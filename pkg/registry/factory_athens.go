package registry

import (
	"fmt"
	"path/filepath"
)

// NewGoModuleProxy creates a new GoModuleProxy implementation based on the configuration.
// Currently only supports Athens as the implementation.
// Returns an error if the config is invalid.
func NewGoModuleProxy(config GoModuleConfig) (GoModuleProxy, error) {
	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid Go module proxy config: %w", err)
	}

	// Construct dependencies explicitly and use canonical DI constructor.
	binaryManager := NewAthensBinaryManager(config.Storage, "0.14.1")
	processManager := NewProcessManager(ProcessConfig{
		PIDFile: filepath.Join(config.Storage, "athens.pid"),
		LogFile: filepath.Join(config.Storage, "athens.log"),
	})

	mgr, err := NewAthensManager(config, binaryManager, processManager)
	if err != nil {
		return nil, err
	}
	return mgr, nil
}
