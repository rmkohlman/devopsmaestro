package registry

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/registry/envinjector"
)

// =============================================================================
// Test Helpers / Mocks
// =============================================================================

// MockManagerFactory is a mock implementation of ManagerFactory for testing.
type MockManagerFactory struct {
	CreateManagerFunc func(reg *models.Registry) (ServiceManager, error)
}

func (m *MockManagerFactory) CreateManager(reg *models.Registry) (ServiceManager, error) {
	if m.CreateManagerFunc != nil {
		return m.CreateManagerFunc(reg)
	}
	return &MockServiceManager{}, nil
}

// newTestCoordinator creates a BuildRegistryCoordinator with the provided store
// and an optionally-customised factory and real EnvironmentInjector.
func newTestCoordinator(store db.DataStore, factory ManagerFactory) *BuildRegistryCoordinator {
	if factory == nil {
		factory = &MockManagerFactory{}
	}
	return NewBuildRegistryCoordinator(store, factory, envinjector.NewEnvironmentInjector())
}

// makeRegistry builds a Registry model with sensible defaults for a given type.
func makeRegistry(name, regType, lifecycle string, enabled bool, port int) *models.Registry {
	return &models.Registry{
		ID:        1,
		Name:      name,
		Type:      regType,
		Enabled:   enabled,
		Port:      port,
		Lifecycle: lifecycle,
		Storage:   "/var/lib/" + regType,
		Status:    "stopped",
	}
}

// populateMockStore inserts the provided registries into the mock store.
func populateMockStore(store *db.MockDataStore, registries ...*models.Registry) {
	for i, reg := range registries {
		reg.ID = i + 1
		store.Registries[reg.Name] = reg
	}
}

// =============================================================================
// Test 1: No registries in the store → empty result
// =============================================================================

func TestBuildRegistryCoordinator_NoRegistries(t *testing.T) {
	store := db.NewMockDataStore()
	c := newTestCoordinator(store, nil)

	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Managers, "no managers should be created when store is empty")
	assert.Empty(t, result.Registries, "no registries should be returned when store is empty")
	assert.Empty(t, result.EnvVars, "no env vars should be set when store is empty")
	assert.Empty(t, result.OCIEndpoint, "OCI endpoint should be empty when no zot registry exists")
	assert.Empty(t, result.Warnings, "no warnings should be produced for an empty store")
}

// =============================================================================
// Test 2: All registries disabled → empty result
// =============================================================================

func TestBuildRegistryCoordinator_AllDisabled(t *testing.T) {
	store := db.NewMockDataStore()
	populateMockStore(store,
		makeRegistry("go-proxy", "athens", "on-demand", false, 3000),
		makeRegistry("pip-proxy", "devpi", "on-demand", false, 3141),
		makeRegistry("npm-proxy", "verdaccio", "persistent", false, 4873),
	)

	factoryCalled := false
	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			factoryCalled = true
			return &MockServiceManager{}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Managers, "disabled registries should not create managers")
	assert.Empty(t, result.Registries, "disabled registries should not appear in result")
	assert.Empty(t, result.EnvVars, "disabled registries should not produce env vars")
	assert.False(t, factoryCalled, "factory should not be called for disabled registries")
}

// =============================================================================
// Test 3: Lifecycle "manual" is skipped even when enabled
// =============================================================================

func TestBuildRegistryCoordinator_ManualLifecycleSkipped(t *testing.T) {
	store := db.NewMockDataStore()
	populateMockStore(store,
		makeRegistry("manual-reg", "athens", "manual", true, 3000),
	)

	factoryCalled := false
	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			factoryCalled = true
			return &MockServiceManager{}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.Managers, "manual lifecycle should be skipped")
	assert.Empty(t, result.EnvVars, "manual lifecycle should produce no env vars")
	assert.False(t, factoryCalled, "factory should not be called for manual lifecycle")
}

// =============================================================================
// Test 4: on-demand Athens registry is started and GOPROXY is injected
// =============================================================================

