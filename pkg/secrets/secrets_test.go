package secrets

import (
	"context"
	"errors"
	"testing"
	"time"
)

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestMockProvider_ImplementsSecretProvider(t *testing.T) {
	var _ SecretProvider = (*MockProvider)(nil)
}

// =============================================================================
// MockProvider Tests
// =============================================================================

func TestMockProvider_GetSecret(t *testing.T) {
	ctx := context.Background()

	t.Run("returns configured secret", func(t *testing.T) {
		mock := NewMockProvider()
		mock.SetSecret("github-token", "ghp_test123")

		value, err := mock.GetSecret(ctx, SecretRequest{Name: "github-token"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != "ghp_test123" {
			t.Errorf("got %q, want %q", value, "ghp_test123")
		}
	})

	t.Run("returns ErrSecretNotFound for missing secret", func(t *testing.T) {
		mock := NewMockProvider()

		_, err := mock.GetSecret(ctx, SecretRequest{Name: "nonexistent"})
		if !errors.Is(err, ErrSecretNotFound) {
			t.Errorf("got %v, want ErrSecretNotFound", err)
		}
	})

	t.Run("returns configured error", func(t *testing.T) {
		mock := NewMockProvider()
		customErr := errors.New("permission denied")
		mock.SetError("restricted", customErr)

		_, err := mock.GetSecret(ctx, SecretRequest{Name: "restricted"})
		if !errors.Is(err, customErr) {
			t.Errorf("got %v, want %v", err, customErr)
		}
	})

	t.Run("records calls for verification", func(t *testing.T) {
		mock := NewMockProvider()
		mock.SetSecret("test", "value")

		mock.GetSecret(ctx, SecretRequest{Name: "test", Key: "field"})
		mock.GetSecret(ctx, SecretRequest{Name: "other"})

		calls := mock.GetCalls()
		if len(calls) != 2 {
			t.Fatalf("got %d calls, want 2", len(calls))
		}
		if calls[0].Name != "test" || calls[0].Key != "field" {
			t.Errorf("first call: got %+v, want name=test key=field", calls[0])
		}
		if calls[1].Name != "other" {
			t.Errorf("second call: got %+v, want name=other", calls[1])
		}
	})

	t.Run("handles structured secrets with key", func(t *testing.T) {
		mock := NewMockProvider()
		mock.SetSecret("db-config:password", "secret123")

		value, err := mock.GetSecret(ctx, SecretRequest{Name: "db-config", Key: "password"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != "secret123" {
			t.Errorf("got %q, want %q", value, "secret123")
		}
	})
}

func TestMockProvider_IsAvailable(t *testing.T) {
	t.Run("default is available", func(t *testing.T) {
		mock := NewMockProvider()
		if !mock.IsAvailable() {
			t.Error("expected mock to be available by default")
		}
	})

	t.Run("can be set unavailable", func(t *testing.T) {
		mock := NewMockProvider(WithMockAvailable(false))
		if mock.IsAvailable() {
			t.Error("expected mock to be unavailable")
		}
	})

	t.Run("can toggle availability", func(t *testing.T) {
		mock := NewMockProvider()
		mock.SetAvailable(false)
		if mock.IsAvailable() {
			t.Error("expected mock to be unavailable after SetAvailable(false)")
		}
		mock.SetAvailable(true)
		if !mock.IsAvailable() {
			t.Error("expected mock to be available after SetAvailable(true)")
		}
	})
}

// =============================================================================
// ProviderFactory Tests
// =============================================================================

func TestProviderFactory_Register(t *testing.T) {
	t.Run("registers provider", func(t *testing.T) {
		factory := NewProviderFactory()
		mock := NewMockProvider(WithMockName("test-provider"))

		factory.Register(mock)

		provider, err := factory.Get("test-provider")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if provider.Name() != "test-provider" {
			t.Errorf("got %q, want %q", provider.Name(), "test-provider")
		}
	})

	t.Run("first registered provider becomes default", func(t *testing.T) {
		factory := NewProviderFactory()
		mock := NewMockProvider(WithMockName("first"))

		factory.Register(mock)

		if factory.DefaultName() != "first" {
			t.Errorf("default: got %q, want %q", factory.DefaultName(), "first")
		}
	})

	t.Run("replaces existing provider with same name", func(t *testing.T) {
		factory := NewProviderFactory()
		mock1 := NewMockProvider(WithMockName("test"))
		mock1.SetSecret("key", "value1")
		mock2 := NewMockProvider(WithMockName("test"))
		mock2.SetSecret("key", "value2")

		factory.Register(mock1)
		factory.Register(mock2)

		provider, _ := factory.Get("test")
		value, _ := provider.GetSecret(context.Background(), SecretRequest{Name: "key"})
		if value != "value2" {
			t.Errorf("got %q, want %q (second mock)", value, "value2")
		}
	})
}

func TestProviderFactory_Get(t *testing.T) {
	t.Run("returns ErrProviderNotFound for unknown provider", func(t *testing.T) {
		factory := NewProviderFactory()

		_, err := factory.Get("nonexistent")
		if !IsProviderNotFound(err) {
			t.Errorf("got %v, want ErrProviderNotFound", err)
		}
	})

	t.Run("returns ErrProviderNotAvailable when not available", func(t *testing.T) {
		factory := NewProviderFactory()
		mock := NewMockProvider(WithMockName("unavailable"), WithMockAvailable(false))
		factory.Register(mock)

		_, err := factory.Get("unavailable")
		if !IsProviderNotAvailable(err) {
			t.Errorf("got %v, want ErrProviderNotAvailable", err)
		}
	})
}

func TestProviderFactory_GetDefault(t *testing.T) {
	t.Run("returns ErrNoDefaultProvider when empty", func(t *testing.T) {
		factory := NewProviderFactory()

		_, err := factory.GetDefault()
		if !errors.Is(err, ErrNoDefaultProvider) {
			t.Errorf("got %v, want ErrNoDefaultProvider", err)
		}
	})

	t.Run("returns first registered provider as default", func(t *testing.T) {
		factory := NewProviderFactory()
		factory.Register(NewMockProvider(WithMockName("first")))
		factory.Register(NewMockProvider(WithMockName("second")))

		provider, err := factory.GetDefault()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if provider.Name() != "first" {
			t.Errorf("got %q, want %q", provider.Name(), "first")
		}
	})

	t.Run("respects SetDefault", func(t *testing.T) {
		factory := NewProviderFactory()
		factory.Register(NewMockProvider(WithMockName("first")))
		factory.Register(NewMockProvider(WithMockName("second")))

		factory.SetDefault("second")

		provider, err := factory.GetDefault()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if provider.Name() != "second" {
			t.Errorf("got %q, want %q", provider.Name(), "second")
		}
	})
}

func TestProviderFactory_SetDefault(t *testing.T) {
	t.Run("returns error for unregistered provider", func(t *testing.T) {
		factory := NewProviderFactory()

		err := factory.SetDefault("nonexistent")
		if !IsProviderNotFound(err) {
			t.Errorf("got %v, want ErrProviderNotFound", err)
		}
	})
}

func TestProviderFactory_List(t *testing.T) {
	factory := NewProviderFactory()
	factory.Register(NewMockProvider(WithMockName("alpha")))
	factory.Register(NewMockProvider(WithMockName("beta")))

	names := factory.List()
	if len(names) != 2 {
		t.Fatalf("got %d providers, want 2", len(names))
	}

	// Check both are present (order not guaranteed)
	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}
	if !found["alpha"] || !found["beta"] {
		t.Errorf("got %v, want alpha and beta", names)
	}
}

func TestProviderFactory_ListAvailable(t *testing.T) {
	factory := NewProviderFactory()
	factory.Register(NewMockProvider(WithMockName("available"), WithMockAvailable(true)))
	factory.Register(NewMockProvider(WithMockName("unavailable"), WithMockAvailable(false)))

	names := factory.ListAvailable()
	if len(names) != 1 {
		t.Fatalf("got %d available providers, want 1", len(names))
	}
	if names[0] != "available" {
		t.Errorf("got %q, want %q", names[0], "available")
	}
}

// =============================================================================
// Cache Tests
// =============================================================================

func TestCache_GetSet(t *testing.T) {
	t.Run("stores and retrieves values", func(t *testing.T) {
		cache := NewCache()
		cache.Set("provider", "secret-name", "", "secret-value")

		value, ok := cache.Get("provider", "secret-name", "")
		if !ok {
			t.Fatal("expected cache hit")
		}
		if value != "secret-value" {
			t.Errorf("got %q, want %q", value, "secret-value")
		}
	})

	t.Run("returns false for missing entry", func(t *testing.T) {
		cache := NewCache()

		_, ok := cache.Get("provider", "nonexistent", "")
		if ok {
			t.Error("expected cache miss")
		}
	})

	t.Run("handles key in cache key", func(t *testing.T) {
		cache := NewCache()
		cache.Set("provider", "secret", "field1", "value1")
		cache.Set("provider", "secret", "field2", "value2")

		value1, ok := cache.Get("provider", "secret", "field1")
		if !ok || value1 != "value1" {
			t.Errorf("field1: got %q, ok=%v, want %q", value1, ok, "value1")
		}

		value2, ok := cache.Get("provider", "secret", "field2")
		if !ok || value2 != "value2" {
			t.Errorf("field2: got %q, ok=%v, want %q", value2, ok, "value2")
		}
	})
}

func TestCache_TTL(t *testing.T) {
	// Use a fixed time for testing
	now := time.Now()
	currentTime := now

	cache := NewCache(
		WithTTL(5*time.Minute),
		withNowFunc(func() time.Time { return currentTime }),
	)

	cache.Set("provider", "secret", "", "value")

	// Should be cached initially
	if _, ok := cache.Get("provider", "secret", ""); !ok {
		t.Fatal("expected cache hit immediately after set")
	}

	// Advance time past TTL
	currentTime = now.Add(6 * time.Minute)

	// Should be expired now
	if _, ok := cache.Get("provider", "secret", ""); ok {
		t.Error("expected cache miss after TTL expiration")
	}
}

func TestCache_Clear(t *testing.T) {
	cache := NewCache()
	cache.Set("p1", "s1", "", "v1")
	cache.Set("p2", "s2", "", "v2")

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("got size %d, want 0 after clear", cache.Size())
	}
}

func TestCache_Delete(t *testing.T) {
	cache := NewCache()
	cache.Set("provider", "keep", "", "value1")
	cache.Set("provider", "delete", "", "value2")

	cache.Delete("provider", "delete", "")

	if _, ok := cache.Get("provider", "delete", ""); ok {
		t.Error("expected deleted entry to be missing")
	}
	if _, ok := cache.Get("provider", "keep", ""); !ok {
		t.Error("expected kept entry to still exist")
	}
}

func TestCache_Prune(t *testing.T) {
	now := time.Now()
	currentTime := now

	cache := NewCache(
		WithTTL(5*time.Minute),
		withNowFunc(func() time.Time { return currentTime }),
	)

	cache.Set("provider", "expired", "", "value1")
	currentTime = now.Add(6 * time.Minute)
	cache.Set("provider", "fresh", "", "value2")

	cache.Prune()

	if cache.Size() != 1 {
		t.Errorf("got size %d, want 1 after prune", cache.Size())
	}
	if _, ok := cache.Get("provider", "fresh", ""); !ok {
		t.Error("expected fresh entry to survive prune")
	}
}

// =============================================================================
// Resolver Tests
// =============================================================================

func TestResolver_ResolveInline(t *testing.T) {
	ctx := context.Background()

	t.Run("resolves single secret with default provider", func(t *testing.T) {
		factory := NewProviderFactory()
		mock := NewMockProvider()
		mock.SetSecret("github-token", "ghp_test123")
		factory.Register(mock)

		resolver := NewResolver(factory)
		result, err := resolver.ResolveInline(ctx, "token: ${secret:github-token}")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "token: ghp_test123" {
			t.Errorf("got %q, want %q", result, "token: ghp_test123")
		}
	})

	t.Run("resolves secret with explicit provider", func(t *testing.T) {
		factory := NewProviderFactory()
		envMock := NewMockProvider(WithMockName("env"))
		envMock.SetSecret("api-key", "env-value")
		keychainMock := NewMockProvider(WithMockName("keychain"))
		keychainMock.SetSecret("api-key", "keychain-value")
		factory.Register(envMock)
		factory.Register(keychainMock)

		resolver := NewResolver(factory)

		result, err := resolver.ResolveInline(ctx, "key: ${secret:api-key:keychain}")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "key: keychain-value" {
			t.Errorf("got %q, want %q", result, "key: keychain-value")
		}
	})

	t.Run("resolves multiple secrets", func(t *testing.T) {
		factory := NewProviderFactory()
		mock := NewMockProvider()
		mock.SetSecret("user", "admin")
		mock.SetSecret("pass", "secret")
		factory.Register(mock)

		resolver := NewResolver(factory)
		result, err := resolver.ResolveInline(ctx, "user=${secret:user} pass=${secret:pass}")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "user=admin pass=secret" {
			t.Errorf("got %q, want %q", result, "user=admin pass=secret")
		}
	})

	t.Run("returns error for missing secret", func(t *testing.T) {
		factory := NewProviderFactory()
		mock := NewMockProvider()
		factory.Register(mock)

		resolver := NewResolver(factory)
		_, err := resolver.ResolveInline(ctx, "token: ${secret:missing}")

		if err == nil {
			t.Fatal("expected error for missing secret")
		}
	})

	t.Run("returns content unchanged without secrets", func(t *testing.T) {
		factory := NewProviderFactory()
		factory.Register(NewMockProvider())

		resolver := NewResolver(factory)
		content := "no secrets here: ${other:pattern}"
		result, err := resolver.ResolveInline(ctx, content)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != content {
			t.Errorf("got %q, want %q", result, content)
		}
	})
}

func TestResolver_ResolveReference(t *testing.T) {
	ctx := context.Background()

	t.Run("resolves reference with default provider", func(t *testing.T) {
		factory := NewProviderFactory()
		mock := NewMockProvider()
		mock.SetSecret("db-password", "supersecret")
		factory.Register(mock)

		resolver := NewResolver(factory)
		value, err := resolver.ResolveReference(ctx, SecretReference{Name: "db-password"})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != "supersecret" {
			t.Errorf("got %q, want %q", value, "supersecret")
		}
	})

	t.Run("returns error for empty name", func(t *testing.T) {
		factory := NewProviderFactory()
		factory.Register(NewMockProvider())

		resolver := NewResolver(factory)
		_, err := resolver.ResolveReference(ctx, SecretReference{Name: ""})

		if !errors.Is(err, ErrInvalidReference) {
			t.Errorf("got %v, want ErrInvalidReference", err)
		}
	})
}

func TestResolver_Caching(t *testing.T) {
	ctx := context.Background()

	factory := NewProviderFactory()
	mock := NewMockProvider()
	mock.SetSecret("cached", "value")
	factory.Register(mock)

	resolver := NewResolver(factory)

	// First call
	resolver.ResolveInline(ctx, "${secret:cached}")
	// Second call (should use cache)
	resolver.ResolveInline(ctx, "${secret:cached}")

	calls := mock.GetCalls()
	if len(calls) != 1 {
		t.Errorf("got %d provider calls, want 1 (cached)", len(calls))
	}
}

func TestResolver_ClearCache(t *testing.T) {
	ctx := context.Background()

	factory := NewProviderFactory()
	mock := NewMockProvider()
	mock.SetSecret("test", "value")
	factory.Register(mock)

	resolver := NewResolver(factory)

	resolver.ResolveInline(ctx, "${secret:test}")
	resolver.ClearCache()
	resolver.ResolveInline(ctx, "${secret:test}")

	calls := mock.GetCalls()
	if len(calls) != 2 {
		t.Errorf("got %d provider calls, want 2 (cache cleared)", len(calls))
	}
}

// =============================================================================
// Helper Function Tests
// =============================================================================

func TestHasSecretReferences(t *testing.T) {
	tests := []struct {
		content string
		want    bool
	}{
		{"${secret:name}", true},
		{"${secret:name:provider}", true},
		{"prefix ${secret:name} suffix", true},
		{"no secrets here", false},
		{"${other:pattern}", false},
		{"${secret}", false}, // Missing name
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			got := HasSecretReferences(tt.content)
			if got != tt.want {
				t.Errorf("HasSecretReferences(%q) = %v, want %v", tt.content, got, tt.want)
			}
		})
	}
}

