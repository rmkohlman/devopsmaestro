package db

import (
	"devopsmaestro/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DataStore Interface Tests for Registry Operations
// =============================================================================

func TestDataStore_CreateRegistry(t *testing.T) {
	ds := createTestDataStore(t)

	tests := []struct {
		name     string
		registry *models.Registry
		wantErr  bool
		errMsg   string
	}{
		{
			name: "create valid zot registry",
			registry: &models.Registry{
				Name:      "zot-registry",
				Type:      "zot",
				Port:      5000,
				Lifecycle: "persistent",
				Storage:   "/var/lib/zot",
			},
			wantErr: false,
		},
		{
			name: "create valid athens registry",
			registry: &models.Registry{
				Name:      "athens-proxy",
				Type:      "athens",
				Port:      3000,
				Lifecycle: "on-demand",
				Storage:   "/var/lib/athens",
			},
			wantErr: false,
		},
		{
			name: "create with duplicate name",
			registry: &models.Registry{
				Name:      "zot-registry", // Duplicate from first test
				Type:      "zot",
				Port:      5001,
				Lifecycle: "manual",
				Storage:   "/var/lib/zot-dup",
			},
			wantErr: true,
			errMsg:  "unique constraint violation",
		},
		{
			name: "create with invalid type",
			registry: &models.Registry{
				Name:      "bad-registry",
				Type:      "invalid",
				Port:      5000,
				Lifecycle: "manual",
				Storage:   "/var/lib/bad",
			},
			wantErr: true,
			errMsg:  "unsupported registry type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ds.CreateRegistry(tt.registry)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.registry.ID, "ID should be set after creation")
			}
		})
	}
}

func TestDataStore_GetRegistryByName(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registry
	reg := &models.Registry{
		Name:      "test-registry",
		Type:      "zot",
		Port:      5010, // Unique port to avoid conflicts
		Lifecycle: "persistent",
		Storage:   "/var/lib/zot",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	tests := []struct {
		name    string
		regName string
		wantErr bool
		wantReg bool
	}{
		{
			name:    "get existing registry",
			regName: "test-registry",
			wantErr: false,
			wantReg: true,
		},
		{
			name:    "get non-existent registry",
			regName: "does-not-exist",
			wantErr: true,
			wantReg: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ds.GetRegistryByName(tt.regName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantReg {
					require.NotNil(t, result)
					assert.Equal(t, tt.regName, result.Name)
					assert.Equal(t, "zot", result.Type)
					assert.Equal(t, 5010, result.Port)
					assert.Equal(t, "persistent", result.Lifecycle)
				}
			}
		})
	}
}

func TestDataStore_GetRegistryByID(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registry
	reg := &models.Registry{
		Name:      "test-registry-id",
		Type:      "verdaccio",
		Port:      4873,
		Lifecycle: "manual",
		Storage:   "/var/lib/verdaccio",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)
	require.NotZero(t, reg.ID)

	tests := []struct {
		name    string
		id      int
		wantErr bool
	}{
		{
			name:    "get by valid ID",
			id:      reg.ID,
			wantErr: false,
		},
		{
			name:    "get by non-existent ID",
			id:      99999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ds.GetRegistryByID(tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, reg.ID, result.ID)
				assert.Equal(t, reg.Name, result.Name)
			}
		})
	}
}

func TestDataStore_ListRegistries(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registries
	registries := []*models.Registry{
		{Name: "zot-list-1", Type: "zot", Port: 5020, Lifecycle: "persistent", Storage: "/var/lib/zot"},
		{Name: "athens-list-1", Type: "athens", Port: 3020, Lifecycle: "on-demand", Storage: "/var/lib/athens"},
		{Name: "devpi-list-1", Type: "devpi", Port: 3141, Lifecycle: "manual", Storage: "/var/lib/devpi"},
	}

	for _, reg := range registries {
		err := ds.CreateRegistry(reg)
		require.NoError(t, err)
	}

	// List all
	results, err := ds.ListRegistries()
	assert.NoError(t, err)
	assert.Len(t, results, 3)

	// Verify all are present
	names := make(map[string]bool)
	for _, r := range results {
		names[r.Name] = true
	}
	assert.True(t, names["zot-list-1"])
	assert.True(t, names["athens-list-1"])
	assert.True(t, names["devpi-list-1"])
}

