package operators

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

// PlatformType identifies which container platform is being used
type PlatformType string

const (
	PlatformOrbStack      PlatformType = "orbstack"
	PlatformColima        PlatformType = "colima"
	PlatformDockerDesktop PlatformType = "docker-desktop"
	PlatformPodman        PlatformType = "podman"
	PlatformLinuxNative   PlatformType = "linux-native"
	PlatformUnknown       PlatformType = "unknown"
)

// Platform contains information about a detected container platform
type Platform struct {
	Type       PlatformType
	SocketPath string
	Profile    string // For platforms that support profiles (e.g., Colima)
	Name       string // Human-readable name
	HomeDir    string // Home directory (for building paths)
}

// PlatformDetector detects available container platforms
type PlatformDetector struct {
	homeDir string
}

// NewPlatformDetector creates a new platform detector
func NewPlatformDetector() (*PlatformDetector, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	return &PlatformDetector{homeDir: homeDir}, nil
}

// Detect finds the container platform to use
// Priority: DVM_PLATFORM env var > config file > auto-detect
func (pd *PlatformDetector) Detect() (*Platform, error) {
	// Check for explicit platform selection via env var
	if platformEnv := os.Getenv("DVM_PLATFORM"); platformEnv != "" {
		return pd.detectSpecific(PlatformType(strings.ToLower(platformEnv)))
	}

	// Check config file
	if platformConfig := viper.GetString("runtime.platform"); platformConfig != "" && platformConfig != "auto" {
		return pd.detectSpecific(PlatformType(strings.ToLower(platformConfig)))
	}

	// Auto-detect: check platforms in order of preference
	return pd.autoDetect()
}

// detectSpecific tries to detect a specific platform
func (pd *PlatformDetector) detectSpecific(platformType PlatformType) (*Platform, error) {
	var platform *Platform

	switch platformType {
	case PlatformOrbStack:
		platform = pd.detectOrbStack()
	case PlatformColima:
		platform = pd.detectColima()
	case PlatformDockerDesktop, "docker":
		platform = pd.detectDockerDesktop()
	case PlatformPodman:
		platform = pd.detectPodman()
	case PlatformLinuxNative:
		platform = pd.detectLinuxNative()
	default:
		return nil, fmt.Errorf("unknown platform type: %s (valid: orbstack, colima, docker-desktop, podman)", platformType)
	}

	if platform == nil {
		return nil, fmt.Errorf("platform %s is not available. %s", platformType, pd.getInstallHint(platformType))
	}

	return platform, nil
}

// autoDetect finds the first available platform
func (pd *PlatformDetector) autoDetect() (*Platform, error) {
	detectors := []func() *Platform{
		pd.detectOrbStack,
		pd.detectColima,
		pd.detectDockerDesktop,
		pd.detectPodman,
		pd.detectLinuxNative,
	}

	for _, detect := range detectors {
		if platform := detect(); platform != nil {
			return platform, nil
		}
	}

	return nil, fmt.Errorf("no container runtime found. Please install one of: OrbStack, Colima, Docker Desktop, or Podman")
}

// DetectAll finds all available container platforms
func (pd *PlatformDetector) DetectAll() []*Platform {
	var platforms []*Platform

	detectors := []func() *Platform{
		pd.detectOrbStack,
		pd.detectColima,
		pd.detectDockerDesktop,
		pd.detectPodman,
		pd.detectLinuxNative,
	}

	for _, detect := range detectors {
		if platform := detect(); platform != nil {
			platforms = append(platforms, platform)
		}
	}

	return platforms
}

