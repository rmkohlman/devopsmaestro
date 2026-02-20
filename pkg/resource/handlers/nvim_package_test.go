package handlers

import (
	"errors"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"
	nvimpackage "devopsmaestro/pkg/nvimops/package"
	"devopsmaestro/pkg/resource"
)

func TestNvimPackageHandler_Kind(t *testing.T) {
	h := NewNvimPackageHandler()
	if h.Kind() != KindNvimPackage {
		t.Errorf("Kind() = %v, want %v", h.Kind(), KindNvimPackage)
	}
}

func TestNvimPackageHandler_Apply(t *testing.T) {
	h := NewNvimPackageHandler()
	mockStore := db.NewMockDataStore()

	// Set up mock to handle GetPackage (not found) and CreatePackage
	mockStore.GetPackageErr = errors.New("package not found")
	mockStore.CreatePackageErr = nil

	ctx := resource.Context{
		DataStore: mockStore,
	}

	yaml := `apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: core
  description: "Essential plugins"
  category: core
  labels:
    source: library
spec:
  plugins:
    - telescope
    - treesitter
    - lspconfig`

	res, err := h.Apply(ctx, []byte(yaml))
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if res.GetKind() != KindNvimPackage {
		t.Errorf("Apply() resource.Kind = %v, want %v", res.GetKind(), KindNvimPackage)
	}
	if res.GetName() != "core" {
		t.Errorf("Apply() resource.Name = %v, want core", res.GetName())
	}

	// Verify CreatePackage was called by checking method calls
	calls := mockStore.GetCalls()
	createCalled := false
	for _, call := range calls {
		if call.Method == "CreatePackage" {
			createCalled = true
			break
		}
	}
	if !createCalled {
		t.Error("CreatePackage was not called")
	}
}

func TestNvimPackageHandler_Apply_Update(t *testing.T) {
	h := NewNvimPackageHandler()
	mockStore := db.NewMockDataStore()

	// Set up mock to return existing package, then allow update
	existing := &models.NvimPackageDB{
		ID:   1,
		Name: "core",
	}
	existing.SetPlugins([]string{"telescope"})

	// Pre-populate the mock store
	mockStore.Packages["core"] = existing
	mockStore.GetPackageErr = nil
	mockStore.UpdatePackageErr = nil

	ctx := resource.Context{
		DataStore: mockStore,
	}

	yaml := `apiVersion: devopsmaestro.io/v1
kind: NvimPackage
metadata:
  name: core
  description: "Essential plugins updated"
spec:
  plugins:
    - telescope
    - treesitter`

	res, err := h.Apply(ctx, []byte(yaml))
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}

	if res.GetName() != "core" {
		t.Errorf("Apply() resource.Name = %v, want core", res.GetName())
	}

	// Verify UpdatePackage was called
	calls := mockStore.GetCalls()
	updateCalled := false
	for _, call := range calls {
		if call.Method == "UpdatePackage" {
			updateCalled = true
			break
		}
	}
	if !updateCalled {
		t.Error("UpdatePackage was not called")
	}
}

func TestNvimPackageHandler_Get(t *testing.T) {
	h := NewNvimPackageHandler()
	mockStore := db.NewMockDataStore()

	// Set up mock result
	dbPkg := &models.NvimPackageDB{
		ID:   1,
		Name: "test-package",
	}
	dbPkg.Description.String = "Test description"
	dbPkg.Description.Valid = true
	dbPkg.SetPlugins([]string{"telescope", "treesitter"})

	// Pre-populate the mock store
	mockStore.Packages["test-package"] = dbPkg
	mockStore.GetPackageErr = nil

	ctx := resource.Context{
		DataStore: mockStore,
	}

	res, err := h.Get(ctx, "test-package")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if res.GetName() != "test-package" {
		t.Errorf("Get() resource.Name = %v, want test-package", res.GetName())
	}

	// Cast to access underlying package
	pr, ok := res.(*NvimPackageResource)
	if !ok {
		t.Fatalf("Get() result is not *NvimPackageResource")
	}
	if pr.Package().Description != "Test description" {
		t.Errorf("Get() package.Description = %v, want Test description", pr.Package().Description)
	}
	if len(pr.Package().Plugins) != 2 {
		t.Errorf("Get() package.Plugins length = %d, want 2", len(pr.Package().Plugins))
	}
}

func TestNvimPackageHandler_List(t *testing.T) {
	h := NewNvimPackageHandler()
	mockStore := db.NewMockDataStore()

	// Set up mock results
	dbPkg1 := &models.NvimPackageDB{ID: 1, Name: "package1"}
	dbPkg1.SetPlugins([]string{"telescope"})
	dbPkg2 := &models.NvimPackageDB{ID: 2, Name: "package2"}
	dbPkg2.SetPlugins([]string{"treesitter"})

	// Pre-populate the mock store
	mockStore.Packages["package1"] = dbPkg1
	mockStore.Packages["package2"] = dbPkg2
	mockStore.ListPackagesErr = nil

	ctx := resource.Context{
		DataStore: mockStore,
	}

	resources, err := h.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// List now includes both database packages AND library packages
	// Library has 5 packages (core, full, go-dev, python-dev, rmkohlman)
	// Database has 2 packages (package1, package2)
	// Total should be at least 7 (could be more if library grows)
	if len(resources) < 7 {
		t.Errorf("List() returned %d resources, want at least 7 (2 db + 5 library)", len(resources))
	}

	// Verify our database packages are included
	names := make(map[string]bool)
	for _, res := range resources {
		names[res.GetName()] = true
	}
	if !names["package1"] || !names["package2"] {
		t.Error("List() did not return expected database package names")
	}
	// Also verify library packages are included
	if !names["core"] {
		t.Error("List() did not return library package 'core'")
	}
}

