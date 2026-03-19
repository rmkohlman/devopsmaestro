package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Theme Operations
// =============================================================================

// CreateTheme inserts a new nvim theme.
func (ds *SQLDataStore) CreateTheme(theme *models.NvimThemeDB) error {
	query := fmt.Sprintf(`INSERT INTO nvim_themes (name, description, author, category, plugin_repo, 
		plugin_branch, plugin_tag, style, transparent, colors, options, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		theme.Name, theme.Description, theme.Author, theme.Category, theme.PluginRepo,
		theme.PluginBranch, theme.PluginTag, theme.Style, theme.Transparent,
		theme.Colors, theme.Options, theme.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create theme: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		theme.ID = int(id)
	}

	return nil
}

// GetThemeByName retrieves a theme by its name.
func (ds *SQLDataStore) GetThemeByName(name string) (*models.NvimThemeDB, error) {
	theme := &models.NvimThemeDB{}
	query := `SELECT id, name, description, author, category, plugin_repo, plugin_branch, plugin_tag,
		style, transparent, colors, options, is_active, created_at, updated_at
		FROM nvim_themes WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(
		&theme.ID, &theme.Name, &theme.Description, &theme.Author, &theme.Category, &theme.PluginRepo,
		&theme.PluginBranch, &theme.PluginTag, &theme.Style, &theme.Transparent,
		&theme.Colors, &theme.Options, &theme.IsActive, &theme.CreatedAt, &theme.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("theme", name)
		}
		return nil, fmt.Errorf("failed to scan theme: %w", err)
	}

	return theme, nil
}

// GetThemeByID retrieves a theme by its ID.
func (ds *SQLDataStore) GetThemeByID(id int) (*models.NvimThemeDB, error) {
	theme := &models.NvimThemeDB{}
	query := `SELECT id, name, description, author, category, plugin_repo, plugin_branch, plugin_tag,
		style, transparent, colors, options, is_active, created_at, updated_at
		FROM nvim_themes WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(
		&theme.ID, &theme.Name, &theme.Description, &theme.Author, &theme.Category, &theme.PluginRepo,
		&theme.PluginBranch, &theme.PluginTag, &theme.Style, &theme.Transparent,
		&theme.Colors, &theme.Options, &theme.IsActive, &theme.CreatedAt, &theme.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("theme", id)
		}
		return nil, fmt.Errorf("failed to scan theme: %w", err)
	}

	return theme, nil
}

// UpdateTheme updates an existing theme.
func (ds *SQLDataStore) UpdateTheme(theme *models.NvimThemeDB) error {
	query := fmt.Sprintf(`UPDATE nvim_themes SET description = ?, author = ?, category = ?, plugin_repo = ?,
		plugin_branch = ?, plugin_tag = ?, style = ?, transparent = ?, colors = ?, options = ?,
		is_active = ?, updated_at = %s WHERE name = ?`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		theme.Description, theme.Author, theme.Category, theme.PluginRepo,
		theme.PluginBranch, theme.PluginTag, theme.Style, theme.Transparent,
		theme.Colors, theme.Options, theme.IsActive, theme.Name)

	if err != nil {
		return fmt.Errorf("failed to update theme: %w", err)
	}
	return nil
}

// DeleteTheme removes a theme by name.
func (ds *SQLDataStore) DeleteTheme(name string) error {
	return ds.deleteByName("nvim_themes", "theme", name)
}

// ListThemes retrieves all themes.
func (ds *SQLDataStore) ListThemes() ([]*models.NvimThemeDB, error) {
	query := `SELECT id, name, description, author, category, plugin_repo, plugin_branch, plugin_tag,
		style, transparent, colors, options, is_active, created_at, updated_at
		FROM nvim_themes ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list themes: %w", err)
	}
	defer rows.Close()

	var themes []*models.NvimThemeDB
	for rows.Next() {
		theme := &models.NvimThemeDB{}
		if err := rows.Scan(
			&theme.ID, &theme.Name, &theme.Description, &theme.Author, &theme.Category, &theme.PluginRepo,
			&theme.PluginBranch, &theme.PluginTag, &theme.Style, &theme.Transparent,
			&theme.Colors, &theme.Options, &theme.IsActive, &theme.CreatedAt, &theme.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan theme: %w", err)
		}
		themes = append(themes, theme)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over themes: %w", err)
	}

	return themes, nil
}

// ListThemesByCategory retrieves themes filtered by category.
func (ds *SQLDataStore) ListThemesByCategory(category string) ([]*models.NvimThemeDB, error) {
	query := `SELECT id, name, description, author, category, plugin_repo, plugin_branch, plugin_tag,
		style, transparent, colors, options, is_active, created_at, updated_at
		FROM nvim_themes WHERE category = ? ORDER BY name`

	rows, err := ds.driver.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list themes by category: %w", err)
	}
	defer rows.Close()

	var themes []*models.NvimThemeDB
	for rows.Next() {
		theme := &models.NvimThemeDB{}
		if err := rows.Scan(
			&theme.ID, &theme.Name, &theme.Description, &theme.Author, &theme.Category, &theme.PluginRepo,
			&theme.PluginBranch, &theme.PluginTag, &theme.Style, &theme.Transparent,
			&theme.Colors, &theme.Options, &theme.IsActive, &theme.CreatedAt, &theme.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan theme: %w", err)
		}
		themes = append(themes, theme)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over themes: %w", err)
	}

	return themes, nil
}

// GetActiveTheme retrieves the currently active theme.
func (ds *SQLDataStore) GetActiveTheme() (*models.NvimThemeDB, error) {
	theme := &models.NvimThemeDB{}
	query := `SELECT id, name, description, author, category, plugin_repo, plugin_branch, plugin_tag,
		style, transparent, colors, options, is_active, created_at, updated_at
		FROM nvim_themes WHERE is_active = 1 LIMIT 1`

	row := ds.driver.QueryRow(query)
	if err := row.Scan(
		&theme.ID, &theme.Name, &theme.Description, &theme.Author, &theme.Category, &theme.PluginRepo,
		&theme.PluginBranch, &theme.PluginTag, &theme.Style, &theme.Transparent,
		&theme.Colors, &theme.Options, &theme.IsActive, &theme.CreatedAt, &theme.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active theme
		}
		return nil, fmt.Errorf("failed to scan active theme: %w", err)
	}

	return theme, nil
}

// SetActiveTheme sets the active theme by name (deactivates others).
// The deactivate-all + activate-one runs in a transaction to ensure atomicity.
func (ds *SQLDataStore) SetActiveTheme(name string) error {
	// First, verify the theme exists (outside transaction — read-only check)
	if _, err := ds.GetThemeByName(name); err != nil {
		return err
	}

	tx, err := ds.driver.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	// Deactivate all themes
	if _, err := tx.Execute(`UPDATE nvim_themes SET is_active = 0`); err != nil {
		return fmt.Errorf("failed to clear active theme: %w", err)
	}

	// Activate the specified theme
	query := fmt.Sprintf(`UPDATE nvim_themes SET is_active = 1, updated_at = %s WHERE name = ?`,
		ds.queryBuilder.Now())
	if _, err := tx.Execute(query, name); err != nil {
		return fmt.Errorf("failed to set active theme: %w", err)
	}

	return tx.Commit()
}

// ClearActiveTheme deactivates all themes.
func (ds *SQLDataStore) ClearActiveTheme() error {
	query := `UPDATE nvim_themes SET is_active = 0`
	_, err := ds.driver.Execute(query)
	if err != nil {
		return fmt.Errorf("failed to clear active theme: %w", err)
	}
	return nil
}
