package cmd

import (
	"devopsmaestro/db"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestComputeLibraryFingerprint_Deterministic verifies that the fingerprint
// is deterministic — calling it twice yields the same hash.
func TestComputeLibraryFingerprint_Deterministic(t *testing.T) {
	fp1, err := ComputeLibraryFingerprint()
	require.NoError(t, err, "first fingerprint computation should succeed")
	require.NotEmpty(t, fp1, "fingerprint should not be empty")

	fp2, err := ComputeLibraryFingerprint()
	require.NoError(t, err, "second fingerprint computation should succeed")

	assert.Equal(t, fp1, fp2, "fingerprints should be deterministic (same on repeated calls)")
}

// TestComputeLibraryFingerprint_IsSHA256 verifies the fingerprint looks
// like a valid hex-encoded SHA-256 (64 hex chars).
func TestComputeLibraryFingerprint_IsSHA256(t *testing.T) {
	fp, err := ComputeLibraryFingerprint()
	require.NoError(t, err)
	assert.Len(t, fp, 64, "SHA-256 hex digest should be 64 characters")
}

// TestEnsureLibrarySynced_ImportsWhenNoFingerprint verifies that when the
// database has no stored fingerprint (fresh install), auto-sync imports
// all library types and stores the fingerprint.
func TestEnsureLibrarySynced_ImportsWhenNoFingerprint(t *testing.T) {
	mock := db.NewMockDataStore()

	err := EnsureLibrarySynced(mock)
	require.NoError(t, err, "auto-sync should succeed on fresh database")

	// Verify fingerprint was stored
	stored, err := mock.GetDefault(defaultKeyLibraryFingerprint)
	require.NoError(t, err)
	assert.NotEmpty(t, stored, "fingerprint should be stored after sync")
	assert.Len(t, stored, 64, "stored fingerprint should be SHA-256 hex")

	// Verify plugins were imported (at least some nvim plugins exist)
	plugins, err := mock.ListPlugins()
	require.NoError(t, err)
	assert.Greater(t, len(plugins), 0, "nvim plugins should be imported")

	// Verify themes were imported
	themes, err := mock.ListThemes()
	require.NoError(t, err)
	assert.Greater(t, len(themes), 0, "nvim themes should be imported")
}

// TestEnsureLibrarySynced_SkipsWhenFingerprintMatches verifies that when the
// stored fingerprint matches the current embedded library, no import occurs.
func TestEnsureLibrarySynced_SkipsWhenFingerprintMatches(t *testing.T) {
	mock := db.NewMockDataStore()

	// First sync to populate fingerprint
	err := EnsureLibrarySynced(mock)
	require.NoError(t, err)

	// Record the call count after first sync
	firstCallCount := len(mock.Calls)

	// Second sync — should skip because fingerprints match
	err = EnsureLibrarySynced(mock)
	require.NoError(t, err)

	// Should have only added GetDefault call (the fingerprint check), no imports
	secondCallCount := len(mock.Calls)
	newCalls := secondCallCount - firstCallCount

	// The only new call should be GetDefault to check the fingerprint
	assert.Equal(t, 1, newCalls,
		"when fingerprint matches, only GetDefault should be called (no imports)")
	lastCall := mock.Calls[len(mock.Calls)-1]
	assert.Equal(t, "GetDefault", lastCall.Method,
		"the only call should be GetDefault for fingerprint check")
}

// TestEnsureLibrarySynced_ReimportsWhenFingerprintDiffers verifies that when
// the stored fingerprint differs from the current library, a full reimport occurs.
func TestEnsureLibrarySynced_ReimportsWhenFingerprintDiffers(t *testing.T) {
	mock := db.NewMockDataStore()

	// Simulate stale fingerprint from a previous library version
	mock.Defaults = map[string]string{
		defaultKeyLibraryFingerprint: "stale_fingerprint_from_old_version",
	}

	err := EnsureLibrarySynced(mock)
	require.NoError(t, err, "auto-sync should succeed when fingerprint differs")

	// Verify fingerprint was updated
	stored, err := mock.GetDefault(defaultKeyLibraryFingerprint)
	require.NoError(t, err)
	assert.NotEqual(t, "stale_fingerprint_from_old_version", stored,
		"fingerprint should be updated after reimport")
	assert.Len(t, stored, 64, "new fingerprint should be SHA-256 hex")

	// Verify imports happened
	plugins, err := mock.ListPlugins()
	require.NoError(t, err)
	assert.Greater(t, len(plugins), 0, "plugins should be reimported")
}

// TestEnsureLibrarySynced_DefaultKeyConstant verifies the defaults key
// used for fingerprint storage.
func TestEnsureLibrarySynced_DefaultKeyConstant(t *testing.T) {
	assert.Equal(t, "library.fingerprint", defaultKeyLibraryFingerprint,
		"fingerprint key should be 'library.fingerprint'")
}

// TestTruncateHash_Variations verifies the truncation helper.
func TestTruncateHash_Variations(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"long hash", "abcdef1234567890", "abcdef123456"},
		{"exactly 12", "abcdef123456", "abcdef123456"},
		{"short string", "abc", "abc"},
		{"empty string", "", "(none)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, truncateHash(tt.input))
		})
	}
}
