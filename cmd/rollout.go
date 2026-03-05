package cmd

import (
	"context"
	"database/sql"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/render"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// rolloutCmd is the parent command for rollout operations
var rolloutCmd = &cobra.Command{
	Use:   "rollout",
	Short: "Manage rollouts (restart, status, history, undo)",
	Long: `Manage resource rollouts following kubectl patterns.

Supports restart, status, history, and undo operations for:
  - registry

Examples:
  # Restart a registry
  dvm rollout restart registry my-registry

  # Check rollout status
  dvm rollout status registry my-registry

  # View rollout history
  dvm rollout history registry my-registry

  # Rollback to previous version
  dvm rollout undo registry my-registry`,
}

// ========== RESTART Subcommands ==========

var rolloutRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart a resource",
	Long:  `Restart a resource (stops then starts it).`,
}

var rolloutRestartRegistryCmd = &cobra.Command{
	Use:     "registry <name>",
	Aliases: []string{"reg"},
	Short:   "Restart a registry",
	Long: `Restart a registry by stopping it and starting it again.

This creates a new revision entry in the rollout history.

Examples:
  # Restart a registry
  dvm rollout restart registry my-registry
  dvm rollout restart reg my-registry`,
	Args: cobra.ExactArgs(1),
	RunE: runRolloutRestartRegistry,
}

// ========== STATUS Subcommands ==========

var rolloutStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show rollout status",
	Long:  `Show the current rollout status of a resource.`,
}

var rolloutStatusRegistryCmd = &cobra.Command{
	Use:     "registry <name>",
	Aliases: []string{"reg"},
	Short:   "Show registry rollout status",
	Long: `Show the current rollout status of a registry.

Displays the current revision, configuration, and runtime status.

Output formats:
  table (default) - Human-readable table
  yaml            - YAML output
  json            - JSON output

Examples:
  # Show status
  dvm rollout status registry my-registry

  # Show status as JSON
  dvm rollout status registry my-registry -o json`,
	Args: cobra.ExactArgs(1),
	RunE: runRolloutStatusRegistry,
}

// ========== HISTORY Subcommands ==========

var rolloutHistoryCmd = &cobra.Command{
	Use:   "history",
	Short: "Show rollout history",
	Long:  `Show the rollout history for a resource.`,
}

var rolloutHistoryRegistryCmd = &cobra.Command{
	Use:     "registry <name>",
	Aliases: []string{"reg"},
	Short:   "Show registry rollout history",
	Long: `Show the rollout history for a registry.

Displays all past revisions with actions, status, and timestamps.

Output formats:
  table (default) - Human-readable table
  yaml            - YAML output
  json            - JSON output

Examples:
  # Show history
  dvm rollout history registry my-registry

  # Show history as JSON
  dvm rollout history registry my-registry -o json`,
	Args: cobra.ExactArgs(1),
	RunE: runRolloutHistoryRegistry,
}

// ========== UNDO Subcommands ==========

var rolloutUndoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Rollback to previous version",
	Long:  `Rollback a resource to its previous version.`,
}

var rolloutUndoRegistryCmd = &cobra.Command{
	Use:     "registry <name>",
	Aliases: []string{"reg"},
	Short:   "Rollback a registry to previous version",
	Long: `Rollback a registry to its previous successful revision.

This restores the configuration from the previous revision and restarts
the registry with that configuration.

Examples:
  # Undo last rollout
  dvm rollout undo registry my-registry
  dvm rollout undo reg my-registry`,
	Args: cobra.ExactArgs(1),
	RunE: runRolloutUndoRegistry,
}

// ========== Init Function ==========

func init() {
	// Register rollout as top-level command
	rootCmd.AddCommand(rolloutCmd)

	// Add subcommands to rollout
	rolloutCmd.AddCommand(rolloutRestartCmd)
	rolloutCmd.AddCommand(rolloutStatusCmd)
	rolloutCmd.AddCommand(rolloutHistoryCmd)
	rolloutCmd.AddCommand(rolloutUndoCmd)

	// Add resource-specific commands to each subcommand
	rolloutRestartCmd.AddCommand(rolloutRestartRegistryCmd)
	rolloutStatusCmd.AddCommand(rolloutStatusRegistryCmd)
	rolloutHistoryCmd.AddCommand(rolloutHistoryRegistryCmd)
	rolloutUndoCmd.AddCommand(rolloutUndoRegistryCmd)

	// Add output format flags to status and history
	rolloutStatusRegistryCmd.Flags().StringP("output", "o", "", "Output format (json, yaml, table)")
	rolloutHistoryRegistryCmd.Flags().StringP("output", "o", "", "Output format (json, yaml, table)")
}

// ========== Implementation Functions ==========

