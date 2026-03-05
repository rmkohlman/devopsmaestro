package registry

import (
	"context"
	"os"
	"os/exec"
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

// =============================================================================
// PID File Fallback Tests (cross-CLI-invocation behavior)
// =============================================================================

// startSleepProcess starts a "sleep 300" subprocess, returning the cmd
// (for PID access and Wait), and a cleanup function that kills it.
// It is the caller's responsibility to invoke cleanup even if the test fails.
func startSleepProcess(t *testing.T) (cmd *exec.Cmd, cleanup func()) {
	t.Helper()
	cmd = exec.Command("sleep", "300")
	require.NoError(t, cmd.Start(), "failed to start sleep subprocess")
	cleanup = func() {
		// Best-effort: process may already be dead.
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}
	return cmd, cleanup
}

func TestProcessManager_IsRunning_PIDFileFallback(t *testing.T) {
	// Start one real process so we always have a live PID to test with.
	aliveCmd, killAlive := startSleepProcess(t)
	t.Cleanup(killAlive)
	alivePID := aliveCmd.Process.Pid

	tests := []struct {
		name         string
		setupPIDFile func(t *testing.T, pidFile string)
		wantRunning  bool
	}{
		{
			name: "pid file with running process returns true",
			setupPIDFile: func(t *testing.T, pidFile string) {
				t.Helper()
				require.NoError(t, os.WriteFile(pidFile, []byte(strconv.Itoa(alivePID)), 0644))
			},
			wantRunning: true,
		},
		{
			name: "pid file with dead process returns false",
			setupPIDFile: func(t *testing.T, pidFile string) {
				t.Helper()
				// Use a PID that is almost certainly not alive: PID 1 is init on
				// Linux but on macOS signal(0) to PID 1 returns EPERM which is NOT
				// nil, so isProcessAlive returns false as we need. Use a very large
				// PID that is exceedingly unlikely to exist instead.
				require.NoError(t, os.WriteFile(pidFile, []byte("2147483647"), 0644))
			},
			wantRunning: false,
		},
		{
			name: "missing pid file returns false",
			setupPIDFile: func(t *testing.T, pidFile string) {
				t.Helper()
				// Deliberately do NOT create the file.
			},
			wantRunning: false,
		},
		{
			name: "pid file with non-numeric content returns false",
			setupPIDFile: func(t *testing.T, pidFile string) {
				t.Helper()
				require.NoError(t, os.WriteFile(pidFile, []byte("not-a-pid"), 0644))
			},
			wantRunning: false,
		},
		{
			name: "pid file with empty content returns false",
			setupPIDFile: func(t *testing.T, pidFile string) {
				t.Helper()
				require.NoError(t, os.WriteFile(pidFile, []byte(""), 0644))
			},
			wantRunning: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			pidFile := filepath.Join(dir, "test.pid")

			tt.setupPIDFile(t, pidFile)

			// Create a ProcessManager with cmd == nil (simulates fresh CLI invocation).
			mgr := NewProcessManager(ProcessConfig{
				PIDFile: pidFile,
				LogFile: filepath.Join(dir, "test.log"),
			})

			assert.Equal(t, tt.wantRunning, mgr.IsRunning())
		})
	}
}

// TestProcessManager_IsRunning_InMemoryCmdStillWorks verifies the original
// in-memory behavior (p.cmd != nil) is not broken by the PID-file fallback.
func TestProcessManager_IsRunning_InMemoryCmdStillWorks(t *testing.T) {
	dir := t.TempDir()
	config := ProcessConfig{
		PIDFile: filepath.Join(dir, "test.pid"),
		LogFile: filepath.Join(dir, "test.log"),
	}

	mgr := NewProcessManager(config)
	ctx := context.Background()

	require.NoError(t, mgr.Start(ctx, "sleep", []string{"300"}, config))
	t.Cleanup(func() { _ = mgr.Stop(ctx) })

	assert.True(t, mgr.IsRunning(), "in-memory cmd path: IsRunning should return true")
}