func TestDataStore_UpdateRegistry(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registry
	reg := &models.Registry{
		Name:      "update-test",
		Type:      "zot",
		Port:      5030,
		Lifecycle: "manual",
		Storage:   "/var/lib/zot",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	// Update fields
	reg.Port = 5031
	reg.Lifecycle = "persistent"
	reg.Status = "running"

	err = ds.UpdateRegistry(reg)
	assert.NoError(t, err)

	// Verify update
	updated, err := ds.GetRegistryByID(reg.ID)
	require.NoError(t, err)
	assert.Equal(t, 5031, updated.Port)
	assert.Equal(t, "persistent", updated.Lifecycle)
	assert.Equal(t, "running", updated.Status)
}

func TestDataStore_DeleteRegistry(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registry
	reg := &models.Registry{
		Name:      "delete-test",
		Type:      "squid",
		Port:      3040,
		Lifecycle: "manual",
		Storage:   "/var/cache/squid",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	// Delete by name
	err = ds.DeleteRegistry(reg.Name)
	assert.NoError(t, err)

	// Verify deletion
	_, err = ds.GetRegistryByName(reg.Name)
	assert.Error(t, err)
	assert.True(t, IsNotFound(err), "error should be ErrNotFound")
}

func TestDataStore_GetRegistryByPort(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registries on different ports
	reg1 := &models.Registry{
		Name:      "reg-5050",
		Type:      "zot",
		Port:      5050,
		Lifecycle: "persistent",
		Storage:   "/var/lib/zot",
	}
	err := ds.CreateRegistry(reg1)
	require.NoError(t, err)

	reg2 := &models.Registry{
		Name:      "reg-5051",
		Type:      "zot",
		Port:      5051,
		Lifecycle: "persistent",
		Storage:   "/var/lib/zot-2",
	}
	err = ds.CreateRegistry(reg2)
	require.NoError(t, err)

	tests := []struct {
		name    string
		port    int
		wantErr bool
		wantReg string
	}{
		{
			name:    "find registry on port 5050",
			port:    5050,
			wantErr: false,
			wantReg: "reg-5050",
		},
		{
			name:    "find registry on port 5051",
			port:    5051,
			wantErr: false,
			wantReg: "reg-5051",
		},
		{
			name:    "no registry on port 9999",
			port:    9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ds.GetRegistryByPort(tt.port)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.wantReg, result.Name)
				assert.Equal(t, tt.port, result.Port)
			}
		})
	}
}

