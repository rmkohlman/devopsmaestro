package db

import (
	"database/sql"
	"testing"

	"devopsmaestro/models"
)

// =============================================================================
// MoveSystem tests — issue #397
// =============================================================================

// moveTestSetup creates two ecosystems each with one domain, and one system
// (under eco1/dom1) that owns one app. Returns IDs for use in tests.
type moveSetup struct {
	eco1ID, eco2ID int
	dom1ID, dom2ID int
	systemID       int
	appID          int
}

func setupMoveTest(t *testing.T, ds *SQLDataStore) moveSetup {
	t.Helper()

	eco1 := &models.Ecosystem{Name: "move-eco-1"}
	if err := ds.CreateEcosystem(eco1); err != nil {
		t.Fatalf("CreateEcosystem(eco1) error = %v", err)
	}
	eco2 := &models.Ecosystem{Name: "move-eco-2"}
	if err := ds.CreateEcosystem(eco2); err != nil {
		t.Fatalf("CreateEcosystem(eco2) error = %v", err)
	}

	dom1 := &models.Domain{EcosystemID: validNullInt64(eco1.ID), Name: "move-dom-1"}
	if err := ds.CreateDomain(dom1); err != nil {
		t.Fatalf("CreateDomain(dom1) error = %v", err)
	}
	dom2 := &models.Domain{EcosystemID: validNullInt64(eco2.ID), Name: "move-dom-2"}
	if err := ds.CreateDomain(dom2); err != nil {
		t.Fatalf("CreateDomain(dom2) error = %v", err)
	}

	sys := &models.System{
		EcosystemID: validNullInt64(eco1.ID),
		DomainID:    validNullInt64(dom1.ID),
		Name:        "move-sys",
	}
	if err := ds.CreateSystem(sys); err != nil {
		t.Fatalf("CreateSystem error = %v", err)
	}

	app := &models.App{
		DomainID: validNullInt64(dom1.ID),
		SystemID: validNullInt64(sys.ID),
		Name:     "move-app",
		Path:     "/tmp/move-app",
	}
	if err := ds.CreateApp(app); err != nil {
		t.Fatalf("CreateApp error = %v", err)
	}

	return moveSetup{
		eco1ID: eco1.ID, eco2ID: eco2.ID,
		dom1ID: dom1.ID, dom2ID: dom2.ID,
		systemID: sys.ID, appID: app.ID,
	}
}

func TestSQLDataStore_MoveSystem_ReparentToNewDomain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()
	s := setupMoveTest(t, ds)

	// Move system from dom1/eco1 to dom2/eco2.
	if err := ds.MoveSystem(s.systemID, validNullInt64(s.dom2ID)); err != nil {
		t.Fatalf("MoveSystem() error = %v", err)
	}

	got, err := ds.GetSystemByID(s.systemID)
	if err != nil {
		t.Fatalf("GetSystemByID() error = %v", err)
	}
	if !got.DomainID.Valid || got.DomainID.Int64 != int64(s.dom2ID) {
		t.Errorf("system.DomainID = %v, want %d", got.DomainID, s.dom2ID)
	}
	if !got.EcosystemID.Valid || got.EcosystemID.Int64 != int64(s.eco2ID) {
		t.Errorf("system.EcosystemID = %v, want %d (recomputed from new domain)", got.EcosystemID, s.eco2ID)
	}

	// Cascade: child app's DomainID must follow the system to dom2.
	app, err := ds.GetAppByID(s.appID)
	if err != nil {
		t.Fatalf("GetAppByID() error = %v", err)
	}
	if !app.DomainID.Valid || app.DomainID.Int64 != int64(s.dom2ID) {
		t.Errorf("cascade: app.DomainID = %v, want %d", app.DomainID, s.dom2ID)
	}
	if !app.SystemID.Valid || app.SystemID.Int64 != int64(s.systemID) {
		t.Errorf("app.SystemID changed unexpectedly = %v, want %d", app.SystemID, s.systemID)
	}
}

func TestSQLDataStore_MoveSystem_DetachToNullDomain(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()
	s := setupMoveTest(t, ds)

	if err := ds.MoveSystem(s.systemID, sql.NullInt64{}); err != nil {
		t.Fatalf("MoveSystem(detach) error = %v", err)
	}

	got, _ := ds.GetSystemByID(s.systemID)
	if got.DomainID.Valid {
		t.Errorf("system.DomainID = %v, want NULL", got.DomainID)
	}
	if got.EcosystemID.Valid {
		t.Errorf("system.EcosystemID = %v, want NULL", got.EcosystemID)
	}

	// Cascade: child app's DomainID should also be NULL.
	app, _ := ds.GetAppByID(s.appID)
	if app.DomainID.Valid {
		t.Errorf("cascade: app.DomainID = %v, want NULL", app.DomainID)
	}
}

func TestSQLDataStore_MoveSystem_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.MoveSystem(99999, sql.NullInt64{})
	if err == nil {
		t.Fatalf("MoveSystem() expected error for missing system")
	}
	if _, ok := err.(*ErrNotFound); !ok {
		t.Errorf("MoveSystem() error type = %T, want *ErrNotFound: %v", err, err)
	}
}

