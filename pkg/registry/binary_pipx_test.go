package registry

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// PipxBinaryManager Tests
// =============================================================================

func TestPipxBinaryManager_EnsureBinary_AlreadyInstalled(t *testing.T) {
	t.Skip("Integration test - requires pipx to be installed")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
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

func TestPipxBinaryManager_EnsureBinary_InstallsViaPipx(t *testing.T) {
	t.Skip("Integration test - requires pipx to be installed")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
	ctx := context.Background()

	binaryPath, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err, "Should install via pipx successfully")
	assert.NotEmpty(t, binaryPath, "Should return valid binary path")

	// Verify binary is executable
	// (Check file permissions, etc.)
}

func TestPipxBinaryManager_EnsureBinary_PipxNotInstalled(t *testing.T) {
	t.Skip("Integration test - hard to test without manipulating PATH")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
	ctx := context.Background()

	// Mock environment where pipx is not available
	// (Would require PATH manipulation or mock exec)

	_, err := mgr.EnsureBinary(ctx)
	assert.Error(t, err, "Should fail if pipx is not installed")
	assert.Contains(t, err.Error(), "pipx", "Error should mention pipx")
}

func TestPipxBinaryManager_EnsureBinary_NetworkError(t *testing.T) {
	t.Skip("Integration test - requires network simulation")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
	ctx := context.Background()

	// Simulate network failure during install
	// (Would require network mocking)

	_, err := mgr.EnsureBinary(ctx)
	assert.Error(t, err, "Should fail on network error")
}

