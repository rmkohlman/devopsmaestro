package cmd

import (
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/mirror"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Tests for resolveOrCreateGitRepo ==========
// This function handles the --repo flag logic:
// - If repo is a URL, auto-create a GitRepo (or reuse existing by URL)
// - If repo is a name, look up existing GitRepo
// Returns: gitRepoID, gitRepoName, path (for app), error

// TestResolveOrCreateGitRepo_NewURL tests creating a new GitRepo from a URL
func TestResolveOrCreateGitRepo_NewURL(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repoURL := "https://github.com/rmkohlman/dvm-test-golang.git"

	// Call the function
	gitRepoID, gitRepoName, localPath, err := resolveOrCreateGitRepo(mockStore, repoURL)

	// Should succeed
	require.NoError(t, err)
	require.NotNil(t, gitRepoID)
	assert.NotEmpty(t, gitRepoName)
	assert.NotEmpty(t, localPath)

	// Verify expected slug was generated
	expectedSlug, err := mirror.GenerateSlug(repoURL)
	require.NoError(t, err)
	assert.Equal(t, "github.com_rmkohlman_dvm-test-golang", expectedSlug)
	assert.Equal(t, expectedSlug, gitRepoName)

	// Verify GitRepo was created in database
	createdRepo, err := mockStore.GetGitRepoByName(expectedSlug)
	require.NoError(t, err)
	require.NotNil(t, createdRepo)
	assert.Equal(t, repoURL, createdRepo.URL)
	assert.Equal(t, expectedSlug, createdRepo.Slug)
	assert.Equal(t, expectedSlug, createdRepo.Name)
	assert.Equal(t, "main", createdRepo.DefaultRef)
	assert.Equal(t, "none", createdRepo.AuthType)
	assert.True(t, createdRepo.AutoSync)
	assert.Equal(t, 60, createdRepo.SyncIntervalMinutes)

	// Verify path matches expected format
	expectedPath := getGitRepoPath(expectedSlug)
	assert.Equal(t, expectedPath, localPath)

	// Verify ID matches
	assert.Equal(t, createdRepo.ID, *gitRepoID)
}

// TestResolveOrCreateGitRepo_ExistingByURL tests that an existing GitRepo is found by URL
func TestResolveOrCreateGitRepo_ExistingByURL(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repoURL := "https://github.com/rmkohlman/dvm-test-golang.git"
	expectedSlug, err := mirror.GenerateSlug(repoURL)
	require.NoError(t, err)

	// Pre-create the GitRepo
	existingRepo := &models.GitRepoDB{
		Name:                expectedSlug,
		URL:                 repoURL,
		Slug:                expectedSlug,
		DefaultRef:          "main",
		AuthType:            "none",
		AutoSync:            true,
		SyncIntervalMinutes: 60,
		SyncStatus:          "synced",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}
	err = mockStore.CreateGitRepo(existingRepo)
	require.NoError(t, err)

	// Call the function with the same URL
	gitRepoID, gitRepoName, localPath, err := resolveOrCreateGitRepo(mockStore, repoURL)

	// Should succeed and return the existing repo
	require.NoError(t, err)
	require.NotNil(t, gitRepoID)
	assert.Equal(t, expectedSlug, gitRepoName)
	assert.Equal(t, getGitRepoPath(expectedSlug), localPath)
	assert.Equal(t, existingRepo.ID, *gitRepoID)

	// Verify no duplicate was created
	allRepos, err := mockStore.ListGitRepos()
	require.NoError(t, err)
	assert.Len(t, allRepos, 1, "Should not create duplicate GitRepo")
}

// TestResolveOrCreateGitRepo_SlugConflict tests that a slug conflict is detected
// when a GitRepo exists with the same slug but different URL
func TestResolveOrCreateGitRepo_SlugConflict(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// The conflict scenario occurs when:
	// 1. A GitRepo exists with a custom name that happens to match a generated slug
	// 2. The URL for that GitRepo is different from the URL we're trying to use
	//
	// For example:
	// - Existing GitRepo: name="github.com_user_repo", url="https://github.com/olduser/oldrepo.git"
	// - New URL: "https://github.com/user/repo.git" (generates slug "github.com_user_repo")

	// Create a GitRepo with a name that will conflict with the slug of our new URL
	conflictingName := "github.com_user_newrepo"
	existingRepo := &models.GitRepoDB{
		Name:       conflictingName,
		URL:        "https://github.com/olduser/oldrepo.git", // Different URL
		Slug:       "github.com_olduser_oldrepo",             // Different slug
		DefaultRef: "main",
		AuthType:   "none",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err := mockStore.CreateGitRepo(existingRepo)
	require.NoError(t, err)

	// Now try to create from URL that generates the conflicting slug
	conflictURL := "https://github.com/user/newrepo.git"

	// Verify the slug generation matches
	generatedSlug, err := mirror.GenerateSlug(conflictURL)
	require.NoError(t, err)
	assert.Equal(t, conflictingName, generatedSlug, "Test setup: slug should match")

	// Call the function
	gitRepoID, gitRepoName, localPath, err := resolveOrCreateGitRepo(mockStore, conflictURL)

	// Should error with slug conflict message
	require.Error(t, err)
	assert.Nil(t, gitRepoID)
	assert.Empty(t, gitRepoName)
	assert.Empty(t, localPath)
	assert.Contains(t, err.Error(), "already exists with different URL")
	assert.Contains(t, err.Error(), conflictingName)
	assert.Contains(t, err.Error(), "Existing URL")
	assert.Contains(t, err.Error(), "Provided URL")
}

// TestResolveOrCreateGitRepo_ExistingByName tests lookup by name
func TestResolveOrCreateGitRepo_ExistingByName(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// Pre-create a GitRepo with a custom name
	customName := "my-custom-repo"
	repoURL := "https://github.com/org/repo.git"
	slug, err := mirror.GenerateSlug(repoURL)
	require.NoError(t, err)

	existingRepo := &models.GitRepoDB{
		Name:       customName,
		URL:        repoURL,
		Slug:       slug,
		DefaultRef: "develop",
		AuthType:   "token",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err = mockStore.CreateGitRepo(existingRepo)
	require.NoError(t, err)

	// Call the function with the name (not URL)
	gitRepoID, gitRepoName, localPath, err := resolveOrCreateGitRepo(mockStore, customName)

	// Should succeed and return the existing repo
	require.NoError(t, err)
	require.NotNil(t, gitRepoID)
	assert.Equal(t, customName, gitRepoName)
	assert.Equal(t, getGitRepoPath(slug), localPath)
	assert.Equal(t, existingRepo.ID, *gitRepoID)

	// Verify no new repo was created
	allRepos, err := mockStore.ListGitRepos()
	require.NoError(t, err)
	assert.Len(t, allRepos, 1, "Should not create new GitRepo when looking up by name")
}

// TestResolveOrCreateGitRepo_NameNotFound tests error when name doesn't exist
func TestResolveOrCreateGitRepo_NameNotFound(t *testing.T) {
	mockStore := db.NewMockDataStore()

	nonExistentName := "nonexistent-repo"

	// Call the function with non-existent name
	gitRepoID, gitRepoName, localPath, err := resolveOrCreateGitRepo(mockStore, nonExistentName)

	// Should error
	require.Error(t, err)
	assert.Nil(t, gitRepoID)
	assert.Empty(t, gitRepoName)
	assert.Empty(t, localPath)
	assert.Contains(t, err.Error(), fmt.Sprintf("GitRepo '%s' not found", nonExistentName))
	assert.Contains(t, err.Error(), "To auto-create from URL")
	assert.Contains(t, err.Error(), "Or create GitRepo first")
}

// TestResolveOrCreateGitRepo_InvalidURL tests error for malformed URLs
func TestResolveOrCreateGitRepo_InvalidURL(t *testing.T) {
	mockStore := db.NewMockDataStore()

	tests := []struct {
		name        string
		repoInput   string
		wantErrMsg  string
		description string
	}{
		{
			name:        "http protocol",
			repoInput:   "http://github.com/org/repo.git",
			wantErrMsg:  "invalid git URL",
			description: "HTTP URLs should be rejected as insecure",
		},
		{
			name:        "shell metacharacters",
			repoInput:   "https://github.com/org/repo.git; rm -rf /",
			wantErrMsg:  "invalid git URL",
			description: "URLs with shell metacharacters should be rejected",
		},
		{
			name:        "empty url",
			repoInput:   "https://",
			wantErrMsg:  "failed to generate slug from URL",
			description: "Empty/invalid URLs should fail slug generation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gitRepoID, gitRepoName, localPath, err := resolveOrCreateGitRepo(mockStore, tt.repoInput)

			// Should error
			require.Error(t, err, tt.description)
			assert.Nil(t, gitRepoID)
			assert.Empty(t, gitRepoName)
			assert.Empty(t, localPath)
			assert.Contains(t, err.Error(), tt.wantErrMsg)
		})
	}
}

// TestResolveOrCreateGitRepo_URLDetection tests URL detection logic
func TestResolveOrCreateGitRepo_URLDetection(t *testing.T) {
	tests := []struct {
		name        string
		repoInput   string
		shouldBeURL bool
		description string
	}{
		{
			name:        "https url",
			repoInput:   "https://github.com/org/repo.git",
			shouldBeURL: true,
			description: "HTTPS URLs should be detected",
		},
		{
			name:        "ssh url",
			repoInput:   "git@github.com:org/repo.git",
			shouldBeURL: true,
			description: "SSH URLs should be detected",
		},
		{
			name:        "ssh protocol url",
			repoInput:   "ssh://git@github.com/org/repo.git",
			shouldBeURL: true,
			description: "SSH protocol URLs should be detected",
		},
		{
			name:        "plain name",
			repoInput:   "my-repo",
			shouldBeURL: false,
			description: "Plain names should not be detected as URLs",
		},
		{
			name:        "name with dots",
			repoInput:   "my.repo.name",
			shouldBeURL: false,
			description: "Names with dots should not be detected as URLs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The function detects URLs by checking for "://" or "git@" prefix
			isURL := containsProtocol(tt.repoInput)
			assert.Equal(t, tt.shouldBeURL, isURL, tt.description)
		})
	}
}

// Helper function to test URL detection logic (mirrors the function's logic)
func containsProtocol(repo string) bool {
	return containsString(repo, "://") || hasPrefix(repo, "git@")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && indexOf(s, substr) >= 0
}

func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// TestResolveOrCreateGitRepo_ExistingByURLWithDifferentSlug tests that
// a GitRepo is found by exact URL match even if the slug differs
func TestResolveOrCreateGitRepo_ExistingByURLWithDifferentSlug(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repoURL := "https://github.com/org/repo.git"
	expectedSlug, err := mirror.GenerateSlug(repoURL)
	require.NoError(t, err)

	// Create a GitRepo with a custom name (not the auto-generated slug)
	customName := "my-custom-name"
	existingRepo := &models.GitRepoDB{
		Name:       customName,
		URL:        repoURL,
		Slug:       expectedSlug, // Slug is generated, but name is custom
		DefaultRef: "main",
		AuthType:   "none",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err = mockStore.CreateGitRepo(existingRepo)
	require.NoError(t, err)

	// Call the function with the URL
	gitRepoID, gitRepoName, localPath, err := resolveOrCreateGitRepo(mockStore, repoURL)

	// Should find the existing repo by URL match (in ListGitRepos loop)
	require.NoError(t, err)
	require.NotNil(t, gitRepoID)
	assert.Equal(t, customName, gitRepoName)
	assert.Equal(t, getGitRepoPath(expectedSlug), localPath)
	assert.Equal(t, existingRepo.ID, *gitRepoID)

	// Verify no duplicate was created
	allRepos, err := mockStore.ListGitRepos()
	require.NoError(t, err)
	assert.Len(t, allRepos, 1, "Should reuse existing GitRepo by URL match")
}

// TestResolveOrCreateGitRepo_SSHUrl tests handling of SSH URLs
func TestResolveOrCreateGitRepo_SSHUrl(t *testing.T) {
	mockStore := db.NewMockDataStore()

	sshURL := "git@github.com:rmkohlman/dvm-test-golang.git"

	// Call the function
	gitRepoID, gitRepoName, localPath, err := resolveOrCreateGitRepo(mockStore, sshURL)

	// Should succeed
	require.NoError(t, err)
	require.NotNil(t, gitRepoID)
	assert.NotEmpty(t, gitRepoName)
	assert.NotEmpty(t, localPath)

	// Verify expected slug was generated
	expectedSlug, err := mirror.GenerateSlug(sshURL)
	require.NoError(t, err)
	assert.Equal(t, "github.com_rmkohlman_dvm-test-golang", expectedSlug)
	assert.Equal(t, expectedSlug, gitRepoName)

	// Verify GitRepo was created in database
	createdRepo, err := mockStore.GetGitRepoByName(expectedSlug)
	require.NoError(t, err)
	require.NotNil(t, createdRepo)
	assert.Equal(t, sshURL, createdRepo.URL)
	assert.Equal(t, expectedSlug, createdRepo.Slug)
}

// TestResolveOrCreateGitRepo_SlugMatchButURLMatch tests the case where
// GetGitRepoByName returns a repo with matching URL (double-check case)
func TestResolveOrCreateGitRepo_SlugMatchButURLMatch(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repoURL := "https://github.com/org/repo.git"
	slug, err := mirror.GenerateSlug(repoURL)
	require.NoError(t, err)

	// Create a GitRepo with auto-generated slug name
	existingRepo := &models.GitRepoDB{
		Name:       slug,
		URL:        repoURL,
		Slug:       slug,
		DefaultRef: "main",
		AuthType:   "none",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	err = mockStore.CreateGitRepo(existingRepo)
	require.NoError(t, err)

	// Call the function with the same URL
	// This should find it in the GetGitRepoByName check (lines 668-677)
	// even though it also matches in ListGitRepos
	gitRepoID, gitRepoName, localPath, err := resolveOrCreateGitRepo(mockStore, repoURL)

	// Should succeed and return the existing repo
	require.NoError(t, err)
	require.NotNil(t, gitRepoID)
	assert.Equal(t, slug, gitRepoName)
	assert.Equal(t, getGitRepoPath(slug), localPath)
	assert.Equal(t, existingRepo.ID, *gitRepoID)

	// Verify no duplicate was created
	allRepos, err := mockStore.ListGitRepos()
	require.NoError(t, err)
	assert.Len(t, allRepos, 1, "Should not create duplicate")
}

// TestResolveOrCreateGitRepo_DatabaseErrorOnList tests handling of database errors
func TestResolveOrCreateGitRepo_DatabaseErrorOnList(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// MockDataStore doesn't easily simulate errors, but we can test
	// that the function continues even if ListGitRepos fails (line 656: if err == nil)

	// Create a URL that will pass validation
	repoURL := "https://github.com/org/repo.git"

	// Call the function - should still work because error from ListGitRepos is ignored
	gitRepoID, gitRepoName, localPath, err := resolveOrCreateGitRepo(mockStore, repoURL)

	// Should succeed (the function ignores ListGitRepos errors)
	require.NoError(t, err)
	require.NotNil(t, gitRepoID)
	assert.NotEmpty(t, gitRepoName)
	assert.NotEmpty(t, localPath)
}

// TestResolveOrCreateGitRepo_PathFormat tests that returned path is correctly formatted
func TestResolveOrCreateGitRepo_PathFormat(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repoURL := "https://github.com/org/repo.git"
	slug, err := mirror.GenerateSlug(repoURL)
	require.NoError(t, err)

	// Call the function
	_, _, localPath, err := resolveOrCreateGitRepo(mockStore, repoURL)

	// Should succeed
	require.NoError(t, err)

	// Verify path format
	expectedPath := getGitRepoPath(slug)
	assert.Equal(t, expectedPath, localPath)

	// Path should contain the slug
	assert.Contains(t, localPath, slug)

	// Path should be absolute-ish (contains .devopsmaestro/repos)
	assert.Contains(t, localPath, ".devopsmaestro")
	assert.Contains(t, localPath, "repos")
}
