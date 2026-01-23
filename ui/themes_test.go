package ui

import (
	"os"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/assert"
)

func TestAvailableThemes(t *testing.T) {
	themes := AvailableThemes()

	assert.NotEmpty(t, themes, "Should have available themes")
	assert.Contains(t, themes, ThemeAuto, "Should include auto theme")
	assert.Contains(t, themes, ThemeCatppuccinMocha, "Should include catppuccin-mocha theme")
	assert.Contains(t, themes, ThemeTokyoNight, "Should include tokyo-night theme")
	assert.Contains(t, themes, ThemeNord, "Should include nord theme")
	assert.Contains(t, themes, ThemeDracula, "Should include dracula theme")
	assert.Contains(t, themes, ThemeGruvboxDark, "Should include gruvbox-dark theme")
	assert.Contains(t, themes, ThemeGruvboxLight, "Should include gruvbox-light theme")

	// Should have at least 8 themes
	assert.GreaterOrEqual(t, len(themes), 8, "Should have at least 8 themes")
}

func TestGetTheme_Auto(t *testing.T) {
	theme := GetTheme(ThemeAuto)

	assert.Equal(t, ThemeAuto, theme.Name)
	assert.NotNil(t, theme.Primary, "Primary color should be set")
	assert.NotNil(t, theme.Secondary, "Secondary color should be set")
	assert.NotNil(t, theme.Success, "Success color should be set")

	// Auto theme should use AdaptiveColor
	_, ok := theme.Primary.(lipgloss.AdaptiveColor)
	assert.True(t, ok, "Auto theme should use AdaptiveColor for Primary")
}

func TestGetTheme_CatppuccinMocha(t *testing.T) {
	theme := GetTheme(ThemeCatppuccinMocha)

	assert.Equal(t, ThemeCatppuccinMocha, theme.Name)
	assert.NotNil(t, theme.Primary)
	assert.NotNil(t, theme.Secondary)
	assert.NotNil(t, theme.Success)
	assert.NotNil(t, theme.Warning)
	assert.NotNil(t, theme.Error)
	assert.NotNil(t, theme.Info)
	assert.NotNil(t, theme.Muted)
	assert.NotNil(t, theme.Highlight)
	assert.NotNil(t, theme.Background)

	// Should use fixed Color, not AdaptiveColor
	_, ok := theme.Primary.(lipgloss.Color)
	assert.True(t, ok, "Catppuccin Mocha should use fixed Color")
}

func TestGetTheme_CatppuccinLatte(t *testing.T) {
	theme := GetTheme(ThemeCatppuccinLatte)

	assert.Equal(t, ThemeCatppuccinLatte, theme.Name)
	assert.NotNil(t, theme.Primary)

	// Light theme should have different colors than dark theme
	mochaTheme := GetTheme(ThemeCatppuccinMocha)
	assert.NotEqual(t, theme.Primary, mochaTheme.Primary, "Light and dark themes should have different colors")
}

func TestGetTheme_TokyoNight(t *testing.T) {
	theme := GetTheme(ThemeTokyoNight)

	assert.Equal(t, ThemeTokyoNight, theme.Name)
	assert.NotNil(t, theme.Primary)
	assert.NotNil(t, theme.Secondary)
}

func TestGetTheme_Nord(t *testing.T) {
	theme := GetTheme(ThemeNord)

	assert.Equal(t, ThemeNord, theme.Name)
	assert.NotNil(t, theme.Primary)
	assert.NotNil(t, theme.Secondary)
}

func TestGetTheme_Dracula(t *testing.T) {
	theme := GetTheme(ThemeDracula)

	assert.Equal(t, ThemeDracula, theme.Name)
	assert.NotNil(t, theme.Primary)
	assert.NotNil(t, theme.Secondary)
}

func TestGetTheme_GruvboxDark(t *testing.T) {
	theme := GetTheme(ThemeGruvboxDark)

	assert.Equal(t, ThemeGruvboxDark, theme.Name)
	assert.NotNil(t, theme.Primary)
	assert.NotNil(t, theme.Secondary)
}

func TestGetTheme_GruvboxLight(t *testing.T) {
	theme := GetTheme(ThemeGruvboxLight)

	assert.Equal(t, ThemeGruvboxLight, theme.Name)
	assert.NotNil(t, theme.Primary)

	// Light theme should differ from dark theme
	darkTheme := GetTheme(ThemeGruvboxDark)
	assert.NotEqual(t, theme.Primary, darkTheme.Primary, "Gruvbox light and dark should differ")
}

func TestGetTheme_Invalid(t *testing.T) {
	// Invalid theme name should default to auto
	theme := GetTheme(ThemeName("invalid-theme"))

	assert.Equal(t, ThemeAuto, theme.Name, "Invalid theme should default to auto")
}

func TestSetTheme(t *testing.T) {
	// Save original theme
	originalTheme := GetCurrentTheme()
	defer SetTheme(originalTheme.Name) // Restore after test

	// Set to Catppuccin Mocha
	SetTheme(ThemeCatppuccinMocha)
	currentTheme := GetCurrentTheme()
	assert.Equal(t, ThemeCatppuccinMocha, currentTheme.Name)

	// Set to Tokyo Night
	SetTheme(ThemeTokyoNight)
	currentTheme = GetCurrentTheme()
	assert.Equal(t, ThemeTokyoNight, currentTheme.Name)

	// Colors should be updated
	assert.NotNil(t, PrimaryColor)
	assert.NotNil(t, SecondaryColor)
}

