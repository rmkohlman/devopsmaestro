package registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rmkohlman/MaestroSDK/paths"
)

// npmInstallMu serialises npm install operations across all NpmBinaryManager
// instances within the same process. Each VerdaccioManager creates its own
// NpmBinaryManager, so the per-instance mutex does not prevent concurrent
// installs when --concurrency > 1 launches multiple goroutines. This
// package-level map keyed by package name ensures only one goroutine runs
// npm install for a given package at a time.
//
// For cross-process protection, installPackage also acquires a file lock
// (see acquireInstallLock).
var (
	npmInstallMuMap   = make(map[string]*sync.Mutex)
	npmInstallMuMapMu sync.Mutex
)

// npmInstallMutex returns the process-wide mutex for the given npm package name.
func npmInstallMutex(packageName string) *sync.Mutex {
	npmInstallMuMapMu.Lock()
	defer npmInstallMuMapMu.Unlock()
	if mu, ok := npmInstallMuMap[packageName]; ok {
		return mu
	}
	mu := &sync.Mutex{}
	npmInstallMuMap[packageName] = mu
	return mu
}

const (
	// npmInstallMaxRetries is the maximum number of attempts for npm install.
	npmInstallMaxRetries = 3

	// npmInstallBaseDelay is the initial delay between retries (doubled each attempt).
	npmInstallBaseDelay = 2 * time.Second
)

// NpmBinaryManager manages binaries installed via npm global install.
type NpmBinaryManager struct {
	packageName string
	version     string
	binDir      string

	mu            sync.RWMutex
	binaryPath    string
	cachedVersion string
}

// NewNpmBinaryManager creates a new NpmBinaryManager for the given package.
func NewNpmBinaryManager(packageName, version string) *NpmBinaryManager {
	// Default npm global bin directory — this is a system path (not a DVM config path),
	// so we use paths.Default() to get the home directory consistently.
	var binDir string
	if pc, err := paths.Default(); err == nil {
		homeDir := filepath.Dir(pc.Root())
		binDir = filepath.Join(homeDir, ".npm-global", "bin")
	}

	return &NpmBinaryManager{
		packageName: packageName,
		version:     version,
		binDir:      binDir,
	}
}

// EnsureBinary ensures the binary exists, installing via npm if necessary.
// Returns the path to the binary.
//
// Concurrency safety: this method is safe to call from multiple goroutines
// and across multiple NpmBinaryManager instances for the same package.
// A process-level mutex serialises npm install attempts per package name,
// and a file-based lock prevents cross-process races (e.g. parallel dvm
// invocations). After acquiring the install lock, the method re-checks
// whether another caller already completed the installation.
func (m *NpmBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Fast path: binary is already cached and still exists on disk.
	if m.binaryPath != "" {
		if _, err := os.Stat(m.binaryPath); err == nil {
			return m.binaryPath, nil
		}
	}

	// Ensure npm is installed
	if err := m.ensureNpmInstalled(ctx); err != nil {
		return "", fmt.Errorf("npm not available: %w", err)
	}

	// Check if package is already installed and binary exists on disk.
	if binaryPath, ok, err := m.checkExistingInstallation(ctx); err != nil {
		return "", err
	} else if ok {
		m.binaryPath = binaryPath
		return m.binaryPath, nil
	}

	// Acquire the process-level mutex so only one goroutine installs at a time.
	installMu := npmInstallMutex(m.packageName)
	// Release per-instance lock while waiting for the install mutex to avoid
	// holding it for the duration of a potentially long npm install.
	m.mu.Unlock()
	installMu.Lock()
	defer installMu.Unlock()
	m.mu.Lock()

	// Re-check after acquiring the install lock — another goroutine may have
	// completed the installation while we were waiting.
	if binaryPath, ok, err := m.checkExistingInstallation(ctx); err != nil {
		return "", err
	} else if ok {
		m.binaryPath = binaryPath
		return m.binaryPath, nil
	}

	// Install package via npm (with file lock, cleanup, and retry).
	if err := m.installPackageWithRetry(ctx); err != nil {
		return "", fmt.Errorf("failed to install package: %w", err)
	}

	// Find binary path
	binaryPath := filepath.Join(m.binDir, m.packageName)
	if _, err := os.Stat(binaryPath); err != nil {
		return "", fmt.Errorf("binary not found after installation: %w", err)
	}

	m.binaryPath = binaryPath
	return m.binaryPath, nil
}

