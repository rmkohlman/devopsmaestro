package cmd

import (
	"context"
	"devopsmaestro/db"
	"log"

	"github.com/spf13/cobra"
)

// snapshotCmd represents the snapshot command
var snapshotCmd = &cobra.Command{
	Use:   "snapshot",
	Short: "Create a snapshot of the database",
	Long: `The snapshot command creates a snapshot of the current database state.
This can be used to quickly capture the state of the database before making changes, 
allowing for easy rollback if needed.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		snapshotCommand(ctx)
	},
}

func snapshotCommand(ctx context.Context) {
	database := ctx.Value("database").(db.Database) // Pull the database from ctx.Value
	err := db.SnapshotDatabase(database)
	if err != nil {
		log.Fatalf("Failed to create database snapshot: %v", err)
	}

}

func init() {
	// Register the snapshot command under the admin command
	adminCmd.AddCommand(snapshotCmd)
}
