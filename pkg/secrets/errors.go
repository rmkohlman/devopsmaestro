package secrets

import (
	"errors"
	"fmt"
)

// Sentinel errors for secret operations.
var (
	// ErrSecretNotFound indicates the requested secret does not exist.
	ErrSecretNotFound = errors.New("secret not found")

	// ErrProviderNotFound indicates the requested provider is not registered.
	ErrProviderNotFound = errors.New("secret provider not found")

	// ErrInvalidReference indicates the secret reference is malformed.
	ErrInvalidReference = errors.New("invalid secret reference")

	// ErrProviderNotAvailable indicates the provider exists but cannot be used
	// in the current environment (e.g., keychain on Linux).
	ErrProviderNotAvailable = errors.New("secret provider not available in this environment")

	// ErrNoDefaultProvider indicates no default provider has been configured.
	ErrNoDefaultProvider = errors.New("no default secret provider configured")
)

// ProviderError wraps an error with provider context.
type ProviderError struct {
	Provider string
	Op       string
	Err      error
}

func (e *ProviderError) Error() string {
	return fmt.Sprintf("secret provider %q %s: %v", e.Provider, e.Op, e.Err)
}

func (e *ProviderError) Unwrap() error {
	return e.Err
}

// SecretError wraps an error with secret context.
// Note: Never include the secret value in error messages.
type SecretError struct {
	Name string
	Op   string
	Err  error
}

func (e *SecretError) Error() string {
	return fmt.Sprintf("secret %q %s: %v", e.Name, e.Op, e.Err)
}

func (e *SecretError) Unwrap() error {
	return e.Err
}

// ResolveError indicates an error during inline secret resolution.
type ResolveError struct {
	Pattern string
	Err     error
}

func (e *ResolveError) Error() string {
	// Don't include the actual pattern in case it might leak secret info
	return fmt.Sprintf("failed to resolve secret reference: %v", e.Err)
}

func (e *ResolveError) Unwrap() error {
	return e.Err
}

// IsNotFound returns true if the error indicates a secret was not found.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrSecretNotFound)
}

// IsProviderNotFound returns true if the error indicates a provider was not found.
func IsProviderNotFound(err error) bool {
	return errors.Is(err, ErrProviderNotFound)
}

// IsProviderNotAvailable returns true if the error indicates a provider is not available.
func IsProviderNotAvailable(err error) bool {
	return errors.Is(err, ErrProviderNotAvailable)
}
