package db

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test SetMockDBInstance function
func TestSetMockDBInstance(t *testing.T) {
	mockDB := &MockDB{}
	SetMockDBInstance(mockDB)

	assert.Equal(t, mockDBInstance, mockDB, "mockDBInstance should be set correctly")
}

// Test NewMockDB function
func TestNewMockDB(t *testing.T) {
	// Ensure a new instance is created
	mockDBInstance = nil
	db, err := NewMockDB()
	assert.NoError(t, err)
	assert.NotNil(t, db, "NewMockDB should return a valid instance")

	// Ensure it returns the existing instance
	db2, err := NewMockDB()
	assert.NoError(t, err)
	assert.Equal(t, db, db2, "NewMockDB should return the same instance")
}

// Test Connect function
func TestMockDB_Connect(t *testing.T) {
	mockDB := &MockDB{}
	mockDB.On("Connect").Return(nil)

	err := mockDB.Connect()
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

// Test Close function
func TestMockDB_Close(t *testing.T) {
	mockDB := &MockDB{}
	mockDB.On("Close").Return(nil)

	err := mockDB.Close()
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

// Test Execute function
func TestMockDB_Execute(t *testing.T) {
	mockDB := &MockDB{}
	query := "INSERT INTO projects (name) VALUES (?)"
	mockDB.On("Execute", query, mock.Anything).Return(int64(1), nil)

	result, err := mockDB.Execute(query, "test_project")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result, "Execute should return correct mock result")

	mockDB.AssertExpectations(t)
}

// Test FetchOne function
func TestMockDB_FetchOne(t *testing.T) {
	mockDB := &MockDB{}
	query := "SELECT name FROM projects WHERE id = ?"
	expectedRow := &sql.Row{}

	mockDB.On("FetchOne", query, mock.Anything).Return(expectedRow, nil)

	row, err := mockDB.FetchOne(query, 1)
	assert.NoError(t, err)
	assert.Equal(t, expectedRow, row, "FetchOne should return the correct row")

	mockDB.AssertExpectations(t)
}

// Test FetchMany function
func TestMockDB_FetchMany(t *testing.T) {
	mockDB := &MockDB{}
	query := "SELECT * FROM projects"

	expectedRows := &sql.Rows{}
	mockDB.On("FetchMany", query, mock.Anything).Return(expectedRows, nil)

	rows, err := mockDB.FetchMany(query)
	assert.NoError(t, err)
	assert.Equal(t, expectedRows, rows, "FetchMany should return the correct result")

	mockDB.AssertExpectations(t)
}

// Test DSN function
func TestMockDB_DSN(t *testing.T) {
	mockDB := &MockDB{}
	expectedDSN := "sqlite://test.db"

	mockDB.On("DSN").Return(expectedDSN)

	dsn := mockDB.DSN()
	assert.Equal(t, expectedDSN, dsn, "DSN should return the correct string")

	mockDB.AssertExpectations(t)
}

// Test MigrationDSN function
func TestMockDB_MigrationDSN(t *testing.T) {
	mockDB := &MockDB{}
	expectedDSN := "sqlite://migrations/test.db"

	mockDB.On("MigrationDSN").Return(expectedDSN)

	dsn := mockDB.MigrationDSN()
	assert.Equal(t, expectedDSN, dsn, "MigrationDSN should return the correct string")

	mockDB.AssertExpectations(t)
}

// Test GetDBInstance function
func TestMockDB_GetDBInstance(t *testing.T) {
	mockDB := &MockDB{}
	mockDB.On("GetDBInstance").Return(mockDB)

	instance := mockDB.GetDBInstance()
	assert.Equal(t, mockDB, instance, "GetDBInstance should return the same mockDB instance")

	mockDB.AssertExpectations(t)
}

// Test InitializeDBConnection function
func TestMockDB_InitializeDBConnection(t *testing.T) {
	mockDB := &MockDB{}
	mockDB.On("InitializeDBConnection").Return(nil)

	err := mockDB.InitializeDBConnection()
	assert.NoError(t, err)
	mockDB.AssertExpectations(t)
}

// Test MockRow Scan function
func TestMockRow_Scan(t *testing.T) {
	mockRow := &MockRow{values: []interface{}{"test_project"}}
	var name string

	err := mockRow.Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "test_project", name, "Scan should correctly assign values")
}
