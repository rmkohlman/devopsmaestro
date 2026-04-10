package registry

import (
	"context"
	"fmt"
	"log/slog"

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
	Managers       []ServiceManager
	Registries     []*models.Registry
	EnvVars        map[string]string
	OCIEndpoint    string
	Warnings       []string
	CacheReadiness CacheReadiness
}

// CacheReadiness reports aggregate health status of build cache registries.
type CacheReadiness struct {
	AllHealthy   bool
	HealthyCount int
	TotalEnabled int
	Started      []string // names of registries that started successfully
	StartedTypes []string // types of registries that started successfully (e.g. "zot", "devpi")
	Unhealthy    []string // names of registries that failed to start
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

		result.CacheReadiness.TotalEnabled++

		manager, err := c.factory.CreateManager(reg)
		if err != nil {
			slog.Warn("registry failed to create manager",
				"registry", reg.Name, "type", reg.Type, "error", err)
			result.Warnings = append(result.Warnings, fmt.Sprintf("registry '%s' failed to start: %v", reg.Name, err))
			result.CacheReadiness.Unhealthy = append(result.CacheReadiness.Unhealthy, reg.Name)
			continue
		}

		if err := manager.Start(ctx); err != nil {
			slog.Warn("registry failed to start",
				"registry", reg.Name, "type", reg.Type, "error", err)
			result.Warnings = append(result.Warnings, fmt.Sprintf("registry '%s' failed to start: %v", reg.Name, err))
			result.CacheReadiness.Unhealthy = append(result.CacheReadiness.Unhealthy, reg.Name)
			continue
		}

		result.Managers = append(result.Managers, manager)
		result.Registries = append(result.Registries, reg)
		result.CacheReadiness.HealthyCount++
		result.CacheReadiness.Started = append(result.CacheReadiness.Started, reg.Name)
		result.CacheReadiness.StartedTypes = append(result.CacheReadiness.StartedTypes, reg.Type)

		for k, v := range c.injector.InjectForBuild(reg) {
			result.EnvVars[k] = v
		}

		if reg.Type == "zot" {
			result.OCIEndpoint = manager.GetEndpoint()
		}
	}

	result.CacheReadiness.AllHealthy = result.CacheReadiness.HealthyCount == result.CacheReadiness.TotalEnabled

	return result, nil
}

// EnsureCachesReady returns the readiness status of all enabled registries.
// It does NOT abort the build — it only reports which registries started
// successfully and which failed. Use this for cache status summaries.
func (c *BuildRegistryCoordinator) EnsureCachesReady(ctx context.Context) (*CacheReadiness, error) {
	registries, err := c.store.ListRegistries()
	if err != nil {
		return nil, fmt.Errorf("failed to list registries: %w", err)
	}

	readiness := &CacheReadiness{}

	for _, reg := range registries {
		if !reg.Enabled || reg.Lifecycle == "manual" {
			continue
		}
		readiness.TotalEnabled++

		manager, err := c.factory.CreateManager(reg)
		if err != nil {
			slog.Warn("cache readiness: failed to create manager",
				"registry", reg.Name, "error", err)
			readiness.Unhealthy = append(readiness.Unhealthy, reg.Name)
			continue
		}

		if err := manager.Start(ctx); err != nil {
			slog.Warn("cache readiness: registry not available",
				"registry", reg.Name, "error", err)
			readiness.Unhealthy = append(readiness.Unhealthy, reg.Name)
			continue
		}

		readiness.HealthyCount++
	}

	readiness.AllHealthy = readiness.HealthyCount == readiness.TotalEnabled
	return readiness, nil
}

// IsStarted reports whether a specific registry (by name) started successfully.
func (r *CacheReadiness) IsStarted(name string) bool {
	for _, s := range r.Started {
		if s == name {
			return true
		}
	}
	return false
}

// IsTypeStarted reports whether any registry of the given type started successfully.
func (r *CacheReadiness) IsTypeStarted(regType string) bool {
	for _, t := range r.StartedTypes {
		if t == regType {
			return true
		}
	}
	return false
}

// FormatSummary returns a human-readable cache readiness summary line.
// Example: "Cache: 4/5 registries active (devpi failed)"
func (r *CacheReadiness) FormatSummary() string {
	if r.TotalEnabled == 0 {
		return "Cache: no registries enabled"
	}
	if r.AllHealthy {
		return fmt.Sprintf("Cache: %d/%d registries active", r.HealthyCount, r.TotalEnabled)
	}
	return fmt.Sprintf("Cache: %d/%d registries active (%s failed)",
		r.HealthyCount, r.TotalEnabled, joinNames(r.Unhealthy))
}

// joinNames joins registry names with commas for display.
func joinNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	result := names[0]
	for _, n := range names[1:] {
		result += ", " + n
	}
	return result
}
