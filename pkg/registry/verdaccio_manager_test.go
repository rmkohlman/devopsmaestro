package registry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupTestVerdaccioManager creates a VerdaccioManager with test configuration.
func setupTestVerdaccioManager(t *testing.T) NpmProxy {
	t.Helper()

	config := NpmProxyConfig{
		Enabled:     true,
		Lifecycle:   "manual",
		Port:        14873, // Use high port for tests
		Storage:     t.TempDir(),
		IdleTimeout: 1 * time.Minute,
		Upstreams: []NpmUpstreamConfig{
			{
				Name: "npmjs",
				URL:  "https://registry.npmjs.org",
			},
		},
	}

	// Use mock binary manager for unit tests
	mockBinary := NewMockNpmBinaryManager(config.Storage, "5.28.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/verdaccio.pid",
		LogFile: config.Storage + "/verdaccio.log",
	})

	return NewVerdaccioManagerWithDeps(config, mockBinary, mockProcess)
}

// =============================================================================
// Task 1: VerdaccioManager Start Tests
// =============================================================================

func TestVerdaccioManager_Start_Success(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary with HTTP server")

	mgr := setupTestVerdaccioManager(t)
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

func TestVerdaccioManager_Start_AlreadyRunning(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	mgr := setupTestVerdaccioManager(t)
	ctx := context.Background()

	// Start first time
	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Start second time - should handle gracefully (idempotent)
	err = mgr.Start(ctx)
	assert.NoError(t, err, "Starting already-running proxy should be idempotent")
}

func TestVerdaccioManager_Start_BinaryNotFound(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	// Create manager with invalid binary path
	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      14874,
		Storage:   t.TempDir(),
	}

	mgr := NewVerdaccioManager(config)
	ctx := context.Background()

	// Mock BinaryManager to return error
	// (This test will need proper mock injection in implementation)

	err := mgr.Start(ctx)
	assert.Error(t, err, "Start should fail if binary not found")
	assert.Contains(t, err.Error(), "binary", "Error should mention binary")
}

func TestVerdaccioManager_Start_PortInUse(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	// Start first proxy
	mgr1 := setupTestVerdaccioManager(t)
	ctx := context.Background()

	err := mgr1.Start(ctx)
	require.NoError(t, err)
	defer mgr1.Stop(ctx)

	// Try to start second proxy on same port
	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      14873, // Same port
		Storage:   t.TempDir(),
	}
	mgr2 := NewVerdaccioManager(config)

	err = mgr2.Start(ctx)
	assert.Error(t, err, "Start should fail if port is in use")
	assert.ErrorIs(t, err, ErrPortInUse, "Should return ErrPortInUse")
	assert.Contains(t, err.Error(), "port", "Error should mention port conflict")
}

func TestVerdaccioManager_Start_ConfigFileGeneration(t *testing.T) {
	t.Skip("Integration test - requires verdaccio binary")

	mgr := setupTestVerdaccioManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Verify config file was generated
	// (Would need to check actual file system)
	assert.True(t, mgr.IsRunning(ctx))
}

// =============================================================================
// Task 2: VerdaccioManager Stop Tests
// =============================================================================

