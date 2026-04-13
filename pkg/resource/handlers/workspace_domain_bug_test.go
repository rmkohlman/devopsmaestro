package handlers

// =============================================================================
// TDD Phase 2 (RED): Bug #130 — Workspace export missing domainName and gitRepoName
//
// get_all.go line 195: handlers.NewWorkspaceResource(w, appNames[w.AppID])
// — only passes appName, no domainName or gitRepoName.
//
// Fix design: Expand NewWorkspaceResource signature to:
//   func NewWorkspaceResource(workspace *models.Workspace, appName, domainName, gitRepoName string) *WorkspaceResource
// And update get_all.go to pass precomputed domain/gitrepo names.
//
// These tests verify the round-trip requirement:
//   1. Workspace YAML includes metadata.domain when exported via NewWorkspaceResource
//   2. Workspace YAML includes spec.gitrepo when workspace has a linked git repo
//   3. export → import preserves domain and gitrepo associations
//   4. NewWorkspaceResource with new signature compiles and works correctly
//
// ALL tests in this section WILL FAIL until:
//   - NewWorkspaceResource signature is expanded to (workspace, appName, domainName, gitRepoName)
//   - get_all.go is updated to pass resolved domainName and gitRepoName
// =============================================================================

import (
	"database/sql"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"
)

// TestNewWorkspaceResource_WithDomainName verifies that NewWorkspaceResource
// with the new expanded signature correctly populates domainName so that
// ToYAML() includes metadata.domain.
//
// RED: Fails because NewWorkspaceResource(workspace, appName) only takes 2
// args — passing domainName requires the new 4-arg signature.
func TestNewWorkspaceResource_WithDomainName(t *testing.T) {
	h := NewWorkspaceHandler()

	ws := &models.Workspace{
		ID:        1,
		AppID:     1,
		Name:      "ws-with-domain",
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}

	// ── COMPILE ERROR EXPECTED BELOW — current signature is (workspace, appName) ──
	// New signature: NewWorkspaceResource(workspace, appName, domainName, gitRepoName)
	res := NewWorkspaceResource(ws, "my-app", "my-domain", "") // COMPILE ERROR: too many arguments
	// ────────────────────────────────────────────────────────────────────────────

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	if !strings.Contains(yamlStr, "domain: my-domain") {
		t.Errorf("ToYAML() must include 'domain: my-domain' in metadata; got:\n%s", yamlStr)
	}
}

// TestNewWorkspaceResource_WithGitRepoName verifies that NewWorkspaceResource
// with the new signature correctly populates gitRepoName so that ToYAML()
// includes spec.gitrepo.
//
// RED: Same compile error — current signature doesn't accept gitRepoName.
func TestNewWorkspaceResource_WithGitRepoName(t *testing.T) {
	h := NewWorkspaceHandler()

	ws := &models.Workspace{
		ID:        1,
		AppID:     1,
		Name:      "ws-with-gitrepo",
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Int64: 5, Valid: true},
	}

	// ── COMPILE ERROR EXPECTED BELOW — too many arguments ──────────────────────
	res := NewWorkspaceResource(ws, "my-app", "my-domain", "my-repo") // COMPILE ERROR
	// ────────────────────────────────────────────────────────────────────────────

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	if !strings.Contains(yamlStr, "gitrepo: my-repo") {
		t.Errorf("ToYAML() must include 'gitrepo: my-repo' in spec; got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "domain: my-domain") {
		t.Errorf("ToYAML() must include 'domain: my-domain' in metadata; got:\n%s", yamlStr)
	}
}

