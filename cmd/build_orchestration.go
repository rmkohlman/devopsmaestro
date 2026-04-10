package cmd

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// resolveWorkspacesForParallelBuild queries the DataStore for workspaces
// matching the given scope flags. When buildAll is true and no scope flags
// are set, it returns ALL workspaces. When scope flags are set, it filters
// accordingly. Returns an error if no workspaces match the given criteria.
func resolveWorkspacesForParallelBuild(ds db.DataStore, flags HierarchyFlags, buildAll bool) ([]*models.WorkspaceWithHierarchy, error) {
	filter := flags.ToFilter()

	// When --all with no scope flags, use empty filter to get everything.
	// When scope flags are set (with or without --all), use the filter.
	workspaces, err := ds.FindWorkspaces(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to query workspaces: %w", err)
	}

	if len(workspaces) == 0 {
		// Build a descriptive error based on which scope flag was set
		if flags.Ecosystem != "" {
			return nil, fmt.Errorf("%s", FormatBuildNoMatchScopeError("ecosystem", flags.Ecosystem))
		}
		if flags.Domain != "" {
			return nil, fmt.Errorf("%s", FormatBuildNoMatchScopeError("domain", flags.Domain))
		}
		if flags.App != "" {
			return nil, fmt.Errorf("%s", FormatBuildNoMatchScopeError("app", flags.App))
		}
		if flags.Workspace != "" {
			return nil, fmt.Errorf("%s", FormatBuildNoMatchScopeError("workspace", flags.Workspace))
		}
		// --all with no matches at all
		return nil, fmt.Errorf("no workspaces found")
	}

	return workspaces, nil
}

// shouldRouteToParallelBuild returns true when the build command should use
// the multi-workspace parallel build path instead of the single-workspace
// path. This is true when --all is set OR any scope flag is set.
func shouldRouteToParallelBuild(flags HierarchyFlags, buildAll bool) bool {
	return buildAll || flags.HasAnyFlag()
}

// runParallelBuild is the entry point for the multi-workspace parallel build
// path. It resolves workspaces from scope flags, then builds them in parallel.
func runParallelBuild(cmd *cobra.Command) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	allSet, _ := cmd.Flags().GetBool("all")
	workspaces, err := resolveWorkspacesForParallelBuild(ds, buildFlags, allSet)
	if err != nil {
		return err
	}

	// Dry-run: preview what would be built
	if buildDryRun {
		for _, ws := range workspaces {
			render.Plain(FormatBuildDryRunTable(ws.Workspace.Name, ws.App.Name))
		}
		render.Plain(FormatBuildDryRunSummary(len(workspaces)))
		return nil
	}

	// Determine scope label for the progress header
	scopeLabel, scopeValue := parallelBuildScopeLabel(buildFlags)
	render.Plain(FormatParallelBuildHeader(len(workspaces), scopeLabel, scopeValue, buildConcurrency))

	// Build function wraps the single-workspace build orchestrator
	buildFn := func(ws *models.WorkspaceWithHierarchy) error {
		render.Info(fmt.Sprintf("Building: %s/%s", ws.App.Name, ws.Workspace.Name))
		// TODO: wire into actual single-workspace build phases
		return nil
	}

	// Detach mode: launch in background, return session ID
	if buildDetach {
		sessionID, detachErr := buildWorkspacesInParallelDetached(
			workspaces, buildConcurrency, buildFn)
		if detachErr != nil {
			return detachErr
		}
		render.Plain(FormatBuildSessionID(sessionID))
		return nil
	}

	// Foreground mode: wait for all builds to complete
	buildErr := buildWorkspacesInParallel(workspaces, buildConcurrency, buildFn)

	succeeded := len(workspaces)
	failed := 0
	if buildErr != nil {
		// Count failures from the error message (rough — detailed tracking
		// will come with BuildSession from #205)
		failed = countFailedFromWorkspaces(workspaces, buildFn)
		succeeded = len(workspaces) - failed
	}
	render.Plain(FormatBuildSummaryLine(succeeded, failed, len(workspaces)))

	return buildErr
}

// parallelBuildScopeLabel returns a (label, value) pair for the progress header.
func parallelBuildScopeLabel(flags HierarchyFlags) (string, string) {
	if flags.Ecosystem != "" {
		return "ecosystem", flags.Ecosystem
	}
	if flags.Domain != "" {
		return "domain", flags.Domain
	}
	if flags.App != "" {
		return "app", flags.App
	}
	return "", ""
}

// countFailedFromWorkspaces is a placeholder for failure counting until
// BuildSession (#205) provides proper tracking. Returns 0 as a fallback.
func countFailedFromWorkspaces(_ []*models.WorkspaceWithHierarchy, _ func(*models.WorkspaceWithHierarchy) error) int {
	return 0
}
