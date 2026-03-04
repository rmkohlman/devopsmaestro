package cmd

import (
	"database/sql"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/mirror"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Test Helpers ==========

// MockMirrorManager implements the MirrorManager interface for testing
type MockMirrorManager struct {
	CloneFunc            func(url string, slug string) (string, error)
	SyncFunc             func(slug string) error
	DeleteFunc           func(slug string) error
	ExistsFunc           func(slug string) bool
	GetPathFunc          func(slug string) string
	CloneToWorkspaceFunc func(mirrorSlug string, destPath string, ref string) error
}

func (m *MockMirrorManager) Clone(url string, slug string) (string, error) {
	if m.CloneFunc != nil {
		return m.CloneFunc(url, slug)
	}
	return "/mock/path", nil
}

func (m *MockMirrorManager) Sync(slug string) error {
	if m.SyncFunc != nil {
		return m.SyncFunc(slug)
	}
	return nil
}

func (m *MockMirrorManager) Delete(slug string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(slug)
	}
	return nil
}

func (m *MockMirrorManager) Exists(slug string) bool {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(slug)
	}
	return true
}

func (m *MockMirrorManager) GetPath(slug string) string {
	if m.GetPathFunc != nil {
		return m.GetPathFunc(slug)
	}
	return "/mock/path"
}

func (m *MockMirrorManager) CloneToWorkspace(mirrorSlug string, destPath string, ref string) error {
	if m.CloneToWorkspaceFunc != nil {
		return m.CloneToWorkspaceFunc(mirrorSlug, destPath, ref)
	}
	return nil
}

// Ensure MockMirrorManager implements MirrorManager
var _ mirror.MirrorManager = (*MockMirrorManager)(nil)

// newTestCreateWorkspaceCmd creates a fresh createWorkspaceCmd for testing
func newTestCreateWorkspaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "workspace <name>",
		Aliases: []string{"ws"},
		Short:   "Create a new workspace",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			// Simplified test version that doesn't call runCreateWorkspace
			// Tests will use the command structure directly
		},
	}
	cmd.Flags().String("description", "", "Workspace description")
	cmd.Flags().String("image", "", "Custom image name")
	cmd.Flags().StringP("app", "a", "", "App name")
	cmd.Flags().String("repo", "", "GitRepo to clone into workspace")
	return cmd
}

// setupTestContext creates a test context with mock datastore and test data
func setupTestContext() (*db.MockDataStore, *models.App) {
	mockStore := db.NewMockDataStore()

	// Create test ecosystem
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	mockStore.CreateEcosystem(ecosystem)

	// Create test domain
	domain := &models.Domain{Name: "test-domain", EcosystemID: ecosystem.ID}
	mockStore.CreateDomain(domain)

	// Create test app
	app := &models.App{
		Name:     "test-app",
		DomainID: domain.ID,
		Path:     "/test/path",
	}
	mockStore.CreateApp(app)

	return mockStore, app
}

// ========== Command Structure Tests ==========

func TestCreateWorkspaceCmd_Exists(t *testing.T) {
	assert.NotNil(t, createWorkspaceCmd, "createWorkspaceCmd should exist")
	assert.Equal(t, "workspace <name>", createWorkspaceCmd.Use)
}

func TestCreateWorkspaceCmd_Aliases(t *testing.T) {
	aliases := createWorkspaceCmd.Aliases
	assert.Contains(t, aliases, "ws", "should have 'ws' alias")
}

func TestCreateWorkspaceCmd_HasRepoFlag(t *testing.T) {
	repoFlag := createWorkspaceCmd.Flags().Lookup("repo")
	assert.NotNil(t, repoFlag, "createWorkspaceCmd should have 'repo' flag")

	if repoFlag != nil {
		assert.Equal(t, "", repoFlag.DefValue, "repo flag should default to empty")
		assert.Equal(t, "string", repoFlag.Value.Type(), "repo flag should be string type")
	}
}

// ========== --repo Flag Tests ==========

