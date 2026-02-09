package db

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"devopsmaestro/models"
)

// =============================================================================
// MockDriver Tests
// =============================================================================

func TestNewMockDriver(t *testing.T) {
	driver := NewMockDriver()

	if driver == nil {
		t.Fatal("NewMockDriver() returned nil")
	}

	if driver.Type() != DriverMemory {
		t.Errorf("Type() = %v, want %v", driver.Type(), DriverMemory)
	}

	if driver.DSN() != "mock://test" {
		t.Errorf("DSN() = %q, want %q", driver.DSN(), "mock://test")
	}
}

func TestMockDriver_Connect(t *testing.T) {
	driver := NewMockDriver()

	// Default should succeed
	if err := driver.Connect(); err != nil {
		t.Errorf("Connect() error = %v, want nil", err)
	}

	// Verify call was recorded
	calls := driver.GetCalls()
	if len(calls) != 1 || calls[0].Method != "Connect" {
		t.Errorf("Expected Connect call to be recorded, got %v", calls)
	}

	// Test with custom error
	expectedErr := errors.New("connection failed")
	driver.ConnectFunc = func() error { return expectedErr }
	driver.ResetCalls()

	if err := driver.Connect(); err != expectedErr {
		t.Errorf("Connect() error = %v, want %v", err, expectedErr)
	}
}

func TestMockDriver_Execute(t *testing.T) {
	driver := NewMockDriver()

	result, err := driver.Execute("INSERT INTO test VALUES (?)", 1)
	if err != nil {
		t.Errorf("Execute() error = %v, want nil", err)
	}
	if result == nil {
		t.Error("Execute() returned nil result")
	}

	// Verify call was recorded with args
	calls := driver.GetCalls()
	if len(calls) != 1 {
		t.Fatalf("Expected 1 call, got %d", len(calls))
	}
	if calls[0].Method != "Execute" {
		t.Errorf("Method = %q, want %q", calls[0].Method, "Execute")
	}
}

func TestMockDriver_QueryRow(t *testing.T) {
	driver := NewMockDriver()

	// Set up custom row
	driver.QueryRowFunc = func(query string, args ...interface{}) Row {
		return &MockRow{
			Values: []interface{}{1, "test"},
		}
	}

	row := driver.QueryRow("SELECT id, name FROM test WHERE id = ?", 1)
	if row == nil {
		t.Fatal("QueryRow() returned nil")
	}

	var id int
	var name string
	if err := row.Scan(&id, &name); err != nil {
		t.Errorf("Scan() error = %v", err)
	}
	if id != 1 || name != "test" {
		t.Errorf("Scan() = (%d, %q), want (1, \"test\")", id, name)
	}
}

func TestMockDriver_Query(t *testing.T) {
	driver := NewMockDriver()

	// Set up custom rows
	driver.QueryFunc = func(query string, args ...interface{}) (Rows, error) {
		return &MockRows{
			Data: [][]interface{}{
				{1, "first"},
				{2, "second"},
			},
			ColumnList: []string{"id", "name"},
		}, nil
	}

	rows, err := driver.Query("SELECT id, name FROM test")
	if err != nil {
		t.Fatalf("Query() error = %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		count++
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			t.Errorf("Scan() error = %v", err)
		}
	}

	if count != 2 {
		t.Errorf("Expected 2 rows, got %d", count)
	}
}

func TestMockDriver_Transaction(t *testing.T) {
	driver := NewMockDriver()

	tx, err := driver.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	mockTx := tx.(*MockTransaction)

	// Execute within transaction
	_, err = tx.Execute("INSERT INTO test VALUES (?)", 1)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}

	// Commit
	if err := tx.Commit(); err != nil {
		t.Errorf("Commit() error = %v", err)
	}

	if !mockTx.Committed {
		t.Error("Transaction was not marked as committed")
	}
}

