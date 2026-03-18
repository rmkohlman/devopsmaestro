// Package resolver_test contains Phase 2 RED tests for the hierarchical CA certs resolver.
//
// RED PHASE (v0.56.0): This file tests pkg/cacerts/resolver, which does NOT
// have a full implementation yet. The interface.go file defines the types,
// but the implementation file (hierarchy.go) does NOT EXIST YET.
//
// These tests WILL NOT COMPILE until the following is implemented:
//
//   - pkg/cacerts/resolver/hierarchy.go
//     (HierarchyCACertsResolver, NewHierarchyCACertsResolver, Resolve)
//
// Additionally, the models and DB must be extended with:
//   - models.Ecosystem.CACerts   sql.NullString  (JSON: []CACertConfig)
//   - models.Domain.CACerts      sql.NullString  (JSON: []CACertConfig)
//   - AppBuildConfig.CACerts     []CACertConfig  (within app's build_config JSON)
//
// Design contract (from v0.56.0 sprint plan):
//
//	CACertsResolution {
//	    Certs   []CACertEntry             // Merged result — what build actually uses
//	    Sources map[string]HierarchyLevel // Per-name provenance (for --effective display)
//	    Path    []CACertsResolutionStep   // Full resolution path (all 5 levels, always present)
//	}
//	CACertsResolutionStep {
//	    Level HierarchyLevel
//	    Name  string
//	    Certs []CACertEntry // Certs at that specific level (may be empty)
//	    Found bool          // true if this level had any certs
//	}
//
// Cascade / override semantics:
//
//	global < ecosystem < domain < app < workspace
//	Override by cert Name — a more-specific cert with the same Name wins.
//	Unique-named certs from all levels are ALL included in the merged result.
//
// Security contract:
//   - Merged result must not exceed 10 certs (validated via ValidateCACerts())
//   - Cert Name must pass CACertConfig.Validate() (name regex, 64-char limit)
package resolver_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	// RED: this package's implementation (hierarchy.go) does not exist yet.
	// The interface.go file provides the types, but NewHierarchyCACertsResolver
	// will be undefined until hierarchy.go is created.
	"devopsmaestro/pkg/cacerts/resolver"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test helpers — mock store wiring
// =============================================================================

// caCertsMockStore is a thin wrapper around db.MockDataStore that adds
// CA certs fields to ecosystem, domain, app, and workspace models.
// It mirrors the buildArgsMockStore pattern in pkg/buildargs/resolver/resolver_test.go.
type caCertsMockStore struct {
	*db.MockDataStore
}

func newCACertsMockStore() *caCertsMockStore {
	return &caCertsMockStore{
		MockDataStore: db.NewMockDataStore(),
	}
}

// makeCACertJSON serialises a slice of CACertConfig to JSON for DB storage.
func makeCACertJSON(certs []models.CACertConfig) string {
	b, _ := json.Marshal(certs)
	return string(b)
}

// addEcosystem adds an ecosystem with optional CA certs.
// RED: Ecosystem.CACerts sql.NullString does not exist yet (v0.56.0 WI).
func (m *caCertsMockStore) addEcosystem(id int, name string, certs []models.CACertConfig) {
	eco := &models.Ecosystem{ID: id, Name: name}
	if len(certs) > 0 {
		// RED: Ecosystem.CACerts field does not exist yet
		eco.CACerts = sql.NullString{String: makeCACertJSON(certs), Valid: true}
	}
	m.MockDataStore.Ecosystems[name] = eco
}

// addDomain adds a domain with optional CA certs.
// RED: Domain.CACerts sql.NullString does not exist yet (v0.56.0 WI).
func (m *caCertsMockStore) addDomain(id, ecosystemID int, name string, certs []models.CACertConfig) {
	d := &models.Domain{ID: id, EcosystemID: ecosystemID, Name: name}
	if len(certs) > 0 {
		// RED: Domain.CACerts field does not exist yet
		d.CACerts = sql.NullString{String: makeCACertJSON(certs), Valid: true}
	}
	m.MockDataStore.Domains[id] = d
}