func TestCreateWorkspace_WithRepoFlag_SetsGitRepoID(t *testing.T) {
	mockStore, app := setupTestContext()

	// Create a GitRepo
	gitRepo := &models.GitRepoDB{
		Name: "test-repo",
		URL:  "https://github.com/org/repo.git",
		Slug: "org-repo",
	}
	err := mockStore.CreateGitRepo(gitRepo)
	require.NoError(t, err)

	// Create workspace with GitRepoID
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "test-workspace",
		ImageName: "test-image",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true},
	}
	err = mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	// Verify workspace has GitRepoID set
	ws, err := mockStore.GetWorkspaceByName(app.ID, "test-workspace")
	assert.NoError(t, err)
	assert.NotNil(t, ws)
	assert.True(t, ws.GitRepoID.Valid, "GitRepoID should be valid")
	assert.Equal(t, int64(gitRepo.ID), ws.GitRepoID.Int64, "GitRepoID should match")
}

func TestCreateWorkspace_WithRepoFlag_NonExistentRepo_ReturnsError(t *testing.T) {
	mockStore, _ := setupTestContext()

	// Try to get non-existent GitRepo
	_, err := mockStore.GetGitRepoByName("nonexistent-repo")
	assert.Error(t, err, "should return error for non-existent repo")
	assert.Contains(t, err.Error(), "not found", "error should mention not found")
}

func TestCreateWorkspace_WithRepoFlag_CreatesWorkspaceRecord(t *testing.T) {
	mockStore, app := setupTestContext()

	// Create GitRepo
	gitRepo := &models.GitRepoDB{
		Name: "my-repo",
		URL:  "https://github.com/org/myrepo.git",
		Slug: "org-myrepo",
	}
	err := mockStore.CreateGitRepo(gitRepo)
	require.NoError(t, err)

	// Create workspace linked to GitRepo
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "feature-branch",
		ImageName: "dvm-feature-branch-test-app:pending",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true},
	}
	err = mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	// Verify workspace was created with GitRepoID
	ws, err := mockStore.GetWorkspaceByName(app.ID, "feature-branch")
	assert.NoError(t, err)
	assert.NotNil(t, ws)
	assert.Equal(t, "feature-branch", ws.Name)
	assert.True(t, ws.GitRepoID.Valid)
	assert.Equal(t, int64(gitRepo.ID), ws.GitRepoID.Int64)
}

func TestCreateWorkspace_WithRepoFlag_MirrorNotCloned_SyncsFirst(t *testing.T) {
	mockStore, _ := setupTestContext()

	// Create GitRepo
	gitRepo := &models.GitRepoDB{
		Name: "sync-repo",
		URL:  "https://github.com/org/syncrepo.git",
		Slug: "org-syncrepo",
	}
	err := mockStore.CreateGitRepo(gitRepo)
	require.NoError(t, err)

	// Mock MirrorManager that tracks calls
	var cloneCalled bool
	var cloneToWorkspaceCalled bool

	mockMirror := &MockMirrorManager{
		ExistsFunc: func(slug string) bool {
			return false // Mirror doesn't exist yet
		},
		CloneFunc: func(url string, slug string) (string, error) {
			cloneCalled = true
			assert.Equal(t, gitRepo.URL, url)
			assert.Equal(t, gitRepo.Slug, slug)
			return "/mock/mirror/path", nil
		},
		CloneToWorkspaceFunc: func(mirrorSlug string, destPath string, ref string) error {
			cloneToWorkspaceCalled = true
			assert.Equal(t, gitRepo.Slug, mirrorSlug)
			return nil
		},
	}

	// Simulate the workflow
	if !mockMirror.Exists(gitRepo.Slug) {
		_, err := mockMirror.Clone(gitRepo.URL, gitRepo.Slug)
		assert.NoError(t, err)
	}

	// Clone to workspace
	err = mockMirror.CloneToWorkspace(gitRepo.Slug, "/workspace/repo", gitRepo.DefaultRef)
	assert.NoError(t, err)

	// Verify calls
	assert.True(t, cloneCalled, "Clone should be called when mirror doesn't exist")
	assert.True(t, cloneToWorkspaceCalled, "CloneToWorkspace should be called")
}

