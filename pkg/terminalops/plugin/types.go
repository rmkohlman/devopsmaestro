// Package plugin provides types and utilities for shell plugin management.
// This package has ZERO dependencies on dvm internals and can be used standalone.
package plugin

import (
	"time"
)

// PluginManager identifies the shell plugin manager to use.
type PluginManager string

const (
	// PluginManagerZinit is the Zinit plugin manager (fast, flexible)
	PluginManagerZinit PluginManager = "zinit"
	// PluginManagerOhMyZsh is Oh My Zsh framework
	PluginManagerOhMyZsh PluginManager = "oh-my-zsh"
	// PluginManagerAntigen is the Antigen plugin manager
	PluginManagerAntigen PluginManager = "antigen"
	// PluginManagerManual is manual git clone installation
	PluginManagerManual PluginManager = "manual"
	// PluginManagerSheldon is the Sheldon plugin manager (Rust-based)
	PluginManagerSheldon PluginManager = "sheldon"
)

// LoadMode specifies when/how the plugin should be loaded.
type LoadMode string

const (
	// LoadModeImmediate loads the plugin immediately at shell startup
	LoadModeImmediate LoadMode = "immediate"
	// LoadModeDeferred loads the plugin after prompt is displayed (faster startup)
	LoadModeDeferred LoadMode = "deferred"
	// LoadModeLazy loads the plugin only when first used
	LoadModeLazy LoadMode = "lazy"
)

// Plugin represents a shell plugin configuration.
// This is the canonical type used throughout terminal-manager.
type Plugin struct {
	// Core identification
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Source - where to get the plugin
	Repo   string `json:"repo,omitempty" yaml:"repo,omitempty"`     // GitHub repo (e.g., "zsh-users/zsh-autosuggestions")
	Source string `json:"source,omitempty" yaml:"source,omitempty"` // Full URL or local path
	Branch string `json:"branch,omitempty" yaml:"branch,omitempty"`
	Tag    string `json:"tag,omitempty" yaml:"tag,omitempty"`

	// Loading configuration
	Manager  PluginManager `json:"manager,omitempty" yaml:"manager,omitempty"`
	LoadMode LoadMode      `json:"load_mode,omitempty" yaml:"load_mode,omitempty"`

	// For oh-my-zsh plugins that are built-in
	OhMyZshPlugin string `json:"ohmyzsh_plugin,omitempty" yaml:"ohmyzsh_plugin,omitempty"`

	// Files to source (if not auto-detected)
	SourceFiles []string `json:"source_files,omitempty" yaml:"source_files,omitempty"`

	// Configuration to run after loading
	Config string `json:"config,omitempty" yaml:"config,omitempty"`

	// Environment variables to set
	Env map[string]string `json:"env,omitempty" yaml:"env,omitempty"`

	// Dependencies (other plugins that must be loaded first)
	Dependencies []string `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`

	// Metadata
	Category string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Enabled  bool     `json:"enabled" yaml:"enabled"`
	Priority int      `json:"priority,omitempty" yaml:"priority,omitempty"` // Lower = load first

	// Timestamps (optional, used when stored)
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"-"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"-"`
}

// PluginYAML represents the full YAML document format (kubectl-style).
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
	Repo          string            `yaml:"repo,omitempty"`
	Source        string            `yaml:"source,omitempty"`
	Branch        string            `yaml:"branch,omitempty"`
	Tag           string            `yaml:"tag,omitempty"`
	Manager       PluginManager     `yaml:"manager,omitempty"`
	LoadMode      LoadMode          `yaml:"loadMode,omitempty"`
	OhMyZshPlugin string            `yaml:"ohmyzshPlugin,omitempty"`
	SourceFiles   []string          `yaml:"sourceFiles,omitempty"`
	Config        string            `yaml:"config,omitempty"`
	Env           map[string]string `yaml:"env,omitempty"`
	Dependencies  []string          `yaml:"dependencies,omitempty"`
	Priority      int               `yaml:"priority,omitempty"`
	Enabled       *bool             `yaml:"enabled,omitempty"`
}

// NewPlugin creates a new Plugin with default values.
func NewPlugin(name, repo string) *Plugin {
	return &Plugin{
		Name:     name,
		Repo:     repo,
		Manager:  PluginManagerManual,
		LoadMode: LoadModeImmediate,
		Enabled:  true,
	}
}

// NewPluginYAML creates a new PluginYAML with proper defaults.
func NewPluginYAML(name, repo string) *PluginYAML {
	return &PluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalPlugin",
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
	enabled := true
	if py.Spec.Enabled != nil {
		enabled = *py.Spec.Enabled
	}

	manager := py.Spec.Manager
	if manager == "" {
		manager = PluginManagerManual
	}

	loadMode := py.Spec.LoadMode
	if loadMode == "" {
		loadMode = LoadModeImmediate
	}

	return &Plugin{
		Name:          py.Metadata.Name,
		Description:   py.Metadata.Description,
		Category:      py.Metadata.Category,
		Tags:          py.Metadata.Tags,
		Repo:          py.Spec.Repo,
		Source:        py.Spec.Source,
		Branch:        py.Spec.Branch,
		Tag:           py.Spec.Tag,
		Manager:       manager,
		LoadMode:      loadMode,
		OhMyZshPlugin: py.Spec.OhMyZshPlugin,
		SourceFiles:   py.Spec.SourceFiles,
		Config:        py.Spec.Config,
		Env:           py.Spec.Env,
		Dependencies:  py.Spec.Dependencies,
		Priority:      py.Spec.Priority,
		Enabled:       enabled,
	}
}

// ToYAML converts a Plugin to the PluginYAML format.
func (p *Plugin) ToYAML() *PluginYAML {
	var enabledPtr *bool
	if !p.Enabled {
		enabledPtr = &p.Enabled
	}

	return &PluginYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalPlugin",
		Metadata: PluginMetadata{
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
			Tags:        p.Tags,
		},
		Spec: PluginSpec{
			Repo:          p.Repo,
			Source:        p.Source,
			Branch:        p.Branch,
			Tag:           p.Tag,
			Manager:       p.Manager,
			LoadMode:      p.LoadMode,
			OhMyZshPlugin: p.OhMyZshPlugin,
			SourceFiles:   p.SourceFiles,
			Config:        p.Config,
			Env:           p.Env,
			Dependencies:  p.Dependencies,
			Priority:      p.Priority,
			Enabled:       enabledPtr,
		},
	}
}

// IsOhMyZshBuiltin returns true if this is a built-in oh-my-zsh plugin.
func (p *Plugin) IsOhMyZshBuiltin() bool {
	return p.OhMyZshPlugin != "" && p.Repo == ""
}

// GetSourceURL returns the full source URL for the plugin.
func (p *Plugin) GetSourceURL() string {
	if p.Source != "" {
		return p.Source
	}
	if p.Repo != "" {
		return "https://github.com/" + p.Repo
	}
	return ""
}
