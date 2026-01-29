package builders

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// SaveDockerfile saves the generated Dockerfile content to disk.
// Returns the path to the saved Dockerfile.
func SaveDockerfile(content, projectPath string) (string, error) {
	dockerfilePath := filepath.Join(projectPath, "Dockerfile.dvm")

	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	fmt.Printf("✓ Generated Dockerfile: %s\n", dockerfilePath)
	return dockerfilePath, nil
}

// GetColimaProfile returns the active Colima profile from environment or default.
func GetColimaProfile() string {
	if profile := os.Getenv("COLIMA_ACTIVE_PROFILE"); profile != "" {
		return profile
	}
	if profile := os.Getenv("COLIMA_DOCKER_PROFILE"); profile != "" {
		return profile
	}
	return "default"
}

// IsColimaRunning checks if Colima is running with the specified profile.
func IsColimaRunning(profile string) bool {
	cmd := exec.Command("colima", "status", "--profile", profile)
	return cmd.Run() == nil
}
