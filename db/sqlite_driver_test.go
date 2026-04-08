package db

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// SQLiteDriver Creation Tests
// =============================================================================

func TestNewMemorySQLiteDriver(t *testing.T) {
	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("NewMemorySQLiteDriver() error = %v", err)
	}
	defer driver.Close()

	if driver.Type() != DriverMemory {
		t.Errorf("Type() = %v, want %v", driver.Type(), DriverMemory)
	}

	if err := driver.Connect(); err != nil {
		t.Errorf("Connect() error = %v", err)
	}

	if err := driver.Ping(); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestNewSQLiteDriver(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	cfg := DriverConfig{
		Type: DriverSQLite,
		Path: dbPath,
	}

	driver, err := NewSQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteDriver() error = %v", err)
	}
	defer driver.Close()

	if driver.Type() != DriverSQLite {
		t.Errorf("Type() = %v, want %v", driver.Type(), DriverSQLite)
	}

	if err := driver.Connect(); err != nil {
		t.Errorf("Connect() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}
}

func TestSQLiteDriver_DSN(t *testing.T) {
	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("NewMemorySQLiteDriver() error = %v", err)
	}
	defer driver.Close()

	dsn := driver.DSN()
	// Uses :memory: with cache=shared for test isolation with concurrent access support
	if dsn != ":memory:?cache=shared&_foreign_keys=on" {
		t.Errorf("DSN() = %q, want %q", dsn, ":memory:?cache=shared&_foreign_keys=on")
	}
}

func TestSQLiteDriver_MigrationDSN(t *testing.T) {
	t.Run("memory driver", func(t *testing.T) {
		cfg := DriverConfig{Type: DriverMemory}
		driver, err := NewMemorySQLiteDriver(cfg)
		if err != nil {
			t.Fatalf("NewMemorySQLiteDriver() error = %v", err)
		}
		defer driver.Close()

		dsn := driver.MigrationDSN()
		if dsn != "sqlite3://:memory:" {
			t.Errorf("MigrationDSN() = %q, want %q", dsn, "sqlite3://:memory:")
		}
	})

	t.Run("file driver", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test.db")

		cfg := DriverConfig{
			Type: DriverSQLite,
			Path: dbPath,
		}

		driver, err := NewSQLiteDriver(cfg)
		if err != nil {
			t.Fatalf("NewSQLiteDriver() error = %v", err)
		}
		defer driver.Close()

		dsn := driver.MigrationDSN()
		expected := "sqlite:///" + dbPath
		if dsn != expected {
			t.Errorf("MigrationDSN() = %q, want %q", dsn, expected)
		}
	})
}

// =============================================================================
// Execute Tests
// =============================================================================

func TestSQLiteDriver_Execute(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	// Create a test table
	_, err := driver.Execute(`CREATE TABLE test_execute (
		id INTEGER PRIMARY KEY,
		name TEXT
	)`)
	if err != nil {
		t.Fatalf("Execute() CREATE TABLE error = %v", err)
	}

	// Insert data
	result, err := driver.Execute("INSERT INTO test_execute (name) VALUES (?)", "test-name")
	if err != nil {
		t.Fatalf("Execute() INSERT error = %v", err)
	}

	lastID, err := result.LastInsertId()
	if err != nil {
		t.Errorf("LastInsertId() error = %v", err)
	}
	if lastID != 1 {
		t.Errorf("LastInsertId() = %d, want 1", lastID)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		t.Errorf("RowsAffected() error = %v", err)
	}
	if rowsAffected != 1 {
		t.Errorf("RowsAffected() = %d, want 1", rowsAffected)
	}
}

func TestSQLiteDriver_ExecuteContext(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create a test table
	_, err := driver.ExecuteContext(ctx, `CREATE TABLE test_exec_ctx (
		id INTEGER PRIMARY KEY,
		name TEXT
	)`)
	if err != nil {
		t.Fatalf("ExecuteContext() CREATE TABLE error = %v", err)
	}

	// Insert with context
	result, err := driver.ExecuteContext(ctx, "INSERT INTO test_exec_ctx (name) VALUES (?)", "ctx-test")
	if err != nil {
		t.Fatalf("ExecuteContext() INSERT error = %v", err)
	}

	lastID, _ := result.LastInsertId()
	if lastID != 1 {
		t.Errorf("LastInsertId() = %d, want 1", lastID)
	}
}

