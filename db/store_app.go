package db

import (
	"database/sql"
	"errors"
	"fmt"

	"devopsmaestro/models"
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
		if errors.Is(err, sql.ErrNoRows) {
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
		if errors.Is(err, sql.ErrNoRows) {
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
		if errors.Is(err, sql.ErrNoRows) {
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
// Also cleans up orphaned credentials scoped to this app and its child workspaces
// (polymorphic scope_type/scope_id has no FK constraint).
// The entire operation runs in a transaction to ensure data integrity.
func (ds *SQLDataStore) DeleteApp(id int) error {
	tx, err := ds.driver.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	// Clean up credentials scoped to workspaces under this app
	if _, err := tx.Execute(`DELETE FROM credentials WHERE scope_type = 'workspace' AND scope_id IN (SELECT id FROM workspaces WHERE app_id = ?)`, id); err != nil {
		return fmt.Errorf("failed to delete workspace credentials for app: %w", err)
	}
	// Clean up credentials scoped to this app
	if _, err := tx.Execute(`DELETE FROM credentials WHERE scope_type = 'app' AND scope_id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete app credentials: %w", err)
	}

	query := `DELETE FROM apps WHERE id = ?`
	result, err := tx.Execute(query, id)
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

	return tx.Commit()
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

// FindAppsByName retrieves all apps with the given name across all domains,
// including their full hierarchy (domain and ecosystem).
// Returns an empty slice (not an error) if no apps match.
func (ds *SQLDataStore) FindAppsByName(name string) ([]*models.AppWithHierarchy, error) {
	query := `SELECT 
		a.id, a.domain_id, a.name, a.path, a.description, a.theme, a.language, a.build_config, a.git_repo_id, a.created_at, a.updated_at,
		d.id, d.ecosystem_id, d.name, d.description, d.theme, d.build_args, d.ca_certs, d.created_at, d.updated_at,
		e.id, e.name, e.description, e.theme, e.build_args, e.ca_certs, e.created_at, e.updated_at
	FROM apps a
	JOIN domains d ON a.domain_id = d.id
	JOIN ecosystems e ON d.ecosystem_id = e.id
	WHERE a.name = ?
	ORDER BY e.name, d.name`

	rows, err := ds.driver.Query(query, name)
	if err != nil {
		return nil, fmt.Errorf("failed to find apps by name: %w", err)
	}
	defer rows.Close()

	var results []*models.AppWithHierarchy
	for rows.Next() {
		app := &models.App{}
		domain := &models.Domain{}
		ecosystem := &models.Ecosystem{}

		if err := rows.Scan(
			// App fields
			&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.GitRepoID, &app.CreatedAt, &app.UpdatedAt,
			// Domain fields
			&domain.ID, &domain.EcosystemID, &domain.Name, &domain.Description, &domain.Theme, &domain.BuildArgs, &domain.CACerts, &domain.CreatedAt, &domain.UpdatedAt,
			// Ecosystem fields
			&ecosystem.ID, &ecosystem.Name, &ecosystem.Description, &ecosystem.Theme, &ecosystem.BuildArgs, &ecosystem.CACerts, &ecosystem.CreatedAt, &ecosystem.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan app with hierarchy: %w", err)
		}

		results = append(results, &models.AppWithHierarchy{
			App:       app,
			Domain:    domain,
			Ecosystem: ecosystem,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over apps: %w", err)
	}

	return results, nil
}
