package mirror

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Test Helpers
// =============================================================================

// setupTestMirrorManager creates a GitMirrorManager with a temporary base directory.
// Uses direct struct construction since this is an internal (same-package) test.
func setupTestMirrorManager(t *testing.T) *GitMirrorManager {
	t.Helper()
	baseDir := t.TempDir()
	return &GitMirrorManager{baseDir: baseDir}
}

// createTestMirror creates a minimal bare git repository for testing.
// This is used to set up test fixtures where a mirror should already exist.
func createTestMirror(t *testing.T, baseDir, slug string) string {
	t.Helper()

	mirrorPath := filepath.Join(baseDir, slug)
	err := os.MkdirAll(mirrorPath, 0700)
	require.NoError(t, err)

	// Initialize bare git repository
	cmd := exec.Command("git", "init", "--bare", mirrorPath)
	err = cmd.Run()
	require.NoError(t, err, "failed to initialize test bare repo")

	return mirrorPath
}

// createTestRemoteRepo creates a minimal git repository to use as a remote.
// Returns the path to the repository.
func createTestRemoteRepo(t *testing.T) string {
	t.Helper()

	remoteDir := t.TempDir()

	// Initialize repository
	cmd := exec.Command("git", "init", remoteDir)
	err := cmd.Run()
	require.NoError(t, err, "failed to initialize test remote repo")

	// Configure git
	cmd = exec.Command("git", "-C", remoteDir, "config", "user.name", "Test User")
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "-C", remoteDir, "config", "user.email", "test@example.com")
	err = cmd.Run()
	require.NoError(t, err)

	// Create a test file
	testFile := filepath.Join(remoteDir, "README.md")
	err = os.WriteFile(testFile, []byte("# Test Repo\n"), 0644)
	require.NoError(t, err)

	// Add and commit
	cmd = exec.Command("git", "-C", remoteDir, "add", ".")
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "-C", remoteDir, "commit", "-m", "Initial commit")
	err = cmd.Run()
	require.NoError(t, err)

	return remoteDir
}

// =============================================================================
// Task 2.3: Bare Mirror Creation Tests
// =============================================================================

func TestMirrorManager_Clone(t *testing.T) {
	mgr := setupTestMirrorManager(t)
	remoteRepo := createTestRemoteRepo(t)

	tests := []struct {
		name    string
		url     string
		slug    string
		wantErr bool
	}{
		{
			name:    "creates bare mirror at expected path",
			url:     remoteRepo,
			slug:    "test.com_user_repo",
			wantErr: false,
		},
		{
			name:    "creates mirror with nested slug",
			url:     remoteRepo,
			slug:    "github.com_org_team_project",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := mgr.Clone(tt.url, tt.slug)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, path)

			// Verify mirror exists at expected path
			expectedPath := mgr.GetPath(tt.slug)
			assert.Equal(t, expectedPath, path)

			// Verify it's a bare repository
			gitDir := filepath.Join(path, "HEAD")
			assert.FileExists(t, gitDir, "bare repo should have HEAD file in root")

			// Verify directory permissions are 0700
			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.True(t, info.IsDir())
			assert.Equal(t, os.FileMode(0700), info.Mode().Perm(), "mirror directory should have 0700 permissions")
		})
	}
}

