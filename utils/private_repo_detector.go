package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// PrivateRepoInfo contains information about private repository usage
type PrivateRepoInfo struct {
	NeedsGit          bool
	NeedsGitConfig    bool
	NeedsSSH          bool
	RequiredBuildArgs []string
	DetectedInFiles   []string
	GitURLType        string            // "https", "ssh", or "mixed"
	SystemDeps        []string          // System packages needed by Python C extensions
	SystemDepSources  map[string]string // Maps system dep to Python package that requires it (for comments)
}

// DetectPrivateRepos scans app files for private repository references
func DetectPrivateRepos(appPath, language string) *PrivateRepoInfo {
	info := &PrivateRepoInfo{
		RequiredBuildArgs: []string{},
		DetectedInFiles:   []string{},
	}

	switch language {
	case "python":
		detectPythonPrivateRepos(appPath, info)
	case "golang":
		detectGoPrivateRepos(appPath, info)
	case "nodejs":
		detectNodePrivateRepos(appPath, info)
	case "java":
		detectJavaPrivateRepos(appPath, info)
	case "rust":
		detectRustPrivateRepos(appPath, info)
	}

	return info
}

func detectPythonPrivateRepos(appPath string, info *PrivateRepoInfo) {
	reqFile := filepath.Join(appPath, "requirements.txt")
	content, err := os.ReadFile(reqFile)
	if err != nil {
		return
	}

	text := string(content)

	// Pattern: git+https://${VAR}:${VAR}@github.com or git+ssh://git@github.com
	httpsGitPattern := regexp.MustCompile(`git\+https?://`)
	// Match git+ssh:// anywhere, or git@ only when NOT preceded by a dot
	// (prevents ".git@v1.0" in HTTPS URLs like repo.git@v1.0 from false-matching)
	sshGitPattern := regexp.MustCompile(`git\+ssh://|[^.]git@|^git@`)
	varPattern := regexp.MustCompile(`\$\{([^}]+)\}`)

	hasHTTPS := httpsGitPattern.MatchString(text)
	hasSSH := sshGitPattern.MatchString(text)

	if hasHTTPS || hasSSH {
		info.NeedsGit = true
		info.DetectedInFiles = append(info.DetectedInFiles, "requirements.txt")

		if hasSSH {
			info.NeedsSSH = true
			if hasHTTPS {
				info.GitURLType = "mixed"
			} else {
				info.GitURLType = "ssh"
			}
		} else {
			info.GitURLType = "https"
		}

		// Extract variable names (for HTTPS with tokens)
		if hasHTTPS {
			matches := varPattern.FindAllStringSubmatch(text, -1)
			seen := make(map[string]bool)
			for _, match := range matches {
				if len(match) > 1 {
					varName := match[1]
					if !seen[varName] {
						info.RequiredBuildArgs = append(info.RequiredBuildArgs, varName)
						seen[varName] = true
					}
				}
			}
		}
	}

	// Detect system dependencies needed by Python C extensions
	detectPythonSystemDeps(appPath, info)
}

func detectGoPrivateRepos(appPath string, info *PrivateRepoInfo) {
	goModFile := filepath.Join(appPath, "go.mod")
	content, err := os.ReadFile(goModFile)
	if err != nil {
		return
	}

	text := string(content)

	// Check for SSH URLs (git@github.com:)
	sshPattern := regexp.MustCompile(`git@[a-zA-Z0-9.-]+:`)

	// Look for private GitHub/GitLab repos in require statements
	// Pattern: github.com/company/private-repo (not in public repos like golang.org)
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "require") || strings.HasPrefix(line, "replace") {
			// Check for SSH URLs
			if sshPattern.MatchString(line) {
				info.NeedsGit = true
				info.NeedsSSH = true
				info.GitURLType = "ssh"
				info.DetectedInFiles = append(info.DetectedInFiles, "go.mod")
				break
			}

			// Check for common private repo patterns (HTTPS)
			if strings.Contains(line, "github.com/") && !strings.Contains(line, "golang.org") {
				// Likely private if it's a company repo
				info.NeedsGit = true
				info.NeedsGitConfig = true
				info.DetectedInFiles = append(info.DetectedInFiles, "go.mod")

				if info.GitURLType == "" {
					info.GitURLType = "https"
				}

				// Go typically uses GITHUB_TOKEN or GOPRIVATE
				if !contains(info.RequiredBuildArgs, "GITHUB_TOKEN") {
					info.RequiredBuildArgs = append(info.RequiredBuildArgs, "GITHUB_TOKEN")
				}
				break
			}
		}
	}
}

