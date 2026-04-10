package db

import (
	"fmt"
	"time"

	"devopsmaestro/models"
)

// =============================================================================
// Build Session — List, Delete, Workspace Operations
// =============================================================================

// GetBuildSessions retrieves the most recent build sessions up to limit.
func (ds *SQLDataStore) GetBuildSessions(limit int) ([]*models.BuildSession, error) {
	query := `SELECT ` + buildSessionColumns + ` FROM build_sessions ORDER BY started_at DESC LIMIT ?`

	rows, err := ds.driver.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list build sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*models.BuildSession
	for rows.Next() {
		session := &models.BuildSession{}
		if err := rows.Scan(
			&session.ID,
			&session.StartedAt,
			&session.CompletedAt,
			&session.Status,
			&session.TotalWorkspaces,
			&session.Succeeded,
			&session.Failed,
			&session.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan build session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating build sessions: %w", err)
	}

	return sessions, nil
}

// DeleteBuildSessionsOlderThan removes build sessions older than the cutoff time.
// Cascade deletes also remove associated build_session_workspaces rows.
func (ds *SQLDataStore) DeleteBuildSessionsOlderThan(cutoff time.Time) (int64, error) {
	query := `DELETE FROM build_sessions WHERE started_at < ?`

	result, err := ds.driver.Execute(query, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old build sessions: %w", err)
	}

	count, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return count, nil
}

// CreateBuildSessionWorkspace inserts a new workspace entry for a build session.
func (ds *SQLDataStore) CreateBuildSessionWorkspace(bsw *models.BuildSessionWorkspace) error {
	query := `INSERT INTO build_session_workspaces 
		(session_id, workspace_id, status, started_at, completed_at, duration_seconds, image_tag, error_message) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := ds.driver.Execute(query,
		bsw.SessionID,
		bsw.WorkspaceID,
		bsw.Status,
		bsw.StartedAt,
		bsw.CompletedAt,
		bsw.DurationSeconds,
		bsw.ImageTag,
		bsw.ErrorMessage,
	)
	if err != nil {
		return fmt.Errorf("failed to create build session workspace: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		bsw.ID = int(id)
	}

	return nil
}

// UpdateBuildSessionWorkspace updates a workspace entry within a build session.
func (ds *SQLDataStore) UpdateBuildSessionWorkspace(bsw *models.BuildSessionWorkspace) error {
	query := `UPDATE build_session_workspaces 
		SET status = ?, started_at = ?, completed_at = ?, duration_seconds = ?, image_tag = ?, error_message = ? 
		WHERE id = ?`

	result, err := ds.driver.Execute(query,
		bsw.Status,
		bsw.StartedAt,
		bsw.CompletedAt,
		bsw.DurationSeconds,
		bsw.ImageTag,
		bsw.ErrorMessage,
		bsw.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update build session workspace: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return NewErrNotFound("build session workspace", bsw.ID)
	}

	return nil
}

// GetBuildSessionWorkspaces retrieves all workspace entries for a build session.
func (ds *SQLDataStore) GetBuildSessionWorkspaces(sessionID string) ([]*models.BuildSessionWorkspace, error) {
	query := `SELECT id, session_id, workspace_id, status, started_at, completed_at, duration_seconds, image_tag, error_message 
		FROM build_session_workspaces WHERE session_id = ? ORDER BY id`

	rows, err := ds.driver.Query(query, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list build session workspaces: %w", err)
	}
	defer rows.Close()

	var results []*models.BuildSessionWorkspace
	for rows.Next() {
		bsw := &models.BuildSessionWorkspace{}
		if err := rows.Scan(
			&bsw.ID,
			&bsw.SessionID,
			&bsw.WorkspaceID,
			&bsw.Status,
			&bsw.StartedAt,
			&bsw.CompletedAt,
			&bsw.DurationSeconds,
			&bsw.ImageTag,
			&bsw.ErrorMessage,
		); err != nil {
			return nil, fmt.Errorf("failed to scan build session workspace: %w", err)
		}
		results = append(results, bsw)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating build session workspaces: %w", err)
	}

	return results, nil
}

// GetBuildSessionStats returns the succeeded and failed counts for a build session.
func (ds *SQLDataStore) GetBuildSessionStats(sessionID string) (succeeded int, failed int, err error) {
	query := `SELECT 
		COALESCE(SUM(CASE WHEN status = 'succeeded' THEN 1 ELSE 0 END), 0),
		COALESCE(SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END), 0)
		FROM build_session_workspaces WHERE session_id = ?`

	row := ds.driver.QueryRow(query, sessionID)
	if err := row.Scan(&succeeded, &failed); err != nil {
		return 0, 0, fmt.Errorf("failed to get build session stats: %w", err)
	}

	return succeeded, failed, nil
}

// UpdateWorkspaceImage updates the image_name field of a workspace by ID.
func (ds *SQLDataStore) UpdateWorkspaceImage(workspaceID int, imageTag string) error {
	query := fmt.Sprintf(`UPDATE workspaces SET image_name = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, imageTag, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to update workspace image: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return NewErrNotFound("workspace", workspaceID)
	}

	return nil
}
