package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLDataStore_DefaultOperations(t *testing.T) {
	// Create in-memory database for testing
	dataStore := setupTestDB(t)
	defer dataStore.Close()

	// Create the defaults table since it won't exist in the test schema
	_, err := dataStore.Driver().Execute(`
		CREATE TABLE IF NOT EXISTS defaults (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	require.NoError(t, err)

	t.Run("GetDefault returns empty string for non-existent key", func(t *testing.T) {
		value, err := dataStore.GetDefault("nonexistent")
		require.NoError(t, err)
		assert.Equal(t, "", value)
	})

	t.Run("SetDefault and GetDefault work correctly", func(t *testing.T) {
		// Set a default
		err := dataStore.SetDefault("theme", "catppuccin")
		require.NoError(t, err)

		// Get the default
		value, err := dataStore.GetDefault("theme")
		require.NoError(t, err)
		assert.Equal(t, "catppuccin", value)
	})

	t.Run("SetDefault with JSON value works correctly", func(t *testing.T) {
		// Set a plugins default with JSON
		pluginsJSON := `["telescope.nvim", "nvim-tree.lua", "lualine.nvim"]`
		err := dataStore.SetDefault("plugins", pluginsJSON)
		require.NoError(t, err)

		// Get the plugins default
		value, err := dataStore.GetDefault("plugins")
		require.NoError(t, err)
		assert.Equal(t, pluginsJSON, value)
	})

	t.Run("SetDefault upserts existing values", func(t *testing.T) {
		// Set initial value
		err := dataStore.SetDefault("prompt", "starship")
		require.NoError(t, err)

		// Update the value
		err = dataStore.SetDefault("prompt", "oh-my-posh")
		require.NoError(t, err)

		// Verify the update
		value, err := dataStore.GetDefault("prompt")
		require.NoError(t, err)
		assert.Equal(t, "oh-my-posh", value)
	})

	t.Run("ListDefaults returns all defaults", func(t *testing.T) {
		// Set multiple defaults
		err := dataStore.SetDefault("key1", "value1")
		require.NoError(t, err)
		err = dataStore.SetDefault("key2", "value2")
		require.NoError(t, err)
		err = dataStore.SetDefault("key3", "value3")
		require.NoError(t, err)

		// List all defaults
		defaults, err := dataStore.ListDefaults()
		require.NoError(t, err)

		// Should contain at least our test keys (plus any from previous tests)
		assert.Contains(t, defaults, "key1")
		assert.Contains(t, defaults, "key2")
		assert.Contains(t, defaults, "key3")
		assert.Equal(t, "value1", defaults["key1"])
		assert.Equal(t, "value2", defaults["key2"])
		assert.Equal(t, "value3", defaults["key3"])
	})

	t.Run("DeleteDefault removes existing keys", func(t *testing.T) {
		// Set a default
		err := dataStore.SetDefault("temp_key", "temp_value")
		require.NoError(t, err)

		// Verify it exists
		value, err := dataStore.GetDefault("temp_key")
		require.NoError(t, err)
		assert.Equal(t, "temp_value", value)

		// Delete it
		err = dataStore.DeleteDefault("temp_key")
		require.NoError(t, err)

		// Verify it's gone
		value, err = dataStore.GetDefault("temp_key")
		require.NoError(t, err)
		assert.Equal(t, "", value)
	})

	t.Run("DeleteDefault does not error on non-existent keys", func(t *testing.T) {
		// Delete a non-existent key
		err := dataStore.DeleteDefault("definitely_does_not_exist")
		require.NoError(t, err) // Should not error
	})

	t.Run("ListDefaults returns empty map when no defaults exist", func(t *testing.T) {
		// Create fresh database for this test
		freshDB := setupTestDB(t)
		defer freshDB.Close()

		// Create the defaults table since it won't exist in the test schema
		_, err := freshDB.Driver().Execute(`
			CREATE TABLE IF NOT EXISTS defaults (
				key TEXT PRIMARY KEY,
				value TEXT NOT NULL,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			)
		`)
		require.NoError(t, err)

		// Clear any existing data (since setupTestDB may reuse tables)
		_, err = freshDB.Driver().Execute(`DELETE FROM defaults`)
		require.NoError(t, err)

		// List defaults on empty database
		defaults, err := freshDB.ListDefaults()
		require.NoError(t, err)
		assert.Empty(t, defaults)
	})
}

func TestMockDataStore_DefaultOperations(t *testing.T) {
	mockStore := &MockDataStore{}
	mockStore.Reset() // Initialize maps

	t.Run("GetDefault returns empty string for non-existent key", func(t *testing.T) {
		value, err := mockStore.GetDefault("nonexistent")
		require.NoError(t, err)
		assert.Equal(t, "", value)
	})

	t.Run("SetDefault and GetDefault work correctly", func(t *testing.T) {
		err := mockStore.SetDefault("theme", "catppuccin")
		require.NoError(t, err)

		value, err := mockStore.GetDefault("theme")
		require.NoError(t, err)
		assert.Equal(t, "catppuccin", value)
	})

	t.Run("ListDefaults works correctly", func(t *testing.T) {
		err := mockStore.SetDefault("key1", "value1")
		require.NoError(t, err)
		err = mockStore.SetDefault("key2", "value2")
		require.NoError(t, err)

		defaults, err := mockStore.ListDefaults()
		require.NoError(t, err)
		assert.Equal(t, "value1", defaults["key1"])
		assert.Equal(t, "value2", defaults["key2"])
	})

	t.Run("DeleteDefault works correctly", func(t *testing.T) {
		err := mockStore.SetDefault("temp", "temp_value")
		require.NoError(t, err)

		err = mockStore.DeleteDefault("temp")
		require.NoError(t, err)

		value, err := mockStore.GetDefault("temp")
		require.NoError(t, err)
		assert.Equal(t, "", value)
	})

	t.Run("Error injection works correctly", func(t *testing.T) {
		// Test GetDefault error injection
		mockStore.GetDefaultErr = assert.AnError
		_, err := mockStore.GetDefault("test")
		assert.Error(t, err)
		mockStore.GetDefaultErr = nil

		// Test SetDefault error injection
		mockStore.SetDefaultErr = assert.AnError
		err = mockStore.SetDefault("test", "value")
		assert.Error(t, err)
		mockStore.SetDefaultErr = nil

		// Test DeleteDefault error injection
		mockStore.DeleteDefaultErr = assert.AnError
		err = mockStore.DeleteDefault("test")
		assert.Error(t, err)
		mockStore.DeleteDefaultErr = nil

		// Test ListDefaults error injection
		mockStore.ListDefaultsErr = assert.AnError
		_, err = mockStore.ListDefaults()
		assert.Error(t, err)
		mockStore.ListDefaultsErr = nil
	})
}

// Helper function to set up test database (reusing existing pattern)
func setupTestDB(t *testing.T) DataStore {
	return createTestDataStore(t)
}
