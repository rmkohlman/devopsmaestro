package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/render"
	"fmt"

	"github.com/spf13/cobra"
)

// stopCmd is the parent command for stopping resources
var stopCmd = &cobra.Command{
	Use:   "stop [resource]",
	Short: "Stop resources (registry, app, workspace)",
	Long: `Stop various resources in the DevOpsMaestro system.

This command follows kubectl patterns where you specify the resource type
as a subcommand, then provide the resource name as an argument.

Available resources:
  registry    Stop a registry instance

Examples:
  # Stop a registry gracefully
  dvm stop registry my-registry

  # Force stop a registry
  dvm stop registry my-registry --force`,
}

// stopRegistryCmd stops a registry
var stopRegistryCmd = &cobra.Command{
	Use:     "registry <name>",
	Aliases: []string{"reg"},
	Short:   "Stop a registry",
	Long: `Stop a local OCI registry gracefully.

By default, sends SIGTERM and waits for graceful shutdown.
Use --force to send SIGKILL immediately.

Examples:
  # Graceful stop
  dvm stop registry my-registry

  # Force stop with SIGKILL
  dvm stop registry my-registry --force`,
	Args: cobra.ExactArgs(1),
	RunE: runStopRegistry,
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.AddCommand(stopRegistryCmd)

	// Stop command flags
	stopRegistryCmd.Flags().Bool("force", false, "Force kill (SIGKILL instead of graceful shutdown)")
}

// runStopRegistry implements the stop registry command
func runStopRegistry(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	name := args[0] // Get name from positional arg (Args validator ensures it exists)

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

	// Check if running
	if !mgr.IsRunning(ctx) {
		render.Info(fmt.Sprintf("Registry '%s' is not running", name))
		return nil
	}

	// Stop the registry
	force, _ := cmd.Flags().GetBool("force")
	if force {
		render.Progress(fmt.Sprintf("Force stopping registry '%s'...", name))
	} else {
		render.Progress(fmt.Sprintf("Stopping registry '%s'...", name))
	}

	if err := mgr.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop registry: %w", err)
	}

	render.Success(fmt.Sprintf("Registry '%s' stopped", name))
	return nil
}
