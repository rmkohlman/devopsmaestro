package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmkohlman/MaestroPalette"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTerminal/terminalops/wezterm"
	weztermlib "github.com/rmkohlman/MaestroTerminal/terminalops/wezterm/library"
	"github.com/rmkohlman/MaestroTheme/library"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
	Use:   "get",
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
	Use:   "describe <name>",
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

var weztermUseCmd = &cobra.Command{
	Use:     "use <name>",
	Aliases: []string{"apply"},
	Short:   "Use a WezTerm configuration preset",
	Long: `Use a WezTerm configuration preset by writing to ~/.wezterm.lua.

Sets the active WezTerm configuration from a named preset.
Automatically creates a timestamped backup before overwriting.

Examples:
  dvt wezterm use default
  dvt wezterm use default --output-file ~/custom.wezterm.lua
  dvt wezterm use default --no-backup
  dvt wezterm use default --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		output, _ := cmd.Flags().GetString("output-file")
		noBackup, _ := cmd.Flags().GetBool("no-backup")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

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

		// Create backup if file exists and --no-backup is not set
		if !noBackup {
			if _, err := os.Stat(outputPath); err == nil {
				backupPath, err := wezterm.BackupFile(outputPath)
				if err != nil {
					render.WarningfToStderr("Failed to create backup: %v", err)
				} else if backupPath != "" {
					render.Infof("Backup created: %s", backupPath)
				}
			}
		}

		// Dry run mode - just show what would be written
		if dryRun {
			render.Infof("Dry run - would write to: %s", outputPath)
			fmt.Print(lua)
			return nil
		}

		// Write file
		if err := os.WriteFile(outputPath, []byte(lua), 0644); err != nil {
			return fmt.Errorf("failed to write config: %w", err)
		}

		render.Successf("WezTerm configuration written to %s", outputPath)
		return nil
	},
}

var weztermParseCmd = &cobra.Command{
	Use:   "parse <path>",
	Short: "Parse a WezTerm Lua config file",
	Long: `Parse an existing WezTerm Lua configuration file and output as dvt-compatible YAML.

Reads ~/.wezterm.lua (or specified path) and extracts: fonts, window settings,
colors, theme, leader key, keybindings, tab bar, pane settings, scrollback, workspace.

Examples:
  dvt wezterm parse ~/.wezterm.lua
  dvt wezterm parse ~/custom.wezterm.lua -o json`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var configPath string

		if len(args) > 0 {
			configPath = args[0]
		} else {
			// Default to ~/.wezterm.lua
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			configPath = filepath.Join(home, ".wezterm.lua")
		}

		// Expand ~ in path
		if strings.HasPrefix(configPath, "~") {
			home, _ := os.UserHomeDir()
			configPath = filepath.Join(home, configPath[1:])
		}

		// Parse the Lua config
		parser := wezterm.NewLuaParser()
		config, err := parser.ParseFromFile(configPath)
		if err != nil {
			return fmt.Errorf("failed to parse config: %w", err)
		}

		// Convert to YAML
		yamlConfig := config.ToYAML()

		// Output in requested format
		outputFormat, _ := cmd.Flags().GetString("output")
		if outputFormat == "json" {
			jsonBytes, err := yamlConfigToJSON(yamlConfig)
			if err != nil {
				return fmt.Errorf("failed to convert to JSON: %w", err)
			}
			fmt.Print(string(jsonBytes))
		} else {
			yamlBytes, err := yamlConfigToYAML(yamlConfig)
			if err != nil {
				return fmt.Errorf("failed to convert to YAML: %w", err)
			}
			fmt.Print(string(yamlBytes))
		}

		return nil
	},
}

var weztermDiffCmd = &cobra.Command{
	Use:   "diff <preset>",
	Short: "Show diff between preset and current config",
	Long: `Show the differences between a preset and the current ~/.wezterm.lua.

Compares the named preset configuration against the existing WezTerm config
and displays the differences in a unified diff format.

Examples:
  dvt wezterm diff default
  dvt wezterm diff default ~/.wezterm.lua`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presetName := args[0]

		// Get preset config
		lib, err := weztermlib.NewWeztermLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		presetConfig, err := lib.Get(presetName)
		if err != nil {
			return fmt.Errorf("preset not found: %s", presetName)
		}

		// Resolve theme colors if ThemeRef is set
		if err := resolveThemeColors(presetConfig); err != nil {
			render.WarningfToStderr("Could not resolve theme colors: %v", err)
		}

		// Generate Lua for preset
		generator := wezterm.NewLuaGenerator()
		presetLua, err := generator.GenerateFromConfig(presetConfig)
		if err != nil {
			return fmt.Errorf("failed to generate preset config: %w", err)
		}

		// Get current config path
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		currentPath := filepath.Join(home, ".wezterm.lua")

		// Check if current config exists
		currentData, err := os.ReadFile(currentPath)
		if err != nil {
			if os.IsNotExist(err) {
				render.Infof("No existing config at %s", currentPath)
				render.Info("Preset config:")
				fmt.Print(presetLua)
				return nil
			}
			return fmt.Errorf("failed to read current config: %w", err)
		}

		// Show diff
		currentLua := string(currentData)
		diff := computeDiff(currentLua, presetLua)
		if diff == "" {
			render.Info("No differences found")
		} else {
			fmt.Print(diff)
		}

		return nil
	},
}

var weztermMergeCmd = &cobra.Command{
	Use:   "merge <preset>",
	Short: "Merge preset with existing config",
	Long: `Merge a preset with the existing ~/.wezterm.lua configuration.

Merges the named preset with the current WezTerm config. The preset values
take precedence over existing values. Optionally, a merge file can provide
additional overrides.

Examples:
  dvt wezterm merge default
  dvt wezterm merge default --merge-file ~/overrides.lua
  dvt wezterm merge default --dry-run`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		presetName := args[0]
		mergeFile, _ := cmd.Flags().GetString("merge-file")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		output, _ := cmd.Flags().GetString("output-file")

		// Get preset config
		lib, err := weztermlib.NewWeztermLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		presetConfig, err := lib.Get(presetName)
		if err != nil {
			return fmt.Errorf("preset not found: %s", presetName)
		}

		// Resolve theme colors if ThemeRef is set
		if err := resolveThemeColors(presetConfig); err != nil {
			render.WarningfToStderr("Could not resolve theme colors: %v", err)
		}

		// Get current config if exists
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		currentPath := filepath.Join(home, ".wezterm.lua")

		var existingConfig *wezterm.WezTerm
		if currentData, err := os.ReadFile(currentPath); err == nil {
			parser := wezterm.NewLuaParser()
			existingConfig, err = parser.Parse(currentData)
			if err != nil {
				render.WarningfToStderr("Could not parse existing config: %v", err)
			}
		}

		// Parse merge file if provided
		var mergeOverrides *wezterm.WezTerm
		if mergeFile != "" {
			parser := wezterm.NewLuaParser()
			mergeOverrides, err = parser.ParseFromFile(mergeFile)
			if err != nil {
				return fmt.Errorf("failed to parse merge file: %w", err)
			}
		}

		// Merge configs: existing < preset < merge_overrides
		mergedConfig := mergeWezTermConfigs(existingConfig, presetConfig, mergeOverrides)

		// Generate Lua
		generator := wezterm.NewLuaGenerator()
		lua, err := generator.GenerateFromConfig(mergedConfig)
		if err != nil {
			return fmt.Errorf("failed to generate merged config: %w", err)
		}

		// Determine output path
		outputPath := output
		if outputPath == "" {
			outputPath = currentPath
		}

		// Expand ~ in path
		if strings.HasPrefix(outputPath, "~") {
			outputPath = filepath.Join(home, outputPath[1:])
		}

		// Dry run mode
		if dryRun {
			render.Infof("Dry run - merged config for: %s", outputPath)
			fmt.Print(lua)
			return nil
		}

		// Create backup
		if _, err := os.Stat(outputPath); err == nil {
			backupPath, err := wezterm.BackupFile(outputPath)
			if err != nil {
				render.WarningfToStderr("Failed to create backup: %v", err)
			} else if backupPath != "" {
				render.Infof("Backup created: %s", backupPath)
			}
		}

		// Write merged config
		if err := os.WriteFile(outputPath, []byte(lua), 0644); err != nil {
			return fmt.Errorf("failed to write merged config: %w", err)
		}

		render.Successf("Merged WezTerm configuration written to %s", outputPath)
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

// yamlConfigToYAML converts a WeztermYAML to YAML string.
func yamlConfigToYAML(config *wezterm.WeztermYAML) ([]byte, error) {
	return yaml.Marshal(config)
}

// yamlConfigToJSON converts a WeztermYAML to JSON string.
func yamlConfigToJSON(config *wezterm.WeztermYAML) ([]byte, error) {
	return json.Marshal(config)
}

// computeDiff computes a unified diff between two strings.
func computeDiff(old, new string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var diff []string
	diff = append(diff, "--- current config")
	diff = append(diff, "+++ preset config")

	// Simple line-by-line comparison
	maxLines := len(oldLines)
	if len(newLines) > maxLines {
		maxLines = len(newLines)
	}

	for i := 0; i < maxLines; i++ {
		oldLine := ""
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		newLine := ""
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine != newLine {
			if oldLine != "" {
				diff = append(diff, fmt.Sprintf("-%d: %s", i+1, oldLine))
			}
			if newLine != "" {
				diff = append(diff, fmt.Sprintf("+%d: %s", i+1, newLine))
			}
		}
	}

	return strings.Join(diff, "\n")
}

// mergeWezTermConfigs merges three WezTerm configurations.
// Priority: existing < preset < mergeOverrides
func mergeWezTermConfigs(existing, preset, overrides *wezterm.WezTerm) *wezterm.WezTerm {
	// Start with a copy of existing (lowest priority base)
	var result wezterm.WezTerm
	if existing != nil {
		result = *existing
	}

	// Helper to copy non-zero values from source to target
	copyIfSet := func(target, source *wezterm.WezTerm) {
		if source == nil {
			return
		}
		// Font
		if source.Font.Family != "" {
			target.Font.Family = source.Font.Family
		}
		if source.Font.Size > 0 {
			target.Font.Size = source.Font.Size
		}
		// Window
		if source.Window.Opacity > 0 {
			target.Window.Opacity = source.Window.Opacity
		}
		if source.Window.Blur > 0 {
			target.Window.Blur = source.Window.Blur
		}
		if source.Window.Decorations != "" {
			target.Window.Decorations = source.Window.Decorations
		}
		if source.Window.InitialRows > 0 {
			target.Window.InitialRows = source.Window.InitialRows
		}
		if source.Window.InitialCols > 0 {
			target.Window.InitialCols = source.Window.InitialCols
		}
		if source.Window.CloseOnExit != "" {
			target.Window.CloseOnExit = source.Window.CloseOnExit
		}
		if source.Window.PaddingLeft > 0 {
			target.Window.PaddingLeft = source.Window.PaddingLeft
		}
		if source.Window.PaddingRight > 0 {
			target.Window.PaddingRight = source.Window.PaddingRight
		}
		if source.Window.PaddingTop > 0 {
			target.Window.PaddingTop = source.Window.PaddingTop
		}
		if source.Window.PaddingBottom > 0 {
			target.Window.PaddingBottom = source.Window.PaddingBottom
		}
		// Colors
		if source.Colors != nil {
			if target.Colors == nil {
				target.Colors = &wezterm.ColorConfig{}
			}
			if source.Colors.Foreground != "" {
				target.Colors.Foreground = source.Colors.Foreground
			}
			if source.Colors.Background != "" {
				target.Colors.Background = source.Colors.Background
			}
			if source.Colors.CursorBg != "" {
				target.Colors.CursorBg = source.Colors.CursorBg
			}
			if source.Colors.CursorFg != "" {
				target.Colors.CursorFg = source.Colors.CursorFg
			}
			if source.Colors.CursorBorder != "" {
				target.Colors.CursorBorder = source.Colors.CursorBorder
			}
			if source.Colors.SelectionBg != "" {
				target.Colors.SelectionBg = source.Colors.SelectionBg
			}
			if source.Colors.SelectionFg != "" {
				target.Colors.SelectionFg = source.Colors.SelectionFg
			}
			if len(source.Colors.ANSI) > 0 {
				target.Colors.ANSI = source.Colors.ANSI
			}
			if len(source.Colors.Brights) > 0 {
				target.Colors.Brights = source.Colors.Brights
			}
		}
		// ThemeRef - only override if source has it and target doesn't
		if source.ThemeRef != "" && target.ThemeRef == "" {
			target.ThemeRef = source.ThemeRef
		}
		// Leader
		if source.Leader != nil {
			if target.Leader == nil {
				target.Leader = &wezterm.LeaderKey{}
			}
			if source.Leader.Key != "" {
				target.Leader.Key = source.Leader.Key
			}
			if source.Leader.Mods != "" {
				target.Leader.Mods = source.Leader.Mods
			}
			if source.Leader.Timeout > 0 {
				target.Leader.Timeout = source.Leader.Timeout
			}
		}
		// Keys - append unique keybindings
		if len(source.Keys) > 0 {
			existingKeys := make(map[string]bool)
			for _, k := range target.Keys {
				existingKeys[k.Key+k.Mods] = true
			}
			for _, k := range source.Keys {
				key := k.Key + k.Mods
				if !existingKeys[key] {
					target.Keys = append(target.Keys, k)
					existingKeys[key] = true
				}
			}
		}
		// KeyTables
		if len(source.KeyTables) > 0 {
			if target.KeyTables == nil {
				target.KeyTables = make(map[string][]wezterm.Keybinding)
			}
			for tableName, bindings := range source.KeyTables {
				if _, exists := target.KeyTables[tableName]; !exists {
					target.KeyTables[tableName] = bindings
				}
			}
		}
		// TabBar
		if source.TabBar != nil {
			if target.TabBar == nil {
				target.TabBar = &wezterm.TabBarConfig{}
			}
			if source.TabBar.Enabled {
				target.TabBar.Enabled = true
			}
			if source.TabBar.Position != "" {
				target.TabBar.Position = source.TabBar.Position
			}
			if source.TabBar.MaxWidth > 0 {
				target.TabBar.MaxWidth = source.TabBar.MaxWidth
			}
			if source.TabBar.ShowNewTab {
				target.TabBar.ShowNewTab = true
			}
			if source.TabBar.FancyTabBar {
				target.TabBar.FancyTabBar = true
			}
			if source.TabBar.HideTabBarIfOnly {
				target.TabBar.HideTabBarIfOnly = true
			}
		}
		// Pane
		if source.Pane != nil {
			if target.Pane == nil {
				target.Pane = &wezterm.PaneConfig{}
			}
			if source.Pane.InactiveSaturation > 0 {
				target.Pane.InactiveSaturation = source.Pane.InactiveSaturation
			}
			if source.Pane.InactiveBrightness > 0 {
				target.Pane.InactiveBrightness = source.Pane.InactiveBrightness
			}
		}
		// Scrollback
		if source.Scrollback > 0 {
			target.Scrollback = source.Scrollback
		}
		// Workspace
		if source.Workspace != "" {
			target.Workspace = source.Workspace
		}
		// Plugins
		if len(source.Plugins) > 0 {
			target.Plugins = source.Plugins
		}
	}

	// Apply preset config (middle priority)
	if preset != nil {
		copyIfSet(&result, preset)
	}

	// Apply overrides last (highest priority)
	if overrides != nil {
		copyIfSet(&result, overrides)
	}

	return &result
}

func init() {
	// WezTerm subcommands
	weztermCmd.AddCommand(weztermListCmd)
	weztermCmd.AddCommand(weztermShowCmd)
	weztermCmd.AddCommand(weztermGenerateCmd)
	weztermCmd.AddCommand(weztermUseCmd)
	weztermCmd.AddCommand(weztermParseCmd)
	weztermCmd.AddCommand(weztermDiffCmd)
	weztermCmd.AddCommand(weztermMergeCmd)

	// Flags for wezterm use
	weztermUseCmd.Flags().String("output-file", "", "Output file (default: ~/.wezterm.lua)")
	weztermUseCmd.Flags().Bool("no-backup", false, "Skip automatic backup before writing")
	weztermUseCmd.Flags().Bool("dry-run", false, "Show what would be written without writing")

	// Flags for wezterm parse
	weztermParseCmd.Flags().String("output", "yaml", "Output format (yaml or json)")

	// Flags for wezterm merge
	weztermMergeCmd.Flags().String("merge-file", "", "Additional merge override file")
	weztermMergeCmd.Flags().Bool("dry-run", false, "Show merged config without writing")
	weztermMergeCmd.Flags().String("output-file", "", "Output file (default: ~/.wezterm.lua)")

	// Hidden backward-compat aliases for deprecated verbs in wezterm (after flags)
	weztermCmd.AddCommand(hiddenAlias("list", weztermListCmd))
	weztermCmd.AddCommand(hiddenAlias("show", weztermShowCmd))
}
