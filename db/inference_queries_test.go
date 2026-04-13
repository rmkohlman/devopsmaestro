package db

import (
	"database/sql"
	"testing"

	"devopsmaestro/models"
)

// =============================================================================
// TestFindAppsByName
// =============================================================================

func TestFindAppsByName(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, ds *SQLDataStore)
		searchFor string
		wantCount int
		validate  func(t *testing.T, results []*models.AppWithHierarchy)
	}{
		{
			name:      "empty db returns empty slice",
			setup:     func(t *testing.T, ds *SQLDataStore) {},
			searchFor: "api",
			wantCount: 0,
			validate:  nil,
		},
		{
			name: "one app with that name returns 1 result with correct hierarchy",
			setup: func(t *testing.T, ds *SQLDataStore) {
				eco := &models.Ecosystem{Name: "findapps-eco-single"}
				if err := ds.CreateEcosystem(eco); err != nil {
					t.Fatalf("setup: CreateEcosystem: %v", err)
				}
				dom := &models.Domain{EcosystemID: validNullInt64(eco.ID), Name: "findapps-dom-single"}
				if err := ds.CreateDomain(dom); err != nil {
					t.Fatalf("setup: CreateDomain: %v", err)
				}
				app := &models.App{
					DomainID: validNullInt64(dom.ID),
					Name:        "api",
					Path:        "/services/api",
					Description: sql.NullString{String: "The API service", Valid: true},
				}
				if err := ds.CreateApp(app); err != nil {
					t.Fatalf("setup: CreateApp: %v", err)
				}
			},
			searchFor: "api",
			wantCount: 1,
			validate: func(t *testing.T, results []*models.AppWithHierarchy) {
				t.Helper()
				r := results[0]
				if r.App == nil {
					t.Fatal("App should not be nil")
				}
				if r.App.Name != "api" {
					t.Errorf("App.Name = %q, want %q", r.App.Name, "api")
				}
				if r.App.Path != "/services/api" {
					t.Errorf("App.Path = %q, want %q", r.App.Path, "/services/api")
				}
				if r.Domain == nil {
					t.Fatal("Domain should not be nil")
				}
				if r.Domain.Name != "findapps-dom-single" {
					t.Errorf("Domain.Name = %q, want %q", r.Domain.Name, "findapps-dom-single")
				}
				if r.Ecosystem == nil {
					t.Fatal("Ecosystem should not be nil")
				}
				if r.Ecosystem.Name != "findapps-eco-single" {
					t.Errorf("Ecosystem.Name = %q, want %q", r.Ecosystem.Name, "findapps-eco-single")
				}
			},
		},
		{
			name: "same app name in multiple domains returns all matches",
			setup: func(t *testing.T, ds *SQLDataStore) {
				// Ecosystem 1 → Domain 1 → App "api"
				eco1 := &models.Ecosystem{Name: "findapps-eco-multi1"}
				if err := ds.CreateEcosystem(eco1); err != nil {
					t.Fatalf("setup: CreateEcosystem eco1: %v", err)
				}
				dom1 := &models.Domain{EcosystemID: validNullInt64(eco1.ID), Name: "findapps-dom-multi1"}
				if err := ds.CreateDomain(dom1); err != nil {
					t.Fatalf("setup: CreateDomain dom1: %v", err)
				}
				app1 := &models.App{DomainID: validNullInt64(dom1.ID), Name: "api", Path: "/team1/api"}
				if err := ds.CreateApp(app1); err != nil {
					t.Fatalf("setup: CreateApp app1: %v", err)
				}

				// Ecosystem 2 → Domain 2 → App "api" (same name, different hierarchy)
				eco2 := &models.Ecosystem{Name: "findapps-eco-multi2"}
				if err := ds.CreateEcosystem(eco2); err != nil {
					t.Fatalf("setup: CreateEcosystem eco2: %v", err)
				}
				dom2 := &models.Domain{EcosystemID: validNullInt64(eco2.ID), Name: "findapps-dom-multi2"}
				if err := ds.CreateDomain(dom2); err != nil {
					t.Fatalf("setup: CreateDomain dom2: %v", err)
				}
				app2 := &models.App{DomainID: validNullInt64(dom2.ID), Name: "api", Path: "/team2/api"}
				if err := ds.CreateApp(app2); err != nil {
					t.Fatalf("setup: CreateApp app2: %v", err)
				}

				// Extra app with different name — should NOT appear in results
				app3 := &models.App{DomainID: validNullInt64(dom2.ID), Name: "other-service", Path: "/team2/other"}
				if err := ds.CreateApp(app3); err != nil {
					t.Fatalf("setup: CreateApp app3: %v", err)
				}
			},
			searchFor: "api",
			wantCount: 2,
			validate: func(t *testing.T, results []*models.AppWithHierarchy) {
				t.Helper()
				for _, r := range results {
					if r.App == nil {
						t.Fatal("App should not be nil")
					}
					if r.App.Name != "api" {
						t.Errorf("App.Name = %q, want %q", r.App.Name, "api")
					}
					if r.Domain == nil {
						t.Error("Domain should not be nil")
					}
					if r.Ecosystem == nil {
						t.Error("Ecosystem should not be nil")
					}
				}
				// Verify we get both distinct domains
				domains := make(map[string]bool)
				for _, r := range results {
					domains[r.Domain.Name] = true
				}
				if !domains["findapps-dom-multi1"] {
					t.Error("expected findapps-dom-multi1 in results")
				}
				if !domains["findapps-dom-multi2"] {
					t.Error("expected findapps-dom-multi2 in results")
				}
			},
		},
		{
			name: "name that doesn't exist returns empty slice",
			setup: func(t *testing.T, ds *SQLDataStore) {
				eco := &models.Ecosystem{Name: "findapps-eco-nomatch"}
				if err := ds.CreateEcosystem(eco); err != nil {
					t.Fatalf("setup: CreateEcosystem: %v", err)
				}
				dom := &models.Domain{EcosystemID: validNullInt64(eco.ID), Name: "findapps-dom-nomatch"}
				if err := ds.CreateDomain(dom); err != nil {
					t.Fatalf("setup: CreateDomain: %v", err)
				}
				app := &models.App{DomainID: validNullInt64(dom.ID), Name: "real-app", Path: "/real/path"}
				if err := ds.CreateApp(app); err != nil {
					t.Fatalf("setup: CreateApp: %v", err)
				}
			},
			searchFor: "nonexistent-app",
			wantCount: 0,
			validate:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			tt.setup(t, ds)

			// Act
			results, err := ds.FindAppsByName(tt.searchFor)
			if err != nil {
				t.Fatalf("FindAppsByName(%q) error = %v", tt.searchFor, err)
			}

			// Assert count
			if len(results) != tt.wantCount {
				t.Errorf("FindAppsByName(%q) returned %d results, want %d", tt.searchFor, len(results), tt.wantCount)
			}

			// Assert content
			if tt.validate != nil && len(results) > 0 {
				tt.validate(t, results)
			}
		})
	}
}

