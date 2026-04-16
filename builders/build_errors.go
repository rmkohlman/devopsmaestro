package builders

import (
	"fmt"
	"strings"
)

// EnhanceBuildError wraps a build error with actionable context when the root
// cause is a known infrastructure issue (e.g., squid proxy down → network
// timeout). This helps users diagnose build failures without reading logs.
func EnhanceBuildError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()

	// Detect network timeout during image pull — likely squid proxy is down (#377).
	if isNetworkTimeout(msg) {
		return fmt.Errorf("build failed: %w\n\n"+
			"  Hint: Network timeout during image pull. The HTTP proxy (squid) may not be running.\n"+
			"  Registry cache unavailable. Run 'dvm restart registries' to fix.\n"+
			"  If the problem persists, run 'dvm get registries' to check proxy status.", err)
	}

	return fmt.Errorf("build failed: %w", err)
}

// isNetworkTimeout returns true if the error message indicates a network
// connectivity failure typical of a missing HTTP proxy.
func isNetworkTimeout(msg string) bool {
	timeoutPatterns := []string{
		"i/o timeout",
		"dial tcp",
		"connection refused",
		"no such host",
		"network is unreachable",
	}
	lower := strings.ToLower(msg)
	for _, p := range timeoutPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
}
