package registry

import (
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// B1: Zot version mismatch — factory.go:22 and strategy.go:87 must use the same version.
// Both code paths must use the same version ("2.1.15").
// =============================================================================

// TestZotBinaryVersion_FactoryUsesCorrectVersion verifies that NewZotManager (used
// via factory.go:22) creates a DefaultBinaryManager with version "2.1.15".
func TestZotBinaryVersion_FactoryUsesCorrectVersion(t *testing.T) {
	tests := []struct {
		name        string
		wantVersion string
	}{
		{
			name:        "factory path must use zot 2.1.15",
			wantVersion: "2.1.15",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := RegistryConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      5001,
				Storage:   t.TempDir(),
			}

			mgr := NewZotManager(config)
			require.NotNil(t, mgr)

			// We are in the same package, so we can inspect the unexported field.
			// Cast binaryManager to *DefaultBinaryManager to read the version field.
			binMgr, ok := mgr.binaryManager.(*DefaultBinaryManager)
			require.True(t, ok, "binaryManager should be a *DefaultBinaryManager")

			assert.Equal(t, tt.wantVersion, binMgr.version,
				"NewZotManager (factory path) must use version %q, not %q",
				tt.wantVersion, binMgr.version,
			)
		})
	}
}

// TestZotBinaryVersion_StrategyAndFactoryAgree verifies that both code paths that
// create a ZotManager — NewZotManager (factory.go) and ZotStrategy.CreateManager
// (strategy.go) — use the same binary version.
// Both paths must agree on the same version ("2.1.15").
func TestZotBinaryVersion_StrategyAndFactoryAgree(t *testing.T) {
	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   t.TempDir(),
	}

	// Factory path: NewZotManager (factory.go line 22)
	factoryMgr := NewZotManager(config)
	require.NotNil(t, factoryMgr)

	factoryBinMgr, ok := factoryMgr.binaryManager.(*DefaultBinaryManager)
	require.True(t, ok, "factory binaryManager should be *DefaultBinaryManager")

	// Strategy path: ZotStrategy.CreateManager (strategy.go line 65)
	strategy := NewZotStrategy()
	reg := &models.Registry{
		ID:        1,
		Name:      "test-zot",
		Type:      "zot",
		Lifecycle: "persistent",
		Port:      5001,
	}

	strategyManager, err := strategy.CreateManager(reg)
	require.NoError(t, err)
	require.NotNil(t, strategyManager)

	// ZotStrategy.CreateManager returns a *ZotManager wrapped as ServiceManager
	zotMgr, ok := strategyManager.(*ZotManager)
	require.True(t, ok, "strategy CreateManager should return *ZotManager")

	strategyBinMgr, ok := zotMgr.binaryManager.(*DefaultBinaryManager)
	require.True(t, ok, "strategy binaryManager should be *DefaultBinaryManager")

	assert.Equal(t, factoryBinMgr.version, strategyBinMgr.version,
		"factory path (%q) and strategy path (%q) must use the same binary version",
		factoryBinMgr.version, strategyBinMgr.version,
	)
}

// =============================================================================
// B2: Default port mismatch — ZotStrategy.GetDefaultPort() returns 5000 but
// DefaultRegistryConfig().Port returns 5001. Both must return 5001.
// =============================================================================

// TestZotStrategy_GetDefaultPort_Is5001 verifies that ZotStrategy.GetDefaultPort
// returns 5001.
// This FAILS today because strategy.go:83 returns 5000.
func TestZotStrategy_GetDefaultPort_Is5001(t *testing.T) {
	s := NewZotStrategy()
	if got := s.GetDefaultPort(); got != 5001 {
		t.Errorf("GetDefaultPort() = %d, want 5001", got)
	}
}

// TestDefaultPort_Consistency verifies that ZotStrategy.GetDefaultPort() and
// DefaultRegistryConfig().Port agree on the same value (5001).
// This FAILS today because ZotStrategy.GetDefaultPort() returns 5000 while
// DefaultRegistryConfig().Port returns 5001.
func TestDefaultPort_Consistency(t *testing.T) {
	s := NewZotStrategy()
	cfg := DefaultRegistryConfig()

	if s.GetDefaultPort() != cfg.Port {
		t.Errorf("ZotStrategy.GetDefaultPort() = %d, but DefaultRegistryConfig().Port = %d; they must be equal",
			s.GetDefaultPort(), cfg.Port,
		)
	}
}

