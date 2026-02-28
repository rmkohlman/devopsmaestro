package mirror

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// GitMirrorManager implements MirrorManager for managing bare git mirrors.
type GitMirrorManager struct {
	baseDir string // e.g., ~/.devopsmaestro/repos/
}

// NewGitMirrorManager creates a new GitMirrorManager with the specified base directory.
func NewGitMirrorManager(baseDir string) *GitMirrorManager {
	return &GitMirrorManager{
		baseDir: baseDir,
	}
}

// Clone creates a new bare mirror from a remote URL.
func (g *GitMirrorManager) Clone(url string, slug string) (string, error) {
	// Validate URL
	if err := ValidateGitURL(url); err != nil {
		return "", err
	}

	// Validate slug
	if err := ValidateSlug(slug); err != nil {
		return "", err
	}

	// Get mirror path
	mirrorPath := g.GetPath(slug)

	// Check if mirror already exists
	if g.Exists(slug) {
		return "", fmt.Errorf("mirror already exists at %s", mirrorPath)
	}

	// Create base directory if needed
	if err := os.MkdirAll(g.baseDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create base directory: %w", err)
	}

	// 5 minute timeout for clone operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Execute: git clone --mirror -- <url> <mirrorPath>
	cmd := exec.CommandContext(ctx, "git", "clone", "--mirror", "--", url, mirrorPath)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return "", fmt.Errorf("git clone timed out after 5 minutes")
	}
	if err != nil {
		// Clean up partial clone
		os.RemoveAll(mirrorPath)
		return "", fmt.Errorf("git clone failed: %w: %s", err, sanitizeGitOutput(output))
	}

	// Set directory permissions to 0700
	if err := os.Chmod(mirrorPath, 0700); err != nil {
		// Clean up and return error
		os.RemoveAll(mirrorPath)
		return "", fmt.Errorf("failed to set mirror permissions: %w", err)
	}

	return mirrorPath, nil
}

// Sync updates an existing mirror from its remote.
func (g *GitMirrorManager) Sync(slug string) error {
	// Validate slug
	if err := ValidateSlug(slug); err != nil {
		return err
	}

	// Verify mirror exists
	mirrorPath := g.GetPath(slug)
	if !g.Exists(slug) {
		return fmt.Errorf("mirror does not exist: %s", mirrorPath)
	}

	// 5 minute timeout for sync operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Execute: git remote update --prune
	cmd := exec.CommandContext(ctx, "git", "-C", mirrorPath, "remote", "update", "--prune")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("git remote update timed out after 5 minutes")
	}
	if err != nil {
		return fmt.Errorf("git remote update failed: %w: %s", err, sanitizeGitOutput(output))
	}

	return nil
}

// Delete removes a mirror from disk.
func (g *GitMirrorManager) Delete(slug string) error {
	// Validate slug
	if err := ValidateSlug(slug); err != nil {
		return err
	}

	mirrorPath := g.GetPath(slug)
	if _, err := os.Stat(mirrorPath); os.IsNotExist(err) {
		return nil // Idempotent - already deleted
	}

	// Atomic-ish delete: rename first, then remove
	// This prevents partial deletion issues
	tmpPath := mirrorPath + ".deleting." + fmt.Sprintf("%d", time.Now().UnixNano())
	if err := os.Rename(mirrorPath, tmpPath); err != nil {
		return fmt.Errorf("failed to prepare mirror for deletion: %w", err)
	}

	return os.RemoveAll(tmpPath)
}

// Exists checks if a mirror exists locally.
func (g *GitMirrorManager) Exists(slug string) bool {
	mirrorPath := g.GetPath(slug)

	// Check if directory exists
	info, err := os.Stat(mirrorPath)
	if err != nil {
		return false
	}
	if !info.IsDir() {
		return false
	}

	// Check if it's a git repository (has HEAD file in bare repo)
	headFile := filepath.Join(mirrorPath, "HEAD")
	if _, err := os.Stat(headFile); err != nil {
		return false
	}

	return true
}

// GetPath returns the filesystem path for a mirror.
func (g *GitMirrorManager) GetPath(slug string) string {
	return filepath.Join(g.baseDir, slug)
}

// CloneToWorkspace clones from a mirror to a workspace path.
func (g *GitMirrorManager) CloneToWorkspace(mirrorSlug string, destPath string, ref string) error {
	// Validate slug
	if err := ValidateSlug(mirrorSlug); err != nil {
		return err
	}

	// Validate destination path
	if err := ValidateDestPath(destPath); err != nil {
		return err
	}

	// Validate ref if provided
	if err := ValidateGitRef(ref); err != nil {
		return err
	}

	// Verify mirror exists
	mirrorPath := g.GetPath(mirrorSlug)
	if !g.Exists(mirrorSlug) {
		return fmt.Errorf("mirror does not exist: %s", mirrorPath)
	}

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return fmt.Errorf("destination already exists: %s", destPath)
	}

	// 5 minute timeout for clone operations
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Execute: git clone -- <mirrorPath> <destPath>
	cmd := exec.CommandContext(ctx, "git", "clone", "--", mirrorPath, destPath)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.CombinedOutput()
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("git clone timed out after 5 minutes")
	}
	if err != nil {
		return fmt.Errorf("git clone failed: %w: %s", err, sanitizeGitOutput(output))
	}

	// If ref is provided, checkout that ref
	if ref != "" {
		ctx2, cancel2 := context.WithTimeout(context.Background(), 1*time.Minute)
		defer cancel2()

		// Try checkout as-is first (works for tags and local branches)
		cmd = exec.CommandContext(ctx2, "git", "-C", destPath, "checkout", "--", ref)
		cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

		output, err = cmd.CombinedOutput()
		if err != nil {
			// Checkout failed - might be a remote branch
			// Try creating a local branch tracking the remote branch
			cmd2 := exec.CommandContext(ctx2, "git", "-C", destPath, "checkout", "-b", ref, "origin/"+ref)
			cmd2.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

			_, err2 := cmd2.CombinedOutput()
			if err2 != nil {
				// Both failed - return original error with more context
				if ctx2.Err() == context.DeadlineExceeded {
					return fmt.Errorf("git checkout timed out after 1 minute")
				}
				return fmt.Errorf("git checkout failed: %w: %s (also tried origin/%s)", err, sanitizeGitOutput(output), ref)
			}
		}
	}

	// Set remote URL to original
	originalURL, err := g.getOriginalURL(mirrorPath)
	if err != nil {
		return fmt.Errorf("failed to get original URL: %w", err)
	}

	ctx3, cancel3 := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel3()

	cmd = exec.CommandContext(ctx3, "git", "-C", destPath, "remote", "set-url", "origin", "--", originalURL)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err = cmd.CombinedOutput()
	if ctx3.Err() == context.DeadlineExceeded {
		return fmt.Errorf("git remote set-url timed out after 30 seconds")
	}
	if err != nil {
		return fmt.Errorf("git remote set-url failed: %w: %s", err, sanitizeGitOutput(output))
	}

	return nil
}

// getOriginalURL reads the original URL from the mirror's git config.
func (g *GitMirrorManager) getOriginalURL(mirrorPath string) (string, error) {
	cmd := exec.Command("git", "-C", mirrorPath, "config", "--get", "remote.origin.url")
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	url := strings.TrimSpace(string(output))

	// Re-validate the URL before using it
	if err := ValidateGitURL(url); err != nil {
		return "", fmt.Errorf("stored URL is invalid: %w", err)
	}

	return url, nil
}
