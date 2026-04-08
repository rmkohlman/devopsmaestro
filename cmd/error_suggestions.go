package cmd

import (
	"fmt"
	"strings"
)

// ErrorWithSuggestion wraps an error message with one or more actionable suggestions.
// The suggestions are appended after a blank line, each prefixed with "  → ".
// Example output:
//
//	no active app context
//
//	Suggestions:
//	  → Set active app: dvm use app <name>
//	  → List available apps: dvm get apps
func ErrorWithSuggestion(msg string, suggestions ...string) error {
	if len(suggestions) == 0 {
		return fmt.Errorf("%s", msg)
	}
	var b strings.Builder
	b.WriteString(msg)
	b.WriteString("\n\nSuggestions:")
	for _, s := range suggestions {
		b.WriteString("\n  → ")
		b.WriteString(s)
	}
	return fmt.Errorf("%s", b.String())
}

// FormatSuggestions formats suggestions as a multi-line string suitable for
// render.Info() or render.Plain(). It does NOT create an error — use this
// when you want to display suggestions before returning an existing error.
func FormatSuggestions(suggestions ...string) string {
	if len(suggestions) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("Suggestions:")
	for _, s := range suggestions {
		b.WriteString("\n  → ")
		b.WriteString(s)
	}
	return b.String()
}

// --- Predefined suggestion sets for the most common error scenarios ---

// SuggestNoActiveApp returns suggestions for when no active app context is set.
func SuggestNoActiveApp() []string {
	return []string{
		"Set active app: dvm use app <name>",
		"List available apps: dvm get apps",
		"Or use the --app / -a flag",
	}
}

// SuggestNoActiveWorkspace returns suggestions for when no active workspace context is set.
func SuggestNoActiveWorkspace() []string {
	return []string{
		"Set active workspace: dvm use workspace <name>",
		"List available workspaces: dvm get workspaces",
		"Or use the --workspace / -w flag",
	}
}

// SuggestNoActiveEcosystem returns suggestions for when no active ecosystem context is set.
func SuggestNoActiveEcosystem() []string {
	return []string{
		"Set active ecosystem: dvm use ecosystem <name>",
		"List available ecosystems: dvm get ecosystems",
		"Or use the --ecosystem / -e flag",
	}
}

// SuggestNoActiveDomain returns suggestions for when no active domain context is set.
func SuggestNoActiveDomain() []string {
	return []string{
		"Set active domain: dvm use domain <name>",
		"List available domains: dvm get domains",
		"Or use the --domain / -d flag",
	}
}

// SuggestResourceNotFound returns suggestions for when a named resource is not found.
func SuggestResourceNotFound(resourceType, name, listCmd string) []string {
	return []string{
		fmt.Sprintf("Check the spelling of %q", name),
		fmt.Sprintf("List available %ss: %s", resourceType, listCmd),
	}
}

// SuggestWorkspaceNotFound returns suggestions for when a workspace is not found.
func SuggestWorkspaceNotFound(name string) []string {
	return SuggestResourceNotFound("workspace", name, "dvm get workspaces")
}

// SuggestAppNotFound returns suggestions for when an app is not found.
func SuggestAppNotFound(name string) []string {
	return SuggestResourceNotFound("app", name, "dvm get apps")
}

// SuggestEcosystemNotFound returns suggestions for when an ecosystem is not found.
func SuggestEcosystemNotFound(name string) []string {
	return SuggestResourceNotFound("ecosystem", name, "dvm get ecosystems")
}

// SuggestDomainNotFound returns suggestions for when a domain is not found.
func SuggestDomainNotFound(name string) []string {
	return SuggestResourceNotFound("domain", name, "dvm get domains")
}

// SuggestNoContainerRuntime returns suggestions for when no container runtime is found.
func SuggestNoContainerRuntime() []string {
	return []string{
		"Install a container runtime: OrbStack, Docker Desktop, or Colima",
		"Ensure the runtime daemon is running",
	}
}

// SuggestWorkspaceNotBuilt returns suggestions for when a workspace image hasn't been built.
func SuggestWorkspaceNotBuilt() []string {
	return []string{
		"Build the workspace first: dvm build",
		"Or build a specific workspace: dvm build -a <app> -w <workspace>",
	}
}

// SuggestAmbiguousWorkspace returns suggestions for ambiguous workspace matches.
func SuggestAmbiguousWorkspace() []string {
	return []string{
		"Narrow your selection with additional flags: -e, -d, -a, -w",
		"List all workspaces: dvm get workspaces --all",
	}
}