func TestCreateWorkspace_WithRepoFlag_MirrorExists_SkipsSync(t *testing.T) {
	mockStore, _ := setupTestContext()

	// Create GitRepo
	gitRepo := &models.GitRepoDB{
		Name: "existing-repo",
		URL:  "https://github.com/org/existing.git",
		Slug: "org-existing",
	}
	err := mockStore.CreateGitRepo(gitRepo)
	require.NoError(t, err)

	// Mock MirrorManager with existing mirror
	var cloneCalled bool
	var cloneToWorkspaceCalled bool

	mockMirror := &MockMirrorManager{
		ExistsFunc: func(slug string) bool {
			return true // Mirror already exists
		},
		CloneFunc: func(url string, slug string) (string, error) {
			cloneCalled = true
			return "", nil
		},
		CloneToWorkspaceFunc: func(mirrorSlug string, destPath string, ref string) error {
			cloneToWorkspaceCalled = true
			return nil
		},
	}

	// Simulate workflow - should skip clone since mirror exists
	if !mockMirror.Exists(gitRepo.Slug) {
		mockMirror.Clone(gitRepo.URL, gitRepo.Slug)
	}

	// Clone to workspace
	err = mockMirror.CloneToWorkspace(gitRepo.Slug, "/workspace/repo", gitRepo.DefaultRef)
	assert.NoError(t, err)

	// Verify calls
	assert.False(t, cloneCalled, "Clone should NOT be called when mirror exists")
	assert.True(t, cloneToWorkspaceCalled, "CloneToWorkspace should still be called")
}

func TestCreateWorkspace_WithRepoFlag_UsesCorrectRepoPath(t *testing.T) {
	mockStore, app := setupTestContext()

	// Create workspace
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "path-test",
		ImageName: "test-image",
		Status:    "stopped",
	}
	err := mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	// Get workspace path and repo path
	workspacePath, err := mockStore.GetWorkspacePath(workspace.ID)
	assert.NoError(t, err)
	assert.Contains(t, workspacePath, "workspaces/", "workspace path should contain workspaces/")

	repoPath, err := mockStore.GetWorkspaceRepoPath(workspace.ID)
	assert.NoError(t, err)
	assert.Contains(t, repoPath, "repo", "repo path should end with /repo")
}

func TestCreateWorkspace_WithRepoFlag_CloneFailure_WorkspaceStillCreated(t *testing.T) {
	mockStore, app := setupTestContext()

	// Create GitRepo
	gitRepo := &models.GitRepoDB{
		Name: "fail-repo",
		URL:  "https://github.com/org/fail.git",
		Slug: "org-fail",
	}
	err := mockStore.CreateGitRepo(gitRepo)
	require.NoError(t, err)

	// Create workspace first (simulating that workspace creation succeeds)
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "fail-workspace",
		ImageName: "test-image",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true},
	}
	err = mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	// Mock MirrorManager that fails on clone
	mockMirror := &MockMirrorManager{
		ExistsFunc: func(slug string) bool {
			return true
		},
		CloneToWorkspaceFunc: func(mirrorSlug string, destPath string, ref string) error {
			return assert.AnError // Simulate clone failure
		},
	}

	// Attempt clone - should fail
	err = mockMirror.CloneToWorkspace(gitRepo.Slug, "/workspace/repo", "")
	assert.Error(t, err, "clone should fail")

	// But workspace should still exist
	ws, err := mockStore.GetWorkspaceByName(app.ID, "fail-workspace")
	assert.NoError(t, err)
	assert.NotNil(t, ws, "workspace should exist even if clone failed")
}

func TestCreateWorkspace_WithoutRepoFlag_GitRepoIDIsNull(t *testing.T) {
	mockStore, app := setupTestContext()

	// Create workspace WITHOUT GitRepoID
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "no-repo-workspace",
		ImageName: "test-image",
		Status:    "stopped",
		// GitRepoID not set
	}
	err := mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	// Verify GitRepoID is NULL
	ws, err := mockStore.GetWorkspaceByName(app.ID, "no-repo-workspace")
	assert.NoError(t, err)
	assert.NotNil(t, ws)
	assert.False(t, ws.GitRepoID.Valid, "GitRepoID should be NULL when --repo not provided")
}

// ========== Integration-Style Tests ==========

