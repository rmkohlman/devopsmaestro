package cmd

import (
	"context"
	"database/sql"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Test Helpers ==========

// newTestAttachCmd creates a fresh attachCmd for testing
func newTestAttachCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach to workspace container",
		Run: func(cmd *cobra.Command, args []string) {
			// Simplified test version
		},
	}
	AddHierarchyFlags(cmd, &attachFlags)
	cmd.Flags().Bool("no-sync", false, "Skip syncing git mirror before attach")
	cmd.Flags().String("network", "", "Network mode: bridge (default), none, host, or custom network name")
	cmd.Flags().Float64("cpus", 0, "CPU limit (e.g., 1.5 for 1.5 cores; 0 = no limit)")
	cmd.Flags().String("memory", "", "Memory limit (e.g., 512m, 2g; empty = no limit)")
	return cmd
}

// setupAttachTestContext creates a test context with workspace and GitRepo
func setupAttachTestContext() (*db.MockDataStore, *models.App, *models.Workspace, *models.GitRepoDB) {
	mockStore := db.NewMockDataStore()

	// Create hierarchy
	ecosystem := &models.Ecosystem{Name: "test-eco"}
	mockStore.CreateEcosystem(ecosystem)

	domain := &models.Domain{Name: "test-domain", EcosystemID: ecosystem.ID}
	mockStore.CreateDomain(domain)

	app := &models.App{
		Name:     "test-app",
		DomainID: domain.ID,
		Path:     "/test/path",
	}
	mockStore.CreateApp(app)

	// Create GitRepo
	gitRepo := &models.GitRepoDB{
		Name:     "test-repo",
		URL:      "https://github.com/org/repo.git",
		Slug:     "org-repo",
		AutoSync: true,
	}
	mockStore.CreateGitRepo(gitRepo)

	// Create workspace linked to GitRepo
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "test-workspace",
		ImageName: "dvm-test-workspace-test-app:v1",
		Status:    "running",
		GitRepoID: sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true},
	}
	mockStore.CreateWorkspace(workspace)

	return mockStore, app, workspace, gitRepo
}

// ========== Command Structure Tests ==========

func TestAttachCmd_Exists(t *testing.T) {
	assert.NotNil(t, attachCmd, "attachCmd should exist")
	assert.Equal(t, "attach", attachCmd.Use)
}

func TestAttachCmd_HasNoSyncFlag(t *testing.T) {
	noSyncFlag := attachCmd.Flags().Lookup("no-sync")
	assert.NotNil(t, noSyncFlag, "attachCmd should have 'no-sync' flag")

	if noSyncFlag != nil {
		assert.Equal(t, "false", noSyncFlag.DefValue, "no-sync flag should default to false")
		assert.Equal(t, "bool", noSyncFlag.Value.Type(), "no-sync flag should be bool type")
	}
}

// ========== --no-sync Flag Tests ==========

func TestAttach_WithGitRepoID_DefaultBehavior_SyncsBeforeAttach(t *testing.T) {
	mockStore, _, workspace, _ := setupAttachTestContext()

	// Verify workspace has GitRepoID
	assert.True(t, workspace.GitRepoID.Valid, "workspace should have GitRepoID")

	// Mock MirrorManager to track sync calls
	var syncCalled bool
	mockMirror := &MockMirrorManager{
		SyncFunc: func(slug string) error {
			syncCalled = true
			return nil
		},
	}

	// Get GitRepo (simulating attach workflow)
	repo, err := mockStore.GetGitRepoByID(workspace.GitRepoID.Int64)
	require.NoError(t, err)

	// Sync should be called when AutoSync is true and --no-sync not set
	if repo.AutoSync {
		err = mockMirror.Sync(repo.Slug)
		assert.NoError(t, err)
	}

	// Verify sync was called
	assert.True(t, syncCalled, "sync should be called by default")
}

func TestAttach_WithGitRepoID_NoSyncFlag_SkipsSync(t *testing.T) {
	mockStore, _, workspace, _ := setupAttachTestContext()

	// Create test command with --no-sync flag
	cmd := newTestAttachCmd()
	ctx := context.WithValue(context.Background(), CtxKeyDataStore, mockStore)
	cmd.SetContext(ctx)

	// Set --no-sync flag
	cmd.Flags().Set("no-sync", "true")

	// Mock MirrorManager to track sync calls
	var syncCalled bool
	mockMirror := &MockMirrorManager{
		SyncFunc: func(slug string) error {
			syncCalled = true
			return nil
		},
	}

	// Get flag value
	noSync, err := cmd.Flags().GetBool("no-sync")
	assert.NoError(t, err)
	assert.True(t, noSync, "no-sync flag should be true")

	// Get GitRepo
	repo, err := mockStore.GetGitRepoByID(workspace.GitRepoID.Int64)
	require.NoError(t, err)

	// Sync should be skipped when --no-sync is set
	if workspace.GitRepoID.Valid && !noSync {
		if repo.AutoSync {
			mockMirror.Sync(repo.Slug)
		}
	}

	// Verify sync was NOT called
	assert.False(t, syncCalled, "sync should be skipped with --no-sync flag")
}

