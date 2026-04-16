package db

import (
	"database/sql"
	"errors"
	"fmt"

	"devopsmaestro/models"
)

// =============================================================================
// Workspace Operations
// =============================================================================

// CreateWorkspace inserts a new workspace.
// Callers must ensure defaults (nvim config, slug) are set via
// workspace.PrepareDefaults() before calling this method.
func (ds *SQLDataStore) CreateWorkspace(workspace *models.Workspace) error {
	// Default env to empty JSON object if not set
	if !workspace.Env.Valid {
		workspace.Env = sql.NullString{String: "{}", Valid: true}
	}

	query := fmt.Sprintf(`INSERT INTO workspaces (app_id, name, slug, description, image_name, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, nvim_package, git_repo_id, env, build_config, git_credential_mounting, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, workspace.AppID, workspace.Name, workspace.Slug, workspace.Description, workspace.ImageName, workspace.Status, workspace.SSHAgentForwarding, workspace.NvimStructure, workspace.NvimPlugins, workspace.Theme, workspace.TerminalPrompt, workspace.TerminalPlugins, workspace.TerminalPackage, workspace.NvimPackage, workspace.GitRepoID, workspace.Env, workspace.BuildConfig, workspace.GitCredentialMounting)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		workspace.ID = int(id)
	}

	return nil
}

// GetWorkspaceByName retrieves a workspace by app ID and name.
func (ds *SQLDataStore) GetWorkspaceByName(appID int, name string) (*models.Workspace, error) {
	workspace := &models.Workspace{}
	query := `SELECT id, app_id, name, slug, description, image_name, container_id, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, nvim_package, git_repo_id, env, build_config, git_credential_mounting, created_at, updated_at 
		FROM workspaces WHERE app_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, appID, name)
	if err := row.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Slug, &workspace.Description,
		&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.SSHAgentForwarding, &workspace.NvimStructure,
		&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.NvimPackage, &workspace.GitRepoID, &workspace.Env, &workspace.BuildConfig, &workspace.GitCredentialMounting, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("workspace", name)
		}
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return workspace, nil
}

// GetWorkspaceByID retrieves a workspace by its ID.
func (ds *SQLDataStore) GetWorkspaceByID(id int) (*models.Workspace, error) {
	workspace := &models.Workspace{}
	query := `SELECT id, app_id, name, slug, description, image_name, container_id, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, nvim_package, git_repo_id, env, build_config, git_credential_mounting, created_at, updated_at 
		FROM workspaces WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Slug, &workspace.Description,
		&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.SSHAgentForwarding, &workspace.NvimStructure,
		&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.NvimPackage, &workspace.GitRepoID, &workspace.Env, &workspace.BuildConfig, &workspace.GitCredentialMounting, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("workspace", id)
		}
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return workspace, nil
}

// GetWorkspaceBySlug retrieves a workspace by its hierarchical slug.
func (ds *SQLDataStore) GetWorkspaceBySlug(slug string) (*models.Workspace, error) {
	workspace := &models.Workspace{}
	query := `SELECT id, app_id, name, slug, description, image_name, container_id, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, nvim_package, git_repo_id, env, build_config, git_credential_mounting, created_at, updated_at 
		FROM workspaces WHERE slug = ?`

	row := ds.driver.QueryRow(query, slug)
	if err := row.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Slug, &workspace.Description,
		&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.SSHAgentForwarding, &workspace.NvimStructure,
		&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.NvimPackage, &workspace.GitRepoID, &workspace.Env, &workspace.BuildConfig, &workspace.GitCredentialMounting, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, NewErrNotFound("workspace", slug)
		}
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return workspace, nil
}

// UpdateWorkspace updates an existing workspace.
func (ds *SQLDataStore) UpdateWorkspace(workspace *models.Workspace) error {
	query := fmt.Sprintf(`UPDATE workspaces SET name = ?, slug = ?, description = ?, image_name = ?, container_id = ?, 
		status = ?, ssh_agent_forwarding = ?, nvim_structure = ?, nvim_plugins = ?, theme = ?, terminal_prompt = ?, terminal_plugins = ?, terminal_package = ?, nvim_package = ?, git_repo_id = ?, env = ?, build_config = ?, git_credential_mounting = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, workspace.Name, workspace.Slug, workspace.Description, workspace.ImageName,
		workspace.ContainerID, workspace.Status, workspace.SSHAgentForwarding, workspace.NvimStructure, workspace.NvimPlugins, workspace.Theme, workspace.TerminalPrompt, workspace.TerminalPlugins, workspace.TerminalPackage, workspace.NvimPackage, workspace.GitRepoID, workspace.Env, workspace.BuildConfig, workspace.GitCredentialMounting, workspace.ID)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}
	return nil
}

