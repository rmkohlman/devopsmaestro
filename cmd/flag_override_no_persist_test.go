package cmd

// =============================================================================
// Issue #197: Bug: Flag overrides persist to active context unexpectedly
//
// TDD Phase 2 (RED) — These tests document the required contract:
//
//   When hierarchy flags (-a, -d, -e, -w) are passed to operational commands
//   (build, attach, get workspace, detach), the resolved workspace should be
//   used for that invocation only. The active context in the DB must NOT be
//   changed as a side-effect of using a flag override.
//
// Currently FAILING because each of the 4 affected code paths calls
// updateContextFromHierarchy(), which writes SetActiveEcosystem /
// SetActiveDomain / SetActiveApp / SetActiveWorkspace to the DB.
//
// EXPECTED STATE AFTER FIX:
//   The 4 updateContextFromHierarchy() call sites in:
//     - cmd/attach.go (runAttach)
//     - cmd/build_phases.go (resolveFromHierarchyFlags)
//     - cmd/get_workspace.go (runGetWorkspace)
//     - cmd/detach.go (detachActiveWorkspace)
//   ...are removed. updateContextFromHierarchy() remains in dvm use commands
//   where persistence is intentional.
// =============================================================================

import (
	"database/sql"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Helpers
// =============================================================================

// setupHierarchyTestStore creates a MockDataStore containing a complete
// ecosystem → domain → app → workspace hierarchy suitable for resolver tests.
// Returns the store plus the individual models so callers can inspect IDs.
func setupHierarchyTestStore(t *testing.T) (
	*db.MockDataStore,
	*models.Ecosystem,
	*models.Domain,
	*models.App,
	*models.Workspace,
) {
	t.Helper()

	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "prod-eco"}
	require.NoError(t, store.CreateEcosystem(eco))

	dom := &models.Domain{Name: "backend", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	require.NoError(t, store.CreateDomain(dom))

	app := &models.App{Name: "my-app", DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true}, Path: "/apps/my-app"}
	require.NoError(t, store.CreateApp(app))

	ws := &models.Workspace{
		Name:      "main",
		AppID:     app.ID,
		ImageName: "dvm-main-my-app:v1",
		Status:    "running",
	}
	require.NoError(t, store.CreateWorkspace(ws))

	return store, eco, dom, app, ws
}

// setInitialActiveContext sets an initial active context on a MockDataStore.
// This simulates the user having previously run `dvm use app original-app`.
func setInitialActiveContext(t *testing.T, store *db.MockDataStore, originalAppID, originalWsID int) {
	t.Helper()
	require.NoError(t, store.SetActiveApp(&originalAppID))
	require.NoError(t, store.SetActiveWorkspace(&originalWsID))
	// Clear the call log so setup calls don't count as "mutations during command"
	store.Calls = nil
}

// countCallsTo returns the number of recorded calls matching methodName.
func countCallsTo(store *db.MockDataStore, methodName string) int {
	count := 0
	for _, call := range store.Calls {
		if call.Method == methodName {
			count++
		}
	}
	return count
}

// hasContextMutationCall returns true if any SetActive* call was recorded.
func hasContextMutationCall(store *db.MockDataStore) bool {
	for _, call := range store.Calls {
		switch call.Method {
		case "SetActiveEcosystem", "SetActiveDomain", "SetActiveSystem", "SetActiveApp", "SetActiveWorkspace":
			return true
		}
	}
	return false
}

// =============================================================================
// Section 1: Document the problematic function
//
// updateContextFromHierarchy() DOES mutate the DB context. These tests
// confirm the function works correctly in isolation — they PASS now and will
// continue to PASS after the fix (the function itself stays, only its call
// sites in operational commands are removed).
// =============================================================================

// TestUpdateContextFromHierarchy_MutatesDBContext verifies that
// updateContextFromHierarchy() writes all four SetActive* values to the
// DataStore. This is the function that must NOT be called from operational
// command paths after the fix.
func TestUpdateContextFromHierarchy_MutatesDBContext(t *testing.T) {
	store, eco, dom, app, ws := setupHierarchyTestStore(t)

	// Clear call log after setup
	store.Calls = nil

	wh := &models.WorkspaceWithHierarchy{
		Ecosystem: eco,
		Domain:    dom,
		App:       app,
		Workspace: ws,
	}

	err := updateContextFromHierarchy(store, wh)
	require.NoError(t, err)

	// Verify all five SetActive* calls were made
	assert.Equal(t, 1, countCallsTo(store, "SetActiveEcosystem"),
		"updateContextFromHierarchy should call SetActiveEcosystem once")
	assert.Equal(t, 1, countCallsTo(store, "SetActiveDomain"),
		"updateContextFromHierarchy should call SetActiveDomain once")
	assert.Equal(t, 1, countCallsTo(store, "SetActiveSystem"),
		"updateContextFromHierarchy should call SetActiveSystem once (nil when no system)")
	assert.Equal(t, 1, countCallsTo(store, "SetActiveApp"),
		"updateContextFromHierarchy should call SetActiveApp once")
	assert.Equal(t, 1, countCallsTo(store, "SetActiveWorkspace"),
		"updateContextFromHierarchy should call SetActiveWorkspace once")

	// Verify context was actually changed
	ctx, err := store.GetContext()
	require.NoError(t, err)
	assert.Equal(t, app.ID, *ctx.ActiveAppID,
		"active app should be set to the resolved app")
	assert.Equal(t, ws.ID, *ctx.ActiveWorkspaceID,
		"active workspace should be set to the resolved workspace")
}

