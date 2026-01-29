package db

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDriver implements the Driver interface for SQLite databases.
type SQLiteDriver struct {
	conn *sql.DB
	cfg  DriverConfig
	dsn  string
}

// sqliteRow wraps sql.Row to implement the Row interface.
type sqliteRow struct {
	row *sql.Row
}

func (r *sqliteRow) Scan(dest ...interface{}) error {
	return r.row.Scan(dest...)
}

// sqliteRows wraps sql.Rows to implement the Rows interface.
type sqliteRows struct {
	rows *sql.Rows
}

func (r *sqliteRows) Next() bool {
	return r.rows.Next()
}

func (r *sqliteRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

func (r *sqliteRows) Close() error {
	return r.rows.Close()
}

func (r *sqliteRows) Err() error {
	return r.rows.Err()
}

func (r *sqliteRows) Columns() ([]string, error) {
	return r.rows.Columns()
}

// sqliteResult wraps sql.Result to implement the Result interface.
type sqliteResult struct {
	result sql.Result
}

func (r *sqliteResult) LastInsertId() (int64, error) {
	return r.result.LastInsertId()
}

func (r *sqliteResult) RowsAffected() (int64, error) {
	return r.result.RowsAffected()
}

// sqliteTransaction wraps sql.Tx to implement the Transaction interface.
type sqliteTransaction struct {
	tx *sql.Tx
}

func (t *sqliteTransaction) Execute(query string, args ...interface{}) (Result, error) {
	result, err := t.tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return &sqliteResult{result: result}, nil
}

func (t *sqliteTransaction) QueryRow(query string, args ...interface{}) Row {
	return &sqliteRow{row: t.tx.QueryRow(query, args...)}
}

func (t *sqliteTransaction) Query(query string, args ...interface{}) (Rows, error) {
	rows, err := t.tx.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return &sqliteRows{rows: rows}, nil
}

func (t *sqliteTransaction) Commit() error {
	return t.tx.Commit()
}

func (t *sqliteTransaction) Rollback() error {
	return t.tx.Rollback()
}

// Register SQLite driver on package init
func init() {
	RegisterDriver(DriverSQLite, NewSQLiteDriver)
	RegisterDriver(DriverMemory, NewMemorySQLiteDriver)
}

// NewSQLiteDriver creates a new SQLite driver from configuration.
func NewSQLiteDriver(cfg DriverConfig) (Driver, error) {
	path, err := ExpandPath(cfg.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to expand path: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?cache=shared&mode=rwc", path)

	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database: %w", err)
	}

	// Apply connection pool settings
	if cfg.MaxOpenConns > 0 {
		conn.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		conn.SetMaxIdleConns(cfg.MaxIdleConns)
	}

	return &SQLiteDriver{
		conn: conn,
		cfg:  cfg,
		dsn:  dsn,
	}, nil
}

// NewMemorySQLiteDriver creates an in-memory SQLite driver for testing.
func NewMemorySQLiteDriver(cfg DriverConfig) (Driver, error) {
	dsn := "file::memory:?cache=shared"

	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open in-memory SQLite database: %w", err)
	}

	return &SQLiteDriver{
		conn: conn,
		cfg:  cfg,
		dsn:  dsn,
	}, nil
}

// Connect establishes the database connection.
func (d *SQLiteDriver) Connect() error {
	return d.conn.Ping()
}

// Close closes the database connection.
func (d *SQLiteDriver) Close() error {
	return d.conn.Close()
}

// Ping verifies the database connection is alive.
func (d *SQLiteDriver) Ping() error {
	return d.conn.Ping()
}

// Execute runs a command that doesn't return rows.
func (d *SQLiteDriver) Execute(query string, args ...interface{}) (Result, error) {
	result, err := d.conn.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return &sqliteResult{result: result}, nil
}

// ExecuteContext runs a command with context support.
func (d *SQLiteDriver) ExecuteContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	result, err := d.conn.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &sqliteResult{result: result}, nil
}

// QueryRow executes a query expected to return at most one row.
func (d *SQLiteDriver) QueryRow(query string, args ...interface{}) Row {
	return &sqliteRow{row: d.conn.QueryRow(query, args...)}
}

// QueryRowContext executes a query with context support.
func (d *SQLiteDriver) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return &sqliteRow{row: d.conn.QueryRowContext(ctx, query, args...)}
}

// Query executes a query that returns multiple rows.
func (d *SQLiteDriver) Query(query string, args ...interface{}) (Rows, error) {
	rows, err := d.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return &sqliteRows{rows: rows}, nil
}

// QueryContext executes a query with context support.
func (d *SQLiteDriver) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	rows, err := d.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return &sqliteRows{rows: rows}, nil
}

// Begin starts a new transaction.
func (d *SQLiteDriver) Begin() (Transaction, error) {
	tx, err := d.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &sqliteTransaction{tx: tx}, nil
}

// BeginContext starts a new transaction with context.
func (d *SQLiteDriver) BeginContext(ctx context.Context) (Transaction, error) {
	tx, err := d.conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &sqliteTransaction{tx: tx}, nil
}

// Type returns the driver type.
func (d *SQLiteDriver) Type() DriverType {
	if d.cfg.Type == DriverMemory {
		return DriverMemory
	}
	return DriverSQLite
}

// DSN returns the data source name.
func (d *SQLiteDriver) DSN() string {
	return d.dsn
}

// MigrationDSN returns the DSN formatted for golang-migrate.
func (d *SQLiteDriver) MigrationDSN() string {
	if d.cfg.Type == DriverMemory {
		return "sqlite3://:memory:"
	}
	path, _ := ExpandPath(d.cfg.Path)
	return fmt.Sprintf("sqlite:///%s", path)
}

// Stats returns connection pool statistics.
func (d *SQLiteDriver) Stats() DriverStats {
	stats := d.conn.Stats()
	return DriverStats{
		OpenConnections:    stats.OpenConnections,
		InUse:              stats.InUse,
		Idle:               stats.Idle,
		WaitCount:          stats.WaitCount,
		MaxOpenConnections: stats.MaxOpenConnections,
	}
}

// Ensure SQLiteDriver implements Driver interface
var _ Driver = (*SQLiteDriver)(nil)

// Ensure wrapper types implement their interfaces
var _ Row = (*sqliteRow)(nil)
var _ Rows = (*sqliteRows)(nil)
var _ Result = (*sqliteResult)(nil)
var _ Transaction = (*sqliteTransaction)(nil)
