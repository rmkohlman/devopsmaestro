package operators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/moby/term"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// ContainerdRuntimeV2 is a clean implementation using containerd v2 API
type ContainerdRuntimeV2 struct {
	client    *client.Client
	platform  *Platform
	namespace string
}

// NewContainerdRuntimeV2 creates a new containerd v2 runtime instance
func NewContainerdRuntimeV2() (*ContainerdRuntimeV2, error) {
	// Detect platform
	detector, err := NewPlatformDetector()
	if err != nil {
		return nil, err
	}

	platform, err := detector.Detect()
	if err != nil {
		return nil, err
	}

	return NewContainerdRuntimeV2WithPlatform(platform)
}

// NewContainerdRuntimeV2WithPlatform creates a new containerd v2 runtime with a specific platform
func NewContainerdRuntimeV2WithPlatform(platform *Platform) (*ContainerdRuntimeV2, error) {
	// Verify this platform supports containerd
	if !platform.IsContainerd() {
		return nil, fmt.Errorf("platform %s does not support containerd API directly. Use docker runtime instead", platform.Name)
	}

	// Set namespace for DVM containers
	namespace := "devopsmaestro"

	// Create containerd client
	ctdClient, err := client.New(platform.SocketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w\n%s", err, platform.GetStartHint())
	}

	// Verify connection
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	version, err := ctdClient.Version(ctx)
	if err != nil {
		ctdClient.Close()
		return nil, fmt.Errorf("failed to connect to containerd: %w\n%s", err, platform.GetStartHint())
	}

	fmt.Printf("Connected to containerd %s (%s, namespace: %s)\n",
		version.Version, platform.Name, namespace)

	return &ContainerdRuntimeV2{
		client:    ctdClient,
		platform:  platform,
		namespace: namespace,
	}, nil
}

// GetRuntimeType returns the runtime type
func (r *ContainerdRuntimeV2) GetRuntimeType() string {
	return "containerd-v2"
}

// BuildImage is not implemented (use BuildKit API instead)
func (r *ContainerdRuntimeV2) BuildImage(ctx context.Context, opts BuildOptions) error {
	return fmt.Errorf("use 'dvm build' command which uses BuildKit API")
}

// StartWorkspace creates and starts a workspace container
func (r *ContainerdRuntimeV2) StartWorkspace(ctx context.Context, opts StartOptions) (string, error) {
	// For Colima, use nerdctl via SSH to handle host path mounting correctly
	// The containerd API cannot handle macOS host paths directly since containerd
	// runs inside the Colima VM
	if r.platform.Type == PlatformColima {
		return r.startWorkspaceViaColima(ctx, opts)
	}

	// For other containerd platforms, use direct API (may need similar handling)
	return r.startWorkspaceDirectAPI(ctx, opts)
}

