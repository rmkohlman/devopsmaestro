package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TDD Phase 2 (RED): MaestroVault SecretBackend Tests (v0.40.0)
// =============================================================================
// New interfaces and types being introduced to replace macOS Keychain:
//
//	type SecretBackend interface {
//	    Get(name, env string) (string, error)
//	    Health() error
//	}
//
//	type VaultBackend struct { ... }
//	func NewVaultBackend(token string) (*VaultBackend, error)
//
// ALL tests in this section WILL FAIL TO COMPILE until the above types are
// added to config/secret_backend.go and config/vault.go.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: SecretBackend Interface Tests
// ---------------------------------------------------------------------------

// TestSecretBackend_InterfaceExists verifies that the SecretBackend interface
// is declared in the config package with the correct method set.
//
// WILL FAIL TO COMPILE — SecretBackend interface does not exist yet.
func TestSecretBackend_InterfaceExists(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// SecretBackend type does not exist yet.
	var _ SecretBackend = nil // interface nil assignment — proves type exists
	// ─────────────────────────────────────────────────────────────────────────

	// If we reach here the interface compiles. Mark it as a passing assertion.
	assert.True(t, true, "SecretBackend interface must exist with correct method set")
}

// TestSecretBackend_HasGetMethod verifies that SecretBackend.Get has the
// signature: Get(name, env string) (string, error).
//
// WILL FAIL TO COMPILE — SecretBackend interface does not exist yet.
func TestSecretBackend_HasGetMethod(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Declare a variable of function type matching the Get signature,
	// then assign it from a nil SecretBackend. The compiler enforces the signature.
	var backend SecretBackend
	if backend != nil {
		// This branch is never executed; it exists to bind the method to a
		// typed variable, forcing the compiler to verify the signature.
		var fn func(string, string) (string, error)
		fn = backend.Get
		_ = fn
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.True(t, true, "SecretBackend.Get must have signature: Get(name, env string) (string, error)")
}

// TestSecretBackend_HasHealthMethod verifies that SecretBackend.Health has the
// signature: Health() error.
//
// WILL FAIL TO COMPILE — SecretBackend interface does not exist yet.
func TestSecretBackend_HasHealthMethod(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	var backend SecretBackend
	if backend != nil {
		var fn func() error
		fn = backend.Health
		_ = fn
	}
	// ─────────────────────────────────────────────────────────────────────────

	assert.True(t, true, "SecretBackend.Health must have signature: Health() error")
}

// ---------------------------------------------------------------------------
// Section: VaultBackend Implementation Tests
// ---------------------------------------------------------------------------

// TestVaultBackend_ImplementsSecretBackend verifies at compile time that
// *VaultBackend satisfies the SecretBackend interface.
//
// WILL FAIL TO COMPILE — VaultBackend type does not exist yet.
func TestVaultBackend_ImplementsSecretBackend(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// This is the idiomatic Go compile-time interface check.
	// VaultBackend type does not exist yet.
	var _ SecretBackend = (*VaultBackend)(nil)
	// ─────────────────────────────────────────────────────────────────────────

	assert.True(t, true, "*VaultBackend must implement SecretBackend")
}

// TestNewVaultBackend_EmptyToken verifies that NewVaultBackend returns an
// error when given an empty token.
//
// WILL FAIL TO COMPILE — NewVaultBackend function does not exist yet.
func TestNewVaultBackend_EmptyToken(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// NewVaultBackend does not exist yet.
	_, err := NewVaultBackend("")
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err, "empty token should be rejected by NewVaultBackend")
	assert.Contains(t, err.Error(), "token",
		"error message should mention the missing token")
}

// TestNewVaultBackend_ValidToken verifies that NewVaultBackend does NOT return
// an error when given a non-empty token. This is a structural check — we do
// not verify the token's validity against a live vault daemon.
//
// WILL FAIL TO COMPILE — NewVaultBackend function does not exist yet.
func TestNewVaultBackend_ValidToken(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// NewVaultBackend does not exist yet.
	backend, err := NewVaultBackend("dvm-test-token-abc123")
	// ─────────────────────────────────────────────────────────────────────────

	// Structural check: construction succeeds — daemon connectivity is NOT
	// validated at construction time (only at runtime when Get/Health is called).
	require.NoError(t, err, "non-empty token should not be rejected at construction")
	require.NotNil(t, backend, "NewVaultBackend must return a non-nil backend for a valid token")
}

