package handlers

import (
	"database/sql"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// MoveTarget specifies the destination of a Move operation.
//
// Semantics by handler:
//   - SystemHandler.Move: DomainName + EcosystemName describe the new parent Domain.
//   - AppHandler.Move:    SystemName (with optional DomainName + EcosystemName as
//     disambiguators) describes the new parent System.
//     If SystemName is "" and DomainName is set, the App is
//     moved to the Domain with no System.
//
// EcosystemName is always optional and is used as a disambiguator when a Domain
// or System name is not unique across ecosystems.
type MoveTarget struct {
	EcosystemName string
	DomainName    string
	SystemName    string
}

// MoveResult summarizes the outcome of a Move operation for kubectl-style output.
type MoveResult struct {
	Kind         string // "system" or "app"
	Name         string
	FromParent   string // e.g. "domain/old-domain" or "system/old-system"
	ToParent     string // e.g. "domain/new-domain" or "system/new-system"
	CascadedApps int    // number of child apps reparented (System moves only)
	NoOp         bool   // true if already at target (idempotent success)
}

// String renders the result kubectl-style:
//
//	system/foo moved to domain/bar (3 apps cascaded)
//	app/baz moved to system/qux
//	app/baz detached from system (now at ecosystem level)
//	app/baz already at system/qux (no-op)
func (r *MoveResult) String() string {
	if r.NoOp {
		return fmt.Sprintf("%s/%s already at %s (no-op)", r.Kind, r.Name, r.ToParent)
	}
	base := fmt.Sprintf("%s/%s moved to %s", r.Kind, r.Name, r.ToParent)
	if r.CascadedApps > 0 {
		base += fmt.Sprintf(" (%d app%s cascaded)", r.CascadedApps, plural(r.CascadedApps))
	}
	return base
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

// resolveDomainTarget looks up a Domain by name, optionally scoped to an Ecosystem
// when the ecosystem name is provided. Returns the Domain, its owning Ecosystem
// name (for display), and an error.
//
// When ecosystemName is empty and the domain name is ambiguous across multiple
// ecosystems, returns a clear error asking the caller to disambiguate.
func resolveDomainTarget(ds db.DataStore, ecosystemName, domainName string) (*models.Domain, string, error) {
	if domainName == "" {
		return nil, "", fmt.Errorf("target domain name is required")
	}
	if ecosystemName != "" {
		eco, err := ds.GetEcosystemByName(ecosystemName)
		if err != nil {
			return nil, "", fmt.Errorf("ecosystem '%s' not found: %w", ecosystemName, err)
		}
		dom, err := ds.GetDomainByName(sql.NullInt64{Int64: int64(eco.ID), Valid: true}, domainName)
		if err != nil {
			return nil, "", fmt.Errorf("domain '%s' not found in ecosystem '%s': %w", domainName, ecosystemName, err)
		}
		return dom, eco.Name, nil
	}
	matches, err := ds.FindDomainsByName(domainName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to look up domain '%s': %w", domainName, err)
	}
	if len(matches) == 0 {
		return nil, "", fmt.Errorf("domain '%s' not found", domainName)
	}
	if len(matches) > 1 {
		return nil, "", fmt.Errorf("domain '%s' exists in multiple ecosystems; specify ecosystem to disambiguate", domainName)
	}
	ecoName := ""
	if matches[0].Ecosystem != nil {
		ecoName = matches[0].Ecosystem.Name
	}
	return matches[0].Domain, ecoName, nil
}

// resolveSystemTarget looks up a System by name, optionally scoped by Domain
// (and transitively Ecosystem). Returns the System, its owning Domain (which
// the caller will use as the new App.DomainID), and an error.
func resolveSystemTarget(ds db.DataStore, ecosystemName, domainName, systemName string) (*models.System, *models.Domain, error) {
	if systemName == "" {
		return nil, nil, fmt.Errorf("target system name is required")
	}
	matches, err := ds.FindSystemsByName(systemName)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to look up system '%s': %w", systemName, err)
	}
	if len(matches) == 0 {
		return nil, nil, fmt.Errorf("system '%s' not found", systemName)
	}

	// Filter by ecosystem/domain hints when provided.
	var filtered []*models.SystemWithHierarchy
	for _, m := range matches {
		if ecosystemName != "" {
			if m.Ecosystem == nil || m.Ecosystem.Name != ecosystemName {
				continue
			}
		}
		if domainName != "" {
			if m.Domain == nil || m.Domain.Name != domainName {
				continue
			}
		}
		filtered = append(filtered, m)
	}
	if len(filtered) == 0 {
		// Ambiguity-or-not-found after filter.
		return nil, nil, fmt.Errorf("system '%s' not found matching ecosystem=%q domain=%q",
			systemName, ecosystemName, domainName)
	}
	if len(filtered) > 1 {
		return nil, nil, fmt.Errorf("system '%s' is ambiguous across multiple domains; specify --to-domain (and -e) to disambiguate", systemName)
	}
	chosen := filtered[0]
	if chosen.Domain == nil {
		return nil, nil, fmt.Errorf("system '%s' has no parent domain; cannot use as move target", systemName)
	}
	return chosen.System, chosen.Domain, nil
}

// dataStoreFromCtx is a small helper to keep handler call sites tidy.
func dataStoreFromCtx(ctx resource.Context) (db.DataStore, error) {
	return resource.DataStoreAs[db.DataStore](ctx)
}
