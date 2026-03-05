package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// =============================================================================
// MockBinaryManager — consolidated mock for BinaryManager interface
// =============================================================================

// MockBinaryManager is a mock implementation of BinaryManager for testing.
// It supports all 5 registry types via configurable binary name.
type MockBinaryManager struct {
	binDir     string
	version    string
	binaryName string

	// Hooks for customizing behavior in tests
	EnsureBinaryFunc func(ctx context.Context) (string, error)
	GetVersionFunc   func(ctx context.Context) (string, error)
	NeedsUpdateFunc  func(ctx context.Context) (bool, error)
	UpdateFunc       func(ctx context.Context) error
}

// NewMockBinaryManager creates a MockBinaryManager for testing.
// The binaryName defaults to "athens" for backward compatibility.
func NewMockBinaryManager(binDir, version string) *MockBinaryManager {
	return &MockBinaryManager{
		binDir:     binDir,
		version:    version,
		binaryName: "athens",
	}
}

// NewMockBinaryManagerNamed creates a MockBinaryManager with an explicit binary name.
func NewMockBinaryManagerNamed(binDir, version, binaryName string) *MockBinaryManager {
	return &MockBinaryManager{
		binDir:     binDir,
		version:    version,
		binaryName: binaryName,
	}
}

// EnsureBinary creates a fake binary file for testing.
func (m *MockBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	if m.EnsureBinaryFunc != nil {
		return m.EnsureBinaryFunc(ctx)
	}

	binaryPath := filepath.Join(m.binDir, m.binaryName)

	// Check if already exists
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	// Create directory
	if err := os.MkdirAll(m.binDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Create fake executable script that handles version command
	script := fmt.Sprintf(`#!/bin/bash
if [ "$1" = "--version" ] || [ "$1" = "version" ]; then
    echo "%s v%s"
    exit 0
fi
# For serve/start command, sleep forever to simulate running server
sleep infinity
`, m.binaryName, m.version)

	if err := os.WriteFile(binaryPath, []byte(script), 0755); err != nil {
		return "", fmt.Errorf("failed to create fake binary: %w", err)
	}

	return binaryPath, nil
}

// GetVersion returns the mock version.
func (m *MockBinaryManager) GetVersion(ctx context.Context) (string, error) {
	if m.GetVersionFunc != nil {
		return m.GetVersionFunc(ctx)
	}

	return m.version, nil
}

// NeedsUpdate always returns false for mock.
func (m *MockBinaryManager) NeedsUpdate(ctx context.Context) (bool, error) {
	if m.NeedsUpdateFunc != nil {
		return m.NeedsUpdateFunc(ctx)
	}

	return false, nil
}

// Update is a no-op for mock.
func (m *MockBinaryManager) Update(ctx context.Context) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx)
	}

	return nil
}

// =============================================================================
// MockBrewBinaryManager — Brew-specific mock with extra Install/Uninstall methods
// =============================================================================

// MockBrewBinaryManager is a mock implementation of BinaryManager for testing.
// It simulates Homebrew operations without actually calling brew.
type MockBrewBinaryManager struct {
	// StorageDir is where the mock binary is stored
	StorageDir string

	// Version is the mock version to return
	Version string

	// InstalledState tracks whether the binary is "installed"
	InstalledState bool

	// Custom function overrides for testing
	InstallFunc       func(ctx context.Context) error
	UninstallFunc     func(ctx context.Context) error
	IsInstalledFunc   func(ctx context.Context) (bool, error)
	GetVersionFunc    func(ctx context.Context) (string, error)
	GetBinaryPathFunc func(ctx context.Context) (string, error)
	EnsureBinaryFunc  func(ctx context.Context) (string, error)
	NeedsUpdateFunc   func(ctx context.Context) (bool, error)
	UpdateFunc        func(ctx context.Context) error
}

// NewMockBrewBinaryManager creates a new mock binary manager.
func NewMockBrewBinaryManager(storageDir string, version string) *MockBrewBinaryManager {
	return &MockBrewBinaryManager{
		StorageDir:     storageDir,
		Version:        version,
		InstalledState: true, // Default to installed
	}
}

// Install simulates installing the binary via Homebrew.
func (m *MockBrewBinaryManager) Install(ctx context.Context) error {
	if m.InstallFunc != nil {
		return m.InstallFunc(ctx)
	}

	m.InstalledState = true

	binaryPath := filepath.Join(m.StorageDir, "squid")
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to create mock binary: %w", err)
	}
	defer f.Close()

	if err := os.Chmod(binaryPath, 0755); err != nil {
		return fmt.Errorf("failed to make binary executable: %w", err)
	}

	return nil
}

