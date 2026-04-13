// use_context_test.go.pending — TDD Phase 2 (RED) for issue #201 Change 1
//
// Tests validate desired behavior AFTER the dual-write elimination fix.
//
// Status of each test:
//   RED  — TestUseClear_ClearsDBContext: --clear currently only clears YAML, not DB (BUG)
//   RED  — TestUseApp_NoYAMLWrite: use app writes to YAML; after fix, YAML file must not be modified
//   RED  — TestUseWorkspace_NoYAMLWrite: use workspace writes to YAML; after fix, must not modify YAML
//
// To activate: rename to use_context_test.go after implementation is complete.

package cmd

import (
	"database/sql"
	"context"
	"os"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Helpers
// =============================================================================

// newCmdContextWithMock builds a context.Context carrying a MockDataStore,
// matching the pattern consumed by getDataStore().
func newCmdContextWithMock(ds *db.MockDataStore) context.Context {
	return context.WithValue(context.Background(), CtxKeyDataStore, ds)
}

// wasMethodCalled checks the MockDataStore call log for a specific method name.
func wasMethodCalled(mock *db.MockDataStore, method string) bool {
	for _, call := range mock.GetCalls() {
		if call.Method == method {
			return true
		}
	}
	return false
}

// =============================================================================
// TestUseClear_ClearsDBContext
//
// Expected: dvm use --clear clears ALL 4 DB context fields (eco, domain, app, ws).
// Current (BUG): --clear only calls contextMgr.ClearApp() on YAML, never touches DB.
// This test is RED until Change 1 is implemented.
// =============================================================================

func TestUseClear_ClearsDBContext(t *testing.T) {
	mock := db.NewMockDataStore()

	// Pre-populate all 4 active IDs in the DB context
	ecoID, domID, appID, wsID := 1, 2, 3, 4
	mock.Context.ActiveEcosystemID = &ecoID
	mock.Context.ActiveDomainID = &domID
	mock.Context.ActiveAppID = &appID
	mock.Context.ActiveWorkspaceID = &wsID

	// Set --clear flag and run
	require.NoError(t, useCmd.Flags().Set("clear", "true"))
	defer useCmd.Flags().Set("clear", "false") // cleanup

	useCmd.SetContext(newCmdContextWithMock(mock))
	err := useCmd.RunE(useCmd, []string{})

	// Must succeed
	require.NoError(t, err)

	// All 4 DB context fields must be nil — BUG: currently none are cleared
	assert.Nil(t, mock.Context.ActiveEcosystemID, "use --clear must clear DB ecosystem context")
	assert.Nil(t, mock.Context.ActiveDomainID, "use --clear must clear DB domain context")
	assert.Nil(t, mock.Context.ActiveAppID, "use --clear must clear DB app context")
	assert.Nil(t, mock.Context.ActiveWorkspaceID, "use --clear must clear DB workspace context")
}

// =============================================================================
// TestUseApp_NoYAMLWrite
//
// Expected: dvm use app <name> writes ONLY to DB — no context.yaml file is created/modified.
// Current: use app calls contextMgr.SetApp() which writes to ~/.config/devopsmaestro/context.yaml.
// This test is RED until Change 1 removes the contextMgr.SetApp() call.
//
// Strategy: run the command with HOME pointing to a fresh temp dir. After the fix,
// no YAML file should be written there. Before the fix, context.yaml is created.
// =============================================================================

func TestUseApp_NoYAMLWrite(t *testing.T) {
	// Create an isolated HOME so we can detect filesystem writes
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	mock := db.NewMockDataStore()
	app := &models.App{ID: 5, Name: "my-api", DomainID: sql.NullInt64{Int64: 1, Valid: true}}
	mock.Apps[5] = app

	useAppCmd.SetContext(newCmdContextWithMock(mock))
	err := useAppCmd.RunE(useAppCmd, []string{"my-api"})

	require.NoError(t, err)

	// DB must be updated
	require.NotNil(t, mock.Context.ActiveAppID, "SetActiveApp must be called after use app")
	assert.Equal(t, 5, *mock.Context.ActiveAppID, "ActiveAppID must be set to the app's ID")

	// YAML file must NOT be written — after the fix, no context.yaml is created
	// Before the fix: contextMgr.SetApp() creates/modifies context.yaml at
	// $HOME/.devopsmaestro/context.yaml
	contextYAML := tempHome + "/.devopsmaestro/context.yaml"
	_, statErr := os.Stat(contextYAML)
	assert.True(t, os.IsNotExist(statErr),
		"YAML context file must NOT be written by 'dvm use app' after dual-write removal; file found at %s", contextYAML)
}