// addApp adds an app with optional CA certs in its build config.
// RED: AppBuildConfig.CACerts within the stored JSON does not yet cascade
// from app level. The app's BuildConfig JSON must include a "caCerts" key.
func (m *caCertsMockStore) addApp(id, domainID int, name string, certs []models.CACertConfig) {
	a := &models.App{ID: id, DomainID: domainID, Name: name, Path: "/code/" + name}
	if len(certs) > 0 {
		buildConfig := models.AppBuildConfig{CACerts: certs}
		b, _ := json.Marshal(buildConfig)
		a.BuildConfig = sql.NullString{String: string(b), Valid: true}
	}
	m.MockDataStore.Apps[id] = a
}

// addWorkspace adds a workspace with optional CA certs in its build config.
func (m *caCertsMockStore) addWorkspace(id, appID int, name string, certs []models.CACertConfig) {
	ws := &models.Workspace{ID: id, AppID: appID, Name: name}
	if len(certs) > 0 {
		devBuildConfig := models.DevBuildConfig{CACerts: certs}
		b, _ := json.Marshal(devBuildConfig)
		ws.BuildConfig = sql.NullString{String: string(b), Valid: true}
	}
	m.MockDataStore.Workspaces[id] = ws
}

// =============================================================================
// Tests: resolver.HierarchyCACertsResolver — all 5 levels, merge semantics
// =============================================================================

// TestCACertsResolver_WorkspaceOnly verifies that when only the workspace has
// CA certs (no higher levels have any), the workspace certs appear in the result.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver does not exist yet.
func TestCACertsResolver_WorkspaceOnly(t *testing.T) {
	store := newCACertsMockStore()

	store.addEcosystem(1, "corp-eco", nil)
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", []models.CACertConfig{
		{Name: "corp-root-ca", VaultSecret: "corp-root-ca-pem"},
	})

	// RED: NewHierarchyCACertsResolver does not exist yet
	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1 /*workspaceID*/)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	require.Len(t, resolution.Certs, 1,
		"workspace-only: should have exactly 1 cert in result")
	assert.Equal(t, "corp-root-ca", resolution.Certs[0].Name,
		"workspace-only: cert name should match")
	assert.Equal(t, "corp-root-ca-pem", resolution.Certs[0].VaultSecret,
		"workspace-only: vault secret should match")
	assert.Equal(t, resolver.LevelWorkspace, resolution.Certs[0].Source,
		"workspace-only: cert source should be LevelWorkspace")
}

// TestCACertsResolver_EcosystemCascade verifies that ecosystem-level certs
// cascade down to the workspace when no more-specific level overrides them.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver does not exist yet.
func TestCACertsResolver_EcosystemCascade(t *testing.T) {
	store := newCACertsMockStore()

	// Only ecosystem has certs; workspace has none
	store.addEcosystem(1, "corp-eco", []models.CACertConfig{
		{Name: "corp-root-ca", VaultSecret: "eco-root-ca-pem"},
		{Name: "corp-int-ca", VaultSecret: "eco-int-ca-pem"},
	})
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", nil)

	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	require.Len(t, resolution.Certs, 2,
		"ecosystem cascade: both ecosystem certs should appear in merged result")

	// Build a name→entry map for order-independent assertions
	certsByName := make(map[string]resolver.CACertEntry)
	for _, c := range resolution.Certs {
		certsByName[c.Name] = c
	}

	require.Contains(t, certsByName, "corp-root-ca")
	assert.Equal(t, resolver.LevelEcosystem, certsByName["corp-root-ca"].Source,
		"ecosystem cascade: corp-root-ca source should be LevelEcosystem")

	require.Contains(t, certsByName, "corp-int-ca")
	assert.Equal(t, resolver.LevelEcosystem, certsByName["corp-int-ca"].Source,
		"ecosystem cascade: corp-int-ca source should be LevelEcosystem")
}

