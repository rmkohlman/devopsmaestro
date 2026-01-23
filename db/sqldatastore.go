package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

func init() {
	RegisterStore("sql", func(db Database) (DataStore, error) {
		return NewSQLDataStore(db, &SQLQueryBuilder{}), nil
	})
}

// SQLDataStore is a concrete implementation of the DataStore interface for SQL databases.
type SQLDataStore struct {
	db           Database
	queryBuilder *SQLQueryBuilder
}

// NewSQLDataStore creates a new instance of SQLDataStore
func NewSQLDataStore(db Database, queryBuilder *SQLQueryBuilder) *SQLDataStore {
	return &SQLDataStore{
		db:           db,
		queryBuilder: queryBuilder,
	}
}

// CreateProject inserts a new project into the database
func (ds *SQLDataStore) CreateProject(project *models.Project) error {
	query, values := ds.queryBuilder.BuildInsertQuery(project)
	_, err := ds.db.Execute(query, values...)
	if err != nil {
		return fmt.Errorf("failed to create project: %v", err)
	}
	return nil
}

// GetProjectByName retrieves a project by its name from the database
func (ds *SQLDataStore) GetProjectByName(name string) (*models.Project, error) {
	project := &models.Project{}
	query, values := ds.queryBuilder.BuildSelectQuery(project, "name = ?", name)

	row, err := ds.db.FetchOne(query, values...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project: %v", err)
	}

	if err := row.(*sql.Row).Scan(&project.ID, &project.Name, &project.Path, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan project: %v", err)
	}

	return project, nil
}

// UpdateProject updates an existing project in the database
func (ds *SQLDataStore) UpdateProject(project *models.Project) error {
	query, values := ds.queryBuilder.BuildUpdateQuery(project, "id = ?", project.ID)
	_, err := ds.db.Execute(query, values...)
	if err != nil {
		return fmt.Errorf("failed to update project: %v", err)
	}
	return nil
}

// ListProjects retrieves all projects from the database
func (ds *SQLDataStore) ListProjects() ([]*models.Project, error) {
	query := "SELECT id, name, path, description, created_at, updated_at FROM projects"
	rows, err := ds.db.FetchMany(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %v", err)
	}

	defer rows.(*sql.Rows).Close()

	var projects []*models.Project
	for rows.(*sql.Rows).Next() {
		project := &models.Project{}
		if err := rows.(*sql.Rows).Scan(&project.ID, &project.Name, &project.Path, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %v", err)
		}
		projects = append(projects, project)
	}

	if err := rows.(*sql.Rows).Err(); err != nil {
		return nil, fmt.Errorf("error iterating over project rows: %v", err)
	}

	return projects, nil
}

// CreateWorkspace inserts a new workspace into the database
func (ds *SQLDataStore) CreateWorkspace(workspace *models.Workspace) error {
	query := `INSERT INTO workspaces (project_id, name, description, image_name, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, datetime('now'), datetime('now'))`
	_, err := ds.db.Execute(query, workspace.ProjectID, workspace.Name, workspace.Description, workspace.ImageName, workspace.Status)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %v", err)
	}
	return nil
}

// GetWorkspaceByName retrieves a workspace by project ID and name
func (ds *SQLDataStore) GetWorkspaceByName(projectID int, name string) (*models.Workspace, error) {
	workspace := &models.Workspace{}
	query := `SELECT id, project_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at 
		FROM workspaces WHERE project_id = ? AND name = ?`

	row, err := ds.db.FetchOne(query, projectID, name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch workspace: %v", err)
	}

	if err := row.(*sql.Row).Scan(&workspace.ID, &workspace.ProjectID, &workspace.Name, &workspace.Description, &workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure, &workspace.NvimPlugins, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan workspace: %v", err)
	}

	return workspace, nil
}

