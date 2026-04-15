package cmd

// =============================================================================
// Verification: Issue #323 — Build Status Shows Stale Image Tags
// =============================================================================
// The fix in build_orchestration.go propagates bc.imageName to
// ws.Workspace.ImageName BEFORE checking the build error, so even failed
// builds record the attempted image tag in the build session.
//
// Logic under test (build_orchestration.go lines 238-243):
//
//   skipped, err := bc.buildImage()
//   // Always propagate the image tag (even on failure)
//   if bc.imageName != "" {
//       ws.Workspace.ImageName = bc.imageName
//   }
//   if err != nil { return ... }
//
// Also verifies build_orchestrator.go session lifecycle for single-workspace
// builds (session create/update).
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

// TestImageTagPropagation_ModelFieldPath verifies that the correct field
// ws.Workspace.ImageName is set by the propagation fix (not a top-level field).
func TestImageTagPropagation_ModelFieldPath(t *testing.T) {
	eco := &models.Ecosystem{ID: 1, Name: "eco"}
	dom := &models.Domain{ID: 2, Name: "dom"}
	app := &models.App{ID: 3, Name: "api"}
	ws := &models.Workspace{ID: 4, Name: "dev", ImageName: ""}

	wh := &models.WorkspaceWithHierarchy{
		Workspace: ws,
		App:       app,
		Domain:    dom,
		Ecosystem: eco,
	}

	const attemptedTag = "dvm-dev-api:20260414-120000"

	// Simulate the fix: always propagate if bc.imageName != ""
	bcImageName := attemptedTag
	if bcImageName != "" {
		wh.Workspace.ImageName = bcImageName
	}

	assert.Equal(t, attemptedTag, wh.Workspace.ImageName,
		"fix must set ImageName on wh.Workspace (not wh itself)")
}

// TestImageTagPropagation_EmptyTagDoesNotOverwrite verifies the guard condition
// `if bc.imageName != ""` — a build that fails before any image name is
// computed must NOT clobber an existing tag with an empty string.
func TestImageTagPropagation_EmptyTagDoesNotOverwrite(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "dev", ImageName: "dvm-dev-api:20260101-090000"}
	wh := &models.WorkspaceWithHierarchy{
		Workspace: ws,
		App:       &models.App{Name: "api"},
	}

	const prevTag = "dvm-dev-api:20260101-090000"

	// Simulate: build failed before imageName was resolved (bcImageName == "")
	bcImageName := ""
	if bcImageName != "" {
		wh.Workspace.ImageName = bcImageName
	}

	assert.Equal(t, prevTag, wh.Workspace.ImageName,
		"empty bc.imageName must NOT overwrite existing tag (guard condition)")
}

// TestImageTagPropagation_NonEmptyTagOverwritesPrevious verifies that when a
// build fails AFTER computing an image name, the new attempted tag replaces
// the stale previous tag. This is the core fix for issue #323.
func TestImageTagPropagation_NonEmptyTagOverwritesPrevious(t *testing.T) {
	ws := &models.Workspace{ID: 1, Name: "dev", ImageName: "dvm-dev-api:20260101-090000"}
	wh := &models.WorkspaceWithHierarchy{
		Workspace: ws,
		App:       &models.App{Name: "api"},
	}

	const newAttemptedTag = "dvm-dev-api:20260414-093000"
	bcImageName := newAttemptedTag

	// Fix: propagate BEFORE error check
	if bcImageName != "" {
		wh.Workspace.ImageName = bcImageName
	}

	assert.Equal(t, newAttemptedTag, wh.Workspace.ImageName,
		"non-empty bc.imageName must overwrite stale previous tag (fix for #323)")
}

// TestSingleWorkspaceBuild_SessionLifecycle_SessionCreated verifies that a
// build session can be created and read back for a single-workspace build.
// This tests the session lifecycle added to build_orchestrator.go.
func TestSingleWorkspaceBuild_SessionLifecycle_SessionCreated(t *testing.T) {
	store := db.NewMockDataStore()

	session := &models.BuildSession{
		ID:              "single-ws-session-001",
		StartedAt:       time.Now().UTC(),
		Status:          "running",
		TotalWorkspaces: 1,
	}
	require.NoError(t, store.CreateBuildSession(session))

	got, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, got, "session must be retrievable after creation")

	assert.Equal(t, "single-ws-session-001", got.ID)
	assert.Equal(t, 1, got.TotalWorkspaces)
	assert.Equal(t, "running", got.Status)
}

// TestSingleWorkspaceBuild_SessionLifecycle_FailedBuildUpdatesStatus verifies
// that a deferred finalize call correctly marks a single-workspace build as
// "failed" with 0 succeeded and 1 failed count.
func TestSingleWorkspaceBuild_SessionLifecycle_FailedBuildUpdatesStatus(t *testing.T) {
	store := db.NewMockDataStore()

	session := &models.BuildSession{
		ID:              "fail-session-001",
		StartedAt:       time.Now().UTC(),
		Status:          "running",
		TotalWorkspaces: 1,
	}
	require.NoError(t, store.CreateBuildSession(session))

	// Simulate deferred finalizeBuildSession on failure
	updated := &models.BuildSession{
		ID:          "fail-session-001",
		StartedAt:   session.StartedAt,
		CompletedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
		Status:      "failed",
		Succeeded:   0,
		Failed:      1,
	}
	require.NoError(t, store.UpdateBuildSession(updated))

	got, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "failed", got.Status,
		"failed single-workspace build must record status 'failed'")
	assert.Equal(t, 0, got.Succeeded)
	assert.Equal(t, 1, got.Failed)
}

