package operators

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPlatformDetector(t *testing.T) {
	detector, err := NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}
	if detector == nil {
		t.Fatal("NewPlatformDetector() returned nil")
	}
	if detector.homeDir == "" {
		t.Error("NewPlatformDetector() homeDir is empty")
	}
}

func TestPlatformDetector_DetectWithEnvVar(t *testing.T) {
	// Skip if no platforms are available
	detector, err := NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platforms := detector.DetectAll()
	if len(platforms) == 0 {
		t.Skip("No container platforms available for testing")
	}

	tests := []struct {
		name         string
		envValue     string
		wantType     PlatformType
		wantErr      bool
		skipIfNoSock bool
	}{
		{
			name:         "orbstack env var",
			envValue:     "orbstack",
			wantType:     PlatformOrbStack,
			skipIfNoSock: true,
		},
		{
			name:         "colima env var",
			envValue:     "colima",
			wantType:     PlatformColima,
			skipIfNoSock: true,
		},
		{
			name:         "podman env var",
			envValue:     "podman",
			wantType:     PlatformPodman,
			skipIfNoSock: true,
		},
		{
			name:         "docker-desktop env var",
			envValue:     "docker-desktop",
			wantType:     PlatformDockerDesktop,
			skipIfNoSock: true,
		},
		{
			name:     "invalid platform",
			envValue: "invalid-platform",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if platform is available before testing
			if tt.skipIfNoSock {
				found := false
				for _, p := range platforms {
					if p.Type == tt.wantType {
						found = true
						break
					}
				}
				if !found {
					t.Skipf("Platform %s not available", tt.wantType)
				}
			}

			// Set env var
			oldVal := os.Getenv("DVM_PLATFORM")
			os.Setenv("DVM_PLATFORM", tt.envValue)
			defer os.Setenv("DVM_PLATFORM", oldVal)

			platform, err := detector.Detect()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Detect() expected error for env=%s", tt.envValue)
				}
				return
			}

			if err != nil {
				t.Errorf("Detect() error = %v for env=%s", err, tt.envValue)
				return
			}

			if platform.Type != tt.wantType {
				t.Errorf("Detect() got type %s, want %s", platform.Type, tt.wantType)
			}
		})
	}
}

func TestPlatformDetector_DetectAll(t *testing.T) {
	detector, err := NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platforms := detector.DetectAll()

	// Just verify we can call DetectAll without error
	// The actual platforms detected depend on the environment
	t.Logf("Detected %d platforms:", len(platforms))
	for _, p := range platforms {
		t.Logf("  - %s: %s (%s)", p.Type, p.Name, p.SocketPath)
	}
}

func TestPlatformDetector_AutoDetect(t *testing.T) {
	detector, err := NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	// Clear env var to ensure auto-detection
	oldVal := os.Getenv("DVM_PLATFORM")
	os.Unsetenv("DVM_PLATFORM")
	defer os.Setenv("DVM_PLATFORM", oldVal)

	platform, err := detector.Detect()

	// If no platforms are available, we expect an error
	platforms := detector.DetectAll()
	if len(platforms) == 0 {
		if err == nil {
			t.Error("Detect() should return error when no platforms available")
		}
		return
	}

	// Otherwise, we should get the first available platform
	if err != nil {
		t.Errorf("Detect() error = %v", err)
		return
	}

	if platform == nil {
		t.Error("Detect() returned nil platform")
		return
	}

	t.Logf("Auto-detected platform: %s (%s)", platform.Type, platform.Name)
}

