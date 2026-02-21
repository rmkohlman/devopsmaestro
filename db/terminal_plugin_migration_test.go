package db

import (
	"database/sql"
	"embed"
	"io/fs"
	"path/filepath"
	"testing"

	"devopsmaestro/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed migrations/*
var testMigrationsFS embed.FS

func TestTerminalPluginMigration(t *testing.T) {
	// Create a temporary database file
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	// Create a new data store (which will run migrations)
	cfg := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver, err := NewSQLiteDriver(cfg)
	require.NoError(t, err)

	err = driver.Connect()
	require.NoError(t, err)
	defer driver.Close()

	store := &SQLDataStore{
		driver:       driver,
		queryBuilder: NewSQLiteQueryBuilder(),
	}

	// Run migrations (this should create the terminal_plugins table)
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err)

	// Verify the table exists and has the correct structure by creating a plugin
	plugin := &models.TerminalPluginDB{
		Name:         "migration-test",
		Description:  sql.NullString{String: "Migration test plugin", Valid: true},
		Repo:         "test/migration-test",
		Category:     sql.NullString{String: "test", Valid: true},
		Shell:        "zsh",
		Manager:      "manual",
		LoadCommand:  sql.NullString{String: "echo 'test'", Valid: true},
		Dependencies: `["curl"]`,
		EnvVars:      `{"TEST": "value"}`,
		Labels:       `{"env": "test"}`,
		Enabled:      true,
	}

	// Test all CRUD operations
	err = store.CreateTerminalPlugin(plugin)
	assert.NoError(t, err)
	assert.NotZero(t, plugin.ID)

	// Read back
	retrieved, err := store.GetTerminalPlugin("migration-test")
	assert.NoError(t, err)
	assert.Equal(t, plugin.Name, retrieved.Name)
	assert.Equal(t, plugin.Description, retrieved.Description)
	assert.Equal(t, plugin.Repo, retrieved.Repo)
	assert.Equal(t, plugin.Dependencies, retrieved.Dependencies)
	assert.Equal(t, plugin.EnvVars, retrieved.EnvVars)
	assert.Equal(t, plugin.Labels, retrieved.Labels)

	// Update
	plugin.Description = sql.NullString{String: "Updated description", Valid: true}
	err = store.UpdateTerminalPlugin(plugin)
	assert.NoError(t, err)

	// Verify update
	updated, err := store.GetTerminalPlugin("migration-test")
	assert.NoError(t, err)
	assert.Equal(t, "Updated description", updated.Description.String)

	// List operations
	plugins, err := store.ListTerminalPlugins()
	assert.NoError(t, err)
	assert.Len(t, plugins, 1)

	byCategory, err := store.ListTerminalPluginsByCategory("test")
	assert.NoError(t, err)
	assert.Len(t, byCategory, 1)

	byShell, err := store.ListTerminalPluginsByShell("zsh")
	assert.NoError(t, err)
	assert.Len(t, byShell, 1)

	byManager, err := store.ListTerminalPluginsByManager("manual")
	assert.NoError(t, err)
	assert.Len(t, byManager, 1)

	// Delete
	err = store.DeleteTerminalPlugin("migration-test")
	assert.NoError(t, err)

	// Verify deletion
	_, err = store.GetTerminalPlugin("migration-test")
	assert.Error(t, err)
}
