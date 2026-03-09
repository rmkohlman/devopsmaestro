package db

import (
	"database/sql"
	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// =============================================================================
// Workspace Operations
// =============================================================================

// CreateWorkspace inserts a new workspace.
func (ds *SQLDataStore) CreateWorkspace(workspace *models.Workspace) error {
	// Apply default nvim config if not specified
	if !workspace.NvimStructure.Valid || workspace.NvimStructure.String == "" {
		defaultConfig := nvimops.DefaultNvimConfig()
		workspace.NvimStructure = sql.NullString{String: defaultConfig.Structure, Valid: true}
	}

	// Generate slug if not provided
	if workspace.Slug == "" {
		// Get hierarchy to generate slug
		app, err := ds.GetAppByID(workspace.AppID)
		if err != nil {
			return fmt.Errorf("failed to get app for slug generation: %w", err)
		}
		domain, err := ds.GetDomainByID(app.DomainID)
		if err != nil {
			return fmt.Errorf("failed to get domain for slug generation: %w", err)
		}
		ecosystem, err := ds.GetEcosystemByID(domain.EcosystemID)
		if err != nil {
			return fmt.Errorf("failed to get ecosystem for slug generation: %w", err)
		}
		workspace.Slug = ds.GenerateWorkspaceSlug(ecosystem.Name, domain.Name, app.Name, workspace.Name)
	}

	query := fmt.Sprintf(`INSERT INTO workspaces (app_id, name, slug, description, image_name, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, git_repo_id, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, workspace.AppID, workspace.Name, workspace.Slug, workspace.Description, workspace.ImageName, workspace.Status, workspace.SSHAgentForwarding, workspace.NvimStructure, workspace.NvimPlugins, workspace.Theme, workspace.TerminalPrompt, workspace.TerminalPlugins, workspace.TerminalPackage, workspace.GitRepoID)
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
	query := `SELECT id, app_id, name, slug, description, image_name, container_id, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, git_repo_id, created_at, updated_at 
		FROM workspaces WHERE app_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, appID, name)
	if err := row.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Slug, &workspace.Description,
		&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.SSHAgentForwarding, &workspace.NvimStructure,
		&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.GitRepoID, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workspace not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return workspace, nil
}

// GetWorkspaceByID retrieves a workspace by its ID.
func (ds *SQLDataStore) GetWorkspaceByID(id int) (*models.Workspace, error) {
	workspace := &models.Workspace{}
	query := `SELECT id, app_id, name, slug, description, image_name, container_id, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, git_repo_id, created_at, updated_at 
		FROM workspaces WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Slug, &workspace.Description,
		&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.SSHAgentForwarding, &workspace.NvimStructure,
		&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.GitRepoID, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workspace not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return workspace, nil
}

// GetWorkspaceBySlug retrieves a workspace by its hierarchical slug.
func (ds *SQLDataStore) GetWorkspaceBySlug(slug string) (*models.Workspace, error) {
	workspace := &models.Workspace{}
	query := `SELECT id, app_id, name, slug, description, image_name, container_id, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, git_repo_id, created_at, updated_at 
		FROM workspaces WHERE slug = ?`

	row := ds.driver.QueryRow(query, slug)
	if err := row.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Slug, &workspace.Description,
		&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.SSHAgentForwarding, &workspace.NvimStructure,
		&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.GitRepoID, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workspace not found: %s", slug)
		}
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return workspace, nil
}

// UpdateWorkspace updates an existing workspace.
func (ds *SQLDataStore) UpdateWorkspace(workspace *models.Workspace) error {
	query := fmt.Sprintf(`UPDATE workspaces SET name = ?, slug = ?, description = ?, image_name = ?, container_id = ?, 
		status = ?, ssh_agent_forwarding = ?, nvim_structure = ?, nvim_plugins = ?, theme = ?, terminal_prompt = ?, terminal_plugins = ?, terminal_package = ?, git_repo_id = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, workspace.Name, workspace.Slug, workspace.Description, workspace.ImageName,
		workspace.ContainerID, workspace.Status, workspace.SSHAgentForwarding, workspace.NvimStructure, workspace.NvimPlugins, workspace.Theme, workspace.TerminalPrompt, workspace.TerminalPlugins, workspace.TerminalPackage, workspace.GitRepoID, workspace.ID)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}
	return nil
}

// DeleteWorkspace removes a workspace by ID.
func (ds *SQLDataStore) DeleteWorkspace(id int) error {
	query := `DELETE FROM workspaces WHERE id = ?`
	result, err := ds.driver.Execute(query, id)
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
	return nil
}