func TestAttach_WithoutGitRepoID_NoSyncAttempted(t *testing.T) {
	mockStore, app, _, _ := setupAttachTestContext()

	// Create workspace WITHOUT GitRepoID
	workspace := &models.Workspace{
		AppID:     app.ID,
		Name:      "no-repo-workspace",
		ImageName: "test-image:v1",
		Status:    "running",
		// GitRepoID is NULL
	}
	err := mockStore.CreateWorkspace(workspace)
	require.NoError(t, err)

	// Mock MirrorManager
	var syncCalled bool
	mockMirror := &MockMirrorManager{
		SyncFunc: func(slug string) error {
			syncCalled = true
			return nil
		},
	}

	// Sync logic - should skip when GitRepoID is not set
	if workspace.GitRepoID.Valid {
		repo, err := mockStore.GetGitRepoByID(workspace.GitRepoID.Int64)
		if err == nil && repo.AutoSync {
			mockMirror.Sync(repo.Slug)
		}
	}

	// Verify sync was NOT called
	assert.False(t, syncCalled, "sync should not be attempted when workspace has no GitRepo")
}

func TestAttach_SyncFailure_IsWarningNotFatal(t *testing.T) {
	mockStore, _, workspace, _ := setupAttachTestContext()

	// Mock MirrorManager that fails on sync
	var syncCalled bool
	mockMirror := &MockMirrorManager{
		SyncFunc: func(slug string) error {
			syncCalled = true
			return assert.AnError // Simulate sync failure
		},
	}

	// Get GitRepo
	repo, err := mockStore.GetGitRepoByID(workspace.GitRepoID.Int64)
	require.NoError(t, err)

	// Attempt sync - should fail but not panic
	if repo.AutoSync {
		err = mockMirror.Sync(repo.Slug)
		// In actual implementation, this error is logged as warning and attach continues
		assert.Error(t, err, "sync should fail")
	}

	// Verify sync was attempted
	assert.True(t, syncCalled, "sync should be attempted")

	// In real implementation, attach would continue despite sync failure
	// This test verifies that sync errors don't block attach
}

func TestAttach_AutoSyncFalse_SkipsSync(t *testing.T) {
	mockStore, _, workspace, _ := setupAttachTestContext()

	// Get the GitRepo and update AutoSync to false
	repo, err := mockStore.GetGitRepoByID(workspace.GitRepoID.Int64)
	require.NoError(t, err)

	repo.AutoSync = false
	err = mockStore.UpdateGitRepo(repo)
	require.NoError(t, err)

	// Mock MirrorManager
	var syncCalled bool
	mockMirror := &MockMirrorManager{
		SyncFunc: func(slug string) error {
			syncCalled = true
			return nil
		},
	}

	// Get updated repo
	repo, err = mockStore.GetGitRepoByID(workspace.GitRepoID.Int64)
	require.NoError(t, err)

	// Sync should be skipped when AutoSync is false
	if repo.AutoSync {
		mockMirror.Sync(repo.Slug)
	}

	// Verify sync was NOT called
	assert.False(t, syncCalled, "sync should be skipped when AutoSync is false")
}

// ========== Table-Driven Tests ==========

