package operators

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// detectRuntime auto-detects which runtime is available by checking Colima sockets
func detectRuntime() string {
	// Get Colima profile from environment
	profile := os.Getenv("COLIMA_DOCKER_PROFILE")
	if profile == "" {
		profile = os.Getenv("COLIMA_ACTIVE_PROFILE")
	}
	if profile == "" {
		profile = "default"
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "docker" // Fallback
	}

	// Check for docker.sock first (backward compatibility with Docker runtime)
	dockerSock := filepath.Join(homeDir, ".colima", profile, "docker.sock")
	if _, err := os.Stat(dockerSock); err == nil {
		return "docker"
	}

	// Check for containerd.sock
	containerdSock := filepath.Join(homeDir, ".colima", profile, "containerd.sock")
	if _, err := os.Stat(containerdSock); err == nil {
		return "containerd"
	}

	// Default to docker for backward compatibility
	return "docker"
}

// NewContainerRuntime creates the appropriate container runtime based on configuration
// It auto-detects the runtime if not specified in config
func NewContainerRuntime() (ContainerRuntime, error) {
	// Check config or environment variable
	runtimeType := viper.GetString("runtime.type")
	if runtimeType == "" {
		runtimeType = os.Getenv("DVM_RUNTIME")
	}

	// Auto-detect if not specified or explicitly set to "auto"
	if runtimeType == "" || runtimeType == "auto" {
		runtimeType = detectRuntime()
	}

	switch runtimeType {
	case "docker":
		return NewDockerRuntime()
	case "containerd":
		// Use V2 implementation
		return NewContainerdRuntimeV2()
	case "kubernetes", "k8s":
		// For Phase 3
		return nil, fmt.Errorf("kubernetes runtime not yet implemented (coming in Phase 3)")
	default:
		return nil, fmt.Errorf("unknown runtime type: %s (supported: docker, containerd, auto)", runtimeType)
	}
}

// GetActiveRuntime returns information about the active runtime
func GetActiveRuntime() (string, error) {
	runtime, err := NewContainerRuntime()
	if err != nil {
		return "", err
	}
	return runtime.GetRuntimeType(), nil
}
