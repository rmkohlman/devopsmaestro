package resolver

import (
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
)

// setupTestData creates a mock data store with test hierarchy data
func setupTestData() *db.MockDataStore {
	store := db.NewMockDataStore()

	// Create ecosystems
	eco1 := &models.Ecosystem{Name: "healthcare"}
	eco2 := &models.Ecosystem{Name: "fintech"}
	store.CreateEcosystem(eco1)
	store.CreateEcosystem(eco2)

	// Create domains
	dom1 := &models.Domain{EcosystemID: eco1.ID, Name: "billing"}
	dom2 := &models.Domain{EcosystemID: eco1.ID, Name: "claims"}
	dom3 := &models.Domain{EcosystemID: eco2.ID, Name: "payments"}
	store.CreateDomain(dom1)
	store.CreateDomain(dom2)
	store.CreateDomain(dom3)

	// Create apps
	app1 := &models.App{DomainID: dom1.ID, Name: "portal"}
	app2 := &models.App{DomainID: dom2.ID, Name: "portal"} // Same name, different domain
	app3 := &models.App{DomainID: dom3.ID, Name: "api"}
	store.CreateApp(app1)
	store.CreateApp(app2)
	store.CreateApp(app3)

	// Create workspaces
	ws1 := &models.Workspace{AppID: app1.ID, Name: "dev", ImageName: "portal-dev:latest", Status: "created"}
	ws2 := &models.Workspace{AppID: app1.ID, Name: "staging", ImageName: "portal-staging:latest", Status: "created"}
	ws3 := &models.Workspace{AppID: app2.ID, Name: "dev", ImageName: "portal-claims-dev:latest", Status: "created"}
	ws4 := &models.Workspace{AppID: app3.ID, Name: "dev", ImageName: "api-dev:latest", Status: "created"}
	store.CreateWorkspace(ws1)
	store.CreateWorkspace(ws2)
	store.CreateWorkspace(ws3)
	store.CreateWorkspace(ws4)

	return store
}

