package registry

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// Task 1: GoModuleConfig Validation Tests
// =============================================================================

func TestGoModuleConfig_Validate_ValidConfig(t *testing.T) {
	config := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      3000,
		Storage:   "/tmp/athens",
	}

	err := config.Validate()
	assert.NoError(t, err, "Valid config should pass validation")
}

func TestGoModuleConfig_Validate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"port zero", 0},
		{"port negative", -1},
		{"port too low", 500},
		{"port too high", 70000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GoModuleConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      tt.port,
				Storage:   "/tmp/athens",
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid port should fail validation")
			assert.ErrorIs(t, err, ErrInvalidConfig)
		})
	}
}

func TestGoModuleConfig_Validate_EmptyStorage(t *testing.T) {
	config := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      3000,
		Storage:   "",
	}

	err := config.Validate()
	assert.Error(t, err, "Empty storage path should fail validation")
	assert.ErrorIs(t, err, ErrInvalidConfig)
}

func TestGoModuleConfig_Validate_InvalidLifecycle(t *testing.T) {
	tests := []struct {
		name      string
		lifecycle string
		wantErr   bool
	}{
		{"persistent valid", "persistent", false},
		{"on-demand valid", "on-demand", false},
		{"manual valid", "manual", false},
		{"empty valid", "", false}, // Empty defaults to on-demand
		{"invalid lifecycle", "invalid", true},
		{"random string", "foobar", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := GoModuleConfig{
				Enabled:   true,
				Lifecycle: tt.lifecycle,
				Port:      3000,
				Storage:   "/tmp/athens",
			}

			err := config.Validate()
			if tt.wantErr {
				assert.Error(t, err, "Invalid lifecycle should fail validation")
				assert.ErrorIs(t, err, ErrInvalidConfig)
			} else {
				assert.NoError(t, err, "Valid lifecycle should pass validation")
			}
		})
	}
}

// =============================================================================
// Task 2: Athens Config Generation Tests
// =============================================================================

func TestAthensConfig_GenerateFromGoModuleConfig(t *testing.T) {
	goModConfig := GoModuleConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      3000,
		Storage:   "/tmp/athens",
	}

	athensConfig, err := GenerateAthensConfig(goModConfig)
	require.NoError(t, err, "GenerateAthensConfig should succeed")
	assert.NotEmpty(t, athensConfig, "Generated config should not be empty")

	// Verify config contains expected TOML sections
	assert.Contains(t, athensConfig, "Port", "Config should have Port")
	assert.Contains(t, athensConfig, "[Storage]", "Config should have Storage section")
	assert.Contains(t, athensConfig, "[Storage.Disk]", "Config should have Disk storage config")
}

func TestAthensConfig_CustomPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"default port", 3000},
		{"custom port", 3333},
		{"high port", 65000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goModConfig := GoModuleConfig{
				Enabled: true,
				Port:    tt.port,
				Storage: "/tmp/athens",
			}

			athensConfig, err := GenerateAthensConfig(goModConfig)
			require.NoError(t, err)

			// Verify port in TOML config string
			expectedPort := `Port = "127.0.0.1:` + strconv.Itoa(tt.port) + `"`
			assert.Contains(t, athensConfig, expectedPort, "Config should contain specified port")
		})
	}
}

func TestAthensConfig_CustomStorage(t *testing.T) {
	tests := []struct {
		name        string
		storagePath string
	}{
		{"tmp directory", "/tmp/athens"},
		{"home directory", "~/.devopsmaestro/athens"},
		{"custom path", "/var/lib/athens"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			goModConfig := GoModuleConfig{
				Enabled: true,
				Port:    3000,
				Storage: tt.storagePath,
			}

			athensConfig, err := GenerateAthensConfig(goModConfig)
			require.NoError(t, err)

			// Verify storage path in TOML config
			assert.Contains(t, athensConfig, "Type = \"disk\"", "Storage type should be disk")
			assert.Contains(t, athensConfig, `RootPath = "`+tt.storagePath+`"`, "Config should contain storage path")
		})
	}
}

func TestAthensConfig_UpstreamProxies(t *testing.T) {
	goModConfig := GoModuleConfig{
		Enabled: true,
		Port:    3000,
		Storage: "/tmp/athens",
		Upstreams: []UpstreamProxyConfig{
			{
				Name: "proxy.golang.org",
				URL:  "https://proxy.golang.org",
			},
		},
	}

	athensConfig, err := GenerateAthensConfig(goModConfig)
	require.NoError(t, err)

	// Verify upstreams are configured
	// Athens uses GoEnv GOPROXY for upstreams
	assert.Contains(t, athensConfig, "[GoEnv]", "Config should have GoEnv section")
	assert.Contains(t, athensConfig, "https://proxy.golang.org", "Config should have upstream URL")
}

func TestAthensConfig_ValidTOML(t *testing.T) {
	goModConfig := GoModuleConfig{
		Enabled: true,
		Port:    3000,
		Storage: "/tmp/athens",
	}

	athensConfig, err := GenerateAthensConfig(goModConfig)
	require.NoError(t, err)

	// Verify it's a non-empty TOML string
	assert.NotEmpty(t, athensConfig, "Config should not be empty")
	assert.Contains(t, athensConfig, "Port =", "Config should be valid TOML")
	assert.Contains(t, athensConfig, "[Storage]", "Config should have TOML sections")
}

