// Package plugin provides types and utilities for shell plugin management.
package plugin

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseFile reads and parses a plugin YAML file.
func ParseFile(filename string) (*Plugin, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return Parse(data)
}

// Parse parses YAML data into a Plugin.
func Parse(data []byte) (*Plugin, error) {
	var pluginYAML PluginYAML
	if err := yaml.Unmarshal(data, &pluginYAML); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate API version and kind
	if pluginYAML.APIVersion == "" {
		return nil, fmt.Errorf("missing apiVersion")
	}
	if pluginYAML.Kind != "TerminalPlugin" {
		return nil, fmt.Errorf("invalid kind: expected TerminalPlugin, got %s", pluginYAML.Kind)
	}

	// Validate required fields
	if pluginYAML.Metadata.Name == "" {
		return nil, fmt.Errorf("missing metadata.name")
	}

	// Must have either repo, source, or ohmyzsh_plugin
	if pluginYAML.Spec.Repo == "" && pluginYAML.Spec.Source == "" && pluginYAML.Spec.OhMyZshPlugin == "" {
		return nil, fmt.Errorf("must specify repo, source, or ohmyzshPlugin")
	}

	// Validate manager if specified
	if pluginYAML.Spec.Manager != "" {
		switch pluginYAML.Spec.Manager {
		case PluginManagerZinit, PluginManagerOhMyZsh, PluginManagerAntigen, PluginManagerManual, PluginManagerSheldon:
			// Valid
		default:
			return nil, fmt.Errorf("invalid manager: %s", pluginYAML.Spec.Manager)
		}
	}

	// Validate load mode if specified
	if pluginYAML.Spec.LoadMode != "" {
		switch pluginYAML.Spec.LoadMode {
		case LoadModeImmediate, LoadModeDeferred, LoadModeLazy:
			// Valid
		default:
			return nil, fmt.Errorf("invalid loadMode: %s", pluginYAML.Spec.LoadMode)
		}
	}

	return pluginYAML.ToPlugin(), nil
}

// ToYAMLBytes converts a Plugin to YAML bytes.
func ToYAMLBytes(p *Plugin) ([]byte, error) {
	py := p.ToYAML()
	return yaml.Marshal(py)
}

// ToYAMLString converts a Plugin to a YAML string.
func ToYAMLString(p *Plugin) (string, error) {
	data, err := ToYAMLBytes(p)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
