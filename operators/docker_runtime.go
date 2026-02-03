package operators

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/moby/go-archive"
	"github.com/moby/term"
)

// DockerRuntime implements ContainerRuntime for Docker-compatible platforms
// Supports: OrbStack, Colima, Docker Desktop, Podman, native Linux Docker
type DockerRuntime struct {
	client   *client.Client
	platform *Platform
}

// NewDockerRuntime creates a new Docker runtime instance for the given platform
func NewDockerRuntime(platform *Platform) (*DockerRuntime, error) {
	if platform == nil {
		return nil, fmt.Errorf("platform cannot be nil")
	}

	// Set Docker host based on platform socket
	dockerHost := fmt.Sprintf("unix://%s", platform.SocketPath)
	os.Setenv("DOCKER_HOST", dockerHost)

	// Create Docker client
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	// Verify connection
	if _, err := cli.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w\n%s",
			platform.Name, err, platform.GetStartHint())
	}

	return &DockerRuntime{
		client:   cli,
		platform: platform,
	}, nil
}

// BuildImage builds a container image using Docker
func (d *DockerRuntime) BuildImage(ctx context.Context, opts BuildOptions) error {
	fmt.Printf("Building image '%s' using %s...\n", opts.ImageName, d.platform.Name)

	// Create build context tarball
	buildCtx, err := archive.TarWithOptions(opts.BuildContext, &archive.TarOptions{})
	if err != nil {
		return fmt.Errorf("failed to create build context: %w", err)
	}
	defer buildCtx.Close()

	// Prepare build options
	buildArgs := make(map[string]*string)
	for k, v := range opts.BuildArgs {
		val := v // Create a copy
		buildArgs[k] = &val
	}

	buildOpts := types.ImageBuildOptions{
		Tags:       append([]string{opts.ImageName}, opts.Tags...),
		Dockerfile: filepath.Base(opts.Dockerfile),
		Remove:     true,
		BuildArgs:  buildArgs,
	}

	// Start build
	buildResp, err := d.client.ImageBuild(ctx, buildCtx, buildOpts)
	if err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}
	defer buildResp.Body.Close()

	// Stream build output to terminal
	termFd, isTerm := term.GetFdInfo(os.Stdout)
	if err := jsonmessage.DisplayJSONMessagesStream(buildResp.Body, os.Stdout, termFd, isTerm, nil); err != nil {
		return fmt.Errorf("error during build: %w", err)
	}

	fmt.Printf("✓ Image '%s' built successfully\n", opts.ImageName)
	return nil
}

// StartWorkspace starts a Docker container as a workspace
// It handles three cases:
// 1. Container exists and is running -> return its ID
// 2. Container exists but is stopped -> start it and return its ID
// 3. Container doesn't exist -> create and start it
func (d *DockerRuntime) StartWorkspace(ctx context.Context, opts StartOptions) (string, error) {
	// Determine container name
	containerName := opts.ContainerName
	if containerName == "" {
		containerName = opts.WorkspaceName
	}

	// Check if container already exists
	existingContainers, err := d.client.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("name", fmt.Sprintf("^%s$", containerName)),
		),
	})
	if err != nil {
		return "", fmt.Errorf("failed to check for existing container: %w", err)
	}

	if len(existingContainers) > 0 {
		existing := existingContainers[0]

		// Check if it's running
		if existing.State == "running" {
			return existing.ID[:12], nil
		}

		// Container exists but is stopped - start it
		if err := d.client.ContainerStart(ctx, existing.ID, container.StartOptions{}); err != nil {
			return "", fmt.Errorf("failed to start existing container: %w", err)
		}

		return existing.ID[:12], nil
	}

	// Container doesn't exist - create it
	// Set default command to keep container running
	command := opts.Command
	if len(command) == 0 {
		command = []string{"/bin/sleep", "infinity"}
	}

	// Set default working directory
	workingDir := opts.WorkingDir
	if workingDir == "" {
		workingDir = "/workspace"
	}

	// Build environment variables
	env := envMapToSlice(opts.Env)
	if opts.ProjectName != "" {
		env = append(env, fmt.Sprintf("DVM_PROJECT=%s", opts.ProjectName))
	}
	if opts.WorkspaceName != "" {
		env = append(env, fmt.Sprintf("DVM_WORKSPACE=%s", opts.WorkspaceName))
	}

	// Create container configuration with proper DVM labels
	containerConfig := &container.Config{
		Image:      opts.ImageName,
		Cmd:        command,
		WorkingDir: workingDir,
		Tty:        true,
		OpenStdin:  true,
		Env:        env,
		Labels: map[string]string{
			"io.devopsmaestro.managed":   "true",
			"io.devopsmaestro.namespace": "devopsmaestro",
			"io.devopsmaestro.project":   opts.ProjectName,
			"io.devopsmaestro.workspace": opts.WorkspaceName,
		},
	}

	// Build volume mounts
	binds := []string{
		fmt.Sprintf("%s:/workspace", opts.ProjectPath),
	}

	// Mount SSH keys if they exist
	homeDir, _ := os.UserHomeDir()
	sshDir := filepath.Join(homeDir, ".ssh")
	if _, err := os.Stat(sshDir); err == nil {
		binds = append(binds, fmt.Sprintf("%s:/home/dev/.ssh:ro", sshDir))
	}

	// Create host configuration
	hostConfig := &container.HostConfig{
		Binds: binds,
	}

	// Create container
	resp, err := d.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil,
		nil,
		containerName,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	return resp.ID[:12], nil
}

