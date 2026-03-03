package style

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPromptStyleYAML_Parse(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		checks  func(t *testing.T, psy *PromptStyleYAML)
	}{
		{
			name: "valid powerline style",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: PromptStyle
metadata:
  name: powerline-segments
  description: Powerline style with segment transitions
spec:
  segments:
    - name: identity
      position: 1
      startTransition: "[](red)"
      endColor: red
      modules: [os, username]
    - name: git
      position: 4
      startColor: sky
      endColor: sky
      modules: [git_branch, git_status]
  suffix: "$line_break$character"`,
			wantErr: false,
			checks: func(t *testing.T, psy *PromptStyleYAML) {
				if psy.Metadata.Name != "powerline-segments" {
					t.Errorf("Name = %q, want %q", psy.Metadata.Name, "powerline-segments")
				}
				if len(psy.Spec.Segments) != 2 {
					t.Errorf("Segments count = %d, want 2", len(psy.Spec.Segments))
				}
				if psy.Spec.Suffix != "$line_break$character" {
					t.Errorf("Suffix = %q, want %q", psy.Spec.Suffix, "$line_break$character")
				}
			},
		},
		{
			name: "minimal style",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: PromptStyle
metadata:
  name: minimal
spec:
  segments:
    - name: simple
      position: 1
      endColor: blue
      modules: [directory]
  suffix: "$ "`,
			wantErr: false,
			checks: func(t *testing.T, psy *PromptStyleYAML) {
				if psy.Metadata.Name != "minimal" {
					t.Errorf("Name = %q, want %q", psy.Metadata.Name, "minimal")
				}
				if len(psy.Spec.Segments) != 1 {
					t.Errorf("Segments count = %d, want 1", len(psy.Spec.Segments))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var psy PromptStyleYAML
			err := yaml.Unmarshal([]byte(tt.yaml), &psy)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checks != nil {
				tt.checks(t, &psy)
			}
		})
	}
}

func TestPromptStyle_GetSegments(t *testing.T) {
	tests := []struct {
		name      string
		style     *PromptStyle
		wantLen   int
		wantOrder []int // Expected position order
	}{
		{
			name: "segments in order",
			style: &PromptStyle{
				Name: "test",
				Segments: []Segment{
					{Name: "first", Position: 1},
					{Name: "second", Position: 2},
					{Name: "third", Position: 3},
				},
			},
			wantLen:   3,
			wantOrder: []int{1, 2, 3},
		},
		{
			name: "segments out of order",
			style: &PromptStyle{
				Name: "test",
				Segments: []Segment{
					{Name: "third", Position: 3},
					{Name: "first", Position: 1},
					{Name: "second", Position: 2},
				},
			},
			wantLen:   3,
			wantOrder: []int{1, 2, 3},
		},
		{
			name: "segments with gaps",
			style: &PromptStyle{
				Name: "test",
				Segments: []Segment{
					{Name: "first", Position: 1},
					{Name: "third", Position: 5},
					{Name: "second", Position: 3},
				},
			},
			wantLen:   3,
			wantOrder: []int{1, 3, 5},
		},
		{
			name: "empty segments",
			style: &PromptStyle{
				Name:     "test",
				Segments: []Segment{},
			},
			wantLen:   0,
			wantOrder: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.style.GetSegments()

			if len(got) != tt.wantLen {
				t.Errorf("GetSegments() length = %d, want %d", len(got), tt.wantLen)
				return
			}

			for i, seg := range got {
				if seg.Position != tt.wantOrder[i] {
					t.Errorf("GetSegments()[%d].Position = %d, want %d", i, seg.Position, tt.wantOrder[i])
				}
			}
		})
	}
}

