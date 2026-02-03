// Package prompt provides types and utilities for terminal prompt management.
// This package has ZERO dependencies on dvm internals and can be used standalone.
package prompt

import (
	"time"
)

// PromptType identifies the type of terminal prompt system.
type PromptType string

const (
	// PromptTypeStarship is the Starship cross-shell prompt
	PromptTypeStarship PromptType = "starship"
	// PromptTypePowerlevel10k is the Powerlevel10k Zsh prompt
	PromptTypePowerlevel10k PromptType = "powerlevel10k"
	// PromptTypeOhMyPosh is the Oh My Posh cross-platform prompt
	PromptTypeOhMyPosh PromptType = "oh-my-posh"
)

// Prompt represents a terminal prompt configuration.
// This is the canonical type used throughout terminal-manager.
type Prompt struct {
	// Core identification
	Name        string     `json:"name" yaml:"name"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Type        PromptType `json:"type" yaml:"type"`

	// Prompt format string (Starship/Oh-My-Posh format)
	Format string `json:"format,omitempty" yaml:"format,omitempty"`

	// Module configurations
	Modules map[string]ModuleConfig `json:"modules,omitempty" yaml:"modules,omitempty"`

	// Character/prompt symbol configuration
	Character *CharacterConfig `json:"character,omitempty" yaml:"character,omitempty"`

	// Color palette reference (can link to nvim theme)
	PaletteRef string `json:"palette_ref,omitempty" yaml:"palette_ref,omitempty"`

	// Custom color overrides
	Colors map[string]string `json:"colors,omitempty" yaml:"colors,omitempty"`

	// Raw config (for advanced users who want to specify raw config)
	RawConfig string `json:"raw_config,omitempty" yaml:"raw_config,omitempty"`

	// Metadata
	Category string   `json:"category,omitempty" yaml:"category,omitempty"`
	Tags     []string `json:"tags,omitempty" yaml:"tags,omitempty"`
	Enabled  bool     `json:"enabled" yaml:"enabled"`

	// Timestamps (optional, used when stored)
	CreatedAt *time.Time `json:"created_at,omitempty" yaml:"-"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" yaml:"-"`
}

// ModuleConfig represents configuration for a prompt module.
// Different modules have different available options.
type ModuleConfig struct {
	// Disabled completely disables the module
	Disabled bool `json:"disabled,omitempty" yaml:"disabled,omitempty"`

	// Format string for the module
	Format string `json:"format,omitempty" yaml:"format,omitempty"`

	// Style for the module (color/formatting)
	Style string `json:"style,omitempty" yaml:"style,omitempty"`

	// Symbol to display
	Symbol string `json:"symbol,omitempty" yaml:"symbol,omitempty"`

	// Additional options (module-specific)
	Options map[string]any `json:"options,omitempty" yaml:"options,omitempty"`
}

// CharacterConfig configures the prompt character/symbol.
type CharacterConfig struct {
	// SuccessSymbol is shown when the last command succeeded
	SuccessSymbol string `json:"success_symbol,omitempty" yaml:"success_symbol,omitempty"`
	// ErrorSymbol is shown when the last command failed
	ErrorSymbol string `json:"error_symbol,omitempty" yaml:"error_symbol,omitempty"`
	// ViCmdSymbol is shown in vi command mode
	ViCmdSymbol string `json:"vicmd_symbol,omitempty" yaml:"vicmd_symbol,omitempty"`
	// ViInsSymbol is shown in vi insert mode
	ViInsSymbol string `json:"viins_symbol,omitempty" yaml:"viins_symbol,omitempty"`
}

// PromptYAML represents the full YAML document format (kubectl-style).
// This is the user-facing format for prompt definition files.
type PromptYAML struct {
	APIVersion string         `yaml:"apiVersion"`
	Kind       string         `yaml:"kind"`
	Metadata   PromptMetadata `yaml:"metadata"`
	Spec       PromptSpec     `yaml:"spec"`
}

// PromptMetadata contains prompt metadata in the YAML format.
type PromptMetadata struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Category    string            `yaml:"category,omitempty"`
	Tags        []string          `yaml:"tags,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

// PromptSpec contains the prompt specification in the YAML format.
type PromptSpec struct {
	Type       PromptType              `yaml:"type"`
	Format     string                  `yaml:"format,omitempty"`
	Modules    map[string]ModuleConfig `yaml:"modules,omitempty"`
	Character  *CharacterConfig        `yaml:"character,omitempty"`
	PaletteRef string                  `yaml:"paletteRef,omitempty"`
	Colors     map[string]string       `yaml:"colors,omitempty"`
	RawConfig  string                  `yaml:"rawConfig,omitempty"`
	Enabled    *bool                   `yaml:"enabled,omitempty"`
}

// NewPrompt creates a new Prompt with default values.
func NewPrompt(name string, promptType PromptType) *Prompt {
	return &Prompt{
		Name:    name,
		Type:    promptType,
		Enabled: true,
	}
}

// NewPromptYAML creates a new PromptYAML with proper defaults.
func NewPromptYAML(name string, promptType PromptType) *PromptYAML {
	return &PromptYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalPrompt",
		Metadata: PromptMetadata{
			Name: name,
		},
		Spec: PromptSpec{
			Type: promptType,
		},
	}
}

// ToPrompt converts PromptYAML to the canonical Prompt type.
func (py *PromptYAML) ToPrompt() *Prompt {
	// Default to enabled unless explicitly set to false
	enabled := true
	if py.Spec.Enabled != nil {
		enabled = *py.Spec.Enabled
	}

	p := &Prompt{
		Name:        py.Metadata.Name,
		Description: py.Metadata.Description,
		Category:    py.Metadata.Category,
		Tags:        py.Metadata.Tags,
		Type:        py.Spec.Type,
		Format:      py.Spec.Format,
		Modules:     py.Spec.Modules,
		Character:   py.Spec.Character,
		PaletteRef:  py.Spec.PaletteRef,
		Colors:      py.Spec.Colors,
		RawConfig:   py.Spec.RawConfig,
		Enabled:     enabled,
	}

	return p
}

// ToYAML converts a Prompt to the PromptYAML format.
func (p *Prompt) ToYAML() *PromptYAML {
	// Only include enabled field if disabled (to avoid cluttering YAML)
	var enabledPtr *bool
	if !p.Enabled {
		enabledPtr = &p.Enabled
	}

	py := &PromptYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "TerminalPrompt",
		Metadata: PromptMetadata{
			Name:        p.Name,
			Description: p.Description,
			Category:    p.Category,
			Tags:        p.Tags,
		},
		Spec: PromptSpec{
			Type:       p.Type,
			Format:     p.Format,
			Modules:    p.Modules,
			Character:  p.Character,
			PaletteRef: p.PaletteRef,
			Colors:     p.Colors,
			RawConfig:  p.RawConfig,
			Enabled:    enabledPtr,
		},
	}

	return py
}

// IsStarship returns true if this is a Starship prompt.
func (p *Prompt) IsStarship() bool {
	return p.Type == PromptTypeStarship
}

// IsPowerlevel10k returns true if this is a Powerlevel10k prompt.
func (p *Prompt) IsPowerlevel10k() bool {
	return p.Type == PromptTypePowerlevel10k
}

// IsOhMyPosh returns true if this is an Oh My Posh prompt.
func (p *Prompt) IsOhMyPosh() bool {
	return p.Type == PromptTypeOhMyPosh
}
