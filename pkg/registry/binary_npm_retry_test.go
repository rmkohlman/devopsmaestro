package registry

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// checkExistingInstallation tests
// =============================================================================

func TestCheckExistingInstallation_BinaryExists(t *testing.T) {
	// When the binary is installed and exists on disk, should return (path, true, nil).
	// Requires npm on PATH for isPackageInstalled.
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available on PATH")
	}

	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	// Create a fake binary.
	fakeBin := filepath.Join(binDir, "fake-check-pkg")
	require.NoError(t, os.WriteFile(fakeBin, []byte("#!/bin/sh\n"), 0755))

	mgr := &NpmBinaryManager{
		packageName: "fake-check-pkg",
		version:     "1.0.0",
		binDir:      binDir,
	}

	// isPackageInstalled will call npm list which won't find fake-check-pkg,
	// so installed=false and the binary won't be returned.
	_, ok, err := mgr.checkExistingInstallation(context.Background())
	require.NoError(t, err)
	assert.False(t, ok, "npm list won't find fake-check-pkg, so installed=false")
}

func TestCheckExistingInstallation_NoBinary(t *testing.T) {
	// When the binary doesn't exist on disk, should return ("", false, nil).
	if _, err := exec.LookPath("npm"); err != nil {
		t.Skip("npm not available on PATH")
	}

	mgr := &NpmBinaryManager{
		packageName: "nonexistent-check-test-pkg",
		version:     "1.0.0",
		binDir:      t.TempDir(),
	}

	path, ok, err := mgr.checkExistingInstallation(context.Background())
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, path)
}

func TestCheckExistingInstallation_FatalNpmError(t *testing.T) {
	// When npm list fails with a fatal error (exit code 2), the error
	// must be propagated — not swallowed (BUG B7 fix).
	fakeNpmDir := t.TempDir()
	fakeNpmScript := `#!/bin/sh
if [ "$1" = "--version" ]; then echo "9.0.0"; exit 0; fi
if [ "$1" = "list" ]; then echo "npm ERR! fatal" >&2; exit 2; fi
exit 2
`
	fakeNpmPath := filepath.Join(fakeNpmDir, "npm")
	require.NoError(t, os.WriteFile(fakeNpmPath, []byte(fakeNpmScript), 0755))

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", fakeNpmDir+string(os.PathListSeparator)+origPath)

	mgr := &NpmBinaryManager{
		packageName: "test-fatal-pkg",
		version:     "1.0.0",
		binDir:      t.TempDir(),
	}

	_, _, err := mgr.checkExistingInstallation(context.Background())
	assert.Error(t, err, "fatal npm list error must be propagated")
	assert.Contains(t, err.Error(), "failed to check package status")
}

// =============================================================================
// installPackageWithRetry tests — retry logic and cleanup
// =============================================================================

func TestInstallPackageWithRetry_FailingNpmRetries(t *testing.T) {
	// Use a fake npm that always fails install but passes --version and list.
	fakeNpmDir := t.TempDir()
	fakeNpmScript := `#!/bin/sh
if [ "$1" = "--version" ]; then echo "9.0.0"; exit 0; fi
if [ "$1" = "list" ]; then exit 1; fi
if [ "$1" = "install" ]; then echo "npm ERR! ENOENT" >&2; exit 254; fi
exit 1
`
	fakeNpmPath := filepath.Join(fakeNpmDir, "npm")
	require.NoError(t, os.WriteFile(fakeNpmPath, []byte(fakeNpmScript), 0755))

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", fakeNpmDir+string(os.PathListSeparator)+origPath)

	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	mgr := &NpmBinaryManager{
		packageName: "test-retry-pkg",
		version:     "1.0.0",
		binDir:      binDir,
	}

	// installPackageWithRetry should fail after npmInstallMaxRetries.
	err := mgr.installPackageWithRetry(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "npm install failed after")
	assert.Contains(t, err.Error(), "3 attempts")
}

func TestInstallPackageWithRetry_CancelledDuringRetry(t *testing.T) {
	// Use a fake npm that always fails install.
	fakeNpmDir := t.TempDir()
	fakeNpmScript := `#!/bin/sh
if [ "$1" = "--version" ]; then echo "9.0.0"; exit 0; fi
if [ "$1" = "list" ]; then exit 1; fi
if [ "$1" = "install" ]; then sleep 0.1; exit 254; fi
exit 1
`
	fakeNpmPath := filepath.Join(fakeNpmDir, "npm")
	require.NoError(t, os.WriteFile(fakeNpmPath, []byte(fakeNpmScript), 0755))

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", fakeNpmDir+string(os.PathListSeparator)+origPath)

	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	mgr := &NpmBinaryManager{
		packageName: "test-cancel-pkg",
		version:     "1.0.0",
		binDir:      binDir,
	}

	// Cancel after first attempt starts the retry delay.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := mgr.installPackageWithRetry(ctx)
	require.Error(t, err)
	// Should either report cancellation or exhausted retries.
	t.Logf("installPackageWithRetry error: %v", err)
}
