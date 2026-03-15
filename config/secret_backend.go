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

// FieldCapableBackend extends SecretBackend with field-level secret access.
// Only backends that support N fields per secret implement this.
type FieldCapableBackend interface {
	SecretBackend
	GetField(name, env, field string) (string, error)
}
