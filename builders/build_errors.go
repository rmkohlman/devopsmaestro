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

	// Detect BuildKit/containerd cache corruption (#378).
	if isCacheCorruptionMsg(msg) {
		return fmt.Errorf("build failed: %w\n\n"+
			"  Hint: BuildKit cache corruption detected (blob not found / snapshot missing).\n"+
			"  Try: dvm build --clean-cache    (prune BuildKit cache and retry)\n"+
			"  Or:  dvm cache clear --buildkit  (manual cache cleanup)\n"+
			"  Last resort: delete the Colima disk and recreate the VM.", err)
	}

	// Detect network timeout during image pull — likely squid proxy is down (#377).
	if isNetworkTimeout(msg) {
		return fmt.Errorf("build failed: %w\n\n"+
			"  Hint: Network timeout during image pull. The HTTP proxy (squid) may not be running.\n"+
			"  Registry cache unavailable. Run 'dvm restart registries' to fix.\n"+
			"  If the problem persists, run 'dvm get registries' to check proxy status.", err)
	}

	return fmt.Errorf("build failed: %w", err)
}

// IsCacheCorruption returns true if the error message indicates BuildKit/containerd
// cache corruption — typically "blob not found" or "parent snapshot does not exist"
// errors that make all subsequent builds fail until the cache is wiped (#378).
func IsCacheCorruption(err error) bool {
	if err == nil {
		return false
	}
	return isCacheCorruptionMsg(err.Error())
}

func isCacheCorruptionMsg(msg string) bool {
	corruptionPatterns := []string{
		"blob not found",
		"parent snapshot does not exist",
		"failed to get reader from content store",
		"missing content",
		"failed to verify",
		"unexpected end of json input",
		"index.json: no such file or directory",
	}
	lower := strings.ToLower(msg)
	for _, p := range corruptionPatterns {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return false
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
