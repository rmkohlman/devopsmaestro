package db

import (
	"database/sql"
	"io/fs"
	"path/filepath"
	"testing"

	"devopsmaestro/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminalEmulatorMigration(t *testing.T) {
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

	// Run migrations (this should create the terminal_emulators table)
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err)

	// Verify the table exists and has the correct structure by creating an emulator
	emulator := &models.TerminalEmulatorDB{
		Name:        "migration-test",
		Description: sql.NullString{String: "Migration test emulator", Valid: true},
		Type:        "wezterm",
		Config:      `{"theme": "dark"}`,
		ThemeRef:    sql.NullString{String: "dark-theme", Valid: true},
		Category:    sql.NullString{String: "test", Valid: true},
		Labels:      `{"env": "test"}`,
		Workspace:   sql.NullString{String: "test-workspace", Valid: true},
		Enabled:     true,
	}

	// Test all CRUD operations
	err = store.CreateTerminalEmulator(emulator)
	assert.NoError(t, err)
	assert.NotZero(t, emulator.ID)

	// Read back
	retrieved, err := store.GetTerminalEmulator("migration-test")
	assert.NoError(t, err)
	assert.Equal(t, emulator.Name, retrieved.Name)
	assert.Equal(t, emulator.Description, retrieved.Description)
	assert.Equal(t, emulator.Type, retrieved.Type)
	assert.Equal(t, emulator.Config, retrieved.Config)
	assert.Equal(t, emulator.ThemeRef, retrieved.ThemeRef)
	assert.Equal(t, emulator.Labels, retrieved.Labels)

	// Update
	emulator.Description = sql.NullString{String: "Updated description", Valid: true}
	err = store.UpdateTerminalEmulator(emulator)
	assert.NoError(t, err)

	// Verify update
	updated, err := store.GetTerminalEmulator("migration-test")
	assert.NoError(t, err)
	assert.Equal(t, "Updated description", updated.Description.String)

	// List operations
	emulators, err := store.ListTerminalEmulators()
	assert.NoError(t, err)
	assert.Len(t, emulators, 1)

	byType, err := store.ListTerminalEmulatorsByType("wezterm")
	assert.NoError(t, err)
	assert.Len(t, byType, 1)

	byWorkspace, err := store.ListTerminalEmulatorsByWorkspace("test-workspace")
	assert.NoError(t, err)
	assert.Len(t, byWorkspace, 1)

	// Delete
	err = store.DeleteTerminalEmulator("migration-test")
	assert.NoError(t, err)

	// Verify deletion
	_, err = store.GetTerminalEmulator("migration-test")
	assert.Error(t, err)
}