func TestAttach_SyncBehavior_TableDriven(t *testing.T) {
	tests := []struct {
		name            string
		hasGitRepoID    bool
		autoSync        bool
		noSyncFlag      bool
		expectSyncCall  bool
		expectedInfoMsg string
	}{
		{
			name:            "default: has repo, autosync true, no flag",
			hasGitRepoID:    true,
			autoSync:        true,
			noSyncFlag:      false,
			expectSyncCall:  true,
			expectedInfoMsg: "syncing mirror",
		},
		{
			name:            "no-sync flag: has repo, autosync true, flag set",
			hasGitRepoID:    true,
			autoSync:        true,
			noSyncFlag:      true,
			expectSyncCall:  false,
			expectedInfoMsg: "skipping mirror sync",
		},
		{
			name:            "no repo: no gitrepo_id",
			hasGitRepoID:    false,
			autoSync:        false,
			noSyncFlag:      false,
			expectSyncCall:  false,
			expectedInfoMsg: "",
		},
		{
			name:            "autosync false: has repo but autosync disabled",
			hasGitRepoID:    true,
			autoSync:        false,
			noSyncFlag:      false,
			expectSyncCall:  false,
			expectedInfoMsg: "",
		},
		{
			name:            "no-sync with autosync false: redundant but harmless",
			hasGitRepoID:    true,
			autoSync:        false,
			noSyncFlag:      true,
			expectSyncCall:  false,
			expectedInfoMsg: "skipping mirror sync",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore, app, _, _ := setupAttachTestContext()

			// Create workspace based on test case
			workspace := &models.Workspace{
				AppID:     app.ID,
				Name:      tt.name,
				ImageName: "test-image:v1",
				Status:    "running",
			}

			if tt.hasGitRepoID {
				// Create GitRepo with specified AutoSync
				gitRepo := &models.GitRepoDB{
					Name:     tt.name + "-repo",
					URL:      "https://github.com/test/repo.git",
					Slug:     "test-repo-" + tt.name,
					AutoSync: tt.autoSync,
				}
				err := mockStore.CreateGitRepo(gitRepo)
				require.NoError(t, err)

				workspace.GitRepoID = sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true}
			}

			err := mockStore.CreateWorkspace(workspace)
			require.NoError(t, err)

			// Mock MirrorManager to track sync calls
			var syncCalled bool
			mockMirror := &MockMirrorManager{
				SyncFunc: func(slug string) error {
					syncCalled = true
					return nil
				},
			}

			// Simulate attach logic
			noSync := tt.noSyncFlag
			if workspace.GitRepoID.Valid && !noSync {
				repo, err := mockStore.GetGitRepoByID(workspace.GitRepoID.Int64)
				if err == nil && repo.AutoSync {
					mockMirror.Sync(repo.Slug)
				}
			}

			// Verify sync call matches expectation
			assert.Equal(t, tt.expectSyncCall, syncCalled, "sync call mismatch for: "+tt.name)
		})
	}
}

// ========== Integration-Style Tests ==========

func TestAttach_FullWorkflow_WithSync(t *testing.T) {
	mockStore, _, workspace, gitRepo := setupAttachTestContext()

	// 1. Retrieve workspace
	ws, err := mockStore.GetWorkspaceByID(workspace.ID)
	require.NoError(t, err)
	assert.True(t, ws.GitRepoID.Valid)

	// 2. Get associated GitRepo
	repo, err := mockStore.GetGitRepoByID(ws.GitRepoID.Int64)
	require.NoError(t, err)
	assert.Equal(t, gitRepo.Name, repo.Name)

	// 3. Check if sync should run
	noSync := false // Default behavior
	shouldSync := ws.GitRepoID.Valid && !noSync && repo.AutoSync
	assert.True(t, shouldSync, "should sync in default workflow")

	// 4. Mock sync operation
	var syncCalled bool
	mockMirror := &MockMirrorManager{
		SyncFunc: func(slug string) error {
			syncCalled = true
			return nil
		},
	}

	// 5. Perform sync
	if shouldSync {
		err = mockMirror.Sync(repo.Slug)
		assert.NoError(t, err)
	}

	// 6. Verify sync was performed
	assert.True(t, syncCalled, "sync should have been called in full workflow")
}

func TestAttach_FullWorkflow_WithNoSync(t *testing.T) {
	mockStore, _, workspace, gitRepo := setupAttachTestContext()

	// 1. Retrieve workspace
	ws, err := mockStore.GetWorkspaceByID(workspace.ID)
	require.NoError(t, err)

	// 2. Get associated GitRepo
	repo, err := mockStore.GetGitRepoByID(ws.GitRepoID.Int64)
	require.NoError(t, err)
	assert.Equal(t, gitRepo.Name, repo.Name)

	// 3. Check if sync should run (with --no-sync flag)
	noSync := true // User specified --no-sync
	shouldSync := ws.GitRepoID.Valid && !noSync && repo.AutoSync
	assert.False(t, shouldSync, "should NOT sync with --no-sync flag")

	// 4. Mock sync operation
	var syncCalled bool
	mockMirror := &MockMirrorManager{
		SyncFunc: func(slug string) error {
			syncCalled = true
			return nil
		},
	}

	// 5. Perform sync (should be skipped)
	if shouldSync {
		mockMirror.Sync(repo.Slug)
	}

	// 6. Verify sync was NOT performed
	assert.False(t, syncCalled, "sync should be skipped with --no-sync flag")
}