// Uninstall simulates uninstalling the binary via Homebrew.
func (m *MockBrewBinaryManager) Uninstall(ctx context.Context) error {
	if m.UninstallFunc != nil {
		return m.UninstallFunc(ctx)
	}

	m.InstalledState = false

	binaryPath := filepath.Join(m.StorageDir, "squid")
	if err := os.Remove(binaryPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove mock binary: %w", err)
	}

	return nil
}

// IsInstalled checks if the binary is "installed".
func (m *MockBrewBinaryManager) IsInstalled(ctx context.Context) (bool, error) {
	if m.IsInstalledFunc != nil {
		return m.IsInstalledFunc(ctx)
	}

	return m.InstalledState, nil
}

// GetVersion returns the mock version.
func (m *MockBrewBinaryManager) GetVersion(ctx context.Context) (string, error) {
	if m.GetVersionFunc != nil {
		return m.GetVersionFunc(ctx)
	}

	if !m.InstalledState {
		return "", fmt.Errorf("binary not installed")
	}

	return m.Version, nil
}

// GetBinaryPath returns the path to the mock binary.
func (m *MockBrewBinaryManager) GetBinaryPath(ctx context.Context) (string, error) {
	if m.GetBinaryPathFunc != nil {
		return m.GetBinaryPathFunc(ctx)
	}

	if !m.InstalledState {
		return "", fmt.Errorf("binary not installed")
	}

	return filepath.Join(m.StorageDir, "squid"), nil
}

// EnsureBinary ensures the binary exists, installing if necessary.
func (m *MockBrewBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	if m.EnsureBinaryFunc != nil {
		return m.EnsureBinaryFunc(ctx)
	}

	if !m.InstalledState {
		if err := m.Install(ctx); err != nil {
			return "", err
		}
	}

	return m.GetBinaryPath(ctx)
}

// NeedsUpdate checks if the binary needs to be updated.
func (m *MockBrewBinaryManager) NeedsUpdate(ctx context.Context) (bool, error) {
	if m.NeedsUpdateFunc != nil {
		return m.NeedsUpdateFunc(ctx)
	}

	return false, nil
}

// Update updates the binary to the latest version.
func (m *MockBrewBinaryManager) Update(ctx context.Context) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx)
	}

	return m.Install(ctx)
}

// Verify MockBrewBinaryManager implements BinaryManager
var _ BinaryManager = (*MockBrewBinaryManager)(nil)

// =============================================================================
// MockNpmBinaryManager — npm-specific mock (for Verdaccio)
// =============================================================================

// MockNpmBinaryManager is a mock implementation of BinaryManager for testing.
type MockNpmBinaryManager struct {
	packageName string
	version     string
	binDir      string
	binaryPath  string

	// Function overrides for custom behavior
	EnsureBinaryFunc func(ctx context.Context) (string, error)
	GetVersionFunc   func(ctx context.Context) (string, error)
	NeedsUpdateFunc  func(ctx context.Context) (bool, error)
	UpdateFunc       func(ctx context.Context) error
}

// NewMockNpmBinaryManager creates a new MockNpmBinaryManager.
func NewMockNpmBinaryManager(storage, version string) *MockNpmBinaryManager {
	return &MockNpmBinaryManager{
		packageName: "verdaccio",
		version:     version,
		binDir:      filepath.Join(storage, "bin"),
	}
}

// EnsureBinary creates a fake binary for testing.
func (m *MockNpmBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	if m.EnsureBinaryFunc != nil {
		return m.EnsureBinaryFunc(ctx)
	}

	if m.binaryPath == "" {
		binPath := filepath.Join(m.binDir, "verdaccio")

		if err := os.MkdirAll(m.binDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create bin directory: %w", err)
		}

		f, err := os.Create(binPath)
		if err != nil {
			return "", fmt.Errorf("failed to create fake binary: %w", err)
		}
		f.Close()

		if err := os.Chmod(binPath, 0755); err != nil {
			return "", fmt.Errorf("failed to make binary executable: %w", err)
		}

		m.binaryPath = binPath
	}

	return m.binaryPath, nil
}

// GetVersion returns the configured version.
func (m *MockNpmBinaryManager) GetVersion(ctx context.Context) (string, error) {
	if m.GetVersionFunc != nil {
		return m.GetVersionFunc(ctx)
	}
	return m.version, nil
}

// NeedsUpdate always returns false for the mock.
func (m *MockNpmBinaryManager) NeedsUpdate(ctx context.Context) (bool, error) {
	if m.NeedsUpdateFunc != nil {
		return m.NeedsUpdateFunc(ctx)
	}
	return false, nil
}

