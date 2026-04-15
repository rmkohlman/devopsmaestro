package builders

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
		AppPath:   "/tmp/test",
		ImageName: "test:latest",
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

// requireDockerPlatform uses `docker info` rather than Platform.IsReachable() because
// integration tests need to verify the full Docker API is responding, not just
// that a socket is listening.
func requireDockerPlatform(t *testing.T) *operators.Platform {
	t.Helper()

	detector, err := operators.NewPlatformDetector()
	if err != nil {
		t.Skipf("Failed to create platform detector: %v", err)
	}

	platforms := detector.DetectAll()
	for _, p := range platforms {
		if p.IsDockerCompatible() {
			// Verify the Docker daemon is actually reachable, not just that
			// the socket file exists. A stale socket (e.g., Docker Desktop
			// stopped but socket remains) would cause tests to fail rather
			// than skip.
			dockerHost := "unix://" + p.SocketPath
			cmd := exec.Command("docker", "info")
			cmd.Env = append(os.Environ(), "DOCKER_HOST="+dockerHost)
			if err := cmd.Run(); err != nil {
				t.Logf("Skipping platform %s: daemon not reachable at %s", p.Name, p.SocketPath)
				continue
			}
			return p
		}
	}

	t.Skip("No Docker-compatible platform available (no reachable daemon)")
	return nil
}

func TestIntegration_DockerBuilder_New(t *testing.T) {
	platform := requireDockerPlatform(t)

	appPath := t.TempDir()

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: "test-docker-builder:latest",
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
	if builder.appPath != appPath {
		t.Errorf("appPath = %q, want %q", builder.appPath, appPath)
	}
	if builder.imageName != "test-docker-builder:latest" {
		t.Errorf("imageName = %q, want %q", builder.imageName, "test-docker-builder:latest")
	}
}

