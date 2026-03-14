package db

import (
	"database/sql"
	"devopsmaestro/models"
	"testing"
	"time"
)

// =============================================================================
// GitRepo CRUD Tests
// =============================================================================

func TestSQLDataStore_CreateGitRepo(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	repo := &models.GitRepoDB{
		Name:                "test-repo",
		URL:                 "https://github.com/user/test-repo",
		Slug:                "github.com_user_test-repo",
		DefaultRef:          "main",
		AuthType:            "none",
		AutoSync:            false,
		SyncIntervalMinutes: 0,
		SyncStatus:          "pending",
	}

	err := ds.CreateGitRepo(repo)
	if err != nil {
		t.Fatalf("CreateGitRepo() error = %v", err)
	}

	if repo.ID == 0 {
		t.Errorf("CreateGitRepo() did not set repo.ID")
	}
}

func TestSQLDataStore_CreateGitRepo_DuplicateName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	repo1 := &models.GitRepoDB{
		Name:       "duplicate-repo",
		URL:        "https://github.com/user/repo1",
		Slug:       "github.com_user_repo1",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}

	err := ds.CreateGitRepo(repo1)
	if err != nil {
		t.Fatalf("Setup: CreateGitRepo() error = %v", err)
	}

	// Try to create another repo with the same name
	repo2 := &models.GitRepoDB{
		Name:       "duplicate-repo",
		URL:        "https://github.com/user/repo2",
		Slug:       "github.com_user_repo2",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}

	err = ds.CreateGitRepo(repo2)
	if err == nil {
		t.Errorf("CreateGitRepo() with duplicate name should have returned error")
	}
}

func TestSQLDataStore_GetGitRepoByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create a credential to satisfy FK constraint
	svc := "test-service"
	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     0,
		Name:        "test-cred-for-gitrepo",
		Source:      "vault",
		VaultSecret: &svc,
	}
	if err := ds.CreateCredential(cred); err != nil {
		t.Fatalf("Setup: CreateCredential() error = %v", err)
	}

	repo := &models.GitRepoDB{
		Name:       "findme-repo",
		URL:        "https://github.com/user/findme",
		Slug:       "github.com_user_findme",
		DefaultRef: "main",
		AuthType:   "ssh",
		CredentialID: sql.NullInt64{
			Int64: cred.ID,
			Valid: true,
		},
		AutoSync:            true,
		SyncIntervalMinutes: 30,
		SyncStatus:          "synced",
	}

	if err := ds.CreateGitRepo(repo); err != nil {
		t.Fatalf("Setup: CreateGitRepo() error = %v", err)
	}

	retrieved, err := ds.GetGitRepoByName("findme-repo")
	if err != nil {
		t.Fatalf("GetGitRepoByName() error = %v", err)
	}

	if retrieved.Name != "findme-repo" {
		t.Errorf("GetGitRepoByName() Name = %q, want %q", retrieved.Name, "findme-repo")
	}
	if retrieved.URL != "https://github.com/user/findme" {
		t.Errorf("GetGitRepoByName() URL = %q, want %q", retrieved.URL, "https://github.com/user/findme")
	}
	if retrieved.Slug != "github.com_user_findme" {
		t.Errorf("GetGitRepoByName() Slug = %q, want %q", retrieved.Slug, "github.com_user_findme")
	}
	if retrieved.DefaultRef != "main" {
		t.Errorf("GetGitRepoByName() DefaultRef = %q, want %q", retrieved.DefaultRef, "main")
	}
	if retrieved.AuthType != "ssh" {
		t.Errorf("GetGitRepoByName() AuthType = %q, want %q", retrieved.AuthType, "ssh")
	}
	if !retrieved.CredentialID.Valid || retrieved.CredentialID.Int64 != cred.ID {
		t.Errorf("GetGitRepoByName() CredentialID = %v, want %d", retrieved.CredentialID, cred.ID)
	}
	if !retrieved.AutoSync {
		t.Errorf("GetGitRepoByName() AutoSync = %v, want true", retrieved.AutoSync)
	}
	if retrieved.SyncIntervalMinutes != 30 {
		t.Errorf("GetGitRepoByName() SyncIntervalMinutes = %d, want 30", retrieved.SyncIntervalMinutes)
	}
	if retrieved.SyncStatus != "synced" {
		t.Errorf("GetGitRepoByName() SyncStatus = %q, want %q", retrieved.SyncStatus, "synced")
	}
}

func TestSQLDataStore_GetGitRepoByName_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	_, err := ds.GetGitRepoByName("nonexistent-repo")
	if err == nil {
		t.Errorf("GetGitRepoByName() expected error for nonexistent repo")
	}
}

