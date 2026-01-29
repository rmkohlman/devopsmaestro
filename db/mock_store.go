package db

import (
	"fmt"
	"sync"

	"devopsmaestro/models"
)

// MockDataStore implements the DataStore interface for testing.
// It stores data in memory and records all method calls.
type MockDataStore struct {
	mu sync.Mutex

	// In-memory storage
	Projects   map[string]*models.Project
	Workspaces map[int]*models.Workspace
	Plugins    map[string]*models.NvimPluginDB
	Context    *models.Context

	// WorkspacePlugins maps workspaceID -> pluginIDs
	WorkspacePlugins map[int]map[int]bool

	// Mock driver (optional, for Driver() method)
	MockDriver *MockDriver

	// Error injection for testing error paths
	CreateProjectErr             error
	GetProjectByNameErr          error
	GetProjectByIDErr            error
	UpdateProjectErr             error
	DeleteProjectErr             error
	ListProjectsErr              error
	CreateWorkspaceErr           error
	GetWorkspaceByNameErr        error
	GetWorkspaceByIDErr          error
	UpdateWorkspaceErr           error
	DeleteWorkspaceErr           error
	ListWorkspacesByProjectErr   error
	ListAllWorkspacesErr         error
	GetContextErr                error
	SetActiveProjectErr          error
	SetActiveWorkspaceErr        error
	CreatePluginErr              error
	GetPluginByNameErr           error
	GetPluginByIDErr             error
	UpdatePluginErr              error
	DeletePluginErr              error
	ListPluginsErr               error
	ListPluginsByCategoryErr     error
	ListPluginsByTagsErr         error
	AddPluginToWorkspaceErr      error
	RemovePluginFromWorkspaceErr error
	GetWorkspacePluginsErr       error
	SetWorkspacePluginEnabledErr error
	CloseErr                     error
	PingErr                      error

	// Call tracking
	Calls []MockDataStoreCall

	// Auto-increment IDs
	nextProjectID   int
	nextWorkspaceID int
	nextPluginID    int
}

// MockDataStoreCall represents a recorded method call
type MockDataStoreCall struct {
	Method string
	Args   []interface{}
}

// NewMockDataStore creates a new mock data store with initialized storage
func NewMockDataStore() *MockDataStore {
	return &MockDataStore{
		Projects:         make(map[string]*models.Project),
		Workspaces:       make(map[int]*models.Workspace),
		Plugins:          make(map[string]*models.NvimPluginDB),
		WorkspacePlugins: make(map[int]map[int]bool),
		Context:          &models.Context{ID: 1},
		MockDriver:       NewMockDriver(),
		nextProjectID:    1,
		nextWorkspaceID:  1,
		nextPluginID:     1,
	}
}

func (m *MockDataStore) recordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockDataStoreCall{Method: method, Args: args})
}

// =============================================================================
// Project Operations
// =============================================================================

