package cmd

// use_ecosystem_domain_test.go — TDD Phase 2 (failing tests) for issue #198
//
// Tests for:
//   - dvm use ecosystem <name>  (useEcosystemCmd)
//   - dvm use domain <name>     (useDomainCmd)
//   - --export flag on all dvm use subcommands
//
// These tests FAIL until the implementation is added in cmd/use.go.

import (
	"database/sql"
	"bytes"
	"context"
	"strings"
	"testing"

	"devopsmaestro/db"
	"devopsmaestro/models"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Helpers
// =============================================================================

// newCmdContextWithDS returns a context.Context carrying a MockDataStore,
// matching the pattern used by getDataStore().
func newCmdContextWithDS(ds *db.MockDataStore) context.Context {
	return context.WithValue(context.Background(), CtxKeyDataStore, ds)
}

// =============================================================================
// TestUseEcosystemCmd — Command Structure
// =============================================================================

// TestUseEcosystemCmdExists verifies the use ecosystem subcommand is registered.
func TestUseEcosystemCmdExists(t *testing.T) {
	assert.NotNil(t, useEcosystemCmd, "useEcosystemCmd should exist")
}

// TestUseEcosystemCmdRegisteredUnderUse verifies 'ecosystem' appears as a subcommand of 'use'.
func TestUseEcosystemCmdRegisteredUnderUse(t *testing.T) {
	subcommands := useCmd.Commands()
	names := make([]string, 0, len(subcommands))
	for _, c := range subcommands {
		names = append(names, c.Name())
	}
	assert.Contains(t, names, "ecosystem", "use should have 'ecosystem' subcommand")
}

// TestUseEcosystemCmdRequiresOneArg verifies use ecosystem requires exactly 1 argument.
func TestUseEcosystemCmdRequiresOneArg(t *testing.T) {
	assert.NotNil(t, useEcosystemCmd.Args, "useEcosystemCmd should have Args validator")
	// Zero args should fail
	err := cobra.ExactArgs(1)(nil, []string{})
	assert.Error(t, err, "use ecosystem should require exactly 1 arg")
	// One arg should pass
	err = cobra.ExactArgs(1)(nil, []string{"myeco"})
	assert.NoError(t, err)
}

// TestUseEcosystemCmdHasRunE verifies useEcosystemCmd uses RunE.
func TestUseEcosystemCmdHasRunE(t *testing.T) {
	assert.NotNil(t, useEcosystemCmd.RunE, "useEcosystemCmd should have RunE")
}

// TestUseEcosystemCmdShortDescription verifies Short description is set.
func TestUseEcosystemCmdShortDescription(t *testing.T) {
	assert.NotEmpty(t, useEcosystemCmd.Short, "useEcosystemCmd should have Short description")
}

// TestUseEcosystemCmdLongDescriptionMentionsNone verifies 'none' is mentioned.
func TestUseEcosystemCmdLongDescriptionMentionsNone(t *testing.T) {
	assert.Contains(t, useEcosystemCmd.Long, "none",
		"useEcosystemCmd Long should mention 'none' for clearing context")
}

// =============================================================================
// TestUseEcosystemCmd — Behavior
// =============================================================================

// TestUseEcosystem_SetsActiveEcosystemInDB verifies that dvm use ecosystem <name>
// calls SetActiveEcosystem with the correct ecosystem ID in the database.
func TestUseEcosystem_SetsActiveEcosystemInDB(t *testing.T) {
	mock := db.NewMockDataStore()

	// Seed an ecosystem
	eco := &models.Ecosystem{ID: 10, Name: "myeco"}
	mock.Ecosystems["myeco"] = eco

	// Execute the command
	useEcosystemCmd.SetContext(newCmdContextWithDS(mock))
	err := useEcosystemCmd.RunE(useEcosystemCmd, []string{"myeco"})

	require.NoError(t, err)
	require.NotNil(t, mock.Context.ActiveEcosystemID, "ActiveEcosystemID should be set")
	assert.Equal(t, 10, *mock.Context.ActiveEcosystemID, "ActiveEcosystemID should be 10")
}

// TestUseEcosystem_ClearsCascadingContext verifies that switching ecosystem clears
// domain, app, and workspace context (cascade clear).
func TestUseEcosystem_ClearsCascadingContext(t *testing.T) {
	mock := db.NewMockDataStore()

	// Seed ecosystem
	eco := &models.Ecosystem{ID: 1, Name: "myeco"}
	mock.Ecosystems["myeco"] = eco

	// Pre-set domain, app, and workspace IDs to simulate existing context
	domID, appID, wsID := 5, 10, 20
	mock.Context.ActiveDomainID = &domID
	mock.Context.ActiveAppID = &appID
	mock.Context.ActiveWorkspaceID = &wsID

	useEcosystemCmd.SetContext(newCmdContextWithDS(mock))
	err := useEcosystemCmd.RunE(useEcosystemCmd, []string{"myeco"})

	require.NoError(t, err)
	assert.Nil(t, mock.Context.ActiveDomainID,
		"domain should be cleared when switching ecosystem")
	assert.Nil(t, mock.Context.ActiveAppID,
		"app should be cleared when switching ecosystem")
	assert.Nil(t, mock.Context.ActiveWorkspaceID,
		"workspace should be cleared when switching ecosystem")
}

// TestUseEcosystem_NoneClears verifies that dvm use ecosystem none clears ecosystem context.
func TestUseEcosystem_NoneClears(t *testing.T) {
	mock := db.NewMockDataStore()

	// Pre-set an active ecosystem
	ecoID := 5
	mock.Context.ActiveEcosystemID = &ecoID

	useEcosystemCmd.SetContext(newCmdContextWithDS(mock))
	err := useEcosystemCmd.RunE(useEcosystemCmd, []string{"none"})

	require.NoError(t, err)
	assert.Nil(t, mock.Context.ActiveEcosystemID,
		"ecosystem context should be cleared with 'none'")
}

// TestUseEcosystem_NonExistentReturnsError verifies that a non-existent ecosystem
// results in an error.
func TestUseEcosystem_NonExistentReturnsError(t *testing.T) {
	mock := db.NewMockDataStore()
	// No ecosystems seeded

	useEcosystemCmd.SetContext(newCmdContextWithDS(mock))
	err := useEcosystemCmd.RunE(useEcosystemCmd, []string{"does-not-exist"})

	// Should return errSilent (error path, rendered via render.Error)
	assert.Error(t, err, "non-existent ecosystem should return an error")
}

// =============================================================================
// TestUseDomainCmd — Command Structure
// =============================================================================

// TestUseDomainCmdExists verifies the use domain subcommand is registered.
func TestUseDomainCmdExists(t *testing.T) {
	assert.NotNil(t, useDomainCmd, "useDomainCmd should exist")
}

// TestUseDomainCmdRegisteredUnderUse verifies 'domain' appears as a subcommand of 'use'.
func TestUseDomainCmdRegisteredUnderUse(t *testing.T) {
	subcommands := useCmd.Commands()
	names := make([]string, 0, len(subcommands))
	for _, c := range subcommands {
		names = append(names, c.Name())
	}
	assert.Contains(t, names, "domain", "use should have 'domain' subcommand")
}

// TestUseDomainCmdRequiresOneArg verifies use domain requires exactly 1 argument.
func TestUseDomainCmdRequiresOneArg(t *testing.T) {
	assert.NotNil(t, useDomainCmd.Args, "useDomainCmd should have Args validator")
	// Zero args should fail
	err := cobra.ExactArgs(1)(nil, []string{})
	assert.Error(t, err, "use domain should require exactly 1 arg")
	// One arg should pass
	err = cobra.ExactArgs(1)(nil, []string{"mydom"})
	assert.NoError(t, err)
}

// TestUseDomainCmdHasRunE verifies useDomainCmd uses RunE.
func TestUseDomainCmdHasRunE(t *testing.T) {
	assert.NotNil(t, useDomainCmd.RunE, "useDomainCmd should have RunE")
}

// TestUseDomainCmdShortDescription verifies Short description is set.
func TestUseDomainCmdShortDescription(t *testing.T) {
	assert.NotEmpty(t, useDomainCmd.Short, "useDomainCmd should have Short description")
}

// TestUseDomainCmdLongDescriptionMentionsNone verifies 'none' is mentioned.
func TestUseDomainCmdLongDescriptionMentionsNone(t *testing.T) {
	assert.Contains(t, useDomainCmd.Long, "none",
		"useDomainCmd Long should mention 'none' for clearing context")
}

// =============================================================================
// TestUseDomainCmd — Behavior
// =============================================================================

// TestUseDomain_SetsActiveDomainInDB verifies that dvm use domain <name>
// calls SetActiveDomain with the correct domain ID in the database.
func TestUseDomain_SetsActiveDomainInDB(t *testing.T) {
	mock := db.NewMockDataStore()

	// Seed ecosystem and domain
	ecoID := 1
	mock.Context.ActiveEcosystemID = &ecoID
	mock.Ecosystems["prod"] = &models.Ecosystem{ID: 1, Name: "prod"}
	domain := &models.Domain{ID: 7, Name: "mydom", EcosystemID: sql.NullInt64{Int64: 1, Valid: true}}
	mock.Domains[7] = domain

	useDomainCmd.SetContext(newCmdContextWithDS(mock))
	err := useDomainCmd.RunE(useDomainCmd, []string{"mydom"})

	require.NoError(t, err)
	require.NotNil(t, mock.Context.ActiveDomainID, "ActiveDomainID should be set")
	assert.Equal(t, 7, *mock.Context.ActiveDomainID, "ActiveDomainID should be 7")
}

// TestUseDomain_ClearsCascadingContext verifies that switching domain clears
// app and workspace context (but NOT ecosystem).
func TestUseDomain_ClearsCascadingContext(t *testing.T) {
	mock := db.NewMockDataStore()

	// Seed ecosystem and domain
	ecoID := 1
	mock.Context.ActiveEcosystemID = &ecoID
	mock.Ecosystems["prod"] = &models.Ecosystem{ID: 1, Name: "prod"}
	domain := &models.Domain{ID: 3, Name: "mydom", EcosystemID: sql.NullInt64{Int64: 1, Valid: true}}
	mock.Domains[3] = domain

	// Pre-set app and workspace to simulate existing context
	appID, wsID := 10, 20
	mock.Context.ActiveAppID = &appID
	mock.Context.ActiveWorkspaceID = &wsID

	useDomainCmd.SetContext(newCmdContextWithDS(mock))
	err := useDomainCmd.RunE(useDomainCmd, []string{"mydom"})

	require.NoError(t, err)
	assert.Nil(t, mock.Context.ActiveAppID,
		"app should be cleared when switching domain")
	assert.Nil(t, mock.Context.ActiveWorkspaceID,
		"workspace should be cleared when switching domain")
	// Ecosystem must NOT be cleared
	assert.NotNil(t, mock.Context.ActiveEcosystemID,
		"ecosystem should NOT be cleared when switching domain")
}

// TestUseDomain_NoneClears verifies that dvm use domain none clears domain context.
func TestUseDomain_NoneClears(t *testing.T) {
	mock := db.NewMockDataStore()

	// Pre-set an active domain
	domID := 5
	mock.Context.ActiveDomainID = &domID

	useDomainCmd.SetContext(newCmdContextWithDS(mock))
	err := useDomainCmd.RunE(useDomainCmd, []string{"none"})

	require.NoError(t, err)
	assert.Nil(t, mock.Context.ActiveDomainID,
		"domain context should be cleared with 'none'")
}

// TestUseDomain_NonExistentReturnsError verifies that a non-existent domain
// results in an error.
func TestUseDomain_NonExistentReturnsError(t *testing.T) {
	mock := db.NewMockDataStore()

	// Set active ecosystem
	ecoID := 1
	mock.Context.ActiveEcosystemID = &ecoID
	mock.Ecosystems["prod"] = &models.Ecosystem{ID: 1, Name: "prod"}
	// No domains seeded

	useDomainCmd.SetContext(newCmdContextWithDS(mock))
	err := useDomainCmd.RunE(useDomainCmd, []string{"does-not-exist"})

	assert.Error(t, err, "non-existent domain should return an error")
}

// =============================================================================
// TestUseCommandHierarchy_IncludesEcosystemAndDomain
// =============================================================================

// TestUseCommandHierarchy_IncludesEcosystemAndDomain extends the existing
// TestUseCommandHierarchy test to also check for the new subcommands.
func TestUseCommandHierarchy_IncludesEcosystemAndDomain(t *testing.T) {
	subcommands := useCmd.Commands()
	names := make([]string, 0, len(subcommands))
	for _, c := range subcommands {
		names = append(names, c.Name())
	}

	assert.Contains(t, names, "ecosystem", "use should have 'ecosystem' subcommand")
	assert.Contains(t, names, "domain", "use should have 'domain' subcommand")
	assert.Contains(t, names, "app", "use should still have 'app' subcommand")
	assert.Contains(t, names, "workspace", "use should still have 'workspace' subcommand")
}

// =============================================================================
// TestUseExportFlag — --export flag
// =============================================================================

// TestUseExportFlagRegistered_App verifies that --export flag is registered on useAppCmd.
func TestUseExportFlagRegistered_App(t *testing.T) {
	flag := useAppCmd.Flags().Lookup("export")
	require.NotNil(t, flag, "useAppCmd should have --export flag")
	assert.Equal(t, "false", flag.DefValue, "--export flag should default to false")
	assert.Equal(t, "bool", flag.Value.Type(), "--export flag should be bool type")
}

// TestUseExportFlagRegistered_Workspace verifies that --export flag is registered on useWorkspaceCmd.
func TestUseExportFlagRegistered_Workspace(t *testing.T) {
	flag := useWorkspaceCmd.Flags().Lookup("export")
	require.NotNil(t, flag, "useWorkspaceCmd should have --export flag")
	assert.Equal(t, "false", flag.DefValue, "--export flag should default to false")
	assert.Equal(t, "bool", flag.Value.Type(), "--export flag should be bool type")
}

// TestUseExportFlagRegistered_Ecosystem verifies that --export flag is registered on useEcosystemCmd.
func TestUseExportFlagRegistered_Ecosystem(t *testing.T) {
	flag := useEcosystemCmd.Flags().Lookup("export")
	require.NotNil(t, flag, "useEcosystemCmd should have --export flag")
	assert.Equal(t, "false", flag.DefValue, "--export flag should default to false")
	assert.Equal(t, "bool", flag.Value.Type(), "--export flag should be bool type")
}

// TestUseExportFlagRegistered_Domain verifies that --export flag is registered on useDomainCmd.
func TestUseExportFlagRegistered_Domain(t *testing.T) {
	flag := useDomainCmd.Flags().Lookup("export")
	require.NotNil(t, flag, "useDomainCmd should have --export flag")
	assert.Equal(t, "false", flag.DefValue, "--export flag should default to false")
	assert.Equal(t, "bool", flag.Value.Type(), "--export flag should be bool type")
}

// TestUseExport_AppOutputsExportStatement verifies that dvm use app <name> --export
// prints "export DVM_APP=<name>" to stdout.
func TestUseExport_AppOutputsExportStatement(t *testing.T) {
	mock := db.NewMockDataStore()

	// Seed an app
	app := &models.App{ID: 1, Name: "myapi", DomainID: sql.NullInt64{Int64: 1, Valid: true}}
	mock.Apps[1] = app

	// Capture output by redirecting useAppCmd's Out
	var buf bytes.Buffer
	useAppCmd.SetOut(&buf)

	// Set --export flag
	useAppCmd.Flags().Set("export", "true")
	defer useAppCmd.Flags().Set("export", "false") // cleanup

	useAppCmd.SetContext(newCmdContextWithDS(mock))
	err := useAppCmd.RunE(useAppCmd, []string{"myapi"})

	// Should succeed
	require.NoError(t, err)

	// Output should contain the export statement
	output := buf.String()
	assert.True(t, strings.Contains(output, "export DVM_APP=myapi"),
		"--export flag should output 'export DVM_APP=myapi', got: %q", output)
}

// TestUseExport_WorkspaceOutputsExportStatement verifies that dvm use workspace <name> --export
// prints "export DVM_WORKSPACE=<name>" to stdout.
func TestUseExport_WorkspaceOutputsExportStatement(t *testing.T) {
	mock := db.NewMockDataStore()

	// Seed context: active app + workspace
	appID := 1
	mock.Context.ActiveAppID = &appID
	app := &models.App{ID: 1, Name: "myapi", DomainID: sql.NullInt64{Int64: 1, Valid: true}}
	mock.Apps[1] = app
	ws := &models.Workspace{ID: 5, Name: "dev", AppID: 1}
	mock.Workspaces[5] = ws

	var buf bytes.Buffer
	useWorkspaceCmd.SetOut(&buf)

	useWorkspaceCmd.Flags().Set("export", "true")
	defer useWorkspaceCmd.Flags().Set("export", "false")

	useWorkspaceCmd.SetContext(newCmdContextWithDS(mock))
	err := useWorkspaceCmd.RunE(useWorkspaceCmd, []string{"dev"})

	require.NoError(t, err)

	output := buf.String()
	assert.True(t, strings.Contains(output, "export DVM_WORKSPACE=dev"),
		"--export flag should output 'export DVM_WORKSPACE=dev', got: %q", output)
}

// TestUseExport_EcosystemOutputsExportStatement verifies that dvm use ecosystem <name> --export
// prints "export DVM_ECOSYSTEM=<name>" to stdout.
func TestUseExport_EcosystemOutputsExportStatement(t *testing.T) {
	mock := db.NewMockDataStore()

	eco := &models.Ecosystem{ID: 3, Name: "prod"}
	mock.Ecosystems["prod"] = eco

	var buf bytes.Buffer
	useEcosystemCmd.SetOut(&buf)

	useEcosystemCmd.Flags().Set("export", "true")
	defer useEcosystemCmd.Flags().Set("export", "false")

	useEcosystemCmd.SetContext(newCmdContextWithDS(mock))
	err := useEcosystemCmd.RunE(useEcosystemCmd, []string{"prod"})

	require.NoError(t, err)

	output := buf.String()
	assert.True(t, strings.Contains(output, "export DVM_ECOSYSTEM=prod"),
		"--export flag should output 'export DVM_ECOSYSTEM=prod', got: %q", output)
}

// TestUseExport_DomainOutputsExportStatement verifies that dvm use domain <name> --export
// prints "export DVM_DOMAIN=<name>" to stdout.
func TestUseExport_DomainOutputsExportStatement(t *testing.T) {
	mock := db.NewMockDataStore()

	// Set active ecosystem
	ecoID := 1
	mock.Context.ActiveEcosystemID = &ecoID
	mock.Ecosystems["prod"] = &models.Ecosystem{ID: 1, Name: "prod"}
	domain := &models.Domain{ID: 2, Name: "backend", EcosystemID: sql.NullInt64{Int64: 1, Valid: true}}
	mock.Domains[2] = domain

	var buf bytes.Buffer
	useDomainCmd.SetOut(&buf)

	useDomainCmd.Flags().Set("export", "true")
	defer useDomainCmd.Flags().Set("export", "false")

	useDomainCmd.SetContext(newCmdContextWithDS(mock))
	err := useDomainCmd.RunE(useDomainCmd, []string{"backend"})

	require.NoError(t, err)

	output := buf.String()
	assert.True(t, strings.Contains(output, "export DVM_DOMAIN=backend"),
		"--export flag should output 'export DVM_DOMAIN=backend', got: %q", output)
}

// =============================================================================
// TestUseCmdHelp — Documentation
// =============================================================================

// TestUseCmdHelpMentionsEcosystem verifies 'use --help' mentions ecosystem.
func TestUseCmdHelpMentionsEcosystem(t *testing.T) {
	buf := new(bytes.Buffer)
	useCmd.SetOut(buf)
	useCmd.Help()
	helpText := buf.String()

	assert.Contains(t, helpText, "ecosystem",
		"use help should mention 'ecosystem' subcommand")
}

// TestUseCmdHelpMentionsDomain verifies 'use --help' mentions domain.
func TestUseCmdHelpMentionsDomain(t *testing.T) {
	buf := new(bytes.Buffer)
	useCmd.SetOut(buf)
	useCmd.Help()
	helpText := buf.String()

	assert.Contains(t, helpText, "domain",
		"use help should mention 'domain' subcommand")
}

// TestUseCmdLongMentionsEnvVars verifies 'use' long description documents env vars.
func TestUseCmdLongMentionsEnvVars(t *testing.T) {
	long := useCmd.Long
	assert.Contains(t, long, "DVM_ECOSYSTEM",
		"use Long description should document DVM_ECOSYSTEM env var")
	assert.Contains(t, long, "DVM_DOMAIN",
		"use Long description should document DVM_DOMAIN env var")
	assert.Contains(t, long, "DVM_APP",
		"use Long description should document DVM_APP env var")
	assert.Contains(t, long, "DVM_WORKSPACE",
		"use Long description should document DVM_WORKSPACE env var")
}