// ListWorkspacesByApp retrieves all workspaces for an app.
func (ds *SQLDataStore) ListWorkspacesByApp(appID int) ([]*models.Workspace, error) {
	query := `SELECT id, app_id, name, slug, description, image_name, container_id, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, git_repo_id, created_at, updated_at 
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
			&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.GitRepoID, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
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
	query := `SELECT id, app_id, name, slug, description, image_name, container_id, status, ssh_agent_forwarding, nvim_structure, nvim_plugins, theme, terminal_prompt, terminal_plugins, terminal_package, git_repo_id, created_at, updated_at 
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
			&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.GitRepoID, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
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
// Returns workspaces with their full hierarchy information (ecosystem, domain, app).
// Use this for smart workspace resolution when the user provides partial criteria.
func (ds *SQLDataStore) FindWorkspaces(filter models.WorkspaceFilter) ([]*models.WorkspaceWithHierarchy, error) {
	// Build query with JOINs to get full hierarchy
	query := `SELECT 
		w.id, w.app_id, w.name, w.description, w.image_name, w.container_id, w.status, w.nvim_structure, w.nvim_plugins, w.theme, w.terminal_prompt, w.terminal_plugins, w.terminal_package, w.slug, w.ssh_agent_forwarding, w.git_repo_id, w.created_at, w.updated_at,
		a.id, a.domain_id, a.name, a.path, a.description, a.language, a.build_config, a.created_at, a.updated_at,
		d.id, d.ecosystem_id, d.name, d.description, d.created_at, d.updated_at,
		e.id, e.name, e.description, e.created_at, e.updated_at
	FROM workspaces w
	JOIN apps a ON w.app_id = a.id
	JOIN domains d ON a.domain_id = d.id
	JOIN ecosystems e ON d.ecosystem_id = e.id
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
		domain := &models.Domain{}
		ecosystem := &models.Ecosystem{}

		if err := rows.Scan(
			// Workspace fields
			&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Description,
			&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure,
			&workspace.NvimPlugins, &workspace.Theme, &workspace.TerminalPrompt, &workspace.TerminalPlugins, &workspace.TerminalPackage, &workspace.Slug, &workspace.SSHAgentForwarding, &workspace.GitRepoID, &workspace.CreatedAt, &workspace.UpdatedAt,
			// App fields
			&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description,
			&app.Language, &app.BuildConfig, &app.CreatedAt, &app.UpdatedAt,
			// Domain fields
			&domain.ID, &domain.EcosystemID, &domain.Name, &domain.Description,
			&domain.CreatedAt, &domain.UpdatedAt,
			// Ecosystem fields
			&ecosystem.ID, &ecosystem.Name, &ecosystem.Description,
			&ecosystem.CreatedAt, &ecosystem.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan workspace with hierarchy: %w", err)
		}

		results = append(results, &models.WorkspaceWithHierarchy{
			Workspace: workspace,
			App:       app,
			Domain:    domain,
			Ecosystem: ecosystem,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspaces: %w", err)
	}

	return results, nil
}

// GetWorkspacePath returns the filesystem path for a workspace.
// Returns: ~/.devopsmaestro/workspaces/{slug}/
func (ds *SQLDataStore) GetWorkspacePath(workspaceID int) (string, error) {
	slug, err := ds.GetWorkspaceSlug(workspaceID)
	if err != nil {
		return "", err
	}

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return fmt.Sprintf("%s/.devopsmaestro/workspaces/%s/", homeDir, slug), nil
}

// GetWorkspaceRepoPath returns the path to the workspace's git clone directory.
// Returns: ~/.devopsmaestro/workspaces/{slug}/repo/
func (ds *SQLDataStore) GetWorkspaceRepoPath(workspaceID int) (string, error) {
	basePath, err := ds.GetWorkspacePath(workspaceID)
	if err != nil {
		return "", err
	}
	return filepath.Join(basePath, "repo"), nil
}

// GetWorkspaceSlug returns the slug for a workspace.
func (ds *SQLDataStore) GetWorkspaceSlug(workspaceID int) (string, error) {
	var slug string
	query := `SELECT slug FROM workspaces WHERE id = ?`

	row := ds.driver.QueryRow(query, workspaceID)
	if err := row.Scan(&slug); err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("workspace not found: %d", workspaceID)
		}
		return "", fmt.Errorf("failed to get workspace slug: %w", err)
	}

	return slug, nil
}

// GenerateWorkspaceSlug creates a slug from hierarchy names.
// Format: {ecosystem}-{domain}-{app}-{workspace}
// Names are sanitized: lowercased, spaces/underscores converted to hyphens
func (ds *SQLDataStore) GenerateWorkspaceSlug(ecosystemName, domainName, appName, workspaceName string) string {
	sanitize := func(s string) string {
		// Convert to lowercase
		s = strings.ToLower(s)
		// Replace spaces and underscores with hyphens
		s = strings.ReplaceAll(s, " ", "-")
		s = strings.ReplaceAll(s, "_", "-")
		return s
	}

	return fmt.Sprintf("%s-%s-%s-%s",
		sanitize(ecosystemName),
		sanitize(domainName),
		sanitize(appName),
		sanitize(workspaceName),
	)
}
