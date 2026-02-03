// Package shell provides types and utilities for shell configuration management.
package shell

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFile reads and parses a shell YAML file.
func ParseFile(filename string) (*Shell, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return Parse(data)
}

// Parse parses YAML data into a Shell.
func Parse(data []byte) (*Shell, error) {
	var shellYAML ShellYAML
	if err := yaml.Unmarshal(data, &shellYAML); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate API version and kind
	if shellYAML.APIVersion == "" {
		return nil, fmt.Errorf("missing apiVersion")
	}
	if shellYAML.Kind != "TerminalShell" {
		return nil, fmt.Errorf("invalid kind: expected TerminalShell, got %s", shellYAML.Kind)
	}

	// Validate required fields
	if shellYAML.Metadata.Name == "" {
		return nil, fmt.Errorf("missing metadata.name")
	}
	if shellYAML.Spec.ShellType == "" {
		return nil, fmt.Errorf("missing spec.shellType")
	}

	// Validate shell type
	switch shellYAML.Spec.ShellType {
	case ShellTypeZsh, ShellTypeBash, ShellTypeFish:
		// Valid
	default:
		return nil, fmt.Errorf("invalid shellType: %s (must be zsh, bash, or fish)", shellYAML.Spec.ShellType)
	}

	return shellYAML.ToShell(), nil
}

// ToYAMLBytes converts a Shell to YAML bytes.
func ToYAMLBytes(s *Shell) ([]byte, error) {
	sy := s.ToYAML()
	return yaml.Marshal(sy)
}

// ToYAMLString converts a Shell to a YAML string.
func ToYAMLString(s *Shell) (string, error) {
	data, err := ToYAMLBytes(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
