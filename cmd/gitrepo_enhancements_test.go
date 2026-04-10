package cmd

// =============================================================================
// TDD Phase 2 (RED): GitRepo Enhancements — Issue #223
// =============================================================================
// These tests drive the implementation of new CLI commands:
//   - dvm describe gitrepo <name>  (runDescribeGitRepo)
//   - dvm get branches --repo <name>  (runGetBranches)
//   - dvm get tags --repo <name>      (runGetTags)
//   - dvm edit gitrepo <name>         (runEditGitRepo — non-interactive path)
//
// RED state: ALL tests FAIL because:
//   - runDescribeGitRepo, runGetBranches, runGetTags do not exist
//   - CtxKeyMirrorManager / getMirrorManager DI helper not implemented
//   - MirrorInspector interface not implemented
//   - ListAppsByGitRepoID, ListWorkspacesByGitRepoID not in DataStore
//
// GREEN state: After @dvm-core implements Issue #223.
// =============================================================================

import (
	"bytes"
	"context"
	"database/sql"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/mirror"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Mock MirrorInspector for command tests
// =============================================================================

// mockMirrorInspector implements both mirror.MirrorManager and mirror.MirrorInspector.
// It allows command tests to avoid real git operations.
type mockMirrorInspector struct {
	// MirrorManager fields
	existsFunc  func(slug string) bool
	getPathFunc func(slug string) string

	// MirrorInspector fields
	branches   []mirror.RefInfo
	tags       []mirror.RefInfo
	diskUsage  int64
	verifyErr  error
	inspectErr error
}

func (m *mockMirrorInspector) Clone(url, slug string) (string, error) { return "", nil }
func (m *mockMirrorInspector) Sync(slug string) error                 { return nil }
func (m *mockMirrorInspector) Delete(slug string) error               { return nil }
func (m *mockMirrorInspector) Exists(slug string) bool {
	if m.existsFunc != nil {
		return m.existsFunc(slug)
	}
	return true
}
func (m *mockMirrorInspector) GetPath(slug string) string {
	if m.getPathFunc != nil {
		return m.getPathFunc(slug)
	}
	return "/mock/path/" + slug
}
func (m *mockMirrorInspector) CloneToWorkspace(mirrorSlug, destPath, ref string) error { return nil }

// MirrorInspector methods
func (m *mockMirrorInspector) ListBranches(slug string) ([]mirror.RefInfo, error) {
	if m.inspectErr != nil {
		return nil, m.inspectErr
	}
	return m.branches, nil
}
func (m *mockMirrorInspector) ListTags(slug string) ([]mirror.RefInfo, error) {
	if m.inspectErr != nil {
		return nil, m.inspectErr
	}
	return m.tags, nil
}
func (m *mockMirrorInspector) DiskUsage(slug string) (int64, error) {
	if m.inspectErr != nil {
		return 0, m.inspectErr
	}
	return m.diskUsage, nil
}
func (m *mockMirrorInspector) Verify(slug string) error {
	return m.verifyErr
}

// =============================================================================
// Test helpers — build describe/get commands
// =============================================================================

func newTestDescribeGitRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "gitrepo <name>",
		Args: cobra.ExactArgs(1),
		RunE: runDescribeGitRepo,
	}
	cmd.Flags().StringP("output", "o", "yaml", "Output format (yaml, json)")
	return cmd
}

func newTestGetBranchesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "branches",
		RunE: runGetBranches,
	}
	cmd.Flags().String("repo", "", "GitRepo name (required)")
	cmd.Flags().StringP("output", "o", "table", "Output format")
	return cmd
}

func newTestGetTagsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "tags",
		RunE: runGetTags,
	}
	cmd.Flags().String("repo", "", "GitRepo name (required)")
	cmd.Flags().StringP("output", "o", "table", "Output format")
	return cmd
}

