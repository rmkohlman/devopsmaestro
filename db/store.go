package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// SQLDataStore is a concrete implementation of the DataStore interface.
// It uses the Driver interface for database operations and QueryBuilder
// for dialect-specific SQL generation.
type SQLDataStore struct {
	driver       Driver
	queryBuilder QueryBuilder
}

// NewSQLDataStore creates a new SQLDataStore with the given driver.
// If queryBuilder is nil, the appropriate builder is selected based on driver type.
func NewSQLDataStore(driver Driver, queryBuilder QueryBuilder) *SQLDataStore {
	if queryBuilder == nil {
		queryBuilder = QueryBuilderFor(driver.Type())
	}
	return &SQLDataStore{
		driver:       driver,
		queryBuilder: queryBuilder,
	}
}

// NewDataStore creates a DataStore from configuration.
// This is the recommended way to create a DataStore.
func NewDataStore(cfg DataStoreConfig) (DataStore, error) {
	if cfg.Driver == nil {
		return nil, fmt.Errorf("driver is required")
	}
	return NewSQLDataStore(cfg.Driver, cfg.QueryBuilder), nil
}

// Driver returns the underlying database driver.
func (ds *SQLDataStore) Driver() Driver {
	return ds.driver
}

// Close releases any resources held by the DataStore.
func (ds *SQLDataStore) Close() error {
	return ds.driver.Close()
}

// Ping verifies the database connection is alive.
func (ds *SQLDataStore) Ping() error {
	return ds.driver.Ping()
}

// =============================================================================
// Project Operations
// =============================================================================