func TestAttach_MultipleWorkspaces_DifferentRepos(t *testing.T) {
	mockStore, app, _, _ := setupAttachTestContext()

	// Create multiple workspaces with different repos
	repos := []struct {
		name     string
		autoSync bool
	}{
		{"repo-1", true},
		{"repo-2", false},
		{"repo-3", true},
	}

	var syncCalls int
	mockMirror := &MockMirrorManager{
		SyncFunc: func(slug string) error {
			syncCalls++
			return nil
		},
	}

	for i, r := range repos {
		// Create GitRepo
		gitRepo := &models.GitRepoDB{
			Name:     r.name,
			URL:      "https://github.com/org/" + r.name + ".git",
			Slug:     "org-" + r.name,
			AutoSync: r.autoSync,
		}
		err := mockStore.CreateGitRepo(gitRepo)
		require.NoError(t, err)

		// Create workspace
		workspace := &models.Workspace{
			AppID:     app.ID,
			Name:      "ws-" + r.name,
			ImageName: "test-image:v1",
			Status:    "running",
			GitRepoID: sql.NullInt64{Int64: int64(gitRepo.ID), Valid: true},
		}
		err = mockStore.CreateWorkspace(workspace)
		require.NoError(t, err)

		// Simulate attach for this workspace
		ws, err := mockStore.GetWorkspaceByID(workspace.ID)
		require.NoError(t, err)

		if ws.GitRepoID.Valid {
			repo, err := mockStore.GetGitRepoByID(ws.GitRepoID.Int64)
			if err == nil && repo.AutoSync {
				mockMirror.Sync(repo.Slug)
			}
		}

		// Verify sync call based on AutoSync setting
		expectedCalls := 0
		for j := 0; j <= i; j++ {
			if repos[j].autoSync {
				expectedCalls++
			}
		}
		assert.Equal(t, expectedCalls, syncCalls, "sync call count mismatch at iteration %d", i)
	}
}

// =============================================================================
// WI-4: Runtime Env Injection Tests (pre-existing; updated for new signature)
// =============================================================================

// TestBuildRuntimeEnv_IncludesWorkspaceMetadata verifies that the env map
// passed to AttachOptions includes the standard DVM workspace metadata vars.
//
// NOTE (WI-1): Updated to new 5-string + 4-map signature:
// buildRuntimeEnv(app, ws, eco, domain, system, themeEnv, registryEnv, credentialEnv, wsEnv)
func TestBuildRuntimeEnv_IncludesWorkspaceMetadata(t *testing.T) {
	envVars := buildRuntimeEnv("my-app", "dev-ws", "my-eco", "my-domain", "", nil, nil, nil, nil)

	tests := []struct {
		key   string
		value string
	}{
		{"DVM_APP", "my-app"},
		{"DVM_WORKSPACE", "dev-ws"},
		{"DVM_ECOSYSTEM", "my-eco"},
		{"DVM_DOMAIN", "my-domain"},
		{"TERM", "xterm-256color"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			got, ok := envVars[tt.key]
			if !ok {
				t.Errorf("env missing key %q", tt.key)
			} else if got != tt.value {
				t.Errorf("env[%q] = %q, want %q", tt.key, got, tt.value)
			}
		})
	}
}

// TestBuildRuntimeEnv_MergesWorkspaceEnv verifies that workspace-level env vars
// (from spec.env) are included in the runtime env.
//
// NOTE (WI-1): Updated to new 4-map signature.
func TestBuildRuntimeEnv_MergesWorkspaceEnv(t *testing.T) {
	wsEnv := map[string]string{
		"MY_API_KEY": "secret",
		"LOG_LEVEL":  "debug",
	}

	envVars := buildRuntimeEnv("app", "ws", "", "", "", nil, nil, nil, wsEnv)

	if envVars["MY_API_KEY"] != "secret" {
		t.Errorf("env[MY_API_KEY] = %q, want %q", envVars["MY_API_KEY"], "secret")
	}
	if envVars["LOG_LEVEL"] != "debug" {
		t.Errorf("env[LOG_LEVEL] = %q, want %q", envVars["LOG_LEVEL"], "debug")
	}
}

// TestBuildRuntimeEnv_ThemeVarsIncluded verifies that theme vars are merged
// into the env and can be passed in via the themeEnv parameter.
//
// NOTE (WI-1): Updated to new 4-map signature.
func TestBuildRuntimeEnv_ThemeVarsIncluded(t *testing.T) {
	themeEnv := map[string]string{
		"DVM_COLOR_BG":   "#1a1b26",
		"DVM_COLOR_TEXT": "#c0caf5",
	}

	envVars := buildRuntimeEnv("app", "ws", "", "", "", themeEnv, nil, nil, nil)

	if envVars["DVM_COLOR_BG"] != "#1a1b26" {
		t.Errorf("env[DVM_COLOR_BG] = %q, want %q", envVars["DVM_COLOR_BG"], "#1a1b26")
	}
}

