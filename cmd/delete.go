package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// deleteCmd is the root 'delete' command for kubectl-style resource deletion
// Usage: dvm delete nvim plugin <name>
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources",
	Long: `Delete resources by name.

Resource aliases (kubectl-style):
  app       → a
  workspace → ws

Examples:
  dvm delete app my-api                        # Delete app and its workspaces
  dvm delete a my-api                          # Short form
  dvm delete workspace dev                     # Delete workspace from active app
  dvm delete ws dev                            # Short form
  dvm delete workspace dev -p myapp            # Delete workspace from specific app
  dvm delete nvim plugin telescope             # Delete nvim plugin`,
}

// deleteNvimCmd is the 'nvim' subcommand under 'delete'
var deleteNvimCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Delete nvim resources",
	Long: `Delete nvim-related resources (plugins, themes).

Examples:
  dvm delete nvim plugin telescope
  dvm delete nvim theme tokyonight`,
}

// Flags for delete nvim plugin
var (
	deleteNvimWorkspaceFlag string
	deleteNvimAppFlag       string
)

// deleteNvimPluginCmd deletes a nvim plugin (from global library or workspace)
// Usage: dvm delete nvim plugin <name>           # Delete from global library
// Usage: dvm delete nvim plugin -w <ws> <name>   # Remove from workspace config
var deleteNvimPluginCmd = &cobra.Command{
	Use:   "plugin [names...]",
	Short: "Delete nvim plugin(s)",
	Long: `Delete nvim plugin(s) from the global library or remove from a workspace.

Without -w flag: Deletes plugin definition from global library (~/.nvp/plugins/).
With -w flag:    Removes plugin(s) from workspace configuration (keeps in library).

Examples:
  dvm delete nvim plugin telescope              # Delete from global library
  dvm delete nvim plugin telescope --force      # Skip confirmation
  dvm delete nvim plugin -w dev telescope       # Remove from workspace 'dev'
  dvm delete nvim plugin -w dev treesitter lsp  # Remove multiple from workspace`,
	Args: cobra.MinimumNArgs(1),
	RunE: runDeleteNvimPlugin,
}

func runDeleteNvimPlugin(cmd *cobra.Command, args []string) error {
	// If workspace flag provided, remove from workspace config
	if deleteNvimWorkspaceFlag != "" {
		return runDeleteWorkspacePlugins(cmd, args)
	}

	// Otherwise, delete from global library (original behavior)
	return runDeleteGlobalPlugin(cmd, args[0])
}

func runDeleteGlobalPlugin(cmd *cobra.Command, name string) error {
	// Build resource context and use unified handler
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return fmt.Errorf("failed to get resource context: %v", err)
	}

	// Check if plugin exists
	_, err = resource.Get(ctx, handlers.KindNvimPlugin, name)
	if err != nil {
		return fmt.Errorf("plugin not found: %s", name)
	}

	// Confirm deletion
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		fmt.Printf("Delete plugin definition '%s' from global library? (y/N): ", name)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			render.Info("Aborted")
			return nil
		}
	}

	// Delete plugin
	if err := resource.Delete(ctx, handlers.KindNvimPlugin, name); err != nil {
		return fmt.Errorf("failed to delete plugin: %v", err)
	}

	render.Success(fmt.Sprintf("Plugin '%s' removed from global library", name))
	return nil
}

