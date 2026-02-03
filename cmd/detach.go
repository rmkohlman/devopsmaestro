package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/operators"
	"devopsmaestro/render"
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var (
	detachAll bool
)

// detachCmd stops the active workspace container
var detachCmd = &cobra.Command{
	Use:   "detach",
	Short: "Stop and detach from a workspace container",
	Long: `Stop and detach from a workspace container.

By default, stops the currently active workspace. The container is stopped
but not removed, so you can quickly re-attach later with 'dvm attach'.

Use --all to stop all DVM workspace containers.

Examples:
  dvm detach              # Stop active workspace
  dvm detach --all        # Stop all DVM workspaces
  dvm detach -a           # Short form`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDetach(cmd)
	},
}

func init() {
	rootCmd.AddCommand(detachCmd)
	detachCmd.Flags().BoolVarP(&detachAll, "all", "a", false, "Stop all DVM workspace containers")
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
	// Get context manager
	contextMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to initialize context manager: %w", err)
	}

	// Get active project and workspace
	projectName, err := contextMgr.GetActiveProject()
	if err != nil {
		render.Warning("No active project set")
		render.Info("Set active project with: dvm use project <name>")
		return nil
	}

	workspaceName, err := contextMgr.GetActiveWorkspace()
	if err != nil {
		render.Warning("No active workspace set")
		render.Info("Set active workspace with: dvm use workspace <name>")
		return nil
	}

	slog.Debug("detach context", "project", projectName, "workspace", workspaceName)

	// Get datastore from context
	ctx := cmd.Context()
	dataStore := ctx.Value("dataStore").(*db.DataStore)
	if dataStore == nil {
		return fmt.Errorf("dataStore not initialized")
	}

	ds := *dataStore

	// Verify project and workspace exist
	project, err := ds.GetProjectByName(projectName)
	if err != nil {
		return fmt.Errorf("failed to get project '%s': %w", projectName, err)
	}

	_, err = ds.GetWorkspaceByName(project.ID, workspaceName)
	if err != nil {
		return fmt.Errorf("failed to get workspace '%s': %w", workspaceName, err)
	}

	// Stop the container
	containerName := fmt.Sprintf("dvm-%s-%s", projectName, workspaceName)
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
