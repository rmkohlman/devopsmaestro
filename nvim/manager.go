package nvim

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Manager handles Neovim configuration management operations
type Manager interface {
	// Initialize local Neovim configuration
	Init(opts InitOptions) error

	// Sync local config with remote workspace
	Sync(workspace string, direction SyncDirection) error

	// Push local config to remote workspace
	Push(workspace string) error

	// Get status of local config
	Status() (*Status, error)

	// List available workspaces
	ListWorkspaces() ([]Workspace, error)
}

// manager is the default implementation of Manager
type manager struct {
	configPath string
	statusFile string
}

// NewManager creates a new Neovim configuration manager
func NewManager() Manager {
	homeDir, _ := os.UserHomeDir()
	configPath := filepath.Join(homeDir, ".config", "nvim")
	statusFile := filepath.Join(homeDir, ".devopsmaestro", ".nvim-sync-status")

	return &manager{
		configPath: configPath,
		statusFile: statusFile,
	}
}

// NewManagerWithPath creates a manager with a custom config path
func NewManagerWithPath(configPath string) Manager {
	homeDir, _ := os.UserHomeDir()
	statusFile := filepath.Join(homeDir, ".devopsmaestro", ".nvim-sync-status")

	return &manager{
		configPath: configPath,
		statusFile: statusFile,
	}
}

// InitOptions contains options for initializing Neovim config
type InitOptions struct {
	ConfigPath string // Default: ~/.config/nvim
	Template   string // kickstart, lazyvim, astronvim, custom, minimal
	Overwrite  bool   // Overwrite existing config
	GitClone   bool   // Clone from git (vs copy template)
	GitURL     string // Custom git URL to clone from
	Subdir     string // Subdirectory within repo to use (e.g., "templates/starter")
}

// SyncDirection specifies the direction of the sync operation
type SyncDirection int

const (
	// SyncPull syncs from workspace to local (workspace → local)
	SyncPull SyncDirection = iota
	// SyncPush syncs from local to workspace (local → workspace)
	SyncPush
	// SyncBidirectional merges both ways
	SyncBidirectional
)

// String returns the string representation of SyncDirection
func (d SyncDirection) String() string {
	switch d {
	case SyncPull:
		return "pull"
	case SyncPush:
		return "push"
	case SyncBidirectional:
		return "bidirectional"
	default:
		return "unknown"
	}
}

// Status represents the current status of local Neovim config
type Status struct {
	ConfigPath    string    // Path to local config
	Exists        bool      // Does config exist
	LastSync      time.Time // Last sync timestamp
	SyncedWith    string    // Workspace ID last synced with
	LocalChanges  bool      // Has local changes since last sync
	RemoteChanges bool      // Has remote changes since last sync (if applicable)
	Template      string    // Template used for initialization
}

// Workspace represents a DevOpsMaestro workspace
type Workspace struct {
	ID       string
	Name     string
	Active   bool
	NvimPath string // Path to Neovim config in workspace container
}

// Init initializes the local Neovim configuration
func (m *manager) Init(opts InitOptions) error {
	// Use manager's config path if not specified
	if opts.ConfigPath == "" {
		opts.ConfigPath = m.configPath
	}

	// Check if config already exists
	if _, err := os.Stat(opts.ConfigPath); err == nil && !opts.Overwrite {
		return fmt.Errorf("config already exists at %s (use --overwrite to replace)", opts.ConfigPath)
	}

	// Remove existing config if overwriting
	if opts.Overwrite {
		if err := os.RemoveAll(opts.ConfigPath); err != nil {
			return fmt.Errorf("failed to remove existing config: %w", err)
		}
	}

	// Create config directory
	if err := os.MkdirAll(opts.ConfigPath, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Initialize based on template
	if err := m.initFromTemplate(opts); err != nil {
		return fmt.Errorf("failed to initialize from template: %w", err)
	}

	// Save status
	status := &Status{
		ConfigPath: opts.ConfigPath,
		Exists:     true,
		LastSync:   time.Now(),
		Template:   opts.Template,
	}
	if err := m.saveStatus(status); err != nil {
		return fmt.Errorf("failed to save status: %w", err)
	}

	return nil
}

// Sync synchronizes local config with remote workspace
func (m *manager) Sync(workspace string, direction SyncDirection) error {
	// Implementation will be added in next iteration
	return fmt.Errorf("sync not yet implemented")
}

// Push pushes local config to remote workspace
func (m *manager) Push(workspace string) error {
	return m.Sync(workspace, SyncPush)
}

// Status returns the current status of local Neovim config
func (m *manager) Status() (*Status, error) {
	status := &Status{
		ConfigPath: m.configPath,
	}

	// Check if config exists
	if info, err := os.Stat(m.configPath); err == nil && info.IsDir() {
		status.Exists = true
	}

	// Try to load saved status
	savedStatus, err := m.loadStatus()
	if err == nil {
		status.LastSync = savedStatus.LastSync
		status.SyncedWith = savedStatus.SyncedWith
		status.Template = savedStatus.Template
	}

	// Check for local changes (simplified - just check if any files exist)
	if status.Exists {
		entries, err := os.ReadDir(m.configPath)
		if err == nil && len(entries) > 0 {
			// If we have files and they were modified after last sync, mark as changed
			for _, entry := range entries {
				info, _ := entry.Info()
				if info != nil && info.ModTime().After(status.LastSync) {
					status.LocalChanges = true
					break
				}
			}
		}
	}

	return status, nil
}

// ListWorkspaces returns available workspaces (stub for now)
func (m *manager) ListWorkspaces() ([]Workspace, error) {
	// This will integrate with DevOpsMaestro's workspace DB in the future
	return []Workspace{}, nil
}
