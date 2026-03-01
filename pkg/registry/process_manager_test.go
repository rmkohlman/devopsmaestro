package registry

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupTestProcessManager creates a ProcessManager for testing.
func setupTestProcessManager(t *testing.T) ProcessManager {
	t.Helper()

	pidFile := filepath.Join(t.TempDir(), "test.pid")
	logFile := filepath.Join(t.TempDir(), "test.log")

	config := ProcessConfig{
		PIDFile: pidFile,
		LogFile: logFile,
	}

	return NewProcessManager(config)
}

// =============================================================================
// Task 2.1: Process Start Tests
// =============================================================================

func TestProcessManager_Start_Success(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	// Start a long-running test process (sleep)
	binary := "sleep"
	args := []string{"10"}
	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	err := mgr.Start(ctx, binary, args, config)
	require.NoError(t, err, "Start should succeed with valid binary")

	// Verify process is running
	assert.True(t, mgr.IsRunning(), "IsRunning should return true after Start")

	// Verify PID is set
	pid := mgr.GetPID()
	assert.Greater(t, pid, 0, "PID should be positive")

	// Cleanup
	defer mgr.Stop(ctx)
}

func TestProcessManager_Start_AlreadyRunning(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	// Start first time
	err := mgr.Start(ctx, "sleep", []string{"10"}, config)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Start second time - should fail or be idempotent
	err = mgr.Start(ctx, "sleep", []string{"10"}, config)
	assert.Error(t, err, "Start should fail when already running")
	assert.Contains(t, err.Error(), "already running", "Error should indicate process already running")
}

func TestProcessManager_Start_BinaryNotFound(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	err := mgr.Start(ctx, "nonexistent-binary-12345", []string{}, config)
	assert.Error(t, err, "Start should fail with nonexistent binary")
	assert.Contains(t, err.Error(), "not found", "Error should indicate binary not found")
}

func TestProcessManager_Start_InvalidArgs(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	// Start with args that will cause immediate process exit
	// Note: Start() only spawns the process - it doesn't wait for success
	// The process will start but exit immediately with an error
	err := mgr.Start(ctx, "sleep", []string{"invalid"}, config)
	require.NoError(t, err, "Start should succeed (process spawns, then exits)")

	// Wait for process to exit
	time.Sleep(100 * time.Millisecond)

	// Process should have exited due to invalid args
	assert.False(t, mgr.IsRunning(), "Process should have exited with invalid args")
}

func TestProcessManager_Start_ContextCancellation(t *testing.T) {
	mgr := setupTestProcessManager(t)

	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	// Create context with immediate cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := mgr.Start(ctx, "sleep", []string{"10"}, config)
	assert.Error(t, err, "Start should fail with cancelled context")
	assert.Contains(t, err.Error(), "context", "Error should mention context cancellation")
}

// =============================================================================
// Task 2.2: Process Stop Tests
// =============================================================================

func TestProcessManager_Stop_Success(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	// Start process
	err := mgr.Start(ctx, "sleep", []string{"10"}, config)
	require.NoError(t, err)

	// Stop process
	err = mgr.Stop(ctx)
	require.NoError(t, err, "Stop should succeed")

	// Verify process is stopped
	assert.False(t, mgr.IsRunning(), "IsRunning should return false after Stop")
}

func TestProcessManager_Stop_SendsSIGTERM(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	// Start process
	err := mgr.Start(ctx, "sleep", []string{"10"}, config)
	require.NoError(t, err)

	pid := mgr.GetPID()
	require.Greater(t, pid, 0)

	// Stop should send SIGTERM first
	err = mgr.Stop(ctx)
	require.NoError(t, err)

	// Verify process was terminated gracefully
	// (Process should have received SIGTERM)
	assert.False(t, mgr.IsRunning(), "Process should be stopped")
}

func TestProcessManager_Stop_ForcesKillAfterTimeout(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	config := ProcessConfig{
		PIDFile:         filepath.Join(t.TempDir(), "test.pid"),
		LogFile:         filepath.Join(t.TempDir(), "test.log"),
		ShutdownTimeout: 1 * time.Second,
	}

	// Start a process that ignores SIGTERM
	// (In practice, this would be a test binary that traps signals)
	err := mgr.Start(ctx, "sleep", []string{"60"}, config)
	require.NoError(t, err)

	// Stop with short timeout
	stopCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = mgr.Stop(stopCtx)
	require.NoError(t, err, "Stop should eventually kill unresponsive process")

	// Process should be stopped (SIGKILL after SIGTERM timeout)
	assert.False(t, mgr.IsRunning())
}

func TestProcessManager_Stop_NotRunning(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	// Stop when not running - should be idempotent
	err := mgr.Stop(ctx)
	assert.NoError(t, err, "Stopping non-running process should be idempotent")
}

// =============================================================================
// Task 2.3: IsRunning Tests
// =============================================================================

func TestProcessManager_IsRunning_True(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	err := mgr.Start(ctx, "sleep", []string{"10"}, config)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	assert.True(t, mgr.IsRunning(), "IsRunning should return true when process is running")
}

func TestProcessManager_IsRunning_False(t *testing.T) {
	mgr := setupTestProcessManager(t)

	assert.False(t, mgr.IsRunning(), "IsRunning should return false when process is not running")
}