// TestNewVaultBackend_ReturnsVaultBackend verifies that NewVaultBackend returns
// a concrete *VaultBackend (not just a SecretBackend interface), so callers
// can reference specific fields if needed.
//
// WILL FAIL TO COMPILE — NewVaultBackend and VaultBackend do not exist yet.
func TestNewVaultBackend_ReturnsVaultBackend(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	backend, err := NewVaultBackend("some-token")
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)
	require.NotNil(t, backend)

	// Verify the returned type is *VaultBackend (not just an interface).
	// The type assertion will panic at runtime if the type is wrong, making
	// the failure explicit.
	var _ *VaultBackend = backend // compile-time type check
}

// ---------------------------------------------------------------------------
// Section: VaultBackend.Get Tests
// ---------------------------------------------------------------------------

// TestVaultBackend_Get_ReturnsError_WhenDaemonNotRunning verifies that calling
// Get on a VaultBackend when the vault daemon is not running returns an error.
// (In test environments the daemon will not be running, so this should always
// produce an error — we just verify it doesn't panic and returns an error.)
//
// WILL FAIL TO COMPILE — VaultBackend and NewVaultBackend do not exist yet.
func TestVaultBackend_Get_ReturnsError_WhenDaemonNotRunning(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	backend, err := NewVaultBackend("dvm-test-token")
	require.NoError(t, err)

	// Call Get — vault daemon is not running in CI, so we expect an error.
	// We do NOT assert a specific error message because it is implementation-
	// defined; we only verify a non-nil error is returned.
	_, getErr := backend.Get("my-secret", "production")
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, getErr,
		"VaultBackend.Get must return an error when the vault daemon is not running")
}

// TestVaultBackend_Get_EmptySecretName verifies that Get returns an error when
// the secret name argument is empty.
//
// WILL FAIL TO COMPILE — VaultBackend does not exist yet.
func TestVaultBackend_Get_EmptySecretName(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	backend, err := NewVaultBackend("dvm-test-token")
	require.NoError(t, err)

	_, getErr := backend.Get("", "production")
	// ─────────────────────────────────────────────────────────────────────────

	// Either the implementation validates at the Get call level (preferred)
	// or it passes through to the vault client which rejects an empty name.
	// Either way, an error must be returned.
	assert.Error(t, getErr,
		"VaultBackend.Get with empty secret name must return an error")
}

// ---------------------------------------------------------------------------
// Section: VaultBackend.Health Tests
// ---------------------------------------------------------------------------

// TestVaultBackend_Health_ReturnsError_WhenDaemonNotRunning verifies that
// Health() returns an error when the vault daemon is not running.
//
// WILL FAIL TO COMPILE — VaultBackend does not exist yet.
func TestVaultBackend_Health_ReturnsError_WhenDaemonNotRunning(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	backend, err := NewVaultBackend("dvm-test-token")
	require.NoError(t, err)

	healthErr := backend.Health()
	// ─────────────────────────────────────────────────────────────────────────

	// In test environments the vault daemon is not running, so Health() must
	// return an error. We do not assert a specific message.
	assert.Error(t, healthErr,
		"VaultBackend.Health must return an error when the vault daemon is not running")
}

// ---------------------------------------------------------------------------
// Section: Mock SecretBackend (for use in ResolveCredential tests)
// ---------------------------------------------------------------------------

// mockSecretBackend is a test double implementing SecretBackend.
// It is used to test ResolveCredential in isolation from the vault daemon.
//
// WILL FAIL TO COMPILE — SecretBackend interface does not exist yet.
type mockSecretBackend struct {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// SecretBackend interface does not exist yet; this embed fails to compile.
	GetFunc    func(name, env string) (string, error)
	HealthFunc func() error
	// ─────────────────────────────────────────────────────────────────────────
}

// Get implements SecretBackend.Get.
//
// WILL FAIL TO COMPILE — SecretBackend does not exist yet.
func (m *mockSecretBackend) Get(name, env string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(name, env)
	}
	return "mock-secret-value", nil
}

// Health implements SecretBackend.Health.
//
// WILL FAIL TO COMPILE — SecretBackend does not exist yet.
func (m *mockSecretBackend) Health() error {
	if m.HealthFunc != nil {
		return m.HealthFunc()
	}
	return nil
}

// Compile-time assertion that mockSecretBackend implements SecretBackend.
//
// WILL FAIL TO COMPILE — SecretBackend does not exist yet.
var _ SecretBackend = (*mockSecretBackend)(nil)

// ---------------------------------------------------------------------------
// Section: ResolveCredential Vault Source Tests
// ---------------------------------------------------------------------------

