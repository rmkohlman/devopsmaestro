package operators

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
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
func (d *DockerRuntime) StartWorkspace(ctx context.Context, opts StartOptions) (string, error) {
	fmt.Printf("Starting workspace '%s' using %s...\n", opts.WorkspaceName, d.platform.Name)

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

	// Create container configuration
	containerConfig := &container.Config{
		Image:      opts.ImageName,
		Cmd:        command,
		WorkingDir: workingDir,
		Tty:        true,
		OpenStdin:  true,
		Env:        envMapToSlice(opts.Env),
		Labels: map[string]string{
			"devopsmaestro.workspace": opts.WorkspaceName,
			"devopsmaestro.managed":   "true",
		},
	}

	// Create host configuration (volume mounts, etc.)
	hostConfig := &container.HostConfig{
		Binds: []string{
			fmt.Sprintf("%s:/workspace", opts.ProjectPath),
		},
		AutoRemove: true, // Ephemeral containers
	}

	// Create container
	resp, err := d.client.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil,
		nil,
		opts.WorkspaceName,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	// Start container
	if err := d.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", fmt.Errorf("failed to start container: %w", err)
	}

	fmt.Printf("✓ Workspace started (Container ID: %s)\n", resp.ID[:12])
	return resp.ID, nil
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

// GetPlatform returns the platform this runtime is using
func (d *DockerRuntime) GetPlatform() *Platform {
	return d.platform
}

// Helper function to convert map to env slice
func envMapToSlice(envMap map[string]string) []string {
	var envSlice []string
	for key, value := range envMap {
		envSlice = append(envSlice, fmt.Sprintf("%s=%s", key, value))
	}
	return envSlice
}
