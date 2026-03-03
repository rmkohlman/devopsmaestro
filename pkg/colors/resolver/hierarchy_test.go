package resolver

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops/theme"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDataStore implements db.DataStore for testing
type MockDataStore struct {
	ecosystems map[int]*models.Ecosystem
	domains    map[int]*models.Domain
	apps       map[int]*models.App
	workspaces map[int]*models.Workspace
	defaults   map[string]string

	// Error simulation
	getEcosystemError bool
	getDomainError    bool
	getAppError       bool
	getWorkspaceError bool
}

func NewMockDataStore() *MockDataStore {
	return &MockDataStore{
		ecosystems: make(map[int]*models.Ecosystem),
		domains:    make(map[int]*models.Domain),
		apps:       make(map[int]*models.App),
		workspaces: make(map[int]*models.Workspace),
		defaults:   make(map[string]string),
	}
}

func (m *MockDataStore) AddEcosystem(id int, name string, theme *string) {
	ecosystem := &models.Ecosystem{
		ID:   id,
		Name: name,
	}
	if theme != nil {
		ecosystem.Theme = sql.NullString{String: *theme, Valid: true}
	}
	m.ecosystems[id] = ecosystem
}

func (m *MockDataStore) AddDomain(id, ecosystemID int, name string, theme *string) {
	domain := &models.Domain{
		ID:          id,
		EcosystemID: ecosystemID,
		Name:        name,
	}
	if theme != nil {
		domain.Theme = sql.NullString{String: *theme, Valid: true}
	}
	m.domains[id] = domain
}

func (m *MockDataStore) AddApp(id, domainID int, name string, theme *string) {
	app := &models.App{
		ID:       id,
		DomainID: domainID,
		Name:     name,
	}
	if theme != nil {
		app.Theme = sql.NullString{String: *theme, Valid: true}
	}
	m.apps[id] = app
}

func (m *MockDataStore) AddWorkspace(id, appID int, name string) {
	workspace := &models.Workspace{
		ID:    id,
		AppID: appID,
		Name:  name,
	}
	m.workspaces[id] = workspace
}

func (m *MockDataStore) AddWorkspaceWithNvimStructure(id, appID int, name, nvimStructure string) {
	workspace := &models.Workspace{
		ID:            id,
		AppID:         appID,
		Name:          name,
		NvimStructure: sql.NullString{String: nvimStructure, Valid: true},
	}
	m.workspaces[id] = workspace
}

// Implement required DataStore methods for testing
func (m *MockDataStore) GetEcosystemByID(id int) (*models.Ecosystem, error) {
	if m.getEcosystemError {
		return nil, errors.New("mock ecosystem error")
	}
	if ecosystem, ok := m.ecosystems[id]; ok {
		return ecosystem, nil
	}
	return nil, errors.New("ecosystem not found")
}

func (m *MockDataStore) GetDomainByID(id int) (*models.Domain, error) {
	if m.getDomainError {
		return nil, errors.New("mock domain error")
	}
	if domain, ok := m.domains[id]; ok {
		return domain, nil
	}
	return nil, errors.New("domain not found")
}

func (m *MockDataStore) GetAppByID(id int) (*models.App, error) {
	if m.getAppError {
		return nil, errors.New("mock app error")
	}
	if app, ok := m.apps[id]; ok {
		return app, nil
	}
	return nil, errors.New("app not found")
}

func (m *MockDataStore) GetWorkspaceByID(id int) (*models.Workspace, error) {
	if m.getWorkspaceError {
		return nil, errors.New("mock workspace error")
	}
	if workspace, ok := m.workspaces[id]; ok {
		return workspace, nil
	}
	return nil, errors.New("workspace not found")
}

// Stub implementations for other required methods
func (m *MockDataStore) CreateEcosystem(ecosystem *models.Ecosystem) error         { return nil }
func (m *MockDataStore) GetEcosystemByName(name string) (*models.Ecosystem, error) { return nil, nil }
func (m *MockDataStore) UpdateEcosystem(ecosystem *models.Ecosystem) error         { return nil }
func (m *MockDataStore) DeleteEcosystem(name string) error                         { return nil }
func (m *MockDataStore) ListEcosystems() ([]*models.Ecosystem, error)              { return nil, nil }

func (m *MockDataStore) CreateDomain(domain *models.Domain) error { return nil }
func (m *MockDataStore) GetDomainByName(ecosystemID int, name string) (*models.Domain, error) {
	return nil, nil
}
func (m *MockDataStore) UpdateDomain(domain *models.Domain) error { return nil }
func (m *MockDataStore) DeleteDomain(id int) error                { return nil }
func (m *MockDataStore) ListDomainsByEcosystem(ecosystemID int) ([]*models.Domain, error) {
	return nil, nil
}
func (m *MockDataStore) ListAllDomains() ([]*models.Domain, error) { return nil, nil }

func (m *MockDataStore) CreateApp(app *models.App) error                             { return nil }
func (m *MockDataStore) GetAppByName(domainID int, name string) (*models.App, error) { return nil, nil }
func (m *MockDataStore) GetAppByNameGlobal(name string) (*models.App, error)         { return nil, nil }
func (m *MockDataStore) UpdateApp(app *models.App) error                             { return nil }
func (m *MockDataStore) DeleteApp(id int) error                                      { return nil }
func (m *MockDataStore) ListAppsByDomain(domainID int) ([]*models.App, error)        { return nil, nil }
func (m *MockDataStore) ListAllApps() ([]*models.App, error)                         { return nil, nil }

func (m *MockDataStore) CreateWorkspace(workspace *models.Workspace) error { return nil }
func (m *MockDataStore) GetWorkspaceByName(appID int, name string) (*models.Workspace, error) {
	return nil, nil
}
func (m *MockDataStore) GetWorkspaceByNameInApp(appID int, name string) (*models.Workspace, error) {
	return nil, nil
}
func (m *MockDataStore) UpdateWorkspace(workspace *models.Workspace) error          { return nil }
func (m *MockDataStore) DeleteWorkspace(id int) error                               { return nil }
func (m *MockDataStore) DeleteWorkspaceByID(id int) error                           { return nil }
func (m *MockDataStore) ListWorkspaces() ([]*models.Workspace, error)               { return nil, nil }
func (m *MockDataStore) ListWorkspacesByApp(appID int) ([]*models.Workspace, error) { return nil, nil }
func (m *MockDataStore) ListAllWorkspaces() ([]*models.Workspace, error)            { return nil, nil }
func (m *MockDataStore) FindWorkspaces(filter models.WorkspaceFilter) ([]*models.WorkspaceWithHierarchy, error) {
	return nil, nil
}

