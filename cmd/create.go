package cmd

import (
	"database/sql"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/mirror"
	"devopsmaestro/render"
	"fmt"
	"path/filepath"

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
	workspaceRepo        string
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
  
  # Clone from a GitRepo mirror
  dvm create workspace feature-x --repo my-repo
  dvm create workspace feature-x --app myapp --repo my-repo
  
  # Create with description
  dvm create workspace feature-auth --description "Auth feature branch"
  
  # Create with custom image name
  dvm create workspace staging --image my-app:staging`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		workspaceName := args[0]

		// Validate name is not empty
		if err := ValidateResourceName(workspaceName, "workspace"); err != nil {
			return err
		}

		// Get app from flag or context
		appFlag, _ := cmd.Flags().GetString("app")
		repoFlag, _ := cmd.Flags().GetString("repo")

		contextMgr, err := operators.NewContextManager()
		if err != nil {
			return fmt.Errorf("failed to initialize context manager: %w", err)
		}

		var appName string
		if appFlag != "" {
			appName = appFlag
		} else {
			appName, err = contextMgr.GetActiveApp()
			if err != nil {
				render.Error("No app specified")
				render.Info("Hint: Use --app <name> or 'dvm use app <name>' to select an app first")
				return nil
			}
		}

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			return fmt.Errorf("DataStore not initialized")
		}

		ds := *dataStore

		// Get app to get its ID (search globally across all domains)
		app, err := ds.GetAppByNameGlobal(appName)
		if err != nil {
			render.Error(fmt.Sprintf("App '%s' not found: %v", appName, err))
			render.Info("Hint: List available apps with: dvm get apps --all")
			return nil
		}

		// Check if workspace already exists
		existingWorkspaces, err := ds.ListWorkspacesByApp(app.ID)
		if err == nil {
			for _, ws := range existingWorkspaces {
				if ws.Name == workspaceName {
					return fmt.Errorf("workspace '%s' already exists in app '%s'", workspaceName, appName)
				}
			}
		}

		// Determine image name
		// Use "pending" tag for new workspaces - actual tag set at build time
		imageName := workspaceImage
		if imageName == "" {
			imageName = fmt.Sprintf("dvm-%s-%s:pending", workspaceName, appName)
		}

		// If --repo flag is provided, look up the GitRepo
		var gitRepo *models.GitRepoDB
		if repoFlag != "" {
			gitRepo, err = ds.GetGitRepoByName(repoFlag)
			if err != nil {
				return fmt.Errorf("gitrepo '%s' not found", repoFlag)
			}
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

		// Set GitRepoID if --repo was provided
		if gitRepo != nil {
			workspace.GitRepoID = sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true}
		}

		if err := ds.CreateWorkspace(workspace); err != nil {
			return fmt.Errorf("failed to create workspace: %w", err)
		}

		// Clone from mirror if --repo was provided
		if gitRepo != nil {
			render.Progress(fmt.Sprintf("Cloning from mirror '%s'...", repoFlag))

			// Get workspace path and clone to repo/ subdirectory
			workspacePath, err := ds.GetWorkspacePath(workspace.ID)
			if err != nil {
				render.Warning(fmt.Sprintf("Failed to get workspace path: %v", err))
			} else {
				repoPath := filepath.Join(workspacePath, "repo")
				baseDir := getGitRepoBaseDir()
				mirrorMgr := mirror.NewGitMirrorManager(baseDir)

				// Check if mirror exists, sync if needed
				if !mirrorMgr.Exists(gitRepo.Slug) {
					render.Info("Mirror not yet cloned, syncing from remote...")
					if _, err := mirrorMgr.Clone(gitRepo.URL, gitRepo.Slug); err != nil {
						render.Error(fmt.Sprintf("Failed to sync mirror: %v", err))
						render.Info("Workspace created, but repository clone failed")
						render.Info(fmt.Sprintf("Try: dvm sync gitrepo %s", repoFlag))
						return nil
					}
				}

				// Clone from local mirror to workspace
				if err := mirrorMgr.CloneToWorkspace(gitRepo.Slug, repoPath, gitRepo.DefaultRef); err != nil {
					render.Error(fmt.Sprintf("Failed to clone to workspace: %v", err))
					render.Info("Workspace created, but repository clone failed")
					return nil
				}
				render.Success("Cloned repository to workspace")
			}
		}

		render.Success(fmt.Sprintf("Workspace '%s' created successfully", workspaceName))
		render.Info(fmt.Sprintf("App: %s", appName))
		if gitRepo != nil {
			render.Info(fmt.Sprintf("GitRepo: %s (cloned)", repoFlag))
		}
		render.Info(fmt.Sprintf("Image:   %s", imageName))

		fmt.Println()
		render.Info("Next steps:")
		render.Info("  1. Switch to this workspace:")
		render.Info(fmt.Sprintf("     dvm use workspace %s", workspaceName))
		render.Info("  2. Build and attach:")
		render.Info("     dvm build && dvm attach")
		return nil
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
	createWorkspaceCmd.Flags().StringVar(&workspaceRepo, "repo", "", "GitRepo to clone into workspace (see: dvm get gitrepos)")
}
