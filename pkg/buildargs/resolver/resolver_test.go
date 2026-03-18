// Package resolver_test contains Phase 2 RED tests for the hierarchical build args resolver.
//
// RED PHASE: This file tests pkg/buildargs/resolver, which does NOT EXIST YET.
// These tests WILL NOT COMPILE until the following are implemented:
//
//   - pkg/buildargs/resolver/interface.go  (HierarchyLevel, BuildArgsResolution, BuildArgsStep, BuildArgsResolver)
//   - pkg/buildargs/resolver/hierarchy.go  (HierarchyBuildArgsResolver, NewHierarchyBuildArgsResolver, Resolve)
//
// Design contract (from v0.55.0 sprint plan):
//
//	BuildArgsResolution {
//	    Args    map[string]string         // Merged result — what the build actually uses
//	    Sources map[string]HierarchyLevel // Per-key provenance (for --effective display)
//	    Path    []BuildArgsStep           // Full resolution path (all 5 levels, always present)
//	}
//	BuildArgsStep {
//	    Level HierarchyLevel
//	    Name  string
//	    Args  map[string]string // Args at that specific level (may be empty)
//	    Found bool              // true if this level had any build args
//	}
//
// Cascade order (lowest → highest precedence):
//
//	global < ecosystem < domain < app < workspace
//
// Security contract:
//   - Keys failing ValidateEnvKey() are silently filtered from cascade output
//   - Keys in IsDangerousEnvVar() denylist are silently filtered from cascade output
//   - Both checks are defence-in-depth (data was validated at entry points too)
package resolver_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	// RED: this package does not exist yet — import will fail to compile
	"devopsmaestro/pkg/buildargs/resolver"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test helpers — minimal mock wiring that mirrors hierarchy_test.go pattern
// =============================================================================

// buildArgsMockStore is a thin wrapper around db.MockDataStore that adds
// build-args fields to ecosystem, domain, app, and workspace models.
// It mirrors the local MockDataStore in hierarchy_test.go but uses the
// shared db.MockDataStore so we stay consistent across the code-base.
type buildArgsMockStore struct {
	*db.MockDataStore
}

func newBuildArgsMockStore() *buildArgsMockStore {
	return &buildArgsMockStore{
		MockDataStore: db.NewMockDataStore(),
	}
}

// addEcosystem adds an ecosystem with optional build args.
func (m *buildArgsMockStore) addEcosystem(id int, name string, args map[string]string) {
	eco := &models.Ecosystem{ID: id, Name: name}
	if len(args) > 0 {
		// RED: Ecosystem.BuildArgs (sql.NullString) does not exist yet
		b, _ := json.Marshal(args)
		eco.BuildArgs = sql.NullString{String: string(b), Valid: true}
	}
	m.MockDataStore.Ecosystems[name] = eco
}

// addDomain adds a domain with optional build args.
func (m *buildArgsMockStore) addDomain(id, ecosystemID int, name string, args map[string]string) {
	d := &models.Domain{ID: id, EcosystemID: ecosystemID, Name: name}
	if len(args) > 0 {
		// RED: Domain.BuildArgs (sql.NullString) does not exist yet
		b, _ := json.Marshal(args)
		d.BuildArgs = sql.NullString{String: string(b), Valid: true}
	}
	m.MockDataStore.Domains[id] = d
}

// addApp adds an app with optional build args.
// RED: models.App does not yet have a BuildConfig field that preserves Args-only builds.
// Once WI-2 is implemented, args will be stored in app.BuildConfig.
func (m *buildArgsMockStore) addApp(id, domainID int, name string, args map[string]string) {
	a := &models.App{ID: id, DomainID: domainID, Name: name}
	if len(args) > 0 {
		// RED: AppBuildConfig args persistence fix (WI-2) not yet implemented.
		// For now we wire through BuildConfig JSON as if WI-2 is complete.
		b, _ := json.Marshal(map[string]interface{}{"args": args})
		a.BuildConfig = sql.NullString{String: string(b), Valid: true}
	}
	m.MockDataStore.Apps[id] = a
}

// addWorkspace adds a workspace with optional build args.
// RED: Workspace.BuildConfig column (WI-1) does not yet exist.
func (m *buildArgsMockStore) addWorkspace(id, appID int, name string, args map[string]string) {
	ws := &models.Workspace{ID: id, AppID: appID, Name: name}
	if len(args) > 0 {
		// RED: Workspace.BuildConfig (sql.NullString) does not exist yet
		b, _ := json.Marshal(map[string]interface{}{"args": args})
		ws.BuildConfig = sql.NullString{String: string(b), Valid: true}
	}
	m.MockDataStore.Workspaces[id] = ws
}

