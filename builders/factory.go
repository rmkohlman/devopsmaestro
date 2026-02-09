package builders

import (
	"fmt"

	"devopsmaestro/operators"
)

// BuilderConfig contains configuration for creating an ImageBuilder.
// This decouples the builder creation from specific platform details.
type BuilderConfig struct {
	// Platform is the container platform to use (required)
	Platform *operators.Platform

	// Namespace is the container namespace (e.g., "devopsmaestro")
	Namespace string

	// AppPath is the path to the app directory (required)
	AppPath string

	// ImageName is the name/tag for the built image (required)
	ImageName string

	// Dockerfile is the path to the Dockerfile (optional, defaults to AppPath/Dockerfile)
	Dockerfile string
}

// Validate checks that required fields are set
func (c BuilderConfig) Validate() error {
	if c.Platform == nil {
		return fmt.Errorf("platform is required")
	}
	if c.AppPath == "" {
		return fmt.Errorf("app path is required")
	}
	if c.ImageName == "" {
		return fmt.Errorf("image name is required")
	}
	return nil
}

// NewImageBuilder creates an appropriate ImageBuilder based on the platform.
// This is the factory function that decouples consumers from specific implementations.
//
// Platform selection:
//   - Docker-compatible (OrbStack, Docker Desktop, Podman): DockerBuilder
//   - Containerd (Colima with containerd): BuildKitBuilder
func NewImageBuilder(cfg BuilderConfig) (ImageBuilder, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	// Choose builder based on platform capabilities
	if cfg.Platform.IsDockerCompatible() {
		return NewDockerBuilder(cfg)
	}

	if cfg.Platform.IsContainerd() {
		return NewBuildKitBuilder(cfg)
	}

	return nil, fmt.Errorf("unsupported platform type: %s", cfg.Platform.Type)
}
