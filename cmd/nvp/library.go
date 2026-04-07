package main

import (
	"fmt"
	"log/slog"

	"github.com/rmkohlman/MaestroNvim/nvimops/library"
	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

// =============================================================================
// LIBRARY COMMANDS
// =============================================================================

var libraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Browse and import from the plugin library",
	Long: `The plugin library contains curated, pre-configured plugin definitions
that work well together. Use these commands to explore and install plugins.`,
}

var libraryListCmd = &cobra.Command{
	Use:   "get",
	Short: "List all plugins in the library",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		plugins := lib.List()

		// Filter by category if specified
		category, _ := cmd.Flags().GetString("category")
		if category != "" {
			plugins = lib.ListByCategory(category)
		}

		// Filter by tag if specified
		tag, _ := cmd.Flags().GetString("tag")
		if tag != "" {
			plugins = lib.ListByTag(tag)
		}

		if len(plugins) == 0 {
			render.Info("No plugins found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugins(plugins, format)
	},
}

var libraryShowCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Show details of a library plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		p, ok := lib.Get(name)
		if !ok {
			return fmt.Errorf("plugin not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugin(p, format)
	},
}

var libraryInstallCmd = &cobra.Command{
	Use:   "import <name>...",
	Short: "Import plugins from library to local store",
	Long: `Copy plugin definitions from the built-in library to your local store.
You can then customize them with 'nvp get' and 'nvp apply'.

Examples:
  nvp library import telescope
  nvp library import telescope treesitter lspconfig
  nvp library import --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if !all && len(args) == 0 {
			return fmt.Errorf("specify plugin names or use --all")
		}

		slog.Debug("loading library")
		lib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		var plugins []*plugin.Plugin
		if all {
			plugins = lib.List()
			slog.Info("installing all plugins from library", "count", len(plugins))
		} else {
			for _, name := range args {
				p, ok := lib.Get(name)
				if !ok {
					slog.Warn("plugin not found in library", "name", name)
					render.WarningfToStderr("plugin not found in library: %s", name)
					continue
				}
				plugins = append(plugins, p)
			}
			slog.Info("installing plugins from library", "count", len(plugins), "names", args)
		}

		for _, p := range plugins {
			p.Enabled = true
			if err := mgr.Apply(p); err != nil {
				slog.Error("failed to install plugin", "name", p.Name, "error", err)
				render.WarningfToStderr("failed to install %s: %v", p.Name, err)
				continue
			}
			slog.Debug("installed plugin", "name", p.Name)
			render.Successf("Installed %s", p.Name)
		}

		return nil
	},
}

var libraryCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List all plugin categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		categories := lib.Categories()
		render.Infof("Categories (%d):", len(categories))
		for _, c := range categories {
			plugins := lib.ListByCategory(c)
			render.Plainf("  %-15s (%d plugins)", c, len(plugins))
		}
		return nil
	},
}

var libraryTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all plugin tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		tags := lib.Tags()
		render.Infof("Tags (%d):", len(tags))
		for _, t := range tags {
			plugins := lib.ListByTag(t)
			render.Plainf("  %-20s (%d plugins)", t, len(plugins))
		}
		return nil
	},
}

func init() {
	libraryCmd.AddCommand(libraryListCmd)
	libraryCmd.AddCommand(libraryShowCmd)
	libraryCmd.AddCommand(libraryInstallCmd)
	libraryCmd.AddCommand(libraryCategoriesCmd)
	libraryCmd.AddCommand(libraryTagsCmd)

	libraryListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	libraryListCmd.Flags().StringP("category", "c", "", "Filter by category")
	libraryListCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	libraryShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	libraryInstallCmd.Flags().Bool("all", false, "Import all plugins from library")

	// Hidden backward-compat aliases for deprecated verbs (list→get, show→describe, install→import)
	// MUST be after flag definitions — shallow copy captures FlagSet pointer at copy time
	libraryCmd.AddCommand(hiddenAlias("list", libraryListCmd))
	libraryCmd.AddCommand(hiddenAlias("show", libraryShowCmd))
	libraryCmd.AddCommand(hiddenAlias("install", libraryInstallCmd))
}
