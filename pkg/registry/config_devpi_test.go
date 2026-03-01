package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// PyPIProxyConfig Validation Tests
// =============================================================================

func TestPyPIProxyConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name   string
		config PyPIProxyConfig
	}{
		{
			name: "valid persistent config",
			config: PyPIProxyConfig{
				Enabled:     true,
				Lifecycle:   "persistent",
				Port:        3141,
				Storage:     "/tmp/devpi",
				IdleTimeout: 30 * time.Minute,
			},
		},
		{
			name: "valid on-demand config",
			config: PyPIProxyConfig{
				Enabled:     true,
				Lifecycle:   "on-demand",
				Port:        3142,
				Storage:     "/var/lib/devpi",
				IdleTimeout: 15 * time.Minute,
			},
		},
		{
			name: "valid manual config",
			config: PyPIProxyConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      3143,
				Storage:   "/opt/devpi",
			},
		},
		{
			name: "valid with upstreams",
			config: PyPIProxyConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      3141,
				Storage:   "/tmp/devpi",
				Upstreams: []PyPIUpstreamConfig{
					{
						Name: "pypi",
						URL:  "https://pypi.org/simple",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			assert.NoError(t, err, "Valid config should pass validation")
		})
	}
}

func TestPyPIProxyConfig_Validate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"port too low", 0},
		{"port negative", -1},
		{"port too high", 70000},
		{"port reserved", 80},
		{"port privileged", 443},
		{"port just below valid range", 1023},
		{"port just above valid range", 65536},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PyPIProxyConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      tt.port,
				Storage:   "/tmp/devpi",
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid port should fail validation")
			assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
		})
	}
}

func TestPyPIProxyConfig_Validate_EmptyStorage(t *testing.T) {
	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      3141,
		Storage:   "",
	}

	err := config.Validate()
	assert.Error(t, err, "Empty storage path should fail validation")
	assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
}

func TestPyPIProxyConfig_Validate_InvalidLifecycle(t *testing.T) {
	tests := []struct {
		name      string
		lifecycle string
	}{
		{"invalid lifecycle", "invalid"},
		{"typo in lifecycle", "persistant"},
		{"empty string", ""},
		{"uppercase", "PERSISTENT"},
		{"mixed case", "On-Demand"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PyPIProxyConfig{
				Enabled:   true,
				Lifecycle: tt.lifecycle,
				Port:      3141,
				Storage:   "/tmp/devpi",
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid lifecycle should fail validation")
			assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
		})
	}
}

func TestPyPIProxyConfig_Validate_InvalidUpstreamURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty URL", ""},
		{"not a URL", "not-a-url"},
		{"missing protocol", "pypi.org/simple"},
		{"ftp protocol", "ftp://pypi.org/simple"},
		{"just domain", "pypi.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := PyPIProxyConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      3141,
				Storage:   "/tmp/devpi",
				Upstreams: []PyPIUpstreamConfig{
					{
						Name: "pypi",
						URL:  tt.url,
					},
				},
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid upstream URL should fail validation")
		})
	}
}

// =============================================================================
// PyPIProxyConfig Defaults Tests
// =============================================================================

func TestDefaultPyPIProxyConfig(t *testing.T) {
	config := DefaultPyPIProxyConfig()

	assert.True(t, config.Enabled, "PyPI proxy should be enabled by default")
	assert.Equal(t, "on-demand", config.Lifecycle, "Default lifecycle should be on-demand")
	assert.Equal(t, 3141, config.Port, "Default port should be 3141")
	assert.NotEmpty(t, config.Storage, "Default storage should be set")
	assert.Contains(t, config.Storage, ".devopsmaestro", "Storage should be in .devopsmaestro directory")
	assert.Contains(t, config.Storage, "devpi", "Storage should contain 'devpi'")
	assert.Equal(t, 30*time.Minute, config.IdleTimeout, "Default idle timeout should be 30 minutes")
	assert.NotEmpty(t, config.Upstreams, "Default upstreams should be configured")
}

func TestPyPIUpstreamConfig_DefaultUpstreams(t *testing.T) {
	upstreams := defaultPyPIUpstreams()

	assert.NotEmpty(t, upstreams, "Default upstreams should not be empty")
	assert.Len(t, upstreams, 1, "Should have 1 default upstream")

	// Verify PyPI is in default upstreams
	foundPyPI := false
	for _, u := range upstreams {
		if u.URL == "https://pypi.org/simple" {
			foundPyPI = true
			assert.Equal(t, "pypi", u.Name, "PyPI upstream should be named 'pypi'")
		}
	}
	assert.True(t, foundPyPI, "Default upstreams should include PyPI")
}

func TestPyPIProxyConfig_Defaults_CustomValues(t *testing.T) {
	config := DefaultPyPIProxyConfig()

	// Modify values
	config.Port = 4141
	config.Lifecycle = "persistent"

	// Validate still works
	err := config.Validate()
	assert.NoError(t, err, "Modified default config should still be valid")
}

// =============================================================================
// PyPIProxyConfig YAML Serialization Tests
// =============================================================================

