package registry

import (
	"context"
	"fmt"
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

// setupTestDevpiManager creates a DevpiManager with test configuration.
func setupTestDevpiManager(t *testing.T) PyPIProxy {
	t.Helper()

	config := PyPIProxyConfig{
		Enabled:     true,
		Lifecycle:   "manual",
		Port:        13141, // Use high port for tests
		Storage:     t.TempDir(),
		IdleTimeout: 1 * time.Minute,
		Upstreams: []PyPIUpstreamConfig{
			{
				Name: "pypi",
				URL:  "https://pypi.org/simple",
			},
		},
	}

	// Use mock binary manager for unit tests
	mockBinary := NewMockPipxBinaryManager(config.Storage, "6.2.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/devpi.pid",
		LogFile: config.Storage + "/devpi.log",
	})

	mgr, err := NewDevpiManager(config, mockBinary, mockProcess)
	require.NoError(t, err)
	return mgr
}

// =============================================================================
// Task 1: DevpiManager Start Tests
// =============================================================================

func TestDevpiManager_Start_Success(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary with HTTP server")

	mgr := setupTestDevpiManager(t)
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

func TestDevpiManager_Start_AlreadyRunning(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	mgr := setupTestDevpiManager(t)
	ctx := context.Background()

	// Start first time
	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	// Start second time - should handle gracefully (idempotent)
	err = mgr.Start(ctx)
	assert.NoError(t, err, "Starting already-running proxy should be idempotent")
}

func TestDevpiManager_Start_BinaryNotFound(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	// Create manager with invalid binary path
	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      13142,
		Storage:   t.TempDir(),
	}

	mgr, err := NewDevpiManagerDefault(config)
	require.NoError(t, err)
	ctx := context.Background()

	// Mock BinaryManager to return error
	// (This test will need proper mock injection in implementation)

	err = mgr.Start(ctx)
	assert.Error(t, err, "Start should fail if binary not found")
	assert.Contains(t, err.Error(), "binary", "Error should mention binary")
}

func TestDevpiManager_Start_PortInUse(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	// Start first proxy
	mgr1 := setupTestDevpiManager(t)
	ctx := context.Background()

	err := mgr1.Start(ctx)
	require.NoError(t, err)
	defer mgr1.Stop(ctx)

	// Try to start second proxy on same port
	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      13141, // Same port
		Storage:   t.TempDir(),
	}
	mgr2, err := NewDevpiManagerDefault(config)
	require.NoError(t, err)

	err = mgr2.Start(ctx)
	assert.Error(t, err, "Start should fail if port is in use")
	assert.ErrorIs(t, err, ErrPortInUse, "Should return ErrPortInUse")
	assert.Contains(t, err.Error(), "port", "Error should mention port conflict")
}

// =============================================================================
// Task 2: DevpiManager Stop Tests
// =============================================================================