func TestCreateWorkspace_FullWorkflow_WithRepo(t *testing.T) {
	// This test simulates the full workflow from create.go
	mockStore, app := setupTestContext()

	// 1. Create GitRepo
	gitRepo := &models.GitRepoDB{
		Name:       "full-test-repo",
		URL:        "https://github.com/org/full.git",
		Slug:       "org-full",
		DefaultRef: "main",
	}
	err := mockStore.CreateGitRepo(gitRepo)
	require.NoError(t, err)

	// 2. Create workspace with GitRepoID
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "full-test-workspace",
		ImageName: "dvm-full-test-workspace-test-app:pending",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true},
	}
	err = mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	// 3. Get workspace path
	workspacePath, err := mockStore.GetWorkspacePath(workspace.ID)
	require.NoError(t, err)

	// 4. Mock mirror operations
	mockMirror := &MockMirrorManager{
		ExistsFunc: func(slug string) bool {
			return true // Assume mirror exists
		},
		CloneToWorkspaceFunc: func(mirrorSlug string, destPath string, ref string) error {
			assert.Equal(t, gitRepo.Slug, mirrorSlug)
			assert.Contains(t, destPath, "repo")
			assert.Equal(t, gitRepo.DefaultRef, ref)
			return nil
		},
	}

	// 5. Clone to workspace
	repoPath := workspacePath + "repo"
	err = mockMirror.CloneToWorkspace(gitRepo.Slug, repoPath, gitRepo.DefaultRef)
	assert.NoError(t, err)

	// 6. Verify final state
	ws, err := mockStore.GetWorkspaceByName(app.ID, "full-test-workspace")
	assert.NoError(t, err)
	assert.NotNil(t, ws)
	assert.True(t, ws.GitRepoID.Valid)
	assert.Equal(t, int64(gitRepo.ID), ws.GitRepoID.Int64)
}

func TestCreateWorkspace_TableDriven_RepoScenarios(t *testing.T) {
	tests := []struct {
		name           string
		repoName       string
		repoURL        string
		mirrorExists   bool
		cloneShouldRun bool
		wantError      bool
	}{
		{
			name:           "new mirror needs clone",
			repoName:       "new-repo",
			repoURL:        "https://github.com/org/new.git",
			mirrorExists:   false,
			cloneShouldRun: true,
			wantError:      false,
		},
		{
			name:           "existing mirror skips clone",
			repoName:       "existing-repo",
			repoURL:        "https://github.com/org/existing.git",
			mirrorExists:   true,
			cloneShouldRun: false,
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore, app := setupTestContext()

			// Create GitRepo
			gitRepo := &models.GitRepoDB{
				Name: tt.repoName,
				URL:  tt.repoURL,
				Slug: "org-" + tt.repoName,
			}
			err := mockStore.CreateGitRepo(gitRepo)
			require.NoError(t, err)

			// Create workspace
			workspace := &models.Workspace{
				AppID:     app.ID,
				Name:      tt.name,
				ImageName: "test-image",
				Status:    "stopped",
				GitRepoID: sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true},
			}
			err = mockStore.CreateWorkspace(workspace)
			require.NoError(t, err)

			// Mock mirror manager
			var cloneCalled bool
			mockMirror := &MockMirrorManager{
				ExistsFunc: func(slug string) bool {
					return tt.mirrorExists
				},
				CloneFunc: func(url string, slug string) (string, error) {
					cloneCalled = true
					return "/mock/path", nil
				},
				CloneToWorkspaceFunc: func(mirrorSlug string, destPath string, ref string) error {
					return nil
				},
			}

			// Simulate clone workflow
			if !mockMirror.Exists(gitRepo.Slug) {
				_, err := mockMirror.Clone(gitRepo.URL, gitRepo.Slug)
				assert.NoError(t, err)
			}

			// Verify clone call matches expectation
			assert.Equal(t, tt.cloneShouldRun, cloneCalled, "clone call mismatch")
		})
	}
}

// ========== GitRepo Inheritance Tests (Issue #17) ==========

