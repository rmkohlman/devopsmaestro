package cmd

// =============================================================================
// Issue #215: Scope filters auto-build all matches instead of erroring
// =============================================================================
// These tests verify that when scope filters (--ecosystem, --domain, --app)
// match multiple workspaces, the build command treats them as a batch build
// (same as -A behavior) instead of erroring with "ambiguous workspace selection".
//
// The fix routes scope-filtered builds through the parallel build path
// (resolveWorkspacesForParallelBuild) rather than the single-workspace resolver
// (resolveFromHierarchyFlags) which throws AmbiguousError on multi-match.
//
// Interactive commands (attach, exec, detach) must still error on ambiguity.
// =============================================================================

import (
	"database/sql"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	wsresolver "devopsmaestro/pkg/resolver"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Core behavior: scope filters → batch build (not ambiguous error)
// =============================================================================

// TestScopeBatch_EcosystemFilter_MultiMatch_BuildsAll verifies that
// --ecosystem matching multiple workspaces returns all of them for batch
// building, NOT an "ambiguous workspace selection" error.
// This is the primary fix for issue #215.
func TestScopeBatch_EcosystemFilter_MultiMatch_BuildsAll(t *testing.T) {
	store := setupScopeBatchStore(t)

	flags := HierarchyFlags{Ecosystem: "beans-modules"}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, false)

	require.NoError(t, err,
		"--ecosystem with multiple matches must NOT error (issue #215)")
	assert.Len(t, workspaces, 2,
		"--ecosystem beans-modules should return both workspaces")

	// Verify all returned workspaces belong to the correct ecosystem
	for _, ws := range workspaces {
		assert.Equal(t, "beans-modules", ws.Ecosystem.Name,
			"all returned workspaces must belong to ecosystem beans-modules")
	}
}

// TestScopeBatch_DomainFilter_MultiMatch_BuildsAll verifies that --domain
// matching multiple workspaces returns all of them for batch building.
func TestScopeBatch_DomainFilter_MultiMatch_BuildsAll(t *testing.T) {
	store := setupScopeBatchStore(t)

	flags := HierarchyFlags{Domain: "beans-ray-actor-modules"}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, false)

	require.NoError(t, err,
		"--domain with multiple matches must NOT error (issue #215)")
	assert.Len(t, workspaces, 2,
		"--domain beans-ray-actor-modules should return both workspaces")

	for _, ws := range workspaces {
		assert.Equal(t, "beans-ray-actor-modules", ws.Domain.Name)
	}
}

// TestScopeBatch_AppFilter_MultiMatch_BuildsAll verifies that --app matching
// multiple workspaces (e.g., dev + staging) returns all of them.
func TestScopeBatch_AppFilter_MultiMatch_BuildsAll(t *testing.T) {
	store := setupMultiWorkspaceAppStore(t)

	flags := HierarchyFlags{App: "my-service"}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, false)

	require.NoError(t, err,
		"--app with multiple workspaces (dev+staging) must NOT error (issue #215)")
	assert.Len(t, workspaces, 2,
		"--app my-service should return both dev and staging workspaces")
}

// TestScopeBatch_SingleMatch_StillWorks verifies that scope filters matching
// exactly one workspace work correctly (regression check).
func TestScopeBatch_SingleMatch_StillWorks(t *testing.T) {
	store := setupScopeBatchStore(t)

	flags := HierarchyFlags{App: "beans-ray-actorkit"}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, false)

	require.NoError(t, err)
	assert.Len(t, workspaces, 1,
		"--app with single match should return exactly 1 workspace")
	assert.Equal(t, "beans-ray-actorkit", workspaces[0].App.Name)
}

// TestScopeBatch_NoMatch_ReturnsError verifies that scope filters matching
// zero workspaces still return an error (not silently succeed).
func TestScopeBatch_NoMatch_ReturnsError(t *testing.T) {
	store := setupScopeBatchStore(t)

	flags := HierarchyFlags{Ecosystem: "nonexistent-eco"}
	_, err := resolveWorkspacesForParallelBuild(store, flags, false)

	require.Error(t, err, "no matches should return an error")
	assert.Contains(t, err.Error(), "nonexistent-eco",
		"error should mention the ecosystem name that wasn't found")
}

// =============================================================================
// Routing: scope flags go to parallel build, not single-workspace path
// =============================================================================

// TestScopeBatch_RoutingDecision_ScopeFlagsGoToParallel verifies that the
// routing function directs scope-filtered builds to the parallel path,
// preventing the ambiguous error in resolveFromHierarchyFlags.
func TestScopeBatch_RoutingDecision_ScopeFlagsGoToParallel(t *testing.T) {
	tests := []struct {
		name  string
		flags HierarchyFlags
	}{
		{"ecosystem only", HierarchyFlags{Ecosystem: "beans-modules"}},
		{"domain only", HierarchyFlags{Domain: "payments"}},
		{"app only", HierarchyFlags{App: "portal"}},
		{"ecosystem + domain", HierarchyFlags{Ecosystem: "beans-modules", Domain: "payments"}},
		{"workspace only", HierarchyFlags{Workspace: "dev"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shouldRouteToParallelBuild(tt.flags, false)
			assert.True(t, result,
				"scope flags %+v must route to parallel build path (issue #215)",
				tt.flags)
		})
	}
}

