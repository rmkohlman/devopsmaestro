package db

import (
	"devopsmaestro/models"
)

// DataStore is the high-level interface for application data operations.
// It provides a business-logic oriented API that abstracts away database specifics.
// Implementations can use any underlying Driver (SQLite, PostgreSQL, DuckDB, etc.).
type DataStore interface {
	// Ecosystem Operations (top-level grouping)

	// CreateEcosystem inserts a new ecosystem into the database.
	CreateEcosystem(ecosystem *models.Ecosystem) error

	// GetEcosystemByName retrieves an ecosystem by its name.
	GetEcosystemByName(name string) (*models.Ecosystem, error)

	// GetEcosystemByID retrieves an ecosystem by its ID.
	GetEcosystemByID(id int) (*models.Ecosystem, error)

	// UpdateEcosystem updates an existing ecosystem.
	UpdateEcosystem(ecosystem *models.Ecosystem) error

	// DeleteEcosystem removes an ecosystem by name.
	DeleteEcosystem(name string) error

	// ListEcosystems retrieves all ecosystems.
	ListEcosystems() ([]*models.Ecosystem, error)

	// Domain Operations (bounded context within an ecosystem)

	// CreateDomain inserts a new domain into the database.
	CreateDomain(domain *models.Domain) error

	// GetDomainByName retrieves a domain by ecosystem ID and name.
	GetDomainByName(ecosystemID int, name string) (*models.Domain, error)

	// GetDomainByID retrieves a domain by its ID.
	GetDomainByID(id int) (*models.Domain, error)

	// UpdateDomain updates an existing domain.
	UpdateDomain(domain *models.Domain) error

	// DeleteDomain removes a domain by ID.
	DeleteDomain(id int) error

	// ListDomainsByEcosystem retrieves all domains for an ecosystem.
	ListDomainsByEcosystem(ecosystemID int) ([]*models.Domain, error)

	// ListAllDomains retrieves all domains across all ecosystems.
	ListAllDomains() ([]*models.Domain, error)

	// App Operations (codebase/application within a domain)

	// CreateApp inserts a new app into the database.
	CreateApp(app *models.App) error

	// GetAppByName retrieves an app by domain ID and name.
	GetAppByName(domainID int, name string) (*models.App, error)

	// GetAppByNameGlobal retrieves an app by name across all domains.
	// Returns the first match if multiple apps have the same name in different domains.
	// This is useful for CLI convenience when the user doesn't want to specify domain context.
	GetAppByNameGlobal(name string) (*models.App, error)

	// GetAppByID retrieves an app by its ID.
	GetAppByID(id int) (*models.App, error)

	// UpdateApp updates an existing app.
	UpdateApp(app *models.App) error

	// DeleteApp removes an app by ID.
	DeleteApp(id int) error

	// ListAppsByDomain retrieves all apps for a domain.
	ListAppsByDomain(domainID int) ([]*models.App, error)

	// ListAllApps retrieves all apps across all domains.
	ListAllApps() ([]*models.App, error)

	// Project Operations (DEPRECATED: migrate to Domain/App)

	// CreateProject inserts a new project into the database.
	CreateProject(project *models.Project) error

	// GetProjectByName retrieves a project by its name.
	GetProjectByName(name string) (*models.Project, error)

	// GetProjectByID retrieves a project by its ID.
	GetProjectByID(id int) (*models.Project, error)

	// UpdateProject updates an existing project.
	UpdateProject(project *models.Project) error

	// DeleteProject removes a project by name.
	DeleteProject(name string) error

	// ListProjects retrieves all projects.
	ListProjects() ([]*models.Project, error)

	// Workspace Operations

	// CreateWorkspace inserts a new workspace.
	CreateWorkspace(workspace *models.Workspace) error

	// GetWorkspaceByName retrieves a workspace by app ID and name.
	GetWorkspaceByName(appID int, name string) (*models.Workspace, error)

	// GetWorkspaceByID retrieves a workspace by its ID.
	GetWorkspaceByID(id int) (*models.Workspace, error)

	// UpdateWorkspace updates an existing workspace.
	UpdateWorkspace(workspace *models.Workspace) error

	// DeleteWorkspace removes a workspace by ID.
	DeleteWorkspace(id int) error

	// ListWorkspacesByApp retrieves all workspaces for an app.
	ListWorkspacesByApp(appID int) ([]*models.Workspace, error)

	// ListAllWorkspaces retrieves all workspaces across all apps.
	ListAllWorkspaces() ([]*models.Workspace, error)

	// FindWorkspaces searches for workspaces matching the given filter criteria.
	// Returns workspaces with their full hierarchy information (ecosystem, domain, app).
	// Use this for smart workspace resolution when the user provides partial criteria.
	FindWorkspaces(filter models.WorkspaceFilter) ([]*models.WorkspaceWithHierarchy, error)

	// Context Operations (active selection state tracking)
	// The hierarchy is: Ecosystem -> Domain -> App -> Workspace

	// GetContext retrieves the current context.
	GetContext() (*models.Context, error)

	// SetActiveEcosystem sets the active ecosystem in the context.
	SetActiveEcosystem(ecosystemID *int) error

	// SetActiveDomain sets the active domain in the context.
	SetActiveDomain(domainID *int) error

	// SetActiveApp sets the active app in the context.
	SetActiveApp(appID *int) error

	// SetActiveWorkspace sets the active workspace in the context.
	SetActiveWorkspace(workspaceID *int) error

	// SetActiveProject sets the active project in the context.
	// DEPRECATED: Use SetActiveApp instead. Will be removed in v0.9.0.
	SetActiveProject(projectID *int) error

	// Plugin Operations

	// CreatePlugin inserts a new nvim plugin.
	CreatePlugin(plugin *models.NvimPluginDB) error

	// GetPluginByName retrieves a plugin by its name.
	GetPluginByName(name string) (*models.NvimPluginDB, error)

	// GetPluginByID retrieves a plugin by its ID.
	GetPluginByID(id int) (*models.NvimPluginDB, error)

	// UpdatePlugin updates an existing plugin.
	UpdatePlugin(plugin *models.NvimPluginDB) error

	// DeletePlugin removes a plugin by name.
	DeletePlugin(name string) error

	// ListPlugins retrieves all plugins.
	ListPlugins() ([]*models.NvimPluginDB, error)

	// ListPluginsByCategory retrieves plugins filtered by category.
	ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error)

	// ListPluginsByTags retrieves plugins that have any of the specified tags.
	ListPluginsByTags(tags []string) ([]*models.NvimPluginDB, error)

	// Workspace Plugin Associations

	// AddPluginToWorkspace associates a plugin with a workspace.
	AddPluginToWorkspace(workspaceID int, pluginID int) error

	// RemovePluginFromWorkspace removes a plugin association from a workspace.
	RemovePluginFromWorkspace(workspaceID int, pluginID int) error

	// GetWorkspacePlugins retrieves all plugins associated with a workspace.
	GetWorkspacePlugins(workspaceID int) ([]*models.NvimPluginDB, error)

	// SetWorkspacePluginEnabled enables or disables a plugin for a workspace.
	SetWorkspacePluginEnabled(workspaceID int, pluginID int, enabled bool) error

	// Theme Operations

	// CreateTheme inserts a new nvim theme.
	CreateTheme(theme *models.NvimThemeDB) error

	// GetThemeByName retrieves a theme by its name.
	GetThemeByName(name string) (*models.NvimThemeDB, error)

	// GetThemeByID retrieves a theme by its ID.
	GetThemeByID(id int) (*models.NvimThemeDB, error)

	// UpdateTheme updates an existing theme.
	UpdateTheme(theme *models.NvimThemeDB) error

	// DeleteTheme removes a theme by name.
	DeleteTheme(name string) error

	// ListThemes retrieves all themes.
	ListThemes() ([]*models.NvimThemeDB, error)

	// ListThemesByCategory retrieves themes filtered by category.
	ListThemesByCategory(category string) ([]*models.NvimThemeDB, error)

	// GetActiveTheme retrieves the currently active theme.
	GetActiveTheme() (*models.NvimThemeDB, error)

	// SetActiveTheme sets the active theme by name (deactivates others).
	SetActiveTheme(name string) error

	// ClearActiveTheme deactivates all themes.
	ClearActiveTheme() error

	// Terminal Prompt Operations

	// CreateTerminalPrompt inserts a new terminal prompt.
	CreateTerminalPrompt(prompt *models.TerminalPromptDB) error

	// GetTerminalPromptByName retrieves a terminal prompt by its name.
	GetTerminalPromptByName(name string) (*models.TerminalPromptDB, error)

	// UpdateTerminalPrompt updates an existing terminal prompt.
	UpdateTerminalPrompt(prompt *models.TerminalPromptDB) error

	// UpsertTerminalPrompt creates or updates a terminal prompt (by name).
	UpsertTerminalPrompt(prompt *models.TerminalPromptDB) error

	// DeleteTerminalPrompt removes a terminal prompt by name.
	DeleteTerminalPrompt(name string) error

	// ListTerminalPrompts retrieves all terminal prompts.
	ListTerminalPrompts() ([]*models.TerminalPromptDB, error)

	// ListTerminalPromptsByType retrieves terminal prompts filtered by type.
	ListTerminalPromptsByType(promptType string) ([]*models.TerminalPromptDB, error)

	// ListTerminalPromptsByCategory retrieves terminal prompts filtered by category.
	ListTerminalPromptsByCategory(category string) ([]*models.TerminalPromptDB, error)

	// Terminal Profile Operations

	// CreateTerminalProfile inserts a new terminal profile.
	CreateTerminalProfile(profile *models.TerminalProfileDB) error

	// GetTerminalProfileByName retrieves a terminal profile by its name.
	GetTerminalProfileByName(name string) (*models.TerminalProfileDB, error)

	// UpdateTerminalProfile updates an existing terminal profile.
	UpdateTerminalProfile(profile *models.TerminalProfileDB) error

	// UpsertTerminalProfile creates or updates a terminal profile (by name).
	UpsertTerminalProfile(profile *models.TerminalProfileDB) error

	// DeleteTerminalProfile removes a terminal profile by name.
	DeleteTerminalProfile(name string) error

	// ListTerminalProfiles retrieves all terminal profiles.
	ListTerminalProfiles() ([]*models.TerminalProfileDB, error)

	// ListTerminalProfilesByCategory retrieves terminal profiles filtered by category.
	ListTerminalProfilesByCategory(category string) ([]*models.TerminalProfileDB, error)

	// Credential Operations

	// CreateCredential inserts a new credential configuration.
	CreateCredential(credential *models.CredentialDB) error

	// GetCredential retrieves a credential by scope and name.
	GetCredential(scopeType models.CredentialScopeType, scopeID int64, name string) (*models.CredentialDB, error)

	// UpdateCredential updates an existing credential.
	UpdateCredential(credential *models.CredentialDB) error

	// DeleteCredential removes a credential by scope and name.
	DeleteCredential(scopeType models.CredentialScopeType, scopeID int64, name string) error

	// ListCredentialsByScope retrieves all credentials for a specific scope.
	ListCredentialsByScope(scopeType models.CredentialScopeType, scopeID int64) ([]*models.CredentialDB, error)

	// ListAllCredentials retrieves all credentials across all scopes.
	ListAllCredentials() ([]*models.CredentialDB, error)

	// Default Operations

	// GetDefault retrieves a default value by key.
	// Returns empty string if key is not found (not an error).
	GetDefault(key string) (string, error)

	// SetDefault sets a default value for the given key.
	// Uses upsert behavior (INSERT OR REPLACE).
	SetDefault(key, value string) error

	// DeleteDefault removes a default value by key.
	// No error if key doesn't exist.
	DeleteDefault(key string) error

	// ListDefaults retrieves all default values as a key-value map.
	ListDefaults() (map[string]string, error)

	// Package Operations

	// CreatePackage inserts a new nvim package.
	CreatePackage(pkg *models.NvimPackageDB) error

	// UpdatePackage updates an existing nvim package.
	UpdatePackage(pkg *models.NvimPackageDB) error

	// UpsertPackage creates or updates an nvim package (by name).
	UpsertPackage(pkg *models.NvimPackageDB) error

	// DeletePackage removes a package by name.
	DeletePackage(name string) error

	// GetPackage retrieves a package by its name.
	GetPackage(name string) (*models.NvimPackageDB, error)

	// ListPackages retrieves all packages.
	ListPackages() ([]*models.NvimPackageDB, error)

	// ListPackagesByLabel retrieves packages that have a specific label key-value pair.
	ListPackagesByLabel(key, value string) ([]*models.NvimPackageDB, error)

	// Terminal Package Operations

	// CreateTerminalPackage inserts a new terminal package.
	CreateTerminalPackage(pkg *models.TerminalPackageDB) error

	// UpdateTerminalPackage updates an existing terminal package.
	UpdateTerminalPackage(pkg *models.TerminalPackageDB) error

	// UpsertTerminalPackage creates or updates a terminal package (by name).
	UpsertTerminalPackage(pkg *models.TerminalPackageDB) error

	// DeleteTerminalPackage removes a terminal package by name.
	DeleteTerminalPackage(name string) error

	// GetTerminalPackage retrieves a terminal package by its name.
	GetTerminalPackage(name string) (*models.TerminalPackageDB, error)

	// ListTerminalPackages retrieves all terminal packages.
	ListTerminalPackages() ([]*models.TerminalPackageDB, error)

	// ListTerminalPackagesByLabel retrieves terminal packages that have a specific label key-value pair.
	ListTerminalPackagesByLabel(key, value string) ([]*models.TerminalPackageDB, error)

	// Terminal Plugin Operations

	// CreateTerminalPlugin inserts a new terminal plugin.
	CreateTerminalPlugin(plugin *models.TerminalPluginDB) error

	// UpdateTerminalPlugin updates an existing terminal plugin.
	UpdateTerminalPlugin(plugin *models.TerminalPluginDB) error

	// UpsertTerminalPlugin creates or updates a terminal plugin (by name).
	UpsertTerminalPlugin(plugin *models.TerminalPluginDB) error

	// DeleteTerminalPlugin removes a terminal plugin by name.
	DeleteTerminalPlugin(name string) error

	// GetTerminalPlugin retrieves a terminal plugin by its name.
	GetTerminalPlugin(name string) (*models.TerminalPluginDB, error)

	// ListTerminalPlugins retrieves all terminal plugins.
	ListTerminalPlugins() ([]*models.TerminalPluginDB, error)

	// ListTerminalPluginsByCategory retrieves terminal plugins filtered by category.
	ListTerminalPluginsByCategory(category string) ([]*models.TerminalPluginDB, error)

	// ListTerminalPluginsByShell retrieves terminal plugins filtered by shell.
	ListTerminalPluginsByShell(shell string) ([]*models.TerminalPluginDB, error)

	// ListTerminalPluginsByManager retrieves terminal plugins filtered by manager.
	ListTerminalPluginsByManager(manager string) ([]*models.TerminalPluginDB, error)

	// Terminal Emulator Operations

	// CreateTerminalEmulator inserts a new terminal emulator config.
	CreateTerminalEmulator(emulator *models.TerminalEmulatorDB) error

	// UpdateTerminalEmulator updates an existing terminal emulator config.
	UpdateTerminalEmulator(emulator *models.TerminalEmulatorDB) error

	// UpsertTerminalEmulator creates or updates a terminal emulator (by name).
	UpsertTerminalEmulator(emulator *models.TerminalEmulatorDB) error

	// DeleteTerminalEmulator removes a terminal emulator by name.
	DeleteTerminalEmulator(name string) error

	// GetTerminalEmulator retrieves a terminal emulator by its name.
	GetTerminalEmulator(name string) (*models.TerminalEmulatorDB, error)

	// ListTerminalEmulators retrieves all terminal emulators.
	ListTerminalEmulators() ([]*models.TerminalEmulatorDB, error)

	// ListTerminalEmulatorsByType retrieves terminal emulators filtered by type.
	ListTerminalEmulatorsByType(emulatorType string) ([]*models.TerminalEmulatorDB, error)

	// ListTerminalEmulatorsByWorkspace retrieves terminal emulators for a workspace.
	ListTerminalEmulatorsByWorkspace(workspace string) ([]*models.TerminalEmulatorDB, error)

	// Driver Access

	// Driver returns the underlying database driver.
	// Useful for advanced operations or transactions.
	Driver() Driver

	// Health and Maintenance

	// Close releases any resources held by the DataStore.
	Close() error

	// Ping verifies the database connection is alive.
	Ping() error
}

// DataStoreConfig provides configuration for creating a DataStore.
type DataStoreConfig struct {
	// Driver is the database driver to use.
	Driver Driver

	// QueryBuilder is the SQL dialect-specific query builder.
	// If nil, a default builder will be selected based on the driver type.
	QueryBuilder QueryBuilder
}
