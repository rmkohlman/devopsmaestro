// Package palette provides shared color palette types and utilities for theming.
// This package is designed to be used by multiple sub-tools (nvimops, terminalops)
// to ensure consistent color handling across the DevOpsMaestro ecosystem.
package palette

// Palette represents a color palette that can be used across different tools.
// Colors are stored as hex values (e.g., "#1a1b26") and can be converted
// to tool-specific formats by each consumer.
type Palette struct {
	// Name is the palette identifier
	Name string `yaml:"name" json:"name"`

	// Description provides details about the palette
	Description string `yaml:"description,omitempty" json:"description,omitempty"`

	// Author is the palette creator
	Author string `yaml:"author,omitempty" json:"author,omitempty"`

	// Category indicates the palette type (dark, light, both)
	Category Category `yaml:"category,omitempty" json:"category,omitempty"`

	// Colors maps semantic color names to hex values
	Colors map[string]string `yaml:"colors,omitempty" json:"colors,omitempty"`
}

// Category represents the type of color palette.
type Category string

const (
	// CategoryDark indicates a dark color scheme
	CategoryDark Category = "dark"
	// CategoryLight indicates a light color scheme
	CategoryLight Category = "light"
	// CategoryBoth indicates support for both dark and light modes
	CategoryBoth Category = "both"
)

// =============================================================================
// Semantic Color Keys
// =============================================================================
// These constants define semantic color names that are shared across tools.
// Each tool interprets these colors according to its own needs.

// Background colors
const (
	// ColorBg is the primary background color
	ColorBg = "bg"
	// ColorBgDark is a darker background variant
	ColorBgDark = "bg_dark"
	// ColorBgHighlight is the background for highlighted elements
	ColorBgHighlight = "bg_highlight"
	// ColorBgSearch is the background for search matches
	ColorBgSearch = "bg_search"
	// ColorBgVisual is the background for visual selections
	ColorBgVisual = "bg_visual"
	// ColorBgFloat is the background for floating windows/panels
	ColorBgFloat = "bg_float"
	// ColorBgPopup is the background for popups/menus
	ColorBgPopup = "bg_popup"
	// ColorBgSidebar is the background for sidebars
	ColorBgSidebar = "bg_sidebar"
	// ColorBgStatusline is the background for status lines
	ColorBgStatusline = "bg_statusline"
)

// Foreground colors
const (
	// ColorFg is the primary foreground/text color
	ColorFg = "fg"
	// ColorFgDark is a darker foreground variant
	ColorFgDark = "fg_dark"
	// ColorFgGutter is the foreground for gutter/line numbers
	ColorFgGutter = "fg_gutter"
	// ColorFgSidebar is the foreground for sidebar text
	ColorFgSidebar = "fg_sidebar"
)

// UI element colors
const (
	// ColorBorder is the color for borders
	ColorBorder = "border"
	// ColorComment is the color for comments/muted text
	ColorComment = "comment"
)

// Diagnostic/status colors
const (
	// ColorError is the color for errors
	ColorError = "error"
	// ColorWarning is the color for warnings
	ColorWarning = "warning"
	// ColorInfo is the color for informational messages
	ColorInfo = "info"
	// ColorHint is the color for hints
	ColorHint = "hint"
	// ColorSuccess is the color for success states
	ColorSuccess = "success"
)

// Accent colors
const (
	// ColorPrimary is the primary accent color
	ColorPrimary = "primary"
	// ColorSecondary is the secondary accent color
	ColorSecondary = "secondary"
	// ColorAccent is a general accent color
	ColorAccent = "accent"
)

// =============================================================================
// Standard Color Names (used in themes and terminal configs)
// =============================================================================
// These are the standard color names found in theme files that can be
// mapped to terminal ANSI colors.

const (
	ColorBlack   = "black"
	ColorRed     = "red"
	ColorGreen   = "green"
	ColorYellow  = "yellow"
	ColorBlue    = "blue"
	ColorMagenta = "magenta"
	ColorCyan    = "cyan"
	ColorWhite   = "white"
	ColorOrange  = "orange"
	ColorPurple  = "purple"
	ColorTeal    = "teal"
	ColorPink    = "pink"
)

// =============================================================================
// Terminal-Specific Color Keys (ANSI 16-color palette)
// =============================================================================
// These are the standard ANSI terminal color names that terminal emulators use.

const (
	// Normal colors (0-7)
	TermBlack   = "ansi_black"
	TermRed     = "ansi_red"
	TermGreen   = "ansi_green"
	TermYellow  = "ansi_yellow"
	TermBlue    = "ansi_blue"
	TermMagenta = "ansi_magenta"
	TermCyan    = "ansi_cyan"
	TermWhite   = "ansi_white"

	// Bright colors (8-15)
	TermBrightBlack   = "ansi_bright_black"
	TermBrightRed     = "ansi_bright_red"
	TermBrightGreen   = "ansi_bright_green"
	TermBrightYellow  = "ansi_bright_yellow"
	TermBrightBlue    = "ansi_bright_blue"
	TermBrightMagenta = "ansi_bright_magenta"
	TermBrightCyan    = "ansi_bright_cyan"
	TermBrightWhite   = "ansi_bright_white"

	// Terminal UI colors
	TermCursor        = "cursor"
	TermCursorText    = "cursor_text"
	TermSelection     = "selection"
	TermSelectionText = "selection_text"
)

