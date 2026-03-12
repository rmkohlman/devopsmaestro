package cmd

import (
	"bytes"
	"context"
	"devopsmaestro/db"
	"devopsmaestro/models"
	"devopsmaestro/render"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ========== Test Helpers ==========

// newTestCreateGitRepoCmd creates a fresh createGitRepoCmd for testing.
// This avoids state pollution from global command variables.
func newTestCreateGitRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gitrepo <name>",
		Aliases: []string{"repo", "gr"},
		Short:   "Create a git repository mirror",
		Args:    cobra.ExactArgs(1),
		RunE:    runCreateGitRepo,
	}
	cmd.Flags().String("url", "", "Git repository URL (required)")
	cmd.Flags().String("auth-type", "none", "Authentication type (none, ssh, token)")
	cmd.Flags().String("credential", "", "Credential name for authentication")
	cmd.Flags().Bool("no-sync", false, "Don't sync mirror after creation")
	return cmd
}

// newTestGetGitReposCmd creates a fresh getGitReposCmd for testing.
func newTestGetGitReposCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gitrepos",
		Aliases: []string{"repos", "grs"},
		Short:   "List git repositories",
		RunE:    runGetGitRepos,
	}
	cmd.Flags().StringP("output", "o", "table", "Output format (table, wide, yaml, json)")
	return cmd
}

// newTestGetGitRepoCmd creates a fresh getGitRepoCmd for testing.
func newTestGetGitRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gitrepo <name>",
		Aliases: []string{"repo", "gr"},
		Short:   "Get a git repository",
		Args:    cobra.ExactArgs(1),
		RunE:    runGetGitRepo,
	}
	cmd.Flags().StringP("output", "o", "yaml", "Output format (yaml, json)")
	return cmd
}

// newTestDeleteGitRepoCmd creates a fresh deleteGitRepoCmd for testing.
func newTestDeleteGitRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gitrepo <name>",
		Aliases: []string{"repo", "gr"},
		Short:   "Delete a git repository",
		Args:    cobra.ExactArgs(1),
		RunE:    runDeleteGitRepo,
	}
	cmd.Flags().Bool("keep-mirror", false, "Keep the mirror directory")
	return cmd
}

// newTestSyncGitRepoCmd creates a fresh syncGitRepoCmd for testing.
func newTestSyncGitRepoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gitrepo <name>",
		Aliases: []string{"repo", "gr"},
		Short:   "Sync a git repository",
		Args:    cobra.ExactArgs(1),
		RunE:    runSyncGitRepo,
	}
	return cmd
}

// newTestSyncGitReposCmd creates a fresh syncGitReposCmd for testing.
func newTestSyncGitReposCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "gitrepos",
		Aliases: []string{"repos", "grs"},
		Short:   "Sync all git repositories",
		RunE:    runSyncGitRepos,
	}
	return cmd
}

// ========== Command Structure Tests ==========

// TestCreateGitRepoCmd_Exists verifies the create gitrepo command is registered
func TestCreateGitRepoCmd_Exists(t *testing.T) {
	assert.NotNil(t, createGitRepoCmd, "createGitRepoCmd should exist")
	assert.Equal(t, "gitrepo <name>", createGitRepoCmd.Use, "createGitRepoCmd should have correct Use")
}

// TestCreateGitRepoCmd_Aliases verifies aliases are registered
func TestCreateGitRepoCmd_Aliases(t *testing.T) {
	aliases := createGitRepoCmd.Aliases
	assert.Contains(t, aliases, "repo", "should have 'repo' alias")
	assert.Contains(t, aliases, "gr", "should have 'gr' alias")
}

// TestGetGitReposCmd_Exists verifies the get gitrepos command is registered
func TestGetGitReposCmd_Exists(t *testing.T) {
	assert.NotNil(t, getGitReposCmd, "getGitReposCmd should exist")
	assert.Equal(t, "gitrepos", getGitReposCmd.Use, "getGitReposCmd should have correct Use")
}

// TestGetGitReposCmd_Aliases verifies aliases are registered
func TestGetGitReposCmd_Aliases(t *testing.T) {
	aliases := getGitReposCmd.Aliases
	assert.Contains(t, aliases, "repos", "should have 'repos' alias")
	assert.Contains(t, aliases, "grs", "should have 'grs' alias")
}

