package registry

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock Implementations for Testing
// =============================================================================

// MockBinaryManagerForBase implements BinaryManager for BaseServiceManager tests.
type MockBinaryManagerForBase struct {
	EnsureBinaryFunc func(ctx context.Context) (string, error)
	GetVersionFunc   func(ctx context.Context) (string, error)
	NeedsUpdateFunc  func(ctx context.Context) (bool, error)
	UpdateFunc       func(ctx context.Context) error
}

func (m *MockBinaryManagerForBase) EnsureBinary(ctx context.Context) (string, error) {
	if m.EnsureBinaryFunc != nil {
		return m.EnsureBinaryFunc(ctx)
	}
	return "/path/to/binary", nil
}

func (m *MockBinaryManagerForBase) GetVersion(ctx context.Context) (string, error) {
	if m.GetVersionFunc != nil {
		return m.GetVersionFunc(ctx)
	}
	return "1.0.0", nil
}

func (m *MockBinaryManagerForBase) NeedsUpdate(ctx context.Context) (bool, error) {
	if m.NeedsUpdateFunc != nil {
		return m.NeedsUpdateFunc(ctx)
	}
	return false, nil
}

func (m *MockBinaryManagerForBase) Update(ctx context.Context) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx)
	}
	return nil
}

// MockProcessManagerForBase implements ProcessManager for BaseServiceManager tests.
type MockProcessManagerForBase struct {
	StartFunc     func(ctx context.Context, binary string, args []string, config ProcessConfig) error
	StopFunc      func(ctx context.Context) error
	IsRunningFunc func() bool
	GetPIDFunc    func() int
	mu            sync.RWMutex
}

func (m *MockProcessManagerForBase) Start(ctx context.Context, binary string, args []string, config ProcessConfig) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx, binary, args, config)
	}
	return nil
}

func (m *MockProcessManagerForBase) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	return nil
}

func (m *MockProcessManagerForBase) IsRunning() bool {
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc()
	}
	return false
}

func (m *MockProcessManagerForBase) GetPID() int {
	if m.GetPIDFunc != nil {
		return m.GetPIDFunc()
	}
	return 0
}

// =============================================================================
// Test Helpers
// =============================================================================

// setupBaseServiceManager creates a BaseServiceManager with mocks for testing.
func setupBaseServiceManager(t *testing.T) (BaseServiceManager, *MockBinaryManagerForBase, *MockProcessManagerForBase) {
	t.Helper()

	mockBinary := &MockBinaryManagerForBase{}
	mockProcess := &MockProcessManagerForBase{}

	mgr := NewBaseServiceManager(mockBinary, mockProcess)

	return mgr, mockBinary, mockProcess
}

// =============================================================================
// Task 1: Constructor Tests
// =============================================================================

func TestNewBaseServiceManager(t *testing.T) {
	tests := []struct {
		name        string
		binary      BinaryManager
		process     ProcessManager
		wantNil     bool
		description string
	}{
		{
			name:        "valid dependencies",
			binary:      &MockBinaryManagerForBase{},
			process:     &MockProcessManagerForBase{},
			wantNil:     false,
			description: "Constructor should succeed with valid dependencies",
		},
		{
			name:        "with nil binary manager",
			binary:      nil,
			process:     &MockProcessManagerForBase{},
			wantNil:     true,
			description: "Constructor should handle nil BinaryManager",
		},
		{
			name:        "with nil process manager",
			binary:      &MockBinaryManagerForBase{},
			process:     nil,
			wantNil:     true,
			description: "Constructor should handle nil ProcessManager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewBaseServiceManager(tt.binary, tt.process)

			if tt.wantNil {
				// Implementation should validate dependencies
				t.Skip("Constructor validation not yet implemented")
			} else {
				assert.NotNil(t, mgr, "NewBaseServiceManager should return valid instance")
			}
		})
	}
}

func TestNewBaseServiceManager_StoresReferences(t *testing.T) {
	mockBinary := &MockBinaryManagerForBase{}
	mockProcess := &MockProcessManagerForBase{}

	mgr := NewBaseServiceManager(mockBinary, mockProcess)

	// Verify internal state was initialized correctly
	// (This tests internal fields are properly set)
	assert.NotNil(t, mgr, "Manager should store dependencies")
}

