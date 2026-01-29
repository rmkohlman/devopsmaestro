package db

import (
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/spf13/viper"
)

// Database is the legacy interface for database operations.
// Deprecated: Use Driver interface instead.
type Database interface {
	Connect() error
	Close() error
	Execute(query string, args ...interface{}) (interface{}, error)
	FetchOne(query string, args ...interface{}) (interface{}, error)
	FetchMany(query string, args ...interface{}) (interface{}, error)
	DSN() string
	MigrationDSN() string
}

// InitializeDriver creates and connects a driver based on viper configuration.
func InitializeDriver() (Driver, error) {
	dbType := viper.GetString("database.type")
	if dbType == "" {
		dbType = "sqlite"
	}

	cfg := DriverConfig{
		Type:     DriverType(dbType),
		Path:     viper.GetString("database.path"),
		Host:     viper.GetString("database.host"),
		Port:     viper.GetString("database.port"),
		Database: viper.GetString("database.name"),
		Username: viper.GetString("database.username"),
		Password: viper.GetString("database.password"),
		SSLMode:  viper.GetString("database.sslmode"),
	}

	driver, err := NewDriver(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	if err := driver.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return driver, nil
}

// RunMigrations runs database migrations using the provided driver.
func RunMigrations(driver Driver, migrationsFS fs.FS) error {
	if driver == nil {
		return fmt.Errorf("driver is nil")
	}

	// Get the subdirectory for this database type
	dbType := string(driver.Type())
	if dbType == string(DriverMemory) {
		dbType = "sqlite" // Memory driver uses sqlite migrations
	}

	subFS, err := fs.Sub(migrationsFS, dbType)
	if err != nil {
		return fmt.Errorf("failed to get migrations subdirectory for %s: %w", dbType, err)
	}

	// Create iofs source driver
	sourceDriver, err := iofs.New(subFS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	// Get the Migration DSN from the driver
	migrationDSN := driver.MigrationDSN()

	// Initialize the migrations
	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, migrationDSN)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}
	defer m.Close()

	// Apply the migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// =============================================================================
// Legacy Functions (for backward compatibility during migration)
// =============================================================================

// InitializeDBConnection sets up the database connection when the application starts.
// Deprecated: Use InitializeDriver instead.
func InitializeDBConnection() (Database, error) {
	dbInstance, err := DatabaseFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	if err := dbInstance.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return dbInstance, nil
}

// InitializeDatabase runs the migrations using the dbInstance created by the factory.
// Deprecated: Use RunMigrations instead.
func InitializeDatabase(dbInstance Database, migrationsFS fs.FS) error {
	if dbInstance == nil {
		return fmt.Errorf("database instance is not initialized")
	}

	dbType := viper.GetString("database.type")
	if dbType == "" {
		dbType = "sqlite"
	}

	subFS, err := fs.Sub(migrationsFS, dbType)
	if err != nil {
		return fmt.Errorf("failed to get migrations subdirectory for %s: %w", dbType, err)
	}

	sourceDriver, err := iofs.New(subFS, ".")
	if err != nil {
		return fmt.Errorf("failed to create migration source: %w", err)
	}

	migrationDSN := dbInstance.MigrationDSN()

	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, migrationDSN)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// SnapshotDatabase creates a snapshot of the current database state.
// Deprecated: Use DataStore methods instead.
func SnapshotDatabase(database Database) error {
	return nil
}

// BackupDatabase creates a backup of the current database state.
// Deprecated: Use DataStore methods instead.
func BackupDatabase(database Database) error {
	return nil
}
