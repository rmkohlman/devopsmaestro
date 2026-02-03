package cmd

import (
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/render"
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

		// Get nvim manager (uses DBStoreAdapter internally)
		mgr, err := getNvimManager(cmd)
		if err != nil {
			return fmt.Errorf("failed to get nvim manager: %v", err)
		}
		defer mgr.Close()

		// Get plugin
		p, err := mgr.Get(name)
		if err != nil {
			return fmt.Errorf("plugin not found: %s", name)
		}

		// Convert to YAML using the plugin package
		pluginYAML := p.ToYAML()
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
		editCommand := exec.Command(editor, tmpfile.Name())
		editCommand.Stdin = os.Stdin
		editCommand.Stdout = os.Stdout
		editCommand.Stderr = os.Stderr

		if err := editCommand.Run(); err != nil {
			return fmt.Errorf("editor exited with error: %v", err)
		}

		// Read edited content
		editedData, err := os.ReadFile(tmpfile.Name())
		if err != nil {
			return fmt.Errorf("failed to read edited file: %v", err)
		}

		// Parse edited YAML using the plugin package
		editedPlugin, err := plugin.ParseYAML(editedData)
		if err != nil {
			return fmt.Errorf("failed to parse edited YAML: %v", err)
		}

		// Apply (upsert) the updated plugin
		if err := mgr.Apply(editedPlugin); err != nil {
			return fmt.Errorf("failed to update plugin: %v", err)
		}

		render.Success(fmt.Sprintf("Plugin '%s' updated successfully", editedPlugin.Name))
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
		render.Info("Theme editing is currently available via the standalone 'nvp' CLI.")
		render.Info("")
		render.Info("Use these commands instead:\n  nvp theme get <name> -o yaml > theme.yaml\n  $EDITOR theme.yaml\n  nvp theme apply -f theme.yaml")
		render.Info("")
		render.Info("Integration with dvm is planned for a future release.")
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
