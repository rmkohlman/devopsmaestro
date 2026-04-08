package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"devopsmaestro/models"
)

// =============================================================================
// Registry History Operations
// =============================================================================

// CreateRegistryHistory inserts a new registry history entry.
func (ds *SQLDataStore) CreateRegistryHistory(history *models.RegistryHistory) error {
	query := fmt.Sprintf(`INSERT INTO registry_history 
		(registry_id, revision, config, enabled, lifecycle, port, storage, idle_timeout, 
		 action, status, user, error_message, previous_revision, registry_version, created_at, completed_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, ?)`, ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		history.RegistryID,
		history.Revision,
		history.Config,
		history.Enabled,
		history.Lifecycle,
		history.Port,
		history.Storage,
		history.IdleTimeout,
		history.Action,
		history.Status,
		history.User,
		history.ErrorMessage,
		history.PreviousRevision,
		history.RegistryVersion,
		history.CompletedAt,
	)

	if err != nil {
		// Check for unique constraint violations
		if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "unique") {
			return NewErrUniqueViolation("revision", fmt.Sprintf("registry_id=%d, revision=%d", history.RegistryID, history.Revision))
		}
		return fmt.Errorf("failed to create registry history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}
	history.ID = id

	// Get the created_at timestamp
	row := ds.driver.QueryRow("SELECT created_at FROM registry_history WHERE id = ?", history.ID)
	if err := row.Scan(&history.CreatedAt); err != nil {
		// Non-fatal: ID is set, just timestamp missing
		return nil
	}

	return nil
}

// GetRegistryHistory retrieves a specific registry history entry by registryID and revision.
func (ds *SQLDataStore) GetRegistryHistory(registryID int, revision int) (*models.RegistryHistory, error) {
	query := `SELECT id, registry_id, revision, config, enabled, lifecycle, port, storage, idle_timeout,
		action, status, user, error_message, previous_revision, registry_version, created_at, completed_at
		FROM registry_history 
		WHERE registry_id = ? AND revision = ?`

	row := ds.driver.QueryRow(query, registryID, revision)

	history := &models.RegistryHistory{}
	err := row.Scan(
		&history.ID,
		&history.RegistryID,
		&history.Revision,
		&history.Config,
		&history.Enabled,
		&history.Lifecycle,
		&history.Port,
		&history.Storage,
		&history.IdleTimeout,
		&history.Action,
		&history.Status,
		&history.User,
		&history.ErrorMessage,
		&history.PreviousRevision,
		&history.RegistryVersion,
		&history.CreatedAt,
		&history.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("registry history", fmt.Sprintf("registry_id=%d, revision=%d", registryID, revision))
		}
		return nil, fmt.Errorf("failed to get registry history: %w", err)
	}

	return history, nil
}

// GetLatestRegistryHistory retrieves the most recent history entry for a registry.
func (ds *SQLDataStore) GetLatestRegistryHistory(registryID int) (*models.RegistryHistory, error) {
	query := `SELECT id, registry_id, revision, config, enabled, lifecycle, port, storage, idle_timeout,
		action, status, user, error_message, previous_revision, registry_version, created_at, completed_at
		FROM registry_history 
		WHERE registry_id = ?
		ORDER BY revision DESC
		LIMIT 1`

	row := ds.driver.QueryRow(query, registryID)

	history := &models.RegistryHistory{}
	err := row.Scan(
		&history.ID,
		&history.RegistryID,
		&history.Revision,
		&history.Config,
		&history.Enabled,
		&history.Lifecycle,
		&history.Port,
		&history.Storage,
		&history.IdleTimeout,
		&history.Action,
		&history.Status,
		&history.User,
		&history.ErrorMessage,
		&history.PreviousRevision,
		&history.RegistryVersion,
		&history.CreatedAt,
		&history.CompletedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("registry history", fmt.Sprintf("registry_id=%d", registryID))
		}
		return nil, fmt.Errorf("failed to get latest registry history: %w", err)
	}

	return history, nil
}

// ListRegistryHistory retrieves all history entries for a registry, ordered by revision DESC.
func (ds *SQLDataStore) ListRegistryHistory(registryID int) ([]*models.RegistryHistory, error) {
	query := `SELECT id, registry_id, revision, config, enabled, lifecycle, port, storage, idle_timeout,
		action, status, user, error_message, previous_revision, registry_version, created_at, completed_at
		FROM registry_history 
		WHERE registry_id = ?
		ORDER BY revision DESC`

	rows, err := ds.driver.Query(query, registryID)
	if err != nil {
		return nil, fmt.Errorf("failed to list registry history: %w", err)
	}
	defer rows.Close()

	var histories []*models.RegistryHistory
	for rows.Next() {
		history := &models.RegistryHistory{}
		err := rows.Scan(
			&history.ID,
			&history.RegistryID,
			&history.Revision,
			&history.Config,
			&history.Enabled,
			&history.Lifecycle,
			&history.Port,
			&history.Storage,
			&history.IdleTimeout,
			&history.Action,
			&history.Status,
			&history.User,
			&history.ErrorMessage,
			&history.PreviousRevision,
			&history.RegistryVersion,
			&history.CreatedAt,
			&history.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registry history: %w", err)
		}
		histories = append(histories, history)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating registry history: %w", err)
	}

	return histories, nil
}

// GetNextRevisionNumber returns the next available revision number for a registry.
// Returns 1 if no history exists yet.
func (ds *SQLDataStore) GetNextRevisionNumber(registryID int) (int, error) {
	query := `SELECT COALESCE(MAX(revision), 0) + 1 FROM registry_history WHERE registry_id = ?`

	row := ds.driver.QueryRow(query, registryID)

	var nextRevision int
	err := row.Scan(&nextRevision)
	if err != nil {
		return 0, fmt.Errorf("failed to get next revision number: %w", err)
	}

	return nextRevision, nil
}
