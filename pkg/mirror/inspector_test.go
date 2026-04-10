package mirror

// =============================================================================
// TDD Phase 2 (RED): MirrorInspector Interface — Issue #223
// =============================================================================
// These tests drive the implementation of the MirrorInspector interface,
// which provides read-only inspection of bare git mirrors.
//
// RED state: ALL tests FAIL because:
//   - MirrorInspector interface does not exist in interfaces.go
//   - RefInfo type does not exist
//   - ListBranches, ListTags, DiskUsage, Verify methods not on GitMirrorManager
//
// GREEN state: After @dvm-core implements MirrorInspector per Issue #223.
// =============================================================================

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Setup helpers for inspector tests
// =============================================================================

// createTestMirrorWithBranches creates a bare mirror that has commits on
// multiple branches, so ListBranches returns non-empty results.
func createTestMirrorWithBranches(t *testing.T, baseDir, slug string, branches []string) string {
	t.Helper()

	// Create a remote repo with commits on each branch
	remoteDir := t.TempDir()
	runGit(t, remoteDir, "init")
	runGit(t, remoteDir, "config", "user.name", "Test User")
	runGit(t, remoteDir, "config", "user.email", "test@example.com")

	// Commit on main/master
	err := os.WriteFile(filepath.Join(remoteDir, "README.md"), []byte("# test\n"), 0644)
	require.NoError(t, err)
	runGit(t, remoteDir, "add", ".")
	runGit(t, remoteDir, "commit", "-m", "Initial commit")

	// Create additional branches
	for _, branch := range branches {
		if branch != "main" && branch != "master" {
			runGit(t, remoteDir, "checkout", "-b", branch)
			err = os.WriteFile(filepath.Join(remoteDir, branch+".txt"), []byte(branch), 0644)
			require.NoError(t, err)
			runGit(t, remoteDir, "add", ".")
			runGit(t, remoteDir, "commit", "-m", "Branch "+branch)
			runGit(t, remoteDir, "checkout", "-")
		}
	}

	// Clone as bare mirror
	mirrorPath := filepath.Join(baseDir, slug)
	cmd := exec.Command("git", "clone", "--mirror", "--", remoteDir, mirrorPath)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git clone --mirror failed: %s", out)

	return mirrorPath
}

// createTestMirrorWithTags creates a bare mirror with annotated tags.
func createTestMirrorWithTags(t *testing.T, baseDir, slug string, tags []string) string {
	t.Helper()

	remoteDir := t.TempDir()
	runGit(t, remoteDir, "init")
	runGit(t, remoteDir, "config", "user.name", "Test User")
	runGit(t, remoteDir, "config", "user.email", "test@example.com")

	err := os.WriteFile(filepath.Join(remoteDir, "README.md"), []byte("# test\n"), 0644)
	require.NoError(t, err)
	runGit(t, remoteDir, "add", ".")
	runGit(t, remoteDir, "commit", "-m", "Initial commit")

	// Create lightweight tags
	for _, tag := range tags {
		runGit(t, remoteDir, "tag", tag)
	}

	// Clone as bare mirror
	mirrorPath := filepath.Join(baseDir, slug)
	cmd := exec.Command("git", "clone", "--mirror", "--", remoteDir, mirrorPath)
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git clone --mirror failed: %s", out)

	return mirrorPath
}

// runGit is a test helper to run a git command in a directory.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, "git %s failed: %s", strings.Join(args, " "), out)
}

// =============================================================================
// ListBranches tests
// =============================================================================

func TestListBranches(t *testing.T) {
	baseDir := t.TempDir()
	mgr := NewGitMirrorManager(baseDir)

	slug := "github.com_org_branches-repo"
	wantBranches := []string{"feature-a", "feature-b"}
	createTestMirrorWithBranches(t, baseDir, slug, wantBranches)

	// Act: MirrorInspector interface
	inspector, ok := mgr.(MirrorInspector)
	require.True(t, ok, "GitMirrorManager must implement MirrorInspector")

	branches, err := inspector.ListBranches(slug)
	require.NoError(t, err)

	// Assert: at least the feature branches exist
	names := make(map[string]bool)
	for _, b := range branches {
		names[b.Name] = true
	}
	for _, want := range wantBranches {
		assert.True(t, names[want], "ListBranches() missing branch %q", want)
	}
}

func TestListBranches_MirrorNotFound(t *testing.T) {
	baseDir := t.TempDir()
	mgr := NewGitMirrorManager(baseDir)

	inspector, ok := mgr.(MirrorInspector)
	require.True(t, ok, "GitMirrorManager must implement MirrorInspector")

	_, err := inspector.ListBranches("nonexistent-slug")
	assert.Error(t, err, "ListBranches() on nonexistent mirror should return error")
}

// =============================================================================
// ListTags tests
// =============================================================================

func TestListTags(t *testing.T) {
	baseDir := t.TempDir()
	mgr := NewGitMirrorManager(baseDir)

	slug := "github.com_org_tags-repo"
	wantTags := []string{"v1.0.0", "v1.1.0", "v2.0.0"}
	createTestMirrorWithTags(t, baseDir, slug, wantTags)

	inspector, ok := mgr.(MirrorInspector)
	require.True(t, ok, "GitMirrorManager must implement MirrorInspector")

	tags, err := inspector.ListTags(slug)
	require.NoError(t, err)

	names := make(map[string]bool)
	for _, tag := range tags {
		names[tag.Name] = true
	}
	for _, want := range wantTags {
		assert.True(t, names[want], "ListTags() missing tag %q", want)
	}
}

func TestListTags_EmptyRepo(t *testing.T) {
	baseDir := t.TempDir()
	mgr := NewGitMirrorManager(baseDir)

	slug := "github.com_org_notags-repo"
	createTestMirrorWithTags(t, baseDir, slug, []string{})

	inspector, ok := mgr.(MirrorInspector)
	require.True(t, ok, "GitMirrorManager must implement MirrorInspector")

	tags, err := inspector.ListTags(slug)
	require.NoError(t, err)

	// Empty slice, not nil
	assert.NotNil(t, tags, "ListTags() should return empty slice, not nil")
	assert.Empty(t, tags, "ListTags() should return empty slice for repo with no tags")
}

// =============================================================================
// DiskUsage test
// =============================================================================

func TestDiskUsage(t *testing.T) {
	baseDir := t.TempDir()
	mgr := NewGitMirrorManager(baseDir)

	slug := "github.com_org_diskusage-repo"
	createTestMirrorWithBranches(t, baseDir, slug, []string{})

	inspector, ok := mgr.(MirrorInspector)
	require.True(t, ok, "GitMirrorManager must implement MirrorInspector")

	bytes, err := inspector.DiskUsage(slug)
	require.NoError(t, err)
	assert.Greater(t, bytes, int64(0), "DiskUsage() should return positive bytes for a real mirror")
}

// =============================================================================
// Fsck (Verify) test
// =============================================================================

func TestFsck(t *testing.T) {
	baseDir := t.TempDir()
	mgr := NewGitMirrorManager(baseDir)

	slug := "github.com_org_fsck-repo"
	createTestMirrorWithBranches(t, baseDir, slug, []string{})

	inspector, ok := mgr.(MirrorInspector)
	require.True(t, ok, "GitMirrorManager must implement MirrorInspector")

	// A freshly cloned mirror should pass fsck
	err := inspector.Verify(slug)
	assert.NoError(t, err, "Verify() on a healthy mirror should return nil")
}
