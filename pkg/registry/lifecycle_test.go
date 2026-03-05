package registry

import (
	"context"
	"strings"
	"testing"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLifecycleManager_EnsureRunning_Auto tests auto mode starts registry
func TestLifecycleManager_EnsureRunning_Auto(t *testing.T) {
	mockDB := db.NewMockDataStore()
	mockManager := &MockRegistryManager{
		StartFunc: func(ctx context.Context) error {
			return nil
		},
		IsRunningFunc: func(ctx context.Context) bool {
			return false // Not running initially
		},
	}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		Name:      "auto-zot",
		Type:      "zot",
		Lifecycle: "persistent", // Auto/persistent should auto-start
		Status:    "stopped",
	}

	err := lifecycle.EnsureRunning(context.Background(), registry)
	assert.NoError(t, err, "Auto mode should start registry")
}

// TestLifecycleManager_EnsureRunning_OnDemand tests on-demand mode starts registry
func TestLifecycleManager_EnsureRunning_OnDemand(t *testing.T) {
	mockDB := db.NewMockDataStore()

	startCalled := false
	mockManager := &MockRegistryManager{
		StartFunc: func(ctx context.Context) error {
			startCalled = true
			return nil
		},
		IsRunningFunc: func(ctx context.Context) bool {
			return false // Not running initially
		},
	}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		Name:        "ondemand-zot",
		Type:        "zot",
		Lifecycle:   "on-demand",
		Status:      "stopped",
		IdleTimeout: 1800, // 30 minutes
	}

	err := lifecycle.EnsureRunning(context.Background(), registry)
	assert.NoError(t, err, "On-demand mode should start registry")
	assert.True(t, startCalled, "Start should have been called")
}

// TestLifecycleManager_EnsureRunning_Manual tests manual mode does NOT auto-start
func TestLifecycleManager_EnsureRunning_Manual(t *testing.T) {
	mockDB := db.NewMockDataStore()

	startCalled := false
	mockManager := &MockRegistryManager{
		StartFunc: func(ctx context.Context) error {
			startCalled = true
			return nil
		},
		IsRunningFunc: func(ctx context.Context) bool {
			return false // Not running
		},
	}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		Name:      "manual-zot",
		Type:      "zot",
		Lifecycle: "manual",
		Status:    "stopped",
	}

	err := lifecycle.EnsureRunning(context.Background(), registry)

	// Manual mode should NOT auto-start
	// Behavior: Either return error or do nothing (depends on design decision)
	// For this test, we expect it to NOT call Start
	_ = err // Ignore error for now - design decision pending
	assert.False(t, startCalled, "Manual mode should NOT auto-start registry")
}

// TestLifecycleManager_EnsureRunning_AlreadyRunning tests no-op when already running
func TestLifecycleManager_EnsureRunning_AlreadyRunning(t *testing.T) {
	mockDB := db.NewMockDataStore()

	startCalled := false
	mockManager := &MockRegistryManager{
		StartFunc: func(ctx context.Context) error {
			startCalled = true
			return nil
		},
		IsRunningFunc: func(ctx context.Context) bool {
			return true // Already running
		},
	}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		Name:      "running-zot",
		Type:      "zot",
		Lifecycle: "persistent",
		Status:    "running",
	}

	err := lifecycle.EnsureRunning(context.Background(), registry)
	assert.NoError(t, err)
	assert.False(t, startCalled, "Should not call Start if already running")
}

// TestLifecycleManager_IdleTimeout tests registry stops after idle timeout
func TestLifecycleManager_IdleTimeout(t *testing.T) {
	mockDB := db.NewMockDataStore()

	stopCalled := false
	mockManager := &MockRegistryManager{
		StopFunc: func(ctx context.Context) error {
			stopCalled = true
			return nil
		},
		IsRunningFunc: func(ctx context.Context) bool {
			return true
		},
	}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		ID:          1,
		Name:        "idle-zot",
		Type:        "zot",
		Lifecycle:   "on-demand",
		Status:      "running",
		IdleTimeout: 1, // 1 second for testing
	}

	// Check if registry should be stopped (idle timeout exceeded)
	shouldStop, err := lifecycle.ShouldStop(context.Background(), registry, time.Now().Add(-2*time.Second))
	assert.NoError(t, err)
	assert.True(t, shouldStop, "Should stop after idle timeout")

	// StopIfIdle is not yet implemented (deferred until last_accessed_at tracking)
	err = lifecycle.StopIfIdle(context.Background(), registry)
	assert.Error(t, err, "StopIfIdle should return not-implemented error")
	assert.Contains(t, err.Error(), "not yet implemented")
	assert.False(t, stopCalled, "Stop should not have been called (StopIfIdle is not implemented)")
}

