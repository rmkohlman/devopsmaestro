package registry

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/rmkohlman/MaestroSDK/paths"

	"gopkg.in/yaml.v3"
)

// NpmProxyConfig holds configuration for the verdaccio npm proxy.
type NpmProxyConfig struct {
	// Enabled determines if the proxy should be used
	Enabled bool `yaml:"enabled"`

	// Lifecycle determines when the proxy runs: "persistent", "on-demand", or "manual"
	Lifecycle string `yaml:"lifecycle"`

	// Port is the port the proxy listens on
	Port int `yaml:"port"`

	// Storage is the directory where proxy data is stored
	Storage string `yaml:"storage"`

	// ConfigFile is the path to the verdaccio config file (optional)
	ConfigFile string `yaml:"configFile,omitempty"`

	// IdleTimeout is how long to wait before shutting down an idle proxy (on-demand mode)
	IdleTimeout time.Duration `yaml:"idleTimeout"`

	// MaxBodySize is the maximum allowed package size (e.g., "10mb")
	MaxBodySize string `yaml:"maxBodySize,omitempty"`

	// Upstreams are the upstream npm registries to use
	Upstreams []NpmUpstreamConfig `yaml:"upstreams"`

	// Auth configures authentication/authorization
	Auth NpmAuthConfig `yaml:"auth,omitempty"`
}

// NpmUpstreamConfig configures an upstream npm registry.
type NpmUpstreamConfig struct {
	// Name is a friendly name for this upstream
	Name string `yaml:"name"`

	// URL is the upstream registry URL (e.g., "https://registry.npmjs.org")
	URL string `yaml:"url"`
}

// NpmAuthConfig configures authentication for the npm proxy.
type NpmAuthConfig struct {
	// Enabled determines if authentication is enabled
	Enabled bool `yaml:"enabled"`

	// Type is the authentication type ("htpasswd" is the only supported type currently)
	Type string `yaml:"type,omitempty"`

	// HTPasswd configures htpasswd-based authentication
	HTPasswd *HTPasswdConfig `yaml:"htpasswd,omitempty"`
}

// HTPasswdConfig configures htpasswd authentication.
type HTPasswdConfig struct {
	// File is the path to the htpasswd file
	File string `yaml:"file"`

	// MaxUsers is the maximum number of users allowed
	MaxUsers int `yaml:"max_users,omitempty"`
}

// NpmProxyStatus represents the current state of the npm proxy.
type NpmProxyStatus struct {
	// State is "running" or "stopped"
	State string

	// PID is the process ID (0 if stopped)
	PID int

	// Port is the configured port
	Port int

	// Storage is the storage directory path
	Storage string

	// Version is the verdaccio version
	Version string

	// Uptime is how long the proxy has been running
	Uptime time.Duration

	// PackageCount is the number of packages cached
	PackageCount int

	// DiskUsage is the total disk space used (bytes)
	DiskUsage int64
}

// Validate checks if the npm proxy configuration is valid.
func (c *NpmProxyConfig) Validate() error {
	// Port must be in valid range and not a reserved/privileged port
	if c.Port < 1024 || c.Port > 65535 {
		return ErrInvalidConfig
	}
	if c.Storage == "" {
		return ErrInvalidConfig
	}
	if c.Lifecycle != "persistent" && c.Lifecycle != "on-demand" && c.Lifecycle != "manual" {
		return ErrInvalidConfig
	}

	// Validate upstream URLs
	for _, upstream := range c.Upstreams {
		if err := upstream.Validate(); err != nil {
			return err
		}
	}

	// Validate auth config
	if c.Auth.Enabled {
		if err := c.Auth.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// Validate checks if the upstream registry configuration is valid.
func (u *NpmUpstreamConfig) Validate() error {
	if u.URL == "" {
		return ErrInvalidConfig
	}
	// Simple URL validation - check for http:// or https://
	if len(u.URL) < 8 || (u.URL[:7] != "http://" && u.URL[:8] != "https://") {
		return ErrInvalidConfig
	}
	return nil
}

// Validate checks if the auth configuration is valid.
func (a *NpmAuthConfig) Validate() error {
	if !a.Enabled {
		return nil // Disabled auth is always valid
	}

	// If auth is enabled, type must be specified and valid
	if a.Type == "" {
		return ErrInvalidConfig
	}

	// Check supported auth types
	supportedTypes := map[string]bool{
		"htpasswd": true,
		"ldap":     true,
	}

	if !supportedTypes[a.Type] {
		return ErrInvalidConfig
	}

	return nil
}

// GenerateVerdaccioConfig generates a verdaccio configuration from an NpmProxyConfig.
// Returns YAML configuration as a string.
func GenerateVerdaccioConfig(cfg NpmProxyConfig) (string, error) {
	// Validate config first
	if err := cfg.Validate(); err != nil {
		return "", fmt.Errorf("invalid configuration: %w", err)
	}

	// Use default upstreams if none specified
	upstreams := cfg.Upstreams
	if len(upstreams) == 0 {
		upstreams = defaultNpmUpstreams()
	}

	// Build verdaccio YAML config structure
	config := map[string]interface{}{
		"storage": filepath.Join(cfg.Storage, "storage"),
		"auth": map[string]interface{}{
			"htpasswd": map[string]interface{}{
				"file": "./htpasswd",
			},
		},
		"uplinks": make(map[string]interface{}),
		"packages": map[string]interface{}{
			"@*/*": map[string]interface{}{
				"access":  "$all",
				"publish": "$authenticated",
				"proxy":   []string{},
			},
			"**": map[string]interface{}{
				"access":  "$all",
				"publish": "$authenticated",
				"proxy":   []string{},
			},
		},
		"listen": fmt.Sprintf("127.0.0.1:%d", cfg.Port),
	}

	// Add max_body_size if specified
	if cfg.MaxBodySize != "" {
		config["max_body_size"] = cfg.MaxBodySize
	}

	// Add upstreams
	uplinks := config["uplinks"].(map[string]interface{})
	proxyList := []string{}
	for _, upstream := range upstreams {
		uplinks[upstream.Name] = map[string]interface{}{
			"url": upstream.URL,
		}
		proxyList = append(proxyList, upstream.Name)
	}

	// Update package proxies
	packages := config["packages"].(map[string]interface{})
	for key := range packages {
		pkg := packages[key].(map[string]interface{})
		pkg["proxy"] = proxyList
	}

	// Convert to YAML
	yamlBytes, err := yaml.Marshal(config)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(yamlBytes), nil
}

// DefaultNpmProxyConfig returns an NpmProxyConfig with sensible defaults.
func DefaultNpmProxyConfig() NpmProxyConfig {
	var storage string
	if pc, err := paths.Default(); err == nil {
		storage = pc.VerdaccioStorage()
	}
	return NpmProxyConfig{
		Enabled:     true,
		Lifecycle:   "on-demand",
		Port:        4873,
		Storage:     storage,
		IdleTimeout: 30 * time.Minute,
		MaxBodySize: "10mb",
		Upstreams:   defaultNpmUpstreams(),
	}
}

// defaultNpmUpstreams returns the default set of upstream npm registries.
func defaultNpmUpstreams() []NpmUpstreamConfig {
	return []NpmUpstreamConfig{
		{
			Name: "npmjs",
			URL:  "https://registry.npmjs.org",
		},
	}
}
