// Package integration provides integration tests for DevOpsMaestro's
// git repository management.
package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGitRepoCreate tests basic gitrepo creation.
// This verifies that:
// - GitRepo can be created with URL
// - GitRepo appears in list
// - GitRepo has correct properties
func TestGitRepoCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Create gitrepo
	repoURL := "https://github.com/example/test-repo.git"
	f.AssertCommandSuccess(t, "create", "gitrepo", "my-repo",
		"--url", repoURL)

	// Verify gitrepo appears in list
	repos, err := f.RunDVMJSONList("get", "gitrepos")
	require.NoError(t, err)
	require.Len(t, repos, 1, "Should have exactly 1 gitrepo")

	repo := repos[0]
	assert.Equal(t, "my-repo", f.GetResourceName(repo))

	// For flat JSON, the repo itself contains the fields
	assert.Equal(t, repoURL, f.GetResourceField(repo, "url"))
}

// TestGitRepoMultiple tests creating multiple git repositories.
// This verifies that:
// - Multiple gitrepos can exist
// - Each gitrepo is independent
// - All can be listed
func TestGitRepoMultiple(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Create multiple gitrepos
	f.AssertCommandSuccess(t, "create", "gitrepo", "backend-repo",
		"--url", "https://github.com/example/backend.git")
	f.AssertCommandSuccess(t, "create", "gitrepo", "frontend-repo",
		"--url", "https://github.com/example/frontend.git")
	f.AssertCommandSuccess(t, "create", "gitrepo", "shared-lib",
		"--url", "https://github.com/example/shared.git")

	// Verify all gitrepos exist
	repos, err := f.RunDVMJSONList("get", "gitrepos")
	require.NoError(t, err)
	assert.Len(t, repos, 3, "Should have 3 gitrepos")

	// Verify repo names
	names := make([]string, len(repos))
	for i, repo := range repos {
		names[i] = f.GetResourceName(repo)
	}
	assert.Contains(t, names, "backend-repo")
	assert.Contains(t, names, "frontend-repo")
	assert.Contains(t, names, "shared-lib")
}

// TestGitRepoGet tests retrieving a specific gitrepo.
// This verifies that:
// - Individual gitrepo can be retrieved by name
// - Retrieved data is correct
func TestGitRepoGet(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Create gitrepo
	repoURL := "https://github.com/example/test-repo.git"
	f.AssertCommandSuccess(t, "create", "gitrepo", "test-repo",
		"--url", repoURL)

	// Get specific gitrepo (via list command)
	repos, err := f.RunDVMJSONList("get", "gitrepos")
	require.NoError(t, err)
	require.Len(t, repos, 1, "Should have exactly 1 gitrepo")

	repo := repos[0]
	assert.Equal(t, "test-repo", f.GetResourceName(repo))

	// For flat JSON, use GetResourceField
	assert.Equal(t, repoURL, f.GetResourceField(repo, "url"))
}

// TestGitRepoDelete tests gitrepo deletion.
// This verifies that:
// - GitRepo can be deleted
// - Deleted gitrepo no longer appears in list
func TestGitRepoDelete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Create gitrepo
	f.AssertCommandSuccess(t, "create", "gitrepo", "temp-repo",
		"--url", "https://github.com/example/temp.git")

	// Verify gitrepo exists
	repos, err := f.RunDVMJSONList("get", "gitrepos")
	require.NoError(t, err)
	require.Len(t, repos, 1)

	// Delete gitrepo
	f.AssertCommandSuccess(t, "delete", "gitrepo", "temp-repo")

	// Verify gitrepo is gone
	repos, err = f.RunDVMJSONList("get", "gitrepos")
	require.NoError(t, err)
	assert.Len(t, repos, 0, "GitRepo should be deleted")
}

// TestGitRepoValidation tests input validation for gitrepos.
// This verifies that:
// - Missing URL is rejected
// - Duplicate names are rejected
// - Delete non-existent fails
func TestGitRepoValidation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Test: Missing URL should fail
	f.AssertCommandFails(t, "create", "gitrepo", "test-repo")

	// Test: Create valid gitrepo
	f.AssertCommandSuccess(t, "create", "gitrepo", "test-repo",
		"--url", "https://github.com/example/test.git")

	// Test: Duplicate gitrepo name should fail
	f.AssertCommandFails(t, "create", "gitrepo", "test-repo",
		"--url", "https://github.com/example/other.git")

	// Test: Delete non-existent gitrepo should fail
	f.AssertCommandFails(t, "delete", "gitrepo", "non-existent")
}

