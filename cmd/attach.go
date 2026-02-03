package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/operators"
	"devopsmaestro/render"
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
			render.Error(fmt.Sprintf("Failed to initialize context manager: %v", err))
			return
		}

		// Get active project and workspace
		projectName, err := contextMgr.GetActiveProject()
		if err != nil {
			slog.Debug("no active project", "error", err)
			render.Error(fmt.Sprintf("%v", err))
			render.Info("Hint: Set active project with: dvm use project <name>")
			return
		}

		workspaceName, err := contextMgr.GetActiveWorkspace()
		if err != nil {
			slog.Debug("no active workspace", "error", err)
			render.Error(fmt.Sprintf("%v", err))
			render.Info("Hint: Set active workspace with: dvm use workspace <name>")
			return
		}

		slog.Debug("attach context", "project", projectName, "workspace", workspaceName)
		render.Info(fmt.Sprintf("Project: %s | Workspace: %s", projectName, workspaceName))

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			slog.Error("datastore not initialized in context")
			render.Error("DataStore not initialized")
			return
		}

		ds := *dataStore

		// Get project
		project, err := ds.GetProjectByName(projectName)
		if err != nil {
			slog.Error("failed to get project", "name", projectName, "error", err)
			render.Error(fmt.Sprintf("Failed to get project: %v", err))
			return
		}

		// Get workspace
		workspace, err := ds.GetWorkspaceByName(project.ID, workspaceName)
		if err != nil {
			slog.Error("failed to get workspace", "name", workspaceName, "project_id", project.ID, "error", err)
			render.Error(fmt.Sprintf("Failed to get workspace: %v", err))
			return
		}

		slog.Debug("resolved workspace", "image", workspace.ImageName, "project_path", project.Path)

		// Detect container platform
		detector, err := operators.NewPlatformDetector()
		if err != nil {
			slog.Error("failed to create platform detector", "error", err)
			render.Error(fmt.Sprintf("Failed to detect container platform: %v", err))
			return
		}

		platform, err := detector.Detect()
		if err != nil {
			slog.Error("failed to detect platform", "error", err)
			render.Error(fmt.Sprintf("No container runtime found: %v", err))
			render.Info("Hint: Install OrbStack, Docker Desktop, or Colima")
			return
		}

		slog.Info("detected platform", "name", platform.Name, "type", platform.Type, "socket", platform.SocketPath)
		render.Info(fmt.Sprintf("Platform: %s", platform.Name))

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

		// Route to the appropriate attach implementation based on platform
		switch platform.Type {
		case operators.PlatformColima:
			attachColima(platform, containerName, imageName, project.Path, projectName, workspaceName)
		default:
			// OrbStack, Docker Desktop, Podman - all use Docker CLI
			attachDocker(platform, containerName, imageName, project.Path, projectName, workspaceName)
		}
	},
}

