package cmd

import (
	"context"
	"database/sql"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// Legacy createProject function - kept for backwards compatibility
func createProject(ctx context.Context, cmd *cobra.Command, args []string) {
	dataStore := ctx.Value("dataStore").(db.DataStore) // Pull the dataStore from ctx.Value

	projectName := args[0]
	description, err := cmd.Flags().GetString("description")
	if err != nil {
		fmt.Printf("Error getting description flag: %v\n", err)
		return
	}
	project := models.Project{
		Name: projectName,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err = dataStore.CreateProject(&project)
	if err != nil {
		fmt.Printf("Error creating project: %v\n", err)
		return
	}
	fmt.Printf("Project '%s' created successfully.\n", project.Name)
}
