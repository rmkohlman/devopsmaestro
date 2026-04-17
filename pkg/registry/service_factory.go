package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devopsmaestro/models"

	"github.com/rmkohlman/MaestroSDK/paths"
)

// ServiceFactory creates ServiceManager instances based on registry type.
// It uses the strategy pattern to delegate type-specific behavior.
type ServiceFactory struct {
	strategies map[string]RegistryStrategy
}

// NewServiceFactory creates a new ServiceFactory with all registered strategies.
func NewServiceFactory() *ServiceFactory {
	return &ServiceFactory{
		strategies: map[string]RegistryStrategy{
			"zot":       NewZotStrategy(),
			"athens":    NewAthensStrategy(),
			"devpi":     NewDevpiStrategy(),
			"verdaccio": NewVerdaccioStrategy(),
			"squid":     NewSquidStrategy(),
		},
	}
}

// GetStrategy returns the strategy for a given registry type.
func (f *ServiceFactory) GetStrategy(registryType string) (RegistryStrategy, error) {
	strategy, ok := f.strategies[registryType]
	if !ok {
		return nil, fmt.Errorf("unsupported registry type: %s", registryType)
	}
	return strategy, nil
}

// CreateManager creates a ServiceManager for the given Registry resource.
// It delegates to the appropriate strategy based on the registry type.
func (f *ServiceFactory) CreateManager(reg *models.Registry) (ServiceManager, error) {
	// Validate registry
	if err := reg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid registry: %w", err)
	}

	// Get strategy for registry type
	strategy, err := f.GetStrategy(reg.Type)
	if err != nil {
		return nil, err
	}

	// Validate registry-specific config
	if reg.Config.Valid && reg.Config.String != "" {
		if err := strategy.ValidateConfig([]byte(reg.Config.String)); err != nil {
			return nil, fmt.Errorf("invalid config for %s registry: %w", reg.Type, err)
		}
	}

	// Create manager using strategy
	return strategy.CreateManager(reg)
}

// GetDefaultPort returns the default port for a registry type.
func (f *ServiceFactory) GetDefaultPort(registryType string) (int, error) {
	strategy, err := f.GetStrategy(registryType)
	if err != nil {
		return 0, err
	}
	return strategy.GetDefaultPort(), nil
}

// GetDefaultStorage returns the default storage path for a registry type.
func (f *ServiceFactory) GetDefaultStorage(registryType string) (string, error) {
	strategy, err := f.GetStrategy(registryType)
	if err != nil {
		return "", err
	}
	return strategy.GetDefaultStorage(), nil
}

// SupportedTypes returns a list of all supported registry types.
func (f *ServiceFactory) SupportedTypes() []string {
	types := make([]string, 0, len(f.strategies))
	for t := range f.strategies {
		types = append(types, t)
	}
	return types
}

// DetectVersion returns the installed version for a registry.
// It first checks reg.Version, then the strategy default, then detects from the binary.
func (f *ServiceFactory) DetectVersion(ctx context.Context, reg *models.Registry) string {
	// If version is already set on the model, use it
	if reg.Version != "" {
		return reg.Version
	}

	// Check strategy default
	strategy, err := f.GetStrategy(reg.Type)
	if err != nil {
		return ""
	}
	if v := strategy.GetDefaultVersion(); v != "" {
		return v
	}

	// Detect from binary manager based on registry type
	bm := f.createBinaryManager(reg)
	if bm == nil {
		return ""
	}
	version, err := bm.GetVersion(ctx)
	if err != nil {
		return ""
	}
	return version
}

// createBinaryManager creates the appropriate BinaryManager for a registry type.
func (f *ServiceFactory) createBinaryManager(reg *models.Registry) BinaryManager {
	storagePath, err := resolveStoragePath(reg, "storage")
	if err != nil {
		return nil
	}
	binDir := filepath.Join(storagePath, "bin")

	switch reg.Type {
	case "zot":
		version := reg.Version
		if version == "" {
			if s, err := f.GetStrategy("zot"); err == nil {
				version = s.GetDefaultVersion()
			}
		}
		return NewBinaryManager(binDir, version)
	case "athens":
		return NewAthensBinaryManager(storagePath, "0.14.1")
	case "devpi":
		return NewPipxBinaryManager("devpi-server", "")
	case "verdaccio":
		return NewNpmBinaryManager("verdaccio", "")
	case "squid":
		return NewSquidBinaryManager()
	default:
		return nil
	}
}

// pidFileNames maps registry types to their PID file names.
var pidFileNames = map[string]string{
	"zot":       "zot.pid",
	"athens":    "athens.pid",
	"devpi":     "devpi.pid",
	"verdaccio": "verdaccio.pid",
	"squid":     "squid.pid",
}

// GetUptime returns the uptime for a running registry by checking the PID file modification time.
// Returns 0 if the PID file cannot be found or stat'd.
func (f *ServiceFactory) GetUptime(reg *models.Registry) time.Duration {
	pidName, ok := pidFileNames[reg.Type]
	if !ok {
		return 0
	}

	// Resolve storage path — mirrors the logic in each strategy's CreateManager.
	// For most registries, the PID file is at {storage}/{type}.pid.
	// For squid, storage resolution uses "cacheDir" key and may need parent dir.
	var storagePath string
	var err error
	if reg.Type == "squid" {
		storagePath, err = resolveSquidStoragePath(reg)
	} else {
		storagePath, err = resolveStoragePath(reg, "storage")
	}
	if err != nil {
		return 0
	}

	pidFile := filepath.Join(storagePath, pidName)
	info, err := os.Stat(pidFile)
	if err != nil {
		return 0
	}
	return time.Since(info.ModTime())
}

// resolveSquidStoragePath resolves the storage path for squid, mirroring SquidStrategy.CreateManager.
func resolveSquidStoragePath(reg *models.Registry) (string, error) {
	pc, err := paths.Default()
	if err != nil {
		return "", err
	}
	defaultPath := pc.RegistryDir(reg.Name)
	storagePath, err := resolveStoragePath(reg, "cacheDir")
	if err != nil {
		return "", err
	}
	if storagePath != defaultPath {
		// cacheDir was found in config — use its parent as storage root
		storagePath = filepath.Dir(storagePath)
	}
	return storagePath, nil
}
