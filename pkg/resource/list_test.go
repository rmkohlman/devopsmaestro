package resource

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Mock helpers for list tests
// ---------------------------------------------------------------------------

// mockToYAMLHandler extends mockHandler with canned ToYAML output.
// It overrides the base mockHandler.ToYAML to return a full kubectl-style
// resource document so that BuildList can unmarshal it into map[string]any.
type mockToYAMLHandler struct {
	mockHandler
	yamlTemplate string // Go format string: receives resource name
}

func newMockToYAMLHandler(kind, yamlTemplate string) *mockToYAMLHandler {
	return &mockToYAMLHandler{
		mockHandler:  mockHandler{kind: kind, resources: make(map[string]*mockResource)},
		yamlTemplate: yamlTemplate,
	}
}

func (h *mockToYAMLHandler) ToYAML(res Resource) ([]byte, error) {
	out := fmt.Sprintf(h.yamlTemplate, res.GetName())
	return []byte(out), nil
}

// makeTestResource creates a mockResource directly for use in BuildList tests.
func makeTestResource(kind, name string) Resource {
	return &mockResource{kind: kind, name: name}
}

// ---------------------------------------------------------------------------
// TestNewResourceList
// ---------------------------------------------------------------------------

// TestNewResourceList verifies that NewResourceList returns a List with the
// correct envelope fields set and empty items.
func TestNewResourceList(t *testing.T) {
	rl := NewResourceList()

	require.NotNil(t, rl, "NewResourceList() should not return nil")
	assert.Equal(t, "devopsmaestro.io/v1", rl.APIVersion, "APIVersion should be 'devopsmaestro.io/v1'")
	assert.Equal(t, "List", rl.Kind, "Kind should be 'List'")
	assert.NotNil(t, rl.Metadata, "Metadata should not be nil (should be empty map)")
	assert.Empty(t, rl.Metadata, "Metadata should be an empty map")
	assert.NotNil(t, rl.Items, "Items should not be nil (should be empty slice)")
	assert.Empty(t, rl.Items, "Items should be empty on a new ResourceList")
}

// ---------------------------------------------------------------------------
// TestDependencyOrder
// ---------------------------------------------------------------------------

// TestDependencyOrder verifies the DependencyOrder constant contains all 13
// kinds in the correct order.
func TestDependencyOrder(t *testing.T) {
	expectedOrder := []string{
		"Ecosystem",
		"Domain",
		"App",
		"GitRepo",
		"Registry",
		"Credential",
		"Workspace",
		"NvimPlugin",
		"NvimTheme",
		"NvimPackage",
		"TerminalPrompt",
		"TerminalPackage",
	}

	assert.Equal(t, expectedOrder, DependencyOrder,
		"DependencyOrder must match the expected order exactly")
	assert.Len(t, DependencyOrder, len(expectedOrder),
		"DependencyOrder should have %d entries", len(expectedOrder))
}

// TestDependencyOrder_EcosystemFirst verifies Ecosystem is always first.
func TestDependencyOrder_EcosystemFirst(t *testing.T) {
	require.NotEmpty(t, DependencyOrder)
	assert.Equal(t, "Ecosystem", DependencyOrder[0],
		"Ecosystem must be first in DependencyOrder (no dependencies)")
}

// TestDependencyOrder_WorkspacesAfterApps verifies dependency chain ordering.
func TestDependencyOrder_WorkspacesAfterApps(t *testing.T) {
	kindIndex := func(kind string) int {
		for i, k := range DependencyOrder {
			if k == kind {
				return i
			}
		}
		return -1
	}

	assert.Less(t, kindIndex("Ecosystem"), kindIndex("Domain"),
		"Ecosystem must come before Domain")
	assert.Less(t, kindIndex("Domain"), kindIndex("App"),
		"Domain must come before App")
	assert.Less(t, kindIndex("App"), kindIndex("Workspace"),
		"App must come before Workspace")
}

// ---------------------------------------------------------------------------
// TestBuildList_Empty
// ---------------------------------------------------------------------------