func TestMockDriver_TransactionRollback(t *testing.T) {
	driver := NewMockDriver()

	tx, err := driver.Begin()
	if err != nil {
		t.Fatalf("Begin() error = %v", err)
	}

	mockTx := tx.(*MockTransaction)

	if err := tx.Rollback(); err != nil {
		t.Errorf("Rollback() error = %v", err)
	}

	if !mockTx.RolledBack {
		t.Error("Transaction was not marked as rolled back")
	}
}

func TestMockDriver_ContextMethods(t *testing.T) {
	driver := NewMockDriver()
	ctx := context.Background()

	// ExecuteContext
	_, err := driver.ExecuteContext(ctx, "INSERT INTO test VALUES (?)", 1)
	if err != nil {
		t.Errorf("ExecuteContext() error = %v", err)
	}

	// QueryRowContext
	row := driver.QueryRowContext(ctx, "SELECT * FROM test WHERE id = ?", 1)
	if row == nil {
		t.Error("QueryRowContext() returned nil")
	}

	// QueryContext
	rows, err := driver.QueryContext(ctx, "SELECT * FROM test")
	if err != nil {
		t.Errorf("QueryContext() error = %v", err)
	}
	rows.Close()

	// BeginContext
	tx, err := driver.BeginContext(ctx)
	if err != nil {
		t.Errorf("BeginContext() error = %v", err)
	}
	tx.Rollback()
}

// =============================================================================
// MockResult Tests
// =============================================================================

func TestMockResult(t *testing.T) {
	result := &MockResult{
		LastID:       42,
		AffectedRows: 5,
	}

	id, err := result.LastInsertId()
	if err != nil || id != 42 {
		t.Errorf("LastInsertId() = (%d, %v), want (42, nil)", id, err)
	}

	rows, err := result.RowsAffected()
	if err != nil || rows != 5 {
		t.Errorf("RowsAffected() = (%d, %v), want (5, nil)", rows, err)
	}
}

func TestMockResult_WithErrors(t *testing.T) {
	expectedErr := errors.New("not supported")
	result := &MockResult{
		LastIDErr:       expectedErr,
		AffectedRowsErr: expectedErr,
	}

	_, err := result.LastInsertId()
	if err != expectedErr {
		t.Errorf("LastInsertId() error = %v, want %v", err, expectedErr)
	}

	_, err = result.RowsAffected()
	if err != expectedErr {
		t.Errorf("RowsAffected() error = %v, want %v", err, expectedErr)
	}
}

// =============================================================================
// MockRows Tests
// =============================================================================

func TestMockRows_Iteration(t *testing.T) {
	rows := &MockRows{
		Data: [][]interface{}{
			{1, "first"},
			{2, "second"},
			{3, "third"},
		},
		ColumnList: []string{"id", "name"},
	}

	count := 0
	for rows.Next() {
		count++
	}

	if count != 3 {
		t.Errorf("Iteration count = %d, want 3", count)
	}

	// After iteration, Next should return false
	if rows.Next() {
		t.Error("Next() should return false after iteration complete")
	}
}

func TestMockRows_Columns(t *testing.T) {
	rows := &MockRows{
		ColumnList: []string{"id", "name", "email"},
	}

	cols, err := rows.Columns()
	if err != nil {
		t.Errorf("Columns() error = %v", err)
	}
	if len(cols) != 3 {
		t.Errorf("Columns() returned %d columns, want 3", len(cols))
	}
}

func TestMockRows_CloseAndErr(t *testing.T) {
	rows := &MockRows{}

	if err := rows.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// After close, Next should return false
	if rows.Next() {
		t.Error("Next() should return false after Close()")
	}

	if err := rows.Err(); err != nil {
		t.Errorf("Err() = %v, want nil", err)
	}
}

// =============================================================================
// MockDataStore Tests
// =============================================================================

