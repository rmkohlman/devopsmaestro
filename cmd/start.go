package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/render"
	"fmt"

	"github.com/spf13/cobra"
)

// startCmd is the parent command for starting resources
var startCmd = &cobra.Command{
	Use:   "start [resource]",
	Short: "Start resources (registry, app, workspace)",
	Long: `Start various resources in the DevOpsMaestro system.

This command follows kubectl patterns where you specify the resource type
as a subcommand, then provide the resource name as an argument.

Available resources:
  registry    Start a registry instance

Examples:
  # Start a registry
  dvm start registry my-registry

  # Start with foreground mode (future)
  dvm start registry my-registry --foreground`,
}

// startRegistryCmd starts a registry
var startRegistryCmd = &cobra.Command{
	Use:     "registry <name>",
	Aliases: []string{"reg"},
	Short:   "Start a registry",
	Long: `Start a local OCI registry for image caching.

The registry will be started according to the configured lifecycle mode:
  - persistent: Stays running until explicitly stopped
  - on-demand: Stops automatically after idle timeout
  - manual: Stays running until explicitly stopped

The Zot binary will be automatically downloaded on first use.

Examples:
  # Start a registry by name
  dvm start registry my-registry

  # Start with a specific port (future)
  dvm start registry my-registry --port 5002

  # Run in foreground for debugging (future)
  dvm start registry my-registry --foreground`,
	Args: cobra.ExactArgs(1),
	RunE: runStartRegistry,
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.AddCommand(startRegistryCmd)

	// Future flags (not yet implemented)
	// startRegistryCmd.Flags().Int("port", 0, "Port to run on")
	// startRegistryCmd.Flags().Bool("foreground", false, "Run in foreground")
}

// runStartRegistry implements the start registry command
func runStartRegistry(cmd *cobra.Command, args []string) error {
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

	// Check if already running
	if mgr.IsRunning(ctx) {
		render.Info(fmt.Sprintf("Registry '%s' already running", name))
		render.Info(fmt.Sprintf("Endpoint: %s", mgr.GetEndpoint()))
		return nil
	}

	// Start the registry
	render.Progress(fmt.Sprintf("Starting registry '%s'...", name))
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start registry: %w", err)
	}

	// Update DB status to running
	reg.Status = "running"
	if err := store.UpdateRegistry(reg); err != nil {
		render.Warning(fmt.Sprintf("Registry started but failed to update status: %v", err))
	}

	render.Success(fmt.Sprintf("Registry '%s' started", name))
	render.Info(fmt.Sprintf("Endpoint: %s", mgr.GetEndpoint()))

	return nil
}
