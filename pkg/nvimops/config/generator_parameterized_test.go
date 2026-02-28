package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// v0.19.0 Parameterized Config Generator Tests - NvimOps
// Tests verify generators accept output path parameter and never write to host.
// Implementation complete - all generators support custom output paths.
// =============================================================================

// TestLuaGeneratorWithCustomOutputPath verifies generator accepts custom output path
func TestLuaGeneratorWithCustomOutputPath(t *testing.T) {
	tempDir := t.TempDir()
	workspaceSlug := "test-eco-domain-app-ws"
	customOutputPath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", workspaceSlug, ".dvm", "nvim")

	// Create minimal config
	cfg := &CoreConfig{
		Namespace: "workspace",
	}

	gen := NewGenerator()

	// WriteToDirectory now accepts a custom output path parameter
	err := gen.WriteToDirectory(cfg, nil, customOutputPath)
	if err != nil {
		t.Fatalf("WriteToDirectory() with custom path error = %v", err)
	}

	// Verify files were written to custom path
	expectedFiles := []string{
		"init.lua",
		filepath.Join("lua", "workspace", "lazy.lua"),
		filepath.Join("lua", "workspace", "core", "init.lua"),
	}

	for _, file := range expectedFiles {
		fullPath := filepath.Join(customOutputPath, file)
		if _, err := os.Stat(fullPath); err != nil {
			t.Errorf("Expected file %q not found at custom path: %v", file, err)
		}
	}
}

// TestLuaGeneratorWritesToWorkspacePath verifies writes go to workspace path
func TestLuaGeneratorWritesToWorkspacePath(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name          string
		workspaceSlug string
		expectedPath  string
	}{
		{
			name:          "dev workspace",
			workspaceSlug: "personal-tools-dvm-dev",
			expectedPath:  ".devopsmaestro/workspaces/personal-tools-dvm-dev/.dvm/nvim",
		},
		{
			name:          "production workspace",
			workspaceSlug: "enterprise-payments-api-prod",
			expectedPath:  ".devopsmaestro/workspaces/enterprise-payments-api-prod/.dvm/nvim",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputPath := filepath.Join(tempDir, tt.expectedPath)

			cfg := &CoreConfig{
				Namespace: "workspace",
			}

			gen := NewGenerator()

			// Verify that output path parameter is used correctly
			err := gen.WriteToDirectory(cfg, nil, outputPath)
			if err != nil {
				t.Fatalf("WriteToDirectory() error = %v", err)
			}

			// Verify output is in the workspace path
			if !strings.Contains(outputPath, tt.workspaceSlug) {
				t.Errorf("Output path %q does not contain workspace slug %q", outputPath, tt.workspaceSlug)
			}

			// Verify .dvm/nvim subdirectory exists
			if !strings.HasSuffix(outputPath, ".dvm/nvim") && !strings.HasSuffix(outputPath, ".dvm"+string(filepath.Separator)+"nvim") {
				t.Errorf("Output path %q does not end with .dvm/nvim", outputPath)
			}

			// Verify init.lua was created in workspace path
			initLua := filepath.Join(outputPath, "init.lua")
			if _, err := os.Stat(initLua); err != nil {
				t.Errorf("init.lua not found in workspace path: %v", err)
			}
		})
	}
}

// TestLuaGeneratorDoesNotWriteToHostPath verifies NEVER writes to ~/.config/nvim/
func TestLuaGeneratorDoesNotWriteToHostPath(t *testing.T) {
	// Get user's home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skipf("Cannot get home directory: %v", err)
	}

	hostConfigPath := filepath.Join(homeDir, ".config", "nvim")

	// Record initial state of host config directory
	hostConfigExists := false
	var hostConfigModTime int64
	if info, err := os.Stat(hostConfigPath); err == nil {
		hostConfigExists = true
		hostConfigModTime = info.ModTime().Unix()
	}

	// Create temporary workspace path
	tempDir := t.TempDir()
	workspacePath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", "test-ws", ".dvm", "nvim")

	cfg := &CoreConfig{
		Namespace: "workspace",
	}

	gen := NewGenerator()

	// CRITICAL: Generator must NEVER write to ~/.config/nvim when given a workspace path
	err = gen.WriteToDirectory(cfg, nil, workspacePath)
	if err != nil {
		t.Fatalf("WriteToDirectory() error = %v", err)
	}

	// Verify host config directory was NOT touched
	if hostConfigExists {
		// If it existed before, verify it wasn't modified
		info, err := os.Stat(hostConfigPath)
		if err == nil {
			if info.ModTime().Unix() > hostConfigModTime {
				t.Errorf("Host config directory %q was modified - generator leaked to host!", hostConfigPath)
			}
		}
	} else {
		// If it didn't exist, verify it still doesn't
		if _, err := os.Stat(hostConfigPath); err == nil {
			t.Errorf("Host config directory %q was created - generator leaked to host!", hostConfigPath)
		}
	}

	// Verify output went to workspace path instead
	workspaceInitLua := filepath.Join(workspacePath, "init.lua")
	if _, err := os.Stat(workspaceInitLua); err != nil {
		t.Errorf("init.lua not found in workspace path: %v", err)
	}

	// Verify NO files in host .config/nvim have workspace namespace
	if hostConfigExists {
		filepath.Walk(hostConfigPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // Skip errors
			}
			if !info.IsDir() && filepath.Ext(path) == ".lua" {
				content, err := os.ReadFile(path)
				if err == nil {
					// Check if file contains our workspace namespace
					if strings.Contains(string(content), "-- Generated by DevOpsMaestro") {
						t.Errorf("Host config file %q contains DevOpsMaestro marker - generator leaked!", path)
					}
				}
			}
			return nil
		})
	}
}