func TestDevpiManager_Stop_Success(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	mgr := setupTestDevpiManager(t)
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

func TestDevpiManager_Stop_NotRunning(t *testing.T) {
	mgr := setupTestDevpiManager(t)
	ctx := context.Background()

	// Stop when not running - should be idempotent
	err := mgr.Stop(ctx)
	assert.NoError(t, err, "Stopping non-running proxy should be idempotent")
}

func TestDevpiManager_Stop_GracefulShutdown(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	mgr := setupTestDevpiManager(t)
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
// Task 3: DevpiManager Status Tests
// =============================================================================

func TestDevpiManager_Status_Running(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	mgr := setupTestDevpiManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	status, err := mgr.Status(ctx)
	require.NoError(t, err, "Status should succeed when running")

	assert.Equal(t, "running", status.State, "State should be 'running'")
	assert.Greater(t, status.PID, 0, "PID should be set")
	assert.Equal(t, 13141, status.Port, "Port should match config")
	assert.NotEmpty(t, status.Storage, "Storage path should be set")
	assert.NotEmpty(t, status.Version, "Version should be set")
	assert.GreaterOrEqual(t, status.Uptime, time.Duration(0), "Uptime should be non-negative")
}

func TestDevpiManager_Status_Stopped(t *testing.T) {
	mgr := setupTestDevpiManager(t)
	ctx := context.Background()

	status, err := mgr.Status(ctx)
	require.NoError(t, err, "Status should succeed even when stopped")

	assert.Equal(t, "stopped", status.State, "State should be 'stopped'")
	assert.Equal(t, 0, status.PID, "PID should be 0 when stopped")
	assert.Equal(t, 13141, status.Port, "Port should still be in config")
	assert.NotEmpty(t, status.Storage, "Storage path should be in config")
}

func TestDevpiManager_Status_PackageCount(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	mgr := setupTestDevpiManager(t)
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

func TestDevpiManager_EnsureRunning_StartsIfStopped(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	mgr := setupTestDevpiManager(t)
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

func TestDevpiManager_EnsureRunning_NoopIfRunning(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	mgr := setupTestDevpiManager(t)
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

func TestDevpiManager_IsRunning_True(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	mgr := setupTestDevpiManager(t)
	ctx := context.Background()

	err := mgr.Start(ctx)
	require.NoError(t, err)
	defer mgr.Stop(ctx)

	assert.True(t, mgr.IsRunning(ctx), "IsRunning should return true when proxy is running")
}

func TestDevpiManager_IsRunning_False(t *testing.T) {
	mgr := setupTestDevpiManager(t)
	ctx := context.Background()

	assert.False(t, mgr.IsRunning(ctx), "IsRunning should return false when proxy is not running")
}

func TestDevpiManager_IsRunning_AfterCrash(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server binary")

	mgr := setupTestDevpiManager(t)
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

func TestDevpiManager_GetEndpoint_ReturnsCorrectFormat(t *testing.T) {
	tests := []struct {
		name    string
		port    int
		wantFmt string
	}{
		{
			name:    "default port 3141",
			port:    3141,
			wantFmt: "http://localhost:3141",
		},
		{
			name:    "custom port 3142",
			port:    3142,
			wantFmt: "http://localhost:3142",
		},
		{
			name:    "high port",
			port:    13141,
			wantFmt: "http://localhost:13141",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PyPIProxyConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      tt.port,
				Storage:   t.TempDir(),
			}
			mgr, err := NewDevpiManagerDefault(config)
			require.NoError(t, err)

			endpoint := mgr.GetEndpoint()
			assert.Equal(t, tt.wantFmt, endpoint, "Endpoint format should be http://localhost:port")
		})
	}
}

// =============================================================================
// Task 7: GetPipEnv Tests
// =============================================================================

func TestDevpiManager_GetPipEnv_ReturnsCorrectVariables(t *testing.T) {
	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      3141,
		Storage:   t.TempDir(),
	}
	mgr, err := NewDevpiManagerDefault(config)
	require.NoError(t, err)

	pipEnv := mgr.GetPipEnv()

	// Should include PIP_INDEX_URL
	assert.Contains(t, pipEnv, "PIP_INDEX_URL", "Should include PIP_INDEX_URL variable")
	assert.Equal(t, "http://localhost:3141/root/pypi/+simple/", pipEnv["PIP_INDEX_URL"], "PIP_INDEX_URL should point to devpi endpoint")

	// Should include PIP_TRUSTED_HOST for local development
	assert.Contains(t, pipEnv, "PIP_TRUSTED_HOST", "Should include PIP_TRUSTED_HOST for local development")
	assert.Contains(t, pipEnv["PIP_TRUSTED_HOST"], "localhost", "PIP_TRUSTED_HOST should include localhost")
}

func TestDevpiManager_GetPipEnv_DifferentPorts(t *testing.T) {
	tests := []struct {
		name              string
		port              int
		wantPIP_INDEX_URL string
	}{
		{
			name:              "default port",
			port:              3141,
			wantPIP_INDEX_URL: "http://localhost:3141/root/pypi/+simple/",
		},
		{
			name:              "custom port",
			port:              3142,
			wantPIP_INDEX_URL: "http://localhost:3142/root/pypi/+simple/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PyPIProxyConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      tt.port,
				Storage:   t.TempDir(),
			}
			mgr, err := NewDevpiManagerDefault(config)
			require.NoError(t, err)

			pipEnv := mgr.GetPipEnv()

			assert.Equal(t, tt.wantPIP_INDEX_URL, pipEnv["PIP_INDEX_URL"])
		})
	}
}

func TestDevpiManager_GetPipEnv_NotRunning(t *testing.T) {
	// GetPipEnv should return variables even if proxy is not running
	// (user might want to set env vars before starting)

	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      3141,
		Storage:   t.TempDir(),
	}
	mgr, err := NewDevpiManagerDefault(config)
	require.NoError(t, err)

	pipEnv := mgr.GetPipEnv()

	// Should still return valid environment variables
	assert.Contains(t, pipEnv, "PIP_INDEX_URL")
	assert.NotEmpty(t, pipEnv["PIP_INDEX_URL"])
}