// TestZotStrategy_DefaultPortMatchesTableDriven verifies the default port is 5001
// across all strategy construction methods, table-driven for clarity.
// This FAILS today because GetDefaultPort() currently returns 5000.
func TestZotStrategy_DefaultPortMatchesTableDriven(t *testing.T) {
	tests := []struct {
		name     string
		strategy *ZotStrategy
		wantPort int
	}{
		{
			name:     "newly created strategy returns 5001",
			strategy: NewZotStrategy(),
			wantPort: 5001,
		},
		{
			name:     "zero-value strategy returns 5001",
			strategy: &ZotStrategy{},
			wantPort: 5001,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.strategy.GetDefaultPort()
			assert.Equal(t, tt.wantPort, got,
				"ZotStrategy.GetDefaultPort() must return %d (the canonical Zot port), got %d",
				tt.wantPort, got,
			)
		})
	}
}

// =============================================================================
// B9: Missing nil check in ZotStrategy.CreateManager and AthensStrategy.CreateManager.
// Passing nil must return an error, not panic.
// =============================================================================

// TestZotStrategy_CreateManager_NilRegistry verifies that ZotStrategy.CreateManager
// returns an error when passed a nil registry rather than panicking.
// This FAILS today because strategy.go has no nil guard for the Zot strategy:
// it panics with a nil pointer dereference instead of returning an error.
// assert.NotPanics captures the panic and reports it as a test failure (not a crash).
func TestZotStrategy_CreateManager_NilRegistry(t *testing.T) {
	s := NewZotStrategy()

	// Phase 1: must not panic (currently panics → FAIL)
	var err error
	assert.NotPanics(t, func() {
		_, err = s.CreateManager(nil)
	}, "ZotStrategy.CreateManager(nil) must not panic")

	// Phase 2: when fixed, must return an error (not succeed silently)
	if err == nil {
		t.Fatal("expected error for nil registry, got nil")
	}
}

// TestAthensStrategy_CreateManager_NilRegistry verifies that AthensStrategy.CreateManager
// returns an error when passed a nil registry rather than panicking.
// This FAILS today because strategy.go has no nil guard for the Athens strategy:
// it panics with a nil pointer dereference instead of returning an error.
// assert.NotPanics captures the panic and reports it as a test failure (not a crash).
func TestAthensStrategy_CreateManager_NilRegistry(t *testing.T) {
	s := NewAthensStrategy()

	// Phase 1: must not panic (currently panics → FAIL)
	var err error
	assert.NotPanics(t, func() {
		_, err = s.CreateManager(nil)
	}, "AthensStrategy.CreateManager(nil) must not panic")

	// Phase 2: when fixed, must return an error (not succeed silently)
	if err == nil {
		t.Fatal("expected error for nil registry, got nil")
	}
}

// TestZotStrategy_CreateManager_NilRegistry_ErrorMessage verifies that the error
// returned for nil registry is descriptive.
// This FAILS today because the code panics before it can return any error.
func TestZotStrategy_CreateManager_NilRegistry_ErrorMessage(t *testing.T) {
	s := NewZotStrategy()

	var err error
	// assert.NotPanics prevents a crash; once the nil guard is added the inner
	// assertions verify the error message is meaningful.
	assert.NotPanics(t, func() {
		_, err = s.CreateManager(nil)
	}, "ZotStrategy.CreateManager(nil) must not panic")

	require.Error(t, err, "CreateManager(nil) must return an error")
	assert.Contains(t, err.Error(), "nil",
		"error message should mention nil registry, got: %v", err)
}

// TestAthensStrategy_CreateManager_NilRegistry_ErrorMessage verifies that the error
// returned for nil registry is descriptive.
// This FAILS today because the code panics before it can return any error.
func TestAthensStrategy_CreateManager_NilRegistry_ErrorMessage(t *testing.T) {
	s := NewAthensStrategy()

	var err error
	// assert.NotPanics prevents a crash; once the nil guard is added the inner
	// assertions verify the error message is meaningful.
	assert.NotPanics(t, func() {
		_, err = s.CreateManager(nil)
	}, "AthensStrategy.CreateManager(nil) must not panic")

	require.Error(t, err, "CreateManager(nil) must return an error")
	assert.Contains(t, err.Error(), "nil",
		"error message should mention nil registry, got: %v", err)
}