// UpdateWorkspace updates an existing workspace in the database
func (ds *SQLDataStore) UpdateWorkspace(workspace *models.Workspace) error {
	query := `UPDATE workspaces SET name = ?, description = ?, updated_at = datetime('now') WHERE id = ?`
	_, err := ds.db.Execute(query, workspace.Name, workspace.Description, workspace.ID)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %v", err)
	}
	return nil
}

// GetContext retrieves the current context (always id=1)
func (ds *SQLDataStore) GetContext() (*models.Context, error) {
	context := &models.Context{}
	query := `SELECT id, active_project_id, active_workspace_id, updated_at FROM context WHERE id = 1`

	row, err := ds.db.FetchOne(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch context: %v", err)
	}

	if err := row.(*sql.Row).Scan(&context.ID, &context.ActiveProjectID, &context.ActiveWorkspaceID, &context.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan context: %v", err)
	}

	return context, nil
}

// SetActiveProject sets the active project in the context
func (ds *SQLDataStore) SetActiveProject(projectID *int) error {
	query := `UPDATE context SET active_project_id = ?, updated_at = datetime('now') WHERE id = 1`
	_, err := ds.db.Execute(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to set active project: %v", err)
	}
	return nil
}

// SetActiveWorkspace sets the active workspace in the context
func (ds *SQLDataStore) SetActiveWorkspace(workspaceID *int) error {
	query := `UPDATE context SET active_workspace_id = ?, updated_at = datetime('now') WHERE id = 1`
	_, err := ds.db.Execute(query, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to set active workspace: %v", err)
	}
	return nil
}

// CreatePlugin inserts a new plugin into the database
func (ds *SQLDataStore) CreatePlugin(plugin *models.NvimPluginDB) error {
	query := `INSERT INTO nvim_plugins (name, description, repo, branch, version, priority, lazy, 
		event, ft, keys, cmd, dependencies, build, config, init, opts, keymaps, category, tags, enabled, 
		created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'), datetime('now'))`

	result, err := ds.db.Execute(query,
		plugin.Name, plugin.Description, plugin.Repo, plugin.Branch, plugin.Version, plugin.Priority,
		plugin.Lazy, plugin.Event, plugin.Ft, plugin.Keys, plugin.Cmd, plugin.Dependencies, plugin.Build,
		plugin.Config, plugin.Init, plugin.Opts, plugin.Keymaps, plugin.Category, plugin.Tags, plugin.Enabled)

	if err != nil {
		return fmt.Errorf("failed to create plugin: %v", err)
	}

	// Get the inserted ID
	if result != nil {
		if sqlResult, ok := result.(sql.Result); ok {
			id, err := sqlResult.LastInsertId()
			if err == nil {
				plugin.ID = int(id)
			}
		}
	}

	return nil
}

// GetPluginByName retrieves a plugin by its name from the database
func (ds *SQLDataStore) GetPluginByName(name string) (*models.NvimPluginDB, error) {
	plugin := &models.NvimPluginDB{}
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd, 
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE name = ?`

	row, err := ds.db.FetchOne(query, name)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch plugin: %v", err)
	}

	if err := row.(*sql.Row).Scan(
		&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
		&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
		&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
		&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plugin not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan plugin: %v", err)
	}

	return plugin, nil
}

// UpdatePlugin updates an existing plugin in the database
func (ds *SQLDataStore) UpdatePlugin(plugin *models.NvimPluginDB) error {
	query := `UPDATE nvim_plugins SET description = ?, repo = ?, branch = ?, version = ?, priority = ?, 
		lazy = ?, event = ?, ft = ?, keys = ?, cmd = ?, dependencies = ?, build = ?, config = ?, init = ?,
		opts = ?, keymaps = ?, category = ?, tags = ?, enabled = ?, updated_at = datetime('now') 
		WHERE name = ?`

	_, err := ds.db.Execute(query,
		plugin.Description, plugin.Repo, plugin.Branch, plugin.Version, plugin.Priority,
		plugin.Lazy, plugin.Event, plugin.Ft, plugin.Keys, plugin.Cmd, plugin.Dependencies, plugin.Build,
		plugin.Config, plugin.Init, plugin.Opts, plugin.Keymaps, plugin.Category, plugin.Tags, plugin.Enabled,
		plugin.Name)

	if err != nil {
		return fmt.Errorf("failed to update plugin: %v", err)
	}
	return nil
}

// ListPlugins retrieves all plugins from the database
func (ds *SQLDataStore) ListPlugins() ([]*models.NvimPluginDB, error) {
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd,
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins ORDER BY name`

	rows, err := ds.db.FetchMany(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %v", err)
	}

	defer rows.(*sql.Rows).Close()

	var plugins []*models.NvimPluginDB
	for rows.(*sql.Rows).Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.(*sql.Rows).Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %v", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.(*sql.Rows).Err(); err != nil {
		return nil, fmt.Errorf("error iterating over plugin rows: %v", err)
	}

	return plugins, nil
}

