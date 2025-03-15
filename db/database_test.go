package db

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file" // Import the file source driver
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func init() {
	viper.Set("database", "MOCK")
	RegisterDatabase("MOCK", NewMockDB)
}

func TestInitializeDBConnection(t *testing.T) {
	viper.Set("database", "MOCK")
	RegisterDatabase("MOCK", NewMockDB)
	mockDB := &MockDB{}
	SetMockDBInstance(mockDB)
	mockDB.On("Connect").Return(nil)

	dbInstance, err := InitializeDBConnection()
	assert.NoError(t, err)
	assert.NotNil(t, dbInstance)
	mockDB.AssertExpectations(t)
}

func TestInitializeDBConnection_FactoryError(t *testing.T) {
	viper.Set("database", "UNSUPPORTED")
	// Mock the DatabaseFactory function to return an error
	dbInstance, err := InitializeDBConnection()
	assert.Error(t, err)
	assert.Nil(t, dbInstance)
}

func TestInitializeDBConnection_ConnectError(t *testing.T) {
	viper.Set("database", "MOCK")
	RegisterDatabase("MOCK", NewMockDB)

	mockDB := &MockDB{}
	mockDB.On("Connect").Return(fmt.Errorf("connect error"))
	SetMockDBInstance(mockDB)

	dbInstance, err := InitializeDBConnection()
	assert.Error(t, err)
	assert.Nil(t, dbInstance)
	mockDB.AssertExpectations(t)
}

func TestInitializeDatabaseWithSqlite(t *testing.T) {
	// Set the databaseFile to be in the testutils directory
	databaseFilePath := "testutils/test.db"

	viper.Set("database", "SQLITE")
	viper.Set("database_file_path", databaseFilePath)

	sqliteDB, err := NewSQLiteDB()
	assert.NoError(t, err)

	err = InitializeDatabase(sqliteDB)
	assert.NoError(t, err)

	// Confirm that the table projects was created and has the expected columns from model/project.go
	rows, err := sqliteDB.FetchMany("SELECT name FROM sqlite_master WHERE type='table' AND name='projects'")
	assert.NoError(t, err)
	assert.NotNil(t, rows)

	sqlRows := rows.(*sql.Rows)
	defer func() {
		if sqlRows != nil {
			sqlRows.Close()
		}
	}()

	var tableName string
	for sqlRows.Next() {
		err := sqlRows.Scan(&tableName)
		assert.NoError(t, err)
		assert.Equal(t, "projects", tableName)
	}

	// Clean up the database file
	err = os.Remove(databaseFilePath)
	assert.NoError(t, err)
}

func TestInitializeMockDatabase(t *testing.T) {
	viper.Set("database", "MOCK")
	RegisterDatabase("MOCK", NewInMemoryTestDB)
	mockDB := &MockDB{}
	mockDB.On("DSN").Return("mock_dsn")
	mockDB.On("MigrationDSN").Return("sqlite3://testutils/test.db")

	err := InitializeDatabase(mockDB)
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

func TestInitializeDatabase_NoInstance(t *testing.T) {
	err := InitializeDatabase(nil)
	assert.Error(t, err)
}

func TestInitializeDatabase_MigrationError(t *testing.T) {
	viper.Set("database", "MOCK")
	RegisterDatabase("MOCK", NewMockDB)

	mockDB := &MockDB{}
	mockDB.On("DSN").Return("mock_dsn")
	mockDB.On("MigrationDSN").Return("mock_migration_dsn")

	// Mock viper configuration
	viper.Set("database.type", "invalid")

	err := InitializeDatabase(mockDB)
	assert.Error(t, err)
	mockDB.AssertExpectations(t)
}
