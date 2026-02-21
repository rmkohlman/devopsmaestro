package emulator

import (
	"strings"
	"testing"
)

func TestParse_Valid(t *testing.T) {
	yamlData := `
apiVersion: devopsmaestro.io/v1
kind: TerminalEmulator
metadata:
  name: test-emulator
  description: Test emulator configuration
  category: test
  labels:
    author: test
    style: minimal
spec:
  type: wezterm
  config:
    font:
      family: "JetBrains Mono"
      size: 12
    window:
      opacity: 0.95
  themeRef: tokyonight
  enabled: true
`

	emulator, err := Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to parse valid YAML: %v", err)
	}

	if emulator.Name != "test-emulator" {
		t.Errorf("Expected name 'test-emulator', got %s", emulator.Name)
	}

	if emulator.Description != "Test emulator configuration" {
		t.Errorf("Expected description 'Test emulator configuration', got %s", emulator.Description)
	}

	if emulator.Type != EmulatorTypeWezterm {
		t.Errorf("Expected type wezterm, got %s", emulator.Type)
	}

	if emulator.Category != "test" {
		t.Errorf("Expected category 'test', got %s", emulator.Category)
	}

	if emulator.ThemeRef != "tokyonight" {
		t.Errorf("Expected themeRef 'tokyonight', got %s", emulator.ThemeRef)
	}

	if !emulator.Enabled {
		t.Error("Expected emulator to be enabled")
	}

	// Test config
	if emulator.Config == nil {
		t.Fatal("Config is nil")
	}

	// Test labels
	if emulator.Labels == nil {
		t.Fatal("Labels is nil")
	}

	author, ok := emulator.GetLabel("author")
	if !ok || author != "test" {
		t.Errorf("Expected author label 'test', got %s (exists: %v)", author, ok)
	}
}

func TestParse_MissingAPIVersion(t *testing.T) {
	yamlData := `
kind: TerminalEmulator
metadata:
  name: test-emulator
spec:
  type: wezterm
`

	_, err := Parse([]byte(yamlData))
	if err == nil {
		t.Error("Expected error for missing apiVersion")
	}

	if !strings.Contains(err.Error(), "missing apiVersion") {
		t.Errorf("Expected 'missing apiVersion' error, got: %v", err)
	}
}

func TestParse_InvalidKind(t *testing.T) {
	yamlData := `
apiVersion: devopsmaestro.io/v1
kind: InvalidKind
metadata:
  name: test-emulator
spec:
  type: wezterm
`

	_, err := Parse([]byte(yamlData))
	if err == nil {
		t.Error("Expected error for invalid kind")
	}

	if !strings.Contains(err.Error(), "invalid kind") {
		t.Errorf("Expected 'invalid kind' error, got: %v", err)
	}
}

func TestParse_MissingName(t *testing.T) {
	yamlData := `
apiVersion: devopsmaestro.io/v1
kind: TerminalEmulator
metadata:
  description: Test emulator
spec:
  type: wezterm
`

	_, err := Parse([]byte(yamlData))
	if err == nil {
		t.Error("Expected error for missing name")
	}

	if !strings.Contains(err.Error(), "missing metadata.name") {
		t.Errorf("Expected 'missing metadata.name' error, got: %v", err)
	}
}

func TestParse_InvalidEmulatorType(t *testing.T) {
	yamlData := `
apiVersion: devopsmaestro.io/v1
kind: TerminalEmulator
metadata:
  name: test-emulator
spec:
  type: invalid-type
`

	_, err := Parse([]byte(yamlData))
	if err == nil {
		t.Error("Expected error for invalid emulator type")
	}

	if !strings.Contains(err.Error(), "invalid emulator type") {
		t.Errorf("Expected 'invalid emulator type' error, got: %v", err)
	}
}