func TestVerdaccioManager_Stop_Success(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	mgr := setupTestVerdaccioManager(t)
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

func TestVerdaccioManager_Stop_NotRunning(t *testing.T) {
	mgr := setupTestVerdaccioManager(t)
	ctx := context.Background()

	// Stop when not running - should be idempotent
	err := mgr.Stop(ctx)
	assert.NoError(t, err, "Stopping non-running proxy should be idempotent")
}

func TestVerdaccioManager_Stop_GracefulShutdown(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	mgr := setupTestVerdaccioManager(t)
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
// Task 3: VerdaccioManager Status Tests
// =============================================================================

func TestVerdaccioManager_Status_Running(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	mgr := setupTestVerdaccioManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	status, err := mgr.Status(ctx)
	require.NoError(t, err, "Status should succeed when running")

	assert.Equal(t, "running", status.State, "State should be 'running'")
	assert.Greater(t, status.PID, 0, "PID should be set")
	assert.Equal(t, 14873, status.Port, "Port should match config")
	assert.NotEmpty(t, status.Storage, "Storage path should be set")
	assert.NotEmpty(t, status.Version, "Version should be set")
	assert.GreaterOrEqual(t, status.Uptime, time.Duration(0), "Uptime should be non-negative")
}

func TestVerdaccioManager_Status_Stopped(t *testing.T) {
	mgr := setupTestVerdaccioManager(t)
	ctx := context.Background()

	status, err := mgr.Status(ctx)
	require.NoError(t, err, "Status should succeed even when stopped")

	assert.Equal(t, "stopped", status.State, "State should be 'stopped'")
	assert.Equal(t, 0, status.PID, "PID should be 0 when stopped")
	assert.Equal(t, 14873, status.Port, "Port should still be in config")
	assert.NotEmpty(t, status.Storage, "Storage path should be in config")
}

func TestVerdaccioManager_Status_PackageCount(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	mgr := setupTestVerdaccioManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	status, err := mgr.Status(ctx)
	require.NoError(t, err)

	// Initially should be 0
	assert.GreaterOrEqual(t, status.PackageCount, 0, "PackageCount should be non-negative")
	assert.GreaterOrEqual(t, status.DiskUsage, int64(0), "DiskUsage should be non-negative")
}

// =============================================================================
// Task 4: EnsureRunning Tests
// =============================================================================

func TestVerdaccioManager_EnsureRunning_StartsIfStopped(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	mgr := setupTestVerdaccioManager(t)
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

func TestVerdaccioManager_EnsureRunning_NoopIfRunning(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	mgr := setupTestVerdaccioManager(t)
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

func TestVerdaccioManager_IsRunning_True(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	mgr := setupTestVerdaccioManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	assert.True(t, mgr.IsRunning(ctx), "IsRunning should return true when proxy is running")
}

func TestVerdaccioManager_IsRunning_False(t *testing.T) {
	mgr := setupTestVerdaccioManager(t)
	ctx := context.Background()

	assert.False(t, mgr.IsRunning(ctx), "IsRunning should return false when proxy is not running")
}

func TestVerdaccioManager_IsRunning_AfterCrash(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio binary")

	mgr := setupTestVerdaccioManager(t)
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

func TestVerdaccioManager_GetEndpoint_ReturnsCorrectFormat(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantFmt string
	}{
		{
			name:    "default port 4873",
			port:    4873,
			wantFmt: "http://localhost:4873",
		},
		{
			name:    "custom port 4874",
			port:    4874,
			wantFmt: "http://localhost:4874",
		},
		{
			name:    "high port",
			port:    14873,
			wantFmt: "http://localhost:14873",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NpmProxyConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      tt.port,
				Storage:   t.TempDir(),
			}
			mgr := NewVerdaccioManager(config)

			endpoint := mgr.GetEndpoint()
			assert.Equal(t, tt.wantFmt, endpoint, "Endpoint format should be http://localhost:port")
		})
	}
}

// =============================================================================
// Task 7: GetNpmEnv Tests
// =============================================================================

func TestVerdaccioManager_GetNpmEnv_ReturnsCorrectVariables(t *testing.T) {
	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      4873,
		Storage:   t.TempDir(),
	}
	mgr := NewVerdaccioManager(config)

	npmEnv := mgr.GetNpmEnv()

	// Should include npm_config_registry
	assert.Contains(t, npmEnv, "npm_config_registry", "Should include npm_config_registry variable")
	assert.Equal(t, "http://localhost:4873", npmEnv["npm_config_registry"], "npm_config_registry should point to verdaccio endpoint")

	// Should include NPM_CONFIG_REGISTRY (uppercase)
	assert.Contains(t, npmEnv, "NPM_CONFIG_REGISTRY", "Should include uppercase NPM_CONFIG_REGISTRY")
}

func TestVerdaccioManager_GetNpmEnv_DifferentPorts(t *testing.T) {
	tests := []struct {
		name                    string
		port                    int
		wantNPM_CONFIG_REGISTRY string
	}{
		{
			name:                    "default port",
			port:                    4873,
			wantNPM_CONFIG_REGISTRY: "http://localhost:4873",
		},
		{
			name:                    "custom port",
			port:                    4874,
			wantNPM_CONFIG_REGISTRY: "http://localhost:4874",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NpmProxyConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      tt.port,
				Storage:   t.TempDir(),
			}
			mgr := NewVerdaccioManager(config)

			npmEnv := mgr.GetNpmEnv()

			assert.Equal(t, tt.wantNPM_CONFIG_REGISTRY, npmEnv["npm_config_registry"])
		})
	}
}

