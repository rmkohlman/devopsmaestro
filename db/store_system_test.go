package db

import (
	"database/sql"
	"testing"

	"devopsmaestro/models"
)

// =============================================================================
// System CRUD Tests
// =============================================================================

func TestSQLDataStore_CreateSystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create parent hierarchy
	ecosystem := &models.Ecosystem{Name: "system-test-ecosystem"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup CreateEcosystem() error = %v", err)
	}
	domain := &models.Domain{EcosystemID: validNullInt64(ecosystem.ID), Name: "system-test-domain"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup CreateDomain() error = %v", err)
	}

	system := &models.System{
		EcosystemID: sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true},
		DomainID:    sql.NullInt64{Int64: int64(domain.ID), Valid: true},
		Name:        "test-system",
		Description: sql.NullString{String: "Test system description", Valid: true},
	}

	err := ds.CreateSystem(system)
	if err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	if system.ID == 0 {
		t.Errorf("CreateSystem() did not set system.ID")
	}
}

func TestSQLDataStore_CreateSystem_NoDomain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	system := &models.System{
		Name:        "orphan-system",
		Description: sql.NullString{String: "System without domain", Valid: true},
	}

	err := ds.CreateSystem(system)
	if err != nil {
		t.Fatalf("CreateSystem() error = %v", err)
	}

	if system.ID == 0 {
		t.Errorf("CreateSystem() did not set system.ID")
	}
}

func TestSQLDataStore_GetSystemByID(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "system-getbyid-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}
	domain := &models.Domain{EcosystemID: validNullInt64(ecosystem.ID), Name: "system-getbyid-dom"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	system := &models.System{
		EcosystemID: sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true},
		DomainID:    sql.NullInt64{Int64: int64(domain.ID), Valid: true},
		Name:        "getbyid-system",
	}
	if err := ds.CreateSystem(system); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	retrieved, err := ds.GetSystemByID(system.ID)
	if err != nil {
		t.Fatalf("GetSystemByID() error = %v", err)
	}

	if retrieved.ID != system.ID {
		t.Errorf("GetSystemByID() ID = %d, want %d", retrieved.ID, system.ID)
	}
	if retrieved.Name != "getbyid-system" {
		t.Errorf("GetSystemByID() Name = %q, want %q", retrieved.Name, "getbyid-system")
	}
}

func TestSQLDataStore_GetSystemByID_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	_, err := ds.GetSystemByID(9999)
	if err == nil {
		t.Errorf("GetSystemByID() expected error for nonexistent system")
	}
}

func TestSQLDataStore_GetSystemByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "system-getbyname-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}
	domain := &models.Domain{EcosystemID: validNullInt64(ecosystem.ID), Name: "system-getbyname-dom"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	system := &models.System{
		EcosystemID: sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true},
		DomainID:    sql.NullInt64{Int64: int64(domain.ID), Valid: true},
		Name:        "findme-system",
		Description: sql.NullString{String: "Find me", Valid: true},
	}
	if err := ds.CreateSystem(system); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	domainID := sql.NullInt64{Int64: int64(domain.ID), Valid: true}
	retrieved, err := ds.GetSystemByName(domainID, "findme-system")
	if err != nil {
		t.Fatalf("GetSystemByName() error = %v", err)
	}

	if retrieved.Name != "findme-system" {
		t.Errorf("GetSystemByName() Name = %q, want %q", retrieved.Name, "findme-system")
	}
}

func TestSQLDataStore_GetSystemByName_NullDomain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	system := &models.System{
		Name:        "null-domain-system",
		Description: sql.NullString{String: "No domain", Valid: true},
	}
	if err := ds.CreateSystem(system); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	nullDomainID := sql.NullInt64{Valid: false}
	retrieved, err := ds.GetSystemByName(nullDomainID, "null-domain-system")
	if err != nil {
		t.Fatalf("GetSystemByName() error = %v", err)
	}

	if retrieved.Name != "null-domain-system" {
		t.Errorf("GetSystemByName() Name = %q, want %q", retrieved.Name, "null-domain-system")
	}
}

func TestSQLDataStore_GetSystemByName_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	domainID := sql.NullInt64{Int64: 1, Valid: true}
	_, err := ds.GetSystemByName(domainID, "nonexistent")
	if err == nil {
		t.Errorf("GetSystemByName() expected error for nonexistent system")
	}
}

