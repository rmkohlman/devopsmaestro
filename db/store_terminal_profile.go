package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Terminal Profile Operations
// =============================================================================

// CreateTerminalProfile inserts a new terminal profile.
func (ds *SQLDataStore) CreateTerminalProfile(profile *models.TerminalProfileDB) error {
	query := fmt.Sprintf(`INSERT INTO terminal_profiles (name, description, category, prompt_ref, 
		plugin_refs, shell_ref, theme_ref, tags, labels, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		profile.Name, profile.Description, profile.Category, profile.PromptRef,
		profile.PluginRefs, profile.ShellRef, profile.ThemeRef, profile.Tags,
		profile.Labels, profile.Enabled)

	if err != nil {
		return fmt.Errorf("failed to create terminal profile: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		profile.ID = int(id)
	}

	return nil
}

// GetTerminalProfileByName retrieves a terminal profile by its name.
func (ds *SQLDataStore) GetTerminalProfileByName(name string) (*models.TerminalProfileDB, error) {
	profile := &models.TerminalProfileDB{}
	query := `SELECT id, name, description, category, prompt_ref, plugin_refs, shell_ref, 
		theme_ref, tags, labels, enabled, created_at, updated_at
		FROM terminal_profiles WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(
		&profile.ID, &profile.Name, &profile.Description, &profile.Category, &profile.PromptRef,
		&profile.PluginRefs, &profile.ShellRef, &profile.ThemeRef, &profile.Tags,
		&profile.Labels, &profile.Enabled, &profile.CreatedAt, &profile.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("terminal profile not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan terminal profile: %w", err)
	}

	return profile, nil
}

// UpdateTerminalProfile updates an existing terminal profile.
func (ds *SQLDataStore) UpdateTerminalProfile(profile *models.TerminalProfileDB) error {
	query := fmt.Sprintf(`UPDATE terminal_profiles SET description = ?, category = ?, prompt_ref = ?, 
		plugin_refs = ?, shell_ref = ?, theme_ref = ?, tags = ?, labels = ?, enabled = ?, 
		updated_at = %s WHERE name = ?`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		profile.Description, profile.Category, profile.PromptRef, profile.PluginRefs,
		profile.ShellRef, profile.ThemeRef, profile.Tags, profile.Labels, profile.Enabled, profile.Name)

	if err != nil {
		return fmt.Errorf("failed to update terminal profile: %w", err)
	}
	return nil
}

// UpsertTerminalProfile creates or updates a terminal profile.
func (ds *SQLDataStore) UpsertTerminalProfile(profile *models.TerminalProfileDB) error {
	// Check if the profile exists
	existing, err := ds.GetTerminalProfileByName(profile.Name)
	if err != nil {
		// Profile doesn't exist, create it
		return ds.CreateTerminalProfile(profile)
	}

	// Profile exists, update it with the existing ID
	profile.ID = existing.ID
	return ds.UpdateTerminalProfile(profile)
}

// DeleteTerminalProfile removes a terminal profile by name.
func (ds *SQLDataStore) DeleteTerminalProfile(name string) error {
	query := `DELETE FROM terminal_profiles WHERE name = ?`
	result, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete terminal profile: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("terminal profile", name)
	}
	return nil
}

// ListTerminalProfiles retrieves all terminal profiles.
func (ds *SQLDataStore) ListTerminalProfiles() ([]*models.TerminalProfileDB, error) {
	query := `SELECT id, name, description, category, prompt_ref, plugin_refs, shell_ref,
		theme_ref, tags, labels, enabled, created_at, updated_at
		FROM terminal_profiles ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal profiles: %w", err)
	}
	defer rows.Close()

	var profiles []*models.TerminalProfileDB
	for rows.Next() {
		profile := &models.TerminalProfileDB{}
		if err := rows.Scan(
			&profile.ID, &profile.Name, &profile.Description, &profile.Category, &profile.PromptRef,
			&profile.PluginRefs, &profile.ShellRef, &profile.ThemeRef, &profile.Tags,
			&profile.Labels, &profile.Enabled, &profile.CreatedAt, &profile.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal profile: %w", err)
		}
		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal profiles: %w", err)
	}

	return profiles, nil
}

// ListTerminalProfilesByCategory retrieves terminal profiles filtered by category.
func (ds *SQLDataStore) ListTerminalProfilesByCategory(category string) ([]*models.TerminalProfileDB, error) {
	query := `SELECT id, name, description, category, prompt_ref, plugin_refs, shell_ref,
		theme_ref, tags, labels, enabled, created_at, updated_at
		FROM terminal_profiles WHERE category = ? ORDER BY name`

	rows, err := ds.driver.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal profiles by category: %w", err)
	}
	defer rows.Close()

	var profiles []*models.TerminalProfileDB
	for rows.Next() {
		profile := &models.TerminalProfileDB{}
		if err := rows.Scan(
			&profile.ID, &profile.Name, &profile.Description, &profile.Category, &profile.PromptRef,
			&profile.PluginRefs, &profile.ShellRef, &profile.ThemeRef, &profile.Tags,
			&profile.Labels, &profile.Enabled, &profile.CreatedAt, &profile.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal profile: %w", err)
		}
		profiles = append(profiles, profile)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal profiles by category: %w", err)
	}

	return profiles, nil
}