// TestCACertsResolver_DomainOverridesEcosystem verifies that a domain cert with
// the same Name as an ecosystem cert replaces the ecosystem cert in the result.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver does not exist yet.
func TestCACertsResolver_DomainOverridesEcosystem(t *testing.T) {
	store := newCACertsMockStore()

	store.addEcosystem(1, "corp-eco", []models.CACertConfig{
		// ecosystem provides corp-root-ca pointing at eco-level vault secret
		{Name: "corp-root-ca", VaultSecret: "eco-root-ca-pem"},
	})
	store.addDomain(1, 1, "data-sci", []models.CACertConfig{
		// domain overrides corp-root-ca with a domain-specific vault secret
		{Name: "corp-root-ca", VaultSecret: "domain-root-ca-pem"},
	})
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", nil)

	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	// Only 1 cert: domain overrides ecosystem by same Name
	require.Len(t, resolution.Certs, 1,
		"domain-overrides-ecosystem: same-named cert should appear only once in merged result")
	assert.Equal(t, "corp-root-ca", resolution.Certs[0].Name)
	assert.Equal(t, "domain-root-ca-pem", resolution.Certs[0].VaultSecret,
		"domain-overrides-ecosystem: domain vault secret should win over ecosystem")
	assert.Equal(t, resolver.LevelDomain, resolution.Certs[0].Source,
		"domain-overrides-ecosystem: source should be LevelDomain")
}

// TestCACertsResolver_AppOverridesDomain verifies that an app cert with the
// same Name as a domain cert replaces the domain cert.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver / AppBuildConfig.CACerts
// does not exist yet.
func TestCACertsResolver_AppOverridesDomain(t *testing.T) {
	store := newCACertsMockStore()

	store.addEcosystem(1, "corp-eco", nil)
	store.addDomain(1, 1, "data-sci", []models.CACertConfig{
		{Name: "corp-root-ca", VaultSecret: "domain-root-ca-pem"},
	})
	store.addApp(1, 1, "ml-api", []models.CACertConfig{
		// App overrides corp-root-ca at the app level
		{Name: "corp-root-ca", VaultSecret: "app-root-ca-pem"},
	})
	store.addWorkspace(1, 1, "dev", nil)

	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	require.Len(t, resolution.Certs, 1,
		"app-overrides-domain: only 1 cert — app wins over domain by same Name")
	assert.Equal(t, "app-root-ca-pem", resolution.Certs[0].VaultSecret,
		"app-overrides-domain: app vault secret should win over domain")
	assert.Equal(t, resolver.LevelApp, resolution.Certs[0].Source,
		"app-overrides-domain: source should be LevelApp")
}

// TestCACertsResolver_WorkspaceOverridesAll verifies the full override chain:
// when all 5 levels define a cert with the same Name, the workspace cert wins.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver / fields do not exist yet.
func TestCACertsResolver_WorkspaceOverridesAll(t *testing.T) {
	store := newCACertsMockStore()

	certName := "corp-root-ca"
	store.addEcosystem(1, "corp-eco", []models.CACertConfig{
		{Name: certName, VaultSecret: "global-root-ca-pem"},
	})
	store.addDomain(1, 1, "data-sci", []models.CACertConfig{
		{Name: certName, VaultSecret: "eco-root-ca-pem"},
	})
	store.addApp(1, 1, "ml-api", []models.CACertConfig{
		{Name: certName, VaultSecret: "domain-root-ca-pem"},
	})
	store.addWorkspace(1, 1, "dev", []models.CACertConfig{
		// workspace is most specific — must win
		{Name: certName, VaultSecret: "workspace-root-ca-pem"},
	})

	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	require.Len(t, resolution.Certs, 1,
		"workspace-overrides-all: all same-named certs collapse to exactly 1")
	assert.Equal(t, "workspace-root-ca-pem", resolution.Certs[0].VaultSecret,
		"workspace-overrides-all: workspace vault secret must win over all other levels")
	assert.Equal(t, resolver.LevelWorkspace, resolution.Certs[0].Source,
		"workspace-overrides-all: cert source must be LevelWorkspace")
}