// Context operations
func (m *MockDataStore) GetContext() (*models.Context, error)      { return nil, nil }
func (m *MockDataStore) SetActiveEcosystem(ecosystemID *int) error { return nil }
func (m *MockDataStore) SetActiveDomain(domainID *int) error       { return nil }
func (m *MockDataStore) SetActiveApp(appID *int) error             { return nil }
func (m *MockDataStore) SetActiveWorkspace(workspaceID *int) error { return nil }

// Plugin operations
func (m *MockDataStore) CreatePlugin(plugin *models.NvimPluginDB) error            { return nil }
func (m *MockDataStore) GetPluginByName(name string) (*models.NvimPluginDB, error) { return nil, nil }
func (m *MockDataStore) GetPluginByID(id int) (*models.NvimPluginDB, error)        { return nil, nil }
func (m *MockDataStore) UpdatePlugin(plugin *models.NvimPluginDB) error            { return nil }
func (m *MockDataStore) UpsertPlugin(plugin *models.NvimPluginDB) error            { return nil }
func (m *MockDataStore) DeletePlugin(name string) error                            { return nil }
func (m *MockDataStore) ListPlugins() ([]*models.NvimPluginDB, error)              { return nil, nil }
func (m *MockDataStore) ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error) {
	return nil, nil
}
func (m *MockDataStore) ListPluginsByTags(tags []string) ([]*models.NvimPluginDB, error) {
	return nil, nil
}

// Workspace plugin associations
func (m *MockDataStore) AddPluginToWorkspace(workspaceID int, pluginID int) error      { return nil }
func (m *MockDataStore) RemovePluginFromWorkspace(workspaceID int, pluginID int) error { return nil }
func (m *MockDataStore) GetWorkspacePlugins(workspaceID int) ([]*models.NvimPluginDB, error) {
	return nil, nil
}
func (m *MockDataStore) SetWorkspacePluginEnabled(workspaceID int, pluginID int, enabled bool) error {
	return nil
}

// Theme operations
func (m *MockDataStore) CreateTheme(theme *models.NvimThemeDB) error             { return nil }
func (m *MockDataStore) GetThemeByName(name string) (*models.NvimThemeDB, error) { return nil, nil }
func (m *MockDataStore) GetThemeByID(id int) (*models.NvimThemeDB, error)        { return nil, nil }
func (m *MockDataStore) UpdateTheme(theme *models.NvimThemeDB) error             { return nil }
func (m *MockDataStore) DeleteTheme(name string) error                           { return nil }
func (m *MockDataStore) ListThemes() ([]*models.NvimThemeDB, error)              { return nil, nil }
func (m *MockDataStore) ListThemesByCategory(category string) ([]*models.NvimThemeDB, error) {
	return nil, nil
}
func (m *MockDataStore) SetActiveTheme(name string) error             { return nil }
func (m *MockDataStore) GetActiveTheme() (*models.NvimThemeDB, error) { return nil, nil }
func (m *MockDataStore) ClearActiveTheme() error                      { return nil }

// Terminal prompt operations
func (m *MockDataStore) CreateTerminalPrompt(prompt *models.TerminalPromptDB) error { return nil }
func (m *MockDataStore) GetTerminalPromptByName(name string) (*models.TerminalPromptDB, error) {
	return nil, nil
}
func (m *MockDataStore) UpdateTerminalPrompt(prompt *models.TerminalPromptDB) error { return nil }
func (m *MockDataStore) DeleteTerminalPrompt(name string) error                     { return nil }
func (m *MockDataStore) ListTerminalPrompts() ([]*models.TerminalPromptDB, error)   { return nil, nil }
func (m *MockDataStore) ListTerminalPromptsByType(promptType string) ([]*models.TerminalPromptDB, error) {
	return nil, nil
}
func (m *MockDataStore) ListTerminalPromptsByCategory(category string) ([]*models.TerminalPromptDB, error) {
	return nil, nil
}
func (m *MockDataStore) UpsertTerminalPrompt(prompt *models.TerminalPromptDB) error { return nil }

// Close method for DataStore interface
func (m *MockDataStore) Close() error { return nil }

// Credential operations
func (m *MockDataStore) CreateCredential(credential *models.CredentialDB) error { return nil }
func (m *MockDataStore) GetCredential(scopeType models.CredentialScopeType, scopeID int64, name string) (*models.CredentialDB, error) {
	return nil, nil
}
func (m *MockDataStore) UpdateCredential(credential *models.CredentialDB) error { return nil }
func (m *MockDataStore) DeleteCredential(scopeType models.CredentialScopeType, scopeID int64, name string) error {
	return nil
}
func (m *MockDataStore) ListCredentialsByScope(scopeType models.CredentialScopeType, scopeID int64) ([]*models.CredentialDB, error) {
	return nil, nil
}
func (m *MockDataStore) ListAllCredentials() ([]*models.CredentialDB, error) { return nil, nil }

// Driver access
func (m *MockDataStore) Driver() db.Driver { return nil }

// Connection methods
func (m *MockDataStore) Ping() error { return nil }

// Defaults methods
func (m *MockDataStore) GetDefault(key string) (string, error) {
	if val, ok := m.defaults[key]; ok {
		return val, nil
	}
	return "", nil
}
func (m *MockDataStore) SetDefault(key, value string) error {
	m.defaults[key] = value
	return nil
}
func (m *MockDataStore) DeleteDefault(key string) error {
	delete(m.defaults, key)
	return nil
}
func (m *MockDataStore) ListDefaults() (map[string]string, error) {
	return m.defaults, nil
}

// Registry operations (stub implementations - these tests don't use registries)
func (m *MockDataStore) CreateRegistry(registry *models.Registry) error { return nil }
func (m *MockDataStore) GetRegistryByName(name string) (*models.Registry, error) {
	return nil, nil
}
func (m *MockDataStore) GetRegistryByID(id int) (*models.Registry, error) { return nil, nil }
func (m *MockDataStore) GetRegistryByPort(port int) (*models.Registry, error) {
	return nil, nil
}
func (m *MockDataStore) UpdateRegistry(registry *models.Registry) error { return nil }
func (m *MockDataStore) DeleteRegistry(name string) error               { return nil }
func (m *MockDataStore) ListRegistries() ([]*models.Registry, error)    { return nil, nil }
func (m *MockDataStore) ListRegistriesByType(registryType string) ([]*models.Registry, error) {
	return nil, nil
}
func (m *MockDataStore) ListRegistriesByStatus(status string) ([]*models.Registry, error) {
	return nil, nil
}

