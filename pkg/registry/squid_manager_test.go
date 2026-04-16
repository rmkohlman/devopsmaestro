package registry

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupTestSquidManager creates a SquidManager with test configuration.
func setupTestSquidManager(t *testing.T) HttpProxy {
	t.Helper()

	config := HttpProxyConfig{
		Port:            13128, // Use high port for tests
		CacheDir:        t.TempDir() + "/cache",
		LogDir:          t.TempDir() + "/logs",
		PidFile:         t.TempDir() + "/squid.pid",
		CacheSizeMB:     100,
		MaxObjectSizeMB: 10,
		MemoryCacheMB:   64,
	}

	// Use mock binary manager for unit tests
	mockBinary := NewMockBrewBinaryManager(t.TempDir(), "6.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.PidFile,
		LogFile: config.LogDir + "/squid.log",
	})

	return NewSquidManagerWithDeps(config, mockBinary, mockProcess)
}

// =============================================================================
// SquidManager Start Tests
// =============================================================================

func TestSquidManager_Start_Success(t *testing.T) {
	t.Skip("Integration test - requires real squid binary with HTTP server")

	mgr := setupTestSquidManager(t)
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

func TestSquidManager_Start_GeneratesConfig(t *testing.T) {
	t.Skip("Integration test - requires squid binary")

	mgr := setupTestSquidManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Verify squid.conf was generated
	// (Would need to check actual file system)
	assert.True(t, mgr.IsRunning(ctx))
}

func TestSquidManager_Start_InitializesCache(t *testing.T) {
	t.Skip("Integration test - requires squid binary")

	mgr := setupTestSquidManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Verify cache directories were initialized (squid -z)
	// (Would need to check cache directory structure)
	assert.True(t, mgr.IsRunning(ctx))
}

func TestSquidManager_Start_AlreadyRunning(t *testing.T) {
	t.Skip("Integration test - requires real squid binary")

	mgr := setupTestSquidManager(t)
	ctx := context.Background()

	// Start first time
	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Start second time - should handle gracefully (idempotent)
	err = mgr.Start(ctx)
	assert.NoError(t, err, "Starting already-running proxy should be idempotent")
}

func TestSquidManager_Start_BinaryNotFound(t *testing.T) {
	t.Skip("Integration test - requires real squid binary")

	// Create manager with invalid binary path
	config := HttpProxyConfig{
		Port:            13129,
		CacheDir:        t.TempDir() + "/cache",
		LogDir:          t.TempDir() + "/logs",
		PidFile:         t.TempDir() + "/squid.pid",
		CacheSizeMB:     100,
		MaxObjectSizeMB: 10,
		MemoryCacheMB:   64,
	}

	mgr := NewSquidManager(config)
	ctx := context.Background()

	err := mgr.Start(ctx)
	assert.Error(t, err, "Start should fail if binary not found")
	assert.Contains(t, err.Error(), "binary", "Error should mention binary")
}

func TestSquidManager_Start_PortInUse(t *testing.T) {
	t.Skip("Integration test - requires real squid binary")

	// Start first proxy
	mgr1 := setupTestSquidManager(t)
	ctx := context.Background()

	err := mgr1.Start(ctx)
	require.NoError(t, err)
	defer mgr1.Stop(ctx)

	// Try to start second proxy on same port
	config := HttpProxyConfig{
		Port:            13128, // Same port
		CacheDir:        t.TempDir() + "/cache",
		LogDir:          t.TempDir() + "/logs",
		PidFile:         t.TempDir() + "/squid.pid",
		CacheSizeMB:     100,
		MaxObjectSizeMB: 10,
		MemoryCacheMB:   64,
	}
	mgr2 := NewSquidManager(config)

	err = mgr2.Start(ctx)
	assert.Error(t, err, "Start should fail if port is in use")
	assert.ErrorIs(t, err, ErrPortInUse, "Should return ErrPortInUse")
	assert.Contains(t, err.Error(), "port", "Error should mention port conflict")
}

// =============================================================================
// SquidManager Stop Tests
// =============================================================================

func TestSquidManager_Stop_Success(t *testing.T) {
	t.Skip("Integration test - requires real squid binary")

	mgr := setupTestSquidManager(t)
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

func TestSquidManager_Stop_NotRunning(t *testing.T) {
	mgr := setupTestSquidManager(t)
	ctx := context.Background()

	// Stop when not running - should be idempotent
	err := mgr.Stop(ctx)
	assert.NoError(t, err, "Stopping non-running proxy should be idempotent")
}

func TestSquidManager_Stop_GracefulShutdown(t *testing.T) {
	t.Skip("Integration test - requires real squid binary")

	mgr := setupTestSquidManager(t)
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
// SquidManager Status Tests
// =============================================================================

func TestSquidManager_Status_Running(t *testing.T) {
	t.Skip("Integration test - requires real squid binary")

	mgr := setupTestSquidManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	status, err := mgr.Status(ctx)
	require.NoError(t, err, "Status should succeed when running")

	assert.Equal(t, "running", status.State, "State should be 'running'")
	assert.Greater(t, status.PID, 0, "PID should be set")
	assert.Equal(t, 13128, status.Port, "Port should match config")
	assert.NotEmpty(t, status.CacheDir, "CacheDir should be set")
	assert.NotEmpty(t, status.Version, "Version should be set")
	assert.GreaterOrEqual(t, status.Uptime, time.Duration(0), "Uptime should be non-negative")
}

func TestSquidManager_Status_Stopped(t *testing.T) {
	mgr := setupTestSquidManager(t)
	ctx := context.Background()

	status, err := mgr.Status(ctx)
	require.NoError(t, err, "Status should succeed even when stopped")

	assert.Equal(t, "stopped", status.State, "State should be 'stopped'")
	assert.Equal(t, 0, status.PID, "PID should be 0 when stopped")
	assert.Equal(t, 13128, status.Port, "Port should still be in config")
	assert.NotEmpty(t, status.CacheDir, "CacheDir should be in config")
}

// =============================================================================
// SquidManager EnsureRunning Tests
// =============================================================================

func TestSquidManager_EnsureRunning_StartsIfStopped(t *testing.T) {
	t.Skip("Integration test - requires real squid binary")

	mgr := setupTestSquidManager(t)
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

func TestSquidManager_EnsureRunning_NoopIfRunning(t *testing.T) {
	t.Skip("Integration test - requires real squid binary")

	mgr := setupTestSquidManager(t)
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
// SquidManager IsRunning Tests
// =============================================================================

func TestSquidManager_IsRunning_True(t *testing.T) {
	t.Skip("Integration test - requires real squid binary")

	mgr := setupTestSquidManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	assert.True(t, mgr.IsRunning(ctx), "IsRunning should return true when proxy is running")
}

func TestSquidManager_IsRunning_False(t *testing.T) {
	mgr := setupTestSquidManager(t)
	ctx := context.Background()

	assert.False(t, mgr.IsRunning(ctx), "IsRunning should return false when proxy is not running")
}

func TestSquidManager_IsRunning_AfterCrash(t *testing.T) {
	t.Skip("Integration test - requires real squid binary")

	mgr := setupTestSquidManager(t)
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
// SquidManager GetEndpoint Tests
// =============================================================================

func TestSquidManager_GetEndpoint_ReturnsCorrectFormat(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantFmt string
	}{
		{
			name:    "default port 3128",
			port:    3128,
			wantFmt: "http://localhost:3128",
		},
		{
			name:    "custom port 8080",
			port:    8080,
			wantFmt: "http://localhost:8080",
		},
		{
			name:    "high port",
			port:    13128,
			wantFmt: "http://localhost:13128",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HttpProxyConfig{
				Port:            tt.port,
				CacheDir:        t.TempDir() + "/cache",
				LogDir:          t.TempDir() + "/logs",
				PidFile:         t.TempDir() + "/squid.pid",
				CacheSizeMB:     100,
				MaxObjectSizeMB: 10,
				MemoryCacheMB:   64,
			}
			mgr := NewSquidManager(config)

			endpoint := mgr.GetEndpoint()
			assert.Equal(t, tt.wantFmt, endpoint, "Endpoint format should be http://localhost:port")
		})
	}
}

// =============================================================================
// SquidManager GetProxyEnv Tests
// =============================================================================

func TestSquidManager_GetProxyEnv_ReturnsCorrectVariables(t *testing.T) {
	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        t.TempDir() + "/cache",
		LogDir:          t.TempDir() + "/logs",
		PidFile:         t.TempDir() + "/squid.pid",
		CacheSizeMB:     100,
		MaxObjectSizeMB: 10,
		MemoryCacheMB:   64,
	}
	mgr := NewSquidManager(config)

	proxyEnv := mgr.GetProxyEnv()

	// Should include HTTP_PROXY
	assert.Contains(t, proxyEnv, "HTTP_PROXY", "Should include HTTP_PROXY variable")
	assert.Equal(t, "http://localhost:3128", proxyEnv["HTTP_PROXY"], "HTTP_PROXY should point to squid endpoint")

	// Should include HTTPS_PROXY
	assert.Contains(t, proxyEnv, "HTTPS_PROXY", "Should include HTTPS_PROXY variable")
	assert.Equal(t, "http://localhost:3128", proxyEnv["HTTPS_PROXY"], "HTTPS_PROXY should point to squid endpoint")

	// Should include NO_PROXY
	assert.Contains(t, proxyEnv, "NO_PROXY", "Should include NO_PROXY variable")
	assert.Equal(t, "localhost,127.0.0.1", proxyEnv["NO_PROXY"], "NO_PROXY should exclude localhost")
}

func TestSquidManager_GetProxyEnv_DifferentPorts(t *testing.T) {
	tests := []struct {
		name            string
		port            int
		wantHTTP_PROXY  string
		wantHTTPS_PROXY string
	}{
		{
			name:            "default port",
			port:            3128,
			wantHTTP_PROXY:  "http://localhost:3128",
			wantHTTPS_PROXY: "http://localhost:3128",
		},
		{
			name:            "custom port",
			port:            8080,
			wantHTTP_PROXY:  "http://localhost:8080",
			wantHTTPS_PROXY: "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HttpProxyConfig{
				Port:            tt.port,
				CacheDir:        t.TempDir() + "/cache",
				LogDir:          t.TempDir() + "/logs",
				PidFile:         t.TempDir() + "/squid.pid",
				CacheSizeMB:     100,
				MaxObjectSizeMB: 10,
				MemoryCacheMB:   64,
			}
			mgr := NewSquidManager(config)

			proxyEnv := mgr.GetProxyEnv()

			assert.Equal(t, tt.wantHTTP_PROXY, proxyEnv["HTTP_PROXY"])
			assert.Equal(t, tt.wantHTTPS_PROXY, proxyEnv["HTTPS_PROXY"])
		})
	}
}

func TestSquidManager_GetProxyEnv_NotRunning(t *testing.T) {
	// GetProxyEnv should return variables even if proxy is not running
	// (user might want to set env vars before starting)

	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        t.TempDir() + "/cache",
		LogDir:          t.TempDir() + "/logs",
		PidFile:         t.TempDir() + "/squid.pid",
		CacheSizeMB:     100,
		MaxObjectSizeMB: 10,
		MemoryCacheMB:   64,
	}
	mgr := NewSquidManager(config)

	proxyEnv := mgr.GetProxyEnv()

	// Should still return valid environment variables
	assert.Contains(t, proxyEnv, "HTTP_PROXY")
	assert.NotEmpty(t, proxyEnv["HTTP_PROXY"])
}

