package db

// =============================================================================
// TDD Phase 2 (GREEN): Build Session DB Layer — Issue #217
// =============================================================================
// These tests verify the SQLite implementation of BuildSessionStore.
// Phase 1 (database agent) has already implemented the store methods, so
// these tests are expected to PASS and serve as a regression harness.
//
// Covered operations:
//   (a) CreateBuildSession / GetBuildSession / GetLatestBuildSession
//   (b) UpdateBuildSession
//   (c) GetBuildSessions (list with limit)
//   (d) DeleteBuildSessionsOlderThan (cascade)
//   (e) BuildSessionWorkspace CRUD
//   (f) GetBuildSessionStats aggregation
//   (g) UpdateWorkspaceImage targeted update
//   (h) Edge cases: empty DB, not-found, FK cascade
// =============================================================================

import (
	"database/sql"
	"testing"
	"time"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// newTestSession creates a BuildSession suitable for test inserts.
func newTestSession(id, status string, totalWorkspaces int) *models.BuildSession {
	return &models.BuildSession{
		ID:              id,
		StartedAt:       time.Now().UTC().Truncate(time.Second),
		Status:          status,
		TotalWorkspaces: totalWorkspaces,
		Succeeded:       0,
		Failed:          0,
	}
}

// createTestWorkspaceForSession creates a full hierarchy and workspace for FK use.
func createTestWorkspaceForSession(t *testing.T, ds *SQLDataStore, suffix string) *models.Workspace {
	t.Helper()
	app := createTestApp(t, ds, "bsws-"+suffix)
	ws := &models.Workspace{
		AppID:     app.ID,
		Name:      "bsws-" + suffix,
		Slug:      "eco-dom-app-bsws-" + suffix,
		ImageName: ":pending",
		Status:    "stopped",
	}
	require.NoError(t, ds.CreateWorkspace(ws), "setup: create workspace")
	return ws
}

// =============================================================================
// (a) CreateBuildSession / GetBuildSession / GetLatestBuildSession
// =============================================================================

func TestSQLDataStore_CreateBuildSession_SetsID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	session := newTestSession("test-session-001", "running", 3)
	require.NoError(t, ds.CreateBuildSession(session))
	assert.Equal(t, "test-session-001", session.ID)
}

func TestSQLDataStore_GetBuildSession_RoundTrip(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	orig := newTestSession("roundtrip-001", "running", 5)
	require.NoError(t, ds.CreateBuildSession(orig))

	got, err := ds.GetBuildSession("roundtrip-001")
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "roundtrip-001", got.ID)
	assert.Equal(t, "running", got.Status)
	assert.Equal(t, 5, got.TotalWorkspaces)
	assert.Equal(t, 0, got.Succeeded)
	assert.Equal(t, 0, got.Failed)
}

func TestSQLDataStore_GetBuildSession_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	_, err := ds.GetBuildSession("does-not-exist")
	require.Error(t, err, "GetBuildSession for missing ID should return error")
}

func TestSQLDataStore_GetLatestBuildSession_EmptyDB(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	got, err := ds.GetLatestBuildSession()
	require.NoError(t, err, "GetLatestBuildSession on empty DB should return nil, nil")
	assert.Nil(t, got, "empty DB should return nil session")
}

func TestSQLDataStore_GetLatestBuildSession_ReturnsMostRecent(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	older := newTestSession("older-001", "completed", 2)
	older.StartedAt = time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	require.NoError(t, ds.CreateBuildSession(older))

	newer := newTestSession("newer-001", "running", 4)
	newer.StartedAt = time.Now().UTC().Add(-1 * time.Minute).Truncate(time.Second)
	require.NoError(t, ds.CreateBuildSession(newer))

	got, err := ds.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "newer-001", got.ID,
		"GetLatestBuildSession should return the session with the most recent started_at")
}

// =============================================================================
// (b) UpdateBuildSession
// =============================================================================

func TestSQLDataStore_UpdateBuildSession_UpdatesStatus(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	session := newTestSession("update-001", "running", 3)
	require.NoError(t, ds.CreateBuildSession(session))

	session.Status = "completed"
	session.Succeeded = 3
	session.Failed = 0
	session.CompletedAt = sql.NullTime{Time: time.Now().UTC().Truncate(time.Second), Valid: true}
	require.NoError(t, ds.UpdateBuildSession(session))

	got, err := ds.GetBuildSession("update-001")
	require.NoError(t, err)
	assert.Equal(t, "completed", got.Status)
	assert.Equal(t, 3, got.Succeeded)
	assert.True(t, got.CompletedAt.Valid)
}

