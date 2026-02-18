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
	Ecosystems  map[string]*models.Ecosystem
	Domains     map[int]*models.Domain // keyed by ID for easier lookup
	Apps        map[int]*models.App    // keyed by ID for easier lookup
	Projects    map[string]*models.Project
	Workspaces  map[int]*models.Workspace
	Plugins     map[string]*models.NvimPluginDB
	Themes      map[string]*models.NvimThemeDB
	Credentials map[string]*models.CredentialDB // keyed by "scopeType:scopeID:name"
	ActiveTheme string
	Context     *models.Context

	// WorkspacePlugins maps workspaceID -> pluginIDs
	WorkspacePlugins map[int]map[int]bool

	// Mock driver (optional, for Driver() method)
	MockDriver *MockDriver

	// Error injection for testing error paths
	CreateEcosystemErr           error
	GetEcosystemByNameErr        error
	GetEcosystemByIDErr          error
	UpdateEcosystemErr           error
	DeleteEcosystemErr           error
	ListEcosystemsErr            error
	CreateDomainErr              error
	GetDomainByNameErr           error
	GetDomainByIDErr             error
	UpdateDomainErr              error
	DeleteDomainErr              error
	ListDomainsByEcosystemErr    error
	ListAllDomainsErr            error
	CreateAppErr                 error
	GetAppByNameErr              error
	GetAppByIDErr                error
	UpdateAppErr                 error
	DeleteAppErr                 error
	ListAppsByDomainErr          error
	ListAllAppsErr               error
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
	ListWorkspacesByAppErr       error
	ListAllWorkspacesErr         error
	FindWorkspacesErr            error
	GetContextErr                error
	SetActiveEcosystemErr        error
	SetActiveDomainErr           error
	SetActiveAppErr              error
	SetActiveWorkspaceErr        error
	SetActiveProjectErr          error
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
	CreateThemeErr               error
	GetThemeByNameErr            error
	GetThemeByIDErr              error
	UpdateThemeErr               error
	DeleteThemeErr               error
	ListThemesErr                error
	ListThemesByCategoryErr      error
	GetActiveThemeErr            error
	SetActiveThemeErr            error
	ClearActiveThemeErr          error
	CloseErr                     error
	PingErr                      error

	// Call tracking
	Calls []MockDataStoreCall

	// Auto-increment IDs
	nextEcosystemID int
	nextDomainID    int
	nextAppID       int
	nextProjectID   int
	nextWorkspaceID int
	nextPluginID    int
	nextThemeID     int
}

// MockDataStoreCall represents a recorded method call
type MockDataStoreCall struct {
	Method string
	Args   []interface{}
}

// NewMockDataStore creates a new mock data store with initialized storage
func NewMockDataStore() *MockDataStore {
	return &MockDataStore{
		Ecosystems:       make(map[string]*models.Ecosystem),
		Domains:          make(map[int]*models.Domain),
		Apps:             make(map[int]*models.App),
		Projects:         make(map[string]*models.Project),
		Workspaces:       make(map[int]*models.Workspace),
		Plugins:          make(map[string]*models.NvimPluginDB),
		Themes:           make(map[string]*models.NvimThemeDB),
		WorkspacePlugins: make(map[int]map[int]bool),
		Context:          &models.Context{ID: 1},
		MockDriver:       NewMockDriver(),
		nextEcosystemID:  1,
		nextDomainID:     1,
		nextAppID:        1,
		nextProjectID:    1,
		nextWorkspaceID:  1,
		nextPluginID:     1,
		nextThemeID:      1,
	}
}

func (m *MockDataStore) recordCall(method string, args ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, MockDataStoreCall{Method: method, Args: args})
}

// =============================================================================
// Ecosystem Operations
// =============================================================================

