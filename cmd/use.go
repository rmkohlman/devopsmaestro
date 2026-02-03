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
	Long: `Switch the active project or workspace context (kubectl-style).

Use 'none' as the name to clear the context, or use --clear to clear all context.

Resource aliases (kubectl-style):
  project   → proj
  workspace → ws

Examples:
  dvm use project my-api        # Set active project
  dvm use proj my-api           # Short form
  dvm use workspace dev         # Set active workspace
  dvm use ws dev                # Short form
  dvm use project none          # Clear project context
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

			if err := contextMgr.ClearProject(); err != nil {
				return fmt.Errorf("failed to clear context: %v", err)
			}

			render.Success("Cleared all context (project and workspace)")
			return nil
		}

		// If no --clear flag and no subcommand, show help
		return cmd.Help()
	},
}

// useProjectCmd switches the active project
var useProjectCmd = &cobra.Command{
	Use:     "project <name>",
	Aliases: []string{"proj"},
	Short:   "Switch to a project",
	Long: `Set the specified project as the active context.

Use 'none' as the name to clear the project context (also clears workspace).

Examples:
  dvm use project my-api        # Set active project
  dvm use proj my-api           # Short form
  dvm use project frontend      # Switch to another project
  dvm use project none          # Clear project context`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectName := args[0]

		// Handle "none" to clear context
		if projectName == "none" {
			contextMgr, err := operators.NewContextManager()
			if err != nil {
				return fmt.Errorf("failed to initialize context manager: %v", err)
			}

			if err := contextMgr.ClearProject(); err != nil {
				return fmt.Errorf("failed to clear project context: %v", err)
			}

			// Also clear database context
			ctx := cmd.Context()
			dataStore := ctx.Value("dataStore").(*db.DataStore)
			if dataStore != nil {
				ds := *dataStore
				ds.SetActiveProject(nil)
				ds.SetActiveWorkspace(nil)
			}

			render.Success("Cleared project context (workspace also cleared)")
			return nil
		}

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			return fmt.Errorf("dataStore not initialized")
		}

		ds := *dataStore

		// Verify project exists
		project, err := ds.GetProjectByName(projectName)
		if err != nil {
			render.Error(fmt.Sprintf("Project '%s' not found: %v", projectName, err))
			render.Info("Hint: List available projects with: dvm get projects")
			return nil
		}

		// Set project as active in context manager
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			return fmt.Errorf("failed to initialize context manager: %v", err)
		}

		if err := contextMgr.SetProject(projectName); err != nil {
			return fmt.Errorf("failed to set active project: %v", err)
		}

		// Also update database context
		if err := ds.SetActiveProject(&project.ID); err != nil {
			render.Warning(fmt.Sprintf("Failed to update database context: %v", err))
		}

		render.Success(fmt.Sprintf("Switched to project '%s'", projectName))
		render.Info(fmt.Sprintf("Path: %s", project.Path))
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
Requires an active project to be set first (unless clearing with 'none').

Use 'none' as the name to clear the workspace context (keeps project).

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

		// Get active project
		projectName, err := contextMgr.GetActiveProject()
		if err != nil {
			render.Error("No active project set")
			render.Info("Hint: Set active project first with: dvm use project <name>")
			return nil
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
			return fmt.Errorf("failed to get project: %v", err)
		}

		// Verify workspace exists
		workspace, err := ds.GetWorkspaceByName(project.ID, workspaceName)
		if err != nil {
			render.Error(fmt.Sprintf("Workspace '%s' not found in project '%s': %v", workspaceName, projectName, err))
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

		render.Success(fmt.Sprintf("Switched to workspace '%s' in project '%s'", workspaceName, projectName))
		fmt.Println()
		render.Info("Next: Attach to your workspace with: dvm attach")
		return nil
	},
}

// Initializes the 'use' command and links subcommands
func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.AddCommand(useProjectCmd)
	useCmd.AddCommand(useWorkspaceCmd)
	useCmd.Flags().Bool("clear", false, "Clear all context (project and workspace)")
}