func (pd *PlatformDetector) detectOrbStack() *Platform {
	// OrbStack can expose socket at multiple locations
	socketPaths := []string{
		filepath.Join(pd.homeDir, ".orbstack", "run", "docker.sock"),
		"/var/run/docker.sock", // OrbStack can also symlink here when set as default
	}

	for _, socketPath := range socketPaths {
		if _, err := os.Stat(socketPath); err == nil {
			// Verify it's actually OrbStack by checking for OrbStack-specific indicators
			// OrbStack directory should exist
			orbstackDir := filepath.Join(pd.homeDir, ".orbstack")
			if _, err := os.Stat(orbstackDir); err == nil {
				// For /var/run/docker.sock, only claim it if it's a symlink to OrbStack
				// or if no other platform (Colima, Docker Desktop) is detected
				if socketPath == "/var/run/docker.sock" {
					// Check if socket is symlink to OrbStack
					if target, err := os.Readlink(socketPath); err == nil {
						if strings.Contains(target, "orbstack") {
							return &Platform{
								Type:       PlatformOrbStack,
								SocketPath: socketPath,
								Name:       "OrbStack",
								HomeDir:    pd.homeDir,
							}
						}
					}
					// Not a symlink to OrbStack, skip
					continue
				}

				return &Platform{
					Type:       PlatformOrbStack,
					SocketPath: socketPath,
					Name:       "OrbStack",
					HomeDir:    pd.homeDir,
				}
			}
		}
	}
	return nil
}

func (pd *PlatformDetector) detectColima() *Platform {
	// Get Colima profile from environment
	profile := os.Getenv("COLIMA_DOCKER_PROFILE")
	if profile == "" {
		profile = os.Getenv("COLIMA_ACTIVE_PROFILE")
	}
	if profile == "" {
		profile = "default"
	}

	// Check for docker socket first
	dockerSock := filepath.Join(pd.homeDir, ".colima", profile, "docker.sock")
	if _, err := os.Stat(dockerSock); err == nil {
		return &Platform{
			Type:       PlatformColima,
			SocketPath: dockerSock,
			Profile:    profile,
			Name:       fmt.Sprintf("Colima (profile: %s)", profile),
			HomeDir:    pd.homeDir,
		}
	}

	// Check for containerd socket
	containerdSock := filepath.Join(pd.homeDir, ".colima", profile, "containerd.sock")
	if _, err := os.Stat(containerdSock); err == nil {
		return &Platform{
			Type:       PlatformColima,
			SocketPath: containerdSock,
			Profile:    profile,
			Name:       fmt.Sprintf("Colima containerd (profile: %s)", profile),
			HomeDir:    pd.homeDir,
		}
	}

	return nil
}

func (pd *PlatformDetector) detectDockerDesktop() *Platform {
	// macOS Docker Desktop locations
	locations := []string{
		filepath.Join(pd.homeDir, ".docker", "run", "docker.sock"),
		"/var/run/docker.sock",
	}

	for _, socketPath := range locations {
		if _, err := os.Stat(socketPath); err == nil {
			// Verify it's actually Docker Desktop, not another runtime
			// by checking that other runtimes aren't detected
			if pd.detectOrbStack() == nil && pd.detectColima() == nil {
				return &Platform{
					Type:       PlatformDockerDesktop,
					SocketPath: socketPath,
					Name:       "Docker Desktop",
					HomeDir:    pd.homeDir,
				}
			}
		}
	}

	return nil
}

func (pd *PlatformDetector) detectPodman() *Platform {
	// Check for Podman socket - try multiple locations
	// macOS uses a temp directory socket, Linux uses /run/podman
	locations := []string{
		filepath.Join(pd.homeDir, ".local", "share", "containers", "podman", "machine", "podman.sock"),
		"/run/podman/podman.sock",
		"/run/user/1000/podman/podman.sock",
	}

	for _, socketPath := range locations {
		if _, err := os.Stat(socketPath); err == nil {
			return &Platform{
				Type:       PlatformPodman,
				SocketPath: socketPath,
				Name:       "Podman",
				HomeDir:    pd.homeDir,
			}
		}
	}

	// On macOS, try to find the Podman socket in temp directory
	// Pattern: /var/folders/.../T/podman/podman-machine-default-api.sock
	tmpDir := os.TempDir()
	podmanTmpSock := filepath.Join(filepath.Dir(tmpDir), "podman", "podman-machine-default-api.sock")
	if _, err := os.Stat(podmanTmpSock); err == nil {
		return &Platform{
			Type:       PlatformPodman,
			SocketPath: podmanTmpSock,
			Name:       "Podman",
			HomeDir:    pd.homeDir,
		}
	}

	// Try glob pattern for macOS temp directories
	matches, _ := filepath.Glob("/var/folders/*/*/T/podman/podman-machine-default-api.sock")
	if len(matches) > 0 {
		return &Platform{
			Type:       PlatformPodman,
			SocketPath: matches[0],
			Name:       "Podman",
			HomeDir:    pd.homeDir,
		}
	}

	return nil
}