func detectNodePrivateRepos(appPath string, info *PrivateRepoInfo) {
	pkgFile := filepath.Join(appPath, "package.json")
	content, err := os.ReadFile(pkgFile)
	if err != nil {
		return
	}

	text := string(content)

	// Pattern: git+https:// or @company/package (scoped packages)
	gitPattern := regexp.MustCompile(`"git\+https?://`)
	scopedPattern := regexp.MustCompile(`"@[a-zA-Z0-9-]+/`)
	varPattern := regexp.MustCompile(`\$\{([^}]+)\}`)

	if gitPattern.MatchString(text) || scopedPattern.MatchString(text) {
		info.NeedsGit = true
		info.DetectedInFiles = append(info.DetectedInFiles, "package.json")

		// Extract variable names
		matches := varPattern.FindAllStringSubmatch(text, -1)
		seen := make(map[string]bool)
		for _, match := range matches {
			if len(match) > 1 {
				varName := match[1]
				if !seen[varName] {
					info.RequiredBuildArgs = append(info.RequiredBuildArgs, varName)
					seen[varName] = true
				}
			}
		}

		// Common token for npm
		if len(info.RequiredBuildArgs) == 0 {
			info.RequiredBuildArgs = append(info.RequiredBuildArgs, "NPM_TOKEN")
		}
	}
}

func detectJavaPrivateRepos(appPath string, info *PrivateRepoInfo) {
	pomFile := filepath.Join(appPath, "pom.xml")
	if _, err := os.Stat(pomFile); err == nil {
		content, err := os.ReadFile(pomFile)
		if err != nil {
			return
		}

		// Look for private repository declarations
		if strings.Contains(string(content), "<repository>") {
			info.DetectedInFiles = append(info.DetectedInFiles, "pom.xml")
			info.RequiredBuildArgs = append(info.RequiredBuildArgs, "MAVEN_USERNAME", "MAVEN_PASSWORD")
		}
	}

	// Check for Gradle
	gradleFile := filepath.Join(appPath, "build.gradle")
	if _, err := os.Stat(gradleFile); err == nil {
		content, err := os.ReadFile(gradleFile)
		if err != nil {
			return
		}

		if strings.Contains(string(content), "maven {") {
			info.DetectedInFiles = append(info.DetectedInFiles, "build.gradle")
			info.RequiredBuildArgs = append(info.RequiredBuildArgs, "MAVEN_USERNAME", "MAVEN_PASSWORD")
		}
	}
}

func detectRustPrivateRepos(appPath string, info *PrivateRepoInfo) {
	cargoFile := filepath.Join(appPath, "Cargo.toml")
	content, err := os.ReadFile(cargoFile)
	if err != nil {
		return
	}

	text := string(content)

	// Pattern: git = "https://github.com/..." or git = "ssh://git@github.com/..."
	httpsGitPattern := regexp.MustCompile(`git\s*=\s*"https?://`)
	sshGitPattern := regexp.MustCompile(`git\s*=\s*"(ssh://git@|git@)`)

	hasHTTPS := httpsGitPattern.MatchString(text)
	hasSSH := sshGitPattern.MatchString(text)

	if hasHTTPS || hasSSH {
		info.NeedsGit = true
		info.DetectedInFiles = append(info.DetectedInFiles, "Cargo.toml")

		if hasSSH {
			info.NeedsSSH = true
			if hasHTTPS {
				info.GitURLType = "mixed"
			} else {
				info.GitURLType = "ssh"
			}
		} else {
			info.GitURLType = "https"
			// Rust typically uses SSH keys or CARGO_NET_GIT_FETCH_WITH_CLI
			// For MVP, we'll use GITHUB_TOKEN
			if !contains(info.RequiredBuildArgs, "GITHUB_TOKEN") {
				info.RequiredBuildArgs = append(info.RequiredBuildArgs, "GITHUB_TOKEN")
			}
		}
	}
}