func TestPlatform_IsContainerd(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     bool
	}{
		{
			name: "colima with containerd socket",
			platform: Platform{
				Type:       PlatformColima,
				SocketPath: "/home/user/.colima/default/containerd.sock",
			},
			want: true,
		},
		{
			name: "colima with docker socket",
			platform: Platform{
				Type:       PlatformColima,
				SocketPath: "/home/user/.colima/default/docker.sock",
			},
			want: false,
		},
		{
			name: "orbstack",
			platform: Platform{
				Type:       PlatformOrbStack,
				SocketPath: "/home/user/.orbstack/run/docker.sock",
			},
			want: false,
		},
		{
			name: "podman",
			platform: Platform{
				Type:       PlatformPodman,
				SocketPath: "/run/podman/podman.sock",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.IsContainerd(); got != tt.want {
				t.Errorf("IsContainerd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatform_IsDockerCompatible(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		want     bool
	}{
		{
			name: "orbstack",
			platform: Platform{
				Type: PlatformOrbStack,
			},
			want: true,
		},
		{
			name: "docker desktop",
			platform: Platform{
				Type: PlatformDockerDesktop,
			},
			want: true,
		},
		{
			name: "podman",
			platform: Platform{
				Type: PlatformPodman,
			},
			want: true,
		},
		{
			name: "linux native",
			platform: Platform{
				Type: PlatformLinuxNative,
			},
			want: true,
		},
		{
			name: "colima with docker",
			platform: Platform{
				Type:       PlatformColima,
				SocketPath: "/home/user/.colima/default/docker.sock",
			},
			want: true,
		},
		{
			name: "colima with containerd",
			platform: Platform{
				Type:       PlatformColima,
				SocketPath: "/home/user/.colima/default/containerd.sock",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.IsDockerCompatible(); got != tt.want {
				t.Errorf("IsDockerCompatible() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatform_GetBuildKitSocket(t *testing.T) {
	homeDir := "/home/testuser"

	tests := []struct {
		name     string
		platform Platform
		want     string
	}{
		{
			name: "colima default profile",
			platform: Platform{
				Type:    PlatformColima,
				Profile: "default",
				HomeDir: homeDir,
			},
			want: filepath.Join(homeDir, ".colima", "default", "buildkitd.sock"),
		},
		{
			name: "colima custom profile",
			platform: Platform{
				Type:    PlatformColima,
				Profile: "myprofile",
				HomeDir: homeDir,
			},
			want: filepath.Join(homeDir, ".colima", "myprofile", "buildkitd.sock"),
		},
		{
			name: "colima no profile",
			platform: Platform{
				Type:    PlatformColima,
				Profile: "",
				HomeDir: homeDir,
			},
			want: filepath.Join(homeDir, ".colima", "default", "buildkitd.sock"),
		},
		{
			name: "orbstack (no buildkit socket)",
			platform: Platform{
				Type:    PlatformOrbStack,
				HomeDir: homeDir,
			},
			want: "",
		},
		{
			name: "docker desktop (no buildkit socket)",
			platform: Platform{
				Type:    PlatformDockerDesktop,
				HomeDir: homeDir,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.GetBuildKitSocket(); got != tt.want {
				t.Errorf("GetBuildKitSocket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatform_GetContainerdSocket(t *testing.T) {
	homeDir := "/home/testuser"

	tests := []struct {
		name     string
		platform Platform
		want     string
	}{
		{
			name: "colima default profile",
			platform: Platform{
				Type:    PlatformColima,
				Profile: "default",
				HomeDir: homeDir,
			},
			want: filepath.Join(homeDir, ".colima", "default", "containerd.sock"),
		},
		{
			name: "colima custom profile",
			platform: Platform{
				Type:    PlatformColima,
				Profile: "devenv",
				HomeDir: homeDir,
			},
			want: filepath.Join(homeDir, ".colima", "devenv", "containerd.sock"),
		},
		{
			name: "orbstack (no containerd socket)",
			platform: Platform{
				Type:    PlatformOrbStack,
				HomeDir: homeDir,
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.platform.GetContainerdSocket(); got != tt.want {
				t.Errorf("GetContainerdSocket() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPlatform_GetStartHint(t *testing.T) {
	tests := []struct {
		name     string
		platform Platform
		contains string
	}{
		{
			name:     "orbstack",
			platform: Platform{Type: PlatformOrbStack},
			contains: "OrbStack",
		},
		{
			name:     "colima default",
			platform: Platform{Type: PlatformColima, Profile: "default"},
			contains: "colima start",
		},
		{
			name:     "colima custom profile",
			platform: Platform{Type: PlatformColima, Profile: "myvm"},
			contains: "--profile myvm",
		},
		{
			name:     "docker desktop",
			platform: Platform{Type: PlatformDockerDesktop},
			contains: "Docker Desktop",
		},
		{
			name:     "podman",
			platform: Platform{Type: PlatformPodman},
			contains: "podman machine start",
		},
		{
			name:     "linux native",
			platform: Platform{Type: PlatformLinuxNative},
			contains: "systemctl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hint := tt.platform.GetStartHint()
			if hint == "" {
				t.Error("GetStartHint() returned empty string")
			}
			if tt.contains != "" && !contains(hint, tt.contains) {
				t.Errorf("GetStartHint() = %q, should contain %q", hint, tt.contains)
			}
		})
	}
}

func TestPlatformTypes(t *testing.T) {
	// Verify platform type constants
	tests := []struct {
		platformType PlatformType
		expected     string
	}{
		{PlatformOrbStack, "orbstack"},
		{PlatformColima, "colima"},
		{PlatformDockerDesktop, "docker-desktop"},
		{PlatformPodman, "podman"},
		{PlatformLinuxNative, "linux-native"},
		{PlatformUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.platformType) != tt.expected {
				t.Errorf("PlatformType = %q, want %q", tt.platformType, tt.expected)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Integration tests that require actual runtimes running

func TestIntegration_OrbStackDetection(t *testing.T) {
	detector, err := NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platform := detector.detectOrbStack()
	if platform == nil {
		t.Skip("OrbStack not available")
	}

	if platform.Type != PlatformOrbStack {
		t.Errorf("detectOrbStack() type = %s, want %s", platform.Type, PlatformOrbStack)
	}

	if platform.SocketPath == "" {
		t.Error("detectOrbStack() socketPath is empty")
	}

	if _, err := os.Stat(platform.SocketPath); err != nil {
		t.Errorf("detectOrbStack() socket does not exist: %v", err)
	}

	t.Logf("OrbStack detected: %s", platform.SocketPath)
}

func TestIntegration_ColimaDetection(t *testing.T) {
	detector, err := NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platform := detector.detectColima()
	if platform == nil {
		t.Skip("Colima not available")
	}

	if platform.Type != PlatformColima {
		t.Errorf("detectColima() type = %s, want %s", platform.Type, PlatformColima)
	}

	if platform.SocketPath == "" {
		t.Error("detectColima() socketPath is empty")
	}

	if _, err := os.Stat(platform.SocketPath); err != nil {
		t.Errorf("detectColima() socket does not exist: %v", err)
	}

	t.Logf("Colima detected: %s (profile: %s, containerd: %v)",
		platform.SocketPath, platform.Profile, platform.IsContainerd())
}

func TestIntegration_PodmanDetection(t *testing.T) {
	detector, err := NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platform := detector.detectPodman()
	if platform == nil {
		t.Skip("Podman not available")
	}

	if platform.Type != PlatformPodman {
		t.Errorf("detectPodman() type = %s, want %s", platform.Type, PlatformPodman)
	}

	if platform.SocketPath == "" {
		t.Error("detectPodman() socketPath is empty")
	}

	if _, err := os.Stat(platform.SocketPath); err != nil {
		t.Errorf("detectPodman() socket does not exist: %v", err)
	}

	t.Logf("Podman detected: %s", platform.SocketPath)
}
