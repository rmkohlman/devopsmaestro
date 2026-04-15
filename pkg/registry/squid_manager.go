package registry

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// HttpProxy defines the interface for managing an HTTP proxy (Squid).
// All implementations must be safe for concurrent use.
type HttpProxy interface {
	// Start starts the HTTP proxy process.
	// Returns an error if the proxy fails to start.
	Start(ctx context.Context) error

	// Stop stops the proxy process gracefully.
	// Sends SIGTERM, then SIGKILL if process doesn't exit within timeout.
	Stop(ctx context.Context) error

	// Status returns the current status of the proxy.
	Status(ctx context.Context) (*HttpProxyStatus, error)

	// EnsureRunning starts the proxy if it's not running.
	// Idempotent - does nothing if already running.
	EnsureRunning(ctx context.Context) error

	// IsRunning checks if the proxy is currently running.
	IsRunning(ctx context.Context) bool

	// GetEndpoint returns the proxy endpoint (e.g., "http://localhost:3128").
	GetEndpoint() string

	// GetProxyEnv returns the proxy environment variables.
	// Returns a map with keys like "HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY".
	GetProxyEnv() map[string]string
}

// SquidManager manages squid HTTP proxy lifecycle.
type SquidManager struct {
	BaseServiceManager
	config HttpProxyConfig
}

// NewSquidManager creates a new SquidManager with the given configuration.
func NewSquidManager(config HttpProxyConfig) *SquidManager {
	// Apply defaults for any unset fields
	config.ApplyDefaults()

	// Create real binary and process managers
	binaryManager := NewSquidBinaryManager()
	processManager := NewProcessManager(ProcessConfig{
		PIDFile: config.PidFile,
		LogFile: filepath.Join(config.LogDir, "squid.log"),
	})

	return &SquidManager{
		BaseServiceManager: NewBaseServiceManager(binaryManager, processManager),
		config:             config,
	}
}

// NewSquidManagerWithDeps creates a SquidManager with injected dependencies for testing.
func NewSquidManagerWithDeps(config HttpProxyConfig, binaryManager BinaryManager, processManager ProcessManager) *SquidManager {
	// Apply defaults for any unset fields
	config.ApplyDefaults()

	return &SquidManager{
		BaseServiceManager: NewBaseServiceManager(binaryManager, processManager),
		config:             config,
	}
}

// Start starts the squid proxy process.
func (m *SquidManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check context first
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// Check if already running.
	// If running, verify the on-disk config matches what we would generate.
	// A stale process (from a previous dvm version) might be listening on
	// 127.0.0.1 instead of 0.0.0.0 — we must restart it. (see #353)
	if m.processManager.IsRunning() {
		configPath := filepath.Join(m.config.LogDir, "squid.conf")
		if needsRestart, _ := m.configStale(configPath); needsRestart {
			// Stop the stale process so we can restart with the new config.
			m.processManager.Stop(ctx)
		} else {
			if m.startTime.IsZero() {
				m.RecordStartLocked()
			}
			return nil
		}
	}

	// Validate config
	if err := m.config.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Ensure binary exists
	binaryPath, err := m.binaryManager.EnsureBinary(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure binary: %w", err)
	}

	// Check if port is available
	if !IsPortAvailable(m.config.Port) {
		return fmt.Errorf("%w: port %d is already in use", ErrPortInUse, m.config.Port)
	}

	// Ensure directories exist
	if err := os.MkdirAll(m.config.CacheDir, 0700); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	if err := os.MkdirAll(m.config.LogDir, 0700); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Generate squid.conf file
	configPath := filepath.Join(m.config.LogDir, "squid.conf")
	if err := m.generateConfig(configPath); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Initialize cache directories with `squid -z -f <config>`
	if err := m.initializeCacheDir(ctx, binaryPath, configPath); err != nil {
		return fmt.Errorf("failed to initialize cache directories: %w", err)
	}

	// Prepare process config
	procConfig := ProcessConfig{
		PIDFile:         m.config.PidFile,
		LogFile:         filepath.Join(m.config.LogDir, "squid.log"),
		WorkingDir:      m.config.LogDir,
		ShutdownTimeout: 10 * time.Second,
	}

	// Start squid in foreground mode: `squid -N -f <config>`
	args := []string{
		"-N", // Foreground mode (no daemon)
		"-f", configPath,
	}

	if err := m.processManager.Start(ctx, binaryPath, args, procConfig); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	// Record start time
	m.RecordStartLocked()

	// Wait for proxy to be ready
	if err := m.waitForReady(ctx); err != nil {
		m.processManager.Stop(ctx)
		return fmt.Errorf("proxy failed to become ready: %w", err)
	}

	return nil
}

