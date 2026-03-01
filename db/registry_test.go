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
			},
			wantErr: true,
			errMsg:  "already exists",
		},
		{
			name: "create with invalid type",
			registry: &models.Registry{
				Name:      "bad-registry",
				Type:      "invalid",
				Port:      5000,
				Lifecycle: "manual",
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
		{Name: "zot-list-1", Type: "zot", Port: 5020, Lifecycle: "persistent"},
		{Name: "athens-list-1", Type: "athens", Port: 3020, Lifecycle: "on-demand"},
		{Name: "devpi-list-1", Type: "devpi", Port: 3141, Lifecycle: "manual"},
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
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	// Delete by name
	err = ds.DeleteRegistry(reg.Name)
	assert.NoError(t, err)

	// Verify deletion
	_, err = ds.GetRegistryByName(reg.Name)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDataStore_GetRegistryByPort(t *testing.T) {
	ds := createTestDataStore(t)

	// Create test registries on different ports
	reg1 := &models.Registry{
		Name:      "reg-5050",
		Type:      "zot",
		Port:      5050,
		Lifecycle: "persistent",
	}
	err := ds.CreateRegistry(reg1)
	require.NoError(t, err)

	reg2 := &models.Registry{
		Name:      "reg-5051",
		Type:      "zot",
		Port:      5051,
		Lifecycle: "persistent",
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
	}
	err := ds.CreateRegistry(reg1)
	require.NoError(t, err)

	// Try to create another registry on the same port
	reg2 := &models.Registry{
		Name:      "reg-conflict-2",
		Type:      "athens",
		Port:      5060, // Same port - should conflict
		Lifecycle: "on-demand",
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
		{Name: "zot-type-1", Type: "zot", Port: 5080, Lifecycle: "persistent"},
		{Name: "zot-type-2", Type: "zot", Port: 5081, Lifecycle: "manual"},
		{Name: "athens-type-1", Type: "athens", Port: 3080, Lifecycle: "on-demand"},
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
		{Name: "running-status-1", Type: "zot", Port: 5090, Status: "running"},
		{Name: "running-status-2", Type: "athens", Port: 3090, Status: "running"},
		{Name: "stopped-status-1", Type: "devpi", Port: 3091, Status: "stopped"},
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
