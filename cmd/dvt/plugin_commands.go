package main

import (
	"fmt"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTerminal/terminalops/plugin"
	pluginlibrary "github.com/rmkohlman/MaestroTerminal/terminalops/plugin/library"

	"github.com/spf13/cobra"
)

// =============================================================================
// PLUGIN COMMANDS
// =============================================================================

var pluginCmd = &cobra.Command{
	Use:     "plugin",
	Aliases: []string{"pl"},
	Short:   "Manage shell plugins (zsh-autosuggestions, etc.)",
	Long: `Manage shell plugin configurations.

Plugins enhance your shell with features like autosuggestions, syntax highlighting,
and fuzzy finding. The library provides pre-configured plugins that work well together.`,
}

var pluginLibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Browse and import plugins from the library",
}

var pluginLibraryListCmd = &cobra.Command{
	Use:   "get",
	Short: "List available plugins in the library",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := pluginlibrary.NewPluginLibrary()
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
			var filtered []*plugin.Plugin
			for _, p := range plugins {
				for _, t := range p.Tags {
					if t == tag {
						filtered = append(filtered, p)
						break
					}
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
	},
}

var pluginLibraryShowCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Show details of a library plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lib, err := pluginlibrary.NewPluginLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		p, err := lib.Get(name)
		if err != nil {
			return fmt.Errorf("plugin not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugin(p, format)
	},
}

var pluginLibraryInstallCmd = &cobra.Command{
	Use:   "import <name>...",
	Short: "Import plugins from library to local store",
	Long: `Copy plugin definitions from the built-in library to your local store.

Examples:
  dvt plugin library import zsh-autosuggestions
  dvt plugin library import zsh-autosuggestions zsh-syntax-highlighting
  dvt plugin library import --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if !all && len(args) == 0 {
			return fmt.Errorf("specify plugin names or use --all")
		}

		lib, err := pluginlibrary.NewPluginLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		store, err := getPluginStore(cmd)
		if err != nil {
			return err
		}

		var plugins []*plugin.Plugin
		if all {
			plugins = lib.List()
		} else {
			for _, name := range args {
				p, err := lib.Get(name)
				if err != nil {
					render.WarningfToStderr("plugin not found in library: %s", name)
					continue
				}
				plugins = append(plugins, p)
			}
		}

		for _, p := range plugins {
			if err := store.Upsert(p); err != nil {
				render.WarningfToStderr("failed to install %s: %v", p.Name, err)
				continue
			}
			render.Successf("Installed %s", p.Name)
		}

		return nil
	},
}

var pluginLibraryCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List plugin categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := pluginlibrary.NewPluginLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		categories := lib.Categories()
		render.Infof("Categories (%d):", len(categories))
		for _, c := range categories {
			plugins := lib.ListByCategory(c)
			render.Plainf("  %-20s (%d plugins)", c, len(plugins))
		}
		return nil
	},
}

var pluginGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get plugin definition(s)",
	Long: `Get shell plugin definitions.

With no arguments, lists all installed plugins.
With a name argument, gets a specific plugin definition.

Examples:
  dvt plugin get                        # List installed plugins
  dvt plugin get zsh-autosuggestions    # Get specific plugin`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// List mode
			store, err := getPluginStore(cmd)
			if err != nil {
				return err
			}
			plugins, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list plugins: %w", err)
			}

			if len(plugins) == 0 {
				render.Info("No plugins installed")
				render.Info("Use 'dvt plugin library get' to see available plugins")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			return outputPlugins(plugins, format)
		}
		// Single get mode
		name := args[0]
		store, err := getPluginStore(cmd)
		if err != nil {
			return err
		}

		p, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("plugin not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugin(p, format)
	},
}

var pluginGenerateCmd = &cobra.Command{
	Use:   "generate [name...]",
	Short: "Generate .zshrc plugin section (stdout)",
	Long: `Generate shell configuration for installing and loading plugins.

If no names specified, generates for all installed plugins.
Supports different plugin managers: zinit, oh-my-zsh, antigen, sheldon, manual.

Examples:
  dvt plugin generate                          # All installed plugins
  dvt plugin generate zsh-autosuggestions      # Specific plugin
  dvt plugin generate --manager zinit          # Use zinit format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getPluginStore(cmd)
		if err != nil {
			return err
		}
		_ = cmd.Flags().Lookup("manager") // manager flag not used with current API

		var plugins []*plugin.Plugin

		if len(args) > 0 {
			// Specific plugins
			for _, name := range args {
				p, err := store.Get(name)
				if err != nil {
					// Try library
					lib, _ := pluginlibrary.NewPluginLibrary()
					if libPlugin, libErr := lib.Get(name); libErr == nil {
						p = libPlugin
					} else {
						render.WarningfToStderr("plugin not found: %s", name)
						continue
					}
				}
				plugins = append(plugins, p)
			}
		} else {
			// All installed plugins
			plugins, err = store.List()
			if err != nil {
				return fmt.Errorf("failed to list plugins: %w", err)
			}
		}

		if len(plugins) == 0 {
			fmt.Println("# No plugins to generate")
			return nil
		}

		gen := plugin.NewZshGenerator("")
		output, err := gen.Generate(plugins)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		fmt.Print(output)
		return nil
	},
}

func init() {
	// Plugin subcommands
	pluginCmd.AddCommand(pluginLibraryCmd)
	pluginCmd.AddCommand(pluginGetCmd)
	pluginCmd.AddCommand(pluginGenerateCmd)

	// Plugin library subcommands
	pluginLibraryCmd.AddCommand(pluginLibraryListCmd)
	pluginLibraryCmd.AddCommand(pluginLibraryShowCmd)
	pluginLibraryCmd.AddCommand(pluginLibraryInstallCmd)
	pluginLibraryCmd.AddCommand(pluginLibraryCategoriesCmd)

	// Flags
	pluginLibraryListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	pluginLibraryListCmd.Flags().StringP("category", "c", "", "Filter by category")
	pluginLibraryListCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	pluginLibraryShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	pluginLibraryInstallCmd.Flags().Bool("all", false, "Import all plugins from library")
	pluginGetCmd.Flags().StringP("output", "o", "yaml", "Output format: table, yaml, json")
	pluginGenerateCmd.Flags().StringP("manager", "m", "manual", "Plugin manager: zinit, oh-my-zsh, antigen, sheldon, manual")

	// Hidden backward-compat aliases for deprecated verbs in plugin (after flags)
	pluginLibraryCmd.AddCommand(hiddenAlias("list", pluginLibraryListCmd))
	pluginLibraryCmd.AddCommand(hiddenAlias("show", pluginLibraryShowCmd))
	pluginLibraryCmd.AddCommand(hiddenAlias("install", pluginLibraryInstallCmd))
	pluginCmd.AddCommand(hiddenAlias("list", pluginGetCmd))
}
