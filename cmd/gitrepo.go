package cmd

import (
	"database/sql"
	"devopsmaestro/models"
	"devopsmaestro/pkg/mirror"
	"devopsmaestro/utils"
	"fmt"
	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"
	"time"

	"github.com/spf13/cobra"
)

// Dry-run flags for gitrepo commands
var (
	createGitRepoDryRun bool
	deleteGitRepoDryRun bool
)

// gitrepoCmd is a placeholder that is never used directly
// It exists to support help text and organization
var gitrepoCmd = &cobra.Command{
	Use:   "gitrepo",
	Short: "Manage git repositories",
}

// createGitRepoCmd creates a new git repository mirror
var createGitRepoCmd = &cobra.Command{
	Use:     "gitrepo <name>",
	Aliases: []string{"repo", "gr"},
	Short:   "Create a git repository mirror",
	Long: `Create a git repository mirror configuration.

A git repository mirror stores a bare clone for use with workspaces.
You can specify authentication via --auth-type and --credential flags.

Examples:
  # Create a public repository
  dvm create gitrepo my-repo --url https://github.com/org/repo.git
  
  # Create with SSH authentication
  dvm create gitrepo private-repo --url git@github.com:org/repo.git --auth-type ssh --credential github-ssh
  
  # Create without immediate sync
  dvm create gitrepo my-repo --url https://github.com/org/repo.git --no-sync
  
  # Short aliases
  dvm create repo my-repo --url https://github.com/org/repo.git
  dvm create gr my-repo --url https://github.com/org/repo.git`,
	Args: cobra.ExactArgs(1),
	RunE: runCreateGitRepo,
}

// getGitReposCmd lists git repositories
var getGitReposCmd = &cobra.Command{
	Use:     "gitrepos",
	Aliases: []string{"repos", "grs"},
	Short:   "List git repositories",
	Long: `List all git repository mirrors.

Output formats:
  table (default) - Human-readable table
  wide            - Table with additional columns
  yaml            - YAML output
  json            - JSON output

Examples:
  # List all repositories
  dvm get gitrepos
  
  # Wide output with extra columns
  dvm get gitrepos -o wide
  
  # YAML output
  dvm get gitrepos -o yaml
  
  # JSON output
  dvm get gitrepos -o json
  
  # Short aliases
  dvm get repos
  dvm get grs`,
	RunE: runGetGitRepos,
}

// getGitRepoCmd gets a single git repository
var getGitRepoCmd = &cobra.Command{
	Use:     "gitrepo <name>",
	Aliases: []string{"repo", "gr"},
	Short:   "Get a git repository",
	Long: `Get details for a specific git repository mirror.

Output formats:
  yaml (default) - YAML output
  json           - JSON output

Examples:
  # Get repository details
  dvm get gitrepo my-repo
  
  # JSON output
  dvm get gitrepo my-repo -o json
  
  # Short aliases
  dvm get repo my-repo
  dvm get gr my-repo`,
	Args: cobra.ExactArgs(1),
	RunE: runGetGitRepo,
}

// deleteGitRepoCmd deletes a git repository
var deleteGitRepoCmd = &cobra.Command{
	Use:     "gitrepo <name>",
	Aliases: []string{"repo", "gr"},
	Short:   "Delete a git repository",
	Long: `Delete a git repository mirror configuration.

By default, this removes both the database record and the mirror directory.
Use --keep-mirror to only remove the database record.
By default, you will be prompted for confirmation. Use --force to skip.

Examples:
  # Delete repository and mirror
  dvm delete gitrepo my-repo
  
  # Delete database record but keep mirror directory
  dvm delete gitrepo my-repo --keep-mirror
  
  # Skip confirmation
  dvm delete gitrepo my-repo --force
  
  # Short aliases
  dvm delete repo my-repo
  dvm delete gr my-repo`,
	Args: cobra.ExactArgs(1),
	RunE: runDeleteGitRepo,
}

// syncGitRepoCmd syncs a single repository
var syncGitRepoCmd = &cobra.Command{
	Use:     "gitrepo <name>",
	Aliases: []string{"repo", "gr"},
	Short:   "Sync a git repository",
	Long: `Sync a git repository mirror with its remote.

This fetches the latest changes from the remote repository.

Examples:
  # Sync a repository
  dvm sync gitrepo my-repo
  
  # Short aliases
  dvm sync repo my-repo
  dvm sync gr my-repo`,
	Args: cobra.ExactArgs(1),
	RunE: runSyncGitRepo,
}