func TestSQLDataStore_UpdateSystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "system-update-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}
	domain := &models.Domain{EcosystemID: validNullInt64(ecosystem.ID), Name: "system-update-dom"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	system := &models.System{
		EcosystemID: sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true},
		DomainID:    sql.NullInt64{Int64: int64(domain.ID), Valid: true},
		Name:        "update-system",
	}
	if err := ds.CreateSystem(system); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	system.Description = sql.NullString{String: "Updated description", Valid: true}
	if err := ds.UpdateSystem(system); err != nil {
		t.Fatalf("UpdateSystem() error = %v", err)
	}

	retrieved, err := ds.GetSystemByID(system.ID)
	if err != nil {
		t.Fatalf("GetSystemByID() error = %v", err)
	}
	if retrieved.Description.String != "Updated description" {
		t.Errorf("UpdateSystem() Description = %q, want %q", retrieved.Description.String, "Updated description")
	}
}

func TestSQLDataStore_DeleteSystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	system := &models.System{Name: "delete-system"}
	if err := ds.CreateSystem(system); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	if err := ds.DeleteSystem(system.ID); err != nil {
		t.Fatalf("DeleteSystem() error = %v", err)
	}

	_, err := ds.GetSystemByID(system.ID)
	if err == nil {
		t.Errorf("DeleteSystem() system should not exist after deletion")
	}
}

func TestSQLDataStore_DeleteSystem_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.DeleteSystem(9999)
	if err == nil {
		t.Errorf("DeleteSystem() expected error for nonexistent system")
	}
}

func TestSQLDataStore_ListSystems(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "system-list-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}
	domain := &models.Domain{EcosystemID: validNullInt64(ecosystem.ID), Name: "system-list-dom"}
	if err := ds.CreateDomain(domain); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	for i := 0; i < 3; i++ {
		system := &models.System{
			DomainID: sql.NullInt64{Int64: int64(domain.ID), Valid: true},
			Name:     "list-system-" + string(rune('a'+i)),
		}
		if err := ds.CreateSystem(system); err != nil {
			t.Fatalf("Setup CreateSystem() error = %v", err)
		}
	}

	systems, err := ds.ListSystems()
	if err != nil {
		t.Fatalf("ListSystems() error = %v", err)
	}

	if len(systems) != 3 {
		t.Errorf("ListSystems() returned %d systems, want 3", len(systems))
	}
}

func TestSQLDataStore_ListSystemsByDomain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	ecosystem := &models.Ecosystem{Name: "system-listbydomain-eco"}
	if err := ds.CreateEcosystem(ecosystem); err != nil {
		t.Fatalf("Setup error: %v", err)
	}
	domain1 := &models.Domain{EcosystemID: validNullInt64(ecosystem.ID), Name: "system-listbydomain-dom1"}
	if err := ds.CreateDomain(domain1); err != nil {
		t.Fatalf("Setup error: %v", err)
	}
	domain2 := &models.Domain{EcosystemID: validNullInt64(ecosystem.ID), Name: "system-listbydomain-dom2"}
	if err := ds.CreateDomain(domain2); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	// Create 2 systems in domain1 and 1 in domain2
	for i := 0; i < 2; i++ {
		system := &models.System{
			DomainID: sql.NullInt64{Int64: int64(domain1.ID), Valid: true},
			Name:     "dom1-system-" + string(rune('a'+i)),
		}
		if err := ds.CreateSystem(system); err != nil {
			t.Fatalf("Setup CreateSystem() error = %v", err)
		}
	}
	system3 := &models.System{
		DomainID: sql.NullInt64{Int64: int64(domain2.ID), Valid: true},
		Name:     "dom2-system",
	}
	if err := ds.CreateSystem(system3); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	systems, err := ds.ListSystemsByDomain(domain1.ID)
	if err != nil {
		t.Fatalf("ListSystemsByDomain() error = %v", err)
	}

	if len(systems) != 2 {
		t.Errorf("ListSystemsByDomain() returned %d systems, want 2", len(systems))
	}
}