func TestSQLDataStore_GetGitRepoBySlug(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	repo := &models.GitRepoDB{
		Name:       "slug-test-repo",
		URL:        "https://gitlab.com/team/project",
		Slug:       "gitlab.com_team_project",
		DefaultRef: "develop",
		AuthType:   "none",
		SyncStatus: "pending",
	}

	if err := ds.CreateGitRepo(repo); err != nil {
		t.Fatalf("Setup: CreateGitRepo() error = %v", err)
	}

	retrieved, err := ds.GetGitRepoBySlug("gitlab.com_team_project")
	if err != nil {
		t.Fatalf("GetGitRepoBySlug() error = %v", err)
	}

	if retrieved.Name != "slug-test-repo" {
		t.Errorf("GetGitRepoBySlug() Name = %q, want %q", retrieved.Name, "slug-test-repo")
	}
	if retrieved.Slug != "gitlab.com_team_project" {
		t.Errorf("GetGitRepoBySlug() Slug = %q, want %q", retrieved.Slug, "gitlab.com_team_project")
	}
	if retrieved.DefaultRef != "develop" {
		t.Errorf("GetGitRepoBySlug() DefaultRef = %q, want %q", retrieved.DefaultRef, "develop")
	}
}

func TestSQLDataStore_GetGitRepoBySlug_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	_, err := ds.GetGitRepoBySlug("nonexistent_slug")
	if err == nil {
		t.Errorf("GetGitRepoBySlug() expected error for nonexistent slug")
	}
}

func TestSQLDataStore_UpdateGitRepo(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create a credential to satisfy FK constraint on update
	svc := "update-service"
	cred := &models.CredentialDB{
		ScopeType:   models.CredentialScopeEcosystem,
		ScopeID:     0,
		Name:        "test-cred-for-update",
		Source:      "vault",
		VaultSecret: &svc,
	}
	if err := ds.CreateCredential(cred); err != nil {
		t.Fatalf("Setup: CreateCredential() error = %v", err)
	}

	repo := &models.GitRepoDB{
		Name:                "update-repo",
		URL:                 "https://github.com/user/original",
		Slug:                "github.com_user_original",
		DefaultRef:          "main",
		AuthType:            "none",
		AutoSync:            false,
		SyncIntervalMinutes: 0,
		SyncStatus:          "pending",
	}

	if err := ds.CreateGitRepo(repo); err != nil {
		t.Fatalf("Setup: CreateGitRepo() error = %v", err)
	}

	// Update the repo
	repo.DefaultRef = "develop"
	repo.AuthType = "ssh"
	repo.CredentialID = sql.NullInt64{Int64: cred.ID, Valid: true}
	repo.AutoSync = true
	repo.SyncIntervalMinutes = 60
	repo.SyncStatus = "synced"
	repo.LastSyncedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err := ds.UpdateGitRepo(repo); err != nil {
		t.Fatalf("UpdateGitRepo() error = %v", err)
	}

	// Verify update
	retrieved, err := ds.GetGitRepoByName("update-repo")
	if err != nil {
		t.Fatalf("Verification: GetGitRepoByName() error = %v", err)
	}

	if retrieved.DefaultRef != "develop" {
		t.Errorf("UpdateGitRepo() DefaultRef = %q, want %q", retrieved.DefaultRef, "develop")
	}
	if retrieved.AuthType != "ssh" {
		t.Errorf("UpdateGitRepo() AuthType = %q, want %q", retrieved.AuthType, "ssh")
	}
	if !retrieved.CredentialID.Valid || retrieved.CredentialID.Int64 != cred.ID {
		t.Errorf("UpdateGitRepo() CredentialID = %v, want %d", retrieved.CredentialID, cred.ID)
	}
	if !retrieved.AutoSync {
		t.Errorf("UpdateGitRepo() AutoSync = %v, want true", retrieved.AutoSync)
	}
	if retrieved.SyncIntervalMinutes != 60 {
		t.Errorf("UpdateGitRepo() SyncIntervalMinutes = %d, want 60", retrieved.SyncIntervalMinutes)
	}
	if retrieved.SyncStatus != "synced" {
		t.Errorf("UpdateGitRepo() SyncStatus = %q, want %q", retrieved.SyncStatus, "synced")
	}
	if !retrieved.LastSyncedAt.Valid {
		t.Errorf("UpdateGitRepo() LastSyncedAt should be valid")
	}
}

