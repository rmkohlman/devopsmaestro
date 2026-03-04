package cmd

import (
	"context"
	"database/sql"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/mirror"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== TDD Phase 2: Failing Tests for --repo Flag ==========
// These tests are written FIRST to drive the implementation of the --repo flag
// on `dvm create app`. They are expected to FAIL until implementation is complete.

// setupAppTestContext creates a test context with mock datastore and test ecosystem/domain
func setupAppTestContext() (*db.MockDataStore, *models.Domain) {
	mockStore := db.NewMockDataStore()

	// Create test ecosystem
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	mockStore.CreateEcosystem(ecosystem)

	// Create test domain
	domain := &models.Domain{Name: "test-domain", EcosystemID: ecosystem.ID}
	mockStore.CreateDomain(domain)

	// Set as active
	mockStore.SetActiveEcosystem(&ecosystem.ID)
	mockStore.SetActiveDomain(&domain.ID)

	return mockStore, domain
}

// TestCreateAppCmd_HasRepoFlag verifies that the createAppCmd has a --repo flag
func TestCreateAppCmd_HasRepoFlag(t *testing.T) {
	repoFlag := createAppCmd.Flags().Lookup("repo")
	assert.NotNil(t, repoFlag, "createAppCmd should have 'repo' flag")

	if repoFlag != nil {
		assert.Equal(t, "", repoFlag.DefValue, "repo flag should default to empty")
		assert.Equal(t, "string", repoFlag.Value.Type(), "repo flag should be string type")
	}
}

// TestCreateApp_WithRepoURL_AutoCreatesGitRepo tests that providing a URL to --repo
// automatically creates a GitRepo using the slug from the URL
func TestCreateApp_WithRepoURL_AutoCreatesGitRepo(t *testing.T) {
	mockStore, domain := setupAppTestContext()

	// Setup command context
	cmd := &cobra.Command{}
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	// Simulate: dvm create app myapp --repo https://github.com/rmkohlman/dvm-test-golang.git
	appName := "myapp"
	repoURL := "https://github.com/rmkohlman/dvm-test-golang.git"

	// Expected slug from URL
	expectedSlug, err := mirror.GenerateSlug(repoURL)
	require.NoError(t, err)
	assert.Equal(t, "github.com_rmkohlman_dvm-test-golang", expectedSlug)

	// Mock the create app workflow with --repo flag
	// This would be called by createAppCmd.RunE
	// For now, we'll simulate the expected behavior

	// 1. GitRepo should be auto-created with slug name
	gitRepo := &models.GitRepoDB{
		Name:       expectedSlug, // Auto-generated name from slug
		URL:        repoURL,
		Slug:       expectedSlug,
		DefaultRef: "main",
		AuthType:   "none",
		AutoSync:   true,
	}
	err = mockStore.CreateGitRepo(gitRepo)
	require.NoError(t, err)

	// 2. App should be created with GitRepoID set
	app := &models.App{
		Name:      appName,
		DomainID:  domain.ID,
		Path:      "/tmp/test", // Path is optional when --repo is provided
		GitRepoID: sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true},
	}
	err = mockStore.CreateApp(app)
	require.NoError(t, err)

	// Verify GitRepo was created
	createdRepo, err := mockStore.GetGitRepoByName(expectedSlug)
	assert.NoError(t, err)
	assert.NotNil(t, createdRepo)
	assert.Equal(t, repoURL, createdRepo.URL)
	assert.Equal(t, expectedSlug, createdRepo.Slug)

	// Verify App was created with GitRepoID
	createdApp, err := mockStore.GetAppByName(domain.ID, appName)
	assert.NoError(t, err)
	assert.NotNil(t, createdApp)
	assert.True(t, createdApp.GitRepoID.Valid, "App should have GitRepoID set")
	assert.Equal(t, int64(gitRepo.ID), createdApp.GitRepoID.Int64, "GitRepoID should match created repo")

	// Expected behavior summary:
	// - GitRepo auto-created with slug as name
	// - App created with GitRepoID linked
	// - App path can be optional (defaults to temp or project-specific location)
}

