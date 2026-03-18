// Package cmd_test contains Phase 2 RED tests for global build arg helper functions.
//
// RED PHASE: This file tests functions in cmd/build_arg_helpers.go, which does NOT EXIST YET.
// These tests WILL NOT COMPILE until the following are implemented:
//
//   - cmd/build_arg_helpers.go with:
//     func GetGlobalBuildArgs(ds db.DataStore) (map[string]string, error)
//     func SetGlobalBuildArg(ds db.DataStore, key, value string) error
//     func DeleteGlobalBuildArg(ds db.DataStore, key string) error
//
// Design contract (from v0.55.0 sprint plan WI-4):
//   - Global build args are stored in the `defaults` table under key "build-args"
//   - The value is a JSON-encoded map[string]string
//   - GetGlobalBuildArgs returns an empty (non-nil) map when no args exist
//   - SetGlobalBuildArg is a read-modify-write operation: it reads the current
//     JSON map, sets/updates the given key, and writes back
//   - DeleteGlobalBuildArg is a read-modify-delete: it reads, removes the key, writes back
//   - Both Set and Delete validate the key with ValidateEnvKey() before any DB write
//   - Keys that are dangerous (IsDangerousEnvVar()) are also rejected
//
// This mirrors the pattern established by `dvm set theme` where the global
// default theme is stored in defaults["theme"].
package cmd

import (
	"testing"

	// RED: GetGlobalBuildArgs, SetGlobalBuildArg, DeleteGlobalBuildArg do not exist yet
	// They will live in cmd/build_arg_helpers.go (same package as this test file)

	"devopsmaestro/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// GetGlobalBuildArgs
// =============================================================================

// TestGetGlobalBuildArgs_Empty verifies that GetGlobalBuildArgs returns an empty
// non-nil map when the "build-args" key does not exist in the defaults table.
func TestGetGlobalBuildArgs_Empty(t *testing.T) {
	// Arrange: fresh store with no defaults
	store := db.NewMockDataStore()

	// Act
	// RED: GetGlobalBuildArgs does not exist yet
	args, err := GetGlobalBuildArgs(store)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, args,
		"GetGlobalBuildArgs must return a non-nil map when no args are set")
	assert.Empty(t, args,
		"GetGlobalBuildArgs must return an empty map when 'build-args' key is absent")
}

// TestGetGlobalBuildArgs_ReturnsStoredValues verifies that args previously stored
// via SetGlobalBuildArg are returned by GetGlobalBuildArgs.
func TestGetGlobalBuildArgs_ReturnsStoredValues(t *testing.T) {
	store := db.NewMockDataStore()

	require.NoError(t, SetGlobalBuildArg(store, "PIP_INDEX_URL", "https://pypi.example"))

	args, err := GetGlobalBuildArgs(store)

	require.NoError(t, err)
	assert.Equal(t, "https://pypi.example", args["PIP_INDEX_URL"])
}

// =============================================================================
// SetGlobalBuildArg
// =============================================================================

// TestSetGlobalBuildArg_NewKey verifies that setting a new key creates it in the
// defaults table JSON blob and is retrievable via GetGlobalBuildArgs.
func TestSetGlobalBuildArg_NewKey(t *testing.T) {
	store := db.NewMockDataStore()

	// RED: SetGlobalBuildArg does not exist yet
	err := SetGlobalBuildArg(store, "PIP_INDEX_URL", "https://pypi.example")

	require.NoError(t, err)

	args, err := GetGlobalBuildArgs(store)
	require.NoError(t, err)
	assert.Equal(t, "https://pypi.example", args["PIP_INDEX_URL"],
		"newly set key must be retrievable")
}

// TestSetGlobalBuildArg_UpdateExistingKey verifies that setting a key that already
// exists updates it (last-write-wins), not duplicates it.
func TestSetGlobalBuildArg_UpdateExistingKey(t *testing.T) {
	store := db.NewMockDataStore()

	require.NoError(t, SetGlobalBuildArg(store, "PIP_INDEX_URL", "old-value"))
	require.NoError(t, SetGlobalBuildArg(store, "PIP_INDEX_URL", "new-value"))

	args, err := GetGlobalBuildArgs(store)
	require.NoError(t, err)
	assert.Equal(t, "new-value", args["PIP_INDEX_URL"],
		"second Set must overwrite first Set for the same key")
	assert.Len(t, args, 1,
		"there must be exactly one entry — not two — after updating an existing key")
}

