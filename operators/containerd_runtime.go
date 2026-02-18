package operators

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/core/containers"
	"github.com/containerd/containerd/v2/pkg/cio"
	"github.com/containerd/containerd/v2/pkg/namespaces"
	"github.com/containerd/containerd/v2/pkg/oci"
	"github.com/moby/term"
	specs "github.com/opencontainers/runtime-spec/specs-go"
)

// ContainerdRuntime implements ContainerRuntime for containerd (Colima-compatible)
type ContainerdRuntime struct {
	client    *client.Client
	profile   string
	namespace string
}

// NewContainerdRuntime creates a new containerd runtime instance
// It automatically detects and connects to the active Colima profile's containerd socket
func NewContainerdRuntime() (*ContainerdRuntime, error) {
	// Detect active Colima profile
	profile := os.Getenv("COLIMA_DOCKER_PROFILE")
	if profile == "" {
		profile = os.Getenv("COLIMA_ACTIVE_PROFILE")
	}
	if profile == "" {
		profile = "default"
	}

	// Set namespace for DVM containers (isolated from other workloads)
	namespace := "devopsmaestro"

	// Build socket path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	socketPath := filepath.Join(homeDir, ".colima", profile, "client.sock")
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("containerd socket not found at %s\nMake sure Colima profile '%s' is running: colima start %s", socketPath, profile, profile)
	}

	// Create containerd client
	client, err := client.New(socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create containerd client for profile '%s': %w", profile, err)
	}

	// Verify connection by checking version
	ctx := namespaces.WithNamespace(context.Background(), namespace)
	version, err := client.Version(ctx)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to containerd (profile '%s'): %w", profile, err)
	}

	fmt.Printf("Connected to containerd %s (profile: %s, namespace: %s)\n", version.Version, profile, namespace)

	return &ContainerdRuntime{
		client:    client,
		profile:   profile,
		namespace: namespace,
	}, nil
}

// BuildImage builds a container image (not implemented in MVP)
func (c *ContainerdRuntime) BuildImage(ctx context.Context, opts BuildOptions) error {
	return fmt.Errorf("image building not yet supported with containerd runtime\n\n"+
		"Please build manually using nerdctl:\n"+
		"  colima nerdctl --profile %s -- build --namespace %s -t %s %s\n\n"+
		"Or use the docker alias after running 'exec zsh':\n"+
		"  cd %s\n"+
		"  docker build -t %s .",
		c.profile, c.namespace, opts.ImageName, opts.BuildContext,
		opts.BuildContext, opts.ImageName)
}

