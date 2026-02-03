// Package terminalops provides tools for managing terminal configurations.
//
// This package is designed to be:
// - Standalone: Can be used independently via the dvt CLI
// - Importable: Can be imported as a library by dvm for container integration
// - Portable: Enables sharing terminal setups as portable YAML files
//
// # Architecture
//
// The package is organized into sub-packages:
//   - prompt: Prompt types (Starship/P10k), YAML parsing, config generation
//   - plugin: Shell plugin types (zsh plugins), YAML parsing, install scripts
//   - shell: Shell config types (aliases, env, functions), YAML parsing
//   - profile: Aggregates prompt + plugins + shell into complete configs
//
// # Basic Usage
//
//	import "devopsmaestro/pkg/terminalops"
//
//	// Create a manager with default file storage
//	mgr, err := terminalops.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Install a preset profile
//	err = mgr.InstallPreset("default")
//
//	// Generate config files
//	err = mgr.GenerateProfile("default", "~/.config")
package terminalops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"devopsmaestro/pkg/terminalops/plugin"
	"devopsmaestro/pkg/terminalops/profile"
	"devopsmaestro/pkg/terminalops/prompt"
	"devopsmaestro/pkg/terminalops/shell"
)

// Manager provides high-level operations for terminal configuration management.
// It coordinates prompts, plugins, shell configs, and profiles.
type Manager struct {
	promptStore  PromptStore
	pluginStore  PluginStore
	shellStore   ShellStore
	profileStore ProfileStore

	promptGen  PromptGenerator
	pluginGen  PluginGenerator
	shellGen   ShellGenerator
	profileGen ProfileGenerator

	configDir string
}

// Options configures the Manager.
type Options struct {
	// ConfigDir is the base directory for configuration storage.
	// Defaults to ~/.dvt
	ConfigDir string

	// Stores - if nil, file-based stores are created
	PromptStore  PromptStore
	PluginStore  PluginStore
	ShellStore   ShellStore
	ProfileStore ProfileStore

	// Generators - if nil, default generators are created
	PromptGen  PromptGenerator
	PluginGen  PluginGenerator
	ShellGen   ShellGenerator
	ProfileGen ProfileGenerator
}

// New creates a new Manager with default options.
func New() (*Manager, error) {
	return NewWithOptions(Options{})
}

