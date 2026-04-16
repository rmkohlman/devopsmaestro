package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"

	"devopsmaestro/operators"
)

// pruneOldImages removes old container images for the same workspace after a
// successful build. It keeps only the newly built image (bc.imageName) and
// deletes any older images that share the same repository prefix
// (e.g. "dvm-dev-myapp"). This is called automatically from postBuild.
func (bc *buildContext) pruneOldImages() {
	if bc.platform == nil || bc.imageName == "" {
		return
	}

	// Extract the repository name (without tag) from the new image.
	// Image names follow the pattern "dvm-<workspace>-<app>:<timestamp>".
	repo := imageRepo(bc.imageName)
	if repo == "" {
		return
	}

	pruned, err := pruneImagesForRepo(bc.ctx, bc.platform, repo, bc.imageName)
	if err != nil {
		slog.Warn("failed to prune old images", "repo", repo, "error", err)
		return
	}
	if len(pruned) > 0 {
		bc.renderInfof("Pruned %d old image(s):", len(pruned))
		for _, name := range pruned {
			bc.renderInfof("  - %s", name)
		}
		slog.Info("pruned old images", "count", len(pruned), "repo", repo)
	}
}

// imageRepo extracts the repository portion from an image reference.
// "dvm-dev-myapp:20260415-234218" → "dvm-dev-myapp"
func imageRepo(imageName string) string {
	if idx := strings.LastIndex(imageName, ":"); idx > 0 {
		return imageName[:idx]
	}
	return imageName
}

// pruneImagesForRepo lists all images matching the given repository prefix
// and removes any that are not the keepImage. Returns the names of removed images.
func pruneImagesForRepo(
	ctx context.Context,
	platform *operators.Platform,
	repo string,
	keepImage string,
) ([]string, error) {
	if !platform.IsDockerCompatible() {
		slog.Debug("image pruning skipped: platform is not Docker-compatible",
			"platform", platform.Name)
		return nil, nil
	}

	cli, err := client.NewClientWithOpts(
		client.WithHost(fmt.Sprintf("unix://%s", platform.SocketPath)),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client for pruning: %w", err)
	}
	defer cli.Close()

	// List images filtered by reference to the repo name.
	filterArgs := filters.NewArgs()
	filterArgs.Add("reference", repo+"*")
	images, err := cli.ImageList(ctx, image.ListOptions{Filters: filterArgs})
	if err != nil {
		return nil, fmt.Errorf("failed to list images for %s: %w", repo, err)
	}

	var pruned []string
	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == keepImage {
				continue
			}
			// Only prune images with the exact same repo name.
			if imageRepo(tag) != repo {
				continue
			}
			slog.Info("pruning old image", "image", tag)
			_, removeErr := cli.ImageRemove(ctx, tag, image.RemoveOptions{
				Force:         false,
				PruneChildren: true,
			})
			if removeErr != nil {
				slog.Warn("failed to remove old image",
					"image", tag, "error", removeErr)
				continue
			}
			pruned = append(pruned, tag)
		}
	}
	return pruned, nil
}