func TestNvimPackageHandler_Delete(t *testing.T) {
	h := NewNvimPackageHandler()
	mockStore := db.NewMockDataStore()

	// Pre-populate with a package to delete
	dbPkg := &models.NvimPackageDB{
		ID:   1,
		Name: "to-delete",
	}
	dbPkg.SetPlugins([]string{"telescope"})
	mockStore.Packages["to-delete"] = dbPkg

	mockStore.DeletePackageErr = nil

	ctx := resource.Context{
		DataStore: mockStore,
	}

	err := h.Delete(ctx, "to-delete")
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify DeletePackage was called
	calls := mockStore.GetCalls()
	deleteCalled := false
	var deleteCall *db.MockDataStoreCall
	for i := range calls {
		if calls[i].Method == "DeletePackage" {
			deleteCalled = true
			deleteCall = &calls[i]
			break
		}
	}
	if !deleteCalled {
		t.Error("DeletePackage was not called")
	}

	// Verify correct name was passed
	if deleteCall != nil && len(deleteCall.Args) > 0 {
		if deleteCall.Args[0] != "to-delete" {
			t.Errorf("DeletePackage called with name %v, want to-delete", deleteCall.Args[0])
		}
	}
}

func TestNvimPackageHandler_ToYAML(t *testing.T) {
	h := NewNvimPackageHandler()

	res := &NvimPackageResource{
		pkg: &nvimpackage.Package{
			Name:        "test-package",
			Description: "Test description",
			Category:    "test",
			Plugins:     []string{"telescope", "treesitter"},
			Enabled:     true,
		},
	}

	yaml, err := h.ToYAML(res)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	// Verify YAML contains expected content
	yamlStr := string(yaml)
	if !contains(yamlStr, "kind: NvimPackage") {
		t.Error("ToYAML() missing 'kind: NvimPackage'")
	}
	if !contains(yamlStr, "name: test-package") {
		t.Error("ToYAML() missing 'name: test-package'")
	}
	if !contains(yamlStr, "telescope") {
		t.Error("ToYAML() missing 'telescope' plugin")
	}
}

func TestNvimPackageResource_Validate(t *testing.T) {
	tests := []struct {
		name    string
		pkg     *nvimpackage.Package
		wantErr bool
	}{
		{
			name:    "valid package",
			pkg:     &nvimpackage.Package{Name: "test", Plugins: []string{"telescope"}},
			wantErr: false,
		},
		{
			name:    "missing name",
			pkg:     &nvimpackage.Package{Name: "", Plugins: []string{"telescope"}},
			wantErr: true,
		},
		{
			name:    "no plugins",
			pkg:     &nvimpackage.Package{Name: "test", Plugins: []string{}},
			wantErr: true,
		},
		{
			name:    "empty plugin name",
			pkg:     &nvimpackage.Package{Name: "test", Plugins: []string{"telescope", ""}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &NvimPackageResource{pkg: tt.pkg}
			err := res.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNvimPackageHandler_DBModelConversion(t *testing.T) {
	h := NewNvimPackageHandler()

	// Test Package -> DB conversion
	pkg := &nvimpackage.Package{
		Name:        "test",
		Description: "Test package",
		Category:    "test-category",
		Tags:        []string{"tag1", "tag2"},
		Plugins:     []string{"telescope", "treesitter"},
		Extends:     "base",
		Enabled:     false,
	}

	dbPkg, err := h.toDBModel(pkg)
	if err != nil {
		t.Fatalf("toDBModel() error = %v", err)
	}

	if dbPkg.Name != "test" {
		t.Errorf("toDBModel() Name = %v, want test", dbPkg.Name)
	}
	if !dbPkg.Description.Valid || dbPkg.Description.String != "Test package" {
		t.Errorf("toDBModel() Description = %v, want Test package", dbPkg.Description)
	}
	if !dbPkg.Extends.Valid || dbPkg.Extends.String != "base" {
		t.Errorf("toDBModel() Extends = %v, want base", dbPkg.Extends)
	}

	// Check plugins
	plugins := dbPkg.GetPlugins()
	if len(plugins) != 2 || plugins[0] != "telescope" || plugins[1] != "treesitter" {
		t.Errorf("toDBModel() Plugins = %v, want [telescope treesitter]", plugins)
	}

	// Check labels (tags and enabled status)
	labels := dbPkg.GetLabels()
	if labels["tags"] != "tag1,tag2" {
		t.Errorf("toDBModel() tags label = %v, want tag1,tag2", labels["tags"])
	}
	if labels["enabled"] != "false" {
		t.Errorf("toDBModel() enabled label = %v, want false", labels["enabled"])
	}

	// Test DB -> Package conversion
	converted, err := h.fromDBModel(dbPkg)
	if err != nil {
		t.Fatalf("fromDBModel() error = %v", err)
	}

	if converted.Name != pkg.Name {
		t.Errorf("fromDBModel() Name = %v, want %v", converted.Name, pkg.Name)
	}
	if converted.Description != pkg.Description {
		t.Errorf("fromDBModel() Description = %v, want %v", converted.Description, pkg.Description)
	}
	if converted.Enabled != pkg.Enabled {
		t.Errorf("fromDBModel() Enabled = %v, want %v", converted.Enabled, pkg.Enabled)
	}
	if len(converted.Tags) != len(pkg.Tags) {
		t.Errorf("fromDBModel() Tags length = %d, want %d", len(converted.Tags), len(pkg.Tags))
	}
}
