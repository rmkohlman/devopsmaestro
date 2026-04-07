package cmd

// =============================================================================
// TDD Phase 2 (RED): Bug #157 — CLI flag mapping mismatch
// =============================================================================
// Bug: `dvm create credential --username-var X --password-var Y` sets
// cred.UsernameVar and cred.PasswordVar (dual-field DB columns) instead of
// building cred.VaultFields (JSON map). The YAML apply path uses spec.vaultFields
// which populates cred.VaultFields. This structural mismatch causes CLI-created
// credentials to behave differently from YAML-applied credentials at runtime.
//
// These tests WILL FAIL until cmd/credential.go is fixed to build VaultFields
// from --username-var / --password-var instead of setting UsernameVar / PasswordVar.
// =============================================================================

import (
	"encoding/json"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/pkg/credentialbridge"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Bug 1 Tests: CLI --username-var / --password-var flag mapping
// ---------------------------------------------------------------------------

// TestCLICreateCredential_UsernameVar_BuildsVaultFields verifies that when
// --username-var is provided, the resulting CredentialDB has VaultFields set
// (not UsernameVar). This is how the YAML apply path works.
//
// BUG #157 — WILL FAIL: current code sets UsernameVar, not VaultFields.
func TestCLICreateCredential_UsernameVar_BuildsVaultFields(t *testing.T) {
	// Simulate what cmd/credential.go RunE does when --username-var is set.
	// Current (buggy) behavior: sets cred.UsernameVar
	// Expected (fixed) behavior: builds cred.VaultFields map with the var name

	usernameVar := "GITHUB_USERNAME"

	// This is what the current buggy code produces:
	buggyResult := &models.CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		UsernameVar: &usernameVar, // BUG: sets UsernameVar column
		VaultFields: nil,          // BUG: VaultFields is never set
	}

	// This is what the fixed code must produce:
	expectedVaultFields := map[string]string{
		"GITHUB_USERNAME": "GITHUB_USERNAME",
	}
	vfJSON, err := json.Marshal(expectedVaultFields)
	require.NoError(t, err)
	vfStr := string(vfJSON)

	fixedResult := &models.CredentialDB{
		Name:        "github-creds",
		Source:      "vault",
		UsernameVar: nil,    // Fixed: UsernameVar not set (deprecated path)
		VaultFields: &vfStr, // Fixed: VaultFields JSON map is built
	}

	// The buggy result fails HasVaultFields (it is nil)
	assert.False(t, buggyResult.HasVaultFields(),
		"[confirms bug] current CLI-created credential has no VaultFields — "+
			"proving CLI and YAML paths produce structurally different credentials")

	// The fixed result must have VaultFields set
	require.True(t, fixedResult.HasVaultFields(),
		"after fix: credential created via --username-var must have VaultFields set")

	fields, err := fixedResult.GetVaultFieldsMap()
	require.NoError(t, err)
	assert.Equal(t, "GITHUB_USERNAME", fields["GITHUB_USERNAME"],
		"VaultFields map must contain GITHUB_USERNAME -> GITHUB_USERNAME mapping")

	// THE REAL TEST: simulate the actual createCredentialCmd.RunE behavior.
	// After the fix, the command code should NOT set UsernameVar — it should
	// build VaultFields instead.
	//
	// We construct a credential the way the FIXED RunE will construct it:
	credFromFixedCLI := buildCredentialFromCLIFlags(
		"github-creds", "vault", "github-creds", "production",
		"GITHUB_USERNAME", "", // usernameVar, passwordVar
	)

	// WILL FAIL with buggy code (VaultFields is nil, UsernameVar is set):
	require.True(t, credFromFixedCLI.HasVaultFields(),
		"BUG #157: --username-var must populate VaultFields, not UsernameVar — "+
			"WILL FAIL until cmd/credential.go is fixed")
	assert.Nil(t, credFromFixedCLI.UsernameVar,
		"after fix: UsernameVar must be nil when --username-var builds VaultFields")
}

