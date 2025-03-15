package db

import (
	"devopsmaestro/models"
)

// SQLMockDataStore is a mock implementation of the SQLDataStore interface.
type MockDataStore struct {
	DB *MockDB
}

// CreateProject inserts a new project into the mock database.
func (m *MockDataStore) CreateProject(project *models.Project) error {
	args := m.DB.Called(project)
	return args.Error(0)
}

// GetProjectByName retrieves a project by its name from the mock database.
func (m *MockDataStore) GetProjectByName(name string) (*models.Project, error) {
	args := m.DB.Called(name)
	return args.Get(0).(*models.Project), args.Error(1)
}

// UpdateProject updates an existing project in the mock database.
func (m *MockDataStore) UpdateProject(project *models.Project) error {
	args := m.DB.Called(project)
	return args.Error(0)
}

// ListProjects retrieves all projects from the mock database.
func (m *MockDataStore) ListProjects() ([]*models.Project, error) {
	args := m.DB.Called()
	return args.Get(0).([]*models.Project), args.Error(1)
}

var mockDStoreInstance *MockDataStore

func SetMockDSInstance(ds *MockDataStore) {
	mockDStoreInstance = ds
}

func NewMockDataStore() *MockDataStore {
	if mockDStoreInstance == nil {
		if mockDBInstance == nil {
			SetMockDBInstance(&MockDB{})
		}
		SetMockDSInstance(&MockDataStore{mockDBInstance})
	}
	return mockDStoreInstance
}
