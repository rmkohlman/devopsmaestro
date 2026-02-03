// Package profile provides the TerminalProfile type that aggregates
// prompt, plugins, and shell configuration into a complete terminal setup.
package profile

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFile reads and parses a profile YAML file.
func ParseFile(filename string) (*Profile, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return Parse(data)
}

// Parse parses YAML data into a Profile.
func Parse(data []byte) (*Profile, error) {
	var profileYAML ProfileYAML
	if err := yaml.Unmarshal(data, &profileYAML); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate API version and kind
	if profileYAML.APIVersion == "" {
		return nil, fmt.Errorf("missing apiVersion")
	}
	if profileYAML.Kind != "TerminalProfile" {
		return nil, fmt.Errorf("invalid kind: expected TerminalProfile, got %s", profileYAML.Kind)
	}

	// Validate required fields
	if profileYAML.Metadata.Name == "" {
		return nil, fmt.Errorf("missing metadata.name")
	}

	return profileYAML.ToProfile(), nil
}

// ToYAMLBytes converts a Profile to YAML bytes.
func ToYAMLBytes(p *Profile) ([]byte, error) {
	py := p.ToYAML()
	return yaml.Marshal(py)
}

// ToYAMLString converts a Profile to a YAML string.
func ToYAMLString(p *Profile) (string, error) {
	data, err := ToYAMLBytes(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