func TestPyPIProxyConfig_MarshalYAML(t *testing.T) {
	config := PyPIProxyConfig{
		Enabled:     true,
		Lifecycle:   "persistent",
		Port:        3141,
		Storage:     "/tmp/devpi",
		IdleTimeout: 30 * time.Minute,
		ServerDir:   "/tmp/devpi/server",
		Upstreams: []PyPIUpstreamConfig{
			{
				Name: "pypi",
				URL:  "https://pypi.org/simple",
			},
		},
	}

	yamlBytes, err := yaml.Marshal(config)
	require.NoError(t, err, "Config should marshal to YAML")

	yamlStr := string(yamlBytes)
	assert.Contains(t, yamlStr, "enabled: true")
	assert.Contains(t, yamlStr, "lifecycle: persistent")
	assert.Contains(t, yamlStr, "port: 3141")
	assert.Contains(t, yamlStr, "pypi")
}

func TestPyPIProxyConfig_UnmarshalYAML(t *testing.T) {
	yamlStr := `
enabled: true
lifecycle: persistent
port: 3141
storage: /tmp/devpi
idleTimeout: 30m
serverDir: /tmp/devpi/server
upstreams:
  - name: pypi
    url: https://pypi.org/simple
`

	var config PyPIProxyConfig
	err := yaml.Unmarshal([]byte(yamlStr), &config)
	require.NoError(t, err, "YAML should unmarshal to config")

	assert.True(t, config.Enabled)
	assert.Equal(t, "persistent", config.Lifecycle)
	assert.Equal(t, 3141, config.Port)
	assert.Equal(t, "/tmp/devpi", config.Storage)
	assert.Equal(t, 30*time.Minute, config.IdleTimeout)
	assert.Equal(t, "/tmp/devpi/server", config.ServerDir)
	assert.Len(t, config.Upstreams, 1)
	assert.Equal(t, "pypi", config.Upstreams[0].Name)
}

// =============================================================================
// PyPIUpstreamConfig Validation Tests
// =============================================================================

func TestPyPIUpstreamConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name     string
		upstream PyPIUpstreamConfig
	}{
		{
			name: "valid PyPI URL",
			upstream: PyPIUpstreamConfig{
				Name: "pypi",
				URL:  "https://pypi.org/simple",
			},
		},
		{
			name: "valid custom index",
			upstream: PyPIUpstreamConfig{
				Name: "custom",
				URL:  "https://custom.pypi.org/simple",
			},
		},
		{
			name: "valid with http",
			upstream: PyPIUpstreamConfig{
				Name: "local",
				URL:  "http://localhost:3141/simple",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.upstream.Validate()
			assert.NoError(t, err, "Valid upstream config should pass validation")
		})
	}
}

func TestPyPIUpstreamConfig_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		upstream PyPIUpstreamConfig
		wantErr  bool
	}{
		{
			name: "empty URL",
			upstream: PyPIUpstreamConfig{
				Name: "pypi",
				URL:  "",
			},
			wantErr: true,
		},
		{
			name: "invalid URL format",
			upstream: PyPIUpstreamConfig{
				Name: "pypi",
				URL:  "not-a-url",
			},
			wantErr: true,
		},
		{
			name: "missing protocol",
			upstream: PyPIUpstreamConfig{
				Name: "pypi",
				URL:  "pypi.org/simple",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.upstream.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Invalid upstream should fail validation")
			} else {
				assert.NoError(t, err, "Valid upstream should pass validation")
			}
		})
	}
}

// =============================================================================
// GenerateDevpiConfig Tests
// =============================================================================

func TestGenerateDevpiConfig_Basic(t *testing.T) {
	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      3141,
		Storage:   "/tmp/devpi",
	}

	devpiConfig, err := GenerateDevpiConfig(config)
	require.NoError(t, err, "Should generate config successfully")
	assert.NotEmpty(t, devpiConfig, "Generated config should not be empty")

	// Config should contain key settings
	assert.Contains(t, devpiConfig, "3141", "Config should contain port")
	assert.Contains(t, devpiConfig, "/tmp/devpi", "Config should contain storage path")
}

func TestGenerateDevpiConfig_WithUpstreams(t *testing.T) {
	config := PyPIProxyConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      3141,
		Storage:   "/tmp/devpi",
		Upstreams: []PyPIUpstreamConfig{
			{
				Name: "pypi",
				URL:  "https://pypi.org/simple",
			},
		},
	}

	devpiConfig, err := GenerateDevpiConfig(config)
	require.NoError(t, err, "Should generate config with upstreams")
	assert.Contains(t, devpiConfig, "pypi.org", "Config should reference upstream URL")
}

func TestGenerateDevpiConfig_InvalidConfig(t *testing.T) {
	config := PyPIProxyConfig{
		Enabled: true,
		Port:    0, // Invalid port
		Storage: "/tmp/devpi",
	}

	_, err := GenerateDevpiConfig(config)
	assert.Error(t, err, "Should fail with invalid config")
	assert.Contains(t, err.Error(), "invalid", "Error should mention validation failure")
}