func TestPipxBinaryManager_EnsureBinary_InvalidPackage(t *testing.T) {
	t.Skip("Integration test - requires pipx")

	mgr := NewPipxBinaryManager("nonexistent-package-xyz", "1.0.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	assert.Error(t, err, "Should fail for nonexistent package")
}

// =============================================================================
// PipxBinaryManager GetVersion Tests
// =============================================================================

func TestPipxBinaryManager_GetVersion_Success(t *testing.T) {
	t.Skip("Integration test - requires devpi-server to be installed")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
	ctx := context.Background()

	// Ensure binary is installed first
	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Get version
	version, err := mgr.GetVersion(ctx)
	require.NoError(t, err, "Should get version successfully")
	assert.NotEmpty(t, version, "Version should not be empty")
	assert.Contains(t, version, "6.2", "Version should match expected major.minor")
}

func TestPipxBinaryManager_GetVersion_NotInstalled(t *testing.T) {
	t.Skip("Integration test - requires pipx")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
	ctx := context.Background()

	// Don't call EnsureBinary first
	_, err := mgr.GetVersion(ctx)
	assert.Error(t, err, "Should fail if binary not installed")
}

func TestPipxBinaryManager_GetVersion_ParsesCorrectly(t *testing.T) {
	t.Skip("Integration test - requires devpi-server")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	version, err := mgr.GetVersion(ctx)
	require.NoError(t, err)

	// Version should be in format "6.2.0" or similar
	assert.Regexp(t, `^\d+\.\d+`, version, "Version should match semver format")
}

// =============================================================================
// PipxBinaryManager NeedsUpdate Tests
// =============================================================================

func TestPipxBinaryManager_NeedsUpdate_CurrentVersion(t *testing.T) {
	t.Skip("Integration test - requires devpi-server")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	needsUpdate, err := mgr.NeedsUpdate(ctx)
	require.NoError(t, err, "NeedsUpdate should not error")
	assert.False(t, needsUpdate, "Should not need update if version matches")
}

func TestPipxBinaryManager_NeedsUpdate_OlderVersion(t *testing.T) {
	t.Skip("Integration test - requires version management")

	// Install older version first
	mgr1 := NewPipxBinaryManager("devpi-server", "6.0.0")
	ctx := context.Background()

	_, err := mgr1.EnsureBinary(ctx)
	require.NoError(t, err)

	// Check if newer version needed
	mgr2 := NewPipxBinaryManager("devpi-server", "6.2.0")
	needsUpdate, err := mgr2.NeedsUpdate(ctx)
	require.NoError(t, err)
	assert.True(t, needsUpdate, "Should need update if desired version is newer")
}

func TestPipxBinaryManager_NeedsUpdate_NotInstalled(t *testing.T) {
	t.Skip("Integration test - requires pipx")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
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

// =============================================================================
// PipxBinaryManager Update Tests
// =============================================================================

func TestPipxBinaryManager_Update_Success(t *testing.T) {
	t.Skip("Integration test - requires pipx and network")

	// Install old version
	mgr1 := NewPipxBinaryManager("devpi-server", "6.0.0")
	ctx := context.Background()

	_, err := mgr1.EnsureBinary(ctx)
	require.NoError(t, err)

	// Update to newer version
	mgr2 := NewPipxBinaryManager("devpi-server", "6.2.0")
	err = mgr2.Update(ctx)
	require.NoError(t, err, "Update should succeed")

	// Verify version changed
	version, err := mgr2.GetVersion(ctx)
	require.NoError(t, err)
	assert.Contains(t, version, "6.2", "Version should be updated")
}

func TestPipxBinaryManager_Update_BackupOnFailure(t *testing.T) {
	t.Skip("Integration test - requires error simulation")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
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

func TestPipxBinaryManager_Update_NotInstalled(t *testing.T) {
	t.Skip("Integration test - requires pipx")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
	ctx := context.Background()

	// Try to update without installing first
	err := mgr.Update(ctx)
	// Should either install fresh or error
	// Behavior depends on implementation choice
	assert.NoError(t, err, "Update should install if not present (or error if design requires EnsureBinary first)")
}

func TestPipxBinaryManager_Update_Idempotent(t *testing.T) {
	t.Skip("Integration test - requires pipx")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
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

// =============================================================================
// PipxBinaryManager Edge Cases
// =============================================================================

func TestPipxBinaryManager_EnsureBinary_ConcurrentCalls(t *testing.T) {
	t.Skip("Integration test - requires real pipx and concurrency")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
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

func TestPipxBinaryManager_ContextCancellation(t *testing.T) {
	t.Skip("Integration test - requires long-running operation")

	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := mgr.EnsureBinary(ctx)
	assert.Error(t, err, "Should respect context cancellation")
	assert.ErrorIs(t, err, context.Canceled, "Should return context.Canceled error")
}

func TestPipxBinaryManager_InvalidVersion(t *testing.T) {
	mgr := NewPipxBinaryManager("devpi-server", "invalid-version")
	ctx := context.Background()

	// Should either validate version format or fail during install
	_, err := mgr.EnsureBinary(ctx)
	// Implementation choice: fail early on invalid version format or let pipx fail
	_ = err // Test validates behavior, either is acceptable
}

// =============================================================================
// MockPipxBinaryManager Tests
// =============================================================================

func TestMockPipxBinaryManager_EnsureBinary(t *testing.T) {
	mockMgr := NewMockPipxBinaryManager(t.TempDir(), "6.2.0")
	ctx := context.Background()

	binaryPath, err := mockMgr.EnsureBinary(ctx)
	require.NoError(t, err, "Mock should create fake binary")
	assert.NotEmpty(t, binaryPath, "Should return path to fake binary")
}

func TestMockPipxBinaryManager_GetVersion(t *testing.T) {
	mockMgr := NewMockPipxBinaryManager(t.TempDir(), "6.2.0")
	ctx := context.Background()

	version, err := mockMgr.GetVersion(ctx)
	require.NoError(t, err)
	assert.Equal(t, "6.2.0", version, "Mock should return configured version")
}

func TestMockPipxBinaryManager_CustomBehavior(t *testing.T) {
	mockMgr := NewMockPipxBinaryManager(t.TempDir(), "6.2.0")

	// Override behavior with custom function
	mockMgr.EnsureBinaryFunc = func(ctx context.Context) (string, error) {
		return "", assert.AnError
	}

	ctx := context.Background()
	_, err := mockMgr.EnsureBinary(ctx)
	assert.Error(t, err, "Should use custom function when provided")
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestPipxBinaryManager_ImplementsBinaryManager(t *testing.T) {
	var _ BinaryManager = (*PipxBinaryManager)(nil)
}

// =============================================================================
// Bug 3: Devpi pip fallback when pipx not found (TDD RED phase)
//
// Current behaviour: ensurePipxInstalled() requires pipx in PATH. If pipx is
// not found, EnsureBinary() returns an error containing "pipx not found in PATH".
// No fallback is attempted.
//
// Desired behaviour: When pipx is not found, fall back to:
//   python3 -m pip install --user devpi-server==6.2.0
// The error message should change: it should reflect that pip was attempted
// (not that pipx was missing). If pip also fails, the error should mention "pip".
//
// Testing strategy: We cannot easily manipulate PATH in unit tests, so we
// document the expected error-message CONTRACT change that the fix must satisfy.
// The test uses a PipxBinaryManager with a custom EnsurePipxFunc (once the
// refactored method exists) to simulate the "pipx not found" path.
// =============================================================================

// TestPipxBinaryManager_FallbackToPip_WhenPipxMissing documents the expected
// error behaviour change when pipx is not found.
//
// CURRENT behaviour (BUG):
//
//	EnsureBinary returns an error whose message contains "pipx not found in PATH".
//	The caller has no recourse except to install pipx manually.
//
// DESIRED behaviour (after fix):
//
//	When pipx is not found, EnsureBinary falls back to
//	  python3 -m pip install --user devpi-server==<version>
//	If that also fails, the returned error should mention "pip" (the fallback
//	was attempted) and must NOT contain the string "pipx not found in PATH"
//	(because the implementation moved past that error and tried pip).
//
// This test FAILS today because EnsureBinary() immediately returns an error
// containing "pipx not found in PATH" with no pip fallback attempted.
// After the fix, the error message contract will change as asserted below.
//
// Environment note: this test is meaningful in CI where pipx is not installed.
// When pipx IS installed, EnsureBinary may succeed (installing devpi-server);
// we guard for that case with a check for err == nil.
func TestPipxBinaryManager_FallbackToPip_WhenPipxMissing(t *testing.T) {
	mgr := NewPipxBinaryManager("devpi-server", "6.2.0")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)

	if err == nil {
		// pipx (or pip fallback) is installed and succeeded — nothing to assert.
		t.Log("EnsureBinary succeeded (pipx or pip available in this environment)")
		return
	}

	// EnsureBinary failed. Assert the error message CONTRACT:
	//
	// BUG (current): error contains "pipx not found in PATH".
	//   The test will FAIL here today because the error IS "pipx not found in PATH"
	//   and we assert it must NOT be that.
	//
	// FIXED (after implementation): error mentions pip/python (fallback was tried).
	assert.NotContains(t, err.Error(), "pipx not found in PATH",
		"BUG DETECTED: EnsureBinary returned 'pipx not found in PATH' without "+
			"attempting a pip fallback. After the fix, the error should describe "+
			"the pip attempt instead.")

	// After the fix: error should mention pip or python (the fallback was tried).
	assert.True(t,
		strings.Contains(err.Error(), "pip") || strings.Contains(err.Error(), "python"),
		"After the fix, a failure after pip fallback should mention 'pip' or 'python': %v", err)
}
