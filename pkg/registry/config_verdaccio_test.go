package registry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// NpmProxyConfig Validation Tests
// =============================================================================

func TestNpmProxyConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name   string
		config NpmProxyConfig
	}{
		{
			name: "valid persistent config",
			config: NpmProxyConfig{
				Enabled:     true,
				Lifecycle:   "persistent",
				Port:        4873,
				Storage:     "/tmp/verdaccio",
				IdleTimeout: 30 * time.Minute,
			},
		},
		{
			name: "valid on-demand config",
			config: NpmProxyConfig{
				Enabled:     true,
				Lifecycle:   "on-demand",
				Port:        4874,
				Storage:     "/var/lib/verdaccio",
				IdleTimeout: 15 * time.Minute,
			},
		},
		{
			name: "valid manual config",
			config: NpmProxyConfig{
				Enabled:   true,
				Lifecycle: "manual",
				Port:      4875,
				Storage:   "/opt/verdaccio",
			},
		},
		{
			name: "valid with upstreams",
			config: NpmProxyConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      4873,
				Storage:   "/tmp/verdaccio",
				Upstreams: []NpmUpstreamConfig{
					{
						Name: "npmjs",
						URL:  "https://registry.npmjs.org",
					},
				},
			},
		},
		{
			name: "valid with auth",
			config: NpmProxyConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      4873,
				Storage:   "/tmp/verdaccio",
				Auth: NpmAuthConfig{
					Enabled: true,
					Type:    "htpasswd",
				},
			},
		},
		{
			name: "valid with max body size",
			config: NpmProxyConfig{
				Enabled:     true,
				Lifecycle:   "persistent",
				Port:        4873,
				Storage:     "/tmp/verdaccio",
				MaxBodySize: "10mb",
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

func TestNpmProxyConfig_Validate_InvalidPort(t *testing.T) {
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
			config := NpmProxyConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      tt.port,
				Storage:   "/tmp/verdaccio",
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid port should fail validation")
			assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
		})
	}
}

func TestNpmProxyConfig_Validate_EmptyStorage(t *testing.T) {
	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      4873,
		Storage:   "",
	}

	err := config.Validate()
	assert.Error(t, err, "Empty storage path should fail validation")
	assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
}

func TestNpmProxyConfig_Validate_InvalidLifecycle(t *testing.T) {
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
			config := NpmProxyConfig{
				Enabled:   true,
				Lifecycle: tt.lifecycle,
				Port:      4873,
				Storage:   "/tmp/verdaccio",
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid lifecycle should fail validation")
			assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
		})
	}
}

func TestNpmProxyConfig_Validate_InvalidUpstreamURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty URL", ""},
		{"not a URL", "not-a-url"},
		{"missing protocol", "registry.npmjs.org"},
		{"ftp protocol", "ftp://registry.npmjs.org"},
		{"just domain", "npmjs.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NpmProxyConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      4873,
				Storage:   "/tmp/verdaccio",
				Upstreams: []NpmUpstreamConfig{
					{
						Name: "npmjs",
						URL:  tt.url,
					},
				},
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid upstream URL should fail validation")
		})
	}
}

func TestNpmProxyConfig_Validate_InvalidAuthType(t *testing.T) {
	tests := []struct {
		name     string
		authType string
	}{
		{"unknown auth type", "unknown"},
		{"typo in auth type", "htpaswd"},
		{"empty string", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := NpmProxyConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      4873,
				Storage:   "/tmp/verdaccio",
				Auth: NpmAuthConfig{
					Enabled: true,
					Type:    tt.authType,
				},
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid auth type should fail validation")
		})
	}
}

// =============================================================================
// NpmProxyConfig Defaults Tests
// =============================================================================

func TestDefaultNpmProxyConfig(t *testing.T) {
	config := DefaultNpmProxyConfig()

	assert.True(t, config.Enabled, "npm proxy should be enabled by default")
	assert.Equal(t, "on-demand", config.Lifecycle, "Default lifecycle should be on-demand")
	assert.Equal(t, 4873, config.Port, "Default port should be 4873")
	assert.NotEmpty(t, config.Storage, "Default storage should be set")
	assert.Contains(t, config.Storage, ".devopsmaestro", "Storage should be in .devopsmaestro directory")
	assert.Contains(t, config.Storage, "verdaccio", "Storage should contain 'verdaccio'")
	assert.Equal(t, 30*time.Minute, config.IdleTimeout, "Default idle timeout should be 30 minutes")
	assert.NotEmpty(t, config.Upstreams, "Default upstreams should be configured")
	assert.Equal(t, "10mb", config.MaxBodySize, "Default max body size should be 10mb")
}