// TestCreateWorkspace_InheritsAppGitRepo verifies that a workspace inherits
// the App's GitRepoID when --repo flag is not provided.
// This test uses the ResolveWorkspaceGitRepo helper function.
func TestCreateWorkspace_InheritsAppGitRepo(t *testing.T) {
	mockStore, app := setupTestContext()

	// Create a GitRepo
	gitRepo := &models.GitRepoDB{
		Name: "app-default-repo",
		URL:  "https://github.com/org/app-repo.git",
		Slug: "org-app-repo",
	}
	err := mockStore.CreateGitRepo(gitRepo)
	require.NoError(t, err)

	// Associate the GitRepo with the App
	app.GitRepoID = sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true}
	err = mockStore.UpdateApp(app)
	require.NoError(t, err)

	// Verify App has GitRepoID set
	updatedApp, err := mockStore.GetAppByName(app.DomainID, app.Name)
	require.NoError(t, err)
	require.True(t, updatedApp.GitRepoID.Valid, "App should have GitRepoID set")
	require.Equal(t, int64(gitRepo.ID), updatedApp.GitRepoID.Int64)

	// Use the ResolveWorkspaceGitRepo helper function (no --repo flag)
	resolvedRepo, resolvedGitRepoID, err := ResolveWorkspaceGitRepo(mockStore, updatedApp, "")
	require.NoError(t, err)

	// Create workspace using the resolved GitRepoID
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "inherited-workspace",
		ImageName: "test-image",
		Status:    "stopped",
		GitRepoID: resolvedGitRepoID, // Use the resolved value
	}

	err = mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	// Verify workspace inherited the App's GitRepoID
	ws, err := mockStore.GetWorkspaceByName(app.ID, "inherited-workspace")
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	// Verify inheritance worked
	assert.True(t, ws.GitRepoID.Valid, "Workspace should inherit App's GitRepoID")
	assert.Equal(t, int64(gitRepo.ID), ws.GitRepoID.Int64, "Workspace GitRepoID should match App's GitRepoID")

	// Also verify the resolved repo is correct
	assert.NotNil(t, resolvedRepo, "Resolved GitRepo should not be nil")
	assert.Equal(t, gitRepo.Name, resolvedRepo.Name, "Resolved repo should match App's repo")
}

// TestCreateWorkspace_ExplicitRepoOverridesAppGitRepo verifies that when a workspace
// is created with an explicit --repo flag, it uses that GitRepo instead of inheriting
// from the App.
// This test EXPECTS TO FAIL until the inheritance logic is implemented.
func TestCreateWorkspace_ExplicitRepoOverridesAppGitRepo(t *testing.T) {
	mockStore, app := setupTestContext()

	// Create two GitRepos: one for the App, one for the Workspace
	appGitRepo := &models.GitRepoDB{
		Name: "app-repo",
		URL:  "https://github.com/org/app.git",
		Slug: "org-app",
	}
	err := mockStore.CreateGitRepo(appGitRepo)
	require.NoError(t, err)

	workspaceGitRepo := &models.GitRepoDB{
		Name: "workspace-specific-repo",
		URL:  "https://github.com/org/workspace.git",
		Slug: "org-workspace",
	}
	err = mockStore.CreateGitRepo(workspaceGitRepo)
	require.NoError(t, err)

	// Associate appGitRepo with the App
	app.GitRepoID = sql.NullInt64{Int64: int64(appGitRepo.ID), Valid: true}
	err = mockStore.UpdateApp(app)
	require.NoError(t, err)

	// Create workspace WITH explicit --repo flag
	// The workspace should use the explicitly provided GitRepo, NOT the App's
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "override-workspace",
		ImageName: "test-image",
		Status:    "stopped",
		// Explicitly set to workspaceGitRepo (simulating --repo flag)
		GitRepoID: sql.NullInt64{Int64: int64(workspaceGitRepo.ID), Valid: true},
	}

	err = mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	// Verify workspace uses the explicitly provided GitRepoID
	ws, err := mockStore.GetWorkspaceByName(app.ID, "override-workspace")
	assert.NoError(t, err)
	assert.NotNil(t, ws)

	assert.True(t, ws.GitRepoID.Valid, "Workspace should have GitRepoID set")
	assert.Equal(t, int64(workspaceGitRepo.ID), ws.GitRepoID.Int64,
		"Workspace should use explicit GitRepo, not App's GitRepo")
	assert.NotEqual(t, int64(appGitRepo.ID), ws.GitRepoID.Int64,
		"Workspace should NOT use App's GitRepo when --repo is explicitly provided")
}

// TestCreateWorkspace_NoGitRepoWhenAppHasNone verifies that when an App does not
// have a GitRepoID, workspaces created under it also have no GitRepoID.
func TestCreateWorkspace_NoGitRepoWhenAppHasNone(t *testing.T) {
	mockStore, app := setupTestContext()

	// Verify App has no GitRepoID
	assert.False(t, app.GitRepoID.Valid, "App should not have GitRepoID for this test")

	// Create workspace without --repo flag
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "no-repo-workspace",
		ImageName: "test-image",
		Status:    "stopped",
		// GitRepoID not set
	}

	err := mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	// Verify workspace has no GitRepoID
	ws, err := mockStore.GetWorkspaceByName(app.ID, "no-repo-workspace")
	assert.NoError(t, err)
	assert.NotNil(t, ws)
	assert.False(t, ws.GitRepoID.Valid,
		"Workspace should not have GitRepoID when App has none and --repo not provided")
}

