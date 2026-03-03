package composer

import (
	"strings"
	"testing"

	"devopsmaestro/pkg/terminalops/prompt/extension"
	"devopsmaestro/pkg/terminalops/prompt/style"
)

func TestFormatBuilder_Build(t *testing.T) {
	tests := []struct {
		name       string
		style      *style.PromptStyle
		extensions []*extension.PromptExtension
		wantChecks []string // Strings that should be in the format
	}{
		{
			name: "all segments included",
			style: &style.PromptStyle{
				Name: "test-style",
				Segments: []style.Segment{
					{Name: "identity", Position: 1, StartTransition: "[" + PowerlineArrow + "](red)", EndColor: "red", Modules: []string{"os", "username"}},
					{Name: "git", Position: 2, StartColor: "sky", EndColor: "sky", Modules: []string{"git_branch", "git_status"}},
				},
				Suffix: "$line_break$character",
			},
			extensions: []*extension.PromptExtension{
				{
					Name:     "identity-ext",
					Segment:  "identity",
					Provides: []string{"os", "username"},
				},
				{
					Name:     "git-ext",
					Segment:  "git",
					Provides: []string{"git_branch", "git_status"},
				},
			},
			wantChecks: []string{
				"[" + PowerlineArrow + "](red)", // First segment transition
				"$os",                           // First segment modules
				"$username",                     // First segment modules
				"[" + PowerlineArrow + "](bg:sky fg:red)", // Second segment transition
				"$git_branch",           // Second segment modules
				"$git_status",           // Second segment modules
				"$line_break$character", // Suffix
			},
		},
		{
			name: "middle segment skipped",
			style: &style.PromptStyle{
				Name: "test-style",
				Segments: []style.Segment{
					{Name: "identity", Position: 1, EndColor: "red", Modules: []string{"os"}},
					{Name: "git", Position: 2, StartColor: "sky", EndColor: "sky", Modules: []string{"git_branch"}},
					{Name: "path", Position: 3, StartColor: "peach", EndColor: "peach", Modules: []string{"directory"}},
				},
				Suffix: "$ ",
			},
			extensions: []*extension.PromptExtension{
				{
					Name:     "identity-ext",
					Segment:  "identity",
					Provides: []string{"os"},
				},
				// git segment has NO extension - should be skipped
				{
					Name:     "path-ext",
					Segment:  "path",
					Provides: []string{"directory"},
				},
			},
			wantChecks: []string{
				"$os", // First segment
				"[" + PowerlineArrow + "](bg:peach fg:red)", // Transition should skip git segment color
				"$directory", // Third segment
				"$ ",         // Suffix
			},
		},
		{
			name: "first segment skipped",
			style: &style.PromptStyle{
				Name: "test-style",
				Segments: []style.Segment{
					{Name: "identity", Position: 1, EndColor: "red", Modules: []string{"os"}},
					{Name: "git", Position: 2, StartColor: "sky", EndColor: "sky", Modules: []string{"git_branch"}},
				},
				Suffix: "$character",
			},
			extensions: []*extension.PromptExtension{
				// identity segment has NO extension - skipped
				{
					Name:     "git-ext",
					Segment:  "git",
					Provides: []string{"git_branch"},
				},
			},
			wantChecks: []string{
				"[" + PowerlineArrow + "](sky)", // First rendered segment (git) has no previous
				"$git_branch",                   // Git segment module
				"$character",                    // Suffix
			},
		},
		{
			name: "multiple consecutive segments skipped",
			style: &style.PromptStyle{
				Name: "test-style",
				Segments: []style.Segment{
					{Name: "identity", Position: 1, EndColor: "red", Modules: []string{"os"}},
					{Name: "git", Position: 2, StartColor: "sky", EndColor: "sky", Modules: []string{"git_branch"}},
					{Name: "lang", Position: 3, StartColor: "green", EndColor: "green", Modules: []string{"nodejs"}},
					{Name: "path", Position: 4, StartColor: "peach", EndColor: "peach", Modules: []string{"directory"}},
				},
				Suffix: "$ ",
			},
			extensions: []*extension.PromptExtension{
				{
					Name:     "identity-ext",
					Segment:  "identity",
					Provides: []string{"os"},
				},
				// git and lang segments skipped (no extensions)
				{
					Name:     "path-ext",
					Segment:  "path",
					Provides: []string{"directory"},
				},
			},
			wantChecks: []string{
				"$os", // First segment
				"[" + PowerlineArrow + "](bg:peach fg:red)", // Should skip git and lang colors
				"$directory", // Last segment
				"$ ",         // Suffix
			},
		},
		{
			name: "format includes suffix at end",
			style: &style.PromptStyle{
				Name: "test-style",
				Segments: []style.Segment{
					{Name: "identity", Position: 1, EndColor: "red", Modules: []string{"os"}},
				},
				Suffix: "$line_break$character",
			},
			extensions: []*extension.PromptExtension{
				{
					Name:     "identity-ext",
					Segment:  "identity",
					Provides: []string{"os"},
				},
			},
			wantChecks: []string{
				"$os",
				"$line_break$character", // Suffix at the very end
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fb := NewFormatBuilder()
			got := fb.Build(tt.style, tt.extensions)

			for _, check := range tt.wantChecks {
				if !strings.Contains(got, check) {
					t.Errorf("Build() missing expected string %q\nGot: %s", check, got)
				}
			}

			// Verify suffix is at the end
			if tt.style.Suffix != "" && !strings.HasSuffix(got, tt.style.Suffix) {
				t.Errorf("Build() does not end with suffix %q\nGot: %s", tt.style.Suffix, got)
			}
		})
	}
}