// runRolloutRestartRegistry implements the rollout restart registry command
func runRolloutRestartRegistry(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	name := args[0]

	// Get DataStore from context
	dataStore, ok := ctx.Value("dataStore").(*db.DataStore)
	if !ok || dataStore == nil {
		return fmt.Errorf("database not initialized")
	}
	store := *dataStore

	// Look up Registry by name
	reg, err := store.GetRegistryByName(name)
	if err != nil {
		return fmt.Errorf("registry '%s' not found: %w", name, err)
	}

	// Create ServiceManager via factory
	factory := registry.NewServiceFactory()
	mgr, err := factory.CreateManager(reg)
	if err != nil {
		return fmt.Errorf("failed to create registry manager: %w", err)
	}

	// Record restart start
	render.Progress(fmt.Sprintf("Restarting registry '%s'...", name))

	// Stop if running
	if mgr.IsRunning(ctx) {
		if err := mgr.Stop(ctx); err != nil {
			return fmt.Errorf("failed to stop registry: %w", err)
		}
		// Update DB status for the stop phase
		reg.Status = "stopped"
		_ = store.UpdateRegistry(reg)
		// Small delay to ensure clean shutdown
		time.Sleep(500 * time.Millisecond)
	}

	// Start the registry
	if err := mgr.Start(ctx); err != nil {
		// Record failure in history
		_ = store.CreateRegistryHistory(&models.RegistryHistory{
			RegistryID: reg.ID,
			Action:     "restart",
			Status:     "failed",
			Config:     mustMarshalJSON(reg),
			CreatedAt:  time.Now(),
		})
		return fmt.Errorf("failed to start registry: %w", err)
	}

	// Update DB status to running
	reg.Status = "running"
	if err := store.UpdateRegistry(reg); err != nil {
		render.Warning(fmt.Sprintf("Registry restarted but failed to update status: %v", err))
	}

	// Record successful restart in history
	if err := store.CreateRegistryHistory(&models.RegistryHistory{
		RegistryID: reg.ID,
		Action:     "restart",
		Status:     "success",
		Config:     mustMarshalJSON(reg),
		CreatedAt:  time.Now(),
		CompletedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}); err != nil {
		// Log but don't fail the command
		render.Warning(fmt.Sprintf("Failed to record restart in history: %v", err))
	}

	render.Success(fmt.Sprintf("Registry '%s' restarted", name))
	render.Info(fmt.Sprintf("Endpoint: %s", mgr.GetEndpoint()))
	return nil
}

// runRolloutStatusRegistry implements the rollout status registry command
func runRolloutStatusRegistry(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	name := args[0]
	format, _ := cmd.Flags().GetString("output")

	// Get DataStore from context
	dataStore, ok := ctx.Value("dataStore").(*db.DataStore)
	if !ok || dataStore == nil {
		return fmt.Errorf("database not initialized")
	}
	store := *dataStore

	// Look up Registry by name
	reg, err := store.GetRegistryByName(name)
	if err != nil {
		return fmt.Errorf("registry '%s' not found: %w", name, err)
	}

	// Create ServiceManager to get runtime status
	factory := registry.NewServiceFactory()
	mgr, err := factory.CreateManager(reg)
	if err != nil {
		return fmt.Errorf("failed to create registry manager: %w", err)
	}

	// Get latest history entry
	history, err := store.ListRegistryHistory(reg.ID)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	// Build status data
	status := map[string]interface{}{
		"name":     reg.Name,
		"enabled":  reg.Enabled,
		"running":  mgr.IsRunning(ctx),
		"endpoint": mgr.GetEndpoint(),
		"config": map[string]interface{}{
			"lifecycle":    reg.Lifecycle,
			"port":         reg.Port,
			"storage":      reg.Storage,
			"idle_timeout": reg.IdleTimeout,
		},
	}

	if len(history) > 0 {
		latest := history[0]
		status["latest_revision"] = latest.Revision
		status["latest_action"] = latest.Action
		status["latest_status"] = latest.Status
		status["last_updated"] = latest.CreatedAt
	}

	// Output based on format
	return outputData(ctx, format, status)
}

// runRolloutHistoryRegistry implements the rollout history registry command
func runRolloutHistoryRegistry(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	name := args[0]
	format, _ := cmd.Flags().GetString("output")

	// Get DataStore from context
	dataStore, ok := ctx.Value("dataStore").(*db.DataStore)
	if !ok || dataStore == nil {
		return fmt.Errorf("database not initialized")
	}
	store := *dataStore

	// Look up Registry by name
	reg, err := store.GetRegistryByName(name)
	if err != nil {
		return fmt.Errorf("registry '%s' not found: %w", name, err)
	}

	// Get history
	history, err := store.ListRegistryHistory(reg.ID)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	if len(history) == 0 {
		render.Info(fmt.Sprintf("No rollout history for registry '%s'", name))
		return nil
	}

	// Output based on format
	return outputData(ctx, format, history)
}

// runRolloutUndoRegistry implements the rollout undo registry command
func runRolloutUndoRegistry(cmd *cobra.Command, args []string) error {
	// TODO: Implement rollback functionality
	// This requires:
	// 1. Get previous successful revision from history
	// 2. Restore that configuration (update registry record)
	// 3. Restart with old config
	// 4. Record rollback in history
	return fmt.Errorf("rollout undo is not yet implemented")
}

// ========== Helper Functions ==========

// outputData handles output formatting for status and history commands
func outputData(ctx context.Context, format string, data interface{}) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(render.GetWriter())
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)
	case "yaml":
		encoder := yaml.NewEncoder(render.GetWriter())
		encoder.SetIndent(2)
		return encoder.Encode(data)
	case "table", "":
		// Use render.OutputWith for default table formatting
		return render.OutputWith(format, data, render.Options{})
	default:
		return fmt.Errorf("unsupported output format: %s (use json, yaml, or table)", format)
	}
}

// mustMarshalJSON marshals to JSON or returns empty string on error
func mustMarshalJSON(v interface{}) string {
	data, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(data)
}
