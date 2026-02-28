package db

import (
	"devopsmaestro/models"
	"fmt"
	"os"
	"strings"
	"testing"
)

// =============================================================================
// Task 2.1: Workspace Slug Generation and Path Tests (v0.19.0)
// =============================================================================

// TestWorkspaceSlugGeneration verifies that workspace slugs are computed correctly
// from the hierarchy: {ecosystem}-{domain}-{app}-{workspace}
func TestWorkspaceSlugGeneration(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	tests := []struct {
		name         string
		ecosystem    string
		domain       string
		app          string
		workspace    string
		expectedSlug string
	}{
		{
			name:         "simple slug",
			ecosystem:    "personal",
			domain:       "tools",
			app:          "dvm",
			workspace:    "dev",
			expectedSlug: "personal-tools-dvm-dev",
		},
		{
			name:         "production workspace",
			ecosystem:    "enterprise",
			domain:       "payments",
			app:          "api",
			workspace:    "prod",
			expectedSlug: "enterprise-payments-api-prod",
		},
		{
			name:         "multi-word names with hyphens",
			ecosystem:    "my-org",
			domain:       "data-platform",
			app:          "etl-service",
			workspace:    "local-dev",
			expectedSlug: "my-org-data-platform-etl-service-local-dev",
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
				ImageName: "test:latest",
				Status:    "stopped",
			}

			if err := ds.CreateWorkspace(workspace); err != nil {
				t.Fatalf("CreateWorkspace() error = %v", err)
			}

			// Get workspace and verify slug
			retrieved, err := ds.GetWorkspaceByID(workspace.ID)
			if err != nil {
				t.Fatalf("GetWorkspaceByID() error = %v", err)
			}

			// FIXME: This test will FAIL - Slug field doesn't exist yet
			// After Phase 3 implementation, Workspace model should have:
			// type Workspace struct {
			//     ...
			//     Slug string `db:"slug" json:"slug" yaml:"slug"`
			// }
			if retrieved.Slug != tt.expectedSlug {
				t.Errorf("GetWorkspaceByID() Slug = %q, want %q", retrieved.Slug, tt.expectedSlug)
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
		ImageName: "test:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace1); err != nil {
		t.Fatalf("CreateWorkspace() first error = %v", err)
	}

	// Attempt to create second workspace with same name (same slug)
	workspace2 := &models.Workspace{
		AppID:     app.ID,
		Name:      "dev", // Same name = same slug
		ImageName: "test:latest",
		Status:    "stopped",
	}
	err := ds.CreateWorkspace(workspace2)

	// FIXME: This test will FAIL - Slug uniqueness constraint doesn't exist yet
	// After Phase 3, database should have: UNIQUE(slug) or UNIQUE(app_id, name)
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

// TestWorkspaceSlugFormat validates that slugs follow the correct format
func TestWorkspaceSlugFormat(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	tests := []struct {
		name               string
		ecosystemName      string
		domainName         string
		appName            string
		workspaceName      string
		wantSlugPattern    string // Expected pattern
		wantComponentCount int    // Should be 4 parts
	}{
		{
			name:               "standard format",
			ecosystemName:      "eco",
			domainName:         "dom",
			appName:            "app",
			workspaceName:      "ws",
			wantSlugPattern:    "eco-dom-app-ws",
			wantComponentCount: 4,
		},
		{
			name:               "hyphenated names",
			ecosystemName:      "my-eco",
			domainName:         "my-domain",
			appName:            "my-app",
			workspaceName:      "my-ws",
			wantSlugPattern:    "my-eco-my-domain-my-app-my-ws",
			wantComponentCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create hierarchy
			ecosystem := &models.Ecosystem{Name: tt.ecosystemName}
			if err := ds.CreateEcosystem(ecosystem); err != nil {
				t.Fatalf("Setup error: %v", err)
			}

			domain := &models.Domain{
				EcosystemID: ecosystem.ID,
				Name:        tt.domainName,
			}
			if err := ds.CreateDomain(domain); err != nil {
				t.Fatalf("Setup error: %v", err)
			}

			app := &models.App{
				DomainID: domain.ID,
				Name:     tt.appName,
				Path:     "/path",
			}
			if err := ds.CreateApp(app); err != nil {
				t.Fatalf("Setup error: %v", err)
			}

			workspace := &models.Workspace{
				AppID:     app.ID,
				Name:      tt.workspaceName,
				ImageName: "test:latest",
				Status:    "stopped",
			}
			if err := ds.CreateWorkspace(workspace); err != nil {
				t.Fatalf("CreateWorkspace() error = %v", err)
			}

			// Retrieve workspace and check slug format
			retrieved, err := ds.GetWorkspaceByID(workspace.ID)
			if err != nil {
				t.Fatalf("GetWorkspaceByID() error = %v", err)
			}

			// FIXME: This test will FAIL - Slug field doesn't exist yet
			if retrieved.Slug != tt.wantSlugPattern {
				t.Errorf("Slug format = %q, want %q", retrieved.Slug, tt.wantSlugPattern)
			}

			// Verify slug has correct number of components
			components := strings.Split(retrieved.Slug, "-")
			// Note: With hyphenated names, count actual hierarchy levels not hyphens
			// We should verify ecosystem-domain-app-workspace structure
			if len(components) < tt.wantComponentCount {
				t.Errorf("Slug has %d components, want at least %d", len(components), tt.wantComponentCount)
			}
		})
	}
}

// TestGetWorkspacePath returns correct base path for workspace
func TestGetWorkspacePath(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create hierarchy
	ecosystem := &models.Ecosystem{Name: "personal"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "tools",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "dvm",
		Path:     "/path",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "dev",
		ImageName: "test:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}

	// FIXME: This test will FAIL - GetWorkspacePath() method doesn't exist yet
	// After Phase 3, DataStore should have:
	// GetWorkspacePath(workspaceID int) (string, error)
	// Returns: ~/.devopsmaestro/workspaces/{slug}/
	path, err := ds.GetWorkspacePath(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspacePath() error = %v", err)
	}

	expectedPath := fmt.Sprintf("%s/.devopsmaestro/workspaces/personal-tools-dvm-dev/", homeDir())
	if path != expectedPath {
		t.Errorf("GetWorkspacePath() = %q, want %q", path, expectedPath)
	}

	// Verify subdirectories can be constructed
	tests := []struct {
		name         string
		subdirectory string
		wantContains string
	}{
		{
			name:         "repo directory",
			subdirectory: "repo",
			wantContains: "/workspaces/personal-tools-dvm-dev/repo",
		},
		{
			name:         "volume directory",
			subdirectory: "volume",
			wantContains: "/workspaces/personal-tools-dvm-dev/volume",
		},
		{
			name:         "generated config directory",
			subdirectory: ".dvm",
			wantContains: "/workspaces/personal-tools-dvm-dev/.dvm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := fmt.Sprintf("%s%s", path, tt.subdirectory)
			if !strings.Contains(fullPath, tt.wantContains) {
				t.Errorf("Path %q does not contain %q", fullPath, tt.wantContains)
			}
		})
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

	// FIXME: This test will FAIL - SSHAgentForwarding field doesn't exist yet
	// After Phase 3, Workspace model should have:
	// type Workspace struct {
	//     ...
	//     SSHAgentForwarding bool `db:"ssh_agent_forwarding" json:"ssh_agent_forwarding" yaml:"ssh_agent_forwarding"`
	// }
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
		ImageName:          "test:latest",
		Status:             "stopped",
		SSHAgentForwarding: true, // FIXME: Field doesn't exist yet
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}

	// Retrieve and verify SSH agent is enabled
	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspaceByID() error = %v", err)
	}

	// FIXME: This test will FAIL - SSHAgentForwarding field doesn't exist yet
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

	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "ws1",
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

	// FIXME: This test will FAIL - Slug field doesn't exist yet
	// After Phase 3, all workspace queries should populate Slug
	found := false
	for _, ws := range workspaces {
		if ws.ID == workspace.ID {
			found = true
			expectedSlug := "eco-dom-app-ws1"
			if ws.Slug != expectedSlug {
				t.Errorf("ListAllWorkspaces() Slug = %q, want %q", ws.Slug, expectedSlug)
			}
		}
	}

	if !found {
		t.Error("ListAllWorkspaces() did not return created workspace")
	}
}

// homeDir returns a mock home directory for testing
// In real implementation, this would use os.UserHomeDir()
func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/home/testuser" // Fallback for testing
	}
	return home
}
