package main

import (
	"context"
	"devopsmaestro/cmd"
	"devopsmaestro/db"
	"testing"

	"github.com/spf13/cobra"
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

// write a test to test the main.go run function
func TestRun(t *testing.T) {
	// Create a mock instance of the Database
	mockDB := new(db.MockDB)
	// Create a mock instance of the DataStore
	mockDS := new(db.MockDataStore)
	// Create a mock instance of the Executor
	mockExecutor := new(MockExecutor)

	mockExecutor.On("&Execute", mock.Anything).Return(nil)
	// Call the run function with the mock instances
	run(mockDB, mockDS, mockExecutor)

	// Assert that the expectations were met
	mockExecutor.AssertExpectations(t)
}

func TestMainFunction(t *testing.T) {
	// Mock the os.Exit function
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	exitCode := 0
	osExit = func(code int) {
		exitCode = code
	}

	// Mock the db.InitializeDBConnection function
	oldInitializeDBConnection := db.InitializeDBConnection
	defer func() { db.InitializeDBConnection = oldInitializeDBConnection }()
	db.InitializeDBConnection = func() (db.Database, error) {
		return &db.MockDB{}, nil
	}

	// Mock the db.StoreFactory function
	oldStoreFactory := db.StoreFactory
	defer func() { db.StoreFactory = oldStoreFactory }()
	db.StoreFactory = func(dbInstance db.Database) (db.DataStore, error) {
		return &db.MockDataStore{}, nil
	}

	// Mock the cmd.NewExecutor function
	oldNewExecutor := cmd.NewExecutor
	defer func() { cmd.NewExecutor = oldNewExecutor }()
	cmd.NewExecutor = func(dataStore db.DataStore, dbInstance db.Database) cmd.Executor {
		return &MockExecutor{}
	}

	// Add a simple command to rootCmd that calls executor.Execute
	cmd.RootCmd.AddCommand(&cobra.Command{
		Use:   "test",
		Short: "Test command",
		RunE: func(cmd *cobra.Command, args []string) error {
			executor := cmd.Context().Value("executor").(cmd.Executor)
			return executor.Execute(cmd.Context())
		},
	})

	main()

	assert.Equal(t, 0, exitCode)
}
