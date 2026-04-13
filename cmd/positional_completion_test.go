package cmd

// positional_completion_test.go — Tests for ValidArgsFunction on positional resource args
//
// Issue: #188
// These tests verify that commands accepting resource names as positional arguments
// have ValidArgsFunction registered so tab completion works.
//
// FAILING TESTS (gap in registerAllResourceCompletions):
//   - attachCmd  — no ValidArgsFunction registered for workspace positional arg
//   - buildCmd   — no ValidArgsFunction registered for workspace positional arg
//   - detachCmd  — no ValidArgsFunction registered for workspace positional arg
//
// PASSING TESTS (already registered):
//   - useEcosystemCmd — registered in registerAllResourceCompletions()
//   - useDomainCmd    — registered in registerAllResourceCompletions()
//   - useAppCmd       — registered in use.go init()
//   - useWorkspaceCmd — registered in use.go init()

import (
	"context"
	"database/sql"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/pkg/resource/handlers"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// TestPositionalCompletion_WorkspaceCommands
//
// Verifies that commands accepting a workspace name as a positional argument
// have ValidArgsFunction registered. These commands should complete workspace
// names when the user presses TAB after the command name.
//
// All three sub-tests currently FAIL because registerAllResourceCompletions()
// does not register ValidArgsFunction for attachCmd, buildCmd, or detachCmd.
// ---------------------------------------------------------------------------

func TestPositionalCompletion_WorkspaceCommands(t *testing.T) {
	type workspaceCmd struct {
		name string
		cmd  *cobra.Command
	}

	workspaceCmds := []workspaceCmd{
		{name: "attachCmd", cmd: attachCmd},
		{name: "buildCmd", cmd: buildCmd},
		{name: "detachCmd", cmd: detachCmd},
	}

	for _, tc := range workspaceCmds {
		tc := tc // capture
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.cmd, "%s must not be nil", tc.name)

			// ValidArgsFunction should be registered to provide workspace completions
			assert.NotNil(t, tc.cmd.ValidArgsFunction,
				"%s: ValidArgsFunction is nil — positional workspace completion not registered (gap in registerAllResourceCompletions). "+
					"Fix: add '%s.ValidArgsFunction = completeWorkspaces' to registerAllResourceCompletions()",
				tc.name, tc.name)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPositionalCompletion_UseSubcommands
//
// Verifies that 'dvm use <resource>' subcommands have ValidArgsFunction
// registered so TAB completes the resource name.
//
// These should all PASS — they are already wired up:
//   - useEcosystemCmd via registerAllResourceCompletions()
//   - useDomainCmd    via registerAllResourceCompletions()
//   - useAppCmd       via use.go init()
//   - useWorkspaceCmd via use.go init()
// ---------------------------------------------------------------------------

func TestPositionalCompletion_UseSubcommands(t *testing.T) {
	type useCmd struct {
		name string
		cmd  *cobra.Command
	}

	useCmds := []useCmd{
		{name: "useEcosystemCmd", cmd: useEcosystemCmd},
		{name: "useDomainCmd", cmd: useDomainCmd},
		{name: "useAppCmd", cmd: useAppCmd},
		{name: "useWorkspaceCmd", cmd: useWorkspaceCmd},
	}

	for _, tc := range useCmds {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.cmd, "%s must not be nil", tc.name)

			assert.NotNil(t, tc.cmd.ValidArgsFunction,
				"%s: ValidArgsFunction is nil — positional resource completion not registered",
				tc.name)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPositionalCompletion_ReturnValue_WorkspaceCommands
//
// Verifies that when ValidArgsFunction IS registered on attach/build/detach,
// calling it returns ShellCompDirectiveNoFileComp (the correct directive for
// resource name completion — no filename fallback).
//
// These tests FAIL because ValidArgsFunction is nil on those commands,
// so they prove the gap before and verify behavior after implementation.
// ---------------------------------------------------------------------------

func TestPositionalCompletion_ReturnValue_WorkspaceCommands(t *testing.T) {
	// Ensure handlers are registered for resource listing
	handlers.RegisterAll()

	// Create a test datastore with a known workspace
	dataStore := createTestDataStore(t)
	defer dataStore.Close()

	ecosystem := &models.Ecosystem{
		Name:        "test-eco",
		Description: sql.NullString{String: "Test ecosystem", Valid: true},
	}
	require.NoError(t, dataStore.CreateEcosystem(ecosystem))

	domain := &models.Domain{
		Name:        "test-domain",
		EcosystemID: sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true},
		Description: sql.NullString{String: "Test domain", Valid: true},
	}
	require.NoError(t, dataStore.CreateDomain(domain))

	app := &models.App{
		Name:        "test-app",
		Path:        "/path/to/app",
		DomainID: sql.NullInt64{Int64: int64(domain.ID), Valid: true},
		Description: sql.NullString{String: "Test app", Valid: true},
	}
	require.NoError(t, dataStore.CreateApp(app))

	workspace := &models.Workspace{
		Name:        "dev",
		AppID:       app.ID,
		Description: sql.NullString{String: "Development workspace", Valid: true},
	}
	require.NoError(t, dataStore.CreateWorkspace(workspace))

	// Create a command context with the test datastore
	testCmd := &cobra.Command{Use: "test"}
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, dataStore)
	testCmd.SetContext(ctx)

	type testCase struct {
		name string
		cmd  *cobra.Command
	}

	tests := []testCase{
		{name: "attachCmd", cmd: attachCmd},
		{name: "buildCmd", cmd: buildCmd},
		{name: "detachCmd", cmd: detachCmd},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name+" returns NoFileComp directive", func(t *testing.T) {
			require.NotNil(t, tc.cmd, "%s must not be nil", tc.name)

			// This test fails if ValidArgsFunction is nil
			if tc.cmd.ValidArgsFunction == nil {
				t.Errorf("%s: ValidArgsFunction is nil — cannot test completion behavior. "+
					"Register 'ValidArgsFunction = completeWorkspaces' to fix.",
					tc.name)
				return
			}

			// Set the test datastore context on the actual command
			tc.cmd.SetContext(ctx)
			defer tc.cmd.SetContext(context.Background())

			completions, directive := tc.cmd.ValidArgsFunction(tc.cmd, []string{}, "")

			// Must return NoFileComp (workspace names, not files)
			assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive,
				"%s: completion should return ShellCompDirectiveNoFileComp", tc.name)

			// Must return a non-nil slice
			assert.NotNil(t, completions,
				"%s: completion should return non-nil slice", tc.name)
		})

		t.Run(tc.name+" completes workspace names", func(t *testing.T) {
			require.NotNil(t, tc.cmd, "%s must not be nil", tc.name)

			if tc.cmd.ValidArgsFunction == nil {
				t.Errorf("%s: ValidArgsFunction is nil — cannot test that workspace names appear in completions.",
					tc.name)
				return
			}

			tc.cmd.SetContext(ctx)
			defer tc.cmd.SetContext(context.Background())

			completions, _ := tc.cmd.ValidArgsFunction(tc.cmd, []string{}, "")

			// Should contain our test workspace "dev"
			found := false
			for _, c := range completions {
				if len(c) >= len("dev") && c[:len("dev")] == "dev" {
					found = true
					break
				}
			}
			assert.True(t, found,
				"%s: completion should include workspace name 'dev', got: %v",
				tc.name, completions)
		})

		t.Run(tc.name+" stops completing after first arg", func(t *testing.T) {
			require.NotNil(t, tc.cmd, "%s must not be nil", tc.name)

			if tc.cmd.ValidArgsFunction == nil {
				t.Skipf("%s: ValidArgsFunction is nil — skipping second-arg behavior test", tc.name)
				return
			}

			tc.cmd.SetContext(ctx)
			defer tc.cmd.SetContext(context.Background())

			// Once workspace is provided as first arg, no more positional completions
			completions, directive := tc.cmd.ValidArgsFunction(tc.cmd, []string{"dev"}, "")

			assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive,
				"%s: after first arg, directive should still be NoFileComp", tc.name)
			assert.Nil(t, completions,
				"%s: after first arg is provided, should return nil completions", tc.name)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPositionalCompletion_UseSubcommand_ReturnValues
//
// Verifies that use subcommands return the correct resource type from their
// ValidArgsFunction. These tests PASS because the use subcommands are
// already registered.
// ---------------------------------------------------------------------------

func TestPositionalCompletion_UseSubcommand_ReturnValues(t *testing.T) {
	handlers.RegisterAll()

	dataStore := createTestDataStore(t)
	defer dataStore.Close()

	// Seed test data: one of each hierarchy resource
	ecosystem := &models.Ecosystem{
		Name:        "my-platform",
		Description: sql.NullString{String: "Platform ecosystem", Valid: true},
	}
	require.NoError(t, dataStore.CreateEcosystem(ecosystem))

	domain := &models.Domain{
		Name:        "backend",
		EcosystemID: sql.NullInt64{Int64: int64(ecosystem.ID), Valid: true},
		Description: sql.NullString{String: "Backend domain", Valid: true},
	}
	require.NoError(t, dataStore.CreateDomain(domain))

	app := &models.App{
		Name:        "api-service",
		Path:        "/path/to/api",
		DomainID: sql.NullInt64{Int64: int64(domain.ID), Valid: true},
		Description: sql.NullString{String: "API service", Valid: true},
	}
	require.NoError(t, dataStore.CreateApp(app))

	workspace := &models.Workspace{
		Name:        "staging",
		AppID:       app.ID,
		Description: sql.NullString{String: "Staging workspace", Valid: true},
	}
	require.NoError(t, dataStore.CreateWorkspace(workspace))

	ctx := context.WithValue(context.Background(), CtxKeyDataStore, dataStore)

	tests := []struct {
		name         string
		cmd          *cobra.Command
		expectedName string // the resource name that should appear in completions
	}{
		{
			name:         "useEcosystemCmd completes ecosystem names",
			cmd:          useEcosystemCmd,
			expectedName: "my-platform",
		},
		{
			name:         "useDomainCmd completes domain names",
			cmd:          useDomainCmd,
			expectedName: "backend",
		},
		{
			name:         "useAppCmd completes app names",
			cmd:          useAppCmd,
			expectedName: "api-service",
		},
		{
			name:         "useWorkspaceCmd completes workspace names",
			cmd:          useWorkspaceCmd,
			expectedName: "staging",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.cmd, "command must not be nil")
			require.NotNil(t, tt.cmd.ValidArgsFunction,
				"ValidArgsFunction must not be nil for %s", tt.name)

			tt.cmd.SetContext(ctx)
			defer tt.cmd.SetContext(context.Background())

			completions, directive := tt.cmd.ValidArgsFunction(tt.cmd, []string{}, "")

			assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive,
				"should return ShellCompDirectiveNoFileComp")
			assert.NotNil(t, completions)

			// Verify the expected resource name appears in completions
			found := false
			for _, c := range completions {
				if len(c) >= len(tt.expectedName) && c[:len(tt.expectedName)] == tt.expectedName {
					found = true
					break
				}
			}
			assert.True(t, found,
				"completions should include %q, got: %v", tt.expectedName, completions)
		})
	}
}