func TestNewMockDataStore(t *testing.T) {
	store := NewMockDataStore()

	if store == nil {
		t.Fatal("NewMockDataStore() returned nil")
	}

	if store.Projects == nil {
		t.Error("Projects map is nil")
	}
	if store.Workspaces == nil {
		t.Error("Workspaces map is nil")
	}
	if store.Plugins == nil {
		t.Error("Plugins map is nil")
	}
	if store.MockDriver == nil {
		t.Error("MockDriver is nil")
	}
}

func TestMockDataStore_ProjectCRUD(t *testing.T) {
	store := NewMockDataStore()

	// Create
	project := &models.Project{Name: "test-project", Path: "/test/path"}
	if err := store.CreateProject(project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}
	if project.ID == 0 {
		t.Error("Project ID was not set")
	}

	// Read by name
	retrieved, err := store.GetProjectByName("test-project")
	if err != nil {
		t.Fatalf("GetProjectByName() error = %v", err)
	}
	if retrieved.Path != "/test/path" {
		t.Errorf("Path = %q, want %q", retrieved.Path, "/test/path")
	}

	// Read by ID
	retrieved, err = store.GetProjectByID(project.ID)
	if err != nil {
		t.Fatalf("GetProjectByID() error = %v", err)
	}
	if retrieved.Name != "test-project" {
		t.Errorf("Name = %q, want %q", retrieved.Name, "test-project")
	}

	// Update
	project.Description = sql.NullString{String: "Updated description", Valid: true}
	if err := store.UpdateProject(project); err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}

	// List
	projects, err := store.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(projects) != 1 {
		t.Errorf("ListProjects() returned %d projects, want 1", len(projects))
	}

	// Delete
	if err := store.DeleteProject("test-project"); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	projects, _ = store.ListProjects()
	if len(projects) != 0 {
		t.Error("Project was not deleted")
	}
}

func TestMockDataStore_WorkspaceCRUD(t *testing.T) {
	store := NewMockDataStore()

	// Create project first
	project := &models.Project{Name: "test-project", Path: "/test"}
	store.CreateProject(project)

	// Create workspace
	workspace := &models.Workspace{
		AppID:  project.ID, // Using project.ID as AppID during migration
		Name:   "main",
		Status: "stopped",
	}
	if err := store.CreateWorkspace(workspace); err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}
	if workspace.ID == 0 {
		t.Error("Workspace ID was not set")
	}

	// Read by name
	retrieved, err := store.GetWorkspaceByName(project.ID, "main")
	if err != nil {
		t.Fatalf("GetWorkspaceByName() error = %v", err)
	}
	if retrieved.Status != "stopped" {
		t.Errorf("Status = %q, want %q", retrieved.Status, "stopped")
	}

	// Read by ID
	retrieved, err = store.GetWorkspaceByID(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspaceByID() error = %v", err)
	}

	// List by app (using project.ID as appID during migration)
	workspaces, err := store.ListWorkspacesByApp(project.ID)
	if err != nil {
		t.Fatalf("ListWorkspacesByApp() error = %v", err)
	}
	if len(workspaces) != 1 {
		t.Errorf("ListWorkspacesByApp() returned %d workspaces, want 1", len(workspaces))
	}

	// List all
	allWorkspaces, err := store.ListAllWorkspaces()
	if err != nil {
		t.Fatalf("ListAllWorkspaces() error = %v", err)
	}
	if len(allWorkspaces) != 1 {
		t.Errorf("ListAllWorkspaces() returned %d workspaces, want 1", len(allWorkspaces))
	}

	// Delete
	if err := store.DeleteWorkspace(workspace.ID); err != nil {
		t.Fatalf("DeleteWorkspace() error = %v", err)
	}
}

