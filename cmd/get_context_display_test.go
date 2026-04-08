// get_context_display_test.go.pending — TDD Phase 2 (RED) for issue #201 Change 2
//
// Tests validate desired behavior AFTER the env var display fix in getContext().
//
// Status of each test:
//   RED  — TestGetContext_ShowsEcosystemEnvVar: DVM_ECOSYSTEM not checked in getContext()
//   RED  — TestGetContext_ShowsDomainEnvVar: DVM_DOMAIN not checked in getContext()
//   RED  — TestGetContext_EnvVarOverridesDB: DVM_ECOSYSTEM overrides DB value (not implemented)
//   RED  — TestGetContext_LongDescMentionsEcosystemDomainEnvVars: help text not updated yet
//   GREEN — TestGetContext_ShowsAppEnvVar: DVM_APP already supported (regression guard)
//   GREEN — TestGetContext_ShowsWorkspaceEnvVar: DVM_WORKSPACE already supported
//
// To activate: rename to get_context_display_test.go after implementation is complete.

package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Helpers
// =============================================================================

// captureGetContext executes getContext() with a MockDataStore and captures all
// render output using render.SetWriter — matching the gitrepo_test.go pattern.
func captureGetContext(t *testing.T, mock *db.MockDataStore) (string, error) {
	t.Helper()

	ctx := context.WithValue(context.Background(), CtxKeyDataStore, mock)
	getContextCmd.SetContext(ctx)

	var buf bytes.Buffer
	originalWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(originalWriter)

	originalFormat := getOutputFormat
	getOutputFormat = ""
	defer func() { getOutputFormat = originalFormat }()

	err := getContextCmd.RunE(getContextCmd, []string{})
	return buf.String(), err
}

// =============================================================================
// TestGetContext_ShowsAppEnvVar (GREEN — regression guard)
// DVM_APP already handled in getContext(). Confirms it keeps working.
// =============================================================================

func TestGetContext_ShowsAppEnvVar(t *testing.T) {
	mock := db.NewMockDataStore()
	mock.Context.ActiveAppID = nil
	// Add workspace so context isn't fully empty (avoids "No active context" path)
	wsID := 1
	mock.Context.ActiveWorkspaceID = &wsID
	mock.Workspaces[1] = &models.Workspace{ID: 1, Name: "dev", AppID: 1}

	t.Setenv("DVM_APP", "my-env-app")

	output, err := captureGetContext(t, mock)
	require.NoError(t, err)
	assert.Contains(t, output, "my-env-app",
		"DVM_APP env var should override display (already implemented)")
}

// =============================================================================
// TestGetContext_ShowsWorkspaceEnvVar (GREEN — regression guard)
// =============================================================================

func TestGetContext_ShowsWorkspaceEnvVar(t *testing.T) {
	mock := db.NewMockDataStore()
	appID := 1
	mock.Context.ActiveAppID = &appID
	mock.Apps[1] = &models.App{ID: 1, Name: "my-api", DomainID: 1}

	t.Setenv("DVM_WORKSPACE", "my-env-ws")

	output, err := captureGetContext(t, mock)
	require.NoError(t, err)
	assert.Contains(t, output, "my-env-ws",
		"DVM_WORKSPACE env var should override display (already implemented)")
}

// =============================================================================
// TestGetContext_ShowsEcosystemEnvVar
// RED: DVM_ECOSYSTEM is NOT checked in getContext() (get_resources.go:63-69).
// Expected: output shows "myeco" as ecosystem.
// =============================================================================

func TestGetContext_ShowsEcosystemEnvVar(t *testing.T) {
	mock := db.NewMockDataStore()
	mock.Context.ActiveEcosystemID = nil
	appID := 1
	mock.Context.ActiveAppID = &appID
	mock.Apps[1] = &models.App{ID: 1, Name: "my-api", DomainID: 1}

	t.Setenv("DVM_ECOSYSTEM", "myeco")

	output, err := captureGetContext(t, mock)
	require.NoError(t, err)

	// RED before fix: ecosystem shows "(none)" even with DVM_ECOSYSTEM set
	assert.True(t, strings.Contains(output, "myeco"),
		"DVM_ECOSYSTEM must override ecosystem display; got: %q", output)
}
