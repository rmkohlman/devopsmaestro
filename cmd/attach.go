package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/operators"
	"devopsmaestro/render"
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/cobra"
)

// attachCmd attaches to the active workspace
var attachCmd = &cobra.Command{
	Use:   "attach",
	Short: "Attach to your workspace container",
	Long: `Attach an interactive terminal to your active workspace container.
If the workspace is not running, it will be started automatically.
If the container image doesn't exist, it will be built first.

The workspace provides your complete dev environment with:
- Neovim configuration
- oh-my-zsh + Powerlevel10k theme
- Your project files mounted at /workspace

Press Ctrl+D to detach from the workspace.

Examples:
  dvm attach
  DVM_WORKSPACE=dev dvm attach  # Override with env var`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runAttach(cmd); err != nil {
			render.Error(err.Error())
		}
	},
}

func runAttach(cmd *cobra.Command) error {
	slog.Info("starting attach")

	// Get context manager
	contextMgr, err := operators.NewContextManager()
	if err != nil {
		return fmt.Errorf("failed to initialize context manager: %w", err)
	}

	// Get active project and workspace
	projectName, err := contextMgr.GetActiveProject()
	if err != nil {
		render.Info("Hint: Set active project with: dvm use project <name>")
		return err
	}

	workspaceName, err := contextMgr.GetActiveWorkspace()
	if err != nil {
		render.Info("Hint: Set active workspace with: dvm use workspace <name>")
		return err
	}

	slog.Debug("attach context", "project", projectName, "workspace", workspaceName)
	render.Info(fmt.Sprintf("Project: %s | Workspace: %s", projectName, workspaceName))

	// Get datastore from context
	ctx := cmd.Context()
	dataStore := ctx.Value("dataStore").(*db.DataStore)
	if dataStore == nil {
		return fmt.Errorf("dataStore not initialized")
	}

	ds := *dataStore

	// Get project
	project, err := ds.GetProjectByName(projectName)
	if err != nil {
		return fmt.Errorf("failed to get project '%s': %w", projectName, err)
	}

	// Get workspace
	workspace, err := ds.GetWorkspaceByName(project.ID, workspaceName)
	if err != nil {
		return fmt.Errorf("failed to get workspace '%s': %w", workspaceName, err)
	}

	slog.Debug("resolved workspace", "image", workspace.ImageName, "project_path", project.Path)

	// Create container runtime using factory
	runtime, err := operators.NewContainerRuntime()
	if err != nil {
		render.Info("Hint: Install OrbStack, Docker Desktop, or Colima")
		return fmt.Errorf("failed to create container runtime: %w", err)
	}

	slog.Info("using runtime", "type", runtime.GetRuntimeType(), "platform", runtime.GetPlatformName())
	render.Info(fmt.Sprintf("Platform: %s", runtime.GetPlatformName()))

	// Use image name from workspace
	imageName := workspace.ImageName

	// If the image name doesn't have the dvm- prefix, it might be the original default
	// and the workspace hasn't been built yet
	if !strings.HasPrefix(imageName, "dvm-") {
		slog.Warn("workspace image may not be built", "image", imageName)
		render.Warning(fmt.Sprintf("Workspace image '%s' may not be built.", imageName))
		render.Info("Run 'dvm build' first to build the development container.")
		fmt.Println()
	}

	containerName := fmt.Sprintf("dvm-%s-%s", projectName, workspaceName)
	slog.Debug("container details", "name", containerName, "image", imageName)

	// Start workspace (handles existing containers automatically)
	render.Progress("Starting workspace container...")

	containerID, err := runtime.StartWorkspace(context.Background(), operators.StartOptions{
		ImageName:     imageName,
		WorkspaceName: workspaceName,
		ContainerName: containerName,
		ProjectName:   projectName,
		ProjectPath:   project.Path,
	})
	if err != nil {
		return fmt.Errorf("failed to start workspace: %w", err)
	}

	slog.Info("workspace started", "container_id", containerID)

	// Attach to workspace
	render.Progress("Attaching to workspace...")
	slog.Info("attaching to container", "name", containerName)

	if err := runtime.AttachToWorkspace(context.Background(), containerName); err != nil {
		return fmt.Errorf("failed to attach: %w", err)
	}

	slog.Info("session ended", "container", containerName)
	render.Info("Session ended.")
	return nil
}

// Initializes the attach command
func init() {
	rootCmd.AddCommand(attachCmd)
}
