package builders

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

const dvmBuilderName = "dvm-builder"

// EnsureDVMBuilder ensures a buildx builder named "dvm-builder" exists with
// the given buildkitd.toml config. If the builder exists but has a different
// config, it is recreated. If no config path is given, returns empty string
// (use default builder).
//
// Returns the builder name to use with --builder, or empty string for default.
func EnsureDVMBuilder(configPath string, dockerHost string) string {
	if configPath == "" {
		return ""
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		slog.Warn("buildkit config file not found, using default builder", "path", configPath)
		return ""
	}

	// Check if builder already exists with correct config
	if builderHasConfig(configPath, dockerHost) {
		slog.Debug("dvm-builder already exists with correct config")
		return dvmBuilderName
	}

	// Remove existing builder if it exists (config mismatch)
	removeDVMBuilder(dockerHost)

	// Create new builder with config
	if err := createDVMBuilder(configPath, dockerHost); err != nil {
		slog.Warn("failed to create dvm-builder, using default builder", "error", err)
		return ""
	}

	slog.Info("created dvm-builder with registry mirror config", "config", configPath)
	return dvmBuilderName
}

// createDVMBuilder creates a new buildx builder with the given config.
func createDVMBuilder(configPath string, dockerHost string) error {
	args := []string{"buildx", "create", "--name", dvmBuilderName,
		"--driver", "docker-container", "--config", configPath}
	cmd := exec.Command("docker", args...)
	if dockerHost != "" {
		cmd.Env = append(os.Environ(), "DOCKER_HOST="+dockerHost)
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker buildx create failed: %w: %s", err, string(output))
	}
	return nil
}

// removeDVMBuilder removes the dvm-builder if it exists.
func removeDVMBuilder(dockerHost string) {
	cmd := exec.Command("docker", "buildx", "rm", dvmBuilderName)
	if dockerHost != "" {
		cmd.Env = append(os.Environ(), "DOCKER_HOST="+dockerHost)
	}
	_ = cmd.Run() // Ignore errors — builder may not exist
}

// builderHasConfig checks if the dvm-builder exists and its config matches.
// It compares the hash of the current config file with a stored hash marker.
func builderHasConfig(configPath string, dockerHost string) bool {
	// Check if builder exists
	cmd := exec.Command("docker", "buildx", "inspect", dvmBuilderName)
	if dockerHost != "" {
		cmd.Env = append(os.Environ(), "DOCKER_HOST="+dockerHost)
	}
	if err := cmd.Run(); err != nil {
		return false // Builder doesn't exist
	}

	// Compare config hash with marker file
	configData, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}
	currentHash := fmt.Sprintf("%x", sha256.Sum256(configData))

	markerPath := configPath + ".hash"
	storedHash, err := os.ReadFile(markerPath)
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(storedHash)) == currentHash
}

// WriteConfigHash writes a hash marker file alongside the config for change detection.
func WriteConfigHash(configPath string) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(data))
	_ = os.WriteFile(configPath+".hash", []byte(hash), 0600)
}
