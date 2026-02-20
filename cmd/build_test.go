package cmd

import (
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
	expectedCommand := `echo "[test-app]"`
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
