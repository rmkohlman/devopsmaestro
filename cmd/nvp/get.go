package main

import (
	"fmt"

	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

// =============================================================================
// GET COMMAND
// =============================================================================

var getCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get plugin definition(s) from local store",
	Long: `Get plugins from the local store.

With no arguments, lists all plugins in the local store.
With a name argument, gets a specific plugin definition.

Examples:
  nvp get                    # List all plugins
  nvp get -c lsp             # List plugins filtered by category
  nvp get telescope          # Get specific plugin as YAML
  nvp get telescope -o json  # Get specific plugin as JSON`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// List mode
			mgr, err := getManager()
			if err != nil {
				return err
			}
			defer mgr.Close()

			plugins, err := mgr.List()
			if err != nil {
				return fmt.Errorf("failed to list plugins: %w", err)
			}

			// Filter by category
			category, _ := cmd.Flags().GetString("category")
			if category != "" {
				var filtered []*plugin.Plugin
				for _, p := range plugins {
					if p.Category == category {
						filtered = append(filtered, p)
					}
				}
				plugins = filtered
			}

			// Filter by enabled/disabled
			enabled, _ := cmd.Flags().GetBool("enabled")
			disabled, _ := cmd.Flags().GetBool("disabled")
			if enabled || disabled {
				var filtered []*plugin.Plugin
				for _, p := range plugins {
					if (enabled && p.Enabled) || (disabled && !p.Enabled) {
						filtered = append(filtered, p)
					}
				}
				plugins = filtered
			}

			if len(plugins) == 0 {
				render.Info("No plugins found")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			return outputPlugins(plugins, format)
		}
		// Single get mode
		name := args[0]

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		p, err := mgr.Get(name)
		if err != nil {
			return fmt.Errorf("plugin not found: %s", name)
		}

		// Show dependency tree if requested
		showDeps, _ := cmd.Flags().GetBool("show-deps")
		if showDeps {
			allPlugins, listErr := mgr.List()
			if listErr != nil {
				return fmt.Errorf("failed to list plugins for dependency resolution: %w", listErr)
			}
			resolver := plugin.NewDependencyResolver(allPlugins)
			tree := resolver.BuildTree(p.Repo)
			if tree != nil {
				fmt.Print(plugin.FormatTree(tree))
			} else {
				render.Infof("%s has no resolvable dependencies", name)
			}
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugin(p, format)
	},
}

func init() {
	getCmd.Flags().StringP("output", "o", "yaml", "Output format: table, yaml, json")
	getCmd.Flags().StringP("category", "c", "", "Filter by category")
	getCmd.Flags().Bool("enabled", false, "Show only enabled plugins")
	getCmd.Flags().Bool("disabled", false, "Show only disabled plugins")
	getCmd.Flags().Bool("show-deps", false, "Show dependency tree for a plugin")
}
