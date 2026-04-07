package handlers

// =============================================================================
// Workspace Apply Filesystem Tests — Issue #193
//
// These tests verify that:
//   1. workspace.CreateWorkspaceDirectories() works in isolation (infrastructure)
//   2. mirror.CloneToWorkspace() works in isolation (infrastructure)
//   3. WorkspaceHandler.Apply() creates the workspace directory tree and,
//      when a GitRepo is associated, clones the repo from the mirror (contract)
// =============================================================================

import (
	"os"
	"path/filepath"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/pkg/mirror"
	"devopsmaestro/pkg/workspace"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// -----------------------------------------------------------------------------
// Test 4: workspace.CreateWorkspaceDirectories() creates expected subdirs
// Expected: PASS — function already exists and works
// -----------------------------------------------------------------------------

func TestCreateWorkspaceDirectoriesCreatesRepoDir(t *testing.T) {
	tempBase := t.TempDir()
	wsPath := filepath.Join(tempBase, "workspaces", "eco-domain-app-dev")

	if err := workspace.CreateWorkspaceDirectories(wsPath); err != nil {
		t.Fatalf("CreateWorkspaceDirectories() error: %v", err)
	}

	// All required subdirectories must exist
	expectedDirs := []string{
		"repo",
		"volume",
		"volume/nvim-data",
		"volume/nvim-state",
		"volume/cache",
		".dvm",
		".dvm/nvim",
		".dvm/shell",
		".dvm/starship",
	}

	for _, rel := range expectedDirs {
		fullPath := filepath.Join(wsPath, rel)
		info, err := os.Stat(fullPath)
		if err != nil {
			t.Errorf("directory %q was not created: %v", rel, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("path %q exists but is not a directory", rel)
		}
	}
}

// -----------------------------------------------------------------------------
// Test 5: mirror.CloneToWorkspace() creates a working git checkout at destPath
// Expected: PASS — CloneToWorkspace already works end-to-end
// -----------------------------------------------------------------------------

func TestCloneToWorkspaceCreatesRepoCheckout(t *testing.T) {
	mgr, _ := newTestMirrorManager(t) // helper defined in gitrepo_apply_filesystem_test.go
	remoteURL := createLocalTestRemoteRepo(t)
	slug := "test.example.com_user_checkout-test"

	// Create the bare mirror first
	_, err := mgr.Clone(remoteURL, slug)
	if err != nil {
		t.Fatalf("Clone() error: %v", err)
	}

	destPath := filepath.Join(t.TempDir(), "workspace", "repo")

	// CloneToWorkspace — no specific ref needed for this test
	if err := mgr.CloneToWorkspace(slug, destPath, ""); err != nil {
		t.Fatalf("CloneToWorkspace() error: %v", err)
	}

	// destPath must exist and be a working (non-bare) git repo
	if _, err := os.Stat(destPath); err != nil {
		t.Fatalf("destination path does not exist after CloneToWorkspace(): %v", err)
	}

	dotGit := filepath.Join(destPath, ".git")
	if _, err := os.Stat(dotGit); err != nil {
		t.Errorf(".git directory missing — CloneToWorkspace did not produce a working checkout: %v", err)
	}

	// README.md from the remote must be present in the checkout
	readme := filepath.Join(destPath, "README.md")
	if _, err := os.Stat(readme); err != nil {
		t.Errorf("README.md missing in workspace checkout — content not cloned: %v", err)
	}
}

// =============================================================================
// Handler Contract Tests — Issue #193
// =============================================================================

// TestWorkspaceApplyHandlerContract_DirectoryShouldExistAfterApply verifies
// that after Apply() returns successfully, the workspace directory tree
// (repo/, volume/, .dvm/ etc.) exists on disk.
func TestWorkspaceApplyHandlerContract_DirectoryShouldExistAfterApply(t *testing.T) {
	// Arrange: full hierarchy in mock DB
	h := NewWorkspaceHandler()
	store, _, _, _ := setupWorkspaceTest(t) // helper from workspace_test.go

	// Inject the workspace base dir so the handler creates directories in
	// our temp dir instead of ~/.devopsmaestro/workspaces/.
	workspacesBaseDir := t.TempDir()
	h.WorkspacesBaseDir = workspacesBaseDir

	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: contract-ws
  app: ws-app
spec:
  image:
    name: ubuntu:22.04
`)

	// Act
	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() returned unexpected error: %v", err)
	}

	// Assert: DB record must exist (already passes today)
	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("result is not *WorkspaceResource")
	}
	wsModel := wr.Workspace()
	if wsModel.Name != "contract-ws" {
		t.Errorf("workspace Name = %q, want %q", wsModel.Name, "contract-ws")
	}

	// The slug uniquely identifies the workspace directory.
	if wsModel.Slug == "" {
		t.Fatal("workspace Slug is empty after Apply() — cannot compute expected directory path")
	}

	// Assert: workspace directory tree MUST exist after Apply().
	expectedWsPath := filepath.Join(workspacesBaseDir, wsModel.Slug)
	repoDir := filepath.Join(expectedWsPath, "repo")

	if _, statErr := os.Stat(repoDir); statErr != nil {
		t.Errorf(
			"workspace repo/ directory does not exist after Apply() — "+
				"handler must call workspace.CreateWorkspaceDirectories(path) during Apply().\n"+
				"Expected path: %s\n"+
				"Fix: wire workspace.CreateWorkspaceDirectories() into WorkspaceHandler.Apply() (Issue #193)",
			repoDir,
		)
	}
}

// TestWorkspaceApplyHandlerContract_SlugIsPopulated verifies that after Apply()
// the workspace model has a non-empty Slug. This is a prerequisite for knowing
// the on-disk path. Expected: PASS (slug generation already works).
func TestWorkspaceApplyHandlerContract_SlugIsPopulated(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, _ := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: slug-check-ws
  app: ws-app
spec:
  image:
    name: ubuntu:22.04
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("result is not *WorkspaceResource")
	}

	if wr.Workspace().Slug == "" {
		t.Error("workspace Slug is empty after Apply() — cannot compute directory path for filesystem init")
	}
}

// TestWorkspaceApplyHandlerContract_WithGitRepo_RepoCheckoutShouldExist
// verifies that when a Workspace has an associated GitRepo and the mirror
// exists, Apply() clones the repo from the mirror into the workspace repo/
// directory.
func TestWorkspaceApplyHandlerContract_WithGitRepo_RepoCheckoutShouldExist(t *testing.T) {
	store, _, _, _ := setupWorkspaceTest(t)

	// Create a GitRepo DB record with a known slug.
	const gitRepoHTTPSURL = "https://github.com/example/contract-ws-gitrepo"
	slug, slugErr := mirror.GenerateSlug(gitRepoHTTPSURL)
	if slugErr != nil {
		t.Fatalf("GenerateSlug: %v", slugErr)
	}

	gitRepoModel := &models.GitRepoDB{
		Name:       "contract-gitrepo",
		URL:        gitRepoHTTPSURL,
		Slug:       slug,
		DefaultRef: "main",
		AuthType:   "none",
		SyncStatus: "pending",
	}
	if err := store.CreateGitRepo(gitRepoModel); err != nil {
		t.Fatalf("CreateGitRepo: %v", err)
	}

	// Pre-seed the mirror using a local git repo as the actual source.
	localRemote := createLocalTestRemoteRepo(t)
	mirrorBaseDir := t.TempDir()
	mgr := mirror.NewGitMirrorManager(mirrorBaseDir)
	if _, err := mgr.Clone(localRemote, slug); err != nil {
		t.Fatalf("pre-seed mirror Clone(): %v", err)
	}

	workspacesBaseDir := t.TempDir()

	// Set up handler with injectable base dirs.
	h := NewWorkspaceHandler()
	h.WorkspacesBaseDir = workspacesBaseDir
	h.MirrorBaseDir = mirrorBaseDir

	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: gitrepo-contract-ws
  app: ws-app
spec:
  image:
    name: ubuntu:22.04
  gitrepo: contract-gitrepo
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error: %v", err)
	}

	wr := res.(*WorkspaceResource)
	wsModel := wr.Workspace()

	// GitRepoID must be set (already passes today)
	if !wsModel.GitRepoID.Valid {
		t.Error("GitRepoID not set after Apply() with gitrepo specified")
	}

	// workspace repo/ must contain a git checkout (.git directory).
	expectedRepoDir := filepath.Join(workspacesBaseDir, wsModel.Slug, "repo")
	dotGit := filepath.Join(expectedRepoDir, ".git")

	if _, statErr := os.Stat(dotGit); statErr != nil {
		t.Errorf(
			"workspace repo/.git does not exist after Apply() with associated GitRepo — "+
				"handler must call CloneToWorkspace() during Apply().\n"+
				"Expected path: %s\n"+
				"Fix: wire GitMirrorManager.CloneToWorkspace() into WorkspaceHandler.Apply() (Issue #193)",
			dotGit,
		)
	}
}
