package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

// =============================================================================
// LOCK COMMAND
// =============================================================================

var lockCmd = &cobra.Command{
	Use:   "lock",
	Short: "Generate or verify a lazy-lock.json lock file",
	Long: `Manage the lazy-lock.json lock file for reproducible plugin versions.

Without flags, generates/updates the lock file from current plugin configuration.
With --verify, checks if the current config matches the existing lock file.

The lock file pins specific git commits for each plugin, ensuring
reproducible builds across environments.

Examples:
  nvp lock                 # Generate/update lazy-lock.json
  nvp lock --verify        # Check config matches lock file
  nvp lock --output /path  # Write lock file to custom location`,
	RunE: func(cmd *cobra.Command, args []string) error {
		verify, _ := cmd.Flags().GetBool("verify")
		if verify {
			return runLockVerify(cmd)
		}
		return runLockGenerate(cmd)
	},
}

func runLockGenerate(cmd *cobra.Command) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}
	defer mgr.Close()

	plugins, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	// Filter to enabled
	var enabled []*plugin.Plugin
	for _, p := range plugins {
		if p.Enabled {
			enabled = append(enabled, p)
		}
	}

	if len(enabled) == 0 {
		render.Info("No enabled plugins to lock")
		return nil
	}

	lf := plugin.GenerateLockFile(enabled)

	// Determine output path
	outputPath, _ := cmd.Flags().GetString("output")
	if outputPath == "" {
		outputPath = filepath.Join(getConfigDir(), "lazy-lock.json")
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := lf.WriteTo(outputPath); err != nil {
		return fmt.Errorf("failed to write lock file: %w", err)
	}

	render.Successf("Lock file written to %s (%d plugins)", outputPath, len(lf.Entries))
	return nil
}

func runLockVerify(cmd *cobra.Command) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}
	defer mgr.Close()

	// Determine lock file path
	lockPath, _ := cmd.Flags().GetString("output")
	if lockPath == "" {
		lockPath = filepath.Join(getConfigDir(), "lazy-lock.json")
	}

	// Parse existing lock file
	lf, err := plugin.ParseLockFile(lockPath)
	if err != nil {
		if os.IsNotExist(err) {
			render.Error("No lock file found at " + lockPath)
			render.Info("Run 'nvp lock' to generate one")
			return errSilent
		}
		return fmt.Errorf("failed to read lock file: %w", err)
	}

	// Get current plugins
	plugins, err := mgr.List()
	if err != nil {
		return fmt.Errorf("failed to list plugins: %w", err)
	}

	mismatches := lf.Verify(plugins)
	if len(mismatches) == 0 {
		render.Success("Lock file is up to date")
		return nil
	}

	render.Warningf("Lock file has %d mismatch(es):", len(mismatches))
	for _, m := range mismatches {
		render.Plainf("  %s", m.String())
	}
	render.Info("Run 'nvp lock' to update the lock file")
	return errSilent
}

func init() {
	lockCmd.Flags().Bool("verify", false, "Verify config matches lock file")
	lockCmd.Flags().String("output", "", "Lock file path (default: ~/.nvp/lazy-lock.json)")
}
