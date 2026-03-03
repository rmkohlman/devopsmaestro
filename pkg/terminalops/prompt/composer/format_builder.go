// Package composer provides format string building for prompts.
package composer

import (
	"fmt"
	"strings"

	"devopsmaestro/pkg/terminalops/prompt/extension"
	"devopsmaestro/pkg/terminalops/prompt/style"
)

// FormatBuilder builds starship format strings from styles and extensions.
type FormatBuilder struct {
	tracker  *TransitionTracker
	segments []string // Accumulated format parts
}

// NewFormatBuilder creates a new FormatBuilder.
func NewFormatBuilder() *FormatBuilder {
	return &FormatBuilder{
		tracker:  NewTransitionTracker(),
		segments: make([]string, 0),
	}
}

// Build builds a format string from a style and extensions.
// Returns the format string with segments and suffix.
func (fb *FormatBuilder) Build(promptStyle *style.PromptStyle, extensions []*extension.PromptExtension) string {
	// Reset for fresh build
	fb.segments = make([]string, 0)
	fb.tracker = NewTransitionTracker()

	// Create extension map for quick lookup by segment name
	extensionMap := make(map[string]*extension.PromptExtension)
	for _, ext := range extensions {
		extensionMap[ext.Segment] = ext
	}

	// Get segments in order
	orderedSegments := promptStyle.GetSegments()

	// Process each segment
	for _, segment := range orderedSegments {
		ext, hasExtension := extensionMap[segment.Name]

		if hasExtension {
			// Segment has extensions - add it to the format
			fb.AddSegment(&segment, ext.Provides)
		} else {
			// Segment has no extensions - skip it
			fb.SkipSegment(&segment)
		}
	}

	// Build final string with suffix
	return fb.BuildString(promptStyle.Suffix)
}

// AddSegment adds a segment to the format string.
func (fb *FormatBuilder) AddSegment(segment *style.Segment, moduleRefs []string) {
	var segmentParts []string

	// Use StartTransition if provided, otherwise build transition
	var transition string
	if segment.StartTransition != "" {
		transition = segment.StartTransition
		// Still update tracker with segment's end color
		fb.tracker.SetLastEndColor(segment.EndColor)
	} else {
		// Build transition using tracker
		transition = fb.tracker.BuildTransition(segment.StartColor)
		// Update tracker with segment's end color
		fb.tracker.SetLastEndColor(segment.EndColor)
	}
	segmentParts = append(segmentParts, transition)

	// Add module references
	for _, moduleRef := range moduleRefs {
		// Format as $module_name or ${custom.module_name}
		if strings.Contains(moduleRef, ".") {
			segmentParts = append(segmentParts, fmt.Sprintf("${%s}", moduleRef))
		} else {
			segmentParts = append(segmentParts, fmt.Sprintf("$%s", moduleRef))
		}
	}

	// Join parts and add to segments
	fb.segments = append(fb.segments, strings.Join(segmentParts, ""))
}

// SkipSegment skips a segment without adding it to the format.
func (fb *FormatBuilder) SkipSegment(segment *style.Segment) {
	// Call tracker's Skip method - does NOT update lastEndColor
	fb.tracker.Skip(segment)

	// Do NOT add anything to fb.segments
}

// BuildString returns the final format string with suffix.
func (fb *FormatBuilder) BuildString(suffix string) string {
	// Join all segment strings
	formatString := strings.Join(fb.segments, "")

	// Append suffix at the end
	formatString += suffix

	return formatString
}
