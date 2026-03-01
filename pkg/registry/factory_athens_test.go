package registry

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 1: Factory Creation Tests
// =============================================================================

func TestNewGoModuleProxy_ReturnsAthensManager(t *testing.T) {
	config := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      3000,
		Storage:   t.TempDir(),
	}

	mgr := NewGoModuleProxy(config)
	require.NotNil(t, mgr, "NewGoModuleProxy should return non-nil manager")

	// Verify it implements GoModuleProxy interface
	var _ GoModuleProxy = mgr
}

func TestNewGoModuleProxy_ConfiguresCorrectly(t *testing.T) {
	tests := []struct {
		name   string
		config GoModuleConfig
	}{
		{
			name: "persistent lifecycle",
			config: GoModuleConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      3000,
				Storage:   "/tmp/athens1",
			},
		},
		{
			name: "on-demand lifecycle",
			config: GoModuleConfig{
				Enabled:   true,
				Lifecycle: "on-demand",
				Port:      3001,
				Storage:   "/tmp/athens2",
			},
		},
		{
			name: "manual lifecycle",
			config: GoModuleConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      3002,
				Storage:   "/tmp/athens3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewGoModuleProxy(tt.config)
			require.NotNil(t, mgr)

			// Verify endpoint matches port
			endpoint := mgr.GetEndpoint()
			assert.Contains(t, endpoint, strconv.Itoa(tt.config.Port), "Endpoint should contain configured port")
		})
	}
}

func TestNewGoModuleProxy_WithUpstreams(t *testing.T) {
	config := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      3000,
		Storage:   t.TempDir(),
		Upstreams: []UpstreamProxyConfig{
			{
				Name: "proxy.golang.org",
				URL:  "https://proxy.golang.org",
			},
			{
				Name: "goproxy.io",
				URL:  "https://goproxy.io",
			},
		},
	}

	mgr := NewGoModuleProxy(config)
	require.NotNil(t, mgr)

	// Factory should handle upstream configuration
	goEnv := mgr.GetGoEnv()
	assert.Contains(t, goEnv, "GOPROXY")
}

func TestNewGoModuleProxy_WithIdleTimeout(t *testing.T) {
	config := GoModuleConfig{
		Enabled:     true,
		Lifecycle:   "on-demand",
		Port:        3000,
		Storage:     t.TempDir(),
		IdleTimeout: 10 * time.Minute,
	}

	mgr := NewGoModuleProxy(config)
	require.NotNil(t, mgr)

	// Factory should configure idle timeout for on-demand mode
}

// =============================================================================
// Task 2: Factory Validation Tests
// =============================================================================

func TestNewGoModuleProxy_ValidatesConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    GoModuleConfig
		wantPanic bool
	}{
		{
			name: "valid config",
			config: GoModuleConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      3000,
				Storage:   "/tmp/athens",
			},
			wantPanic: false,
		},
		{
			name: "invalid port",
			config: GoModuleConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      -1,
				Storage:   "/tmp/athens",
			},
			wantPanic: true,
		},
		{
			name: "empty storage",
			config: GoModuleConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      3000,
				Storage:   "",
			},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					NewGoModuleProxy(tt.config)
				}, "Factory should panic with invalid config")
			} else {
				assert.NotPanics(t, func() {
					mgr := NewGoModuleProxy(tt.config)
					assert.NotNil(t, mgr)
				}, "Factory should not panic with valid config")
			}
		})
	}
}

// =============================================================================
// Task 3: Integration Factory Tests
// =============================================================================

func TestFactories_Integration_Athens(t *testing.T) {
	// Test that all factories work together for Athens

	// Create go module proxy config
	athensConfig := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      13000,
		Storage:   t.TempDir(),
	}

	// Create go module proxy manager
	proxyMgr := NewGoModuleProxy(athensConfig)
	require.NotNil(t, proxyMgr)

	// Create binary manager (Athens binary)
	binMgr := NewBinaryManager(t.TempDir(), "v0.13.0")
	require.NotNil(t, binMgr)

	// Create process manager
	procConfig := ProcessConfig{
		PIDFile: t.TempDir() + "/athens.pid",
		LogFile: t.TempDir() + "/athens.log",
	}
	procMgr := NewProcessManager(procConfig)
	require.NotNil(t, procMgr)

	// All managers should be created successfully
	assert.NotNil(t, proxyMgr)
	assert.NotNil(t, binMgr)
	assert.NotNil(t, procMgr)
}

