package registry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupTestAthensManager creates an AthensManager with test configuration.
func setupTestAthensManager(t *testing.T) GoModuleProxy {
	t.Helper()

	config := GoModuleConfig{
		Enabled:     true,
		Lifecycle:   "manual",
		Port:        13000, // Use high port for tests
		Storage:     t.TempDir(),
		IdleTimeout: 1 * time.Minute,
		Upstreams: []UpstreamProxyConfig{
			{
				Name: "proxy.golang.org",
				URL:  "https://proxy.golang.org",
			},
		},
	}

	// Use mock binary manager for unit tests
	mockBinary := NewMockBinaryManager(config.Storage, "v0.13.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/athens.pid",
		LogFile: config.Storage + "/athens.log",
	})

	return NewAthensManagerWithDeps(config, mockBinary, mockProcess)
}

// =============================================================================
// Task 1: Athens Start Tests
// =============================================================================

func TestAthensManager_Start_Success(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err, "Start should succeed with valid config")

	// Verify proxy is running
	assert.True(t, mgr.IsRunning(ctx), "IsRunning should return true after Start")

	// Verify endpoint is available
	endpoint := mgr.GetEndpoint()
	assert.NotEmpty(t, endpoint, "GetEndpoint should return non-empty string")
	assert.Contains(t, endpoint, "http://", "Endpoint should be HTTP URL")
	assert.Contains(t, endpoint, "localhost", "Endpoint should contain localhost")

	// Cleanup
	defer mgr.Stop(ctx)
}

func TestAthensManager_Start_AlreadyRunning(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	// Start first time
	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Start second time - should handle gracefully
	err = mgr.Start(ctx)
	assert.NoError(t, err, "Starting already-running proxy should be idempotent")
}

func TestAthensManager_Start_BinaryNotFound(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	// Create manager with invalid binary path
	config := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      13002,
		Storage:   t.TempDir(),
	}

	mgr := NewAthensManager(config)
	ctx := context.Background()

	// Mock BinaryManager to return error
	// (This test will need proper mock injection in implementation)

	err := mgr.Start(ctx)
	assert.Error(t, err, "Start should fail if binary not found")
	assert.Contains(t, err.Error(), "binary", "Error should mention binary")
}

func TestAthensManager_Start_PortInUse(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	// Start first proxy
	mgr1 := setupTestAthensManager(t)
	ctx := context.Background()

	err := mgr1.Start(ctx)
	require.NoError(t, err)
	defer mgr1.Stop(ctx)

	// Try to start second proxy on same port
	config := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      13000, // Same port
		Storage:   t.TempDir(),
	}
	mgr2 := NewAthensManager(config)

	err = mgr2.Start(ctx)
	assert.Error(t, err, "Start should fail if port is in use")
	assert.Contains(t, err.Error(), "port", "Error should mention port conflict")
}

// =============================================================================
// Task 2: Athens Stop Tests
// =============================================================================

func TestAthensManager_Stop_Success(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	// Start proxy
	err := mgr.Start(ctx)
	require.NoError(t, err)

	// Stop proxy
	err = mgr.Stop(ctx)
	require.NoError(t, err, "Stop should succeed")

	// Verify proxy is stopped
	assert.False(t, mgr.IsRunning(ctx), "IsRunning should return false after Stop")
}

func TestAthensManager_Stop_NotRunning(t *testing.T) {
	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	// Stop when not running - should be idempotent
	err := mgr.Stop(ctx)
	assert.NoError(t, err, "Stopping non-running proxy should be idempotent")
}

func TestAthensManager_Stop_GracefulShutdown(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)

	// Get PID before stopping
	status, err := mgr.Status(ctx)
	require.NoError(t, err)
	pid := status.PID

	// Stop proxy with timeout
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = mgr.Stop(stopCtx)
	require.NoError(t, err, "Stop should complete within timeout")

	// Verify process is actually terminated
	assert.Greater(t, pid, 0, "PID should have been set when running")
	assert.False(t, mgr.IsRunning(ctx), "Process should be stopped")
}

// =============================================================================
// Task 3: Athens Status Tests
// =============================================================================

func TestAthensManager_Status_Running(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	status, err := mgr.Status(ctx)
	require.NoError(t, err, "Status should succeed when running")

	assert.Equal(t, "running", status.State, "State should be 'running'")
	assert.Greater(t, status.PID, 0, "PID should be set")
	assert.Equal(t, 13000, status.Port, "Port should match config")
	assert.NotEmpty(t, status.Storage, "Storage path should be set")
	assert.NotEmpty(t, status.Version, "Version should be set")
	assert.GreaterOrEqual(t, status.Uptime, time.Duration(0), "Uptime should be non-negative")
}

func TestAthensManager_Status_Stopped(t *testing.T) {
	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	status, err := mgr.Status(ctx)
	require.NoError(t, err, "Status should succeed even when stopped")

	assert.Equal(t, "stopped", status.State, "State should be 'stopped'")
	assert.Equal(t, 0, status.PID, "PID should be 0 when stopped")
	assert.Equal(t, 13000, status.Port, "Port should still be in config")
	assert.NotEmpty(t, status.Storage, "Storage path should be in config")
}

