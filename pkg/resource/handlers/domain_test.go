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
// Helpers for domain tests
// =============================================================================

// setupDomainTest creates a mock store pre-populated with an ecosystem and
// sets it as the active ecosystem in the context. Returns the ecosystem ID.
func setupDomainTest(t *testing.T) (*db.MockDataStore, int) {
	t.Helper()
	store := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "test-eco"}
	if err := store.CreateEcosystem(eco); err != nil {
		t.Fatalf("failed to create ecosystem: %v", err)
	}
	// Set active ecosystem
	ecoID := eco.ID
	store.Context.ActiveEcosystemID = &ecoID
	return store, ecoID
}

// =============================================================================
// DomainHandler Tests - Kind
// =============================================================================

func TestDomainHandler_Kind(t *testing.T) {
	h := NewDomainHandler()
	if h.Kind() != KindDomain {
		t.Errorf("Kind() = %q, want %q", h.Kind(), KindDomain)
	}
}

// =============================================================================
// DomainHandler Tests - Apply
// =============================================================================

func TestDomainHandler_Apply_Create(t *testing.T) {
	h := NewDomainHandler()
	store, _ := setupDomainTest(t)
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: test-domain
  ecosystem: test-eco
  annotations:
    description: A test domain
spec: {}
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if res.GetKind() != KindDomain {
		t.Errorf("Apply() Kind = %q, want %q", res.GetKind(), KindDomain)
	}
	if res.GetName() != "test-domain" {
		t.Errorf("Apply() Name = %q, want %q", res.GetName(), "test-domain")
	}

	// Verify the ecosystem name is propagated
	dr, ok := res.(*DomainResource)
	if !ok {
		t.Fatalf("result is not *DomainResource")
	}
	if dr.EcosystemName() != "test-eco" {
		t.Errorf("EcosystemName() = %q, want %q", dr.EcosystemName(), "test-eco")
	}
}

func TestDomainHandler_Apply_Update(t *testing.T) {
	h := NewDomainHandler()
	store, ecoID := setupDomainTest(t)
	ctx := resource.Context{DataStore: store}

	// Pre-populate domain
	_ = store.CreateDomain(&models.Domain{Name: "domain-upd", EcosystemID: sql.NullInt64{Int64: int64(ecoID), Valid: true}})

	updateYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: domain-upd
  ecosystem: test-eco
  annotations:
    description: Updated
spec: {}
`)

	res, err := h.Apply(ctx, updateYAML)
	if err != nil {
		t.Fatalf("Apply() update error = %v", err)
	}
	if res.GetName() != "domain-upd" {
		t.Errorf("Apply() Name = %q, want %q", res.GetName(), "domain-upd")
	}

	dr, ok := res.(*DomainResource)
	if !ok {
		t.Fatalf("result is not *DomainResource")
	}
	if !dr.Domain().Description.Valid || dr.Domain().Description.String != "Updated" {
		t.Errorf("Apply() description = %v, want %q", dr.Domain().Description, "Updated")
	}
}

func TestDomainHandler_Apply_MissingEcosystem(t *testing.T) {
	h := NewDomainHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// YAML missing ecosystem field
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: orphan-domain
spec: {}
`)

	_, err := h.Apply(ctx, yamlData)
	if err == nil {
		t.Error("Apply() expected error for missing ecosystem, got nil")
	}
	if !strings.Contains(err.Error(), "ecosystem") {
		t.Errorf("Apply() error = %q, want it to mention 'ecosystem'", err.Error())
	}
}