// checkExistingInstallation returns the binary path and true if the package
// is already installed and the binary exists on disk. Returns a non-nil error
// if isPackageInstalled fails with a fatal error (e.g. npm exits with code 2),
// which must be propagated to the caller rather than silently falling through
// to the install path (see BUG B7).
func (m *NpmBinaryManager) checkExistingInstallation(ctx context.Context) (string, bool, error) {
	installed, err := m.isPackageInstalled(ctx)
	if err != nil {
		return "", false, fmt.Errorf("failed to check package status: %w", err)
	}
	if !installed {
		return "", false, nil
	}
	binaryPath := filepath.Join(m.binDir, m.packageName)
	if _, err := os.Stat(binaryPath); err != nil {
		return "", false, nil
	}
	return binaryPath, true, nil
}

// GetVersion returns the version of the currently installed binary.
func (m *NpmBinaryManager) GetVersion(ctx context.Context) (string, error) {
	m.mu.RLock()
	if m.cachedVersion != "" {
		version := m.cachedVersion
		m.mu.RUnlock()
		return version, nil
	}
	m.mu.RUnlock()

	// Ensure binary exists
	_, err := m.EnsureBinary(ctx)
	if err != nil {
		return "", fmt.Errorf("binary not installed: %w", err)
	}

	// Get version from package
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", m.packageName, "--depth=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w (output: %s)", err, string(output))
	}

	// Parse version from output
	// Example: verdaccio@6.1.2
	version := m.parseVersionFromList(string(output))
	if version == "" {
		return "", fmt.Errorf("failed to parse version from npm list output")
	}

	m.mu.Lock()
	m.cachedVersion = version
	m.mu.Unlock()

	return version, nil
}

// NeedsUpdate checks if the binary needs to be updated to the desired version.
func (m *NpmBinaryManager) NeedsUpdate(ctx context.Context) (bool, error) {
	currentVersion, err := m.GetVersion(ctx)
	if err != nil {
		// If we can't get version, assume update is needed
		return true, nil
	}

	// Compare versions (simple string comparison for now)
	return currentVersion != m.version, nil
}

// Update downloads and installs the latest version of the binary.
func (m *NpmBinaryManager) Update(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure npm is installed
	if err := m.ensureNpmInstalled(ctx); err != nil {
		return fmt.Errorf("npm not available: %w", err)
	}

	// Clean up previous installation to prevent ENOTEMPTY errors
	globalDir := filepath.Join(m.binDir, "..", "lib", "node_modules", m.packageName)
	if _, err := os.Stat(globalDir); err == nil {
		if err := os.RemoveAll(globalDir); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to clean up previous installation at %s: %v\n", globalDir, err)
		}
	}

	// Update package via npm
	packageSpec := m.packageName
	if m.version != "" {
		packageSpec = fmt.Sprintf("%s@%s", m.packageName, m.version)
	}

	cmd := exec.CommandContext(ctx, "npm", "install", "-g", packageSpec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm install failed: %w (output: %s)", err, string(output))
	}

	// Invalidate cached version
	m.cachedVersion = ""

	// Update binary path
	binaryPath := filepath.Join(m.binDir, m.packageName)
	if _, err := os.Stat(binaryPath); err == nil {
		m.binaryPath = binaryPath
	}

	return nil
}

// ensureNpmInstalled checks if npm is available.
func (m *NpmBinaryManager) ensureNpmInstalled(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "npm", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("npm not found: %w", err)
	}
	return nil
}

// isPackageInstalled checks if the package is already installed globally.
func (m *NpmBinaryManager) isPackageInstalled(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "npm", "list", "-g", m.packageName, "--depth=0")
	output, err := cmd.CombinedOutput()

	if err != nil {
		// npm list exits with code 1 when the package is not found — that is
		// expected and means "not installed". Any other error (signal, code 2+,
		// missing binary, context cancellation) is a real failure.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			// Exit code 1: package not found — check output as before
			return strings.Contains(string(output), m.packageName), nil
		}
		// Non-exit-code-1 error: propagate it
		return false, fmt.Errorf("npm list failed: %w", err)
	}

	// npm list succeeded (exit 0): package is installed if its name appears in output
	return strings.Contains(string(output), m.packageName), nil
}

