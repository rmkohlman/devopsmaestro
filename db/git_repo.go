package db

import (
	"database/sql"
	"errors"
	"fmt"

	"devopsmaestro/models"
)

// =============================================================================
// GitRepo Operations (v0.20.0 Mirror)
// =============================================================================

// CreateGitRepo inserts a new git repository configuration.
func (ds *SQLDataStore) CreateGitRepo(repo *models.GitRepoDB) error {
	query := fmt.Sprintf(`
		INSERT INTO git_repos (
			name, url, slug, default_ref, auth_type, credential_id,
			auto_sync, sync_interval_minutes, sync_status, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		repo.Name,
		repo.URL,
		repo.Slug,
		repo.DefaultRef,
		repo.AuthType,
		repo.CredentialID,
		repo.AutoSync,
		repo.SyncIntervalMinutes,
		repo.SyncStatus,
	)
	if err != nil {
		return fmt.Errorf("failed to create git repo: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		repo.ID = int(id)
	}

	return nil
}

// GetGitRepoByName retrieves a git repository by its name.
func (ds *SQLDataStore) GetGitRepoByName(name string) (*models.GitRepoDB, error) {
	repo := &models.GitRepoDB{}
	query := `
		SELECT id, name, url, slug, default_ref, auth_type, credential_id,
		       auto_sync, sync_interval_minutes, last_synced_at, sync_status, 
		       sync_error, created_at, updated_at
		FROM git_repos
		WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	err := row.Scan(
		&repo.ID,
		&repo.Name,
		&repo.URL,
		&repo.Slug,
		&repo.DefaultRef,
		&repo.AuthType,
		&repo.CredentialID,
		&repo.AutoSync,
		&repo.SyncIntervalMinutes,
		&repo.LastSyncedAt,
		&repo.SyncStatus,
		&repo.SyncError,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("git repo", name)
		}
		return nil, fmt.Errorf("failed to get git repo by name: %w", err)
	}

	return repo, nil
}

// GetGitRepoByID retrieves a git repository by its ID.
func (ds *SQLDataStore) GetGitRepoByID(id int64) (*models.GitRepoDB, error) {
	repo := &models.GitRepoDB{}
	query := `
		SELECT id, name, url, slug, default_ref, auth_type, credential_id,
		       auto_sync, sync_interval_minutes, last_synced_at, sync_status, 
		       sync_error, created_at, updated_at
		FROM git_repos
		WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	err := row.Scan(
		&repo.ID,
		&repo.Name,
		&repo.URL,
		&repo.Slug,
		&repo.DefaultRef,
		&repo.AuthType,
		&repo.CredentialID,
		&repo.AutoSync,
		&repo.SyncIntervalMinutes,
		&repo.LastSyncedAt,
		&repo.SyncStatus,
		&repo.SyncError,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("git repo", id)
		}
		return nil, fmt.Errorf("failed to get git repo by id: %w", err)
	}

	return repo, nil
}

// GetGitRepoBySlug retrieves a git repository by its slug.
func (ds *SQLDataStore) GetGitRepoBySlug(slug string) (*models.GitRepoDB, error) {
	repo := &models.GitRepoDB{}
	query := `
		SELECT id, name, url, slug, default_ref, auth_type, credential_id,
		       auto_sync, sync_interval_minutes, last_synced_at, sync_status, 
		       sync_error, created_at, updated_at
		FROM git_repos
		WHERE slug = ?`

	row := ds.driver.QueryRow(query, slug)
	err := row.Scan(
		&repo.ID,
		&repo.Name,
		&repo.URL,
		&repo.Slug,
		&repo.DefaultRef,
		&repo.AuthType,
		&repo.CredentialID,
		&repo.AutoSync,
		&repo.SyncIntervalMinutes,
		&repo.LastSyncedAt,
		&repo.SyncStatus,
		&repo.SyncError,
		&repo.CreatedAt,
		&repo.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("git repo", slug)
		}
		return nil, fmt.Errorf("failed to get git repo by slug: %w", err)
	}

	return repo, nil
}

// UpdateGitRepo updates an existing git repository configuration.
func (ds *SQLDataStore) UpdateGitRepo(repo *models.GitRepoDB) error {
	query := fmt.Sprintf(`
		UPDATE git_repos
		SET url = ?, slug = ?, default_ref = ?, auth_type = ?, credential_id = ?,
		    auto_sync = ?, sync_interval_minutes = ?, last_synced_at = ?,
		    sync_status = ?, sync_error = ?, updated_at = %s
		WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		repo.URL,
		repo.Slug,
		repo.DefaultRef,
		repo.AuthType,
		repo.CredentialID,
		repo.AutoSync,
		repo.SyncIntervalMinutes,
		repo.LastSyncedAt,
		repo.SyncStatus,
		repo.SyncError,
		repo.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update git repo: %w", err)
	}

	return nil
}

