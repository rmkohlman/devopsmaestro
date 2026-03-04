package registry

import (
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// Task 2.1: Zot Config Generation Tests
// =============================================================================

func TestZotConfig_GenerateFromDvmConfig(t *testing.T) {
	dvmConfig := RegistryConfig{
		Enabled:     true,
		Lifecycle:   "persistent",
		Port:        5001,
		Storage:     "/tmp/registry",
		IdleTimeout: 30 * time.Minute,
		Mirrors: []MirrorConfig{
			{
				Name:     "docker-hub",
				URL:      "https://index.docker.io",
				OnDemand: true,
				Prefix:   "docker.io",
			},
		},
	}

	zotConfig, err := GenerateZotConfig(dvmConfig)
	require.NoError(t, err, "GenerateZotConfig should succeed")
	assert.NotNil(t, zotConfig, "Generated config should not be nil")

	// Verify config structure
	assert.Contains(t, zotConfig, "http", "Config should have http section")
	assert.Contains(t, zotConfig, "storage", "Config should have storage section")
	assert.Contains(t, zotConfig, "extensions", "Config should have extensions section")
}

func TestZotConfig_DefaultMirrors(t *testing.T) {
	dvmConfig := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   "/tmp/registry",
		// No mirrors specified - should use defaults
	}

	zotConfig, err := GenerateZotConfig(dvmConfig)
	require.NoError(t, err)

	// Verify default mirrors are included
	configStr := zotConfig["extensions"].(map[string]interface{})
	sync := configStr["sync"].(map[string]interface{})
	registries := sync["registries"].([]interface{})

	assert.NotEmpty(t, registries, "Default mirrors should be configured")

	// Should include docker.io by default
	foundDockerHub := false
	for _, reg := range registries {
		regMap := reg.(map[string]interface{})
		// Zot v2.0+ uses "urls" (array) instead of "url"
		if urls, ok := regMap["urls"].([]string); ok {
			for _, url := range urls {
				if url == "https://index.docker.io" {
					foundDockerHub = true
					break
				}
			}
		}
	}
	assert.True(t, foundDockerHub, "docker.io should be in default mirrors")
}

func TestZotConfig_CustomPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"default port", 5001},
		{"custom port", 5555},
		{"high port", 65000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dvmConfig := RegistryConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      tt.port,
				Storage:   "/tmp/registry",
			}

			zotConfig, err := GenerateZotConfig(dvmConfig)
			require.NoError(t, err)

			// Verify port in config (Zot v2.0+ uses port as string, separate from address)
			http := zotConfig["http"].(map[string]interface{})
			port := http["port"].(string)
			assert.Equal(t, strconv.Itoa(tt.port), port, "Config should have specified port")
		})
	}
}

func TestZotConfig_CustomStorage(t *testing.T) {
	tests := []struct {
		name        string
		storagePath string
	}{
		{"tmp directory", "/tmp/zot"},
		{"home directory", "~/.devopsmaestro/registry"},
		{"custom path", "/var/lib/registry"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dvmConfig := RegistryConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      5001,
				Storage:   tt.storagePath,
			}

			zotConfig, err := GenerateZotConfig(dvmConfig)
			require.NoError(t, err)

			// Verify storage path in config
			storage := zotConfig["storage"].(map[string]interface{})
			rootDir := storage["rootDirectory"].(string)
			assert.Equal(t, tt.storagePath, rootDir, "Storage path should match config")
		})
	}
}

func TestZotConfig_MultipleUpstreams(t *testing.T) {
	dvmConfig := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   "/tmp/registry",
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
			{
				Name:     "quay",
				URL:      "https://quay.io",
				OnDemand: false,
				Prefix:   "quay.io",
			},
		},
	}

	zotConfig, err := GenerateZotConfig(dvmConfig)
	require.NoError(t, err)

	// Verify all upstreams are configured
	extensions := zotConfig["extensions"].(map[string]interface{})
	sync := extensions["sync"].(map[string]interface{})
	registries := sync["registries"].([]interface{})

	assert.Len(t, registries, 3, "Should have 3 configured mirrors")

	// Verify each mirror - Zot v2.0+ uses "urls" (array)
	allUrls := []string{}
	for _, reg := range registries {
		regMap := reg.(map[string]interface{})
		if urls, ok := regMap["urls"].([]string); ok {
			allUrls = append(allUrls, urls...)
		}
	}

	assert.Contains(t, allUrls, "https://index.docker.io")
	assert.Contains(t, allUrls, "https://ghcr.io")
	assert.Contains(t, allUrls, "https://quay.io")
}

