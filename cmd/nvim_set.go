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
	"encoding/json"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// Flags for set nvim plugin command
var (
	setNvimWorkspaceFlag string
	setNvimAppFlag       string
	setNvimAllFlag       bool
	setNvimClearFlag     bool
	setNvimGlobal        bool
	setNvimPluginOutput  string
	setNvimPluginDryRun  bool
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
	Short: "Add plugins to a workspace or set global defaults",
	Long: `Add nvim plugins to a workspace's configuration or set global defaults.

Plugins must exist in the global library (~/.nvp/plugins/).
Use 'dvm get nvim plugins' to see available plugins.

For workspace operations, the -w flag is required to specify which workspace to configure.

Examples:
  dvm set nvim plugin -w dev treesitter lspconfig telescope
  dvm set nvim plugin -w dev --all      # Add all global plugins
  dvm set nvim plugin -w dev --clear    # Remove all plugins from workspace
  dvm set nvim plugin -a myapp -w dev treesitter  # Explicit app
  
  # Global defaults (replace-only semantics):
  dvm set nvim plugin lazygit telescope --global    # Set default plugins
  dvm set nvim plugin --all --global                # Set all plugins as defaults
  dvm set nvim plugin --clear --global              # Clear default plugins`,
	RunE: runSetNvimPlugin,
}

func init() {
	rootCmd.AddCommand(setCmd)
	setCmd.AddCommand(setNvimCmd)
	setNvimCmd.AddCommand(setNvimPluginCmd)

	// Add flags
	setNvimPluginCmd.Flags().StringVarP(&setNvimWorkspaceFlag, "workspace", "w", "", "Workspace to configure")
	setNvimPluginCmd.Flags().StringVarP(&setNvimAppFlag, "app", "a", "", "App for workspace (defaults to active)")
	setNvimPluginCmd.Flags().BoolVar(&setNvimAllFlag, "all", false, "Add all plugins from global library")
	setNvimPluginCmd.Flags().BoolVar(&setNvimClearFlag, "clear", false, "Remove all plugins from workspace or clear global defaults")
	setNvimPluginCmd.Flags().BoolVar(&setNvimGlobal, "global", false, "Set global default plugins")

	// Add kubectl-style flags
	setNvimPluginCmd.Flags().StringVarP(&setNvimPluginOutput, "output", "o", "", "Output format (json, yaml, plain, table, colored)")
	setNvimPluginCmd.Flags().BoolVar(&setNvimPluginDryRun, "dry-run", false, "Preview changes without applying")

	// Make workspace and global mutually exclusive
	setNvimPluginCmd.MarkFlagsMutuallyExclusive("workspace", "global")
}

func runSetNvimPlugin(cmd *cobra.Command, args []string) error {
	// Handle global defaults first
	if setNvimGlobal {
		return runSetGlobalDefaultPlugins(cmd, args)
	}

	// Workspace operations require the workspace flag
	if setNvimWorkspaceFlag == "" {
		return fmt.Errorf("workspace flag (-w) is required for workspace operations")
	}

	// Validate flags
	if setNvimClearFlag {
		return runClearWorkspacePlugins(cmd)
	}

	if !setNvimAllFlag && len(args) == 0 {
		return fmt.Errorf("specify plugin names, use --all, or use --clear")
	}

	// Get workspace
	workspace, appName, err := getWorkspaceForPlugins(cmd, setNvimAppFlag, setNvimWorkspaceFlag)
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
		_ = appName // Used for context in messages if needed
	}

	return nil
}

