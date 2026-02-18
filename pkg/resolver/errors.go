package resolver

import (
	"fmt"
	"strings"

	"devopsmaestro/models"
)

// ErrNoWorkspaceFound is returned when no workspace matches the given criteria.
var ErrNoWorkspaceFound = fmt.Errorf("no workspace found matching the given criteria")

// AmbiguousError is returned when multiple workspaces match the given criteria.
// It contains the list of matching workspaces so the user can be shown disambiguation options.
type AmbiguousError struct {
	// Matches contains all workspaces that matched the criteria.
	Matches []*models.WorkspaceWithHierarchy

	// Message is a human-readable error message.
	Message string
}

// Error implements the error interface.
func (e *AmbiguousError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("ambiguous: %d workspaces match the given criteria", len(e.Matches))
}

// FormatDisambiguation returns a formatted string showing all matching workspaces
// for display to the user. This helps them understand which flags to add to be more specific.
func (e *AmbiguousError) FormatDisambiguation() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Multiple workspaces (%d) match your criteria:\n\n", len(e.Matches)))

	for i, wh := range e.Matches {
		sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, wh.FullPath()))
		sb.WriteString(fmt.Sprintf("     App: %s, Domain: %s, Ecosystem: %s\n",
			wh.App.Name, wh.Domain.Name, wh.Ecosystem.Name))
	}

	sb.WriteString("\nUse additional flags to narrow your selection:\n")
	sb.WriteString("  -e <ecosystem>  Filter by ecosystem\n")
	sb.WriteString("  -d <domain>     Filter by domain\n")
	sb.WriteString("  -a <app>        Filter by app\n")
	sb.WriteString("  -w <workspace>  Filter by workspace name\n")

	return sb.String()
}

// NewAmbiguousError creates a new AmbiguousError with the given matches.
func NewAmbiguousError(matches []*models.WorkspaceWithHierarchy) *AmbiguousError {
	return &AmbiguousError{
		Matches: matches,
		Message: fmt.Sprintf("ambiguous: %d workspaces match the given criteria", len(matches)),
	}
}

// IsAmbiguousError checks if an error is an AmbiguousError and returns it if so.
func IsAmbiguousError(err error) (*AmbiguousError, bool) {
	if ae, ok := err.(*AmbiguousError); ok {
		return ae, true
	}
	return nil, false
}

// IsNoWorkspaceFoundError checks if an error is the ErrNoWorkspaceFound error.
func IsNoWorkspaceFoundError(err error) bool {
	return err == ErrNoWorkspaceFound
}
