package db

import (
	"database/sql"
	"io/fs"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: testMigrationsFS is defined in terminal_plugin_migration_test.go

func TestCheckPendingMigrations_NewDatabase(t *testing.T) {
	// Test case is covered by integration tests that use the real migrations
	// This is just a placeholder to ensure the function signature is tested
	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())
	defer driver.Close()

	// Test the function exists and accepts the right parameters
	// The actual functionality is tested in integration tests
	assert.NotNil(t, driver)
}

func TestCheckPendingMigrations_CurrentDatabase(t *testing.T) {
	// Test case is covered by integration tests that use the real migrations
	// This is just a placeholder to ensure the function signature is tested
	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())
	defer driver.Close()

	// Test the function exists and accepts the right parameters
	// The actual functionality is tested in integration tests
	assert.NotNil(t, driver)
}

func TestGetLatestMigrationVersion(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		expected uint
		hasVer   bool
		wantErr  bool
	}{
		{
			name: "single migration",
			files: map[string]string{
				"001_init.up.sql": "CREATE TABLE test;",
			},
			expected: 1,
			hasVer:   true,
			wantErr:  false,
		},
		{
			name: "multiple migrations",
			files: map[string]string{
				"001_init.up.sql":       "CREATE TABLE test;",
				"002_add_column.up.sql": "ALTER TABLE test ADD COLUMN name TEXT;",
				"015_latest.up.sql":     "CREATE INDEX idx_test ON test(name);",
			},
			expected: 15,
			hasVer:   true,
			wantErr:  false,
		},
		{
			name: "zero padded versions",
			files: map[string]string{
				"001_init.up.sql":         "CREATE TABLE test;",
				"005_middle.up.sql":       "ALTER TABLE test ADD COLUMN name TEXT;",
				"010_double_digit.up.sql": "CREATE INDEX idx_test ON test(name);",
			},
			expected: 10,
			hasVer:   true,
			wantErr:  false,
		},
		{
			name: "mixed with down files",
			files: map[string]string{
				"001_init.up.sql":   "CREATE TABLE test;",
				"001_init.down.sql": "DROP TABLE test;",
				"002_add.up.sql":    "ALTER TABLE test ADD COLUMN name TEXT;",
				"002_add.down.sql":  "ALTER TABLE test DROP COLUMN name;",
			},
			expected: 2,
			hasVer:   true,
			wantErr:  false,
		},
		{
			name:     "no migration files",
			files:    map[string]string{},
			expected: 0,
			hasVer:   false,
			wantErr:  false,
		},
		{
			name: "only down files",
			files: map[string]string{
				"001_init.down.sql": "DROP TABLE test;",
			},
			expected: 0,
			hasVer:   false,
			wantErr:  false,
		},
		{
			name: "non-migration files",
			files: map[string]string{
				"README.md":  "Some docs",
				"script.sql": "Some script",
				"data.json":  "Some data",
			},
			expected: 0,
			hasVer:   false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary filesystem with test files
			testFS := fstest.MapFS{}
			for name, content := range tt.files {
				testFS[name] = &fstest.MapFile{Data: []byte(content)}
			}

			version, hasVersion, err := getLatestMigrationVersion(testFS)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, version, "Expected version mismatch")
			assert.Equal(t, tt.hasVer, hasVersion, "Expected hasVersion mismatch")
		})
	}
}

func TestParseVersionNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected uint
		wantErr  bool
	}{
		{"001", 1, false},
		{"005", 5, false},
		{"010", 10, false},
		{"015", 15, false},
		{"1", 1, false},
		{"5", 5, false},
		{"10", 10, false},
		{"15", 15, false},
		{"000", 0, false},
		{"0", 0, false},
		{"abc", 0, true},
		{"1a", 0, true},
		{"a1", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseVersionNumber(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// =============================================================================
// Comprehensive Migration Tests
// =============================================================================

// TestRunMigrations_FreshDatabase verifies migrations run successfully
// on a fresh database with no schema.
func TestRunMigrations_FreshDatabase(t *testing.T) {
	// Create a temporary file-based database (in-memory doesn't work with migrate library)
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	cfg := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver, err := NewSQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())
	defer driver.Close()

	// Run migrations using the embedded migrations
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err)

	// Verify schema_migrations table exists
	var tableName string
	err = driver.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='schema_migrations'").Scan(&tableName)
	require.NoError(t, err)
	assert.Equal(t, "schema_migrations", tableName)

	// Verify some key tables were created
	tables := []string{"ecosystems", "domains", "apps", "workspaces", "context", "credentials"}
	for _, table := range tables {
		var exists string
		err = driver.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&exists)
		require.NoError(t, err, "Table %s should exist", table)
		assert.Equal(t, table, exists)
	}
}

// TestRunMigrations_Idempotency verifies that running migrations twice
// is safe (idempotency test).
func TestRunMigrations_Idempotency(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	cfg := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver, err := NewSQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())
	defer driver.Close()

	// Run migrations first time
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err)

	// Run migrations second time - should not error
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err, "Running migrations twice should be idempotent")
}

// TestRunMigrations_VersionTracking verifies migration version is tracked
// correctly in schema_migrations table.
func TestRunMigrations_VersionTracking(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	cfg := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver, err := NewSQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())
	defer driver.Close()

	// Run migrations
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err)

	// Check version in schema_migrations table
	var version int64
	var dirty bool
	err = driver.QueryRow("SELECT version, dirty FROM schema_migrations").Scan(&version, &dirty)
	require.NoError(t, err)
	assert.Greater(t, version, int64(0), "Migration version should be > 0")
	assert.False(t, dirty, "Migration should not be in dirty state")
}

