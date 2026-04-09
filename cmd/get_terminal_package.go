// Package cmd implements the 'dvm get terminal-package' command.
// It displays the resolved terminal package for the current workspace context,
// walking the hierarchy: workspace → app → domain → ecosystem → global.
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getTermPkgShowCascade bool

// getTerminalPackageCmd displays the resolved terminal package
var getTerminalPackageCmd = &cobra.Command{
	Use:   "terminal-package",
	Short: "Show resolved terminal package for current context",
	Long: `Display the effective terminal package for the current workspace context.

The package is resolved by walking the hierarchy:
  workspace → app → domain → ecosystem → global default

Use --show-cascade to see where each level's package is set.

Examples:
  dvm get terminal-package
  dvm get terminal-package --show-cascade
  dvm get terminal-package -o yaml`,
	RunE: runGetTerminalPackage,
}

func init() {
	getCmd.AddCommand(getTerminalPackageCmd)
	getTerminalPackageCmd.Flags().BoolVar(&getTermPkgShowCascade, "show-cascade", false, "Show full hierarchy walk")
}

func runGetTerminalPackage(cmd *cobra.Command, args []string) error {
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	ws, level, objectID, err := resolveCurrentWorkspaceForPackage(ctx)
	if err != nil {
		return err
	}

	resolution, err := resolvePackageCascade(ctx, "terminal", level, objectID)
	if err != nil {
		return fmt.Errorf("failed to resolve terminal package: %w", err)
	}

	return renderPackageResolution(resolution, ws, getTermPkgShowCascade, getOutputFormat)
}
