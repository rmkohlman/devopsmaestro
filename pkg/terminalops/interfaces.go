package terminalops

import (
	"devopsmaestro/pkg/terminalops/plugin"
	"devopsmaestro/pkg/terminalops/profile"
	"devopsmaestro/pkg/terminalops/prompt"
	"devopsmaestro/pkg/terminalops/shell"
)

// =============================================================================
// STORE INTERFACES
// =============================================================================

// PromptStore defines storage operations for prompts.
// Implementations can use files, databases, or memory.
type PromptStore interface {
	// Save stores a prompt (creates or updates).
	Save(p *prompt.Prompt) error

	// Get retrieves a prompt by name.
	Get(name string) (*prompt.Prompt, error)

	// List returns all prompts.
	List() ([]*prompt.Prompt, error)

	// Delete removes a prompt by name.
	Delete(name string) error

	// Exists checks if a prompt exists.
	Exists(name string) bool

	// Close releases resources.
	Close() error
}

// PluginStore defines storage operations for plugins.
type PluginStore interface {
	// Save stores a plugin (creates or updates).
	Save(p *plugin.Plugin) error

	// Get retrieves a plugin by name.
	Get(name string) (*plugin.Plugin, error)

	// List returns all plugins.
	List() ([]*plugin.Plugin, error)

	// Delete removes a plugin by name.
	Delete(name string) error

	// Close releases resources.
	Close() error
}

// ShellStore defines storage operations for shell configs.
type ShellStore interface {
	// Save stores a shell config (creates or updates).
	Save(s *shell.Shell) error

	// Get retrieves a shell config by name.
	Get(name string) (*shell.Shell, error)

	// List returns all shell configs.
	List() ([]*shell.Shell, error)

	// Delete removes a shell config by name.
	Delete(name string) error

	// Close releases resources.
	Close() error
}

// ProfileStore defines storage operations for profiles.
type ProfileStore interface {
	// Save stores a profile (creates or updates).
	Save(p *profile.Profile) error

	// Get retrieves a profile by name.
	Get(name string) (*profile.Profile, error)

	// List returns all profiles.
	List() ([]*profile.Profile, error)

	// Delete removes a profile by name.
	Delete(name string) error

	// SetActive sets the active profile.
	SetActive(name string) error

	// GetActive returns the active profile.
	GetActive() (*profile.Profile, error)

	// Close releases resources.
	Close() error
}

// =============================================================================
// GENERATOR INTERFACES
// =============================================================================

// PromptGenerator generates config files from prompts.
type PromptGenerator interface {
	// Generate creates the config file content (e.g., starship.toml).
	Generate(p *prompt.Prompt) (string, error)
}

// PluginGenerator generates shell plugin loading scripts.
type PluginGenerator interface {
	// Generate creates shell script content for loading plugins.
	Generate(plugins []*plugin.Plugin) (string, error)
}

// ShellGenerator generates shell configuration scripts.
type ShellGenerator interface {
	// Generate creates shell script content (aliases, env, functions).
	Generate(s *shell.Shell) (string, error)
}

// ProfileGenerator generates complete terminal configurations.
type ProfileGenerator interface {
	// Generate creates all config files from a profile.
	Generate(p *profile.Profile) (*profile.GeneratedConfig, error)
}

// =============================================================================
// LIBRARY INTERFACES
// =============================================================================

// PromptLibrary provides access to pre-built prompt definitions.
type PromptLibrary interface {
	// Get retrieves a prompt by name.
	Get(name string) (*prompt.Prompt, error)

	// List returns all prompts.
	List() []*prompt.Prompt

	// ListByCategory returns prompts in a category.
	ListByCategory(category string) []*prompt.Prompt

	// Categories returns all available categories.
	Categories() []string

	// Names returns all prompt names.
	Names() []string
}

// PluginLibrary provides access to pre-built plugin definitions.
type PluginLibrary interface {
	// Get retrieves a plugin by name.
	Get(name string) (*plugin.Plugin, error)

	// List returns all plugins.
	List() []*plugin.Plugin

	// ListByCategory returns plugins in a category.
	ListByCategory(category string) []*plugin.Plugin

	// Categories returns all available categories.
	Categories() []string

	// Names returns all plugin names.
	Names() []string
}
