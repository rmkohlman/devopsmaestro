// Package extension provides types for PromptExtension resources.
// This package has ZERO dependencies on dvm internals and can be used standalone.
package extension

import (
	"fmt"
	"regexp"
	"time"
)

// Kind constant for PromptExtension resource.
const KindPromptExtension = "PromptExtension"

// PromptExtension represents a module extension for a prompt segment.
// This is the canonical type used throughout terminal-manager.
type PromptExtension struct {
	// Core identification
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Target segment
	Segment string `json:"segment,omitempty" yaml:"segment,omitempty"`

	// Module definitions
	Provides []string                   `json:"provides,omitempty" yaml:"provides,omitempty"`
	Modules  map[string]ExtensionModule `json:"modules,omitempty" yaml:"modules,omitempty"`

	// Metadata
	Category string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Enabled  bool     `json:"enabled" yaml:"enabled"`

	// Timestamps (optional, used when stored)
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"-"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"-"`
}

// ExtensionModule represents a single module configuration within an extension.
type ExtensionModule struct {
	Disabled bool           `json:"disabled,omitempty" yaml:"disabled,omitempty"`
	Symbol   string         `json:"symbol,omitempty" yaml:"symbol,omitempty"`
	Format   string         `json:"format,omitempty" yaml:"format,omitempty"`
	Style    string         `json:"style,omitempty" yaml:"style,omitempty"`
	Options  map[string]any `json:"options,omitempty" yaml:"options,omitempty"`
}

// PromptExtensionYAML represents the full YAML document format (kubectl-style).
type PromptExtensionYAML struct {
	APIVersion string                  `yaml:"apiVersion"`
	Kind       string                  `yaml:"kind"`
	Metadata   PromptExtensionMetadata `yaml:"metadata"`
	Spec       PromptExtensionSpec     `yaml:"spec"`
}

// PromptExtensionMetadata contains extension metadata in the YAML format.
type PromptExtensionMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// PromptExtensionSpec contains the extension specification in the YAML format.
type PromptExtensionSpec struct {
	Segment  string                     `yaml:"segment,omitempty"`
	Provides []string                   `yaml:"provides,omitempty"`
	Modules  map[string]ExtensionModule `yaml:"modules,omitempty"`
	Enabled  *bool                      `yaml:"enabled,omitempty"`
}

// NewPromptExtension creates a new PromptExtension with default values.
func NewPromptExtension(name string) *PromptExtension {
	return &PromptExtension{
		Name:     name,
		Provides: make([]string, 0),
		Modules:  make(map[string]ExtensionModule),
		Enabled:  true,
	}
}

// NewPromptExtensionYAML creates a new PromptExtensionYAML with proper defaults.
func NewPromptExtensionYAML(name string) *PromptExtensionYAML {
	return &PromptExtensionYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       KindPromptExtension,
		Metadata: PromptExtensionMetadata{
			Name: name,
		},
		Spec: PromptExtensionSpec{
			Provides: make([]string, 0),
			Modules:  make(map[string]ExtensionModule),
		},
	}
}

// ToPromptExtension converts PromptExtensionYAML to the canonical PromptExtension type.
func (pey *PromptExtensionYAML) ToPromptExtension() *PromptExtension {
	// Default to enabled unless explicitly set to false
	enabled := true
	if pey.Spec.Enabled != nil {
		enabled = *pey.Spec.Enabled
	}

	return &PromptExtension{
		Name:        pey.Metadata.Name,
		Description: pey.Metadata.Description,
		Category:    pey.Metadata.Category,
		Tags:        pey.Metadata.Tags,
		Segment:     pey.Spec.Segment,
		Provides:    pey.Spec.Provides,
		Modules:     pey.Spec.Modules,
		Enabled:     enabled,
	}
}

// GetProvides returns the list of module names this extension provides.
func (pe *PromptExtension) GetProvides() []string {
	return pe.Provides
}

// GetSegment returns the segment name this extension targets.
func (pe *PromptExtension) GetSegment() string {
	return pe.Segment
}

// Resource interface methods for PromptExtensionYAML

// GetKind returns the resource kind.
func (pey *PromptExtensionYAML) GetKind() string {
	return KindPromptExtension
}

// GetName returns the resource name.
func (pey *PromptExtensionYAML) GetName() string {
	return pey.Metadata.Name
}

// GetAPIVersion returns the API version.
func (pey *PromptExtensionYAML) GetAPIVersion() string {
	return pey.APIVersion
}

// Validate checks if the PromptExtensionYAML resource is valid.
func (pey *PromptExtensionYAML) Validate() error {
	if pey.Metadata.Name == "" {
		return fmt.Errorf("metadata.name is required")
	}

	if pey.Spec.Segment == "" {
		return fmt.Errorf("spec.segment is required")
	}

	if len(pey.Spec.Modules) == 0 {
		return fmt.Errorf("at least one module is required")
	}

	// Check that module styles use palette keys only (no hex colors)
	hexColorRegex := regexp.MustCompile(`#[0-9A-Fa-f]{3,8}`)
	for moduleName, module := range pey.Spec.Modules {
		if module.Style != "" {
			// Parse style string format: "bg:color fg:color"
			// Check if any color value is a hex color
			if hexColorRegex.MatchString(module.Style) {
				return fmt.Errorf("module %q style contains hex color, must use palette keys only", moduleName)
			}
		}
	}

	return nil
}
