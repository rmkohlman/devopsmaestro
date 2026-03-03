package registry

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

// ValidateRegistryURL validates a registry URL for security.
// It checks:
// - URL is not empty
// - Scheme is http or https
// - No embedded credentials (userinfo)
// - Valid port range (1-65535)
// - Valid hostname
func ValidateRegistryURL(registryURL string) error {
	// Check for empty URL
	if registryURL == "" {
		return fmt.Errorf("registry URL cannot be empty")
	}

	// Parse URL
	parsedURL, err := url.Parse(registryURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	// Validate scheme (must be http or https)
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("invalid URL scheme '%s': only http and https are allowed", parsedURL.Scheme)
	}

	// Check for embedded credentials (security risk)
	if parsedURL.User != nil {
		return fmt.Errorf("registry URL must not contain embedded credentials")
	}

	// Validate hostname is present
	if parsedURL.Host == "" {
		return fmt.Errorf("registry URL must have a valid hostname")
	}

	// Check for hostname-only port (e.g., ":5000")
	hostname := parsedURL.Hostname()
	if hostname == "" {
		return fmt.Errorf("registry URL must have a valid hostname")
	}

	// Check for IPv6 without brackets (e.g., "http://::1:5000" instead of "http://[::1]:5000")
	// If the host contains multiple colons, it's likely an IPv6 address
	// and it should have been wrapped in brackets
	if strings.Count(parsedURL.Host, ":") > 1 && !strings.Contains(parsedURL.Host, "[") {
		return fmt.Errorf("IPv6 addresses must be enclosed in brackets (e.g., http://[::1]:5000)")
	}

	// Validate port if specified
	if parsedURL.Port() != "" {
		port, err := strconv.Atoi(parsedURL.Port())
		if err != nil {
			return fmt.Errorf("invalid port: %w", err)
		}
		if port < 1 || port > 65535 {
			return fmt.Errorf("port must be between 1 and 65535, got %d", port)
		}
	}

	return nil
}

// ValidateRegistryURLWithWarning validates a URL and returns a warning message
// if the URL points to an external (non-localhost) registry.
// Returns empty string if no warning needed.
func ValidateRegistryURLWithWarning(registryURL string) string {
	// First validate the URL is structurally valid
	if err := ValidateRegistryURL(registryURL); err != nil {
		return ""
	}

	// Parse URL to extract hostname
	parsedURL, err := url.Parse(registryURL)
	if err != nil {
		return ""
	}

	// Get hostname without port
	hostname := parsedURL.Hostname()

	// Check if it's localhost
	if IsLocalHost(hostname) {
		return "" // No warning for localhost
	}

	// External URL - return warning
	return fmt.Sprintf("Warning: using external registry %s - ensure it is trusted", hostname)
}

// IsLocalHost checks if a host is localhost, 127.0.0.1, ::1, 0.0.0.0, or host.docker.internal.
// The host parameter can include a port (e.g., "localhost:5000"), which will be stripped.
func IsLocalHost(host string) bool {
	// Strip port if present
	hostname := host
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		// For IPv6, check if there are brackets
		if strings.Contains(host, "]") {
			// IPv6 with port: [::1]:5000
			hostname = host[:idx]
		} else if strings.Count(host, ":") == 1 {
			// IPv4 with port: localhost:5000
			hostname = host[:idx]
		}
		// else: IPv6 without port: ::1 (no stripping needed)
	}

	// Remove brackets from IPv6 addresses
	hostname = strings.Trim(hostname, "[]")

	// Normalize to lowercase for comparison
	hostname = strings.ToLower(hostname)

	// Check against known localhost patterns
	switch hostname {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0", "host.docker.internal":
		return true
	default:
		return false
	}
}
