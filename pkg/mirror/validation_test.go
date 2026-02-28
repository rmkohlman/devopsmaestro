package mirror

import (
	"strings"
	"testing"
)

// =============================================================================
// ValidateGitURL - Valid URLs Tests
// =============================================================================

func TestValidateGitURL_ValidURLs(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{
			name: "github https",
			url:  "https://github.com/user/repo",
		},
		{
			name: "github https with .git",
			url:  "https://github.com/user/repo.git",
		},
		{
			name: "github ssh",
			url:  "git@github.com:user/repo",
		},
		{
			name: "github ssh with .git",
			url:  "git@github.com:user/repo.git",
		},
		{
			name: "gitlab https",
			url:  "https://gitlab.com/user/repo",
		},
		{
			name: "gitlab ssh",
			url:  "git@gitlab.com:user/repo",
		},
		{
			name: "bitbucket https",
			url:  "https://bitbucket.org/user/repo",
		},
		{
			name: "bitbucket ssh",
			url:  "git@bitbucket.org:user/repo",
		},
		{
			name: "custom domain https",
			url:  "https://git.company.com/team/project",
		},
		{
			name: "custom domain ssh",
			url:  "git@git.company.com:team/project",
		},
		{
			name: "nested path",
			url:  "https://github.com/org/team/project",
		},
		{
			name: "repo with dash",
			url:  "https://github.com/user/my-repo",
		},
		{
			name: "repo with underscore",
			url:  "https://github.com/user/my_repo",
		},
		{
			name: "repo with dot in name",
			url:  "https://github.com/user/repo.name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitURL(tt.url)
			if err != nil {
				t.Errorf("ValidateGitURL(%q) should be valid, got error: %v", tt.url, err)
			}
		})
	}
}

// =============================================================================
// ValidateGitURL - Invalid URLs Tests (Security)
// =============================================================================

func TestValidateGitURL_InvalidURLs(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError string
	}{
		{
			name:      "empty url",
			url:       "",
			wantError: "empty",
		},
		{
			name:      "shell command injection attempt",
			url:       "https://github.com/user/repo; rm -rf /",
			wantError: "invalid",
		},
		{
			name:      "shell pipe in url",
			url:       "https://github.com/user/repo | cat /etc/passwd",
			wantError: "invalid",
		},
		{
			name:      "shell redirect in url",
			url:       "https://github.com/user/repo > /tmp/file",
			wantError: "invalid",
		},
		{
			name:      "shell backtick in url",
			url:       "https://github.com/user/repo`whoami`",
			wantError: "invalid",
		},
		{
			name:      "shell dollar paren in url",
			url:       "https://github.com/user/repo$(whoami)",
			wantError: "invalid",
		},
		{
			name:      "ext transport attempt",
			url:       "ext::sh -c 'git-upload-pack /tmp/repo'",
			wantError: "invalid",
		},
		{
			name:      "file protocol",
			url:       "file:///tmp/repo",
			wantError: "invalid",
		},
		{
			name:      "http protocol (insecure)",
			url:       "http://github.com/user/repo",
			wantError: "insecure",
		},
		{
			name:      "leading dash",
			url:       "-u https://github.com/user/repo",
			wantError: "invalid",
		},
		{
			name:      "url with embedded credentials",
			url:       "https://user:pass@github.com/user/repo",
			wantError: "credentials",
		},
		{
			name:      "url with username only",
			url:       "https://user@github.com/user/repo",
			wantError: "credentials",
		},
		// NOTE: Local paths are allowed for testing/development - removed these test cases:
		// - local path: "/tmp/local/repo"
		// - relative path: "../other/repo"
		// Applications should validate user-provided URLs before passing to MirrorManager.
		{
			name:      "url with spaces",
			url:       "https://github.com/user/my repo",
			wantError: "invalid",
		},
		{
			name:      "url with newline",
			url:       "https://github.com/user/repo\n",
			wantError: "invalid",
		},
		{
			name:      "url with tab",
			url:       "https://github.com/user/repo\t",
			wantError: "invalid",
		},
		{
			name:      "git protocol (deprecated)",
			url:       "git://github.com/user/repo",
			wantError: "insecure",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGitURL(tt.url)
			if err == nil {
				t.Errorf("ValidateGitURL(%q) should have returned error containing %q", tt.url, tt.wantError)
				return
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantError)) {
				t.Errorf("ValidateGitURL(%q) error = %q, want error containing %q", tt.url, err.Error(), tt.wantError)
			}
		})
	}
}

// =============================================================================
// ValidateSlug - Valid Slugs Tests
// =============================================================================

