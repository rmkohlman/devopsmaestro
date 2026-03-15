package config

import (
	"fmt"
	"os"
)

// CredentialSource defines where a credential value comes from
type CredentialSource string

const (
	// SourceVault retrieves the credential from MaestroVault
	SourceVault CredentialSource = "vault"
	// SourceEnv retrieves the credential from an environment variable
	SourceEnv CredentialSource = "env"
)

// CredentialConfig defines how to retrieve a single credential
type CredentialConfig struct {
	// Source specifies where to retrieve the credential from
	Source CredentialSource `yaml:"source" json:"source" mapstructure:"source"`
	// VaultSecret is the secret name in MaestroVault (when Source is "vault")
	VaultSecret string `yaml:"vaultSecret,omitempty" json:"vaultSecret,omitempty" mapstructure:"vaultSecret"`
	// VaultEnv is the vault environment (when Source is "vault", optional)
	VaultEnv string `yaml:"vaultEnvironment,omitempty" json:"vaultEnvironment,omitempty" mapstructure:"vaultEnvironment"`
	// VaultField is a specific field within a vault secret (when Source is "vault", optional)
	VaultField string `yaml:"vaultField,omitempty" json:"vaultField,omitempty" mapstructure:"vaultField"`
	// EnvVar is the environment variable name (when Source is "env")
	EnvVar string `yaml:"env,omitempty" json:"env,omitempty" mapstructure:"env"`
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

// ResolveCredential resolves a single credential config to its actual value.
// For vault-sourced credentials, use ResolveCredentialWithBackend instead.
func ResolveCredential(cfg CredentialConfig) (string, error) {
	switch cfg.Source {
	case SourceVault:
		return "", fmt.Errorf("vault credentials require a backend; use ResolveCredentialWithBackend")
	case SourceEnv:
		if cfg.EnvVar == "" {
			return "", fmt.Errorf("env source requires env var name")
		}
		return os.Getenv(cfg.EnvVar), nil
	default:
		return "", fmt.Errorf("unknown credential source: %s", cfg.Source)
	}
}

// ResolveCredentialWithBackend resolves a credential using the provided SecretBackend
// for vault-sourced credentials. For env-sourced credentials, the backend is ignored.
func ResolveCredentialWithBackend(cfg CredentialConfig, backend SecretBackend) (string, error) {
	switch cfg.Source {
	case SourceVault:
		if backend == nil {
			return "", fmt.Errorf("vault backend is required for vault-sourced credentials")
		}
		if cfg.VaultSecret == "" {
			return "", fmt.Errorf("vault source requires vault secret name")
		}
		if cfg.VaultField != "" {
			fb, ok := backend.(FieldCapableBackend)
			if !ok {
				return "", fmt.Errorf("backend does not support field-level access")
			}
			return fb.GetField(cfg.VaultSecret, cfg.VaultEnv, cfg.VaultField)
		}
		return backend.Get(cfg.VaultSecret, cfg.VaultEnv)
	case SourceEnv:
		if cfg.EnvVar == "" {
			return "", fmt.Errorf("env source requires env var name")
		}
		return os.Getenv(cfg.EnvVar), nil
	default:
		return "", fmt.Errorf("unknown credential source: %s", cfg.Source)
	}
}

// ResolveCredentials resolves all credentials from multiple scopes,
// applying inheritance (later scopes override earlier ones).
// Environment variables always take highest priority.
func ResolveCredentials(scopes ...CredentialScope) map[string]string {
	result, _ := ResolveCredentialsWithErrors(scopes...)
	return result
}

// ResolveCredentialsWithErrors is like ResolveCredentials but also returns errors
func ResolveCredentialsWithErrors(scopes ...CredentialScope) (map[string]string, map[string]error) {
	result := make(map[string]string)
	errors := make(map[string]error)
	allNames := make(map[string]struct{}) // Track every credential name

	// Process scopes in order (global -> ecosystem -> domain -> app -> workspace)
	for _, scope := range scopes {
		for name, cfg := range scope.Credentials {
			allNames[name] = struct{}{} // Register name regardless of resolution outcome
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

	// Environment variables always win — check ALL names, not just resolved ones
	for name := range allNames {
		if envVal := os.Getenv(name); envVal != "" {
			result[name] = envVal
			delete(errors, name) // Env var rescued this credential
		}
	}

	return result, errors
}

// ResolveCredentialsWithBackend is like ResolveCredentialsWithErrors but accepts a SecretBackend
// for resolving vault-sourced credentials. If backend is nil, vault credentials will fail
// (same behavior as ResolveCredentialsWithErrors).
func ResolveCredentialsWithBackend(backend SecretBackend, scopes ...CredentialScope) (map[string]string, map[string]error) {
	result := make(map[string]string)
	errors := make(map[string]error)
	allNames := make(map[string]struct{})

	for _, scope := range scopes {
		for name, cfg := range scope.Credentials {
			allNames[name] = struct{}{}
			var val string
			var err error
			if cfg.Source == SourceVault && backend != nil {
				val, err = ResolveCredentialWithBackend(cfg, backend)
			} else {
				val, err = ResolveCredential(cfg)
			}
			if err != nil {
				errors[name] = fmt.Errorf("[%s] %w", scope.Type, err)
				continue
			}
			if val != "" {
				result[name] = val
				delete(errors, name)
			}
		}
	}

	// Environment variables always win
	for name := range allNames {
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
