package registry

import (
	"fmt"
	"path/filepath"
)

// NewGoModuleProxy creates a new GoModuleProxy implementation based on the configuration.
// Currently only supports Athens as the implementation.
func NewGoModuleProxy(config GoModuleConfig) GoModuleProxy {
	// Validate config
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid Go module proxy config: %v", err))
	}

	// For now, always return AthensManager
	// In the future, we could support other proxy implementations
	return NewAthensManager(config)
}

// newAthensManagerInternal creates a new AthensManager with default dependencies.
// This is an internal helper for the factory.
func newAthensManagerInternal(config GoModuleConfig) *AthensManager {
	binaryManager := NewAthensBinaryManager(config.Storage, "0.14.1")
	processManager := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(config.Storage, "athens.pid"),
		LogFile: filepath.Join(config.Storage, "athens.log"),
	})

	return NewAthensManagerWithDeps(config, binaryManager, processManager)
}
