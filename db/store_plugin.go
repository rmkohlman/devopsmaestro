package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
	"strings"
)

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
	return ds.deleteByName("nvim_plugins", "plugin", name)
}

// UpsertPlugin creates or updates a plugin (by name).
func (ds *SQLDataStore) UpsertPlugin(plugin *models.NvimPluginDB) error {
	// First try to get the existing plugin
	existing, err := ds.GetPluginByName(plugin.Name)
	if err == nil {
		// Plugin exists, update it
		plugin.ID = existing.ID
		return ds.UpdatePlugin(plugin)
	}

	// Plugin doesn't exist, create it
	return ds.CreatePlugin(plugin)
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
	query += "(" + strings.Join(conditions, " OR ") + ") ORDER BY name"

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