// Package Operations (added to satisfy interface)
func (m *MockDataStore) CreatePackage(pkg *models.NvimPackageDB) error         { return nil }
func (m *MockDataStore) UpdatePackage(pkg *models.NvimPackageDB) error         { return nil }
func (m *MockDataStore) UpsertPackage(pkg *models.NvimPackageDB) error         { return nil }
func (m *MockDataStore) DeletePackage(name string) error                       { return nil }
func (m *MockDataStore) GetPackage(name string) (*models.NvimPackageDB, error) { return nil, nil }
func (m *MockDataStore) ListPackages() ([]*models.NvimPackageDB, error)        { return nil, nil }
func (m *MockDataStore) ListPackagesByLabel(key, value string) ([]*models.NvimPackageDB, error) {
	return nil, nil
}

// Terminal Package Operations (added to satisfy interface)
func (m *MockDataStore) CreateTerminalPackage(pkg *models.TerminalPackageDB) error { return nil }
func (m *MockDataStore) UpdateTerminalPackage(pkg *models.TerminalPackageDB) error { return nil }
func (m *MockDataStore) UpsertTerminalPackage(pkg *models.TerminalPackageDB) error { return nil }
func (m *MockDataStore) DeleteTerminalPackage(name string) error                   { return nil }
func (m *MockDataStore) GetTerminalPackage(name string) (*models.TerminalPackageDB, error) {
	return nil, nil
}
func (m *MockDataStore) ListTerminalPackages() ([]*models.TerminalPackageDB, error) { return nil, nil }
func (m *MockDataStore) ListTerminalPackagesByLabel(key, value string) ([]*models.TerminalPackageDB, error) {
	return nil, nil
}

// Terminal Plugin methods
func (m *MockDataStore) CreateTerminalPlugin(plugin *models.TerminalPluginDB) error { return nil }
func (m *MockDataStore) GetTerminalPlugin(name string) (*models.TerminalPluginDB, error) {
	return nil, nil
}
func (m *MockDataStore) UpdateTerminalPlugin(plugin *models.TerminalPluginDB) error { return nil }
func (m *MockDataStore) UpsertTerminalPlugin(plugin *models.TerminalPluginDB) error { return nil }
func (m *MockDataStore) DeleteTerminalPlugin(name string) error                     { return nil }
func (m *MockDataStore) ListTerminalPlugins() ([]*models.TerminalPluginDB, error)   { return nil, nil }
func (m *MockDataStore) ListTerminalPluginsByCategory(category string) ([]*models.TerminalPluginDB, error) {
	return nil, nil
}
func (m *MockDataStore) ListTerminalPluginsByShell(shell string) ([]*models.TerminalPluginDB, error) {
	return nil, nil
}
func (m *MockDataStore) ListTerminalPluginsByManager(manager string) ([]*models.TerminalPluginDB, error) {
	return nil, nil
}

// Terminal Emulator Operations
func (m *MockDataStore) CreateTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	return nil
}
func (m *MockDataStore) UpdateTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	return nil
}
func (m *MockDataStore) UpsertTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	return nil
}
func (m *MockDataStore) DeleteTerminalEmulator(name string) error {
	return nil
}
func (m *MockDataStore) GetTerminalEmulator(name string) (*models.TerminalEmulatorDB, error) {
	return nil, nil
}
func (m *MockDataStore) ListTerminalEmulators() ([]*models.TerminalEmulatorDB, error) {
	return nil, nil
}
func (m *MockDataStore) ListTerminalEmulatorsByType(emulatorType string) ([]*models.TerminalEmulatorDB, error) {
	return nil, nil
}
func (m *MockDataStore) ListTerminalEmulatorsByWorkspace(workspace string) ([]*models.TerminalEmulatorDB, error) {
	return nil, nil
}

// Terminal Profile Operations
func (m *MockDataStore) CreateTerminalProfile(profile *models.TerminalProfileDB) error { return nil }
func (m *MockDataStore) GetTerminalProfileByName(name string) (*models.TerminalProfileDB, error) {
	return nil, nil
}
func (m *MockDataStore) UpdateTerminalProfile(profile *models.TerminalProfileDB) error { return nil }
func (m *MockDataStore) UpsertTerminalProfile(profile *models.TerminalProfileDB) error { return nil }
func (m *MockDataStore) DeleteTerminalProfile(name string) error                       { return nil }
func (m *MockDataStore) ListTerminalProfiles() ([]*models.TerminalProfileDB, error)    { return nil, nil }
func (m *MockDataStore) ListTerminalProfilesByCategory(category string) ([]*models.TerminalProfileDB, error) {
	return nil, nil
}

// GitRepo Operations
func (m *MockDataStore) CreateGitRepo(repo *models.GitRepoDB) error { return nil }
func (m *MockDataStore) GetGitRepoByName(name string) (*models.GitRepoDB, error) {
	return nil, nil
}
func (m *MockDataStore) GetGitRepoByID(id int64) (*models.GitRepoDB, error) { return nil, nil }
func (m *MockDataStore) GetGitRepoBySlug(slug string) (*models.GitRepoDB, error) {
	return nil, nil
}
func (m *MockDataStore) UpdateGitRepo(repo *models.GitRepoDB) error { return nil }
func (m *MockDataStore) DeleteGitRepo(name string) error            { return nil }
func (m *MockDataStore) ListGitRepos() ([]models.GitRepoDB, error)  { return nil, nil }

// Registry History Operations (stub implementations - these tests don't use registry history)
func (m *MockDataStore) CreateRegistryHistory(history *models.RegistryHistory) error { return nil }
func (m *MockDataStore) GetRegistryHistory(registryID int, revision int) (*models.RegistryHistory, error) {
	return nil, nil
}
func (m *MockDataStore) GetLatestRegistryHistory(registryID int) (*models.RegistryHistory, error) {
	return nil, nil
}
func (m *MockDataStore) ListRegistryHistory(registryID int) ([]*models.RegistryHistory, error) {
	return nil, nil
}
func (m *MockDataStore) GetNextRevisionNumber(registryID int) (int, error) { return 0, nil }

// GenerateWorkspaceSlug generates a slug for a workspace (mock implementation)
func (m *MockDataStore) GenerateWorkspaceSlug(ecosystemName, domainName, appName, workspaceName string) string {
	return ecosystemName + "-" + domainName + "-" + appName + "-" + workspaceName
}

