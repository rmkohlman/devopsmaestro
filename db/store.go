package db

import (
	"database/sql"
	"devopsmaestro/models"
	"devopsmaestro/pkg/nvimops"
	"fmt"
)

// SQLDataStore is a concrete implementation of the DataStore interface.
// It uses the Driver interface for database operations and QueryBuilder
// for dialect-specific SQL generation.
type SQLDataStore struct {
	driver       Driver
	queryBuilder QueryBuilder
}

// NewSQLDataStore creates a new SQLDataStore with the given driver.
// If queryBuilder is nil, the appropriate builder is selected based on driver type.
func NewSQLDataStore(driver Driver, queryBuilder QueryBuilder) *SQLDataStore {
	if queryBuilder == nil {
		queryBuilder = QueryBuilderFor(driver.Type())
	}
	return &SQLDataStore{
		driver:       driver,
		queryBuilder: queryBuilder,
	}
}

// NewDataStore creates a DataStore from configuration.
// This is the recommended way to create a DataStore.
func NewDataStore(cfg DataStoreConfig) (DataStore, error) {
	if cfg.Driver == nil {
		return nil, fmt.Errorf("driver is required")
	}
	return NewSQLDataStore(cfg.Driver, cfg.QueryBuilder), nil
}

// Driver returns the underlying database driver.
func (ds *SQLDataStore) Driver() Driver {
	return ds.driver
}

// Close releases any resources held by the DataStore.
func (ds *SQLDataStore) Close() error {
	return ds.driver.Close()
}

// Ping verifies the database connection is alive.
func (ds *SQLDataStore) Ping() error {
	return ds.driver.Ping()
}

// =============================================================================
// Ecosystem Operations
// =============================================================================

// CreateEcosystem inserts a new ecosystem into the database.
func (ds *SQLDataStore) CreateEcosystem(ecosystem *models.Ecosystem) error {
	query := fmt.Sprintf(`INSERT INTO ecosystems (name, description, theme, created_at, updated_at) 
		VALUES (?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, ecosystem.Name, ecosystem.Description, ecosystem.Theme)
	if err != nil {
		return fmt.Errorf("failed to create ecosystem: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		ecosystem.ID = int(id)
	}

	return nil
}

// GetEcosystemByName retrieves an ecosystem by its name.
func (ds *SQLDataStore) GetEcosystemByName(name string) (*models.Ecosystem, error) {
	ecosystem := &models.Ecosystem{}
	query := `SELECT id, name, description, theme, created_at, updated_at FROM ecosystems WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&ecosystem.ID, &ecosystem.Name, &ecosystem.Description, &ecosystem.Theme, &ecosystem.CreatedAt, &ecosystem.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ecosystem not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan ecosystem: %w", err)
	}

	return ecosystem, nil
}

// GetEcosystemByID retrieves an ecosystem by its ID.
func (ds *SQLDataStore) GetEcosystemByID(id int) (*models.Ecosystem, error) {
	ecosystem := &models.Ecosystem{}
	query := `SELECT id, name, description, theme, created_at, updated_at FROM ecosystems WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&ecosystem.ID, &ecosystem.Name, &ecosystem.Description, &ecosystem.Theme, &ecosystem.CreatedAt, &ecosystem.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("ecosystem not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan ecosystem: %w", err)
	}

	return ecosystem, nil
}

// UpdateEcosystem updates an existing ecosystem.
func (ds *SQLDataStore) UpdateEcosystem(ecosystem *models.Ecosystem) error {
	query := fmt.Sprintf(`UPDATE ecosystems SET name = ?, description = ?, theme = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, ecosystem.Name, ecosystem.Description, ecosystem.Theme, ecosystem.ID)
	if err != nil {
		return fmt.Errorf("failed to update ecosystem: %w", err)
	}
	return nil
}

// DeleteEcosystem removes an ecosystem by name.
func (ds *SQLDataStore) DeleteEcosystem(name string) error {
	query := `DELETE FROM ecosystems WHERE name = ?`
	_, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete ecosystem: %w", err)
	}
	return nil
}

// ListEcosystems retrieves all ecosystems.
func (ds *SQLDataStore) ListEcosystems() ([]*models.Ecosystem, error) {
	query := `SELECT id, name, description, theme, created_at, updated_at FROM ecosystems ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list ecosystems: %w", err)
	}
	defer rows.Close()

	var ecosystems []*models.Ecosystem
	for rows.Next() {
		ecosystem := &models.Ecosystem{}
		if err := rows.Scan(&ecosystem.ID, &ecosystem.Name, &ecosystem.Description, &ecosystem.Theme, &ecosystem.CreatedAt, &ecosystem.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan ecosystem: %w", err)
		}
		ecosystems = append(ecosystems, ecosystem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over ecosystems: %w", err)
	}

	return ecosystems, nil
}

// =============================================================================
// Domain Operations
// =============================================================================

// CreateDomain inserts a new domain into the database.
func (ds *SQLDataStore) CreateDomain(domain *models.Domain) error {
	query := fmt.Sprintf(`INSERT INTO domains (ecosystem_id, name, description, theme, created_at, updated_at) 
		VALUES (?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, domain.EcosystemID, domain.Name, domain.Description, domain.Theme)
	if err != nil {
		return fmt.Errorf("failed to create domain: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		domain.ID = int(id)
	}

	return nil
}

// GetDomainByName retrieves a domain by ecosystem ID and name.
func (ds *SQLDataStore) GetDomainByName(ecosystemID int, name string) (*models.Domain, error) {
	domain := &models.Domain{}
	query := `SELECT id, ecosystem_id, name, description, theme, created_at, updated_at FROM domains WHERE ecosystem_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, ecosystemID, name)
	if err := row.Scan(&domain.ID, &domain.EcosystemID, &domain.Name, &domain.Description, &domain.Theme, &domain.CreatedAt, &domain.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("domain not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan domain: %w", err)
	}

	return domain, nil
}

// GetDomainByID retrieves a domain by its ID.
func (ds *SQLDataStore) GetDomainByID(id int) (*models.Domain, error) {
	domain := &models.Domain{}
	query := `SELECT id, ecosystem_id, name, description, theme, created_at, updated_at FROM domains WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&domain.ID, &domain.EcosystemID, &domain.Name, &domain.Description, &domain.Theme, &domain.CreatedAt, &domain.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("domain not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan domain: %w", err)
	}

	return domain, nil
}

// UpdateDomain updates an existing domain.
func (ds *SQLDataStore) UpdateDomain(domain *models.Domain) error {
	query := fmt.Sprintf(`UPDATE domains SET ecosystem_id = ?, name = ?, description = ?, theme = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, domain.EcosystemID, domain.Name, domain.Description, domain.Theme, domain.ID)
	if err != nil {
		return fmt.Errorf("failed to update domain: %w", err)
	}
	return nil
}

// DeleteDomain removes a domain by ID.
func (ds *SQLDataStore) DeleteDomain(id int) error {
	query := `DELETE FROM domains WHERE id = ?`
	_, err := ds.driver.Execute(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete domain: %w", err)
	}
	return nil
}

// ListDomainsByEcosystem retrieves all domains for an ecosystem.
func (ds *SQLDataStore) ListDomainsByEcosystem(ecosystemID int) ([]*models.Domain, error) {
	query := `SELECT id, ecosystem_id, name, description, theme, created_at, updated_at FROM domains WHERE ecosystem_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, ecosystemID)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}
	defer rows.Close()

	var domains []*models.Domain
	for rows.Next() {
		domain := &models.Domain{}
		if err := rows.Scan(&domain.ID, &domain.EcosystemID, &domain.Name, &domain.Description, &domain.Theme, &domain.CreatedAt, &domain.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan domain: %w", err)
		}
		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over domains: %w", err)
	}

	return domains, nil
}

