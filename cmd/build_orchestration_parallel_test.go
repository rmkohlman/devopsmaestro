package cmd

// =============================================================================
// TDD Phase 2 (RED): Parallel Build Execution Behavior — Issue #213
// =============================================================================
// These tests verify the worker-pool-based parallel execution, failure
// isolation, exit code aggregation, and detach mode behavior.
//
// RED state: All tests FAIL because buildWorkspacesInParallel, the worker
//            pool, and the detach path do not exist yet.
// GREEN state: After dvm-core implements Issue #213.
// =============================================================================

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildOperatorFunc is a function type that simulates building one workspace.
// Tests inject this to observe and control build behavior without real Docker.
type buildOperatorFunc func(ws *models.WorkspaceWithHierarchy) error

// =============================================================================
// (c) Parallel Execution — Concurrency Bounded by --concurrency
// =============================================================================

// TestBuildOrchestration_ParallelExecution_CallsOperatorForEachWorkspace
// verifies that buildWorkspacesInParallel calls the build operator exactly
// once for each workspace provided.
//
// RED: buildWorkspacesInParallel does not exist yet.
func TestBuildOrchestration_ParallelExecution_CallsOperatorForEachWorkspace(t *testing.T) {
	workspaces := makeTestWorkspaceList(3)

	var callCount atomic.Int32
	operator := func(ws *models.WorkspaceWithHierarchy) error {
		callCount.Add(1)
		return nil
	}

	err := buildWorkspacesInParallel(workspaces, 4, operator)
	require.NoError(t, err)
	assert.Equal(t, int32(3), callCount.Load(),
		"operator must be called exactly once per workspace")
}

// TestBuildOrchestration_ParallelExecution_BoundedByConcurrency verifies that
// at most --concurrency builds run simultaneously. Uses a channel-based counter
// to measure peak concurrent executions.
//
// RED: buildWorkspacesInParallel does not exist yet.
func TestBuildOrchestration_ParallelExecution_BoundedByConcurrency(t *testing.T) {
	const numWorkspaces = 8
	const concurrencyLimit = 3

	workspaces := makeTestWorkspaceList(numWorkspaces)

	var mu sync.Mutex
	activeConcurrent := 0
	peakConcurrent := 0

	operator := func(ws *models.WorkspaceWithHierarchy) error {
		mu.Lock()
		activeConcurrent++
		if activeConcurrent > peakConcurrent {
			peakConcurrent = activeConcurrent
		}
		mu.Unlock()

		// Simulate work
		time.Sleep(5 * time.Millisecond)

		mu.Lock()
		activeConcurrent--
		mu.Unlock()
		return nil
	}

	err := buildWorkspacesInParallel(workspaces, concurrencyLimit, operator)
	require.NoError(t, err)

	assert.LessOrEqual(t, peakConcurrent, concurrencyLimit,
		"peak concurrent builds (%d) must not exceed concurrency limit (%d)",
		peakConcurrent, concurrencyLimit)
	assert.Greater(t, peakConcurrent, 1,
		"peak concurrent builds should be > 1 to confirm parallelism")
}

// TestBuildOrchestration_ParallelExecution_IsActuallyParallel verifies that
// N workspaces with concurrency=N complete faster than serial execution would.
//
// RED: buildWorkspacesInParallel does not exist yet.
func TestBuildOrchestration_ParallelExecution_IsActuallyParallel(t *testing.T) {
	const numWorkspaces = 4
	const workDuration = 20 * time.Millisecond
	const concurrencyLimit = numWorkspaces // all can run at once

	workspaces := makeTestWorkspaceList(numWorkspaces)

	operator := func(ws *models.WorkspaceWithHierarchy) error {
		time.Sleep(workDuration)
		return nil
	}

	start := time.Now()
	err := buildWorkspacesInParallel(workspaces, concurrencyLimit, operator)
	elapsed := time.Since(start)

	require.NoError(t, err)

	// Serial execution would take numWorkspaces * workDuration.
	// Parallel should be much faster — allow 2x slack for timing jitter.
	serialTime := time.Duration(numWorkspaces) * workDuration
	assert.Less(t, elapsed, serialTime,
		"parallel execution (elapsed=%v) should be faster than serial (%v)", elapsed, serialTime)
}

// =============================================================================
// (d) Build Failure Isolation
// =============================================================================

