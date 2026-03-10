package operators

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// validProfileName matches safe Colima profile names: alphanumeric start,
// then alphanumeric, dots, hyphens, or underscores. No path traversal.
var validProfileName = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// isValidColimaProfileName returns true if name is a safe Colima profile name.
func isValidColimaProfileName(name string) bool {
	return validProfileName.MatchString(name) && !strings.Contains(name, "..")
}

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

// PlatformDetector defines the interface for detecting available container platforms.
// Implementations detect installed runtimes (OrbStack, Colima, Docker Desktop, Podman)
// by checking for platform-specific socket files and configuration directories.
type PlatformDetector interface {
	// Detect finds the best container platform to use.
	// Priority: DVM_PLATFORM env var > config file > auto-detect.
	Detect() (*Platform, error)

	// DetectAll finds all installed container platforms (socket files on disk).
	// Returns platforms regardless of whether they are currently running.
	DetectAll() []*Platform

	// DetectReachable finds all container platforms that are actively listening.
	// Filters DetectAll() to only platforms where the socket is reachable.
	DetectReachable() []*Platform
}

// Compile-time interface compliance check
var _ PlatformDetector = (*DefaultPlatformDetector)(nil)

// DefaultPlatformDetector is the standard implementation of PlatformDetector
// that detects platforms by inspecting the local filesystem for socket files.
type DefaultPlatformDetector struct {
	homeDir string
}

// NewPlatformDetector creates a new platform detector.
// Returns the PlatformDetector interface for loose coupling.
func NewPlatformDetector() (PlatformDetector, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	return &DefaultPlatformDetector{homeDir: homeDir}, nil
}

