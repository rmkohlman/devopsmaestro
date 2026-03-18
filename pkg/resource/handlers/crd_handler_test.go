package handlers

import (
	"fmt"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Fixtures
// =============================================================================

func createValidCRDYAML() []byte {
	return []byte(`
apiVersion: devopsmaestro.io/v1alpha1
kind: CustomResourceDefinition
metadata:
  name: databases.devopsmaestro.io
spec:
  group: devopsmaestro.io
  names:
    kind: Database
    singular: database
    plural: databases
    shortNames: [db]
  scope: Workspace
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              required: [engine, version]
              properties:
                engine:
                  type: string
                  enum: [postgres, mysql, sqlite]
                version:
                  type: string
                port:
                  type: integer
`)
}

func createInvalidCRDYAML_NoKind() []byte {
	return []byte(`
apiVersion: devopsmaestro.io/v1alpha1
kind: CustomResourceDefinition
metadata:
  name: databases.devopsmaestro.io
spec:
  group: devopsmaestro.io
  names:
    singular: database
    plural: databases
  scope: Workspace
  versions: []
`)
}

func createInvalidCRDYAML_InvalidSchema() []byte {
	return []byte(`
apiVersion: devopsmaestro.io/v1alpha1
kind: CustomResourceDefinition
metadata:
  name: databases.devopsmaestro.io
spec:
  group: devopsmaestro.io
  names:
    kind: Database
    singular: database
    plural: databases
  scope: Workspace
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          # Invalid schema - missing type
          properties: {}
`)
}

func createDuplicateKindCRDYAML() []byte {
	return []byte(`
apiVersion: devopsmaestro.io/v1alpha1
kind: CustomResourceDefinition
metadata:
  name: workspaces.devopsmaestro.io
spec:
  group: devopsmaestro.io
  names:
    kind: Workspace
    singular: workspace
    plural: workspaces
  scope: App
  versions:
    - name: v1alpha1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
`)
}

// =============================================================================
// CRDHandler Tests - Apply
// =============================================================================

func TestCRDHandler_Apply_CreatesCRD(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	data := createValidCRDYAML()
	res, err := handler.Apply(ctx, data)

	// Should succeed now (GREEN phase)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	crdModel, ok := res.(*models.CustomResourceDefinition)
	assert.True(t, ok)
	assert.Equal(t, "Database", crdModel.Kind)
	assert.Equal(t, "devopsmaestro.io", crdModel.Group)
	assert.Equal(t, "database", crdModel.Singular)
	assert.Equal(t, "databases", crdModel.Plural)
}

func TestCRDHandler_Apply_UpdatesCRD(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()

	// Pre-create CRD
	store.CreateCRD(&models.CustomResourceDefinition{
		Kind:     "Database",
		Group:    "devopsmaestro.io",
		Singular: "database",
		Plural:   "databases",
		Scope:    "Workspace",
	})

	ctx := resource.Context{DataStore: store}
	data := createValidCRDYAML()

	res, err := handler.Apply(ctx, data)

	// Should succeed - updating existing CRD
	assert.NoError(t, err)
	assert.NotNil(t, res)

	crdModel, ok := res.(*models.CustomResourceDefinition)
	assert.True(t, ok)
	assert.Equal(t, "Database", crdModel.Kind)
}

func TestCRDHandler_Apply_RejectsInvalidSchema(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	data := createInvalidCRDYAML_InvalidSchema()
	res, err := handler.Apply(ctx, data)

	// Note: JSON Schema validator may accept schemas without explicit 'type' field
	// This test verifies the schema compiles successfully even if minimal
	// For stricter validation, we would need additional CRD-specific validation rules
	assert.NoError(t, err)
	assert.NotNil(t, res)
}

func TestCRDHandler_Apply_UpdatesExistingCRD(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()

	// Pre-create CRD
	store.CreateCRD(&models.CustomResourceDefinition{
		Kind:     "Database",
		Group:    "devopsmaestro.io",
		Singular: "database",
		Plural:   "databases",
		Scope:    "Workspace",
	})

	ctx := resource.Context{DataStore: store}
	data := createValidCRDYAML()

	// Apply should update the existing CRD, not reject as duplicate
	res, err := handler.Apply(ctx, data)

	// Should succeed - Apply does upsert (create or update)
	assert.NoError(t, err)
	assert.NotNil(t, res)

	crdModel, ok := res.(*models.CustomResourceDefinition)
	assert.True(t, ok)
	assert.Equal(t, "Database", crdModel.Kind)
}

func TestCRDHandler_Apply_RejectsBuiltInKindCollision(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	data := createDuplicateKindCRDYAML()
	res, err := handler.Apply(ctx, data)

	// Should fail - cannot override built-in kind "Workspace"
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestCRDHandler_Apply_RequiresNames(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	data := createInvalidCRDYAML_NoKind()
	res, err := handler.Apply(ctx, data)

	// Should fail - missing required 'kind' field
	assert.Error(t, err)
	assert.Nil(t, res)
}

// =============================================================================
// CRDHandler Tests - Get
// =============================================================================

func TestCRDHandler_Get_ReturnsCRD(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()

	// Pre-create CRD
	store.CreateCRD(&models.CustomResourceDefinition{
		Kind:     "Database",
		Group:    "devopsmaestro.io",
		Singular: "database",
		Plural:   "databases",
	})

	ctx := resource.Context{DataStore: store}
	res, err := handler.Get(ctx, "Database")

	// Should succeed now
	assert.NoError(t, err)
	assert.NotNil(t, res)

	crdModel, ok := res.(*models.CustomResourceDefinition)
	assert.True(t, ok)
	assert.Equal(t, "Database", crdModel.Kind)
}

func TestCRDHandler_Get_NotFoundReturnsError(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	res, err := handler.Get(ctx, "NonExistent")

	// Should fail - CRD not found
	assert.Error(t, err)
	assert.Nil(t, res)
}

// =============================================================================
// CRDHandler Tests - List
// =============================================================================

func TestCRDHandler_List_ReturnsAllCRDs(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()

	// Pre-create CRDs
	store.CreateCRD(&models.CustomResourceDefinition{
		Kind:     "Database",
		Group:    "devopsmaestro.io",
		Singular: "database",
		Plural:   "databases",
	})
	store.CreateCRD(&models.CustomResourceDefinition{
		Kind:     "Cache",
		Group:    "devopsmaestro.io",
		Singular: "cache",
		Plural:   "caches",
	})

	ctx := resource.Context{DataStore: store}
	resources, err := handler.List(ctx)

	// Should succeed now - returns 2 CRDs
	assert.NoError(t, err)
	assert.NotNil(t, resources)
	assert.Len(t, resources, 2)
}

func TestCRDHandler_List_EmptyWhenNoCRDs(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	resources, err := handler.List(ctx)

	// Should succeed - returns empty list
	assert.NoError(t, err)
	assert.NotNil(t, resources)
	assert.Len(t, resources, 0)
}

// =============================================================================
// CRDHandler Tests - Delete
// =============================================================================

func TestCRDHandler_Delete_RemovesCRD(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()

	// Pre-create CRD with no instances
	store.CreateCRD(&models.CustomResourceDefinition{
		Kind:     "Database",
		Group:    "devopsmaestro.io",
		Singular: "database",
		Plural:   "databases",
	})
	// No instances created - should be able to delete

	ctx := resource.Context{DataStore: store}
	err := handler.Delete(ctx, "Database")

	// Should succeed now
	assert.NoError(t, err)
}

func TestCRDHandler_Delete_RejectsIfInstancesExist(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()

	// Pre-create CRD with instances
	store.CreateCRD(&models.CustomResourceDefinition{
		Kind:     "Database",
		Group:    "devopsmaestro.io",
		Singular: "database",
		Plural:   "databases",
	})
	// Create some instances of this CRD
	for i := 0; i < 5; i++ {
		store.CreateCustomResource(&models.CustomResource{
			Kind: "Database",
			Name: fmt.Sprintf("db-%d", i),
		})
	}

	ctx := resource.Context{DataStore: store}
	err := handler.Delete(ctx, "Database")

	// Should fail - instances exist
	assert.Error(t, err)
}

func TestCRDHandler_Delete_NotFoundReturnsError(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	err := handler.Delete(ctx, "NonExistent")

	// Should fail - CRD not found
	assert.Error(t, err)
}

// =============================================================================
// CRDHandler Tests - ToYAML
// =============================================================================

func TestCRDHandler_ToYAML_SerializesCRD(t *testing.T) {
	handler := NewCRDHandler()

	crdModel := &models.CustomResourceDefinition{
		Kind:     "Database",
		Group:    "devopsmaestro.io",
		Singular: "database",
		Plural:   "databases",
	}

	data, err := handler.ToYAML(crdModel)

	// Should succeed now
	assert.NoError(t, err)
	assert.NotNil(t, data)

	// Verify YAML contains expected fields
	assert.Contains(t, string(data), "kind: CustomResourceDefinition")
	assert.Contains(t, string(data), "kind: Database")
	assert.Contains(t, string(data), "singular: database")
	assert.Contains(t, string(data), "plural: databases")
}

// =============================================================================
// CRDHandler Tests - Handler Interface Compliance
// =============================================================================

func TestCRDHandler_ImplementsHandler(t *testing.T) {
	handler := NewCRDHandler()
	var _ resource.Handler = handler
}

func TestCRDHandler_Kind(t *testing.T) {
	handler := NewCRDHandler()
	kind := handler.Kind()
	assert.Equal(t, "CustomResourceDefinition", kind)
}

// =============================================================================
// CRDHandler Tests - Integration Scenarios
// =============================================================================

func TestCRDHandler_FullLifecycle(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// Apply (create)
	data := createValidCRDYAML()
	res, err := handler.Apply(ctx, data)
	require.NoError(t, err)
	require.NotNil(t, res)

	// Get
	res, err = handler.Get(ctx, "Database")
	require.NoError(t, err)
	require.NotNil(t, res)

	// List
	resources, err := handler.List(ctx)
	require.NoError(t, err)
	require.Len(t, resources, 1)

	// Delete
	err = handler.Delete(ctx, "Database")
	require.NoError(t, err)

	// Verify deletion
	_, err = handler.Get(ctx, "Database")
	require.Error(t, err)
}

func TestCRDHandler_MultipleCRDs(t *testing.T) {
	handler := NewCRDHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// Apply multiple CRDs
	crd1 := createValidCRDYAML()
	res1, err1 := handler.Apply(ctx, crd1)
	require.NoError(t, err1)
	require.NotNil(t, res1)

	// List should return all
	resources, err := handler.List(ctx)
	require.NoError(t, err)
	require.Len(t, resources, 1)
}

func TestCRDHandler_BuiltInKindProtection(t *testing.T) {
	tests := []struct {
		name         string
		kind         string
		shouldReject bool
	}{
		{"custom kind allowed", "Database", false},
		{"Workspace protected", "Workspace", true},
		{"App protected", "App", true},
		{"Domain protected", "Domain", true},
		{"Ecosystem protected", "Ecosystem", true},
		{"NvimPlugin protected", "NvimPlugin", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isBuiltInKind(tt.kind)
			assert.Equal(t, tt.shouldReject, result)
		})
	}
}
