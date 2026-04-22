package buildlog_test

// =============================================================================
// TDD Phase 2 — RED security tests for pkg/buildlog
//
// These tests document the security hardening requirements from the security
// review comment on issue #400.
//
// Run: go test ./pkg/buildlog/... -v -run TestSecurity -count=1
// =============================================================================

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devopsmaestro/pkg/buildlog"
)

// =============================================================================
// SECURITY-1: File and directory permission modes
// =============================================================================

// TestSecurity_DirectoryMode_Is0700 verifies that the log directory is created
// with mode 0700, regardless of the process umask.
func TestSecurity_DirectoryMode_Is0700(t *testing.T) {
	parent := t.TempDir()
	logDir := filepath.Join(parent, "logs", "builds")

	opts := buildlog.Options{
		Enabled:    true,
		Directory:  logDir,
		MaxSizeMB:  100,
		MaxAgeDays: 7,
		MaxBackups: 5,
	}
	logger, err := buildlog.New(opts)
	require.NoError(t, err)
	require.NoError(t, logger.Open("perm-dir-session"))
	require.NoError(t, logger.Close())

	info, statErr := os.Stat(logDir)
	require.NoError(t, statErr)
	got := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0o700), got,
		"log directory must be mode 0700, got %04o", got)
}

// TestSecurity_FileMode_Is0600 verifies that the log file itself is created
// with mode 0600, regardless of the process umask.
func TestSecurity_FileMode_Is0600(t *testing.T) {
	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	sid := "perm-file-session-001"
	require.NoError(t, logger.Open(sid))
	_, _ = fmt.Fprint(logger.Writer("ws"), "some data\n")
	require.NoError(t, logger.Close())

	info, statErr := os.Stat(filepath.Join(opts.Directory, sid+".log"))
	require.NoError(t, statErr)
	got := info.Mode().Perm()
	assert.Equal(t, os.FileMode(0o600), got,
		"log file must be mode 0600, got %04o", got)
}

// =============================================================================
// SECURITY-2: Atomic symlink swap
// =============================================================================

// TestSecurity_SymlinkSwap_IsAtomic verifies that the latest.log symlink is
// swapped via a .tmp file + Rename (atomic on POSIX), and never via a
// Remove + Symlink sequence (which leaves a window with no symlink).
//
// This test asserts the _outcome_: the symlink must exist and be valid
// immediately after Open, with no observable ENOENT window. We also assert
// that no stale "latest.log.tmp" file remains after Open completes.
func TestSecurity_SymlinkSwap_IsAtomic(t *testing.T) {
	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)
	require.NoError(t, logger.Open("atomic-symlink-session"))
	defer logger.Close() //nolint:errcheck

	symlinkPath := filepath.Join(opts.Directory, "latest.log")
	tmpPath := symlinkPath + ".tmp"

	// The symlink must exist.
	_, lstatErr := os.Lstat(symlinkPath)
	assert.NoError(t, lstatErr, "latest.log symlink must exist after Open")

	// No stale .tmp file must remain.
	_, tmpErr := os.Lstat(tmpPath)
	assert.True(t, os.IsNotExist(tmpErr),
		"latest.log.tmp must not exist after Open completes (atomic rename left stale file)")
}

// TestSecurity_SymlinkTarget_IsNotFollowed verifies that the logger uses
// Lstat (not Stat) when inspecting the latest.log path — i.e., it never
// follows an existing symlink to check existence before swapping.
//
// We simulate a pre-existing "latest.log" pointing to a file in a different
// location and verify the logger overwrites the symlink safely without
// following the old target.
func TestSecurity_SymlinkTarget_IsNotFollowed(t *testing.T) {
	opts := newEnabledOpts(t)

	// Plant a symlink at latest.log pointing to a file outside the log dir.
	outsideFile := filepath.Join(t.TempDir(), "outside.txt")
	require.NoError(t, os.WriteFile(outsideFile, []byte("original"), 0o600))
	existingSymlink := filepath.Join(opts.Directory, "latest.log")
	require.NoError(t, os.Symlink(outsideFile, existingSymlink))

	logger, err := buildlog.New(opts)
	require.NoError(t, err)
	require.NoError(t, logger.Open("symlink-overwrite-session"))
	defer logger.Close() //nolint:errcheck

	// The outside file must not have been written to — the logger replaced the
	// symlink rather than following it to write into the outside file.
	data, readErr := os.ReadFile(outsideFile)
	require.NoError(t, readErr)
	assert.Equal(t, "original", string(data),
		"logger must not follow existing symlink and write through it")

	// The symlink must now point to the new session file.
	target, readlinkErr := os.Readlink(existingSymlink)
	require.NoError(t, readlinkErr)
	assert.Contains(t, target, "symlink-overwrite-session")
}

