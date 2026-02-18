package colors

import (
	"devopsmaestro/pkg/palette"
)

// ThemeColorProvider implements ColorProvider using a palette.
// This is the concrete implementation that bridges themes to color consumers.
type ThemeColorProvider struct {
	palette  *palette.Palette
	name     string
	category palette.Category
}

// NewThemeColorProvider creates a ColorProvider from a palette.
// The palette should contain the standard color keys defined in pkg/palette.
func NewThemeColorProvider(p *palette.Palette) ColorProvider {
	if p == nil {
		// Return a default provider if nil palette
		return NewDefaultColorProvider()
	}
	return &ThemeColorProvider{
		palette:  p,
		name:     p.Name,
		category: p.Category,
	}
}

// getColor retrieves a color from the palette with fallback support.
func (p *ThemeColorProvider) getColor(key, fallback string) string {
	if color := p.palette.Get(key); color != "" {
		return color
	}
	return fallback
}

// getColorWithFallbacks tries multiple keys in order, returning the first found.
func (p *ThemeColorProvider) getColorWithFallbacks(fallback string, keys ...string) string {
	for _, key := range keys {
		if color := p.palette.Get(key); color != "" {
			return color
		}
	}
	return fallback
}

// =============================================================================
// Primary Colors
// =============================================================================

// Primary returns the main brand/accent color.
func (p *ThemeColorProvider) Primary() string {
	return p.getColorWithFallbacks(DefaultDarkColors["primary"],
		palette.ColorPrimary, palette.ColorAccent, "blue")
}

// Secondary returns the secondary brand color.
func (p *ThemeColorProvider) Secondary() string {
	return p.getColorWithFallbacks(DefaultDarkColors["secondary"],
		palette.ColorSecondary, "purple", "magenta")
}

// Accent returns the highlight/focus color.
func (p *ThemeColorProvider) Accent() string {
	return p.getColorWithFallbacks(DefaultDarkColors["accent"],
		palette.ColorAccent, palette.ColorPrimary, "blue")
}

// =============================================================================
// Status Colors
// =============================================================================

// Success returns the success state color (typically green).
func (p *ThemeColorProvider) Success() string {
	return p.getColorWithFallbacks(DefaultDarkColors["success"],
		palette.ColorSuccess, "green")
}

// Warning returns the warning state color (typically yellow/orange).
func (p *ThemeColorProvider) Warning() string {
	return p.getColorWithFallbacks(DefaultDarkColors["warning"],
		palette.ColorWarning, "yellow", "orange")
}

// Error returns the error state color (typically red).
func (p *ThemeColorProvider) Error() string {
	return p.getColorWithFallbacks(DefaultDarkColors["error"],
		palette.ColorError, "red")
}

// Info returns the info state color (typically blue/cyan).
func (p *ThemeColorProvider) Info() string {
	return p.getColorWithFallbacks(DefaultDarkColors["info"],
		palette.ColorInfo, "cyan", "blue")
}

// =============================================================================
// UI Colors
// =============================================================================

// Foreground returns the main text color.
func (p *ThemeColorProvider) Foreground() string {
	return p.getColor(palette.ColorFg, DefaultDarkColors["foreground"])
}

// Background returns the main background color.
func (p *ThemeColorProvider) Background() string {
	return p.getColor(palette.ColorBg, DefaultDarkColors["background"])
}

// Muted returns the subdued/disabled text color.
func (p *ThemeColorProvider) Muted() string {
	return p.getColorWithFallbacks(DefaultDarkColors["muted"],
		palette.ColorComment, palette.ColorFgGutter)
}

// Highlight returns the selection/hover background color.
func (p *ThemeColorProvider) Highlight() string {
	return p.getColorWithFallbacks(DefaultDarkColors["highlight"],
		palette.ColorBgHighlight, palette.ColorBgVisual)
}

// Border returns the border/separator color.
func (p *ThemeColorProvider) Border() string {
	return p.getColorWithFallbacks(DefaultDarkColors["border"],
		palette.ColorBorder, palette.ColorComment)
}

// =============================================================================
// Metadata
// =============================================================================

// Name returns the theme/palette name.
func (p *ThemeColorProvider) Name() string {
	return p.name
}

// IsLight returns true if this is a light theme.
func (p *ThemeColorProvider) IsLight() bool {
	return p.category == palette.CategoryLight
}

// =============================================================================
// Default Color Provider
// =============================================================================

// DefaultDarkColors provides sensible defaults for dark themes (Tokyo Night inspired).
var DefaultDarkColors = map[string]string{
	"primary":    "#7aa2f7",
	"secondary":  "#bb9af7",
	"accent":     "#7aa2f7",
	"success":    "#9ece6a",
	"warning":    "#e0af68",
	"error":      "#f7768e",
	"info":       "#7dcfff",
	"foreground": "#c0caf5",
	"background": "#1a1b26",
	"muted":      "#565f89",
	"highlight":  "#283457",
	"border":     "#414868",
}

// DefaultLightColors provides sensible defaults for light themes.
var DefaultLightColors = map[string]string{
	"primary":    "#3d59a1",
	"secondary":  "#9854f1",
	"accent":     "#3d59a1",
	"success":    "#587539",
	"warning":    "#8c6c3e",
	"error":      "#c64343",
	"info":       "#007197",
	"foreground": "#24292e",
	"background": "#ffffff",
	"muted":      "#586069",
	"highlight":  "#e1e4e8",
	"border":     "#d0d7de",
}

// defaultColorProvider is a fallback implementation with default colors.
type defaultColorProvider struct {
	colors  map[string]string
	name    string
	isLight bool
}

// NewDefaultColorProvider creates a ColorProvider with default dark theme colors.
func NewDefaultColorProvider() ColorProvider {
	return &defaultColorProvider{
		colors:  DefaultDarkColors,
		name:    "default",
		isLight: false,
	}
}

// NewDefaultLightColorProvider creates a ColorProvider with default light theme colors.
func NewDefaultLightColorProvider() ColorProvider {
	return &defaultColorProvider{
		colors:  DefaultLightColors,
		name:    "default-light",
		isLight: true,
	}
}

func (d *defaultColorProvider) Primary() string    { return d.colors["primary"] }
func (d *defaultColorProvider) Secondary() string  { return d.colors["secondary"] }
func (d *defaultColorProvider) Accent() string     { return d.colors["accent"] }
func (d *defaultColorProvider) Success() string    { return d.colors["success"] }
func (d *defaultColorProvider) Warning() string    { return d.colors["warning"] }
func (d *defaultColorProvider) Error() string      { return d.colors["error"] }
func (d *defaultColorProvider) Info() string       { return d.colors["info"] }
func (d *defaultColorProvider) Foreground() string { return d.colors["foreground"] }
func (d *defaultColorProvider) Background() string { return d.colors["background"] }
func (d *defaultColorProvider) Muted() string      { return d.colors["muted"] }
func (d *defaultColorProvider) Highlight() string  { return d.colors["highlight"] }
func (d *defaultColorProvider) Border() string     { return d.colors["border"] }
func (d *defaultColorProvider) Name() string       { return d.name }
func (d *defaultColorProvider) IsLight() bool      { return d.isLight }
