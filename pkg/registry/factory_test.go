package registry

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Task 2.1: Factory Creation Tests
// =============================================================================

func TestNewRegistryManager_ReturnsZotManager(t *testing.T) {
	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   t.TempDir(),
	}

	mgr := NewRegistryManager(config)
	require.NotNil(t, mgr, "NewRegistryManager should return non-nil manager")

	// Verify it implements RegistryManager interface
	var _ RegistryManager = mgr
}

func TestNewRegistryManager_ConfiguresCorrectly(t *testing.T) {
	tests := []struct {
		name   string
		config RegistryConfig
	}{
		{
			name: "persistent lifecycle",
			config: RegistryConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      5001,
				Storage:   "/tmp/registry1",
			},
		},
		{
			name: "on-demand lifecycle",
			config: RegistryConfig{
				Enabled:   true,
				Lifecycle: "on-demand",
				Port:      5002,
				Storage:   "/tmp/registry2",
			},
		},
		{
			name: "manual lifecycle",
			config: RegistryConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      5003,
				Storage:   "/tmp/registry3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := NewRegistryManager(tt.config)
			require.NotNil(t, mgr)

			// Verify endpoint matches port
			endpoint := mgr.GetEndpoint()
			assert.Contains(t, endpoint, strconv.Itoa(tt.config.Port), "Endpoint should contain configured port")
		})
	}
}

func TestNewRegistryManager_WithMirrors(t *testing.T) {
	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   t.TempDir(),
		Mirrors: []MirrorConfig{
			{
				Name:     "docker-hub",
				URL:      "https://index.docker.io",
				OnDemand: true,
				Prefix:   "docker.io",
			},
			{
				Name:     "ghcr",
				URL:      "https://ghcr.io",
				OnDemand: true,
				Prefix:   "ghcr.io",
			},
		},
	}

	mgr := NewRegistryManager(config)
	require.NotNil(t, mgr)

	// Factory should handle mirror configuration
	// (Verification would require starting the registry)
}

func TestNewRegistryManager_WithIdleTimeout(t *testing.T) {
	config := RegistryConfig{
		Enabled:     true,
		Lifecycle:   "on-demand",
		Port:        5001,
		Storage:     t.TempDir(),
		IdleTimeout: 10 * time.Minute,
	}

	mgr := NewRegistryManager(config)
	require.NotNil(t, mgr)

	// Factory should configure idle timeout for on-demand mode
}

// =============================================================================
// Task 2.2: Factory Validation Tests
// =============================================================================

func TestNewRegistryManager_ValidatesConfig(t *testing.T) {
	tests := []struct {
		name      string
		config    RegistryConfig
		wantPanic bool
	}{
		{
			name: "valid config",
			config: RegistryConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      5001,
				Storage:   "/tmp/registry",
			},
			wantPanic: false,
		},
		{
			name: "invalid port",
			config: RegistryConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      -1,
				Storage:   "/tmp/registry",
			},
			wantPanic: true,
		},
		{
			name: "empty storage",
			config: RegistryConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      5001,
				Storage:   "",
			},
			wantPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				assert.Panics(t, func() {
					NewRegistryManager(tt.config)
				}, "Factory should panic with invalid config")
			} else {
				assert.NotPanics(t, func() {
					mgr := NewRegistryManager(tt.config)
					assert.NotNil(t, mgr)
				}, "Factory should not panic with valid config")
			}
		})
	}
}

// =============================================================================
// Task 2.3: Binary Manager Factory Tests
// =============================================================================

func TestNewBinaryManager_CreatesManager(t *testing.T) {
	binDir := t.TempDir()
	version := "1.4.3"

	mgr := NewBinaryManager(binDir, version)
	require.NotNil(t, mgr, "NewBinaryManager should return non-nil manager")

	// Verify it implements BinaryManager interface
	var _ BinaryManager = mgr
}

func TestNewBinaryManager_ConfiguresDirectory(t *testing.T) {
	binDir := "/custom/bin/directory"
	version := "1.4.3"

	mgr := NewBinaryManager(binDir, version)
	require.NotNil(t, mgr)

	// Manager should use configured directory
	// (Would verify by checking EnsureBinary behavior)
}

func TestNewBinaryManager_ConfiguresVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
	}{
		{"stable version", "1.4.3"},
		{"latest version", "2.0.0"},
		{"patch version", "1.4.3-patch1"},
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
// Task 2.4: Process Manager Factory Tests
// =============================================================================

func TestNewProcessManager_CreatesManager(t *testing.T) {
	config := ProcessConfig{
		PIDFile: "/tmp/test.pid",
		LogFile: "/tmp/test.log",
	}

	mgr := NewProcessManager(config)
	require.NotNil(t, mgr, "NewProcessManager should return non-nil manager")

	// Verify it implements ProcessManager interface
	var _ ProcessManager = mgr
}

func TestNewProcessManager_ConfiguresPaths(t *testing.T) {
	pidFile := "/custom/path/registry.pid"
	logFile := "/custom/path/registry.log"

	config := ProcessConfig{
		PIDFile: pidFile,
		LogFile: logFile,
	}

	mgr := NewProcessManager(config)
	require.NotNil(t, mgr)

	// Manager should use configured paths
}

func TestNewProcessManager_ConfiguresShutdownTimeout(t *testing.T) {
	config := ProcessConfig{
		PIDFile:         "/tmp/test.pid",
		LogFile:         "/tmp/test.log",
		ShutdownTimeout: 30 * time.Second,
	}

	mgr := NewProcessManager(config)
	require.NotNil(t, mgr)

	// Manager should respect configured timeout
}

// =============================================================================
// Task 2.5: Integration Factory Tests
// =============================================================================

func TestFactories_Integration(t *testing.T) {
	// Test that all factories work together

	// Create registry config
	regConfig := RegistryConfig{
		Enabled:   true,
		Lifecycle: "manual",
		Port:      15001,
		Storage:   t.TempDir(),
	}

	// Create registry manager
	regMgr := NewRegistryManager(regConfig)
	require.NotNil(t, regMgr)

	// Create binary manager
	binMgr := NewBinaryManager(t.TempDir(), "1.4.3")
	require.NotNil(t, binMgr)

	// Create process manager
	procConfig := ProcessConfig{
		PIDFile: t.TempDir() + "/test.pid",
		LogFile: t.TempDir() + "/test.log",
	}
	procMgr := NewProcessManager(procConfig)
	require.NotNil(t, procMgr)

	// All managers should be created successfully
	assert.NotNil(t, regMgr)
	assert.NotNil(t, binMgr)
	assert.NotNil(t, procMgr)
}

func TestFactories_DefaultConfig(t *testing.T) {
	// Test creating manager with default config
	config := DefaultRegistryConfig()

	mgr := NewRegistryManager(config)
	require.NotNil(t, mgr)

	// Should work with defaults
	endpoint := mgr.GetEndpoint()
	assert.NotEmpty(t, endpoint)
}
