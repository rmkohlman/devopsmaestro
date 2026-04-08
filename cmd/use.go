package cmd

import (
	"devopsmaestro/db"
	"encoding/json"
	"fmt"
	"github.com/rmkohlman/MaestroNvim/nvimops/package/library"
	"github.com/rmkohlman/MaestroSDK/render"
	terminalpkglib "github.com/rmkohlman/MaestroTerminal/terminalops/package/library"

	"github.com/spf13/cobra"
)

// Dry-run flags for use subcommands
var (
	useAppDryRun       bool
	useWorkspaceDryRun bool
)

// previousContext is the JSON structure stored in defaults under "context.previous".
// Fields are pointers so absent context levels can be represented as null/absent.
type previousContext struct {
	EcosystemID *int `json:"ecosystem_id,omitempty"`
	DomainID    *int `json:"domain_id,omitempty"`
	AppID       *int `json:"app_id,omitempty"`
	WorkspaceID *int `json:"workspace_id,omitempty"`
}

// saveCurrentContext captures the current DB context and stores it as JSON
// in the defaults table under key "context.previous".
func saveCurrentContext(ds db.DataStore) error {
	ctx, err := ds.GetContext()
	if err != nil {
		return fmt.Errorf("failed to read current context: %w", err)
	}

	prev := previousContext{}
	if ctx != nil {
		prev.EcosystemID = ctx.ActiveEcosystemID
		prev.DomainID = ctx.ActiveDomainID
		prev.AppID = ctx.ActiveAppID
		prev.WorkspaceID = ctx.ActiveWorkspaceID
	}

	data, err := json.Marshal(prev)
	if err != nil {
		return fmt.Errorf("failed to serialize context: %w", err)
	}

	return ds.SetDefault("context.previous", string(data))
}

// restorePreviousContext reads "context.previous" from defaults, restores it
// as the active context, and saves the old active context as the new previous.
func restorePreviousContext(ds db.DataStore) error {
	// Read previous context
	prevJSON, err := ds.GetDefault("context.previous")
	if err != nil || prevJSON == "" {
		return fmt.Errorf("no previous context to switch to")
	}

	var prev previousContext
	if err := json.Unmarshal([]byte(prevJSON), &prev); err != nil {
		return fmt.Errorf("failed to parse previous context: %w", err)
	}

	// Save current context as new previous before restoring
	if err := saveCurrentContext(ds); err != nil {
		return fmt.Errorf("failed to save current context: %w", err)
	}

	// Restore all 4 context fields from previous
	if err := ds.SetActiveEcosystem(prev.EcosystemID); err != nil {
		return fmt.Errorf("failed to restore ecosystem: %w", err)
	}
	if err := ds.SetActiveDomain(prev.DomainID); err != nil {
		return fmt.Errorf("failed to restore domain: %w", err)
	}
	if err := ds.SetActiveApp(prev.AppID); err != nil {
		return fmt.Errorf("failed to restore app: %w", err)
	}
	if err := ds.SetActiveWorkspace(prev.WorkspaceID); err != nil {
		return fmt.Errorf("failed to restore workspace: %w", err)
	}

	render.Success("Switched to previous context")
	return nil
}