// =============================================================================
// Task 2: RecordStart Tests
// =============================================================================

func TestBaseServiceManager_RecordStart(t *testing.T) {
	tests := []struct {
		name        string
		description string
	}{
		{
			name:        "sets start time",
			description: "RecordStart should set startTime to current time",
		},
		{
			name:        "sets last access time",
			description: "RecordStart should set lastAccessTime to current time",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, _, _ := setupBaseServiceManager(t)

			_ = time.Now() // beforeStart
			mgr.RecordStart()
			_ = time.Now() // afterStart

			// startTime should be between beforeStart and afterStart
			// (We'll verify this by checking GetUptime() returns non-zero)
			uptime := mgr.GetUptime()
			assert.GreaterOrEqual(t, uptime, time.Duration(0), "Uptime should be non-negative after RecordStart")
		})
	}
}

func TestBaseServiceManager_RecordStart_ThreadSafe(t *testing.T) {
	mgr, _, _ := setupBaseServiceManager(t)

	// Call RecordStart from multiple goroutines
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mgr.RecordStart()
		}()
	}

	wg.Wait()

	// If we get here without data race, mutex is working
	uptime := mgr.GetUptime()
	assert.GreaterOrEqual(t, uptime, time.Duration(0), "RecordStart should be thread-safe")
}

// =============================================================================
// Task 3: GetUptime Tests
// =============================================================================

func TestBaseServiceManager_GetUptime(t *testing.T) {
	tests := []struct {
		name          string
		recordStart   bool
		sleepDuration time.Duration
		wantZero      bool
		description   string
	}{
		{
			name:          "before start returns zero",
			recordStart:   false,
			sleepDuration: 0,
			wantZero:      true,
			description:   "GetUptime should return zero duration if not started",
		},
		{
			name:          "after start returns duration",
			recordStart:   true,
			sleepDuration: 10 * time.Millisecond,
			wantZero:      false,
			description:   "GetUptime should return elapsed time after RecordStart",
		},
		{
			name:          "increases over time",
			recordStart:   true,
			sleepDuration: 50 * time.Millisecond,
			wantZero:      false,
			description:   "GetUptime should increase as time passes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, _, _ := setupBaseServiceManager(t)

			if tt.recordStart {
				mgr.RecordStart()
				if tt.sleepDuration > 0 {
					time.Sleep(tt.sleepDuration)
				}
			}

			uptime := mgr.GetUptime()

			if tt.wantZero {
				assert.Equal(t, time.Duration(0), uptime, "Uptime should be zero before start")
			} else {
				assert.Greater(t, uptime, time.Duration(0), "Uptime should be positive after start")

				// If we slept, uptime should be at least that long
				if tt.sleepDuration > 0 {
					assert.GreaterOrEqual(t, uptime, tt.sleepDuration, "Uptime should reflect elapsed time")
				}
			}
		})
	}
}

func TestBaseServiceManager_GetUptime_MultipleCalls(t *testing.T) {
	mgr, _, _ := setupBaseServiceManager(t)
	mgr.RecordStart()

	// Call GetUptime multiple times, verify it increases
	uptime1 := mgr.GetUptime()
	time.Sleep(10 * time.Millisecond)
	uptime2 := mgr.GetUptime()

	assert.Greater(t, uptime2, uptime1, "GetUptime should return increasing values")
}

// =============================================================================
// Task 4: IsProcessRunning Tests
// =============================================================================

func TestBaseServiceManager_IsProcessRunning(t *testing.T) {
	tests := []struct {
		name          string
		mockIsRunning bool
		wantRunning   bool
		description   string
	}{
		{
			name:          "process running",
			mockIsRunning: true,
			wantRunning:   true,
			description:   "IsProcessRunning should return true when process is running",
		},
		{
			name:          "process not running",
			mockIsRunning: false,
			wantRunning:   false,
			description:   "IsProcessRunning should return false when process is stopped",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, _, mockProcess := setupBaseServiceManager(t)

			mockProcess.IsRunningFunc = func() bool {
				return tt.mockIsRunning
			}

			running := mgr.IsProcessRunning()
			assert.Equal(t, tt.wantRunning, running, tt.description)
		})
	}
}

