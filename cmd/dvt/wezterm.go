package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmkohlman/MaestroTerminal/terminalops/wezterm"
	weztermlib "github.com/rmkohlman/MaestroTerminal/terminalops/wezterm/library"
	"github.com/rmkohlman/MaestroPalette"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTheme/library"

	"github.com/spf13/cobra"
)

// =============================================================================
// WEZTERM COMMANDS
// =============================================================================

var weztermCmd = &cobra.Command{
	Use:   "wezterm",
	Short: "Manage WezTerm terminal configuration",
	Long:  `Manage WezTerm terminal configurations using YAML presets.`,
}

var weztermListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available WezTerm presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := weztermlib.NewWeztermLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		configs := lib.List()
		render.Info("Available WezTerm presets:")
		render.Blank()
		for _, c := range configs {
			render.Plainf("  %s - %s", c.Name, c.Description)
		}
		return nil
	},
}

var weztermShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show details of a WezTerm preset",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lib, err := weztermlib.NewWeztermLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		config, err := lib.Get(name)
		if err != nil {
			return fmt.Errorf("preset not found: %s", name)
		}

		render.Infof("Preset: %s", config.Name)
		render.Plainf("Description: %s", config.Description)
		if config.Font.Family != "" {
			render.Plainf("Font: %s (%.0fpt)", config.Font.Family, config.Font.Size)
		}
		if config.Window.Opacity > 0 && config.Window.Opacity < 1.0 {
			render.Plainf("Opacity: %.1f", config.Window.Opacity)
		}
		if config.ThemeRef != "" {
			render.Plainf("Theme reference: %s", config.ThemeRef)
		}
		if len(config.Keys) > 0 {
			render.Plainf("Key bindings: %d configured", len(config.Keys))
		}
		return nil
	},
}

var weztermGenerateCmd = &cobra.Command{
	Use:   "generate <name>",
	Short: "Generate WezTerm configuration to stdout",
	Long: `Generate WezTerm configuration and output to stdout.

Examples:
  dvt wezterm generate default
  dvt wezterm generate default > custom.wezterm.lua`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lib, err := weztermlib.NewWeztermLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		config, err := lib.Get(name)
		if err != nil {
			return fmt.Errorf("preset not found: %s", name)
		}

		// Resolve theme colors if ThemeRef is set
		if err := resolveThemeColors(config); err != nil {
			// Log warning but continue - colors just won't be embedded
			render.WarningfToStderr("Could not resolve theme colors: %v", err)
		}

		generator := wezterm.NewLuaGenerator()
		lua, err := generator.GenerateFromConfig(config)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		fmt.Print(lua)
		return nil
	},
}

var weztermApplyCmd = &cobra.Command{
	Use:   "apply <name>",
	Short: "Apply WezTerm configuration",
	Long: `Applies WezTerm configuration by writing to ~/.wezterm.lua.

Examples:
  dvt wezterm apply default
  dvt wezterm apply default --output ~/custom.wezterm.lua`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		output, _ := cmd.Flags().GetString("output")

		lib, err := weztermlib.NewWeztermLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		config, err := lib.Get(name)
		if err != nil {
			return fmt.Errorf("preset not found: %s", name)
		}

		// Resolve theme colors if ThemeRef is set
		if err := resolveThemeColors(config); err != nil {
			// Log warning but continue - colors just won't be embedded
			render.WarningfToStderr("Could not resolve theme colors: %v", err)
		}

		generator := wezterm.NewLuaGenerator()
		lua, err := generator.GenerateFromConfig(config)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		// Determine output path
		outputPath := output
		if outputPath == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			outputPath = filepath.Join(home, ".wezterm.lua")
		}

		// Expand ~ in path
		if strings.HasPrefix(outputPath, "~") {
			home, _ := os.UserHomeDir()
			outputPath = filepath.Join(home, outputPath[1:])
		}

		// Write file
		if err := os.WriteFile(outputPath, []byte(lua), 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		render.Successf("WezTerm configuration applied to %s", outputPath)
		return nil
	},
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// resolveThemeColors loads colors from a theme and applies them to the config
func resolveThemeColors(config *wezterm.WezTerm) error {
	if config.ThemeRef == "" {
		return nil // No theme to resolve
	}

	// Load theme from library
	theme, err := library.Get(config.ThemeRef)
	if err != nil {
		return fmt.Errorf("theme not found: %s", config.ThemeRef)
	}

	// Get terminal colors from theme
	termColors := theme.ToTerminalColors()
	if len(termColors) == 0 {
		return fmt.Errorf("theme %s has no terminal colors", config.ThemeRef)
	}

	// Helper to get color with fallback
	getColor := func(key, fallback string) string {
		if color, ok := termColors[key]; ok && color != "" {
			return color
		}
		return fallback
	}

	// Build WezTerm ColorConfig
	config.Colors = &wezterm.ColorConfig{
		Foreground:  getColor(palette.ColorFg, "#c0caf5"),
		Background:  getColor(palette.ColorBg, "#1a1b26"),
		CursorBg:    getColor(palette.TermCursor, termColors[palette.ColorFg]),
		CursorFg:    getColor(palette.TermCursorText, termColors[palette.ColorBg]),
		SelectionBg: getColor(palette.TermSelection, "#283457"),
		SelectionFg: getColor(palette.TermSelectionText, termColors[palette.ColorFg]),
		ANSI: []string{
			getColor(palette.TermBlack, "#15161e"),
			getColor(palette.TermRed, "#f7768e"),
			getColor(palette.TermGreen, "#9ece6a"),
			getColor(palette.TermYellow, "#e0af68"),
			getColor(palette.TermBlue, "#7aa2f7"),
			getColor(palette.TermMagenta, "#bb9af7"),
			getColor(palette.TermCyan, "#7dcfff"),
			getColor(palette.TermWhite, "#a9b1d6"),
		},
		Brights: []string{
			getColor(palette.TermBrightBlack, "#414868"),
			getColor(palette.TermBrightRed, "#f7768e"),
			getColor(palette.TermBrightGreen, "#9ece6a"),
			getColor(palette.TermBrightYellow, "#e0af68"),
			getColor(palette.TermBrightBlue, "#7aa2f7"),
			getColor(palette.TermBrightMagenta, "#bb9af7"),
			getColor(palette.TermBrightCyan, "#7dcfff"),
			getColor(palette.TermBrightWhite, "#c0caf5"),
		},
	}

	return nil
}

func init() {
	// WezTerm subcommands
	weztermCmd.AddCommand(weztermListCmd)
	weztermCmd.AddCommand(weztermShowCmd)
	weztermCmd.AddCommand(weztermGenerateCmd)
	weztermCmd.AddCommand(weztermApplyCmd)

	// Flags
	weztermApplyCmd.Flags().StringP("output", "o", "", "Output file (default: ~/.wezterm.lua)")
}