// useCmd represents the base 'use' command (kubectl-style context switching)
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Switch active context",
	Long: `Switch the active context (kubectl-style).

Use 'none' as the name to clear the context, or use --clear to clear all context.

Resource aliases (kubectl-style):
  ecosystem → eco
  domain    → dom
  app       → a, application
  workspace → ws

Environment variables for per-terminal-tab context:
  DVM_ECOSYSTEM    Override active ecosystem
  DVM_DOMAIN       Override active domain
  DVM_APP          Override active app
  DVM_WORKSPACE    Override active workspace

Resolution order: flags > env vars > stored context

Examples:
  dvm use ecosystem my-platform  # Set active ecosystem
  dvm use domain backend         # Set active domain
  dvm use app my-api             # Set active app
  dvm use a my-api               # Short form
  dvm use workspace dev          # Set active workspace
  dvm use ws dev                 # Short form
  dvm use app none               # Clear app context
  dvm use workspace none         # Clear workspace context
  dvm use --clear                # Clear all context
  dvm use app myapi --export     # Print 'export DVM_APP=myapi' for shell eval`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle "dvm use -" to toggle to previous context
		if len(args) == 1 && args[0] == "-" {
			ds, err := getDataStore(cmd)
			if err != nil {
				return fmt.Errorf("dataStore not initialized: %w", err)
			}
			return restorePreviousContext(ds)
		}

		// Check if --clear flag was passed
		clearAll, _ := cmd.Flags().GetBool("clear")
		if clearAll {
			ds, err := getDataStore(cmd)
			if err != nil {
				return fmt.Errorf("dataStore not initialized: %w", err)
			}

			// Clear all 4 DB context fields
			if err := ds.SetActiveEcosystem(nil); err != nil {
				return fmt.Errorf("failed to clear ecosystem context: %v", err)
			}
			if err := ds.SetActiveDomain(nil); err != nil {
				return fmt.Errorf("failed to clear domain context: %v", err)
			}
			if err := ds.SetActiveApp(nil); err != nil {
				return fmt.Errorf("failed to clear app context: %v", err)
			}
			if err := ds.SetActiveWorkspace(nil); err != nil {
				return fmt.Errorf("failed to clear workspace context: %v", err)
			}

			render.Success("Cleared all context (ecosystem, domain, app, and workspace)")
			return nil
		}

		// If no --clear flag and no subcommand, show help
		return cmd.Help()
	},
}

// useAppCmd switches the active app
var useAppCmd = &cobra.Command{
	Use:     "app <name>",
	Aliases: []string{"a", "application"},
	Short:   "Switch to an app",
	Long: `Set the specified app as the active context.

Use 'none' as the name to clear the app context (also clears workspace).

Examples:
  dvm use app my-api            # Set active app
  dvm use a my-api              # Short form
  dvm use app frontend          # Switch to another app
  dvm use app none              # Clear app context`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		appName := args[0]

		// Handle "none" to clear context
		if appName == "none" {
			ds, err := getDataStore(cmd)
			if err != nil {
				return fmt.Errorf("dataStore not initialized: %w", err)
			}

			if err := ds.SetActiveApp(nil); err != nil {
				return fmt.Errorf("failed to clear app context: %v", err)
			}
			if err := ds.SetActiveWorkspace(nil); err != nil {
				return fmt.Errorf("failed to clear workspace context: %v", err)
			}

			render.Success("Cleared app context (workspace also cleared)")
			return nil
		}

		// Get datastore from context
		ds, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("dataStore not initialized: %w", err)
		}

		// Verify app exists (search globally across all domains)
		app, err := ds.GetAppByNameGlobal(appName)
		if err != nil {
			render.Error(fmt.Sprintf("App '%s' not found: %v", appName, err))
			render.Plain(FormatSuggestions(SuggestAppNotFound(appName)...))
			return errSilent
		}

		// Handle --export flag: print export statement and return
		exportFlag, _ := cmd.Flags().GetBool("export")
		if exportFlag {
			fmt.Fprintf(cmd.OutOrStdout(), "export DVM_APP=%s\n", appName)
			return nil
		}

		// Dry-run: preview what would happen
		if useAppDryRun {
			render.Plain(fmt.Sprintf("Would switch active app to %q", appName))
			return nil
		}

		// Save current context before switching
		if err := saveCurrentContext(ds); err != nil {
			return fmt.Errorf("failed to save previous context: %w", err)
		}

		// Set app as active in database context
		if err := ds.SetActiveApp(&app.ID); err != nil {
			return fmt.Errorf("failed to set active app: %v", err)
		}

		render.Success(fmt.Sprintf("Switched to app '%s'", appName))
		render.Info(fmt.Sprintf("Path: %s", app.Path))
		render.Blank()
		render.Info("Next: Select a workspace with: dvm use workspace <name>")
		return nil
	},
}

