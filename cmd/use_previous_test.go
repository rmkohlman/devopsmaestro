// use_previous_test.go.pending — TDD Phase 2 (RED) for issue #202 Feature 1
//
// Tests for "dvm use -" (previous context toggle, like cd - in shell).
//
// All tests are RED — the feature does not exist yet. These tests drive implementation.
//
// Design:
//   - Before any context switch via dvm use <resource>, current context is serialized
//     to the defaults table with key "context.previous" as JSON.
//   - "dvm use -" reads context.previous, restores it as active, saves old active as
//     the new context.previous (so repeated "dvm use -" toggles between two contexts).
//   - JSON format: {"ecosystem_id":1,"domain_id":2,"app_id":3,"workspace_id":4}
//     (null/missing fields = not set)
//
// To activate: rename to use_previous_test.go after implementation is complete.

package cmd

import (
	"database/sql"
	"encoding/json"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Helpers
// =============================================================================

// previousContextKey is the defaults table key used to store previous context.
// Must match the implementation constant.
const previousContextKey = "context.previous"

// previousContextJSON is the expected JSON structure stored in defaults.
// Fields are pointers so absent context levels can be represented as null/absent.
type previousContextJSON struct {
	EcosystemID *int `json:"ecosystem_id,omitempty"`
	DomainID    *int `json:"domain_id,omitempty"`
	AppID       *int `json:"app_id,omitempty"`
	WorkspaceID *int `json:"workspace_id,omitempty"`
}

// buildPreviousContextJSON serialises context IDs to the JSON string stored in defaults.
func buildPreviousContextJSON(appID, workspaceID *int) string {
	v := previousContextJSON{
		AppID:       appID,
		WorkspaceID: workspaceID,
	}
	b, _ := json.Marshal(v)
	return string(b)
}

// =============================================================================
// TestUseDash_NoPreviousContext
//
// When no previous context has been stored ("context.previous" key absent in defaults),
// "dvm use -" must return an error containing "no previous context".
//
// RED: useCmd does not handle "-" argument at all yet.
// =============================================================================

func TestUseDash_NoPreviousContext(t *testing.T) {
	mock := db.NewMockDataStore()
	// Defaults table is empty — no "context.previous" key
	// mock.Defaults is initialised empty by NewMockDataStore

	useCmd.SetContext(newCmdContextWithMock(mock))
	err := useCmd.RunE(useCmd, []string{"-"})

	// Must return an error (not nil)
	require.Error(t, err, "dvm use - must error when no previous context exists")

	// Error message must be user-friendly
	assert.True(t,
		strings.Contains(strings.ToLower(err.Error()), "no previous context") ||
			strings.Contains(strings.ToLower(err.Error()), "previous"),
		"error message should mention 'previous context'; got: %q", err.Error(),
	)
}

// =============================================================================
// TestUseDash_TogglesBetweenContexts
//
// Full toggle cycle:
//   1. Set active app to A (ID=1), workspace to WS-A (ID=10) — represents "context A"
//   2. "dvm use app B" → saves context A as previous, sets B (ID=2) as active
//   3. "dvm use -" → restores context A (appID=1, wsID=10), saves B as previous
//   4. "dvm use -" again → restores context B (appID=2), saves A as previous
//
// RED: "dvm use -" handler not implemented; "dvm use app" does not save previous.
// =============================================================================

func TestUseDash_TogglesBetweenContexts(t *testing.T) {
	mock := db.NewMockDataStore()

	// Seed two apps
	appA := &models.App{ID: 1, Name: "app-alpha", DomainID: sql.NullInt64{Int64: 1, Valid: true}}
	appB := &models.App{ID: 2, Name: "app-beta", DomainID: sql.NullInt64{Int64: 1, Valid: true}}
	mock.Apps[1] = appA
	mock.Apps[2] = appB

	// Seed workspaces
	wsA := &models.Workspace{ID: 10, Name: "ws-alpha", AppID: 1}
	mock.Workspaces[10] = wsA

	// --- Step 1: establish "context A" as active ---
	appAID := 1
	wsAID := 10
	mock.Context.ActiveAppID = &appAID
	mock.Context.ActiveWorkspaceID = &wsAID

	// --- Step 2: switch to app-beta → must save context A as previous ---
	useAppCmd.SetContext(newCmdContextWithMock(mock))
	err := useAppCmd.RunE(useAppCmd, []string{"app-beta"})
	require.NoError(t, err, "switching to app-beta should succeed")

	// Active context must now be B
	require.NotNil(t, mock.Context.ActiveAppID)
	assert.Equal(t, 2, *mock.Context.ActiveAppID, "active app must be app-beta (ID=2) after switch")

	// Previous context must have been saved in defaults
	prevJSON, ok := mock.Defaults[previousContextKey]
	require.True(t, ok, "defaults must contain %q after context switch", previousContextKey)
	require.NotEmpty(t, prevJSON, "previous context JSON must not be empty")

	// Decode and verify it captured app A
	var prev previousContextJSON
	require.NoError(t, json.Unmarshal([]byte(prevJSON), &prev), "previous context must be valid JSON")
	require.NotNil(t, prev.AppID, "previous context must include AppID")
	assert.Equal(t, 1, *prev.AppID, "previous context must point to app-alpha (ID=1)")

	// --- Step 3: dvm use - → restores context A, saves B as previous ---
	useCmd.SetContext(newCmdContextWithMock(mock))
	err = useCmd.RunE(useCmd, []string{"-"})
	require.NoError(t, err, "dvm use - should succeed when previous context exists")

	// Active context must now be A again
	require.NotNil(t, mock.Context.ActiveAppID)
	assert.Equal(t, 1, *mock.Context.ActiveAppID, "active app must be restored to app-alpha (ID=1)")

	// Previous context must now be B
	prevJSON2, ok2 := mock.Defaults[previousContextKey]
	require.True(t, ok2, "defaults must still contain %q after dvm use -", previousContextKey)
	var prev2 previousContextJSON
	require.NoError(t, json.Unmarshal([]byte(prevJSON2), &prev2))
	require.NotNil(t, prev2.AppID, "new previous context must include AppID")
	assert.Equal(t, 2, *prev2.AppID, "new previous context must point to app-beta (ID=2)")

	// --- Step 4: dvm use - again → restores B ---
	useCmd.SetContext(newCmdContextWithMock(mock))
	err = useCmd.RunE(useCmd, []string{"-"})
	require.NoError(t, err, "second dvm use - should succeed")

	require.NotNil(t, mock.Context.ActiveAppID)
	assert.Equal(t, 2, *mock.Context.ActiveAppID,
		"second dvm use - must restore app-beta (ID=2)")
}

// =============================================================================
// TestUseDash_SavesPreviousOnAppSwitch
//
// When switching from app A to app B via "dvm use app B", the CURRENT context
// (app A + any active workspace) must be serialised and stored in defaults
// under key "context.previous" BEFORE the new context is applied.
//
// RED: "dvm use app" does not call SetDefault("context.previous", ...) yet.
// =============================================================================

func TestUseDash_SavesPreviousOnAppSwitch(t *testing.T) {
	mock := db.NewMockDataStore()

	// Seed app A as the current active context (ID=5)
	appAID := 5
	wsAID := 20
	mock.Context.ActiveAppID = &appAID
	mock.Context.ActiveWorkspaceID = &wsAID

	// Seed the target app B
	appB := &models.App{ID: 7, Name: "new-app", DomainID: sql.NullInt64{Int64: 1, Valid: true}}
	mock.Apps[7] = appB

	// Switch to app B
	useAppCmd.SetContext(newCmdContextWithMock(mock))
	err := useAppCmd.RunE(useAppCmd, []string{"new-app"})
	require.NoError(t, err, "switching to new-app should succeed")

	// Verify that "context.previous" was saved in the defaults table
	prevRaw, exists := mock.Defaults[previousContextKey]
	require.True(t, exists,
		"defaults must contain key %q after dvm use app; got defaults: %v",
		previousContextKey, mock.Defaults)
	require.NotEmpty(t, prevRaw, "context.previous must not be empty string")

	// Decode and verify it captured app A's IDs
	var prev previousContextJSON
	err = json.Unmarshal([]byte(prevRaw), &prev)
	require.NoError(t, err, "context.previous must be valid JSON; got: %q", prevRaw)

	require.NotNil(t, prev.AppID, "context.previous JSON must contain app_id")
	assert.Equal(t, 5, *prev.AppID,
		"context.previous app_id must be the OLD active app (ID=5); got JSON: %s", prevRaw)

	// Workspace should also be captured (it was set)
	require.NotNil(t, prev.WorkspaceID, "context.previous JSON must contain workspace_id when workspace was active")
	assert.Equal(t, 20, *prev.WorkspaceID,
		"context.previous workspace_id must be the OLD active workspace (ID=20); got JSON: %s", prevRaw)
}
