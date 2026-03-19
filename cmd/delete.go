package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
)

// deleteCmd is the root 'delete' command for kubectl-style resource deletion
// Usage: dvm delete nvim plugin <name>
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources",
	Long: `Delete resources by name.

Resource aliases (kubectl-style):
  app       → a, application
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
			render.Plainf("  - %s", p)
		}
	}

	if len(notFound) > 0 {
		render.Warning(fmt.Sprintf("Not found in workspace (%d):", len(notFound)))
		for _, p := range notFound {
			render.Plainf("  ? %s", p)
		}
	}

	remaining := mgr.ListPlugins(workspace)
	if len(remaining) == 0 && len(removed) > 0 {
		render.Info("Workspace has no plugins configured")
		render.Info("Build will now use all plugins from global library")
	}

	if len(removed) > 0 {
		render.Blank()
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

		// Get datastore from context
		ds, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("DataStore not initialized: %w", err)
		}

		// Get app from flag or context
		appFlag, _ := cmd.Flags().GetString("app")

		var appName string
		if appFlag != "" {
			appName = appFlag
		} else {
			// Fall back to active app context (DB-backed)
			var err error
			appName, err = getActiveAppFromContext(ds)
			if err != nil {
				return fmt.Errorf("no app specified. Use --app <name> or 'dvm use app <name>' first")
			}
		}

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

		// Check if this is the active workspace before deleting
		// (ON DELETE SET NULL in the context table handles the FK, but we
		// want to inform the user that the active context was cleared)
		wasActive := false
		activeApp, _ := getActiveAppFromContext(ds)
		activeWorkspace, _ := getActiveWorkspaceFromContext(ds)
		if activeApp == appName && activeWorkspace == workspaceName {
			wasActive = true
		}

		// Delete the workspace
		if err := ds.DeleteWorkspace(workspace.ID); err != nil {
			return fmt.Errorf("failed to delete workspace: %v", err)
		}

		if wasActive {
			render.Info("Cleared active workspace context")
		}

		render.Success(fmt.Sprintf("Workspace '%s' deleted from app '%s'", workspaceName, appName))
		return nil
	},
}

// =============================================================================
// Credential Resource Commands (dvm delete credential <name>)
// =============================================================================

// deleteCredentialCmd deletes a credential
var deleteCredentialCmd = &cobra.Command{
	Use:     "credential <name>",
	Aliases: []string{"cred"},
	Short:   "Delete a credential",
	Long: `Delete a credential by name within a scope.

Requires exactly one scope flag to identify which credential to delete.
By default, you will be prompted for confirmation.

Examples:
  dvm delete credential github-token --app my-api
  dvm delete credential api-key --ecosystem prod --force
  dvm delete cred db-pass --domain backend -f`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		credName := args[0]

		// Get DataStore from context
		ds, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("dataStore not found in context")
		}

		// Resolve scope
		scopeType, scopeID, err := resolveCredentialScopeFromFlags(cmd, ds)
		if err != nil {
			return err
		}

		// Verify credential exists
		cred, err := ds.GetCredential(scopeType, scopeID, credName)
		if err != nil {
			return fmt.Errorf("credential '%s' not found in %s scope: %w", credName, scopeType, err)
		}

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Delete credential '%s' (scope: %s, source: %s)? (y/N): ", cred.Name, cred.ScopeType, cred.Source)
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(response)
			if response != "y" && response != "Y" {
				render.Info("Aborted")
				return nil
			}
		}

		// Delete the credential
		if err := ds.DeleteCredential(scopeType, scopeID, credName); err != nil {
			return fmt.Errorf("failed to delete credential: %w", err)
		}

		render.Success(fmt.Sprintf("Credential '%s' deleted (scope: %s)", credName, scopeType))
		return nil
	},
}

// =============================================================================
// Registry Resource Commands (dvm delete registry <name>)
// =============================================================================

// deleteRegistryCmd deletes a registry
var deleteRegistryCmd = &cobra.Command{
	Use:     "registry <name>",
	Aliases: []string{"reg"},
	Short:   "Delete a registry",
	Long: `Delete a package registry by name.

This permanently removes the registry from DVM's database.
It does NOT delete any container data or volumes associated with the registry.
By default, you will be prompted for confirmation.

Examples:
  dvm delete registry my-zot
  dvm delete reg my-zot            # Short form
  dvm delete registry my-zot -f    # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteRegistry(cmd, args[0])
	},
}

func deleteRegistry(cmd *cobra.Command, name string) error {
	// Get datastore from context
	ds, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("DataStore not initialized: %w", err)
	}

	// Check if registry exists (for confirmation prompt display)
	reg, err := ds.GetRegistryByName(name)
	if err != nil {
		return fmt.Errorf("registry '%s' not found", name)
	}

	// Confirm deletion
	force, _ := cmd.Flags().GetBool("force")
	if !force {
		fmt.Printf("Delete registry '%s' (type: %s)? (y/N): ", name, reg.Type)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			render.Info("Aborted")
			return nil
		}
	}

	// Use the core function with auto-stop logic
	factory := registry.NewServiceFactory()
	if err := deleteRegistryCore(cmd.Context(), ds, factory, name, force); err != nil {
		return err
	}

	return nil
}

// deleteRegistryCore is the testable core of registry deletion.
// It checks if the registry is running and stops it before deleting the DB record.
// If Stop() fails, the DB record is NOT deleted and an error is returned.
func deleteRegistryCore(ctx context.Context, ds db.RegistryStore, factory registry.ManagerFactory, name string, force bool) error {
	// Get registry from DB
	reg, err := ds.GetRegistryByName(name)
	if err != nil {
		return fmt.Errorf("registry '%s' not found", name)
	}

	// Create a service manager from the factory
	mgr, err := factory.CreateManager(reg)
	if err != nil {
		return fmt.Errorf("failed to create registry manager: %w", err)
	}

	// Check if running and auto-stop (Option C: always auto-stop)
	if mgr.IsRunning(ctx) {
		render.Progress(fmt.Sprintf("Stopping running registry '%s'...", name))
		if err := mgr.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop registry '%s': %w", name, err)
		}
		render.Success(fmt.Sprintf("Registry '%s' stopped", name))
	}

	// Delete the DB record
	if err := ds.DeleteRegistry(name); err != nil {
		return fmt.Errorf("failed to delete registry: %w", err)
	}

	render.Success(fmt.Sprintf("Registry '%s' deleted", name))
	return nil
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
	deleteNvimPluginCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	deleteNvimPluginCmd.Flags().StringVarP(&deleteNvimWorkspaceFlag, "workspace", "w", "", "Remove from workspace (instead of global library)")
	deleteNvimPluginCmd.Flags().StringVarP(&deleteNvimAppFlag, "app", "a", "", "App for workspace (defaults to active)")

	// Add flags for workspace
	deleteWorkspaceCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	deleteWorkspaceCmd.Flags().StringP("app", "a", "", "App name (defaults to active app)")

	// Registry command
	deleteCmd.AddCommand(deleteRegistryCmd)

	// Registry deletion flags
	deleteRegistryCmd.Flags().Bool("force", false, "Skip confirmation prompt")

	// Credential command
	deleteCmd.AddCommand(deleteCredentialCmd)

	// Credential deletion flags
	deleteCredentialCmd.Flags().Bool("force", false, "Skip confirmation prompt")
	addCredentialScopeFlags(deleteCredentialCmd)
}