// syncGitReposCmd syncs all repositories
var syncGitReposCmd = &cobra.Command{
	Use:     "gitrepos",
	Aliases: []string{"repos", "grs"},
	Short:   "Sync all git repositories",
	Long: `Sync all git repository mirrors with their remotes.

This fetches the latest changes from all remote repositories.

Examples:
  # Sync all repositories
  dvm sync gitrepos
  
  # Short aliases
  dvm sync repos
  dvm sync grs`,
	RunE: runSyncGitRepos,
}

// syncCmd is the parent command for sync operations
var syncCmd *cobra.Command

func init() {
	// Register create subcommand
	createCmd.AddCommand(createGitRepoCmd)
	createGitRepoCmd.Flags().String("url", "", "Git repository URL (required)")
	createGitRepoCmd.Flags().String("auth-type", "none", "Authentication type (none, ssh, token)")
	createGitRepoCmd.Flags().String("credential", "", "Credential name for authentication")
	createGitRepoCmd.Flags().Bool("no-sync", false, "Skip initial sync")
	createGitRepoCmd.MarkFlagRequired("url")
	createGitRepoCmd.Flags().String("default-ref", "", "Default branch name (auto-detected if not specified)")
	AddDryRunFlag(createGitRepoCmd, &createGitRepoDryRun)

	// Register get subcommands
	getCmd.AddCommand(getGitReposCmd)
	getCmd.AddCommand(getGitRepoCmd)

	// Register delete subcommand
	deleteCmd.AddCommand(deleteGitRepoCmd)
	deleteGitRepoCmd.Flags().Bool("keep-mirror", false, "Keep mirror directory on disk")
	AddForceConfirmFlag(deleteGitRepoCmd)
	AddDryRunFlag(deleteGitRepoCmd, &deleteGitRepoDryRun)

	// Create or get sync command
	idx := findCommandIndex(rootCmd, "sync")
	if idx >= 0 {
		syncCmd = rootCmd.Commands()[idx]
	} else {
		syncCmd = &cobra.Command{
			Use:   "sync",
			Short: "Sync resources",
			Long:  `Sync resources with their remote sources.`,
		}
		rootCmd.AddCommand(syncCmd)
	}

	// Register sync subcommands
	syncCmd.AddCommand(syncGitRepoCmd)
	syncCmd.AddCommand(syncGitReposCmd)
}

// findCommandIndex finds the index of a command by name
func findCommandIndex(parent *cobra.Command, name string) int {
	for i, cmd := range parent.Commands() {
		if cmd.Name() == name {
			return i
		}
	}
	return -1
}

// =============================================================================
// Command Implementations
// =============================================================================

