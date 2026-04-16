package registry

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// npmInstallMutex tests — process-level per-package mutex
// =============================================================================

func TestNpmInstallMutex_ReturnsSameMutexForSamePackage(t *testing.T) {
	mu1 := npmInstallMutex("test-pkg-same")
	mu2 := npmInstallMutex("test-pkg-same")
	assert.Same(t, mu1, mu2, "same package name must return the same mutex")
}

func TestNpmInstallMutex_ReturnsDifferentMutexForDifferentPackages(t *testing.T) {
	mu1 := npmInstallMutex("test-pkg-a")
	mu2 := npmInstallMutex("test-pkg-b")
	assert.NotSame(t, mu1, mu2, "different package names must return different mutexes")
}

func TestNpmInstallMutex_ConcurrentAccess(t *testing.T) {
	// Verify the map is safe under concurrent access.
	var wg sync.WaitGroup
	results := make([]*sync.Mutex, 50)
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = npmInstallMutex("test-pkg-concurrent")
		}(i)
	}
	wg.Wait()

	// All goroutines must get the same mutex.
	for i := 1; i < len(results); i++ {
		assert.Same(t, results[0], results[i],
			"concurrent calls for same package must return the same mutex")
	}
}

// =============================================================================
// acquireInstallLock tests — file-based cross-process locking
// =============================================================================

func TestAcquireInstallLock_CreatesLockFile(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "lib", ".test-install.lock")

	unlock, err := acquireInstallLock(context.Background(), lockPath)
	require.NoError(t, err, "should acquire lock on new file")
	defer unlock()

	_, statErr := os.Stat(lockPath)
	assert.NoError(t, statErr, "lock file should exist on disk")
}

func TestAcquireInstallLock_UnlockReleasesLock(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "lib", ".test-install.lock")

	unlock, err := acquireInstallLock(context.Background(), lockPath)
	require.NoError(t, err)
	unlock()

	// Should be able to re-acquire after unlock.
	unlock2, err := acquireInstallLock(context.Background(), lockPath)
	require.NoError(t, err, "should re-acquire lock after unlock")
	defer unlock2()
}

func TestAcquireInstallLock_SerializesConcurrentCallers(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "lib", ".test-install.lock")
	var counter int64

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unlock, err := acquireInstallLock(context.Background(), lockPath)
			if err != nil {
				t.Errorf("acquireInstallLock failed: %v", err)
				return
			}
			// Critical section: read-modify-write must be serialised.
			val := atomic.LoadInt64(&counter)
			time.Sleep(5 * time.Millisecond) // small window for races
			atomic.StoreInt64(&counter, val+1)
			unlock()
		}()
	}
	wg.Wait()

	assert.Equal(t, int64(5), atomic.LoadInt64(&counter),
		"all 5 goroutines must complete the critical section exactly once")
}

func TestAcquireInstallLock_RespectsContextCancellation(t *testing.T) {
	lockPath := filepath.Join(t.TempDir(), "lib", ".test-install.lock")

	// Hold the lock.
	unlock, err := acquireInstallLock(context.Background(), lockPath)
	require.NoError(t, err)
	defer unlock()

	// Try to acquire with an already-cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = acquireInstallLock(ctx, lockPath)
	assert.Error(t, err, "should fail when context is already cancelled")
	assert.Contains(t, err.Error(), "context cancelled",
		"error should mention context cancellation")
}