func TestDomainHandler_Apply_EcosystemNotFound(t *testing.T) {
	h := NewDomainHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: some-domain
  ecosystem: nonexistent-eco
spec: {}
`)

	_, err := h.Apply(ctx, yamlData)
	if err == nil {
		t.Error("Apply() expected error for nonexistent ecosystem, got nil")
	}
}

func TestDomainHandler_Apply_InvalidYAML(t *testing.T) {
	h := NewDomainHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_, err := h.Apply(ctx, []byte("{\x00invalid: [yaml"))
	if err == nil {
		t.Error("Apply() expected error for invalid YAML, got nil")
	}
}

// =============================================================================
// DomainHandler Tests - Get
// =============================================================================

func TestDomainHandler_Get_Found(t *testing.T) {
	h := NewDomainHandler()
	store, ecoID := setupDomainTest(t)
	ctx := resource.Context{DataStore: store}

	_ = store.CreateDomain(&models.Domain{Name: "get-domain", EcosystemID: sql.NullInt64{Int64: int64(ecoID), Valid: true}})

	res, err := h.Get(ctx, "get-domain")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if res.GetName() != "get-domain" {
		t.Errorf("Get() Name = %q, want %q", res.GetName(), "get-domain")
	}
}

func TestDomainHandler_Get_NoActiveEcosystem(t *testing.T) {
	h := NewDomainHandler()
	store := db.NewMockDataStore() // no active ecosystem set
	ctx := resource.Context{DataStore: store}

	_, err := h.Get(ctx, "any-domain")
	if err == nil {
		t.Error("Get() expected error when no active ecosystem, got nil")
	}
	if !strings.Contains(err.Error(), "active ecosystem") {
		t.Errorf("Get() error = %q, want it to mention 'active ecosystem'", err.Error())
	}
}

func TestDomainHandler_Get_NotFound(t *testing.T) {
	h := NewDomainHandler()
	store, _ := setupDomainTest(t)
	ctx := resource.Context{DataStore: store}

	_, err := h.Get(ctx, "does-not-exist")
	if err == nil {
		t.Error("Get() expected error for non-existent domain, got nil")
	}
}

// =============================================================================
// DomainHandler Tests - List
// =============================================================================

func TestDomainHandler_List_WithActiveEcosystem(t *testing.T) {
	h := NewDomainHandler()
	store, ecoID := setupDomainTest(t)
	ctx := resource.Context{DataStore: store}

	_ = store.CreateDomain(&models.Domain{Name: "d1", EcosystemID: sql.NullInt64{Int64: int64(ecoID), Valid: true}})
	_ = store.CreateDomain(&models.Domain{Name: "d2", EcosystemID: sql.NullInt64{Int64: int64(ecoID), Valid: true}})

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("List() returned %d items, want 2", len(resources))
	}
	for _, r := range resources {
		if r.GetKind() != KindDomain {
			t.Errorf("List() item Kind = %q, want %q", r.GetKind(), KindDomain)
		}
	}
}

func TestDomainHandler_List_NoActiveEcosystem(t *testing.T) {
	h := NewDomainHandler()
	store := db.NewMockDataStore() // no active ecosystem
	ctx := resource.Context{DataStore: store}

	// Pre-populate two ecosystems with one domain each
	eco1 := &models.Ecosystem{Name: "eco1"}
	eco2 := &models.Ecosystem{Name: "eco2"}
	_ = store.CreateEcosystem(eco1)
	_ = store.CreateEcosystem(eco2)
	_ = store.CreateDomain(&models.Domain{Name: "d-eco1", EcosystemID: sql.NullInt64{Int64: int64(eco1.ID), Valid: true}})
	_ = store.CreateDomain(&models.Domain{Name: "d-eco2", EcosystemID: sql.NullInt64{Int64: int64(eco2.ID), Valid: true}})

	// No active ecosystem: should call ListAllDomains
	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("List() returned %d items, want 2", len(resources))
	}
}

// =============================================================================
// DomainHandler Tests - Delete
// =============================================================================

func TestDomainHandler_Delete_Found(t *testing.T) {
	h := NewDomainHandler()
	store, ecoID := setupDomainTest(t)
	ctx := resource.Context{DataStore: store}

	domain := &models.Domain{Name: "del-domain", EcosystemID: sql.NullInt64{Int64: int64(ecoID), Valid: true}}
	_ = store.CreateDomain(domain)

	if err := h.Delete(ctx, "del-domain"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify removed
	_, err := store.GetDomainByName(sql.NullInt64{Int64: int64(ecoID), Valid: true}, "del-domain")
	if err == nil {
		t.Error("Delete() did not remove domain from store")
	}
}

func TestDomainHandler_Delete_NoActiveEcosystem(t *testing.T) {
	h := NewDomainHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	err := h.Delete(ctx, "any-domain")
	if err == nil {
		t.Error("Delete() expected error when no active ecosystem, got nil")
	}
	if !strings.Contains(err.Error(), "active ecosystem") {
		t.Errorf("Delete() error = %q, want it to mention 'active ecosystem'", err.Error())
	}
}

// =============================================================================
// DomainHandler Tests - ToYAML
// =============================================================================

func TestDomainHandler_ToYAML(t *testing.T) {
	h := NewDomainHandler()

	domain := &models.Domain{
		ID:          1,
		EcosystemID: sql.NullInt64{Int64: 1, Valid: true},
		Name:        "yaml-domain",
	}
	res := NewDomainResource(domain, "my-eco")

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	if !strings.Contains(yamlStr, "kind: Domain") {
		t.Errorf("ToYAML() missing 'kind: Domain', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "name: yaml-domain") {
		t.Errorf("ToYAML() missing 'name: yaml-domain', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "ecosystem: my-eco") {
		t.Errorf("ToYAML() missing 'ecosystem: my-eco', got:\n%s", yamlStr)
	}
}

func TestDomainHandler_ToYAML_WrongType(t *testing.T) {
	h := NewDomainHandler()
	wrongRes := NewEcosystemResource(&models.Ecosystem{Name: "wrong"})
	_, err := h.ToYAML(wrongRes)
	if err == nil {
		t.Error("ToYAML() expected error for wrong resource type, got nil")
	}
}

// =============================================================================
// DomainResource Validate Tests
// =============================================================================

func TestDomainResource_Validate(t *testing.T) {
	tests := []struct {
		name    string
		domain  *models.Domain
		wantErr bool
	}{
		{
			name:    "valid domain",
			domain:  &models.Domain{Name: "good-domain", EcosystemID: sql.NullInt64{Int64: 1, Valid: true}},
			wantErr: false,
		},
		{
			name:    "missing name",
			domain:  &models.Domain{Name: "", EcosystemID: sql.NullInt64{Int64: 1, Valid: true}},
			wantErr: true,
		},
		{
			name:    "no ecosystem_id is allowed",
			domain:  &models.Domain{Name: "no-eco"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewDomainResource(tt.domain, "")
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
