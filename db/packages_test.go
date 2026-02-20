package db

import (
	"database/sql"
	"devopsmaestro/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Simple integration test to verify package operations work correctly
func TestPackageOperations_Integration(t *testing.T) {
	// Use the existing pattern from other tests
	store := NewMockDataStore()

	t.Run("BasicCRUD", func(t *testing.T) {
		// Create
		pkg := &models.NvimPackageDB{
			Name:        "test-package",
			Description: sql.NullString{String: "Test package", Valid: true},
			Category:    sql.NullString{String: "test", Valid: true},
			Plugins:     `["plugin1", "plugin2"]`,
		}

		err := pkg.SetLabels(map[string]string{
			"source": "test",
			"env":    "dev",
		})
		require.NoError(t, err)

		err = store.CreatePackage(pkg)
		require.NoError(t, err)
		assert.NotZero(t, pkg.ID)

		// Read
		retrieved, err := store.GetPackage("test-package")
		require.NoError(t, err)
		assert.Equal(t, "test-package", retrieved.Name)
		assert.Equal(t, "Test package", retrieved.Description.String)
		assert.Equal(t, "test", retrieved.Category.String)

		plugins := retrieved.GetPlugins()
		assert.Equal(t, []string{"plugin1", "plugin2"}, plugins)

		labels := retrieved.GetLabels()
		assert.Equal(t, "test", labels["source"])
		assert.Equal(t, "dev", labels["env"])

		// Update
		retrieved.Description = sql.NullString{String: "Updated package", Valid: true}
		err = retrieved.SetPlugins([]string{"plugin1", "plugin2", "plugin3"})
		require.NoError(t, err)

		err = store.UpdatePackage(retrieved)
		require.NoError(t, err)

		updated, err := store.GetPackage("test-package")
		require.NoError(t, err)
		assert.Equal(t, "Updated package", updated.Description.String)

		updatedPlugins := updated.GetPlugins()
		assert.Contains(t, updatedPlugins, "plugin3")

		// List
		packages, err := store.ListPackages()
		require.NoError(t, err)
		assert.Len(t, packages, 1)

		// List by label
		packages, err = store.ListPackagesByLabel("source", "test")
		require.NoError(t, err)
		assert.Len(t, packages, 1)
		assert.Equal(t, "test-package", packages[0].Name)

		// Delete
		err = store.DeletePackage("test-package")
		require.NoError(t, err)

		_, err = store.GetPackage("test-package")
		assert.Error(t, err)
	})

	t.Run("Upsert", func(t *testing.T) {
		pkg := &models.NvimPackageDB{
			Name:    "upsert-test",
			Plugins: `["initial"]`,
		}

		// First upsert should create
		err := store.UpsertPackage(pkg)
		require.NoError(t, err)
		assert.NotZero(t, pkg.ID)

		// Second upsert should update
		pkg.Description = sql.NullString{String: "Updated via upsert", Valid: true}
		err = pkg.SetPlugins([]string{"updated"})
		require.NoError(t, err)

		err = store.UpsertPackage(pkg)
		require.NoError(t, err)

		retrieved, err := store.GetPackage("upsert-test")
		require.NoError(t, err)
		assert.Equal(t, "Updated via upsert", retrieved.Description.String)

		plugins := retrieved.GetPlugins()
		assert.Equal(t, []string{"updated"}, plugins)
	})
}

func TestNvimPackageDB_HelperMethods(t *testing.T) {
	pkg := &models.NvimPackageDB{}

	t.Run("LabelsHandling", func(t *testing.T) {
		// Empty initially
		labels := pkg.GetLabels()
		assert.Equal(t, map[string]string{}, labels)

		// Set some labels
		testLabels := map[string]string{
			"source":  "lazyvim",
			"version": "15.0",
		}

		err := pkg.SetLabels(testLabels)
		require.NoError(t, err)
		assert.True(t, pkg.Labels.Valid)

		retrieved := pkg.GetLabels()
		assert.Equal(t, testLabels, retrieved)

		// Set empty labels
		err = pkg.SetLabels(map[string]string{})
		require.NoError(t, err)
		assert.False(t, pkg.Labels.Valid)

		empty := pkg.GetLabels()
		assert.Equal(t, map[string]string{}, empty)
	})

	t.Run("PluginsHandling", func(t *testing.T) {
		// Initially empty JSON array
		plugins := pkg.GetPlugins()
		assert.Equal(t, []string{}, plugins)

		// Set some plugins
		testPlugins := []string{"telescope", "treesitter", "nvim-cmp"}

		err := pkg.SetPlugins(testPlugins)
		require.NoError(t, err)

		retrieved := pkg.GetPlugins()
		assert.Equal(t, testPlugins, retrieved)

		// Set empty plugins
		err = pkg.SetPlugins([]string{})
		require.NoError(t, err)

		empty := pkg.GetPlugins()
		assert.Equal(t, []string{}, empty)

		// Set nil plugins
		err = pkg.SetPlugins(nil)
		require.NoError(t, err)

		nilResult := pkg.GetPlugins()
		assert.Equal(t, []string{}, nilResult)
	})
}
