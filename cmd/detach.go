package cmd

import (
	"context"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/resolver"
	"fmt"
	"github.com/rmkohlman/MaestroSDK/render"
	"log/slog"
	"time"

	"github.com/spf13/cobra"
)

var (
	detachAll     bool
	detachTimeout time.Duration
	detachFlags   HierarchyFlags
	detachDryRun  bool
)

// detachCmd stops the active workspace container
var detachCmd = &cobra.Command{
	Use:   "detach",
	Short: "Stop and detach from a workspace container",
	Long: `Stop and detach from a workspace container.

By default, stops the currently active workspace. The container is stopped
but not removed, so you can quickly re-attach later with 'dvm attach'.

Use -A (--all) to stop all DVM workspace containers.

Flags:
  -e, --ecosystem   Filter by ecosystem name
  -d, --domain      Filter by domain name  
  -a, --app         Filter by app name
  -w, --workspace   Filter by workspace name
  -A, --all         Stop all DVM workspace containers

Examples:
  dvm detach                         # Stop active workspace
  dvm detach -a portal               # Stop workspace in 'portal' app
  dvm detach -e healthcare -a portal # Specify ecosystem and app
  dvm detach -A                      # Stop all DVM workspaces`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDetach(cmd)
	},
}

func init() {
	rootCmd.AddCommand(detachCmd)
	detachCmd.Flags().BoolVarP(&detachAll, "all", "A", false, "Stop all DVM workspace containers")
	detachCmd.Flags().DurationVar(&detachTimeout, "timeout", 5*time.Minute, "Timeout for the detach operation (e.g., 5m, 30s)")
	AddHierarchyFlags(detachCmd, &detachFlags)
	AddDryRunFlag(detachCmd, &detachDryRun)
}

func runDetach(cmd *cobra.Command) error {
	// Dry-run: preview what would be stopped
	if detachDryRun {
		if detachAll {
			render.Plain("Would stop all DVM workspace containers")
			return nil
		}
		render.Plain("Would stop the active workspace container")
		return nil
	}

	// Create timeout context from flag
	ctx := context.Background()
	if detachTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, detachTimeout)
		defer cancel()
	}

	// Create container runtime using factory
	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		return fmt.Errorf("failed to create container runtime: %w", err)
	}

	slog.Debug("using runtime", "type", runtime.GetRuntimeType(), "platform", runtime.GetPlatformName())

	if detachAll {
		return detachAllWorkspaces(ctx, runtime)
	}

	return detachActiveWorkspace(cmd, ctx, runtime)
}

