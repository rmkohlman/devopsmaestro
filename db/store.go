package db

import (
	"fmt"
)

// SQLDataStore is a concrete implementation of the DataStore interface.
// It uses the Driver interface for database operations and QueryBuilder
// for dialect-specific SQL generation.
type SQLDataStore struct {
	driver       Driver
	queryBuilder QueryBuilder
}

// NewSQLDataStore creates a new SQLDataStore with the given driver.
// If queryBuilder is nil, the appropriate builder is selected based on driver type.
func NewSQLDataStore(driver Driver, queryBuilder QueryBuilder) *SQLDataStore {
	if queryBuilder == nil {
		queryBuilder = QueryBuilderFor(driver.Type())
	}
	return &SQLDataStore{
		driver:       driver,
		queryBuilder: queryBuilder,
	}
}

// NewDataStore creates a DataStore from configuration.
// This is the recommended way to create a DataStore.
func NewDataStore(cfg DataStoreConfig) (DataStore, error) {
	if cfg.Driver == nil {
		return nil, fmt.Errorf("driver is required")
	}
	return NewSQLDataStore(cfg.Driver, cfg.QueryBuilder), nil
}

// Driver returns the underlying database driver.
func (ds *SQLDataStore) Driver() Driver {
	return ds.driver
}

// Close releases any resources held by the DataStore.
func (ds *SQLDataStore) Close() error {
	return ds.driver.Close()
}

// Ping verifies the database connection is alive.
func (ds *SQLDataStore) Ping() error {
	return ds.driver.Ping()
}

// Ensure SQLDataStore implements DataStore interface
var _ DataStore = (*SQLDataStore)(nil)
