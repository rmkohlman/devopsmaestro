package main

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/pkg/terminalbridge"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTerminal/terminalops/prompt"
	promptlibrary "github.com/rmkohlman/MaestroTerminal/terminalops/prompt/library"

	"github.com/spf13/cobra"
)

// =============================================================================
// PROMPT COMMANDS
// =============================================================================

var promptCmd = &cobra.Command{
	Use:     "prompt",
	Aliases: []string{"pr"},
	Short:   "Manage terminal prompts (Starship, P10k)",
	Long: `Manage terminal prompt configurations.

Prompts define how your shell prompt looks using tools like Starship or P10k.
Use the library to get started with pre-configured prompts, then customize as needed.`,
}

var promptLibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Browse and import prompts from the library",
}

var promptLibraryListCmd = &cobra.Command{
	Use:   "get",
	Short: "List available prompts in the library",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := promptlibrary.NewPromptLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		prompts := lib.List()

		// Filter by category if specified
		category, _ := cmd.Flags().GetString("category")
		if category != "" {
			prompts = lib.ListByCategory(category)
		}

		if len(prompts) == 0 {
			render.Info("No prompts found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPrompts(prompts, format)
	},
}

var promptLibraryShowCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Show details of a library prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lib, err := promptlibrary.NewPromptLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		p, err := lib.Get(name)
		if err != nil {
			return fmt.Errorf("prompt not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPrompt(p, format)
	},
}

var promptLibraryInstallCmd = &cobra.Command{
	Use:   "import <name>...",
	Short: "Import prompts from library to local store",
	Long: `Copy prompt definitions from the built-in library to your local store.
You can then customize them or use them directly.

Examples:
  dvt prompt library import starship-default
  dvt prompt library import starship-minimal starship-powerline
  dvt prompt library import --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if !all && len(args) == 0 {
			return fmt.Errorf("specify prompt names or use --all")
		}

		lib, err := promptlibrary.NewPromptLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		fileStore := getPromptStore()

		// Also write to database so dvt prompt get/generate/set/delete work
		var dbStore *terminalbridge.DBPromptStore
		if ds := cmd.Context().Value("dataStore"); ds != nil {
			if dataStore, ok := ds.(*db.DataStore); ok {
				dbStore = terminalbridge.NewDBPromptStore(*dataStore)
			}
		}

		var prompts []*prompt.Prompt
		if all {
			prompts = lib.List()
		} else {
			for _, name := range args {
				p, err := lib.Get(name)
				if err != nil {
					render.WarningfToStderr("prompt not found in library: %s", name)
					continue
				}
				prompts = append(prompts, p)
			}
		}

		for _, p := range prompts {
			if err := fileStore.Save(p); err != nil {
				render.WarningfToStderr("failed to install %s: %v", p.Name, err)
				continue
			}
			// Sync to database
			if dbStore != nil {
				if err := dbStore.Upsert(p); err != nil {
					render.WarningfToStderr("installed %s to file store but failed to sync to database: %v", p.Name, err)
				}
			}
			render.Successf("Installed %s", p.Name)
		}

		return nil
	},
}

var promptLibraryCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List prompt categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := promptlibrary.NewPromptLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		categories := lib.Categories()
		render.Infof("Categories (%d):", len(categories))
		for _, c := range categories {
			prompts := lib.ListByCategory(c)
			render.Plainf("  %-15s (%d prompts)", c, len(prompts))
		}
		return nil
	},
}

var promptGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get prompt definition(s)",
	Long: `Get terminal prompts stored in the database.

With no arguments, lists all installed prompts.
With a name argument, gets a specific prompt definition.

Uses Resource/Handler pattern with database storage.

Examples:
  dvt prompt get                     # List all installed prompts
  dvt prompt get coolnight           # Get prompt as YAML
  dvt prompt get coolnight -o json   # Get prompt as JSON
  dvt prompt get coolnight -o table  # Get prompt as table`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// List mode
			store := getPromptStore()
			prompts, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list prompts: %w", err)
			}

			if len(prompts) == 0 {
				render.Info("No prompts installed")
				render.Info("Use 'dvt prompt library get' to see available prompts")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			return outputPrompts(prompts, format)
		}
		// Single get mode - delegate to resource handler
		return promptResourceGet(cmd, args)
	},
}

var promptApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a prompt definition from file (kubectl-style)",
	Long: `Apply a terminal prompt configuration from a YAML file using Resource/Handler pattern.

Uses database storage and supports theme variable resolution.

Examples:
  dvt prompt apply -f my-prompt.yaml
  dvt prompt apply -f -               # Read from stdin
  dvt prompt apply -f prompt1.yaml -f prompt2.yaml  # Apply multiple`,
	RunE: promptResourceApply,
}

var promptDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a prompt (kubectl-style)",
	Long: `Delete a terminal prompt from the database using Resource/Handler pattern.

Requires confirmation unless --force is used.

Examples:
  dvt prompt delete coolnight        # Delete with confirmation
  dvt prompt delete coolnight --force  # Delete without confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: promptResourceDelete,
}

var promptGenerateCmd = &cobra.Command{
	Use:   "generate <name>",
	Short: "Generate config file for a prompt (kubectl-style)",
	Long: `Generate the configuration file for a terminal prompt stored in the database.

Uses Resource/Handler pattern with theme variable resolution.
For Starship prompts, this outputs starship.toml content.

Examples:
  dvt prompt generate coolnight                    # Output to stdout
  dvt prompt generate coolnight > ~/.config/starship.toml  # Save to file`,
	Args: cobra.ExactArgs(1),
	RunE: promptResourceGenerate,
}

func init() {
	// Prompt subcommands
	promptCmd.AddCommand(promptLibraryCmd)
	promptCmd.AddCommand(promptGetCmd)
	promptCmd.AddCommand(promptApplyCmd)
	promptCmd.AddCommand(promptDeleteCmd)
	promptCmd.AddCommand(promptGenerateCmd)
	promptCmd.AddCommand(promptSetCmd)

	// Prompt library subcommands
	promptLibraryCmd.AddCommand(promptLibraryListCmd)
	promptLibraryCmd.AddCommand(promptLibraryShowCmd)
	promptLibraryCmd.AddCommand(promptLibraryInstallCmd)
	promptLibraryCmd.AddCommand(promptLibraryCategoriesCmd)

	// Flags
	promptLibraryListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	promptLibraryListCmd.Flags().StringP("category", "c", "", "Filter by category")
	promptLibraryShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	promptLibraryInstallCmd.Flags().Bool("all", false, "Import all prompts from library")
	promptGetCmd.Flags().StringP("output", "o", "yaml", "Output format: table, yaml, json")
	promptApplyCmd.Flags().StringSliceP("filename", "f", nil, "Prompt YAML file(s)")
	promptDeleteCmd.Flags().Bool("force", false, "Skip confirmation")

	// Hidden backward-compat aliases for deprecated verbs in prompt (after flags)
	promptLibraryCmd.AddCommand(hiddenAlias("list", promptLibraryListCmd))
	promptLibraryCmd.AddCommand(hiddenAlias("show", promptLibraryShowCmd))
	promptLibraryCmd.AddCommand(hiddenAlias("install", promptLibraryInstallCmd))
	promptCmd.AddCommand(hiddenAlias("list", promptGetCmd))
}
