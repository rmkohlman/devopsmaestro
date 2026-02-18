package colors

import (
	"devopsmaestro/pkg/palette"
)

// PaletteProvider is an interface for getting palettes.
// This allows the factory to work without importing the theme package directly,
// maintaining the decoupled architecture.
type PaletteProvider interface {
	// GetActivePalette returns the palette for the active theme.
	// Returns nil if no active theme is set.
	GetActivePalette() (*palette.Palette, error)

	// GetPalette returns the palette for a specific theme by name.
	GetPalette(name string) (*palette.Palette, error)
}

// ProviderFactory creates ColorProvider instances.
// This is the main entry point for obtaining color providers.
type ProviderFactory interface {
	// CreateFromActive creates a ColorProvider from the active theme.
	// Returns a default provider if no active theme is set.
	CreateFromActive() (ColorProvider, error)

	// CreateFromTheme creates a ColorProvider from a specific theme.
	CreateFromTheme(themeName string) (ColorProvider, error)

	// CreateDefault creates a ColorProvider with default colors.
	// If isLight is true, uses light theme defaults.
	CreateDefault(isLight bool) ColorProvider
}

// providerFactory implements ProviderFactory.
type providerFactory struct {
	paletteProvider PaletteProvider
}

// NewProviderFactory creates a new factory with a palette provider.
// The palette provider bridges to the theme system without direct import.
func NewProviderFactory(pp PaletteProvider) ProviderFactory {
	return &providerFactory{
		paletteProvider: pp,
	}
}

// CreateFromActive creates a ColorProvider from the active theme.
func (f *providerFactory) CreateFromActive() (ColorProvider, error) {
	if f.paletteProvider == nil {
		return NewDefaultColorProvider(), nil
	}

	p, err := f.paletteProvider.GetActivePalette()
	if err != nil {
		// Return default on error (no active theme, etc.)
		return NewDefaultColorProvider(), nil
	}

	return NewThemeColorProvider(p), nil
}

// CreateFromTheme creates a ColorProvider from a specific theme.
func (f *providerFactory) CreateFromTheme(themeName string) (ColorProvider, error) {
	if f.paletteProvider == nil {
		return nil, &NoProviderError{message: "no palette provider configured"}
	}

	p, err := f.paletteProvider.GetPalette(themeName)
	if err != nil {
		return nil, err
	}

	return NewThemeColorProvider(p), nil
}

// CreateDefault creates a ColorProvider with default colors.
func (f *providerFactory) CreateDefault(isLight bool) ColorProvider {
	if isLight {
		return NewDefaultLightColorProvider()
	}
	return NewDefaultColorProvider()
}

// =============================================================================
// Static Factory Functions (for simple use cases)
// =============================================================================

// FromPalette creates a ColorProvider directly from a palette.
// This is useful when you already have a palette and don't need the full factory.
func FromPalette(p *palette.Palette) ColorProvider {
	return NewThemeColorProvider(p)
}

// Default returns a default dark theme ColorProvider.
func Default() ColorProvider {
	return NewDefaultColorProvider()
}

// DefaultLight returns a default light theme ColorProvider.
func DefaultLight() ColorProvider {
	return NewDefaultLightColorProvider()
}

// =============================================================================
// Errors
// =============================================================================

// NoProviderError indicates no palette provider is configured.
type NoProviderError struct {
	message string
}

func (e *NoProviderError) Error() string {
	return e.message
}
