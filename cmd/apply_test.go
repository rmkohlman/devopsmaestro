package cmd

import (
	"testing"
)

// Test isURL detection
func TestIsURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// URLs that should be detected
		{"http://example.com/file.yaml", true},
		{"https://example.com/file.yaml", true},
		{"https://raw.githubusercontent.com/user/repo/main/plugin.yaml", true},
		{"github:user/repo/path/file.yaml", true},
		{"github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml", true},

		// Local files that should NOT be detected as URLs
		{"./local/file.yaml", false},
		{"/absolute/path/file.yaml", false},
		{"file.yaml", false},
		{"relative/path/file.yaml", false},
		{"-", false}, // stdin
		{"", false},

		// Edge cases - strings that start with similar prefixes but aren't URLs
		{"httpnotaurl", false},
		{"httpsnotaurl", false},
		{"githubnotaurl", false},
		{"http", false},
		{"https", false},
		{"github", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isURL(tt.input)
			if got != tt.want {
				t.Errorf("isURL(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