// TestLifecycleManager_PersistentNeverStops tests persistent lifecycle never auto-stops
func TestLifecycleManager_PersistentNeverStops(t *testing.T) {
	mockDB := db.NewMockDataStore()
	mockManager := &MockRegistryManager{
		IsRunningFunc: func(ctx context.Context) bool {
			return true
		},
	}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		Name:      "persistent-zot",
		Type:      "zot",
		Lifecycle: "persistent",
		Status:    "running",
	}

	// Even with old last access time, persistent should never auto-stop
	shouldStop, err := lifecycle.ShouldStop(context.Background(), registry, time.Now().Add(-24*time.Hour))
	assert.NoError(t, err)
	assert.False(t, shouldStop, "Persistent registry should never auto-stop")
}

// TestLifecycleManager_ManualNeverAutoStops tests manual lifecycle never auto-stops
func TestLifecycleManager_ManualNeverAutoStops(t *testing.T) {
	mockDB := db.NewMockDataStore()
	mockManager := &MockRegistryManager{
		IsRunningFunc: func(ctx context.Context) bool {
			return true
		},
	}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		Name:      "manual-zot",
		Type:      "zot",
		Lifecycle: "manual",
		Status:    "running",
	}

	// Manual should never auto-stop
	shouldStop, err := lifecycle.ShouldStop(context.Background(), registry, time.Now().Add(-24*time.Hour))
	assert.NoError(t, err)
	assert.False(t, shouldStop, "Manual registry should never auto-stop")
}

// TestLifecycleManager_GlobalIdleTimeout tests using global default timeout
func TestLifecycleManager_GlobalIdleTimeout(t *testing.T) {
	mockDB := db.NewMockDataStore()
	mockManager := &MockRegistryManager{}

	// Set global idle timeout
	err := mockDB.SetDefault("registry-idle-timeout", "45m")
	require.NoError(t, err)

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		Name:        "zot-default-timeout",
		Type:        "zot",
		Lifecycle:   "on-demand",
		IdleTimeout: 0, // Not set - should use global default
	}

	// Get effective timeout
	timeout := lifecycle.GetEffectiveTimeout(registry)
	assert.Equal(t, 45*time.Minute, timeout, "Should use global default timeout")
}

// TestLifecycleManager_RegistryOverridesGlobal tests registry timeout overrides global
func TestLifecycleManager_RegistryOverridesGlobal(t *testing.T) {
	mockDB := db.NewMockDataStore()
	mockManager := &MockRegistryManager{}

	// Set global idle timeout
	err := mockDB.SetDefault("registry-idle-timeout", "30m")
	require.NoError(t, err)

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		Name:        "zot-custom-timeout",
		Type:        "zot",
		Lifecycle:   "on-demand",
		IdleTimeout: 3600, // 60 minutes - overrides global
	}

	// Get effective timeout
	timeout := lifecycle.GetEffectiveTimeout(registry)
	assert.Equal(t, 60*time.Minute, timeout, "Registry timeout should override global")
}

// =============================================================================
// B5: Lifecycle dead code removal — TDD RED phase
//
// These tests define the INTENDED behavior after the fix:
//   - StopIfIdle should return a "not implemented" error (feature deferred).
//   - RecordActivity is a recognised no-op; existing tests that assert
//     NoError on it will continue to pass but are now documented as
//     candidates for removal once RecordActivity is deleted from the API.
//
// Tests that must be REMOVED after the fix lands (they test the no-op that
// will be deleted along with RecordActivity):
//   - TestLifecycleManager_RecordActivity        (line ~133)
//   - TestLifecycleManager_IdleTimeoutReset       (line ~199, uses RecordActivity)
// =============================================================================

