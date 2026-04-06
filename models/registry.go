package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

// Registry represents a package registry (zot, athens, devpi, verdaccio, squid)
type Registry struct {
	ID          int
	Name        string
	Type        string // zot, athens, devpi, verdaccio, squid
	Version     string // Desired binary version (e.g., "2.1.15")
	Enabled     bool   // Whether the registry is enabled
	Port        int
	Lifecycle   string // persistent, on-demand, manual
	Storage     string // Storage path for registry data
	IdleTimeout int    // Idle timeout in seconds (for on-demand registries)
	Description sql.NullString
	Config      sql.NullString // JSON config specific to registry type
	Status      string         // stopped, starting, running, error
	CreatedAt   string
	UpdatedAt   string
}

// RegistryYAML represents the YAML structure for a Registry resource
type RegistryYAML struct {
	APIVersion string              `yaml:"apiVersion" json:"apiVersion"`
	Kind       string              `yaml:"kind" json:"kind"`
	Metadata   RegistryMetadata    `yaml:"metadata" json:"metadata"`
	Spec       RegistrySpec        `yaml:"spec" json:"spec"`
	Status     *RegistryStatusYAML `yaml:"status,omitempty" json:"status,omitempty"`
}

// RegistryStatusYAML represents the live status section of a Registry resource
type RegistryStatusYAML struct {
	State    string `yaml:"state" json:"state"`
	Endpoint string `yaml:"endpoint" json:"endpoint"`
}

type RegistryMetadata struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

type RegistrySpec struct {
	Type        string                 `yaml:"type"`
	Version     string                 `yaml:"version"`
	Enabled     *bool                  `yaml:"enabled,omitempty"`
	Port        int                    `yaml:"port,omitempty"`
	Lifecycle   string                 `yaml:"lifecycle,omitempty"`
	Storage     string                 `yaml:"storage,omitempty"`
	IdleTimeout int                    `yaml:"idleTimeout,omitempty"`
	Config      map[string]interface{} `yaml:"config,omitempty"`
}

// Minimum idle timeout for on-demand registries (seconds).
// Values below this threshold cause rapid start/stop cycling.
const minIdleTimeout = 60

// Valid registry types
var validRegistryTypes = map[string]bool{
	"zot":       true,
	"athens":    true,
	"devpi":     true,
	"verdaccio": true,
	"squid":     true,
}

// Valid lifecycle values
var validLifecycles = map[string]bool{
	"persistent": true,
	"on-demand":  true,
	"manual":     true,
}

// Default ports for each registry type
var defaultPorts = map[string]int{
	"zot":       5001,
	"athens":    3000,
	"devpi":     3141,
	"verdaccio": 4873,
	"squid":     3128,
}

// Default storage paths for each registry type
var defaultStorage = map[string]string{
	"zot":       "/var/lib/zot",
	"athens":    "/var/lib/athens",
	"devpi":     "/var/lib/devpi",
	"verdaccio": "/var/lib/verdaccio",
	"squid":     "/var/cache/squid",
}

// ToYAML converts Registry to RegistryYAML
func (r *Registry) ToYAML() RegistryYAML {
	enabled := r.Enabled
	yaml := RegistryYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Registry",
		Metadata: RegistryMetadata{
			Name: r.Name,
		},
		Spec: RegistrySpec{
			Type:        r.Type,
			Version:     r.Version,
			Enabled:     &enabled,
			Port:        r.Port,
			Lifecycle:   r.Lifecycle,
			Storage:     r.Storage,
			IdleTimeout: r.IdleTimeout,
		},
	}

	if r.Description.Valid {
		yaml.Metadata.Description = r.Description.String
	}

	if r.Config.Valid && r.Config.String != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(r.Config.String), &config); err == nil {
			yaml.Spec.Config = config
		}
	}

	return yaml
}

// FromYAML populates Registry from RegistryYAML
func (r *Registry) FromYAML(yaml RegistryYAML) {
	r.Name = yaml.Metadata.Name
	r.Type = yaml.Spec.Type
	r.Version = yaml.Spec.Version
	r.Port = yaml.Spec.Port
	r.Lifecycle = yaml.Spec.Lifecycle
	r.Storage = yaml.Spec.Storage
	r.IdleTimeout = yaml.Spec.IdleTimeout

	// Default Enabled=true when the field is omitted from YAML.
	// A nil pointer means the field was absent; Go's bool zero value (false)
	// would silently disable all registries restored from legacy YAML.
	if yaml.Spec.Enabled == nil {
		r.Enabled = true
	} else {
		r.Enabled = *yaml.Spec.Enabled
	}

	if yaml.Metadata.Description != "" {
		r.Description = sql.NullString{String: yaml.Metadata.Description, Valid: true}
	}

	if len(yaml.Spec.Config) > 0 {
		configJSON, err := json.Marshal(yaml.Spec.Config)
		if err == nil {
			r.Config = sql.NullString{String: string(configJSON), Valid: true}
		}
	}

	// Apply defaults if not specified
	if r.Lifecycle == "" {
		r.Lifecycle = "manual"
	}
	if r.Status == "" {
		r.Status = "stopped"
	}
}

