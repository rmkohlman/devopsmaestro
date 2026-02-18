package pkg

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ParseYAMLFile reads a package YAML file and returns a Package.
func ParseYAMLFile(path string) (*Package, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}
	return ParseYAML(data)
}

// ParseYAML parses YAML data and returns a Package.
func ParseYAML(data []byte) (*Package, error) {
	var py PackageYAML
	if err := yaml.Unmarshal(data, &py); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate required fields
	if err := validatePackageYAML(&py); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	return py.ToPackage(), nil
}

// ParseYAMLMultiple parses a YAML file that may contain multiple packages.
// Documents are separated by "---".
func ParseYAMLMultiple(data []byte) ([]*Package, error) {
	var packages []*Package
	decoder := yaml.NewDecoder(bytesReader(data))

	for {
		var py PackageYAML
		if err := decoder.Decode(&py); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, fmt.Errorf("failed to parse YAML: %w", err)
		}

		// Skip empty documents
		if py.Metadata.Name == "" && len(py.Spec.Plugins) == 0 {
			continue
		}

		if err := validatePackageYAML(&py); err != nil {
			return nil, fmt.Errorf("validation failed for package %q: %w", py.Metadata.Name, err)
		}

		packages = append(packages, py.ToPackage())
	}

	return packages, nil
}

// validatePackageYAML validates required fields in a PackageYAML.
func validatePackageYAML(py *PackageYAML) error {
	// Validate API version
	if py.APIVersion != "" && py.APIVersion != "devopsmaestro.io/v1" {
		return fmt.Errorf("unsupported apiVersion: expected 'devopsmaestro.io/v1', got %q", py.APIVersion)
	}

	// Validate kind
	if py.Kind != "" && py.Kind != "NvimPackage" {
		return fmt.Errorf("invalid kind: expected 'NvimPackage', got %q", py.Kind)
	}

	// Validate required fields
	if py.Metadata.Name == "" {
		return fmt.Errorf("package metadata.name is required")
	}

	// Validate plugin names (basic check)
	for i, pluginName := range py.Spec.Plugins {
		if pluginName == "" {
			return fmt.Errorf("plugin name at index %d cannot be empty", i)
		}
	}

	// Validate extends field (if present, must not be empty)
	if py.Spec.Extends == py.Metadata.Name {
		return fmt.Errorf("package cannot extend itself")
	}

	return nil
}

// ToYAMLBytes serializes a Package to YAML bytes.
func (p *Package) ToYAMLBytes() ([]byte, error) {
	py := p.ToYAML()
	data, err := yaml.Marshal(py)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal package %q to YAML: %w", p.Name, err)
	}
	return data, nil
}

// bytesReader wraps a byte slice in a Reader for yaml.NewDecoder.
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
