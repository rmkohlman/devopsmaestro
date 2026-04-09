package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// buildStatusCmd is the `dvm build status` subcommand that shows the current
// build session status when running hierarchical/parallel builds.
var buildStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show build session status",
	Long: `Show the status of the current or most recent build session.

Displays a table of all workspaces being built with their current status,
duration, and any errors encountered.

Examples:
  dvm build status
  dvm build status -o json
  dvm build status -o yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuildStatus(cmd)
	},
}

func init() {
	AddOutputFlag(buildStatusCmd, "")
}

// BuildStatusTableHeaders returns the ALL-CAPS column headers for the build
// status table, as required by the CLI architect review (Requirement R5).
func BuildStatusTableHeaders() []string {
	return []string{"WORKSPACE", "APP", "STATUS", "DURATION", "ERROR"}
}

// FormatBuildSummaryLine formats the build completion summary line.
// Example: "Build complete: 3 succeeded, 1 failed (4 total)"
func FormatBuildSummaryLine(succeeded, failed, total int) string {
	return fmt.Sprintf("Build complete: %d succeeded, %d failed (%d total)",
		succeeded, failed, total)
}

// FormatNoActiveBuildSessionMessage returns the message shown when `dvm build
// status` is called but no build session is active.
func FormatNoActiveBuildSessionMessage() string {
	return "no active build session"
}

// FormatBuildDryRunStatusValue returns the string used in the STATUS column
// of the build status/dry-run table when a build is in dry-run mode.
func FormatBuildDryRunStatusValue() string {
	return "(dry-run)"
}

// runBuildStatus implements the `dvm build status` command logic.
func runBuildStatus(cmd *cobra.Command) error {
	// TODO: When BuildSession from #205 is available, query active session
	// and display status table. For now, report no active session.
	return fmt.Errorf("%s", FormatNoActiveBuildSessionMessage())
}
