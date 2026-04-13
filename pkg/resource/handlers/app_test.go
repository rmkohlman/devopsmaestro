package handlers

import (
	"database/sql"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// =============================================================================
// Helpers for app tests
// =============================================================================

// setupAppTest creates a mock store pre-populated with an ecosystem and domain,
// and sets the active ecosystem + domain IDs on the context.
// Returns the store, ecosystemID, and domainID.
func setupAppTest(t *testing.T) (*db.MockDataStore, int, int) {
	t.Helper()
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "app-eco"}
	if err := store.CreateEcosystem(eco); err != nil {
		t.Fatalf("failed to create ecosystem: %v", err)
	}

	domain := &models.Domain{Name: "app-domain", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	if err := store.CreateDomain(domain); err != nil {
		t.Fatalf("failed to create domain: %v", err)
	}

	ecoID := eco.ID
	domainID := domain.ID
	store.Context.ActiveEcosystemID = &ecoID
	store.Context.ActiveDomainID = &domainID

	return store, ecoID, domainID
}

// =============================================================================
// AppHandler Tests - Kind
// =============================================================================

func TestAppHandler_Kind(t *testing.T) {
	h := NewAppHandler()
	if h.Kind() != KindApp {
		t.Errorf("Kind() = %q, want %q", h.Kind(), KindApp)
	}
}

// =============================================================================
// AppHandler Tests - Apply
// =============================================================================

func TestAppHandler_Apply_Create(t *testing.T) {
	h := NewAppHandler()
	store, _, _ := setupAppTest(t)
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: test-app
  domain: app-domain
spec:
  path: /home/user/myapp
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if res.GetKind() != KindApp {
		t.Errorf("Apply() Kind = %q, want %q", res.GetKind(), KindApp)
	}
	if res.GetName() != "test-app" {
		t.Errorf("Apply() Name = %q, want %q", res.GetName(), "test-app")
	}

	ar, ok := res.(*AppResource)
	if !ok {
		t.Fatalf("result is not *AppResource")
	}
	if ar.DomainName() != "app-domain" {
		t.Errorf("DomainName() = %q, want %q", ar.DomainName(), "app-domain")
	}
	if ar.App().Path != "/home/user/myapp" {
		t.Errorf("App().Path = %q, want %q", ar.App().Path, "/home/user/myapp")
	}
}

func TestAppHandler_Apply_Update(t *testing.T) {
	h := NewAppHandler()
	store, _, domainID := setupAppTest(t)
	ctx := resource.Context{DataStore: store}

	// Pre-populate app
	_ = store.CreateApp(&models.App{Name: "app-upd", DomainID: sql.NullInt64{Int64: int64(domainID), Valid: true}, Path: "/old/path"})

	updateYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: app-upd
  domain: app-domain
spec:
  path: /new/path
`)

	res, err := h.Apply(ctx, updateYAML)
	if err != nil {
		t.Fatalf("Apply() update error = %v", err)
	}
	if res.GetName() != "app-upd" {
		t.Errorf("Apply() Name = %q, want %q", res.GetName(), "app-upd")
	}

	ar, ok := res.(*AppResource)
	if !ok {
		t.Fatalf("result is not *AppResource")
	}
	if ar.App().Path != "/new/path" {
		t.Errorf("Apply() Path = %q, want %q", ar.App().Path, "/new/path")
	}
}

func TestAppHandler_Apply_MissingDomain(t *testing.T) {
	h := NewAppHandler()
	store, _, _ := setupAppTest(t)
	ctx := resource.Context{DataStore: store}

	// YAML missing domain metadata
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: orphan-app
spec:
  path: /some/path
`)

	_, err := h.Apply(ctx, yamlData)
	if err == nil {
		t.Error("Apply() expected error for missing domain, got nil")
	}
	if !strings.Contains(err.Error(), "domain") {
		t.Errorf("Apply() error = %q, want it to mention 'domain'", err.Error())
	}
}

