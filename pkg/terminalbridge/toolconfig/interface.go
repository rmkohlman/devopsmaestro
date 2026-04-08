// Package toolconfig provides config generators for terminal tools (bat, delta,
// fzf, dircolors) that produce tool-specific config snippets from a workspace
// theme's color palette. This ensures visual consistency across all terminal tools.
package toolconfig

import (
	"github.com/rmkohlman/MaestroPalette"
)

// ToolConfigGenerator defines the interface for generating tool-specific
// configuration from a color palette. Each implementation maps palette colors
// to the target tool's configuration format.
type ToolConfigGenerator interface {
	// Generate produces a config snippet for the tool using the given palette.
	// Returns the config string (env var value, config file content, etc.).
	Generate(pal *palette.Palette) (string, error)

	// ToolName returns the name of the tool this generator targets.
	ToolName() string

	// Description returns a short description of what this generator produces.
	Description() string
}

// OutputFormat controls how the generated config is formatted.
type OutputFormat string

const (
	// FormatEnv produces shell export statements (e.g., export FOO="bar").
	FormatEnv OutputFormat = "env"
	// FormatRaw produces just the config value without shell wrapping.
	FormatRaw OutputFormat = "raw"
)

// paletteGet safely retrieves a color from the palette with a fallback default.
func paletteGet(pal *palette.Palette, key, defaultColor string) string {
	if pal == nil {
		return defaultColor
	}
	return pal.GetOrDefault(key, defaultColor)
}