// Detect finds the container platform to use
// Priority: DVM_PLATFORM env var > config file > auto-detect
func (pd *DefaultPlatformDetector) Detect() (*Platform, error) {
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
func (pd *DefaultPlatformDetector) detectSpecific(platformType PlatformType) (*Platform, error) {
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
func (pd *DefaultPlatformDetector) autoDetect() (*Platform, error) {
	// Check non-Colima detectors and Colima separately so we can
	// enumerate all Colima profiles while keeping other detectors as-is.

	// OrbStack first (highest priority)
	if platform := pd.detectOrbStack(); platform != nil && platform.IsReachable() {
		return platform, nil
	}

	// All Colima profiles
	for _, platform := range pd.detectAllColima() {
		if platform.IsReachable() {
			return platform, nil
		}
	}

	// Remaining detectors in priority order
	remainingDetectors := []func() *Platform{
		pd.detectDockerDesktop,
		pd.detectPodman,
		pd.detectLinuxNative,
	}

	for _, detect := range remainingDetectors {
		if platform := detect(); platform != nil && platform.IsReachable() {
			return platform, nil
		}
	}

	return nil, fmt.Errorf("no container runtime found. Please install one of: OrbStack, Colima, Docker Desktop, or Podman")
}

// DetectAll finds all installed container platforms (socket files on disk).
// This returns platforms regardless of whether they are currently running.
// Use DetectReachable() to get only platforms that are actively listening.
func (pd *DefaultPlatformDetector) DetectAll() []*Platform {
	var platforms []*Platform

	// OrbStack
	if platform := pd.detectOrbStack(); platform != nil {
		platforms = append(platforms, platform)
	}

	// All Colima profiles
	platforms = append(platforms, pd.detectAllColima()...)

	// Remaining single-platform detectors
	remainingDetectors := []func() *Platform{
		pd.detectDockerDesktop,
		pd.detectPodman,
		pd.detectLinuxNative,
	}

	for _, detect := range remainingDetectors {
		if platform := detect(); platform != nil {
			platforms = append(platforms, platform)
		}
	}

	return platforms
}

// DetectReachable finds all container platforms that are actively listening.
// This calls DetectAll() and then filters to only platforms where the socket
// is reachable, i.e., the runtime is currently running.
func (pd *DefaultPlatformDetector) DetectReachable() []*Platform {
	all := pd.DetectAll()
	reachable := make([]*Platform, 0, len(all))
	for _, p := range all {
		if p.IsReachable() {
			reachable = append(reachable, p)
		}
	}
	return reachable
}

func (pd *DefaultPlatformDetector) detectOrbStack() *Platform {
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

func (pd *DefaultPlatformDetector) detectColima() *Platform {
	platforms := pd.detectAllColima()
	if len(platforms) > 0 {
		return platforms[0]
	}
	return nil
}

// detectAllColima enumerates all Colima profiles and returns platforms for each.
// If COLIMA_DOCKER_PROFILE or COLIMA_ACTIVE_PROFILE is set, only that profile is checked.
// Otherwise, all profile directories under ~/.colima/ are scanned.
func (pd *DefaultPlatformDetector) detectAllColima() []*Platform {
	// If an env var is set, only check that specific profile
	if profile := os.Getenv("COLIMA_DOCKER_PROFILE"); profile != "" {
		if !isValidColimaProfileName(profile) {
			// Invalid profile name in env var — fall through to enumeration
		} else {
			return pd.detectColimaProfile(profile)
		}
	}
	if profile := os.Getenv("COLIMA_ACTIVE_PROFILE"); profile != "" {
		if !isValidColimaProfileName(profile) {
			// Invalid profile name in env var — fall through to enumeration
		} else {
			return pd.detectColimaProfile(profile)
		}
	}

	// Enumerate all profile directories under ~/.colima/
	colimaDir := filepath.Join(pd.homeDir, ".colima")
	entries, err := os.ReadDir(colimaDir)
	if err != nil {
		return nil
	}

	var platforms []*Platform
	for _, entry := range entries {
		// Skip non-directories and symlinks (explicit symlink check for safety)
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			continue
		}
		name := entry.Name()
		// Skip internal directories (prefixed with _) and hidden directories (prefixed with .)
		if strings.HasPrefix(name, "_") || strings.HasPrefix(name, ".") {
			continue
		}
		// Validate profile name to prevent path traversal or injection
		if !isValidColimaProfileName(name) {
			continue
		}
		platforms = append(platforms, pd.detectColimaProfile(name)...)
	}

	return platforms
}

// detectColimaProfile checks a single Colima profile for docker.sock and containerd.sock.
// If both exist, docker.sock is preferred (matches existing behavior).
func (pd *DefaultPlatformDetector) detectColimaProfile(profile string) []*Platform {
	profileDir := filepath.Join(pd.homeDir, ".colima", profile)

	// Check for docker socket first (preferred)
	dockerSock := filepath.Join(profileDir, "docker.sock")
	if _, err := os.Stat(dockerSock); err == nil {
		return []*Platform{{
			Type:       PlatformColima,
			SocketPath: dockerSock,
			Profile:    profile,
			Name:       fmt.Sprintf("Colima (profile: %s)", profile),
			HomeDir:    pd.homeDir,
		}}
	}

	// Check for containerd socket
	containerdSock := filepath.Join(profileDir, "containerd.sock")
	if _, err := os.Stat(containerdSock); err == nil {
		return []*Platform{{
			Type:       PlatformColima,
			SocketPath: containerdSock,
			Profile:    profile,
			Name:       fmt.Sprintf("Colima containerd (profile: %s)", profile),
			HomeDir:    pd.homeDir,
		}}
	}

	return nil
}

func (pd *DefaultPlatformDetector) detectDockerDesktop() *Platform {
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

func (pd *DefaultPlatformDetector) detectPodman() *Platform {
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

func (pd *DefaultPlatformDetector) detectLinuxNative() *Platform {
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
func (pd *DefaultPlatformDetector) getInstallHint(platformType PlatformType) string {
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

// IsReachable returns true if the platform's socket is actually listening.
// This is a lightweight dial check — no Docker or containerd API calls are made.
func (p *Platform) IsReachable() bool {
	conn, err := net.DialTimeout("unix", p.SocketPath, 150*time.Millisecond)
	if err != nil {
		return false
	}
	conn.Close()
	return true
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
