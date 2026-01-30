package theme

import (
	"strings"
	"testing"
)

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()

	theme := &Theme{
		Name:        "test-theme",
		Description: "Test theme",
		Plugin:      ThemePlugin{Repo: "folke/tokyonight.nvim"},
		Style:       "night",
		Transparent: false,
		Colors: map[string]string{
			"bg":           "#1a1b26",
			"bg_dark":      "#16161e",
			"bg_highlight": "#292e42",
			"fg":           "#c0caf5",
			"fg_dark":      "#a9b1d6",
		},
	}

	generated, err := gen.Generate(theme)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	// Check palette.lua
	if !strings.Contains(generated.PaletteLua, "M.name = \"test-theme\"") {
		t.Error("PaletteLua should contain theme name")
	}
	if !strings.Contains(generated.PaletteLua, "bg = \"#1a1b26\"") {
		t.Error("PaletteLua should contain bg color")
	}
	if !strings.Contains(generated.PaletteLua, "function M.get(name, fallback)") {
		t.Error("PaletteLua should contain get function")
	}

	// Check init.lua
	if !strings.Contains(generated.InitLua, "M.palette = require(\"theme.palette\")") {
		t.Error("InitLua should require palette")
	}
	if !strings.Contains(generated.InitLua, "function M.lualine_theme()") {
		t.Error("InitLua should contain lualine_theme function")
	}
	if !strings.Contains(generated.InitLua, "function M.bufferline_highlights()") {
		t.Error("InitLua should contain bufferline_highlights function")
	}

	// Check plugin.lua
	if !strings.Contains(generated.PluginLua, "\"folke/tokyonight.nvim\"") {
		t.Error("PluginLua should contain plugin repo")
	}
	if !strings.Contains(generated.PluginLua, "priority = 1000") {
		t.Error("PluginLua should have priority 1000")
	}
	if !strings.Contains(generated.PluginLua, "local palette = require(\"theme.palette\")") {
		t.Error("PluginLua should require palette")
	}
}

func TestGenerator_Catppuccin(t *testing.T) {
	gen := NewGenerator()

	theme := &Theme{
		Name:   "catppuccin-test",
		Plugin: ThemePlugin{Repo: "catppuccin/nvim"},
		Style:  "mocha",
		Colors: map[string]string{
			"bg": "#1e1e2e",
			"fg": "#cdd6f4",
		},
	}

	generated, err := gen.Generate(theme)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(generated.PluginLua, "catppuccin/nvim") {
		t.Error("should contain catppuccin repo")
	}
	if !strings.Contains(generated.PluginLua, "flavour = \"mocha\"") {
		t.Error("should contain flavour setting")
	}
}

func TestGenerator_Gruvbox(t *testing.T) {
	gen := NewGenerator()

	theme := &Theme{
		Name:        "gruvbox-test",
		Plugin:      ThemePlugin{Repo: "ellisonleao/gruvbox.nvim"},
		Style:       "hard",
		Transparent: true,
		Colors: map[string]string{
			"bg": "#282828",
		},
	}

	generated, err := gen.Generate(theme)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if !strings.Contains(generated.PluginLua, "gruvbox.nvim") {
		t.Error("should contain gruvbox repo")
	}
	if !strings.Contains(generated.PluginLua, "transparent_mode = true") {
		t.Error("should have transparent mode enabled")
	}
}

func TestGenerator_InvalidTheme(t *testing.T) {
	gen := NewGenerator()

	// Missing name
	theme := &Theme{
		Plugin: ThemePlugin{Repo: "user/repo"},
	}

	_, err := gen.Generate(theme)
	if err == nil {
		t.Error("expected error for invalid theme")
	}
}

func TestGetColorschemeCommand(t *testing.T) {
	tests := []struct {
		theme    Theme
		expected string
	}{
		{
			theme:    Theme{Plugin: ThemePlugin{Repo: "folke/tokyonight.nvim"}, Style: "night"},
			expected: "tokyonight-night",
		},
		{
			theme:    Theme{Plugin: ThemePlugin{Repo: "folke/tokyonight.nvim"}},
			expected: "tokyonight",
		},
		{
			theme:    Theme{Plugin: ThemePlugin{Repo: "catppuccin/nvim"}, Style: "mocha"},
			expected: "catppuccin-mocha",
		},
		{
			theme:    Theme{Plugin: ThemePlugin{Repo: "unknown/awesome-theme.nvim"}, Style: ""},
			expected: "awesome-theme",
		},
	}

	for _, tt := range tests {
		got := tt.theme.GetColorschemeCommand()
		if got != tt.expected {
			t.Errorf("GetColorschemeCommand() = %s, want %s", got, tt.expected)
		}
	}
}
