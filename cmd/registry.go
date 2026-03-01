package cmd

import (
	"bufio"
	"context"
	"devopsmaestro/config"
	"devopsmaestro/pkg/registry"
	"devopsmaestro/render"
	"fmt"
	"os"
	"path/filepath"
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
	Use:   "start",
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
	Use:   "stop",
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
	Use:   "status",
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

	// Get registry config
	cfg := config.GetRegistryConfig()

	// Override port if specified
	if portFlag, _ := cmd.Flags().GetInt("port"); portFlag > 0 {
		cfg.Port = portFlag
	}

	// Check foreground flag
	foreground, _ := cmd.Flags().GetBool("foreground")
	if foreground {
		render.Info("Running in foreground mode (Ctrl+C to stop)")
	}

	// Create registry manager
	regCfg := convertToRegistryConfig(cfg)
	if err := regCfg.Validate(); err != nil {
		return fmt.Errorf("invalid registry config: %w", err)
	}
	mgr := registry.NewRegistryManager(regCfg)

	// Check if already running
	if mgr.IsRunning(ctx) {
		status, _ := mgr.Status(ctx)
		if status != nil {
			render.Info(fmt.Sprintf("Registry already running (PID: %d, Port: %d)", status.PID, status.Port))
		} else {
			render.Info("Registry already running")
		}
		return nil
	}

	// Start the registry
	render.Progress("Starting registry...")
	if err := mgr.Start(ctx); err != nil {
		return fmt.Errorf("failed to start registry: %w", err)
	}

	// Get status to show details
	status, err := mgr.Status(ctx)
	if err != nil {
		render.Success("Registry started")
		return nil
	}

	render.Success(fmt.Sprintf("Registry started (PID: %d, Port: %d)", status.PID, status.Port))
	render.Info(fmt.Sprintf("Endpoint: %s", mgr.GetEndpoint()))

	// If foreground, wait for interrupt
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

	// Get registry config
	cfg := config.GetRegistryConfig()

	// Create registry manager
	regCfg := convertToRegistryConfig(cfg)
	if err := regCfg.Validate(); err != nil {
		return fmt.Errorf("invalid registry config: %w", err)
	}
	mgr := registry.NewRegistryManager(regCfg)

	// Check if running
	if !mgr.IsRunning(ctx) {
		render.Info("Registry is not running")
		return nil
	}

	// Check force flag
	force, _ := cmd.Flags().GetBool("force")
	if force {
		render.Progress("Force stopping registry...")
	} else {
		render.Progress("Stopping registry...")
	}

	// Stop the registry
	if err := mgr.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop registry: %w", err)
	}

	render.Success("Registry stopped")
	return nil
}

// runRegistryStatus implements the registry status command
func runRegistryStatus(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get output format
	format, _ := cmd.Flags().GetString("output")

	// Get registry config
	cfg := config.GetRegistryConfig()

	// Create registry manager
	regCfg := convertToRegistryConfig(cfg)
	if err := regCfg.Validate(); err != nil {
		return fmt.Errorf("invalid registry config: %w", err)
	}
	mgr := registry.NewRegistryManager(regCfg)

	// Get status
	status, err := mgr.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get registry status: %w", err)
	}

	// Handle YAML/JSON output
	if format == "yaml" || format == "json" {
		return render.OutputWith(format, registryStatusToMap(status, cfg), render.Options{})
	}

	// Determine if wide format
	isWide := format == "wide"

	// Build table data
	var headers []string
	var row []string

	if isWide {
		headers = []string{"STATE", "PID", "PORT", "UPTIME", "IMAGES", "DISK_USAGE", "VERSION", "STORAGE"}
		row = []string{
			status.State,
			formatPID(status.PID),
			fmt.Sprintf("%d", status.Port),
			formatDuration(status.Uptime),
			fmt.Sprintf("%d", status.ImageCount),
			formatBytes(status.DiskUsage),
			status.Version,
			status.Storage,
		}
	} else {
		headers = []string{"STATE", "PID", "PORT", "UPTIME", "IMAGES", "DISK_USAGE"}
		row = []string{
			status.State,
			formatPID(status.PID),
			fmt.Sprintf("%d", status.Port),
			formatDuration(status.Uptime),
			fmt.Sprintf("%d", status.ImageCount),
			formatBytes(status.DiskUsage),
		}
	}

	tableData := render.TableData{
		Headers: headers,
		Rows:    [][]string{row},
	}

	return render.OutputWith(format, tableData, render.Options{
		Type: render.TypeTable,
	})
}

