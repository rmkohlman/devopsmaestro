package db

import (
	"database/sql"
	"devopsmaestro/models"
	"strings"
	"testing"
)

// =============================================================================
// Test Setup Helpers
// =============================================================================

// createTestDataStore creates an in-memory SQLite database with schema for testing
func createTestDataStore(t *testing.T) *SQLDataStore {
	t.Helper()

	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("Failed to create test driver: %v", err)
	}

	if err := driver.Connect(); err != nil {
		t.Fatalf("Failed to connect test driver: %v", err)
	}

	// Create schema
	if err := createTestSchema(driver); err != nil {
		driver.Close()
		t.Fatalf("Failed to create test schema: %v", err)
	}

	return NewSQLDataStore(driver, nil)
}

// createTestSchema creates the required tables for testing
func createTestSchema(driver Driver) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS ecosystems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ecosystem_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id),
			UNIQUE(ecosystem_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS apps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			description TEXT,
			language TEXT,
			build_config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (domain_id) REFERENCES domains(id),
			UNIQUE(domain_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS projects (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			path TEXT,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			app_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			image_name TEXT,
			container_id TEXT,
			status TEXT DEFAULT 'stopped',
			nvim_structure TEXT,
			nvim_plugins TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (app_id) REFERENCES apps(id),
			UNIQUE(app_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS context (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			active_ecosystem_id INTEGER,
			active_domain_id INTEGER,
			active_app_id INTEGER,
			active_workspace_id INTEGER,
			active_project_id INTEGER,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (active_ecosystem_id) REFERENCES ecosystems(id),
			FOREIGN KEY (active_domain_id) REFERENCES domains(id),
			FOREIGN KEY (active_app_id) REFERENCES apps(id),
			FOREIGN KEY (active_workspace_id) REFERENCES workspaces(id),
			FOREIGN KEY (active_project_id) REFERENCES projects(id)
		)`,
		`CREATE TABLE IF NOT EXISTS nvim_plugins (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			repo TEXT NOT NULL,
			branch TEXT,
			version TEXT,
			priority INTEGER,
			lazy INTEGER DEFAULT 0,
			event TEXT,
			ft TEXT,
			keys TEXT,
			cmd TEXT,
			dependencies TEXT,
			build TEXT,
			config TEXT,
			init TEXT,
			opts TEXT,
			keymaps TEXT,
			category TEXT,
			tags TEXT,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS workspace_plugins (
			workspace_id INTEGER NOT NULL,
			plugin_id INTEGER NOT NULL,
			enabled INTEGER DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (workspace_id, plugin_id),
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE,
			FOREIGN KEY (plugin_id) REFERENCES nvim_plugins(id) ON DELETE CASCADE
		)`,
		// Initialize context with a single row
		`INSERT OR IGNORE INTO context (id) VALUES (1)`,
	}

	for _, query := range queries {
		if _, err := driver.Execute(query); err != nil {
			return err
		}
	}

	return nil
}

// =============================================================================
// Ecosystem CRUD Tests
// =============================================================================

func TestSQLDataStore_CreateEcosystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{
		Name: "test-ecosystem",
		Description: sql.NullString{
			String: "Test description",
			Valid:  true,
		},
	}

	err := ds.CreateEcosystem(ecosystem)
	if err != nil {
		t.Fatalf("CreateEcosystem() error = %v", err)
	}

	if ecosystem.ID == 0 {
		t.Errorf("CreateEcosystem() did not set ecosystem.ID")
	}
}

func TestSQLDataStore_GetEcosystemByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create an ecosystem first
	ecosystem := &models.Ecosystem{
		Name: "findme-ecosystem",
		Description: sql.NullString{
			String: "Find me",
			Valid:  true,
		},
	}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup CreateEcosystem() error = %v", err)
	}

	// Retrieve by name
	retrieved, err := ds.GetEcosystemByName("findme-ecosystem")
	if err != nil {
		t.Fatalf("GetEcosystemByName() error = %v", err)
	}

	if retrieved.Name != "findme-ecosystem" {
		t.Errorf("GetEcosystemByName() Name = %q, want %q", retrieved.Name, "findme-ecosystem")
	}
	if retrieved.Description.String != "Find me" {
		t.Errorf("GetEcosystemByName() Description = %q, want %q", retrieved.Description.String, "Find me")
	}
}

func TestSQLDataStore_GetEcosystemByName_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	_, err := ds.GetEcosystemByName("nonexistent")
	if err == nil {
		t.Errorf("GetEcosystemByName() expected error for nonexistent ecosystem")
	}
}

func TestSQLDataStore_GetEcosystemByID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{
		Name: "getbyid-ecosystem",
	}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetEcosystemByID(ecosystem.ID)
	if err != nil {
		t.Fatalf("GetEcosystemByID() error = %v", err)
	}

	if retrieved.ID != ecosystem.ID {
		t.Errorf("GetEcosystemByID() ID = %d, want %d", retrieved.ID, ecosystem.ID)
	}
}

func TestSQLDataStore_UpdateEcosystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{
		Name: "update-ecosystem",
	}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Update the ecosystem
	ecosystem.Description = sql.NullString{String: "Updated description", Valid: true}

	if err := ds.UpdateEcosystem(ecosystem); err != nil {
		t.Fatalf("UpdateEcosystem() error = %v", err)
	}

	// Verify update
	retrieved, err := ds.GetEcosystemByID(ecosystem.ID)
	if err != nil {
		t.Fatalf("Verification error: %v", err)
	}

	if retrieved.Description.String != "Updated description" {
		t.Errorf("UpdateEcosystem() Description = %q, want %q", retrieved.Description.String, "Updated description")
	}
}

func TestSQLDataStore_DeleteEcosystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{
		Name: "delete-ecosystem",
	}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	if err := ds.DeleteEcosystem("delete-ecosystem"); err != nil {
		t.Fatalf("DeleteEcosystem() error = %v", err)
	}

	// Verify deletion
	_, err := ds.GetEcosystemByName("delete-ecosystem")
	if err == nil {
		t.Errorf("DeleteEcosystem() ecosystem should not exist after deletion")
	}
}

func TestSQLDataStore_ListEcosystems(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create multiple ecosystems
	for i := 1; i <= 3; i++ {
		ecosystem := &models.Ecosystem{
			Name: "list-ecosystem-" + string(rune('0'+i)),
		}
		if err := ds.CreateEcosystem(ecosystem); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	ecosystems, err := ds.ListEcosystems()
	if err != nil {
		t.Fatalf("ListEcosystems() error = %v", err)
	}

	if len(ecosystems) != 3 {
		t.Errorf("ListEcosystems() returned %d ecosystems, want 3", len(ecosystems))
	}
}

// =============================================================================
// Domain CRUD Tests
// =============================================================================

func TestSQLDataStore_CreateDomain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create an ecosystem first
	ecosystem := &models.Ecosystem{Name: "domain-test-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup CreateEcosystem() error = %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "test-domain",
		Description: sql.NullString{
			String: "Test domain description",
			Valid:  true,
		},
	}

	err := ds.CreateDomain(domain)
	if err != nil {
		t.Fatalf("CreateDomain() error = %v", err)
	}

	if domain.ID == 0 {
		t.Errorf("CreateDomain() did not set domain.ID")
	}
}

func TestSQLDataStore_GetDomainByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create an ecosystem first
	ecosystem := &models.Ecosystem{Name: "domain-findme-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup CreateEcosystem() error = %v", err)
	}

	// Create a domain
	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "findme-domain",
		Description: sql.NullString{
			String: "Find me",
			Valid:  true,
		},
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup CreateDomain() error = %v", err)
	}

	// Retrieve by name
	retrieved, err := ds.GetDomainByName(ecosystem.ID, "findme-domain")
	if err != nil {
		t.Fatalf("GetDomainByName() error = %v", err)
	}

	if retrieved.Name != "findme-domain" {
		t.Errorf("GetDomainByName() Name = %q, want %q", retrieved.Name, "findme-domain")
	}
	if retrieved.EcosystemID != ecosystem.ID {
		t.Errorf("GetDomainByName() EcosystemID = %d, want %d", retrieved.EcosystemID, ecosystem.ID)
	}
}

func TestSQLDataStore_GetDomainByName_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create an ecosystem first
	ecosystem := &models.Ecosystem{Name: "domain-notfound-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	_, err := ds.GetDomainByName(ecosystem.ID, "nonexistent")
	if err == nil {
		t.Errorf("GetDomainByName() expected error for nonexistent domain")
	}
}

func TestSQLDataStore_GetDomainByID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "domain-getbyid-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "getbyid-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetDomainByID(domain.ID)
	if err != nil {
		t.Fatalf("GetDomainByID() error = %v", err)
	}

	if retrieved.ID != domain.ID {
		t.Errorf("GetDomainByID() ID = %d, want %d", retrieved.ID, domain.ID)
	}
}

func TestSQLDataStore_UpdateDomain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "domain-update-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "update-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Update the domain
	domain.Description = sql.NullString{String: "Updated description", Valid: true}

	if err := ds.UpdateDomain(domain); err != nil {
		t.Fatalf("UpdateDomain() error = %v", err)
	}

	// Verify update
	retrieved, err := ds.GetDomainByID(domain.ID)
	if err != nil {
		t.Fatalf("Verification error: %v", err)
	}

	if retrieved.Description.String != "Updated description" {
		t.Errorf("UpdateDomain() Description = %q, want %q", retrieved.Description.String, "Updated description")
	}
}

func TestSQLDataStore_DeleteDomain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "domain-delete-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "delete-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	if err := ds.DeleteDomain(domain.ID); err != nil {
		t.Fatalf("DeleteDomain() error = %v", err)
	}

	// Verify deletion
	_, err := ds.GetDomainByID(domain.ID)
	if err == nil {
		t.Errorf("DeleteDomain() domain should not exist after deletion")
	}
}

func TestSQLDataStore_ListDomainsByEcosystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "domain-list-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create multiple domains
	for i := 1; i <= 3; i++ {
		domain := &models.Domain{
			EcosystemID: ecosystem.ID,
			Name:        "list-domain-" + string(rune('0'+i)),
		}
		if err := ds.CreateDomain(domain); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	domains, err := ds.ListDomainsByEcosystem(ecosystem.ID)
	if err != nil {
		t.Fatalf("ListDomainsByEcosystem() error = %v", err)
	}

	if len(domains) != 3 {
		t.Errorf("ListDomainsByEcosystem() returned %d domains, want 3", len(domains))
	}
}

func TestSQLDataStore_ListAllDomains(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create two ecosystems with domains
	for e := 1; e <= 2; e++ {
		ecosystem := &models.Ecosystem{
			Name: "listall-eco-" + string(rune('0'+e)),
		}
		if err := ds.CreateEcosystem(ecosystem); err != nil {
			t.Fatalf("Setup error: %v", err)
		}

		for d := 1; d <= 2; d++ {
			domain := &models.Domain{
				EcosystemID: ecosystem.ID,
				Name:        "domain-" + string(rune('0'+d)),
			}
			if err := ds.CreateDomain(domain); err != nil {
				t.Fatalf("Setup error: %v", err)
			}
		}
	}

	domains, err := ds.ListAllDomains()
	if err != nil {
		t.Fatalf("ListAllDomains() error = %v", err)
	}

	if len(domains) != 4 {
		t.Errorf("ListAllDomains() returned %d domains, want 4", len(domains))
	}
}

// =============================================================================
// App CRUD Tests
// =============================================================================

func TestSQLDataStore_CreateApp(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem and domain first
	ecosystem := &models.Ecosystem{Name: "app-test-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-test-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "test-app",
		Path:     "/path/to/app",
		Description: sql.NullString{
			String: "Test app description",
			Valid:  true,
		},
	}

	err := ds.CreateApp(app)
	if err != nil {
		t.Fatalf("CreateApp() error = %v", err)
	}

	if app.ID == 0 {
		t.Errorf("CreateApp() did not set app.ID")
	}
}

func TestSQLDataStore_GetAppByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem and domain
	ecosystem := &models.Ecosystem{Name: "app-findme-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-findme-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create an app
	app := &models.App{
		DomainID: domain.ID,
		Name:     "findme-app",
		Path:     "/path/to/findme",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Retrieve by name
	retrieved, err := ds.GetAppByName(domain.ID, "findme-app")
	if err != nil {
		t.Fatalf("GetAppByName() error = %v", err)
	}

	if retrieved.Name != "findme-app" {
		t.Errorf("GetAppByName() Name = %q, want %q", retrieved.Name, "findme-app")
	}
	if retrieved.Path != "/path/to/findme" {
		t.Errorf("GetAppByName() Path = %q, want %q", retrieved.Path, "/path/to/findme")
	}
}

func TestSQLDataStore_GetAppByName_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem and domain
	ecosystem := &models.Ecosystem{Name: "app-notfound-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-notfound-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	_, err := ds.GetAppByName(domain.ID, "nonexistent")
	if err == nil {
		t.Errorf("GetAppByName() expected error for nonexistent app")
	}
}

func TestSQLDataStore_GetAppByID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "app-getbyid-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-getbyid-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "getbyid-app",
		Path:     "/path/to/getbyid",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetAppByID(app.ID)
	if err != nil {
		t.Fatalf("GetAppByID() error = %v", err)
	}

	if retrieved.ID != app.ID {
		t.Errorf("GetAppByID() ID = %d, want %d", retrieved.ID, app.ID)
	}
}

func TestSQLDataStore_UpdateApp(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "app-update-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-update-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "update-app",
		Path:     "/original/path",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Update the app
	app.Path = "/updated/path"
	app.Description = sql.NullString{String: "Updated description", Valid: true}

	if err := ds.UpdateApp(app); err != nil {
		t.Fatalf("UpdateApp() error = %v", err)
	}

	// Verify update
	retrieved, err := ds.GetAppByID(app.ID)
	if err != nil {
		t.Fatalf("Verification error: %v", err)
	}

	if retrieved.Path != "/updated/path" {
		t.Errorf("UpdateApp() Path = %q, want %q", retrieved.Path, "/updated/path")
	}
	if retrieved.Description.String != "Updated description" {
		t.Errorf("UpdateApp() Description = %q, want %q", retrieved.Description.String, "Updated description")
	}
}

func TestSQLDataStore_DeleteApp(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "app-delete-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-delete-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "delete-app",
		Path:     "/path/to/delete",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	if err := ds.DeleteApp(app.ID); err != nil {
		t.Fatalf("DeleteApp() error = %v", err)
	}

	// Verify deletion
	_, err := ds.GetAppByID(app.ID)
	if err == nil {
		t.Errorf("DeleteApp() app should not exist after deletion")
	}
}

func TestSQLDataStore_ListAppsByDomain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "app-list-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-list-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create multiple apps
	for i := 1; i <= 3; i++ {
		app := &models.App{
			DomainID: domain.ID,
			Name:     "list-app-" + string(rune('0'+i)),
			Path:     "/path/" + string(rune('0'+i)),
		}
		if err := ds.CreateApp(app); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	apps, err := ds.ListAppsByDomain(domain.ID)
	if err != nil {
		t.Fatalf("ListAppsByDomain() error = %v", err)
	}

	if len(apps) != 3 {
		t.Errorf("ListAppsByDomain() returned %d apps, want 3", len(apps))
	}
}

func TestSQLDataStore_ListAllApps(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "app-listall-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create two domains with apps
	for d := 1; d <= 2; d++ {
		domain := &models.Domain{
			EcosystemID: ecosystem.ID,
			Name:        "listall-domain-" + string(rune('0'+d)),
		}
		if err := ds.CreateDomain(domain); err != nil {
			t.Fatalf("Setup error: %v", err)
		}

		for a := 1; a <= 2; a++ {
			app := &models.App{
				DomainID: domain.ID,
				Name:     "app-" + string(rune('0'+a)),
				Path:     "/path/" + string(rune('0'+a)),
			}
			if err := ds.CreateApp(app); err != nil {
				t.Fatalf("Setup error: %v", err)
			}
		}
	}

	apps, err := ds.ListAllApps()
	if err != nil {
		t.Fatalf("ListAllApps() error = %v", err)
	}

	if len(apps) != 4 {
		t.Errorf("ListAllApps() returned %d apps, want 4", len(apps))
	}
}

// =============================================================================
// App Language/BuildConfig Tests
// =============================================================================

func TestSQLDataStore_CreateApp_WithLanguage(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "app-lang-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-lang-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create app with language config (JSON serialized)
	langJSON := `{"name":"golang","version":"1.22"}`
	app := &models.App{
		DomainID: domain.ID,
		Name:     "lang-test-app",
		Path:     "/path/to/golang-app",
		Language: sql.NullString{String: langJSON, Valid: true},
	}

	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("CreateApp() error = %v", err)
	}

	// Retrieve and verify
	retrieved, err := ds.GetAppByID(app.ID)
	if err != nil {
		t.Fatalf("GetAppByID() error = %v", err)
	}

	if !retrieved.Language.Valid {
		t.Error("Retrieved app Language should be valid")
	}
	if retrieved.Language.String != langJSON {
		t.Errorf("Language = %q, want %q", retrieved.Language.String, langJSON)
	}
}

func TestSQLDataStore_CreateApp_WithBuildConfig(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "app-build-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-build-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create app with build config (JSON serialized)
	buildJSON := `{"dockerfile":"Dockerfile.dev","args":{"BUILD_ENV":"development","DEBUG":"true"}}`
	app := &models.App{
		DomainID:    domain.ID,
		Name:        "build-test-app",
		Path:        "/path/to/build-app",
		BuildConfig: sql.NullString{String: buildJSON, Valid: true},
	}

	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("CreateApp() error = %v", err)
	}

	// Retrieve and verify
	retrieved, err := ds.GetAppByID(app.ID)
	if err != nil {
		t.Fatalf("GetAppByID() error = %v", err)
	}

	if !retrieved.BuildConfig.Valid {
		t.Error("Retrieved app BuildConfig should be valid")
	}
	if retrieved.BuildConfig.String != buildJSON {
		t.Errorf("BuildConfig = %q, want %q", retrieved.BuildConfig.String, buildJSON)
	}
}

func TestSQLDataStore_UpdateApp_Language(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "app-updatelang-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-updatelang-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create app without language initially
	app := &models.App{
		DomainID: domain.ID,
		Name:     "update-lang-app",
		Path:     "/path/to/app",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("CreateApp() error = %v", err)
	}

	// Update with language
	langJSON := `{"name":"python","version":"3.11"}`
	app.Language = sql.NullString{String: langJSON, Valid: true}

	if err := ds.UpdateApp(app); err != nil {
		t.Fatalf("UpdateApp() error = %v", err)
	}

	// Verify update
	retrieved, err := ds.GetAppByID(app.ID)
	if err != nil {
		t.Fatalf("GetAppByID() error = %v", err)
	}

	if !retrieved.Language.Valid {
		t.Error("Updated app Language should be valid")
	}
	if retrieved.Language.String != langJSON {
		t.Errorf("Language = %q, want %q", retrieved.Language.String, langJSON)
	}
}

func TestSQLDataStore_App_PreservesLanguageAndBuildConfig(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "app-preserve-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "app-preserve-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create app with both language and build config
	langJSON := `{"name":"nodejs","version":"20"}`
	buildJSON := `{"dockerfile":"Dockerfile","args":{"NODE_ENV":"development"},"target":"dev"}`

	app := &models.App{
		DomainID:    domain.ID,
		Name:        "full-config-app",
		Path:        "/path/to/node-app",
		Description: sql.NullString{String: "Full config test", Valid: true},
		Language:    sql.NullString{String: langJSON, Valid: true},
		BuildConfig: sql.NullString{String: buildJSON, Valid: true},
	}

	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("CreateApp() error = %v", err)
	}

	// Test GetAppByName
	byName, err := ds.GetAppByName(domain.ID, "full-config-app")
	if err != nil {
		t.Fatalf("GetAppByName() error = %v", err)
	}
	if byName.Language.String != langJSON {
		t.Errorf("GetAppByName() Language = %q, want %q", byName.Language.String, langJSON)
	}
	if byName.BuildConfig.String != buildJSON {
		t.Errorf("GetAppByName() BuildConfig = %q, want %q", byName.BuildConfig.String, buildJSON)
	}

	// Test GetAppByNameGlobal
	byGlobal, err := ds.GetAppByNameGlobal("full-config-app")
	if err != nil {
		t.Fatalf("GetAppByNameGlobal() error = %v", err)
	}
	if byGlobal.Language.String != langJSON {
		t.Errorf("GetAppByNameGlobal() Language = %q, want %q", byGlobal.Language.String, langJSON)
	}
	if byGlobal.BuildConfig.String != buildJSON {
		t.Errorf("GetAppByNameGlobal() BuildConfig = %q, want %q", byGlobal.BuildConfig.String, buildJSON)
	}

	// Test GetAppByID
	byID, err := ds.GetAppByID(app.ID)
	if err != nil {
		t.Fatalf("GetAppByID() error = %v", err)
	}
	if byID.Language.String != langJSON {
		t.Errorf("GetAppByID() Language = %q, want %q", byID.Language.String, langJSON)
	}
	if byID.BuildConfig.String != buildJSON {
		t.Errorf("GetAppByID() BuildConfig = %q, want %q", byID.BuildConfig.String, buildJSON)
	}

	// Test ListAppsByDomain
	domainApps, err := ds.ListAppsByDomain(domain.ID)
	if err != nil {
		t.Fatalf("ListAppsByDomain() error = %v", err)
	}
	if len(domainApps) != 1 {
		t.Fatalf("ListAppsByDomain() returned %d apps, want 1", len(domainApps))
	}
	if domainApps[0].Language.String != langJSON {
		t.Errorf("ListAppsByDomain() Language = %q, want %q", domainApps[0].Language.String, langJSON)
	}
	if domainApps[0].BuildConfig.String != buildJSON {
		t.Errorf("ListAppsByDomain() BuildConfig = %q, want %q", domainApps[0].BuildConfig.String, buildJSON)
	}

	// Test ListAllApps
	allApps, err := ds.ListAllApps()
	if err != nil {
		t.Fatalf("ListAllApps() error = %v", err)
	}
	var found *models.App
	for _, a := range allApps {
		if a.Name == "full-config-app" {
			found = a
			break
		}
	}
	if found == nil {
		t.Fatal("ListAllApps() did not return full-config-app")
	}
	if found.Language.String != langJSON {
		t.Errorf("ListAllApps() Language = %q, want %q", found.Language.String, langJSON)
	}
	if found.BuildConfig.String != buildJSON {
		t.Errorf("ListAllApps() BuildConfig = %q, want %q", found.BuildConfig.String, buildJSON)
	}
}

// =============================================================================
// Project CRUD Tests
// =============================================================================

func TestSQLDataStore_CreateProject(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{
		Name: "test-project",
		Path: "/path/to/project",
		Description: sql.NullString{
			String: "Test description",
			Valid:  true,
		},
	}

	err := ds.CreateProject(project)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	if project.ID == 0 {
		t.Errorf("CreateProject() did not set project.ID")
	}
}

func TestSQLDataStore_GetProjectByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create a project first
	project := &models.Project{
		Name: "findme-project",
		Path: "/path/to/findme",
	}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup CreateProject() error = %v", err)
	}

	// Retrieve by name
	retrieved, err := ds.GetProjectByName("findme-project")
	if err != nil {
		t.Fatalf("GetProjectByName() error = %v", err)
	}

	if retrieved.Name != "findme-project" {
		t.Errorf("GetProjectByName() Name = %q, want %q", retrieved.Name, "findme-project")
	}
	if retrieved.Path != "/path/to/findme" {
		t.Errorf("GetProjectByName() Path = %q, want %q", retrieved.Path, "/path/to/findme")
	}
}

func TestSQLDataStore_GetProjectByName_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	_, err := ds.GetProjectByName("nonexistent")
	if err == nil {
		t.Errorf("GetProjectByName() expected error for nonexistent project")
	}
}

func TestSQLDataStore_GetProjectByID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{
		Name: "getbyid-project",
		Path: "/path/getbyid",
	}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetProjectByID(project.ID)
	if err != nil {
		t.Fatalf("GetProjectByID() error = %v", err)
	}

	if retrieved.ID != project.ID {
		t.Errorf("GetProjectByID() ID = %d, want %d", retrieved.ID, project.ID)
	}
}

func TestSQLDataStore_UpdateProject(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{
		Name: "update-project",
		Path: "/original/path",
	}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Update the project
	project.Path = "/updated/path"
	project.Description = sql.NullString{String: "Updated description", Valid: true}

	if err := ds.UpdateProject(project); err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}

	// Verify update
	retrieved, err := ds.GetProjectByID(project.ID)
	if err != nil {
		t.Fatalf("Verification error: %v", err)
	}

	if retrieved.Path != "/updated/path" {
		t.Errorf("UpdateProject() Path = %q, want %q", retrieved.Path, "/updated/path")
	}
	if retrieved.Description.String != "Updated description" {
		t.Errorf("UpdateProject() Description = %q, want %q", retrieved.Description.String, "Updated description")
	}
}

func TestSQLDataStore_DeleteProject(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{
		Name: "delete-project",
		Path: "/to/delete",
	}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	if err := ds.DeleteProject("delete-project"); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	// Verify deletion
	_, err := ds.GetProjectByName("delete-project")
	if err == nil {
		t.Errorf("DeleteProject() project should not exist after deletion")
	}
}

func TestSQLDataStore_ListProjects(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create multiple projects
	for i := 1; i <= 3; i++ {
		project := &models.Project{
			Name: "list-project-" + string(rune('0'+i)),
			Path: "/path/" + string(rune('0'+i)),
		}
		if err := ds.CreateProject(project); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	projects, err := ds.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	if len(projects) != 3 {
		t.Errorf("ListProjects() returned %d projects, want 3", len(projects))
	}
}

// =============================================================================
// Workspace CRUD Tests
// =============================================================================

func TestSQLDataStore_CreateWorkspace(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create a project first
	project := &models.Project{Name: "ws-project", Path: "/ws/path"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		AppID:     project.ID, // Using project.ID as AppID during migration
		Name:      "test-workspace",
		ImageName: "test:latest",
		Status:    "stopped",
	}

	err := ds.CreateWorkspace(workspace)
	if err != nil {
		t.Fatalf("CreateWorkspace() error = %v", err)
	}

	if workspace.ID == 0 {
		t.Errorf("CreateWorkspace() did not set workspace.ID")
	}
}

func TestSQLDataStore_GetWorkspaceByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "ws-project2", Path: "/ws/path2"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		AppID:     project.ID, // Using project.ID as AppID during migration
		Name:      "findme-ws",
		ImageName: "image:v1",
		Status:    "running",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetWorkspaceByName(project.ID, "findme-ws")
	if err != nil {
		t.Fatalf("GetWorkspaceByName() error = %v", err)
	}

	if retrieved.Name != "findme-ws" {
		t.Errorf("GetWorkspaceByName() Name = %q, want %q", retrieved.Name, "findme-ws")
	}
	if retrieved.ImageName != "image:v1" {
		t.Errorf("GetWorkspaceByName() ImageName = %q, want %q", retrieved.ImageName, "image:v1")
	}
}

func TestSQLDataStore_GetWorkspaceByID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "ws-project3", Path: "/ws/path3"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		AppID:     project.ID, // Using project.ID as AppID during migration
		Name:      "getbyid-ws",
		ImageName: "test:v2",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspaceByID() error = %v", err)
	}

	if retrieved.ID != workspace.ID {
		t.Errorf("GetWorkspaceByID() ID = %d, want %d", retrieved.ID, workspace.ID)
	}
}

func TestSQLDataStore_UpdateWorkspace(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "ws-project4", Path: "/ws/path4"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		AppID:     project.ID, // Using project.ID as AppID during migration
		Name:      "update-ws",
		ImageName: "old:image",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Update
	workspace.ImageName = "new:image"
	workspace.Status = "running"

	if err := ds.UpdateWorkspace(workspace); err != nil {
		t.Fatalf("UpdateWorkspace() error = %v", err)
	}

	// Verify
	retrieved, err := ds.GetWorkspaceByID(workspace.ID)
	if err != nil {
		t.Fatalf("Verification error: %v", err)
	}

	if retrieved.ImageName != "new:image" {
		t.Errorf("UpdateWorkspace() ImageName = %q, want %q", retrieved.ImageName, "new:image")
	}
	if retrieved.Status != "running" {
		t.Errorf("UpdateWorkspace() Status = %q, want %q", retrieved.Status, "running")
	}
}

func TestSQLDataStore_DeleteWorkspace(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "ws-project5", Path: "/ws/path5"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		AppID:     project.ID, // Using project.ID as AppID during migration
		Name:      "delete-ws",
		ImageName: "img:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	if err := ds.DeleteWorkspace(workspace.ID); err != nil {
		t.Fatalf("DeleteWorkspace() error = %v", err)
	}

	_, err := ds.GetWorkspaceByID(workspace.ID)
	if err == nil {
		t.Errorf("DeleteWorkspace() workspace should not exist after deletion")
	}
}

func TestSQLDataStore_ListWorkspacesByApp(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "ws-project6", Path: "/ws/path6"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create multiple workspaces
	for i := 1; i <= 3; i++ {
		ws := &models.Workspace{
			AppID:     project.ID, // Using project.ID as AppID during migration
			Name:      "ws-" + string(rune('0'+i)),
			ImageName: "img:v" + string(rune('0'+i)),
			Status:    "stopped",
		}
		if err := ds.CreateWorkspace(ws); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	workspaces, err := ds.ListWorkspacesByApp(project.ID)
	if err != nil {
		t.Fatalf("ListWorkspacesByApp() error = %v", err)
	}

	if len(workspaces) != 3 {
		t.Errorf("ListWorkspacesByApp() returned %d workspaces, want 3", len(workspaces))
	}
}

func TestSQLDataStore_ListAllWorkspaces(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create two projects with workspaces
	for p := 1; p <= 2; p++ {
		project := &models.Project{
			Name: "project-" + string(rune('0'+p)),
			Path: "/path/" + string(rune('0'+p)),
		}
		if err := ds.CreateProject(project); err != nil {
			t.Fatalf("Setup error: %v", err)
		}

		for w := 1; w <= 2; w++ {
			ws := &models.Workspace{
				AppID:     project.ID, // Using project.ID as AppID during migration
				Name:      "ws-" + string(rune('0'+w)),
				ImageName: "img:latest",
				Status:    "stopped",
			}
			if err := ds.CreateWorkspace(ws); err != nil {
				t.Fatalf("Setup error: %v", err)
			}
		}
	}

	workspaces, err := ds.ListAllWorkspaces()
	if err != nil {
		t.Fatalf("ListAllWorkspaces() error = %v", err)
	}

	if len(workspaces) != 4 {
		t.Errorf("ListAllWorkspaces() returned %d workspaces, want 4", len(workspaces))
	}
}

// =============================================================================
// Context Tests
// =============================================================================

func TestSQLDataStore_Context(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create a project and workspace
	project := &models.Project{Name: "ctx-project", Path: "/ctx/path"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		AppID:     project.ID, // Using project.ID as AppID during migration
		Name:      "ctx-workspace",
		ImageName: "ctx:img",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Set active project
	if err := ds.SetActiveProject(&project.ID); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}

	// Get context and verify
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveProjectID == nil || *ctx.ActiveProjectID != project.ID {
		t.Errorf("GetContext() ActiveProjectID = %v, want %d", ctx.ActiveProjectID, project.ID)
	}

	// Set active workspace
	if err := ds.SetActiveWorkspace(&workspace.ID); err != nil {
		t.Fatalf("SetActiveWorkspace() error = %v", err)
	}

	ctx, err = ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveWorkspaceID == nil || *ctx.ActiveWorkspaceID != workspace.ID {
		t.Errorf("GetContext() ActiveWorkspaceID = %v, want %d", ctx.ActiveWorkspaceID, workspace.ID)
	}
}

func TestSQLDataStore_ClearContext(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	project := &models.Project{Name: "clear-ctx-project", Path: "/clear"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Set then clear
	if err := ds.SetActiveProject(&project.ID); err != nil {
		t.Fatalf("SetActiveProject() error = %v", err)
	}

	if err := ds.SetActiveProject(nil); err != nil {
		t.Fatalf("SetActiveProject(nil) error = %v", err)
	}

	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveProjectID != nil {
		t.Errorf("SetActiveProject(nil) should clear ActiveProjectID")
	}
}

func TestSQLDataStore_Context_Ecosystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create an ecosystem
	ecosystem := &models.Ecosystem{Name: "ctx-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Set active ecosystem
	if err := ds.SetActiveEcosystem(&ecosystem.ID); err != nil {
		t.Fatalf("SetActiveEcosystem() error = %v", err)
	}

	// Get context and verify
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveEcosystemID == nil || *ctx.ActiveEcosystemID != ecosystem.ID {
		t.Errorf("GetContext() ActiveEcosystemID = %v, want %d", ctx.ActiveEcosystemID, ecosystem.ID)
	}

	// Clear active ecosystem
	if err := ds.SetActiveEcosystem(nil); err != nil {
		t.Fatalf("SetActiveEcosystem(nil) error = %v", err)
	}

	ctx, err = ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() after clear error = %v", err)
	}

	if ctx.ActiveEcosystemID != nil {
		t.Errorf("SetActiveEcosystem(nil) should clear ActiveEcosystemID")
	}
}

func TestSQLDataStore_Context_Domain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem and domain
	ecosystem := &models.Ecosystem{Name: "ctx-domain-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "ctx-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Set active domain
	if err := ds.SetActiveDomain(&domain.ID); err != nil {
		t.Fatalf("SetActiveDomain() error = %v", err)
	}

	// Get context and verify
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveDomainID == nil || *ctx.ActiveDomainID != domain.ID {
		t.Errorf("GetContext() ActiveDomainID = %v, want %d", ctx.ActiveDomainID, domain.ID)
	}

	// Clear active domain
	if err := ds.SetActiveDomain(nil); err != nil {
		t.Fatalf("SetActiveDomain(nil) error = %v", err)
	}

	ctx, err = ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() after clear error = %v", err)
	}

	if ctx.ActiveDomainID != nil {
		t.Errorf("SetActiveDomain(nil) should clear ActiveDomainID")
	}
}

func TestSQLDataStore_Context_App(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create ecosystem, domain, and app
	ecosystem := &models.Ecosystem{Name: "ctx-app-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "ctx-app-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "ctx-app",
		Path:     "/path/to/ctx-app",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Set active app
	if err := ds.SetActiveApp(&app.ID); err != nil {
		t.Fatalf("SetActiveApp() error = %v", err)
	}

	// Get context and verify
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveAppID == nil || *ctx.ActiveAppID != app.ID {
		t.Errorf("GetContext() ActiveAppID = %v, want %d", ctx.ActiveAppID, app.ID)
	}

	// Clear active app
	if err := ds.SetActiveApp(nil); err != nil {
		t.Fatalf("SetActiveApp(nil) error = %v", err)
	}

	ctx, err = ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() after clear error = %v", err)
	}

	if ctx.ActiveAppID != nil {
		t.Errorf("SetActiveApp(nil) should clear ActiveAppID")
	}
}

func TestSQLDataStore_Context_FullHierarchy(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create full hierarchy: Ecosystem -> Domain -> App -> Workspace
	ecosystem := &models.Ecosystem{Name: "full-ctx-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "full-ctx-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	app := &models.App{
		DomainID: domain.ID,
		Name:     "full-ctx-app",
		Path:     "/path/to/full-ctx-app",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	project := &models.Project{Name: "full-ctx-project", Path: "/full/ctx"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		AppID:     project.ID, // Using project.ID as AppID during migration
		Name:      "full-ctx-workspace",
		ImageName: "full:img",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Set all active entities
	if err := ds.SetActiveEcosystem(&ecosystem.ID); err != nil {
		t.Fatalf("SetActiveEcosystem() error = %v", err)
	}
	if err := ds.SetActiveDomain(&domain.ID); err != nil {
		t.Fatalf("SetActiveDomain() error = %v", err)
	}
	if err := ds.SetActiveApp(&app.ID); err != nil {
		t.Fatalf("SetActiveApp() error = %v", err)
	}
	if err := ds.SetActiveWorkspace(&workspace.ID); err != nil {
		t.Fatalf("SetActiveWorkspace() error = %v", err)
	}

	// Get context and verify all
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveEcosystemID == nil || *ctx.ActiveEcosystemID != ecosystem.ID {
		t.Errorf("GetContext() ActiveEcosystemID = %v, want %d", ctx.ActiveEcosystemID, ecosystem.ID)
	}
	if ctx.ActiveDomainID == nil || *ctx.ActiveDomainID != domain.ID {
		t.Errorf("GetContext() ActiveDomainID = %v, want %d", ctx.ActiveDomainID, domain.ID)
	}
	if ctx.ActiveAppID == nil || *ctx.ActiveAppID != app.ID {
		t.Errorf("GetContext() ActiveAppID = %v, want %d", ctx.ActiveAppID, app.ID)
	}
	if ctx.ActiveWorkspaceID == nil || *ctx.ActiveWorkspaceID != workspace.ID {
		t.Errorf("GetContext() ActiveWorkspaceID = %v, want %d", ctx.ActiveWorkspaceID, workspace.ID)
	}
}

// =============================================================================
// Plugin Tests
// =============================================================================

func TestSQLDataStore_CreatePlugin(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	plugin := &models.NvimPluginDB{
		Name:     "test-plugin",
		Repo:     "user/test-plugin",
		Category: sql.NullString{String: "editor", Valid: true},
		Enabled:  true,
	}

	err := ds.CreatePlugin(plugin)
	if err != nil {
		t.Fatalf("CreatePlugin() error = %v", err)
	}

	if plugin.ID == 0 {
		t.Errorf("CreatePlugin() did not set plugin.ID")
	}
}

func TestSQLDataStore_GetPluginByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	plugin := &models.NvimPluginDB{
		Name:     "findme-plugin",
		Repo:     "user/findme-plugin",
		Category: sql.NullString{String: "lsp", Valid: true},
		Enabled:  true,
	}
	if err := ds.CreatePlugin(plugin); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	retrieved, err := ds.GetPluginByName("findme-plugin")
	if err != nil {
		t.Fatalf("GetPluginByName() error = %v", err)
	}

	if retrieved.Repo != "user/findme-plugin" {
		t.Errorf("GetPluginByName() Repo = %q, want %q", retrieved.Repo, "user/findme-plugin")
	}
}

func TestSQLDataStore_ListPlugins(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	for i := 1; i <= 3; i++ {
		plugin := &models.NvimPluginDB{
			Name:     "plugin-" + string(rune('0'+i)),
			Repo:     "user/plugin-" + string(rune('0'+i)),
			Category: sql.NullString{String: "test", Valid: true},
			Enabled:  true,
		}
		if err := ds.CreatePlugin(plugin); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	plugins, err := ds.ListPlugins()
	if err != nil {
		t.Fatalf("ListPlugins() error = %v", err)
	}

	if len(plugins) != 3 {
		t.Errorf("ListPlugins() returned %d plugins, want 3", len(plugins))
	}
}

func TestSQLDataStore_ListPluginsByCategory(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create plugins with different categories
	categories := []string{"lsp", "lsp", "ui", "editor"}
	for i, cat := range categories {
		plugin := &models.NvimPluginDB{
			Name:     "cat-plugin-" + string(rune('0'+i)),
			Repo:     "user/cat-plugin-" + string(rune('0'+i)),
			Category: sql.NullString{String: cat, Valid: true},
			Enabled:  true,
		}
		if err := ds.CreatePlugin(plugin); err != nil {
			t.Fatalf("Setup error: %v", err)
		}
	}

	lspPlugins, err := ds.ListPluginsByCategory("lsp")
	if err != nil {
		t.Fatalf("ListPluginsByCategory() error = %v", err)
	}

	if len(lspPlugins) != 2 {
		t.Errorf("ListPluginsByCategory(lsp) returned %d plugins, want 2", len(lspPlugins))
	}
}

// =============================================================================
// Workspace-Plugin Association Tests
// =============================================================================

func TestSQLDataStore_WorkspacePluginAssociation(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Setup
	project := &models.Project{Name: "wp-project", Path: "/wp/path"}
	if err := ds.CreateProject(project); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	workspace := &models.Workspace{
		AppID:     project.ID, // Using project.ID as AppID during migration
		Name:      "wp-workspace",
		ImageName: "img:latest",
		Status:    "stopped",
	}
	if err := ds.CreateWorkspace(workspace); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	plugin := &models.NvimPluginDB{
		Name:    "wp-plugin",
		Repo:    "user/wp-plugin",
		Enabled: true,
	}
	if err := ds.CreatePlugin(plugin); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Add plugin to workspace
	if err := ds.AddPluginToWorkspace(workspace.ID, plugin.ID); err != nil {
		t.Fatalf("AddPluginToWorkspace() error = %v", err)
	}

	// Get workspace plugins
	plugins, err := ds.GetWorkspacePlugins(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspacePlugins() error = %v", err)
	}

	if len(plugins) != 1 {
		t.Errorf("GetWorkspacePlugins() returned %d plugins, want 1", len(plugins))
	}

	// Remove plugin from workspace
	if err := ds.RemovePluginFromWorkspace(workspace.ID, plugin.ID); err != nil {
		t.Fatalf("RemovePluginFromWorkspace() error = %v", err)
	}

	plugins, err = ds.GetWorkspacePlugins(workspace.ID)
	if err != nil {
		t.Fatalf("GetWorkspacePlugins() after remove error = %v", err)
	}

	if len(plugins) != 0 {
		t.Errorf("GetWorkspacePlugins() after remove returned %d plugins, want 0", len(plugins))
	}
}

// =============================================================================
// Driver and Health Tests
// =============================================================================

func TestSQLDataStore_Driver(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	driver := ds.Driver()
	if driver == nil {
		t.Errorf("Driver() returned nil")
	}

	if driver.Type() != DriverMemory {
		t.Errorf("Driver().Type() = %v, want %v", driver.Type(), DriverMemory)
	}
}

func TestSQLDataStore_Ping(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	if err := ds.Ping(); err != nil {
		t.Errorf("Ping() error = %v", err)
	}
}

func TestSQLDataStore_Close(t *testing.T) {
	ds := createTestDataStore(t)

	if err := ds.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}

	// Subsequent operations should fail
	if err := ds.Ping(); err == nil {
		t.Errorf("Ping() after Close() should fail")
	}
}

// =============================================================================
// Interface Compliance Tests
// =============================================================================

func TestSQLDataStore_ImplementsDataStore(t *testing.T) {
	var _ DataStore = (*SQLDataStore)(nil)
}

// =============================================================================
// Migration Schema Tests (v0.8.2 fixes)
// =============================================================================

// TestSQLDataStore_MigrationSchema_AppsTableHasLanguageAndBuildConfig verifies
// that the apps table schema includes language and build_config columns.
// This test was added to prevent regression of the v0.8.2 fix where these
// columns were missing from the migration file 007_add_apps.up.sql.
func TestSQLDataStore_MigrationSchema_AppsTableHasLanguageAndBuildConfig(t *testing.T) {
	// Create a driver with a fresh database
	cfg := DriverConfig{Type: DriverMemory}
	driver, err := NewMemorySQLiteDriver(cfg)
	if err != nil {
		t.Fatalf("Failed to create test driver: %v", err)
	}
	defer driver.Close()

	if err := driver.Connect(); err != nil {
		t.Fatalf("Failed to connect test driver: %v", err)
	}

	// Apply the exact schema that should match migration 007_add_apps.up.sql
	// This verifies that apps can be created with language and build_config columns
	migrationSQL := `
		CREATE TABLE IF NOT EXISTS ecosystems (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS domains (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ecosystem_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (ecosystem_id) REFERENCES ecosystems(id),
			UNIQUE(ecosystem_id, name)
		);
		
		CREATE TABLE IF NOT EXISTS apps (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			path TEXT NOT NULL,
			description TEXT,
			language TEXT,
			build_config TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (domain_id) REFERENCES domains(id) ON DELETE CASCADE,
			UNIQUE(domain_id, name)
		);
	`

	// Execute migration-style SQL
	_, err = driver.Execute(migrationSQL)
	if err != nil {
		t.Fatalf("Migration SQL failed: %v", err)
	}

	// Create DataStore
	ds := NewSQLDataStore(driver, nil)

	// Create ecosystem
	ecosystem := &models.Ecosystem{Name: "migration-test-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("CreateEcosystem() error = %v", err)
	}

	// Create domain
	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "migration-test-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("CreateDomain() error = %v", err)
	}

	// Create app with language and build_config - this is the critical test
	// This would have failed before the v0.8.2 fix with:
	// "table apps has no column named language"
	langJSON := `{"name":"golang","version":"1.25"}`
	buildJSON := `{"dockerfile":"Dockerfile","args":{"CGO_ENABLED":"1"}}`

	app := &models.App{
		DomainID:    domain.ID,
		Name:        "migration-test-app",
		Path:        "/path/to/app",
		Language:    sql.NullString{String: langJSON, Valid: true},
		BuildConfig: sql.NullString{String: buildJSON, Valid: true},
	}

	err = ds.CreateApp(app)
	if err != nil {
		t.Fatalf("CreateApp() with language and build_config failed: %v\n"+
			"This indicates the migration schema is missing required columns.", err)
	}

	// Verify the data was stored correctly
	retrieved, err := ds.GetAppByID(app.ID)
	if err != nil {
		t.Fatalf("GetAppByID() error = %v", err)
	}

	if !retrieved.Language.Valid || retrieved.Language.String != langJSON {
		t.Errorf("Language not persisted correctly: got %q, want %q",
			retrieved.Language.String, langJSON)
	}

	if !retrieved.BuildConfig.Valid || retrieved.BuildConfig.String != buildJSON {
		t.Errorf("BuildConfig not persisted correctly: got %q, want %q",
			retrieved.BuildConfig.String, buildJSON)
	}
}

// =============================================================================
// Error Message Tests (v0.8.2 fixes)
// =============================================================================

// TestSQLDataStore_CreateApp_ErrorNotDuplicated verifies that CreateApp errors
// do not contain duplicate error wrapping like "failed to create app: failed to create app:"
// This test was added after the v0.8.2 fix that removed duplicate error wrapping.
func TestSQLDataStore_CreateApp_ErrorNotDuplicated(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Setup: Create ecosystem and domain
	ecosystem := &models.Ecosystem{Name: "error-test-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup CreateEcosystem() error = %v", err)
	}

	domain := &models.Domain{
		EcosystemID: ecosystem.ID,
		Name:        "error-test-domain",
	}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup CreateDomain() error = %v", err)
	}

	// Create first app
	app1 := &models.App{
		DomainID: domain.ID,
		Name:     "duplicate-app",
		Path:     "/path/to/app",
	}
	if err := ds.CreateApp(app1); err != nil {
		t.Fatalf("CreateApp() first app error = %v", err)
	}

	// Try to create duplicate app - should fail with unique constraint
	app2 := &models.App{
		DomainID: domain.ID,
		Name:     "duplicate-app", // Same name, same domain
		Path:     "/path/to/other",
	}

	err := ds.CreateApp(app2)
	if err == nil {
		t.Fatal("CreateApp() with duplicate name should fail")
	}

	errMsg := err.Error()

	// The error should NOT contain duplicated phrases
	// Before fix: "failed to create app: failed to create app: UNIQUE constraint failed"
	// After fix: "UNIQUE constraint failed" (raw error from sqlite)
	if strings.Contains(errMsg, "failed to create app: failed to create app") {
		t.Errorf("CreateApp() error message contains duplicate wrapping: %q", errMsg)
	}

	// Verify it's still a meaningful error
	if !strings.Contains(errMsg, "UNIQUE") {
		t.Errorf("CreateApp() error should mention UNIQUE constraint: %q", errMsg)
	}
}
