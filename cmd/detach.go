package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/operators"
	"devopsmaestro/pkg/resolver"
	"devopsmaestro/render"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var (
	detachAll   bool
	detachFlags HierarchyFlags
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
	AddHierarchyFlags(detachCmd, &detachFlags)
}

func runDetach(cmd *cobra.Command) error {
	// Create container runtime using factory
	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		return fmt.Errorf("failed to create container runtime: %w", err)
	}

	slog.Debug("using runtime", "type", runtime.GetRuntimeType(), "platform", runtime.GetPlatformName())

	if detachAll {
		return detachAllWorkspaces(runtime)
	}

	return detachActiveWorkspace(cmd, runtime)
}

func detachActiveWorkspace(cmd *cobra.Command, runtime operators.ContainerRuntime) error {
	// Get datastore from context
	ctx := cmd.Context()
	dataStore := ctx.Value("dataStore").(*db.DataStore)
	if dataStore == nil {
		return fmt.Errorf("dataStore not initialized")
	}
	ds := *dataStore

	var app *models.App
	var workspace *models.Workspace
	var appName, workspaceName string
	var ecosystemName, domainName string // For hierarchical container naming

	// Check if hierarchy flags were provided
	if detachFlags.HasAnyFlag() {
		// Use resolver to find workspace
		slog.Debug("using hierarchy flags", "ecosystem", detachFlags.Ecosystem,
			"domain", detachFlags.Domain, "app", detachFlags.App, "workspace", detachFlags.Workspace)

		wsResolver := resolver.NewWorkspaceResolver(ds)
		result, err := wsResolver.Resolve(detachFlags.ToFilter())
		if err != nil {
			// Check if ambiguous and provide helpful output
			if ambiguousErr, ok := resolver.IsAmbiguousError(err); ok {
				render.Warning("Multiple workspaces match your criteria")
				fmt.Println(ambiguousErr.FormatDisambiguation())
				return fmt.Errorf("ambiguous workspace selection")
			}
			if resolver.IsNoWorkspaceFoundError(err) {
				render.Warning("No workspace found matching your criteria")
				render.Info("Hint: Use 'dvm get workspaces' to see available workspaces")
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

		// Update context to the resolved workspace
		if err := updateContextFromHierarchy(ds, result); err != nil {
			slog.Warn("failed to update context", "error", err)
			// Continue anyway - this is not fatal
		}

		render.Info(fmt.Sprintf("Resolved: %s", result.FullPath()))
	} else {
		// Fall back to existing context-based behavior
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			return fmt.Errorf("failed to initialize context manager: %w", err)
		}

		// Get active app and workspace
		appName, err = contextMgr.GetActiveApp()
		if err != nil {
			render.Warning("No active app set")
			render.Info("Set active app with: dvm use app <name>")
			render.Info("      Or use flags: dvm detach -a <app>")
			return nil
		}

		workspaceName, err = contextMgr.GetActiveWorkspace()
		if err != nil {
			render.Warning("No active workspace set")
			render.Info("Set active workspace with: dvm use workspace <name>")
			render.Info("      Or use flags: dvm detach -w <workspace>")
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
	containerName := namingStrategy.GenerateName(ecosystemName, domainName, appName, workspaceName)
	return stopWorkspace(runtime, containerName)
}

func detachAllWorkspaces(runtime operators.ContainerRuntime) error {
	render.Progress("Finding all DVM workspace containers...")

	stopped, err := runtime.StopAllWorkspaces(context.Background())
	if err != nil {
		return fmt.Errorf("failed to stop workspaces: %w", err)
	}

	if stopped == 0 {
		render.Info("No running DVM workspace containers found")
		return nil
	}

	fmt.Println()
	render.Success(fmt.Sprintf("Stopped %d workspace container(s)", stopped))
	return nil
}

func stopWorkspace(runtime operators.ContainerRuntime, containerName string) error {
	render.Progress(fmt.Sprintf("Stopping workspace '%s'...", containerName))

	// Check if workspace exists and is running
	workspace, err := runtime.FindWorkspace(context.Background(), containerName)
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
	if err := runtime.StopWorkspace(context.Background(), containerName); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	slog.Info("workspace stopped", "name", containerName)
	render.Success(fmt.Sprintf("Workspace '%s' stopped", containerName))
	fmt.Println()
	render.Info("Re-attach with: dvm attach")

	return nil
}

// containsRunning checks if the status string indicates a running container
// Docker status can be "Up 5 minutes" or similar
func containsRunning(status string) bool {
	return len(status) >= 2 && status[:2] == "Up"
}
