package db

import (
	"devopsmaestro/models"
)

// DataStore is the high-level interface for application data operations.
// It provides a business-logic oriented API that abstracts away database specifics.
// Implementations can use any underlying Driver (SQLite, PostgreSQL, DuckDB, etc.).
type DataStore interface {
	// Project Operations

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

	// GetWorkspaceByName retrieves a workspace by project ID and name.
	GetWorkspaceByName(projectID int, name string) (*models.Workspace, error)

	// GetWorkspaceByID retrieves a workspace by its ID.
	GetWorkspaceByID(id int) (*models.Workspace, error)

	// UpdateWorkspace updates an existing workspace.
	UpdateWorkspace(workspace *models.Workspace) error

	// DeleteWorkspace removes a workspace by ID.
	DeleteWorkspace(id int) error

	// ListWorkspacesByProject retrieves all workspaces for a project.
	ListWorkspacesByProject(projectID int) ([]*models.Workspace, error)

	// ListAllWorkspaces retrieves all workspaces across all projects.
	ListAllWorkspaces() ([]*models.Workspace, error)

	// Context Operations (active project/workspace tracking)

	// GetContext retrieves the current context.
	GetContext() (*models.Context, error)

	// SetActiveProject sets the active project in the context.
	SetActiveProject(projectID *int) error

	// SetActiveWorkspace sets the active workspace in the context.
	SetActiveWorkspace(workspaceID *int) error

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
