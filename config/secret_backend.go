package config

// SecretBackend abstracts secret retrieval from a vault backend.
// This enables mocking in tests and future provider swaps.
type SecretBackend interface {
	// Get retrieves a secret by name and environment.
	// Returns the secret value or an error if retrieval fails.
	Get(name, env string) (string, error)

	// Health checks if the backend is reachable.
	Health() error
}
