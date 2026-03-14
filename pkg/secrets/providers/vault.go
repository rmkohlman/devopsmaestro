// Package providers contains concrete implementations of SecretProvider.
package providers

import (
	"context"
	"fmt"
	"os"

	"devopsmaestro/pkg/secrets"

	"github.com/rmkohlman/MaestroVault/pkg/client"
)

// VaultProvider reads secrets from MaestroVault.
// It uses the MaestroVault Go client library to retrieve secrets via Unix socket.
//
// Authentication requires MAV_TOKEN environment variable to be set.
// The vault daemon must be running (see: mav serve --no-touchid).
//
// Secret names are used as-is (no transformation).
// The "environment" option can be passed via SecretRequest.Options["environment"].
type VaultProvider struct {
	token  string
	client *client.Client
}

// VaultProviderOption is a functional option for configuring VaultProvider.
type VaultProviderOption func(*VaultProvider)

// WithVaultToken sets an explicit token instead of reading from MAV_TOKEN.
func WithVaultToken(token string) VaultProviderOption {
	return func(p *VaultProvider) {
		p.token = token
	}
}

// NewVaultProvider creates a MaestroVault secret provider.
// By default, reads the token from MAV_TOKEN environment variable.
func NewVaultProvider(opts ...VaultProviderOption) *VaultProvider {
	p := &VaultProvider{
		token: os.Getenv("MAV_TOKEN"),
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Name returns the provider identifier.
func (p *VaultProvider) Name() string {
	return "vault"
}

// IsAvailable returns true if the MAV_TOKEN is set (provider can attempt connections).
func (p *VaultProvider) IsAvailable() bool {
	return p.token != ""
}

// GetSecret retrieves a secret from MaestroVault.
// The secret name is used directly (no transformation).
// An optional "environment" key in req.Options specifies the vault environment.
func (p *VaultProvider) GetSecret(ctx context.Context, req secrets.SecretRequest) (string, error) {
	// Check availability first
	if !p.IsAvailable() {
		return "", secrets.ErrProviderNotAvailable
	}

	// Check context cancellation
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	// Lazy-initialize client
	if p.client == nil {
		c, err := client.New(p.token)
		if err != nil {
			return "", &secrets.ProviderError{
				Provider: "vault",
				Op:       "connect",
				Err:      fmt.Errorf("failed to create vault client: %w", err),
			}
		}
		p.client = c
	}

	// Determine environment from options
	env := ""
	if override, ok := req.Options["environment"]; ok && override != "" {
		env = override
	}

	// Get the secret
	entry, err := p.client.Get(req.Name, env)
	if err != nil {
		// Check if it's a "not found" type error
		return "", &secrets.ProviderError{
			Provider: "vault",
			Op:       "get",
			Err:      fmt.Errorf("secret %q not found (env=%q): %w", req.Name, env, err),
		}
	}

	return entry.Value, nil
}

// Ensure VaultProvider implements SecretProvider
var _ secrets.SecretProvider = (*VaultProvider)(nil)
