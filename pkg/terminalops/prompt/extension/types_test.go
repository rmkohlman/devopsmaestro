package extension

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestPromptExtensionYAML_Parse(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		checks  func(t *testing.T, pey *PromptExtensionYAML)
	}{
		{
			name: "valid git extension",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: PromptExtension
metadata:
  name: git-detailed
  description: Detailed git status extension
spec:
  segment: git
  provides: [git_branch, git_status]
  modules:
    git_branch:
      symbol: " "
      format: "[$symbol$branch]($style)"
      style: "bg:sky fg:crust"
    git_status:
      format: "[($modified$staged)]($style)"
      style: "bg:sky fg:crust"`,
			wantErr: false,
			checks: func(t *testing.T, pey *PromptExtensionYAML) {
				if pey.Metadata.Name != "git-detailed" {
					t.Errorf("Name = %q, want %q", pey.Metadata.Name, "git-detailed")
				}
				if pey.Spec.Segment != "git" {
					t.Errorf("Segment = %q, want %q", pey.Spec.Segment, "git")
				}
				if len(pey.Spec.Provides) != 2 {
					t.Errorf("Provides count = %d, want 2", len(pey.Spec.Provides))
				}
				if len(pey.Spec.Modules) != 2 {
					t.Errorf("Modules count = %d, want 2", len(pey.Spec.Modules))
				}
			},
		},
		{
			name: "minimal extension",
			yaml: `apiVersion: devopsmaestro.io/v1
kind: PromptExtension
metadata:
  name: directory-simple
spec:
  segment: path
  provides: [directory]
  modules:
    directory:
      format: "[$path]($style)"
      style: "bg:blue fg:crust"`,
			wantErr: false,
			checks: func(t *testing.T, pey *PromptExtensionYAML) {
				if pey.Metadata.Name != "directory-simple" {
					t.Errorf("Name = %q, want %q", pey.Metadata.Name, "directory-simple")
				}
				if len(pey.Spec.Provides) != 1 {
					t.Errorf("Provides count = %d, want 1", len(pey.Spec.Provides))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var pey PromptExtensionYAML
			err := yaml.Unmarshal([]byte(tt.yaml), &pey)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checks != nil {
				tt.checks(t, &pey)
			}
		})
	}
}

func TestPromptExtension_GetProvides(t *testing.T) {
	tests := []struct {
		name      string
		extension *PromptExtension
		want      []string
	}{
		{
			name: "multiple modules",
			extension: &PromptExtension{
				Name:     "test",
				Provides: []string{"git_branch", "git_status"},
			},
			want: []string{"git_branch", "git_status"},
		},
		{
			name: "single module",
			extension: &PromptExtension{
				Name:     "test",
				Provides: []string{"directory"},
			},
			want: []string{"directory"},
		},
		{
			name: "empty provides",
			extension: &PromptExtension{
				Name:     "test",
				Provides: []string{},
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.extension.GetProvides()

			if len(got) != len(tt.want) {
				t.Errorf("GetProvides() length = %d, want %d", len(got), len(tt.want))
				return
			}

			for i, module := range got {
				if module != tt.want[i] {
					t.Errorf("GetProvides()[%d] = %q, want %q", i, module, tt.want[i])
				}
			}
		})
	}
}

func TestPromptExtension_GetSegment(t *testing.T) {
	tests := []struct {
		name      string
		extension *PromptExtension
		want      string
	}{
		{
			name: "git segment",
			extension: &PromptExtension{
				Name:    "git-ext",
				Segment: "git",
			},
			want: "git",
		},
		{
			name: "path segment",
			extension: &PromptExtension{
				Name:    "path-ext",
				Segment: "path",
			},
			want: "path",
		},
		{
			name: "empty segment",
			extension: &PromptExtension{
				Name:    "no-segment",
				Segment: "",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.extension.GetSegment()
			if got != tt.want {
				t.Errorf("GetSegment() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPromptExtensionYAML_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ext     *PromptExtensionYAML
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid extension",
			ext: &PromptExtensionYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptExtension,
				Metadata: PromptExtensionMetadata{
					Name: "valid-ext",
				},
				Spec: PromptExtensionSpec{
					Segment:  "git",
					Provides: []string{"git_branch"},
					Modules: map[string]ExtensionModule{
						"git_branch": {
							Style: "bg:sky fg:crust",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "missing name",
			ext: &PromptExtensionYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptExtension,
				Metadata: PromptExtensionMetadata{
					Name: "",
				},
				Spec: PromptExtensionSpec{
					Segment: "git",
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "hex color in style",
			ext: &PromptExtensionYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptExtension,
				Metadata: PromptExtensionMetadata{
					Name: "hex-color",
				},
				Spec: PromptExtensionSpec{
					Segment:  "git",
					Provides: []string{"git_branch"},
					Modules: map[string]ExtensionModule{
						"git_branch": {
							Style: "bg:#ff0000 fg:#00ff00", // Hex colors not allowed!
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "must use palette keys",
		},
		{
			name: "valid palette key styles",
			ext: &PromptExtensionYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptExtension,
				Metadata: PromptExtensionMetadata{
					Name: "valid-colors",
				},
				Spec: PromptExtensionSpec{
					Segment:  "git",
					Provides: []string{"git_branch", "git_status"},
					Modules: map[string]ExtensionModule{
						"git_branch": {
							Style: "bg:sky fg:crust",
						},
						"git_status": {
							Style: "bg:peach fg:base",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "partial hex color detection",
			ext: &PromptExtensionYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       KindPromptExtension,
				Metadata: PromptExtensionMetadata{
					Name: "partial-hex",
				},
				Spec: PromptExtensionSpec{
					Segment:  "git",
					Provides: []string{"git_branch"},
					Modules: map[string]ExtensionModule{
						"git_branch": {
							Style: "bg:sky fg:#123456", // One hex color
						},
					},
				},
			},
			wantErr: true,
			errMsg:  "must use palette keys",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ext.Validate()

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

func TestPromptExtensionYAML_ToPromptExtension(t *testing.T) {
	tests := []struct {
		name   string
		yaml   *PromptExtensionYAML
		checks func(t *testing.T, pe *PromptExtension)
	}{
		{
			name: "basic conversion",
			yaml: &PromptExtensionYAML{
				Metadata: PromptExtensionMetadata{
					Name:        "test-ext",
					Description: "A test extension",
				},
				Spec: PromptExtensionSpec{
					Segment:  "git",
					Provides: []string{"git_branch"},
					Modules: map[string]ExtensionModule{
						"git_branch": {
							Symbol: " ",
							Style:  "bg:sky",
						},
					},
				},
			},
			checks: func(t *testing.T, pe *PromptExtension) {
				if pe.Name != "test-ext" {
					t.Errorf("Name = %q, want %q", pe.Name, "test-ext")
				}
				if pe.Description != "A test extension" {
					t.Errorf("Description = %q, want %q", pe.Description, "A test extension")
				}
				if pe.Segment != "git" {
					t.Errorf("Segment = %q, want %q", pe.Segment, "git")
				}
				if len(pe.Provides) != 1 {
					t.Errorf("Provides length = %d, want 1", len(pe.Provides))
				}
				if len(pe.Modules) != 1 {
					t.Errorf("Modules length = %d, want 1", len(pe.Modules))
				}
			},
		},
		{
			name: "enabled defaults to true",
			yaml: &PromptExtensionYAML{
				Metadata: PromptExtensionMetadata{
					Name: "enabled-default",
				},
				Spec: PromptExtensionSpec{},
			},
			checks: func(t *testing.T, pe *PromptExtension) {
				if !pe.Enabled {
					t.Errorf("Enabled = false, want true")
				}
			},
		},
		{
			name: "enabled explicitly false",
			yaml: &PromptExtensionYAML{
				Metadata: PromptExtensionMetadata{
					Name: "disabled",
				},
				Spec: PromptExtensionSpec{
					Enabled: boolPtr(false),
				},
			},
			checks: func(t *testing.T, pe *PromptExtension) {
				if pe.Enabled {
					t.Errorf("Enabled = true, want false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pe := tt.yaml.ToPromptExtension()
			tt.checks(t, pe)
		})
	}
}

func TestPromptExtensionYAML_ResourceInterface(t *testing.T) {
	pey := NewPromptExtensionYAML("test-extension")

	if pey.GetKind() != KindPromptExtension {
		t.Errorf("GetKind() = %q, want %q", pey.GetKind(), KindPromptExtension)
	}

	if pey.GetName() != "test-extension" {
		t.Errorf("GetName() = %q, want %q", pey.GetName(), "test-extension")
	}

	if pey.GetAPIVersion() != "devopsmaestro.io/v1" {
		t.Errorf("GetAPIVersion() = %q, want %q", pey.GetAPIVersion(), "devopsmaestro.io/v1")
	}
}

// Helper function for bool pointers
func boolPtr(b bool) *bool {
	return &b
}