// AttachToWorkspace attaches an interactive terminal to a running workspace
func (d *DockerRuntime) AttachToWorkspace(ctx context.Context, workspaceID string) error {
	fmt.Printf("Attaching to workspace (press Ctrl+D to exit)...\n")

	// Execute zsh in the container with interactive TTY
	execConfig := container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{"/bin/zsh"},
	}

	execResp, err := d.client.ContainerExecCreate(ctx, workspaceID, execConfig)
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	// Attach to the exec
	attachResp, err := d.client.ContainerExecAttach(ctx, execResp.ID, container.ExecStartOptions{
		Tty: true,
	})
	if err != nil {
		return fmt.Errorf("failed to attach: %w", err)
	}
	defer attachResp.Close()

	// Put terminal in raw mode
	oldState, err := term.SetRawTerminal(os.Stdin.Fd())
	if err != nil {
		return fmt.Errorf("failed to set raw terminal: %w", err)
	}
	defer term.RestoreTerminal(os.Stdin.Fd(), oldState)

	// Stream I/O
	errChan := make(chan error, 2)

	// Copy from container to stdout/stderr
	go func() {
		_, err := io.Copy(os.Stdout, attachResp.Reader)
		errChan <- err
	}()

	// Copy from stdin to container
	go func() {
		_, err := io.Copy(attachResp.Conn, os.Stdin)
		errChan <- err
	}()

	// Wait for I/O to finish
	<-errChan

	fmt.Printf("\n✓ Detached from workspace\n")
	return nil
}

// StopWorkspace stops a running workspace
func (d *DockerRuntime) StopWorkspace(ctx context.Context, workspaceID string) error {
	fmt.Printf("Stopping workspace...\n")

	timeout := 10
	if err := d.client.ContainerStop(ctx, workspaceID, container.StopOptions{Timeout: &timeout}); err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}

	fmt.Printf("✓ Workspace stopped\n")
	return nil
}

// GetWorkspaceStatus returns the status of a workspace
func (d *DockerRuntime) GetWorkspaceStatus(ctx context.Context, workspaceID string) (string, error) {
	containerJSON, err := d.client.ContainerInspect(ctx, workspaceID)
	if err != nil {
		return "unknown", fmt.Errorf("failed to inspect container: %w", err)
	}

	if containerJSON.State.Running {
		return "running", nil
	} else if containerJSON.State.Paused {
		return "paused", nil
	} else if containerJSON.State.Restarting {
		return "restarting", nil
	} else if containerJSON.State.Dead {
		return "dead", nil
	}

	return "stopped", nil
}

// GetRuntimeType returns "docker"
func (d *DockerRuntime) GetRuntimeType() string {
	return "docker"
}

// GetPlatformName returns the human-readable platform name
func (d *DockerRuntime) GetPlatformName() string {
	return d.platform.Name
}

// GetPlatform returns the platform this runtime is using
func (d *DockerRuntime) GetPlatform() *Platform {
	return d.platform
}

// ListWorkspaces lists all DVM-managed workspaces
func (d *DockerRuntime) ListWorkspaces(ctx context.Context) ([]WorkspaceInfo, error) {
	// List containers with DVM label
	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		All: true, // Include stopped containers
		Filters: filters.NewArgs(
			filters.Arg("label", "io.devopsmaestro.managed=true"),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var workspaces []WorkspaceInfo
	for _, c := range containers {
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
			// Remove leading slash from Docker container names
			if len(name) > 0 && name[0] == '/' {
				name = name[1:]
			}
		}

		workspaces = append(workspaces, WorkspaceInfo{
			ID:        c.ID[:12],
			Name:      name,
			Status:    c.Status,
			Image:     c.Image,
			Project:   c.Labels["io.devopsmaestro.project"],
			Workspace: c.Labels["io.devopsmaestro.workspace"],
			Labels:    c.Labels,
		})
	}

	return workspaces, nil
}

// FindWorkspace finds a workspace by name and returns its info
func (d *DockerRuntime) FindWorkspace(ctx context.Context, name string) (*WorkspaceInfo, error) {
	// List containers with DVM label and name filter
	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "io.devopsmaestro.managed=true"),
			filters.Arg("name", name),
		),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find container: %w", err)
	}

	if len(containers) == 0 {
		return nil, nil // Not found
	}

	c := containers[0]
	containerName := ""
	if len(c.Names) > 0 {
		containerName = c.Names[0]
		if len(containerName) > 0 && containerName[0] == '/' {
			containerName = containerName[1:]
		}
	}

	return &WorkspaceInfo{
		ID:        c.ID[:12],
		Name:      containerName,
		Status:    c.Status,
		Image:     c.Image,
		Project:   c.Labels["io.devopsmaestro.project"],
		Workspace: c.Labels["io.devopsmaestro.workspace"],
		Labels:    c.Labels,
	}, nil
}

// StopAllWorkspaces stops all DVM-managed workspaces
func (d *DockerRuntime) StopAllWorkspaces(ctx context.Context) (int, error) {
	// List running DVM containers
	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		Filters: filters.NewArgs(
			filters.Arg("label", "io.devopsmaestro.managed=true"),
			filters.Arg("status", "running"),
		),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list containers: %w", err)
	}

	stopped := 0
	timeout := 10
	for _, c := range containers {
		if err := d.client.ContainerStop(ctx, c.ID, container.StopOptions{Timeout: &timeout}); err != nil {
			// Log error but continue stopping others
			continue
		}
		stopped++
	}

	return stopped, nil
}

// Helper function to convert map to env slice
func envMapToSlice(envMap map[string]string) []string {
	var envSlice []string
	for key, value := range envMap {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, value))
	}
	return envSlice
}