// ValidateType checks if the registry type is valid
func (r *Registry) ValidateType() error {
	if r.Type == "" {
		return fmt.Errorf("type is required")
	}
	if !validRegistryTypes[r.Type] {
		return fmt.Errorf("unsupported registry type: %s (valid types: zot, athens, devpi, verdaccio, squid)", r.Type)
	}
	return nil
}

// ValidatePort checks if the port is in valid range (1024-65535)
func (r *Registry) ValidatePort() error {
	// Port 0 means auto-assign, which is allowed
	if r.Port == 0 {
		return nil
	}
	if r.Port < 1024 || r.Port > 65535 {
		return fmt.Errorf("port must be between 1024 and 65535 (got %d)", r.Port)
	}
	return nil
}

// ValidateLifecycle checks if the lifecycle is valid
func (r *Registry) ValidateLifecycle() error {
	// Empty lifecycle defaults to manual
	if r.Lifecycle == "" {
		return nil
	}
	if !validLifecycles[r.Lifecycle] {
		return fmt.Errorf("unsupported lifecycle: %s (valid: persistent, on-demand, manual)", r.Lifecycle)
	}
	return nil
}

// ValidateVersion checks if the version string is valid semver (without leading 'v').
// Empty string is valid (means "use strategy default").
func (r *Registry) ValidateVersion() error {
	if r.Version == "" {
		return nil
	}
	// Light semver regex: MAJOR.MINOR.PATCH with optional pre-release
	matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+(-[a-zA-Z0-9]+(\.[a-zA-Z0-9]+)*)?$`, r.Version)
	if !matched {
		return fmt.Errorf("invalid version %q: must be semver format (e.g., 2.1.15, 1.0.0-rc1)", r.Version)
	}
	return nil
}

// Validate performs all validation checks
func (r *Registry) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if err := r.ValidateType(); err != nil {
		return err
	}

	if err := r.ValidatePort(); err != nil {
		return err
	}

	if err := r.ValidateLifecycle(); err != nil {
		return err
	}

	if err := r.ValidateVersion(); err != nil {
		return err
	}

	// Minimum IdleTimeout for on-demand registries: values between 1 and
	// minIdleTimeout-1 would cause rapid start/stop cycling. Zero means
	// "use default" (ApplyDefaults sets it to 1800s) and is always allowed.
	if r.IsOnDemand() && r.IdleTimeout > 0 && r.IdleTimeout < minIdleTimeout {
		return fmt.Errorf("IdleTimeout must be at least %d seconds for on-demand registries (got %d)", minIdleTimeout, r.IdleTimeout)
	}

	// Validate storage path is not empty
	if r.Storage == "" {
		return fmt.Errorf("storage path is required (NOT NULL constraint)")
	}

	// Validate config JSON if present
	if r.Config.Valid && r.Config.String != "" {
		var config map[string]interface{}
		if err := json.Unmarshal([]byte(r.Config.String), &config); err != nil {
			return fmt.Errorf("invalid config JSON: %w", err)
		}
	}

	return nil
}

// GetDefaultPort returns the default port for the registry type
func (r *Registry) GetDefaultPort() int {
	if port, ok := defaultPorts[r.Type]; ok {
		return port
	}
	return 0
}

// GetDefaultStorage returns the default storage path for the registry type
func (r *Registry) GetDefaultStorage() string {
	if storage, ok := defaultStorage[r.Type]; ok {
		return storage
	}
	return ""
}

// IsOnDemand returns true if this registry uses on-demand lifecycle
func (r *Registry) IsOnDemand() bool {
	return r.Lifecycle == "on-demand"
}

// ShouldStopAfterIdle returns true if registry should auto-stop after idle timeout
func (r *Registry) ShouldStopAfterIdle() bool {
	return r.IsOnDemand() && r.IdleTimeout > 0
}

// GetIdleTimeoutDuration returns idle timeout as time.Duration
func (r *Registry) GetIdleTimeoutDuration() time.Duration {
	return time.Duration(r.IdleTimeout) * time.Second
}

// ApplyDefaults sets default values for Port, Storage, and IdleTimeout
func (r *Registry) ApplyDefaults() {
	// Apply default port if not set
	if r.Port == 0 {
		if port, ok := defaultPorts[r.Type]; ok {
			r.Port = port
		}
	}

	// Apply default storage if not set
	if r.Storage == "" {
		if storage, ok := defaultStorage[r.Type]; ok {
			r.Storage = storage
		}
	}

	// Apply default idle timeout for on-demand registries
	if r.IsOnDemand() && r.IdleTimeout == 0 {
		r.IdleTimeout = 1800 // 30 minutes default
	}
}
