package utils

import (
	"os"
	"path/filepath"
	"strings"
)

// Language represents a detected programming language
type Language struct {
	Name    string
	Version string
	Files   []string // Files that indicated this language
}

// DetectLanguage attempts to detect the primary language of a project
func DetectLanguage(projectPath string) (*Language, error) {
	// Check for language-specific files
	indicators := map[string][]string{
		"python": {"requirements.txt", "setup.py", "pyproject.toml", "Pipfile", "*.py"},
		"golang": {"go.mod", "go.sum", "*.go"},
		"nodejs": {"package.json", "package-lock.json", "node_modules"},
		"rust":   {"Cargo.toml", "Cargo.lock", "*.rs"},
		"ruby":   {"Gemfile", "Gemfile.lock", "*.rb"},
		"java":   {"pom.xml", "build.gradle", "*.java"},
	}

	detected := make(map[string]int)

	// Walk the project directory
	err := filepath.Walk(projectPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden directories and common non-source directories
		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") ||
				name == "node_modules" ||
				name == "vendor" ||
				name == "__pycache__" ||
				name == "venv" ||
				name == "target" {
				return filepath.SkipDir
			}
			return nil
		}

		// Check against language indicators
		for lang, patterns := range indicators {
			for _, pattern := range patterns {
				matched, _ := filepath.Match(pattern, info.Name())
				if matched {
					detected[lang]++
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Find the language with the most indicators
	var bestLang string
	var maxCount int
	for lang, count := range detected {
		if count > maxCount {
			maxCount = count
			bestLang = lang
		}
	}

	if bestLang == "" {
		return nil, nil // No language detected
	}

	return &Language{
		Name: bestLang,
	}, nil
}

// HasDockerfile checks if a project has a Dockerfile
func HasDockerfile(projectPath string) (bool, string) {
	dockerfilePaths := []string{
		"Dockerfile",
		"dockerfile",
		"Dockerfile.prod",
		"Dockerfile.production",
	}

	for _, name := range dockerfilePaths {
		path := filepath.Join(projectPath, name)
		if _, err := os.Stat(path); err == nil {
			return true, path
		}
	}

	return false, ""
}

// DetectVersion attempts to detect the version of a language/runtime
func DetectVersion(lang, projectPath string) string {
	switch lang {
	case "python":
		return detectPythonVersion(projectPath)
	case "golang":
		return detectGoVersion(projectPath)
	case "nodejs":
		return detectNodeVersion(projectPath)
	default:
		return ""
	}
}

func detectPythonVersion(projectPath string) string {
	// Check for .python-version file
	versionFile := filepath.Join(projectPath, ".python-version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return strings.TrimSpace(string(data))
	}

	// Check pyproject.toml or setup.py for python_requires
	// For now, return a sensible default
	return "3.11"
}

func detectGoVersion(projectPath string) string {
	// Check go.mod for go version
	goMod := filepath.Join(projectPath, "go.mod")
	if data, err := os.ReadFile(goMod); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), "go ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					return parts[1]
				}
			}
		}
	}

	return "1.22"
}

func detectNodeVersion(projectPath string) string {
	// Check .nvmrc
	nvmrc := filepath.Join(projectPath, ".nvmrc")
	if data, err := os.ReadFile(nvmrc); err == nil {
		return strings.TrimSpace(string(data))
	}

	return "20"
}