// TestCACertsResolver_MultipleLevelsMerge verifies additive merge semantics:
// when each level has a cert with a UNIQUE Name, all certs appear in the result.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver / fields do not exist yet.
func TestCACertsResolver_MultipleLevelsMerge(t *testing.T) {
	store := newCACertsMockStore()

	store.addEcosystem(1, "corp-eco", []models.CACertConfig{
		{Name: "eco-ca", VaultSecret: "eco-ca-pem"},
	})
	store.addDomain(1, 1, "data-sci", []models.CACertConfig{
		{Name: "domain-ca", VaultSecret: "domain-ca-pem"},
	})
	store.addApp(1, 1, "ml-api", []models.CACertConfig{
		{Name: "app-ca", VaultSecret: "app-ca-pem"},
	})
	store.addWorkspace(1, 1, "dev", []models.CACertConfig{
		{Name: "workspace-ca", VaultSecret: "workspace-ca-pem"},
	})

	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	// All 4 certs have unique names — all should appear in result
	require.Len(t, resolution.Certs, 4,
		"multiple-levels-merge: all distinct-named certs from all levels should be present")

	certsByName := make(map[string]resolver.CACertEntry)
	for _, c := range resolution.Certs {
		certsByName[c.Name] = c
	}

	assert.Contains(t, certsByName, "eco-ca", "eco-ca should be in merged result")
	assert.Contains(t, certsByName, "domain-ca", "domain-ca should be in merged result")
	assert.Contains(t, certsByName, "app-ca", "app-ca should be in merged result")
	assert.Contains(t, certsByName, "workspace-ca", "workspace-ca should be in merged result")

	assert.Equal(t, resolver.LevelEcosystem, certsByName["eco-ca"].Source)
	assert.Equal(t, resolver.LevelDomain, certsByName["domain-ca"].Source)
	assert.Equal(t, resolver.LevelApp, certsByName["app-ca"].Source)
	assert.Equal(t, resolver.LevelWorkspace, certsByName["workspace-ca"].Source)
}

// TestCACertsResolver_GlobalCerts verifies that global-level CA certs
// (stored in the defaults table under key "ca-certs") flow down to workspace.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver / global level does not exist yet.
func TestCACertsResolver_GlobalCerts(t *testing.T) {
	store := newCACertsMockStore()

	// Set global CA certs in defaults table
	globalCerts := []models.CACertConfig{
		{Name: "global-root-ca", VaultSecret: "global-root-ca-pem"},
	}
	b, _ := json.Marshal(globalCerts)
	if store.MockDataStore.Defaults == nil {
		store.MockDataStore.Defaults = make(map[string]string)
	}
	store.MockDataStore.Defaults["ca-certs"] = string(b)

	store.addEcosystem(1, "corp-eco", nil)
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", nil)

	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	require.Len(t, resolution.Certs, 1,
		"global-certs: global cert should cascade to workspace when no overrides")
	assert.Equal(t, "global-root-ca", resolution.Certs[0].Name)
	assert.Equal(t, resolver.LevelGlobal, resolution.Certs[0].Source,
		"global-certs: source must be LevelGlobal")
}

// TestCACertsResolver_EmptyLevels verifies that when no CA certs exist at any
// level, Resolve returns an empty (non-nil) slice.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver does not exist yet.
func TestCACertsResolver_EmptyLevels(t *testing.T) {
	store := newCACertsMockStore()

	store.addEcosystem(1, "corp-eco", nil)
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", nil)

	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.NotNil(t, resolution.Certs,
		"empty-levels: Certs must be a non-nil empty slice, not nil")
	assert.Empty(t, resolution.Certs,
		"empty-levels: Certs must be empty when no certs at any level")
	assert.NotNil(t, resolution.Sources,
		"empty-levels: Sources must be a non-nil empty map, not nil")
	assert.Empty(t, resolution.Sources,
		"empty-levels: Sources must be empty when no certs at any level")
}

