package toolconfig

import (
	"sort"
	"testing"
)

func TestNewToolConfigGenerator(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		wantErr  bool
		wantTool string
	}{
		{name: "bat", toolName: "bat", wantTool: "bat"},
		{name: "delta", toolName: "delta", wantTool: "delta"},
		{name: "fzf", toolName: "fzf", wantTool: "fzf"},
		{name: "dircolors", toolName: "dircolors", wantTool: "dircolors"},
		{name: "unknown", toolName: "unknown-tool", wantErr: true},
		{name: "empty", toolName: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen, err := NewToolConfigGenerator(tt.toolName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewToolConfigGenerator(%q) error = %v, wantErr %v",
					tt.toolName, err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if gen.ToolName() != tt.wantTool {
				t.Errorf("ToolName() = %q, want %q", gen.ToolName(), tt.wantTool)
			}
		})
	}
}

func TestAllGenerators(t *testing.T) {
	gens := AllGenerators()
	if len(gens) != 4 {
		t.Fatalf("AllGenerators() returned %d generators, want 4", len(gens))
	}

	// Verify all generators are unique
	names := make(map[string]bool)
	for _, g := range gens {
		if names[g.ToolName()] {
			t.Errorf("duplicate generator for tool %q", g.ToolName())
		}
		names[g.ToolName()] = true
	}
}

func TestAvailableTools(t *testing.T) {
	tools := AvailableTools()
	want := []string{"bat", "delta", "dircolors", "fzf"}

	sort.Strings(tools)
	sort.Strings(want)

	if len(tools) != len(want) {
		t.Fatalf("AvailableTools() = %v, want %v", tools, want)
	}
	for i := range tools {
		if tools[i] != want[i] {
			t.Errorf("AvailableTools()[%d] = %q, want %q", i, tools[i], want[i])
		}
	}
}

func TestAllGenerators_ProduceOutput(t *testing.T) {
	pal := testPalette()
	for _, gen := range AllGenerators() {
		t.Run(gen.ToolName(), func(t *testing.T) {
			output, err := gen.Generate(pal)
			if err != nil {
				t.Fatalf("%s.Generate() error = %v", gen.ToolName(), err)
			}
			if output == "" {
				t.Errorf("%s.Generate() returned empty output", gen.ToolName())
			}
			if gen.Description() == "" {
				t.Errorf("%s.Description() returned empty", gen.ToolName())
			}
		})
	}
}
