package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseYAML parses a CoreConfig from YAML bytes.
func ParseYAML(data []byte) (*CoreConfig, error) {
	var cfg CoreConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse core config YAML: %w", err)
	}

	// Apply defaults for missing fields
	if cfg.Namespace == "" {
		cfg.Namespace = "workspace"
	}
	if cfg.Leader == "" {
		cfg.Leader = " "
	}

	return &cfg, nil
}

// ParseYAMLFile parses a CoreConfig from a YAML file.
func ParseYAMLFile(path string) (*CoreConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	return ParseYAML(data)
}

// ToYAML serializes the CoreConfig to YAML.
func (c *CoreConfig) ToYAML() ([]byte, error) {
	return yaml.Marshal(c)
}

// WriteYAMLFile writes the CoreConfig to a YAML file.
func (c *CoreConfig) WriteYAMLFile(path string) error {
	data, err := c.ToYAML()
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
