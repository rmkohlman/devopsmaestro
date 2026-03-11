package handlers

import (
	"database/sql"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/pkg/resource"
)

// =============================================================================
// Helpers for workspace tests
// =============================================================================

// setupWorkspaceTest creates a mock store pre-populated with an ecosystem,
// domain, and app, and sets the active IDs on the context.
// Returns the store, ecosystemID, domainID, and appID.
func setupWorkspaceTest(t *testing.T) (*db.MockDataStore, int, int, int) {
	t.Helper()
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "ws-eco"}
	if err := store.CreateEcosystem(eco); err != nil {
		t.Fatalf("failed to create ecosystem: %v", err)
	}

	domain := &models.Domain{Name: "ws-domain", EcosystemID: eco.ID}
	if err := store.CreateDomain(domain); err != nil {
		t.Fatalf("failed to create domain: %v", err)
	}

	app := &models.App{Name: "ws-app", DomainID: domain.ID, Path: "/ws/app"}
	if err := store.CreateApp(app); err != nil {
		t.Fatalf("failed to create app: %v", err)
	}

	ecoID := eco.ID
	domainID := domain.ID
	appID := app.ID
	store.Context.ActiveEcosystemID = &ecoID
	store.Context.ActiveDomainID = &domainID
	store.Context.ActiveAppID = &appID

	return store, ecoID, domainID, appID
}

// =============================================================================
// WorkspaceHandler Tests - Kind
// =============================================================================

func TestWorkspaceHandler_Kind(t *testing.T) {
	h := NewWorkspaceHandler()
	if h.Kind() != KindWorkspace {
		t.Errorf("Kind() = %q, want %q", h.Kind(), KindWorkspace)
	}
}

// =============================================================================
// WorkspaceHandler Tests - Apply
// =============================================================================

func TestWorkspaceHandler_Apply_Create(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, _ := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: my-workspace
  app: ws-app
spec:
  image:
    name: ubuntu:22.04
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if res.GetKind() != KindWorkspace {
		t.Errorf("Apply() Kind = %q, want %q", res.GetKind(), KindWorkspace)
	}
	if res.GetName() != "my-workspace" {
		t.Errorf("Apply() Name = %q, want %q", res.GetName(), "my-workspace")
	}

	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("result is not *WorkspaceResource")
	}
	if wr.AppName() != "ws-app" {
		t.Errorf("AppName() = %q, want %q", wr.AppName(), "ws-app")
	}
	if wr.Workspace().ImageName != "ubuntu:22.04" {
		t.Errorf("Workspace().ImageName = %q, want %q", wr.Workspace().ImageName, "ubuntu:22.04")
	}
}

func TestWorkspaceHandler_Apply_Update(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, appID := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	// Pre-populate a workspace
	existing := &models.Workspace{
		Name:        "upd-workspace",
		AppID:       appID,
		ImageName:   "old-image:latest",
		Status:      "stopped",
		Slug:        "preserved-slug",
		ContainerID: sql.NullString{String: "old-container-id", Valid: true},
	}
	if err := store.CreateWorkspace(existing); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	updateYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: upd-workspace
  app: ws-app
spec:
  image:
    name: new-image:2.0
`)

	res, err := h.Apply(ctx, updateYAML)
	if err != nil {
		t.Fatalf("Apply() update error = %v", err)
	}
	if res.GetName() != "upd-workspace" {
		t.Errorf("Apply() Name = %q, want %q", res.GetName(), "upd-workspace")
	}

	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("result is not *WorkspaceResource")
	}
	if wr.Workspace().ImageName != "new-image:2.0" {
		t.Errorf("Apply() ImageName = %q, want %q", wr.Workspace().ImageName, "new-image:2.0")
	}
}

func TestWorkspaceHandler_Apply_PreservesFields(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, appID := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	// Pre-populate a workspace with rich fields
	existing := &models.Workspace{
		Name:            "preserve-ws",
		AppID:           appID,
		ImageName:       "base:1.0",
		Status:          "running",
		Slug:            "my-unique-slug",
		ContainerID:     sql.NullString{String: "container-abc", Valid: true},
		NvimStructure:   sql.NullString{String: "lazyvim", Valid: true},
		NvimPlugins:     sql.NullString{String: "plugin1,plugin2", Valid: true},
		TerminalPrompt:  sql.NullString{String: "starship", Valid: true},
		TerminalPackage: sql.NullString{String: "my-pkg", Valid: true},
	}
	if err := store.CreateWorkspace(existing); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	// Apply YAML without specifying nvim/terminal fields
	updateYAML := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: preserve-ws
  app: ws-app
spec:
  image:
    name: updated:2.0
`)

	res, err := h.Apply(ctx, updateYAML)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("result is not *WorkspaceResource")
	}

	ws := wr.Workspace()
	// Slug and ContainerID must be preserved
	if ws.Slug != "my-unique-slug" {
		t.Errorf("Slug not preserved: got %q, want %q", ws.Slug, "my-unique-slug")
	}
	if ws.ContainerID.String != "container-abc" {
		t.Errorf("ContainerID not preserved: got %q, want %q", ws.ContainerID.String, "container-abc")
	}
	// Optional fields preserved when not provided in YAML
	if ws.NvimStructure.String != "lazyvim" {
		t.Errorf("NvimStructure not preserved: got %q, want %q", ws.NvimStructure.String, "lazyvim")
	}
	if ws.NvimPlugins.String != "plugin1,plugin2" {
		t.Errorf("NvimPlugins not preserved: got %q, want %q", ws.NvimPlugins.String, "plugin1,plugin2")
	}
	if ws.TerminalPrompt.String != "starship" {
		t.Errorf("TerminalPrompt not preserved: got %q, want %q", ws.TerminalPrompt.String, "starship")
	}
	if ws.TerminalPackage.String != "my-pkg" {
		t.Errorf("TerminalPackage not preserved: got %q, want %q", ws.TerminalPackage.String, "my-pkg")
	}
}

