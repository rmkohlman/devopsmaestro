package builders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ImageBuilder builds container images using nerdctl (containerd)
type ImageBuilder struct {
	profile     string
	namespace   string
	projectPath string
	imageName   string
	dockerfile  string
}

// NewImageBuilder creates a new image builder
func NewImageBuilder(profile, namespace, projectPath, imageName, dockerfile string) *ImageBuilder {
	return &ImageBuilder{
		profile:     profile,
		namespace:   namespace,
		projectPath: projectPath,
		imageName:   imageName,
		dockerfile:  dockerfile,
	}
}

// Build builds the container image using nerdctl
func (b *ImageBuilder) Build(buildArgs map[string]string, target string) error {
	// Generate nerdctl build command
	// Use colima nerdctl which handles the socket automatically
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
	for key, value := range buildArgs {
		args = append(args, "--build-arg", fmt.Sprintf("%s=%s", key, value))
	}

	// Add target if specified
	if target != "" {
		args = append(args, "--target", target)
	}

	// Add image tag
	args = append(args, "-t", b.imageName)

	// Add progress output
	args = append(args, "--progress", "plain")

	// Add build context last
	args = append(args, b.projectPath)

	fmt.Printf("Building image: %s\n", b.imageName)
	fmt.Printf("Command: colima %s\n\n", strings.Join(args, " "))

	// Execute colima nerdctl build
	cmd := exec.Command("colima", args...)
	cmd.Dir = b.projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	fmt.Printf("\n✓ Image built successfully: %s\n", b.imageName)
	return nil
}

// ImageExists checks if an image already exists
func (b *ImageBuilder) ImageExists() (bool, error) {
	args := []string{
		"nerdctl",
		"--profile", b.profile,
		"--",
		"--namespace", b.namespace,
		"images",
		"-q",
		b.imageName,
	}

	cmd := exec.Command("colima", args...)
	output, err := cmd.Output()
	if err != nil {
		return false, nil // If command fails, assume image doesn't exist
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// SaveDockerfile saves the generated Dockerfile to disk
func SaveDockerfile(content, projectPath string) (string, error) {
	dockerfilePath := filepath.Join(projectPath, "Dockerfile.dvm")

	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	fmt.Printf("✓ Generated Dockerfile: %s\n", dockerfilePath)
	return dockerfilePath, nil
}

// GetColimaProfile returns the active Colima profile
func GetColimaProfile() string {
	profile := os.Getenv("COLIMA_ACTIVE_PROFILE")
	if profile == "" {
		// Try to detect running profile
		cmd := exec.Command("colima", "list", "--json")
		_, err := cmd.Output()
		if err == nil {
			// Parse JSON to find running profile
			// For now, use default
			profile = "local-med"
		} else {
			profile = "default"
		}
	}
	return profile
}
