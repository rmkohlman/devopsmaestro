package cmd

import (
	"fmt"
	"os"

	"devopsmaestro/db"

	"github.com/spf13/cobra"
)

// getDataStore extracts the DataStore from the cobra command context.
func getDataStore(cmd *cobra.Command) (db.DataStore, error) {
	ctx := cmd.Context()
	dataStore := ctx.Value("dataStore").(*db.DataStore)
	if dataStore == nil {
		return nil, fmt.Errorf("dataStore not initialized")
	}

	return *dataStore, nil
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
