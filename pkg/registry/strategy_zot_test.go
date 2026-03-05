package registry

import (
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// B1: Zot version mismatch — factory.go:22 uses "1.4.3", strategy.go:65 uses "2.0.0"
// Both code paths must use the same version ("2.0.0").
// =============================================================================

// TestZotBinaryVersion_FactoryUsesCorrectVersion verifies that NewZotManager (used
// via factory.go:22) creates a DefaultBinaryManager with version "2.0.0", not "1.4.3".
// This FAILS today because factory.go:22 hardcodes "1.4.3".
func TestZotBinaryVersion_FactoryUsesCorrectVersion(t *testing.T) {
	tests := []struct {
		name        string
		wantVersion string
	}{
		{
			name:        "factory path must use zot 2.0.0",
			wantVersion: "2.0.0",
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
// This FAILS today because the two paths are out of sync (1.4.3 vs 2.0.0).
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