// TestNewWorkspaceResource_BackwardsCompatibility verifies that the current
// 2-arg call site still works when called with empty domainName and gitRepoName.
// This test verifies the EXISTING behavior (ToYAML without domain/gitrepo)
// and documents that it should remain valid after the signature change.
//
// GREEN today (tests existing behavior). After signature change, callers
// must be updated to pass "" for the new params — this test should still pass.
func TestNewWorkspaceResource_BackwardsCompatibility(t *testing.T) {
	h := NewWorkspaceHandler()

	ws := &models.Workspace{
		ID:        1,
		AppID:     1,
		Name:      "compat-ws",
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}

	// Current 2-arg call — this is what get_all.go does today (the bug)
	// After fix, this should become NewWorkspaceResource(ws, "my-app", "", "")
	res := NewWorkspaceResource(ws, "my-app", "", "")

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// Basic fields must be present
	if !strings.Contains(yamlStr, "name: compat-ws") {
		t.Errorf("ToYAML() missing 'name: compat-ws'; got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "app: my-app") {
		t.Errorf("ToYAML() missing 'app: my-app'; got:\n%s", yamlStr)
	}
	// Without domain provided, domain should be absent (no empty string in YAML)
	// This is existing behavior — no change needed here
}

// TestWorkspace_RoundTrip_DomainPreserved verifies the full export→apply
// round-trip preserves the domain association: a workspace exported with
// domain info in metadata can be re-applied without active context.
//
// RED: Fails because NewWorkspaceResource(w, appName) in get_all.go doesn't
// pass domainName, so exported YAML lacks metadata.domain.
func TestWorkspace_RoundTrip_DomainPreserved(t *testing.T) {
	h := NewWorkspaceHandler()

	// --- Source store: full hierarchy ---
	sourceStore := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "ws-rt-eco"}
	if err := sourceStore.CreateEcosystem(eco); err != nil {
		t.Fatalf("CreateEcosystem: %v", err)
	}
	domain := &models.Domain{Name: "ws-rt-domain", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	if err := sourceStore.CreateDomain(domain); err != nil {
		t.Fatalf("CreateDomain: %v", err)
	}
	app := &models.App{Name: "ws-rt-app", DomainID: sql.NullInt64{Int64: int64(domain.ID), Valid: true}, Path: "/rt"}
	if err := sourceStore.CreateApp(app); err != nil {
		t.Fatalf("CreateApp: %v", err)
	}
	ws := &models.Workspace{
		Name:      "ws-rt",
		AppID:     app.ID,
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	if err := sourceStore.CreateWorkspace(ws); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Export: simulates what get_all.go SHOULD do after fix
	// RED: NewWorkspaceResource(ws, appNames[ws.AppID]) misses domainName/gitRepoName
	// ── COMPILE ERROR EXPECTED BELOW — too many arguments ──────────────────────
	exportRes := NewWorkspaceResource(ws, "ws-rt-app", "ws-rt-domain", "") // COMPILE ERROR
	// ────────────────────────────────────────────────────────────────────────────

	yamlBytes, err := h.ToYAML(exportRes)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// Must include domain in metadata for context-free apply
	if !strings.Contains(yamlStr, "domain: ws-rt-domain") {
		t.Errorf("exported Workspace YAML must contain 'domain: ws-rt-domain'; got:\n%s", yamlStr)
	}

	// --- Destination store: fresh, no active context ---
	destStore := db.NewMockDataStore()
	dstEco := &models.Ecosystem{Name: "ws-rt-eco"}
	if err := destStore.CreateEcosystem(dstEco); err != nil {
		t.Fatalf("dest CreateEcosystem: %v", err)
	}
	dstDomain := &models.Domain{Name: "ws-rt-domain", EcosystemID: sql.NullInt64{Int64: int64(dstEco.ID), Valid: true}}
	if err := destStore.CreateDomain(dstDomain); err != nil {
		t.Fatalf("dest CreateDomain: %v", err)
	}
	dstApp := &models.App{Name: "ws-rt-app", DomainID: sql.NullInt64{Int64: int64(dstDomain.ID), Valid: true}, Path: "/rt"}
	if err := destStore.CreateApp(dstApp); err != nil {
		t.Fatalf("dest CreateApp: %v", err)
	}

	// Apply with no active context — metadata.domain should resolve the app
	destCtx := resource.Context{DataStore: destStore}
	_, err = h.Apply(destCtx, yamlBytes)
	if err != nil {
		t.Fatalf("Apply() on fresh store should succeed with metadata.domain in YAML, got: %v", err)
	}
}

// TestWorkspace_RoundTrip_GitRepoPreserved verifies that a workspace with a
// linked git repo exports spec.gitrepo and can be re-applied preserving the link.
//
// RED: NewWorkspaceResource(w, appName) doesn't pass gitRepoName, so exported
// YAML lacks spec.gitrepo.
func TestWorkspace_RoundTrip_GitRepoPreserved(t *testing.T) {
	h := NewWorkspaceHandler()

	sourceStore := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "ws-git-eco"}
	if err := sourceStore.CreateEcosystem(eco); err != nil {
		t.Fatalf("CreateEcosystem: %v", err)
	}
	domain := &models.Domain{Name: "ws-git-domain", EcosystemID: sql.NullInt64{Int64: int64(eco.ID), Valid: true}}
	if err := sourceStore.CreateDomain(domain); err != nil {
		t.Fatalf("CreateDomain: %v", err)
	}
	app := &models.App{Name: "ws-git-app", DomainID: sql.NullInt64{Int64: int64(domain.ID), Valid: true}, Path: "/git"}
	if err := sourceStore.CreateApp(app); err != nil {
		t.Fatalf("CreateApp: %v", err)
	}
	gitRepo := &models.GitRepoDB{Name: "api-repo", URL: "https://github.com/org/api.git"}
	if err := sourceStore.CreateGitRepo(gitRepo); err != nil {
		t.Fatalf("CreateGitRepo: %v", err)
	}
	ws := &models.Workspace{
		Name:      "ws-git",
		AppID:     app.ID,
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
		GitRepoID: sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true},
	}
	if err := sourceStore.CreateWorkspace(ws); err != nil {
		t.Fatalf("CreateWorkspace: %v", err)
	}

	// Export with gitRepoName — requires new 4-arg signature
	// ── COMPILE ERROR EXPECTED BELOW — too many arguments ──────────────────────
	exportRes := NewWorkspaceResource(ws, "ws-git-app", "ws-git-domain", "api-repo") // COMPILE ERROR
	// ────────────────────────────────────────────────────────────────────────────

	yamlBytes, err := h.ToYAML(exportRes)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	if !strings.Contains(yamlStr, "gitrepo: api-repo") {
		t.Errorf("exported Workspace YAML must contain 'gitrepo: api-repo'; got:\n%s", yamlStr)
	}

	// Reimport: fresh store with both the git repo and hierarchy present
	destStore := db.NewMockDataStore()
	dstEco := &models.Ecosystem{Name: "ws-git-eco"}
	if err := destStore.CreateEcosystem(dstEco); err != nil {
		t.Fatalf("dest CreateEcosystem: %v", err)
	}
	dstDomain := &models.Domain{Name: "ws-git-domain", EcosystemID: sql.NullInt64{Int64: int64(dstEco.ID), Valid: true}}
	if err := destStore.CreateDomain(dstDomain); err != nil {
		t.Fatalf("dest CreateDomain: %v", err)
	}
	dstApp := &models.App{Name: "ws-git-app", DomainID: sql.NullInt64{Int64: int64(dstDomain.ID), Valid: true}, Path: "/git"}
	if err := destStore.CreateApp(dstApp); err != nil {
		t.Fatalf("dest CreateApp: %v", err)
	}
	dstRepo := &models.GitRepoDB{Name: "api-repo", URL: "https://github.com/org/api.git"}
	if err := destStore.CreateGitRepo(dstRepo); err != nil {
		t.Fatalf("dest CreateGitRepo: %v", err)
	}

	destCtx := resource.Context{DataStore: destStore}
	res, err := h.Apply(destCtx, yamlBytes)
	if err != nil {
		t.Fatalf("Apply() on fresh store with gitrepo should succeed, got: %v", err)
	}

	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("result is not *WorkspaceResource")
	}

	if !wr.Workspace().GitRepoID.Valid {
		t.Error("Apply() should have set GitRepoID from spec.gitrepo in YAML")
	}
	if wr.Workspace().GitRepoID.Int64 != int64(dstRepo.ID) {
		t.Errorf("Apply() GitRepoID = %d, want %d", wr.Workspace().GitRepoID.Int64, dstRepo.ID)
	}
}
