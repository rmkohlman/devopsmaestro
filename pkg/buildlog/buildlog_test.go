package buildlog_test

// =============================================================================
// TDD Phase 2 — RED tests for pkg/buildlog
//
// All tests in this file are expected to FAIL until dvm-core implements the
// package. The tests document the exact contract the implementation must
// satisfy.
//
// Run: go test ./pkg/buildlog/... -v -count=1
// =============================================================================

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devopsmaestro/pkg/buildlog"
)

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func newEnabledOpts(t *testing.T) buildlog.Options {
	t.Helper()
	return buildlog.Options{
		Enabled:    true,
		Directory:  t.TempDir(),
		MaxSizeMB:  100,
		MaxAgeDays: 7,
		MaxBackups: 10,
		Compress:   false, // keep tests fast — no gzip
	}
}

// -----------------------------------------------------------------------------
// Factory tests
// -----------------------------------------------------------------------------

// TestNew_ReturnsNoop_WhenDisabled verifies that New returns a no-op logger
// that writes nothing to disk when Enabled=false.
func TestNew_ReturnsNoop_WhenDisabled(t *testing.T) {
	dir := t.TempDir()
	opts := buildlog.Options{
		Enabled:   false,
		Directory: dir,
	}
	logger, err := buildlog.New(opts)
	require.NoError(t, err)
	require.NotNil(t, logger)

	require.NoError(t, logger.Open("session-noop-001"))
	_, writeErr := fmt.Fprint(logger.Writer("ws"), "should not reach disk\n")
	require.NoError(t, writeErr)
	require.NoError(t, logger.Close())

	// No files should have been created in the directory.
	entries, readErr := os.ReadDir(dir)
	require.NoError(t, readErr)
	assert.Empty(t, entries, "noop logger must not create any files")
}

// TestNew_ReturnsFileLogger_WhenEnabled verifies that New returns a working
// file logger when Enabled=true.
func TestNew_ReturnsFileLogger_WhenEnabled(t *testing.T) {
	logger, err := buildlog.New(newEnabledOpts(t))
	require.NoError(t, err)
	require.NotNil(t, logger)
}

// -----------------------------------------------------------------------------
// Open / Write / Close lifecycle
// -----------------------------------------------------------------------------

// TestFileLogger_Open_CreatesFile verifies that Open creates the log file at
// <directory>/<sessionID>.log.
func TestFileLogger_Open_CreatesFile(t *testing.T) {
	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	sid := "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
	require.NoError(t, logger.Open(sid))
	defer logger.Close() //nolint:errcheck

	want := filepath.Join(opts.Directory, sid+".log")
	_, statErr := os.Stat(want)
	assert.NoError(t, statErr, "log file %s must exist after Open", want)
}

// TestFileLogger_Path_ReturnsAbsolutePath verifies that Path() returns the
// absolute path of the active log file after Open.
func TestFileLogger_Path_ReturnsAbsolutePath(t *testing.T) {
	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	sid := "path-test-session-001"
	require.NoError(t, logger.Open(sid))
	defer logger.Close() //nolint:errcheck

	p := logger.Path()
	assert.True(t, filepath.IsAbs(p), "Path() must return an absolute path, got %q", p)
	assert.Equal(t, filepath.Join(opts.Directory, sid+".log"), p)
}

// TestFileLogger_Write_AppendsLines verifies that Writer.Write appends lines
// to the log file.
func TestFileLogger_Write_AppendsLines(t *testing.T) {
	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	sid := "write-test-session-001"
	require.NoError(t, logger.Open(sid))

	w := logger.Writer("workspace-a")
	lines := []string{"line one\n", "line two\n", "line three\n"}
	for _, l := range lines {
		_, writeErr := fmt.Fprint(w, l)
		require.NoError(t, writeErr)
	}
	require.NoError(t, logger.Close())

	data, readErr := os.ReadFile(filepath.Join(opts.Directory, sid+".log"))
	require.NoError(t, readErr)
	content := string(data)
	for _, l := range lines {
		assert.Contains(t, content, strings.TrimSuffix(l, "\n"))
	}
}

// TestFileLogger_Open_IsIdempotent verifies that calling Open twice does not
// reset or truncate the log file.
func TestFileLogger_Open_IsIdempotent(t *testing.T) {
	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	sid := "idempotent-open-001"
	require.NoError(t, logger.Open(sid))
	_, _ = fmt.Fprint(logger.Writer("ws"), "first write\n")

	// Second Open must not error and must not truncate.
	require.NoError(t, logger.Open(sid))
	_, _ = fmt.Fprint(logger.Writer("ws"), "second write\n")
	require.NoError(t, logger.Close())

	data, readErr := os.ReadFile(filepath.Join(opts.Directory, sid+".log"))
	require.NoError(t, readErr)
	assert.Contains(t, string(data), "first write")
	assert.Contains(t, string(data), "second write")
}

// -----------------------------------------------------------------------------
// Concurrency safety
// -----------------------------------------------------------------------------