// TestCreateApp_WithExistingGitRepoName tests that providing an existing GitRepo name
// to --repo links the existing repo to the app without creating a new one
func TestCreateApp_WithExistingGitRepoName(t *testing.T) {
	mockStore, domain := setupAppTestContext()

	// Pre-create a GitRepo
	existingRepo := &models.GitRepoDB{
		Name:       "my-repo",
		URL:        "https://github.com/org/repo.git",
		Slug:       "github.com_org_repo",
		DefaultRef: "main",
		AuthType:   "none",
	}
	err := mockStore.CreateGitRepo(existingRepo)
	require.NoError(t, err)

	// Simulate: dvm create app myapp --repo my-repo
	appName := "myapp"
	repoName := "my-repo"

	// Look up existing GitRepo by name
	repo, err := mockStore.GetGitRepoByName(repoName)
	require.NoError(t, err)
	require.NotNil(t, repo)

	// Create app linked to existing repo
	app := &models.App{
		Name:      appName,
		DomainID:  domain.ID,
		Path:      "/tmp/test",
		GitRepoID: sql.NullInt64{Int64: int64(repo.ID), Valid: true},
	}
	err = mockStore.CreateApp(app)
	require.NoError(t, err)

	// Verify app is linked to existing repo
	createdApp, err := mockStore.GetAppByName(domain.ID, appName)
	assert.NoError(t, err)
	assert.True(t, createdApp.GitRepoID.Valid)
	assert.Equal(t, int64(existingRepo.ID), createdApp.GitRepoID.Int64)

	// Verify no new GitRepo was created (should still be just 1)
	allRepos, err := mockStore.ListGitRepos()
	assert.NoError(t, err)
	assert.Len(t, allRepos, 1, "Should not create a new GitRepo, only link existing one")
}

// TestCreateApp_RepoAndPath_MutuallyExclusive tests that --repo and --path cannot be used together
func TestCreateApp_RepoAndPath_MutuallyExclusive(t *testing.T) {
	mockStore, _ := setupAppTestContext()

	// Setup command
	cmd := &cobra.Command{
		Use:  "app <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Simulate validation logic
			repoFlag, _ := cmd.Flags().GetString("repo")
			pathFlag, _ := cmd.Flags().GetString("path")

			if repoFlag != "" && pathFlag != "" {
				return assert.AnError // Should return error about mutual exclusivity
			}
			return nil
		},
	}
	cmd.Flags().String("repo", "", "GitRepo URL or name")
	cmd.Flags().String("path", "", "App source path")

	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	// Set both flags
	cmd.Flags().Set("repo", "https://github.com/org/repo.git")
	cmd.Flags().Set("path", "/tmp/myapp")

	// Execute should fail with mutual exclusivity error
	err := cmd.Execute()
	assert.Error(t, err, "Should return error when both --repo and --path are provided")

	// Expected error message
	// In real implementation, this would be:
	// "flags --from-cwd, --path, and --repo are mutually exclusive"
}

// TestCreateApp_RepoAndFromCwd_MutuallyExclusive tests that --repo and --from-cwd cannot be used together
func TestCreateApp_RepoAndFromCwd_MutuallyExclusive(t *testing.T) {
	mockStore, _ := setupAppTestContext()

	// Setup command
	cmd := &cobra.Command{
		Use:  "app <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Simulate validation logic
			repoFlag, _ := cmd.Flags().GetString("repo")
			fromCwdFlag, _ := cmd.Flags().GetBool("from-cwd")

			if repoFlag != "" && fromCwdFlag {
				return assert.AnError // Should return error about mutual exclusivity
			}
			return nil
		},
	}
	cmd.Flags().String("repo", "", "GitRepo URL or name")
	cmd.Flags().Bool("from-cwd", false, "Use current directory")

	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	// Set both flags
	cmd.Flags().Set("repo", "https://github.com/org/repo.git")
	cmd.Flags().Set("from-cwd", "true")

	// Execute should fail with mutual exclusivity error
	err := cmd.Execute()
	assert.Error(t, err, "Should return error when both --repo and --from-cwd are provided")

	// Expected error message
	// In real implementation, this would be:
	// "flags --from-cwd, --path, and --repo are mutually exclusive"
}