// StartWorkspace starts a containerd container as a workspace
func (c *ContainerdRuntime) StartWorkspace(ctx context.Context, opts StartOptions) (string, error) {
	fmt.Printf("Starting workspace '%s' in containerd (profile: %s, namespace: %s)...\n",
		opts.WorkspaceName, c.profile, c.namespace)

	// Use namespace context
	ctx = namespaces.WithNamespace(ctx, c.namespace)

	// Normalize image name - containerd needs docker.io prefix for local images
	imageName := opts.ImageName
	if !strings.Contains(imageName, "/") {
		// Simple name without registry/repo, try docker.io/library prefix
		imageName = "docker.io/library/" + imageName
	} else if !strings.Contains(imageName, ".") {
		// Has slash but no dot (no registry), add docker.io prefix
		imageName = "docker.io/" + imageName
	}

	// Try to get the image with normalized name
	image, err := c.client.GetImage(ctx, imageName)
	if err != nil {
		// Try original name without prefix
		image, err = c.client.GetImage(ctx, opts.ImageName)
		if err != nil {
			fmt.Printf("Image '%s' not found locally, attempting to pull...\n", opts.ImageName)
			image, err = c.client.Pull(ctx, imageName, client.WithPullUnpack)
			if err != nil {
				return "", fmt.Errorf("failed to pull image '%s': %w\nMake sure the image exists or build it manually", opts.ImageName, err)
			}
			fmt.Printf("✓ Image '%s' pulled successfully\n", opts.ImageName)
		}
	}

	// Set default command if not specified
	command := opts.Command
	if len(command) == 0 {
		command = []string{"/bin/zsh"}
	}

	// Set default working directory
	workingDir := opts.WorkingDir
	if workingDir == "" {
		workingDir = "/workspace"
	}

	// Convert env map to slice
	envSlice := envMapToSliceContainerd(opts.Env)

	// Clean up any existing container with the same name
	if existingContainer, err := c.client.LoadContainer(ctx, opts.ContainerName); err == nil {
		// Container exists, delete it
		if task, err := existingContainer.Task(ctx, nil); err == nil {
			// Task exists, kill it first
			task.Kill(ctx, syscall.SIGKILL)
			task.Wait(ctx)
		}
		existingContainer.Delete(ctx, client.WithSnapshotCleanup)
	}

	// Create container with OCI spec
	// Manually create a minimal Linux spec since oci helpers don't always work
	withLinuxSpec := func(ctx context.Context, client oci.Client, c *containers.Container, s *specs.Spec) error {
		if s.Linux == nil {
			s.Linux = &specs.Linux{
				Resources: &specs.LinuxResources{
					Devices: []specs.LinuxDeviceCgroup{
						// Deny all by default
						{Allow: false, Access: "rwm"},
						// Allow specific devices (matching nerdctl defaults)
						{Allow: true, Type: "c", Major: toPtr(int64(1)), Minor: toPtr(int64(3)), Access: "rwm"}, // /dev/null
						{Allow: true, Type: "c", Major: toPtr(int64(1)), Minor: toPtr(int64(8)), Access: "rwm"}, // /dev/random
						{Allow: true, Type: "c", Major: toPtr(int64(1)), Minor: toPtr(int64(7)), Access: "rwm"}, // /dev/full
						{Allow: true, Type: "c", Major: toPtr(int64(5)), Minor: toPtr(int64(0)), Access: "rwm"}, // /dev/tty
						{Allow: true, Type: "c", Major: toPtr(int64(1)), Minor: toPtr(int64(5)), Access: "rwm"}, // /dev/zero
						{Allow: true, Type: "c", Major: toPtr(int64(1)), Minor: toPtr(int64(9)), Access: "rwm"}, // /dev/urandom
						{Allow: true, Type: "c", Major: toPtr(int64(5)), Minor: toPtr(int64(1)), Access: "rwm"}, // /dev/console
						{Allow: true, Type: "c", Major: toPtr(int64(136)), Access: "rwm"},                       // /dev/pts/*
						{Allow: true, Type: "c", Major: toPtr(int64(5)), Minor: toPtr(int64(2)), Access: "rwm"}, // /dev/ptmx
					},
				},
				Namespaces: []specs.LinuxNamespace{
					{Type: specs.PIDNamespace},
					{Type: specs.IPCNamespace},
					{Type: specs.UTSNamespace},
					{Type: specs.MountNamespace},
					{Type: specs.NetworkNamespace},
					{Type: specs.CgroupNamespace},
				},
			}
		}
		if s.Process == nil {
			s.Process = &specs.Process{}
		}
		if s.Process.User.AdditionalGids == nil {
			// Add additional groups (matching nerdctl defaults)
			s.Process.User.AdditionalGids = []uint32{0, 1, 2, 3, 4, 6, 10, 11, 20, 26, 27}
		}
		if s.Process.Capabilities == nil {
			// Add basic capabilities
			s.Process.Capabilities = &specs.LinuxCapabilities{
				Bounding: []string{
					"CAP_CHOWN",
					"CAP_DAC_OVERRIDE",
					"CAP_FSETID",
					"CAP_FOWNER",
					"CAP_MKNOD",
					"CAP_NET_RAW",
					"CAP_SETGID",
					"CAP_SETUID",
					"CAP_SETFCAP",
					"CAP_SETPCAP",
					"CAP_NET_BIND_SERVICE",
					"CAP_SYS_CHROOT",
					"CAP_KILL",
					"CAP_AUDIT_WRITE",
				},
				Effective: []string{
					"CAP_CHOWN",
					"CAP_DAC_OVERRIDE",
					"CAP_FSETID",
					"CAP_FOWNER",
					"CAP_MKNOD",
					"CAP_NET_RAW",
					"CAP_SETGID",
					"CAP_SETUID",
					"CAP_SETFCAP",
					"CAP_SETPCAP",
					"CAP_NET_BIND_SERVICE",
					"CAP_SYS_CHROOT",
					"CAP_KILL",
					"CAP_AUDIT_WRITE",
				},
				Permitted: []string{
					"CAP_CHOWN",
					"CAP_DAC_OVERRIDE",
					"CAP_FSETID",
					"CAP_FOWNER",
					"CAP_MKNOD",
					"CAP_NET_RAW",
					"CAP_SETGID",
					"CAP_SETUID",
					"CAP_SETFCAP",
					"CAP_SETPCAP",
					"CAP_NET_BIND_SERVICE",
					"CAP_SYS_CHROOT",
					"CAP_KILL",
					"CAP_AUDIT_WRITE",
				},
			}
		}
		// Add /dev tmpfs mount and other critical mounts
		if s.Mounts == nil {
			s.Mounts = []specs.Mount{}
		}
		// Prepend critical system mounts
		criticalMounts := []specs.Mount{
			{
				Destination: "/proc",
				Type:        "proc",
				Source:      "proc",
				Options:     []string{"nosuid", "noexec", "nodev"},
			},
			{
				Destination: "/dev",
				Type:        "tmpfs",
				Source:      "tmpfs",
				Options:     []string{"nosuid", "strictatime", "mode=755", "size=65536k"},
			},
			{
				Destination: "/dev/pts",
				Type:        "devpts",
				Source:      "devpts",
				Options:     []string{"nosuid", "noexec", "newinstance", "ptmxmode=0666", "mode=0620", "gid=5"},
			},
			{
				Destination: "/dev/shm",
				Type:        "tmpfs",
				Source:      "shm",
				Options:     []string{"nosuid", "noexec", "nodev", "mode=1777", "size=65536k"},
			},
			{
				Destination: "/sys",
				Type:        "sysfs",
				Source:      "sysfs",
				Options:     []string{"nosuid", "noexec", "nodev", "ro"},
			},
		}
		s.Mounts = append(criticalMounts, s.Mounts...)

		// Set runtime name
		c.Runtime = containers.RuntimeInfo{
			Name: "io.client.runc.v2",
		}
		return nil
	}

	container, err := c.client.NewContainer(
		ctx,
		opts.ContainerName,
		client.WithImage(image),
		client.WithSnapshotter("overlayfs"),
		client.WithNewSnapshot(opts.ContainerName+"-snapshot", image),
		client.WithNewSpec(
			withLinuxSpec, // Apply this first to create the Linux section
			oci.WithImageConfigArgs(image, command),
			oci.WithDefaultUnixDevices, // Add default devices like /dev/null
			oci.WithHostname(opts.ContainerName),
			oci.WithProcessCwd(workingDir),
			oci.WithEnv(envSlice),
			// Skip TTY for now since we're using NullIO
			oci.WithMounts([]specs.Mount{
				{
					Source:      opts.AppPath,
					Destination: "/workspace",
					Type:        "bind",
					Options:     []string{"rbind", "rw"},
				},
			}),
		),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Create task (start container)
	// Use cio.NullIO since we're connecting remotely and will attach separately
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

	fmt.Printf("✓ Workspace started (Container: %s)\n", container.ID())
	return container.ID(), nil
}

// AttachToWorkspace attaches an interactive terminal to a running workspace
func (c *ContainerdRuntime) AttachToWorkspace(ctx context.Context, workspaceID string) error {
	fmt.Printf("Attaching to workspace (press Ctrl+D to exit)...\n")

	// Use namespace context
	ctx = namespaces.WithNamespace(ctx, c.namespace)

	// Load container
	container, err := c.client.LoadContainer(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to load container '%s': %w", workspaceID, err)
	}

	// Get task
	task, err := container.Task(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to get task: %w", err)
	}

	// Create exec process for interactive shell
	spec, err := container.Spec(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container spec: %w", err)
	}

	// Get current terminal size
	ws, err := term.GetWinsize(os.Stdin.Fd())
	if err != nil {
		// Default size if we can't get it
		ws = &term.Winsize{Height: 24, Width: 80}
	}

	// Create exec process
	execID := fmt.Sprintf("%s-exec", workspaceID)
	execSpec := &specs.Process{
		Args:     []string{"/bin/zsh"},
		Cwd:      spec.Process.Cwd,
		Env:      spec.Process.Env,
		Terminal: true,
	}
	if ws != nil {
		execSpec.ConsoleSize = &specs.Box{
			Width:  uint(ws.Width),
			Height: uint(ws.Height),
		}
	}

	// Create exec process
	process, err := task.Exec(ctx, execID, execSpec, cio.NewCreator(cio.WithStreams(os.Stdin, os.Stdout, os.Stderr)))
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}
	defer process.Delete(ctx)

	// Put terminal in raw mode
	oldState, err := term.SetRawTerminal(os.Stdin.Fd())
	if err != nil {
		return fmt.Errorf("failed to set raw terminal: %w", err)
	}
	defer term.RestoreTerminal(os.Stdin.Fd(), oldState)

	// Start exec
	if err := process.Start(ctx); err != nil {
		return fmt.Errorf("failed to start exec: %w", err)
	}

	// Wait for exec to finish
	statusC, err := process.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for exec: %w", err)
	}

	// Monitor for terminal resize
	go func() {
		for {
			ws, err := term.GetWinsize(os.Stdin.Fd())
			if err != nil {
				return
			}
			if err := process.Resize(ctx, uint32(ws.Width), uint32(ws.Height)); err != nil {
				return
			}
		}
	}()

	// Wait for process to complete
	status := <-statusC
	code, _, err := status.Result()
	if err != nil {
		return fmt.Errorf("exec failed: %w", err)
	}

	if code != 0 {
		fmt.Printf("\nProcess exited with code %d\n", code)
	}

	fmt.Printf("✓ Detached from workspace\n")
	return nil
}