func (m *MockDataStore) CreateEcosystem(ecosystem *models.Ecosystem) error {
	m.recordCall("CreateEcosystem", ecosystem)
	if m.CreateEcosystemErr != nil {
		return m.CreateEcosystemErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	ecosystem.ID = m.nextEcosystemID
	m.nextEcosystemID++
	m.Ecosystems[ecosystem.Name] = ecosystem
	return nil
}

func (m *MockDataStore) GetEcosystemByName(name string) (*models.Ecosystem, error) {
	m.recordCall("GetEcosystemByName", name)
	if m.GetEcosystemByNameErr != nil {
		return nil, m.GetEcosystemByNameErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.Ecosystems[name]; ok {
		return e, nil
	}
	return nil, fmt.Errorf("ecosystem not found: %s", name)
}

func (m *MockDataStore) GetEcosystemByID(id int) (*models.Ecosystem, error) {
	m.recordCall("GetEcosystemByID", id)
	if m.GetEcosystemByIDErr != nil {
		return nil, m.GetEcosystemByIDErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.Ecosystems {
		if e.ID == id {
			return e, nil
		}
	}
	return nil, fmt.Errorf("ecosystem not found: %d", id)
}

func (m *MockDataStore) UpdateEcosystem(ecosystem *models.Ecosystem) error {
	m.recordCall("UpdateEcosystem", ecosystem)
	if m.UpdateEcosystemErr != nil {
		return m.UpdateEcosystemErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Ecosystems[ecosystem.Name] = ecosystem
	return nil
}

func (m *MockDataStore) DeleteEcosystem(name string) error {
	m.recordCall("DeleteEcosystem", name)
	if m.DeleteEcosystemErr != nil {
		return m.DeleteEcosystemErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Ecosystems, name)
	return nil
}

func (m *MockDataStore) ListEcosystems() ([]*models.Ecosystem, error) {
	m.recordCall("ListEcosystems")
	if m.ListEcosystemsErr != nil {
		return nil, m.ListEcosystemsErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var ecosystems []*models.Ecosystem
	for _, e := range m.Ecosystems {
		ecosystems = append(ecosystems, e)
	}
	return ecosystems, nil
}

// =============================================================================
// Domain Operations
// =============================================================================

func (m *MockDataStore) CreateDomain(domain *models.Domain) error {
	m.recordCall("CreateDomain", domain)
	if m.CreateDomainErr != nil {
		return m.CreateDomainErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	domain.ID = m.nextDomainID
	m.nextDomainID++
	m.Domains[domain.ID] = domain
	return nil
}

func (m *MockDataStore) GetDomainByName(ecosystemID int, name string) (*models.Domain, error) {
	m.recordCall("GetDomainByName", ecosystemID, name)
	if m.GetDomainByNameErr != nil {
		return nil, m.GetDomainByNameErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, d := range m.Domains {
		if d.EcosystemID == ecosystemID && d.Name == name {
			return d, nil
		}
	}
	return nil, fmt.Errorf("domain not found: %s", name)
}

func (m *MockDataStore) GetDomainByID(id int) (*models.Domain, error) {
	m.recordCall("GetDomainByID", id)
	if m.GetDomainByIDErr != nil {
		return nil, m.GetDomainByIDErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if d, ok := m.Domains[id]; ok {
		return d, nil
	}
	return nil, fmt.Errorf("domain not found: %d", id)
}

func (m *MockDataStore) UpdateDomain(domain *models.Domain) error {
	m.recordCall("UpdateDomain", domain)
	if m.UpdateDomainErr != nil {
		return m.UpdateDomainErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Domains[domain.ID] = domain
	return nil
}

func (m *MockDataStore) DeleteDomain(id int) error {
	m.recordCall("DeleteDomain", id)
	if m.DeleteDomainErr != nil {
		return m.DeleteDomainErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Domains, id)
	return nil
}

func (m *MockDataStore) ListDomainsByEcosystem(ecosystemID int) ([]*models.Domain, error) {
	m.recordCall("ListDomainsByEcosystem", ecosystemID)
	if m.ListDomainsByEcosystemErr != nil {
		return nil, m.ListDomainsByEcosystemErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var domains []*models.Domain
	for _, d := range m.Domains {
		if d.EcosystemID == ecosystemID {
			domains = append(domains, d)
		}
	}
	return domains, nil
}

func (m *MockDataStore) ListAllDomains() ([]*models.Domain, error) {
	m.recordCall("ListAllDomains")
	if m.ListAllDomainsErr != nil {
		return nil, m.ListAllDomainsErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var domains []*models.Domain
	for _, d := range m.Domains {
		domains = append(domains, d)
	}
	return domains, nil
}

// =============================================================================
// App Operations
// =============================================================================

func (m *MockDataStore) CreateApp(app *models.App) error {
	m.recordCall("CreateApp", app)
	if m.CreateAppErr != nil {
		return m.CreateAppErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	app.ID = m.nextAppID
	m.nextAppID++
	m.Apps[app.ID] = app
	return nil
}

func (m *MockDataStore) GetAppByName(domainID int, name string) (*models.App, error) {
	m.recordCall("GetAppByName", domainID, name)
	if m.GetAppByNameErr != nil {
		return nil, m.GetAppByNameErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.Apps {
		if a.DomainID == domainID && a.Name == name {
			return a, nil
		}
	}
	return nil, fmt.Errorf("app not found: %s", name)
}

func (m *MockDataStore) GetAppByNameGlobal(name string) (*models.App, error) {
	m.recordCall("GetAppByNameGlobal", name)
	if m.GetAppByNameErr != nil {
		return nil, m.GetAppByNameErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, a := range m.Apps {
		if a.Name == name {
			return a, nil
		}
	}
	return nil, fmt.Errorf("app not found: %s", name)
}

func (m *MockDataStore) GetAppByID(id int) (*models.App, error) {
	m.recordCall("GetAppByID", id)
	if m.GetAppByIDErr != nil {
		return nil, m.GetAppByIDErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if a, ok := m.Apps[id]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("app not found: %d", id)
}

func (m *MockDataStore) UpdateApp(app *models.App) error {
	m.recordCall("UpdateApp", app)
	if m.UpdateAppErr != nil {
		return m.UpdateAppErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Apps[app.ID] = app
	return nil
}

func (m *MockDataStore) DeleteApp(id int) error {
	m.recordCall("DeleteApp", id)
	if m.DeleteAppErr != nil {
		return m.DeleteAppErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Apps, id)
	return nil
}

func (m *MockDataStore) ListAppsByDomain(domainID int) ([]*models.App, error) {
	m.recordCall("ListAppsByDomain", domainID)
	if m.ListAppsByDomainErr != nil {
		return nil, m.ListAppsByDomainErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var apps []*models.App
	for _, a := range m.Apps {
		if a.DomainID == domainID {
			apps = append(apps, a)
		}
	}
	return apps, nil
}

func (m *MockDataStore) ListAllApps() ([]*models.App, error) {
	m.recordCall("ListAllApps")
	if m.ListAllAppsErr != nil {
		return nil, m.ListAllAppsErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var apps []*models.App
	for _, a := range m.Apps {
		apps = append(apps, a)
	}
	return apps, nil
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

func (m *MockDataStore) GetWorkspaceByName(appID int, name string) (*models.Workspace, error) {
	m.recordCall("GetWorkspaceByName", appID, name)
	if m.GetWorkspaceByNameErr != nil {
		return nil, m.GetWorkspaceByNameErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, ws := range m.Workspaces {
		if ws.AppID == appID && ws.Name == name {
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

func (m *MockDataStore) ListWorkspacesByApp(appID int) ([]*models.Workspace, error) {
	m.recordCall("ListWorkspacesByApp", appID)
	if m.ListWorkspacesByAppErr != nil {
		return nil, m.ListWorkspacesByAppErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var workspaces []*models.Workspace
	for _, ws := range m.Workspaces {
		if ws.AppID == appID {
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

func (m *MockDataStore) FindWorkspaces(filter models.WorkspaceFilter) ([]*models.WorkspaceWithHierarchy, error) {
	m.recordCall("FindWorkspaces", filter)
	if m.FindWorkspacesErr != nil {
		return nil, m.FindWorkspacesErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	var results []*models.WorkspaceWithHierarchy

	for _, ws := range m.Workspaces {
		// Get the app for this workspace
		app, ok := m.Apps[ws.AppID]
		if !ok {
			continue
		}

		// Get the domain for this app
		domain, ok := m.Domains[app.DomainID]
		if !ok {
			continue
		}

		// Get the ecosystem for this domain
		var ecosystem *models.Ecosystem
		for _, e := range m.Ecosystems {
			if e.ID == domain.EcosystemID {
				ecosystem = e
				break
			}
		}
		if ecosystem == nil {
			continue
		}

		// Apply filters
		if filter.EcosystemName != "" && ecosystem.Name != filter.EcosystemName {
			continue
		}
		if filter.DomainName != "" && domain.Name != filter.DomainName {
			continue
		}
		if filter.AppName != "" && app.Name != filter.AppName {
			continue
		}
		if filter.WorkspaceName != "" && ws.Name != filter.WorkspaceName {
			continue
		}

		results = append(results, &models.WorkspaceWithHierarchy{
			Workspace: ws,
			App:       app,
			Domain:    domain,
			Ecosystem: ecosystem,
		})
	}

	return results, nil
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

func (m *MockDataStore) SetActiveEcosystem(ecosystemID *int) error {
	m.recordCall("SetActiveEcosystem", ecosystemID)
	if m.SetActiveEcosystemErr != nil {
		return m.SetActiveEcosystemErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Context.ActiveEcosystemID = ecosystemID
	return nil
}

func (m *MockDataStore) SetActiveDomain(domainID *int) error {
	m.recordCall("SetActiveDomain", domainID)
	if m.SetActiveDomainErr != nil {
		return m.SetActiveDomainErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Context.ActiveDomainID = domainID
	return nil
}

func (m *MockDataStore) SetActiveApp(appID *int) error {
	m.recordCall("SetActiveApp", appID)
	if m.SetActiveAppErr != nil {
		return m.SetActiveAppErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Context.ActiveAppID = appID
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
// Theme Operations
// =============================================================================

func (m *MockDataStore) CreateTheme(theme *models.NvimThemeDB) error {
	m.recordCall("CreateTheme", theme)
	if m.CreateThemeErr != nil {
		return m.CreateThemeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	theme.ID = m.nextThemeID
	m.nextThemeID++
	m.Themes[theme.Name] = theme
	return nil
}

func (m *MockDataStore) GetThemeByName(name string) (*models.NvimThemeDB, error) {
	m.recordCall("GetThemeByName", name)
	if m.GetThemeByNameErr != nil {
		return nil, m.GetThemeByNameErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if t, ok := m.Themes[name]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("theme not found: %s", name)
}

func (m *MockDataStore) GetThemeByID(id int) (*models.NvimThemeDB, error) {
	m.recordCall("GetThemeByID", id)
	if m.GetThemeByIDErr != nil {
		return nil, m.GetThemeByIDErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, t := range m.Themes {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("theme not found: %d", id)
}

func (m *MockDataStore) UpdateTheme(theme *models.NvimThemeDB) error {
	m.recordCall("UpdateTheme", theme)
	if m.UpdateThemeErr != nil {
		return m.UpdateThemeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Themes[theme.Name] = theme
	return nil
}

func (m *MockDataStore) DeleteTheme(name string) error {
	m.recordCall("DeleteTheme", name)
	if m.DeleteThemeErr != nil {
		return m.DeleteThemeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Themes, name)
	return nil
}

func (m *MockDataStore) ListThemes() ([]*models.NvimThemeDB, error) {
	m.recordCall("ListThemes")
	if m.ListThemesErr != nil {
		return nil, m.ListThemesErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var themes []*models.NvimThemeDB
	for _, t := range m.Themes {
		themes = append(themes, t)
	}
	return themes, nil
}

func (m *MockDataStore) ListThemesByCategory(category string) ([]*models.NvimThemeDB, error) {
	m.recordCall("ListThemesByCategory", category)
	if m.ListThemesByCategoryErr != nil {
		return nil, m.ListThemesByCategoryErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var themes []*models.NvimThemeDB
	for _, t := range m.Themes {
		if t.Category.Valid && t.Category.String == category {
			themes = append(themes, t)
		}
	}
	return themes, nil
}

func (m *MockDataStore) GetActiveTheme() (*models.NvimThemeDB, error) {
	m.recordCall("GetActiveTheme")
	if m.GetActiveThemeErr != nil {
		return nil, m.GetActiveThemeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.ActiveTheme == "" {
		return nil, nil
	}
	if t, ok := m.Themes[m.ActiveTheme]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("active theme not found: %s", m.ActiveTheme)
}

func (m *MockDataStore) SetActiveTheme(name string) error {
	m.recordCall("SetActiveTheme", name)
	if m.SetActiveThemeErr != nil {
		return m.SetActiveThemeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	// Verify theme exists
	if _, ok := m.Themes[name]; !ok {
		return fmt.Errorf("theme not found: %s", name)
	}
	m.ActiveTheme = name
	return nil
}

func (m *MockDataStore) ClearActiveTheme() error {
	m.recordCall("ClearActiveTheme")
	if m.ClearActiveThemeErr != nil {
		return m.ClearActiveThemeErr
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ActiveTheme = ""
	return nil
}

// =============================================================================
// Credential Operations
// =============================================================================

func (m *MockDataStore) credentialKey(scopeType models.CredentialScopeType, scopeID int64, name string) string {
	return fmt.Sprintf("%s:%d:%s", scopeType, scopeID, name)
}

func (m *MockDataStore) CreateCredential(credential *models.CredentialDB) error {
	m.recordCall("CreateCredential", credential)
	if m.Credentials == nil {
		m.Credentials = make(map[string]*models.CredentialDB)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.credentialKey(credential.ScopeType, credential.ScopeID, credential.Name)
	if _, exists := m.Credentials[key]; exists {
		return fmt.Errorf("credential already exists: %s", credential.Name)
	}
	m.Credentials[key] = credential
	return nil
}

func (m *MockDataStore) GetCredential(scopeType models.CredentialScopeType, scopeID int64, name string) (*models.CredentialDB, error) {
	m.recordCall("GetCredential", scopeType, scopeID, name)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Credentials == nil {
		return nil, fmt.Errorf("credential not found: %s", name)
	}
	key := m.credentialKey(scopeType, scopeID, name)
	if cred, ok := m.Credentials[key]; ok {
		return cred, nil
	}
	return nil, fmt.Errorf("credential not found: %s", name)
}

func (m *MockDataStore) UpdateCredential(credential *models.CredentialDB) error {
	m.recordCall("UpdateCredential", credential)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Credentials == nil {
		return fmt.Errorf("credential not found: %s", credential.Name)
	}
	key := m.credentialKey(credential.ScopeType, credential.ScopeID, credential.Name)
	if _, exists := m.Credentials[key]; !exists {
		return fmt.Errorf("credential not found: %s", credential.Name)
	}
	m.Credentials[key] = credential
	return nil
}

func (m *MockDataStore) DeleteCredential(scopeType models.CredentialScopeType, scopeID int64, name string) error {
	m.recordCall("DeleteCredential", scopeType, scopeID, name)
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.Credentials == nil {
		return fmt.Errorf("credential not found: %s", name)
	}
	key := m.credentialKey(scopeType, scopeID, name)
	if _, exists := m.Credentials[key]; !exists {
		return fmt.Errorf("credential not found: %s", name)
	}
	delete(m.Credentials, key)
	return nil
}

func (m *MockDataStore) ListCredentialsByScope(scopeType models.CredentialScopeType, scopeID int64) ([]*models.CredentialDB, error) {
	m.recordCall("ListCredentialsByScope", scopeType, scopeID)
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*models.CredentialDB
	prefix := fmt.Sprintf("%s:%d:", scopeType, scopeID)
	for key, cred := range m.Credentials {
		if len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			result = append(result, cred)
		}
	}
	return result, nil
}

func (m *MockDataStore) ListAllCredentials() ([]*models.CredentialDB, error) {
	m.recordCall("ListAllCredentials")
	m.mu.Lock()
	defer m.mu.Unlock()
	var result []*models.CredentialDB
	for _, cred := range m.Credentials {
		result = append(result, cred)
	}
	return result, nil
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
	m.Ecosystems = make(map[string]*models.Ecosystem)
	m.Domains = make(map[int]*models.Domain)
	m.Apps = make(map[int]*models.App)
	m.Projects = make(map[string]*models.Project)
	m.Workspaces = make(map[int]*models.Workspace)
	m.Plugins = make(map[string]*models.NvimPluginDB)
	m.Themes = make(map[string]*models.NvimThemeDB)
	m.Credentials = make(map[string]*models.CredentialDB)
	m.WorkspacePlugins = make(map[int]map[int]bool)
	m.ActiveTheme = ""
	m.Context = &models.Context{ID: 1}
	m.Calls = nil
	m.nextEcosystemID = 1
	m.nextDomainID = 1
	m.nextAppID = 1
	m.nextProjectID = 1
	m.nextWorkspaceID = 1
	m.nextPluginID = 1
	m.nextThemeID = 1
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
