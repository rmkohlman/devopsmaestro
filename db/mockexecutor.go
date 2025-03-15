package db

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockExecutor is a mock implementation of the Executor interface
type MockExecutor struct {
	mock.Mock
}

// Execute is the mock implementation of the Execute method
func (m *MockExecutor) Execute(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