// =============================================================================
// Task 8: GetPipConfig Tests
// =============================================================================

func TestDevpiManager_GetPipConfig_ReturnsCorrectFormat(t *testing.T) {
	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      3141,
		Storage:   t.TempDir(),
	}
	mgr, err := NewDevpiManagerDefault(config)
	require.NoError(t, err)

	pipConfig := mgr.GetPipConfig()

	// Should contain [global] section
	assert.Contains(t, pipConfig, "[global]", "Should contain [global] section")

	// Should contain index-url setting
	assert.Contains(t, pipConfig, "index-url", "Should contain index-url setting")
	assert.Contains(t, pipConfig, "http://localhost:3141", "Should reference devpi endpoint")

	// Should contain trusted-host setting
	assert.Contains(t, pipConfig, "trusted-host", "Should contain trusted-host setting")
	assert.Contains(t, pipConfig, "localhost", "Should trust localhost")
}

func TestDevpiManager_GetPipConfig_DifferentPorts(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"default port", 3141},
		{"custom port", 3142},
		{"high port", 13141},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PyPIProxyConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      tt.port,
				Storage:   t.TempDir(),
			}
			mgr, err := NewDevpiManagerDefault(config)
			require.NoError(t, err)

			pipConfig := mgr.GetPipConfig()

			// Should contain the port number
			assert.Contains(t, pipConfig, fmt.Sprintf("%d", tt.port))
		})
	}
}

// =============================================================================
// Task 9: IdleTimeout Tests (On-Demand Mode)
// =============================================================================

func TestDevpiManager_IdleTimeout_OnDemand(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server and time-based behavior")

	config := PyPIProxyConfig{
		Enabled:     true,
		Lifecycle:   "on-demand",
		Port:        13143,
		Storage:     t.TempDir(),
		IdleTimeout: 500 * time.Millisecond, // Short timeout for testing
	}

	mgr, err := NewDevpiManagerDefault(config)
	require.NoError(t, err)
	ctx := context.Background()

	// Start proxy
	err = mgr.Start(ctx)
	require.NoError(t, err)

	// Verify running
	assert.True(t, mgr.IsRunning(ctx))

	// Wait for idle timeout
	time.Sleep(1 * time.Second)

	// Should have stopped automatically
	assert.False(t, mgr.IsRunning(ctx), "Proxy should stop after idle timeout")
}

