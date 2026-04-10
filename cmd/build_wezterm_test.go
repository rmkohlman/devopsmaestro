package cmd

import (
	"testing"

	theme "github.com/rmkohlman/MaestroTheme"
)

// =============================================================================
// themeToWeztermColors tests
// =============================================================================

func TestThemeToWeztermColors_NilTheme_ReturnsNil(t *testing.T) {
	result := themeToWeztermColors(nil)
	if result != nil {
		t.Errorf("themeToWeztermColors(nil) = %+v, want nil", result)
	}
}

func TestThemeToWeztermColors_ThemeWithNoColors_ReturnsNil(t *testing.T) {
	// A theme with no Colors map will produce no terminal colors
	thm := &theme.Theme{
		Name:     "empty-theme",
		Category: "dark",
		Colors:   nil,
	}
	result := themeToWeztermColors(thm)
	// ToTerminalColors() on a theme with no colors returns empty map → nil
	if result != nil {
		t.Errorf("themeToWeztermColors(empty) = %+v, want nil (no terminal colors)", result)
	}
}

func TestThemeToWeztermColors_ThemeWithColors_ReturnsNonNil(t *testing.T) {
	thm := &theme.Theme{
		Name:     "test-theme",
		Category: "dark",
		Colors: map[string]string{
			"bg":    "#1a1b26",
			"fg":    "#c0caf5",
			"red":   "#f7768e",
			"green": "#9ece6a",
		},
	}
	result := themeToWeztermColors(thm)
	if result == nil {
		t.Fatal("themeToWeztermColors(theme with colors) = nil, want non-nil ColorConfig")
	}
}

// TestThemeToWeztermColors_MapsBackgroundAndForeground verifies that bg/fg
// map to Background and Foreground fields respectively.
func TestThemeToWeztermColors_MapsBackgroundAndForeground(t *testing.T) {
	thm := &theme.Theme{
		Name:     "tokyonight-test",
		Category: "dark",
		Colors: map[string]string{
			"bg":    "#1a1b26",
			"fg":    "#c0caf5",
			"red":   "#f7768e",
			"green": "#9ece6a",
			"blue":  "#7aa2f7",
		},
	}
	result := themeToWeztermColors(thm)
	if result == nil {
		t.Fatal("expected non-nil ColorConfig")
	}
	if result.Background != "#1a1b26" {
		t.Errorf("Background = %q, want %q", result.Background, "#1a1b26")
	}
	if result.Foreground != "#c0caf5" {
		t.Errorf("Foreground = %q, want %q", result.Foreground, "#c0caf5")
	}
}

// TestThemeToWeztermColors_ANSISliceHas8Entries verifies the ANSI array is always 8 entries.
func TestThemeToWeztermColors_ANSISliceHas8Entries(t *testing.T) {
	thm := &theme.Theme{
		Name:     "test-theme",
		Category: "dark",
		Colors: map[string]string{
			"bg":  "#1a1b26",
			"fg":  "#c0caf5",
			"red": "#f7768e",
		},
	}
	result := themeToWeztermColors(thm)
	if result == nil {
		t.Fatal("expected non-nil ColorConfig")
	}
	if len(result.ANSI) != 8 {
		t.Errorf("len(ANSI) = %d, want 8", len(result.ANSI))
	}
	if len(result.Brights) != 8 {
		t.Errorf("len(Brights) = %d, want 8", len(result.Brights))
	}
}

// TestThemeToWeztermColors_FallbackDefaultsUsed verifies that when a color is
// missing from the theme, the TokyoNight-inspired defaults are used.
func TestThemeToWeztermColors_FallbackDefaultsUsed(t *testing.T) {
	// Theme only has bg/fg — ANSI colors should fall back to defaults
	thm := &theme.Theme{
		Name:     "minimal-theme",
		Category: "dark",
		Colors: map[string]string{
			"bg": "#000000",
			"fg": "#ffffff",
		},
	}
	result := themeToWeztermColors(thm)
	if result == nil {
		t.Fatal("expected non-nil ColorConfig")
	}
	// With only bg/fg, red should use the fallback "#f7768e"
	if result.ANSI[1] != "#f7768e" {
		t.Errorf("ANSI[1] (red fallback) = %q, want %q", result.ANSI[1], "#f7768e")
	}
	// Background should use the theme's bg
	if result.Background != "#000000" {
		t.Errorf("Background = %q, want %q", result.Background, "#000000")
	}
}

