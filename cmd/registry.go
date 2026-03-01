package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/render"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// registryCmd is the parent command for registry operations
var registryCmd = &cobra.Command{
	Use:     "registry",
	Aliases: []string{"reg"},
	Short:   "Manage the local OCI registry",
	Long: `Manage the local Zot OCI registry used for image caching.

The registry provides pull-through caching for container images, reducing
network usage and improving build times. It can also store locally built
workspace images.

Examples:
  # Start the registry
  dvm registry start

  # Check registry status
  dvm registry status

  # View recent logs
  dvm registry logs

  # Stop the registry
  dvm registry stop

  # Clean up old images
  dvm registry prune --older-than 30d`,
}

// registryStartCmd starts the registry
var registryStartCmd = &cobra.Command{
	Use:   "start <name>",
	Args:  cobra.ExactArgs(1),
	Short: "Start the local registry",
	Long: `Start the local Zot OCI registry for image caching.

The registry will be started according to the configured lifecycle mode:
  - persistent: Stays running until explicitly stopped
  - on-demand: Stops automatically after idle timeout
  - manual: Stays running until explicitly stopped

The Zot binary will be automatically downloaded on first use.

Examples:
  # Start with default settings
  dvm registry start

  # Start on a different port
  dvm registry start --port 5002

  # Run in foreground (for debugging)
  dvm registry start --foreground`,
	RunE: runRegistryStart,
}

// registryStopCmd stops the registry
var registryStopCmd = &cobra.Command{
	Use:   "stop <name>",
	Args:  cobra.ExactArgs(1),
	Short: "Stop the local registry",
	Long: `Stop the local Zot OCI registry gracefully.

By default, sends SIGTERM and waits for graceful shutdown.
Use --force to send SIGKILL immediately.

Examples:
  # Graceful stop
  dvm registry stop

  # Force stop
  dvm registry stop --force`,
	RunE: runRegistryStop,
}

// registryStatusCmd shows registry status
var registryStatusCmd = &cobra.Command{
	Use:   "status [name]",
	Args:  cobra.MaximumNArgs(1),
	Short: "Show registry status",
	Long: `Show the current status of the local OCI registry.

Output formats:
  table (default) - Human-readable table
  wide            - Table with additional columns
  yaml            - YAML output
  json            - JSON output

Examples:
  # Default table output
  dvm registry status

  # Wide output with extra columns
  dvm registry status -o wide

  # JSON output for scripting
  dvm registry status -o json`,
	RunE: runRegistryStatus,
}

// registryLogsCmd shows registry logs
var registryLogsCmd = &cobra.Command{
	Use:   "logs",
	Short: "View registry logs",
	Long: `View recent logs from the Zot registry.

Examples:
  # Show last 50 lines (default)
  dvm registry logs

  # Show last 100 lines
  dvm registry logs -n 100

  # Show logs from the last 2 hours
  dvm registry logs --since 2h`,
	RunE: runRegistryLogs,
}

// registryPruneCmd cleans up registry images
var registryPruneCmd = &cobra.Command{
	Use:   "prune",
	Short: "Remove cached images",
	Long: `Remove cached images from the local registry.

By default, prompts for confirmation before removing images.
Use --force to skip the confirmation prompt.

Examples:
  # Remove images older than 30 days
  dvm registry prune --older-than 30d

  # Show what would be removed without removing
  dvm registry prune --all --dry-run

  # Remove all cached images without confirmation
  dvm registry prune --all --force`,
	RunE: runRegistryPrune,
}

func init() {
	// Register registry command to root
	rootCmd.AddCommand(registryCmd)

	// Register subcommands
	registryCmd.AddCommand(registryStartCmd)
	registryCmd.AddCommand(registryStopCmd)
	registryCmd.AddCommand(registryStatusCmd)
	registryCmd.AddCommand(registryLogsCmd)
	registryCmd.AddCommand(registryPruneCmd)

	// Start command flags
	registryStartCmd.Flags().Int("port", 0, "Port to run on (default: from config or 5001)")
	registryStartCmd.Flags().Bool("foreground", false, "Run in foreground (don't daemonize)")

	// Stop command flags
	registryStopCmd.Flags().Bool("force", false, "Force kill (SIGKILL instead of graceful shutdown)")

	// Status command flags (no flags needed - args determine behavior)

	// Logs command flags
	registryLogsCmd.Flags().IntP("lines", "n", 50, "Number of lines to show")
	registryLogsCmd.Flags().String("since", "", "Show logs since duration (e.g., '2h', '30m')")

	// Prune command flags
	registryPruneCmd.Flags().Bool("all", false, "Remove all images (not just unused)")
	registryPruneCmd.Flags().String("older-than", "", "Remove images older than duration (e.g., '7d', '30d')")
	registryPruneCmd.Flags().Bool("dry-run", false, "Show what would be removed without removing")
	registryPruneCmd.Flags().Bool("force", false, "Skip confirmation prompt")
}

// =============================================================================
// Command Implementations
// =============================================================================

