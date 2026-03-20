package registry

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// PipxBinaryManager manages Python binaries installed via pipx.
type PipxBinaryManager struct {
	packageName string
	version     string
	binDir      string

	mu            sync.RWMutex
	binaryPath    string
	cachedVersion string
}

// NewPipxBinaryManager creates a new PipxBinaryManager for the given package.
func NewPipxBinaryManager(packageName, version string) *PipxBinaryManager {
	// Default pipx bin directory
	homeDir, _ := os.UserHomeDir()
	binDir := filepath.Join(homeDir, ".local", "bin")

	return &PipxBinaryManager{
		packageName: packageName,
		version:     version,
		binDir:      binDir,
	}
}

// EnsureBinary ensures the binary exists, installing via pipx if necessary.
// Returns the path to the binary.
func (m *PipxBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if binary is already available
	if m.binaryPath != "" {
		if _, err := os.Stat(m.binaryPath); err == nil {
			return m.binaryPath, nil
		}
	}

	// Ensure pipx is installed; if not, try pip as fallback
	if err := m.ensurePipxInstalled(ctx); err != nil {
		// pipx not available — fall back to pip
		binaryPath, pipErr := m.fallbackPipInstall(ctx)
		if pipErr != nil {
			return "", fmt.Errorf("failed to install %s: pipx unavailable and pip fallback failed: %w", m.packageName, pipErr)
		}
		m.binaryPath = binaryPath
		return m.binaryPath, nil
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

	// Install package via pipx
	if err := m.installPackage(ctx); err != nil {
		return "", fmt.Errorf("failed to install package: %w", err)
	}

	// Find binary path
	binaryPath := filepath.Join(m.binDir, m.packageName)
	if _, err := os.Stat(binaryPath); err != nil {
		return "", fmt.Errorf("pipx install succeeded but binary not found at %s: %w", binaryPath, err)
	}

	m.binaryPath = binaryPath
	return m.binaryPath, nil
}

// GetVersion returns the version of the currently installed binary.
func (m *PipxBinaryManager) GetVersion(ctx context.Context) (string, error) {
	m.mu.RLock()
	if m.cachedVersion != "" {
		version := m.cachedVersion
		m.mu.RUnlock()
		return version, nil
	}
	m.mu.RUnlock()

	// Ensure binary exists
	binaryPath, err := m.EnsureBinary(ctx)
	if err != nil {
		return "", fmt.Errorf("binary not installed: %w", err)
	}

	// Get version from pipx list
	cmd := exec.CommandContext(ctx, "pipx", "list", "--short")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w", err)
	}

	// Parse version from output
	// Format: "package-name 1.2.3"
	lines := strings.Split(string(output), "\n")
	versionRegex := regexp.MustCompile(fmt.Sprintf(`%s\s+(\d+\.\d+(?:\.\d+)?)`, regexp.QuoteMeta(m.packageName)))

	for _, line := range lines {
		if matches := versionRegex.FindStringSubmatch(line); len(matches) > 1 {
			m.mu.Lock()
			m.cachedVersion = matches[1]
			version := m.cachedVersion
			m.mu.Unlock()
			return version, nil
		}
	}

	// Fallback: Try running the binary with --version
	cmd = exec.CommandContext(ctx, binaryPath, "--version")
	output, err = cmd.CombinedOutput()
	if err == nil {
		// Extract version from output (first number sequence)
		versionRegex = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)
		if match := versionRegex.FindString(string(output)); match != "" {
			m.mu.Lock()
			m.cachedVersion = match
			version := m.cachedVersion
			m.mu.Unlock()
			return version, nil
		}
	}

	return "", fmt.Errorf("failed to determine version")
}

// NeedsUpdate checks if the binary needs to be updated to the desired version.
func (m *PipxBinaryManager) NeedsUpdate(ctx context.Context) (bool, error) {
	// Check if package is installed
	installed, err := m.isPackageInstalled(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check package status: %w", err)
	}

	if !installed {
		return true, nil // Needs installation
	}

	// Get current version
	currentVersion, err := m.GetVersion(ctx)
	if err != nil {
		return true, nil // Can't get version, assume needs update
	}

	// Compare versions (simple string comparison for now)
	// In production, use semver comparison
	return currentVersion != m.version, nil
}

// Update downloads and installs the latest version of the binary.
func (m *PipxBinaryManager) Update(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if package is installed
	installed, err := m.isPackageInstalled(ctx)
	if err != nil {
		return fmt.Errorf("failed to check package status: %w", err)
	}

	if !installed {
		// Install fresh if not present
		return m.installPackage(ctx)
	}

	// Upgrade existing installation
	cmd := exec.CommandContext(ctx, "pipx", "upgrade", m.packageName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to upgrade package: %w (output: %s)", err, string(output))
	}

	// Clear cached version
	m.cachedVersion = ""

	return nil
}

// ensurePipxInstalled checks if pipx is available.
func (m *PipxBinaryManager) ensurePipxInstalled(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "pipx", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pipx not found in PATH")
	}
	return nil
}

// isPackageInstalled checks if the package is already installed via pipx.
func (m *PipxBinaryManager) isPackageInstalled(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "pipx", "list", "--short")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("failed to list packages: %w", err)
	}

	// Check if package name appears in output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), m.packageName+" ") || strings.TrimSpace(line) == m.packageName {
			return true, nil
		}
	}

	return false, nil
}

// installPackage installs the package via pipx.
func (m *PipxBinaryManager) installPackage(ctx context.Context) error {
	// Build install command
	args := []string{"install"}

	// If version is specified, use package==version syntax
	if m.version != "" {
		args = append(args, fmt.Sprintf("%s==%s", m.packageName, m.version))
	} else {
		args = append(args, m.packageName)
	}

	cmd := exec.CommandContext(ctx, "pipx", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install package: %w (output: %s)", err, string(output))
	}

	// Clear cached version after install
	m.cachedVersion = ""

	return nil
}

// fallbackPipInstall attempts to install the package using pip when pipx is
// not available. This uses "python3 -m pip install --user" which installs to
// the user's local bin directory (~/.local/bin on Linux/macOS).
func (m *PipxBinaryManager) fallbackPipInstall(ctx context.Context) (string, error) {
	// Build the package spec
	packageSpec := m.packageName
	if m.version != "" {
		packageSpec = fmt.Sprintf("%s==%s", m.packageName, m.version)
	}

	// Try python3 -m pip install --user
	cmd := exec.CommandContext(ctx, "python3", "-m", "pip", "install", "--user", packageSpec)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("pip install failed: %w (output: %s)", err, string(output))
	}

	// Find the installed binary
	// pip --user installs to ~/.local/bin on Linux/macOS
	binaryPath := filepath.Join(m.binDir, m.packageName)
	if _, err := os.Stat(binaryPath); err != nil {
		// Also check if python3 knows the user bin path
		userBase, err2 := m.getPythonUserBase(ctx)
		if err2 == nil {
			altPath := filepath.Join(userBase, "bin", m.packageName)
			if _, err3 := os.Stat(altPath); err3 == nil {
				return altPath, nil
			}
		}
		return "", fmt.Errorf("binary not found after pip install: %w", err)
	}

	return binaryPath, nil
}

// getPythonUserBase returns the Python user base directory (site.USER_BASE).
func (m *PipxBinaryManager) getPythonUserBase(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "python3", "-m", "site", "--user-base")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
