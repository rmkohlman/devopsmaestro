package cmd

import (
	"devopsmaestro/db"
	"fmt"
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
		database := ctx.Value("datsbase").(db.Database) // Pull the dataStore from ctx.Value
		// Initialize the database connection
		var err error

		if err != nil {
			fmt.Printf("Failed to initialize the database connection: %v\n", err)
			os.Exit(1)
		}

		// Optionally perform a backup or snapshot before migrating
		if backupFlag {
			fmt.Println("Performing a backup before applying migrations...")
			err := db.BackupDatabase(database)
			if err != nil {
				fmt.Printf("Failed to backup the database: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Backup completed successfully.")
		}

		if snapshotFlag {
			fmt.Println("Creating a snapshot before applying migrations...")
			err := db.SnapshotDatabase(database)
			if err != nil {
				fmt.Printf("Failed to create a snapshot of the database: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Snapshot completed successfully.")
		}

		// Run the necessary migrations to set up the database schema
		if err := db.InitializeDatabase(database); err != nil {
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