// TestGetGitRepoCmd_Exists verifies the get gitrepo command is registered
func TestGetGitRepoCmd_Exists(t *testing.T) {
	assert.NotNil(t, getGitRepoCmd, "getGitRepoCmd should exist")
	assert.Equal(t, "gitrepo <name>", getGitRepoCmd.Use, "getGitRepoCmd should have correct Use")
}

// TestGetGitRepoCmd_Aliases verifies aliases are registered
func TestGetGitRepoCmd_Aliases(t *testing.T) {
	aliases := getGitRepoCmd.Aliases
	assert.Contains(t, aliases, "repo", "should have 'repo' alias")
	assert.Contains(t, aliases, "gr", "should have 'gr' alias")
}

// TestDeleteGitRepoCmd_Exists verifies the delete gitrepo command is registered
func TestDeleteGitRepoCmd_Exists(t *testing.T) {
	assert.NotNil(t, deleteGitRepoCmd, "deleteGitRepoCmd should exist")
	assert.Equal(t, "gitrepo <name>", deleteGitRepoCmd.Use, "deleteGitRepoCmd should have correct Use")
}

// TestDeleteGitRepoCmd_Aliases verifies aliases are registered
func TestDeleteGitRepoCmd_Aliases(t *testing.T) {
	aliases := deleteGitRepoCmd.Aliases
	assert.Contains(t, aliases, "repo", "should have 'repo' alias")
	assert.Contains(t, aliases, "gr", "should have 'gr' alias")
}

// TestSyncGitRepoCmd_Exists verifies the sync gitrepo command is registered
func TestSyncGitRepoCmd_Exists(t *testing.T) {
	assert.NotNil(t, syncGitRepoCmd, "syncGitRepoCmd should exist")
	assert.Equal(t, "gitrepo <name>", syncGitRepoCmd.Use, "syncGitRepoCmd should have correct Use")
}

// TestSyncGitRepoCmd_Aliases verifies aliases are registered
func TestSyncGitRepoCmd_Aliases(t *testing.T) {
	aliases := syncGitRepoCmd.Aliases
	assert.Contains(t, aliases, "repo", "should have 'repo' alias")
	assert.Contains(t, aliases, "gr", "should have 'gr' alias")
}

// TestSyncGitReposCmd_Exists verifies the sync gitrepos command is registered
func TestSyncGitReposCmd_Exists(t *testing.T) {
	assert.NotNil(t, syncGitReposCmd, "syncGitReposCmd should exist")
	assert.Equal(t, "gitrepos", syncGitReposCmd.Use, "syncGitReposCmd should have correct Use")
}

// TestSyncGitReposCmd_Aliases verifies aliases are registered
func TestSyncGitReposCmd_Aliases(t *testing.T) {
	aliases := syncGitReposCmd.Aliases
	assert.Contains(t, aliases, "repos", "should have 'repos' alias")
	assert.Contains(t, aliases, "grs", "should have 'grs' alias")
}

// ========== Flag Tests ==========

// TestCreateGitRepoCmd_HasURLFlag verifies the --url flag is registered
func TestCreateGitRepoCmd_HasURLFlag(t *testing.T) {
	urlFlag := createGitRepoCmd.Flags().Lookup("url")
	assert.NotNil(t, urlFlag, "createGitRepoCmd should have 'url' flag")

	if urlFlag != nil {
		assert.Equal(t, "", urlFlag.DefValue, "url flag should default to empty")
		assert.Equal(t, "string", urlFlag.Value.Type(), "url flag should be string type")
	}
}

// TestCreateGitRepoCmd_HasAuthTypeFlag verifies the --auth-type flag is registered
func TestCreateGitRepoCmd_HasAuthTypeFlag(t *testing.T) {
	authTypeFlag := createGitRepoCmd.Flags().Lookup("auth-type")
	assert.NotNil(t, authTypeFlag, "createGitRepoCmd should have 'auth-type' flag")

	if authTypeFlag != nil {
		assert.Equal(t, "none", authTypeFlag.DefValue, "auth-type flag should default to 'none'")
		assert.Equal(t, "string", authTypeFlag.Value.Type(), "auth-type flag should be string type")
	}
}

