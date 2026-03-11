package cmd

import (
	"database/sql"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	ws "devopsmaestro/pkg/workspace"
)

// TestGetMountPath_WithGitRepoID_UsesWorkspaceRepoPath verifies that when a workspace
// has a GitRepoID (created with --repo flag), the mount path should be the workspace
// repo path, not the app.Path. This prevents mounting an empty directory.
func TestGetMountPath_WithGitRepoID_UsesWorkspaceRepoPath(t *testing.T) {
	// Arrange: Create a workspace with GitRepoID and Slug set
	workspace := &models.Workspace{
		ID:        123,
		Name:      "feature-branch",
		Slug:      "eco-domain-my-app-feature-branch",
		GitRepoID: sql.NullInt64{Int64: 1, Valid: true},
	}

	app := &models.App{
		ID:   456,
		Name: "my-app",
		Path: "/tmp/dvm-test-apps/my-app", // This is NOT where the code is
	}

	// getMountPath now uses ws.GetWorkspaceRepoPath(slug) directly,
	// so the DataStore is not consulted for path lookups.
	mockStore := db.NewMockDataStore()

	// Act: Call getMountPath
	mountPath, err := getMountPath(mockStore, workspace, app.Path)

	// Assert: Should use workspace repo path, not app.Path
	if err != nil {
		t.Fatalf("getMountPath returned error: %v", err)
	}

	expectedRepoPath, _ := ws.GetWorkspaceRepoPath(workspace.Slug)
	if mountPath != expectedRepoPath {
		t.Errorf("Expected mount path to be workspace repo path %q, got %q", expectedRepoPath, mountPath)
	}

	// This is the key assertion: MUST NOT use app.Path when GitRepoID is set
	if mountPath == app.Path {
		t.Errorf("BUG: Mount path should NOT be app.Path (%q) when GitRepoID is set. Expected workspace repo path.", app.Path)
	}
}

// TestGetMountPath_WithoutGitRepoID_UsesAppPath verifies that when a workspace
// doesn't have a GitRepoID (traditional local app), it uses the app.Path as expected.
func TestGetMountPath_WithoutGitRepoID_UsesAppPath(t *testing.T) {
	// Arrange: Create a workspace WITHOUT GitRepoID
	workspace := &models.Workspace{
		ID:        123,
		Name:      "dev",
		GitRepoID: sql.NullInt64{Valid: false}, // NOT set
	}

	app := &models.App{
		ID:   456,
		Name: "local-app",
		Path: "/home/user/code/local-app", // This IS where the code is
	}

	mockStore := db.NewMockDataStore()

	// Act
	mountPath, err := getMountPath(mockStore, workspace, app.Path)

	// Assert: Should use app.Path since no GitRepoID
	if err != nil {
		t.Fatalf("getMountPath returned error: %v", err)
	}

	if mountPath != app.Path {
		t.Errorf("Expected mount path to be app.Path %q, got %q", app.Path, mountPath)
	}
}

// TestAttachMountPath_IntegrationScenario tests the integration scenario
// where workspace is created with --repo flag and should mount from workspace repo path
func TestAttachMountPath_IntegrationScenario(t *testing.T) {
	// This test documents the expected behavior:
	// 1. User creates workspace with: dvm create workspace dev --repo my-git-repo
	// 2. dvm clones the repo to ~/.devopsmaestro/workspaces/{slug}/repo/
	// 3. When attaching, the container mount should be from the repo path, not app.Path

	workspaceSlug := "test-eco-test-domain-golang-app-main"

	workspace := &models.Workspace{
		ID:        11,
		Name:      "main",
		Slug:      workspaceSlug,
		ImageName: "dvm-main-golang-app:20260303-202258",
		GitRepoID: sql.NullInt64{Int64: 5, Valid: true},
	}

	app := &models.App{
		ID:   2,
		Name: "golang-app",
		Path: "/tmp/dvm-test-apps/golang-app", // Empty directory
	}

	// getMountPath now uses ws.GetWorkspaceRepoPath(slug) directly
	mockStore := db.NewMockDataStore()
	expectedMountPath, _ := ws.GetWorkspaceRepoPath(workspaceSlug)

	// Act
	mountPath, err := getMountPath(mockStore, workspace, app.Path)

	// Assert
	if err != nil {
		t.Fatalf("getMountPath returned error: %v", err)
	}

	// Must use workspace repo path, NOT app.Path
	if mountPath == app.Path {
		t.Errorf("BUG: Mount path should NOT be app.Path (%q) when GitRepoID is set", app.Path)
	}

	if mountPath != expectedMountPath {
		t.Errorf("Expected mount path %q, got %q", expectedMountPath, mountPath)
	}
}
