package cmd

import (
	"github.com/spf13/cobra"
)

var (
	getOutputFormat    string
	getWorkspacesFlags HierarchyFlags
	getWorkspaceFlags  HierarchyFlags
	showTheme          bool // Flag to show theme resolution information
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [resource]",
	Short: "Get resources (kubectl-style)",
	Long: `Get resources in various formats (colored, yaml, json, plain).

Resource aliases (kubectl-style):
  apps       → a, app
  workspaces → ws
  workspace  → ws
  context    → ctx
  platforms  → plat
  nvim plugins → np
  nvim themes  → nt

Examples:
  dvm get apps
  dvm get a                       # Same as 'get apps'
  dvm get workspaces
  dvm get ws                      # Same as 'get workspaces'
  dvm get workspace main
  dvm get ws main                 # Same as 'get workspace main'
  dvm get context
  dvm get ctx                     # Same as 'get context'
  dvm get np                      # Same as 'get nvim plugins'
  dvm get nt                      # Same as 'get nvim themes' (34+ library themes)
  dvm get nvim theme coolnight-ocean    # Library theme (no install needed)
  dvm get workspace main -o yaml
  dvm get app my-api -o json
`,
}

// getWorkspacesCmd lists all workspaces in current app
var getWorkspacesCmd = &cobra.Command{
	Use:     "workspaces",
	Aliases: []string{"ws"},
	Short:   "List all workspaces in an app",
	Long: `List all workspaces in an app.

Flags:
  -A, --all         List all workspaces across all apps/domains/ecosystems
  -e, --ecosystem   Filter by ecosystem name
  -d, --domain      Filter by domain name  
  -a, --app         Filter by app name
  -w, --workspace   Filter by workspace name

Examples:
  dvm get workspaces              # List workspaces in active app
  dvm get ws                      # Short form
  dvm get workspaces -A           # List ALL workspaces across everything
  dvm get workspaces -a myapp     # List workspaces in specific app
  dvm get workspaces -e healthcare -a portal`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getWorkspaces(cmd)
	},
}

// getWorkspaceCmd gets a specific workspace
var getWorkspaceCmd = &cobra.Command{
	Use:     "workspace [name]",
	Aliases: []string{"ws"},
	Short:   "Get a specific workspace",
	Long: `Get a specific workspace by name.

Flags:
  -e, --ecosystem   Filter by ecosystem name
  -d, --domain      Filter by domain name  
  -a, --app         Filter by app name
  -w, --workspace   Filter by workspace name (alternative to positional arg)

Examples:
  dvm get workspace main              # Get workspace from active app
  dvm get ws main                     # Short form
  dvm get workspace main -a myapp     # Get workspace from specific app
  dvm get workspace -a portal         # Get workspace if only one exists
  dvm get workspace main -o yaml      # Output as YAML`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := ""
		if len(args) > 0 {
			name = args[0]
		}
		return getWorkspace(cmd, name)
	},
}

// getPlatformsCmd lists all detected container platforms
var getPlatformsCmd = &cobra.Command{
	Use:     "platforms",
	Aliases: []string{"plat"},
	Short:   "List all detected container platforms",
	Long: `List all detected container platforms (OrbStack, Colima, Docker Desktop, Podman).

Examples:
  dvm get platforms
  dvm get plat          # Short form
  dvm get platforms -o yaml
  dvm get platforms -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getPlatforms(cmd)
	},
}

// getContextCmd displays the current active context
var getContextCmd = &cobra.Command{
	Use:     "context",
	Aliases: []string{"ctx"},
	Short:   "Display the current context",
	Long: `Display the current active app and workspace context.

The context determines which app and workspace commands operate on by default.
Set context with 'dvm use app <name>' and 'dvm use workspace <name>'.

Context can also be set via environment variables:
  DVM_APP        - Override active app
  DVM_WORKSPACE  - Override active workspace

Examples:
  dvm get context       # Show current context
  dvm get ctx           # Short form
  dvm get context -o yaml
  dvm get context -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getContext(cmd)
	},
}

// getNvimPluginsShortCmd is a top-level shortcut for 'dvm get nvim plugins'
// Usage: dvm get np
var getNvimPluginsShortCmd = &cobra.Command{
	Use:   "np",
	Short: "List all nvim plugins (shortcut for 'nvim plugins')",
	Long: `List all nvim plugins stored in the database.

This is a shortcut for 'dvm get nvim plugins'.

Examples:
  dvm get np
  dvm get np -o yaml
  dvm get np -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getPlugins(cmd)
	},
}

// getNvimThemesShortCmd is a top-level shortcut for 'dvm get nvim themes'
// Usage: dvm get nt
var getNvimThemesShortCmd = &cobra.Command{
	Use:   "nt",
	Short: "List all nvim themes (shortcut for 'nvim themes')",
	Long: `List all nvim themes from user store and embedded library.

This is a shortcut for 'dvm get nvim themes'.
Shows 34+ library themes automatically available without installation.

Examples:
  dvm get nt
  dvm get nt -o yaml
  dvm get nt -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getThemes(cmd)
	},
}

// getDefaultsCmd displays default configuration values
var getDefaultsCmd = &cobra.Command{
	Use:   "defaults",
	Short: "Display default configuration values",
	Long: `Display default configuration values for containers and shells.

Shows the default values used when creating new workspaces if no explicit 
configuration is provided.

Examples:
  dvm get defaults
  dvm get defaults -o yaml
  dvm get defaults -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getDefaults(cmd)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.AddCommand(getWorkspacesCmd)
	getCmd.AddCommand(getWorkspaceCmd)
	getCmd.AddCommand(getPlatformsCmd)
	getCmd.AddCommand(getContextCmd)
	getCmd.AddCommand(getDefaultsCmd)

	// Add top-level shortcuts for nvim resources
	getCmd.AddCommand(getNvimPluginsShortCmd)
	getCmd.AddCommand(getNvimThemesShortCmd)

	// Registry resource commands
	getCmd.AddCommand(getRegistriesCmd)
	getCmd.AddCommand(getRegistryCmd)

	// Add output format flag to all get commands
	// Maps to render package: json, yaml, plain, table, colored (default)
	getCmd.PersistentFlags().StringVarP(&getOutputFormat, "output", "o", "", "Output format (json, yaml, plain, table, colored)")

	// Add hierarchy flags for workspace commands
	AddHierarchyFlags(getWorkspacesCmd, &getWorkspacesFlags)
	AddHierarchyFlags(getWorkspaceCmd, &getWorkspaceFlags)

	// Add --all flag to get workspaces (with -A shorthand for consistency)
	getWorkspacesCmd.Flags().BoolP("all", "A", false, "List all workspaces across all apps/domains/ecosystems")

	// Add --show-theme flag to hierarchy commands
	getWorkspacesCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
	getWorkspaceCmd.Flags().BoolVar(&showTheme, "show-theme", false, "Show theme resolution information")
}