// GetWorkspacePath returns the path for a workspace (mock implementation)
func (m *MockDataStore) GetWorkspacePath(workspaceID int) (string, error) {
	if m.getWorkspaceError {
		return "", errors.New("mock workspace error")
	}
	if workspace, ok := m.workspaces[workspaceID]; ok {
		return "/mock/path/" + workspace.Name, nil
	}
	return "", sql.ErrNoRows
}

// GetWorkspaceSlug returns the slug for a workspace (mock implementation)
func (m *MockDataStore) GetWorkspaceSlug(workspaceID int) (string, error) {
	if m.getWorkspaceError {
		return "", errors.New("mock workspace error")
	}
	if workspace, ok := m.workspaces[workspaceID]; ok {
		return "mock-workspace-" + workspace.Name, nil
	}
	return "", sql.ErrNoRows
}

// GetWorkspaceBySlug retrieves a workspace by its hierarchical slug (mock implementation)
func (m *MockDataStore) GetWorkspaceBySlug(slug string) (*models.Workspace, error) {
	if m.getWorkspaceError {
		return nil, errors.New("mock workspace error")
	}
	// Simple mock: iterate through workspaces to find one with matching slug
	for _, workspace := range m.workspaces {
		if workspace.Slug == slug {
			return workspace, nil
		}
	}
	return nil, sql.ErrNoRows
}

// GetWorkspaceRepoPath returns the path to the workspace's git clone directory (mock implementation)
func (m *MockDataStore) GetWorkspaceRepoPath(workspaceID int) (string, error) {
	if m.getWorkspaceError {
		return "", errors.New("mock workspace error")
	}
	if workspace, ok := m.workspaces[workspaceID]; ok {
		return "/mock/path/" + workspace.Name + "/repo", nil
	}
	return "", sql.ErrNoRows
}

// CRD Operations (stub implementations for interface compliance)
func (m *MockDataStore) CreateCRD(crd *models.CustomResourceDefinition) error { return nil }
func (m *MockDataStore) GetCRDByKind(kind string) (*models.CustomResourceDefinition, error) {
	return nil, nil
}
func (m *MockDataStore) GetCRDByID(id int) (*models.CustomResourceDefinition, error) { return nil, nil }
func (m *MockDataStore) UpdateCRD(crd *models.CustomResourceDefinition) error        { return nil }
func (m *MockDataStore) DeleteCRD(kind string) error                                 { return nil }
func (m *MockDataStore) ListCRDs() ([]*models.CustomResourceDefinition, error)       { return nil, nil }
func (m *MockDataStore) ListCRDsByScope(scope string) ([]*models.CustomResourceDefinition, error) {
	return nil, nil
}

// Custom Resource Operations (stub implementations for interface compliance)
func (m *MockDataStore) CreateCustomResource(resource *models.CustomResource) error { return nil }
func (m *MockDataStore) GetCustomResource(kind, name, namespace string) (*models.CustomResource, error) {
	return nil, nil
}
func (m *MockDataStore) GetCustomResourceByID(id int) (*models.CustomResource, error) {
	return nil, nil
}
func (m *MockDataStore) UpdateCustomResource(resource *models.CustomResource) error { return nil }
func (m *MockDataStore) DeleteCustomResource(kind, name, namespace string) error    { return nil }
func (m *MockDataStore) ListCustomResources(kind string) ([]*models.CustomResource, error) {
	return nil, nil
}
func (m *MockDataStore) ListCustomResourcesByNamespace(kind, namespace string) ([]*models.CustomResource, error) {
	return nil, nil
}

// MockThemeStore implements theme.Store for testing
type MockThemeStore struct {
	themes   map[string]*theme.Theme
	getError bool
}

func NewMockThemeStore() *MockThemeStore {
	return &MockThemeStore{
		themes: make(map[string]*theme.Theme),
	}
}

func (m *MockThemeStore) AddTheme(name string, t *theme.Theme) {
	m.themes[name] = t
}

func (m *MockThemeStore) Get(name string) (*theme.Theme, error) {
	if m.getError {
		return nil, errors.New("mock theme store error")
	}
	if theme, ok := m.themes[name]; ok {
		return theme, nil
	}
	return nil, errors.New("theme not found")
}

func (m *MockThemeStore) List() ([]*theme.Theme, error)    { return nil, nil }
func (m *MockThemeStore) Save(*theme.Theme) error          { return nil }
func (m *MockThemeStore) Delete(string) error              { return nil }
func (m *MockThemeStore) GetActive() (*theme.Theme, error) { return nil, nil }
func (m *MockThemeStore) SetActive(name string) error      { return nil }
func (m *MockThemeStore) Path() string                     { return "" }

func TestNewHierarchyThemeResolver(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)

	assert.NotNil(t, resolver)
	assert.Equal(t, dataStore, resolver.dataStore)
	assert.Equal(t, themeStore, resolver.themeStore)
	assert.Equal(t, DefaultTheme, resolver.defaultTheme)
}

func TestHierarchyThemeResolver_ResolveDefault(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	// Add the default theme to the store
	defaultTheme := &theme.Theme{
		Name:     DefaultTheme,
		Category: "dark",
		Plugin:   theme.ThemePlugin{Repo: "test/default"},
		Colors:   map[string]string{"bg": "#1a1b26", "fg": "#c0caf5"},
	}
	themeStore.AddTheme(DefaultTheme, defaultTheme)

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)

	resolution, err := resolver.ResolveDefault()

	require.NoError(t, err)
	assert.NotNil(t, resolution)
	assert.Equal(t, defaultTheme, resolution.Theme)
	assert.Equal(t, LevelGlobal, resolution.Source)
	assert.Equal(t, "global default", resolution.SourceName)
	assert.Equal(t, 0, resolution.SourceID)
	assert.Len(t, resolution.Path, 1)

	step := resolution.Path[0]
	assert.Equal(t, LevelGlobal, step.Level)
	assert.Equal(t, "global default", step.Name)
	assert.Equal(t, DefaultTheme, step.ThemeName)
	assert.True(t, step.Found)
	assert.Empty(t, step.Error)
}