func TestZotConfig_ValidJSON(t *testing.T) {
	dvmConfig := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   "/tmp/registry",
		Mirrors: []MirrorConfig{
			{
				Name:     "docker-hub",
				URL:      "https://index.docker.io",
				OnDemand: true,
				Prefix:   "docker.io",
			},
		},
	}

	zotConfig, err := GenerateZotConfig(dvmConfig)
	require.NoError(t, err)

	// Marshal to JSON
	jsonBytes, err := json.Marshal(zotConfig)
	require.NoError(t, err, "Config should be valid JSON")

	// Verify it can be unmarshaled
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonBytes, &parsed)
	require.NoError(t, err, "Config should unmarshal successfully")

	// Verify structure
	assert.Contains(t, parsed, "http")
	assert.Contains(t, parsed, "storage")
	assert.Contains(t, parsed, "extensions")
}

// =============================================================================
// Task 2.2: Zot Config Structure Tests
// =============================================================================

func TestZotConfig_HTTPSection(t *testing.T) {
	dvmConfig := RegistryConfig{
		Enabled: true,
		Port:    5001,
		Storage: "/tmp/registry",
	}

	zotConfig, err := GenerateZotConfig(dvmConfig)
	require.NoError(t, err)

	http := zotConfig["http"].(map[string]interface{})

	// Verify required HTTP fields
	assert.Contains(t, http, "address", "HTTP section should have address")
	assert.Contains(t, http, "port", "HTTP section should have port") // OR address contains port
}

func TestZotConfig_StorageSection(t *testing.T) {
	dvmConfig := RegistryConfig{
		Enabled: true,
		Port:    5001,
		Storage: "/tmp/registry",
	}

	zotConfig, err := GenerateZotConfig(dvmConfig)
	require.NoError(t, err)

	storage := zotConfig["storage"].(map[string]interface{})

	// Verify required storage fields
	assert.Contains(t, storage, "rootDirectory", "Storage should have rootDirectory")
	assert.Equal(t, "/tmp/registry", storage["rootDirectory"])

	// Verify deduplication is enabled (recommended for cache)
	assert.Contains(t, storage, "dedupe", "Storage should have dedupe enabled")
}

func TestZotConfig_ExtensionsSection(t *testing.T) {
	dvmConfig := RegistryConfig{
		Enabled: true,
		Port:    5001,
		Storage: "/tmp/registry",
		Mirrors: []MirrorConfig{
			{Name: "docker-hub", URL: "https://index.docker.io"},
		},
	}

	zotConfig, err := GenerateZotConfig(dvmConfig)
	require.NoError(t, err)

	extensions := zotConfig["extensions"].(map[string]interface{})

	// Verify sync extension is configured
	assert.Contains(t, extensions, "sync", "Extensions should have sync")

	sync := extensions["sync"].(map[string]interface{})
	assert.Contains(t, sync, "registries", "Sync should have registries")
	assert.Contains(t, sync, "enable", "Sync should have enable flag")
}

func TestZotConfig_OnDemandPullthrough(t *testing.T) {
	dvmConfig := RegistryConfig{
		Enabled: true,
		Port:    5001,
		Storage: "/tmp/registry",
		Mirrors: []MirrorConfig{
			{
				Name:     "docker-hub",
				URL:      "https://index.docker.io",
				OnDemand: true,
				Prefix:   "docker.io",
			},
		},
	}

	zotConfig, err := GenerateZotConfig(dvmConfig)
	require.NoError(t, err)

	extensions := zotConfig["extensions"].(map[string]interface{})
	sync := extensions["sync"].(map[string]interface{})
	registries := sync["registries"].([]interface{})

	reg := registries[0].(map[string]interface{})

	// Verify on-demand pull-through is enabled
	assert.True(t, reg["onDemand"].(bool), "Mirror should have onDemand enabled")
}

// =============================================================================
// Task 2.3: Config Validation Tests
// =============================================================================

