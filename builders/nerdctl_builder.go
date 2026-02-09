package builders

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// NerdctlBuilder builds container images using nerdctl CLI (Colima with containerd).
// This is a CLI-based fallback builder - prefer BuildKitBuilder for better performance.
//
// Works with: Colima (containerd mode via nerdctl CLI)
type NerdctlBuilder struct {
	profile    string
	namespace  string
	appPath    string
	imageName  string
	dockerfile string
}

// NewNerdctlBuilder creates a new nerdctl-based image builder.
// This builder shells out to `colima nerdctl` for build operations.
func NewNerdctlBuilder(profile, namespace, appPath, imageName, dockerfile string) *NerdctlBuilder {
	if profile == "" {
		profile = GetColimaProfile()
	}
	return &NerdctlBuilder{
		profile:    profile,
		namespace:  namespace,
		appPath:    appPath,
		imageName:  imageName,
		dockerfile: dockerfile,
	}
}

// Build builds the container image using nerdctl CLI.
func (b *NerdctlBuilder) Build(ctx context.Context, opts BuildOptions) error {
	// Generate nerdctl build command via colima
	args := []string{
		"nerdctl",
		"--profile", b.profile,
		"--",
		"--namespace", b.namespace,
		"build",
	}

	// Add dockerfile flag if specified
	if b.dockerfile != "" {
		args = append(args, "-f", b.dockerfile)
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

	// Add image tag
	args = append(args, "-t", b.imageName)

	// Add progress output
	args = append(args, "--progress", "plain")

	// Add build context last
	args = append(args, b.appPath)

	fmt.Printf("Building image: %s\n", b.imageName)
	fmt.Printf("Command: colima %s\n\n", strings.Join(args, " "))

	// Execute colima nerdctl build
	cmd := exec.CommandContext(ctx, "colima", args...)
	cmd.Dir = b.appPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	fmt.Printf("\n✓ Image built successfully: %s\n", b.imageName)
	return nil
}

// ImageExists checks if an image already exists using nerdctl CLI.
func (b *NerdctlBuilder) ImageExists(ctx context.Context) (bool, error) {
	args := []string{
		"nerdctl",
		"--profile", b.profile,
		"--",
		"--namespace", b.namespace,
		"images",
		"-q",
		b.imageName,
	}

	cmd := exec.CommandContext(ctx, "colima", args...)
	output, err := cmd.Output()
	if err != nil {
		return false, nil // If command fails, assume image doesn't exist
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// Close is a no-op for the CLI-based builder.
func (b *NerdctlBuilder) Close() error {
	return nil
}
