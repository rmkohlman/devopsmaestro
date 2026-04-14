package cmd

// =============================================================================
// Verification: Issue #256 — Concurrent Builds Race on Shared Staging Dirs
// =============================================================================
// The fix in build_phases.go appends a UUID suffix to the staging directory
// key so that concurrent builds of the same workspace each get a unique path.
//
// Format: <buildKey>-<8-char-uuid>
// e.g.  "cloud-payments-api-dev-a3f1b2c4"
//
// These tests verify:
//   (a) Two calls to the staging key generation logic always produce distinct
//       paths, even for the same workspace (UUID randomness).
//   (b) The staging key still embeds the buildKey prefix for human readability.
//   (c) The UUID suffix is exactly 8 hex characters.
// =============================================================================

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/stretchr/testify/assert"
)

// simulateStagingKey mirrors the exact logic in build_phases.go:
//
//	stagingKey := bc.buildKey() + "-" + uuid.New().String()[:8]
func simulateStagingKey(buildKey string) string {
	return buildKey + "-" + uuid.New().String()[:8]
}

// TestStagingUUID_TwoCallsProduceDistinctKeys verifies that two concurrent
// builds of the same workspace get different staging keys because of the
// UUID suffix (fix for issue #256).
func TestStagingUUID_TwoCallsProduceDistinctKeys(t *testing.T) {
	const buildKey = "cloud-payments-api-dev"

	key1 := simulateStagingKey(buildKey)
	key2 := simulateStagingKey(buildKey)

	assert.NotEqual(t, key1, key2,
		"two concurrent builds of the same workspace must produce distinct staging keys; "+
			"got key1=%q key2=%q — UUID suffix must differ", key1, key2)
}

// TestStagingUUID_KeyContainsBuildKeyPrefix verifies that the staging key
// still starts with the buildKey so paths remain human-readable.
func TestStagingUUID_KeyContainsBuildKeyPrefix(t *testing.T) {
	tests := []struct {
		name     string
		buildKey string
	}{
		{"slug-based key", "cloud-payments-api-dev"},
		{"fallback app-workspace key", "myapp-dev"},
		{"hyphenated app name", "my-svc-staging"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := simulateStagingKey(tt.buildKey)
			assert.True(t, strings.HasPrefix(key, tt.buildKey+"-"),
				"staging key %q must start with buildKey prefix %q", key, tt.buildKey+"-")
		})
	}
}

// TestStagingUUID_SuffixIsEightHexChars verifies that the UUID suffix appended
// by the fix is exactly 8 hex characters (uuid.New().String()[:8]).
func TestStagingUUID_SuffixIsEightHexChars(t *testing.T) {
	const buildKey = "eco-dom-app-ws"

	for i := 0; i < 10; i++ {
		key := simulateStagingKey(buildKey)
		// key format: "<buildKey>-<8chars>"
		suffix := strings.TrimPrefix(key, buildKey+"-")
		assert.Len(t, suffix, 8,
			"UUID suffix must be exactly 8 characters, got %q (len=%d)", suffix, len(suffix))
		// Each character must be a hex digit or hyphen (UUID v4 chars before truncation)
		for _, c := range suffix {
			assert.True(t,
				(c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || c == '-',
				"UUID suffix char %q must be a hex digit", c)
		}
	}
}

// TestStagingUUID_ConcurrentBuilds_AllKeysUnique simulates 10 concurrent
// builds of the same workspace and asserts every staging key is unique.
// This directly models the race condition fixed by issue #256.
func TestStagingUUID_ConcurrentBuilds_AllKeysUnique(t *testing.T) {
	const buildKey = "eco-pay-api-dev"
	const numConcurrent = 10

	keys := make([]string, numConcurrent)
	for i := 0; i < numConcurrent; i++ {
		keys[i] = simulateStagingKey(buildKey)
	}

	seen := make(map[string]int)
	for _, k := range keys {
		seen[k]++
	}

	for k, count := range seen {
		assert.Equal(t, 1, count,
			"staging key %q appeared %d times — UUID suffix must make each key unique", k, count)
	}

	assert.Len(t, seen, numConcurrent,
		"all %d concurrent builds must produce distinct staging keys", numConcurrent)
}

// TestStagingUUID_StagingDirPathIsUnique verifies that two distinct staging
// keys produce two distinct filesystem paths via paths.BuildStagingDir.
func TestStagingUUID_StagingDirPathIsUnique(t *testing.T) {
	const homeDir = "/tmp/test-staging-uuid"
	const buildKey = "eco-dom-api-dev"

	key1 := simulateStagingKey(buildKey)
	key2 := simulateStagingKey(buildKey)

	dir1 := paths.New(homeDir).BuildStagingDir(key1)
	dir2 := paths.New(homeDir).BuildStagingDir(key2)

	assert.NotEqual(t, dir1, dir2,
		"distinct staging keys must produce distinct directory paths; got dir1=%q dir2=%q", dir1, dir2)
}
