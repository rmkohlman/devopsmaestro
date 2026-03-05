package registry

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// VerdaccioManager manages a verdaccio npm proxy instance.
type VerdaccioManager struct {
	config         NpmProxyConfig
	binaryManager  BinaryManager
	processManager ProcessManager

	mu             sync.RWMutex
	startTime      time.Time
	idleTimer      *time.Timer
	lastAccessTime time.Time
}

// NewVerdaccioManager creates a new VerdaccioManager with the given configuration.
func NewVerdaccioManager(config NpmProxyConfig) *VerdaccioManager {
	// Validate config
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid verdaccio config: %v", err))
	}

	// Create real binary and process managers
	binaryManager := NewNpmBinaryManager("verdaccio", "5.28.0")
	processManager := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(config.Storage, "verdaccio.pid"),
		LogFile: filepath.Join(config.Storage, "verdaccio.log"),
	})

	return &VerdaccioManager{
		config:         config,
		binaryManager:  binaryManager,
		processManager: processManager,
	}
}

// NewVerdaccioManagerWithDeps creates a VerdaccioManager with injected dependencies.
func NewVerdaccioManagerWithDeps(config NpmProxyConfig, binary BinaryManager, process ProcessManager) *VerdaccioManager {
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid verdaccio config: %v", err))
	}

	return &VerdaccioManager{
		config:         config,
		binaryManager:  binary,
		processManager: process,
	}
}

// Start starts the verdaccio proxy process.
func (m *VerdaccioManager) Start(ctx context.Context) error {
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
	if !m.isPortAvailable(m.config.Port) {
		return fmt.Errorf("%w: port %d is already in use", ErrPortInUse, m.config.Port)
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(m.config.Storage, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Generate verdaccio config file
	if err := m.generateConfig(); err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Prepare process config
	procConfig := ProcessConfig{
		PIDFile:         filepath.Join(m.config.Storage, "verdaccio.pid"),
		LogFile:         filepath.Join(m.config.Storage, "verdaccio.log"),
		WorkingDir:      m.config.Storage,
		ShutdownTimeout: 10 * time.Second,
	}

	// Build verdaccio arguments
	configFile := filepath.Join(m.config.Storage, "config.yaml")
	args := []string{
		"--config", configFile,
		"--listen", fmt.Sprintf("127.0.0.1:%d", m.config.Port),
	}

	// Start verdaccio process
	if err := m.processManager.Start(ctx, binaryPath, args, procConfig); err != nil {
		return fmt.Errorf("failed to start proxy: %w", err)
	}

	// Record start time
	m.startTime = time.Now()
	m.lastAccessTime = m.startTime

	// Setup idle timer if on-demand mode
	if m.config.Lifecycle == "on-demand" && m.config.IdleTimeout > 0 {
		m.resetIdleTimer()
	}

	// Wait for proxy to be ready
	if err := m.waitForReady(ctx); err != nil {
		m.processManager.Stop(ctx)
		return fmt.Errorf("proxy failed to become ready: %w", err)
	}

	return nil
}

// Stop stops the verdaccio proxy process gracefully.
func (m *VerdaccioManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop idle timer if running
	if m.idleTimer != nil {
		m.idleTimer.Stop()
		m.idleTimer = nil
	}

	// Stop process (idempotent)
	return m.processManager.Stop(ctx)
}

// Status returns the current status of the verdaccio proxy.
func (m *VerdaccioManager) Status(ctx context.Context) (*NpmProxyStatus, error) {
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

	return &NpmProxyStatus{
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
func (m *VerdaccioManager) EnsureRunning(ctx context.Context) error {
	if m.IsRunning(ctx) {
		return nil
	}
	return m.Start(ctx)
}

// IsRunning checks if the verdaccio proxy is currently running.
func (m *VerdaccioManager) IsRunning(ctx context.Context) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processManager.IsRunning()
}

// GetEndpoint returns the verdaccio proxy endpoint.
func (m *VerdaccioManager) GetEndpoint() string {
	// Reset idle timer on access
	m.resetIdleTimer()
	return fmt.Sprintf("http://localhost:%d", m.config.Port)
}

// GetNpmEnv returns the npm environment variables to use this proxy.
func (m *VerdaccioManager) GetNpmEnv() map[string]string {
	endpoint := m.GetEndpoint()

	return map[string]string{
		"npm_config_registry": endpoint,
		"NPM_CONFIG_REGISTRY": endpoint, // Uppercase variant
	}
}

// GetNpmrc returns the npm configuration in .npmrc format.
func (m *VerdaccioManager) GetNpmrc() string {
	endpoint := m.GetEndpoint()

	return fmt.Sprintf("registry=%s\n", endpoint)
}

// GetYarnConfig returns the yarn configuration in .yarnrc.yml format.
func (m *VerdaccioManager) GetYarnConfig() string {
	endpoint := m.GetEndpoint()

	// Yarn 2+ uses .yarnrc.yml with YAML format
	return fmt.Sprintf("npmRegistryServer: \"%s\"\n", endpoint)
}

// resetIdleTimer resets the idle shutdown timer.
func (m *VerdaccioManager) resetIdleTimer() {
	if m.config.Lifecycle != "on-demand" || m.config.IdleTimeout == 0 {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.lastAccessTime = time.Now()

	if m.idleTimer != nil {
		m.idleTimer.Stop()
	}

	m.idleTimer = time.AfterFunc(m.config.IdleTimeout, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		m.Stop(ctx)
	})
}

// isPortAvailable checks if a port is available for binding.
func (m *VerdaccioManager) isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// generateConfig generates the verdaccio config.yaml file.
func (m *VerdaccioManager) generateConfig() error {
	// Generate config content
	configContent, err := GenerateVerdaccioConfig(m.config)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Write config file
	configPath := filepath.Join(m.config.Storage, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// waitForReady waits for the proxy to become ready.
func (m *VerdaccioManager) waitForReady(ctx context.Context) error {
	// Verdaccio health check endpoint
	endpoint := fmt.Sprintf("http://localhost:%d/-/ping", m.config.Port)
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("proxy did not become ready within timeout")
		case <-ticker.C:
			// Try to connect
			req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
			if err != nil {
				continue
			}
			resp, err := healthCheckClient.Do(req)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil // Proxy is ready
				}
			}
		}
	}
}

// getProxyStats counts packages and disk usage.
func (m *VerdaccioManager) getProxyStats(ctx context.Context) (int, int64) {
	// Verdaccio stores packages in storage/packageName
	packageCount := 0
	var diskUsage int64

	storageDir := filepath.Join(m.config.Storage, "storage")

	// Walk the storage directory
	filepath.Walk(storageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			diskUsage += info.Size()
			// Count .tgz files as packages
			if strings.HasSuffix(path, ".tgz") {
				packageCount++
			}
		}
		return nil
	})

	return packageCount, diskUsage
}