func TestMockDataStore_Context(t *testing.T) {
	store := NewMockDataStore()

	// Get initial context
	ctx, err := store.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}
	if ctx.ID != 1 {
		t.Errorf("Context ID = %d, want 1", ctx.ID)
	}

	// Set active project
	projectID := 42
	if err := store.SetActiveProject(&projectID); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}

	ctx, _ = store.GetContext()
	if ctx.ActiveProjectID == nil || *ctx.ActiveProjectID != 42 {
		t.Error("Active project ID was not set")
	}

	// Set active workspace
	workspaceID := 7
	if err := store.SetActiveWorkspace(&workspaceID); err != nil {
		t.Fatalf("SetActiveWorkspace() error = %v", err)
	}

	ctx, _ = store.GetContext()
	if ctx.ActiveWorkspaceID == nil || *ctx.ActiveWorkspaceID != 7 {
		t.Error("Active workspace ID was not set")
	}
}

func TestMockDataStore_ErrorInjection(t *testing.T) {
	store := NewMockDataStore()
	expectedErr := errors.New("injected error")

	store.CreateProjectErr = expectedErr
	if err := store.CreateProject(&models.Project{}); err != expectedErr {
		t.Errorf("CreateProject() error = %v, want %v", err, expectedErr)
	}

	store.GetProjectByNameErr = expectedErr
	if _, err := store.GetProjectByName("test"); err != expectedErr {
		t.Errorf("GetProjectByName() error = %v, want %v", err, expectedErr)
	}

	store.ListProjectsErr = expectedErr
	if _, err := store.ListProjects(); err != expectedErr {
		t.Errorf("ListProjects() error = %v, want %v", err, expectedErr)
	}
}

func TestMockDataStore_CallTracking(t *testing.T) {
	store := NewMockDataStore()

	// Make some calls
	store.CreateProject(&models.Project{Name: "test"})
	store.ListProjects()
	store.GetProjectByName("test")

	calls := store.GetCalls()
	if len(calls) != 3 {
		t.Errorf("Expected 3 calls, got %d", len(calls))
	}

	expectedMethods := []string{"CreateProject", "ListProjects", "GetProjectByName"}
	for i, expected := range expectedMethods {
		if calls[i].Method != expected {
			t.Errorf("Call %d method = %q, want %q", i, calls[i].Method, expected)
		}
	}

	// Reset calls
	store.ResetCalls()
	if len(store.GetCalls()) != 0 {
		t.Error("Calls were not reset")
	}
}

func TestMockDataStore_Reset(t *testing.T) {
	store := NewMockDataStore()

	// Add some data
	store.CreateProject(&models.Project{Name: "test"})
	store.CreateWorkspace(&models.Workspace{Name: "main"})

	// Reset
	store.Reset()

	// Verify all data is cleared
	if len(store.Projects) != 0 {
		t.Error("Projects not cleared")
	}
	if len(store.Workspaces) != 0 {
		t.Error("Workspaces not cleared")
	}
	if len(store.GetCalls()) != 0 {
		t.Error("Calls not cleared")
	}
}

func TestMockDataStore_DriverAccess(t *testing.T) {
	store := NewMockDataStore()

	driver := store.Driver()
	if driver == nil {
		t.Fatal("Driver() returned nil")
	}

	if driver.Type() != DriverMemory {
		t.Errorf("Driver type = %v, want %v", driver.Type(), DriverMemory)
	}
}

func TestMockDataStore_HealthMethods(t *testing.T) {
	store := NewMockDataStore()

	// Close
	if err := store.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Ping
	if err := store.Ping(); err != nil {
		t.Errorf("Ping() error = %v", err)
	}

	// With errors
	expectedErr := errors.New("closed")
	store.CloseErr = expectedErr
	if err := store.Close(); err != expectedErr {
		t.Errorf("Close() error = %v, want %v", err, expectedErr)
	}

	store.PingErr = expectedErr
	if err := store.Ping(); err != expectedErr {
		t.Errorf("Ping() error = %v, want %v", err, expectedErr)
	}
}

// Interface compliance tests
func TestMockDriver_ImplementsDriver(t *testing.T) {
	var _ Driver = (*MockDriver)(nil)
}

func TestMockDataStore_ImplementsDataStore(t *testing.T) {
	var _ DataStore = (*MockDataStore)(nil)
}
