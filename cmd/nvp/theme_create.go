package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rmkohlman/MaestroSDK/render"
	theme "github.com/rmkohlman/MaestroTheme"
	"github.com/rmkohlman/MaestroTheme/parametric"
	"github.com/spf13/cobra"
)

var themeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom theme using parametric generation",
	Long: `Create a new theme using parametric generation from a base color.
	
This command uses the CoolNight Ocean theme as a reference and generates
a new theme by shifting the color hue while maintaining the same saturation
and lightness relationships. Semantic colors (red, green, yellow) are kept
fixed for consistency.

The --from flag accepts:
  - Hex colors: #FF5733, #3498db
  - Hue values: 270 (degrees 0-360)
  - Preset names: coolnight-synthwave, coolnight-matrix, coolnight-arctic

Examples:
  nvp theme create --from "#8B00FF" --name my-purple-theme
  nvp theme create --from "150" --name my-green-theme
  nvp theme create --from "coolnight-synthwave" --dry-run
  nvp theme create --from "#FF6B35" --name sunset-coding -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fromValue, _ := cmd.Flags().GetString("from")
		name, _ := cmd.Flags().GetString("name")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		output, _ := cmd.Flags().GetString("output")

		if fromValue == "" {
			return fmt.Errorf("--from flag is required (hex color, hue, or preset name)")
		}

		// Generate theme name if not provided and not dry run
		if name == "" && !dryRun {
			return fmt.Errorf("--name flag is required (or use --dry-run to preview)")
		}

		generator := parametric.NewGenerator()

		var generatedTheme *theme.Theme
		var err error

		// Check if it's a preset name first
		if preset, exists := parametric.GetPreset(fromValue); exists {
			_ = preset // Use preset to avoid unused variable
			generatedTheme, err = parametric.GeneratePreset(fromValue)
			if err != nil {
				return fmt.Errorf("failed to generate preset %s: %w", fromValue, err)
			}
		} else {
			// Try parsing as hex color
			if strings.HasPrefix(fromValue, "#") {
				description := fmt.Sprintf("Custom theme generated from %s", fromValue)
				if name != "" {
					description = fmt.Sprintf("Custom %s theme", name)
				}
				generatedTheme, err = generator.GenerateFromHex(fromValue, name, description)
				if err != nil {
					return fmt.Errorf("invalid hex color %s: %w", fromValue, err)
				}
			} else {
				// Try parsing as hue value
				hue, parseErr := strconv.ParseFloat(fromValue, 64)
				if parseErr != nil {
					return fmt.Errorf("invalid input %s: expected hex color (#rrggbb), hue (0-360), or preset name", fromValue)
				}

				if hue < 0 || hue >= 360 {
					return fmt.Errorf("hue must be between 0 and 360, got %.1f", hue)
				}

				description := fmt.Sprintf("Custom theme generated from hue %.1f°", hue)
				if name != "" {
					description = fmt.Sprintf("Custom %s theme (hue %.1f°)", name, hue)
				}
				generatedTheme = generator.GenerateFromHue(hue, name, description)
			}
		}

		// Dry run - just output the theme
		if dryRun {
			render.Info("Generated theme preview:")
			return outputTheme(generatedTheme, output)
		}

		// Save theme to store
		themeStore := getThemeStore()
		if err := themeStore.Init(); err != nil {
			return err
		}

		if err := themeStore.Save(generatedTheme); err != nil {
			return fmt.Errorf("failed to save theme: %w", err)
		}

		render.Successf("Created theme '%s'", generatedTheme.Name)

		// Optionally set as active
		setActive, _ := cmd.Flags().GetBool("use")
		if setActive {
			if err := themeStore.SetActive(generatedTheme.Name); err != nil {
				return err
			}
			render.Successf("Set '%s' as active theme", generatedTheme.Name)
		}

		// Show what to do next
		render.Blank()
		render.Info("Next steps:")
		if !setActive {
			render.Plainf("  nvp theme use %s      # Set as active theme", generatedTheme.Name)
		}
		render.Plain("  nvp generate          # Generate Lua files")
		render.Plainf("  nvp theme preview %s  # Preview colors", generatedTheme.Name)

		return nil
	},
}

var themeGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Lua files for the active theme",
	Long: `Generate Lua files for the active theme including:
  - theme/palette.lua   - Color palette module for other plugins
  - theme/init.lua      - Theme setup and helpers
  - plugins/colorscheme.lua - Lazy.nvim plugin spec

Other plugins can use the palette:
  local palette = require("theme").palette
  local bg = palette.colors.bg`,
	RunE: func(cmd *cobra.Command, args []string) error {
		themeStore := getThemeStore()

		t, err := themeStore.GetActive()
		if err != nil {
			return err
		}
		if t == nil {
			return fmt.Errorf("no active theme set. Use 'nvp theme use <name>' first")
		}

		outputDir, _ := cmd.Flags().GetString("output-dir")
		if outputDir == "" {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, ".config", "nvim", "lua")
		}

		// Expand ~
		if strings.HasPrefix(outputDir, "~") {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, outputDir[1:])
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		gen := theme.NewGenerator()
		generated, err := gen.Generate(t)
		if err != nil {
			return fmt.Errorf("failed to generate theme: %w", err)
		}

		files := map[string]string{
			filepath.Join(outputDir, "theme", "palette.lua"):              generated.PaletteLua,
			filepath.Join(outputDir, "theme", "init.lua"):                 generated.InitLua,
			filepath.Join(outputDir, "plugins", "nvp", "colorscheme.lua"): generated.PluginLua,
		}

		// Add standalone colorscheme file for standalone themes
		if t.IsStandalone() && generated.ColorschemeLua != "" {
			files[filepath.Join(outputDir, "theme", "colorscheme.lua")] = generated.ColorschemeLua
		}

		if dryRun {
			render.Infof("Would generate theme files for '%s':", t.Name)
			for path := range files {
				render.Plainf("  %s", path)
			}
			return nil
		}

		for path, content := range files {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", path, err)
			}
			if verbose {
				render.Plainf("  Generated %s", path)
			}
		}

		render.Successf("Generated theme '%s' to %s", t.Name, outputDir)
		render.Info("Other plugins can now use: require(\"theme\").palette")
		return nil
	},
}
