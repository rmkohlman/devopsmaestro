package mirror

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Compile-time check: GitMirrorManager implements MirrorInspector.
var _ MirrorInspector = (*GitMirrorManager)(nil)

// ListBranches returns branch refs from a bare mirror.
func (g *GitMirrorManager) ListBranches(slug string) ([]RefInfo, error) {
	return g.listRefs(slug, "refs/heads/")
}

// ListTags returns tag refs from a bare mirror.
func (g *GitMirrorManager) ListTags(slug string) ([]RefInfo, error) {
	return g.listRefs(slug, "refs/tags/")
}

// listRefs runs git for-each-ref and parses the output into RefInfo slices.
func (g *GitMirrorManager) listRefs(slug, refPrefix string) ([]RefInfo, error) {
	mirrorPath := g.GetPath(slug)
	if !g.Exists(slug) {
		return nil, fmt.Errorf("mirror does not exist: %s", mirrorPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	format := "%(refname:short)\t%(objectname:short)\t%(creatordate:iso8601)"
	cmd := exec.CommandContext(ctx, "git", "-C", mirrorPath,
		"for-each-ref", "--format="+format, refPrefix)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, fmt.Errorf("git for-each-ref timed out after 1 minute")
	}
	if err != nil {
		return nil, fmt.Errorf("git for-each-ref failed: %w: %s", err, sanitizeGitOutput(output))
	}

	refs := parseRefOutput(string(output))
	return refs, nil
}

// parseRefOutput parses tab-separated git for-each-ref output into RefInfo.
func parseRefOutput(output string) []RefInfo {
	refs := make([]RefInfo, 0)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		ref := RefInfo{Name: parts[0]}
		if len(parts) > 1 {
			ref.Hash = parts[1]
		}
		if len(parts) > 2 {
			ref.Date = parts[2]
		}
		refs = append(refs, ref)
	}
	return refs
}

// DiskUsage returns the total size in bytes of a mirror directory on disk.
func (g *GitMirrorManager) DiskUsage(slug string) (int64, error) {
	mirrorPath := g.GetPath(slug)
	if !g.Exists(slug) {
		return 0, fmt.Errorf("mirror does not exist: %s", mirrorPath)
	}

	var totalSize int64
	err := filepath.WalkDir(mirrorPath, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			info, infoErr := d.Info()
			if infoErr != nil {
				return infoErr
			}
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to calculate disk usage: %w", err)
	}

	return totalSize, nil
}

// Verify runs git fsck on a mirror and returns nil if the mirror is healthy.
func (g *GitMirrorManager) Verify(slug string) error {
	mirrorPath := g.GetPath(slug)
	if !g.Exists(slug) {
		return fmt.Errorf("mirror does not exist: %s", mirrorPath)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", "-C", mirrorPath, "fsck", "--no-progress")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("git fsck timed out after 5 minutes")
	}
	if err != nil {
		return fmt.Errorf("git fsck failed: %w: %s", err, sanitizeGitOutput(output))
	}

	return nil
}
