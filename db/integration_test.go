package db

import (
	"context"
	"database/sql"
	"devopsmaestro/models"
	"testing"
	"time"
)

// =============================================================================
// Integration Tests - Testing Through Interfaces
//
// These tests verify that all layers work together correctly when accessed
// through their interfaces. This ensures the decoupling is working properly
// and that different implementations can be swapped without breaking behavior.
// =============================================================================

// =============================================================================
// Test Factory Functions (Interface Entry Points)
// =============================================================================

func TestIntegration_NewDriver_CreatesWorkingDriver(t *testing.T) {
	// Test that the factory creates a driver that works through the interface
	cfg := DriverConfig{Type: DriverMemory}

	// Use the factory function (entry point)
	driver, err := NewDriver(cfg)
	if err != nil {
		t.Fatalf("NewDriver() error = %v", err)
	}
	defer driver.Close()

	// Verify we can use it through the Driver interface
	if err := driver.Connect(); err != nil {
		t.Fatalf("Driver.Connect() error = %v", err)
	}

	if err := driver.Ping(); err != nil {
		t.Fatalf("Driver.Ping() error = %v", err)
	}

	// Execute through interface
	_, err = driver.Execute(`CREATE TABLE test_factory (id INTEGER PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("Driver.Execute() error = %v", err)
	}
}

func TestIntegration_NewDataStore_CreatesWorkingStore(t *testing.T) {
	// Create driver through factory
	driver, err := NewDriver(DriverConfig{Type: DriverMemory})
	if err != nil {
		t.Fatalf("NewDriver() error = %v", err)
	}
	driver.Connect()

	// Create DataStore through factory
	store, err := NewDataStore(DataStoreConfig{Driver: driver})
	if err != nil {
		t.Fatalf("NewDataStore() error = %v", err)
	}
	defer store.Close()

	// Verify we can use it through the DataStore interface
	if err := store.Ping(); err != nil {
		t.Fatalf("DataStore.Ping() error = %v", err)
	}

	// Verify Driver() returns the interface
	d := store.Driver()
	if d == nil {
		t.Fatalf("DataStore.Driver() returned nil")
	}
	if d.Type() != DriverMemory {
		t.Errorf("DataStore.Driver().Type() = %v, want %v", d.Type(), DriverMemory)
	}
}

// =============================================================================
// Full Stack Integration: DataStore -> Driver -> Database
// =============================================================================

func TestIntegration_FullStack_ProjectWorkflow(t *testing.T) {
	// This test exercises the full stack:
	// DataStore interface -> SQLDataStore -> Driver interface -> SQLiteDriver -> SQLite

	store := createIntegrationTestStore(t)
	defer store.Close()

	// --- Create Project (DataStore -> Driver -> DB) ---
	project := &models.Project{
		Name:        "integration-project",
		Path:        "/integration/path",
		Description: sql.NullString{String: "Integration test project", Valid: true},
	}

	err := store.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject() through interface error = %v", err)
	}

	if project.ID == 0 {
		t.Error("CreateProject() should set ID through all layers")
	}

	// --- Retrieve Project (validates data went through all layers) ---
	retrieved, err := store.GetProjectByName("integration-project")
	if err != nil {
		t.Fatalf("GetProjectByName() through interface error = %v", err)
	}

	if retrieved.Name != project.Name {
		t.Errorf("Data integrity: Name = %q, want %q", retrieved.Name, project.Name)
	}
	if retrieved.Path != project.Path {
		t.Errorf("Data integrity: Path = %q, want %q", retrieved.Path, project.Path)
	}

	// --- Update Project ---
	retrieved.Path = "/updated/path"
	if err := store.UpdateProject(retrieved); err != nil {
		t.Fatalf("UpdateProject() through interface error = %v", err)
	}

	updated, _ := store.GetProjectByID(retrieved.ID)
	if updated.Path != "/updated/path" {
		t.Errorf("Update did not propagate: Path = %q, want %q", updated.Path, "/updated/path")
	}

	// --- Delete Project ---
	if err := store.DeleteProject("integration-project"); err != nil {
		t.Fatalf("DeleteProject() through interface error = %v", err)
	}

	_, err = store.GetProjectByName("integration-project")
	if err == nil {
		t.Error("Delete did not propagate through layers")
	}
}

func TestIntegration_FullStack_WorkspaceWorkflow(t *testing.T) {
	store := createIntegrationTestStore(t)
	defer store.Close()

	// Create project first
	project := &models.Project{Name: "ws-int-project", Path: "/ws/int"}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// --- Create Workspace ---
	workspace := &models.Workspace{
		AppID:       project.ID, // Using project.ID as AppID during migration
		Name:        "integration-workspace",
		Description: sql.NullString{String: "Integration workspace", Valid: true},
		ImageName:   "integration:v1",
		Status:      "stopped",
	}

	if err := store.CreateWorkspace(workspace); err != nil {
		t.Fatalf("CreateWorkspace() through interface error = %v", err)
	}

	// --- Retrieve and verify ---
	retrieved, err := store.GetWorkspaceByName(project.ID, "integration-workspace")
	if err != nil {
		t.Fatalf("GetWorkspaceByName() through interface error = %v", err)
	}

	if retrieved.ImageName != "integration:v1" {
		t.Errorf("Data integrity: ImageName = %q, want %q", retrieved.ImageName, "integration:v1")
	}

	// --- Update ---
	retrieved.Status = "running"
	if err := store.UpdateWorkspace(retrieved); err != nil {
		t.Fatalf("UpdateWorkspace() error = %v", err)
	}

	updated, _ := store.GetWorkspaceByID(retrieved.ID)
	if updated.Status != "running" {
		t.Errorf("Update did not propagate: Status = %q, want %q", updated.Status, "running")
	}

	// --- List ---
	workspaces, err := store.ListWorkspacesByApp(project.ID)
	if err != nil {
		t.Fatalf("ListWorkspacesByApp() error = %v", err)
	}
	if len(workspaces) != 1 {
		t.Errorf("List returned %d workspaces, want 1", len(workspaces))
	}
}

func TestIntegration_FullStack_ContextWorkflow(t *testing.T) {
	store := createIntegrationTestStore(t)
	defer store.Close()

	// Setup
	project := &models.Project{Name: "ctx-int-project", Path: "/ctx/int"}
	store.CreateProject(project)

	workspace := &models.Workspace{
		AppID:     project.ID, // Using project.ID as AppID during migration
		Name:      "ctx-int-workspace",
		ImageName: "ctx:latest",
		Status:    "stopped",
	}
	store.CreateWorkspace(workspace)

	// --- Set and Get Context ---
	if err := store.SetActiveProject(&project.ID); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}

	if err := store.SetActiveWorkspace(&workspace.ID); err != nil {
		t.Fatalf("SetActiveWorkspace() error = %v", err)
	}

	ctx, err := store.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveProjectID == nil || *ctx.ActiveProjectID != project.ID {
		t.Errorf("Context.ActiveProjectID = %v, want %d", ctx.ActiveProjectID, project.ID)
	}
	if ctx.ActiveWorkspaceID == nil || *ctx.ActiveWorkspaceID != workspace.ID {
		t.Errorf("Context.ActiveWorkspaceID = %v, want %d", ctx.ActiveWorkspaceID, workspace.ID)
	}
}

// =============================================================================
// Transaction Integration Tests
// =============================================================================

func TestIntegration_Transaction_CommitPropagates(t *testing.T) {
	store := createIntegrationTestStore(t)
	defer store.Close()

	// Get driver through interface
	driver := store.Driver()

	// Begin transaction through interface
	tx, err := driver.Begin()
	if err != nil {
		t.Fatalf("Driver.Begin() error = %v", err)
	}

	// Execute within transaction
	_, err = tx.Execute(`INSERT INTO projects (name, path) VALUES (?, ?)`, "tx-project", "/tx/path")
	if err != nil {
		tx.Rollback()
		t.Fatalf("tx.Execute() error = %v", err)
	}

	// Commit
	if err := tx.Commit(); err != nil {
		t.Fatalf("tx.Commit() error = %v", err)
	}

	// Verify through DataStore interface
	project, err := store.GetProjectByName("tx-project")
	if err != nil {
		t.Fatalf("Transaction commit did not propagate: %v", err)
	}
	if project.Path != "/tx/path" {
		t.Errorf("Data mismatch after commit: Path = %q", project.Path)
	}
}

func TestIntegration_Transaction_RollbackPropagates(t *testing.T) {
	store := createIntegrationTestStore(t)
	defer store.Close()

	driver := store.Driver()

	tx, err := driver.Begin()
	if err != nil {
		t.Fatalf("Driver.Begin() error = %v", err)
	}

	// Insert data
	_, err = tx.Execute(`INSERT INTO projects (name, path) VALUES (?, ?)`, "rollback-project", "/rb/path")
	if err != nil {
		tx.Rollback()
		t.Fatalf("tx.Execute() error = %v", err)
	}

	// Rollback
	if err := tx.Rollback(); err != nil {
		t.Fatalf("tx.Rollback() error = %v", err)
	}

	// Verify data was NOT persisted
	_, err = store.GetProjectByName("rollback-project")
	if err == nil {
		t.Error("Transaction rollback did not propagate - data should not exist")
	}
}

func TestIntegration_Transaction_WithContext(t *testing.T) {
	store := createIntegrationTestStore(t)
	defer store.Close()

	driver := store.Driver()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := driver.BeginContext(ctx)
	if err != nil {
		t.Fatalf("Driver.BeginContext() error = %v", err)
	}

	_, err = tx.Execute(`INSERT INTO projects (name, path) VALUES (?, ?)`, "ctx-tx-project", "/ctx/tx")
	if err != nil {
		tx.Rollback()
		t.Fatalf("tx.Execute() error = %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("tx.Commit() error = %v", err)
	}

	// Verify
	project, err := store.GetProjectByName("ctx-tx-project")
	if err != nil {
		t.Fatalf("Context transaction did not work: %v", err)
	}
	if project.Name != "ctx-tx-project" {
		t.Error("Data integrity issue with context transaction")
	}
}

// =============================================================================
// QueryBuilder Integration Tests
// =============================================================================

func TestIntegration_QueryBuilder_DialectCorrect(t *testing.T) {
	store := createIntegrationTestStore(t)
	defer store.Close()

	// The SQLDataStore should use SQLite query builder for memory driver
	// This is verified by the fact that queries work correctly

	// Create project - uses Now() from query builder
	project := &models.Project{
		Name: "qb-test",
		Path: "/qb/path",
	}

	if err := store.CreateProject(project); err != nil {
		t.Fatalf("Query with dialect-specific NOW() failed: %v", err)
	}

	// Retrieve and check timestamps were set
	retrieved, _ := store.GetProjectByName("qb-test")
	if retrieved.CreatedAt.IsZero() {
		t.Error("QueryBuilder.Now() did not work - CreatedAt is zero")
	}
}

// =============================================================================
// Interface Swappability Tests
// =============================================================================

func TestIntegration_InterfaceSwappability_MockDriver(t *testing.T) {
	// Verify that MockDriver can be used in place of SQLiteDriver
	// through the Driver interface

	mockDriver := NewMockDriver()
	mockDriver.TypeValue = DriverMemory

	// Create DataStore with mock driver
	store := NewSQLDataStore(mockDriver, NewSQLiteQueryBuilder())

	// The store should work with the mock through the interface
	driver := store.Driver()
	if driver.Type() != DriverMemory {
		t.Errorf("Interface swappability: Type() = %v, want %v", driver.Type(), DriverMemory)
	}

	// Verify method calls are recorded (mock-specific behavior)
	mockDriver.Ping()
	calls := mockDriver.GetCalls()
	found := false
	for _, c := range calls {
		if c.Method == "Ping" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Mock driver calls not being tracked through interface")
	}
}

func TestIntegration_InterfaceSwappability_MockDataStore(t *testing.T) {
	// Verify MockDataStore can be used in place of SQLDataStore
	// through the DataStore interface

	var store DataStore = NewMockDataStore()
	defer store.Close()

	// Should work through interface
	project := &models.Project{
		Name: "mock-project",
		Path: "/mock/path",
	}

	if err := store.CreateProject(project); err != nil {
		t.Fatalf("MockDataStore.CreateProject() through interface error = %v", err)
	}

	retrieved, err := store.GetProjectByName("mock-project")
	if err != nil {
		t.Fatalf("MockDataStore.GetProjectByName() through interface error = %v", err)
	}

	if retrieved.Name != "mock-project" {
		t.Errorf("Data through mock interface: Name = %q, want %q", retrieved.Name, "mock-project")
	}
}

// =============================================================================
// Error Propagation Tests
// =============================================================================

func TestIntegration_ErrorPropagation_DriverToStore(t *testing.T) {
	// Verify errors propagate correctly through layers
	store := createIntegrationTestStore(t)

	// Close the store (and driver)
	store.Close()

	// Operations should now fail and propagate errors
	err := store.Ping()
	if err == nil {
		t.Error("Error should propagate from closed driver through store")
	}
}

func TestIntegration_ErrorPropagation_NotFound(t *testing.T) {
	store := createIntegrationTestStore(t)
	defer store.Close()

	_, err := store.GetProjectByName("nonexistent-project")
	if err == nil {
		t.Error("Not found error should propagate through layers")
	}
}

// =============================================================================
// Concurrent Access Tests
// =============================================================================

func TestIntegration_ConcurrentAccess(t *testing.T) {
	store := createIntegrationTestStore(t)
	defer store.Close()

	// Create initial project
	project := &models.Project{Name: "concurrent-project", Path: "/concurrent"}
	store.CreateProject(project)

	// Concurrent reads
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			_, err := store.GetProjectByName("concurrent-project")
			if err != nil {
				t.Errorf("Concurrent read failed: %v", err)
			}
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// createIntegrationTestStore creates a fully-wired test store
func createIntegrationTestStore(t *testing.T) DataStore {
	t.Helper()

	// Create through factory (real entry point)
	driver, err := NewDriver(DriverConfig{Type: DriverMemory})
	if err != nil {
		t.Fatalf("NewDriver() error = %v", err)
	}

	if err := driver.Connect(); err != nil {
		t.Fatalf("driver.Connect() error = %v", err)
	}

	// Create schema
	if err := createIntegrationSchema(driver); err != nil {
		driver.Close()
		t.Fatalf("createIntegrationSchema() error = %v", err)
	}

	// Create DataStore through factory
	store, err := NewDataStore(DataStoreConfig{Driver: driver})
	if err != nil {
		driver.Close()
		t.Fatalf("NewDataStore() error = %v", err)
	}

	return store
}

// createIntegrationSchema creates the database schema for integration tests
func createIntegrationSchema(driver Driver) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS ecosystems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ecosystem_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id),
			UNIQUE(ecosystem_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS apps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			description TEXT,
			language TEXT,
			build_config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (domain_id) REFERENCES domains(id),
			UNIQUE(domain_id, name)
		)`,
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
			app_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			image_name TEXT,
			container_id TEXT,
			status TEXT DEFAULT 'stopped',
			nvim_structure TEXT,
			nvim_plugins TEXT,
			theme TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (app_id) REFERENCES apps(id),
			UNIQUE(app_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS context (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			active_ecosystem_id INTEGER,
			active_domain_id INTEGER,
			active_app_id INTEGER,
			active_workspace_id INTEGER,
			active_project_id INTEGER,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (active_ecosystem_id) REFERENCES ecosystems(id),
			FOREIGN KEY (active_domain_id) REFERENCES domains(id),
			FOREIGN KEY (active_app_id) REFERENCES apps(id),
			FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id),
			FOREIGN KEY (active_project_id) REFERENCES projects(id)
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
			PRIMARY KEY (workspace_id, plugin_id)
		)`,
		`INSERT OR IGNORE INTO context (id) VALUES (1)`,
	}

	for _, query := range queries {
		if _, err := driver.Execute(query); err != nil {
			return err
		}
	}

	return nil
}