func TestSQLDataStore_UpdateBuildSession_NotFound_ReturnsError(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ghost := &models.BuildSession{
		ID:     "does-not-exist",
		Status: "completed",
	}
	err := ds.UpdateBuildSession(ghost)
	require.Error(t, err, "UpdateBuildSession on missing session should return error")
}

// =============================================================================
// (c) GetBuildSessions (list with limit)
// =============================================================================

func TestSQLDataStore_GetBuildSessions_RespectsLimit(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	for i := 0; i < 5; i++ {
		s := newTestSession(
			"list-session-"+string(rune('a'+i)),
			"completed",
			1,
		)
		s.StartedAt = time.Now().UTC().Add(time.Duration(-i) * time.Minute).Truncate(time.Second)
		require.NoError(t, ds.CreateBuildSession(s))
	}

	sessions, err := ds.GetBuildSessions(3)
	require.NoError(t, err)
	assert.Len(t, sessions, 3, "GetBuildSessions(3) should return at most 3 sessions")
}

func TestSQLDataStore_GetBuildSessions_EmptyDB(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	sessions, err := ds.GetBuildSessions(10)
	require.NoError(t, err)
	assert.Empty(t, sessions)
}

func TestSQLDataStore_GetBuildSessions_OrderedByStartedAtDesc(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	s1 := newTestSession("ordered-oldest", "completed", 1)
	s1.StartedAt = time.Now().UTC().Add(-3 * time.Hour).Truncate(time.Second)
	require.NoError(t, ds.CreateBuildSession(s1))

	s2 := newTestSession("ordered-middle", "completed", 1)
	s2.StartedAt = time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	require.NoError(t, ds.CreateBuildSession(s2))

	s3 := newTestSession("ordered-newest", "running", 1)
	s3.StartedAt = time.Now().UTC().Add(-1 * time.Minute).Truncate(time.Second)
	require.NoError(t, ds.CreateBuildSession(s3))

	sessions, err := ds.GetBuildSessions(10)
	require.NoError(t, err)
	require.Len(t, sessions, 3)
	assert.Equal(t, "ordered-newest", sessions[0].ID,
		"first result should be newest session")
}

// =============================================================================
// (d) DeleteBuildSessionsOlderThan (cascade)
// =============================================================================

func TestSQLDataStore_DeleteBuildSessionsOlderThan_DeletesMatching(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	old := newTestSession("old-delete-001", "completed", 1)
	old.StartedAt = time.Now().UTC().Add(-48 * time.Hour).Truncate(time.Second)
	require.NoError(t, ds.CreateBuildSession(old))

	recent := newTestSession("recent-keep-001", "completed", 1)
	recent.StartedAt = time.Now().UTC().Add(-1 * time.Hour).Truncate(time.Second)
	require.NoError(t, ds.CreateBuildSession(recent))

	cutoff := time.Now().UTC().Add(-24 * time.Hour)
	count, err := ds.DeleteBuildSessionsOlderThan(cutoff)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "should delete exactly 1 old session")

	_, err = ds.GetBuildSession("old-delete-001")
	assert.Error(t, err, "old session should be gone after deletion")

	kept, err := ds.GetBuildSession("recent-keep-001")
	require.NoError(t, err)
	assert.NotNil(t, kept, "recent session should still exist")
}

func TestSQLDataStore_DeleteBuildSessionsOlderThan_NoMatches(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	s := newTestSession("no-delete-001", "completed", 1)
	s.StartedAt = time.Now().UTC()
	require.NoError(t, ds.CreateBuildSession(s))

	cutoff := time.Now().UTC().Add(-48 * time.Hour)
	count, err := ds.DeleteBuildSessionsOlderThan(cutoff)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count, "should delete 0 sessions when none are old enough")
}

// =============================================================================
// (e) BuildSessionWorkspace CRUD
// =============================================================================

