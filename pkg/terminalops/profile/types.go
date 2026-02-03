// Package profile provides the TerminalProfile type that aggregates
// prompt, plugins, and shell configuration into a complete terminal setup.
package profile

import (
	"fmt"
	"time"

	"devopsmaestro/pkg/terminalops/plugin"
	"devopsmaestro/pkg/terminalops/prompt"
	"devopsmaestro/pkg/terminalops/shell"
)

// Profile represents a complete terminal configuration.
// It aggregates prompt, plugins, and shell settings into a single deployable unit.
type Profile struct {
	// Core identification
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	// Components (can be inline or references)
	Prompt  *prompt.Prompt   `json:"prompt,omitempty" yaml:"prompt,omitempty"`
	Plugins []*plugin.Plugin `json:"plugins,omitempty" yaml:"plugins,omitempty"`
	Shell   *shell.Shell     `json:"shell,omitempty" yaml:"shell,omitempty"`

	// References to library items (alternative to inline)
	PromptRef  string   `json:"prompt_ref,omitempty" yaml:"prompt_ref,omitempty"`
	PluginRefs []string `json:"plugin_refs,omitempty" yaml:"plugin_refs,omitempty"`
	ShellRef   string   `json:"shell_ref,omitempty" yaml:"shell_ref,omitempty"`

	// Theme reference (for color coordination with nvim)
	ThemeRef string `json:"theme_ref,omitempty" yaml:"theme_ref,omitempty"`

	// Metadata
	Category string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Enabled  bool     `json:"enabled" yaml:"enabled"`

	// Timestamps
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"-"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"-"`
}

// ProfileYAML represents the full YAML document format (kubectl-style).
type ProfileYAML struct {
	APIVersion string          `yaml:"apiVersion"`
	Kind       string          `yaml:"kind"`
	Metadata   ProfileMetadata `yaml:"metadata"`
	Spec       ProfileSpec     `yaml:"spec"`
}

// ProfileMetadata contains profile metadata.
type ProfileMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// ProfileSpec contains the profile specification.
type ProfileSpec struct {
	// Inline configurations
	Prompt  *prompt.PromptSpec  `yaml:"prompt,omitempty"`
	Plugins []plugin.PluginSpec `yaml:"plugins,omitempty"`
	Shell   *shell.ShellSpec    `yaml:"shell,omitempty"`

	// References to library items
	PromptRef  string   `yaml:"promptRef,omitempty"`
	PluginRefs []string `yaml:"pluginRefs,omitempty"`
	ShellRef   string   `yaml:"shellRef,omitempty"`

	// Theme coordination
	ThemeRef string `yaml:"themeRef,omitempty"`

	Enabled *bool `yaml:"enabled,omitempty"`
}

// NewProfile creates a new Profile with default values.
func NewProfile(name string) *Profile {
	return &Profile{
		Name:    name,
		Enabled: true,
	}
}

// NewProfileYAML creates a new ProfileYAML with proper defaults.
func NewProfileYAML(name string) *ProfileYAML {
	return &ProfileYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalProfile",
		Metadata: ProfileMetadata{
			Name: name,
		},
		Spec: ProfileSpec{},
	}
}

// ToProfile converts ProfileYAML to the canonical Profile type.
// Note: This only converts inline specs. References need to be resolved separately.
func (py *ProfileYAML) ToProfile() *Profile {
	enabled := true
	if py.Spec.Enabled != nil {
		enabled = *py.Spec.Enabled
	}

	p := &Profile{
		Name:        py.Metadata.Name,
		Description: py.Metadata.Description,
		Category:    py.Metadata.Category,
		Tags:        py.Metadata.Tags,
		PromptRef:   py.Spec.PromptRef,
		PluginRefs:  py.Spec.PluginRefs,
		ShellRef:    py.Spec.ShellRef,
		ThemeRef:    py.Spec.ThemeRef,
		Enabled:     enabled,
	}

	// Convert inline prompt spec if present
	if py.Spec.Prompt != nil {
		promptYAML := &prompt.PromptYAML{
			APIVersion: "devopsmaestro.io/v1",
			Kind:       "TerminalPrompt",
			Metadata: prompt.PromptMetadata{
				Name: py.Metadata.Name + "-prompt",
			},
			Spec: *py.Spec.Prompt,
		}
		p.Prompt = promptYAML.ToPrompt()
	}

	// Convert inline plugin specs if present
	for i, pluginSpec := range py.Spec.Plugins {
		pluginYAML := &plugin.PluginYAML{
			APIVersion: "devopsmaestro.io/v1",
			Kind:       "TerminalPlugin",
			Metadata: plugin.PluginMetadata{
				Name: pluginSpec.Repo, // Use repo as default name
			},
			Spec: pluginSpec,
		}
		if pluginYAML.Metadata.Name == "" {
			pluginYAML.Metadata.Name = fmt.Sprintf("%s-plugin-%d", py.Metadata.Name, i)
		}
		p.Plugins = append(p.Plugins, pluginYAML.ToPlugin())
	}

	// Convert inline shell spec if present
	if py.Spec.Shell != nil {
		shellYAML := &shell.ShellYAML{
			APIVersion: "devopsmaestro.io/v1",
			Kind:       "TerminalShell",
			Metadata: shell.ShellMetadata{
				Name: py.Metadata.Name + "-shell",
			},
			Spec: *py.Spec.Shell,
		}
		p.Shell = shellYAML.ToShell()
	}

	return p
}

// ToYAML converts a Profile to the ProfileYAML format.
func (p *Profile) ToYAML() *ProfileYAML {
	var enabledPtr *bool
	if !p.Enabled {
		enabledPtr = &p.Enabled
	}

	py := &ProfileYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalProfile",
		Metadata: ProfileMetadata{
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
			Tags:        p.Tags,
		},
		Spec: ProfileSpec{
			PromptRef:  p.PromptRef,
			PluginRefs: p.PluginRefs,
			ShellRef:   p.ShellRef,
			ThemeRef:   p.ThemeRef,
			Enabled:    enabledPtr,
		},
	}

	// Convert inline prompt if present
	if p.Prompt != nil {
		promptYAML := p.Prompt.ToYAML()
		py.Spec.Prompt = &promptYAML.Spec
	}

	// Convert inline plugins if present
	for _, plug := range p.Plugins {
		pluginYAML := plug.ToYAML()
		py.Spec.Plugins = append(py.Spec.Plugins, pluginYAML.Spec)
	}

	// Convert inline shell if present
	if p.Shell != nil {
		shellYAML := p.Shell.ToYAML()
		py.Spec.Shell = &shellYAML.Spec
	}

	return py
}

// HasInlineConfig returns true if the profile has any inline configuration.
func (p *Profile) HasInlineConfig() bool {
	return p.Prompt != nil || len(p.Plugins) > 0 || p.Shell != nil
}

// HasReferences returns true if the profile has any library references.
func (p *Profile) HasReferences() bool {
	return p.PromptRef != "" || len(p.PluginRefs) > 0 || p.ShellRef != ""
}
