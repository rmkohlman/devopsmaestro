package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// =============================================================================
// HttpProxyConfig Validation Tests
// =============================================================================

func TestHttpProxyConfig_Validate_Valid(t *testing.T) {
	tests := []struct {
		name   string
		config HttpProxyConfig
	}{
		{
			name: "valid default config",
			config: HttpProxyConfig{
				Port:            3128,
				CacheDir:        "/tmp/squid/cache",
				LogDir:          "/tmp/squid/logs",
				PidFile:         "/tmp/squid/squid.pid",
				CacheSizeMB:     1000,
				MaxObjectSizeMB: 100,
				MemoryCacheMB:   256,
			},
		},
		{
			name: "valid with custom port",
			config: HttpProxyConfig{
				Port:            8080,
				CacheDir:        "/var/cache/squid",
				LogDir:          "/var/log/squid",
				PidFile:         "/var/run/squid.pid",
				CacheSizeMB:     5000,
				MaxObjectSizeMB: 500,
				MemoryCacheMB:   512,
			},
		},
		{
			name: "valid with minimum values",
			config: HttpProxyConfig{
				Port:            1024,
				CacheDir:        "/tmp/squid",
				LogDir:          "/tmp/logs",
				PidFile:         "/tmp/squid.pid",
				CacheSizeMB:     100,
				MaxObjectSizeMB: 10,
				MemoryCacheMB:   64,
			},
		},
		{
			name: "valid with large cache",
			config: HttpProxyConfig{
				Port:            3128,
				CacheDir:        "/var/cache/squid",
				LogDir:          "/var/log/squid",
				PidFile:         "/var/run/squid.pid",
				CacheSizeMB:     50000,
				MaxObjectSizeMB: 5000,
				MemoryCacheMB:   2048,
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

func TestHttpProxyConfig_Validate_InvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"port zero", 0},
		{"port negative", -1},
		{"port too high", 70000},
		{"port reserved", 80},
		{"port privileged", 443},
		{"port just below valid range", 1023},
		{"port just above valid range", 65536},
		{"port extremely negative", -9999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HttpProxyConfig{
				Port:            tt.port,
				CacheDir:        "/tmp/squid/cache",
				LogDir:          "/tmp/squid/logs",
				PidFile:         "/tmp/squid/squid.pid",
				CacheSizeMB:     1000,
				MaxObjectSizeMB: 100,
				MemoryCacheMB:   256,
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid port should fail validation")
			assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
		})
	}
}

func TestHttpProxyConfig_Validate_EmptyCacheDir(t *testing.T) {
	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	err := config.Validate()
	assert.Error(t, err, "Empty CacheDir should fail validation")
	assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
}

func TestHttpProxyConfig_Validate_EmptyLogDir(t *testing.T) {
	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	err := config.Validate()
	assert.Error(t, err, "Empty LogDir should fail validation")
	assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
}

func TestHttpProxyConfig_Validate_EmptyPidFile(t *testing.T) {
	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	err := config.Validate()
	assert.Error(t, err, "Empty PidFile should fail validation")
	assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
}

func TestHttpProxyConfig_Validate_InvalidCacheSizes(t *testing.T) {
	tests := []struct {
		name            string
		cacheSizeMB     int
		maxObjectSizeMB int
		memoryCacheMB   int
	}{
		{
			name:            "zero cache size",
			cacheSizeMB:     0,
			maxObjectSizeMB: 100,
			memoryCacheMB:   256,
		},
		{
			name:            "negative cache size",
			cacheSizeMB:     -1,
			maxObjectSizeMB: 100,
			memoryCacheMB:   256,
		},
		{
			name:            "zero max object size",
			cacheSizeMB:     1000,
			maxObjectSizeMB: 0,
			memoryCacheMB:   256,
		},
		{
			name:            "negative max object size",
			cacheSizeMB:     1000,
			maxObjectSizeMB: -1,
			memoryCacheMB:   256,
		},
		{
			name:            "zero memory cache",
			cacheSizeMB:     1000,
			maxObjectSizeMB: 100,
			memoryCacheMB:   0,
		},
		{
			name:            "negative memory cache",
			cacheSizeMB:     1000,
			maxObjectSizeMB: 100,
			memoryCacheMB:   -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := HttpProxyConfig{
				Port:            3128,
				CacheDir:        "/tmp/squid/cache",
				LogDir:          "/tmp/squid/logs",
				PidFile:         "/tmp/squid/squid.pid",
				CacheSizeMB:     tt.cacheSizeMB,
				MaxObjectSizeMB: tt.maxObjectSizeMB,
				MemoryCacheMB:   tt.memoryCacheMB,
			}

			err := config.Validate()
			assert.Error(t, err, "Invalid cache size should fail validation")
			assert.ErrorIs(t, err, ErrInvalidConfig, "Should return ErrInvalidConfig")
		})
	}
}

