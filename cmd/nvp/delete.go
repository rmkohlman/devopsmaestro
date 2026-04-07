package main

import (
	"fmt"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// =============================================================================
// DELETE COMMAND
// =============================================================================

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a plugin from the local store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		// Check exists
		if _, err := mgr.Get(name); err != nil {
			return fmt.Errorf("plugin not found: %s", name)
		}

		// Confirm unless forced
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Delete plugin '%s'? (y/N): ", name)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				render.Info("Aborted")
				return nil
			}
		}

		if err := mgr.Delete(name); err != nil {
			return fmt.Errorf("failed to delete plugin: %w", err)
		}

		render.Successf("Plugin '%s' deleted", name)
		return nil
	},
}

func init() {
	deleteCmd.Flags().Bool("force", false, "Skip confirmation")
}
