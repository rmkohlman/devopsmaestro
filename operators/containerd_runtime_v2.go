package operators

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	profile   string
	namespace string
}

// NewContainerdRuntimeV2 creates a new containerd v2 runtime instance
func NewContainerdRuntimeV2() (*ContainerdRuntimeV2, error) {
	// Detect active Colima profile
	profile := os.Getenv("COLIMA_ACTIVE_PROFILE")
	if profile == "" {
		profile = os.Getenv("COLIMA_DOCKER_PROFILE")
	}
	if profile == "" {
		profile = "default"
	}

	// Set namespace for DVM containers
	namespace := "devopsmaestro"

	// Build socket path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	socketPath := filepath.Join(homeDir, ".colima", profile, "containerd.sock")
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("containerd socket not found at %s\nMake sure Colima profile '%s' is running: colima start %s", socketPath, profile, profile)
	}

	// Create containerd client
	ctdClient, err := client.New(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client: %w", err)
	}

	// Verify connection
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	version, err := ctdClient.Version(ctx)
	if err != nil {
		ctdClient.Close()
		return nil, fmt.Errorf("failed to connect to containerd: %w", err)
	}

	fmt.Printf("Connected to containerd %s (profile: %s, namespace: %s)\n",
		version.Version, profile, namespace)

	return &ContainerdRuntimeV2{
		client:    ctdClient,
		profile:   profile,
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
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	// Get the image
	image, err := r.client.GetImage(ctx, opts.ImageName)
	if err != nil {
		return "", fmt.Errorf("image not found: %s\nRun 'dvm build' first", opts.ImageName)
	}

	// Check if container already exists
	existingContainer, err := r.client.LoadContainer(ctx, opts.WorkspaceName)
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

	// Set default command
	command := opts.Command
	if len(command) == 0 {
		command = []string{"/bin/zsh", "-l"}
	}

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
			Source:      opts.ProjectPath,
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
		opts.WorkspaceName,
		client.WithImage(image),
		client.WithSnapshotter("overlayfs"),
		client.WithNewSnapshot(opts.WorkspaceName+"-snapshot", image),
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
func (r *ContainerdRuntimeV2) AttachToWorkspace(ctx context.Context, containerID string) error {
	ctx = namespaces.WithNamespace(ctx, r.namespace)

	// Load the container
	container, err := r.client.LoadContainer(ctx, containerID)
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

	// Use nerdctl exec via SSH for interactive attach (works with Colima)
	// This avoids FIFO issues when client is on host and containerd is in VM
	cmd := fmt.Sprintf("sudo nerdctl --namespace %s exec -it %s /bin/zsh -l", r.namespace, containerID)
	execCmd := []string{"colima", "--profile", r.profile, "ssh", "-t", "--", "sh", "-c", cmd}

	// Create exec command
	execProc := &exec.Cmd{
		Path:   "/usr/bin/colima",
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

// StopWorkspace stops a running workspace
func (r *ContainerdRuntimeV2) StopWorkspace(ctx context.Context, containerID string) error {
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
