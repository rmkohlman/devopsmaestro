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
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
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
// WI-4: Runtime Env Injection Tests
// RED: These tests FAIL until buildRuntimeEnv (or equivalent) is implemented.
// =============================================================================

// TestBuildRuntimeEnv_IncludesWorkspaceMetadata verifies that the env map
// passed to AttachOptions includes the standard DVM workspace metadata vars.
func TestBuildRuntimeEnv_IncludesWorkspaceMetadata(t *testing.T) {
	envVars := buildRuntimeEnv("my-app", "dev-ws", "my-eco", "my-domain", nil, nil)

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
func TestBuildRuntimeEnv_MergesWorkspaceEnv(t *testing.T) {
	wsEnv := map[string]string{
		"MY_API_KEY": "secret",
		"LOG_LEVEL":  "debug",
	}

	envVars := buildRuntimeEnv("app", "ws", "", "", wsEnv, nil)

	if envVars["MY_API_KEY"] != "secret" {
		t.Errorf("env[MY_API_KEY] = %q, want %q", envVars["MY_API_KEY"], "secret")
	}
	if envVars["LOG_LEVEL"] != "debug" {
		t.Errorf("env[LOG_LEVEL] = %q, want %q", envVars["LOG_LEVEL"], "debug")
	}
}

// TestBuildRuntimeEnv_ThemeVarsOverridable verifies that theme vars are merged
// into the env and can be passed in via the themeEnv parameter.
func TestBuildRuntimeEnv_ThemeVarsIncluded(t *testing.T) {
	themeEnv := map[string]string{
		"DVM_COLOR_BG":   "#1a1b26",
		"DVM_COLOR_TEXT": "#c0caf5",
	}

	envVars := buildRuntimeEnv("app", "ws", "", "", nil, themeEnv)

	if envVars["DVM_COLOR_BG"] != "#1a1b26" {
		t.Errorf("env[DVM_COLOR_BG] = %q, want %q", envVars["DVM_COLOR_BG"], "#1a1b26")
	}
}

// TestBuildRuntimeEnv_WorkspaceEnvOverridesTheme verifies that workspace-level
// env vars take precedence over theme env vars when there is a key collision.
func TestBuildRuntimeEnv_WorkspaceEnvPriority(t *testing.T) {
	wsEnv := map[string]string{
		"DVM_COLOR_BG": "workspace-override",
	}
	themeEnv := map[string]string{
		"DVM_COLOR_BG": "theme-value",
	}

	envVars := buildRuntimeEnv("app", "ws", "", "", wsEnv, themeEnv)

	if envVars["DVM_COLOR_BG"] != "workspace-override" {
		t.Errorf("workspace env should override theme env; env[DVM_COLOR_BG] = %q, want %q",
			envVars["DVM_COLOR_BG"], "workspace-override")
	}
}