// TestSetGlobalBuildArg_MultipleKeys verifies that multiple independent keys can
// coexist in the global build args store.
func TestSetGlobalBuildArg_MultipleKeys(t *testing.T) {
	store := db.NewMockDataStore()

	require.NoError(t, SetGlobalBuildArg(store, "KEY_A", "val-a"))
	require.NoError(t, SetGlobalBuildArg(store, "KEY_B", "val-b"))
	require.NoError(t, SetGlobalBuildArg(store, "KEY_C", "val-c"))

	args, err := GetGlobalBuildArgs(store)
	require.NoError(t, err)
	assert.Equal(t, "val-a", args["KEY_A"])
	assert.Equal(t, "val-b", args["KEY_B"])
	assert.Equal(t, "val-c", args["KEY_C"])
	assert.Len(t, args, 3,
		"all 3 distinct keys must be stored independently")
}

// TestSetGlobalBuildArg_InvalidKey verifies that SetGlobalBuildArg returns an error
// when the key fails ValidateEnvKey() — e.g. starts with a digit or contains hyphens.
func TestSetGlobalBuildArg_InvalidKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr string
	}{
		{
			name:    "starts with digit",
			key:     "123-bad",
			wantErr: "invalid",
		},
		{
			name:    "contains lowercase letters",
			key:     "pip_index_url",
			wantErr: "invalid",
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: "invalid",
		},
		{
			name:    "contains hyphen",
			key:     "PIP-INDEX-URL",
			wantErr: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := db.NewMockDataStore()

			// RED: SetGlobalBuildArg does not exist yet
			err := SetGlobalBuildArg(store, tt.key, "some-value")

			assert.Error(t, err,
				"SetGlobalBuildArg must return an error for invalid key %q", tt.key)
			assert.ErrorContains(t, err, tt.wantErr,
				"error message for key %q must mention 'invalid'", tt.key)
		})
	}
}

// TestSetGlobalBuildArg_DangerousKey verifies that SetGlobalBuildArg returns an error
// when the key is in the IsDangerousEnvVar() security denylist, preventing injection
// of LD_PRELOAD, BASH_ENV, etc. at the global level.
func TestSetGlobalBuildArg_DangerousKey(t *testing.T) {
	dangerousKeys := []string{
		"LD_PRELOAD",
		"LD_LIBRARY_PATH",
		"DYLD_INSERT_LIBRARIES",
		"DYLD_LIBRARY_PATH",
		"NODE_OPTIONS",
		"BASH_ENV",
	}

	for _, key := range dangerousKeys {
		t.Run(key, func(t *testing.T) {
			store := db.NewMockDataStore()

			// RED: SetGlobalBuildArg does not exist yet
			err := SetGlobalBuildArg(store, key, "/evil.so")

			assert.Error(t, err,
				"SetGlobalBuildArg must reject dangerous key %q", key)
			// The error should mention forbidden/dangerous/denylist
			assert.True(t,
				errorContainsAny(err, "forbidden", "dangerous", "denylist", "security"),
				"error for dangerous key %q must mention security restriction, got: %v", key, err)
		})
	}
}

// TestSetGlobalBuildArg_DVMReservedKey verifies that keys using the reserved DVM_
// prefix are rejected by SetGlobalBuildArg.
func TestSetGlobalBuildArg_DVMReservedKey(t *testing.T) {
	store := db.NewMockDataStore()

	err := SetGlobalBuildArg(store, "DVM_INTERNAL", "value")

	assert.Error(t, err,
		"SetGlobalBuildArg must reject DVM_ reserved prefix keys")
}

// =============================================================================
// DeleteGlobalBuildArg
// =============================================================================

