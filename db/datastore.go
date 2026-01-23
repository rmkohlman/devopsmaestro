package db

import (
	"devopsmaestro/models"
)

// DataStore is the interface for higher-level application operations that interact with the database.
type DataStore interface {
	CreateProject(project *models.Project) error
	GetProjectByName(name string) (*models.Project, error)
	UpdateProject(project *models.Project) error
	ListProjects() ([]*models.Project, error)

	// Plugin operations
	CreatePlugin(plugin *models.NvimPluginDB) error
	GetPluginByName(name string) (*models.NvimPluginDB, error)
	UpdatePlugin(plugin *models.NvimPluginDB) error
	ListPlugins() ([]*models.NvimPluginDB, error)
	ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error)
	DeletePlugin(name string) error

	// Workspace plugin associations
	AddPluginToWorkspace(workspaceID int, pluginID int) error
	RemovePluginFromWorkspace(workspaceID int, pluginID int) error
	GetWorkspacePlugins(workspaceID int) ([]*models.NvimPluginDB, error)
}