// CreateProject inserts a new project into the database.
func (ds *SQLDataStore) CreateProject(project *models.Project) error {
	query := fmt.Sprintf(`INSERT INTO projects (name, path, description, created_at, updated_at) 
		VALUES (?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, project.Name, project.Path, project.Description)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		project.ID = int(id)
	}

	return nil
}

// GetProjectByName retrieves a project by its name.
func (ds *SQLDataStore) GetProjectByName(name string) (*models.Project, error) {
	project := &models.Project{}
	query := `SELECT id, name, path, description, created_at, updated_at FROM projects WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&project.ID, &project.Name, &project.Path, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	return project, nil
}

// GetProjectByID retrieves a project by its ID.
func (ds *SQLDataStore) GetProjectByID(id int) (*models.Project, error) {
	project := &models.Project{}
	query := `SELECT id, name, path, description, created_at, updated_at FROM projects WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&project.ID, &project.Name, &project.Path, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	return project, nil
}

// UpdateProject updates an existing project.
func (ds *SQLDataStore) UpdateProject(project *models.Project) error {
	query := fmt.Sprintf(`UPDATE projects SET name = ?, path = ?, description = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, project.Name, project.Path, project.Description, project.ID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

// DeleteProject removes a project by name.
func (ds *SQLDataStore) DeleteProject(name string) error {
	query := `DELETE FROM projects WHERE name = ?`
	_, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// ListProjects retrieves all projects.
func (ds *SQLDataStore) ListProjects() ([]*models.Project, error) {
	query := `SELECT id, name, path, description, created_at, updated_at FROM projects ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		if err := rows.Scan(&project.ID, &project.Name, &project.Path, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over projects: %w", err)
	}

	return projects, nil
}

// =============================================================================
// Workspace Operations
// =============================================================================

// CreateWorkspace inserts a new workspace.
func (ds *SQLDataStore) CreateWorkspace(workspace *models.Workspace) error {
	query := fmt.Sprintf(`INSERT INTO workspaces (project_id, name, description, image_name, status, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, workspace.ProjectID, workspace.Name, workspace.Description, workspace.ImageName, workspace.Status)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		workspace.ID = int(id)
	}

	return nil
}

// GetWorkspaceByName retrieves a workspace by project ID and name.
func (ds *SQLDataStore) GetWorkspaceByName(projectID int, name string) (*models.Workspace, error) {
	workspace := &models.Workspace{}
	query := `SELECT id, project_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at 
		FROM workspaces WHERE project_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, projectID, name)
	if err := row.Scan(&workspace.ID, &workspace.ProjectID, &workspace.Name, &workspace.Description,
		&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure,
		&workspace.NvimPlugins, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workspace not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return workspace, nil
}

// GetWorkspaceByID retrieves a workspace by its ID.
func (ds *SQLDataStore) GetWorkspaceByID(id int) (*models.Workspace, error) {
	workspace := &models.Workspace{}
	query := `SELECT id, project_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at 
		FROM workspaces WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&workspace.ID, &workspace.ProjectID, &workspace.Name, &workspace.Description,
		&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure,
		&workspace.NvimPlugins, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workspace not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return workspace, nil
}

// UpdateWorkspace updates an existing workspace.
func (ds *SQLDataStore) UpdateWorkspace(workspace *models.Workspace) error {
	query := fmt.Sprintf(`UPDATE workspaces SET name = ?, description = ?, image_name = ?, container_id = ?, 
		status = ?, nvim_structure = ?, nvim_plugins = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, workspace.Name, workspace.Description, workspace.ImageName,
		workspace.ContainerID, workspace.Status, workspace.NvimStructure, workspace.NvimPlugins, workspace.ID)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}
	return nil
}

// DeleteWorkspace removes a workspace by ID.
func (ds *SQLDataStore) DeleteWorkspace(id int) error {
	query := `DELETE FROM workspaces WHERE id = ?`
	_, err := ds.driver.Execute(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}

// ListWorkspacesByProject retrieves all workspaces for a project.
func (ds *SQLDataStore) ListWorkspacesByProject(projectID int) ([]*models.Workspace, error) {
	query := `SELECT id, project_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at 
		FROM workspaces WHERE project_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*models.Workspace
	for rows.Next() {
		workspace := &models.Workspace{}
		if err := rows.Scan(&workspace.ID, &workspace.ProjectID, &workspace.Name, &workspace.Description,
			&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure,
			&workspace.NvimPlugins, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspaces: %w", err)
	}

	return workspaces, nil
}

// ListAllWorkspaces retrieves all workspaces across all projects.
func (ds *SQLDataStore) ListAllWorkspaces() ([]*models.Workspace, error) {
	query := `SELECT id, project_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at 
		FROM workspaces ORDER BY project_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*models.Workspace
	for rows.Next() {
		workspace := &models.Workspace{}
		if err := rows.Scan(&workspace.ID, &workspace.ProjectID, &workspace.Name, &workspace.Description,
			&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure,
			&workspace.NvimPlugins, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspaces: %w", err)
	}

	return workspaces, nil
}

// =============================================================================
// Context Operations
// =============================================================================

// GetContext retrieves the current context.
func (ds *SQLDataStore) GetContext() (*models.Context, error) {
	context := &models.Context{}
	query := `SELECT id, active_project_id, active_workspace_id, updated_at FROM context WHERE id = 1`

	row := ds.driver.QueryRow(query)
	if err := row.Scan(&context.ID, &context.ActiveProjectID, &context.ActiveWorkspaceID, &context.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("context not found")
		}
		return nil, fmt.Errorf("failed to scan context: %w", err)
	}

	return context, nil
}

// SetActiveProject sets the active project in the context.
func (ds *SQLDataStore) SetActiveProject(projectID *int) error {
	query := fmt.Sprintf(`UPDATE context SET active_project_id = ?, updated_at = %s WHERE id = 1`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to set active project: %w", err)
	}
	return nil
}

// SetActiveWorkspace sets the active workspace in the context.
func (ds *SQLDataStore) SetActiveWorkspace(workspaceID *int) error {
	query := fmt.Sprintf(`UPDATE context SET active_workspace_id = ?, updated_at = %s WHERE id = 1`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to set active workspace: %w", err)
	}
	return nil
}

// =============================================================================
// Plugin Operations
// =============================================================================

// CreatePlugin inserts a new nvim plugin.
func (ds *SQLDataStore) CreatePlugin(plugin *models.NvimPluginDB) error {
	query := fmt.Sprintf(`INSERT INTO nvim_plugins (name, description, repo, branch, version, priority, lazy, 
		event, ft, keys, cmd, dependencies, build, config, init, opts, keymaps, category, tags, enabled, 
		created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		plugin.Name, plugin.Description, plugin.Repo, plugin.Branch, plugin.Version, plugin.Priority,
		plugin.Lazy, plugin.Event, plugin.Ft, plugin.Keys, plugin.Cmd, plugin.Dependencies, plugin.Build,
		plugin.Config, plugin.Init, plugin.Opts, plugin.Keymaps, plugin.Category, plugin.Tags, plugin.Enabled)

	if err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		plugin.ID = int(id)
	}

	return nil
}

// GetPluginByName retrieves a plugin by its name.
func (ds *SQLDataStore) GetPluginByName(name string) (*models.NvimPluginDB, error) {
	plugin := &models.NvimPluginDB{}
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd, 
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(
		&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
		&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
		&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
		&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plugin not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan plugin: %w", err)
	}

	return plugin, nil
}

// GetPluginByID retrieves a plugin by its ID.
func (ds *SQLDataStore) GetPluginByID(id int) (*models.NvimPluginDB, error) {
	plugin := &models.NvimPluginDB{}
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd, 
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(
		&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
		&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
		&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
		&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plugin not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan plugin: %w", err)
	}

	return plugin, nil
}

// UpdatePlugin updates an existing plugin.
func (ds *SQLDataStore) UpdatePlugin(plugin *models.NvimPluginDB) error {
	query := fmt.Sprintf(`UPDATE nvim_plugins SET description = ?, repo = ?, branch = ?, version = ?, priority = ?, 
		lazy = ?, event = ?, ft = ?, keys = ?, cmd = ?, dependencies = ?, build = ?, config = ?, init = ?,
		opts = ?, keymaps = ?, category = ?, tags = ?, enabled = ?, updated_at = %s 
		WHERE name = ?`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		plugin.Description, plugin.Repo, plugin.Branch, plugin.Version, plugin.Priority,
		plugin.Lazy, plugin.Event, plugin.Ft, plugin.Keys, plugin.Cmd, plugin.Dependencies, plugin.Build,
		plugin.Config, plugin.Init, plugin.Opts, plugin.Keymaps, plugin.Category, plugin.Tags, plugin.Enabled,
		plugin.Name)

	if err != nil {
		return fmt.Errorf("failed to update plugin: %w", err)
	}
	return nil
}

// DeletePlugin removes a plugin by name.
func (ds *SQLDataStore) DeletePlugin(name string) error {
	query := `DELETE FROM nvim_plugins WHERE name = ?`
	_, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete plugin: %w", err)
	}
	return nil
}

// ListPlugins retrieves all plugins.
func (ds *SQLDataStore) ListPlugins() ([]*models.NvimPluginDB, error) {
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd,
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*models.NvimPluginDB
	for rows.Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over plugins: %w", err)
	}

	return plugins, nil
}

// ListPluginsByCategory retrieves plugins filtered by category.
func (ds *SQLDataStore) ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error) {
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd,
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE category = ? ORDER BY name`

	rows, err := ds.driver.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by category: %w", err)
	}
	defer rows.Close()

	var plugins []*models.NvimPluginDB
	for rows.Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over plugins: %w", err)
	}

	return plugins, nil
}

// ListPluginsByTags retrieves plugins that have any of the specified tags.
func (ds *SQLDataStore) ListPluginsByTags(tags []string) ([]*models.NvimPluginDB, error) {
	if len(tags) == 0 {
		return []*models.NvimPluginDB{}, nil
	}

	// Build query with LIKE clauses for each tag
	// Tags are stored as comma-separated string
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd,
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE `

	var conditions []string
	var args []interface{}
	for _, tag := range tags {
		conditions = append(conditions, "tags LIKE ?")
		args = append(args, "%"+tag+"%")
	}
	query += "(" + joinStrings(conditions, " OR ") + ") ORDER BY name"

	rows, err := ds.driver.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by tags: %w", err)
	}
	defer rows.Close()

	var plugins []*models.NvimPluginDB
	for rows.Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over plugins: %w", err)
	}

	return plugins, nil
}