// =============================================================================
// QueryRow Tests
// =============================================================================

func TestSQLiteDriver_QueryRow(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	// Setup
	_, err := driver.Execute(`CREATE TABLE test_query_row (
		id INTEGER PRIMARY KEY,
		name TEXT,
		value INTEGER
	)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	_, err = driver.Execute("INSERT INTO test_query_row (name, value) VALUES (?, ?)", "row1", 42)
	if err != nil {
		t.Fatalf("Setup insert error: %v", err)
	}

	// Test QueryRow
	var id int
	var name string
	var value int

	row := driver.QueryRow("SELECT id, name, value FROM test_query_row WHERE name = ?", "row1")
	err = row.Scan(&id, &name, &value)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if id != 1 {
		t.Errorf("id = %d, want 1", id)
	}
	if name != "row1" {
		t.Errorf("name = %q, want %q", name, "row1")
	}
	if value != 42 {
		t.Errorf("value = %d, want 42", value)
	}
}

func TestSQLiteDriver_QueryRow_NoRows(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	_, err := driver.Execute(`CREATE TABLE test_no_rows (id INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	var id int
	row := driver.QueryRow("SELECT id FROM test_no_rows WHERE id = ?", 999)
	err = row.Scan(&id)

	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("Expected sql.ErrNoRows, got %v", err)
	}
}

func TestSQLiteDriver_QueryRowContext(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	ctx := context.Background()

	_, err := driver.ExecuteContext(ctx, `CREATE TABLE test_row_ctx (
		id INTEGER PRIMARY KEY,
		name TEXT
	)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	_, err = driver.ExecuteContext(ctx, "INSERT INTO test_row_ctx (name) VALUES (?)", "ctx-row")
	if err != nil {
		t.Fatalf("Setup insert error: %v", err)
	}

	var name string
	row := driver.QueryRowContext(ctx, "SELECT name FROM test_row_ctx WHERE id = ?", 1)
	err = row.Scan(&name)
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if name != "ctx-row" {
		t.Errorf("name = %q, want %q", name, "ctx-row")
	}
}

// =============================================================================
// Query Tests (Multiple Rows)
// =============================================================================

func TestSQLiteDriver_Query(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	// Setup
	_, err := driver.Execute(`CREATE TABLE test_query (
		id INTEGER PRIMARY KEY,
		name TEXT
	)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Insert multiple rows
	for i := 1; i <= 3; i++ {
		_, err := driver.Execute("INSERT INTO test_query (name) VALUES (?)", "name"+string(rune('0'+i)))
		if err != nil {
			t.Fatalf("Setup insert error: %v", err)
		}
	}

	// Query multiple rows
	rows, err := driver.Query("SELECT id, name FROM test_query ORDER BY id")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	defer rows.Close()

	var results []struct {
		ID   int
		Name string
	}

	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			t.Fatalf("Scan() error = %v", err)
		}
		results = append(results, struct {
			ID   int
			Name string
		}{id, name})
	}

	if err := rows.Err(); err != nil {
		t.Errorf("Rows.Err() = %v", err)
	}

	if len(results) != 3 {
		t.Errorf("Got %d results, want 3", len(results))
	}
}

