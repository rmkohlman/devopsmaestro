package builders

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"devopsmaestro/operators"
)

func TestBuildKitBuilder_Implements_ImageBuilder(t *testing.T) {
	// Compile-time check that BuildKitBuilder implements ImageBuilder
	var _ ImageBuilder = (*BuildKitBuilder)(nil)
}

func TestNewBuildKitBuilder_MissingContainerdSocket(t *testing.T) {
	// Platform without containerd socket should fail
	config := BuilderConfig{
		Platform: &operators.Platform{
			Type:       operators.PlatformOrbStack, // OrbStack doesn't have containerd socket
			SocketPath: "/test/docker.sock",
			HomeDir:    "/home/user",
		},
		AppPath:   "/tmp/test",
		ImageName: "test:latest",
	}

	builder, err := NewBuildKitBuilder(config)
	if err == nil {
		t.Error("NewBuildKitBuilder() should fail without containerd socket")
		if builder != nil {
			builder.Close()
		}
	}
}

func TestNewBuildKitBuilder_InvalidSockets(t *testing.T) {
	// Skip if Colima is running (because it will use real sockets)
	detector, err := operators.NewPlatformDetector()
	if err == nil {
		for _, p := range detector.DetectAll() {
			if p.Type == operators.PlatformColima && p.IsContainerd() {
				t.Skip("Colima is running, cannot test invalid socket behavior")
			}
		}
	}

	config := BuilderConfig{
		Platform: &operators.Platform{
			Type:       operators.PlatformColima,
			SocketPath: "/nonexistent/containerd.sock",
			Profile:    "nonexistent",
			HomeDir:    "/nonexistent-home-dir-12345",
		},
		AppPath:   "/tmp/test",
		ImageName: "test:latest",
	}

	builder, err := NewBuildKitBuilder(config)
	if err == nil {
		t.Error("NewBuildKitBuilder() should fail with invalid sockets")
		if builder != nil {
			builder.Close()
		}
	}
}

// Integration tests requiring Colima with containerd

func requireContainerdPlatform(t *testing.T) *operators.Platform {
	t.Helper()

	detector, err := operators.NewPlatformDetector()
	if err != nil {
		t.Skipf("Failed to create platform detector: %v", err)
	}

	platforms := detector.DetectAll()
	for _, p := range platforms {
		if p.IsContainerd() {
			return p
		}
	}

	t.Skip("No containerd platform available (Colima with containerd required)")
	return nil
}

func TestIntegration_BuildKitBuilder_New(t *testing.T) {
	platform := requireContainerdPlatform(t)

	appPath := t.TempDir()

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: "test-buildkit-builder:latest",
	}

	builder, err := NewBuildKitBuilder(config)
	if err != nil {
		t.Fatalf("NewBuildKitBuilder() error = %v", err)
	}
	defer builder.Close()

	// Verify internal state
	if builder.namespace != "devopsmaestro" {
		t.Errorf("namespace = %q, want %q", builder.namespace, "devopsmaestro")
	}
	if builder.appPath != appPath {
		t.Errorf("appPath = %q, want %q", builder.appPath, appPath)
	}
	if builder.imageName != "test-buildkit-builder:latest" {
		t.Errorf("imageName = %q, want %q", builder.imageName, "test-buildkit-builder:latest")
	}
	if builder.containerdClient == nil {
		t.Error("containerdClient is nil")
	}
	if builder.buildkitClient == nil {
		t.Error("buildkitClient is nil")
	}
}

func TestIntegration_BuildKitBuilder_Close(t *testing.T) {
	platform := requireContainerdPlatform(t)

	appPath := t.TempDir()

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: "test-buildkit-close:latest",
	}

	builder, err := NewBuildKitBuilder(config)
	if err != nil {
		t.Fatalf("NewBuildKitBuilder() error = %v", err)
	}

	// Close should release resources
	err = builder.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Verify clients are closed (second close should be safe)
	err = builder.Close()
	// May or may not error on double close, but shouldn't panic
	_ = err
}

