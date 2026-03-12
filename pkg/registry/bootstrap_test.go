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

// =============================================================================
// TDD Phase 2 (RED): WI-5 — Bootstrap all 5 default registries on init
// =============================================================================

// aliasToDefaultKey maps each alias to its expected defaults store key.
// Used to assert SetDefault is called with the right key per alias.
var aliasToDefaultKey = map[string]string{
	AliasOCI:  DefaultKeyOCI,
	AliasPyPI: DefaultKeyPyPI,
	AliasNPM:  DefaultKeyNPM,
	AliasGo:   DefaultKeyGo,
	AliasHTTP: DefaultKeyHTTP,
}

// aliasToConcreteType maps each alias to the expected concrete registry type.
var aliasToConcreteType = map[string]string{
	AliasOCI:  TypeZot,
	AliasPyPI: TypeDevpi,
	AliasNPM:  TypeVerdaccio,
	AliasGo:   TypeAthens,
	AliasHTTP: TypeSquid,
}

// TestEnsureDefaultRegistry_AllAliases verifies that EnsureDefaultRegistry works
// correctly for each of the 5 supported aliases: "oci", "pypi", "npm", "go", "http".
// Each alias must resolve to the correct concrete type and set the right defaults key.
//
// This test WILL FAIL (RED) until EnsureDefaultRegistry has been verified to
// handle all 5 aliases end-to-end (WI-5: bootstrap all aliases on init).
func TestEnsureDefaultRegistry_AllAliases(t *testing.T) {
	allAliases := []string{AliasOCI, AliasPyPI, AliasNPM, AliasGo, AliasHTTP}

	for _, alias := range allAliases {
		alias := alias // capture for sub-test closure
		t.Run("alias="+alias, func(t *testing.T) {
			ctx := context.Background()

			registryStore := &mockRegistryStore{
				// Default: GetRegistryByName returns "not found", triggering creation.
			}
			defaultsStore := &mockDefaultsStore{}

			created, err := EnsureDefaultRegistry(ctx, registryStore, defaultsStore, alias, "on-demand")
			require.NoError(t, err, "alias %q must not return an error", alias)
			assert.True(t, created, "alias %q should create a new registry (created=true)", alias)

			// Verify CreateRegistry was called exactly once.
			require.Len(t, registryStore.createRegistryCalls, 1,
				"alias %q: CreateRegistry should be called exactly once", alias)

			reg := registryStore.createRegistryCalls[0]
			expectedType := aliasToConcreteType[alias]

			assert.Equal(t, expectedType, reg.Type,
				"alias %q: registry type should be %q", alias, expectedType)
			assert.Equal(t, expectedType, reg.Name,
				"alias %q: registry name should equal the concrete type %q", alias, expectedType)
			assert.True(t, reg.Enabled,
				"alias %q: registry should be enabled", alias)
			assert.Equal(t, "stopped", reg.Status,
				"alias %q: initial status should be stopped", alias)

			// Verify SetDefault was called with the correct defaults key.
			require.Len(t, defaultsStore.setDefaultCalls, 1,
				"alias %q: SetDefault must be called exactly once", alias)

			expectedKey := aliasToDefaultKey[alias]
			assert.Equal(t, expectedKey, defaultsStore.setDefaultCalls[0].Key,
				"alias %q: SetDefault key should be %q", alias, expectedKey)
			assert.Equal(t, expectedType, defaultsStore.setDefaultCalls[0].Value,
				"alias %q: SetDefault value should be %q", alias, expectedType)
		})
	}
}