// setGlobalBuildArgs stores global build args in the defaults table under key "build-args".
func (m *buildArgsMockStore) setGlobalBuildArgs(args map[string]string) {
	b, _ := json.Marshal(args)
	if m.MockDataStore.Defaults == nil {
		m.MockDataStore.Defaults = make(map[string]string)
	}
	m.MockDataStore.Defaults["build-args"] = string(b)
}

// =============================================================================
// Tests: resolver.HierarchyBuildArgsResolver — all 5 levels, merge semantics
// =============================================================================

// TestResolve_AllLevels_MostSpecificWins verifies that when PIP_INDEX_URL is set
// at every hierarchy level, the workspace (most specific) value wins.
func TestResolve_AllLevels_MostSpecificWins(t *testing.T) {
	// RED: HierarchyBuildArgsResolver does not exist yet
	store := newBuildArgsMockStore()

	// Arrange: every level sets the same key with a unique value
	store.setGlobalBuildArgs(map[string]string{
		"PIP_INDEX_URL": "global-value",
	})
	store.addEcosystem(1, "corp-eco", map[string]string{
		"PIP_INDEX_URL": "eco-value",
	})
	store.addDomain(1, 1, "data-sci", map[string]string{
		"PIP_INDEX_URL": "domain-value",
	})
	store.addApp(1, 1, "ml-api", map[string]string{
		"PIP_INDEX_URL": "app-value",
	})
	store.addWorkspace(1, 1, "dev", map[string]string{
		"PIP_INDEX_URL": "workspace-value", // most specific — should win
	})

	// RED: NewHierarchyBuildArgsResolver does not exist yet
	r := resolver.NewHierarchyBuildArgsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1 /*workspaceID*/)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, "workspace-value", resolution.Args["PIP_INDEX_URL"],
		"workspace value must win when set at every level")
	assert.Equal(t, resolver.LevelWorkspace, resolution.Sources["PIP_INDEX_URL"],
		"source for PIP_INDEX_URL must be LevelWorkspace")
}

// TestResolve_PartialCascade_EcosystemAndWorkspace verifies that when only
// ecosystem and workspace have build args, both sets of keys appear in the result.
func TestResolve_PartialCascade_EcosystemAndWorkspace(t *testing.T) {
	store := newBuildArgsMockStore()

	store.setGlobalBuildArgs(nil) // no global args
	store.addEcosystem(1, "corp-eco", map[string]string{
		"PIP_INDEX_URL": "https://internal.pypi.example",
	})
	store.addDomain(1, 1, "data-sci", nil) // no domain args
	store.addApp(1, 1, "ml-api", nil)      // no app args
	store.addWorkspace(1, 1, "dev", map[string]string{
		"DEBUG_BUILD": "true",
	})

	r := resolver.NewHierarchyBuildArgsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, "https://internal.pypi.example", resolution.Args["PIP_INDEX_URL"],
		"ecosystem arg must be present in merged result")
	assert.Equal(t, "true", resolution.Args["DEBUG_BUILD"],
		"workspace arg must be present in merged result")
	assert.Len(t, resolution.Args, 2,
		"merged result must contain exactly the 2 keys set across the hierarchy")
}

// TestResolve_EmptyCascade_ReturnsEmptyMap verifies that when no build args
// exist at any level, Resolve returns an empty (non-nil) map.
func TestResolve_EmptyCascade_ReturnsEmptyMap(t *testing.T) {
	store := newBuildArgsMockStore()

	// No build args at any level
	store.addEcosystem(1, "corp-eco", nil)
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", nil)

	r := resolver.NewHierarchyBuildArgsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.NotNil(t, resolution.Args,
		"Args must be a non-nil empty map, not nil")
	assert.Empty(t, resolution.Args,
		"Args must be empty when no build args are set anywhere")
}

// TestResolve_Override_WorkspaceOverridesEcosystem verifies the override
// semantics: a workspace value for PIP_INDEX_URL replaces the ecosystem value.
func TestResolve_Override_WorkspaceOverridesEcosystem(t *testing.T) {
	store := newBuildArgsMockStore()

	store.addEcosystem(1, "corp-eco", map[string]string{
		"PIP_INDEX_URL": "eco-value",
	})
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", map[string]string{
		"PIP_INDEX_URL": "ws-value", // overrides ecosystem
	})

	r := resolver.NewHierarchyBuildArgsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, "ws-value", resolution.Args["PIP_INDEX_URL"],
		"workspace value must override ecosystem value for the same key")
	assert.Equal(t, resolver.LevelWorkspace, resolution.Sources["PIP_INDEX_URL"],
		"source must be LevelWorkspace, not LevelEcosystem")
}

