package cmd

import (
	"devopsmaestro/db"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
)

var (
	backupFlag   bool
	snapshotFlag bool
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Apply database migrations",
	Long:  `This command applies the necessary database migrations to ensure your schema is up-to-date.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		database := ctx.Value("database").(*db.Database)
		if database == nil {
			fmt.Println("Error: Database not initialized")
			os.Exit(1)
		}

		// Get migrations filesystem from context
		migrationsFS := ctx.Value("migrationsFS").(fs.FS)
		if migrationsFS == nil {
			fmt.Println("Error: Migrations filesystem not available")
			os.Exit(1)
		}

		// Optionally perform a backup or snapshot before migrating
		if backupFlag {
			fmt.Println("Performing a backup before applying migrations...")
			err := db.BackupDatabase(*database)
			if err != nil {
				fmt.Printf("Failed to backup the database: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Backup completed successfully.")
		}

		if snapshotFlag {
			fmt.Println("Creating a snapshot before applying migrations...")
			err := db.SnapshotDatabase(*database)
			if err != nil {
				fmt.Printf("Failed to create a snapshot of the database: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Snapshot completed successfully.")
		}

		// Run the necessary migrations to set up the database schema
		if err := db.InitializeDatabase(*database, migrationsFS); err != nil {
			fmt.Printf("Failed to apply migrations: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Migrations applied successfully.")
	},
}

func init() {
	adminCmd.AddCommand(migrateCmd)
	migrateCmd.Flags().BoolVar(&backupFlag, "backup", false, "Backup the database before applying migrations")
	migrateCmd.Flags().BoolVar(&snapshotFlag, "snapshot", false, "Create a snapshot of the database before applying migrations")
}