func TestHierarchyThemeResolver_ResolveDefault_ThemeLoadError(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()
	themeStore.getError = true // Force theme loading error

	// Use a custom resolver with a theme name that won't exist in library
	resolver := &HierarchyThemeResolver{
		dataStore:    dataStore,
		themeStore:   themeStore,
		defaultTheme: "nonexistent-theme-name",
	}

	resolution, err := resolver.ResolveDefault()

	// Should get error since theme store errors and library fallback will fail
	require.Error(t, err)
	assert.NotNil(t, resolution)
	assert.Nil(t, resolution.Theme)
	// Check that we get a reasonable error about the theme not being found
	assert.Contains(t, err.Error(), "not found in store or library")
	// The path error should also indicate the loading failure
	if len(resolution.Path) > 0 {
		assert.Contains(t, resolution.Path[0].Error, "failed to load default theme")
	}
}

func TestHierarchyThemeResolver_Resolve_HierarchyWalk(t *testing.T) {
	tests := []struct {
		name           string
		setupData      func(*MockDataStore)
		setupThemes    func(*MockThemeStore)
		level          HierarchyLevel
		objectID       int
		expectedTheme  string
		expectedSource HierarchyLevel
		expectedName   string
		expectError    bool
	}{
		{
			name: "theme found at workspace level (not currently implemented)",
			setupData: func(ds *MockDataStore) {
				ds.AddEcosystem(1, "test-ecosystem", stringPtr("ecosystem-theme"))
				ds.AddDomain(1, 1, "test-domain", stringPtr("domain-theme"))
				ds.AddApp(1, 1, "test-app", stringPtr("app-theme"))
				ds.AddWorkspace(1, 1, "test-workspace")
				// Note: workspace theme is not implemented in current code
			},
			setupThemes: func(ts *MockThemeStore) {
				ts.AddTheme("app-theme", &theme.Theme{Name: "app-theme"})
			},
			level:          LevelWorkspace,
			objectID:       1,
			expectedTheme:  "app-theme", // Falls through to app theme
			expectedSource: LevelApp,
			expectedName:   "test-app",
		},
		{
			name: "theme found at app level",
			setupData: func(ds *MockDataStore) {
				ds.AddEcosystem(1, "test-ecosystem", stringPtr("ecosystem-theme"))
				ds.AddDomain(1, 1, "test-domain", stringPtr("domain-theme"))
				ds.AddApp(1, 1, "test-app", stringPtr("app-theme"))
			},
			setupThemes: func(ts *MockThemeStore) {
				ts.AddTheme("app-theme", &theme.Theme{Name: "app-theme"})
			},
			level:          LevelApp,
			objectID:       1,
			expectedTheme:  "app-theme",
			expectedSource: LevelApp,
			expectedName:   "test-app",
		},
		{
			name: "theme found at domain level",
			setupData: func(ds *MockDataStore) {
				ds.AddEcosystem(1, "test-ecosystem", stringPtr("ecosystem-theme"))
				ds.AddDomain(1, 1, "test-domain", stringPtr("domain-theme"))
				ds.AddApp(1, 1, "test-app", nil) // No app theme
			},
			setupThemes: func(ts *MockThemeStore) {
				ts.AddTheme("domain-theme", &theme.Theme{Name: "domain-theme"})
			},
			level:          LevelApp,
			objectID:       1,
			expectedTheme:  "domain-theme",
			expectedSource: LevelDomain,
			expectedName:   "test-domain",
		},
		{
			name: "theme found at ecosystem level",
			setupData: func(ds *MockDataStore) {
				ds.AddEcosystem(1, "test-ecosystem", stringPtr("ecosystem-theme"))
				ds.AddDomain(1, 1, "test-domain", nil) // No domain theme
				ds.AddApp(1, 1, "test-app", nil)       // No app theme
			},
			setupThemes: func(ts *MockThemeStore) {
				ts.AddTheme("ecosystem-theme", &theme.Theme{Name: "ecosystem-theme"})
			},
			level:          LevelApp,
			objectID:       1,
			expectedTheme:  "ecosystem-theme",
			expectedSource: LevelEcosystem,
			expectedName:   "test-ecosystem",
		},
		{
			name: "no themes found - uses default",
			setupData: func(ds *MockDataStore) {
				ds.AddEcosystem(1, "test-ecosystem", nil)
				ds.AddDomain(1, 1, "test-domain", nil)
				ds.AddApp(1, 1, "test-app", nil)
			},
			setupThemes: func(ts *MockThemeStore) {
				ts.AddTheme(DefaultTheme, &theme.Theme{Name: DefaultTheme})
			},
			level:          LevelApp,
			objectID:       1,
			expectedTheme:  DefaultTheme,
			expectedSource: LevelGlobal,
			expectedName:   "global default",
		},
		{
			name: "theme loading fails - continues up hierarchy",
			setupData: func(ds *MockDataStore) {
				ds.AddEcosystem(1, "test-ecosystem", stringPtr("ecosystem-theme"))
				ds.AddDomain(1, 1, "test-domain", nil)
				ds.AddApp(1, 1, "test-app", stringPtr("nonexistent-theme"))
			},
			setupThemes: func(ts *MockThemeStore) {
				ts.AddTheme("ecosystem-theme", &theme.Theme{Name: "ecosystem-theme"})
				// Don't add "nonexistent-theme"
			},
			level:          LevelApp,
			objectID:       1,
			expectedTheme:  "ecosystem-theme",
			expectedSource: LevelEcosystem,
			expectedName:   "test-ecosystem",
		},
		{
			name: "object not found error",
			setupData: func(ds *MockDataStore) {
				// Don't add any data
			},
			setupThemes: func(ts *MockThemeStore) {
				ts.AddTheme(DefaultTheme, &theme.Theme{Name: DefaultTheme})
			},
			level:          LevelApp,
			objectID:       999, // Non-existent ID
			expectedTheme:  DefaultTheme,
			expectedSource: LevelGlobal,
			expectedName:   "global default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dataStore := NewMockDataStore()
			themeStore := NewMockThemeStore()

			tt.setupData(dataStore)
			tt.setupThemes(themeStore)

			resolver := NewHierarchyThemeResolver(dataStore, themeStore)

			ctx := context.Background()
			resolution, err := resolver.Resolve(ctx, tt.level, tt.objectID)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resolution)
			require.NotNil(t, resolution.Theme)

			assert.Equal(t, tt.expectedTheme, resolution.Theme.Name)
			assert.Equal(t, tt.expectedSource, resolution.Source)
			assert.Equal(t, tt.expectedName, resolution.SourceName)
			assert.NotEmpty(t, resolution.Path)
		})
	}
}

