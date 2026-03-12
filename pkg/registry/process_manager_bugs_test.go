package registry

// =============================================================================
// BUG #2 (P1): Premature log file close in DefaultProcessManager.Start()
//
// Root cause: Start() has `defer logFile.Close()` at line 70 of process_manager.go.
// This means the log file is closed as soon as Start() returns — before the
// spawned process has finished writing to stdout/stderr — potentially
// corrupting or truncating log output.
//
// Expected fix:
//   - Add a `logFile *os.File` field to DefaultProcessManager.
//   - In Start(), assign logFile to p.logFile instead of deferring Close().
//   - In Stop(), close p.logFile (and set it to nil).
//
// Test strategy:
//   After Start() returns successfully, access p.logFile on DefaultProcessManager.
//
//   RED  today : DefaultProcessManager has no logFile field → compile error (RED).
//   GREEN after: field exists and is non-nil after Start().
// =============================================================================

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestProcessManager_LogFileRemainsOpenAfterStart verifies that the log file
// is kept open after Start() returns, not closed prematurely via defer.
//
// BUG #2 (P1): Start() currently uses `defer logFile.Close()` which closes
// the log file descriptor immediately when Start() returns. The spawned
// process still holds the same fd open (via Stdout/Stderr redirect), but
// on systems that enforce fd validity this can corrupt log output.
//
// RED  today : DefaultProcessManager has no `logFile` field → compile error.
// GREEN after: p.logFile is non-nil after Start() succeeds.
func TestProcessManager_LogFileRemainsOpenAfterStart(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")
	pidPath := filepath.Join(tmpDir, "test.pid")

	pm := &DefaultProcessManager{}
	ctx := context.Background()

	cfg := ProcessConfig{
		LogFile: logPath,
		PIDFile: pidPath,
	}

	// Start a real process that exits quickly (sleep 0).
	err := pm.Start(ctx, "/bin/sleep", []string{"0"}, cfg)
	if err != nil {
		t.Skipf("could not start /bin/sleep (non-Unix?): %v", err)
	}
	t.Cleanup(func() {
		_ = pm.Stop(context.Background())
	})

	// BUG present : p.logFile field does not exist → compile error (RED).
	// BUG fixed   : p.logFile is non-nil after Start() → assertion passes (GREEN).
	if pm.logFile == nil {
		t.Errorf("BUG #2: DefaultProcessManager.logFile is nil after Start(). " +
			"The log file was closed prematurely by `defer logFile.Close()` in Start(). " +
			"Fix: store the *os.File in pm.logFile and close it in Stop() instead.")
	}

	// Also verify the log file is still open (not closed) — try writing to it.
	if pm.logFile != nil {
		if _, err := pm.logFile.WriteString("probe\n"); err != nil {
			t.Errorf("BUG #2: log file fd is already closed after Start(): %v", err)
		}
	}

	// Verify the log file exists on disk.
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("log file was not created at %s", logPath)
	}
}
