// Package wezterm provides types and utilities for WezTerm terminal configuration management.
package wezterm

import "gopkg.in/yaml.v3"

// Parser defines the interface for parsing WezTerm configurations.
type Parser interface {
	Parse(data []byte) (*WezTerm, error)
	ParseYAML(data []byte) (*WeztermYAML, error)
}

// Store defines the interface for WezTerm configuration storage.
type Store interface {
	// Get retrieves a WezTerm configuration by name
	Get(name string) (*WezTerm, error)

	// List returns all WezTerm configurations
	List() ([]*WezTerm, error)

	// Save stores a WezTerm configuration
	Save(config *WezTerm) error

	// Delete removes a WezTerm configuration
	Delete(name string) error

	// Exists checks if a WezTerm configuration exists
	Exists(name string) bool
}

// Library defines the interface for accessing embedded WezTerm configuration presets.
type Library interface {
	// Get retrieves a preset configuration by name
	Get(name string) (*WezTerm, error)

	// List returns all preset configurations
	List() []*WezTerm

	// ListByCategory returns presets in a specific category
	ListByCategory(category string) []*WezTerm

	// Names returns all preset names
	Names() []string

	// Categories returns all unique categories
	Categories() []string

	// Count returns the number of presets
	Count() int
}

// YAMLParser implements Parser for YAML configurations.
type YAMLParser struct{}

// NewParser creates a new YAML parser for WezTerm configurations.
func NewParser() Parser {
	return &YAMLParser{}
}

// Parse parses a WezTerm configuration from YAML bytes.
func (p *YAMLParser) Parse(data []byte) (*WezTerm, error) {
	yamlConfig, err := p.ParseYAML(data)
	if err != nil {
		return nil, err
	}
	return yamlConfig.ToWezTerm(), nil
}

// ParseYAML parses a WeztermYAML configuration from YAML bytes.
func (p *YAMLParser) ParseYAML(data []byte) (*WeztermYAML, error) {
	var config WeztermYAML
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
