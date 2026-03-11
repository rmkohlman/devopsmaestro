package db

import (
	"devopsmaestro/models"
	"strings"
	"testing"
)

// =============================================================================
// Workspace Slug Storage and SSH Agent Tests
// =============================================================================
//
// NOTE: Slug generation logic is tested in pkg/workspace/slug_test.go.
// Path construction logic is tested in pkg/workspace/directory_test.go.
// These tests verify the DB layer correctly stores and retrieves slugs and
// workspace fields.

// TestWorkspaceSlugRoundTrip verifies that slugs are stored and retrieved correctly.
// Slug generation itself is now done by pkg/workspace.GenerateSlug(), so callers
// must set the Slug field before calling CreateWorkspace.
func TestWorkspaceSlugRoundTrip(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	tests := []struct {
		name      string
		ecosystem string
		domain    string
		app       string
		workspace string
		slug      string
	}{
		{
			name:      "simple slug",
			ecosystem: "personal",
			domain:    "tools",
			app:       "dvm",
			workspace: "dev",
			slug:      "personal-tools-dvm-dev",
		},
		{
			name:      "production workspace",
			ecosystem: "enterprise",
			domain:    "payments",
			app:       "api",
			workspace: "prod",
			slug:      "enterprise-payments-api-prod",
		},
		{
			name:      "multi-word names with hyphens",
			ecosystem: "my-org",
			domain:    "data-platform",
			app:       "etl-service",
			workspace: "local-dev",
			slug:      "my-org-data-platform-etl-service-local-dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create hierarchy
			ecosystem := &models.Ecosystem{Name: tt.ecosystem}
			if err := ds.CreateEcosystem(ecosystem); err != nil {
				t.Fatalf("Setup CreateEcosystem() error = %v", err)
			}

			domain := &models.Domain{
				EcosystemID: ecosystem.ID,
				Name:        tt.domain,
			}
			if err := ds.CreateDomain(domain); err != nil {
				t.Fatalf("Setup CreateDomain() error = %v", err)
			}

			app := &models.App{
				DomainID: domain.ID,
				Name:     tt.app,
				Path:     "/path/to/app",
			}
			if err := ds.CreateApp(app); err != nil {
				t.Fatalf("Setup CreateApp() error = %v", err)
			}

			workspace := &models.Workspace{
				AppID:     app.ID,
				Name:      tt.workspace,
				Slug:      tt.slug, // Caller must set slug before CreateWorkspace
				ImageName: "test:latest",
				Status:    "stopped",
			}

			if err := ds.CreateWorkspace(workspace); err != nil {
				t.Fatalf("CreateWorkspace() error = %v", err)
			}

			// Get workspace and verify slug round-trips correctly
			retrieved, err := ds.GetWorkspaceByID(workspace.ID)
			if err != nil {
				t.Fatalf("GetWorkspaceByID() error = %v", err)
			}

			if retrieved.Slug != tt.slug {
				t.Errorf("GetWorkspaceByID() Slug = %q, want %q", retrieved.Slug, tt.slug)
			}
		})
	}
}

// TestWorkspaceSlugUniqueness verifies that duplicate slugs are prevented
func TestWorkspaceSlugUniqueness(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create hierarchy
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "test-app",
		Path:     "/path",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create first workspace
	workspace1 := &models.Workspace{
		AppID:     app.ID,
		Name:      "dev",
		Slug:      "test-eco-test-domain-test-app-dev",
		ImageName: "test:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace1); err != nil {
		t.Fatalf("CreateWorkspace() first error = %v", err)
	}

	// Attempt to create second workspace with same slug
	workspace2 := &models.Workspace{
		AppID:     app.ID,
		Name:      "dev",
		Slug:      "test-eco-test-domain-test-app-dev", // Same slug
		ImageName: "test:latest",
		Status:    "stopped",
	}
	err := ds.CreateWorkspace(workspace2)

	if err == nil {
		t.Errorf("CreateWorkspace() expected error for duplicate slug, got nil")
	}
	if err != nil {
		errLower := strings.ToLower(err.Error())
		if !strings.Contains(errLower, "unique") && !strings.Contains(errLower, "duplicate") {
			t.Errorf("CreateWorkspace() expected uniqueness error, got: %v", err)
		}
	}
}

// TestWorkspaceSSHAgentDefault verifies new workspaces have ssh_agent_forwarding = false
func TestWorkspaceSSHAgentDefault(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create hierarchy
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "test-app",
		Path:     "/path",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create workspace without setting SSH agent
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "dev",
		Slug:      "test-eco-test-domain-test-app-dev",
		ImageName: "test:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}

	// Retrieve and verify SSH agent is disabled by default
	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspaceByID() error = %v", err)
	}

	if retrieved.SSHAgentForwarding {
		t.Errorf("CreateWorkspace() default SSHAgentForwarding = true, want false")
	}
}

// TestWorkspaceSSHAgentExplicitEnable verifies SSH agent can be explicitly enabled
func TestWorkspaceSSHAgentExplicitEnable(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create hierarchy
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "test-app",
		Path:     "/path",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create workspace with SSH agent explicitly enabled
	workspace := &models.Workspace{
		AppID:              app.ID,
		Name:               "dev",
		Slug:               "test-eco-test-domain-test-app-dev",
		ImageName:          "test:latest",
		Status:             "stopped",
		SSHAgentForwarding: true,
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}

	// Retrieve and verify SSH agent is enabled
	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspaceByID() error = %v", err)
	}

	if !retrieved.SSHAgentForwarding {
		t.Errorf("CreateWorkspace() SSHAgentForwarding = false, want true")
	}
}

// TestWorkspaceSlugInList verifies that ListAllWorkspaces includes slug
func TestWorkspaceSlugInList(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create hierarchy
	ecosystem := &models.Ecosystem{Name: "eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "dom",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "app",
		Path:     "/path",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	expectedSlug := "eco-dom-app-ws1"
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "ws1",
		Slug:      expectedSlug,
		ImageName: "test:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}

	// List all workspaces
	workspaces, err := ds.ListAllWorkspaces()
	if err != nil {
		t.Fatalf("ListAllWorkspaces() error = %v", err)
	}

	if len(workspaces) == 0 {
		t.Fatal("ListAllWorkspaces() returned empty list")
	}

	found := false
	for _, ws := range workspaces {
		if ws.ID == workspace.ID {
			found = true
			if ws.Slug != expectedSlug {
				t.Errorf("ListAllWorkspaces() Slug = %q, want %q", ws.Slug, expectedSlug)
			}
		}
	}

	if !found {
		t.Error("ListAllWorkspaces() did not return created workspace")
	}
}