// TestLuaGeneratorPathParameterRequired verifies nil/empty path is rejected
func TestLuaGeneratorPathParameterRequired(t *testing.T) {
	cfg := &CoreConfig{
		Namespace: "workspace",
	}

	gen := NewGenerator()

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "empty path should error",
			path:    "",
			wantErr: true,
		},
		{
			name:    "valid path succeeds",
			path:    t.TempDir(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := gen.WriteToDirectory(cfg, nil, tt.path)

			if tt.wantErr && err == nil {
				t.Errorf("WriteToDirectory() with path=%q expected error, got nil", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("WriteToDirectory() with path=%q unexpected error = %v", tt.path, err)
			}
		})
	}
}

// TestLuaGeneratorPreservesExistingHostConfig verifies existing host config unchanged
func TestLuaGeneratorPreservesExistingHostConfig(t *testing.T) {
	// Create mock host config directory
	mockHostDir := t.TempDir()
	hostConfigPath := filepath.Join(mockHostDir, ".config", "nvim")
	if err := os.MkdirAll(hostConfigPath, 0755); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create existing host init.lua
	hostInitPath := filepath.Join(hostConfigPath, "init.lua")
	originalContent := "-- Existing user config\nprint('hello')\n"
	if err := os.WriteFile(hostInitPath, []byte(originalContent), 0644); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create workspace output path
	workspacePath := filepath.Join(t.TempDir(), ".dvm", "nvim")

	cfg := &CoreConfig{
		Namespace: "workspace",
	}

	gen := NewGenerator()

	// Generate to workspace path
	err := gen.WriteToDirectory(cfg, nil, workspacePath)
	if err != nil {
		t.Fatalf("WriteToDirectory() error = %v", err)
	}

	// Verify host config is unchanged
	hostContent, err := os.ReadFile(hostInitPath)
	if err != nil {
		t.Fatalf("Failed to read host config: %v", err)
	}

	if string(hostContent) != originalContent {
		t.Errorf("Host config was modified:\nOriginal: %q\nCurrent:  %q", originalContent, string(hostContent))
	}

	// Verify workspace has different content
	workspaceInitPath := filepath.Join(workspacePath, "init.lua")
	workspaceContent, err := os.ReadFile(workspaceInitPath)
	if err != nil {
		t.Fatalf("Failed to read workspace config: %v", err)
	}

	if string(workspaceContent) == originalContent {
		t.Errorf("Workspace config should be different from host config")
	}
}

// TestPluginGeneratorWithWorkspacePath verifies plugin configs go to workspace path
func TestPluginGeneratorWithWorkspacePath(t *testing.T) {
	tempDir := t.TempDir()
	workspaceSlug := "test-ws"
	pluginOutputPath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", workspaceSlug, ".dvm", "nvim", "lua", "workspace", "plugins")

	cfg := &CoreConfig{
		Namespace: "workspace",
	}

	gen := NewGenerator()

	// Verify plugin configs are written to the workspace path
	// Plugin configurations should be written to: {workspace}/.dvm/nvim/lua/workspace/plugins/
	workspacePath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", workspaceSlug, ".dvm", "nvim")
	err := gen.WriteToDirectory(cfg, nil, workspacePath)
	if err != nil {
		t.Fatalf("WriteToDirectory() error = %v", err)
	}

	// Verify plugin directory exists in workspace path
	if _, err := os.Stat(pluginOutputPath); err != nil {
		t.Errorf("Plugin output directory not found in workspace path: %v", err)
	}

	// Verify it contains workspace slug in path
	if !strings.Contains(pluginOutputPath, workspaceSlug) {
		t.Errorf("Plugin path %q does not contain workspace slug %q", pluginOutputPath, workspaceSlug)
	}
}