// useWorkspaceCmd switches the active workspace
var useWorkspaceCmd = &cobra.Command{
	Use:     "workspace <name>",
	Aliases: []string{"ws"},
	Short:   "Switch to a workspace",
	Long: `Set the specified workspace as the active context.
Requires an active app to be set first (unless clearing with 'none').

Use 'none' as the name to clear the workspace context (keeps app).

Examples:
  dvm use workspace main        # Set active workspace
  dvm use ws main               # Short form
  dvm use workspace dev         # Switch to another workspace
  dvm use workspace none        # Clear workspace context`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceName := args[0]

		// Handle "none" to clear context
		if workspaceName == "none" {
			ds, err := getDataStore(cmd)
			if err != nil {
				return fmt.Errorf("dataStore not initialized: %w", err)
			}

			if err := ds.SetActiveWorkspace(nil); err != nil {
				return fmt.Errorf("failed to clear workspace context: %v", err)
			}

			render.Success("Cleared workspace context")
			return nil
		}

		// Get datastore from context
		ds, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("dataStore not initialized: %w", err)
		}

		// Get active app (DB-backed)
		appName, err := getActiveAppFromContext(ds)
		if err != nil {
			render.Error("No active app set")
			render.Plain(FormatSuggestions(SuggestNoActiveApp()...))
			return errSilent
		}

		// Get app to get its ID
		app, err := ds.GetAppByNameGlobal(appName)
		if err != nil {
			return fmt.Errorf("failed to get app: %v", err)
		}

		// Verify workspace exists
		workspace, err := ds.GetWorkspaceByName(app.ID, workspaceName)
		if err != nil {
			render.Error(fmt.Sprintf("Workspace '%s' not found in app '%s': %v", workspaceName, appName, err))
			render.Plain(FormatSuggestions(SuggestWorkspaceNotFound(workspaceName)...))
			return errSilent
		}

		// Handle --export flag: print export statement and return
		exportFlag, _ := cmd.Flags().GetBool("export")
		if exportFlag {
			fmt.Fprintf(cmd.OutOrStdout(), "export DVM_WORKSPACE=%s\n", workspaceName)
			return nil
		}

		// Dry-run: preview what would happen
		if useWorkspaceDryRun {
			render.Plain(fmt.Sprintf("Would switch active workspace to %q in app %q", workspaceName, appName))
			return nil
		}

		// Save current context before switching
		if err := saveCurrentContext(ds); err != nil {
			return fmt.Errorf("failed to save previous context: %w", err)
		}

		// Update database context
		if err := ds.SetActiveWorkspace(&workspace.ID); err != nil {
			return fmt.Errorf("failed to set active workspace: %v", err)
		}

		render.Success(fmt.Sprintf("Switched to workspace '%s' in app '%s'", workspaceName, appName))
		render.Blank()
		render.Info("Next: Attach to your workspace with: dvm attach")
		return nil
	},
}

// useNvimCmd manages nvim-related use subcommands
var useNvimCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Manage Neovim defaults",
	Long: `Manage Neovim defaults (kubectl-style).

Use these commands to set default Neovim configurations.

Examples:
  dvm use nvim package core     # Set default package to 'core'
  dvm use nvim package none     # Clear default package`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// useNvimPackageCmd sets the default nvim package
var useNvimPackageCmd = &cobra.Command{
	Use:   "package <name>",
	Short: "Set default nvim package",
	Long: `Set the default Neovim package for new workspaces.

The default package will be used when creating new workspaces that don't 
specify a custom package. Packages group related plugins into reusable bundles.

Use 'none' to clear the default package.

Available packages can be found via: nvp package list

Examples:
  dvm use nvim package core     # Set default to 'core' package
  dvm use nvim package go-dev   # Set default to 'go-dev' package  
  dvm use nvim package none     # Clear default package`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packageName := args[0]

		// Get datastore from context
		ds, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("dataStore not initialized: %w", err)
		}

		// Handle "none" to clear default
		if packageName == "none" {
			if err := ds.DeleteDefault("nvim-package"); err != nil {
				return fmt.Errorf("failed to clear default nvim package: %v", err)
			}

			render.Success("Default nvim package cleared")
			return nil
		}

		// Validate package exists
		if err := validatePackageExists(packageName, ds); err != nil {
			render.Error(fmt.Sprintf("Package '%s' not found: %v", packageName, err))
			render.Info("Hint: List available packages with: nvp package list")
			return errSilent
		}

		// Set the default package
		if err := ds.SetDefault("nvim-package", packageName); err != nil {
			return fmt.Errorf("failed to set default nvim package: %v", err)
		}

		render.Success(fmt.Sprintf("Default nvim package set to '%s'", packageName))
		render.Info("This package will be used for new workspaces that don't specify a custom package")
		return nil
	},
}

// validatePackageExists checks if a package exists in database or library
func validatePackageExists(packageName string, ds db.DataStore) error {
	// First check if package exists in database (user packages)
	_, err := ds.GetPackage(packageName)
	if err == nil {
		return nil // Found in database
	}

	// If not in database, check library packages
	lib, err := library.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load package library: %w", err)
	}

	if _, ok := lib.Get(packageName); ok {
		return nil // Found in library
	}

	return fmt.Errorf("package not found in database or library")
}

