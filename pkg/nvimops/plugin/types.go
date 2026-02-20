// Package plugin provides types and utilities for Neovim plugin management.
// This package has ZERO dependencies on dvm internals and can be used standalone.
package plugin

import (
	"time"
)

// Plugin represents a Neovim plugin with all its configuration.
// This is the canonical type used throughout nvim-manager - no database types.
type Plugin struct {
	// Core identification
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Repo        string `json:"repo" yaml:"repo"`

	// Version control
	Branch  string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`

	// Lazy loading configuration
	Priority int      `json:"priority,omitempty" yaml:"priority,omitempty"`
	Lazy     bool     `json:"lazy,omitempty" yaml:"lazy,omitempty"`
	Event    []string `json:"event,omitempty" yaml:"event,omitempty"`
	Ft       []string `json:"ft,omitempty" yaml:"ft,omitempty"`
	Cmd      []string `json:"cmd,omitempty" yaml:"cmd,omitempty"`
	Keys     []Keymap `json:"keys,omitempty" yaml:"keys,omitempty"`

	// Dependencies
	Dependencies []Dependency `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`

	// Build and configuration
	Build  string      `json:"build,omitempty" yaml:"build,omitempty"`
	Config string      `json:"config,omitempty" yaml:"config,omitempty"` // Lua code
	Init   string      `json:"init,omitempty" yaml:"init,omitempty"`     // Lua code
	Opts   interface{} `json:"opts,omitempty" yaml:"opts,omitempty"`

	// Additional keymaps (separate from lazy keys)
	Keymaps []Keymap `json:"keymaps,omitempty" yaml:"keymaps,omitempty"`

	// Metadata
	Category string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Enabled  bool     `json:"enabled" yaml:"enabled"`

	// Timestamps (optional, used when stored)
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"-"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"-"`
}

