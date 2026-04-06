package handlers

// =============================================================================
// GitRepo-Credential Round-Trip Tests — Issue #179
//
// These tests prove the bug: GitRepo credential associations are lost on YAML
// export → wipe → restore because:
//
//  1. GitRepoSpec has no Credential field
//  2. GitRepoDB.ToYAML() never resolves CredentialID → name
//  3. GitRepoHandler.Apply() never resolves a credential name → ID
//
// All tests MUST compile but MUST fail (red phase of TDD) until the fix is in.
// Compare to the App-GitRepo fix (#175) which correctly implements this pattern.
// =============================================================================

import (
	"database/sql"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

// =============================================================================
// TestGitRepoCredentialYAMLExport_IncludesCredentialName
//
// Arrange: Create a GitRepoDB with CredentialID set. Wrap it in a
//
//	GitRepoResource (the way the handler returns it today).
//
// Act:     Call GitRepoHandler.ToYAML().
//
// Assert:  The resulting YAML contains a "credential:" field with the resolved
//
//	credential name (not the numeric ID).
//
// WHY IT FAILS: GitRepoResource has no credentialName field.
//
//	GitRepoDB.ToYAML() never looks up CredentialID → name.
//	GitRepoSpec has no Credential field.
//	The YAML output will never contain a credential reference.
//
// =============================================================================
func TestGitRepoCredentialYAMLExport_IncludesCredentialName(t *testing.T) {
	h := NewGitRepoHandler()

	// Simulate a git repo that was created with --auth-type https --credential my-cred
	repo := &models.GitRepoDB{
		ID:           1,
		Name:         "secure-repo",
		URL:          "https://github.com/acme/secure-repo",
		AuthType:     "https",
		CredentialID: sql.NullInt64{Int64: 42, Valid: true},
	}

	// A fixed implementation carries the resolved credential name in
	// GitRepoResource (mirroring AppResource.gitRepoName pattern).
	// The credential name for CredentialID=42 would be resolved by Get()/List().
	// For this unit test we pass it directly to NewGitRepoResource.
	res := NewGitRepoResource(repo, "my-cred")

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() unexpected error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// MUST FAIL: no credential field will appear in the output because:
	//   1. GitRepoSpec has no Credential field
	//   2. GitRepoDB.ToYAML() never resolves CredentialID → name
	//   3. The current implementation silently drops the credential association
	if !strings.Contains(yamlStr, "credential:") {
		t.Errorf("ToYAML() output is missing 'credential' field — the credential association will be lost on export/restore.\n"+
			"GitRepoDB.CredentialID was set to 42 but was not serialised to YAML.\n"+
			"Got YAML:\n%s", yamlStr)
	}
}

// =============================================================================
// TestGitRepoCredentialApply_ResolvesNameToID
//
// Arrange: Pre-populate the mock store with a credential named "my-cred".
//
//	Provide GitRepo YAML that references spec.credential: my-cred.
//
// Act:     Call GitRepoHandler.Apply().
//
// Assert:  The resulting GitRepoResource's GitRepoDB has CredentialID set to
//
//	the ID of the "my-cred" record.
//
// WHY IT FAILS: GitRepoSpec has no Credential field, so the name is never
//
//	parsed. Even if it were parsed, Apply() has no code to resolve it
//	to an ID. The resulting GitRepoDB.CredentialID will be the zero value
//	(not valid).
//
// =============================================================================
func TestGitRepoCredentialApply_ResolvesNameToID(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// Pre-populate a credential
	cred := &models.CredentialDB{
		Name:      "my-cred",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   1,
		Source:    "env",
		EnvVar:    strPtr("MY_CRED_TOKEN"),
	}
	if err := store.CreateCredential(cred); err != nil {
		t.Fatalf("CreateCredential() unexpected error = %v", err)
	}

	// YAML with spec.credential — this is the contract the fix must honour
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: cred-repo
spec:
  url: https://github.com/acme/cred-repo
  authType: https
  credential: my-cred
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() unexpected error = %v", err)
	}

	grr, ok := res.(*GitRepoResource)
	if !ok {
		t.Fatalf("Apply() result is not *GitRepoResource, got %T", res)
	}

	// MUST FAIL: Apply() never reads spec.credential, so CredentialID is never set.
	if !grr.GitRepo().CredentialID.Valid {
		t.Errorf("Apply(): GitRepoDB.CredentialID is not set — spec.credential was not resolved to a CredentialID FK.\n"+
			"Expected CredentialID = %d (ID of 'my-cred'), got zero value.\n"+
			"GitRepoDB struct: %+v", cred.ID, grr.GitRepo())
	} else if grr.GitRepo().CredentialID.Int64 != cred.ID {
		t.Errorf("Apply(): GitRepoDB.CredentialID = %d, want %d (ID of 'my-cred')",
			grr.GitRepo().CredentialID.Int64, cred.ID)
	}

	// Also verify the association persisted in the store
	storedRepo, err := store.GetGitRepoByName("cred-repo")
	if err != nil {
		t.Fatalf("GetGitRepoByName() after Apply() error = %v", err)
	}
	if !storedRepo.CredentialID.Valid {
		t.Errorf("stored GitRepoDB.CredentialID is not set — credential association was not persisted to the DB")
	} else if storedRepo.CredentialID.Int64 != cred.ID {
		t.Errorf("stored GitRepoDB.CredentialID = %d, want %d", storedRepo.CredentialID.Int64, cred.ID)
	}
}

