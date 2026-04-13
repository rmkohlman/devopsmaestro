package handlers

// =============================================================================
// App-GitRepo Round-Trip Tests — Issue #175
//
// These tests prove the bug: App-GitRepo associations are lost on YAML
// export → wipe → restore because:
//
//  1. AppSpec has no GitRepo field
//  2. AppHandler.ToYAML() never resolves GitRepoID → name
//  3. AppHandler.Apply() never resolves a git repo name → ID
//
// All tests MUST compile but MUST fail (red phase of TDD) until the fix is in.
// Compare to WorkspaceHandler which correctly implements this pattern.
// =============================================================================

import (
	"database/sql"
	"strings"
	"testing"

	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/resource"

	"gopkg.in/yaml.v3"
)

// =============================================================================
// TestAppHandler_ToYAML_IncludesGitRepo
//
// Arrange: Create an App model with GitRepoID set. Wrap it in an AppResource
//
//	(the way the handler returns it today).
//
// Act:     Call AppHandler.ToYAML().
//
// Assert:  The resulting YAML contains a "gitRepo" or "gitrepo" field with
//
//	the resolved git repo name.
//
// WHY IT FAILS: AppResource has no gitRepoName field, AppHandler.ToYAML()
//
//	passes nil to app.ToYAML(), and AppSpec has no GitRepo field.
//	The YAML output will never contain a git repo reference.
//
// =============================================================================
func TestAppHandler_ToYAML_IncludesGitRepo(t *testing.T) {
	h := NewAppHandler()

	app := &models.App{
		ID:       1,
		DomainID: sql.NullInt64{Int64: 1, Valid: true},
		Name:     "git-app",
		Path:     "/my/app",
		// GitRepoID is set — simulates an app created with `dvm create app --repo`
		GitRepoID: sql.NullInt64{Int64: 42, Valid: true},
	}

	// A fixed implementation carries the resolved git repo name in
	// AppResource (mirroring WorkspaceResource.gitRepoName).
	// NewAppResource now accepts gitRepoName as an optional extra parameter.
	res := NewAppResource(app, "my-domain", "my-eco", "my-git-repo")

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() unexpected error = %v", err)
	}
	yamlStr := string(yamlBytes)

	// MUST FAIL: no gitRepo field will appear in the output because
	// the current implementation never resolves GitRepoID → name.
	if !strings.Contains(yamlStr, "gitRepo:") && !strings.Contains(yamlStr, "gitrepo:") {
		t.Errorf("ToYAML() output is missing 'gitRepo' field — the git repo association will be lost on export/restore.\n"+
			"App.GitRepoID was set to 42 but was not serialised to YAML.\n"+
			"Got YAML:\n%s", yamlStr)
	}
}

// =============================================================================
// TestAppHandler_FromYAML_ReadsGitRepo
//
// Arrange: Construct raw YAML bytes that include spec.gitRepo.
//
// Act:     Unmarshal into models.AppYAML using gopkg.in/yaml.v3.
//
// Assert:  The parsed AppSpec retains the git repo name in a GitRepo field.
//
// WHY IT FAILS: AppSpec has no GitRepo field. The yaml decoder will silently
//
//	discard the spec.gitRepo key. The test proves the field is absent.
//
// =============================================================================
func TestAppHandler_FromYAML_ReadsGitRepo(t *testing.T) {
	// This is the YAML that a fixed exporter would produce.
	// The spec.gitRepo key must survive a round-trip through AppYAML.
	rawYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: repo-app
  domain: test-domain
spec:
  path: /app/path
  gitRepo: my-git-repo
`)

	var parsed models.AppYAML
	if err := yaml.Unmarshal(rawYAML, &parsed); err != nil {
		t.Fatalf("yaml.Unmarshal() unexpected error = %v", err)
	}

	// MUST FAIL: AppSpec has no GitRepo field.
	// The yaml decoder silently discards unknown fields, so the git repo name
	// is lost during unmarshalling. There is no way to read it back.
	//
	// A fix must add `GitRepo string \`yaml:"gitRepo,omitempty"\`` to AppSpec,
	// then this assertion will start passing.
	//
	// We detect the absence by checking that the spec path did parse (confirming
	// the YAML was valid and parsed correctly) but no git repo field exists.
	if parsed.Spec.Path != "/app/path" {
		t.Fatalf("sanity check failed: Spec.Path = %q, want %q — YAML was not parsed at all", parsed.Spec.Path, "/app/path")
	}

	// AppSpec currently has no GitRepo field at all, so we cannot reference
	// parsed.Spec.GitRepo directly — it would be a compile error.
	// Instead we re-marshal the parsed spec to check whether the gitRepo key
	// survived. This is equivalent and compiles against the current (broken) code.
	reMarshalled, err := yaml.Marshal(parsed.Spec)
	if err != nil {
		t.Fatalf("yaml.Marshal(parsed.Spec) unexpected error = %v", err)
	}
	reStr := string(reMarshalled)

	// MUST FAIL: the re-marshalled spec will not contain "gitRepo" because
	// AppSpec has no such field — the value was discarded on unmarshal.
	if !strings.Contains(reStr, "gitRepo") && !strings.Contains(reStr, "gitrepo") {
		t.Errorf("FromYAML: spec.gitRepo was silently discarded — AppSpec has no GitRepo field.\n"+
			"Input YAML contained 'gitRepo: my-git-repo' but it was lost.\n"+
			"Re-marshalled AppSpec:\n%s", reStr)
	}
}

