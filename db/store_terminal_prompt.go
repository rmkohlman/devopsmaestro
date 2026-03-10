package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Terminal Prompt Operations
// =============================================================================

// scanTerminalPrompt scans a single row into a TerminalPromptDB struct.
func scanTerminalPrompt(s interface{ Scan(dest ...any) error }) (*models.TerminalPromptDB, error) {
	prompt := &models.TerminalPromptDB{}
	if err := s.Scan(
		&prompt.ID, &prompt.Name, &prompt.Description, &prompt.Type, &prompt.AddNewline,
		&prompt.Palette, &prompt.Format, &prompt.Modules, &prompt.Character, &prompt.PaletteRef,
		&prompt.Colors, &prompt.RawConfig, &prompt.Category, &prompt.Tags, &prompt.Enabled,
		&prompt.CreatedAt, &prompt.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return prompt, nil
}

// CreateTerminalPrompt inserts a new terminal prompt.
func (ds *SQLDataStore) CreateTerminalPrompt(prompt *models.TerminalPromptDB) error {
	query := fmt.Sprintf(`INSERT INTO terminal_prompts (name, description, type, add_newline, palette, format,
		modules, character, palette_ref, colors, raw_config, category, tags, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		prompt.Name, prompt.Description, prompt.Type, prompt.AddNewline, prompt.Palette, prompt.Format,
		prompt.Modules, prompt.Character, prompt.PaletteRef, prompt.Colors, prompt.RawConfig,
		prompt.Category, prompt.Tags, prompt.Enabled)

	if err != nil {
		return fmt.Errorf("failed to create terminal prompt: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		prompt.ID = int(id)
	}

	return nil
}

// GetTerminalPromptByName retrieves a terminal prompt by its name.
func (ds *SQLDataStore) GetTerminalPromptByName(name string) (*models.TerminalPromptDB, error) {
	query := `SELECT id, name, description, type, add_newline, palette, format, modules, character, 
		palette_ref, colors, raw_config, category, tags, enabled, created_at, updated_at
		FROM terminal_prompts WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	prompt, err := scanTerminalPrompt(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("terminal prompt not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan terminal prompt: %w", err)
	}

	return prompt, nil
}

// UpdateTerminalPrompt updates an existing terminal prompt.
func (ds *SQLDataStore) UpdateTerminalPrompt(prompt *models.TerminalPromptDB) error {
	query := fmt.Sprintf(`UPDATE terminal_prompts SET description = ?, type = ?, add_newline = ?, palette = ?, 
		format = ?, modules = ?, character = ?, palette_ref = ?, colors = ?, raw_config = ?, 
		category = ?, tags = ?, enabled = ?, updated_at = %s 
		WHERE name = ?`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		prompt.Description, prompt.Type, prompt.AddNewline, prompt.Palette, prompt.Format,
		prompt.Modules, prompt.Character, prompt.PaletteRef, prompt.Colors, prompt.RawConfig,
		prompt.Category, prompt.Tags, prompt.Enabled, prompt.Name)

	if err != nil {
		return fmt.Errorf("failed to update terminal prompt: %w", err)
	}
	return nil
}

// UpsertTerminalPrompt creates or updates a terminal prompt.
func (ds *SQLDataStore) UpsertTerminalPrompt(prompt *models.TerminalPromptDB) error {
	// Check if the prompt exists
	existing, err := ds.GetTerminalPromptByName(prompt.Name)
	if err != nil {
		// Prompt doesn't exist, create it
		return ds.CreateTerminalPrompt(prompt)
	}

	// Prompt exists, update it with the existing ID
	prompt.ID = existing.ID
	return ds.UpdateTerminalPrompt(prompt)
}

// DeleteTerminalPrompt removes a terminal prompt by name.
func (ds *SQLDataStore) DeleteTerminalPrompt(name string) error {
	return ds.deleteByName("terminal_prompts", "terminal prompt", name)
}

// ListTerminalPrompts retrieves all terminal prompts.
func (ds *SQLDataStore) ListTerminalPrompts() ([]*models.TerminalPromptDB, error) {
	query := `SELECT id, name, description, type, add_newline, palette, format, modules, character,
		palette_ref, colors, raw_config, category, tags, enabled, created_at, updated_at
		FROM terminal_prompts ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal prompts: %w", err)
	}
	defer rows.Close()

	var prompts []*models.TerminalPromptDB
	for rows.Next() {
		prompt, err := scanTerminalPrompt(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan terminal prompt: %w", err)
		}
		prompts = append(prompts, prompt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal prompts: %w", err)
	}

	return prompts, nil
}

// ListTerminalPromptsByType retrieves terminal prompts filtered by type.
func (ds *SQLDataStore) ListTerminalPromptsByType(promptType string) ([]*models.TerminalPromptDB, error) {
	query := `SELECT id, name, description, type, add_newline, palette, format, modules, character,
		palette_ref, colors, raw_config, category, tags, enabled, created_at, updated_at
		FROM terminal_prompts WHERE type = ? ORDER BY name`

	rows, err := ds.driver.Query(query, promptType)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal prompts by type: %w", err)
	}
	defer rows.Close()

	var prompts []*models.TerminalPromptDB
	for rows.Next() {
		prompt, err := scanTerminalPrompt(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan terminal prompt: %w", err)
		}
		prompts = append(prompts, prompt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal prompts by type: %w", err)
	}

	return prompts, nil
}

// ListTerminalPromptsByCategory retrieves terminal prompts filtered by category.
func (ds *SQLDataStore) ListTerminalPromptsByCategory(category string) ([]*models.TerminalPromptDB, error) {
	query := `SELECT id, name, description, type, add_newline, palette, format, modules, character,
		palette_ref, colors, raw_config, category, tags, enabled, created_at, updated_at
		FROM terminal_prompts WHERE category = ? ORDER BY name`

	rows, err := ds.driver.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal prompts by category: %w", err)
	}
	defer rows.Close()

	var prompts []*models.TerminalPromptDB
	for rows.Next() {
		prompt, err := scanTerminalPrompt(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan terminal prompt: %w", err)
		}
		prompts = append(prompts, prompt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal prompts by category: %w", err)
	}

	return prompts, nil
}
