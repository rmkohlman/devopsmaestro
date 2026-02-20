package pkg

import (
	"gopkg.in/yaml.v3"
)

// UnmarshalYAML implements custom unmarshaling for StringOrSlice.
// It can handle both string and []string values in YAML.
func (s *StringOrSlice) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.ScalarNode:
		// Single string value
		var str string
		if err := node.Decode(&str); err != nil {
			return err
		}
		*s = StringOrSlice{str}
	case yaml.SequenceNode:
		// Array of strings
		var strs []string
		if err := node.Decode(&strs); err != nil {
			return err
		}
		*s = StringOrSlice(strs)
	default:
		// Try to decode as []string for other cases
		var strs []string
		if err := node.Decode(&strs); err != nil {
			return err
		}
		*s = StringOrSlice(strs)
	}
	return nil
}

// MarshalYAML implements custom marshaling for StringOrSlice.
// Single items are marshaled as scalars, multiple items as arrays.
func (s StringOrSlice) MarshalYAML() (interface{}, error) {
	switch len(s) {
	case 0:
		return []string{}, nil
	case 1:
		return s[0], nil
	default:
		return []string(s), nil
	}
}
