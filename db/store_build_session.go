package db

import (
	"database/sql"
	"errors"
	"fmt"

	"devopsmaestro/models"
)

// =============================================================================
// Build Session Operations
// =============================================================================

// CreateBuildSession inserts a new build session.
func (ds *SQLDataStore) CreateBuildSession(session *models.BuildSession) error {
	query := fmt.Sprintf(`INSERT INTO build_sessions 
		(id, started_at, completed_at, status, total_workspaces, succeeded, failed, created_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, %s)`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		session.ID,
		session.StartedAt,
		session.CompletedAt,
		session.Status,
		session.TotalWorkspaces,
		session.Succeeded,
		session.Failed,
	)
	if err != nil {
		return fmt.Errorf("failed to create build session: %w", err)
	}

	// Read back created_at timestamp
	row := ds.driver.QueryRow("SELECT created_at FROM build_sessions WHERE id = ?", session.ID)
	if err := row.Scan(&session.CreatedAt); err != nil {
		// Non-fatal: session is created, just timestamp missing
		return nil
	}

	return nil
}

// UpdateBuildSession updates an existing build session.
func (ds *SQLDataStore) UpdateBuildSession(session *models.BuildSession) error {
	query := `UPDATE build_sessions 
		SET completed_at = ?, status = ?, total_workspaces = ?, succeeded = ?, failed = ? 
		WHERE id = ?`

	result, err := ds.driver.Execute(query,
		session.CompletedAt,
		session.Status,
		session.TotalWorkspaces,
		session.Succeeded,
		session.Failed,
		session.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update build session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return NewErrNotFound("build session", session.ID)
	}

	return nil
}

// buildSessionColumns is the column list for build_sessions queries.
const buildSessionColumns = `id, started_at, completed_at, status, total_workspaces, succeeded, failed, created_at`

// scanBuildSession scans a single build session row into a BuildSession struct.
func scanBuildSession(row Row) (*models.BuildSession, error) {
	session := &models.BuildSession{}
	err := row.Scan(
		&session.ID,
		&session.StartedAt,
		&session.CompletedAt,
		&session.Status,
		&session.TotalWorkspaces,
		&session.Succeeded,
		&session.Failed,
		&session.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// GetLatestBuildSession retrieves the most recent build session.
func (ds *SQLDataStore) GetLatestBuildSession() (*models.BuildSession, error) {
	query := `SELECT ` + buildSessionColumns + ` FROM build_sessions ORDER BY started_at DESC LIMIT 1`

	row := ds.driver.QueryRow(query)
	session, err := scanBuildSession(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get latest build session: %w", err)
	}

	return session, nil
}

// GetBuildSession retrieves a build session by its ID.
func (ds *SQLDataStore) GetBuildSession(id string) (*models.BuildSession, error) {
	query := `SELECT ` + buildSessionColumns + ` FROM build_sessions WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	session, err := scanBuildSession(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("build session", id)
		}
		return nil, fmt.Errorf("failed to get build session: %w", err)
	}

	return session, nil
}
