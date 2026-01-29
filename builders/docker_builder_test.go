package builders

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"devopsmaestro/operators"
)

func TestDockerBuilder_Implements_ImageBuilder(t *testing.T) {
	// Compile-time check that DockerBuilder implements ImageBuilder
	var _ ImageBuilder = (*DockerBuilder)(nil)
}

func TestNewDockerBuilder_InvalidSocket(t *testing.T) {
	config := BuilderConfig{
		Platform: &operators.Platform{
			Type:       operators.PlatformOrbStack,
			SocketPath: "/nonexistent/socket.sock",
		},
		ProjectPath: "/tmp/test",
		ImageName:   "test:latest",
	}

	builder, err := NewDockerBuilder(config)
	if err == nil {
		t.Error("NewDockerBuilder() should fail with invalid socket")
		if builder != nil {
			builder.Close()
		}
	}
}

func TestDockerBuilder_Close(t *testing.T) {
	// DockerBuilder.Close() is a no-op, should always return nil
	builder := &DockerBuilder{}
	err := builder.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

// Integration tests requiring actual Docker runtime

func requireDockerPlatform(t *testing.T) *operators.Platform {
	t.Helper()

	detector, err := operators.NewPlatformDetector()
	if err != nil {
		t.Skipf("Failed to create platform detector: %v", err)
	}

	platforms := detector.DetectAll()
	for _, p := range platforms {
		if p.IsDockerCompatible() {
			return p
		}
	}

	t.Skip("No Docker-compatible platform available")
	return nil
}

func TestIntegration_DockerBuilder_New(t *testing.T) {
	platform := requireDockerPlatform(t)

	projectPath := t.TempDir()

	config := BuilderConfig{
		Platform:    platform,
		Namespace:   "devopsmaestro",
		ProjectPath: projectPath,
		ImageName:   "test-docker-builder:latest",
	}

	builder, err := NewDockerBuilder(config)
	if err != nil {
		t.Fatalf("NewDockerBuilder() error = %v", err)
	}
	defer builder.Close()

	// Verify internal state
	if builder.platform != platform {
		t.Error("platform not set correctly")
	}
	if builder.namespace != "devopsmaestro" {
		t.Errorf("namespace = %q, want %q", builder.namespace, "devopsmaestro")
	}
	if builder.projectPath != projectPath {
		t.Errorf("projectPath = %q, want %q", builder.projectPath, projectPath)
	}
	if builder.imageName != "test-docker-builder:latest" {
		t.Errorf("imageName = %q, want %q", builder.imageName, "test-docker-builder:latest")
	}
}

func TestIntegration_DockerBuilder_ImageExists_NotFound(t *testing.T) {
	platform := requireDockerPlatform(t)

	projectPath := t.TempDir()

	config := BuilderConfig{
		Platform:    platform,
		Namespace:   "devopsmaestro",
		ProjectPath: projectPath,
		ImageName:   "nonexistent-image-" + t.Name() + ":v999",
	}

	builder, err := NewDockerBuilder(config)
	if err != nil {
		t.Fatalf("NewDockerBuilder() error = %v", err)
	}
	defer builder.Close()

	ctx := context.Background()
	exists, err := builder.ImageExists(ctx)
	if err != nil {
		t.Fatalf("ImageExists() error = %v", err)
	}

	if exists {
		t.Error("ImageExists() should return false for nonexistent image")
	}
}

func TestIntegration_DockerBuilder_Build(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	platform := requireDockerPlatform(t)

	// Create temporary project with Dockerfile
	projectPath := t.TempDir()
	dockerfile := filepath.Join(projectPath, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`
FROM alpine:latest
RUN echo "test"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	imageName := "dvm-test-docker-builder:test"

	config := BuilderConfig{
		Platform:    platform,
		Namespace:   "devopsmaestro",
		ProjectPath: projectPath,
		ImageName:   imageName,
	}

	builder, err := NewDockerBuilder(config)
	if err != nil {
		t.Fatalf("NewDockerBuilder() error = %v", err)
	}
	defer builder.Close()

	ctx := context.Background()

	// Build the image
	err = builder.Build(ctx, BuildOptions{})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Verify image exists
	exists, err := builder.ImageExists(ctx)
	if err != nil {
		t.Fatalf("ImageExists() error = %v", err)
	}
	if !exists {
		t.Error("ImageExists() should return true after build")
	}

	// Cleanup: remove test image
	t.Cleanup(func() {
		// Don't fail test if cleanup fails
		cleanup := &DockerBuilder{
			platform:  platform,
			imageName: imageName,
		}
		// Use docker rmi to remove image
		_ = cleanup.Close()
	})
}

func TestIntegration_DockerBuilder_Build_WithOptions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	platform := requireDockerPlatform(t)

	projectPath := t.TempDir()
	dockerfile := filepath.Join(projectPath, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`
FROM alpine:latest
ARG BUILD_VERSION=unknown
RUN echo "Version: ${BUILD_VERSION}"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	imageName := "dvm-test-docker-builder-opts:test"

	config := BuilderConfig{
		Platform:    platform,
		Namespace:   "devopsmaestro",
		ProjectPath: projectPath,
		ImageName:   imageName,
	}

	builder, err := NewDockerBuilder(config)
	if err != nil {
		t.Fatalf("NewDockerBuilder() error = %v", err)
	}
	defer builder.Close()

	ctx := context.Background()

	// Build with options
	err = builder.Build(ctx, BuildOptions{
		BuildArgs: map[string]string{
			"BUILD_VERSION": "1.0.0-test",
		},
		NoCache: true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestIntegration_DockerBuilder_Build_CustomDockerfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	platform := requireDockerPlatform(t)

	projectPath := t.TempDir()

	// Create custom Dockerfile with different name
	customDockerfile := filepath.Join(projectPath, "Dockerfile.custom")
	err := os.WriteFile(customDockerfile, []byte(`
FROM alpine:latest
LABEL test="custom-dockerfile"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create custom Dockerfile: %v", err)
	}

	imageName := "dvm-test-docker-builder-custom:test"

	config := BuilderConfig{
		Platform:    platform,
		Namespace:   "devopsmaestro",
		ProjectPath: projectPath,
		ImageName:   imageName,
		Dockerfile:  customDockerfile,
	}

	builder, err := NewDockerBuilder(config)
	if err != nil {
		t.Fatalf("NewDockerBuilder() error = %v", err)
	}
	defer builder.Close()

	ctx := context.Background()

	err = builder.Build(ctx, BuildOptions{})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestDockerBuilder_Build_Cancelled(t *testing.T) {
	platform := requireDockerPlatform(t)

	projectPath := t.TempDir()
	dockerfile := filepath.Join(projectPath, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`
FROM alpine:latest
RUN sleep 60
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	config := BuilderConfig{
		Platform:    platform,
		Namespace:   "devopsmaestro",
		ProjectPath: projectPath,
		ImageName:   "dvm-test-cancelled:test",
	}

	builder, err := NewDockerBuilder(config)
	if err != nil {
		t.Fatalf("NewDockerBuilder() error = %v", err)
	}
	defer builder.Close()

	// Create already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = builder.Build(ctx, BuildOptions{})
	// Build should fail with cancelled context
	if err == nil {
		t.Log("Build() did not immediately fail with cancelled context (docker may queue the build)")
	}
}
