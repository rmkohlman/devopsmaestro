package db

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseFactory(t *testing.T) {

	viper.Set("database", "MOCK")
	RegisterDatabase("MOCK", NewMockDB)

	db, err := DatabaseFactory()
	assert.NoError(t, err)
	assert.NotNil(t, db)
}

func TestDatabaseFactory_UnsupportedType(t *testing.T) {
	viper.Set("database", "UNSUPPORTED")

	db, err := DatabaseFactory()
	assert.Error(t, err)
	assert.Nil(t, db)
}
func TestStoreFactory(t *testing.T) {

	viper.Set("store", "MOCK")
	RegisterStore("MOCK", func(db Database) (DataStore, error) {
		return NewMockDataStore(), nil
	})

	mockDB := new(MockDB)
	ds, err := StoreFactory(mockDB)
	assert.NoError(t, err)
	assert.NotNil(t, ds)
}

func TestStoreFactory_UnsupportedType(t *testing.T) {
	viper.Set("store", "UNSUPPORTED")
	RegisterStore("MOCK", func(db Database) (DataStore, error) {
		return NewMockDataStore(), nil
	})
	mockDB := new(MockDB)
	ds, err := StoreFactory(mockDB)
	assert.Error(t, err)
	assert.Nil(t, ds)
}