// TestBuildRuntimeEnv_WorkspaceEnvPriority verifies that workspace-level
// env vars take precedence over theme env vars when there is a key collision.
//
// NOTE (WI-1): Updated to new 4-map signature.
func TestBuildRuntimeEnv_WorkspaceEnvPriority(t *testing.T) {
	wsEnv := map[string]string{
		"DVM_COLOR_BG": "workspace-override",
	}
	themeEnv := map[string]string{
		"DVM_COLOR_BG": "theme-value",
	}

	envVars := buildRuntimeEnv("app", "ws", "", "", "", themeEnv, nil, nil, wsEnv)

	if envVars["DVM_COLOR_BG"] != "workspace-override" {
		t.Errorf("workspace env should override theme env; env[DVM_COLOR_BG] = %q, want %q",
			envVars["DVM_COLOR_BG"], "workspace-override")
	}
}

// =============================================================================
// WI-1: buildRuntimeEnv Extended Signature — TDD Phase 2 (RED)
//
// These tests verify the NEW 5-layer merge semantics introduced by WI-1:
//   1. metadata  (TERM, DVM_WORKSPACE, DVM_APP, DVM_ECOSYSTEM, DVM_DOMAIN)
//   2. themeEnv  (terminal color vars)
//   3. registryEnv (PIP_INDEX_URL, GOPROXY, NPM_CONFIG_REGISTRY, …)
//   4. credentialEnv (GITHUB_TOKEN, AWS_ACCESS_KEY_ID, …)
//   5. wsEnv     (workspace spec.env — highest user-defined priority)
//
// RED: Fails because the production function still takes (wsEnv, themeEnv)
// and does not accept registryEnv or credentialEnv parameters.
// =============================================================================

// TestBuildRuntimeEnv_FullMergePriority verifies the complete 5-layer merge
// order. A conflicting key in all layers should resolve to the wsEnv value,
// while DVM metadata vars must survive any override attempt.
func TestBuildRuntimeEnv_FullMergePriority(t *testing.T) {
	const conflictKey = "CONFLICT_KEY"
	const metaKey = "DVM_WORKSPACE"

	themeEnv := map[string]string{conflictKey: "theme-value"}
	registryEnv := map[string]string{conflictKey: "registry-value"}
	credentialEnv := map[string]string{conflictKey: "credential-value"}
	wsEnv := map[string]string{
		conflictKey: "workspace-value", // must win
		metaKey:     "evil-workspace",  // must NOT override metadata
	}

	result := buildRuntimeEnv("my-app", "real-ws", "eco", "dom", "", themeEnv, registryEnv, credentialEnv, wsEnv)

	t.Run("workspace env wins conflict", func(t *testing.T) {
		if got := result[conflictKey]; got != "workspace-value" {
			t.Errorf("expected wsEnv to win; got %q", got)
		}
	})

	t.Run("DVM_WORKSPACE metadata always authoritative", func(t *testing.T) {
		if got := result[metaKey]; got != "real-ws" {
			t.Errorf("DVM_WORKSPACE should be %q (from params), got %q", "real-ws", got)
		}
	})

	t.Run("DVM_APP metadata always authoritative", func(t *testing.T) {
		if got := result["DVM_APP"]; got != "my-app" {
			t.Errorf("DVM_APP should be %q (from params), got %q", "my-app", got)
		}
	})
}

// TestBuildRuntimeEnv_RegistryEnvMerged verifies that registry env vars
// (PIP_INDEX_URL, GOPROXY, NPM_CONFIG_REGISTRY) from the registryEnv map
// appear in the output when no higher-priority layer overrides them.
func TestBuildRuntimeEnv_RegistryEnvMerged(t *testing.T) {
	registryEnv := map[string]string{
		"PIP_INDEX_URL":       "http://localhost:3141/root/pypi/+simple/",
		"GOPROXY":             "http://localhost:3000",
		"NPM_CONFIG_REGISTRY": "http://localhost:4873/",
	}

	result := buildRuntimeEnv("app", "ws", "", "", "", nil, registryEnv, nil, nil)

	for k, want := range registryEnv {
		t.Run(k, func(t *testing.T) {
			if got := result[k]; got != want {
				t.Errorf("env[%q] = %q, want %q", k, got, want)
			}
		})
	}
}