func TestRegistryConfig_Validate_ValidConfig(t *testing.T) {
	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   "/tmp/registry",
	}

	err := config.Validate()
	assert.NoError(t, err, "Valid config should pass validation")
}

func TestRegistryConfig_Validate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"port too low", 0},
		{"port negative", -1},
		{"port too high", 70000},
		{"port reserved", 80}, // May want to warn on privileged ports
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := RegistryConfig{
				Enabled:   true,
				Lifecycle: "persistent",
				Port:      tt.port,
				Storage:   "/tmp/registry",
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid port should fail validation")
		})
	}
}

func TestRegistryConfig_Validate_EmptyStorage(t *testing.T) {
	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   "",
	}

	err := config.Validate()
	assert.Error(t, err, "Empty storage path should fail validation")
}

func TestRegistryConfig_Validate_InvalidLifecycle(t *testing.T) {
	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "invalid",
		Port:      5001,
		Storage:   "/tmp/registry",
	}

	err := config.Validate()
	assert.Error(t, err, "Invalid lifecycle should fail validation")
}

func TestRegistryConfig_Validate_InvalidMirrorURL(t *testing.T) {
	config := RegistryConfig{
		Enabled:   true,
		Lifecycle: "persistent",
		Port:      5001,
		Storage:   "/tmp/registry",
		Mirrors: []MirrorConfig{
			{
				Name:   "bad-mirror",
				URL:    "not-a-url",
				Prefix: "bad",
			},
		},
	}

	err := config.Validate()
	assert.Error(t, err, "Invalid mirror URL should fail validation")
}

// =============================================================================
// Task 2.4: Config YAML Serialization Tests
// =============================================================================

func TestRegistryConfig_MarshalYAML(t *testing.T) {
	config := RegistryConfig{
		Enabled:     true,
		Lifecycle:   "persistent",
		Port:        5001,
		Storage:     "/tmp/registry",
		IdleTimeout: 30 * time.Minute,
		Mirrors: []MirrorConfig{
			{
				Name:     "docker-hub",
				URL:      "https://index.docker.io",
				OnDemand: true,
				Prefix:   "docker.io",
			},
		},
	}

	yamlBytes, err := yaml.Marshal(config)
	require.NoError(t, err, "Config should marshal to YAML")

	yamlStr := string(yamlBytes)
	assert.Contains(t, yamlStr, "enabled: true")
	assert.Contains(t, yamlStr, "lifecycle: persistent")
	assert.Contains(t, yamlStr, "port: 5001")
	assert.Contains(t, yamlStr, "docker-hub")
}

func TestRegistryConfig_UnmarshalYAML(t *testing.T) {
	yamlStr := `
enabled: true
lifecycle: persistent
port: 5001
storage: /tmp/registry
idleTimeout: 30m
mirrors:
  - name: docker-hub
    url: https://index.docker.io
    onDemand: true
    prefix: docker.io
`

	var config RegistryConfig
	err := yaml.Unmarshal([]byte(yamlStr), &config)
	require.NoError(t, err, "YAML should unmarshal to config")

	assert.True(t, config.Enabled)
	assert.Equal(t, "persistent", config.Lifecycle)
	assert.Equal(t, 5001, config.Port)
	assert.Equal(t, "/tmp/registry", config.Storage)
	assert.Equal(t, 30*time.Minute, config.IdleTimeout)
	assert.Len(t, config.Mirrors, 1)
	assert.Equal(t, "docker-hub", config.Mirrors[0].Name)
}

// =============================================================================
// Task 2.5: Default Config Tests
// =============================================================================

func TestDefaultRegistryConfig(t *testing.T) {
	config := DefaultRegistryConfig()

	assert.True(t, config.Enabled, "Registry should be enabled by default")
	assert.Equal(t, "on-demand", config.Lifecycle, "Default lifecycle should be on-demand")
	assert.Equal(t, 5001, config.Port, "Default port should be 5001")
	assert.NotEmpty(t, config.Storage, "Default storage should be set")
	assert.NotEmpty(t, config.Mirrors, "Default mirrors should be configured")

	// Verify docker.io is in default mirrors
	foundDockerHub := false
	for _, m := range config.Mirrors {
		if m.URL == "https://index.docker.io" {
			foundDockerHub = true
		}
	}
	assert.True(t, foundDockerHub, "Default mirrors should include docker.io")
}
