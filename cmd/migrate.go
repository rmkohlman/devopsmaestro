package cmd

import (
	"devopsmaestro/db"
	"github.com/rmkohlman/MaestroSDK/render"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply database migrations",
	Long:  `This command applies the necessary database migrations to ensure your schema is up-to-date.`,
	Run: func(cmd *cobra.Command, args []string) {
		ds, dsErr := getDataStore(cmd)
		if dsErr != nil {
			render.Error("DataStore not initialized")
			os.Exit(1)
		}

		driver := ds.Driver()
		if driver == nil {
			render.Error("Database driver not available")
			os.Exit(1)
		}

		// Get migrations filesystem from context
		ctx := cmd.Context()
		migrationsFS := ctx.Value("migrationsFS").(fs.FS)
		if migrationsFS == nil {
			render.Error("Migrations filesystem not available")
			os.Exit(1)
		}

		// Run the necessary migrations to set up the database schema
		if err := db.RunMigrations(driver, migrationsFS); err != nil {
			render.Errorf("Failed to apply migrations: %v", err)
			os.Exit(1)
		}

		render.Success("Migrations applied successfully.")
	},
}

func init() {
	adminCmd.AddCommand(migrateCmd)
}