// ListAllDomains retrieves all domains across all ecosystems.
func (ds *SQLDataStore) ListAllDomains() ([]*models.Domain, error) {
	query := `SELECT id, ecosystem_id, name, description, theme, created_at, updated_at FROM domains ORDER BY ecosystem_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all domains: %w", err)
	}
	defer rows.Close()

	var domains []*models.Domain
	for rows.Next() {
		domain := &models.Domain{}
		if err := rows.Scan(&domain.ID, &domain.EcosystemID, &domain.Name, &domain.Description, &domain.Theme, &domain.CreatedAt, &domain.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan domain: %w", err)
		}
		domains = append(domains, domain)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over domains: %w", err)
	}

	return domains, nil
}

// =============================================================================
// App Operations
// =============================================================================

// CreateApp inserts a new app into the database.
func (ds *SQLDataStore) CreateApp(app *models.App) error {
	query := fmt.Sprintf(`INSERT INTO apps (domain_id, name, path, description, theme, language, build_config, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, app.DomainID, app.Name, app.Path, app.Description, app.Theme, app.Language, app.BuildConfig)
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
	query := `SELECT id, domain_id, name, path, description, theme, language, build_config, created_at, updated_at FROM apps WHERE domain_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, domainID, name)
	if err := row.Scan(&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.CreatedAt, &app.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("app not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan app: %w", err)
	}

	return app, nil
}

// GetAppByNameGlobal retrieves an app by name across all domains.
// Returns the first match if multiple apps have the same name in different domains.
func (ds *SQLDataStore) GetAppByNameGlobal(name string) (*models.App, error) {
	app := &models.App{}
	query := `SELECT id, domain_id, name, path, description, theme, language, build_config, created_at, updated_at FROM apps WHERE name = ? LIMIT 1`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.CreatedAt, &app.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("app not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan app: %w", err)
	}

	return app, nil
}

// GetAppByID retrieves an app by its ID.
func (ds *SQLDataStore) GetAppByID(id int) (*models.App, error) {
	app := &models.App{}
	query := `SELECT id, domain_id, name, path, description, theme, language, build_config, created_at, updated_at FROM apps WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.CreatedAt, &app.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("app not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan app: %w", err)
	}

	return app, nil
}

// UpdateApp updates an existing app.
func (ds *SQLDataStore) UpdateApp(app *models.App) error {
	query := fmt.Sprintf(`UPDATE apps SET domain_id = ?, name = ?, path = ?, description = ?, theme = ?, language = ?, build_config = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, app.DomainID, app.Name, app.Path, app.Description, app.Theme, app.Language, app.BuildConfig, app.ID)
	if err != nil {
		return fmt.Errorf("failed to update app: %w", err)
	}
	return nil
}

// DeleteApp removes an app by ID.
func (ds *SQLDataStore) DeleteApp(id int) error {
	query := `DELETE FROM apps WHERE id = ?`
	_, err := ds.driver.Execute(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete app: %w", err)
	}
	return nil
}

// ListAppsByDomain retrieves all apps for a domain.
func (ds *SQLDataStore) ListAppsByDomain(domainID int) ([]*models.App, error) {
	query := `SELECT id, domain_id, name, path, description, theme, language, build_config, created_at, updated_at FROM apps WHERE domain_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, domainID)
	if err != nil {
		return nil, fmt.Errorf("failed to list apps: %w", err)
	}
	defer rows.Close()

	var apps []*models.App
	for rows.Next() {
		app := &models.App{}
		if err := rows.Scan(&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.CreatedAt, &app.UpdatedAt); err != nil {
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
	query := `SELECT id, domain_id, name, path, description, theme, language, build_config, created_at, updated_at FROM apps ORDER BY domain_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all apps: %w", err)
	}
	defer rows.Close()

	var apps []*models.App
	for rows.Next() {
		app := &models.App{}
		if err := rows.Scan(&app.ID, &app.DomainID, &app.Name, &app.Path, &app.Description, &app.Theme, &app.Language, &app.BuildConfig, &app.CreatedAt, &app.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan app: %w", err)
		}
		apps = append(apps, app)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over apps: %w", err)
	}

	return apps, nil
}

// =============================================================================
// Project Operations
// =============================================================================

// CreateProject inserts a new project into the database.
func (ds *SQLDataStore) CreateProject(project *models.Project) error {
	query := fmt.Sprintf(`INSERT INTO projects (name, path, description, created_at, updated_at) 
		VALUES (?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, project.Name, project.Path, project.Description)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		project.ID = int(id)
	}

	return nil
}

// GetProjectByName retrieves a project by its name.
func (ds *SQLDataStore) GetProjectByName(name string) (*models.Project, error) {
	project := &models.Project{}
	query := `SELECT id, name, path, description, created_at, updated_at FROM projects WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&project.ID, &project.Name, &project.Path, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	return project, nil
}

// GetProjectByID retrieves a project by its ID.
func (ds *SQLDataStore) GetProjectByID(id int) (*models.Project, error) {
	project := &models.Project{}
	query := `SELECT id, name, path, description, created_at, updated_at FROM projects WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&project.ID, &project.Name, &project.Path, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan project: %w", err)
	}

	return project, nil
}

