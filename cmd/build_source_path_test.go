package cmd

import (
	"database/sql"
	"devopsmaestro/db"
	"devopsmaestro/models"
	ws "devopsmaestro/pkg/workspace"
	"testing"
)

// TestBuildSourcePath_WithGitRepoID_UsesWorkspaceRepoPath verifies that when a workspace
// has a GitRepoID set (created with --repo flag), the build uses the workspace repo path
// (where the code was cloned) instead of the app path.
func TestBuildSourcePath_WithGitRepoID_UsesWorkspaceRepoPath(t *testing.T) {
	// Arrange: Create a workspace with GitRepoID
	mockStore := db.NewMockDataStore()

	// Create ecosystem, domain, and app
	ecosystem := &models.Ecosystem{
		Name: "test-ecosystem",
	}
	mockStore.CreateEcosystem(ecosystem)

	domain := &models.Domain{
		Name:        "test-domain",
		EcosystemID: sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true},
	}
	mockStore.CreateDomain(domain)

	app := &models.App{
		Name:     "test-app",
		DomainID: sql.NullInt64{Int64: int64(domain.ID), Valid: true},
		Path:     "/original/app/path",
	}
	mockStore.CreateApp(app)

	// Create workspace with GitRepoID (indicating it was created with --repo)
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "test-workspace",
		Slug:      "test-ecosystem-test-domain-test-app-test-workspace",
		ImageName: "test-image",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Int64: 1, Valid: true}, // Has a git repo
	}
	mockStore.CreateWorkspace(workspace)

	// Act: Get the source path that should be used for building
	sourcePath, err := getBuildSourcePath(mockStore, workspace, app.Path)

	// Assert: Should use workspace repo path, not app path
	if err != nil {
		t.Fatalf("getBuildSourcePath returned error: %v", err)
	}

	// When GitRepoID is set, we MUST use the repo path, even if it doesn't exist yet
	// The directory will be created during the git clone step
	expectedRepoPath, _ := ws.GetWorkspaceRepoPath(workspace.Slug)
	if sourcePath != expectedRepoPath {
		t.Errorf("Expected source path to be workspace repo path %q, got %q", expectedRepoPath, sourcePath)
	}

	// This is the key assertion: MUST NOT use app.Path when GitRepoID is set
	if sourcePath == app.Path {
		t.Errorf("BUG: Source path should NOT be app.Path (%q) when GitRepoID is set. Expected workspace repo path.", app.Path)
	}
}

// TestBuildSourcePath_WithoutGitRepoID_UsesAppPath verifies that when a workspace
// does NOT have a GitRepoID (regular workspace without --repo flag), the build uses
// the app.Path as expected.
func TestBuildSourcePath_WithoutGitRepoID_UsesAppPath(t *testing.T) {
	// Arrange: Create a workspace without GitRepoID
	mockStore := db.NewMockDataStore()

	// Create ecosystem, domain, and app
	ecosystem := &models.Ecosystem{
		Name: "test-ecosystem",
	}
	mockStore.CreateEcosystem(ecosystem)

	domain := &models.Domain{
		Name:        "test-domain",
		EcosystemID: sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true},
	}
	mockStore.CreateDomain(domain)

	app := &models.App{
		Name:     "test-app",
		DomainID: sql.NullInt64{Int64: int64(domain.ID), Valid: true},
		Path:     "/original/app/path",
	}
	mockStore.CreateApp(app)

	// Create workspace WITHOUT GitRepoID (regular workspace)
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "test-workspace",
		Slug:      "test-ecosystem-test-domain-test-app-test-workspace",
		ImageName: "test-image",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Valid: false}, // No git repo
	}
	mockStore.CreateWorkspace(workspace)

	// Act: Get the source path that should be used for building
	sourcePath, err := getBuildSourcePath(mockStore, workspace, app.Path)

	// Assert: Should use app.Path
	if err != nil {
		t.Fatalf("getBuildSourcePath returned error: %v", err)
	}

	if sourcePath != app.Path {
		t.Errorf("Expected source path to be app.Path %q, got %q", app.Path, sourcePath)
	}
}

// TestBuildSourcePath_WithGitRepoID_AlwaysUsesRepoPath verifies that
// when a workspace has a GitRepoID, we ALWAYS use the repo path,
// even if the directory doesn't exist yet (it will be created during git clone).
func TestBuildSourcePath_WithGitRepoID_AlwaysUsesRepoPath(t *testing.T) {
	// Arrange: Create a workspace with GitRepoID
	mockStore := db.NewMockDataStore()

	// Create ecosystem, domain, and app
	ecosystem := &models.Ecosystem{
		Name: "test-ecosystem",
	}
	mockStore.CreateEcosystem(ecosystem)

	domain := &models.Domain{
		Name:        "test-domain",
		EcosystemID: sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true},
	}
	mockStore.CreateDomain(domain)

	app := &models.App{
		Name:     "test-app",
		DomainID: sql.NullInt64{Int64: int64(domain.ID), Valid: true},
		Path:     "/original/app/path",
	}
	mockStore.CreateApp(app)

	// Create workspace with GitRepoID
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "test-workspace",
		Slug:      "test-ecosystem-test-domain-test-app-test-workspace",
		ImageName: "test-image",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Int64: 1, Valid: true},
	}
	mockStore.CreateWorkspace(workspace)

	// Act: Get the source path
	sourcePath, err := getBuildSourcePath(mockStore, workspace, app.Path)

	// Assert: Should ALWAYS use workspace repo path when GitRepoID is set
	if err != nil {
		t.Fatalf("getBuildSourcePath returned error: %v", err)
	}

	// The key requirement: when GitRepoID is set, always use repo path
	expectedRepoPath, _ := ws.GetWorkspaceRepoPath(workspace.Slug)
	if sourcePath != expectedRepoPath {
		t.Errorf("Expected source path to be workspace repo path %q, got %q", expectedRepoPath, sourcePath)
	}

	// Must NOT use app.Path
	if sourcePath == app.Path {
		t.Errorf("BUG: Source path should NOT be app.Path (%q) when GitRepoID is set", app.Path)
	}
}

// TestBuildCmd_Integration_UsesCorrectSourcePath is an integration test that verifies
// the entire build command uses the correct source path based on workspace.GitRepoID.
// This is skipped by default as it requires a real workspace setup.
func TestBuildCmd_Integration_UsesCorrectSourcePath(t *testing.T) {
	t.Skip("Integration test - requires full workspace setup with git repo")

	// This test would:
	// 1. Create a real workspace with --repo flag
	// 2. Run dvm build
	// 3. Verify that prepareStagingDirectory received the workspace repo path, not app.Path
	// 4. Verify that the staging directory contains code from the git clone
}

// Note: getBuildSourcePath is now implemented in build.go
// Tests import it from there
