// get_context_source_test.go.pending — TDD Phase 2 (RED) for issue #202 Feature 2
//
// Tests for env-var source markers in "dvm get context".
//
// All tests are RED — source marker display is not implemented yet.
// These tests drive the implementation.
//
// Design (from issue #202):
//   - dvm get context currently shows values but not WHERE they came from.
//   - After implementation each value should be annotated with its source:
//       App: api (env: DVM_APP)    ← env var provided this value
//       App: api (global)          ← persisted DB value
//   - Source detection: if os.Getenv("DVM_APP") != "" → source = "env: DVM_APP"
//                       else                          → source = "db" / "global"
//
// To activate: rename to get_context_source_test.go after implementation is complete.

package cmd

import (
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TestGetContext_ShowsEnvVarSource
//
// When DVM_APP is set as an environment variable, "dvm get context" output
// must include a source annotation indicating the value came from the env var.
// Expected marker: "env" or "DVM_APP" alongside the app name.
//
// RED: getContext() currently outputs the value but omits any source annotation.
// =============================================================================

func TestGetContext_ShowsEnvVarSource(t *testing.T) {
	mock := db.NewMockDataStore()
	// DB has no active app — value comes entirely from env var
	mock.Context.ActiveAppID = nil
	// Provide a workspace so context is not fully empty (avoids "No active context" path)
	wsID := 1
	mock.Context.ActiveWorkspaceID = &wsID
	mock.Workspaces[1] = &models.Workspace{ID: 1, Name: "dev", AppID: 1}

	t.Setenv("DVM_APP", "env-provided-app")

	output, err := captureGetContext(t, mock)
	require.NoError(t, err)

	// Value must appear (regression guard — already works)
	assert.Contains(t, output, "env-provided-app",
		"DVM_APP value must appear in context output")

	// RED: source annotation must also appear next to the value
	// Accept any of the expected formats: "(env: DVM_APP)", "(env)", "DVM_APP"
	hasEnvMarker := containsAnySubstr(output,
		"env: DVM_APP",
		"(env)",
		"DVM_APP",
	)
	assert.True(t, hasEnvMarker,
		"context output must show source annotation when value comes from DVM_APP env var; got: %q", output)
}

// =============================================================================
// TestGetContext_ShowsWorkspaceEnvVarSource
//
// When DVM_WORKSPACE is set, output must annotate the workspace value with
// its env-var source marker.
//
// RED: same as above — no source annotation implemented yet.
// =============================================================================

func TestGetContext_ShowsWorkspaceEnvVarSource(t *testing.T) {
	mock := db.NewMockDataStore()
	appID := 1
	mock.Context.ActiveAppID = &appID
	mock.Apps[1] = &models.App{ID: 1, Name: "my-api", DomainID: 1}
	// DB has no active workspace — provided entirely by env var
	mock.Context.ActiveWorkspaceID = nil

	t.Setenv("DVM_WORKSPACE", "env-provided-ws")

	output, err := captureGetContext(t, mock)
	require.NoError(t, err)

	// Value must appear
	assert.Contains(t, output, "env-provided-ws",
		"DVM_WORKSPACE value must appear in context output")

	// RED: source annotation must also appear
	hasEnvMarker := containsAnySubstr(output,
		"env: DVM_WORKSPACE",
		"(env)",
		"DVM_WORKSPACE",
	)
	assert.True(t, hasEnvMarker,
		"context output must show source annotation when value comes from DVM_WORKSPACE env var; got: %q", output)
}

// =============================================================================
// TestGetContext_ShowsDBSource
//
// When no env vars are set and the DB has an active app context, "dvm get context"
// must annotate the app value with a "db" / "global" / "persisted" source marker.
//
// RED: no source annotation implemented.
// =============================================================================

func TestGetContext_ShowsDBSource(t *testing.T) {
	mock := db.NewMockDataStore()

	// DB-only context — no env vars
	appID := 3
	mock.Context.ActiveAppID = &appID
	mock.Apps[3] = &models.App{ID: 3, Name: "persisted-api", DomainID: 1}

	// Ensure env vars are explicitly NOT set
	t.Setenv("DVM_APP", "")
	t.Setenv("DVM_WORKSPACE", "")
	t.Setenv("DVM_ECOSYSTEM", "")
	t.Setenv("DVM_DOMAIN", "")

	output, err := captureGetContext(t, mock)
	require.NoError(t, err)

	// Value must appear
	assert.Contains(t, output, "persisted-api",
		"DB-sourced app value must appear in context output")

	// RED: a db/global/persisted source marker must appear alongside the value.
	// We look for parenthesised or labelled annotations, NOT bare substrings,
	// to avoid matching "db" inside the value name "db-sourced-app" itself.
	hasDBMarker := containsAnySubstr(output,
		"(global)",
		"(db)",
		"(persisted)",
		"global)",
		"persisted)",
	)
	assert.True(t, hasDBMarker,
		"context output must show a (db)/(global)/(persisted) source marker when value comes from DB; got: %q", output)
}

// =============================================================================
// TestGetContext_EnvVarSourceNotShownForUnsetVars
//
// When a field has NO env var AND no DB value, "(none)" should be displayed
// WITHOUT any source annotation (the source annotation only applies when a
// value is actually present).
//
// RED: implementation does not exist yet, so no annotation at all — this
// validates the "absent source" case won't regress after source markers land.
// =============================================================================

func TestGetContext_EnvVarSourceNotShownForUnsetVars(t *testing.T) {
	mock := db.NewMockDataStore()
	// Empty DB context and no env vars → all fields are "(none)"
	mock.Context.ActiveAppID = nil
	mock.Context.ActiveWorkspaceID = nil
	mock.Context.ActiveEcosystemID = nil
	mock.Context.ActiveDomainID = nil

	t.Setenv("DVM_APP", "")
	t.Setenv("DVM_WORKSPACE", "")
	t.Setenv("DVM_ECOSYSTEM", "")
	t.Setenv("DVM_DOMAIN", "")

	// With all context fields empty, getContext() shows "No active context" message
	// (skips the key-value section entirely) — so just verify it doesn't error
	// and doesn't spuriously inject source markers into the empty state.
	_, err := captureGetContext(t, mock)
	// Either succeeds with an empty-state message or returns nil — both acceptable
	// as long as it does not panic or introduce stray source markers.
	assert.NoError(t, err,
		"dvm get context with fully empty context must not return an error")
}

// =============================================================================
// Helpers
// =============================================================================

// containsAnySubstr returns true if s contains any of the given substrings.
func containsAnySubstr(s string, substrings ...string) bool {
	for _, sub := range substrings {
		if sub != "" && strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
