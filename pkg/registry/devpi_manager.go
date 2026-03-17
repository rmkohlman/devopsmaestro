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

// DevpiManager manages a devpi PyPI proxy instance.
type DevpiManager struct {
	BaseServiceManager
	config      PyPIProxyConfig
	initialized bool
}

// NewDevpiManager creates a DevpiManager with injected dependencies.
// This is the canonical constructor — callers provide their own BinaryManager
// and ProcessManager, enabling testability and explicit dependency control.
// Returns an error if the config is invalid.
func NewDevpiManager(config PyPIProxyConfig, binary BinaryManager, process ProcessManager) (*DevpiManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid devpi config: %w", err)
	}

	return &DevpiManager{
		BaseServiceManager: NewBaseServiceManager(binary, process),
		config:             config,
	}, nil
}

// NewDevpiManagerDefault creates a DevpiManager with default production
// dependencies (PipxBinaryManager for devpi-server v6.2.0, ProcessManager
// with standard PID/log paths). Prefer NewDevpiManager for testability.
func NewDevpiManagerDefault(config PyPIProxyConfig) (*DevpiManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid devpi config: %w", err)
	}

	binaryManager := NewPipxBinaryManager("devpi-server", "6.2.0")
	processManager := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(config.Storage, "devpi.pid"),
		LogFile: filepath.Join(config.Storage, "devpi.log"),
	})

	return &DevpiManager{
		BaseServiceManager: NewBaseServiceManager(binaryManager, processManager),
		config:             config,
	}, nil
}

// Start starts the devpi proxy process.
func (m *DevpiManager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already running - idempotent
	if m.processManager.IsRunning() {
		return nil
	}

	// Ensure binary exists
	binaryPath, err := m.binaryManager.EnsureBinary(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure binary: %w", err)
	}

	// Check if port is available
	if !IsPortAvailable(m.config.Port) {
		// Port is taken — check if our service is already running on it
		if ProbeServiceHealth(m.config.Port, "/", []int{200, 302}) {
			return nil // Adopt running instance (idempotent)
		}
		return fmt.Errorf("%w: port %d is already in use", ErrPortInUse, m.config.Port)
	}

	// Determine server directory
	serverDir := m.config.ServerDir
	if serverDir == "" {
		serverDir = filepath.Join(m.config.Storage, "server")
	}

	// Initialize devpi server if not already initialized
	if !m.initialized {
		if err := m.initializeDevpiServer(ctx, serverDir); err != nil {
			return fmt.Errorf("failed to initialize devpi server: %w", err)
		}
		m.initialized = true
	}

	// Prepare process config
	procConfig := ProcessConfig{
		PIDFile:         filepath.Join(m.config.Storage, "devpi.pid"),
		LogFile:         filepath.Join(m.config.Storage, "devpi.log"),
		WorkingDir:      m.config.Storage,
		ShutdownTimeout: 10 * time.Second,
	}

	// Build devpi-server arguments
	args := []string{
		"--serverdir", serverDir,
		"--port", fmt.Sprintf("%d", m.config.Port),
		"--host", "127.0.0.1",
	}

	// Start devpi process
	if err := m.processManager.Start(ctx, binaryPath, args, procConfig); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	// Record start time
	m.RecordStartLocked()

	// Setup idle timer if on-demand mode
	m.ResetIdleTimerLocked(m.config.Lifecycle, m.config.IdleTimeout, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		m.Stop(ctx)
	})

	// Wait for proxy to be ready
	if err := m.waitForReady(ctx); err != nil {
		m.processManager.Stop(ctx)
		return fmt.Errorf("proxy failed to become ready: %w", err)
	}

	return nil
}

// Stop stops the devpi proxy process gracefully.
func (m *DevpiManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop idle timer if running
	m.StopIdleTimerLocked()

	// Stop process (idempotent)
	return m.processManager.Stop(ctx)
}

