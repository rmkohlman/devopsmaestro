package db

import (
	"devopsmaestro/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TDD Phase 2 (RED) — Registry Version Column
// =============================================================================
//
// These tests validate that the `version` column exists in the registries table
// and is correctly handled by all CRUD operations. They will NOT COMPILE until:
//   - models.Registry gains a Version field (RC-5)
//   - Migration 008 adds the column to the schema (RC-3)
//   - store_registry.go includes version in all SQL queries (RC-5)
//   - The test schema in store_test.go is updated to include version
//
// Related: migrations/sqlite/008_add_registry_version.{up,down}.sql
// =============================================================================

// TestCreateRegistry_WithVersion verifies that a registry created with a
// Version value persists that value and can be read back.
func TestCreateRegistry_WithVersion(t *testing.T) {
	ds := createTestDataStore(t)

	reg := &models.Registry{
		Name:      "zot-versioned",
		Type:      "zot",
		Port:      5600,
		Lifecycle: "persistent",
		Storage:   "/var/lib/zot",
		Version:   "2.1.15",
	}

	err := ds.CreateRegistry(reg)
	require.NoError(t, err, "CreateRegistry should succeed with Version set")
	require.NotZero(t, reg.ID, "ID should be assigned after creation")

	// Read back and verify the version persisted
	retrieved, err := ds.GetRegistryByID(reg.ID)
	require.NoError(t, err)
	assert.Equal(t, "2.1.15", retrieved.Version, "Version should be persisted and returned")
}

// TestGetRegistryByName_ReturnsVersion verifies that GetRegistryByName
// populates the Version field from the database.
func TestGetRegistryByName_ReturnsVersion(t *testing.T) {
	ds := createTestDataStore(t)

	reg := &models.Registry{
		Name:      "athens-versioned",
		Type:      "athens",
		Port:      3700,
		Lifecycle: "on-demand",
		Storage:   "/var/lib/athens",
		Version:   "0.13.0",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	retrieved, err := ds.GetRegistryByName("athens-versioned")
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, "athens-versioned", retrieved.Name)
	assert.Equal(t, "athens", retrieved.Type)
	assert.Equal(t, "0.13.0", retrieved.Version, "GetRegistryByName should return the Version field")
}

// TestGetRegistryByID_ReturnsVersion verifies that GetRegistryByID
// populates the Version field from the database.
func TestGetRegistryByID_ReturnsVersion(t *testing.T) {
	ds := createTestDataStore(t)

	reg := &models.Registry{
		Name:      "devpi-versioned",
		Type:      "devpi",
		Port:      3800,
		Lifecycle: "manual",
		Storage:   "/var/lib/devpi",
		Version:   "6.10.0",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)
	require.NotZero(t, reg.ID)

	retrieved, err := ds.GetRegistryByID(reg.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, reg.ID, retrieved.ID)
	assert.Equal(t, "devpi-versioned", retrieved.Name)
	assert.Equal(t, "6.10.0", retrieved.Version, "GetRegistryByID should return the Version field")
}

// TestUpdateRegistry_UpdatesVersion verifies that updating a registry's
// Version field persists the change.
func TestUpdateRegistry_UpdatesVersion(t *testing.T) {
	ds := createTestDataStore(t)

	// Create with initial version
	reg := &models.Registry{
		Name:      "zot-upgrade",
		Type:      "zot",
		Port:      5700,
		Lifecycle: "persistent",
		Storage:   "/var/lib/zot",
		Version:   "2.0.0",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	// Verify initial version
	before, err := ds.GetRegistryByID(reg.ID)
	require.NoError(t, err)
	assert.Equal(t, "2.0.0", before.Version, "Initial version should be 2.0.0")

	// Update to new version
	reg.Version = "2.1.15"
	err = ds.UpdateRegistry(reg)
	require.NoError(t, err, "UpdateRegistry should succeed when changing Version")

	// Verify version was updated
	after, err := ds.GetRegistryByID(reg.ID)
	require.NoError(t, err)
	assert.Equal(t, "2.1.15", after.Version, "Version should be updated to 2.1.15")
}

// TestListRegistries_IncludesVersion verifies that ListRegistries returns
// the Version field for every registry in the result set.
func TestListRegistries_IncludesVersion(t *testing.T) {
	ds := createTestDataStore(t)

	registries := []*models.Registry{
		{
			Name:      "list-ver-zot",
			Type:      "zot",
			Port:      5800,
			Lifecycle: "persistent",
			Storage:   "/var/lib/zot",
			Version:   "2.1.15",
		},
		{
			Name:      "list-ver-athens",
			Type:      "athens",
			Port:      3900,
			Lifecycle: "on-demand",
			Storage:   "/var/lib/athens",
			Version:   "0.13.0",
		},
	}

	for _, reg := range registries {
		err := ds.CreateRegistry(reg)
		require.NoError(t, err)
	}

	results, err := ds.ListRegistries()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "Should return at least 2 registries")

	// Build lookup by name
	found := make(map[string]*models.Registry)
	for _, r := range results {
		found[r.Name] = r
	}

	r1, ok := found["list-ver-zot"]
	require.True(t, ok, "list-ver-zot should be in results")
	assert.Equal(t, "2.1.15", r1.Version, "list-ver-zot should have version 2.1.15")

	r2, ok := found["list-ver-athens"]
	require.True(t, ok, "list-ver-athens should be in results")
	assert.Equal(t, "0.13.0", r2.Version, "list-ver-athens should have version 0.13.0")
}

// TestListRegistriesByType_IncludesVersion verifies that ListRegistriesByType
// returns the Version field for filtered registries.
func TestListRegistriesByType_IncludesVersion(t *testing.T) {
	ds := createTestDataStore(t)

	registries := []*models.Registry{
		{
			Name:      "type-ver-zot-1",
			Type:      "zot",
			Port:      5900,
			Lifecycle: "persistent",
			Storage:   "/var/lib/zot-1",
			Version:   "2.1.15",
		},
		{
			Name:      "type-ver-zot-2",
			Type:      "zot",
			Port:      5901,
			Lifecycle: "manual",
			Storage:   "/var/lib/zot-2",
			Version:   "2.0.0",
		},
		{
			Name:      "type-ver-athens",
			Type:      "athens",
			Port:      3950,
			Lifecycle: "on-demand",
			Storage:   "/var/lib/athens",
			Version:   "0.13.0",
		},
	}

	for _, reg := range registries {
		err := ds.CreateRegistry(reg)
		require.NoError(t, err)
	}

	// Filter by "zot" — should get 2 results
	results, err := ds.ListRegistriesByType("zot")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "Should return at least 2 zot registries")

	found := make(map[string]*models.Registry)
	for _, r := range results {
		found[r.Name] = r
	}

	r1, ok := found["type-ver-zot-1"]
	require.True(t, ok, "type-ver-zot-1 should be in results")
	assert.Equal(t, "2.1.15", r1.Version, "type-ver-zot-1 should have version 2.1.15")

	r2, ok := found["type-ver-zot-2"]
	require.True(t, ok, "type-ver-zot-2 should be in results")
	assert.Equal(t, "2.0.0", r2.Version, "type-ver-zot-2 should have version 2.0.0")
}

