package registry

import (
	"context"
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

// TestLifecycleManager_RecordActivity tests updating last access time
func TestLifecycleManager_RecordActivity(t *testing.T) {
	mockDB := db.NewMockDataStore()
	mockManager := &MockRegistryManager{}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		ID:          1,
		Name:        "active-zot",
		Type:        "zot",
		Lifecycle:   "on-demand",
		Status:      "running",
		IdleTimeout: 1800,
	}

	// Create registry in DB
	err := mockDB.CreateRegistry(registry)
	require.NoError(t, err)

	// Record activity
	err = lifecycle.RecordActivity(context.Background(), registry.ID)
	assert.NoError(t, err, "Should successfully record activity")

	// Verify last access time was updated (implementation detail)
	// This would check a "last_accessed_at" field in the registry history
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

	// Stop the registry
	err = lifecycle.StopIfIdle(context.Background(), registry)
	assert.NoError(t, err)
	assert.True(t, stopCalled, "Should have called Stop")
}

// TestLifecycleManager_IdleTimeoutReset tests activity resets timeout
func TestLifecycleManager_IdleTimeoutReset(t *testing.T) {
	mockDB := db.NewMockDataStore()
	mockManager := &MockRegistryManager{
		IsRunningFunc: func(ctx context.Context) bool {
			return true
		},
	}

	lifecycle := NewLifecycleManager(mockDB, mockManager)

	registry := &models.Registry{
		ID:          1,
		Name:        "active-zot",
		Type:        "zot",
		Lifecycle:   "on-demand",
		Status:      "running",
		IdleTimeout: 1800, // 30 minutes
	}

	// Create registry
	err := mockDB.CreateRegistry(registry)
	require.NoError(t, err)

	// Record activity (resets timer)
	err = lifecycle.RecordActivity(context.Background(), registry.ID)
	assert.NoError(t, err)

	// Check if should stop - should be false since we just recorded activity
	shouldStop, err := lifecycle.ShouldStop(context.Background(), registry, time.Now())
	assert.NoError(t, err)
	assert.False(t, shouldStop, "Should NOT stop immediately after activity")
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
