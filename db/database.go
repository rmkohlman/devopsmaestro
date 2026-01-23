package db

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
)

// Database is a generic interface for database operations
type Database interface {
	Connect() error
	Close() error
	Execute(query string, args ...interface{}) (interface{}, error)   // For executing commands (e.g., INSERT, UPDATE, DELETE)
	FetchOne(query string, args ...interface{}) (interface{}, error)  // For fetching a single record
	FetchMany(query string, args ...interface{}) (interface{}, error) // For fetching multiple records
	DSN() string                                                      // Method to provide the data source name
	MigrationDSN() string                                             // Method to provide the migration data source name
}

// InitializeDBConnection sets up the database connection when the application starts
func InitializeDBConnection() (Database, error) {
	dbInstance, err := DatabaseFactory() // Use the factory to get the appropriate database connection
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	if err := dbInstance.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}

	return dbInstance, nil
}

// InitializeDatabase runs the migrations using the dbInstance created by the factory
func InitializeDatabase(dbInstance Database) error {
	if dbInstance == nil {
		return fmt.Errorf("database instance is not initialized")
	}

	// Get the database type from the configuration
	dbType := strings.ToLower(viper.GetString("database.type"))
	fmt.Printf("Database type: %s\n", dbType)

	// Get the current working directory
	workingDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}
	fmt.Printf("Working directory: %s\n", workingDir)

	// Construct the path to the migrations directory
	migrationDir, err := filepath.Abs(filepath.Join(workingDir, "migrations", dbType))
	if err != nil {
		return fmt.Errorf("failed to get absolute migration directory path: %v", err)
	}
	fmt.Printf("Migration directory: %s\n", migrationDir)

	// Format the migration path as a URL
	migrationPath := fmt.Sprintf("file://%s", migrationDir)
	fmt.Printf("Migration path: %s\n", migrationPath)
	// Get the Data Source Name (DSN) and Migration DSN from the database instance
	dsn := dbInstance.DSN()
	migration_dns := dbInstance.MigrationDSN()
	fmt.Printf("DSN: %s\n", dsn)
	fmt.Printf("Migration DSN: %s\n", migration_dns)
	fmt.Printf("Migration DSN: %s\n", migration_dns)

	// Initialize the migrations
	m, err := migrate.New(migrationPath, migration_dns) // No need for dbInstance.DSN()
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %v, file: %v, dsn: %v", err, migrationPath, migration_dns)
	}

	// Apply the migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %v", err)
	}

	return nil
}

// SnapshotDatabase creates a snapshot of the current database state and stores it as a YAML file
func SnapshotDatabase(database Database) error {
	// Implementation to create a snapshot of the database
	// and save it to the specified file path
	return nil
}

// BackupDatabase creates a backup of the current database state and stores it as a YAML file
func BackupDatabase(database Database) error {
	// Implementation to create a backup of the database
	// and save it to the specified file path
	return nil
}
