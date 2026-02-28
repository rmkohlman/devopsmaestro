package mirror

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Shell metacharacters that could be used for command injection
var shellMetachars = []string{";", "|", "&", "$", "`", "(", ")", "<", ">"}

// ValidateGitURL validates a git URL for security.
// It rejects:
// - Empty URLs
// - Shell metacharacters (command injection)
// - ext:: prefix (Git remote-ext exploit)
// - Leading dash (command injection)
// - Insecure protocols (http://, git://)
// - file:// protocol
// - Embedded credentials (except git@host SSH format)
// - Whitespace (space, tab, newline)
//
// It allows:
// - HTTPS URLs (https://...)
// - SSH URLs (git@host:path or ssh://...)
// - Local filesystem paths (absolute/relative) for testing and development
//
// NOTE: Local filesystem paths are allowed to support testing and local development
// workflows. In production CLI usage, applications should validate that user-provided
// URLs are remote URLs before calling MirrorManager methods. The use of `--` separator
// in git commands provides defense-in-depth against option injection.
func ValidateGitURL(url string) error {
	// Empty check
	if url == "" {
		return fmt.Errorf("git URL is empty")
	}

	// Whitespace check (space, tab, newline, carriage return)
	if strings.ContainsAny(url, " \t\n\r") {
		return fmt.Errorf("git URL is invalid: contains whitespace")
	}

	// Leading dash (command injection via git options)
	if strings.HasPrefix(url, "-") {
		return fmt.Errorf("git URL is invalid: starts with dash")
	}

	// ext:: prefix (Git remote-ext exploit CVE-2017-1000117)
	if strings.HasPrefix(strings.ToLower(url), "ext::") {
		return fmt.Errorf("git URL is invalid: ext transport not allowed")
	}

	// Shell metacharacters check
	for _, char := range shellMetachars {
		if strings.Contains(url, char) {
			return fmt.Errorf("git URL is invalid: contains shell metacharacter")
		}
	}

	// file:// protocol (local filesystem access - explicit protocol is rejected)
	if strings.HasPrefix(strings.ToLower(url), "file://") {
		return fmt.Errorf("git URL is invalid: file protocol not allowed")
	}

	// Insecure protocols
	if strings.HasPrefix(strings.ToLower(url), "http://") {
		return fmt.Errorf("git URL is insecure: http protocol not allowed, use https")
	}
	if strings.HasPrefix(strings.ToLower(url), "git://") {
		return fmt.Errorf("git URL is insecure: git protocol not allowed, use https or ssh")
	}

	// Check for embedded credentials
	if err := ValidateURLNoCredentials(url); err != nil {
		return err
	}

	// SECURITY: Allowlist approach - only allow known-safe protocols
	// After all blocklist checks pass, verify URL uses an allowed protocol
	lowerURL := strings.ToLower(url)

	// Check for allowed protocols
	isHTTPS := strings.HasPrefix(lowerURL, "https://")
	isSSH := strings.HasPrefix(lowerURL, "ssh://")
	isGitSSH := strings.HasPrefix(url, "git@") // Case-sensitive for git@

	// Check for local filesystem path (no protocol, for testing)
	isLocalPath := !strings.Contains(url, "://")

	if !isHTTPS && !isSSH && !isGitSSH && !isLocalPath {
		return fmt.Errorf("git URL is invalid: unsupported protocol (use https://, ssh://, or git@host:path)")
	}

	return nil
}

// ValidateSlug validates a slug for filesystem safety.
// It rejects:
// - Empty slugs
// - Path traversal (..)
// - Path separators (/ or \)
// - Absolute paths
// - Null bytes
// - Whitespace
// - Shell metacharacters
// - Leading dash or dot
// - Single dot (.) or double dot (..)
func ValidateSlug(slug string) error {
	// Empty check
	if slug == "" {
		return fmt.Errorf("slug is empty")
	}

	// Single dot or double dot
	if slug == "." || slug == ".." {
		return fmt.Errorf("slug is invalid: cannot be . or ..")
	}

	// Leading dash (could be interpreted as option)
	if strings.HasPrefix(slug, "-") {
		return fmt.Errorf("slug is invalid: cannot start with dash")
	}

	// Leading dot (hidden file)
	if strings.HasPrefix(slug, ".") {
		return fmt.Errorf("slug is invalid: cannot start with dot")
	}

	// Path traversal
	if strings.Contains(slug, "..") {
		return fmt.Errorf("slug is invalid: path traversal not allowed")
	}

	// Path separators
	if strings.Contains(slug, "/") {
		return fmt.Errorf("slug is invalid: forward slash not allowed")
	}
	if strings.Contains(slug, "\\") {
		return fmt.Errorf("slug is invalid: backslash not allowed")
	}

	// Null byte
	if strings.Contains(slug, "\x00") {
		return fmt.Errorf("slug is invalid: null byte not allowed")
	}

	// Whitespace (space, tab, newline, carriage return)
	if strings.ContainsAny(slug, " \t\n\r") {
		return fmt.Errorf("slug is invalid: whitespace not allowed")
	}

	// Shell metacharacters
	for _, char := range shellMetachars {
		if strings.Contains(slug, char) {
			return fmt.Errorf("slug is invalid: shell metacharacter not allowed")
		}
	}

	return nil
}

