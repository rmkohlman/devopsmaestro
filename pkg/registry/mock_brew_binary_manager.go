package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

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

	// Simulate installation
	m.InstalledState = true

	// Create a fake binary file
	binaryPath := filepath.Join(m.StorageDir, "squid")
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(binaryPath)
	if err != nil {
		return fmt.Errorf("failed to create mock binary: %w", err)
	}
	defer f.Close()

	// Make it executable
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

	// Simulate uninstallation
	m.InstalledState = false

	// Remove the fake binary file
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

	// If not installed, install it
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

	// Mock always returns false (no update needed)
	return false, nil
}

// Update updates the binary to the latest version.
func (m *MockBrewBinaryManager) Update(ctx context.Context) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx)
	}

	// Simulate update by reinstalling
	return m.Install(ctx)
}

// Verify MockBrewBinaryManager implements BinaryManager
var _ BinaryManager = (*MockBrewBinaryManager)(nil)