// =============================================================================
// Task 3: GoModuleConfig YAML Serialization Tests
// =============================================================================

func TestGoModuleConfig_MarshalYAML(t *testing.T) {
	config := GoModuleConfig{
		Enabled:     true,
		Lifecycle:   "persistent",
		Port:        3000,
		Storage:     "/tmp/athens",
		IdleTimeout: 30 * time.Minute,
		Upstreams: []UpstreamProxyConfig{
			{
				Name: "proxy.golang.org",
				URL:  "https://proxy.golang.org",
			},
		},
	}

	yamlBytes, err := yaml.Marshal(config)
	require.NoError(t, err, "Config should marshal to YAML")

	yamlStr := string(yamlBytes)
	assert.Contains(t, yamlStr, "enabled: true")
	assert.Contains(t, yamlStr, "lifecycle: persistent")
	assert.Contains(t, yamlStr, "port: 3000")
	assert.Contains(t, yamlStr, "proxy.golang.org")
}

func TestGoModuleConfig_UnmarshalYAML(t *testing.T) {
	yamlStr := `
enabled: true
lifecycle: persistent
port: 3000
storage: /tmp/athens
idleTimeout: 30m
upstreams:
  - name: proxy.golang.org
    url: https://proxy.golang.org
`

	var config GoModuleConfig
	err := yaml.Unmarshal([]byte(yamlStr), &config)
	require.NoError(t, err, "YAML should unmarshal to config")

	assert.True(t, config.Enabled)
	assert.Equal(t, "persistent", config.Lifecycle)
	assert.Equal(t, 3000, config.Port)
	assert.Equal(t, "/tmp/athens", config.Storage)
	assert.Equal(t, 30*time.Minute, config.IdleTimeout)
	assert.Len(t, config.Upstreams, 1)
	assert.Equal(t, "proxy.golang.org", config.Upstreams[0].Name)
}

// =============================================================================
// Task 4: Default Config Tests
// =============================================================================

func TestDefaultGoModuleConfig(t *testing.T) {
	config := DefaultGoModuleConfig()

	assert.True(t, config.Enabled, "Go module proxy should be enabled by default")
	assert.Equal(t, "on-demand", config.Lifecycle, "Default lifecycle should be on-demand")
	assert.Equal(t, 3000, config.Port, "Default port should be 3000")
	assert.NotEmpty(t, config.Storage, "Default storage should be set")
	assert.NotEmpty(t, config.Upstreams, "Default upstreams should be configured")

	// Verify proxy.golang.org is in default upstreams
	foundProxyGolang := false
	for _, u := range config.Upstreams {
		if u.URL == "https://proxy.golang.org" {
			foundProxyGolang = true
		}
	}
	assert.True(t, foundProxyGolang, "Default upstreams should include proxy.golang.org")
}

// =============================================================================
// Task 5: UpstreamProxyConfig Tests
// =============================================================================

func TestUpstreamProxyConfig_Validate_Valid(t *testing.T) {
	upstream := UpstreamProxyConfig{
		Name: "proxy.golang.org",
		URL:  "https://proxy.golang.org",
	}

	err := upstream.Validate()
	assert.NoError(t, err, "Valid upstream config should pass validation")
}

func TestUpstreamProxyConfig_Validate_EmptyURL(t *testing.T) {
	upstream := UpstreamProxyConfig{
		Name: "proxy",
		URL:  "",
	}

	err := upstream.Validate()
	assert.Error(t, err, "Empty URL should fail validation")
	assert.ErrorIs(t, err, ErrInvalidConfig)
}

func TestUpstreamProxyConfig_Validate_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"not a url", "not-a-url"},
		{"missing protocol", "proxy.golang.org"},
		{"invalid protocol", "ftp://proxy.golang.org"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := UpstreamProxyConfig{
				Name: "proxy",
				URL:  tt.url,
			}

			err := upstream.Validate()
			assert.Error(t, err, "Invalid URL should fail validation")
			assert.ErrorIs(t, err, ErrInvalidConfig)
		})
	}
}

// =============================================================================
// Task 6: GoModuleProxyStatus Tests
// =============================================================================

func TestGoModuleProxyStatus_Fields(t *testing.T) {
	status := GoModuleProxyStatus{
		State:       "running",
		PID:         12345,
		Port:        3000,
		Storage:     "/tmp/athens",
		Version:     "v0.13.0",
		Uptime:      5 * time.Minute,
		ModuleCount: 42,
		DiskUsage:   1024 * 1024 * 100, // 100MB
	}

	assert.Equal(t, "running", status.State)
	assert.Equal(t, 12345, status.PID)
	assert.Equal(t, 3000, status.Port)
	assert.Equal(t, "/tmp/athens", status.Storage)
	assert.Equal(t, "v0.13.0", status.Version)
	assert.Equal(t, 5*time.Minute, status.Uptime)
	assert.Equal(t, 42, status.ModuleCount)
	assert.Equal(t, int64(1024*1024*100), status.DiskUsage)
}
