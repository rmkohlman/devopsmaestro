package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

// =============================================================================
// Task 2.3: Workspace Directory Creation Tests (v0.19.0)
// Tests for creating and managing workspace directory structure
// =============================================================================

// TestCreateWorkspaceDirectories verifies all required subdirectories are created
func TestCreateWorkspaceDirectories(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	workspaceSlug := "test-eco-test-domain-test-app-dev"
	workspacePath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", workspaceSlug)

	tests := []struct {
		name             string
		expectedDirs     []string
		checkPermissions bool
		expectedFileMode os.FileMode
	}{
		{
			name: "create all workspace directories",
			expectedDirs: []string{
				"repo",
				"volume",
				"volume/nvim-data",
				"volume/nvim-state",
				"volume/cache",
				".dvm",
				".dvm/nvim",
				".dvm/shell",
				".dvm/starship",
			},
			checkPermissions: true,
			expectedFileMode: 0700, // User-only access
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// FIXME: This test will FAIL - CreateWorkspaceDirectories() doesn't exist yet
			// After Phase 3, should have:
			// func CreateWorkspaceDirectories(workspacePath string) error
			err := CreateWorkspaceDirectories(workspacePath)
			if err != nil {
				t.Fatalf("CreateWorkspaceDirectories() error = %v", err)
			}

			// Verify all expected directories exist
			for _, dir := range tt.expectedDirs {
				fullPath := filepath.Join(workspacePath, dir)
				info, err := os.Stat(fullPath)
				if err != nil {
					t.Errorf("Directory %q was not created: %v", dir, err)
					continue
				}

				if !info.IsDir() {
					t.Errorf("Path %q exists but is not a directory", dir)
				}

				// Check permissions if requested
				if tt.checkPermissions {
					actualMode := info.Mode().Perm()
					if actualMode != tt.expectedFileMode {
						t.Errorf("Directory %q has mode %o, want %o", dir, actualMode, tt.expectedFileMode)
					}
				}
			}
		})
	}
}

// TestWorkspaceDirectoryPermissions verifies directories have correct permissions (0700)
func TestWorkspaceDirectoryPermissions(t *testing.T) {
	tempDir := t.TempDir()
	workspaceSlug := "permissions-test"
	workspacePath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", workspaceSlug)

	// FIXME: This test will FAIL - CreateWorkspaceDirectories() doesn't exist yet
	err := CreateWorkspaceDirectories(workspacePath)
	if err != nil {
		t.Fatalf("CreateWorkspaceDirectories() error = %v", err)
	}

	// Test that sensitive directories have restricted permissions
	sensitiveDirs := []string{
		workspacePath, // Base directory
		filepath.Join(workspacePath, "repo"),
		filepath.Join(workspacePath, "volume"),
		filepath.Join(workspacePath, ".dvm"),
	}

	for _, dir := range sensitiveDirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Failed to stat %q: %v", dir, err)
			continue
		}

		mode := info.Mode().Perm()
		// Should be 0700 (rwx------)
		if mode != 0700 {
			t.Errorf("Directory %q has mode %o, want 0700 (user-only access)", dir, mode)
		}

		// Verify group and others have no permissions
		if mode&0077 != 0 {
			t.Errorf("Directory %q has group/other permissions %o, should be 0", dir, mode&0077)
		}
	}
}

// TestWorkspaceDirectoryIdempotent verifies calling twice doesn't error
func TestWorkspaceDirectoryIdempotent(t *testing.T) {
	tempDir := t.TempDir()
	workspaceSlug := "idempotent-test"
	workspacePath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", workspaceSlug)

	// FIXME: This test will FAIL - CreateWorkspaceDirectories() doesn't exist yet
	// First call should succeed
	err := CreateWorkspaceDirectories(workspacePath)
	if err != nil {
		t.Fatalf("First CreateWorkspaceDirectories() error = %v", err)
	}

	// Second call should also succeed (idempotent)
	err = CreateWorkspaceDirectories(workspacePath)
	if err != nil {
		t.Errorf("Second CreateWorkspaceDirectories() should be idempotent, got error = %v", err)
	}

	// Verify directories still exist and are correct
	expectedDirs := []string{
		"repo",
		"volume/nvim-data",
		".dvm/nvim",
	}

	for _, dir := range expectedDirs {
		fullPath := filepath.Join(workspacePath, dir)
		if _, err := os.Stat(fullPath); err != nil {
			t.Errorf("Directory %q missing after second call: %v", dir, err)
		}
	}
}

