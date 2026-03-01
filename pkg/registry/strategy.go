package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devopsmaestro/models"
)

// --- Zot Strategy ---

// ZotStrategy implements RegistryStrategy for Zot container registry.
type ZotStrategy struct{}

// NewZotStrategy creates a new ZotStrategy.
func NewZotStrategy() *ZotStrategy {
	return &ZotStrategy{}
}

// ValidateConfig validates Zot-specific configuration.
func (s *ZotStrategy) ValidateConfig(config json.RawMessage) error {
	if len(config) == 0 {
		return nil // Empty config is valid
	}

	// Parse config to verify it's valid JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal(config, &configMap); err != nil {
		return fmt.Errorf("invalid JSON config: %w", err)
	}

	// Zot config validation would go here
	// For now, we accept any valid JSON
	return nil
}

// CreateManager creates a ZotManager from a Registry resource.
func (s *ZotStrategy) CreateManager(reg *models.Registry) (ServiceManager, error) {
	// Convert Registry to RegistryConfig
	config := RegistryConfig{
		Enabled:     true,
		Lifecycle:   reg.Lifecycle,
		Port:        reg.Port,
		Storage:     s.getStoragePath(reg),
		IdleTimeout: 30 * time.Minute,
	}

	// Apply defaults
	if config.Port == 0 {
		config.Port = s.GetDefaultPort()
	}

	// Parse custom config if provided
	if reg.Config.Valid && reg.Config.String != "" {
		// Could merge custom config here
		// For now, we use defaults
	}

	// Create managers (for now, use real implementations)
	// TODO: Make these injectable for testing
	binaryManager := NewBinaryManager(config.Storage, "2.0.0") // Zot v2.0.0
	processManager := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(config.Storage, "zot.pid"),
		LogFile: filepath.Join(config.Storage, "zot.log"),
	})

	// Create and return ZotManager wrapped as ServiceManager
	manager := &ZotManager{
		config:         config,
		binaryManager:  binaryManager,
		processManager: processManager,
	}

	return manager, nil
}

// GetDefaultPort returns the default Zot port (5000).
func (s *ZotStrategy) GetDefaultPort() int {
	return 5000
}

// GetDefaultStorage returns the default Zot storage path.
func (s *ZotStrategy) GetDefaultStorage() string {
	return "/var/lib/zot"
}

// getStoragePath determines the storage path for a registry.
func (s *ZotStrategy) getStoragePath(reg *models.Registry) string {
	// If config specifies storage, use it
	if reg.Config.Valid && reg.Config.String != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(reg.Config.String), &configMap); err == nil {
			if storage, ok := configMap["storage"].(string); ok && storage != "" {
				return storage
			}
		}
	}

	// Otherwise use default path under ~/.devopsmaestro
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".devopsmaestro", "registries", reg.Name)
}

// --- Athens Strategy ---

// AthensStrategy implements RegistryStrategy for Athens Go module proxy.
type AthensStrategy struct{}

// NewAthensStrategy creates a new AthensStrategy.
func NewAthensStrategy() *AthensStrategy {
	return &AthensStrategy{}
}

// ValidateConfig validates Athens-specific configuration.
func (s *AthensStrategy) ValidateConfig(config json.RawMessage) error {
	if len(config) == 0 {
		return nil // Empty config is valid
	}

	// Parse config to verify it's valid JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal(config, &configMap); err != nil {
		return fmt.Errorf("invalid JSON config: %w", err)
	}

	// Athens config validation would go here
	return nil
}

// CreateManager creates an AthensManagerAdapter from a Registry resource.
func (s *AthensStrategy) CreateManager(reg *models.Registry) (ServiceManager, error) {
	// Convert Registry to GoModuleConfig
	config := GoModuleConfig{
		Enabled:     true,
		Lifecycle:   reg.Lifecycle,
		Port:        reg.Port,
		Storage:     s.getStoragePath(reg),
		IdleTimeout: 30 * time.Minute,
		Upstreams:   defaultUpstreams(),
	}

	// Apply defaults
	if config.Port == 0 {
		config.Port = s.GetDefaultPort()
	}

	// Create AthensManager and wrap it in adapter
	athensManager := NewAthensManager(config)
	return &AthensManagerAdapter{manager: athensManager}, nil
}

// GetDefaultPort returns the default Athens port (3000).
func (s *AthensStrategy) GetDefaultPort() int {
	return 3000
}

// GetDefaultStorage returns the default Athens storage path.
func (s *AthensStrategy) GetDefaultStorage() string {
	return "/var/lib/athens"
}

