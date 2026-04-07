package main

import (
	"fmt"

	"github.com/rmkohlman/MaestroSDK/render"
	themelibrary "github.com/rmkohlman/MaestroTheme/library"

	"github.com/spf13/cobra"
)

// Theme library commands
var themeLibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Browse and import themes from the library",
}

var themeLibraryListCmd = &cobra.Command{
	Use:   "get",
	Short: "List available themes in the library",
	RunE: func(cmd *cobra.Command, args []string) error {
		themes, err := themelibrary.List()
		if err != nil {
			return fmt.Errorf("failed to list library themes: %w", err)
		}

		if len(themes) == 0 {
			render.Info("No themes in library")
			return nil
		}

		// Filter by category
		category, _ := cmd.Flags().GetString("category")
		if category != "" {
			themes, err = themelibrary.ListByCategory(category)
			if err != nil {
				return err
			}
		}

		format, _ := cmd.Flags().GetString("output")
		return outputThemeInfos(themes, format)
	},
}

var themeLibraryShowCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Show details of a library theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		t, err := themelibrary.Get(name)
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("output")
		return outputTheme(t, format)
	},
}

var themeLibraryInstallCmd = &cobra.Command{
	Use:   "import <name>...",
	Short: "Import themes from library",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		themeStore := getThemeStore()
		if err := themeStore.Init(); err != nil {
			return err
		}

		setActive, _ := cmd.Flags().GetBool("use")
		var lastInstalled string

		for _, name := range args {
			t, err := themelibrary.Get(name)
			if err != nil {
				render.WarningfToStderr("theme not found in library: %s", name)
				continue
			}

			if err := themeStore.Save(t); err != nil {
				render.WarningfToStderr("failed to install %s: %v", name, err)
				continue
			}

			render.Successf("Installed theme '%s'", t.Name)
			lastInstalled = t.Name
		}

		// Set active if requested
		if setActive && lastInstalled != "" {
			if err := themeStore.SetActive(lastInstalled); err != nil {
				return err
			}
			render.Successf("Active theme set to '%s'", lastInstalled)
		}

		return nil
	},
}

var themeLibraryCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List theme categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		categories, err := themelibrary.Categories()
		if err != nil {
			return err
		}

		render.Infof("Categories (%d):", len(categories))
		for _, c := range categories {
			themes, _ := themelibrary.ListByCategory(c)
			render.Plainf("  %-10s (%d themes)", c, len(themes))
		}
		return nil
	},
}