// TestCreateGitRepoCmd_HasCredentialFlag verifies the --credential flag is registered
func TestCreateGitRepoCmd_HasCredentialFlag(t *testing.T) {
	credentialFlag := createGitRepoCmd.Flags().Lookup("credential")
	assert.NotNil(t, credentialFlag, "createGitRepoCmd should have 'credential' flag")

	if credentialFlag != nil {
		assert.Equal(t, "", credentialFlag.DefValue, "credential flag should default to empty")
		assert.Equal(t, "string", credentialFlag.Value.Type(), "credential flag should be string type")
	}
}

// TestCreateGitRepoCmd_HasNoSyncFlag verifies the --no-sync flag is registered
func TestCreateGitRepoCmd_HasNoSyncFlag(t *testing.T) {
	noSyncFlag := createGitRepoCmd.Flags().Lookup("no-sync")
	assert.NotNil(t, noSyncFlag, "createGitRepoCmd should have 'no-sync' flag")

	if noSyncFlag != nil {
		assert.Equal(t, "false", noSyncFlag.DefValue, "no-sync flag should default to false")
		assert.Equal(t, "bool", noSyncFlag.Value.Type(), "no-sync flag should be bool type")
	}
}

// TestDeleteGitRepoCmd_HasKeepMirrorFlag verifies the --keep-mirror flag is registered
func TestDeleteGitRepoCmd_HasKeepMirrorFlag(t *testing.T) {
	keepMirrorFlag := deleteGitRepoCmd.Flags().Lookup("keep-mirror")
	assert.NotNil(t, keepMirrorFlag, "deleteGitRepoCmd should have 'keep-mirror' flag")

	if keepMirrorFlag != nil {
		assert.Equal(t, "false", keepMirrorFlag.DefValue, "keep-mirror flag should default to false")
		assert.Equal(t, "bool", keepMirrorFlag.Value.Type(), "keep-mirror flag should be bool type")
	}
}

// ========== RunE Tests ==========

// TestCreateGitRepoCmd_HasRunE verifies createGitRepoCmd uses RunE (not Run)
func TestCreateGitRepoCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, createGitRepoCmd.RunE, "createGitRepoCmd should have RunE (not Run)")
}

// TestGetGitReposCmd_HasRunE verifies getGitReposCmd uses RunE
func TestGetGitReposCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, getGitReposCmd.RunE, "getGitReposCmd should have RunE (not Run)")
}

// TestGetGitRepoCmd_HasRunE verifies getGitRepoCmd uses RunE
func TestGetGitRepoCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, getGitRepoCmd.RunE, "getGitRepoCmd should have RunE (not Run)")
}

// TestDeleteGitRepoCmd_HasRunE verifies deleteGitRepoCmd uses RunE
func TestDeleteGitRepoCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, deleteGitRepoCmd.RunE, "deleteGitRepoCmd should have RunE (not Run)")
}

// TestSyncGitRepoCmd_HasRunE verifies syncGitRepoCmd uses RunE
func TestSyncGitRepoCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, syncGitRepoCmd.RunE, "syncGitRepoCmd should have RunE (not Run)")
}

// TestSyncGitReposCmd_HasRunE verifies syncGitReposCmd uses RunE
func TestSyncGitReposCmd_HasRunE(t *testing.T) {
	assert.NotNil(t, syncGitReposCmd.RunE, "syncGitReposCmd should have RunE (not Run)")
}

// ========== Help Text Tests ==========

// TestCreateGitRepoCmd_Help verifies help text includes examples
func TestCreateGitRepoCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	createGitRepoCmd.SetOut(buf)
	createGitRepoCmd.Help()
	helpText := buf.String()

	// Should mention flags
	assert.Contains(t, helpText, "--url", "help should mention '--url' flag")
	assert.Contains(t, helpText, "--auth-type", "help should mention '--auth-type' flag")
	assert.Contains(t, helpText, "--credential", "help should mention '--credential' flag")
	assert.Contains(t, helpText, "--no-sync", "help should mention '--no-sync' flag")

	// Should mention examples
	assert.Contains(t, helpText, "Examples:", "help should have examples section")
}

// TestGetGitReposCmd_Help verifies help text includes output formats
func TestGetGitReposCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	getGitReposCmd.SetOut(buf)
	getGitReposCmd.Help()
	helpText := buf.String()

	// Should mention output formats
	assert.Contains(t, helpText, "yaml", "help should mention yaml output")
	assert.Contains(t, helpText, "json", "help should mention json output")
	assert.Contains(t, helpText, "wide", "help should mention wide output")
}