// =============================================================================
// TestGitRepoCredentialRoundTrip_ExportRestorePreservesCredential
//
// Table-driven end-to-end round-trip test.
// Proves the complete export → restore path is broken for git repos with
// credential associations.
//
// WHY IT FAILS: ToYAML() never includes the credential reference, so the
//
//	round-trip (export → apply) would silently drop the association.
//
// =============================================================================
func TestGitRepoCredentialRoundTrip_ExportRestorePreservesCredential(t *testing.T) {
	tests := []struct {
		name     string
		credName string
		repoName string
		repoURL  string
		authType string
	}{
		{
			name:     "https with env credential",
			credName: "github-token",
			repoName: "acme-backend",
			repoURL:  "https://github.com/acme/backend",
			authType: "https",
		},
		{
			name:     "https with vault credential",
			credName: "vault-git-cred",
			repoName: "internal-api",
			repoURL:  "https://gitlab.internal/org/api",
			authType: "https",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewGitRepoHandler()
			store := db.NewMockDataStore()

			// Create the credential
			cred := &models.CredentialDB{
				Name:      tt.credName,
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   1,
				Source:    "env",
				EnvVar:    strPtr("SOME_TOKEN"),
			}
			if err := store.CreateCredential(cred); err != nil {
				t.Fatalf("CreateCredential() error = %v", err)
			}

			// Simulate a GitRepo that was created with --auth-type https --credential <name>
			// and therefore has CredentialID set in the DB.
			repo := &models.GitRepoDB{
				ID:           1,
				Name:         tt.repoName,
				URL:          tt.repoURL,
				AuthType:     tt.authType,
				CredentialID: sql.NullInt64{Int64: cred.ID, Valid: true},
			}

			// Step 1: Export to YAML
			// A fixed implementation: Get() resolves CredentialID → name and
			// stores it in GitRepoResource (mirroring the App-GitRepo fix).
			repoRes := NewGitRepoResource(repo, tt.credName)

			yamlBytes, err := h.ToYAML(repoRes)
			if err != nil {
				t.Fatalf("ToYAML() error = %v", err)
			}
			yamlStr := string(yamlBytes)

			// MUST FAIL: no credential field in output
			if !strings.Contains(yamlStr, "credential:") {
				t.Errorf("[%s] ToYAML() is missing 'credential' field — association will be silently lost on restore.\n"+
					"GitRepoDB had CredentialID=%d (%s) but it was not serialised.\nYAML:\n%s",
					tt.name, cred.ID, tt.credName, yamlStr)
			}
			if !strings.Contains(yamlStr, tt.credName) {
				t.Errorf("[%s] ToYAML() YAML does not contain credential name %q.\nYAML:\n%s",
					tt.name, tt.credName, yamlStr)
			}

			// Step 2: Simulate restore by applying the exported YAML to a fresh store
			freshStore := db.NewMockDataStore()
			// Credential must exist in the restore target for the FK to resolve
			restoreCred := &models.CredentialDB{
				Name:      tt.credName,
				ScopeType: models.CredentialScopeEcosystem,
				ScopeID:   1,
				Source:    "env",
				EnvVar:    strPtr("SOME_TOKEN"),
			}
			if err := freshStore.CreateCredential(restoreCred); err != nil {
				t.Fatalf("CreateCredential (restore) error = %v", err)
			}
			restoreCtx := resource.Context{DataStore: freshStore}

			res, err := h.Apply(restoreCtx, yamlBytes)
			if err != nil {
				t.Fatalf("[%s] Apply() (restore) error = %v", tt.name, err)
			}

			grr, ok := res.(*GitRepoResource)
			if !ok {
				t.Fatalf("[%s] Apply() result is not *GitRepoResource, got %T", tt.name, res)
			}

			// MUST FAIL: CredentialID will not be set after restore
			if !grr.GitRepo().CredentialID.Valid {
				t.Errorf("[%s] After round-trip restore: GitRepoDB.CredentialID is not set.\n"+
					"Original CredentialID was %d (%s) but was not preserved through export/restore.\n"+
					"GitRepoDB struct: %+v",
					tt.name, cred.ID, tt.credName, grr.GitRepo())
			}
		})
	}
}

