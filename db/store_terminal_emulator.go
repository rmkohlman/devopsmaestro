package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Terminal Emulator Operations
// =============================================================================

// scanTerminalEmulator scans a single row into a TerminalEmulatorDB struct.
func scanTerminalEmulator(s interface{ Scan(dest ...any) error }) (*models.TerminalEmulatorDB, error) {
	emulator := &models.TerminalEmulatorDB{}
	if err := s.Scan(
		&emulator.ID, &emulator.Name, &emulator.Description, &emulator.Type, &emulator.Config,
		&emulator.ThemeRef, &emulator.Category, &emulator.Labels, &emulator.Workspace,
		&emulator.Enabled, &emulator.CreatedAt, &emulator.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return emulator, nil
}

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

	row := ds.driver.QueryRow(query, name)
	emulator, err := scanTerminalEmulator(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("terminal emulator", name)
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

// UpsertTerminalEmulator creates or updates a terminal emulator (by name) atomically using ON CONFLICT.
func (ds *SQLDataStore) UpsertTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	// Ensure required JSON fields have proper defaults
	if emulator.Config == "" {
		emulator.Config = "{}"
	}
	if emulator.Labels == "" {
		emulator.Labels = "{}"
	}

	query := fmt.Sprintf(`INSERT INTO terminal_emulators (name, description, type, config, theme_ref, category, 
		labels, workspace, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)
		%s, updated_at = %s`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now(),
		ds.queryBuilder.UpsertSuffix([]string{"name"}, []string{
			"description", "type", "config", "theme_ref", "category", "labels", "workspace", "enabled",
		}),
		ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		emulator.Name, emulator.Description, emulator.Type, emulator.Config, emulator.ThemeRef,
		emulator.Category, emulator.Labels, emulator.Workspace, emulator.Enabled)
	if err != nil {
		return fmt.Errorf("failed to upsert terminal emulator: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		emulator.ID = int(id)
	}

	return nil
}

// DeleteTerminalEmulator removes a terminal emulator by name.
func (ds *SQLDataStore) DeleteTerminalEmulator(name string) error {
	return ds.deleteByName("terminal_emulators", "terminal emulator", name)
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
		emulator, err := scanTerminalEmulator(rows)
		if err != nil {
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
		emulator, err := scanTerminalEmulator(rows)
		if err != nil {
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
		emulator, err := scanTerminalEmulator(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan terminal emulator: %w", err)
		}
		emulators = append(emulators, emulator)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal emulators by workspace: %w", err)
	}

	return emulators, nil
}