func TestSQLDataStore_CreateBuildSessionWorkspace_SetsID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	session := newTestSession("bsw-create-001", "running", 1)
	require.NoError(t, ds.CreateBuildSession(session))

	ws := createTestWorkspaceForSession(t, ds, "create001")

	bsw := &models.BuildSessionWorkspace{
		SessionID:   session.ID,
		WorkspaceID: ws.ID,
		Status:      "queued",
	}
	require.NoError(t, ds.CreateBuildSessionWorkspace(bsw))
	assert.NotZero(t, bsw.ID, "CreateBuildSessionWorkspace should set the auto-increment ID")
}

func TestSQLDataStore_UpdateBuildSessionWorkspace_UpdatesStatus(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	session := newTestSession("bsw-update-001", "running", 1)
	require.NoError(t, ds.CreateBuildSession(session))
	ws := createTestWorkspaceForSession(t, ds, "update001")

	bsw := &models.BuildSessionWorkspace{
		SessionID:   session.ID,
		WorkspaceID: ws.ID,
		Status:      "queued",
	}
	require.NoError(t, ds.CreateBuildSessionWorkspace(bsw))

	bsw.Status = "succeeded"
	bsw.ImageTag = sql.NullString{String: "myapp:abc123", Valid: true}
	bsw.DurationSeconds = sql.NullInt64{Int64: 45, Valid: true}
	require.NoError(t, ds.UpdateBuildSessionWorkspace(bsw))

	workspaces, err := ds.GetBuildSessionWorkspaces(session.ID)
	require.NoError(t, err)
	require.Len(t, workspaces, 1)
	assert.Equal(t, "succeeded", workspaces[0].Status)
	assert.Equal(t, "myapp:abc123", workspaces[0].ImageTag.String)
	assert.Equal(t, int64(45), workspaces[0].DurationSeconds.Int64)
}

func TestSQLDataStore_UpdateBuildSessionWorkspace_NotFound_ReturnsError(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ghost := &models.BuildSessionWorkspace{
		ID:     99999,
		Status: "succeeded",
	}
	err := ds.UpdateBuildSessionWorkspace(ghost)
	require.Error(t, err, "UpdateBuildSessionWorkspace on missing entry should return error")
}

func TestSQLDataStore_GetBuildSessionWorkspaces_ReturnsAllForSession(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	session := newTestSession("bsw-list-001", "running", 3)
	require.NoError(t, ds.CreateBuildSession(session))

	for i := 0; i < 3; i++ {
		ws := createTestWorkspaceForSession(t, ds, "list"+string(rune('a'+i)))
		bsw := &models.BuildSessionWorkspace{
			SessionID:   session.ID,
			WorkspaceID: ws.ID,
			Status:      "queued",
		}
		require.NoError(t, ds.CreateBuildSessionWorkspace(bsw))
	}

	workspaces, err := ds.GetBuildSessionWorkspaces(session.ID)
	require.NoError(t, err)
	assert.Len(t, workspaces, 3)
}

func TestSQLDataStore_GetBuildSessionWorkspaces_EmptyForUnknownSession(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	workspaces, err := ds.GetBuildSessionWorkspaces("no-such-session")
	require.NoError(t, err)
	assert.Empty(t, workspaces)
}

// =============================================================================
// (f) GetBuildSessionStats aggregation
// =============================================================================

func TestSQLDataStore_GetBuildSessionStats_Aggregates(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	session := newTestSession("stats-001", "completed", 4)
	require.NoError(t, ds.CreateBuildSession(session))

	statusMap := []string{"succeeded", "succeeded", "succeeded", "failed"}
	for i, status := range statusMap {
		ws := createTestWorkspaceForSession(t, ds, "stats"+string(rune('a'+i)))
		bsw := &models.BuildSessionWorkspace{
			SessionID:   session.ID,
			WorkspaceID: ws.ID,
			Status:      status,
		}
		require.NoError(t, ds.CreateBuildSessionWorkspace(bsw))
	}

	succeeded, failed, err := ds.GetBuildSessionStats(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 3, succeeded, "GetBuildSessionStats should count 3 succeeded")
	assert.Equal(t, 1, failed, "GetBuildSessionStats should count 1 failed")
}

