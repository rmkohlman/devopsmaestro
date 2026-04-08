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

	"github.com/rmkohlman/MaestroSDK/paths"
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
func (m *NpmBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if binary is already available
	if m.binaryPath != "" {
		if _, err := os.Stat(m.binaryPath); err == nil {
			return m.binaryPath, nil
		}
	}

	// Ensure npm is installed
	if err := m.ensureNpmInstalled(ctx); err != nil {
		return "", fmt.Errorf("npm not available: %w", err)
	}

	// Check if package is already installed
	if installed, err := m.isPackageInstalled(ctx); err != nil {
		return "", fmt.Errorf("failed to check package status: %w", err)
	} else if installed {
		// Package is installed, find the binary
		binaryPath := filepath.Join(m.binDir, m.packageName)
		if _, err := os.Stat(binaryPath); err == nil {
			m.binaryPath = binaryPath
			return m.binaryPath, nil
		}
	}

	// Install package via npm
	if err := m.installPackage(ctx); err != nil {
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
	// Example: verdaccio@5.28.0
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

// installPackage installs the package via npm global install.
func (m *NpmBinaryManager) installPackage(ctx context.Context) error {
	// Ensure bin directory exists
	if err := os.MkdirAll(m.binDir, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
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
// Example output: "verdaccio@5.28.0"
func (m *NpmBinaryManager) parseVersionFromList(output string) string {
	// Pattern: packageName@version
	pattern := regexp.MustCompile(m.packageName + `@([\d\.]+)`)
	matches := pattern.FindStringSubmatch(output)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}