// ListPluginsByCategory retrieves plugins filtered by category
func (ds *SQLDataStore) ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error) {
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd,
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE category = ? ORDER BY name`

	rows, err := ds.db.FetchMany(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by category: %v", err)
	}

	defer rows.(*sql.Rows).Close()

	var plugins []*models.NvimPluginDB
	for rows.(*sql.Rows).Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.(*sql.Rows).Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %v", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.(*sql.Rows).Err(); err != nil {
		return nil, fmt.Errorf("error iterating over plugin rows: %v", err)
	}

	return plugins, nil
}

// DeletePlugin removes a plugin from the database
func (ds *SQLDataStore) DeletePlugin(name string) error {
	query := `DELETE FROM nvim_plugins WHERE name = ?`
	_, err := ds.db.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete plugin: %v", err)
	}
	return nil
}

// AddPluginToWorkspace associates a plugin with a workspace
func (ds *SQLDataStore) AddPluginToWorkspace(workspaceID int, pluginID int) error {
	query := `INSERT OR IGNORE INTO workspace_plugins (workspace_id, plugin_id, enabled, created_at)
		VALUES (?, ?, TRUE, datetime('now'))`
	_, err := ds.db.Execute(query, workspaceID, pluginID)
	if err != nil {
		return fmt.Errorf("failed to add plugin to workspace: %v", err)
	}
	return nil
}

// RemovePluginFromWorkspace removes a plugin association from a workspace
func (ds *SQLDataStore) RemovePluginFromWorkspace(workspaceID int, pluginID int) error {
	query := `DELETE FROM workspace_plugins WHERE workspace_id = ? AND plugin_id = ?`
	_, err := ds.db.Execute(query, workspaceID, pluginID)
	if err != nil {
		return fmt.Errorf("failed to remove plugin from workspace: %v", err)
	}
	return nil
}

// GetWorkspacePlugins retrieves all plugins associated with a workspace
func (ds *SQLDataStore) GetWorkspacePlugins(workspaceID int) ([]*models.NvimPluginDB, error) {
	query := `SELECT p.id, p.name, p.description, p.repo, p.branch, p.version, p.priority, p.lazy, 
		p.event, p.ft, p.keys, p.cmd, p.dependencies, p.build, p.config, p.init, p.opts, p.keymaps,
		p.category, p.tags, p.enabled, p.created_at, p.updated_at
		FROM nvim_plugins p
		JOIN workspace_plugins wp ON p.id = wp.plugin_id
		WHERE wp.workspace_id = ? AND wp.enabled = TRUE
		ORDER BY p.priority DESC, p.name`

	rows, err := ds.db.FetchMany(query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace plugins: %v", err)
	}

	defer rows.(*sql.Rows).Close()

	var plugins []*models.NvimPluginDB
	for rows.(*sql.Rows).Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.(*sql.Rows).Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %v", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.(*sql.Rows).Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspace plugin rows: %v", err)
	}

	return plugins, nil
}