// pepNormRe matches runs of [-_.] for PEP 503 package name normalization.
var pepNormRe = regexp.MustCompile(`[-_.]+`)

// normalizePythonPkgName normalizes a Python package name per PEP 503:
// lowercase and replace all runs of [-_.] with a single dash.
func normalizePythonPkgName(name string) string {
	if name == "" {
		return ""
	}
	lower := strings.ToLower(name)
	return pepNormRe.ReplaceAllString(lower, "-")
}

// extractPkgName extracts the package name from a requirements.txt line,
// stripping version specifiers, extras, and comments.
func extractPkgName(line string) string {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
		return ""
	}

	// Handle @ syntax: "mylib @ git+https://..."
	if atIdx := strings.Index(line, " @ "); atIdx != -1 {
		return strings.TrimSpace(line[:atIdx])
	}

	// Strip version specifiers and extras
	for i, ch := range line {
		if ch == '[' || ch == '>' || ch == '<' || ch == '=' || ch == '!' || ch == '~' || ch == ';' {
			return strings.TrimSpace(line[:i])
		}
	}

	return strings.TrimSpace(line)
}

// pythonSystemDepsMap maps normalized Python package names to the system packages
// (Debian/Ubuntu) they require for building C extensions.
var pythonSystemDepsMap = map[string][]string{
	"psycopg2":     {"libpq-dev"},
	"mysqlclient":  {"default-libmysqlclient-dev"},
	"pillow":       {"libjpeg-dev", "zlib1g-dev", "libfreetype6-dev"},
	"lxml":         {"libxml2-dev", "libxslt1-dev"},
	"cryptography": {"libffi-dev", "libssl-dev"},
	"cffi":         {"libffi-dev", "libssl-dev"},
	"pyyaml":       {"libyaml-dev"},
	"python-ldap":  {"libldap2-dev", "libsasl2-dev"},
	"gevent":       {"libev-dev", "libevent-dev"},
	"pycairo":      {"libcairo2-dev", "pkg-config"},
	"h5py":         {"libhdf5-dev"},
}

// detectPythonSystemDeps scans requirements.txt for Python packages that need
// system library headers (e.g., psycopg2 -> libpq-dev) and populates
// info.SystemDeps and info.SystemDepSources.
func detectPythonSystemDeps(appPath string, info *PrivateRepoInfo) {
	reqFile := filepath.Join(appPath, "requirements.txt")
	content, err := os.ReadFile(reqFile)
	if err != nil {
		return
	}

	seen := make(map[string]bool)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		pkgName := extractPkgName(line)
		if pkgName == "" {
			continue
		}

		normalized := normalizePythonPkgName(pkgName)
		deps, ok := pythonSystemDepsMap[normalized]
		if !ok {
			continue
		}

		for _, dep := range deps {
			if !seen[dep] {
				seen[dep] = true
				info.SystemDeps = append(info.SystemDeps, dep)
				if info.SystemDepSources == nil {
					info.SystemDepSources = make(map[string]string)
				}
				// First package that needs this dep wins the source attribution
				if _, exists := info.SystemDepSources[dep]; !exists {
					info.SystemDepSources[dep] = normalized
				}
			}
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
