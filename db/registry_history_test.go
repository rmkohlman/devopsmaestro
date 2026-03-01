package db

import (
	"database/sql"
	"devopsmaestro/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// DataStore Interface Tests for Registry History Operations
// =============================================================================

// TestDataStore_CreateRegistryHistory tests creating registry history entries
func TestDataStore_CreateRegistryHistory(t *testing.T) {
	ds := createTestDataStore(t)

	// Setup: Create a test registry
	reg := &models.Registry{
		Name:      "test-registry-history",
		Type:      "zot",
		Port:      5100,
		Lifecycle: "persistent",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)
	require.NotZero(t, reg.ID)

	tests := []struct {
		name    string
		history *models.RegistryHistory
		wantErr bool
		errMsg  string
	}{
		{
			name: "create valid history entry - start action",
			history: &models.RegistryHistory{
				RegistryID:      reg.ID,
				Revision:        1,
				Config:          `{"type":"zot","port":5100,"lifecycle":"persistent"}`,
				Enabled:         true,
				Lifecycle:       "persistent",
				Port:            5100,
				Storage:         "/data/zot",
				Action:          "start",
				Status:          "success",
				User:            sql.NullString{String: "testuser", Valid: true},
				RegistryVersion: sql.NullString{String: "v2.0.0", Valid: true},
			},
			wantErr: false,
		},
		{
			name: "create with config_change action",
			history: &models.RegistryHistory{
				RegistryID:       reg.ID,
				Revision:         2,
				Config:           `{"type":"zot","port":5101,"lifecycle":"on-demand"}`,
				Enabled:          true,
				Lifecycle:        "on-demand",
				Port:             5101,
				Storage:          "/data/zot",
				Action:           "config_change",
				Status:           "success",
				PreviousRevision: sql.NullInt64{Int64: 1, Valid: true},
			},
			wantErr: false,
		},
		{
			name: "create with rollback action",
			history: &models.RegistryHistory{
				RegistryID:       reg.ID,
				Revision:         3,
				Config:           `{"type":"zot","port":5100,"lifecycle":"persistent"}`,
				Enabled:          true,
				Lifecycle:        "persistent",
				Port:             5100,
				Storage:          "/data/zot",
				Action:           "rollback",
				Status:           "success",
				PreviousRevision: sql.NullInt64{Int64: 2, Valid: true},
			},
			wantErr: false,
		},
		{
			name: "create with idle_timeout",
			history: &models.RegistryHistory{
				RegistryID:  reg.ID,
				Revision:    4,
				Config:      `{"type":"zot","port":5100,"lifecycle":"on-demand","idle_timeout":300}`,
				Enabled:     true,
				Lifecycle:   "on-demand",
				Port:        5100,
				Storage:     "/data/zot",
				IdleTimeout: sql.NullInt64{Int64: 300, Valid: true},
				Action:      "config_change",
				Status:      "success",
			},
			wantErr: false,
		},
		{
			name: "create with failed status",
			history: &models.RegistryHistory{
				RegistryID:   reg.ID,
				Revision:     5,
				Config:       `{"type":"zot","port":5100}`,
				Enabled:      true,
				Lifecycle:    "persistent",
				Port:         5100,
				Storage:      "/data/zot",
				Action:       "restart",
				Status:       "failed",
				ErrorMessage: sql.NullString{String: "container failed to start", Valid: true},
			},
			wantErr: false,
		},
		{
			name: "create with minimal fields",
			history: &models.RegistryHistory{
				RegistryID: reg.ID,
				Revision:   6,
				Config:     `{}`,
				Enabled:    false,
				Lifecycle:  "manual",
				Port:       0,
				Action:     "config_change",
				Status:     "in_progress",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ds.CreateRegistryHistory(tt.history)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.history.ID, "ID should be set after creation")
				assert.NotZero(t, tt.history.CreatedAt, "CreatedAt should be set")

				// Verify all fields persisted correctly
				retrieved, err := ds.GetRegistryHistory(tt.history.RegistryID, tt.history.Revision)
				require.NoError(t, err)
				assert.Equal(t, tt.history.ID, retrieved.ID)
				assert.Equal(t, tt.history.RegistryID, retrieved.RegistryID)
				assert.Equal(t, tt.history.Revision, retrieved.Revision)
				assert.Equal(t, tt.history.Config, retrieved.Config)
				assert.Equal(t, tt.history.Enabled, retrieved.Enabled)
				assert.Equal(t, tt.history.Lifecycle, retrieved.Lifecycle)
				assert.Equal(t, tt.history.Port, retrieved.Port)
				assert.Equal(t, tt.history.Storage, retrieved.Storage)
				assert.Equal(t, tt.history.Action, retrieved.Action)
				assert.Equal(t, tt.history.Status, retrieved.Status)
			}
		})
	}
}