// getStoragePath determines the storage path for a registry.
func (s *AthensStrategy) getStoragePath(reg *models.Registry) string {
	// If config specifies storage, use it
	if reg.Config.Valid && reg.Config.String != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(reg.Config.String), &configMap); err == nil {
			if storage, ok := configMap["storage"].(string); ok && storage != "" {
				return storage
			}
		}
	}

	// Otherwise use default path under ~/.devopsmaestro
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".devopsmaestro", "registries", reg.Name)
}

// AthensManagerAdapter adapts AthensManager to ServiceManager interface.
// Athens has a slightly different interface (returns full URLs), so we adapt it.
type AthensManagerAdapter struct {
	manager *AthensManager
}

// Start starts the Athens proxy.
func (a *AthensManagerAdapter) Start(ctx context.Context) error {
	return a.manager.Start(ctx)
}

// Stop stops the Athens proxy.
func (a *AthensManagerAdapter) Stop(ctx context.Context) error {
	return a.manager.Stop(ctx)
}

// IsRunning checks if Athens is running.
func (a *AthensManagerAdapter) IsRunning(ctx context.Context) bool {
	return a.manager.IsRunning(ctx)
}

// GetEndpoint returns the Athens endpoint (full URL).
func (a *AthensManagerAdapter) GetEndpoint() string {
	return a.manager.GetEndpoint()
}

// --- Devpi Strategy ---

// DevpiStrategy implements RegistryStrategy for devpi PyPI proxy.
type DevpiStrategy struct{}

// NewDevpiStrategy creates a new DevpiStrategy.
func NewDevpiStrategy() *DevpiStrategy {
	return &DevpiStrategy{}
}

// ValidateConfig validates devpi-specific configuration.
func (s *DevpiStrategy) ValidateConfig(config json.RawMessage) error {
	if len(config) == 0 {
		return nil // Empty config is valid
	}

	// Parse config to verify it's valid JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal(config, &configMap); err != nil {
		return fmt.Errorf("invalid JSON config: %w", err)
	}

	// Devpi config validation would go here
	return nil
}

// CreateManager creates a DevpiManagerAdapter from a Registry resource.
func (s *DevpiStrategy) CreateManager(reg *models.Registry) (ServiceManager, error) {
	if reg == nil {
		return nil, fmt.Errorf("registry cannot be nil")
	}

	// Convert Registry to PyPIProxyConfig
	config := PyPIProxyConfig{
		Enabled:     true,
		Lifecycle:   reg.Lifecycle,
		Port:        reg.Port,
		Storage:     s.getStoragePath(reg),
		IdleTimeout: 30 * time.Minute,
		Upstreams:   defaultPyPIUpstreams(),
	}

	// Apply defaults
	if config.Port == 0 {
		config.Port = s.GetDefaultPort()
	}

	// Create DevpiManager and wrap it in adapter
	devpiManager := NewDevpiManager(config)
	return &DevpiManagerAdapter{manager: devpiManager}, nil
}

// GetDefaultPort returns the default devpi port (3141).
func (s *DevpiStrategy) GetDefaultPort() int {
	return 3141
}

// GetDefaultStorage returns the default devpi storage path.
func (s *DevpiStrategy) GetDefaultStorage() string {
	return "/var/lib/devpi"
}

// getStoragePath determines the storage path for a registry.
func (s *DevpiStrategy) getStoragePath(reg *models.Registry) string {
	// If config specifies storage, use it
	if reg.Config.Valid && reg.Config.String != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(reg.Config.String), &configMap); err == nil {
			if storage, ok := configMap["storage"].(string); ok && storage != "" {
				return storage
			}
		}
	}

	// Otherwise use default path under ~/.devopsmaestro
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".devopsmaestro", "registries", reg.Name)
}

// DevpiManagerAdapter adapts DevpiManager to ServiceManager interface.
type DevpiManagerAdapter struct {
	manager *DevpiManager
}

// Start starts the devpi proxy.
func (d *DevpiManagerAdapter) Start(ctx context.Context) error {
	return d.manager.Start(ctx)
}

// Stop stops the devpi proxy.
func (d *DevpiManagerAdapter) Stop(ctx context.Context) error {
	return d.manager.Stop(ctx)
}

// IsRunning checks if devpi is running.
func (d *DevpiManagerAdapter) IsRunning(ctx context.Context) bool {
	return d.manager.IsRunning(ctx)
}

// GetEndpoint returns the devpi endpoint (full URL).
func (d *DevpiManagerAdapter) GetEndpoint() string {
	return d.manager.GetEndpoint()
}

// --- Verdaccio Strategy ---

// VerdaccioStrategy implements RegistryStrategy for verdaccio npm proxy.
type VerdaccioStrategy struct{}

