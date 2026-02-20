package sources

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	nvimpackage "devopsmaestro/pkg/nvimops/package"
	"devopsmaestro/pkg/nvimops/sync"
	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLazyVimHandlerAutoPackageCreation(t *testing.T) {
	// Create temporary directories for testing
	tempDir, err := os.MkdirTemp("", "nvp-sync-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	pluginsDir := filepath.Join(tempDir, "plugins")
	packagesDir := filepath.Join(tempDir, "packages")

	handler := NewLazyVimHandler()

	t.Run("CreatesPackageAfterSync", func(t *testing.T) {
		// Create package creator
		packageCreator := nvimpackage.NewFilePackageCreator(packagesDir)

		// Build sync options
		options := sync.NewSyncOptions().
			DryRun(false).
			WithTargetDir(pluginsDir).
			WithPackageCreator(packageCreator).
			Build()

		// Perform sync (this will make real HTTP requests to LazyVim)
		result, err := handler.Sync(context.Background(), options)
		require.NoError(t, err)
		assert.NotNil(t, result)

		// Check that plugins were synced (at least some should be available)
		if result.TotalAvailable > 0 {
			// Verify plugin files were created
			pluginFiles, err := os.ReadDir(pluginsDir)
			require.NoError(t, err)
			assert.Greater(t, len(pluginFiles), 0, "Should have created plugin files")

			// Verify package file was created
			packageFile := filepath.Join(packagesDir, "lazyvim.yaml")
			_, err = os.Stat(packageFile)
			assert.NoError(t, err, "Package file should exist")

			// Verify package was added to result
			assert.Contains(t, result.PackagesCreated, "lazyvim")

			// Read and validate package content
			packageData, err := os.ReadFile(packageFile)
			require.NoError(t, err)

			pkg, err := nvimpackage.ParseYAML(packageData)
			require.NoError(t, err)

			assert.Equal(t, "lazyvim", pkg.Name)
			assert.Contains(t, pkg.Description, "lazyvim source sync")
			assert.Equal(t, "source-sync", pkg.Category)
			assert.True(t, pkg.Enabled)
			assert.Greater(t, len(pkg.Plugins), 0, "Package should contain plugins")
		}
	})

	t.Run("DryRunMode", func(t *testing.T) {
		// Clean up directories
		os.RemoveAll(pluginsDir)
		os.RemoveAll(packagesDir)

		packageCreator := nvimpackage.NewFilePackageCreator(packagesDir)

		options := sync.NewSyncOptions().
			DryRun(true).
			WithTargetDir(pluginsDir).
			WithPackageCreator(packageCreator).
			Build()

		result, err := handler.Sync(context.Background(), options)
		require.NoError(t, err)

		// In dry run mode, no actual files should be created
		_, err = os.Stat(pluginsDir)
		assert.True(t, os.IsNotExist(err), "Plugins directory should not exist in dry run")

		_, err = os.Stat(packagesDir)
		assert.True(t, os.IsNotExist(err), "Packages directory should not exist in dry run")

		// But the result should still show what would be created
		if result.TotalAvailable > 0 {
			assert.Greater(t, len(result.PluginsCreated), 0)
			assert.Contains(t, result.PackagesCreated, "lazyvim")
		}
	})

	t.Run("WithoutPackageCreator", func(t *testing.T) {
		// Test that sync works without package creator (backward compatibility)
		options := sync.NewSyncOptions().
			DryRun(true).
			WithTargetDir(pluginsDir).
			Build()

		result, err := handler.Sync(context.Background(), options)
		require.NoError(t, err)

		// Should not have any packages created
		assert.Empty(t, result.PackagesCreated)
		assert.Empty(t, result.PackagesUpdated)
	})
}

// TestFilePackageCreator tests the FilePackageCreator directly
func TestFilePackageCreator(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "nvp-package-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	creator := nvimpackage.NewFilePackageCreator(tempDir)

	t.Run("CreatePackage", func(t *testing.T) {
		plugins := []string{"telescope", "treesitter", "lsp-config"}
		err := creator.CreatePackage("test-source", plugins)
		require.NoError(t, err)

		// Verify file was created
		packageFile := filepath.Join(tempDir, "test-source.yaml")
		_, err = os.Stat(packageFile)
		assert.NoError(t, err)

		// Read and validate content
		packageData, err := os.ReadFile(packageFile)
		require.NoError(t, err)

		pkg, err := nvimpackage.ParseYAML(packageData)
		require.NoError(t, err)

		assert.Equal(t, "test-source", pkg.Name)
		assert.Contains(t, pkg.Description, "test-source source sync")
		assert.Equal(t, "source-sync", pkg.Category)
		assert.True(t, pkg.Enabled)
		assert.Equal(t, plugins, pkg.Plugins)
		// Note: CreatedAt and UpdatedAt are not serialized to YAML (yaml:"-" tag)
		// so we don't check them here

		// Verify the YAML structure is kubectl-style by parsing raw YAML
		var packageYAMLRaw nvimpackage.PackageYAML
		err = yaml.Unmarshal(packageData, &packageYAMLRaw)
		require.NoError(t, err)

		assert.Equal(t, "devopsmaestro.io/v1", packageYAMLRaw.APIVersion)
		assert.Equal(t, "NvimPackage", packageYAMLRaw.Kind)
		assert.Equal(t, "test-source", packageYAMLRaw.Metadata.Name)
		assert.NotNil(t, packageYAMLRaw.Metadata.Labels)
		assert.Equal(t, "test-source", packageYAMLRaw.Metadata.Labels["source"])
		assert.Equal(t, "true", packageYAMLRaw.Metadata.Labels["auto-generated"])
	})

	t.Run("InvalidDirectory", func(t *testing.T) {
		invalidCreator := nvimpackage.NewFilePackageCreator("")
		err := invalidCreator.CreatePackage("test", []string{"plugin1"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "packages directory not set")
	})
}
