package registry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/registry/envinjector"

	"github.com/rmkohlman/MaestroSDK/paths"
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

	// BuildKitConfigPath is the path to a generated buildkitd.toml that
	// configures the Zot registry as a pull-through mirror for buildx builds.
	// Empty if Zot is not running or config generation failed.
	BuildKitConfigPath string

	// ContainerdCertsDir is the path to a generated certs.d directory that
	// configures the Zot registry as a pull-through mirror for containerd/nerdctl.
	// Empty if Zot is not running or config generation failed.
	ContainerdCertsDir string
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
			// Distinguish "binary not installed" from "binary installed but failed to start".
			// ErrBinaryNotInstalled = user needs to install; use slog.Info (actionable, not alarming).
			// Other errors = binary exists but something else failed; use slog.Warn (unexpected).
			if errors.Is(err, ErrBinaryNotInstalled) {
				slog.Info("registry binary not installed",
					"registry", reg.Name, "type", reg.Type, "error", err)
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("⚠️  %s proxy not available. Install with: brew install %s for faster builds",
						reg.Name, reg.Type))
			} else {
				slog.Warn("registry failed to start",
					"registry", reg.Name, "type", reg.Type, "error", err)
				result.Warnings = append(result.Warnings,
					fmt.Sprintf("registry '%s' installed but failed to start: %v", reg.Name, err))
			}
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

	// Generate BuildKit and containerd mirror configs if Zot is available
	if result.OCIEndpoint != "" {
		c.generateMirrorConfigs(result)
	}

	return result, nil
}

// generateMirrorConfigs generates BuildKit and containerd mirror configuration
// files so that image pulls are routed through the local Zot registry.
// Failures are non-fatal — warnings are appended to the result.
func (c *BuildRegistryCoordinator) generateMirrorConfigs(result *BuildRegistryResult) {
	endpoint := EndpointFromURL(result.OCIEndpoint)
	mirrors := defaultMirrors()

	// Resolve output base dir from paths
	var outputBase string
	if pc, err := paths.Default(); err == nil {
		outputBase = pc.Root()
	} else {
		slog.Warn("mirror config: cannot resolve paths, skipping config generation", "error", err)
		return
	}

	// Generate buildkitd.toml
	bkGen := NewBuildKitConfigGenerator()
	bkDir := filepath.Join(outputBase, "buildkit")
	bkPath, err := bkGen.Generate(endpoint, mirrors, bkDir)
	if err != nil {
		slog.Warn("failed to generate buildkitd.toml", "error", err)
		result.Warnings = append(result.Warnings, fmt.Sprintf("buildkitd.toml generation failed: %v", err))
	} else {
		result.BuildKitConfigPath = bkPath
		slog.Info("generated buildkitd.toml", "path", bkPath)
	}

	// Generate containerd hosts.toml files
	ctGen := NewContainerdConfigGenerator()
	ctDir := filepath.Join(outputBase, "containerd")
	ctPath, err := ctGen.Generate(endpoint, mirrors, ctDir)
	if err != nil {
		slog.Warn("failed to generate containerd hosts.toml", "error", err)
		result.Warnings = append(result.Warnings, fmt.Sprintf("containerd hosts.toml generation failed: %v", err))
	} else {
		result.ContainerdCertsDir = ctPath
		slog.Info("generated containerd hosts.toml", "path", ctPath)
	}
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
// When registries are unhealthy, the message includes a warning emoji
// to make the degraded state visually prominent in build output.
func (r *CacheReadiness) FormatSummary() string {
	if r.TotalEnabled == 0 {
		return "Cache: no registries enabled"
	}
	if r.AllHealthy {
		return fmt.Sprintf("✅ Cache: %d/%d registries active", r.HealthyCount, r.TotalEnabled)
	}
	return fmt.Sprintf("⚠️  Cache: %d/%d registries active (%s failed)",
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