func TestWorkspaceHandler_Apply_WithGitRepo(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, _ := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	// Pre-populate a gitrepo
	gitRepo := &models.GitRepoDB{Name: "my-repo", URL: "https://github.com/user/repo"}
	if err := store.CreateGitRepo(gitRepo); err != nil {
		t.Fatalf("failed to create git repo: %v", err)
	}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: repo-workspace
  app: ws-app
spec:
  image:
    name: ubuntu:22.04
  gitrepo: my-repo
`)

	res, err := h.Apply(ctx, yamlData)
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	wr, ok := res.(*WorkspaceResource)
	if !ok {
		t.Fatalf("result is not *WorkspaceResource")
	}
	if !wr.Workspace().GitRepoID.Valid {
		t.Error("Apply() GitRepoID should be set when gitrepo is specified")
	}
	if wr.Workspace().GitRepoID.Int64 != int64(gitRepo.ID) {
		t.Errorf("Apply() GitRepoID = %d, want %d", wr.Workspace().GitRepoID.Int64, gitRepo.ID)
	}
}

func TestWorkspaceHandler_Apply_MissingApp(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, _ := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	// YAML missing app metadata
	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: orphan-ws
spec:
  image:
    name: ubuntu:22.04
`)

	_, err := h.Apply(ctx, yamlData)
	if err == nil {
		t.Error("Apply() expected error for missing app, got nil")
	}
	if !strings.Contains(err.Error(), "app") {
		t.Errorf("Apply() error = %q, want it to mention 'app'", err.Error())
	}
}

func TestWorkspaceHandler_Apply_NoActiveDomain(t *testing.T) {
	h := NewWorkspaceHandler()
	// Store with no active domain
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	yamlData := []byte(`
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: any-ws
  app: some-app
spec:
  image:
    name: ubuntu:22.04
`)

	_, err := h.Apply(ctx, yamlData)
	if err == nil {
		t.Error("Apply() expected error when no active domain, got nil")
	}
	if !strings.Contains(err.Error(), "active domain") {
		t.Errorf("Apply() error = %q, want it to mention 'active domain'", err.Error())
	}
}

func TestWorkspaceHandler_Apply_InvalidYAML(t *testing.T) {
	h := NewWorkspaceHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_, err := h.Apply(ctx, []byte("{\x00invalid: [yaml"))
	if err == nil {
		t.Error("Apply() expected error for invalid YAML, got nil")
	}
}

// =============================================================================
// WorkspaceHandler Tests - Get
// =============================================================================

func TestWorkspaceHandler_Get_Found(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, appID := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	ws := &models.Workspace{Name: "get-ws", AppID: appID, ImageName: "img:1", Status: "stopped"}
	if err := store.CreateWorkspace(ws); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	res, err := h.Get(ctx, "get-ws")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if res.GetName() != "get-ws" {
		t.Errorf("Get() Name = %q, want %q", res.GetName(), "get-ws")
	}
	if res.GetKind() != KindWorkspace {
		t.Errorf("Get() Kind = %q, want %q", res.GetKind(), KindWorkspace)
	}
}

func TestWorkspaceHandler_Get_NoActiveApp(t *testing.T) {
	h := NewWorkspaceHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	_, err := h.Get(ctx, "any-ws")
	if err == nil {
		t.Error("Get() expected error when no active app, got nil")
	}
	if !strings.Contains(err.Error(), "active app") {
		t.Errorf("Get() error = %q, want it to mention 'active app'", err.Error())
	}
}

func TestWorkspaceHandler_Get_NotFound(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, _ := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	_, err := h.Get(ctx, "does-not-exist")
	if err == nil {
		t.Error("Get() expected error for non-existent workspace, got nil")
	}
}

// =============================================================================
// WorkspaceHandler Tests - List
// =============================================================================

func TestWorkspaceHandler_List_WithActiveApp(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, appID := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	_ = store.CreateWorkspace(&models.Workspace{Name: "ws1", AppID: appID, ImageName: "img:1", Status: "stopped"})
	_ = store.CreateWorkspace(&models.Workspace{Name: "ws2", AppID: appID, ImageName: "img:2", Status: "stopped"})

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("List() returned %d items, want 2", len(resources))
	}
	for _, r := range resources {
		if r.GetKind() != KindWorkspace {
			t.Errorf("List() item Kind = %q, want %q", r.GetKind(), KindWorkspace)
		}
	}
}

