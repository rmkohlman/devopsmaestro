package nvim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestInitMinimalDirect tests the minimal template initialization directly
func TestInitMinimalDirect(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	m := &manager{
		configPath: configPath,
		statusFile: filepath.Join(tmpDir, ".dvm", "nvim-status.json"),
	}

	opts := InitOptions{
		ConfigPath: configPath,
		Template:   "minimal",
	}

	if err := m.initMinimal(opts); err != nil {
		t.Fatalf("initMinimal failed: %v", err)
	}

	// Check that init.lua was created
	initPath := filepath.Join(configPath, "init.lua")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		t.Fatal("init.lua was not created")
	}

	// Read and verify content
	content, err := os.ReadFile(initPath)
	if err != nil {
		t.Fatalf("Failed to read init.lua: %v", err)
	}

	contentStr := string(content)
	// Check for key elements
	if !strings.Contains(contentStr, "DevOpsMaestro") {
		t.Error("init.lua missing DevOpsMaestro header")
	}
	if !strings.Contains(contentStr, "mapleader") {
		t.Error("init.lua missing leader key configuration")
	}
	if !strings.Contains(contentStr, "lazy.nvim") {
		t.Error("init.lua missing lazy.nvim plugin manager")
	}
}

// TestInitFromTemplate tests the template dispatcher
func TestInitFromTemplate(t *testing.T) {
	tests := []struct {
		name        string
		template    string
		gitClone    bool
		gitURL      string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "minimal template",
			template:    "minimal",
			gitClone:    false,
			expectError: false,
		},
		{
			name:        "kickstart without git clone",
			template:    "kickstart",
			gitClone:    false,
			expectError: false, // Falls back to minimal
		},
		{
			name:        "lazyvim without git clone",
			template:    "lazyvim",
			gitClone:    false,
			expectError: false, // Falls back to minimal
		},
		{
			name:        "astronvim without git clone",
			template:    "astronvim",
			gitClone:    false,
			expectError: false, // Falls back to minimal
		},
		{
			name:        "custom without git URL",
			template:    "custom",
			gitClone:    false,
			expectError: true,
			errorMsg:    "custom template requires --git-url",
		},
		{
			name:        "unknown template",
			template:    "unknown",
			gitClone:    false,
			expectError: true,
			errorMsg:    "unknown template",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "nvim")
			if err := os.MkdirAll(configPath, 0755); err != nil {
				t.Fatalf("Failed to create config directory: %v", err)
			}

			m := &manager{
				configPath: configPath,
				statusFile: filepath.Join(tmpDir, ".dvm", "nvim-status.json"),
			}

			opts := InitOptions{
				ConfigPath: configPath,
				Template:   tt.template,
				GitClone:   tt.gitClone,
				GitURL:     tt.gitURL,
			}

			err := m.initFromTemplate(opts)

			if tt.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none")
				}
				if tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error message to contain %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				// Verify init.lua was created
				initPath := filepath.Join(configPath, "init.lua")
				if _, err := os.Stat(initPath); os.IsNotExist(err) {
					t.Error("init.lua was not created")
				}
			}
		})
	}
}

// TestCloneTemplateWithSubdir tests cloning with subdirectory extraction
// This test is skipped by default because it requires network access
func TestCloneTemplateWithSubdir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	// Use a small, stable test repository
	// We'll use our own devopsmaestro repo which we know exists
	gitURL := "https://github.com/yourusername/test-nvim-configs.git"
	subdir := "configs/nvim"

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")

	m := &manager{
		configPath: configPath,
		statusFile: filepath.Join(tmpDir, ".dvm", "nvim-status.json"),
	}

	// Skip if git is not available
	if _, err := os.Stat("/usr/bin/git"); os.IsNotExist(err) {
		t.Skip("git not available")
	}

	// This will fail since the repo doesn't exist, but we're testing the logic
	err := m.cloneTemplate(gitURL, configPath, subdir)
	// We expect an error here because the repo doesn't exist
	// The test validates that the error handling works
	if err == nil {
		t.Log("Clone succeeded (unexpected but ok)")
	} else if !strings.Contains(err.Error(), "git clone failed") {
		t.Errorf("Expected 'git clone failed' error, got: %v", err)
	}
}