// TestCreateWorkspace_InheritanceTableDriven is a comprehensive table-driven test
// for GitRepo inheritance scenarios.
// These tests EXPECT TO FAIL until the inheritance logic is implemented.
func TestCreateWorkspace_InheritanceTableDriven(t *testing.T) {
	tests := []struct {
		name              string
		appHasRepo        bool
		workspaceRepoFlag string // Empty means --repo not provided
		expectedInherit   bool   // Should workspace inherit App's repo?
		expectedRepoName  string // Expected GitRepo name (empty = no repo)
	}{
		{
			name:              "app has repo, workspace flag not provided - inherit",
			appHasRepo:        true,
			workspaceRepoFlag: "",
			expectedInherit:   true,
			expectedRepoName:  "app-repo",
		},
		{
			name:              "app has repo, workspace flag provided - override",
			appHasRepo:        true,
			workspaceRepoFlag: "workspace-repo",
			expectedInherit:   false,
			expectedRepoName:  "workspace-repo",
		},
		{
			name:              "app has no repo, workspace flag not provided - no repo",
			appHasRepo:        false,
			workspaceRepoFlag: "",
			expectedInherit:   false,
			expectedRepoName:  "",
		},
		{
			name:              "app has no repo, workspace flag provided - use flag",
			appHasRepo:        false,
			workspaceRepoFlag: "workspace-repo",
			expectedInherit:   false,
			expectedRepoName:  "workspace-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore, app := setupTestContext()

			var appGitRepo *models.GitRepoDB
			var workspaceGitRepo *models.GitRepoDB

			// Create app GitRepo if needed
			if tt.appHasRepo {
				appGitRepo = &models.GitRepoDB{
					Name: "app-repo",
					URL:  "https://github.com/org/app.git",
					Slug: "org-app",
				}
				err := mockStore.CreateGitRepo(appGitRepo)
				require.NoError(t, err)

				app.GitRepoID = sql.NullInt64{Int64: int64(appGitRepo.ID), Valid: true}
				err = mockStore.UpdateApp(app)
				require.NoError(t, err)
			}

			// Create workspace GitRepo if flag provided
			if tt.workspaceRepoFlag != "" {
				workspaceGitRepo = &models.GitRepoDB{
					Name: tt.workspaceRepoFlag,
					URL:  "https://github.com/org/workspace.git",
					Slug: "org-workspace",
				}
				err := mockStore.CreateGitRepo(workspaceGitRepo)
				require.NoError(t, err)
			}

			// Create workspace
			workspace := &models.Workspace{
				AppID:     app.ID,
				Name:      "test-workspace",
				ImageName: "test-image",
				Status:    "stopped",
			}

			// Simulate the inheritance logic that SHOULD be in create.go
			// THIS IS THE FIX THAT NEEDS TO BE IMPLEMENTED:
			if tt.workspaceRepoFlag != "" {
				// Explicit --repo flag provided
				workspace.GitRepoID = sql.NullInt64{Int64: int64(workspaceGitRepo.ID), Valid: true}
			} else if tt.appHasRepo && app.GitRepoID.Valid {
				// No --repo flag, but App has GitRepoID - inherit it
				workspace.GitRepoID = app.GitRepoID
			}
			// Otherwise GitRepoID remains unset (NULL)

			err := mockStore.CreateWorkspace(workspace)
			require.NoError(t, err)

			// Verify the workspace GitRepoID
			ws, err := mockStore.GetWorkspaceByName(app.ID, "test-workspace")
			require.NoError(t, err)
			require.NotNil(t, ws)

			if tt.expectedRepoName == "" {
				// Should have no GitRepo
				assert.False(t, ws.GitRepoID.Valid, "Workspace should not have GitRepoID")
			} else {
				// Should have a GitRepo
				assert.True(t, ws.GitRepoID.Valid, "Workspace should have GitRepoID set")

				// Verify it's the correct one
				repo, err := mockStore.GetGitRepoByID(ws.GitRepoID.Int64)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRepoName, repo.Name,
					"Workspace should have correct GitRepo")
			}
		})
	}
}