// attachDocker handles attach for Docker-compatible runtimes (OrbStack, Docker Desktop, Podman)
func attachDocker(platform *operators.Platform, containerName, imageName, projectPath, projectName, workspaceName string) {
	render.Progress("Starting workspace container...")

	// Check if container already exists and is running
	checkCmd := exec.Command("docker", "-H", "unix://"+platform.SocketPath,
		"ps", "-q", "-f", fmt.Sprintf("name=^%s$", containerName))
	output, _ := checkCmd.Output()

	if len(strings.TrimSpace(string(output))) == 0 {
		// Check if container exists but is stopped
		checkStoppedCmd := exec.Command("docker", "-H", "unix://"+platform.SocketPath,
			"ps", "-aq", "-f", fmt.Sprintf("name=^%s$", containerName))
		stoppedOutput, _ := checkStoppedCmd.Output()

		if len(strings.TrimSpace(string(stoppedOutput))) > 0 {
			// Container exists but is stopped - start it
			render.Progress("Starting existing container...")
			slog.Debug("starting stopped container", "name", containerName)

			startCmd := exec.Command("docker", "-H", "unix://"+platform.SocketPath,
				"start", containerName)
			startCmd.Stdout = os.Stdout
			startCmd.Stderr = os.Stderr

			if err := startCmd.Run(); err != nil {
				slog.Error("failed to start container", "name", containerName, "error", err)
				render.Error(fmt.Sprintf("Failed to start container: %v", err))
				return
			}
			slog.Info("container started", "name", containerName)
		} else {
			// Container doesn't exist - create it
			render.Progress("Creating container...")
			slog.Debug("creating new container", "name", containerName, "project_path", projectPath)

			createArgs := []string{
				"-H", "unix://" + platform.SocketPath,
				"run",
				"-d", // Detached
				"--name", containerName,
				"--label", "io.devopsmaestro.managed=true",
				"--label", "io.devopsmaestro.project=" + projectName,
				"--label", "io.devopsmaestro.workspace=" + workspaceName,
				"-v", fmt.Sprintf("%s:/workspace", projectPath),
				"-w", "/workspace",
				"-e", fmt.Sprintf("DVM_PROJECT=%s", projectName),
				"-e", fmt.Sprintf("DVM_WORKSPACE=%s", workspaceName),
			}

			// Mount SSH keys if they exist
			sshDir := os.Getenv("HOME") + "/.ssh"
			if _, err := os.Stat(sshDir); err == nil {
				createArgs = append(createArgs, "-v", fmt.Sprintf("%s:/home/dev/.ssh:ro", sshDir))
			}

			createArgs = append(createArgs, imageName, "/bin/sleep", "infinity")

			createCmd := exec.Command("docker", createArgs...)
			createCmd.Stdout = os.Stdout
			createCmd.Stderr = os.Stderr

			if err := createCmd.Run(); err != nil {
				slog.Error("failed to create container", "name", containerName, "error", err)
				render.Error(fmt.Sprintf("Failed to create container: %v", err))
				return
			}
			slog.Info("container created", "name", containerName)
		}
	} else {
		slog.Debug("container already running", "name", containerName)
		render.Info("Container already running")
	}

	// Attach to container
	render.Progress("Attaching to workspace...")
	slog.Info("attaching to container", "name", containerName)

	attachCmd := exec.Command("docker", "-H", "unix://"+platform.SocketPath,
		"exec", "-it", containerName, "/bin/zsh", "-l")
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr

	if err := attachCmd.Run(); err != nil {
		slog.Error("failed to attach to container", "name", containerName, "error", err)
		render.Error(fmt.Sprintf("Failed to attach: %v", err))
		return
	}

	slog.Info("session ended", "container", containerName)
	render.Info("Session ended.")
}

// attachColima handles attach for Colima with nerdctl
func attachColima(platform *operators.Platform, containerName, imageName, projectPath, projectName, workspaceName string) {
	render.Progress("Starting workspace container...")

	profile := os.Getenv("COLIMA_ACTIVE_PROFILE")
	if profile == "" {
		profile = "default"
	}

	slog.Debug("using colima profile", "profile", profile)

	// Check if container already exists and is running
	checkCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "devopsmaestro", "ps", "-q", "-f", fmt.Sprintf("name=%s", containerName))
	output, _ := checkCmd.Output()

	if len(output) == 0 {
		// Container doesn't exist or isn't running - create it
		render.Progress("Creating container...")
		slog.Debug("creating new container", "name", containerName, "project_path", projectPath)

		// Convert project path to VM path (assumes home directory is mounted)
		vmProjectPath := projectPath

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
			render.Error(fmt.Sprintf("Failed to create container: %v", err))
			return
		}
		slog.Info("container created", "name", containerName)
	} else {
		slog.Debug("container already running", "name", containerName)
		render.Info("Container already running")
	}

	// Attach to container using nerdctl exec
	render.Progress("Attaching to workspace...")
	slog.Info("attaching to container", "name", containerName)
	attachCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
		"sudo", "nerdctl", "--namespace", "devopsmaestro", "exec", "-it", containerName, "/bin/zsh", "-l")
	attachCmd.Stdin = os.Stdin
	attachCmd.Stdout = os.Stdout
	attachCmd.Stderr = os.Stderr

	if err := attachCmd.Run(); err != nil {
		slog.Error("failed to attach to container", "name", containerName, "error", err)
		render.Error(fmt.Sprintf("Failed to attach: %v", err))
		return
	}

	slog.Info("session ended", "container", containerName)
	render.Info("Session ended.")
}

// Initializes the attach command
func init() {
	rootCmd.AddCommand(attachCmd)
}
