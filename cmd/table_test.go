package cmd

import (
	"database/sql"
	"strings"
	"testing"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/models"
	"github.com/rmkohlman/MaestroSDK/render"
)

// =============================================================================
// Helpers
// =============================================================================

func ptrInt(i int) *int { return &i }

func mustTime(s string) time.Time {
	t, err := time.Parse("2006-01-02 15:04", s)
	if err != nil {
		panic(err)
	}
	return t
}

func ptrString(s string) *string { return &s }

// =============================================================================
// ecosystemTableBuilder
// =============================================================================

func TestEcosystemTableBuilder_Headers_Default(t *testing.T) {
	b := &ecosystemTableBuilder{}
	headers := b.Headers(false)
	want := []string{"NAME", "DESCRIPTION", "THEME", "CREATED"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestEcosystemTableBuilder_Headers_Wide(t *testing.T) {
	b := &ecosystemTableBuilder{}
	headers := b.Headers(true)
	want := []string{"NAME", "DESCRIPTION", "THEME", "CREATED", "ID"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestEcosystemTableBuilder_Row_Default(t *testing.T) {
	b := &ecosystemTableBuilder{}
	eco := &models.Ecosystem{
		ID:          42,
		Name:        "prod",
		Description: sql.NullString{String: "Production", Valid: true},
		CreatedAt:   mustTime("2024-01-15 09:30"),
	}
	row := b.Row(eco, false)
	if len(row) != 4 {
		t.Fatalf("want 4 columns, got %d: %v", len(row), row)
	}
	if row[0] != "prod" {
		t.Errorf("row[0] (name) = %q, want %q", row[0], "prod")
	}
	if row[1] != "Production" {
		t.Errorf("row[1] (description) = %q, want %q", row[1], "Production")
	}
	if row[2] != "" {
		t.Errorf("row[2] (theme) = %q, want empty (no theme set)", row[2])
	}
	if row[3] != "2024-01-15 09:30" {
		t.Errorf("row[3] (created) = %q, want %q", row[3], "2024-01-15 09:30")
	}
}

func TestEcosystemTableBuilder_Row_Wide(t *testing.T) {
	b := &ecosystemTableBuilder{}
	eco := &models.Ecosystem{
		ID:          42,
		Name:        "prod",
		Description: sql.NullString{String: "Production", Valid: true},
		CreatedAt:   mustTime("2024-01-15 09:30"),
	}
	row := b.Row(eco, true)
	if len(row) != 5 {
		t.Fatalf("want 5 columns, got %d: %v", len(row), row)
	}
	if row[4] != "42" {
		t.Errorf("row[4] (id) = %q, want %q", row[4], "42")
	}
}

func TestEcosystemTableBuilder_Row_ActiveMarker(t *testing.T) {
	activeID := 42
	b := &ecosystemTableBuilder{ActiveID: &activeID}
	eco := &models.Ecosystem{ID: 42, Name: "prod", CreatedAt: mustTime("2024-01-15 09:30")}

	row := b.Row(eco, false)
	if !strings.HasPrefix(row[0], "●") {
		t.Errorf("active ecosystem: row[0] = %q, want prefix '●'", row[0])
	}
}

func TestEcosystemTableBuilder_Row_ActiveMarker_Absent(t *testing.T) {
	otherID := 99
	b := &ecosystemTableBuilder{ActiveID: &otherID}
	eco := &models.Ecosystem{ID: 42, Name: "prod", CreatedAt: mustTime("2024-01-15 09:30")}

	row := b.Row(eco, false)
	if strings.HasPrefix(row[0], "●") {
		t.Errorf("inactive ecosystem: row[0] = %q should NOT have '●' prefix", row[0])
	}
}

func TestEcosystemTableBuilder_Row_DescriptionTruncation(t *testing.T) {
	b := &ecosystemTableBuilder{}
	// 41 chars description – should be truncated to 37 + "..."
	longDesc := "This is a very long description text!!XXX"
	eco := &models.Ecosystem{
		ID:          1,
		Name:        "test",
		Description: sql.NullString{String: longDesc, Valid: true},
		CreatedAt:   mustTime("2024-01-15 09:30"),
	}
	row := b.Row(eco, false)
	if len(row[1]) != 40 {
		t.Errorf("truncated description length = %d, want 40 (37+3 ellipsis)", len(row[1]))
	}
	if row[1][37:] != "..." {
		t.Errorf("truncated description suffix = %q, want %q", row[1][37:], "...")
	}
}

func TestEcosystemTableBuilder_Row_DescriptionNotTruncated(t *testing.T) {
	b := &ecosystemTableBuilder{}
	// exactly 40 chars – no truncation
	exactDesc := "1234567890123456789012345678901234567890"
	eco := &models.Ecosystem{
		ID:          1,
		Name:        "test",
		Description: sql.NullString{String: exactDesc, Valid: true},
		CreatedAt:   mustTime("2024-01-15 09:30"),
	}
	row := b.Row(eco, false)
	if row[1] != exactDesc {
		t.Errorf("description = %q, want %q (should not be truncated)", row[1], exactDesc)
	}
}

func TestEcosystemTableBuilder_Row_NullDescription(t *testing.T) {
	b := &ecosystemTableBuilder{}
	eco := &models.Ecosystem{
		ID:          1,
		Name:        "test",
		Description: sql.NullString{Valid: false},
		CreatedAt:   mustTime("2024-01-15 09:30"),
	}
	row := b.Row(eco, false)
	if row[1] != "" {
		t.Errorf("null description = %q, want empty string", row[1])
	}
}

func TestEcosystemTableBuilder_Row_ThemeSet(t *testing.T) {
	b := &ecosystemTableBuilder{}
	eco := &models.Ecosystem{
		ID:        1,
		Name:      "test",
		Theme:     sql.NullString{String: "catppuccin-mocha", Valid: true},
		CreatedAt: mustTime("2024-01-15 09:30"),
	}
	row := b.Row(eco, false)
	if row[2] != "catppuccin-mocha" {
		t.Errorf("row[2] (theme) = %q, want %q", row[2], "catppuccin-mocha")
	}
}

func TestEcosystemTableBuilder_Row_ThemeNull(t *testing.T) {
	b := &ecosystemTableBuilder{}
	eco := &models.Ecosystem{
		ID:        1,
		Name:      "test",
		Theme:     sql.NullString{Valid: false},
		CreatedAt: mustTime("2024-01-15 09:30"),
	}
	row := b.Row(eco, false)
	if row[2] != "" {
		t.Errorf("row[2] (theme) = %q, want empty string for null theme", row[2])
	}
}

// =============================================================================
// domainTableBuilder
// =============================================================================

func TestDomainTableBuilder_Headers_Default(t *testing.T) {
	b := &domainTableBuilder{}
	headers := b.Headers(false)
	want := []string{"NAME", "ECOSYSTEM", "DESCRIPTION", "THEME", "CREATED"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestDomainTableBuilder_Headers_Wide(t *testing.T) {
	b := &domainTableBuilder{}
	headers := b.Headers(true)
	want := []string{"NAME", "ECOSYSTEM", "DESCRIPTION", "THEME", "CREATED", "ID"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestDomainTableBuilder_Row_Default(t *testing.T) {
	mockDS := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "prod"}
	_ = mockDS.CreateEcosystem(eco)
	// eco.ID is now 1

	b := &domainTableBuilder{DataStore: mockDS}
	domain := &models.Domain{
		ID:          10,
		EcosystemID: eco.ID,
		Name:        "backend",
		Description: sql.NullString{String: "Backend domain", Valid: true},
		CreatedAt:   mustTime("2024-02-20 10:00"),
	}
	row := b.Row(domain, false)
	if len(row) != 5 {
		t.Fatalf("want 5 columns, got %d: %v", len(row), row)
	}
	if row[0] != "backend" {
		t.Errorf("row[0] (name) = %q, want %q", row[0], "backend")
	}
	if row[1] != "prod" {
		t.Errorf("row[1] (ecosystem) = %q, want %q", row[1], "prod")
	}
	if row[2] != "Backend domain" {
		t.Errorf("row[2] (description) = %q, want %q", row[2], "Backend domain")
	}
}

func TestDomainTableBuilder_Row_Wide(t *testing.T) {
	mockDS := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "prod"}
	_ = mockDS.CreateEcosystem(eco)

	b := &domainTableBuilder{DataStore: mockDS}
	domain := &models.Domain{
		ID:          10,
		EcosystemID: eco.ID,
		Name:        "backend",
		CreatedAt:   mustTime("2024-02-20 10:00"),
	}
	row := b.Row(domain, true)
	if len(row) != 6 {
		t.Fatalf("want 6 columns, got %d: %v", len(row), row)
	}
	if row[5] != "10" {
		t.Errorf("row[5] (id) = %q, want %q", row[5], "10")
	}
}

func TestDomainTableBuilder_Row_ActiveMarker(t *testing.T) {
	mockDS := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "prod"}
	_ = mockDS.CreateEcosystem(eco)

	activeID := 10
	b := &domainTableBuilder{DataStore: mockDS, ActiveID: &activeID}
	domain := &models.Domain{
		ID:          10,
		EcosystemID: eco.ID,
		Name:        "backend",
		CreatedAt:   mustTime("2024-02-20 10:00"),
	}
	row := b.Row(domain, false)
	if !strings.HasPrefix(row[0], "●") {
		t.Errorf("active domain: row[0] = %q, want prefix '●'", row[0])
	}
}

func TestDomainTableBuilder_Row_ActiveMarker_Absent(t *testing.T) {
	mockDS := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "prod"}
	_ = mockDS.CreateEcosystem(eco)

	otherID := 99
	b := &domainTableBuilder{DataStore: mockDS, ActiveID: &otherID}
	domain := &models.Domain{
		ID:          10,
		EcosystemID: eco.ID,
		Name:        "backend",
		CreatedAt:   mustTime("2024-02-20 10:00"),
	}
	row := b.Row(domain, false)
	if strings.HasPrefix(row[0], "●") {
		t.Errorf("inactive domain: row[0] = %q should NOT have '●' prefix", row[0])
	}
}

func TestDomainTableBuilder_Row_DescriptionTruncation(t *testing.T) {
	mockDS := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "prod"}
	_ = mockDS.CreateEcosystem(eco)

	b := &domainTableBuilder{DataStore: mockDS}
	// 31 chars – truncated to 27 + "..."
	longDesc := "1234567890123456789012345678901"
	domain := &models.Domain{
		ID:          10,
		EcosystemID: eco.ID,
		Name:        "backend",
		Description: sql.NullString{String: longDesc, Valid: true},
		CreatedAt:   mustTime("2024-02-20 10:00"),
	}
	row := b.Row(domain, false)
	if len(row[2]) != 30 {
		t.Errorf("truncated description length = %d, want 30 (27+3 ellipsis)", len(row[2]))
	}
	if row[2][27:] != "..." {
		t.Errorf("truncated description suffix = %q, want %q", row[2][27:], "...")
	}
}

// =============================================================================
// appTableBuilder
// =============================================================================

func TestAppTableBuilder_Headers_Default(t *testing.T) {
	b := &appTableBuilder{}
	headers := b.Headers(false)
	want := []string{"NAME", "DOMAIN", "PATH", "THEME", "CREATED"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestAppTableBuilder_Headers_Wide(t *testing.T) {
	b := &appTableBuilder{}
	headers := b.Headers(true)
	want := []string{"NAME", "DOMAIN", "PATH", "THEME", "CREATED", "ID", "GITREPO"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestAppTableBuilder_Row_Default(t *testing.T) {
	mockDS := db.NewMockDataStore()
	domain := &models.Domain{EcosystemID: 1, Name: "backend"}
	_ = mockDS.CreateDomain(domain)

	b := &appTableBuilder{DataStore: mockDS}
	app := &models.App{
		ID:        5,
		DomainID:  domain.ID,
		Name:      "api",
		Path:      "/home/user/projects/api",
		CreatedAt: mustTime("2024-03-01 12:00"),
	}
	row := b.Row(app, false)
	if len(row) != 5 {
		t.Fatalf("want 5 columns, got %d: %v", len(row), row)
	}
	if row[0] != "api" {
		t.Errorf("row[0] (name) = %q, want %q", row[0], "api")
	}
	if row[1] != "backend" {
		t.Errorf("row[1] (domain) = %q, want %q", row[1], "backend")
	}
	if row[2] != "/home/user/projects/api" {
		t.Errorf("row[2] (path) = %q, want %q", row[2], "/home/user/projects/api")
	}
}

func TestAppTableBuilder_Row_Wide_ShowsGitRepoNone(t *testing.T) {
	mockDS := db.NewMockDataStore()
	domain := &models.Domain{EcosystemID: 1, Name: "backend"}
	_ = mockDS.CreateDomain(domain)

	b := &appTableBuilder{DataStore: mockDS}
	app := &models.App{
		ID:        5,
		DomainID:  domain.ID,
		Name:      "api",
		Path:      "/code/api",
		CreatedAt: mustTime("2024-03-01 12:00"),
	}
	row := b.Row(app, true)
	if len(row) != 7 {
		t.Fatalf("want 7 columns, got %d: %v", len(row), row)
	}
	if row[5] != "5" {
		t.Errorf("row[5] (id) = %q, want %q", row[5], "5")
	}
	if row[6] != "<none>" {
		t.Errorf("row[6] (gitrepo) = %q, want %q", row[6], "<none>")
	}
}

func TestAppTableBuilder_Row_ActiveMarker(t *testing.T) {
	mockDS := db.NewMockDataStore()
	domain := &models.Domain{EcosystemID: 1, Name: "backend"}
	_ = mockDS.CreateDomain(domain)

	activeID := 5
	b := &appTableBuilder{DataStore: mockDS, ActiveID: &activeID}
	app := &models.App{
		ID:        5,
		DomainID:  domain.ID,
		Name:      "api",
		Path:      "/code/api",
		CreatedAt: mustTime("2024-03-01 12:00"),
	}
	row := b.Row(app, false)
	if !strings.HasPrefix(row[0], "●") {
		t.Errorf("active app: row[0] = %q, want prefix '●'", row[0])
	}
}

func TestAppTableBuilder_Row_ActiveMarker_Absent(t *testing.T) {
	mockDS := db.NewMockDataStore()
	domain := &models.Domain{EcosystemID: 1, Name: "backend"}
	_ = mockDS.CreateDomain(domain)

	otherID := 99
	b := &appTableBuilder{DataStore: mockDS, ActiveID: &otherID}
	app := &models.App{
		ID:        5,
		DomainID:  domain.ID,
		Name:      "api",
		Path:      "/code/api",
		CreatedAt: mustTime("2024-03-01 12:00"),
	}
	row := b.Row(app, false)
	if strings.HasPrefix(row[0], "●") {
		t.Errorf("inactive app: row[0] = %q should NOT have '●' prefix", row[0])
	}
}

func TestAppTableBuilder_Row_PathTruncation(t *testing.T) {
	mockDS := db.NewMockDataStore()
	domain := &models.Domain{EcosystemID: 1, Name: "backend"}
	_ = mockDS.CreateDomain(domain)

	b := &appTableBuilder{DataStore: mockDS}
	// 41 chars path – truncated to "..." + last 37 chars
	longPath := "/home/user/projects/very-deep/nested/api"
	app := &models.App{
		ID:        5,
		DomainID:  domain.ID,
		Name:      "api",
		Path:      longPath,
		CreatedAt: mustTime("2024-03-01 12:00"),
	}
	row := b.Row(app, false)
	if len(row[2]) != 40 {
		t.Errorf("truncated path length = %d, want 40 (3+37)", len(row[2]))
	}
	if row[2][:3] != "..." {
		t.Errorf("truncated path prefix = %q, want %q", row[2][:3], "...")
	}
	// last 37 chars of longPath
	want := longPath[len(longPath)-37:]
	if row[2][3:] != want {
		t.Errorf("truncated path suffix = %q, want %q", row[2][3:], want)
	}
}

func TestAppTableBuilder_Row_PathNotTruncated(t *testing.T) {
	mockDS := db.NewMockDataStore()
	domain := &models.Domain{EcosystemID: 1, Name: "backend"}
	_ = mockDS.CreateDomain(domain)

	b := &appTableBuilder{DataStore: mockDS}
	exactPath := "/home/user/projects/api-service-long/" // 37 chars
	app := &models.App{
		ID:        5,
		DomainID:  domain.ID,
		Name:      "api",
		Path:      exactPath,
		CreatedAt: mustTime("2024-03-01 12:00"),
	}
	row := b.Row(app, false)
	if row[2] != exactPath {
		t.Errorf("path = %q, want %q (should not be truncated)", row[2], exactPath)
	}
}

// =============================================================================
// workspaceTableBuilder
// =============================================================================

func TestWorkspaceTableBuilder_Headers_Default(t *testing.T) {
	b := &workspaceTableBuilder{}
	headers := b.Headers(false)
	want := []string{"NAME", "APP", "IMAGE", "STATUS", "THEME"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestWorkspaceTableBuilder_Headers_Wide(t *testing.T) {
	b := &workspaceTableBuilder{}
	headers := b.Headers(true)
	want := []string{"NAME", "APP", "IMAGE", "STATUS", "THEME", "CREATED", "CONTAINER-ID"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestWorkspaceTableBuilder_Row_Default(t *testing.T) {
	mockDS := db.NewMockDataStore()
	app := &models.App{DomainID: 1, Name: "api", Path: "/code"}
	_ = mockDS.CreateApp(app)

	b := &workspaceTableBuilder{DataStore: mockDS}
	ws := &models.Workspace{
		ID:        7,
		AppID:     app.ID,
		Name:      "dev",
		ImageName: "ubuntu:22.04",
		Status:    "running",
		CreatedAt: mustTime("2024-04-10 08:00"),
	}
	row := b.Row(ws, false)
	if len(row) != 5 {
		t.Fatalf("want 5 columns, got %d: %v", len(row), row)
	}
	if row[0] != "dev" {
		t.Errorf("row[0] (name) = %q, want %q", row[0], "dev")
	}
	if row[1] != "api" {
		t.Errorf("row[1] (app) = %q, want %q", row[1], "api")
	}
	if row[2] != "ubuntu:22.04" {
		t.Errorf("row[2] (image) = %q, want %q", row[2], "ubuntu:22.04")
	}
	if row[3] != "running" {
		t.Errorf("row[3] (status) = %q, want %q", row[3], "running")
	}
	if row[4] != "" {
		t.Errorf("row[4] (theme) = %q, want empty (no theme set)", row[4])
	}
}

func TestWorkspaceTableBuilder_Row_Wide_WithContainerID(t *testing.T) {
	mockDS := db.NewMockDataStore()
	app := &models.App{DomainID: 1, Name: "api", Path: "/code"}
	_ = mockDS.CreateApp(app)

	b := &workspaceTableBuilder{DataStore: mockDS}
	ws := &models.Workspace{
		ID:          7,
		AppID:       app.ID,
		Name:        "dev",
		ImageName:   "ubuntu:22.04",
		Status:      "running",
		ContainerID: sql.NullString{String: "abc123def456789012345", Valid: true},
		CreatedAt:   mustTime("2024-04-10 08:00"),
	}
	row := b.Row(ws, true)
	if len(row) != 7 {
		t.Fatalf("want 7 columns, got %d: %v", len(row), row)
	}
	// Container-ID truncated to 12 chars
	if row[6] != "abc123def456" {
		t.Errorf("row[6] (container-id) = %q, want %q", row[6], "abc123def456")
	}
}

func TestWorkspaceTableBuilder_Row_Wide_NullContainerID(t *testing.T) {
	mockDS := db.NewMockDataStore()
	app := &models.App{DomainID: 1, Name: "api", Path: "/code"}
	_ = mockDS.CreateApp(app)

	b := &workspaceTableBuilder{DataStore: mockDS}
	ws := &models.Workspace{
		ID:          7,
		AppID:       app.ID,
		Name:        "dev",
		ImageName:   "ubuntu:22.04",
		Status:      "stopped",
		ContainerID: sql.NullString{Valid: false},
		CreatedAt:   mustTime("2024-04-10 08:00"),
	}
	row := b.Row(ws, true)
	if len(row) != 7 {
		t.Fatalf("want 7 columns, got %d: %v", len(row), row)
	}
	if row[6] != "<none>" {
		t.Errorf("row[6] (container-id) = %q, want %q", row[6], "<none>")
	}
}

func TestWorkspaceTableBuilder_Row_Wide_ShortContainerID(t *testing.T) {
	mockDS := db.NewMockDataStore()
	app := &models.App{DomainID: 1, Name: "api", Path: "/code"}
	_ = mockDS.CreateApp(app)

	b := &workspaceTableBuilder{DataStore: mockDS}
	ws := &models.Workspace{
		ID:          7,
		AppID:       app.ID,
		Name:        "dev",
		ImageName:   "ubuntu:22.04",
		Status:      "running",
		ContainerID: sql.NullString{String: "abc123", Valid: true},
		CreatedAt:   mustTime("2024-04-10 08:00"),
	}
	row := b.Row(ws, true)
	// Short container IDs should not be truncated
	if row[6] != "abc123" {
		t.Errorf("row[6] (container-id) = %q, want %q (no truncation)", row[6], "abc123")
	}
}

func TestWorkspaceTableBuilder_Row_ActiveMarker(t *testing.T) {
	mockDS := db.NewMockDataStore()
	app := &models.App{DomainID: 1, Name: "api", Path: "/code"}
	_ = mockDS.CreateApp(app)

	b := &workspaceTableBuilder{DataStore: mockDS, ActiveWorkspaceName: "dev"}
	ws := &models.Workspace{
		ID:        7,
		AppID:     app.ID,
		Name:      "dev",
		ImageName: "ubuntu:22.04",
		Status:    "running",
		CreatedAt: mustTime("2024-04-10 08:00"),
	}
	row := b.Row(ws, false)
	if !strings.HasPrefix(row[0], "●") {
		t.Errorf("active workspace: row[0] = %q, want prefix '●'", row[0])
	}
}

func TestWorkspaceTableBuilder_Row_ActiveMarker_Absent(t *testing.T) {
	mockDS := db.NewMockDataStore()
	app := &models.App{DomainID: 1, Name: "api", Path: "/code"}
	_ = mockDS.CreateApp(app)

	b := &workspaceTableBuilder{DataStore: mockDS, ActiveWorkspaceName: "other-ws"}
	ws := &models.Workspace{
		ID:        7,
		AppID:     app.ID,
		Name:      "dev",
		ImageName: "ubuntu:22.04",
		Status:    "running",
		CreatedAt: mustTime("2024-04-10 08:00"),
	}
	row := b.Row(ws, false)
	if strings.HasPrefix(row[0], "●") {
		t.Errorf("inactive workspace: row[0] = %q should NOT have '●' prefix", row[0])
	}
}

// =============================================================================
// credentialTableBuilder
// =============================================================================

func TestCredentialTableBuilder_Headers(t *testing.T) {
	b := &credentialTableBuilder{}
	// No wide mode for credentials – same either way
	for _, wide := range []bool{false, true} {
		headers := b.Headers(wide)
		want := []string{"NAME", "SCOPE", "SOURCE", "TARGET", "DESCRIPTION"}
		if len(headers) != len(want) {
			t.Fatalf("wide=%v: want %d headers, got %d: %v", wide, len(want), len(headers), headers)
		}
		for i, h := range want {
			if headers[i] != h {
				t.Errorf("wide=%v: headers[%d] = %q, want %q", wide, i, headers[i], h)
			}
		}
	}
}

func TestCredentialTableBuilder_Row_Default(t *testing.T) {
	mockDS := db.NewMockDataStore()
	app := &models.App{DomainID: 1, Name: "api", Path: "/code"}
	_ = mockDS.CreateApp(app)

	b := &credentialTableBuilder{DataStore: mockDS}
	desc := "GitHub token"
	cred := &models.CredentialDB{
		Name:        "github-token",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "vault",
		EnvVar:      ptrString("GITHUB_TOKEN"),
		Description: &desc,
	}
	row := b.Row(cred, false)
	if len(row) != 5 {
		t.Fatalf("want 5 columns, got %d: %v", len(row), row)
	}
	if row[0] != "github-token" {
		t.Errorf("row[0] (name) = %q, want %q", row[0], "github-token")
	}
	// row[1] is scope resolved via resolveScopeName
	if row[2] != "vault" {
		t.Errorf("row[2] (source) = %q, want %q", row[2], "vault")
	}
	// row[3] is target vars – EnvVar is GITHUB_TOKEN
	if row[3] != "GITHUB_TOKEN" {
		t.Errorf("row[3] (target) = %q, want %q", row[3], "GITHUB_TOKEN")
	}
	if row[4] != "GitHub token" {
		t.Errorf("row[4] (description) = %q, want %q", row[4], "GitHub token")
	}
}

func TestCredentialTableBuilder_Row_NilDescription(t *testing.T) {
	mockDS := db.NewMockDataStore()
	app := &models.App{DomainID: 1, Name: "api", Path: "/code"}
	_ = mockDS.CreateApp(app)

	b := &credentialTableBuilder{DataStore: mockDS}
	cred := &models.CredentialDB{
		Name:        "github-token",
		ScopeType:   models.CredentialScopeApp,
		ScopeID:     int64(app.ID),
		Source:      "env",
		EnvVar:      ptrString("GITHUB_TOKEN"),
		Description: nil,
	}
	row := b.Row(cred, false)
	if row[4] != "" {
		t.Errorf("row[4] (description) = %q, want empty string for nil description", row[4])
	}
}

// =============================================================================
// registryTableBuilder
// =============================================================================

func TestRegistryTableBuilder_Headers_Default(t *testing.T) {
	b := &registryTableBuilder{}
	headers := b.Headers(false)
	want := []string{"NAME", "TYPE", "VERSION", "PORT", "LIFECYCLE", "STATE", "UPTIME"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestRegistryTableBuilder_Headers_Wide(t *testing.T) {
	b := &registryTableBuilder{}
	headers := b.Headers(true)
	want := []string{"NAME", "TYPE", "VERSION", "PORT", "LIFECYCLE", "STATE", "UPTIME", "CREATED"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestRegistryTableBuilder_Row_Default(t *testing.T) {
	statusMap := map[string]string{
		"zot-registry": "running/2h30m",
	}
	b := &registryTableBuilder{StatusMap: statusMap}
	reg := &models.Registry{
		Name:      "zot-registry",
		Type:      "zot",
		Version:   "2.1.15",
		Port:      5001,
		Lifecycle: "persistent",
		Status:    "running",
		CreatedAt: "2024-05-01",
	}
	row := b.Row(reg, false)
	if len(row) != 7 {
		t.Fatalf("want 7 columns, got %d: %v", len(row), row)
	}
	if row[0] != "zot-registry" {
		t.Errorf("row[0] (name) = %q, want %q", row[0], "zot-registry")
	}
	if row[1] != "zot" {
		t.Errorf("row[1] (type) = %q, want %q", row[1], "zot")
	}
	if row[2] != "2.1.15" {
		t.Errorf("row[2] (version) = %q, want %q", row[2], "2.1.15")
	}
	if row[3] != "5001" {
		t.Errorf("row[3] (port) = %q, want %q", row[3], "5001")
	}
	if row[4] != "persistent" {
		t.Errorf("row[4] (lifecycle) = %q, want %q", row[4], "persistent")
	}
}

func TestRegistryTableBuilder_Row_Wide(t *testing.T) {
	statusMap := map[string]string{}
	b := &registryTableBuilder{StatusMap: statusMap}
	reg := &models.Registry{
		Name:      "athens",
		Type:      "athens",
		Version:   "0.13.0",
		Port:      3000,
		Lifecycle: "on-demand",
		Status:    "stopped",
		CreatedAt: "2024-05-01",
	}
	row := b.Row(reg, true)
	if len(row) != 8 {
		t.Fatalf("want 8 columns, got %d: %v", len(row), row)
	}
	if row[7] != "2024-05-01" {
		t.Errorf("row[7] (created) = %q, want %q", row[7], "2024-05-01")
	}
}

func TestRegistryTableBuilder_Row_StatusFromStatusMap(t *testing.T) {
	statusMap := map[string]string{
		"zot-registry": "running/1h",
	}
	b := &registryTableBuilder{StatusMap: statusMap}
	reg := &models.Registry{
		Name:      "zot-registry",
		Type:      "zot",
		Version:   "2.1.15",
		Port:      5001,
		Lifecycle: "persistent",
		Status:    "stopped", // DB status overridden by StatusMap
		CreatedAt: "2024-05-01",
	}
	row := b.Row(reg, false)
	// STATE and UPTIME come from StatusMap
	if row[5] != "running" {
		t.Errorf("row[5] (state) = %q, want %q (from StatusMap)", row[5], "running")
	}
}

// =============================================================================
// gitRepoTableBuilder
// =============================================================================

func TestGitRepoTableBuilder_Headers_Default(t *testing.T) {
	b := &gitRepoTableBuilder{}
	headers := b.Headers(false)
	want := []string{"NAME", "URL", "STATUS", "LAST_SYNCED"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestGitRepoTableBuilder_Headers_Wide(t *testing.T) {
	b := &gitRepoTableBuilder{}
	headers := b.Headers(true)
	want := []string{"NAME", "URL", "STATUS", "LAST_SYNCED", "SLUG", "REF", "AUTO_SYNC"}
	if len(headers) != len(want) {
		t.Fatalf("want %d headers, got %d: %v", len(want), len(headers), headers)
	}
	for i, h := range want {
		if headers[i] != h {
			t.Errorf("headers[%d] = %q, want %q", i, headers[i], h)
		}
	}
}

func TestGitRepoTableBuilder_Row_Default(t *testing.T) {
	b := &gitRepoTableBuilder{}
	syncedAt := time.Now()
	repo := &models.GitRepoDB{
		Name:         "my-repo",
		URL:          "https://github.com/org/repo.git",
		SyncStatus:   "synced",
		LastSyncedAt: sql.NullTime{Time: syncedAt, Valid: true},
		Slug:         "my-repo",
		DefaultRef:   "main",
		AutoSync:     true,
	}
	row := b.Row(repo, false)
	if len(row) != 4 {
		t.Fatalf("want 4 columns, got %d: %v", len(row), row)
	}
	if row[0] != "my-repo" {
		t.Errorf("row[0] (name) = %q, want %q", row[0], "my-repo")
	}
	if row[1] != "https://github.com/org/repo.git" {
		t.Errorf("row[1] (url) = %q, want %q", row[1], "https://github.com/org/repo.git")
	}
	if row[2] != "synced" {
		t.Errorf("row[2] (status) = %q, want %q", row[2], "synced")
	}
	// row[3] is formatted date – not empty
	if row[3] == "never" {
		t.Errorf("row[3] (last_synced) = %q, should not be 'never' when valid", row[3])
	}
}

func TestGitRepoTableBuilder_Row_NeverSynced(t *testing.T) {
	b := &gitRepoTableBuilder{}
	repo := &models.GitRepoDB{
		Name:         "my-repo",
		URL:          "https://github.com/org/repo.git",
		SyncStatus:   "pending",
		LastSyncedAt: sql.NullTime{Valid: false},
		Slug:         "my-repo",
		DefaultRef:   "main",
		AutoSync:     false,
	}
	row := b.Row(repo, false)
	if row[3] != "never" {
		t.Errorf("row[3] (last_synced) = %q, want %q when null", row[3], "never")
	}
}

func TestGitRepoTableBuilder_Row_Wide(t *testing.T) {
	b := &gitRepoTableBuilder{}
	repo := &models.GitRepoDB{
		Name:         "my-repo",
		URL:          "https://github.com/org/repo.git",
		SyncStatus:   "synced",
		LastSyncedAt: sql.NullTime{Valid: false},
		Slug:         "my-repo-slug",
		DefaultRef:   "main",
		AutoSync:     true,
	}
	row := b.Row(repo, true)
	if len(row) != 7 {
		t.Fatalf("want 7 columns, got %d: %v", len(row), row)
	}
	if row[4] != "my-repo-slug" {
		t.Errorf("row[4] (slug) = %q, want %q", row[4], "my-repo-slug")
	}
	if row[5] != "main" {
		t.Errorf("row[5] (ref) = %q, want %q", row[5], "main")
	}
	if row[6] != "yes" {
		t.Errorf("row[6] (auto_sync) = %q, want %q", row[6], "yes")
	}
}

func TestGitRepoTableBuilder_Row_Wide_AutoSyncNo(t *testing.T) {
	b := &gitRepoTableBuilder{}
	repo := &models.GitRepoDB{
		Name:         "my-repo",
		URL:          "https://github.com/org/repo.git",
		SyncStatus:   "pending",
		LastSyncedAt: sql.NullTime{Valid: false},
		Slug:         "my-repo",
		DefaultRef:   "develop",
		AutoSync:     false,
	}
	row := b.Row(repo, true)
	if row[6] != "no" {
		t.Errorf("row[6] (auto_sync) = %q, want %q", row[6], "no")
	}
}

// =============================================================================
// nvimPluginTableBuilder
// =============================================================================

func TestNvimPluginTableBuilder_Headers(t *testing.T) {
	b := &nvimPluginTableBuilder{}
	// No wide mode for nvim plugins
	for _, wide := range []bool{false, true} {
		headers := b.Headers(wide)
		want := []string{"NAME", "CATEGORY", "REPO", "VERSION"}
		if len(headers) != len(want) {
			t.Fatalf("wide=%v: want %d headers, got %d: %v", wide, len(want), len(headers), headers)
		}
		for i, h := range want {
			if headers[i] != h {
				t.Errorf("wide=%v: headers[%d] = %q, want %q", wide, i, headers[i], h)
			}
		}
	}
}

func TestNvimPluginTableBuilder_Row_Enabled(t *testing.T) {
	b := &nvimPluginTableBuilder{}
	plugin := &models.NvimPluginDB{
		Name:     "telescope",
		Category: sql.NullString{String: "fuzzy-finder", Valid: true},
		Repo:     "nvim-telescope/telescope.nvim",
		Version:  sql.NullString{String: "0.1.5", Valid: true},
		Branch:   sql.NullString{Valid: false},
		Enabled:  true,
	}
	row := b.Row(plugin, false)
	if len(row) != 4 {
		t.Fatalf("want 4 columns, got %d: %v", len(row), row)
	}
	// enabled: name ends with " ✓"
	if len(row[0]) < 2 {
		t.Fatalf("row[0] too short: %q", row[0])
	}
	// Check the mark is present (either ✓ or ✗)
	if row[0] != "telescope ✓" {
		t.Errorf("row[0] (name+enabled) = %q, want %q", row[0], "telescope ✓")
	}
	if row[1] != "fuzzy-finder" {
		t.Errorf("row[1] (category) = %q, want %q", row[1], "fuzzy-finder")
	}
	if row[2] != "nvim-telescope/telescope.nvim" {
		t.Errorf("row[2] (repo) = %q, want %q", row[2], "nvim-telescope/telescope.nvim")
	}
	if row[3] != "0.1.5" {
		t.Errorf("row[3] (version) = %q, want %q", row[3], "0.1.5")
	}
}

func TestNvimPluginTableBuilder_Row_Disabled(t *testing.T) {
	b := &nvimPluginTableBuilder{}
	plugin := &models.NvimPluginDB{
		Name:     "cmp",
		Category: sql.NullString{Valid: false},
		Repo:     "hrsh7th/nvim-cmp",
		Version:  sql.NullString{Valid: false},
		Branch:   sql.NullString{Valid: false},
		Enabled:  false,
	}
	row := b.Row(plugin, false)
	if row[0] != "cmp ✗" {
		t.Errorf("row[0] (name+disabled) = %q, want %q", row[0], "cmp ✗")
	}
}

func TestNvimPluginTableBuilder_Row_VersionFromBranch(t *testing.T) {
	b := &nvimPluginTableBuilder{}
	plugin := &models.NvimPluginDB{
		Name:    "lazy",
		Repo:    "folke/lazy.nvim",
		Version: sql.NullString{Valid: false},
		Branch:  sql.NullString{String: "stable", Valid: true},
		Enabled: true,
	}
	row := b.Row(plugin, false)
	if row[3] != "branch:stable" {
		t.Errorf("row[3] (version from branch) = %q, want %q", row[3], "branch:stable")
	}
}

func TestNvimPluginTableBuilder_Row_VersionLatestDefault(t *testing.T) {
	b := &nvimPluginTableBuilder{}
	plugin := &models.NvimPluginDB{
		Name:    "lualine",
		Repo:    "nvim-lualine/lualine.nvim",
		Version: sql.NullString{Valid: false},
		Branch:  sql.NullString{Valid: false},
		Enabled: true,
	}
	row := b.Row(plugin, false)
	if row[3] != "latest" {
		t.Errorf("row[3] (version default) = %q, want %q", row[3], "latest")
	}
}

func TestNvimPluginTableBuilder_Row_VersionPriority(t *testing.T) {
	// Explicit version field takes priority over branch
	b := &nvimPluginTableBuilder{}
	plugin := &models.NvimPluginDB{
		Name:    "alpha",
		Repo:    "goolord/alpha-nvim",
		Version: sql.NullString{String: "1.2.3", Valid: true},
		Branch:  sql.NullString{String: "stable", Valid: true},
		Enabled: true,
	}
	row := b.Row(plugin, false)
	if row[3] != "1.2.3" {
		t.Errorf("row[3] (version priority) = %q, want %q (version field takes priority)", row[3], "1.2.3")
	}
}

// =============================================================================
// nvimThemeTableBuilder
// =============================================================================

func TestNvimThemeTableBuilder_Headers(t *testing.T) {
	b := &nvimThemeTableBuilder{}
	// No wide mode for nvim themes
	for _, wide := range []bool{false, true} {
		headers := b.Headers(wide)
		want := []string{"NAME", "CATEGORY", "PLUGIN", "STYLE"}
		if len(headers) != len(want) {
			t.Fatalf("wide=%v: want %d headers, got %d: %v", wide, len(want), len(headers), headers)
		}
		for i, h := range want {
			if headers[i] != h {
				t.Errorf("wide=%v: headers[%d] = %q, want %q", wide, i, headers[i], h)
			}
		}
	}
}

func TestNvimThemeTableBuilder_Row_Default(t *testing.T) {
	b := &nvimThemeTableBuilder{}
	theme := &models.NvimThemeDB{
		Name:       "tokyonight",
		Category:   sql.NullString{String: "dark", Valid: true},
		PluginRepo: "folke/tokyonight.nvim",
		Style:      sql.NullString{String: "night", Valid: true},
	}
	row := b.Row(theme, false)
	if len(row) != 4 {
		t.Fatalf("want 4 columns, got %d: %v", len(row), row)
	}
	if row[0] != "tokyonight" {
		t.Errorf("row[0] (name) = %q, want %q", row[0], "tokyonight")
	}
	if row[1] != "dark" {
		t.Errorf("row[1] (category) = %q, want %q", row[1], "dark")
	}
	if row[2] != "folke/tokyonight.nvim" {
		t.Errorf("row[2] (plugin) = %q, want %q", row[2], "folke/tokyonight.nvim")
	}
	if row[3] != "night" {
		t.Errorf("row[3] (style) = %q, want %q", row[3], "night")
	}
}

func TestNvimThemeTableBuilder_Row_NullCategory(t *testing.T) {
	b := &nvimThemeTableBuilder{}
	theme := &models.NvimThemeDB{
		Name:       "gruvbox",
		Category:   sql.NullString{Valid: false},
		PluginRepo: "ellisonleao/gruvbox.nvim",
		Style:      sql.NullString{String: "dark", Valid: true},
	}
	row := b.Row(theme, false)
	if row[1] != "-" {
		t.Errorf("row[1] (category) = %q, want %q for empty category", row[1], "-")
	}
}

func TestNvimThemeTableBuilder_Row_NullStyle(t *testing.T) {
	b := &nvimThemeTableBuilder{}
	theme := &models.NvimThemeDB{
		Name:       "gruvbox",
		Category:   sql.NullString{String: "dark", Valid: true},
		PluginRepo: "ellisonleao/gruvbox.nvim",
		Style:      sql.NullString{Valid: false},
	}
	row := b.Row(theme, false)
	if row[3] != "default" {
		t.Errorf("row[3] (style) = %q, want %q for empty style", row[3], "default")
	}
}

// =============================================================================
// BuildTable generic helper
// =============================================================================

func TestBuildTable_AssemblesHeadersAndRows(t *testing.T) {
	b := &gitRepoTableBuilder{}
	repos := []*models.GitRepoDB{
		{
			Name:         "repo-a",
			URL:          "https://github.com/org/a.git",
			SyncStatus:   "synced",
			LastSyncedAt: sql.NullTime{Valid: false},
			Slug:         "repo-a",
			DefaultRef:   "main",
			AutoSync:     false,
		},
		{
			Name:         "repo-b",
			URL:          "https://github.com/org/b.git",
			SyncStatus:   "pending",
			LastSyncedAt: sql.NullTime{Valid: false},
			Slug:         "repo-b",
			DefaultRef:   "develop",
			AutoSync:     true,
		},
	}
	tableData := BuildTable(b, repos, false)

	// Verify it's a render.TableData with correct structure
	var _ render.TableData = tableData // compile-time check

	if len(tableData.Headers) != 4 {
		t.Fatalf("want 4 headers, got %d", len(tableData.Headers))
	}
	if len(tableData.Rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(tableData.Rows))
	}
	if tableData.Rows[0][0] != "repo-a" {
		t.Errorf("row[0][0] = %q, want %q", tableData.Rows[0][0], "repo-a")
	}
	if tableData.Rows[1][0] != "repo-b" {
		t.Errorf("row[1][0] = %q, want %q", tableData.Rows[1][0], "repo-b")
	}
}

func TestBuildTable_EmptyItems(t *testing.T) {
	b := &gitRepoTableBuilder{}
	repos := []*models.GitRepoDB{}
	tableData := BuildTable(b, repos, false)

	if len(tableData.Headers) != 4 {
		t.Fatalf("want 4 headers, got %d", len(tableData.Headers))
	}
	if len(tableData.Rows) != 0 {
		t.Fatalf("want 0 rows for empty slice, got %d", len(tableData.Rows))
	}
}

func TestBuildTable_WideMode(t *testing.T) {
	b := &gitRepoTableBuilder{}
	repos := []*models.GitRepoDB{
		{
			Name:         "repo-a",
			URL:          "https://github.com/org/a.git",
			SyncStatus:   "synced",
			LastSyncedAt: sql.NullTime{Valid: false},
			Slug:         "repo-a-slug",
			DefaultRef:   "main",
			AutoSync:     true,
		},
	}
	tableData := BuildTable(b, repos, true)

	if len(tableData.Headers) != 7 {
		t.Fatalf("wide mode: want 7 headers, got %d", len(tableData.Headers))
	}
	if len(tableData.Rows[0]) != 7 {
		t.Fatalf("wide mode: want 7 columns in row, got %d", len(tableData.Rows[0]))
	}
}

func TestBuildTable_EcosystemBuilder(t *testing.T) {
	activeID := 1
	b := &ecosystemTableBuilder{ActiveID: &activeID}
	ecos := []*models.Ecosystem{
		{ID: 1, Name: "prod", CreatedAt: mustTime("2024-01-01 00:00")},
		{ID: 2, Name: "staging", CreatedAt: mustTime("2024-01-02 00:00")},
	}
	tableData := BuildTable(b, ecos, false)

	if len(tableData.Rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(tableData.Rows))
	}
	// First ecosystem (ID=1) matches activeID=1, should have ● prefix
	if !strings.HasPrefix(tableData.Rows[0][0], "●") {
		t.Errorf("row[0][0] = %q, want '●' prefix (active ecosystem)", tableData.Rows[0][0])
	}
	// Second ecosystem (ID=2) is not active
	if strings.HasPrefix(tableData.Rows[1][0], "●") {
		t.Errorf("row[1][0] = %q should NOT have '●' prefix (inactive)", tableData.Rows[1][0])
	}
}

// =============================================================================
// THEME column tests — domain, app, workspace
// These verify THEME column (added as part of hierarchy theme resolution work)
// =============================================================================

func TestDomainTableBuilder_Row_ThemeSet(t *testing.T) {
	mockDS := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "prod"}
	_ = mockDS.CreateEcosystem(eco)

	b := &domainTableBuilder{DataStore: mockDS}
	domain := &models.Domain{
		ID:          10,
		EcosystemID: eco.ID,
		Name:        "backend",
		Theme:       sql.NullString{String: "tokyonight-night", Valid: true},
		CreatedAt:   mustTime("2024-02-20 10:00"),
	}
	row := b.Row(domain, false)
	// THEME is column index 3 (NAME, ECOSYSTEM, DESCRIPTION, THEME, CREATED)
	if row[3] != "tokyonight-night" {
		t.Errorf("row[3] (theme) = %q, want %q", row[3], "tokyonight-night")
	}
}

func TestDomainTableBuilder_Row_ThemeNull(t *testing.T) {
	mockDS := db.NewMockDataStore()
	eco := &models.Ecosystem{Name: "prod"}
	_ = mockDS.CreateEcosystem(eco)

	b := &domainTableBuilder{DataStore: mockDS}
	domain := &models.Domain{
		ID:          10,
		EcosystemID: eco.ID,
		Name:        "backend",
		Theme:       sql.NullString{Valid: false},
		CreatedAt:   mustTime("2024-02-20 10:00"),
	}
	row := b.Row(domain, false)
	if row[3] != "" {
		t.Errorf("row[3] (theme) = %q, want empty string for null theme", row[3])
	}
}

func TestAppTableBuilder_Row_ThemeSet(t *testing.T) {
	mockDS := db.NewMockDataStore()
	domain := &models.Domain{EcosystemID: 1, Name: "backend"}
	_ = mockDS.CreateDomain(domain)

	b := &appTableBuilder{DataStore: mockDS}
	app := &models.App{
		ID:        5,
		DomainID:  domain.ID,
		Name:      "api",
		Path:      "/code/api",
		Theme:     sql.NullString{String: "catppuccin-mocha", Valid: true},
		CreatedAt: mustTime("2024-03-01 12:00"),
	}
	row := b.Row(app, false)
	// THEME is column index 3 (NAME, DOMAIN, PATH, THEME, CREATED)
	if row[3] != "catppuccin-mocha" {
		t.Errorf("row[3] (theme) = %q, want %q", row[3], "catppuccin-mocha")
	}
}

func TestAppTableBuilder_Row_ThemeNull(t *testing.T) {
	mockDS := db.NewMockDataStore()
	domain := &models.Domain{EcosystemID: 1, Name: "backend"}
	_ = mockDS.CreateDomain(domain)

	b := &appTableBuilder{DataStore: mockDS}
	app := &models.App{
		ID:        5,
		DomainID:  domain.ID,
		Name:      "api",
		Path:      "/code/api",
		Theme:     sql.NullString{Valid: false},
		CreatedAt: mustTime("2024-03-01 12:00"),
	}
	row := b.Row(app, false)
	if row[3] != "" {
		t.Errorf("row[3] (theme) = %q, want empty string for null theme", row[3])
	}
}

func TestWorkspaceTableBuilder_Row_ThemeSet(t *testing.T) {
	mockDS := db.NewMockDataStore()
	app := &models.App{DomainID: 1, Name: "api", Path: "/code"}
	_ = mockDS.CreateApp(app)

	b := &workspaceTableBuilder{DataStore: mockDS}
	ws := &models.Workspace{
		ID:        7,
		AppID:     app.ID,
		Name:      "dev",
		ImageName: "ubuntu:22.04",
		Status:    "running",
		Theme:     sql.NullString{String: "gruvbox-dark", Valid: true},
		CreatedAt: mustTime("2024-04-10 08:00"),
	}
	row := b.Row(ws, false)
	// THEME is column index 4 (NAME, APP, IMAGE, STATUS, THEME)
	if row[4] != "gruvbox-dark" {
		t.Errorf("row[4] (theme) = %q, want %q", row[4], "gruvbox-dark")
	}
}

func TestWorkspaceTableBuilder_Row_ThemeNull(t *testing.T) {
	mockDS := db.NewMockDataStore()
	app := &models.App{DomainID: 1, Name: "api", Path: "/code"}
	_ = mockDS.CreateApp(app)

	b := &workspaceTableBuilder{DataStore: mockDS}
	ws := &models.Workspace{
		ID:        7,
		AppID:     app.ID,
		Name:      "dev",
		ImageName: "ubuntu:22.04",
		Status:    "running",
		Theme:     sql.NullString{Valid: false},
		CreatedAt: mustTime("2024-04-10 08:00"),
	}
	row := b.Row(ws, false)
	if row[4] != "" {
		t.Errorf("row[4] (theme) = %q, want empty string for null theme", row[4])
	}
}
