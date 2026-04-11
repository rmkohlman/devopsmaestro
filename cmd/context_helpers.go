package cmd

import (
	"fmt"
	"os"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/mirror"

	"github.com/spf13/cobra"
)

// getDataStore extracts the DataStore from the cobra command context.
// It safely handles all the ways a DataStore may have been stored:
//   - *db.DataStore  (production: main.go passes pointer-to-interface)
//   - db.DataStore   (tests that store the interface directly)
//   - *db.MockDataStore / db.MockDataStore (tests with mock)
func getDataStore(cmd *cobra.Command) (db.DataStore, error) {
	ctx := cmd.Context()
	val := ctx.Value(CtxKeyDataStore)
	if val == nil {
		return nil, fmt.Errorf("dataStore not found in context")
	}

	switch ds := val.(type) {
	case *db.DataStore:
		if ds == nil {
			return nil, fmt.Errorf("dataStore not initialized")
		}
		return *ds, nil
	case db.DataStore:
		return ds, nil
	case *db.MockDataStore:
		return ds, nil
	default:
		return nil, fmt.Errorf("invalid dataStore type in context: %T", val)
	}
}

// getActiveAppFromContext returns the active app name from DB context, with env var override.
// Precedence: DVM_APP env var > DB context (active_app_id) > error
func getActiveAppFromContext(ds db.DataStore) (string, error) {
	// Check environment variable first
	if app := os.Getenv("DVM_APP"); app != "" {
		return app, nil
	}

	// Read from database context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return "", fmt.Errorf("no active app context (use 'dvm use app <name>' or set DVM_APP)")
	}

	if dbCtx == nil || dbCtx.ActiveAppID == nil {
		return "", fmt.Errorf("no active app context (use 'dvm use app <name>' or set DVM_APP)")
	}

	app, err := ds.GetAppByID(*dbCtx.ActiveAppID)
	if err != nil {
		return "", fmt.Errorf("no active app context (use 'dvm use app <name>' or set DVM_APP)")
	}

	return app.Name, nil
}

// getActiveWorkspaceFromContext returns the active workspace name from DB context, with env var override.
// Precedence: DVM_WORKSPACE env var > DB context (active_workspace_id) > error
func getActiveWorkspaceFromContext(ds db.DataStore) (string, error) {
	// Check environment variable first
	if workspace := os.Getenv("DVM_WORKSPACE"); workspace != "" {
		return workspace, nil
	}

	// Read from database context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return "", fmt.Errorf("no active workspace context (use 'dvm use workspace <name>' or set DVM_WORKSPACE)")
	}

	if dbCtx == nil || dbCtx.ActiveWorkspaceID == nil {
		return "", fmt.Errorf("no active workspace context (use 'dvm use workspace <name>' or set DVM_WORKSPACE)")
	}

	ws, err := ds.GetWorkspaceByID(*dbCtx.ActiveWorkspaceID)
	if err != nil {
		return "", fmt.Errorf("no active workspace context (use 'dvm use workspace <name>' or set DVM_WORKSPACE)")
	}

	return ws.Name, nil
}

// getActiveEcosystemFromContext returns the active ecosystem name.
// Checks DVM_ECOSYSTEM env var first, then DB context.
func getActiveEcosystemFromContext(ds db.DataStore) (string, error) {
	// Check environment variable first
	if eco := os.Getenv("DVM_ECOSYSTEM"); eco != "" {
		return eco, nil
	}

	// Read from database context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return "", fmt.Errorf("no active ecosystem context (use 'dvm use ecosystem <name>' or set DVM_ECOSYSTEM)")
	}

	if dbCtx == nil || dbCtx.ActiveEcosystemID == nil {
		return "", fmt.Errorf("no active ecosystem context (use 'dvm use ecosystem <name>' or set DVM_ECOSYSTEM)")
	}

	eco, err := ds.GetEcosystemByID(*dbCtx.ActiveEcosystemID)
	if err != nil {
		return "", fmt.Errorf("no active ecosystem context (use 'dvm use ecosystem <name>' or set DVM_ECOSYSTEM)")
	}

	return eco.Name, nil
}

// getActiveDomainFromContext returns the active domain name.
// Checks DVM_DOMAIN env var first, then DB context.
func getActiveDomainFromContext(ds db.DataStore) (string, error) {
	// Check environment variable first
	if dom := os.Getenv("DVM_DOMAIN"); dom != "" {
		return dom, nil
	}

	// Read from database context
	dbCtx, err := ds.GetContext()
	if err != nil {
		return "", fmt.Errorf("no active domain context (use 'dvm use domain <name>' or set DVM_DOMAIN)")
	}

	if dbCtx == nil || dbCtx.ActiveDomainID == nil {
		return "", fmt.Errorf("no active domain context (use 'dvm use domain <name>' or set DVM_DOMAIN)")
	}

	dom, err := ds.GetDomainByID(*dbCtx.ActiveDomainID)
	if err != nil {
		return "", fmt.Errorf("no active domain context (use 'dvm use domain <name>' or set DVM_DOMAIN)")
	}

	return dom.Name, nil
}

// getMirrorManager extracts the MirrorManager from the cobra command context.
// It checks the context first (for testing), then falls back to creating a real manager.
func getMirrorManager(cmd *cobra.Command) mirror.MirrorManager {
	ctx := cmd.Context()
	if val := ctx.Value(CtxKeyMirrorManager); val != nil {
		if mm, ok := val.(mirror.MirrorManager); ok {
			return mm
		}
	}
	// Fall back to real manager using default git repo base directory
	return mirror.NewGitMirrorManager(getGitRepoBaseDir())
}

// resolveAppByNameScoped resolves an app by name, scoped to the active ecosystem
// context. When an active ecosystem is set, it prefers the app within that ecosystem
// to avoid cross-ecosystem workspace creation (issue #250). Falls back to global
// lookup when no ecosystem context is active.
func resolveAppByNameScoped(ds db.DataStore, appName string) (*models.App, error) {
	// Try ecosystem-scoped resolution first
	eco, ecoErr := getActiveEcosystem(ds)
	if ecoErr == nil && eco != nil {
		matches, err := ds.FindAppsByName(appName)
		if err == nil && len(matches) > 0 {
			// Filter to apps in the active ecosystem
			for _, m := range matches {
				if m.Ecosystem.ID == eco.ID {
					return m.App, nil
				}
			}
			// App exists but not in the active ecosystem — still return the first
			// match so existing behavior is preserved for single-app names
		}
	}

	// Fall back to global lookup (no active ecosystem or app not found in ecosystem)
	return ds.GetAppByNameGlobal(appName)
}