// UpdateProject updates an existing project.
func (ds *SQLDataStore) UpdateProject(project *models.Project) error {
	query := fmt.Sprintf(`UPDATE projects SET name = ?, path = ?, description = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, project.Name, project.Path, project.Description, project.ID)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

// DeleteProject removes a project by name.
func (ds *SQLDataStore) DeleteProject(name string) error {
	query := `DELETE FROM projects WHERE name = ?`
	_, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}
	return nil
}

// ListProjects retrieves all projects.
func (ds *SQLDataStore) ListProjects() ([]*models.Project, error) {
	query := `SELECT id, name, path, description, created_at, updated_at FROM projects ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []*models.Project
	for rows.Next() {
		project := &models.Project{}
		if err := rows.Scan(&project.ID, &project.Name, &project.Path, &project.Description, &project.CreatedAt, &project.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over projects: %w", err)
	}

	return projects, nil
}

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

	query := fmt.Sprintf(`INSERT INTO workspaces (app_id, name, description, image_name, status, nvim_structure, nvim_plugins, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, workspace.AppID, workspace.Name, workspace.Description, workspace.ImageName, workspace.Status, workspace.NvimStructure, workspace.NvimPlugins)
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
	query := `SELECT id, app_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at 
		FROM workspaces WHERE app_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, appID, name)
	if err := row.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Description,
		&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure,
		&workspace.NvimPlugins, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
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
	query := `SELECT id, app_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at 
		FROM workspaces WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Description,
		&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure,
		&workspace.NvimPlugins, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("workspace not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan workspace: %w", err)
	}

	return workspace, nil
}

// UpdateWorkspace updates an existing workspace.
func (ds *SQLDataStore) UpdateWorkspace(workspace *models.Workspace) error {
	query := fmt.Sprintf(`UPDATE workspaces SET name = ?, description = ?, image_name = ?, container_id = ?, 
		status = ?, nvim_structure = ?, nvim_plugins = ?, updated_at = %s WHERE id = ?`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, workspace.Name, workspace.Description, workspace.ImageName,
		workspace.ContainerID, workspace.Status, workspace.NvimStructure, workspace.NvimPlugins, workspace.ID)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}
	return nil
}

// DeleteWorkspace removes a workspace by ID.
func (ds *SQLDataStore) DeleteWorkspace(id int) error {
	query := `DELETE FROM workspaces WHERE id = ?`
	_, err := ds.driver.Execute(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete workspace: %w", err)
	}
	return nil
}

// ListWorkspacesByApp retrieves all workspaces for an app.
func (ds *SQLDataStore) ListWorkspacesByApp(appID int) ([]*models.Workspace, error) {
	query := `SELECT id, app_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at 
		FROM workspaces WHERE app_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, appID)
	if err != nil {
		return nil, fmt.Errorf("failed to list workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*models.Workspace
	for rows.Next() {
		workspace := &models.Workspace{}
		if err := rows.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Description,
			&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure,
			&workspace.NvimPlugins, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
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
	query := `SELECT id, app_id, name, description, image_name, container_id, status, nvim_structure, nvim_plugins, created_at, updated_at 
		FROM workspaces ORDER BY app_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*models.Workspace
	for rows.Next() {
		workspace := &models.Workspace{}
		if err := rows.Scan(&workspace.ID, &workspace.AppID, &workspace.Name, &workspace.Description,
			&workspace.ImageName, &workspace.ContainerID, &workspace.Status, &workspace.NvimStructure,
			&workspace.NvimPlugins, &workspace.CreatedAt, &workspace.UpdatedAt); err != nil {
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
		w.id, w.app_id, w.name, w.description, w.image_name, w.container_id, w.status, w.nvim_structure, w.nvim_plugins, w.created_at, w.updated_at,
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
			&workspace.NvimPlugins, &workspace.CreatedAt, &workspace.UpdatedAt,
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

// =============================================================================
// Context Operations
// =============================================================================

// GetContext retrieves the current context.
func (ds *SQLDataStore) GetContext() (*models.Context, error) {
	context := &models.Context{}
	query := `SELECT id, active_ecosystem_id, active_domain_id, active_app_id, active_workspace_id, active_project_id, updated_at FROM context WHERE id = 1`

	row := ds.driver.QueryRow(query)
	if err := row.Scan(&context.ID, &context.ActiveEcosystemID, &context.ActiveDomainID, &context.ActiveAppID, &context.ActiveWorkspaceID, &context.ActiveProjectID, &context.UpdatedAt); err != nil {
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

// SetActiveProject sets the active project in the context.
// DEPRECATED: Use SetActiveApp instead. Will be removed in v0.9.0.
func (ds *SQLDataStore) SetActiveProject(projectID *int) error {
	query := fmt.Sprintf(`UPDATE context SET active_project_id = ?, updated_at = %s WHERE id = 1`,
		ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, projectID)
	if err != nil {
		return fmt.Errorf("failed to set active project: %w", err)
	}
	return nil
}

// =============================================================================
// Plugin Operations
// =============================================================================

// CreatePlugin inserts a new nvim plugin.
func (ds *SQLDataStore) CreatePlugin(plugin *models.NvimPluginDB) error {
	query := fmt.Sprintf(`INSERT INTO nvim_plugins (name, description, repo, branch, version, priority, lazy, 
		event, ft, keys, cmd, dependencies, build, config, init, opts, keymaps, category, tags, enabled, 
		created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		plugin.Name, plugin.Description, plugin.Repo, plugin.Branch, plugin.Version, plugin.Priority,
		plugin.Lazy, plugin.Event, plugin.Ft, plugin.Keys, plugin.Cmd, plugin.Dependencies, plugin.Build,
		plugin.Config, plugin.Init, plugin.Opts, plugin.Keymaps, plugin.Category, plugin.Tags, plugin.Enabled)

	if err != nil {
		return fmt.Errorf("failed to create plugin: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		plugin.ID = int(id)
	}

	return nil
}

// GetPluginByName retrieves a plugin by its name.
func (ds *SQLDataStore) GetPluginByName(name string) (*models.NvimPluginDB, error) {
	plugin := &models.NvimPluginDB{}
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd, 
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(
		&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
		&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
		&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
		&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plugin not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan plugin: %w", err)
	}

	return plugin, nil
}

// GetPluginByID retrieves a plugin by its ID.
func (ds *SQLDataStore) GetPluginByID(id int) (*models.NvimPluginDB, error) {
	plugin := &models.NvimPluginDB{}
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd, 
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(
		&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
		&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
		&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
		&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("plugin not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan plugin: %w", err)
	}

	return plugin, nil
}

// UpdatePlugin updates an existing plugin.
func (ds *SQLDataStore) UpdatePlugin(plugin *models.NvimPluginDB) error {
	query := fmt.Sprintf(`UPDATE nvim_plugins SET description = ?, repo = ?, branch = ?, version = ?, priority = ?, 
		lazy = ?, event = ?, ft = ?, keys = ?, cmd = ?, dependencies = ?, build = ?, config = ?, init = ?,
		opts = ?, keymaps = ?, category = ?, tags = ?, enabled = ?, updated_at = %s 
		WHERE name = ?`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		plugin.Description, plugin.Repo, plugin.Branch, plugin.Version, plugin.Priority,
		plugin.Lazy, plugin.Event, plugin.Ft, plugin.Keys, plugin.Cmd, plugin.Dependencies, plugin.Build,
		plugin.Config, plugin.Init, plugin.Opts, plugin.Keymaps, plugin.Category, plugin.Tags, plugin.Enabled,
		plugin.Name)

	if err != nil {
		return fmt.Errorf("failed to update plugin: %w", err)
	}
	return nil
}

// DeletePlugin removes a plugin by name.
func (ds *SQLDataStore) DeletePlugin(name string) error {
	query := `DELETE FROM nvim_plugins WHERE name = ?`
	_, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete plugin: %w", err)
	}
	return nil
}

// ListPlugins retrieves all plugins.
func (ds *SQLDataStore) ListPlugins() ([]*models.NvimPluginDB, error) {
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd,
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*models.NvimPluginDB
	for rows.Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over plugins: %w", err)
	}

	return plugins, nil
}

// ListPluginsByCategory retrieves plugins filtered by category.
func (ds *SQLDataStore) ListPluginsByCategory(category string) ([]*models.NvimPluginDB, error) {
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd,
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE category = ? ORDER BY name`

	rows, err := ds.driver.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by category: %w", err)
	}
	defer rows.Close()

	var plugins []*models.NvimPluginDB
	for rows.Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over plugins: %w", err)
	}

	return plugins, nil
}

// ListPluginsByTags retrieves plugins that have any of the specified tags.
func (ds *SQLDataStore) ListPluginsByTags(tags []string) ([]*models.NvimPluginDB, error) {
	if len(tags) == 0 {
		return []*models.NvimPluginDB{}, nil
	}

	// Build query with LIKE clauses for each tag
	// Tags are stored as comma-separated string
	query := `SELECT id, name, description, repo, branch, version, priority, lazy, event, ft, keys, cmd,
		dependencies, build, config, init, opts, keymaps, category, tags, enabled, created_at, updated_at
		FROM nvim_plugins WHERE `

	var conditions []string
	var args []interface{}
	for _, tag := range tags {
		conditions = append(conditions, "tags LIKE ?")
		args = append(args, "%"+tag+"%")
	}
	query += "(" + joinStrings(conditions, " OR ") + ") ORDER BY name"

	rows, err := ds.driver.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list plugins by tags: %w", err)
	}
	defer rows.Close()

	var plugins []*models.NvimPluginDB
	for rows.Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over plugins: %w", err)
	}

	return plugins, nil
}

// =============================================================================
// Workspace Plugin Associations
// =============================================================================

// AddPluginToWorkspace associates a plugin with a workspace.
func (ds *SQLDataStore) AddPluginToWorkspace(workspaceID int, pluginID int) error {
	query := fmt.Sprintf(`INSERT OR IGNORE INTO workspace_plugins (workspace_id, plugin_id, enabled, created_at)
		VALUES (?, ?, %s, %s)`, ds.queryBuilder.Boolean(true), ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query, workspaceID, pluginID)
	if err != nil {
		return fmt.Errorf("failed to add plugin to workspace: %w", err)
	}
	return nil
}

// RemovePluginFromWorkspace removes a plugin association from a workspace.
func (ds *SQLDataStore) RemovePluginFromWorkspace(workspaceID int, pluginID int) error {
	query := `DELETE FROM workspace_plugins WHERE workspace_id = ? AND plugin_id = ?`
	_, err := ds.driver.Execute(query, workspaceID, pluginID)
	if err != nil {
		return fmt.Errorf("failed to remove plugin from workspace: %w", err)
	}
	return nil
}

// GetWorkspacePlugins retrieves all plugins associated with a workspace.
func (ds *SQLDataStore) GetWorkspacePlugins(workspaceID int) ([]*models.NvimPluginDB, error) {
	query := fmt.Sprintf(`SELECT p.id, p.name, p.description, p.repo, p.branch, p.version, p.priority, p.lazy, 
		p.event, p.ft, p.keys, p.cmd, p.dependencies, p.build, p.config, p.init, p.opts, p.keymaps,
		p.category, p.tags, p.enabled, p.created_at, p.updated_at
		FROM nvim_plugins p
		JOIN workspace_plugins wp ON p.id = wp.plugin_id
		WHERE wp.workspace_id = ? AND wp.enabled = %s
		ORDER BY p.priority DESC, p.name`, ds.queryBuilder.Boolean(true))

	rows, err := ds.driver.Query(query, workspaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspace plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*models.NvimPluginDB
	for rows.Next() {
		plugin := &models.NvimPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Branch, &plugin.Version,
			&plugin.Priority, &plugin.Lazy, &plugin.Event, &plugin.Ft, &plugin.Keys, &plugin.Cmd,
			&plugin.Dependencies, &plugin.Build, &plugin.Config, &plugin.Init, &plugin.Opts, &plugin.Keymaps,
			&plugin.Category, &plugin.Tags, &plugin.Enabled, &plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over workspace plugins: %w", err)
	}

	return plugins, nil
}

// SetWorkspacePluginEnabled enables or disables a plugin for a workspace.
func (ds *SQLDataStore) SetWorkspacePluginEnabled(workspaceID int, pluginID int, enabled bool) error {
	query := `UPDATE workspace_plugins SET enabled = ? WHERE workspace_id = ? AND plugin_id = ?`
	_, err := ds.driver.Execute(query, enabled, workspaceID, pluginID)
	if err != nil {
		return fmt.Errorf("failed to set workspace plugin enabled: %w", err)
	}
	return nil
}

// =============================================================================
// Theme Operations
// =============================================================================

// CreateTheme inserts a new nvim theme.
func (ds *SQLDataStore) CreateTheme(theme *models.NvimThemeDB) error {
	query := fmt.Sprintf(`INSERT INTO nvim_themes (name, description, author, category, plugin_repo, 
		plugin_branch, plugin_tag, style, transparent, colors, options, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		theme.Name, theme.Description, theme.Author, theme.Category, theme.PluginRepo,
		theme.PluginBranch, theme.PluginTag, theme.Style, theme.Transparent,
		theme.Colors, theme.Options, theme.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create theme: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		theme.ID = int(id)
	}

	return nil
}

// GetThemeByName retrieves a theme by its name.
func (ds *SQLDataStore) GetThemeByName(name string) (*models.NvimThemeDB, error) {
	theme := &models.NvimThemeDB{}
	query := `SELECT id, name, description, author, category, plugin_repo, plugin_branch, plugin_tag,
		style, transparent, colors, options, is_active, created_at, updated_at
		FROM nvim_themes WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(
		&theme.ID, &theme.Name, &theme.Description, &theme.Author, &theme.Category, &theme.PluginRepo,
		&theme.PluginBranch, &theme.PluginTag, &theme.Style, &theme.Transparent,
		&theme.Colors, &theme.Options, &theme.IsActive, &theme.CreatedAt, &theme.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("theme not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan theme: %w", err)
	}

	return theme, nil
}

// GetThemeByID retrieves a theme by its ID.
func (ds *SQLDataStore) GetThemeByID(id int) (*models.NvimThemeDB, error) {
	theme := &models.NvimThemeDB{}
	query := `SELECT id, name, description, author, category, plugin_repo, plugin_branch, plugin_tag,
		style, transparent, colors, options, is_active, created_at, updated_at
		FROM nvim_themes WHERE id = ?`

	row := ds.driver.QueryRow(query, id)
	if err := row.Scan(
		&theme.ID, &theme.Name, &theme.Description, &theme.Author, &theme.Category, &theme.PluginRepo,
		&theme.PluginBranch, &theme.PluginTag, &theme.Style, &theme.Transparent,
		&theme.Colors, &theme.Options, &theme.IsActive, &theme.CreatedAt, &theme.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("theme not found: %d", id)
		}
		return nil, fmt.Errorf("failed to scan theme: %w", err)
	}

	return theme, nil
}

// UpdateTheme updates an existing theme.
func (ds *SQLDataStore) UpdateTheme(theme *models.NvimThemeDB) error {
	query := fmt.Sprintf(`UPDATE nvim_themes SET description = ?, author = ?, category = ?, plugin_repo = ?,
		plugin_branch = ?, plugin_tag = ?, style = ?, transparent = ?, colors = ?, options = ?,
		is_active = ?, updated_at = %s WHERE name = ?`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		theme.Description, theme.Author, theme.Category, theme.PluginRepo,
		theme.PluginBranch, theme.PluginTag, theme.Style, theme.Transparent,
		theme.Colors, theme.Options, theme.IsActive, theme.Name)

	if err != nil {
		return fmt.Errorf("failed to update theme: %w", err)
	}
	return nil
}

// DeleteTheme removes a theme by name.
func (ds *SQLDataStore) DeleteTheme(name string) error {
	query := `DELETE FROM nvim_themes WHERE name = ?`
	_, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete theme: %w", err)
	}
	return nil
}

// ListThemes retrieves all themes.
func (ds *SQLDataStore) ListThemes() ([]*models.NvimThemeDB, error) {
	query := `SELECT id, name, description, author, category, plugin_repo, plugin_branch, plugin_tag,
		style, transparent, colors, options, is_active, created_at, updated_at
		FROM nvim_themes ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list themes: %w", err)
	}
	defer rows.Close()

	var themes []*models.NvimThemeDB
	for rows.Next() {
		theme := &models.NvimThemeDB{}
		if err := rows.Scan(
			&theme.ID, &theme.Name, &theme.Description, &theme.Author, &theme.Category, &theme.PluginRepo,
			&theme.PluginBranch, &theme.PluginTag, &theme.Style, &theme.Transparent,
			&theme.Colors, &theme.Options, &theme.IsActive, &theme.CreatedAt, &theme.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan theme: %w", err)
		}
		themes = append(themes, theme)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over themes: %w", err)
	}

	return themes, nil
}

// ListThemesByCategory retrieves themes filtered by category.
func (ds *SQLDataStore) ListThemesByCategory(category string) ([]*models.NvimThemeDB, error) {
	query := `SELECT id, name, description, author, category, plugin_repo, plugin_branch, plugin_tag,
		style, transparent, colors, options, is_active, created_at, updated_at
		FROM nvim_themes WHERE category = ? ORDER BY name`

	rows, err := ds.driver.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list themes by category: %w", err)
	}
	defer rows.Close()

	var themes []*models.NvimThemeDB
	for rows.Next() {
		theme := &models.NvimThemeDB{}
		if err := rows.Scan(
			&theme.ID, &theme.Name, &theme.Description, &theme.Author, &theme.Category, &theme.PluginRepo,
			&theme.PluginBranch, &theme.PluginTag, &theme.Style, &theme.Transparent,
			&theme.Colors, &theme.Options, &theme.IsActive, &theme.CreatedAt, &theme.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan theme: %w", err)
		}
		themes = append(themes, theme)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over themes: %w", err)
	}

	return themes, nil
}

