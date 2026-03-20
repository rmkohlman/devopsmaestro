package handlers

// =============================================================================
// TDD Phase 2 (RED): Bug #128 — Credential export missing scopeName
//
// get_all.go line 190: handlers.NewCredentialResource(c, "") — empty scopeName
// credential.go List() line 260: &CredentialResource{credential: cred} — no scopeName
//
// These tests verify the round-trip requirement: credentials exported via
// get_all or List() must have a non-empty scopeName so that:
//   1. ToYAML() produces valid YAML with the correct scope field
//   2. Validate() passes (ValidateCredentialYAML requires exactly one scope)
//   3. The exported YAML can be re-applied without error
//
// ALL tests in this section WILL FAIL until:
//   - get_all.go is fixed to pass a resolved scopeName to NewCredentialResource
//   - credential.go List() is fixed to resolve and attach scopeName
// =============================================================================

import (
	"strings"
	"testing"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// TestNewCredentialResource_EmptyScopeName_FailsValidation verifies that the
// current bug is demonstrable: NewCredentialResource called with "" scopeName
// produces a resource whose Validate() fails (as it does in get_all.go today).
//
// RED: This test documents the bug. It passes right now because the bug
// exists — Validate() DOES fail with empty scopeName. Once the bug is fixed
// (scopeName is always populated), this test should still pass (it tests that
// empty scopeName correctly fails validation — that invariant should be kept).
func TestNewCredentialResource_EmptyScopeName_FailsValidation(t *testing.T) {
	vaultSecret := "github/token"
	cred := &models.CredentialDB{
		Name:        "my-cred",
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     1,
		Source:      "vault",
		VaultSecret: &vaultSecret,
	}

	// Simulate the bug: get_all.go calls NewCredentialResource(c, "")
	res := NewCredentialResource(cred, "")
	err := res.Validate()
	if err == nil {
		t.Error("Validate() should fail when scopeName is empty — this is the documented bug in get_all.go")
	}
}

// TestCredentialHandler_List_ReturnsResourcesWithScopeNames verifies that
// CredentialHandler.List() returns CredentialResources with non-empty scopeNames.
//
// RED: Fails because credential.go List() (line 260) creates
// &CredentialResource{credential: cred} with no scopeName.
func TestCredentialHandler_List_ReturnsResourcesWithScopeNames(t *testing.T) {
	h := NewCredentialHandler()
	ds := createCredentialTestDataStore(t)
	defer ds.Close()

	// Seed hierarchy
	eco := &models.Ecosystem{Name: "list-scope-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("CreateEcosystem: %v", err)
	}

	// Create a credential scoped to this ecosystem
	vaultSecret := "list/secret"
	cred := &models.CredentialDB{
		Name:        "list-cred",
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Source:      "vault",
		VaultSecret: &vaultSecret,
	}
	if err := ds.CreateCredential(cred); err != nil {
		t.Fatalf("CreateCredential: %v", err)
	}

	ctx := resource.Context{DataStore: ds}
	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(resources) == 0 {
		t.Fatal("List() returned 0 resources, want at least 1")
	}

	for i, r := range resources {
		credRes, ok := r.(*CredentialResource)
		if !ok {
			t.Fatalf("resource[%d] is not *CredentialResource", i)
		}
		// Bug: scopeName is currently "" because List() doesn't resolve it
		if credRes.ScopeName() == "" {
			t.Errorf("resource[%d] (%q) has empty scopeName — bug in credential.go List()", i, credRes.GetName())
		}
	}
}

// TestCredentialHandler_List_YAML_PassesValidation verifies that credentials
// returned by List() can be serialized to YAML and that YAML passes validation.
//
// RED: Fails because List() doesn't set scopeName, so ToYAML produces YAML
// with no scope field, which fails ValidateCredentialYAML.
func TestCredentialHandler_List_YAML_PassesValidation(t *testing.T) {
	h := NewCredentialHandler()
	ds := createCredentialTestDataStore(t)
	defer ds.Close()

	eco := &models.Ecosystem{Name: "list-yaml-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("CreateEcosystem: %v", err)
	}

	vaultSecret := "yaml/secret"
	cred := &models.CredentialDB{
		Name:        "yaml-cred",
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     int64(eco.ID),
		Source:      "vault",
		VaultSecret: &vaultSecret,
	}
	if err := ds.CreateCredential(cred); err != nil {
		t.Fatalf("CreateCredential: %v", err)
	}

	ctx := resource.Context{DataStore: ds}
	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(resources) == 0 {
		t.Fatal("List() returned 0 resources")
	}

	for i, r := range resources {
		// Validate() calls ValidateCredentialYAML internally — must pass
		if err := r.(*CredentialResource).Validate(); err != nil {
			t.Errorf("resource[%d] (%q) Validate() failed: %v — scope name was not resolved in List()", i, r.GetName(), err)
		}
	}
}

// TestNewCredentialResource_WithScopeName_ProducesCorrectYAML verifies that
// when NewCredentialResource IS called with a non-empty scopeName (the fixed
// behavior), ToYAML produces YAML with the correct scope field for each scope type.
//
// GREEN: This test already passes (NewCredentialResource itself is fine — the
// bug is the caller passing ""). Included here to pin the correct behavior.
func TestNewCredentialResource_WithScopeName_ProducesCorrectYAML(t *testing.T) {
	h := NewCredentialHandler()

	tests := []struct {
		name          string
		scopeType     models.CredentialScopeType
		scopeName     string
		wantYAMLField string
	}{
		{
			name:          "ecosystem scope",
			scopeType:     models.CredentialScopeEcosystem,
			scopeName:     "my-ecosystem",
			wantYAMLField: "ecosystem: my-ecosystem",
		},
		{
			name:          "domain scope",
			scopeType:     models.CredentialScopeDomain,
			scopeName:     "my-domain",
			wantYAMLField: "domain: my-domain",
		},
		{
			name:          "app scope",
			scopeType:     models.CredentialScopeApp,
			scopeName:     "my-app",
			wantYAMLField: "app: my-app",
		},
		{
			name:          "workspace scope",
			scopeType:     models.CredentialScopeWorkspace,
			scopeName:     "my-workspace",
			wantYAMLField: "workspace: my-workspace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vaultSecret := "some/secret"
			cred := &models.CredentialDB{
				Name:        "scoped-cred",
				ScopeType:   tt.scopeType,
				ScopeID:     1,
				Source:      "vault",
				VaultSecret: &vaultSecret,
			}
			res := NewCredentialResource(cred, tt.scopeName)

			yamlBytes, err := h.ToYAML(res)
			if err != nil {
				t.Fatalf("ToYAML() error = %v", err)
			}
			yamlStr := string(yamlBytes)

			if !strings.Contains(yamlStr, tt.wantYAMLField) {
				t.Errorf("ToYAML() missing %q in output:\n%s", tt.wantYAMLField, yamlStr)
			}

			// Also validate the round-trip
			if err := res.Validate(); err != nil {
				t.Errorf("Validate() should pass with scopeName %q, got: %v", tt.scopeName, err)
			}
		})
	}
}