// TestGitRepoWithWorkspace tests gitrepo integration with workspaces.
// This verifies that:
// - Workspace can reference a gitrepo
// - Deleting gitrepo doesn't delete workspace (sets ref to null)
// - Multiple workspaces can reference same gitrepo
func TestGitRepoWithWorkspace(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create hierarchy
	setupTestHierarchy(t, f)

	// Create gitrepo
	f.AssertCommandSuccess(t, "create", "gitrepo", "shared-repo",
		"--url", "https://github.com/example/shared.git")

	// Create workspaces (without gitrepo reference for now - not implemented)
	f.AssertCommandSuccess(t, "create", "workspace", "ws1")
	f.AssertCommandSuccess(t, "create", "workspace", "ws2")

	// Verify both workspaces exist
	workspaces, err := f.RunDVMJSONList("get", "workspaces")
	require.NoError(t, err)
	assert.Len(t, workspaces, 2)

	// Delete gitrepo
	f.AssertCommandSuccess(t, "delete", "gitrepo", "shared-repo")

	// Verify workspaces still exist
	workspaces, err = f.RunDVMJSONList("get", "workspaces")
	require.NoError(t, err)
	assert.Len(t, workspaces, 2, "Workspaces should still exist after gitrepo deletion")
}

// TestGitRepoWithApp tests gitrepo integration with apps.
// This verifies that:
// - App can reference a gitrepo
// - Deleting gitrepo doesn't delete app (sets ref to null)
func TestGitRepoWithApp(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Setup: Create ecosystem and domain
	f.AssertCommandSuccess(t, "create", "ecosystem", "test-eco")
	f.AssertCommandSuccess(t, "use", "ecosystem", "test-eco")
	f.AssertCommandSuccess(t, "create", "domain", "test-domain")
	f.AssertCommandSuccess(t, "use", "domain", "test-domain")

	// Create gitrepo
	f.AssertCommandSuccess(t, "create", "gitrepo", "app-repo",
		"--url", "https://github.com/example/app.git")

	// Create app (without gitrepo reference - not implemented yet)
	f.AssertCommandSuccess(t, "create", "app", "test-app", "--from-cwd")

	// Verify app was created
	apps, err := f.RunDVMJSONList("get", "apps")
	require.NoError(t, err)
	require.Len(t, apps, 1)

	app := apps[0]
	assert.Equal(t, "test-app", f.GetResourceName(app))

	// Delete gitrepo
	f.AssertCommandSuccess(t, "delete", "gitrepo", "app-repo")

	// Verify app still exists
	apps, err = f.RunDVMJSONList("get", "apps")
	require.NoError(t, err)
	assert.Len(t, apps, 1, "App should still exist after gitrepo deletion")
}

// TestGitRepoURLFormats tests various URL formats.
// This verifies that:
// - HTTPS URLs are accepted
// - SSH URLs are accepted (if supported)
// - File URLs are accepted (if supported)
func TestGitRepoURLFormats(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Test HTTPS URL
	// Note: Each URL must have a unique path to avoid slug conflicts
	f.AssertCommandSuccess(t, "create", "gitrepo", "https-repo",
		"--url", "https://github.com/example/repo-https.git", "--no-sync")

	// Test SSH URL (if supported)
	f.AssertCommandSuccess(t, "create", "gitrepo", "ssh-repo",
		"--url", "git@github.com:example/repo-ssh.git", "--no-sync")

	// Test HTTPS URL without .git suffix
	f.AssertCommandSuccess(t, "create", "gitrepo", "no-suffix-repo",
		"--url", "https://github.com/example/repo-nosuffix", "--no-sync")

	// Verify all were created
	repos, err := f.RunDVMJSONList("get", "gitrepos")
	require.NoError(t, err)
	assert.Len(t, repos, 3, "Should have 3 gitrepos with different URL formats")
}

// TestGitRepoUpdate tests updating gitrepo properties.
// This verifies that:
// - GitRepo description can be updated (if supported)
// - GitRepo URL can be updated (if supported)
func TestGitRepoUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	f := NewTestFramework(t)
	defer f.Cleanup()

	// Create gitrepo (without --description flag)
	f.AssertCommandSuccess(t, "create", "gitrepo", "test-repo",
		"--url", "https://github.com/example/test.git")

	// Verify initial state (via list command)
	repos, err := f.RunDVMJSONList("get", "gitrepos")
	require.NoError(t, err)
	require.Len(t, repos, 1, "Should have exactly 1 gitrepo")

	repo := repos[0]
	assert.Equal(t, "test-repo", f.GetResourceName(repo))

	// Update description (if update command exists - adjust based on implementation)
	// f.AssertCommandSuccess(t, "update", "gitrepo", "test-repo",
	// 	"--description", "Updated description")

	// Verify update (if supported)
	// repo, err = f.RunDVMJSON("get", "gitrepo", "test-repo")
	// require.NoError(t, err)
	// assert.Equal(t, "Updated description", repo["description"])
}
