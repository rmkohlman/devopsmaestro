package models

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TDD Phase 2 (RED): RegistryHistoryYAML DTO Tests — Bug 6
// =============================================================================
// These tests verify that RegistryHistory.ToYAML() exists and returns a clean
// RegistryHistoryYAML DTO — without sql.Null* wrapper types in the output.
//
// These tests FAIL until:
//   1. RegistryHistoryYAML struct is added to models/registry_history.go
//   2. ToYAML() method is added to RegistryHistory
// =============================================================================

// TestRegistryHistory_ToYAML_NullableFieldsNil verifies that when sql.NullString
// fields have Valid=false, the YAML struct exposes empty string (not the raw
// sql.NullString object).
func TestRegistryHistory_ToYAML_NullableFieldsNil(t *testing.T) {
	h := &RegistryHistory{
		ID:               1,
		RegistryID:       42,
		Revision:         1,
		Config:           `{}`,
		Enabled:          true,
		Lifecycle:        "persistent",
		Port:             5001,
		Storage:          "/var/lib/zot",
		IdleTimeout:      sql.NullInt64{Valid: false},
		Action:           "start",
		Status:           "success",
		User:             sql.NullString{Valid: false},
		ErrorMessage:     sql.NullString{Valid: false},
		PreviousRevision: sql.NullInt64{Valid: false},
		RegistryVersion:  sql.NullString{Valid: false},
		CreatedAt:        time.Now(),
		CompletedAt:      sql.NullTime{Valid: false},
	}

	yamlDTO := h.ToYAML()

	// Nullable fields with Valid=false should be empty/zero values, not sql.Null* wrappers
	assert.Equal(t, "", yamlDTO.User, "User should be empty string when sql.NullString is invalid")
	assert.Equal(t, "", yamlDTO.ErrorMessage, "ErrorMessage should be empty string when sql.NullString is invalid")
	assert.Equal(t, "", yamlDTO.RegistryVersion, "RegistryVersion should be empty string when sql.NullString is invalid")

	// Non-nullable fields should be present
	assert.Equal(t, 1, yamlDTO.Revision, "Revision should be set")
	assert.Equal(t, "start", yamlDTO.Action, "Action should be set")
	assert.Equal(t, "success", yamlDTO.Status, "Status should be set")
}

// TestRegistryHistory_ToYAML_NullableFieldsPresent verifies that when
// sql.NullString has Valid=true, the YAML struct contains the actual value.
func TestRegistryHistory_ToYAML_NullableFieldsPresent(t *testing.T) {
	h := &RegistryHistory{
		ID:         2,
		RegistryID: 42,
		Revision:   3,
		Config:     `{"storage":"/var/lib/zot"}`,
		Enabled:    true,
		Lifecycle:  "on-demand",
		Port:       5001,
		Storage:    "/var/lib/zot",
		Action:     "restart",
		Status:     "success",
		User: sql.NullString{
			String: "admin",
			Valid:  true,
		},
		ErrorMessage: sql.NullString{
			String: "",
			Valid:  false,
		},
		RegistryVersion: sql.NullString{
			String: "v2.0.3",
			Valid:  true,
		},
		PreviousRevision: sql.NullInt64{
			Int64: 2,
			Valid: true,
		},
		CreatedAt: time.Now(),
		CompletedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
	}

	yamlDTO := h.ToYAML()

	// Present values should appear as plain strings
	assert.Equal(t, "admin", yamlDTO.User, "User should be the actual string value")
	assert.Equal(t, "v2.0.3", yamlDTO.RegistryVersion, "RegistryVersion should be the actual string value")
	assert.Equal(t, "", yamlDTO.ErrorMessage, "ErrorMessage should be empty string when invalid")
	assert.Equal(t, int64(2), yamlDTO.PreviousRevision, "PreviousRevision should be the int64 value")
	assert.False(t, yamlDTO.CompletedAt.IsZero(), "CompletedAt should be set")
}

// TestRegistryHistory_ToYAML_ConfigParsed verifies that the Config string
// (which is a JSON blob) is parsed into a map[string]interface{} in the YAML
// output, not left as a raw string.
func TestRegistryHistory_ToYAML_ConfigParsed(t *testing.T) {
	h := &RegistryHistory{
		ID:         3,
		RegistryID: 42,
		Revision:   5,
		Config:     `{"port":5001,"storage":"/var/lib/zot","lifecycle":"persistent"}`,
		Enabled:    true,
		Lifecycle:  "persistent",
		Port:       5001,
		Storage:    "/var/lib/zot",
		Action:     "config_change",
		Status:     "success",
		CreatedAt:  time.Now(),
	}

	yamlDTO := h.ToYAML()

	// Config should be a parsed map, not a raw string
	require.NotNil(t, yamlDTO.Config, "Config should not be nil after parsing")

	configMap, ok := yamlDTO.Config.(map[string]interface{})
	require.True(t, ok, "Config should be map[string]interface{}, got %T", yamlDTO.Config)

	// The parsed map should contain expected keys
	assert.Equal(t, float64(5001), configMap["port"], "Config should have port key")
	assert.Equal(t, "/var/lib/zot", configMap["storage"], "Config should have storage key")
	assert.Equal(t, "persistent", configMap["lifecycle"], "Config should have lifecycle key")
}

// TestRegistryHistory_ToYAML_JSONSerialization verifies that marshaling the
// YAML struct to JSON produces clean output without sql.Null wrapper types.
func TestRegistryHistory_ToYAML_JSONSerialization(t *testing.T) {
	h := &RegistryHistory{
		ID:         4,
		RegistryID: 42,
		Revision:   7,
		Config:     `{"port":5001}`,
		Enabled:    true,
		Lifecycle:  "manual",
		Port:       5001,
		Storage:    "/var/lib/zot",
		Action:     "start",
		Status:     "failed",
		User: sql.NullString{
			String: "bob",
			Valid:  true,
		},
		ErrorMessage: sql.NullString{
			String: "binary not found",
			Valid:  true,
		},
		RegistryVersion: sql.NullString{Valid: false},
		CreatedAt:       time.Now(),
	}

	yamlDTO := h.ToYAML()

	// Marshal to JSON
	data, err := json.Marshal(yamlDTO)
	require.NoError(t, err, "JSON marshaling should succeed")

	jsonStr := string(data)

	// JSON output must NOT contain sql.Null wrapper patterns like {"String":"...","Valid":true}
	assert.NotContains(t, jsonStr, `"Valid"`, "JSON should not contain sql.Null 'Valid' field")
	assert.NotContains(t, jsonStr, `"String"`, "JSON should not contain sql.Null 'String' field")

	// JSON output SHOULD contain the actual values
	assert.Contains(t, jsonStr, `"bob"`, "JSON should contain the user value 'bob'")
	assert.Contains(t, jsonStr, `"binary not found"`, "JSON should contain the error message")
}