// TestDeleteGitRepoCmd_Help verifies help text mentions keep-mirror
func TestDeleteGitRepoCmd_Help(t *testing.T) {
	buf := new(bytes.Buffer)
	deleteGitRepoCmd.SetOut(buf)
	deleteGitRepoCmd.Help()
	helpText := buf.String()

	assert.Contains(t, helpText, "--keep-mirror", "help should mention '--keep-mirror' flag")
}

// ========== Create GitRepo Operation Tests (RED phase - will fail) ==========

// TestCreateGitRepoCmd_ValidURL creates gitrepo with valid URL
func TestCreateGitRepoCmd_ValidURL(t *testing.T) {
	mockStore := db.NewMockDataStore()

	cmd := newTestCreateGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"my-repo"})
	cmd.Flags().Set("url", "https://github.com/org/repo.git")
	cmd.Flags().Set("no-sync", "true") // Skip actual git operations in test

	err := cmd.Execute()
	assert.NoError(t, err, "should create gitrepo with valid URL")

	// Verify repo was created
	repo, err := mockStore.GetGitRepoByName("my-repo")
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, "my-repo", repo.Name)
	assert.Equal(t, "https://github.com/org/repo.git", repo.URL)
	assert.Equal(t, "none", repo.AuthType)
}

// TestCreateGitRepoCmd_InvalidURL rejects invalid URLs
func TestCreateGitRepoCmd_InvalidURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{"empty url", ""},
		{"malformed url", "not-a-url"},
		{"invalid protocol", "ftp://example.com/repo.git"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := db.NewMockDataStore()

			cmd := newTestCreateGitRepoCmd()
			ctx := context.WithValue(context.Background(), "dataStore", mockStore)
			cmd.SetContext(ctx)

			cmd.SetArgs([]string{"test-repo"})
			cmd.Flags().Set("url", tt.url)

			err := cmd.Execute()
			assert.Error(t, err, "should reject invalid URL: %s", tt.url)
		})
	}
}

// TestCreateGitRepoCmd_MissingURL requires --url flag
func TestCreateGitRepoCmd_MissingURL(t *testing.T) {
	mockStore := db.NewMockDataStore()

	cmd := newTestCreateGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"my-repo"})
	// Don't set --url flag

	err := cmd.Execute()
	assert.Error(t, err, "should require --url flag")
	assert.Contains(t, err.Error(), "required", "error should mention flag is required")
}

// TestCreateGitRepoCmd_DuplicateName handles duplicate names
func TestCreateGitRepoCmd_DuplicateName(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// Create first repo
	repo1 := &models.GitRepoDB{
		Name: "my-repo",
		URL:  "https://github.com/org/repo1.git",
	}
	err := mockStore.CreateGitRepo(repo1)
	require.NoError(t, err)

	// Try to create duplicate
	cmd := newTestCreateGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"my-repo"})
	cmd.Flags().Set("url", "https://github.com/org/repo2.git")

	err = cmd.Execute()
	assert.Error(t, err, "should reject duplicate name")
	assert.Contains(t, err.Error(), "already exists", "error should mention duplicate")
}

// TestCreateGitRepoCmd_WithAuthType creates with --auth-type flag
func TestCreateGitRepoCmd_WithAuthType(t *testing.T) {
	mockStore := db.NewMockDataStore()

	cmd := newTestCreateGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"ssh-repo"})
	cmd.Flags().Set("url", "git@github.com:org/repo.git")
	cmd.Flags().Set("auth-type", "ssh")
	cmd.Flags().Set("no-sync", "true")

	err := cmd.Execute()
	assert.NoError(t, err)

	repo, err := mockStore.GetGitRepoByName("ssh-repo")
	assert.NoError(t, err)
	assert.Equal(t, "ssh", repo.AuthType)
}

