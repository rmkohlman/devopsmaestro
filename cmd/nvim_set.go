// Package cmd provides CLI commands for DevOpsMaestro.
// This file implements kubectl-style 'set' commands for nvim resources.
//
// Usage:
//
//	dvm set nvim plugin -w <workspace> <names...>   # Add plugins to workspace
//	dvm set nvim plugin -w <workspace> --all        # Add all global plugins
//	dvm set nvim plugin -w <workspace> --clear      # Remove all plugins
package cmd

import (
	"fmt"

	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// Flags for set nvim plugin command
var (
	setNvimWorkspaceFlag string
	setNvimProjectFlag   string
	setNvimAllFlag       bool
	setNvimClearFlag     bool
)

// setCmd is the root 'set' command
var setCmd = &cobra.Command{
	Use:   "set",
	Short: "Set resource configurations",
	Long: `Set configurations for resources.

Examples:
  dvm set nvim plugin -w dev treesitter lspconfig
  dvm set nvim plugin -w dev --all
  dvm set nvim plugin -w dev --clear`,
}

// setNvimCmd is the 'nvim' subcommand under 'set'
var setNvimCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Set nvim configurations",
	Long: `Set nvim-related configurations.

Examples:
  dvm set nvim plugin -w dev treesitter lspconfig`,
}

// setNvimPluginCmd adds plugins to a workspace
// Usage: dvm set nvim plugin -w <workspace> <names...>
var setNvimPluginCmd = &cobra.Command{
	Use:   "plugin [names...]",
	Short: "Add plugins to a workspace",
	Long: `Add nvim plugins to a workspace's configuration.

Plugins must exist in the global library (~/.nvp/plugins/).
Use 'dvm get nvim plugins' to see available plugins.

The -w flag is required to specify which workspace to configure.

Examples:
  dvm set nvim plugin -w dev treesitter lspconfig telescope
  dvm set nvim plugin -w dev --all      # Add all global plugins
  dvm set nvim plugin -w dev --clear    # Remove all plugins from workspace
  dvm set nvim plugin -p myproj -w dev treesitter  # Explicit project`,
	RunE: runSetNvimPlugin,
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setNvimCmd)
	setNvimCmd.AddCommand(setNvimPluginCmd)

	// Add flags
	setNvimPluginCmd.Flags().StringVarP(&setNvimWorkspaceFlag, "workspace", "w", "", "Workspace to configure (required)")
	setNvimPluginCmd.Flags().StringVarP(&setNvimProjectFlag, "project", "p", "", "Project for workspace (defaults to active)")
	setNvimPluginCmd.Flags().BoolVar(&setNvimAllFlag, "all", false, "Add all plugins from global library")
	setNvimPluginCmd.Flags().BoolVar(&setNvimClearFlag, "clear", false, "Remove all plugins from workspace")

	// Mark workspace as required
	setNvimPluginCmd.MarkFlagRequired("workspace")
}

func runSetNvimPlugin(cmd *cobra.Command, args []string) error {
	// Validate flags
	if setNvimClearFlag {
		return runClearWorkspacePlugins(cmd)
	}

	if !setNvimAllFlag && len(args) == 0 {
		return fmt.Errorf("specify plugin names, use --all, or use --clear")
	}

	// Get workspace
	workspace, projectName, err := getWorkspaceForPlugins(cmd, setNvimProjectFlag, setNvimWorkspaceFlag)
	if err != nil {
		return err
	}

	// Get datastore for saving
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get workspace plugin manager
	wsMgr, err := NewWorkspacePluginManager()
	if err != nil {
		return err
	}

	// Get global plugins from database via nvimops manager
	nvimMgr, err := getNvimManager(cmd)
	if err != nil {
		return err
	}
	defer nvimMgr.Close()

	allPlugins, err := nvimMgr.List()
	if err != nil {
		return fmt.Errorf("failed to list global plugins: %w", err)
	}

	// Extract enabled plugin names
	var globalPlugins []string
	for _, p := range allPlugins {
		if p.Enabled {
			globalPlugins = append(globalPlugins, p.Name)
		}
	}

	if len(globalPlugins) == 0 {
		render.Error("No plugins found in database")
		render.Info("Import plugins first with: nvp library install --all")
		return nil
	}

	// Determine which plugins to add
	var toAdd []string
	if setNvimAllFlag {
		toAdd = globalPlugins
	} else {
		toAdd = args
	}

	// Add plugins
	added, skipped, notFound := wsMgr.AddPlugins(workspace, toAdd, globalPlugins)

	// Save to database
	if err := ds.UpdateWorkspace(workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	// Report results
	if len(added) > 0 {
		render.Success(fmt.Sprintf("Added %d plugin(s) to workspace '%s':", len(added), workspace.Name))
		for _, p := range added {
			fmt.Printf("  + %s\n", p)
		}
	}

	if len(skipped) > 0 {
		render.Info(fmt.Sprintf("Skipped %d plugin(s) (already configured):", len(skipped)))
		for _, p := range skipped {
			fmt.Printf("  • %s\n", p)
		}
	}

	if len(notFound) > 0 {
		render.Warning(fmt.Sprintf("Not found in global library (%d):", len(notFound)))
		for _, p := range notFound {
			fmt.Printf("  ? %s\n", p)
		}
		render.Info("Install missing plugins with: nvp library install <name>")
	}

	if len(added) > 0 {
		fmt.Println()
		render.Info(fmt.Sprintf("View configured plugins: dvm get nvim plugins -w %s", workspace.Name))
		render.Info(fmt.Sprintf("Rebuild workspace to apply: dvm build %s --force", workspace.Name))
		_ = projectName // Used for context in messages if needed
	}

	return nil
}

func runClearWorkspacePlugins(cmd *cobra.Command) error {
	// Get workspace
	workspace, _, err := getWorkspaceForPlugins(cmd, setNvimProjectFlag, setNvimWorkspaceFlag)
	if err != nil {
		return err
	}

	// Get datastore
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get workspace plugin manager
	mgr, err := NewWorkspacePluginManager()
	if err != nil {
		return err
	}

	count := mgr.ClearPlugins(workspace)

	if count == 0 {
		render.Info(fmt.Sprintf("Workspace '%s' has no plugins configured", workspace.Name))
		return nil
	}

	// Save to database
	if err := ds.UpdateWorkspace(workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	render.Success(fmt.Sprintf("Cleared %d plugin(s) from workspace '%s'", count, workspace.Name))
	render.Info("Build will now use all plugins from global library")
	fmt.Println()
	render.Info(fmt.Sprintf("Rebuild workspace to apply: dvm build %s --force", workspace.Name))

	return nil
}
