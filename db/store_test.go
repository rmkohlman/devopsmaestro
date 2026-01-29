package db

import (
	"database/sql"
	"devopsmaestro/models"
	"testing"
)

// =============================================================================
// Test Setup Helpers
// =============================================================================

// createTestDataStore creates an in-memory SQLite database with schema for testing
func createTestDataStore(t *testing.T) *SQLDataStore {
	t.Helper()

	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("Failed to create test driver: %v", err)
	}

	if err := driver.Connect(); err != nil {
		t.Fatalf("Failed to connect test driver: %v", err)
	}

	// Create schema
	if err := createTestSchema(driver); err != nil {
		driver.Close()
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return NewSQLDataStore(driver, nil)
}

// createTestSchema creates the required tables for testing
func createTestSchema(driver Driver) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			path TEXT,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			project_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			image_name TEXT,
			container_id TEXT,
			status TEXT DEFAULT 'stopped',
			nvim_structure TEXT,
			nvim_plugins TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (project_id) REFERENCES projects(id),
			UNIQUE(project_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS context (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			active_project_id INTEGER,
			active_workspace_id INTEGER,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (active_project_id) REFERENCES projects(id),
			FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id)
		)`,
		`CREATE TABLE IF NOT EXISTS nvim_plugins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			repo TEXT NOT NULL,
			branch TEXT,
			version TEXT,
			priority INTEGER,
			lazy INTEGER DEFAULT 0,
			event TEXT,
			ft TEXT,
			keys TEXT,
			cmd TEXT,
			dependencies TEXT,
			build TEXT,
			config TEXT,
			init TEXT,
			opts TEXT,
			keymaps TEXT,
			category TEXT,
			tags TEXT,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_plugins (
			workspace_id INTEGER NOT NULL,
			plugin_id INTEGER NOT NULL,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (workspace_id, plugin_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			FOREIGN KEY (plugin_id) REFERENCES nvim_plugins(id) ON DELETE CASCADE
		)`,
		// Initialize context with a single row
		`INSERT OR IGNORE INTO context (id) VALUES (1)`,
	}

	for _, query := range queries {
		if _, err := driver.Execute(query); err != nil {
			return err
		}
	}

	return nil
}

// =============================================================================
// Project CRUD Tests
// =============================================================================

func TestSQLDataStore_CreateProject(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{
		Name: "test-project",
		Path: "/path/to/project",
		Description: sql.NullString{
			String: "Test description",
			Valid:  true,
		},
	}

	err := ds.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	if project.ID == 0 {
		t.Errorf("CreateProject() did not set project.ID")
	}
}

func TestSQLDataStore_GetProjectByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create a project first
	project := &models.Project{
		Name: "findme-project",
		Path: "/path/to/findme",
	}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup CreateProject() error = %v", err)
	}

	// Retrieve by name
	retrieved, err := ds.GetProjectByName("findme-project")
	if err != nil {
		t.Fatalf("GetProjectByName() error = %v", err)
	}

	if retrieved.Name != "findme-project" {
		t.Errorf("GetProjectByName() Name = %q, want %q", retrieved.Name, "findme-project")
	}
	if retrieved.Path != "/path/to/findme" {
		t.Errorf("GetProjectByName() Path = %q, want %q", retrieved.Path, "/path/to/findme")
	}
}

func TestSQLDataStore_GetProjectByName_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	_, err := ds.GetProjectByName("nonexistent")
	if err == nil {
		t.Errorf("GetProjectByName() expected error for nonexistent project")
	}
}

func TestSQLDataStore_GetProjectByID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{
		Name: "getbyid-project",
		Path: "/path/getbyid",
	}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetProjectByID(project.ID)
	if err != nil {
		t.Fatalf("GetProjectByID() error = %v", err)
	}

	if retrieved.ID != project.ID {
		t.Errorf("GetProjectByID() ID = %d, want %d", retrieved.ID, project.ID)
	}
}

func TestSQLDataStore_UpdateProject(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{
		Name: "update-project",
		Path: "/original/path",
	}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Update the project
	project.Path = "/updated/path"
	project.Description = sql.NullString{String: "Updated description", Valid: true}

	if err := ds.UpdateProject(project); err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}

	// Verify update
	retrieved, err := ds.GetProjectByID(project.ID)
	if err != nil {
		t.Fatalf("Verification error: %v", err)
	}

	if retrieved.Path != "/updated/path" {
		t.Errorf("UpdateProject() Path = %q, want %q", retrieved.Path, "/updated/path")
	}
	if retrieved.Description.String != "Updated description" {
		t.Errorf("UpdateProject() Description = %q, want %q", retrieved.Description.String, "Updated description")
	}
}

func TestSQLDataStore_DeleteProject(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{
		Name: "delete-project",
		Path: "/to/delete",
	}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	if err := ds.DeleteProject("delete-project"); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	// Verify deletion
	_, err := ds.GetProjectByName("delete-project")
	if err == nil {
		t.Errorf("DeleteProject() project should not exist after deletion")
	}
}

func TestSQLDataStore_ListProjects(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create multiple projects
	for i := 1; i <= 3; i++ {
		project := &models.Project{
			Name: "list-project-" + string(rune('0'+i)),
			Path: "/path/" + string(rune('0'+i)),
		}
		if err := ds.CreateProject(project); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	projects, err := ds.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	if len(projects) != 3 {
		t.Errorf("ListProjects() returned %d projects, want 3", len(projects))
	}
}

// =============================================================================
// Workspace CRUD Tests
// =============================================================================

func TestSQLDataStore_CreateWorkspace(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create a project first
	project := &models.Project{Name: "ws-project", Path: "/ws/path"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		ProjectID: project.ID,
		Name:      "test-workspace",
		ImageName: "test:latest",
		Status:    "stopped",
	}

	err := ds.CreateWorkspace(workspace)
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}

	if workspace.ID == 0 {
		t.Errorf("CreateWorkspace() did not set workspace.ID")
	}
}

func TestSQLDataStore_GetWorkspaceByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "ws-project2", Path: "/ws/path2"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		ProjectID: project.ID,
		Name:      "findme-ws",
		ImageName: "image:v1",
		Status:    "running",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetWorkspaceByName(project.ID, "findme-ws")
	if err != nil {
		t.Fatalf("GetWorkspaceByName() error = %v", err)
	}

	if retrieved.Name != "findme-ws" {
		t.Errorf("GetWorkspaceByName() Name = %q, want %q", retrieved.Name, "findme-ws")
	}
	if retrieved.ImageName != "image:v1" {
		t.Errorf("GetWorkspaceByName() ImageName = %q, want %q", retrieved.ImageName, "image:v1")
	}
}

func TestSQLDataStore_GetWorkspaceByID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "ws-project3", Path: "/ws/path3"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		ProjectID: project.ID,
		Name:      "getbyid-ws",
		ImageName: "test:v2",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspaceByID() error = %v", err)
	}

	if retrieved.ID != workspace.ID {
		t.Errorf("GetWorkspaceByID() ID = %d, want %d", retrieved.ID, workspace.ID)
	}
}

func TestSQLDataStore_UpdateWorkspace(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "ws-project4", Path: "/ws/path4"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		ProjectID: project.ID,
		Name:      "update-ws",
		ImageName: "old:image",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Update
	workspace.ImageName = "new:image"
	workspace.Status = "running"

	if err := ds.UpdateWorkspace(workspace); err != nil {
		t.Fatalf("UpdateWorkspace() error = %v", err)
	}

	// Verify
	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	if err != nil {
		t.Fatalf("Verification error: %v", err)
	}

	if retrieved.ImageName != "new:image" {
		t.Errorf("UpdateWorkspace() ImageName = %q, want %q", retrieved.ImageName, "new:image")
	}
	if retrieved.Status != "running" {
		t.Errorf("UpdateWorkspace() Status = %q, want %q", retrieved.Status, "running")
	}
}

func TestSQLDataStore_DeleteWorkspace(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "ws-project5", Path: "/ws/path5"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		ProjectID: project.ID,
		Name:      "delete-ws",
		ImageName: "img:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	if err := ds.DeleteWorkspace(workspace.ID); err != nil {
		t.Fatalf("DeleteWorkspace() error = %v", err)
	}

	_, err := ds.GetWorkspaceByID(workspace.ID)
	if err == nil {
		t.Errorf("DeleteWorkspace() workspace should not exist after deletion")
	}
}

func TestSQLDataStore_ListWorkspacesByProject(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "ws-project6", Path: "/ws/path6"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create multiple workspaces
	for i := 1; i <= 3; i++ {
		ws := &models.Workspace{
			ProjectID: project.ID,
			Name:      "ws-" + string(rune('0'+i)),
			ImageName: "img:v" + string(rune('0'+i)),
			Status:    "stopped",
		}
		if err := ds.CreateWorkspace(ws); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	workspaces, err := ds.ListWorkspacesByProject(project.ID)
	if err != nil {
		t.Fatalf("ListWorkspacesByProject() error = %v", err)
	}

	if len(workspaces) != 3 {
		t.Errorf("ListWorkspacesByProject() returned %d workspaces, want 3", len(workspaces))
	}
}

func TestSQLDataStore_ListAllWorkspaces(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create two projects with workspaces
	for p := 1; p <= 2; p++ {
		project := &models.Project{
			Name: "project-" + string(rune('0'+p)),
			Path: "/path/" + string(rune('0'+p)),
		}
		if err := ds.CreateProject(project); err != nil {
			t.Fatalf("Setup error: %v", err)
		}

		for w := 1; w <= 2; w++ {
			ws := &models.Workspace{
				ProjectID: project.ID,
				Name:      "ws-" + string(rune('0'+w)),
				ImageName: "img:latest",
				Status:    "stopped",
			}
			if err := ds.CreateWorkspace(ws); err != nil {
				t.Fatalf("Setup error: %v", err)
			}
		}
	}

	workspaces, err := ds.ListAllWorkspaces()
	if err != nil {
		t.Fatalf("ListAllWorkspaces() error = %v", err)
	}

	if len(workspaces) != 4 {
		t.Errorf("ListAllWorkspaces() returned %d workspaces, want 4", len(workspaces))
	}
}

// =============================================================================
// Context Tests
// =============================================================================

func TestSQLDataStore_Context(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create a project and workspace
	project := &models.Project{Name: "ctx-project", Path: "/ctx/path"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		ProjectID: project.ID,
		Name:      "ctx-workspace",
		ImageName: "ctx:img",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Set active project
	if err := ds.SetActiveProject(&project.ID); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}

	// Get context and verify
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveProjectID == nil || *ctx.ActiveProjectID != project.ID {
		t.Errorf("GetContext() ActiveProjectID = %v, want %d", ctx.ActiveProjectID, project.ID)
	}

	// Set active workspace
	if err := ds.SetActiveWorkspace(&workspace.ID); err != nil {
		t.Fatalf("SetActiveWorkspace() error = %v", err)
	}

	ctx, err = ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveWorkspaceID == nil || *ctx.ActiveWorkspaceID != workspace.ID {
		t.Errorf("GetContext() ActiveWorkspaceID = %v, want %d", ctx.ActiveWorkspaceID, workspace.ID)
	}
}

func TestSQLDataStore_ClearContext(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "clear-ctx-project", Path: "/clear"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Set then clear
	if err := ds.SetActiveProject(&project.ID); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}

	if err := ds.SetActiveProject(nil); err != nil {
		t.Fatalf("SetActiveProject(nil) error = %v", err)
	}

	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveProjectID != nil {
		t.Errorf("SetActiveProject(nil) should clear ActiveProjectID")
	}
}

// =============================================================================
// Plugin Tests
// =============================================================================

func TestSQLDataStore_CreatePlugin(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	plugin := &models.NvimPluginDB{
		Name:     "test-plugin",
		Repo:     "user/test-plugin",
		Category: sql.NullString{String: "editor", Valid: true},
		Enabled:  true,
	}

	err := ds.CreatePlugin(plugin)
	if err != nil {
		t.Fatalf("CreatePlugin() error = %v", err)
	}

	if plugin.ID == 0 {
		t.Errorf("CreatePlugin() did not set plugin.ID")
	}
}

func TestSQLDataStore_GetPluginByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	plugin := &models.NvimPluginDB{
		Name:     "findme-plugin",
		Repo:     "user/findme-plugin",
		Category: sql.NullString{String: "lsp", Valid: true},
		Enabled:  true,
	}
	if err := ds.CreatePlugin(plugin); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetPluginByName("findme-plugin")
	if err != nil {
		t.Fatalf("GetPluginByName() error = %v", err)
	}

	if retrieved.Repo != "user/findme-plugin" {
		t.Errorf("GetPluginByName() Repo = %q, want %q", retrieved.Repo, "user/findme-plugin")
	}
}

func TestSQLDataStore_ListPlugins(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	for i := 1; i <= 3; i++ {
		plugin := &models.NvimPluginDB{
			Name:     "plugin-" + string(rune('0'+i)),
			Repo:     "user/plugin-" + string(rune('0'+i)),
			Category: sql.NullString{String: "test", Valid: true},
			Enabled:  true,
		}
		if err := ds.CreatePlugin(plugin); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	plugins, err := ds.ListPlugins()
	if err != nil {
		t.Fatalf("ListPlugins() error = %v", err)
	}

	if len(plugins) != 3 {
		t.Errorf("ListPlugins() returned %d plugins, want 3", len(plugins))
	}
}

func TestSQLDataStore_ListPluginsByCategory(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create plugins with different categories
	categories := []string{"lsp", "lsp", "ui", "editor"}
	for i, cat := range categories {
		plugin := &models.NvimPluginDB{
			Name:     "cat-plugin-" + string(rune('0'+i)),
			Repo:     "user/cat-plugin-" + string(rune('0'+i)),
			Category: sql.NullString{String: cat, Valid: true},
			Enabled:  true,
		}
		if err := ds.CreatePlugin(plugin); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	lspPlugins, err := ds.ListPluginsByCategory("lsp")
	if err != nil {
		t.Fatalf("ListPluginsByCategory() error = %v", err)
	}

	if len(lspPlugins) != 2 {
		t.Errorf("ListPluginsByCategory(lsp) returned %d plugins, want 2", len(lspPlugins))
	}
}

// =============================================================================
// Workspace-Plugin Association Tests
// =============================================================================

func TestSQLDataStore_WorkspacePluginAssociation(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Setup
	project := &models.Project{Name: "wp-project", Path: "/wp/path"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		ProjectID: project.ID,
		Name:      "wp-workspace",
		ImageName: "img:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	plugin := &models.NvimPluginDB{
		Name:    "wp-plugin",
		Repo:    "user/wp-plugin",
		Enabled: true,
	}
	if err := ds.CreatePlugin(plugin); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Add plugin to workspace
	if err := ds.AddPluginToWorkspace(workspace.ID, plugin.ID); err != nil {
		t.Fatalf("AddPluginToWorkspace() error = %v", err)
	}

	// Get workspace plugins
	plugins, err := ds.GetWorkspacePlugins(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspacePlugins() error = %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("GetWorkspacePlugins() returned %d plugins, want 1", len(plugins))
	}

	// Remove plugin from workspace
	if err := ds.RemovePluginFromWorkspace(workspace.ID, plugin.ID); err != nil {
		t.Fatalf("RemovePluginFromWorkspace() error = %v", err)
	}

	plugins, err = ds.GetWorkspacePlugins(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspacePlugins() after remove error = %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("GetWorkspacePlugins() after remove returned %d plugins, want 0", len(plugins))
	}
}

// =============================================================================
// Driver and Health Tests
// =============================================================================

func TestSQLDataStore_Driver(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	driver := ds.Driver()
	if driver == nil {
		t.Errorf("Driver() returned nil")
	}

	if driver.Type() != DriverMemory {
		t.Errorf("Driver().Type() = %v, want %v", driver.Type(), DriverMemory)
	}
}

func TestSQLDataStore_Ping(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	if err := ds.Ping(); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestSQLDataStore_Close(t *testing.T) {
	ds := createTestDataStore(t)

	if err := ds.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Subsequent operations should fail
	if err := ds.Ping(); err == nil {
		t.Errorf("Ping() after Close() should fail")
	}
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestSQLDataStore_ImplementsDataStore(t *testing.T) {
	var _ DataStore = (*SQLDataStore)(nil)
}
