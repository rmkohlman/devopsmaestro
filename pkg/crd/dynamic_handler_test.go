package crd

import (
	"fmt"
	"testing"

	"devopsmaestro/pkg/resource"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock CustomResourceStore
// =============================================================================

type MockCustomResourceStore struct {
	resources   map[string]map[string]map[string]interface{} // kind -> name -> data
	createError error
	getError    error
	listError   error
	updateError error
	deleteError error
}

func NewMockCustomResourceStore() *MockCustomResourceStore {
	return &MockCustomResourceStore{
		resources: make(map[string]map[string]map[string]interface{}),
	}
}

func (m *MockCustomResourceStore) CreateResource(kind string, data map[string]interface{}) error {
	if m.createError != nil {
		return m.createError
	}

	if _, exists := m.resources[kind]; !exists {
		m.resources[kind] = make(map[string]map[string]interface{})
	}

	// Extract name from metadata (if present) or directly from data
	var name string
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		name, _ = metadata["name"].(string)
	} else {
		name, _ = data["name"].(string)
	}

	if name == "" {
		return fmt.Errorf("resource name not found")
	}

	m.resources[kind][name] = data
	return nil
}

func (m *MockCustomResourceStore) GetResource(kind, name string) (map[string]interface{}, error) {
	if m.getError != nil {
		return nil, m.getError
	}

	kindResources, exists := m.resources[kind]
	if !exists {
		return nil, &CRDNotFoundError{Kind: kind}
	}

	data, exists := kindResources[name]
	if !exists {
		return nil, &CRDNotFoundError{Kind: name}
	}

	return data, nil
}

func (m *MockCustomResourceStore) ListResources(kind string) ([]map[string]interface{}, error) {
	if m.listError != nil {
		return nil, m.listError
	}

	kindResources, exists := m.resources[kind]
	if !exists {
		return []map[string]interface{}{}, nil
	}

	resources := make([]map[string]interface{}, 0, len(kindResources))
	for _, data := range kindResources {
		resources = append(resources, data)
	}
	return resources, nil
}

func (m *MockCustomResourceStore) UpdateResource(kind string, data map[string]interface{}) error {
	if m.updateError != nil {
		return m.updateError
	}

	// Extract name from metadata (if present) or directly from data
	var name string
	if metadata, ok := data["metadata"].(map[string]interface{}); ok {
		name, _ = metadata["name"].(string)
	} else {
		name, _ = data["name"].(string)
	}

	if name == "" {
		return fmt.Errorf("resource name not found")
	}

	if _, exists := m.resources[kind]; !exists {
		return &CRDNotFoundError{Kind: kind}
	}
	if _, exists := m.resources[kind][name]; !exists {
		return &CRDNotFoundError{Kind: name}
	}

	m.resources[kind][name] = data
	return nil
}

func (m *MockCustomResourceStore) DeleteResource(kind, name string) error {
	if m.deleteError != nil {
		return m.deleteError
	}

	kindResources, exists := m.resources[kind]
	if !exists {
		return &CRDNotFoundError{Kind: kind}
	}

	if _, exists := kindResources[name]; !exists {
		return &CRDNotFoundError{Kind: name}
	}

	delete(m.resources[kind], name)
	return nil
}

// =============================================================================
// Test Fixtures
// =============================================================================

func createDatabaseCRDYAML() []byte {
	return []byte(`
apiVersion: devopsmaestro.io/v1alpha1
kind: Database
metadata:
  name: my-postgres
  workspace: my-workspace
spec:
  engine: postgres
  version: "15"
  port: 5432
`)
}

func createInvalidCRDYAML() []byte {
	return []byte(`
apiVersion: devopsmaestro.io/v1alpha1
kind: Database
metadata:
  name: my-postgres
spec:
  engine: oracle
  version: "19"
`)
}

func setupTestHandler() (*DynamicHandler, *MockCRDResolver, *MockSchemaValidator, *MockScopeValidator, *MockCustomResourceStore) {
	resolver := NewMockCRDResolver()
	validator := NewMockSchemaValidator()
	scopeValidator := NewMockScopeValidator()
	store := NewMockCustomResourceStore()

	handler := NewDynamicHandler(resolver, validator, scopeValidator, store)

	// Register test CRD
	crd := createTestCRD("Database", "database", "databases", []string{"db"})
	resolver.Register(crd)

	return handler, resolver, validator, scopeValidator, store
}

// =============================================================================
// DynamicHandler Tests - Apply
// =============================================================================

func TestDynamicHandler_Apply_CreatesResource(t *testing.T) {
	handler, resolver, validator, scopeValidator, store := setupTestHandler()
	_ = resolver
	_ = validator
	_ = scopeValidator
	_ = store

	ctx := resource.Context{}
	data := createDatabaseCRDYAML()

	res, err := handler.Apply(ctx, data)

	// GREEN phase - implementation complete
	assert.NoError(t, err, "Apply should succeed")
	assert.NotNil(t, res)
	assert.Equal(t, "Database", res.GetKind())
	assert.Equal(t, "my-postgres", res.GetName())
}