func TestPromptStyleYAML_Validate(t *testing.T) {
	tests := []struct {
		name    string
		style   *PromptStyleYAML
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid style",
			style: &PromptStyleYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptStyle,
				Metadata: PromptStyleMetadata{
					Name: "valid-style",
				},
				Spec: PromptStyleSpec{
					Segments: []Segment{
						{Name: "identity", Position: 1, EndColor: "red"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			style: &PromptStyleYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptStyle,
				Metadata: PromptStyleMetadata{
					Name: "",
				},
				Spec: PromptStyleSpec{
					Segments: []Segment{
						{Name: "identity", Position: 1},
					},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "duplicate positions",
			style: &PromptStyleYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptStyle,
				Metadata: PromptStyleMetadata{
					Name: "duplicate-positions",
				},
				Spec: PromptStyleSpec{
					Segments: []Segment{
						{Name: "first", Position: 1},
						{Name: "second", Position: 1}, // Duplicate!
					},
				},
			},
			wantErr: true,
			errMsg:  "duplicate position",
		},
		{
			name: "hex color in startColor",
			style: &PromptStyleYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptStyle,
				Metadata: PromptStyleMetadata{
					Name: "hex-color",
				},
				Spec: PromptStyleSpec{
					Segments: []Segment{
						{Name: "identity", Position: 1, StartColor: "#ff0000"}, // Hex color not allowed
					},
				},
			},
			wantErr: true,
			errMsg:  "must be palette key",
		},
		{
			name: "hex color in endColor",
			style: &PromptStyleYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptStyle,
				Metadata: PromptStyleMetadata{
					Name: "hex-end-color",
				},
				Spec: PromptStyleSpec{
					Segments: []Segment{
						{Name: "identity", Position: 1, EndColor: "#00ff00"}, // Hex color not allowed
					},
				},
			},
			wantErr: true,
			errMsg:  "must be palette key",
		},
		{
			name: "valid palette key colors",
			style: &PromptStyleYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptStyle,
				Metadata: PromptStyleMetadata{
					Name: "valid-colors",
				},
				Spec: PromptStyleSpec{
					Segments: []Segment{
						{Name: "identity", Position: 1, StartColor: "red", EndColor: "blue"},
						{Name: "git", Position: 2, StartColor: "sky", EndColor: "peach"},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.style.Validate()

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %q, want error containing %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestPromptStyleYAML_ToPromptStyle(t *testing.T) {
	tests := []struct {
		name   string
		yaml   *PromptStyleYAML
		checks func(t *testing.T, ps *PromptStyle)
	}{
		{
			name: "basic conversion",
			yaml: &PromptStyleYAML{
				Metadata: PromptStyleMetadata{
					Name:        "test-style",
					Description: "A test style",
				},
				Spec: PromptStyleSpec{
					Segments: []Segment{
						{Name: "identity", Position: 1},
					},
					Suffix: "$ ",
				},
			},
			checks: func(t *testing.T, ps *PromptStyle) {
				if ps.Name != "test-style" {
					t.Errorf("Name = %q, want %q", ps.Name, "test-style")
				}
				if ps.Description != "A test style" {
					t.Errorf("Description = %q, want %q", ps.Description, "A test style")
				}
				if ps.Suffix != "$ " {
					t.Errorf("Suffix = %q, want %q", ps.Suffix, "$ ")
				}
				if len(ps.Segments) != 1 {
					t.Errorf("Segments length = %d, want 1", len(ps.Segments))
				}
			},
		},
		{
			name: "enabled defaults to true",
			yaml: &PromptStyleYAML{
				Metadata: PromptStyleMetadata{
					Name: "enabled-default",
				},
				Spec: PromptStyleSpec{},
			},
			checks: func(t *testing.T, ps *PromptStyle) {
				if !ps.Enabled {
					t.Errorf("Enabled = false, want true")
				}
			},
		},
		{
			name: "enabled explicitly false",
			yaml: &PromptStyleYAML{
				Metadata: PromptStyleMetadata{
					Name: "disabled",
				},
				Spec: PromptStyleSpec{
					Enabled: boolPtr(false),
				},
			},
			checks: func(t *testing.T, ps *PromptStyle) {
				if ps.Enabled {
					t.Errorf("Enabled = true, want false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := tt.yaml.ToPromptStyle()
			tt.checks(t, ps)
		})
	}
}

func TestPromptStyleYAML_ResourceInterface(t *testing.T) {
	psy := NewPromptStyleYAML("test-style")

	if psy.GetKind() != KindPromptStyle {
		t.Errorf("GetKind() = %q, want %q", psy.GetKind(), KindPromptStyle)
	}

	if psy.GetName() != "test-style" {
		t.Errorf("GetName() = %q, want %q", psy.GetName(), "test-style")
	}

	if psy.GetAPIVersion() != "devopsmaestro.io/v1" {
		t.Errorf("GetAPIVersion() = %q, want %q", psy.GetAPIVersion(), "devopsmaestro.io/v1")
	}
}

// Helper function for bool pointers
func boolPtr(b bool) *bool {
	return &b
}