func TestDataStore_Registry_PortConflictDetection(t *testing.T) {
	ds := createTestDataStore(t)

	// Create first registry on port 5060
	reg1 := &models.Registry{
		Name:      "reg-conflict-1",
		Type:      "zot",
		Port:      5060,
		Lifecycle: "persistent",
		Storage:   "/var/lib/zot",
	}
	err := ds.CreateRegistry(reg1)
	require.NoError(t, err)

	// Try to create another registry on the same port
	reg2 := &models.Registry{
		Name:      "reg-conflict-2",
		Type:      "athens",
		Port:      5060, // Same port - should conflict
		Lifecycle: "on-demand",
		Storage:   "/var/lib/athens",
	}
	err = ds.CreateRegistry(reg2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port 5060 is already in use")
}

func TestDataStore_Registry_DefaultStatus(t *testing.T) {
	ds := createTestDataStore(t)

	reg := &models.Registry{
		Name:      "status-test",
		Type:      "zot",
		Port:      5070,
		Lifecycle: "manual",
		Storage:   "/var/lib/zot",
		// Status not explicitly set
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	// Retrieve and check default status
	result, err := ds.GetRegistryByID(reg.ID)
	require.NoError(t, err)
	assert.Equal(t, "stopped", result.Status, "Default status should be 'stopped'")
}

func TestDataStore_ListRegistriesByType(t *testing.T) {
	ds := createTestDataStore(t)

	// Create registries of different types
	registries := []*models.Registry{
		{Name: "zot-type-1", Type: "zot", Port: 5080, Lifecycle: "persistent", Storage: "/var/lib/zot-1"},
		{Name: "zot-type-2", Type: "zot", Port: 5081, Lifecycle: "manual", Storage: "/var/lib/zot-2"},
		{Name: "athens-type-1", Type: "athens", Port: 3080, Lifecycle: "on-demand", Storage: "/var/lib/athens"},
	}

	for _, reg := range registries {
		err := ds.CreateRegistry(reg)
		require.NoError(t, err)
	}

	// List by type
	zotRegs, err := ds.ListRegistriesByType("zot")
	assert.NoError(t, err)
	assert.Len(t, zotRegs, 2)

	athensRegs, err := ds.ListRegistriesByType("athens")
	assert.NoError(t, err)
	assert.Len(t, athensRegs, 1)

	// Non-existent type
	devpiRegs, err := ds.ListRegistriesByType("devpi")
	assert.NoError(t, err)
	assert.Len(t, devpiRegs, 0)
}

func TestDataStore_ListRegistriesByStatus(t *testing.T) {
	ds := createTestDataStore(t)

	// Create registries with different statuses
	registries := []*models.Registry{
		{Name: "running-status-1", Type: "zot", Port: 5090, Status: "running", Storage: "/var/lib/zot"},
		{Name: "running-status-2", Type: "athens", Port: 3090, Status: "running", Storage: "/var/lib/athens"},
		{Name: "stopped-status-1", Type: "devpi", Port: 3091, Status: "stopped", Storage: "/var/lib/devpi"},
	}

	for _, reg := range registries {
		err := ds.CreateRegistry(reg)
		require.NoError(t, err)
	}

	// List running registries
	running, err := ds.ListRegistriesByStatus("running")
	assert.NoError(t, err)
	assert.Len(t, running, 2)

	// List stopped registries
	stopped, err := ds.ListRegistriesByStatus("stopped")
	assert.NoError(t, err)
	assert.Len(t, stopped, 1)
}

// =============================================================================
// Integration Tests (Skipped - Require Full Runtime)
// =============================================================================

func TestDataStore_Registry_RuntimeIntegration(t *testing.T) {
	t.Skip("Integration test - requires runtime manager")

	// This test would verify:
	// 1. Creating registry triggers runtime manager
	// 2. Status updates propagate to database
	// 3. Container lifecycle matches database state
}

func TestDataStore_Registry_StartStop(t *testing.T) {
	t.Skip("Integration test - requires container runtime")

	// This test would verify:
	// 1. Start updates status to "running"
	// 2. Stop updates status to "stopped"
	// 3. Container ID is tracked correctly
}

// =============================================================================
// TDD Phase 2 (RED) - Tests for GitHub Issue #5
// Missing columns: storage, enabled, idle_timeout
// =============================================================================

func TestDataStore_CreateRegistry_WithStorageEnabledIdleTimeout(t *testing.T) {
	ds := createTestDataStore(t)

	tests := []struct {
		name        string
		registry    *models.Registry
		wantErr     bool
		errMsg      string
		checkFields bool
	}{
		{
			name: "create registry with all fields including storage, enabled, idle_timeout",
			registry: &models.Registry{
				Name:        "zot-full-fields",
				Type:        "zot",
				Port:        5100,
				Lifecycle:   "on-demand",
				Enabled:     true,
				Storage:     "/var/lib/zot",
				IdleTimeout: 1800,
			},
			wantErr:     false,
			checkFields: true,
		},
		{
			name: "create registry with enabled=false",
			registry: &models.Registry{
				Name:        "zot-disabled",
				Type:        "zot",
				Port:        5101,
				Lifecycle:   "manual",
				Enabled:     false,
				Storage:     "/custom/storage",
				IdleTimeout: 3600,
			},
			wantErr:     false,
			checkFields: true,
		},
		{
			name: "create registry with custom storage path",
			registry: &models.Registry{
				Name:        "athens-custom-storage",
				Type:        "athens",
				Port:        3100,
				Lifecycle:   "persistent",
				Enabled:     true,
				Storage:     "/mnt/registry/athens",
				IdleTimeout: 0,
			},
			wantErr:     false,
			checkFields: true,
		},
		{
			name: "create registry with on-demand and idle timeout",
			registry: &models.Registry{
				Name:        "devpi-on-demand",
				Type:        "devpi",
				Port:        3150,
				Lifecycle:   "on-demand",
				Enabled:     true,
				Storage:     "/data/devpi",
				IdleTimeout: 900, // 15 minutes
			},
			wantErr:     false,
			checkFields: true,
		},
		{
			name: "create registry without storage should fail (NOT NULL constraint)",
			registry: &models.Registry{
				Name:        "no-storage",
				Type:        "zot",
				Port:        5102,
				Lifecycle:   "manual",
				Enabled:     true,
				Storage:     "", // Empty storage
				IdleTimeout: 0,
			},
			wantErr: true,
			errMsg:  "NOT NULL", // Expected SQLite constraint error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ds.CreateRegistry(tt.registry)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.registry.ID, "ID should be set after creation")

				if tt.checkFields {
					// Retrieve and verify all fields were persisted
					retrieved, err := ds.GetRegistryByID(tt.registry.ID)
					require.NoError(t, err)
					assert.Equal(t, tt.registry.Enabled, retrieved.Enabled, "Enabled field should match")
					assert.Equal(t, tt.registry.Storage, retrieved.Storage, "Storage field should match")
					assert.Equal(t, tt.registry.IdleTimeout, retrieved.IdleTimeout, "IdleTimeout field should match")
				}
			}
		})
	}
}

