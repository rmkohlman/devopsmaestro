package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestGetTheme_FromEnvironment(t *testing.T) {
	// Save and restore original env
	originalEnv := os.Getenv("DVM_THEME")
	defer func() {
		if originalEnv != "" {
			os.Setenv("DVM_THEME", originalEnv)
		} else {
			os.Unsetenv("DVM_THEME")
		}
	}()

	// Reset viper
	viper.Reset()

	// Set environment variable
	os.Setenv("DVM_THEME", "dracula")

	theme := GetTheme()
	assert.Equal(t, "dracula", theme, "Should use theme from environment variable")
}

func TestGetTheme_FromConfig(t *testing.T) {
	// Save and restore original env
	originalEnv := os.Getenv("DVM_THEME")
	defer func() {
		if originalEnv != "" {
			os.Setenv("DVM_THEME", originalEnv)
		} else {
			os.Unsetenv("DVM_THEME")
		}
		viper.Reset()
	}()

	// Clear environment
	os.Unsetenv("DVM_THEME")

	// Reset and set config
	viper.Reset()
	viper.Set("theme", "tokyo-night")

	theme := GetTheme()
	assert.Equal(t, "tokyo-night", theme, "Should use theme from config")
}

func TestGetTheme_Default(t *testing.T) {
	// Save and restore
	originalEnv := os.Getenv("DVM_THEME")
	defer func() {
		if originalEnv != "" {
			os.Setenv("DVM_THEME", originalEnv)
		} else {
			os.Unsetenv("DVM_THEME")
		}
		viper.Reset()
	}()

	// Clear everything
	os.Unsetenv("DVM_THEME")
	viper.Reset()

	theme := GetTheme()
	assert.Equal(t, "auto", theme, "Should default to 'auto'")
}

func TestGetTheme_PriorityOrder(t *testing.T) {
	// Save and restore
	originalEnv := os.Getenv("DVM_THEME")
	defer func() {
		if originalEnv != "" {
			os.Setenv("DVM_THEME", originalEnv)
		} else {
			os.Unsetenv("DVM_THEME")
		}
		viper.Reset()
	}()

	// Set both env and config
	viper.Reset()
	viper.Set("theme", "nord")
	os.Setenv("DVM_THEME", "catppuccin-mocha")

	theme := GetTheme()
	assert.Equal(t, "catppuccin-mocha", theme, "Environment variable should take priority over config")
}

func TestGetConfig(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	viper.Set("theme", "gruvbox-dark")

	cfg := GetConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "gruvbox-dark", cfg.Theme)
}

func TestGetConfig_Defaults(t *testing.T) {
	viper.Reset()
	defer viper.Reset()

	// Don't set anything
	cfg := GetConfig()
	assert.NotNil(t, cfg)
	assert.Equal(t, "auto", cfg.Theme, "Should return default theme")
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config directory
	tmpDir := t.TempDir()

	// Create a test config file
	configContent := []byte("theme: nord\n")
	configFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configFile, configContent, 0644)
	assert.NoError(t, err)

	// Reset viper
	viper.Reset()
	defer viper.Reset()

	// Load config
	LoadConfig(tmpDir)

	// Check if theme was loaded
	theme := viper.GetString("theme")
	assert.Equal(t, "nord", theme)
}

func TestLoadConfig_MissingFile(t *testing.T) {
	// Create a temporary directory with no config file
	tmpDir := t.TempDir()

	viper.Reset()
	defer viper.Reset()

	// Should not panic when config file is missing
	assert.NotPanics(t, func() {
		LoadConfig(tmpDir)
	})

	// Should use default
	theme := viper.GetString("theme")
	assert.Equal(t, "auto", theme, "Should use default when config file is missing")
}

func TestCreateDefaultConfig(t *testing.T) {
	// Create a temporary directory
	tmpDir := t.TempDir()

	// Create default config
	err := CreateDefaultConfig(tmpDir)
	assert.NoError(t, err)

	// Check if config file was created
	configFile := filepath.Join(tmpDir, "config.yaml")
	_, err = os.Stat(configFile)
	assert.NoError(t, err, "Config file should exist")

	// Read and verify content
	content, err := os.ReadFile(configFile)
	assert.NoError(t, err)
	assert.Contains(t, string(content), "theme: auto", "Config should contain default theme")
	assert.Contains(t, string(content), "DevOpsMaestro Configuration", "Config should have header comment")
}

func TestLoadOrCreateConfig_ExistingFile(t *testing.T) {
	// Create a temporary directory with existing config
	tmpDir := t.TempDir()

	configContent := []byte("theme: dracula\n")
	configFile := filepath.Join(tmpDir, "config.yaml")
	err := os.WriteFile(configFile, configContent, 0644)
	assert.NoError(t, err)

	viper.Reset()
	defer viper.Reset()

	// Load or create config (should load existing)
	LoadOrCreateConfig(tmpDir)

	// Verify it loaded the existing config
	theme := viper.GetString("theme")
	assert.Equal(t, "dracula", theme)
}

func TestLoadOrCreateConfig_NewFile(t *testing.T) {
	// Create a temporary directory without config
	tmpDir := t.TempDir()

	viper.Reset()
	defer viper.Reset()

	// Load or create config (should create new)
	LoadOrCreateConfig(tmpDir)

	// Check if config file was created
	configFile := filepath.Join(tmpDir, "config.yaml")
	_, err := os.Stat(configFile)
	assert.NoError(t, err, "Config file should be created")

	// Verify default was set
	theme := viper.GetString("theme")
	assert.Equal(t, "auto", theme)
}

func TestConfigStruct(t *testing.T) {
	cfg := Config{
		Theme: "catppuccin-latte",
	}

	assert.Equal(t, "catppuccin-latte", cfg.Theme)
}
