package cmd

import (
	"github.com/spf13/cobra"
)

// nvimGetCmd is the 'nvim' subcommand under 'get' for kubectl-style namespacing
// Usage: dvm get nvim plugins, dvm get nvim plugin <name>
var nvimGetCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Get nvim resources (plugins, themes)",
	Long: `Get nvim-related resources in kubectl-style namespaced format.

Examples:
  dvm get nvim plugins              # List all nvim plugins
  dvm get nvim plugin telescope     # Get specific plugin
  dvm get nvim plugins -o yaml      # Output as YAML
  dvm get nvim themes               # List all nvim themes (future)
`,
}

// nvimGetPluginsCmd lists all nvim plugins (namespaced version)
// Usage: dvm get nvim plugins
var nvimGetPluginsCmd = &cobra.Command{
	Use:     "plugins",
	Aliases: []string{"np"},
	Short:   "List all nvim plugins",
	Long: `List all nvim plugins stored in the database.

Examples:
  dvm get nvim plugins
  dvm get nvim plugins -o yaml
  dvm get nvim plugins -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Reuse existing getPlugins function from get.go
		return getPlugins(cmd)
	},
}

// nvimGetPluginCmd gets a specific nvim plugin (namespaced version)
// Usage: dvm get nvim plugin <name>
var nvimGetPluginCmd = &cobra.Command{
	Use:   "plugin [name]",
	Short: "Get a specific nvim plugin",
	Long: `Get a specific nvim plugin by name.

Examples:
  dvm get nvim plugin telescope
  dvm get nvim plugin telescope -o yaml
  dvm get nvim plugin lspconfig -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Reuse existing getPlugin function from get.go
		return getPlugin(cmd, args[0])
	},
}

// nvimGetThemesCmd lists all nvim themes (namespaced version)
// Usage: dvm get nvim themes
var nvimGetThemesCmd = &cobra.Command{
	Use:     "themes",
	Aliases: []string{"nt"},
	Short:   "List all nvim themes",
	Long: `List all nvim themes stored in the database.

Examples:
  dvm get nvim themes
  dvm get nvim themes -o yaml
  dvm get nvim themes -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getThemes(cmd)
	},
}

// nvimGetThemeCmd gets a specific nvim theme (namespaced version)
// Usage: dvm get nvim theme <name>
var nvimGetThemeCmd = &cobra.Command{
	Use:   "theme [name]",
	Short: "Get a specific nvim theme",
	Long: `Get a specific nvim theme by name.

Examples:
  dvm get nvim theme tokyonight
  dvm get nvim theme tokyonight -o yaml
  dvm get nvim theme catppuccin -o json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return getTheme(cmd, args[0])
	},
}

func init() {
	// Add nvim subcommand to get
	getCmd.AddCommand(nvimGetCmd)

	// Add resource types under nvim
	nvimGetCmd.AddCommand(nvimGetPluginsCmd)
	nvimGetCmd.AddCommand(nvimGetPluginCmd)
	nvimGetCmd.AddCommand(nvimGetThemesCmd)
	nvimGetCmd.AddCommand(nvimGetThemeCmd)
}