// =============================================================================
// SquidManager Config Generation Tests
// =============================================================================

func TestSquidManager_GenerateSquidConf_Basic(t *testing.T) {
	t.Skip("Integration test - requires squid.conf template")

	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	squidConf, err := GenerateSquidConfig(config)
	require.NoError(t, err, "Should generate config successfully")
	assert.NotEmpty(t, squidConf, "Generated config should not be empty")

	// Config should contain key settings
	assert.Contains(t, squidConf, "http_port 0.0.0.0:3128", "Config should contain explicit 0.0.0.0 binding for container access")
	assert.Contains(t, squidConf, "/tmp/squid/cache", "Config should contain cache dir")
	assert.Contains(t, squidConf, "cache_dir", "Config should have cache_dir directive")
}

func TestSquidManager_GenerateSquidConf_CacheSizes(t *testing.T) {
	t.Skip("Integration test - requires squid.conf template")

	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     5000,
		MaxObjectSizeMB: 500,
		MemoryCacheMB:   512,
	}

	squidConf, err := GenerateSquidConfig(config)
	require.NoError(t, err)

	// Verify cache sizes are in config
	assert.Contains(t, squidConf, "5000", "Config should reference cache size")
	assert.Contains(t, squidConf, "maximum_object_size", "Config should have max object size")
	assert.Contains(t, squidConf, "cache_mem", "Config should have memory cache setting")
}

