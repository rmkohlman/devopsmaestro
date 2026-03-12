package models

import (
	"devopsmaestro/config"
	"fmt"
	"time"
)

// CredentialScopeType represents the level at which a credential is defined
type CredentialScopeType string

const (
	CredentialScopeEcosystem CredentialScopeType = "ecosystem"
	CredentialScopeDomain    CredentialScopeType = "domain"
	CredentialScopeApp       CredentialScopeType = "app"
	CredentialScopeWorkspace CredentialScopeType = "workspace"
)

// CredentialDB represents a credential configuration stored in the database
type CredentialDB struct {
	ID          int64               `db:"id" json:"id"`
	ScopeType   CredentialScopeType `db:"scope_type" json:"scope_type"`
	ScopeID     int64               `db:"scope_id" json:"scope_id"`
	Name        string              `db:"name" json:"name"`
	Source      string              `db:"source" json:"source"`   // "keychain", "env"
	Service     *string             `db:"service" json:"service"` // Keychain service name
	EnvVar      *string             `db:"env_var" json:"env_var"` // Environment variable name
	Description *string             `db:"description" json:"description"`
	CreatedAt   time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time           `db:"updated_at" json:"updated_at"`
}

// ToConfig converts a CredentialDB to a config.CredentialConfig
func (c *CredentialDB) ToConfig() config.CredentialConfig {
	cfg := config.CredentialConfig{
		Source: config.CredentialSource(c.Source),
	}
	if c.Service != nil {
		cfg.Service = *c.Service
	}
	if c.EnvVar != nil {
		cfg.EnvVar = *c.EnvVar
	}
	return cfg
}

// CredentialsToMap converts a slice of CredentialDB to a config.Credentials map
func CredentialsToMap(creds []*CredentialDB) config.Credentials {
	result := make(config.Credentials)
	for _, c := range creds {
		result[c.Name] = c.ToConfig()
	}
	return result
}

// =============================================================================
// Credential YAML Types
// =============================================================================

// CredentialYAML represents the YAML structure for a Credential resource
type CredentialYAML struct {
	APIVersion string             `yaml:"apiVersion" json:"apiVersion"`
	Kind       string             `yaml:"kind" json:"kind"`
	Metadata   CredentialMetadata `yaml:"metadata" json:"metadata"`
	Spec       CredentialSpec     `yaml:"spec" json:"spec"`
}

// CredentialMetadata contains credential metadata including scope hierarchy
type CredentialMetadata struct {
	Name      string `yaml:"name" json:"name"`
	Ecosystem string `yaml:"ecosystem,omitempty" json:"ecosystem,omitempty"`
	Domain    string `yaml:"domain,omitempty" json:"domain,omitempty"`
	App       string `yaml:"app,omitempty" json:"app,omitempty"`
	Workspace string `yaml:"workspace,omitempty" json:"workspace,omitempty"`
}

// CredentialSpec contains the credential specification
type CredentialSpec struct {
	Source      string `yaml:"source" json:"source"`
	Service     string `yaml:"service,omitempty" json:"service,omitempty"`
	EnvVar      string `yaml:"envVar,omitempty" json:"envVar,omitempty"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
}

// ScopeInfo returns the scope type and scope name from metadata
func (m *CredentialMetadata) ScopeInfo() (CredentialScopeType, string) {
	switch {
	case m.Ecosystem != "":
		return CredentialScopeEcosystem, m.Ecosystem
	case m.Domain != "":
		return CredentialScopeDomain, m.Domain
	case m.App != "":
		return CredentialScopeApp, m.App
	case m.Workspace != "":
		return CredentialScopeWorkspace, m.Workspace
	default:
		return "", ""
	}
}

// ToYAML converts a CredentialDB to CredentialYAML.
// scopeName is the name of the scope target (ecosystem name, domain name, etc.)
func (c *CredentialDB) ToYAML(scopeName string) CredentialYAML {
	y := CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata: CredentialMetadata{
			Name: c.Name,
		},
		Spec: CredentialSpec{
			Source: c.Source,
		},
	}

	// Set scope in metadata based on scope type
	switch c.ScopeType {
	case CredentialScopeEcosystem:
		y.Metadata.Ecosystem = scopeName
	case CredentialScopeDomain:
		y.Metadata.Domain = scopeName
	case CredentialScopeApp:
		y.Metadata.App = scopeName
	case CredentialScopeWorkspace:
		y.Metadata.Workspace = scopeName
	}

	// Set optional fields from pointers
	if c.Service != nil {
		y.Spec.Service = *c.Service
	}
	if c.EnvVar != nil {
		y.Spec.EnvVar = *c.EnvVar
	}
	if c.Description != nil {
		y.Spec.Description = *c.Description
	}

	return y
}

// FromYAML populates a CredentialDB from CredentialYAML.
// Note: ScopeType and ScopeID must be resolved separately by the caller.
func (c *CredentialDB) FromYAML(y CredentialYAML) {
	c.Name = y.Metadata.Name
	c.Source = y.Spec.Source

	if y.Spec.Service != "" {
		s := y.Spec.Service
		c.Service = &s
	}
	if y.Spec.EnvVar != "" {
		e := y.Spec.EnvVar
		c.EnvVar = &e
	}
	if y.Spec.Description != "" {
		d := y.Spec.Description
		c.Description = &d
	}
}

// ValidateCredentialYAML validates a CredentialYAML for correctness
func ValidateCredentialYAML(y CredentialYAML) error {
	if y.Kind != "Credential" {
		return fmt.Errorf("kind must be 'Credential', got '%s'", y.Kind)
	}

	if y.Metadata.Name == "" {
		return fmt.Errorf("name is required")
	}

	// Validate exactly one scope is specified
	scopeCount := 0
	if y.Metadata.Ecosystem != "" {
		scopeCount++
	}
	if y.Metadata.Domain != "" {
		scopeCount++
	}
	if y.Metadata.App != "" {
		scopeCount++
	}
	if y.Metadata.Workspace != "" {
		scopeCount++
	}
	if scopeCount != 1 {
		return fmt.Errorf("exactly one scope (ecosystem, domain, app, or workspace) must be specified, got %d", scopeCount)
	}

	// Validate source
	if y.Spec.Source == "" {
		return fmt.Errorf("source is required")
	}
	if y.Spec.Source != "keychain" && y.Spec.Source != "env" {
		return fmt.Errorf("source must be 'keychain' or 'env', got '%s'", y.Spec.Source)
	}

	// Validate source-specific fields
	if y.Spec.Source == "keychain" && y.Spec.Service == "" {
		return fmt.Errorf("service is required for keychain source")
	}
	if y.Spec.Source == "env" && y.Spec.EnvVar == "" {
		return fmt.Errorf("env-var is required for env source")
	}

	return nil
}
