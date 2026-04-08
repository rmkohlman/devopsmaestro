// Package shellgen generates workspace-scoped .zshrc files by compositing
// the host's shell config with workspace-specific environment variables,
// prompt configuration, and plugin source lines.
package shellgen

import (
	"github.com/rmkohlman/MaestroTerminal/terminalops/plugin"
	"github.com/rmkohlman/MaestroTerminal/terminalops/shell"
)

// ShellConfig holds all inputs needed to generate a .zshrc.workspace file.
type ShellConfig struct {
	// HostZshrcPath is the path to the user's host ~/.zshrc.
	// If empty, defaults to ~/.zshrc. If the file doesn't exist, it's skipped.
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

	// IncludeHostZshrc controls whether to source the host's .zshrc.
	// Defaults to true when not explicitly set.
	IncludeHostZshrc *bool
}

// ShouldIncludeHostZshrc returns whether host .zshrc should be included.
func (c *ShellConfig) ShouldIncludeHostZshrc() bool {
	if c.IncludeHostZshrc == nil {
		return true
	}
	return *c.IncludeHostZshrc
}
