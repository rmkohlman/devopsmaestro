package theme

import (
	"strings"
	"testing"
)

func TestGenerator_StandaloneTheme(t *testing.T) {
	generator := NewGenerator()

	// Create a standalone theme
	standaloneTheme := &Theme{
		Name:        "test-standalone",
		Description: "Test standalone theme",
		Plugin:      ThemePlugin{Repo: ""}, // Empty repo = standalone
		Colors: map[string]string{
			"bg":           "#1a1b26",
			"fg":           "#c0caf5",
			"red":          "#f7768e",
			"green":        "#9ece6a",
			"blue":         "#7aa2f7",
			"yellow":       "#e0af68",
			"purple":       "#9d7cd8",
			"cyan":         "#7dcfff",
			"orange":       "#ff9e64",
			"comment":      "#565f89",
			"bg_highlight": "#292e42",
			"bg_visual":    "#283457",
			"bg_search":    "#3d59a1",
			"error":        "#f7768e",
			"warning":      "#e0af68",
			"info":         "#7aa2f7",
			"hint":         "#1abc9c",
		},
	}

	// Generate theme files
	generated, err := generator.Generate(standaloneTheme)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that standalone colorscheme was generated
	if generated.ColorschemeLua == "" {
		t.Error("ColorschemeLua should not be empty for standalone theme")
	}

	// Check that the colorscheme contains the expected elements
	colorscheme := generated.ColorschemeLua
	expectedElements := []string{
		"vim.g.colors_name = \"test-standalone\"",
		"vim.api.nvim_set_hl(0, name, val)",
		"require(\"theme.palette\")",
		"colors.bg",
		"colors.fg",
		"hl(\"Normal\"",
		"hl(\"Comment\"",
		"hl(\"DiagnosticError\"",
		"M.setup()",
	}

	for _, element := range expectedElements {
		if !strings.Contains(colorscheme, element) {
			t.Errorf("Colorscheme should contain %q, got:\n%s", element, colorscheme)
		}
	}

	// Check that plugin spec is correct for standalone
	pluginSpec := generated.PluginLua
	expectedInPlugin := []string{
		"dir = vim.fn.stdpath(\"config\") .. \"/lua/theme\"",
		"name = \"test-standalone\"",
		"require(\"theme.colorscheme\").setup()",
	}

	for _, element := range expectedInPlugin {
		if !strings.Contains(pluginSpec, element) {
			t.Errorf("Plugin spec should contain %q, got:\n%s", element, pluginSpec)
		}
	}

	// Should not contain external repo reference
	if strings.Contains(pluginSpec, "folke/") || strings.Contains(pluginSpec, "github.com") {
		t.Errorf("Standalone plugin spec should not reference external repos, got:\n%s", pluginSpec)
	}
}

func TestGenerator_PluginTheme_StillWorks(t *testing.T) {
	generator := NewGenerator()

	// Create a regular plugin-based theme
	pluginTheme := &Theme{
		Name:        "test-plugin",
		Description: "Test plugin theme",
		Plugin:      ThemePlugin{Repo: "folke/tokyonight.nvim"}, // Has repo = plugin-based
		Style:       "night",
		Colors: map[string]string{
			"bg": "#1a1b26",
			"fg": "#c0caf5",
		},
	}

	// Generate theme files
	generated, err := generator.Generate(pluginTheme)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check that standalone colorscheme was NOT generated
	if generated.ColorschemeLua != "" {
		t.Error("ColorschemeLua should be empty for plugin-based theme")
	}

	// Check that plugin spec references external repo
	pluginSpec := generated.PluginLua
	if !strings.Contains(pluginSpec, "folke/tokyonight.nvim") {
		t.Error("Plugin-based theme should reference external repo")
	}

	// Should contain setup call for external plugin
	if !strings.Contains(pluginSpec, "require(\"tokyonight\").setup") {
		t.Error("Plugin-based theme should call external plugin setup")
	}
}

func TestTheme_IsStandalone(t *testing.T) {
	tests := []struct {
		name     string
		theme    *Theme
		expected bool
	}{
		{
			name: "plugin-based theme",
			theme: &Theme{
				Plugin: ThemePlugin{Repo: "folke/tokyonight.nvim"},
			},
			expected: false,
		},
		{
			name: "standalone theme",
			theme: &Theme{
				Plugin: ThemePlugin{Repo: ""},
			},
			expected: true,
		},
		{
			name: "nil theme",
			theme: &Theme{
				Plugin: ThemePlugin{},
			},
			expected: true, // Empty repo = standalone
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.theme.IsStandalone(); got != tt.expected {
				t.Errorf("IsStandalone() = %v, want %v", got, tt.expected)
			}
		})
	}
}