// Status returns the current status of the devpi proxy.
func (m *DevpiManager) Status(ctx context.Context) (*PyPIProxyStatus, error) {
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

	// Get package count and disk usage
	packageCount := 0
	var diskUsage int64

	if running {
		// Count packages in storage directory
		packageCount, diskUsage = m.getProxyStats(ctx)
	}

	return &PyPIProxyStatus{
		State:        state,
		PID:          m.processManager.GetPID(),
		Port:         m.config.Port,
		Storage:      m.config.Storage,
		Version:      version,
		Uptime:       uptime,
		PackageCount: packageCount,
		DiskUsage:    diskUsage,
	}, nil
}

// EnsureRunning starts the proxy if it's not running.
func (m *DevpiManager) EnsureRunning(ctx context.Context) error {
	if m.IsRunning(ctx) {
		return nil
	}
	return m.Start(ctx)
}

// IsRunning checks if the devpi proxy is currently running.
// Falls back to a health probe when no PID file exists (e.g. adopted instances).
func (m *DevpiManager) IsRunning(ctx context.Context) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.processManager.IsRunning() {
		return true
	}
	return ProbeServiceHealth(m.config.Port, "/", []int{200, 302})
}

// GetEndpoint returns the devpi proxy endpoint.
func (m *DevpiManager) GetEndpoint() string {
	// Reset idle timer on access
	if m.config.Lifecycle == "on-demand" && m.config.IdleTimeout > 0 {
		m.ResetIdleTimer(m.config.Lifecycle, m.config.IdleTimeout, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			m.Stop(ctx)
		})
	}
	return fmt.Sprintf("http://localhost:%d", m.config.Port)
}

// GetPipEnv returns the pip environment variables to use this proxy.
func (m *DevpiManager) GetPipEnv() map[string]string {
	endpoint := m.GetEndpoint()

	// Devpi uses the root/pypi index by default
	indexURL := fmt.Sprintf("%s/root/pypi/+simple/", endpoint)

	return map[string]string{
		"PIP_INDEX_URL":    indexURL,
		"PIP_TRUSTED_HOST": "localhost", // Trust localhost for local development
	}
}

// GetPipConfig returns the pip configuration in pip.conf format.
func (m *DevpiManager) GetPipConfig() string {
	endpoint := m.GetEndpoint()
	indexURL := fmt.Sprintf("%s/root/pypi/+simple/", endpoint)

	return fmt.Sprintf(`[global]
index-url = %s
trusted-host = localhost
`, indexURL)
}

// initializeDevpiServer runs devpi-init if the server hasn't been initialized.
func (m *DevpiManager) initializeDevpiServer(ctx context.Context, serverDir string) error {
	// Check if already initialized
	if _, err := os.Stat(filepath.Join(serverDir, ".serverversion")); err == nil {
		return nil // Already initialized
	}

	// Ensure binary exists
	binaryPath, err := m.binaryManager.EnsureBinary(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure binary: %w", err)
	}

	// Get the devpi-init binary path (should be in same directory)
	initBinary := strings.Replace(binaryPath, "devpi-server", "devpi-init", 1)

	// Ensure directory exists
	if err := os.MkdirAll(serverDir, 0755); err != nil {
		return fmt.Errorf("failed to create server directory: %w", err)
	}

	// Run devpi-init
	cmd := exec.CommandContext(ctx, initBinary, "--serverdir", serverDir)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("devpi-init failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// waitForReady waits for the proxy to become ready.
func (m *DevpiManager) waitForReady(ctx context.Context) error {
	endpoint := fmt.Sprintf("http://localhost:%d/", m.config.Port)
	return WaitForReady(ctx, endpoint, []int{200, 302}, 10*time.Second)
}

// getProxyStats counts packages and disk usage.
func (m *DevpiManager) getProxyStats(ctx context.Context) (int, int64) {
	// Devpi stores packages in a directory structure
	// Count directories/files in storage path
	packageCount := 0
	var diskUsage int64

	serverDir := m.config.ServerDir
	if serverDir == "" {
		serverDir = filepath.Join(m.config.Storage, "server")
	}

	// Walk the storage directory
	filepath.Walk(serverDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			diskUsage += info.Size()
			// Count .tar.gz or .whl files as packages
			if strings.HasSuffix(path, ".tar.gz") || strings.HasSuffix(path, ".whl") {
				packageCount++
			}
		}
		return nil
	})

	return packageCount, diskUsage
}
