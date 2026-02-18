package models

import (
	"devopsmaestro/config"
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
	Source      string              `db:"source" json:"source"`   // "keychain", "env", "value"
	Service     *string             `db:"service" json:"service"` // Keychain service name
	EnvVar      *string             `db:"env_var" json:"env_var"` // Environment variable name
	Value       *string             `db:"value" json:"value"`     // Plaintext value (not recommended)
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
	if c.Value != nil {
		cfg.Value = *c.Value
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