// NewWithOptions creates a new Manager with the specified options.
func NewWithOptions(opts Options) (*Manager, error) {
	configDir := opts.ConfigDir
	if configDir == "" {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".dvt")
	}

	// Expand ~ if present
	if strings.HasPrefix(configDir, "~") {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, configDir[2:])
	}

	// Create directories
	dirs := []string{
		filepath.Join(configDir, "prompts"),
		filepath.Join(configDir, "plugins"),
		filepath.Join(configDir, "shells"),
		filepath.Join(configDir, "profiles"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	mgr := &Manager{configDir: configDir}

	// Initialize stores
	if opts.PromptStore != nil {
		mgr.promptStore = opts.PromptStore
	} else {
		mgr.promptStore = NewFilePromptStore(filepath.Join(configDir, "prompts"))
	}

	if opts.PluginStore != nil {
		mgr.pluginStore = opts.PluginStore
	} else {
		mgr.pluginStore = NewFilePluginStore(filepath.Join(configDir, "plugins"))
	}

	if opts.ShellStore != nil {
		mgr.shellStore = opts.ShellStore
	} else {
		mgr.shellStore = NewFileShellStore(filepath.Join(configDir, "shells"))
	}

	if opts.ProfileStore != nil {
		mgr.profileStore = opts.ProfileStore
	} else {
		mgr.profileStore = NewFileProfileStore(filepath.Join(configDir, "profiles"))
	}

	// Initialize generators
	if opts.PromptGen != nil {
		mgr.promptGen = opts.PromptGen
	} else {
		mgr.promptGen = prompt.NewStarshipGenerator()
	}

	if opts.PluginGen != nil {
		mgr.pluginGen = opts.PluginGen
	} else {
		mgr.pluginGen = plugin.NewZshGenerator("")
	}

	if opts.ShellGen != nil {
		mgr.shellGen = opts.ShellGen
	} else {
		mgr.shellGen = shell.NewGenerator()
	}

	if opts.ProfileGen != nil {
		mgr.profileGen = opts.ProfileGen
	} else {
		mgr.profileGen = profile.NewGenerator("")
	}

	return mgr, nil
}

// =============================================================================
// PROMPT OPERATIONS
// =============================================================================

// GetPrompt retrieves a prompt by name.
func (m *Manager) GetPrompt(name string) (*prompt.Prompt, error) {
	return m.promptStore.Get(name)
}

// ListPrompts returns all stored prompts.
func (m *Manager) ListPrompts() ([]*prompt.Prompt, error) {
	return m.promptStore.List()
}

// SavePrompt saves a prompt to the store.
func (m *Manager) SavePrompt(p *prompt.Prompt) error {
	return m.promptStore.Save(p)
}

// DeletePrompt removes a prompt by name.
func (m *Manager) DeletePrompt(name string) error {
	return m.promptStore.Delete(name)
}

// GeneratePromptConfig generates the config file content for a prompt.
func (m *Manager) GeneratePromptConfig(p *prompt.Prompt) (string, error) {
	return m.promptGen.Generate(p)
}

// =============================================================================
// PLUGIN OPERATIONS
// =============================================================================

// GetPlugin retrieves a plugin by name.
func (m *Manager) GetPlugin(name string) (*plugin.Plugin, error) {
	return m.pluginStore.Get(name)
}

// ListPlugins returns all stored plugins.
func (m *Manager) ListPlugins() ([]*plugin.Plugin, error) {
	return m.pluginStore.List()
}

// SavePlugin saves a plugin to the store.
func (m *Manager) SavePlugin(p *plugin.Plugin) error {
	return m.pluginStore.Save(p)
}

// DeletePlugin removes a plugin by name.
func (m *Manager) DeletePlugin(name string) error {
	return m.pluginStore.Delete(name)
}

// GeneratePluginConfig generates shell config for plugins.
func (m *Manager) GeneratePluginConfig(plugins []*plugin.Plugin) (string, error) {
	return m.pluginGen.Generate(plugins)
}

// =============================================================================
// SHELL OPERATIONS
// =============================================================================

// GetShell retrieves a shell config by name.
func (m *Manager) GetShell(name string) (*shell.Shell, error) {
	return m.shellStore.Get(name)
}

// ListShells returns all stored shell configs.
func (m *Manager) ListShells() ([]*shell.Shell, error) {
	return m.shellStore.List()
}

// SaveShell saves a shell config to the store.
func (m *Manager) SaveShell(s *shell.Shell) error {
	return m.shellStore.Save(s)
}

// DeleteShell removes a shell config by name.
func (m *Manager) DeleteShell(name string) error {
	return m.shellStore.Delete(name)
}

// GenerateShellConfig generates shell config content.
func (m *Manager) GenerateShellConfig(s *shell.Shell) (string, error) {
	return m.shellGen.Generate(s)
}

// =============================================================================
// PROFILE OPERATIONS
// =============================================================================

// GetProfile retrieves a profile by name.
func (m *Manager) GetProfile(name string) (*profile.Profile, error) {
	return m.profileStore.Get(name)
}

// ListProfiles returns all stored profiles.
func (m *Manager) ListProfiles() ([]*profile.Profile, error) {
	return m.profileStore.List()
}

// SaveProfile saves a profile to the store.
func (m *Manager) SaveProfile(p *profile.Profile) error {
	return m.profileStore.Save(p)
}

// DeleteProfile removes a profile by name.
func (m *Manager) DeleteProfile(name string) error {
	return m.profileStore.Delete(name)
}

// InstallPreset installs a preset profile by name (default, minimal, power-user).
func (m *Manager) InstallPreset(name string) error {
	var p *profile.Profile
	switch name {
	case "default":
		p = profile.DefaultProfile()
	case "minimal":
		p = profile.MinimalProfile()
	case "power-user":
		p = profile.PowerUserProfile()
	default:
		return fmt.Errorf("unknown preset: %s (available: default, minimal, power-user)", name)
	}
	return m.profileStore.Save(p)
}

// GenerateProfile generates all config files for a profile.
func (m *Manager) GenerateProfile(p *profile.Profile) (*profile.GeneratedConfig, error) {
	return m.profileGen.Generate(p)
}

// GenerateProfileToDir generates all config files for a profile and writes to disk.
func (m *Manager) GenerateProfileToDir(p *profile.Profile, outputDir string) error {
	// Expand ~ if present
	if strings.HasPrefix(outputDir, "~") {
		home, _ := os.UserHomeDir()
		outputDir = filepath.Join(home, outputDir[1:])
	}

	config, err := m.profileGen.Generate(p)
	if err != nil {
		return fmt.Errorf("failed to generate profile: %w", err)
	}

	// Write starship.toml
	if config.StarshipTOML != "" {
		starshipPath := filepath.Join(outputDir, "starship.toml")
		if err := os.MkdirAll(filepath.Dir(starshipPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(starshipPath, []byte(config.StarshipTOML), 0644); err != nil {
			return fmt.Errorf("failed to write starship.toml: %w", err)
		}
	}

	// Write zshrc content
	if config.ZshrcFull != "" {
		zshrcPath := filepath.Join(outputDir, ".zshrc.dvt")
		if err := os.WriteFile(zshrcPath, []byte(config.ZshrcFull), 0644); err != nil {
			return fmt.Errorf("failed to write .zshrc.dvt: %w", err)
		}
	}

	return nil
}

// =============================================================================
// ACCESSORS
// =============================================================================

// ConfigDir returns the configuration directory.
func (m *Manager) ConfigDir() string {
	return m.configDir
}

// PromptStore returns the underlying prompt store.
func (m *Manager) PromptStore() PromptStore {
	return m.promptStore
}

// PluginStore returns the underlying plugin store.
func (m *Manager) PluginStore() PluginStore {
	return m.pluginStore
}

// ShellStore returns the underlying shell store.
func (m *Manager) ShellStore() ShellStore {
	return m.shellStore
}

// ProfileStore returns the underlying profile store.
func (m *Manager) ProfileStore() ProfileStore {
	return m.profileStore
}

// Close releases resources held by the manager.
func (m *Manager) Close() error {
	var errs []error
	if err := m.promptStore.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := m.pluginStore.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := m.shellStore.Close(); err != nil {
		errs = append(errs, err)
	}
	if err := m.profileStore.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing stores: %v", errs)
	}
	return nil
}
