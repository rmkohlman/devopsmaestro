package db

import (
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
