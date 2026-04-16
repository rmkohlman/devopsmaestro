package cmd

import (
	"fmt"
	"log/slog"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

// preCleanCacheCleanup performs aggressive cleanup before building when
// --clean-cache is set (#383). This includes:
// 1. Deleting existing images for the target workspace (pre-build image replacement)
// 2. Pruning dangling images to reclaim space
// 3. Logging registry health status
func (bc *buildContext) preCleanCacheCleanup() {
	if bc.platform == nil {
		return
	}

	// Step 1: Delete existing images for this workspace
	bc.deleteExistingWorkspaceImages()

	// Step 2: Prune dangling images
	bc.pruneDanglingImages()

	// Step 3: Log registry cache status
	bc.logRegistryCacheStatus()
}

// deleteExistingWorkspaceImages removes all existing images for the
// current workspace before building. This ensures only the latest
// version exists after a successful build, preventing image accumulation.
func (bc *buildContext) deleteExistingWorkspaceImages() {
	if !bc.platform.IsDockerCompatible() {
		slog.Debug("skipping pre-build image cleanup: not Docker-compatible")
		return
	}

	repo := fmt.Sprintf("dvm-%s-%s", bc.workspaceName, bc.appName)
	bc.renderProgressf("Removing old images for %s...", repo)

	cli, err := client.NewClientWithOpts(
		client.WithHost(fmt.Sprintf("unix://%s", bc.platform.SocketPath)),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		slog.Warn("failed to create Docker client for pre-build cleanup", "error", err)
		return
	}
	defer cli.Close()

	filterArgs := filters.NewArgs()
	filterArgs.Add("reference", repo+"*")
	images, err := cli.ImageList(bc.ctx, image.ListOptions{Filters: filterArgs})
	if err != nil {
		slog.Warn("failed to list existing workspace images", "repo", repo, "error", err)
		return
	}

	removed := 0
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if imageRepo(tag) != repo {
				continue
			}
			slog.Info("removing old workspace image (pre-build)", "image", tag)
			_, removeErr := cli.ImageRemove(bc.ctx, tag, image.RemoveOptions{
				Force:         true,
				PruneChildren: true,
			})
			if removeErr != nil {
				slog.Warn("failed to remove old workspace image", "image", tag, "error", removeErr)
				continue
			}
			removed++
		}
	}

	if removed > 0 {
		bc.renderSuccessf("Removed %d old workspace image(s)", removed)
	} else {
		bc.renderInfo("No old workspace images to remove")
	}
}

// pruneDanglingImages removes dangling (untagged) images to free space.
func (bc *buildContext) pruneDanglingImages() {
	if !bc.platform.IsDockerCompatible() {
		return
	}

	cli, err := client.NewClientWithOpts(
		client.WithHost(fmt.Sprintf("unix://%s", bc.platform.SocketPath)),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		slog.Warn("failed to create Docker client for dangling prune", "error", err)
		return
	}
	defer cli.Close()

	bc.renderProgress("Pruning dangling images...")
	report, err := cli.ImagesPrune(bc.ctx, filters.NewArgs(
		filters.Arg("dangling", "true"),
	))
	if err != nil {
		slog.Warn("failed to prune dangling images", "error", err)
		return
	}

	if len(report.ImagesDeleted) > 0 || report.SpaceReclaimed > 0 {
		bc.renderSuccessf("Pruned %d dangling image(s), reclaimed %s",
			len(report.ImagesDeleted), formatBytes(int64(report.SpaceReclaimed)))
		slog.Info("pruned dangling images",
			"count", len(report.ImagesDeleted),
			"reclaimed_bytes", report.SpaceReclaimed)
	}
}

// logRegistryCacheStatus logs the registry cache readiness for visibility.
func (bc *buildContext) logRegistryCacheStatus() {
	if bc.cacheReadiness == nil {
		bc.renderWarning("No registry caches available — build will pull from upstream")
		return
	}
	if bc.cacheReadiness.AllHealthy {
		bc.renderSuccess(bc.cacheReadiness.FormatSummary())
	} else {
		bc.renderWarning(bc.cacheReadiness.FormatSummary())
	}
}

// postCleanCacheCleanup performs cleanup after a successful build when
// --clean-cache is set (#383). Prunes dangling images and BuildKit cache.
func (bc *buildContext) postCleanCacheCleanup() {
	if bc.platform == nil {
		return
	}

	bc.renderBlank()
	bc.renderProgress("Post-build cleanup (--clean-cache)...")

	// Prune dangling images created during build
	bc.pruneDanglingImages()

	// Prune BuildKit cache to minimize footprint
	bc.renderProgress("Pruning BuildKit cache after build...")
	if pruneErr := bc.pruneBuildKitCache(); pruneErr != nil {
		slog.Warn("post-build cache prune failed", "error", pruneErr)
		bc.renderWarning("Post-build cache prune failed (non-fatal)")
	} else {
		bc.renderSuccess("BuildKit cache pruned after build")
	}
}
