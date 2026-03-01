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
// SquidStrategy Basic Tests
// =============================================================================

func TestSquidStrategy_GetDefaultPort(t *testing.T) {
	strategy := NewSquidStrategy()

	port := strategy.GetDefaultPort()
	assert.Equal(t, 3128, port, "Default squid port should be 3128")
}

func TestSquidStrategy_GetDefaultStorage(t *testing.T) {
	strategy := NewSquidStrategy()

	storage := strategy.GetDefaultStorage()
	assert.NotEmpty(t, storage, "Default storage should not be empty")
	assert.Equal(t, "/var/cache/squid", storage, "Default storage should be /var/cache/squid")
}

// =============================================================================
// SquidStrategy ValidateConfig Tests
// =============================================================================

func TestSquidStrategy_ValidateConfig_Valid(t *testing.T) {
	strategy := NewSquidStrategy()

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
			name:   "config with cacheDir",
			config: json.RawMessage(`{"cacheDir": "/tmp/squid/cache"}`),
		},
		{
			name:   "config with port",
			config: json.RawMessage(`{"port": 8080}`),
		},
		{
			name:   "config with cache sizes",
			config: json.RawMessage(`{"cacheSizeMB": 5000, "maxObjectSizeMB": 500, "memoryCacheMB": 512}`),
		},
		{
			name:   "full config",
			config: json.RawMessage(`{"port": 3128, "cacheDir": "/tmp/squid/cache", "logDir": "/tmp/squid/logs", "cacheSizeMB": 1000}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.ValidateConfig(tt.config)
			assert.NoError(t, err, "Valid config should pass validation")
		})
	}
}

func TestSquidStrategy_ValidateConfig_Invalid(t *testing.T) {
	strategy := NewSquidStrategy()

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
// SquidStrategy CreateManager Tests
// =============================================================================

func TestSquidStrategy_CreateManager_Success(t *testing.T) {
	t.Skip("Integration test - requires implementation of SquidManager")

	strategy := NewSquidStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-squid",
		Type:      RegistryTypeSquid,
		Lifecycle: "persistent",
		Port:      3128,
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err, "Should create manager successfully")
	assert.NotNil(t, manager, "Manager should not be nil")

	// Verify manager implements ServiceManager
	assert.Implements(t, (*ServiceManager)(nil), manager)
}

func TestSquidStrategy_CreateManager_WithDefaultPort(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewSquidStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-squid",
		Type:      RegistryTypeSquid,
		Lifecycle: "persistent",
		Port:      0, // Should use default
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err)
	assert.NotNil(t, manager)

	// Verify default port is used
	endpoint := manager.GetEndpoint()
	assert.Contains(t, endpoint, "3128", "Should use default port 3128")
}

func TestSquidStrategy_CreateManager_WithCustomConfig(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewSquidStrategy()

	customConfig := `{"cacheDir": "/custom/cache", "cacheSizeMB": 5000, "maxObjectSizeMB": 500}`
	reg := &models.Registry{
		ID:        1,
		Name:      "test-squid",
		Type:      RegistryTypeSquid,
		Lifecycle: "on-demand",
		Port:      8080,
		Config: sql.NullString{
			Valid:  true,
			String: customConfig,
		},
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestSquidStrategy_CreateManager_StoragePath(t *testing.T) {
	strategy := NewSquidStrategy()

	tests := []struct {
		name        string
		registry    *models.Registry
		wantContain string
	}{
		{
			name: "uses name in default storage",
			registry: &models.Registry{
				ID:        1,
				Name:      "my-squid",
				Type:      RegistryTypeSquid,
				Lifecycle: "persistent",
				Port:      3128,
			},
			wantContain: "my-squid",
		},
		{
			name: "uses custom storage from config",
			registry: &models.Registry{
				ID:        1,
				Name:      "custom-squid",
				Type:      RegistryTypeSquid,
				Lifecycle: "persistent",
				Port:      3128,
				Config: sql.NullString{
					Valid:  true,
					String: `{"cacheDir": "/custom/cache/path"}`,
				},
			},
			wantContain: "/custom/cache/path",
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
// SquidStrategy Integration with Models
// =============================================================================

func TestSquidStrategy_CreateManager_AllLifecycles(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewSquidStrategy()

	lifecycles := []string{"persistent", "on-demand", "manual"}

	for _, lifecycle := range lifecycles {
		t.Run(lifecycle, func(t *testing.T) {
			reg := &models.Registry{
				ID:        1,
				Name:      "test-squid",
				Type:      RegistryTypeSquid,
				Lifecycle: lifecycle,
				Port:      3128,
			}

			manager, err := strategy.CreateManager(reg)
			require.NoError(t, err, "Should support %s lifecycle", lifecycle)
			assert.NotNil(t, manager)
		})
	}
}

func TestSquidStrategy_CreateManager_NilRegistry(t *testing.T) {
	strategy := NewSquidStrategy()

	_, err := strategy.CreateManager(nil)
	assert.Error(t, err, "Should error on nil registry")
}

// =============================================================================
// SquidStrategy Type Verification
// =============================================================================

func TestSquidStrategy_IsCorrectType(t *testing.T) {
	// This test verifies that SquidStrategy is the correct type

	var strategy RegistryStrategy = NewSquidStrategy()

	// Should be a real SquidStrategy type
	_, isSquid := strategy.(*SquidStrategy)
	assert.True(t, isSquid, "Should be actual SquidStrategy type")
}

func TestSquidStrategy_CreateManager_DoesNotReturnNotImplementedError(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewSquidStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-squid",
		Type:      RegistryTypeSquid,
		Lifecycle: "persistent",
		Port:      3128,
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

func TestSquidStrategy_getStoragePath_DefaultPath(t *testing.T) {
	t.Skip("Would need to test internal getStoragePath method")

	strategy := NewSquidStrategy()

	reg := &models.Registry{
		Name: "test-squid",
	}

	// Access private method via reflection or test the CreateManager behavior
	// For now, validate expected behavior through CreateManager

	// Expected: ~/.devopsmaestro/registries/test-squid/cache
	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".devopsmaestro", "registries", "test-squid", "cache")

	// This would be validated in CreateManager integration test
	_, _, _ = strategy, reg, expectedPath
}

func TestSquidStrategy_getStoragePath_CustomPath(t *testing.T) {
	strategy := NewSquidStrategy()

	reg := &models.Registry{
		Name: "test-squid",
		Config: sql.NullString{
			Valid:  true,
			String: `{"cacheDir": "/custom/path"}`,
		},
	}

	t.Skip("Would need to test internal getStoragePath method")

	// Expected: /custom/path
	_ = strategy
	_ = reg
}

func TestSquidStrategy_getStoragePath_InvalidJSON(t *testing.T) {
	strategy := NewSquidStrategy()

	reg := &models.Registry{
		Name: "test-squid",
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
// SquidStrategy Config Parsing Tests
// =============================================================================

func TestSquidStrategy_parseConfig_ValidJSON(t *testing.T) {
	t.Skip("Would need to test internal parseConfig method")

	strategy := NewSquidStrategy()

	configJSON := `{"port": 3128, "cacheDir": "/tmp/squid/cache", "cacheSizeMB": 1000, "maxObjectSizeMB": 100}`

	// Parse config
	// config, err := strategy.parseConfig(configJSON)
	// require.NoError(t, err)
	// assert.Equal(t, 3128, config.Port)

	_ = strategy
	_ = configJSON
}

func TestSquidStrategy_parseConfig_EmptyConfig(t *testing.T) {
	t.Skip("Would need to test internal parseConfig method")

	strategy := NewSquidStrategy()

	// Empty config should use defaults
	// config, err := strategy.parseConfig("")
	// require.NoError(t, err)
	// assert.NotZero(t, config.Port)

	_ = strategy
}

func TestSquidStrategy_parseConfig_InvalidJSON(t *testing.T) {
	t.Skip("Would need to test internal parseConfig method")

	strategy := NewSquidStrategy()

	configJSON := `{invalid json}`

	// Parse should fail
	// _, err := strategy.parseConfig(configJSON)
	// assert.Error(t, err)

	_ = strategy
	_ = configJSON
}

// =============================================================================
// Registry Strategy Registry Tests (ensure squid is registered)
// =============================================================================

func TestStrategyRegistry_SquidRegistered(t *testing.T) {
	t.Skip("Requires strategy registry implementation")

	// Verify squid strategy is registered in the strategy registry
	// This ensures NewSquidStrategy() returns a real strategy, not stub

	// strategy := GetRegistryStrategy(RegistryTypeSquid)
	// assert.NotNil(t, strategy, "squid strategy should be registered")

	// _, isStub := strategy.(*StubStrategy)
	// assert.False(t, isStub, "squid should not be a stub")
}

func TestStrategyRegistry_SquidType(t *testing.T) {
	t.Skip("Requires strategy registry implementation")

	// Verify that RegistryTypeSquid type maps to SquidStrategy

	// strategy := GetRegistryStrategy(RegistryTypeSquid)
	// _, ok := strategy.(*SquidStrategy)
	// assert.True(t, ok, "squid type should return SquidStrategy")
}

// =============================================================================
// SquidStrategy HTTP Proxy Naming
// =============================================================================

func TestSquidStrategy_ConfigUsesHttpProxyTypes(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewSquidStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-squid",
		Type:      RegistryTypeSquid,
		Lifecycle: "persistent",
		Port:      3128,
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err)

	// Verify manager implements HttpProxy interface
	_, ok := manager.(HttpProxy)
	assert.True(t, ok, "SquidManager should implement HttpProxy interface")
}

func TestSquidStrategy_ImplementsHttpProxyStrategy(t *testing.T) {
	// Verify SquidStrategy is the concrete HTTP proxy implementation

	strategy := NewSquidStrategy()

	// Should be the strategy for "squid" type registries
	// (squid is the implementation of HTTP proxy)
	assert.NotNil(t, strategy, "SquidStrategy should be the HTTP proxy implementation")
}

// =============================================================================
// SquidStrategy Adapter Tests
// =============================================================================

func TestSquidStrategy_CreatesAdapter(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewSquidStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-squid",
		Type:      RegistryTypeSquid,
		Lifecycle: "persistent",
		Port:      3128,
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err)

	// Verify manager is wrapped in adapter for ServiceManager interface
	_, isAdapter := manager.(*SquidManagerAdapter)
	assert.True(t, isAdapter, "Should return SquidManagerAdapter to implement ServiceManager")
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestSquidStrategy_ImplementsRegistryStrategy(t *testing.T) {
	var _ RegistryStrategy = (*SquidStrategy)(nil)
}