// =============================================================================
// HttpProxyConfig Defaults Tests
// =============================================================================

func TestDefaultHttpProxyConfig(t *testing.T) {
	config := DefaultHttpProxyConfig()

	assert.Equal(t, 3128, config.Port, "Default port should be 3128")
	assert.NotEmpty(t, config.CacheDir, "Default CacheDir should be set")
	assert.NotEmpty(t, config.LogDir, "Default LogDir should be set")
	assert.NotEmpty(t, config.PidFile, "Default PidFile should be set")
	assert.Equal(t, 1000, config.CacheSizeMB, "Default cache size should be 1000MB")
	assert.Equal(t, 100, config.MaxObjectSizeMB, "Default max object size should be 100MB")
	assert.Equal(t, 256, config.MemoryCacheMB, "Default memory cache should be 256MB")

	// Verify paths contain .devopsmaestro
	assert.Contains(t, config.CacheDir, ".devopsmaestro", "CacheDir should be in .devopsmaestro directory")
	assert.Contains(t, config.LogDir, ".devopsmaestro", "LogDir should be in .devopsmaestro directory")
	assert.Contains(t, config.PidFile, ".devopsmaestro", "PidFile should be in .devopsmaestro directory")

	// Verify paths contain squid
	assert.Contains(t, config.CacheDir, "squid", "CacheDir should contain 'squid'")
	assert.Contains(t, config.LogDir, "squid", "LogDir should contain 'squid'")
}

func TestHttpProxyConfig_Defaults_CustomValues(t *testing.T) {
	config := DefaultHttpProxyConfig()

	// Modify values
	config.Port = 8080
	config.CacheSizeMB = 5000

	// Validate still works
	err := config.Validate()
	assert.NoError(t, err, "Modified default config should still be valid")
}

// =============================================================================
// HttpProxyConfig YAML Serialization Tests
// =============================================================================

func TestHttpProxyConfig_MarshalYAML(t *testing.T) {
	config := HttpProxyConfig{
		Port:            3128,
		CacheDir:        "/tmp/squid/cache",
		LogDir:          "/tmp/squid/logs",
		PidFile:         "/tmp/squid/squid.pid",
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}

	yamlBytes, err := yaml.Marshal(config)
	require.NoError(t, err, "Config should marshal to YAML")

	yamlStr := string(yamlBytes)
	assert.Contains(t, yamlStr, "port: 3128")
	assert.Contains(t, yamlStr, "cacheDir: /tmp/squid/cache")
	assert.Contains(t, yamlStr, "logDir: /tmp/squid/logs")
	assert.Contains(t, yamlStr, "cacheSizeMB: 1000")
	assert.Contains(t, yamlStr, "maxObjectSizeMB: 100")
	assert.Contains(t, yamlStr, "memoryCacheMB: 256")
}

func TestHttpProxyConfig_UnmarshalYAML(t *testing.T) {
	yamlStr := `
port: 3128
cacheDir: /tmp/squid/cache
logDir: /tmp/squid/logs
pidFile: /tmp/squid/squid.pid
cacheSizeMB: 1000
maxObjectSizeMB: 100
memoryCacheMB: 256
`

	var config HttpProxyConfig
	err := yaml.Unmarshal([]byte(yamlStr), &config)
	require.NoError(t, err, "YAML should unmarshal to config")

	assert.Equal(t, 3128, config.Port)
	assert.Equal(t, "/tmp/squid/cache", config.CacheDir)
	assert.Equal(t, "/tmp/squid/logs", config.LogDir)
	assert.Equal(t, "/tmp/squid/squid.pid", config.PidFile)
	assert.Equal(t, 1000, config.CacheSizeMB)
	assert.Equal(t, 100, config.MaxObjectSizeMB)
	assert.Equal(t, 256, config.MemoryCacheMB)
}

// =============================================================================
// HttpProxyStatus Tests
// =============================================================================

func TestHttpProxyStatus_Structure(t *testing.T) {
	// Verify HttpProxyStatus has expected fields
	status := HttpProxyStatus{
		State:    "running",
		PID:      12345,
		Port:     3128,
		CacheDir: "/tmp/squid/cache",
		Version:  "6.0",
	}

	assert.Equal(t, "running", status.State)
	assert.Equal(t, 12345, status.PID)
	assert.Equal(t, 3128, status.Port)
	assert.Equal(t, "/tmp/squid/cache", status.CacheDir)
	assert.Equal(t, "6.0", status.Version)
}
