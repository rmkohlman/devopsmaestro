package colors

import (
	"devopsmaestro/pkg/nvimops/theme"
	"devopsmaestro/pkg/palette"
)

// ThemeStoreAdapter bridges theme.Store to PaletteProvider interface.
// This adapter allows the colors package to use theme functionality
// without creating circular imports or tight coupling.
type ThemeStoreAdapter struct {
	store theme.Store
}

// NewThemeStoreAdapter creates a new adapter that wraps a theme store.
// The adapter implements PaletteProvider to bridge the theme system
// with the ColorProvider factory.
func NewThemeStoreAdapter(store theme.Store) PaletteProvider {
	return &ThemeStoreAdapter{
		store: store,
	}
}

// GetActivePalette returns the palette for the currently active theme.
// If no active theme is set, returns nil with no error.
// This allows the factory to fall back to default colors gracefully.
func (a *ThemeStoreAdapter) GetActivePalette() (*palette.Palette, error) {
	if a.store == nil {
		return nil, nil
	}

	activeTheme, err := a.store.GetActive()
	if err != nil {
		return nil, err
	}

	if activeTheme == nil {
		return nil, nil // No active theme set
	}

	return activeTheme.ToPalette(), nil
}

// GetPalette returns the palette for a specific theme by name.
// Returns an error if the theme is not found.
func (a *ThemeStoreAdapter) GetPalette(name string) (*palette.Palette, error) {
	if a.store == nil {
		return nil, &NoProviderError{message: "no theme store configured"}
	}

	themeData, err := a.store.Get(name)
	if err != nil {
		return nil, err
	}

	return themeData.ToPalette(), nil
}
