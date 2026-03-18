package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/paths"
)

// validateStoragePath ensures the resolved path is within the DVM root directory (~/.devopsmaestro/).
// Returns an error if the path escapes the allowed base directory.
// The homeDir parameter is accepted explicitly to support deterministic testing.
func validateStoragePath(resolved, homeDir string) error {
	if !filepath.IsAbs(resolved) {
		return fmt.Errorf("storage path must be absolute: %s", resolved)
	}

	baseDir := filepath.Join(homeDir, ".devopsmaestro")

	absResolved, err := filepath.Abs(filepath.Clean(resolved))
	if err != nil {
		return fmt.Errorf("invalid storage path: %w", err)
	}
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return fmt.Errorf("invalid base directory: %w", err)
	}

	// Check that the resolved path is within the base directory.
	// filepath.Rel returns a ".." prefix when the target escapes the base.
	rel, err := filepath.Rel(absBase, absResolved)
	if err != nil || strings.HasPrefix(rel, "..") {
		return fmt.Errorf("storage path %q must be within %s", resolved, baseDir)
	}

	return nil
}

// resolveStoragePath extracts the storage path from a registry's config,
// falling back to the default ~/.devopsmaestro/registries/{name} path.
// The configKey parameter specifies which JSON key to look up (e.g. "storage", "cacheDir").
// Returns an error if a user-provided path fails validation.
func resolveStoragePath(reg *models.Registry, configKey string) (string, error) {
	if reg.Config.Valid && reg.Config.String != "" {
		var configMap map[string]interface{}
		if err := json.Unmarshal([]byte(reg.Config.String), &configMap); err == nil {
			if val, ok := configMap[configKey].(string); ok && val != "" {
				homeDir, err := os.UserHomeDir()
				if err != nil {
					return "", fmt.Errorf("cannot determine home directory: %w", err)
				}
				if err := validateStoragePath(val, homeDir); err != nil {
					return "", err
				}
				return val, nil
			}
		}
	}

	// Otherwise use default path under ~/.devopsmaestro
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return paths.New(homeDir).RegistryDir(reg.Name), nil
}

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
	if reg == nil {
		return nil, fmt.Errorf("registry cannot be nil")
	}

	// Resolve and validate storage path
	storagePath, err := resolveStoragePath(reg, "storage")
	if err != nil {
		return nil, fmt.Errorf("invalid storage path for registry %q: %w", reg.Name, err)
	}

	// Convert Registry to RegistryConfig
	config := RegistryConfig{
		Enabled:     true,
		Lifecycle:   reg.Lifecycle,
		Port:        reg.Port,
		Storage:     storagePath,
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

	// Create managers with explicit dependency construction.
	// Use explicit version from registry if set, otherwise fall back to strategy default (RC-2)
	version := reg.Version
	if version == "" {
		version = s.GetDefaultVersion()
	}
	binaryManager := NewBinaryManager(filepath.Join(config.Storage, "bin"), version)
	processManager := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(config.Storage, "zot.pid"),
		LogFile: filepath.Join(config.Storage, "zot.log"),
	})

	// Create and return ZotManager wrapped as ServiceManager
	manager := NewZotManagerWithDeps(config, binaryManager, processManager)

	return manager, nil
}

// GetDefaultPort returns the default Zot port (5001).
func (s *ZotStrategy) GetDefaultPort() int {
	return 5001
}

// GetDefaultStorage returns the default Zot storage path.
func (s *ZotStrategy) GetDefaultStorage() string {
	return "/var/lib/zot"
}

