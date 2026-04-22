package cmd

// =============================================================================
// TDD Phase 2 (RED): Build Status Stale-Session Healing -- Issue #399
// =============================================================================
// These tests verify that `dvm build status` heals stale sessions that are
// stuck in "running" with non-terminal workspace entries (the post-SIGINT
// state). The current heal logic in renderBuildSession only fires when ALL
// workspaces are in terminal states; it does not handle the "building"/"queued"
// stuck-state that results from SIGINT.
//
// RED state: ALL assertions in this file FAIL because:
//   - allWorkspacesTerminal() returns false for "building"/"queued" entries
//   - renderBuildSession only heals when allWorkspacesTerminal is true
//   - There is no time-based staleness check (started_at age)
//   - No heal path writes "interrupted" for non-terminal workspace rows
//
// GREEN state: After dvm-core implements Issue #399, renderBuildSession will:
//   - Detect sessions older than N minutes still "running" with non-terminal
//     workspaces and mark them "interrupted"
//   - Update those workspace rows to "interrupted"/"cancelled" as appropriate
// =============================================================================

import (
	"database/sql"
	"testing"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildStaleInterruptedSession seeds a store with a session that looks like
// what SIGINT leaves behind: status="running", workspaces in "building",
// started_at far in the past (1 hour ago).
func buildStaleInterruptedSession(t *testing.T, store *db.MockDataStore, startedAgo time.Duration) (sessionID string) {
	t.Helper()

	eco := &models.Ecosystem{Name: "stale-eco"}
	require.NoError(t, store.CreateEcosystem(eco))
	dom := &models.Domain{
		Name:        "stale-dom",
		EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true},
	}
	require.NoError(t, store.CreateDomain(dom))

	app := &models.App{
		Name:     "stale-app",
		DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true},
	}
	require.NoError(t, store.CreateApp(app))
	ws := &models.Workspace{Name: "dev", AppID: app.ID, ImageName: ":pending"}
	require.NoError(t, store.CreateWorkspace(ws))

	startedAt := time.Now().UTC().Add(-startedAgo)
	sessionID = "stale-session-" + startedAt.Format("20060102T150405")

	session := &models.BuildSession{
		ID:              sessionID,
		StartedAt:       startedAt,
		Status:          "running",
		TotalWorkspaces: 2,
	}
	require.NoError(t, store.CreateBuildSession(session))

	// Workspace 1: was "building" (picked up by a worker before SIGINT)
	bsw1 := &models.BuildSessionWorkspace{
		SessionID:   sessionID,
		WorkspaceID: ws.ID,
		Status:      "building",
		StartedAt:   sql.NullTime{Time: startedAt, Valid: true},
	}
	require.NoError(t, store.CreateBuildSessionWorkspace(bsw1))

	// Workspace 2: still "queued" (never started before SIGINT)
	bsw2 := &models.BuildSessionWorkspace{
		SessionID:   sessionID,
		WorkspaceID: ws.ID,
		Status:      "queued",
		StartedAt:   sql.NullTime{Time: startedAt, Valid: true},
	}
	require.NoError(t, store.CreateBuildSessionWorkspace(bsw2))

	return sessionID
}

// =============================================================================
// 1. Stale "running" session with non-terminal workspaces is healed
// =============================================================================

// TestBuildStatus_HealsStaleRunningSessionAsInterrupted verifies that when
// renderBuildSession is called with a session that is:
//   - status = "running"
//   - started_at > staleThreshold minutes ago
//   - has workspace entries in non-terminal states ("building", "queued")
//
// ...it heals the session to status="interrupted" and updates the database.
//
// RED: renderBuildSession only heals via allWorkspacesTerminal, which returns
// false for "building"/"queued". The stale session stays as "running" forever.
func TestBuildStatus_HealsStaleRunningSessionAsInterrupted(t *testing.T) {
	store := db.NewMockDataStore()
	sessionID := buildStaleInterruptedSession(t, store, 2*time.Hour)

	session, err := store.GetBuildSession(sessionID)
	require.NoError(t, err)
	require.NotNil(t, session)
	require.Equal(t, "running", session.Status)

	// Calling renderBuildSession should trigger stale-session healing.
	// The function writes to stdout which we ignore in this test.
	_ = renderBuildSession(store, session)

	// After healing, the persisted session must reflect "interrupted".
	healed, err := store.GetBuildSession(sessionID)
	require.NoError(t, err)
	require.NotNil(t, healed)

	assert.Equal(t, "interrupted", healed.Status,
		"[RED] renderBuildSession must heal a stale 'running' session (started 2h ago, "+
			"workspaces stuck in 'building'/'queued') to status='interrupted'. "+
			"Currently allWorkspacesTerminal() returns false so no heal occurs. "+
			"Fix: add a time-based staleness check in renderBuildSession.")

	assert.True(t, healed.CompletedAt.Valid,
		"[RED] CompletedAt must be set when session is healed to 'interrupted'.")
}

// =============================================================================
// 2. Workspace rows are updated to non-terminal-interrupt states during heal
// =============================================================================

// TestBuildStatus_HealsWorkspaceRowsDuringStaleSessionHeal verifies that when
// a stale session is healed, its workspace entries in "building"/"queued" are
// also updated so that subsequent `dvm build status` calls show accurate data.
//
// RED: workspace rows are never updated by the heal path.
func TestBuildStatus_HealsWorkspaceRowsDuringStaleSessionHeal(t *testing.T) {
	store := db.NewMockDataStore()
	sessionID := buildStaleInterruptedSession(t, store, 90*time.Minute)

	session, err := store.GetBuildSession(sessionID)
	require.NoError(t, err)
	require.NotNil(t, session)

	_ = renderBuildSession(store, session)

	entries, err := store.GetBuildSessionWorkspaces(sessionID)
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	for _, e := range entries {
		isHealed := e.Status == "interrupted" || e.Status == "cancelled"
		assert.True(t, isHealed,
			"[RED] Workspace %d has status '%s' after stale-session heal. "+
				"Expected 'interrupted' (was building) or 'cancelled' (was queued). "+
				"Fix: heal path must also update non-terminal workspace rows.",
			e.WorkspaceID, e.Status)
	}
}

// =============================================================================
// 3. Recent "running" sessions are NOT incorrectly healed as interrupted
// =============================================================================

// TestBuildStatus_DoesNotHealRecentRunningSession verifies that a session
// started recently (within the stale threshold) with "building" workspaces is
// NOT healed to "interrupted" -- it may still be actively building.
//
// This is the boundary condition: stale healing must only fire after a session
// has been "running" longer than the configured staleness threshold.
//
// GREEN / ASSERTION: Currently this passes trivially because no healing occurs
// at all. After the fix it must STILL pass (i.e., the fix must be time-gated).
func TestBuildStatus_DoesNotHealRecentRunningSession(t *testing.T) {
	store := db.NewMockDataStore()
	// Session started 30 seconds ago -- definitely not stale.
	sessionID := buildStaleInterruptedSession(t, store, 30*time.Second)

	session, err := store.GetBuildSession(sessionID)
	require.NoError(t, err)
	require.NotNil(t, session)

	_ = renderBuildSession(store, session)

	stillRunning, err := store.GetBuildSession(sessionID)
	require.NoError(t, err)
	require.NotNil(t, stillRunning)

	assert.Equal(t, "running", stillRunning.Status,
		"A session started 30s ago with 'building' workspaces must NOT be healed "+
			"to 'interrupted' -- it may be actively building. "+
			"The stale-heal threshold must be at least several minutes.")
}
