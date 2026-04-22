package cmd

// =============================================================================
// TDD Phase 2 (RED): Build Session Interrupt Handling -- Issue #399
// =============================================================================
// These tests verify that when a build is interrupted (context cancelled /
// SIGINT), the orchestration engine must write "interrupted" status to the
// build_sessions and build_session_workspaces tables before returning.
//
// RED state: ALL assertions in this file FAIL because:
//   - buildWorkspacesInParallel does not accept a context.Context parameter
//   - There is no signal handler in main.go
//   - Defers do not fire on SIGINT, so session finalization is bypassed
//   - The engine maps all buildFn errors to "failed", not "interrupted"
//
// GREEN state: After dvm-core implements Issue #399:
//   - buildWorkspacesInParallel accepts a context.Context
//   - On ctx.Done(), non-terminal workspace rows are written "interrupted"
//   - The session row is written "interrupted" with completed_at = now
//   - Uses sync.Once so cleanup runs exactly once (normal or signal path)
// =============================================================================

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupInterruptTestStore creates a MockDataStore pre-seeded with N workspaces.
func setupInterruptTestStore(t *testing.T, count int) (*db.MockDataStore, []*models.WorkspaceWithHierarchy) {
	t.Helper()
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "interrupt-eco"}
	require.NoError(t, store.CreateEcosystem(eco))
	dom := &models.Domain{
		Name:        "interrupt-dom",
		EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true},
	}
	require.NoError(t, store.CreateDomain(dom))

	var workspaces []*models.WorkspaceWithHierarchy
	for i := 0; i < count; i++ {
		app := &models.App{
			Name:     "interrupt-app",
			DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true},
		}
		require.NoError(t, store.CreateApp(app))
		ws := &models.Workspace{
			Name:      "dev",
			AppID:     app.ID,
			ImageName: ":pending",
		}
		require.NoError(t, store.CreateWorkspace(ws))
		workspaces = append(workspaces, &models.WorkspaceWithHierarchy{
			Workspace: ws,
			App:       app,
			Domain:    dom,
			Ecosystem: eco,
		})
	}
	return store, workspaces
}

// =============================================================================
// 1. Session marked "interrupted" when context is cancelled mid-build
// =============================================================================

// TestBuildSession_MarkedInterruptedOnContextCancel verifies that cancelling
// the context while a build is running causes the session to be finalized with
// status="interrupted" rather than "running", "failed", or "partial".
//
// RED: buildWorkspacesInParallel does not accept a context.Context.
// The engine maps context.Canceled to a generic build error ("failed").
// Fix: detect ctx.Err()==context.Canceled and write status="interrupted".
func TestBuildSession_MarkedInterruptedOnContextCancel(t *testing.T) {
	store, workspaces := setupInterruptTestStore(t, 2)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buildStarted := make(chan struct{}, len(workspaces))
	slowBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		buildStarted <- struct{}{}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			return nil
		}
	}

	done := make(chan error, 1)
	go func() {
		// NOTE TO IMPLEMENTER: this call must become
		// buildWorkspacesInParallel(ctx, workspaces, 2, slowBuildFn, store)
		// once the context parameter is added (Issue #399).
		done <- buildWorkspacesInParallel(workspaces, 2, slowBuildFn, store)
	}()

	for i := 0; i < len(workspaces); i++ {
		select {
		case <-buildStarted:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for build goroutines to start")
		}
	}
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("buildWorkspacesInParallel did not return after context cancel")
	}

	session, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, session)

	assert.Equal(t, "interrupted", session.Status,
		"[RED] Build session status MUST be 'interrupted' after context cancellation. "+
			"Currently the engine maps ctx.Canceled errors to status='failed'. "+
			"Fix: detect ctx.Err()==context.Canceled and write status='interrupted'.")

	assert.True(t, session.CompletedAt.Valid,
		"[RED] Build session CompletedAt must be set when interrupted.")
}

// =============================================================================
// 2. Workspace entries updated to "interrupted" (not stuck as "building")
// =============================================================================

// TestBuildSession_WorkspacesMarkedInterruptedOnContextCancel verifies that
// workspace entries in non-terminal states are updated to "interrupted" when
// the context is cancelled. This is the core of bug #399 -- rows left as
// "building" can never be healed by the existing allWorkspacesTerminal logic.
//
// RED: workspace rows are left as "failed" (from ctx.Canceled error) or
// remain "building" because the engine has no ctx.Done() watcher.
func TestBuildSession_WorkspacesMarkedInterruptedOnContextCancel(t *testing.T) {
	store, workspaces := setupInterruptTestStore(t, 3)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	buildStarted := make(chan struct{}, len(workspaces))
	slowBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		buildStarted <- struct{}{}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			return nil
		}
	}

	done := make(chan error, 1)
	go func() {
		done <- buildWorkspacesInParallel(workspaces, 3, slowBuildFn, store)
	}()

	for i := 0; i < len(workspaces); i++ {
		select {
		case <-buildStarted:
		case <-time.After(5 * time.Second):
			t.Fatal("timed out waiting for build goroutines to start")
		}
	}
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("buildWorkspacesInParallel did not return after context cancel")
	}

	session, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, session)

	entries, err := store.GetBuildSessionWorkspaces(session.ID)
	require.NoError(t, err)
	require.Len(t, entries, 3, "all workspace entries must exist")

	for _, e := range entries {
		assert.Equal(t, "interrupted", e.Status,
			"[RED] Workspace %d status must be 'interrupted' after context cancel -- "+
				"currently 'failed' (ctx.Canceled treated as build error). "+
				"Fix: on ctx.Done(), update non-terminal workspace rows to 'interrupted'.",
			e.WorkspaceID)
	}
}

// =============================================================================
// 3. Queued workspaces (not yet started) marked "cancelled" on interrupt
// =============================================================================

// TestBuildSession_QueuedWorkspacesMarkedCancelledOnContextCancel verifies
// that workspace entries still in "queued" state (never picked up by a worker)
// are set to "cancelled" rather than left as "queued" indefinitely.
//
// RED: With concurrency=1 and 4 workspaces, 3 will remain "queued" when Ctrl-C
// fires. The current engine leaves them as "queued" (or "failed" if the worker
// loop reaches them). Neither is acceptable.
func TestBuildSession_QueuedWorkspacesMarkedCancelledOnContextCancel(t *testing.T) {
	store, workspaces := setupInterruptTestStore(t, 4)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	firstStarted := make(chan struct{}, 1)
	slowBuildFn := func(ws *models.WorkspaceWithHierarchy) error {
		select {
		case firstStarted <- struct{}{}:
		default:
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Second):
			return nil
		}
	}

	done := make(chan error, 1)
	go func() {
		done <- buildWorkspacesInParallel(workspaces, 1, slowBuildFn, store)
	}()

	select {
	case <-firstStarted:
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for first build to start")
	}
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("buildWorkspacesInParallel did not return after context cancel")
	}

	session, err := store.GetLatestBuildSession()
	require.NoError(t, err)
	require.NotNil(t, session)

	entries, err := store.GetBuildSessionWorkspaces(session.ID)
	require.NoError(t, err)

	for _, e := range entries {
		isTerminalInterrupt := e.Status == "interrupted" || e.Status == "cancelled"
		assert.True(t, isTerminalInterrupt,
			"[RED] Workspace %d has status '%s' after context cancel. "+
				"Expected 'interrupted' (was building) or 'cancelled' (was queued). "+
				"Fix: on ctx.Done(), set building->interrupted and queued->cancelled.",
			e.WorkspaceID, e.Status)
	}
}
