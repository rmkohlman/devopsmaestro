package resolver

// =============================================================================
// Issue #200: Feature: Smart context inference — relax resolver guards,
//             single-ecosystem skip
//
// TDD Phase 2 (RED) — These tests document the required contract for smart
// inference features. They test the InferenceResolver (not yet implemented)
// which wraps WorkspaceResolver with additional logic:
//
//   A. Single-ecosystem inference: when only 1 ecosystem exists, auto-fill it
//   B. Cascade inference via --app: auto-infer domain/ecosystem when unambiguous
//   C. Cascade inference via --domain: auto-infer ecosystem when unambiguous
//   D. No active context required: flags alone are sufficient
//
// All tests in this file FAIL until InferenceResolver is implemented.
//
// The InferenceResolver interface is expected to be:
//
//   type InferenceResolver interface {
//       ResolveWithInference(filter models.WorkspaceFilter) (*models.WorkspaceWithHierarchy, error)
//   }
//
//   func NewInferenceResolver(store db.DataStore) InferenceResolver
//
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
// Test Setup Helpers
// =============================================================================

// setupSingleEcosystemStore creates a store with exactly ONE ecosystem,
// one domain, one app "myapi", and one workspace "main".
// Used for single-ecosystem inference tests.
func setupSingleEcosystemStore(t *testing.T) (
	*db.MockDataStore,
	*models.Ecosystem,
	*models.Domain,
	*models.App,
	*models.Workspace,
) {
	t.Helper()

	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "production"}
	require.NoError(t, store.CreateEcosystem(eco))

	dom := &models.Domain{Name: "backend", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	require.NoError(t, store.CreateDomain(dom))

	app := &models.App{Name: "myapi", DomainID: sql.NullInt64{Int64: int64(dom.ID), Valid: true}, Path: "/apps/myapi"}
	require.NoError(t, store.CreateApp(app))

	ws := &models.Workspace{
		Name:      "main",
		AppID:     app.ID,
		ImageName: "dvm-main-myapi:latest",
		Status:    "created",
	}
	require.NoError(t, store.CreateWorkspace(ws))

	return store, eco, dom, app, ws
}

// setupMultiEcosystemStore creates a store with TWO ecosystems.
// Each ecosystem has a domain with an app named "api".
// Used for ambiguity detection tests.
func setupMultiEcosystemStore(t *testing.T) *db.MockDataStore {
	t.Helper()

	store := db.NewMockDataStore()

	// Ecosystem 1: healthcare
	eco1 := &models.Ecosystem{Name: "healthcare"}
	require.NoError(t, store.CreateEcosystem(eco1))
	dom1 := &models.Domain{Name: "billing", EcosystemID: sql.NullInt64{Int64: int64(eco1.ID), Valid: true}}
	require.NoError(t, store.CreateDomain(dom1))
	app1 := &models.App{Name: "api", DomainID: sql.NullInt64{Int64: int64(dom1.ID), Valid: true}, Path: "/apps/healthcare-api"}
	require.NoError(t, store.CreateApp(app1))
	ws1 := &models.Workspace{
		Name:      "dev",
		AppID:     app1.ID,
		ImageName: "dvm-dev-api:latest",
		Status:    "created",
	}
	require.NoError(t, store.CreateWorkspace(ws1))

	// Ecosystem 2: fintech
	eco2 := &models.Ecosystem{Name: "fintech"}
	require.NoError(t, store.CreateEcosystem(eco2))
	dom2 := &models.Domain{Name: "payments", EcosystemID: sql.NullInt64{Int64: int64(eco2.ID), Valid: true}}
	require.NoError(t, store.CreateDomain(dom2))
	app2 := &models.App{Name: "api", DomainID: sql.NullInt64{Int64: int64(dom2.ID), Valid: true}, Path: "/apps/fintech-api"}
	require.NoError(t, store.CreateApp(app2))
	ws2 := &models.Workspace{
		Name:      "dev",
		AppID:     app2.ID,
		ImageName: "dvm-dev-fintech-api:latest",
		Status:    "created",
	}
	require.NoError(t, store.CreateWorkspace(ws2))

	return store
}