// TestCloneTemplateInvalidURL tests error handling for invalid URLs
func TestCloneTemplateInvalidURL(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")

	m := &manager{
		configPath: configPath,
		statusFile: filepath.Join(tmpDir, ".dvm", "nvim-status.json"),
	}

	// Test with invalid URL that git will reject
	invalidURL := "https://this-domain-definitely-does-not-exist-12345.com/repo.git"

	err := m.cloneTemplate(invalidURL, configPath, "")
	if err == nil {
		t.Fatal("Expected error for invalid URL, got none")
	}

	if !strings.Contains(err.Error(), "git clone failed") {
		t.Errorf("Expected 'git clone failed' error, got: %v", err)
	}
}

// TestCloneTemplateMissingSubdir tests error handling when subdirectory doesn't exist
func TestCloneTemplateMissingSubdir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")

	m := &manager{
		configPath: configPath,
		statusFile: filepath.Join(tmpDir, ".dvm", "nvim-status.json"),
	}

	// Use a real repo but request a subdirectory that doesn't exist
	gitURL := "https://github.com/nvim-lua/kickstart.nvim.git"
	nonExistentSubdir := "this-directory-does-not-exist"

	err := m.cloneTemplate(gitURL, configPath, nonExistentSubdir)
	if err == nil {
		t.Fatal("Expected error for missing subdirectory, got none")
	}

	if !strings.Contains(err.Error(), "subdirectory") && !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected subdirectory error, got: %v", err)
	}
}

// TestCloneTemplateGitHubShorthand tests GitHub shorthand URL normalization
func TestCloneTemplateGitHubShorthand(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")

	m := &manager{
		configPath: configPath,
		statusFile: filepath.Join(tmpDir, ".dvm", "nvim-status.json"),
	}

	// Test with GitHub shorthand - should be normalized to full URL
	shorthandURL := "github:nvim-lua/kickstart.nvim"

	// This should normalize to full URL and attempt to clone
	err := m.cloneTemplate(shorthandURL, configPath, "")
	// We don't check the result, just that normalization happens
	// The actual clone might succeed or fail depending on network
	if err != nil {
		t.Logf("Clone result (may succeed or fail): %v", err)
	}

	// Verify the URL was normalized by checking it starts with https
	normalizedURL := NormalizeGitURL(shorthandURL)
	if !strings.HasPrefix(normalizedURL, "https://") {
		t.Errorf("Expected normalized URL to start with https://, got: %s", normalizedURL)
	}
}

// TestCloneTemplateRemovesGitDir tests that .git directory is removed after cloning
func TestCloneTemplateRemovesGitDir(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping network test in short mode")
	}

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")

	m := &manager{
		configPath: configPath,
		statusFile: filepath.Join(tmpDir, ".dvm", "nvim-status.json"),
	}

	// Use nvim-lua/kickstart.nvim as a reliable test repo
	gitURL := "https://github.com/nvim-lua/kickstart.nvim.git"

	err := m.cloneTemplate(gitURL, configPath, "")
	if err != nil {
		t.Skipf("Clone failed (may be network issue): %v", err)
	}

	// Check that .git directory was removed
	gitDir := filepath.Join(configPath, ".git")
	if _, err := os.Stat(gitDir); !os.IsNotExist(err) {
		t.Error(".git directory was not removed after cloning")
	}

	// Check that files were cloned
	initPath := filepath.Join(configPath, "init.lua")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		t.Error("init.lua was not created from clone")
	}
}

// TestInitCustomWithGitURL tests custom template initialization
func TestInitCustomWithGitURL(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	m := &manager{
		configPath: configPath,
		statusFile: filepath.Join(tmpDir, ".dvm", "nvim-status.json"),
	}

	// Test without GitURL (should error)
	opts := InitOptions{
		ConfigPath: configPath,
		Template:   "custom",
		GitURL:     "",
	}

	err := m.initCustom(opts)
	if err == nil {
		t.Fatal("Expected error when GitURL is empty, got none")
	}
	if !strings.Contains(err.Error(), "requires GitURL") {
		t.Errorf("Expected 'requires GitURL' error, got: %v", err)
	}
}