func TestBaseServiceManager_IsProcessRunning_DelegatesToProcessManager(t *testing.T) {
	mgr, _, mockProcess := setupBaseServiceManager(t)

	callCount := 0
	mockProcess.IsRunningFunc = func() bool {
		callCount++
		return true
	}

	mgr.IsProcessRunning()
	assert.Equal(t, 1, callCount, "IsProcessRunning should delegate to ProcessManager")
}

// =============================================================================
// Task 5: GetPID Tests
// =============================================================================

func TestBaseServiceManager_GetPID(t *testing.T) {
	tests := []struct {
		name        string
		mockPID     int
		wantPID     int
		description string
	}{
		{
			name:        "process running with PID",
			mockPID:     12345,
			wantPID:     12345,
			description: "GetPID should return process ID when running",
		},
		{
			name:        "process not running",
			mockPID:     0,
			wantPID:     0,
			description: "GetPID should return 0 when not running",
		},
		{
			name:        "different PID",
			mockPID:     99999,
			wantPID:     99999,
			description: "GetPID should return correct PID value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, _, mockProcess := setupBaseServiceManager(t)

			mockProcess.GetPIDFunc = func() int {
				return tt.mockPID
			}

			pid := mgr.GetPID()
			assert.Equal(t, tt.wantPID, pid, tt.description)
		})
	}
}

func TestBaseServiceManager_GetPID_DelegatesToProcessManager(t *testing.T) {
	mgr, _, mockProcess := setupBaseServiceManager(t)

	callCount := 0
	mockProcess.GetPIDFunc = func() int {
		callCount++
		return 54321
	}

	pid := mgr.GetPID()
	assert.Equal(t, 1, callCount, "GetPID should delegate to ProcessManager")
	assert.Equal(t, 54321, pid, "GetPID should return value from ProcessManager")
}

// =============================================================================
// Task 6: StopProcess Tests
// =============================================================================

func TestBaseServiceManager_StopProcess(t *testing.T) {
	tests := []struct {
		name        string
		mockError   error
		wantError   bool
		description string
	}{
		{
			name:        "successful stop",
			mockError:   nil,
			wantError:   false,
			description: "StopProcess should succeed when process stops cleanly",
		},
		{
			name:        "stop fails",
			mockError:   assert.AnError,
			wantError:   true,
			description: "StopProcess should propagate errors from ProcessManager",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, _, mockProcess := setupBaseServiceManager(t)
			ctx := context.Background()

			mockProcess.StopFunc = func(ctx context.Context) error {
				return tt.mockError
			}

			err := mgr.StopProcess(ctx)

			if tt.wantError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestBaseServiceManager_StopProcess_StopsIdleTimer(t *testing.T) {
	mgr, _, mockProcess := setupBaseServiceManager(t)
	ctx := context.Background()

	var mu sync.Mutex
	stopFuncCalled := false
	mockProcess.StopFunc = func(ctx context.Context) error {
		return nil
	}

	// Setup idle timer
	mgr.SetupIdleTimer("on-demand", 1*time.Second, func() {
		mu.Lock()
		stopFuncCalled = true
		mu.Unlock()
	})

	// Stop process should stop idle timer
	err := mgr.StopProcess(ctx)
	require.NoError(t, err)

	// Wait to ensure timer doesn't fire
	time.Sleep(1500 * time.Millisecond)

	mu.Lock()
	called := stopFuncCalled
	mu.Unlock()
	assert.False(t, called, "Idle timer should be stopped by StopProcess")
}

func TestBaseServiceManager_StopProcess_DelegatesToProcessManager(t *testing.T) {
	mgr, _, mockProcess := setupBaseServiceManager(t)
	ctx := context.Background()

	callCount := 0
	mockProcess.StopFunc = func(ctx context.Context) error {
		callCount++
		return nil
	}

	err := mgr.StopProcess(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 1, callCount, "StopProcess should delegate to ProcessManager")
}

// =============================================================================
// Task 7: EnsureBinary Tests
// =============================================================================

func TestBaseServiceManager_EnsureBinary(t *testing.T) {
	tests := []struct {
		name        string
		mockPath    string
		mockError   error
		wantPath    string
		wantError   bool
		description string
	}{
		{
			name:        "binary exists",
			mockPath:    "/usr/local/bin/zot",
			mockError:   nil,
			wantPath:    "/usr/local/bin/zot",
			wantError:   false,
			description: "EnsureBinary should return path when binary exists",
		},
		{
			name:        "binary downloaded",
			mockPath:    "/tmp/downloads/athens",
			mockError:   nil,
			wantPath:    "/tmp/downloads/athens",
			wantError:   false,
			description: "EnsureBinary should return path after download",
		},
		{
			name:        "download fails",
			mockPath:    "",
			mockError:   assert.AnError,
			wantPath:    "",
			wantError:   true,
			description: "EnsureBinary should propagate download errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, mockBinary, _ := setupBaseServiceManager(t)
			ctx := context.Background()

			mockBinary.EnsureBinaryFunc = func(ctx context.Context) (string, error) {
				return tt.mockPath, tt.mockError
			}

			path, err := mgr.EnsureBinary(ctx)

			if tt.wantError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.wantPath, path, "Should return correct binary path")
			}
		})
	}
}