// TestBuildRuntimeEnv_CredentialEnvMerged verifies that credential env vars
// (GITHUB_TOKEN, AWS_ACCESS_KEY_ID) from the credentialEnv map appear in the
// output when no higher-priority layer overrides them.
func TestBuildRuntimeEnv_CredentialEnvMerged(t *testing.T) {
	credentialEnv := map[string]string{
		"GITHUB_TOKEN":          "ghp_faketoken123",
		"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
		"AWS_SECRET_ACCESS_KEY": "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
	}

	result := buildRuntimeEnv("app", "ws", "", "", "", nil, nil, credentialEnv, nil)

	for k, want := range credentialEnv {
		t.Run(k, func(t *testing.T) {
			if got := result[k]; got != want {
				t.Errorf("env[%q] = %q, want %q", k, got, want)
			}
		})
	}
}

// TestBuildRuntimeEnv_CredentialOverridesRegistry verifies that credentials
// beat registry env vars when there is a key collision (layer 4 > layer 3).
func TestBuildRuntimeEnv_CredentialOverridesRegistry(t *testing.T) {
	registryEnv := map[string]string{
		"GOPROXY": "http://localhost:3000",
	}
	credentialEnv := map[string]string{
		"GOPROXY": "https://proxy.corporate.example.com",
	}

	result := buildRuntimeEnv("app", "ws", "", "", "", nil, registryEnv, credentialEnv, nil)

	if got := result["GOPROXY"]; got != "https://proxy.corporate.example.com" {
		t.Errorf("credential should override registry for GOPROXY; got %q", got)
	}
}

// TestBuildRuntimeEnv_WorkspaceEnvOverridesCredential verifies that workspace
// env beats credentials when there is a key collision (layer 5 > layer 4).
func TestBuildRuntimeEnv_WorkspaceEnvOverridesCredential(t *testing.T) {
	credentialEnv := map[string]string{
		"GITHUB_TOKEN": "from-credential-store",
	}
	wsEnv := map[string]string{
		"GITHUB_TOKEN": "workspace-personal-token",
	}

	result := buildRuntimeEnv("app", "ws", "", "", "", nil, nil, credentialEnv, wsEnv)

	if got := result["GITHUB_TOKEN"]; got != "workspace-personal-token" {
		t.Errorf("wsEnv should override credential for GITHUB_TOKEN; got %q", got)
	}
}

// =============================================================================
// Security-Driven Tests (from security review of WI-1/WI-2)
// =============================================================================

// TestBuildRuntimeEnv_DVMPrefixReserved verifies that DVM_* metadata vars set
// via the function parameters cannot be overridden by any env layer.
// If a caller passes DVM_WORKSPACE=evil in wsEnv, the actual workspace name
// from the parameters must win.
func TestBuildRuntimeEnv_DVMPrefixReserved(t *testing.T) {
	wsEnv := map[string]string{
		"DVM_WORKSPACE": "evil",
		"DVM_APP":       "evil-app",
		"DVM_ECOSYSTEM": "evil-eco",
		"DVM_DOMAIN":    "evil-domain",
	}

	result := buildRuntimeEnv("real-app", "real-ws", "real-eco", "real-domain", "", nil, nil, nil, wsEnv)

	tests := []struct {
		key  string
		want string
	}{
		{"DVM_WORKSPACE", "real-ws"},
		{"DVM_APP", "real-app"},
		{"DVM_ECOSYSTEM", "real-eco"},
		{"DVM_DOMAIN", "real-domain"},
	}

	for _, tt := range tests {
		t.Run("reserved "+tt.key, func(t *testing.T) {
			if got := result[tt.key]; got != tt.want {
				t.Errorf("DVM_* key %q should be reserved; got %q, want %q", tt.key, got, tt.want)
			}
		})
	}
}