// runCreateGitRepo implements the create gitrepo command
func runCreateGitRepo(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Validate name is not empty
	if err := ValidateResourceName(name, "gitrepo"); err != nil {
		return err
	}

	// Get URL from flag
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return fmt.Errorf("failed to get url flag: %w", err)
	}

	// Validate URL - required flag is handled by cobra, but validate content
	if url == "" {
		return fmt.Errorf("required flag \"url\" not set")
	}

	// Validate URL using mirror package
	if err := mirror.ValidateGitURL(url); err != nil {
		return fmt.Errorf("invalid git URL: %w", err)
	}

	// Generate slug from URL
	slug, err := mirror.GenerateSlug(url)
	if err != nil {
		return fmt.Errorf("failed to generate slug: %w", err)
	}

	// Get optional flags
	authType, _ := cmd.Flags().GetString("auth-type")
	credential, _ := cmd.Flags().GetString("credential")
	noSync, _ := cmd.Flags().GetBool("no-sync")

	// Determine default branch
	defaultRef, _ := cmd.Flags().GetString("default-ref")

	// Dry-run: preview what would be created
	if createGitRepoDryRun {
		render.Plain(fmt.Sprintf("Would create gitrepo %q", name))
		render.Plain(fmt.Sprintf("  url: %s", url))
		render.Plain(fmt.Sprintf("  auth-type: %s", authType))
		if credential != "" {
			render.Plain(fmt.Sprintf("  credential: %s", credential))
		}
		if noSync {
			render.Plain("  --no-sync: would skip initial sync")
		}
		return nil
	}

	if defaultRef == "" {
		render.Progress("Detecting default branch...")
		defaultRef = utils.DetectDefaultBranch(url)
		render.Info(fmt.Sprintf("Detected default branch: %s", defaultRef))
	}

	// Get dataStore from context
	dataStore, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Check if repo with same name already exists
	existingRepo, err := dataStore.GetGitRepoByName(name)
	if err == nil && existingRepo != nil {
		return fmt.Errorf("gitrepo '%s' already exists", name)
	}

	// Create the GitRepoDB model
	repo := &models.GitRepoDB{
		Name:                name,
		URL:                 url,
		Slug:                slug,
		DefaultRef:          defaultRef,
		AuthType:            authType,
		AutoSync:            true,
		SyncIntervalMinutes: 60,
		SyncStatus:          "pending",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Set credential ID if provided
	if credential != "" {
		// Validate auth-type/credential consistency
		if authType == "none" {
			return fmt.Errorf("cannot use --credential with --auth-type none")
		}

		// Look up credential by name
		cred, err := dataStore.GetCredentialByName(credential)
		if err != nil {
			return fmt.Errorf("credential '%s' not found: %w", credential, err)
		}
		repo.CredentialID = sql.NullInt64{Int64: cred.ID, Valid: true}
	}

	// Create the repo in the database
	if err := dataStore.CreateGitRepo(repo); err != nil {
		return fmt.Errorf("failed to create gitrepo: %w", err)
	}

	// Clone the mirror if not --no-sync
	if !noSync {
		baseDir := getGitRepoBaseDir()
		mirrorMgr := mirror.NewGitMirrorManager(baseDir)
		if _, err := mirrorMgr.Clone(url, slug); err != nil {
			// Update sync status to failed
			repo.SyncStatus = "failed"
			repo.SyncError = sql.NullString{String: err.Error(), Valid: true}
			dataStore.UpdateGitRepo(repo)
			// Don't fail the create, just warn
			render.Warning(fmt.Sprintf("Created gitrepo '%s' but initial sync failed: %v", name, err))
			return nil
		}

		// Update sync status
		repo.LastSyncedAt = sql.NullTime{Time: time.Now(), Valid: true}
		repo.SyncStatus = "synced"
		dataStore.UpdateGitRepo(repo)
	}

	render.Success(fmt.Sprintf("Created gitrepo '%s'", name))
	return nil
}

// runGetGitRepos implements the get gitrepos command
func runGetGitRepos(cmd *cobra.Command, args []string) error {
	// Get dataStore from context
	dataStore, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get output format
	format, _ := cmd.Flags().GetString("output")

	// List all repos
	repos, err := dataStore.ListGitRepos()
	if err != nil {
		return fmt.Errorf("failed to list gitrepos: %w", err)
	}

	// Handle empty list
	if len(repos) == 0 {
		render.Info("No git repositories found")
		return nil
	}

	// Handle YAML/JSON output
	if format == "yaml" || format == "json" {
		return render.OutputWith(format, gitReposToYAML(repos), render.Options{})
	}

	// Determine if wide format
	isWide := format == "wide"

	// Convert to pointer slice for BuildTable
	repoPtrs := make([]*models.GitRepoDB, len(repos))
	for i := range repos {
		repoPtrs[i] = &repos[i]
	}

	// Build table using shared builder (with constraints)
	tableData := BuildTable(&gitRepoTableBuilder{}, repoPtrs, isWide)

	return render.OutputWith(format, tableData, render.Options{
		Type: render.TypeTable,
	})
}

// runGetGitRepo implements the get gitrepo command
func runGetGitRepo(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Get dataStore from context
	dataStore, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get the repo
	repo, err := dataStore.GetGitRepoByName(name)
	if err != nil {
		return fmt.Errorf("gitrepo '%s' not found", name)
	}

	// Get output format
	format, _ := cmd.Flags().GetString("output")

	// JSON/YAML: use ToYAML for apiVersion/kind/metadata/spec format (issue #183)
	if format == "json" || format == "yaml" {
		// Resolve credential name if associated
		credentialName := ""
		if repo.CredentialID.Valid {
			creds, credErr := dataStore.ListAllCredentials()
			if credErr == nil {
				for _, c := range creds {
					if c.ID == repo.CredentialID.Int64 {
						credentialName = c.Name
						break
					}
				}
			}
		}
		yamlDoc := repo.ToYAML(credentialName)
		return render.OutputWith(format, yamlDoc, render.Options{})
	}

	// Human output: use KeyValueData
	return render.OutputWith(format, gitRepoToYAML(repo), render.Options{
		Type:  render.TypeKeyValue,
		Title: "GitRepo Details",
	})
}

