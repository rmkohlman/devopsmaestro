package registry

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// skipIfShort skips integration tests that require Zot binary in short mode.
func skipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("Skipping integration test that requires Zot binary")
	}
}

// setupTestRegistryManager creates a ZotManager with test configuration.
func setupTestRegistryManager(t *testing.T) RegistryManager {
	t.Helper()

	config := RegistryConfig{
		Enabled:     true,
		Lifecycle:   "manual",
		Port:        15001, // Use high port for tests
		Storage:     t.TempDir(),
		IdleTimeout: 1 * time.Minute,
		Mirrors: []MirrorConfig{
			{
				Name:     "docker-hub",
				URL:      "https://index.docker.io",
				OnDemand: true,
				Prefix:   "docker.io",
			},
		},
	}

	// Use mock binary manager for unit tests
	mockBinary := NewMockBinaryManager(config.Storage, "1.4.3")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/zot.pid",
		LogFile: config.Storage + "/zot.log",
	})

	return NewZotManagerWithDeps(config, mockBinary, mockProcess)
}

// =============================================================================
// Task 2.1: Registry Start Tests
// =============================================================================

func TestZotManager_Start_Success(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err, "Start should succeed with valid config")

	// Verify registry is running
	assert.True(t, mgr.IsRunning(ctx), "IsRunning should return true after Start")

	// Verify endpoint is available
	endpoint := mgr.GetEndpoint()
	assert.NotEmpty(t, endpoint, "GetEndpoint should return non-empty string")
	assert.Contains(t, endpoint, "localhost", "Endpoint should contain localhost")

	// Cleanup
	defer mgr.Stop(ctx)
}

func TestZotManager_Start_AlreadyRunning(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	// Start first time
	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Start second time - should handle gracefully
	err = mgr.Start(ctx)
	assert.NoError(t, err, "Starting already-running registry should be idempotent")
}

func TestZotManager_Start_BinaryNotFound(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	// Create manager with invalid binary path
	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      15002,
		Storage:   t.TempDir(),
	}

	mgr := NewZotManager(config)
	ctx := context.Background()

	// Mock BinaryManager to return error
	// (This test will need proper mock injection in implementation)

	err := mgr.Start(ctx)
	assert.Error(t, err, "Start should fail if binary not found")
	assert.Contains(t, err.Error(), "binary", "Error should mention binary")
}

func TestZotManager_Start_PortInUse(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	// Start first registry
	mgr1 := setupTestRegistryManager(t)
	ctx := context.Background()

	err := mgr1.Start(ctx)
	require.NoError(t, err)
	defer mgr1.Stop(ctx)

	// Try to start second registry on same port
	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      15001, // Same port
		Storage:   t.TempDir(),
	}
	mgr2 := NewZotManager(config)

	err = mgr2.Start(ctx)
	assert.Error(t, err, "Start should fail if port is in use")
	assert.Contains(t, err.Error(), "port", "Error should mention port conflict")
}

// =============================================================================
// Task 2.2: Registry Stop Tests
// =============================================================================

func TestZotManager_Stop_Success(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	// Start registry
	err := mgr.Start(ctx)
	require.NoError(t, err)

	// Stop registry
	err = mgr.Stop(ctx)
	require.NoError(t, err, "Stop should succeed")

	// Verify registry is stopped
	assert.False(t, mgr.IsRunning(ctx), "IsRunning should return false after Stop")
}

func TestZotManager_Stop_NotRunning(t *testing.T) {
	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	// Stop when not running - should be idempotent
	err := mgr.Stop(ctx)
	assert.NoError(t, err, "Stopping non-running registry should be idempotent")
}

func TestZotManager_Stop_GracefulShutdown(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)

	// Get PID before stopping
	status, err := mgr.Status(ctx)
	require.NoError(t, err)
	pid := status.PID

	// Stop registry with timeout
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = mgr.Stop(stopCtx)
	require.NoError(t, err, "Stop should complete within timeout")

	// Verify process is actually terminated
	assert.Greater(t, pid, 0, "PID should have been set when running")
	assert.False(t, mgr.IsRunning(ctx), "Process should be stopped")
}