func TestHierarchyThemeResolver_GetResolutionPath(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	// Set up hierarchy with theme at domain level
	dataStore.AddEcosystem(1, "test-ecosystem", stringPtr("ecosystem-theme"))
	dataStore.AddDomain(1, 1, "test-domain", stringPtr("domain-theme"))
	dataStore.AddApp(1, 1, "test-app", nil)
	dataStore.AddWorkspace(1, 1, "test-workspace")

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)

	ctx := context.Background()
	resolution, err := resolver.GetResolutionPath(ctx, LevelWorkspace, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)

	// Should trace path but not load themes
	assert.Nil(t, resolution.Theme)
	assert.Equal(t, LevelDomain, resolution.Source)
	assert.Equal(t, "test-domain", resolution.SourceName)
	assert.Equal(t, 1, resolution.SourceID)

	// Should have complete path but stops when theme is found
	require.Len(t, resolution.Path, 3) // workspace, app, domain (stops at domain since it found theme)

	// Verify path steps
	assert.Equal(t, LevelWorkspace, resolution.Path[0].Level)
	assert.Equal(t, "test-workspace", resolution.Path[0].Name)
	assert.False(t, resolution.Path[0].Found)

	assert.Equal(t, LevelApp, resolution.Path[1].Level)
	assert.Equal(t, "test-app", resolution.Path[1].Name)
	assert.False(t, resolution.Path[1].Found)

	assert.Equal(t, LevelDomain, resolution.Path[2].Level)
	assert.Equal(t, "test-domain", resolution.Path[2].Name)
	assert.Equal(t, "domain-theme", resolution.Path[2].ThemeName)
	assert.True(t, resolution.Path[2].Found)
}

func TestHierarchyThemeResolver_ResolveAtLevel(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	// Set up test data
	dataStore.AddEcosystem(1, "test-ecosystem", stringPtr("ecosystem-theme"))
	dataStore.AddDomain(1, 1, "test-domain", stringPtr("domain-theme"))
	dataStore.AddApp(1, 1, "test-app", stringPtr("app-theme"))
	dataStore.AddWorkspace(1, 1, "test-workspace")
	dataStore.AddWorkspaceWithNvimStructure(2, 1, "test-workspace-with-nvim", "structure: lazyvim")

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)
	ctx := context.Background()

	tests := []struct {
		name          string
		level         HierarchyLevel
		objectID      int
		expectedName  string
		expectedTheme string
		expectedFound bool
		expectError   bool
	}{
		{
			name:          "ecosystem level",
			level:         LevelEcosystem,
			objectID:      1,
			expectedName:  "test-ecosystem",
			expectedTheme: "ecosystem-theme",
			expectedFound: true,
		},
		{
			name:          "domain level",
			level:         LevelDomain,
			objectID:      1,
			expectedName:  "test-domain",
			expectedTheme: "domain-theme",
			expectedFound: true,
		},
		{
			name:          "app level",
			level:         LevelApp,
			objectID:      1,
			expectedName:  "test-app",
			expectedTheme: "app-theme",
			expectedFound: true,
		},
		{
			name:          "workspace level with nvim structure (no theme support yet)",
			level:         LevelWorkspace,
			objectID:      2,
			expectedName:  "test-workspace-with-nvim",
			expectedTheme: "",
			expectedFound: false,
		},
		{
			name:          "global level",
			level:         LevelGlobal,
			objectID:      0,
			expectedName:  "global default",
			expectedTheme: DefaultTheme,
			expectedFound: true,
		},
		{
			name:        "invalid level",
			level:       HierarchyLevel(999),
			objectID:    1,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			step := resolver.resolveAtLevel(ctx, tt.level, tt.objectID)

			if tt.expectError {
				assert.NotEmpty(t, step.Error)
				return
			}

			assert.Equal(t, tt.level, step.Level)
			assert.Equal(t, tt.objectID, step.ObjectID)
			assert.Equal(t, tt.expectedName, step.Name)
			assert.Equal(t, tt.expectedTheme, step.ThemeName)
			assert.Equal(t, tt.expectedFound, step.Found)
			assert.Empty(t, step.Error)
		})
	}
}

func TestHierarchyThemeResolver_GetParent(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	// Set up hierarchy
	dataStore.AddEcosystem(1, "test-ecosystem", nil)
	dataStore.AddDomain(1, 1, "test-domain", nil)
	dataStore.AddApp(1, 1, "test-app", nil)
	dataStore.AddWorkspace(1, 1, "test-workspace")

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)
	ctx := context.Background()

	tests := []struct {
		name             string
		level            HierarchyLevel
		objectID         int
		expectedParentID int
		expectedLevel    HierarchyLevel
	}{
		{
			name:             "workspace to app",
			level:            LevelWorkspace,
			objectID:         1,
			expectedParentID: 1,
			expectedLevel:    LevelApp,
		},
		{
			name:             "app to domain",
			level:            LevelApp,
			objectID:         1,
			expectedParentID: 1,
			expectedLevel:    LevelDomain,
		},
		{
			name:             "domain to ecosystem",
			level:            LevelDomain,
			objectID:         1,
			expectedParentID: 1,
			expectedLevel:    LevelEcosystem,
		},
		{
			name:             "ecosystem to global",
			level:            LevelEcosystem,
			objectID:         1,
			expectedParentID: 0,
			expectedLevel:    LevelGlobal,
		},
		{
			name:             "global stays global",
			level:            LevelGlobal,
			objectID:         0,
			expectedParentID: 0,
			expectedLevel:    LevelGlobal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentID, parentLevel := resolver.getParent(ctx, tt.level, tt.objectID)

			assert.Equal(t, tt.expectedParentID, parentID)
			assert.Equal(t, tt.expectedLevel, parentLevel)
		})
	}
}

func TestHierarchyThemeResolver_GetParent_DatabaseErrors(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)
	ctx := context.Background()

	// Test error handling - should fall back to global when object not found
	tests := []struct {
		name          string
		level         HierarchyLevel
		objectID      int
		expectedLevel HierarchyLevel
	}{
		{
			name:          "workspace not found",
			level:         LevelWorkspace,
			objectID:      999,
			expectedLevel: LevelGlobal,
		},
		{
			name:          "app not found",
			level:         LevelApp,
			objectID:      999,
			expectedLevel: LevelGlobal,
		},
		{
			name:          "domain not found",
			level:         LevelDomain,
			objectID:      999,
			expectedLevel: LevelGlobal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parentID, parentLevel := resolver.getParent(ctx, tt.level, tt.objectID)

			assert.Equal(t, 0, parentID)
			assert.Equal(t, tt.expectedLevel, parentLevel)
		})
	}
}