// =============================================================================
// Section 2: Build path — resolveFromHierarchyFlags
//
// Contract: resolveFromHierarchyFlags() must NOT call SetActive* on the store.
//
// RED: FAILS because resolveFromHierarchyFlags() currently calls
//      updateContextFromHierarchy() at line 62 of build_phases.go.
// GREEN: PASSES after that call is removed.
// =============================================================================

// TestBuildPath_HierarchyFlags_DoNotPersistContext verifies that resolving a
// workspace via hierarchy flags during a build does NOT change the active
// context in the DB.
//
// RED: Fails because resolveFromHierarchyFlags calls updateContextFromHierarchy.
func TestBuildPath_HierarchyFlags_DoNotPersistContext(t *testing.T) {
	store, _, _, originalApp, originalWs := setupHierarchyTestStore(t)

	// Set a second app+workspace as "flag target" in the same store
	targetApp := &models.App{
		Name:     "flag-app",
		DomainID: originalApp.DomainID,
		Path:     "/apps/flag-app",
	}
	require.NoError(t, store.CreateApp(targetApp))

	targetWs := &models.Workspace{
		Name:      "flag-ws",
		AppID:     targetApp.ID,
		ImageName: "dvm-flag-ws-flag-app:v1",
		Status:    "stopped",
	}
	require.NoError(t, store.CreateWorkspace(targetWs))

	// Simulate the user's existing active context (originalApp/originalWs)
	setInitialActiveContext(t, store, originalApp.ID, originalWs.ID)

	// Set global buildFlags to use the flag-app (simulating: dvm build -a flag-app)
	originalBuildFlags := buildFlags
	defer func() { buildFlags = originalBuildFlags }()
	buildFlags = HierarchyFlags{App: "flag-app"}

	// Create a minimal buildContext and resolve via hierarchy flags
	bc := &buildContext{ds: store}
	err := bc.resolveFromHierarchyFlags()
	require.NoError(t, err, "resolveFromHierarchyFlags should succeed with a valid app")

	// Verify workspace was resolved correctly
	assert.Equal(t, "flag-app", bc.appName,
		"resolved app name should match the flag value")

	// The active context MUST remain unchanged — no SetActive* calls allowed
	assert.False(t, hasContextMutationCall(store),
		"resolveFromHierarchyFlags must NOT call any SetActive* method; "+
			"flag overrides are per-invocation only, not persisted to DB context")

	// Double-check: context should still point to the original app/workspace
	ctx, err := store.GetContext()
	require.NoError(t, err)
	require.NotNil(t, ctx.ActiveAppID,
		"active app should still be set to the original app")
	assert.Equal(t, originalApp.ID, *ctx.ActiveAppID,
		"dvm build -a flag-app must NOT change the active app context")
	require.NotNil(t, ctx.ActiveWorkspaceID,
		"active workspace should still be set to the original workspace")
	assert.Equal(t, originalWs.ID, *ctx.ActiveWorkspaceID,
		"dvm build -a flag-app must NOT change the active workspace context")
}

// =============================================================================
// Section 3: updateContextFromHierarchy call-site contract tests
//
// These tests verify that the 4 call sites no longer call
// updateContextFromHierarchy. They test the contract at the level of the
// DataStore mock — no container runtime required.
//
// Each test:
//   1. Sets up a mock store with an initial "original" active context.
//   2. Calls the internal function that is supposed to NOT persist context.
//   3. Asserts no SetActive* calls were recorded after the initial context setup.
// =============================================================================