// TerminalColorMapping defines how theme colors map to terminal ANSI colors.
// This allows a single theme palette to be used for both editor and terminal.
var TerminalColorMapping = map[string]string{
	TermBlack:   ColorBgDark, // or "black" if present
	TermRed:     ColorRed,
	TermGreen:   ColorGreen,
	TermYellow:  ColorYellow,
	TermBlue:    ColorBlue,
	TermMagenta: ColorMagenta,
	TermCyan:    ColorCyan,
	TermWhite:   ColorFg,

	TermBrightBlack:   ColorComment,
	TermBrightRed:     ColorError,
	TermBrightGreen:   ColorGreen,
	TermBrightYellow:  ColorWarning,
	TermBrightBlue:    ColorInfo,
	TermBrightMagenta: ColorMagenta,
	TermBrightCyan:    ColorCyan,
	TermBrightWhite:   ColorFgDark,

	TermCursor:        ColorFg,
	TermCursorText:    ColorBg,
	TermSelection:     ColorBgVisual,
	TermSelectionText: ColorFg,
}

// =============================================================================
// Palette Methods
// =============================================================================

// Get retrieves a color value by key, returning empty string if not found.
func (p *Palette) Get(key string) string {
	if p.Colors == nil {
		return ""
	}
	return p.Colors[key]
}

// GetOrDefault retrieves a color value by key, returning the default if not found.
func (p *Palette) GetOrDefault(key, defaultColor string) string {
	if p.Colors == nil {
		return defaultColor
	}
	if color, ok := p.Colors[key]; ok && color != "" {
		return color
	}
	return defaultColor
}

// Set sets a color value for a key.
func (p *Palette) Set(key, value string) {
	if p.Colors == nil {
		p.Colors = make(map[string]string)
	}
	p.Colors[key] = value
}

// Has checks if a color key exists in the palette.
func (p *Palette) Has(key string) bool {
	if p.Colors == nil {
		return false
	}
	_, ok := p.Colors[key]
	return ok
}

// Merge combines another palette's colors into this one.
// Existing colors are not overwritten unless overwrite is true.
func (p *Palette) Merge(other *Palette, overwrite bool) {
	if other == nil || other.Colors == nil {
		return
	}
	if p.Colors == nil {
		p.Colors = make(map[string]string)
	}
	for key, value := range other.Colors {
		if overwrite || !p.Has(key) {
			p.Colors[key] = value
		}
	}
}

// Clone creates a deep copy of the palette.
func (p *Palette) Clone() *Palette {
	if p == nil {
		return nil
	}
	clone := &Palette{
		Name:        p.Name,
		Description: p.Description,
		Author:      p.Author,
		Category:    p.Category,
	}
	if p.Colors != nil {
		clone.Colors = make(map[string]string, len(p.Colors))
		for k, v := range p.Colors {
			clone.Colors[k] = v
		}
	}
	return clone
}

// ToTerminalColors extracts ANSI 16-color terminal palette from theme colors.
// It uses TerminalColorMapping to map semantic theme colors to terminal colors,
// with fallbacks for missing colors.
func (p *Palette) ToTerminalColors() map[string]string {
	if p == nil {
		return nil
	}

	terminal := make(map[string]string)

	// If no colors, return empty map
	if p.Colors == nil {
		return terminal
	}

	// Helper to get color with fallback chain
	getColor := func(keys ...string) string {
		for _, key := range keys {
			if color := p.Get(key); color != "" {
				return color
			}
		}
		return ""
	}

	// Map standard ANSI colors from theme
	// Normal colors (0-7)
	terminal[TermBlack] = getColor("black", ColorBgDark, ColorBg)
	terminal[TermRed] = getColor("red", ColorError)
	terminal[TermGreen] = getColor("green", ColorSuccess)
	terminal[TermYellow] = getColor("yellow", ColorWarning)
	terminal[TermBlue] = getColor("blue", ColorInfo, ColorPrimary)
	terminal[TermMagenta] = getColor("magenta", "purple", "pink")
	terminal[TermCyan] = getColor("cyan", "teal")
	terminal[TermWhite] = getColor("white", ColorFg)

	// Bright colors (8-15) - use brighter variants or same as normal
	terminal[TermBrightBlack] = getColor("bright_black", ColorComment, ColorFgGutter)
	terminal[TermBrightRed] = getColor("bright_red", "red", ColorError)
	terminal[TermBrightGreen] = getColor("bright_green", "green", ColorSuccess)
	terminal[TermBrightYellow] = getColor("bright_yellow", "yellow", "orange", ColorWarning)
	terminal[TermBrightBlue] = getColor("bright_blue", "blue", ColorInfo)
	terminal[TermBrightMagenta] = getColor("bright_magenta", "magenta", "purple", "lavender")
	terminal[TermBrightCyan] = getColor("bright_cyan", "cyan", "teal", "sky")
	terminal[TermBrightWhite] = getColor("bright_white", ColorFgDark, ColorFg)

	// Terminal UI colors
	terminal[ColorBg] = getColor(ColorBg)
	terminal[ColorFg] = getColor(ColorFg)
	terminal[TermCursor] = getColor(TermCursor, ColorFg)
	terminal[TermCursorText] = getColor(TermCursorText, ColorBg)
	terminal[TermSelection] = getColor(TermSelection, ColorBgVisual, ColorBgHighlight)
	terminal[TermSelectionText] = getColor(TermSelectionText, ColorFg)

	// Remove empty values
	for k, v := range terminal {
		if v == "" {
			delete(terminal, k)
		}
	}

	return terminal
}

// TerminalPalette creates a new Palette containing only terminal-relevant colors.
func (p *Palette) TerminalPalette() *Palette {
	if p == nil {
		return nil
	}

	return &Palette{
		Name:        p.Name + "-terminal",
		Description: p.Description,
		Author:      p.Author,
		Category:    p.Category,
		Colors:      p.ToTerminalColors(),
	}
}
