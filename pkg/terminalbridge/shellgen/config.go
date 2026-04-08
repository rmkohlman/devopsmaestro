// Package shellgen generates workspace-scoped shell configuration files by
// compositing the host's shell config with workspace-specific environment
// variables, prompt configuration, and plugin source lines.
// Supports Zsh, Bash, and Fish shells.
package shellgen

import (
	"github.com/rmkohlman/MaestroTerminal/terminalops/plugin"
	"github.com/rmkohlman/MaestroTerminal/terminalops/shell"
)

// ShellConfig holds all inputs needed to generate a workspace shell config file.
type ShellConfig struct {
	// HostShellConfigPath is the path to the user's host shell config file.
	// For Zsh: ~/.zshrc, Bash: ~/.bashrc, Fish: ~/.config/fish/config.fish.
	// If empty, defaults based on shell type. If the file doesn't exist, it's skipped.
	HostShellConfigPath string

	// HostZshrcPath is deprecated — use HostShellConfigPath instead.
	// Kept for backward compatibility with existing callers.
	HostZshrcPath string

	// EnvVars are workspace-specific environment variables to inject.
	EnvVars map[string]string

	// PromptConfig is raw prompt initialization to append (e.g., starship init).
	PromptConfig string

	// Plugins are the shell plugins to include in load order.
	Plugins []*plugin.Plugin

	// PluginDir overrides the default plugin installation directory.
	PluginDir string

	// ShellConfigs are additional dvt shell configurations to include.
	ShellConfigs []*shell.Shell

	// WorkspaceName identifies the workspace (used in comments/headers).
	WorkspaceName string

	// IncludeHostConfig controls whether to source the host's shell config.
	// Defaults to true when not explicitly set.
	IncludeHostConfig *bool

	// IncludeHostZshrc is deprecated — use IncludeHostConfig instead.
	// Kept for backward compatibility with existing callers.
	IncludeHostZshrc *bool
}

// ShouldIncludeHostConfig returns whether the host shell config should be included.
func (c *ShellConfig) ShouldIncludeHostConfig() bool {
	// Check the new field first
	if c.IncludeHostConfig != nil {
		return *c.IncludeHostConfig
	}
	// Fall back to the deprecated field for backward compatibility
	if c.IncludeHostZshrc != nil {
		return *c.IncludeHostZshrc
	}
	return true
}

// ShouldIncludeHostZshrc returns whether host .zshrc should be included.
// Deprecated: use ShouldIncludeHostConfig instead.
func (c *ShellConfig) ShouldIncludeHostZshrc() bool {
	return c.ShouldIncludeHostConfig()
}

// GetHostConfigPath returns the effective host config path, checking both
// the new HostShellConfigPath and the deprecated HostZshrcPath fields.
func (c *ShellConfig) GetHostConfigPath() string {
	if c.HostShellConfigPath != "" {
		return c.HostShellConfigPath
	}
	return c.HostZshrcPath
}
