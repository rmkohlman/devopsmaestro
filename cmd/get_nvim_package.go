// Package cmd implements the 'dvm get nvim-package' command.
// It displays the resolved nvim package for the current workspace context,
// walking the hierarchy: workspace → app → domain → ecosystem → global.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getNvimPkgShowCascade bool

// getNvimPackageCmd displays the resolved nvim package
var getNvimPackageCmd = &cobra.Command{
	Use:   "nvim-package",
	Short: "Show resolved nvim package for current context",
	Long: `Display the effective nvim plugin package for the current workspace context.

The package is resolved by walking the hierarchy:
  workspace → app → domain → ecosystem → global default

Use --show-cascade to see where each level's package is set.

Examples:
  dvm get nvim-package
  dvm get nvim-package --show-cascade
  dvm get nvim-package -o yaml`,
	RunE: runGetNvimPackage,
}

func init() {
	getCmd.AddCommand(getNvimPackageCmd)
	getNvimPackageCmd.Flags().BoolVar(&getNvimPkgShowCascade, "show-cascade", false, "Show full hierarchy walk")
}

func runGetNvimPackage(cmd *cobra.Command, args []string) error {
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	// Get current workspace from context
	ws, level, objectID, err := resolveCurrentWorkspaceForPackage(ctx)
	if err != nil {
		return err
	}

	resolution, err := resolvePackageCascade(ctx, "nvim", level, objectID)
	if err != nil {
		return fmt.Errorf("failed to resolve nvim package: %w", err)
	}

	return renderPackageResolution(resolution, ws, getNvimPkgShowCascade, getOutputFormat)
}
