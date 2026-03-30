package models

import (
	"devopsmaestro/config"
	"devopsmaestro/pkg/envvalidation"
	"encoding/json"
	"fmt"
	"log/slog"
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
	ID                  int64               `db:"id" json:"id"`
	ScopeType           CredentialScopeType `db:"scope_type" json:"scope_type"`
	ScopeID             int64               `db:"scope_id" json:"scope_id"`
	Name                string              `db:"name" json:"name"`
	Source              string              `db:"source" json:"source"` // "vault", "env"
	VaultSecret         *string             `db:"vault_secret" json:"vault_secret,omitempty"`
	VaultEnv            *string             `db:"vault_env" json:"vault_env,omitempty"`
	VaultUsernameSecret *string             `db:"vault_username_secret" json:"vault_username_secret,omitempty"`
	VaultFields         *string             `db:"vault_fields" json:"vault_fields,omitempty"`
	EnvVar              *string             `db:"env_var" json:"env_var"` // Environment variable name
	Description         *string             `db:"description" json:"description"`
	UsernameVar         *string             `db:"username_var" json:"username_var,omitempty"`
	PasswordVar         *string             `db:"password_var" json:"password_var,omitempty"`
	CreatedAt           time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time           `db:"updated_at" json:"updated_at"`
}

// IsDualField returns true if the credential has explicit username or password var names.
func (c *CredentialDB) IsDualField() bool {
	return c.UsernameVar != nil || c.PasswordVar != nil
}

// HasVaultFields returns true if the credential has vault field mappings.
func (c *CredentialDB) HasVaultFields() bool {
	if c.VaultFields == nil || *c.VaultFields == "" || *c.VaultFields == "{}" {
		return false
	}
	return true
}

// GetVaultFieldsMap parses the VaultFields JSON string into a map.
// Returns nil, nil if no vault fields are set.
func (c *CredentialDB) GetVaultFieldsMap() (map[string]string, error) {
	if !c.HasVaultFields() {
		return nil, nil
	}
	var fields map[string]string
	if err := json.Unmarshal([]byte(*c.VaultFields), &fields); err != nil {
		return nil, fmt.Errorf("failed to parse vault fields JSON: %w", err)
	}
	return fields, nil
}

// ToFieldConfig creates a CredentialConfig for a single vault field.
func (c *CredentialDB) ToFieldConfig(field string) config.CredentialConfig {
	return config.CredentialConfig{
		Source:      config.SourceVault,
		VaultSecret: deref(c.VaultSecret),
		VaultEnv:    deref(c.VaultEnv),
		VaultField:  field,
	}
}

// ToMapEntries converts a single CredentialDB into its config.Credentials map entries.
// Handles all three fan-out cases: vault fields, dual-field, and single credential.
func (c *CredentialDB) ToMapEntries() config.Credentials {
	result := make(config.Credentials)
	switch {
	case c.HasVaultFields():
		fields, err := c.GetVaultFieldsMap()
		if err != nil {
			slog.Warn("failed to parse vault_fields JSON, skipping credential", "name", c.Name, "error", err)
			return result
		}
		for envVar, field := range fields {
			result[envVar] = c.ToFieldConfig(field)
		}
	case c.IsDualField():
		if c.UsernameVar != nil {
			result[*c.UsernameVar] = c.ToUsernameConfig()
		}
		if c.PasswordVar != nil {
			result[*c.PasswordVar] = c.ToPasswordConfig()
		}
	default:
		result[c.Name] = c.ToConfig()
	}
	return result
}

// Validate checks for mutual exclusivity constraints on the credential.
func (c *CredentialDB) Validate() error {
	if c.HasVaultFields() && c.IsDualField() {
		return fmt.Errorf("vault fields cannot be used with username/password vars")
	}
	if c.HasVaultFields() && c.VaultUsernameSecret != nil && *c.VaultUsernameSecret != "" {
		return fmt.Errorf("vault fields cannot be used with vault username secret")
	}
	return nil
}