// TestScopeBatch_RoutingDecision_NoFlagsNoAll_SingleWorkspacePath verifies
// that without any flags, the single-workspace path is used (backward compat).
func TestScopeBatch_RoutingDecision_NoFlagsNoAll_SingleWorkspacePath(t *testing.T) {
	flags := HierarchyFlags{}
	result := shouldRouteToParallelBuild(flags, false)
	assert.False(t, result,
		"no flags and no --all must use single-workspace path")
}

// =============================================================================
// Guard: interactive commands still error on ambiguity
// =============================================================================

// TestScopeBatch_InteractiveCommands_StillErrorOnAmbiguity verifies that
// the workspace resolver used by attach/exec/detach still returns
// AmbiguousError when multiple workspaces match. These commands can only
// operate on a single workspace and MUST require disambiguation.
func TestScopeBatch_InteractiveCommands_StillErrorOnAmbiguity(t *testing.T) {
	store := setupScopeBatchStore(t)

	// The workspace resolver (used by attach/exec/detach) should return
	// AmbiguousError when multiple workspaces match a filter.
	resolver := wsresolver.NewWorkspaceResolver(store)
	filter := models.WorkspaceFilter{
		EcosystemName: "beans-modules",
	}

	_, err := resolver.Resolve(filter)
	require.Error(t, err, "resolver must error when multiple workspaces match")

	ambiguousErr, ok := wsresolver.IsAmbiguousError(err)
	assert.True(t, ok,
		"error must be AmbiguousError for interactive commands (attach/exec/detach)")
	assert.GreaterOrEqual(t, len(ambiguousErr.Matches), 2,
		"AmbiguousError must contain the matching workspaces for disambiguation")
}

// =============================================================================
// Combined filter tests
// =============================================================================

// TestScopeBatch_CombinedFilters_NarrowsScope verifies that combining multiple
// scope flags narrows the result set (additive filtering).
func TestScopeBatch_CombinedFilters_NarrowsScope(t *testing.T) {
	store := setupScopeBatchStore(t)

	// Ecosystem alone → 2 workspaces
	flags := HierarchyFlags{Ecosystem: "beans-modules"}
	all, err := resolveWorkspacesForParallelBuild(store, flags, false)
	require.NoError(t, err)
	assert.Len(t, all, 2)

	// Ecosystem + app → 1 workspace
	flags = HierarchyFlags{Ecosystem: "beans-modules", App: "beans-ray-actorkit"}
	narrowed, err := resolveWorkspacesForParallelBuild(store, flags, false)
	require.NoError(t, err)
	assert.Len(t, narrowed, 1,
		"combining ecosystem + app should narrow to 1 workspace")
	assert.Equal(t, "beans-ray-actorkit", narrowed[0].App.Name)
}

// =============================================================================
// Test data helpers
// =============================================================================

// setupScopeBatchStore creates a test store matching the issue #215 scenario:
// ecosystem "beans-modules" with domain "beans-ray-actor-modules" containing
// two apps (beans-ray-actorkit and beans-ray-actors), each with a "dev" workspace.
func setupScopeBatchStore(t *testing.T) *db.MockDataStore {
	t.Helper()
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "beans-modules"}
	require.NoError(t, store.CreateEcosystem(eco))

	dom := &models.Domain{
		Name:        "beans-ray-actor-modules",
		EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true},
	}
	require.NoError(t, store.CreateDomain(dom))

	app1 := &models.App{
		Name:     "beans-ray-actorkit",
		DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true},
	}
	require.NoError(t, store.CreateApp(app1))
	ws1 := &models.Workspace{Name: "dev", AppID: app1.ID}
	require.NoError(t, store.CreateWorkspace(ws1))

	app2 := &models.App{
		Name:     "beans-ray-actors",
		DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true},
	}
	require.NoError(t, store.CreateApp(app2))
	ws2 := &models.Workspace{Name: "dev", AppID: app2.ID}
	require.NoError(t, store.CreateWorkspace(ws2))

	return store
}

// setupMultiWorkspaceAppStore creates a test store with a single app that has
// multiple workspaces (dev + staging), for testing --app multi-match.
func setupMultiWorkspaceAppStore(t *testing.T) *db.MockDataStore {
	t.Helper()
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "my-eco"}
	require.NoError(t, store.CreateEcosystem(eco))

	dom := &models.Domain{
		Name:        "my-domain",
		EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true},
	}
	require.NoError(t, store.CreateDomain(dom))

	app := &models.App{
		Name:     "my-service",
		DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true},
	}
	require.NoError(t, store.CreateApp(app))

	ws1 := &models.Workspace{Name: "dev", AppID: app.ID}
	require.NoError(t, store.CreateWorkspace(ws1))

	ws2 := &models.Workspace{Name: "staging", AppID: app.ID}
	require.NoError(t, store.CreateWorkspace(ws2))

	return store
}
