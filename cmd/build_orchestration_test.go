package cmd

// =============================================================================
// TDD Phase 2 (RED): Parallel Build Orchestration — Issue #213
// =============================================================================
// These tests verify the ACTUAL parallel build orchestration behavior.
// They are distinct from build_parallel_test.go which only tested format
// helpers and flag registration.
//
// What's tested here:
//   (a) Workspace auto-discovery for --all
//   (b) Scope-based workspace filtering (--ecosystem, --domain, --app)
//   (c) Parallel execution (worker pool, concurrency bounding)
//   (d) Build failure isolation (one failure doesn't block others)
//   (e) No active workspace required for --all / scope flags
//
// RED state: All tests FAIL because resolveWorkspacesForParallelBuild
//            and buildWorkspacesInParallel do not exist yet.
// GREEN state: After dvm-core implements Issue #213.
// =============================================================================

import (
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupParallelBuildTestStore creates a MockDataStore with an ecosystem,
// domain, and N workspaces for use in parallel build tests.
func setupParallelBuildTestStore(t *testing.T, ecosystemName, domainName string, appWorkspacePairs []struct{ app, workspace string }) *db.MockDataStore {
	t.Helper()
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: ecosystemName}
	require.NoError(t, store.CreateEcosystem(eco))

	dom := &models.Domain{Name: domainName, EcosystemID: eco.ID}
	require.NoError(t, store.CreateDomain(dom))

	for _, pair := range appWorkspacePairs {
		app := &models.App{Name: pair.app, DomainID: dom.ID}
		require.NoError(t, store.CreateApp(app))

		ws := &models.Workspace{Name: pair.workspace, AppID: app.ID}
		require.NoError(t, store.CreateWorkspace(ws))
	}

	return store
}

// =============================================================================
// (a) Workspace Auto-discovery for --all
// =============================================================================

// TestBuildOrchestration_AllFlag_QueriesAllWorkspaces verifies that when
// --all is set, resolveWorkspacesForParallelBuild queries the DataStore
// for ALL workspaces (using FindWorkspaces with empty filter).
//
// RED: resolveWorkspacesForParallelBuild does not exist yet.
func TestBuildOrchestration_AllFlag_QueriesAllWorkspaces(t *testing.T) {
	pairs := []struct{ app, workspace string }{
		{"app-a", "dev"},
		{"app-b", "dev"},
		{"app-c", "dev"},
	}
	store := setupParallelBuildTestStore(t, "my-eco", "my-domain", pairs)

	// --all with no scope flags = empty filter = all workspaces
	flags := HierarchyFlags{}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, true)

	require.NoError(t, err, "--all should succeed without active workspace")
	assert.Len(t, workspaces, 3,
		"--all should return all 3 workspaces, got %d", len(workspaces))
}

// TestBuildOrchestration_AllFlag_NoActiveWorkspaceRequired verifies that
// --all does not fail when no active workspace is set. The function must not
// call resolveFromActiveContext when buildAll is true.
//
// RED: resolveWorkspacesForParallelBuild does not exist yet.
func TestBuildOrchestration_AllFlag_NoActiveWorkspaceRequired(t *testing.T) {
	// Store with no active workspace context
	store := db.NewMockDataStore()
	// store.Context is initialized with empty workspace ID (nil) by default

	eco := &models.Ecosystem{Name: "eco"}
	require.NoError(t, store.CreateEcosystem(eco))
	dom := &models.Domain{Name: "dom", EcosystemID: eco.ID}
	require.NoError(t, store.CreateDomain(dom))
	app := &models.App{Name: "myapp", DomainID: dom.ID}
	require.NoError(t, store.CreateApp(app))
	ws := &models.Workspace{Name: "dev", AppID: app.ID}
	require.NoError(t, store.CreateWorkspace(ws))

	flags := HierarchyFlags{}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, true)

	require.NoError(t, err,
		"--all must NOT error when no active workspace is set")
	assert.NotEmpty(t, workspaces,
		"--all must return workspaces even with no active context")
}

// =============================================================================
// (b) Scope-based Workspace Filtering
// =============================================================================