func TestDataStore_GetRegistryByName_ReturnsAllFields(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registry with all fields
	reg := &models.Registry{
		Name:        "test-get-all-fields",
		Type:        "verdaccio",
		Port:        4900,
		Lifecycle:   "on-demand",
		Enabled:     true,
		Storage:     "/var/lib/verdaccio",
		IdleTimeout: 2400,
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	// Retrieve and verify all fields
	retrieved, err := ds.GetRegistryByName("test-get-all-fields")
	assert.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, reg.Name, retrieved.Name)
	assert.Equal(t, reg.Type, retrieved.Type)
	assert.Equal(t, reg.Port, retrieved.Port)
	assert.Equal(t, reg.Lifecycle, retrieved.Lifecycle)
	assert.Equal(t, reg.Enabled, retrieved.Enabled, "Enabled should be retrieved")
	assert.Equal(t, reg.Storage, retrieved.Storage, "Storage should be retrieved")
	assert.Equal(t, reg.IdleTimeout, retrieved.IdleTimeout, "IdleTimeout should be retrieved")
}

func TestDataStore_GetRegistryByID_ReturnsAllFields(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registry with all fields
	reg := &models.Registry{
		Name:        "test-get-by-id-all-fields",
		Type:        "squid",
		Port:        3200,
		Lifecycle:   "persistent",
		Enabled:     false, // Disabled
		Storage:     "/var/cache/squid",
		IdleTimeout: 0,
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)
	require.NotZero(t, reg.ID)

	// Retrieve and verify all fields
	retrieved, err := ds.GetRegistryByID(reg.ID)
	assert.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, reg.ID, retrieved.ID)
	assert.Equal(t, reg.Name, retrieved.Name)
	assert.Equal(t, reg.Enabled, retrieved.Enabled, "Enabled should be retrieved (false)")
	assert.Equal(t, reg.Storage, retrieved.Storage, "Storage should be retrieved")
	assert.Equal(t, reg.IdleTimeout, retrieved.IdleTimeout, "IdleTimeout should be retrieved")
}

