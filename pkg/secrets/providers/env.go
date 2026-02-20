// Package providers contains concrete implementations of SecretProvider.
package providers

import (
	"context"
	"os"
	"strings"

	"devopsmaestro/pkg/secrets"
)

// DefaultEnvPrefix is the default prefix for environment variable secrets.
const DefaultEnvPrefix = "DVM_SECRET_"

// EnvProvider reads secrets from environment variables.
// Secret names are converted to environment variable format with a configurable prefix.
//
// Conversion rules:
//   - github-token -> DVM_SECRET_GITHUB_TOKEN
//   - api.key -> DVM_SECRET_API_KEY
//
// The prefix can be customized via options.
type EnvProvider struct {
	prefix string
}

// EnvProviderOption is a functional option for configuring EnvProvider.
type EnvProviderOption func(*EnvProvider)

// WithEnvPrefix sets a custom prefix for environment variables.
func WithEnvPrefix(prefix string) EnvProviderOption {
	return func(p *EnvProvider) {
		p.prefix = prefix
	}
}

// NewEnvProvider creates an environment variable secret provider.
// By default, uses prefix "DVM_SECRET_".
func NewEnvProvider(opts ...EnvProviderOption) *EnvProvider {
	p := &EnvProvider{
		prefix: DefaultEnvPrefix,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Name returns the provider identifier.
func (p *EnvProvider) Name() string {
	return "env"
}

// IsAvailable returns true - environment variables are always available.
func (p *EnvProvider) IsAvailable() bool {
	return true
}

// GetSecret retrieves a secret from environment variables.
// The secret name is converted to an environment variable name using convertToEnvVar.
//
// Example: GetSecret("github-token") looks up DVM_SECRET_GITHUB_TOKEN
//
// For backward compatibility, if the prefixed variable is not found,
// it also checks for common non-prefixed variants (e.g., GITHUB_TOKEN).
func (p *EnvProvider) GetSecret(ctx context.Context, req secrets.SecretRequest) (string, error) {
	// Check context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Convert name to environment variable format
	envVar := p.convertToEnvVar(req.Name)

	// Look up the environment variable
	value, ok := os.LookupEnv(envVar)
	if !ok {
		// Fallback: check for non-prefixed env var for backward compatibility
		// This allows GITHUB_TOKEN to work in addition to DVM_SECRET_GITHUB_TOKEN
		fallbackVar := p.convertToFallbackEnvVar(req.Name)
		if fallbackVar != "" {
			value, ok = os.LookupEnv(fallbackVar)
			if !ok {
				return "", secrets.ErrSecretNotFound
			}
		} else {
			return "", secrets.ErrSecretNotFound
		}
	}

	// Handle key extraction for JSON-like values (simple case)
	if req.Key != "" {
		// For now, we don't support structured extraction from env vars
		// The full value is returned and caller can parse if needed
		return value, nil
	}

	return value, nil
}

// convertToEnvVar converts a secret name to environment variable format.
// Example: "github-token" -> "DVM_SECRET_GITHUB_TOKEN"
func (p *EnvProvider) convertToEnvVar(name string) string {
	// Replace hyphens and dots with underscores
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	// Convert to uppercase
	name = strings.ToUpper(name)
	// Add prefix
	return p.prefix + name
}

// convertToFallbackEnvVar converts a secret name to a non-prefixed env var for backward compatibility.
// This allows common environment variables like GITHUB_TOKEN to work without the DVM_SECRET_ prefix.
// Example: "github-token" -> "GITHUB_TOKEN"
func (p *EnvProvider) convertToFallbackEnvVar(name string) string {
	// Replace hyphens and dots with underscores
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	// Convert to uppercase (no prefix)
	return strings.ToUpper(name)
}

// Prefix returns the current prefix for environment variables.
func (p *EnvProvider) Prefix() string {
	return p.prefix
}

// Ensure EnvProvider implements SecretProvider
var _ secrets.SecretProvider = (*EnvProvider)(nil)
