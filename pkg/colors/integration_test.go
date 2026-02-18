package colors_test

import (
	"context"
	"os"
	"testing"

	"devopsmaestro/pkg/colors"
	"devopsmaestro/pkg/nvimops/theme"
)

// TestColorProviderIntegration demonstrates the complete workflow
// from theme store to ColorProvider using the new adapter system.
func TestColorProviderIntegration(t *testing.T) {
	// Create temporary directory for test theme store
	tempDir := t.TempDir()

	// Create a test theme
	testTheme := &theme.Theme{
		Name:        "test-theme",
		Description: "Test theme for integration testing",
		Category:    "dark",
		Plugin: theme.ThemePlugin{
			Repo: "test/theme",
		},
		Colors: map[string]string{
			"bg":      "#1a1b26",
			"fg":      "#c0caf5",
			"primary": "#7aa2f7",
			"error":   "#f7768e",
			"success": "#9ece6a",
			"warning": "#e0af68",
		},
	}

	// Save the theme to file store
	store := theme.NewFileStore(tempDir)
	err := store.Save(testTheme)
	if err != nil {
		t.Fatalf("Failed to save test theme: %v", err)
	}

	// Set as active theme
	err = store.SetActive("test-theme")
	if err != nil {
		t.Fatalf("Failed to set active theme: %v", err)
	}

	t.Run("InitColorProviderForCommand with active theme", func(t *testing.T) {
		ctx := context.Background()

		// Initialize ColorProvider using the command helper
		ctx, err := colors.InitColorProviderForCommand(ctx, tempDir, false)
		if err != nil {
			t.Fatalf("InitColorProviderForCommand failed: %v", err)
		}

		// Get provider from context
		provider, ok := colors.FromContext(ctx)
		if !ok {
			t.Fatal("ColorProvider not found in context")
		}

		// Verify it has the expected colors from our test theme
		if provider.Primary() != "#7aa2f7" {
			t.Errorf("Expected primary color #7aa2f7, got %s", provider.Primary())
		}
		if provider.Error() != "#f7768e" {
			t.Errorf("Expected error color #f7768e, got %s", provider.Error())
		}
		if provider.Foreground() != "#c0caf5" {
			t.Errorf("Expected foreground color #c0caf5, got %s", provider.Foreground())
		}
		if provider.Name() != "test-theme" {
			t.Errorf("Expected theme name 'test-theme', got %s", provider.Name())
		}
	})

	t.Run("InitColorProviderForCommand with no-color flag", func(t *testing.T) {
		ctx := context.Background()

		// Initialize with no-color flag
		ctx, err := colors.InitColorProviderForCommand(ctx, tempDir, true)
		if err != nil {
			t.Fatalf("InitColorProviderForCommand failed: %v", err)
		}

		// Get provider from context
		provider, ok := colors.FromContext(ctx)
		if !ok {
			t.Fatal("ColorProvider not found in context")
		}

		// Verify all colors are empty (no-color mode)
		if provider.Primary() != "" {
			t.Errorf("Expected empty primary color in no-color mode, got %s", provider.Primary())
		}
		if provider.Error() != "" {
			t.Errorf("Expected empty error color in no-color mode, got %s", provider.Error())
		}
		if provider.Name() != "no-color" {
			t.Errorf("Expected provider name 'no-color', got %s", provider.Name())
		}
	})

	t.Run("InitColorProviderForCommand with NO_COLOR env var", func(t *testing.T) {
		// Set NO_COLOR environment variable
		originalValue := os.Getenv("NO_COLOR")
		os.Setenv("NO_COLOR", "1")
		defer func() {
			if originalValue == "" {
				os.Unsetenv("NO_COLOR")
			} else {
				os.Setenv("NO_COLOR", originalValue)
			}
		}()

		ctx := context.Background()

		// Initialize without no-color flag (should still respect env var)
		ctx, err := colors.InitColorProviderForCommand(ctx, tempDir, false)
		if err != nil {
			t.Fatalf("InitColorProviderForCommand failed: %v", err)
		}

		// Get provider from context
		provider, ok := colors.FromContext(ctx)
		if !ok {
			t.Fatal("ColorProvider not found in context")
		}

		// Verify it's the no-color provider
		if provider.Name() != "no-color" {
			t.Errorf("Expected no-color provider due to NO_COLOR env var, got %s", provider.Name())
		}
	})

	t.Run("InitColorProviderWithTheme specific theme", func(t *testing.T) {
		ctx := context.Background()

		// Initialize with specific theme name
		ctx, err := colors.InitColorProviderWithTheme(ctx, tempDir, "test-theme", false)
		if err != nil {
			t.Fatalf("InitColorProviderWithTheme failed: %v", err)
		}

		// Get provider from context
		provider, ok := colors.FromContext(ctx)
		if !ok {
			t.Fatal("ColorProvider not found in context")
		}

		// Verify it loaded the correct theme
		if provider.Name() != "test-theme" {
			t.Errorf("Expected theme name 'test-theme', got %s", provider.Name())
		}
		if provider.Primary() != "#7aa2f7" {
			t.Errorf("Expected primary color #7aa2f7, got %s", provider.Primary())
		}
	})

	t.Run("InitColorProviderWithTheme non-existent theme", func(t *testing.T) {
		ctx := context.Background()

		// Try to initialize with non-existent theme
		_, err := colors.InitColorProviderWithTheme(ctx, tempDir, "non-existent", false)
		if err == nil {
			t.Fatal("Expected error for non-existent theme, got nil")
		}
	})

	t.Run("Fallback to default colors", func(t *testing.T) {
		// Use empty theme path to trigger fallback
		ctx := context.Background()

		ctx, err := colors.InitColorProviderForCommand(ctx, "", false)
		if err != nil {
			t.Fatalf("InitColorProviderForCommand failed: %v", err)
		}

		provider, ok := colors.FromContext(ctx)
		if !ok {
			t.Fatal("ColorProvider not found in context")
		}

		// Should get default provider
		if provider.Name() != "default" {
			t.Errorf("Expected default provider, got %s", provider.Name())
		}

		// Should have default colors (not empty)
		if provider.Primary() == "" {
			t.Error("Expected default primary color, got empty string")
		}
	})
}

