package registry

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Clean-before-retry: installPackage removes corrupted node_modules
// =============================================================================

func TestInstallPackage_CleansCorruptedDir(t *testing.T) {
	// Verify that installPackage removes an existing node_modules/<pkg> dir
	// before running npm install — preventing ENOTEMPTY errors.
	fakeNpmDir := t.TempDir()
	// Fake npm: --version passes, install "succeeds" (exit 0, no real install).
	fakeNpmScript := `#!/bin/sh
if [ "$1" = "--version" ]; then echo "9.0.0"; exit 0; fi
if [ "$1" = "install" ]; then exit 0; fi
exit 1
`
	require.NoError(t, os.WriteFile(
		filepath.Join(fakeNpmDir, "npm"),
		[]byte(fakeNpmScript), 0755))

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", fakeNpmDir+string(os.PathListSeparator)+origPath)

	// Set up directory structure: binDir/../lib/node_modules/<pkg>/
	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	globalDir := filepath.Join(tmpDir, "lib", "node_modules", "test-clean-pkg")
	require.NoError(t, os.MkdirAll(binDir, 0755))
	require.NoError(t, os.MkdirAll(globalDir, 0755))

	// Plant a "corrupted" file in the node_modules dir.
	corruptedFile := filepath.Join(globalDir, "corrupted.txt")
	require.NoError(t, os.WriteFile(corruptedFile, []byte("corrupted"), 0644))

	mgr := &NpmBinaryManager{
		packageName: "test-clean-pkg",
		version:     "1.0.0",
		binDir:      binDir,
	}

	err := mgr.installPackage(context.Background())
	require.NoError(t, err)

	// The corrupted file should have been removed by the clean-before-install.
	_, statErr := os.Stat(corruptedFile)
	assert.True(t, os.IsNotExist(statErr),
		"corrupted file should be removed before npm install")
}

// =============================================================================
// Concurrent EnsureBinary — the core bug scenario from #372
// =============================================================================

func TestEnsureBinary_ConcurrentCallsSerialize(t *testing.T) {
	// This tests that concurrent EnsureBinary calls for the same package
	// do NOT corrupt the installation. When --concurrency > 1, multiple
	// VerdaccioManagers each create their own NpmBinaryManager and call
	// EnsureBinary concurrently. The fix adds a process-level mutex keyed
	// by package name that serialises the actual npm install.
	//
	// We use a fake npm that:
	//  - Passes --version
	//  - Reports "not installed" for list (exit 1, no package in output)
	//  - "Installs" by creating the binary file on disk (exit 0)

	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("requires /bin/sh")
	}

	tmpDir := t.TempDir()
	binDir := filepath.Join(tmpDir, "bin")
	require.NoError(t, os.MkdirAll(binDir, 0755))

	fakeNpmDir := t.TempDir()
	// The fake npm install creates the binary file on success.
	binTarget := filepath.Join(binDir, "conc-test-pkg")
	fakeNpmScript := `#!/bin/sh
if [ "$1" = "--version" ]; then echo "9.0.0"; exit 0; fi
if [ "$1" = "list" ]; then exit 1; fi
if [ "$1" = "install" ]; then
  sleep 0.05
  touch "` + binTarget + `"
  exit 0
fi
exit 1
`
	require.NoError(t, os.WriteFile(
		filepath.Join(fakeNpmDir, "npm"),
		[]byte(fakeNpmScript), 0755))

	origPath := os.Getenv("PATH")
	t.Setenv("PATH", fakeNpmDir+string(os.PathListSeparator)+origPath)

	const goroutines = 5
	var wg sync.WaitGroup
	paths := make([]string, goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			// Each goroutine creates its own NpmBinaryManager — simulating
			// multiple VerdaccioManagers with --concurrency > 1.
			mgr := &NpmBinaryManager{
				packageName: "conc-test-pkg",
				version:     "1.0.0",
				binDir:      binDir,
			}
			p, err := mgr.EnsureBinary(context.Background())
			paths[idx] = p
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	// All goroutines should succeed.
	for i := 0; i < goroutines; i++ {
		assert.NoError(t, errs[i], "goroutine %d should succeed", i)
		assert.Equal(t, binTarget, paths[i],
			"goroutine %d should return the correct binary path", i)
	}
}
