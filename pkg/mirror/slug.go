package mirror

import (
	"fmt"
	"net/url"
	"strings"
)

// NormalizeGitURL normalizes a git URL to a standard format.
// Always adds .git suffix if not present.
// Handles both HTTPS and SSH formats.
func NormalizeGitURL(rawURL string) string {
	// Already has .git suffix
	if strings.HasSuffix(rawURL, ".git") {
		return rawURL
	}

	// Add .git suffix
	return rawURL + ".git"
}

// GenerateSlug creates a filesystem-safe slug from a git URL.
// Format: {host}_{owner}_{repo}
// Examples:
//   - https://github.com/user/repo -> github.com_user_repo
//   - git@github.com:user/repo -> github.com_user_repo
//   - https://gitlab.com/group/subgroup/project -> gitlab.com_group_subgroup_project
func GenerateSlug(rawURL string) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	// Strip .git suffix if present for parsing
	cleanURL := strings.TrimSuffix(rawURL, ".git")

	var host, path string

	// Handle SSH format: git@github.com:user/repo
	if strings.HasPrefix(cleanURL, "git@") {
		parts := strings.SplitN(cleanURL, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid SSH URL format: %s", rawURL)
		}
		// Remove "git@" prefix
		host = strings.TrimPrefix(parts[0], "git@")
		path = parts[1]
	} else {
		// Handle HTTPS format
		parsed, err := url.Parse(cleanURL)
		if err != nil {
			return "", fmt.Errorf("invalid URL: %w", err)
		}
		host = parsed.Host
		path = strings.TrimPrefix(parsed.Path, "/")
	}

	if host == "" || path == "" {
		return "", fmt.Errorf("could not extract host and path from URL: %s", rawURL)
	}

	// Replace path separators with underscores
	pathSlug := strings.ReplaceAll(path, "/", "_")

	// Create slug: host_path
	slug := host + "_" + pathSlug

	return slug, nil
}