func TestSQLDataStore_FindSystemsByName(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Create two ecosystems with domains, each with a system named "shared-name"
	eco1 := &models.Ecosystem{Name: "find-system-eco1"}
	if err := ds.CreateEcosystem(eco1); err != nil {
		t.Fatalf("Setup error: %v", err)
	}
	dom1 := &models.Domain{EcosystemID: validNullInt64(eco1.ID), Name: "find-system-dom1"}
	if err := ds.CreateDomain(dom1); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	eco2 := &models.Ecosystem{Name: "find-system-eco2"}
	if err := ds.CreateEcosystem(eco2); err != nil {
		t.Fatalf("Setup error: %v", err)
	}
	dom2 := &models.Domain{EcosystemID: validNullInt64(eco2.ID), Name: "find-system-dom2"}
	if err := ds.CreateDomain(dom2); err != nil {
		t.Fatalf("Setup error: %v", err)
	}

	sys1 := &models.System{
		EcosystemID: sql.NullInt64{Int64: int64(eco1.ID), Valid: true},
		DomainID:    sql.NullInt64{Int64: int64(dom1.ID), Valid: true},
		Name:        "shared-name",
	}
	if err := ds.CreateSystem(sys1); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	sys2 := &models.System{
		EcosystemID: sql.NullInt64{Int64: int64(eco2.ID), Valid: true},
		DomainID:    sql.NullInt64{Int64: int64(dom2.ID), Valid: true},
		Name:        "shared-name",
	}
	if err := ds.CreateSystem(sys2); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	results, err := ds.FindSystemsByName("shared-name")
	if err != nil {
		t.Fatalf("FindSystemsByName() error = %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("FindSystemsByName() returned %d results, want 2", len(results))
	}

	// Verify hierarchy is populated
	for _, r := range results {
		if r.System == nil {
			t.Error("FindSystemsByName() System should not be nil")
		}
		if r.Domain == nil {
			t.Error("FindSystemsByName() Domain should not be nil")
		}
		if r.Ecosystem == nil {
			t.Error("FindSystemsByName() Ecosystem should not be nil")
		}
	}
}

func TestSQLDataStore_FindSystemsByName_NullParents(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// System with no domain or ecosystem
	system := &models.System{Name: "orphan-find"}
	if err := ds.CreateSystem(system); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	results, err := ds.FindSystemsByName("orphan-find")
	if err != nil {
		t.Fatalf("FindSystemsByName() error = %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("FindSystemsByName() returned %d results, want 1", len(results))
	}

	if results[0].Domain != nil {
		t.Error("FindSystemsByName() Domain should be nil for orphan system")
	}
	if results[0].Ecosystem != nil {
		t.Error("FindSystemsByName() Ecosystem should be nil for orphan system")
	}
}

func TestSQLDataStore_FindSystemsByName_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	results, err := ds.FindSystemsByName("nonexistent")
	if err != nil {
		t.Fatalf("FindSystemsByName() error = %v", err)
	}

	if len(results) != 0 {
		t.Errorf("FindSystemsByName() returned %d results, want 0", len(results))
	}
}

func TestSQLDataStore_CountSystems(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	// Initial count should be 0
	count, err := ds.CountSystems()
	if err != nil {
		t.Fatalf("CountSystems() error = %v", err)
	}
	if count != 0 {
		t.Errorf("CountSystems() = %d, want 0", count)
	}

	// Create a system and verify count
	system := &models.System{Name: "count-system"}
	if err := ds.CreateSystem(system); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	count, err = ds.CountSystems()
	if err != nil {
		t.Fatalf("CountSystems() error = %v", err)
	}
	if count != 1 {
		t.Errorf("CountSystems() = %d, want 1", count)
	}
}

// =============================================================================
// Context System Tests
// =============================================================================

func TestSQLDataStore_SetActiveSystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	system := &models.System{Name: "context-system"}
	if err := ds.CreateSystem(system); err != nil {
		t.Fatalf("Setup CreateSystem() error = %v", err)
	}

	// Set active system
	if err := ds.SetActiveSystem(&system.ID); err != nil {
		t.Fatalf("SetActiveSystem() error = %v", err)
	}

	// Verify via GetContext
	ctx, err := ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveSystemID == nil {
		t.Fatal("GetContext() ActiveSystemID should not be nil")
	}
	if *ctx.ActiveSystemID != system.ID {
		t.Errorf("GetContext() ActiveSystemID = %d, want %d", *ctx.ActiveSystemID, system.ID)
	}

	// Clear active system
	if err := ds.SetActiveSystem(nil); err != nil {
		t.Fatalf("SetActiveSystem(nil) error = %v", err)
	}

	ctx, err = ds.GetContext()
	if err != nil {
		t.Fatalf("GetContext() error = %v", err)
	}

	if ctx.ActiveSystemID != nil {
		t.Errorf("GetContext() ActiveSystemID should be nil after clearing, got %d", *ctx.ActiveSystemID)
	}
}
