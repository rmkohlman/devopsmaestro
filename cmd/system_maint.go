package cmd

import (
	"github.com/spf13/cobra"
)

// systemMaintCmd is the top-level `dvm system` command for runtime maintenance.
// It groups system-level subcommands: info, df, prune.
var systemMaintCmd = &cobra.Command{
	Use:   "system",
	Short: "System maintenance and runtime management",
	Long: `Inspect and maintain your container runtime and dvm resources.

Subcommands:
  info    Show platform, runtime, and disk usage summary
  df      Show disk usage breakdown (Docker-style)
  prune   Clean up unused images and build caches

Examples:
  dvm system info
  dvm system df
  dvm system prune --dry-run
  dvm system prune --buildkit --force`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	rootCmd.AddCommand(systemMaintCmd)
}
