package providers

import (
	"context"
	"os"
	"testing"

	"devopsmaestro/pkg/secrets"
)

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestEnvProvider_ImplementsSecretProvider(t *testing.T) {
	var _ secrets.SecretProvider = (*EnvProvider)(nil)
}

// =============================================================================
// EnvProvider Tests
// =============================================================================

func TestEnvProvider_Name(t *testing.T) {
	provider := NewEnvProvider()
	if provider.Name() != "env" {
		t.Errorf("got %q, want %q", provider.Name(), "env")
	}
}

func TestEnvProvider_IsAvailable(t *testing.T) {
	provider := NewEnvProvider()
	if !provider.IsAvailable() {
		t.Error("EnvProvider should always be available")
	}
}

func TestEnvProvider_GetSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("retrieves secret from environment", func(t *testing.T) {
		// Set up test environment variable
		envKey := "DVM_SECRET_TEST_TOKEN"
		envValue := "test-secret-value"
		os.Setenv(envKey, envValue)
		defer os.Unsetenv(envKey)

		provider := NewEnvProvider()
		value, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "test-token"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != envValue {
			t.Errorf("got %q, want %q", value, envValue)
		}
	})

	t.Run("returns ErrSecretNotFound for missing env var", func(t *testing.T) {
		provider := NewEnvProvider()
		_, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "nonexistent-secret"})

		if !secrets.IsNotFound(err) {
			t.Errorf("got %v, want ErrSecretNotFound", err)
		}
	})

	t.Run("converts name with hyphens", func(t *testing.T) {
		envKey := "DVM_SECRET_GITHUB_API_TOKEN"
		os.Setenv(envKey, "token123")
		defer os.Unsetenv(envKey)

		provider := NewEnvProvider()
		value, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "github-api-token"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != "token123" {
			t.Errorf("got %q, want %q", value, "token123")
		}
	})

	t.Run("converts name with dots", func(t *testing.T) {
		envKey := "DVM_SECRET_API_KEY"
		os.Setenv(envKey, "key456")
		defer os.Unsetenv(envKey)

		provider := NewEnvProvider()
		value, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "api.key"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != "key456" {
			t.Errorf("got %q, want %q", value, "key456")
		}
	})

	t.Run("respects custom prefix", func(t *testing.T) {
		envKey := "CUSTOM_PREFIX_MY_SECRET"
		os.Setenv(envKey, "custom-value")
		defer os.Unsetenv(envKey)

		provider := NewEnvProvider(WithEnvPrefix("CUSTOM_PREFIX_"))
		value, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "my-secret"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != "custom-value" {
			t.Errorf("got %q, want %q", value, "custom-value")
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		provider := NewEnvProvider()
		_, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "test"})

		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})
}

func TestEnvProvider_Prefix(t *testing.T) {
	t.Run("returns default prefix", func(t *testing.T) {
		provider := NewEnvProvider()
		if provider.Prefix() != DefaultEnvPrefix {
			t.Errorf("got %q, want %q", provider.Prefix(), DefaultEnvPrefix)
		}
	})

	t.Run("returns custom prefix", func(t *testing.T) {
		provider := NewEnvProvider(WithEnvPrefix("CUSTOM_"))
		if provider.Prefix() != "CUSTOM_" {
			t.Errorf("got %q, want %q", provider.Prefix(), "CUSTOM_")
		}
	})
}

// =============================================================================
// TDD Phase 2 (GREEN): VaultProvider Tests (v0.40.0)
// =============================================================================
// These tests verify the VaultProvider implementation in vault.go.
// VaultProvider reads secrets from MaestroVault via the MAV client.
//
// Key design:
//   - NewVaultProvider() reads MAV_TOKEN env var by default
//   - NewVaultProvider(WithVaultToken("tok")) uses an explicit token
//   - IsAvailable() returns true iff the token is non-empty
//   - GetSecret respects context cancellation
// =============================================================================

