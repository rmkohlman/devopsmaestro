package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectPythonPrivateRepos(t *testing.T) {
	tests := []struct {
		name                  string
		requirementsContent   string // empty string means no requirements.txt
		wantNeedsSSH          bool
		wantGitURLType        string
		wantRequiredBuildArgs []string
		wantNeedsGit          bool
	}{
		{
			name: "HTTPS-only with tokens",
			requirementsContent: "flask==2.3.0\n" +
				"git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/Org/repo.git@v1.0\n",
			wantNeedsSSH:          false,
			wantGitURLType:        "https",
			wantRequiredBuildArgs: []string{"GITHUB_USERNAME", "GITHUB_PAT"},
			wantNeedsGit:          true,
		},
		{
			name:                  "SSH-only",
			requirementsContent:   "mylib @ git+ssh://git@github.com/Org/repo.git@v1.0\n",
			wantNeedsSSH:          true,
			wantGitURLType:        "ssh",
			wantRequiredBuildArgs: []string{},
			wantNeedsGit:          true,
		},
		{
			name: "Mixed HTTPS token and SSH",
			requirementsContent: "git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/Org/private-lib.git@v1.0\n" +
				"mylib @ git+ssh://git@github.com/Org/repo.git@v2.0\n",
			wantNeedsSSH:          true,
			wantGitURLType:        "mixed",
			wantRequiredBuildArgs: []string{"GITHUB_USERNAME", "GITHUB_PAT"},
			wantNeedsGit:          true,
		},
		{
			name:                  "No private repos — plain packages",
			requirementsContent:   "flask==2.3.0\nrequests>=2.28\nnumpy==1.25.0\n",
			wantNeedsSSH:          false,
			wantGitURLType:        "",
			wantRequiredBuildArgs: []string{},
			wantNeedsGit:          false,
		},
		{
			name:                  "No requirements.txt file",
			requirementsContent:   "", // signals: do not create the file
			wantNeedsSSH:          false,
			wantGitURLType:        "",
			wantRequiredBuildArgs: []string{},
			wantNeedsGit:          false,
		},
		{
			name:                  "HTTPS without variables",
			requirementsContent:   "git+https://github.com/Org/repo.git\n",
			wantNeedsSSH:          false,
			wantGitURLType:        "https",
			wantRequiredBuildArgs: []string{},
			wantNeedsGit:          true,
		},
		{
			name: "Multiple token variables — deduplicated",
			requirementsContent: "git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/Org/repo.git@v1.0\n" +
				"git+https://${GITHUB_USERNAME}:${ARTIFACTORY_TOKEN}@artifactory.example.com/Org/lib.git@v2.0\n",
			wantNeedsSSH:          false,
			wantGitURLType:        "https",
			wantRequiredBuildArgs: []string{"GITHUB_USERNAME", "GITHUB_PAT", "ARTIFACTORY_TOKEN"},
			wantNeedsGit:          true,
		},
		{
			name:                  "git@ SSH syntax",
			requirementsContent:   "mylib @ git+https://git@github.com/Org/repo.git\n",
			wantNeedsSSH:          true,
			wantGitURLType:        "mixed", // has git+https:// AND git@ pattern
			wantRequiredBuildArgs: []string{},
			wantNeedsGit:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create requirements.txt only when content is non-empty
			if tt.requirementsContent != "" {
				reqPath := filepath.Join(tmpDir, "requirements.txt")
				if err := os.WriteFile(reqPath, []byte(tt.requirementsContent), 0644); err != nil {
					t.Fatalf("failed to write requirements.txt: %v", err)
				}
			}

			got := DetectPrivateRepos(tmpDir, "python")

			if got.NeedsSSH != tt.wantNeedsSSH {
				t.Errorf("NeedsSSH = %v, want %v", got.NeedsSSH, tt.wantNeedsSSH)
			}

			if got.GitURLType != tt.wantGitURLType {
				t.Errorf("GitURLType = %q, want %q", got.GitURLType, tt.wantGitURLType)
			}

			if got.NeedsGit != tt.wantNeedsGit {
				t.Errorf("NeedsGit = %v, want %v", got.NeedsGit, tt.wantNeedsGit)
			}

			// Check RequiredBuildArgs — order-insensitive set comparison
			if len(got.RequiredBuildArgs) != len(tt.wantRequiredBuildArgs) {
				t.Errorf("RequiredBuildArgs = %v (len %d), want %v (len %d)",
					got.RequiredBuildArgs, len(got.RequiredBuildArgs),
					tt.wantRequiredBuildArgs, len(tt.wantRequiredBuildArgs))
			} else {
				wantSet := make(map[string]bool, len(tt.wantRequiredBuildArgs))
				for _, a := range tt.wantRequiredBuildArgs {
					wantSet[a] = true
				}
				for _, a := range got.RequiredBuildArgs {
					if !wantSet[a] {
						t.Errorf("RequiredBuildArgs contains unexpected entry %q; got %v, want %v",
							a, got.RequiredBuildArgs, tt.wantRequiredBuildArgs)
					}
				}
			}
		})
	}
}
