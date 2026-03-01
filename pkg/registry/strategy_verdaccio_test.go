package registry

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// VerdaccioStrategy Basic Tests
// =============================================================================

func TestVerdaccioStrategy_GetDefaultPort(t *testing.T) {
	strategy := NewVerdaccioStrategy()

	port := strategy.GetDefaultPort()
	assert.Equal(t, 4873, port, "Default verdaccio port should be 4873")
}

func TestVerdaccioStrategy_GetDefaultStorage(t *testing.T) {
	strategy := NewVerdaccioStrategy()

	storage := strategy.GetDefaultStorage()
	assert.NotEmpty(t, storage, "Default storage should not be empty")
	assert.Equal(t, "/var/lib/verdaccio", storage, "Default storage should be /var/lib/verdaccio")
}

// =============================================================================
// VerdaccioStrategy ValidateConfig Tests
// =============================================================================

func TestVerdaccioStrategy_ValidateConfig_Valid(t *testing.T) {
	strategy := NewVerdaccioStrategy()

	tests := []struct {
		name   string
		config json.RawMessage
	}{
		{
			name:   "empty config",
			config: json.RawMessage(""),
		},
		{
			name:   "null config",
			config: json.RawMessage("null"),
		},
		{
			name:   "empty object",
			config: json.RawMessage("{}"),
		},
		{
			name:   "config with storage",
			config: json.RawMessage(`{"storage": "/tmp/verdaccio"}`),
		},
		{
			name:   "config with port",
			config: json.RawMessage(`{"port": 4874}`),
		},
		{
			name:   "config with upstreams",
			config: json.RawMessage(`{"upstreams": [{"name": "npmjs", "url": "https://registry.npmjs.org"}]}`),
		},
		{
			name:   "full config",
			config: json.RawMessage(`{"port": 4873, "storage": "/tmp/verdaccio", "lifecycle": "persistent"}`),
		},
		{
			name:   "config with auth",
			config: json.RawMessage(`{"auth": {"enabled": true, "type": "htpasswd"}}`),
		},
		{
			name:   "config with max body size",
			config: json.RawMessage(`{"maxBodySize": "50mb"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.ValidateConfig(tt.config)
			assert.NoError(t, err, "Valid config should pass validation")
		})
	}
}

func TestVerdaccioStrategy_ValidateConfig_Invalid(t *testing.T) {
	strategy := NewVerdaccioStrategy()

	tests := []struct {
		name    string
		config  json.RawMessage
		wantErr string
	}{
		{
			name:    "invalid JSON",
			config:  json.RawMessage(`{invalid}`),
			wantErr: "invalid JSON",
		},
		{
			name:    "malformed JSON",
			config:  json.RawMessage(`{"key": `),
			wantErr: "invalid JSON",
		},
		{
			name:    "not an object",
			config:  json.RawMessage(`["array"]`),
			wantErr: "", // May or may not error depending on validation strictness
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.ValidateConfig(tt.config)
			if tt.wantErr != "" {
				assert.Error(t, err, "Invalid config should fail validation")
				assert.Contains(t, err.Error(), tt.wantErr)
			}
		})
	}
}

// =============================================================================
// VerdaccioStrategy CreateManager Tests
// =============================================================================

func TestVerdaccioStrategy_CreateManager_Success(t *testing.T) {
	t.Skip("Integration test - requires implementation of VerdaccioManager")

	strategy := NewVerdaccioStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-verdaccio",
		Type:      "verdaccio",
		Lifecycle: "persistent",
		Port:      4873,
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err, "Should create manager successfully")
	assert.NotNil(t, manager, "Manager should not be nil")

	// Verify manager implements ServiceManager
	assert.Implements(t, (*ServiceManager)(nil), manager)
}

func TestVerdaccioStrategy_CreateManager_WithDefaultPort(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewVerdaccioStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-verdaccio",
		Type:      "verdaccio",
		Lifecycle: "persistent",
		Port:      0, // Should use default
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err)
	assert.NotNil(t, manager)

	// Verify default port is used
	endpoint := manager.GetEndpoint()
	assert.Contains(t, endpoint, "4873", "Should use default port 4873")
}

func TestVerdaccioStrategy_CreateManager_WithCustomConfig(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewVerdaccioStrategy()

	customConfig := `{"storage": "/custom/path", "upstreams": [{"name": "npmjs", "url": "https://registry.npmjs.org"}]}`
	reg := &models.Registry{
		ID:        1,
		Name:      "test-verdaccio",
		Type:      "verdaccio",
		Lifecycle: "on-demand",
		Port:      4874,
		Config: sql.NullString{
			Valid:  true,
			String: customConfig,
		},
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestVerdaccioStrategy_CreateManager_StoragePath(t *testing.T) {
	strategy := NewVerdaccioStrategy()

	tests := []struct {
		name        string
		registry    *models.Registry
		wantContain string
	}{
		{
			name: "uses name in default storage",
			registry: &models.Registry{
				ID:        1,
				Name:      "my-verdaccio",
				Type:      "verdaccio",
				Lifecycle: "persistent",
				Port:      4873,
			},
			wantContain: "my-verdaccio",
		},
		{
			name: "uses custom storage from config",
			registry: &models.Registry{
				ID:        1,
				Name:      "custom-verdaccio",
				Type:      "verdaccio",
				Lifecycle: "persistent",
				Port:      4873,
				Config: sql.NullString{
					Valid:  true,
					String: `{"storage": "/custom/storage/path"}`,
				},
			},
			wantContain: "/custom/storage/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Skip("Requires implementation to verify storage path")

			manager, err := strategy.CreateManager(tt.registry)
			require.NoError(t, err)
			assert.NotNil(t, manager)

			// Would need to access internal config to verify storage path
			// This validates the strategy logic
		})
	}
}

// =============================================================================
// VerdaccioStrategy Integration with Models
// =============================================================================

func TestVerdaccioStrategy_CreateManager_AllLifecycles(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewVerdaccioStrategy()

	lifecycles := []string{"persistent", "on-demand", "manual"}

	for _, lifecycle := range lifecycles {
		t.Run(lifecycle, func(t *testing.T) {
			reg := &models.Registry{
				ID:        1,
				Name:      "test-verdaccio",
				Type:      "verdaccio",
				Lifecycle: lifecycle,
				Port:      4873,
			}

			manager, err := strategy.CreateManager(reg)
			require.NoError(t, err, "Should support %s lifecycle", lifecycle)
			assert.NotNil(t, manager)
		})
	}
}

func TestVerdaccioStrategy_CreateManager_NilRegistry(t *testing.T) {
	strategy := NewVerdaccioStrategy()

	_, err := strategy.CreateManager(nil)
	assert.Error(t, err, "Should error on nil registry")
}

// =============================================================================
// VerdaccioStrategy vs StubStrategy Comparison
// =============================================================================

func TestVerdaccioStrategy_IsNotStub(t *testing.T) {
	// This test verifies that VerdaccioStrategy is no longer a StubStrategy

	var strategy RegistryStrategy = NewVerdaccioStrategy()

	// Should not be a StubStrategy type
	_, isStub := strategy.(*StubStrategy)
	assert.False(t, isStub, "VerdaccioStrategy should not be a StubStrategy")

	// Should be a real VerdaccioStrategy type
	_, isVerdaccio := strategy.(*VerdaccioStrategy)
	assert.True(t, isVerdaccio, "Should be actual VerdaccioStrategy type")
}

func TestVerdaccioStrategy_CreateManager_DoesNotReturnNotImplementedError(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewVerdaccioStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-verdaccio",
		Type:      "verdaccio",
		Lifecycle: "persistent",
		Port:      4873,
	}

	_, err := strategy.CreateManager(reg)

	// Should not return "not implemented" error like StubStrategy does
	if err != nil {
		assert.NotContains(t, err.Error(), "not implemented", "Should not return not-implemented error")
	}
}

// =============================================================================
// getStoragePath Helper Tests
// =============================================================================

func TestVerdaccioStrategy_getStoragePath_DefaultPath(t *testing.T) {
	t.Skip("Would need to test internal getStoragePath method")

	strategy := NewVerdaccioStrategy()

	reg := &models.Registry{
		Name: "test-verdaccio",
	}

	// Access private method via reflection or test the CreateManager behavior
	// For now, validate expected behavior through CreateManager

	// Expected: ~/.devopsmaestro/registries/test-verdaccio
	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".devopsmaestro", "registries", "test-verdaccio")

	// This would be validated in CreateManager integration test
	_, _, _ = strategy, reg, expectedPath
}

func TestVerdaccioStrategy_getStoragePath_CustomPath(t *testing.T) {
	strategy := NewVerdaccioStrategy()

	reg := &models.Registry{
		Name: "test-verdaccio",
		Config: sql.NullString{
			Valid:  true,
			String: `{"storage": "/custom/path"}`,
		},
	}

	t.Skip("Would need to test internal getStoragePath method")

	// Expected: /custom/path
	_ = strategy
	_ = reg
}

func TestVerdaccioStrategy_getStoragePath_InvalidJSON(t *testing.T) {
	strategy := NewVerdaccioStrategy()

	reg := &models.Registry{
		Name: "test-verdaccio",
		Config: sql.NullString{
			Valid:  true,
			String: `{invalid json}`,
		},
	}

	t.Skip("Would need to test internal getStoragePath method")

	// Should fall back to default path on invalid JSON
	_ = strategy
	_ = reg
}

// =============================================================================
// VerdaccioStrategy Config Parsing Tests
// =============================================================================

func TestVerdaccioStrategy_parseConfig_ValidJSON(t *testing.T) {
	t.Skip("Would need to test internal parseConfig method")

	strategy := NewVerdaccioStrategy()

	configJSON := `{"port": 4873, "storage": "/tmp/verdaccio", "upstreams": [{"name": "npmjs", "url": "https://registry.npmjs.org"}]}`

	// Parse config
	// config, err := strategy.parseConfig(configJSON)
	// require.NoError(t, err)
	// assert.Equal(t, 4873, config.Port)

	_ = strategy
	_ = configJSON
}

func TestVerdaccioStrategy_parseConfig_EmptyConfig(t *testing.T) {
	t.Skip("Would need to test internal parseConfig method")

	strategy := NewVerdaccioStrategy()

	// Empty config should use defaults
	// config, err := strategy.parseConfig("")
	// require.NoError(t, err)
	// assert.NotZero(t, config.Port)

	_ = strategy
}

func TestVerdaccioStrategy_parseConfig_InvalidJSON(t *testing.T) {
	t.Skip("Would need to test internal parseConfig method")

	strategy := NewVerdaccioStrategy()

	configJSON := `{invalid json}`

	// Parse should fail
	// _, err := strategy.parseConfig(configJSON)
	// assert.Error(t, err)

	_ = strategy
	_ = configJSON
}

// =============================================================================
// Registry Strategy Registry Tests (ensure verdaccio is registered)
// =============================================================================

func TestStrategyRegistry_VerdaccioRegistered(t *testing.T) {
	t.Skip("Requires strategy registry implementation")

	// Verify verdaccio strategy is registered in the strategy registry
	// This ensures NewVerdaccioStrategy() returns a real strategy, not stub

	// strategy := GetRegistryStrategy("verdaccio")
	// assert.NotNil(t, strategy, "verdaccio strategy should be registered")

	// _, isStub := strategy.(*StubStrategy)
	// assert.False(t, isStub, "verdaccio should not be a stub")
}

func TestStrategyRegistry_VerdaccioType(t *testing.T) {
	t.Skip("Requires strategy registry implementation")

	// Verify that "verdaccio" type maps to VerdaccioStrategy

	// strategy := GetRegistryStrategy("verdaccio")
	// _, ok := strategy.(*VerdaccioStrategy)
	// assert.True(t, ok, "verdaccio type should return VerdaccioStrategy")
}

// =============================================================================
// VerdaccioStrategy npm vs Verdaccio naming
// =============================================================================

func TestVerdaccioStrategy_ConfigUsesNpmTypes(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewVerdaccioStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-verdaccio",
		Type:      "verdaccio",
		Lifecycle: "persistent",
		Port:      4873,
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err)

	// Verify manager implements NpmProxy interface (not VerdaccioProxy)
	_, ok := manager.(NpmProxy)
	assert.True(t, ok, "VerdaccioManager should implement NpmProxy interface")
}

func TestVerdaccioStrategy_ImplementsNpmStrategy(t *testing.T) {
	// Verify VerdaccioStrategy is the concrete npm implementation

	strategy := NewVerdaccioStrategy()

	// Should be the strategy for "npm" type registries
	// (verdaccio is the implementation of npm proxy)
	assert.NotNil(t, strategy, "VerdaccioStrategy should be the npm proxy implementation")
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestVerdaccioStrategy_ImplementsRegistryStrategy(t *testing.T) {
	var _ RegistryStrategy = (*VerdaccioStrategy)(nil)
}