// TestVaultProvider_ImplementsSecretProvider verifies that VaultProvider
// satisfies the secrets.SecretProvider interface at compile time.
func TestVaultProvider_ImplementsSecretProvider(t *testing.T) {
	var _ secrets.SecretProvider = (*VaultProvider)(nil)
}

// TestVaultProvider_Name verifies that VaultProvider.Name() returns "vault".
func TestVaultProvider_Name(t *testing.T) {
	provider := NewVaultProvider(WithVaultToken("test-token"))

	if provider.Name() != "vault" {
		t.Errorf("got %q, want %q", provider.Name(), "vault")
	}
}

// TestVaultProvider_IsAvailable verifies that VaultProvider.IsAvailable()
// returns true when a token is configured.
func TestVaultProvider_IsAvailable(t *testing.T) {
	provider := NewVaultProvider(WithVaultToken("test-token"))

	if !provider.IsAvailable() {
		t.Error("VaultProvider should always be available when a token is configured")
	}
}

// TestVaultProvider_IsAvailable_NoToken verifies that VaultProvider.IsAvailable()
// returns false when no token is provided.
func TestVaultProvider_IsAvailable_NoToken(t *testing.T) {
	provider := NewVaultProvider() // no token — not available

	if provider.IsAvailable() {
		t.Error("VaultProvider with empty token should NOT be available")
	}
}

// TestVaultProvider_GetSecret_ContextCancellation verifies that VaultProvider
// respects context cancellation and returns context.Canceled.
func TestVaultProvider_GetSecret_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	provider := NewVaultProvider(WithVaultToken("test-token"))

	_, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "test"})
	if err != context.Canceled {
		t.Errorf("got %v, want context.Canceled", err)
	}
}

// TestVaultProvider_GetSecret_Unavailable verifies that calling GetSecret
// on a VaultProvider with no token returns ErrProviderNotAvailable.
func TestVaultProvider_GetSecret_Unavailable(t *testing.T) {
	ctx := context.Background()

	provider := NewVaultProvider() // no token → not available

	_, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "test"})
	if !secrets.IsProviderNotAvailable(err) {
		t.Errorf("got %v, want ErrProviderNotAvailable", err)
	}
}

// TestVaultProvider_Name_TableDriven is a table-driven test that verifies
// VaultProvider properties across different configurations.
func TestVaultProvider_Name_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		opts          []VaultProviderOption
		wantName      string
		wantAvailable bool
	}{
		{
			name:          "with valid token",
			opts:          []VaultProviderOption{WithVaultToken("my-vault-token")},
			wantName:      "vault",
			wantAvailable: true,
		},
		{
			name:          "with empty token (no options)",
			opts:          nil,
			wantName:      "vault",
			wantAvailable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			provider := NewVaultProvider(tt.opts...)

			if provider.Name() != tt.wantName {
				t.Errorf("Name() = %q, want %q", provider.Name(), tt.wantName)
			}
			if provider.IsAvailable() != tt.wantAvailable {
				t.Errorf("IsAvailable() = %v, want %v", provider.IsAvailable(), tt.wantAvailable)
			}
		})
	}
}

// =============================================================================
// Integration Tests (with real providers)
// =============================================================================

func TestEnvProvider_Integration(t *testing.T) {
	ctx := context.Background()

	// Set up multiple test environment variables
	testSecrets := map[string]string{
		"DVM_SECRET_INTEGRATION_TEST_1": "value1",
		"DVM_SECRET_INTEGRATION_TEST_2": "value2",
	}

	for k, v := range testSecrets {
		os.Setenv(k, v)
		defer os.Unsetenv(k)
	}

	provider := NewEnvProvider()

	// Verify all secrets can be retrieved
	value1, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "integration-test-1"})
	if err != nil {
		t.Fatalf("failed to get integration-test-1: %v", err)
	}
	if value1 != "value1" {
		t.Errorf("integration-test-1: got %q, want %q", value1, "value1")
	}

	value2, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "integration-test-2"})
	if err != nil {
		t.Fatalf("failed to get integration-test-2: %v", err)
	}
	if value2 != "value2" {
		t.Errorf("integration-test-2: got %q, want %q", value2, "value2")
	}
}
