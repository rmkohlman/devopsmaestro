package main

import (
	"fmt"
	"log/slog"

	"devopsmaestro/pkg/source"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroSDK/resource"
	theme "github.com/rmkohlman/MaestroTheme"
	"github.com/spf13/cobra"
)

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage Neovim themes",
	Long: `Manage Neovim colorscheme themes using YAML definitions.

Themes define:
  - The colorscheme plugin to use (tokyonight, catppuccin, etc.)
  - Color palette that other plugins can reference
  - Custom color overrides

The active theme's palette is exported as a Lua module that other plugins
(lualine, bufferline, telescope, etc.) can use for consistent styling.

Examples:
  nvp theme library get               # See available themes
  nvp theme library import catppuccin-mocha
  nvp theme use catppuccin-mocha      # Set as active theme
  nvp theme get                       # Show active theme`,
}

var themeGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get theme(s)",
	Long: `Get Neovim themes.

With no arguments, lists all installed themes (active theme marked with *).
With a name argument, gets a specific theme definition.

Examples:
  nvp theme get                       # List all installed themes
  nvp theme get catppuccin-mocha      # Get specific theme
  nvp theme get catppuccin-mocha -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		themeStore := getThemeStore()

		if len(args) == 0 {
			// List mode
			themes, err := themeStore.List()
			if err != nil {
				return fmt.Errorf("failed to list themes: %w", err)
			}

			if len(themes) == 0 {
				render.Info("No themes installed")
				render.Info("Use 'nvp theme library get' to see available themes")
				return nil
			}

			// Get active theme
			active, _ := themeStore.GetActive()
			activeName := ""
			if active != nil {
				activeName = active.Name
			}

			format, _ := cmd.Flags().GetString("output")
			return outputThemes(themes, format, activeName)
		}

		// Single get mode
		var t *theme.Theme
		var err error

		t, err = themeStore.Get(args[0])
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("output")
		return outputTheme(t, format)
	},
}

var themeApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a theme from file or URL",
	Long: `Apply a theme definition from a YAML file or URL.

The -f flag accepts local files, URLs, or stdin (use '-' for stdin).
URLs starting with http://, https://, or github: are fetched automatically.

GitHub shorthand: github:user/repo/path/file.yaml
   
Examples:
  nvp theme apply -f my-theme.yaml
  nvp theme apply -f github:user/repo/themes/custom.yaml`,
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

			slog.Info("resource applied", "kind", res.GetKind(), "name", res.GetName(), "source", displayName)
			render.Successf("%s '%s' applied (from %s)", res.GetKind(), res.GetName(), displayName)
		}

		return nil
	},
}

var themeDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		themeStore := getThemeStore()

		// Check exists
		if _, err := themeStore.Get(name); err != nil {
			return fmt.Errorf("theme not found: %s", name)
		}

		// Confirm unless forced
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Delete theme '%s'? (y/N): ", name)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				render.Info("Aborted")
				return nil
			}
		}

		if err := themeStore.Delete(name); err != nil {
			return fmt.Errorf("failed to delete theme: %w", err)
		}

		render.Successf("Theme '%s' deleted", name)
		return nil
	},
}

var themeUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the active theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		themeStore := getThemeStore()

		if err := themeStore.SetActive(name); err != nil {
			return err
		}

		render.Successf("Active theme set to '%s'", name)
		render.Info("Run 'nvp generate' to regenerate Lua files with the new theme")
		return nil
	},
}

func init() {
	// Theme subcommands
	themeCmd.AddCommand(themeGetCmd)
	themeCmd.AddCommand(themeApplyCmd)
	themeCmd.AddCommand(themeCreateCmd)
	themeCmd.AddCommand(themeDeleteCmd)
	themeCmd.AddCommand(themeUseCmd)
	themeCmd.AddCommand(themeLibraryCmd)
	themeCmd.AddCommand(themeGenerateCmd)
	themeCmd.AddCommand(themePreviewCmd)

	// Theme library subcommands
	themeLibraryCmd.AddCommand(themeLibraryListCmd)
	themeLibraryCmd.AddCommand(themeLibraryShowCmd)
	themeLibraryCmd.AddCommand(themeLibraryInstallCmd)
	themeLibraryCmd.AddCommand(themeLibraryCategoriesCmd)

	// Flags
	themeGetCmd.Flags().StringP("output", "o", "yaml", "Output format: table, yaml, json")
	themeApplyCmd.Flags().StringSliceP("filename", "f", nil, "Theme YAML file(s) or URL(s) to apply")
	themeCreateCmd.Flags().String("from", "", "Base color (hex #rrggbb, hue 0-360, or preset name)")
	themeCreateCmd.Flags().String("name", "", "Theme name (required unless --dry-run)")
	themeCreateCmd.Flags().Bool("dry-run", false, "Preview without saving")
	themeCreateCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json, table")
	themeCreateCmd.Flags().Bool("use", false, "Set as active theme after creation")
	themeDeleteCmd.Flags().Bool("force", false, "Skip confirmation")
	themePreviewCmd.Flags().Bool("all", false, "Preview all library themes")
	themeLibraryListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	themeLibraryListCmd.Flags().StringP("category", "c", "", "Filter by category (dark, light)")
	themeLibraryShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	themeLibraryInstallCmd.Flags().Bool("use", false, "Set as active theme after install")
	themeGenerateCmd.Flags().String("output-dir", "", "Output directory (default: ~/.config/nvim/lua)")
	themeGenerateCmd.Flags().Bool("dry-run", false, "Show what would be generated")

	// Hidden backward-compat aliases for deprecated verbs in theme library
	// MUST be after flag definitions — shallow copy captures FlagSet pointer at copy time
	themeLibraryCmd.AddCommand(hiddenAlias("list", themeLibraryListCmd))
	themeLibraryCmd.AddCommand(hiddenAlias("show", themeLibraryShowCmd))
	themeLibraryCmd.AddCommand(hiddenAlias("install", themeLibraryInstallCmd))
}
