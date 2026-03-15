package config

import (
	"fmt"

	"github.com/rmkohlman/MaestroVault/pkg/client"
)

// VaultBackend implements SecretBackend using MaestroVault.
type VaultBackend struct {
	client *client.Client
}

// NewVaultBackend creates a new VaultBackend with the given token.
// The token is used to authenticate with the MaestroVault daemon.
// Returns an error if the token is empty.
// Does NOT validate connectivity at construction time — that happens on first Get/Health call.
func NewVaultBackend(token string) (*VaultBackend, error) {
	if token == "" {
		return nil, fmt.Errorf("vault token is required: set MAV_TOKEN environment variable")
	}

	c, err := client.New(token)
	if err != nil {
		return nil, fmt.Errorf("failed to create vault client: %w", err)
	}

	return &VaultBackend{client: c}, nil
}

// Get retrieves a secret from MaestroVault by name and environment.
func (v *VaultBackend) Get(name, env string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("vault secret name is required")
	}

	entry, err := v.client.Get(name, env)
	if err != nil {
		return "", fmt.Errorf("vault lookup failed for %q (env=%q): %w", name, env, err)
	}

	return entry.Value, nil
}

// Health checks if the MaestroVault daemon is reachable.
func (v *VaultBackend) Health() error {
	return v.client.Health()
}

// GetField retrieves a single field from a MaestroVault secret by name, environment, and field key.
func (v *VaultBackend) GetField(name, env, field string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("vault secret name is required")
	}
	if field == "" {
		return "", fmt.Errorf("vault field name is required")
	}
	value, err := v.client.GetField(name, env, field)
	if err != nil {
		return "", fmt.Errorf("vault field lookup failed for %q field %q (env=%q): %w", name, field, env, err)
	}
	return value, nil
}

// Compile-time interface check
var _ FieldCapableBackend = (*VaultBackend)(nil)
