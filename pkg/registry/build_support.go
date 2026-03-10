package registry

import (
	"context"
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/registry/envinjector"
)

// ManagerFactory creates ServiceManager instances from Registry models.
// ServiceFactory implements this interface.
type ManagerFactory interface {
	CreateManager(reg *models.Registry) (ServiceManager, error)
}

// BuildRegistryCoordinator prepares registries for build operations.
type BuildRegistryCoordinator struct {
	store    db.RegistryStore
	factory  ManagerFactory
	injector *envinjector.EnvironmentInjector
}

// BuildRegistryResult contains the output of registry preparation for builds.
type BuildRegistryResult struct {
	Managers    []ServiceManager
	Registries  []*models.Registry
	EnvVars     map[string]string
	OCIEndpoint string
	Warnings    []string
}

// NewBuildRegistryCoordinator creates a new BuildRegistryCoordinator.
func NewBuildRegistryCoordinator(store db.RegistryStore, factory ManagerFactory, injector *envinjector.EnvironmentInjector) *BuildRegistryCoordinator {
	return &BuildRegistryCoordinator{
		store:    store,
		factory:  factory,
		injector: injector,
	}
}

// Prepare prepares all eligible registries for a build operation.
// It lists enabled registries, filters by auto-startable lifecycle,
// starts each one, and returns the collected managers, env vars, and OCI endpoint.
func (c *BuildRegistryCoordinator) Prepare(ctx context.Context) (*BuildRegistryResult, error) {
	registries, err := c.store.ListRegistries()
	if err != nil {
		return nil, err
	}

	result := &BuildRegistryResult{
		EnvVars: make(map[string]string),
	}

	for _, reg := range registries {
		if !reg.Enabled {
			continue
		}
		if reg.Lifecycle == "manual" {
			continue
		}

		manager, err := c.factory.CreateManager(reg)
		if err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("registry '%s' failed to start: %v", reg.Name, err))
			continue
		}

		if err := manager.Start(ctx); err != nil {
			result.Warnings = append(result.Warnings, fmt.Sprintf("registry '%s' failed to start: %v", reg.Name, err))
			continue
		}

		result.Managers = append(result.Managers, manager)
		result.Registries = append(result.Registries, reg)

		for k, v := range c.injector.InjectForBuild(reg) {
			result.EnvVars[k] = v
		}

		if reg.Type == "zot" {
			result.OCIEndpoint = manager.GetEndpoint()
		}
	}

	return result, nil
}