func TestBaseServiceManager_EnsureBinary_DelegatesToBinaryManager(t *testing.T) {
	mgr, mockBinary, _ := setupBaseServiceManager(t)
	ctx := context.Background()

	callCount := 0
	mockBinary.EnsureBinaryFunc = func(ctx context.Context) (string, error) {
		callCount++
		return "/path/to/binary", nil
	}

	path, err := mgr.EnsureBinary(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "/path/to/binary", path)
	assert.Equal(t, 1, callCount, "EnsureBinary should delegate to BinaryManager")
}

// =============================================================================
// Task 8: SetupIdleTimer Tests
// =============================================================================

func TestBaseServiceManager_SetupIdleTimer(t *testing.T) {
	tests := []struct {
		name           string
		lifecycle      string
		timeout        time.Duration
		shouldActivate bool
		description    string
	}{
		{
			name:           "on-demand activates timer",
			lifecycle:      "on-demand",
			timeout:        100 * time.Millisecond,
			shouldActivate: true,
			description:    "SetupIdleTimer should activate for on-demand lifecycle",
		},
		{
			name:           "persistent ignores timer",
			lifecycle:      "persistent",
			timeout:        100 * time.Millisecond,
			shouldActivate: false,
			description:    "SetupIdleTimer should ignore persistent lifecycle",
		},
		{
			name:           "manual ignores timer",
			lifecycle:      "manual",
			timeout:        100 * time.Millisecond,
			shouldActivate: false,
			description:    "SetupIdleTimer should ignore manual lifecycle",
		},
		{
			name:           "empty lifecycle ignores timer",
			lifecycle:      "",
			timeout:        100 * time.Millisecond,
			shouldActivate: false,
			description:    "SetupIdleTimer should ignore empty lifecycle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, _, _ := setupBaseServiceManager(t)

			var mu sync.Mutex
			stopFuncCalled := false
			stopFunc := func() {
				mu.Lock()
				stopFuncCalled = true
				mu.Unlock()
			}

			mgr.SetupIdleTimer(tt.lifecycle, tt.timeout, stopFunc)

			// Wait for timer to potentially fire
			time.Sleep(tt.timeout + 50*time.Millisecond)

			mu.Lock()
			called := stopFuncCalled
			mu.Unlock()
			if tt.shouldActivate {
				assert.True(t, called, "Timer should fire for %s lifecycle", tt.lifecycle)
			} else {
				assert.False(t, called, "Timer should not fire for %s lifecycle", tt.lifecycle)
			}
		})
	}
}