func (pd *PlatformDetector) detectLinuxNative() *Platform {
	// Only on Linux, check for native Docker
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		// Make sure we're on Linux and it's not another platform
		if pd.detectOrbStack() == nil && pd.detectColima() == nil && pd.detectDockerDesktop() == nil {
			return &Platform{
				Type:       PlatformLinuxNative,
				SocketPath: "/var/run/docker.sock",
				Name:       "Docker (native)",
				HomeDir:    pd.homeDir,
			}
		}
	}

	return nil
}

// getInstallHint returns installation instructions for a platform
func (pd *PlatformDetector) getInstallHint(platformType PlatformType) string {
	switch platformType {
	case PlatformOrbStack:
		return "Install OrbStack from https://orbstack.dev or run: brew install orbstack"
	case PlatformColima:
		return "Install Colima with: brew install colima && colima start"
	case PlatformDockerDesktop:
		return "Install Docker Desktop from https://docker.com/products/docker-desktop"
	case PlatformPodman:
		return "Install Podman with: brew install podman && podman machine init && podman machine start"
	default:
		return "Please install a container runtime"
	}
}

// GetStartHint returns a helpful message for starting the platform
func (p *Platform) GetStartHint() string {
	switch p.Type {
	case PlatformOrbStack:
		return "Start OrbStack from the menu bar or run: open -a OrbStack"
	case PlatformColima:
		if p.Profile != "" && p.Profile != "default" {
			return fmt.Sprintf("Start Colima with: colima start --profile %s", p.Profile)
		}
		return "Start Colima with: colima start"
	case PlatformDockerDesktop:
		return "Start Docker Desktop from Applications"
	case PlatformPodman:
		return "Start Podman machine with: podman machine start"
	case PlatformLinuxNative:
		return "Start Docker daemon with: sudo systemctl start docker"
	default:
		return "Please start your container runtime"
	}
}

// IsContainerd returns true if this platform uses containerd (vs Docker API)
func (p *Platform) IsContainerd() bool {
	if p.Type == PlatformColima {
		return filepath.Base(p.SocketPath) == "containerd.sock"
	}
	return false
}

// IsDockerCompatible returns true if this platform supports the Docker API
func (p *Platform) IsDockerCompatible() bool {
	switch p.Type {
	case PlatformOrbStack, PlatformDockerDesktop, PlatformPodman, PlatformLinuxNative:
		return true
	case PlatformColima:
		return !p.IsContainerd()
	default:
		return false
	}
}

// GetBuildKitSocket returns the BuildKit socket path for this platform (if available)
func (p *Platform) GetBuildKitSocket() string {
	switch p.Type {
	case PlatformColima:
		if p.Profile != "" {
			return filepath.Join(p.HomeDir, ".colima", p.Profile, "buildkitd.sock")
		}
		return filepath.Join(p.HomeDir, ".colima", "default", "buildkitd.sock")
	default:
		// OrbStack, Docker Desktop, etc. use Docker's built-in buildkit
		return ""
	}
}

// GetContainerdSocket returns the containerd socket path for this platform (if available)
func (p *Platform) GetContainerdSocket() string {
	switch p.Type {
	case PlatformColima:
		if p.Profile != "" {
			return filepath.Join(p.HomeDir, ".colima", p.Profile, "containerd.sock")
		}
		return filepath.Join(p.HomeDir, ".colima", "default", "containerd.sock")
	default:
		return ""
	}
}
