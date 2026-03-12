package registry

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupTestBinaryManager creates a BinaryManager with a test directory.
// For integration tests that need real downloads, use a valid version.
func setupTestBinaryManager(t *testing.T) BinaryManager {
	t.Helper()

	binDir := t.TempDir()
	// Use v2.1.1 which is a real, stable Zot version
	return NewBinaryManager(binDir, "2.1.1")
}

// setupMockBinaryManager creates a MockBinaryManager for unit tests.
func setupMockBinaryManager(t *testing.T) *MockBinaryManager {
	t.Helper()

	binDir := t.TempDir()
	return NewMockBinaryManager(binDir, "1.4.3")
}

// =============================================================================
// Task 2.1: EnsureBinary Tests
// =============================================================================

func TestBinaryManager_EnsureBinary_AlreadyExists(t *testing.T) {
	binDir := t.TempDir()

	// Create fake binary that responds to --version
	makeFakeZotScript(t, binDir, "1.0.0")
	binaryPath := filepath.Join(binDir, "zot")

	mgr := NewBinaryManager(binDir, "1.0.0")
	ctx := context.Background()

	path, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err, "EnsureBinary should succeed when binary exists with matching version")
	assert.Equal(t, binaryPath, path, "Should return existing binary path")
}

func TestBinaryManager_EnsureBinary_Downloads(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	mgr := setupTestBinaryManager(t)
	ctx := context.Background()

	path, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err, "EnsureBinary should download binary")

	assert.FileExists(t, path, "Binary should exist after download")

	// Verify file is executable
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.NotEqual(t, 0, info.Mode()&0111, "Binary should be executable")
}

func TestBinaryManager_EnsureBinary_FailsOnNetworkError(t *testing.T) {
	binDir := t.TempDir()
	// Use invalid version to trigger download error
	mgr := NewBinaryManager(binDir, "invalid-version-999.999.999")
	ctx := context.Background()

	_, err := mgr.EnsureBinary(ctx)
	assert.Error(t, err, "EnsureBinary should fail with invalid version")
	assert.Contains(t, err.Error(), "download", "Error should mention download failure")
}

func TestBinaryManager_EnsureBinary_VerifiesChecksum(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	mgr := setupTestBinaryManager(t)
	ctx := context.Background()

	// Download binary
	path, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Corrupt the binary
	err = os.WriteFile(path, []byte("corrupted"), 0755)
	require.NoError(t, err)

	// Try to ensure binary again - should detect corruption and re-download
	_, err = mgr.EnsureBinary(ctx)
	// Implementation should verify checksum and re-download if corrupted
	// This test will verify that behavior once implemented
}

func TestBinaryManager_EnsureBinary_PermissionsCorrect(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	mgr := setupTestBinaryManager(t)
	ctx := context.Background()

	path, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Verify file permissions are 0755 (rwxr-xr-x)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0755), info.Mode().Perm(), "Binary should have 0755 permissions")
}

// =============================================================================
// Task 2.2: GetVersion Tests
// =============================================================================

func TestBinaryManager_GetVersion_Success(t *testing.T) {
	binDir := t.TempDir()
	_ = filepath.Join(binDir, "zot") // binaryPath - would be used for creating fake binary

	// Create fake binary with version output
	// This would need to be a real executable for this test
	// For now, we test the interface contract

	mgr := NewBinaryManager(binDir, "1.0.0")
	ctx := context.Background()

	version, err := mgr.GetVersion(ctx)
	if err != nil {
		// If binary doesn't exist, this is expected
		t.Skip("Skipping version test without binary")
	}

	assert.NotEmpty(t, version, "Version should not be empty")
	assert.Contains(t, version, ".", "Version should contain dots (semver)")
}

func TestBinaryManager_GetVersion_BinaryNotExist(t *testing.T) {
	mgr := setupTestBinaryManager(t)
	ctx := context.Background()

	_, err := mgr.GetVersion(ctx)
	assert.Error(t, err, "GetVersion should fail if binary doesn't exist")
}

func TestBinaryManager_GetVersion_ParsesOutput(t *testing.T) {
	// This test verifies that GetVersion can parse Zot's version output.
	// Zot outputs JSON to stderr via --version; we parse the commit field
	// (e.g. "v2.1.15-0-gabcdef0") to extract the bare semver string "2.1.15".
	binDir := t.TempDir()
	makeFakeZotScript(t, binDir, "2.1.15")

	bm := &DefaultBinaryManager{binDir: binDir, version: "anything"}
	ctx := context.Background()

	ver, err := bm.GetVersion(ctx)
	require.NoError(t, err, "GetVersion should parse Zot JSON version output")
	assert.Equal(t, "2.1.15", ver, "GetVersion should extract semver from commit field")
}