func TestDevpiManager_IdleTimeout_ResetOnAccess(t *testing.T) {
	t.Skip("Integration test - requires real devpi-server and time-based behavior")

	config := PyPIProxyConfig{
		Enabled:     true,
		Lifecycle:   "on-demand",
		Port:        13144,
		Storage:     t.TempDir(),
		IdleTimeout: 1 * time.Second,
	}

	mgr, err := NewDevpiManagerDefault(config)
	require.NoError(t, err)
	ctx := context.Background()

	err = mgr.Start(ctx)
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

func TestDevpiManager_ImplementsPyPIProxy(t *testing.T) {
	var _ PyPIProxy = (*DevpiManager)(nil)
}

// =============================================================================
// Bug 1: Health check probe when port is in use (TDD RED phase)
//
// Current behaviour: IsPortAvailable(port) == false → immediately return ErrPortInUse.
//
// Desired behaviour: Before returning ErrPortInUse, probe the service's health
// endpoint (http://localhost:{port}/ → 200 or 302). If healthy, adopt the
// running instance and return nil (idempotent). Only return ErrPortInUse when
// the health probe fails.
// =============================================================================

// TestDevpiManager_Start_AdoptsRunningInstance verifies that Start() returns
// nil when the configured port is occupied by a healthy devpi instance (i.e.
// / returns 200). This is the "adopt already-running service" path.
//
// This test FAILS today because Start() returns ErrPortInUse without probing
// the health endpoint first.
// bindAndServeDevpi binds a TCP listener on 0.0.0.0:0 (so that
// IsPortAvailable returns false for the chosen port) and begins serving HTTP
// requests using the provided handler. The bound port is returned.
// The server is shut down when the test ends via t.Cleanup.
func bindAndServeDevpi(t *testing.T, handler http.Handler) int {
	t.Helper()
	ln, err := net.Listen("tcp", "0.0.0.0:0")
	require.NoError(t, err, "bindAndServeDevpi: net.Listen failed")
	port := ln.Addr().(*net.TCPAddr).Port
	srv := &http.Server{Handler: handler}
	go srv.Serve(ln) //nolint:errcheck
	t.Cleanup(func() { srv.Close() })
	return port
}

// TestDevpiManager_Start_AdoptsRunningInstance verifies that Start() returns
// nil when the configured port is occupied by a healthy devpi instance (i.e.
// / returns 200). This is the "adopt already-running service" path.
//
// This test FAILS today because Start() returns ErrPortInUse without probing
// the health endpoint first.
func TestDevpiManager_Start_AdoptsRunningInstance(t *testing.T) {
	// Stand up a fake "devpi" that responds 200 on /.
	// We bind on 0.0.0.0 so that IsPortAvailable() sees the port as taken.
	port := bindAndServeDevpi(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))

	// Build a DevpiManager pointing at that port.
	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      port,
		Storage:   t.TempDir(),
	}
	mockBinary := NewMockPipxBinaryManager(config.Storage, "6.2.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/devpi.pid",
		LogFile: config.Storage + "/devpi.log",
	})
	mgr, err := NewDevpiManager(config, mockBinary, mockProcess)
	require.NoError(t, err)

	ctx := context.Background()

	// BUG: currently returns ErrPortInUse.
	// FIXED: should probe /, see 200, and return nil.
	startErr := mgr.Start(ctx)
	assert.NoError(t, startErr,
		"Start should return nil when the port is occupied by a healthy devpi instance "+
			"(adopt-running-instance path)")
}

// TestDevpiManager_Start_AdoptsRunningInstance_Redirect verifies that Start()
// also adopts a running instance when / returns 302 (devpi redirects to the
// index on the root path).
//
// This test FAILS today because Start() returns ErrPortInUse without probing.
func TestDevpiManager_Start_AdoptsRunningInstance_Redirect(t *testing.T) {
	// Stand up a fake "devpi" that responds 302 on / (typical redirect).
	// We bind on 0.0.0.0 so that IsPortAvailable() sees the port as taken.
	port := bindAndServeDevpi(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/root/pypi/", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	}))

	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      port,
		Storage:   t.TempDir(),
	}
	mockBinary := NewMockPipxBinaryManager(config.Storage, "6.2.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/devpi.pid",
		LogFile: config.Storage + "/devpi.log",
	})
	mgr, err := NewDevpiManager(config, mockBinary, mockProcess)
	require.NoError(t, err)

	ctx := context.Background()

	// BUG: currently returns ErrPortInUse.
	// FIXED: 302 on / is a healthy devpi response — adopt it.
	startErr := mgr.Start(ctx)
	assert.NoError(t, startErr,
		"Start should return nil when the port is occupied by a devpi instance "+
			"that redirects on / (302)")
}