// TestResolve_Additive_DifferentKeysFromDifferentLevels verifies the additive
// merge: unique keys from all 5 levels all appear in the resolved result.
func TestResolve_Additive_DifferentKeysFromDifferentLevels(t *testing.T) {
	store := newBuildArgsMockStore()

	store.setGlobalBuildArgs(map[string]string{
		"KEY_A": "global",
	})
	store.addEcosystem(1, "corp-eco", map[string]string{
		"KEY_B": "eco",
	})
	store.addDomain(1, 1, "data-sci", map[string]string{
		"KEY_C": "domain",
	})
	store.addApp(1, 1, "ml-api", map[string]string{
		"KEY_D": "app",
	})
	store.addWorkspace(1, 1, "dev", map[string]string{
		"KEY_E": "workspace",
	})

	r := resolver.NewHierarchyBuildArgsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.Equal(t, "global", resolution.Args["KEY_A"], "global key must appear in result")
	assert.Equal(t, "eco", resolution.Args["KEY_B"], "ecosystem key must appear in result")
	assert.Equal(t, "domain", resolution.Args["KEY_C"], "domain key must appear in result")
	assert.Equal(t, "app", resolution.Args["KEY_D"], "app key must appear in result")
	assert.Equal(t, "workspace", resolution.Args["KEY_E"], "workspace key must appear in result")
	assert.Len(t, resolution.Args, 5,
		"all 5 distinct keys from 5 levels must be present")
}

// TestResolve_Provenance_SourcesTrackOrigin verifies that the Sources map
// correctly records which hierarchy level each key came from.
func TestResolve_Provenance_SourcesTrackOrigin(t *testing.T) {
	store := newBuildArgsMockStore()

	store.addEcosystem(1, "corp-eco", map[string]string{
		"PIP_INDEX_URL": "eco-pip",
	})
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", map[string]string{
		"DEBUG_BUILD": "true",
	})

	r := resolver.NewHierarchyBuildArgsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	require.NotNil(t, resolution.Sources,
		"Sources map must not be nil")
	assert.Equal(t, resolver.LevelEcosystem, resolution.Sources["PIP_INDEX_URL"],
		"PIP_INDEX_URL came from ecosystem")
	assert.Equal(t, resolver.LevelWorkspace, resolution.Sources["DEBUG_BUILD"],
		"DEBUG_BUILD came from workspace")
}

// TestResolve_DangerousKey_Filtered verifies that keys in the IsDangerousEnvVar()
// denylist are removed from the cascade output as a defence-in-depth measure.
func TestResolve_DangerousKey_Filtered(t *testing.T) {
	store := newBuildArgsMockStore()

	// Attempt to inject a dangerous key at the ecosystem level
	store.addEcosystem(1, "corp-eco", map[string]string{
		"LD_PRELOAD":    "/evil.so", // dangerous — must be filtered
		"PIP_INDEX_URL": "safe-value",
	})
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", nil)

	r := resolver.NewHierarchyBuildArgsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.NotContains(t, resolution.Args, "LD_PRELOAD",
		"LD_PRELOAD must be filtered by IsDangerousEnvVar() defence-in-depth check")
	assert.Contains(t, resolution.Args, "PIP_INDEX_URL",
		"safe keys must still pass through after dangerous key is filtered")
}

