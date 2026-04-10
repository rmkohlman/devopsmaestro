package db

// =============================================================================
// TDD Phase 2 (RED): GitRepo Relationship Queries — Issue #223
// =============================================================================
// These tests drive the implementation of two new DataStore methods:
//   - ListAppsByGitRepoID(gitRepoID int64) ([]*models.App, error)
//   - ListWorkspacesByGitRepoID(gitRepoID int64) ([]*models.Workspace, error)
//
// RED state: ALL tests FAIL — the methods do not exist yet on GitRepoStore
// or DataStore.
//
// GREEN state: After @database implements Issue #223 Phase 1.
// =============================================================================

import (
	"database/sql"
	"testing"

	"devopsmaestro/models"
)

// =============================================================================
// TestListAppsByGitRepoID
// =============================================================================

func TestListAppsByGitRepoID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Setup: ecosystem, domain, gitrepo
	eco := &models.Ecosystem{Name: "rel-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("setup: CreateEcosystem: %v", err)
	}
	dom := &models.Domain{EcosystemID: eco.ID, Name: "rel-dom"}
	if err := ds.CreateDomain(dom); err != nil {
		t.Fatalf("setup: CreateDomain: %v", err)
	}
	repo := &models.GitRepoDB{
		Name:       "rel-repo",
		URL:        "https://github.com/org/rel-repo",
		Slug:       "github.com_org_rel-repo",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	if err := ds.CreateGitRepo(repo); err != nil {
		t.Fatalf("setup: CreateGitRepo: %v", err)
	}

	// Create 2 apps linked to the gitrepo, 1 unlinked
	app1 := &models.App{
		DomainID:  dom.ID,
		Name:      "linked-app-1",
		Path:      "/path/linked-1",
		GitRepoID: sql.NullInt64{Int64: int64(repo.ID), Valid: true},
	}
	app2 := &models.App{
		DomainID:  dom.ID,
		Name:      "linked-app-2",
		Path:      "/path/linked-2",
		GitRepoID: sql.NullInt64{Int64: int64(repo.ID), Valid: true},
	}
	unlinked := &models.App{
		DomainID: dom.ID,
		Name:     "unlinked-app",
		Path:     "/path/unlinked",
	}
	for _, app := range []*models.App{app1, app2, unlinked} {
		if err := ds.CreateApp(app); err != nil {
			t.Fatalf("setup: CreateApp %q: %v", app.Name, err)
		}
	}

	// Act
	apps, err := ds.ListAppsByGitRepoID(int64(repo.ID))
	if err != nil {
		t.Fatalf("ListAppsByGitRepoID() error = %v", err)
	}

	// Assert: only 2 linked apps returned
	if len(apps) != 2 {
		t.Errorf("ListAppsByGitRepoID() returned %d apps, want 2", len(apps))
	}
	for _, app := range apps {
		if !app.GitRepoID.Valid || app.GitRepoID.Int64 != int64(repo.ID) {
			t.Errorf("ListAppsByGitRepoID() returned app %q not linked to gitrepo %d", app.Name, repo.ID)
		}
	}
}