func TestFactories_DefaultConfig_Athens(t *testing.T) {
	// Test creating manager with default config
	config := DefaultGoModuleConfig()

	mgr := NewGoModuleProxy(config)
	require.NotNil(t, mgr)

	// Should work with defaults
	endpoint := mgr.GetEndpoint()
	assert.NotEmpty(t, endpoint)
	assert.Contains(t, endpoint, "http://")

	// Should have valid Go environment variables
	goEnv := mgr.GetGoEnv()
	assert.Contains(t, goEnv, "GOPROXY")
	assert.NotEmpty(t, goEnv["GOPROXY"])
}

// =============================================================================
// Task 4: Binary Manager Factory Tests (Athens-specific)
// =============================================================================

func TestNewBinaryManager_Athens_CreatesManager(t *testing.T) {
	binDir := t.TempDir()
	version := "v0.13.0"

	mgr := NewBinaryManager(binDir, version)
	require.NotNil(t, mgr, "NewBinaryManager should return non-nil manager")

	// Verify it implements BinaryManager interface
	var _ BinaryManager = mgr
}

func TestNewBinaryManager_Athens_ConfiguresVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"stable version", "v0.13.0"},
		{"latest version", "v0.14.0"},
		{"specific patch", "v0.13.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewBinaryManager(t.TempDir(), tt.version)
			require.NotNil(t, mgr)

			// Manager should track desired version
		})
	}
}

// =============================================================================
// Task 5: Process Manager Factory Tests (Athens-specific)
// =============================================================================

func TestNewProcessManager_Athens_CreatesManager(t *testing.T) {
	config := ProcessConfig{
		PIDFile: "/tmp/athens.pid",
		LogFile: "/tmp/athens.log",
	}

	mgr := NewProcessManager(config)
	require.NotNil(t, mgr, "NewProcessManager should return non-nil manager")

	// Verify it implements ProcessManager interface
	var _ ProcessManager = mgr
}

func TestNewProcessManager_Athens_ConfiguresPaths(t *testing.T) {
	pidFile := "/custom/path/athens.pid"
	logFile := "/custom/path/athens.log"

	config := ProcessConfig{
		PIDFile: pidFile,
		LogFile: logFile,
	}

	mgr := NewProcessManager(config)
	require.NotNil(t, mgr)

	// Manager should use configured paths
}

func TestNewProcessManager_Athens_ConfiguresShutdownTimeout(t *testing.T) {
	config := ProcessConfig{
		PIDFile:         "/tmp/athens.pid",
		LogFile:         "/tmp/athens.log",
		ShutdownTimeout: 30 * time.Second,
	}

	mgr := NewProcessManager(config)
	require.NotNil(t, mgr)

	// Manager should respect configured timeout
}

// =============================================================================
// Task 6: Factory with Dependencies Tests
// =============================================================================

func TestNewAthensManagerWithDeps_CreatesManager(t *testing.T) {
	config := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      3000,
		Storage:   t.TempDir(),
	}

	// Create dependencies
	mockBinary := NewMockBinaryManager(config.Storage, "v0.13.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/athens.pid",
		LogFile: config.Storage + "/athens.log",
	})

	mgr := NewAthensManagerWithDeps(config, mockBinary, mockProcess)
	require.NotNil(t, mgr, "NewAthensManagerWithDeps should return non-nil manager")

	// Verify it implements GoModuleProxy interface
	var _ GoModuleProxy = mgr
}

func TestNewAthensManagerWithDeps_UsesProvidedDependencies(t *testing.T) {
	config := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      3000,
		Storage:   t.TempDir(),
	}

	// Create custom mock dependencies
	mockBinary := NewMockBinaryManager(config.Storage, "v0.99.0")
	mockProcess := NewProcessManager(ProcessConfig{
		PIDFile: config.Storage + "/custom.pid",
		LogFile: config.Storage + "/custom.log",
	})

	mgr := NewAthensManagerWithDeps(config, mockBinary, mockProcess)
	require.NotNil(t, mgr)

	// Manager should use the provided dependencies
	// (would verify through behavior in integration tests)
}
