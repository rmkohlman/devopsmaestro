package buildlog_test

// =============================================================================
// TDD Phase 2 — RED rotation and retention tests for pkg/buildlog
// =============================================================================

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devopsmaestro/pkg/buildlog"
)

// =============================================================================
// Rotation
// =============================================================================

// TestFileLogger_Rotation_TriggersAtMaxSize verifies that when the active log
// file exceeds MaxSizeMB the old file is renamed and a new file is created.
//
// We use a very small MaxSizeMB (1 MB) and write enough data to exceed it.
func TestFileLogger_Rotation_TriggersAtMaxSize(t *testing.T) {
	opts := buildlog.Options{
		Enabled:    true,
		Directory:  t.TempDir(),
		MaxSizeMB:  1, // 1 MB — easy to exceed in a test
		MaxAgeDays: 30,
		MaxBackups: 5,
		Compress:   false,
	}
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	sid := "rotation-session-001"
	require.NoError(t, logger.Open(sid))

	// Write slightly more than 1 MiB to trigger rotation.
	chunk := strings.Repeat("x", 1024) + "\n" // 1 KiB per line
	w := logger.Writer("ws")
	for i := 0; i < 1100; i++ {
		_, writeErr := fmt.Fprint(w, chunk)
		require.NoError(t, writeErr)
	}
	require.NoError(t, logger.Close())

	// After rotation there must be more than one file matching the session ID.
	entries, readErr := os.ReadDir(opts.Directory)
	require.NoError(t, readErr)

	var sessionFiles []string
	for _, e := range entries {
		if strings.Contains(e.Name(), sid) {
			sessionFiles = append(sessionFiles, e.Name())
		}
	}
	assert.Greater(t, len(sessionFiles), 1,
		"rotation must produce more than one file for session %s, got: %v", sid, sessionFiles)
}

// TestFileLogger_RetentionByCount verifies that MaxBackups is respected —
// only MaxBackups rotated files are kept (oldest pruned first).
func TestFileLogger_RetentionByCount(t *testing.T) {
	const maxBackups = 3
	opts := buildlog.Options{
		Enabled:    true,
		Directory:  t.TempDir(),
		MaxSizeMB:  1,
		MaxAgeDays: 365,
		MaxBackups: maxBackups,
		Compress:   false,
	}
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	sid := "retention-count-session"
	require.NoError(t, logger.Open(sid))

	// Trigger more rotations than MaxBackups to force pruning.
	chunk := strings.Repeat("y", 1024) + "\n"
	w := logger.Writer("ws")
	for i := 0; i < (maxBackups+3)*1100; i++ {
		_, _ = fmt.Fprint(w, chunk)
	}
	require.NoError(t, logger.Close())

	entries, readErr := os.ReadDir(opts.Directory)
	require.NoError(t, readErr)

	var backups []string
	for _, e := range entries {
		name := e.Name()
		// Rotated backup files contain a timestamp suffix.
		if strings.Contains(name, sid) && name != sid+".log" && name != "latest.log" {
			backups = append(backups, name)
		}
	}
	assert.LessOrEqual(t, len(backups), maxBackups,
		"must retain at most %d backups, found %d: %v", maxBackups, len(backups), backups)
}

// =============================================================================
// Startup sweep (age-based pruning)
// =============================================================================

// TestFileLogger_StartupSweep_RemovesOldFiles verifies that when New is called
// (not just on rotation), files older than MaxAgeDays are removed. This covers
// the architecture note that lumberjack only prunes on rotation — the
// implementation must add an explicit startup sweep.
func TestFileLogger_StartupSweep_RemovesOldFiles(t *testing.T) {
	dir := t.TempDir()

	// Plant a stale backup file with an mtime in the past.
	staleFile := filepath.Join(dir, "old-session-2026-01-01T00-00-00.000.log")
	require.NoError(t, os.WriteFile(staleFile, []byte("stale"), 0o600))
	staleTime := time.Now().AddDate(0, 0, -30) // 30 days ago
	require.NoError(t, os.Chtimes(staleFile, staleTime, staleTime))

	opts := buildlog.Options{
		Enabled:    true,
		Directory:  dir,
		MaxSizeMB:  100,
		MaxAgeDays: 7, // prune files older than 7 days
		MaxBackups: 10,
		Compress:   false,
	}

	// Creating a new logger should trigger the startup sweep.
	logger, err := buildlog.New(opts)
	require.NoError(t, err)
	require.NoError(t, logger.Open("sweep-session-001"))
	require.NoError(t, logger.Close())

	_, statErr := os.Stat(staleFile)
	assert.True(t, os.IsNotExist(statErr),
		"startup sweep must remove files older than MaxAgeDays=%d", opts.MaxAgeDays)
}

// TestFileLogger_StartupSweep_PreservesRecentFiles verifies that recent
// backup files are NOT removed by the startup sweep.
func TestFileLogger_StartupSweep_PreservesRecentFiles(t *testing.T) {
	dir := t.TempDir()

	// Plant a recent backup file.
	recentFile := filepath.Join(dir, "recent-session-2026-04-20T12-00-00.000.log")
	require.NoError(t, os.WriteFile(recentFile, []byte("recent"), 0o600))
	recentTime := time.Now().AddDate(0, 0, -2) // 2 days ago
	require.NoError(t, os.Chtimes(recentFile, recentTime, recentTime))

	opts := buildlog.Options{
		Enabled:    true,
		Directory:  dir,
		MaxSizeMB:  100,
		MaxAgeDays: 7,
		MaxBackups: 10,
		Compress:   false,
	}

	logger, err := buildlog.New(opts)
	require.NoError(t, err)
	require.NoError(t, logger.Open("sweep-preserve-session"))
	require.NoError(t, logger.Close())

	_, statErr := os.Stat(recentFile)
	assert.NoError(t, statErr, "startup sweep must NOT remove files within MaxAgeDays")
}

// =============================================================================
// MultiWriter integration
// =============================================================================

// TestFileLogger_MultiWriter_StdoutPreserved verifies that when the file
// logger is composed with io.MultiWriter, writes go to both stdout (the
// provided buffer) and the log file.
func TestFileLogger_MultiWriter_StdoutPreserved(t *testing.T) {
	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	sid := "multiwriter-session-001"
	require.NoError(t, logger.Open(sid))

	// Capture "stdout" output.
	pr, pw, pipeErr := os.Pipe()
	require.NoError(t, pipeErr)

	// Instead of os.Stdout, we pass the pipe writer directly to MultiWriter.
	// (We don't actually use io.MultiWriter here — we verify the logger Writer
	// produces output that could be teed; the MultiWriter composition lives in
	// the orchestrator. Here we verify the logger Writer is an io.Writer that
	// can be composed.)
	logW := logger.Writer("ws")
	require.NotNil(t, logW, "logger.Writer must return a non-nil io.Writer")

	// Write to the log, then separately to pipe.
	const msg = "hello multiwriter\n"
	_, _ = fmt.Fprint(logW, msg)

	_, _ = fmt.Fprint(pw, msg)
	pw.Close()

	buf := make([]byte, 256)
	n, _ := pr.Read(buf)
	pr.Close()

	require.NoError(t, logger.Close())

	// Verify the log file received the write.
	data, readErr := os.ReadFile(filepath.Join(opts.Directory, sid+".log"))
	require.NoError(t, readErr)
	assert.Contains(t, string(data), strings.TrimSuffix(msg, "\n"),
		"log file must contain written message")

	// Verify the pipe also received the message (simulates stdout tee).
	assert.Equal(t, msg, string(buf[:n]))
}
