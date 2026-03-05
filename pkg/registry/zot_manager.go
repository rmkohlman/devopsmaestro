package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// ZotManager implements RegistryManager for Zot registry.
type ZotManager struct {
	BaseServiceManager
	config RegistryConfig
}

// Start starts the registry process.
func (z *ZotManager) Start(ctx context.Context) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	// Check if already running - idempotent
	if z.processManager.IsRunning() {
		return nil
	}

	// Ensure binary exists
	binaryPath, err := z.binaryManager.EnsureBinary(ctx)
	if err != nil {
		return fmt.Errorf("failed to ensure binary: %w", err)
	}

	// Generate Zot config
	zotConfig, err := GenerateZotConfig(z.config)
	if err != nil {
		return fmt.Errorf("failed to generate config: %w", err)
	}

	// Write config file
	configPath := filepath.Join(z.config.Storage, "config.json")
	if err := z.writeConfigFile(configPath, zotConfig); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	// Check if port is available
	if !IsPortAvailable(z.config.Port) {
		return fmt.Errorf("%w: port %d is already in use", ErrPortInUse, z.config.Port)
	}

	// Prepare process config
	procConfig := ProcessConfig{
		PIDFile:         filepath.Join(z.config.Storage, "zot.pid"),
		LogFile:         filepath.Join(z.config.Storage, "zot.log"),
		WorkingDir:      z.config.Storage,
		ShutdownTimeout: 10 * time.Second,
	}

	// Start Zot process
	args := []string{"serve", configPath}
	if err := z.processManager.Start(ctx, binaryPath, args, procConfig); err != nil {
		return fmt.Errorf("failed to start registry: %w", err)
	}

	// Record start time
	z.RecordStartLocked()

	// Setup idle timer if on-demand mode
	z.ResetIdleTimerLocked(z.config.Lifecycle, z.config.IdleTimeout, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		z.Stop(ctx)
	})

	// Wait for registry to be ready
	if err := z.waitForReady(ctx); err != nil {
		z.processManager.Stop(ctx)
		return fmt.Errorf("registry failed to become ready: %w", err)
	}

	return nil
}

// Stop stops the registry process gracefully.
func (z *ZotManager) Stop(ctx context.Context) error {
	z.mu.Lock()
	defer z.mu.Unlock()

	// Stop idle timer if running
	z.StopIdleTimerLocked()

	// Stop process (idempotent)
	return z.processManager.Stop(ctx)
}

// Status returns the current status of the registry.
func (z *ZotManager) Status(ctx context.Context) (*RegistryStatus, error) {
	z.mu.RLock()
	defer z.mu.RUnlock()

	running := z.processManager.IsRunning()
	state := "stopped"
	if running {
		state = "running"
	}

	var uptime time.Duration
	if running && !z.startTime.IsZero() {
		uptime = time.Since(z.startTime)
	}

	// Get version
	version, _ := z.binaryManager.GetVersion(ctx)

	// Get image count and disk usage from registry API
	imageCount := 0
	var diskUsage int64

	if running {
		// Query registry API for stats
		imageCount, diskUsage = z.getRegistryStats(ctx)
	}

	return &RegistryStatus{
		State:      state,
		PID:        z.processManager.GetPID(),
		Port:       z.config.Port,
		Storage:    z.config.Storage,
		Version:    version,
		Uptime:     uptime,
		ImageCount: imageCount,
		DiskUsage:  diskUsage,
	}, nil
}

// EnsureRunning starts the registry if it's not running.
func (z *ZotManager) EnsureRunning(ctx context.Context) error {
	if z.IsRunning(ctx) {
		return nil
	}
	return z.Start(ctx)
}

// IsRunning checks if the registry is currently running.
func (z *ZotManager) IsRunning(ctx context.Context) bool {
	z.mu.RLock()
	defer z.mu.RUnlock()
	return z.processManager.IsRunning()
}

// GetEndpoint returns the registry endpoint.
func (z *ZotManager) GetEndpoint() string {
	// Reset idle timer on access
	if z.config.Lifecycle == "on-demand" && z.config.IdleTimeout > 0 {
		z.ResetIdleTimer(z.config.Lifecycle, z.config.IdleTimeout, func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			z.Stop(ctx)
		})
	}
	return fmt.Sprintf("http://localhost:%d", z.config.Port)
}

// Prune removes unused images from the registry.
func (z *ZotManager) Prune(ctx context.Context, opts PruneOptions) (*PruneResult, error) {
	// Check if registry is running
	if !z.IsRunning(ctx) {
		return nil, fmt.Errorf("%w: cannot prune when registry is not running", ErrNotRunning)
	}

	result := &PruneResult{
		Images: []string{},
	}

	// For dry run, just report what would be removed
	if opts.DryRun {
		// Query registry API to get list of images
		// For now, return empty result
		return result, nil
	}

	// For actual pruning, would call Zot's GC API
	// This is not implemented in v1 of Zot, so we just return empty result
	// Future: Use Zot's scrub/GC features when available

	return result, nil
}

// writeConfigFile writes the Zot config to a file.
func (z *ZotManager) writeConfigFile(path string, config map[string]interface{}) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	// Marshal config to JSON
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(path, data, 0644)
}

// waitForReady waits for the registry to become ready.
func (z *ZotManager) waitForReady(ctx context.Context) error {
	endpoint := fmt.Sprintf("http://localhost:%d/v2/", z.config.Port)
	return WaitForReady(ctx, endpoint, []int{http.StatusOK, http.StatusUnauthorized}, 10*time.Second)
}

// getRegistryStats queries the registry API for image count and disk usage.
func (z *ZotManager) getRegistryStats(ctx context.Context) (int, int64) {
	// Query Zot's catalog API to get image count
	endpoint := fmt.Sprintf("http://localhost:%d/v2/_catalog", z.config.Port)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return 0, 0
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, 0
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0
	}

	// Parse catalog response
	var catalog struct {
		Repositories []string `json:"repositories"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&catalog); err != nil {
		return 0, 0
	}

	imageCount := len(catalog.Repositories)

	// Get disk usage by walking storage directory
	var diskUsage int64
	filepath.Walk(z.config.Storage, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			diskUsage += info.Size()
		}
		return nil
	})

	return imageCount, diskUsage
}
