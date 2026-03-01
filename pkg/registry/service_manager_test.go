package registry

import (
	"context"
	"encoding/json"
	"testing"

	"devopsmaestro/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestServiceManagerInterface verifies that existing managers implement ServiceManager
func TestServiceManagerInterface(t *testing.T) {
	t.Run("ZotManager implements ServiceManager", func(t *testing.T) {
		var _ ServiceManager = &ZotManager{}
	})

	t.Run("AthensManager implements ServiceManager", func(t *testing.T) {
		var _ ServiceManager = &AthensManagerAdapter{}
	})
}

// TestZotStrategy tests the Zot registry strategy
func TestZotStrategy(t *testing.T) {
	strategy := NewZotStrategy()

	t.Run("ValidateConfig accepts valid zot config", func(t *testing.T) {
		config := map[string]interface{}{
			"storage": "/var/lib/zot",
		}
		configJSON, _ := json.Marshal(config)

		err := strategy.ValidateConfig(configJSON)
		assert.NoError(t, err)
	})

	t.Run("ValidateConfig rejects invalid JSON", func(t *testing.T) {
		err := strategy.ValidateConfig([]byte("{invalid"))
		assert.Error(t, err)
	})

	t.Run("GetDefaultPort returns 5000", func(t *testing.T) {
		assert.Equal(t, 5000, strategy.GetDefaultPort())
	})

	t.Run("GetDefaultStorage returns zot path", func(t *testing.T) {
		assert.Equal(t, "/var/lib/zot", strategy.GetDefaultStorage())
	})

	t.Run("CreateManager creates a ZotManager", func(t *testing.T) {
		reg := &models.Registry{
			ID:   1,
			Name: "test-zot",
			Type: "zot",
			Port: 5001,
		}

		manager, err := strategy.CreateManager(reg)
		require.NoError(t, err)
		assert.NotNil(t, manager)

		// Verify it's a ServiceManager
		_, ok := manager.(ServiceManager)
		assert.True(t, ok, "CreateManager should return a ServiceManager")
	})
}

// TestAthensStrategy tests the Athens proxy strategy
func TestAthensStrategy(t *testing.T) {
	strategy := NewAthensStrategy()

	t.Run("ValidateConfig accepts valid athens config", func(t *testing.T) {
		config := map[string]interface{}{
			"storage": "/var/lib/athens",
		}
		configJSON, _ := json.Marshal(config)

		err := strategy.ValidateConfig(configJSON)
		assert.NoError(t, err)
	})

	t.Run("GetDefaultPort returns 3000", func(t *testing.T) {
		assert.Equal(t, 3000, strategy.GetDefaultPort())
	})

	t.Run("GetDefaultStorage returns athens path", func(t *testing.T) {
		assert.Equal(t, "/var/lib/athens", strategy.GetDefaultStorage())
	})

	t.Run("CreateManager creates an AthensManagerAdapter", func(t *testing.T) {
		reg := &models.Registry{
			ID:   1,
			Name: "test-athens",
			Type: "athens",
			Port: 3001,
		}

		manager, err := strategy.CreateManager(reg)
		require.NoError(t, err)
		assert.NotNil(t, manager)

		// Verify it's a ServiceManager
		_, ok := manager.(ServiceManager)
		assert.True(t, ok, "CreateManager should return a ServiceManager")
	})
}

// TestSquidStrategy tests the Squid proxy strategy
func TestSquidStrategy(t *testing.T) {
	strategy := NewSquidStrategy()

	t.Run("ValidateConfig accepts valid squid config", func(t *testing.T) {
		config := map[string]interface{}{
			"port":     3128,
			"cacheDir": "/var/cache/squid",
		}
		configJSON, _ := json.Marshal(config)

		err := strategy.ValidateConfig(configJSON)
		assert.NoError(t, err)
	})

	t.Run("ValidateConfig rejects invalid JSON", func(t *testing.T) {
		err := strategy.ValidateConfig([]byte("{invalid"))
		assert.Error(t, err)
	})

	t.Run("GetDefaultPort returns 3128", func(t *testing.T) {
		assert.Equal(t, 3128, strategy.GetDefaultPort())
	})

	t.Run("GetDefaultStorage returns squid path", func(t *testing.T) {
		assert.Equal(t, "/var/cache/squid", strategy.GetDefaultStorage())
	})

	t.Run("CreateManager creates a SquidManagerAdapter", func(t *testing.T) {
		reg := &models.Registry{
			ID:   1,
			Name: "test-squid",
			Type: "squid",
			Port: 3128,
		}

		manager, err := strategy.CreateManager(reg)
		require.NoError(t, err)
		assert.NotNil(t, manager)

		// Verify it's a ServiceManager
		_, ok := manager.(ServiceManager)
		assert.True(t, ok, "CreateManager should return a ServiceManager")
	})
}

// TestServiceFactory tests the service factory
func TestServiceFactory(t *testing.T) {
	factory := NewServiceFactory()

	t.Run("GetStrategy for zot", func(t *testing.T) {
		strategy, err := factory.GetStrategy("zot")
		require.NoError(t, err)
		assert.NotNil(t, strategy)
		assert.Equal(t, 5000, strategy.GetDefaultPort())
	})

	t.Run("GetStrategy for athens", func(t *testing.T) {
		strategy, err := factory.GetStrategy("athens")
		require.NoError(t, err)
		assert.NotNil(t, strategy)
		assert.Equal(t, 3000, strategy.GetDefaultPort())
	})

	t.Run("GetStrategy for devpi", func(t *testing.T) {
		strategy, err := factory.GetStrategy("devpi")
		require.NoError(t, err)
		assert.NotNil(t, strategy)
		assert.Equal(t, 3141, strategy.GetDefaultPort())
	})

	t.Run("GetStrategy for unsupported type", func(t *testing.T) {
		_, err := factory.GetStrategy("unknown")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported")
	})

	t.Run("CreateManager for zot registry", func(t *testing.T) {
		reg := &models.Registry{
			ID:   1,
			Name: "my-zot",
			Type: "zot",
			Port: 5001,
		}

		manager, err := factory.CreateManager(reg)
		require.NoError(t, err)
		assert.NotNil(t, manager)

		// Verify endpoint
		endpoint := manager.GetEndpoint()
		assert.Equal(t, "localhost:5001", endpoint)
	})

	t.Run("CreateManager for athens registry", func(t *testing.T) {
		reg := &models.Registry{
			ID:   2,
			Name: "my-athens",
			Type: "athens",
			Port: 3001,
		}

		manager, err := factory.CreateManager(reg)
		require.NoError(t, err)
		assert.NotNil(t, manager)

		// Athens returns full URL
		endpoint := manager.GetEndpoint()
		assert.Contains(t, endpoint, "3001")
	})

	t.Run("CreateManager for unsupported type", func(t *testing.T) {
		reg := &models.Registry{
			Name: "unknown",
			Type: "unknown",
		}

		_, err := factory.CreateManager(reg)
		assert.Error(t, err)
	})
}

// TestServiceManagerLifecycle tests the common lifecycle methods
func TestServiceManagerLifecycle(t *testing.T) {
	// This test uses a mock manager to verify the ServiceManager interface
	// In real usage, we'd test against actual ZotManager/AthensManager implementations

	t.Run("ServiceManager interface methods", func(t *testing.T) {
		ctx := context.Background()

		// Create a mock manager
		mock := &MockServiceManager{
			StartFunc: func(ctx context.Context) error {
				return nil
			},
			StopFunc: func(ctx context.Context) error {
				return nil
			},
			IsRunningFunc: func(ctx context.Context) bool {
				return true
			},
			GetEndpointFunc: func() string {
				return "localhost:5000"
			},
		}

		// Test Start
		err := mock.Start(ctx)
		assert.NoError(t, err)

		// Test IsRunning
		running := mock.IsRunning(ctx)
		assert.True(t, running)

		// Test GetEndpoint
		endpoint := mock.GetEndpoint()
		assert.Equal(t, "localhost:5000", endpoint)

		// Test Stop
		err = mock.Stop(ctx)
		assert.NoError(t, err)
	})
}

// MockServiceManager is a mock implementation for testing
type MockServiceManager struct {
	StartFunc       func(ctx context.Context) error
	StopFunc        func(ctx context.Context) error
	IsRunningFunc   func(ctx context.Context) bool
	GetEndpointFunc func() string
}

func (m *MockServiceManager) Start(ctx context.Context) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx)
	}
	return nil
}

func (m *MockServiceManager) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	return nil
}

func (m *MockServiceManager) IsRunning(ctx context.Context) bool {
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc(ctx)
	}
	return false
}

func (m *MockServiceManager) GetEndpoint() string {
	if m.GetEndpointFunc != nil {
		return m.GetEndpointFunc()
	}
	return ""
}
