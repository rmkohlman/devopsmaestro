package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/operators"
	"fmt"
	"os"
	"os/exec"

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
		// Get context manager
		contextMgr, err := operators.NewContextManager()
		if err != nil {
			fmt.Printf("Error: Failed to initialize context manager: %v\n", err)
			return
		}

		// Get active project and workspace
		projectName, err := contextMgr.GetActiveProject()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("\nHint: Set active project with: dvm use project <name>")
			return
		}

		workspaceName, err := contextMgr.GetActiveWorkspace()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("\nHint: Set active workspace with: dvm use workspace <name>")
			return
		}

		fmt.Printf("Project: %s | Workspace: %s\n", projectName, workspaceName)

		// Get datastore from context
		ctx := cmd.Context()
		dataStore := ctx.Value("dataStore").(*db.DataStore)
		if dataStore == nil {
			fmt.Println("Error: DataStore not initialized")
			return
		}

		sqlDS, ok := (*dataStore).(*db.SQLDataStore)
		if !ok {
			fmt.Println("Error: Expected SQLDataStore")
			return
		}

		// Get project
		project, err := sqlDS.GetProjectByName(projectName)
		if err != nil {
			fmt.Printf("Error: Failed to get project: %v\n", err)
			return
		}

		// Get workspace
		workspace, err := sqlDS.GetWorkspaceByName(project.ID, workspaceName)
		if err != nil {
			fmt.Printf("Error: Failed to get workspace: %v\n", err)
			return
		}

		// Initialize container runtime
		runtime, err := operators.NewContainerRuntime()
		if err != nil {
			fmt.Printf("Error: Failed to initialize container runtime: %v\n", err)
			return
		}

		fmt.Printf("Using %s runtime\n", runtime.GetRuntimeType())

		// Use image name from workspace
		imageName := workspace.ImageName

		// TODO: Check if image exists, build if needed
		// For MVP, we'll assume the image exists or will implement later
		// This would involve checking for a Dockerfile in the project directory
		// and building a multi-stage image with dev tools

		fmt.Printf("Starting workspace container...\n")

		// For Colima, we need to use nerdctl via SSH instead of direct containerd API
		// because mounts and other operations don't work across host/VM boundary
		profile := os.Getenv("COLIMA_ACTIVE_PROFILE")
		if profile == "" {
			profile = "default"
		}

		containerName := fmt.Sprintf("dvm-%s-%s", projectName, workspaceName)

		// Check if container already exists and is running
		checkCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
			"sudo", "nerdctl", "--namespace", "devopsmaestro", "ps", "-q", "-f", fmt.Sprintf("name=%s", containerName))
		output, _ := checkCmd.Output()

		if len(output) == 0 {
			// Container doesn't exist or isn't running - create it
			fmt.Println("Creating container...")

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
				fmt.Printf("Error: Failed to create container: %v\n", err)
				return
			}
		} else {
			fmt.Println("Container already running")
		}

		// Attach to container using nerdctl exec
		fmt.Println("Attaching to workspace...")
		attachCmd := exec.Command("colima", "--profile", profile, "ssh", "--",
			"sudo", "nerdctl", "--namespace", "devopsmaestro", "exec", "-it", containerName, "/bin/zsh", "-l")
		attachCmd.Stdin = os.Stdin
		attachCmd.Stdout = os.Stdout
		attachCmd.Stderr = os.Stderr

		if err := attachCmd.Run(); err != nil {
			fmt.Printf("Error: Failed to attach: %v\n", err)
			return
		}

		fmt.Println("Session ended.")
	},
}

// Initializes the attach command
func init() {
	rootCmd.AddCommand(attachCmd)
}
