package db

import (
	"database/sql"
	"io/fs"
	"path/filepath"
	"testing"

	"devopsmaestro/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTerminalEmulator_IntegrationTest(t *testing.T) {
	// Create a temporary database file for testing to avoid race conditions
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	cfg := DriverConfig{Type: DriverSQLite, Path: dbPath}
	driver, err := NewSQLiteDriver(cfg)
	require.NoError(t, err)

	err = driver.Connect()
	require.NoError(t, err)
	defer driver.Close()

	store := &SQLDataStore{
		driver:       driver,
		queryBuilder: NewSQLiteQueryBuilder(),
	}

	// Run migrations
	migrationsSubFS, err := fs.Sub(testMigrationsFS, "migrations")
	require.NoError(t, err)
	err = AutoMigrate(driver, migrationsSubFS)
	require.NoError(t, err)

	// Test data
	emulator1 := &models.TerminalEmulatorDB{
		Name:        "wezterm-dark",
		Description: sql.NullString{String: "WezTerm with dark theme", Valid: true},
		Type:        "wezterm",
		Config:      `{"theme": "dark", "font_size": 14}`,
		ThemeRef:    sql.NullString{String: "dark-pro", Valid: true},
		Category:    sql.NullString{String: "terminal", Valid: true},
		Labels:      `{"type": "terminal", "theme": "dark"}`,
		Workspace:   sql.NullString{String: "main-workspace", Valid: true},
		Enabled:     true,
	}

	emulator2 := &models.TerminalEmulatorDB{
		Name:        "alacritty-light",
		Description: sql.NullString{String: "Alacritty with light theme", Valid: true},
		Type:        "alacritty",
		Config:      `{"theme": "light", "opacity": 0.9}`,
		ThemeRef:    sql.NullString{String: "light-theme", Valid: true},
		Category:    sql.NullString{String: "terminal", Valid: true},
		Labels:      `{"type": "terminal", "theme": "light"}`,
		Workspace:   sql.NullString{String: "dev-workspace", Valid: true},
		Enabled:     true,
	}

	// Test Create
	t.Run("Create", func(t *testing.T) {
		err := store.CreateTerminalEmulator(emulator1)
		assert.NoError(t, err)
		assert.NotZero(t, emulator1.ID)

		err = store.CreateTerminalEmulator(emulator2)
		assert.NoError(t, err)
		assert.NotZero(t, emulator2.ID)
	})

	// Test Get
	t.Run("Get", func(t *testing.T) {
		retrieved, err := store.GetTerminalEmulator("wezterm-dark")
		assert.NoError(t, err)
		assert.Equal(t, emulator1.Name, retrieved.Name)
		assert.Equal(t, emulator1.Type, retrieved.Type)
		assert.Equal(t, emulator1.Config, retrieved.Config)
		assert.Equal(t, emulator1.ThemeRef, retrieved.ThemeRef)
		assert.Equal(t, emulator1.Labels, retrieved.Labels)
	})

	// Test Update
	t.Run("Update", func(t *testing.T) {
		emulator1.Config = `{"theme": "dark", "font_size": 16}`
		emulator1.Description = sql.NullString{String: "Updated WezTerm config", Valid: true}

		err := store.UpdateTerminalEmulator(emulator1)
		assert.NoError(t, err)

		updated, err := store.GetTerminalEmulator("wezterm-dark")
		assert.NoError(t, err)
		assert.Equal(t, "Updated WezTerm config", updated.Description.String)
		assert.Equal(t, `{"theme": "dark", "font_size": 16}`, updated.Config)
	})

	// Test Upsert (create new)
	t.Run("UpsertNew", func(t *testing.T) {
		newEmulator := &models.TerminalEmulatorDB{
			Name:    "kitty-minimal",
			Type:    "kitty",
			Config:  `{"minimal": true}`,
			Labels:  `{"style": "minimal"}`,
			Enabled: true,
		}

		err := store.UpsertTerminalEmulator(newEmulator)
		assert.NoError(t, err)
		assert.NotZero(t, newEmulator.ID)

		retrieved, err := store.GetTerminalEmulator("kitty-minimal")
		assert.NoError(t, err)
		assert.Equal(t, "kitty", retrieved.Type)
	})

	// Test Upsert (update existing)
	t.Run("UpsertExisting", func(t *testing.T) {
		emulator1.Config = `{"theme": "dark", "font_size": 18}`

		err := store.UpsertTerminalEmulator(emulator1)
		assert.NoError(t, err)

		updated, err := store.GetTerminalEmulator("wezterm-dark")
		assert.NoError(t, err)
		assert.Equal(t, `{"theme": "dark", "font_size": 18}`, updated.Config)
	})

	// Test List
	t.Run("List", func(t *testing.T) {
		emulators, err := store.ListTerminalEmulators()
		assert.NoError(t, err)
		assert.Len(t, emulators, 3) // wezterm-dark, alacritty-light, kitty-minimal
	})

	// Test ListByType
	t.Run("ListByType", func(t *testing.T) {
		weztermEmulators, err := store.ListTerminalEmulatorsByType("wezterm")
		assert.NoError(t, err)
		assert.Len(t, weztermEmulators, 1)
		assert.Equal(t, "wezterm-dark", weztermEmulators[0].Name)

		alacrittyEmulators, err := store.ListTerminalEmulatorsByType("alacritty")
		assert.NoError(t, err)
		assert.Len(t, alacrittyEmulators, 1)
		assert.Equal(t, "alacritty-light", alacrittyEmulators[0].Name)
	})

	// Test ListByWorkspace
	t.Run("ListByWorkspace", func(t *testing.T) {
		mainWorkspaceEmulators, err := store.ListTerminalEmulatorsByWorkspace("main-workspace")
		assert.NoError(t, err)
		assert.Len(t, mainWorkspaceEmulators, 1)
		assert.Equal(t, "wezterm-dark", mainWorkspaceEmulators[0].Name)

		devWorkspaceEmulators, err := store.ListTerminalEmulatorsByWorkspace("dev-workspace")
		assert.NoError(t, err)
		assert.Len(t, devWorkspaceEmulators, 1)
		assert.Equal(t, "alacritty-light", devWorkspaceEmulators[0].Name)
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		err := store.DeleteTerminalEmulator("kitty-minimal")
		assert.NoError(t, err)

		_, err = store.GetTerminalEmulator("kitty-minimal")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "terminal emulator not found")

		// Verify other emulators still exist
		emulators, err := store.ListTerminalEmulators()
		assert.NoError(t, err)
		assert.Len(t, emulators, 2)
	})

	// Test helper methods
	t.Run("HelperMethods", func(t *testing.T) {
		retrieved, err := store.GetTerminalEmulator("wezterm-dark")
		assert.NoError(t, err)

		// Test GetConfig
		config, err := retrieved.GetConfig()
		assert.NoError(t, err)
		assert.Equal(t, "dark", config["theme"])
		assert.Equal(t, float64(18), config["font_size"]) // JSON unmarshals numbers as float64

		// Test GetLabels
		labels, err := retrieved.GetLabels()
		assert.NoError(t, err)
		assert.Equal(t, "terminal", labels["type"])
		assert.Equal(t, "dark", labels["theme"])
	})

	// Test YAML conversion
	t.Run("YAMLConversion", func(t *testing.T) {
		retrieved, err := store.GetTerminalEmulator("wezterm-dark")
		assert.NoError(t, err)

		// Convert to YAML
		yamlObj, err := retrieved.ToYAML()
		assert.NoError(t, err)
		assert.Equal(t, "devopsmaestro.io/v1", yamlObj.APIVersion)
		assert.Equal(t, "TerminalEmulator", yamlObj.Kind)
		assert.Equal(t, "wezterm-dark", yamlObj.Metadata.Name)
		assert.Equal(t, "wezterm", yamlObj.Spec.Type)
		assert.Equal(t, "dark-pro", yamlObj.Spec.ThemeRef)
		assert.Equal(t, "main-workspace", yamlObj.Spec.Workspace)

		// Convert back from YAML
		newEmulator := &models.TerminalEmulatorDB{}
		err = newEmulator.FromYAML(yamlObj)
		assert.NoError(t, err)
		assert.Equal(t, retrieved.Name, newEmulator.Name)
		assert.Equal(t, retrieved.Type, newEmulator.Type)
		assert.Equal(t, retrieved.ThemeRef, newEmulator.ThemeRef)
		assert.Equal(t, retrieved.Workspace, newEmulator.Workspace)
	})
}