func TestDataStore_GetRegistryByPort_ReturnsAllFields(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registry with all fields
	reg := &models.Registry{
		Name:        "test-get-by-port-all-fields",
		Type:        "zot",
		Port:        5200,
		Lifecycle:   "on-demand",
		Enabled:     true,
		Storage:     "/mnt/zot",
		IdleTimeout: 3600,
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	// Retrieve by port and verify all fields
	retrieved, err := ds.GetRegistryByPort(5200)
	assert.NoError(t, err)
	require.NotNil(t, retrieved)

	assert.Equal(t, reg.Name, retrieved.Name)
	assert.Equal(t, reg.Port, retrieved.Port)
	assert.Equal(t, reg.Enabled, retrieved.Enabled, "Enabled should be retrieved")
	assert.Equal(t, reg.Storage, retrieved.Storage, "Storage should be retrieved")
	assert.Equal(t, reg.IdleTimeout, retrieved.IdleTimeout, "IdleTimeout should be retrieved")
}

func TestDataStore_UpdateRegistry_WithStorageEnabledIdleTimeout(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registry
	reg := &models.Registry{
		Name:        "update-all-fields-test",
		Type:        "athens",
		Port:        3300,
		Lifecycle:   "manual",
		Enabled:     false,
		Storage:     "/var/lib/athens",
		IdleTimeout: 0,
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	// Update all fields including storage, enabled, idle_timeout
	reg.Port = 3301
	reg.Lifecycle = "on-demand"
	reg.Enabled = true
	reg.Storage = "/mnt/custom/athens"
	reg.IdleTimeout = 2700
	reg.Status = "running"

	err = ds.UpdateRegistry(reg)
	assert.NoError(t, err, "UpdateRegistry should succeed")

	// Verify all updates persisted
	updated, err := ds.GetRegistryByID(reg.ID)
	require.NoError(t, err)
	assert.Equal(t, 3301, updated.Port)
	assert.Equal(t, "on-demand", updated.Lifecycle)
	assert.Equal(t, true, updated.Enabled, "Enabled should be updated")
	assert.Equal(t, "/mnt/custom/athens", updated.Storage, "Storage should be updated")
	assert.Equal(t, 2700, updated.IdleTimeout, "IdleTimeout should be updated")
	assert.Equal(t, "running", updated.Status)
}

func TestDataStore_ListRegistries_ReturnsAllFields(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registries with varying field values
	registries := []*models.Registry{
		{
			Name:        "list-all-1",
			Type:        "zot",
			Port:        5300,
			Lifecycle:   "persistent",
			Enabled:     true,
			Storage:     "/var/lib/zot",
			IdleTimeout: 0,
		},
		{
			Name:        "list-all-2",
			Type:        "athens",
			Port:        3400,
			Lifecycle:   "on-demand",
			Enabled:     false,
			Storage:     "/mnt/athens",
			IdleTimeout: 1800,
		},
		{
			Name:        "list-all-3",
			Type:        "devpi",
			Port:        3500,
			Lifecycle:   "manual",
			Enabled:     true,
			Storage:     "/data/devpi",
			IdleTimeout: 3600,
		},
	}

	for _, reg := range registries {
		err := ds.CreateRegistry(reg)
		require.NoError(t, err)
	}

	// List all and verify fields are present
	results, err := ds.ListRegistries()
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 3, "Should retrieve at least 3 registries")

	// Find our test registries and verify fields
	found := make(map[string]*models.Registry)
	for _, r := range results {
		if r.Name == "list-all-1" || r.Name == "list-all-2" || r.Name == "list-all-3" {
			found[r.Name] = r
		}
	}

	// Verify list-all-1
	if r, ok := found["list-all-1"]; ok {
		assert.Equal(t, true, r.Enabled, "list-all-1: Enabled should be retrieved")
		assert.Equal(t, "/var/lib/zot", r.Storage, "list-all-1: Storage should be retrieved")
		assert.Equal(t, 0, r.IdleTimeout, "list-all-1: IdleTimeout should be retrieved")
	} else {
		t.Error("list-all-1 not found in results")
	}

	// Verify list-all-2
	if r, ok := found["list-all-2"]; ok {
		assert.Equal(t, false, r.Enabled, "list-all-2: Enabled should be retrieved")
		assert.Equal(t, "/mnt/athens", r.Storage, "list-all-2: Storage should be retrieved")
		assert.Equal(t, 1800, r.IdleTimeout, "list-all-2: IdleTimeout should be retrieved")
	} else {
		t.Error("list-all-2 not found in results")
	}

	// Verify list-all-3
	if r, ok := found["list-all-3"]; ok {
		assert.Equal(t, true, r.Enabled, "list-all-3: Enabled should be retrieved")
		assert.Equal(t, "/data/devpi", r.Storage, "list-all-3: Storage should be retrieved")
		assert.Equal(t, 3600, r.IdleTimeout, "list-all-3: IdleTimeout should be retrieved")
	} else {
		t.Error("list-all-3 not found in results")
	}
}

func TestDataStore_ListRegistriesByType_ReturnsAllFields(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registries of the same type
	registries := []*models.Registry{
		{
			Name:        "zot-type-all-1",
			Type:        "zot",
			Port:        5400,
			Lifecycle:   "persistent",
			Enabled:     true,
			Storage:     "/var/lib/zot-1",
			IdleTimeout: 0,
		},
		{
			Name:        "zot-type-all-2",
			Type:        "zot",
			Port:        5401,
			Lifecycle:   "on-demand",
			Enabled:     false,
			Storage:     "/var/lib/zot-2",
			IdleTimeout: 2400,
		},
	}

	for _, reg := range registries {
		err := ds.CreateRegistry(reg)
		require.NoError(t, err)
	}

	// List by type and verify fields
	results, err := ds.ListRegistriesByType("zot")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "Should retrieve at least 2 zot registries")

	// Find our test registries
	found := make(map[string]*models.Registry)
	for _, r := range results {
		if r.Name == "zot-type-all-1" || r.Name == "zot-type-all-2" {
			found[r.Name] = r
		}
	}

	// Verify zot-type-all-1
	if r, ok := found["zot-type-all-1"]; ok {
		assert.Equal(t, true, r.Enabled, "zot-type-all-1: Enabled should be retrieved")
		assert.Equal(t, "/var/lib/zot-1", r.Storage, "zot-type-all-1: Storage should be retrieved")
		assert.Equal(t, 0, r.IdleTimeout, "zot-type-all-1: IdleTimeout should be retrieved")
	} else {
		t.Error("zot-type-all-1 not found in results")
	}

	// Verify zot-type-all-2
	if r, ok := found["zot-type-all-2"]; ok {
		assert.Equal(t, false, r.Enabled, "zot-type-all-2: Enabled should be retrieved")
		assert.Equal(t, "/var/lib/zot-2", r.Storage, "zot-type-all-2: Storage should be retrieved")
		assert.Equal(t, 2400, r.IdleTimeout, "zot-type-all-2: IdleTimeout should be retrieved")
	} else {
		t.Error("zot-type-all-2 not found in results")
	}
}