func TestHierarchyThemeResolver_LoadTheme(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	// Add theme to store
	testTheme := &theme.Theme{Name: "test-theme"}
	themeStore.AddTheme("test-theme", testTheme)

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)

	t.Run("theme found in store", func(t *testing.T) {
		theme, err := resolver.loadTheme("test-theme")

		require.NoError(t, err)
		assert.Equal(t, testTheme, theme)
	})

	t.Run("theme not found", func(t *testing.T) {
		theme, err := resolver.loadTheme("nonexistent-theme")

		assert.Error(t, err)
		assert.Nil(t, theme)
		assert.Contains(t, err.Error(), "not found in store or library")
	})

	t.Run("theme loading fails - uses fallback library", func(t *testing.T) {
		themeStore.getError = true // Force theme store error
		_, err := resolver.loadTheme("test-theme")

		// The library fallback will still fail since we don't have library setup
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found in store or library")
	})
}

func TestHierarchyThemeResolver_EdgeCases(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	// Set up data with edge cases
	dataStore.AddEcosystem(1, "eco-with-empty-theme", stringPtr(""))
	dataStore.AddDomain(1, 1, "domain-with-empty-theme", stringPtr(""))
	dataStore.AddApp(1, 1, "app-with-empty-theme", stringPtr(""))

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)
	ctx := context.Background()

	t.Run("empty theme names should be skipped", func(t *testing.T) {
		// Add a default theme so we don't fail completely
		defaultTheme := &theme.Theme{Name: DefaultTheme}
		themeStore.AddTheme(DefaultTheme, defaultTheme)

		resolution, err := resolver.Resolve(ctx, LevelApp, 1)

		require.NoError(t, err)
		assert.NotNil(t, resolution)
		assert.Equal(t, DefaultTheme, resolution.Theme.Name)
		assert.Equal(t, LevelGlobal, resolution.Source)

		// Check that the path includes steps for empty themes but they're not found
		foundEmptyThemeSteps := 0
		for _, step := range resolution.Path {
			if step.ThemeName == "" && step.Level != LevelWorkspace {
				foundEmptyThemeSteps++
				assert.False(t, step.Found) // Should not be found
			}
		}
		assert.Greater(t, foundEmptyThemeSteps, 0)
	})

	t.Run("walkHierarchyTrace with no themes", func(t *testing.T) {
		// Test the trace path functionality with no themes set
		resolution, err := resolver.GetResolutionPath(ctx, LevelApp, 1)

		require.NoError(t, err)
		assert.NotNil(t, resolution)
		assert.Nil(t, resolution.Theme) // No theme loaded in trace mode
		assert.Equal(t, LevelGlobal, resolution.Source)
		assert.Equal(t, "global default", resolution.SourceName)

		// Should have the global default step added at the end
		require.NotEmpty(t, resolution.Path)
		lastStep := resolution.Path[len(resolution.Path)-1]
		assert.Equal(t, LevelGlobal, lastStep.Level)
		assert.Equal(t, "global default", lastStep.Name)
		assert.Equal(t, DefaultTheme, lastStep.ThemeName)
		assert.True(t, lastStep.Found)
	})
}

