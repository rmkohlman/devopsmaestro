package preflight

import (
	"context"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/registry"
)

// RegistryCheck verifies registry health and availability
type RegistryCheck struct {
	store   db.DataStore
	manager registry.RegistryManager
}

// NewRegistryCheck creates a new RegistryCheck
func NewRegistryCheck(store db.DataStore, manager registry.RegistryManager) *RegistryCheck {
	return &RegistryCheck{
		store:   store,
		manager: manager,
	}
}

// Name returns the check name
func (rc *RegistryCheck) Name() string {
	return "Registry Health"
}

// Run executes the registry health check
func (rc *RegistryCheck) Run(ctx context.Context) CheckResult {
	// Get all registries from database
	registries, err := rc.store.ListRegistries()
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: "Failed to retrieve registries: " + err.Error(),
		}
	}

	// Check if no registries configured
	if len(registries) == 0 {
		return CheckResult{
			Status:  StatusWarning,
			Message: "No registries configured",
			Details: map[string]interface{}{
				"recommendation": "Configure at least one OCI registry for image caching",
			},
		}
	}

	// Get enabled registries (those with defaults set)
	defaults, err := rc.getEnabledRegistries(ctx)
	if err != nil {
		return CheckResult{
			Status:  StatusError,
			Message: "Failed to check registry defaults: " + err.Error(),
		}
	}

	// Check if no OCI registry is set as default
	hasOCI := false
	for regType := range defaults {
		if regType == "oci" || regType == "zot" {
			hasOCI = true
			break
		}
	}

	// Track registry health
	details := make(map[string]interface{})
	allHealthy := true
	hasWarnings := false
	statusMessages := []string{}

	// Check each enabled registry
	for regType, regName := range defaults {
		reg, err := rc.store.GetRegistryByName(regName)
		if err != nil {
			statusMessages = append(statusMessages, regType+" ("+regName+"): not found")
			allHealthy = false
			continue
		}

		// Check health based on lifecycle
		status := rc.checkRegistryStatus(ctx, reg)
		statusMessages = append(statusMessages, status.message)

		if status.isError {
			allHealthy = false
		}
		if status.isWarning {
			hasWarnings = true
		}
	}

	details["registries"] = statusMessages

	// Determine overall status
	if !allHealthy {
		return CheckResult{
			Status:  StatusError,
			Message: "One or more enabled registries are unhealthy",
			Details: details,
		}
	}

	if !hasOCI {
		return CheckResult{
			Status:  StatusWarning,
			Message: "No OCI registry configured as default",
			Details: details,
		}
	}

	if hasWarnings {
		return CheckResult{
			Status:  StatusWarning,
			Message: "All registries operational with warnings",
			Details: details,
		}
	}

	return CheckResult{
		Status:  StatusOK,
		Message: "All enabled registries are healthy",
		Details: details,
	}
}

// registryCheckStatus represents the result of checking a single registry
type registryCheckStatus struct {
	message   string
	isError   bool
	isWarning bool
}

// checkRegistryStatus checks the health of a single registry
func (rc *RegistryCheck) checkRegistryStatus(ctx context.Context, reg *models.Registry) registryCheckStatus {
	// For manual lifecycle registries, we skip health checks
	if reg.Lifecycle == "manual" {
		if rc.manager.IsRunning(ctx) {
			return registryCheckStatus{
				message:   reg.Type + " (" + reg.Name + "): running (manual)",
				isError:   false,
				isWarning: false,
			}
		}
		return registryCheckStatus{
			message:   reg.Type + " (" + reg.Name + "): stopped (manual - skipped)",
			isError:   false,
			isWarning: false,
		}
	}

	// For auto/persistent/on-demand registries, auto-start if needed
	if rc.shouldAutoStart(reg) {
		if !rc.manager.IsRunning(ctx) {
			// Try to start the registry
			if err := rc.manager.EnsureRunning(ctx); err != nil {
				return registryCheckStatus{
					message:   reg.Type + " (" + reg.Name + "): failed to auto-start - " + err.Error(),
					isError:   true,
					isWarning: false,
				}
			}
		}
	}

	// Check if registry is running
	if !rc.manager.IsRunning(ctx) {
		return registryCheckStatus{
			message:   reg.Type + " (" + reg.Name + "): not running",
			isError:   true,
			isWarning: false,
		}
	}

	// Check health via Status call
	status, err := rc.manager.Status(ctx)
	if err != nil {
		return registryCheckStatus{
			message:   reg.Type + " (" + reg.Name + "): health check failed - " + err.Error(),
			isError:   false,
			isWarning: true,
		}
	}

	if status.State != "running" {
		return registryCheckStatus{
			message:   reg.Type + " (" + reg.Name + "): unhealthy (state: " + status.State + ")",
			isError:   true,
			isWarning: false,
		}
	}

	return registryCheckStatus{
		message:   reg.Type + " (" + reg.Name + "): healthy",
		isError:   false,
		isWarning: false,
	}
}

// shouldAutoStart determines if a registry should be auto-started
func (rc *RegistryCheck) shouldAutoStart(reg *models.Registry) bool {
	// Auto-start for auto, persistent, and on-demand lifecycles
	return reg.Lifecycle == "auto" ||
		reg.Lifecycle == "persistent" ||
		reg.Lifecycle == "on-demand"
}

// getEnabledRegistries returns a map of registry types to registry names that have defaults set
func (rc *RegistryCheck) getEnabledRegistries(ctx context.Context) (map[string]string, error) {
	defaults := registry.NewRegistryDefaults(rc.store)
	allDefaults, err := defaults.GetAllDefaults(ctx)
	if err != nil {
		return nil, err
	}

	enabled := make(map[string]string)
	for key, value := range allDefaults {
		if value != "" && key != registry.DefaultKeyIdleTimeout {
			// Convert key like "registry-oci" to "oci"
			regType := key
			if len(key) > 9 && key[:9] == "registry-" {
				regType = key[9:]
			}
			enabled[regType] = value
		}
	}

	return enabled, nil
}
