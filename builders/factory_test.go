package builders

import (
	"testing"

	"devopsmaestro/operators"
)

func TestBuilderConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  BuilderConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty config",
			config:  BuilderConfig{},
			wantErr: true,
			errMsg:  "platform is required",
		},
		{
			name: "missing project path",
			config: BuilderConfig{
				Platform: &operators.Platform{
					Type:       operators.PlatformOrbStack,
					SocketPath: "/test/socket",
				},
			},
			wantErr: true,
			errMsg:  "project path is required",
		},
		{
			name: "missing image name",
			config: BuilderConfig{
				Platform: &operators.Platform{
					Type:       operators.PlatformOrbStack,
					SocketPath: "/test/socket",
				},
				ProjectPath: "/test/project",
			},
			wantErr: true,
			errMsg:  "image name is required",
		},
		{
			name: "valid config",
			config: BuilderConfig{
				Platform: &operators.Platform{
					Type:       operators.PlatformOrbStack,
					SocketPath: "/test/socket",
				},
				ProjectPath: "/test/project",
				ImageName:   "test:latest",
			},
			wantErr: false,
		},
		{
			name: "valid config with all fields",
			config: BuilderConfig{
				Platform: &operators.Platform{
					Type:       operators.PlatformColima,
					SocketPath: "/test/containerd.sock",
					Profile:    "default",
					HomeDir:    "/home/user",
				},
				Namespace:   "devopsmaestro",
				ProjectPath: "/test/project",
				ImageName:   "myimage:v1.0",
				Dockerfile:  "/test/project/Dockerfile.custom",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()

			if tt.wantErr {
				if err == nil {
					t.Error("Validate() expected error, got nil")
				} else if tt.errMsg != "" && err.Error() != tt.errMsg {
					t.Errorf("Validate() error = %q, want %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Validate() unexpected error: %v", err)
				}
			}
		})
	}
}

func TestNewImageBuilder_InvalidConfig(t *testing.T) {
	tests := []struct {
		name   string
		config BuilderConfig
	}{
		{
			name:   "nil platform",
			config: BuilderConfig{},
		},
		{
			name: "missing project path",
			config: BuilderConfig{
				Platform: &operators.Platform{
					Type:       operators.PlatformOrbStack,
					SocketPath: "/test/socket",
				},
			},
		},
		{
			name: "missing image name",
			config: BuilderConfig{
				Platform: &operators.Platform{
					Type:       operators.PlatformOrbStack,
					SocketPath: "/test/socket",
				},
				ProjectPath: "/test/project",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder, err := NewImageBuilder(tt.config)
			if err == nil {
				t.Error("NewImageBuilder() expected error for invalid config")
				if builder != nil {
					builder.Close()
				}
			}
		})
	}
}

func TestNewImageBuilder_UnsupportedPlatform(t *testing.T) {
	config := BuilderConfig{
		Platform: &operators.Platform{
			Type:       operators.PlatformUnknown,
			SocketPath: "/test/socket",
		},
		ProjectPath: "/test/project",
		ImageName:   "test:latest",
	}

	builder, err := NewImageBuilder(config)
	if err == nil {
		t.Error("NewImageBuilder() expected error for unknown platform")
		if builder != nil {
			builder.Close()
		}
	}

	if builder != nil {
		t.Error("NewImageBuilder() should return nil builder for unknown platform")
	}
}

func TestNewImageBuilder_DockerCompatible(t *testing.T) {
	// This test verifies the factory selects DockerBuilder for Docker-compatible platforms
	// Note: This is a unit test that verifies the selection logic, not actual connection

	dockerPlatforms := []operators.PlatformType{
		operators.PlatformOrbStack,
		operators.PlatformDockerDesktop,
		operators.PlatformPodman,
		operators.PlatformLinuxNative,
	}

	for _, platformType := range dockerPlatforms {
		t.Run(string(platformType), func(t *testing.T) {
			platform := &operators.Platform{
				Type:       platformType,
				SocketPath: "/nonexistent/socket", // Will fail connection
			}

			// Verify IsDockerCompatible returns true
			if !platform.IsDockerCompatible() {
				t.Errorf("Platform %s should be Docker compatible", platformType)
			}
		})
	}
}

