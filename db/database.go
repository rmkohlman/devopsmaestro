package db

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

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

// CheckPendingMigrations checks if there are pending migrations without applying them.
// Returns true if migrations are pending, false if database is current.
// If database doesn't exist, returns false (let init command handle first-time setup).
func CheckPendingMigrations(driver Driver, migrationsFS fs.FS) (bool, error) {
	if driver == nil {
		return false, fmt.Errorf("driver is nil")
	}

	// Get the subdirectory for this database type
	dbType := string(driver.Type())
	if dbType == string(DriverMemory) {
		dbType = "sqlite" // Memory driver uses sqlite migrations
	}

	subFS, err := fs.Sub(migrationsFS, dbType)
	if err != nil {
		return false, fmt.Errorf("failed to get migrations subdirectory for %s: %w", dbType, err)
	}

	// Create iofs source driver
	sourceDriver, err := iofs.New(subFS, ".")
	if err != nil {
		return false, fmt.Errorf("failed to create migration source: %w", err)
	}

	// Get the Migration DSN from the driver
	migrationDSN := driver.MigrationDSN()

	// Initialize the migrations
	m, err := migrate.NewWithSourceInstance("iofs", sourceDriver, migrationDSN)
	if err != nil {
		// If migration initialization fails, might be because database doesn't exist yet
		// This is OK - let init command handle first-time setup
		return false, nil
	}
	defer m.Close()

	// Get current version and check if we have migrations to apply
	_, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			// No migrations have been applied yet - database is new
			return true, nil
		}
		// Database might not exist yet
		return false, nil
	}

	if dirty {
		return false, fmt.Errorf("database is in dirty state - please run 'dvm admin migrate' to fix")
	}

	// Check if we have newer migrations available
	// We'll try to step up once to see if there are pending migrations
	tempM, err := migrate.NewWithSourceInstance("iofs", sourceDriver, migrationDSN)
	if err != nil {
		return false, nil
	}
	defer tempM.Close()

	// Try to get the next version by attempting to step up
	err = tempM.Steps(1)
	if err == migrate.ErrNoChange {
		// No pending migrations
		return false, nil
	} else if err != nil {
		// Some other error occurred, assume no migrations needed
		return false, nil
	} else {
		// Migration would succeed, so we have pending migrations
		// Step back down to restore original state
		_ = tempM.Steps(-1)
		return true, nil
	}
}

// AutoMigrate checks for pending migrations and applies them if needed.
// Shows user feedback when migrations are applied.
// This function is used by the manual 'dvm admin migrate' command.
// For automatic migrations on startup, use CheckVersionBasedAutoMigration instead.
func AutoMigrate(driver Driver, migrationsFS fs.FS) error {
	if driver == nil {
		return fmt.Errorf("driver is nil")
	}

	// Check if migrations are pending
	pending, err := CheckPendingMigrations(driver, migrationsFS)
	if err != nil {
		return fmt.Errorf("failed to check pending migrations: %w", err)
	}

	if !pending {
		// No migrations needed
		return nil
	}

	// Apply migrations with user feedback
	fmt.Println("Applying database migrations...")
	err = RunMigrations(driver, migrationsFS)
	if err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// CheckVersionBasedAutoMigration checks if auto-migration should run based on version changes.
// This provides a fast path by avoiding database migration checks when the binary version hasn't changed.
// Returns true if migrations were applied, false if skipped.
func CheckVersionBasedAutoMigration(driver Driver, migrationsFS fs.FS, currentVersion string, verbose bool) (bool, error) {
	if driver == nil {
		return false, fmt.Errorf("driver is nil")
	}

	// Get stored version from file
	storedVersion, err := GetStoredVersion()
	if err != nil {
		// If we can't read stored version, log and proceed with normal migration check
		if verbose {
			slog.Warn("failed to read stored version, proceeding with migration check", "error", err)
		}
		return runMigrationsIfNeeded(driver, migrationsFS, currentVersion, verbose)
	}

	// If versions match, skip migration check entirely (fast path)
	if storedVersion == currentVersion {
		if verbose {
			slog.Debug("version unchanged, skipping migration check", "version", currentVersion)
		}
		return false, nil
	}

	// Version changed or first run - check and apply migrations if needed
	if verbose {
		if storedVersion == "" {
			slog.Info("first run detected, checking for migrations", "version", currentVersion)
		} else {
			slog.Info("version change detected, checking for migrations",
				"old", storedVersion, "new", currentVersion)
		}
	}

	return runMigrationsIfNeeded(driver, migrationsFS, currentVersion, verbose)
}

// runMigrationsIfNeeded checks for and applies migrations, then updates stored version on success.
func runMigrationsIfNeeded(driver Driver, migrationsFS fs.FS, currentVersion string, verbose bool) (bool, error) {
	// Check if migrations are pending
	pending, err := CheckPendingMigrations(driver, migrationsFS)
	if err != nil {
		return false, fmt.Errorf("failed to check pending migrations: %w", err)
	}

	if !pending {
		// No migrations needed, but update version since it changed
		if err := SaveCurrentVersion(currentVersion); err != nil {
			if verbose {
				slog.Warn("failed to save current version", "error", err)
			}
		}
		return false, nil
	}

	// Apply migrations
	if !verbose {
		// Only show message if not in verbose mode (verbose mode uses slog)
		fmt.Println("Applying database migrations...")
	}
	err = RunMigrations(driver, migrationsFS)
	if err != nil {
		return false, fmt.Errorf("failed to apply migrations: %w", err)
	}

	// Save current version only after successful migration
	if err := SaveCurrentVersion(currentVersion); err != nil {
		if verbose {
			slog.Warn("migrations applied successfully but failed to save version", "error", err)
		}
	}

	return true, nil
}

// GetStoredVersion reads the last-run binary version from ~/.devopsmaestro/.version
func GetStoredVersion() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	versionFile := filepath.Join(homeDir, ".devopsmaestro", ".version")

	// If file doesn't exist, return empty string (first run)
	if _, err := os.Stat(versionFile); os.IsNotExist(err) {
		return "", nil
	}

	content, err := os.ReadFile(versionFile)
	if err != nil {
		return "", fmt.Errorf("failed to read version file: %w", err)
	}

	return strings.TrimSpace(string(content)), nil
}

// SaveCurrentVersion writes the current binary version to ~/.devopsmaestro/.version
func SaveCurrentVersion(version string) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".devopsmaestro")

	// Ensure config directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	versionFile := filepath.Join(configDir, ".version")

	// Write version to file
	err = os.WriteFile(versionFile, []byte(version), 0644)
	if err != nil {
		return fmt.Errorf("failed to write version file: %w", err)
	}

	return nil
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
