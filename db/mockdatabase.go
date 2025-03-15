package db

import (
	"database/sql"

	"github.com/stretchr/testify/mock"
)

type MockDB struct {
	mock.Mock
}

var mockDBInstance *MockDB

func SetMockDBInstance(db *MockDB) {
	mockDBInstance = db
}

func NewMockDB() (Database, error) {
	if mockDBInstance == nil {
		SetMockDBInstance(&MockDB{})
	}
	return mockDBInstance, nil
}

func (m *MockDB) InitializeDBConnection() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) GetDBInstance() Database {
	args := m.Called()
	return args.Get(0).(Database)
}

func (m *MockDB) Connect() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) Execute(query string, args ...interface{}) (interface{}, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0), arguments.Error(1)
}

func (m *MockDB) FetchOne(query string, args ...interface{}) (interface{}, error) {
	arguments := m.Called(query, args)
	row := sql.Row{}
	// Populate the row with mock data if needed
	return &row, arguments.Error(1)
}

func (m *MockDB) FetchMany(query string, args ...interface{}) (interface{}, error) {
	arguments := m.Called(query, args)
	return arguments.Get(0), arguments.Error(1)
}

func (m *MockDB) Close() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockDB) DSN() string {

	args := m.Called()
	return args.String(0)
}

func (m *MockDB) MigrationDSN() string {
	args := m.Called()
	return args.String(0)
}

// MockRow implements the sql.Row interface
type MockRow struct {
	values []interface{}
}

func (r *MockRow) Scan(dest ...interface{}) error {
	if len(dest) != len(r.values) {
		return sql.ErrNoRows // Return an error if the scan fields do not match
	}
	for i, v := range r.values {
		switch d := dest[i].(type) {
		case *string:
			*d = v.(string)
		case *int:
			*d = v.(int)
		case *float64:
			*d = v.(float64)
		default:
			return sql.ErrNoRows
		}
	}
	return nil
}