func TestNpmUpstreamConfig_DefaultUpstreams(t *testing.T) {
	upstreams := defaultNpmUpstreams()

	assert.NotEmpty(t, upstreams, "Default upstreams should not be empty")
	assert.Len(t, upstreams, 1, "Should have 1 default upstream")

	// Verify npmjs is in default upstreams
	foundNpmjs := false
	for _, u := range upstreams {
		if u.URL == "https://registry.npmjs.org" {
			foundNpmjs = true
			assert.Equal(t, "npmjs", u.Name, "npmjs upstream should be named 'npmjs'")
		}
	}
	assert.True(t, foundNpmjs, "Default upstreams should include npmjs")
}

func TestNpmProxyConfig_Defaults_CustomValues(t *testing.T) {
	config := DefaultNpmProxyConfig()

	// Modify values
	config.Port = 5873
	config.Lifecycle = "persistent"

	// Validate still works
	err := config.Validate()
	assert.NoError(t, err, "Modified default config should still be valid")
}

// =============================================================================
// NpmProxyConfig YAML Serialization Tests
// =============================================================================

func TestNpmProxyConfig_MarshalYAML(t *testing.T) {
	config := NpmProxyConfig{
		Enabled:     true,
		Lifecycle:   "persistent",
		Port:        4873,
		Storage:     "/tmp/verdaccio",
		ConfigFile:  "/tmp/verdaccio/config.yaml",
		IdleTimeout: 30 * time.Minute,
		MaxBodySize: "10mb",
		Upstreams: []NpmUpstreamConfig{
			{
				Name: "npmjs",
				URL:  "https://registry.npmjs.org",
			},
		},
		Auth: NpmAuthConfig{
			Enabled: false,
		},
	}

	yamlBytes, err := yaml.Marshal(config)
	require.NoError(t, err, "Config should marshal to YAML")

	yamlStr := string(yamlBytes)
	assert.Contains(t, yamlStr, "enabled: true")
	assert.Contains(t, yamlStr, "lifecycle: persistent")
	assert.Contains(t, yamlStr, "port: 4873")
	assert.Contains(t, yamlStr, "npmjs")
}

func TestNpmProxyConfig_UnmarshalYAML(t *testing.T) {
	yamlStr := `
enabled: true
lifecycle: persistent
port: 4873
storage: /tmp/verdaccio
configFile: /tmp/verdaccio/config.yaml
idleTimeout: 30m
maxBodySize: 10mb
upstreams:
  - name: npmjs
    url: https://registry.npmjs.org
auth:
  enabled: false
`

	var config NpmProxyConfig
	err := yaml.Unmarshal([]byte(yamlStr), &config)
	require.NoError(t, err, "YAML should unmarshal to config")

	assert.True(t, config.Enabled)
	assert.Equal(t, "persistent", config.Lifecycle)
	assert.Equal(t, 4873, config.Port)
	assert.Equal(t, "/tmp/verdaccio", config.Storage)
	assert.Equal(t, "/tmp/verdaccio/config.yaml", config.ConfigFile)
	assert.Equal(t, 30*time.Minute, config.IdleTimeout)
	assert.Equal(t, "10mb", config.MaxBodySize)
	assert.Len(t, config.Upstreams, 1)
	assert.Equal(t, "npmjs", config.Upstreams[0].Name)
	assert.False(t, config.Auth.Enabled)
}

// =============================================================================
// NpmUpstreamConfig Validation Tests
// =============================================================================

func TestNpmUpstreamConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name     string
		upstream NpmUpstreamConfig
	}{
		{
			name: "valid npmjs URL",
			upstream: NpmUpstreamConfig{
				Name: "npmjs",
				URL:  "https://registry.npmjs.org",
			},
		},
		{
			name: "valid custom registry",
			upstream: NpmUpstreamConfig{
				Name: "custom",
				URL:  "https://custom.registry.com",
			},
		},
		{
			name: "valid with http",
			upstream: NpmUpstreamConfig{
				Name: "local",
				URL:  "http://localhost:4873",
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

func TestNpmUpstreamConfig_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name     string
		upstream NpmUpstreamConfig
		wantErr  bool
	}{
		{
			name: "empty URL",
			upstream: NpmUpstreamConfig{
				Name: "npmjs",
				URL:  "",
			},
			wantErr: true,
		},
		{
			name: "invalid URL format",
			upstream: NpmUpstreamConfig{
				Name: "npmjs",
				URL:  "not-a-url",
			},
			wantErr: true,
		},
		{
			name: "missing protocol",
			upstream: NpmUpstreamConfig{
				Name: "npmjs",
				URL:  "registry.npmjs.org",
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
// NpmAuthConfig Validation Tests
// =============================================================================

func TestNpmAuthConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name string
		auth NpmAuthConfig
	}{
		{
			name: "auth disabled",
			auth: NpmAuthConfig{
				Enabled: false,
			},
		},
		{
			name: "htpasswd auth",
			auth: NpmAuthConfig{
				Enabled: true,
				Type:    "htpasswd",
			},
		},
		{
			name: "ldap auth",
			auth: NpmAuthConfig{
				Enabled: true,
				Type:    "ldap",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.auth.Validate()
			assert.NoError(t, err, "Valid auth config should pass validation")
		})
	}
}

func TestNpmAuthConfig_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		auth    NpmAuthConfig
		wantErr bool
	}{
		{
			name: "enabled with no type",
			auth: NpmAuthConfig{
				Enabled: true,
				Type:    "",
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			auth: NpmAuthConfig{
				Enabled: true,
				Type:    "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.auth.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Invalid auth should fail validation")
			} else {
				assert.NoError(t, err, "Valid auth should pass validation")
			}
		})
	}
}

