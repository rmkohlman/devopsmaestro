package cmd

import "fmt"

// FormatBuildNoMatchScopeError formats an error message when no workspaces
// match a scope filter (e.g., --ecosystem "does-not-exist"). Includes a hint
// directing users to 'dvm get workspaces -A' to list available workspaces.
//
// Example output:
//
//	error: no workspaces found matching --ecosystem "does-not-exist"
//	Hint: Run 'dvm get workspaces -A' to see all available workspaces
func FormatBuildNoMatchScopeError(scopeFlag, scopeValue string) string {
	return fmt.Sprintf(
		"error: no workspaces found matching --%s %q\nHint: Run 'dvm get workspaces -A' to see all available workspaces",
		scopeFlag, scopeValue,
	)
}

// FormatParallelBuildHeader formats the streaming progress header shown at the
// start of a parallel build session. When scopeLabel is empty, the "in <scope>"
// portion is omitted (used for --all without a scope filter).
//
// Example output:
//
//	Building 6 workspace(s) in ecosystem "healthcare" (concurrency: 4)
//	Building 12 workspace(s) (concurrency: 8)
func FormatParallelBuildHeader(count int, scopeLabel, scopeValue string, concurrency int) string {
	if scopeLabel != "" && scopeValue != "" {
		return fmt.Sprintf("Building %d workspace(s) in %s %q (concurrency: %d)",
			count, scopeLabel, scopeValue, concurrency)
	}
	return fmt.Sprintf("Building %d workspace(s) (concurrency: %d)",
		count, concurrency)
}

// AggregateBuildExitCode returns the exit code for a parallel build session.
// Returns 0 when all workspaces succeeded (failed == 0), and 1 when any
// workspace failed. This follows standard POSIX exit code conventions.
func AggregateBuildExitCode(succeeded, failed int) int {
	if failed > 0 {
		return 1
	}
	return 0
}

// FormatBuildSessionID formats the session ID output for detach/background
// builds. The output includes the session identifier and a hint to monitor
// progress via 'dvm build status'.
//
// Example output:
//
//	Build session started: abc123
//	Monitor with: dvm build status
func FormatBuildSessionID(id string) string {
	return fmt.Sprintf("Build session started: %s\nMonitor with: dvm build status", id)
}
