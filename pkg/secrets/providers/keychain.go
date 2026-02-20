package providers

import (
	"bytes"
	"context"
	"os/exec"
	"runtime"
	"strings"

	"devopsmaestro/pkg/secrets"
)

// DefaultKeychainService is the default service name for keychain secrets.
const DefaultKeychainService = "devopsmaestro"

// KeychainProvider reads secrets from macOS Keychain.
// It uses the `security` CLI tool to retrieve generic passwords.
//
// Keychain entries are expected to be stored with:
//   - Service: devopsmaestro (configurable)
//   - Account: the secret name
//
// To add a secret to keychain:
//
//	security add-generic-password -s devopsmaestro -a github-token -w "your-token"
type KeychainProvider struct {
	service string
	// execCommand is injectable for testing
	execCommand func(ctx context.Context, name string, args ...string) *exec.Cmd
}

// KeychainProviderOption is a functional option for configuring KeychainProvider.
type KeychainProviderOption func(*KeychainProvider)

// WithKeychainService sets a custom service name.
func WithKeychainService(service string) KeychainProviderOption {
	return func(p *KeychainProvider) {
		p.service = service
	}
}

// withExecCommand is used for testing to mock exec.Command.
func withExecCommand(fn func(ctx context.Context, name string, args ...string) *exec.Cmd) KeychainProviderOption {
	return func(p *KeychainProvider) {
		p.execCommand = fn
	}
}

// NewKeychainProvider creates a macOS Keychain secret provider.
// By default, uses service name "devopsmaestro".
func NewKeychainProvider(opts ...KeychainProviderOption) *KeychainProvider {
	p := &KeychainProvider{
		service:     DefaultKeychainService,
		execCommand: exec.CommandContext,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Name returns the provider identifier.
func (p *KeychainProvider) Name() string {
	return "keychain"
}

// IsAvailable returns true only on macOS.
func (p *KeychainProvider) IsAvailable() bool {
	return runtime.GOOS == "darwin"
}

// GetSecret retrieves a secret from macOS Keychain.
// Uses: security find-generic-password -s <service> -a <name> -w
//
// The -w flag returns only the password value without any metadata.
func (p *KeychainProvider) GetSecret(ctx context.Context, req secrets.SecretRequest) (string, error) {
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

	// Determine service name (allow override via options)
	service := p.service
	if override, ok := req.Options["service"]; ok && override != "" {
		service = override
	}

	// Build the security command
	// security find-generic-password -s <service> -a <account> -w
	cmd := p.execCommand(ctx,
		"security",
		"find-generic-password",
		"-s", service,
		"-a", req.Name,
		"-w",
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		// Check for "item not found" error
		stderrStr := stderr.String()
		if strings.Contains(stderrStr, "could not be found") ||
			strings.Contains(stderrStr, "SecKeychainSearchCopyNext") ||
			strings.Contains(stderrStr, "The specified item could not be found") {
			return "", secrets.ErrSecretNotFound
		}

		// Other error (permission denied, etc.)
		return "", &secrets.ProviderError{
			Provider: "keychain",
			Op:       "find-generic-password",
			Err:      err,
		}
	}

	// Trim trailing newline from output
	value := strings.TrimSuffix(stdout.String(), "\n")

	return value, nil
}

// Service returns the current service name.
func (p *KeychainProvider) Service() string {
	return p.service
}

// Ensure KeychainProvider implements SecretProvider
var _ secrets.SecretProvider = (*KeychainProvider)(nil)