func TestDataStore_ListRegistriesByStatus_ReturnsAllFields(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registries with the same status
	registries := []*models.Registry{
		{
			Name:        "running-all-1",
			Type:        "zot",
			Port:        5500,
			Status:      "running",
			Enabled:     true,
			Storage:     "/var/lib/zot",
			IdleTimeout: 1200,
		},
		{
			Name:        "running-all-2",
			Type:        "athens",
			Port:        3600,
			Status:      "running",
			Enabled:     false,
			Storage:     "/mnt/athens",
			IdleTimeout: 3000,
		},
	}

	for _, reg := range registries {
		err := ds.CreateRegistry(reg)
		require.NoError(t, err)
	}

	// List by status and verify fields
	results, err := ds.ListRegistriesByStatus("running")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(results), 2, "Should retrieve at least 2 running registries")

	// Find our test registries
	found := make(map[string]*models.Registry)
	for _, r := range results {
		if r.Name == "running-all-1" || r.Name == "running-all-2" {
			found[r.Name] = r
		}
	}

	// Verify running-all-1
	if r, ok := found["running-all-1"]; ok {
		assert.Equal(t, true, r.Enabled, "running-all-1: Enabled should be retrieved")
		assert.Equal(t, "/var/lib/zot", r.Storage, "running-all-1: Storage should be retrieved")
		assert.Equal(t, 1200, r.IdleTimeout, "running-all-1: IdleTimeout should be retrieved")
	} else {
		t.Error("running-all-1 not found in results")
	}

	// Verify running-all-2
	if r, ok := found["running-all-2"]; ok {
		assert.Equal(t, false, r.Enabled, "running-all-2: Enabled should be retrieved")
		assert.Equal(t, "/mnt/athens", r.Storage, "running-all-2: Storage should be retrieved")
		assert.Equal(t, 3000, r.IdleTimeout, "running-all-2: IdleTimeout should be retrieved")
	} else {
		t.Error("running-all-2 not found in results")
	}
}
