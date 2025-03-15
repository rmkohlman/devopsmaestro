package cmd

import (
	"github.com/spf13/cobra"
)

// adminCmd represents the base 'admin' command
var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "Administrative commands for managing the tool itself",
	Long:  `The 'admin' command provides various subcommands for managing the internal workings of the tool, such as migrations, backups, and snapshots.`,
}

func init() {
	rootCmd.AddCommand(adminCmd)
}
