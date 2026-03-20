package cmd

import (
	"github.com/spf13/cobra"
)

// libraryCmd is the root 'library' command
var libraryCmd = &cobra.Command{
	Use:     "library",
	Aliases: []string{"lib"},
	Short:   "Browse plugin and theme libraries",
	Long: `Browse plugin and theme libraries.

View available plugins, themes, prompts, and packages from the embedded libraries.

Examples:
  dvm library get plugins                # List nvim plugins
  dvm library get themes                 # List nvim themes
  dvm library get terminal prompts       # List terminal prompts
  dvm lib ls np                          # Short form: nvim plugins
  dvm library describe plugin telescope      # Show plugin details
  dvm library describe theme coolnight-ocean # Show theme details`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// libraryListCmd is the 'get' subcommand (formerly 'list')
var libraryListCmd = &cobra.Command{
	Use:     "get [resource-type]",
	Aliases: []string{"ls"},
	Short:   "List library resources",
	Long: `List library resources (plugins, themes, prompts, packages).

Resource types:
  plugins               - List nvim plugins (alias: np)
  themes                - List nvim themes (alias: nt)
  nvim packages         - List nvim plugin bundles
  terminal prompts      - List terminal prompts (alias: tp)
  terminal plugins      - List shell plugins (alias: tpl)
  terminal packages     - List terminal bundles

Examples:
  dvm library get plugins
  dvm library get themes
  dvm library get terminal prompts
  dvm lib ls np                    # Short form
  dvm library get plugins -o yaml`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLibraryList,
}

// libraryShowCmd is the 'describe' subcommand (formerly 'show')
var libraryShowCmd = &cobra.Command{
	Use:   "describe [resource-type] [name]",
	Short: "Show library resource details",
	Long: `Show details of a specific library resource.

Resource types:
  plugin [name]               - Show nvim plugin details
  theme [name]                - Show nvim theme details
  nvim-package [name]         - Show nvim package details
  terminal prompt [name]      - Show terminal prompt details
  terminal plugin [name]      - Show terminal plugin details
  terminal-package [name]     - Show terminal package details

Examples:
  dvm library describe plugin telescope
  dvm library describe theme coolnight-ocean
  dvm library describe nvim-package core
  dvm library describe terminal prompt starship-default
  dvm library describe terminal plugin zsh-autosuggestions
  dvm library describe terminal-package core`,
	Args: cobra.MinimumNArgs(2),
	RunE: runLibraryShow,
}

// libraryImportCmd is the 'import' subcommand
var libraryImportCmd = &cobra.Command{
	Use:   "import [resource-type...]",
	Short: "Import library resources into the database",
	Long: `Import library resources into the database.

Imports embedded library resources into the local database for use with workspaces.

Resource types:
  nvim-plugins          - Import nvim plugins
  nvim-themes           - Import nvim themes
  nvim-packages         - Import nvim packages
  terminal-prompts      - Import terminal prompts
  terminal-plugins      - Import terminal plugins
  terminal-packages     - Import terminal packages

Flags:
  --all                 - Import all resource types

Examples:
  dvm library import nvim-plugins          # Import just nvim plugins
  dvm library import nvim-themes           # Import just nvim themes
  dvm library import --all                 # Import all resource types`,
	RunE: runLibraryImport,
}

func init() {
	// Add output format flag to list and show commands
	libraryListCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")
	libraryShowCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")

	// Add flags to import command
	libraryImportCmd.Flags().Bool("all", false, "Import all resource types")
	libraryImportCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")

	// Add subcommands
	libraryCmd.AddCommand(libraryListCmd)
	libraryCmd.AddCommand(libraryShowCmd)
	libraryCmd.AddCommand(libraryImportCmd)

	// Hidden backward-compat aliases for deprecated verbs (list→get, show→describe)
	libraryCmd.AddCommand(hiddenAlias("list", libraryListCmd))
	libraryCmd.AddCommand(hiddenAlias("show", libraryShowCmd))

	// Add library command to root
	rootCmd.AddCommand(libraryCmd)
}

// hiddenAlias creates a hidden command that delegates to the target command.
// Used to keep deprecated verb names (list, show, install) working without
// showing them in --help output.
func hiddenAlias(name string, target *cobra.Command) *cobra.Command {
	alias := *target
	alias.Use = name
	alias.Aliases = nil
	alias.Hidden = true
	alias.Short = target.Short + " (deprecated: use " + target.Name() + ")"
	// Deprecated field causes Cobra to print a deprecation notice when used
	alias.Deprecated = "use '" + target.Name() + "' instead"
	return &alias
}