func detachActiveWorkspace(cmd *cobra.Command, ctx context.Context, runtime operators.ContainerRuntime) error {
	// Get datastore from context
	ds, err := getDataStore(cmd)
	if err != nil {
		return fmt.Errorf("dataStore not initialized: %w", err)
	}

	var app *models.App
	var workspace *models.Workspace
	var appName, workspaceName string
	var ecosystemName, domainName, systemName string // For hierarchical container naming

	// Check if hierarchy flags were provided
	if detachFlags.HasAnyFlag() {
		// Use resolver to find workspace
		slog.Debug("using hierarchy flags", "ecosystem", detachFlags.Ecosystem,
			"domain", detachFlags.Domain, "system", detachFlags.System, "app", detachFlags.App, "workspace", detachFlags.Workspace)

		wsResolver := resolver.NewWorkspaceResolver(ds)
		result, err := wsResolver.Resolve(detachFlags.ToFilter())
		if err != nil {
			// Check if ambiguous and provide helpful output
			if ambiguousErr, ok := resolver.IsAmbiguousError(err); ok {
				render.Warning("Multiple workspaces match your criteria")
				render.Plain(ambiguousErr.FormatDisambiguation())
				render.Plain(FormatSuggestions(SuggestAmbiguousWorkspace()...))
				return fmt.Errorf("ambiguous workspace selection")
			}
			if resolver.IsNoWorkspaceFoundError(err) {
				render.Warning("No workspace found matching your criteria")
				render.Plain(FormatSuggestions(SuggestWorkspaceNotFound("")...))
				return err
			}
			return fmt.Errorf("failed to resolve workspace: %w", err)
		}

		// Use resolved workspace and app
		workspace = result.Workspace
		app = result.App
		appName = app.Name
		workspaceName = workspace.Name
		ecosystemName = result.Ecosystem.Name
		domainName = result.Domain.Name
		if result.System != nil {
			systemName = result.System.Name
		}

		render.Info(fmt.Sprintf("Resolved: %s", result.FullPath()))
	} else {
		// Fall back to existing context-based behavior (DB-backed)
		var err error
		appName, err = getActiveAppFromContext(ds)
		if err != nil {
			render.Warning("No active app set")
			render.Plain(FormatSuggestions(SuggestNoActiveApp()...))
			return nil
		}

		workspaceName, err = getActiveWorkspaceFromContext(ds)
		if err != nil {
			render.Warning("No active workspace set")
			render.Plain(FormatSuggestions(SuggestNoActiveWorkspace()...))
			return nil
		}

		slog.Debug("detach context", "app", appName, "workspace", workspaceName)

		// Verify app and workspace exist (search globally across all domains)
		app, err = ds.GetAppByNameGlobal(appName)
		if err != nil {
			return fmt.Errorf("failed to get app '%s': %w", appName, err)
		}

		workspace, err = ds.GetWorkspaceByName(app.ID, workspaceName)
		if err != nil {
			return fmt.Errorf("failed to get workspace '%s': %w", workspaceName, err)
		}
	}

	// Use workspace and app to ensure they're referenced (avoid unused variable warnings)
	_ = workspace

	// Stop the container using hierarchical naming strategy
	namingStrategy := operators.NewHierarchicalNamingStrategy()
	containerName := namingStrategy.GenerateName(ecosystemName, domainName, systemName, appName, workspaceName)
	return stopWorkspace(ctx, runtime, containerName)
}

func detachAllWorkspaces(ctx context.Context, runtime operators.ContainerRuntime) error {
	render.Progress("Finding all DVM workspace containers...")

	stopped, err := runtime.StopAllWorkspaces(ctx)
	if err != nil {
		return fmt.Errorf("failed to stop workspaces: %w", err)
	}

	if stopped == 0 {
		render.Info("No running DVM workspace containers found")
		return nil
	}

	render.Blank()
	render.Success(fmt.Sprintf("Stopped %d workspace container(s)", stopped))
	return nil
}

func stopWorkspace(ctx context.Context, runtime operators.ContainerRuntime, containerName string) error {
	render.Progress(fmt.Sprintf("Stopping workspace '%s'...", containerName))

	// Check if workspace exists and is running
	workspace, err := runtime.FindWorkspace(ctx, containerName)
	if err != nil {
		return fmt.Errorf("failed to find workspace: %w", err)
	}

	if workspace == nil {
		render.Info(fmt.Sprintf("Workspace '%s' not found", containerName))
		return nil
	}

	// Check if already stopped
	if workspace.Status != "running" && workspace.Status != "Up" && !containsRunning(workspace.Status) {
		render.Info(fmt.Sprintf("Workspace '%s' is not running (status: %s)", containerName, workspace.Status))
		return nil
	}

	// Stop the workspace
	if err := runtime.StopWorkspace(ctx, containerName); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	slog.Info("workspace stopped", "name", containerName)
	render.Success(fmt.Sprintf("Workspace '%s' stopped", containerName))
	render.Blank()
	render.Info("Re-attach with: dvm attach")

	return nil
}

// containsRunning checks if the status string indicates a running container
// Docker status can be "Up 5 minutes" or similar
func containsRunning(status string) bool {
	return len(status) >= 2 && status[:2] == "Up"
}
