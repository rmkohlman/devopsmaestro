package wezterm

import (
	"strings"
	"testing"
)

func TestLuaGenerator(t *testing.T) {
	generator := NewLuaGenerator()

	// Create a test spec
	spec := &WeztermSpec{
		Font: FontConfig{
			Family: "Test Font",
			Size:   16,
		},
		Window: WindowConfig{
			Opacity:     0.9,
			Blur:        10,
			Decorations: "RESIZE",
		},
		Colors: &ColorConfig{
			Foreground:  "#ffffff",
			Background:  "#000000",
			CursorBg:    "#ff0000",
			CursorFg:    "#00ff00",
			SelectionBg: "#0000ff",
			SelectionFg: "#ffff00",
			ANSI:        []string{"#000000", "#ff0000"},
			Brights:     []string{"#808080", "#ff8080"},
		},
		Leader: &LeaderKey{
			Key:     "a",
			Mods:    "CTRL",
			Timeout: 1000,
		},
		Keys: []Keybinding{
			{
				Key:    "t",
				Mods:   "CMD",
				Action: "SpawnTab",
				Args:   "CurrentPaneDomain",
			},
		},
		Scrollback: 5000,
		Workspace:  "test",
	}

	lua, err := generator.Generate(spec)
	if err != nil {
		t.Fatalf("Failed to generate Lua: %v", err)
	}

	// Check for required elements
	requiredStrings := []string{
		"local wezterm = require(\"wezterm\")",
		"local act = wezterm.action",
		"local config = wezterm.config_builder()",
		"config.font = wezterm.font(\"Test Font\")",
		"config.font_size = 16",
		"config.window_background_opacity = 0.9",
		"config.macos_window_background_blur = 10",
		"config.window_decorations = \"RESIZE\"",
		"config.colors = {",
		"foreground = \"#ffffff\"",
		"background = \"#000000\"",
		"cursor_bg = \"#ff0000\"",
		"cursor_fg = \"#00ff00\"",
		"selection_bg = \"#0000ff\"",
		"selection_fg = \"#ffff00\"",
		"config.leader = { key = \"a\", mods = \"CTRL\", timeout_milliseconds = 1000 }",
		"config.keys = {",
		"{ key = \"t\", mods = \"CMD\", action = act.SpawnTab(\"CurrentPaneDomain\") }",
		"config.scrollback_lines = 5000",
		"config.default_workspace = \"test\"",
		"return config",
	}

	for _, required := range requiredStrings {
		if !strings.Contains(lua, required) {
			t.Errorf("Generated Lua missing required string: %s", required)
		}
	}
}

func TestLuaGeneratorFromConfig(t *testing.T) {
	generator := NewLuaGenerator()

	config := NewWezTerm("test")
	config.Font.Family = "Test Font"
	config.Font.Size = 16
	config.Window.Opacity = 0.9
	config.Scrollback = 5000

	lua, err := generator.GenerateFromConfig(config)
	if err != nil {
		t.Fatalf("Failed to generate Lua from config: %v", err)
	}

	if !strings.Contains(lua, "config.font = wezterm.font(\"Test Font\")") {
		t.Error("Generated Lua missing font configuration")
	}

	if !strings.Contains(lua, "config.font_size = 16") {
		t.Error("Generated Lua missing font size")
	}

	if !strings.Contains(lua, "config.window_background_opacity = 0.9") {
		t.Error("Generated Lua missing window opacity")
	}

	if !strings.Contains(lua, "config.scrollback_lines = 5000") {
		t.Error("Generated Lua missing scrollback lines")
	}
}

func TestFormatLuaString(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "\"\""},
		{"hello", "\"hello\""},
		{"hello \"world\"", "\"hello \\\"world\\\"\""},
		{"hello\\world", "\"hello\\\\world\""},
	}

	for _, test := range tests {
		result := formatLuaString(test.input)
		if result != test.expected {
			t.Errorf("formatLuaString(%q) = %q, expected %q", test.input, result, test.expected)
		}
	}
}

func TestFormatLuaTable(t *testing.T) {
	table := map[string]any{
		"string_val": "hello",
		"int_val":    42,
		"bool_val":   true,
		"float_val":  3.14,
	}

	result := formatLuaTable(table, 0)

	// Should be properly formatted Lua table
	if !strings.Contains(result, "{") || !strings.Contains(result, "}") {
		t.Error("Result should contain table braces")
	}

	if !strings.Contains(result, "string_val = \"hello\"") {
		t.Error("Result should contain string value")
	}

	if !strings.Contains(result, "int_val = 42") {
		t.Error("Result should contain integer value")
	}

	if !strings.Contains(result, "bool_val = true") {
		t.Error("Result should contain boolean value")
	}

	if !strings.Contains(result, "float_val = 3.14") {
		t.Error("Result should contain float value")
	}
}

func TestGeneratorNilHandling(t *testing.T) {
	generator := NewLuaGenerator()

	_, err := generator.Generate(nil)
	if err == nil {
		t.Error("Expected error when generating from nil spec")
	}

	_, err = generator.GenerateFromConfig(nil)
	if err == nil {
		t.Error("Expected error when generating from nil config")
	}
}