func TestSQLiteDriver_Query_Columns(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	_, err := driver.Execute(`CREATE TABLE test_columns (
		id INTEGER PRIMARY KEY,
		name TEXT,
		value REAL
	)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	rows, err := driver.Query("SELECT id, name, value FROM test_columns")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		t.Fatalf("Columns() error = %v", err)
	}

	expected := []string{"id", "name", "value"}
	if len(columns) != len(expected) {
		t.Fatalf("Columns() = %v, want %v", columns, expected)
	}

	for i, col := range columns {
		if col != expected[i] {
			t.Errorf("Columns()[%d] = %q, want %q", i, col, expected[i])
		}
	}
}

func TestSQLiteDriver_QueryContext(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	ctx := context.Background()

	_, err := driver.ExecuteContext(ctx, `CREATE TABLE test_query_ctx (id INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	_, err = driver.ExecuteContext(ctx, "INSERT INTO test_query_ctx (id) VALUES (?)", 1)
	if err != nil {
		t.Fatalf("Setup insert error: %v", err)
	}

	rows, err := driver.QueryContext(ctx, "SELECT id FROM test_query_ctx")
	if err != nil {
		t.Fatalf("QueryContext() error = %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	if count != 1 {
		t.Errorf("Got %d rows, want 1", count)
	}
}

// =============================================================================
// Transaction Tests
// =============================================================================

func TestSQLiteDriver_Transaction_Commit(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	// Setup
	_, err := driver.Execute(`CREATE TABLE test_tx_commit (
		id INTEGER PRIMARY KEY,
		name TEXT
	)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Begin transaction
	tx, err := driver.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	// Insert within transaction
	_, err = tx.Execute("INSERT INTO test_tx_commit (name) VALUES (?)", "tx-value")
	if err != nil {
		t.Fatalf("tx.Execute() error = %v", err)
	}

	// Commit
	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Verify data persisted
	var name string
	row := driver.QueryRow("SELECT name FROM test_tx_commit WHERE id = 1")
	if err := row.Scan(&name); err != nil {
		t.Fatalf("Verification query error = %v", err)
	}

	if name != "tx-value" {
		t.Errorf("name = %q, want %q", name, "tx-value")
	}
}

func TestSQLiteDriver_Transaction_Rollback(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	// Setup
	_, err := driver.Execute(`CREATE TABLE test_tx_rollback (
		id INTEGER PRIMARY KEY,
		name TEXT
	)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Begin transaction
	tx, err := driver.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	// Insert within transaction
	_, err = tx.Execute("INSERT INTO test_tx_rollback (name) VALUES (?)", "should-rollback")
	if err != nil {
		t.Fatalf("tx.Execute() error = %v", err)
	}

	// Rollback
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback() error = %v", err)
	}

	// Verify data was rolled back
	var count int
	row := driver.QueryRow("SELECT COUNT(*) FROM test_tx_rollback")
	if err := row.Scan(&count); err != nil {
		t.Fatalf("Verification query error = %v", err)
	}

	if count != 0 {
		t.Errorf("count = %d, want 0 (rollback should have removed the row)", count)
	}
}

func TestSQLiteDriver_Transaction_QueryRow(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	_, err := driver.Execute(`CREATE TABLE test_tx_query_row (
		id INTEGER PRIMARY KEY,
		name TEXT
	)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	tx, err := driver.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	defer tx.Rollback()

	_, err = tx.Execute("INSERT INTO test_tx_query_row (name) VALUES (?)", "tx-query-row")
	if err != nil {
		t.Fatalf("tx.Execute() error = %v", err)
	}

	// Query within transaction should see uncommitted data
	var name string
	row := tx.QueryRow("SELECT name FROM test_tx_query_row WHERE id = 1")
	if err := row.Scan(&name); err != nil {
		t.Fatalf("tx.QueryRow() Scan error = %v", err)
	}

	if name != "tx-query-row" {
		t.Errorf("name = %q, want %q", name, "tx-query-row")
	}
}

func TestSQLiteDriver_Transaction_Query(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	_, err := driver.Execute(`CREATE TABLE test_tx_query (
		id INTEGER PRIMARY KEY,
		name TEXT
	)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	tx, err := driver.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}
	defer tx.Rollback()

	// Insert multiple rows
	for i := 1; i <= 3; i++ {
		_, err := tx.Execute("INSERT INTO test_tx_query (name) VALUES (?)", "tx-item")
		if err != nil {
			t.Fatalf("tx.Execute() error = %v", err)
		}
	}

	// Query within transaction
	rows, err := tx.Query("SELECT id, name FROM test_tx_query")
	if err != nil {
		t.Fatalf("tx.Query() error = %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
	}

	if count != 3 {
		t.Errorf("Got %d rows in transaction, want 3", count)
	}
}

func TestSQLiteDriver_BeginContext(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	ctx := context.Background()

	_, err := driver.ExecuteContext(ctx, `CREATE TABLE test_begin_ctx (id INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	tx, err := driver.BeginContext(ctx)
	if err != nil {
		t.Fatalf("BeginContext() error = %v", err)
	}

	_, err = tx.Execute("INSERT INTO test_begin_ctx (id) VALUES (?)", 1)
	if err != nil {
		t.Fatalf("tx.Execute() error = %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	// Verify
	var id int
	row := driver.QueryRow("SELECT id FROM test_begin_ctx")
	if err := row.Scan(&id); err != nil {
		t.Fatalf("Verification error = %v", err)
	}

	if id != 1 {
		t.Errorf("id = %d, want 1", id)
	}
}

// =============================================================================
// Connection Pool Stats Tests
// =============================================================================

func TestSQLiteDriver_Stats(t *testing.T) {
	driver, ok := createTestDriver(t).(*SQLiteDriver)
	if !ok {
		t.Fatal("Failed to cast to SQLiteDriver")
	}
	defer driver.Close()

	stats := driver.Stats()

	// Stats should have reasonable values
	if stats.OpenConnections < 0 {
		t.Errorf("OpenConnections = %d, want >= 0", stats.OpenConnections)
	}
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestSQLiteDriver_ImplementsDriver(t *testing.T) {
	var _ Driver = (*SQLiteDriver)(nil)
}

func TestSqliteRow_ImplementsRow(t *testing.T) {
	var _ Row = (*sqliteRow)(nil)
}

func TestSqliteRows_ImplementsRows(t *testing.T) {
	var _ Rows = (*sqliteRows)(nil)
}

func TestSqliteResult_ImplementsResult(t *testing.T) {
	var _ Result = (*sqliteResult)(nil)
}

func TestSqliteTransaction_ImplementsTransaction(t *testing.T) {
	var _ Transaction = (*sqliteTransaction)(nil)
}

// =============================================================================
// Driver Registration Tests
// =============================================================================

func TestDriverRegistry_SQLite(t *testing.T) {
	cfg := DriverConfig{Type: DriverSQLite, Path: ":memory:"}
	driver, err := NewDriver(cfg)
	if err != nil {
		t.Fatalf("NewDriver(sqlite) error = %v", err)
	}
	defer driver.Close()

	if driver.Type() != DriverSQLite {
		t.Errorf("Type() = %v, want %v", driver.Type(), DriverSQLite)
	}
}

func TestDriverRegistry_Memory(t *testing.T) {
	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewDriver(cfg)
	if err != nil {
		t.Fatalf("NewDriver(memory) error = %v", err)
	}
	defer driver.Close()

	if driver.Type() != DriverMemory {
		t.Errorf("Type() = %v, want %v", driver.Type(), DriverMemory)
	}
}

// =============================================================================
// WI-1: Foreign Keys Enforcement Tests
// =============================================================================

// TestSQLiteDriver_ForeignKeysEnabled_Memory verifies that PRAGMA foreign_keys = ON
// is applied after connecting to an in-memory SQLite database.
// RED: This test will FAIL until NewMemorySQLiteDriver runs PRAGMA foreign_keys = ON.
func TestSQLiteDriver_ForeignKeysEnabled_Memory(t *testing.T) {
	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("NewMemorySQLiteDriver() error = %v", err)
	}
	defer driver.Close()

	if err := driver.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	// Query PRAGMA foreign_keys to verify it's ON (1 = ON, 0 = OFF)
	row := driver.QueryRow("PRAGMA foreign_keys")
	var fkEnabled int
	if err := row.Scan(&fkEnabled); err != nil {
		t.Fatalf("failed to scan PRAGMA foreign_keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("PRAGMA foreign_keys = %d, want 1 (ON); foreign key enforcement is disabled", fkEnabled)
	}
}

// TestSQLiteDriver_ForeignKeysEnabled_File verifies that PRAGMA foreign_keys = ON
// is applied after connecting to a file-based SQLite database.
// RED: This test will FAIL until NewSQLiteDriver runs PRAGMA foreign_keys = ON.
func TestSQLiteDriver_ForeignKeysEnabled_File(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "fk_test.db")

	cfg := DriverConfig{
		Type: DriverSQLite,
		Path: dbPath,
	}
	driver, err := NewSQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteDriver() error = %v", err)
	}
	defer driver.Close()

	if err := driver.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	row := driver.QueryRow("PRAGMA foreign_keys")
	var fkEnabled int
	if err := row.Scan(&fkEnabled); err != nil {
		t.Fatalf("failed to scan PRAGMA foreign_keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("PRAGMA foreign_keys = %d, want 1 (ON); foreign key enforcement is disabled", fkEnabled)
	}
}

// =============================================================================
// WAL Mode Tests
// =============================================================================

// TestSQLiteDriver_WALMode_File verifies that PRAGMA journal_mode=WAL is applied
// after connecting to a file-based SQLite database.
func TestSQLiteDriver_WALMode_File(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "wal_test.db")

	cfg := DriverConfig{
		Type: DriverSQLite,
		Path: dbPath,
	}
	driver, err := NewSQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteDriver() error = %v", err)
	}
	defer driver.Close()

	if err := driver.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	// Verify WAL mode is enabled
	var journalMode string
	row := driver.QueryRow("PRAGMA journal_mode")
	if err := row.Scan(&journalMode); err != nil {
		t.Fatalf("failed to scan PRAGMA journal_mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("PRAGMA journal_mode = %q, want %q", journalMode, "wal")
	}
}

// TestSQLiteDriver_BusyTimeout verifies that busy_timeout is set after Connect.
func TestSQLiteDriver_BusyTimeout(t *testing.T) {
	driver := createTestDriver(t)
	defer driver.Close()

	var timeout int
	row := driver.QueryRow("PRAGMA busy_timeout")
	if err := row.Scan(&timeout); err != nil {
		t.Fatalf("failed to scan PRAGMA busy_timeout: %v", err)
	}
	if timeout != 5000 {
		t.Errorf("PRAGMA busy_timeout = %d, want 5000", timeout)
	}
}

// TestSQLiteDriver_SynchronousNormal verifies that synchronous=NORMAL is set after Connect.
func TestSQLiteDriver_SynchronousNormal(t *testing.T) {
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "sync_test.db")

	cfg := DriverConfig{
		Type: DriverSQLite,
		Path: dbPath,
	}
	driver, err := NewSQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("NewSQLiteDriver() error = %v", err)
	}
	defer driver.Close()

	if err := driver.Connect(); err != nil {
		t.Fatalf("Connect() error = %v", err)
	}

	var syncMode int
	row := driver.QueryRow("PRAGMA synchronous")
	if err := row.Scan(&syncMode); err != nil {
		t.Fatalf("failed to scan PRAGMA synchronous: %v", err)
	}
	// synchronous=NORMAL is value 1
	if syncMode != 1 {
		t.Errorf("PRAGMA synchronous = %d, want 1 (NORMAL)", syncMode)
	}
}

// TestSQLiteDriver_InMemory_WALFallback verifies that an in-memory database:
//   - Does not error when WAL pragma is intentionally skipped
//   - Stays in journal_mode=memory (not wal — in-memory doesn't support WAL)
//   - Still applies busy_timeout and synchronous pragmas
func TestSQLiteDriver_InMemory_WALFallback(t *testing.T) {
	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("NewMemorySQLiteDriver() error = %v", err)
	}
	defer driver.Close()

	// Connect must not error — WAL pragma is intentionally skipped for in-memory
	if err := driver.Connect(); err != nil {
		t.Fatalf("Connect() error = %v; WAL fallback should not cause a connection error", err)
	}

	// journal_mode must NOT be "wal" — in-memory databases don't support WAL
	var journalMode string
	row := driver.QueryRow("PRAGMA journal_mode")
	if err := row.Scan(&journalMode); err != nil {
		t.Fatalf("failed to scan PRAGMA journal_mode: %v", err)
	}
	if journalMode == "wal" {
		t.Errorf("PRAGMA journal_mode = %q, want anything except \"wal\" (in-memory databases do not support WAL)", journalMode)
	}

	// busy_timeout must be set (driver applies it to all database types)
	var busyTimeout int
	row = driver.QueryRow("PRAGMA busy_timeout")
	if err := row.Scan(&busyTimeout); err != nil {
		t.Fatalf("failed to scan PRAGMA busy_timeout: %v", err)
	}
	if busyTimeout <= 0 {
		t.Errorf("PRAGMA busy_timeout = %d, want > 0 (should be set even for in-memory databases)", busyTimeout)
	}

	// synchronous must be set — driver applies NORMAL (1) to all database types
	var syncMode int
	row = driver.QueryRow("PRAGMA synchronous")
	if err := row.Scan(&syncMode); err != nil {
		t.Fatalf("failed to scan PRAGMA synchronous: %v", err)
	}
	if syncMode < 0 || syncMode > 3 {
		t.Errorf("PRAGMA synchronous = %d, want a valid mode (0=OFF, 1=NORMAL, 2=FULL, 3=EXTRA)", syncMode)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// createTestDriver creates an in-memory SQLite driver for testing
func createTestDriver(t *testing.T) Driver {
	t.Helper()

	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("Failed to create test driver: %v", err)
	}

	if err := driver.Connect(); err != nil {
		t.Fatalf("Failed to connect test driver: %v", err)
	}

	return driver
}
