package handlers

import (
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// =============================================================================
// EcosystemHandler Tests - Kind
// =============================================================================

func TestEcosystemHandler_Kind(t *testing.T) {
	h := NewEcosystemHandler()
	if h.Kind() != KindEcosystem {
		t.Errorf("Kind() = %q, want %q", h.Kind(), KindEcosystem)
	}
}

// =============================================================================
// EcosystemHandler Tests - Apply
// =============================================================================

func TestEcosystemHandler_Apply_Create(t *testing.T) {
	h := NewEcosystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: test-eco
  annotations:
    description: A test ecosystem
spec:
  theme: tokyonight
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if res.GetKind() != KindEcosystem {
		t.Errorf("Apply() resource.Kind = %q, want %q", res.GetKind(), KindEcosystem)
	}
	if res.GetName() != "test-eco" {
		t.Errorf("Apply() resource.Name = %q, want %q", res.GetName(), "test-eco")
	}

	// Verify stored in mock
	stored, err := store.GetEcosystemByName("test-eco")
	if err != nil {
		t.Fatalf("ecosystem not found in store: %v", err)
	}
	if stored.Name != "test-eco" {
		t.Errorf("stored Name = %q, want %q", stored.Name, "test-eco")
	}
}

func TestEcosystemHandler_Apply_Update(t *testing.T) {
	h := NewEcosystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// Pre-populate
	_ = store.CreateEcosystem(&models.Ecosystem{Name: "eco-update", ID: 0})

	updateYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: eco-update
  annotations:
    description: Updated description
spec: {}
`)

	res, err := h.Apply(ctx, updateYAML)
	if err != nil {
		t.Fatalf("Apply() update error = %v", err)
	}
	if res.GetName() != "eco-update" {
		t.Errorf("Apply() resource.Name = %q, want %q", res.GetName(), "eco-update")
	}

	// Verify description was updated
	ecoRes, ok := res.(*EcosystemResource)
	if !ok {
		t.Fatalf("result is not *EcosystemResource")
	}
	if !ecoRes.Ecosystem().Description.Valid || ecoRes.Ecosystem().Description.String != "Updated description" {
		t.Errorf("Apply() description = %v, want %q", ecoRes.Ecosystem().Description, "Updated description")
	}
}

func TestEcosystemHandler_Apply_InvalidYAML(t *testing.T) {
	h := NewEcosystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_, err := h.Apply(ctx, []byte("{\x00invalid: [yaml"))
	if err == nil {
		t.Error("Apply() expected error for invalid YAML, got nil")
	}
}

// =============================================================================
// EcosystemHandler Tests - Get
// =============================================================================

func TestEcosystemHandler_Get_Found(t *testing.T) {
	h := NewEcosystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_ = store.CreateEcosystem(&models.Ecosystem{Name: "eco-get"})

	res, err := h.Get(ctx, "eco-get")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if res.GetName() != "eco-get" {
		t.Errorf("Get() Name = %q, want %q", res.GetName(), "eco-get")
	}
	if res.GetKind() != KindEcosystem {
		t.Errorf("Get() Kind = %q, want %q", res.GetKind(), KindEcosystem)
	}
}

func TestEcosystemHandler_Get_NotFound(t *testing.T) {
	h := NewEcosystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_, err := h.Get(ctx, "does-not-exist")
	if err == nil {
		t.Error("Get() expected error for non-existent ecosystem, got nil")
	}
}

// =============================================================================
// EcosystemHandler Tests - List
// =============================================================================

func TestEcosystemHandler_List_Empty(t *testing.T) {
	h := NewEcosystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("List() returned %d items, want 0", len(resources))
	}
}

func TestEcosystemHandler_List_Multiple(t *testing.T) {
	h := NewEcosystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_ = store.CreateEcosystem(&models.Ecosystem{Name: "eco-a"})
	_ = store.CreateEcosystem(&models.Ecosystem{Name: "eco-b"})
	_ = store.CreateEcosystem(&models.Ecosystem{Name: "eco-c"})

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 3 {
		t.Errorf("List() returned %d items, want 3", len(resources))
	}
	for _, r := range resources {
		if r.GetKind() != KindEcosystem {
			t.Errorf("List() item Kind = %q, want %q", r.GetKind(), KindEcosystem)
		}
	}
}

// =============================================================================
// EcosystemHandler Tests - Delete
// =============================================================================

func TestEcosystemHandler_Delete_Found(t *testing.T) {
	h := NewEcosystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_ = store.CreateEcosystem(&models.Ecosystem{Name: "eco-delete"})

	if err := h.Delete(ctx, "eco-delete"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify removed
	_, err := store.GetEcosystemByName("eco-delete")
	if err == nil {
		t.Error("Delete() did not remove ecosystem from store")
	}
}

func TestEcosystemHandler_Delete_NotFound(t *testing.T) {
	h := NewEcosystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	err := h.Delete(ctx, "nonexistent")
	if err == nil {
		t.Error("Delete() expected error for non-existent ecosystem, got nil")
	}
}

// =============================================================================
// EcosystemHandler Tests - ToYAML
// =============================================================================

func TestEcosystemHandler_ToYAML(t *testing.T) {
	h := NewEcosystemHandler()

	eco := &models.Ecosystem{
		ID:   1,
		Name: "yaml-eco",
	}
	res := NewEcosystemResource(eco)

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	if !strings.Contains(yamlStr, "kind: Ecosystem") {
		t.Errorf("ToYAML() missing 'kind: Ecosystem', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "name: yaml-eco") {
		t.Errorf("ToYAML() missing 'name: yaml-eco', got:\n%s", yamlStr)
	}
}

func TestEcosystemHandler_ToYAML_WrongType(t *testing.T) {
	h := NewEcosystemHandler()
	// Pass a wrong resource type
	wrongRes := &DomainResource{domain: &models.Domain{Name: "wrong"}}
	_, err := h.ToYAML(wrongRes)
	if err == nil {
		t.Error("ToYAML() expected error for wrong resource type, got nil")
	}
}

// =============================================================================
// EcosystemResource Validate Tests
// =============================================================================

func TestEcosystemResource_Validate(t *testing.T) {
	tests := []struct {
		name    string
		eco     *models.Ecosystem
		wantErr bool
	}{
		{
			name:    "valid ecosystem",
			eco:     &models.Ecosystem{Name: "valid-eco"},
			wantErr: false,
		},
		{
			name:    "missing name",
			eco:     &models.Ecosystem{Name: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewEcosystemResource(tt.eco)
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