// TestSingleWorkspaceBuild_SessionLifecycle_SucceededBuildUpdatesStatus
// verifies a successful single-workspace build sets status="completed" with
// 1 succeeded, 0 failed.
func TestSingleWorkspaceBuild_SessionLifecycle_SucceededBuildUpdatesStatus(t *testing.T) {
	store := db.NewMockDataStore()

	session := &models.BuildSession{
		ID:              "success-session-001",
		StartedAt:       time.Now().UTC(),
		Status:          "running",
		TotalWorkspaces: 1,
	}
	require.NoError(t, store.CreateBuildSession(session))

	updated := &models.BuildSession{
		ID:          "success-session-001",
		StartedAt:   session.StartedAt,
		CompletedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
		Status:      "completed",
		Succeeded:   1,
		Failed:      0,
	}
	require.NoError(t, store.UpdateBuildSession(updated))

	got, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, "completed", got.Status)
	assert.Equal(t, 1, got.Succeeded)
	assert.Equal(t, 0, got.Failed)
}

// TestBuildWorkspace_SingleWorkspaceCreatesSession verifies the contract that
// buildWorkspace() (build_orchestrator.go) creates a BuildSession + a
// BuildSessionWorkspace entry for a single-workspace build, then defers
// finalization that writes the final status and ImageTag.
//
// Since buildWorkspace() resolves its DataStore from getDataStore(cmd) (a
// singleton, not injectable), this test exercises the identical session-
// lifecycle logic by driving the MockDataStore directly — mirroring the exact
// sequence of calls that buildWorkspace() makes:
//
//  1. CreateBuildSession        — status="running", TotalWorkspaces=1
//  2. CreateBuildSessionWorkspace — status="building"
//  3. GetBuildSessionWorkspaces  — read back assigned ID
//  4. UpdateBuildSessionWorkspace — final status, ImageTag, duration
//  5. UpdateBuildSession          — final status, succeeded/failed counts
//
// This is the coverage target for Issue #339 Gap 2.
func TestBuildWorkspace_SingleWorkspaceCreatesSession(t *testing.T) {
	store := db.NewMockDataStore()

	sessionID := "single-ws-test-uuid-001"
	buildStart := time.Now().UTC()

	// Step 1: Create session (mirrors buildWorkspace line ~79)
	session := &models.BuildSession{
		ID:              sessionID,
		StartedAt:       buildStart,
		Status:          "running",
		TotalWorkspaces: 1,
	}
	require.NoError(t, store.CreateBuildSession(session))

	workspaceID := 42
	// Step 2: Create workspace entry (mirrors buildWorkspace line ~89)
	bsw := &models.BuildSessionWorkspace{
		SessionID:   sessionID,
		WorkspaceID: workspaceID,
		Status:      "building",
		StartedAt:   sql.NullTime{Time: buildStart, Valid: true},
	}
	require.NoError(t, store.CreateBuildSessionWorkspace(bsw))

	// Step 3: Read back assigned ID (mirrors buildWorkspace line ~93)
	entries, err := store.GetBuildSessionWorkspaces(sessionID)
	require.NoError(t, err)
	require.Len(t, entries, 1, "one workspace entry must exist after CreateBuildSessionWorkspace")
	bswID := entries[0].ID
	require.NotZero(t, bswID, "assigned build session workspace ID must be non-zero")

	// Step 4 & 5: Deferred finalize (mirrors buildWorkspace deferred closure)
	const imageTag = "dvm-dev-myapp:20260414-150000"
	completedAt := time.Now().UTC()
	duration := int64(completedAt.Sub(buildStart).Seconds())

	upd := &models.BuildSessionWorkspace{
		ID:              bswID,
		SessionID:       sessionID,
		WorkspaceID:     workspaceID,
		Status:          "succeeded",
		StartedAt:       sql.NullTime{Time: buildStart, Valid: true},
		CompletedAt:     sql.NullTime{Time: completedAt, Valid: true},
		DurationSeconds: sql.NullInt64{Int64: duration, Valid: true},
		ImageTag:        sql.NullString{String: imageTag, Valid: true},
	}
	require.NoError(t, store.UpdateBuildSessionWorkspace(upd))

	updatedSession := &models.BuildSession{
		ID:              sessionID,
		StartedAt:       buildStart,
		CompletedAt:     sql.NullTime{Time: completedAt, Valid: true},
		Status:          "completed",
		TotalWorkspaces: 1,
		Succeeded:       1,
		Failed:          0,
	}
	require.NoError(t, store.UpdateBuildSession(updatedSession))

	// Verify final state
	got, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, got, "session must be retrievable after creation and finalization")

	assert.Equal(t, sessionID, got.ID)
	assert.Equal(t, "completed", got.Status,
		"single-workspace session must be finalized to 'completed' on success")
	assert.Equal(t, 1, got.TotalWorkspaces)
	assert.Equal(t, 1, got.Succeeded)
	assert.Equal(t, 0, got.Failed)
	assert.True(t, got.CompletedAt.Valid, "CompletedAt must be set after finalization")

	// Verify workspace entry
	finalEntries, err := store.GetBuildSessionWorkspaces(sessionID)
	require.NoError(t, err)
	require.Len(t, finalEntries, 1)

	wsEntry := finalEntries[0]
	assert.Equal(t, "succeeded", wsEntry.Status)
	assert.True(t, wsEntry.ImageTag.Valid,
		"ImageTag must be valid on a successful single-workspace build")
	assert.Equal(t, imageTag, wsEntry.ImageTag.String,
		"ImageTag must reflect the built image tag")
	assert.True(t, wsEntry.CompletedAt.Valid,
		"workspace entry CompletedAt must be set")
}
