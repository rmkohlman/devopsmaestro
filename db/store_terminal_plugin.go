package db

import (
	"database/sql"
	"errors"
	"fmt"

	"devopsmaestro/models"
)

// =============================================================================
// Terminal Plugin Operations
// =============================================================================

// scanTerminalPlugin scans a single row into a TerminalPluginDB struct.
func scanTerminalPlugin(s interface{ Scan(dest ...any) error }) (*models.TerminalPluginDB, error) {
	plugin := &models.TerminalPluginDB{}
	if err := s.Scan(
		&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Category,
		&plugin.Shell, &plugin.Manager, &plugin.LoadCommand, &plugin.SourceFile,
		&plugin.Dependencies, &plugin.EnvVars, &plugin.Labels, &plugin.Enabled,
		&plugin.CreatedAt, &plugin.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return plugin, nil
}

// CreateTerminalPlugin inserts a new terminal plugin.
func (ds *SQLDataStore) CreateTerminalPlugin(plugin *models.TerminalPluginDB) error {
	// Ensure required JSON fields have proper defaults
	if plugin.Dependencies == "" {
		plugin.Dependencies = "[]"
	}
	if plugin.EnvVars == "" {
		plugin.EnvVars = "{}"
	}
	if plugin.Labels == "" {
		plugin.Labels = "{}"
	}

	query := fmt.Sprintf(`INSERT INTO terminal_plugins (name, description, repo, category, shell, manager, 
		load_command, source_file, dependencies, env_vars, labels, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		plugin.Name, plugin.Description, plugin.Repo, plugin.Category, plugin.Shell, plugin.Manager,
		plugin.LoadCommand, plugin.SourceFile, plugin.Dependencies, plugin.EnvVars, plugin.Labels, plugin.Enabled)

	if err != nil {
		return fmt.Errorf("failed to create terminal plugin: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		plugin.ID = int(id)

		// Read back the created record to get timestamps
		created, err := ds.GetTerminalPlugin(plugin.Name)
		if err == nil {
			plugin.CreatedAt = created.CreatedAt
			plugin.UpdatedAt = created.UpdatedAt
		}
	}

	return nil
}

// GetTerminalPlugin retrieves a terminal plugin by its name.
func (ds *SQLDataStore) GetTerminalPlugin(name string) (*models.TerminalPluginDB, error) {
	query := `SELECT id, name, description, repo, category, shell, manager, load_command, 
		source_file, dependencies, env_vars, labels, enabled, created_at, updated_at
		FROM terminal_plugins WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	plugin, err := scanTerminalPlugin(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("terminal plugin", name)
		}
		return nil, fmt.Errorf("failed to scan terminal plugin: %w", err)
	}

	return plugin, nil
}

// UpdateTerminalPlugin updates an existing terminal plugin.
func (ds *SQLDataStore) UpdateTerminalPlugin(plugin *models.TerminalPluginDB) error {
	query := fmt.Sprintf(`UPDATE terminal_plugins SET description = ?, repo = ?, category = ?, shell = ?, 
		manager = ?, load_command = ?, source_file = ?, dependencies = ?, env_vars = ?, labels = ?, 
		enabled = ?, updated_at = %s 
		WHERE name = ?`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		plugin.Description, plugin.Repo, plugin.Category, plugin.Shell, plugin.Manager,
		plugin.LoadCommand, plugin.SourceFile, plugin.Dependencies, plugin.EnvVars, plugin.Labels,
		plugin.Enabled, plugin.Name)

	if err != nil {
		return fmt.Errorf("failed to update terminal plugin: %w", err)
	}
	return nil
}

// UpsertTerminalPlugin creates or updates a terminal plugin (by name) atomically using ON CONFLICT.
func (ds *SQLDataStore) UpsertTerminalPlugin(plugin *models.TerminalPluginDB) error {
	// Ensure required JSON fields have proper defaults
	if plugin.Dependencies == "" {
		plugin.Dependencies = "[]"
	}
	if plugin.EnvVars == "" {
		plugin.EnvVars = "{}"
	}
	if plugin.Labels == "" {
		plugin.Labels = "{}"
	}

	query := fmt.Sprintf(`INSERT INTO terminal_plugins (name, description, repo, category, shell, manager, 
		load_command, source_file, dependencies, env_vars, labels, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)
		%s, updated_at = %s`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now(),
		ds.queryBuilder.UpsertSuffix([]string{"name"}, []string{
			"description", "repo", "category", "shell", "manager", "load_command",
			"source_file", "dependencies", "env_vars", "labels", "enabled",
		}),
		ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		plugin.Name, plugin.Description, plugin.Repo, plugin.Category, plugin.Shell, plugin.Manager,
		plugin.LoadCommand, plugin.SourceFile, plugin.Dependencies, plugin.EnvVars, plugin.Labels, plugin.Enabled)
	if err != nil {
		return fmt.Errorf("failed to upsert terminal plugin: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		plugin.ID = int(id)
	}

	return nil
}

// DeleteTerminalPlugin removes a terminal plugin by name.
func (ds *SQLDataStore) DeleteTerminalPlugin(name string) error {
	return ds.deleteByName("terminal_plugins", "terminal plugin", name)
}

// ListTerminalPlugins retrieves all terminal plugins.
func (ds *SQLDataStore) ListTerminalPlugins() ([]*models.TerminalPluginDB, error) {
	query := `SELECT id, name, description, repo, category, shell, manager, load_command,
		source_file, dependencies, env_vars, labels, enabled, created_at, updated_at
		FROM terminal_plugins ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*models.TerminalPluginDB
	for rows.Next() {
		plugin, err := scanTerminalPlugin(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan terminal plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal plugins: %w", err)
	}

	return plugins, nil
}

// ListTerminalPluginsByCategory retrieves terminal plugins filtered by category.
func (ds *SQLDataStore) ListTerminalPluginsByCategory(category string) ([]*models.TerminalPluginDB, error) {
	query := `SELECT id, name, description, repo, category, shell, manager, load_command,
		source_file, dependencies, env_vars, labels, enabled, created_at, updated_at
		FROM terminal_plugins WHERE category = ? ORDER BY name`

	rows, err := ds.driver.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal plugins by category: %w", err)
	}
	defer rows.Close()

	var plugins []*models.TerminalPluginDB
	for rows.Next() {
		plugin, err := scanTerminalPlugin(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan terminal plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal plugins by category: %w", err)
	}

	return plugins, nil
}

// ListTerminalPluginsByShell retrieves terminal plugins filtered by shell.
func (ds *SQLDataStore) ListTerminalPluginsByShell(shell string) ([]*models.TerminalPluginDB, error) {
	query := `SELECT id, name, description, repo, category, shell, manager, load_command,
		source_file, dependencies, env_vars, labels, enabled, created_at, updated_at
		FROM terminal_plugins WHERE shell = ? ORDER BY name`

	rows, err := ds.driver.Query(query, shell)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal plugins by shell: %w", err)
	}
	defer rows.Close()

	var plugins []*models.TerminalPluginDB
	for rows.Next() {
		plugin, err := scanTerminalPlugin(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan terminal plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal plugins by shell: %w", err)
	}

	return plugins, nil
}

// ListTerminalPluginsByManager retrieves terminal plugins filtered by manager.
func (ds *SQLDataStore) ListTerminalPluginsByManager(manager string) ([]*models.TerminalPluginDB, error) {
	query := `SELECT id, name, description, repo, category, shell, manager, load_command,
		source_file, dependencies, env_vars, labels, enabled, created_at, updated_at
		FROM terminal_plugins WHERE manager = ? ORDER BY name`

	rows, err := ds.driver.Query(query, manager)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal plugins by manager: %w", err)
	}
	defer rows.Close()

	var plugins []*models.TerminalPluginDB
	for rows.Next() {
		plugin, err := scanTerminalPlugin(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan terminal plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal plugins by manager: %w", err)
	}

	return plugins, nil
}