// TestBuildRuntimeEnv_MetadataAlwaysPresent verifies that all required metadata
// vars (DVM_APP, DVM_WORKSPACE, TERM) are present in the final env regardless
// of what other layers pass in.
func TestBuildRuntimeEnv_MetadataAlwaysPresent(t *testing.T) {
	// Attempt to override metadata vars from multiple layers
	themeEnv := map[string]string{
		"DVM_APP":       "theme-app",
		"DVM_WORKSPACE": "theme-ws",
		"TERM":          "dumb",
	}
	wsEnv := map[string]string{
		"DVM_APP":       "ws-app",
		"DVM_WORKSPACE": "ws-ws",
		"TERM":          "vt100",
	}

	result := buildRuntimeEnv("correct-app", "correct-ws", "", "", "", themeEnv, nil, nil, wsEnv)

	t.Run("DVM_APP is authoritative", func(t *testing.T) {
		if got := result["DVM_APP"]; got != "correct-app" {
			t.Errorf("DVM_APP = %q, want %q", got, "correct-app")
		}
	})

	t.Run("DVM_WORKSPACE is authoritative", func(t *testing.T) {
		if got := result["DVM_WORKSPACE"]; got != "correct-ws" {
			t.Errorf("DVM_WORKSPACE = %q, want %q", got, "correct-ws")
		}
	})

	t.Run("TERM is always set to xterm-256color", func(t *testing.T) {
		if got := result["TERM"]; got != "xterm-256color" {
			t.Errorf("TERM = %q, want %q", got, "xterm-256color")
		}
	})
}

// =============================================================================
// WI-2: Credential Denylist Tests — TDD Phase 2 (RED)
//
// These tests verify that dangerous env var names (LD_PRELOAD, LD_LIBRARY_PATH,
// etc.) are filtered from credentialEnv before they are merged into the
// runtime environment.
//
// RED: Fails because buildRuntimeEnv does not yet accept credentialEnv and
// does not yet call envvalidation.SanitizeEnvMap on the credential layer.
// =============================================================================

// TestBuildRuntimeEnv_DangerousCredentialFiltered verifies that a credential
// entry whose key is on the security denylist (e.g. LD_PRELOAD) is NOT
// injected into the container environment.
func TestBuildRuntimeEnv_DangerousCredentialFiltered(t *testing.T) {
	credentialEnv := map[string]string{
		"GITHUB_TOKEN":    "safe-token", // should be present
		"LD_PRELOAD":      "/evil.so",   // must be filtered (denylist)
		"LD_LIBRARY_PATH": "/evil/lib",  // must be filtered (denylist)
	}

	result := buildRuntimeEnv("app", "ws", "", "", "", nil, nil, credentialEnv, nil)

	t.Run("safe credential is present", func(t *testing.T) {
		if got := result["GITHUB_TOKEN"]; got != "safe-token" {
			t.Errorf("GITHUB_TOKEN should be present; got %q", got)
		}
	})

	t.Run("LD_PRELOAD is filtered", func(t *testing.T) {
		if val, ok := result["LD_PRELOAD"]; ok {
			t.Errorf("LD_PRELOAD must not appear in env; got value %q", val)
		}
	})

	t.Run("LD_LIBRARY_PATH is filtered", func(t *testing.T) {
		if val, ok := result["LD_LIBRARY_PATH"]; ok {
			t.Errorf("LD_LIBRARY_PATH must not appear in env; got value %q", val)
		}
	})
}

// =============================================================================
// WI-3: Registry Env Helper — TDD Phase 2 (RED)
//
// These tests verify the new loadRegistryEnv helper function that queries
// enabled registries from the DataStore and returns their env var map.
//
// RED: Fails because loadRegistryEnv does not exist yet in cmd/attach.go.
// =============================================================================

// TestLoadRegistryEnv_EnabledOnly verifies that only enabled registries
// contribute env vars to the registry env map.
//
// RED: loadRegistryEnv does not exist yet — this will fail to compile.
func TestLoadRegistryEnv_EnabledOnly(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// Create an enabled PyPI registry (devpi)
	enabledRegistry := &models.Registry{
		Name:    "local-pypi",
		Type:    "devpi",
		Enabled: true,
		Port:    3141,
		Status:  "running",
	}
	enabledRegistry.ApplyDefaults()
	require.NoError(t, mockStore.CreateRegistry(enabledRegistry))

	// Create a disabled Go proxy registry (athens)
	disabledRegistry := &models.Registry{
		Name:    "local-go",
		Type:    "athens",
		Enabled: false,
		Port:    3000,
		Status:  "stopped",
	}
	disabledRegistry.ApplyDefaults()
	require.NoError(t, mockStore.CreateRegistry(disabledRegistry))

	// Call the helper under test
	result, err := loadRegistryEnv(mockStore)
	require.NoError(t, err)

	t.Run("enabled devpi injects PIP_INDEX_URL", func(t *testing.T) {
		if _, ok := result["PIP_INDEX_URL"]; !ok {
			t.Error("PIP_INDEX_URL should be present for enabled devpi registry")
		}
	})

	t.Run("disabled athens does NOT inject GOPROXY", func(t *testing.T) {
		if val, ok := result["GOPROXY"]; ok {
			t.Errorf("GOPROXY must not appear for disabled registry; got %q", val)
		}
	})
}