func runClearWorkspacePlugins(cmd *cobra.Command) error {
	// Get workspace
	workspace, _, err := getWorkspaceForPlugins(cmd, setNvimAppFlag, setNvimWorkspaceFlag)
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

// runSetGlobalDefaultPlugins sets or clears the global default plugins using the defaults table
func runSetGlobalDefaultPlugins(cmd *cobra.Command, args []string) error {
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Handle clear operation
	if setNvimClearFlag {
		return clearGlobalDefaultPlugins(cmd, ds)
	}

	// Get available plugins for validation and --all flag
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
	var availablePlugins []string
	for _, p := range allPlugins {
		if p.Enabled {
			availablePlugins = append(availablePlugins, p.Name)
		}
	}

	if len(availablePlugins) == 0 {
		render.Error("No plugins found in database")
		render.Info("Import plugins first with: nvp library install --all")
		return nil
	}

	// Determine which plugins to set as defaults
	var targetPlugins []string
	if setNvimAllFlag {
		targetPlugins = availablePlugins
	} else if len(args) == 0 {
		return fmt.Errorf("specify plugin names, use --all, or use --clear")
	} else {
		targetPlugins = args
	}

	// Validate that all specified plugins exist
	var notFound []string
	pluginMap := make(map[string]bool)
	for _, p := range availablePlugins {
		pluginMap[p] = true
	}

	for _, pluginName := range targetPlugins {
		if !pluginMap[pluginName] {
			notFound = append(notFound, pluginName)
		}
	}

	if len(notFound) > 0 {
		render.Error(fmt.Sprintf("Plugin(s) not found in library (%d):", len(notFound)))
		for _, p := range notFound {
			fmt.Printf("  ? %s\n", p)
		}
		render.Info("Install missing plugins with: nvp library install <name>")
		return fmt.Errorf("invalid plugin names provided")
	}

	// Get previous global default plugins
	previousPluginsJSON, err := ds.GetDefault("plugins")
	if err != nil {
		return fmt.Errorf("failed to get previous global default plugins: %w", err)
	}

	var previousPlugins []string
	if previousPluginsJSON != "" {
		// Parse existing JSON array
		if err := json.Unmarshal([]byte(previousPluginsJSON), &previousPlugins); err != nil {
			// If parsing fails, treat as empty
			previousPlugins = []string{}
		}
	}

	// Handle dry run
	if setNvimPluginDryRun {
		result := map[string]interface{}{
			"level":           "global",
			"objectName":      "global-defaults",
			"plugins":         targetPlugins,
			"previousPlugins": previousPlugins,
			"operation":       "set-global-default-plugins",
		}

		opts := render.Options{
			Type:  render.TypeKeyValue,
			Title: "Global Default Plugins Set (dry-run)",
		}

		return render.OutputWith(setNvimPluginOutput, result, opts)
	}

	// Convert plugins list to JSON for storage
	pluginsJSON, err := json.Marshal(targetPlugins)
	if err != nil {
		return fmt.Errorf("failed to encode plugins as JSON: %w", err)
	}

	// Set the new global default plugins
	if err := ds.SetDefault("plugins", string(pluginsJSON)); err != nil {
		return fmt.Errorf("failed to set global default plugins: %w", err)
	}

	// Report results
	render.Success(fmt.Sprintf("Set %d plugin(s) as global defaults:", len(targetPlugins)))
	for _, p := range targetPlugins {
		fmt.Printf("  + %s\n", p)
	}

	if len(previousPlugins) > 0 {
		fmt.Println()
		render.Info(fmt.Sprintf("Replaced %d previous default(s):", len(previousPlugins)))
		for _, p := range previousPlugins {
			fmt.Printf("  - %s\n", p)
		}
	}

	fmt.Println()
	render.Info("These plugins will be used as defaults when creating new workspaces")

	return nil
}

// clearGlobalDefaultPlugins removes the global default plugins setting
func clearGlobalDefaultPlugins(cmd *cobra.Command, ds db.DataStore) error {
	// Get current defaults for reporting
	currentPluginsJSON, err := ds.GetDefault("plugins")
	if err != nil {
		return fmt.Errorf("failed to get current global default plugins: %w", err)
	}

	var currentPlugins []string
	if currentPluginsJSON != "" {
		if err := json.Unmarshal([]byte(currentPluginsJSON), &currentPlugins); err != nil {
			// If parsing fails, treat as empty
			currentPlugins = []string{}
		}
	}

	// Handle dry run
	if setNvimPluginDryRun {
		result := map[string]interface{}{
			"level":           "global",
			"objectName":      "global-defaults",
			"plugins":         []string{},
			"previousPlugins": currentPlugins,
			"operation":       "clear-global-default-plugins",
		}

		opts := render.Options{
			Type:  render.TypeKeyValue,
			Title: "Global Default Plugins Cleared (dry-run)",
		}

		return render.OutputWith(setNvimPluginOutput, result, opts)
	}

	// Clear the global defaults by deleting the key
	if err := ds.DeleteDefault("plugins"); err != nil {
		return fmt.Errorf("failed to clear global default plugins: %w", err)
	}

	if len(currentPlugins) == 0 {
		render.Info("No global default plugins were set")
		return nil
	}

	render.Success(fmt.Sprintf("Cleared %d global default plugin(s):", len(currentPlugins)))
	for _, p := range currentPlugins {
		fmt.Printf("  - %s\n", p)
	}

	fmt.Println()
	render.Info("New workspaces will now include all available plugins")

	return nil
}