// GetActiveTheme retrieves the currently active theme.
func (ds *SQLDataStore) GetActiveTheme() (*models.NvimThemeDB, error) {
	theme := &models.NvimThemeDB{}
	query := `SELECT id, name, description, author, category, plugin_repo, plugin_branch, plugin_tag,
		style, transparent, colors, options, is_active, created_at, updated_at
		FROM nvim_themes WHERE is_active = 1 LIMIT 1`

	row := ds.driver.QueryRow(query)
	if err := row.Scan(
		&theme.ID, &theme.Name, &theme.Description, &theme.Author, &theme.Category, &theme.PluginRepo,
		&theme.PluginBranch, &theme.PluginTag, &theme.Style, &theme.Transparent,
		&theme.Colors, &theme.Options, &theme.IsActive, &theme.CreatedAt, &theme.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No active theme
		}
		return nil, fmt.Errorf("failed to scan active theme: %w", err)
	}

	return theme, nil
}

// SetActiveTheme sets the active theme by name (deactivates others).
func (ds *SQLDataStore) SetActiveTheme(name string) error {
	// First, verify the theme exists
	if _, err := ds.GetThemeByName(name); err != nil {
		return err
	}

	// Deactivate all themes
	if err := ds.ClearActiveTheme(); err != nil {
		return err
	}

	// Activate the specified theme
	query := fmt.Sprintf(`UPDATE nvim_themes SET is_active = 1, updated_at = %s WHERE name = ?`,
		ds.queryBuilder.Now())
	_, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to set active theme: %w", err)
	}

	return nil
}

// ClearActiveTheme deactivates all themes.
func (ds *SQLDataStore) ClearActiveTheme() error {
	query := `UPDATE nvim_themes SET is_active = 0`
	_, err := ds.driver.Execute(query)
	if err != nil {
		return fmt.Errorf("failed to clear active theme: %w", err)
	}
	return nil
}

// =============================================================================
// Terminal Prompt Operations
// =============================================================================