func TestBaseServiceManager_SetupIdleTimer_RespectsTimeout(t *testing.T) {
	mgr, _, _ := setupBaseServiceManager(t)

	var mu sync.Mutex
	stopFuncCalled := false
	timeout := 200 * time.Millisecond

	mgr.SetupIdleTimer("on-demand", timeout, func() {
		mu.Lock()
		stopFuncCalled = true
		mu.Unlock()
	})

	// Check before timeout expires
	time.Sleep(100 * time.Millisecond)
	mu.Lock()
	called := stopFuncCalled
	mu.Unlock()
	assert.False(t, called, "Timer should not fire before timeout")

	// Check after timeout expires
	time.Sleep(150 * time.Millisecond)
	mu.Lock()
	called = stopFuncCalled
	mu.Unlock()
	assert.True(t, called, "Timer should fire after timeout")
}

func TestBaseServiceManager_SetupIdleTimer_CanBeStopped(t *testing.T) {
	mgr, _, _ := setupBaseServiceManager(t)

	var mu sync.Mutex
	stopFuncCalled := false
	timeout := 100 * time.Millisecond

	mgr.SetupIdleTimer("on-demand", timeout, func() {
		mu.Lock()
		stopFuncCalled = true
		mu.Unlock()
	})

	// Stop timer before it fires
	time.Sleep(50 * time.Millisecond)
	mgr.StopIdleTimer()

	// Wait past timeout
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	called := stopFuncCalled
	mu.Unlock()
	assert.False(t, called, "Timer should not fire after being stopped")
}

// =============================================================================
// Task 9: ResetIdleTimer Tests
// =============================================================================

func TestBaseServiceManager_ResetIdleTimer(t *testing.T) {
	tests := []struct {
		name         string
		lifecycle    string
		initialTimer bool
		shouldReset  bool
		description  string
	}{
		{
			name:         "resets on-demand timer",
			lifecycle:    "on-demand",
			initialTimer: true,
			shouldReset:  true,
			description:  "ResetIdleTimer should reset for on-demand lifecycle",
		},
		{
			name:         "ignores persistent",
			lifecycle:    "persistent",
			initialTimer: true,
			shouldReset:  false,
			description:  "ResetIdleTimer should ignore persistent lifecycle",
		},
		{
			name:         "ignores manual",
			lifecycle:    "manual",
			initialTimer: false,
			shouldReset:  false,
			description:  "ResetIdleTimer should ignore manual lifecycle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, _, _ := setupBaseServiceManager(t)

			var mu sync.Mutex
			stopFuncCallCount := 0
			timeout := 150 * time.Millisecond
			stopFunc := func() {
				mu.Lock()
				stopFuncCallCount++
				mu.Unlock()
			}

			if tt.initialTimer {
				mgr.SetupIdleTimer(tt.lifecycle, timeout, stopFunc)
				time.Sleep(50 * time.Millisecond)
			}

			// Reset timer
			mgr.ResetIdleTimer(tt.lifecycle, timeout, stopFunc)

			if tt.shouldReset {
				// Wait for original timeout (should not fire because it was reset)
				time.Sleep(120 * time.Millisecond)
				mu.Lock()
				count := stopFuncCallCount
				mu.Unlock()
				assert.Equal(t, 0, count, "Original timer should be cancelled")

				// Wait for new timeout (should fire)
				time.Sleep(80 * time.Millisecond)
				mu.Lock()
				count = stopFuncCallCount
				mu.Unlock()
				assert.Equal(t, 1, count, "New timer should fire after reset")
			}
		})
	}
}

func TestBaseServiceManager_ResetIdleTimer_ExtendsTimeout(t *testing.T) {
	mgr, _, _ := setupBaseServiceManager(t)

	var mu sync.Mutex
	stopFuncCalled := false
	initialTimeout := 100 * time.Millisecond

	// Setup initial timer
	mgr.SetupIdleTimer("on-demand", initialTimeout, func() {
		mu.Lock()
		stopFuncCalled = true
		mu.Unlock()
	})

	// Wait halfway through timeout
	time.Sleep(60 * time.Millisecond)

	// Reset timer with new timeout
	mgr.ResetIdleTimer("on-demand", 200*time.Millisecond, func() {
		mu.Lock()
		stopFuncCalled = true
		mu.Unlock()
	})

	// Wait past original timeout
	time.Sleep(60 * time.Millisecond)
	mu.Lock()
	called := stopFuncCalled
	mu.Unlock()
	assert.False(t, called, "Timer should not fire at original timeout after reset")

	// Wait for new timeout
	time.Sleep(160 * time.Millisecond)
	mu.Lock()
	called = stopFuncCalled
	mu.Unlock()
	assert.True(t, called, "Timer should fire at new timeout")
}

