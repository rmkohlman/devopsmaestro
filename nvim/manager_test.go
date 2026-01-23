package nvim

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestNewManagerWithPath(t *testing.T) {
	customPath := "/custom/path/nvim"
	mgr := NewManagerWithPath(customPath)
	if mgr == nil {
		t.Fatal("NewManagerWithPath() returned nil")
	}
}

func TestInitMinimal(t *testing.T) {
	// Create temp directory for test
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")

	mgr := NewManager()
	opts := InitOptions{
		ConfigPath: configPath,
		Template:   "minimal",
		Overwrite:  false,
		GitClone:   false,
	}

	// Test initialization
	err := mgr.Init(opts)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Verify init.lua was created
	initLua := filepath.Join(configPath, "init.lua")
	if _, err := os.Stat(initLua); os.IsNotExist(err) {
		t.Error("init.lua was not created")
	}

	// Verify content contains expected text
	content, err := os.ReadFile(initLua)
	if err != nil {
		t.Fatalf("Failed to read init.lua: %v", err)
	}

	expectedText := "DevOpsMaestro Minimal"
	if !containsString(string(content), expectedText) {
		t.Errorf("init.lua does not contain expected text %q", expectedText)
	}
}

func TestInitOverwrite(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")

	// Create existing config
	err := os.MkdirAll(configPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	existingContent := "existing config"
	err = os.WriteFile(filepath.Join(configPath, "init.lua"), []byte(existingContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write existing config: %v", err)
	}

	// Try to init without overwrite - should fail
	mgr := NewManager()
	opts := InitOptions{
		ConfigPath: configPath,
		Template:   "minimal",
		Overwrite:  false,
	}

	err = mgr.Init(opts)
	if err == nil {
		t.Error("Init() should have failed without overwrite flag")
	}

	// Try with overwrite - should succeed
	opts.Overwrite = true
	err = mgr.Init(opts)
	if err != nil {
		t.Fatalf("Init() with overwrite failed: %v", err)
	}

	// Verify content was overwritten
	content, _ := os.ReadFile(filepath.Join(configPath, "init.lua"))
	if containsString(string(content), existingContent) {
		t.Error("Config was not overwritten")
	}
}

func TestStatus(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")

	mgr := NewManagerWithPath(configPath)

	// Status before init - should show not exists
	status, err := mgr.Status()
	if err != nil {
		t.Fatalf("Status() failed: %v", err)
	}

	if status.Exists {
		t.Error("Status.Exists should be false before init")
	}

	if status.ConfigPath != configPath {
		t.Errorf("Status.ConfigPath = %q, want %q", status.ConfigPath, configPath)
	}

	// Initialize config
	opts := InitOptions{
		ConfigPath: configPath,
		Template:   "minimal",
	}
	err = mgr.Init(opts)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Status after init - should show exists
	status, err = mgr.Status()
	if err != nil {
		t.Fatalf("Status() after init failed: %v", err)
	}

	if !status.Exists {
		t.Error("Status.Exists should be true after init")
	}

	if status.Template != "minimal" {
		t.Errorf("Status.Template = %q, want %q", status.Template, "minimal")
	}

	if status.LastSync.IsZero() {
		t.Error("Status.LastSync should not be zero after init")
	}
}

func TestStatusSaveLoad(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")

	// Create home-like structure for status file
	homeDir := filepath.Join(tmpDir, "home")
	os.Setenv("HOME", homeDir)
	defer os.Unsetenv("HOME")

	mgr := NewManagerWithPath(configPath)

	// Initialize
	opts := InitOptions{
		ConfigPath: configPath,
		Template:   "minimal",
	}
	err := mgr.Init(opts)
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Get status
	status1, _ := mgr.Status()

	// Create new manager instance
	mgr2 := NewManagerWithPath(configPath)

	// Status should be loaded from file
	status2, _ := mgr2.Status()

	if status2.Template != status1.Template {
		t.Errorf("Loaded template = %q, want %q", status2.Template, status1.Template)
	}

	// Times should be within 1 second (accounting for serialization)
	timeDiff := status1.LastSync.Sub(status2.LastSync)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("LastSync time difference too large: %v", timeDiff)
	}
}

func TestSyncDirection(t *testing.T) {
	tests := []struct {
		direction SyncDirection
		expected  string
	}{
		{SyncPull, "pull"},
		{SyncPush, "push"},
		{SyncBidirectional, "bidirectional"},
		{SyncDirection(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.direction.String()
			if result != tt.expected {
				t.Errorf("SyncDirection.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && s != "" &&
		(s == substr || (len(s) >= len(substr) && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
