package shell

import (
	"testing"
)

func TestGetDefaults(t *testing.T) {
	defaults := GetDefaults()

	// Check expected keys exist
	if defaults["type"] != "zsh" {
		t.Errorf("Expected default shell type to be 'zsh', got %v", defaults["type"])
	}

	if defaults["framework"] != "oh-my-zsh" {
		t.Errorf("Expected default framework to be 'oh-my-zsh', got %v", defaults["framework"])
	}

	if defaults["theme"] != "starship" {
		t.Errorf("Expected default theme to be 'starship', got %v", defaults["theme"])
	}

	// Check that all expected keys are present
	expectedKeys := []string{"type", "framework", "theme"}
	for _, key := range expectedKeys {
		if _, exists := defaults[key]; !exists {
			t.Errorf("Expected key '%s' not found in defaults", key)
		}
	}
}