// StopWorkspace stops a running workspace
func (c *ContainerdRuntime) StopWorkspace(ctx context.Context, workspaceID string) error {
	fmt.Printf("Stopping workspace '%s'...\n", workspaceID)

	// Use namespace context
	ctx = namespaces.WithNamespace(ctx, c.namespace)

	// Load container
	container, err := c.client.LoadContainer(ctx, workspaceID)
	if err != nil {
		return fmt.Errorf("failed to load container: %w", err)
	}

	// Get task
	task, err := container.Task(ctx, nil)
	if err != nil {
		// Task might not exist, try to delete container anyway
		if err := container.Delete(ctx, client.WithSnapshotCleanup); err != nil {
			return fmt.Errorf("failed to delete container: %w", err)
		}
		fmt.Printf("✓ Workspace stopped (container deleted)\n")
		return nil
	}

	// Kill task
	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to kill task: %w", err)
	}

	// Wait for task to exit
	statusC, err := task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to wait for task: %w", err)
	}

	// Wait with timeout
	select {
	case <-statusC:
		// Task exited
	case <-ctx.Done():
		// Force kill if timeout
		task.Kill(ctx, syscall.SIGKILL)
		<-statusC
	}

	// Delete task
	if _, err := task.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	// Delete container
	if err := container.Delete(ctx, client.WithSnapshotCleanup); err != nil {
		return fmt.Errorf("failed to delete container: %w", err)
	}

	fmt.Printf("✓ Workspace stopped\n")
	return nil
}