// Stop stops the squid proxy process gracefully.
func (m *SquidManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If not running, this is idempotent
	if !m.processManager.IsRunning() {
		return nil
	}

	// Stop process
	return m.processManager.Stop(ctx)
}

// Status returns the current status of the squid proxy.
func (m *SquidManager) Status(ctx context.Context) (*HttpProxyStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	running := m.processManager.IsRunning()
	state := "stopped"
	if running {
		state = "running"
	}

	var uptime time.Duration
	if running && !m.startTime.IsZero() {
		uptime = time.Since(m.startTime)
	}

	// Get version
	version, _ := m.binaryManager.GetVersion(ctx)

	return &HttpProxyStatus{
		State:    state,
		PID:      m.processManager.GetPID(),
		Port:     m.config.Port,
		CacheDir: m.config.CacheDir,
		Version:  version,
		Uptime:   uptime,
	}, nil
}

// EnsureRunning starts the proxy if it's not running.
func (m *SquidManager) EnsureRunning(ctx context.Context) error {
	if m.IsRunning(ctx) {
		return nil
	}
	return m.Start(ctx)
}

// IsRunning checks if the squid proxy is currently running.
func (m *SquidManager) IsRunning(ctx context.Context) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processManager.IsRunning()
}

// GetEndpoint returns the squid proxy endpoint.
func (m *SquidManager) GetEndpoint() string {
	return fmt.Sprintf("http://localhost:%d", m.config.Port)
}

// GetProxyEnv returns the proxy environment variables to use this proxy.
func (m *SquidManager) GetProxyEnv() map[string]string {
	endpoint := m.GetEndpoint()

	return map[string]string{
		"HTTP_PROXY":  endpoint,
		"HTTPS_PROXY": endpoint,
		"http_proxy":  endpoint,
		"https_proxy": endpoint,
		"NO_PROXY":    "localhost,127.0.0.1",
		"no_proxy":    "localhost,127.0.0.1",
	}
}