func TestIntegration_BuildKitBuilder_ImageExists_NotFound(t *testing.T) {
	platform := requireContainerdPlatform(t)

	appPath := t.TempDir()

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: "nonexistent-buildkit-image-" + t.Name() + ":v999",
	}

	builder, err := NewBuildKitBuilder(config)
	if err != nil {
		t.Fatalf("NewBuildKitBuilder() error = %v", err)
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

func TestIntegration_BuildKitBuilder_Build(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	platform := requireContainerdPlatform(t)

	// Create temporary project with Dockerfile
	appPath := t.TempDir()
	dockerfile := filepath.Join(appPath, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`
FROM alpine:latest
RUN echo "buildkit test"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	imageName := "dvm-test-buildkit-builder:test"

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: imageName,
	}

	builder, err := NewBuildKitBuilder(config)
	if err != nil {
		t.Fatalf("NewBuildKitBuilder() error = %v", err)
	}
	defer builder.Close()

	ctx := context.Background()

	// Build the image
	err = builder.Build(ctx, BuildOptions{})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Note: ImageExists checks containerd's image store, but BuildKit exports
	// images to its own storage by default. The image won't be visible in
	// containerd until it's explicitly imported or the export type is changed
	// to "containerd" or "oci". For now, we just verify the build succeeded.
	//
	// TODO: Consider adding "containerd" export or implementing ImageExists
	// to check BuildKit's image store directly.
}

func TestIntegration_BuildKitBuilder_Build_WithBuildArgs(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	platform := requireContainerdPlatform(t)

	appPath := t.TempDir()
	dockerfile := filepath.Join(appPath, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`
FROM alpine:latest
ARG VERSION=unknown
ARG ENVIRONMENT=dev
RUN echo "Version: ${VERSION}, Env: ${ENVIRONMENT}"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: "dvm-test-buildkit-args:test",
	}

	builder, err := NewBuildKitBuilder(config)
	if err != nil {
		t.Fatalf("NewBuildKitBuilder() error = %v", err)
	}
	defer builder.Close()

	ctx := context.Background()

	err = builder.Build(ctx, BuildOptions{
		BuildArgs: map[string]string{
			"VERSION":     "1.2.3",
			"ENVIRONMENT": "production",
		},
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestIntegration_BuildKitBuilder_Build_WithTarget(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	platform := requireContainerdPlatform(t)

	appPath := t.TempDir()
	dockerfile := filepath.Join(appPath, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`
FROM alpine:latest AS builder
RUN echo "building"

FROM alpine:latest AS runtime
RUN echo "runtime"

FROM alpine:latest AS test
RUN echo "testing"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: "dvm-test-buildkit-target:test",
	}

	builder, err := NewBuildKitBuilder(config)
	if err != nil {
		t.Fatalf("NewBuildKitBuilder() error = %v", err)
	}
	defer builder.Close()

	ctx := context.Background()

	// Build only the builder stage
	err = builder.Build(ctx, BuildOptions{
		Target: "builder",
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestIntegration_BuildKitBuilder_Build_NoCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	platform := requireContainerdPlatform(t)

	appPath := t.TempDir()
	dockerfile := filepath.Join(appPath, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`
FROM alpine:latest
RUN echo "no cache test"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: "dvm-test-buildkit-nocache:test",
	}

	builder, err := NewBuildKitBuilder(config)
	if err != nil {
		t.Fatalf("NewBuildKitBuilder() error = %v", err)
	}
	defer builder.Close()

	ctx := context.Background()

	err = builder.Build(ctx, BuildOptions{
		NoCache: true,
	})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}

func TestIntegration_BuildKitBuilder_Build_CustomDockerfile(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	platform := requireContainerdPlatform(t)

	appPath := t.TempDir()

	// Create custom Dockerfile
	customDockerfile := filepath.Join(appPath, "Dockerfile.buildkit")
	err := os.WriteFile(customDockerfile, []byte(`
FROM alpine:latest
LABEL builder="buildkit"
RUN echo "custom dockerfile for buildkit"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create custom Dockerfile: %v", err)
	}

	config := BuilderConfig{
		Platform:   platform,
		Namespace:  "devopsmaestro",
		AppPath:    appPath,
		ImageName:  "dvm-test-buildkit-custom:test",
		Dockerfile: customDockerfile,
	}

	builder, err := NewBuildKitBuilder(config)
	if err != nil {
		t.Fatalf("NewBuildKitBuilder() error = %v", err)
	}
	defer builder.Close()

	ctx := context.Background()

	err = builder.Build(ctx, BuildOptions{})
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
}
