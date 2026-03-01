package registry

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

// BrewBinaryManager manages binaries installed via Homebrew.
type BrewBinaryManager struct {
	formulaName string
	binaryName  string

	mu            sync.RWMutex
	binaryPath    string
	cachedVersion string
}

// NewBrewBinaryManager creates a new BrewBinaryManager for the given formula.
// The binaryName defaults to formulaName if not provided separately.
func NewBrewBinaryManager(formulaName string) *BrewBinaryManager {
	return &BrewBinaryManager{
		formulaName: formulaName,
		binaryName:  formulaName, // Default: binary name matches formula name
	}
}

// NewBrewBinaryManagerWithBinaryName creates a BrewBinaryManager with a custom binary name.
// Use this when the binary name differs from the formula name.
func NewBrewBinaryManagerWithBinaryName(formulaName, binaryName string) *BrewBinaryManager {
	return &BrewBinaryManager{
		formulaName: formulaName,
		binaryName:  binaryName,
	}
}

// NewSquidBinaryManager creates a BrewBinaryManager configured for Squid.
func NewSquidBinaryManager() *BrewBinaryManager {
	return NewBrewBinaryManager("squid")
}

// Install installs the formula via Homebrew.
func (m *BrewBinaryManager) Install(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Validate formula name
	if m.formulaName == "" {
		return fmt.Errorf("formula name cannot be empty")
	}

	// Ensure brew is installed
	if err := m.ensureBrewInstalled(ctx); err != nil {
		return fmt.Errorf("brew not available: %w", err)
	}

	// Run brew install
	cmd := exec.CommandContext(ctx, "brew", "install", m.formulaName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Check for specific error patterns
		outputStr := string(output)
		if strings.Contains(outputStr, "No formulae or casks found") ||
			strings.Contains(outputStr, "No available formula") {
			return fmt.Errorf("formula not found: %s", m.formulaName)
		}
		return fmt.Errorf("brew install failed: %w (output: %s)", err, outputStr)
	}

	// Clear cached data after successful install
	m.binaryPath = ""
	m.cachedVersion = ""

	return nil
}

// Uninstall removes the formula via Homebrew.
func (m *BrewBinaryManager) Uninstall(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure brew is installed
	if err := m.ensureBrewInstalled(ctx); err != nil {
		return fmt.Errorf("brew not available: %w", err)
	}

	// Check if formula is installed first (for idempotency)
	installed, err := m.isInstalledLocked(ctx)
	if err != nil {
		return fmt.Errorf("failed to check installation status: %w", err)
	}
	if !installed {
		// Already uninstalled, nothing to do
		return nil
	}

	// Run brew uninstall
	cmd := exec.CommandContext(ctx, "brew", "uninstall", m.formulaName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("brew uninstall failed: %w (output: %s)", err, string(output))
	}

	// Clear cached data
	m.binaryPath = ""
	m.cachedVersion = ""

	return nil
}

// IsInstalled checks if the formula is installed via Homebrew.
func (m *BrewBinaryManager) IsInstalled(ctx context.Context) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.isInstalledLocked(ctx)
}

// isInstalledLocked checks if the formula is installed (caller must hold lock).
func (m *BrewBinaryManager) isInstalledLocked(ctx context.Context) (bool, error) {
	// Ensure brew is installed
	if err := m.ensureBrewInstalled(ctx); err != nil {
		return false, fmt.Errorf("brew not available: %w", err)
	}

	// Run brew list to check if formula is installed
	cmd := exec.CommandContext(ctx, "brew", "list", m.formulaName)
	err := cmd.Run()
	if err != nil {
		// brew list returns exit code 1 if formula not installed
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to check formula status: %w", err)
	}

	return true, nil
}

// GetVersion returns the version of the installed binary.
// For Squid, parses output like "Squid Cache: Version X.Y" to extract "X.Y".
func (m *BrewBinaryManager) GetVersion(ctx context.Context) (string, error) {
	m.mu.RLock()
	if m.cachedVersion != "" {
		version := m.cachedVersion
		m.mu.RUnlock()
		return version, nil
	}
	m.mu.RUnlock()

	// Get binary path
	binaryPath, err := m.GetBinaryPath(ctx)
	if err != nil {
		return "", fmt.Errorf("binary not installed: %w", err)
	}

	// Run binary with -v flag to get version
	cmd := exec.CommandContext(ctx, binaryPath, "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get version: %w (output: %s)", err, string(output))
	}

	// Parse version from output
	version := m.parseVersion(string(output))
	if version == "" {
		return "", fmt.Errorf("failed to parse version from output: %s", string(output))
	}

	m.mu.Lock()
	m.cachedVersion = version
	m.mu.Unlock()

	return version, nil
}

