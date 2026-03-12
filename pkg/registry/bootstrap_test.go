package registry

import (
	"context"
	"errors"
	"testing"

	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Inline mocks for RegistryStore and DefaultsStore
//
// We use fine-grained function-hook mocks here rather than the shared
// MockDataStore so each test case can precisely assert which methods were
// called, with which arguments, and inject arbitrary errors.
// =============================================================================

// mockRegistryStore is a test-local implementation of db.RegistryStore.
type mockRegistryStore struct {
	// GetRegistryByNameFunc controls what GetRegistryByName returns.
	// If nil, returns a "not found" error (registry absent).
	GetRegistryByNameFunc func(name string) (*models.Registry, error)

	// CreateRegistryFunc controls what CreateRegistry returns.
	// If nil, records the call and returns nil.
	CreateRegistryFunc func(registry *models.Registry) error

	// Recorded calls for assertion.
	createRegistryCalls []*models.Registry
}

func (m *mockRegistryStore) GetRegistryByName(name string) (*models.Registry, error) {
	if m.GetRegistryByNameFunc != nil {
		return m.GetRegistryByNameFunc(name)
	}
	return nil, errors.New("registry '" + name + "' not found")
}

func (m *mockRegistryStore) CreateRegistry(registry *models.Registry) error {
	if m.CreateRegistryFunc != nil {
		return m.CreateRegistryFunc(registry)
	}
	m.createRegistryCalls = append(m.createRegistryCalls, registry)
	return nil
}

// Satisfy remaining RegistryStore interface methods (unused in bootstrap).
func (m *mockRegistryStore) GetRegistryByID(id int) (*models.Registry, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRegistryStore) GetRegistryByPort(port int) (*models.Registry, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRegistryStore) UpdateRegistry(registry *models.Registry) error {
	return errors.New("not implemented")
}
func (m *mockRegistryStore) DeleteRegistry(name string) error {
	return errors.New("not implemented")
}
func (m *mockRegistryStore) ListRegistries() ([]*models.Registry, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRegistryStore) ListRegistriesByType(registryType string) ([]*models.Registry, error) {
	return nil, errors.New("not implemented")
}
func (m *mockRegistryStore) ListRegistriesByStatus(status string) ([]*models.Registry, error) {
	return nil, errors.New("not implemented")
}

// mockDefaultsStore is a test-local implementation of db.DefaultsStore.
type mockDefaultsStore struct {
	// SetDefaultFunc controls what SetDefault returns.
	// If nil, records the call and returns nil.
	SetDefaultFunc func(key, value string) error

	// setDefaultCalls records each (key, value) pair passed to SetDefault.
	setDefaultCalls []struct{ Key, Value string }
}

func (m *mockDefaultsStore) GetDefault(key string) (string, error) {
	return "", nil
}

func (m *mockDefaultsStore) SetDefault(key, value string) error {
	if m.SetDefaultFunc != nil {
		return m.SetDefaultFunc(key, value)
	}
	m.setDefaultCalls = append(m.setDefaultCalls, struct{ Key, Value string }{key, value})
	return nil
}

func (m *mockDefaultsStore) DeleteDefault(key string) error {
	return nil
}

func (m *mockDefaultsStore) ListDefaults() (map[string]string, error) {
	return nil, nil
}

// =============================================================================
// Tests for EnsureDefaultRegistry
// =============================================================================

// TestEnsureDefaultRegistry_CreatesNewRegistry verifies that when no registry
// exists for the resolved type, a registry is created with correct defaults,
// set as the type default, and (true, nil) is returned.
func TestEnsureDefaultRegistry_CreatesNewRegistry(t *testing.T) {
	ctx := context.Background()

	registryStore := &mockRegistryStore{
		// GetRegistryByName returns "not found" (the default behaviour when func is nil).
	}
	defaultsStore := &mockDefaultsStore{}

	created, err := EnsureDefaultRegistry(ctx, registryStore, defaultsStore, AliasOCI, "on-demand")
	require.NoError(t, err)
	assert.True(t, created, "should return true when a new registry was created")

	// Verify CreateRegistry was called exactly once.
	require.Len(t, registryStore.createRegistryCalls, 1, "CreateRegistry must be called once")

	reg := registryStore.createRegistryCalls[0]

	// The alias "oci" resolves to concrete type "zot".
	assert.Equal(t, TypeZot, reg.Name, "registry name should be the concrete type name")
	assert.Equal(t, TypeZot, reg.Type, "registry type should be zot")
	assert.Equal(t, 5001, reg.Port, "registry port should be the zot default (5001)")
	assert.Equal(t, "on-demand", reg.Lifecycle, "lifecycle should match the argument")
	assert.True(t, reg.Enabled, "registry should be enabled")
	assert.Equal(t, "stopped", reg.Status, "initial status should be stopped")

	// Verify SetDefault was called to register this as the default for "oci".
	require.Len(t, defaultsStore.setDefaultCalls, 1, "SetDefault must be called once")
	assert.Equal(t, DefaultKeyOCI, defaultsStore.setDefaultCalls[0].Key,
		"SetDefault key should be the OCI defaults key")
	assert.Equal(t, TypeZot, defaultsStore.setDefaultCalls[0].Value,
		"SetDefault value should be the registry name (zot)")
}

// TestEnsureDefaultRegistry_SkipsCreationWhenExists verifies that when a
// registry with the resolved type name already exists, CreateRegistry is NOT
// called, SetDefault IS called, and (false, nil) is returned.
func TestEnsureDefaultRegistry_SkipsCreationWhenExists(t *testing.T) {
	ctx := context.Background()

	existing := &models.Registry{
		ID:        1,
		Name:      TypeZot,
		Type:      TypeZot,
		Port:      5001,
		Enabled:   true,
		Status:    "stopped",
		Lifecycle: "on-demand",
		Storage:   "/var/lib/zot",
	}

	registryStore := &mockRegistryStore{
		GetRegistryByNameFunc: func(name string) (*models.Registry, error) {
			if name == TypeZot {
				return existing, nil
			}
			return nil, errors.New("registry '" + name + "' not found")
		},
	}
	defaultsStore := &mockDefaultsStore{}

	created, err := EnsureDefaultRegistry(ctx, registryStore, defaultsStore, AliasOCI, "on-demand")
	require.NoError(t, err)
	assert.False(t, created, "should return false when registry already exists")

	// CreateRegistry must NOT have been called.
	assert.Empty(t, registryStore.createRegistryCalls,
		"CreateRegistry must not be called when registry already exists")

	// SetDefault must still have been called to (re-)register the default.
	require.Len(t, defaultsStore.setDefaultCalls, 1, "SetDefault must be called even when registry exists")
	assert.Equal(t, DefaultKeyOCI, defaultsStore.setDefaultCalls[0].Key)
	assert.Equal(t, TypeZot, defaultsStore.setDefaultCalls[0].Value)
}

// TestEnsureDefaultRegistry_ErrorOnUnknownAlias verifies that passing an alias
// that does not map to a known type causes EnsureDefaultRegistry to return an error.
func TestEnsureDefaultRegistry_ErrorOnUnknownAlias(t *testing.T) {
	ctx := context.Background()

	// "invalid" is not a known alias and is not a valid registry type.
	registryStore := &mockRegistryStore{}
	defaultsStore := &mockDefaultsStore{}

	_, err := EnsureDefaultRegistry(ctx, registryStore, defaultsStore, "invalid", "on-demand")
	assert.Error(t, err, "should return an error for an unknown alias")
}

// TestEnsureDefaultRegistry_ErrorWhenCreateFails verifies that when
// CreateRegistry returns an error, EnsureDefaultRegistry propagates it.
func TestEnsureDefaultRegistry_ErrorWhenCreateFails(t *testing.T) {
	ctx := context.Background()

	createErr := errors.New("disk full")

	registryStore := &mockRegistryStore{
		// No GetRegistryByNameFunc → returns "not found", so creation will be attempted.
		CreateRegistryFunc: func(registry *models.Registry) error {
			return createErr
		},
	}
	defaultsStore := &mockDefaultsStore{}

	_, err := EnsureDefaultRegistry(ctx, registryStore, defaultsStore, AliasOCI, "on-demand")
	require.Error(t, err, "error from CreateRegistry must be propagated")
	assert.ErrorIs(t, err, createErr, "propagated error should wrap or equal the original error")

	// SetDefault must NOT have been called if creation failed.
	assert.Empty(t, defaultsStore.setDefaultCalls,
		"SetDefault must not be called when CreateRegistry fails")
}

// TestEnsureDefaultRegistry_ErrorWhenSetDefaultFails verifies that when
// CreateRegistry succeeds but SetDefault fails, the error is propagated.
func TestEnsureDefaultRegistry_ErrorWhenSetDefaultFails(t *testing.T) {
	ctx := context.Background()

	setErr := errors.New("db write error")

	registryStore := &mockRegistryStore{
		// No GetRegistryByNameFunc → "not found", creation will be attempted.
		// No CreateRegistryFunc → records call and returns nil (success).
	}
	defaultsStore := &mockDefaultsStore{
		SetDefaultFunc: func(key, value string) error {
			return setErr
		},
	}

	_, err := EnsureDefaultRegistry(ctx, registryStore, defaultsStore, AliasOCI, "on-demand")
	require.Error(t, err, "error from SetDefault must be propagated")
	assert.ErrorIs(t, err, setErr, "propagated error should wrap or equal the original error")
}
