// Package source provides unified source resolution for reading resource data
// from various locations: files, URLs, stdin, and GitHub shorthand.
//
// # Usage
//
//	// Resolve a source string to a Source
//	src := source.Resolve("github:user/repo/path/file.yaml")
//
//	// Read the data
//	data, displayName, err := src.Read()
//
// # Supported Source Types
//
//   - File: Local filesystem paths (e.g., "./plugin.yaml", "/path/to/file.yaml")
//   - URL: HTTP/HTTPS URLs (e.g., "https://example.com/plugin.yaml")
//   - GitHub: Shorthand for raw GitHub content (e.g., "github:user/repo/path/file.yaml")
//   - Stdin: Read from standard input (use "-" as the source)
package source

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// Source represents a location that can provide resource data.
type Source interface {
	// Read returns the data from this source.
	// Also returns a display name suitable for user messages.
	Read() (data []byte, displayName string, err error)

	// Type returns the source type for logging/debugging.
	Type() string
}

// Resolve determines the source type from a string and returns the appropriate Source.
// Supported formats:
//   - "-" → StdinSource
//   - "http://" or "https://" → URLSource
//   - "github:user/repo/path" → GitHubSource (converted to URLSource)
//   - anything else → FileSource
func Resolve(s string) Source {
	if s == "-" {
		return &StdinSource{}
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return &URLSource{URL: s}
	}
	if strings.HasPrefix(s, "github:") {
		return NewGitHubSource(s)
	}
	return &FileSource{Path: s}
}

// IsURL returns true if the string looks like a URL (http://, https://, or github:)
func IsURL(s string) bool {
	return strings.HasPrefix(s, "http://") ||
		strings.HasPrefix(s, "https://") ||
		strings.HasPrefix(s, "github:")
}

// FileSource reads data from a local file.
type FileSource struct {
	Path string
}

func (s *FileSource) Read() ([]byte, string, error) {
	slog.Debug("reading file", "path", s.Path)
	data, err := os.ReadFile(s.Path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file %s: %w", s.Path, err)
	}
	return data, s.Path, nil
}

func (s *FileSource) Type() string { return "file" }

// StdinSource reads data from standard input.
type StdinSource struct{}

func (s *StdinSource) Read() ([]byte, string, error) {
	slog.Debug("reading from stdin")
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read from stdin: %w", err)
	}
	return data, "stdin", nil
}

func (s *StdinSource) Type() string { return "stdin" }

// URLSource reads data from an HTTP/HTTPS URL.
type URLSource struct {
	URL     string
	Timeout time.Duration
}

func (s *URLSource) Read() ([]byte, string, error) {
	timeout := s.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	slog.Debug("fetching URL", "url", s.URL)

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(s.URL)
	if err != nil {
		slog.Error("HTTP request failed", "url", s.URL, "error", err)
		return nil, "", fmt.Errorf("failed to fetch %s: %w", s.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("HTTP request returned error", "url", s.URL, "status", resp.StatusCode)
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read response body", "url", s.URL, "error", err)
		return nil, "", fmt.Errorf("failed to read response: %w", err)
	}

	slog.Info("fetched URL successfully", "url", s.URL, "bytes", len(data))
	return data, s.URL, nil
}

func (s *URLSource) Type() string { return "url" }

// GitHubSource converts GitHub shorthand to a raw.githubusercontent.com URL.
// Format: github:user/repo/path/to/file.yaml
// Branch defaults to "main".
type GitHubSource struct {
	Original string    // Original shorthand (e.g., "github:user/repo/path/file.yaml")
	URL      string    // Resolved URL
	inner    URLSource // Delegate to URLSource for actual fetching
}

// NewGitHubSource creates a GitHubSource from shorthand notation.
// Format: github:user/repo/path/to/file.yaml
func NewGitHubSource(shorthand string) *GitHubSource {
	path := strings.TrimPrefix(shorthand, "github:")
	parts := strings.SplitN(path, "/", 3)

	var url string
	if len(parts) >= 3 {
		user := parts[0]
		repo := parts[1]
		filePath := parts[2]
		url = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", user, repo, filePath)
	} else {
		// Invalid format, will fail on Read()
		url = shorthand
	}

	slog.Debug("converted GitHub shorthand", "original", shorthand, "url", url)

	return &GitHubSource{
		Original: shorthand,
		URL:      url,
		inner:    URLSource{URL: url},
	}
}

func (s *GitHubSource) Read() ([]byte, string, error) {
	data, _, err := s.inner.Read()
	if err != nil {
		return nil, "", err
	}
	// Return the resolved URL as display name for clarity
	return data, s.URL, nil
}

func (s *GitHubSource) Type() string { return "github" }

// ReadMultiple reads data from multiple sources.
// Returns a slice of results, one per source.
// Continues on error but records all errors.
type ReadResult struct {
	Data        []byte
	DisplayName string
	Source      Source
	Err         error
}

// ReadAll reads from multiple source strings and returns all results.
func ReadAll(sources []string) []ReadResult {
	results := make([]ReadResult, len(sources))
	for i, s := range sources {
		src := Resolve(s)
		data, name, err := src.Read()
		results[i] = ReadResult{
			Data:        data,
			DisplayName: name,
			Source:      src,
			Err:         err,
		}
	}
	return results
}
