package cmd

import (
	"context"
	"devopsmaestro/db"
	"fmt"

	"github.com/stretchr/testify/mock"
)

type Executor interface {
	Execute(ctx context.Context) error
}

type DefaultExecutor struct {
	DataStore db.DataStore
	Database  db.Database
}

func (e *DefaultExecutor) Execute(ctx context.Context) error {
	// Access the dataStore and perform some actions
	if e.DataStore == nil {
		return fmt.Errorf("datastore is not initialized")
	}

	// Example logic: perform operations using DataStore
	fmt.Println("Executing command with datastore...")
	// Insert business logic here

	return nil
}

func NewExecutor(dataStore db.DataStore, dbInstance db.Database) Executor {
	return &DefaultExecutor{
		DataStore: dataStore,
		Database:  dbInstance,
	}
}

// MockExecutor is a mock implementation of the Executor interface.
type MockExecutor struct {
	mock.Mock
}

func (m *MockExecutor) Execute(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
