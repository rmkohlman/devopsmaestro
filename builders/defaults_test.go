package builders

import (
	"testing"
)

func TestGetContainerDefaults(t *testing.T) {
	defaults := GetContainerDefaults()

	// Check expected keys exist
	if defaults["user"] != "dev" {
		t.Errorf("Expected default user to be 'dev', got %v", defaults["user"])
	}

	if defaults["uid"] != 1000 {
		t.Errorf("Expected default uid to be 1000, got %v", defaults["uid"])
	}

	if defaults["gid"] != 1000 {
		t.Errorf("Expected default gid to be 1000, got %v", defaults["gid"])
	}

	if defaults["workingDir"] != "/workspace" {
		t.Errorf("Expected default workingDir to be '/workspace', got %v", defaults["workingDir"])
	}

	// Check command is a slice
	if command, ok := defaults["command"].([]string); !ok {
		t.Errorf("Expected command to be []string, got %T", defaults["command"])
	} else if len(command) != 2 || command[0] != "/bin/zsh" || command[1] != "-l" {
		t.Errorf("Expected command to be ['/bin/zsh', '-l'], got %v", command)
	}

	// Check that all expected keys are present
	expectedKeys := []string{"user", "uid", "gid", "workingDir", "command"}
	for _, key := range expectedKeys {
		if _, exists := defaults[key]; !exists {
			t.Errorf("Expected key '%s' not found in defaults", key)
		}
	}
}
