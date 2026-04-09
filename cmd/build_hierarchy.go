package cmd

import (
	"fmt"
	"strings"
)

// FormatBuildDisambiguationError formats an error message when a name is
// ambiguous across multiple parent scopes (e.g., app name found in multiple
// domains). The parentKind is the type to disambiguate by (e.g., "domain",
// "ecosystem") and parentMatches lists the matching parent names.
//
// Example output:
//
//	error: ambiguous app name "my-app" — found in multiple domains:
//	  my-app  (domain: backend)
//	  my-app  (domain: frontend)
//	Hint: Add --domain to disambiguate: dvm build --app my-app --domain backend
func FormatBuildDisambiguationError(name, parentKind string, parentMatches []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "error: ambiguous name %q — found in multiple %ss:\n", name, parentKind)
	for _, match := range parentMatches {
		fmt.Fprintf(&b, "  %s  (%s: %s)\n", name, parentKind, match)
	}
	if len(parentMatches) > 0 {
		fmt.Fprintf(&b, "Hint: Add --%s to disambiguate: dvm build --app %s --%s %s",
			parentKind, name, parentKind, parentMatches[0])
	}
	return b.String()
}

// FormatBuildNotFoundError formats a clear error message when no workspaces
// are found for a given hierarchy flag value.
//
// Example: no app named "does-not-exist" found
func FormatBuildNotFoundError(resourceKind, name string) string {
	return fmt.Sprintf("no %s named %q found", resourceKind, name)
}

// FormatBuildDryRunTable formats a single row of the dry-run table for
// a hierarchical build. Each row contains a workspace name and app name.
func FormatBuildDryRunTable(workspace, app string) string {
	return fmt.Sprintf("%-20s %-15s %s", workspace, app, FormatBuildDryRunStatusValue())
}

// FormatBuildDryRunSummary formats the summary line printed after the dry-run
// table for a hierarchical build.
//
// Example: "Would build 3 workspace(s). Use without --dry-run to execute."
func FormatBuildDryRunSummary(count int) string {
	return fmt.Sprintf("Would build %d workspace(s). Use without --dry-run to execute.", count)
}

// validateBuildAllMutualExclusion checks that --all is not combined with
// any hierarchy scope flags (--app, --domain, --ecosystem, --workspace).
// Returns an error if mutual exclusion is violated.
func validateBuildAllMutualExclusion(allFlag bool, flags HierarchyFlags) error {
	if !allFlag {
		return nil
	}
	if flags.HasAnyFlag() {
		return fmt.Errorf("--all is mutually exclusive with --app, --domain, --ecosystem, --workspace")
	}
	return nil
}
