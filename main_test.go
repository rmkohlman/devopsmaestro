package main

import (
	"context"
	"devopsmaestro/cmd"
	"devopsmaestro/db"
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockExecutor is a mock implementation of the cmd.Executor interface for testing purposes.
type MockExecutor struct {
	mock.Mock
}

var _ cmd.Executor = (*MockExecutor)(nil)

func (m *MockExecutor) Execute(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// testMigrationsFS creates an in-memory filesystem with test migrations
func testMigrationsFS() fs.FS {
	return fstest.MapFS{
		"sqlite/001_init.up.sql": &fstest.MapFile{
			Data: []byte(`
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    path TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`),
		},
	}
}

// TestRun tests the run function with mock dependencies
func TestRun(t *testing.T) {
	// Create a mock instance of the Database (legacy interface)
	mockDB := db.NewMockDB()
	// Create a mock instance of the DataStore
	mockDS := db.NewMockDataStore()
	// Create a mock instance of the Executor
	mockExecutor := new(MockExecutor)

	// Call the run function with the mock instances
	// Note: run() calls cmd.Execute which will parse args and run commands
	// Since we're testing with no args, it should just show help
	exitCode := run(mockDB, mockDS, mockExecutor, testMigrationsFS())

	// run should return 0 for success
	assert.Equal(t, 0, exitCode)
}
