package handlers

import (
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
)

// =============================================================================
// GitRepoHandler Tests - Kind
// =============================================================================

func TestGitRepoHandler_Kind(t *testing.T) {
	h := NewGitRepoHandler()
	if h.Kind() != KindGitRepo {
		t.Errorf("Kind() = %q, want %q", h.Kind(), KindGitRepo)
	}
}

// =============================================================================
// GitRepoHandler Tests - Apply
// =============================================================================

func TestGitRepoHandler_Apply_Create(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: test-repo
  annotations:
    description: A test git repository
spec:
  url: https://github.com/user/test-repo
  defaultRef: main
  authType: none
  autoSync: true
  syncIntervalMinutes: 30
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if res.GetKind() != KindGitRepo {
		t.Errorf("Apply() resource.Kind = %q, want %q", res.GetKind(), KindGitRepo)
	}
	if res.GetName() != "test-repo" {
		t.Errorf("Apply() resource.Name = %q, want %q", res.GetName(), "test-repo")
	}

	// Verify stored in mock
	stored, err := store.GetGitRepoByName("test-repo")
	if err != nil {
		t.Fatalf("git repo not found in store: %v", err)
	}
	if stored.Name != "test-repo" {
		t.Errorf("stored Name = %q, want %q", stored.Name, "test-repo")
	}
	if stored.URL != "https://github.com/user/test-repo" {
		t.Errorf("stored URL = %q, want %q", stored.URL, "https://github.com/user/test-repo")
	}
}

func TestGitRepoHandler_Apply_Update(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	// Pre-populate
	_ = store.CreateGitRepo(&models.GitRepoDB{
		Name: "repo-update",
		URL:  "https://github.com/user/repo-update",
	})

	updateYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: repo-update
spec:
  url: https://github.com/user/repo-update-changed
  defaultRef: develop
  authType: none
`)

	res, err := h.Apply(ctx, updateYAML)
	if err != nil {
		t.Fatalf("Apply() update error = %v", err)
	}
	if res.GetName() != "repo-update" {
		t.Errorf("Apply() resource.Name = %q, want %q", res.GetName(), "repo-update")
	}

	// Verify URL was updated
	repoRes, ok := res.(*GitRepoResource)
	if !ok {
		t.Fatalf("result is not *GitRepoResource")
	}
	if repoRes.GitRepo().URL != "https://github.com/user/repo-update-changed" {
		t.Errorf("Apply() URL = %q, want %q", repoRes.GitRepo().URL, "https://github.com/user/repo-update-changed")
	}
}

func TestGitRepoHandler_Apply_InvalidYAML(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_, err := h.Apply(ctx, []byte("{\x00invalid: [yaml"))
	if err == nil {
		t.Error("Apply() expected error for invalid YAML, got nil")
	}
}

// =============================================================================
// GitRepoHandler Tests - Get
// =============================================================================

func TestGitRepoHandler_Get_Found(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_ = store.CreateGitRepo(&models.GitRepoDB{
		Name: "repo-get",
		URL:  "https://github.com/user/repo-get",
	})

	res, err := h.Get(ctx, "repo-get")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if res.GetName() != "repo-get" {
		t.Errorf("Get() Name = %q, want %q", res.GetName(), "repo-get")
	}
	if res.GetKind() != KindGitRepo {
		t.Errorf("Get() Kind = %q, want %q", res.GetKind(), KindGitRepo)
	}
}

func TestGitRepoHandler_Get_NotFound(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_, err := h.Get(ctx, "does-not-exist")
	if err == nil {
		t.Error("Get() expected error for non-existent git repo, got nil")
	}
}

// =============================================================================
// GitRepoHandler Tests - List
// =============================================================================

func TestGitRepoHandler_List_Empty(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("List() returned %d items, want 0", len(resources))
	}
}

func TestGitRepoHandler_List_Multiple(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_ = store.CreateGitRepo(&models.GitRepoDB{Name: "repo-a", URL: "https://github.com/user/repo-a"})
	_ = store.CreateGitRepo(&models.GitRepoDB{Name: "repo-b", URL: "https://github.com/user/repo-b"})
	_ = store.CreateGitRepo(&models.GitRepoDB{Name: "repo-c", URL: "https://github.com/user/repo-c"})

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 3 {
		t.Errorf("List() returned %d items, want 3", len(resources))
	}
	for _, r := range resources {
		if r.GetKind() != KindGitRepo {
			t.Errorf("List() item Kind = %q, want %q", r.GetKind(), KindGitRepo)
		}
	}
}

// =============================================================================
// GitRepoHandler Tests - Delete
// =============================================================================

func TestGitRepoHandler_Delete_Found(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_ = store.CreateGitRepo(&models.GitRepoDB{
		Name: "repo-delete",
		URL:  "https://github.com/user/repo-delete",
	})

	if err := h.Delete(ctx, "repo-delete"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify removed
	_, err := store.GetGitRepoByName("repo-delete")
	if err == nil {
		t.Error("Delete() did not remove git repo from store")
	}
}

func TestGitRepoHandler_Delete_NotFound(t *testing.T) {
	h := NewGitRepoHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	err := h.Delete(ctx, "nonexistent")
	if err == nil {
		t.Error("Delete() expected error for non-existent git repo, got nil")
	}
}

// =============================================================================
// GitRepoHandler Tests - ToYAML
// =============================================================================

func TestGitRepoHandler_ToYAML(t *testing.T) {
	h := NewGitRepoHandler()

	repo := &models.GitRepoDB{
		ID:   1,
		Name: "yaml-repo",
		URL:  "https://github.com/user/yaml-repo",
	}
	res := NewGitRepoResource(repo)

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	if !strings.Contains(yamlStr, "kind: GitRepo") {
		t.Errorf("ToYAML() missing 'kind: GitRepo', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "apiVersion: devopsmaestro.io/v1") {
		t.Errorf("ToYAML() missing 'apiVersion: devopsmaestro.io/v1', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "name: yaml-repo") {
		t.Errorf("ToYAML() missing 'name: yaml-repo', got:\n%s", yamlStr)
	}
}

func TestGitRepoHandler_ToYAML_WrongType(t *testing.T) {
	h := NewGitRepoHandler()
	// Pass a wrong resource type
	wrongRes := &EcosystemResource{ecosystem: &models.Ecosystem{Name: "wrong"}}
	_, err := h.ToYAML(wrongRes)
	if err == nil {
		t.Error("ToYAML() expected error for wrong resource type, got nil")
	}
}

// =============================================================================
// GitRepoResource Validate Tests
// =============================================================================

func TestGitRepoResource_Validate(t *testing.T) {
	tests := []struct {
		name    string
		repo    *models.GitRepoDB
		wantErr bool
	}{
		{
			name:    "valid git repo",
			repo:    &models.GitRepoDB{Name: "valid-repo", URL: "https://github.com/user/valid-repo"},
			wantErr: false,
		},
		{
			name:    "missing name",
			repo:    &models.GitRepoDB{Name: "", URL: "https://github.com/user/repo"},
			wantErr: true,
		},
		{
			name:    "missing URL",
			repo:    &models.GitRepoDB{Name: "no-url-repo", URL: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewGitRepoResource(tt.repo)
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
