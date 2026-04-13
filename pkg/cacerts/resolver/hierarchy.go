package resolver

import (
	"context"
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
)

// defaultsCACertsKey is the key used in the defaults table to store global CA certs.
const defaultsCACertsKey = "ca-certs"

// maxMergedCACerts is the maximum number of CA certificates allowed after cascade
// resolution. This mirrors the validation limit in models.ValidateCACerts().
const maxMergedCACerts = 10

// HierarchyCACertsResolver implements CACertsResolver by walking the full
// 6-level hierarchy: global → ecosystem → domain → system → app → workspace.
// More-specific levels override less-specific levels for certs with the same Name.
type HierarchyCACertsResolver struct {
	store db.DataStore
}

// NewHierarchyCACertsResolver creates a new CACertsResolver backed
// by the provided DataStore.
func NewHierarchyCACertsResolver(store db.DataStore) CACertsResolver {
	return &HierarchyCACertsResolver{store: store}
}

// Resolve walks the full hierarchy for the given workspace and merges all CA
// certs from global → ecosystem → domain → system → app → workspace (workspace wins).
// Certs with the same Name at a more-specific level override those from less-specific
// levels. Uniquely-named certs from all levels are additively merged.
// The system level is optional — if the app has no system, that level is skipped gracefully.
// Returns an error if the merged result exceeds the 10-cert maximum.
func (r *HierarchyCACertsResolver) Resolve(ctx context.Context, workspaceID int) (*CACertsResolution, error) {
	// ─── Load workspace ───────────────────────────────────────────────────────
	ws, err := r.store.GetWorkspaceByID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("resolving CA certs: getting workspace %d: %w", workspaceID, err)
	}

	// ─── Load app ─────────────────────────────────────────────────────────────
	app, err := r.store.GetAppByID(ws.AppID)
	if err != nil {
		return nil, fmt.Errorf("resolving CA certs: getting app %d: %w", ws.AppID, err)
	}

	// ─── Load domain ──────────────────────────────────────────────────────────
	domain, err := r.store.GetDomainByID(app.DomainID)
	if err != nil {
		return nil, fmt.Errorf("resolving CA certs: getting domain %d: %w", app.DomainID, err)
	}

	// ─── Load system (optional) ───────────────────────────────────────────────
	var systemName string
	var systemCerts []models.CACertConfig
	if app.SystemID.Valid {
		system, sErr := r.store.GetSystemByID(int(app.SystemID.Int64))
		if sErr != nil {
			return nil, fmt.Errorf("resolving CA certs: getting system %d: %w", app.SystemID.Int64, sErr)
		}
		systemName = system.Name
		systemCerts = parseDirectCACerts(nullableString(system.CACerts.String, system.CACerts.Valid))
	}

	// ─── Load ecosystem ───────────────────────────────────────────────────────
	eco, err := r.store.GetEcosystemByID(domain.EcosystemID)
	if err != nil {
		return nil, fmt.Errorf("resolving CA certs: getting ecosystem %d: %w", domain.EcosystemID, err)
	}

	// ─── Parse each level ────────────────────────────────────────────────────
	globalCerts := parseGlobalCACerts(r.store)
	ecoCerts := parseDirectCACerts(nullableString(eco.CACerts.String, eco.CACerts.Valid))
	domainCerts := parseDirectCACerts(nullableString(domain.CACerts.String, domain.CACerts.Valid))
	appCerts := parseWrappedCACerts(nullableString(app.BuildConfig.String, app.BuildConfig.Valid))
	wsCerts := parseWrappedDevBuildCACerts(nullableString(ws.BuildConfig.String, ws.BuildConfig.Valid))

	// ─── Build path (always 6 entries) ───────────────────────────────────────
	path := []CACertsResolutionStep{
		{Level: LevelGlobal, Name: "global", Certs: toCACertEntries(globalCerts, LevelGlobal), Found: len(globalCerts) > 0},
		{Level: LevelEcosystem, Name: eco.Name, Certs: toCACertEntries(ecoCerts, LevelEcosystem), Found: len(ecoCerts) > 0},
		{Level: LevelDomain, Name: domain.Name, Certs: toCACertEntries(domainCerts, LevelDomain), Found: len(domainCerts) > 0},
		{Level: LevelSystem, Name: systemName, Certs: toCACertEntries(systemCerts, LevelSystem), Found: len(systemCerts) > 0},
		{Level: LevelApp, Name: app.Name, Certs: toCACertEntries(appCerts, LevelApp), Found: len(appCerts) > 0},
		{Level: LevelWorkspace, Name: ws.Name, Certs: toCACertEntries(wsCerts, LevelWorkspace), Found: len(wsCerts) > 0},
	}

	// ─── Cascade merge (global first, workspace last = highest precedence) ────
	// Use a map keyed by cert Name to handle override semantics.
	mergedByName := make(map[string]CACertEntry)
	sources := make(map[string]HierarchyLevel)

	for _, step := range path {
		for _, cert := range step.Certs {
			mergedByName[cert.Name] = cert
			sources[cert.Name] = step.Level
		}
	}

	// ─── Convert merged map to ordered slice ────────────────────────────────
	merged := make([]CACertEntry, 0, len(mergedByName))
	for _, cert := range mergedByName {
		merged = append(merged, cert)
	}

	// ─── Validate merged count against maximum ──────────────────────────────
	if len(merged) > maxMergedCACerts {
		return nil, fmt.Errorf("CA certs cascade resolved to %d certificates, exceeding the maximum of %d", len(merged), maxMergedCACerts)
	}

	return &CACertsResolution{
		Certs:   merged,
		Sources: sources,
		Path:    path,
	}, nil
}