// Update is a no-op for the mock.
func (m *MockNpmBinaryManager) Update(ctx context.Context) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx)
	}
	return nil
}

// =============================================================================
// MockPipxBinaryManager — pipx-specific mock (for Devpi)
// =============================================================================

// MockPipxBinaryManager is a mock implementation of BinaryManager for testing.
type MockPipxBinaryManager struct {
	packageName string
	version     string
	binDir      string
	binaryPath  string

	// Function overrides for custom behavior
	EnsureBinaryFunc func(ctx context.Context) (string, error)
	GetVersionFunc   func(ctx context.Context) (string, error)
	NeedsUpdateFunc  func(ctx context.Context) (bool, error)
	UpdateFunc       func(ctx context.Context) error
}

// NewMockPipxBinaryManager creates a new MockPipxBinaryManager.
func NewMockPipxBinaryManager(storage, version string) *MockPipxBinaryManager {
	return &MockPipxBinaryManager{
		packageName: "devpi-server",
		version:     version,
		binDir:      filepath.Join(storage, "bin"),
	}
}

// EnsureBinary creates a fake binary for testing.
func (m *MockPipxBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	if m.EnsureBinaryFunc != nil {
		return m.EnsureBinaryFunc(ctx)
	}

	if m.binaryPath == "" {
		binPath := filepath.Join(m.binDir, "devpi-server")

		if err := os.MkdirAll(m.binDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create bin directory: %w", err)
		}

		f, err := os.Create(binPath)
		if err != nil {
			return "", fmt.Errorf("failed to create fake binary: %w", err)
		}
		f.Close()

		if err := os.Chmod(binPath, 0755); err != nil {
			return "", fmt.Errorf("failed to make binary executable: %w", err)
		}

		m.binaryPath = binPath
	}

	return m.binaryPath, nil
}

// GetVersion returns the configured version.
func (m *MockPipxBinaryManager) GetVersion(ctx context.Context) (string, error) {
	if m.GetVersionFunc != nil {
		return m.GetVersionFunc(ctx)
	}
	return m.version, nil
}

// NeedsUpdate always returns false for the mock.
func (m *MockPipxBinaryManager) NeedsUpdate(ctx context.Context) (bool, error) {
	if m.NeedsUpdateFunc != nil {
		return m.NeedsUpdateFunc(ctx)
	}
	return false, nil
}

// Update is a no-op for the mock.
func (m *MockPipxBinaryManager) Update(ctx context.Context) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx)
	}
	return nil
}

// =============================================================================
// MockGoModuleProxy — mock for GoModuleProxy interface
// =============================================================================

// MockGoModuleProxy is a mock implementation of GoModuleProxy for testing.
type MockGoModuleProxy struct {
	StartFunc         func(ctx context.Context) error
	StopFunc          func(ctx context.Context) error
	StatusFunc        func(ctx context.Context) (*GoModuleProxyStatus, error)
	EnsureRunningFunc func(ctx context.Context) error
	IsRunningFunc     func(ctx context.Context) bool
	GetEndpointFunc   func() string
	GetGoEnvFunc      func() map[string]string
}

// Start calls the mock's StartFunc.
func (m *MockGoModuleProxy) Start(ctx context.Context) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx)
	}
	return nil
}

// Stop calls the mock's StopFunc.
func (m *MockGoModuleProxy) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	return nil
}

// Status calls the mock's StatusFunc.
func (m *MockGoModuleProxy) Status(ctx context.Context) (*GoModuleProxyStatus, error) {
	if m.StatusFunc != nil {
		return m.StatusFunc(ctx)
	}
	return &GoModuleProxyStatus{}, nil
}

// EnsureRunning calls the mock's EnsureRunningFunc.
func (m *MockGoModuleProxy) EnsureRunning(ctx context.Context) error {
	if m.EnsureRunningFunc != nil {
		return m.EnsureRunningFunc(ctx)
	}
	return nil
}

// IsRunning calls the mock's IsRunningFunc.
func (m *MockGoModuleProxy) IsRunning(ctx context.Context) bool {
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc(ctx)
	}
	return false
}

// GetEndpoint calls the mock's GetEndpointFunc.
func (m *MockGoModuleProxy) GetEndpoint() string {
	if m.GetEndpointFunc != nil {
		return m.GetEndpointFunc()
	}
	return "http://localhost:3000"
}

// GetGoEnv calls the mock's GetGoEnvFunc.
func (m *MockGoModuleProxy) GetGoEnv() map[string]string {
	if m.GetGoEnvFunc != nil {
		return m.GetGoEnvFunc()
	}
	return map[string]string{
		"GOPROXY": "http://localhost:3000",
	}
}