// TestListRegistriesByStatus_IncludesVersion verifies that ListRegistriesByStatus
// returns the Version field for filtered registries.
func TestListRegistriesByStatus_IncludesVersion(t *testing.T) {
	ds := createTestDataStore(t)

	registries := []*models.Registry{
		{
			Name:    "status-ver-running-1",
			Type:    "zot",
			Port:    6000,
			Status:  "running",
			Storage: "/var/lib/zot",
			Version: "2.1.15",
		},
		{
			Name:    "status-ver-running-2",
			Type:    "verdaccio",
			Port:    4900,
			Status:  "running",
			Storage: "/var/lib/verdaccio",
			Version: "5.31.0",
		},
		{
			Name:    "status-ver-stopped",
			Type:    "devpi",
			Port:    3960,
			Status:  "stopped",
			Storage: "/var/lib/devpi",
			Version: "6.10.0",
		},
	}

	for _, reg := range registries {
		err := ds.CreateRegistry(reg)
		require.NoError(t, err)
	}

	// Filter by "running" — should get 2 results
	results, err := ds.ListRegistriesByStatus("running")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "Should return at least 2 running registries")

	found := make(map[string]*models.Registry)
	for _, r := range results {
		found[r.Name] = r
	}

	r1, ok := found["status-ver-running-1"]
	require.True(t, ok, "status-ver-running-1 should be in results")
	assert.Equal(t, "2.1.15", r1.Version, "status-ver-running-1 should have version 2.1.15")

	r2, ok := found["status-ver-running-2"]
	require.True(t, ok, "status-ver-running-2 should be in results")
	assert.Equal(t, "5.31.0", r2.Version, "status-ver-running-2 should have version 5.31.0")
}

// TestCreateRegistry_EmptyVersion verifies that creating a registry with an
// empty Version string is allowed and persists correctly. Non-zot registry
// types may not have a meaningful version, so empty string is valid.
func TestCreateRegistry_EmptyVersion(t *testing.T) {
	ds := createTestDataStore(t)

	reg := &models.Registry{
		Name:      "squid-no-version",
		Type:      "squid",
		Port:      3200,
		Lifecycle: "manual",
		Storage:   "/var/cache/squid",
		Version:   "",
	}

	err := ds.CreateRegistry(reg)
	require.NoError(t, err, "CreateRegistry should succeed with empty Version")
	require.NotZero(t, reg.ID)

	// Read back and verify empty string persists (not NULL, not some default)
	retrieved, err := ds.GetRegistryByID(reg.ID)
	require.NoError(t, err)
	assert.Equal(t, "", retrieved.Version, "Empty version should persist as empty string")
}