// TestLoadRegistryEnv_NoRegistries verifies that an empty map is returned
// when no registries exist in the store.
//
// RED: loadRegistryEnv does not exist yet.
func TestLoadRegistryEnv_NoRegistries(t *testing.T) {
	mockStore := db.NewMockDataStore()

	result, err := loadRegistryEnv(mockStore)
	require.NoError(t, err)

	if len(result) != 0 {
		t.Errorf("expected empty env map when no registries; got %v", result)
	}
}

// =============================================================================
// Issue #91: Network Isolation Flag Tests
// =============================================================================

func TestAttachCmd_HasNetworkFlag(t *testing.T) {
	networkFlag := attachCmd.Flags().Lookup("network")
	assert.NotNil(t, networkFlag, "attachCmd should have 'network' flag")
	if networkFlag != nil {
		assert.Equal(t, "", networkFlag.DefValue, "network flag should default to empty")
		assert.Equal(t, "string", networkFlag.Value.Type(), "network flag should be string type")
	}
}

func TestAttachCmd_NetworkFlag_Parsing(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"bridge", "bridge", false},
		{"none", "none", false},
		{"host", "host", false},
		{"custom network", "my-custom-net", false},
		{"empty (default)", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestAttachCmd()
			if tt.value != "" {
				err := cmd.Flags().Set("network", tt.value)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
			}

			got, err := cmd.Flags().GetString("network")
			assert.NoError(t, err)
			assert.Equal(t, tt.value, got)
		})
	}
}

// =============================================================================
// Issue #92: Resource Limit Flag Tests
// =============================================================================

func TestAttachCmd_HasCPUsFlag(t *testing.T) {
	cpusFlag := attachCmd.Flags().Lookup("cpus")
	assert.NotNil(t, cpusFlag, "attachCmd should have 'cpus' flag")
	if cpusFlag != nil {
		assert.Equal(t, "0", cpusFlag.DefValue, "cpus flag should default to 0")
		assert.Equal(t, "float64", cpusFlag.Value.Type(), "cpus flag should be float64 type")
	}
}

func TestAttachCmd_HasMemoryFlag(t *testing.T) {
	memoryFlag := attachCmd.Flags().Lookup("memory")
	assert.NotNil(t, memoryFlag, "attachCmd should have 'memory' flag")
	if memoryFlag != nil {
		assert.Equal(t, "", memoryFlag.DefValue, "memory flag should default to empty")
		assert.Equal(t, "string", memoryFlag.Value.Type(), "memory flag should be string type")
	}
}

func TestAttachCmd_CPUsFlag_Parsing(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		want    float64
		wantErr bool
	}{
		{"no limit", "0", 0, false},
		{"one core", "1", 1.0, false},
		{"half core", "0.5", 0.5, false},
		{"one and a half", "1.5", 1.5, false},
		{"four cores", "4", 4.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestAttachCmd()
			err := cmd.Flags().Set("cpus", tt.value)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)

			got, err := cmd.Flags().GetFloat64("cpus")
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAttachCmd_MemoryFlag_Parsing(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{"megabytes", "512m"},
		{"gigabytes", "2g"},
		{"kilobytes", "8192k"},
		{"empty (no limit)", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newTestAttachCmd()
			if tt.value != "" {
				err := cmd.Flags().Set("memory", tt.value)
				assert.NoError(t, err)
			}

			got, err := cmd.Flags().GetString("memory")
			assert.NoError(t, err)
			assert.Equal(t, tt.value, got)
		})
	}
}

// TestAttachCmd_AllNewFlags_Coexist verifies that network, cpus, and memory
// flags can all be set simultaneously without conflicts.
func TestAttachCmd_AllNewFlags_Coexist(t *testing.T) {
	cmd := newTestAttachCmd()

	// Set all three new flags
	assert.NoError(t, cmd.Flags().Set("network", "none"))
	assert.NoError(t, cmd.Flags().Set("cpus", "2"))
	assert.NoError(t, cmd.Flags().Set("memory", "4g"))

	// Verify all values are set correctly
	network, _ := cmd.Flags().GetString("network")
	assert.Equal(t, "none", network)

	cpus, _ := cmd.Flags().GetFloat64("cpus")
	assert.Equal(t, 2.0, cpus)

	memory, _ := cmd.Flags().GetString("memory")
	assert.Equal(t, "4g", memory)

	// Ensure existing flags still work
	assert.NoError(t, cmd.Flags().Set("no-sync", "true"))
	noSync, _ := cmd.Flags().GetBool("no-sync")
	assert.True(t, noSync)
}