// GetDefaultVersion returns the default Zot binary version.
func (s *ZotStrategy) GetDefaultVersion() string {
	return "2.1.15"
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
	if reg == nil {
		return nil, fmt.Errorf("registry cannot be nil")
	}

	// Resolve and validate storage path
	storagePath, err := resolveStoragePath(reg, "storage")
	if err != nil {
		return nil, fmt.Errorf("invalid storage path for registry %q: %w", reg.Name, err)
	}

	// Convert Registry to GoModuleConfig
	config := GoModuleConfig{
		Enabled:     true,
		Lifecycle:   reg.Lifecycle,
		Port:        reg.Port,
		Storage:     storagePath,
		IdleTimeout: 30 * time.Minute,
		Upstreams:   defaultUpstreams(),
	}

	// Apply defaults
	if config.Port == 0 {
		config.Port = s.GetDefaultPort()
	}

	// Create AthensManager with explicit dependency construction and wrap in adapter.
	athensManager, err := NewAthensManagerDefault(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Athens manager: %w", err)
	}
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

// GetDefaultVersion returns empty string (Athens is externally managed).
func (s *AthensStrategy) GetDefaultVersion() string {
	return ""
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

	// Resolve and validate storage path
	storagePath, err := resolveStoragePath(reg, "storage")
	if err != nil {
		return nil, fmt.Errorf("invalid storage path for registry %q: %w", reg.Name, err)
	}

	// Convert Registry to PyPIProxyConfig
	config := PyPIProxyConfig{
		Enabled:     true,
		Lifecycle:   reg.Lifecycle,
		Port:        reg.Port,
		Storage:     storagePath,
		IdleTimeout: 30 * time.Minute,
		Upstreams:   defaultPyPIUpstreams(),
	}

	// Apply defaults
	if config.Port == 0 {
		config.Port = s.GetDefaultPort()
	}

	// Create DevpiManager with explicit dependency construction and wrap in adapter.
	devpiManager, err := NewDevpiManagerDefault(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create devpi manager: %w", err)
	}
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

// GetDefaultVersion returns empty string (devpi is externally managed via pipx).
func (s *DevpiStrategy) GetDefaultVersion() string {
	return ""
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

	// Resolve and validate storage path
	storagePath, err := resolveStoragePath(reg, "storage")
	if err != nil {
		return nil, fmt.Errorf("invalid storage path for registry %q: %w", reg.Name, err)
	}

	// Convert Registry to NpmProxyConfig
	config := NpmProxyConfig{
		Enabled:     true,
		Lifecycle:   reg.Lifecycle,
		Port:        reg.Port,
		Storage:     storagePath,
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

	// Create VerdaccioManager with explicit dependency construction and wrap in adapter.
	verdaccioManager, err := NewVerdaccioManagerDefault(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create verdaccio manager: %w", err)
	}
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

// GetDefaultVersion returns empty string (verdaccio is externally managed via npm).
func (s *VerdaccioStrategy) GetDefaultVersion() string {
	return ""
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

// --- Registry Type Constants ---

// RegistryTypeSquid is the registry type for squid HTTP proxy.
const RegistryTypeSquid = "squid"

// --- Squid Strategy ---

// SquidStrategy implements RegistryStrategy for squid HTTP proxy/cache.
type SquidStrategy struct{}

// NewSquidStrategy creates a new SquidStrategy.
func NewSquidStrategy() *SquidStrategy {
	return &SquidStrategy{}
}

// ValidateConfig validates squid-specific configuration.
func (s *SquidStrategy) ValidateConfig(config json.RawMessage) error {
	if len(config) == 0 {
		return nil // Empty config is valid
	}

	// Handle null JSON
	if string(config) == "null" {
		return nil
	}

	// Parse config as HttpProxyConfig
	var proxyConfig HttpProxyConfig
	if err := json.Unmarshal(config, &proxyConfig); err != nil {
		return fmt.Errorf("invalid JSON config: %w", err)
	}

	// Config parsed successfully - validation happens when applying defaults
	return nil
}

// CreateManager creates a SquidManagerAdapter from a Registry resource.
func (s *SquidStrategy) CreateManager(reg *models.Registry) (ServiceManager, error) {
	if reg == nil {
		return nil, fmt.Errorf("registry cannot be nil")
	}

	// Start with default config
	config := DefaultHttpProxyConfig()

	// Override port if specified
	if reg.Port != 0 {
		config.Port = reg.Port
	}

	// Parse custom config if provided
	if reg.Config.Valid && reg.Config.String != "" {
		var customConfig HttpProxyConfig
		if err := json.Unmarshal([]byte(reg.Config.String), &customConfig); err == nil {
			// Merge custom config values
			if customConfig.Port != 0 {
				config.Port = customConfig.Port
			}
			if customConfig.CacheDir != "" {
				config.CacheDir = customConfig.CacheDir
			}
			if customConfig.LogDir != "" {
				config.LogDir = customConfig.LogDir
			}
			if customConfig.PidFile != "" {
				config.PidFile = customConfig.PidFile
			}
			if customConfig.CacheSizeMB != 0 {
				config.CacheSizeMB = customConfig.CacheSizeMB
			}
			if customConfig.MaxObjectSizeMB != 0 {
				config.MaxObjectSizeMB = customConfig.MaxObjectSizeMB
			}
			if customConfig.MemoryCacheMB != 0 {
				config.MemoryCacheMB = customConfig.MemoryCacheMB
			}
		}
	}

	// Determine storage path - override defaults if not set in custom config
	// For Squid, if "cacheDir" is specified in config, use its parent directory
	// as the base storage path. Otherwise fall back to the default registry path.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	defaultPath := paths.New(homeDir).RegistryDir(reg.Name)
	storagePath, err := resolveStoragePath(reg, "cacheDir")
	if err != nil {
		return nil, fmt.Errorf("invalid storage path for registry %q: %w", reg.Name, err)
	}
	if storagePath != defaultPath {
		// cacheDir was found in config — use its parent as the storage root
		storagePath = filepath.Dir(storagePath)
	}
	if config.CacheDir == "" || config.CacheDir == DefaultHttpProxyConfig().CacheDir {
		config.CacheDir = filepath.Join(storagePath, "cache")
	}
	if config.LogDir == "" || config.LogDir == DefaultHttpProxyConfig().LogDir {
		config.LogDir = filepath.Join(storagePath, "logs")
	}
	if config.PidFile == "" || config.PidFile == DefaultHttpProxyConfig().PidFile {
		config.PidFile = filepath.Join(storagePath, "squid.pid")
	}

	// Apply any remaining defaults
	config.ApplyDefaults()

	// Create SquidManager and wrap it in adapter
	squidManager := NewSquidManager(config)
	return NewSquidManagerAdapter(squidManager), nil
}

// GetDefaultPort returns the default squid port (3128).
func (s *SquidStrategy) GetDefaultPort() int {
	return 3128
}

// GetDefaultStorage returns the default squid storage path.
func (s *SquidStrategy) GetDefaultStorage() string {
	return "/var/cache/squid"
}

// GetDefaultVersion returns empty string (squid is externally managed via brew).
func (s *SquidStrategy) GetDefaultVersion() string {
	return ""
}
