package cmd

import (
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/render"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// applyCmd is the root 'apply' command for kubectl-style resource application
// Usage: dvm apply -f <file> or dvm apply nvim plugin -f <file>
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a configuration from file",
	Long: `Apply a configuration to a resource from a YAML file.

Examples:
  dvm apply -f workspace.yaml           # Apply any resource
  dvm apply nvim plugin -f plugin.yaml  # Apply nvim plugin
  dvm apply -f plugin1.yaml -f plugin2.yaml  # Apply multiple files`,
}

// applyNvimCmd is the 'nvim' subcommand under 'apply'
var applyNvimCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Apply nvim resources",
	Long: `Apply nvim-related resources (plugins, themes) from YAML files.

Examples:
  dvm apply nvim plugin -f telescope.yaml
  dvm apply nvim theme -f tokyonight.yaml`,
}

// applyNvimPluginCmd applies a nvim plugin from file
// Usage: dvm apply nvim plugin -f <file>
var applyNvimPluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Apply a nvim plugin from file",
	Long: `Apply a nvim plugin definition from a YAML file to the database.
If the plugin already exists, it will be updated.

Examples:
  dvm apply nvim plugin -f telescope.yaml
  dvm apply nvim plugin -f plugin1.yaml -f plugin2.yaml
  cat plugin.yaml | dvm apply nvim plugin -f -`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file with -f flag")
		}

		for _, filePath := range files {
			if err := applyNvimPluginFromFile(cmd, filePath); err != nil {
				return err
			}
		}

		return nil
	},
}

// applyNvimPluginFromFile applies a single plugin file using nvimops.Manager.
// This provides a unified storage mechanism shared with nvp CLI.
func applyNvimPluginFromFile(cmd *cobra.Command, filePath string) error {
	var data []byte
	var err error
	var source string

	// Read from stdin if filePath is "-"
	if filePath == "-" {
		data, err = os.ReadFile("/dev/stdin")
		source = "stdin"
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %v", err)
		}
	} else {
		data, err = os.ReadFile(filePath)
		source = filePath
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
	}

	// Parse YAML using the nvimops plugin parser
	p, err := plugin.ParseYAML(data)
	if err != nil {
		return fmt.Errorf("failed to parse plugin YAML: %v", err)
	}

	// Get nvim manager (uses DBStoreAdapter internally)
	mgr, err := getNvimManager(cmd)
	if err != nil {
		return fmt.Errorf("failed to get nvim manager: %v", err)
	}
	defer mgr.Close()

	// Check if plugin already exists (for messaging)
	existing, _ := mgr.Get(p.Name)
	action := "created"
	if existing != nil {
		action = "configured"
	}

	// Apply (upsert) the plugin
	if err := mgr.Apply(p); err != nil {
		return fmt.Errorf("failed to apply plugin: %v", err)
	}

	render.Success(fmt.Sprintf("Plugin '%s' %s (from %s)", p.Name, action, source))
	return nil
}

// applyNvimThemeCmd applies a nvim theme from file (placeholder for future)
var applyNvimThemeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Apply a nvim theme from file",
	Long: `Apply a nvim theme definition from a YAML file.

Note: Theme management is currently available via the standalone 'nvp' CLI.
This command will be integrated in a future version.

For now, use: nvp theme apply -f <file>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		render.Info("Theme management is currently available via the standalone 'nvp' CLI.")
		render.Info("")
		render.Info("Use this command instead:\n  nvp theme apply -f <file>")
		render.Info("")
		render.Info("Integration with dvm is planned for a future release.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Add nvim subcommand to apply
	applyCmd.AddCommand(applyNvimCmd)

	// Add plugin and theme under nvim
	applyNvimCmd.AddCommand(applyNvimPluginCmd)
	applyNvimCmd.AddCommand(applyNvimThemeCmd)

	// Add flags
	applyNvimPluginCmd.Flags().StringSliceP("filename", "f", []string{}, "Plugin YAML file(s) to apply (use '-' for stdin)")
	applyNvimPluginCmd.MarkFlagRequired("filename")

	applyNvimThemeCmd.Flags().StringSliceP("filename", "f", []string{}, "Theme YAML file(s) to apply (use '-' for stdin)")
}
