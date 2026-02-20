// Package source provides unified source resolution for reading resource data
// from various locations including GitHub directories.
package source

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"runtime"
	"strings"
	"time"

	"devopsmaestro/pkg/secrets"
	"devopsmaestro/pkg/secrets/providers"
)

// GitHubDirectorySource lists and fetches all YAML files from a GitHub directory.
type GitHubDirectorySource struct {
	Original string // Original source string (e.g., "github:user/repo/plugins/")
	Owner    string // GitHub owner/user
	Repo     string // Repository name
	Path     string // Path within the repository
	Branch   string // Branch name (defaults to "main")
}

// GitHubFile represents a file entry from the GitHub Contents API.
type GitHubFile struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Type        string `json:"type"`         // "file" or "dir"
	DownloadURL string `json:"download_url"` // Raw content URL
}

// NewGitHubDirectorySource creates a GitHubDirectorySource from various URL formats.
// Supported formats:
//   - github:user/repo/path/
//   - github:user/repo/path (no extension = directory)
//   - https://github.com/user/repo/tree/branch/path/
func NewGitHubDirectorySource(s string) *GitHubDirectorySource {
	src := &GitHubDirectorySource{
		Original: s,
		Branch:   "main", // Default branch
	}

	// Handle github: shorthand
	if strings.HasPrefix(s, "github:") {
		path := strings.TrimPrefix(s, "github:")
		path = strings.TrimSuffix(path, "/")
		parts := strings.SplitN(path, "/", 3)

		if len(parts) >= 2 {
			src.Owner = parts[0]
			src.Repo = parts[1]
			if len(parts) >= 3 {
				src.Path = parts[2]
			}
		}
		return src
	}

	// Handle https://github.com/user/repo/tree/branch/path/ URLs
	if strings.HasPrefix(s, "https://github.com/") {
		path := strings.TrimPrefix(s, "https://github.com/")
		path = strings.TrimSuffix(path, "/")
		parts := strings.Split(path, "/")

		if len(parts) >= 2 {
			src.Owner = parts[0]
			src.Repo = parts[1]

			// Check for /tree/branch/ pattern
			if len(parts) >= 4 && parts[2] == "tree" {
				src.Branch = parts[3]
				if len(parts) >= 5 {
					src.Path = strings.Join(parts[4:], "/")
				}
			} else if len(parts) >= 3 {
				// Just user/repo/path
				src.Path = strings.Join(parts[2:], "/")
			}
		}
		return src
	}

	return src
}

// Type returns the directory source type for logging/debugging.
func (s *GitHubDirectorySource) Type() string { return "github-directory" }

// ListFiles returns a list of Source for each YAML file in this directory.
// This implements the DirectorySource interface.
func (s *GitHubDirectorySource) ListFiles() ([]Source, error) {
	if s.Owner == "" || s.Repo == "" {
		return nil, fmt.Errorf("invalid GitHub directory source: missing owner or repo in %q", s.Original)
	}

	// Build the GitHub API URL
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", s.Owner, s.Repo, s.Path)
	if s.Branch != "" && s.Branch != "main" {
		apiURL += "?ref=" + s.Branch
	}

	slog.Debug("fetching GitHub directory listing", "url", apiURL, "owner", s.Owner, "repo", s.Repo, "path", s.Path)

	// Create HTTP request
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authorization header if GitHub token is available
	if token := getGitHubToken(); token != "" {
		req.Header.Set("Authorization", "token "+token)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "dvm")

	// Make the request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch directory listing: %w", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting
	if resp.StatusCode == http.StatusForbidden {
		remaining := resp.Header.Get("X-RateLimit-Remaining")
		if remaining == "0" {
			return nil, fmt.Errorf("GitHub API rate limit exceeded. Set GITHUB_TOKEN env var for higher limits (5000/hour vs 60/hour)")
		}
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var files []GitHubFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	// Filter for YAML files only
	var sources []Source
	for _, f := range files {
		if f.Type != "file" {
			continue
		}

		lowerName := strings.ToLower(f.Name)
		if !strings.HasSuffix(lowerName, ".yaml") && !strings.HasSuffix(lowerName, ".yml") {
			continue
		}

		// Create a GitHubSource for each file
		// Use the download URL directly
		src := &GitHubSource{
			Original: fmt.Sprintf("github:%s/%s/%s", s.Owner, s.Repo, f.Path),
			URL:      f.DownloadURL,
			inner:    URLSource{URL: f.DownloadURL},
		}
		sources = append(sources, src)

		slog.Debug("found YAML file", "name", f.Name, "path", f.Path)
	}

	slog.Info("listed GitHub directory", "path", s.Path, "yaml_files", len(sources))

	return sources, nil
}

// Name returns the filename portion of a GitHubSource.
func (s *GitHubSource) Name() string {
	// Extract filename from the path
	parts := strings.Split(s.Original, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return s.Original
}

// getGitHubToken retrieves the GitHub token using the secret provider system.
// It tries providers in this order:
//  1. Keychain (macOS only) - looks for "github-token" in devopsmaestro service
//  2. Environment variable DVM_SECRET_GITHUB_TOKEN
//  3. Environment variable GITHUB_TOKEN (backward compatibility)
//
// Returns empty string if no token is found (graceful degradation).
func getGitHubToken() string {
	ctx := context.Background()

	// Try keychain first on macOS
	if runtime.GOOS == "darwin" {
		kc := providers.NewKeychainProvider()
		if kc.IsAvailable() {
			token, err := kc.GetSecret(ctx, secrets.SecretRequest{Name: "github-token"})
			if err == nil && token != "" {
				slog.Debug("using GitHub token from keychain")
				return token
			}
			// Log if there was an error other than "not found"
			if err != nil && !secrets.IsNotFound(err) {
				slog.Debug("keychain lookup failed", "error", err)
			}
		}
	}

	// Fallback to environment variable (checks DVM_SECRET_GITHUB_TOKEN then GITHUB_TOKEN)
	env := providers.NewEnvProvider()
	token, err := env.GetSecret(ctx, secrets.SecretRequest{Name: "github-token"})
	if err == nil && token != "" {
		slog.Debug("using GitHub token from environment variable")
		return token
	}

	return ""
}
