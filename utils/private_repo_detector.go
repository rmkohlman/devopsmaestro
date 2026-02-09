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
	GitURLType        string // "https", "ssh", or "mixed"
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
	sshGitPattern := regexp.MustCompile(`git\+ssh://|git@`)
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

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
