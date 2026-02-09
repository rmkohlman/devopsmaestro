package builders

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"devopsmaestro/operators"
)

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
func (b *DockerBuilder) Build(ctx context.Context, opts BuildOptions) error {
	fmt.Printf("Building image: %s\n", b.imageName)
	fmt.Printf("Using Docker CLI (%s)\n", b.platform.Name)
	fmt.Printf("Socket: %s\n\n", b.platform.SocketPath)

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

	fmt.Printf("Command: docker %s\n\n", strings.Join(args, " "))

	// Execute docker build
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Dir = b.appPath
	cmd.Env = append(os.Environ(), "DOCKER_HOST=unix://"+b.platform.SocketPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	fmt.Printf("\n✓ Image built successfully: %s\n", b.imageName)
	return nil
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
