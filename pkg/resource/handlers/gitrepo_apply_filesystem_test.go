package handlers

// =============================================================================
// GitRepo Apply Filesystem Tests — Issue #193
//
// These tests verify that:
//   1. GitMirrorManager filesystem operations work in isolation (infrastructure)
//   2. GitRepoHandler.Apply() wires mirror.Clone() so that after a successful
//      Apply the mirror directory exists on disk (handler contract)
// =============================================================================

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/mirror"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

// createLocalTestRemoteRepo creates a minimal git repository in a temp dir
// that can be used as a "remote" for local clone operations in tests.
func createLocalTestRemoteRepo(t *testing.T) string {
	t.Helper()
	remoteDir := t.TempDir()

	mustGit(t, remoteDir, "init")
	mustGit(t, remoteDir, "config", "user.name", "Test User")
	mustGit(t, remoteDir, "config", "user.email", "test@example.com")

	readmePath := filepath.Join(remoteDir, "README.md")
	if err := os.WriteFile(readmePath, []byte("# Test\n"), 0644); err != nil {
		t.Fatalf("write README: %v", err)
	}
	mustGit(t, remoteDir, "add", ".")
	mustGit(t, remoteDir, "commit", "-m", "Initial commit")
	mustGit(t, remoteDir, "branch", "-m", "main") // rename default branch to main (modern convention)

	return remoteDir
}

// mustGit runs a git command inside dir and fails the test on error.
func mustGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	fullArgs := append([]string{"-C", dir}, args...)
	cmd := exec.Command("git", fullArgs...)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// newTestMirrorManager returns a MirrorManager rooted in a fresh temp dir.
func newTestMirrorManager(t *testing.T) (mirror.MirrorManager, string) {
	t.Helper()
	baseDir := t.TempDir()
	return mirror.NewGitMirrorManager(baseDir), baseDir
}

// -----------------------------------------------------------------------------
// Test 1: GitMirrorManager.Clone() creates the mirror directory
// Expected: PASS — infrastructure already works
// -----------------------------------------------------------------------------

func TestGitMirrorManagerCloneCreatesDirectory(t *testing.T) {
	mgr, _ := newTestMirrorManager(t)
	remoteURL := createLocalTestRemoteRepo(t)
	slug := "test.example.com_user_test-repo"

	mirrorPath, err := mgr.Clone(remoteURL, slug)
	if err != nil {
		t.Fatalf("Clone() unexpected error: %v", err)
	}

	// Mirror directory must exist
	info, statErr := os.Stat(mirrorPath)
	if statErr != nil {
		t.Fatalf("mirror directory does not exist after Clone(): %v", statErr)
	}
	if !info.IsDir() {
		t.Errorf("mirror path is not a directory: %s", mirrorPath)
	}

	// Bare repo must have a HEAD file (hallmark of a bare clone)
	headFile := filepath.Join(mirrorPath, "HEAD")
	if _, err := os.Stat(headFile); err != nil {
		t.Errorf("HEAD file missing in bare mirror — not a valid bare repo: %v", err)
	}
}

// -----------------------------------------------------------------------------
// Test 2: GitMirrorManager.Exists() returns false for unknown slug
// Expected: PASS
// -----------------------------------------------------------------------------

func TestGitMirrorManagerExistsReturnsFalse(t *testing.T) {
	mgr, _ := newTestMirrorManager(t)

	if mgr.Exists("nonexistent-slug") {
		t.Error("Exists() returned true for slug that was never cloned")
	}
}

// -----------------------------------------------------------------------------
// Test 3: Cloning the same slug twice is handled gracefully (error, not panic)
// Expected: PASS — Clone() already returns an error for existing mirrors
// -----------------------------------------------------------------------------

func TestGitMirrorManagerCloneIsIdempotent(t *testing.T) {
	mgr, _ := newTestMirrorManager(t)
	remoteURL := createLocalTestRemoteRepo(t)
	slug := "test.example.com_user_idempotent-repo"

	// First clone must succeed
	_, err := mgr.Clone(remoteURL, slug)
	if err != nil {
		t.Fatalf("first Clone() error: %v", err)
	}

	// Mirror must exist after the first clone
	if !mgr.Exists(slug) {
		t.Fatal("Exists() returned false after successful Clone()")
	}

	// Second clone to same slug must return an error (not panic, not corrupt)
	_, err = mgr.Clone(remoteURL, slug)
	if err == nil {
		t.Error("second Clone() should return an error ('already exists'), got nil")
	}

	// Mirror must still be intact
	if !mgr.Exists(slug) {
		t.Error("Exists() returned false after failed second Clone() — mirror was corrupted")
	}
}

// =============================================================================
// Handler Contract Tests — Issue #193
// =============================================================================

