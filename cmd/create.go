package cmd

import (
	"database/sql"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// createCmd represents the base 'create' command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create resources",
	Long:  `Create various resources like projects, workspaces, dependencies, etc.`,
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
	Use:   "project <name>",
	Short: "Create a new project",
	Long: `Create a new project with the specified name.
	
Examples:
  # Create a project from the current directory
  dvm create project my-api --from-cwd
  
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
				fmt.Printf("Error: Failed to get current directory: %v\n", err)
				return
			}
		} else if projectPath != "" {
			path, err = filepath.Abs(projectPath)
			if err != nil {
				fmt.Printf("Error: Invalid path: %v\n", err)
				return
			}
		} else {
			fmt.Println("Error: Must specify either --from-cwd or --path")
			return
		}

		// Verify path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("Error: Path does not exist: %s\n", path)
			return
		}

		fmt.Printf("Creating project '%s' at %s...\n", projectName, path)

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			fmt.Println("Error: DataStore not initialized")
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
			fmt.Printf("Error: Failed to create project: %v\n", err)
			return
		}

		// Get the created project to get its ID
		createdProject, err := ds.GetProjectByName(projectName)
		if err != nil {
			fmt.Printf("Error: Failed to retrieve created project: %v\n", err)
			return
		}

		fmt.Printf("✓ Project '%s' created successfully (ID: %d)\n", projectName, createdProject.ID)

		// Auto-create "main" workspace
		fmt.Println("Creating default 'main' workspace...")
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
			fmt.Printf("Warning: Failed to create main workspace: %v\n", err)
			fmt.Println("You can create it manually later with: dvm create workspace main")
		} else {
			fmt.Println("✓ Default 'main' workspace created")
		}

		// Set project as active context
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			fmt.Printf("Warning: Failed to initialize context manager: %v\n", err)
		} else {
			if err := contextMgr.SetProject(projectName); err != nil {
				fmt.Printf("Warning: Failed to set active project: %v\n", err)
			} else {
				fmt.Printf("✓ Set '%s' as active project\n", projectName)
			}
		}

		fmt.Println("\nNext steps:")
		fmt.Println("  1. Use the main workspace:")
		fmt.Println("     dvm use workspace main")
		fmt.Println("  2. Start coding in your containerized environment:")
		fmt.Println("     dvm attach")
	},
}

// createWorkspaceCmd creates a new workspace in the current project
var createWorkspaceCmd = &cobra.Command{
	Use:   "workspace <name>",
	Short: "Create a new workspace",
	Long: `Create a new workspace in a project.

A workspace is an isolated development environment within a project.
You can have multiple workspaces per project (e.g., main, dev, feature-x).

Examples:
  # Create a workspace named 'dev' in active project
  dvm create workspace dev
  
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
			fmt.Printf("Error: Failed to initialize context manager: %v\n", err)
			return
		}

		var projectName string
		if projectFlag != "" {
			projectName = projectFlag
		} else {
			projectName, err = contextMgr.GetActiveProject()
			if err != nil {
				fmt.Println("Error: No project specified")
				fmt.Println("Hint: Use -p <project> or 'dvm use project <name>' to select a project first")
				return
			}
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
			fmt.Printf("Error: Project '%s' not found: %v\n", projectName, err)
			return
		}

		// Check if workspace already exists
		existingWorkspaces, err := ds.ListWorkspacesByProject(project.ID)
		if err == nil {
			for _, ws := range existingWorkspaces {
				if ws.Name == workspaceName {
					fmt.Printf("Error: Workspace '%s' already exists in project '%s'\n", workspaceName, projectName)
					return
				}
			}
		}

		// Determine image name
		imageName := workspaceImage
		if imageName == "" {
			imageName = fmt.Sprintf("dvm-%s-%s:latest", workspaceName, projectName)
		}

		fmt.Printf("Creating workspace '%s' in project '%s'...\n", workspaceName, projectName)

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
			fmt.Printf("Error: Failed to create workspace: %v\n", err)
			return
		}

		fmt.Printf("✓ Workspace '%s' created successfully\n", workspaceName)
		fmt.Printf("  Project: %s\n", projectName)
		fmt.Printf("  Image:   %s\n", imageName)

		fmt.Println("\nNext steps:")
		fmt.Println("  1. Switch to this workspace:")
		fmt.Printf("     dvm use workspace %s\n", workspaceName)
		fmt.Println("  2. Build and attach:")
		fmt.Println("     dvm build && dvm attach")
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
	createWorkspaceCmd.Flags().StringVar(&workspaceImage, "image", "", "Custom image name (default: dvm-<workspace>-<project>:latest)")
	createWorkspaceCmd.Flags().StringP("project", "p", "", "Project name (defaults to active project)")
}
