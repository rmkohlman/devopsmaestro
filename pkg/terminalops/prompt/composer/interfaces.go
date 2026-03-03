// Package composer provides interfaces for composing prompts.
package composer

import (
	"devopsmaestro/pkg/terminalops/prompt/extension"
	"devopsmaestro/pkg/terminalops/prompt/style"
)

// PromptComposer composes a prompt from a style and extensions.
type PromptComposer interface {
	// Compose creates a ComposedPrompt from a style and extensions.
	Compose(promptStyle *style.PromptStyle, extensions []*extension.PromptExtension) (*ComposedPrompt, error)
}

// ComposedPrompt represents the result of composing a prompt style with extensions.
type ComposedPrompt struct {
	// Format is the starship format string
	Format string

	// Modules contains the module configurations
	Modules map[string]ModuleConfig

	// Palette is the palette name reference
	Palette string
}

// ModuleConfig represents a module configuration in the composed prompt.
type ModuleConfig struct {
	Disabled bool           `json:"disabled,omitempty"`
	Symbol   string         `json:"symbol,omitempty"`
	Format   string         `json:"format,omitempty"`
	Style    string         `json:"style,omitempty"`
	Options  map[string]any `json:"options,omitempty"`
}
