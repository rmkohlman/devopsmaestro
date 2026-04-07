package main

import (
	"fmt"
	"io"
	"os"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTerminal/terminalops/shell"

	"github.com/spf13/cobra"
)

// =============================================================================
// SHELL COMMANDS
// =============================================================================

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Manage shell configurations (aliases, env vars, functions)",
	Long: `Manage shell configuration like aliases, environment variables, and functions.

Shell configs define the non-prompt parts of your shell setup. Use profiles
to combine shell configs with prompts and plugins.`,
}

var shellApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a shell configuration from file",
	Long: `Apply a shell configuration from a YAML file.

Examples:
  dvt shell apply -f my-shell.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file with -f flag")
		}

		store := getShellStore()

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

			s, err := shell.Parse(data)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", source, err)
			}

			existing, _ := store.Get(s.Name)
			action := "created"
			if existing != nil {
				action = "updated"
			}

			if err := store.Save(s); err != nil {
				return fmt.Errorf("failed to save shell config: %w", err)
			}

			render.Successf("Shell config '%s' %s (from %s)", s.Name, action, source)
		}

		return nil
	},
}

var shellGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get shell configuration(s)",
	Long: `Get shell configurations.

With no arguments, lists all installed shell configurations.
With a name argument, gets a specific shell configuration.

Examples:
  dvt shell get              # List installed shell configs
  dvt shell get my-shell     # Get specific shell config`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// List mode
			store := getShellStore()
			shells, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list shell configs: %w", err)
			}

			if len(shells) == 0 {
				render.Info("No shell configs installed")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			return outputShells(shells, format)
		}
		// Single get mode
		name := args[0]
		store := getShellStore()

		s, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("shell config not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputShell(s, format)
	},
}

var shellGenerateCmd = &cobra.Command{
	Use:   "generate <name>",
	Short: "Generate shell config section (stdout)",
	Long: `Generate the shell configuration section for a shell config.

This outputs aliases, environment variables, and functions for your .zshrc/.bashrc.

Examples:
  dvt shell generate my-shell
  dvt shell generate my-shell >> ~/.zshrc`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store := getShellStore()

		s, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("shell config not found: %s", name)
		}

		gen := shell.NewGenerator()
		output, err := gen.Generate(s)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		fmt.Print(output)
		return nil
	},
}

func init() {
	// Shell subcommands
	shellCmd.AddCommand(shellApplyCmd)
	shellCmd.AddCommand(shellGetCmd)
	shellCmd.AddCommand(shellGenerateCmd)

	// Flags
	shellApplyCmd.Flags().StringSliceP("filename", "f", nil, "Shell YAML file(s)")
	shellGetCmd.Flags().StringP("output", "o", "yaml", "Output format: table, yaml, json")

	// Hidden backward-compat alias for deprecated verb in shell (after flags)
	shellCmd.AddCommand(hiddenAlias("list", shellGetCmd))
}
