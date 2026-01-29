package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/operators"
	"fmt"

	"github.com/spf13/cobra"
)

// useCmd represents the base 'use' command (kubectl-style context switching)
var useCmd = &cobra.Command{
	Use:   "use",
	Short: "Switch active context",
	Long:  `Switch the active project or workspace context (kubectl-style).`,
}

// useProjectCmd switches the active project
var useProjectCmd = &cobra.Command{
	Use:   "project <name>",
	Short: "Switch to a project",
	Long: `Set the specified project as the active context.

Examples:
  dvm use project my-api
  dvm use project frontend`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			fmt.Println("Error: DataStore not initialized")
			return
		}

		ds := *dataStore

		// Verify project exists
		project, err := ds.GetProjectByName(projectName)
		if err != nil {
			fmt.Printf("Error: Project '%s' not found: %v\n", projectName, err)
			fmt.Println("\nHint: List available projects with: dvm list projects")
			return
		}

		// Set project as active in context manager
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			fmt.Printf("Error: Failed to initialize context manager: %v\n", err)
			return
		}

		if err := contextMgr.SetProject(projectName); err != nil {
			fmt.Printf("Error: Failed to set active project: %v\n", err)
			return
		}

		// Also update database context
		if err := ds.SetActiveProject(&project.ID); err != nil {
			fmt.Printf("Warning: Failed to update database context: %v\n", err)
		}

		fmt.Printf("✓ Switched to project '%s'\n", projectName)
		fmt.Printf("  Path: %s\n", project.Path)

		fmt.Println("\nNext: Select a workspace with: dvm use workspace <name>")
	},
}

// useWorkspaceCmd switches the active workspace
var useWorkspaceCmd = &cobra.Command{
	Use:   "workspace <name>",
	Short: "Switch to a workspace",
	Long: `Set the specified workspace as the active context.
Requires an active project to be set first.

Examples:
  dvm use workspace main
  dvm use workspace dev`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workspaceName := args[0]

		// Get context manager
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			fmt.Printf("Error: Failed to initialize context manager: %v\n", err)
			return
		}

		// Get active project
		projectName, err := contextMgr.GetActiveProject()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("\nHint: Set active project first with: dvm use project <name>")
			return
		}

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			fmt.Println("Error: DataStore not initialized")
			return
		}

		ds := *dataStore

		// Get project to get its ID
		project, err := ds.GetProjectByName(projectName)
		if err != nil {
			fmt.Printf("Error: Failed to get project: %v\n", err)
			return
		}

		// Verify workspace exists
		workspace, err := ds.GetWorkspaceByName(project.ID, workspaceName)
		if err != nil {
			fmt.Printf("Error: Workspace '%s' not found in project '%s': %v\n", workspaceName, projectName, err)
			fmt.Println("\nHint: List available workspaces with: dvm list workspaces")
			return
		}

		// Set workspace as active in context manager
		if err := contextMgr.SetWorkspace(workspaceName); err != nil {
			fmt.Printf("Error: Failed to set active workspace: %v\n", err)
			return
		}

		// Also update database context
		if err := ds.SetActiveWorkspace(&workspace.ID); err != nil {
			fmt.Printf("Warning: Failed to update database context: %v\n", err)
		}

		fmt.Printf("✓ Switched to workspace '%s' in project '%s'\n", workspaceName, projectName)

		fmt.Println("\nNext: Attach to your workspace with: dvm attach")
	},
}

// Initializes the 'use' command and links subcommands
func init() {
	rootCmd.AddCommand(useCmd)
	useCmd.AddCommand(useProjectCmd)
	useCmd.AddCommand(useWorkspaceCmd)
}
