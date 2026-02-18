package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the application configuration
type Config struct {
	Theme       string      `mapstructure:"theme"`       // UI theme (auto, catppuccin-mocha, etc.)
	Credentials Credentials `mapstructure:"credentials"` // Global credentials for builds
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

	err := viper.ReadInConfig()
	if err != nil {
		// It's okay if config file doesn't exist, we'll use defaults
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("Warning: Error loading config file: %v", err)
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
			log.Printf("Warning: Could not create default config: %v", err)
		}
	}

	LoadConfig(configPath)
}

// CreateDefaultConfig creates a default configuration file
func CreateDefaultConfig(configPath string) error {
	// Ensure config directory exists
	if err := os.MkdirAll(configPath, 0755); err != nil {
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
#   GITHUB_USERNAME:
#     source: keychain
#     service: dvm-github-username
#   GITHUB_PAT:
#     source: keychain
#     service: dvm-github-pat
#   NPM_TOKEN:
#     source: env
#     env: MY_NPM_TOKEN
#
# To add credentials to macOS Keychain:
#   security add-generic-password -s "dvm-github-pat" -a "$USER" -w "ghp_yourtoken"
credentials: {}
`

	return os.WriteFile(configFile, []byte(defaultConfig), 0644)
}
