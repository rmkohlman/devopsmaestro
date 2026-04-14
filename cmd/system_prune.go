package cmd

import (
	"context"
	"devopsmaestro/operators"
	"fmt"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

var (
	pruneBuildKit bool
	pruneImages   bool
	pruneAll      bool
	pruneDryRun   bool
	pruneForce    bool
)

var systemPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Clean up unused images and build caches",
	Long: `Remove unused dvm workspace images and BuildKit build cache.

By default (no flags), performs a full cleanup (--all) with a confirmation
prompt. Use --dry-run to preview what would be deleted.

Safety: Images backing active workspace containers are NEVER deleted.

Flags:
  --buildkit   Prune BuildKit cache only (docker buildx prune)
  --images     Remove unused dvm workspace images only
  --all        Prune both BuildKit cache and images (default behavior)
  --dry-run    Preview what would be deleted without deleting
  --force      Skip confirmation prompt

Examples:
  dvm system prune                   # Full cleanup (with confirmation)
  dvm system prune --dry-run         # Preview cleanup
  dvm system prune --buildkit        # BuildKit cache only
  dvm system prune --images          # Unused dvm images only
  dvm system prune --all --force     # Full cleanup, no prompt`,
	RunE: runSystemPrune,
}

func init() {
	systemMaintCmd.AddCommand(systemPruneCmd)
	systemPruneCmd.Flags().BoolVar(&pruneBuildKit, "buildkit", false, "Prune BuildKit build cache")
	systemPruneCmd.Flags().BoolVar(&pruneImages, "images", false, "Remove unused dvm workspace images")
	systemPruneCmd.Flags().BoolVar(&pruneAll, "all", false, "Prune BuildKit cache and images")
	systemPruneCmd.Flags().BoolVar(&pruneDryRun, "dry-run", false, "Preview changes without applying")
	systemPruneCmd.Flags().BoolVar(&pruneForce, "force", false, "Skip confirmation prompt")
}

func runSystemPrune(cmd *cobra.Command, args []string) error {
	// Default to --all when no specific flag is set
	if !pruneBuildKit && !pruneImages && !pruneAll {
		pruneAll = true
	}
	if pruneAll {
		pruneBuildKit = true
		pruneImages = true
	}

	// Detect platform
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		return fmt.Errorf("failed to initialize platform detector: %w", err)
	}
	platform, err := detector.Detect()
	if err != nil {
		return fmt.Errorf("no container runtime found: %w", err)
	}
	if !platform.IsReachable() {
		render.Error("Container runtime is not running")
		render.Info(platform.GetStartHint())
		return errSilent
	}

	cleaner := operators.NewSystemCleaner(platform)
	ctx := context.Background()

	// Confirmation prompt (unless --force or --dry-run)
	if !pruneForce && !pruneDryRun {
		msg := "This will remove unused dvm resources"
		confirmed, confirmErr := confirmDelete(msg+"?", false)
		if confirmErr != nil {
			return confirmErr
		}
		if !confirmed {
			return nil
		}
	}

	var results []*operators.PruneResult

	// BuildKit prune
	if pruneBuildKit {
		render.Progress("Pruning BuildKit cache...")
		result, pruneErr := cleaner.PruneBuildKit(ctx, pruneDryRun)
		if pruneErr != nil {
			render.Warning(fmt.Sprintf("BuildKit prune: %v", pruneErr))
		} else {
			results = append(results, result)
		}
	}

	// Image prune
	if pruneImages {
		render.Progress("Checking dvm workspace images...")
		activeNames, listErr := getActiveContainerNames(platform)
		if listErr != nil {
			render.Warning(fmt.Sprintf("Could not list active containers: %v", listErr))
			activeNames = nil
		}

		result, pruneErr := cleaner.PruneImages(ctx, activeNames, pruneDryRun)
		if pruneErr != nil {
			render.Warning(fmt.Sprintf("Image prune: %v", pruneErr))
		} else {
			results = append(results, result)
		}
	}

	// Report results
	render.Blank()
	renderPruneResults(results, pruneDryRun)
	return nil
}

// getActiveContainerNames returns the names of containers currently running
// via the container runtime. These are used as a safety check — their backing
// images must not be deleted.
func getActiveContainerNames(platform *operators.Platform) ([]string, error) {
	runtime, err := operators.NewContainerRuntimeWith(
		&staticPlatformDetector{platform: platform},
	)
	if err != nil {
		return nil, err
	}
	workspaces, err := runtime.ListWorkspaces(context.Background())
	if err != nil {
		return nil, err
	}
	var names []string
	for _, ws := range workspaces {
		if isRunning(ws.Status) {
			names = append(names, ws.Image)
		}
	}
	return names, nil
}

// staticPlatformDetector always returns a fixed platform. Used to pass an
// already-detected Platform into NewContainerRuntimeWith.
type staticPlatformDetector struct {
	platform *operators.Platform
}

func (s *staticPlatformDetector) Detect() (*operators.Platform, error) {
	return s.platform, nil
}
func (s *staticPlatformDetector) DetectAll() []*operators.Platform {
	return []*operators.Platform{s.platform}
}
func (s *staticPlatformDetector) DetectReachable() []*operators.Platform {
	if s.platform.IsReachable() {
		return []*operators.Platform{s.platform}
	}
	return nil
}

func renderPruneResults(results []*operators.PruneResult, dryRun bool) {
	if dryRun {
		render.Info("Dry-run results (no changes made):")
	} else {
		render.Info("Prune results:")
	}

	for _, r := range results {
		if dryRun {
			render.Plain(fmt.Sprintf("  [%s] %s", r.Type, r.Details))
		} else {
			if r.ItemsRemoved > 0 {
				render.Success(fmt.Sprintf("  [%s] Removed %d item(s), freed %s",
					r.Type, r.ItemsRemoved, formatBytes(r.BytesFreed)))
			} else if r.Details != "" {
				render.Success(fmt.Sprintf("  [%s] %s", r.Type, r.Details))
			} else {
				render.Info(fmt.Sprintf("  [%s] Nothing to prune", r.Type))
			}
		}
	}
	render.Blank()
}