// TestBuildList_Empty verifies BuildList with no resources returns a List
// with empty items (not nil).
func TestBuildList_Empty(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	ctx := Context{}
	rl, err := BuildList(ctx, []Resource{})

	require.NoError(t, err, "BuildList() with empty slice should not error")
	require.NotNil(t, rl, "BuildList() should return non-nil ResourceList")
	assert.Equal(t, "List", rl.Kind)
	assert.Equal(t, "devopsmaestro.io/v1", rl.APIVersion)
	assert.Empty(t, rl.Items, "BuildList() with empty resources should produce empty items")
}

// ---------------------------------------------------------------------------
// TestBuildList_SingleResource
// ---------------------------------------------------------------------------

// TestBuildList_SingleResource verifies BuildList with one resource produces
// a single-item list with the correct structure.
func TestBuildList_SingleResource(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	ecoYAML := `apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: %s
spec:
  description: "test"
`
	handler := newMockToYAMLHandler("Ecosystem", ecoYAML)
	Register(handler)

	ctx := Context{}
	resources := []Resource{makeTestResource("Ecosystem", "my-eco")}

	rl, err := BuildList(ctx, resources)

	require.NoError(t, err, "BuildList() should not error with a valid resource")
	require.NotNil(t, rl)
	assert.Len(t, rl.Items, 1, "should have exactly 1 item")

	// The item should be a map with 'kind' and 'metadata' keys
	item, ok := rl.Items[0].(map[string]any)
	require.True(t, ok, "item should be a map[string]any")
	assert.Equal(t, "Ecosystem", item["kind"], "item kind should be 'Ecosystem'")
	assert.NotNil(t, item["metadata"], "item should have metadata")
}

// ---------------------------------------------------------------------------
// TestBuildList_MultipleResources
// ---------------------------------------------------------------------------

