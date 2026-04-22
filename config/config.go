package config

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// VaultConfig holds MaestroVault-related configuration.
type VaultConfig struct {
	Token string `mapstructure:"token"` // Vault token (can also be set via MAV_TOKEN env or .vault_token file)
}

// BuildLogsConfig controls per-session build log file capture and rotation.
// See pkg/buildlog for the implementation.
type BuildLogsConfig struct {
	Enabled    bool   `mapstructure:"enabled"`    // default true
	Directory  string `mapstructure:"directory"`  // default ~/.devopsmaestro/logs/builds
	MaxSizeMB  int    `mapstructure:"maxSizeMB"`  // default 100
	MaxAgeDays int    `mapstructure:"maxAgeDays"` // default 7
	MaxBackups int    `mapstructure:"maxBackups"` // default 10
	Compress   bool   `mapstructure:"compress"`   // default true
}

// Config represents the application configuration
type Config struct {
	Theme       string          `mapstructure:"theme"`       // UI theme (auto, catppuccin-mocha, etc.)
	Credentials Credentials     `mapstructure:"credentials"` // Global credentials for builds
	Vault       VaultConfig     `mapstructure:"vault"`       // MaestroVault configuration
	BuildLogs   BuildLogsConfig `mapstructure:"buildLogs"`   // Build log capture / rotation
}

// GetConfig returns the current configuration
func GetConfig() *Config {
	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		// Return defaults if unmarshal fails
		return &Config{
			Theme:       "auto",
			Credentials: make(Credentials),
		}
	}

	// If theme is empty, set default
	if cfg.Theme == "" {
		cfg.Theme = "auto"
	}

	// Ensure credentials map is initialized
	if cfg.Credentials == nil {
		cfg.Credentials = make(Credentials)
	}

	return &cfg
}

// GetTheme returns the configured theme, checking in order:
// 1. DVM_THEME environment variable
// 2. config file theme setting
// 3. default "auto"
func GetTheme() string {
	// Check environment variable first
	if theme := os.Getenv("DVM_THEME"); theme != "" {
		return theme
	}

	// Check config file
	if viper.IsSet("theme") {
		return viper.GetString("theme")
	}

	// Default to auto
	return "auto"
}

// LoadConfig loads configuration from the specified path
func LoadConfig(configPath string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()

	// Set defaults
	viper.SetDefault("theme", "auto")
	viper.SetDefault("buildLogs.enabled", true)
	viper.SetDefault("buildLogs.directory", "~/.devopsmaestro/logs/builds")
	viper.SetDefault("buildLogs.maxSizeMB", 100)
	viper.SetDefault("buildLogs.maxAgeDays", 7)
	viper.SetDefault("buildLogs.maxBackups", 10)
	viper.SetDefault("buildLogs.compress", true)

	err := viper.ReadInConfig()
	if err != nil {
		// It's okay if config file doesn't exist, we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Warn("error loading config file", "error", err)
		}
	}
}

// LoadOrCreateConfig loads config or creates a default one
func LoadOrCreateConfig(configPath string) {
	configFile := filepath.Join(configPath, "config.yaml")

	// Check if config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		// Create default config
		if err := CreateDefaultConfig(configPath); err != nil {
			slog.Warn("could not create default config", "error", err)
		}
	}

	LoadConfig(configPath)
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig(configPath string) error {
	// Ensure config directory exists
	if err := os.MkdirAll(configPath, 0700); err != nil {
		return err
	}

	configFile := filepath.Join(configPath, "config.yaml")

	defaultConfig := `# DevOpsMaestro Configuration File
# For more information, see: https://github.com/rmkohlman/devopsmaestro

# UI Theme
# Options: auto, catppuccin-mocha, catppuccin-latte, tokyo-night, nord, dracula, gruvbox-dark, gruvbox-light
# Default: auto (automatically adapts to your terminal's light/dark theme)
theme: auto

# Global Credentials
# These are used during 'dvm build' for private repository access.
# Credentials are inherited: Global -> Ecosystem -> Domain -> App -> Workspace
# Environment variables always take highest priority.
#
# Example:
# credentials:
#   GITHUB_PAT:
#     source: vault
#     vaultSecret: github-pat
#     vaultEnvironment: production
#   NPM_TOKEN:
#     source: env
#     env: MY_NPM_TOKEN
#
# To store secrets in MaestroVault:
#   mav set github-pat production "ghp_yourtoken"
credentials: {}
`

	return os.WriteFile(configFile, []byte(defaultConfig), 0600)
}
