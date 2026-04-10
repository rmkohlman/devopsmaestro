package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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

	// Build function wraps the single-workspace build phases for each workspace
	buildFn := func(ws *models.WorkspaceWithHierarchy) error {
		render.Info(fmt.Sprintf("Building: %s/%s", ws.App.Name, ws.Workspace.Name))
		return buildSingleWorkspaceForParallel(ds, ws)
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
	buildErr := buildWorkspacesInParallel(workspaces, buildConcurrency, buildFn, ds)

	// Read accurate succeeded/failed counts from the persisted build session.
	// The engine writes per-workspace results to the DB; querying here avoids
	// the broken placeholder counter that always returned 0 failures.
	succeeded, failed := getBuildSessionCounts(ds, len(workspaces), buildErr)
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

// getBuildSessionCounts retrieves succeeded/failed counts from the most
// recent persisted build session. Falls back to inferring from the total
// workspace count and whether an error was returned when the DB query fails.
func getBuildSessionCounts(ds db.DataStore, total int, buildErr error) (succeeded, failed int) {
	session, err := ds.GetLatestBuildSession()
	if err == nil && session != nil {
		return session.Succeeded, session.Failed
	}
	// Fallback: if no session available, use error presence as signal
	if buildErr != nil {
		return 0, total
	}
	return total, 0
}

// buildSingleWorkspaceForParallel executes the full build pipeline for a
// single workspace within the parallel build path. It mirrors the phase
// sequence from buildWorkspace() (build_orchestrator.go) but accepts a
// pre-resolved WorkspaceWithHierarchy instead of resolving from flags.
//
// On success, ws.Workspace.ImageName is updated to the built image tag
// (e.g., "dvm-dev-myapp:20260410-123456") so the engine can persist it.
func buildSingleWorkspaceForParallel(ds db.DataStore, ws *models.WorkspaceWithHierarchy) error {
	ctx := context.Background()
	if buildTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, buildTimeout)
		defer cancel()
	}

	bc := &buildContext{
		ds:            ds,
		ctx:           ctx,
		app:           ws.App,
		workspace:     ws.Workspace,
		appName:       ws.App.Name,
		workspaceName: ws.Workspace.Name,
	}

	// Phase 1: Validate app path
	if err := bc.validateAppPath(); err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}

	// Phase 2: Platform & registry
	if err := bc.detectBuildPlatform(); err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}
	if err := bc.prepareRegistry(); err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}

	// Phase 3: Dockerfile detection & workspace spec
	bc.checkDockerfile()
	if err := bc.prepareWorkspaceSpec(); err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}

	// Phase 4: Source, staging, language detection
	if err := bc.prepareSourceAndStaging(); err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}
	defer func() {
		if err := os.RemoveAll(bc.stagingDir); err != nil {
			slog.Warn("failed to clean up staging directory",
				"path", bc.stagingDir, "error", err)
		}
	}()

	// Phase 5: CA certs & nvim config
	if err := bc.resolveCACerts(); err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}
	if err := bc.generateNvimConfiguration(); err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}

	// Phase 6: Dockerfile generation & build
	if err := bc.generateDockerfileAndResolveArgs(); err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}

	skipped, err := bc.buildImage()
	if bc.builder != nil {
		defer bc.builder.Close()
	}
	if err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}
	if skipped {
		// Image already existed — still update the workspace record
		ws.Workspace.ImageName = bc.imageName
		return nil
	}

	// Phase 7: Post-build (DB update, registry push, summary)
	bc.postBuild()

	// Propagate built image tag back to the hierarchy struct so the
	// engine can persist it in the build session workspace entry.
	ws.Workspace.ImageName = bc.imageName

	return nil
}
