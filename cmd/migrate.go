package cmd

import (
	"devopsmaestro/db"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply database migrations",
	Long:  `This command applies the necessary database migrations to ensure your schema is up-to-date.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil || *dataStore == nil {
			fmt.Println("Error: DataStore not initialized")
			os.Exit(1)
		}

		driver := (*dataStore).Driver()
		if driver == nil {
			fmt.Println("Error: Database driver not available")
			os.Exit(1)
		}

		// Get migrations filesystem from context
		migrationsFS := ctx.Value("migrationsFS").(fs.FS)
		if migrationsFS == nil {
			fmt.Println("Error: Migrations filesystem not available")
			os.Exit(1)
		}

		// Run the necessary migrations to set up the database schema
		if err := db.RunMigrations(driver, migrationsFS); err != nil {
			fmt.Printf("Failed to apply migrations: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Migrations applied successfully.")
	},
}

func init() {
	adminCmd.AddCommand(migrateCmd)
}