// =============================================================================
// SECURITY-3: Session ID validation
// =============================================================================

// TestSecurity_SessionID_Validation verifies that Open rejects session IDs
// that could cause path traversal or other injection attacks.
func TestSecurity_SessionID_Validation(t *testing.T) {
	tests := []struct {
		name      string
		sessionID string
		wantErr   bool
	}{
		// Valid UUIDs and safe identifiers must be accepted.
		{"valid uuid", "a1b2c3d4-e5f6-7890-abcd-ef1234567890", false},
		{"valid alphanumeric", "session123", false},
		{"valid with dots", "session.1.2.3", false},
		{"valid with underscore", "session_foo", false},

		// Malicious inputs must be rejected.
		{"path traversal dotdot", "../etc/passwd", true},
		{"path traversal slash", "session/../../etc", true},
		{"absolute path", "/etc/passwd", true},
		{"NUL byte", "session\x00evil", true},
		{"leading dot", ".hidden", true},
		{"too long (>128 chars)", fmt.Sprintf("%0129d", 0), true},
		{"empty string", "", true},
		{"backslash", `session\evil`, true},
		{"control char", "session\x01foo", true},
	}

	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			openErr := logger.Open(tc.sessionID)
			if tc.wantErr {
				assert.Error(t, openErr,
					"Open(%q) must return error for unsafe session ID", tc.sessionID)
			} else {
				assert.NoError(t, openErr,
					"Open(%q) must not error for safe session ID", tc.sessionID)
				_ = logger.Close()
			}
		})
	}
}

// TestSecurity_WorkspaceName_ControlCharsStripped verifies that control
// characters and ANSI escape sequences in workspace names are sanitised
// before being written to the log (prevents terminal-escape log injection
// when the log is viewed with `tail -f`).
func TestSecurity_WorkspaceName_ControlCharsStripped(t *testing.T) {
	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)
	require.NoError(t, logger.Open("ws-sanitise-session-001"))
	defer logger.Close() //nolint:errcheck

	// Write using a workspace name that contains an ANSI escape sequence.
	maliciousWS := "ws\x1b[31mRED\x1b[0m"
	w := logger.Writer(maliciousWS)
	_, writeErr := fmt.Fprint(w, "test line\n")
	require.NoError(t, writeErr)
	require.NoError(t, logger.Close())

	data, readErr := os.ReadFile(logger.Path())
	require.NoError(t, readErr)
	assert.NotContains(t, string(data), "\x1b",
		"log file must not contain raw ANSI escape sequences from workspace name")
}

// =============================================================================
// SECURITY-4: Directory config path safety
// =============================================================================

// TestSecurity_Directory_RejectsForbiddenSystemPaths verifies that New returns
// an error when the resolved directory path falls under a forbidden system
// prefix (mirroring the pattern in operators/mount_validation.go).
func TestSecurity_Directory_RejectsForbiddenSystemPaths(t *testing.T) {
	forbidden := []string{
		"/etc/buildlogs",
		"/var/log/dvm",
		"/usr/share/dvm",
		"/bin/logs",
		"/sbin/logs",
	}

	for _, dir := range forbidden {
		t.Run(dir, func(t *testing.T) {
			opts := buildlog.Options{
				Enabled:    true,
				Directory:  dir,
				MaxSizeMB:  100,
				MaxAgeDays: 7,
				MaxBackups: 5,
			}
			_, err := buildlog.New(opts)
			assert.Error(t, err,
				"New must reject forbidden system path %q as log directory", dir)
		})
	}
}

// TestSecurity_Directory_RejectsFilesystemRoot verifies that "/" is rejected
// as a log directory.
func TestSecurity_Directory_RejectsFilesystemRoot(t *testing.T) {
	opts := buildlog.Options{
		Enabled:    true,
		Directory:  "/",
		MaxSizeMB:  100,
		MaxAgeDays: 7,
		MaxBackups: 5,
	}
	_, err := buildlog.New(opts)
	assert.Error(t, err, "New must reject filesystem root as log directory")
}

// TestSecurity_Directory_RejectsSymlinkDir verifies that when the resolved
// log directory path is a symlink, New returns an error rather than writing
// through the symlink to a potentially unsafe location.
func TestSecurity_Directory_RejectsSymlinkDir(t *testing.T) {
	realDir := t.TempDir()
	symlinkDir := filepath.Join(t.TempDir(), "link-to-dir")
	require.NoError(t, os.Symlink(realDir, symlinkDir))

	opts := buildlog.Options{
		Enabled:    true,
		Directory:  symlinkDir,
		MaxSizeMB:  100,
		MaxAgeDays: 7,
		MaxBackups: 5,
	}
	_, err := buildlog.New(opts)
	assert.Error(t, err,
		"New must reject a log directory that is itself a symlink")
}
