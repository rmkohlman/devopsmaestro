package shellgen

import (
	"strings"
	"testing"
)

func TestNewShellGenerator(t *testing.T) {
	tests := []struct {
		shell    string
		wantErr  bool
		wantName string
	}{
		{"zsh", false, "zsh"},
		{"bash", false, "bash"},
		{"fish", false, "fish"},
		{"csh", true, ""},
		{"", true, ""},
		{"ZSH", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			gen, err := NewShellGenerator(tt.shell)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gen.ShellName() != tt.wantName {
				t.Errorf("ShellName() = %q, want %q", gen.ShellName(), tt.wantName)
			}
		})
	}
}

func TestAllGenerators_SameConfig(t *testing.T) {
	// Verify all three generators produce valid output from the same config
	config := ShellConfig{
		WorkspaceName: "test-workspace",
		EnvVars: map[string]string{
			"PROJECT":  "/workspace",
			"NODE_ENV": "development",
		},
		PromptConfig: "starship-init-placeholder",
	}

	shells := []string{"zsh", "bash", "fish"}
	for _, shell := range shells {
		t.Run(shell, func(t *testing.T) {
			gen, err := NewShellGenerator(shell)
			if err != nil {
				t.Fatalf("failed to create %s generator: %v", shell, err)
			}

			output, err := gen.Generate(config)
			if err != nil {
				t.Fatalf("%s generator failed: %v", shell, err)
			}

			// All generators should include workspace name
			if !strings.Contains(output, "test-workspace") {
				t.Errorf("%s output missing workspace name", shell)
			}

			// All generators should include prompt config
			if !strings.Contains(output, "starship-init-placeholder") {
				t.Errorf("%s output missing prompt config", shell)
			}

			// All generators should include env vars
			if !strings.Contains(output, "PROJECT") {
				t.Errorf("%s output missing env var PROJECT", shell)
			}
			if !strings.Contains(output, "NODE_ENV") {
				t.Errorf("%s output missing env var NODE_ENV", shell)
			}
		})
	}
}

func TestZshAndBashUseExport_FishUsesSet(t *testing.T) {
	config := ShellConfig{
		EnvVars: map[string]string{"FOO": "bar"},
	}

	// Zsh
	zsh := NewZshGenerator()
	zshOut, _ := zsh.Generate(config)
	if !strings.Contains(zshOut, `export FOO="bar"`) {
		t.Error("Zsh should use 'export' for env vars")
	}

	// Bash
	bash := NewBashGenerator()
	bashOut, _ := bash.Generate(config)
	if !strings.Contains(bashOut, `export FOO="bar"`) {
		t.Error("Bash should use 'export' for env vars")
	}

	// Fish
	fish := NewFishGenerator()
	fishOut, _ := fish.Generate(config)
	if !strings.Contains(fishOut, `set -gx FOO "bar"`) {
		t.Error("Fish should use 'set -gx' for env vars")
	}
	if strings.Contains(fishOut, "export") {
		t.Error("Fish should NOT use 'export'")
	}
}
