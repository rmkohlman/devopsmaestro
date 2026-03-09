package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Terminal Emulator Operations
// =============================================================================

// CreateTerminalEmulator inserts a new terminal emulator config.
func (ds *SQLDataStore) CreateTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	// Ensure required JSON fields have proper defaults
	if emulator.Config == "" {
		emulator.Config = "{}"
	}
	if emulator.Labels == "" {
		emulator.Labels = "{}"
	}

	query := fmt.Sprintf(`INSERT INTO terminal_emulators (name, description, type, config, theme_ref, category, 
		labels, workspace, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		emulator.Name, emulator.Description, emulator.Type, emulator.Config, emulator.ThemeRef,
		emulator.Category, emulator.Labels, emulator.Workspace, emulator.Enabled)
	if err != nil {
		return fmt.Errorf("failed to create terminal emulator: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get terminal emulator ID: %w", err)
	}

	emulator.ID = int(id)
	return nil
}

// GetTerminalEmulator retrieves a terminal emulator by its name.
func (ds *SQLDataStore) GetTerminalEmulator(name string) (*models.TerminalEmulatorDB, error) {
	query := `SELECT id, name, description, type, config, theme_ref, category, labels, workspace, enabled, created_at, updated_at
		FROM terminal_emulators WHERE name = ?`

	emulator := &models.TerminalEmulatorDB{}
	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(
		&emulator.ID, &emulator.Name, &emulator.Description, &emulator.Type, &emulator.Config,
		&emulator.ThemeRef, &emulator.Category, &emulator.Labels, &emulator.Workspace,
		&emulator.Enabled, &emulator.CreatedAt, &emulator.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("terminal emulator not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan terminal emulator: %w", err)
	}

	return emulator, nil
}

// UpdateTerminalEmulator updates an existing terminal emulator config.
func (ds *SQLDataStore) UpdateTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	// Ensure required JSON fields have proper defaults
	if emulator.Config == "" {
		emulator.Config = "{}"
	}
	if emulator.Labels == "" {
		emulator.Labels = "{}"
	}

	query := fmt.Sprintf(`UPDATE terminal_emulators SET description = ?, type = ?, config = ?, theme_ref = ?, 
		category = ?, labels = ?, workspace = ?, enabled = ?, updated_at = %s WHERE name = ?`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		emulator.Description, emulator.Type, emulator.Config, emulator.ThemeRef, emulator.Category,
		emulator.Labels, emulator.Workspace, emulator.Enabled, emulator.Name)
	if err != nil {
		return fmt.Errorf("failed to update terminal emulator: %w", err)
	}
	return nil
}

// UpsertTerminalEmulator creates or updates a terminal emulator (by name).
func (ds *SQLDataStore) UpsertTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	// First try to get the existing emulator
	existing, err := ds.GetTerminalEmulator(emulator.Name)
	if err == nil {
		// Emulator exists, update it
		emulator.ID = existing.ID
		return ds.UpdateTerminalEmulator(emulator)
	}

	// Emulator doesn't exist, create it
	return ds.CreateTerminalEmulator(emulator)
}

// DeleteTerminalEmulator removes a terminal emulator by name.
func (ds *SQLDataStore) DeleteTerminalEmulator(name string) error {
	query := `DELETE FROM terminal_emulators WHERE name = ?`
	result, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete terminal emulator: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("terminal emulator", name)
	}
	return nil
}

// ListTerminalEmulators retrieves all terminal emulators.
func (ds *SQLDataStore) ListTerminalEmulators() ([]*models.TerminalEmulatorDB, error) {
	query := `SELECT id, name, description, type, config, theme_ref, category, labels, workspace, enabled, created_at, updated_at
		FROM terminal_emulators ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal emulators: %w", err)
	}
	defer rows.Close()

	var emulators []*models.TerminalEmulatorDB
	for rows.Next() {
		emulator := &models.TerminalEmulatorDB{}
		if err := rows.Scan(
			&emulator.ID, &emulator.Name, &emulator.Description, &emulator.Type, &emulator.Config,
			&emulator.ThemeRef, &emulator.Category, &emulator.Labels, &emulator.Workspace,
			&emulator.Enabled, &emulator.CreatedAt, &emulator.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal emulator: %w", err)
		}
		emulators = append(emulators, emulator)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal emulators: %w", err)
	}

	return emulators, nil
}

// ListTerminalEmulatorsByType retrieves terminal emulators filtered by type.
func (ds *SQLDataStore) ListTerminalEmulatorsByType(emulatorType string) ([]*models.TerminalEmulatorDB, error) {
	query := `SELECT id, name, description, type, config, theme_ref, category, labels, workspace, enabled, created_at, updated_at
		FROM terminal_emulators WHERE type = ? ORDER BY name`

	rows, err := ds.driver.Query(query, emulatorType)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal emulators by type: %w", err)
	}
	defer rows.Close()

	var emulators []*models.TerminalEmulatorDB
	for rows.Next() {
		emulator := &models.TerminalEmulatorDB{}
		if err := rows.Scan(
			&emulator.ID, &emulator.Name, &emulator.Description, &emulator.Type, &emulator.Config,
			&emulator.ThemeRef, &emulator.Category, &emulator.Labels, &emulator.Workspace,
			&emulator.Enabled, &emulator.CreatedAt, &emulator.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal emulator: %w", err)
		}
		emulators = append(emulators, emulator)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal emulators by type: %w", err)
	}

	return emulators, nil
}

// ListTerminalEmulatorsByWorkspace retrieves terminal emulators for a workspace.
func (ds *SQLDataStore) ListTerminalEmulatorsByWorkspace(workspace string) ([]*models.TerminalEmulatorDB, error) {
	query := `SELECT id, name, description, type, config, theme_ref, category, labels, workspace, enabled, created_at, updated_at
		FROM terminal_emulators WHERE workspace = ? ORDER BY name`

	rows, err := ds.driver.Query(query, workspace)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal emulators by workspace: %w", err)
	}
	defer rows.Close()

	var emulators []*models.TerminalEmulatorDB
	for rows.Next() {
		emulator := &models.TerminalEmulatorDB{}
		if err := rows.Scan(
			&emulator.ID, &emulator.Name, &emulator.Description, &emulator.Type, &emulator.Config,
			&emulator.ThemeRef, &emulator.Category, &emulator.Labels, &emulator.Workspace,
			&emulator.Enabled, &emulator.CreatedAt, &emulator.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal emulator: %w", err)
		}
		emulators = append(emulators, emulator)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal emulators by workspace: %w", err)
	}

	return emulators, nil
}