// TestBuildOrchestration_FailureIsolation_OtherWorkspacesContinue verifies
// that a failure in one workspace does NOT stop other workspaces from building.
// Workspaces 1 and 3 succeed; workspace 2 fails.
//
// RED: buildWorkspacesInParallel does not exist yet.
func TestBuildOrchestration_FailureIsolation_OtherWorkspacesContinue(t *testing.T) {
	workspaces := makeTestWorkspaceList(3)
	failingWorkspace := workspaces[1].Workspace.Name

	var builtWorkspaces []string
	var mu sync.Mutex

	operator := func(ws *models.WorkspaceWithHierarchy) error {
		if ws.Workspace.Name == failingWorkspace {
			return errors.New("build failed: container error")
		}
		mu.Lock()
		builtWorkspaces = append(builtWorkspaces, ws.Workspace.Name)
		mu.Unlock()
		return nil
	}

	// The function should return an error (aggregate), but still build all workspaces
	err := buildWorkspacesInParallel(workspaces, 3, operator)

	// Must return an error because one workspace failed
	assert.Error(t, err,
		"buildWorkspacesInParallel must return error when any workspace fails")

	// But the other 2 workspaces must still have been built
	assert.Len(t, builtWorkspaces, 2,
		"2 workspaces should complete despite 1 failure — failure isolation required")
}

// TestBuildOrchestration_FailureIsolation_AllThreeFail verifies that when all
// workspaces fail, the error is returned but all were attempted.
//
// RED: buildWorkspacesInParallel does not exist yet.
func TestBuildOrchestration_FailureIsolation_AllThreeFail(t *testing.T) {
	workspaces := makeTestWorkspaceList(3)
	var attemptCount atomic.Int32

	operator := func(ws *models.WorkspaceWithHierarchy) error {
		attemptCount.Add(1)
		return errors.New("build failed")
	}

	err := buildWorkspacesInParallel(workspaces, 3, operator)

	assert.Error(t, err,
		"must return error when all workspaces fail")
	assert.Equal(t, int32(3), attemptCount.Load(),
		"all 3 workspaces must be attempted even when they all fail")
}

// =============================================================================
// (e) Exit Code Aggregation (wired into real builds)
// =============================================================================

// TestBuildOrchestration_ExitCode_ZeroWhenAllSucceed verifies that
// buildWorkspacesInParallel returns nil (exit 0 equivalent) when all succeed.
//
// RED: buildWorkspacesInParallel does not exist yet.
func TestBuildOrchestration_ExitCode_ZeroWhenAllSucceed(t *testing.T) {
	workspaces := makeTestWorkspaceList(4)

	operator := func(ws *models.WorkspaceWithHierarchy) error {
		return nil // all succeed
	}

	err := buildWorkspacesInParallel(workspaces, 4, operator)
	assert.NoError(t, err,
		"buildWorkspacesInParallel must return nil when all workspaces succeed")
}

// TestBuildOrchestration_ExitCode_NonZeroOnPartialFailure verifies that
// buildWorkspacesInParallel returns a non-nil error when any workspace fails.
//
// RED: buildWorkspacesInParallel does not exist yet.
func TestBuildOrchestration_ExitCode_NonZeroOnPartialFailure(t *testing.T) {
	tests := []struct {
		name          string
		failIndexes   []int
		numWorkspaces int
	}{
		{"first fails", []int{0}, 3},
		{"middle fails", []int{1}, 3},
		{"last fails", []int{2}, 3},
		{"all fail", []int{0, 1, 2}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspaces := makeTestWorkspaceList(tt.numWorkspaces)
			failSet := make(map[int]bool)
			for _, i := range tt.failIndexes {
				failSet[i] = true
			}

			i := 0
			operator := func(ws *models.WorkspaceWithHierarchy) error {
				// Note: parallel execution means we can't rely on index ordering,
				// but for this test we use workspace name as identity
				if ws.Workspace.Name == workspaces[tt.failIndexes[0]].Workspace.Name {
					return errors.New("build error")
				}
				return nil
			}
			_ = i // suppress unused

			err := buildWorkspacesInParallel(workspaces, 4, operator)
			assert.Error(t, err,
				"must return error when workspace(s) %v fail", tt.failIndexes)
		})
	}
}

// =============================================================================
// (f) Detach Mode — Returns Before Builds Complete
// =============================================================================

// TestBuildOrchestration_DetachMode_ReturnsImmediately verifies that
// buildWorkspacesInParallelDetached returns promptly with a session ID
// before the builds complete, allowing the caller to monitor via
// 'dvm build status'.
//
// RED: buildWorkspacesInParallelDetached does not exist yet.
func TestBuildOrchestration_DetachMode_ReturnsImmediately(t *testing.T) {
	workspaces := makeTestWorkspaceList(3)

	// Operator with significant delay — if detach doesn't work, test will be slow
	slowOperator := func(ws *models.WorkspaceWithHierarchy) error {
		time.Sleep(200 * time.Millisecond)
		return nil
	}

	start := time.Now()
	sessionID, err := buildWorkspacesInParallelDetached(workspaces, 4, slowOperator)
	elapsed := time.Since(start)

	require.NoError(t, err,
		"detached build launch must not return an error")
	assert.NotEmpty(t, sessionID,
		"detached build must return a non-empty session ID")

	// Must return well before the builds complete (200ms * 3 workspaces = 600ms serial)
	assert.Less(t, elapsed, 100*time.Millisecond,
		"--detach must return immediately (elapsed=%v), not wait for builds", elapsed)
}

