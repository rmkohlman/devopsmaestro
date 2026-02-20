// Package pkg provides types and utilities for terminal package management.
//
// Terminal Packages allow grouping related terminal configuration into reusable bundles with single inheritance.
// For example, a "dev-essentials" package might extend "core" and include zsh plugins, prompts, and profiles.
//
// YAML Example:
//
//	apiVersion: devopsmaestro.io/v1
//	kind: TerminalPackage
//	metadata:
//	  name: dev-essentials
//	  description: "Developer essentials terminal setup"
//	  category: development
//	  tags: [zsh, starship]
//	spec:
//	  extends: core
//	  plugins: [zsh-autosuggestions, zsh-syntax-highlighting]
//	  prompts: [starship-minimal]
//	  profiles: [developer]
//
// This package has ZERO dependencies on dvm internals and can be used standalone.
package pkg

import (
	"time"
)

// StringOrSlice handles YAML fields that can be either a string or []string.
type StringOrSlice []string

// Package represents a collection of terminal configuration with metadata.
// This is the canonical type used throughout terminal-manager - no database types.
type Package struct {
	// Core identification
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Metadata
	Category string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`

	// Package structure
	Extends  string   `json:"extends,omitempty" yaml:"extends,omitempty"`   // Single inheritance
	Plugins  []string `json:"plugins,omitempty" yaml:"plugins,omitempty"`   // Shell plugin names
	Prompts  []string `json:"prompts,omitempty" yaml:"prompts,omitempty"`   // Prompt names
	Profiles []string `json:"profiles,omitempty" yaml:"profiles,omitempty"` // Profile preset names

	// Optional embedded configuration
	WezTerm *WezTermConfig `json:"wezterm,omitempty" yaml:"wezterm,omitempty"`

	// State
	Enabled bool `json:"enabled" yaml:"enabled"`

	// Timestamps (optional, used when stored)
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"-"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"-"`
}

// WezTermConfig represents optional WezTerm configuration embedded in a package.
type WezTermConfig struct {
	FontSize    int    `json:"fontSize,omitempty" yaml:"fontSize,omitempty"`
	ColorScheme string `json:"colorScheme,omitempty" yaml:"colorScheme,omitempty"`
	FontFamily  string `json:"fontFamily,omitempty" yaml:"fontFamily,omitempty"`
}

// PackageYAML represents the full YAML document format (kubectl-style).
// This is the user-facing format for package definition files.
type PackageYAML struct {
	APIVersion string          `yaml:"apiVersion"`
	Kind       string          `yaml:"kind"`
	Metadata   PackageMetadata `yaml:"metadata"`
	Spec       PackageSpec     `yaml:"spec"`
}

// PackageMetadata contains package metadata in the YAML format.
type PackageMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// PackageSpec contains the package specification in the YAML format.
type PackageSpec struct {
	Extends  string         `yaml:"extends,omitempty"`
	Plugins  StringOrSlice  `yaml:"plugins,omitempty"`
	Prompts  StringOrSlice  `yaml:"prompts,omitempty"`
	Profiles StringOrSlice  `yaml:"profiles,omitempty"`
	WezTerm  *WezTermConfig `yaml:"wezterm,omitempty"`
	Enabled  *bool          `yaml:"enabled,omitempty"` // Pointer to distinguish unset from false
}

// NewPackage creates a new Package with default values.
func NewPackage(name string) *Package {
	return &Package{
		Name:     name,
		Plugins:  make([]string, 0),
		Prompts:  make([]string, 0),
		Profiles: make([]string, 0),
		Enabled:  true,
	}
}

// NewPackageYAML creates a new PackageYAML with proper defaults.
func NewPackageYAML(name string) *PackageYAML {
	return &PackageYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalPackage",
		Metadata: PackageMetadata{
			Name: name,
		},
		Spec: PackageSpec{
			Plugins:  StringOrSlice{},
			Prompts:  StringOrSlice{},
			Profiles: StringOrSlice{},
		},
	}
}

// ToPackage converts PackageYAML to the canonical Package type.
func (py *PackageYAML) ToPackage() *Package {
	// Default to enabled unless explicitly set to false
	enabled := true
	if py.Spec.Enabled != nil {
		enabled = *py.Spec.Enabled
	}

	p := &Package{
		Name:        py.Metadata.Name,
		Description: py.Metadata.Description,
		Category:    py.Metadata.Category,
		Tags:        py.Metadata.Tags,
		Extends:     py.Spec.Extends,
		Plugins:     []string(py.Spec.Plugins),
		Prompts:     []string(py.Spec.Prompts),
		Profiles:    []string(py.Spec.Profiles),
		WezTerm:     py.Spec.WezTerm,
		Enabled:     enabled,
	}

	return p
}

// ToYAML converts a Package to the PackageYAML format.
func (p *Package) ToYAML() *PackageYAML {
	// Only include enabled field if disabled (to avoid cluttering YAML)
	var enabledPtr *bool
	if !p.Enabled {
		enabledPtr = &p.Enabled
	}

	py := &PackageYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalPackage",
		Metadata: PackageMetadata{
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
			Tags:        p.Tags,
		},
		Spec: PackageSpec{
			Extends:  p.Extends,
			Plugins:  StringOrSlice(p.Plugins),
			Prompts:  StringOrSlice(p.Prompts),
			Profiles: StringOrSlice(p.Profiles),
			WezTerm:  p.WezTerm,
			Enabled:  enabledPtr,
		},
	}

	return py
}
