package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

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

	// Shared mutex for serializing buffer flushes to stdout
	var outputMu sync.Mutex

	// Build function wraps the single-workspace build phases for each workspace.
	// Each workspace gets its own output buffer; when the build completes the
	// buffer is flushed atomically to stdout under a mutex so output from
	// concurrent builds never interleaves.
	buildFn := func(ws *models.WorkspaceWithHierarchy) error {
		var buf bytes.Buffer
		buf.WriteString(fmt.Sprintf("\n─── Building: %s/%s ───\n", ws.App.Name, ws.Workspace.Name))
		err := buildSingleWorkspaceForParallel(ds, ws, &buf)

		// Flush the entire workspace output atomically
		outputMu.Lock()
		_, _ = io.Copy(os.Stdout, &buf)
		outputMu.Unlock()

		return err
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

	// Extract accurate succeeded/failed counts directly from the BuildError
	// returned by the engine. This avoids a DB round-trip that could return
	// stale data from a concurrent or previous session.
	succeeded, failed := getBuildCounts(len(workspaces), buildErr)
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

// getBuildCounts extracts succeeded/failed counts from a BuildError returned
// by buildWorkspacesInParallel. This uses the in-memory counts from the engine
// directly, avoiding a database round-trip that could return stale data from a
// concurrent or previous build session.
func getBuildCounts(total int, buildErr error) (succeeded, failed int) {
	var be *BuildError
	if errors.As(buildErr, &be) {
		return be.Succeeded, be.Failed
	}
	// No BuildError: either all succeeded or the error is from another source
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
func buildSingleWorkspaceForParallel(ds db.DataStore, ws *models.WorkspaceWithHierarchy, out io.Writer) error {
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
		output:        out,
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

	// Phase 6b: Validate staging directory (warn on missing COPY sources)
	if err := bc.validateStagingDirectory(); err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}

	skipped, err := bc.buildImage()
	if bc.builder != nil {
		defer bc.builder.Close()
	}

	// Always propagate the image tag (even on failure) so the build session
	// records which tag was attempted. Without this, failed builds show stale
	// tags from previous sessions instead of the current attempt's tag.
	if bc.imageName != "" {
		ws.Workspace.ImageName = bc.imageName
	}

	if err != nil {
		return fmt.Errorf("%s/%s: %w", ws.App.Name, ws.Workspace.Name, err)
	}
	if skipped {
		return nil
	}

	// Phase 7: Post-build (DB update, registry push, summary)
	bc.postBuild()

	return nil
}