// =============================================================================
// TestGitRepoCredentialFromYAML_ReadsCredentialField
//
// Arrange: Construct raw YAML bytes that include spec.credential.
//
// Act:     Unmarshal into models.GitRepoYAML using gopkg.in/yaml.v3.
//
// Assert:  The parsed GitRepoSpec retains the credential name in a Credential
//
//	field.
//
// WHY IT FAILS: GitRepoSpec has no Credential field. The yaml decoder will
//
//	silently discard the spec.credential key. The test proves the field
//	is absent.
//
// =============================================================================
func TestGitRepoCredentialFromYAML_ReadsCredentialField(t *testing.T) {
	// This is the YAML that a fixed exporter would produce.
	// The spec.credential key must survive a round-trip through GitRepoYAML.
	rawYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: cred-roundtrip-repo
spec:
  url: https://github.com/acme/cred-roundtrip-repo
  authType: https
  credential: my-https-cred
`)

	var parsed models.GitRepoYAML
	if err := yaml.Unmarshal(rawYAML, &parsed); err != nil {
		t.Fatalf("yaml.Unmarshal() unexpected error = %v", err)
	}

	// Sanity check: confirm basic fields parsed correctly
	if parsed.Spec.URL != "https://github.com/acme/cred-roundtrip-repo" {
		t.Fatalf("sanity check failed: Spec.URL = %q, want %q — YAML was not parsed at all",
			parsed.Spec.URL, "https://github.com/acme/cred-roundtrip-repo")
	}

	// MUST FAIL: GitRepoSpec currently has no Credential field at all.
	// We cannot reference parsed.Spec.Credential directly — it would be a compile error.
	// Instead we re-marshal the parsed spec to check whether the credential key
	// survived the round-trip. This mirrors the technique used in the App-GitRepo test.
	reMarshalled, err := yaml.Marshal(parsed.Spec)
	if err != nil {
		t.Fatalf("yaml.Marshal(parsed.Spec) unexpected error = %v", err)
	}
	reStr := string(reMarshalled)

	// MUST FAIL: the re-marshalled spec will not contain "credential" because
	// GitRepoSpec has no such field — the value was discarded on unmarshal.
	if !strings.Contains(reStr, "credential") {
		t.Errorf("FromYAML: spec.credential was silently discarded — GitRepoSpec has no Credential field.\n"+
			"Input YAML contained 'credential: my-https-cred' but it was lost.\n"+
			"Re-marshalled GitRepoSpec:\n%s", reStr)
	}
}

// =============================================================================
// TestGitRepoCredentialList_IncludesCredentialName
//
// Arrange: Pre-populate the store with a git repo that has CredentialID set.
// Act:     Call GitRepoHandler.List().
// Assert:  The returned resources, when serialised via ToYAML(), include the
//
//	credential name.
//
// WHY IT FAILS: List() returns GitRepoResource instances with no credential
//
//	name resolved. ToYAML() then produces output without a credential field.
//	The credential association is invisible via the List path.
//
// =============================================================================
func TestGitRepoCredentialList_IncludesCredentialName(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// Pre-populate a credential
	cred := &models.CredentialDB{
		Name:      "list-test-cred",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   1,
		Source:    "env",
		EnvVar:    strPtr("LIST_TOKEN"),
	}
	if err := store.CreateCredential(cred); err != nil {
		t.Fatalf("CreateCredential() error = %v", err)
	}

	// Pre-populate a git repo that references the credential
	repo := &models.GitRepoDB{
		Name:         "list-repo",
		URL:          "https://github.com/acme/list-repo",
		AuthType:     "https",
		CredentialID: sql.NullInt64{Int64: cred.ID, Valid: true},
	}
	if err := store.CreateGitRepo(repo); err != nil {
		t.Fatalf("CreateGitRepo() error = %v", err)
	}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("List() returned %d items, want 1", len(resources))
	}

	yamlBytes, err := h.ToYAML(resources[0])
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// MUST FAIL: List() does not resolve CredentialID → name, and
	// GitRepoHandler.ToYAML() / GitRepoDB.ToYAML() have no Credential field.
	// The YAML output will not contain the credential reference.
	if !strings.Contains(yamlStr, "credential:") {
		t.Errorf("List()+ToYAML() output is missing 'credential' field.\n"+
			"GitRepoDB had CredentialID=%d ('list-test-cred') but it was not included in list output.\n"+
			"Got YAML:\n%s", cred.ID, yamlStr)
	}
	if !strings.Contains(yamlStr, "list-test-cred") {
		t.Errorf("List()+ToYAML() YAML does not contain credential name 'list-test-cred'.\n"+
			"Got YAML:\n%s", yamlStr)
	}
}

// =============================================================================
// Helpers
// =============================================================================

// strPtr returns a pointer to the given string value.
func strPtr(s string) *string {
	return &s
}
