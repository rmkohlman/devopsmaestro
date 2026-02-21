package cmd

import (
	"devopsmaestro/pkg/terminalops/wezterm"
	"strings"
	"testing"
)

func TestGenerateShellConfig(t *testing.T) {
	t.Skip("generateShellConfig now requires a database connection - needs integration test setup")

	// TODO: Update this test to use proper database mocking
	// The function signature is now: generateShellConfig(stagingDir, appName, workspaceName string, ds db.DataStore)
	// This requires more complex test setup with database mocking
}

func TestCreateDefaultTerminalPrompt(t *testing.T) {
	appName := "test-app"
	workspaceName := "test-workspace"
	prompt := createDefaultTerminalPrompt(appName, workspaceName)

	if prompt == nil {
		t.Fatal("createDefaultTerminalPrompt returned nil")
	}

	if prompt.Kind != "TerminalPrompt" {
		t.Errorf("Expected Kind 'TerminalPrompt', got %s", prompt.Kind)
	}

	expectedName := "dvm-default-test-app-test-workspace"
	if prompt.Metadata.Name != expectedName {
		t.Errorf("Expected Name '%s', got %s", expectedName, prompt.Metadata.Name)
	}

	// Verify expected modules exist
	expectedModules := []string{"custom.dvm", "directory", "character"}
	for _, module := range expectedModules {
		if _, exists := prompt.Spec.Modules[module]; !exists {
			t.Errorf("Expected module %s not found", module)
		}
	}

	// Verify custom.dvm module has the app name command
	customDvm := prompt.Spec.Modules["custom.dvm"]
	expectedCommand := `echo '[test-app]'`
	if command, ok := customDvm.Options["command"].(string); !ok || command != expectedCommand {
		t.Errorf("Expected command %q, got %q", expectedCommand, command)
	}
}

func TestCreateDefaultPalette(t *testing.T) {
	palette := createDefaultPalette()

	if palette == nil {
		t.Fatal("createDefaultPalette returned nil")
	}

	if palette.Name != "default" {
		t.Errorf("Expected Name 'default', got %s", palette.Name)
	}

	// Verify essential colors exist
	essentialColors := []string{
		"bg", "fg",
		"ansi_red", "ansi_green", "ansi_blue", "ansi_cyan",
		"error", "warning", "info", "success",
	}

	for _, color := range essentialColors {
		if palette.Get(color) == "" {
			t.Errorf("Essential color %s is missing or empty", color)
		}
	}

	// Verify colors are valid hex format
	for key, color := range palette.Colors {
		if color != "" && !strings.HasPrefix(color, "#") {
			t.Errorf("Color %s (%s) should be in hex format", key, color)
		}
	}
}

