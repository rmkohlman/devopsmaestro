package colorbridge

import (
	"fmt"

	"github.com/rmkohlman/MaestroPalette"
	"github.com/rmkohlman/MaestroSDK/colors"
	theme "github.com/rmkohlman/MaestroTheme"
)

// ThemeStoreAdapter bridges theme.Store to colors.PaletteProvider interface.
// This adapter allows the colors package to use theme functionality
// without creating circular imports or tight coupling.
//
// This lives in a separate "colorbridge" package because it needs to import both
// the SDK colors package (github.com/rmkohlman/MaestroSDK/colors) and the local
// theme package (devopsmaestro/pkg/nvimops/theme). Having this in pkg/colors/
// would create a package naming conflict with the SDK's colors package.
type ThemeStoreAdapter struct {
	store theme.Store
}

// NewThemeStoreAdapter creates a new adapter that wraps a theme store.
// The adapter implements colors.PaletteProvider to bridge the theme system
// with the ColorProvider factory.
func NewThemeStoreAdapter(store theme.Store) colors.PaletteProvider {
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
		// NoProviderError.message is unexported in the SDK, so we use fmt.Errorf
		return nil, fmt.Errorf("no theme store configured")
	}

	themeData, err := a.store.Get(name)
	if err != nil {
		return nil, err
	}

	return themeData.ToPalette(), nil
}