// TestFileLogger_Writer_ConcurrencySafe verifies that parallel writes from
// multiple workspaces do not interleave bytes within a single line.
//
// Each goroutine writes lines of the form "[workspace-N] line M\n". After all
// goroutines finish, every line in the file must still be a complete,
// non-corrupted record.
func TestFileLogger_Writer_ConcurrencySafe(t *testing.T) {
	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	require.NoError(t, logger.Open("concurrency-session-001"))

	const numWorkspaces = 10
	const linesPerWorker = 200

	var wg sync.WaitGroup
	for i := range numWorkspaces {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			w := logger.Writer(fmt.Sprintf("workspace-%d", n))
			for j := range linesPerWorker {
				_, _ = fmt.Fprintf(w, "[workspace-%d] line %d\n", n, j)
			}
		}(i)
	}
	wg.Wait()
	require.NoError(t, logger.Close())

	data, readErr := os.ReadFile(logger.Path())
	require.NoError(t, readErr)

	// Every line in the file must match the expected prefix pattern and must
	// not be empty or garbled.
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	totalExpected := numWorkspaces * linesPerWorker
	assert.Len(t, lines, totalExpected,
		"expected %d total lines, got %d — concurrent writes caused loss or interleaving",
		totalExpected, len(lines))

	for _, line := range lines {
		assert.Regexp(t, `^\[workspace-\d+\] line \d+$`, line,
			"line is garbled (partial write or interleaving): %q", line)
	}
}

// -----------------------------------------------------------------------------
// latest.log symlink
// -----------------------------------------------------------------------------

// TestFileLogger_LatestSymlink_PointsToCurrentSession verifies that after
// Open, a symlink named "latest.log" exists in the log directory and resolves
// to the active session file.
func TestFileLogger_LatestSymlink_PointsToCurrentSession(t *testing.T) {
	opts := newEnabledOpts(t)
	logger, err := buildlog.New(opts)
	require.NoError(t, err)

	sid := "symlink-session-001"
	require.NoError(t, logger.Open(sid))
	defer logger.Close() //nolint:errcheck

	symlinkPath := filepath.Join(opts.Directory, "latest.log")
	target, readlinkErr := os.Readlink(symlinkPath)
	require.NoError(t, readlinkErr, "latest.log symlink must exist")

	// The symlink target must be or resolve to the active session file.
	wantBase := sid + ".log"
	assert.Equal(t, filepath.Join(opts.Directory, wantBase), target,
		"latest.log must point to the active session file")
}

// TestFileLogger_LatestSymlink_UpdatedAcrossSessions verifies that opening a
// second session updates "latest.log" to point to the new session file.
func TestFileLogger_LatestSymlink_UpdatedAcrossSessions(t *testing.T) {
	opts := newEnabledOpts(t)

	open := func(sid string) {
		logger, err := buildlog.New(opts)
		require.NoError(t, err)
		require.NoError(t, logger.Open(sid))
		_, _ = fmt.Fprint(logger.Writer("ws"), "data\n")
		require.NoError(t, logger.Close())
	}

	open("first-session-001")
	open("second-session-002")

	target, err := os.Readlink(filepath.Join(opts.Directory, "latest.log"))
	require.NoError(t, err)
	assert.Contains(t, target, "second-session-002",
		"latest.log must point to the most recently opened session")
}

// -----------------------------------------------------------------------------
// Config validation
// -----------------------------------------------------------------------------

// TestNew_ConfigValidation exercises numeric bound checks on Options fields.
func TestNew_ConfigValidation(t *testing.T) {
	base := newEnabledOpts(t)

	tests := []struct {
		name    string
		mutate  func(*buildlog.Options)
		wantErr string
	}{
		{
			name:    "zero MaxSizeMB",
			mutate:  func(o *buildlog.Options) { o.MaxSizeMB = 0 },
			wantErr: "MaxSizeMB",
		},
		{
			name:    "negative MaxSizeMB",
			mutate:  func(o *buildlog.Options) { o.MaxSizeMB = -1 },
			wantErr: "MaxSizeMB",
		},
		{
			name:    "absurdly large MaxSizeMB",
			mutate:  func(o *buildlog.Options) { o.MaxSizeMB = 100_001 },
			wantErr: "MaxSizeMB",
		},
		{
			name:    "zero MaxAgeDays",
			mutate:  func(o *buildlog.Options) { o.MaxAgeDays = 0 },
			wantErr: "MaxAgeDays",
		},
		{
			name:    "negative MaxAgeDays",
			mutate:  func(o *buildlog.Options) { o.MaxAgeDays = -5 },
			wantErr: "MaxAgeDays",
		},
		{
			name:    "zero MaxBackups",
			mutate:  func(o *buildlog.Options) { o.MaxBackups = 0 },
			wantErr: "MaxBackups",
		},
		{
			name:    "negative MaxBackups",
			mutate:  func(o *buildlog.Options) { o.MaxBackups = -3 },
			wantErr: "MaxBackups",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := base
			tc.mutate(&opts)
			_, err := buildlog.New(opts)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErr)
		})
	}
}

// TestNew_DisabledLogger_IgnoresInvalidNumericConfig verifies that a noop
// logger is returned without error even when numeric bounds would otherwise
// fail — because validation is skipped when Enabled=false.
func TestNew_DisabledLogger_IgnoresInvalidNumericConfig(t *testing.T) {
	opts := buildlog.Options{
		Enabled:    false,
		MaxSizeMB:  -999, // would fail if enabled
		MaxAgeDays: 0,
		MaxBackups: -1,
	}
	logger, err := buildlog.New(opts)
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Noop logger must still satisfy the full interface.
	require.NoError(t, logger.Open("noop-novalidate"))
	_, _ = io.WriteString(logger.Writer("ws"), "x\n")
	require.NoError(t, logger.Close())
}
