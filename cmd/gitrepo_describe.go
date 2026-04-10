package cmd

import (
	"fmt"
	"strings"
	"time"

	"devopsmaestro/db"
	"devopsmaestro/pkg/mirror"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// runDescribeGitRepo implements the describe gitrepo command.
// It shows rich status with mirror health, disk usage, ref counts, and linked resources.
func runDescribeGitRepo(cmd *cobra.Command, args []string) error {
	name := args[0]

	dataStore, err := getDataStore(cmd)
	if err != nil {
		return err
	}

	repo, err := dataStore.GetGitRepoByName(name)
	if err != nil {
		return fmt.Errorf("gitrepo '%s' not found", name)
	}

	mm := getMirrorManager(cmd)

	// Build basic key-value pairs from DB data
	lastSynced := "never"
	if repo.LastSyncedAt.Valid {
		lastSynced = repo.LastSyncedAt.Time.Format(time.RFC3339)
	}
	syncError := "(none)"
	if repo.SyncError.Valid && repo.SyncError.String != "" {
		syncError = repo.SyncError.String
	}

	pairs := []render.KeyValue{
		{Key: "Name", Value: repo.Name},
		{Key: "URL", Value: repo.URL},
		{Key: "Slug", Value: repo.Slug},
		{Key: "Default Ref", Value: repo.DefaultRef},
		{Key: "Auth Type", Value: repo.AuthType},
		{Key: "Sync Status", Value: repo.SyncStatus},
		{Key: "Last Synced", Value: lastSynced},
		{Key: "Sync Error", Value: syncError},
	}

	// Add mirror inspection data (fire-and-forget for non-critical errors)
	pairs = appendMirrorInfo(mm, repo.Slug, pairs)

	// Add linked resources
	pairs = appendLinkedResources(dataStore, int64(repo.ID), pairs)

	format, _ := cmd.Flags().GetString("output")
	kvData := render.NewOrderedKeyValueData(pairs...)
	return render.OutputTo(cmd.OutOrStdout(), format, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "GitRepo Details",
	})
}

// appendMirrorInfo adds disk usage, branch count, tag count, and health to pairs.
func appendMirrorInfo(mm mirror.MirrorManager, slug string, pairs []render.KeyValue) []render.KeyValue {
	inspector, ok := mm.(mirror.MirrorInspector)
	if !ok {
		pairs = append(pairs, render.KeyValue{Key: "Mirror Health", Value: "unknown (no inspector)"})
		return pairs
	}

	// Mirror path
	pairs = append(pairs, render.KeyValue{Key: "Mirror Path", Value: mm.GetPath(slug)})

	// Disk usage
	if bytes, err := inspector.DiskUsage(slug); err == nil {
		pairs = append(pairs, render.KeyValue{Key: "Disk Usage", Value: formatBytes(bytes)})
	} else {
		pairs = append(pairs, render.KeyValue{Key: "Disk Usage", Value: "unavailable"})
	}

	// Branch count
	if branches, err := inspector.ListBranches(slug); err == nil {
		pairs = append(pairs, render.KeyValue{Key: "Branch Count", Value: fmt.Sprintf("%d", len(branches))})
	} else {
		pairs = append(pairs, render.KeyValue{Key: "Branch Count", Value: "unavailable"})
	}

	// Tag count
	if tags, err := inspector.ListTags(slug); err == nil {
		pairs = append(pairs, render.KeyValue{Key: "Tag Count", Value: fmt.Sprintf("%d", len(tags))})
	} else {
		pairs = append(pairs, render.KeyValue{Key: "Tag Count", Value: "unavailable"})
	}

	// Mirror health (git fsck)
	if err := inspector.Verify(slug); err == nil {
		pairs = append(pairs, render.KeyValue{Key: "Mirror Health", Value: "ok"})
	} else {
		pairs = append(pairs, render.KeyValue{Key: "Mirror Health", Value: fmt.Sprintf("unhealthy: %v", err)})
	}

	return pairs
}

// appendLinkedResources adds linked apps and workspaces to the describe output.
func appendLinkedResources(ds db.DataStore, repoID int64, pairs []render.KeyValue) []render.KeyValue {
	// Linked apps
	if apps, err := ds.ListAppsByGitRepoID(repoID); err == nil && len(apps) > 0 {
		names := make([]string, len(apps))
		for i, a := range apps {
			names[i] = a.Name
		}
		pairs = append(pairs, render.KeyValue{Key: "Linked Apps", Value: strings.Join(names, ", ")})
	} else {
		pairs = append(pairs, render.KeyValue{Key: "Linked Apps", Value: "(none)"})
	}

	// Linked workspaces
	if workspaces, err := ds.ListWorkspacesByGitRepoID(repoID); err == nil && len(workspaces) > 0 {
		names := make([]string, len(workspaces))
		for i, ws := range workspaces {
			names[i] = ws.Name
		}
		pairs = append(pairs, render.KeyValue{Key: "Linked Workspaces", Value: strings.Join(names, ", ")})
	} else {
		pairs = append(pairs, render.KeyValue{Key: "Linked Workspaces", Value: "(none)"})
	}

	return pairs
}

// formatBytes converts bytes to a human-readable string.
func formatBytes(b int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case b >= gb:
		return fmt.Sprintf("%.1f GB", float64(b)/float64(gb))
	case b >= mb:
		return fmt.Sprintf("%.1f MB", float64(b)/float64(mb))
	case b >= kb:
		return fmt.Sprintf("%.1f KB", float64(b)/float64(kb))
	default:
		return fmt.Sprintf("%d B", b)
	}
}