// =============================================================================
// GenerateVerdaccioConfig Tests
// =============================================================================

func TestGenerateVerdaccioConfig_Basic(t *testing.T) {
	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      4873,
		Storage:   "/tmp/verdaccio",
	}

	verdaccioConfig, err := GenerateVerdaccioConfig(config)
	require.NoError(t, err, "Should generate config successfully")
	assert.NotEmpty(t, verdaccioConfig, "Generated config should not be empty")

	// Config should contain key settings
	assert.Contains(t, verdaccioConfig, "4873", "Config should contain port")
	assert.Contains(t, verdaccioConfig, "/tmp/verdaccio", "Config should contain storage path")
	assert.Contains(t, verdaccioConfig, "storage:", "Config should have storage section")
}

func TestGenerateVerdaccioConfig_WithUpstreams(t *testing.T) {
	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      4873,
		Storage:   "/tmp/verdaccio",
		Upstreams: []NpmUpstreamConfig{
			{
				Name: "npmjs",
				URL:  "https://registry.npmjs.org",
			},
		},
	}

	verdaccioConfig, err := GenerateVerdaccioConfig(config)
	require.NoError(t, err, "Should generate config with upstreams")
	assert.Contains(t, verdaccioConfig, "registry.npmjs.org", "Config should reference upstream URL")
	assert.Contains(t, verdaccioConfig, "uplinks:", "Config should have uplinks section")
}

func TestGenerateVerdaccioConfig_WithAuth(t *testing.T) {
	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      4873,
		Storage:   "/tmp/verdaccio",
		Auth: NpmAuthConfig{
			Enabled: true,
			Type:    "htpasswd",
		},
	}

	verdaccioConfig, err := GenerateVerdaccioConfig(config)
	require.NoError(t, err, "Should generate config with auth")
	assert.Contains(t, verdaccioConfig, "auth:", "Config should have auth section")
	assert.Contains(t, verdaccioConfig, "htpasswd", "Config should reference htpasswd plugin")
}

func TestGenerateVerdaccioConfig_WithMaxBodySize(t *testing.T) {
	config := NpmProxyConfig{
		Enabled:     true,
		Lifecycle:   "persistent",
		Port:        4873,
		Storage:     "/tmp/verdaccio",
		MaxBodySize: "50mb",
	}

	verdaccioConfig, err := GenerateVerdaccioConfig(config)
	require.NoError(t, err, "Should generate config with max body size")
	assert.Contains(t, verdaccioConfig, "max_body_size:", "Config should have max_body_size setting")
	assert.Contains(t, verdaccioConfig, "50mb", "Config should contain max body size value")
}

func TestGenerateVerdaccioConfig_InvalidConfig(t *testing.T) {
	config := NpmProxyConfig{
		Enabled: true,
		Port:    0, // Invalid port
		Storage: "/tmp/verdaccio",
	}

	_, err := GenerateVerdaccioConfig(config)
	assert.Error(t, err, "Should fail with invalid config")
	assert.Contains(t, err.Error(), "invalid", "Error should mention validation failure")
}

func TestGenerateVerdaccioConfig_YAMLFormat(t *testing.T) {
	config := NpmProxyConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      4873,
		Storage:   "/tmp/verdaccio",
	}

	verdaccioConfig, err := GenerateVerdaccioConfig(config)
	require.NoError(t, err)

	// Should be valid YAML
	var parsed map[string]interface{}
	err = yaml.Unmarshal([]byte(verdaccioConfig), &parsed)
	assert.NoError(t, err, "Generated config should be valid YAML")
	assert.NotEmpty(t, parsed, "Parsed config should have content")
}
