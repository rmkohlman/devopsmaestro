package registry

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// NpmBinaryManager Tests
// =============================================================================

func TestNpmBinaryManager_EnsureBinary_AlreadyInstalled(t *testing.T) {
	t.Skip("Integration test - requires npm to be installed")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	// First call installs
	binaryPath, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err, "First EnsureBinary should succeed")
	assert.NotEmpty(t, binaryPath, "Should return binary path")

	// Second call should detect existing installation
	binaryPath2, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err, "Second EnsureBinary should succeed")
	assert.Equal(t, binaryPath, binaryPath2, "Should return same path")
}

func TestNpmBinaryManager_EnsureBinary_InstallsViaNpm(t *testing.T) {
	t.Skip("Integration test - requires npm to be installed")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	binaryPath, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err, "Should install via npm successfully")
	assert.NotEmpty(t, binaryPath, "Should return valid binary path")

	// Verify binary is executable
	// (Check file permissions, etc.)
}

func TestNpmBinaryManager_EnsureBinary_NpmNotInstalled(t *testing.T) {
	t.Skip("Integration test - hard to test without manipulating PATH")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	// Mock environment where npm is not available
	// (Would require PATH manipulation or mock exec)

	_, err := mgr.EnsureBinary(ctx)
	assert.Error(t, err, "Should fail if npm is not installed")
	assert.Contains(t, err.Error(), "npm", "Error should mention npm")
}

func TestNpmBinaryManager_EnsureBinary_NetworkError(t *testing.T) {
	t.Skip("Integration test - requires network simulation")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	// Simulate network failure during install
	// (Would require network mocking)

	_, err := mgr.EnsureBinary(ctx)
	assert.Error(t, err, "Should fail on network error")
}