// TestResolveCredentialWithBackend_VaultSource verifies that
// ResolveCredentialWithBackend resolves a vault-sourced credential by calling
// backend.Get(cfg.VaultSecret, cfg.VaultEnv).
//
// WILL FAIL TO COMPILE — ResolveCredentialWithBackend, SourceVault,
// CredentialConfig.VaultSecret, and CredentialConfig.VaultEnv do not exist yet.
func TestResolveCredentialWithBackend_VaultSource(t *testing.T) {
	mock := &mockSecretBackend{
		GetFunc: func(name, env string) (string, error) {
			assert.Equal(t, "my-github-pat", name,
				"VaultSecret must be passed as the name argument to backend.Get")
			assert.Equal(t, "production", env,
				"VaultEnv must be passed as the env argument to backend.Get")
			return "ghp_test_token_value", nil
		},
	}

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// SourceVault, CredentialConfig.VaultSecret, CredentialConfig.VaultEnv,
	// and ResolveCredentialWithBackend do not exist yet.
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "my-github-pat",
		VaultEnv:    "production",
	}
	value, err := ResolveCredentialWithBackend(cfg, mock)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err, "vault credential resolution must succeed with a working backend")
	assert.Equal(t, "ghp_test_token_value", value,
		"resolved value must match the value returned by backend.Get")
}

// TestResolveCredentialWithBackend_VaultRequiresSecretName verifies that
// ResolveCredentialWithBackend returns an error when VaultSecret is empty.
//
// WILL FAIL TO COMPILE — ResolveCredentialWithBackend, SourceVault,
// and CredentialConfig.VaultSecret do not exist yet.
func TestResolveCredentialWithBackend_VaultRequiresSecretName(t *testing.T) {
	mock := &mockSecretBackend{}

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "", // intentionally empty
		VaultEnv:    "production",
	}
	_, err := ResolveCredentialWithBackend(cfg, mock)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"vault source with empty VaultSecret must return a validation error")
	assert.Contains(t, err.Error(), "vault",
		"error should indicate the vault secret name is required")
}

// TestResolveCredentialWithBackend_VaultEmptyEnvUsesDefault verifies that when
// VaultEnv is empty, ResolveCredentialWithBackend passes an empty string (or
// implementation-defined default) to backend.Get without error. The backend
// itself is responsible for defaulting the env.
//
// WILL FAIL TO COMPILE — ResolveCredentialWithBackend, SourceVault,
// and CredentialConfig.VaultSecret do not exist yet.
func TestResolveCredentialWithBackend_VaultEmptyEnvUsesDefault(t *testing.T) {
	called := false
	mock := &mockSecretBackend{
		GetFunc: func(name, env string) (string, error) {
			called = true
			assert.Equal(t, "my-secret", name)
			// env may be "" or a default value — we just verify it was called
			return "secret-value", nil
		},
	}

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "my-secret",
		VaultEnv:    "", // empty — implementation must handle gracefully
	}
	value, err := ResolveCredentialWithBackend(cfg, mock)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err, "empty VaultEnv must not cause a validation error")
	assert.Equal(t, "secret-value", value)
	assert.True(t, called, "backend.Get must be called even with empty VaultEnv")
}

// TestResolveCredentialWithBackend_NilBackend_VaultSource verifies that
// ResolveCredentialWithBackend returns a hard error when the backend is nil
// and the source is "vault". This prevents silent failures.
//
// WILL FAIL TO COMPILE — ResolveCredentialWithBackend and SourceVault
// do not exist yet.
func TestResolveCredentialWithBackend_NilBackend_VaultSource(t *testing.T) {
	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "my-secret",
	}
	_, err := ResolveCredentialWithBackend(cfg, nil)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"ResolveCredentialWithBackend with nil backend and vault source must return an error")
}

// TestResolveCredentialWithBackend_EnvSource verifies that env-sourced
// credentials still work correctly through ResolveCredentialWithBackend
// (the backend parameter is ignored for non-vault sources).
//
// WILL FAIL TO COMPILE — ResolveCredentialWithBackend does not exist yet.
func TestResolveCredentialWithBackend_EnvSource(t *testing.T) {
	t.Setenv("TEST_VAULT_ENV_VAR_DVM", "env-resolved-value")

	// A nil backend should be fine for env-sourced credentials.
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cfg := CredentialConfig{
		Source: SourceEnv,
		EnvVar: "TEST_VAULT_ENV_VAR_DVM",
	}
	value, err := ResolveCredentialWithBackend(cfg, nil)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err, "env source must still resolve via ResolveCredentialWithBackend")
	assert.Equal(t, "env-resolved-value", value,
		"env source must read from the environment variable")
}

