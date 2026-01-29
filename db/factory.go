package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
)

// =============================================================================
// Primary Factory Interface (Recommended)
// =============================================================================

// DataStoreFactory is the primary interface for creating DataStore instances.
// This allows different factory implementations to be swapped in for testing
// or different deployment scenarios.
type DataStoreFactory interface {
	// Create returns a fully configured, connected DataStore.
	Create() (DataStore, error)
}

// =============================================================================
// Default Factory Implementation
// =============================================================================

// DefaultDataStoreFactory creates DataStore instances using viper configuration.
type DefaultDataStoreFactory struct{}

// NewDataStoreFactory creates the default factory.
func NewDataStoreFactory() DataStoreFactory {
	return &DefaultDataStoreFactory{}
}

// Create returns a DataStore based on viper configuration.
func (f *DefaultDataStoreFactory) Create() (DataStore, error) {
	return CreateDataStore()
}

// CreateDataStore is the recommended way to create a fully configured DataStore.
// It creates the driver based on viper configuration and wraps it in a DataStore.
func CreateDataStore() (DataStore, error) {
	driver, err := DriverFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	if err := driver.Connect(); err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return NewSQLDataStore(driver, nil), nil
}

// CreateDataStoreWithDriver creates a DataStore using a provided Driver.
// This is useful for testing or when you need custom driver configuration.
func CreateDataStoreWithDriver(driver Driver) DataStore {
	return NewSQLDataStore(driver, nil)
}

// =============================================================================
// Driver Factory (uses driver.go's NewDriver and registry)
// =============================================================================

// DriverFactory creates a Driver based on viper configuration.
func DriverFactory() (Driver, error) {
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

	// Use NewDriver which uses the driver registry from driver.go
	return NewDriver(cfg)
}

// =============================================================================
// Legacy Support (Deprecated - for backward compatibility)
// =============================================================================

// DatabaseCreator is a function type that creates a Database.
// Deprecated: Use DriverCreator instead.
type DatabaseCreator func() (Database, error)

// StoreCreator is a function type that creates a DataStore.
// Deprecated: Use CreateDataStore instead.
type StoreCreator func(Database) (DataStore, error)

// databaseFactories holds registered legacy database creators
var databaseFactories = map[string]DatabaseCreator{}

// storeFactories holds registered legacy store creators
var storeFactories = map[string]StoreCreator{}

// RegisterDatabase registers a legacy database creator.
// Deprecated: Use RegisterDriver instead.
func RegisterDatabase(dbType string, creator DatabaseCreator) {
	databaseFactories[dbType] = creator
}

// RegisterStore registers a legacy store creator.
// Deprecated: No longer needed - use CreateDataStore.
func RegisterStore(storeType string, creator StoreCreator) {
	storeFactories[storeType] = creator
}

// Register the default "sql" store type for legacy compatibility
func init() {
	RegisterStore("sql", func(database Database) (DataStore, error) {
		// Wrap the legacy Database in a Driver adapter
		driver := NewDatabaseDriverAdapter(database)
		return NewSQLDataStore(driver, nil), nil
	})
}

