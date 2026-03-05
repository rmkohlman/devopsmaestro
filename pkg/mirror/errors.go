package mirror

import (
	"errors"
	"fmt"
)

// ClonePhaseError represents an error that occurred during a specific phase
// of the CloneToWorkspace operation (clone, checkout, or remote-set-url).
type ClonePhaseError struct {
	Phase   string // "clone", "checkout", "remote-set-url"
	Ref     string // git ref involved (branch, tag, commit)
	Wrapped error  // underlying error
}

// Error returns a human-readable error message that identifies the phase.
func (e *ClonePhaseError) Error() string {
	if e.Ref != "" {
		return fmt.Sprintf("git %s failed for ref '%s': %v", e.Phase, e.Ref, e.Wrapped)
	}
	return fmt.Sprintf("git %s failed: %v", e.Phase, e.Wrapped)
}

// Unwrap returns the underlying error for use with errors.Is/errors.As.
func (e *ClonePhaseError) Unwrap() error {
	return e.Wrapped
}

// IsCheckoutFailure returns true if the error is a ClonePhaseError with Phase "checkout".
func IsCheckoutFailure(err error) bool {
	var phaseErr *ClonePhaseError
	if errors.As(err, &phaseErr) {
		return phaseErr.Phase == "checkout"
	}
	return false
}

// IsCloneFailure returns true if the error is a ClonePhaseError with Phase "clone".
func IsCloneFailure(err error) bool {
	var phaseErr *ClonePhaseError
	if errors.As(err, &phaseErr) {
		return phaseErr.Phase == "clone"
	}
	return false
}
