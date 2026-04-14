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

func TestDetectDotnetPrivateRepos(t *testing.T) {
	tests := []struct {
		name                  string
		nugetConfigContent    string // empty string means no NuGet.config
		wantRequiredBuildArgs []string
	}{
		{
			name: "Private feed detected",
			nugetConfigContent: `<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" />
    <add key="PrivateFeed" value="https://pkgs.dev.azure.com/myorg/_packaging/myfeed/nuget/v3/index.json" />
  </packageSources>
</configuration>`,
			wantRequiredBuildArgs: []string{"NUGET_API_KEY"},
		},
		{
			name: "Only public nuget.org — no private feeds",
			nugetConfigContent: `<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <add key="nuget.org" value="https://api.nuget.org/v3/index.json" />
  </packageSources>
</configuration>`,
			wantRequiredBuildArgs: []string{},
		},
		{
			name:                  "No NuGet.config file",
			nugetConfigContent:    "",
			wantRequiredBuildArgs: []string{},
		},
		{
			name: "GitHub Packages private feed",
			nugetConfigContent: `<?xml version="1.0" encoding="utf-8"?>
<configuration>
  <packageSources>
    <add key="github" value="https://nuget.pkg.github.com/myorg/index.json" />
  </packageSources>
</configuration>`,
			wantRequiredBuildArgs: []string{"NUGET_API_KEY"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.nugetConfigContent != "" {
				nugetPath := filepath.Join(tmpDir, "NuGet.config")
				if err := os.WriteFile(nugetPath, []byte(tt.nugetConfigContent), 0644); err != nil {
					t.Fatalf("failed to write NuGet.config: %v", err)
				}
			}

			got := DetectPrivateRepos(tmpDir, "dotnet")

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

func TestDetectPhpPrivateRepos(t *testing.T) {
	tests := []struct {
		name                  string
		composerContent       string // empty means no composer.json
		wantNeedsGit          bool
		wantNeedsSSH          bool
		wantGitURLType        string
		wantRequiredBuildArgs []string
	}{
		{
			name: "Private VCS repo via HTTPS",
			composerContent: `{
				"repositories": [
					{"type": "vcs", "url": "https://github.com/myorg/private-package"}
				],
				"require": {"myorg/private-package": "^1.0"}
			}`,
			wantNeedsGit:          true,
			wantNeedsSSH:          false,
			wantGitURLType:        "https",
			wantRequiredBuildArgs: []string{"COMPOSER_AUTH"},
		},
		{
			name: "Private VCS repo via SSH",
			composerContent: `{
				"repositories": [
					{"type": "vcs", "url": "git@github.com:myorg/private-package.git"}
				],
				"require": {"myorg/private-package": "^1.0"}
			}`,
			wantNeedsGit:          true,
			wantNeedsSSH:          true,
			wantGitURLType:        "ssh",
			wantRequiredBuildArgs: []string{"COMPOSER_AUTH"},
		},
		{
			name: "Mixed HTTPS and SSH repos",
			composerContent: `{
				"repositories": [
					{"type": "vcs", "url": "https://github.com/myorg/package-a"},
					{"type": "vcs", "url": "git@github.com:myorg/package-b.git"}
				],
				"require": {"myorg/package-a": "^1.0", "myorg/package-b": "^2.0"}
			}`,
			wantNeedsGit:          true,
			wantNeedsSSH:          true,
			wantGitURLType:        "mixed",
			wantRequiredBuildArgs: []string{"COMPOSER_AUTH"},
		},
		{
			name: "No repositories key — public packages only",
			composerContent: `{
				"require": {"laravel/framework": "^11.0"}
			}`,
			wantNeedsGit:          false,
			wantNeedsSSH:          false,
			wantGitURLType:        "",
			wantRequiredBuildArgs: []string{},
		},
		{
			name:                  "No composer.json file",
			composerContent:       "",
			wantNeedsGit:          false,
			wantNeedsSSH:          false,
			wantGitURLType:        "",
			wantRequiredBuildArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.composerContent != "" {
				composerPath := filepath.Join(tmpDir, "composer.json")
				if err := os.WriteFile(composerPath, []byte(tt.composerContent), 0644); err != nil {
					t.Fatalf("failed to write composer.json: %v", err)
				}
			}

			got := DetectPrivateRepos(tmpDir, "php")

			if got.NeedsGit != tt.wantNeedsGit {
				t.Errorf("NeedsGit = %v, want %v", got.NeedsGit, tt.wantNeedsGit)
			}
			if got.NeedsSSH != tt.wantNeedsSSH {
				t.Errorf("NeedsSSH = %v, want %v", got.NeedsSSH, tt.wantNeedsSSH)
			}
			if got.GitURLType != tt.wantGitURLType {
				t.Errorf("GitURLType = %q, want %q", got.GitURLType, tt.wantGitURLType)
			}

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

func TestDetectElixirPrivateRepos(t *testing.T) {
	tests := []struct {
		name                  string
		mixContent            string // empty means no mix.exs
		wantNeedsGit          bool
		wantNeedsSSH          bool
		wantGitURLType        string
		wantRequiredBuildArgs []string
	}{
		{
			name: "Private git dep via HTTPS",
			mixContent: `defmodule MyApp.MixProject do
  defp deps do
    [{:private_lib, git: "https://github.com/myorg/private_lib.git", tag: "v1.0"}]
  end
end`,
			wantNeedsGit:          true,
			wantNeedsSSH:          false,
			wantGitURLType:        "https",
			wantRequiredBuildArgs: []string{"GITHUB_TOKEN"},
		},
		{
			name: "Private git dep via SSH",
			mixContent: `defmodule MyApp.MixProject do
  defp deps do
    [{:private_lib, git: "git@github.com:myorg/private_lib.git"}]
  end
end`,
			wantNeedsGit:          true,
			wantNeedsSSH:          true,
			wantGitURLType:        "ssh",
			wantRequiredBuildArgs: []string{},
		},
		{
			name: "Mixed HTTPS and SSH git deps",
			mixContent: `defmodule MyApp.MixProject do
  defp deps do
    [
      {:lib_a, git: "https://github.com/myorg/lib_a.git"},
      {:lib_b, git: "git@github.com:myorg/lib_b.git"}
    ]
  end
end`,
			wantNeedsGit:          true,
			wantNeedsSSH:          true,
			wantGitURLType:        "mixed",
			wantRequiredBuildArgs: []string{"GITHUB_TOKEN"},
		},
		{
			name: "No private deps — public Hex packages only",
			mixContent: `defmodule MyApp.MixProject do
  defp deps do
    [{:phoenix, "~> 1.7"}, {:ecto, "~> 3.11"}]
  end
end`,
			wantNeedsGit:          false,
			wantNeedsSSH:          false,
			wantGitURLType:        "",
			wantRequiredBuildArgs: []string{},
		},
		{
			name:                  "No mix.exs file",
			mixContent:            "",
			wantNeedsGit:          false,
			wantNeedsSSH:          false,
			wantGitURLType:        "",
			wantRequiredBuildArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.mixContent != "" {
				mixPath := filepath.Join(tmpDir, "mix.exs")
				if err := os.WriteFile(mixPath, []byte(tt.mixContent), 0644); err != nil {
					t.Fatalf("failed to write mix.exs: %v", err)
				}
			}

			got := DetectPrivateRepos(tmpDir, "elixir")

			if got.NeedsGit != tt.wantNeedsGit {
				t.Errorf("NeedsGit = %v, want %v", got.NeedsGit, tt.wantNeedsGit)
			}
			if got.NeedsSSH != tt.wantNeedsSSH {
				t.Errorf("NeedsSSH = %v, want %v", got.NeedsSSH, tt.wantNeedsSSH)
			}
			if got.GitURLType != tt.wantGitURLType {
				t.Errorf("GitURLType = %q, want %q", got.GitURLType, tt.wantGitURLType)
			}

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

func TestDetectScalaPrivateRepos(t *testing.T) {
	tests := []struct {
		name                  string
		buildSbtContent       string
		wantRequiredBuildArgs []string
	}{
		{
			name: "build.sbt with resolvers — private repo detected",
			buildSbtContent: `name := "myapp"
scalaVersion := "3.6.4"
resolvers += "My Private Repo" at "https://nexus.example.com/repository/maven/"`,
			wantRequiredBuildArgs: []string{"MAVEN_USERNAME", "MAVEN_PASSWORD"},
		},
		{
			name: "build.sbt without resolvers — no private repos",
			buildSbtContent: `name := "myapp"
scalaVersion := "3.6.4"
libraryDependencies += "org.scalatest" %% "scalatest" % "3.2.18" % Test`,
			wantRequiredBuildArgs: []string{},
		},
		{
			name:                  "no build.sbt — no private repos",
			buildSbtContent:       "",
			wantRequiredBuildArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.buildSbtContent != "" {
				sbtPath := filepath.Join(tmpDir, "build.sbt")
				if err := os.WriteFile(sbtPath, []byte(tt.buildSbtContent), 0644); err != nil {
					t.Fatalf("failed to write build.sbt: %v", err)
				}
			}

			got := DetectPrivateRepos(tmpDir, "scala")

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

func TestDetectSwiftPrivateRepos(t *testing.T) {
	tests := []struct {
		name                  string
		packageSwiftContent   string // empty means no Package.swift
		wantNeedsGit          bool
		wantNeedsSSH          bool
		wantGitURLType        string
		wantRequiredBuildArgs []string
	}{
		{
			name: "Private package via HTTPS",
			packageSwiftContent: `// swift-tools-version:6.1
import PackageDescription
let package = Package(
    name: "MyApp",
    dependencies: [
        .package(url: "https://github.com/myorg/private-lib.git", from: "1.0.0"),
    ]
)`,
			wantNeedsGit:          true,
			wantNeedsSSH:          false,
			wantGitURLType:        "https",
			wantRequiredBuildArgs: []string{"GITHUB_TOKEN"},
		},
		{
			name: "Private package via SSH",
			packageSwiftContent: `// swift-tools-version:6.1
import PackageDescription
let package = Package(
    name: "MyApp",
    dependencies: [
        .package(url: "git@github.com:myorg/private-lib.git", from: "1.0.0"),
    ]
)`,
			wantNeedsGit:          true,
			wantNeedsSSH:          true,
			wantGitURLType:        "ssh",
			wantRequiredBuildArgs: []string{},
		},
		{
			name: "Mixed HTTPS and SSH packages",
			packageSwiftContent: `// swift-tools-version:6.1
import PackageDescription
let package = Package(
    name: "MyApp",
    dependencies: [
        .package(url: "https://github.com/myorg/lib-a.git", from: "1.0.0"),
        .package(url: "git@github.com:myorg/lib-b.git", from: "2.0.0"),
    ]
)`,
			wantNeedsGit:          true,
			wantNeedsSSH:          true,
			wantGitURLType:        "mixed",
			wantRequiredBuildArgs: []string{"GITHUB_TOKEN"},
		},
		{
			name: "No private deps — no git package URLs",
			packageSwiftContent: `// swift-tools-version:6.1
import PackageDescription
let package = Package(
    name: "MyApp",
    targets: [
        .executableTarget(name: "MyApp"),
    ]
)`,
			wantNeedsGit:          false,
			wantNeedsSSH:          false,
			wantGitURLType:        "",
			wantRequiredBuildArgs: []string{},
		},
		{
			name:                  "No Package.swift file",
			packageSwiftContent:   "",
			wantNeedsGit:          false,
			wantNeedsSSH:          false,
			wantGitURLType:        "",
			wantRequiredBuildArgs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			if tt.packageSwiftContent != "" {
				pkgPath := filepath.Join(tmpDir, "Package.swift")
				if err := os.WriteFile(pkgPath, []byte(tt.packageSwiftContent), 0644); err != nil {
					t.Fatalf("failed to write Package.swift: %v", err)
				}
			}

			got := DetectPrivateRepos(tmpDir, "swift")

			if got.NeedsGit != tt.wantNeedsGit {
				t.Errorf("NeedsGit = %v, want %v", got.NeedsGit, tt.wantNeedsGit)
			}
			if got.NeedsSSH != tt.wantNeedsSSH {
				t.Errorf("NeedsSSH = %v, want %v", got.NeedsSSH, tt.wantNeedsSSH)
			}
			if got.GitURLType != tt.wantGitURLType {
				t.Errorf("GitURLType = %q, want %q", got.GitURLType, tt.wantGitURLType)
			}

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
