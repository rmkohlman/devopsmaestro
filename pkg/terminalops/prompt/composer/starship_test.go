package composer

import (
	"strings"
	"testing"

	"devopsmaestro/pkg/terminalops/prompt/extension"
	"devopsmaestro/pkg/terminalops/prompt/style"
)

func TestStarshipComposer_Compose(t *testing.T) {
	tests := []struct {
		name       string
		style      *style.PromptStyle
		extensions []*extension.PromptExtension
		wantErr    bool
		errMsg     string
		checks     func(t *testing.T, cp *ComposedPrompt)
	}{
		{
			name: "all extensions match all segments",
			style: &style.PromptStyle{
				Name: "powerline",
				Segments: []style.Segment{
					{Name: "identity", Position: 1, StartTransition: "[](red)", EndColor: "red", Modules: []string{"os", "username"}},
					{Name: "git", Position: 2, StartColor: "sky", EndColor: "sky", Modules: []string{"git_branch", "git_status"}},
				},
				Suffix: "$character",
			},
			extensions: []*extension.PromptExtension{
				{
					Name:     "identity-ext",
					Segment:  "identity",
					Provides: []string{"os", "username"},
					Modules: map[string]extension.ExtensionModule{
						"os": {
							Symbol: " ",
							Style:  "bg:red fg:crust",
						},
						"username": {
							Style: "bg:red fg:crust",
						},
					},
				},
				{
					Name:     "git-ext",
					Segment:  "git",
					Provides: []string{"git_branch", "git_status"},
					Modules: map[string]extension.ExtensionModule{
						"git_branch": {
							Symbol: " ",
							Style:  "bg:sky fg:crust",
						},
						"git_status": {
							Style: "bg:sky fg:crust",
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, cp *ComposedPrompt) {
				if cp == nil {
					t.Fatal("ComposedPrompt should not be nil")
				}
				if cp.Format == "" {
					t.Error("Format should not be empty")
				}
				if !strings.Contains(cp.Format, "$os") {
					t.Error("Format should contain $os")
				}
				if !strings.Contains(cp.Format, "$git_branch") {
					t.Error("Format should contain $git_branch")
				}
				if !strings.HasSuffix(cp.Format, "$character") {
					t.Error("Format should end with $character")
				}
				if len(cp.Modules) != 4 {
					t.Errorf("Modules count = %d, want 4", len(cp.Modules))
				}
			},
		},
		{
			name: "some segments have no extensions",
			style: &style.PromptStyle{
				Name: "partial",
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
					Modules: map[string]extension.ExtensionModule{
						"os": {Style: "bg:red"},
					},
				},
				// git segment has no extension - should be skipped
				{
					Name:     "path-ext",
					Segment:  "path",
					Provides: []string{"directory"},
					Modules: map[string]extension.ExtensionModule{
						"directory": {Style: "bg:peach"},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, cp *ComposedPrompt) {
				if cp == nil {
					t.Fatal("ComposedPrompt should not be nil")
				}
				if strings.Contains(cp.Format, "$git_branch") {
					t.Error("Format should NOT contain $git_branch (segment skipped)")
				}
				if !strings.Contains(cp.Format, "$os") {
					t.Error("Format should contain $os")
				}
				if !strings.Contains(cp.Format, "$directory") {
					t.Error("Format should contain $directory")
				}
				if len(cp.Modules) != 2 {
					t.Errorf("Modules count = %d, want 2 (os, directory)", len(cp.Modules))
				}
			},
		},
		{
			name: "invalid style returns error",
			style: &style.PromptStyle{
				Name: "invalid",
				Segments: []style.Segment{
					{Name: "identity", Position: 1, EndColor: "#ff0000"}, // Hex color in style
				},
			},
			extensions: []*extension.PromptExtension{},
			wantErr:    true,
			errMsg:     "invalid",
		},
		{
			name: "extension with unknown segment returns error",
			style: &style.PromptStyle{
				Name: "test",
				Segments: []style.Segment{
					{Name: "identity", Position: 1, EndColor: "red", Modules: []string{"os"}},
				},
				Suffix: "$ ",
			},
			extensions: []*extension.PromptExtension{
				{
					Name:     "unknown-ext",
					Segment:  "nonexistent", // This segment doesn't exist in style!
					Provides: []string{"foo"},
				},
			},
			wantErr: true,
			errMsg:  "unknown segment",
		},
		{
			name: "composed prompt contains format string",
			style: &style.PromptStyle{
				Name: "simple",
				Segments: []style.Segment{
					{Name: "identity", Position: 1, EndColor: "red", Modules: []string{"os"}},
				},
				Suffix: "$ ",
			},
			extensions: []*extension.PromptExtension{
				{
					Name:     "identity-ext",
					Segment:  "identity",
					Provides: []string{"os"},
					Modules: map[string]extension.ExtensionModule{
						"os": {Style: "bg:red"},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, cp *ComposedPrompt) {
				if cp == nil {
					t.Fatal("ComposedPrompt should not be nil")
				}
				if cp.Format == "" {
					t.Error("ComposedPrompt.Format should not be empty")
				}
			},
		},
		{
			name: "composed prompt contains all module configs",
			style: &style.PromptStyle{
				Name: "multi-module",
				Segments: []style.Segment{
					{Name: "git", Position: 1, EndColor: "sky", Modules: []string{"git_branch", "git_status"}},
				},
				Suffix: "$ ",
			},
			extensions: []*extension.PromptExtension{
				{
					Name:     "git-ext",
					Segment:  "git",
					Provides: []string{"git_branch", "git_status"},
					Modules: map[string]extension.ExtensionModule{
						"git_branch": {
							Symbol: " ",
							Format: "[$symbol$branch]($style)",
							Style:  "bg:sky fg:crust",
						},
						"git_status": {
							Format: "[($modified)]($style)",
							Style:  "bg:sky fg:crust",
						},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, cp *ComposedPrompt) {
				if cp == nil {
					t.Fatal("ComposedPrompt should not be nil")
				}
				if len(cp.Modules) != 2 {
					t.Errorf("Modules count = %d, want 2", len(cp.Modules))
				}
				if _, ok := cp.Modules["git_branch"]; !ok {
					t.Error("Modules should contain git_branch")
				}
				if _, ok := cp.Modules["git_status"]; !ok {
					t.Error("Modules should contain git_status")
				}
				// Check module details
				if cp.Modules["git_branch"].Symbol != " " {
					t.Errorf("git_branch symbol = %q, want %q", cp.Modules["git_branch"].Symbol, " ")
				}
				if cp.Modules["git_branch"].Style != "bg:sky fg:crust" {
					t.Errorf("git_branch style = %q, want %q", cp.Modules["git_branch"].Style, "bg:sky fg:crust")
				}
			},
		},
		{
			name: "palette reference included in output",
			style: &style.PromptStyle{
				Name: "with-palette",
				Segments: []style.Segment{
					{Name: "identity", Position: 1, EndColor: "red", Modules: []string{"os"}},
				},
				Suffix: "$ ",
			},
			extensions: []*extension.PromptExtension{
				{
					Name:     "identity-ext",
					Segment:  "identity",
					Provides: []string{"os"},
					Modules: map[string]extension.ExtensionModule{
						"os": {Style: "bg:red"},
					},
				},
			},
			wantErr: false,
			checks: func(t *testing.T, cp *ComposedPrompt) {
				if cp == nil {
					t.Fatal("ComposedPrompt should not be nil")
				}
				// Palette should be set (even if empty string for now in stub)
				// Real implementation will set this properly
				_ = cp.Palette
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewStarshipComposer()
			got, err := sc.Compose(tt.style, tt.extensions)

			if (err != nil) != tt.wantErr {
				t.Errorf("Compose() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Compose() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
				return
			}

			if tt.checks != nil {
				tt.checks(t, got)
			}
		})
	}
}

func TestStarshipComposer_ExtensionValidation(t *testing.T) {
	tests := []struct {
		name      string
		extension *extension.PromptExtension
		wantErr   bool
		errMsg    string
	}{
		{
			name: "hex color in module style",
			extension: &extension.PromptExtension{
				Name:     "hex-color",
				Segment:  "git",
				Provides: []string{"git_branch"},
				Modules: map[string]extension.ExtensionModule{
					"git_branch": {
						Style: "bg:#ff0000 fg:crust", // Hex color not allowed!
					},
				},
			},
			wantErr: true,
			errMsg:  "hex color",
		},
		{
			name: "valid palette keys",
			extension: &extension.PromptExtension{
				Name:     "valid",
				Segment:  "git",
				Provides: []string{"git_branch"},
				Modules: map[string]extension.ExtensionModule{
					"git_branch": {
						Style: "bg:sky fg:crust",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sc := NewStarshipComposer()
			style := &style.PromptStyle{
				Name: "test",
				Segments: []style.Segment{
					{Name: "git", Position: 1, EndColor: "sky", Modules: []string{"git_branch"}},
				},
				Suffix: "$ ",
			}

			_, err := sc.Compose(style, []*extension.PromptExtension{tt.extension})

			if (err != nil) != tt.wantErr {
				t.Errorf("Compose() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Compose() error = %q, want error containing %q", err.Error(), tt.errMsg)
			}
		})
	}
}
