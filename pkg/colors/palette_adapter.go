package colors

import (
	"devopsmaestro/pkg/palette"
)

// PaletteAdapter defines the interface for converting to a palette.Palette.
// This is a separate concern from ColorProvider, following the single responsibility principle.
// It allows consumers (like StarshipRenderer) to get the full palette when needed
// without polluting the ColorProvider interface with palette-specific concerns.
type PaletteAdapter interface {
	// ToPalette converts the provider's colors to a *palette.Palette.
	// Returns a palette with semantic color mappings suitable for terminal/starship use.
	ToPalette() *palette.Palette
}

// ColorToPaletteAdapter implements the Adapter Pattern to convert a ColorProvider
// to a PaletteAdapter. This bridges the gap between the clean ColorProvider interface
// and the palette system needed by terminal rendering components like StarshipRenderer.
//
// The adapter maps ColorProvider methods to palette.Palette semantic color constants:
//   - Primary() → palette.ColorPrimary
//   - Secondary() → palette.ColorSecondary
//   - Success() → palette.ColorSuccess
//   - Warning() → palette.ColorWarning
//   - Error() → palette.ColorError
//   - Info() → palette.ColorInfo
//   - Foreground() → palette.ColorFg
//   - Background() → palette.ColorBg
//   - Muted() → palette.ColorComment (maps to muted text)
//   - Highlight() → palette.ColorBgHighlight
//   - Border() → palette.ColorBorder
//   - Accent() → palette.ColorAccent
//
// For NoColorProvider, this returns a palette with empty color values instead of nil,
// ensuring that palette-based rendering doesn't fail but simply produces no colors.
type ColorToPaletteAdapter struct {
	provider ColorProvider
}

// NewColorToPaletteAdapter creates a PaletteAdapter from a ColorProvider.
// This factory function implements the Adapter Pattern, allowing any ColorProvider
// to be used where a PaletteAdapter is expected.
//
// Example usage:
//
//	provider := colors.FromContext(ctx)
//	paletteAdapter := colors.NewColorToPaletteAdapter(provider)
//	palette := paletteAdapter.ToPalette()
//	starshipRenderer.Generate(palette)
func NewColorToPaletteAdapter(provider ColorProvider) PaletteAdapter {
	return &ColorToPaletteAdapter{
		provider: provider,
	}
}

// ToPalette converts the ColorProvider's colors to a *palette.Palette.
// The returned palette contains semantic color mappings that are compatible
// with terminal rendering and the existing palette system.
//
// Color mapping strategy:
//   - Maps each ColorProvider method to its corresponding palette constant
//   - Uses the provider's Name() and IsLight() for palette metadata
//   - For NoColorProvider, returns a valid palette with empty colors (not nil)
//   - Sets palette category based on IsLight(): "light" or "dark"
//
// The resulting palette can be used with:
//   - StarshipRenderer for terminal prompt generation
//   - Terminal color scheme generation
//   - Any palette-based rendering system
func (a *ColorToPaletteAdapter) ToPalette() *palette.Palette {
	// Determine category from provider
	category := palette.CategoryDark
	if a.provider.IsLight() {
		category = palette.CategoryLight
	}

	// Create palette with metadata
	p := &palette.Palette{
		Name:        a.provider.Name(),
		Description: "Generated from ColorProvider",
		Category:    category,
		Colors:      make(map[string]string),
	}

	// Map ColorProvider methods to palette semantic constants
	// Only set colors that are non-empty (handles NoColorProvider gracefully)
	if color := a.provider.Primary(); color != "" {
		p.Colors[palette.ColorPrimary] = color
	}
	if color := a.provider.Secondary(); color != "" {
		p.Colors[palette.ColorSecondary] = color
	}
	if color := a.provider.Accent(); color != "" {
		p.Colors[palette.ColorAccent] = color
	}

	// Status colors
	if color := a.provider.Success(); color != "" {
		p.Colors[palette.ColorSuccess] = color
	}
	if color := a.provider.Warning(); color != "" {
		p.Colors[palette.ColorWarning] = color
	}
	if color := a.provider.Error(); color != "" {
		p.Colors[palette.ColorError] = color
	}
	if color := a.provider.Info(); color != "" {
		p.Colors[palette.ColorInfo] = color
	}

	// UI colors
	if color := a.provider.Foreground(); color != "" {
		p.Colors[palette.ColorFg] = color
	}
	if color := a.provider.Background(); color != "" {
		p.Colors[palette.ColorBg] = color
	}
	if color := a.provider.Muted(); color != "" {
		// Map Muted to ColorComment as they serve similar purposes (subdued text)
		p.Colors[palette.ColorComment] = color
	}
	if color := a.provider.Highlight(); color != "" {
		p.Colors[palette.ColorBgHighlight] = color
	}
	if color := a.provider.Border(); color != "" {
		p.Colors[palette.ColorBorder] = color
	}

	return p
}

// ToPalette is a convenience function that creates an adapter and converts in one call.
// This is useful for simple cases where you just need to convert a ColorProvider
// to a palette without keeping the adapter around.
//
// Example usage:
//
//	provider := colors.FromContext(ctx)
//	palette := colors.ToPalette(provider)
//	starshipRenderer.Generate(palette)
//
// This is equivalent to:
//
//	adapter := colors.NewColorToPaletteAdapter(provider)
//	palette := adapter.ToPalette()
func ToPalette(provider ColorProvider) *palette.Palette {
	adapter := NewColorToPaletteAdapter(provider)
	return adapter.ToPalette()
}