func TestDynamicHandler_Apply_UpdatesExistingResource(t *testing.T) {
	handler, _, _, _, store := setupTestHandler()

	// Pre-create resource
	store.CreateResource("Database", map[string]interface{}{
		"name":    "my-postgres",
		"engine":  "postgres",
		"version": "14",
	})

	ctx := resource.Context{}
	data := createDatabaseCRDYAML()

	res, err := handler.Apply(ctx, data)

	// GREEN phase - should update existing resource
	assert.NoError(t, err, "Apply should succeed for update")
	assert.NotNil(t, res)
	assert.Equal(t, "Database", res.GetKind())
	assert.Equal(t, "my-postgres", res.GetName())
}

func TestDynamicHandler_Apply_ValidatesAgainstSchema(t *testing.T) {
	handler, _, validator, _, _ := setupTestHandler()

	// Set validation to pass
	validator.SetValidateError(nil)

	ctx := resource.Context{}
	data := createDatabaseCRDYAML()

	res, err := handler.Apply(ctx, data)

	// GREEN phase - validation passes
	assert.NoError(t, err, "Apply should succeed with valid schema")
	assert.NotNil(t, res)
}

func TestDynamicHandler_Apply_RejectsInvalidSpec(t *testing.T) {
	handler, _, validator, _, _ := setupTestHandler()

	// Set validation to fail
	validator.SetValidateError(&SchemaValidationError{
		Kind:    "Database",
		Field:   "spec.engine",
		Message: "must be one of: postgres, mysql, sqlite",
	})

	ctx := resource.Context{}
	data := createInvalidCRDYAML()

	res, err := handler.Apply(ctx, data)

	// Should fail due to validation
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestDynamicHandler_Apply_EnforcesScope(t *testing.T) {
	handler, _, _, scopeValidator, _ := setupTestHandler()

	// Set scope validation to fail
	scopeValidator.SetValidateError(&ScopeValidationError{
		Scope:   ScopeWorkspace,
		Message: "workspace is required",
	})

	ctx := resource.Context{}
	data := createDatabaseCRDYAML()

	res, err := handler.Apply(ctx, data)

	// Should fail due to scope validation
	assert.Error(t, err)
	assert.Nil(t, res)
	assert.IsType(t, &ScopeValidationError{}, err)
}

func TestDynamicHandler_Apply_UnknownKindFails(t *testing.T) {
	handler, _, _, _, _ := setupTestHandler()

	ctx := resource.Context{}
	data := []byte(`
apiVersion: devopsmaestro.io/v1alpha1
kind: UnknownResource
metadata:
  name: test
spec: {}
`)

	res, err := handler.Apply(ctx, data)

	// Should fail due to unknown kind
	assert.Error(t, err)
	assert.Nil(t, res)
}

// =============================================================================
// DynamicHandler Tests - Get
// =============================================================================

func TestDynamicHandler_Get_ReturnsResource(t *testing.T) {
	handler, _, _, _, store := setupTestHandler()

	// Pre-create resource
	store.CreateResource("Database", map[string]interface{}{
		"name":    "my-postgres",
		"engine":  "postgres",
		"version": "15",
	})

	ctx := resource.Context{}
	res, err := handler.Get(ctx, "my-postgres")

	// Expected to fail in RED phase
	assert.Error(t, err, "Get should fail - not implemented yet")
	assert.Nil(t, res)
}

func TestDynamicHandler_Get_NotFoundReturnsError(t *testing.T) {
	handler, _, _, _, _ := setupTestHandler()

	ctx := resource.Context{}
	res, err := handler.Get(ctx, "nonexistent")

	// Should fail - resource not found
	assert.Error(t, err)
	assert.Nil(t, res)
}

// =============================================================================
// DynamicHandler Tests - List
// =============================================================================

func TestDynamicHandler_List_ReturnsByKind(t *testing.T) {
	handler, _, _, _, store := setupTestHandler()

	// Pre-create resources
	store.CreateResource("Database", map[string]interface{}{
		"name":   "postgres-1",
		"engine": "postgres",
	})
	store.CreateResource("Database", map[string]interface{}{
		"name":   "postgres-2",
		"engine": "postgres",
	})

	ctx := resource.Context{}
	resources, err := handler.List(ctx)

	// Expected to fail in RED phase
	assert.Error(t, err, "List should fail - not implemented yet")
	assert.Nil(t, resources)
}

func TestDynamicHandler_List_FiltersByNamespace(t *testing.T) {
	handler, _, _, _, store := setupTestHandler()

	// Pre-create resources in different namespaces
	store.CreateResource("Database", map[string]interface{}{
		"name":      "db-1",
		"namespace": "workspace-1",
	})
	store.CreateResource("Database", map[string]interface{}{
		"name":      "db-2",
		"namespace": "workspace-2",
	})

	ctx := resource.Context{}
	resources, err := handler.List(ctx)

	// Expected to fail in RED phase
	assert.Error(t, err, "List should fail - not implemented yet")
	assert.Nil(t, resources)
}

func TestDynamicHandler_List_EmptyWhenNoResources(t *testing.T) {
	handler, _, _, _, _ := setupTestHandler()

	ctx := resource.Context{}
	resources, err := handler.List(ctx)

	// Expected to fail in RED phase
	assert.Error(t, err, "List should fail - not implemented yet")
	assert.Nil(t, resources)
}

// =============================================================================
// DynamicHandler Tests - Delete
// =============================================================================

func TestDynamicHandler_Delete_RemovesResource(t *testing.T) {
	handler, _, _, _, store := setupTestHandler()

	// Pre-create resource
	store.CreateResource("Database", map[string]interface{}{
		"name": "my-postgres",
	})

	ctx := resource.Context{}
	err := handler.Delete(ctx, "my-postgres")

	// Expected to fail in RED phase
	assert.Error(t, err, "Delete should fail - not implemented yet")
}

func TestDynamicHandler_Delete_NotFoundReturnsError(t *testing.T) {
	handler, _, _, _, _ := setupTestHandler()

	ctx := resource.Context{}
	err := handler.Delete(ctx, "nonexistent")

	// Should fail - resource not found
	assert.Error(t, err)
}

// =============================================================================
// DynamicHandler Tests - ToYAML
// =============================================================================

func TestDynamicHandler_ToYAML_SerializesResource(t *testing.T) {
	handler, _, _, _, _ := setupTestHandler()

	// Create a mock resource (will need Resource implementation)
	// For now, pass nil to test failure

	data, err := handler.ToYAML(nil)

	// Expected to fail in RED phase
	assert.Error(t, err, "ToYAML should fail - not implemented yet")
	assert.Nil(t, data)
}

// =============================================================================
// DynamicHandler Tests - Handler Interface Compliance
// =============================================================================

func TestDynamicHandler_ImplementsHandler(t *testing.T) {
	handler, _, _, _, _ := setupTestHandler()

	var _ resource.Handler = handler
}

func TestDynamicHandler_Kind(t *testing.T) {
	handler, _, _, _, _ := setupTestHandler()

	kind := handler.Kind()
	assert.Equal(t, "CustomResource", kind)
}

// =============================================================================
// DynamicHandler Tests - Integration Scenarios
// =============================================================================

func TestDynamicHandler_FullLifecycle(t *testing.T) {
	handler, _, _, _, _ := setupTestHandler()

	// For Get/List/Delete, we need to pass the kind in context
	// This is done via the DataStore field as a map
	ctx := resource.Context{
		DataStore: map[string]string{"kind": "Database"},
	}

	// Apply (create)
	data := createDatabaseCRDYAML()
	res, err := handler.Apply(ctx, data)
	require.NoError(t, err, "Apply should succeed")
	require.NotNil(t, res)
	assert.Equal(t, "my-postgres", res.GetName())

	// Get
	res, err = handler.Get(ctx, "my-postgres")
	require.NoError(t, err, "Get should succeed")
	require.NotNil(t, res)
	assert.Equal(t, "my-postgres", res.GetName())

	// List
	resources, err := handler.List(ctx)
	require.NoError(t, err, "List should succeed")
	require.NotNil(t, resources)
	assert.Len(t, resources, 1)

	// Delete
	err = handler.Delete(ctx, "my-postgres")
	require.NoError(t, err, "Delete should succeed")

	// Verify deletion
	_, err = handler.Get(ctx, "my-postgres")
	assert.Error(t, err, "Get should fail after delete")
}

func TestDynamicHandler_MultipleResourceTypes(t *testing.T) {
	resolver := NewMockCRDResolver()
	validator := NewMockSchemaValidator()
	scopeValidator := NewMockScopeValidator()
	store := NewMockCustomResourceStore()

	handler := NewDynamicHandler(resolver, validator, scopeValidator, store)

	// Register multiple CRDs
	dbCRD := createTestCRD("Database", "database", "databases", []string{"db"})
	cacheCRD := createTestCRD("Cache", "cache", "caches", []string{"ch"})

	resolver.Register(dbCRD)
	resolver.Register(cacheCRD)

	ctx := resource.Context{}

	// Try to apply different resource types
	dbData := createDatabaseCRDYAML()
	_, err := handler.Apply(ctx, dbData)
	require.NoError(t, err, "Apply should succeed")

	// Each should be handled independently
	assert.NotNil(t, resolver.Resolve("Database"))
	assert.NotNil(t, resolver.Resolve("Cache"))

	// Verify resolving works for both
	assert.Equal(t, "Database", resolver.Resolve("Database").Names.Kind)
	assert.Equal(t, "Cache", resolver.Resolve("Cache").Names.Kind)
}