func TestAthensManager_Status_ModuleCount(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Fetch a test module (would require Go client or curl)
	// This would require integration with Go tools

	status, err := mgr.Status(ctx)
	require.NoError(t, err)

	// Initially should be 0
	assert.GreaterOrEqual(t, status.ModuleCount, 0, "ModuleCount should be non-negative")
	assert.GreaterOrEqual(t, status.DiskUsage, int64(0), "DiskUsage should be non-negative")
}

// =============================================================================
// Task 4: EnsureRunning Tests
// =============================================================================

func TestAthensManager_EnsureRunning_StartsIfStopped(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	mgr := setupTestAthensManager(t)
	ctx := context.Background()
	defer mgr.Stop(ctx)

	// Verify not running initially
	assert.False(t, mgr.IsRunning(ctx), "Should not be running initially")

	// EnsureRunning should start the proxy
	err := mgr.EnsureRunning(ctx)
	require.NoError(t, err, "EnsureRunning should succeed")

	// Verify now running
	assert.True(t, mgr.IsRunning(ctx), "Should be running after EnsureRunning")
}

func TestAthensManager_EnsureRunning_NoopIfRunning(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	// Start proxy
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
// Task 5: IsRunning Tests
// =============================================================================

func TestAthensManager_IsRunning_True(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	assert.True(t, mgr.IsRunning(ctx), "IsRunning should return true when proxy is running")
}

func TestAthensManager_IsRunning_False(t *testing.T) {
	mgr := setupTestAthensManager(t)
	ctx := context.Background()

	assert.False(t, mgr.IsRunning(ctx), "IsRunning should return false when proxy is not running")
}

func TestAthensManager_IsRunning_AfterCrash(t *testing.T) {
	t.Skip("Integration test - requires real Athens binary with HTTP server")

	mgr := setupTestAthensManager(t)
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
// Task 6: GetEndpoint Tests
// =============================================================================

func TestAthensManager_GetEndpoint_ReturnsCorrectFormat(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantFmt string
	}{
		{
			name:    "default port 3000",
			port:    3000,
			wantFmt: "http://localhost:3000",
		},
		{
			name:    "custom port 3333",
			port:    3333,
			wantFmt: "http://localhost:3333",
		},
		{
			name:    "high port",
			port:    13000,
			wantFmt: "http://localhost:13000",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GoModuleConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      tt.port,
				Storage:   t.TempDir(),
			}
			mgr := NewAthensManager(config)

			endpoint := mgr.GetEndpoint()
			assert.Equal(t, tt.wantFmt, endpoint, "Endpoint format should be http://localhost:port")
		})
	}
}

// =============================================================================
// Task 7: GetGoEnv Tests
// =============================================================================

func TestAthensManager_GetGoEnv_ReturnsCorrectVariables(t *testing.T) {
	config := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      3000,
		Storage:   t.TempDir(),
	}
	mgr := NewAthensManager(config)

	goEnv := mgr.GetGoEnv()

	// Should include GOPROXY
	assert.Contains(t, goEnv, "GOPROXY", "Should include GOPROXY variable")
	assert.Equal(t, "http://localhost:3000", goEnv["GOPROXY"], "GOPROXY should point to Athens endpoint")

	// Should include GONOSUMDB for local development
	assert.Contains(t, goEnv, "GONOSUMDB", "Should include GONOSUMDB for local development")

	// Should include GOPRIVATE as alternative
	assert.Contains(t, goEnv, "GOPRIVATE", "Should include GOPRIVATE as alternative")
}

func TestAthensManager_GetGoEnv_DifferentPorts(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		wantGOPROXY string
	}{
		{
			name:        "default port",
			port:        3000,
			wantGOPROXY: "http://localhost:3000",
		},
		{
			name:        "custom port",
			port:        3333,
			wantGOPROXY: "http://localhost:3333",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GoModuleConfig{
				Enabled: true,
				Port:    tt.port,
				Storage: t.TempDir(),
			}
			mgr := NewAthensManager(config)

			goEnv := mgr.GetGoEnv()

			assert.Equal(t, tt.wantGOPROXY, goEnv["GOPROXY"])
		})
	}
}

func TestAthensManager_GetGoEnv_NotRunning(t *testing.T) {
	// GetGoEnv should return variables even if proxy is not running
	// (user might want to set env vars before starting)

	config := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      3000,
		Storage:   t.TempDir(),
	}
	mgr := NewAthensManager(config)

	goEnv := mgr.GetGoEnv()

	// Should still return valid environment variables
	assert.Contains(t, goEnv, "GOPROXY")
	assert.NotEmpty(t, goEnv["GOPROXY"])
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestAthensManager_ImplementsGoModuleProxy(t *testing.T) {
	var _ GoModuleProxy = (*AthensManager)(nil)
}