// =============================================================================
// TestAppHandler_Apply_ResolvesGitRepoName
//
// Arrange: Pre-populate the mock store with a git repo named "my-git-repo".
//
//	Provide App YAML that references spec.gitRepo: my-git-repo.
//
// Act:     Call AppHandler.Apply().
//
// Assert:  The resulting AppResource's App has GitRepoID set to the ID of the
//
//	"my-git-repo" record.
//
// WHY IT FAILS: AppSpec has no GitRepo field, so the name is never parsed.
//
//	Even if it were parsed, Apply() has no code to resolve it to an ID.
//	The resulting App.GitRepoID will be the zero value (not valid).
//
// =============================================================================
func TestAppHandler_Apply_ResolvesGitRepoName(t *testing.T) {
	h := NewAppHandler()
	store, _, domainID := setupAppTest(t)
	ctx := resource.Context{DataStore: store}

	// Pre-populate a git repo
	gitRepo := &models.GitRepoDB{
		Name: "my-git-repo",
		URL:  "https://github.com/user/my-git-repo",
	}
	if err := store.CreateGitRepo(gitRepo); err != nil {
		t.Fatalf("CreateGitRepo() unexpected error = %v", err)
	}

	// YAML with spec.gitRepo — this is the contract the fix must honour
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: linked-app
  domain: app-domain
spec:
  path: /linked/app
  gitRepo: my-git-repo
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() unexpected error = %v", err)
	}

	ar, ok := res.(*AppResource)
	if !ok {
		t.Fatalf("Apply() result is not *AppResource, got %T", res)
	}

	// MUST FAIL: Apply() never reads spec.gitRepo, so GitRepoID is never set.
	if !ar.App().GitRepoID.Valid {
		t.Errorf("Apply(): App.GitRepoID is not set — spec.gitRepo was not resolved to a GitRepoID FK.\n"+
			"Expected GitRepoID = %d (ID of 'my-git-repo'), got zero value.\n"+
			"App struct: %+v", gitRepo.ID, ar.App())
	} else if ar.App().GitRepoID.Int64 != int64(gitRepo.ID) {
		t.Errorf("Apply(): App.GitRepoID = %d, want %d (ID of 'my-git-repo')",
			ar.App().GitRepoID.Int64, gitRepo.ID)
	}

	// Also verify the association persisted in the store
	storedApp, err := store.GetAppByName(sql.NullInt64{Int64: int64(domainID), Valid: true}, "linked-app")
	if err != nil {
		t.Fatalf("GetAppByName() after Apply() error = %v", err)
	}
	if !storedApp.GitRepoID.Valid {
		t.Errorf("stored App.GitRepoID is not set — git repo association was not persisted to the DB")
	} else if storedApp.GitRepoID.Int64 != int64(gitRepo.ID) {
		t.Errorf("stored App.GitRepoID = %d, want %d", storedApp.GitRepoID.Int64, gitRepo.ID)
	}
}

// =============================================================================
// TestAppHandler_ToYAML_RoundTrip_PreservesGitRepo
//
// Table-driven end-to-end round-trip test.
// Proves the complete export path is broken for apps with git repo associations.
//
// WHY IT FAILS: ToYAML() never includes the git repo reference, so the
//
//	round-trip (export → apply) would silently drop the association.
//
// =============================================================================
func TestAppHandler_ToYAML_RoundTrip_PreservesGitRepo(t *testing.T) {
	tests := []struct {
		name        string
		gitRepoName string
		gitRepoURL  string
	}{
		{
			name:        "github repo",
			gitRepoName: "acme-backend",
			gitRepoURL:  "https://github.com/acme/backend",
		},
		{
			name:        "gitlab repo",
			gitRepoName: "internal-api",
			gitRepoURL:  "https://gitlab.internal/org/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewAppHandler()
			store, _, _ := setupAppTest(t)

			// Create the git repo
			gitRepo := &models.GitRepoDB{Name: tt.gitRepoName, URL: tt.gitRepoURL}
			if err := store.CreateGitRepo(gitRepo); err != nil {
				t.Fatalf("CreateGitRepo() error = %v", err)
			}

			// Simulate an App that was created with `dvm create app --repo`
			// and therefore has GitRepoID set in the DB.
			app := &models.App{
				ID:        1,
				Name:      "roundtrip-app",
				DomainID:  sql.NullInt64{Int64: 1, Valid: true},
				Path:      "/rt/app",
				GitRepoID: sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true},
			}

			// A fixed implementation: Get() resolves GitRepoID → name and
			// stores it in AppResource. NewAppResource now accepts gitRepoName
			// as an optional extra parameter, mirroring the workspace pattern.
			appRes := NewAppResource(app, "app-domain", "app-eco", tt.gitRepoName)

			// Step: Export to YAML
			yamlBytes, err := h.ToYAML(appRes)
			if err != nil {
				t.Fatalf("ToYAML() error = %v", err)
			}
			yamlStr := string(yamlBytes)

			// MUST FAIL: no gitRepo field in output
			if !strings.Contains(yamlStr, "gitRepo:") && !strings.Contains(yamlStr, "gitrepo:") {
				t.Errorf("[%s] ToYAML() is missing gitRepo field — association will be silently lost on restore.\n"+
					"App had GitRepoID=%d (%s) but it was not serialised.\nYAML:\n%s",
					tt.name, gitRepo.ID, tt.gitRepoName, yamlStr)
			}
			if !strings.Contains(yamlStr, tt.gitRepoName) {
				t.Errorf("[%s] ToYAML() YAML does not contain git repo name %q.\nYAML:\n%s",
					tt.name, tt.gitRepoName, yamlStr)
			}
		})
	}
}