// TestNilRegistry_AllStrategies_TableDriven is a table-driven test verifying that
// every strategy handles nil gracefully (returns error, not panic).
// Zot and Athens FAIL today; Devpi, Verdaccio, and Squid already pass (they have nil guards).
// assert.NotPanics is used so panics are reported as failures, not binary crashes.
func TestNilRegistry_AllStrategies_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		strategy    RegistryStrategy
		alreadyFail bool // true = currently panics (RED); false = control (GREEN)
	}{
		{
			name:        "ZotStrategy must not panic on nil",
			strategy:    NewZotStrategy(),
			alreadyFail: true,
		},
		{
			name:        "AthensStrategy must not panic on nil",
			strategy:    NewAthensStrategy(),
			alreadyFail: true,
		},
		{
			name:        "DevpiStrategy must not panic on nil (control — already passing)",
			strategy:    NewDevpiStrategy(),
			alreadyFail: false,
		},
		{
			name:        "VerdaccioStrategy must not panic on nil (control — already passing)",
			strategy:    NewVerdaccioStrategy(),
			alreadyFail: false,
		},
		{
			name:        "SquidStrategy must not panic on nil (control — already passing)",
			strategy:    NewSquidStrategy(),
			alreadyFail: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			// assert.NotPanics wraps the call so a panic becomes a test failure
			// rather than crashing the binary — essential for RED-phase tests.
			didNotPanic := assert.NotPanics(t, func() {
				_, err = tt.strategy.CreateManager(nil)
			}, "CreateManager(nil) must not panic")

			if didNotPanic {
				assert.Error(t, err,
					"CreateManager(nil) must return an error, not succeed")
			}
		})
	}
}

// =============================================================================
// Interface compliance
// =============================================================================

// TestZotStrategy_ImplementsRegistryStrategy is a compile-time interface check.
func TestZotStrategy_ImplementsRegistryStrategy(t *testing.T) {
	var _ RegistryStrategy = (*ZotStrategy)(nil)
}

// TestAthensStrategy_ImplementsRegistryStrategy is a compile-time interface check.
func TestAthensStrategy_ImplementsRegistryStrategy(t *testing.T) {
	var _ RegistryStrategy = (*AthensStrategy)(nil)
}

// =============================================================================
// TDD Phase 2 (RED): GetDefaultVersion() — Declarative Registry Version (v0.35.1)
// RC-1: defaultVersions belongs on Strategy layer (GetDefaultVersion on RegistryStrategy)
// RC-4: All 5 strategy implementations must have GetDefaultVersion()
//
// These tests WILL NOT COMPILE until Phase 3 adds:
//   - GetDefaultVersion() string method to all 5 strategy types
//   - GetDefaultVersion() string method to the RegistryStrategy interface
//
// WHY THEY FAIL:
//   - ZotStrategy/AthensStrategy/DevpiStrategy/VerdaccioStrategy/SquidStrategy
//     have no GetDefaultVersion() method → compiler error
//   - RegistryStrategy interface has no GetDefaultVersion() → compiler error
//   - Registry.Version field used in CreateManager tests → compiler error
// =============================================================================

// TestAllStrategies_GetDefaultVersion_TableDriven is a table-driven test that
// calls GetDefaultVersion() on every strategy and asserts the correct version.
//
// Per RC-1 and RC-4:
//   - ZotStrategy    → "2.1.15"  (managed binary — always has a pinned version)
//   - AthensStrategy → ""        (externally managed; no version to pin)
//   - DevpiStrategy  → ""        (externally managed via pipx; no version to pin)
//   - VerdaccioStrategy → ""     (externally managed via npm; no version to pin)
//   - SquidStrategy  → ""        (externally managed via brew; no version to pin)
//
// FAILS TO COMPILE: GetDefaultVersion() method does not exist on any strategy.
func TestAllStrategies_GetDefaultVersion_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		strategy      RegistryStrategy
		wantVersion   string
		managedBinary bool   // true = dvm manages the binary download
		externalTool  string // non-empty = managed by this external tool
	}{
		{
			name:          "ZotStrategy returns 2.1.15 (managed binary)",
			strategy:      NewZotStrategy(),
			wantVersion:   "2.1.15",
			managedBinary: true,
		},
		{
			name:         "AthensStrategy returns empty (externally managed)",
			strategy:     NewAthensStrategy(),
			wantVersion:  "",
			externalTool: "docker/brew",
		},
		{
			name:         "DevpiStrategy returns empty (externally managed via pipx)",
			strategy:     NewDevpiStrategy(),
			wantVersion:  "",
			externalTool: "pipx",
		},
		{
			name:         "VerdaccioStrategy returns empty (externally managed via npm)",
			strategy:     NewVerdaccioStrategy(),
			wantVersion:  "",
			externalTool: "npm",
		},
		{
			name:         "SquidStrategy returns empty (externally managed via brew)",
			strategy:     NewSquidStrategy(),
			wantVersion:  "",
			externalTool: "brew",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// COMPILE ERROR: RegistryStrategy interface has no GetDefaultVersion() method.
			got := tt.strategy.GetDefaultVersion()
			assert.Equal(t, tt.wantVersion, got,
				"GetDefaultVersion() = %q, want %q", got, tt.wantVersion)
		})
	}
}