func TestSQLDataStore_MoveSystem_TargetDomainNotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()
	s := setupMoveTest(t, ds)

	err := ds.MoveSystem(s.systemID, validNullInt64(99999))
	if err == nil {
		t.Fatalf("MoveSystem() expected error for missing target domain")
	}

	// Verify rollback: original system state intact.
	got, _ := ds.GetSystemByID(s.systemID)
	if !got.DomainID.Valid || got.DomainID.Int64 != int64(s.dom1ID) {
		t.Errorf("rollback failed: system.DomainID = %v, want %d", got.DomainID, s.dom1ID)
	}
}

// =============================================================================
// MoveApp tests — issue #397
// =============================================================================

func TestSQLDataStore_MoveApp_ReparentToNewDomainAndSystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()
	s := setupMoveTest(t, ds)

	// Create a second system under dom2 and move the app to it.
	sys2 := &models.System{
		EcosystemID: validNullInt64(s.eco2ID),
		DomainID:    validNullInt64(s.dom2ID),
		Name:        "move-sys-2",
	}
	if err := ds.CreateSystem(sys2); err != nil {
		t.Fatalf("CreateSystem(sys2) error = %v", err)
	}

	if err := ds.MoveApp(s.appID, validNullInt64(s.dom2ID), validNullInt64(sys2.ID)); err != nil {
		t.Fatalf("MoveApp() error = %v", err)
	}

	app, _ := ds.GetAppByID(s.appID)
	if !app.DomainID.Valid || app.DomainID.Int64 != int64(s.dom2ID) {
		t.Errorf("app.DomainID = %v, want %d", app.DomainID, s.dom2ID)
	}
	if !app.SystemID.Valid || app.SystemID.Int64 != int64(sys2.ID) {
		t.Errorf("app.SystemID = %v, want %d", app.SystemID, sys2.ID)
	}
}

func TestSQLDataStore_MoveApp_DomainOnlyNoSystem(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()
	s := setupMoveTest(t, ds)

	// Move app to dom2 with no system attached.
	if err := ds.MoveApp(s.appID, validNullInt64(s.dom2ID), sql.NullInt64{}); err != nil {
		t.Fatalf("MoveApp() error = %v", err)
	}

	app, _ := ds.GetAppByID(s.appID)
	if !app.DomainID.Valid || app.DomainID.Int64 != int64(s.dom2ID) {
		t.Errorf("app.DomainID = %v, want %d", app.DomainID, s.dom2ID)
	}
	if app.SystemID.Valid {
		t.Errorf("app.SystemID = %v, want NULL", app.SystemID)
	}
}

func TestSQLDataStore_MoveApp_FullDetach(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()
	s := setupMoveTest(t, ds)

	// Detach: BOTH domain and system NULL — `dvm app detach`.
	if err := ds.MoveApp(s.appID, sql.NullInt64{}, sql.NullInt64{}); err != nil {
		t.Fatalf("MoveApp(detach) error = %v", err)
	}

	app, _ := ds.GetAppByID(s.appID)
	if app.DomainID.Valid {
		t.Errorf("app.DomainID = %v, want NULL", app.DomainID)
	}
	if app.SystemID.Valid {
		t.Errorf("app.SystemID = %v, want NULL", app.SystemID)
	}
}

func TestSQLDataStore_MoveApp_NotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()

	err := ds.MoveApp(99999, sql.NullInt64{}, sql.NullInt64{})
	if err == nil {
		t.Fatalf("MoveApp() expected error for missing app")
	}
	if _, ok := err.(*ErrNotFound); !ok {
		t.Errorf("MoveApp() error type = %T, want *ErrNotFound: %v", err, err)
	}
}

func TestSQLDataStore_MoveApp_TargetDomainNotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()
	s := setupMoveTest(t, ds)

	err := ds.MoveApp(s.appID, validNullInt64(99999), sql.NullInt64{})
	if err == nil {
		t.Fatalf("MoveApp() expected error for missing target domain")
	}

	// Verify rollback.
	app, _ := ds.GetAppByID(s.appID)
	if !app.DomainID.Valid || app.DomainID.Int64 != int64(s.dom1ID) {
		t.Errorf("rollback failed: app.DomainID = %v, want %d", app.DomainID, s.dom1ID)
	}
}

func TestSQLDataStore_MoveApp_TargetSystemNotFound(t *testing.T) {
	ds := createTestDataStore(t)
	defer ds.Close()
	s := setupMoveTest(t, ds)

	err := ds.MoveApp(s.appID, validNullInt64(s.dom2ID), validNullInt64(99999))
	if err == nil {
		t.Fatalf("MoveApp() expected error for missing target system")
	}

	// Verify rollback: app stayed on dom1/sys1.
	app, _ := ds.GetAppByID(s.appID)
	if !app.DomainID.Valid || app.DomainID.Int64 != int64(s.dom1ID) {
		t.Errorf("rollback failed: app.DomainID = %v, want %d", app.DomainID, s.dom1ID)
	}
	if !app.SystemID.Valid || app.SystemID.Int64 != int64(s.systemID) {
		t.Errorf("rollback failed: app.SystemID = %v, want %d", app.SystemID, s.systemID)
	}
}
