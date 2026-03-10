package registry

import (
	"fmt"
	"path/filepath"
)

// NewRegistryManager creates a RegistryManager with the given config.
// Returns an error if the config is invalid.
func NewRegistryManager(config RegistryConfig) (RegistryManager, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid registry config: %w", err)
	}
	return NewZotManager(config), nil
}

// NewZotManager creates a ZotManager with injected dependencies.
func NewZotManager(config RegistryConfig) *ZotManager {
	binDir := filepath.Join(config.Storage, "bin")
	return NewZotManagerWithDeps(
		config,
		NewBinaryManager(binDir, "2.0.0"),
		NewProcessManager(ProcessConfig{
			PIDFile: filepath.Join(config.Storage, "zot.pid"),
			LogFile: filepath.Join(config.Storage, "zot.log"),
		}),
	)
}

// NewZotManagerWithDeps creates a ZotManager with explicit dependency injection.
// This is useful for testing with mock dependencies.
func NewZotManagerWithDeps(config RegistryConfig, binaryMgr BinaryManager, processMgr ProcessManager) *ZotManager {
	return &ZotManager{
		BaseServiceManager: NewBaseServiceManager(binaryMgr, processMgr),
		config:             config,
	}
}

// NewBinaryManager creates a BinaryManager for the specified version.
func NewBinaryManager(binDir, version string) BinaryManager {
	return &DefaultBinaryManager{
		binDir:  binDir,
		version: version,
	}
}

// NewProcessManager creates a ProcessManager with the given config.
func NewProcessManager(config ProcessConfig) ProcessManager {
	return &DefaultProcessManager{
		config: config,
	}
}