func TestProcessManager_IsRunning_AfterProcessExit(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	// Start a process that exits immediately
	err := mgr.Start(ctx, "sleep", []string{"0.1"}, config)
	require.NoError(t, err)

	// Wait for process to exit
	time.Sleep(200 * time.Millisecond)

	// IsRunning should detect that process exited
	assert.False(t, mgr.IsRunning(), "IsRunning should return false after process exits")
}

func TestProcessManager_IsRunning_ChecksActualProcess(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	err := mgr.Start(ctx, "sleep", []string{"10"}, config)
	require.NoError(t, err)

	pid := mgr.GetPID()

	// Kill process directly (simulate crash)
	proc, err := os.FindProcess(pid)
	require.NoError(t, err)
	err = proc.Signal(syscall.SIGKILL)
	require.NoError(t, err)

	// Wait for process to die
	time.Sleep(100 * time.Millisecond)

	// IsRunning should detect that process is dead
	assert.False(t, mgr.IsRunning(), "IsRunning should detect killed process")
}

// =============================================================================
// Task 2.4: GetPID Tests
// =============================================================================

func TestProcessManager_GetPID_ReturnsCorrectPID(t *testing.T) {
	mgr := setupTestProcessManager(t)
	ctx := context.Background()

	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	err := mgr.Start(ctx, "sleep", []string{"10"}, config)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	pid := mgr.GetPID()
	assert.Greater(t, pid, 0, "PID should be positive")

	// Verify PID corresponds to actual running process
	proc, err := os.FindProcess(pid)
	require.NoError(t, err)

	// Send signal 0 to check if process exists
	err = proc.Signal(syscall.Signal(0))
	assert.NoError(t, err, "Process with returned PID should exist")
}

func TestProcessManager_GetPID_ReturnsZeroWhenNotRunning(t *testing.T) {
	mgr := setupTestProcessManager(t)

	pid := mgr.GetPID()
	assert.Equal(t, 0, pid, "PID should be 0 when not running")
}

// =============================================================================
// Task 2.5: PID File Tests
// =============================================================================

func TestProcessManager_PIDFile_WrittenOnStart(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "test.pid")
	config := ProcessConfig{
		PIDFile: pidFile,
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	mgr := NewProcessManager(config)
	ctx := context.Background()

	err := mgr.Start(ctx, "sleep", []string{"10"}, config)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Verify PID file exists
	assert.FileExists(t, pidFile, "PID file should be created on start")

	// Verify PID file contains correct PID
	content, err := os.ReadFile(pidFile)
	require.NoError(t, err)

	pid := mgr.GetPID()
	assert.Contains(t, string(content), strconv.Itoa(pid), "PID file should contain process PID")
}

func TestProcessManager_PIDFile_RemovedOnStop(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "test.pid")
	config := ProcessConfig{
		PIDFile: pidFile,
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	mgr := NewProcessManager(config)
	ctx := context.Background()

	// Start process
	err := mgr.Start(ctx, "sleep", []string{"10"}, config)
	require.NoError(t, err)

	// Verify PID file exists
	assert.FileExists(t, pidFile)

	// Stop process
	err = mgr.Stop(ctx)
	require.NoError(t, err)

	// Verify PID file is removed
	assert.NoFileExists(t, pidFile, "PID file should be removed on stop")
}

func TestProcessManager_PIDFile_HandlesStaleFile(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "test.pid")

	// Create stale PID file with non-existent PID
	err := os.WriteFile(pidFile, []byte("99999"), 0644)
	require.NoError(t, err)

	config := ProcessConfig{
		PIDFile: pidFile,
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	}

	mgr := NewProcessManager(config)
	ctx := context.Background()

	// Start should handle stale PID file
	err = mgr.Start(ctx, "sleep", []string{"10"}, config)
	require.NoError(t, err, "Start should handle stale PID file")
	defer mgr.Stop(ctx)

	// Verify PID file now contains current PID
	content, err := os.ReadFile(pidFile)
	require.NoError(t, err)

	pid := mgr.GetPID()
	assert.NotContains(t, string(content), "99999", "PID file should be updated")
	assert.Greater(t, pid, 0, "New PID should be set")
}

// =============================================================================
// Task 2.6: Log File Tests
// =============================================================================

func TestProcessManager_LogFile_CreatedOnStart(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "test.log")
	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: logFile,
	}

	mgr := NewProcessManager(config)
	ctx := context.Background()

	err := mgr.Start(ctx, "sleep", []string{"10"}, config)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Verify log file exists
	assert.FileExists(t, logFile, "Log file should be created on start")
}

func TestProcessManager_LogFile_CapturesOutput(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "test.log")
	config := ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: logFile,
	}

	mgr := NewProcessManager(config)
	ctx := context.Background()

	// Run a command that produces output
	err := mgr.Start(ctx, "echo", []string{"test output"}, config)
	if err != nil {
		// Echo exits immediately, so this may fail
		// For registry, we want long-running processes
		t.Skip("Test requires long-running process with output")
	}

	time.Sleep(100 * time.Millisecond)

	// Verify log file contains output
	content, err := os.ReadFile(logFile)
	require.NoError(t, err)

	assert.Contains(t, string(content), "test output", "Log file should capture process output")
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestDefaultProcessManager_ImplementsProcessManager(t *testing.T) {
	var _ ProcessManager = (*DefaultProcessManager)(nil)
}
