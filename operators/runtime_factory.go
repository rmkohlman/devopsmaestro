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

// NewContainerRuntime creates the appropriate container runtime using a default
// PlatformDetector. Prefer NewContainerRuntimeWith for testability.
func NewContainerRuntime() (ContainerRuntime, error) {
	detector, err := NewPlatformDetector()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize platform detector: %w", err)
	}
	return NewContainerRuntimeWith(detector)
}

// NewContainerRuntimeWith creates the appropriate container runtime using the
// supplied PlatformDetector. This is the DI-friendly constructor — callers that
// need control over platform detection (tests, composition roots) should use
// this variant.
func NewContainerRuntimeWith(detector PlatformDetector) (ContainerRuntime, error) {
	config, err := resolveRuntimeConfig(detector)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve runtime configuration: %w", err)
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

// resolveRuntimeConfig determines which runtime and platform to use.
// The PlatformDetector is accepted as a parameter rather than created
// internally to support dependency injection.
func resolveRuntimeConfig(detector PlatformDetector) (*RuntimeConfig, error) {
	// Check config or environment variable for explicit runtime type
	runtimeType := viper.GetString("runtime.type")
	if runtimeType == "" {
		runtimeType = os.Getenv("DVM_RUNTIME")
	}

	platform, err := detector.Detect()
	if err != nil {
		return nil, fmt.Errorf("failed to detect container platform: %w", err)
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
		return "", fmt.Errorf("failed to get active runtime: %w", err)
	}
	return runtime.GetRuntimeType(), nil
}

// GetDetectedPlatformInfo returns user-friendly information about the detected
// platform. Uses a default PlatformDetector.
func GetDetectedPlatformInfo() (string, error) {
	detector, err := NewPlatformDetector()
	if err != nil {
		return "", fmt.Errorf("failed to initialize platform detector: %w", err)
	}
	return GetDetectedPlatformInfoWith(detector)
}

// GetDetectedPlatformInfoWith returns platform info using the supplied detector.
func GetDetectedPlatformInfoWith(detector PlatformDetector) (string, error) {
	platform, err := detector.Detect()
	if err != nil {
		return "", fmt.Errorf("failed to detect platform: %w", err)
	}
	return platform.Name, nil
}

// ListAvailablePlatforms returns all detected container platforms.
// Uses a default PlatformDetector.
func ListAvailablePlatforms() ([]*Platform, error) {
	detector, err := NewPlatformDetector()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize platform detector: %w", err)
	}
	return ListAvailablePlatformsWith(detector)
}

// ListAvailablePlatformsWith returns all platforms using the supplied detector.
func ListAvailablePlatformsWith(detector PlatformDetector) ([]*Platform, error) {
	return detector.DetectAll(), nil
}
