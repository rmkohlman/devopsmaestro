package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	nvimconfig "github.com/rmkohlman/MaestroNvim/nvimops/config"
	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroSDK/render"
	theme "github.com/rmkohlman/MaestroTheme"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Neovim core configuration",
	Long: `Manage Neovim core configuration (options, keymaps, autocmds).

The config command generates a complete Neovim configuration structure
from YAML definitions, matching the lua/workspace/ directory pattern.

Generated structure:
  ~/.config/nvim/
  ├── init.lua                    # Entry point
  └── lua/workspace/
      ├── lazy.lua                # lazy.nvim bootstrap
      ├── core/
      │   ├── init.lua
      │   ├── options.lua         # vim.opt settings
      │   ├── keymaps.lua         # Key mappings
      │   └── autocmds.lua        # Autocommands
      └── plugins/
          ├── init.lua            # Base plugins
          └── *.lua               # Plugin configs

Quick Start:
  nvp config init                 # Create default core.yaml
  nvp config describe             # View current config
  nvp config generate             # Generate full nvim structure`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize core.yaml with default settings",
	Long: `Create a default core.yaml configuration file.

This creates a sensible starting configuration with:
  - Common vim options (line numbers, tabs, search, etc.)
  - Essential keymaps (leader key, window splits, etc.)
  - Useful autocmds (yank highlight, etc.)
  - Base plugins (plenary, tmux-navigator)

The file is created at ~/.nvp/core.yaml by default.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getConfigDir()
		configPath := filepath.Join(dir, "core.yaml")

		// Check if already exists
		force, _ := cmd.Flags().GetBool("force")
		if _, err := os.Stat(configPath); err == nil && !force {
			return fmt.Errorf("core.yaml already exists at %s (use --force to overwrite)", configPath)
		}

		// Create directory if needed
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Write default config
		cfg := nvimconfig.DefaultCoreConfig()
		if err := cfg.WriteYAMLFile(configPath); err != nil {
			return fmt.Errorf("failed to write core.yaml: %w", err)
		}

		render.Successf("Created %s", configPath)
		render.Info("Edit this file to customize your Neovim configuration.")
		render.Info("Then run 'nvp config generate' to create the Lua files.")
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "describe",
	Short: "Show current core configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadCoreConfig()
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("output")
		switch format {
		case "yaml", "":
			data, err := cfg.ToYAML()
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		case "json":
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		default:
			return fmt.Errorf("unknown format: %s", format)
		}
		return nil
	},
}

var configGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate complete Neovim configuration",
	Long: `Generate a complete Neovim configuration from core.yaml and installed plugins.

