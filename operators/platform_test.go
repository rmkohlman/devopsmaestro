package operators

import (
	"net"
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
	// Skip if no reachable platforms are available.
	// DetectAll() now returns ALL detected platforms (including unreachable), so we
	// use DetectReachable() here to only proceed when something is actually running.
	detector, err := NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platforms := detector.DetectReachable()
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

	// If no reachable platforms are available, we expect an error.
	// DetectAll() now returns unfiltered results; use DetectReachable() to
	// determine whether any runtime is actually running.
	platforms := detector.DetectReachable()
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

// ---------------------------------------------------------------------------
// New tests for detectAllColima, IsReachable, and DetectAll filtering
// ---------------------------------------------------------------------------

func TestPlatformDetector_DetectAllColima(t *testing.T) {
	tests := []struct {
		name         string
		setupFunc    func(t *testing.T, colimaDir string)
		envProfile   string // set COLIMA_DOCKER_PROFILE when non-empty
		wantCount    int
		wantProfiles []string // profiles that must appear in results (order-independent)
	}{
		{
			name: "two profiles one docker sock one containerd sock returns two platforms",
			setupFunc: func(t *testing.T, colimaDir string) {
				t.Helper()
				// profile "alpha" – docker socket
				alphaDir := filepath.Join(colimaDir, "alpha")
				if err := os.MkdirAll(alphaDir, 0o755); err != nil {
					t.Fatalf("mkdir alpha: %v", err)
				}
				if _, err := os.Create(filepath.Join(alphaDir, "docker.sock")); err != nil {
					t.Fatalf("create docker.sock: %v", err)
				}
				// profile "beta" – containerd socket
				betaDir := filepath.Join(colimaDir, "beta")
				if err := os.MkdirAll(betaDir, 0o755); err != nil {
					t.Fatalf("mkdir beta: %v", err)
				}
				if _, err := os.Create(filepath.Join(betaDir, "containerd.sock")); err != nil {
					t.Fatalf("create containerd.sock: %v", err)
				}
			},
			wantCount:    2,
			wantProfiles: []string{"alpha", "beta"},
		},
		{
			name: "one profile with both sockets returns one platform preferring docker",
			setupFunc: func(t *testing.T, colimaDir string) {
				t.Helper()
				profileDir := filepath.Join(colimaDir, "default")
				if err := os.MkdirAll(profileDir, 0o755); err != nil {
					t.Fatalf("mkdir default: %v", err)
				}
				if _, err := os.Create(filepath.Join(profileDir, "docker.sock")); err != nil {
					t.Fatalf("create docker.sock: %v", err)
				}
				if _, err := os.Create(filepath.Join(profileDir, "containerd.sock")); err != nil {
					t.Fatalf("create containerd.sock: %v", err)
				}
			},
			wantCount:    1,
			wantProfiles: []string{"default"},
		},
		{
			name: "empty colima dir returns empty slice",
			setupFunc: func(t *testing.T, colimaDir string) {
				t.Helper()
				// colimaDir already exists (created by test harness); nothing to add
			},
			wantCount: 0,
		},
		{
			name: "only internal dirs are skipped",
			setupFunc: func(t *testing.T, colimaDir string) {
				t.Helper()
				for _, internal := range []string{"_lima", "_store"} {
					dir := filepath.Join(colimaDir, internal)
					if err := os.MkdirAll(dir, 0o755); err != nil {
						t.Fatalf("mkdir %s: %v", internal, err)
					}
					// Even if a socket file exists inside, it must be ignored
					if _, err := os.Create(filepath.Join(dir, "docker.sock")); err != nil {
						t.Fatalf("create docker.sock in %s: %v", internal, err)
					}
				}
			},
			wantCount: 0,
		},
		{
			name: "COLIMA_DOCKER_PROFILE set returns only that profile",
			setupFunc: func(t *testing.T, colimaDir string) {
				t.Helper()
				// profile "selected" with docker socket
				selectedDir := filepath.Join(colimaDir, "selected")
				if err := os.MkdirAll(selectedDir, 0o755); err != nil {
					t.Fatalf("mkdir selected: %v", err)
				}
				if _, err := os.Create(filepath.Join(selectedDir, "docker.sock")); err != nil {
					t.Fatalf("create docker.sock: %v", err)
				}
				// profile "other" that should be ignored
				otherDir := filepath.Join(colimaDir, "other")
				if err := os.MkdirAll(otherDir, 0o755); err != nil {
					t.Fatalf("mkdir other: %v", err)
				}
				if _, err := os.Create(filepath.Join(otherDir, "docker.sock")); err != nil {
					t.Fatalf("create docker.sock in other: %v", err)
				}
			},
			envProfile:   "selected",
			wantCount:    1,
			wantProfiles: []string{"selected"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a fake home dir: <tempDir>/home
			tempHome := t.TempDir()
			colimaDir := filepath.Join(tempHome, ".colima")
			if err := os.MkdirAll(colimaDir, 0o755); err != nil {
				t.Fatalf("mkdir .colima: %v", err)
			}

			tt.setupFunc(t, colimaDir)

			// Manage COLIMA_DOCKER_PROFILE env var
			if tt.envProfile != "" {
				oldVal := os.Getenv("COLIMA_DOCKER_PROFILE")
				os.Setenv("COLIMA_DOCKER_PROFILE", tt.envProfile)
				defer os.Setenv("COLIMA_DOCKER_PROFILE", oldVal)
			}

			detector := &PlatformDetector{homeDir: tempHome}
			platforms := detector.detectAllColima()

			if len(platforms) != tt.wantCount {
				t.Errorf("detectAllColima() returned %d platforms, want %d", len(platforms), tt.wantCount)
				for _, p := range platforms {
					t.Logf("  got: profile=%s socket=%s", p.Profile, p.SocketPath)
				}
			}

			// Verify that all expected profiles are present
			for _, wantProfile := range tt.wantProfiles {
				found := false
				for _, p := range platforms {
					if p.Profile == wantProfile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("detectAllColima() missing expected profile %q", wantProfile)
				}
			}

			// Verify all returned platforms have correct type and non-empty socket
			for _, p := range platforms {
				if p.Type != PlatformColima {
					t.Errorf("platform type = %s, want %s", p.Type, PlatformColima)
				}
				if p.SocketPath == "" {
					t.Errorf("platform profile=%s has empty SocketPath", p.Profile)
				}
				if p.HomeDir != tempHome {
					t.Errorf("platform profile=%s HomeDir = %s, want %s", p.Profile, p.HomeDir, tempHome)
				}
			}
		})
	}
}

func TestPlatform_IsReachable(t *testing.T) {
	t.Run("reachable unix socket returns true", func(t *testing.T) {
		// Use /tmp instead of t.TempDir() because macOS has a 104-byte limit
		// for Unix socket paths. t.TempDir() produces paths under /var/folders/...
		// that easily exceed this limit (e.g. 120+ bytes with subtest names).
		tmpDir, err := os.MkdirTemp("/tmp", "sock")
		if err != nil {
			t.Fatalf("os.MkdirTemp: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		socketPath := filepath.Join(tmpDir, "test.sock")

		ln, err := net.Listen("unix", socketPath)
		if err != nil {
			t.Fatalf("net.Listen unix: %v", err)
		}
		defer ln.Close()

		p := &Platform{
			Type:       PlatformColima,
			SocketPath: socketPath,
		}
		if !p.IsReachable() {
			t.Error("IsReachable() = false, want true for active unix listener")
		}
	})

	t.Run("regular file not listening returns false", func(t *testing.T) {
		socketPath := filepath.Join(t.TempDir(), "fake.sock")
		if _, err := os.Create(socketPath); err != nil {
			t.Fatalf("os.Create: %v", err)
		}

		p := &Platform{
			Type:       PlatformColima,
			SocketPath: socketPath,
		}
		if p.IsReachable() {
			t.Error("IsReachable() = true, want false for plain file (not a listener)")
		}
	})

	t.Run("non-existent path returns false", func(t *testing.T) {
		p := &Platform{
			Type:       PlatformColima,
			SocketPath: filepath.Join(t.TempDir(), "does_not_exist.sock"),
		}
		if p.IsReachable() {
			t.Error("IsReachable() = true, want false for non-existent path")
		}
	})
}

// TestPlatformDetector_DetectAll_IncludesUnreachable verifies that DetectAll() returns
// ALL detected platforms regardless of reachability (stale sockets included).
// It does NOT assert that sockets are reachable — that is DetectReachable()'s job.
func TestPlatformDetector_DetectAll_IncludesUnreachable(t *testing.T) {
	detector, err := NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platforms := detector.DetectAll()

	// Smoke test: DetectAll must not panic or error.
	// The returned slice may contain unreachable platforms (stale sockets).
	t.Logf("DetectAll() returned %d platform(s) (may include unreachable):", len(platforms))
	for _, p := range platforms {
		t.Logf("  - type=%s name=%q socket=%s reachable=%v",
			p.Type, p.Name, p.SocketPath, p.IsReachable())
	}

	// Every returned platform must have a non-empty SocketPath.
	// We do NOT assert os.Stat succeeds — sockets may be stale but still returned.
	for _, p := range platforms {
		if p.SocketPath == "" {
			t.Errorf("platform %s has empty SocketPath", p.Type)
		}
	}
}

// TestPlatformDetector_DetectReachable verifies that DetectReachable() only returns
// platforms where IsReachable() returns true.
func TestPlatformDetector_DetectReachable(t *testing.T) {
	detector, err := NewPlatformDetector()
	if err != nil {
		t.Fatalf("NewPlatformDetector() error = %v", err)
	}

	platforms := detector.DetectReachable()

	t.Logf("DetectReachable() returned %d platform(s):", len(platforms))
	for _, p := range platforms {
		t.Logf("  - type=%s name=%q socket=%s", p.Type, p.Name, p.SocketPath)
	}

	// Every platform returned by DetectReachable() MUST pass IsReachable().
	for _, p := range platforms {
		if !p.IsReachable() {
			t.Errorf("DetectReachable() returned platform %s (socket=%s) that is NOT reachable",
				p.Type, p.SocketPath)
		}
	}
}

// TestIsValidColimaProfileName exercises the profile-name validation regex.
func TestIsValidColimaProfileName(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		want    bool
	}{
		// Valid profile names
		{name: "default", profile: "default", want: true},
		{name: "k3s", profile: "k3s", want: true},
		{name: "my-profile", profile: "my-profile", want: true},
		{name: "dev.env", profile: "dev.env", want: true},
		{name: "test_1", profile: "test_1", want: true},
		// Invalid profile names
		{name: "empty string", profile: "", want: false},
		{name: "path traversal", profile: "../etc", want: false},
		{name: "hidden file dot prefix", profile: ".hidden", want: false},
		{name: "underscore prefix", profile: "_internal", want: false},
		{name: "has space", profile: "has space", want: false},
		{name: "semicolon", profile: "semi;colon", want: false},
		{name: "pipe character", profile: "pipe|char", want: false},
		{name: "backtick", profile: "back`tick", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidColimaProfileName(tt.profile)
			if got != tt.want {
				t.Errorf("isValidColimaProfileName(%q) = %v, want %v", tt.profile, got, tt.want)
			}
		})
	}
}

// TestPlatformDetector_DetectAllColima_SymlinksSkipped verifies that detectAllColima()
// ignores symlink entries in the ~/.colima directory, even when the symlink target
// contains a valid socket file.
func TestPlatformDetector_DetectAllColima_SymlinksSkipped(t *testing.T) {
	tempHome := t.TempDir()
	colimaDir := filepath.Join(tempHome, ".colima")
	if err := os.MkdirAll(colimaDir, 0o755); err != nil {
		t.Fatalf("mkdir .colima: %v", err)
	}

	// Create a real profile directory with a socket file.
	realProfileDir := filepath.Join(tempHome, "real-profile-outside-colima")
	if err := os.MkdirAll(realProfileDir, 0o755); err != nil {
		t.Fatalf("mkdir real profile dir: %v", err)
	}
	if _, err := os.Create(filepath.Join(realProfileDir, "docker.sock")); err != nil {
		t.Fatalf("create docker.sock: %v", err)
	}

	// Create a symlink inside .colima/ pointing to the real directory.
	symlinkPath := filepath.Join(colimaDir, "symlinked-profile")
	if err := os.Symlink(realProfileDir, symlinkPath); err != nil {
		t.Fatalf("os.Symlink: %v", err)
	}

	// Also create a legitimate non-symlink profile so we can verify real ones still work.
	legitimateDir := filepath.Join(colimaDir, "legitimate")
	if err := os.MkdirAll(legitimateDir, 0o755); err != nil {
		t.Fatalf("mkdir legitimate: %v", err)
	}
	if _, err := os.Create(filepath.Join(legitimateDir, "docker.sock")); err != nil {
		t.Fatalf("create docker.sock in legitimate: %v", err)
	}

	detector := &PlatformDetector{homeDir: tempHome}
	platforms := detector.detectAllColima()

	// The symlinked entry must be skipped; only "legitimate" should be returned.
	if len(platforms) != 1 {
		t.Errorf("detectAllColima() returned %d platforms, want 1 (symlink must be skipped)", len(platforms))
		for _, p := range platforms {
			t.Logf("  got: profile=%s socket=%s", p.Profile, p.SocketPath)
		}
		return
	}

	if platforms[0].Profile != "legitimate" {
		t.Errorf("detectAllColima() returned profile=%q, want %q", platforms[0].Profile, "legitimate")
	}
}

// TestPlatformDetector_DetectAllColima_InvalidProfileNames verifies that detectAllColima()
// skips directory entries whose names fail profile-name validation (e.g. path traversal,
// semicolons) and only returns entries with valid names.
func TestPlatformDetector_DetectAllColima_InvalidProfileNames(t *testing.T) {
	tempHome := t.TempDir()
	colimaDir := filepath.Join(tempHome, ".colima")
	if err := os.MkdirAll(colimaDir, 0o755); err != nil {
		t.Fatalf("mkdir .colima: %v", err)
	}

	// Create directories with names that isValidColimaProfileName should reject.
	// Note: some shell-hostile characters (e.g. "|") are illegal in directory names
	// on certain filesystems, so we only use names that are actually creatable.
	invalidNames := []string{
		"semi;colon",
		"has space",
	}
	for _, name := range invalidNames {
		dir := filepath.Join(colimaDir, name)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			// Some filesystems reject these names — skip rather than fail.
			t.Logf("skipping %q: os.MkdirAll failed: %v", name, err)
			continue
		}
		if _, err := os.Create(filepath.Join(dir, "docker.sock")); err != nil {
			t.Fatalf("create docker.sock in %q: %v", name, err)
		}
	}

	// Add one valid profile so we can confirm valid names still pass through.
	validDir := filepath.Join(colimaDir, "valid-profile")
	if err := os.MkdirAll(validDir, 0o755); err != nil {
		t.Fatalf("mkdir valid-profile: %v", err)
	}
	if _, err := os.Create(filepath.Join(validDir, "docker.sock")); err != nil {
		t.Fatalf("create docker.sock in valid-profile: %v", err)
	}

	detector := &PlatformDetector{homeDir: tempHome}
	platforms := detector.detectAllColima()

	// Only "valid-profile" should survive; all invalid-name dirs must be skipped.
	if len(platforms) != 1 {
		t.Errorf("detectAllColima() returned %d platforms, want 1 (invalid names must be skipped)", len(platforms))
		for _, p := range platforms {
			t.Logf("  got: profile=%s socket=%s", p.Profile, p.SocketPath)
		}
		return
	}

	if platforms[0].Profile != "valid-profile" {
		t.Errorf("detectAllColima() returned profile=%q, want %q", platforms[0].Profile, "valid-profile")
	}
}
