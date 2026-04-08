package shellgen

import "testing"

func TestDetectShellFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		shellEnv string
		want     string
	}{
		{"empty defaults to zsh", "", "zsh"},
		{"zsh path", "/bin/zsh", "zsh"},
		{"bash path", "/bin/bash", "bash"},
		{"fish path", "/usr/bin/fish", "fish"},
		{"usr local zsh", "/usr/local/bin/zsh", "zsh"},
		{"usr local bash", "/usr/local/bin/bash", "bash"},
		{"homebrew fish", "/opt/homebrew/bin/fish", "fish"},
		{"nix bash", "/nix/store/abc123/bin/bash", "bash"},
		{"unknown shell defaults to zsh", "/bin/csh", "zsh"},
		{"sh defaults to zsh", "/bin/sh", "zsh"},
		{"tcsh defaults to zsh", "/bin/tcsh", "zsh"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectShellFromEnv(tt.shellEnv)
			if got != tt.want {
				t.Errorf("detectShellFromEnv(%q) = %q, want %q", tt.shellEnv, got, tt.want)
			}
		})
	}
}
