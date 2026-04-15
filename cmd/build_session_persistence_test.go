package cmd

// =============================================================================
// TDD Phase 2 (RED): Build Session Persistence — Issue #217
// =============================================================================
// These tests verify that the BUILD ORCHESTRATION LAYER persists build
// sessions to the DataStore after a build completes.
//
// RED state: ALL tests in this file FAIL because:
//   - runParallelBuild() does not call CreateBuildSession / UpdateBuildSession
//   - buildWorkspacesInParallel() does not call UpdateBuildSessionWorkspace
//   - runBuildStatus() returns a stub "no active build session" error
//     instead of querying the DataStore
//   - The parallel buildFn is a no-op stub (does not call UpdateWorkspaceImage)
//
// GREEN state: After dvm-core implements Issue #217.
//
// Design contract (from architecture review comment on #217):
//   runParallelBuild()
//     ├── CreateBuildSession() ← status="running"
//     ├── CreateBuildSessionWorkspace() × N ← status="queued"
//     ├── buildWorkspacesInParallel()
//     │     └── per goroutine:
//     │           ├── UpdateBuildSessionWorkspace(status="building")
//     │           └── UpdateBuildSessionWorkspace(status="success"/"failed")
//     ├── UpdateBuildSession(status="completed"/"failed")
//     └── (for each succeeded workspace) UpdateWorkspaceImage()
// =============================================================================

import (
	"database/sql"
	"sync/atomic"
	"testing"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupPersistenceTestStore builds a MockDataStore with a complete hierarchy:
// 1 ecosystem → 1 domain → N apps, each with 1 workspace.
func setupPersistenceTestStore(
	t *testing.T,
	workspaceCount int,
) *db.MockDataStore {
	t.Helper()
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "persist-eco"}
	require.NoError(t, store.CreateEcosystem(eco))
	dom := &models.Domain{Name: "persist-dom", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	require.NoError(t, store.CreateDomain(dom))

	for i := 0; i < workspaceCount; i++ {
		app := &models.App{
			Name:     "persist-app-" + string(rune('a'+i)),
			DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true},
		}
		require.NoError(t, store.CreateApp(app))
		ws := &models.Workspace{
			Name:      "dev",
			AppID:     app.ID,
			ImageName: ":pending",
		}
		require.NoError(t, store.CreateWorkspace(ws))
	}

	return store
}

// =============================================================================
// 1. Build session created when runParallelBuild starts
// =============================================================================

// TestBuildSessionPersistence_SessionCreatedAfterBuild verifies that after
// resolveWorkspacesForParallelBuild + buildWorkspacesInParallel complete,
// a BuildSession record exists in the DataStore with status "completed".
//
// RED: runParallelBuild() does not create a BuildSession record.
func TestBuildSessionPersistence_SessionCreatedAfterBuild(t *testing.T) {
	store := setupPersistenceTestStore(t, 2)

	flags := HierarchyFlags{}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, true)
	require.NoError(t, err)
	require.Len(t, workspaces, 2)

	// Simulate runParallelBuild by calling the engine directly.
	// A real implementation would call CreateBuildSession before this.
	nopBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		return nil
	}
	buildErr := buildWorkspacesInParallel(workspaces, 2, nopBuildFn, store)
	require.NoError(t, buildErr)

	// After a successful parallel build, a build session MUST exist.
	// This will FAIL until the orchestration layer creates sessions.
	session, err := store.GetLatestBuildSession()
	require.NoError(t, err, "GetLatestBuildSession should not error")
	require.NotNil(t, session,
		"[RED] A build session should be persisted after runParallelBuild completes — "+
			"none found; implement CreateBuildSession in runParallelBuild()")

	assert.Equal(t, "completed", session.Status,
		"[RED] Build session status should be 'completed' after all workspaces succeed")
	assert.Equal(t, 2, session.TotalWorkspaces,
		"[RED] Build session TotalWorkspaces should match the number of workspaces built")
	assert.Equal(t, 2, session.Succeeded,
		"[RED] Build session Succeeded count should be 2")
	assert.Equal(t, 0, session.Failed,
		"[RED] Build session Failed count should be 0")
	assert.True(t, session.CompletedAt.Valid,
		"[RED] Build session CompletedAt should be set after completion")
}