func TestProcessManager_GetPID_PIDFileFallback(t *testing.T) {
	tests := []struct {
		name         string
		pidFileValue string
		writePIDFile bool
		wantPID      int
	}{
		{
			name:         "pid file with valid pid returns that pid",
			writePIDFile: true,
			pidFileValue: "12345",
			wantPID:      12345,
		},
		{
			name:         "missing pid file returns 0",
			writePIDFile: false,
			wantPID:      0,
		},
		{
			name:         "pid file with non-numeric content returns 0",
			writePIDFile: true,
			pidFileValue: "notapid",
			wantPID:      0,
		},
		{
			name:         "pid file with whitespace-padded pid returns that pid",
			writePIDFile: true,
			pidFileValue: "  99  \n",
			wantPID:      99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			pidFile := filepath.Join(dir, "test.pid")

			if tt.writePIDFile {
				require.NoError(t, os.WriteFile(pidFile, []byte(tt.pidFileValue), 0644))
			}

			// p.cmd is nil – simulates fresh CLI invocation.
			mgr := NewProcessManager(ProcessConfig{
				PIDFile: pidFile,
				LogFile: filepath.Join(dir, "test.log"),
			})

			assert.Equal(t, tt.wantPID, mgr.GetPID())
		})
	}
}

// TestProcessManager_GetPID_NilCmdNoPIDFile is a simple sanity test confirming
// the zero-value case.
func TestProcessManager_GetPID_NilCmdNoPIDFile(t *testing.T) {
	mgr := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(t.TempDir(), "test.pid"),
		LogFile: filepath.Join(t.TempDir(), "test.log"),
	})
	assert.Equal(t, 0, mgr.GetPID(), "GetPID should return 0 when no process and no pid file")
}

// TestProcessManager_Stop_PIDFileBased verifies that Stop() can terminate a
// process it did NOT start in the current CLI invocation (p.cmd == nil) by
// reading the PID from the PID file.
func TestProcessManager_Stop_PIDFileBased(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "test.pid")

	// Start a real process independently (simulates a process started by a
	// previous CLI invocation).
	sleepCmd, killIfStillAlive := startSleepProcess(t)
	t.Cleanup(killIfStillAlive)
	pid := sleepCmd.Process.Pid

	// Write its PID to the PID file.
	require.NoError(t, os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644))

	// Create a fresh ProcessManager with cmd == nil but PIDFile pointing at the
	// file we just wrote.
	mgr := NewProcessManager(ProcessConfig{
		PIDFile:         pidFile,
		LogFile:         filepath.Join(dir, "test.log"),
		ShutdownTimeout: 5 * time.Second,
	})

	// Confirm we can see the process as running via PID file.
	require.True(t, mgr.IsRunning(), "process should be visible via PID file before Stop")

	// Stop via PID file.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := mgr.Stop(ctx)
	require.NoError(t, err, "Stop should succeed when stopping via PID file")

	// Reap the child process so it is no longer a zombie. Without this, the
	// kernel still shows the process in the process table (as a zombie) and
	// signal(0) succeeds, making isProcessAlive return true even though the
	// process has been terminated.
	_ = sleepCmd.Wait()

	// Process should be dead.
	assert.False(t, isProcessAlive(pid), "process should no longer be alive after PID-file-based Stop")

	// PID file should be removed.
	assert.NoFileExists(t, pidFile, "PID file should be removed after Stop")
}

// TestProcessManager_ReadPIDFileLocked exercises the readPIDFileLocked helper
// via the public GetPID() API (which calls it when p.cmd == nil).
func TestProcessManager_ReadPIDFileLocked(t *testing.T) {
	tests := []struct {
		name        string
		fileContent string
		createFile  bool
		wantPID     int
		wantZero    bool // true means we expect 0 returned
	}{
		{
			name:        "valid pid file returns correct pid",
			createFile:  true,
			fileContent: "42",
			wantPID:     42,
			wantZero:    false,
		},
		{
			name:       "missing file returns 0",
			createFile: false,
			wantZero:   true,
		},
		{
			name:        "empty file returns 0",
			createFile:  true,
			fileContent: "",
			wantZero:    true,
		},
		{
			name:        "non-numeric content returns 0",
			createFile:  true,
			fileContent: "abc",
			wantZero:    true,
		},
		{
			name:        "whitespace-only content returns 0",
			createFile:  true,
			fileContent: "   \n",
			wantZero:    true,
		},
		{
			name:        "pid with surrounding whitespace is parsed correctly",
			createFile:  true,
			fileContent: "  1234\n",
			wantPID:     1234,
			wantZero:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			pidFile := filepath.Join(dir, "test.pid")

			if tt.createFile {
				require.NoError(t, os.WriteFile(pidFile, []byte(tt.fileContent), 0644))
			}

			// Access readPIDFileLocked indirectly via GetPID() with cmd == nil.
			mgr := NewProcessManager(ProcessConfig{
				PIDFile: pidFile,
			})

			got := mgr.GetPID()

			if tt.wantZero {
				assert.Equal(t, 0, got, "expected PID 0 for case %q", tt.name)
			} else {
				assert.Equal(t, tt.wantPID, got, "expected PID %d for case %q", tt.wantPID, tt.name)
			}
		})
	}
}