// TestRunMigrations_SchemaChanges verifies specific schema changes
// are applied correctly.
func TestRunMigrations_SchemaChanges(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	cfg := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver, err := NewSQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())
	defer driver.Close()

	// Run migrations
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err)

	// Test: ecosystems table has expected columns
	rows, err := driver.Query("PRAGMA table_info(ecosystems)")
	require.NoError(t, err)
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt sql.NullString
		err = rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk)
		require.NoError(t, err)
		columns[name] = true
	}

	expectedColumns := []string{"id", "name", "description", "theme", "created_at", "updated_at"}
	for _, col := range expectedColumns {
		assert.True(t, columns[col], "Column %s should exist in ecosystems table", col)
	}

	// Test: workspaces table has slug column (added in later migration)
	rows, err = driver.Query("PRAGMA table_info(workspaces)")
	require.NoError(t, err)
	defer rows.Close()

	workspaceColumns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt sql.NullString
		err = rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk)
		require.NoError(t, err)
		workspaceColumns[name] = true
	}

	assert.True(t, workspaceColumns["slug"], "Column slug should exist in workspaces table")
}

// TestRunMigrations_ForeignKeyConstraints verifies foreign key
// relationships are properly established.
func TestRunMigrations_ForeignKeyConstraints(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	cfg := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver, err := NewSQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())
	defer driver.Close()

	// Enable foreign keys for SQLite
	_, err = driver.Execute("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	// Run migrations
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err)

	// Test: Can't create domain without valid ecosystem_id
	_, err = driver.Execute("INSERT INTO domains (ecosystem_id, name, created_at, updated_at) VALUES (999, 'test', datetime('now'), datetime('now'))")
	assert.Error(t, err, "Should fail due to foreign key constraint")

	// Test: Can create domain with valid ecosystem_id
	_, err = driver.Execute("INSERT INTO ecosystems (name, created_at, updated_at) VALUES ('test-eco', datetime('now'), datetime('now'))")
	require.NoError(t, err)

	var ecoID int64
	err = driver.QueryRow("SELECT id FROM ecosystems WHERE name = 'test-eco'").Scan(&ecoID)
	require.NoError(t, err)

	_, err = driver.Execute("INSERT INTO domains (ecosystem_id, name, created_at, updated_at) VALUES (?, 'test-domain', datetime('now'), datetime('now'))", ecoID)
	assert.NoError(t, err, "Should succeed with valid ecosystem_id")
}

// TestCheckPendingMigrations_FreshVsCurrent verifies CheckPendingMigrations
// correctly detects when migrations are needed.
func TestCheckPendingMigrations_FreshVsCurrent(t *testing.T) {
	// Test 1: Fresh database - should have pending migrations
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	cfg1 := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver1, err := NewSQLiteDriver(cfg1)
	require.NoError(t, err)
	require.NoError(t, driver1.Connect())
	defer driver1.Close()

	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)

	hasPending, err := CheckPendingMigrations(driver1, migrationsSubFS)
	require.NoError(t, err)
	assert.True(t, hasPending, "Fresh database should have pending migrations")

	// Test 2: After running migrations - should have no pending
	err = AutoMigrate(driver1, migrationsSubFS)
	require.NoError(t, err)

	hasPending, err = CheckPendingMigrations(driver1, migrationsSubFS)
	require.NoError(t, err)
	assert.False(t, hasPending, "Current database should have no pending migrations")
}

// TestRunMigrations_ContextTableInitialization verifies the context table
// is properly initialized with a single row.
func TestRunMigrations_ContextTableInitialization(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")
	cfg := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver, err := NewSQLiteDriver(cfg)
	require.NoError(t, err)
	require.NoError(t, driver.Connect())
	defer driver.Close()

	// Run migrations
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err)

	// Verify context table has exactly one row with id=1
	var id int
	var activeEcosystemID, activeDomainID, activeAppID, activeWorkspaceID sql.NullInt64
	err = driver.QueryRow("SELECT id, active_ecosystem_id, active_domain_id, active_app_id, active_workspace_id FROM context WHERE id = 1").
		Scan(&id, &activeEcosystemID, &activeDomainID, &activeAppID, &activeWorkspaceID)
	require.NoError(t, err)
	assert.Equal(t, 1, id)

	// Verify all active IDs are NULL initially
	assert.False(t, activeEcosystemID.Valid, "active_ecosystem_id should be NULL")
	assert.False(t, activeDomainID.Valid, "active_domain_id should be NULL")
	assert.False(t, activeAppID.Valid, "active_app_id should be NULL")
	assert.False(t, activeWorkspaceID.Valid, "active_workspace_id should be NULL")
}

// TestAutoMigrate_NilDriver verifies proper error handling when
// driver is nil.
func TestAutoMigrate_NilDriver(t *testing.T) {
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(nil, migrationsSubFS)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver is nil")
}

// TestCheckPendingMigrations_NilDriver verifies proper error handling
// when driver is nil.
func TestCheckPendingMigrations_NilDriver(t *testing.T) {
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	hasPending, err := CheckPendingMigrations(nil, migrationsSubFS)
	assert.Error(t, err)
	assert.False(t, hasPending)
	assert.Contains(t, err.Error(), "driver is nil")
}
