package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDB represents a SQLite database connection
type SQLiteDB struct {
	conn *sql.DB
	_dsn string
}

// buildDSN constructs the DSN string for SQLite connection
func buildSqliteDSN(filePath string) string {
	return fmt.Sprintf("file:%s?cache=shared&mode=rwc", filePath)
}

// getDSN retrieves the configuration values and constructs the DSN string
func sqliteDSN() string {
	// Get path from config
	filePath := viper.GetString("database.path")

	// If path is not set, use default
	if filePath == "" {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, ".devopsmaestro", "devopsmaestro.db")
	}

	// Expand tilde in path
	if filePath[:2] == "~/" {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, filePath[2:])
	}

	// Create the directory if it does not exist
	if _, err := os.Stat(filepath.Dir(filePath)); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(filePath), 0755)
	}
	return buildSqliteDSN(filePath)
}

func (s *SQLiteDB) DSN() string {
	if s._dsn == "" {
		filePath := viper.GetString("database.path")
		if filePath == "" {
			homeDir, _ := os.UserHomeDir()
			filePath = filepath.Join(homeDir, ".devopsmaestro", "devopsmaestro.db")
		}
		// Expand tilde
		if len(filePath) >= 2 && filePath[:2] == "~/" {
			homeDir, _ := os.UserHomeDir()
			filePath = filepath.Join(homeDir, filePath[2:])
		}
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			fmt.Printf("Error getting absolute path: %v\n", err)
			absPath = filePath
		}
		// Normal SQLite DSN for sql.Open()
		s._dsn = fmt.Sprintf("file:%s?cache=shared&mode=rwc", absPath)
	}
	return s._dsn
}

// MigrationDSN returns the correct DSN for golang-migrate
func (s *SQLiteDB) MigrationDSN() string {
	filePath := viper.GetString("database.path")
	if filePath == "" {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, ".devopsmaestro", "devopsmaestro.db")
	}
	// Expand tilde
	if len(filePath) >= 2 && filePath[:2] == "~/" {
		homeDir, _ := os.UserHomeDir()
		filePath = filepath.Join(homeDir, filePath[2:])
	}
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		fmt.Printf("Error getting absolute path for migrations: %v\n", err)
		absPath = filePath
	}
	return fmt.Sprintf("sqlite:///%s", absPath) // Correct DSN for golang-migrate
}

// Register the SQLite implementation with the factory
func init() {
	RegisterDatabase("sqlite", NewSQLiteDB)
	RegisterDatabase("SQLITE", NewSQLiteDB) // Keep uppercase for backward compatibility
}

// NewSQLiteDB creates a new SQLite database connection and returns it as a SQLiteDB instance
func NewSQLiteDB() (Database, error) {
	return connectSQLite(sqliteDSN())
}

// connectSQLite establishes the database connection using the DSN and returns a SQLiteDB instance
func connectSQLite(dataSourceName string) (*SQLiteDB, error) {
	conn, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &SQLiteDB{conn: conn, _dsn: dataSourceName}, nil
}

// Connect is part of the Database interface, ensuring the connection to the database
func (s *SQLiteDB) Connect() error {
	return s.conn.Ping()
}

// Close closes the database connection as part of the Database interface
func (s *SQLiteDB) Close() error {
	return s.conn.Close()
}

// Execute runs a command that doesn't return rows (e.g., INSERT, UPDATE, DELETE)
func (s *SQLiteDB) Execute(query string, args ...interface{}) (interface{}, error) {
	result, err := s.conn.Exec(query, args...)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FetchOne retrieves a single record from the database (e.g., SELECT WHERE id = ?)
func (s *SQLiteDB) FetchOne(query string, args ...interface{}) (interface{}, error) {
	return s.conn.QueryRow(query, args...), nil
}

// FetchMany retrieves multiple records from the database (e.g., SELECT * FROM table)
func (s *SQLiteDB) FetchMany(query string, args ...interface{}) (interface{}, error) {
	rows, err := s.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// Ensure SQLiteDB implements the Database interface
var _ Database = (*SQLiteDB)(nil)
