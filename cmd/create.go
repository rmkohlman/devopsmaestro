package cmd

import (
	"database/sql"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/render"
	"fmt"

	"github.com/spf13/cobra"
)

// createCmd represents the base 'create' command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources",
	Long: `Create various resources like apps, workspaces, dependencies, etc.

Resource aliases (kubectl-style):
  app       → a
  workspace → ws

Examples:
  dvm create app my-api --from-cwd
  dvm create a my-api --from-cwd       # Short form
  dvm create workspace dev
  dvm create ws dev                    # Short form`,
}

var (
	workspaceDescription string
	workspaceImage       string
)

// createWorkspaceCmd creates a new workspace in the current app
var createWorkspaceCmd = &cobra.Command{
	Use:     "workspace <name>",
	Aliases: []string{"ws"},
	Short:   "Create a new workspace",
	Long: `Create a new workspace in an app.

A workspace is an isolated development environment within an app.
You can have multiple workspaces per app (e.g., main, dev, feature-x).

Examples:
  # Create a workspace named 'dev' in active app
  dvm create workspace dev
  dvm create ws dev                # Short form
  
  # Create a workspace in a specific app
  dvm create workspace dev --app myapp
  
  # Create with description
  dvm create workspace feature-auth --description "Auth feature branch"
  
  # Create with custom image name
  dvm create workspace staging --image my-app:staging`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workspaceName := args[0]

		// Get app from flag or context
		appFlag, _ := cmd.Flags().GetString("app")

		contextMgr, err := operators.NewContextManager()
		if err != nil {
			render.Error(fmt.Sprintf("Failed to initialize context manager: %v", err))
			return
		}

		var appName string
		if appFlag != "" {
			appName = appFlag
		} else {
			appName, err = contextMgr.GetActiveApp()
			if err != nil {
				render.Error("No app specified")
				render.Info("Hint: Use --app <name> or 'dvm use app <name>' to select an app first")
				return
			}
		}

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			render.Error("DataStore not initialized")
			return
		}

		ds := *dataStore

		// Get app to get its ID (search globally across all domains)
		app, err := ds.GetAppByNameGlobal(appName)
		if err != nil {
			render.Error(fmt.Sprintf("App '%s' not found: %v", appName, err))
			render.Info("Hint: List available apps with: dvm get apps --all")
			return
		}

		// Check if workspace already exists
		existingWorkspaces, err := ds.ListWorkspacesByApp(app.ID)
		if err == nil {
			for _, ws := range existingWorkspaces {
				if ws.Name == workspaceName {
					render.Error(fmt.Sprintf("Workspace '%s' already exists in app '%s'", workspaceName, appName))
					return
				}
			}
		}

		// Determine image name
		// Use "pending" tag for new workspaces - actual tag set at build time
		imageName := workspaceImage
		if imageName == "" {
			imageName = fmt.Sprintf("dvm-%s-%s:pending", workspaceName, appName)
		}

		render.Progress(fmt.Sprintf("Creating workspace '%s' in app '%s'...", workspaceName, appName))

		// Create workspace
		workspace := &models.Workspace{
			AppID: app.ID,
			Name:  workspaceName,
			Description: sql.NullString{
				String: workspaceDescription,
				Valid:  workspaceDescription != "",
			},
			ImageName: imageName,
			Status:    "stopped",
		}

		if err := ds.CreateWorkspace(workspace); err != nil {
			render.Error(fmt.Sprintf("Failed to create workspace: %v", err))
			return
		}

		render.Success(fmt.Sprintf("Workspace '%s' created successfully", workspaceName))
		render.Info(fmt.Sprintf("App: %s", appName))
		render.Info(fmt.Sprintf("Image:   %s", imageName))

		fmt.Println()
		render.Info("Next steps:")
		render.Info("  1. Switch to this workspace:")
		render.Info(fmt.Sprintf("     dvm use workspace %s", workspaceName))
		render.Info("  2. Build and attach:")
		render.Info("     dvm build && dvm attach")
	},
}

// Initializes the 'create' command and links subcommands
func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createWorkspaceCmd)

	// Workspace creation flags
	createWorkspaceCmd.Flags().StringVar(&workspaceDescription, "description", "", "Workspace description")
	createWorkspaceCmd.Flags().StringVar(&workspaceImage, "image", "", "Custom image name (default: dvm-<workspace>-<app>:<timestamp>)")
	createWorkspaceCmd.Flags().StringP("app", "a", "", "App name (defaults to active app)")
}