// startWorkspaceViaColima starts a workspace using nerdctl via colima ssh
// This handles the host path mounting correctly through Colima's mount system
func (r *ContainerdRuntimeV2) startWorkspaceViaColima(ctx context.Context, opts StartOptions) (string, error) {
	profile := r.platform.Profile
	if profile == "" {
		profile = "default"
	}

	// First check if container already exists and what image it's using
	containerName := opts.ComputeContainerName()

	// Check if container exists by trying to inspect it
	// We check for existence separately from the image label because
	// containers created before v0.18.18 don't have the image label
	existsCmd := fmt.Sprintf("sudo nerdctl --namespace %s inspect %s >/dev/null 2>&1 && echo EXISTS || echo NOTFOUND",
		r.namespace, containerName)
	existsExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", existsCmd)
	existsOutput, _ := existsExec.Output()
	containerExists := strings.TrimSpace(string(existsOutput)) == "EXISTS"

	if containerExists {
		// Get the image the existing container was created with from our label
		// Containers created before v0.18.18 won't have this label
		imageCmd := fmt.Sprintf("sudo nerdctl --namespace %s inspect -f '{{index .Config.Labels \"io.devopsmaestro.image\"}}' %s 2>/dev/null || echo ''",
			r.namespace, containerName)
		imageExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", imageCmd)
		imageOutput, _ := imageExec.Output()
		existingImage := strings.TrimSpace(string(imageOutput))

		// If no label (pre-v0.18.18 container) or image changed, recreate the container
		hasImageLabel := existingImage != "" && existingImage != "''" && existingImage != "<no value>"
		imageChanged := hasImageLabel && existingImage != opts.ImageName

		if !hasImageLabel {
			fmt.Printf("Container exists without image label (pre-v0.18.18), recreating...\n")
			// Remove old container to recreate with proper labels
			rmCmd := fmt.Sprintf("sudo nerdctl --namespace %s rm -f %s 2>/dev/null || true",
				r.namespace, containerName)
			rmExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", rmCmd)
			rmExec.Run()
			// Fall through to create new container
		} else if imageChanged {
			fmt.Printf("Image changed: %s -> %s\n", existingImage, opts.ImageName)
			fmt.Printf("Recreating container with new image...\n")

			// Stop and remove the old container regardless of state
			rmCmd := fmt.Sprintf("sudo nerdctl --namespace %s rm -f %s 2>/dev/null || true",
				r.namespace, containerName)
			rmExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", rmCmd)
			rmExec.Run()

			// Fall through to create new container
		} else {
			// Same image - check if it's running
			statusCmd := fmt.Sprintf("sudo nerdctl --namespace %s inspect -f '{{.State.Status}}' %s 2>/dev/null || echo stopped",
				r.namespace, containerName)
			statusExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", statusCmd)
			statusOutput, _ := statusExec.Output()
			status := strings.TrimSpace(string(statusOutput))

			if status == "running" {
				// Already running with same image, return the container name as ID
				return containerName, nil
			}

			// Container exists but not running, try to start it
			startCmd := fmt.Sprintf("sudo nerdctl --namespace %s start %s 2>/dev/null",
				r.namespace, containerName)
			startExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", startCmd)
			if err := startExec.Run(); err == nil {
				return containerName, nil
			}

			// Start failed, remove and recreate
			rmCmd := fmt.Sprintf("sudo nerdctl --namespace %s rm -f %s 2>/dev/null || true",
				r.namespace, containerName)
			rmExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", rmCmd)
			rmExec.Run()
		}
	}

	// Build nerdctl run command
	// Set command to keep container running using helper
	command := opts.ComputeCommand()

	// Set default working directory
	workingDir := opts.WorkingDir
	if workingDir == "" {
		workingDir = "/workspace"
	}

	// Build the nerdctl run command parts
	nerdctlArgs := []string{
		"sudo", "nerdctl",
		"--namespace", r.namespace,
		"run", "-d",
		"--name", containerName,
		"-w", workingDir,
	}

	// Add volume mounts
	// nerdctl handles path translation for Colima's mounted volumes
	nerdctlArgs = append(nerdctlArgs, "-v", fmt.Sprintf("%s:/workspace", opts.AppPath))

	// Add SSH key mount if directory exists
	homeDir, _ := os.UserHomeDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	if _, err := os.Stat(sshDir); err == nil {
		nerdctlArgs = append(nerdctlArgs, "-v", fmt.Sprintf("%s:/root/.ssh:ro", sshDir))
	}

	// Add environment variables
	for k, v := range opts.Env {
		nerdctlArgs = append(nerdctlArgs, "-e", fmt.Sprintf("%s=%s", k, v))
	}

	// Add labels for DVM management
	nerdctlArgs = append(nerdctlArgs,
		"--label", "io.devopsmaestro.managed=true",
		"--label", fmt.Sprintf("io.devopsmaestro.app=%s", opts.AppName),
		"--label", fmt.Sprintf("io.devopsmaestro.workspace=%s", opts.WorkspaceName),
		"--label", fmt.Sprintf("io.devopsmaestro.image=%s", opts.ImageName),
	)

	// Add image and command
	nerdctlArgs = append(nerdctlArgs, opts.ImageName)
	nerdctlArgs = append(nerdctlArgs, command...)

	// Build the full command string for SSH
	cmdStr := ""
	for i, arg := range nerdctlArgs {
		if i > 0 {
			cmdStr += " "
		}
		// Quote arguments that contain spaces or special characters
		if needsQuoting(arg) {
			cmdStr += fmt.Sprintf("'%s'", arg)
		} else {
			cmdStr += arg
		}
	}

	// Execute via colima ssh
	execCmd := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", cmdStr)
	execCmd.Stderr = os.Stderr

	output, err := execCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to start workspace via nerdctl: %w", err)
	}

	// nerdctl run -d returns the container ID
	containerID := string(output)
	if len(containerID) > 12 {
		containerID = containerID[:12]
	}

	// Return the container name as the ID (that's what we use to reference it)
	return containerName, nil
}