// TestDeleteWorkspaceDirectories verifies all workspace directories are removed
func TestDeleteWorkspaceDirectories(t *testing.T) {
	tempDir := t.TempDir()
	workspaceSlug := "delete-test"
	workspacePath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", workspaceSlug)

	// Create directories first
	// FIXME: Both functions will FAIL - don't exist yet
	err := CreateWorkspaceDirectories(workspacePath)
	if err != nil {
		t.Fatalf("Setup CreateWorkspaceDirectories() error = %v", err)
	}

	// Create some test files
	testFile := filepath.Join(workspacePath, "repo", "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0600); err != nil {
		t.Fatalf("Setup failed to create test file: %v", err)
	}

	// Verify directory exists before deletion
	if _, err := os.Stat(workspacePath); err != nil {
		t.Fatalf("Workspace directory doesn't exist before deletion: %v", err)
	}

	// FIXME: This test will FAIL - DeleteWorkspaceDirectories() doesn't exist yet
	// After Phase 3, should have:
	// func DeleteWorkspaceDirectories(workspacePath string) error
	err = DeleteWorkspaceDirectories(workspacePath)
	if err != nil {
		t.Fatalf("DeleteWorkspaceDirectories() error = %v", err)
	}

	// Verify directory no longer exists
	if _, err := os.Stat(workspacePath); !os.IsNotExist(err) {
		t.Errorf("DeleteWorkspaceDirectories() directory still exists after deletion")
	}

	// Parent directories should still exist
	parentDir := filepath.Join(tempDir, ".devopsmaestro", "workspaces")
	if _, err := os.Stat(parentDir); err != nil {
		t.Errorf("Parent directory was deleted, should be preserved: %v", err)
	}
}

// TestWorkspaceDirectoryPathsCorrect verifies paths match expected structure
func TestWorkspaceDirectoryPathsCorrect(t *testing.T) {
	tempDir := t.TempDir()
	workspaceSlug := "path-test-eco-domain-app-ws"
	workspacePath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", workspaceSlug)

	// FIXME: This test will FAIL - CreateWorkspaceDirectories() doesn't exist yet
	err := CreateWorkspaceDirectories(workspacePath)
	if err != nil {
		t.Fatalf("CreateWorkspaceDirectories() error = %v", err)
	}

	tests := []struct {
		name         string
		relativePath string
		description  string
	}{
		{
			name:         "repo directory for git clone",
			relativePath: "repo",
			description:  "Should contain git repository clone",
		},
		{
			name:         "nvim data directory",
			relativePath: "volume/nvim-data",
			description:  "Maps to XDG_DATA_HOME/nvim in container",
		},
		{
			name:         "nvim state directory",
			relativePath: "volume/nvim-state",
			description:  "Maps to XDG_STATE_HOME/nvim in container",
		},
		{
			name:         "cache directory",
			relativePath: "volume/cache",
			description:  "Maps to XDG_CACHE_HOME in container",
		},
		{
			name:         "generated nvim config",
			relativePath: ".dvm/nvim",
			description:  "Contains init.lua and plugin configs",
		},
		{
			name:         "generated shell config",
			relativePath: ".dvm/shell",
			description:  "Contains .zshrc.workspace",
		},
		{
			name:         "generated starship config",
			relativePath: ".dvm/starship",
			description:  "Contains starship.toml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := filepath.Join(workspacePath, tt.relativePath)

			info, err := os.Stat(fullPath)
			if err != nil {
				t.Errorf("Path %q does not exist: %v", tt.relativePath, err)
				return
			}

			if !info.IsDir() {
				t.Errorf("Path %q is not a directory", tt.relativePath)
			}

			// Verify path components
			if !filepath.IsAbs(filepath.Join(tempDir, ".devopsmaestro")) {
				// Make it absolute for this test
			}

			// Check that workspace slug is in the path
			if !contains(fullPath, workspaceSlug) {
				t.Errorf("Path %q does not contain workspace slug %q", fullPath, workspaceSlug)
			}
		})
	}
}