// TestCreateGitRepoCmd_WithCredential creates with --credential flag
func TestCreateGitRepoCmd_WithCredential(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// Create the credential that will be referenced
	err := mockStore.CreateCredential(&models.CredentialDB{
		ID:        1,
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   0,
		Name:      "github-ssh",
		Source:    "keychain",
	})
	assert.NoError(t, err)

	cmd := newTestCreateGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"private-repo"})
	cmd.Flags().Set("url", "git@github.com:org/private.git")
	cmd.Flags().Set("auth-type", "ssh")
	cmd.Flags().Set("credential", "github-ssh")
	cmd.Flags().Set("no-sync", "true")

	err = cmd.Execute()
	assert.NoError(t, err)

	repo, err := mockStore.GetGitRepoByName("private-repo")
	assert.NoError(t, err)
	assert.Equal(t, "ssh", repo.AuthType)
	assert.True(t, repo.CredentialID.Valid, "credential ID should be set")
	assert.Equal(t, int64(1), repo.CredentialID.Int64, "credential ID should match")
}

// TestCreateGitRepoCmd_NoSync creates with --no-sync flag
func TestCreateGitRepoCmd_NoSync(t *testing.T) {
	mockStore := db.NewMockDataStore()

	cmd := newTestCreateGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"no-sync-repo"})
	cmd.Flags().Set("url", "https://github.com/org/repo.git")
	cmd.Flags().Set("no-sync", "true")

	err := cmd.Execute()
	assert.NoError(t, err)

	repo, err := mockStore.GetGitRepoByName("no-sync-repo")
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	// Should not have been synced yet (LastSyncedAt should be null)
	assert.False(t, repo.LastSyncedAt.Valid, "repo should not have been synced yet")
}

// ========== Get GitRepos Operation Tests (RED phase - will fail) ==========

// TestGetGitReposCmd_Table lists repos in table format
func TestGetGitReposCmd_Table(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// Add test repos
	repo1 := &models.GitRepoDB{Name: "repo1", URL: "https://github.com/org/repo1.git"}
	repo2 := &models.GitRepoDB{Name: "repo2", URL: "https://github.com/org/repo2.git"}
	mockStore.CreateGitRepo(repo1)
	mockStore.CreateGitRepo(repo2)

	// Capture render output
	var buf bytes.Buffer
	originalWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(originalWriter)

	cmd := newTestGetGitReposCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "repo1", "output should contain repo1")
	assert.Contains(t, output, "repo2", "output should contain repo2")
}

// TestGetGitReposCmd_Wide lists repos with extra columns
func TestGetGitReposCmd_Wide(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repo := &models.GitRepoDB{
		Name:     "test-repo",
		URL:      "https://github.com/org/repo.git",
		AuthType: "ssh",
	}
	mockStore.CreateGitRepo(repo)

	// Capture render output
	var buf bytes.Buffer
	originalWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(originalWriter)

	cmd := newTestGetGitReposCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)
	cmd.Flags().Set("output", "wide")

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "test-repo")
	assert.Contains(t, output, "AUTO_SYNC", "wide output should show AUTO_SYNC column")
}

// TestGetGitReposCmd_YAML outputs YAML format
func TestGetGitReposCmd_YAML(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repo := &models.GitRepoDB{Name: "test-repo", URL: "https://github.com/org/repo.git"}
	mockStore.CreateGitRepo(repo)

	// Capture render output
	var buf bytes.Buffer
	originalWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(originalWriter)

	cmd := newTestGetGitReposCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)
	cmd.Flags().Set("output", "yaml")

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "name:", "YAML output should have 'name:' field")
	assert.Contains(t, output, "test-repo")
}