// GetWorkspaceStatus returns the status of a workspace
func (c *ContainerdRuntime) GetWorkspaceStatus(ctx context.Context, workspaceID string) (string, error) {
	// Use namespace context
	ctx = namespaces.WithNamespace(ctx, c.namespace)

	// Load container
	container, err := c.client.LoadContainer(ctx, workspaceID)
	if err != nil {
		return "unknown", fmt.Errorf("failed to load container: %w", err)
	}

	// Get task
	task, err := container.Task(ctx, nil)
	if err != nil {
		// No task means container exists but not running
		return "stopped", nil
	}

	// Get task status
	status, err := task.Status(ctx)
	if err != nil {
		return "unknown", fmt.Errorf("failed to get task status: %w", err)
	}

	switch status.Status {
	case client.Running:
		return "running", nil
	case client.Paused:
		return "paused", nil
	case client.Stopped:
		return "stopped", nil
	default:
		return string(status.Status), nil
	}
}

// GetRuntimeType returns "containerd"
func (c *ContainerdRuntime) GetRuntimeType() string {
	return "containerd"
}

// Close closes the containerd client connection
func (c *ContainerdRuntime) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

// Helper function to convert env map to slice for containerd
func envMapToSliceContainerd(envMap map[string]string) []string {
	var envSlice []string
	for key, value := range envMap {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, value))
	}
	return envSlice
}

// Helper function to create pointer to int64 for device cgroup specs
func toPtr(v int64) *int64 {
	return &v
}