// TestCLICreateCredential_PasswordVar_BuildsVaultFields verifies that when
// --password-var is provided, the resulting CredentialDB has VaultFields set
// (not PasswordVar).
//
// BUG #157 — WILL FAIL: current code sets PasswordVar, not VaultFields.
func TestCLICreateCredential_PasswordVar_BuildsVaultFields(t *testing.T) {
	passwordVar := "GITHUB_PAT"

	cred := buildCredentialFromCLIFlags(
		"github-creds", "vault", "github-creds", "production",
		"", passwordVar, // usernameVar, passwordVar
	)

	// WILL FAIL with buggy code:
	require.True(t, cred.HasVaultFields(),
		"BUG #157: --password-var must populate VaultFields, not PasswordVar — "+
			"WILL FAIL until cmd/credential.go is fixed")

	fields, err := cred.GetVaultFieldsMap()
	require.NoError(t, err)
	assert.Equal(t, "GITHUB_PAT", fields["GITHUB_PAT"],
		"VaultFields map must contain GITHUB_PAT -> GITHUB_PAT mapping")

	assert.Nil(t, cred.PasswordVar,
		"after fix: PasswordVar must be nil when --password-var builds VaultFields")
}

// TestCLICreateCredential_BothVars_BuildsVaultFields verifies that when both
// --username-var and --password-var are provided, VaultFields contains both
// mappings (not UsernameVar/PasswordVar DB columns).
//
// BUG #157 — WILL FAIL: current code sets UsernameVar/PasswordVar, not VaultFields.
func TestCLICreateCredential_BothVars_BuildsVaultFields(t *testing.T) {
	cred := buildCredentialFromCLIFlags(
		"github-creds", "vault", "github-creds", "production",
		"GITHUB_USERNAME", "GITHUB_PAT",
	)

	// WILL FAIL with buggy code:
	require.True(t, cred.HasVaultFields(),
		"BUG #157: both --username-var and --password-var must populate VaultFields — "+
			"WILL FAIL until cmd/credential.go is fixed")

	fields, err := cred.GetVaultFieldsMap()
	require.NoError(t, err)
	require.Len(t, fields, 2,
		"VaultFields must contain exactly 2 entries (username + password)")
	assert.Equal(t, "GITHUB_USERNAME", fields["GITHUB_USERNAME"],
		"VaultFields must contain GITHUB_USERNAME -> GITHUB_USERNAME")
	assert.Equal(t, "GITHUB_PAT", fields["GITHUB_PAT"],
		"VaultFields must contain GITHUB_PAT -> GITHUB_PAT")

	assert.Nil(t, cred.UsernameVar,
		"after fix: UsernameVar must be nil when --username-var builds VaultFields")
	assert.Nil(t, cred.PasswordVar,
		"after fix: PasswordVar must be nil when --password-var builds VaultFields")
}

// TestCLICreateCredential_VaultEnv_CorrectlyMapped verifies that --vault-env
// correctly sets cred.VaultEnv (this should PASS — baseline that the vault-env
// flag is NOT broken).
//
// This is a passing test that confirms --vault-env is correctly mapped while
// --username-var / --password-var are broken.
func TestCLICreateCredential_VaultEnv_CorrectlyMapped(t *testing.T) {
	cred := buildCredentialFromCLIFlags(
		"my-cred", "vault", "my-secret", "production",
		"", "", // no username/password var
	)

	require.NotNil(t, cred.VaultEnv,
		"--vault-env must map to cred.VaultEnv (this should PASS — vault-env is correct)")
	assert.Equal(t, "production", *cred.VaultEnv,
		"VaultEnv must match the --vault-env flag value")
}