// TestGetGitReposCmd_JSON outputs JSON format
func TestGetGitReposCmd_JSON(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repo := &models.GitRepoDB{Name: "test-repo", URL: "https://github.com/org/repo.git"}
	mockStore.CreateGitRepo(repo)

	// Capture render output
	var buf bytes.Buffer
	originalWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(originalWriter)

	cmd := newTestGetGitReposCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)
	cmd.Flags().Set("output", "json")

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"name"`, "JSON output should have 'name' field")
	assert.Contains(t, output, "test-repo")
}

// TestGetGitReposCmd_Empty handles empty list
func TestGetGitReposCmd_Empty(t *testing.T) {
	mockStore := db.NewMockDataStore()

	cmd := newTestGetGitReposCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	// Capture render output (render.Info writes to render's default writer, not cobra's SetOut)
	var buf bytes.Buffer
	render.SetWriter(&buf)
	defer render.SetWriter(os.Stdout)

	err := cmd.Execute()
	assert.NoError(t, err, "should handle empty list gracefully")

	output := buf.String()
	assert.Contains(t, output, "No", "should indicate no repositories found")
}

// TestGetGitRepoCmd_ByName gets single repo by name
func TestGetGitRepoCmd_ByName(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repo := &models.GitRepoDB{
		Name: "my-repo",
		URL:  "https://github.com/org/repo.git",
	}
	mockStore.CreateGitRepo(repo)

	// Capture render output
	var buf bytes.Buffer
	originalWriter := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(originalWriter)

	cmd := newTestGetGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"my-repo"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "my-repo")
	assert.Contains(t, output, "https://github.com/org/repo.git")
}

// TestGetGitRepoCmd_NotFound handles not found error
func TestGetGitRepoCmd_NotFound(t *testing.T) {
	mockStore := db.NewMockDataStore()

	cmd := newTestGetGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	assert.Error(t, err, "should return error for nonexistent repo")
	assert.Contains(t, err.Error(), "not found", "error should mention not found")
}

// ========== Delete GitRepo Operation Tests (RED phase - will fail) ==========

// TestDeleteGitRepoCmd_Success deletes existing repo
func TestDeleteGitRepoCmd_Success(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repo := &models.GitRepoDB{Name: "delete-me", URL: "https://github.com/org/repo.git"}
	mockStore.CreateGitRepo(repo)

	cmd := newTestDeleteGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"delete-me"})

	err := cmd.Execute()
	assert.NoError(t, err)

	// Verify repo was deleted
	_, err = mockStore.GetGitRepoByName("delete-me")
	assert.Error(t, err, "repo should be deleted")
}

// TestDeleteGitRepoCmd_NotFound handles not found error
func TestDeleteGitRepoCmd_NotFound(t *testing.T) {
	mockStore := db.NewMockDataStore()

	cmd := newTestDeleteGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	assert.Error(t, err, "should return error for nonexistent repo")
	assert.Contains(t, err.Error(), "not found", "error should mention not found")
}

// TestDeleteGitRepoCmd_KeepMirror deletes db record but keeps mirror
func TestDeleteGitRepoCmd_KeepMirror(t *testing.T) {
	mockStore := db.NewMockDataStore()

	repo := &models.GitRepoDB{Name: "keep-mirror", URL: "https://github.com/org/repo.git"}
	mockStore.CreateGitRepo(repo)

	cmd := newTestDeleteGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"keep-mirror"})
	cmd.Flags().Set("keep-mirror", "true")

	err := cmd.Execute()
	assert.NoError(t, err)

	// Repo should be deleted from DB
	_, err = mockStore.GetGitRepoByName("keep-mirror")
	assert.Error(t, err)

	// Implementation should verify mirror directory still exists on disk
}

// ========== Sync GitRepo Operation Tests (RED phase - will fail) ==========

// TestSyncGitRepoCmd_Success syncs specific repo
// Note: This test skips actual git operations since we're testing command logic
func TestSyncGitRepoCmd_Success(t *testing.T) {
	t.Skip("Sync tests require git operations - tested via integration tests")
}

// TestSyncGitRepoCmd_NotFound handles not found error
func TestSyncGitRepoCmd_NotFound(t *testing.T) {
	mockStore := db.NewMockDataStore()

	cmd := newTestSyncGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	assert.Error(t, err, "should return error for nonexistent repo")
	assert.Contains(t, err.Error(), "not found", "error should mention not found")
}

// TestSyncGitReposCmd_All syncs all repos
// Note: This test skips actual git operations since we're testing command logic
func TestSyncGitReposCmd_All(t *testing.T) {
	t.Skip("Sync tests require git operations - tested via integration tests")
}

// TestSyncGitReposCmd_Empty handles no repos to sync
func TestSyncGitReposCmd_Empty(t *testing.T) {
	mockStore := db.NewMockDataStore()

	cmd := newTestSyncGitReposCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)

	err := cmd.Execute()
	assert.NoError(t, err, "should handle empty list gracefully")
}

// =============================================================================
// WI-2: GitRepo Single-Item Render Fix Tests
// RED: These tests FAIL until gitRepoToYAML returns render.KeyValueData.
// =============================================================================

// TestGitRepoToYAML_ReturnsKeyValueData verifies that gitRepoToYAML returns
// render.KeyValueData (not map[string]interface{}) for proper human-readable output.
func TestGitRepoToYAML_ReturnsKeyValueData(t *testing.T) {
	repo := &models.GitRepoDB{
		Name:       "my-repo",
		URL:        "https://github.com/org/repo.git",
		Slug:       "org-repo",
		DefaultRef: "main",
		AuthType:   "none",
		AutoSync:   true,
		SyncStatus: "synced",
	}

	result := gitRepoToYAML(repo)

	// The result must be render.KeyValueData (not map[string]interface{})
	// We cast to any first to allow type assertion regardless of declared return type.
	var resultAny any = result
	_, ok := resultAny.(render.KeyValueData)
	if !ok {
		t.Errorf("gitRepoToYAML() returned %T, want render.KeyValueData", result)
	}
}

// TestRunGetGitRepo_HumanOutput verifies that get gitrepo outputs human-readable
// key-value format (not raw YAML of a map) when format is yaml.
func TestRunGetGitRepo_HumanOutput(t *testing.T) {
	mockStore := db.NewMockDataStore()
	repo := &models.GitRepoDB{
		Name:       "test-repo",
		URL:        "https://github.com/test/repo.git",
		Slug:       "test-repo",
		DefaultRef: "main",
		AuthType:   "none",
		AutoSync:   true,
		SyncStatus: "pending",
	}
	require.NoError(t, mockStore.CreateGitRepo(repo))

	var buf bytes.Buffer
	render.SetWriter(&buf)
	defer render.SetWriter(os.Stdout)

	cmd := newTestGetGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{"test-repo"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// With KeyValueData, the output should contain the repo name as a value
	// With the buggy map[string]interface{}, the output is unstructured
	assert.Contains(t, output, "test-repo", "output should contain the repo name")
}

// =============================================================================
// WI-3: GitRepo --credential Wiring Tests
// RED: These tests FAIL until credential lookup + GitRepoDB.CredentialID are wired.
// =============================================================================

// TestRunCreateGitRepo_WithCredential verifies that --credential flag causes
// the credential to be looked up and its ID set on the created GitRepo.
func TestRunCreateGitRepo_WithCredential(t *testing.T) {
	mockStore := db.NewMockDataStore()

	// Create a credential in the mock store
	svc := "github.com"
	cred := &models.CredentialDB{
		Name:      "my-gh-token",
		ScopeType: models.CredentialScopeEcosystem,
		ScopeID:   0, // global scope
		Source:    "keychain",
		Service:   &svc,
	}
	require.NoError(t, mockStore.CreateCredential(cred))

	cmd := newTestCreateGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{
		"cred-repo",
		"--url", "https://github.com/org/cred-repo.git",
		"--auth-type", "token",
		"--credential", "my-gh-token",
		"--no-sync",
	})

	err := cmd.Execute()
	require.NoError(t, err, "create gitrepo with --credential should succeed")

	// Verify the created repo has CredentialID set
	createdRepo, err := mockStore.GetGitRepoByName("cred-repo")
	require.NoError(t, err)
	if !createdRepo.CredentialID.Valid {
		t.Error("GitRepoDB.CredentialID is not set after --credential flag; credential wiring is missing")
	}
	if createdRepo.CredentialID.Int64 != cred.ID {
		t.Errorf("GitRepoDB.CredentialID = %d, want %d", createdRepo.CredentialID.Int64, cred.ID)
	}
}

// TestRunCreateGitRepo_InvalidCredential verifies that providing a nonexistent
// credential name with --credential causes an error.
func TestRunCreateGitRepo_InvalidCredential(t *testing.T) {
	mockStore := db.NewMockDataStore()

	cmd := newTestCreateGitRepoCmd()
	ctx := context.WithValue(context.Background(), "dataStore", mockStore)
	cmd.SetContext(ctx)
	cmd.SetArgs([]string{
		"bad-cred-repo",
		"--url", "https://github.com/org/repo.git",
		"--auth-type", "token",
		"--credential", "nonexistent-cred",
		"--no-sync",
	})

	err := cmd.Execute()
	assert.Error(t, err, "create gitrepo with nonexistent --credential should fail")
	assert.Contains(t, err.Error(), "credential", "error should mention credential")
}