// CreateTerminalPrompt inserts a new terminal prompt.
func (ds *SQLDataStore) CreateTerminalPrompt(prompt *models.TerminalPromptDB) error {
	query := fmt.Sprintf(`INSERT INTO terminal_prompts (name, description, type, add_newline, palette, format,
		modules, character, palette_ref, colors, raw_config, category, tags, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		prompt.Name, prompt.Description, prompt.Type, prompt.AddNewline, prompt.Palette, prompt.Format,
		prompt.Modules, prompt.Character, prompt.PaletteRef, prompt.Colors, prompt.RawConfig,
		prompt.Category, prompt.Tags, prompt.Enabled)

	if err != nil {
		return fmt.Errorf("failed to create terminal prompt: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		prompt.ID = int(id)
	}

	return nil
}

// GetTerminalPromptByName retrieves a terminal prompt by its name.
func (ds *SQLDataStore) GetTerminalPromptByName(name string) (*models.TerminalPromptDB, error) {
	prompt := &models.TerminalPromptDB{}
	query := `SELECT id, name, description, type, add_newline, palette, format, modules, character, 
		palette_ref, colors, raw_config, category, tags, enabled, created_at, updated_at
		FROM terminal_prompts WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(
		&prompt.ID, &prompt.Name, &prompt.Description, &prompt.Type, &prompt.AddNewline,
		&prompt.Palette, &prompt.Format, &prompt.Modules, &prompt.Character, &prompt.PaletteRef,
		&prompt.Colors, &prompt.RawConfig, &prompt.Category, &prompt.Tags, &prompt.Enabled,
		&prompt.CreatedAt, &prompt.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("terminal prompt not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan terminal prompt: %w", err)
	}

	return prompt, nil
}

// UpdateTerminalPrompt updates an existing terminal prompt.
func (ds *SQLDataStore) UpdateTerminalPrompt(prompt *models.TerminalPromptDB) error {
	query := fmt.Sprintf(`UPDATE terminal_prompts SET description = ?, type = ?, add_newline = ?, palette = ?, 
		format = ?, modules = ?, character = ?, palette_ref = ?, colors = ?, raw_config = ?, 
		category = ?, tags = ?, enabled = ?, updated_at = %s 
		WHERE name = ?`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		prompt.Description, prompt.Type, prompt.AddNewline, prompt.Palette, prompt.Format,
		prompt.Modules, prompt.Character, prompt.PaletteRef, prompt.Colors, prompt.RawConfig,
		prompt.Category, prompt.Tags, prompt.Enabled, prompt.Name)

	if err != nil {
		return fmt.Errorf("failed to update terminal prompt: %w", err)
	}
	return nil
}

// DeleteTerminalPrompt removes a terminal prompt by name.
func (ds *SQLDataStore) DeleteTerminalPrompt(name string) error {
	query := `DELETE FROM terminal_prompts WHERE name = ?`
	_, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete terminal prompt: %w", err)
	}
	return nil
}

// ListTerminalPrompts retrieves all terminal prompts.
func (ds *SQLDataStore) ListTerminalPrompts() ([]*models.TerminalPromptDB, error) {
	query := `SELECT id, name, description, type, add_newline, palette, format, modules, character,
		palette_ref, colors, raw_config, category, tags, enabled, created_at, updated_at
		FROM terminal_prompts ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal prompts: %w", err)
	}
	defer rows.Close()

	var prompts []*models.TerminalPromptDB
	for rows.Next() {
		prompt := &models.TerminalPromptDB{}
		if err := rows.Scan(
			&prompt.ID, &prompt.Name, &prompt.Description, &prompt.Type, &prompt.AddNewline,
			&prompt.Palette, &prompt.Format, &prompt.Modules, &prompt.Character, &prompt.PaletteRef,
			&prompt.Colors, &prompt.RawConfig, &prompt.Category, &prompt.Tags, &prompt.Enabled,
			&prompt.CreatedAt, &prompt.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal prompt: %w", err)
		}
		prompts = append(prompts, prompt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal prompts: %w", err)
	}

	return prompts, nil
}

// ListTerminalPromptsByType retrieves terminal prompts filtered by type.
func (ds *SQLDataStore) ListTerminalPromptsByType(promptType string) ([]*models.TerminalPromptDB, error) {
	query := `SELECT id, name, description, type, add_newline, palette, format, modules, character,
		palette_ref, colors, raw_config, category, tags, enabled, created_at, updated_at
		FROM terminal_prompts WHERE type = ? ORDER BY name`

	rows, err := ds.driver.Query(query, promptType)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal prompts by type: %w", err)
	}
	defer rows.Close()

	var prompts []*models.TerminalPromptDB
	for rows.Next() {
		prompt := &models.TerminalPromptDB{}
		if err := rows.Scan(
			&prompt.ID, &prompt.Name, &prompt.Description, &prompt.Type, &prompt.AddNewline,
			&prompt.Palette, &prompt.Format, &prompt.Modules, &prompt.Character, &prompt.PaletteRef,
			&prompt.Colors, &prompt.RawConfig, &prompt.Category, &prompt.Tags, &prompt.Enabled,
			&prompt.CreatedAt, &prompt.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal prompt: %w", err)
		}
		prompts = append(prompts, prompt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal prompts by type: %w", err)
	}

	return prompts, nil
}

// ListTerminalPromptsByCategory retrieves terminal prompts filtered by category.
func (ds *SQLDataStore) ListTerminalPromptsByCategory(category string) ([]*models.TerminalPromptDB, error) {
	query := `SELECT id, name, description, type, add_newline, palette, format, modules, character,
		palette_ref, colors, raw_config, category, tags, enabled, created_at, updated_at
		FROM terminal_prompts WHERE category = ? ORDER BY name`

	rows, err := ds.driver.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal prompts by category: %w", err)
	}
	defer rows.Close()

	var prompts []*models.TerminalPromptDB
	for rows.Next() {
		prompt := &models.TerminalPromptDB{}
		if err := rows.Scan(
			&prompt.ID, &prompt.Name, &prompt.Description, &prompt.Type, &prompt.AddNewline,
			&prompt.Palette, &prompt.Format, &prompt.Modules, &prompt.Character, &prompt.PaletteRef,
			&prompt.Colors, &prompt.RawConfig, &prompt.Category, &prompt.Tags, &prompt.Enabled,
			&prompt.CreatedAt, &prompt.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal prompt: %w", err)
		}
		prompts = append(prompts, prompt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal prompts by category: %w", err)
	}

	return prompts, nil
}

// =============================================================================
// Terminal Plugin Operations
// =============================================================================

// CreateTerminalPlugin inserts a new terminal plugin.
func (ds *SQLDataStore) CreateTerminalPlugin(plugin *models.TerminalPluginDB) error {
	// Ensure required JSON fields have proper defaults
	if plugin.Dependencies == "" {
		plugin.Dependencies = "[]"
	}
	if plugin.EnvVars == "" {
		plugin.EnvVars = "{}"
	}
	if plugin.Labels == "" {
		plugin.Labels = "{}"
	}

	query := fmt.Sprintf(`INSERT INTO terminal_plugins (name, description, repo, category, shell, manager, 
		load_command, source_file, dependencies, env_vars, labels, enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`,
		ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		plugin.Name, plugin.Description, plugin.Repo, plugin.Category, plugin.Shell, plugin.Manager,
		plugin.LoadCommand, plugin.SourceFile, plugin.Dependencies, plugin.EnvVars, plugin.Labels, plugin.Enabled)

	if err != nil {
		return fmt.Errorf("failed to create terminal plugin: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		plugin.ID = int(id)

		// Read back the created record to get timestamps
		created, err := ds.GetTerminalPlugin(plugin.Name)
		if err == nil {
			plugin.CreatedAt = created.CreatedAt
			plugin.UpdatedAt = created.UpdatedAt
		}
	}

	return nil
}

// GetTerminalPlugin retrieves a terminal plugin by its name.
func (ds *SQLDataStore) GetTerminalPlugin(name string) (*models.TerminalPluginDB, error) {
	plugin := &models.TerminalPluginDB{}
	query := `SELECT id, name, description, repo, category, shell, manager, load_command, 
		source_file, dependencies, env_vars, labels, enabled, created_at, updated_at
		FROM terminal_plugins WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(
		&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Category,
		&plugin.Shell, &plugin.Manager, &plugin.LoadCommand, &plugin.SourceFile,
		&plugin.Dependencies, &plugin.EnvVars, &plugin.Labels, &plugin.Enabled,
		&plugin.CreatedAt, &plugin.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("terminal plugin not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan terminal plugin: %w", err)
	}

	return plugin, nil
}

// UpdateTerminalPlugin updates an existing terminal plugin.
func (ds *SQLDataStore) UpdateTerminalPlugin(plugin *models.TerminalPluginDB) error {
	query := fmt.Sprintf(`UPDATE terminal_plugins SET description = ?, repo = ?, category = ?, shell = ?, 
		manager = ?, load_command = ?, source_file = ?, dependencies = ?, env_vars = ?, labels = ?, 
		enabled = ?, updated_at = %s 
		WHERE name = ?`, ds.queryBuilder.Now())

	_, err := ds.driver.Execute(query,
		plugin.Description, plugin.Repo, plugin.Category, plugin.Shell, plugin.Manager,
		plugin.LoadCommand, plugin.SourceFile, plugin.Dependencies, plugin.EnvVars, plugin.Labels,
		plugin.Enabled, plugin.Name)

	if err != nil {
		return fmt.Errorf("failed to update terminal plugin: %w", err)
	}
	return nil
}

// UpsertTerminalPlugin creates or updates a terminal plugin (by name).
func (ds *SQLDataStore) UpsertTerminalPlugin(plugin *models.TerminalPluginDB) error {
	// First try to get the existing plugin
	existing, err := ds.GetTerminalPlugin(plugin.Name)
	if err == nil {
		// Plugin exists, update it
		plugin.ID = existing.ID
		return ds.UpdateTerminalPlugin(plugin)
	}

	// Plugin doesn't exist, create it
	return ds.CreateTerminalPlugin(plugin)
}

// DeleteTerminalPlugin removes a terminal plugin by name.
func (ds *SQLDataStore) DeleteTerminalPlugin(name string) error {
	query := `DELETE FROM terminal_plugins WHERE name = ?`
	_, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete terminal plugin: %w", err)
	}
	return nil
}

