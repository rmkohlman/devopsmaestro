package colors

import (
	"context"
	"os"

	"devopsmaestro/pkg/nvimops/theme"
)

// InitColorProviderForCommand initializes ColorProvider for CLI commands.
// This is the main entry point for commands to set up color providers
// with proper fallback handling and no-color support.
//
// Parameters:
//   - ctx: Base context to inject the ColorProvider into
//   - themePath: Path to the theme store directory (e.g., ~/.config/dvm/themes)
//   - noColor: Whether to disable colors (--no-color flag)
//
// Returns:
//   - context.Context with ColorProvider injected
//   - error if theme initialization fails (falls back to default on error)
//
// Behavior:
//   - If noColor is true or NO_COLOR env var is set, returns NoColorProvider
//   - Tries to load active theme from store
//   - Falls back to default colors if no active theme or on error
//   - Never returns an error (always provides some ColorProvider)
func InitColorProviderForCommand(ctx context.Context, themePath string, noColor bool) (context.Context, error) {
	// Check for no-color mode first
	if noColor || os.Getenv("NO_COLOR") != "" {
		return WithProvider(ctx, NewNoColorProvider()), nil
	}

	// Check if theme path is provided
	if themePath == "" {
		// Use default colors if no theme path
		return WithProvider(ctx, NewDefaultColorProvider()), nil
	}

	// Create theme store and adapter
	store := theme.NewFileStore(themePath)
	adapter := NewThemeStoreAdapter(store)
	factory := NewProviderFactory(adapter)

	// Try to create provider from active theme
	provider, err := factory.CreateFromActive()
	if err != nil {
		// Fall back to default on any error (no active theme, invalid theme, etc.)
		provider = NewDefaultColorProvider()
	}

	return WithProvider(ctx, provider), nil
}

// InitColorProviderWithTheme initializes ColorProvider for a specific theme.
// This is used when a command wants to use a specific theme rather than
// the active one (e.g., for preview commands).
//
// Parameters:
//   - ctx: Base context to inject the ColorProvider into
//   - themePath: Path to the theme store directory
//   - themeName: Name of the specific theme to load
//   - noColor: Whether to disable colors
//
// Returns:
//   - context.Context with ColorProvider injected
//   - error if the specific theme cannot be loaded
//
// Unlike InitColorProviderForCommand, this function returns an error
// if the requested theme cannot be found, since the caller explicitly
// requested a specific theme.
func InitColorProviderWithTheme(ctx context.Context, themePath, themeName string, noColor bool) (context.Context, error) {
	// Check for no-color mode first
	if noColor || os.Getenv("NO_COLOR") != "" {
		return WithProvider(ctx, NewNoColorProvider()), nil
	}

	// Validate inputs
	if themePath == "" {
		return WithProvider(ctx, NewDefaultColorProvider()), nil
	}
	if themeName == "" {
		return WithProvider(ctx, NewDefaultColorProvider()), nil
	}

	// Create theme store and adapter
	store := theme.NewFileStore(themePath)
	adapter := NewThemeStoreAdapter(store)
	factory := NewProviderFactory(adapter)

	// Try to create provider from specific theme
	provider, err := factory.CreateFromTheme(themeName)
	if err != nil {
		return ctx, err // Return error for specific theme requests
	}

	return WithProvider(ctx, provider), nil
}

// GetDefaultThemePath returns the default path for theme storage.
// This is a convenience function for commands that need to determine
// the theme path based on user configuration or environment.
func GetDefaultThemePath() string {
	// Check environment variable first
	if path := os.Getenv("DVM_THEME_PATH"); path != "" {
		return path
	}

	// Check XDG config directory
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return xdgConfig + "/dvm/themes"
	}

	// Fall back to home directory
	if home := os.Getenv("HOME"); home != "" {
		return home + "/.config/dvm/themes"
	}

	// Last resort - current directory
	return "./themes"
}

// IsNoColorRequested checks if color should be disabled based on
// environment variables or explicit flags.
func IsNoColorRequested(noColorFlag bool) bool {
	return noColorFlag ||
		os.Getenv("NO_COLOR") != "" ||
		os.Getenv("TERM") == "dumb"
}