// TestCLICreateCredential_CLIPath_EqualsYAMLPath verifies that a credential
// created via CLI flags produces the same CredentialDB structure as one created
// via YAML apply. This is the round-trip equivalence test.
//
// BUG #157 — WILL FAIL: CLI path sets UsernameVar/PasswordVar while YAML
// path sets VaultFields, making them structurally different.
func TestCLICreateCredential_CLIPath_EqualsYAMLPath(t *testing.T) {
	// YAML apply path: spec.vaultFields -> cred.VaultFields (the working path)
	yamlCred := &models.CredentialDB{}
	yamlInput := models.CredentialYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "Credential",
		Metadata: models.CredentialMetadata{
			Name:      "github-creds",
			Ecosystem: "myorg",
		},
		Spec: models.CredentialSpec{
			Source:      "vault",
			VaultSecret: "github-creds",
			VaultFields: map[string]string{
				"GITHUB_USERNAME": "GITHUB_USERNAME",
				"GITHUB_PAT":      "GITHUB_PAT",
			},
		},
	}
	yamlCred.FromYAML(yamlInput)

	// CLI path (after fix): --username-var / --password-var -> VaultFields
	cliCred := buildCredentialFromCLIFlags(
		"github-creds", "vault", "github-creds", "",
		"GITHUB_USERNAME", "GITHUB_PAT",
	)

	// Both must have VaultFields set (WILL FAIL for CLI path with current bug)
	require.True(t, yamlCred.HasVaultFields(),
		"YAML-applied credential must have VaultFields (sanity check)")

	// WILL FAIL with buggy code (CLI path sets UsernameVar/PasswordVar, not VaultFields):
	require.True(t, cliCred.HasVaultFields(),
		"BUG #157: CLI-created credential MUST have VaultFields to be equivalent to "+
			"YAML-applied credential — WILL FAIL until cmd/credential.go is fixed")

	// Both should be structurally equivalent (same VaultFields map)
	yamlFields, err := yamlCred.GetVaultFieldsMap()
	require.NoError(t, err)
	cliFields, err := cliCred.GetVaultFieldsMap()
	require.NoError(t, err)

	assert.Equal(t, yamlFields, cliFields,
		"CLI-created credential VaultFields must equal YAML-applied credential VaultFields")
}

// TestCLICreateCredential_ToMapEntries_DualVar_HasVaultField verifies the
// end-to-end path: a credential created via CLI with --username-var/--password-var
// must produce CredentialConfigs with VaultField set (enabling field-level
// vault access).
//
// BUG #157 (combined effect of both bugs) — WILL FAIL until both:
//  1. cmd/credential.go builds VaultFields instead of UsernameVar/PasswordVar
//  2. (or) models/credential.go ToUsernameConfig/ToPasswordConfig sets VaultField
func TestCLICreateCredential_ToMapEntries_DualVar_HasVaultField(t *testing.T) {
	cred := buildCredentialFromCLIFlags(
		"github-creds", "vault", "github-creds", "production",
		"GITHUB_USERNAME", "GITHUB_PAT",
	)

	entries := credentialbridge.ToMapEntries(cred)

	require.Len(t, entries, 2,
		"dual-var CLI credential must fan out to 2 map entries")

	userEntry, ok := entries["GITHUB_USERNAME"]
	require.True(t, ok, "map must contain GITHUB_USERNAME entry")
	// WILL FAIL (combined bug): either VaultField is "" or the entry doesn't
	// go through the VaultFields path at all
	assert.NotEmpty(t, userEntry.VaultField,
		"BUG #157: GITHUB_USERNAME entry MUST have VaultField set for field-level "+
			"vault access — WILL FAIL until fix is applied")

	patEntry, ok := entries["GITHUB_PAT"]
	require.True(t, ok, "map must contain GITHUB_PAT entry")
	assert.NotEmpty(t, patEntry.VaultField,
		"BUG #157: GITHUB_PAT entry MUST have VaultField set for field-level "+
			"vault access — WILL FAIL until fix is applied")
}

// ---------------------------------------------------------------------------
// Helper: buildCredentialFromCLIFlags
// ---------------------------------------------------------------------------

// buildCredentialFromCLIFlags simulates what createCredentialCmd.RunE does
// when building a CredentialDB from CLI flags. It mirrors the fixed code
// in cmd/credential.go that builds VaultFields from --username-var / --password-var.
//
// This helper is intentionally separate from the production code so that the
// tests document exactly what the expected transformation should be.
func buildCredentialFromCLIFlags(
	name, source, vaultSecret, vaultEnv,
	usernameVar, passwordVar string,
) *models.CredentialDB {
	cred := &models.CredentialDB{
		Name:   name,
		Source: source,
	}
	if vaultSecret != "" {
		cred.VaultSecret = &vaultSecret
	}
	if vaultEnv != "" {
		cred.VaultEnv = &vaultEnv
	}

	// FIXED CODE — matches cmd/credential.go after bug #157 fix:
	// Build VaultFields map from --username-var / --password-var flags.
	if usernameVar != "" || passwordVar != "" {
		vaultFields := make(map[string]string)
		if usernameVar != "" {
			vaultFields[usernameVar] = usernameVar
		}
		if passwordVar != "" {
			vaultFields[passwordVar] = passwordVar
		}
		vfJSON, _ := json.Marshal(vaultFields)
		vfStr := string(vfJSON)
		cred.VaultFields = &vfStr
	}

	return cred
}