// ListTerminalPlugins retrieves all terminal plugins.
func (ds *SQLDataStore) ListTerminalPlugins() ([]*models.TerminalPluginDB, error) {
	query := `SELECT id, name, description, repo, category, shell, manager, load_command,
		source_file, dependencies, env_vars, labels, enabled, created_at, updated_at
		FROM terminal_plugins ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal plugins: %w", err)
	}
	defer rows.Close()

	var plugins []*models.TerminalPluginDB
	for rows.Next() {
		plugin := &models.TerminalPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Category,
			&plugin.Shell, &plugin.Manager, &plugin.LoadCommand, &plugin.SourceFile,
			&plugin.Dependencies, &plugin.EnvVars, &plugin.Labels, &plugin.Enabled,
			&plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal plugins: %w", err)
	}

	return plugins, nil
}

// ListTerminalPluginsByCategory retrieves terminal plugins filtered by category.
func (ds *SQLDataStore) ListTerminalPluginsByCategory(category string) ([]*models.TerminalPluginDB, error) {
	query := `SELECT id, name, description, repo, category, shell, manager, load_command,
		source_file, dependencies, env_vars, labels, enabled, created_at, updated_at
		FROM terminal_plugins WHERE category = ? ORDER BY name`

	rows, err := ds.driver.Query(query, category)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal plugins by category: %w", err)
	}
	defer rows.Close()

	var plugins []*models.TerminalPluginDB
	for rows.Next() {
		plugin := &models.TerminalPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Category,
			&plugin.Shell, &plugin.Manager, &plugin.LoadCommand, &plugin.SourceFile,
			&plugin.Dependencies, &plugin.EnvVars, &plugin.Labels, &plugin.Enabled,
			&plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal plugins by category: %w", err)
	}

	return plugins, nil
}

// ListTerminalPluginsByShell retrieves terminal plugins filtered by shell.
func (ds *SQLDataStore) ListTerminalPluginsByShell(shell string) ([]*models.TerminalPluginDB, error) {
	query := `SELECT id, name, description, repo, category, shell, manager, load_command,
		source_file, dependencies, env_vars, labels, enabled, created_at, updated_at
		FROM terminal_plugins WHERE shell = ? ORDER BY name`

	rows, err := ds.driver.Query(query, shell)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal plugins by shell: %w", err)
	}
	defer rows.Close()

	var plugins []*models.TerminalPluginDB
	for rows.Next() {
		plugin := &models.TerminalPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Category,
			&plugin.Shell, &plugin.Manager, &plugin.LoadCommand, &plugin.SourceFile,
			&plugin.Dependencies, &plugin.EnvVars, &plugin.Labels, &plugin.Enabled,
			&plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal plugins by shell: %w", err)
	}

	return plugins, nil
}

// ListTerminalPluginsByManager retrieves terminal plugins filtered by manager.
func (ds *SQLDataStore) ListTerminalPluginsByManager(manager string) ([]*models.TerminalPluginDB, error) {
	query := `SELECT id, name, description, repo, category, shell, manager, load_command,
		source_file, dependencies, env_vars, labels, enabled, created_at, updated_at
		FROM terminal_plugins WHERE manager = ? ORDER BY name`

	rows, err := ds.driver.Query(query, manager)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal plugins by manager: %w", err)
	}
	defer rows.Close()

	var plugins []*models.TerminalPluginDB
	for rows.Next() {
		plugin := &models.TerminalPluginDB{}
		if err := rows.Scan(
			&plugin.ID, &plugin.Name, &plugin.Description, &plugin.Repo, &plugin.Category,
			&plugin.Shell, &plugin.Manager, &plugin.LoadCommand, &plugin.SourceFile,
			&plugin.Dependencies, &plugin.EnvVars, &plugin.Labels, &plugin.Enabled,
			&plugin.CreatedAt, &plugin.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal plugin: %w", err)
		}
		plugins = append(plugins, plugin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal plugins by manager: %w", err)
	}

	return plugins, nil
}

// =============================================================================
// Terminal Emulator Operations
// =============================================================================

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

	emulator := &models.TerminalEmulatorDB{}
	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(
		&emulator.ID, &emulator.Name, &emulator.Description, &emulator.Type, &emulator.Config,
		&emulator.ThemeRef, &emulator.Category, &emulator.Labels, &emulator.Workspace,
		&emulator.Enabled, &emulator.CreatedAt, &emulator.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("terminal emulator not found: %s", name)
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

// UpsertTerminalEmulator creates or updates a terminal emulator (by name).
func (ds *SQLDataStore) UpsertTerminalEmulator(emulator *models.TerminalEmulatorDB) error {
	// First try to get the existing emulator
	existing, err := ds.GetTerminalEmulator(emulator.Name)
	if err == nil {
		// Emulator exists, update it
		emulator.ID = existing.ID
		return ds.UpdateTerminalEmulator(emulator)
	}

	// Emulator doesn't exist, create it
	return ds.CreateTerminalEmulator(emulator)
}

// DeleteTerminalEmulator removes a terminal emulator by name.
func (ds *SQLDataStore) DeleteTerminalEmulator(name string) error {
	query := `DELETE FROM terminal_emulators WHERE name = ?`
	_, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete terminal emulator: %w", err)
	}
	return nil
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
		emulator := &models.TerminalEmulatorDB{}
		if err := rows.Scan(
			&emulator.ID, &emulator.Name, &emulator.Description, &emulator.Type, &emulator.Config,
			&emulator.ThemeRef, &emulator.Category, &emulator.Labels, &emulator.Workspace,
			&emulator.Enabled, &emulator.CreatedAt, &emulator.UpdatedAt,
		); err != nil {
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
		emulator := &models.TerminalEmulatorDB{}
		if err := rows.Scan(
			&emulator.ID, &emulator.Name, &emulator.Description, &emulator.Type, &emulator.Config,
			&emulator.ThemeRef, &emulator.Category, &emulator.Labels, &emulator.Workspace,
			&emulator.Enabled, &emulator.CreatedAt, &emulator.UpdatedAt,
		); err != nil {
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
		emulator := &models.TerminalEmulatorDB{}
		if err := rows.Scan(
			&emulator.ID, &emulator.Name, &emulator.Description, &emulator.Type, &emulator.Config,
			&emulator.ThemeRef, &emulator.Category, &emulator.Labels, &emulator.Workspace,
			&emulator.Enabled, &emulator.CreatedAt, &emulator.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan terminal emulator: %w", err)
		}
		emulators = append(emulators, emulator)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate terminal emulators by workspace: %w", err)
	}

	return emulators, nil
}

// =============================================================================
// Credential Operations
// =============================================================================

// CreateCredential inserts a new credential configuration.
func (ds *SQLDataStore) CreateCredential(credential *models.CredentialDB) error {
	query := fmt.Sprintf(`INSERT INTO credentials (scope_type, scope_id, name, source, service, env_var, value, description, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		credential.ScopeType,
		credential.ScopeID,
		credential.Name,
		credential.Source,
		credential.Service,
		credential.EnvVar,
		credential.Value,
		credential.Description,
	)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		credential.ID = id
	}

	return nil
}

