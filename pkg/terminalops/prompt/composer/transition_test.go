package composer

import (
	"testing"

	"devopsmaestro/pkg/terminalops/prompt/style"
)

func TestTransitionTracker_BuildTransition(t *testing.T) {
	tests := []struct {
		name          string
		prevEndColor  string
		newStartColor string
		want          string
	}{
		{
			name:          "first segment no previous",
			prevEndColor:  "",
			newStartColor: "red",
			want:          "[" + PowerlineArrow + "](red)",
		},
		{
			name:          "standard transition",
			prevEndColor:  "red",
			newStartColor: "sky",
			want:          "[" + PowerlineArrow + "](bg:sky fg:red)",
		},
		{
			name:          "transition from sky to peach",
			prevEndColor:  "sky",
			newStartColor: "peach",
			want:          "[" + PowerlineArrow + "](bg:peach fg:sky)",
		},
		{
			name:          "transition from peach to blue",
			prevEndColor:  "peach",
			newStartColor: "blue",
			want:          "[" + PowerlineArrow + "](bg:blue fg:peach)",
		},
		{
			name:          "same color transition",
			prevEndColor:  "red",
			newStartColor: "red",
			want:          "[" + PowerlineArrow + "](bg:red fg:red)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTransitionTracker()
			if tt.prevEndColor != "" {
				tracker.SetLastEndColor(tt.prevEndColor)
			}
			got := tracker.BuildTransition(tt.newStartColor)
			if got != tt.want {
				t.Errorf("BuildTransition() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTransitionTracker_Skip(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(tt *TransitionTracker)
		segmentToSkip  *style.Segment
		nextStartColor string
		wantTransition string
	}{
		{
			name: "skip middle segment uses pre-skip endColor",
			setup: func(tt *TransitionTracker) {
				// First segment: red -> red
				tt.SetLastEndColor("red")
			},
			segmentToSkip: &style.Segment{
				Name:       "skipped",
				Position:   2,
				StartColor: "sky",
				EndColor:   "sky",
			},
			nextStartColor: "peach",
			// Should use "red" (before skip), not "sky" (the skipped segment)
			wantTransition: "[" + PowerlineArrow + "](bg:peach fg:red)",
		},
		{
			name: "skip after first segment",
			setup: func(tt *TransitionTracker) {
				tt.SetLastEndColor("blue")
			},
			segmentToSkip: &style.Segment{
				Name:       "skipped",
				Position:   2,
				StartColor: "green",
				EndColor:   "green",
			},
			nextStartColor: "yellow",
			wantTransition: "[" + PowerlineArrow + "](bg:yellow fg:blue)",
		},
		{
			name: "consecutive skips",
			setup: func(tt *TransitionTracker) {
				tt.SetLastEndColor("red")
			},
			segmentToSkip: &style.Segment{
				Name:       "skip1",
				Position:   2,
				StartColor: "sky",
				EndColor:   "sky",
			},
			// We'll test by skipping multiple then building
			nextStartColor: "peach",
			wantTransition: "[" + PowerlineArrow + "](bg:peach fg:red)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTransitionTracker()
			tt.setup(tracker)

			// Skip the segment
			tracker.Skip(tt.segmentToSkip)

			// Build transition to next segment
			got := tracker.BuildTransition(tt.nextStartColor)
			if got != tt.wantTransition {
				t.Errorf("After Skip(), BuildTransition() = %q, want %q", got, tt.wantTransition)
			}
		})
	}
}

func TestTransitionTracker_ConsecutiveSkips(t *testing.T) {
	tracker := NewTransitionTracker()

	// Set initial color
	tracker.SetLastEndColor("red")

	// Skip segment 1
	tracker.Skip(&style.Segment{
		Name:       "skip1",
		Position:   2,
		StartColor: "sky",
		EndColor:   "sky",
	})

	// Skip segment 2
	tracker.Skip(&style.Segment{
		Name:       "skip2",
		Position:   3,
		StartColor: "peach",
		EndColor:   "peach",
	})

	// Build transition should still use "red" (before all skips)
	got := tracker.BuildTransition("blue")
	want := "[" + PowerlineArrow + "](bg:blue fg:red)"

	if got != want {
		t.Errorf("After consecutive skips, BuildTransition() = %q, want %q", got, want)
	}
}

func TestTransitionTracker_GetLastEndColor(t *testing.T) {
	tests := []struct {
		name          string
		operations    func(tt *TransitionTracker)
		wantLastColor string
	}{
		{
			name: "after BuildTransition",
			operations: func(tt *TransitionTracker) {
				tt.BuildTransition("red")
			},
			wantLastColor: "red",
		},
		{
			name: "after multiple BuildTransition calls",
			operations: func(tt *TransitionTracker) {
				tt.BuildTransition("red")
				tt.BuildTransition("sky")
				tt.BuildTransition("peach")
			},
			wantLastColor: "peach",
		},
		{
			name: "after Skip does not change lastEndColor",
			operations: func(tt *TransitionTracker) {
				tt.SetLastEndColor("red")
				tt.Skip(&style.Segment{
					Name:       "skipped",
					StartColor: "sky",
					EndColor:   "sky",
				})
			},
			wantLastColor: "red",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tracker := NewTransitionTracker()
			tt.operations(tracker)
			got := tracker.GetLastEndColor()
			if got != tt.wantLastColor {
				t.Errorf("GetLastEndColor() = %q, want %q", got, tt.wantLastColor)
			}
		})
	}
}

func TestTransitionTracker_SkipFirstSegment(t *testing.T) {
	tracker := NewTransitionTracker()

	// Skip the first segment (no previous color)
	tracker.Skip(&style.Segment{
		Name:       "skip-first",
		Position:   1,
		StartColor: "red",
		EndColor:   "red",
	})

	// Next transition should still work as if it's the first
	got := tracker.BuildTransition("sky")
	want := "[" + PowerlineArrow + "](sky)"

	if got != want {
		t.Errorf("After skipping first segment, BuildTransition() = %q, want %q", got, want)
	}
}