// DeleteWorkspace removes a workspace by ID.
// Also cleans up orphaned credentials scoped to this workspace
// (polymorphic scope_type/scope_id has no FK constraint).
// The entire operation runs in a transaction to ensure data integrity.
func (ds *SQLDataStore) DeleteWorkspace(id int) error {
	tx, err := ds.driver.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback after commit is a no-op

	// Clean up orphaned credentials (polymorphic scope_id has no FK constraint)
	if _, err := tx.Execute(`DELETE FROM credentials WHERE scope_type = 'workspace' AND scope_id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete workspace credentials: %w", err)
	}

	query := `DELETE FROM workspaces WHERE id = ?`
	result, err := tx.Execute(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return NewErrNotFound("workspace", id)
	}

	return tx.Commit()
}

// ListWorkspacesByApp retrieves all workspaces for an app.
func (ds *SQLDataStore) ListWorkspacesByApp(appID int) ([]*models.Workspace, error) {
	query := `SELECT id, app_id, name, slug, description, image_name, container_id, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, nvim_package, git_repo_id, env, build_config, git_credential_mounting, created_at, updated_at 
		FROM workspaces WHERE app_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*models.Workspace
	for rows.Next() {
		workspace := &models.Workspace{}
		if err := rows.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Slug, &workspace.Description,
			&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.SSHAgentForwarding, &workspace.NvimStructure,
			&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.NvimPackage, &workspace.GitRepoID, &workspace.Env, &workspace.BuildConfig, &workspace.GitCredentialMounting, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspaces: %w", err)
	}

	return workspaces, nil
}

// ListAllWorkspaces retrieves all workspaces across all apps.
func (ds *SQLDataStore) ListAllWorkspaces() ([]*models.Workspace, error) {
	query := `SELECT id, app_id, name, slug, description, image_name, container_id, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, nvim_package, git_repo_id, env, build_config, git_credential_mounting, created_at, updated_at 
		FROM workspaces ORDER BY app_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*models.Workspace
	for rows.Next() {
		workspace := &models.Workspace{}
		if err := rows.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Slug, &workspace.Description,
			&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.SSHAgentForwarding, &workspace.NvimStructure,
			&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.NvimPackage, &workspace.GitRepoID, &workspace.Env, &workspace.BuildConfig, &workspace.GitCredentialMounting, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan workspace: %w", err)
		}
		workspaces = append(workspaces, workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspaces: %w", err)
	}

	return workspaces, nil
}

// FindWorkspaces searches for workspaces matching the given filter criteria.
// Returns workspaces with their full hierarchy information (ecosystem, domain, system, app).
// Use this for smart workspace resolution when the user provides partial criteria.
func (ds *SQLDataStore) FindWorkspaces(filter models.WorkspaceFilter) ([]*models.WorkspaceWithHierarchy, error) {
	// Build query with JOINs to get full hierarchy (LEFT JOIN on systems since system is optional)
	query := `SELECT 
		w.id, w.app_id, w.name, w.description, w.image_name, w.container_id, w.status, w.nvim_structure, w.nvim_plugins, w.theme, w.terminal_prompt, w.terminal_plugins, w.terminal_package, w.nvim_package, w.slug, w.ssh_agent_forwarding, w.git_repo_id, w.env, w.build_config, w.git_credential_mounting, w.created_at, w.updated_at,
		a.id, a.domain_id, a.system_id, a.name, a.path, a.description, a.language, a.build_config, a.created_at, a.updated_at,
		s.id, s.ecosystem_id, s.domain_id, s.name, s.description, s.theme, s.nvim_package, s.terminal_package, s.build_args, s.ca_certs, s.created_at, s.updated_at,
		d.id, d.ecosystem_id, d.name, d.description, d.created_at, d.updated_at,
		e.id, e.name, e.description, e.created_at, e.updated_at
	FROM workspaces w
	JOIN apps a ON w.app_id = a.id
	LEFT JOIN systems s ON a.system_id = s.id
	LEFT JOIN domains d ON a.domain_id = d.id
	LEFT JOIN ecosystems e ON d.ecosystem_id = e.id
	WHERE 1=1`

	var args []interface{}

	// Add filter conditions
	if filter.EcosystemName != "" {
		query += " AND e.name = ?"
		args = append(args, filter.EcosystemName)
	}
	if filter.DomainName != "" {
		query += " AND d.name = ?"
		args = append(args, filter.DomainName)
	}
	if filter.SystemName != "" {
		query += " AND s.name = ?"
		args = append(args, filter.SystemName)
	}
	if filter.AppName != "" {
		query += " AND a.name = ?"
		args = append(args, filter.AppName)
	}
	if filter.WorkspaceName != "" {
		query += " AND w.name = ?"
		args = append(args, filter.WorkspaceName)
	}

	query += " ORDER BY e.name, d.name, a.name, w.name"

	rows, err := ds.driver.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to find workspaces: %w", err)
	}
	defer rows.Close()

	var results []*models.WorkspaceWithHierarchy
	for rows.Next() {
		workspace := &models.Workspace{}
		app := &models.App{}

		// System fields are nullable via LEFT JOIN
		var sysID, sysEcoID, sysDomainID sql.NullInt64
		var sysName, sysDesc, sysTheme sql.NullString
		var sysNvimPkg, sysTermPkg, sysBuildArgs, sysCACerts sql.NullString
		var sysCreatedAt, sysUpdatedAt sql.NullTime

		// Domain fields are nullable via LEFT JOIN
		var domID sql.NullInt64
		var domEcoID sql.NullInt64
		var domName, domDesc sql.NullString
		var domCreatedAt, domUpdatedAt sql.NullTime

		// Ecosystem fields are nullable via LEFT JOIN
		var ecoID sql.NullInt64
		var ecoName, ecoDesc sql.NullString
		var ecoCreatedAt, ecoUpdatedAt sql.NullTime

		if err := rows.Scan(
			// Workspace fields
			&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Description,
			&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure,
			&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.NvimPackage, &workspace.Slug, &workspace.SSHAgentForwarding, &workspace.GitRepoID, &workspace.Env, &workspace.BuildConfig, &workspace.GitCredentialMounting, &workspace.CreatedAt, &workspace.UpdatedAt,
			// App fields (now includes system_id)
			&app.ID, &app.DomainID, &app.SystemID, &app.Name, &app.Path, &app.Description,
			&app.Language, &app.BuildConfig, &app.CreatedAt, &app.UpdatedAt,
			// System fields (nullable via LEFT JOIN)
			&sysID, &sysEcoID, &sysDomainID, &sysName, &sysDesc, &sysTheme,
			&sysNvimPkg, &sysTermPkg, &sysBuildArgs, &sysCACerts,
			&sysCreatedAt, &sysUpdatedAt,
			// Domain fields (nullable via LEFT JOIN)
			&domID, &domEcoID, &domName, &domDesc,
			&domCreatedAt, &domUpdatedAt,
			// Ecosystem fields (nullable via LEFT JOIN)
			&ecoID, &ecoName, &ecoDesc,
			&ecoCreatedAt, &ecoUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan workspace with hierarchy: %w", err)
		}

		result := &models.WorkspaceWithHierarchy{
			Workspace: workspace,
			App:       app,
		}

		// Populate system if present
		if sysID.Valid {
			result.System = &models.System{
				ID:              int(sysID.Int64),
				EcosystemID:     sysEcoID,
				DomainID:        sysDomainID,
				Name:            sysName.String,
				Description:     sysDesc,
				Theme:           sysTheme,
				NvimPackage:     sysNvimPkg,
				TerminalPackage: sysTermPkg,
				BuildArgs:       sysBuildArgs,
				CACerts:         sysCACerts,
			}
			if sysCreatedAt.Valid {
				result.System.CreatedAt = sysCreatedAt.Time
			}
			if sysUpdatedAt.Valid {
				result.System.UpdatedAt = sysUpdatedAt.Time
			}
		}

		// Populate domain if present
		if domID.Valid {
			result.Domain = &models.Domain{
				ID:          int(domID.Int64),
				EcosystemID: domEcoID,
				Name:        domName.String,
				Description: domDesc,
			}
			if domCreatedAt.Valid {
				result.Domain.CreatedAt = domCreatedAt.Time
			}
			if domUpdatedAt.Valid {
				result.Domain.UpdatedAt = domUpdatedAt.Time
			}
		}

		// Populate ecosystem if present
		if ecoID.Valid {
			result.Ecosystem = &models.Ecosystem{
				ID:          int(ecoID.Int64),
				Name:        ecoName.String,
				Description: ecoDesc,
			}
			if ecoCreatedAt.Valid {
				result.Ecosystem.CreatedAt = ecoCreatedAt.Time
			}
			if ecoUpdatedAt.Valid {
				result.Ecosystem.UpdatedAt = ecoUpdatedAt.Time
			}
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspaces: %w", err)
	}

	return results, nil
}

// GetWorkspaceSlug returns the slug for a workspace.
func (ds *SQLDataStore) GetWorkspaceSlug(workspaceID int) (string, error) {
	var slug string
	query := `SELECT slug FROM workspaces WHERE id = ?`

	row := ds.driver.QueryRow(query, workspaceID)
	if err := row.Scan(&slug); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", NewErrNotFound("workspace", workspaceID)
		}
		return "", fmt.Errorf("failed to get workspace slug: %w", err)
	}

	return slug, nil
}