func TestListAppsByGitRepoID_NoResults(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create a gitrepo with no linked apps
	repo := &models.GitRepoDB{
		Name:       "empty-repo",
		URL:        "https://github.com/org/empty-repo",
		Slug:       "github.com_org_empty-repo",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	if err := ds.CreateGitRepo(repo); err != nil {
		t.Fatalf("setup: CreateGitRepo: %v", err)
	}

	// Act
	apps, err := ds.ListAppsByGitRepoID(int64(repo.ID))
	if err != nil {
		t.Fatalf("ListAppsByGitRepoID() error = %v, want nil", err)
	}

	// Assert: empty slice, not nil
	if apps == nil {
		t.Errorf("ListAppsByGitRepoID() returned nil, want empty slice")
	}
	if len(apps) != 0 {
		t.Errorf("ListAppsByGitRepoID() returned %d apps, want 0", len(apps))
	}
}

// =============================================================================
// TestListWorkspacesByGitRepoID
// =============================================================================

func TestListWorkspacesByGitRepoID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Setup: full hierarchy
	eco := &models.Ecosystem{Name: "ws-rel-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("setup: CreateEcosystem: %v", err)
	}
	dom := &models.Domain{EcosystemID: eco.ID, Name: "ws-rel-dom"}
	if err := ds.CreateDomain(dom); err != nil {
		t.Fatalf("setup: CreateDomain: %v", err)
	}
	app := &models.App{DomainID: dom.ID, Name: "ws-rel-app", Path: "/ws/rel"}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("setup: CreateApp: %v", err)
	}
	repo := &models.GitRepoDB{
		Name:       "ws-rel-repo",
		URL:        "https://github.com/org/ws-rel-repo",
		Slug:       "github.com_org_ws-rel-repo",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	if err := ds.CreateGitRepo(repo); err != nil {
		t.Fatalf("setup: CreateGitRepo: %v", err)
	}

	// Create 2 workspaces linked to gitrepo, 1 unlinked
	ws1 := &models.Workspace{
		AppID:     app.ID,
		Name:      "linked-ws-1",
		Slug:      "ws-rel-eco.ws-rel-dom.ws-rel-app.linked-ws-1",
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Int64: int64(repo.ID), Valid: true},
	}
	ws2 := &models.Workspace{
		AppID:     app.ID,
		Name:      "linked-ws-2",
		Slug:      "ws-rel-eco.ws-rel-dom.ws-rel-app.linked-ws-2",
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Int64: int64(repo.ID), Valid: true},
	}
	wsUnlinked := &models.Workspace{
		AppID:     app.ID,
		Name:      "unlinked-ws",
		Slug:      "ws-rel-eco.ws-rel-dom.ws-rel-app.unlinked-ws",
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	for _, ws := range []*models.Workspace{ws1, ws2, wsUnlinked} {
		if err := ds.CreateWorkspace(ws); err != nil {
			t.Fatalf("setup: CreateWorkspace %q: %v", ws.Name, err)
		}
	}

	// Act
	workspaces, err := ds.ListWorkspacesByGitRepoID(int64(repo.ID))
	if err != nil {
		t.Fatalf("ListWorkspacesByGitRepoID() error = %v", err)
	}

	// Assert: only 2 linked workspaces returned
	if len(workspaces) != 2 {
		t.Errorf("ListWorkspacesByGitRepoID() returned %d workspaces, want 2", len(workspaces))
	}
	for _, ws := range workspaces {
		if !ws.GitRepoID.Valid || ws.GitRepoID.Int64 != int64(repo.ID) {
			t.Errorf("ListWorkspacesByGitRepoID() returned workspace %q not linked to gitrepo %d", ws.Name, repo.ID)
		}
	}
}

func TestListWorkspacesByGitRepoID_NoResults(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create a gitrepo with no linked workspaces
	repo := &models.GitRepoDB{
		Name:       "ws-empty-repo",
		URL:        "https://github.com/org/ws-empty-repo",
		Slug:       "github.com_org_ws-empty-repo",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	if err := ds.CreateGitRepo(repo); err != nil {
		t.Fatalf("setup: CreateGitRepo: %v", err)
	}

	// Act
	workspaces, err := ds.ListWorkspacesByGitRepoID(int64(repo.ID))
	if err != nil {
		t.Fatalf("ListWorkspacesByGitRepoID() error = %v, want nil", err)
	}

	// Assert: empty slice, not nil
	if workspaces == nil {
		t.Errorf("ListWorkspacesByGitRepoID() returned nil, want empty slice")
	}
	if len(workspaces) != 0 {
		t.Errorf("ListWorkspacesByGitRepoID() returned %d workspaces, want 0", len(workspaces))
	}
}