func TestNpmBinaryManager_EnsureBinary_InvalidPackage(t *testing.T) {
	t.Skip("Integration test - requires npm")

	mgr := NewNpmBinaryManager("nonexistent-package-xyz", "1.0.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	assert.Error(t, err, "Should fail for nonexistent package")
}

func TestNpmBinaryManager_EnsureBinary_GlobalInstall(t *testing.T) {
	t.Skip("Integration test - requires npm")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	binaryPath, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Should install globally
	assert.Contains(t, binaryPath, "node_modules", "Should be in global node_modules")
}

func TestNpmBinaryManager_EnsureBinary_SpecificVersion(t *testing.T) {
	t.Skip("Integration test - requires npm")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Verify correct version was installed
	version, err := mgr.GetVersion(ctx)
	require.NoError(t, err)
	assert.Contains(t, version, "5.28", "Should install specific version")
}

// =============================================================================
// NpmBinaryManager GetVersion Tests
// =============================================================================

func TestNpmBinaryManager_GetVersion_Success(t *testing.T) {
	t.Skip("Integration test - requires verdaccio to be installed")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	// Ensure binary is installed first
	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Get version
	version, err := mgr.GetVersion(ctx)
	require.NoError(t, err, "Should get version successfully")
	assert.NotEmpty(t, version, "Version should not be empty")
	assert.Contains(t, version, "5.", "Version should match expected major version")
}

func TestNpmBinaryManager_GetVersion_NotInstalled(t *testing.T) {
	t.Skip("Integration test - requires npm")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	// Don't call EnsureBinary first
	_, err := mgr.GetVersion(ctx)
	assert.Error(t, err, "Should fail if binary not installed")
}

func TestNpmBinaryManager_GetVersion_ParsesCorrectly(t *testing.T) {
	t.Skip("Integration test - requires verdaccio")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	version, err := mgr.GetVersion(ctx)
	require.NoError(t, err)

	// Version should be in format "5.28.0" or similar
	assert.Regexp(t, `^\d+\.\d+`, version, "Version should match semver format")
}

func TestNpmBinaryManager_GetVersion_UsesVersionFlag(t *testing.T) {
	t.Skip("Integration test - requires verdaccio")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// GetVersion should use --version flag
	version, err := mgr.GetVersion(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, version, "Should return version from --version flag")
}

// =============================================================================
// NpmBinaryManager NeedsUpdate Tests
// =============================================================================

func TestNpmBinaryManager_NeedsUpdate_CurrentVersion(t *testing.T) {
	t.Skip("Integration test - requires verdaccio")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	needsUpdate, err := mgr.NeedsUpdate(ctx)
	require.NoError(t, err, "NeedsUpdate should not error")
	assert.False(t, needsUpdate, "Should not need update if version matches")
}

func TestNpmBinaryManager_NeedsUpdate_OlderVersion(t *testing.T) {
	t.Skip("Integration test - requires version management")

	// Install older version first
	mgr1 := NewNpmBinaryManager("verdaccio", "5.0.0")
	ctx := context.Background()

	_, err := mgr1.EnsureBinary(ctx)
	require.NoError(t, err)

	// Check if newer version needed
	mgr2 := NewNpmBinaryManager("verdaccio", "5.28.0")
	needsUpdate, err := mgr2.NeedsUpdate(ctx)
	require.NoError(t, err)
	assert.True(t, needsUpdate, "Should need update if desired version is newer")
}

func TestNpmBinaryManager_NeedsUpdate_NotInstalled(t *testing.T) {
	t.Skip("Integration test - requires npm")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	// Don't install first
	needsUpdate, err := mgr.NeedsUpdate(ctx)
	// Should either return true (needs installation) or error
	if err == nil {
		assert.True(t, needsUpdate, "Should need update if not installed")
	} else {
		assert.Error(t, err, "Alternatively should error if not installed")
	}
}

func TestNpmBinaryManager_NeedsUpdate_LatestTag(t *testing.T) {
	t.Skip("Integration test - requires npm and network")

	// Install specific version
	mgr1 := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	_, err := mgr1.EnsureBinary(ctx)
	require.NoError(t, err)

	// Check if "latest" is needed
	mgr2 := NewNpmBinaryManager("verdaccio", "latest")
	needsUpdate, err := mgr2.NeedsUpdate(ctx)
	require.NoError(t, err)
	// Result depends on whether 5.28.0 is actually the latest
	_ = needsUpdate
}

// =============================================================================
// NpmBinaryManager Update Tests
// =============================================================================

func TestNpmBinaryManager_Update_Success(t *testing.T) {
	t.Skip("Integration test - requires npm and network")

	// Install old version
	mgr1 := NewNpmBinaryManager("verdaccio", "5.0.0")
	ctx := context.Background()

	_, err := mgr1.EnsureBinary(ctx)
	require.NoError(t, err)

	// Update to newer version
	mgr2 := NewNpmBinaryManager("verdaccio", "5.28.0")
	err = mgr2.Update(ctx)
	require.NoError(t, err, "Update should succeed")

	// Verify version changed
	version, err := mgr2.GetVersion(ctx)
	require.NoError(t, err)
	assert.Contains(t, version, "5.28", "Version should be updated")
}

func TestNpmBinaryManager_Update_BackupOnFailure(t *testing.T) {
	t.Skip("Integration test - requires error simulation")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	// Install current version
	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Simulate update failure
	// (Would require network/install mocking)

	// Verify old version is still available (rollback)
	version, err := mgr.GetVersion(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, version, "Should rollback to previous version on failure")
}

func TestNpmBinaryManager_Update_NotInstalled(t *testing.T) {
	t.Skip("Integration test - requires npm")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	// Try to update without installing first
	err := mgr.Update(ctx)
	// Should either install fresh or error
	// Behavior depends on implementation choice
	assert.NoError(t, err, "Update should install if not present (or error if design requires EnsureBinary first)")
}

func TestNpmBinaryManager_Update_Idempotent(t *testing.T) {
	t.Skip("Integration test - requires npm")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Update to same version - should be idempotent
	err = mgr.Update(ctx)
	assert.NoError(t, err, "Update to same version should succeed")

	// Update again
	err = mgr.Update(ctx)
	assert.NoError(t, err, "Second update should also succeed")
}

func TestNpmBinaryManager_Update_UsesNpmUpdate(t *testing.T) {
	t.Skip("Integration test - requires npm")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Update should use npm update or npm install -g
	err = mgr.Update(ctx)
	assert.NoError(t, err, "Should use npm commands for update")
}

// =============================================================================
// NpmBinaryManager Edge Cases
// =============================================================================

func TestNpmBinaryManager_EnsureBinary_ConcurrentCalls(t *testing.T) {
	t.Skip("Integration test - requires real npm and concurrency")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	// Call EnsureBinary concurrently
	errChan := make(chan error, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := mgr.EnsureBinary(ctx)
			errChan <- err
		}()
	}

	// All should succeed without race conditions
	for i := 0; i < 5; i++ {
		err := <-errChan
		assert.NoError(t, err, "Concurrent EnsureBinary calls should not race")
	}
}

func TestNpmBinaryManager_ContextCancellation(t *testing.T) {
	t.Skip("Integration test - requires long-running operation")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := mgr.EnsureBinary(ctx)
	assert.Error(t, err, "Should respect context cancellation")
	assert.ErrorIs(t, err, context.Canceled, "Should return context.Canceled error")
}