// TestBuildList_MultipleResources verifies BuildList with mixed resource types
// preserves order and content.
func TestBuildList_MultipleResources(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	ecoYAML := `apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: %s
spec: {}
`
	domainYAML := `apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: %s
spec: {}
`
	Register(newMockToYAMLHandler("Ecosystem", ecoYAML))
	Register(newMockToYAMLHandler("Domain", domainYAML))

	ctx := Context{}
	resources := []Resource{
		makeTestResource("Ecosystem", "eco-1"),
		makeTestResource("Domain", "dom-1"),
		makeTestResource("Ecosystem", "eco-2"),
	}

	rl, err := BuildList(ctx, resources)

	require.NoError(t, err)
	require.NotNil(t, rl)
	assert.Len(t, rl.Items, 3, "should have 3 items")

	// Verify order is preserved
	item0, ok := rl.Items[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Ecosystem", item0["kind"])

	item1, ok := rl.Items[1].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Domain", item1["kind"])

	item2, ok := rl.Items[2].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Ecosystem", item2["kind"])
}

// ---------------------------------------------------------------------------
// TestBuildList_PreservesDependencyOrder
// ---------------------------------------------------------------------------

// TestBuildList_PreservesDependencyOrder verifies that BuildList preserves
// the input order (caller is responsible for sorting by DependencyOrder).
func TestBuildList_PreservesDependencyOrder(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	ecoYAML := `apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: %s
spec: {}
`
	appYAML := `apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: %s
spec: {}
`
	wsYAML := `apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: %s
spec: {}
`
	Register(newMockToYAMLHandler("Ecosystem", ecoYAML))
	Register(newMockToYAMLHandler("App", appYAML))
	Register(newMockToYAMLHandler("Workspace", wsYAML))

	ctx := Context{}
	// Pass in dependency order: Ecosystem -> App -> Workspace
	resources := []Resource{
		makeTestResource("Ecosystem", "my-eco"),
		makeTestResource("App", "my-app"),
		makeTestResource("Workspace", "my-ws"),
	}

	rl, err := BuildList(ctx, resources)

	require.NoError(t, err)
	require.NotNil(t, rl)
	require.Len(t, rl.Items, 3)

	// BuildList preserves the caller's ordering
	kinds := make([]string, 0, len(rl.Items))
	for _, item := range rl.Items {
		m, ok := item.(map[string]any)
		require.True(t, ok)
		kinds = append(kinds, m["kind"].(string))
	}

	assert.Equal(t, []string{"Ecosystem", "App", "Workspace"}, kinds,
		"BuildList should preserve caller's dependency ordering")
}

// ---------------------------------------------------------------------------
// TestResourceList_MarshalYAML
// ---------------------------------------------------------------------------

// TestResourceList_MarshalYAML verifies marshaling a ResourceList produces
// valid YAML with kind: List envelope.
func TestResourceList_MarshalYAML(t *testing.T) {
	rl := &ResourceList{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "List",
		Metadata:   map[string]any{},
		Items: []any{
			map[string]any{
				"apiVersion": "devopsmaestro.io/v1",
				"kind":       "Ecosystem",
				"metadata":   map[string]any{"name": "test-eco"},
				"spec":       map[string]any{},
			},
		},
	}

	data, err := yaml.Marshal(rl)
	require.NoError(t, err, "yaml.Marshal(ResourceList) should not error")

	yamlStr := string(data)
	assert.Contains(t, yamlStr, "apiVersion: devopsmaestro.io/v1", "YAML should have apiVersion")
	assert.Contains(t, yamlStr, "kind: List", "YAML should have kind: List")
	assert.Contains(t, yamlStr, "metadata:", "YAML should have metadata:")
	assert.Contains(t, yamlStr, "items:", "YAML should have items:")
	assert.Contains(t, yamlStr, "kind: Ecosystem", "YAML items should contain Ecosystem kind")
	assert.Contains(t, yamlStr, "test-eco", "YAML items should contain the ecosystem name")

	// Verify it round-trips back correctly
	var decoded ResourceList
	err = yaml.Unmarshal(data, &decoded)
	require.NoError(t, err, "yaml.Unmarshal of ResourceList YAML should not error")
	assert.Equal(t, "List", decoded.Kind)
	assert.Equal(t, "devopsmaestro.io/v1", decoded.APIVersion)
	assert.Len(t, decoded.Items, 1)
}

// ---------------------------------------------------------------------------
// TestResourceList_MarshalJSON
// ---------------------------------------------------------------------------

// TestResourceList_MarshalJSON verifies marshaling a ResourceList produces
// valid JSON with kind: List envelope.
func TestResourceList_MarshalJSON(t *testing.T) {
	rl := &ResourceList{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "List",
		Metadata:   map[string]any{},
		Items: []any{
			map[string]any{
				"apiVersion": "devopsmaestro.io/v1",
				"kind":       "Ecosystem",
				"metadata":   map[string]any{"name": "eco-json"},
			},
		},
	}

	data, err := json.Marshal(rl)
	require.NoError(t, err, "json.Marshal(ResourceList) should not error")

	var decoded map[string]any
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err, "JSON output should be valid")

	assert.Equal(t, "devopsmaestro.io/v1", decoded["apiVersion"])
	assert.Equal(t, "List", decoded["kind"])
	assert.NotNil(t, decoded["metadata"], "JSON should include metadata field")
	assert.NotNil(t, decoded["items"], "JSON should include items field")

	items, ok := decoded["items"].([]any)
	require.True(t, ok, "items should be a JSON array")
	assert.Len(t, items, 1)

	item, ok := items[0].(map[string]any)
	require.True(t, ok)
	assert.Equal(t, "Ecosystem", item["kind"])
}

// ---------------------------------------------------------------------------
// TestDetectKind_List
// ---------------------------------------------------------------------------

// TestDetectKind_List verifies that DetectKind() correctly identifies the
// "List" kind from a List YAML document. This tests existing functionality
// against the new List kind.
func TestDetectKind_List(t *testing.T) {
	listYAML := []byte(`apiVersion: devopsmaestro.io/v1
kind: List
metadata: {}
items: []
`)

	kind, err := DetectKind(listYAML)
	require.NoError(t, err, "DetectKind should not error on a valid List YAML")
	assert.Equal(t, "List", kind, "DetectKind should return 'List' for a List document")
}

