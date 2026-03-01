package db

import (
	"database/sql"
	"devopsmaestro/models"
	"testing"
)

// =============================================================================
// Delete Operations - ErrNotFound Tests
// =============================================================================
// These tests verify that all Delete methods return ErrNotFound when attempting
// to delete a resource that doesn't exist.

func TestDeleteEcosystem_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteEcosystem("nonexistent")
	if err == nil {
		t.Fatal("DeleteEcosystem() should return error for nonexistent ecosystem")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteEcosystem() error should be ErrNotFound, got: %v", err)
	}

	// Verify error message
	expectedMsg := "ecosystem 'nonexistent' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteDomain_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteDomain(999)
	if err == nil {
		t.Fatal("DeleteDomain() should return error for nonexistent domain")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteDomain() error should be ErrNotFound, got: %v", err)
	}

	// Verify error message contains expected parts
	expectedMsg := "domain '999' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteApp_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteApp(999)
	if err == nil {
		t.Fatal("DeleteApp() should return error for nonexistent app")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteApp() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "app '999' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteWorkspace_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteWorkspace(999)
	if err == nil {
		t.Fatal("DeleteWorkspace() should return error for nonexistent workspace")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteWorkspace() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "workspace '999' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeletePlugin_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeletePlugin("nonexistent-plugin")
	if err == nil {
		t.Fatal("DeletePlugin() should return error for nonexistent plugin")
	}
	if !IsNotFound(err) {
		t.Errorf("DeletePlugin() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "plugin 'nonexistent-plugin' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteTheme_NotFound(t *testing.T) {
	t.Skip("Skipping: nvim_themes table not in test schema")
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteTheme("nonexistent-theme")
	if err == nil {
		t.Fatal("DeleteTheme() should return error for nonexistent theme")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteTheme() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "theme 'nonexistent-theme' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteTerminalPrompt_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteTerminalPrompt("nonexistent-prompt")
	if err == nil {
		t.Fatal("DeleteTerminalPrompt() should return error for nonexistent prompt")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteTerminalPrompt() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "terminal prompt 'nonexistent-prompt' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteTerminalProfile_NotFound(t *testing.T) {
	t.Skip("Skipping: terminal_profiles table not in test schema")
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteTerminalProfile("nonexistent-profile")
	if err == nil {
		t.Fatal("DeleteTerminalProfile() should return error for nonexistent profile")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteTerminalProfile() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "terminal profile 'nonexistent-profile' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteTerminalPlugin_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteTerminalPlugin("nonexistent-plugin")
	if err == nil {
		t.Fatal("DeleteTerminalPlugin() should return error for nonexistent plugin")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteTerminalPlugin() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "terminal plugin 'nonexistent-plugin' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteTerminalEmulator_NotFound(t *testing.T) {
	t.Skip("Skipping: terminal_emulators table not in test schema")
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteTerminalEmulator("nonexistent-emulator")
	if err == nil {
		t.Fatal("DeleteTerminalEmulator() should return error for nonexistent emulator")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteTerminalEmulator() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "terminal emulator 'nonexistent-emulator' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteCredential_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Use ecosystem scope for testing (no "global" scope exists)
	err := ds.DeleteCredential(models.CredentialScopeEcosystem, 999, "nonexistent-cred")
	if err == nil {
		t.Fatal("DeleteCredential() should return error for nonexistent credential")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteCredential() error should be ErrNotFound, got: %v", err)
	}
}

func TestDeleteDefault_NotFound(t *testing.T) {
	t.Skip("Skipping: defaults table not in test schema")
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteDefault("nonexistent-default")
	if err == nil {
		t.Fatal("DeleteDefault() should return error for nonexistent default")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteDefault() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "default 'nonexistent-default' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeletePackage_NotFound(t *testing.T) {
	t.Skip("Skipping: nvim_packages table not in test schema")
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeletePackage("nonexistent-package")
	if err == nil {
		t.Fatal("DeletePackage() should return error for nonexistent package")
	}
	if !IsNotFound(err) {
		t.Errorf("DeletePackage() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "package 'nonexistent-package' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteTerminalPackage_NotFound(t *testing.T) {
	t.Skip("Skipping: terminal_packages table not in test schema")
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteTerminalPackage("nonexistent-package")
	if err == nil {
		t.Fatal("DeleteTerminalPackage() should return error for nonexistent package")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteTerminalPackage() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "terminal package 'nonexistent-package' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

func TestDeleteGitRepo_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteGitRepo("nonexistent-repo")
	if err == nil {
		t.Fatal("DeleteGitRepo() should return error for nonexistent git repo")
	}
	if !IsNotFound(err) {
		t.Errorf("DeleteGitRepo() error should be ErrNotFound, got: %v", err)
	}

	expectedMsg := "git repo 'nonexistent-repo' not found"
	if err.Error() != expectedMsg {
		t.Errorf("Error message = %q, want %q", err.Error(), expectedMsg)
	}
}

// =============================================================================
// Table-Driven Delete Tests
// =============================================================================
// These tests verify the entire delete lifecycle: create, delete, verify gone.

func TestDeleteOperations_Lifecycle(t *testing.T) {
	tests := []struct {
		name         string
		createFunc   func(*SQLDataStore) (interface{}, error)
		deleteFunc   func(*SQLDataStore) error
		verifyFunc   func(*SQLDataStore) error
		resourceName string
	}{
		{
			name: "ecosystem lifecycle",
			createFunc: func(ds *SQLDataStore) (interface{}, error) {
				eco := &models.Ecosystem{
					Name:        "test-eco",
					Description: sql.NullString{String: "Test", Valid: true},
				}
				return eco, ds.CreateEcosystem(eco)
			},
			deleteFunc: func(ds *SQLDataStore) error {
				return ds.DeleteEcosystem("test-eco")
			},
			verifyFunc: func(ds *SQLDataStore) error {
				_, err := ds.GetEcosystemByName("test-eco")
				return err
			},
			resourceName: "ecosystem",
		},
		{
			name: "plugin lifecycle",
			createFunc: func(ds *SQLDataStore) (interface{}, error) {
				plugin := &models.NvimPluginDB{
					Name: "test-plugin",
					Repo: "https://github.com/test/plugin",
				}
				return plugin, ds.CreatePlugin(plugin)
			},
			deleteFunc: func(ds *SQLDataStore) error {
				return ds.DeletePlugin("test-plugin")
			},
			verifyFunc: func(ds *SQLDataStore) error {
				_, err := ds.GetPluginByName("test-plugin")
				return err
			},
			resourceName: "plugin",
		},
		// Theme lifecycle test skipped - nvim_themes table not in test schema
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			// Create resource
			_, err := tt.createFunc(ds)
			if err != nil {
				t.Fatalf("Failed to create %s: %v", tt.resourceName, err)
			}

			// Delete resource
			err = tt.deleteFunc(ds)
			if err != nil {
				t.Fatalf("Failed to delete %s: %v", tt.resourceName, err)
			}

			// Verify it's gone (Get should fail)
			err = tt.verifyFunc(ds)
			if err == nil {
				t.Errorf("Get after delete should fail for %s", tt.resourceName)
			}

			// Try to delete again - should get ErrNotFound
			err = tt.deleteFunc(ds)
			if err == nil {
				t.Fatalf("Second delete should return error for %s", tt.resourceName)
			}
			if !IsNotFound(err) {
				t.Errorf("Second delete should return ErrNotFound for %s, got: %v", tt.resourceName, err)
			}
		})
	}
}

// =============================================================================
// Hierarchical Delete Tests
// =============================================================================
// Test delete operations with foreign key relationships

func TestDeleteDomain_WithHierarchy(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem
	eco := &models.Ecosystem{
		Name:        "test-eco",
		Description: sql.NullString{String: "Test", Valid: true},
	}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("Failed to create ecosystem: %v", err)
	}

	// Create domain
	domain := &models.Domain{Name: "test-domain", EcosystemID: eco.ID}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Failed to create domain: %v", err)
	}

	// Delete domain
	err := ds.DeleteDomain(domain.ID)
	if err != nil {
		t.Fatalf("Failed to delete domain: %v", err)
	}

	// Try to delete again - should get ErrNotFound
	err = ds.DeleteDomain(domain.ID)
	if err == nil {
		t.Fatal("Second delete should return error")
	}
	if !IsNotFound(err) {
		t.Errorf("Second delete should return ErrNotFound, got: %v", err)
	}
}

func TestDeleteApp_WithHierarchy(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem → domain → app
	eco := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("Failed to create ecosystem: %v", err)
	}

	domain := &models.Domain{Name: "test-domain", EcosystemID: eco.ID}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Failed to create domain: %v", err)
	}

	app := &models.App{Name: "test-app", DomainID: domain.ID}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	// Delete app
	err := ds.DeleteApp(app.ID)
	if err != nil {
		t.Fatalf("Failed to delete app: %v", err)
	}

	// Try to delete again - should get ErrNotFound
	err = ds.DeleteApp(app.ID)
	if err == nil {
		t.Fatal("Second delete should return error")
	}
	if !IsNotFound(err) {
		t.Errorf("Second delete should return ErrNotFound, got: %v", err)
	}
}

func TestDeleteWorkspace_WithHierarchy(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create full hierarchy
	eco := &models.Ecosystem{Name: "test-eco"}
	if err := ds.CreateEcosystem(eco); err != nil {
		t.Fatalf("Failed to create ecosystem: %v", err)
	}

	domain := &models.Domain{Name: "test-domain", EcosystemID: eco.ID}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Failed to create domain: %v", err)
	}

	app := &models.App{Name: "test-app", DomainID: domain.ID}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Failed to create app: %v", err)
	}

	workspace := &models.Workspace{Name: "test-ws", AppID: app.ID}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Failed to create workspace: %v", err)
	}

	// Delete workspace
	err := ds.DeleteWorkspace(workspace.ID)
	if err != nil {
		t.Fatalf("Failed to delete workspace: %v", err)
	}

	// Try to delete again - should get ErrNotFound
	err = ds.DeleteWorkspace(workspace.ID)
	if err == nil {
		t.Fatal("Second delete should return error")
	}
	if !IsNotFound(err) {
		t.Errorf("Second delete should return ErrNotFound, got: %v", err)
	}
}
