package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// =============================================================================
// VERSION COMMAND
// =============================================================================

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		short, _ := cmd.Flags().GetBool("short")
		if short {
			fmt.Println(Version)
			return
		}
		fmt.Printf("nvp (NvimOps) %s\n", Version)
		fmt.Printf("  Build time: %s\n", BuildTime)
		fmt.Printf("  Commit:     %s\n", Commit)
	},
}

func init() {
	versionCmd.Flags().Bool("short", false, "Print only version number")
}

// =============================================================================
// INIT COMMAND
// =============================================================================

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize nvp configuration",
	Long: `Initialize the nvp configuration directory and plugin store.

This creates:
  ~/.nvp/
  ~/.nvp/plugins/     # Plugin YAML storage
  ~/.nvp/config.yaml  # Configuration file (optional)

You can specify a custom directory with --config or NVP_CONFIG_DIR.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getConfigDir()

		// Create directories
		pluginsDir := filepath.Join(dir, "plugins")
		if err := os.MkdirAll(pluginsDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		render.Successf("Initialized nvp at %s", dir)
		return nil
	},
}