// runDeleteGitRepo implements the delete gitrepo command
func runDeleteGitRepo(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Get dataStore from context
	dataStore, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get flags
	keepMirror, _ := cmd.Flags().GetBool("keep-mirror")
	force, _ := cmd.Flags().GetBool("force")

	// Get the repo first to get slug for mirror deletion and confirmation info
	repo, err := dataStore.GetGitRepoByName(name)
	if err != nil {
		return fmt.Errorf("gitrepo '%s' not found", name)
	}

	// Build confirmation message
	msg := fmt.Sprintf("Delete gitrepo '%s' (URL: %s)", name, repo.URL)
	if !keepMirror {
		msg += " and its mirror directory?"
	} else {
		msg += " (keeping mirror directory)?"
	}

	// Dry-run: preview what would be deleted
	if deleteGitRepoDryRun {
		render.Plain(fmt.Sprintf("Would delete gitrepo %q (URL: %s)", name, repo.URL))
		if !keepMirror {
			render.Plain("  Would also delete mirror directory")
		}
		return nil
	}

	confirmed, err := confirmDelete(msg, force)
	if err != nil {
		return err
	}
	if !confirmed {
		return nil
	}

	// Delete from database
	if err := dataStore.DeleteGitRepo(name); err != nil {
		return fmt.Errorf("failed to delete gitrepo: %w", err)
	}

	// Delete mirror directory if not keeping
	if !keepMirror && repo.Slug != "" {
		baseDir := getGitRepoBaseDir()
		mirrorMgr := mirror.NewGitMirrorManager(baseDir)
		if mirrorMgr.Exists(repo.Slug) {
			if err := mirrorMgr.Delete(repo.Slug); err != nil {
				render.Warning(fmt.Sprintf("Deleted gitrepo '%s' from database but failed to delete mirror: %v", name, err))
				return nil
			}
		}
	}

	render.Success(fmt.Sprintf("Deleted gitrepo '%s'", name))
	return nil
}

// runSyncGitRepo implements the sync gitrepo command
func runSyncGitRepo(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Get dataStore from context
	dataStore, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// Get the repo
	repo, err := dataStore.GetGitRepoByName(name)
	if err != nil {
		return fmt.Errorf("gitrepo '%s' not found", name)
	}

	// Get MirrorManager
	baseDir := getGitRepoBaseDir()
	mirrorMgr := mirror.NewGitMirrorManager(baseDir)

	// If mirror doesn't exist, clone it first
	if !mirrorMgr.Exists(repo.Slug) {
		if _, err := mirrorMgr.Clone(repo.URL, repo.Slug); err != nil {
			repo.SyncStatus = "failed"
			repo.SyncError = sql.NullString{String: err.Error(), Valid: true}
			dataStore.UpdateGitRepo(repo)
			return fmt.Errorf("failed to clone mirror: %w", err)
		}
	} else {
		// Sync the mirror
		if err := mirrorMgr.Sync(repo.Slug); err != nil {
			repo.SyncStatus = "failed"
			repo.SyncError = sql.NullString{String: err.Error(), Valid: true}
			dataStore.UpdateGitRepo(repo)
			return fmt.Errorf("failed to sync mirror: %w", err)
		}
	}

	// Update repo status
	repo.LastSyncedAt = sql.NullTime{Time: time.Now(), Valid: true}
	repo.SyncStatus = "synced"
	repo.SyncError = sql.NullString{Valid: false}
	if err := dataStore.UpdateGitRepo(repo); err != nil {
		return fmt.Errorf("failed to update repo status: %w", err)
	}

	render.Success(fmt.Sprintf("Synced gitrepo '%s'", name))
	return nil
}

