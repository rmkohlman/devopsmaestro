// Package prompt provides types and utilities for terminal prompt management.
package prompt

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"devopsmaestro/pkg/palette"
)

// PromptRenderer converts TerminalPrompt definitions to config files.
type PromptRenderer interface {
	// Render generates config file content from a PromptYAML
	// using the provided palette for ${theme.X} variable resolution.
	Render(prompt *PromptYAML, pal *palette.Palette) (string, error)

	// RenderToFile renders and writes to a file path.
	RenderToFile(prompt *PromptYAML, pal *palette.Palette, path string) error
}

// StarshipRenderer converts TerminalPrompt definitions to starship.toml configuration.
type StarshipRenderer struct {
	// IndentSize is the number of spaces per indentation level (default: 2)
	IndentSize int
}

// NewRenderer creates a new StarshipRenderer with default settings.
func NewRenderer() *StarshipRenderer {
	return &StarshipRenderer{
		IndentSize: 2,
	}
}

// NewStarshipRenderer creates a new StarshipRenderer with default settings.
// Alias for NewRenderer for explicit naming.
func NewStarshipRenderer() *StarshipRenderer {
	return NewRenderer()
}

// themeVarRegex matches ${theme.X} variables in strings.
var themeVarRegex = regexp.MustCompile(`\$\{theme\.([a-zA-Z_][a-zA-Z0-9_]*)\}`)

// themeVarMapping maps theme variable names to palette color keys.
// This allows users to write ${theme.red} and have it resolve to the correct palette key.
var themeVarMapping = map[string]string{
	// Background/foreground
	"bg": palette.ColorBg,
	"fg": palette.ColorFg,

	// Accent colors
	"primary":   palette.ColorPrimary,
	"secondary": palette.ColorSecondary,
	"accent":    palette.ColorAccent,

	// Standard colors (map to ANSI terminal colors)
	"red":     palette.TermRed,
	"green":   palette.TermGreen,
	"yellow":  palette.TermYellow,
	"blue":    palette.TermBlue,
	"magenta": palette.TermMagenta,
	"cyan":    palette.TermCyan,
	"black":   palette.TermBlack,
	"white":   palette.TermWhite,

	// Bright variants
	"bright_red":     palette.TermBrightRed,
	"bright_green":   palette.TermBrightGreen,
	"bright_yellow":  palette.TermBrightYellow,
	"bright_blue":    palette.TermBrightBlue,
	"bright_magenta": palette.TermBrightMagenta,
	"bright_cyan":    palette.TermBrightCyan,
	"bright_black":   palette.TermBrightBlack,
	"bright_white":   palette.TermBrightWhite,

	// Diagnostic colors
	"error":   palette.ColorError,
	"warning": palette.ColorWarning,
	"info":    palette.ColorInfo,
	"hint":    palette.ColorHint,
	"success": palette.ColorSuccess,

	// UI colors
	"comment": palette.ColorComment,
	"border":  palette.ColorBorder,

	// Catppuccin-style names (commonly used in starship prompts)
	// These map to semantic/ANSI colors for cross-theme compatibility
	"crust":     palette.ColorBg,     // Dark background / text-on-colored-bg
	"mantle":    palette.ColorBgDark, // Darker background
	"base":      palette.ColorBg,     // Base background
	"surface0":  palette.ColorBgHighlight,
	"surface1":  palette.ColorBgHighlight,
	"surface2":  palette.ColorBgHighlight,
	"text":      palette.ColorFg,      // Main text
	"subtext0":  palette.ColorComment, // Muted text
	"subtext1":  palette.ColorComment, // Muted text
	"overlay0":  palette.ColorBorder,
	"overlay1":  palette.ColorBorder,
	"overlay2":  palette.ColorBorder,
	"lavender":  palette.TermBrightMagenta, // Light purple/lavender
	"sky":       palette.TermCyan,          // Light blue/cyan
	"sapphire":  palette.TermBlue,          // Blue
	"teal":      palette.ColorTeal,         // Teal
	"peach":     palette.ColorOrange,       // Orange/peach
	"maroon":    palette.TermRed,           // Dark red
	"pink":      palette.ColorPink,         // Pink
	"mauve":     palette.TermMagenta,       // Mauve/purple
	"flamingo":  palette.ColorPink,         // Pink variant
	"rosewater": palette.ColorFg,           // Light pink/fg
}