// TestCACertsResolver_SourceTracking verifies that the Sources map correctly
// records which HierarchyLevel each cert Name came from.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver does not exist yet.
func TestCACertsResolver_SourceTracking(t *testing.T) {
	store := newCACertsMockStore()

	store.addEcosystem(1, "corp-eco", []models.CACertConfig{
		{Name: "eco-ca", VaultSecret: "eco-ca-pem"},
	})
	store.addDomain(1, 1, "data-sci", []models.CACertConfig{
		// domain overrides eco-ca AND adds its own
		{Name: "eco-ca", VaultSecret: "domain-overrides-eco-ca-pem"},
		{Name: "domain-ca", VaultSecret: "domain-ca-pem"},
	})
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", []models.CACertConfig{
		{Name: "workspace-ca", VaultSecret: "workspace-ca-pem"},
	})

	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	require.NotNil(t, resolution.Sources, "Sources map must not be nil")

	// eco-ca was overridden by domain → source is LevelDomain
	assert.Equal(t, resolver.LevelDomain, resolution.Sources["eco-ca"],
		"source-tracking: eco-ca was overridden by domain, source must be LevelDomain")

	// domain-ca came from domain
	assert.Equal(t, resolver.LevelDomain, resolution.Sources["domain-ca"],
		"source-tracking: domain-ca source must be LevelDomain")

	// workspace-ca came from workspace
	assert.Equal(t, resolver.LevelWorkspace, resolution.Sources["workspace-ca"],
		"source-tracking: workspace-ca source must be LevelWorkspace")

	// Total: 3 unique-named certs (eco-ca overridden by domain, domain-ca, workspace-ca)
	require.Len(t, resolution.Certs, 3,
		"source-tracking: 3 distinct-named certs after eco-ca override")
}

// =============================================================================
// Table-driven: HierarchyLevel constant values
// =============================================================================

// TestCACertsHierarchyLevel_Constants verifies the HierarchyLevel constants
// are defined and their String() representations match expected display strings.
func TestCACertsHierarchyLevel_Constants(t *testing.T) {
	tests := []struct {
		level   resolver.HierarchyLevel
		wantStr string
	}{
		{resolver.LevelGlobal, "global"},
		{resolver.LevelEcosystem, "ecosystem"},
		{resolver.LevelDomain, "domain"},
		{resolver.LevelApp, "app"},
		{resolver.LevelWorkspace, "workspace"},
	}

	for _, tt := range tests {
		t.Run(tt.wantStr, func(t *testing.T) {
			assert.Equal(t, tt.wantStr, tt.level.String(),
				"HierarchyLevel.String() must return %q for level %d", tt.wantStr, tt.level)
		})
	}
}

// =============================================================================
// Table-driven: comprehensive cascade scenarios
// =============================================================================