// =============================================================================
// TestFindDomainsByName
// =============================================================================

func TestFindDomainsByName(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, ds *SQLDataStore)
		searchFor string
		wantCount int
		validate  func(t *testing.T, results []*models.DomainWithHierarchy)
	}{
		{
			name:      "empty db returns empty slice",
			setup:     func(t *testing.T, ds *SQLDataStore) {},
			searchFor: "backend",
			wantCount: 0,
			validate:  nil,
		},
		{
			name: "one domain with that name returns 1 result with ecosystem",
			setup: func(t *testing.T, ds *SQLDataStore) {
				eco := &models.Ecosystem{
					Name:        "finddoms-eco-single",
					Description: sql.NullString{String: "Single test ecosystem", Valid: true},
				}
				if err := ds.CreateEcosystem(eco); err != nil {
					t.Fatalf("setup: CreateEcosystem: %v", err)
				}
				dom := &models.Domain{
					EcosystemID: validNullInt64(eco.ID),
					Name:        "backend",
					Description: sql.NullString{String: "Backend services", Valid: true},
				}
				if err := ds.CreateDomain(dom); err != nil {
					t.Fatalf("setup: CreateDomain: %v", err)
				}
			},
			searchFor: "backend",
			wantCount: 1,
			validate: func(t *testing.T, results []*models.DomainWithHierarchy) {
				t.Helper()
				r := results[0]
				if r.Domain == nil {
					t.Fatal("Domain should not be nil")
				}
				if r.Domain.Name != "backend" {
					t.Errorf("Domain.Name = %q, want %q", r.Domain.Name, "backend")
				}
				if r.Ecosystem == nil {
					t.Fatal("Ecosystem should not be nil")
				}
				if r.Ecosystem.Name != "finddoms-eco-single" {
					t.Errorf("Ecosystem.Name = %q, want %q", r.Ecosystem.Name, "finddoms-eco-single")
				}
			},
		},
		{
			name: "same domain name in multiple ecosystems returns all matches",
			setup: func(t *testing.T, ds *SQLDataStore) {
				// Ecosystem A → Domain "backend"
				ecoA := &models.Ecosystem{Name: "finddoms-eco-a"}
				if err := ds.CreateEcosystem(ecoA); err != nil {
					t.Fatalf("setup: CreateEcosystem ecoA: %v", err)
				}
				domA := &models.Domain{EcosystemID: validNullInt64(ecoA.ID), Name: "backend"}
				if err := ds.CreateDomain(domA); err != nil {
					t.Fatalf("setup: CreateDomain domA: %v", err)
				}

				// Ecosystem B → Domain "backend" (same name, different ecosystem)
				ecoB := &models.Ecosystem{Name: "finddoms-eco-b"}
				if err := ds.CreateEcosystem(ecoB); err != nil {
					t.Fatalf("setup: CreateEcosystem ecoB: %v", err)
				}
				domB := &models.Domain{EcosystemID: validNullInt64(ecoB.ID), Name: "backend"}
				if err := ds.CreateDomain(domB); err != nil {
					t.Fatalf("setup: CreateDomain domB: %v", err)
				}

				// Ecosystem B also has a different domain — should NOT appear
				domC := &models.Domain{EcosystemID: validNullInt64(ecoB.ID), Name: "frontend"}
				if err := ds.CreateDomain(domC); err != nil {
					t.Fatalf("setup: CreateDomain domC: %v", err)
				}
			},
			searchFor: "backend",
			wantCount: 2,
			validate: func(t *testing.T, results []*models.DomainWithHierarchy) {
				t.Helper()
				for _, r := range results {
					if r.Domain == nil {
						t.Fatal("Domain should not be nil")
					}
					if r.Domain.Name != "backend" {
						t.Errorf("Domain.Name = %q, want %q", r.Domain.Name, "backend")
					}
					if r.Ecosystem == nil {
						t.Error("Ecosystem should not be nil")
					}
				}
				// Verify we get both distinct ecosystems
				ecosystems := make(map[string]bool)
				for _, r := range results {
					ecosystems[r.Ecosystem.Name] = true
				}
				if !ecosystems["finddoms-eco-a"] {
					t.Error("expected finddoms-eco-a in results")
				}
				if !ecosystems["finddoms-eco-b"] {
					t.Error("expected finddoms-eco-b in results")
				}
			},
		},
		{
			name: "name that doesn't exist returns empty slice",
			setup: func(t *testing.T, ds *SQLDataStore) {
				eco := &models.Ecosystem{Name: "finddoms-eco-nomatch"}
				if err := ds.CreateEcosystem(eco); err != nil {
					t.Fatalf("setup: CreateEcosystem: %v", err)
				}
				dom := &models.Domain{EcosystemID: validNullInt64(eco.ID), Name: "real-domain"}
				if err := ds.CreateDomain(dom); err != nil {
					t.Fatalf("setup: CreateDomain: %v", err)
				}
			},
			searchFor: "nonexistent-domain",
			wantCount: 0,
			validate:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			tt.setup(t, ds)

			// Act
			results, err := ds.FindDomainsByName(tt.searchFor)
			if err != nil {
				t.Fatalf("FindDomainsByName(%q) error = %v", tt.searchFor, err)
			}

			// Assert count
			if len(results) != tt.wantCount {
				t.Errorf("FindDomainsByName(%q) returned %d results, want %d", tt.searchFor, len(results), tt.wantCount)
			}

			// Assert content
			if tt.validate != nil && len(results) > 0 {
				tt.validate(t, results)
			}
		})
	}
}

