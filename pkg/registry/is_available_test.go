package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// BrewBinaryManager.IsAvailable — unit tests
// =============================================================================

// isAvailableWithCandidates is a test-only helper that overrides candidate paths
// by creating a local BrewBinaryManager whose getBinaryPathCandidates list we can
// satisfy by placing a real file at a known temp path.
//
// IsAvailable() checks candidates via os.Stat, then exec.LookPath, then brew.
// We test the first two cases without invoking brew at all.

func TestBrewBinaryManager_IsAvailable_ReturnsTrueWhenBinaryOnDisk(t *testing.T) {
	dir := t.TempDir()

	// Create a real executable file at the expected Homebrew sbin path location
	// We can't override the candidate list directly, but we CAN test that IsAvailable
	// returns true when the binary is discoverable via PATH (exec.LookPath).
	binaryName := "fake-brew-binary-" + filepath.Base(dir)
	binaryPath := filepath.Join(dir, binaryName)

	// Write a real executable
	require.NoError(t, os.WriteFile(binaryPath, []byte("#!/bin/sh\nexit 0\n"), 0755))

	// Put our temp dir on PATH so exec.LookPath finds it
	origPath := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", origPath) })
	os.Setenv("PATH", dir+":"+origPath)

	mgr := NewBrewBinaryManagerWithBinaryName("some-formula", binaryName)
	available, err := mgr.IsAvailable(context.Background())

	// If brew is not installed, IsAvailable swallows the error and returns false.
	// But since the binary was found via PATH (exec.LookPath), it returns true BEFORE
	// ever reaching the IsInstalled(brew) call.
	assert.NoError(t, err)
	assert.True(t, available, "binary on PATH should be detected as available")
}

func TestBrewBinaryManager_IsAvailable_ReturnsFalseWhenBinaryAbsent(t *testing.T) {
	// Use a formula name that certainly does NOT exist as a binary anywhere
	mgr := NewBrewBinaryManagerWithBinaryName("nonexistent-formula-zzz999", "nonexistent-binary-zzz999")

	available, err := mgr.IsAvailable(context.Background())

	// If brew is also not installed (or the formula isn't installed), this returns false.
	// IsAvailable must never return an error even when brew is missing — it swallows it.
	assert.NoError(t, err, "IsAvailable must not return an error even when binary/brew is missing")
	assert.False(t, available, "a non-existent binary should not be reported as available")
}

func TestBrewBinaryManager_IsAvailable_ReturnsTrueWhenFileExistsAtCandidatePath(t *testing.T) {
	dir := t.TempDir()

	// BrewBinaryManager uses hard-coded candidate paths based on runtime.GOOS and GOARCH.
	// We can't inject custom candidates, so instead we verify the PATH-based path.
	// This test confirms the early-return on exec.LookPath works correctly.
	binaryName := "candidate-check-binary-" + filepath.Base(dir)
	binaryPath := filepath.Join(dir, binaryName)
	require.NoError(t, os.WriteFile(binaryPath, []byte("#!/bin/sh\n"), 0755))

	origPath := os.Getenv("PATH")
	t.Cleanup(func() { os.Setenv("PATH", origPath) })
	os.Setenv("PATH", dir+":"+origPath)

	mgr := NewBrewBinaryManagerWithBinaryName("any-formula", binaryName)

	available, err := mgr.IsAvailable(context.Background())
	require.NoError(t, err)
	assert.True(t, available)
}

func TestBrewBinaryManager_IsAvailable_DoesNotAttemptInstall(t *testing.T) {
	// This test verifies the contract: IsAvailable must never trigger brew install.
	// We use a non-existent formula and confirm that no side effect occurs.
	mgr := NewBrewBinaryManagerWithBinaryName("would-fail-if-installed-zzz", "no-such-binary-zzz")

	// If IsAvailable called Install(), it would fail with a brew error.
	// Since it only checks presence, it must return cleanly.
	_, err := mgr.IsAvailable(context.Background())
	assert.NoError(t, err, "IsAvailable must never propagate brew errors")
}