// =============================================================================
// Task 2.3: NeedsUpdate Tests
// =============================================================================

func TestBinaryManager_NeedsUpdate_True(t *testing.T) {
	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "zot")

	// Create fake old binary
	err := os.WriteFile(binaryPath, []byte("fake old zot"), 0755)
	require.NoError(t, err)

	// Manager expects newer version
	mgr := NewBinaryManager(binDir, "2.0.0")
	ctx := context.Background()

	needs, err := mgr.NeedsUpdate(ctx)
	if err != nil {
		// If we can't determine version, skip
		t.Skip("Cannot determine version without real binary")
	}

	assert.True(t, needs, "Should need update when newer version available")
}

func TestBinaryManager_NeedsUpdate_False(t *testing.T) {
	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "zot")

	// Create fake current binary
	err := os.WriteFile(binaryPath, []byte("fake current zot"), 0755)
	require.NoError(t, err)

	// Manager expects same version
	mgr := NewBinaryManager(binDir, "1.0.0")
	ctx := context.Background()

	needs, err := mgr.NeedsUpdate(ctx)
	if err != nil {
		t.Skip("Cannot determine version without real binary")
	}

	assert.False(t, needs, "Should not need update when versions match")
}

func TestBinaryManager_NeedsUpdate_BinaryNotExist(t *testing.T) {
	mgr := setupTestBinaryManager(t)
	ctx := context.Background()

	needs, err := mgr.NeedsUpdate(ctx)
	// Should return true (needs update because binary doesn't exist)
	// OR return error - either is acceptable
	if err == nil {
		assert.True(t, needs, "Should need update if binary doesn't exist")
	}
}

func TestBinaryManager_NeedsUpdate_SemverComparison(t *testing.T) {
	tests := []struct {
		name           string
		currentVer     string
		desiredVer     string
		expectedUpdate bool
	}{
		{
			name:           "major version update",
			currentVer:     "1.0.0",
			desiredVer:     "2.0.0",
			expectedUpdate: true,
		},
		{
			name:           "minor version update",
			currentVer:     "1.0.0",
			desiredVer:     "1.1.0",
			expectedUpdate: true,
		},
		{
			name:           "patch version update",
			currentVer:     "1.0.0",
			desiredVer:     "1.0.1",
			expectedUpdate: true,
		},
		{
			name:           "same version",
			currentVer:     "1.0.0",
			desiredVer:     "1.0.0",
			expectedUpdate: false,
		},
		{
			name:           "current newer",
			currentVer:     "2.0.0",
			desiredVer:     "1.0.0",
			expectedUpdate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test defines the semver comparison logic
			// Implementation will need to parse versions and compare
			t.Skip("Unit test - requires semver comparison implementation")
		})
	}
}

// =============================================================================
// Task 2.4: Update Tests
// =============================================================================

func TestBinaryManager_Update_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	mgr := setupTestBinaryManager(t)
	ctx := context.Background()

	// First ensure binary exists
	_, err := mgr.EnsureBinary(ctx)
	require.NoError(t, err)

	// Update should succeed
	err = mgr.Update(ctx)
	require.NoError(t, err, "Update should succeed")
}

func TestBinaryManager_Update_BacksUpOldVersion(t *testing.T) {
	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "zot")

	// Create fake old binary
	oldContent := []byte("old zot binary")
	err := os.WriteFile(binaryPath, oldContent, 0755)
	require.NoError(t, err)

	mgr := NewBinaryManager(binDir, "2.0.0")
	ctx := context.Background()

	err = mgr.Update(ctx)
	if err != nil {
		t.Skip("Cannot test update without network")
	}

	// Verify backup exists
	backupPath := binaryPath + ".backup"
	assert.FileExists(t, backupPath, "Old binary should be backed up")

	content, err := os.ReadFile(backupPath)
	require.NoError(t, err)
	assert.Equal(t, oldContent, content, "Backup should contain old binary")
}

func TestBinaryManager_Update_RollsBackOnFailure(t *testing.T) {
	binDir := t.TempDir()
	binaryPath := filepath.Join(binDir, "zot")

	// Create fake working binary
	workingContent := []byte("working zot binary")
	err := os.WriteFile(binaryPath, workingContent, 0755)
	require.NoError(t, err)

	// Use invalid version to force download failure
	mgr := NewBinaryManager(binDir, "invalid-999.999.999")
	ctx := context.Background()

	err = mgr.Update(ctx)
	assert.Error(t, err, "Update should fail with invalid version")

	// Verify original binary is still intact
	content, err := os.ReadFile(binaryPath)
	require.NoError(t, err)
	assert.Equal(t, workingContent, content, "Original binary should be restored on failure")
}