// ─── internal helpers ─────────────────────────────────────────────────────────

// nullableString returns the string if valid, otherwise "".
func nullableString(s string, valid bool) string {
	if !valid {
		return ""
	}
	return s
}

// parseGlobalCACerts retrieves the global CA certs from the defaults table
// and returns them as a slice of CACertConfig.
func parseGlobalCACerts(store db.DataStore) []models.CACertConfig {
	raw, _ := store.GetDefault(defaultsCACertsKey)
	if raw == "" {
		return nil
	}
	var certs []models.CACertConfig
	if err := json.Unmarshal([]byte(raw), &certs); err != nil {
		return nil
	}
	return certs
}

// parseDirectCACerts parses a JSON string that is directly a []CACertConfig.
// Used for: ecosystem.ca_certs, domain.ca_certs.
func parseDirectCACerts(raw string) []models.CACertConfig {
	if raw == "" {
		return nil
	}
	var certs []models.CACertConfig
	if err := json.Unmarshal([]byte(raw), &certs); err != nil {
		return nil
	}
	return certs
}

// parseWrappedCACerts parses a JSON string of the form {"caCerts": [...]}
// (AppBuildConfig format). Used for: app.build_config.
func parseWrappedCACerts(raw string) []models.CACertConfig {
	if raw == "" {
		return nil
	}
	var wrapper struct {
		CACerts []models.CACertConfig `json:"caCerts"`
	}
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return nil
	}
	return wrapper.CACerts
}

// parseWrappedDevBuildCACerts parses a JSON string of the form {"caCerts": [...]}
// (DevBuildConfig format). Used for: workspace.build_config.
func parseWrappedDevBuildCACerts(raw string) []models.CACertConfig {
	if raw == "" {
		return nil
	}
	var wrapper struct {
		CACerts []models.CACertConfig `json:"caCerts"`
	}
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return nil
	}
	return wrapper.CACerts
}

// toCACertEntries converts a slice of CACertConfig to CACertEntry with the
// given source level.
func toCACertEntries(certs []models.CACertConfig, level HierarchyLevel) []CACertEntry {
	entries := make([]CACertEntry, 0, len(certs))
	for _, c := range certs {
		entries = append(entries, CACertEntry{
			Name:             c.Name,
			VaultSecret:      c.VaultSecret,
			VaultEnvironment: c.VaultEnvironment,
			VaultField:       c.VaultField,
			Source:           level,
		})
	}
	return entries
}
