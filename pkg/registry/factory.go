package registry

import (
	"fmt"
	"path/filepath"
)

// NewRegistryManager creates a RegistryManager with the given config.
// Panics if config is invalid (fail-fast for programming errors).
func NewRegistryManager(config RegistryConfig) RegistryManager {
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid registry config: %v", err))
	}
	return NewZotManager(config)
}

// NewZotManager creates a ZotManager with injected dependencies.
func NewZotManager(config RegistryConfig) *ZotManager {
	binDir := filepath.Join(config.Storage, "bin")
	return NewZotManagerWithDeps(
		config,
		NewBinaryManager(binDir, "1.4.3"),
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
		config:         config,
		binaryManager:  binaryMgr,
		processManager: processMgr,
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
