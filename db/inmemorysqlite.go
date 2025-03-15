package db

import (
	"database/sql"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/source/file" // Import the file source driver
	_ "github.com/mattn/go-sqlite3"                      // Import the SQLite driver
	"github.com/stretchr/testify/assert"
)

// SetupInMemorySQLiteDB sets up an in-memory SQLite database for testing and returns a cleanup function
func SetupInMemorySQLiteDB(t *testing.T) (Database, func()) {
	db, err := NewInMemoryTestDB()
	assert.NoError(t, err)
	// Return the database and a cleanup function
	return db, func() {
		// Cleanup logic if needed
	}
}

func InMemorySQLiteDNS() string {
	return "file::memory:?cache=shared"
}

type InMemorySQLiteDB struct {
	conn *sql.DB
	_dsn string
}

// NewInMemoryTestDB sets up an in-memory SQLite database for testing
func NewInMemoryTestDB() (Database, error) {
	return connectInMemorySQLite(InMemorySQLiteDNS())
}

// DSN returns the DSN string
func (s *InMemorySQLiteDB) DSN() string {
	return InMemorySQLiteDNS()
}

// MigrationDSN returns the DSN string for migrations
func (s *InMemorySQLiteDB) MigrationDSN() string {
	return "file://path/to/migrations"
}

// Register the SQLite implementation with the factory
func init() {
	RegisterDatabase("INMEMORYSQLITE", NewInMemoryTestDB)
}

// connectSQLite establishes the database connection using the DSN and returns a SQLiteDB instance
func connectInMemorySQLite(dataSourceName string) (*InMemorySQLiteDB, error) {
	conn, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &InMemorySQLiteDB{conn: conn, _dsn: dataSourceName}, nil
}

// Connect is part of the Database interface, ensuring the connection to the database
func (s *InMemorySQLiteDB) Connect() error {
	return s.conn.Ping()
}

// Close closes the database connection as part of the Database interface
func (s *InMemorySQLiteDB) Close() error {
	return s.conn.Close()
}

// Execute runs a command that doesn't return rows (e.g., INSERT, UPDATE, DELETE)
func (s *InMemorySQLiteDB) Execute(query string, args ...interface{}) (interface{}, error) {
	return s.conn.Exec(query, args...)
}

// FetchOne retrieves a single record from the database (e.g., SELECT WHERE id = ?)
func (s *InMemorySQLiteDB) FetchOne(query string, args ...interface{}) (interface{}, error) {
	return s.conn.QueryRow(query, args...), nil
}

// FetchMany retrieves multiple records from the database (e.g., SELECT * FROM table)
func (s *InMemorySQLiteDB) FetchMany(query string, args ...interface{}) (interface{}, error) {
	return s.conn.Query(query, args...)
}

// Ensure SQLiteDB implements the Database interface
var _ Database = (*InMemorySQLiteDB)(nil)
