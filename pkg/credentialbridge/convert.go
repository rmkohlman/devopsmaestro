// Package credentialbridge converts models.CredentialDB into config types.
// It bridges the models layer (pure data) with the config layer (runtime
// configuration), keeping models/ free of sibling-package imports.
package credentialbridge

import (
	"devopsmaestro/config"
	"devopsmaestro/models"
	"devopsmaestro/pkg/envvalidation"
	"fmt"
	"log/slog"
)

// deref safely dereferences a string pointer, returning "" for nil.
func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ToFieldConfig creates a CredentialConfig for a single vault field.
func ToFieldConfig(c *models.CredentialDB, field string) config.CredentialConfig {
	return config.CredentialConfig{
		Source:      config.SourceVault,
		VaultSecret: deref(c.VaultSecret),
		VaultEnv:    deref(c.VaultEnv),
		VaultField:  field,
	}
}

// ToMapEntries converts a single CredentialDB into its config.Credentials map entries.
// Handles all three fan-out cases: vault fields, dual-field, and single credential.
func ToMapEntries(c *models.CredentialDB) config.Credentials {
	result := make(config.Credentials)
	switch {
	case c.HasVaultFields():
		fields, err := c.GetVaultFieldsMap()
		if err != nil {
			slog.Warn("failed to parse vault_fields JSON, skipping credential", "name", c.Name, "error", err)
			return result
		}
		for envVar, field := range fields {
			result[envVar] = ToFieldConfig(c, field)
		}
	case c.IsDualField():
		if c.UsernameVar != nil {
			result[*c.UsernameVar] = ToUsernameConfig(c)
		}
		if c.PasswordVar != nil {
			result[*c.PasswordVar] = ToPasswordConfig(c)
		}
	default:
		result[c.Name] = ToConfig(c)
	}
	return result
}

// ToUsernameConfig creates a CredentialConfig targeting the vault username field.
func ToUsernameConfig(c *models.CredentialDB) config.CredentialConfig {
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
func ToPasswordConfig(c *models.CredentialDB) config.CredentialConfig {
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

// ToConfig converts a CredentialDB to a config.CredentialConfig.
func ToConfig(c *models.CredentialDB) config.CredentialConfig {
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
func CredentialsToMap(creds []*models.CredentialDB) config.Credentials {
	result := make(config.Credentials)
	for _, c := range creds {
		for k, v := range ToMapEntries(c) {
			result[k] = v
		}
	}
	return result
}

// ValidateCredentialYAML validates a CredentialYAML for correctness.
func ValidateCredentialYAML(y models.CredentialYAML) error {
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
