package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTerminal/terminalops/profile"

	"github.com/spf13/cobra"
)

// =============================================================================
// PROFILE COMMANDS
// =============================================================================

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage terminal profiles (combines prompt, plugins, shell)",
	Long: `Manage terminal profiles that combine prompts, plugins, and shell settings.

Profiles are the recommended way to manage your terminal configuration. They:
  - Reference a prompt (Starship/P10k)
  - Include multiple plugins (autosuggestions, syntax-highlighting, etc.)
  - Include shell settings (aliases, env vars, functions)

Quick start with presets:
  dvt profile preset get        # See available presets
  dvt profile preset import default
  dvt profile generate          # Generate all config files`,
}

var profilePresetCmd = &cobra.Command{
	Use:   "preset",
	Short: "Manage profile presets",
}

var profilePresetListCmd = &cobra.Command{
	Use:   "get",
	Short: "List available profile presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		presets := []struct {
			name        string
			description string
		}{
			{"default", "Balanced setup with Starship, autosuggestions, and syntax-highlighting"},
			{"minimal", "Lightweight setup with just Starship and basic plugins"},
			{"power-user", "Full-featured setup with all plugins and nerd font support"},
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESCRIPTION")
		for _, p := range presets {
			fmt.Fprintf(w, "%s\t%s\n", p.name, p.description)
		}
		w.Flush()
		return nil
	},
}

var profilePresetInstallCmd = &cobra.Command{
	Use:   "import <name>",
	Short: "Import a profile preset",
	Long: `Import a profile preset and all its dependencies.

Available presets:
  default     - Balanced setup with Starship, autosuggestions, syntax-highlighting
  minimal     - Lightweight setup with just Starship
  power-user  - Full-featured setup with all plugins

Examples:
  dvt profile preset import default`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		var p *profile.Profile
		switch name {
		case "default":
			p = profile.DefaultProfile()
		case "minimal":
			p = profile.MinimalProfile()
		case "power-user":
			p = profile.PowerUserProfile()
		default:
			return fmt.Errorf("unknown preset: %s (available: default, minimal, power-user)", name)
		}

		store := getProfileStore()
		if err := store.Save(p); err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}

		render.Successf("Installed profile preset '%s'", name)
		render.Info("Run 'dvt profile generate' to generate config files")
		return nil
	},
}

var profileApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a profile from file",
	Long: `Apply a profile definition from a YAML file.

Examples:
  dvt profile apply -f my-profile.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file with -f flag")
		}

		store := getProfileStore()

		for _, file := range files {
			var data []byte
			var err error
			var source string

			if file == "-" {
				data, err = io.ReadAll(os.Stdin)
				source = "stdin"
			} else {
				data, err = os.ReadFile(file)
				source = file
			}
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", source, err)
			}

			p, err := profile.Parse(data)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", source, err)
			}

			existing, _ := store.Get(p.Name)
			action := "created"
			if existing != nil {
				action = "updated"
			}

			if err := store.Save(p); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			render.Successf("Profile '%s' %s (from %s)", p.Name, action, source)
		}

		return nil
	},
}

var profileGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get profile definition(s)",
	Long: `Get terminal profile definitions.

With no arguments, lists all installed profiles.
With a name argument, gets a specific profile definition.

Examples:
  dvt profile get              # List installed profiles
  dvt profile get default      # Get specific profile`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// List mode
			store := getProfileStore()
			profiles, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}

			if len(profiles) == 0 {
				render.Info("No profiles installed")
				render.Info("Use 'dvt profile preset get' to see available presets")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			return outputProfiles(profiles, format)
		}
		// Single get mode
		name := args[0]
		store := getProfileStore()

		p, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("profile not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputProfile(p, format)
	},
}