// TestDeleteGlobalBuildArg_ExistingKey verifies that DeleteGlobalBuildArg removes
// the target key while leaving other keys intact.
func TestDeleteGlobalBuildArg_ExistingKey(t *testing.T) {
	store := db.NewMockDataStore()

	require.NoError(t, SetGlobalBuildArg(store, "KEY_A", "val-a"))
	require.NoError(t, SetGlobalBuildArg(store, "KEY_B", "val-b"))

	// RED: DeleteGlobalBuildArg does not exist yet
	err := DeleteGlobalBuildArg(store, "KEY_A")

	require.NoError(t, err)

	args, err := GetGlobalBuildArgs(store)
	require.NoError(t, err)
	assert.NotContains(t, args, "KEY_A",
		"KEY_A must be removed after DeleteGlobalBuildArg")
	assert.Contains(t, args, "KEY_B",
		"KEY_B must remain untouched after deleting KEY_A")
	assert.Equal(t, "val-b", args["KEY_B"],
		"KEY_B value must be unchanged")
}

// TestDeleteGlobalBuildArg_NonExistentKey verifies that deleting a key that was
// never set is a no-op (no error).
func TestDeleteGlobalBuildArg_NonExistentKey(t *testing.T) {
	store := db.NewMockDataStore()
	// No keys set at all

	// RED: DeleteGlobalBuildArg does not exist yet
	err := DeleteGlobalBuildArg(store, "KEY_A")

	assert.NoError(t, err,
		"DeleteGlobalBuildArg on a non-existent key must not return an error")
}

// TestDeleteGlobalBuildArg_LastKey verifies that deleting the last key leaves the
// build args store in a clean state (empty map, not nil).
func TestDeleteGlobalBuildArg_LastKey(t *testing.T) {
	store := db.NewMockDataStore()

	require.NoError(t, SetGlobalBuildArg(store, "ONLY_KEY", "val"))
	require.NoError(t, DeleteGlobalBuildArg(store, "ONLY_KEY"))

	args, err := GetGlobalBuildArgs(store)
	require.NoError(t, err)
	assert.NotNil(t, args,
		"GetGlobalBuildArgs must return a non-nil map even after all keys are deleted")
	assert.Empty(t, args,
		"map must be empty after deleting the last key")
}

// TestDeleteGlobalBuildArg_AllThreeKeys verifies a sequence of set/delete/get
// operations that exercises the read-modify-write lifecycle end-to-end.
func TestDeleteGlobalBuildArg_AllThreeKeys(t *testing.T) {
	store := db.NewMockDataStore()

	// Set 3 keys
	require.NoError(t, SetGlobalBuildArg(store, "KEY_A", "a"))
	require.NoError(t, SetGlobalBuildArg(store, "KEY_B", "b"))
	require.NoError(t, SetGlobalBuildArg(store, "KEY_C", "c"))

	// Delete them one by one and verify state after each deletion
	require.NoError(t, DeleteGlobalBuildArg(store, "KEY_B"))
	args, err := GetGlobalBuildArgs(store)
	require.NoError(t, err)
	assert.Len(t, args, 2)
	assert.NotContains(t, args, "KEY_B")

	require.NoError(t, DeleteGlobalBuildArg(store, "KEY_A"))
	args, err = GetGlobalBuildArgs(store)
	require.NoError(t, err)
	assert.Len(t, args, 1)
	assert.Equal(t, "c", args["KEY_C"])

	require.NoError(t, DeleteGlobalBuildArg(store, "KEY_C"))
	args, err = GetGlobalBuildArgs(store)
	require.NoError(t, err)
	assert.Empty(t, args)
}

// =============================================================================
// Helper
// =============================================================================

// errorContainsAny returns true if the error message contains any of the given substrings.
// This avoids hard-coding the exact error phrasing from envvalidation.
func errorContainsAny(err error, substrings ...string) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	for _, s := range substrings {
		if containsStr(msg, s) {
			return true
		}
	}
	return false
}

// containsStr is a simple case-insensitive substring check.
func containsStr(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if equalFold(s[i:i+len(sub)], sub) {
			return true
		}
	}
	return false
}

// equalFold is a simple byte-level case-insensitive compare for ASCII strings.
func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}