// TestDataStore_GetRegistryHistory tests retrieving history by registryID and revision
func TestDataStore_GetRegistryHistory(t *testing.T) {
	ds := createTestDataStore(t)

	// Setup: Create test registry
	reg := &models.Registry{
		Name:      "get-history-test",
		Type:      "zot",
		Port:      5200,
		Lifecycle: "persistent",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	// Create test history entries
	history1 := &models.RegistryHistory{
		RegistryID: reg.ID,
		Revision:   1,
		Config:     `{"type":"zot","port":5200}`,
		Enabled:    true,
		Lifecycle:  "persistent",
		Port:       5200,
		Action:     "start",
		Status:     "success",
	}
	err = ds.CreateRegistryHistory(history1)
	require.NoError(t, err)

	history2 := &models.RegistryHistory{
		RegistryID:       reg.ID,
		Revision:         2,
		Config:           `{"type":"zot","port":5201}`,
		Enabled:          true,
		Lifecycle:        "on-demand",
		Port:             5201,
		Action:           "config_change",
		Status:           "success",
		PreviousRevision: sql.NullInt64{Int64: 1, Valid: true},
	}
	err = ds.CreateRegistryHistory(history2)
	require.NoError(t, err)

	tests := []struct {
		name       string
		registryID int
		revision   int
		wantErr    bool
		wantPort   int
	}{
		{
			name:       "get revision 1",
			registryID: reg.ID,
			revision:   1,
			wantErr:    false,
			wantPort:   5200,
		},
		{
			name:       "get revision 2",
			registryID: reg.ID,
			revision:   2,
			wantErr:    false,
			wantPort:   5201,
		},
		{
			name:       "get non-existent revision",
			registryID: reg.ID,
			revision:   999,
			wantErr:    true,
		},
		{
			name:       "get non-existent registry",
			registryID: 99999,
			revision:   1,
			wantErr:    true,
		},
		{
			name:       "get with invalid registryID (negative)",
			registryID: -1,
			revision:   1,
			wantErr:    true,
		},
		{
			name:       "get with invalid revision (zero)",
			registryID: reg.ID,
			revision:   0,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ds.GetRegistryHistory(tt.registryID, tt.revision)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.registryID, result.RegistryID)
				assert.Equal(t, tt.revision, result.Revision)
				assert.Equal(t, tt.wantPort, result.Port)
			}
		})
	}
}

// TestDataStore_GetLatestRegistryHistory tests retrieving the latest history entry
func TestDataStore_GetLatestRegistryHistory(t *testing.T) {
	ds := createTestDataStore(t)

	// Setup: Create test registry
	reg := &models.Registry{
		Name:      "latest-history-test",
		Type:      "athens",
		Port:      3200,
		Lifecycle: "on-demand",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	tests := []struct {
		name           string
		setupHistories []int // revisions to create before test
		registryID     int
		wantErr        bool
		wantRevision   int
	}{
		{
			name:         "no history - should error",
			registryID:   reg.ID,
			wantErr:      true,
			wantRevision: 0,
		},
		{
			name:           "single history entry",
			setupHistories: []int{1},
			registryID:     reg.ID,
			wantErr:        false,
			wantRevision:   1,
		},
		{
			name:           "multiple history entries - returns highest",
			setupHistories: []int{2, 3, 4, 5},
			registryID:     reg.ID,
			wantErr:        false,
			wantRevision:   5,
		},
		{
			name:       "non-existent registry",
			registryID: 99999,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create history entries for this test
			for _, rev := range tt.setupHistories {
				history := &models.RegistryHistory{
					RegistryID: reg.ID,
					Revision:   rev,
					Config:     `{"test":true}`,
					Enabled:    true,
					Lifecycle:  "on-demand",
					Port:       3200 + rev,
					Action:     "config_change",
					Status:     "success",
				}
				err := ds.CreateRegistryHistory(history)
				require.NoError(t, err)
			}

			// Test
			result, err := ds.GetLatestRegistryHistory(tt.registryID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.registryID, result.RegistryID)
				assert.Equal(t, tt.wantRevision, result.Revision)
			}
		})
	}
}

