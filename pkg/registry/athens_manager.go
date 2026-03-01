package registry

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AthensManager manages an Athens Go module proxy instance.
type AthensManager struct {
	config         GoModuleConfig
	binaryManager  BinaryManager
	processManager ProcessManager

	mu             sync.RWMutex
	startTime      time.Time
	idleTimer      *time.Timer
	lastAccessTime time.Time
}

// NewAthensManager creates a new AthensManager with the given configuration.
func NewAthensManager(config GoModuleConfig) *AthensManager {
	// Validate config
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid Athens config: %v", err))
	}

	// Create real binary and process managers
	binaryManager := NewAthensBinaryManager(config.Storage, "0.14.1")
	processManager := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(config.Storage, "athens.pid"),
		LogFile: filepath.Join(config.Storage, "athens.log"),
	})

	return &AthensManager{
		config:         config,
		binaryManager:  binaryManager,
		processManager: processManager,
	}
}

// NewAthensManagerWithDeps creates an AthensManager with injected dependencies.
func NewAthensManagerWithDeps(config GoModuleConfig, binary BinaryManager, process ProcessManager) *AthensManager {
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid Athens config: %v", err))
	}

	return &AthensManager{
		config:         config,
		binaryManager:  binary,
		processManager: process,
	}
}

// Start starts the Athens proxy process.
func (m *AthensManager) Start(ctx context.Context) error {
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

	// Generate Athens config
	athensConfig, err := GenerateAthensConfig(m.config)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Write config file
	configPath := filepath.Join(m.config.Storage, "config.toml")
	if err := m.writeConfigFile(configPath, athensConfig); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Check if port is available
	if !m.isPortAvailable(m.config.Port) {
		return fmt.Errorf("%w: port %d is already in use", ErrPortInUse, m.config.Port)
	}

	// Prepare process config
	procConfig := ProcessConfig{
		PIDFile:         filepath.Join(m.config.Storage, "athens.pid"),
		LogFile:         filepath.Join(m.config.Storage, "athens.log"),
		WorkingDir:      m.config.Storage,
		ShutdownTimeout: 10 * time.Second,
	}

	// Start Athens process
	// Athens uses environment variables for config instead of --config flag
	args := []string{}
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

// Stop stops the Athens proxy process gracefully.
func (m *AthensManager) Stop(ctx context.Context) error {
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

// Status returns the current status of the Athens proxy.
func (m *AthensManager) Status(ctx context.Context) (*GoModuleProxyStatus, error) {
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

	// Get module count and disk usage
	moduleCount := 0
	var diskUsage int64

	if running {
		// Count modules in storage directory
		moduleCount, diskUsage = m.getProxyStats(ctx)
	}

	return &GoModuleProxyStatus{
		State:       state,
		PID:         m.processManager.GetPID(),
		Port:        m.config.Port,
		Storage:     m.config.Storage,
		Version:     version,
		Uptime:      uptime,
		ModuleCount: moduleCount,
		DiskUsage:   diskUsage,
	}, nil
}

// EnsureRunning starts the proxy if it's not running.
func (m *AthensManager) EnsureRunning(ctx context.Context) error {
	if m.IsRunning(ctx) {
		return nil
	}
	return m.Start(ctx)
}

// IsRunning checks if the Athens proxy is currently running.
func (m *AthensManager) IsRunning(ctx context.Context) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.processManager.IsRunning()
}

// GetEndpoint returns the Athens proxy endpoint.
func (m *AthensManager) GetEndpoint() string {
	// Reset idle timer on access
	m.resetIdleTimer()
	return fmt.Sprintf("http://localhost:%d", m.config.Port)
}

// GetGoEnv returns the Go environment variables to use this proxy.
func (m *AthensManager) GetGoEnv() map[string]string {
	endpoint := m.GetEndpoint()
	return map[string]string{
		"GOPROXY":   endpoint,
		"GONOSUMDB": "*", // Disable checksum verification for local development
		"GOPRIVATE": "*", // Mark all modules as private for local development
	}
}

// resetIdleTimer resets the idle shutdown timer.
func (m *AthensManager) resetIdleTimer() {
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

// writeConfigFile writes the Athens config to a TOML file.
func (m *AthensManager) writeConfigFile(path string, config string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, []byte(config), 0644)
}

// isPortAvailable checks if a port is available for binding.
func (m *AthensManager) isPortAvailable(port int) bool {
	addr := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return false
	}
	listener.Close()
	return true
}