// TestZotStrategy_GetDefaultVersion verifies the focused case that ZotStrategy
// returns "2.1.15" as its pinned binary version.
//
// This is the canonical version used by dvm to download and install the Zot binary.
// FAILS TO COMPILE: ZotStrategy has no GetDefaultVersion() method.
func TestZotStrategy_GetDefaultVersion(t *testing.T) {
	s := NewZotStrategy()
	// COMPILE ERROR: ZotStrategy has no GetDefaultVersion() method.
	got := s.GetDefaultVersion()
	assert.Equal(t, "2.1.15", got,
		"ZotStrategy.GetDefaultVersion() must return the canonical Zot binary version %q, got %q",
		"2.1.15", got,
	)
}

// TestRegistryStrategy_InterfaceIncludesGetDefaultVersion is a compile-time
// interface compliance check. It verifies that ZotStrategy satisfies an interface
// that includes GetDefaultVersion() — meaning the RegistryStrategy interface
// itself must declare that method.
//
// FAILS TO COMPILE: RegistryStrategy interface has no GetDefaultVersion() method.
func TestRegistryStrategy_InterfaceIncludesGetDefaultVersion(t *testing.T) {
	// This compile-time assertion fails until GetDefaultVersion() is added to
	// the RegistryStrategy interface in service_manager.go.
	// COMPILE ERROR: *ZotStrategy does not implement interface{ GetDefaultVersion() string }
	// because RegistryStrategy (which *ZotStrategy satisfies) doesn't declare GetDefaultVersion().
	var _ interface{ GetDefaultVersion() string } = (*ZotStrategy)(nil)
}

// TestZotStrategy_CreateManager_UsesRegistryVersion verifies that when
// reg.Version is non-empty, ZotStrategy.CreateManager uses that explicit version
// instead of its own default.
//
// This tests the "explicit version wins" rule: a user who pins "2.0.0" in their
// registry YAML gets exactly zot 2.0.0, not the strategy default.
//
// FAILS TO COMPILE:
//   - Registry.Version field doesn't exist (models/registry.go)
func TestZotStrategy_CreateManager_UsesRegistryVersion(t *testing.T) {
	s := NewZotStrategy()
	reg := &models.Registry{
		ID:        1,
		Name:      "test-zot",
		Type:      "zot",
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   t.TempDir(),
		Version:   "2.0.0", // COMPILE ERROR: Registry has no Version field
	}

	mgr, err := s.CreateManager(reg)
	require.NoError(t, err)
	require.NotNil(t, mgr)

	zotMgr, ok := mgr.(*ZotManager)
	require.True(t, ok, "CreateManager should return *ZotManager")

	binMgr, ok := zotMgr.binaryManager.(*DefaultBinaryManager)
	require.True(t, ok, "binaryManager should be *DefaultBinaryManager")

	// When the user pins a version, the manager must use that exact version.
	assert.Equal(t, "2.0.0", binMgr.version,
		"CreateManager must use reg.Version %q when set (not the strategy default %q)",
		"2.0.0", s.GetDefaultVersion()) // COMPILE ERROR: GetDefaultVersion() doesn't exist
}

// TestZotStrategy_CreateManager_FallsBackToDefault verifies that when
// reg.Version is empty, ZotStrategy.CreateManager falls back to
// GetDefaultVersion() ("2.1.15").
//
// This tests the "empty version → use default" rule: when users don't pin a
// version, they always get the strategy's well-tested default.
//
// FAILS TO COMPILE:
//   - Registry.Version field doesn't exist (models/registry.go)
//   - ZotStrategy.GetDefaultVersion() method doesn't exist
func TestZotStrategy_CreateManager_FallsBackToDefault(t *testing.T) {
	s := NewZotStrategy()
	reg := &models.Registry{
		ID:        1,
		Name:      "test-zot",
		Type:      "zot",
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   t.TempDir(),
		// Version intentionally empty — should fall back to GetDefaultVersion()
	}
	// Explicitly confirm Version is empty after construction
	// COMPILE ERROR: Registry has no Version field
	require.Equal(t, "", reg.Version,
		"reg.Version must be empty to test fallback behavior")

	mgr, err := s.CreateManager(reg)
	require.NoError(t, err)
	require.NotNil(t, mgr)

	zotMgr, ok := mgr.(*ZotManager)
	require.True(t, ok, "CreateManager should return *ZotManager")

	binMgr, ok := zotMgr.binaryManager.(*DefaultBinaryManager)
	require.True(t, ok, "binaryManager should be *DefaultBinaryManager")

	// When version is empty, must fall back to GetDefaultVersion() — not hardcode "2.1.15".
	// COMPILE ERROR: GetDefaultVersion() doesn't exist
	wantVersion := s.GetDefaultVersion()
	assert.Equal(t, wantVersion, binMgr.version,
		"CreateManager must use GetDefaultVersion() %q when reg.Version is empty, got %q",
		wantVersion, binMgr.version)
}