// ---------------------------------------------------------------------------
// TestApplyList_EmptyItems
// ---------------------------------------------------------------------------

// TestApplyList_EmptyItems verifies that ApplyList with an empty items list
// returns success immediately with no resources and no error.
func TestApplyList_EmptyItems(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	ctx := Context{}
	listYAML := []byte(`apiVersion: devopsmaestro.io/v1
kind: List
metadata: {}
items: []
`)

	resources, err := ApplyList(ctx, listYAML)

	require.NoError(t, err, "ApplyList with empty items should not error")
	assert.Empty(t, resources, "ApplyList with empty items should return no resources")
}

// ---------------------------------------------------------------------------
// TestApplyList_HappyPath
// ---------------------------------------------------------------------------

// TestApplyList_HappyPath verifies that ApplyList with valid items applies all
// resources in order and returns all applied resources.
func TestApplyList_HappyPath(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	// Register a handler that applies successfully
	handler := newMockHandler("Ecosystem")
	Register(handler)

	ctx := Context{}
	listYAML := []byte(`apiVersion: devopsmaestro.io/v1
kind: List
metadata: {}
items:
  - apiVersion: devopsmaestro.io/v1
    kind: Ecosystem
    metadata:
      name: eco-applied
    spec: {}
  - apiVersion: devopsmaestro.io/v1
    kind: Ecosystem
    metadata:
      name: eco-applied-2
    spec: {}
`)

	resources, err := ApplyList(ctx, listYAML)

	require.NoError(t, err, "ApplyList happy path should not error")
	assert.Len(t, resources, 2, "ApplyList should return 2 applied resources")
}

// ---------------------------------------------------------------------------
// TestApplyList_ContinueOnError
// ---------------------------------------------------------------------------

// TestApplyList_ContinueOnError verifies that when one item fails, the
// remaining items are still applied (kubectl precedent: continue-on-error).
func TestApplyList_ContinueOnError(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	// Register an Ecosystem handler that works
	Register(newMockHandler("Ecosystem"))
	// Do NOT register a Domain handler — so Domain items will fail

	ctx := Context{}
	listYAML := []byte(`apiVersion: devopsmaestro.io/v1
kind: List
metadata: {}
items:
  - apiVersion: devopsmaestro.io/v1
    kind: Ecosystem
    metadata:
      name: eco-ok
    spec: {}
  - apiVersion: devopsmaestro.io/v1
    kind: Domain
    metadata:
      name: dom-fails
    spec: {}
  - apiVersion: devopsmaestro.io/v1
    kind: Ecosystem
    metadata:
      name: eco-also-ok
    spec: {}
`)

	resources, err := ApplyList(ctx, listYAML)

	// Should return an error (one item failed)
	assert.Error(t, err, "ApplyList should return error when any item fails")

	// But should have applied the 2 successful items
	assert.Len(t, resources, 2, "ApplyList should return the 2 successfully applied resources")
}

// ---------------------------------------------------------------------------
// TestApplyList_InvalidItem
// ---------------------------------------------------------------------------

// TestApplyList_InvalidItem verifies that an item missing the 'kind' field
// produces an error but processing continues for remaining items.
func TestApplyList_InvalidItem(t *testing.T) {
	ClearRegistry()
	defer ClearRegistry()

	Register(newMockHandler("Ecosystem"))

	ctx := Context{}
	listYAML := []byte(`apiVersion: devopsmaestro.io/v1
kind: List
metadata: {}
items:
  - apiVersion: devopsmaestro.io/v1
    metadata:
      name: no-kind-item
    spec: {}
  - apiVersion: devopsmaestro.io/v1
    kind: Ecosystem
    metadata:
      name: valid-eco
    spec: {}
`)

	resources, err := ApplyList(ctx, listYAML)

	// Should have an error (one item was invalid)
	assert.Error(t, err, "ApplyList should return error when item is missing kind")

	// The valid item should still have been applied
	assert.Len(t, resources, 1, "ApplyList should return 1 successfully applied resource")
}