// DeleteGitRepo removes a git repository by name.
func (ds *SQLDataStore) DeleteGitRepo(name string) error {
	query := `DELETE FROM git_repos WHERE name = ?`

	result, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete git repo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("git repo", name)
	}

	return nil
}

// ListGitRepos retrieves all git repositories.
func (ds *SQLDataStore) ListGitRepos() ([]models.GitRepoDB, error) {
	query := `
		SELECT id, name, url, slug, default_ref, auth_type, credential_id,
		       auto_sync, sync_interval_minutes, last_synced_at, sync_status, 
		       sync_error, created_at, updated_at
		FROM git_repos
		ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list git repos: %w", err)
	}
	defer rows.Close()

	var repos []models.GitRepoDB
	for rows.Next() {
		var repo models.GitRepoDB
		err := rows.Scan(
			&repo.ID,
			&repo.Name,
			&repo.URL,
			&repo.Slug,
			&repo.DefaultRef,
			&repo.AuthType,
			&repo.CredentialID,
			&repo.AutoSync,
			&repo.SyncIntervalMinutes,
			&repo.LastSyncedAt,
			&repo.SyncStatus,
			&repo.SyncError,
			&repo.CreatedAt,
			&repo.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan git repo: %w", err)
		}
		repos = append(repos, repo)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating git repos: %w", err)
	}

	return repos, nil
}

// ListAppsByGitRepoID retrieves all apps linked to a git repository.
// Returns an empty slice (not nil) when no apps are linked.
func (ds *SQLDataStore) ListAppsByGitRepoID(gitRepoID int64) ([]*models.App, error) {
	query := `
		SELECT id, domain_id, name, path, description, theme, nvim_package,
		       terminal_package, language, build_config, git_repo_id,
		       created_at, updated_at
		FROM apps
		WHERE git_repo_id = ?
		ORDER BY name`

	rows, err := ds.driver.Query(query, gitRepoID)
	if err != nil {
		return nil, fmt.Errorf("failed to list apps by git repo: %w", err)
	}
	defer rows.Close()

	apps := make([]*models.App, 0)
	for rows.Next() {
		var app models.App
		err := rows.Scan(
			&app.ID,
			&app.DomainID,
			&app.Name,
			&app.Path,
			&app.Description,
			&app.Theme,
			&app.NvimPackage,
			&app.TerminalPackage,
			&app.Language,
			&app.BuildConfig,
			&app.GitRepoID,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan app: %w", err)
		}
		apps = append(apps, &app)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating apps: %w", err)
	}

	return apps, nil
}

// ListWorkspacesByGitRepoID retrieves all workspaces linked to a git repository.
// Returns an empty slice (not nil) when no workspaces are linked.
func (ds *SQLDataStore) ListWorkspacesByGitRepoID(gitRepoID int64) ([]*models.Workspace, error) {
	query := `
		SELECT id, app_id, name, slug, description, image_name, container_id,
		       status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme,
		       terminal_prompt, terminal_plugins, terminal_package, nvim_package,
		       git_repo_id, env, build_config, created_at, updated_at
		FROM workspaces
		WHERE git_repo_id = ?
		ORDER BY name`

	rows, err := ds.driver.Query(query, gitRepoID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces by git repo: %w", err)
	}
	defer rows.Close()

	workspaces := make([]*models.Workspace, 0)
	for rows.Next() {
		var ws models.Workspace
		err := rows.Scan(
			&ws.ID, &ws.AppID, &ws.Name, &ws.Slug, &ws.Description,
			&ws.ImageName, &ws.ContainerID, &ws.Status, &ws.SSHAgentForwarding,
			&ws.NvimStructure, &ws.NvimPlugins, &ws.Theme,
			&ws.TerminalPrompt, &ws.TerminalPlugins, &ws.TerminalPackage,
			&ws.NvimPackage, &ws.GitRepoID, &ws.Env, &ws.BuildConfig,
			&ws.CreatedAt, &ws.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, &ws)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating workspaces: %w", err)
	}

	return workspaces, nil
}