func TestSquidManager_GenerateSquidConf_InvalidConfig(t *testing.T) {
	config := HttpProxyConfig{
		Port:            0, // Invalid port
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	_, err := GenerateSquidConfig(config)
	assert.Error(t, err, "Should fail with invalid config")
	assert.Contains(t, err.Error(), "invalid", "Error should mention validation failure")
}

// =============================================================================
// SquidManager Context Cancellation Tests
// =============================================================================

func TestSquidManager_ContextCancellation_Start(t *testing.T) {
	t.Skip("Integration test - requires long-running operation")

	mgr := setupTestSquidManager(t)

	// Create context that cancels immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := mgr.Start(ctx)
	assert.Error(t, err, "Should respect context cancellation")
	assert.ErrorIs(t, err, context.Canceled, "Should return context.Canceled error")
}

func TestSquidManager_ContextCancellation_Stop(t *testing.T) {
	t.Skip("Integration test - requires long-running operation")

	mgr := setupTestSquidManager(t)
	ctx := context.Background()

	// Start first
	err := mgr.Start(ctx)
	require.NoError(t, err)

	// Create context that cancels during stop
	stopCtx, cancel := context.WithCancel(context.Background())
	cancel()

	err = mgr.Stop(stopCtx)
	// Behavior depends on implementation - may succeed or return context error
	_ = err
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestSquidManager_ImplementsHttpProxy(t *testing.T) {
	var _ HttpProxy = (*SquidManager)(nil)
}

func TestSquidManagerAdapter_ImplementsServiceManager(t *testing.T) {
	var _ ServiceManager = (*SquidManagerAdapter)(nil)
}

// =============================================================================
// Config Staleness Detection Tests (#353)
// =============================================================================

func TestSquidManager_ConfigStale_DetectsOldLoopbackBinding(t *testing.T) {
	// Simulate an old config file that used bare port form (pre-#353 fix).
	// The manager should detect this as stale because the new config
	// explicitly binds 0.0.0.0.
	tmpDir := t.TempDir()
	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        tmpDir + "/cache",
		LogDir:          tmpDir + "/logs",
		PidFile:         tmpDir + "/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	mgr := NewSquidManagerWithDeps(config,
		&MockBinaryManager{},
		NewProcessManager(ProcessConfig{}),
	)

	// Write an old-style config with bare port (how v0.96.0 generated it)
	configPath := tmpDir + "/logs/squid.conf"
	if err := os.MkdirAll(tmpDir+"/logs", 0700); err != nil {
		t.Fatal(err)
	}
	oldConfig := "# Generated by DevOpsMaestro\nhttp_port 3128\n"
	if err := os.WriteFile(configPath, []byte(oldConfig), 0600); err != nil {
		t.Fatal(err)
	}

	stale, err := mgr.configStale(configPath)
	assert.NoError(t, err)
	assert.True(t, stale, "Old bare-port config should be detected as stale")
}

func TestSquidManager_ConfigStale_CurrentConfigNotStale(t *testing.T) {
	// A config generated by the current code should not be stale.
	tmpDir := t.TempDir()
	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        tmpDir + "/cache",
		LogDir:          tmpDir + "/logs",
		PidFile:         tmpDir + "/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	mgr := NewSquidManagerWithDeps(config,
		&MockBinaryManager{},
		NewProcessManager(ProcessConfig{}),
	)

	// Generate the current config and write it
	configPath := tmpDir + "/logs/squid.conf"
	if err := os.MkdirAll(tmpDir+"/logs", 0700); err != nil {
		t.Fatal(err)
	}
	currentConfig, err := GenerateSquidConfig(config)
	require.NoError(t, err)
	if err := os.WriteFile(configPath, []byte(currentConfig), 0600); err != nil {
		t.Fatal(err)
	}

	stale, err := mgr.configStale(configPath)
	assert.NoError(t, err)
	assert.False(t, stale, "Current config should not be detected as stale")
}

