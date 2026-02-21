// Package emulator provides types and utilities for terminal emulator management.
package emulator

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// EmulatorYAML represents the YAML structure for terminal emulator configurations.
type EmulatorYAML struct {
	APIVersion string           `yaml:"apiVersion"`
	Kind       string           `yaml:"kind"`
	Metadata   EmulatorMetadata `yaml:"metadata"`
	Spec       EmulatorSpec     `yaml:"spec"`
}

// EmulatorMetadata contains the metadata section of the YAML.
type EmulatorMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
}

// EmulatorSpec contains the specification section of the YAML.
type EmulatorSpec struct {
	Type     EmulatorType   `yaml:"type"`
	Config   map[string]any `yaml:"config,omitempty"`
	ThemeRef string         `yaml:"themeRef,omitempty"`
	Enabled  *bool          `yaml:"enabled,omitempty"`
}

// ParseFile reads and parses an emulator YAML file.
func ParseFile(filename string) (*Emulator, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return Parse(data)
}

// Parse parses YAML data into an Emulator.
func Parse(data []byte) (*Emulator, error) {
	var emulatorYAML EmulatorYAML
	if err := yaml.Unmarshal(data, &emulatorYAML); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate API version and kind
	if emulatorYAML.APIVersion == "" {
		return nil, fmt.Errorf("missing apiVersion")
	}
	if emulatorYAML.Kind != "TerminalEmulator" {
		return nil, fmt.Errorf("invalid kind: expected TerminalEmulator, got %s", emulatorYAML.Kind)
	}

	// Validate required fields
	if emulatorYAML.Metadata.Name == "" {
		return nil, fmt.Errorf("missing metadata.name")
	}

	// Validate emulator type
	switch emulatorYAML.Spec.Type {
	case EmulatorTypeWezterm, EmulatorTypeAlacritty, EmulatorTypeKitty, EmulatorTypeITerm2:
		// Valid
	case "":
		return nil, fmt.Errorf("missing spec.type")
	default:
		return nil, fmt.Errorf("invalid emulator type: %s", emulatorYAML.Spec.Type)
	}

	return emulatorYAML.ToEmulator(), nil
}

// ToEmulator converts EmulatorYAML to an Emulator.
func (ey *EmulatorYAML) ToEmulator() *Emulator {
	emulator := &Emulator{
		Name:        ey.Metadata.Name,
		Description: ey.Metadata.Description,
		Type:        ey.Spec.Type,
		Config:      ey.Spec.Config,
		ThemeRef:    ey.Spec.ThemeRef,
		Category:    ey.Metadata.Category,
		Labels:      ey.Metadata.Labels,
		Enabled:     true, // Default to enabled
	}

	// Override enabled if specified
	if ey.Spec.Enabled != nil {
		emulator.Enabled = *ey.Spec.Enabled
	}

	// Initialize maps if nil
	if emulator.Config == nil {
		emulator.Config = make(map[string]any)
	}
	if emulator.Labels == nil {
		emulator.Labels = make(map[string]string)
	}

	return emulator
}

// ToYAMLBytes converts an Emulator to YAML bytes.
func ToYAMLBytes(e *Emulator) ([]byte, error) {
	ey := e.ToYAML()
	return yaml.Marshal(ey)
}

// ToYAMLString converts an Emulator to a YAML string.
func ToYAMLString(e *Emulator) (string, error) {
	data, err := ToYAMLBytes(e)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToYAML converts an Emulator to EmulatorYAML.
func (e *Emulator) ToYAML() *EmulatorYAML {
	return &EmulatorYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalEmulator",
		Metadata: EmulatorMetadata{
			Name:        e.Name,
			Description: e.Description,
			Category:    e.Category,
			Labels:      e.Labels,
		},
		Spec: EmulatorSpec{
			Type:     e.Type,
			Config:   e.Config,
			ThemeRef: e.ThemeRef,
			Enabled:  &e.Enabled,
		},
	}
}
