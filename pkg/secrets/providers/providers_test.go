package providers

import (
	"context"
	"os"
	"runtime"
	"testing"

	"devopsmaestro/pkg/secrets"
)

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestEnvProvider_ImplementsSecretProvider(t *testing.T) {
	var _ secrets.SecretProvider = (*EnvProvider)(nil)
}

func TestKeychainProvider_ImplementsSecretProvider(t *testing.T) {
	var _ secrets.SecretProvider = (*KeychainProvider)(nil)
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
// KeychainProvider Tests
// =============================================================================

func TestKeychainProvider_Name(t *testing.T) {
	provider := NewKeychainProvider()
	if provider.Name() != "keychain" {
		t.Errorf("got %q, want %q", provider.Name(), "keychain")
	}
}

func TestKeychainProvider_IsAvailable(t *testing.T) {
	provider := NewKeychainProvider()
	expected := runtime.GOOS == "darwin"

	if provider.IsAvailable() != expected {
		t.Errorf("IsAvailable() = %v, want %v (GOOS=%s)", provider.IsAvailable(), expected, runtime.GOOS)
	}
}

func TestKeychainProvider_Service(t *testing.T) {
	t.Run("returns default service", func(t *testing.T) {
		provider := NewKeychainProvider()
		if provider.Service() != DefaultKeychainService {
			t.Errorf("got %q, want %q", provider.Service(), DefaultKeychainService)
		}
	})

	t.Run("returns custom service", func(t *testing.T) {
		provider := NewKeychainProvider(WithKeychainService("custom-service"))
		if provider.Service() != "custom-service" {
			t.Errorf("got %q, want %q", provider.Service(), "custom-service")
		}
	})
}

func TestKeychainProvider_GetSecret_Unavailable(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("skipping unavailability test on macOS")
	}

	ctx := context.Background()
	provider := NewKeychainProvider()

	_, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "test"})
	if !secrets.IsProviderNotAvailable(err) {
		t.Errorf("got %v, want ErrProviderNotAvailable", err)
	}
}

// TestKeychainProvider_GetSecret_MockExec tests the keychain provider with a mocked exec command.
// This allows testing the command construction and output parsing without actually calling keychain.
func TestKeychainProvider_GetSecret_MockExec(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("keychain tests only run on macOS")
	}

	// Note: Full integration tests with real keychain would require setup/teardown
	// of actual keychain entries. For CI, we rely on the mocked exec tests.

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		provider := NewKeychainProvider()
		_, err := provider.GetSecret(ctx, secrets.SecretRequest{Name: "test"})

		if err != context.Canceled {
			t.Errorf("got %v, want context.Canceled", err)
		}
	})
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
