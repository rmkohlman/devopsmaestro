package cmd

import (
	"fmt"
	"strings"

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
	Long:  `Clear BuildKit build cache, npm/pip cache mounts, and build staging directory.`,
	RunE:  runCacheClear,
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

	// Dry-run: preview what would be cleared
	if cacheClearDryRun {
		targets := []string{}
		if all || buildkit {
			targets = append(targets, "BuildKit cache")
		}
		if all || npm {
			targets = append(targets, "npm cache")
		}
		if all || pip {
			targets = append(targets, "pip cache")
		}
		if all || staging {
			targets = append(targets, "build staging directory")
		}
		render.Plain(fmt.Sprintf("Would clear: %s", strings.Join(targets, ", ")))
		return nil
	}

	if !force && !all {
		fmt.Fprintln(cmd.OutOrStdout(), "Use --force to skip confirmation, or --all to clear everything")
	}

	// Placeholder: actual cache clearing will be implemented when
	// container runtime abstraction supports it
	if all || buildkit {
		fmt.Fprintln(cmd.OutOrStdout(), "Clearing BuildKit cache...")
	}
	if all || npm {
		fmt.Fprintln(cmd.OutOrStdout(), "Clearing npm cache...")
	}
	if all || pip {
		fmt.Fprintln(cmd.OutOrStdout(), "Clearing pip cache...")
	}
	if all || staging {
		fmt.Fprintln(cmd.OutOrStdout(), "Clearing build staging...")
	}

	return nil
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