func TestSquidManager_ConfigStale_MissingFileIsStale(t *testing.T) {
	// If the config file doesn't exist, treat as stale.
	tmpDir := t.TempDir()
	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        tmpDir + "/cache",
		LogDir:          tmpDir + "/logs",
		PidFile:         tmpDir + "/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	mgr := NewSquidManagerWithDeps(config,
		&MockBinaryManager{},
		NewProcessManager(ProcessConfig{}),
	)

	stale, _ := mgr.configStale(tmpDir + "/nonexistent/squid.conf")
	assert.True(t, stale, "Missing config file should be treated as stale")
}

func TestGenerateSquidConfig_ExplicitBindAddress(t *testing.T) {
	// Verify that the generated config uses explicit 0.0.0.0 binding
	// rather than bare port form (see #353).
	tests := []struct {
		name string
		port int
		want string
	}{
		{"default port", 3128, "http_port 0.0.0.0:3128"},
		{"custom port", 8080, "http_port 0.0.0.0:8080"},
		{"high port", 13128, "http_port 0.0.0.0:13128"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := HttpProxyConfig{
				Port:            tt.port,
				CacheDir:        "/tmp/squid/cache",
				LogDir:          "/tmp/squid/logs",
				PidFile:         "/tmp/squid/squid.pid",
				CacheSizeMB:     1000,
				MaxObjectSizeMB: 100,
				MemoryCacheMB:   256,
			}

			config, err := GenerateSquidConfig(cfg)
			require.NoError(t, err)
			assert.Contains(t, config, tt.want,
				"Config must use explicit 0.0.0.0 binding, not bare port form")
		})
	}
}