// Dependency represents a plugin dependency.
type Dependency struct {
	Repo    string `json:"repo" yaml:"repo"`
	Build   string `json:"build,omitempty" yaml:"build,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Branch  string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Config  bool   `json:"config,omitempty" yaml:"config,omitempty"`
}

// Keymap represents a key mapping for a plugin.
type Keymap struct {
	Key    string   `json:"key" yaml:"key"`
	Mode   []string `json:"mode,omitempty" yaml:"mode,omitempty"` // Always a slice internally
	Action string   `json:"action,omitempty" yaml:"action,omitempty"`
	Desc   string   `json:"desc,omitempty" yaml:"desc,omitempty"`
}

// PluginYAML represents the full YAML document format (kubectl-style).
// This is the user-facing format for plugin definition files.
type PluginYAML struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   PluginMetadata `yaml:"metadata"`
	Spec       PluginSpec     `yaml:"spec"`
}

// PluginMetadata contains plugin metadata in the YAML format.
type PluginMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// PluginSpec contains the plugin specification in the YAML format.
type PluginSpec struct {
	Repo         string           `yaml:"repo"`
	Branch       string           `yaml:"branch,omitempty"`
	Version      string           `yaml:"version,omitempty"`
	Priority     int              `yaml:"priority,omitempty"`
	Lazy         bool             `yaml:"lazy,omitempty"`
	Enabled      *bool            `yaml:"enabled,omitempty"` // Pointer to distinguish unset from false
	Event        StringOrSlice    `yaml:"event,omitempty"`
	Ft           StringOrSlice    `yaml:"ft,omitempty"`
	Cmd          StringOrSlice    `yaml:"cmd,omitempty"`
	Keys         []KeymapYAML     `yaml:"keys,omitempty"`
	Dependencies []DependencyYAML `yaml:"dependencies,omitempty"`
	Build        string           `yaml:"build,omitempty"`
	Config       string           `yaml:"config,omitempty"`
	Init         string           `yaml:"init,omitempty"`
	Opts         interface{}      `yaml:"opts,omitempty"`
	Keymaps      []KeymapYAML     `yaml:"keymaps,omitempty"`
}

// KeymapYAML represents a keymap in YAML format (flexible mode field).
type KeymapYAML struct {
	Key    string        `yaml:"key"`
	Mode   StringOrSlice `yaml:"mode,omitempty"`
	Action string        `yaml:"action,omitempty"`
	Desc   string        `yaml:"desc,omitempty"`
}

// DependencyYAML can be either a string (repo only) or a full dependency spec.
// This is handled by custom unmarshaling.
type DependencyYAML struct {
	Repo    string `yaml:"repo,omitempty"`
	Build   string `yaml:"build,omitempty"`
	Version string `yaml:"version,omitempty"`
	Branch  string `yaml:"branch,omitempty"`
	Config  bool   `yaml:"config,omitempty"`
}

// StringOrSlice handles YAML fields that can be either a string or []string.
type StringOrSlice []string

// NewPlugin creates a new Plugin with default values.
func NewPlugin(name, repo string) *Plugin {
	return &Plugin{
		Name:    name,
		Repo:    repo,
		Enabled: true,
	}
}

// NewPluginYAML creates a new PluginYAML with proper defaults.
func NewPluginYAML(name, repo string) *PluginYAML {
	return &PluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPlugin",
		Metadata: PluginMetadata{
			Name: name,
		},
		Spec: PluginSpec{
			Repo: repo,
		},
	}
}

// ToPlugin converts PluginYAML to the canonical Plugin type.
func (py *PluginYAML) ToPlugin() *Plugin {
	// Default to enabled unless explicitly set to false
	enabled := true
	if py.Spec.Enabled != nil {
		enabled = *py.Spec.Enabled
	}

	p := &Plugin{
		Name:        py.Metadata.Name,
		Description: py.Metadata.Description,
		Category:    py.Metadata.Category,
		Tags:        py.Metadata.Tags,
		Repo:        py.Spec.Repo,
		Branch:      py.Spec.Branch,
		Version:     py.Spec.Version,
		Priority:    py.Spec.Priority,
		Lazy:        py.Spec.Lazy,
		Event:       []string(py.Spec.Event),
		Ft:          []string(py.Spec.Ft),
		Cmd:         []string(py.Spec.Cmd),
		Build:       py.Spec.Build,
		Config:      py.Spec.Config,
		Init:        py.Spec.Init,
		Opts:        py.Spec.Opts,
		Enabled:     enabled,
	}

	// Convert keys
	for _, k := range py.Spec.Keys {
		p.Keys = append(p.Keys, Keymap{
			Key:    k.Key,
			Mode:   []string(k.Mode),
			Action: k.Action,
			Desc:   k.Desc,
		})
	}

	// Convert keymaps
	for _, k := range py.Spec.Keymaps {
		p.Keymaps = append(p.Keymaps, Keymap{
			Key:    k.Key,
			Mode:   []string(k.Mode),
			Action: k.Action,
			Desc:   k.Desc,
		})
	}

	// Convert dependencies
	for _, d := range py.Spec.Dependencies {
		p.Dependencies = append(p.Dependencies, Dependency{
			Repo:    d.Repo,
			Build:   d.Build,
			Version: d.Version,
			Branch:  d.Branch,
			Config:  d.Config,
		})
	}

	return p
}

// ToYAML converts a Plugin to the PluginYAML format.
func (p *Plugin) ToYAML() *PluginYAML {
	// Only include enabled field if disabled (to avoid cluttering YAML)
	var enabledPtr *bool
	if !p.Enabled {
		enabledPtr = &p.Enabled
	}

	py := &PluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimPlugin",
		Metadata: PluginMetadata{
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
			Tags:        p.Tags,
		},
		Spec: PluginSpec{
			Repo:     p.Repo,
			Branch:   p.Branch,
			Version:  p.Version,
			Priority: p.Priority,
			Lazy:     p.Lazy,
			Enabled:  enabledPtr,
			Event:    StringOrSlice(p.Event),
			Ft:       StringOrSlice(p.Ft),
			Cmd:      StringOrSlice(p.Cmd),
			Build:    p.Build,
			Config:   p.Config,
			Init:     p.Init,
			Opts:     p.Opts,
		},
	}

	// Convert keys
	for _, k := range p.Keys {
		py.Spec.Keys = append(py.Spec.Keys, KeymapYAML{
			Key:    k.Key,
			Mode:   StringOrSlice(k.Mode),
			Action: k.Action,
			Desc:   k.Desc,
		})
	}

	// Convert keymaps
	for _, k := range p.Keymaps {
		py.Spec.Keymaps = append(py.Spec.Keymaps, KeymapYAML{
			Key:    k.Key,
			Mode:   StringOrSlice(k.Mode),
			Action: k.Action,
			Desc:   k.Desc,
		})
	}

	// Convert dependencies
	for _, d := range p.Dependencies {
		py.Spec.Dependencies = append(py.Spec.Dependencies, DependencyYAML{
			Repo:    d.Repo,
			Build:   d.Build,
			Version: d.Version,
			Branch:  d.Branch,
			Config:  d.Config,
		})
	}

	return py
}
