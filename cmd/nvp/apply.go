package main

import (
	"fmt"
	"log/slog"

	"devopsmaestro/pkg/source"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"

	"github.com/spf13/cobra"
)

// =============================================================================
// APPLY COMMAND
// =============================================================================

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a plugin definition from file or URL",
	Long: `Apply a plugin definition from a YAML file or URL to the local store.
If the plugin already exists, it will be updated.

The -f flag accepts local files, URLs, or stdin (use '-' for stdin).
URLs starting with http://, https://, or github: are fetched automatically.

GitHub shorthand: github:user/repo/path/file.yaml
   
Examples:
  nvp apply -f telescope.yaml
  nvp apply -f plugin1.yaml -f plugin2.yaml
  nvp apply -f https://raw.githubusercontent.com/user/repo/main/plugin.yaml
  nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml
  cat plugin.yaml | nvp apply -f -`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file or URL with -f flag")
		}

		// Create resource context for file-based storage
		ctx := resource.Context{
			ConfigDir: getConfigDir(),
		}

		// Process files and URLs using unified source resolution
		for _, src := range files {
			srcObj := source.Resolve(src)
			data, displayName, err := srcObj.Read()
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", src, err)
			}

			// Use unified resource pipeline
			res, err := resource.Apply(ctx, data, displayName)
			if err != nil {
				return fmt.Errorf("failed to apply from %s: %w", displayName, err)
			}

			// Determine if this was a create or update based on the resource type
			// For now, just report success
			slog.Info("resource applied", "kind", res.GetKind(), "name", res.GetName(), "source", displayName)
			render.Successf("%s '%s' applied (from %s)", res.GetKind(), res.GetName(), displayName)
		}

		return nil
	},
}

func init() {
	applyCmd.Flags().StringSliceP("filename", "f", nil, "Plugin YAML file(s) or URL(s) to apply (use '-' for stdin)")
}