// setupMixedAppStore creates a store with ONE ecosystem containing TWO domains,
// each with a DIFFERENT app. Used for unambiguous cascade inference.
func setupMixedAppStore(t *testing.T) *db.MockDataStore {
	t.Helper()

	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "platform"}
	require.NoError(t, store.CreateEcosystem(eco))

	dom1 := &models.Domain{Name: "frontend-domain", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	require.NoError(t, store.CreateDomain(dom1))
	appFrontend := &models.App{Name: "frontend", DomainID: sql.NullInt64{Int64: int64(dom1.ID), Valid: true}, Path: "/apps/frontend"}
	require.NoError(t, store.CreateApp(appFrontend))
	wsFrontend := &models.Workspace{
		Name:      "staging",
		AppID:     appFrontend.ID,
		ImageName: "dvm-staging-frontend:latest",
		Status:    "created",
	}
	require.NoError(t, store.CreateWorkspace(wsFrontend))

	dom2 := &models.Domain{Name: "backend-domain", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	require.NoError(t, store.CreateDomain(dom2))
	appBackend := &models.App{Name: "backend-api", DomainID: sql.NullInt64{Int64: int64(dom2.ID), Valid: true}, Path: "/apps/backend-api"}
	require.NoError(t, store.CreateApp(appBackend))
	wsBackend := &models.Workspace{
		Name:      "staging",
		AppID:     appBackend.ID,
		ImageName: "dvm-staging-backend-api:latest",
		Status:    "created",
	}
	require.NoError(t, store.CreateWorkspace(wsBackend))

	return store
}

// =============================================================================
// A. Single-Ecosystem Inference Tests
// =============================================================================

// TestInferenceResolver_SingleEcosystem_AutoFillsEcosystem verifies that when
// exactly one ecosystem exists and the filter has no ecosystem specified,
// the InferenceResolver auto-fills it and successfully resolves the workspace.
//
// RED: Fails because InferenceResolver does not exist yet.
// GREEN: Passes once InferenceResolver.ResolveWithInference() is implemented
//
//	to call CountEcosystems() and auto-fill when count == 1.
func TestInferenceResolver_SingleEcosystem_AutoFillsEcosystem(t *testing.T) {
	store, _, _, _, _ := setupSingleEcosystemStore(t)

	resolver := NewInferenceResolver(store)

	// Provide only app + workspace — NO ecosystem flag.
	// Should succeed because there's only 1 ecosystem.
	filter := models.WorkspaceFilter{
		AppName:       "myapi",
		WorkspaceName: "main",
	}

	result, err := resolver.ResolveWithInference(filter)

	require.NoError(t, err, "should resolve without ecosystem flag when only 1 ecosystem exists")
	require.NotNil(t, result)
	assert.Equal(t, "main", result.Workspace.Name)
	assert.Equal(t, "myapi", result.App.Name)
	assert.Equal(t, "production", result.Ecosystem.Name,
		"ecosystem should be auto-filled from the single ecosystem")
}

// TestInferenceResolver_SingleEcosystem_AppOnly verifies that with only
// -a myapi and a single ecosystem, the resolver succeeds without -e or -d flags.
//
// RED: Fails because InferenceResolver does not exist yet.
func TestInferenceResolver_SingleEcosystem_AppOnly(t *testing.T) {
	store, _, _, _, _ := setupSingleEcosystemStore(t)

	resolver := NewInferenceResolver(store)

	// Only provide app name — no ecosystem, domain, or workspace.
	// Single ecosystem + single match → should resolve.
	filter := models.WorkspaceFilter{
		AppName: "myapi",
	}

	result, err := resolver.ResolveWithInference(filter)

	require.NoError(t, err, "single ecosystem + unique app name should resolve without additional flags")
	require.NotNil(t, result)
	assert.Equal(t, "myapi", result.App.Name)
	assert.Equal(t, "production", result.Ecosystem.Name)
}

// TestInferenceResolver_MultipleEcosystems_RequiresEcosystem verifies that
// when multiple ecosystems exist AND the app name is ambiguous across them,
// an AmbiguousError is returned (not silent wrong resolution).
//
// RED: Fails because InferenceResolver does not exist yet.
func TestInferenceResolver_MultipleEcosystems_AmbiguousApp_ReturnsAmbiguousError(t *testing.T) {
	store := setupMultiEcosystemStore(t)

	resolver := NewInferenceResolver(store)

	// "api" exists in both healthcare/billing and fintech/payments
	filter := models.WorkspaceFilter{
		AppName:       "api",
		WorkspaceName: "dev",
	}

	_, err := resolver.ResolveWithInference(filter)

	require.Error(t, err, "ambiguous app name across ecosystems should return an error")
	_, ok := IsAmbiguousError(err)
	assert.True(t, ok,
		"error should be AmbiguousError so caller can show disambiguation; got: %T: %v", err, err)
}

// =============================================================================
// B. Cascade Inference via --app flag
// =============================================================================

// TestInferenceResolver_AppInference_UnambiguousApp_AutoInfersDomainAndEcosystem
// verifies that when --app is provided and exactly one app with that name exists,
// the InferenceResolver auto-infers the domain and ecosystem.
//
// Setup: 1 ecosystem with 2 domains. Each domain has a differently-named app.
// With -a frontend and -w staging → should resolve without -e or -d.
//
// RED: Fails because InferenceResolver does not exist yet.
func TestInferenceResolver_AppInference_UnambiguousApp_AutoInfersDomainAndEcosystem(t *testing.T) {
	store := setupMixedAppStore(t)

	resolver := NewInferenceResolver(store)

	// "frontend" exists in exactly one domain → unambiguous cascade
	filter := models.WorkspaceFilter{
		AppName:       "frontend",
		WorkspaceName: "staging",
	}

	result, err := resolver.ResolveWithInference(filter)

	require.NoError(t, err, "unambiguous app name should cascade-infer domain and ecosystem")
	require.NotNil(t, result)
	assert.Equal(t, "frontend", result.App.Name)
	assert.Equal(t, "frontend-domain", result.Domain.Name,
		"domain should be auto-inferred from app lookup")
	assert.Equal(t, "platform", result.Ecosystem.Name,
		"ecosystem should be auto-inferred from domain lookup")
	assert.Equal(t, "staging", result.Workspace.Name)
}

// TestInferenceResolver_AppInference_NonexistentApp_ReturnsNotFound verifies
// that requesting an app that doesn't exist returns ErrNoWorkspaceFound.
//
// RED: Fails because InferenceResolver does not exist yet.
func TestInferenceResolver_AppInference_NonexistentApp_ReturnsNotFound(t *testing.T) {
	store, _, _, _, _ := setupSingleEcosystemStore(t)

	resolver := NewInferenceResolver(store)

	filter := models.WorkspaceFilter{
		AppName:       "nonexistent-app",
		WorkspaceName: "main",
	}

	_, err := resolver.ResolveWithInference(filter)

	require.Error(t, err, "nonexistent app should return an error")
	assert.True(t, IsNoWorkspaceFoundError(err),
		"error should be ErrNoWorkspaceFound for a nonexistent app; got: %T: %v", err, err)
}

// TestInferenceResolver_AppInference_AmbiguousApp_ReturnsAmbiguousError verifies
// that when --app matches multiple apps across different domains/ecosystems,
// an AmbiguousError is returned with all matches listed.
//
// Setup: 2 ecosystems, each with app named "api"
//
// RED: Fails because InferenceResolver does not exist yet.
func TestInferenceResolver_AppInference_AmbiguousApp_ReturnsAmbiguousError(t *testing.T) {
	store := setupMultiEcosystemStore(t)

	resolver := NewInferenceResolver(store)

	// "api" exists in both ecosystems → ambiguous
	filter := models.WorkspaceFilter{
		AppName: "api",
	}

	_, err := resolver.ResolveWithInference(filter)

	require.Error(t, err)
	ambiguousErr, ok := IsAmbiguousError(err)
	require.True(t, ok,
		"error should be AmbiguousError so caller can show disambiguation; got: %T: %v", err, err)
	assert.GreaterOrEqual(t, len(ambiguousErr.Matches), 2,
		"AmbiguousError should contain all matching workspaces")
}

// =============================================================================
// C. Flag-only resolution — no active context required
// =============================================================================

// TestInferenceResolver_NoActiveContext_FlagsAreEnough verifies that the
// InferenceResolver resolves a workspace using flags ALONE, with no active
// context set in the DataStore (no prior `dvm use app`).
//
// This tests the core fix for Part A of issue #200: commands with -a myapi -w main
// should work without requiring `dvm use app` first.
//
// RED: Fails because InferenceResolver does not exist yet.
func TestInferenceResolver_NoActiveContext_FlagsAreEnough(t *testing.T) {
	store, _, _, _, _ := setupSingleEcosystemStore(t)

	// Explicitly: NO active context set (store starts with nil context)
	resolver := NewInferenceResolver(store)

	filter := models.WorkspaceFilter{
		AppName:       "myapi",
		WorkspaceName: "main",
	}

	result, err := resolver.ResolveWithInference(filter)

	require.NoError(t, err,
		"should resolve workspace from flags alone, without needing prior 'dvm use app'")
	require.NotNil(t, result)
	assert.Equal(t, "myapi", result.App.Name)
	assert.Equal(t, "main", result.Workspace.Name)
}

// TestInferenceResolver_NoActiveContext_DoesNotMutateContext verifies that
// InferenceResolver does NOT write SetActive* to the DB — inference is
// read-only and per-invocation only.
//
// RED: Fails because InferenceResolver does not exist yet.
func TestInferenceResolver_NoActiveContext_DoesNotMutateContext(t *testing.T) {
	store, _, _, _, _ := setupSingleEcosystemStore(t)

	// Clear call log before the actual resolver call
	store.Calls = nil

	resolver := NewInferenceResolver(store)

	filter := models.WorkspaceFilter{
		AppName:       "myapi",
		WorkspaceName: "main",
	}

	_, err := resolver.ResolveWithInference(filter)
	require.NoError(t, err)

	// Verify no SetActive* calls were made
	for _, call := range store.Calls {
		switch call.Method {
		case "SetActiveEcosystem", "SetActiveDomain", "SetActiveApp", "SetActiveWorkspace":
			t.Errorf("InferenceResolver must NOT call %s; inference is read-only", call.Method)
		}
	}
}

// =============================================================================
// D. Table-driven inference chain tests
// =============================================================================

// TestInferenceResolver_InferenceChain is a table-driven test covering the
// full matrix of inference scenarios: single ecosystem, unambiguous app,
// ambiguous app, and missing app.
//
// RED: All cases fail because InferenceResolver does not exist yet.
func TestInferenceResolver_InferenceChain(t *testing.T) {
	tests := []struct {
		name         string
		setupStore   func(t *testing.T) *db.MockDataStore
		filter       models.WorkspaceFilter
		wantErr      bool
		wantAmbig    bool
		wantNotFound bool
		wantAppName  string
		wantEcoName  string
	}{
		{
			name: "single ecosystem: app+workspace flags resolve without ecosystem flag",
			setupStore: func(t *testing.T) *db.MockDataStore {
				store, _, _, _, _ := setupSingleEcosystemStore(t)
				return store
			},
			filter: models.WorkspaceFilter{
				AppName:       "myapi",
				WorkspaceName: "main",
			},
			wantErr:     false,
			wantAppName: "myapi",
			wantEcoName: "production",
		},
		{
			name: "single ecosystem: full flags work too",
			setupStore: func(t *testing.T) *db.MockDataStore {
				store, _, _, _, _ := setupSingleEcosystemStore(t)
				return store
			},
			filter: models.WorkspaceFilter{
				EcosystemName: "production",
				DomainName:    "backend",
				AppName:       "myapi",
				WorkspaceName: "main",
			},
			wantErr:     false,
			wantAppName: "myapi",
			wantEcoName: "production",
		},
		{
			name: "unambiguous app: cascade infers domain and ecosystem",
			setupStore: func(t *testing.T) *db.MockDataStore {
				return setupMixedAppStore(t)
			},
			filter: models.WorkspaceFilter{
				AppName:       "frontend",
				WorkspaceName: "staging",
			},
			wantErr:     false,
			wantAppName: "frontend",
			wantEcoName: "platform",
		},
		{
			name: "ambiguous app across 2 ecosystems: returns AmbiguousError",
			setupStore: func(t *testing.T) *db.MockDataStore {
				return setupMultiEcosystemStore(t)
			},
			filter: models.WorkspaceFilter{
				AppName:       "api",
				WorkspaceName: "dev",
			},
			wantErr:   true,
			wantAmbig: true,
		},
		{
			name: "nonexistent app: returns ErrNoWorkspaceFound",
			setupStore: func(t *testing.T) *db.MockDataStore {
				store, _, _, _, _ := setupSingleEcosystemStore(t)
				return store
			},
			filter: models.WorkspaceFilter{
				AppName:       "does-not-exist",
				WorkspaceName: "main",
			},
			wantErr:      true,
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.setupStore(t)
			resolver := NewInferenceResolver(store)

			result, err := resolver.ResolveWithInference(tt.filter)

			if tt.wantErr {
				require.Error(t, err)
				if tt.wantAmbig {
					_, ok := IsAmbiguousError(err)
					assert.True(t, ok,
						"expected AmbiguousError, got %T: %v", err, err)
				}
				if tt.wantNotFound {
					assert.True(t, IsNoWorkspaceFoundError(err),
						"expected ErrNoWorkspaceFound, got %T: %v", err, err)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			if tt.wantAppName != "" {
				assert.Equal(t, tt.wantAppName, result.App.Name)
			}
			if tt.wantEcoName != "" {
				assert.Equal(t, tt.wantEcoName, result.Ecosystem.Name)
			}
		})
	}
}

// =============================================================================
// E. NewInferenceResolver constructor test
// =============================================================================

// TestNewInferenceResolver_ReturnsNonNil verifies the constructor compiles
// and returns a non-nil resolver.
//
// RED: Fails because NewInferenceResolver does not exist yet.
func TestNewInferenceResolver_ReturnsNonNil(t *testing.T) {
	store := db.NewMockDataStore()
	resolver := NewInferenceResolver(store)
	assert.NotNil(t, resolver, "NewInferenceResolver should return a non-nil resolver")
}