// TestBuildOrchestration_EcosystemFilter_ReturnsMatchingWorkspaces verifies
// that --ecosystem <name> returns only workspaces in that ecosystem.
//
// RED: resolveWorkspacesForParallelBuild does not exist yet.
func TestBuildOrchestration_EcosystemFilter_ReturnsMatchingWorkspaces(t *testing.T) {
	store := db.NewMockDataStore()

	// Ecosystem A with 2 workspaces
	ecoA := &models.Ecosystem{Name: "ecosystem-a"}
	require.NoError(t, store.CreateEcosystem(ecoA))
	domA := &models.Domain{Name: "domain-a", EcosystemID: ecoA.ID}
	require.NoError(t, store.CreateDomain(domA))
	appA1 := &models.App{Name: "app-a1", DomainID: domA.ID}
	require.NoError(t, store.CreateApp(appA1))
	require.NoError(t, store.CreateWorkspace(&models.Workspace{Name: "dev", AppID: appA1.ID}))
	appA2 := &models.App{Name: "app-a2", DomainID: domA.ID}
	require.NoError(t, store.CreateApp(appA2))
	require.NoError(t, store.CreateWorkspace(&models.Workspace{Name: "dev", AppID: appA2.ID}))

	// Ecosystem B with 1 workspace
	ecoB := &models.Ecosystem{Name: "ecosystem-b"}
	require.NoError(t, store.CreateEcosystem(ecoB))
	domB := &models.Domain{Name: "domain-b", EcosystemID: ecoB.ID}
	require.NoError(t, store.CreateDomain(domB))
	appB := &models.App{Name: "app-b", DomainID: domB.ID}
	require.NoError(t, store.CreateApp(appB))
	require.NoError(t, store.CreateWorkspace(&models.Workspace{Name: "dev", AppID: appB.ID}))

	flags := HierarchyFlags{Ecosystem: "ecosystem-a"}
	workspaces, err := resolveWorkspacesForParallelBuild(store, flags, false)

	require.NoError(t, err)
	assert.Len(t, workspaces, 2,
		"--ecosystem ecosystem-a should return exactly 2 workspaces (not workspace from ecosystem-b)")

	for _, w := range workspaces {
		assert.Equal(t, "ecosystem-a", w.Ecosystem.Name,
			"all returned workspaces must belong to ecosystem-a")
	}
}

// TestBuildOrchestration_ScopeFilter_TableDriven tests all scope flags return
// only matching workspaces.
//
// RED: resolveWorkspacesForParallelBuild does not exist yet.
func TestBuildOrchestration_ScopeFilter_TableDriven(t *testing.T) {
	tests := []struct {
		name       string
		flags      HierarchyFlags
		wantCount  int
		wantErrMsg string
	}{
		{
			name:      "domain filter returns matching",
			flags:     HierarchyFlags{Domain: "payments"},
			wantCount: 2,
		},
		{
			name:      "app filter returns matching",
			flags:     HierarchyFlags{App: "portal"},
			wantCount: 1,
		},
		{
			name:       "nonexistent ecosystem returns error",
			flags:      HierarchyFlags{Ecosystem: "ghost-eco"},
			wantCount:  0,
			wantErrMsg: "ghost-eco",
		},
	}

	// Setup: one ecosystem, two domains (payments, ui), portal in ui
	store := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "main-eco"}
	require.NoError(t, store.CreateEcosystem(eco))

	domPay := &models.Domain{Name: "payments", EcosystemID: eco.ID}
	require.NoError(t, store.CreateDomain(domPay))
	appPay1 := &models.App{Name: "pay-service", DomainID: domPay.ID}
	require.NoError(t, store.CreateApp(appPay1))
	require.NoError(t, store.CreateWorkspace(&models.Workspace{Name: "dev", AppID: appPay1.ID}))
	appPay2 := &models.App{Name: "pay-worker", DomainID: domPay.ID}
	require.NoError(t, store.CreateApp(appPay2))
	require.NoError(t, store.CreateWorkspace(&models.Workspace{Name: "dev", AppID: appPay2.ID}))

	domUI := &models.Domain{Name: "ui", EcosystemID: eco.ID}
	require.NoError(t, store.CreateDomain(domUI))
	appPortal := &models.App{Name: "portal", DomainID: domUI.ID}
	require.NoError(t, store.CreateApp(appPortal))
	require.NoError(t, store.CreateWorkspace(&models.Workspace{Name: "dev", AppID: appPortal.ID}))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workspaces, err := resolveWorkspacesForParallelBuild(store, tt.flags, false)

			if tt.wantErrMsg != "" {
				require.Error(t, err,
					"scope with no matches should return an error")
				assert.Contains(t, err.Error(), tt.wantErrMsg,
					"error message should contain the nonexistent scope value")
				return
			}

			require.NoError(t, err)
			assert.Len(t, workspaces, tt.wantCount,
				"filter %+v should return %d workspaces", tt.flags, tt.wantCount)
		})
	}
}