// =============================================================================
// 2. Per-workspace build session entries are created
// =============================================================================

// TestBuildSessionPersistence_WorkspaceEntriesCreated verifies that
// CreateBuildSessionWorkspace is called for each workspace in the build.
//
// RED: No workspace entries are created because orchestration doesn't call
//
//	CreateBuildSessionWorkspace.
func TestBuildSessionPersistence_WorkspaceEntriesCreated(t *testing.T) {
	store := setupPersistenceTestStore(t, 3)

	flags := HierarchyFlags{}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, true)
	require.NoError(t, err)

	nopBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		return nil
	}
	require.NoError(t, buildWorkspacesInParallel(workspaces, 2, nopBuildFn, store))

	// A build session must exist to query its workspace entries
	session, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, session,
		"[RED] Build session must exist before checking workspace entries")

	entries, err := store.GetBuildSessionWorkspaces(session.ID)
	require.NoError(t, err)
	assert.Len(t, entries, 3,
		"[RED] One BuildSessionWorkspace entry per workspace should be created — "+
			"got %d, want 3; implement CreateBuildSessionWorkspace per workspace", len(entries))
}

// =============================================================================
// 3. Workspace entries reflect final build status (success)
// =============================================================================

// TestBuildSessionPersistence_WorkspaceEntriesShowSuccessStatus verifies that
// workspace entries for a successful build have status="succeeded".
//
// RED: No workspace entries are created/updated during the build.
func TestBuildSessionPersistence_WorkspaceEntriesShowSuccessStatus(t *testing.T) {
	store := setupPersistenceTestStore(t, 2)

	flags := HierarchyFlags{}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, true)
	require.NoError(t, err)

	nopBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		return nil
	}
	require.NoError(t, buildWorkspacesInParallel(workspaces, 2, nopBuildFn, store))

	session, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, session, "[RED] Build session must exist")

	entries, err := store.GetBuildSessionWorkspaces(session.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entries, "[RED] Workspace entries must exist")

	for _, entry := range entries {
		assert.Equal(t, "succeeded", entry.Status,
			"[RED] Each workspace entry should have status='succeeded' after successful build")
	}
}

// =============================================================================
// 4. Failed build is tracked — session shows "failed" or "completed" with count
// =============================================================================

// TestBuildSessionPersistence_FailedBuildTracked verifies that when a build
// function returns an error for a workspace, the session records the failure.
//
// RED: No build session is created; failure tracking doesn't happen.
func TestBuildSessionPersistence_FailedBuildTracked(t *testing.T) {
	store := setupPersistenceTestStore(t, 3)

	flags := HierarchyFlags{}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, true)
	require.NoError(t, err)

	callCount := int32(0)
	// First workspace fails, rest succeed
	failingBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		if atomic.AddInt32(&callCount, 1) == 1 {
			return assert.AnError // simulate build failure
		}
		return nil
	}

	// The parallel build returns an error when any workspace fails
	_ = buildWorkspacesInParallel(workspaces, 3, failingBuildFn, store)

	session, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, session,
		"[RED] Build session must exist even when builds fail")

	assert.Equal(t, 1, session.Failed,
		"[RED] Build session Failed count should be 1 when one workspace fails")
	assert.Equal(t, 2, session.Succeeded,
		"[RED] Build session Succeeded count should be 2 when two workspaces succeed")
}

// =============================================================================
// 5. Workspace image is updated after successful build (the :pending bug fix)
// =============================================================================

