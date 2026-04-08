package db

import (
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSQLDataStore_MigrationVersion_AfterMigrations verifies that
// MigrationVersion returns the correct version after migrations are applied.
func TestSQLDataStore_MigrationVersion_AfterMigrations(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	cfg := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver, err := NewSQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())
	defer driver.Close()

	store := NewSQLDataStore(driver, nil)

	// Run migrations
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err)

	// Verify MigrationVersion returns a positive version
	version, err := store.MigrationVersion()
	require.NoError(t, err)
	assert.Greater(t, version, 0, "Migration version should be > 0 after running migrations")

	// Verify the version matches the latest migration file
	sqliteFS, err := fs.Sub(migrationsSubFS, "sqlite")
	require.NoError(t, err)
	latestVersion, hasVersion, err := getLatestMigrationVersion(sqliteFS)
	require.NoError(t, err)
	require.True(t, hasVersion)
	assert.Equal(t, int(latestVersion), version,
		"MigrationVersion should match the latest migration file version")
}

// TestSQLDataStore_MigrationVersion_FreshDatabase verifies that
// MigrationVersion returns 0 for a fresh database with no migrations.
func TestSQLDataStore_MigrationVersion_FreshDatabase(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	cfg := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver, err := NewSQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())
	defer driver.Close()

	store := NewSQLDataStore(driver, nil)

	// No migrations applied - should return 0 with no error
	version, err := store.MigrationVersion()
	require.NoError(t, err)
	assert.Equal(t, 0, version, "Fresh database should report version 0")
}

// TestMockDataStore_MigrationVersion verifies the mock implementation.
func TestMockDataStore_MigrationVersion(t *testing.T) {
	store := NewMockDataStore()

	// Default value is 0
	version, err := store.MigrationVersion()
	require.NoError(t, err)
	assert.Equal(t, 0, version)

	// Set a custom version
	store.MigrationVersionValue = 18
	version, err = store.MigrationVersion()
	require.NoError(t, err)
	assert.Equal(t, 18, version)

	// Error injection
	store.MigrationVersionErr = assert.AnError
	_, err = store.MigrationVersion()
	assert.Equal(t, assert.AnError, err)
}