func TestExtractSecretReferences(t *testing.T) {
	content := "a=${secret:first} b=${secret:second:keychain}"
	refs := ExtractSecretReferences(content)

	if len(refs) != 2 {
		t.Fatalf("got %d refs, want 2", len(refs))
	}

	if refs[0].Name != "first" || refs[0].Provider != "" {
		t.Errorf("first ref: got %+v, want name=first provider=", refs[0])
	}
	if refs[1].Name != "second" || refs[1].Provider != "keychain" {
		t.Errorf("second ref: got %+v, want name=second provider=keychain", refs[1])
	}
}

func TestConvertNameToEnvVar(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"github-token", "GITHUB_TOKEN"},
		{"api.key", "API_KEY"},
		{"simple", "SIMPLE"},
		{"ALREADY_UPPER", "ALREADY_UPPER"},
		{"mixed-case.name", "MIXED_CASE_NAME"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ConvertNameToEnvVar(tt.input)
			if got != tt.want {
				t.Errorf("ConvertNameToEnvVar(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// =============================================================================
// Error Tests
// =============================================================================

func TestIsNotFound(t *testing.T) {
	if !IsNotFound(ErrSecretNotFound) {
		t.Error("IsNotFound(ErrSecretNotFound) should return true")
	}

	wrapped := &SecretError{Name: "test", Err: ErrSecretNotFound}
	if !IsNotFound(wrapped) {
		t.Error("IsNotFound should work with wrapped errors")
	}

	if IsNotFound(errors.New("other error")) {
		t.Error("IsNotFound should return false for other errors")
	}
}
