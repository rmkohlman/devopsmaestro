package resolver

import (
	"context"
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/pkg/envvalidation"
)

// defaultsBuildArgsKey is the key used in the defaults table to store global build args.
const defaultsBuildArgsKey = "build-args"

// HierarchyBuildArgsResolver implements BuildArgsResolver by walking the full
// 6-level hierarchy: global → ecosystem → domain → system → app → workspace.
// More-specific levels override less-specific levels for the same key.
type HierarchyBuildArgsResolver struct {
	store db.DataStore
}

// NewHierarchyBuildArgsResolver creates a new BuildArgsResolver backed
// by the provided DataStore.
func NewHierarchyBuildArgsResolver(store db.DataStore) BuildArgsResolver {
	return &HierarchyBuildArgsResolver{store: store}
}

// Resolve walks the full hierarchy for the given workspace and merges all build
// args from global → ecosystem → domain → system → app → workspace (workspace wins).
// Keys that fail ValidateEnvKey() or are in the IsDangerousEnvVar() denylist are
// silently filtered as a defence-in-depth measure.
// The system level is optional — if the app has no system, that level is skipped gracefully.
func (r *HierarchyBuildArgsResolver) Resolve(ctx context.Context, workspaceID int) (*BuildArgsResolution, error) {
	// ─── Load workspace ───────────────────────────────────────────────────────
	ws, err := r.store.GetWorkspaceByID(workspaceID)
	if err != nil {
		return nil, fmt.Errorf("resolving build args: getting workspace %d: %w", workspaceID, err)
	}

	// ─── Load app ─────────────────────────────────────────────────────────────
	app, err := r.store.GetAppByID(ws.AppID)
	if err != nil {
		return nil, fmt.Errorf("resolving build args: getting app %d: %w", ws.AppID, err)
	}

	// ─── Load domain ──────────────────────────────────────────────────────────
	domain, err := r.store.GetDomainByID(app.DomainID)
	if err != nil {
		return nil, fmt.Errorf("resolving build args: getting domain %d: %w", app.DomainID, err)
	}

	// ─── Load system (optional) ───────────────────────────────────────────────
	var systemName string
	var systemArgs map[string]string
	if app.SystemID.Valid {
		system, sErr := r.store.GetSystemByID(int(app.SystemID.Int64))
		if sErr != nil {
			return nil, fmt.Errorf("resolving build args: getting system %d: %w", app.SystemID.Int64, sErr)
		}
		systemName = system.Name
		systemArgs = parseDirectMap(nullableString(system.BuildArgs.String, system.BuildArgs.Valid))
	} else {
		systemArgs = map[string]string{}
	}

	// ─── Load ecosystem ───────────────────────────────────────────────────────
	eco, err := r.store.GetEcosystemByID(domain.EcosystemID)
	if err != nil {
		return nil, fmt.Errorf("resolving build args: getting ecosystem %d: %w", domain.EcosystemID, err)
	}

	// ─── Load global build args ───────────────────────────────────────────────
	globalArgs := parseDirectMap(getDefault(r.store, defaultsBuildArgsKey))

	// ─── Parse each level ────────────────────────────────────────────────────
	ecoArgs := parseDirectMap(nullableString(eco.BuildArgs.String, eco.BuildArgs.Valid))
	domainArgs := parseDirectMap(nullableString(domain.BuildArgs.String, domain.BuildArgs.Valid))
	appArgs := parseWrappedArgs(nullableString(app.BuildConfig.String, app.BuildConfig.Valid))
	wsArgs := parseWrappedArgs(nullableString(ws.BuildConfig.String, ws.BuildConfig.Valid))

	// ─── Build path (always 6 entries) ───────────────────────────────────────
	path := []BuildArgsStep{
		{Level: LevelGlobal, Name: "global", Args: globalArgs, Found: len(globalArgs) > 0},
		{Level: LevelEcosystem, Name: eco.Name, Args: ecoArgs, Found: len(ecoArgs) > 0},
		{Level: LevelDomain, Name: domain.Name, Args: domainArgs, Found: len(domainArgs) > 0},
		{Level: LevelSystem, Name: systemName, Args: systemArgs, Found: len(systemArgs) > 0},
		{Level: LevelApp, Name: app.Name, Args: appArgs, Found: len(appArgs) > 0},
		{Level: LevelWorkspace, Name: ws.Name, Args: wsArgs, Found: len(wsArgs) > 0},
	}

	// ─── Cascade merge (global first, workspace last = highest precedence) ────
	merged := make(map[string]string)
	sources := make(map[string]HierarchyLevel)

	for _, step := range path {
		for k, v := range step.Args {
			if err := envvalidation.ValidateEnvKey(k); err != nil {
				continue // defence-in-depth: silently filter invalid keys
			}
			merged[k] = v
			sources[k] = step.Level
		}
	}

	return &BuildArgsResolution{
		Args:    merged,
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

// getDefault safely retrieves a value from the defaults table, returning "" on error.
func getDefault(store db.DataStore, key string) string {
	v, _ := store.GetDefault(key)
	return v
}

// parseDirectMap parses a JSON string that is directly a map[string]string.
// Used for: global (defaults table), ecosystem.build_args, domain.build_args.
func parseDirectMap(raw string) map[string]string {
	if raw == "" {
		return map[string]string{}
	}
	var m map[string]string
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return map[string]string{}
	}
	if m == nil {
		return map[string]string{}
	}
	return m
}

// parseWrappedArgs parses a JSON string of the form {"args": {...}}.
// Used for: app.build_config, workspace.build_config.
func parseWrappedArgs(raw string) map[string]string {
	if raw == "" {
		return map[string]string{}
	}
	var wrapper struct {
		Args map[string]string `json:"args"`
	}
	if err := json.Unmarshal([]byte(raw), &wrapper); err != nil {
		return map[string]string{}
	}
	if wrapper.Args == nil {
		return map[string]string{}
	}
	return wrapper.Args
}