// ValidateURLNoCredentials checks that a URL doesn't contain embedded credentials.
// It allows:
// - git@host:path SSH format (special case)
// - HTTPS URLs without credentials
// - Colons in paths (not credentials)
// It rejects:
// - user:pass@host
// - user@host (except git@)
// - :pass@host
func ValidateURLNoCredentials(url string) error {
	// Special case: git@host:path SSH format is allowed
	if strings.HasPrefix(url, "git@") {
		return nil
	}

	// Check for @ in the URL before the first / (userinfo section)
	// For https://user:pass@host/path, the @ appears before /host/path

	// Find the protocol separator
	protoEnd := strings.Index(url, "://")
	if protoEnd == -1 {
		// No protocol - could be user@host:/path SSH style (not git@)
		atIdx := strings.Index(url, "@")
		if atIdx != -1 {
			// Check if there's a colon after @ (user@host:path)
			colonAfterAt := strings.Index(url[atIdx:], ":")
			if colonAfterAt != -1 {
				// This is user@host:path SSH format - credentials detected
				return fmt.Errorf("git URL contains credentials")
			}
		}
		return nil
	}

	// Get the part after the protocol (e.g., "user:pass@host/path")
	afterProto := url[protoEnd+3:]

	// Find the first slash (separates authority from path)
	slashIdx := strings.Index(afterProto, "/")
	var authority string
	if slashIdx == -1 {
		authority = afterProto
	} else {
		authority = afterProto[:slashIdx]
	}

	// Check for @ in the authority section (indicates userinfo)
	if strings.Contains(authority, "@") {
		return fmt.Errorf("git URL contains credentials")
	}

	return nil
}

// ValidateGitRef validates a git ref (branch, tag, or commit SHA) for security.
// It rejects:
// - Leading dash (could be interpreted as option)
// - Shell metacharacters
// - Null bytes
// - Whitespace
// - Path traversal (..)
func ValidateGitRef(ref string) error {
	if ref == "" {
		return nil // Empty ref is valid (use default branch)
	}

	// Leading dash - could be interpreted as option
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("git ref is invalid: cannot start with dash")
	}

	// Shell metacharacters
	for _, char := range shellMetachars {
		if strings.Contains(ref, char) {
			return fmt.Errorf("git ref is invalid: contains shell metacharacter")
		}
	}

	// Null byte
	if strings.Contains(ref, "\x00") {
		return fmt.Errorf("git ref is invalid: contains null byte")
	}

	// Whitespace
	if strings.ContainsAny(ref, " \t\n\r") {
		return fmt.Errorf("git ref is invalid: contains whitespace")
	}

	// Path traversal (refs could be like "refs/heads/../../../etc")
	if strings.Contains(ref, "..") {
		return fmt.Errorf("git ref is invalid: path traversal not allowed")
	}

	return nil
}

// ValidateDestPath validates a destination path for cloning operations.
// It checks that the path is absolute and not in critical system directories.
func ValidateDestPath(destPath string) error {
	// Clean the path first
	cleanPath := filepath.Clean(destPath)

	// Must be absolute
	if !filepath.IsAbs(cleanPath) {
		return fmt.Errorf("destination path must be absolute")
	}

	// Reject critical system paths (but allow /var/folders for macOS temp, /tmp, etc.)
	dangerousPrefixes := []string{
		"/etc",
		"/usr",
		"/bin",
		"/sbin",
		"/root",
		"/System",                // macOS system directory
		"/Library/StartupItems",  // macOS critical
		"/Library/LaunchAgents",  // macOS critical
		"/Library/LaunchDaemons", // macOS critical
	}

	for _, prefix := range dangerousPrefixes {
		if strings.HasPrefix(cleanPath, prefix) {
			return fmt.Errorf("destination path %s is not allowed", prefix)
		}
	}

	return nil
}

// sanitizeGitOutput removes potentially sensitive information from git output.
func sanitizeGitOutput(output []byte) string {
	s := string(output)

	// Remove lines containing potential secrets
	sensitivePatterns := []string{
		"password",
		"token",
		"credential",
		"ssh",
		"key",
	}

	lines := strings.Split(s, "\n")
	var sanitized []string
	for _, line := range lines {
		lower := strings.ToLower(line)
		isSensitive := false
		for _, pattern := range sensitivePatterns {
			if strings.Contains(lower, pattern) {
				isSensitive = true
				break
			}
		}
		if !isSensitive {
			sanitized = append(sanitized, line)
		}
	}

	// Truncate if too long
	result := strings.Join(sanitized, "\n")
	if len(result) > 500 {
		result = result[:500] + "... (truncated)"
	}

	return result
}