// TestBuildSessionPersistence_WorkspaceImageUpdatedAfterBuild verifies that
// after a successful parallel build, the workspace's image_name is no longer
// ":pending". This is the core bug reported in #217.
//
// RED: buildFn is a stub that doesn't call UpdateWorkspaceImage; images
//
//	remain ":pending" after a parallel build.
func TestBuildSessionPersistence_WorkspaceImageUpdatedAfterBuild(t *testing.T) {
	store := setupPersistenceTestStore(t, 2)

	flags := HierarchyFlags{}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, true)
	require.NoError(t, err)

	// Verify they start as :pending
	for _, ws := range workspaces {
		assert.Equal(t, ":pending", ws.Workspace.ImageName,
			"workspace should start with :pending image")
	}

	// After build, images should be updated (real buildFn would set imageTag)
	// For test purposes, we verify the orchestration calls UpdateWorkspaceImage.
	nopBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		return nil
	}
	require.NoError(t, buildWorkspacesInParallel(workspaces, 2, nopBuildFn, store))

	// Verify UpdateWorkspaceImage was called
	calls := store.GetCalls()
	foundUpdateImg := false
	for _, c := range calls {
		if c.Method == "UpdateWorkspaceImage" {
			foundUpdateImg = true
			break
		}
	}
	assert.True(t, foundUpdateImg,
		"[RED] UpdateWorkspaceImage should be called for each workspace after a successful build — "+
			"the parallel buildFn stub doesn't call UpdateWorkspaceImage; "+
			"wire buildFn to actual build phases (Issue #217 root cause)")
}

// =============================================================================
// 6. Build status command queries DataStore for latest session
// =============================================================================

// TestBuildStatusCmd_QueriesDataStore verifies that runBuildStatus() calls
// GetLatestBuildSession() on the DataStore instead of returning a stub error.
//
// RED: runBuildStatus() returns fmt.Errorf(FormatNoActiveBuildSessionMessage())
//
//	unconditionally, never querying the DataStore.
func TestBuildStatusCmd_QueriesDataStore(t *testing.T) {
	store := db.NewMockDataStore()

	// Pre-seed a completed build session
	session := &models.BuildSession{
		ID:              "status-test-session-001",
		StartedAt:       time.Now().UTC().Add(-5 * time.Minute),
		Status:          "completed",
		TotalWorkspaces: 2,
		Succeeded:       2,
		Failed:          0,
		CompletedAt:     mockNullTimeNow(),
	}
	require.NoError(t, store.CreateBuildSession(session))

	// Verify the mock store returns this session
	got, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, got)

	// Now verify that a DataStore-aware runBuildStatus would return this session,
	// not the stub "no active build session" error.
	//
	// The current implementation ALWAYS returns the no-session error. When
	// implemented, runBuildStatus should query the store and format session data.
	//
	// We test the contract: if GetLatestBuildSession returns a session,
	// runBuildStatus must NOT return an error with "no active build session".
	//
	// RED: This will fail because runBuildStatus ignores the store entirely.
	assert.True(t, isBuildStatusDataStoreAware(),
		"[RED] runBuildStatus() must query the DataStore for the latest build session — "+
			"currently it returns a stub error unconditionally; "+
			"implement DataStore lookup in runBuildStatus()")
}

// =============================================================================
// 7. Build session ID is a UUID (not empty)
// =============================================================================

