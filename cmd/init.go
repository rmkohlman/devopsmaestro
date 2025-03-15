package cmd

import (
	"devopsmaestro/db"
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the tool and set up the database",
	Long:  `This command initializes the tool by setting up the database schema and creating necessary configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		database := ctx.Value("datsbase").(db.Database) // Pull the dataStore from ctx.Value
		// Run the necessary migrations to set up the database schema
		if err := db.InitializeDatabase(database); err != nil {
			fmt.Printf("Failed to initialize the database schema: %v\n", err)
			return
		}

		fmt.Println("Initialization complete.")
	},
}

func init() {
	adminCmd.AddCommand(initCmd)
}
