package cmd

import (
	"context"
	"devopsmaestro/db"
	"log"

	"github.com/spf13/cobra"
)

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of the database",
	Long: `The backup command creates a backup of the current database state.
This is useful to safeguard data before performing potentially risky operations, 
such as migrations or schema changes.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		backupCommand(ctx)
	},
}

func backupCommand(ctx context.Context) {
	database := ctx.Value("database").(db.Database) // Pull the dataStore from ctx.Value
	err := db.BackupDatabase(database)
	if err != nil {
		log.Fatalf("Failed to backup database: %v", err)
	}
}

func init() {
	// Register the backup command under the admin command
	adminCmd.AddCommand(backupCmd)
}

func init() {
	// Register the backup command under the admin command
	adminCmd.AddCommand(backupCmd)
}
