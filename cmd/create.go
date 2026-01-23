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
	fromCwd     bool
	projectPath string
	description string
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

		sqlDS, ok := (*dataStore).(*db.SQLDataStore)
		if !ok {
			fmt.Println("Error: Expected SQLDataStore")
			return
		}

		if err := sqlDS.CreateProject(project); err != nil {
			fmt.Printf("Error: Failed to create project: %v\n", err)
			return
		}

		// Get the created project to get its ID
		createdProject, err := sqlDS.GetProjectByName(projectName)
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

		if err := sqlDS.CreateWorkspace(workspace); err != nil {
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

// Initializes the 'create' command and links subcommands
func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.AddCommand(createProjectCmd)

	// Project creation flags
	createProjectCmd.Flags().BoolVar(&fromCwd, "from-cwd", false, "Use current working directory as project path")
	createProjectCmd.Flags().StringVar(&projectPath, "path", "", "Specific path for the project")
	createProjectCmd.Flags().StringVar(&description, "description", "", "Project description")
}