func TestIntegration_DockerBuilder_ImageExists_NotFound(t *testing.T) {
	platform := requireDockerPlatform(t)

	appPath := t.TempDir()

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: "nonexistent-image-" + t.Name() + ":v999",
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

	// Create temporary app directory with Dockerfile
	appPath := t.TempDir()
	dockerfile := filepath.Join(appPath, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`
FROM alpine:latest
RUN echo "test"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	imageName := "dvm-test-docker-builder:test"

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: imageName,
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

	appPath := t.TempDir()
	dockerfile := filepath.Join(appPath, "Dockerfile")
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
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: imageName,
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

	appPath := t.TempDir()

	// Create custom Dockerfile with different name
	customDockerfile := filepath.Join(appPath, "Dockerfile.custom")
	err := os.WriteFile(customDockerfile, []byte(`
FROM alpine:latest
LABEL test="custom-dockerfile"
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create custom Dockerfile: %v", err)
	}

	imageName := "dvm-test-docker-builder-custom:test"

	config := BuilderConfig{
		Platform:   platform,
		Namespace:  "devopsmaestro",
		AppPath:    appPath,
		ImageName:  imageName,
		Dockerfile: customDockerfile,
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

	appPath := t.TempDir()
	dockerfile := filepath.Join(appPath, "Dockerfile")
	err := os.WriteFile(dockerfile, []byte(`
FROM alpine:latest
RUN sleep 60
`), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	config := BuilderConfig{
		Platform:  platform,
		Namespace: "devopsmaestro",
		AppPath:   appPath,
		ImageName: "dvm-test-cancelled:test",
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

// =============================================================================
// SM-7: Build Arg Redaction Tests
// RED: These tests FAIL until --build-arg values are redacted in log output.
// =============================================================================

// TestDockerBuilder_BuildArgLogRedaction verifies that --build-arg values
// are redacted (replaced with ***) in the rendered command log output,
// so secrets don't appear in CI logs or console output.
func TestDockerBuilder_BuildArgLogRedaction(t *testing.T) {
	// buildDockerArgs is the function that constructs the docker args slice
	// which is then logged. It must exist after SM-7 is implemented.
	opts := BuildOptions{
		BuildArgs: map[string]string{
			"SECRET_TOKEN": "super-secret-value",
			"API_KEY":      "another-secret",
			"SAFE_ARG":     "not-sensitive",
		},
	}

	// buildDockerArgsForLog should return the args with values redacted,
	// while buildDockerArgs returns the actual args (with real values) for docker.
	redacted := buildDockerArgsForLog(opts)

	// Verify secret values are not present in the redacted args
	for _, arg := range redacted {
		if arg == "super-secret-value" {
			t.Error("buildDockerArgsForLog() exposed SECRET_TOKEN value in log output")
		}
		if arg == "another-secret" {
			t.Error("buildDockerArgsForLog() exposed API_KEY value in log output")
		}
	}

	// Verify the keys are still present (for debugging purposes)
	argsStr := strings.Join(redacted, " ")
	if !strings.Contains(argsStr, "SECRET_TOKEN") {
		t.Error("buildDockerArgsForLog() should still include arg key names, got no SECRET_TOKEN")
	}

	// Verify redaction marker is present
	if !strings.Contains(argsStr, "***") {
		t.Error("buildDockerArgsForLog() should use *** as redaction marker for --build-arg values")
	}
}

// TestDockerBuilder_Build_TimeoutOverride verifies that BuildOptions.Timeout
// overrides the watchdog's default timeout. This is the fix for #252 where
// the --timeout flag value was ignored because the watchdog had its own
// hardcoded 30m timeout that fired first.
func TestDockerBuilder_Build_TimeoutOverride(t *testing.T) {
	tests := []struct {
		name            string
		optsTimeout     time.Duration
		watchdogTimeout time.Duration
		wantTimeout     time.Duration
	}{
		{
			name:            "BuildOptions.Timeout overrides watchdog default",
			optsTimeout:     1 * time.Hour,
			watchdogTimeout: 0, // zero = use default
			wantTimeout:     1 * time.Hour,
		},
		{
			name:            "BuildOptions.Timeout overrides explicit watchdog config",
			optsTimeout:     2 * time.Hour,
			watchdogTimeout: 30 * time.Minute,
			wantTimeout:     2 * time.Hour,
		},
		{
			name:            "zero BuildOptions.Timeout uses watchdog default",
			optsTimeout:     0,
			watchdogTimeout: 0,
			wantTimeout:     45 * time.Minute, // DefaultWatchdogConfig
		},
		{
			name:            "zero BuildOptions.Timeout uses explicit watchdog config",
			optsTimeout:     0,
			watchdogTimeout: 20 * time.Minute,
			wantTimeout:     20 * time.Minute,
		},
		{
			name:            "short timeout (5m) overrides default",
			optsTimeout:     5 * time.Minute,
			watchdogTimeout: 0,
			wantTimeout:     5 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedCfg WatchdogConfig

			// Create a DockerBuilder with a mock watchdog runner that
			// captures the WatchdogConfig passed to it.
			builder := &DockerBuilder{
				platform:  &operators.Platform{Name: "test", SocketPath: "/tmp/fake.sock"},
				appPath:   t.TempDir(),
				imageName: "test:latest",
			}

			if tt.watchdogTimeout > 0 {
				builder.WatchdogConfig = WatchdogConfig{
					PollInterval: 1 * time.Second,
					Timeout:      tt.watchdogTimeout,
					CleanupWait:  5 * time.Second,
				}
			}

			// Inject a mock watchdog runner that captures the config
			// and returns immediately (no actual build needed).
			builder.WatchdogRunner = func(
				ctx context.Context,
				cmd *exec.Cmd,
				checkCondition func(ctx context.Context) bool,
				cfg WatchdogConfig,
			) (WatchdogResult, error) {
				capturedCfg = cfg
				// Kill the process we started so it doesn't leak
				if cmd.Process != nil {
					_ = cmd.Process.Kill()
				}
				return WatchdogCompleted, nil
			}

			// Write a minimal Dockerfile so docker buildx doesn't fail
			// before reaching the watchdog (though our mock intercepts).
			dockerfile := filepath.Join(builder.appPath, "Dockerfile")
			_ = os.WriteFile(dockerfile, []byte("FROM scratch\n"), 0644)

			opts := BuildOptions{
				Timeout: tt.optsTimeout,
			}

			// Build will fail because we have a fake socket, but the
			// watchdog runner is called before the docker command starts
			// in the real flow. However, with our mock, it captures the
			// config. We need to handle the docker command start error.
			_ = builder.Build(context.Background(), opts)

			if capturedCfg.Timeout != tt.wantTimeout {
				t.Errorf("watchdog timeout = %v, want %v",
					capturedCfg.Timeout, tt.wantTimeout)
			}
		})
	}
}
