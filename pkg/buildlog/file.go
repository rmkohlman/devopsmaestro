package buildlog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"syscall"

	"gopkg.in/natefinch/lumberjack.v2"
)

// fileLogger is the lumberjack-backed BuildLogger implementation.
//
// Concurrency model: a single mutex serialises all writes through the
// underlying *lumberjack.Logger. Writer(workspace) returns a thin wrapper
// that takes the mutex on every Write, so concurrent goroutines never
// interleave bytes within a single Write call. Callers should write whole
// lines per Write (fmt.Fprintf does this).
type fileLogger struct {
	opts      Options
	directory string // resolved absolute directory

	mu        sync.Mutex
	opened    bool
	sessionID string
	path      string
	lj        *lumberjack.Logger
	writer    *lockedWriter
}

// newFileLogger constructs a validated, sweep-cleaned fileLogger.
func newFileLogger(opts Options) (*fileLogger, error) {
	if err := validateNumericBounds(opts); err != nil {
		return nil, err
	}
	dir, err := validateDirectory(opts.Directory)
	if err != nil {
		return nil, err
	}
	if err := mkdirSecure(dir); err != nil {
		return nil, err
	}
	if err := startupSweep(dir, opts.MaxAgeDays); err != nil {
		// Sweep failures are not fatal — log to stderr and continue.
		fmt.Fprintf(os.Stderr, "buildlog: startup sweep error: %v\n", err)
	}
	return &fileLogger{
		opts:      opts,
		directory: dir,
	}, nil
}

// mkdirSecure creates the directory tree with mode 0o700, explicitly
// overriding the process umask so the final directory is not world-readable.
func mkdirSecure(dir string) error {
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("buildlog: cannot create directory %q: %w", dir, err)
	}
	// Belt-and-suspenders: chmod even if it already existed with looser perms.
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("buildlog: cannot chmod directory %q to 0700: %w", dir, err)
	}
	return nil
}

// Open prepares the per-session log file. Idempotent: subsequent calls (with
// any sessionID) on an already-opened logger are a no-op.
func (f *fileLogger) Open(sessionID string) error {
	if err := validateSessionID(sessionID); err != nil {
		return err
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.opened {
		return nil
	}
	path := filepath.Join(f.directory, sessionID+".log")

	// Force file creation under a tight umask so lumberjack's first Write
	// produces a 0o600 file even on systems with a permissive umask.
	old := syscall.Umask(0o077)
	defer syscall.Umask(old)

	lj := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    f.opts.MaxSizeMB,
		MaxAge:     f.opts.MaxAgeDays,
		MaxBackups: f.opts.MaxBackups,
		Compress:   f.opts.Compress,
		LocalTime:  true,
	}
	// Eagerly create the file so callers that only invoke Open (no writes)
	// still see the file on disk, and so we can chmod it to 0o600.
	touch, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return fmt.Errorf("buildlog: cannot create log file %q: %w", path, err)
	}
	_ = touch.Close()
	if err := os.Chmod(path, 0o600); err != nil {
		return fmt.Errorf("buildlog: cannot chmod log file %q to 0600: %w", path, err)
	}

	f.lj = lj
	f.path = path
	f.sessionID = sessionID
	f.opened = true
	f.writer = &lockedWriter{mu: &f.mu, w: lj}

	if err := updateLatestSymlink(f.directory, path); err != nil {
		// Symlink failure should not abort logging — warn to stderr.
		fmt.Fprintf(os.Stderr, "buildlog: failed to update latest.log symlink: %v\n", err)
	}
	return nil
}

// Writer returns the shared, mutex-guarded writer. The workspace name is
// accepted for API symmetry but is not echoed to the file (callers prefix
// their own lines as needed). This avoids ANSI/control-char injection via
// workspace names while keeping the concurrency contract simple.
func (f *fileLogger) Writer(workspace string) io.Writer {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.opened {
		return io.Discard
	}
	return f.writer
}

// Path returns the absolute path of the active log file.
func (f *fileLogger) Path() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.path
}

// Close flushes and releases the underlying file handle. Subsequent calls
// are safe and return nil.
func (f *fileLogger) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.opened {
		return nil
	}
	err := f.lj.Close()
	f.opened = false
	return err
}

// updateLatestSymlink atomically updates <dir>/latest.log to point at target.
// It writes the symlink to a .tmp name and renames into place — never
// removes the existing symlink first (avoids an ENOENT window for readers).
// Uses Lstat (not Stat) so it never follows an attacker-planted symlink.
func updateLatestSymlink(dir, target string) error {
	link := filepath.Join(dir, "latest.log")
	tmp := link + ".tmp"

	// Clean any stale .tmp from a prior crash. Lstat to avoid following.
	if _, err := os.Lstat(tmp); err == nil {
		_ = os.Remove(tmp)
	}
	if err := os.Symlink(target, tmp); err != nil {
		return fmt.Errorf("symlink tmp: %w", err)
	}
	if err := os.Rename(tmp, link); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename tmp -> latest.log: %w", err)
	}
	return nil
}