// TestWorkspaceDirectoryStructureMatches verifies structure matches architecture spec
func TestWorkspaceDirectoryStructureMatches(t *testing.T) {
	tempDir := t.TempDir()
	workspaceSlug := "structure-test"
	workspacePath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", workspaceSlug)

	// FIXME: This test will FAIL - CreateWorkspaceDirectories() doesn't exist yet
	err := CreateWorkspaceDirectories(workspacePath)
	if err != nil {
		t.Fatalf("CreateWorkspaceDirectories() error = %v", err)
	}

	// Expected structure from architecture spec:
	// ~/.devopsmaestro/workspaces/{workspace-slug}/
	// ├── repo/             # Git clone
	// ├── volume/           # Persistent data
	// │   ├── nvim-data/    # XDG_DATA_HOME/nvim
	// │   ├── nvim-state/   # XDG_STATE_HOME/nvim
	// │   └── cache/        # XDG_CACHE_HOME
	// └── .dvm/             # Generated configs
	//     ├── nvim/
	//     ├── shell/
	//     └── starship/

	expectedStructure := map[string]bool{
		"repo":              true,
		"volume":            true,
		"volume/nvim-data":  true,
		"volume/nvim-state": true,
		"volume/cache":      true,
		".dvm":              true,
		".dvm/nvim":         true,
		".dvm/shell":        true,
		".dvm/starship":     true,
	}

	// Walk the workspace directory
	err = filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip the root
		if path == workspacePath {
			return nil
		}

		// Get relative path
		relPath, err := filepath.Rel(workspacePath, path)
		if err != nil {
			return err
		}

		// Only check directories we care about
		if info.IsDir() {
			if expected, ok := expectedStructure[relPath]; ok && expected {
				delete(expectedStructure, relPath) // Mark as found
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Error walking workspace directory: %v", err)
	}

	// Check if any expected directories were not found
	if len(expectedStructure) > 0 {
		t.Errorf("Missing expected directories:")
		for dir := range expectedStructure {
			t.Errorf("  - %s", dir)
		}
	}
}

// TestWorkspaceDirectoryNoHostPathLeakage verifies no paths outside ~/.devopsmaestro/
func TestWorkspaceDirectoryNoHostPathLeakage(t *testing.T) {
	tempDir := t.TempDir()
	workspaceSlug := "leakage-test"
	workspacePath := filepath.Join(tempDir, ".devopsmaestro", "workspaces", workspaceSlug)

	// FIXME: This test will FAIL - CreateWorkspaceDirectories() doesn't exist yet
	err := CreateWorkspaceDirectories(workspacePath)
	if err != nil {
		t.Fatalf("CreateWorkspaceDirectories() error = %v", err)
	}

	// Verify base path is under .devopsmaestro
	if !contains(workspacePath, ".devopsmaestro") {
		t.Errorf("Workspace path %q is not under .devopsmaestro/", workspacePath)
	}

	// Verify no symlinks to host paths
	prohibitedPaths := []string{
		"/home", "~/.config", "~/.local", "~/.ssh",
		"/etc", "/usr", "/var",
	}

	err = filepath.Walk(workspacePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if it's a symlink
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := os.Readlink(path)
			if err != nil {
				return err
			}

			// Verify symlink doesn't point to prohibited paths
			for _, prohibited := range prohibitedPaths {
				if contains(target, prohibited) {
					t.Errorf("Found symlink to prohibited path: %s -> %s", path, target)
				}
			}
		}

		return nil
	})

	if err != nil {
		t.Fatalf("Error walking workspace directory: %v", err)
	}
}

// contains is a helper to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			filepath.Dir(s) == substr || filepath.Base(s) == substr ||
			filepath.Join("", s) == filepath.Join("", substr) ||
			len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
