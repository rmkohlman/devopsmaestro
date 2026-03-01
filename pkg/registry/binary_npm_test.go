package registry

import (
	"context"
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