func TestSQLDataStore_GetBuildSessionStats_EmptySession(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	session := newTestSession("stats-empty-001", "running", 0)
	require.NoError(t, ds.CreateBuildSession(session))

	succeeded, failed, err := ds.GetBuildSessionStats(session.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, succeeded)
	assert.Equal(t, 0, failed)
}

func TestSQLDataStore_GetBuildSessionStats_NoSuchSession(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// COALESCE means this returns 0,0 not an error
	succeeded, failed, err := ds.GetBuildSessionStats("ghost-session")
	require.NoError(t, err)
	assert.Equal(t, 0, succeeded)
	assert.Equal(t, 0, failed)
}

// =============================================================================
// (g) UpdateWorkspaceImage targeted update
// =============================================================================

func TestSQLDataStore_UpdateWorkspaceImage_UpdatesImageName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ws := createTestWorkspaceForSession(t, ds, "imgupdate001")
	require.Equal(t, ":pending", ws.ImageName)

	err := ds.UpdateWorkspaceImage(ws.ID, "myapp:abc123")
	require.NoError(t, err)

	got, err := ds.GetWorkspaceByID(ws.ID)
	require.NoError(t, err)
	assert.Equal(t, "myapp:abc123", got.ImageName,
		"UpdateWorkspaceImage should update the workspace image_name field")
}

func TestSQLDataStore_UpdateWorkspaceImage_NotFound_ReturnsError(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.UpdateWorkspaceImage(99999, "ghost:latest")
	require.Error(t, err, "UpdateWorkspaceImage on non-existent workspace should return error")
}

// =============================================================================
// (h) FK cascade: delete session cascades to workspace entries
// =============================================================================

func TestSQLDataStore_DeleteBuildSession_CascadesToWorkspaces(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	session := newTestSession("cascade-001", "completed", 2)
	require.NoError(t, ds.CreateBuildSession(session))

	for i := 0; i < 2; i++ {
		ws := createTestWorkspaceForSession(t, ds, "cascade"+string(rune('a'+i)))
		bsw := &models.BuildSessionWorkspace{
			SessionID:   session.ID,
			WorkspaceID: ws.ID,
			Status:      "succeeded",
		}
		require.NoError(t, ds.CreateBuildSessionWorkspace(bsw))
	}

	// Verify entries exist before deletion
	entries, err := ds.GetBuildSessionWorkspaces(session.ID)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	// Delete the session (should cascade)
	cutoff := time.Now().UTC().Add(1 * time.Second)
	count, err := ds.DeleteBuildSessionsOlderThan(cutoff)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Workspace entries should be gone due to cascade
	entries, err = ds.GetBuildSessionWorkspaces(session.ID)
	require.NoError(t, err)
	assert.Empty(t, entries, "cascade delete should remove build_session_workspaces entries")
}

// =============================================================================
// (i) Table-driven: multiple session lifecycle states
// =============================================================================

func TestSQLDataStore_BuildSession_StatusTransitions_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus string
		finalStatus   string
		succeeded     int
		failed        int
	}{
		{
			name:          "all success",
			initialStatus: "running",
			finalStatus:   "completed",
			succeeded:     3,
			failed:        0,
		},
		{
			name:          "partial failure",
			initialStatus: "running",
			finalStatus:   "completed",
			succeeded:     2,
			failed:        1,
		},
		{
			name:          "all failed",
			initialStatus: "running",
			finalStatus:   "failed",
			succeeded:     0,
			failed:        3,
		},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			id := "lifecycle-" + string(rune('a'+i))
			session := newTestSession(id, tt.initialStatus, tt.succeeded+tt.failed)
			require.NoError(t, ds.CreateBuildSession(session))

			// Transition to final status
			session.Status = tt.finalStatus
			session.Succeeded = tt.succeeded
			session.Failed = tt.failed
			session.CompletedAt = sql.NullTime{
				Time:  time.Now().UTC().Truncate(time.Second),
				Valid: true,
			}
			require.NoError(t, ds.UpdateBuildSession(session))

			got, err := ds.GetBuildSession(id)
			require.NoError(t, err)
			assert.Equal(t, tt.finalStatus, got.Status)
			assert.Equal(t, tt.succeeded, got.Succeeded)
			assert.Equal(t, tt.failed, got.Failed)
			assert.True(t, got.CompletedAt.Valid, "completed session should have completed_at set")
		})
	}
}