// waitForReady waits for the proxy to become ready.
func (m *AthensManager) waitForReady(ctx context.Context) error {
	// Athens health check endpoint
	endpoint := fmt.Sprintf("http://localhost:%d/healthz", m.config.Port)
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
			resp, err := http.Get(endpoint)
			if err == nil {
				resp.Body.Close()
				if resp.StatusCode == http.StatusOK {
					return nil // Proxy is ready
				}
			}
		}
	}
}

// getProxyStats counts modules and disk usage.
func (m *AthensManager) getProxyStats(ctx context.Context) (int, int64) {
	// Athens stores modules in a directory structure
	// Count directories in storage path
	moduleCount := 0
	var diskUsage int64

	// Walk the storage directory
	filepath.Walk(m.config.Storage, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			diskUsage += info.Size()
			// Count go.mod files as modules
			if filepath.Base(path) == "go.mod" {
				moduleCount++
			}
		}
		return nil
	})

	return moduleCount, diskUsage
}

// MockGoModuleProxy is a mock implementation of GoModuleProxy for testing.
type MockGoModuleProxy struct {
	StartFunc         func(ctx context.Context) error
	StopFunc          func(ctx context.Context) error
	StatusFunc        func(ctx context.Context) (*GoModuleProxyStatus, error)
	EnsureRunningFunc func(ctx context.Context) error
	IsRunningFunc     func(ctx context.Context) bool
	GetEndpointFunc   func() string
	GetGoEnvFunc      func() map[string]string
}

// Start calls the mock's StartFunc.
func (m *MockGoModuleProxy) Start(ctx context.Context) error {
	if m.StartFunc != nil {
		return m.StartFunc(ctx)
	}
	return nil
}

// Stop calls the mock's StopFunc.
func (m *MockGoModuleProxy) Stop(ctx context.Context) error {
	if m.StopFunc != nil {
		return m.StopFunc(ctx)
	}
	return nil
}

// Status calls the mock's StatusFunc.
func (m *MockGoModuleProxy) Status(ctx context.Context) (*GoModuleProxyStatus, error) {
	if m.StatusFunc != nil {
		return m.StatusFunc(ctx)
	}
	return &GoModuleProxyStatus{}, nil
}

// EnsureRunning calls the mock's EnsureRunningFunc.
func (m *MockGoModuleProxy) EnsureRunning(ctx context.Context) error {
	if m.EnsureRunningFunc != nil {
		return m.EnsureRunningFunc(ctx)
	}
	return nil
}

// IsRunning calls the mock's IsRunningFunc.
func (m *MockGoModuleProxy) IsRunning(ctx context.Context) bool {
	if m.IsRunningFunc != nil {
		return m.IsRunningFunc(ctx)
	}
	return false
}

// GetEndpoint calls the mock's GetEndpointFunc.
func (m *MockGoModuleProxy) GetEndpoint() string {
	if m.GetEndpointFunc != nil {
		return m.GetEndpointFunc()
	}
	return "http://localhost:3000"
}

// GetGoEnv calls the mock's GetGoEnvFunc.
func (m *MockGoModuleProxy) GetGoEnv() map[string]string {
	if m.GetGoEnvFunc != nil {
		return m.GetGoEnvFunc()
	}
	return map[string]string{
		"GOPROXY": "http://localhost:3000",
	}
}