// setupDescribeTestStore builds a mock store with a gitrepo + linked apps/workspaces.
func setupDescribeTestStore(t *testing.T) (*db.MockDataStore, *models.GitRepoDB) {
	t.Helper()
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "desc-eco"}
	require.NoError(t, store.CreateEcosystem(eco))

	dom := &models.Domain{Name: "desc-dom", EcosystemID: eco.ID}
	require.NoError(t, store.CreateDomain(dom))

	repo := &models.GitRepoDB{
		Name:       "describe-repo",
		URL:        "https://github.com/org/describe-repo.git",
		Slug:       "github.com_org_describe-repo",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "synced",
	}
	require.NoError(t, store.CreateGitRepo(repo))

	// Link 2 apps to the gitrepo
	for i := 1; i <= 2; i++ {
		app := &models.App{
			DomainID:  dom.ID,
			Name:      "desc-app-" + string(rune('0'+i)),
			Path:      "/path/desc/" + string(rune('0'+i)),
			GitRepoID: sql.NullInt64{Int64: int64(repo.ID), Valid: true},
		}
		require.NoError(t, store.CreateApp(app))

		// Link 1 workspace per app
		ws := &models.Workspace{
			AppID:     app.ID,
			Name:      "desc-ws-" + string(rune('0'+i)),
			Slug:      "desc-eco.desc-dom.desc-app." + string(rune('0'+i)),
			ImageName: "ubuntu:22.04",
			Status:    "stopped",
			GitRepoID: sql.NullInt64{Int64: int64(repo.ID), Valid: true},
		}
		require.NoError(t, store.CreateWorkspace(ws))
	}

	return store, repo
}

// =============================================================================
// dvm describe gitrepo <name> tests
// =============================================================================

func TestDescribeGitRepo_ShowsMirrorStatus(t *testing.T) {
	store, repo := setupDescribeTestStore(t)

	inspector := &mockMirrorInspector{
		branches:  []mirror.RefInfo{{Name: "main"}, {Name: "feature-x"}},
		diskUsage: 148 * 1024 * 1024, // 148 MB
		verifyErr: nil,
	}

	cmd := newTestDescribeGitRepoCmd()
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, store)
	ctx = context.WithValue(ctx, CtxKeyMirrorManager, mirror.MirrorManager(inspector))
	cmd.SetContext(ctx)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{repo.Name})

	err := cmd.Execute()
	require.NoError(t, err, "describe gitrepo should succeed")

	output := buf.String()
	// Output must include mirror health, disk usage, branch count
	assert.Contains(t, output, repo.Name, "output should contain repo name")
	assert.Contains(t, output, repo.URL, "output should contain URL")
	assert.Contains(t, output, "Disk", "output should include disk usage field")
	assert.Contains(t, output, "Branch", "output should include branch count field")
}

func TestDescribeGitRepo_ShowsLinkedResources(t *testing.T) {
	store, repo := setupDescribeTestStore(t)

	inspector := &mockMirrorInspector{diskUsage: 1024}

	cmd := newTestDescribeGitRepoCmd()
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, store)
	ctx = context.WithValue(ctx, CtxKeyMirrorManager, mirror.MirrorManager(inspector))
	cmd.SetContext(ctx)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetArgs([]string{repo.Name})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should list linked apps and workspaces
	assert.Contains(t, output, "desc-app-1", "output should list linked app")
	assert.Contains(t, output, "desc-app-2", "output should list linked app")
}

func TestDescribeGitRepo_NotFound(t *testing.T) {
	store := db.NewMockDataStore()

	inspector := &mockMirrorInspector{}

	cmd := newTestDescribeGitRepoCmd()
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, store)
	ctx = context.WithValue(ctx, CtxKeyMirrorManager, mirror.MirrorManager(inspector))
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"nonexistent-repo"})

	err := cmd.Execute()
	assert.Error(t, err, "describe gitrepo should return error for non-existent repo")
}

// =============================================================================
// dvm get branches --repo <name> tests
// =============================================================================

