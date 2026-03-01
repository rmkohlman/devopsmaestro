package handlers

import (
	"database/sql"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
)

// =============================================================================
// Handler Basic Tests
// =============================================================================

func TestRegistryHandler_Kind(t *testing.T) {
	h := NewRegistryHandler()
	if h.Kind() != KindRegistry {
		t.Errorf("Kind() = %v, want %v", h.Kind(), KindRegistry)
	}
}

// =============================================================================
// Handler Apply Tests
// =============================================================================

func TestRegistryHandler_Apply(t *testing.T) {
	h := NewRegistryHandler()
	ds := createTestDataStore(t)

	ctx := resource.Context{
		DataStore: ds,
	}

	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		errMsg  string
	}{
		{
			name: "create valid zot registry",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: my-zot
  description: OCI registry for images
spec:
  type: zot
  port: 5100
  lifecycle: persistent`,
			wantErr: false,
		},
		{
			name: "create valid athens registry",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: my-athens
spec:
  type: athens
  port: 3100
  lifecycle: on-demand`,
			wantErr: false,
		},
		{
			name: "create with invalid type",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: bad-registry
spec:
  type: invalid-type
  port: 5110`,
			wantErr: true,
			errMsg:  "unsupported registry type",
		},
		{
			name: "create with invalid port",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: bad-port
spec:
  type: zot
  port: 100`,
			wantErr: true,
			errMsg:  "port must be between 1024 and 65535",
		},
		{
			name: "create with missing type",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: no-type
spec:
  port: 5120`,
			wantErr: true,
			errMsg:  "type is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := h.Apply(ctx, []byte(tt.yaml))
			if tt.wantErr {
				if err == nil {
					t.Errorf("Apply() expected error but got none")
				} else if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("Apply() error = %v, want error containing %v", err, tt.errMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("Apply() error = %v", err)
				}
				if res.GetKind() != KindRegistry {
					t.Errorf("Apply() resource.Kind = %v, want %v", res.GetKind(), KindRegistry)
				}
			}
		})
	}
}

func TestRegistryHandler_Apply_Update(t *testing.T) {
	h := NewRegistryHandler()
	ds := createTestDataStore(t)

	ctx := resource.Context{
		DataStore: ds,
	}

	// Create initial registry
	createYAML := `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: update-test
spec:
  type: zot
  port: 5130
  lifecycle: manual`

	_, err := h.Apply(ctx, []byte(createYAML))
	if err != nil {
		t.Fatalf("Failed to create initial registry: %v", err)
	}

	// Update with different port
	updateYAML := `apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: update-test
spec:
  type: zot
  port: 5131
  lifecycle: persistent`

	res, err := h.Apply(ctx, []byte(updateYAML))
	if err != nil {
		t.Fatalf("Apply() update error = %v", err)
	}

	// Verify update
	regRes, ok := res.(*RegistryResource)
	if !ok {
		t.Fatalf("Apply() result is not *RegistryResource")
	}
	if regRes.Registry().Port != 5131 {
		t.Errorf("Apply() updated port = %v, want 5131", regRes.Registry().Port)
	}
	if regRes.Registry().Lifecycle != "persistent" {
		t.Errorf("Apply() updated lifecycle = %v, want persistent", regRes.Registry().Lifecycle)
	}
}

// =============================================================================
// Handler Get Tests
// =============================================================================

func TestRegistryHandler_Get(t *testing.T) {
	h := NewRegistryHandler()
	ds := createTestDataStore(t)

	// Pre-populate datastore
	reg := &models.Registry{
		Name:        "test-registry-get",
		Type:        "zot",
		Port:        5140,
		Lifecycle:   "persistent",
		Description: sql.NullString{String: "Test registry", Valid: true},
	}
	err := ds.CreateRegistry(reg)
	if err != nil {
		t.Fatalf("Failed to create test registry: %v", err)
	}

	ctx := resource.Context{
		DataStore: ds,
	}

	res, err := h.Get(ctx, "test-registry-get")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if res.GetName() != "test-registry-get" {
		t.Errorf("Get() resource.Name = %v, want test-registry-get", res.GetName())
	}

	// Cast to access underlying registry
	regRes, ok := res.(*RegistryResource)
	if !ok {
		t.Fatalf("Get() result is not *RegistryResource")
	}
	if regRes.Registry().Type != "zot" {
		t.Errorf("Get() registry.Type = %v, want zot", regRes.Registry().Type)
	}
}

func TestRegistryHandler_Get_NotFound(t *testing.T) {
	h := NewRegistryHandler()
	ds := createTestDataStore(t)

	ctx := resource.Context{
		DataStore: ds,
	}

	_, err := h.Get(ctx, "does-not-exist")
	if err == nil {
		t.Error("Get() expected error for non-existent registry")
	}
}

// =============================================================================
// Handler List Tests
// =============================================================================

func TestRegistryHandler_List(t *testing.T) {
	h := NewRegistryHandler()
	ds := createTestDataStore(t)

	// Pre-populate datastore
	registries := []*models.Registry{
		{Name: "zot-list-h1", Type: "zot", Port: 5150, Lifecycle: "persistent"},
		{Name: "athens-list-h1", Type: "athens", Port: 3150, Lifecycle: "on-demand"},
		{Name: "devpi-list-h1", Type: "devpi", Port: 3151, Lifecycle: "manual"},
	}

	for _, reg := range registries {
		err := ds.CreateRegistry(reg)
		if err != nil {
			t.Fatalf("Failed to create test registry: %v", err)
		}
	}

	ctx := resource.Context{
		DataStore: ds,
	}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(resources) != 3 {
		t.Errorf("List() returned %d resources, want 3", len(resources))
	}

	// Verify all are RegistryResource
	for _, res := range resources {
		if res.GetKind() != KindRegistry {
			t.Errorf("List() resource.Kind = %v, want %v", res.GetKind(), KindRegistry)
		}
	}
}

func TestRegistryHandler_List_Empty(t *testing.T) {
	h := NewRegistryHandler()
	ds := createTestDataStore(t)

	ctx := resource.Context{
		DataStore: ds,
	}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(resources) != 0 {
		t.Errorf("List() returned %d resources, want 0", len(resources))
	}
}

// =============================================================================
// Handler Delete Tests
// =============================================================================

func TestRegistryHandler_Delete(t *testing.T) {
	h := NewRegistryHandler()
	ds := createTestDataStore(t)

	// Pre-populate datastore
	reg := &models.Registry{
		Name:      "to-delete-h",
		Type:      "squid",
		Port:      3160,
		Lifecycle: "manual",
	}
	err := ds.CreateRegistry(reg)
	if err != nil {
		t.Fatalf("Failed to create test registry: %v", err)
	}

	ctx := resource.Context{
		DataStore: ds,
	}

	err = h.Delete(ctx, "to-delete-h")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	_, err = ds.GetRegistryByName("to-delete-h")
	if err == nil {
		t.Error("Delete() did not remove registry from datastore")
	}
}

func TestRegistryHandler_Delete_NotFound(t *testing.T) {
	h := NewRegistryHandler()
	ds := createTestDataStore(t)

	ctx := resource.Context{
		DataStore: ds,
	}

	err := h.Delete(ctx, "does-not-exist")
	if err == nil {
		t.Error("Delete() expected error for non-existent registry")
	}
}

// =============================================================================
// Handler ToYAML Tests
// =============================================================================

func TestRegistryHandler_ToYAML(t *testing.T) {
	h := NewRegistryHandler()

	reg := &models.Registry{
		Name:        "test-registry-yaml",
		Type:        "verdaccio",
		Port:        4873,
		Lifecycle:   "on-demand",
		Description: sql.NullString{String: "npm registry", Valid: true},
	}

	res := &RegistryResource{
		registry: reg,
	}

	yaml, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	// Verify YAML contains expected content
	yamlStr := string(yaml)
	if !contains(yamlStr, "kind: Registry") {
		t.Error("ToYAML() missing 'kind: Registry'")
	}
	if !contains(yamlStr, "name: test-registry-yaml") {
		t.Error("ToYAML() missing 'name: test-registry-yaml'")
	}
	if !contains(yamlStr, "type: verdaccio") {
		t.Error("ToYAML() missing 'type: verdaccio'")
	}
	if !contains(yamlStr, "port: 4873") {
		t.Error("ToYAML() missing 'port: 4873'")
	}
}

// =============================================================================
// RegistryResource Validation Tests
// =============================================================================

func TestRegistryResource_Validate(t *testing.T) {
	tests := []struct {
		name     string
		registry *models.Registry
		wantErr  bool
	}{
		{
			name:     "valid registry",
			registry: &models.Registry{Name: "test", Type: "zot", Port: 5170, Lifecycle: "persistent"},
			wantErr:  false,
		},
		{
			name:     "missing name",
			registry: &models.Registry{Name: "", Type: "zot", Port: 5171, Lifecycle: "persistent"},
			wantErr:  true,
		},
		{
			name:     "invalid type",
			registry: &models.Registry{Name: "test", Type: "invalid", Port: 5172, Lifecycle: "persistent"},
			wantErr:  true,
		},
		{
			name:     "invalid port",
			registry: &models.Registry{Name: "test", Type: "zot", Port: 100, Lifecycle: "persistent"},
			wantErr:  true,
		},
		{
			name:     "invalid lifecycle",
			registry: &models.Registry{Name: "test", Type: "zot", Port: 5173, Lifecycle: "always-on"},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &RegistryResource{registry: tt.registry}
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func createTestDataStore(t *testing.T) *db.SQLDataStore {
	t.Helper()

	cfg := db.DriverConfig{Type: db.DriverMemory}
	driver, err := db.NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("Failed to create test driver: %v", err)
	}

	if err := driver.Connect(); err != nil {
		t.Fatalf("Failed to connect test driver: %v", err)
	}

	// Create schema (including registries table)
	if err := createTestSchema(driver); err != nil {
		driver.Close()
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return db.NewSQLDataStore(driver, nil)
}

func createTestSchema(driver db.Driver) error {
	query := `CREATE TABLE IF NOT EXISTS registries (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		type TEXT NOT NULL,
		port INTEGER NOT NULL,
		lifecycle TEXT NOT NULL DEFAULT 'manual',
		description TEXT,
		config TEXT,
		status TEXT DEFAULT 'stopped',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`

	_, err := driver.Execute(query)
	return err
}
