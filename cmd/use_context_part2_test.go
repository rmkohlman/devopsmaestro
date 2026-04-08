// use_context_test.go.pending — continued (part 2)
//
// TestUseWorkspace_NoYAMLWrite
//
// Expected: dvm use workspace <name> writes ONLY to DB — no YAML file created.
// Current: use workspace calls contextMgr.SetWorkspace() which writes to context.yaml.
// This test is RED until Change 1 removes the contextMgr.SetWorkspace() call.

package cmd

import (
	"os"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUseWorkspace_NoYAMLWrite verifies that dvm use workspace <name> only writes
// to the DB and does NOT write/create a YAML context file.
//
// RED until Change 1: remove contextMgr.SetWorkspace() from use workspace.
//
// Current failure mode: contextMgr.SetWorkspace() reads YAML to find active app name.
// In a clean temp HOME there is no YAML, so it fails with "no active app".
// The DB has the active app correctly set — proving the YAML dependency is the bug.
// After the fix: YAML is never read, command succeeds using only the DB.
func TestUseWorkspace_NoYAMLWrite(t *testing.T) {
	// Isolate filesystem writes to a temp HOME
	tempHome := t.TempDir()
	t.Setenv("HOME", tempHome)

	mock := db.NewMockDataStore()

	// Seed active app (required for use workspace — getActiveAppFromContext reads DB)
	appID := 1
	mock.Context.ActiveAppID = &appID
	app := &models.App{ID: 1, Name: "my-api", DomainID: 1}
	mock.Apps[1] = app

	// Seed a workspace under that app
	ws := &models.Workspace{ID: 7, Name: "dev", AppID: 1}
	mock.Workspaces[7] = ws

	useWorkspaceCmd.SetContext(newCmdContextWithMock(mock))
	err := useWorkspaceCmd.RunE(useWorkspaceCmd, []string{"dev"})

	require.NoError(t, err)

	// DB must be updated
	require.NotNil(t, mock.Context.ActiveWorkspaceID, "SetActiveWorkspace must be called after use workspace")
	assert.Equal(t, 7, *mock.Context.ActiveWorkspaceID, "ActiveWorkspaceID must be set to the workspace's ID")

	// YAML file must NOT be written after the fix
	// Before the fix: contextMgr.SetWorkspace() creates/modifies context.yaml
	contextYAML := tempHome + "/.devopsmaestro/context.yaml"
	_, statErr := os.Stat(contextYAML)
	assert.True(t, os.IsNotExist(statErr),
		"YAML context file must NOT be written by 'dvm use workspace' after dual-write removal; file found at %s", contextYAML)
}

// TestUseAppNone_ClearsDB verifies that dvm use app none clears DB context (not just YAML).
//
// This is currently partially correct — app+workspace are cleared in DB, but
// contextMgr.ClearApp() is still called unnecessarily. After Change 1 it should
// skip contextMgr entirely.
//
// GREEN after current code, but verifies the DB behavior is correct.
func TestUseAppNone_ClearsDB(t *testing.T) {
	mock := db.NewMockDataStore()

	// Pre-seed an active app and workspace
	appID, wsID := 3, 8
	mock.Context.ActiveAppID = &appID
	mock.Context.ActiveWorkspaceID = &wsID

	useAppCmd.SetContext(newCmdContextWithMock(mock))
	err := useAppCmd.RunE(useAppCmd, []string{"none"})

	require.NoError(t, err)
	assert.Nil(t, mock.Context.ActiveAppID, "use app none must clear ActiveAppID in DB")
	assert.Nil(t, mock.Context.ActiveWorkspaceID, "use app none must clear ActiveWorkspaceID in DB")
}

// TestUseWorkspaceNone_ClearsDB verifies that dvm use workspace none clears DB workspace context.
//
// Currently partially correct (DB is cleared) but contextMgr.ClearWorkspace() is still called.
// After Change 1 it should skip contextMgr entirely.
//
// GREEN after current code, verifies DB behavior is correct.
func TestUseWorkspaceNone_ClearsDB(t *testing.T) {
	mock := db.NewMockDataStore()

	// Pre-seed an active app and workspace
	appID, wsID := 2, 6
	mock.Context.ActiveAppID = &appID
	mock.Context.ActiveWorkspaceID = &wsID

	// Provide the app record so getActiveAppFromContext resolves it
	app := &models.App{ID: 2, Name: "my-api", DomainID: 1}
	mock.Apps[2] = app

	useWorkspaceCmd.SetContext(newCmdContextWithMock(mock))
	err := useWorkspaceCmd.RunE(useWorkspaceCmd, []string{"none"})

	require.NoError(t, err)
	assert.Nil(t, mock.Context.ActiveWorkspaceID, "use workspace none must clear ActiveWorkspaceID in DB")
	// App should remain untouched
	require.NotNil(t, mock.Context.ActiveAppID, "use workspace none must NOT clear ActiveAppID")
	assert.Equal(t, 2, *mock.Context.ActiveAppID, "ActiveAppID should remain unchanged")
}