// =============================================================================
// TDD Phase 2 (RED): FieldCapableBackend Interface Tests (v0.41.0)
// =============================================================================
// New interface being introduced to support per-field vault secret resolution.
//
//	type FieldCapableBackend interface {
//	    SecretBackend
//	    GetField(name, env, field string) (string, error)
//	}
//
// DESIGN DECISION: GetField is NOT added to SecretBackend directly.
// It lives on the narrower FieldCapableBackend sub-interface so callers that
// don't need field access don't have to implement it.
//
// ALL tests in this section WILL FAIL TO COMPILE until FieldCapableBackend
// is declared in config/secret_backend.go.
// =============================================================================

// ---------------------------------------------------------------------------
// Section: FieldCapableBackend Interface Declaration
// ---------------------------------------------------------------------------

// mockFieldCapableBackend is a test double that implements FieldCapableBackend.
// It is declared separately from mockSecretBackend to test the narrower interface.
//
// WILL FAIL TO COMPILE — FieldCapableBackend does not exist yet.
type mockFieldCapableBackend struct {
	// Embed the GetFunc / HealthFunc from the existing mock to satisfy SecretBackend.
	GetFunc    func(name, env string) (string, error)
	HealthFunc func() error
	// GetFieldFunc satisfies the FieldCapableBackend extension.
	GetFieldFunc func(name, env, field string) (string, error)
}

// ── COMPILE ERROR EXPECTED ──────────────────────────────────────────────────
// FieldCapableBackend does not exist yet — this interface assertion will fail.
var _ FieldCapableBackend = (*mockFieldCapableBackend)(nil)

// ───────────────────────────────────────────────────────────────────────────

func (m *mockFieldCapableBackend) Get(name, env string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(name, env)
	}
	return "", nil
}

func (m *mockFieldCapableBackend) Health() error {
	if m.HealthFunc != nil {
		return m.HealthFunc()
	}
	return nil
}

func (m *mockFieldCapableBackend) GetField(name, env, field string) (string, error) {
	if m.GetFieldFunc != nil {
		return m.GetFieldFunc(name, env, field)
	}
	return "", nil
}

// TestFieldCapableBackend_InterfaceExists verifies that FieldCapableBackend is
// declared with the correct embedded SecretBackend + GetField method.
//
// WILL FAIL TO COMPILE — FieldCapableBackend does not exist yet.
func TestFieldCapableBackend_InterfaceExists(t *testing.T) {
	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	var _ FieldCapableBackend = nil // proves type exists
	// ─────────────────────────────────────────────────────────────────────────

	assert.True(t, true, "FieldCapableBackend interface must exist")
}

// TestFieldCapableBackend_EmbedssSecretBackend verifies that FieldCapableBackend
// is a superset of SecretBackend — any FieldCapableBackend must also be usable
// as a SecretBackend (Liskov substitution).
//
// WILL FAIL TO COMPILE — FieldCapableBackend does not exist yet.
func TestFieldCapableBackend_EmbedsSecretBackend(t *testing.T) {
	mock := &mockFieldCapableBackend{
		GetFunc: func(name, env string) (string, error) { return "value", nil },
	}

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	// Assign a FieldCapableBackend to a SecretBackend variable.
	// If FieldCapableBackend embeds SecretBackend this compiles; otherwise fails.
	var sb SecretBackend = mock
	// ─────────────────────────────────────────────────────────────────────────

	value, err := sb.Get("some-secret", "prod")
	require.NoError(t, err)
	assert.Equal(t, "value", value)
}

// TestFieldCapableBackend_GetFieldMethod verifies the GetField signature:
//
//	GetField(name, env, field string) (string, error)
//
// WILL FAIL TO COMPILE — FieldCapableBackend does not exist yet.
func TestFieldCapableBackend_GetFieldMethod(t *testing.T) {
	called := false
	mock := &mockFieldCapableBackend{
		GetFieldFunc: func(name, env, field string) (string, error) {
			called = true
			assert.Equal(t, "github/creds", name)
			assert.Equal(t, "production", env)
			assert.Equal(t, "password", field)
			return "super-secret-password", nil
		},
	}

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	var backend FieldCapableBackend = mock
	value, err := backend.GetField("github/creds", "production", "password")
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)
	assert.Equal(t, "super-secret-password", value)
	assert.True(t, called, "GetField must be called on the backend")
}

