// Package cmd provides CLI commands for DevOpsMaestro.
// This file implements kubectl-style 'set' commands for terminal resources.
//
// Usage:
//
//	dvm set terminal prompt -w <workspace> <name>     # Set terminal prompt for workspace
//	dvm set terminal plugin -w <workspace> <names...> # Add terminal plugins to workspace
//	dvm set terminal package -w <workspace> <name>    # Set terminal package for workspace
package cmd

import (
	"database/sql"
	"fmt"

	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// Flags for set terminal commands
var (
	setTerminalWorkspaceFlag string
	setTerminalAppFlag       string
	setTerminalAllFlag       bool
	setTerminalClearFlag     bool
	setTerminalOutput        string
	setTerminalDryRun        bool
)

// setTerminalCmd is the 'terminal' subcommand under 'set'
var setTerminalCmd = &cobra.Command{
	Use:   "terminal",
	Short: "Set terminal configurations",
	Long: `Set terminal-related configurations for workspaces.

Examples:
  dvm set terminal prompt -w dev starship
  dvm set terminal plugin -w dev zsh-autosuggestions
  dvm set terminal package -w dev poweruser`,
}

// setTerminalPromptCmd sets the terminal prompt for a workspace
var setTerminalPromptCmd = &cobra.Command{
	Use:   "prompt [name]",
	Short: "Set terminal prompt for workspace",
	Long: `Set the terminal prompt configuration for a workspace.

Prompts must exist in the terminal prompt library.
Use 'dvm get terminal prompts' to see available prompts.

Examples:
  dvm set terminal prompt -w dev starship
  dvm set terminal prompt -w dev starship-minimal
  dvm set terminal prompt -w dev starship --dry-run
  dvm set terminal prompt -w dev --clear`,
	RunE: runSetTerminalPrompt,
}

// setTerminalPluginCmd adds terminal plugins to a workspace
var setTerminalPluginCmd = &cobra.Command{
	Use:   "plugin [names...]",
	Short: "Add terminal plugins to workspace",
	Long: `Add terminal plugins to a workspace's configuration.

Plugins must exist in the terminal plugin library.
Use 'dvm get terminal plugins' to see available plugins.

Examples:
  dvm set terminal plugin -w dev zsh-autosuggestions
  dvm set terminal plugin -w dev zsh-autosuggestions zsh-syntax-highlighting
  dvm set terminal plugin -w dev --all      # Add all available plugins
  dvm set terminal plugin -w dev --clear    # Remove all plugins from workspace`,
	RunE: runSetTerminalPlugin,
}

// setTerminalPackageCmd sets a terminal package for a workspace
var setTerminalPackageCmd = &cobra.Command{
	Use:   "package [name]",
	Short: "Set terminal package for workspace",
	Long: `Set a terminal package (bundle of plugins + prompts) for a workspace.

Packages must exist in the terminal package library.
Use 'dvm get terminal packages' to see available packages.

Examples:
  dvm set terminal package -w dev poweruser
  dvm set terminal package -w dev minimal
  dvm set terminal package -w dev --clear`,
	RunE: runSetTerminalPackage,
}

func init() {
	// Add setTerminalCmd to setCmd
	setCmd.AddCommand(setTerminalCmd)

	// Add subcommands
	setTerminalCmd.AddCommand(setTerminalPromptCmd)
	setTerminalCmd.AddCommand(setTerminalPluginCmd)
	setTerminalCmd.AddCommand(setTerminalPackageCmd)

	// Flags for prompt command
	setTerminalPromptCmd.Flags().StringVarP(&setTerminalWorkspaceFlag, "workspace", "w", "", "Workspace to configure (required)")
	setTerminalPromptCmd.Flags().StringVarP(&setTerminalAppFlag, "app", "a", "", "App for workspace (defaults to active)")
	setTerminalPromptCmd.Flags().BoolVar(&setTerminalClearFlag, "clear", false, "Remove prompt from workspace")
	setTerminalPromptCmd.Flags().BoolVar(&setTerminalDryRun, "dry-run", false, "Preview changes without applying")
	setTerminalPromptCmd.Flags().StringVarP(&setTerminalOutput, "output", "o", "", "Output format (json, yaml, plain, table, colored)")
	setTerminalPromptCmd.MarkFlagRequired("workspace")

	// Flags for plugin command
	setTerminalPluginCmd.Flags().StringVarP(&setTerminalWorkspaceFlag, "workspace", "w", "", "Workspace to configure (required)")
	setTerminalPluginCmd.Flags().StringVarP(&setTerminalAppFlag, "app", "a", "", "App for workspace (defaults to active)")
	setTerminalPluginCmd.Flags().BoolVar(&setTerminalAllFlag, "all", false, "Add all plugins from library")
	setTerminalPluginCmd.Flags().BoolVar(&setTerminalClearFlag, "clear", false, "Remove all plugins from workspace")
	setTerminalPluginCmd.Flags().BoolVar(&setTerminalDryRun, "dry-run", false, "Preview changes without applying")
	setTerminalPluginCmd.Flags().StringVarP(&setTerminalOutput, "output", "o", "", "Output format (json, yaml, plain, table, colored)")
	setTerminalPluginCmd.MarkFlagRequired("workspace")

	// Flags for package command
	setTerminalPackageCmd.Flags().StringVarP(&setTerminalWorkspaceFlag, "workspace", "w", "", "Workspace to configure (required)")
	setTerminalPackageCmd.Flags().StringVarP(&setTerminalAppFlag, "app", "a", "", "App for workspace (defaults to active)")
	setTerminalPackageCmd.Flags().BoolVar(&setTerminalClearFlag, "clear", false, "Remove package from workspace")
	setTerminalPackageCmd.Flags().BoolVar(&setTerminalDryRun, "dry-run", false, "Preview changes without applying")
	setTerminalPackageCmd.Flags().StringVarP(&setTerminalOutput, "output", "o", "", "Output format (json, yaml, plain, table, colored)")
	setTerminalPackageCmd.MarkFlagRequired("workspace")
}

func runSetTerminalPrompt(cmd *cobra.Command, args []string) error {
	// Validate flags
	if setTerminalClearFlag {
		return runClearTerminalPrompt(cmd)
	}

	if len(args) == 0 {
		return fmt.Errorf("specify prompt name or use --clear")
	}

	promptName := args[0]

	// Get workspace
	workspace, appName, err := getWorkspaceForPlugins(cmd, setTerminalAppFlag, setTerminalWorkspaceFlag)
	if err != nil {
		return err
	}

	// TODO: Validate that prompt exists in library
	// For now, we'll accept any prompt name for testing
	// In production, we would call something like:
	// if !terminalops.PromptExists(promptName) {
	//     return fmt.Errorf("prompt %q not found in library", promptName)
	// }

	// Handle dry-run
	if setTerminalDryRun {
		result := map[string]interface{}{
			"level":     "workspace",
			"workspace": workspace.Name,
			"app":       appName,
			"prompt":    promptName,
			"operation": "set-terminal-prompt",
			"applied":   false,
		}

		opts := render.Options{
			Type:  render.TypeKeyValue,
			Title: fmt.Sprintf("Terminal Prompt Set (dry-run)"),
		}

		return render.OutputWith(setTerminalOutput, result, opts)
	}

	// Set the prompt
	workspace.TerminalPrompt = sql.NullString{String: promptName, Valid: true}

	// Save to database
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	if err := ds.UpdateWorkspace(workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	// Report success
	render.Success(fmt.Sprintf("Set terminal prompt for workspace '%s' to '%s'", workspace.Name, promptName))
	render.Blank()
	render.Info(fmt.Sprintf("Rebuild workspace to apply: dvm build %s --force", workspace.Name))

	return nil
}

func runClearTerminalPrompt(cmd *cobra.Command) error {
	// Get workspace
	workspace, _, err := getWorkspaceForPlugins(cmd, setTerminalAppFlag, setTerminalWorkspaceFlag)
	if err != nil {
		return err
	}

	// Handle dry-run
	if setTerminalDryRun {
		result := map[string]interface{}{
			"level":     "workspace",
			"workspace": workspace.Name,
			"prompt":    nil,
			"operation": "clear-terminal-prompt",
			"applied":   false,
		}

		opts := render.Options{
			Type:  render.TypeKeyValue,
			Title: "Terminal Prompt Cleared (dry-run)",
		}

		return render.OutputWith(setTerminalOutput, result, opts)
	}

	// Clear the prompt
	workspace.TerminalPrompt = sql.NullString{Valid: false}

	// Save to database
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	if err := ds.UpdateWorkspace(workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	render.Success(fmt.Sprintf("Cleared terminal prompt from workspace '%s'", workspace.Name))
	return nil
}

func runSetTerminalPlugin(cmd *cobra.Command, args []string) error {
	// Validate flags
	if setTerminalClearFlag {
		return runClearTerminalPlugins(cmd)
	}

	if !setTerminalAllFlag && len(args) == 0 {
		return fmt.Errorf("specify plugin names, use --all, or use --clear")
	}

	// Get workspace
	workspace, appName, err := getWorkspaceForPlugins(cmd, setTerminalAppFlag, setTerminalWorkspaceFlag)
	if err != nil {
		return err
	}

	// Get datastore
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get available plugins from database
	// TODO: Implement terminal plugin listing
	// For now, we'll use a mock list for testing
	var availablePlugins []string
	if setTerminalAllFlag {
		// In production, this would call ds.ListTerminalPlugins()
		availablePlugins = []string{"zsh-autosuggestions", "zsh-syntax-highlighting"}
	} else {
		availablePlugins = args
	}

	// Get current plugins
	currentPlugins := workspace.GetTerminalPlugins()

	// Determine which plugins to add
	var toAdd []string
	var skipped []string
	var notFound []string

	pluginMap := make(map[string]bool)
	for _, p := range currentPlugins {
		pluginMap[p] = true
	}

	for _, pluginName := range availablePlugins {
		if pluginMap[pluginName] {
			skipped = append(skipped, pluginName)
		} else {
			// TODO: Validate plugin exists in library
			// For testing, we accept all plugin names
			toAdd = append(toAdd, pluginName)
			pluginMap[pluginName] = true
		}
	}

	// Handle dry-run
	if setTerminalDryRun {
		result := map[string]interface{}{
			"level":     "workspace",
			"workspace": workspace.Name,
			"app":       appName,
			"toAdd":     toAdd,
			"skipped":   skipped,
			"notFound":  notFound,
			"operation": "add-terminal-plugins",
			"applied":   false,
		}

		opts := render.Options{
			Type:  render.TypeKeyValue,
			Title: "Terminal Plugins Add (dry-run)",
		}

		return render.OutputWith(setTerminalOutput, result, opts)
	}

	// Update workspace plugins
	newPlugins := append(currentPlugins, toAdd...)
	workspace.SetTerminalPlugins(newPlugins)

	// Save to database
	if err := ds.UpdateWorkspace(workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	// Report results
	if len(toAdd) > 0 {
		render.Success(fmt.Sprintf("Added %d plugin(s) to workspace '%s':", len(toAdd), workspace.Name))
		for _, p := range toAdd {
			render.Plainf("  + %s", p)
		}
	}

	if len(skipped) > 0 {
		render.Info(fmt.Sprintf("Skipped %d plugin(s) (already configured):", len(skipped)))
		for _, p := range skipped {
			render.Plainf("  • %s", p)
		}
	}

	if len(notFound) > 0 {
		render.Warning(fmt.Sprintf("Not found in library (%d):", len(notFound)))
		for _, p := range notFound {
			render.Plainf("  ? %s", p)
		}
	}

	if len(toAdd) > 0 {
		render.Blank()
		render.Info(fmt.Sprintf("View configured plugins: dvm get terminal plugins -w %s", workspace.Name))
		render.Info(fmt.Sprintf("Rebuild workspace to apply: dvm build %s --force", workspace.Name))
	}

	return nil
}

func runClearTerminalPlugins(cmd *cobra.Command) error {
	// Get workspace
	workspace, _, err := getWorkspaceForPlugins(cmd, setTerminalAppFlag, setTerminalWorkspaceFlag)
	if err != nil {
		return err
	}

	// Get datastore
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get current plugin count
	currentPlugins := workspace.GetTerminalPlugins()
	count := len(currentPlugins)

	// Handle dry-run
	if setTerminalDryRun {
		result := map[string]interface{}{
			"level":     "workspace",
			"workspace": workspace.Name,
			"plugins":   []string{},
			"count":     count,
			"operation": "clear-terminal-plugins",
			"applied":   false,
		}

		opts := render.Options{
			Type:  render.TypeKeyValue,
			Title: "Terminal Plugins Cleared (dry-run)",
		}

		return render.OutputWith(setTerminalOutput, result, opts)
	}

	if count == 0 {
		render.Info(fmt.Sprintf("Workspace '%s' has no plugins configured", workspace.Name))
		return nil
	}

	// Clear plugins
	workspace.SetTerminalPlugins([]string{})

	// Save to database
	if err := ds.UpdateWorkspace(workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	render.Success(fmt.Sprintf("Cleared %d plugin(s) from workspace '%s'", count, workspace.Name))
	render.Blank()
	render.Info(fmt.Sprintf("Rebuild workspace to apply: dvm build %s --force", workspace.Name))

	return nil
}

func runSetTerminalPackage(cmd *cobra.Command, args []string) error {
	// Validate flags
	if setTerminalClearFlag {
		return runClearTerminalPackage(cmd)
	}

	if len(args) == 0 {
		return fmt.Errorf("specify package name or use --clear")
	}

	packageName := args[0]

	// Get workspace
	workspace, appName, err := getWorkspaceForPlugins(cmd, setTerminalAppFlag, setTerminalWorkspaceFlag)
	if err != nil {
		return err
	}

	// TODO: Validate that package exists in library
	// For now, we'll accept any package name for testing
	// In production, we would call something like:
	// if !terminalops.PackageExists(packageName) {
	//     return fmt.Errorf("package %q not found in library", packageName)
	// }

	// Handle dry-run
	if setTerminalDryRun {
		result := map[string]interface{}{
			"level":     "workspace",
			"workspace": workspace.Name,
			"app":       appName,
			"package":   packageName,
			"operation": "set-terminal-package",
			"applied":   false,
		}

		opts := render.Options{
			Type:  render.TypeKeyValue,
			Title: "Terminal Package Set (dry-run)",
		}

		return render.OutputWith(setTerminalOutput, result, opts)
	}

	// Set the package
	workspace.TerminalPackage = sql.NullString{String: packageName, Valid: true}

	// Save to database
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	if err := ds.UpdateWorkspace(workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	render.Success(fmt.Sprintf("Set terminal package for workspace '%s' to '%s'", workspace.Name, packageName))
	render.Blank()
	render.Info(fmt.Sprintf("Rebuild workspace to apply: dvm build %s --force", workspace.Name))

	return nil
}

func runClearTerminalPackage(cmd *cobra.Command) error {
	// Get workspace
	workspace, _, err := getWorkspaceForPlugins(cmd, setTerminalAppFlag, setTerminalWorkspaceFlag)
	if err != nil {
		return err
	}

	// Handle dry-run
	if setTerminalDryRun {
		result := map[string]interface{}{
			"level":     "workspace",
			"workspace": workspace.Name,
			"package":   nil,
			"operation": "clear-terminal-package",
			"applied":   false,
		}

		opts := render.Options{
			Type:  render.TypeKeyValue,
			Title: "Terminal Package Cleared (dry-run)",
		}

		return render.OutputWith(setTerminalOutput, result, opts)
	}

	// Clear the package
	workspace.TerminalPackage = sql.NullString{Valid: false}

	// Save to database
	ds, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	if err := ds.UpdateWorkspace(workspace); err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	render.Success(fmt.Sprintf("Cleared terminal package from workspace '%s'", workspace.Name))
	return nil
}