// =============================================================================
// TestCountEcosystems
// =============================================================================

func TestCountEcosystems(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(t *testing.T, ds *SQLDataStore)
		wantCount int
	}{
		{
			name:      "empty db returns 0",
			setup:     func(t *testing.T, ds *SQLDataStore) {},
			wantCount: 0,
		},
		{
			name: "one ecosystem returns 1",
			setup: func(t *testing.T, ds *SQLDataStore) {
				eco := &models.Ecosystem{Name: "counteco-single"}
				if err := ds.CreateEcosystem(eco); err != nil {
					t.Fatalf("setup: CreateEcosystem: %v", err)
				}
			},
			wantCount: 1,
		},
		{
			name: "multiple ecosystems returns correct count",
			setup: func(t *testing.T, ds *SQLDataStore) {
				for i := 1; i <= 5; i++ {
					eco := &models.Ecosystem{
						Name: "counteco-multi-" + string(rune('0'+i)),
					}
					if err := ds.CreateEcosystem(eco); err != nil {
						t.Fatalf("setup: CreateEcosystem #%d: %v", i, err)
					}
				}
			},
			wantCount: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ds := createTestDataStore(t)
			defer ds.Close()

			tt.setup(t, ds)

			// Act
			count, err := ds.CountEcosystems()
			if err != nil {
				t.Fatalf("CountEcosystems() error = %v", err)
			}

			// Assert
			if count != tt.wantCount {
				t.Errorf("CountEcosystems() = %d, want %d", count, tt.wantCount)
			}
		})
	}
}