// TestCACertsResolver_CascadeScenarios covers additional merge edge-cases in a
// table-driven format, mirroring the pattern from buildargs resolver tests.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver does not exist yet.
func TestCACertsResolver_CascadeScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setupStore  func(*caCertsMockStore)
		workspaceID int
		wantNames   []string // expected cert names in merged result (order-independent)
		wantErr     bool
	}{
		{
			name: "ecosystem only cascades to workspace",
			setupStore: func(s *caCertsMockStore) {
				s.addEcosystem(1, "eco", []models.CACertConfig{
					{Name: "eco-ca", VaultSecret: "eco-ca-pem"},
				})
				s.addDomain(1, 1, "dom", nil)
				s.addApp(1, 1, "app", nil)
				s.addWorkspace(1, 1, "ws", nil)
			},
			workspaceID: 1,
			wantNames:   []string{"eco-ca"},
		},
		{
			name: "domain overrides ecosystem same-named cert",
			setupStore: func(s *caCertsMockStore) {
				s.addEcosystem(1, "eco", []models.CACertConfig{
					{Name: "shared-ca", VaultSecret: "eco-shared-pem"},
				})
				s.addDomain(1, 1, "dom", []models.CACertConfig{
					{Name: "shared-ca", VaultSecret: "domain-shared-pem"},
				})
				s.addApp(1, 1, "app", nil)
				s.addWorkspace(1, 1, "ws", nil)
			},
			workspaceID: 1,
			wantNames:   []string{"shared-ca"}, // only 1, domain wins
		},
		{
			name: "app overrides domain",
			setupStore: func(s *caCertsMockStore) {
				s.addEcosystem(1, "eco", nil)
				s.addDomain(1, 1, "dom", []models.CACertConfig{
					{Name: "ca", VaultSecret: "domain-pem"},
				})
				s.addApp(1, 1, "app", []models.CACertConfig{
					{Name: "ca", VaultSecret: "app-pem"},
				})
				s.addWorkspace(1, 1, "ws", nil)
			},
			workspaceID: 1,
			wantNames:   []string{"ca"}, // app wins
		},
		{
			name: "workspace overrides app",
			setupStore: func(s *caCertsMockStore) {
				s.addEcosystem(1, "eco", nil)
				s.addDomain(1, 1, "dom", nil)
				s.addApp(1, 1, "app", []models.CACertConfig{
					{Name: "ca", VaultSecret: "app-pem"},
				})
				s.addWorkspace(1, 1, "ws", []models.CACertConfig{
					{Name: "ca", VaultSecret: "ws-pem"},
				})
			},
			workspaceID: 1,
			wantNames:   []string{"ca"}, // workspace wins
		},
		{
			name: "completely empty hierarchy returns empty slice",
			setupStore: func(s *caCertsMockStore) {
				s.addEcosystem(1, "eco", nil)
				s.addDomain(1, 1, "dom", nil)
				s.addApp(1, 1, "app", nil)
				s.addWorkspace(1, 1, "ws", nil)
			},
			workspaceID: 1,
			wantNames:   []string{},
		},
		{
			name: "additive merge across all levels — distinct names",
			setupStore: func(s *caCertsMockStore) {
				s.addEcosystem(1, "eco", []models.CACertConfig{
					{Name: "eco-ca", VaultSecret: "eco-pem"},
				})
				s.addDomain(1, 1, "dom", []models.CACertConfig{
					{Name: "domain-ca", VaultSecret: "dom-pem"},
				})
				s.addApp(1, 1, "app", []models.CACertConfig{
					{Name: "app-ca", VaultSecret: "app-pem"},
				})
				s.addWorkspace(1, 1, "ws", []models.CACertConfig{
					{Name: "ws-ca", VaultSecret: "ws-pem"},
				})
			},
			workspaceID: 1,
			wantNames:   []string{"eco-ca", "domain-ca", "app-ca", "ws-ca"},
		},
		{
			name: "workspace not found returns error",
			setupStore: func(s *caCertsMockStore) {
				// no workspace with ID 99
			},
			workspaceID: 99,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newCACertsMockStore()
			tt.setupStore(store)

			// RED: NewHierarchyCACertsResolver does not exist yet
			r := resolver.NewHierarchyCACertsResolver(store)
			ctx := context.Background()

			resolution, err := r.Resolve(ctx, tt.workspaceID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resolution)

			// Collect actual cert names for order-independent comparison
			gotNames := make([]string, 0, len(resolution.Certs))
			for _, c := range resolution.Certs {
				gotNames = append(gotNames, c.Name)
			}
			assert.ElementsMatch(t, tt.wantNames, gotNames,
				"merged cert names must match expected (order-independent)")
		})
	}
}