func TestHierarchyThemeResolver_LoadTheme_NilThemeStore(t *testing.T) {
	dataStore := NewMockDataStore()

	// Create resolver with nil theme store to test library fallback
	resolver := &HierarchyThemeResolver{
		dataStore:    dataStore,
		themeStore:   nil, // No theme store
		defaultTheme: DefaultTheme,
	}

	// This should try the library fallback immediately
	theme, err := resolver.loadTheme("nonexistent-theme")

	assert.Error(t, err)
	assert.Nil(t, theme)
	assert.Contains(t, err.Error(), "not found in store or library")
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// ==============================================================================
// BUG EXPOSURE TESTS (TDD Phase 2 - RED)
// These tests expose bugs identified in GitHub Issue #14
// ==============================================================================

// TestBug1_WorkspaceThemeFieldIsRead tests that workspace.Theme field is read
// Bug: resolveWorkspaceTheme() doesn't check workspace.Theme field (hierarchy.go:199-221)
func TestBug1_WorkspaceThemeFieldIsRead(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	// Set up hierarchy with theme at workspace level
	ecosystemThemeName := "ecosystem-default"
	workspaceThemeName := "workspace-override"

	dataStore.AddEcosystem(1, "test-ecosystem", stringPtr(ecosystemThemeName))
	dataStore.AddDomain(1, 1, "test-domain", nil)
	dataStore.AddApp(1, 1, "test-app", nil)

	// Add workspace WITH a theme set
	workspace := &models.Workspace{
		ID:    1,
		AppID: 1,
		Name:  "test-workspace",
		Theme: sql.NullString{String: workspaceThemeName, Valid: true},
	}
	dataStore.workspaces[1] = workspace

	// Add themes to store
	workspaceTheme := &theme.Theme{Name: workspaceThemeName}
	ecosystemTheme := &theme.Theme{Name: ecosystemThemeName}
	themeStore.AddTheme(workspaceThemeName, workspaceTheme)
	themeStore.AddTheme(ecosystemThemeName, ecosystemTheme)

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)
	ctx := context.Background()

	// EXPECTED: Should resolve to workspace theme
	// ACTUAL (BUG): Currently skips workspace.Theme and returns ecosystem theme
	resolution, err := resolver.Resolve(ctx, LevelWorkspace, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	require.NotNil(t, resolution.Theme)

	// THIS WILL FAIL until Bug #1 is fixed
	assert.Equal(t, workspaceThemeName, resolution.Theme.Name,
		"Bug #1: workspace.Theme field should be read and used")
	assert.Equal(t, LevelWorkspace, resolution.Source,
		"Bug #1: resolution should come from workspace level")
	assert.Equal(t, "test-workspace", resolution.SourceName,
		"Bug #1: source name should be workspace name")
}

// TestBug2_GlobalDefaultUsesDatabase tests that global default comes from database
// Bug: LevelGlobal uses hardcoded constant instead of database (hierarchy.go:188-192)
func TestBug2_GlobalDefaultUsesDatabase(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	// Set up database with custom global default
	customGlobalTheme := "custom-global-default"
	dataStore.defaults = map[string]string{
		"theme": customGlobalTheme,
	}

	// Set up hierarchy with NO themes at any level
	dataStore.AddEcosystem(1, "test-ecosystem", nil)
	dataStore.AddDomain(1, 1, "test-domain", nil)
	dataStore.AddApp(1, 1, "test-app", nil)

	// Add custom theme to store
	customTheme := &theme.Theme{Name: customGlobalTheme}
	themeStore.AddTheme(customGlobalTheme, customTheme)

	// Also add hardcoded default in case it falls back
	hardcodedDefault := &theme.Theme{Name: DefaultTheme}
	themeStore.AddTheme(DefaultTheme, hardcodedDefault)

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)
	ctx := context.Background()

	// EXPECTED: Should resolve to custom global default from database
	// ACTUAL (BUG): Currently returns hardcoded DefaultTheme constant
	resolution, err := resolver.Resolve(ctx, LevelApp, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	require.NotNil(t, resolution.Theme)

	// THIS WILL FAIL until Bug #2 is fixed
	assert.Equal(t, customGlobalTheme, resolution.Theme.Name,
		"Bug #2: global default should come from database, not hardcoded constant")
	assert.Equal(t, LevelGlobal, resolution.Source)
}

// TestBug3_FullHierarchyWalk tests complete hierarchy traversal
// Bug: getEffectiveTheme() in cmd/set_theme.go doesn't use hierarchy resolver
// This test verifies the EXPECTED behavior once Bug #3 is fixed
func TestBug3_FullHierarchyWalk(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	// Set up the exact scenario from GitHub Issue #14
	// Workspace dev: (none)
	// App fastapi-test: (none)
	// Domain python-apps: (none)
	// Ecosystem sandbox: coolnight-synthwave
	// Global: coolnight-ocean

	ecosystemThemeName := "coolnight-synthwave"
	globalThemeName := "coolnight-ocean"

	dataStore.AddEcosystem(1, "sandbox", stringPtr(ecosystemThemeName))
	dataStore.AddDomain(1, 1, "python-apps", nil) // No theme
	dataStore.AddApp(1, 1, "fastapi-test", nil)   // No theme
	dataStore.AddWorkspace(1, 1, "dev")           // No theme

	// Set global default
	dataStore.defaults = map[string]string{
		"theme": globalThemeName,
	}

	// Add themes to store
	ecosystemTheme := &theme.Theme{Name: ecosystemThemeName}
	globalTheme := &theme.Theme{Name: globalThemeName}
	themeStore.AddTheme(ecosystemThemeName, ecosystemTheme)
	themeStore.AddTheme(globalThemeName, globalTheme)

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)
	ctx := context.Background()

	// EXPECTED: Should walk workspace → app → domain → ecosystem and find theme at ecosystem
	// Should NOT reach global level
	resolution, err := resolver.Resolve(ctx, LevelWorkspace, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)
	require.NotNil(t, resolution.Theme)

	// THIS WILL FAIL until hierarchy traversal is correct
	assert.Equal(t, ecosystemThemeName, resolution.Theme.Name,
		"Bug #3: should resolve to ecosystem theme (coolnight-synthwave), not global")
	assert.Equal(t, LevelEcosystem, resolution.Source,
		"Bug #3: resolution should come from ecosystem level")
	assert.Equal(t, "sandbox", resolution.SourceName,
		"Bug #3: source name should be ecosystem name")

	// Verify the resolution path shows the walk
	require.NotEmpty(t, resolution.Path)

	// Should have steps for: workspace, app, domain, ecosystem
	// Should NOT have global step (stopped at ecosystem)
	assert.GreaterOrEqual(t, len(resolution.Path), 4,
		"Bug #3: path should include workspace, app, domain, ecosystem")

	// Verify each step in the path
	pathLevels := make([]HierarchyLevel, 0)
	for _, step := range resolution.Path {
		pathLevels = append(pathLevels, step.Level)
	}

	assert.Contains(t, pathLevels, LevelWorkspace, "Path should include workspace level")
	assert.Contains(t, pathLevels, LevelApp, "Path should include app level")
	assert.Contains(t, pathLevels, LevelDomain, "Path should include domain level")
	assert.Contains(t, pathLevels, LevelEcosystem, "Path should include ecosystem level")

	// Ecosystem step should be marked as found
	var ecosystemStep *ThemeStep
	for i := range resolution.Path {
		if resolution.Path[i].Level == LevelEcosystem {
			ecosystemStep = &resolution.Path[i]
			break
		}
	}
	require.NotNil(t, ecosystemStep, "Should have ecosystem step in path")
	assert.True(t, ecosystemStep.Found, "Ecosystem step should be marked as found")
	assert.Equal(t, ecosystemThemeName, ecosystemStep.ThemeName)
}

// TestBug3_GetResolutionPathTraversal tests GetResolutionPath with hierarchy
func TestBug3_GetResolutionPathTraversal(t *testing.T) {
	dataStore := NewMockDataStore()
	themeStore := NewMockThemeStore()

	// Set up hierarchy with theme at domain level
	domainThemeName := "domain-theme"

	dataStore.AddEcosystem(1, "test-ecosystem", nil)
	dataStore.AddDomain(1, 1, "test-domain", stringPtr(domainThemeName))
	dataStore.AddApp(1, 1, "test-app", nil)
	dataStore.AddWorkspace(1, 1, "test-workspace")

	resolver := NewHierarchyThemeResolver(dataStore, themeStore)
	ctx := context.Background()

	// GetResolutionPath should trace the hierarchy without loading themes
	resolution, err := resolver.GetResolutionPath(ctx, LevelWorkspace, 1)

	require.NoError(t, err)
	require.NotNil(t, resolution)

	// Should identify domain as the source
	assert.Equal(t, LevelDomain, resolution.Source)
	assert.Equal(t, "test-domain", resolution.SourceName)
	assert.Equal(t, 1, resolution.SourceID)

	// Should have walked: workspace → app → domain (stopped)
	require.Len(t, resolution.Path, 3)

	// Verify path order
	assert.Equal(t, LevelWorkspace, resolution.Path[0].Level)
	assert.Equal(t, "test-workspace", resolution.Path[0].Name)
	assert.False(t, resolution.Path[0].Found)

	assert.Equal(t, LevelApp, resolution.Path[1].Level)
	assert.Equal(t, "test-app", resolution.Path[1].Name)
	assert.False(t, resolution.Path[1].Found)

	assert.Equal(t, LevelDomain, resolution.Path[2].Level)
	assert.Equal(t, "test-domain", resolution.Path[2].Name)
	assert.True(t, resolution.Path[2].Found)
	assert.Equal(t, domainThemeName, resolution.Path[2].ThemeName)
}
