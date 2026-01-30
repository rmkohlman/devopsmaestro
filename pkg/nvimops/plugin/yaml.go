package plugin

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAML implements custom unmarshaling for StringOrSlice.
// This allows YAML fields to be either a single string or an array of strings.
func (s *StringOrSlice) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		// Single string value
		var str string
		if err := node.Decode(&str); err != nil {
			return err
		}
		if str != "" {
			*s = []string{str}
		}
		return nil

	case yaml.SequenceNode:
		// Array of strings
		var arr []string
		if err := node.Decode(&arr); err != nil {
			return err
		}
		*s = arr
		return nil

	default:
		return fmt.Errorf("expected string or array, got %v", node.Kind)
	}
}

// MarshalYAML implements custom marshaling for StringOrSlice.
// Single-element slices are marshaled as a scalar.
func (s StringOrSlice) MarshalYAML() (interface{}, error) {
	if len(s) == 0 {
		return nil, nil
	}
	if len(s) == 1 {
		return s[0], nil
	}
	return []string(s), nil
}

// UnmarshalYAML implements custom unmarshaling for DependencyYAML.
// This allows dependencies to be either a string (repo only) or a full spec.
func (d *DependencyYAML) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		// Simple string: just the repo
		var repo string
		if err := node.Decode(&repo); err != nil {
			return err
		}
		d.Repo = repo
		return nil

	case yaml.MappingNode:
		// Full dependency spec
		type depAlias DependencyYAML // Avoid recursion
		var alias depAlias
		if err := node.Decode(&alias); err != nil {
			return err
		}
		*d = DependencyYAML(alias)
		return nil

	default:
		return fmt.Errorf("expected string or object for dependency, got %v", node.Kind)
	}
}

// MarshalYAML implements custom marshaling for DependencyYAML.
// Simple dependencies (repo only) are marshaled as a scalar.
func (d DependencyYAML) MarshalYAML() (interface{}, error) {
	if d.Build == "" && d.Version == "" && d.Branch == "" && !d.Config {
		// Simple dependency, just output the repo string
		return d.Repo, nil
	}
	// Full dependency spec
	type depAlias DependencyYAML
	return depAlias(d), nil
}
