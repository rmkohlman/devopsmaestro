package mirror

import (
	"testing"
)

// =============================================================================
// NormalizeGitURL Tests
// =============================================================================

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
	}{
		{
			name:     "https url without .git suffix",
			url:      "https://github.com/user/repo",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "https url with .git suffix",
			url:      "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo.git",
		},
		{
			name:     "ssh url without .git suffix",
			url:      "git@github.com:user/repo",
			expected: "git@github.com:user/repo.git",
		},
		{
			name:     "ssh url with .git suffix",
			url:      "git@github.com:user/repo.git",
			expected: "git@github.com:user/repo.git",
		},
		{
			name:     "gitlab https url",
			url:      "https://gitlab.com/user/repo",
			expected: "https://gitlab.com/user/repo.git",
		},
		{
			name:     "gitlab ssh url",
			url:      "git@gitlab.com:user/repo",
			expected: "git@gitlab.com:user/repo.git",
		},
		{
			name:     "bitbucket https url",
			url:      "https://bitbucket.org/user/repo",
			expected: "https://bitbucket.org/user/repo.git",
		},
		{
			name:     "bitbucket ssh url",
			url:      "git@bitbucket.org:user/repo",
			expected: "git@bitbucket.org:user/repo.git",
		},
		{
			name:     "custom domain https",
			url:      "https://git.company.com/team/project",
			expected: "https://git.company.com/team/project.git",
		},
		{
			name:     "custom domain ssh",
			url:      "git@git.company.com:team/project",
			expected: "git@git.company.com:team/project.git",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeGitURL(tt.url)
			if result != tt.expected {
				t.Errorf("NormalizeGitURL(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// GenerateSlug Tests
// =============================================================================

func TestGenerateSlug(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected string
		wantErr  bool
	}{
		{
			name:     "github https url",
			url:      "https://github.com/user/repo",
			expected: "github.com_user_repo",
			wantErr:  false,
		},
		{
			name:     "github https url with .git",
			url:      "https://github.com/user/repo.git",
			expected: "github.com_user_repo",
			wantErr:  false,
		},
		{
			name:     "github ssh url",
			url:      "git@github.com:user/repo",
			expected: "github.com_user_repo",
			wantErr:  false,
		},
		{
			name:     "github ssh url with .git",
			url:      "git@github.com:user/repo.git",
			expected: "github.com_user_repo",
			wantErr:  false,
		},
		{
			name:     "gitlab url",
			url:      "https://gitlab.com/group/subgroup/project",
			expected: "gitlab.com_group_subgroup_project",
			wantErr:  false,
		},
		{
			name:     "bitbucket url",
			url:      "https://bitbucket.org/team/repo",
			expected: "bitbucket.org_team_repo",
			wantErr:  false,
		},
		{
			name:     "custom domain",
			url:      "https://git.company.com/team/project",
			expected: "git.company.com_team_project",
			wantErr:  false,
		},
		{
			name:     "deep nested path",
			url:      "https://github.com/org/team/group/repo",
			expected: "github.com_org_team_group_repo",
			wantErr:  false,
		},
		{
			name:     "repo with dash",
			url:      "https://github.com/user/my-repo",
			expected: "github.com_user_my-repo",
			wantErr:  false,
		},
		{
			name:     "repo with underscore",
			url:      "https://github.com/user/my_repo",
			expected: "github.com_user_my_repo",
			wantErr:  false,
		},
		{
			name:     "empty url",
			url:      "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid url",
			url:      "not-a-url",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateSlug(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateSlug(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
				return
			}
			if result != tt.expected {
				t.Errorf("GenerateSlug(%q) = %q, want %q", tt.url, result, tt.expected)
			}
		})
	}
}

// =============================================================================
// SSH and HTTPS Same Slug Test
// =============================================================================

func TestSSHAndHTTPSSameSlug(t *testing.T) {
	tests := []struct {
		name      string
		httpsURL  string
		sshURL    string
		wantEqual bool
	}{
		{
			name:      "github repo same slug",
			httpsURL:  "https://github.com/user/repo",
			sshURL:    "git@github.com:user/repo",
			wantEqual: true,
		},
		{
			name:      "github repo with .git same slug",
			httpsURL:  "https://github.com/user/repo.git",
			sshURL:    "git@github.com:user/repo.git",
			wantEqual: true,
		},
		{
			name:      "gitlab repo same slug",
			httpsURL:  "https://gitlab.com/group/project",
			sshURL:    "git@gitlab.com:group/project",
			wantEqual: true,
		},
		{
			name:      "bitbucket repo same slug",
			httpsURL:  "https://bitbucket.org/team/repo",
			sshURL:    "git@bitbucket.org:team/repo",
			wantEqual: true,
		},
		{
			name:      "custom domain same slug",
			httpsURL:  "https://git.company.com/team/project",
			sshURL:    "git@git.company.com:team/project",
			wantEqual: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			httpsSlug, err := GenerateSlug(tt.httpsURL)
			if err != nil {
				t.Fatalf("GenerateSlug(%q) error = %v", tt.httpsURL, err)
			}

			sshSlug, err := GenerateSlug(tt.sshURL)
			if err != nil {
				t.Fatalf("GenerateSlug(%q) error = %v", tt.sshURL, err)
			}

			equal := httpsSlug == sshSlug
			if equal != tt.wantEqual {
				t.Errorf("Slug equality = %v, want %v\nHTTPS slug: %q\nSSH slug:   %q",
					equal, tt.wantEqual, httpsSlug, sshSlug)
			}
		})
	}
}