// runRegistryLogs implements the registry logs command
func runRegistryLogs(cmd *cobra.Command, args []string) error {
	// Get flags
	lines, _ := cmd.Flags().GetInt("lines")
	since, _ := cmd.Flags().GetString("since")

	// Get registry config
	cfg := config.GetRegistryConfig()

	// Determine log file path
	homeDir, _ := os.UserHomeDir()
	logFile := filepath.Join(homeDir, ".devopsmaestro", "registry", "zot.log")

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		render.Info("No registry logs found")
		render.Info(fmt.Sprintf("Log file: %s", logFile))
		return nil
	}

	// Parse since duration if provided
	var sinceTime time.Time
	if since != "" {
		duration, err := parseDuration(since)
		if err != nil {
			return fmt.Errorf("invalid --since value: %w", err)
		}
		sinceTime = time.Now().Add(-duration)
	}

	// Read log file
	file, err := os.Open(logFile)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	// Read all lines (simple approach for now)
	var logLines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		// Filter by time if --since specified
		if !sinceTime.IsZero() {
			// Try to parse timestamp from log line (ISO8601 format)
			// Zot logs typically start with timestamp
			if len(line) > 20 {
				if ts, err := time.Parse(time.RFC3339, line[:20]); err == nil {
					if ts.Before(sinceTime) {
						continue
					}
				}
			}
		}

		logLines = append(logLines, line)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading log file: %w", err)
	}

	// Take last N lines
	if len(logLines) > lines {
		logLines = logLines[len(logLines)-lines:]
	}

	// Output lines
	if len(logLines) == 0 {
		render.Info("No log entries found")
		return nil
	}

	for _, line := range logLines {
		fmt.Println(line)
	}

	// Show log file location
	render.Info(fmt.Sprintf("\nLog file: %s", logFile))
	_ = cfg // Silence unused variable warning

	return nil
}

// runRegistryPrune implements the registry prune command
func runRegistryPrune(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get flags
	all, _ := cmd.Flags().GetBool("all")
	olderThan, _ := cmd.Flags().GetString("older-than")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")

	// Validate flags - must specify either --all or --older-than
	if !all && olderThan == "" {
		return fmt.Errorf("must specify either --all or --older-than")
	}

	// Parse older-than duration if provided
	var olderThanDuration time.Duration
	if olderThan != "" {
		var err error
		olderThanDuration, err = parseDuration(olderThan)
		if err != nil {
			return fmt.Errorf("invalid --older-than value: %w", err)
		}
	}

	// Get registry config
	cfg := config.GetRegistryConfig()

	// Create registry manager
	regCfg := convertToRegistryConfig(cfg)
	if err := regCfg.Validate(); err != nil {
		return fmt.Errorf("invalid registry config: %w", err)
	}
	mgr := registry.NewRegistryManager(regCfg)

	// Check if running
	if !mgr.IsRunning(ctx) {
		return fmt.Errorf("registry is not running - start it first with 'dvm registry start'")
	}

	// Build prune options
	pruneOpts := registry.PruneOptions{
		All:       all,
		OlderThan: olderThanDuration,
		DryRun:    dryRun,
	}

	// Confirm if not --force and not --dry-run
	if !force && !dryRun {
		var description string
		if all {
			description = "all cached images"
		} else {
			description = fmt.Sprintf("images older than %s", olderThan)
		}

		render.Warning(fmt.Sprintf("This will remove %s from the registry.", description))
		fmt.Print("Continue? [y/N]: ")

		var response string
		fmt.Scanln(&response)
		if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
			render.Info("Cancelled")
			return nil
		}
	}

	// Execute prune
	if dryRun {
		render.Progress("Calculating what would be removed...")
	} else {
		render.Progress("Pruning registry...")
	}

	result, err := mgr.Prune(ctx, pruneOpts)
	if err != nil {
		return fmt.Errorf("failed to prune registry: %w", err)
	}

	// Show results
	if dryRun {
		if result.ImagesRemoved == 0 {
			render.Info("No images would be removed")
		} else {
			render.Info(fmt.Sprintf("Would remove %d images, freeing %s",
				result.ImagesRemoved, formatBytes(result.SpaceReclaimed)))
			if len(result.Images) > 0 {
				render.Info("\nImages that would be removed:")
				for _, img := range result.Images {
					fmt.Printf("  - %s\n", img)
				}
			}
		}
	} else {
		if result.ImagesRemoved == 0 {
			render.Info("No images removed")
		} else {
			render.Success(fmt.Sprintf("Removed %d images, freed %s",
				result.ImagesRemoved, formatBytes(result.SpaceReclaimed)))
		}
	}

	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// convertToRegistryConfig converts config.RegistryConfig to registry.RegistryConfig
func convertToRegistryConfig(cfg *config.RegistryConfig) registry.RegistryConfig {
	mirrors := make([]registry.MirrorConfig, len(cfg.Mirrors))
	for i, m := range cfg.Mirrors {
		mirrors[i] = registry.MirrorConfig{
			Name:     m.Name,
			URL:      m.URL,
			OnDemand: m.OnDemand,
			Prefix:   m.Prefix,
		}
	}

	return registry.RegistryConfig{
		Enabled:     cfg.Enabled,
		Lifecycle:   cfg.Lifecycle,
		Port:        cfg.Port,
		Storage:     cfg.Storage,
		IdleTimeout: cfg.IdleTimeout,
		Mirrors:     mirrors,
	}
}

// registryStatusToMap converts RegistryStatus to a map for YAML/JSON output
func registryStatusToMap(status *registry.RegistryStatus, cfg *config.RegistryConfig) map[string]interface{} {
	return map[string]interface{}{
		"state":      status.State,
		"pid":        status.PID,
		"port":       status.Port,
		"storage":    status.Storage,
		"version":    status.Version,
		"uptime":     status.Uptime.String(),
		"imageCount": status.ImageCount,
		"diskUsage":  status.DiskUsage,
		"lifecycle":  cfg.Lifecycle,
		"endpoint":   fmt.Sprintf("localhost:%d", cfg.Port),
	}
}

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