func TestResolve_SingleMatch(t *testing.T) {
	store := setupTestData()
	resolver := NewWorkspaceResolver(store)

	// Filter by unique app name (api only exists in fintech/payments)
	result, err := resolver.Resolve(models.WorkspaceFilter{
		AppName: "api",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected result, got nil")
	}

	if result.Workspace.Name != "dev" {
		t.Errorf("expected workspace 'dev', got '%s'", result.Workspace.Name)
	}

	if result.App.Name != "api" {
		t.Errorf("expected app 'api', got '%s'", result.App.Name)
	}

	if result.Ecosystem.Name != "fintech" {
		t.Errorf("expected ecosystem 'fintech', got '%s'", result.Ecosystem.Name)
	}
}

func TestResolve_SingleMatchWithMultipleFilters(t *testing.T) {
	store := setupTestData()
	resolver := NewWorkspaceResolver(store)

	// Filter by ecosystem, domain, and app to get unique workspace
	result, err := resolver.Resolve(models.WorkspaceFilter{
		EcosystemName: "healthcare",
		DomainName:    "billing",
		AppName:       "portal",
		WorkspaceName: "staging",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Workspace.Name != "staging" {
		t.Errorf("expected workspace 'staging', got '%s'", result.Workspace.Name)
	}

	if result.FullPath() != "healthcare/billing/portal/staging" {
		t.Errorf("expected full path 'healthcare/billing/portal/staging', got '%s'", result.FullPath())
	}
}

func TestResolve_NoMatch(t *testing.T) {
	store := setupTestData()
	resolver := NewWorkspaceResolver(store)

	// Filter by non-existent app
	_, err := resolver.Resolve(models.WorkspaceFilter{
		AppName: "nonexistent",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !IsNoWorkspaceFoundError(err) {
		t.Errorf("expected ErrNoWorkspaceFound, got: %v", err)
	}
}

func TestResolve_AmbiguousMatch(t *testing.T) {
	store := setupTestData()
	resolver := NewWorkspaceResolver(store)

	// Filter by app name that exists in multiple domains
	_, err := resolver.Resolve(models.WorkspaceFilter{
		AppName: "portal",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}

	ambiguousErr, ok := IsAmbiguousError(err)
	if !ok {
		t.Fatalf("expected AmbiguousError, got: %T", err)
	}

	// There should be 3 matches (2 apps named "portal" with workspaces)
	// app1 has 2 workspaces (dev, staging), app2 has 1 workspace (dev)
	if len(ambiguousErr.Matches) != 3 {
		t.Errorf("expected 3 matches, got %d", len(ambiguousErr.Matches))
	}
}

func TestResolve_AmbiguousMatchNarrowedByDomain(t *testing.T) {
	store := setupTestData()
	resolver := NewWorkspaceResolver(store)

	// Filter by app name AND domain to narrow down
	_, err := resolver.Resolve(models.WorkspaceFilter{
		DomainName: "billing",
		AppName:    "portal",
	})

	if err == nil {
		t.Fatal("expected error (still 2 workspaces in billing/portal), got nil")
	}

	ambiguousErr, ok := IsAmbiguousError(err)
	if !ok {
		t.Fatalf("expected AmbiguousError, got: %T", err)
	}

	// Should be 2 matches (dev and staging in billing/portal)
	if len(ambiguousErr.Matches) != 2 {
		t.Errorf("expected 2 matches, got %d", len(ambiguousErr.Matches))
	}
}

func TestResolve_FullyQualifiedPath(t *testing.T) {
	store := setupTestData()
	resolver := NewWorkspaceResolver(store)

	// Filter by all criteria - should get exactly one match
	result, err := resolver.Resolve(models.WorkspaceFilter{
		EcosystemName: "healthcare",
		DomainName:    "claims",
		AppName:       "portal",
		WorkspaceName: "dev",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.FullPath() != "healthcare/claims/portal/dev" {
		t.Errorf("expected 'healthcare/claims/portal/dev', got '%s'", result.FullPath())
	}
}

func TestResolveAll_ReturnsMultipleMatches(t *testing.T) {
	store := setupTestData()
	resolver := NewWorkspaceResolver(store)

	// Filter by workspace name "dev" - exists in multiple apps
	results, err := resolver.ResolveAll(models.WorkspaceFilter{
		WorkspaceName: "dev",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find all workspaces named "dev"
	if len(results) != 3 {
		t.Errorf("expected 3 results, got %d", len(results))
	}
}

func TestResolveAll_EmptyFilter(t *testing.T) {
	store := setupTestData()
	resolver := NewWorkspaceResolver(store)

	// Empty filter should return all workspaces
	results, err := resolver.ResolveAll(models.WorkspaceFilter{})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should find all 4 workspaces
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}
}

func TestWorkspaceWithHierarchy_ShortPath(t *testing.T) {
	wh := &models.WorkspaceWithHierarchy{
		Workspace: &models.Workspace{Name: "dev"},
		App:       &models.App{Name: "portal"},
		Domain:    &models.Domain{Name: "billing"},
		Ecosystem: &models.Ecosystem{Name: "healthcare"},
	}

	if wh.ShortPath() != "portal/dev" {
		t.Errorf("expected 'portal/dev', got '%s'", wh.ShortPath())
	}
}

func TestAmbiguousError_FormatDisambiguation(t *testing.T) {
	matches := []*models.WorkspaceWithHierarchy{
		{
			Workspace: &models.Workspace{Name: "dev"},
			App:       &models.App{Name: "portal"},
			Domain:    &models.Domain{Name: "billing"},
			Ecosystem: &models.Ecosystem{Name: "healthcare"},
		},
		{
			Workspace: &models.Workspace{Name: "staging"},
			App:       &models.App{Name: "portal"},
			Domain:    &models.Domain{Name: "billing"},
			Ecosystem: &models.Ecosystem{Name: "healthcare"},
		},
	}

	err := NewAmbiguousError(matches)
	output := err.FormatDisambiguation()

	// Check that key information is present
	if len(output) == 0 {
		t.Error("expected non-empty disambiguation output")
	}

	// Should contain the full paths
	if !contains(output, "healthcare/billing/portal/dev") {
		t.Error("expected output to contain 'healthcare/billing/portal/dev'")
	}

	if !contains(output, "healthcare/billing/portal/staging") {
		t.Error("expected output to contain 'healthcare/billing/portal/staging'")
	}
}

func TestDefaultFactory_Create(t *testing.T) {
	store := db.NewMockDataStore()
	resolver := DefaultFactory.Create(store)

	if resolver == nil {
		t.Error("expected resolver, got nil")
	}
}

// Helper function
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
