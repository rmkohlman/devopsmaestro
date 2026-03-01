package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// MockPipxBinaryManager is a mock implementation of BinaryManager for testing.
// It doesn't actually install packages, but creates fake binaries for testing.
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
// The storage parameter is used as the bin directory for the mock.
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

	// Create fake binary
	if m.binaryPath == "" {
		binPath := filepath.Join(m.binDir, "devpi-server")

		// Ensure directory exists
		if err := os.MkdirAll(m.binDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create bin directory: %w", err)
		}

		// Create fake binary file
		f, err := os.Create(binPath)
		if err != nil {
			return "", fmt.Errorf("failed to create fake binary: %w", err)
		}
		f.Close()

		// Make it executable
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