This creates the full lua/workspace/ directory structure:
  - init.lua (entry point)
  - lua/workspace/lazy.lua (lazy.nvim bootstrap)
  - lua/workspace/core/*.lua (options, keymaps, autocmds)
  - lua/workspace/plugins/*.lua (plugin configurations)

By default, files are written to ~/.config/nvim/
Use --output-dir to specify a different directory.

Examples:
  nvp config generate
  nvp config generate --output-dir /path/to/nvim/config
  nvp config generate --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load core config
		cfg, err := loadCoreConfig()
		if err != nil {
			// If no core.yaml exists, use defaults
			if os.IsNotExist(err) {
				render.Info("No core.yaml found, using defaults...")
				cfg = nvimconfig.DefaultCoreConfig()
			} else {
				return err
			}
		}

		// Load plugins
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		plugins, err := mgr.List()
		if err != nil {
			return fmt.Errorf("failed to list plugins: %w", err)
		}

		// Filter to enabled
		var enabled []*plugin.Plugin
		for _, p := range plugins {
			if p.Enabled {
				enabled = append(enabled, p)
			}
		}

		// Output directory
		outputDir, _ := cmd.Flags().GetString("output-dir")
		if outputDir == "" {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, ".config", "nvim")
		}

		// Expand ~
		if strings.HasPrefix(outputDir, "~") {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, outputDir[1:])
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		ns := cfg.Namespace
		if ns == "" {
			ns = "workspace"
		}

		if dryRun {
			render.Infof("Would generate Neovim config to %s:", outputDir)
			render.Plain("  init.lua")
			render.Plainf("  lua/%s/lazy.lua", ns)
			render.Plainf("  lua/%s/core/init.lua", ns)
			render.Plainf("  lua/%s/core/options.lua", ns)
			render.Plainf("  lua/%s/core/keymaps.lua", ns)
			render.Plainf("  lua/%s/core/autocmds.lua", ns)
			render.Plainf("  lua/%s/plugins/init.lua", ns)
			for _, p := range enabled {
				render.Plainf("  lua/%s/plugins/%s.lua", ns, p.Name)
			}
			// Check for active theme
			themeStore := getThemeStore()
			if activeTheme, _ := themeStore.GetActive(); activeTheme != nil {
				render.Plainf("  lua/%s/plugins/colorscheme.lua (theme: %s)", ns, activeTheme.Name)
				render.Plain("  lua/theme/init.lua")
				render.Plain("  lua/theme/palette.lua")
			}
			return nil
		}

		// Generate
		gen := nvimconfig.NewGenerator()
		if err := gen.WriteToDirectory(cfg, enabled, outputDir); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		// Generate theme if active
		themeStore := getThemeStore()
		activeTheme, _ := themeStore.GetActive()
		if activeTheme != nil {
			themeGen := theme.NewGenerator()
			generated, err := themeGen.Generate(activeTheme)
			if err != nil {
				render.WarningfToStderr("failed to generate theme: %v", err)
			} else {
				// Write theme files
				themeFiles := map[string]string{
					filepath.Join(outputDir, "lua", "theme", "palette.lua"):           generated.PaletteLua,
					filepath.Join(outputDir, "lua", "theme", "init.lua"):              generated.InitLua,
					filepath.Join(outputDir, "lua", ns, "plugins", "colorscheme.lua"): generated.PluginLua,
				}

				for path, content := range themeFiles {
					dir := filepath.Dir(path)
					if err := os.MkdirAll(dir, 0755); err != nil {
						render.WarningfToStderr("failed to create %s: %v", dir, err)
						continue
					}
					if err := os.WriteFile(path, []byte(content), 0644); err != nil {
						render.WarningfToStderr("failed to write %s: %v", path, err)
						continue
					}
				}
				render.Plainf("  Theme: %s (colorscheme.lua)", activeTheme.Name)
			}
		}

		render.Successf("Generated Neovim configuration to %s", outputDir)
		render.Plainf("  Core files: init.lua, lua/%s/core/*.lua", ns)
		render.Plainf("  Plugin files: %d plugins in lua/%s/plugins/", len(enabled), ns)
		render.Info("Restart Neovim to apply changes.")
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open core.yaml in editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getConfigDir()
		configPath := filepath.Join(dir, "core.yaml")

		// Create default if doesn't exist
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			render.Info("No core.yaml found, creating default...")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			cfg := nvimconfig.DefaultCoreConfig()
			if err := cfg.WriteYAMLFile(configPath); err != nil {
				return err
			}
		}

		// Find editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			editor = "vim"
		}

		// Open editor
		editorCmd := fmt.Sprintf("%s %s", editor, configPath)
		render.Progressf("Opening %s in %s...", configPath, editor)
		return runCommand(editorCmd)
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGenerateCmd)
	configCmd.AddCommand(configEditCmd)

	configInitCmd.Flags().Bool("force", false, "Overwrite existing core.yaml")
	configShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	configGenerateCmd.Flags().String("output-dir", "", "Output directory (default: ~/.config/nvim)")
	configGenerateCmd.Flags().Bool("dry-run", false, "Show what would be generated")

	// Hidden backward-compat alias for deprecated verb (show→describe)
	// MUST be after flag definitions — shallow copy captures FlagSet pointer at copy time
	configCmd.AddCommand(hiddenAlias("show", configShowCmd))
}

func loadCoreConfig() (*nvimconfig.CoreConfig, error) {
	dir := getConfigDir()
	configPath := filepath.Join(dir, "core.yaml")
	return nvimconfig.ParseYAMLFile(configPath)
}

func runCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := parts[0]
	args := parts[1:]

	proc := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	path, err := findExecutable(cmd)
	if err != nil {
		return err
	}

	process, err := os.StartProcess(path, append([]string{cmd}, args...), proc)
	if err != nil {
		return err
	}

	_, err = process.Wait()
	return err
}

func findExecutable(name string) (string, error) {
	if filepath.IsAbs(name) {
		return name, nil
	}

	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	for _, dir := range paths {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("executable not found: %s", name)
}