func runDeleteWorkspacePlugins(cmd *cobra.Command, pluginNames []string) error {
	// Get workspace
	workspace, _, err := getWorkspaceForPlugins(cmd, deleteNvimAppFlag, deleteNvimWorkspaceFlag)
	if err != nil {
		return err
	}

	// Get datastore
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get workspace plugin manager
	mgr, err := NewWorkspacePluginManager()
	if err != nil {
		return err
	}

	// Remove plugins
	removed, notFound := mgr.RemovePlugins(workspace, pluginNames)

	// Save to database
	if err := ds.UpdateWorkspace(workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	// Report results
	if len(removed) > 0 {
		render.Success(fmt.Sprintf("Removed %d plugin(s) from workspace '%s':", len(removed), workspace.Name))
		for _, p := range removed {
			fmt.Printf("  - %s\n", p)
		}
	}

	if len(notFound) > 0 {
		render.Warning(fmt.Sprintf("Not found in workspace (%d):", len(notFound)))
		for _, p := range notFound {
			fmt.Printf("  ? %s\n", p)
		}
	}

	remaining := mgr.ListPlugins(workspace)
	if len(remaining) == 0 && len(removed) > 0 {
		render.Info("Workspace has no plugins configured")
		render.Info("Build will now use all plugins from global library")
	}

	if len(removed) > 0 {
		fmt.Println()
		render.Info(fmt.Sprintf("Rebuild workspace to apply: dvm build %s --force", workspace.Name))
	}

	return nil
}

// deleteNvimThemeCmd deletes a nvim theme (placeholder for future)
var deleteNvimThemeCmd = &cobra.Command{
	Use:   "theme [name]",
	Short: "Delete a nvim theme",
	Long: `Delete a nvim theme definition.

Note: Theme management is currently available via the standalone 'nvp' CLI.
This command will be integrated in a future version.

For now, use: nvp theme delete <name>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		themeName := args[0]
		render.Info("Theme management is currently available via the standalone 'nvp' CLI.")
		render.Info("")
		render.Info(fmt.Sprintf("Use this command instead:\n  nvp theme delete %s", themeName))
		render.Info("")
		render.Info("Integration with dvm is planned for a future release.")
		return nil
	},
}

// deleteWorkspaceCmd deletes a workspace
// Usage: dvm delete workspace <name>
var deleteWorkspaceCmd = &cobra.Command{
	Use:     "workspace [name]",
	Aliases: []string{"ws"},
	Short:   "Delete a workspace",
	Long: `Delete a workspace from an app.

This permanently removes the workspace from DVM's database.
It does NOT delete any container images or running containers.
By default, you will be prompted for confirmation.

Examples:
  dvm delete workspace dev                    # Delete from active app
  dvm delete ws dev                           # Short form
  dvm delete workspace dev --app myapp        # Delete from specific app
  dvm delete workspace dev --force            # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceName := args[0]

		// Get app from flag or context
		appFlag, _ := cmd.Flags().GetString("app")

		ctxMgr, err := operators.NewContextManager()
		if err != nil {
			return fmt.Errorf("failed to create context manager: %v", err)
		}

		var appName string
		if appFlag != "" {
			appName = appFlag
		} else {
			// Fall back to active app context
			appName, err = ctxMgr.GetActiveApp()
			if err != nil {
				return fmt.Errorf("no app specified. Use --app <name> or 'dvm use app <name>' first")
			}
		}

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			return fmt.Errorf("dataStore not initialized")
		}
		ds := *dataStore

		// Get app to get its ID (search globally across all domains)
		app, err := ds.GetAppByNameGlobal(appName)
		if err != nil {
			return fmt.Errorf("app '%s' not found: %v", appName, err)
		}

		// Check if workspace exists
		workspace, err := ds.GetWorkspaceByName(app.ID, workspaceName)
		if err != nil {
			return fmt.Errorf("workspace '%s' not found in app '%s'", workspaceName, appName)
		}

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Delete workspace '%s' from app '%s'? (y/N): ", workspaceName, appName)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				render.Info("Aborted")
				return nil
			}
		}

		// Delete the workspace
		if err := ds.DeleteWorkspace(workspace.ID); err != nil {
			return fmt.Errorf("failed to delete workspace: %v", err)
		}

		// Clear context if this was the active workspace in the active app
		activeApp, _ := ctxMgr.GetActiveApp()
		activeWorkspace, _ := ctxMgr.GetActiveWorkspace()
		if activeApp == appName && activeWorkspace == workspaceName {
			ctxMgr.ClearWorkspace()
			render.Info("Cleared active workspace context")
		}

		render.Success(fmt.Sprintf("Workspace '%s' deleted from app '%s'", workspaceName, appName))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Add nvim subcommand to delete
	deleteCmd.AddCommand(deleteNvimCmd)

	// Add plugin and theme under nvim
	deleteNvimCmd.AddCommand(deleteNvimPluginCmd)
	deleteNvimCmd.AddCommand(deleteNvimThemeCmd)

	// Add workspace command directly under delete
	deleteCmd.AddCommand(deleteWorkspaceCmd)

	// Add flags for nvim plugin
	deleteNvimPluginCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	deleteNvimPluginCmd.Flags().StringVarP(&deleteNvimWorkspaceFlag, "workspace", "w", "", "Remove from workspace (instead of global library)")
	deleteNvimPluginCmd.Flags().StringVarP(&deleteNvimAppFlag, "app", "a", "", "App for workspace (defaults to active)")

	// Add flags for workspace
	deleteWorkspaceCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	deleteWorkspaceCmd.Flags().StringP("app", "a", "", "App name (defaults to active app)")
}