// TestThemeToWeztermColors_CustomColorsOverrideFallbacks verifies custom theme
// colors override the default fallbacks.
func TestThemeToWeztermColors_CustomColorsOverrideFallbacks(t *testing.T) {
	customRed := "#ff0000"
	thm := &theme.Theme{
		Name:     "custom-theme",
		Category: "dark",
		Colors: map[string]string{
			"bg":  "#1a1b26",
			"fg":  "#c0caf5",
			"red": customRed, // Should be mapped to ansi_red via ToTerminalColors
		},
	}
	result := themeToWeztermColors(thm)
	if result == nil {
		t.Fatal("expected non-nil ColorConfig")
	}
	// ANSI[1] is red — should use our custom red
	if result.ANSI[1] != customRed {
		t.Errorf("ANSI[1] (custom red) = %q, want %q", result.ANSI[1], customRed)
	}
}

// TestThemeToWeztermColors_ANSIOrder verifies the exact ANSI order:
// [black, red, green, yellow, blue, magenta, cyan, white]
// When a theme only has bg/fg, the palette resolver maps ansi_black → bg
// and ansi_white → fg (since those are the closest semantic matches).
// Only colors without any semantic fallback use the hardcoded defaults.
func TestThemeToWeztermColors_ANSIOrder(t *testing.T) {
	tests := []struct {
		name  string
		index int
		want  string
	}{
		// ansi_black falls back to ColorBg (#1a1b26) when bg is the only color
		{name: "black", index: 0, want: "#1a1b26"},
		{name: "red", index: 1, want: "#f7768e"},
		{name: "green", index: 2, want: "#9ece6a"},
		{name: "yellow", index: 3, want: "#e0af68"},
		{name: "blue", index: 4, want: "#7aa2f7"},
		{name: "magenta", index: 5, want: "#bb9af7"},
		{name: "cyan", index: 6, want: "#7dcfff"},
		// ansi_white falls back to ColorFg (#c0caf5) when fg is the only color
		{name: "white", index: 7, want: "#c0caf5"},
	}

	// Use a minimal theme so all ANSI colors fall back to palette semantics or defaults
	thm := &theme.Theme{
		Name:     "test-order",
		Category: "dark",
		Colors: map[string]string{
			"bg": "#1a1b26",
			"fg": "#c0caf5",
		},
	}
	result := themeToWeztermColors(thm)
	if result == nil {
		t.Fatal("expected non-nil ColorConfig")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result.ANSI[tt.index] != tt.want {
				t.Errorf("ANSI[%d] (%s) = %q, want %q",
					tt.index, tt.name, result.ANSI[tt.index], tt.want)
			}
		})
	}
}

// TestThemeToWeztermColors_BrightsOrder verifies the bright ANSI color order:
// [bright_black, bright_red, bright_green, bright_yellow, bright_blue, bright_magenta, bright_cyan, bright_white]
func TestThemeToWeztermColors_BrightsOrder(t *testing.T) {
	tests := []struct {
		name     string
		index    int
		fallback string
	}{
		{name: "bright_black", index: 0, fallback: "#414868"},
		{name: "bright_red", index: 1, fallback: "#f7768e"},
		{name: "bright_green", index: 2, fallback: "#9ece6a"},
		{name: "bright_yellow", index: 3, fallback: "#e0af68"},
		{name: "bright_blue", index: 4, fallback: "#7aa2f7"},
		{name: "bright_magenta", index: 5, fallback: "#bb9af7"},
		{name: "bright_cyan", index: 6, fallback: "#7dcfff"},
		{name: "bright_white", index: 7, fallback: "#c0caf5"},
	}

	thm := &theme.Theme{
		Name:     "test-brights-order",
		Category: "dark",
		Colors: map[string]string{
			"bg": "#1a1b26",
			"fg": "#c0caf5",
		},
	}
	result := themeToWeztermColors(thm)
	if result == nil {
		t.Fatal("expected non-nil ColorConfig")
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if result.Brights[tt.index] != tt.fallback {
				t.Errorf("Brights[%d] (%s) = %q, want fallback %q",
					tt.index, tt.name, result.Brights[tt.index], tt.fallback)
			}
		})
	}
}