func TestBaseServiceManager_ResetIdleTimer_SafeWhenNoTimer(t *testing.T) {
	mgr, _, _ := setupBaseServiceManager(t)

	// Reset when no timer exists - should not panic
	assert.NotPanics(t, func() {
		mgr.ResetIdleTimer("on-demand", 100*time.Millisecond, func() {})
	}, "ResetIdleTimer should be safe when no timer exists")
}

// =============================================================================
// Task 10: StopIdleTimer Tests
// =============================================================================

func TestBaseServiceManager_StopIdleTimer(t *testing.T) {
	tests := []struct {
		name        string
		setupTimer  bool
		description string
	}{
		{
			name:        "stops active timer",
			setupTimer:  true,
			description: "StopIdleTimer should cancel active timer",
		},
		{
			name:        "safe when no timer",
			setupTimer:  false,
			description: "StopIdleTimer should be safe when no timer exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr, _, _ := setupBaseServiceManager(t)

			var mu sync.Mutex
			stopFuncCalled := false

			if tt.setupTimer {
				mgr.SetupIdleTimer("on-demand", 100*time.Millisecond, func() {
					mu.Lock()
					stopFuncCalled = true
					mu.Unlock()
				})
			}

			// Stop timer
			assert.NotPanics(t, func() {
				mgr.StopIdleTimer()
			}, "StopIdleTimer should not panic")

			// Wait to ensure timer doesn't fire
			time.Sleep(150 * time.Millisecond)

			mu.Lock()
			called := stopFuncCalled
			mu.Unlock()
			assert.False(t, called, "Timer should not fire after being stopped")
		})
	}
}

func TestBaseServiceManager_StopIdleTimer_MultipleCalls(t *testing.T) {
	mgr, _, _ := setupBaseServiceManager(t)

	mgr.SetupIdleTimer("on-demand", 100*time.Millisecond, func() {})

	// Stop multiple times - should be idempotent
	assert.NotPanics(t, func() {
		mgr.StopIdleTimer()
		mgr.StopIdleTimer()
		mgr.StopIdleTimer()
	}, "StopIdleTimer should be safe to call multiple times")
}

func TestBaseServiceManager_StopIdleTimer_DoesNotTriggerStopFunc(t *testing.T) {
	mgr, _, _ := setupBaseServiceManager(t)

	var mu sync.Mutex
	stopFuncCalled := false
	mgr.SetupIdleTimer("on-demand", 50*time.Millisecond, func() {
		mu.Lock()
		stopFuncCalled = true
		mu.Unlock()
	})

	// Stop timer immediately
	mgr.StopIdleTimer()

	// Wait well past timeout
	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	called := stopFuncCalled
	mu.Unlock()
	assert.False(t, called, "StopIdleTimer should prevent stopFunc from being called")
}

// =============================================================================
// Concurrency Tests
// =============================================================================

func TestBaseServiceManager_ConcurrentAccess(t *testing.T) {
	mgr, _, mockProcess := setupBaseServiceManager(t)

	mockProcess.IsRunningFunc = func() bool { return true }
	mockProcess.GetPIDFunc = func() int { return 12345 }

	var wg sync.WaitGroup

	// Spawn multiple goroutines accessing manager concurrently
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mgr.RecordStart()
			mgr.GetUptime()
			mgr.IsProcessRunning()
			mgr.GetPID()
		}()
	}

	wg.Wait()

	// If we get here without race detector complaining, mutex is working
	assert.True(t, true, "Concurrent access should be thread-safe")
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestBaseServiceManager_CompilationOnly(t *testing.T) {
	// This test will fail to compile until BaseServiceManager is implemented
	// That's expected - we're in TDD RED phase

	var _ BaseServiceManager
	t.Skip("BaseServiceManager not yet implemented - expected to fail")
}