// NewVerdaccioStrategy creates a new VerdaccioStrategy.
func NewVerdaccioStrategy() *VerdaccioStrategy {
	return &VerdaccioStrategy{}
}

// ValidateConfig validates verdaccio-specific configuration.
func (s *VerdaccioStrategy) ValidateConfig(config json.RawMessage) error {
	if len(config) == 0 {
		return nil // Empty config is valid
	}

	// Parse config to verify it's valid JSON
	var configMap map[string]interface{}
	if err := json.Unmarshal(config, &configMap); err != nil {
		return fmt.Errorf("invalid JSON config: %w", err)
	}

	// Verdaccio config validation would go here
	return nil
}

// CreateManager creates a VerdaccioManagerAdapter from a Registry resource.
func (s *VerdaccioStrategy) CreateManager(reg *models.Registry) (ServiceManager, error) {
	if reg == nil {
		return nil, fmt.Errorf("registry cannot be nil")
	}

	// Convert Registry to NpmProxyConfig
	config := NpmProxyConfig{
		Enabled:     true,
		Lifecycle:   reg.Lifecycle,
		Port:        reg.Port,
		Storage:     s.getStoragePath(reg),
		IdleTimeout: 30 * time.Minute,
		Upstreams:   defaultNpmUpstreams(),
	}

	// Apply defaults
	if config.Port == 0 {
		config.Port = s.GetDefaultPort()
	}

	// Set default lifecycle if not specified
	if config.Lifecycle == "" {
		config.Lifecycle = "manual"
	}

	// Create VerdaccioManager and wrap it in adapter
	verdaccioManager := NewVerdaccioManager(config)
	return &VerdaccioManagerAdapter{manager: verdaccioManager}, nil
}

// GetDefaultPort returns the default verdaccio port (4873).
func (s *VerdaccioStrategy) GetDefaultPort() int {
	return 4873
}

// GetDefaultStorage returns the default verdaccio storage path.
func (s *VerdaccioStrategy) GetDefaultStorage() string {
	return "/var/lib/verdaccio"
}

// getStoragePath determines the storage path for a registry.
func (s *VerdaccioStrategy) getStoragePath(reg *models.Registry) string {
	// If config specifies storage, use it
	if reg.Config.Valid && reg.Config.String != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(reg.Config.String), &configMap); err == nil {
			if storage, ok := configMap["storage"].(string); ok && storage != "" {
				return storage
			}
		}
	}

	// Otherwise use default path under ~/.devopsmaestro
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, ".devopsmaestro", "registries", reg.Name)
}

// VerdaccioManagerAdapter adapts VerdaccioManager to ServiceManager interface.
type VerdaccioManagerAdapter struct {
	manager *VerdaccioManager
}

// Start starts the verdaccio proxy.
func (v *VerdaccioManagerAdapter) Start(ctx context.Context) error {
	return v.manager.Start(ctx)
}

// Stop stops the verdaccio proxy.
func (v *VerdaccioManagerAdapter) Stop(ctx context.Context) error {
	return v.manager.Stop(ctx)
}

// IsRunning checks if verdaccio is running.
func (v *VerdaccioManagerAdapter) IsRunning(ctx context.Context) bool {
	return v.manager.IsRunning(ctx)
}

// GetEndpoint returns the verdaccio endpoint (full URL).
func (v *VerdaccioManagerAdapter) GetEndpoint() string {
	return v.manager.GetEndpoint()
}

// --- Stub Strategies for Future Implementation ---

// StubStrategy is a base strategy for registries not yet implemented.
type StubStrategy struct {
	registryType   string
	defaultPort    int
	defaultStorage string
}

// ValidateConfig accepts any valid JSON for stub registries.
func (s *StubStrategy) ValidateConfig(config json.RawMessage) error {
	if len(config) == 0 {
		return nil
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal(config, &configMap); err != nil {
		return fmt.Errorf("invalid JSON config: %w", err)
	}
	return nil
}

// CreateManager returns a not implemented error.
func (s *StubStrategy) CreateManager(reg *models.Registry) (ServiceManager, error) {
	return nil, fmt.Errorf("%s registry not implemented yet", s.registryType)
}

// GetDefaultPort returns the default port for this registry type.
func (s *StubStrategy) GetDefaultPort() int {
	return s.defaultPort
}

// GetDefaultStorage returns the default storage path for this registry type.
func (s *StubStrategy) GetDefaultStorage() string {
	return s.defaultStorage
}

// NewSquidStrategy creates a stub strategy for squid (HTTP proxy/cache).
func NewSquidStrategy() RegistryStrategy {
	return &StubStrategy{
		registryType:   "squid",
		defaultPort:    3128,
		defaultStorage: "/var/cache/squid",
	}
}
