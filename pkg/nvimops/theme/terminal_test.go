package theme

import (
	"strings"
	"testing"
)

func TestTheme_TerminalEnvVars(t *testing.T) {
	theme := &Theme{
		Name:     "test-theme",
		Category: "dark",
		Plugin:   ThemePlugin{Repo: "test/test"},
		Colors: map[string]string{
			// Basic UI colors
			"bg":     "#1a1b26",
			"fg":     "#c0caf5",
			"red":    "#f7768e",
			"green":  "#9ece6a",
			"blue":   "#7aa2f7",
			"yellow": "#e0af68",
		},
	}

	envVars := theme.TerminalEnvVars()

	// Test theme metadata
	if envVars["DVM_THEME"] != "test-theme" {
		t.Errorf("expected DVM_THEME=test-theme, got %s", envVars["DVM_THEME"])
	}
	if envVars["DVM_THEME_CATEGORY"] != "dark" {
		t.Errorf("expected DVM_THEME_CATEGORY=dark, got %s", envVars["DVM_THEME_CATEGORY"])
	}

	// Test UI colors are included
	if envVars["DVM_COLOR_BG"] != "#1a1b26" {
		t.Errorf("expected DVM_COLOR_BG=#1a1b26, got %s", envVars["DVM_COLOR_BG"])
	}
	if envVars["DVM_COLOR_FG"] != "#c0caf5" {
		t.Errorf("expected DVM_COLOR_FG=#c0caf5, got %s", envVars["DVM_COLOR_FG"])
	}

	// Test that we get some ANSI colors mapped
	// The exact mapping depends on ToTerminalColors() implementation
	// We'll just verify that some color keys exist
	found := false
	for key := range envVars {
		if strings.HasPrefix(key, "DVM_COLOR_") {
			found = true
			break
		}
	}
	if !found {
		t.Error("no DVM_COLOR_ environment variables found")
	}
}

func TestTheme_TerminalEnvVars_NilTheme(t *testing.T) {
	var theme *Theme
	envVars := theme.TerminalEnvVars()

	if envVars == nil {
		t.Error("expected empty map, got nil")
	}
	if len(envVars) != 0 {
		t.Errorf("expected empty map, got %d items", len(envVars))
	}
}

func TestTheme_TerminalEnvVars_EmptyCategory(t *testing.T) {
	theme := &Theme{
		Name:   "test-theme",
		Plugin: ThemePlugin{Repo: "test/test"},
		Colors: map[string]string{"bg": "#000000"},
	}

	envVars := theme.TerminalEnvVars()

	if envVars["DVM_THEME_CATEGORY"] != "unknown" {
		t.Errorf("expected DVM_THEME_CATEGORY=unknown for empty category, got %s", envVars["DVM_THEME_CATEGORY"])
	}
}

func TestTerminalKeyToEnvKey(t *testing.T) {
	tests := []struct {
		termKey  string
		expected string
	}{
		// ANSI normal colors
		{"ansi_red", "DVM_COLOR_RED"},
		{"ansi_green", "DVM_COLOR_GREEN"},
		{"ansi_blue", "DVM_COLOR_BLUE"},
		{"ansi_black", "DVM_COLOR_BLACK"},
		{"ansi_white", "DVM_COLOR_WHITE"},
		{"ansi_yellow", "DVM_COLOR_YELLOW"},
		{"ansi_magenta", "DVM_COLOR_MAGENTA"},
		{"ansi_cyan", "DVM_COLOR_CYAN"},

		// ANSI bright colors
		{"ansi_bright_red", "DVM_COLOR_BRIGHT_RED"},
		{"ansi_bright_green", "DVM_COLOR_BRIGHT_GREEN"},
		{"ansi_bright_blue", "DVM_COLOR_BRIGHT_BLUE"},
		{"ansi_bright_black", "DVM_COLOR_BRIGHT_BLACK"},
		{"ansi_bright_white", "DVM_COLOR_BRIGHT_WHITE"},
		{"ansi_bright_yellow", "DVM_COLOR_BRIGHT_YELLOW"},
		{"ansi_bright_magenta", "DVM_COLOR_BRIGHT_MAGENTA"},
		{"ansi_bright_cyan", "DVM_COLOR_BRIGHT_CYAN"},

		// UI colors
		{"bg", "DVM_COLOR_BG"},
		{"fg", "DVM_COLOR_FG"},
		{"cursor", "DVM_COLOR_CURSOR"},
		{"cursor_text", "DVM_COLOR_CURSOR_TEXT"},
		{"selection", "DVM_COLOR_SELECTION"},
		{"selection_text", "DVM_COLOR_SELECTION_TEXT"},

		// Other colors (should get generic mapping)
		{"border", "DVM_COLOR_BORDER"},
		{"comment", "DVM_COLOR_COMMENT"},
	}

	for _, tt := range tests {
		t.Run(tt.termKey, func(t *testing.T) {
			got := terminalKeyToEnvKey(tt.termKey)
			if got != tt.expected {
				t.Errorf("terminalKeyToEnvKey(%s) = %s, want %s", tt.termKey, got, tt.expected)
			}
		})
	}
}