var profileGenerateCmd = &cobra.Command{
	Use:   "generate <name>",
	Short: "Generate all config files for a profile",
	Long: `Generate all configuration files for a profile.

This creates:
  - starship.toml (if using Starship prompt)
  - Plugin installation/loading code
  - Shell aliases, env vars, functions

By default outputs to stdout. Use --output-dir to write to files.

Examples:
  dvt profile generate default
  dvt profile generate default --output-dir ~/.config/`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store := getProfileStore()

		p, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("profile not found: %s", name)
		}

		outputDir, _ := cmd.Flags().GetString("output-dir")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		gen := profile.NewGenerator("")
		result, err := gen.Generate(p)
		if err != nil {
			return fmt.Errorf("failed to generate profile: %w", err)
		}

		if outputDir == "" || dryRun {
			// Output to stdout
			fmt.Println("# Generated by dvt profile generate")
			fmt.Println("#")
			fmt.Printf("# Profile: %s\n", p.Name)
			if p.Description != "" {
				fmt.Printf("# Description: %s\n", p.Description)
			}
			fmt.Println()

			if result.StarshipTOML != "" {
				fmt.Println("# === starship.toml ===")
				fmt.Println(result.StarshipTOML)
				fmt.Println()
			}

			if result.ZshrcPlugins != "" {
				fmt.Println("# === .zshrc (plugins) ===")
				fmt.Println(result.ZshrcPlugins)
				fmt.Println()
			}

			if result.ZshrcShell != "" {
				fmt.Println("# === .zshrc (shell) ===")
				fmt.Println(result.ZshrcShell)
			}

			if dryRun && outputDir != "" {
				fmt.Println("\n# Would write to:")
				fmt.Printf("#   %s/starship.toml\n", outputDir)
				fmt.Printf("#   %s/.zshrc.dvt\n", outputDir)
			}
			return nil
		}

		// Write to files
		// Expand ~
		if strings.HasPrefix(outputDir, "~") {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, outputDir[1:])
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if result.StarshipTOML != "" {
			starshipPath := filepath.Join(outputDir, "starship.toml")
			if err := os.WriteFile(starshipPath, []byte(result.StarshipTOML), 0644); err != nil {
				return fmt.Errorf("failed to write starship.toml: %w", err)
			}
			render.Successf("Wrote %s", starshipPath)
		}

		if result.ZshrcPlugins != "" || result.ZshrcShell != "" {
			zshrcContent := "# Generated by dvt - source this from your .zshrc\n\n"
			if result.ZshrcPlugins != "" {
				zshrcContent += "# Plugins\n" + result.ZshrcPlugins + "\n"
			}
			if result.ZshrcShell != "" {
				zshrcContent += "# Shell config\n" + result.ZshrcShell
			}

			zshrcPath := filepath.Join(outputDir, ".zshrc.dvt")
			if err := os.WriteFile(zshrcPath, []byte(zshrcContent), 0644); err != nil {
				return fmt.Errorf("failed to write .zshrc.dvt: %w", err)
			}
			render.Successf("Wrote %s", zshrcPath)
			render.Blank()
			render.Info("Add this line to your .zshrc:")
			render.Plainf("  source %s", zshrcPath)
		}

		return nil
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store := getProfileStore()

		if err := store.SetActive(name); err != nil {
			return err
		}

		render.Successf("Active profile set to '%s'", name)
		render.Info("Run 'dvt profile generate' to regenerate config files")
		return nil
	},
}

func init() {
	// Profile subcommands
	profileCmd.AddCommand(profilePresetCmd)
	profileCmd.AddCommand(profileApplyCmd)
	profileCmd.AddCommand(profileGetCmd)
	profileCmd.AddCommand(profileGenerateCmd)
	profileCmd.AddCommand(profileUseCmd)

	// Preset subcommands
	profilePresetCmd.AddCommand(profilePresetListCmd)
	profilePresetCmd.AddCommand(profilePresetInstallCmd)

	// Flags
	profileApplyCmd.Flags().StringSliceP("filename", "f", nil, "Profile YAML file(s)")
	profileGetCmd.Flags().StringP("output", "o", "yaml", "Output format: table, yaml, json")
	profileGenerateCmd.Flags().String("output-dir", "", "Output directory (default: stdout)")
	profileGenerateCmd.Flags().Bool("dry-run", false, "Show what would be generated")

	// Hidden backward-compat aliases for deprecated verbs in profile (after flags)
	profilePresetCmd.AddCommand(hiddenAlias("list", profilePresetListCmd))
	profilePresetCmd.AddCommand(hiddenAlias("install", profilePresetInstallCmd))
	profileCmd.AddCommand(hiddenAlias("list", profileGetCmd))
}
