package registry

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// VerdaccioManager manages a verdaccio npm proxy instance.
type VerdaccioManager struct {
	BaseServiceManager
	config NpmProxyConfig
}

// NewVerdaccioManager creates a VerdaccioManager with injected dependencies.
// This is the canonical constructor — callers provide their own BinaryManager
// and ProcessManager, enabling testability and explicit dependency control.
// Returns an error if the config is invalid.
func NewVerdaccioManager(config NpmProxyConfig, binary BinaryManager, process ProcessManager) (*VerdaccioManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid verdaccio config: %w", err)
	}

	return &VerdaccioManager{
		BaseServiceManager: NewBaseServiceManager(binary, process),
		config:             config,
	}, nil
}

// NewVerdaccioManagerDefault creates a VerdaccioManager with default production
// dependencies (NpmBinaryManager for verdaccio v6.1.2, ProcessManager with
// standard PID/log paths). Prefer NewVerdaccioManager for testability.
func NewVerdaccioManagerDefault(config NpmProxyConfig) (*VerdaccioManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid verdaccio config: %w", err)
	}

	binaryManager := NewNpmBinaryManager("verdaccio", "6.1.2")
	processManager := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(config.Storage, "verdaccio.pid"),
		LogFile: filepath.Join(config.Storage, "verdaccio.log"),
	})

	return &VerdaccioManager{
		BaseServiceManager: NewBaseServiceManager(binaryManager, processManager),
		config:             config,
	}, nil
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
	if !IsPortAvailable(m.config.Port) {
		// Port is taken — check if our service is already running on it (#384)
		if ProbeServiceHealth(m.config.Port, "/-/ping", []int{200}) {
			return nil // Adopt running instance (idempotent)
		}
		return fmt.Errorf("%w: port %d is already in use", ErrPortInUse, m.config.Port)
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(m.config.Storage, 0700); err != nil {
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
		// If the process is already running (adopted from stale PID), verify
		// it's actually serving before returning an error (#385).
		if errors.Is(err, ErrProcessAlreadyRunning) {
			if ProbeServiceHealth(m.config.Port, "/-/ping", []int{200}) {
				m.RecordStartLocked()
				return nil // Adopt the running instance
			}
			// PID exists but service not healthy — kill and retry
			m.processManager.Stop(ctx)
			if retryErr := m.processManager.Start(ctx, binaryPath, args, procConfig); retryErr != nil {
				return fmt.Errorf("failed to start proxy after cleanup: %w", retryErr)
			}
		} else {
			return fmt.Errorf("failed to start proxy: %w", err)
		}
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

// Stop stops the verdaccio proxy process gracefully.
func (m *VerdaccioManager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Stop idle timer if running
	m.StopIdleTimerLocked()

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
	if m.config.Lifecycle == "on-demand" && m.config.IdleTimeout > 0 {
		m.ResetIdleTimer(m.config.Lifecycle, m.config.IdleTimeout, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			m.Stop(ctx)
		})
	}
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

// generateConfig generates the verdaccio config.yaml file.
func (m *VerdaccioManager) generateConfig() error {
	// Generate config content
	configContent, err := GenerateVerdaccioConfig(m.config)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Write config file
	configPath := filepath.Join(m.config.Storage, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), configFileMode); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// waitForReady waits for the proxy to become ready.
func (m *VerdaccioManager) waitForReady(ctx context.Context) error {
	endpoint := fmt.Sprintf("http://localhost:%d/-/ping", m.config.Port)
	return WaitForReady(ctx, endpoint, []int{200}, 10*time.Second)
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