func TestFormatBuilder_ModuleReferences(t *testing.T) {
	style := &style.PromptStyle{
		Name: "test",
		Segments: []style.Segment{
			{
				Name:     "git",
				Position: 1,
				EndColor: "sky",
				Modules:  []string{"git_branch", "git_status", "git_commit"},
			},
		},
		Suffix: "$ ",
	}

	extensions := []*extension.PromptExtension{
		{
			Name:     "git-ext",
			Segment:  "git",
			Provides: []string{"git_branch", "git_status", "git_commit"},
		},
	}

	fb := NewFormatBuilder()
	got := fb.Build(style, extensions)

	// All modules should be referenced with $ prefix
	expectedModules := []string{"$git_branch", "$git_status", "$git_commit"}
	for _, module := range expectedModules {
		if !strings.Contains(got, module) {
			t.Errorf("Build() missing module reference %q\nGot: %s", module, got)
		}
	}
}

func TestFormatBuilder_EmptyStyle(t *testing.T) {
	style := &style.PromptStyle{
		Name:     "empty",
		Segments: []style.Segment{},
		Suffix:   "$ ",
	}

	extensions := []*extension.PromptExtension{}

	fb := NewFormatBuilder()
	got := fb.Build(style, extensions)

	// Should just have the suffix
	if got != "$ " {
		t.Errorf("Build() with empty style = %q, want %q", got, "$ ")
	}
}

func TestFormatBuilder_AllSegmentsSkipped(t *testing.T) {
	style := &style.PromptStyle{
		Name: "all-skipped",
		Segments: []style.Segment{
			{Name: "identity", Position: 1, EndColor: "red", Modules: []string{"os"}},
			{Name: "git", Position: 2, StartColor: "sky", EndColor: "sky", Modules: []string{"git_branch"}},
		},
		Suffix: "$ ",
	}

	extensions := []*extension.PromptExtension{
		// No extensions match any segments
	}

	fb := NewFormatBuilder()
	got := fb.Build(style, extensions)

	// Should just have the suffix
	if got != "$ " {
		t.Errorf("Build() with all segments skipped = %q, want %q", got, "$ ")
	}
}