func TestMirrorManager_Clone_InvalidURL(t *testing.T) {
	mgr := setupTestMirrorManager(t)

	tests := []struct {
		name    string
		url     string
		slug    string
		wantErr string
	}{
		{
			name:    "rejects empty url",
			url:     "",
			slug:    "test.com_user_repo",
			wantErr: "empty",
		},
		{
			name:    "rejects url with shell metacharacters",
			url:     "https://github.com/user/repo; rm -rf /",
			slug:    "test.com_user_repo",
			wantErr: "invalid",
		},
		{
			name:    "rejects url with pipe",
			url:     "https://github.com/user/repo | cat /etc/passwd",
			slug:    "test.com_user_repo",
			wantErr: "invalid",
		},
		{
			name:    "rejects insecure http protocol",
			url:     "http://github.com/user/repo",
			slug:    "test.com_user_repo",
			wantErr: "insecure",
		},
		{
			name:    "rejects url with embedded credentials",
			url:     "https://user:pass@github.com/user/repo",
			slug:    "test.com_user_repo",
			wantErr: "credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.Clone(tt.url, tt.slug)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestMirrorManager_Clone_InvalidSlug(t *testing.T) {
	mgr := setupTestMirrorManager(t)
	remoteRepo := createTestRemoteRepo(t)

	tests := []struct {
		name    string
		slug    string
		wantErr string
	}{
		{
			name:    "rejects empty slug",
			slug:    "",
			wantErr: "empty",
		},
		{
			name:    "rejects slug with path traversal",
			slug:    "../parent",
			wantErr: "invalid",
		},
		{
			name:    "rejects slug with forward slash",
			slug:    "github.com/user/repo",
			wantErr: "invalid",
		},
		{
			name:    "rejects slug with shell metacharacter",
			slug:    "repo;cmd",
			wantErr: "invalid",
		},
		{
			name:    "rejects slug starting with dash",
			slug:    "-option",
			wantErr: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mgr.Clone(remoteRepo, tt.slug)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestMirrorManager_Clone_AlreadyExists(t *testing.T) {
	mgr := setupTestMirrorManager(t)
	remoteRepo := createTestRemoteRepo(t)
	slug := "test.com_user_repo"

	// Create mirror first time - should succeed
	_, err := mgr.Clone(remoteRepo, slug)
	require.NoError(t, err)

	// Try to create same mirror again - should handle gracefully
	_, err = mgr.Clone(remoteRepo, slug)
	assert.Error(t, err, "cloning to existing mirror should return error")
	assert.Contains(t, err.Error(), "exists", "error should indicate mirror already exists")
}

func TestMirrorManager_Clone_DirectoryPermissions(t *testing.T) {
	mgr := setupTestMirrorManager(t)
	remoteRepo := createTestRemoteRepo(t)
	slug := "test.com_user_repo"

	path, err := mgr.Clone(remoteRepo, slug)
	require.NoError(t, err)

	// Verify directory has 0700 permissions (rwx------)
	info, err := os.Stat(path)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0700), info.Mode().Perm(),
		"mirror directory must have 0700 permissions for security")
}

// =============================================================================
// Task 2.4: Workspace Clone from Mirror Tests
// =============================================================================

func TestMirrorManager_CloneToWorkspace(t *testing.T) {
	mgr := setupTestMirrorManager(t)
	remoteRepo := createTestRemoteRepo(t)
	slug := "test.com_user_repo"

	// Create mirror first
	_, err := mgr.Clone(remoteRepo, slug)
	require.NoError(t, err)

	// Create destination for workspace clone
	destPath := filepath.Join(t.TempDir(), "workspace")

	err = mgr.CloneToWorkspace(slug, destPath, "")
	require.NoError(t, err)

	// Verify workspace directory exists
	assert.DirExists(t, destPath)

	// Verify it's a working repository (not bare)
	dotGit := filepath.Join(destPath, ".git")
	assert.DirExists(t, dotGit, "workspace clone should have .git directory")

	// Verify README.md from remote is present
	readme := filepath.Join(destPath, "README.md")
	assert.FileExists(t, readme, "workspace should contain files from repository")

	content, err := os.ReadFile(readme)
	require.NoError(t, err)
	assert.Contains(t, string(content), "Test Repo", "file content should match remote")
}

func TestMirrorManager_CloneToWorkspace_MirrorNotExist(t *testing.T) {
	mgr := setupTestMirrorManager(t)
	destPath := filepath.Join(t.TempDir(), "workspace")

	err := mgr.CloneToWorkspace("nonexistent_mirror", destPath, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not exist", "error should indicate mirror doesn't exist")
}

func TestMirrorManager_CloneToWorkspace_WithRef(t *testing.T) {
	mgr := setupTestMirrorManager(t)

	// Create remote with a branch
	remoteRepo := createTestRemoteRepo(t)

	// Create a branch in remote
	cmd := exec.Command("git", "-C", remoteRepo, "checkout", "-b", "feature-branch")
	err := cmd.Run()
	require.NoError(t, err)

	// Add a file on the branch
	branchFile := filepath.Join(remoteRepo, "feature.txt")
	err = os.WriteFile(branchFile, []byte("feature content"), 0644)
	require.NoError(t, err)

	cmd = exec.Command("git", "-C", remoteRepo, "add", ".")
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "-C", remoteRepo, "commit", "-m", "Add feature")
	err = cmd.Run()
	require.NoError(t, err)

	// Switch back to main
	cmd = exec.Command("git", "-C", remoteRepo, "checkout", "master")
	_ = cmd.Run() // Ignore error if master doesn't exist

	// Create mirror
	slug := "test.com_user_repo"
	_, err = mgr.Clone(remoteRepo, slug)
	require.NoError(t, err)

	// Clone to workspace with specific ref
	destPath := filepath.Join(t.TempDir(), "workspace")
	err = mgr.CloneToWorkspace(slug, destPath, "feature-branch")
	require.NoError(t, err)

	// Verify we're on the correct branch
	cmd = exec.Command("git", "-C", destPath, "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(output), "feature-branch", "workspace should be on specified branch")

	// Verify feature file exists
	featureFile := filepath.Join(destPath, "feature.txt")
	assert.FileExists(t, featureFile, "feature file from branch should exist")
}

func TestMirrorManager_CloneToWorkspace_DestExists(t *testing.T) {
	mgr := setupTestMirrorManager(t)
	remoteRepo := createTestRemoteRepo(t)
	slug := "test.com_user_repo"

	// Create mirror
	_, err := mgr.Clone(remoteRepo, slug)
	require.NoError(t, err)

	// Create destination directory with existing content
	destPath := filepath.Join(t.TempDir(), "workspace")
	err = os.MkdirAll(destPath, 0755)
	require.NoError(t, err)

	existingFile := filepath.Join(destPath, "existing.txt")
	err = os.WriteFile(existingFile, []byte("existing content"), 0644)
	require.NoError(t, err)

	// Try to clone to existing destination
	err = mgr.CloneToWorkspace(slug, destPath, "")
	assert.Error(t, err, "cloning to existing directory should return error")
	assert.Contains(t, err.Error(), "exists", "error should indicate destination exists")
}

// =============================================================================
// Task 2.5: Mirror Sync Operations Tests
// =============================================================================

func TestMirrorManager_Sync(t *testing.T) {
	mgr := setupTestMirrorManager(t)
	remoteRepo := createTestRemoteRepo(t)
	slug := "test.com_user_repo"

	// Create mirror
	_, err := mgr.Clone(remoteRepo, slug)
	require.NoError(t, err)

	// Add a new commit to remote
	newFile := filepath.Join(remoteRepo, "new.txt")
	err = os.WriteFile(newFile, []byte("new content"), 0644)
	require.NoError(t, err)

	cmd := exec.Command("git", "-C", remoteRepo, "add", ".")
	err = cmd.Run()
	require.NoError(t, err)

	cmd = exec.Command("git", "-C", remoteRepo, "commit", "-m", "Add new file")
	err = cmd.Run()
	require.NoError(t, err)

	// Sync mirror
	err = mgr.Sync(slug)
	require.NoError(t, err)

	// Verify mirror has the new commit
	mirrorPath := mgr.GetPath(slug)
	cmd = exec.Command("git", "-C", mirrorPath, "log", "--oneline")
	output, err := cmd.Output()
	require.NoError(t, err)
	assert.Contains(t, string(output), "Add new file", "mirror should have new commit after sync")
}

func TestMirrorManager_Sync_NotExist(t *testing.T) {
	mgr := setupTestMirrorManager(t)

	err := mgr.Sync("nonexistent_mirror")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not exist", "error should indicate mirror doesn't exist")
}

func TestMirrorManager_Delete(t *testing.T) {
	mgr := setupTestMirrorManager(t)
	slug := "test.com_user_repo"

	// Create a test mirror
	mirrorPath := createTestMirror(t, mgr.baseDir, slug)
	assert.DirExists(t, mirrorPath)

	// Delete mirror
	err := mgr.Delete(slug)
	require.NoError(t, err)

	// Verify mirror is removed
	assert.NoDirExists(t, mirrorPath, "mirror directory should be removed")
	assert.False(t, mgr.Exists(slug), "Exists should return false after deletion")
}

func TestMirrorManager_Delete_NotExist(t *testing.T) {
	mgr := setupTestMirrorManager(t)

	// Delete non-existent mirror should be handled gracefully
	err := mgr.Delete("nonexistent_mirror")
	assert.NoError(t, err, "deleting non-existent mirror should not error")
}

func TestMirrorManager_Exists(t *testing.T) {
	mgr := setupTestMirrorManager(t)
	slug := "test.com_user_repo"

	// Initially should not exist
	assert.False(t, mgr.Exists(slug))

	// Create mirror
	createTestMirror(t, mgr.baseDir, slug)

	// Now should exist
	assert.True(t, mgr.Exists(slug))

	// Different slug should not exist
	assert.False(t, mgr.Exists("other_mirror"))
}

func TestMirrorManager_GetPath(t *testing.T) {
	mgr := setupTestMirrorManager(t)

	tests := []struct {
		name         string
		slug         string
		wantContains string
	}{
		{
			name:         "simple slug",
			slug:         "github.com_user_repo",
			wantContains: "github.com_user_repo",
		},
		{
			name:         "nested slug",
			slug:         "gitlab.com_group_subgroup_project",
			wantContains: "gitlab.com_group_subgroup_project",
		},
		{
			name:         "slug with dash",
			slug:         "github.com_user_my-repo",
			wantContains: "github.com_user_my-repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := mgr.GetPath(tt.slug)

			// Path should contain base directory
			assert.Contains(t, path, mgr.baseDir)

			// Path should contain slug
			assert.Contains(t, path, tt.wantContains)

			// Path should be absolute
			assert.True(t, filepath.IsAbs(path), "path should be absolute")

			// Path should end with slug
			assert.Equal(t, tt.slug, filepath.Base(path), "path should end with slug")
		})
	}
}

// =============================================================================
// Interface Compliance Test
// =============================================================================

func TestGitMirrorManager_ImplementsMirrorManager(t *testing.T) {
	var _ MirrorManager = (*GitMirrorManager)(nil)
}