func TestParse_MissingType(t *testing.T) {
	yamlData := `
apiVersion: devopsmaestro.io/v1
kind: TerminalEmulator
metadata:
  name: test-emulator
spec:
  config:
    font: "JetBrains Mono"
`

	_, err := Parse([]byte(yamlData))
	if err == nil {
		t.Error("Expected error for missing type")
	}

	if !strings.Contains(err.Error(), "missing spec.type") {
		t.Errorf("Expected 'missing spec.type' error, got: %v", err)
	}
}

func TestParse_DefaultValues(t *testing.T) {
	yamlData := `
apiVersion: devopsmaestro.io/v1
kind: TerminalEmulator
metadata:
  name: minimal-test
spec:
  type: wezterm
`

	emulator, err := Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("Failed to parse minimal YAML: %v", err)
	}

	// Should default to enabled
	if !emulator.Enabled {
		t.Error("Expected default enabled to be true")
	}

	// Config and Labels should be initialized
	if emulator.Config == nil {
		t.Error("Expected Config to be initialized")
	}

	if emulator.Labels == nil {
		t.Error("Expected Labels to be initialized")
	}
}

func TestToYAML(t *testing.T) {
	emulator := &Emulator{
		Name:        "test-emulator",
		Description: "Test emulator",
		Type:        EmulatorTypeWezterm,
		Category:    "test",
		ThemeRef:    "tokyonight",
		Enabled:     true,
		Config: map[string]any{
			"font": map[string]any{
				"family": "JetBrains Mono",
				"size":   12,
			},
		},
		Labels: map[string]string{
			"author": "test",
		},
	}

	yamlEmulator := emulator.ToYAML()

	if yamlEmulator.APIVersion != "devopsmaestro.io/v1" {
		t.Errorf("Expected apiVersion 'devopsmaestro.io/v1', got %s", yamlEmulator.APIVersion)
	}

	if yamlEmulator.Kind != "TerminalEmulator" {
		t.Errorf("Expected kind 'TerminalEmulator', got %s", yamlEmulator.Kind)
	}

	if yamlEmulator.Metadata.Name != "test-emulator" {
		t.Errorf("Expected name 'test-emulator', got %s", yamlEmulator.Metadata.Name)
	}

	if yamlEmulator.Spec.Type != EmulatorTypeWezterm {
		t.Errorf("Expected type wezterm, got %s", yamlEmulator.Spec.Type)
	}

	if yamlEmulator.Spec.Enabled == nil || !*yamlEmulator.Spec.Enabled {
		t.Error("Expected enabled to be true")
	}
}

func TestToYAMLBytes(t *testing.T) {
	emulator := &Emulator{
		Name:    "test-emulator",
		Type:    EmulatorTypeWezterm,
		Enabled: true,
		Config:  make(map[string]any),
		Labels:  make(map[string]string),
	}

	yamlBytes, err := ToYAMLBytes(emulator)
	if err != nil {
		t.Fatalf("Failed to convert to YAML bytes: %v", err)
	}

	if len(yamlBytes) == 0 {
		t.Error("YAML bytes is empty")
	}

	// Should be able to parse it back
	parsedEmulator, err := Parse(yamlBytes)
	if err != nil {
		t.Fatalf("Failed to parse generated YAML: %v", err)
	}

	if parsedEmulator.Name != emulator.Name {
		t.Errorf("Round-trip failed: expected name %s, got %s", emulator.Name, parsedEmulator.Name)
	}
}

func TestToYAMLString(t *testing.T) {
	emulator := &Emulator{
		Name:    "test-emulator",
		Type:    EmulatorTypeWezterm,
		Enabled: true,
		Config:  make(map[string]any),
		Labels:  make(map[string]string),
	}

	yamlString, err := ToYAMLString(emulator)
	if err != nil {
		t.Fatalf("Failed to convert to YAML string: %v", err)
	}

	if yamlString == "" {
		t.Error("YAML string is empty")
	}

	if !strings.Contains(yamlString, "apiVersion: devopsmaestro.io/v1") {
		t.Error("YAML string should contain apiVersion")
	}

	if !strings.Contains(yamlString, "kind: TerminalEmulator") {
		t.Error("YAML string should contain kind")
	}
}
