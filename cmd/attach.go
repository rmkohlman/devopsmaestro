package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/operators"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
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
		slog.Info("starting attach")

		// Get context manager
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			slog.Error("failed to initialize context manager", "error", err)
			fmt.Printf("Error: Failed to initialize context manager: %v\n", err)
			return
		}

		// Get active project and workspace
		projectName, err := contextMgr.GetActiveProject()
		if err != nil {
			slog.Debug("no active project", "error", err)
			fmt.Printf("Error: %v\n", err)
			fmt.Println("\nHint: Set active project with: dvm use project <name>")
			return
		}

		workspaceName, err := contextMgr.GetActiveWorkspace()
		if err != nil {
			slog.Debug("no active workspace", "error", err)
			fmt.Printf("Error: %v\n", err)
			fmt.Println("\nHint: Set active workspace with: dvm use workspace <name>")
			return
		}

		slog.Debug("attach context", "project", projectName, "workspace", workspaceName)
		fmt.Printf("Project: %s | Workspace: %s\n", projectName, workspaceName)

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			slog.Error("datastore not initialized in context")
			fmt.Println("Error: DataStore not initialized")
			return
		}

		ds := *dataStore

		// Get project
		project, err := ds.GetProjectByName(projectName)
		if err != nil {
			slog.Error("failed to get project", "name", projectName, "error", err)
			fmt.Printf("Error: Failed to get project: %v\n", err)
			return
		}

		// Get workspace
		workspace, err := ds.GetWorkspaceByName(project.ID, workspaceName)
		if err != nil {
			slog.Error("failed to get workspace", "name", workspaceName, "project_id", project.ID, "error", err)
			fmt.Printf("Error: Failed to get workspace: %v\n", err)
			return
		}

		slog.Debug("resolved workspace", "image", workspace.ImageName, "project_path", project.Path)

		// Initialize container runtime
		runtime, err := operators.NewContainerRuntime()
		if err != nil {
			slog.Error("failed to initialize container runtime", "error", err)
			fmt.Printf("Error: Failed to initialize container runtime: %v\n", err)
			return
		}

		slog.Info("using container runtime", "type", runtime.GetRuntimeType())
		fmt.Printf("Using %s runtime\n", runtime.GetRuntimeType())

		// Use image name from workspace
		// The build command stores the built image name (e.g., dvm-main-fastapi-test:latest)
		imageName := workspace.ImageName

		// If the image name doesn't have the dvm- prefix, it might be the original default
		// and the workspace hasn't been built yet
		if !strings.HasPrefix(imageName, "dvm-") {
			slog.Warn("workspace image may not be built", "image", imageName)
			fmt.Printf("Warning: Workspace image '%s' may not be built.\n", imageName)
			fmt.Println("Run 'dvm build' first to build the development container.")
			fmt.Println()
		}

		fmt.Printf("Starting workspace container...\n")

		// For Colima, we need to use nerdctl via SSH instead of direct containerd API
		// because mounts and other operations don't work across host/VM boundary
		profile := os.Getenv("COLIMA_ACTIVE_PROFILE")
		if profile == "" {
			profile = "default"
		}

		containerName := fmt.Sprintf("dvm-%s-%s", projectName, workspaceName)
		slog.Debug("container details", "name", containerName, "image", imageName, "profile", profile)

		// Check if container already exists and is running
		checkCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
			"sudo", "nerdctl", "--namespace", "devopsmaestro", "ps", "-q", "-f", fmt.Sprintf("name=%s", containerName))
		output, _ := checkCmd.Output()

		if len(output) == 0 {
			// Container doesn't exist or isn't running - create it
			fmt.Println("Creating container...")
			slog.Debug("creating new container", "name", containerName, "project_path", project.Path)

			// Convert project path to VM path (assumes home directory is mounted)
			vmProjectPath := project.Path

			createCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
				"sudo", "nerdctl", "--namespace", "devopsmaestro", "run",
				"-d", // Detached
				"--name", containerName,
				"-v", fmt.Sprintf("%s:/workspace", vmProjectPath),
				"-v", fmt.Sprintf("%s/.ssh:/root/.ssh:ro", os.Getenv("HOME")),
				"-w", "/workspace",
				"-e", fmt.Sprintf("DVM_PROJECT=%s", projectName),
				"-e", fmt.Sprintf("DVM_WORKSPACE=%s", workspaceName),
				imageName,
				"/bin/sleep", "infinity", // Keep container running
			)
			createCmd.Stdout = os.Stdout
			createCmd.Stderr = os.Stderr

			if err := createCmd.Run(); err != nil {
				slog.Error("failed to create container", "name", containerName, "error", err)
				fmt.Printf("Error: Failed to create container: %v\n", err)
				return
			}
			slog.Info("container created", "name", containerName)
		} else {
			slog.Debug("container already running", "name", containerName)
			fmt.Println("Container already running")
		}

		// Attach to container using nerdctl exec
		fmt.Println("Attaching to workspace...")
		slog.Info("attaching to container", "name", containerName)
		attachCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
			"sudo", "nerdctl", "--namespace", "devopsmaestro", "exec", "-it", containerName, "/bin/zsh", "-l")
		attachCmd.Stdin = os.Stdin
		attachCmd.Stdout = os.Stdout
		attachCmd.Stderr = os.Stderr

		if err := attachCmd.Run(); err != nil {
			slog.Error("failed to attach to container", "name", containerName, "error", err)
			fmt.Printf("Error: Failed to attach: %v\n", err)
			return
		}

		slog.Info("session ended", "container", containerName)
		fmt.Println("Session ended.")
	},
}

// Initializes the attach command
func init() {
	rootCmd.AddCommand(attachCmd)
}