func TestMapConfigToWezTerm(t *testing.T) {
	// Test data
	config := map[string]any{
		"font": map[string]any{
			"family": "JetBrains Mono",
			"size":   16.0,
		},
		"window": map[string]any{
			"opacity":     0.9,
			"blur":        10.0,
			"decorations": "RESIZE",
			"initialRows": 24.0,
			"initialCols": 80.0,
		},
		"colors": map[string]any{
			"foreground":   "#ffffff",
			"background":   "#000000",
			"cursor_bg":    "#ff0000",
			"cursor_fg":    "#00ff00",
			"selection_bg": "#0000ff",
			"selection_fg": "#ffff00",
			"ansi":         []any{"#000000", "#ff0000", "#00ff00", "#ffff00", "#0000ff", "#ff00ff", "#00ffff", "#ffffff"},
			"brights":      []any{"#808080", "#ff8080", "#80ff80", "#ffff80", "#8080ff", "#ff80ff", "#80ffff", "#ffffff"},
		},
		"themeRef":   "tokyo-night",
		"scrollback": 5000.0,
		"leader": map[string]any{
			"key":     "a",
			"mods":    "CTRL",
			"timeout": 1000.0,
		},
		"keys": []any{
			map[string]any{
				"key":    "c",
				"mods":   "CTRL",
				"action": "CopyTo",
				"args":   "Clipboard",
			},
		},
		"tabBar": map[string]any{
			"enabled":  true,
			"position": "Bottom",
			"maxWidth": 32.0,
		},
		"pane": map[string]any{
			"inactiveSaturation": 0.8,
			"inactiveBrightness": 0.6,
		},
	}

	// Create WezTerm config
	wt := &wezterm.WezTerm{
		Name: "test-config",
	}

	// Test the mapping
	err := mapConfigToWezTerm(config, wt)
	if err != nil {
		t.Fatalf("mapConfigToWezTerm failed: %v", err)
	}

	// Verify font mapping
	if wt.Font.Family != "JetBrains Mono" {
		t.Errorf("Expected font family 'JetBrains Mono', got %s", wt.Font.Family)
	}
	if wt.Font.Size != 16.0 {
		t.Errorf("Expected font size 16.0, got %f", wt.Font.Size)
	}

	// Verify window mapping
	if wt.Window.Opacity != 0.9 {
		t.Errorf("Expected window opacity 0.9, got %f", wt.Window.Opacity)
	}
	if wt.Window.Blur != 10 {
		t.Errorf("Expected window blur 10, got %d", wt.Window.Blur)
	}
	if wt.Window.Decorations != "RESIZE" {
		t.Errorf("Expected window decorations 'RESIZE', got %s", wt.Window.Decorations)
	}
	if wt.Window.InitialRows != 24 {
		t.Errorf("Expected initial rows 24, got %d", wt.Window.InitialRows)
	}
	if wt.Window.InitialCols != 80 {
		t.Errorf("Expected initial cols 80, got %d", wt.Window.InitialCols)
	}

	// Verify colors mapping
	if wt.Colors == nil {
		t.Fatal("Colors should not be nil")
	}
	if wt.Colors.Foreground != "#ffffff" {
		t.Errorf("Expected foreground color '#ffffff', got %s", wt.Colors.Foreground)
	}
	if wt.Colors.Background != "#000000" {
		t.Errorf("Expected background color '#000000', got %s", wt.Colors.Background)
	}
	if len(wt.Colors.ANSI) != 8 {
		t.Errorf("Expected 8 ANSI colors, got %d", len(wt.Colors.ANSI))
	}
	if len(wt.Colors.Brights) != 8 {
		t.Errorf("Expected 8 bright colors, got %d", len(wt.Colors.Brights))
	}

	// Verify theme reference
	if wt.ThemeRef != "tokyo-night" {
		t.Errorf("Expected theme ref 'tokyo-night', got %s", wt.ThemeRef)
	}

	// Verify scrollback
	if wt.Scrollback != 5000 {
		t.Errorf("Expected scrollback 5000, got %d", wt.Scrollback)
	}

	// Verify leader key
	if wt.Leader == nil {
		t.Fatal("Leader should not be nil")
	}
	if wt.Leader.Key != "a" {
		t.Errorf("Expected leader key 'a', got %s", wt.Leader.Key)
	}
	if wt.Leader.Mods != "CTRL" {
		t.Errorf("Expected leader mods 'CTRL', got %s", wt.Leader.Mods)
	}
	if wt.Leader.Timeout != 1000 {
		t.Errorf("Expected leader timeout 1000, got %d", wt.Leader.Timeout)
	}

	// Verify keybindings
	if len(wt.Keys) != 1 {
		t.Errorf("Expected 1 keybinding, got %d", len(wt.Keys))
	} else {
		key := wt.Keys[0]
		if key.Key != "c" {
			t.Errorf("Expected key 'c', got %s", key.Key)
		}
		if key.Mods != "CTRL" {
			t.Errorf("Expected mods 'CTRL', got %s", key.Mods)
		}
		if key.Action != "CopyTo" {
			t.Errorf("Expected action 'CopyTo', got %s", key.Action)
		}
		if key.Args != "Clipboard" {
			t.Errorf("Expected args 'Clipboard', got %v", key.Args)
		}
	}

	// Verify tab bar
	if wt.TabBar == nil {
		t.Fatal("TabBar should not be nil")
	}
	if !wt.TabBar.Enabled {
		t.Error("Expected tab bar to be enabled")
	}
	if wt.TabBar.Position != "Bottom" {
		t.Errorf("Expected tab bar position 'Bottom', got %s", wt.TabBar.Position)
	}
	if wt.TabBar.MaxWidth != 32 {
		t.Errorf("Expected tab bar max width 32, got %d", wt.TabBar.MaxWidth)
	}

	// Verify pane config
	if wt.Pane == nil {
		t.Fatal("Pane should not be nil")
	}
	if wt.Pane.InactiveSaturation != 0.8 {
		t.Errorf("Expected inactive saturation 0.8, got %f", wt.Pane.InactiveSaturation)
	}
	if wt.Pane.InactiveBrightness != 0.6 {
		t.Errorf("Expected inactive brightness 0.6, got %f", wt.Pane.InactiveBrightness)
	}
}
