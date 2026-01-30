package plugin

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseYAMLFile reads a plugin YAML file and returns a Plugin.
func ParseYAMLFile(path string) (*Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return ParseYAML(data)
}

// ParseYAML parses YAML data and returns a Plugin.
func ParseYAML(data []byte) (*Plugin, error) {
	var py PluginYAML
	if err := yaml.Unmarshal(data, &py); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if err := validatePluginYAML(&py); err != nil {
		return nil, err
	}

	return py.ToPlugin(), nil
}

// ParseYAMLMultiple parses a YAML file that may contain multiple plugins.
// Documents are separated by "---".
func ParseYAMLMultiple(data []byte) ([]*Plugin, error) {
	var plugins []*Plugin
	decoder := yaml.NewDecoder(bytesReader(data))

	for {
		var py PluginYAML
		if err := decoder.Decode(&py); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		// Skip empty documents
		if py.Metadata.Name == "" && py.Spec.Repo == "" {
			continue
		}

		if err := validatePluginYAML(&py); err != nil {
			return nil, err
		}

		plugins = append(plugins, py.ToPlugin())
	}

	return plugins, nil
}

// validatePluginYAML validates required fields in a PluginYAML.
func validatePluginYAML(py *PluginYAML) error {
	// Validate kind
	if py.Kind != "" && py.Kind != "NvimPlugin" {
		return fmt.Errorf("invalid kind: expected 'NvimPlugin', got '%s'", py.Kind)
	}
	if py.Metadata.Name == "" {
		return fmt.Errorf("plugin metadata.name is required")
	}
	if py.Spec.Repo == "" {
		return fmt.Errorf("plugin spec.repo is required")
	}
	return nil
}

// ToYAMLBytes serializes a Plugin to YAML bytes.
func (p *Plugin) ToYAMLBytes() ([]byte, error) {
	py := p.ToYAML()
	return yaml.Marshal(py)
}

// bytesReader wraps a byte slice in a Reader.
type bytesReaderType struct {
	data []byte
	pos  int
}

func bytesReader(data []byte) *bytesReaderType {
	return &bytesReaderType{data: data}
}

func (r *bytesReaderType) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
