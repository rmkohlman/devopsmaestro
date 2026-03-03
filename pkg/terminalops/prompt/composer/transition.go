// Package composer provides transition tracking for prompt segments.
package composer

import (
	"fmt"

	"devopsmaestro/pkg/terminalops/prompt/style"
)

// Powerline arrow character (U+E0B0)
const PowerlineArrow = "\ue0b0"

// TransitionTracker tracks color transitions between prompt segments.
type TransitionTracker struct {
	lastEndColor string
}

// NewTransitionTracker creates a new TransitionTracker.
func NewTransitionTracker() *TransitionTracker {
	return &TransitionTracker{}
}

// BuildTransition builds a transition string for the current segment.
// For the first segment with no previous color, it returns "[](color)".
// For subsequent segments, it returns "[](bg:newColor fg:prevColor)".
func (tt *TransitionTracker) BuildTransition(newStartColor string) string {
	var transition string

	if tt.lastEndColor == "" {
		// First segment - no previous color
		transition = fmt.Sprintf("[%s](%s)", PowerlineArrow, newStartColor)
	} else {
		// Subsequent segments - include both colors
		transition = fmt.Sprintf("[%s](bg:%s fg:%s)", PowerlineArrow, newStartColor, tt.lastEndColor)
	}

	// Update lastEndColor to newStartColor for the next transition
	tt.lastEndColor = newStartColor

	return transition
}

// Skip advances the tracker without rendering a transition.
// This is used when a segment is skipped (no extensions match).
// CRITICAL: Does NOT update lastEndColor - this preserves correct chaining.
func (tt *TransitionTracker) Skip(segment interface{}) {
	// Skip does not change lastEndColor
	// This ensures the next transition uses the color from before the skip

	// Type assertion to extract segment (for potential future use)
	if _, ok := segment.(*style.Segment); ok {
		// Segment is available if needed in the future
		// For now, we just don't update lastEndColor
	}
}

// SetLastEndColor sets the last end color for testing purposes.
func (tt *TransitionTracker) SetLastEndColor(color string) {
	tt.lastEndColor = color
}

// GetLastEndColor returns the last end color.
func (tt *TransitionTracker) GetLastEndColor() string {
	return tt.lastEndColor
}