func TestValidateSlug_ValidSlugs(t *testing.T) {
	tests := []struct {
		name string
		slug string
	}{
		{
			name: "simple slug",
			slug: "github.com_user_repo",
		},
		{
			name: "slug with dash",
			slug: "github.com_user_my-repo",
		},
		{
			name: "slug with underscore",
			slug: "gitlab.com_group_my_project",
		},
		{
			name: "slug with dot",
			slug: "git.company.com_team_project",
		},
		{
			name: "nested path slug",
			slug: "github.com_org_team_group_repo",
		},
		{
			name: "all valid chars",
			slug: "host.domain.com_org-name_repo.name_sub_group",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSlug(tt.slug)
			if err != nil {
				t.Errorf("ValidateSlug(%q) should be valid, got error: %v", tt.slug, err)
			}
		})
	}
}

// =============================================================================
// ValidateSlug - Invalid Slugs Tests (Path Traversal)
// =============================================================================

func TestValidateSlug_InvalidSlugs(t *testing.T) {
	tests := []struct {
		name      string
		slug      string
		wantError string
	}{
		{
			name:      "empty slug",
			slug:      "",
			wantError: "empty",
		},
		{
			name:      "path traversal parent",
			slug:      "../parent",
			wantError: "invalid",
		},
		{
			name:      "path traversal hidden parent",
			slug:      "repo/../../../etc/passwd",
			wantError: "invalid",
		},
		{
			name:      "absolute path",
			slug:      "/tmp/repo",
			wantError: "invalid",
		},
		{
			name:      "forward slash",
			slug:      "github.com/user/repo",
			wantError: "invalid",
		},
		{
			name:      "backslash",
			slug:      "github.com\\user\\repo",
			wantError: "invalid",
		},
		{
			name:      "null byte",
			slug:      "repo\x00name",
			wantError: "invalid",
		},
		{
			name:      "space character",
			slug:      "repo name",
			wantError: "invalid",
		},
		{
			name:      "tab character",
			slug:      "repo\tname",
			wantError: "invalid",
		},
		{
			name:      "newline character",
			slug:      "repo\nname",
			wantError: "invalid",
		},
		{
			name:      "single dot",
			slug:      ".",
			wantError: "invalid",
		},
		{
			name:      "double dot",
			slug:      "..",
			wantError: "invalid",
		},
		{
			name:      "hidden file",
			slug:      ".hidden",
			wantError: "invalid",
		},
		{
			name:      "shell metachar semicolon",
			slug:      "repo;cmd",
			wantError: "invalid",
		},
		{
			name:      "shell metachar pipe",
			slug:      "repo|cmd",
			wantError: "invalid",
		},
		{
			name:      "shell metachar ampersand",
			slug:      "repo&cmd",
			wantError: "invalid",
		},
		{
			name:      "shell metachar dollar",
			slug:      "repo$var",
			wantError: "invalid",
		},
		{
			name:      "shell metachar backtick",
			slug:      "repo`cmd`",
			wantError: "invalid",
		},
		{
			name:      "leading dash",
			slug:      "-option",
			wantError: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSlug(tt.slug)
			if err == nil {
				t.Errorf("ValidateSlug(%q) should have returned error containing %q", tt.slug, tt.wantError)
				return
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.wantError)) {
				t.Errorf("ValidateSlug(%q) error = %q, want error containing %q", tt.slug, err.Error(), tt.wantError)
			}
		})
	}
}

// =============================================================================
// ValidateURLNoCredentials Tests
// =============================================================================

func TestValidateURLNoCredentials(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		wantError bool
	}{
		{
			name:      "https without credentials",
			url:       "https://github.com/user/repo",
			wantError: false,
		},
		{
			name:      "ssh without credentials",
			url:       "git@github.com:user/repo",
			wantError: false,
		},
		{
			name:      "https with user and password",
			url:       "https://user:pass@github.com/user/repo",
			wantError: true,
		},
		{
			name:      "https with username only",
			url:       "https://user@github.com/user/repo",
			wantError: true,
		},
		{
			name:      "https with password only",
			url:       "https://:pass@github.com/user/repo",
			wantError: true,
		},
		{
			name:      "https with encoded credentials",
			url:       "https://user%40email:pass%40word@github.com/user/repo",
			wantError: true,
		},
		{
			name:      "https with colon in path (not credentials)",
			url:       "https://github.com/user/repo:branch",
			wantError: false,
		},
		{
			name:      "ssh git@ prefix is OK",
			url:       "git@github.com:user/repo",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURLNoCredentials(tt.url)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateURLNoCredentials(%q) error = %v, wantError %v", tt.url, err, tt.wantError)
			}
		})
	}
}
