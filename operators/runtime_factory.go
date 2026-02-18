package operators

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

// RuntimeType identifies the container runtime API to use
type RuntimeType string

const (
	RuntimeDocker     RuntimeType = "docker"
	RuntimeContainerd RuntimeType = "containerd"
	RuntimeKubernetes RuntimeType = "kubernetes"
)

// RuntimeConfig holds configuration for creating a runtime
type RuntimeConfig struct {
	Platform *Platform
	Type     RuntimeType
}

// NewContainerRuntime creates the appropriate container runtime based on configuration
// It auto-detects the platform if not specified in config
func NewContainerRuntime() (ContainerRuntime, error) {
	config, err := resolveRuntimeConfig()
	if err != nil {
		return nil, err
	}

	switch config.Type {
	case RuntimeDocker:
		return NewDockerRuntime(config.Platform)
	case RuntimeContainerd:
		return NewContainerdRuntimeV2WithPlatform(config.Platform)
	case RuntimeKubernetes:
		return nil, fmt.Errorf("kubernetes runtime not yet implemented (coming in Phase 3)")
	default:
		return nil, fmt.Errorf("unknown runtime type: %s (supported: docker, containerd)", config.Type)
	}
}

// resolveRuntimeConfig determines which runtime and platform to use
func resolveRuntimeConfig() (*RuntimeConfig, error) {
	// Check config or environment variable for explicit runtime type
	runtimeType := viper.GetString("runtime.type")
	if runtimeType == "" {
		runtimeType = os.Getenv("DVM_RUNTIME")
	}

	// Detect platform
	detector, err := NewPlatformDetector()
	if err != nil {
		return nil, err
	}

	platform, err := detector.Detect()
	if err != nil {
		return nil, err
	}

	// Determine runtime type
	var rt RuntimeType
	if runtimeType == "" || runtimeType == "auto" {
		// Auto-detect based on platform
		if platform.IsContainerd() {
			rt = RuntimeContainerd
		} else {
			rt = RuntimeDocker
		}
	} else {
		switch runtimeType {
		case "docker":
			rt = RuntimeDocker
		case "containerd":
			rt = RuntimeContainerd
		case "kubernetes", "k8s":
			rt = RuntimeKubernetes
		default:
			return nil, fmt.Errorf("unknown runtime type: %s (supported: docker, containerd, auto)", runtimeType)
		}
	}

	return &RuntimeConfig{
		Platform: platform,
		Type:     rt,
	}, nil
}

// GetActiveRuntime returns information about the active runtime
func GetActiveRuntime() (string, error) {
	runtime, err := NewContainerRuntime()
	if err != nil {
		return "", err
	}
	return runtime.GetRuntimeType(), nil
}

// GetDetectedPlatformInfo returns user-friendly information about the detected platform
func GetDetectedPlatformInfo() (string, error) {
	detector, err := NewPlatformDetector()
	if err != nil {
		return "", err
	}

	platform, err := detector.Detect()
	if err != nil {
		return "", err
	}

	return platform.Name, nil
}

// ListAvailablePlatforms returns all detected container platforms
func ListAvailablePlatforms() ([]*Platform, error) {
	detector, err := NewPlatformDetector()
	if err != nil {
		return nil, err
	}

	return detector.DetectAll(), nil
}