// runRegistryStart implements the registry start command
func runRegistryStart(cmd *cobra.Command, args []string) error {
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

	render.Success(fmt.Sprintf("Registry '%s' started", name))
	render.Info(fmt.Sprintf("Endpoint: %s", mgr.GetEndpoint()))

	// Handle foreground mode
	foreground, _ := cmd.Flags().GetBool("foreground")
	if foreground {
		render.Info("Press Ctrl+C to stop")
		<-ctx.Done()
		mgr.Stop(context.Background())
		render.Info("Registry stopped")
	}

	return nil
}

// runRegistryStop implements the registry stop command
func runRegistryStop(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	name := args[0] // Get name from positional arg

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

// runRegistryStatus implements the registry status command
func runRegistryStatus(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	format, _ := cmd.Flags().GetString("output")

	// Get DataStore from context
	dataStore, ok := ctx.Value("dataStore").(*db.DataStore)
	if !ok || dataStore == nil {
		return fmt.Errorf("database not initialized")
	}
	store := *dataStore

	// If no args, list all registries
	if len(args) == 0 {
		return listAllRegistriesStatus(ctx, store, format)
	}

	// Show specific registry
	name := args[0]
	return showRegistryStatus(ctx, store, name, format)
}

func listAllRegistriesStatus(ctx context.Context, store db.DataStore, format string) error {
	registries, err := store.ListRegistries()
	if err != nil {
		return fmt.Errorf("failed to list registries: %w", err)
	}

	if len(registries) == 0 {
		render.Info("No registries found")
		return nil
	}

	factory := registry.NewServiceFactory()

	// Build table data
	headers := []string{"NAME", "TYPE", "STATE", "ENDPOINT"}
	var rows [][]string

	for _, reg := range registries {
		mgr, err := factory.CreateManager(reg)
		state := "unknown"
		endpoint := "-"

		if err == nil {
			if mgr.IsRunning(ctx) {
				state = "running"
				endpoint = mgr.GetEndpoint()
			} else {
				state = "stopped"
			}
		} else {
			state = "error"
		}

		rows = append(rows, []string{reg.Name, reg.Type, state, endpoint})
	}

	tableData := render.TableData{Headers: headers, Rows: rows}
	return render.OutputWith(format, tableData, render.Options{Type: render.TypeTable})
}

func showRegistryStatus(ctx context.Context, store db.DataStore, name string, format string) error {
	reg, err := store.GetRegistryByName(name)
	if err != nil {
		return fmt.Errorf("registry '%s' not found: %w", name, err)
	}

	factory := registry.NewServiceFactory()
	mgr, err := factory.CreateManager(reg)
	if err != nil {
		return fmt.Errorf("failed to create manager for registry '%s': %w", name, err)
	}

	// Check if running
	isRunning := mgr.IsRunning(ctx)
	endpoint := mgr.GetEndpoint()

	// Handle YAML/JSON output
	if format == "yaml" || format == "json" {
		statusMap := map[string]interface{}{
			"name":     name,
			"type":     reg.Type,
			"state":    getStateString(isRunning),
			"endpoint": endpoint,
		}
		return render.OutputWith(format, statusMap, render.Options{})
	}

	// Build table data
	headers := []string{"NAME", "TYPE", "STATE", "ENDPOINT"}
	row := []string{
		name,
		reg.Type,
		getStateString(isRunning),
		endpoint,
	}

	tableData := render.TableData{
		Headers: headers,
		Rows:    [][]string{row},
	}

	return render.OutputWith(format, tableData, render.Options{
		Type: render.TypeTable,
	})
}

// getStateString returns a human-readable state string
func getStateString(isRunning bool) string {
	if isRunning {
		return "running"
	}
	return "stopped"
}

// runRegistryLogs implements the registry logs command
func runRegistryLogs(cmd *cobra.Command, args []string) error {
	// TODO(v0.21.0): Update to use ServiceFactory pattern with registry name argument
	return fmt.Errorf("registry logs command not yet updated for multi-registry support - use 'docker logs <container>' directly")
}

// runRegistryPrune implements the registry prune command
func runRegistryPrune(cmd *cobra.Command, args []string) error {
	// TODO(v0.21.0): Update to use ServiceFactory pattern with registry name argument
	return fmt.Errorf("registry prune command not yet updated for multi-registry support")
}

// =============================================================================
// Helper Functions
// =============================================================================

// formatPID formats a PID for display
func formatPID(pid int) string {
	if pid == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", pid)
}

// formatDuration formats a duration for display
func formatDuration(d time.Duration) string {
	if d == 0 {
		return "-"
	}

	// Round to seconds
	d = d.Round(time.Second)

	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh%dm", hours, minutes)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

// formatBytes formats bytes as human-readable string
func formatBytes(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}

	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// parseDuration parses a duration string that may include days (e.g., "7d", "30d", "2h")
func parseDuration(s string) (time.Duration, error) {
	// Handle day notation
	if strings.HasSuffix(s, "d") {
		days := strings.TrimSuffix(s, "d")
		var d int
		if _, err := fmt.Sscanf(days, "%d", &d); err != nil {
			return 0, fmt.Errorf("invalid duration: %s", s)
		}
		return time.Duration(d) * 24 * time.Hour, nil
	}

	// Standard Go duration parsing
	return time.ParseDuration(s)
}
