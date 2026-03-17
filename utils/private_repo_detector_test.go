package utils

import (
	"os"
	"path/filepath"
	"sort"
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
		wantSystemDeps        []string
	}{
		{
			name: "HTTPS-only with tokens",
			requirementsContent: "flask==2.3.0\n" +
				"git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/Org/repo.git@v1.0\n",
			wantNeedsSSH:          false,
			wantGitURLType:        "https",
			wantRequiredBuildArgs: []string{"GITHUB_USERNAME", "GITHUB_PAT"},
			wantNeedsGit:          true,
			wantSystemDeps:        nil,
		},
		{
			name:                  "SSH-only",
			requirementsContent:   "mylib @ git+ssh://git@github.com/Org/repo.git@v1.0\n",
			wantNeedsSSH:          true,
			wantGitURLType:        "ssh",
			wantRequiredBuildArgs: []string{},
			wantNeedsGit:          true,
			wantSystemDeps:        nil,
		},
		{
			name: "Mixed HTTPS token and SSH",
			requirementsContent: "git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/Org/private-lib.git@v1.0\n" +
				"mylib @ git+ssh://git@github.com/Org/repo.git@v2.0\n",
			wantNeedsSSH:          true,
			wantGitURLType:        "mixed",
			wantRequiredBuildArgs: []string{"GITHUB_USERNAME", "GITHUB_PAT"},
			wantNeedsGit:          true,
			wantSystemDeps:        nil,
		},
		{
			name:                  "No private repos — plain packages",
			requirementsContent:   "flask==2.3.0\nrequests>=2.28\nnumpy==1.25.0\n",
			wantNeedsSSH:          false,
			wantGitURLType:        "",
			wantRequiredBuildArgs: []string{},
			wantNeedsGit:          false,
			wantSystemDeps:        nil,
		},
		{
			name:                  "No requirements.txt file",
			requirementsContent:   "", // signals: do not create the file
			wantNeedsSSH:          false,
			wantGitURLType:        "",
			wantRequiredBuildArgs: []string{},
			wantNeedsGit:          false,
			wantSystemDeps:        nil,
		},
		{
			name:                  "HTTPS without variables",
			requirementsContent:   "git+https://github.com/Org/repo.git\n",
			wantNeedsSSH:          false,
			wantGitURLType:        "https",
			wantRequiredBuildArgs: []string{},
			wantNeedsGit:          true,
			wantSystemDeps:        nil,
		},
		{
			name: "Multiple token variables — deduplicated",
			requirementsContent: "git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/Org/repo.git@v1.0\n" +
				"git+https://${GITHUB_USERNAME}:${ARTIFACTORY_TOKEN}@artifactory.example.com/Org/lib.git@v2.0\n",
			wantNeedsSSH:          false,
			wantGitURLType:        "https",
			wantRequiredBuildArgs: []string{"GITHUB_USERNAME", "GITHUB_PAT", "ARTIFACTORY_TOKEN"},
			wantNeedsGit:          true,
			wantSystemDeps:        nil,
		},
		{
			name:                  "git@ SSH syntax",
			requirementsContent:   "mylib @ git+https://git@github.com/Org/repo.git\n",
			wantNeedsSSH:          true,
			wantGitURLType:        "mixed", // has git+https:// AND git@ pattern
			wantRequiredBuildArgs: []string{},
			wantNeedsGit:          true,
			wantSystemDeps:        nil,
		},
		{
			name:                  "Private repo with system dep packages",
			requirementsContent:   "git+https://${GITHUB_PAT}@github.com/Org/repo.git@v1.0\npsycopg2==2.9.9\n",
			wantNeedsSSH:          false,
			wantGitURLType:        "https",
			wantRequiredBuildArgs: []string{"GITHUB_PAT"},
			wantNeedsGit:          true,
			wantSystemDeps:        []string{"libpq-dev"},
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

			// Check SystemDeps — order-insensitive set comparison
			if tt.wantSystemDeps != nil {
				if len(got.SystemDeps) != len(tt.wantSystemDeps) {
					t.Errorf("SystemDeps = %v (len %d), want %v (len %d)",
						got.SystemDeps, len(got.SystemDeps),
						tt.wantSystemDeps, len(tt.wantSystemDeps))
				} else {
					wantSet := make(map[string]bool, len(tt.wantSystemDeps))
					for _, d := range tt.wantSystemDeps {
						wantSet[d] = true
					}
					for _, d := range got.SystemDeps {
						if !wantSet[d] {
							t.Errorf("SystemDeps contains unexpected entry %q; got %v, want %v",
								d, got.SystemDeps, tt.wantSystemDeps)
						}
					}
				}
			}
		})
	}
}

