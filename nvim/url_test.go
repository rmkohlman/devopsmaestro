package nvim

import (
	"testing"
)

func TestIsGitURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		// Valid Git URLs
		{"HTTPS URL", "https://github.com/user/repo.git", true},
		{"HTTP URL", "http://github.com/user/repo.git", true},
		{"Git protocol", "git://github.com/user/repo.git", true},
		{"SSH format", "git@github.com:user/repo.git", true},
		{"GitHub shorthand", "github:user/repo", true},
		{"GitLab shorthand", "gitlab:user/repo", true},
		{"Bitbucket shorthand", "bitbucket:user/repo", true},
		{"Ends with .git", "custom.server/path/repo.git", true},

		// Invalid/non-Git strings
		{"Plain template name", "kickstart", false},
		{"Plain word", "minimal", false},
		{"Local path", "/home/user/config", false},
		{"Relative path", "../config", false},
		{"Empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGitURL(tt.input)
			if result != tt.expected {
				t.Errorf("IsGitURL(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "GitHub shorthand",
			input:    "github:nvim-lua/kickstart.nvim",
			expected: "https://github.com/nvim-lua/kickstart.nvim.git",
		},
		{
			name:     "GitLab shorthand",
			input:    "gitlab:user/project",
			expected: "https://gitlab.com/user/project.git",
		},
		{
			name:     "Bitbucket shorthand",
			input:    "bitbucket:team/repo",
			expected: "https://bitbucket.org/team/repo.git",
		},
		{
			name:     "Full HTTPS URL unchanged",
			input:    "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "SSH URL unchanged",
			input:    "git@github.com:user/repo.git",
			expected: "git@github.com:user/repo.git",
		},
		{
			name:     "Git protocol unchanged",
			input:    "git://server.com/repo.git",
			expected: "git://server.com/repo.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeGitURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeGitURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParseGitURL(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedURL      string
		expectedPlatform string
		expectedRepo     string
	}{
		{
			name:             "GitHub HTTPS",
			input:            "https://github.com/user/repo.git",
			expectedURL:      "https://github.com/user/repo.git",
			expectedPlatform: "github",
			expectedRepo:     "repo",
		},
		{
			name:             "GitHub shorthand",
			input:            "github:nvim-lua/kickstart.nvim",
			expectedURL:      "https://github.com/nvim-lua/kickstart.nvim.git",
			expectedPlatform: "github",
			expectedRepo:     "kickstart.nvim",
		},
		{
			name:             "GitLab",
			input:            "https://gitlab.com/user/project.git",
			expectedURL:      "https://gitlab.com/user/project.git",
			expectedPlatform: "gitlab",
			expectedRepo:     "project",
		},
		{
			name:             "Bitbucket",
			input:            "https://bitbucket.org/team/repo.git",
			expectedURL:      "https://bitbucket.org/team/repo.git",
			expectedPlatform: "bitbucket",
			expectedRepo:     "repo",
		},
		{
			name:             "Custom Git server",
			input:            "https://git.company.com/team/project.git",
			expectedURL:      "https://git.company.com/team/project.git",
			expectedPlatform: "git",
			expectedRepo:     "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseGitURL(tt.input)

			if result.FullURL != tt.expectedURL {
				t.Errorf("ParseGitURL(%q).FullURL = %q, want %q", tt.input, result.FullURL, tt.expectedURL)
			}

			if result.Platform != tt.expectedPlatform {
				t.Errorf("ParseGitURL(%q).Platform = %q, want %q", tt.input, result.Platform, tt.expectedPlatform)
			}

			if result.RepoName != tt.expectedRepo {
				t.Errorf("ParseGitURL(%q).RepoName = %q, want %q", tt.input, result.RepoName, tt.expectedRepo)
			}
		})
	}
}