// =============================================================================
// Task 2.3: Registry Status Tests
// =============================================================================

func TestZotManager_Status_Running(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	status, err := mgr.Status(ctx)
	require.NoError(t, err, "Status should succeed when running")

	assert.Equal(t, "running", status.State, "State should be 'running'")
	assert.Greater(t, status.PID, 0, "PID should be set")
	assert.Equal(t, 15001, status.Port, "Port should match config")
	assert.NotEmpty(t, status.Storage, "Storage path should be set")
	assert.NotEmpty(t, status.Version, "Version should be set")
	assert.GreaterOrEqual(t, status.Uptime, time.Duration(0), "Uptime should be non-negative")
}

func TestZotManager_Status_Stopped(t *testing.T) {
	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	status, err := mgr.Status(ctx)
	require.NoError(t, err, "Status should succeed even when stopped")

	assert.Equal(t, "stopped", status.State, "State should be 'stopped'")
	assert.Equal(t, 0, status.PID, "PID should be 0 when stopped")
	assert.Equal(t, 15001, status.Port, "Port should still be in config")
	assert.NotEmpty(t, status.Storage, "Storage path should be in config")
}

func TestZotManager_Status_ImageCount(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Push a test image (mocked or via real registry)
	// This would require integration with container runtime

	status, err := mgr.Status(ctx)
	require.NoError(t, err)

	// Initially should be 0
	assert.GreaterOrEqual(t, status.ImageCount, 0, "ImageCount should be non-negative")
	assert.GreaterOrEqual(t, status.DiskUsage, int64(0), "DiskUsage should be non-negative")
}

// =============================================================================
// Task 2.4: EnsureRunning Tests
// =============================================================================

func TestZotManager_EnsureRunning_StartsIfStopped(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()
	defer mgr.Stop(ctx)

	// Verify not running initially
	assert.False(t, mgr.IsRunning(ctx), "Should not be running initially")

	// EnsureRunning should start the registry
	err := mgr.EnsureRunning(ctx)
	require.NoError(t, err, "EnsureRunning should succeed")

	// Verify now running
	assert.True(t, mgr.IsRunning(ctx), "Should be running after EnsureRunning")
}

func TestZotManager_EnsureRunning_NoopIfRunning(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	// Start registry
	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Get initial status
	status1, err := mgr.Status(ctx)
	require.NoError(t, err)

	// EnsureRunning should be idempotent
	err = mgr.EnsureRunning(ctx)
	require.NoError(t, err)

	// PID should be the same
	status2, err := mgr.Status(ctx)
	require.NoError(t, err)
	assert.Equal(t, status1.PID, status2.PID, "PID should not change")
}

// =============================================================================
// Task 2.5: IsRunning Tests
// =============================================================================

func TestZotManager_IsRunning_True(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	assert.True(t, mgr.IsRunning(ctx), "IsRunning should return true when registry is running")
}

func TestZotManager_IsRunning_False(t *testing.T) {
	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	assert.False(t, mgr.IsRunning(ctx), "IsRunning should return false when registry is not running")
}

func TestZotManager_IsRunning_AfterCrash(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)

	// Simulate crash by killing process directly
	_, err = mgr.Status(ctx)
	require.NoError(t, err)

	// (Kill process with signal - implementation-specific)

	// IsRunning should detect crashed process
	assert.False(t, mgr.IsRunning(ctx), "IsRunning should return false after crash")
}

// =============================================================================
// Task 2.6: GetEndpoint Tests
// =============================================================================

