package cmd

import (
	"database/sql"
	"fmt"
	"log/slog"
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
	// Show workspace details (fetched first so we can derive accurate status/counters)
	entries, err := ds.GetBuildSessionWorkspaces(session.ID)
	if err != nil {
		entries = nil // proceed with session-level data only
	}

	// Compute counters from actual workspace entries (source of truth).
	// The denormalized session.Succeeded/Failed fields may be stale if session
	// finalization failed or was interrupted (#366, #375).
	succeeded, failed := session.Succeeded, session.Failed
	if stats, statsF, statsErr := ds.GetBuildSessionStats(session.ID); statsErr == nil {
		succeeded, failed = stats, statsF
	}

	// Derive session status from workspace entries when the session record
	// is stale (e.g., still "running" after all workspaces completed).
	status := session.Status
	if status == "running" && len(entries) > 0 && allWorkspacesTerminal(entries) {
		if failed > 0 && succeeded == 0 {
			status = "failed"
		} else if failed > 0 {
			status = "partial"
		} else {
			status = "completed"
		}
		// Self-heal: persist the corrected session state so future queries
		// don't need to recompute. Errors are non-fatal.
		healSession(ds, session, status, succeeded, failed)
	}

	duration := "in progress"
	if session.CompletedAt.Valid {
		duration = session.CompletedAt.Time.Sub(session.StartedAt).Round(time.Second).String()
	} else if status != "running" {
		// Session finalization missed setting CompletedAt; estimate from entries
		duration = estimateDurationFromEntries(session.StartedAt, entries)
	}

	render.Plain(fmt.Sprintf("Build Session: %s", session.ID))
	render.Plain(fmt.Sprintf("  Status:    %s", status))
	render.Plain(fmt.Sprintf("  Started:   %s", session.StartedAt.Format("2006-01-02 15:04:05")))
	render.Plain(fmt.Sprintf("  Duration:  %s", duration))
	render.Plain(fmt.Sprintf("  Result:    %d succeeded, %d failed (%d total)",
		succeeded, failed, session.TotalWorkspaces))

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
			// Resolve workspace and app names from DB instead of showing raw ID
			wsLabel := resolveWorkspaceLabel(ds, e.WorkspaceID)
			render.Plain(fmt.Sprintf("    [%s] %s  %s  %s%s",
				e.Status, wsLabel, imageTag, wsDuration, errMsg))
		}
	}

	return nil
}

// resolveWorkspaceLabel looks up a workspace by ID and returns a human-readable
// "app/workspace" label. Falls back to "workspace:<id>" if lookup fails.
func resolveWorkspaceLabel(ds db.DataStore, wsID int) string {
	ws, err := ds.GetWorkspaceByID(wsID)
	if err != nil || ws == nil {
		return fmt.Sprintf("workspace:%d", wsID)
	}
	app, err := ds.GetAppByID(ws.AppID)
	if err != nil || app == nil {
		return ws.Name
	}
	return fmt.Sprintf("%s/%s", app.Name, ws.Name)
}

// allWorkspacesTerminal returns true if every workspace entry is in a terminal
// state (succeeded, failed, or skipped). "queued" and "building" are non-terminal.
func allWorkspacesTerminal(entries []*models.BuildSessionWorkspace) bool {
	for _, e := range entries {
		switch e.Status {
		case "succeeded", "failed", "skipped":
			continue
		default:
			return false
		}
	}
	return true
}

// healSession persists corrected session state back to the database so future
// queries return accurate data. Errors are logged but never propagated.
func healSession(ds db.DataStore, session *models.BuildSession, status string, succeeded, failed int) {
	completedAt := time.Now().UTC()
	healed := &models.BuildSession{
		ID:              session.ID,
		StartedAt:       session.StartedAt,
		CompletedAt:     sql.NullTime{Time: completedAt, Valid: true},
		Status:          status,
		TotalWorkspaces: session.TotalWorkspaces,
		Succeeded:       succeeded,
		Failed:          failed,
	}
	if err := ds.UpdateBuildSession(healed); err != nil {
		slog.Warn("failed to self-heal stale build session",
			"session_id", session.ID, "error", err)
	} else {
		slog.Info("self-healed stale build session",
			"session_id", session.ID, "status", status,
			"succeeded", succeeded, "failed", failed)
	}
}

// estimateDurationFromEntries derives a duration string from the latest
// workspace CompletedAt when the session itself has no CompletedAt.
func estimateDurationFromEntries(startedAt time.Time, entries []*models.BuildSessionWorkspace) string {
	var latest time.Time
	for _, e := range entries {
		if e.CompletedAt.Valid && e.CompletedAt.Time.After(latest) {
			latest = e.CompletedAt.Time
		}
	}
	if latest.IsZero() {
		return "unknown"
	}
	return latest.Sub(startedAt).Round(time.Second).String()
}
