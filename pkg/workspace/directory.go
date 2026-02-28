package workspace

import (
	"fmt"
	"os"
	"path/filepath"
)

// CreateWorkspaceDirectories creates the full directory structure for a workspace
// according to the v0.19.0 isolation architecture:
//
// {workspacePath}/
// ├── repo/             # Git clone
// ├── volume/           # Persistent data
// │   ├── nvim-data/    # XDG_DATA_HOME/nvim (plugins, shada)
// │   ├── nvim-state/   # XDG_STATE_HOME/nvim (undo, swap)
// │   └── cache/        # XDG_CACHE_HOME
// └── .dvm/             # Generated configs
//
//	├── nvim/         # init.lua, plugins/*.lua
//	├── shell/        # .zshrc.workspace
//	└── starship/     # starship.toml
//
// All directories are created with 0700 permissions (user-only access)
// for security isolation.
func CreateWorkspaceDirectories(workspacePath string) error {
	// Define all required subdirectories
	dirs := []string{
		filepath.Join(workspacePath, "repo"),
		filepath.Join(workspacePath, "volume"),
		filepath.Join(workspacePath, "volume", "nvim-data"),
		filepath.Join(workspacePath, "volume", "nvim-state"),
		filepath.Join(workspacePath, "volume", "cache"),
		filepath.Join(workspacePath, ".dvm"),
		filepath.Join(workspacePath, ".dvm", "nvim"),
		filepath.Join(workspacePath, ".dvm", "shell"),
		filepath.Join(workspacePath, ".dvm", "starship"),
	}

	// Create all directories with user-only permissions (0700)
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// DeleteWorkspaceDirectories removes all workspace directories
// This removes the entire workspace directory tree, including:
// - Git repository clone
// - Persistent volume data (nvim plugins, cache, etc.)
// - Generated configurations
func DeleteWorkspaceDirectories(workspacePath string) error {
	if err := os.RemoveAll(workspacePath); err != nil {
		return fmt.Errorf("failed to delete workspace directory %s: %w", workspacePath, err)
	}
	return nil
}

// GetWorkspaceBasePath returns the base path for all workspaces
// Returns: ~/.devopsmaestro
func GetWorkspaceBasePath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return filepath.Join(home, ".devopsmaestro"), nil
}

// GetWorkspacePath returns the full path for a workspace given its slug
// Format: ~/.devopsmaestro/workspaces/{slug}
func GetWorkspacePath(slug string) (string, error) {
	basePath, err := GetWorkspaceBasePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(basePath, "workspaces", slug), nil
}

// GetWorkspaceRepoPath returns the path to the workspace's git repository
// Format: ~/.devopsmaestro/workspaces/{slug}/repo
func GetWorkspaceRepoPath(slug string) (string, error) {
	workspacePath, err := GetWorkspacePath(slug)
	if err != nil {
		return "", err
	}
	return filepath.Join(workspacePath, "repo"), nil
}

// GetWorkspaceVolumePath returns the path to the workspace's persistent volume
// Format: ~/.devopsmaestro/workspaces/{slug}/volume
func GetWorkspaceVolumePath(slug string) (string, error) {
	workspacePath, err := GetWorkspacePath(slug)
	if err != nil {
		return "", err
	}
	return filepath.Join(workspacePath, "volume"), nil
}

// GetWorkspaceConfigPath returns the path to the workspace's generated configs
// Format: ~/.devopsmaestro/workspaces/{slug}/.dvm
func GetWorkspaceConfigPath(slug string) (string, error) {
	workspacePath, err := GetWorkspacePath(slug)
	if err != nil {
		return "", err
	}
	return filepath.Join(workspacePath, ".dvm"), nil
}