// TestDefaultThemePath tests the theme path resolution logic.
func TestDefaultThemePath(t *testing.T) {
	// Save original environment
	originalDVM := os.Getenv("DVM_THEME_PATH")
	originalXDG := os.Getenv("XDG_CONFIG_HOME")
	originalHome := os.Getenv("HOME")

	setOrUnset := func(key, value string) {
		if value == "" {
			os.Unsetenv(key)
		} else {
			os.Setenv(key, value)
		}
	}

	defer func() {
		// Restore environment
		setOrUnset("DVM_THEME_PATH", originalDVM)
		setOrUnset("XDG_CONFIG_HOME", originalXDG)
		setOrUnset("HOME", originalHome)
	}()

	t.Run("DVM_THEME_PATH takes precedence", func(t *testing.T) {
		os.Setenv("DVM_THEME_PATH", "/custom/theme/path")
		os.Setenv("XDG_CONFIG_HOME", "/xdg")
		os.Setenv("HOME", "/home/user")

		path := colors.GetDefaultThemePath()
		expected := "/custom/theme/path"
		if path != expected {
			t.Errorf("Expected %s, got %s", expected, path)
		}
	})

	t.Run("XDG_CONFIG_HOME used when DVM_THEME_PATH not set", func(t *testing.T) {
		os.Unsetenv("DVM_THEME_PATH")
		os.Setenv("XDG_CONFIG_HOME", "/xdg")
		os.Setenv("HOME", "/home/user")

		path := colors.GetDefaultThemePath()
		expected := "/xdg/dvm/themes"
		if path != expected {
			t.Errorf("Expected %s, got %s", expected, path)
		}
	})

	t.Run("HOME/.config used when XDG not set", func(t *testing.T) {
		os.Unsetenv("DVM_THEME_PATH")
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Setenv("HOME", "/home/user")

		path := colors.GetDefaultThemePath()
		expected := "/home/user/.config/dvm/themes"
		if path != expected {
			t.Errorf("Expected %s, got %s", expected, path)
		}
	})

	t.Run("Fallback to ./themes", func(t *testing.T) {
		os.Unsetenv("DVM_THEME_PATH")
		os.Unsetenv("XDG_CONFIG_HOME")
		os.Unsetenv("HOME")

		path := colors.GetDefaultThemePath()
		expected := "./themes"
		if path != expected {
			t.Errorf("Expected %s, got %s", expected, path)
		}
	})
}

// TestNoColorDetection tests the no-color detection logic.
func TestNoColorDetection(t *testing.T) {
	// Save original environment
	originalNoColor := os.Getenv("NO_COLOR")
	originalTerm := os.Getenv("TERM")

	defer func() {
		// Restore environment
		if originalNoColor == "" {
			os.Unsetenv("NO_COLOR")
		} else {
			os.Setenv("NO_COLOR", originalNoColor)
		}
		if originalTerm == "" {
			os.Unsetenv("TERM")
		} else {
			os.Setenv("TERM", originalTerm)
		}
	}()

	t.Run("No-color flag takes precedence", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		os.Setenv("TERM", "xterm-256color")

		if !colors.IsNoColorRequested(true) {
			t.Error("Expected no-color when flag is true")
		}
		if colors.IsNoColorRequested(false) {
			t.Error("Expected color when flag is false and no env vars set")
		}
	})

	t.Run("NO_COLOR environment variable", func(t *testing.T) {
		os.Setenv("NO_COLOR", "1")
		os.Setenv("TERM", "xterm-256color")

		if !colors.IsNoColorRequested(false) {
			t.Error("Expected no-color when NO_COLOR env var is set")
		}
	})

	t.Run("TERM=dumb disables color", func(t *testing.T) {
		os.Unsetenv("NO_COLOR")
		os.Setenv("TERM", "dumb")

		if !colors.IsNoColorRequested(false) {
			t.Error("Expected no-color when TERM=dumb")
		}
	})
}