func TestGetTerminalEnvVarsForTheme(t *testing.T) {
	// Create a memory store with a test theme
	store := NewMemoryStore()
	theme := &Theme{
		Name:     "test-theme",
		Category: "dark",
		Plugin:   ThemePlugin{Repo: "test/test"},
		Colors: map[string]string{
			"bg": "#1a1b26",
			"fg": "#c0caf5",
		},
	}

	err := store.Save(theme)
	if err != nil {
		t.Fatalf("failed to save theme: %v", err)
	}

	// Test successful retrieval
	envVars, err := GetTerminalEnvVarsForTheme(store, "test-theme")
	if err != nil {
		t.Fatalf("GetTerminalEnvVarsForTheme failed: %v", err)
	}

	if envVars["DVM_THEME"] != "test-theme" {
		t.Errorf("expected DVM_THEME=test-theme, got %s", envVars["DVM_THEME"])
	}
}

func TestGetTerminalEnvVarsForTheme_ThemeNotFound(t *testing.T) {
	store := NewMemoryStore()

	_, err := GetTerminalEnvVarsForTheme(store, "nonexistent-theme")
	if err == nil {
		t.Error("expected error for nonexistent theme")
	}
}

func TestGetTerminalEnvVarsForTheme_NilStore(t *testing.T) {
	_, err := GetTerminalEnvVarsForTheme(nil, "test-theme")
	if err == nil {
		t.Error("expected error for nil store")
	}
	if !strings.Contains(err.Error(), "no theme store provided") {
		t.Errorf("expected 'no theme store provided' error, got: %v", err)
	}
}

// TestTerminalEnvVars_WithRealTheme tests with a more complete theme structure
// similar to what's in the library themes
func TestTerminalEnvVars_WithRealTheme(t *testing.T) {
	theme := &Theme{
		Name:     "tokyonight-night",
		Category: "dark",
		Plugin:   ThemePlugin{Repo: "folke/tokyonight.nvim"},
		Style:    "night",
		Colors: map[string]string{
			"bg":      "#1a1b26",
			"bg_dark": "#16161e",
			"fg":      "#c0caf5",
			"fg_dark": "#a9b1d6",
			"red":     "#f7768e",
			"green":   "#9ece6a",
			"blue":    "#7aa2f7",
			"yellow":  "#e0af68",
			"cyan":    "#7dcfff",
			"magenta": "#bb9af7",
			"orange":  "#ff9e64",
			"purple":  "#9d7cd8",
			"teal":    "#1abc9c",
			"comment": "#565f89",
		},
	}

	envVars := theme.TerminalEnvVars()

	// Test that we get the expected metadata
	if envVars["DVM_THEME"] != "tokyonight-night" {
		t.Errorf("expected DVM_THEME=tokyonight-night, got %s", envVars["DVM_THEME"])
	}
	if envVars["DVM_THEME_CATEGORY"] != "dark" {
		t.Errorf("expected DVM_THEME_CATEGORY=dark, got %s", envVars["DVM_THEME_CATEGORY"])
	}

	// Test that we get basic UI colors
	expectedColors := map[string]string{
		"DVM_COLOR_BG": "#1a1b26",
		"DVM_COLOR_FG": "#c0caf5",
	}

	for envKey, expectedColor := range expectedColors {
		if envVars[envKey] != expectedColor {
			t.Errorf("expected %s=%s, got %s", envKey, expectedColor, envVars[envKey])
		}
	}

	// Verify we have some color variables (the exact mapping depends on ToTerminalColors)
	colorCount := 0
	for key := range envVars {
		if strings.HasPrefix(key, "DVM_COLOR_") {
			colorCount++
		}
	}

	if colorCount == 0 {
		t.Error("no color environment variables generated")
	}

	// The exact count will depend on how ToTerminalColors() maps the theme
	// but we should have at least a few colors
	if colorCount < 3 {
		t.Errorf("expected at least 3 color variables, got %d", colorCount)
	}
}
