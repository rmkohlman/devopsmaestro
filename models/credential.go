package models

import (
	"encoding/json"
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
	ExpiresAt           *time.Time          `db:"expires_at" json:"expires_at,omitempty" yaml:"expiresAt,omitempty"`
	CreatedAt           time.Time           `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time           `db:"updated_at" json:"updated_at"`
}

// ExpirationStatus returns the expiration status of a credential.
// Returns "expired", "expiring soon" (≤7 days), "valid", or "" (no expiration set).
func (c *CredentialDB) ExpirationStatus() string {
	return CredentialExpirationStatusAt(c.ExpiresAt, time.Now())
}

// CredentialExpirationStatusAt computes expiration status relative to a given time.
// This enables deterministic testing.
func CredentialExpirationStatusAt(expiresAt *time.Time, now time.Time) string {
	if expiresAt == nil {
		return ""
	}
	if now.After(*expiresAt) {
		return "expired"
	}
	if expiresAt.Sub(now) <= 7*24*time.Hour {
		return "expiring soon"
	}
	return "valid"
}

// IsExpired returns true if the credential has expired.
func (c *CredentialDB) IsExpired() bool {
	return c.ExpirationStatus() == "expired"
}

// IsExpiringSoon returns true if the credential will expire within 7 days.
func (c *CredentialDB) IsExpiringSoon() bool {
	return c.ExpirationStatus() == "expiring soon"
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