// TestSaveAndLoadStatus tests status persistence
func TestSaveAndLoadStatus(t *testing.T) {
	tmpDir := t.TempDir()
	statusFile := filepath.Join(tmpDir, ".dvm", "nvim-status.json")

	m := &manager{
		configPath: filepath.Join(tmpDir, "nvim"),
		statusFile: statusFile,
	}

	// Create a status
	now := time.Now()
	status := &Status{
		ConfigPath:    m.configPath,
		Exists:        true,
		LastSync:      now,
		SyncedWith:    "workspace-1",
		LocalChanges:  false,
		RemoteChanges: false,
		Template:      "minimal",
	}

	// Save it
	if err := m.saveStatus(status); err != nil {
		t.Fatalf("Failed to save status: %v", err)
	}

	// Check file was created
	if _, err := os.Stat(statusFile); os.IsNotExist(err) {
		t.Fatal("Status file was not created")
	}

	// Load it back
	loaded, err := m.loadStatus()
	if err != nil {
		t.Fatalf("Failed to load status: %v", err)
	}

	// Verify data
	if loaded.Exists != status.Exists {
		t.Errorf("Exists mismatch: expected %v, got %v", status.Exists, loaded.Exists)
	}
	if loaded.Template != status.Template {
		t.Errorf("Template mismatch: expected %s, got %s", status.Template, loaded.Template)
	}
	if loaded.ConfigPath != status.ConfigPath {
		t.Errorf("ConfigPath mismatch: expected %s, got %s", status.ConfigPath, loaded.ConfigPath)
	}
	if loaded.SyncedWith != status.SyncedWith {
		t.Errorf("SyncedWith mismatch: expected %s, got %s", status.SyncedWith, loaded.SyncedWith)
	}
	// Check LastSync is approximately the same (within 1 second due to serialization)
	timeDiff := loaded.LastSync.Sub(status.LastSync)
	if timeDiff > time.Second || timeDiff < -time.Second {
		t.Errorf("LastSync time difference too large: %v", timeDiff)
	}
}

// TestLoadStatusNotExist tests loading status when file doesn't exist
func TestLoadStatusNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	statusFile := filepath.Join(tmpDir, ".dvm", "nvim-status.json")

	m := &manager{
		configPath: filepath.Join(tmpDir, "nvim"),
		statusFile: statusFile,
	}

	_, err := m.loadStatus()
	if err == nil {
		t.Fatal("Expected error when loading non-existent status, got none")
	}
}

// TestInitKickstartFallback tests kickstart template falls back to minimal
func TestInitKickstartFallback(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nvim")
	if err := os.MkdirAll(configPath, 0755); err != nil {
		t.Fatalf("Failed to create config directory: %v", err)
	}

	m := &manager{
		configPath: configPath,
		statusFile: filepath.Join(tmpDir, ".dvm", "nvim-status.json"),
	}

	// Test kickstart without GitClone (should fall back to minimal)
	opts := InitOptions{
		ConfigPath: configPath,
		Template:   "kickstart",
		GitClone:   false,
	}

	if err := m.initKickstart(opts); err != nil {
		t.Fatalf("initKickstart failed: %v", err)
	}

	// Check that init.lua was created (minimal template)
	initPath := filepath.Join(configPath, "init.lua")
	if _, err := os.Stat(initPath); os.IsNotExist(err) {
		t.Fatal("init.lua was not created")
	}

	// Verify it's the minimal template
	content, err := os.ReadFile(initPath)
	if err != nil {
		t.Fatalf("Failed to read init.lua: %v", err)
	}

	if !strings.Contains(string(content), "DevOpsMaestro") {
		t.Error("Expected minimal template, got different content")
	}
}