// TestDataStore_ListRegistryHistory tests listing all history for a registry
func TestDataStore_ListRegistryHistory(t *testing.T) {
	ds := createTestDataStore(t)

	// Setup: Create two test registries
	reg1 := &models.Registry{
		Name:      "list-history-reg1",
		Type:      "zot",
		Port:      5300,
		Lifecycle: "persistent",
	}
	err := ds.CreateRegistry(reg1)
	require.NoError(t, err)

	reg2 := &models.Registry{
		Name:      "list-history-reg2",
		Type:      "athens",
		Port:      3300,
		Lifecycle: "on-demand",
	}
	err = ds.CreateRegistry(reg2)
	require.NoError(t, err)

	// Create history for reg1 (revisions 1, 2, 3)
	for i := 1; i <= 3; i++ {
		history := &models.RegistryHistory{
			RegistryID: reg1.ID,
			Revision:   i,
			Config:     `{"test":true}`,
			Enabled:    true,
			Lifecycle:  "persistent",
			Port:       5300 + i,
			Action:     "config_change",
			Status:     "success",
		}
		err := ds.CreateRegistryHistory(history)
		require.NoError(t, err)
	}

	// Create history for reg2 (revisions 1, 2)
	for i := 1; i <= 2; i++ {
		history := &models.RegistryHistory{
			RegistryID: reg2.ID,
			Revision:   i,
			Config:     `{"test":true}`,
			Enabled:    true,
			Lifecycle:  "on-demand",
			Port:       3300 + i,
			Action:     "start",
			Status:     "success",
		}
		err := ds.CreateRegistryHistory(history)
		require.NoError(t, err)
	}

	tests := []struct {
		name          string
		registryID    int
		wantErr       bool
		wantCount     int
		wantRevisions []int // expected revisions in DESC order
	}{
		{
			name:          "list all history for reg1",
			registryID:    reg1.ID,
			wantErr:       false,
			wantCount:     3,
			wantRevisions: []int{3, 2, 1}, // DESC order
		},
		{
			name:          "list all history for reg2",
			registryID:    reg2.ID,
			wantErr:       false,
			wantCount:     2,
			wantRevisions: []int{2, 1}, // DESC order
		},
		{
			name:       "list history for non-existent registry",
			registryID: 99999,
			wantErr:    false, // Should return empty list, not error
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := ds.ListRegistryHistory(tt.registryID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, results, tt.wantCount)

				// Verify order (DESC by revision)
				if tt.wantCount > 0 {
					for i, expectedRev := range tt.wantRevisions {
						assert.Equal(t, expectedRev, results[i].Revision,
							"History should be ordered DESC by revision")
					}

					// Verify all entries belong to correct registry
					for _, h := range results {
						assert.Equal(t, tt.registryID, h.RegistryID,
							"All history entries should belong to the requested registry")
					}
				}
			}
		})
	}
}

// TestDataStore_GetNextRevisionNumber tests getting the next revision number
func TestDataStore_GetNextRevisionNumber(t *testing.T) {
	ds := createTestDataStore(t)

	// Setup: Create test registries
	reg1 := &models.Registry{
		Name:      "revision-test-reg1",
		Type:      "zot",
		Port:      5400,
		Lifecycle: "persistent",
	}
	err := ds.CreateRegistry(reg1)
	require.NoError(t, err)

	reg2 := &models.Registry{
		Name:      "revision-test-reg2",
		Type:      "athens",
		Port:      3400,
		Lifecycle: "on-demand",
	}
	err = ds.CreateRegistry(reg2)
	require.NoError(t, err)

	tests := []struct {
		name           string
		registryID     int
		setupRevisions []int // revisions to create before test
		wantRevision   int
		wantErr        bool
	}{
		{
			name:         "no existing history - should return 1",
			registryID:   reg1.ID,
			wantRevision: 1,
			wantErr:      false,
		},
		{
			name:           "existing history - should return max+1",
			registryID:     reg1.ID,
			setupRevisions: []int{1},
			wantRevision:   2,
			wantErr:        false,
		},
		{
			name:           "multiple revisions - should return max+1",
			registryID:     reg1.ID,
			setupRevisions: []int{2, 3, 4},
			wantRevision:   5,
			wantErr:        false,
		},
		{
			name:         "different registry - independent revision counter",
			registryID:   reg2.ID,
			wantRevision: 1, // reg2 has no history yet
			wantErr:      false,
		},
		{
			name:         "non-existent registry - should still return 1",
			registryID:   99999,
			wantRevision: 1,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup: Create history entries for this test
			for _, rev := range tt.setupRevisions {
				history := &models.RegistryHistory{
					RegistryID: tt.registryID,
					Revision:   rev,
					Config:     `{"test":true}`,
					Enabled:    true,
					Lifecycle:  "persistent",
					Port:       5000,
					Action:     "config_change",
					Status:     "success",
				}
				err := ds.CreateRegistryHistory(history)
				require.NoError(t, err)
			}

			// Test
			nextRev, err := ds.GetNextRevisionNumber(tt.registryID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRevision, nextRev)
			}
		})
	}
}