// TestDevpiManager_Start_PortInUse_UnhealthyService verifies that Start() still
// returns ErrPortInUse when the port is occupied by something that does NOT
// respond with a healthy status on /.
func TestDevpiManager_Start_PortInUse_UnhealthyService(t *testing.T) {
	// Stand up a fake service that returns 500 on all paths.
	// We bind on 0.0.0.0 so that IsPortAvailable() sees the port as taken.
	port := bindAndServeDevpi(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      port,
		Storage:   t.TempDir(),
	}
	mockBinary := NewMockPipxBinaryManager(config.Storage, "6.2.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/devpi.pid",
		LogFile: config.Storage + "/devpi.log",
	})
	mgr, err := NewDevpiManager(config, mockBinary, mockProcess)
	require.NoError(t, err)

	ctx := context.Background()

	startErr := mgr.Start(ctx)
	assert.Error(t, startErr, "Start should return an error when port is occupied by an unhealthy service")
	assert.ErrorIs(t, startErr, ErrPortInUse,
		"Start should return ErrPortInUse when health probe fails")
}

// TestDevpiManager_IsRunning_HealthProbeFallback verifies that IsRunning()
// returns true when the PID file is missing but a healthy devpi instance is
// responding on the configured port. This is the "adopted instance" scenario
// where Start() returned nil without writing a PID file.
func TestDevpiManager_IsRunning_HealthProbeFallback(t *testing.T) {
	// Stand up a fake "devpi" that responds 200 on /.
	port := bindAndServeDevpi(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.WriteHeader(http.StatusOK)
			return
		}
		http.NotFound(w, r)
	}))

	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      port,
		Storage:   t.TempDir(), // Empty dir — no PID file
	}
	mockBinary := NewMockPipxBinaryManager(config.Storage, "6.2.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/devpi.pid",
		LogFile: config.Storage + "/devpi.log",
	})
	mgr, err := NewDevpiManager(config, mockBinary, mockProcess)
	require.NoError(t, err)

	ctx := context.Background()

	assert.True(t, mgr.IsRunning(ctx),
		"IsRunning should return true when PID file is missing but health probe succeeds")
}

// TestDevpiManager_IsRunning_HealthProbeFallback_Redirect verifies that
// IsRunning() returns true when devpi responds with 302 (redirect). A
// redirect means the service is alive and healthy.
func TestDevpiManager_IsRunning_HealthProbeFallback_Redirect(t *testing.T) {
	port := bindAndServeDevpi(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Location", "/+login")
			w.WriteHeader(http.StatusFound)
			return
		}
		http.NotFound(w, r)
	}))

	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      port,
		Storage:   t.TempDir(),
	}
	mockBinary := NewMockPipxBinaryManager(config.Storage, "6.2.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/devpi.pid",
		LogFile: config.Storage + "/devpi.log",
	})
	mgr, err := NewDevpiManager(config, mockBinary, mockProcess)
	require.NoError(t, err)

	ctx := context.Background()

	assert.True(t, mgr.IsRunning(ctx),
		"IsRunning should return true when devpi responds 302 (redirect)")
}

// TestDevpiManager_IsRunning_NoPIDNoHealth verifies that IsRunning() returns
// false when both the PID file is missing AND the health probe fails.
func TestDevpiManager_IsRunning_NoPIDNoHealth(t *testing.T) {
	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      19878, // Nothing listening here
		Storage:   t.TempDir(),
	}
	mockBinary := NewMockPipxBinaryManager(config.Storage, "6.2.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/devpi.pid",
		LogFile: config.Storage + "/devpi.log",
	})
	mgr, err := NewDevpiManager(config, mockBinary, mockProcess)
	require.NoError(t, err)

	ctx := context.Background()

	assert.False(t, mgr.IsRunning(ctx),
		"IsRunning should return false when both PID file and health probe fail")
}
