package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// Context Operations
// =============================================================================

// GetContext retrieves the current context.
func (ds *SQLDataStore) GetContext() (*models.Context, error) {
	context := &models.Context{}
	query := `SELECT id, active_ecosystem_id, active_domain_id, active_app_id, active_workspace_id, updated_at FROM context WHERE id = 1`

	row := ds.driver.QueryRow(query)
	if err := row.Scan(&context.ID, &context.ActiveEcosystemID, &context.ActiveDomainID, &context.ActiveAppID, &context.ActiveWorkspaceID, &context.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("context not found")
		}
		return nil, fmt.Errorf("failed to scan context: %w", err)
	}

	return context, nil
}

// SetActiveEcosystem sets the active ecosystem in the context.
func (ds *SQLDataStore) SetActiveEcosystem(ecosystemID *int) error {
	query := fmt.Sprintf(`UPDATE context SET active_ecosystem_id = ?, updated_at = %s WHERE id = 1`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, ecosystemID)
	if err != nil {
		return fmt.Errorf("failed to set active ecosystem: %w", err)
	}
	return nil
}

// SetActiveDomain sets the active domain in the context.
func (ds *SQLDataStore) SetActiveDomain(domainID *int) error {
	query := fmt.Sprintf(`UPDATE context SET active_domain_id = ?, updated_at = %s WHERE id = 1`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, domainID)
	if err != nil {
		return fmt.Errorf("failed to set active domain: %w", err)
	}
	return nil
}

// SetActiveApp sets the active app in the context.
func (ds *SQLDataStore) SetActiveApp(appID *int) error {
	query := fmt.Sprintf(`UPDATE context SET active_app_id = ?, updated_at = %s WHERE id = 1`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, appID)
	if err != nil {
		return fmt.Errorf("failed to set active app: %w", err)
	}
	return nil
}

// SetActiveWorkspace sets the active workspace in the context.
func (ds *SQLDataStore) SetActiveWorkspace(workspaceID *int) error {
	query := fmt.Sprintf(`UPDATE context SET active_workspace_id = ?, updated_at = %s WHERE id = 1`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to set active workspace: %w", err)
	}
	return nil
}