func TestNormalizePythonPkgName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{input: "psycopg2", want: "psycopg2"},
		{input: "Pillow", want: "pillow"},
		{input: "PyYAML", want: "pyyaml"},
		{input: "python-ldap", want: "python-ldap"},
		{input: "python_ldap", want: "python-ldap"},
		{input: "Some.Package", want: "some-package"},
		{input: "a..b__c--d", want: "a-b-c-d"},
		{input: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizePythonPkgName(tt.input)
			if got != tt.want {
				t.Errorf("normalizePythonPkgName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractPkgName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "version pin", input: "psycopg2==2.9.9", want: "psycopg2"},
		{name: "version gte", input: "Pillow>=10.0", want: "Pillow"},
		{name: "extras", input: "cryptography[ssh]", want: "cryptography"},
		{name: "version ne", input: "lxml!=4.9.0", want: "lxml"},
		{name: "bare package", input: "flask", want: "flask"},
		{name: "comment line", input: "# comment line", want: ""},
		{name: "requirements include", input: "-r other-requirements.txt", want: ""},
		{name: "pip option", input: "--index-url https://pypi.org/simple/", want: ""},
		{name: "empty line", input: "", want: ""},
		{name: "whitespace around", input: "  psycopg2 == 2.9.9  ", want: "psycopg2"},
		{name: "at syntax url", input: "mylib @ git+https://github.com/org/repo.git", want: "mylib"},
		{name: "multiple version constraints", input: "cffi>=1.0,<2.0", want: "cffi"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractPkgName(tt.input)
			if got != tt.want {
				t.Errorf("extractPkgName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestDetectPythonSystemDeps(t *testing.T) {
	tests := []struct {
		name             string
		requirements     string // empty string means no requirements.txt
		wantSystemDeps   []string
		wantSourceKeys   []string          // expected keys in SystemDepSources
		wantSourceValues map[string]string // expected key->value in SystemDepSources
	}{
		{
			name:           "psycopg2 needs libpq-dev",
			requirements:   "psycopg2==2.9.9\nflask==2.3.0\n",
			wantSystemDeps: []string{"libpq-dev"},
			wantSourceKeys: []string{"libpq-dev"},
			wantSourceValues: map[string]string{
				"libpq-dev": "psycopg2",
			},
		},
		{
			name:           "Pillow needs multiple deps",
			requirements:   "Pillow>=10.0\n",
			wantSystemDeps: []string{"libjpeg-dev", "zlib1g-dev", "libfreetype6-dev"},
			wantSourceKeys: []string{"libjpeg-dev", "zlib1g-dev", "libfreetype6-dev"},
			wantSourceValues: map[string]string{
				"libjpeg-dev":      "pillow",
				"zlib1g-dev":       "pillow",
				"libfreetype6-dev": "pillow",
			},
		},
		{
			name:           "multiple packages with shared deps",
			requirements:   "cryptography>=41.0\ncffi>=1.0\n",
			wantSystemDeps: []string{"libffi-dev", "libssl-dev"},
			wantSourceKeys: []string{"libffi-dev", "libssl-dev"},
			wantSourceValues: map[string]string{
				"libffi-dev": "cryptography",
				"libssl-dev": "cryptography",
			},
		},
		{
			name:             "psycopg2-binary needs nothing",
			requirements:     "psycopg2-binary==2.9.9\n",
			wantSystemDeps:   nil,
			wantSourceKeys:   nil,
			wantSourceValues: nil,
		},
		{
			name:             "no matching packages",
			requirements:     "flask==2.3.0\nrequests>=2.28\n",
			wantSystemDeps:   nil,
			wantSourceKeys:   nil,
			wantSourceValues: nil,
		},
		{
			name:             "no requirements.txt",
			requirements:     "", // do not create file
			wantSystemDeps:   nil,
			wantSourceKeys:   nil,
			wantSourceValues: nil,
		},
		{
			name:           "mixed known and unknown",
			requirements:   "psycopg2==2.9.9\nflask==2.3.0\nlxml>=4.0\n",
			wantSystemDeps: []string{"libpq-dev", "libxml2-dev", "libxslt1-dev"},
			wantSourceKeys: []string{"libpq-dev", "libxml2-dev", "libxslt1-dev"},
			wantSourceValues: map[string]string{
				"libpq-dev":    "psycopg2",
				"libxml2-dev":  "lxml",
				"libxslt1-dev": "lxml",
			},
		},
		{
			name:           "comments and pip options ignored",
			requirements:   "# psycopg2\n-r other.txt\n--index-url https://pypi.org\nPillow>=10.0\n",
			wantSystemDeps: []string{"libjpeg-dev", "zlib1g-dev", "libfreetype6-dev"},
			wantSourceKeys: []string{"libjpeg-dev", "zlib1g-dev", "libfreetype6-dev"},
			wantSourceValues: map[string]string{
				"libjpeg-dev":      "pillow",
				"zlib1g-dev":       "pillow",
				"libfreetype6-dev": "pillow",
			},
		},
		{
			name:           "case insensitive matching",
			requirements:   "PSYCOPG2==2.9.9\n",
			wantSystemDeps: []string{"libpq-dev"},
			wantSourceKeys: []string{"libpq-dev"},
			wantSourceValues: map[string]string{
				"libpq-dev": "psycopg2",
			},
		},
		{
			name:           "gevent needs libev",
			requirements:   "gevent>=23.0\n",
			wantSystemDeps: []string{"libev-dev", "libevent-dev"},
			wantSourceKeys: []string{"libev-dev", "libevent-dev"},
			wantSourceValues: map[string]string{
				"libev-dev":    "gevent",
				"libevent-dev": "gevent",
			},
		},
		{
			name:           "pyyaml needs libyaml",
			requirements:   "PyYAML>=6.0\n",
			wantSystemDeps: []string{"libyaml-dev"},
			wantSourceKeys: []string{"libyaml-dev"},
			wantSourceValues: map[string]string{
				"libyaml-dev": "pyyaml",
			},
		},
		{
			name:           "h5py needs libhdf5",
			requirements:   "h5py>=3.0\n",
			wantSystemDeps: []string{"libhdf5-dev"},
			wantSourceKeys: []string{"libhdf5-dev"},
			wantSourceValues: map[string]string{
				"libhdf5-dev": "h5py",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.requirements != "" {
				reqPath := filepath.Join(tmpDir, "requirements.txt")
				if err := os.WriteFile(reqPath, []byte(tt.requirements), 0644); err != nil {
					t.Fatalf("failed to write requirements.txt: %v", err)
				}
			}

			info := &PrivateRepoInfo{
				RequiredBuildArgs: []string{},
				DetectedInFiles:   []string{},
			}
			detectPythonSystemDeps(tmpDir, info)

			// Check SystemDeps — order-insensitive set comparison
			if tt.wantSystemDeps == nil {
				if len(info.SystemDeps) != 0 {
					t.Errorf("SystemDeps = %v, want empty/nil", info.SystemDeps)
				}
			} else {
				if len(info.SystemDeps) != len(tt.wantSystemDeps) {
					t.Errorf("SystemDeps = %v (len %d), want %v (len %d)",
						info.SystemDeps, len(info.SystemDeps),
						tt.wantSystemDeps, len(tt.wantSystemDeps))
				} else {
					wantSet := make(map[string]bool, len(tt.wantSystemDeps))
					for _, d := range tt.wantSystemDeps {
						wantSet[d] = true
					}
					gotSorted := make([]string, len(info.SystemDeps))
					copy(gotSorted, info.SystemDeps)
					sort.Strings(gotSorted)
					for _, d := range gotSorted {
						if !wantSet[d] {
							t.Errorf("SystemDeps contains unexpected entry %q; got %v, want %v",
								d, info.SystemDeps, tt.wantSystemDeps)
						}
					}
				}
			}

			// Check SystemDepSources keys and values
			if tt.wantSourceKeys == nil {
				if len(info.SystemDepSources) != 0 {
					t.Errorf("SystemDepSources = %v, want empty/nil", info.SystemDepSources)
				}
			} else {
				for _, key := range tt.wantSourceKeys {
					if _, ok := info.SystemDepSources[key]; !ok {
						t.Errorf("SystemDepSources missing key %q; got %v", key, info.SystemDepSources)
					}
				}
				for key, wantVal := range tt.wantSourceValues {
					if gotVal, ok := info.SystemDepSources[key]; !ok {
						t.Errorf("SystemDepSources missing key %q", key)
					} else if gotVal != wantVal {
						t.Errorf("SystemDepSources[%q] = %q, want %q", key, gotVal, wantVal)
					}
				}
			}
		})
	}
}