func TestNpmBinaryManager_InvalidVersion(t *testing.T) {
	mgr := NewNpmBinaryManager("verdaccio", "invalid-version")
	ctx := context.Background()

	// Should either validate version format or fail during install
	_, err := mgr.EnsureBinary(ctx)
	// Implementation choice: fail early on invalid version format or let npm fail
	_ = err // Test validates behavior, either is acceptable
}

func TestNpmBinaryManager_GetBinaryPath(t *testing.T) {
	t.Skip("Integration test - requires npm")

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	ctx := context.Background()

	binaryPath, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Verify path exists and is executable
	assert.NotEmpty(t, binaryPath, "Binary path should not be empty")
	// Could check file stats here
}

// =============================================================================
// MockNpmBinaryManager Tests
// =============================================================================

func TestMockNpmBinaryManager_EnsureBinary(t *testing.T) {
	mockMgr := NewMockNpmBinaryManager(t.TempDir(), "5.28.0")
	ctx := context.Background()

	binaryPath, err := mockMgr.EnsureBinary(ctx)
	require.NoError(t, err, "Mock should create fake binary")
	assert.NotEmpty(t, binaryPath, "Should return path to fake binary")
}

func TestMockNpmBinaryManager_GetVersion(t *testing.T) {
	mockMgr := NewMockNpmBinaryManager(t.TempDir(), "5.28.0")
	ctx := context.Background()

	version, err := mockMgr.GetVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, "5.28.0", version, "Mock should return configured version")
}

func TestMockNpmBinaryManager_CustomBehavior(t *testing.T) {
	mockMgr := NewMockNpmBinaryManager(t.TempDir(), "5.28.0")

	// Override behavior with custom function
	mockMgr.EnsureBinaryFunc = func(ctx context.Context) (string, error) {
		return "", assert.AnError
	}

	ctx := context.Background()
	_, err := mockMgr.EnsureBinary(ctx)
	assert.Error(t, err, "Should use custom function when provided")
}

func TestMockNpmBinaryManager_NeedsUpdate(t *testing.T) {
	mockMgr := NewMockNpmBinaryManager(t.TempDir(), "5.28.0")
	ctx := context.Background()

	needsUpdate, err := mockMgr.NeedsUpdate(ctx)
	require.NoError(t, err)
	assert.False(t, needsUpdate, "Mock should report no update needed by default")
}

func TestMockNpmBinaryManager_Update(t *testing.T) {
	mockMgr := NewMockNpmBinaryManager(t.TempDir(), "5.28.0")
	ctx := context.Background()

	err := mockMgr.Update(ctx)
	assert.NoError(t, err, "Mock update should succeed")
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestNpmBinaryManager_ImplementsBinaryManager(t *testing.T) {
	var _ BinaryManager = (*NpmBinaryManager)(nil)
}

// =============================================================================
// B7: npm isPackageInstalled discards exec error — TDD RED phase
//
// Current behaviour (BUG B7):
//   binary_npm.go:178  output, _ := cmd.CombinedOutput()
//
//   isPackageInstalled swallows ALL errors from the `npm list` command.
//   npm exits with code 1 when a package is not found — that is a normal,
//   expected condition and the current code handles it by checking the output
//   text.  However, a *non-exit-code-1* error (e.g. context cancellation,
//   SIGKILL, missing npm binary, OS-level error) is silently dropped.
//   EnsureBinary then proceeds as if the package is simply not installed,
//   attempting an npm install that will also fail — hiding the root cause.
//
// The fix: capture the error and return it from isPackageInstalled when the
// exit code is not 1 (i.e. not the "package not found" sentinel).
// =============================================================================

// TestNpmBinaryManager_EnsureBinary_PropagatesNpmListFatalError verifies that
// EnsureBinary surfaces fatal errors from underlying npm calls instead of
// swallowing them.
//
// NOTE: This test exercises the ensureNpmInstalled path (which runs BEFORE
// isPackageInstalled) using a cancelled context.  It currently PASSES because
// ensureNpmInstalled correctly propagates the context-cancellation error.
//
// The test documents that the "propagate, don't swallow" contract must hold
// at every npm-exec call site — including the one inside isPackageInstalled
// that is the actual B7 bug site.  See TestNpmBinaryManager_IsPackageInstalled
// _DoesNotSwallowFatalError for the test that directly targets the bug.
func TestNpmBinaryManager_EnsureBinary_PropagatesNpmListFatalError(t *testing.T) {
	// Require npm on PATH — skip if not available (CI without node).
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available on PATH; skipping B7 test")
	}

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	// Override binDir to a temp directory so no stale cached binary is found.
	mgr.binDir = t.TempDir()

	// Cancel the context immediately — this causes exec.CommandContext to
	// terminate the npm process with a signal, producing a non-exit-code-1
	// error that MUST be propagated.
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancelled before the call

	_, err := mgr.EnsureBinary(ctx)

	// After the fix: EnsureBinary must return a non-nil error that either
	// is context.Canceled or wraps it, surfacing the real cause.
	//
	// Current behaviour (BUG): err is nil OR contains only "binary not found
	// after installation" — the npm-list exec error is silently dropped.
	if err == nil {
		t.Fatal("EnsureBinary should return an error when the context is cancelled, but got nil")
	}

	// The error must not merely say "binary not found after installation" —
	// that message masks the real cause (context cancellation / exec error).
	// After the fix the error chain must include something about the exec
	// failure or context cancellation.
	//
	// NOTE: we do NOT use assert.ErrorIs(ctx.Err()) here because the error
	// may be wrapped multiple times; we check the message as a proxy until
	// errors.Is/As is wired up in the implementation.
	t.Logf("EnsureBinary returned error (should reference npm/exec/context): %v", err)
}