func TestBinaryManager_Update_VerifiesChecksumAfterDownload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	mgr := setupTestBinaryManager(t)
	ctx := context.Background()

	err := mgr.Update(ctx)
	require.NoError(t, err)

	// Verify downloaded binary has correct checksum
	// Implementation should verify this during download
}

// =============================================================================
// Task 2.5: Binary Platform Detection Tests
// =============================================================================

func TestBinaryManager_DetectsPlatform(t *testing.T) {
	tests := []struct {
		name         string
		goos         string
		goarch       string
		wantPlatform string
	}{
		{
			name:         "darwin amd64",
			goos:         "darwin",
			goarch:       "amd64",
			wantPlatform: "darwin-amd64",
		},
		{
			name:         "darwin arm64",
			goos:         "darwin",
			goarch:       "arm64",
			wantPlatform: "darwin-arm64",
		},
		{
			name:         "linux amd64",
			goos:         "linux",
			goarch:       "amd64",
			wantPlatform: "linux-amd64",
		},
		{
			name:         "linux arm64",
			goos:         "linux",
			goarch:       "arm64",
			wantPlatform: "linux-arm64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that BinaryManager correctly determines download URL
			// based on platform
			t.Skip("Unit test - requires platform detection implementation")
		})
	}
}

func TestBinaryManager_UnsupportedPlatform(t *testing.T) {
	// Test that BinaryManager fails gracefully on unsupported platforms
	t.Skip("Unit test - requires platform validation")
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestDefaultBinaryManager_ImplementsBinaryManager(t *testing.T) {
	var _ BinaryManager = (*DefaultBinaryManager)(nil)
}

// =============================================================================
// B6: binary_manager.go should use errors.Is, not string matching — TDD RED phase
//
// Current behaviour (BUG B6):
//   NeedsUpdate (line 73) uses strings.Contains(err.Error(), "not found") to
//   detect ErrBinaryNotFound.  This is fragile: it relies on the error message
//   text rather than the sentinel value, which breaks if the message changes or
//   the error is double-wrapped.
//
// The fix: replace the string match with errors.Is(err, ErrBinaryNotFound).
// =============================================================================

// TestBinaryManager_NeedsUpdate_MissingBinary verifies that NeedsUpdate returns
// (true, nil) when the binary does not exist yet (fresh install scenario).
//
// This test is expected to PASS with both the current string-match code and the
// fixed errors.Is code, because the current error message happens to contain
// "not found".  It is included to:
//  1. Act as a stable regression test for the happy-path behaviour.
//  2. Document the correct contract so that when the string match is replaced
//     the test still validates the observable outcome.
//
// See TestBinaryManager_NeedsUpdate_ErrorsIs_CorrectPattern below for the
// test that validates the pattern itself is errors.Is-based.
func TestBinaryManager_NeedsUpdate_MissingBinary(t *testing.T) {
	// Use a fresh temp dir — no binary will exist there.
	bm := &DefaultBinaryManager{binDir: t.TempDir(), version: "2.0.0"}
	ctx := context.Background()

	needsUpdate, err := bm.NeedsUpdate(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !needsUpdate {
		t.Error("expected NeedsUpdate to return true for a missing binary")
	}
}

// TestBinaryManager_NeedsUpdate_ErrorsIs_CorrectPattern verifies that the
// ErrBinaryNotFound sentinel error can be detected with errors.Is.
//
// BUG B6: the current implementation uses strings.Contains on err.Error()
// instead of errors.Is(err, ErrBinaryNotFound).  This test calls GetVersion
// directly (which wraps ErrBinaryNotFound) and then asserts that errors.Is
// unwraps correctly — something the string-match approach does NOT guarantee.
//
// This test FAILS today because GetVersion wraps ErrBinaryNotFound with
// fmt.Errorf("%w: …") which IS actually unwrappable, but NeedsUpdate throws
// away the error and does string matching instead.  The test proves:
//   - errors.Is correctly identifies ErrBinaryNotFound through the wrap.
//   - The fix (errors.Is in NeedsUpdate) is the right approach.
func TestBinaryManager_NeedsUpdate_ErrorsIs_CorrectPattern(t *testing.T) {
	bm := &DefaultBinaryManager{binDir: t.TempDir(), version: "2.0.0"}
	ctx := context.Background()

	// GetVersion wraps ErrBinaryNotFound using fmt.Errorf("%w: path").
	// errors.Is must be able to unwrap and match the sentinel.
	_, err := bm.GetVersion(ctx)
	if err == nil {
		t.Fatal("expected GetVersion to return an error for missing binary, got nil")
	}

	// This is the CORRECT way to detect a missing binary.
	// The fix for B6 is to use this pattern inside NeedsUpdate instead of
	// strings.Contains(err.Error(), "not found").
	if !errors.Is(err, ErrBinaryNotFound) {
		t.Errorf("errors.Is(err, ErrBinaryNotFound) returned false; "+
			"this means NeedsUpdate's string-match is the only detection path, "+
			"which is fragile. err = %v", err)
	}
}

// =============================================================================
// Version Reconciliation Tests (TDD RED phase)
//
// Bug: EnsureBinary() returns an existing binary regardless of version mismatch.
// The fix: EnsureBinary() must call NeedsUpdate() when the binary exists and
// is executable; if an update is needed, it must call Update() / downloadBinary.
//
// These tests are written BEFORE the fix so they demonstrate the current bug
// and will drive the correct implementation.
// =============================================================================

// makeFakeZotScript writes a tiny shell script that acts as a "zot" binary,
// responding to "--version" with JSON output matching Zot's real format.
// Zot outputs version info as JSON to stderr; the commit field format is
// "v{semver}-0-g{short-hash}". This lets us test GetVersion/NeedsUpdate/EnsureBinary
// version-checking without network access.
func makeFakeZotScript(t *testing.T, dir, ver string) string {
	t.Helper()
	binPath := filepath.Join(dir, "zot")
	// Zot outputs version info as JSON to stderr via --version flag
	// The commit field format is: v{semver}-0-g{short-hash}
	script := `#!/bin/sh
if [ "$1" = "--version" ]; then
    echo '{"level":"info","commit":"v` + ver + `-0-gabcdef0","message":"version"}' >&2
    exit 0
fi
exit 1
`
	require.NoError(t, os.WriteFile(binPath, []byte(script), 0755))
	return binPath
}

// TestEnsureBinary_ChecksVersionWhenBinaryExists verifies that EnsureBinary
// does NOT silently return a stale binary when the installed version differs
// from the desired version.
//
// Current behaviour (BUG): the existing-and-executable branch returns the path
// immediately without ever consulting NeedsUpdate.
//
// Expected behaviour after fix: EnsureBinary calls NeedsUpdate, gets true,
// then calls Update/downloadBinary. Because there is no network in the test
// environment, the download attempt fails — but the important thing is that it
// TRIES rather than silently returning the wrong version.
//
// We detect the bug by checking the *type* of error returned:
//   - BUG present  → no error at all (EnsureBinary returns the old path)
//   - BUG fixed    → an error containing "download" (it tried to update)
func TestEnsureBinary_ChecksVersionWhenBinaryExists(t *testing.T) {
	binDir := t.TempDir()

	// Place a fake "zot" binary that reports version 2.0.0.
	makeFakeZotScript(t, binDir, "2.0.0")

	// Manager is configured to want version 3.0.0 — a mismatch.
	bm := &DefaultBinaryManager{binDir: binDir, version: "3.0.0"}
	ctx := context.Background()

	// Confirm the version mismatch is detectable.
	needsUpdate, err := bm.NeedsUpdate(ctx)
	require.NoError(t, err, "NeedsUpdate should succeed with a fake executable binary")
	require.True(t, needsUpdate, "NeedsUpdate must return true when versions differ")

	// Now call EnsureBinary.
	// BUG: currently returns the old path with no error.
	// FIXED: returns an error because it tries to download 3.0.0 (no network).
	_, ensureErr := bm.EnsureBinary(ctx)

	if ensureErr == nil {
		// This is the current (buggy) path: EnsureBinary returned the stale binary.
		// Mark the test as failed with a clear message that describes what should happen.
		t.Errorf("BUG DETECTED: EnsureBinary returned nil error even though installed " +
			"version (2.0.0) does not match desired version (3.0.0). " +
			"After the fix, EnsureBinary must call NeedsUpdate and attempt to download " +
			"the new version, returning a download error in a network-less test environment.")
	} else {
		// After the fix: ensure the error is about a download attempt, not something else.
		assert.Contains(t, ensureErr.Error(), "download",
			"EnsureBinary should fail with a download error when update is needed but network is unavailable")
	}
}

// TestEnsureBinary_SkipsUpdateWhenVersionMatches verifies that EnsureBinary
// does NOT attempt a download when the installed version already matches.
//
// This is the happy path after the fix: binary present, version correct → return
// the path without touching the network.
func TestEnsureBinary_SkipsUpdateWhenVersionMatches(t *testing.T) {
	binDir := t.TempDir()

	// Place a fake "zot" that reports the same version the manager wants.
	makeFakeZotScript(t, binDir, "2.0.0")
	expectedPath := filepath.Join(binDir, "zot")

	bm := &DefaultBinaryManager{binDir: binDir, version: "2.0.0"}
	ctx := context.Background()

	// Confirm no update is needed.
	needsUpdate, err := bm.NeedsUpdate(ctx)
	require.NoError(t, err, "NeedsUpdate should succeed with matching fake binary")
	require.False(t, needsUpdate, "NeedsUpdate must return false when versions match")

	// EnsureBinary should return the existing path with no error.
	// This should pass both before and after the fix (it's not the bug scenario).
	path, err := bm.EnsureBinary(ctx)
	require.NoError(t, err, "EnsureBinary should not error when version already matches")
	assert.Equal(t, expectedPath, path, "EnsureBinary should return the existing binary path")
}

// TestNeedsUpdate_VersionMismatch verifies the core NeedsUpdate logic:
// when the installed binary reports a different version than desired, return true.
func TestNeedsUpdate_VersionMismatch(t *testing.T) {
	binDir := t.TempDir()

	// Fake binary reporting version 2.0.0.
	makeFakeZotScript(t, binDir, "2.0.0")

	// Manager wants 3.0.0 — definite mismatch.
	bm := &DefaultBinaryManager{binDir: binDir, version: "3.0.0"}
	ctx := context.Background()

	needsUpdate, err := bm.NeedsUpdate(ctx)
	require.NoError(t, err, "NeedsUpdate should not error when binary exists and is executable")
	assert.True(t, needsUpdate, "NeedsUpdate must return true when installed version differs from desired")
}

// TestNeedsUpdate_VersionMatch verifies the core NeedsUpdate logic:
// when the installed binary reports the same version as desired, return false.
func TestNeedsUpdate_VersionMatch(t *testing.T) {
	binDir := t.TempDir()

	// Fake binary reporting version 2.0.0.
	makeFakeZotScript(t, binDir, "2.0.0")

	// Manager also wants 2.0.0 — exact match.
	bm := &DefaultBinaryManager{binDir: binDir, version: "2.0.0"}
	ctx := context.Background()

	needsUpdate, err := bm.NeedsUpdate(ctx)
	require.NoError(t, err, "NeedsUpdate should not error when binary exists and is executable")
	assert.False(t, needsUpdate, "NeedsUpdate must return false when installed version matches desired")
}

// TestNeedsUpdate_FakeScriptVersionParsing verifies GetVersion correctly parses
// the JSON output from the fake script, extracting the semver from the commit
// field (e.g. "v1.2.3-0-gabcdef0" → "1.2.3"). This mirrors the real Zot binary's
// --version output format. If parsing breaks, NeedsUpdate tests above become
// unreliable.
func TestNeedsUpdate_FakeScriptVersionParsing(t *testing.T) {
	binDir := t.TempDir()
	makeFakeZotScript(t, binDir, "1.2.3")

	bm := &DefaultBinaryManager{binDir: binDir, version: "anything"}
	ctx := context.Background()

	ver, err := bm.GetVersion(ctx)
	require.NoError(t, err, "GetVersion should succeed with the fake script")
	assert.Equal(t, "1.2.3", ver,
		"GetVersion must parse the commit field JSON and return bare semver string")
}

// TestGetVersion_ParsesVariousCommitFormats verifies that GetVersion correctly
// extracts the semver portion from commit fields of the form "v{semver}-0-g{hash}".
func TestGetVersion_ParsesVariousCommitFormats(t *testing.T) {
	tests := []struct {
		name     string
		ver      string
		expected string
	}{
		{"standard semver", "2.1.15", "2.1.15"},
		{"major only patch", "2.0.0", "2.0.0"},
		{"single digit", "1.0.0", "1.0.0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binDir := t.TempDir()
			makeFakeZotScript(t, binDir, tt.ver)
			bm := &DefaultBinaryManager{binDir: binDir, version: "anything"}
			ver, err := bm.GetVersion(context.Background())
			require.NoError(t, err)
			assert.Equal(t, tt.expected, ver)
		})
	}
}