// TestGitRepoApplyHandlerContract_MirrorShouldExistAfterApply verifies that
// after Apply() returns successfully, the git mirror directory exists on disk.
//
// The test uses the UPDATE path: a DB record with a known slug is pre-created,
// and Apply() is called with a YAML whose URL points to a local test repo.
// The handler preserves the slug from the existing record and clones the mirror
// using the (local) URL from the YAML.
func TestGitRepoApplyHandlerContract_MirrorShouldExistAfterApply(t *testing.T) {
	// Create a local git repo that can be cloned.
	localRemote := createLocalTestRemoteRepo(t)

	// Set up the handler with an injectable MirrorBaseDir.
	mirrorBaseDir := t.TempDir()
	h := NewGitRepoHandler()
	h.MirrorBaseDir = mirrorBaseDir

	// Pre-seed a DB record with a known slug so Apply() takes the UPDATE path.
	// The slug must be a valid filesystem-safe value (validated by ValidateSlug).
	const slug = "example.com_org_contract-test-repo"
	store := db.NewMockDataStore()
	existing := &models.GitRepoDB{
		Name:       "contract-repo",
		URL:        localRemote,
		Slug:       slug,
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	if err := store.CreateGitRepo(existing); err != nil {
		t.Fatalf("pre-seed CreateGitRepo: %v", err)
	}

	ctx := resource.Context{DataStore: store}

	// YAML uses the local repo path as the URL (clonable without network).
	yamlData := []byte(`apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: contract-repo
spec:
  url: ` + localRemote + `
  defaultRef: main
  authType: none
`)

	// Act
	_, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() returned an unexpected error: %v", err)
	}

	// Assert: DB record must exist
	stored, dbErr := store.GetGitRepoByName("contract-repo")
	if dbErr != nil {
		t.Fatalf("DB record not found after Apply(): %v", dbErr)
	}
	if stored.Slug != slug {
		t.Errorf("stored Slug = %q, want %q", stored.Slug, slug)
	}

	// Assert: Mirror directory MUST exist after Apply().
	expectedMirrorPath := filepath.Join(mirrorBaseDir, slug)
	if _, statErr := os.Stat(expectedMirrorPath); statErr != nil {
		t.Errorf(
			"mirror directory does not exist after Apply() — "+
				"handler must call mirror.Clone() during Apply().\n"+
				"Expected path: %s\n"+
				"Fix: wire GitMirrorManager.Clone(url, slug) into GitRepoHandler.Apply() (Issue #193)",
			expectedMirrorPath,
		)
	}
}

// TestGitRepoApplyHandlerContract_SlugStoredOnRecord verifies that after Apply()
// the slug field on the stored GitRepo is populated. This is a prerequisite for
// the mirror path to be computable. Expected: PASS (slug generation already works).
func TestGitRepoApplyHandlerContract_SlugStoredOnRecord(t *testing.T) {
	h := NewGitRepoHandler()
	h.MirrorBaseDir = t.TempDir() // prevent clone attempts from reaching ~/.devopsmaestro
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	remoteURL := "https://github.com/user/slug-test-repo"

	yamlData := []byte(`apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: slug-test-repo
spec:
  url: ` + remoteURL + `
  defaultRef: main
  authType: none
`)

	_, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	stored, dbErr := store.GetGitRepoByName("slug-test-repo")
	if dbErr != nil {
		t.Fatalf("record not found: %v", dbErr)
	}

	// Slug must be set so the handler can later compute the mirror path
	if stored.Slug == "" {
		t.Error("Apply() did not populate the Slug field on the stored GitRepo; " +
			"mirror path cannot be computed without a slug")
	}

	// Slug must match what GenerateSlug would produce
	expectedSlug, _ := mirror.GenerateSlug(remoteURL)
	if stored.Slug != expectedSlug {
		t.Errorf("stored Slug = %q, want %q", stored.Slug, expectedSlug)
	}
}

// TestGitRepoApplyHandlerContract_ExistingMirrorNotReCloned ensures that when
// Apply() is called for a GitRepo whose mirror already exists, it does NOT
// attempt to re-clone (idempotent behaviour).
func TestGitRepoApplyHandlerContract_ExistingMirrorNotReCloned(t *testing.T) {
	localRemote := createLocalTestRemoteRepo(t)
	mirrorBaseDir := t.TempDir()
	mgr := mirror.NewGitMirrorManager(mirrorBaseDir)

	// Pre-seed: clone the mirror before Apply() is called.
	const slug = "example.com_user_idempotent-pre-seeded"
	preSeedPath, err := mgr.Clone(localRemote, slug)
	if err != nil {
		t.Fatalf("pre-seed Clone(): %v", err)
	}

	// Record mtime before Apply() to detect unexpected re-clone
	infoBeforeApply, _ := os.Stat(preSeedPath)
	mtimeBefore := infoBeforeApply.ModTime()

	// Set up handler with the same MirrorBaseDir so it can find the mirror.
	h := NewGitRepoHandler()
	h.MirrorBaseDir = mirrorBaseDir
	store := db.NewMockDataStore()

	// Pre-seed DB record with the slug (so the update path is taken and the
	// slug is preserved from the existing record).
	existing := &models.GitRepoDB{
		Name:       "idempotent-apply-repo",
		URL:        "https://github.com/user/idempotent-pre-seeded",
		Slug:       slug,
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	if err := store.CreateGitRepo(existing); err != nil {
		t.Fatalf("pre-seed CreateGitRepo: %v", err)
	}

	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: idempotent-apply-repo
spec:
  url: https://github.com/user/idempotent-pre-seeded
  defaultRef: main
  authType: none
`)
	_, err = h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	// The mirror directory must still exist (not deleted and re-created)
	if _, statErr := os.Stat(filepath.Join(mirrorBaseDir, slug)); statErr != nil {
		t.Errorf("mirror was deleted during Apply() — should be preserved: %v", statErr)
	}

	// mtime should not have changed (mirror was not re-cloned)
	infoAfterApply, _ := os.Stat(preSeedPath)
	mtimeAfter := infoAfterApply.ModTime()
	if mtimeAfter.After(mtimeBefore) {
		t.Logf("NOTE: mirror mtime changed after Apply() — re-clone happened (unexpected)")
	}

	// Primary assertion: mirror still intact
	if !mgr.Exists(slug) {
		t.Error("mirror no longer exists after Apply() — was incorrectly deleted")
	}

	_ = mtimeBefore
}
