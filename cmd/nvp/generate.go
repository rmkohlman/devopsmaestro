package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/rmkohlman/MaestroNvim/nvimops/library"
	"github.com/rmkohlman/MaestroNvim/nvimops/plugin"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

// =============================================================================
// GENERATE COMMANDS
// =============================================================================

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Lua files for all enabled plugins",
	Long: `Generate lazy.nvim compatible Lua files for all enabled plugins.

By default, files are written to ~/.config/nvim/lua/plugins/nvp/
Use --output-dir to specify a different directory.

Examples:
  nvp generate
  nvp generate --output-dir ~/.config/nvim/lua/plugins/managed
  nvp generate --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		outputDir, _ := cmd.Flags().GetString("output-dir")
		if outputDir == "" {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, ".config", "nvim", "lua", "plugins", "nvp")
		}

		// Expand ~
		if strings.HasPrefix(outputDir, "~") {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, outputDir[1:])
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		slog.Debug("generate command", "outputDir", outputDir, "dryRun", dryRun)

		plugins, err := mgr.List()
		if err != nil {
			return fmt.Errorf("failed to list plugins: %w", err)
		}

		// Filter to enabled only
		var enabled []*plugin.Plugin
		for _, p := range plugins {
			if p.Enabled {
				enabled = append(enabled, p)
			}
		}

		slog.Info("generating Lua files", "total", len(plugins), "enabled", len(enabled))

		if len(enabled) == 0 {
			render.Info("No enabled plugins to generate")
			return nil
		}

		if dryRun {
			render.Infof("Would generate %d Lua files to %s:", len(enabled), outputDir)
			for _, p := range enabled {
				render.Plainf("  %s.lua", p.Name)
			}
			return nil
		}

		// Create output directory
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Generate files
		gen := plugin.NewGenerator()
		for _, p := range enabled {
			lua, err := gen.GenerateLuaFile(p)
			if err != nil {
				render.WarningfToStderr("failed to generate %s: %v", p.Name, err)
				continue
			}

			filename := filepath.Join(outputDir, p.Name+".lua")
			if err := os.WriteFile(filename, []byte(lua), 0644); err != nil {
				render.WarningfToStderr("failed to write %s: %v", filename, err)
				continue
			}

			if verbose {
				render.Plainf("  Generated %s", filename)
			}
		}

		render.Successf("Generated %d Lua files to %s", len(enabled), outputDir)
		return nil
	},
}

var generateLuaCmd = &cobra.Command{
	Use:   "generate-lua <name>",
	Short: "Generate Lua for a single plugin (stdout)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		p, err := mgr.Get(name)
		if err != nil {
			// Try library as fallback
			lib, libErr := library.NewLibrary()
			if libErr == nil {
				if libPlugin, ok := lib.Get(name); ok {
					p = libPlugin
				}
			}
			if p == nil {
				return fmt.Errorf("plugin not found: %s", name)
			}
		}

		gen := plugin.NewGenerator()
		lua, err := gen.GenerateLuaFile(p)
		if err != nil {
			return fmt.Errorf("failed to generate Lua: %w", err)
		}

		fmt.Print(lua)
		return nil
	},
}

func init() {
	generateCmd.Flags().String("output-dir", "", "Output directory")
	generateCmd.Flags().Bool("dry-run", false, "Show what would be generated")
}
