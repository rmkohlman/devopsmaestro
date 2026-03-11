package db

import (
	"database/sql"
	"devopsmaestro/models"
	"fmt"
)

// =============================================================================
// App Operations
// =============================================================================

// CreateApp inserts a new app into the database.
func (ds *SQLDataStore) CreateApp(app *models.App) error {
	query := fmt.Sprintf(`INSERT INTO apps (domain_id, name, path, description, theme, language, build_config, git_repo_id, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, app.DomainID, app.Name, app.Path, app.Description, app.Theme, app.Language, app.BuildConfig, app.GitRepoID)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err == nil {
		app.ID = int(id)
	}

	return nil
}

// GetAppByName retrieves an app by domain ID and name.
func (ds *SQLDataStore) GetAppByName(domainID int, name string) (*models.App, error) {
	app := &models.App{}
	query := `SELECT id, domain_id, name, path, description, theme, language, build_config, git_repo_id, created_at, updated_at FROM apps WHERE domain_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, domainID, name)
	if err := row.Scan(&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.GitRepoID, &app.CreatedAt, &app.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("app", name)
		}
		return nil, fmt.Errorf("failed to scan app: %w", err)
	}

	return app, nil
}

// GetAppByNameGlobal retrieves an app by name across all domains.
// Returns the first match if multiple apps have the same name in different domains.
func (ds *SQLDataStore) GetAppByNameGlobal(name string) (*models.App, error) {
	app := &models.App{}
	query := `SELECT id, domain_id, name, path, description, theme, language, build_config, git_repo_id, created_at, updated_at FROM apps WHERE name = ? LIMIT 1`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.GitRepoID, &app.CreatedAt, &app.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("app", name)
		}
		return nil, fmt.Errorf("failed to scan app: %w", err)
	}

	return app, nil
}

// GetAppByID retrieves an app by its ID.
func (ds *SQLDataStore) GetAppByID(id int) (*models.App, error) {
	app := &models.App{}
	query := `SELECT id, domain_id, name, path, description, theme, language, build_config, git_repo_id, created_at, updated_at FROM apps WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.GitRepoID, &app.CreatedAt, &app.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, NewErrNotFound("app", id)
		}
		return nil, fmt.Errorf("failed to scan app: %w", err)
	}

	return app, nil
}

// UpdateApp updates an existing app.
func (ds *SQLDataStore) UpdateApp(app *models.App) error {
	query := fmt.Sprintf(`UPDATE apps SET domain_id = ?, name = ?, path = ?, description = ?, theme = ?, language = ?, build_config = ?, git_repo_id = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, app.DomainID, app.Name, app.Path, app.Description, app.Theme, app.Language, app.BuildConfig, app.GitRepoID, app.ID)
	if err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}
	return nil
}

// DeleteApp removes an app by ID.
func (ds *SQLDataStore) DeleteApp(id int) error {
	query := `DELETE FROM apps WHERE id = ?`
	result, err := ds.driver.Execute(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("app", id)
	}
	return nil
}

// ListAppsByDomain retrieves all apps for a domain.
func (ds *SQLDataStore) ListAppsByDomain(domainID int) ([]*models.App, error) {
	query := `SELECT id, domain_id, name, path, description, theme, language, build_config, git_repo_id, created_at, updated_at FROM apps WHERE domain_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}
	defer rows.Close()

	var apps []*models.App
	for rows.Next() {
		app := &models.App{}
		if err := rows.Scan(&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.GitRepoID, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan app: %w", err)
		}
		apps = append(apps, app)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over apps: %w", err)
	}

	return apps, nil
}

// ListAllApps retrieves all apps across all domains.
func (ds *SQLDataStore) ListAllApps() ([]*models.App, error) {
	query := `SELECT id, domain_id, name, path, description, theme, language, build_config, git_repo_id, created_at, updated_at FROM apps ORDER BY domain_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all apps: %w", err)
	}
	defer rows.Close()

	var apps []*models.App
	for rows.Next() {
		app := &models.App{}
		if err := rows.Scan(&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.GitRepoID, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan app: %w", err)
		}
		apps = append(apps, app)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over apps: %w", err)
	}

	return apps, nil
}
