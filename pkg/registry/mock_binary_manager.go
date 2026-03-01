package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// MockBinaryManager is a mock implementation of BinaryManager for testing.
type MockBinaryManager struct {
	binDir  string
	version string

	// Hooks for customizing behavior in tests
	EnsureBinaryFunc func(ctx context.Context) (string, error)
	GetVersionFunc   func(ctx context.Context) (string, error)
	NeedsUpdateFunc  func(ctx context.Context) (bool, error)
	UpdateFunc       func(ctx context.Context) error
}

// NewMockBinaryManager creates a MockBinaryManager for testing.
func NewMockBinaryManager(binDir, version string) *MockBinaryManager {
	return &MockBinaryManager{
		binDir:  binDir,
		version: version,
	}
}

// EnsureBinary creates a fake binary file for testing.
func (m *MockBinaryManager) EnsureBinary(ctx context.Context) (string, error) {
	if m.EnsureBinaryFunc != nil {
		return m.EnsureBinaryFunc(ctx)
	}

	// Default behavior: create a fake binary
	binaryPath := filepath.Join(m.binDir, "zot")

	// Check if already exists
	if _, err := os.Stat(binaryPath); err == nil {
		return binaryPath, nil
	}

	// Create directory
	if err := os.MkdirAll(m.binDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create bin directory: %w", err)
	}

	// Create fake executable script that handles "version" command
	script := fmt.Sprintf(`#!/bin/bash
if [ "$1" = "version" ]; then
    echo "zot v%s"
    exit 0
fi
# For serve command, sleep forever
sleep infinity
`, m.version)

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