// needsQuoting returns true if a string needs shell quoting
func needsQuoting(s string) bool {
	for _, c := range s {
		if c == ' ' || c == '\'' || c == '"' || c == '=' || c == '$' || c == '\\' || c == '!' || c == '*' {
			return true
		}
	}
	return false
}

// startWorkspaceDirectAPI starts a workspace using the containerd API directly
// This is the original implementation, kept for non-Colima platforms
func (r *ContainerdRuntimeV2) startWorkspaceDirectAPI(ctx context.Context, opts StartOptions) (string, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	// Get the image
	image, err := r.client.GetImage(ctx, opts.ImageName)
	if err != nil {
		return "", fmt.Errorf("image not found: %s\nRun 'dvm build' first", opts.ImageName)
	}

	// Check if container already exists
	containerName := opts.ComputeContainerName()
	existingContainer, err := r.client.LoadContainer(ctx, containerName)
	if err == nil {
		// Container exists - clean it up
		task, err := existingContainer.Task(ctx, nil)
		if err == nil {
			// Task exists - kill and delete it
			status, err := task.Status(ctx)
			if err == nil && status.Status == "running" {
				// Container is already running - return its ID
				return existingContainer.ID(), nil
			}

			// Task exists but not running, try to kill it first
			task.Kill(ctx, syscall.SIGKILL)
			task.Wait(ctx) // Wait for it to die
			task.Delete(ctx, client.WithProcessKill)
		}
		// Delete the existing container with snapshot cleanup
		existingContainer.Delete(ctx, client.WithSnapshotCleanup)
	}

	// Set command to keep container running using helper
	command := opts.ComputeCommand()

	// Set default working directory
	workingDir := opts.WorkingDir
	if workingDir == "" {
		workingDir = "/workspace"
	}

	// Convert env map to slice
	envSlice := make([]string, 0, len(opts.Env))
	for k, v := range opts.Env {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", k, v))
	}

	// Get home directory for SSH keys
	homeDir, _ := os.UserHomeDir()
	sshDir := filepath.Join(homeDir, ".ssh")

	// Create mount specs
	mounts := []specs.Mount{
		{
			Source:      opts.AppPath,
			Destination: "/workspace",
			Type:        "bind",
			Options:     []string{"rbind", "rw"},
		},
	}

	// Add SSH key mount if directory exists
	if _, err := os.Stat(sshDir); err == nil {
		mounts = append(mounts, specs.Mount{
			Source:      sshDir,
			Destination: "/root/.ssh",
			Type:        "bind",
			Options:     []string{"rbind", "ro"},
		})
	}

	// Create container with proper OCI spec
	// Use Compose to combine base spec with customizations
	container, err := r.client.NewContainer(
		ctx,
		containerName,
		client.WithImage(image),
		client.WithSnapshotter("overlayfs"),
		client.WithNewSnapshot(containerName+"-snapshot", image),
		client.WithRuntime("io.containerd.runc.v2", nil), // Explicitly use runc runtime
		client.WithNewSpec(
			oci.Compose(
				oci.WithDefaultSpec(),
				oci.WithDefaultUnixDevices,
				oci.WithImageConfig(image),
				oci.WithProcessArgs(command...),
				oci.WithProcessCwd(workingDir),
				oci.WithEnv(envSlice),
				oci.WithMounts(mounts),
				oci.WithUserID(0), // Run as root
				oci.WithUsername("root"),
				// Don't set TTY here - nerdctl exec will handle TTY allocation
			),
		),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Create the task (starts the container)
	// Use NullIO for now - we'll attach with proper IO later
	task, err := container.NewTask(ctx, cio.NullIO)
	if err != nil {
		container.Delete(ctx, client.WithSnapshotCleanup)
		return "", fmt.Errorf("failed to create task: %w", err)
	}

	// Start the task
	if err := task.Start(ctx); err != nil {
		task.Delete(ctx)
		container.Delete(ctx, client.WithSnapshotCleanup)
		return "", fmt.Errorf("failed to start task: %w", err)
	}

	return container.ID(), nil
}

// AttachToWorkspace attaches to a running workspace container
// For Colima, we use nerdctl via SSH since direct FIFO-based IO doesn't work across host/VM boundary
func (r *ContainerdRuntimeV2) AttachToWorkspace(ctx context.Context, opts AttachOptions) error {
	// For Colima, use nerdctl via SSH (containerd API doesn't work well across VM boundary)
	if r.platform.Type == PlatformColima {
		return r.attachViaColima(ctx, opts)
	}

	// For other platforms, use containerd API directly
	return r.attachDirectAPI(ctx, opts)
}

// attachViaColima attaches to a container using nerdctl via SSH for Colima
func (r *ContainerdRuntimeV2) attachViaColima(ctx context.Context, opts AttachOptions) error {
	profile := r.platform.Profile
	if profile == "" {
		profile = "default"
	}

	// Check if container exists and is running via nerdctl
	statusCmd := fmt.Sprintf("sudo nerdctl --namespace %s inspect -f '{{.State.Status}}' %s 2>/dev/null || echo not_found",
		r.namespace, opts.WorkspaceID)
	statusExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", statusCmd)
	statusOutput, err := statusExec.Output()
	if err != nil {
		return fmt.Errorf("failed to check container status: %w", err)
	}

	status := strings.TrimSpace(string(statusOutput))
	if status == "not_found" {
		return fmt.Errorf("container not found: %s", opts.WorkspaceID)
	}

	if status != "running" {
		return fmt.Errorf("container is not running (status: %s)", status)
	}

	// Build shell command with options
	shell := opts.Shell
	if shell == "" {
		shell = "/bin/zsh"
	}

	// Start building the nerdctl exec command
	var cmdParts []string
	cmdParts = append(cmdParts, "sudo", "nerdctl", "--namespace", r.namespace, "exec", "-it")

	// Add environment variables
	for key, value := range opts.Env {
		cmdParts = append(cmdParts, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add container name
	cmdParts = append(cmdParts, opts.WorkspaceID)

	// Add shell and login flag
	cmdParts = append(cmdParts, shell)
	if opts.LoginShell {
		cmdParts = append(cmdParts, "-l")
	}

	// Convert to command string for SSH execution
	cmd := strings.Join(cmdParts, " ")
	execCmd := []string{"colima", "--profile", profile, "ssh", "--", "sh", "-c", cmd}

	// Find colima in PATH
	colimaPath, err := exec.LookPath("colima")
	if err != nil {
		return fmt.Errorf("colima not found in PATH: %w", err)
	}

	// Create exec command
	execProc := &exec.Cmd{
		Path:   colimaPath,
		Args:   execCmd,
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
	}

	// Run the command
	if err := execProc.Run(); err != nil {
		return fmt.Errorf("failed to attach: %w", err)
	}

	return nil
}

// attachDirectAPI attaches to a container using containerd API directly
func (r *ContainerdRuntimeV2) attachDirectAPI(ctx context.Context, opts AttachOptions) error {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	// Load the container
	container, err := r.client.LoadContainer(ctx, opts.WorkspaceID)
	if err != nil {
		return fmt.Errorf("container not found: %w", err)
	}

	// Get the task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Check if task is running
	status, err := task.Status(ctx)
	if err != nil {
		return fmt.Errorf("failed to get task status: %w", err)
	}

	if status.Status != "running" {
		return fmt.Errorf("container is not running (status: %s)", status.Status)
	}

	// For other platforms, return not implemented for now
	return fmt.Errorf("attach not implemented for platform %s with containerd runtime", r.platform.Name)
}

// StopWorkspace stops a running workspace
func (r *ContainerdRuntimeV2) StopWorkspace(ctx context.Context, containerID string) error {
	// For Colima, use nerdctl via SSH (containerd API doesn't work well across VM boundary)
	if r.platform.Type == PlatformColima {
		return r.stopViaColima(ctx, containerID)
	}

	// For other platforms, use containerd API directly
	return r.stopDirectAPI(ctx, containerID)
}

// stopViaColima stops a container using nerdctl via SSH for Colima
func (r *ContainerdRuntimeV2) stopViaColima(ctx context.Context, containerID string) error {
	profile := r.platform.Profile
	if profile == "" {
		profile = "default"
	}

	// Stop the container using nerdctl
	stopCmd := fmt.Sprintf("sudo nerdctl --namespace %s stop %s 2>/dev/null || true", r.namespace, containerID)
	stopExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", stopCmd)
	if err := stopExec.Run(); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	return nil
}

// stopDirectAPI stops a container using containerd API directly
func (r *ContainerdRuntimeV2) stopDirectAPI(ctx context.Context, containerID string) error {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	container, err := r.client.LoadContainer(ctx, containerID)
	if err != nil {
		return fmt.Errorf("container not found: %w", err)
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		// No task means container is not running
		return nil
	}

	// Kill the task
	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to kill task: %w", err)
	}

	// Wait for task to exit
	statusC, err := task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}

	<-statusC

	// Delete the task
	if _, err := task.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	return nil
}

// GetWorkspaceStatus returns the status of a workspace
func (r *ContainerdRuntimeV2) GetWorkspaceStatus(ctx context.Context, containerID string) (string, error) {
	// For Colima, use nerdctl via SSH (containerd API doesn't work well across VM boundary)
	if r.platform.Type == PlatformColima {
		return r.getStatusViaColima(ctx, containerID)
	}

	// For other platforms, use containerd API directly
	return r.getStatusDirectAPI(ctx, containerID)
}

// getStatusViaColima gets container status using nerdctl via SSH for Colima
func (r *ContainerdRuntimeV2) getStatusViaColima(ctx context.Context, containerID string) (string, error) {
	profile := r.platform.Profile
	if profile == "" {
		profile = "default"
	}

	// Check if container exists and get its status via nerdctl
	statusCmd := fmt.Sprintf("sudo nerdctl --namespace %s inspect -f '{{.State.Status}}' %s 2>/dev/null || echo not_found",
		r.namespace, containerID)
	statusExec := exec.CommandContext(ctx, "colima", "--profile", profile, "ssh", "--", "sh", "-c", statusCmd)
	statusOutput, err := statusExec.Output()
	if err != nil {
		return "unknown", fmt.Errorf("failed to check container status: %w", err)
	}

	status := strings.TrimSpace(string(statusOutput))
	if status == "not_found" {
		return "not_found", nil
	}

	return status, nil
}

// getStatusDirectAPI gets container status using containerd API directly
func (r *ContainerdRuntimeV2) getStatusDirectAPI(ctx context.Context, containerID string) (string, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	container, err := r.client.LoadContainer(ctx, containerID)
	if err != nil {
		return "not_found", nil
	}

	task, err := container.Task(ctx, nil)
	if err != nil {
		return "created", nil
	}

	status, err := task.Status(ctx)
	if err != nil {
		return "unknown", err
	}

	return string(status.Status), nil
}

// Close closes the containerd client
func (r *ContainerdRuntimeV2) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

// GetPlatformName returns the human-readable platform name
func (r *ContainerdRuntimeV2) GetPlatformName() string {
	return r.platform.Name
}

// ListWorkspaces lists all DVM-managed workspaces
func (r *ContainerdRuntimeV2) ListWorkspaces(ctx context.Context) ([]WorkspaceInfo, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	containers, err := r.client.Containers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var workspaces []WorkspaceInfo
	for _, c := range containers {
		labels, err := c.Labels(ctx)
		if err != nil {
			continue
		}

		// Check for DVM management label
		if labels["io.devopsmaestro.managed"] != "true" {
			continue
		}

		// Get status
		status := "created"
		task, err := c.Task(ctx, nil)
		if err == nil {
			taskStatus, err := task.Status(ctx)
			if err == nil {
				status = string(taskStatus.Status)
			}
		}

		// Get image
		image, _ := c.Image(ctx)
		imageName := ""
		if image != nil {
			imageName = image.Name()
		}

		workspaces = append(workspaces, WorkspaceInfo{
			ID:        c.ID()[:12],
			Name:      c.ID(), // containerd uses ID as name
			Status:    status,
			Image:     imageName,
			App:       labels["io.devopsmaestro.app"],
			Workspace: labels["io.devopsmaestro.workspace"],
			Labels:    labels,
		})
	}

	return workspaces, nil
}

// FindWorkspace finds a workspace by name and returns its info
func (r *ContainerdRuntimeV2) FindWorkspace(ctx context.Context, name string) (*WorkspaceInfo, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	container, err := r.client.LoadContainer(ctx, name)
	if err != nil {
		return nil, nil // Not found
	}

	labels, err := container.Labels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	// Verify it's DVM-managed
	if labels["io.devopsmaestro.managed"] != "true" {
		return nil, nil
	}

	// Get status
	status := "created"
	task, err := container.Task(ctx, nil)
	if err == nil {
		taskStatus, err := task.Status(ctx)
		if err == nil {
			status = string(taskStatus.Status)
		}
	}

	// Get image
	image, _ := container.Image(ctx)
	imageName := ""
	if image != nil {
		imageName = image.Name()
	}

	return &WorkspaceInfo{
		ID:        container.ID()[:12],
		Name:      container.ID(),
		Status:    status,
		Image:     imageName,
		App:       labels["io.devopsmaestro.app"],
		Workspace: labels["io.devopsmaestro.workspace"],
		Labels:    labels,
	}, nil
}

// StopAllWorkspaces stops all DVM-managed workspaces
func (r *ContainerdRuntimeV2) StopAllWorkspaces(ctx context.Context) (int, error) {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	workspaces, err := r.ListWorkspaces(ctx)
	if err != nil {
		return 0, err
	}

	stopped := 0
	for _, ws := range workspaces {
		if ws.Status == "running" {
			if err := r.StopWorkspace(ctx, ws.Name); err == nil {
				stopped++
			}
		}
	}

	return stopped, nil
}

// setupTerminal sets up the terminal for raw mode
func setupTerminal() error {
	fd := os.Stdin.Fd()
	state, err := term.SetRawTerminal(fd)
	if err != nil {
		return err
	}
	// Store state for restoration
	terminalState = state
	return nil
}

// restoreTerminal restores the terminal to its original state
func restoreTerminal() {
	if terminalState != nil {
		fd := os.Stdin.Fd()
		term.RestoreTerminal(fd, terminalState)
		terminalState = nil
	}
}

var terminalState *term.State
