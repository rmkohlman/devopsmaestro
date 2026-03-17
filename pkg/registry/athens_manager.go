package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// AthensManager manages an Athens Go module proxy instance.
type AthensManager struct {
	BaseServiceManager
	config GoModuleConfig
}

// NewAthensManager creates an AthensManager with injected dependencies.
// This is the canonical constructor — callers provide their own BinaryManager
// and ProcessManager, enabling testability and explicit dependency control.
// Returns an error if the config is invalid.
func NewAthensManager(config GoModuleConfig, binary BinaryManager, process ProcessManager) (*AthensManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Athens config: %w", err)
	}

	return &AthensManager{
		BaseServiceManager: NewBaseServiceManager(binary, process),
		config:             config,
	}, nil
}

// NewAthensManagerDefault creates an AthensManager with default production
// dependencies (AthensBinaryManager v0.14.1, ProcessManager with standard
// PID/log paths). Prefer NewAthensManager for testability.
func NewAthensManagerDefault(config GoModuleConfig) (*AthensManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Athens config: %w", err)
	}

	binaryManager := NewAthensBinaryManager(config.Storage, "0.14.1")
	processManager := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(config.Storage, "athens.pid"),
		LogFile: filepath.Join(config.Storage, "athens.log"),
	})

	return &AthensManager{
		BaseServiceManager: NewBaseServiceManager(binaryManager, processManager),
		config:             config,
	}, nil
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
	if !IsPortAvailable(m.config.Port) {
		// Port is taken — check if our service is already running on it
		if ProbeServiceHealth(m.config.Port, "/healthz", []int{200}) {
			return nil // Adopt running instance (idempotent)
		}
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

// Stop stops the Athens proxy process gracefully.
func (m *AthensManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop idle timer if running
	m.StopIdleTimerLocked()

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
	if m.config.Lifecycle == "on-demand" && m.config.IdleTimeout > 0 {
		m.ResetIdleTimer(m.config.Lifecycle, m.config.IdleTimeout, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			m.Stop(ctx)
		})
	}
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

// writeConfigFile writes the Athens config to a TOML file.
func (m *AthensManager) writeConfigFile(path string, config string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Write to file — use restricted permissions since config may contain auth settings
	return os.WriteFile(path, []byte(config), configFileMode)
}

// waitForReady waits for the proxy to become ready.
func (m *AthensManager) waitForReady(ctx context.Context) error {
	endpoint := fmt.Sprintf("http://localhost:%d/healthz", m.config.Port)
	return WaitForReady(ctx, endpoint, []int{200}, 10*time.Second)
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
