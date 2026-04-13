// get_context_display_test.go.pending (continued)
// Part 2: Domain, override, combined, and help text tests.

package cmd

import (
	"database/sql"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// TestGetContext_ShowsDomainEnvVar
// RED: DVM_DOMAIN is NOT checked in getContext() (same bug as ecosystem).
// =============================================================================

func TestGetContext_ShowsDomainEnvVar(t *testing.T) {
	mock := db.NewMockDataStore()
	mock.Context.ActiveDomainID = nil
	appID := 1
	mock.Context.ActiveAppID = &appID
	mock.Apps[1] = &models.App{ID: 1, Name: "my-api", DomainID: sql.NullInt64{Int64: 1, Valid: true}}

	t.Setenv("DVM_DOMAIN", "mydom")

	output, err := captureGetContext(t, mock)
	require.NoError(t, err)

	// RED before fix: domain shows "(none)" even with DVM_DOMAIN set
	assert.True(t, strings.Contains(output, "mydom"),
		"DVM_DOMAIN must override domain display; got: %q", output)
}

// =============================================================================
// TestGetContext_EnvVarOverridesDB
// RED: DVM_ECOSYSTEM must OVERRIDE the DB-resolved value.
// DB has ecosystem "prod" but env var says "dev" → output must show "dev".
// =============================================================================

func TestGetContext_EnvVarOverridesDB(t *testing.T) {
	mock := db.NewMockDataStore()
	ecoID := 1
	mock.Context.ActiveEcosystemID = &ecoID
	mock.Ecosystems["prod"] = &models.Ecosystem{ID: 1, Name: "prod"}

	t.Setenv("DVM_ECOSYSTEM", "dev")

	output, err := captureGetContext(t, mock)
	require.NoError(t, err)

	// RED: currently shows "prod" from DB, ignores DVM_ECOSYSTEM
	assert.True(t, strings.Contains(output, "dev"),
		"DVM_ECOSYSTEM=dev must override DB ecosystem 'prod'; got: %q", output)
	assert.False(t, strings.Contains(output, "prod"),
		"DB value 'prod' must NOT appear when DVM_ECOSYSTEM=dev overrides it; got: %q", output)
}

// =============================================================================
// TestGetContext_AllFourEnvVarsTogether
// RED for ecosystem+domain; GREEN for app+workspace.
// All 4 env vars set simultaneously must all appear in output.
// =============================================================================

func TestGetContext_AllFourEnvVarsTogether(t *testing.T) {
	mock := db.NewMockDataStore()
	// Empty DB context
	mock.Context.ActiveEcosystemID = nil
	mock.Context.ActiveDomainID = nil
	mock.Context.ActiveAppID = nil
	mock.Context.ActiveWorkspaceID = nil

	t.Setenv("DVM_ECOSYSTEM", "env-eco")
	t.Setenv("DVM_DOMAIN", "env-dom")
	t.Setenv("DVM_APP", "env-app")
	t.Setenv("DVM_WORKSPACE", "env-ws")

	output, err := captureGetContext(t, mock)
	require.NoError(t, err)

	// GREEN today (already implemented):
	assert.Contains(t, output, "env-app",
		"DVM_APP must appear in context display")
	assert.Contains(t, output, "env-ws",
		"DVM_WORKSPACE must appear in context display")

	// RED until Change 2 is implemented:
	assert.True(t, strings.Contains(output, "env-eco"),
		"DVM_ECOSYSTEM must appear in context display; got: %q", output)
	assert.True(t, strings.Contains(output, "env-dom"),
		"DVM_DOMAIN must appear in context display; got: %q", output)
}

// =============================================================================
// TestGetContext_LongDescMentionsEcosystemDomainEnvVars
// RED: getContextCmd.Long help text currently only mentions DVM_APP and DVM_WORKSPACE.
// After the fix, it must also document DVM_ECOSYSTEM and DVM_DOMAIN.
// =============================================================================

func TestGetContext_LongDescMentionsEcosystemDomainEnvVars(t *testing.T) {
	long := getContextCmd.Long

	// GREEN today (regression guard):
	assert.Contains(t, long, "DVM_APP",
		"getContextCmd help should document DVM_APP")
	assert.Contains(t, long, "DVM_WORKSPACE",
		"getContextCmd help should document DVM_WORKSPACE")

	// RED until help text is updated:
	assert.Contains(t, long, "DVM_ECOSYSTEM",
		"getContextCmd help must document DVM_ECOSYSTEM env var")
	assert.Contains(t, long, "DVM_DOMAIN",
		"getContextCmd help must document DVM_DOMAIN env var")
}
