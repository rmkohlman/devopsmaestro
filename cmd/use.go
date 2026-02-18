package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/operators"
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

// Initializes the 'use' command and links subcommands
func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.AddCommand(useAppCmd)
	useCmd.AddCommand(useWorkspaceCmd)
	useCmd.Flags().Bool("clear", false, "Clear all context (app and workspace)")

	// Register argument completions for subcommands
	useAppCmd.ValidArgsFunction = completeApps
	useWorkspaceCmd.ValidArgsFunction = completeWorkspaces
}