// useTerminalCmd manages terminal-related use subcommands
var useTerminalCmd = &cobra.Command{
	Use:   "terminal",
	Short: "Manage Terminal defaults",
	Long: `Manage Terminal defaults (kubectl-style).

Use these commands to set default Terminal configurations.

Examples:
  dvm use terminal package developer-essentials    # Set default package to 'developer-essentials'
  dvm use terminal package none                    # Clear default package`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

// useTerminalPackageCmd sets the default terminal package
var useTerminalPackageCmd = &cobra.Command{
	Use:   "package <name>",
	Short: "Set default terminal package",
	Long: `Set the default Terminal package for new workspaces.

The default package will be used when creating new workspaces that don't 
specify a custom package. Packages group related plugins, prompts, and 
configurations into reusable bundles.

Use 'none' to clear the default package.

Available packages can be found via the terminal package list command.

Examples:
  dvm use terminal package developer-essentials   # Set default to 'developer-essentials' package
  dvm use terminal package poweruser              # Set default to 'poweruser' package  
  dvm use terminal package none                   # Clear default package`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		packageName := args[0]

		// Get datastore from context
		ds, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("dataStore not initialized: %w", err)
		}

		// Handle "none" to clear default
		if packageName == "none" {
			if err := ds.DeleteDefault("terminal-package"); err != nil {
				return fmt.Errorf("failed to clear default terminal package: %v", err)
			}

			render.Success("Default terminal package cleared")
			return nil
		}

		// Validate package exists
		if err := validateTerminalPackageExists(packageName, ds); err != nil {
			render.Error(fmt.Sprintf("Package '%s' not found: %v", packageName, err))
			render.Info("Hint: List available packages with the terminal package list command")
			return errSilent
		}

		// Set the default package
		if err := ds.SetDefault("terminal-package", packageName); err != nil {
			return fmt.Errorf("failed to set default terminal package: %v", err)
		}

		render.Success(fmt.Sprintf("Default terminal package set to '%s'", packageName))
		render.Info("This package will be used for new workspaces that don't specify a custom package")
		return nil
	},
}

// validateTerminalPackageExists checks if a terminal package exists in database or library
func validateTerminalPackageExists(packageName string, ds db.DataStore) error {
	// First check if package exists in database (user packages)
	_, err := ds.GetTerminalPackage(packageName)
	if err == nil {
		return nil // Found in database
	}

	// If not in database, check library packages
	lib, err := terminalpkglib.NewLibrary()
	if err != nil {
		return fmt.Errorf("failed to load package library: %w", err)
	}

	if _, ok := lib.Get(packageName); ok {
		return nil // Found in library
	}

	return fmt.Errorf("package not found in database or library")
}

// Initializes the 'use' command and links subcommands
func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.AddCommand(useAppCmd)
	useCmd.AddCommand(useWorkspaceCmd)
	useCmd.AddCommand(useNvimCmd)
	useNvimCmd.AddCommand(useNvimPackageCmd)
	useCmd.AddCommand(useTerminalCmd)
	useTerminalCmd.AddCommand(useTerminalPackageCmd)
	useCmd.Flags().Bool("clear", false, "Clear all context (app and workspace)")

	// Register argument completions for subcommands
	useAppCmd.ValidArgsFunction = completeApps
	useWorkspaceCmd.ValidArgsFunction = completeWorkspaces

	// Register --export flag on all 4 use subcommands
	useEcosystemCmd.Flags().Bool("export", false, "Print 'export DVM_ECOSYSTEM=<name>' for shell eval")
	useDomainCmd.Flags().Bool("export", false, "Print 'export DVM_DOMAIN=<name>' for shell eval")
	useAppCmd.Flags().Bool("export", false, "Print 'export DVM_APP=<name>' for shell eval")
	useWorkspaceCmd.Flags().Bool("export", false, "Print 'export DVM_WORKSPACE=<name>' for shell eval")

	// Register --dry-run flag on all 4 use subcommands
	AddDryRunFlag(useAppCmd, &useAppDryRun)
	AddDryRunFlag(useWorkspaceCmd, &useWorkspaceDryRun)
}
