package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/nvimops/package/library"
	"devopsmaestro/render"
	"fmt"

	"github.com/spf13/cobra"
)

// useCmd represents the base 'use' command (kubectl-style context switching)
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Switch active context",
	Long: `Switch the active app or workspace context (kubectl-style).

Use 'none' as the name to clear the context, or use --clear to clear all context.

Resource aliases (kubectl-style):
  app       → a
  workspace → ws

Examples:
  dvm use app my-api            # Set active app
  dvm use a my-api              # Short form
  dvm use workspace dev         # Set active workspace
  dvm use ws dev                # Short form
  dvm use app none              # Clear app context
  dvm use workspace none        # Clear workspace context
  dvm use --clear               # Clear all context`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if --clear flag was passed
		clearAll, _ := cmd.Flags().GetBool("clear")
		if clearAll {
			contextMgr, err := operators.NewContextManager()
			if err != nil {
				return fmt.Errorf("failed to initialize context manager: %v", err)
			}

			if err := contextMgr.ClearApp(); err != nil {
				return fmt.Errorf("failed to clear context: %v", err)
			}

			render.Success("Cleared all context (app and workspace)")
			return nil
		}

		// If no --clear flag and no subcommand, show help
		return cmd.Help()
	},
}

// useAppCmd switches the active app
var useAppCmd = &cobra.Command{
	Use:     "app <name>",
	Aliases: []string{"a"},
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
			contextMgr, err := operators.NewContextManager()
			if err != nil {
				return fmt.Errorf("failed to initialize context manager: %v", err)
			}

			if err := contextMgr.ClearApp(); err != nil {
				return fmt.Errorf("failed to clear app context: %v", err)
			}

			// Also clear database context
			ctx := cmd.Context()
			dataStore := ctx.Value("dataStore").(*db.DataStore)
			if dataStore != nil {
				ds := *dataStore
				ds.SetActiveApp(nil)
				ds.SetActiveWorkspace(nil)
			}

			render.Success("Cleared app context (workspace also cleared)")
			return nil
		}

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			return fmt.Errorf("dataStore not initialized")
		}

		ds := *dataStore

		// Verify app exists (search globally across all domains)
		app, err := ds.GetAppByNameGlobal(appName)
		if err != nil {
			render.Error(fmt.Sprintf("App '%s' not found: %v", appName, err))
			render.Info("Hint: List available apps with: dvm get apps")
			return nil
		}

		// Set app as active in context manager
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			return fmt.Errorf("failed to initialize context manager: %v", err)
		}

		if err := contextMgr.SetApp(appName); err != nil {
			return fmt.Errorf("failed to set active app: %v", err)
		}

		// Also update database context
		if err := ds.SetActiveApp(&app.ID); err != nil {
			render.Warning(fmt.Sprintf("Failed to update database context: %v", err))
		}

		render.Success(fmt.Sprintf("Switched to app '%s'", appName))
		render.Info(fmt.Sprintf("Path: %s", app.Path))
		fmt.Println()
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
			contextMgr, err := operators.NewContextManager()
			if err != nil {
				return fmt.Errorf("failed to initialize context manager: %v", err)
			}

			if err := contextMgr.ClearWorkspace(); err != nil {
				return fmt.Errorf("failed to clear workspace context: %v", err)
			}

			// Also clear database context
			ctx := cmd.Context()
			dataStore := ctx.Value("dataStore").(*db.DataStore)
			if dataStore != nil {
				ds := *dataStore
				ds.SetActiveWorkspace(nil)
			}

			render.Success("Cleared workspace context")
			return nil
		}

		// Get context manager
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			return fmt.Errorf("failed to initialize context manager: %v", err)
		}

		// Get active app
		appName, err := contextMgr.GetActiveApp()
		if err != nil {
			render.Error("No active app set")
			render.Info("Hint: Set active app first with: dvm use app <name>")
			return nil
		}

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			return fmt.Errorf("dataStore not initialized")
		}

		ds := *dataStore

		// Get app to get its ID
		app, err := ds.GetAppByNameGlobal(appName)
		if err != nil {
			return fmt.Errorf("failed to get app: %v", err)
		}

		// Verify workspace exists
		workspace, err := ds.GetWorkspaceByName(app.ID, workspaceName)
		if err != nil {
			render.Error(fmt.Sprintf("Workspace '%s' not found in app '%s': %v", workspaceName, appName, err))
			render.Info("Hint: List available workspaces with: dvm get workspaces")
			return nil
		}

		// Set workspace as active in context manager
		if err := contextMgr.SetWorkspace(workspaceName); err != nil {
			return fmt.Errorf("failed to set active workspace: %v", err)
		}

		// Also update database context
		if err := ds.SetActiveWorkspace(&workspace.ID); err != nil {
			render.Warning(fmt.Sprintf("Failed to update database context: %v", err))
		}

		render.Success(fmt.Sprintf("Switched to workspace '%s' in app '%s'", workspaceName, appName))
		fmt.Println()
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
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			return fmt.Errorf("dataStore not initialized")
		}

		ds := *dataStore

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
			return nil
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

// Initializes the 'use' command and links subcommands
func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.AddCommand(useAppCmd)
	useCmd.AddCommand(useWorkspaceCmd)
	useCmd.AddCommand(useNvimCmd)
	useNvimCmd.AddCommand(useNvimPackageCmd)
	useCmd.Flags().Bool("clear", false, "Clear all context (app and workspace)")

	// Register argument completions for subcommands
	useAppCmd.ValidArgsFunction = completeApps
	useWorkspaceCmd.ValidArgsFunction = completeWorkspaces
}