// runSyncGitRepos implements the sync gitrepos command
func runSyncGitRepos(cmd *cobra.Command, args []string) error {
	// Get dataStore from context
	dataStore, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	// List all repos
	repos, err := dataStore.ListGitRepos()
	if err != nil {
		return fmt.Errorf("failed to list gitrepos: %w", err)
	}

	if len(repos) == 0 {
		render.Info("No git repositories to sync")
		return nil
	}

	// Get MirrorManager
	baseDir := getGitRepoBaseDir()
	mirrorMgr := mirror.NewGitMirrorManager(baseDir)

	synced := 0
	failed := 0

	for _, repo := range repos {
		// Get a copy since we need to modify it
		repoPtr := &repo

		// If mirror doesn't exist, clone it first
		if !mirrorMgr.Exists(repo.Slug) {
			if _, err := mirrorMgr.Clone(repo.URL, repo.Slug); err != nil {
				repoPtr.SyncStatus = "failed"
				repoPtr.SyncError = sql.NullString{String: err.Error(), Valid: true}
				dataStore.UpdateGitRepo(repoPtr)
				failed++
				continue
			}
		} else {
			// Sync the mirror
			if err := mirrorMgr.Sync(repo.Slug); err != nil {
				repoPtr.SyncStatus = "failed"
				repoPtr.SyncError = sql.NullString{String: err.Error(), Valid: true}
				dataStore.UpdateGitRepo(repoPtr)
				failed++
				continue
			}
		}

		// Update repo status
		repoPtr.LastSyncedAt = sql.NullTime{Time: time.Now(), Valid: true}
		repoPtr.SyncStatus = "synced"
		repoPtr.SyncError = sql.NullString{Valid: false}
		dataStore.UpdateGitRepo(repoPtr)
		synced++
	}

	if failed > 0 {
		render.Warning(fmt.Sprintf("Synced %d repos, %d failed", synced, failed))
	} else {
		render.Success(fmt.Sprintf("Synced %d repos", synced))
	}

	return nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// getGitRepoBaseDir returns the base directory for git mirrors
func getGitRepoBaseDir() string {
	pc, err := paths.Default()
	if err != nil {
		return "."
	}
	return pc.ReposDir()
}

// gitRepoToMap converts a GitRepoDB to a map for JSON/YAML serialization
func gitRepoToMap(repo *models.GitRepoDB) map[string]interface{} {
	result := map[string]interface{}{
		"name":       repo.Name,
		"url":        repo.URL,
		"slug":       repo.Slug,
		"defaultRef": repo.DefaultRef,
		"authType":   repo.AuthType,
		"autoSync":   repo.AutoSync,
		"syncStatus": repo.SyncStatus,
	}

	if repo.LastSyncedAt.Valid {
		result["lastSyncedAt"] = repo.LastSyncedAt.Time.Format(time.RFC3339)
	}

	if repo.SyncError.Valid {
		result["syncError"] = repo.SyncError.String
	}

	return result
}

// gitReposToYAML converts a slice of GitRepoDB to YAML-friendly format
func gitReposToYAML(repos []models.GitRepoDB) []map[string]interface{} {
	result := make([]map[string]interface{}, len(repos))
	for i, repo := range repos {
		result[i] = gitRepoToMap(&repo)
	}
	return result
}

// gitRepoToYAML converts a GitRepoDB to KeyValueData for human-readable output.
// This follows the get_registry.go pattern for detail views.
func gitRepoToYAML(repo *models.GitRepoDB) render.KeyValueData {
	lastSynced := "never"
	if repo.LastSyncedAt.Valid {
		lastSynced = repo.LastSyncedAt.Time.Format(time.RFC3339)
	}

	pairs := []render.KeyValue{
		{Key: "Name", Value: repo.Name},
		{Key: "URL", Value: repo.URL},
		{Key: "Slug", Value: repo.Slug},
		{Key: "Default Ref", Value: repo.DefaultRef},
		{Key: "Auth Type", Value: repo.AuthType},
		{Key: "Auto Sync", Value: fmt.Sprintf("%v", repo.AutoSync)},
		{Key: "Sync Status", Value: repo.SyncStatus},
		{Key: "Last Synced", Value: lastSynced},
	}

	if repo.SyncError.Valid {
		pairs = append(pairs, render.KeyValue{Key: "Sync Error", Value: repo.SyncError.String})
	}

	// Show credential if associated
	if repo.CredentialID.Valid {
		pairs = append(pairs, render.KeyValue{Key: "Credential ID", Value: fmt.Sprintf("%d", repo.CredentialID.Int64)})
	}

	return render.NewOrderedKeyValueData(pairs...)
}
