// Package resource provides a unified interface for managing resources
// in DevOpsMaestro. It follows the kubectl pattern where resources are
// identified by Kind and can be applied, retrieved, listed, and deleted
// through a common interface.
//
// # Architecture
//
// The package separates concerns into:
//   - Resource: The data being managed (plugins, themes, workspaces, etc.)
//   - Handler: Knows how to CRUD a specific resource type
//   - Registry: Routes operations to the correct handler by Kind
//
// # Usage
//
//	// Register handlers at startup
//	resource.Register(&NvimPluginHandler{})
//	resource.Register(&NvimThemeHandler{})
//
//	// Apply a resource (auto-detects Kind from YAML)
//	result, err := resource.Apply(ctx, yamlData, "source.yaml")
//
//	// Get a specific resource
//	res, err := resource.Get(ctx, "NvimPlugin", "telescope")
//
//	// List all resources of a type
//	resources, err := resource.List(ctx, "NvimPlugin")
//
//	// Delete a resource
//	err := resource.Delete(ctx, "NvimPlugin", "telescope")
package resource

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Resource represents any manageable resource in DevOpsMaestro.
// All resource types (plugins, themes, workspaces, etc.) implement this interface.
type Resource interface {
	// GetKind returns the resource type (e.g., "NvimPlugin", "NvimTheme", "Workspace")
	GetKind() string

	// GetName returns the unique name of this resource
	GetName() string

	// Validate checks if the resource is valid
	Validate() error
}

// Handler knows how to manage a specific resource type.
// Each resource type (NvimPlugin, NvimTheme, etc.) has its own handler.
type Handler interface {
	// Kind returns the resource type this handler manages (e.g., "NvimPlugin")
	Kind() string

	// Apply creates or updates a resource from YAML data
	// Returns the applied resource and any error
	Apply(ctx Context, data []byte) (Resource, error)

	// Get retrieves a resource by name
	Get(ctx Context, name string) (Resource, error)

	// List returns all resources of this type
	List(ctx Context) ([]Resource, error)

	// Delete removes a resource by name
	Delete(ctx Context, name string) error

	// ToYAML serializes a resource to YAML bytes
	ToYAML(res Resource) ([]byte, error)
}

// Context provides dependencies needed by handlers.
// This allows handlers to access storage without tight coupling.
type Context struct {
	// DataStore is the database store (for dvm)
	// Handlers should type-assert this to their specific store interface.
	// May be nil if using file-based storage.
	DataStore any

	// PluginStore is a pre-configured plugin store (optional).
	// If provided, handlers should use this instead of creating their own.
	PluginStore any

	// ThemeStore is a pre-configured theme store (optional).
	// If provided, handlers should use this instead of creating their own.
	ThemeStore any

	// ConfigDir is the configuration directory (for nvp file-based storage)
	// e.g., ~/.nvp
	ConfigDir string
}

// KindHeader is used to detect the Kind from YAML before full parsing
type KindHeader struct {
	Kind string `yaml:"kind"`
}

// DetectKind extracts the Kind field from YAML data without full parsing
func DetectKind(data []byte) (string, error) {
	var header KindHeader
	if err := yaml.Unmarshal(data, &header); err != nil {
		return "", fmt.Errorf("failed to parse YAML header: %w", err)
	}
	if header.Kind == "" {
		return "", fmt.Errorf("YAML missing required 'kind' field")
	}
	return header.Kind, nil
}