// Render generates starship.toml content from a PromptYAML.
// Theme variables (${theme.X}) are resolved using the provided palette.
func (r *StarshipRenderer) Render(prompt *PromptYAML, pal *palette.Palette) (string, error) {
	if prompt == nil {
		return "", fmt.Errorf("prompt is nil")
	}
	if pal == nil {
		return "", fmt.Errorf("palette is nil")
	}

	var toml strings.Builder

	// Write header comment
	r.writeHeader(&toml, prompt, pal)

	// Write global settings
	r.writeGlobalSettings(&toml, prompt, pal)

	// Write palette section
	r.writePaletteSection(&toml, pal)

	// Write module sections
	r.writeModuleSections(&toml, prompt, pal)

	return toml.String(), nil
}

// RenderToFile renders and writes to a file path.
func (r *StarshipRenderer) RenderToFile(prompt *PromptYAML, pal *palette.Palette, path string) error {
	content, err := r.Render(prompt, pal)
	if err != nil {
		return fmt.Errorf("failed to render: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// writeHeader writes the TOML file header comment.
func (r *StarshipRenderer) writeHeader(toml *strings.Builder, prompt *PromptYAML, pal *palette.Palette) {
	toml.WriteString("# Generated by dvm - do not edit\n")
	toml.WriteString(fmt.Sprintf("# Prompt: %s\n", prompt.Metadata.Name))
	if prompt.Metadata.Description != "" {
		toml.WriteString(fmt.Sprintf("# Description: %s\n", prompt.Metadata.Description))
	}
	toml.WriteString(fmt.Sprintf("# Theme: %s\n", pal.Name))
	toml.WriteString("\n")
}

// writeGlobalSettings writes the global Starship settings.
func (r *StarshipRenderer) writeGlobalSettings(toml *strings.Builder, prompt *PromptYAML, pal *palette.Palette) {
	// Palette reference
	toml.WriteString(fmt.Sprintf("palette = '%s'\n", pal.Name))

	// Add newline setting
	if prompt.Spec.AddNewline {
		toml.WriteString("add_newline = true\n")
	} else {
		toml.WriteString("add_newline = false\n")
	}

	// Format string (with theme variables resolved)
	if prompt.Spec.Format != "" {
		resolvedFormat := r.resolveThemeVars(prompt.Spec.Format, pal)
		toml.WriteString(fmt.Sprintf("\nformat = \"\"\"\n%s\n\"\"\"\n", resolvedFormat))
	}

	toml.WriteString("\n")
}

// writePaletteSection writes the [palettes.X] section with all theme colors.
func (r *StarshipRenderer) writePaletteSection(toml *strings.Builder, pal *palette.Palette) {
	toml.WriteString(fmt.Sprintf("[palettes.%s]\n", pal.Name))

	// Get terminal colors from palette
	termColors := pal.ToTerminalColors()

	// Write colors in a consistent order
	colorKeys := make([]string, 0, len(termColors))
	for k := range termColors {
		colorKeys = append(colorKeys, k)
	}
	sort.Strings(colorKeys)

	for _, key := range colorKeys {
		if color := termColors[key]; color != "" {
			toml.WriteString(fmt.Sprintf("%s = \"%s\"\n", key, color))
		}
	}

	// Also add semantic colors from the palette
	semanticKeys := []string{
		palette.ColorBg, palette.ColorFg,
		palette.ColorPrimary, palette.ColorSecondary, palette.ColorAccent,
		palette.ColorError, palette.ColorWarning, palette.ColorInfo, palette.ColorHint, palette.ColorSuccess,
		palette.ColorComment, palette.ColorBorder,
	}

	for _, key := range semanticKeys {
		if color := pal.Get(key); color != "" {
			// Only write if not already written (check termColors)
			if _, exists := termColors[key]; !exists {
				toml.WriteString(fmt.Sprintf("%s = \"%s\"\n", key, color))
			}
		}
	}

	toml.WriteString("\n")
}

// writeModuleSections writes all module configuration sections.
func (r *StarshipRenderer) writeModuleSections(toml *strings.Builder, prompt *PromptYAML, pal *palette.Palette) {
	// Sort module names for consistent output
	moduleNames := make([]string, 0, len(prompt.Spec.Modules))
	for name := range prompt.Spec.Modules {
		moduleNames = append(moduleNames, name)
	}
	sort.Strings(moduleNames)

	for _, name := range moduleNames {
		module := prompt.Spec.Modules[name]
		r.writeModuleSection(toml, name, &module, pal)
	}
}

// writeModuleSection writes a single module configuration section.
func (r *StarshipRenderer) writeModuleSection(toml *strings.Builder, name string, module *ModuleConfig, pal *palette.Palette) {
	toml.WriteString(fmt.Sprintf("[%s]\n", name))

	// Write disabled
	if module.Disabled {
		toml.WriteString("disabled = true\n")
	} else {
		toml.WriteString("disabled = false\n")
	}

	// Write style (with theme variables resolved)
	if module.Style != "" {
		resolvedStyle := r.resolveThemeVars(module.Style, pal)
		toml.WriteString(fmt.Sprintf("style = \"%s\"\n", resolvedStyle))
	}

	// Write format (with theme variables resolved)
	if module.Format != "" {
		resolvedFormat := r.resolveThemeVars(module.Format, pal)
		toml.WriteString(fmt.Sprintf("format = \"%s\"\n", resolvedFormat))
	}

	// Write symbol
	if module.Symbol != "" {
		toml.WriteString(fmt.Sprintf("symbol = \"%s\"\n", module.Symbol))
	}

	// Write additional options
	for key, value := range module.Options {
		r.writeValue(toml, key, value, name, pal)
	}

	toml.WriteString("\n")
}

// writeValue writes a single TOML key-value pair.
func (r *StarshipRenderer) writeValue(toml *strings.Builder, key string, value any, moduleName string, pal *palette.Palette) {
	switch v := value.(type) {
	case string:
		// Resolve theme variables in string values
		resolved := r.resolveThemeVars(v, pal)
		toml.WriteString(fmt.Sprintf("%s = \"%s\"\n", key, resolved))
	case bool:
		toml.WriteString(fmt.Sprintf("%s = %t\n", key, v))
	case int:
		toml.WriteString(fmt.Sprintf("%s = %d\n", key, v))
	case int64:
		toml.WriteString(fmt.Sprintf("%s = %d\n", key, v))
	case float64:
		// Check if it's actually an integer
		if v == float64(int(v)) {
			toml.WriteString(fmt.Sprintf("%s = %d\n", key, int(v)))
		} else {
			toml.WriteString(fmt.Sprintf("%s = %f\n", key, v))
		}
	case []any:
		toml.WriteString(fmt.Sprintf("%s = [", key))
		for i, item := range v {
			if i > 0 {
				toml.WriteString(", ")
			}
			if str, ok := item.(string); ok {
				toml.WriteString(fmt.Sprintf("\"%s\"", str))
			} else {
				toml.WriteString(fmt.Sprintf("%v", item))
			}
		}
		toml.WriteString("]\n")
	case []string:
		toml.WriteString(fmt.Sprintf("%s = [", key))
		for i, item := range v {
			if i > 0 {
				toml.WriteString(", ")
			}
			toml.WriteString(fmt.Sprintf("\"%s\"", item))
		}
		toml.WriteString("]\n")
	case map[string]any:
		// Handle nested maps (e.g., symbols subsection)
		toml.WriteString(fmt.Sprintf("\n[%s.%s]\n", moduleName, key))
		// Sort keys for consistent output
		subKeys := make([]string, 0, len(v))
		for k := range v {
			subKeys = append(subKeys, k)
		}
		sort.Strings(subKeys)
		for _, subKey := range subKeys {
			r.writeValue(toml, subKey, v[subKey], moduleName+"."+key, pal)
		}
	case map[string]string:
		// Handle string maps (e.g., symbols)
		toml.WriteString(fmt.Sprintf("\n[%s.%s]\n", moduleName, key))
		// Sort keys for consistent output
		subKeys := make([]string, 0, len(v))
		for k := range v {
			subKeys = append(subKeys, k)
		}
		sort.Strings(subKeys)
		for _, subKey := range subKeys {
			// Capitalize first letter for Starship OS symbols (e.g., macos -> Macos)
			capitalizedKey := strings.ToUpper(subKey[:1]) + subKey[1:]
			toml.WriteString(fmt.Sprintf("%s = \"%s\"\n", capitalizedKey, v[subKey]))
		}
	default:
		// Fallback: convert to string
		toml.WriteString(fmt.Sprintf("%s = \"%v\"\n", key, value))
	}
}

// resolveThemeVars replaces ${theme.X} variables with actual color values from the palette.
func (r *StarshipRenderer) resolveThemeVars(input string, pal *palette.Palette) string {
	return themeVarRegex.ReplaceAllStringFunc(input, func(match string) string {
		// Extract the variable name from ${theme.varname}
		submatches := themeVarRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match // Return unchanged if no match
		}

		varName := submatches[1]

		// Look up the palette key for this variable name
		paletteKey, exists := themeVarMapping[varName]
		if !exists {
			// Try direct lookup (user might use palette key directly)
			paletteKey = varName
		}

		// Get the color from the palette
		color := pal.Get(paletteKey)
		if color == "" {
			// Try terminal colors
			termColors := pal.ToTerminalColors()
			if c, ok := termColors[paletteKey]; ok {
				color = c
			}
		}

		if color != "" {
			return color
		}

		// Return original match if color not found
		return match
	})
}

// Verify StarshipRenderer implements PromptRenderer interface
var _ PromptRenderer = (*StarshipRenderer)(nil)