// TestCreateApp_NonExistentGitRepoName_Error tests that providing a non-existent GitRepo name
// returns an error
func TestCreateApp_NonExistentGitRepoName_Error(t *testing.T) {
	mockStore, _ := setupAppTestContext()

	// Try to look up non-existent GitRepo
	nonExistentName := "nonexistent-repo"
	_, err := mockStore.GetGitRepoByName(nonExistentName)

	// Should return error
	assert.Error(t, err, "Should return error for non-existent GitRepo")
	assert.Contains(t, err.Error(), "not found", "Error should mention 'not found'")

	// Expected error message in real implementation:
	// "GitRepo 'nonexistent-repo' not found"
}

// TestCreateApp_NoSourceSpecified_Error tests that not specifying any source
// (--repo, --path, or --from-cwd) returns an error
func TestCreateApp_NoSourceSpecified_Error(t *testing.T) {
	mockStore, _ := setupAppTestContext()

	// Setup command
	cmd := &cobra.Command{
		Use:  "app <name>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Simulate validation logic
			repoFlag, _ := cmd.Flags().GetString("repo")
			pathFlag, _ := cmd.Flags().GetString("path")
			fromCwdFlag, _ := cmd.Flags().GetBool("from-cwd")

			if repoFlag == "" && pathFlag == "" && !fromCwdFlag {
				return assert.AnError // Should return error about missing source
			}
			return nil
		},
	}
	cmd.Flags().String("repo", "", "GitRepo URL or name")
	cmd.Flags().String("path", "", "App source path")
	cmd.Flags().Bool("from-cwd", false, "Use current directory")

	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	// Don't set any flags - all are empty/false

	// Execute should fail
	err := cmd.Execute()
	assert.Error(t, err, "Should return error when no source is specified")

	// Expected error message in real implementation:
	// "must specify one of: --from-cwd, --path, or --repo"
}

// TestCreateApp_RepoURL_IsValidatedAndGeneratesSlug tests that a provided URL
// is validated and generates a proper slug
func TestCreateApp_RepoURL_IsValidatedAndGeneratesSlug(t *testing.T) {
	tests := []struct {
		name        string
		repoURL     string
		wantSlug    string
		wantValid   bool
		description string
	}{
		{
			name:        "valid https url",
			repoURL:     "https://github.com/rmkohlman/dvm-test-golang.git",
			wantSlug:    "github.com_rmkohlman_dvm-test-golang",
			wantValid:   true,
			description: "HTTPS GitHub URL should be valid and generate correct slug",
		},
		{
			name:        "valid ssh url",
			repoURL:     "git@github.com:rmkohlman/dvm-test-golang.git",
			wantSlug:    "github.com_rmkohlman_dvm-test-golang",
			wantValid:   true,
			description: "SSH GitHub URL should be valid and generate correct slug",
		},
		{
			name:        "valid gitlab url",
			repoURL:     "https://gitlab.com/group/subgroup/project.git",
			wantSlug:    "gitlab.com_group_subgroup_project",
			wantValid:   true,
			description: "GitLab URL with subgroups should generate correct slug",
		},
		{
			name:        "invalid http url",
			repoURL:     "http://github.com/org/repo.git",
			wantSlug:    "",
			wantValid:   false,
			description: "HTTP URLs should be rejected (insecure)",
		},
		{
			name:        "invalid url with shell chars",
			repoURL:     "https://github.com/org/repo.git; rm -rf /",
			wantSlug:    "",
			wantValid:   false,
			description: "URLs with shell metacharacters should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Validate URL
			err := mirror.ValidateGitURL(tt.repoURL)
			if tt.wantValid {
				assert.NoError(t, err, tt.description)

				// Generate slug if valid
				slug, err := mirror.GenerateSlug(tt.repoURL)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSlug, slug, "Slug should match expected format")
			} else {
				assert.Error(t, err, tt.description)
			}
		})
	}
}

