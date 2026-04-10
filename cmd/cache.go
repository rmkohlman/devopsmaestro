package cmd

import (
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// cacheCmd is the root 'cache' command
var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage build caches",
	Long:  `Manage BuildKit, npm, pip, and build staging caches.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// cacheClearCmd is the 'clear' subcommand
var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear build caches",
	Long: `Clear BuildKit build cache, local layer cache, and build staging directory.

Examples:
  dvm cache clear                # Clear all caches
  dvm cache clear --buildkit     # Clear Docker BuildKit cache only
  dvm cache clear --staging      # Clear build staging directory only
  dvm cache clear --dry-run      # Preview what would be cleared`,
	RunE: runCacheClear,
}

var cacheClearDryRun bool

func runCacheClear(cmd *cobra.Command, args []string) error {
	all, _ := cmd.Flags().GetBool("all")
	buildkit, _ := cmd.Flags().GetBool("buildkit")
	npm, _ := cmd.Flags().GetBool("npm")
	pip, _ := cmd.Flags().GetBool("pip")
	staging, _ := cmd.Flags().GetBool("staging")
	force, _ := cmd.Flags().GetBool("force")

	// Default to all if no specific flag is set
	if !buildkit && !npm && !pip && !staging {
		all = true
	}

	pc, err := paths.Default()
	if err != nil {
		return fmt.Errorf("failed to determine home directory: %w", err)
	}

	buildCacheDir := filepath.Join(pc.Root(), "build-cache")
	buildStagingDir := filepath.Join(pc.Root(), "build-staging")

	// Dry-run: preview what would be cleared
	if cacheClearDryRun {
		return previewCacheClear(all, buildkit, npm, pip, staging, buildCacheDir, buildStagingDir)
	}

	if !force && !all {
		render.Warning("Use --force to skip confirmation, or --all to clear everything")
		return nil
	}

	var totalFreed int64

	if all || buildkit {
		freed, clearErr := clearBuildKitCache(buildCacheDir)
		totalFreed += freed
		if clearErr != nil {
			slog.Warn("error during BuildKit cache clear", "error", clearErr)
		}
	}

	if all || npm {
		render.Info("npm cache clearing: not yet implemented (uses Docker volume mounts)")
	}

	if all || pip {
		render.Info("pip cache clearing: not yet implemented (uses Docker volume mounts)")
	}

	if all || staging {
		freed, clearErr := clearDirectory(buildStagingDir, "build staging")
		totalFreed += freed
		if clearErr != nil {
			slog.Warn("error during staging clear", "error", clearErr)
		}
	}

	render.Blank()
	render.Successf("Cache clear complete. Space freed: %s", formatBytes(totalFreed))
	return nil
}

// previewCacheClear shows what would be cleared without actually clearing.
func previewCacheClear(all, buildkit, npm, pip, staging bool, buildCacheDir, buildStagingDir string) error {
	targets := []string{}

	if all || buildkit {
		size := dirSize(buildCacheDir)
		targets = append(targets, fmt.Sprintf("local build cache (%s)", formatBytes(size)))
		targets = append(targets, "Docker BuildKit cache (docker buildx prune)")
	}
	if all || npm {
		targets = append(targets, "npm cache (Docker volume)")
	}
	if all || pip {
		targets = append(targets, "pip cache (Docker volume)")
	}
	if all || staging {
		size := dirSize(buildStagingDir)
		targets = append(targets, fmt.Sprintf("build staging directory (%s)", formatBytes(size)))
	}

	render.Plain(fmt.Sprintf("Would clear: %s", strings.Join(targets, ", ")))
	return nil
}

// clearBuildKitCache clears the local build-cache directory and runs docker buildx prune.
func clearBuildKitCache(buildCacheDir string) (int64, error) {
	var totalFreed int64

	// 1. Clear local type=local cache directory
	freed, err := clearDirectory(buildCacheDir, "local build cache")
	totalFreed += freed
	if err != nil {
		return totalFreed, err
	}

	// 2. Run docker buildx prune -f to clear Docker's internal BuildKit cache
	render.Progress("Pruning Docker BuildKit cache...")
	pruneCmd := exec.Command("docker", "buildx", "prune", "-f")
	output, pruneErr := pruneCmd.CombinedOutput()
	if pruneErr != nil {
		slog.Warn("docker buildx prune failed (non-fatal)", "error", pruneErr, "output", string(output))
		render.Warning("Docker BuildKit prune failed (Docker may not be running)")
	} else {
		render.Successf("Docker BuildKit cache pruned")
		slog.Debug("docker buildx prune output", "output", string(output))
	}

	return totalFreed, nil
}

// clearDirectory removes a directory and reports how much space was freed.
func clearDirectory(dir, label string) (int64, error) {
	size := dirSize(dir)

	if size == 0 {
		render.Info(fmt.Sprintf("No %s to clear (%s)", label, dir))
		return 0, nil
	}

	render.Progress(fmt.Sprintf("Clearing %s (%s)...", label, formatBytes(size)))

	if err := os.RemoveAll(dir); err != nil {
		return 0, fmt.Errorf("failed to clear %s at %s: %w", label, dir, err)
	}

	render.Successf("Cleared %s: %s freed", label, formatBytes(size))
	return size, nil
}

// dirSize calculates the total size of a directory tree. Returns 0 if the
// directory does not exist or cannot be read.
func dirSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries
		}
		if !d.IsDir() {
			info, infoErr := d.Info()
			if infoErr == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

func init() {
	cacheClearCmd.Flags().Bool("all", false, "Clear all caches")
	cacheClearCmd.Flags().Bool("buildkit", false, "Clear BuildKit build cache")
	cacheClearCmd.Flags().Bool("npm", false, "Clear npm cache mount")
	cacheClearCmd.Flags().Bool("pip", false, "Clear pip cache mount")
	cacheClearCmd.Flags().Bool("staging", false, "Clear build staging directory")
	AddForceConfirmFlag(cacheClearCmd)
	AddDryRunFlag(cacheClearCmd, &cacheClearDryRun)

	cacheCmd.AddCommand(cacheClearCmd)
	rootCmd.AddCommand(cacheCmd)
}