// installPackageWithRetry wraps installPackage with file-based locking and
// exponential-backoff retries. The file lock prevents cross-process races
// (e.g. multiple dvm processes); the retry loop handles transient npm/tar
// errors caused by stale caches or partially-written node_modules trees.
func (m *NpmBinaryManager) installPackageWithRetry(ctx context.Context) error {
	lockPath := filepath.Join(m.binDir, "..", "lib", fmt.Sprintf(".%s-install.lock", m.packageName))

	// Acquire file-based lock (blocks until lock is available or ctx is cancelled).
	unlock, err := acquireInstallLock(ctx, lockPath)
	if err != nil {
		return fmt.Errorf("failed to acquire install lock: %w", err)
	}
	defer unlock()

	// After acquiring the file lock, another process may have already
	// completed the installation — check one more time.
	if binaryPath, ok, err := m.checkExistingInstallation(ctx); err != nil {
		return err
	} else if ok {
		m.binaryPath = binaryPath
		return nil
	}

	var lastErr error
	delay := npmInstallBaseDelay

	for attempt := 1; attempt <= npmInstallMaxRetries; attempt++ {
		if err := m.installPackage(ctx); err != nil {
			lastErr = fmt.Errorf("attempt %d/%d: %w", attempt, npmInstallMaxRetries, err)

			// Don't sleep after the last attempt.
			if attempt < npmInstallMaxRetries {
				select {
				case <-ctx.Done():
					return fmt.Errorf("install cancelled during retry: %w", ctx.Err())
				case <-time.After(delay):
				}
				delay *= 2 // exponential backoff
			}
			continue
		}
		return nil // success
	}

	return fmt.Errorf("npm install failed after %d attempts: %w", npmInstallMaxRetries, lastErr)
}

// acquireInstallLock acquires an exclusive file lock on lockPath, creating the
// file if necessary. Returns an unlock function that must be called when done.
// The lock uses flock(2) which is automatically released if the process crashes.
func acquireInstallLock(ctx context.Context, lockPath string) (func(), error) {
	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(lockPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create lock directory: %w", err)
	}

	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to open lock file: %w", err)
	}

	// Attempt to acquire the lock in a loop, respecting context cancellation.
	// We use non-blocking flock with polling rather than blocking flock so that
	// context cancellation is honoured.
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	for {
		err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			// Lock acquired.
			return func() {
				syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
				f.Close()
			}, nil
		}

		// EWOULDBLOCK means another process holds the lock — wait and retry.
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			f.Close()
			return nil, fmt.Errorf("flock failed: %w", err)
		}

		select {
		case <-ctx.Done():
			f.Close()
			return nil, fmt.Errorf("context cancelled while waiting for install lock: %w", ctx.Err())
		case <-ticker.C:
			// retry
		}
	}
}

// installPackage installs the package via npm global install.
// Caller must hold the process-level install mutex and file lock.
func (m *NpmBinaryManager) installPackage(ctx context.Context) error {
	// Ensure bin directory exists
	if err := os.MkdirAll(m.binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Clean up any previous installation to prevent ENOTEMPTY errors.
	// A partial or corrupted previous install can leave directories in an
	// inconsistent state that causes npm's rmdir to fail with ENOTEMPTY.
	globalDir := filepath.Join(m.binDir, "..", "lib", "node_modules", m.packageName)
	if _, err := os.Stat(globalDir); err == nil {
		if err := os.RemoveAll(globalDir); err != nil {
			// Non-fatal: log and continue — npm may still succeed
			fmt.Fprintf(os.Stderr, "warning: failed to clean up previous installation at %s: %v\n", globalDir, err)
		}
	}

	// Install package
	packageSpec := m.packageName
	if m.version != "" {
		packageSpec = fmt.Sprintf("%s@%s", m.packageName, m.version)
	}

	cmd := exec.CommandContext(ctx, "npm", "install", "-g", packageSpec, fmt.Sprintf("--prefix=%s/..", m.binDir))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("npm install failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// parseVersionFromList extracts version from npm list output.
// Example output: "verdaccio@6.1.2"
func (m *NpmBinaryManager) parseVersionFromList(output string) string {
	// Pattern: packageName@version
	pattern := regexp.MustCompile(m.packageName + `@([\d\.]+)`)
	matches := pattern.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}