func TestInitTheme_WithEnvironment(t *testing.T) {
	// Save original theme and env
	originalTheme := GetCurrentTheme()
	originalEnv := os.Getenv("DVM_THEME")
	defer func() {
		SetTheme(originalTheme.Name)
		if originalEnv != "" {
			os.Setenv("DVM_THEME", originalEnv)
		} else {
			os.Unsetenv("DVM_THEME")
		}
	}()

	// Test with environment variable
	os.Setenv("DVM_THEME", "dracula")
	InitTheme()

	currentTheme := GetCurrentTheme()
	assert.Equal(t, ThemeDracula, currentTheme.Name, "Should use theme from DVM_THEME env var")
}

func TestInitTheme_WithoutEnvironment(t *testing.T) {
	// Save original env
	originalEnv := os.Getenv("DVM_THEME")
	defer func() {
		if originalEnv != "" {
			os.Setenv("DVM_THEME", originalEnv)
		}
	}()

	// Clear environment variable
	os.Unsetenv("DVM_THEME")

	// Reset to auto theme first
	SetTheme(ThemeAuto)

	// InitTheme should keep current theme if no env var
	InitTheme()
	currentTheme := GetCurrentTheme()
	assert.Equal(t, ThemeAuto, currentTheme.Name, "Should remain auto when no env var set")
}

func TestThemeColorsNotNil(t *testing.T) {
	// Test all themes to ensure no nil colors
	themes := AvailableThemes()

	for _, themeName := range themes {
		t.Run(string(themeName), func(t *testing.T) {
			theme := GetTheme(themeName)

			assert.NotNil(t, theme.Primary, "Primary should not be nil for %s", themeName)
			assert.NotNil(t, theme.Secondary, "Secondary should not be nil for %s", themeName)
			assert.NotNil(t, theme.Success, "Success should not be nil for %s", themeName)
			assert.NotNil(t, theme.Warning, "Warning should not be nil for %s", themeName)
			assert.NotNil(t, theme.Error, "Error should not be nil for %s", themeName)
			assert.NotNil(t, theme.Info, "Info should not be nil for %s", themeName)
			assert.NotNil(t, theme.Muted, "Muted should not be nil for %s", themeName)
			assert.NotNil(t, theme.Highlight, "Highlight should not be nil for %s", themeName)
			assert.NotNil(t, theme.Background, "Background should not be nil for %s", themeName)
		})
	}
}

func TestStylesInitialized(t *testing.T) {
	// After theme initialization, all styles should be set
	SetTheme(ThemeAuto)

	// Test that style variables are not zero values
	assert.NotEqual(t, lipgloss.Style{}, HeaderStyle, "HeaderStyle should be initialized")
	assert.NotEqual(t, lipgloss.Style{}, ActiveStyle, "ActiveStyle should be initialized")
	assert.NotEqual(t, lipgloss.Style{}, TextStyle, "TextStyle should be initialized")
	assert.NotEqual(t, lipgloss.Style{}, MutedStyle, "MutedStyle should be initialized")
	assert.NotEqual(t, lipgloss.Style{}, SuccessStyle, "SuccessStyle should be initialized")
	assert.NotEqual(t, lipgloss.Style{}, ErrorStyle, "ErrorStyle should be initialized")
	assert.NotEqual(t, lipgloss.Style{}, WarningStyle, "WarningStyle should be initialized")
	assert.NotEqual(t, lipgloss.Style{}, InfoStyle, "InfoStyle should be initialized")
}

func TestColorVariablesUpdated(t *testing.T) {
	// Save original
	originalTheme := GetCurrentTheme()
	defer SetTheme(originalTheme.Name)

	// Set a specific theme
	SetTheme(ThemeDracula)

	// Color variables should be updated
	assert.NotNil(t, PrimaryColor, "PrimaryColor should be set")
	assert.NotNil(t, SecondaryColor, "SecondaryColor should be set")
	assert.NotNil(t, SuccessColor, "SuccessColor should be set")
	assert.NotNil(t, WarningColor, "WarningColor should be set")
	assert.NotNil(t, ErrorColor, "ErrorColor should be set")
	assert.NotNil(t, InfoColor, "InfoColor should be set")
	assert.NotNil(t, MutedColor, "MutedColor should be set")
	assert.NotNil(t, HighlightColor, "HighlightColor should be set")
	assert.NotNil(t, BackgroundColor, "BackgroundColor should be set")
}

func TestThemeSwitching(t *testing.T) {
	// Save original
	originalTheme := GetCurrentTheme()
	defer SetTheme(originalTheme.Name)

	// Switch through multiple themes
	themes := []ThemeName{
		ThemeCatppuccinMocha,
		ThemeTokyoNight,
		ThemeNord,
		ThemeDracula,
		ThemeGruvboxDark,
	}

	for _, themeName := range themes {
		SetTheme(themeName)
		current := GetCurrentTheme()
		assert.Equal(t, themeName, current.Name, "Theme should switch to %s", themeName)

		// Verify colors are updated
		assert.NotNil(t, PrimaryColor, "Colors should be set for %s", themeName)
	}
}

func BenchmarkGetTheme(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetTheme(ThemeCatppuccinMocha)
	}
}

func BenchmarkSetTheme(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SetTheme(ThemeCatppuccinMocha)
	}
}
