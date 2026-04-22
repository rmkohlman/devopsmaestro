// Package buildlog provides a rotating per-session build log file facility.
//
// The BuildLogger interface owns the lifecycle of one log file per build
// session. Writers are scoped to a workspace name, are line-buffered, and
// are safe for concurrent use from multiple goroutines.
//
// # Usage
//
//	logger, err := New(cfg)
//	_ = logger.Open(sessionID)
//	defer logger.Close()
//	out := io.MultiWriter(os.Stdout, logger.Writer(workspace))
package buildlog

import "io"

// BuildLogger manages a rotating log file for a single build session.
type BuildLogger interface {
	// Open prepares the per-session log file. Calling Open more than once on
	// the same logger is a no-op after the first successful call.
	Open(sessionID string) error

	// Writer returns an io.Writer scoped to the named workspace. Writes are
	// line-buffered and serialised under a shared mutex so concurrent writes
	// from multiple workspaces never interleave within a single line.
	Writer(workspace string) io.Writer

	// Path returns the absolute path of the active log file, or "" if Open
	// has not been called.
	Path() string

	// Close flushes buffered data and releases the underlying file handle.
	Close() error
}

// Options carries configuration for the file-backed BuildLogger.
type Options struct {
	// Enabled controls whether a real file logger or a no-op is returned by
	// New. When false, New returns a NoopLogger regardless of other fields.
	Enabled bool

	// Directory is the directory in which log files are created. The tilde
	// prefix (~) is expanded to the current user's home directory.
	// Default: ~/.devopsmaestro/logs/builds
	Directory string

	// MaxSizeMB is the maximum size in megabytes of the active log file before
	// it is rotated. Must be > 0. Default: 100.
	MaxSizeMB int

	// MaxAgeDays is the maximum number of days to retain rotated log files.
	// Must be > 0. Default: 7.
	MaxAgeDays int

	// MaxBackups is the maximum number of rotated log files to retain.
	// Must be > 0. Default: 10.
	MaxBackups int

	// Compress enables gzip compression of rotated log files. Default: true.
	Compress bool
}

// New is the factory function for BuildLogger. It returns a noopLogger when
// opts.Enabled is false, and a fileLogger backed by lumberjack otherwise.
//
// When Enabled is true, New validates numeric bounds and the directory path,
// creates the directory with mode 0o700, and runs a startup sweep that
// removes any pre-existing log files older than MaxAgeDays.
func New(opts Options) (BuildLogger, error) {
	if !opts.Enabled {
		return &noopLogger{}, nil
	}
	return newFileLogger(opts)
}