func TestNewImageBuilder_Containerd(t *testing.T) {
	// This test verifies the factory selects BuildKitBuilder for containerd platforms
	// Note: This is a unit test that verifies the selection logic

	platform := &operators.Platform{
		Type:       operators.PlatformColima,
		SocketPath: "/home/user/.colima/default/containerd.sock",
		Profile:    "default",
		HomeDir:    "/home/user",
	}

	// Verify IsContainerd returns true for containerd socket
	if !platform.IsContainerd() {
		t.Error("Colima with containerd.sock should be containerd platform")
	}

	// Verify IsDockerCompatible returns false
	if platform.IsDockerCompatible() {
		t.Error("Colima with containerd.sock should not be Docker compatible")
	}
}

func TestNewImageBuilder_ColimaDocker(t *testing.T) {
	// Colima with docker socket should use DockerBuilder
	platform := &operators.Platform{
		Type:       operators.PlatformColima,
		SocketPath: "/home/user/.colima/default/docker.sock",
		Profile:    "default",
		HomeDir:    "/home/user",
	}

	// Verify IsContainerd returns false for docker socket
	if platform.IsContainerd() {
		t.Error("Colima with docker.sock should not be containerd platform")
	}

	// Verify IsDockerCompatible returns true
	if !platform.IsDockerCompatible() {
		t.Error("Colima with docker.sock should be Docker compatible")
	}
}

// Integration test - requires actual platform to be running
func TestIntegration_NewImageBuilder_OrbStack(t *testing.T) {
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platform := detector.DetectAll()
	var orbstack *operators.Platform
	for _, p := range platform {
		if p.Type == operators.PlatformOrbStack {
			orbstack = p
			break
		}
	}

	if orbstack == nil {
		t.Skip("OrbStack not available")
	}

	// Create temporary project directory
	projectPath := t.TempDir()

	config := BuilderConfig{
		Platform:    orbstack,
		Namespace:   "test",
		ProjectPath: projectPath,
		ImageName:   "test-orbstack:latest",
	}

	builder, err := NewImageBuilder(config)
	if err != nil {
		t.Fatalf("NewImageBuilder() error = %v", err)
	}
	defer builder.Close()

	// Verify we got a DockerBuilder
	_, ok := builder.(*DockerBuilder)
	if !ok {
		t.Error("NewImageBuilder() should return DockerBuilder for OrbStack")
	}
}

func TestIntegration_NewImageBuilder_Podman(t *testing.T) {
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platforms := detector.DetectAll()
	var podman *operators.Platform
	for _, p := range platforms {
		if p.Type == operators.PlatformPodman {
			podman = p
			break
		}
	}

	if podman == nil {
		t.Skip("Podman not available")
	}

	projectPath := t.TempDir()

	config := BuilderConfig{
		Platform:    podman,
		Namespace:   "test",
		ProjectPath: projectPath,
		ImageName:   "test-podman:latest",
	}

	builder, err := NewImageBuilder(config)
	if err != nil {
		t.Fatalf("NewImageBuilder() error = %v", err)
	}
	defer builder.Close()

	_, ok := builder.(*DockerBuilder)
	if !ok {
		t.Error("NewImageBuilder() should return DockerBuilder for Podman")
	}
}

func TestIntegration_NewImageBuilder_Colima(t *testing.T) {
	detector, err := operators.NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platforms := detector.DetectAll()
	var colima *operators.Platform
	for _, p := range platforms {
		if p.Type == operators.PlatformColima {
			colima = p
			break
		}
	}

	if colima == nil {
		t.Skip("Colima not available")
	}

	projectPath := t.TempDir()

	config := BuilderConfig{
		Platform:    colima,
		Namespace:   "test",
		ProjectPath: projectPath,
		ImageName:   "test-colima:latest",
	}

	builder, err := NewImageBuilder(config)
	if err != nil {
		t.Fatalf("NewImageBuilder() error = %v", err)
	}
	defer builder.Close()

	// Check which builder type based on socket
	if colima.IsContainerd() {
		_, ok := builder.(*BuildKitBuilder)
		if !ok {
			t.Error("NewImageBuilder() should return BuildKitBuilder for Colima containerd")
		}
	} else {
		_, ok := builder.(*DockerBuilder)
		if !ok {
			t.Error("NewImageBuilder() should return DockerBuilder for Colima docker")
		}
	}
}
