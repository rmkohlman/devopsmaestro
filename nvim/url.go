package nvim

import (
	"strings"
)

// IsGitURL checks if a string looks like a Git URL
func IsGitURL(s string) bool {
	// Check for common Git URL patterns
	patterns := []string{
		"http://",
		"https://",
		"git://",
		"git@",
		"github:",
		"gitlab:",
		"bitbucket:",
	}

	for _, pattern := range patterns {
		if strings.HasPrefix(s, pattern) {
			return true
		}
	}

	// Check if it ends with .git
	if strings.HasSuffix(s, ".git") {
		return true
	}

	return false
}

// NormalizeGitURL converts shorthand Git URLs to full URLs
func NormalizeGitURL(url string) string {
	// Handle github:user/repo format
	if strings.HasPrefix(url, "github:") {
		repo := strings.TrimPrefix(url, "github:")
		return "https://github.com/" + repo + ".git"
	}

	// Handle gitlab:user/repo format
	if strings.HasPrefix(url, "gitlab:") {
		repo := strings.TrimPrefix(url, "gitlab:")
		return "https://gitlab.com/" + repo + ".git"
	}

	// Handle bitbucket:user/repo format
	if strings.HasPrefix(url, "bitbucket:") {
		repo := strings.TrimPrefix(url, "bitbucket:")
		return "https://bitbucket.org/" + repo + ".git"
	}

	// If already a full URL, return as-is
	return url
}

// ParseGitURL extracts useful information from a Git URL
func ParseGitURL(url string) GitURLInfo {
	info := GitURLInfo{
		FullURL: url,
	}

	// Normalize first
	normalized := NormalizeGitURL(url)
	info.FullURL = normalized

	// Extract repo name (last part before .git)
	parts := strings.Split(normalized, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		info.RepoName = strings.TrimSuffix(lastPart, ".git")
	}

	// Detect platform
	if strings.Contains(normalized, "github.com") {
		info.Platform = "github"
	} else if strings.Contains(normalized, "gitlab.com") {
		info.Platform = "gitlab"
	} else if strings.Contains(normalized, "bitbucket.org") {
		info.Platform = "bitbucket"
	} else {
		info.Platform = "git"
	}

	return info
}

// GitURLInfo contains parsed Git URL information
type GitURLInfo struct {
	FullURL  string // Full normalized URL
	Platform string // github, gitlab, bitbucket, or git
	RepoName string // Repository name
}
