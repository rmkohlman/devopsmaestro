package cmd

import (
	"devopsmaestro/models"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// editCmd is the root 'edit' command for kubectl-style resource editing
// Usage: dvm edit nvim plugin <name>
var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit a resource in your default editor",
	Long: `Edit a resource definition in your default editor ($EDITOR).
After saving and closing the editor, changes are automatically applied.

Examples:
  dvm edit nvim plugin telescope    # Edit nvim plugin in $EDITOR`,
}

// editNvimCmd is the 'nvim' subcommand under 'edit'
var editNvimCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Edit nvim resources",
	Long: `Edit nvim-related resources (plugins, themes) in your default editor.

Examples:
  dvm edit nvim plugin telescope
  dvm edit nvim theme tokyonight`,
}

// editNvimPluginCmd edits a nvim plugin in the user's editor
// Usage: dvm edit nvim plugin <name>
var editNvimPluginCmd = &cobra.Command{
	Use:   "plugin [name]",
	Short: "Edit a nvim plugin in your default editor",
	Long: `Open a nvim plugin definition in your default editor ($EDITOR).
After editing, the plugin will be automatically applied to the database.

Examples:
  dvm edit nvim plugin telescope
  EDITOR=vim dvm edit nvim plugin mason`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Get datastore from context (injected by root command)
		datastore, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("failed to get datastore: %v", err)
		}

		// Get plugin
		plugin, err := datastore.GetPluginByName(name)
		if err != nil {
			return fmt.Errorf("failed to get plugin: %v", err)
		}

		// Convert to YAML
		pluginYAML, err := plugin.ToYAML()
		if err != nil {
			return fmt.Errorf("failed to convert plugin to YAML: %v", err)
		}

		data, err := yaml.Marshal(pluginYAML)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML: %v", err)
		}

		// Create temp file
		tmpfile, err := os.CreateTemp("", fmt.Sprintf("dvm-plugin-%s-*.yaml", name))
		if err != nil {
			return fmt.Errorf("failed to create temp file: %v", err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write(data); err != nil {
			return fmt.Errorf("failed to write temp file: %v", err)
		}
		tmpfile.Close()

		// Get editor from env
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = "vi" // Default to vi
		}

		// Open editor
		editCmd := exec.Command(editor, tmpfile.Name())
		editCmd.Stdin = os.Stdin
		editCmd.Stdout = os.Stdout
		editCmd.Stderr = os.Stderr

		if err := editCmd.Run(); err != nil {
			return fmt.Errorf("editor exited with error: %v", err)
		}

		// Read edited content
		editedData, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			return fmt.Errorf("failed to read edited file: %v", err)
		}

		// Parse and apply
		var editedYAML models.NvimPluginYAML
		if err := yaml.Unmarshal(editedData, &editedYAML); err != nil {
			return fmt.Errorf("failed to parse edited YAML: %v", err)
		}

		// Validate kind
		if editedYAML.Kind != "NvimPlugin" {
			return fmt.Errorf("invalid kind: expected 'NvimPlugin', got '%s'", editedYAML.Kind)
		}

		// Convert and update
		updatedPlugin := &models.NvimPluginDB{}
		if err := updatedPlugin.FromYAML(editedYAML); err != nil {
			return fmt.Errorf("failed to convert plugin: %v", err)
		}

		updatedPlugin.ID = plugin.ID
		updatedPlugin.CreatedAt = plugin.CreatedAt

		if err := datastore.UpdatePlugin(updatedPlugin); err != nil {
			return fmt.Errorf("failed to update plugin: %v", err)
		}

		fmt.Printf("✓ Plugin '%s' updated successfully\n", updatedPlugin.Name)
		return nil
	},
}

// editNvimThemeCmd edits a nvim theme (placeholder for future)
var editNvimThemeCmd = &cobra.Command{
	Use:   "theme [name]",
	Short: "Edit a nvim theme in your default editor",
	Long: `Edit a nvim theme definition in your default editor.

Note: Theme management is currently available via the standalone 'nvp' CLI.
This command will be integrated in a future version.

For now, use: nvp theme get <name> -o yaml > theme.yaml && $EDITOR theme.yaml`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Theme editing is currently available via the standalone 'nvp' CLI.")
		fmt.Println("")
		fmt.Println("Use these commands instead:")
		fmt.Println("  nvp theme get <name> -o yaml > theme.yaml")
		fmt.Println("  $EDITOR theme.yaml")
		fmt.Println("  nvp theme apply -f theme.yaml")
		fmt.Println("")
		fmt.Println("Integration with dvm is planned for a future release.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(editCmd)

	// Add nvim subcommand to edit
	editCmd.AddCommand(editNvimCmd)

	// Add plugin and theme under nvim
	editNvimCmd.AddCommand(editNvimPluginCmd)
	editNvimCmd.AddCommand(editNvimThemeCmd)
}
