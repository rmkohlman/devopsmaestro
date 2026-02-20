// Package secrets provides a pluggable secret provider system for retrieving secrets
// from various backends (Keychain, environment variables, 1Password, etc.) and
// injecting them into YAML during `dvm apply`.
//
// Architecture follows the Interface → Implementation → Factory pattern from STANDARDS.md.
//
// Dependency Flow:
//
//	cmd/ → uses Resolver to process YAML with secrets
//	pkg/secrets/ → defines interfaces, implements providers
//	pkg/secrets/providers/ → concrete provider implementations
package secrets

import (
	"context"
)

// SecretProvider defines the interface for retrieving secrets from a backend.
// Implementations include keychain, environment variables, 1Password, etc.
//
// All implementations must be safe for concurrent use.
type SecretProvider interface {
	// Name returns the provider identifier (e.g., "keychain", "env")
	Name() string

	// IsAvailable checks if this provider can be used in the current environment.
	// For example, keychain is only available on macOS.
	IsAvailable() bool

	// GetSecret retrieves a secret by name.
	// Returns the secret value and nil error if found.
	// Returns "", ErrSecretNotFound if not found.
	// Returns "", error for other failures (permission denied, etc.)
	GetSecret(ctx context.Context, req SecretRequest) (string, error)
}

// SecretRequest contains information needed to retrieve a secret.
type SecretRequest struct {
	// Name is the secret identifier (required).
	// For env provider: "github-token" maps to "DVM_SECRET_GITHUB_TOKEN"
	// For keychain: the account name in the keychain
	Name string

	// Key is a specific field within a structured secret (optional).
	// Used for JSON secrets where you want a specific field.
	Key string

	// Options are provider-specific options.
	// For example, keychain might accept "service" to override the default.
	Options map[string]string
}

// SecretReference represents a reference to a secret in YAML.
// This is used for the valueFrom pattern.
type SecretReference struct {
	// Name is the secret identifier (required)
	Name string `yaml:"name"`

	// Provider overrides the default provider (optional)
	Provider string `yaml:"provider,omitempty"`

	// Key is for structured secrets - retrieves a specific field (optional)
	Key string `yaml:"key,omitempty"`

	// Options are provider-specific options (optional)
	Options map[string]string `yaml:"options,omitempty"`
}

// ValueSource represents the Kubernetes-style valueFrom pattern.
// Used in YAML to reference secrets instead of inline values.
//
// Example YAML:
//
//	env:
//	  - name: GITHUB_TOKEN
//	    valueFrom:
//	      secretRef:
//	        name: github-token
//	        provider: keychain
type ValueSource struct {
	SecretRef *SecretReference `yaml:"secretRef,omitempty"`
}

// Ensure interface compliance is checked at compile time
var _ SecretProvider = (SecretProvider)(nil)