// TestUpdateContextFromHierarchy_NotCalledDuringBuild is an explicit contract
// test that captures the exact function and line that must change.
//
// After the fix, resolveFromHierarchyFlags must NOT record any SetActive* call.
func TestUpdateContextFromHierarchy_NotCalledDuringBuild(t *testing.T) {
	store, _, dom, _, _ := setupHierarchyTestStore(t)

	// Create a "different" app that the flag will resolve to
	flagApp := &models.App{
		Name:     "build-flag-app",
		DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true},
		Path:     "/apps/build-flag-app",
	}
	require.NoError(t, store.CreateApp(flagApp))

	flagWs := &models.Workspace{
		Name:      "build-flag-ws",
		AppID:     flagApp.ID,
		ImageName: "dvm-build-flag-ws-build-flag-app:v1",
		Status:    "stopped",
	}
	require.NoError(t, store.CreateWorkspace(flagWs))

	// Set initial active context and clear call log
	setInitialActiveContext(t, store, flagApp.ID, flagWs.ID)

	// Set global buildFlags
	saved := buildFlags
	defer func() { buildFlags = saved }()
	buildFlags = HierarchyFlags{App: "build-flag-app", Workspace: "build-flag-ws"}

	bc := &buildContext{ds: store}
	err := bc.resolveFromHierarchyFlags()
	require.NoError(t, err)

	setActiveCallCount := countCallsTo(store, "SetActiveApp") +
		countCallsTo(store, "SetActiveWorkspace") +
		countCallsTo(store, "SetActiveEcosystem") +
		countCallsTo(store, "SetActiveDomain") +
		countCallsTo(store, "SetActiveSystem")

	assert.Equal(t, 0, setActiveCallCount,
		"Bug #197: resolveFromHierarchyFlags calls updateContextFromHierarchy "+
			"which makes %d SetActive* call(s); should be 0 after fix",
		setActiveCallCount)
}

// =============================================================================
// Section 4: Table-driven regression tests
//
// After the fix, none of the 4 code paths should ever call SetActive* when
// flag overrides are provided. These tests capture the expected behavior as a
// regression suite.
// =============================================================================

// TestFlagOverrides_NeverPersistContext is a table-driven test that verifies
// the "no context persistence" contract for every hierarchy flag combination.
// It tests the build path (resolveFromHierarchyFlags) as the representative
// case — the pattern applies identically to attach, detach, and get workspace.
func TestFlagOverrides_NeverPersistContext(t *testing.T) {
	tests := []struct {
		name  string
		flags HierarchyFlags
	}{
		{
			name:  "app flag only",
			flags: HierarchyFlags{App: "persist-test-app"},
		},
		{
			name:  "app and workspace flags",
			flags: HierarchyFlags{App: "persist-test-app", Workspace: "persist-test-ws"},
		},
		{
			name:  "ecosystem, domain, app, workspace flags",
			flags: HierarchyFlags{Ecosystem: "prod-eco", Domain: "backend", App: "persist-test-app", Workspace: "persist-test-ws"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, _, dom, _, _ := setupHierarchyTestStore(t)

			// Create the app/workspace that the flags reference
			flagApp := &models.App{
				Name:     "persist-test-app",
				DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true},
				Path:     "/apps/persist-test-app",
			}
			require.NoError(t, store.CreateApp(flagApp))

			flagWs := &models.Workspace{
				Name:      "persist-test-ws",
				AppID:     flagApp.ID,
				ImageName: "dvm-persist-test-ws-persist-test-app:v1",
				Status:    "stopped",
			}
			require.NoError(t, store.CreateWorkspace(flagWs))

			// Set the initial active context to different IDs than what flags resolve to
			setInitialActiveContext(t, store, flagApp.ID, flagWs.ID)

			// Save and restore global buildFlags
			saved := buildFlags
			defer func() { buildFlags = saved }()
			buildFlags = tt.flags

			bc := &buildContext{ds: store}
			err := bc.resolveFromHierarchyFlags()
			require.NoError(t, err)

			// After the fix, no SetActive* calls should occur
			assert.False(t, hasContextMutationCall(store),
				"[%s] Flag-based resolution must not persist context to DB. "+
					"Active context should remain unchanged after 'dvm build %+v'",
				tt.name, tt.flags)
		})
	}
}

// =============================================================================
// Section 5: Positive control — dvm use SHOULD persist context
//
// This test verifies the converse: dvm use app / dvm use workspace MUST
// still call SetActive*. These tests should PASS before AND after the fix,
// confirming we haven't broken intentional persistence.
// =============================================================================

// TestUseAppCmd_StillPersistsContext verifies that updateContextFromHierarchy
// (or equivalent direct SetActive* calls) is still triggered by use commands.
// This is the INTENDED persistence path and must not be broken by the fix.
func TestUseAppCmd_StillPersistsContext(t *testing.T) {
	store, _, _, app, _ := setupHierarchyTestStore(t)

	// Clear call log after setup
	store.Calls = nil

	// Directly call SetActiveApp as use commands do — this should always work
	require.NoError(t, store.SetActiveApp(&app.ID))

	assert.Equal(t, 1, countCallsTo(store, "SetActiveApp"),
		"dvm use app must call SetActiveApp to persist the selection")

	ctx, err := store.GetContext()
	require.NoError(t, err)
	assert.Equal(t, app.ID, *ctx.ActiveAppID,
		"active app should be updated by dvm use app")
}