// TestDataStore_RegistryHistory_CompletedAt tests handling of CompletedAt timestamp
func TestDataStore_RegistryHistory_CompletedAt(t *testing.T) {
	ds := createTestDataStore(t)

	// Setup: Create test registry
	reg := &models.Registry{
		Name:      "completed-at-test",
		Type:      "zot",
		Port:      5500,
		Lifecycle: "persistent",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	tests := []struct {
		name           string
		history        *models.RegistryHistory
		wantCompleted  bool
		checkCompleted bool
	}{
		{
			name: "in_progress status - no CompletedAt",
			history: &models.RegistryHistory{
				RegistryID: reg.ID,
				Revision:   1,
				Config:     `{}`,
				Enabled:    true,
				Lifecycle:  "persistent",
				Port:       5500,
				Action:     "start",
				Status:     "in_progress",
			},
			wantCompleted: false,
		},
		{
			name: "success status - has CompletedAt",
			history: &models.RegistryHistory{
				RegistryID:  reg.ID,
				Revision:    2,
				Config:      `{}`,
				Enabled:     true,
				Lifecycle:   "persistent",
				Port:        5500,
				Action:      "start",
				Status:      "success",
				CompletedAt: sql.NullTime{Time: time.Now(), Valid: true},
			},
			wantCompleted: true,
		},
		{
			name: "failed status - has CompletedAt",
			history: &models.RegistryHistory{
				RegistryID:   reg.ID,
				Revision:     3,
				Config:       `{}`,
				Enabled:      true,
				Lifecycle:    "persistent",
				Port:         5500,
				Action:       "restart",
				Status:       "failed",
				ErrorMessage: sql.NullString{String: "failed to start", Valid: true},
				CompletedAt:  sql.NullTime{Time: time.Now(), Valid: true},
			},
			wantCompleted: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ds.CreateRegistryHistory(tt.history)
			require.NoError(t, err)

			// Retrieve and verify CompletedAt
			retrieved, err := ds.GetRegistryHistory(tt.history.RegistryID, tt.history.Revision)
			require.NoError(t, err)

			if tt.wantCompleted {
				assert.True(t, retrieved.CompletedAt.Valid, "CompletedAt should be set")
				assert.NotZero(t, retrieved.CompletedAt.Time, "CompletedAt time should be non-zero")
			} else {
				assert.False(t, retrieved.CompletedAt.Valid, "CompletedAt should not be set")
			}
		})
	}
}

// TestDataStore_RegistryHistory_EdgeCases tests edge cases and boundary conditions
func TestDataStore_RegistryHistory_EdgeCases(t *testing.T) {
	ds := createTestDataStore(t)

	// Setup: Create test registry
	reg := &models.Registry{
		Name:      "edge-case-test",
		Type:      "zot",
		Port:      5600,
		Lifecycle: "persistent",
	}
	err := ds.CreateRegistry(reg)
	require.NoError(t, err)

	tests := []struct {
		name    string
		history *models.RegistryHistory
		wantErr bool
		errMsg  string
	}{
		{
			name: "empty config string",
			history: &models.RegistryHistory{
				RegistryID: reg.ID,
				Revision:   1,
				Config:     "",
				Enabled:    true,
				Lifecycle:  "manual",
				Port:       0,
				Action:     "config_change",
				Status:     "success",
			},
			wantErr: false,
		},
		{
			name: "very long config JSON",
			history: &models.RegistryHistory{
				RegistryID: reg.ID,
				Revision:   2,
				Config:     `{"key":"` + string(make([]byte, 10000)) + `"}`,
				Enabled:    true,
				Lifecycle:  "persistent",
				Port:       5600,
				Action:     "config_change",
				Status:     "success",
			},
			wantErr: false,
		},
		{
			name: "port zero (disabled registry)",
			history: &models.RegistryHistory{
				RegistryID: reg.ID,
				Revision:   3,
				Config:     `{"enabled":false}`,
				Enabled:    false,
				Lifecycle:  "manual",
				Port:       0,
				Action:     "config_change",
				Status:     "success",
			},
			wantErr: false,
		},
		{
			name: "empty storage path",
			history: &models.RegistryHistory{
				RegistryID: reg.ID,
				Revision:   4,
				Config:     `{}`,
				Enabled:    true,
				Lifecycle:  "persistent",
				Port:       5600,
				Storage:    "",
				Action:     "start",
				Status:     "success",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ds.CreateRegistryHistory(tt.history)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotZero(t, tt.history.ID)
			}
		})
	}
}