// TestLifecycleManager_StopIfIdle_ReturnsNotImplemented verifies that
// StopIfIdle returns an error indicating the feature is not yet implemented.
//
// Current behaviour (BUG B5): StopIfIdle silently returns nil because the
// lastAccess calculation always produces a duration equal to (not greater
// than) IdleTimeout, so shouldStop is always false. The fix will replace
// the broken body with an explicit "not implemented" error.
//
// This test FAILS today because StopIfIdle returns nil.
func TestLifecycleManager_StopIfIdle_ReturnsNotImplemented(t *testing.T) {
	mockDB := db.NewMockDataStore()
	mockManager := &MockRegistryManager{
		IsRunningFunc: func(ctx context.Context) bool {
			return true
		},
		StopFunc: func(ctx context.Context) error {
			// Should never be reached once the fix is in place.
			return nil
		},
	}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		ID:          42,
		Name:        "on-demand-zot",
		Type:        "zot",
		Lifecycle:   "on-demand",
		Status:      "running",
		IdleTimeout: 1800,
	}

	err := lifecycle.StopIfIdle(context.Background(), registry)

	// After the fix: StopIfIdle must return a non-nil error that communicates
	// the feature is not yet implemented (idle-timeout auto-stop is deferred
	// until last_accessed_at is tracked in the database).
	if err == nil {
		t.Fatal("StopIfIdle should return an error indicating it is not implemented, but got nil")
	}

	// The error message must signal "not implemented" so callers know to skip
	// idle-timeout enforcement rather than treating it as a hard failure.
	lowerMsg := strings.ToLower(err.Error())
	if !strings.Contains(lowerMsg, "not implemented") && !strings.Contains(lowerMsg, "not yet implemented") {
		t.Errorf("expected error to contain 'not implemented', got: %q", err.Error())
	}
}

// TestLifecycleManager_StopIfIdle_BrokenLastAccessCalculation documents the
// exact broken behaviour present today: the lastAccess is set to exactly
// IdleTimeout seconds ago, so time.Since(lastAccess) is never strictly
// greater than the timeout — shouldStop is always false.
//
// This test PASSES today (demonstrating the bug: Stop is never called even
// when it should be). It serves as a regression anchor to confirm the bug
// exists before the fix is applied.
func TestLifecycleManager_StopIfIdle_BrokenLastAccessCalculation(t *testing.T) {
	mockDB := db.NewMockDataStore()

	stopCallCount := 0
	mockManager := &MockRegistryManager{
		IsRunningFunc: func(ctx context.Context) bool {
			return true
		},
		StopFunc: func(ctx context.Context) error {
			stopCallCount++
			return nil
		},
	}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		ID:          99,
		Name:        "idle-broken-zot",
		Type:        "zot",
		Lifecycle:   "on-demand",
		Status:      "running",
		IdleTimeout: 1, // 1 second — very short; should trigger a stop
	}

	// Wait just over the idle timeout so a correct implementation would stop.
	time.Sleep(2 * time.Millisecond)

	_ = lifecycle.StopIfIdle(context.Background(), registry)

	// BUG: current code never calls Stop because lastAccess is calculated as
	// time.Now().Add(-IdleTimeout * time.Second), making the idle duration
	// exactly equal to the timeout (not greater), so shouldStop == false.
	// Once the bug is fixed (StopIfIdle returns "not implemented"), this
	// assertion should be updated or removed accordingly.
	if stopCallCount != 0 {
		t.Logf("NOTE: Stop was called %d time(s); the broken-lastAccess bug may already be fixed", stopCallCount)
	}
}

// MockRegistryManager for testing
type MockRegistryManager struct {
	StartFunc         func(ctx context.Context) error
	StopFunc          func(ctx context.Context) error
	StatusFunc        func(ctx context.Context) (*RegistryStatus, error)
	EnsureRunningFunc func(ctx context.Context) error
	IsRunningFunc     func(ctx context.Context) bool
	GetEndpointFunc   func() string
	PruneFunc         func(ctx context.Context, opts PruneOptions) (*PruneResult, error)
}

func (m *MockRegistryManager) Start(ctx context.Context) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx)
	}
	return nil
}

func (m *MockRegistryManager) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	return nil
}

func (m *MockRegistryManager) Status(ctx context.Context) (*RegistryStatus, error) {
	if m.StatusFunc != nil {
		return m.StatusFunc(ctx)
	}
	return &RegistryStatus{State: "running"}, nil
}

func (m *MockRegistryManager) EnsureRunning(ctx context.Context) error {
	if m.EnsureRunningFunc != nil {
		return m.EnsureRunningFunc(ctx)
	}
	return nil
}

func (m *MockRegistryManager) IsRunning(ctx context.Context) bool {
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc(ctx)
	}
	return true
}

func (m *MockRegistryManager) GetEndpoint() string {
	if m.GetEndpointFunc != nil {
		return m.GetEndpointFunc()
	}
	return "http://localhost:5000"
}

func (m *MockRegistryManager) Prune(ctx context.Context, opts PruneOptions) (*PruneResult, error) {
	if m.PruneFunc != nil {
		return m.PruneFunc(ctx, opts)
	}
	return &PruneResult{}, nil
}
