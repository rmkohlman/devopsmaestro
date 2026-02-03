package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// deleteCmd is the root 'delete' command for kubectl-style resource deletion
// Usage: dvm delete nvim plugin <name>
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete resources",
	Long: `Delete resources by name.

Examples:
  dvm delete nvim plugin telescope    # Delete nvim plugin
  dvm delete project my-api           # Delete project (future)
  dvm delete workspace dev            # Delete workspace (future)`,
}

// deleteNvimCmd is the 'nvim' subcommand under 'delete'
var deleteNvimCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Delete nvim resources",
	Long: `Delete nvim-related resources (plugins, themes).

Examples:
  dvm delete nvim plugin telescope
  dvm delete nvim theme tokyonight`,
}

// deleteNvimPluginCmd deletes a nvim plugin
// Usage: dvm delete nvim plugin <name>
var deleteNvimPluginCmd = &cobra.Command{
	Use:   "plugin [name]",
	Short: "Delete a nvim plugin",
	Long: `Delete a nvim plugin definition from DVM's database.

This removes the plugin YAML definition that DVM stores for generating
nvim configurations in workspace containers. It does NOT affect:
- Your local nvim installation
- Any existing container images
- Plugins already installed in running containers

The plugin definition can be re-added later with 'dvm apply nvim plugin -f'.

Examples:
  dvm delete nvim plugin telescope
  dvm delete nvim plugin telescope --force  # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Get datastore from context (injected by root command)
		datastore, err := getDataStore(cmd)
		if err != nil {
			return fmt.Errorf("failed to get datastore: %v", err)
		}

		// Confirm deletion
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Delete plugin definition '%s' from DVM database? (y/N): ", name)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted")
				return nil
			}
		}

		// Delete plugin
		if err := datastore.DeletePlugin(name); err != nil {
			return fmt.Errorf("failed to delete plugin: %v", err)
		}

		fmt.Printf("✓ Plugin definition '%s' removed from DVM database\n", name)
		return nil
	},
}

// deleteNvimThemeCmd deletes a nvim theme (placeholder for future)
var deleteNvimThemeCmd = &cobra.Command{
	Use:   "theme [name]",
	Short: "Delete a nvim theme",
	Long: `Delete a nvim theme definition.

Note: Theme management is currently available via the standalone 'nvp' CLI.
This command will be integrated in a future version.

For now, use: nvp theme delete <name>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		themeName := args[0]
		fmt.Println("Theme management is currently available via the standalone 'nvp' CLI.")
		fmt.Println("")
		fmt.Printf("Use this command instead:\n")
		fmt.Printf("  nvp theme delete %s\n", themeName)
		fmt.Println("")
		fmt.Println("Integration with dvm is planned for a future release.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Add nvim subcommand to delete
	deleteCmd.AddCommand(deleteNvimCmd)

	// Add plugin and theme under nvim
	deleteNvimCmd.AddCommand(deleteNvimPluginCmd)
	deleteNvimCmd.AddCommand(deleteNvimThemeCmd)

	// Add flags
	deleteNvimPluginCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
