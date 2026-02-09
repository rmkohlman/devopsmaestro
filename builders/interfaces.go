// Package builders provides interfaces and implementations for building container images.
//
// Architecture:
//   - ImageBuilder interface defines the contract for all image builders
//   - BuilderConfig provides configuration for creating builders
//   - Factory function NewImageBuilder creates appropriate implementation based on platform
//
// Implementations:
//   - DockerBuilder: Uses Docker API (OrbStack, Docker Desktop, Podman)
//   - BuildKitBuilder: Uses BuildKit gRPC API (Colima with containerd)
//
// Example usage:
//
//	platform, _ := operators.NewPlatformDetector().Detect()
//	builder, _ := builders.NewImageBuilder(builders.BuilderConfig{
//	    Platform:   platform,
//	    Namespace:  "devopsmaestro",
//	    AppPath:    "/path/to/app",
//	    ImageName:  "myimage:latest",
//	    Dockerfile: "/path/to/Dockerfile",
//	})
//	defer builder.Close()
//	err := builder.Build(ctx, builders.BuildOptions{})
package builders

import (
	"context"
)

// ImageBuilder defines the interface for building container images.
// All implementations must be safe for concurrent use.
type ImageBuilder interface {
	// Build builds a container image from a Dockerfile.
	// Returns an error if the build fails.
	Build(ctx context.Context, opts BuildOptions) error

	// ImageExists checks if an image with the configured name already exists.
	// Returns (true, nil) if exists, (false, nil) if not, (false, err) on error.
	ImageExists(ctx context.Context) (bool, error)

	// Close releases any resources held by the builder (connections, etc).
	// Should be called when the builder is no longer needed.
	Close() error
}

// BuildOptions contains options for building an image.
type BuildOptions struct {
	// BuildArgs are build-time variables passed to the Dockerfile
	BuildArgs map[string]string

	// Target specifies the target stage for multi-stage builds
	Target string

	// NoCache disables the build cache when true
	NoCache bool

	// Pull forces pulling the base image even if cached
	Pull bool
}