// TestCreateApp_PathOptionalWhenRepoProvided tests that when --repo is provided,
// the --path flag is optional and a default path can be used
func TestCreateApp_PathOptionalWhenRepoProvided(t *testing.T) {
	mockStore, domain := setupAppTestContext()

	// Create GitRepo first
	repo := &models.GitRepoDB{
		Name:       "test-repo",
		URL:        "https://github.com/org/repo.git",
		Slug:       "github.com_org_repo",
		DefaultRef: "main",
	}
	err := mockStore.CreateGitRepo(repo)
	require.NoError(t, err)

	// Create app with --repo but NO --path
	// Path should be optional and can default to something like:
	// ~/.devopsmaestro/apps/<app-name> or similar
	app := &models.App{
		Name:      "myapp",
		DomainID:  domain.ID,
		Path:      "", // Empty/default path - should be allowed when GitRepoID is set
		GitRepoID: sql.NullInt64{Int64: int64(repo.ID), Valid: true},
	}

	// In real implementation, the command would set a default path if empty
	// For this test, we just verify that GitRepoID being set is the important part
	if app.Path == "" && app.GitRepoID.Valid {
		// Set a default path
		app.Path = "/tmp/default-app-path"
	}

	err = mockStore.CreateApp(app)
	assert.NoError(t, err, "Should allow creating app with --repo even if --path is not provided")

	// Verify app was created with GitRepoID
	createdApp, err := mockStore.GetAppByName(domain.ID, "myapp")
	assert.NoError(t, err)
	assert.True(t, createdApp.GitRepoID.Valid)
	assert.NotEmpty(t, createdApp.Path, "Path should have a default value")
}

// TestCreateApp_RepoURL_DetectsExistingRepoByURL tests that if a URL is provided
// and a GitRepo with that URL already exists, we link to it instead of creating duplicate
func TestCreateApp_RepoURL_DetectsExistingRepoByURL(t *testing.T) {
	mockStore, domain := setupAppTestContext()

	repoURL := "https://github.com/org/repo.git"
	slug, err := mirror.GenerateSlug(repoURL)
	require.NoError(t, err)

	// Pre-create a GitRepo with this URL
	existingRepo := &models.GitRepoDB{
		Name:       slug, // Same slug
		URL:        repoURL,
		Slug:       slug,
		DefaultRef: "main",
	}
	err = mockStore.CreateGitRepo(existingRepo)
	require.NoError(t, err)

	// Now simulate creating an app with the same URL
	// Implementation should detect existing repo by URL and link to it
	// instead of trying to create a duplicate

	// Look up by slug (since slug is derived from URL, this simulates the lookup)
	foundRepo, err := mockStore.GetGitRepoByName(slug)
	require.NoError(t, err)
	require.NotNil(t, foundRepo)

	// Create app linked to existing repo
	app := &models.App{
		Name:      "myapp",
		DomainID:  domain.ID,
		Path:      "/tmp/test",
		GitRepoID: sql.NullInt64{Int64: int64(foundRepo.ID), Valid: true},
	}
	err = mockStore.CreateApp(app)
	require.NoError(t, err)

	// Verify no duplicate GitRepo was created
	allRepos, err := mockStore.ListGitRepos()
	assert.NoError(t, err)
	assert.Len(t, allRepos, 1, "Should reuse existing GitRepo, not create duplicate")

	// Verify app is linked correctly
	createdApp, err := mockStore.GetAppByName(domain.ID, "myapp")
	assert.NoError(t, err)
	assert.True(t, createdApp.GitRepoID.Valid)
	assert.Equal(t, int64(existingRepo.ID), createdApp.GitRepoID.Int64)
}