func TestAppHandler_Apply_DomainNotFound(t *testing.T) {
	h := NewAppHandler()
	store, _, _ := setupAppTest(t)
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: some-app
  domain: nonexistent-domain
spec:
  path: /some/path
`)

	_, err := h.Apply(ctx, yamlData)
	if err == nil {
		t.Error("Apply() expected error for nonexistent domain, got nil")
	}
}

func TestAppHandler_Apply_InvalidYAML(t *testing.T) {
	h := NewAppHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_, err := h.Apply(ctx, []byte("{\x00invalid: [yaml"))
	if err == nil {
		t.Error("Apply() expected error for invalid YAML, got nil")
	}
}

// =============================================================================
// AppHandler Tests - Get
// =============================================================================

func TestAppHandler_Get_Found(t *testing.T) {
	h := NewAppHandler()
	store, _, domainID := setupAppTest(t)
	ctx := resource.Context{DataStore: store}

	_ = store.CreateApp(&models.App{Name: "get-app", DomainID: sql.NullInt64{Int64: int64(domainID), Valid: true}, Path: "/get/path"})

	res, err := h.Get(ctx, "get-app")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if res.GetName() != "get-app" {
		t.Errorf("Get() Name = %q, want %q", res.GetName(), "get-app")
	}
}

func TestAppHandler_Get_NoActiveDomain(t *testing.T) {
	h := NewAppHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_, err := h.Get(ctx, "any-app")
	if err == nil {
		t.Error("Get() expected error when no active domain, got nil")
	}
	if !strings.Contains(err.Error(), "active domain") {
		t.Errorf("Get() error = %q, want it to mention 'active domain'", err.Error())
	}
}

func TestAppHandler_Get_NotFound(t *testing.T) {
	h := NewAppHandler()
	store, _, _ := setupAppTest(t)
	ctx := resource.Context{DataStore: store}

	_, err := h.Get(ctx, "does-not-exist")
	if err == nil {
		t.Error("Get() expected error for non-existent app, got nil")
	}
}

// =============================================================================
// AppHandler Tests - List
// =============================================================================

func TestAppHandler_List_WithActiveDomain(t *testing.T) {
	h := NewAppHandler()
	store, _, domainID := setupAppTest(t)
	ctx := resource.Context{DataStore: store}

	_ = store.CreateApp(&models.App{Name: "a1", DomainID: sql.NullInt64{Int64: int64(domainID), Valid: true}, Path: "/a1"})
	_ = store.CreateApp(&models.App{Name: "a2", DomainID: sql.NullInt64{Int64: int64(domainID), Valid: true}, Path: "/a2"})

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("List() returned %d items, want 2", len(resources))
	}
	for _, r := range resources {
		if r.GetKind() != KindApp {
			t.Errorf("List() item Kind = %q, want %q", r.GetKind(), KindApp)
		}
	}
}

func TestAppHandler_List_NoActiveDomain(t *testing.T) {
	h := NewAppHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// Pre-populate across two domains
	d1 := &models.Domain{Name: "d1", EcosystemID: sql.NullInt64{Int64: 1, Valid: true}}
	d2 := &models.Domain{Name: "d2", EcosystemID: sql.NullInt64{Int64: 1, Valid: true}}
	_ = store.CreateDomain(d1)
	_ = store.CreateDomain(d2)
	_ = store.CreateApp(&models.App{Name: "app-d1", DomainID: sql.NullInt64{Int64: int64(d1.ID), Valid: true}, Path: "/p1"})
	_ = store.CreateApp(&models.App{Name: "app-d2", DomainID: sql.NullInt64{Int64: int64(d2.ID), Valid: true}, Path: "/p2"})

	// No active domain: should call ListAllApps
	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("List() returned %d items, want 2", len(resources))
	}
}

// =============================================================================
// AppHandler Tests - Delete
// =============================================================================

func TestAppHandler_Delete_Found(t *testing.T) {
	h := NewAppHandler()
	store, _, domainID := setupAppTest(t)
	ctx := resource.Context{DataStore: store}

	app := &models.App{Name: "del-app", DomainID: sql.NullInt64{Int64: int64(domainID), Valid: true}, Path: "/del"}
	_ = store.CreateApp(app)

	if err := h.Delete(ctx, "del-app"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify removed
	_, err := store.GetAppByName(sql.NullInt64{Int64: int64(domainID), Valid: true}, "del-app")
	if err == nil {
		t.Error("Delete() did not remove app from store")
	}
}

func TestAppHandler_Delete_NoActiveDomain(t *testing.T) {
	h := NewAppHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	err := h.Delete(ctx, "any-app")
	if err == nil {
		t.Error("Delete() expected error when no active domain, got nil")
	}
	if !strings.Contains(err.Error(), "active domain") {
		t.Errorf("Delete() error = %q, want it to mention 'active domain'", err.Error())
	}
}

// =============================================================================
// AppHandler Tests - ToYAML
// =============================================================================

func TestAppHandler_ToYAML(t *testing.T) {
	h := NewAppHandler()

	app := &models.App{
		ID:       1,
		DomainID: sql.NullInt64{Int64: 1, Valid: true},
		Name:     "yaml-app",
		Path:     "/my/path",
	}
	res := NewAppResource(app, "my-domain", "")

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	if !strings.Contains(yamlStr, "kind: App") {
		t.Errorf("ToYAML() missing 'kind: App', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "name: yaml-app") {
		t.Errorf("ToYAML() missing 'name: yaml-app', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "domain: my-domain") {
		t.Errorf("ToYAML() missing 'domain: my-domain', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "path: /my/path") {
		t.Errorf("ToYAML() missing 'path: /my/path', got:\n%s", yamlStr)
	}
}

func TestAppHandler_ToYAML_WrongType(t *testing.T) {
	h := NewAppHandler()
	wrongRes := NewEcosystemResource(&models.Ecosystem{Name: "wrong"})
	_, err := h.ToYAML(wrongRes)
	if err == nil {
		t.Error("ToYAML() expected error for wrong resource type, got nil")
	}
}

// =============================================================================
// AppResource Validate Tests
// =============================================================================

func TestAppResource_Validate(t *testing.T) {
	tests := []struct {
		name    string
		app     *models.App
		wantErr bool
	}{
		{
			name:    "valid app",
			app:     &models.App{Name: "good-app", DomainID: sql.NullInt64{Int64: 1, Valid: true}, Path: "/valid/path"},
			wantErr: false,
		},
		{
			name:    "missing name",
			app:     &models.App{Name: "", DomainID: sql.NullInt64{Int64: 1, Valid: true}, Path: "/valid/path"},
			wantErr: true,
		},
		{
			name:    "no domain_id is allowed",
			app:     &models.App{Name: "no-domain", Path: "/valid/path"},
			wantErr: false,
		},
		{
			name:    "missing path",
			app:     &models.App{Name: "no-path", DomainID: sql.NullInt64{Int64: 1, Valid: true}, Path: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewAppResource(tt.app, "", "")
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
