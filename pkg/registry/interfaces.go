package registry

import (
	"context"
	"time"
)

// RegistryManager defines the interface for managing a container registry.
// All implementations must be safe for concurrent use.
type RegistryManager interface {
	// Start starts the registry process.
	// Returns an error if the registry fails to start.
	Start(ctx context.Context) error

	// Stop stops the registry process gracefully.
	// Sends SIGTERM, then SIGKILL if process doesn't exit within timeout.
	Stop(ctx context.Context) error

	// Status returns the current status of the registry.
	Status(ctx context.Context) (*RegistryStatus, error)

	// EnsureRunning starts the registry if it's not running.
	// Idempotent - does nothing if already running.
	EnsureRunning(ctx context.Context) error

	// IsRunning checks if the registry is currently running.
	IsRunning(ctx context.Context) bool

	// GetEndpoint returns the registry endpoint (e.g., "localhost:5001").
	GetEndpoint() string

	// Prune removes unused images from the registry.
	Prune(ctx context.Context, opts PruneOptions) (*PruneResult, error)
}

// ProcessManager handles OS process lifecycle management.
type ProcessManager interface {
	// Start spawns a new process with the given binary and arguments.
	Start(ctx context.Context, binary string, args []string, config ProcessConfig) error

	// Stop stops the process gracefully (SIGTERM then SIGKILL).
	Stop(ctx context.Context) error

	// IsRunning checks if the process is currently running.
	IsRunning() bool

	// GetPID returns the process ID, or 0 if not running.
	GetPID() int
}

// BinaryManager handles downloading and version management of binaries.
type BinaryManager interface {
	// EnsureBinary ensures the binary exists, downloading if necessary.
	// Returns the path to the binary.
	EnsureBinary(ctx context.Context) (string, error)

	// GetVersion returns the version of the currently installed binary.
	GetVersion(ctx context.Context) (string, error)

	// NeedsUpdate checks if the binary needs to be updated to the desired version.
	NeedsUpdate(ctx context.Context) (bool, error)

	// Update downloads and installs the latest version of the binary.
	// Backs up old version and rolls back on failure.
	Update(ctx context.Context) error
}

// RegistryConfig holds configuration for the registry.
type RegistryConfig struct {
	// Enabled determines if the registry should be used
	Enabled bool `yaml:"enabled"`

	// Lifecycle determines when the registry runs: "persistent", "on-demand", or "manual"
	Lifecycle string `yaml:"lifecycle"`

	// Port is the port the registry listens on
	Port int `yaml:"port"`

	// Storage is the directory where registry data is stored
	Storage string `yaml:"storage"`

	// IdleTimeout is how long to wait before shutting down an idle registry (on-demand mode)
	IdleTimeout time.Duration `yaml:"idleTimeout"`

	// Mirrors are the upstream registries to mirror/cache
	Mirrors []MirrorConfig `yaml:"mirrors"`
}

// MirrorConfig configures an upstream registry to mirror.
type MirrorConfig struct {
	// Name is a friendly name for this mirror
	Name string `yaml:"name"`

	// URL is the upstream registry URL (e.g., "https://index.docker.io")
	URL string `yaml:"url"`

	// OnDemand enables pull-through caching (download on first pull)
	OnDemand bool `yaml:"onDemand"`

	// Prefix is the registry prefix (e.g., "docker.io")
	Prefix string `yaml:"prefix"`
}

// RegistryStatus represents the current state of the registry.
type RegistryStatus struct {
	// State is "running" or "stopped"
	State string

	// PID is the process ID (0 if stopped)
	PID int

	// Port is the configured port
	Port int

	// Storage is the storage directory path
	Storage string

	// Version is the Zot version
	Version string

	// Uptime is how long the registry has been running
	Uptime time.Duration

	// ImageCount is the number of images stored
	ImageCount int

	// DiskUsage is the total disk space used (bytes)
	DiskUsage int64
}

// PruneOptions controls what gets pruned from the registry.
type PruneOptions struct {
	// All removes all unused images
	All bool

	// OlderThan removes images older than this duration
	OlderThan time.Duration

	// DryRun reports what would be removed without actually removing
	DryRun bool
}

// PruneResult contains the results of a prune operation.
type PruneResult struct {
	// ImagesRemoved is the number of images removed
	ImagesRemoved int

	// SpaceReclaimed is the disk space freed (bytes)
	SpaceReclaimed int64

	// Images is the list of images that were (or would be) removed
	Images []string
}

// ProcessConfig configures process management behavior.
type ProcessConfig struct {
	// PIDFile is where to write the process ID
	PIDFile string

	// LogFile is where to write process output
	LogFile string

	// WorkingDir is the working directory for the process
	WorkingDir string

	// ShutdownTimeout is how long to wait for graceful shutdown before SIGKILL
	ShutdownTimeout time.Duration
}

// Validate checks if the registry configuration is valid.
func (c *RegistryConfig) Validate() error {
	// Port must be in valid range and not a reserved/privileged port
	if c.Port < 1024 || c.Port > 65535 {
		return ErrInvalidConfig
	}
	if c.Storage == "" {
		return ErrInvalidConfig
	}
	if c.Lifecycle != "" && c.Lifecycle != "persistent" && c.Lifecycle != "on-demand" && c.Lifecycle != "manual" {
		return ErrInvalidConfig
	}

	// Validate mirror URLs
	for _, mirror := range c.Mirrors {
		if mirror.URL == "" {
			return ErrInvalidConfig
		}
		// Simple URL validation - check for http:// or https://
		if len(mirror.URL) < 8 || (mirror.URL[:7] != "http://" && mirror.URL[:8] != "https://") {
			return ErrInvalidConfig
		}
	}

	return nil
}

// GoModuleProxy defines the interface for managing a Go module proxy (Athens).
// All implementations must be safe for concurrent use.
type GoModuleProxy interface {
	// Start starts the Go module proxy process.
	// Returns an error if the proxy fails to start.
	Start(ctx context.Context) error

	// Stop stops the proxy process gracefully.
	// Sends SIGTERM, then SIGKILL if process doesn't exit within timeout.
	Stop(ctx context.Context) error

	// Status returns the current status of the proxy.
	Status(ctx context.Context) (*GoModuleProxyStatus, error)

	// EnsureRunning starts the proxy if it's not running.
	// Idempotent - does nothing if already running.
	EnsureRunning(ctx context.Context) error

	// IsRunning checks if the proxy is currently running.
	IsRunning(ctx context.Context) bool

	// GetEndpoint returns the proxy endpoint (e.g., "http://localhost:3000").
	GetEndpoint() string

	// GetGoEnv returns the Go environment variables to use this proxy.
	// Returns a map with keys like "GOPROXY", "GONOSUMDB", "GOPRIVATE".
	GetGoEnv() map[string]string
}

// GoModuleProxyStatus represents the current state of the Go module proxy.
type GoModuleProxyStatus struct {
	// State is "running" or "stopped"
	State string

	// PID is the process ID (0 if stopped)
	PID int

	// Port is the configured port
	Port int

	// Storage is the storage directory path
	Storage string

	// Version is the Athens version
	Version string

	// Uptime is how long the proxy has been running
	Uptime time.Duration

	// ModuleCount is the number of modules cached
	ModuleCount int

	// DiskUsage is the total disk space used (bytes)
	DiskUsage int64
}