// TestResolve_InvalidKey_Filtered verifies that keys failing ValidateEnvKey()
// are removed from the cascade output — e.g. keys that start with a digit.
func TestResolve_InvalidKey_Filtered(t *testing.T) {
	store := newBuildArgsMockStore()

	// Attempt to store a malformed key at the ecosystem level
	store.addEcosystem(1, "corp-eco", map[string]string{
		"123-bad":     "bad-key", // invalid — must be filtered
		"VALID_KEY_A": "good",    // valid
	})
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", nil)

	r := resolver.NewHierarchyBuildArgsResolver(store)
	ctx := context.Background()

	resolution, err := r.Resolve(ctx, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	assert.NotContains(t, resolution.Args, "123-bad",
		"invalid key '123-bad' must be filtered by ValidateEnvKey() defence-in-depth")
	assert.Contains(t, resolution.Args, "VALID_KEY_A",
		"valid keys must still appear after invalid key is filtered")
}

// TestResolve_Path_ShowsAllLevels verifies that the Path slice always contains
// exactly 5 entries (one per level: global, ecosystem, domain, app, workspace),
// with Found=true for levels that had args and Found=false for those that didn't.
func TestResolve_Path_ShowsAllLevels(t *testing.T) {
	store := newBuildArgsMockStore()

	// Only ecosystem and workspace have build args
	store.setGlobalBuildArgs(nil)
	store.addEcosystem(1, "corp-eco", map[string]string{
		"PIP_INDEX_URL": "eco-value",
	})
	store.addDomain(1, 1, "data-sci", nil)
	store.addApp(1, 1, "ml-api", nil)
	store.addWorkspace(1, 1, "dev", map[string]string{
		"DEBUG_BUILD": "true",
	})

	r := resolver.NewHierarchyBuildArgsResolver(store)
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

	// Found flag should reflect whether that level had any args
	assert.True(t, foundFlags[resolver.LevelEcosystem],
		"LevelEcosystem step must have Found=true (had PIP_INDEX_URL)")
	assert.True(t, foundFlags[resolver.LevelWorkspace],
		"LevelWorkspace step must have Found=true (had DEBUG_BUILD)")
	assert.False(t, foundFlags[resolver.LevelGlobal],
		"LevelGlobal step must have Found=false (no global args set)")
	assert.False(t, foundFlags[resolver.LevelDomain],
		"LevelDomain step must have Found=false (no domain args set)")
	assert.False(t, foundFlags[resolver.LevelApp],
		"LevelApp step must have Found=false (no app args set)")
}

// =============================================================================
// Table-driven: HierarchyLevel constant values (mirrors theme resolver pattern)
// =============================================================================

// TestHierarchyLevel_Constants verifies the HierarchyLevel constants are defined
// and their String() representations match the expected display strings.
func TestHierarchyLevel_Constants(t *testing.T) {
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
// Table-driven: comprehensive merge / cascade scenarios
// =============================================================================

// TestResolve_CascadeScenarios covers additional merge edge-cases in a
// table-driven format, following the pattern from hierarchy_test.go.
func TestResolve_CascadeScenarios(t *testing.T) {
	tests := []struct {
		name        string
		setupStore  func(*buildArgsMockStore)
		workspaceID int
		wantArgs    map[string]string
		wantErr     bool
	}{
		{
			name: "global only",
			setupStore: func(s *buildArgsMockStore) {
				s.setGlobalBuildArgs(map[string]string{"GLOBAL_KEY": "gval"})
				s.addEcosystem(1, "eco", nil)
				s.addDomain(1, 1, "dom", nil)
				s.addApp(1, 1, "app", nil)
				s.addWorkspace(1, 1, "ws", nil)
			},
			workspaceID: 1,
			wantArgs:    map[string]string{"GLOBAL_KEY": "gval"},
		},
		{
			name: "ecosystem overrides global",
			setupStore: func(s *buildArgsMockStore) {
				s.setGlobalBuildArgs(map[string]string{"KEY": "global"})
				s.addEcosystem(1, "eco", map[string]string{"KEY": "ecosystem"})
				s.addDomain(1, 1, "dom", nil)
				s.addApp(1, 1, "app", nil)
				s.addWorkspace(1, 1, "ws", nil)
			},
			workspaceID: 1,
			wantArgs:    map[string]string{"KEY": "ecosystem"},
		},
		{
			name: "domain overrides ecosystem",
			setupStore: func(s *buildArgsMockStore) {
				s.addEcosystem(1, "eco", map[string]string{"KEY": "ecosystem"})
				s.addDomain(1, 1, "dom", map[string]string{"KEY": "domain"})
				s.addApp(1, 1, "app", nil)
				s.addWorkspace(1, 1, "ws", nil)
			},
			workspaceID: 1,
			wantArgs:    map[string]string{"KEY": "domain"},
		},
		{
			name: "app overrides domain",
			setupStore: func(s *buildArgsMockStore) {
				s.addEcosystem(1, "eco", nil)
				s.addDomain(1, 1, "dom", map[string]string{"KEY": "domain"})
				s.addApp(1, 1, "app", map[string]string{"KEY": "app"})
				s.addWorkspace(1, 1, "ws", nil)
			},
			workspaceID: 1,
			wantArgs:    map[string]string{"KEY": "app"},
		},
		{
			name: "workspace overrides app",
			setupStore: func(s *buildArgsMockStore) {
				s.addEcosystem(1, "eco", nil)
				s.addDomain(1, 1, "dom", nil)
				s.addApp(1, 1, "app", map[string]string{"KEY": "app"})
				s.addWorkspace(1, 1, "ws", map[string]string{"KEY": "workspace"})
			},
			workspaceID: 1,
			wantArgs:    map[string]string{"KEY": "workspace"},
		},
		{
			name: "workspace not found returns error",
			setupStore: func(s *buildArgsMockStore) {
				// no workspace with ID 99
			},
			workspaceID: 99,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newBuildArgsMockStore()
			tt.setupStore(store)

			r := resolver.NewHierarchyBuildArgsResolver(store)
			ctx := context.Background()

			resolution, err := r.Resolve(ctx, tt.workspaceID)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resolution)
			assert.Equal(t, tt.wantArgs, resolution.Args)
		})
	}
}