// DatabaseFactory creates a Database based on configuration.
// Deprecated: Use DriverFactory and CreateDataStore instead.
func DatabaseFactory() (Database, error) {
	dbType := viper.GetString("database.type")
	if dbType == "" {
		dbType = "sqlite"
	}

	creator, exists := databaseFactories[dbType]
	if !exists {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	return creator()
}

// StoreFactory creates a DataStore from a Database.
// Deprecated: Use CreateDataStore instead.
func StoreFactory(db Database) (DataStore, error) {
	storeType := viper.GetString("store")
	if storeType == "" {
		storeType = "sql"
	}

	creator, exists := storeFactories[storeType]
	if !exists {
		return nil, fmt.Errorf("unsupported store type: %s", storeType)
	}

	return creator(db)
}

// =============================================================================
// DatabaseDriverAdapter - Bridges legacy Database to Driver interface
// =============================================================================

// DatabaseDriverAdapter wraps a legacy Database to implement the Driver interface.
// This allows legacy code using Database to work with the new Driver-based DataStore.
type DatabaseDriverAdapter struct {
	db         Database
	driverType DriverType
}

// NewDatabaseDriverAdapter creates a Driver adapter from a legacy Database.
func NewDatabaseDriverAdapter(db Database) *DatabaseDriverAdapter {
	// Determine driver type from viper config
	dbType := viper.GetString("database.type")
	if dbType == "" {
		dbType = "sqlite"
	}

	return &DatabaseDriverAdapter{
		db:         db,
		driverType: DriverType(dbType),
	}
}

// NewDatabaseDriverAdapterWithType creates a Driver adapter with explicit type.
func NewDatabaseDriverAdapterWithType(db Database, driverType DriverType) *DatabaseDriverAdapter {
	return &DatabaseDriverAdapter{
		db:         db,
		driverType: driverType,
	}
}

// Type returns the driver type.
func (d *DatabaseDriverAdapter) Type() DriverType {
	return d.driverType
}

// Connect establishes the database connection.
func (d *DatabaseDriverAdapter) Connect() error {
	return d.db.Connect()
}

// Close closes the database connection.
func (d *DatabaseDriverAdapter) Close() error {
	return d.db.Close()
}

// Ping verifies the database connection is alive.
func (d *DatabaseDriverAdapter) Ping() error {
	return d.db.Connect()
}

// Execute runs a command that doesn't return rows.
func (d *DatabaseDriverAdapter) Execute(query string, args ...interface{}) (Result, error) {
	result, err := d.db.Execute(query, args...)
	if err != nil {
		return nil, err
	}
	if sqlResult, ok := result.(sql.Result); ok {
		return sqlResult, nil
	}
	return nil, fmt.Errorf("unexpected result type from Execute: %T", result)
}

// ExecuteContext runs a command with context support.
func (d *DatabaseDriverAdapter) ExecuteContext(ctx context.Context, query string, args ...interface{}) (Result, error) {
	return d.Execute(query, args...)
}

// QueryRow executes a query expected to return at most one row.
func (d *DatabaseDriverAdapter) QueryRow(query string, args ...interface{}) Row {
	result, _ := d.db.FetchOne(query, args...)
	if row, ok := result.(*sql.Row); ok {
		return row
	}
	return nil
}

// QueryRowContext executes a query with context support.
func (d *DatabaseDriverAdapter) QueryRowContext(ctx context.Context, query string, args ...interface{}) Row {
	return d.QueryRow(query, args...)
}

// Query executes a query that returns multiple rows.
func (d *DatabaseDriverAdapter) Query(query string, args ...interface{}) (Rows, error) {
	result, err := d.db.FetchMany(query, args...)
	if err != nil {
		return nil, err
	}
	if rows, ok := result.(*sql.Rows); ok {
		return rows, nil
	}
	return nil, fmt.Errorf("unexpected result type from FetchMany: %T", result)
}

// QueryContext executes a query with context support.
func (d *DatabaseDriverAdapter) QueryContext(ctx context.Context, query string, args ...interface{}) (Rows, error) {
	return d.Query(query, args...)
}

// Begin starts a new transaction.
func (d *DatabaseDriverAdapter) Begin() (Transaction, error) {
	return nil, fmt.Errorf("transactions not supported via legacy Database adapter")
}

// BeginContext starts a new transaction with context.
func (d *DatabaseDriverAdapter) BeginContext(ctx context.Context) (Transaction, error) {
	return nil, fmt.Errorf("transactions not supported via legacy Database adapter")
}

// DSN returns the data source name.
func (d *DatabaseDriverAdapter) DSN() string {
	return d.db.DSN()
}

// MigrationDSN returns the DSN formatted for golang-migrate.
func (d *DatabaseDriverAdapter) MigrationDSN() string {
	return d.db.MigrationDSN()
}

// Database returns the underlying legacy Database.
// This allows access to the original Database when needed for migrations.
func (d *DatabaseDriverAdapter) Database() Database {
	return d.db
}

// Ensure DatabaseDriverAdapter implements Driver interface
var _ Driver = (*DatabaseDriverAdapter)(nil)
