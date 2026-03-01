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
// DevpiStrategy Basic Tests
// =============================================================================

func TestDevpiStrategy_GetDefaultPort(t *testing.T) {
	strategy := NewDevpiStrategy()

	port := strategy.GetDefaultPort()
	assert.Equal(t, 3141, port, "Default devpi port should be 3141")
}

func TestDevpiStrategy_GetDefaultStorage(t *testing.T) {
	strategy := NewDevpiStrategy()

	storage := strategy.GetDefaultStorage()
	assert.NotEmpty(t, storage, "Default storage should not be empty")
	assert.Equal(t, "/var/lib/devpi", storage, "Default storage should be /var/lib/devpi")
}

// =============================================================================
// DevpiStrategy ValidateConfig Tests
// =============================================================================

func TestDevpiStrategy_ValidateConfig_Valid(t *testing.T) {
	strategy := NewDevpiStrategy()

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
			config: json.RawMessage(`{"storage": "/tmp/devpi"}`),
		},
		{
			name:   "config with port",
			config: json.RawMessage(`{"port": 3142}`),
		},
		{
			name:   "config with upstreams",
			config: json.RawMessage(`{"upstreams": [{"name": "pypi", "url": "https://pypi.org/simple"}]}`),
		},
		{
			name:   "full config",
			config: json.RawMessage(`{"port": 3141, "storage": "/tmp/devpi", "lifecycle": "persistent"}`),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := strategy.ValidateConfig(tt.config)
			assert.NoError(t, err, "Valid config should pass validation")
		})
	}
}

func TestDevpiStrategy_ValidateConfig_Invalid(t *testing.T) {
	strategy := NewDevpiStrategy()

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
// DevpiStrategy CreateManager Tests
// =============================================================================

func TestDevpiStrategy_CreateManager_Success(t *testing.T) {
	t.Skip("Integration test - requires implementation of DevpiManager")

	strategy := NewDevpiStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-devpi",
		Type:      "devpi",
		Lifecycle: "persistent",
		Port:      3141,
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err, "Should create manager successfully")
	assert.NotNil(t, manager, "Manager should not be nil")

	// Verify manager implements ServiceManager
	assert.Implements(t, (*ServiceManager)(nil), manager)
}

func TestDevpiStrategy_CreateManager_WithDefaultPort(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewDevpiStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-devpi",
		Type:      "devpi",
		Lifecycle: "persistent",
		Port:      0, // Should use default
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err)
	assert.NotNil(t, manager)

	// Verify default port is used
	endpoint := manager.GetEndpoint()
	assert.Contains(t, endpoint, "3141", "Should use default port 3141")
}

func TestDevpiStrategy_CreateManager_WithCustomConfig(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewDevpiStrategy()

	customConfig := `{"storage": "/custom/path", "upstreams": [{"name": "pypi", "url": "https://pypi.org/simple"}]}`
	reg := &models.Registry{
		ID:        1,
		Name:      "test-devpi",
		Type:      "devpi",
		Lifecycle: "on-demand",
		Port:      3142,
		Config: sql.NullString{
			Valid:  true,
			String: customConfig,
		},
	}

	manager, err := strategy.CreateManager(reg)
	require.NoError(t, err)
	assert.NotNil(t, manager)
}

func TestDevpiStrategy_CreateManager_StoragePath(t *testing.T) {
	strategy := NewDevpiStrategy()

	tests := []struct {
		name        string
		registry    *models.Registry
		wantContain string
	}{
		{
			name: "uses name in default storage",
			registry: &models.Registry{
				ID:        1,
				Name:      "my-devpi",
				Type:      "devpi",
				Lifecycle: "persistent",
				Port:      3141,
			},
			wantContain: "my-devpi",
		},
		{
			name: "uses custom storage from config",
			registry: &models.Registry{
				ID:        1,
				Name:      "custom-devpi",
				Type:      "devpi",
				Lifecycle: "persistent",
				Port:      3141,
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
// DevpiStrategy Integration with Models
// =============================================================================

func TestDevpiStrategy_CreateManager_AllLifecycles(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewDevpiStrategy()

	lifecycles := []string{"persistent", "on-demand", "manual"}

	for _, lifecycle := range lifecycles {
		t.Run(lifecycle, func(t *testing.T) {
			reg := &models.Registry{
				ID:        1,
				Name:      "test-devpi",
				Type:      "devpi",
				Lifecycle: lifecycle,
				Port:      3141,
			}

			manager, err := strategy.CreateManager(reg)
			require.NoError(t, err, "Should support %s lifecycle", lifecycle)
			assert.NotNil(t, manager)
		})
	}
}

func TestDevpiStrategy_CreateManager_NilRegistry(t *testing.T) {
	strategy := NewDevpiStrategy()

	_, err := strategy.CreateManager(nil)
	assert.Error(t, err, "Should error on nil registry")
}

// =============================================================================
// DevpiStrategy Type Verification
// =============================================================================

func TestDevpiStrategy_IsCorrectType(t *testing.T) {
	// This test verifies that DevpiStrategy is the correct type

	var strategy RegistryStrategy = NewDevpiStrategy()

	// Should be a real DevpiStrategy type
	_, isDevpi := strategy.(*DevpiStrategy)
	assert.True(t, isDevpi, "Should be actual DevpiStrategy type")
}

func TestDevpiStrategy_CreateManager_DoesNotReturnNotImplementedError(t *testing.T) {
	t.Skip("Integration test - requires implementation")

	strategy := NewDevpiStrategy()

	reg := &models.Registry{
		ID:        1,
		Name:      "test-devpi",
		Type:      "devpi",
		Lifecycle: "persistent",
		Port:      3141,
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

func TestDevpiStrategy_getStoragePath_DefaultPath(t *testing.T) {
	t.Skip("Would need to test internal getStoragePath method")

	strategy := NewDevpiStrategy()

	reg := &models.Registry{
		Name: "test-devpi",
	}

	// Access private method via reflection or test the CreateManager behavior
	// For now, validate expected behavior through CreateManager

	// Expected: ~/.devopsmaestro/registries/test-devpi
	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".devopsmaestro", "registries", "test-devpi")

	// This would be validated in CreateManager integration test
	_, _, _ = strategy, reg, expectedPath
}

func TestDevpiStrategy_getStoragePath_CustomPath(t *testing.T) {
	strategy := NewDevpiStrategy()

	reg := &models.Registry{
		Name: "test-devpi",
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

func TestDevpiStrategy_getStoragePath_InvalidJSON(t *testing.T) {
	strategy := NewDevpiStrategy()

	reg := &models.Registry{
		Name: "test-devpi",
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
// Registry Strategy Registry Tests (ensure devpi is registered)
// =============================================================================

func TestStrategyRegistry_DevpiRegistered(t *testing.T) {
	t.Skip("Requires strategy registry implementation")

	// Verify devpi strategy is registered in the strategy registry
	// This ensures NewDevpiStrategy() returns a real strategy, not stub

	// strategy := GetRegistryStrategy("devpi")
	// assert.NotNil(t, strategy, "devpi strategy should be registered")

	// _, isStub := strategy.(*StubStrategy)
	// assert.False(t, isStub, "devpi should not be a stub")
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestDevpiStrategy_ImplementsRegistryStrategy(t *testing.T) {
	var _ RegistryStrategy = (*DevpiStrategy)(nil)
}