func TestZotManager_GetEndpoint_ReturnsCorrectFormat(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantFmt string
	}{
		{
			name:    "default port 5001",
			port:    5001,
			wantFmt: "http://localhost:5001",
		},
		{
			name:    "custom port 5555",
			port:    5555,
			wantFmt: "http://localhost:5555",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := RegistryConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      tt.port,
				Storage:   t.TempDir(),
			}
			mgr := NewZotManager(config)

			endpoint := mgr.GetEndpoint()
			assert.Equal(t, tt.wantFmt, endpoint, "Endpoint format should be localhost:port")
		})
	}
}

// =============================================================================
// Task 2.7: Prune Tests
// =============================================================================

func TestZotManager_Prune_All(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Add some test images (mocked or real)
	// This would require integration testing

	opts := PruneOptions{
		All:    true,
		DryRun: false,
	}

	result, err := mgr.Prune(ctx, opts)
	require.NoError(t, err, "Prune should succeed")

	assert.GreaterOrEqual(t, result.ImagesRemoved, 0, "ImagesRemoved should be non-negative")
	assert.GreaterOrEqual(t, result.SpaceReclaimed, int64(0), "SpaceReclaimed should be non-negative")
}

func TestZotManager_Prune_OlderThan(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	opts := PruneOptions{
		All:       false,
		OlderThan: 24 * time.Hour,
		DryRun:    false,
	}

	result, err := mgr.Prune(ctx, opts)
	require.NoError(t, err, "Prune with OlderThan should succeed")

	assert.GreaterOrEqual(t, result.ImagesRemoved, 0, "ImagesRemoved should be non-negative")
	assert.GreaterOrEqual(t, result.SpaceReclaimed, int64(0), "SpaceReclaimed should be non-negative")
}

func TestZotManager_Prune_DryRun(t *testing.T) {
	t.Skip("Integration test - requires real Zot binary with HTTP server")

	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	opts := PruneOptions{
		All:    true,
		DryRun: true,
	}

	result, err := mgr.Prune(ctx, opts)
	require.NoError(t, err, "Prune dry-run should succeed")

	// Dry run should not remove anything
	assert.Equal(t, 0, result.ImagesRemoved, "DryRun should not remove images")
	assert.Equal(t, int64(0), result.SpaceReclaimed, "DryRun should not reclaim space")

	// But should report what WOULD be removed
	assert.NotNil(t, result.Images, "DryRun should list images that would be removed")
}

func TestZotManager_Prune_NotRunning(t *testing.T) {
	mgr := setupTestRegistryManager(t)
	ctx := context.Background()

	opts := PruneOptions{
		All:    true,
		DryRun: false,
	}

	_, err := mgr.Prune(ctx, opts)
	assert.Error(t, err, "Prune should fail if registry is not running")
	assert.Contains(t, err.Error(), "not running", "Error should indicate registry is not running")
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestZotManager_ImplementsRegistryManager(t *testing.T) {
	var _ RegistryManager = (*ZotManager)(nil)
}

// =============================================================================
// Bug 1: Health check probe when port is in use (TDD RED phase)
//
// Current behaviour: IsPortAvailable(port) == false → immediately return ErrPortInUse.
//
// Desired behaviour: Before returning ErrPortInUse, probe the service's health
// endpoint (http://localhost:{port}/v2/ → 200 or 401). If healthy, adopt the
// running instance and return nil (idempotent). Only return ErrPortInUse when
// the health probe fails.
// =============================================================================

// TestZotManager_Start_AdoptsRunningInstance verifies that Start() returns nil
// when the configured port is occupied by a healthy Zot instance (i.e.
// /v2/ returns 200). This is the "adopt already-running service" path.
//
// This test FAILS today because Start() returns ErrPortInUse without probing
// the health endpoint first.
// bindAndServeZot binds a TCP listener on 0.0.0.0:0 (so that
// IsPortAvailable returns false for the chosen port) and begins serving HTTP
// requests using the provided handler. The bound port is returned.
// The server is shut down when the test ends via t.Cleanup.
func bindAndServeZot(t *testing.T, handler http.Handler) int {
	t.Helper()
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	require.NoError(t, err, "bindAndServeZot: net.Listen failed")
	port := ln.Addr().(*net.TCPAddr).Port
	srv := &http.Server{Handler: handler}
	go srv.Serve(ln) //nolint:errcheck
	t.Cleanup(func() { srv.Close() })
	return port
}

// TestZotManager_Start_AdoptsRunningInstance verifies that Start() returns nil
// when the configured port is occupied by a healthy Zot instance (i.e.
// /v2/ returns 200). This is the "adopt already-running service" path.
//
// This test FAILS today because Start() returns ErrPortInUse without probing
// the health endpoint first.
func TestZotManager_Start_AdoptsRunningInstance(t *testing.T) {
	// Stand up a fake "Zot" that responds 200 on /v2/.
	// We bind on 0.0.0.0 so that IsPortAvailable() sees the port as taken.
	port := bindAndServeZot(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))

	// Build a ZotManager pointing at that port.
	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      port,
		Storage:   t.TempDir(),
	}
	mockBinary := NewMockBinaryManager(config.Storage, "1.4.3")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/zot.pid",
		LogFile: config.Storage + "/zot.log",
	})
	mgr := NewZotManagerWithDeps(config, mockBinary, mockProcess)

	ctx := context.Background()

	// BUG: currently returns ErrPortInUse.
	// FIXED: should probe /v2/, see 200, and return nil.
	startErr := mgr.Start(ctx)
	assert.NoError(t, startErr,
		"Start should return nil when the port is occupied by a healthy Zot instance "+
			"(adopt-running-instance path)")
}