// GetCredential retrieves a credential by scope and name.
func (ds *SQLDataStore) GetCredential(scopeType models.CredentialScopeType, scopeID int64, name string) (*models.CredentialDB, error) {
	credential := &models.CredentialDB{}
	query := `SELECT id, scope_type, scope_id, name, source, service, env_var, value, description, created_at, updated_at 
		FROM credentials WHERE scope_type = ? AND scope_id = ? AND name = ?`

	row := ds.driver.QueryRow(query, scopeType, scopeID, name)
	if err := row.Scan(
		&credential.ID,
		&credential.ScopeType,
		&credential.ScopeID,
		&credential.Name,
		&credential.Source,
		&credential.Service,
		&credential.EnvVar,
		&credential.Value,
		&credential.Description,
		&credential.CreatedAt,
		&credential.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("credential not found: %s (scope: %s, id: %d)", name, scopeType, scopeID)
		}
		return nil, fmt.Errorf("failed to scan credential: %w", err)
	}

	return credential, nil
}

// UpdateCredential updates an existing credential.
func (ds *SQLDataStore) UpdateCredential(credential *models.CredentialDB) error {
	query := fmt.Sprintf(`UPDATE credentials SET source = ?, service = ?, env_var = ?, value = ?, description = ?, updated_at = %s 
		WHERE scope_type = ? AND scope_id = ? AND name = ?`, ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query,
		credential.Source,
		credential.Service,
		credential.EnvVar,
		credential.Value,
		credential.Description,
		credential.ScopeType,
		credential.ScopeID,
		credential.Name,
	)
	if err != nil {
		return fmt.Errorf("failed to update credential: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("credential not found: %s (scope: %s, id: %d)", credential.Name, credential.ScopeType, credential.ScopeID)
	}

	return nil
}

// DeleteCredential removes a credential by scope and name.
func (ds *SQLDataStore) DeleteCredential(scopeType models.CredentialScopeType, scopeID int64, name string) error {
	query := `DELETE FROM credentials WHERE scope_type = ? AND scope_id = ? AND name = ?`

	result, err := ds.driver.Execute(query, scopeType, scopeID, name)
	if err != nil {
		return fmt.Errorf("failed to delete credential: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("credential not found: %s (scope: %s, id: %d)", name, scopeType, scopeID)
	}

	return nil
}

// ListCredentialsByScope retrieves all credentials for a specific scope.
func (ds *SQLDataStore) ListCredentialsByScope(scopeType models.CredentialScopeType, scopeID int64) ([]*models.CredentialDB, error) {
	query := `SELECT id, scope_type, scope_id, name, source, service, env_var, value, description, created_at, updated_at 
		FROM credentials WHERE scope_type = ? AND scope_id = ? ORDER BY name`

	rows, err := ds.driver.Query(query, scopeType, scopeID)
	if err != nil {
		return nil, fmt.Errorf("failed to list credentials: %w", err)
	}
	defer rows.Close()

	var credentials []*models.CredentialDB
	for rows.Next() {
		credential := &models.CredentialDB{}
		if err := rows.Scan(
			&credential.ID,
			&credential.ScopeType,
			&credential.ScopeID,
			&credential.Name,
			&credential.Source,
			&credential.Service,
			&credential.EnvVar,
			&credential.Value,
			&credential.Description,
			&credential.CreatedAt,
			&credential.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}
		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credentials: %w", err)
	}

	return credentials, nil
}

// ListAllCredentials retrieves all credentials across all scopes.
func (ds *SQLDataStore) ListAllCredentials() ([]*models.CredentialDB, error) {
	query := `SELECT id, scope_type, scope_id, name, source, service, env_var, value, description, created_at, updated_at 
		FROM credentials ORDER BY scope_type, scope_id, name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list all credentials: %w", err)
	}
	defer rows.Close()

	var credentials []*models.CredentialDB
	for rows.Next() {
		credential := &models.CredentialDB{}
		if err := rows.Scan(
			&credential.ID,
			&credential.ScopeType,
			&credential.ScopeID,
			&credential.Name,
			&credential.Source,
			&credential.Service,
			&credential.EnvVar,
			&credential.Value,
			&credential.Description,
			&credential.CreatedAt,
			&credential.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}
		credentials = append(credentials, credential)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating credentials: %w", err)
	}

	return credentials, nil
}

// =============================================================================
// Default Operations
// =============================================================================

// GetDefault retrieves a default value by key.
// Returns empty string if key is not found (not an error).
func (ds *SQLDataStore) GetDefault(key string) (string, error) {
	query := `SELECT value FROM defaults WHERE key = ?`

	var value string
	err := ds.driver.QueryRow(query, key).Scan(&value)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return "", nil // Not found, return empty string (not an error)
		}
		return "", fmt.Errorf("failed to get default for key %s: %w", key, err)
	}

	return value, nil
}

// SetDefault sets a default value for the given key.
// Uses upsert behavior (INSERT OR REPLACE).
func (ds *SQLDataStore) SetDefault(key, value string) error {
	query := `
		INSERT OR REPLACE INTO defaults (key, value, updated_at) 
		VALUES (?, ?, CURRENT_TIMESTAMP)
	`

	_, err := ds.driver.Execute(query, key, value)
	if err != nil {
		return fmt.Errorf("failed to set default for key %s: %w", key, err)
	}

	return nil
}

// DeleteDefault removes a default value by key.
// No error if key doesn't exist.
func (ds *SQLDataStore) DeleteDefault(key string) error {
	query := `DELETE FROM defaults WHERE key = ?`

	_, err := ds.driver.Execute(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete default for key %s: %w", key, err)
	}

	// No error if key doesn't exist (DELETE just affects 0 rows)
	return nil
}

// ListDefaults retrieves all default values as a key-value map.
func (ds *SQLDataStore) ListDefaults() (map[string]string, error) {
	query := `SELECT key, value FROM defaults ORDER BY key`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list defaults: %w", err)
	}
	defer rows.Close()

	defaults := make(map[string]string)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan default: %w", err)
		}
		defaults[key] = value
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating defaults: %w", err)
	}

	return defaults, nil
}

// =============================================================================
// Package Operations
// =============================================================================

// CreatePackage inserts a new nvim package into the database.
func (ds *SQLDataStore) CreatePackage(pkg *models.NvimPackageDB) error {
	query := fmt.Sprintf(`INSERT INTO nvim_packages (name, description, category, labels, plugins, extends, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, pkg.Name, pkg.Description, pkg.Category, pkg.Labels, pkg.Plugins, pkg.Extends)
	if err != nil {
		return fmt.Errorf("failed to create package: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		pkg.ID = int(id)
	}

	return nil
}

// UpdatePackage updates an existing nvim package.
func (ds *SQLDataStore) UpdatePackage(pkg *models.NvimPackageDB) error {
	query := fmt.Sprintf(`UPDATE nvim_packages 
		SET description = ?, category = ?, labels = ?, plugins = ?, extends = ?, updated_at = %s 
		WHERE name = ?`, ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, pkg.Description, pkg.Category, pkg.Labels, pkg.Plugins, pkg.Extends, pkg.Name)
	if err != nil {
		return fmt.Errorf("failed to update package: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return fmt.Errorf("package not found: %s", pkg.Name)
	}

	return nil
}

// UpsertPackage creates or updates an nvim package (by name).
func (ds *SQLDataStore) UpsertPackage(pkg *models.NvimPackageDB) error {
	// Try to get existing package first
	existing, err := ds.GetPackage(pkg.Name)
	if err != nil {
		// Package doesn't exist, create it
		return ds.CreatePackage(pkg)
	}

	// Package exists, update it (preserve the ID)
	pkg.ID = existing.ID
	return ds.UpdatePackage(pkg)
}

// DeletePackage removes a package by name.
func (ds *SQLDataStore) DeletePackage(name string) error {
	query := `DELETE FROM nvim_packages WHERE name = ?`

	result, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete package: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return fmt.Errorf("package not found: %s", name)
	}

	return nil
}

// GetPackage retrieves a package by its name.
func (ds *SQLDataStore) GetPackage(name string) (*models.NvimPackageDB, error) {
	pkg := &models.NvimPackageDB{}
	query := `SELECT id, name, description, category, labels, plugins, extends, created_at, updated_at FROM nvim_packages WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("package not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan package: %w", err)
	}

	return pkg, nil
}

// ListPackages retrieves all packages.
func (ds *SQLDataStore) ListPackages() ([]*models.NvimPackageDB, error) {
	query := `SELECT id, name, description, category, labels, plugins, extends, created_at, updated_at FROM nvim_packages ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}
	defer rows.Close()

	var packages []*models.NvimPackageDB
	for rows.Next() {
		pkg := &models.NvimPackageDB{}
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}
		packages = append(packages, pkg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating packages: %w", err)
	}

	return packages, nil
}

// ListPackagesByLabel retrieves packages that have a specific label key-value pair.
func (ds *SQLDataStore) ListPackagesByLabel(key, value string) ([]*models.NvimPackageDB, error) {
	// Use JSON_EXTRACT if available (SQLite 3.45+) or simple string matching as fallback
	query := `SELECT id, name, description, category, labels, plugins, extends, created_at, updated_at 
		FROM nvim_packages 
		WHERE labels IS NOT NULL 
		AND (json_extract(labels, '$.` + key + `') = ? OR labels LIKE '%"' || ? || '":"' || ? || '"%')
		ORDER BY name`

	rows, err := ds.driver.Query(query, value, key, value)
	if err != nil {
		return nil, fmt.Errorf("failed to list packages by label: %w", err)
	}
	defer rows.Close()

	var packages []*models.NvimPackageDB
	for rows.Next() {
		pkg := &models.NvimPackageDB{}
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan package: %w", err)
		}

		// Double-check the label match in Go for accuracy
		labels := pkg.GetLabels()
		if labels[key] == value {
			packages = append(packages, pkg)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating packages: %w", err)
	}

	return packages, nil
}

// =============================================================================
// Terminal Package Operations
// =============================================================================

// CreateTerminalPackage inserts a new terminal package into the database.
func (ds *SQLDataStore) CreateTerminalPackage(pkg *models.TerminalPackageDB) error {
	query := fmt.Sprintf(`INSERT INTO terminal_packages (name, description, category, labels, plugins, prompts, profiles, wezterm, extends, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, %s, %s)`, ds.queryBuilder.Now(), ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, pkg.Name, pkg.Description, pkg.Category, pkg.Labels, pkg.Plugins, pkg.Prompts, pkg.Profiles, pkg.WezTerm, pkg.Extends)
	if err != nil {
		return fmt.Errorf("failed to create terminal package: %w", err)
	}

	id, err := result.LastInsertId()
	if err == nil {
		pkg.ID = int(id)
	}

	return nil
}

// UpdateTerminalPackage updates an existing terminal package.
func (ds *SQLDataStore) UpdateTerminalPackage(pkg *models.TerminalPackageDB) error {
	query := fmt.Sprintf(`UPDATE terminal_packages 
		SET description = ?, category = ?, labels = ?, plugins = ?, prompts = ?, profiles = ?, wezterm = ?, extends = ?, updated_at = %s 
		WHERE name = ?`, ds.queryBuilder.Now())

	result, err := ds.driver.Execute(query, pkg.Description, pkg.Category, pkg.Labels, pkg.Plugins, pkg.Prompts, pkg.Profiles, pkg.WezTerm, pkg.Extends, pkg.Name)
	if err != nil {
		return fmt.Errorf("failed to update terminal package: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return fmt.Errorf("terminal package not found: %s", pkg.Name)
	}

	return nil
}

// UpsertTerminalPackage creates or updates a terminal package (by name).
func (ds *SQLDataStore) UpsertTerminalPackage(pkg *models.TerminalPackageDB) error {
	// Try to get existing package first
	existing, err := ds.GetTerminalPackage(pkg.Name)
	if err != nil {
		// Package doesn't exist, create it
		return ds.CreateTerminalPackage(pkg)
	}

	// Package exists, update it (preserve the ID)
	pkg.ID = existing.ID
	return ds.UpdateTerminalPackage(pkg)
}

// DeleteTerminalPackage removes a terminal package by name.
func (ds *SQLDataStore) DeleteTerminalPackage(name string) error {
	query := `DELETE FROM terminal_packages WHERE name = ?`

	result, err := ds.driver.Execute(query, name)
	if err != nil {
		return fmt.Errorf("failed to delete terminal package: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err == nil && rowsAffected == 0 {
		return fmt.Errorf("terminal package not found: %s", name)
	}

	return nil
}

// GetTerminalPackage retrieves a terminal package by its name.
func (ds *SQLDataStore) GetTerminalPackage(name string) (*models.TerminalPackageDB, error) {
	pkg := &models.TerminalPackageDB{}
	query := `SELECT id, name, description, category, labels, plugins, prompts, profiles, wezterm, extends, created_at, updated_at FROM terminal_packages WHERE name = ?`

	row := ds.driver.QueryRow(query, name)
	if err := row.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Prompts, &pkg.Profiles, &pkg.WezTerm, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("terminal package not found: %s", name)
		}
		return nil, fmt.Errorf("failed to scan terminal package: %w", err)
	}

	return pkg, nil
}

// ListTerminalPackages retrieves all terminal packages.
func (ds *SQLDataStore) ListTerminalPackages() ([]*models.TerminalPackageDB, error) {
	query := `SELECT id, name, description, category, labels, plugins, prompts, profiles, wezterm, extends, created_at, updated_at FROM terminal_packages ORDER BY name`

	rows, err := ds.driver.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal packages: %w", err)
	}
	defer rows.Close()

	var packages []*models.TerminalPackageDB
	for rows.Next() {
		pkg := &models.TerminalPackageDB{}
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Prompts, &pkg.Profiles, &pkg.WezTerm, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan terminal package: %w", err)
		}
		packages = append(packages, pkg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating terminal packages: %w", err)
	}

	return packages, nil
}

// ListTerminalPackagesByLabel retrieves terminal packages that have a specific label key-value pair.
func (ds *SQLDataStore) ListTerminalPackagesByLabel(key, value string) ([]*models.TerminalPackageDB, error) {
	// Use JSON_EXTRACT if available (SQLite 3.45+) or simple string matching as fallback
	query := `SELECT id, name, description, category, labels, plugins, prompts, profiles, wezterm, extends, created_at, updated_at 
		FROM terminal_packages 
		WHERE labels IS NOT NULL 
		AND (json_extract(labels, '$.` + key + `') = ? OR labels LIKE '%"' || ? || '":"' || ? || '"%')
		ORDER BY name`

	rows, err := ds.driver.Query(query, value, key, value)
	if err != nil {
		return nil, fmt.Errorf("failed to list terminal packages by label: %w", err)
	}
	defer rows.Close()

	var packages []*models.TerminalPackageDB
	for rows.Next() {
		pkg := &models.TerminalPackageDB{}
		if err := rows.Scan(&pkg.ID, &pkg.Name, &pkg.Description, &pkg.Category, &pkg.Labels, &pkg.Plugins, &pkg.Prompts, &pkg.Profiles, &pkg.WezTerm, &pkg.Extends, &pkg.CreatedAt, &pkg.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan terminal package: %w", err)
		}

		// Double-check the label match in Go for accuracy
		labels := pkg.GetLabels()
		if labels[key] == value {
			packages = append(packages, pkg)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating terminal packages: %w", err)
	}

	return packages, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// joinStrings joins strings with a separator (to avoid importing strings package)
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// Ensure SQLDataStore implements DataStore interface
var _ DataStore = (*SQLDataStore)(nil)