func TestBuildRegistryCoordinator_OnDemandStarted(t *testing.T) {
	store := db.NewMockDataStore()
	athensReg := makeRegistry("go-proxy", "athens", "on-demand", true, 3000)
	populateMockStore(store, athensReg)

	startCalled := false
	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			return &MockServiceManager{
				StartFunc: func(ctx context.Context) error {
					startCalled = true
					return nil
				},
				GetEndpointFunc: func() string {
					return "http://host.docker.internal:3000"
				},
			}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, startCalled, "Start() must be called on the manager")
	assert.Len(t, result.Managers, 1, "exactly one manager should be in result")
	assert.Len(t, result.Registries, 1, "exactly one registry should be in result")
	assert.Contains(t, result.EnvVars, "GOPROXY", "GOPROXY must be set for Athens")
	assert.Contains(t, result.EnvVars["GOPROXY"], "3000", "GOPROXY value should include the Athens port")
}

// =============================================================================
// Test 5: persistent Verdaccio registry is started and NPM env is injected
// =============================================================================

func TestBuildRegistryCoordinator_PersistentStarted(t *testing.T) {
	store := db.NewMockDataStore()
	verdaccioReg := makeRegistry("npm-proxy", "verdaccio", "persistent", true, 4873)
	populateMockStore(store, verdaccioReg)

	startCalled := false
	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			return &MockServiceManager{
				StartFunc: func(ctx context.Context) error {
					startCalled = true
					return nil
				},
				GetEndpointFunc: func() string {
					return "http://host.docker.internal:4873"
				},
			}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.True(t, startCalled, "Start() must be called on the manager")
	assert.Len(t, result.Managers, 1, "exactly one manager should be in result")
	// Verdaccio injects NPM_CONFIG_REGISTRY and npm_config_registry
	assert.Contains(t, result.EnvVars, "NPM_CONFIG_REGISTRY", "NPM_CONFIG_REGISTRY must be set for Verdaccio")
	assert.Contains(t, result.EnvVars["NPM_CONFIG_REGISTRY"], "4873", "NPM_CONFIG_REGISTRY should include the Verdaccio port")
}

// =============================================================================
// Test 6: Start failure adds a warning; other registries still processed
// =============================================================================

func TestBuildRegistryCoordinator_StartFailureWarns(t *testing.T) {
	store := db.NewMockDataStore()
	successReg := makeRegistry("go-proxy", "athens", "on-demand", true, 3000)
	failReg := makeRegistry("npm-proxy", "verdaccio", "on-demand", true, 4873)
	populateMockStore(store, successReg, failReg)

	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			if reg.Name == "npm-proxy" {
				return &MockServiceManager{
					StartFunc: func(ctx context.Context) error {
						return fmt.Errorf("verdaccio failed to start: binary not found")
					},
				}, nil
			}
			// "go-proxy" succeeds
			return &MockServiceManager{
				StartFunc: func(ctx context.Context) error { return nil },
				GetEndpointFunc: func() string {
					return "http://host.docker.internal:3000"
				},
			}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	// Should return result, not error — failure is a warning
	require.NoError(t, err, "a single registry failure should not return an error")
	require.NotNil(t, result)

	assert.Len(t, result.Managers, 1, "only the successful manager should be in result")
	assert.Len(t, result.Warnings, 1, "one warning should be recorded for the failed registry")
	assert.Contains(t, result.Warnings[0], "npm-proxy", "warning should identify the failed registry")
	assert.Contains(t, result.EnvVars, "GOPROXY", "successful Athens registry should still inject env vars")
	assert.NotContains(t, result.EnvVars, "NPM_CONFIG_REGISTRY", "failed Verdaccio should not inject env vars")
}

// =============================================================================
// Test 7: Zot registry sets OCIEndpoint
// =============================================================================

func TestBuildRegistryCoordinator_OCIEndpointSet(t *testing.T) {
	store := db.NewMockDataStore()
	zotReg := makeRegistry("oci-registry", "zot", "on-demand", true, 5000)
	populateMockStore(store, zotReg)

	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			return &MockServiceManager{
				StartFunc: func(ctx context.Context) error { return nil },
				GetEndpointFunc: func() string {
					return "host.docker.internal:5000"
				},
			}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NotEmpty(t, result.OCIEndpoint, "OCIEndpoint must be set for a Zot registry")
	assert.Contains(t, result.OCIEndpoint, "5000", "OCIEndpoint should include the Zot port")
}

// =============================================================================
// Test 8: Non-Zot registries do not set OCIEndpoint
// =============================================================================

