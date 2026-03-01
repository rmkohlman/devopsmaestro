package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/viper"
)

// RegistryConfig holds configuration for the local OCI registry.
type RegistryConfig struct {
	// Enabled determines if the registry should be used
	Enabled bool `yaml:"enabled" json:"enabled" mapstructure:"enabled"`

	// Lifecycle determines when the registry runs: "persistent", "on-demand", or "manual"
	// - persistent: Starts on boot, stays running
	// - on-demand: Starts when needed, stops after idle timeout
	// - manual: Only starts/stops via explicit commands
	Lifecycle string `yaml:"lifecycle" json:"lifecycle" mapstructure:"lifecycle"`

	// Port is the port the registry listens on
	Port int `yaml:"port" json:"port" mapstructure:"port"`

	// Storage is the directory where registry data is stored
	Storage string `yaml:"storage" json:"storage" mapstructure:"storage"`

	// IdleTimeout is how long to wait before shutting down an idle registry (on-demand mode)
	IdleTimeout time.Duration `yaml:"idleTimeout" json:"idleTimeout" mapstructure:"idleTimeout"`

	// Mirrors are the upstream registries to mirror/cache
	Mirrors []MirrorConfig `yaml:"mirrors" json:"mirrors" mapstructure:"mirrors"`
}

// MirrorConfig configures an upstream registry to mirror.
type MirrorConfig struct {
	// Name is a friendly name for this mirror
	Name string `yaml:"name" json:"name" mapstructure:"name"`

	// URL is the upstream registry URL (e.g., "https://index.docker.io")
	URL string `yaml:"url" json:"url" mapstructure:"url"`

	// OnDemand enables pull-through caching (download on first pull)
	OnDemand bool `yaml:"onDemand" json:"onDemand" mapstructure:"onDemand"`

	// Prefix is the registry prefix (e.g., "docker.io")
	Prefix string `yaml:"prefix" json:"prefix" mapstructure:"prefix"`
}

// GetRegistryConfig returns the registry configuration from viper.
// If not configured, returns sensible defaults.
func GetRegistryConfig() *RegistryConfig {
	// Check if registry config exists
	if !viper.IsSet("registry") {
		return DefaultRegistryConfig()
	}

	var cfg RegistryConfig
	if err := viper.UnmarshalKey("registry", &cfg); err != nil {
		return DefaultRegistryConfig()
	}

	// Apply defaults for missing values
	if cfg.Port == 0 {
		cfg.Port = 5001
	}
	if cfg.Storage == "" {
		homeDir, _ := os.UserHomeDir()
		cfg.Storage = filepath.Join(homeDir, ".devopsmaestro", "registry")
	}
	if cfg.Lifecycle == "" {
		cfg.Lifecycle = "on-demand"
	}
	if cfg.IdleTimeout == 0 {
		cfg.IdleTimeout = 30 * time.Minute
	}
	if len(cfg.Mirrors) == 0 {
		cfg.Mirrors = defaultMirrors()
	}

	return &cfg
}

// DefaultRegistryConfig returns a RegistryConfig with sensible defaults.
func DefaultRegistryConfig() *RegistryConfig {
	homeDir, _ := os.UserHomeDir()
	return &RegistryConfig{
		Enabled:     true,
		Lifecycle:   "on-demand",
		Port:        5001,
		Storage:     filepath.Join(homeDir, ".devopsmaestro", "registry"),
		IdleTimeout: 30 * time.Minute,
		Mirrors:     defaultMirrors(),
	}
}

// defaultMirrors returns the default set of registry mirrors.
func defaultMirrors() []MirrorConfig {
	return []MirrorConfig{
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
	}
}

// IsRegistryEnabled returns true if the registry is enabled in config.
func IsRegistryEnabled() bool {
	cfg := GetRegistryConfig()
	return cfg.Enabled
}

// GetRegistryEndpoint returns the registry endpoint (e.g., "localhost:5001").
func GetRegistryEndpoint() string {
	cfg := GetRegistryConfig()
	return fmt.Sprintf("localhost:%d", cfg.Port)
}
