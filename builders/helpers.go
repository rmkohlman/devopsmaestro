package builders

import (
	"fmt"
	"os"
	"path/filepath"
)

// SaveDockerfile saves the generated Dockerfile content to disk.
// Returns the path to the saved Dockerfile.
func SaveDockerfile(content, appPath string) (string, error) {
	dockerfilePath := filepath.Join(appPath, "Dockerfile.dvm")

	if err := os.WriteFile(dockerfilePath, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	fmt.Printf("✓ Generated Dockerfile: %s\n", dockerfilePath)
	return dockerfilePath, nil
}
