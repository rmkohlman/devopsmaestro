package registry

import (
	"path/filepath"
	"time"

	"github.com/rmkohlman/MaestroSDK/paths"
)

// HttpProxyConfig represents squid HTTP proxy configuration.
type HttpProxyConfig struct {
	// Port is the port the proxy listens on
	Port int `json:"port" yaml:"port"`

	// CacheDir is the directory where cached objects are stored
	CacheDir string `json:"cacheDir" yaml:"cacheDir"`

	// LogDir is the directory where log files are stored
	LogDir string `json:"logDir" yaml:"logDir"`

	// PidFile is the path to the PID file
	PidFile string `json:"pidFile" yaml:"pidFile"`

	// CacheSizeMB is the maximum disk cache size in megabytes
	CacheSizeMB int `json:"cacheSizeMB" yaml:"cacheSizeMB"`

	// MaxObjectSizeMB is the maximum size of a single cached object in megabytes
	MaxObjectSizeMB int `json:"maxObjectSizeMB" yaml:"maxObjectSizeMB"`

	// MemoryCacheMB is the amount of memory to use for caching in megabytes
	MemoryCacheMB int `json:"memoryCacheMB" yaml:"memoryCacheMB"`
}

// HttpProxyStatus represents the current state of the HTTP proxy.
type HttpProxyStatus struct {
	// State is "running" or "stopped"
	State string `json:"state"`

	// PID is the process ID (0 if stopped)
	PID int `json:"pid"`

	// Port is the configured port
	Port int `json:"port"`

	// CacheDir is the cache directory path
	CacheDir string `json:"cacheDir"`

	// Version is the squid version
	Version string `json:"version"`

	// Uptime is how long the proxy has been running
	Uptime time.Duration `json:"uptime"`
}

// NewHttpProxyConfig returns a new HttpProxyConfig with default values.
func NewHttpProxyConfig() *HttpProxyConfig {
	return &HttpProxyConfig{
		Port:            3128,
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}
}

// DefaultHttpProxyConfig returns an HttpProxyConfig with sensible defaults
// including paths set relative to the user's home directory.
func DefaultHttpProxyConfig() HttpProxyConfig {
	var baseDir string
	if pc, err := paths.Default(); err == nil {
		baseDir = pc.SquidDir()
	}

	return HttpProxyConfig{
		Port:            3128,
		CacheDir:        filepath.Join(baseDir, "cache"),
		LogDir:          filepath.Join(baseDir, "logs"),
		PidFile:         filepath.Join(baseDir, "squid.pid"),
		CacheSizeMB:     1000,
		MaxObjectSizeMB: 100,
		MemoryCacheMB:   256,
	}
}

// Validate checks if the HTTP proxy configuration is valid.
func (c *HttpProxyConfig) Validate() error {
	// Port must be in valid range and not a reserved/privileged port
	if c.Port < 1024 || c.Port > 65535 {
		return ErrInvalidConfig
	}

	// Directory and file paths must not be empty
	if c.CacheDir == "" {
		return ErrInvalidConfig
	}
	if c.LogDir == "" {
		return ErrInvalidConfig
	}
	if c.PidFile == "" {
		return ErrInvalidConfig
	}

	// Cache sizes must be positive
	if c.CacheSizeMB <= 0 {
		return ErrInvalidConfig
	}
	if c.MaxObjectSizeMB <= 0 {
		return ErrInvalidConfig
	}
	if c.MemoryCacheMB <= 0 {
		return ErrInvalidConfig
	}

	return nil
}

// ApplyDefaults applies default values to any unset configuration fields.
func (c *HttpProxyConfig) ApplyDefaults() {
	if c.Port == 0 {
		c.Port = 3128
	}
	if c.CacheSizeMB == 0 {
		c.CacheSizeMB = 1000
	}
	if c.MaxObjectSizeMB == 0 {
		c.MaxObjectSizeMB = 100
	}
	if c.MemoryCacheMB == 0 {
		c.MemoryCacheMB = 256
	}
}
