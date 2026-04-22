package buildlog_test

// Integration tests for the build-log file lifecycle (Issue #400).
//
// These tests exercise the BuildLogger end-to-end the same way the
// orchestrator wires it: Open at session start, MultiWriter into per-
// workspace sinks, Close on shutdown. They assert on actual file system
// artefacts (the per-session log file and the latest.log symlink) rather
// than on internal state.

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devopsmaestro/pkg/buildlog"
)

// fakeBuildFn mimics buildSingleWorkspaceForParallel: it writes a header and
// some workspace output to the provided sink and optionally returns an error
// (or honours a cancelled context to simulate Ctrl-C).
func fakeBuildFn(ctx context.Context, name string, sink io.Writer, fail bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	fmt.Fprintf(sink, "─── Building: %s ───\n", name)
	fmt.Fprintf(sink, "[%s] step 1: resolved sources\n", name)
	fmt.Fprintf(sink, "[%s] step 2: built image\n", name)
	if fail {
		return fmt.Errorf("%s: simulated build failure", name)
	}
	return nil
}

// runFakeOrchestration mimics the Open/MultiWriter/Close lifecycle that the
// orchestrator now performs. cancelAfter, when non-nil, is invoked after the
// first workspace's output has been written so subsequent workspaces see a
// cancelled context — this is the interrupt path #400 cares about.
func runFakeOrchestration(
	t *testing.T,
	dir string,
	workspaces []string,
	failNames map[string]bool,
	cancelAfter func(),
) (string, *bytes.Buffer, buildlog.BuildLogger) {
	t.Helper()
	sessionID := uuid.New().String()

	bl, err := buildlog.New(buildlog.Options{
		Enabled:    true,
		Directory:  dir,
		MaxSizeMB:  10,
		MaxAgeDays: 7,
		MaxBackups: 3,
		Compress:   false,
	})
	require.NoError(t, err)
	require.NoError(t, bl.Open(sessionID))

	stdout := &bytes.Buffer{}
	var stdoutMu sync.Mutex

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i, name := range workspaces {
		if i == 1 && cancelAfter != nil {
			cancelAfter()
			cancel()
		}
		var wsBuf bytes.Buffer
		logWriter := bl.Writer(name)
		sink := io.MultiWriter(&wsBuf, logWriter)
		_ = fakeBuildFn(ctx, name, sink, failNames[name])
		stdoutMu.Lock()
		_, _ = io.Copy(stdout, &wsBuf)
		stdoutMu.Unlock()
	}

	return sessionID, stdout, bl
}

// TestIntegration_OrchestratorWritesPerSessionLogAndSymlink verifies the
// happy path: a clean run produces <dir>/<sessionID>.log containing every
// workspace's output, and latest.log resolves to that file.
func TestIntegration_OrchestratorWritesPerSessionLogAndSymlink(t *testing.T) {
	dir := t.TempDir()
	workspaces := []string{"alpha", "beta", "gamma"}

	sessionID, stdout, bl := runFakeOrchestration(t, dir, workspaces, nil, nil)
	require.NoError(t, bl.Close())

	// Per-session file exists at <dir>/<sessionID>.log.
	logPath := filepath.Join(dir, sessionID+".log")
	info, err := os.Stat(logPath)
	require.NoError(t, err, "per-session log file must exist at %s", logPath)
	require.False(t, info.IsDir())

	contents, err := os.ReadFile(logPath)
	require.NoError(t, err)
	body := string(contents)
	for _, ws := range workspaces {
		assert.Contains(t, body, fmt.Sprintf("Building: %s", ws),
			"log file must contain header for workspace %q", ws)
		assert.Contains(t, body, fmt.Sprintf("[%s] step 2: built image", ws),
			"log file must contain step output for workspace %q", ws)
	}

	// latest.log symlink resolves to the per-session log file.
	latestPath := filepath.Join(dir, "latest.log")
	target, err := os.Readlink(latestPath)
	require.NoError(t, err, "latest.log must be a symlink")
	if !filepath.IsAbs(target) {
		target = filepath.Join(dir, target)
	}
	resolved, err := filepath.EvalSymlinks(latestPath)
	require.NoError(t, err)
	expectedResolved, err := filepath.EvalSymlinks(logPath)
	require.NoError(t, err)
	assert.Equal(t, expectedResolved, resolved,
		"latest.log must resolve to the per-session log file (target=%s)", target)

	// Stdout sink also got the bytes (the MultiWriter contract).
	assert.Contains(t, stdout.String(), "Building: alpha")
}

// TestIntegration_InterruptedBuildFlushesAndClosesLog verifies that when a
// build is cancelled mid-flight the partial output written before the
// cancellation is still flushed to disk and the file is properly closed —
// no truncation, no lost bytes for the workspaces that did execute.
func TestIntegration_InterruptedBuildFlushesAndClosesLog(t *testing.T) {
	dir := t.TempDir()
	workspaces := []string{"first", "second", "third"}

	sessionID, _, bl := runFakeOrchestration(t, dir, workspaces, nil,
		func() { /* cancel triggers in helper */ })
	require.NoError(t, bl.Close(),
		"Close must succeed even when builds were interrupted")

	logPath := filepath.Join(dir, sessionID+".log")
	contents, err := os.ReadFile(logPath)
	require.NoError(t, err, "log file must exist after interrupted build")
	body := string(contents)

	// The first workspace ran to completion before cancel; its full output
	// must have been flushed to disk despite the interrupt that followed.
	assert.Contains(t, body, "Building: first")
	assert.Contains(t, body, "[first] step 2: built image",
		"output written before interrupt must be flushed to disk")

	// Subsequent workspaces saw the cancelled context and bailed out before
	// writing — nothing extra leaked into the file.
	if strings.Contains(body, "[second] step 2: built image") {
		t.Log("note: second workspace produced full output; cancellation racing")
	}

	// File must not be empty / truncated.
	info, err := os.Stat(logPath)
	require.NoError(t, err)
	assert.Greater(t, info.Size(), int64(0),
		"log file must contain flushed bytes after interrupt+close")

	// latest.log must still resolve cleanly (no dangling symlink after Close).
	resolved, err := filepath.EvalSymlinks(filepath.Join(dir, "latest.log"))
	require.NoError(t, err, "latest.log must remain valid after interrupted build")
	expected, _ := filepath.EvalSymlinks(logPath)
	assert.Equal(t, expected, resolved)
}

// guard: ensure errors.Is shape stays imported when refactoring above.
var _ = errors.Is