func TestGetBranches_ListsBranches(t *testing.T) {
	store := db.NewMockDataStore()

	repo := &models.GitRepoDB{
		Name:       "branch-repo",
		URL:        "https://github.com/org/branch-repo.git",
		Slug:       "github.com_org_branch-repo",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "synced",
	}
	require.NoError(t, store.CreateGitRepo(repo))

	inspector := &mockMirrorInspector{
		branches: []mirror.RefInfo{
			{Name: "main"},
			{Name: "feature-login"},
			{Name: "bugfix-crash"},
		},
	}

	cmd := newTestGetBranchesCmd()
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, store)
	ctx = context.WithValue(ctx, CtxKeyMirrorManager, mirror.MirrorManager(inspector))
	cmd.SetContext(ctx)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	require.NoError(t, cmd.Flags().Set("repo", "branch-repo"))

	err := cmd.Execute()
	require.NoError(t, err, "get branches should succeed")

	output := buf.String()
	assert.Contains(t, output, "main", "output should contain branch 'main'")
	assert.Contains(t, output, "feature-login", "output should contain branch 'feature-login'")
	assert.Contains(t, output, "bugfix-crash", "output should contain branch 'bugfix-crash'")
}

func TestGetBranches_RepoRequired(t *testing.T) {
	store := db.NewMockDataStore()
	inspector := &mockMirrorInspector{}

	cmd := newTestGetBranchesCmd()
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, store)
	ctx = context.WithValue(ctx, CtxKeyMirrorManager, mirror.MirrorManager(inspector))
	cmd.SetContext(ctx)

	// Don't set --repo flag
	err := cmd.Execute()
	assert.Error(t, err, "get branches without --repo should return error")
}

// =============================================================================
// dvm get tags --repo <name> tests
// =============================================================================

func TestGetTags_ListsTags(t *testing.T) {
	store := db.NewMockDataStore()

	repo := &models.GitRepoDB{
		Name:       "tag-repo",
		URL:        "https://github.com/org/tag-repo.git",
		Slug:       "github.com_org_tag-repo",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "synced",
	}
	require.NoError(t, store.CreateGitRepo(repo))

	inspector := &mockMirrorInspector{
		tags: []mirror.RefInfo{
			{Name: "v1.0.0"},
			{Name: "v1.1.0"},
			{Name: "v2.0.0"},
		},
	}

	cmd := newTestGetTagsCmd()
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, store)
	ctx = context.WithValue(ctx, CtxKeyMirrorManager, mirror.MirrorManager(inspector))
	cmd.SetContext(ctx)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	require.NoError(t, cmd.Flags().Set("repo", "tag-repo"))

	err := cmd.Execute()
	require.NoError(t, err, "get tags should succeed")

	output := buf.String()
	assert.Contains(t, output, "v1.0.0", "output should contain tag 'v1.0.0'")
	assert.Contains(t, output, "v1.1.0", "output should contain tag 'v1.1.0'")
	assert.Contains(t, output, "v2.0.0", "output should contain tag 'v2.0.0'")
}

// =============================================================================
// dvm edit gitrepo <name> tests
// =============================================================================

func TestEditGitRepo_UpdatesFields(t *testing.T) {
	store := db.NewMockDataStore()

	repo := &models.GitRepoDB{
		Name:       "editable-repo",
		URL:        "https://github.com/org/old-url.git",
		Slug:       "github.com_org_old-url",
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "synced",
	}
	require.NoError(t, store.CreateGitRepo(repo))

	// The edit command calls runEditGitRepo which opens $EDITOR.
	// We test the underlying updateGitRepoFields helper directly.
	// This helper applies field patches without opening an editor.
	newURL := "https://github.com/org/new-url.git"
	newRef := "develop"
	newAuthType := "token"

	err := updateGitRepoFields(store, repo.Name, newURL, newRef, newAuthType, "")
	require.NoError(t, err, "updateGitRepoFields should succeed")

	updated, err := store.GetGitRepoByName(repo.Name)
	require.NoError(t, err)
	assert.Equal(t, newURL, updated.URL, "URL should be updated")
	assert.Equal(t, newRef, updated.DefaultRef, "DefaultRef should be updated")
	assert.Equal(t, newAuthType, updated.AuthType, "AuthType should be updated")
}
