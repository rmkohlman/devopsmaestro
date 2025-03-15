package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var createProjectCmd = &cobra.Command{
	Use:   "create [project name]",
	Short: "Create a new project",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		createProject(ctx, cmd, args)
	},
}

func createProject(ctx context.Context, cmd *cobra.Command, args []string) {
	dataStore := ctx.Value("dataStore").(db.DataStore) // Pull the dataStore from ctx.Value

	projectName := args[0]
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		fmt.Printf("Error getting description flag: %v\n", err)
		return
	}
	project := models.Project{
		Name:        projectName,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = dataStore.CreateProject(&project)
	if err != nil {
		fmt.Printf("Error creating project: %v\n", err)
		return
	}
	fmt.Printf("Project '%s' created successfully.\n", project.Name)
}

func init() {
	createCmd.AddCommand(createProjectCmd)
	createProjectCmd.Flags().StringP("description", "d", "", "Description of the project")
}
