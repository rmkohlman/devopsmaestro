// Package prompt provides types and utilities for terminal prompt management.
package prompt

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFile reads and parses a prompt YAML file.
func ParseFile(filename string) (*Prompt, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return Parse(data)
}

// Parse parses YAML data into a Prompt.
func Parse(data []byte) (*Prompt, error) {
	var promptYAML PromptYAML
	if err := yaml.Unmarshal(data, &promptYAML); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate API version and kind
	if promptYAML.APIVersion == "" {
		return nil, fmt.Errorf("missing apiVersion")
	}
	if promptYAML.Kind != "TerminalPrompt" {
		return nil, fmt.Errorf("invalid kind: expected TerminalPrompt, got %s", promptYAML.Kind)
	}

	// Validate required fields
	if promptYAML.Metadata.Name == "" {
		return nil, fmt.Errorf("missing metadata.name")
	}
	if promptYAML.Spec.Type == "" {
		return nil, fmt.Errorf("missing spec.type")
	}

	// Validate prompt type
	switch promptYAML.Spec.Type {
	case PromptTypeStarship, PromptTypePowerlevel10k, PromptTypeOhMyPosh:
		// Valid types
	default:
		return nil, fmt.Errorf("invalid prompt type: %s (must be starship, powerlevel10k, or oh-my-posh)", promptYAML.Spec.Type)
	}

	return promptYAML.ToPrompt(), nil
}

// ToYAMLBytes converts a Prompt to YAML bytes.
func ToYAMLBytes(p *Prompt) ([]byte, error) {
	py := p.ToYAML()
	return yaml.Marshal(py)
}

// ToYAMLString converts a Prompt to a YAML string.
func ToYAMLString(p *Prompt) (string, error) {
	data, err := ToYAMLBytes(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
