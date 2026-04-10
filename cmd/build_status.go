package cmd

import (
	"fmt"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/rmkohlman/MaestroSDK/render"
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
  dvm build status --session-id <uuid>
  dvm build status --history
  dvm build status -o json
  dvm build status -o yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runBuildStatus(cmd)
	},
}

var (
	buildStatusSessionID string
	buildStatusHistory   bool
)

func init() {
	AddOutputFlag(buildStatusCmd, "")
	buildStatusCmd.Flags().StringVar(&buildStatusSessionID, "session-id", "",
		"Show status for a specific build session ID")
	buildStatusCmd.Flags().BoolVar(&buildStatusHistory, "history", false,
		"Show recent build session history")
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
// status` is called but no build session is active. Includes a hint to start
// a build with --all or a scope flag.
func FormatNoActiveBuildSessionMessage() string {
	return "no active build session\nHint: Run 'dvm build --all' or 'dvm build -e <ecosystem>' to start a build session"
}

// FormatBuildDryRunStatusValue returns the string used in the STATUS column
// of the build status/dry-run table when a build is in dry-run mode.
func FormatBuildDryRunStatusValue() string {
	return "(dry-run)"
}

// runBuildStatus implements the `dvm build status` command logic.
// It queries the DataStore for build session data and displays it.
func runBuildStatus(cmd *cobra.Command) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// History mode: show recent sessions
	if buildStatusHistory {
		return showBuildHistory(ds)
	}

	// Specific session or latest
	if buildStatusSessionID != "" {
		return showBuildSession(ds, buildStatusSessionID)
	}

	return showLatestBuildSession(ds)
}

// showLatestBuildSession displays the most recent build session.
func showLatestBuildSession(ds db.DataStore) error {
	session, err := ds.GetLatestBuildSession()
	if err != nil {
		return fmt.Errorf("failed to query build sessions: %w", err)
	}
	if session == nil {
		return fmt.Errorf("%s", FormatNoActiveBuildSessionMessage())
	}
	return renderBuildSession(ds, session)
}

// showBuildSession displays a specific build session by ID.
func showBuildSession(ds db.DataStore, sessionID string) error {
	session, err := ds.GetBuildSession(sessionID)
	if err != nil {
		return fmt.Errorf("failed to get build session: %w", err)
	}
	return renderBuildSession(ds, session)
}

// showBuildHistory displays recent build sessions.
func showBuildHistory(ds db.DataStore) error {
	sessions, err := ds.GetBuildSessions(10)
	if err != nil {
		return fmt.Errorf("failed to query build sessions: %w", err)
	}
	if len(sessions) == 0 {
		return fmt.Errorf("%s", FormatNoActiveBuildSessionMessage())
	}

	render.Plain("Recent build sessions:")
	render.Plain("")
	for _, s := range sessions {
		duration := "in progress"
		if s.CompletedAt.Valid {
			duration = s.CompletedAt.Time.Sub(s.StartedAt).Round(time.Second).String()
		}
		render.Plain(fmt.Sprintf("  %s  %s  %s  %d/%d succeeded  %s",
			s.ID[:8], s.Status, s.StartedAt.Format("2006-01-02 15:04:05"),
			s.Succeeded, s.TotalWorkspaces, duration))
	}
	return nil
}

// renderBuildSession displays a single build session with workspace details.
func renderBuildSession(ds db.DataStore, session *models.BuildSession) error {
	duration := "in progress"
	if session.CompletedAt.Valid {
		duration = session.CompletedAt.Time.Sub(session.StartedAt).Round(time.Second).String()
	}

	render.Plain(fmt.Sprintf("Build Session: %s", session.ID))
	render.Plain(fmt.Sprintf("  Status:    %s", session.Status))
	render.Plain(fmt.Sprintf("  Started:   %s", session.StartedAt.Format("2006-01-02 15:04:05")))
	render.Plain(fmt.Sprintf("  Duration:  %s", duration))
	render.Plain(fmt.Sprintf("  Result:    %d succeeded, %d failed (%d total)",
		session.Succeeded, session.Failed, session.TotalWorkspaces))

	// Show workspace details
	entries, err := ds.GetBuildSessionWorkspaces(session.ID)
	if err != nil {
		render.Warning(fmt.Sprintf("Failed to query workspace details: %v", err))
		return nil
	}

	if len(entries) > 0 {
		render.Plain("")
		render.Plain("  Workspaces:")
		for _, e := range entries {
			wsDuration := "-"
			if e.DurationSeconds.Valid {
				wsDuration = fmt.Sprintf("%ds", e.DurationSeconds.Int64)
			}
			imageTag := "-"
			if e.ImageTag.Valid && e.ImageTag.String != "" {
				imageTag = e.ImageTag.String
			}
			errMsg := ""
			if e.ErrorMessage.Valid && e.ErrorMessage.String != "" {
				errMsg = fmt.Sprintf("  error: %s", e.ErrorMessage.String)
			}
			render.Plain(fmt.Sprintf("    [%s] workspace:%d  %s  %s%s",
				e.Status, e.WorkspaceID, imageTag, wsDuration, errMsg))
		}
	}

	return nil
}