// generateConfig generates the squid.conf file and writes it to the specified path.
func (m *SquidManager) generateConfig(configPath string) error {
	configContent, err := GenerateSquidConfig(m.config)
	if err != nil {
		return err
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Write config file — use restricted permissions since config may contain sensitive data
	if err := os.WriteFile(configPath, []byte(configContent), configFileMode); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// configStale returns true when the on-disk squid.conf differs from what
// GenerateSquidConfig would produce for the current HttpProxyConfig.
// This detects processes started by older dvm versions that bound
// 127.0.0.1 instead of 0.0.0.0 (see #353).
func (m *SquidManager) configStale(configPath string) (bool, error) {
	existing, err := os.ReadFile(configPath)
	if err != nil {
		// Config file missing or unreadable — treat as stale so we
		// regenerate it before (re)starting squid.
		return true, err
	}

	desired, err := GenerateSquidConfig(m.config)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(existing)) != strings.TrimSpace(desired), nil
}

// initializeCacheDir initializes squid cache directories with `squid -z`.
func (m *SquidManager) initializeCacheDir(ctx context.Context, binaryPath, configPath string) error {
	// Run squid -z -f <config> to initialize cache directories.
	// This creates the swap directories for the disk cache.
	cmd := exec.CommandContext(ctx, binaryPath, "-z", "-N", "-f", configPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		// If Squid reports it's already running, the cache dirs are already
		// initialized from the previous run — safe to proceed.
		if strings.Contains(outputStr, "already running") {
			return nil
		}
		return fmt.Errorf("squid -z failed: %w (output: %s)", err, outputStr)
	}
	return nil
}

// waitForReady waits for the proxy to become ready.
func (m *SquidManager) waitForReady(ctx context.Context) error {
	address := fmt.Sprintf("localhost:%d", m.config.Port)
	return WaitForReadyTCP(ctx, address, 10*time.Second)
}

// GenerateSquidConfig generates a squid configuration from an HttpProxyConfig.
// Returns the squid.conf content as a string.
func GenerateSquidConfig(cfg HttpProxyConfig) (string, error) {
	// Validate config first
	if err := cfg.Validate(); err != nil {
		return "", fmt.Errorf("invalid configuration: %w", err)
	}

	// Build squid.conf content
	//
	// Listening on 0.0.0.0 is required because BuildKit containers inside Colima
	// access the proxy via host.docker.internal, which resolves to the VM's gateway
	// IP (e.g. 192.168.5.2) — NOT 127.0.0.1. If we bind to loopback only,
	// containers get "connection refused" on the gateway address (see #346).
	//
	// Security: the ACL rules below restrict access to RFC1918 private subnets
	// and localhost. The host firewall (macOS) blocks external access to the
	// Colima VM, so binding to 0.0.0.0 inside the VM is safe.
	config := fmt.Sprintf(`# Generated by DevOpsMaestro
# Squid HTTP proxy configuration

# Listen on all interfaces so BuildKit containers can connect via
# host.docker.internal (resolves to VM gateway, not 127.0.0.1).
# ACLs below restrict access to private subnets only. (see #346, #353)
# IMPORTANT: We explicitly bind 0.0.0.0 rather than using bare port form
# because bare port behaviour varies across squid versions and platforms.
http_port 0.0.0.0:%d

# Cache settings
# Squid 7.x changed internal hash-table allocation for ufs stores.
# L1=16 L2=256 caused an xcalloc integer overflow (UINT64_MAX blocks,
# see #363). Using L1=16 L2=64 avoids the overflow while still
# providing 1024 second-level directories for cache distribution.
cache_dir ufs %s %d 16 64
maximum_object_size %d MB
cache_mem %d MB

# Logging
access_log %s/access.log squid
cache_log %s/cache.log
pid_filename %s

# ACLs — restrict to localhost and RFC1918 private subnets
acl localnet src 127.0.0.0/8
acl localnet src 10.0.0.0/8
acl localnet src 172.16.0.0/12
acl localnet src 192.168.0.0/16
acl Safe_ports port 80 443 21 1025-65535
acl CONNECT method CONNECT
acl SSL_ports port 443

# Access rules
http_access deny !Safe_ports
http_access deny CONNECT !SSL_ports
http_access allow localhost
http_access allow localnet
http_access deny all

# Privacy/security settings
forwarded_for delete
visible_hostname dvm-squid-proxy

# Graceful shutdown
shutdown_lifetime 3 seconds
`,
		cfg.Port,
		cfg.CacheDir, cfg.CacheSizeMB,
		cfg.MaxObjectSizeMB,
		cfg.MemoryCacheMB,
		cfg.LogDir,
		cfg.LogDir,
		cfg.PidFile,
	)

	return config, nil
}

// =============================================================================
// SquidManagerAdapter - ServiceManager interface adapter
// =============================================================================

// SquidManagerAdapter wraps SquidManager for ServiceManager interface.
type SquidManagerAdapter struct {
	manager *SquidManager
}

// NewSquidManagerAdapter creates a new SquidManagerAdapter.
func NewSquidManagerAdapter(manager *SquidManager) *SquidManagerAdapter {
	return &SquidManagerAdapter{manager: manager}
}

// Start starts the squid proxy.
func (a *SquidManagerAdapter) Start(ctx context.Context) error {
	return a.manager.Start(ctx)
}

// Stop stops the squid proxy.
func (a *SquidManagerAdapter) Stop(ctx context.Context) error {
	return a.manager.Stop(ctx)
}

// IsRunning checks if squid is running.
func (a *SquidManagerAdapter) IsRunning(ctx context.Context) bool {
	return a.manager.IsRunning(ctx)
}

// GetEndpoint returns the squid endpoint (full URL).
func (a *SquidManagerAdapter) GetEndpoint() string {
	return a.manager.GetEndpoint()
}

// =============================================================================
// Interface Compliance
// =============================================================================

// Verify SquidManager implements HttpProxy interface
var _ HttpProxy = (*SquidManager)(nil)

// Verify SquidManagerAdapter implements ServiceManager interface
var _ ServiceManager = (*SquidManagerAdapter)(nil)
