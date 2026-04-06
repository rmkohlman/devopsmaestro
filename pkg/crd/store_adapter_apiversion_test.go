package crd

// ---------------------------------------------------------------------------
// store_adapter_apiversion_test.go — Issue #180
//
// RED-phase test verifying that toResourceMap() includes an "apiVersion"
// field. Currently it does NOT, which causes custom resource instances to be
// exported without apiVersion — making them impossible to re-apply correctly.
//
// This test MUST FAIL until toResourceMap is updated.
// ---------------------------------------------------------------------------

import (
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToResourceMap_IncludesApiVersion verifies that toResourceMap sets the
// "apiVersion" field on the returned resource map. Without apiVersion, the
// exported YAML cannot be re-applied (consumers need it to identify the schema).
func TestToResourceMap_IncludesApiVersion(t *testing.T) {
	// Arrange: create a minimal CustomResource
	cr := &models.CustomResource{
		Kind: "AppConfig",
		Name: "test-resource",
	}

	// Use a real (no-op) adapter — toResourceMap is a pure value transform, no DB needed
	adapter := &DataStoreAdapter{}

	// Act
	result := adapter.toResourceMap(cr)
	require.NotNil(t, result, "toResourceMap should return a non-nil map")

	// Assert: "apiVersion" must be present
	// This MUST FAIL (RED phase) — toResourceMap never sets "apiVersion"
	apiVersion, exists := result["apiVersion"]
	assert.True(t, exists,
		"toResourceMap result must contain 'apiVersion' key — currently missing (Issue #180)")
	assert.NotEmpty(t, apiVersion,
		"toResourceMap 'apiVersion' must not be empty — expected e.g. 'devopsmaestro.io/v1alpha1'")

	// Assert: "kind" must still be present (regression check)
	kind, kindExists := result["kind"]
	assert.True(t, kindExists, "toResourceMap result should still contain 'kind' key")
	assert.Equal(t, "AppConfig", kind, "toResourceMap 'kind' should match cr.Kind")

	// Assert: "metadata" must still be present (regression check)
	meta, metaExists := result["metadata"]
	assert.True(t, metaExists, "toResourceMap result should still contain 'metadata' key")
	metaMap, ok := meta.(map[string]interface{})
	require.True(t, ok, "'metadata' should be a map")
	assert.Equal(t, "test-resource", metaMap["name"], "'metadata.name' should match cr.Name")
}