// ---------------------------------------------------------------------------
// TestPositionalCompletion_AllCommandsAudit
//
// Audit test: documents which commands have and don't have ValidArgsFunction
// for positional completions. This serves as a living checklist for issue #188.
//
// Commands marked as "MUST HAVE" in the issue spec but currently MISSING
// will fail with a clear message indicating the gap.
// ---------------------------------------------------------------------------

func TestPositionalCompletion_AllCommandsAudit(t *testing.T) {
	type commandAudit struct {
		name     string
		cmd      *cobra.Command
		mustHave bool   // if true, test FAILS when ValidArgsFunction is nil
		reason   string // why it must have completion
	}

	audit := []commandAudit{
		// === Workspace positional arg (MUST HAVE per issue #188) ===
		{
			name:     "attachCmd",
			cmd:      attachCmd,
			mustHave: true,
			reason:   "dvm attach <workspace> must complete workspace names",
		},
		{
			name:     "buildCmd",
			cmd:      buildCmd,
			mustHave: true,
			reason:   "dvm build <workspace> must complete workspace names",
		},
		{
			name:     "detachCmd",
			cmd:      detachCmd,
			mustHave: true,
			reason:   "dvm detach <workspace> must complete workspace names",
		},

		// === use subcommands (already registered — should PASS) ===
		{
			name:     "useEcosystemCmd",
			cmd:      useEcosystemCmd,
			mustHave: true,
			reason:   "dvm use ecosystem <name> must complete ecosystem names",
		},
		{
			name:     "useDomainCmd",
			cmd:      useDomainCmd,
			mustHave: true,
			reason:   "dvm use domain <name> must complete domain names",
		},
		{
			name:     "useAppCmd",
			cmd:      useAppCmd,
			mustHave: true,
			reason:   "dvm use app <name> must complete app names",
		},
		{
			name:     "useWorkspaceCmd",
			cmd:      useWorkspaceCmd,
			mustHave: true,
			reason:   "dvm use workspace <name> must complete workspace names",
		},

		// === get subcommands (already registered — should PASS) ===
		{
			name:     "getEcosystemCmd",
			cmd:      getEcosystemCmd,
			mustHave: true,
			reason:   "dvm get ecosystem <name> must complete ecosystem names",
		},
		{
			name:     "getDomainCmd",
			cmd:      getDomainCmd,
			mustHave: true,
			reason:   "dvm get domain <name> must complete domain names",
		},
		{
			name:     "getAppCmd",
			cmd:      getAppCmd,
			mustHave: true,
			reason:   "dvm get app <name> must complete app names",
		},
		{
			name:     "getWorkspaceCmd",
			cmd:      getWorkspaceCmd,
			mustHave: true,
			reason:   "dvm get workspace <name> must complete workspace names",
		},
		{
			name:     "getCredentialCmd",
			cmd:      getCredentialCmd,
			mustHave: true,
			reason:   "dvm get credential <name> must complete credential names",
		},
		{
			name:     "getRegistryCmd",
			cmd:      getRegistryCmd,
			mustHave: true,
			reason:   "dvm get registry <name> must complete registry names",
		},
		{
			name:     "getGitRepoCmd",
			cmd:      getGitRepoCmd,
			mustHave: true,
			reason:   "dvm get gitrepo <name> must complete gitrepo names",
		},

		// === delete subcommands (already registered — should PASS) ===
		{
			name:     "deleteEcosystemCmd",
			cmd:      deleteEcosystemCmd,
			mustHave: true,
			reason:   "dvm delete ecosystem <name> must complete ecosystem names",
		},
		{
			name:     "deleteDomainCmd",
			cmd:      deleteDomainCmd,
			mustHave: true,
			reason:   "dvm delete domain <name> must complete domain names",
		},
		{
			name:     "deleteAppCmd",
			cmd:      deleteAppCmd,
			mustHave: true,
			reason:   "dvm delete app <name> must complete app names",
		},
		{
			name:     "deleteWorkspaceCmd",
			cmd:      deleteWorkspaceCmd,
			mustHave: true,
			reason:   "dvm delete workspace <name> must complete workspace names",
		},
		{
			name:     "deleteCredentialCmd",
			cmd:      deleteCredentialCmd,
			mustHave: true,
			reason:   "dvm delete credential <name> must complete credential names",
		},
		{
			name:     "deleteRegistryCmd",
			cmd:      deleteRegistryCmd,
			mustHave: true,
			reason:   "dvm delete registry <name> must complete registry names",
		},
		{
			name:     "deleteGitRepoCmd",
			cmd:      deleteGitRepoCmd,
			mustHave: true,
			reason:   "dvm delete gitrepo <name> must complete gitrepo names",
		},
	}

	for _, tc := range audit {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.cmd == nil {
				if tc.mustHave {
					t.Errorf("%s is nil — command var not found; cannot check ValidArgsFunction", tc.name)
				} else {
					t.Skipf("%s is nil — skipping optional audit", tc.name)
				}
				return
			}

			if tc.mustHave {
				assert.NotNil(t, tc.cmd.ValidArgsFunction,
					"MISSING COMPLETION: %s — %s (issue #188)", tc.name, tc.reason)
			} else {
				// Just report current state for informational purposes
				t.Logf("%s: ValidArgsFunction = %v", tc.name, tc.cmd.ValidArgsFunction != nil)
			}
		})
	}
}