// deref safely dereferences a string pointer, returning "" for nil.
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ToUsernameConfig creates a CredentialConfig targeting the vault username field.
func (c *CredentialDB) ToUsernameConfig() config.CredentialConfig {
	cfg := config.CredentialConfig{
		Source: config.SourceVault,
	}
	// Use VaultUsernameSecret if available, otherwise fall back to VaultSecret
	if c.VaultUsernameSecret != nil {
		cfg.VaultSecret = *c.VaultUsernameSecret
	} else if c.VaultSecret != nil {
		cfg.VaultSecret = *c.VaultSecret
	}
	if c.VaultEnv != nil {
		cfg.VaultEnv = *c.VaultEnv
	}
	// Set VaultField to the username var name for field-level vault access
	if c.UsernameVar != nil {
		cfg.VaultField = *c.UsernameVar
	}
	return cfg
}

// ToPasswordConfig creates a CredentialConfig targeting the vault password field.
func (c *CredentialDB) ToPasswordConfig() config.CredentialConfig {
	cfg := config.CredentialConfig{
		Source: config.SourceVault,
	}
	if c.VaultSecret != nil {
		cfg.VaultSecret = *c.VaultSecret
	}
	if c.VaultEnv != nil {
		cfg.VaultEnv = *c.VaultEnv
	}
	// Set VaultField to the password var name for field-level vault access
	if c.PasswordVar != nil {
		cfg.VaultField = *c.PasswordVar
	}
	return cfg
}

// ToConfig converts a CredentialDB to a config.CredentialConfig
func (c *CredentialDB) ToConfig() config.CredentialConfig {
	cfg := config.CredentialConfig{
		Source: config.CredentialSource(c.Source),
	}
	if c.VaultSecret != nil {
		cfg.VaultSecret = *c.VaultSecret
	}
	if c.VaultEnv != nil {
		cfg.VaultEnv = *c.VaultEnv
	}
	if c.EnvVar != nil {
		cfg.EnvVar = *c.EnvVar
	}
	return cfg
}