// parseVersion extracts version from various output formats.
// Handles formats like:
// - "Squid Cache: Version X.Y" -> "X.Y"
// - "version X.Y.Z" -> "X.Y.Z"
// - Just "X.Y.Z" -> "X.Y.Z"
func (m *BrewBinaryManager) parseVersion(output string) string {
	// Try "Squid Cache: Version X.Y" pattern (Squid-specific)
	squidPattern := regexp.MustCompile(`Squid Cache: Version (\d+\.\d+(?:\.\d+)?)`)
	if matches := squidPattern.FindStringSubmatch(output); len(matches) > 1 {
		return matches[1]
	}

	// Try generic "Version X.Y.Z" pattern
	versionPattern := regexp.MustCompile(`[Vv]ersion[:\s]+(\d+\.\d+(?:\.\d+)?)`)
	if matches := versionPattern.FindStringSubmatch(output); len(matches) > 1 {
		return matches[1]
	}

	// Try standalone version number pattern
	standalonePattern := regexp.MustCompile(`^(\d+\.\d+(?:\.\d+)?)\s*$`)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		if matches := standalonePattern.FindStringSubmatch(strings.TrimSpace(line)); len(matches) > 1 {
			return matches[1]
		}
	}

	// Fallback: find any version-like number
	fallbackPattern := regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?)`)
	if matches := fallbackPattern.FindStringSubmatch(output); len(matches) > 1 {
		return matches[1]
	}

	return ""
}

// GetBinaryPath returns the path to the installed binary.
// Checks architecture-specific Homebrew paths:
// - /opt/homebrew/bin/ (macOS ARM64)
// - /usr/local/bin/ (macOS Intel)
// - /home/linuxbrew/.linuxbrew/bin/ (Linux)
func (m *BrewBinaryManager) GetBinaryPath(ctx context.Context) (string, error) {
	m.mu.RLock()
	if m.binaryPath != "" {
		path := m.binaryPath
		m.mu.RUnlock()
		return path, nil
	}
	m.mu.RUnlock()

	// Check if formula is installed first
	installed, err := m.IsInstalled(ctx)
	if err != nil {
		return "", err
	}
	if !installed {
		return "", fmt.Errorf("binary not installed: %s", m.binaryName)
	}

	// Try architecture-specific paths
	candidates := m.getBinaryPathCandidates()

	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			m.mu.Lock()
			m.binaryPath = candidate
			m.mu.Unlock()
			return candidate, nil
		}
	}

	// Fallback: use exec.LookPath
	path, err := exec.LookPath(m.binaryName)
	if err != nil {
		return "", fmt.Errorf("binary not found in PATH: %s", m.binaryName)
	}

	m.mu.Lock()
	m.binaryPath = path
	m.mu.Unlock()

	return path, nil
}

// getBinaryPathCandidates returns the list of candidate paths to check.
func (m *BrewBinaryManager) getBinaryPathCandidates() []string {
	candidates := []string{}

	switch runtime.GOOS {
	case "darwin":
		if runtime.GOARCH == "arm64" {
			// Apple Silicon (M1/M2/M3)
			candidates = append(candidates, fmt.Sprintf("/opt/homebrew/bin/%s", m.binaryName))
			candidates = append(candidates, fmt.Sprintf("/opt/homebrew/sbin/%s", m.binaryName))
		}
		// Intel Mac (also fallback for ARM)
		candidates = append(candidates, fmt.Sprintf("/usr/local/bin/%s", m.binaryName))
		candidates = append(candidates, fmt.Sprintf("/usr/local/sbin/%s", m.binaryName))
	case "linux":
		// Linuxbrew
		candidates = append(candidates, fmt.Sprintf("/home/linuxbrew/.linuxbrew/bin/%s", m.binaryName))
		candidates = append(candidates, fmt.Sprintf("/home/linuxbrew/.linuxbrew/sbin/%s", m.binaryName))
	}

	return candidates
}

// ensureBrewInstalled checks if Homebrew is available.
func (m *BrewBinaryManager) ensureBrewInstalled(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "brew", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("brew not found in PATH")
	}
	return nil
}

// =============================================================================
// BinaryManager Interface Implementation
// =============================================================================

// EnsureBinary ensures the binary exists, installing via Homebrew if necessary.
// Returns the path to the binary.
func (m *BrewBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	// Check if already installed
	installed, err := m.IsInstalled(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to check installation status: %w", err)
	}

	if !installed {
		// Install via Homebrew
		if err := m.Install(ctx); err != nil {
			return "", fmt.Errorf("failed to install: %w", err)
		}
	}

	// Return binary path
	return m.GetBinaryPath(ctx)
}

// NeedsUpdate checks if the binary needs to be updated.
// For Homebrew, this checks if a newer version is available.
func (m *BrewBinaryManager) NeedsUpdate(ctx context.Context) (bool, error) {
	// Check if installed
	installed, err := m.IsInstalled(ctx)
	if err != nil {
		return false, err
	}
	if !installed {
		return true, nil // Not installed = needs "update" (install)
	}

	// Ensure brew is installed
	if err := m.ensureBrewInstalled(ctx); err != nil {
		return false, fmt.Errorf("brew not available: %w", err)
	}

	// Run brew outdated to check if update is available
	cmd := exec.CommandContext(ctx, "brew", "outdated", m.formulaName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// brew outdated can exit with error if formula not found
		return false, nil
	}

	// If output contains the formula name, an update is available
	return strings.Contains(string(output), m.formulaName), nil
}

// Update downloads and installs the latest version of the binary.
func (m *BrewBinaryManager) Update(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure brew is installed
	if err := m.ensureBrewInstalled(ctx); err != nil {
		return fmt.Errorf("brew not available: %w", err)
	}

	// Check if formula is installed
	installed, err := m.isInstalledLocked(ctx)
	if err != nil {
		return fmt.Errorf("failed to check installation status: %w", err)
	}

	if !installed {
		// Install if not present
		m.mu.Unlock()
		err := m.Install(ctx)
		m.mu.Lock()
		return err
	}

	// Upgrade existing installation
	cmd := exec.CommandContext(ctx, "brew", "upgrade", m.formulaName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// brew upgrade might fail if already at latest version
		// This is not really an error
		outputStr := string(output)
		if strings.Contains(outputStr, "already installed") ||
			strings.Contains(outputStr, "already upgraded") {
			return nil
		}
		return fmt.Errorf("brew upgrade failed: %w (output: %s)", err, outputStr)
	}

	// Clear cached data
	m.binaryPath = ""
	m.cachedVersion = ""

	return nil
}

// Verify BrewBinaryManager implements BinaryManager
var _ BinaryManager = (*BrewBinaryManager)(nil)