// TestCACertsResolver_Path_ShowsAllLevels verifies that the Path slice always
// contains exactly 5 entries (one per level), with Found=true for levels that
// had certs and Found=false for those that didn't.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver does not exist yet.
func TestCACertsResolver_Path_ShowsAllLevels(t *testing.T) {
	store := newCACertsMockStore()

	// Only ecosystem and workspace have certs
	store.addEcosystem(1, "corp-eco", []models.CACertConfig{
		{Name: "eco-ca", VaultSecret: "eco-ca-pem"},
	})
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", []models.CACertConfig{
		{Name: "workspace-ca", VaultSecret: "workspace-ca-pem"},
	})

	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)

	// Path must always contain an entry for every level
	require.Len(t, resolution.Path, 5,
		"Path must always have exactly 5 entries — one per hierarchy level")

	// Collect all levels present in path for assertion
	pathLevels := make(map[resolver.HierarchyLevel]bool)
	foundFlags := make(map[resolver.HierarchyLevel]bool)
	for _, step := range resolution.Path {
		pathLevels[step.Level] = true
		foundFlags[step.Level] = step.Found
	}

	// All 5 levels must be represented
	assert.True(t, pathLevels[resolver.LevelGlobal], "Path must include LevelGlobal step")
	assert.True(t, pathLevels[resolver.LevelEcosystem], "Path must include LevelEcosystem step")
	assert.True(t, pathLevels[resolver.LevelDomain], "Path must include LevelDomain step")
	assert.True(t, pathLevels[resolver.LevelApp], "Path must include LevelApp step")
	assert.True(t, pathLevels[resolver.LevelWorkspace], "Path must include LevelWorkspace step")

	// Found flag should reflect whether that level had any certs
	assert.True(t, foundFlags[resolver.LevelEcosystem],
		"LevelEcosystem step must have Found=true (had eco-ca)")
	assert.True(t, foundFlags[resolver.LevelWorkspace],
		"LevelWorkspace step must have Found=true (had workspace-ca)")
	assert.False(t, foundFlags[resolver.LevelGlobal],
		"LevelGlobal step must have Found=false (no global certs set)")
	assert.False(t, foundFlags[resolver.LevelDomain],
		"LevelDomain step must have Found=false (no domain certs set)")
	assert.False(t, foundFlags[resolver.LevelApp],
		"LevelApp step must have Found=false (no app certs set)")
}

// TestCACertsResolver_MaxCertsEnforced verifies that when the merged cascade
// result would exceed 10 certs, Resolve returns an error.
//
// RED: WILL NOT COMPILE — HierarchyCACertsResolver does not exist yet.
func TestCACertsResolver_MaxCertsEnforced(t *testing.T) {
	store := newCACertsMockStore()

	// Create 11 distinct-named certs spread across ecosystem (6) and workspace (5)
	// Total = 11 > max of 10 → should fail
	ecoCerts := make([]models.CACertConfig, 6)
	for i := range ecoCerts {
		ecoCerts[i] = models.CACertConfig{
			Name:        fmt.Sprintf("eco-ca-%d", i),
			VaultSecret: fmt.Sprintf("eco-ca-%d-pem", i),
		}
	}
	wsCerts := make([]models.CACertConfig, 5)
	for i := range wsCerts {
		wsCerts[i] = models.CACertConfig{
			Name:        fmt.Sprintf("ws-ca-%d", i),
			VaultSecret: fmt.Sprintf("ws-ca-%d-pem", i),
		}
	}

	store.addEcosystem(1, "corp-eco", ecoCerts)
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", wsCerts)

	r := resolver.NewHierarchyCACertsResolver(store)
	ctx := context.Background()

	_, err := r.Resolve(ctx, 1)

	require.Error(t, err,
		"max-certs-enforced: Resolve should return error when merged result exceeds 10 certs")
	assert.Contains(t, err.Error(), "10",
		"max-certs-enforced: error message should mention the 10-cert maximum")
}
