package builders

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"devopsmaestro/operators"
	"devopsmaestro/render"
)

// WatchdogRunner is a function type for running commands with a watchdog.
// This allows injection of mock watchdogs for testing.
type WatchdogRunner func(ctx context.Context, cmd *exec.Cmd, checkCondition func(ctx context.Context) bool, cfg WatchdogConfig) (WatchdogResult, error)

// DockerBuilder builds container images using the Docker CLI.
// Works with: OrbStack, Docker Desktop, Podman, Colima (docker mode)
//
// This implementation uses the docker CLI rather than the SDK to avoid
// version compatibility issues and provide consistent behavior across platforms.
type DockerBuilder struct {
	platform   *operators.Platform
	namespace  string
	appPath    string
	imageName  string
	dockerfile string

	// WatchdogConfig allows customizing the build watchdog behavior.
	// If zero, DefaultWatchdogConfig() is used.
	WatchdogConfig WatchdogConfig

	// WatchdogRunner allows injecting a custom watchdog for testing.
	// If nil, RunWithWatchdog is used.
	WatchdogRunner WatchdogRunner
}

// NewDockerBuilder creates a new Docker CLI-based image builder.
func NewDockerBuilder(cfg BuilderConfig) (*DockerBuilder, error) {
	// Verify we can connect to Docker
	dockerHost := "unix://" + cfg.Platform.SocketPath

	cmd := exec.Command("docker", "info")
	cmd.Env = append(os.Environ(), "DOCKER_HOST="+dockerHost)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to connect to Docker at %s: %w\n%s",
			cfg.Platform.SocketPath, err, cfg.Platform.GetStartHint())
	}

	return &DockerBuilder{
		platform:   cfg.Platform,
		namespace:  cfg.Namespace,
		appPath:    cfg.AppPath,
		imageName:  cfg.ImageName,
		dockerfile: cfg.Dockerfile,
	}, nil
}

// Build builds the container image using Docker CLI.
// Includes a watchdog to handle Docker buildx hang issue on Colima.
func (b *DockerBuilder) Build(ctx context.Context, opts BuildOptions) error {
	render.Progressf("Building image: %s", b.imageName)
	render.Infof("Using Docker CLI (%s)", b.platform.Name)
	render.Infof("Socket: %s", b.platform.SocketPath)
	render.Blank()

	// Build docker build command
	args := []string{"build"}

	// Add dockerfile flag if specified
	dockerfilePath := b.dockerfile
	if dockerfilePath != "" {
		// Make path relative to app path if absolute
		if filepath.IsAbs(dockerfilePath) {
			rel, err := filepath.Rel(b.appPath, dockerfilePath)
			if err == nil {
				dockerfilePath = rel
			}
		}
		args = append(args, "-f", dockerfilePath)
	}

	// Add build args
	for key, value := range opts.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	// Add target if specified
	if opts.Target != "" {
		args = append(args, "--target", opts.Target)
	}

	// Add no-cache if specified
	if opts.NoCache {
		args = append(args, "--no-cache")
	}

	// Add pull if specified
	if opts.Pull {
		args = append(args, "--pull")
	}

	// Add labels for namespace tracking
	args = append(args, "--label", "io.devopsmaestro.namespace="+b.namespace)
	args = append(args, "--label", "io.devopsmaestro.managed=true")

	// Add image tag
	args = append(args, "-t", b.imageName)

	// Add progress output
	args = append(args, "--progress", "plain")

	// Add --load flag for Podman (uses buildkit docker-container driver by default)
	// which doesn't auto-load images into local storage
	if b.platform.Type == operators.PlatformPodman {
		args = append(args, "--load")
	}

	// Add build context (project path) last
	args = append(args, ".")

	render.Infof("Command: docker %s", strings.Join(redactBuildArgs(args), " "))
	render.Blank()

	// Prepare docker build command
	cmd := exec.Command("docker", args...)
	cmd.Dir = b.appPath
	cmd.Env = append(os.Environ(), "DOCKER_HOST=unix://"+b.platform.SocketPath)
	stdoutWriter := NewRedactingWriter(os.Stdout, opts.BuildArgs)
	stderrWriter := NewRedactingWriter(os.Stderr, opts.BuildArgs)
	cmd.Stdout = stdoutWriter
	cmd.Stderr = stderrWriter

	// Flush any buffered bytes when build completes
	if rw, ok := stdoutWriter.(*RedactingWriter); ok {
		defer rw.Flush()
	}
	if rw, ok := stderrWriter.(*RedactingWriter); ok {
		defer rw.Flush()
	}

	// Use watchdog config (defaults if not set)
	cfg := b.WatchdogConfig
	if cfg.PollInterval == 0 && cfg.Timeout == 0 {
		cfg = DefaultWatchdogConfig()
	}

	// Use injected watchdog runner or default
	runner := b.WatchdogRunner
	if runner == nil {
		runner = RunWithWatchdog
	}

	// Run build with watchdog to handle Docker buildx + Colima hang
	// where the build completes but the process doesn't exit
	result, err := runner(ctx, cmd, func(checkCtx context.Context) bool {
		exists, _ := b.ImageExists(checkCtx)
		return exists
	}, cfg)

	switch result {
	case WatchdogCompleted:
		if err != nil {
			return fmt.Errorf("failed to build image: %w", err)
		}
		render.Blank()
		render.Successf("Image built successfully: %s", b.imageName)
		return nil

	case WatchdogDetected:
		render.Blank()
		render.Info("[watchdog] Image detected, terminating hung build process...")
		render.Successf("Image built successfully: %s", b.imageName)
		return nil

	case WatchdogTimedOut:
		return fmt.Errorf("build timed out after %v", cfg.Timeout)

	case WatchdogCancelled:
		return err

	default:
		return fmt.Errorf("unexpected watchdog result: %v", result)
	}
}

// ImageExists checks if an image already exists using docker CLI.
func (b *DockerBuilder) ImageExists(ctx context.Context) (bool, error) {
	cmd := exec.CommandContext(ctx, "docker", "images", "-q", b.imageName)
	cmd.Env = append(os.Environ(), "DOCKER_HOST=unix://"+b.platform.SocketPath)

	output, err := cmd.Output()
	if err != nil {
		return false, nil // If command fails, assume image doesn't exist
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// Close is a no-op for the CLI-based builder.
func (b *DockerBuilder) Close() error {
	return nil
}

// redactBuildArgs returns a copy of args with --build-arg values redacted.
// Input:  ["build", "--build-arg", "KEY=secret", "-t", "img"]
// Output: ["build", "--build-arg", "KEY=***", "-t", "img"]
func redactBuildArgs(args []string) []string {
	redacted := make([]string, len(args))
	copy(redacted, args)
	for i, arg := range redacted {
		if i > 0 && redacted[i-1] == "--build-arg" {
			// This arg is a build-arg value (KEY=VALUE format)
			if eqIdx := strings.Index(arg, "="); eqIdx >= 0 {
				redacted[i] = arg[:eqIdx+1] + "***"
			}
		}
	}
	return redacted
}

// buildDockerArgsForLog returns a version of the docker build args suitable for
// logging, with --build-arg values replaced by "***" to prevent secret exposure.
func buildDockerArgsForLog(opts BuildOptions) []string {
	var args []string
	for key := range opts.BuildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=***", key))
	}
	return args
}
