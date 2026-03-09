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
  dvm library list plugins               # List nvim plugins
  dvm library list themes                # List nvim themes
  dvm library list terminal prompts      # List terminal prompts
  dvm lib ls np                          # Short form: nvim plugins
  dvm library show plugin telescope      # Show plugin details
  dvm library show theme coolnight-ocean # Show theme details`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// libraryListCmd is the 'list' subcommand
var libraryListCmd = &cobra.Command{
	Use:     "list [resource-type]",
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
  dvm library list plugins
  dvm library list themes
  dvm library list terminal prompts
  dvm lib ls np                    # Short form
  dvm library list plugins -o yaml`,
	Args: cobra.MinimumNArgs(1),
	RunE: runLibraryList,
}

// libraryShowCmd is the 'show' subcommand
var libraryShowCmd = &cobra.Command{
	Use:   "show [resource-type] [name]",
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
  dvm library show plugin telescope
  dvm library show theme coolnight-ocean
  dvm library show nvim-package core
  dvm library show terminal prompt starship-default
  dvm library show terminal plugin zsh-autosuggestions
  dvm library show terminal-package core`,
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

	// Add library command to root
	rootCmd.AddCommand(libraryCmd)
}