func TestGenerateSquidConfig_NoBarePortForm(t *testing.T) {
	// Ensure the config does NOT contain bare port form (without address).
	// Bare "http_port 3128" has version-dependent binding behavior.
	cfg := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	config, err := GenerateSquidConfig(cfg)
	require.NoError(t, err)

	// The config should contain "http_port 0.0.0.0:3128" NOT "http_port 3128"
	assert.Contains(t, config, "http_port 0.0.0.0:3128")
	// Make sure bare form without address prefix does not appear.
	// We check that "http_port 3128" as a standalone (not part of 0.0.0.0:3128) is absent.
	assert.NotContains(t, config, "http_port 3128\n",
		"Config must NOT use bare port form — it has platform-dependent binding behavior")
}

// =============================================================================
// Fix #363 — Squid xcalloc integer overflow
// =============================================================================

// TestGenerateSquidConfig_CacheDirL2Is64 verifies that the cache_dir directive
// uses L2=64 (not 256) to avoid the xcalloc integer overflow in Squid 7.x (#363).
func TestGenerateSquidConfig_CacheDirL2Is64(t *testing.T) {
	cfg := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	config, err := GenerateSquidConfig(cfg)
	require.NoError(t, err, "GenerateSquidConfig should not error")

	// Fix: L2 must be 64, not 256 — 256 causes xcalloc overflow in Squid 7.x
	assert.Contains(t, config, "16 64",
		"cache_dir must use L2=64 to avoid xcalloc integer overflow in Squid 7.x (#363)")
	assert.NotContains(t, config, "16 256",
		"cache_dir must NOT use L2=256 (causes xcalloc overflow in Squid 7.x, see #363)")
}

// TestGenerateSquidConfig_PidFilenameIsNone verifies that the generated squid.conf
// uses "pid_filename none" so that squid does not interfere with the PID file that
// dvm's DefaultProcessManager writes. Regression test for #373.
func TestGenerateSquidConfig_PidFilenameIsNone(t *testing.T) {
	cfg := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	config, err := GenerateSquidConfig(cfg)
	require.NoError(t, err, "GenerateSquidConfig should not error")

	// Fix #373: pid_filename must be "none" so squid does not overwrite or remove
	// the PID file that dvm's ProcessManager writes after starting the process.
	assert.Contains(t, config, "pid_filename none",
		"squid.conf must set pid_filename none to avoid interfering with dvm PID tracking (#373)")
	assert.NotContains(t, config, "pid_filename /",
		"squid.conf must NOT set a pid_filename path — dvm owns the PID file (#373)")
}

// TestGenerateSquidConfig_CacheDirFormat verifies the complete cache_dir line
// format: "cache_dir ufs <path> <size> 16 64".
func TestGenerateSquidConfig_CacheDirFormat(t *testing.T) {
	cfg := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/var/cache/squid",
		LogDir:          "/var/log/squid",
		PidFile:         "/run/squid.pid",
		CacheSizeMB:     500,
		MaxObjectSizeMB: 50,
		MemoryCacheMB:   128,
	}

	config, err := GenerateSquidConfig(cfg)
	require.NoError(t, err)

	// Verify the full cache_dir directive structure
	assert.Contains(t, config, "cache_dir ufs /var/cache/squid 500 16 64",
		"cache_dir must use format: ufs <path> <size_mb> 16 64 (#363)")
}
