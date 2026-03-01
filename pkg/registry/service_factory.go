package registry

import (
	"fmt"

	"devopsmaestro/models"
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