func TestVerdaccioManager_GetNpmEnv_NotRunning(t *testing.T) {
	// GetNpmEnv should return variables even if proxy is not running
	// (user might want to set env vars before starting)

	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      4873,
		Storage:   t.TempDir(),
	}
	mgr := NewVerdaccioManager(config)

	npmEnv := mgr.GetNpmEnv()

	// Should still return valid environment variables
	assert.Contains(t, npmEnv, "npm_config_registry")
	assert.NotEmpty(t, npmEnv["npm_config_registry"])
}

// =============================================================================
// Task 8: GetNpmrc Tests
// =============================================================================

func TestVerdaccioManager_GetNpmrc_ReturnsCorrectFormat(t *testing.T) {
	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      4873,
		Storage:   t.TempDir(),
	}
	mgr := NewVerdaccioManager(config)

	npmrc := mgr.GetNpmrc()

	// Should contain registry setting
	assert.Contains(t, npmrc, "registry", "Should contain registry setting")
	assert.Contains(t, npmrc, "http://localhost:4873", "Should reference verdaccio endpoint")

	// Should be in .npmrc format (key=value)
	assert.Contains(t, npmrc, "=", "Should use key=value format")
}

func TestVerdaccioManager_GetNpmrc_DifferentPorts(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"default port", 4873},
		{"custom port", 4874},
		{"high port", 14873},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NpmProxyConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      tt.port,
				Storage:   t.TempDir(),
			}
			mgr := NewVerdaccioManager(config)

			npmrc := mgr.GetNpmrc()

			// Should contain the port number
			assert.Contains(t, npmrc, fmt.Sprintf("%d", tt.port))
		})
	}
}

// =============================================================================
// Task 9: GetYarnConfig Tests
// =============================================================================

func TestVerdaccioManager_GetYarnConfig_ReturnsCorrectFormat(t *testing.T) {
	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      4873,
		Storage:   t.TempDir(),
	}
	mgr := NewVerdaccioManager(config)

	yarnConfig := mgr.GetYarnConfig()

	// Should contain yarn registry setting
	assert.Contains(t, yarnConfig, "npmRegistryServer", "Should contain npmRegistryServer setting")
	assert.Contains(t, yarnConfig, "http://localhost:4873", "Should reference verdaccio endpoint")

	// Should be in .yarnrc.yml format (YAML)
	// Could be either format depending on Yarn version
}

func TestVerdaccioManager_GetYarnConfig_DifferentPorts(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"default port", 4873},
		{"custom port", 4874},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NpmProxyConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      tt.port,
				Storage:   t.TempDir(),
			}
			mgr := NewVerdaccioManager(config)

			yarnConfig := mgr.GetYarnConfig()

			// Should contain the port number
			assert.Contains(t, yarnConfig, fmt.Sprintf("%d", tt.port))
		})
	}
}

// =============================================================================
// Task 10: IdleTimeout Tests (On-Demand Mode)
// =============================================================================

func TestVerdaccioManager_IdleTimeout_OnDemand(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio and time-based behavior")

	config := NpmProxyConfig{
		Enabled:     true,
		Lifecycle:   "on-demand",
		Port:        14875,
		Storage:     t.TempDir(),
		IdleTimeout: 500 * time.Millisecond, // Short timeout for testing
	}

	mgr := NewVerdaccioManager(config)
	ctx := context.Background()

	// Start proxy
	err := mgr.Start(ctx)
	require.NoError(t, err)

	// Verify running
	assert.True(t, mgr.IsRunning(ctx))

	// Wait for idle timeout
	time.Sleep(1 * time.Second)

	// Should have stopped automatically
	assert.False(t, mgr.IsRunning(ctx), "Proxy should stop after idle timeout")
}

func TestVerdaccioManager_IdleTimeout_ResetOnAccess(t *testing.T) {
	t.Skip("Integration test - requires real verdaccio and time-based behavior")

	config := NpmProxyConfig{
		Enabled:     true,
		Lifecycle:   "on-demand",
		Port:        14876,
		Storage:     t.TempDir(),
		IdleTimeout: 1 * time.Second,
	}

	mgr := NewVerdaccioManager(config)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)

	// Access endpoint to reset timer
	_ = mgr.GetEndpoint()
	time.Sleep(600 * time.Millisecond)

	// Access again to reset
	_ = mgr.GetEndpoint()
	time.Sleep(600 * time.Millisecond)

	// Should still be running (timer was reset)
	assert.True(t, mgr.IsRunning(ctx), "Timer should reset on access")
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestVerdaccioManager_ImplementsNpmProxy(t *testing.T) {
	var _ NpmProxy = (*VerdaccioManager)(nil)
}
