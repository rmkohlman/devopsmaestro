package db

import (
	"fmt"

	"github.com/spf13/viper"
)

// DatabaseCreator is a function type that creates a Database
type DatabaseCreator func() (Database, error)

// StoreCreator is a function type that creates a DataStore
type StoreCreator func(Database) (DataStore, error)

// databaseFactories is the map that holds all registered database creators
var databaseFactories = map[string]DatabaseCreator{}

// storeFactories is the map that holds all registered store creators
var storeFactories = map[string]StoreCreator{}

// RegisterDatabase registers a new database creator in the factory map
func RegisterDatabase(dbType string, creator DatabaseCreator) {
	databaseFactories[dbType] = creator
}

// RegisterStore registers a new store creator in the factory map
func RegisterStore(storeType string, creator StoreCreator) {
	storeFactories[storeType] = creator
}

// DatabaseFactory creates the appropriate Database based on configuration settings
func DatabaseFactory() (Database, error) {
	dbType := viper.GetString("database.type")

	creator, exists := databaseFactories[dbType]
	if !exists {
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	return creator()
}

// StoreFactory creates the appropriate DataStore based on configuration settings
func StoreFactory(db Database) (DataStore, error) {
	storeType := viper.GetString("store")

	creator, exists := storeFactories[storeType]
	if !exists {
		return nil, fmt.Errorf("unsupported store type: %s", storeType)
	}

	return creator(db)
}