func TestBuildRegistryCoordinator_NoOCIRegistry(t *testing.T) {
	store := db.NewMockDataStore()
	// Athens only — no Zot registry
	athensReg := makeRegistry("go-proxy", "athens", "on-demand", true, 3000)
	populateMockStore(store, athensReg)

	c := newTestCoordinator(store, &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			return &MockServiceManager{
				StartFunc:       func(ctx context.Context) error { return nil },
				GetEndpointFunc: func() string { return "http://host.docker.internal:3000" },
			}, nil
		},
	})

	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Empty(t, result.OCIEndpoint, "OCIEndpoint must be empty when no Zot registry is present")
}

// =============================================================================
// Test 9: Multiple registries (Athens + Devpi + Verdaccio) — all started
// =============================================================================

func TestBuildRegistryCoordinator_MultipleRegistries(t *testing.T) {
	store := db.NewMockDataStore()
	populateMockStore(store,
		makeRegistry("go-proxy", "athens", "on-demand", true, 3000),
		makeRegistry("pip-proxy", "devpi", "on-demand", true, 3141),
		makeRegistry("npm-proxy", "verdaccio", "on-demand", true, 4873),
	)

	startCount := 0
	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			return &MockServiceManager{
				StartFunc: func(ctx context.Context) error {
					startCount++
					return nil
				},
				GetEndpointFunc: func() string {
					return fmt.Sprintf("http://host.docker.internal:%d", reg.Port)
				},
			}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 3, startCount, "all three managers should have been started")
	assert.Len(t, result.Managers, 3, "result should contain all three managers")
	assert.Len(t, result.Registries, 3, "result should reference all three registries")
	assert.Empty(t, result.Warnings, "no warnings expected when all registries start successfully")

	// Check each registry type injected its env vars
	assert.Contains(t, result.EnvVars, "GOPROXY", "Athens should inject GOPROXY")
	assert.Contains(t, result.EnvVars, "PIP_INDEX_URL", "Devpi should inject PIP_INDEX_URL")
	assert.Contains(t, result.EnvVars, "NPM_CONFIG_REGISTRY", "Verdaccio should inject NPM_CONFIG_REGISTRY")
}

// =============================================================================
// Test 10: Mixed enabled/disabled — only enabled registries processed
// =============================================================================

func TestBuildRegistryCoordinator_MixedEnabledDisabled(t *testing.T) {
	store := db.NewMockDataStore()
	populateMockStore(store,
		makeRegistry("go-proxy", "athens", "on-demand", true, 3000),      // enabled
		makeRegistry("pip-proxy", "devpi", "on-demand", false, 3141),     // disabled
		makeRegistry("npm-proxy", "verdaccio", "persistent", true, 4873), // enabled
		makeRegistry("oci-reg", "zot", "on-demand", false, 5000),         // disabled
		makeRegistry("squid-prx", "squid", "manual", false, 3128),        // disabled + manual
	)

	processedNames := []string{}
	factory := &MockManagerFactory{
		CreateManagerFunc: func(reg *models.Registry) (ServiceManager, error) {
			processedNames = append(processedNames, reg.Name)
			return &MockServiceManager{
				StartFunc:       func(ctx context.Context) error { return nil },
				GetEndpointFunc: func() string { return fmt.Sprintf("http://host.docker.internal:%d", reg.Port) },
			}, nil
		},
	}

	c := newTestCoordinator(store, factory)
	result, err := c.Prepare(context.Background())

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Len(t, result.Managers, 2, "only the 2 enabled+auto-startable registries should be processed")
	assert.Len(t, result.Registries, 2, "result should reference exactly 2 registries")
	assert.Empty(t, result.Warnings)

	// The two enabled, auto-startable registries should have been processed
	assert.Contains(t, processedNames, "go-proxy")
	assert.Contains(t, processedNames, "npm-proxy")
	// Disabled and manual ones must NOT have been processed
	assert.NotContains(t, processedNames, "pip-proxy")
	assert.NotContains(t, processedNames, "oci-reg")
	assert.NotContains(t, processedNames, "squid-prx")
}

// =============================================================================
// Test 11: ListRegistries error propagates
// =============================================================================

func TestBuildRegistryCoordinator_ListRegistriesError(t *testing.T) {
	store := db.NewMockDataStore()
	store.ListRegistriesErr = fmt.Errorf("database connection lost")

	c := newTestCoordinator(store, nil)
	result, err := c.Prepare(context.Background())

	assert.Error(t, err, "ListRegistries error must be propagated")
	assert.Nil(t, result, "result must be nil on error")
	assert.Contains(t, err.Error(), "database connection lost")
}
