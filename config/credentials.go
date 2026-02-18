package config

import (
	"fmt"
	"os"
)

// CredentialSource defines where a credential value comes from
type CredentialSource string

const (
	// SourceKeychain retrieves the credential from macOS Keychain
	SourceKeychain CredentialSource = "keychain"
	// SourceEnv retrieves the credential from an environment variable
	SourceEnv CredentialSource = "env"
	// SourceValue uses a plaintext value (not recommended for secrets)
	SourceValue CredentialSource = "value"
)

// CredentialConfig defines how to retrieve a single credential
type CredentialConfig struct {
	// Source specifies where to retrieve the credential from
	Source CredentialSource `yaml:"source" json:"source" mapstructure:"source"`
	// Service is the Keychain service name (when Source is "keychain")
	Service string `yaml:"service,omitempty" json:"service,omitempty" mapstructure:"service"`
	// EnvVar is the environment variable name (when Source is "env")
	EnvVar string `yaml:"env,omitempty" json:"env,omitempty" mapstructure:"env"`
	// Value is the plaintext value (when Source is "value", not recommended)
	Value string `yaml:"value,omitempty" json:"value,omitempty" mapstructure:"value"`
}

// Credentials is a map of credential name to its configuration
type Credentials map[string]CredentialConfig

// CredentialScope represents a level in the hierarchy that can have credentials
type CredentialScope struct {
	Type        string      // "global", "ecosystem", "domain", "app", "workspace"
	ID          int64       // Database ID (0 for global)
	Name        string      // Name of the scope
	Credentials Credentials // Credentials at this scope
}

// ResolveCredential resolves a single credential config to its actual value
func ResolveCredential(cfg CredentialConfig) (string, error) {
	switch cfg.Source {
	case SourceKeychain:
		if cfg.Service == "" {
			return "", fmt.Errorf("keychain source requires service name")
		}
		return GetFromKeychain(cfg.Service)
	case SourceEnv:
		if cfg.EnvVar == "" {
			return "", fmt.Errorf("env source requires env var name")
		}
		return os.Getenv(cfg.EnvVar), nil
	case SourceValue:
		return cfg.Value, nil
	default:
		return "", fmt.Errorf("unknown credential source: %s", cfg.Source)
	}
}

// ResolveCredentials resolves all credentials from multiple scopes,
// applying inheritance (later scopes override earlier ones).
// Environment variables always take highest priority.
func ResolveCredentials(scopes ...CredentialScope) map[string]string {
	result := make(map[string]string)
	errors := make(map[string]error)

	// Process scopes in order (global -> ecosystem -> domain -> app -> workspace)
	for _, scope := range scopes {
		for name, cfg := range scope.Credentials {
			val, err := ResolveCredential(cfg)
			if err != nil {
				errors[name] = err
				continue
			}
			if val != "" {
				result[name] = val
			}
		}
	}

	// Environment variables always win (check all resolved credential names)
	for name := range result {
		if envVal := os.Getenv(name); envVal != "" {
			result[name] = envVal
		}
	}

	return result
}

// ResolveCredentialsWithErrors is like ResolveCredentials but also returns errors
func ResolveCredentialsWithErrors(scopes ...CredentialScope) (map[string]string, map[string]error) {
	result := make(map[string]string)
	errors := make(map[string]error)

	// Process scopes in order (global -> ecosystem -> domain -> app -> workspace)
	for _, scope := range scopes {
		for name, cfg := range scope.Credentials {
			val, err := ResolveCredential(cfg)
			if err != nil {
				errors[name] = fmt.Errorf("[%s] %w", scope.Type, err)
				continue
			}
			if val != "" {
				result[name] = val
				// Clear any previous error since we got a value
				delete(errors, name)
			}
		}
	}

	// Environment variables always win
	for name := range result {
		if envVal := os.Getenv(name); envVal != "" {
			result[name] = envVal
			delete(errors, name)
		}
	}

	return result, errors
}

// GetGlobalCredentials returns credentials from the global config
func GetGlobalCredentials() Credentials {
	cfg := GetConfig()
	return cfg.Credentials
}

// MergeCredentials merges multiple credential maps, later ones override earlier
func MergeCredentials(credMaps ...Credentials) Credentials {
	result := make(Credentials)
	for _, creds := range credMaps {
		for name, cfg := range creds {
			result[name] = cfg
		}
	}
	return result
}