// TestBuildOrchestration_DetachMode_SessionIDIsUnique verifies that two
// concurrent detached build sessions get distinct session IDs.
//
// RED: buildWorkspacesInParallelDetached does not exist yet.
func TestBuildOrchestration_DetachMode_SessionIDIsUnique(t *testing.T) {
	workspaces := makeTestWorkspaceList(1)
	noop := func(ws *models.WorkspaceWithHierarchy) error { return nil }

	id1, err1 := buildWorkspacesInParallelDetached(workspaces, 1, noop)
	id2, err2 := buildWorkspacesInParallelDetached(workspaces, 1, noop)

	require.NoError(t, err1)
	require.NoError(t, err2)
	assert.NotEqual(t, id1, id2,
		"each detached build session must have a unique session ID")
}

// =============================================================================
// (g) No Active Workspace Required (scope flags path)
// =============================================================================

// TestBuildOrchestration_ScopeFlags_NoActiveWorkspaceNeeded verifies that
// when scope flags are provided (HasAnyFlag() == true), the build does NOT
// error on "no active workspace set".
//
// This tests that the RunE routing logic correctly detects scope flags and
// routes to the multi-workspace path, bypassing resolveFromActiveContext.
//
// RED: The routing logic in build.go RunE does not exist yet.
func TestBuildOrchestration_ScopeFlags_NoActiveWorkspaceNeeded(t *testing.T) {
	// HierarchyFlags with any field set must NOT require active workspace
	tests := []struct {
		name  string
		flags HierarchyFlags
	}{
		{"ecosystem flag", HierarchyFlags{Ecosystem: "my-eco"}},
		{"domain flag", HierarchyFlags{Domain: "my-domain"}},
		{"app flag", HierarchyFlags{App: "my-app"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// With any scope flag, ShouldRoutToParallelBuild must be true.
			// This function tells the build orchestrator to skip single-workspace
			// path (which requires active context) and go to multi-workspace path.
			result := shouldRouteToParallelBuild(tt.flags, false)
			assert.True(t, result,
				"scope flag %+v must route to parallel build path (no active workspace needed)",
				tt.flags)
		})
	}
}

// TestBuildOrchestration_NoFlags_NoAllFlag_DoesNotRouteToParallel verifies
// that with no flags and buildAll=false, the single-workspace path is used.
//
// RED: shouldRouteToParallelBuild does not exist yet.
func TestBuildOrchestration_NoFlags_NoAllFlag_DoesNotRouteToParallel(t *testing.T) {
	flags := HierarchyFlags{} // no flags set
	result := shouldRouteToParallelBuild(flags, false)
	assert.False(t, result,
		"no scope flags and --all=false must use the single-workspace path (backward compat)")
}

// TestBuildOrchestration_AllFlag_RoutesToParallelBuild verifies that --all
// alone (no scope flags) routes to the parallel build path.
//
// RED: shouldRouteToParallelBuild does not exist yet.
func TestBuildOrchestration_AllFlag_RoutesToParallelBuild(t *testing.T) {
	flags := HierarchyFlags{}                         // no scope flags
	result := shouldRouteToParallelBuild(flags, true) // buildAll = true
	assert.True(t, result,
		"--all without scope flags must route to parallel build path")
}

// =============================================================================
// Test Data Helpers
// =============================================================================

// makeTestWorkspaceList creates N WorkspaceWithHierarchy entries for testing.
func makeTestWorkspaceList(n int) []*models.WorkspaceWithHierarchy {
	eco := &models.Ecosystem{ID: 1, Name: "test-eco"}
	dom := &models.Domain{ID: 1, Name: "test-domain", EcosystemID: 1}

	result := make([]*models.WorkspaceWithHierarchy, n)
	for i := 0; i < n; i++ {
		app := &models.App{ID: i + 1, Name: "test-app", DomainID: 1}
		ws := &models.Workspace{
			ID:    i + 1,
			Name:  wsName(i),
			AppID: i + 1,
		}
		result[i] = &models.WorkspaceWithHierarchy{
			Workspace: ws,
			App:       app,
			Domain:    dom,
			Ecosystem: eco,
		}
	}
	return result
}

// wsName returns a deterministic workspace name for a given index.
func wsName(i int) string {
	names := []string{"ws-alpha", "ws-beta", "ws-gamma", "ws-delta",
		"ws-epsilon", "ws-zeta", "ws-eta", "ws-theta"}
	if i < len(names) {
		return names[i]
	}
	return "ws-" + string(rune('a'+i))
}