func TestWorkspaceHandler_List_NoActiveApp(t *testing.T) {
	h := NewWorkspaceHandler()
	// Store with multiple apps but no active app set
	store := db.NewMockDataStore()

	eco := &models.Ecosystem{Name: "list-eco"}
	_ = store.CreateEcosystem(eco)
	domain := &models.Domain{Name: "list-domain", EcosystemID: eco.ID}
	_ = store.CreateDomain(domain)
	app1 := &models.App{Name: "list-app1", DomainID: domain.ID, Path: "/p1"}
	app2 := &models.App{Name: "list-app2", DomainID: domain.ID, Path: "/p2"}
	_ = store.CreateApp(app1)
	_ = store.CreateApp(app2)
	_ = store.CreateWorkspace(&models.Workspace{Name: "ws-a1", AppID: app1.ID, ImageName: "img:1", Status: "stopped"})
	_ = store.CreateWorkspace(&models.Workspace{Name: "ws-a2", AppID: app2.ID, ImageName: "img:2", Status: "stopped"})

	ctx := resource.Context{DataStore: store}
	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("List() returned %d items, want 2", len(resources))
	}
}

// =============================================================================
// WorkspaceHandler Tests - Delete
// =============================================================================

func TestWorkspaceHandler_Delete_Found(t *testing.T) {
	h := NewWorkspaceHandler()
	store, _, _, appID := setupWorkspaceTest(t)
	ctx := resource.Context{DataStore: store}

	ws := &models.Workspace{Name: "del-ws", AppID: appID, ImageName: "img:1", Status: "stopped"}
	if err := store.CreateWorkspace(ws); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	if err := h.Delete(ctx, "del-ws"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify removed
	_, err := store.GetWorkspaceByName(appID, "del-ws")
	if err == nil {
		t.Error("Delete() did not remove workspace from store")
	}
}

func TestWorkspaceHandler_Delete_NoActiveApp(t *testing.T) {
	h := NewWorkspaceHandler()
	store := db.NewMockDataStore()
	ctx := resource.Context{DataStore: store}

	err := h.Delete(ctx, "any-ws")
	if err == nil {
		t.Error("Delete() expected error when no active app, got nil")
	}
	if !strings.Contains(err.Error(), "active app") {
		t.Errorf("Delete() error = %q, want it to mention 'active app'", err.Error())
	}
}

// =============================================================================
// WorkspaceHandler Tests - ToYAML
// =============================================================================

func TestWorkspaceHandler_ToYAML(t *testing.T) {
	h := NewWorkspaceHandler()

	ws := &models.Workspace{
		ID:        1,
		AppID:     1,
		Name:      "yaml-ws",
		ImageName: "ubuntu:22.04",
		Status:    "stopped",
	}
	res := NewWorkspaceResource(ws, "yaml-app")

	yamlBytes, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}
	yamlStr := string(yamlBytes)

	if !strings.Contains(yamlStr, "kind: Workspace") {
		t.Errorf("ToYAML() missing 'kind: Workspace', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "name: yaml-ws") {
		t.Errorf("ToYAML() missing 'name: yaml-ws', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "app: yaml-app") {
		t.Errorf("ToYAML() missing 'app: yaml-app', got:\n%s", yamlStr)
	}
	if !strings.Contains(yamlStr, "ubuntu:22.04") {
		t.Errorf("ToYAML() missing 'ubuntu:22.04', got:\n%s", yamlStr)
	}
}

func TestWorkspaceHandler_ToYAML_WrongType(t *testing.T) {
	h := NewWorkspaceHandler()
	wrongRes := NewEcosystemResource(&models.Ecosystem{Name: "wrong"})
	_, err := h.ToYAML(wrongRes)
	if err == nil {
		t.Error("ToYAML() expected error for wrong resource type, got nil")
	}
}

// =============================================================================
// WorkspaceResource Validate Tests
// =============================================================================

func TestWorkspaceResource_Validate(t *testing.T) {
	tests := []struct {
		name      string
		workspace *models.Workspace
		wantErr   bool
	}{
		{
			name: "valid workspace",
			workspace: &models.Workspace{
				Name:      "good-ws",
				AppID:     1,
				ImageName: "ubuntu:22.04",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			workspace: &models.Workspace{
				Name:      "",
				AppID:     1,
				ImageName: "ubuntu:22.04",
			},
			wantErr: true,
		},
		{
			name: "missing app_id",
			workspace: &models.Workspace{
				Name:      "no-app",
				AppID:     0,
				ImageName: "ubuntu:22.04",
			},
			wantErr: true,
		},
		{
			name: "missing image_name",
			workspace: &models.Workspace{
				Name:      "no-image",
				AppID:     1,
				ImageName: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := NewWorkspaceResource(tt.workspace, "")
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