// TestEnsureDefaultRegistry_AllIdempotent verifies that calling EnsureDefaultRegistry
// twice for each alias is idempotent: the second call returns (false, nil) and does
// not create an additional registry, resulting in exactly 5 registries total for 10 calls.
//
// This test WILL FAIL (RED) until the idempotency path works correctly for all 5 aliases.
func TestEnsureDefaultRegistry_AllIdempotent(t *testing.T) {
	ctx := context.Background()
	allAliases := []string{AliasOCI, AliasPyPI, AliasNPM, AliasGo, AliasHTTP}

	// Shared registry store that accumulates all created registries.
	// This simulates a real DB: after first call creates it, second call finds it.
	registryStore := &mockRegistryStore{}
	defaultsStore := &mockDefaultsStore{}

	// Track all created registries by name so the second call can find them.
	createdRegistries := make(map[string]*models.Registry)
	registryStore.CreateRegistryFunc = func(r *models.Registry) error {
		createdRegistries[r.Name] = r
		registryStore.createRegistryCalls = append(registryStore.createRegistryCalls, r)
		return nil
	}
	registryStore.GetRegistryByNameFunc = func(name string) (*models.Registry, error) {
		if r, ok := createdRegistries[name]; ok {
			return r, nil
		}
		return nil, errors.New("registry '" + name + "' not found")
	}

	// First pass: all 5 aliases should create new registries.
	for _, alias := range allAliases {
		created, err := EnsureDefaultRegistry(ctx, registryStore, defaultsStore, alias, "on-demand")
		require.NoError(t, err, "first call for alias %q should not error", alias)
		assert.True(t, created, "first call for alias %q should return created=true", alias)
	}

	// Verify 5 registries were created after the first pass.
	assert.Len(t, registryStore.createRegistryCalls, 5,
		"exactly 5 registries should be created after first pass")

	// Reset create call tracking for the second pass.
	registryStore.createRegistryCalls = nil

	// Second pass: all 5 aliases should find existing registries — no new creation.
	for _, alias := range allAliases {
		created, err := EnsureDefaultRegistry(ctx, registryStore, defaultsStore, alias, "on-demand")
		require.NoError(t, err, "second call for alias %q should not error", alias)
		assert.False(t, created, "second call for alias %q should return created=false (already exists)", alias)
	}

	// Verify no additional registries were created in the second pass.
	assert.Empty(t, registryStore.createRegistryCalls,
		"no registries should be created on the second pass (idempotency)")

	// Total unique registries in createdRegistries map must be exactly 5.
	assert.Len(t, createdRegistries, 5,
		"total unique registries should be exactly 5, not 10")
}

// TestEnsureDefaultRegistry_PartialFailure verifies that a failure for one alias
// does not prevent the others from succeeding. This tests that callers can iterate
// all aliases and that each is independent.
//
// NOTE: EnsureDefaultRegistry itself is per-alias and has no cross-alias logic.
// This test validates the CALLER PATTERN for WI-5: when bootstrapping all 5 aliases,
// a failure in one (e.g., "pypi" CreateRegistry fails) should not prevent the others.
// The test simulates calling EnsureDefaultRegistry in a loop and collecting errors.
//
// This test WILL FAIL (RED) until EnsureAllDefaultRegistries (or equivalent loop
// logic in cmd/init.go) is implemented to continue past partial failures.
func TestEnsureDefaultRegistry_PartialFailure(t *testing.T) {
	ctx := context.Background()

	pypiErr := errors.New("pypi creation failed: disk quota exceeded")

	registryStore := &mockRegistryStore{
		// Default: GetRegistryByName returns "not found" for all → triggers creation.
		CreateRegistryFunc: func(r *models.Registry) error {
			// Fail only for devpi (the pypi concrete type).
			if r.Type == TypeDevpi {
				return pypiErr
			}
			// All other types succeed.
			return nil
		},
	}
	defaultsStore := &mockDefaultsStore{}

	// Track which aliases succeeded and which failed.
	type result struct {
		created bool
		err     error
	}
	results := make(map[string]result)

	allAliases := []string{AliasOCI, AliasPyPI, AliasNPM, AliasGo, AliasHTTP}
	for _, alias := range allAliases {
		created, err := EnsureDefaultRegistry(ctx, registryStore, defaultsStore, alias, "on-demand")
		results[alias] = result{created: created, err: err}
	}

	// "pypi" should have failed.
	require.Error(t, results[AliasPyPI].err,
		"pypi alias should return an error when CreateRegistry fails")
	assert.False(t, results[AliasPyPI].created,
		"pypi alias should not be created on failure")

	// All others should have succeeded.
	successAliases := []string{AliasOCI, AliasNPM, AliasGo, AliasHTTP}
	for _, alias := range successAliases {
		assert.NoError(t, results[alias].err,
			"alias %q should succeed even though pypi failed", alias)
		assert.True(t, results[alias].created,
			"alias %q should be created (created=true)", alias)
	}
}

// =============================================================================
// End of WI-5 tests
// =============================================================================
