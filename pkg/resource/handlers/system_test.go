package handlers

import (
	"database/sql"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// =============================================================================
// Helpers
// =============================================================================

// setupSystemTest creates a mock store with one ecosystem and one domain.
// Returns the store, ecosystemID, and domainID.
func setupSystemTest(t *testing.T) (*db.MockDataStore, int, int) {
	t.Helper()
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "test-eco"}
	if err := store.CreateEcosystem(eco); err != nil {
		t.Fatalf("failed to create ecosystem: %v", err)
	}

	dom := &models.Domain{
		Name:        "test-domain",
		EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true},
	}
	if err := store.CreateDomain(dom); err != nil {
		t.Fatalf("failed to create domain: %v", err)
	}

	return store, eco.ID, dom.ID
}

// =============================================================================
// SystemHandler Tests - Kind
// =============================================================================

func TestSystemHandler_Kind(t *testing.T) {
	h := NewSystemHandler()
	if h.Kind() != KindSystem {
		t.Errorf("Kind() = %q, want %q", h.Kind(), KindSystem)
	}
}

// =============================================================================
// SystemHandler Tests - Apply: ecosystem inference from domain (Issue #392)
// =============================================================================

// TestSystemHandler_Apply_InfersEcosystemFromDomain verifies that when a System
// YAML has metadata.domain but no metadata.ecosystem, Apply succeeds and infers
// the ecosystem from the domain lookup.
func TestSystemHandler_Apply_InfersEcosystemFromDomain(t *testing.T) {
	h := NewSystemHandler()
	store, _, _ := setupSystemTest(t)
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: my-system
  domain: test-domain
spec: {}
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v; want nil (ecosystem should be inferred from domain)", err)
	}
	if res.GetKind() != KindSystem {
		t.Errorf("Apply() Kind = %q, want %q", res.GetKind(), KindSystem)
	}
	if res.GetName() != "my-system" {
		t.Errorf("Apply() Name = %q, want %q", res.GetName(), "my-system")
	}

	sr, ok := res.(*SystemResource)
	if !ok {
		t.Fatalf("result is not *SystemResource")
	}
	if sr.EcosystemName() != "test-eco" {
		t.Errorf("EcosystemName() = %q, want %q (should be inferred)", sr.EcosystemName(), "test-eco")
	}
	if sr.DomainName() != "test-domain" {
		t.Errorf("DomainName() = %q, want %q", sr.DomainName(), "test-domain")
	}
}

// TestSystemHandler_Apply_WithExplicitEcosystem verifies that when both domain
// and ecosystem are specified, Apply succeeds normally.
func TestSystemHandler_Apply_WithExplicitEcosystem(t *testing.T) {
	h := NewSystemHandler()
	store, _, _ := setupSystemTest(t)
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: my-system
  domain: test-domain
  ecosystem: test-eco
spec: {}
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	sr := res.(*SystemResource)
	if sr.EcosystemName() != "test-eco" {
		t.Errorf("EcosystemName() = %q, want %q", sr.EcosystemName(), "test-eco")
	}
}

// TestSystemHandler_Apply_NoDomainNoEcosystem verifies that a System with
// neither domain nor ecosystem is applied successfully (both are optional).
func TestSystemHandler_Apply_NoDomainNoEcosystem(t *testing.T) {
	h := NewSystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: standalone-system
spec: {}
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v; want nil (domain and ecosystem are optional)", err)
	}
	if res.GetName() != "standalone-system" {
		t.Errorf("Apply() Name = %q, want %q", res.GetName(), "standalone-system")
	}
}

// TestSystemHandler_Apply_DomainNotFound verifies that specifying a non-existent
// domain (without ecosystem) returns an appropriate error.
func TestSystemHandler_Apply_DomainNotFound(t *testing.T) {
	h := NewSystemHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: my-system
  domain: nonexistent-domain
spec: {}
`)

	_, err := h.Apply(ctx, yamlData)
	if err == nil {
		t.Fatal("Apply() expected error for nonexistent domain, got nil")
	}
}

// TestSystemHandler_Apply_AmbiguousDomain verifies that when a domain name
// exists in multiple ecosystems, Apply errors asking the user to disambiguate.
func TestSystemHandler_Apply_AmbiguousDomain(t *testing.T) {
	h := NewSystemHandler()
	store := db.NewMockDataStore()

	// Create two ecosystems with the same domain name
	eco1 := &models.Ecosystem{Name: "eco-alpha"}
	eco2 := &models.Ecosystem{Name: "eco-beta"}
	if err := store.CreateEcosystem(eco1); err != nil {
		t.Fatalf("failed to create eco1: %v", err)
	}
	if err := store.CreateEcosystem(eco2); err != nil {
		t.Fatalf("failed to create eco2: %v", err)
	}

	dom1 := &models.Domain{Name: "shared-domain", EcosystemID: sql.NullInt64{Int64: int64(eco1.ID), Valid: true}}
	dom2 := &models.Domain{Name: "shared-domain", EcosystemID: sql.NullInt64{Int64: int64(eco2.ID), Valid: true}}
	if err := store.CreateDomain(dom1); err != nil {
		t.Fatalf("failed to create dom1: %v", err)
	}
	if err := store.CreateDomain(dom2); err != nil {
		t.Fatalf("failed to create dom2: %v", err)
	}

	ctx := resource.Context{DataStore: store}
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: my-system
  domain: shared-domain
spec: {}
`)

	_, err := h.Apply(ctx, yamlData)
	if err == nil {
		t.Fatal("Apply() expected error for ambiguous domain, got nil")
	}
}

// =============================================================================
// SystemHandler Tests - Apply: idempotency
// =============================================================================

func TestSystemHandler_Apply_Idempotent(t *testing.T) {
	h := NewSystemHandler()
	store, _, _ := setupSystemTest(t)
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: idempotent-system
  domain: test-domain
spec: {}
`)

	res1, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("first Apply() error = %v", err)
	}
	res2, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("second Apply() error = %v", err)
	}
	if res1.GetName() != res2.GetName() {
		t.Errorf("idempotency: name mismatch %q vs %q", res1.GetName(), res2.GetName())
	}
}
