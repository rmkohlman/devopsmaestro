package wezterm

import (
	"testing"
)

func TestNewWezTerm(t *testing.T) {
	config := NewWezTerm("test")

	if config.Name != "test" {
		t.Errorf("Expected name 'test', got %s", config.Name)
	}

	if !config.Enabled {
		t.Error("Expected config to be enabled by default")
	}

	if config.Font.Family != "MesloLGS Nerd Font Mono" {
		t.Errorf("Expected default font family, got %s", config.Font.Family)
	}

	if config.Font.Size != 14 {
		t.Errorf("Expected default font size 14, got %f", config.Font.Size)
	}
}

func TestNewWeztermYAML(t *testing.T) {
	yamlConfig := NewWeztermYAML("test")

	if yamlConfig.APIVersion != "devopsmaestro.dev/v1alpha1" {
		t.Errorf("Expected APIVersion 'devopsmaestro.dev/v1alpha1', got %s", yamlConfig.APIVersion)
	}

	if yamlConfig.Kind != "WeztermConfig" {
		t.Errorf("Expected Kind 'WeztermConfig', got %s", yamlConfig.Kind)
	}

	if yamlConfig.Metadata.Name != "test" {
		t.Errorf("Expected name 'test', got %s", yamlConfig.Metadata.Name)
	}
}

func TestToYAMLAndBack(t *testing.T) {
	// Create a config
	config := NewWezTerm("test")
	config.Description = "Test configuration"
	config.Category = "test"
	config.Tags = []string{"test", "example"}
	config.Scrollback = 5000
	config.Workspace = "workspace"

	// Add colors
	config.Colors = &ColorConfig{
		Foreground:  "#ffffff",
		Background:  "#000000",
		CursorBg:    "#ff0000",
		CursorFg:    "#00ff00",
		SelectionBg: "#0000ff",
		SelectionFg: "#ffff00",
		ANSI:        []string{"#000000", "#ff0000", "#00ff00", "#ffff00", "#0000ff", "#ff00ff", "#00ffff", "#ffffff"},
		Brights:     []string{"#808080", "#ff8080", "#80ff80", "#ffff80", "#8080ff", "#ff80ff", "#80ffff", "#ffffff"},
	}

	// Add leader key
	config.Leader = &LeaderKey{
		Key:     "a",
		Mods:    "CTRL",
		Timeout: 1000,
	}

	// Convert to YAML and back
	yamlConfig := config.ToYAML()
	backConfig := yamlConfig.ToWezTerm()

	if backConfig.Name != config.Name {
		t.Errorf("Name mismatch: expected %s, got %s", config.Name, backConfig.Name)
	}

	if backConfig.Description != config.Description {
		t.Errorf("Description mismatch: expected %s, got %s", config.Description, backConfig.Description)
	}

	if backConfig.Scrollback != config.Scrollback {
		t.Errorf("Scrollback mismatch: expected %d, got %d", config.Scrollback, backConfig.Scrollback)
	}

	if backConfig.Colors == nil {
		t.Error("Colors should not be nil")
	} else {
		if backConfig.Colors.Foreground != config.Colors.Foreground {
			t.Errorf("Foreground color mismatch: expected %s, got %s", config.Colors.Foreground, backConfig.Colors.Foreground)
		}
	}

	if backConfig.Leader == nil {
		t.Error("Leader should not be nil")
	} else {
		if backConfig.Leader.Key != config.Leader.Key {
			t.Errorf("Leader key mismatch: expected %s, got %s", config.Leader.Key, backConfig.Leader.Key)
		}
	}
}

func TestHelperMethods(t *testing.T) {
	config := NewWezTerm("test")

	// Test defaults
	if config.HasColors() {
		t.Error("Should not have colors by default")
	}

	if config.HasThemeRef() {
		t.Error("Should not have theme ref by default")
	}

	if config.HasLeader() {
		t.Error("Should not have leader by default")
	}

	if config.HasKeybindings() {
		t.Error("Should not have keybindings by default")
	}

	if config.HasKeyTables() {
		t.Error("Should not have key tables by default")
	}

	if config.HasPlugins() {
		t.Error("Should not have plugins by default")
	}

	// Add components and test
	config.Colors = &ColorConfig{Foreground: "#ffffff", Background: "#000000"}
	config.ThemeRef = "catppuccin"
	config.Leader = &LeaderKey{Key: "a", Mods: "CTRL", Timeout: 1000}
	config.Keys = []Keybinding{{Key: "t", Action: "SpawnTab"}}
	config.KeyTables = map[string][]Keybinding{
		"test": {{Key: "h", Action: "ActivatePaneDirection"}},
	}
	config.Plugins = []PluginConfig{{Name: "test", Source: "github.com/test/plugin"}}

	if !config.HasColors() {
		t.Error("Should have colors now")
	}

	if !config.HasThemeRef() {
		t.Error("Should have theme ref now")
	}

	if !config.HasLeader() {
		t.Error("Should have leader now")
	}

	if !config.HasKeybindings() {
		t.Error("Should have keybindings now")
	}

	if !config.HasKeyTables() {
		t.Error("Should have key tables now")
	}

	if !config.HasPlugins() {
		t.Error("Should have plugins now")
	}
}

func TestYAMLParser(t *testing.T) {
	yamlData := `
apiVersion: devopsmaestro.dev/v1alpha1
kind: WeztermConfig
metadata:
  name: test-config
  description: Test configuration
  category: test
  tags:
    - test
    - example
spec:
  font:
    family: "Test Font"
    size: 16
  window:
    opacity: 0.9
    blur: 10
  scrollback: 5000
  workspace: test
  enabled: true
`

	parser := NewParser()
	config, err := parser.Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	if config.Name != "test-config" {
		t.Errorf("Expected name 'test-config', got %s", config.Name)
	}

	if config.Font.Family != "Test Font" {
		t.Errorf("Expected font family 'Test Font', got %s", config.Font.Family)
	}

	if config.Font.Size != 16 {
		t.Errorf("Expected font size 16, got %f", config.Font.Size)
	}

	if config.Window.Opacity != 0.9 {
		t.Errorf("Expected window opacity 0.9, got %f", config.Window.Opacity)
	}

	if config.Scrollback != 5000 {
		t.Errorf("Expected scrollback 5000, got %d", config.Scrollback)
	}
}