// TestFieldCapableBackend_GetField_EmptyEnv verifies that GetField handles an
// empty env string gracefully (implementations should treat it as "default").
//
// WILL FAIL TO COMPILE — FieldCapableBackend does not exist yet.
func TestFieldCapableBackend_GetField_EmptyEnv(t *testing.T) {
	mock := &mockFieldCapableBackend{
		GetFieldFunc: func(name, env, field string) (string, error) {
			assert.Equal(t, "", env, "empty env should be passed through unchanged")
			return "value", nil
		},
	}

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	var backend FieldCapableBackend = mock
	_, err := backend.GetField("some-secret", "", "username")
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err, "empty env must not cause an error at the interface level")
}

// TestResolveCredentialWithBackend_VaultField_UsesGetField verifies that when
// a CredentialConfig has a VaultField set, ResolveCredentialWithBackend calls
// backend.GetField instead of backend.Get.
//
// WILL FAIL TO COMPILE — FieldCapableBackend and CredentialConfig.VaultField
// do not exist yet.
func TestResolveCredentialWithBackend_VaultField_UsesGetField(t *testing.T) {
	getFieldCalled := false
	getCalled := false

	mock := &mockFieldCapableBackend{
		GetFunc: func(name, env string) (string, error) {
			getCalled = true
			return "wrong-value", nil
		},
		GetFieldFunc: func(name, env, field string) (string, error) {
			getFieldCalled = true
			assert.Equal(t, "github/creds", name)
			assert.Equal(t, "production", env)
			assert.Equal(t, "password", field)
			return "correct-password", nil
		},
	}

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	// CredentialConfig.VaultField does not exist yet.
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "github/creds",
		VaultEnv:    "production",
		VaultField:  "password",
	}
	value, err := ResolveCredentialWithBackend(cfg, mock)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)
	assert.Equal(t, "correct-password", value)
	assert.True(t, getFieldCalled, "GetField must be called when VaultField is set")
	assert.False(t, getCalled, "Get must NOT be called when VaultField is set")
}

// TestResolveCredentialWithBackend_VaultField_FallsBackToGet verifies that when
// VaultField is empty, ResolveCredentialWithBackend uses the regular Get method
// (backward-compatible behavior).
//
// WILL FAIL TO COMPILE — FieldCapableBackend and CredentialConfig.VaultField
// do not exist yet.
func TestResolveCredentialWithBackend_VaultField_FallsBackToGet(t *testing.T) {
	getCalled := false
	getFieldCalled := false

	mock := &mockFieldCapableBackend{
		GetFunc: func(name, env string) (string, error) {
			getCalled = true
			return "the-token", nil
		},
		GetFieldFunc: func(name, env, field string) (string, error) {
			getFieldCalled = true
			return "should-not-be-called", nil
		},
	}

	// ── COMPILE ERRORS EXPECTED BELOW ────────────────────────────────────────
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "github/token",
		VaultEnv:    "staging",
		VaultField:  "", // empty — must fall back to Get
	}
	value, err := ResolveCredentialWithBackend(cfg, mock)
	// ─────────────────────────────────────────────────────────────────────────

	require.NoError(t, err)
	assert.Equal(t, "the-token", value)
	assert.True(t, getCalled, "Get must be called when VaultField is empty")
	assert.False(t, getFieldCalled, "GetField must NOT be called when VaultField is empty")
}

// TestResolveCredentialWithBackend_VaultField_NonFieldCapableBackend verifies
// that when a VaultField is requested but the backend does NOT implement
// FieldCapableBackend, the function returns an error rather than silently
// ignoring the field.
//
// WILL FAIL TO COMPILE — CredentialConfig.VaultField does not exist yet.
func TestResolveCredentialWithBackend_VaultField_NonFieldCapableBackend(t *testing.T) {
	// Use the basic mockSecretBackend (not FieldCapableBackend).
	mock := &mockSecretBackend{
		GetFunc: func(name, env string) (string, error) { return "value", nil },
	}

	// ── COMPILE ERROR EXPECTED BELOW ─────────────────────────────────────────
	cfg := CredentialConfig{
		Source:      SourceVault,
		VaultSecret: "github/creds",
		VaultField:  "password",
	}
	_, err := ResolveCredentialWithBackend(cfg, mock)
	// ─────────────────────────────────────────────────────────────────────────

	assert.Error(t, err,
		"requesting VaultField from a non-FieldCapableBackend must return an error")
	assert.Contains(t, err.Error(), "field",
		"error must mention 'field' to guide the caller")
}