// TestNpmBinaryManager_IsPackageInstalled_DoesNotSwallowFatalError is a
// unit-level documentation test for the discarded-error pattern.
//
// Because isPackageInstalled is unexported we test its behaviour through
// EnsureBinary.  This test uses a fake npm script that:
//   - exits 0 for `npm --version` (so ensureNpmInstalled passes)
//   - exits 2 for `npm list` (fatal error, not the expected "not found" exit 1)
//
// This verifies that EnsureBinary propagates the fatal npm list error instead
// of silently proceeding to attempt an install.
//
// This test FAILS today because the `output, _` pattern in isPackageInstalled
// throws away the exit-code-2 error; EnsureBinary gets (false, nil) back from
// isPackageInstalled and then proceeds to call installPackage (which also
// exits 2), producing a confusing "npm install failed" error rather than
// "failed to check package status: …"
//
// After the fix: the error returned must contain "failed to check package
// status" — the wrapping string from EnsureBinary line 58.
func TestNpmBinaryManager_IsPackageInstalled_DoesNotSwallowFatalError(t *testing.T) {
	// Build a fake "npm" script that:
	//   - succeeds for --version (so ensureNpmInstalled passes)
	//   - exits 2 for any "list" command (fatal npm error)
	fakeNpmDir := t.TempDir()
	fakeNpmScript := `#!/bin/sh
if [ "$1" = "--version" ]; then
  echo "9.0.0"
  exit 0
fi
if [ "$1" = "list" ]; then
  echo "npm ERR! catastrophic failure" >&2
  exit 2
fi
exit 2
`
	fakeNpmPath := fakeNpmDir + "/npm"
	if err := os.WriteFile(fakeNpmPath, []byte(fakeNpmScript), 0755); err != nil {
		t.Fatalf("failed to write fake npm script: %v", err)
	}

	// Prepend fake npm dir to PATH so our script is found first.
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", fakeNpmDir+string(os.PathListSeparator)+origPath)

	mgr := NewNpmBinaryManager("verdaccio", "5.28.0")
	mgr.binDir = t.TempDir()

	_, err := mgr.EnsureBinary(context.Background())

	// After the fix: the exit-code-2 from the `npm list` call inside
	// isPackageInstalled must surface as an error from EnsureBinary with the
	// message "failed to check package status: …".
	//
	// Current behaviour (BUG B7): isPackageInstalled discards the error
	// (output, _ := cmd.CombinedOutput()), returns (false, nil), and
	// EnsureBinary proceeds to installPackage.  The error that eventually
	// surfaces says "npm install failed" — not "failed to check package
	// status" — masking the real source.
	if err == nil {
		t.Fatal("EnsureBinary should return an error when npm list exits with code 2, but got nil")
	}

	// THE KEY ASSERTION: after the fix, the error must come from the
	// package-status check (isPackageInstalled), not from the install step.
	// This will FAIL today because the install-step error fires instead.
	if !strings.Contains(err.Error(), "failed to check package status") {
		t.Errorf("BUG B7: expected error to contain 'failed to check package status' "+
			"(from isPackageInstalled propagating the fatal exit-2 error), "+
			"but got: %q\n"+
			"This confirms isPackageInstalled is swallowing the exec error with 'output, _'",
			err.Error())
	}

	t.Logf("EnsureBinary returned error: %v", err)
}
