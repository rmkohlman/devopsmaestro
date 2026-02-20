package secrets

import (
	"context"
	"regexp"
	"strings"
)

// inlinePattern matches ${secret:name} or ${secret:name:provider}
// Group 1: secret name
// Group 2: optional provider (with leading colon)
var inlinePattern = regexp.MustCompile(`\$\{secret:([^:}]+)(?::([^}]+))?\}`)

// Resolver processes YAML content and replaces secret references.
// It supports both inline patterns (${secret:name}) and the valueFrom pattern.
//
// Thread-safe: the resolver can be used concurrently.
type Resolver struct {
	factory *ProviderFactory
	cache   *Cache
}

// ResolverOption is a functional option for configuring Resolver.
type ResolverOption func(*Resolver)

// WithCache configures a custom cache for the resolver.
func WithCache(cache *Cache) ResolverOption {
	return func(r *Resolver) {
		r.cache = cache
	}
}

// NewResolver creates a new secret resolver.
// If no cache option is provided, a default cache is created.
func NewResolver(factory *ProviderFactory, opts ...ResolverOption) *Resolver {
	r := &Resolver{
		factory: factory,
		cache:   NewCache(),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// ResolveInline handles ${secret:name} and ${secret:name:provider} patterns in strings.
// All matching patterns in the content are replaced with their resolved values.
//
// Examples:
//   - ${secret:github-token} -> uses default provider
//   - ${secret:github-token:keychain} -> uses keychain provider
//   - ${secret:api-key:env} -> uses env provider
//
// Returns the content with all secrets resolved, or an error if any secret fails.
// Security: Never logs or includes secret values in error messages.
func (r *Resolver) ResolveInline(ctx context.Context, content string) (string, error) {
	var resolveErr error

	result := inlinePattern.ReplaceAllStringFunc(content, func(match string) string {
		// If we already have an error, skip further resolution
		if resolveErr != nil {
			return match
		}

		// Parse the match
		parts := inlinePattern.FindStringSubmatch(match)
		if len(parts) < 2 {
			resolveErr = &ResolveError{Pattern: match, Err: ErrInvalidReference}
			return match
		}

		name := parts[1]
		provider := ""
		if len(parts) > 2 && parts[2] != "" {
			provider = parts[2]
		}

		// Resolve the secret
		value, err := r.resolveSecret(ctx, name, provider, "")
		if err != nil {
			resolveErr = &ResolveError{Pattern: match, Err: err}
			return match
		}

		return value
	})

	if resolveErr != nil {
		return "", resolveErr
	}

	return result, nil
}

// ResolveReference resolves a single SecretReference to its value.
// This is used for the valueFrom.secretRef pattern in YAML.
func (r *Resolver) ResolveReference(ctx context.Context, ref SecretReference) (string, error) {
	if ref.Name == "" {
		return "", &SecretError{
			Name: "(empty)",
			Op:   "resolve reference",
			Err:  ErrInvalidReference,
		}
	}

	return r.resolveSecret(ctx, ref.Name, ref.Provider, ref.Key)
}

// resolveSecret handles the actual secret resolution with caching.
func (r *Resolver) resolveSecret(ctx context.Context, name, providerName, key string) (string, error) {
	// Determine which provider to use
	var provider SecretProvider
	var err error

	if providerName == "" {
		provider, err = r.factory.GetDefault()
		if err != nil {
			return "", err
		}
		providerName = provider.Name()
	} else {
		provider, err = r.factory.Get(providerName)
		if err != nil {
			return "", err
		}
	}

	// Check cache first
	if cached, ok := r.cache.Get(providerName, name, key); ok {
		return cached, nil
	}

	// Fetch from provider
	req := SecretRequest{
		Name: name,
		Key:  key,
	}

	value, err := provider.GetSecret(ctx, req)
	if err != nil {
		return "", &SecretError{
			Name: name,
			Op:   "get",
			Err:  err,
		}
	}

	// Cache the result
	r.cache.Set(providerName, name, key, value)

	return value, nil
}

// ClearCache clears the resolver's secret cache.
// This should be called after a command completes.
func (r *Resolver) ClearCache() {
	r.cache.Clear()
}

// HasSecretReferences checks if the content contains any secret references.
// This is useful for determining if resolution is needed.
func HasSecretReferences(content string) bool {
	return inlinePattern.MatchString(content)
}

// ExtractSecretReferences extracts all secret references from content.
// Returns a slice of SecretReference for each found pattern.
func ExtractSecretReferences(content string) []SecretReference {
	matches := inlinePattern.FindAllStringSubmatch(content, -1)
	refs := make([]SecretReference, 0, len(matches))

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		ref := SecretReference{
			Name: match[1],
		}
		if len(match) > 2 && match[2] != "" {
			ref.Provider = match[2]
		}

		refs = append(refs, ref)
	}

	return refs
}

// ValidateSecretReferences validates that all secret references in content can be resolved.
// It attempts to resolve each secret but does not return the values.
// This is useful for dry-run validation.
func (r *Resolver) ValidateSecretReferences(ctx context.Context, content string) error {
	refs := ExtractSecretReferences(content)

	for _, ref := range refs {
		_, err := r.ResolveReference(ctx, ref)
		if err != nil {
			return err
		}
	}

	return nil
}

// ConvertNameToEnvVar converts a secret name to environment variable format.
// Example: "github-token" -> "GITHUB_TOKEN"
func ConvertNameToEnvVar(name string) string {
	// Replace hyphens and dots with underscores
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	// Convert to uppercase
	return strings.ToUpper(name)
}