// TestZotManager_Start_AdoptsRunningInstance_Unauthorized verifies that Start()
// also adopts a running instance when /v2/ returns 401 (auth-enabled Zot).
//
// This test FAILS today because Start() returns ErrPortInUse without probing.
func TestZotManager_Start_AdoptsRunningInstance_Unauthorized(t *testing.T) {
	// Stand up a fake "Zot" that responds 401 on /v2/ (auth-enabled registry).
	// We bind on 0.0.0.0 so that IsPortAvailable() sees the port as taken.
	port := bindAndServeZot(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v2/" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.NotFound(w, r)
	}))

	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      port,
		Storage:   t.TempDir(),
	}
	mockBinary := NewMockBinaryManager(config.Storage, "1.4.3")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/zot.pid",
		LogFile: config.Storage + "/zot.log",
	})
	mgr := NewZotManagerWithDeps(config, mockBinary, mockProcess)

	ctx := context.Background()

	// BUG: currently returns ErrPortInUse.
	// FIXED: 401 on /v2/ is a healthy Zot response (auth-required) — adopt it.
	startErr := mgr.Start(ctx)
	assert.NoError(t, startErr,
		"Start should return nil when the port is occupied by an auth-enabled Zot instance "+
			"(/v2/ returns 401)")
}

// TestZotManager_Start_PortInUse_UnhealthyService verifies that Start() still
// returns ErrPortInUse when the port is occupied by something that does NOT
// respond with a healthy status on /v2/.
func TestZotManager_Start_PortInUse_UnhealthyService(t *testing.T) {
	// Stand up a fake service that returns 500 on all paths.
	// We bind on 0.0.0.0 so that IsPortAvailable() sees the port as taken.
	port := bindAndServeZot(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      port,
		Storage:   t.TempDir(),
	}
	mockBinary := NewMockBinaryManager(config.Storage, "1.4.3")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/zot.pid",
		LogFile: config.Storage + "/zot.log",
	})
	mgr := NewZotManagerWithDeps(config, mockBinary, mockProcess)

	ctx := context.Background()

	startErr := mgr.Start(ctx)
	assert.Error(t, startErr, "Start should return an error when port is occupied by an unhealthy service")
	assert.ErrorIs(t, startErr, ErrPortInUse,
		"Start should return ErrPortInUse when health probe fails")
}