func TestSQLDataStore_UpdateGitRepo_SyncError(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	repo := &models.GitRepoDB{
		Name:       "error-repo",
		URL:        "https://github.com/user/repo",
		Slug:       "github.com_user_repo",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}

	if err := ds.CreateGitRepo(repo); err != nil {
		t.Fatalf("Setup: CreateGitRepo() error = %v", err)
	}

	// Update with sync error
	repo.SyncStatus = "error"
	repo.SyncError = sql.NullString{String: "authentication failed", Valid: true}
	repo.LastSyncedAt = sql.NullTime{Time: time.Now(), Valid: true}

	if err := ds.UpdateGitRepo(repo); err != nil {
		t.Fatalf("UpdateGitRepo() error = %v", err)
	}

	// Verify error was stored
	retrieved, err := ds.GetGitRepoByName("error-repo")
	if err != nil {
		t.Fatalf("Verification: GetGitRepoByName() error = %v", err)
	}

	if retrieved.SyncStatus != "error" {
		t.Errorf("UpdateGitRepo() SyncStatus = %q, want %q", retrieved.SyncStatus, "error")
	}
	if !retrieved.SyncError.Valid || retrieved.SyncError.String != "authentication failed" {
		t.Errorf("UpdateGitRepo() SyncError = %v, want 'authentication failed'", retrieved.SyncError)
	}
}

func TestSQLDataStore_DeleteGitRepo(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	repo := &models.GitRepoDB{
		Name:       "delete-repo",
		URL:        "https://github.com/user/delete",
		Slug:       "github.com_user_delete",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}

	if err := ds.CreateGitRepo(repo); err != nil {
		t.Fatalf("Setup: CreateGitRepo() error = %v", err)
	}

	if err := ds.DeleteGitRepo("delete-repo"); err != nil {
		t.Fatalf("DeleteGitRepo() error = %v", err)
	}

	// Verify deletion
	_, err := ds.GetGitRepoByName("delete-repo")
	if err == nil {
		t.Errorf("DeleteGitRepo() repo should not exist after deletion")
	}
}

func TestSQLDataStore_ListGitRepos(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create multiple repos
	repos := []struct {
		name string
		url  string
		slug string
	}{
		{"repo1", "https://github.com/user/repo1", "github.com_user_repo1"},
		{"repo2", "https://gitlab.com/user/repo2", "gitlab.com_user_repo2"},
		{"repo3", "https://bitbucket.org/user/repo3", "bitbucket.org_user_repo3"},
	}

	for _, r := range repos {
		repo := &models.GitRepoDB{
			Name:       r.name,
			URL:        r.url,
			Slug:       r.slug,
			DefaultRef: "main",
			AuthType:   "none",
			SyncStatus: "pending",
		}
		if err := ds.CreateGitRepo(repo); err != nil {
			t.Fatalf("Setup: CreateGitRepo(%q) error = %v", r.name, err)
		}
	}

	// List all repos
	retrieved, err := ds.ListGitRepos()
	if err != nil {
		t.Fatalf("ListGitRepos() error = %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("ListGitRepos() returned %d repos, want 3", len(retrieved))
	}

	// Verify all repos are present
	names := make(map[string]bool)
	for _, r := range retrieved {
		names[r.Name] = true
	}

	for _, r := range repos {
		if !names[r.name] {
			t.Errorf("ListGitRepos() missing repo %q", r.name)
		}
	}
}

func TestSQLDataStore_ListGitRepos_Empty(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	repos, err := ds.ListGitRepos()
	if err != nil {
		t.Fatalf("ListGitRepos() error = %v", err)
	}

	if len(repos) != 0 {
		t.Errorf("ListGitRepos() on empty database returned %d repos, want 0", len(repos))
	}
}

func TestSQLDataStore_ListGitRepos_FilterByAutoSync(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create repos with different AutoSync settings
	repos := []struct {
		name     string
		autoSync bool
	}{
		{"auto-sync-1", true},
		{"auto-sync-2", true},
		{"manual-sync-1", false},
		{"manual-sync-2", false},
	}

	for _, r := range repos {
		repo := &models.GitRepoDB{
			Name:       r.name,
			URL:        "https://github.com/user/" + r.name,
			Slug:       "github.com_user_" + r.name,
			DefaultRef: "main",
			AuthType:   "none",
			AutoSync:   r.autoSync,
			SyncStatus: "pending",
		}
		if err := ds.CreateGitRepo(repo); err != nil {
			t.Fatalf("Setup: CreateGitRepo(%q) error = %v", r.name, err)
		}
	}

	// This test documents that we'll need a filter method
	// For now, just list all and verify count
	all, err := ds.ListGitRepos()
	if err != nil {
		t.Fatalf("ListGitRepos() error = %v", err)
	}

	if len(all) != 4 {
		t.Errorf("ListGitRepos() returned %d repos, want 4", len(all))
	}

	// Count auto-sync repos manually (implementation will add a filter method)
	autoSyncCount := 0
	for _, repo := range all {
		if repo.AutoSync {
			autoSyncCount++
		}
	}

	if autoSyncCount != 2 {
		t.Errorf("Found %d auto-sync repos, want 2", autoSyncCount)
	}
}