// CredentialsToMap converts a slice of CredentialDB to a config.Credentials map.
// For each credential, delegates to ToMapEntries which handles vault fields,
// dual-field, and single credential fan-out cases.
func CredentialsToMap(creds []*CredentialDB) config.Credentials {
	result := make(config.Credentials)
	for _, c := range creds {
		for k, v := range c.ToMapEntries() {
			result[k] = v
		}
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
	Source              string            `yaml:"source" json:"source"`
	VaultSecret         string            `yaml:"vaultSecret,omitempty" json:"vaultSecret,omitempty"`
	VaultEnvironment    string            `yaml:"vaultEnvironment,omitempty" json:"vaultEnvironment,omitempty"`
	VaultUsernameSecret string            `yaml:"vaultUsernameSecret,omitempty" json:"vaultUsernameSecret,omitempty"`
	VaultFields         map[string]string `yaml:"vaultFields,omitempty" json:"vaultFields,omitempty"`
	EnvVar              string            `yaml:"envVar,omitempty" json:"envVar,omitempty"`
	Description         string            `yaml:"description,omitempty" json:"description,omitempty"`
	UsernameVar         string            `yaml:"usernameVar,omitempty" json:"usernameVar,omitempty"`
	PasswordVar         string            `yaml:"passwordVar,omitempty" json:"passwordVar,omitempty"`
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
	if c.VaultSecret != nil {
		y.Spec.VaultSecret = *c.VaultSecret
	}
	if c.VaultEnv != nil {
		y.Spec.VaultEnvironment = *c.VaultEnv
	}
	if c.VaultUsernameSecret != nil {
		y.Spec.VaultUsernameSecret = *c.VaultUsernameSecret
	}
	if c.HasVaultFields() {
		fields, err := c.GetVaultFieldsMap()
		if err == nil {
			y.Spec.VaultFields = fields
		}
	}
	if c.EnvVar != nil {
		y.Spec.EnvVar = *c.EnvVar
	}
	if c.Description != nil {
		y.Spec.Description = *c.Description
	}
	if c.UsernameVar != nil {
		y.Spec.UsernameVar = *c.UsernameVar
	}
	if c.PasswordVar != nil {
		y.Spec.PasswordVar = *c.PasswordVar
	}

	return y
}

// FromYAML populates a CredentialDB from CredentialYAML.
// Note: ScopeType and ScopeID must be resolved separately by the caller.
func (c *CredentialDB) FromYAML(y CredentialYAML) {
	c.Name = y.Metadata.Name
	c.Source = y.Spec.Source

	if y.Spec.VaultSecret != "" {
		s := y.Spec.VaultSecret
		c.VaultSecret = &s
	}
	if y.Spec.VaultEnvironment != "" {
		e := y.Spec.VaultEnvironment
		c.VaultEnv = &e
	}
	if y.Spec.VaultUsernameSecret != "" {
		u := y.Spec.VaultUsernameSecret
		c.VaultUsernameSecret = &u
	}
	if y.Spec.EnvVar != "" {
		e := y.Spec.EnvVar
		c.EnvVar = &e
	}
	if y.Spec.Description != "" {
		d := y.Spec.Description
		c.Description = &d
	}
	if y.Spec.UsernameVar != "" {
		u := y.Spec.UsernameVar
		c.UsernameVar = &u
	}
	if y.Spec.PasswordVar != "" {
		p := y.Spec.PasswordVar
		c.PasswordVar = &p
	}
	if len(y.Spec.VaultFields) > 0 {
		vfJSON, err := json.Marshal(y.Spec.VaultFields)
		if err == nil {
			s := string(vfJSON)
			c.VaultFields = &s
		}
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
	if y.Spec.Source != "vault" && y.Spec.Source != "env" {
		return fmt.Errorf("source must be 'vault' or 'env', got '%s'", y.Spec.Source)
	}

	// Validate source-specific fields
	if y.Spec.Source == "vault" && y.Spec.VaultSecret == "" {
		return fmt.Errorf("vaultSecret is required for vault source")
	}
	if y.Spec.Source == "env" && y.Spec.EnvVar == "" {
		return fmt.Errorf("env-var is required for env source")
	}

	// Validate dual-field vars are only used with vault source
	if y.Spec.Source != "vault" && (y.Spec.UsernameVar != "" || y.Spec.PasswordVar != "") {
		return fmt.Errorf("usernameVar and passwordVar are only valid with vault source")
	}

	// Cross-validation: vaultUsernameSecret requires usernameVar
	if y.Spec.VaultUsernameSecret != "" && y.Spec.UsernameVar == "" {
		return fmt.Errorf("vaultUsernameSecret requires usernameVar to be set")
	}

	// Validate env key format for dual-field var names
	if y.Spec.UsernameVar != "" {
		if err := envvalidation.ValidateEnvKey(y.Spec.UsernameVar); err != nil {
			return fmt.Errorf("invalid usernameVar: %w", err)
		}
	}
	if y.Spec.PasswordVar != "" {
		if err := envvalidation.ValidateEnvKey(y.Spec.PasswordVar); err != nil {
			return fmt.Errorf("invalid passwordVar: %w", err)
		}
	}

	// vaultFields requires vault source
	if len(y.Spec.VaultFields) > 0 && y.Spec.Source != "vault" {
		return fmt.Errorf("vaultFields is only valid with vault source")
	}
	// vaultFields requires vaultSecret
	if len(y.Spec.VaultFields) > 0 && y.Spec.VaultSecret == "" {
		return fmt.Errorf("vaultFields requires vaultSecret to be set")
	}
	// vaultFields mutually exclusive with dual-field
	if len(y.Spec.VaultFields) > 0 && (y.Spec.UsernameVar != "" || y.Spec.PasswordVar != "") {
		return fmt.Errorf("vaultFields cannot be used with usernameVar/passwordVar")
	}
	// max 50 vault fields
	if len(y.Spec.VaultFields) > 50 {
		return fmt.Errorf("too many vault fields (%d): maximum is 50", len(y.Spec.VaultFields))
	}
	// validate env var keys and field names
	for envVar, fieldName := range y.Spec.VaultFields {
		if err := envvalidation.ValidateEnvKey(envVar); err != nil {
			return fmt.Errorf("invalid env var in vaultFields %q: %w", envVar, err)
		}
		if fieldName == "" {
			return fmt.Errorf("field name cannot be empty for env var %q in vaultFields", envVar)
		}
	}

	return nil
}