// =============================================================================
// Workspace Plugin Associations
// =============================================================================

// AddPluginToWorkspace associates a plugin with a workspace.
func (ds *SQLDataStore) AddPluginToWorkspace(workspaceID int, pluginID int) error {
	query := fmt.Sprintf(`INSERT OR IGNORE INTO workspace_plugins (workspace_id, plugin_id, enabled, created_at)
		VALUES (?, ?, %s, %s)`, ds.queryBuilder.Boolean(true), ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, workspaceID, pluginID)
	if err != nil {
		return fmt.Errorf("failed to add plugin to workspace: %w", err)
	}
	return nil
}

// RemovePluginFromWorkspace removes a plugin association from a workspace.
func (ds *SQLDataStore) RemovePluginFromWorkspace(workspaceID int, pluginID int) error {
	query := `DELETE FROM workspace_plugins WHERE workspace_id = ? AND plugin_id = ?`
	_, err := ds.driver.Execute(query, workspaceID, pluginID)
	if err != nil {
		return fmt.Errorf("failed to remove plugin from workspace: %w", err)
	}
	return nil
}

// GetWorkspacePlugins retrieves all plugins associated with a workspace.
func (ds *SQLDataStore) GetWorkspacePlugins(workspaceID int) ([]*models.NvimPluginDB, error) {
	query := fmt.Sprintf(`SELECT p.id, p.name, p.description, p.repo, p.branch, p.version, p.priority, p.lazy, 
		p.event, p.ft, p.keys, p.cmd, p.dependencies, p.build, p.config, p.init, p.opts, p.keymaps,
		p.category, p.tags, p.enabled, p.created_at, p.updated_at
		FROM nvim_plugins p
		JOIN workspace_plugins wp ON p.id = wp.plugin_id
		WHERE wp.workspace_id = ? AND wp.enabled = %s
		ORDER BY p.priority DESC, p.name`, ds.queryBuilder.Boolean(true))

	rows, err := ds.driver.Query(query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*models.NvimPluginDB
	for rows.Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspace plugins: %w", err)
	}

	return plugins, nil
}

// SetWorkspacePluginEnabled enables or disables a plugin for a workspace.
func (ds *SQLDataStore) SetWorkspacePluginEnabled(workspaceID int, pluginID int, enabled bool) error {
	query := `UPDATE workspace_plugins SET enabled = ? WHERE workspace_id = ? AND plugin_id = ?`
	_, err := ds.driver.Execute(query, enabled, workspaceID, pluginID)
	if err != nil {
		return fmt.Errorf("failed to set workspace plugin enabled: %w", err)
	}
	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// joinStrings joins strings with a separator (to avoid importing strings package)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// Ensure SQLDataStore implements DataStore interface
var _ DataStore = (*SQLDataStore)(nil)