func (m *MockDataStore) CreateProject(project *models.Project) error {
	m.recordCall("CreateProject", project)
	if m.CreateProjectErr != nil {
		return m.CreateProjectErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	project.ID = m.nextProjectID
	m.nextProjectID++
	m.Projects[project.Name] = project
	return nil
}

func (m *MockDataStore) GetProjectByName(name string) (*models.Project, error) {
	m.recordCall("GetProjectByName", name)
	if m.GetProjectByNameErr != nil {
		return nil, m.GetProjectByNameErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.Projects[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("project not found: %s", name)
}

func (m *MockDataStore) GetProjectByID(id int) (*models.Project, error) {
	m.recordCall("GetProjectByID", id)
	if m.GetProjectByIDErr != nil {
		return nil, m.GetProjectByIDErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.Projects {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, fmt.Errorf("project not found: %d", id)
}

func (m *MockDataStore) UpdateProject(project *models.Project) error {
	m.recordCall("UpdateProject", project)
	if m.UpdateProjectErr != nil {
		return m.UpdateProjectErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Projects[project.Name] = project
	return nil
}

func (m *MockDataStore) DeleteProject(name string) error {
	m.recordCall("DeleteProject", name)
	if m.DeleteProjectErr != nil {
		return m.DeleteProjectErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Projects, name)
	return nil
}

func (m *MockDataStore) ListProjects() ([]*models.Project, error) {
	m.recordCall("ListProjects")
	if m.ListProjectsErr != nil {
		return nil, m.ListProjectsErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var projects []*models.Project
	for _, p := range m.Projects {
		projects = append(projects, p)
	}
	return projects, nil
}

// =============================================================================
// Workspace Operations
// =============================================================================

func (m *MockDataStore) CreateWorkspace(workspace *models.Workspace) error {
	m.recordCall("CreateWorkspace", workspace)
	if m.CreateWorkspaceErr != nil {
		return m.CreateWorkspaceErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	workspace.ID = m.nextWorkspaceID
	m.nextWorkspaceID++
	m.Workspaces[workspace.ID] = workspace
	return nil
}

func (m *MockDataStore) GetWorkspaceByName(projectID int, name string) (*models.Workspace, error) {
	m.recordCall("GetWorkspaceByName", projectID, name)
	if m.GetWorkspaceByNameErr != nil {
		return nil, m.GetWorkspaceByNameErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ws := range m.Workspaces {
		if ws.ProjectID == projectID && ws.Name == name {
			return ws, nil
		}
	}
	return nil, fmt.Errorf("workspace not found: %s", name)
}

func (m *MockDataStore) GetWorkspaceByID(id int) (*models.Workspace, error) {
	m.recordCall("GetWorkspaceByID", id)
	if m.GetWorkspaceByIDErr != nil {
		return nil, m.GetWorkspaceByIDErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if ws, ok := m.Workspaces[id]; ok {
		return ws, nil
	}
	return nil, fmt.Errorf("workspace not found: %d", id)
}

func (m *MockDataStore) UpdateWorkspace(workspace *models.Workspace) error {
	m.recordCall("UpdateWorkspace", workspace)
	if m.UpdateWorkspaceErr != nil {
		return m.UpdateWorkspaceErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Workspaces[workspace.ID] = workspace
	return nil
}

func (m *MockDataStore) DeleteWorkspace(id int) error {
	m.recordCall("DeleteWorkspace", id)
	if m.DeleteWorkspaceErr != nil {
		return m.DeleteWorkspaceErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Workspaces, id)
	return nil
}

func (m *MockDataStore) ListWorkspacesByProject(projectID int) ([]*models.Workspace, error) {
	m.recordCall("ListWorkspacesByProject", projectID)
	if m.ListWorkspacesByProjectErr != nil {
		return nil, m.ListWorkspacesByProjectErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var workspaces []*models.Workspace
	for _, ws := range m.Workspaces {
		if ws.ProjectID == projectID {
			workspaces = append(workspaces, ws)
		}
	}
	return workspaces, nil
}

func (m *MockDataStore) ListAllWorkspaces() ([]*models.Workspace, error) {
	m.recordCall("ListAllWorkspaces")
	if m.ListAllWorkspacesErr != nil {
		return nil, m.ListAllWorkspacesErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var workspaces []*models.Workspace
	for _, ws := range m.Workspaces {
		workspaces = append(workspaces, ws)
	}
	return workspaces, nil
}

// =============================================================================
// Context Operations
// =============================================================================

func (m *MockDataStore) GetContext() (*models.Context, error) {
	m.recordCall("GetContext")
	if m.GetContextErr != nil {
		return nil, m.GetContextErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Context, nil
}

func (m *MockDataStore) SetActiveProject(projectID *int) error {
	m.recordCall("SetActiveProject", projectID)
	if m.SetActiveProjectErr != nil {
		return m.SetActiveProjectErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Context.ActiveProjectID = projectID
	return nil
}

func (m *MockDataStore) SetActiveWorkspace(workspaceID *int) error {
	m.recordCall("SetActiveWorkspace", workspaceID)
	if m.SetActiveWorkspaceErr != nil {
		return m.SetActiveWorkspaceErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Context.ActiveWorkspaceID = workspaceID
	return nil
}

// =============================================================================
// Plugin Operations
// =============================================================================

func (m *MockDataStore) CreatePlugin(plugin *models.NvimPluginDB) error {
	m.recordCall("CreatePlugin", plugin)
	if m.CreatePluginErr != nil {
		return m.CreatePluginErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	plugin.ID = m.nextPluginID
	m.nextPluginID++
	m.Plugins[plugin.Name] = plugin
	return nil
}

func (m *MockDataStore) GetPluginByName(name string) (*models.NvimPluginDB, error) {
	m.recordCall("GetPluginByName", name)
	if m.GetPluginByNameErr != nil {
		return nil, m.GetPluginByNameErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if p, ok := m.Plugins[name]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("plugin not found: %s", name)
}

func (m *MockDataStore) GetPluginByID(id int) (*models.NvimPluginDB, error) {
	m.recordCall("GetPluginByID", id)
	if m.GetPluginByIDErr != nil {
		return nil, m.GetPluginByIDErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, p := range m.Plugins {
		if p.ID == id {
			return p, nil
		}
	}
	return nil, fmt.Errorf("plugin not found: %d", id)
}

func (m *MockDataStore) UpdatePlugin(plugin *models.NvimPluginDB) error {
	m.recordCall("UpdatePlugin", plugin)
	if m.UpdatePluginErr != nil {
		return m.UpdatePluginErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Plugins[plugin.Name] = plugin
	return nil
}

func (m *MockDataStore) DeletePlugin(name string) error {
	m.recordCall("DeletePlugin", name)
	if m.DeletePluginErr != nil {
		return m.DeletePluginErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Plugins, name)
	return nil
}

func (m *MockDataStore) ListPlugins() ([]*models.NvimPluginDB, error) {
	m.recordCall("ListPlugins")
	if m.ListPluginsErr != nil {
		return nil, m.ListPluginsErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var plugins []*models.NvimPluginDB
	for _, p := range m.Plugins {
		plugins = append(plugins, p)
	}
	return plugins, nil
}

func (m *MockDataStore) ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error) {
	m.recordCall("ListPluginsByCategory", category)
	if m.ListPluginsByCategoryErr != nil {
		return nil, m.ListPluginsByCategoryErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var plugins []*models.NvimPluginDB
	for _, p := range m.Plugins {
		if p.Category.Valid && p.Category.String == category {
			plugins = append(plugins, p)
		}
	}
	return plugins, nil
}

func (m *MockDataStore) ListPluginsByTags(tags []string) ([]*models.NvimPluginDB, error) {
	m.recordCall("ListPluginsByTags", tags)
	if m.ListPluginsByTagsErr != nil {
		return nil, m.ListPluginsByTagsErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var plugins []*models.NvimPluginDB
	for _, p := range m.Plugins {
		// Simple contains check for tags
		if p.Tags.Valid {
			for _, tag := range tags {
				if contains(p.Tags.String, tag) {
					plugins = append(plugins, p)
					break
				}
			}
		}
	}
	return plugins, nil
}

// =============================================================================
// Workspace Plugin Associations
// =============================================================================

func (m *MockDataStore) AddPluginToWorkspace(workspaceID int, pluginID int) error {
	m.recordCall("AddPluginToWorkspace", workspaceID, pluginID)
	if m.AddPluginToWorkspaceErr != nil {
		return m.AddPluginToWorkspaceErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.WorkspacePlugins[workspaceID] == nil {
		m.WorkspacePlugins[workspaceID] = make(map[int]bool)
	}
	m.WorkspacePlugins[workspaceID][pluginID] = true
	return nil
}

func (m *MockDataStore) RemovePluginFromWorkspace(workspaceID int, pluginID int) error {
	m.recordCall("RemovePluginFromWorkspace", workspaceID, pluginID)
	if m.RemovePluginFromWorkspaceErr != nil {
		return m.RemovePluginFromWorkspaceErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.WorkspacePlugins[workspaceID] != nil {
		delete(m.WorkspacePlugins[workspaceID], pluginID)
	}
	return nil
}

func (m *MockDataStore) GetWorkspacePlugins(workspaceID int) ([]*models.NvimPluginDB, error) {
	m.recordCall("GetWorkspacePlugins", workspaceID)
	if m.GetWorkspacePluginsErr != nil {
		return nil, m.GetWorkspacePluginsErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var plugins []*models.NvimPluginDB
	if pluginIDs, ok := m.WorkspacePlugins[workspaceID]; ok {
		for pluginID, enabled := range pluginIDs {
			if enabled {
				for _, p := range m.Plugins {
					if p.ID == pluginID {
						plugins = append(plugins, p)
						break
					}
				}
			}
		}
	}
	return plugins, nil
}

func (m *MockDataStore) SetWorkspacePluginEnabled(workspaceID int, pluginID int, enabled bool) error {
	m.recordCall("SetWorkspacePluginEnabled", workspaceID, pluginID, enabled)
	if m.SetWorkspacePluginEnabledErr != nil {
		return m.SetWorkspacePluginEnabledErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.WorkspacePlugins[workspaceID] != nil {
		m.WorkspacePlugins[workspaceID][pluginID] = enabled
	}
	return nil
}

// =============================================================================
// Driver Access & Health
// =============================================================================

func (m *MockDataStore) Driver() Driver {
	return m.MockDriver
}

func (m *MockDataStore) Close() error {
	m.recordCall("Close")
	if m.CloseErr != nil {
		return m.CloseErr
	}
	return nil
}

func (m *MockDataStore) Ping() error {
	m.recordCall("Ping")
	if m.PingErr != nil {
		return m.PingErr
	}
	return nil
}

// =============================================================================
// Test Helpers
// =============================================================================

// GetCalls returns all recorded calls
func (m *MockDataStore) GetCalls() []MockDataStoreCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]MockDataStoreCall{}, m.Calls...)
}

// ResetCalls clears recorded calls
func (m *MockDataStore) ResetCalls() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = nil
}

// Reset clears all data and calls
func (m *MockDataStore) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Projects = make(map[string]*models.Project)
	m.Workspaces = make(map[int]*models.Workspace)
	m.Plugins = make(map[string]*models.NvimPluginDB)
	m.WorkspacePlugins = make(map[int]map[int]bool)
	m.Context = &models.Context{ID: 1}
	m.Calls = nil
	m.nextProjectID = 1
	m.nextWorkspaceID = 1
	m.nextPluginID = 1
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Ensure MockDataStore implements DataStore
var _ DataStore = (*MockDataStore)(nil)
