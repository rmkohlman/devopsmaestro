package cmd

import (
	"devopsmaestro/models"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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

// applyNvimPluginFromFile applies a single plugin file
// This reuses the same logic as pluginApplyCmd but with the new command structure
func applyNvimPluginFromFile(cmd *cobra.Command, filePath string) error {
	var data []byte
	var err error

	// Read from stdin if filePath is "-"
	if filePath == "-" {
		data, err = os.ReadFile("/dev/stdin")
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %v", err)
		}
	} else {
		// Read YAML file
		data, err = os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file: %v", err)
		}
	}

	// Parse YAML
	var pluginYAML models.NvimPluginYAML
	if err := yaml.Unmarshal(data, &pluginYAML); err != nil {
		return fmt.Errorf("failed to parse YAML: %v", err)
	}

	// Validate kind
	if pluginYAML.Kind != "NvimPlugin" {
		return fmt.Errorf("invalid kind: expected 'NvimPlugin', got '%s'", pluginYAML.Kind)
	}

	// Convert to database model
	plugin := &models.NvimPluginDB{}
	if err := plugin.FromYAML(pluginYAML); err != nil {
		return fmt.Errorf("failed to convert plugin: %v", err)
	}

	// Get datastore from context (injected by root command)
	datastore, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("failed to get datastore: %v", err)
	}

	// Check if plugin already exists
	existing, err := datastore.GetPluginByName(plugin.Name)
	if err == nil && existing != nil {
		// Update existing plugin
		plugin.ID = existing.ID
		plugin.CreatedAt = existing.CreatedAt
		if err := datastore.UpdatePlugin(plugin); err != nil {
			return fmt.Errorf("failed to update plugin: %v", err)
		}
		source := filePath
		if filePath == "-" {
			source = "stdin"
		}
		fmt.Printf("✓ Plugin '%s' configured (from %s)\n", plugin.Name, source)
	} else {
		// Create new plugin
		if err := datastore.CreatePlugin(plugin); err != nil {
			return fmt.Errorf("failed to create plugin: %v", err)
		}
		source := filePath
		if filePath == "-" {
			source = "stdin"
		}
		fmt.Printf("✓ Plugin '%s' created (from %s)\n", plugin.Name, source)
	}

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
		fmt.Println("Theme management is currently available via the standalone 'nvp' CLI.")
		fmt.Println("")
		fmt.Println("Use this command instead:")
		fmt.Println("  nvp theme apply -f <file>")
		fmt.Println("")
		fmt.Println("Integration with dvm is planned for a future release.")
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
