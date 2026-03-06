package cmd

import (
	"context"
	"errors"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/registry"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TDD Phase 2 (RED): Bug 4 — Auto-stop running registry before delete
// =============================================================================
// The current deleteRegistry() in cmd/delete.go has NO stop logic.
// It calls ds.GetRegistryByName() then immediately ds.DeleteRegistry()
// without checking if the process is running.
//
// These tests FAIL until:
//   1. A deleteRegistryCore(ctx, ds, factory, name, force) function is extracted
//      (or equivalent injectable form) from deleteRegistry() in cmd/delete.go
//   2. deleteRegistryCore checks IsRunning() before deleting
//   3. If running, Stop() is called before ds.DeleteRegistry()
//   4. If Stop() fails, the DB record is NOT deleted and an error is returned
//
// Design:
//   - Option C (always auto-stop): Stop() is called whenever IsRunning() == true
//   - --force only skips the confirmation prompt (not the stop logic)
//   - On Stop() error: abort entirely, do NOT delete DB record
// =============================================================================

// mockRegistryManager is a test double for registry.ServiceManager.
type mockRegistryManager struct {
	isRunning   bool
	stopErr     error
	stopCalled  bool
	startCalled bool
}

func (m *mockRegistryManager) Start(ctx context.Context) error {
	m.startCalled = true
	return nil
}

func (m *mockRegistryManager) Stop(ctx context.Context) error {
	m.stopCalled = true
	return m.stopErr
}

func (m *mockRegistryManager) IsRunning(ctx context.Context) bool {
	return m.isRunning
}

func (m *mockRegistryManager) GetEndpoint() string {
	return "localhost:5001"
}

// Compile-time check: mockRegistryManager implements registry.ServiceManager
var _ registry.ServiceManager = (*mockRegistryManager)(nil)

// mockManagerFactory is a test double for registry.ManagerFactory.
type mockManagerFactory struct {
	manager *mockRegistryManager
	err     error
}

func (f *mockManagerFactory) CreateManager(reg *models.Registry) (registry.ServiceManager, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.manager, nil
}

// Compile-time check: mockManagerFactory implements registry.ManagerFactory
var _ registry.ManagerFactory = (*mockManagerFactory)(nil)

// newTestDeleteDS creates a MockDataStore pre-loaded with a named registry.
func newTestDeleteDS(t *testing.T, name string) *db.MockDataStore {
	t.Helper()
	ds := db.NewMockDataStore()
	ds.Registries[name] = &models.Registry{
		ID:      1,
		Name:    name,
		Type:    "zot",
		Port:    5001,
		Storage: "/var/lib/zot",
		Status:  "running",
	}
	return ds
}

// =============================================================================
// TestDeleteRegistry_StopsRunningProcess
// =============================================================================
// When the registry is running (IsRunning == true), Stop() MUST be called
// before the DB record is deleted.
// =============================================================================

func TestDeleteRegistry_StopsRunningProcess(t *testing.T) {
	mgr := &mockRegistryManager{isRunning: true}
	factory := &mockManagerFactory{manager: mgr}
	ds := newTestDeleteDS(t, "my-zot")

	ctx := context.Background()

	// Call the refactored core function (force=true skips confirmation prompt)
	err := deleteRegistryCore(ctx, ds, factory, "my-zot", true)
	require.NoError(t, err, "deleteRegistryCore should succeed when stop succeeds")

	// Stop() must have been called before delete
	assert.True(t, mgr.stopCalled, "Stop() should be called when registry is running (Bug 4: no stop logic)")

	// After stop, the DB record should be deleted
	_, exists := ds.Registries["my-zot"]
	assert.False(t, exists, "registry should be removed from DB after successful delete")
}

// =============================================================================
// TestDeleteRegistry_SkipsStopWhenStopped
// =============================================================================
// When the registry is NOT running (IsRunning == false), Stop() should NOT
// be called — delete proceeds directly.
// =============================================================================

func TestDeleteRegistry_SkipsStopWhenStopped(t *testing.T) {
	mgr := &mockRegistryManager{isRunning: false} // already stopped
	factory := &mockManagerFactory{manager: mgr}
	ds := newTestDeleteDS(t, "my-zot")

	ctx := context.Background()

	err := deleteRegistryCore(ctx, ds, factory, "my-zot", true)
	require.NoError(t, err, "deleteRegistryCore should succeed when registry is stopped")

	// Stop() must NOT have been called (registry was already stopped)
	assert.False(t, mgr.stopCalled, "Stop() should NOT be called when registry is not running")

	// DB record should still be deleted
	_, exists := ds.Registries["my-zot"]
	assert.False(t, exists, "registry should be removed from DB")
}

// =============================================================================
// TestDeleteRegistry_AbortsIfStopFails
// =============================================================================
// When Stop() returns an error, the DB record must NOT be deleted and
// deleteRegistryCore must return an error.
// =============================================================================

func TestDeleteRegistry_AbortsIfStopFails(t *testing.T) {
	stopErr := errors.New("process kill failed")
	mgr := &mockRegistryManager{
		isRunning: true,
		stopErr:   stopErr,
	}
	factory := &mockManagerFactory{manager: mgr}
	ds := newTestDeleteDS(t, "my-zot")

	ctx := context.Background()

	err := deleteRegistryCore(ctx, ds, factory, "my-zot", true)

	// Must return an error
	require.Error(t, err, "deleteRegistryCore should return error when Stop() fails (Bug 4: no stop logic)")
	assert.Contains(t, err.Error(), "stop", "error message should mention stop failure")

	// DB record must NOT be deleted — rollback protection
	_, exists := ds.Registries["my-zot"]
	assert.True(t, exists, "registry DB record should NOT be deleted when Stop() fails")

	// Stop() was called but failed
	assert.True(t, mgr.stopCalled, "Stop() should have been attempted")
}