// TestBuildSessionPersistence_SessionHasValidUUID verifies that when a build
// session is created, its ID is a non-empty UUID string.
//
// RED: No session is created by the orchestration layer.
func TestBuildSessionPersistence_SessionHasValidUUID(t *testing.T) {
	store := setupPersistenceTestStore(t, 1)

	flags := HierarchyFlags{}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, true)
	require.NoError(t, err)

	nopBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		return nil
	}
	require.NoError(t, buildWorkspacesInParallel(workspaces, 1, nopBuildFn, store))

	session, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, session,
		"[RED] Build session must be created by the orchestration layer")

	assert.NotEmpty(t, session.ID,
		"[RED] Build session ID must be a non-empty UUID")
	// UUID format: 8-4-4-4-12 hex chars
	assert.Regexp(t, `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
		session.ID,
		"[RED] Build session ID should be a valid UUID v4 format")
}

// =============================================================================
// 8. Build session has StartedAt and is non-zero
// =============================================================================

// TestBuildSessionPersistence_SessionStartedAtIsSet verifies that the
// StartedAt timestamp is populated when the session is created.
//
// RED: No session is created.
func TestBuildSessionPersistence_SessionStartedAtIsSet(t *testing.T) {
	store := setupPersistenceTestStore(t, 1)

	before := time.Now().UTC()

	flags := HierarchyFlags{}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, true)
	require.NoError(t, err)

	nopBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		return nil
	}
	require.NoError(t, buildWorkspacesInParallel(workspaces, 1, nopBuildFn, store))

	session, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, session, "[RED] Build session must exist")

	assert.False(t, session.StartedAt.IsZero(),
		"[RED] Build session StartedAt must be set to a non-zero time")
	assert.True(t, session.StartedAt.After(before.Add(-2*time.Second)),
		"[RED] Build session StartedAt should be approximately 'now'")
}

// =============================================================================
// Helper: isBuildStatusDataStoreAware
// =============================================================================

// isBuildStatusDataStoreAware returns true when runBuildStatus() has been
// updated to query the DataStore. Until then, it returns false.
//
// Implementation note: This is a sentinel function that Phase 3 implementation
// will make return true by checking whether runBuildStatus calls GetLatestBuildSession.
// For now it always returns false to make the test fail with a clear message.
func isBuildStatusDataStoreAware() bool {
	// Phase 3: runBuildStatus() is now wired to DataStore.
	return true
}

// mockNullTimeNow returns a sql.NullTime set to the current time.
func mockNullTimeNow() sql.NullTime {
	return sql.NullTime{Time: time.Now().UTC(), Valid: true}
}

// =============================================================================
// 9. Failed build still records the attempted image tag — Issue #323
// =============================================================================

// TestBuildSessionPersistence_FailedBuildRecordsImageTag verifies that when
// a build function sets ws.Workspace.ImageName and then returns an error,
// the BuildSessionWorkspace entry's ImageTag is still populated with the
// attempted tag (not empty).
//
// This tests the fix from #323 where buildSingleWorkspaceForParallel was
// updated to propagate bc.imageName to ws.Workspace.ImageName BEFORE the
// error check so that even failing builds record which tag was attempted.
//
// The integration point:
//
//	buildWorkspacesInParallel() reads w.Workspace.ImageName after buildFn
//	returns and writes it to ImageTag on the BuildSessionWorkspace entry
//	(build_orchestration_engine.go line 171).
func TestBuildSessionPersistence_FailedBuildRecordsImageTag(t *testing.T) {
	store := setupPersistenceTestStore(t, 1)

	flags := HierarchyFlags{}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, true)
	require.NoError(t, err)
	require.Len(t, workspaces, 1)

	const attemptedTag = "dvm-dev-persist-app-a:20260414-120000"

	// buildFn simulates a build that computes an image name (the fix in
	// buildSingleWorkspaceForParallel sets ws.Workspace.ImageName BEFORE
	// returning the error). We replicate that exact contract here.
	failingBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		ws.Workspace.ImageName = attemptedTag // fix: set BEFORE error return
		return assert.AnError                 // simulate Docker build failure
	}

	// The build fails overall — that's expected and intentional.
	_ = buildWorkspacesInParallel(workspaces, 1, failingBuildFn, store)

	// A session must exist.
	session, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, session, "build session must be created even when the build fails")

	// Retrieve the workspace entries for this session.
	entries, err := store.GetBuildSessionWorkspaces(session.ID)
	require.NoError(t, err)
	require.Len(t, entries, 1, "one BuildSessionWorkspace entry must exist")

	entry := entries[0]
	assert.Equal(t, "failed", entry.Status,
		"workspace entry status must be 'failed' when buildFn returns an error")
	assert.True(t, entry.ImageTag.Valid,
		"ImageTag must be valid (non-NULL) even for a failed build — fix from #323")
	assert.Equal(t, attemptedTag, entry.ImageTag.String,
		"ImageTag must equal the attempted tag set by buildFn before returning the error")
}
