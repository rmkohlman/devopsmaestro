package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/operators"
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
  project   → proj
  workspace → ws

Examples:
  dvm delete project my-api                   # Delete project and its workspaces
  dvm delete proj my-api                      # Short form
  dvm delete workspace dev                    # Delete workspace from active project
  dvm delete ws dev                           # Short form
  dvm delete workspace dev -p myproject       # Delete workspace from specific project
  dvm delete nvim plugin telescope            # Delete nvim plugin`,
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
	deleteNvimProjectFlag   string
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
	// Get nvim manager (uses DBStoreAdapter internally)
	mgr, err := getNvimManager(cmd)
	if err != nil {
		return fmt.Errorf("failed to get nvim manager: %v", err)
	}
	defer mgr.Close()

	// Check if plugin exists
	_, err = mgr.Get(name)
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
	if err := mgr.Delete(name); err != nil {
		return fmt.Errorf("failed to delete plugin: %v", err)
	}

	render.Success(fmt.Sprintf("Plugin '%s' removed from global library", name))
	return nil
}

func runDeleteWorkspacePlugins(cmd *cobra.Command, pluginNames []string) error {
	// Get workspace
	workspace, _, err := getWorkspaceForPlugins(cmd, deleteNvimProjectFlag, deleteNvimWorkspaceFlag)
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

// deleteProjectCmd deletes a project
// Usage: dvm delete project <name>
var deleteProjectCmd = &cobra.Command{
	Use:     "project [name]",
	Aliases: []string{"proj"},
	Short:   "Delete a project",
	Long: `Delete a project and optionally all its workspaces.

This permanently removes the project from DVM's database.
By default, you will be prompted for confirmation.

Examples:
  dvm delete project my-api
  dvm delete proj my-api           # Short form
  dvm delete project my-api --force  # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			return fmt.Errorf("dataStore not initialized")
		}
		ds := *dataStore

		// Check if project exists
		project, err := ds.GetProjectByName(projectName)
		if err != nil {
			return fmt.Errorf("project not found: %s", projectName)
		}

		// Check for workspaces
		workspaces, err := ds.ListWorkspacesByProject(project.ID)
		if err != nil {
			return fmt.Errorf("failed to list workspaces: %v", err)
		}

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			if len(workspaces) > 0 {
				render.Warning(fmt.Sprintf("Project '%s' has %d workspace(s) that will also be deleted:", projectName, len(workspaces)))
				for _, ws := range workspaces {
					render.Info(fmt.Sprintf("  - %s", ws.Name))
				}
				render.Info("")
			}
			fmt.Printf("Delete project '%s' and all its workspaces? (y/N): ", projectName)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				render.Info("Aborted")
				return nil
			}
		}

		// Delete all workspaces first
		for _, ws := range workspaces {
			if err := ds.DeleteWorkspace(ws.ID); err != nil {
				return fmt.Errorf("failed to delete workspace '%s': %v", ws.Name, err)
			}
		}

		// Delete the project
		if err := ds.DeleteProject(projectName); err != nil {
			return fmt.Errorf("failed to delete project: %v", err)
		}

		// Clear context if this was the active project
		ctxMgr, err := operators.NewContextManager()
		if err == nil {
			activeProject, _ := ctxMgr.GetActiveProject()
			if activeProject == projectName {
				ctxMgr.ClearProject()
				ctxMgr.ClearWorkspace()
				render.Info("Cleared active project context")
			}
		}

		msg := fmt.Sprintf("Project '%s' deleted", projectName)
		if len(workspaces) > 0 {
			msg += fmt.Sprintf(" (including %d workspace(s))", len(workspaces))
		}
		render.Success(msg)
		return nil
	},
}

// deleteWorkspaceCmd deletes a workspace
// Usage: dvm delete workspace <name>
var deleteWorkspaceCmd = &cobra.Command{
	Use:     "workspace [name]",
	Aliases: []string{"ws"},
	Short:   "Delete a workspace",
	Long: `Delete a workspace from a project.

This permanently removes the workspace from DVM's database.
It does NOT delete any container images or running containers.
By default, you will be prompted for confirmation.

Examples:
  dvm delete workspace dev                    # Delete from active project
  dvm delete ws dev                           # Short form
  dvm delete workspace dev -p myproject       # Delete from specific project
  dvm delete workspace dev --force            # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceName := args[0]

		// Get project from flag or context
		projectFlag, _ := cmd.Flags().GetString("project")

		ctxMgr, err := operators.NewContextManager()
		if err != nil {
			return fmt.Errorf("failed to create context manager: %v", err)
		}

		var projectName string
		if projectFlag != "" {
			// Use the -p flag value
			projectName = projectFlag
		} else {
			// Fall back to active project context
			projectName, err = ctxMgr.GetActiveProject()
			if err != nil {
				return fmt.Errorf("no project specified. Use -p <project> or 'dvm use project <name>' first")
			}
		}

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			return fmt.Errorf("dataStore not initialized")
		}
		ds := *dataStore

		// Get project to get its ID
		project, err := ds.GetProjectByName(projectName)
		if err != nil {
			return fmt.Errorf("project '%s' not found: %v", projectName, err)
		}

		// Check if workspace exists
		workspace, err := ds.GetWorkspaceByName(project.ID, workspaceName)
		if err != nil {
			return fmt.Errorf("workspace '%s' not found in project '%s'", workspaceName, projectName)
		}

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Delete workspace '%s' from project '%s'? (y/N): ", workspaceName, projectName)
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

		// Clear context if this was the active workspace in the active project
		activeProject, _ := ctxMgr.GetActiveProject()
		activeWorkspace, _ := ctxMgr.GetActiveWorkspace()
		if activeProject == projectName && activeWorkspace == workspaceName {
			ctxMgr.ClearWorkspace()
			render.Info("Cleared active workspace context")
		}

		render.Success(fmt.Sprintf("Workspace '%s' deleted from project '%s'", workspaceName, projectName))
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

	// Add project and workspace commands directly under delete
	deleteCmd.AddCommand(deleteProjectCmd)
	deleteCmd.AddCommand(deleteWorkspaceCmd)

	// Add flags for nvim plugin
	deleteNvimPluginCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	deleteNvimPluginCmd.Flags().StringVarP(&deleteNvimWorkspaceFlag, "workspace", "w", "", "Remove from workspace (instead of global library)")
	deleteNvimPluginCmd.Flags().StringVarP(&deleteNvimProjectFlag, "project", "p", "", "Project for workspace (defaults to active)")

	// Add flags for project
	deleteProjectCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	// Add flags for workspace
	deleteWorkspaceCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	deleteWorkspaceCmd.Flags().StringP("project", "p", "", "Project name (defaults to active project)")
}
