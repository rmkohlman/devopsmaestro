package sync

import (
	"testing"

	nvimpackage "devopsmaestro/pkg/nvimops/package"

	"github.com/stretchr/testify/assert"
)

// TestPackageCreatorIntegration verifies that SyncResult properly tracks package operations
func TestPackageCreatorIntegration(t *testing.T) {
	result := &SyncResult{
		SourceName: "test-source",
	}

	// Test adding packages
	assert.Empty(t, result.PackagesCreated)
	assert.Empty(t, result.PackagesUpdated)

	result.AddPackageCreated("source-package")
	assert.Contains(t, result.PackagesCreated, "source-package")
	assert.Len(t, result.PackagesCreated, 1)

	result.AddPackageUpdated("existing-package")
	assert.Contains(t, result.PackagesUpdated, "existing-package")
	assert.Len(t, result.PackagesUpdated, 1)

	// Test multiple packages
	result.AddPackageCreated("another-package")
	assert.Len(t, result.PackagesCreated, 2)
	assert.Contains(t, result.PackagesCreated, "another-package")
}

// TestSyncOptionsWithPackageCreator verifies that SyncOptions can hold a PackageCreator
func TestSyncOptionsWithPackageCreator(t *testing.T) {
	// Create a mock package creator
	var mockCreator MockPackageCreator

	options := NewSyncOptions().
		WithPackageCreator(&mockCreator).
		Build()

	assert.NotNil(t, options.PackageCreator)
	assert.Equal(t, &mockCreator, options.PackageCreator)
}

// MockPackageCreator implements PackageCreator for testing
type MockPackageCreator struct {
	CreatedPackages []string
	CreatedPlugins  []string
}

func (m *MockPackageCreator) CreatePackage(sourceName string, plugins []string) error {
	m.CreatedPackages = append(m.CreatedPackages, sourceName)
	m.CreatedPlugins = append(m.CreatedPlugins, plugins...)
	return nil
}

// Verify MockPackageCreator implements PackageCreator
var _ PackageCreator = (*MockPackageCreator)(nil)

// TestFilePackageCreatorInterface verifies that FilePackageCreator implements PackageCreator
func TestFilePackageCreatorInterface(t *testing.T) {
	creator := nvimpackage.NewFilePackageCreator("/tmp")

	// This should compile - verifies interface compliance
	var _ PackageCreator = creator

	assert.NotNil(t, creator)
}
