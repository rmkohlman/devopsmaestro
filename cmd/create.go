package cmd

import (
	"database/sql"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/render"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// createCmd represents the base 'create' command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources",
	Long: `Create various resources like projects, workspaces, dependencies, etc.

Resource aliases (kubectl-style):
  project   → proj
  workspace → ws

Examples:
  dvm create project my-api --from-cwd
  dvm create proj my-api --from-cwd    # Short form
  dvm create workspace dev
  dvm create ws dev                    # Short form`,
}

var (
	fromCwd              bool
	projectPath          string
	description          string
	workspaceDescription string
	workspaceImage       string
)

// createProjectCmd creates a new project
var createProjectCmd = &cobra.Command{
	Use:     "project <name>",
	Aliases: []string{"proj"},
	Short:   "Create a new project",
	Long: `Create a new project with the specified name.
	
Examples:
  # Create a project from the current directory
  dvm create project my-api --from-cwd
  dvm create proj my-api --from-cwd    # Short form
  
  # Create a project with a specific path
  dvm create project my-api --path ~/code/my-api
  
  # Create with description
  dvm create project my-api --from-cwd --description "My REST API"`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]

		// Determine project path
		var path string
		var err error
		if fromCwd {
			path, err = os.Getwd()
			if err != nil {
				render.Error(fmt.Sprintf("Failed to get current directory: %v", err))
				return
			}
		} else if projectPath != "" {
			path, err = filepath.Abs(projectPath)
			if err != nil {
				render.Error(fmt.Sprintf("Invalid path: %v", err))
				return
			}
		} else {
			render.Error("Must specify either --from-cwd or --path")
			return
		}

		// Verify path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			render.Error(fmt.Sprintf("Path does not exist: %s", path))
			return
		}

		render.Progress(fmt.Sprintf("Creating project '%s' at %s...", projectName, path))

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			render.Error("DataStore not initialized")
			return
		}

		// Create project
		project := &models.Project{
			Name: projectName,
			Path: path,
			Description: sql.NullString{
				String: description,
				Valid:  description != "",
			},
		}

		ds := *dataStore

		if err := ds.CreateProject(project); err != nil {
			render.Error(fmt.Sprintf("Failed to create project: %v", err))
			return
		}

		// Get the created project to get its ID
		createdProject, err := ds.GetProjectByName(projectName)
		if err != nil {
			render.Error(fmt.Sprintf("Failed to retrieve created project: %v", err))
			return
		}

		render.Success(fmt.Sprintf("Project '%s' created successfully (ID: %d)", projectName, createdProject.ID))

		// Auto-create "main" workspace
		render.Progress("Creating default 'main' workspace...")
		workspace := &models.Workspace{
			ProjectID: createdProject.ID,
			Name:      "main",
			Description: sql.NullString{
				String: "Default workspace",
				Valid:  true,
			},
			ImageName: fmt.Sprintf("%s:latest", projectName),
			Status:    "stopped",
		}

		if err := ds.CreateWorkspace(workspace); err != nil {
			render.Warning(fmt.Sprintf("Failed to create main workspace: %v", err))
			render.Info("You can create it manually later with: dvm create workspace main")
		} else {
			render.Success("Default 'main' workspace created")
		}

		// Set project as active context
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			render.Warning(fmt.Sprintf("Failed to initialize context manager: %v", err))
		} else {
			if err := contextMgr.SetProject(projectName); err != nil {
				render.Warning(fmt.Sprintf("Failed to set active project: %v", err))
			} else {
				render.Success(fmt.Sprintf("Set '%s' as active project", projectName))
			}
		}

		fmt.Println()
		render.Info("Next steps:")
		render.Info("  1. Use the main workspace:")
		render.Info("     dvm use workspace main")
		render.Info("  2. Start coding in your containerized environment:")
		render.Info("     dvm attach")
	},
}

// createWorkspaceCmd creates a new workspace in the current project
var createWorkspaceCmd = &cobra.Command{
	Use:     "workspace <name>",
	Aliases: []string{"ws"},
	Short:   "Create a new workspace",
	Long: `Create a new workspace in a project.

A workspace is an isolated development environment within a project.
You can have multiple workspaces per project (e.g., main, dev, feature-x).

Examples:
  # Create a workspace named 'dev' in active project
  dvm create workspace dev
  dvm create ws dev                # Short form
  
  # Create a workspace in a specific project
  dvm create workspace dev -p myproject
  
  # Create with description
  dvm create workspace feature-auth --description "Auth feature branch"
  
  # Create with custom image name
  dvm create workspace staging --image my-project:staging`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		workspaceName := args[0]

		// Get project from flag or context
		projectFlag, _ := cmd.Flags().GetString("project")

		contextMgr, err := operators.NewContextManager()
		if err != nil {
			render.Error(fmt.Sprintf("Failed to initialize context manager: %v", err))
			return
		}

		var projectName string
		if projectFlag != "" {
			projectName = projectFlag
		} else {
			projectName, err = contextMgr.GetActiveProject()
			if err != nil {
				render.Error("No project specified")
				render.Info("Hint: Use -p <project> or 'dvm use project <name>' to select a project first")
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

		// Get project to get its ID
		project, err := ds.GetProjectByName(projectName)
		if err != nil {
			render.Error(fmt.Sprintf("Project '%s' not found: %v", projectName, err))
			return
		}

		// Check if workspace already exists
		existingWorkspaces, err := ds.ListWorkspacesByProject(project.ID)
		if err == nil {
			for _, ws := range existingWorkspaces {
				if ws.Name == workspaceName {
					render.Error(fmt.Sprintf("Workspace '%s' already exists in project '%s'", workspaceName, projectName))
					return
				}
			}
		}

		// Determine image name
		// Use "pending" tag for new workspaces - actual tag set at build time
		imageName := workspaceImage
		if imageName == "" {
			imageName = fmt.Sprintf("dvm-%s-%s:pending", workspaceName, projectName)
		}

		render.Progress(fmt.Sprintf("Creating workspace '%s' in project '%s'...", workspaceName, projectName))

		// Create workspace
		workspace := &models.Workspace{
			ProjectID: project.ID,
			Name:      workspaceName,
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
		render.Info(fmt.Sprintf("Project: %s", projectName))
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
	createCmd.AddCommand(createProjectCmd)
	createCmd.AddCommand(createWorkspaceCmd)

	// Project creation flags
	createProjectCmd.Flags().BoolVar(&fromCwd, "from-cwd", false, "Use current working directory as project path")
	createProjectCmd.Flags().StringVar(&projectPath, "path", "", "Specific path for the project")
	createProjectCmd.Flags().StringVar(&description, "description", "", "Project description")

	// Workspace creation flags
	createWorkspaceCmd.Flags().StringVar(&workspaceDescription, "description", "", "Workspace description")
	createWorkspaceCmd.Flags().StringVar(&workspaceImage, "image", "", "Custom image name (default: dvm-<workspace>-<project>:<timestamp>)")
	createWorkspaceCmd.Flags().StringP("project", "p", "", "Project name (defaults to active project)")
}
