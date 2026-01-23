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

// Plugin CRUD operations (stub implementations for interface compliance)

// CreatePlugin inserts a new plugin into the database.
func (m *MockDataStore) CreatePlugin(plugin *models.NvimPluginDB) error {
	return nil
}

// GetPluginByName retrieves a plugin by its name.
func (m *MockDataStore) GetPluginByName(name string) (*models.NvimPluginDB, error) {
	return nil, nil
}

// UpdatePlugin updates an existing plugin.
func (m *MockDataStore) UpdatePlugin(plugin *models.NvimPluginDB) error {
	return nil
}

// ListPlugins retrieves all plugins.
func (m *MockDataStore) ListPlugins() ([]*models.NvimPluginDB, error) {
	return []*models.NvimPluginDB{}, nil
}

// ListPluginsByCategory retrieves plugins by category.
func (m *MockDataStore) ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error) {
	return []*models.NvimPluginDB{}, nil
}

// DeletePlugin deletes a plugin by name.
func (m *MockDataStore) DeletePlugin(name string) error {
	return nil
}

// AddPluginToWorkspace associates a plugin with a workspace.
func (m *MockDataStore) AddPluginToWorkspace(workspaceID int, pluginID int) error {
	return nil
}

// GetWorkspacePlugins retrieves all plugins for a workspace.
func (m *MockDataStore) GetWorkspacePlugins(workspaceID int) ([]*models.NvimPluginDB, error) {
	return []*models.NvimPluginDB{}, nil
}

// RemovePluginFromWorkspace removes a plugin from a workspace.
func (m *MockDataStore) RemovePluginFromWorkspace(workspaceID int, pluginID int) error {
	return nil
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
