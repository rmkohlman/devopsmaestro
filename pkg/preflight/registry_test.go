package preflight

import (
	"context"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
)

// =============================================================================
// TDD Phase 2 (RED): Registry Preflight Check Tests
// =============================================================================
// These tests verify the registry preflight check that validates registry health
// before operations. The check should:
//   - Verify registry services are reachable
//   - Auto-start on-demand registries if needed
//   - Skip manual lifecycle registries
//   - Provide warnings for unhealthy registries
//
// These tests will FAIL until implementation is added in Phase 3.
// =============================================================================

// ========== Interface Implementation Tests ==========

func TestRegistryCheck_ImplementsCheckInterface(t *testing.T) {
	// Test that RegistryCheck implements the Check interface
	var _ Check = (*RegistryCheck)(nil)
}

// ========== Name Method Tests ==========

func TestRegistryCheck_Name_ReturnsRegistryHealth(t *testing.T) {
	// Test that Name returns "Registry Health"
	check := NewRegistryCheck(nil, nil)
	assert.Equal(t, "Registry Health", check.Name())
}

// ========== Run Method Tests - Success Cases ==========

func TestRegistryCheck_Run_WithAllHealthy_ReturnsSuccess(t *testing.T) {
	// This test will be implemented when mocks are properly integrated
	// For now, we test that the basic structure works
	t.Skip("Phase 3: Awaiting full mock integration")
}

// ========== Run Method Tests - Warning Cases ==========

func TestRegistryCheck_Run_WithUnhealthyRegistry_ReturnsWarning(t *testing.T) {
	t.Skip("Phase 3: Awaiting full mock integration")
}

func TestRegistryCheck_Run_WithNoRegistries_ReturnsInfo(t *testing.T) {
	// Test with no registries - should return warning
	mockStore := db.NewMockDataStore()
	mockStore.Registries = make(map[string]*models.Registry) // Empty map

	check := NewRegistryCheck(mockStore, nil)
	result := check.Run(context.Background())

	assert.Equal(t, StatusWarning, result.Status, "Should return warning when no registries configured")
	assert.Contains(t, result.Message, "No registries configured")
}

func TestRegistryCheck_Run_WithNoOCIRegistry_ReturnsWarning(t *testing.T) {
	t.Skip("Phase 3: Awaiting full mock integration")
}

func TestRegistryCheck_Run_WithMixedHealth_ReturnsPartialSuccess(t *testing.T) {
	t.Skip("Phase 3: Awaiting full mock integration")
}

// ========== Context Cancellation Tests ==========

func TestRegistryCheck_Run_ContextCancellation(t *testing.T) {
	t.Skip("Phase 3: Awaiting full mock integration")
}

// ========== Lifecycle-Specific Tests ==========

func TestRegistryCheck_AutoStartsOnDemandRegistries(t *testing.T) {
	t.Skip("Phase 3: Awaiting full mock integration")
}

func TestRegistryCheck_SkipsManualLifecycleRegistries(t *testing.T) {
	t.Skip("Phase 3: Awaiting full mock integration")
}

// ========== Helper Tests ==========

func TestRegistryCheck_ShouldAutoStart(t *testing.T) {
	check := NewRegistryCheck(nil, nil)

	tests := []struct {
		name      string
		lifecycle string
		expected  bool
	}{
		{"auto lifecycle", "auto", true},
		{"persistent lifecycle", "persistent", true},
		{"on-demand lifecycle", "on-demand", true},
		{"manual lifecycle", "manual", false},
		{"unknown lifecycle", "unknown", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reg := &models.Registry{Lifecycle: tt.lifecycle}
			result := check.shouldAutoStart(reg)
			assert.Equal(t, tt.expected, result, "shouldAutoStart(%s) should return %v", tt.lifecycle, tt.expected)
		})
	}
}

func TestRegistryCheck_GetEnabledRegistries(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// Set some defaults
	mockStore.Defaults = map[string]string{
		"registry-oci":  "zot-main",
		"registry-pypi": "devpi-1",
		"registry-npm":  "", // Empty means not enabled
		"other-key":     "value",
	}

	check := NewRegistryCheck(mockStore, nil)
	enabled, err := check.getEnabledRegistries(context.Background())

	assert.NoError(t, err)
	assert.Len(t, enabled, 2, "Should have 2 enabled registries")
	assert.Equal(t, "zot-main", enabled["oci"])
	assert.Equal(t, "devpi-1", enabled["pypi"])
	assert.NotContains(t, enabled, "npm", "Empty value should not be included")
}

// =============================================================================
// Mock Implementations
// =============================================================================
// NOTE: These are minimal mocks. In Phase 3 implementation, the actual
// db.DataStore and registry.RegistryManager interfaces will be used.
// For now, tests will fail to compile until proper mocks are created.
// This is intentional TDD RED state.
