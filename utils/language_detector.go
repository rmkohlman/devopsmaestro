package utils

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// semverRegex matches the first semantic version pattern (MAJOR.MINOR or MAJOR.MINOR.PATCH)
var semverRegex = regexp.MustCompile(`(\d+\.\d+(?:\.\d+)?)`)

// extractVersion extracts a semantic version from raw file content.
// If the content is empty or contains no version pattern, returns the fallback.
func extractVersion(raw, fallback string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	if m := semverRegex.FindString(raw); m != "" {
		return m
	}
	return fallback
}

// Language represents a detected programming language
type Language struct {
	Name    string
	Version string
	Files   []string // Files that indicated this language
}

// DetectLanguage attempts to detect the primary language of an app
func DetectLanguage(appPath string) (*Language, error) {
	// Check for language-specific files
	indicators := map[string][]string{
		"python":  {"requirements.txt", "setup.py", "pyproject.toml", "Pipfile", "*.py"},
		"golang":  {"go.mod", "go.sum", "*.go"},
		"nodejs":  {"package.json", "package-lock.json", "node_modules"},
		"rust":    {"Cargo.toml", "Cargo.lock", "*.rs"},
		"ruby":    {"Gemfile", "Gemfile.lock", "*.rb"},
		"java":    {"pom.xml", "build.gradle", "*.java"},
		"dotnet":  {"*.csproj", "*.sln", "*.fsproj", "*.cs", "*.fs", "global.json"},
		"gleam":   {"gleam.toml", "*.gleam"},
		"php":     {"composer.json", "composer.lock", "*.php", "artisan"},
		"kotlin":  {"build.gradle.kts", "*.kt", "*.kts"},
		"scala":   {"build.sbt", "*.scala"},
		"elixir":  {"mix.exs", "mix.lock", "*.ex", "*.exs"},
		"swift":   {"Package.swift", "*.swift"},
		"zig":     {"build.zig", "build.zig.zon", "*.zig"},
		"dart":    {"pubspec.yaml", "pubspec.lock", "*.dart"},
		"lua":     {"*.lua", ".luarc.json", ".luacheckrc"},
		"r":       {"DESCRIPTION", "NAMESPACE", "*.R", "*.Rmd", ".Rprofile"},
		"haskell": {"*.cabal", "stack.yaml", "*.hs", "cabal.project"},
		"perl":    {"Makefile.PL", "Build.PL", "cpanfile", "*.pl", "*.pm"},
	}

	detected := make(map[string]int)

	// Walk the app directory
	err := filepath.Walk(appPath, func(path string, info os.FileInfo, err error) error {
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

// HasDockerfile checks if an app has a Dockerfile
func HasDockerfile(appPath string) (bool, string) {
	dockerfilePaths := []string{
		"Dockerfile",
		"dockerfile",
		"Dockerfile.prod",
		"Dockerfile.production",
	}

	for _, name := range dockerfilePaths {
		path := filepath.Join(appPath, name)
		if _, err := os.Stat(path); err == nil {
			return true, path
		}
	}

	return false, ""
}

// DetectVersion attempts to detect the version of a language/runtime
func DetectVersion(lang, appPath string) string {
	switch lang {
	case "python":
		return detectPythonVersion(appPath)
	case "golang":
		return detectGoVersion(appPath)
	case "nodejs":
		return detectNodeVersion(appPath)
	case "dotnet":
		return detectDotnetVersion(appPath)
	case "php":
		return detectPhpVersion(appPath)
	case "kotlin":
		return detectKotlinVersion(appPath)
	case "elixir":
		return detectElixirVersion(appPath)
	case "scala":
		return detectScalaVersion(appPath)
	case "swift":
		return detectSwiftVersion(appPath)
	case "zig":
		return detectZigVersion(appPath)
	case "dart":
		return detectDartVersion(appPath)
	case "lua":
		return detectLuaVersion(appPath)
	case "r":
		return detectRVersion(appPath)
	case "haskell":
		return detectHaskellVersion(appPath)
	case "perl":
		return detectPerlVersion(appPath)
	default:
		return ""
	}
}

func detectPythonVersion(appPath string) string {
	const fallback = "3.11"

	// Check for .python-version file
	versionFile := filepath.Join(appPath, ".python-version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return extractVersion(string(data), fallback)
	}

	// Check pyproject.toml or setup.py for python_requires
	// For now, return a sensible default
	return fallback
}

func detectGoVersion(appPath string) string {
	// Check go.mod for go version
	goMod := filepath.Join(appPath, "go.mod")
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

func detectNodeVersion(appPath string) string {
	const fallback = "20"

	// Check .nvmrc
	nvmrc := filepath.Join(appPath, ".nvmrc")
	if data, err := os.ReadFile(nvmrc); err == nil {
		return extractVersion(string(data), fallback)
	}

	return fallback
}

// dotnetTargetFrameworkRegex extracts the version from <TargetFramework>netX.Y</TargetFramework>.
var dotnetTargetFrameworkRegex = regexp.MustCompile(`<TargetFramework>net(\d+\.\d+)</TargetFramework>`)

// detectDotnetVersion checks global.json for SDK version (highest priority),
// then *.csproj for TargetFramework, then falls back to default.
func detectDotnetVersion(appPath string) string {
	const fallback = "9.0"

	// Priority 1: global.json — SDK version pinning
	if ver := detectDotnetVersionFromGlobalJSON(appPath); ver != "" {
		return ver
	}

	// Priority 2: *.csproj — TargetFramework
	if ver := detectDotnetVersionFromCsproj(appPath); ver != "" {
		return ver
	}

	return fallback
}

// detectDotnetVersionFromGlobalJSON parses global.json for "sdk": { "version": "X.Y.Z" }.
// Returns MAJOR.MINOR (e.g., "9.0" from "9.0.100") or empty string if not found.
func detectDotnetVersionFromGlobalJSON(appPath string) string {
	data, err := os.ReadFile(filepath.Join(appPath, "global.json"))
	if err != nil {
		return ""
	}

	// Simple line-by-line parsing for "version": "X.Y.Z"
	// Avoids importing encoding/json for a single field extraction.
	versionRegex := regexp.MustCompile(`"version"\s*:\s*"(\d+\.\d+)(?:\.\d+)*"`)
	if m := versionRegex.FindStringSubmatch(string(data)); len(m) >= 2 {
		return m[1]
	}
	return ""
}

// detectDotnetVersionFromCsproj scans for *.csproj files and extracts TargetFramework.
// Returns the version (e.g., "8.0" from "<TargetFramework>net8.0</TargetFramework>")
// or empty string if not found.
func detectDotnetVersionFromCsproj(appPath string) string {
	matches, err := filepath.Glob(filepath.Join(appPath, "*.csproj"))
	if err != nil || len(matches) == 0 {
		return ""
	}

	// Use the first .csproj found
	data, err := os.ReadFile(matches[0])
	if err != nil {
		return ""
	}

	if m := dotnetTargetFrameworkRegex.FindStringSubmatch(string(data)); len(m) >= 2 {
		return m[1]
	}
	return ""
}

// phpRequireRegex matches "php": ">=X.Y" or "php": "^X.Y" style constraints in composer.json.
var phpRequireRegex = regexp.MustCompile(`"php"\s*:\s*"([^"]+)"`)

// detectPhpVersion checks composer.json for a PHP version constraint (require.php),
// then falls back to .php-version file, then to default.
func detectPhpVersion(appPath string) string {
	const fallback = "8.4"

	// Priority 1: composer.json — "require": { "php": ">=8.2" }
	composerFile := filepath.Join(appPath, "composer.json")
	if data, err := os.ReadFile(composerFile); err == nil {
		if m := phpRequireRegex.FindStringSubmatch(string(data)); len(m) >= 2 {
			return extractVersion(m[1], fallback)
		}
	}

	// Priority 2: .php-version file
	versionFile := filepath.Join(appPath, ".php-version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return extractVersion(string(data), fallback)
	}

	return fallback
}

// kotlinJvmTargetRegex matches jvmTarget = "X" or jvmTarget = "X.Y" in build.gradle.kts.
var kotlinJvmTargetRegex = regexp.MustCompile(`jvmTarget\s*=\s*"(\d+(?:\.\d+)?)"`)

// detectKotlinVersion checks build.gradle.kts for the JVM target version.
// The JVM target determines the base image version (eclipse-temurin).
// Falls back to "21" (latest LTS JDK).
func detectKotlinVersion(appPath string) string {
	const fallback = "21"

	gradleFile := filepath.Join(appPath, "build.gradle.kts")
	data, err := os.ReadFile(gradleFile)
	if err != nil {
		return fallback
	}

	text := string(data)

	// Priority 1: jvmTarget — this directly determines the JDK version needed
	if m := kotlinJvmTargetRegex.FindStringSubmatch(text); len(m) >= 2 {
		return m[1]
	}

	// Priority 2: If no jvmTarget, fall back to default
	return fallback
}

// elixirVersionRegex matches elixir: "~> X.Y" or elixir: ">= X.Y" style constraints in mix.exs.
var elixirVersionRegex = regexp.MustCompile(`elixir:\s*"([^"]+)"`)

// detectElixirVersion checks mix.exs for an elixir version constraint,
// then .tool-versions for an elixir version, then falls back to default.
func detectElixirVersion(appPath string) string {
	const fallback = "1.18"

	// Priority 1: mix.exs — elixir: "~> 1.16"
	mixFile := filepath.Join(appPath, "mix.exs")
	if data, err := os.ReadFile(mixFile); err == nil {
		if m := elixirVersionRegex.FindStringSubmatch(string(data)); len(m) >= 2 {
			return extractVersion(m[1], fallback)
		}
	}

	// Priority 2: .tool-versions file (asdf format)
	toolVersionsFile := filepath.Join(appPath, ".tool-versions")
	if data, err := os.ReadFile(toolVersionsFile); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "elixir ") {
				parts := strings.Fields(line)
				if len(parts) >= 2 {
					return extractVersion(parts[1], fallback)
				}
			}
		}
	}

	return fallback
}

// scalaVersionRegex matches scalaVersion := "X.Y.Z" in build.sbt.
var scalaVersionRegex = regexp.MustCompile(`scalaVersion\s*:=\s*"(\d+\.\d+(?:\.\d+)?)"`)

// detectScalaVersion checks build.sbt for a Scala version declaration.
// Since Scala runs on the JVM, we need the JDK version for the base image,
// not the Scala language version. Returns the JDK version (default "21").
func detectScalaVersion(appPath string) string {
	const fallback = "21" // JDK version for base image

	buildSbt := filepath.Join(appPath, "build.sbt")
	if data, err := os.ReadFile(buildSbt); err == nil {
		if m := scalaVersionRegex.FindStringSubmatch(string(data)); len(m) >= 2 {
			scalaVer := m[1]
			// Scala 2.x projects typically target older JDKs
			if strings.HasPrefix(scalaVer, "2.") {
				return "17"
			}
		}
	}

	return fallback
}

// swiftToolsVersionRegex matches the swift-tools-version comment at the top of Package.swift.
// Example: // swift-tools-version:5.10 or // swift-tools-version: 6.1
var swiftToolsVersionRegex = regexp.MustCompile(`//\s*swift-tools-version:\s*(\d+\.\d+(?:\.\d+)?)`)

// detectSwiftVersion checks .swift-version (swiftenv convention) and Package.swift
// for the swift-tools-version comment. Falls back to "6.1".
func detectSwiftVersion(appPath string) string {
	const fallback = "6.1"

	// Check .swift-version file (swiftenv convention)
	versionFile := filepath.Join(appPath, ".swift-version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return extractVersion(string(data), fallback)
	}

	// Check Package.swift for swift-tools-version comment
	packageSwift := filepath.Join(appPath, "Package.swift")
	if data, err := os.ReadFile(packageSwift); err == nil {
		if m := swiftToolsVersionRegex.FindStringSubmatch(string(data)); len(m) >= 2 {
			return m[1]
		}
	}

	return fallback
}

// zigMinVersionRegex matches .minimum_zig_version in build.zig.zon.
// Example: .minimum_zig_version = "0.13.0",
var zigMinVersionRegex = regexp.MustCompile(`\.minimum_zig_version\s*=\s*"(\d+\.\d+(?:\.\d+)?)"`)

// detectZigVersion checks .zigversion file and build.zig.zon for
// .minimum_zig_version. Falls back to "0.14".
func detectZigVersion(appPath string) string {
	const fallback = "0.14"

	// Check .zigversion file
	versionFile := filepath.Join(appPath, ".zigversion")
	if data, err := os.ReadFile(versionFile); err == nil {
		return extractVersion(string(data), fallback)
	}

	// Check build.zig.zon for .minimum_zig_version
	buildZigZon := filepath.Join(appPath, "build.zig.zon")
	if data, err := os.ReadFile(buildZigZon); err == nil {
		if m := zigMinVersionRegex.FindStringSubmatch(string(data)); len(m) >= 2 {
			return m[1]
		}
	}

	return fallback
}

// dartSDKConstraintRegex matches the SDK constraint in pubspec.yaml.
// Example: sdk: ">=3.6.0 <4.0.0" or sdk: ^3.7.0 or sdk: '>=3.5.0 <4.0.0'
var dartSDKConstraintRegex = regexp.MustCompile(`sdk:\s*["'^>=]*(\d+\.\d+(?:\.\d+)?)`)

// detectDartVersion checks pubspec.yaml for the environment.sdk constraint
// and .dart_tool/version file. Falls back to "3.7".
func detectDartVersion(appPath string) string {
	const fallback = "3.7"

	// Check pubspec.yaml for environment.sdk constraint
	pubspecPath := filepath.Join(appPath, "pubspec.yaml")
	if data, err := os.ReadFile(pubspecPath); err == nil {
		// Look for sdk constraint under environment section
		lines := strings.Split(string(data), "\n")
		inEnvironment := false
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "environment:" {
				inEnvironment = true
				continue
			}
			// A non-indented line (not starting with space) ends the environment section
			if inEnvironment && len(line) > 0 && line[0] != ' ' && line[0] != '\t' {
				break
			}
			if inEnvironment && strings.Contains(trimmed, "sdk:") {
				if m := dartSDKConstraintRegex.FindStringSubmatch(trimmed); len(m) >= 2 {
					return m[1]
				}
			}
		}
	}

	// Check .dart_tool/version file
	dartToolVersion := filepath.Join(appPath, ".dart_tool", "version")
	if data, err := os.ReadFile(dartToolVersion); err == nil {
		return extractVersion(string(data), fallback)
	}

	return fallback
}

// detectLuaVersion checks .lua-version file for the Lua version.
// Falls back to "5.4".
func detectLuaVersion(appPath string) string {
	const fallback = "5.4"

	// Check .lua-version file
	versionFile := filepath.Join(appPath, ".lua-version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return extractVersion(string(data), fallback)
	}

	return fallback
}

// rDependsRegex matches "Depends: R (>= X.Y)" or "Depends: R (>= X.Y.Z)" in DESCRIPTION files.
// Example: Depends: R (>= 4.1.0) → captures "4.1.0"
var rDependsRegex = regexp.MustCompile(`Depends:\s*R\s*\(\s*>=\s*(\d+\.\d+(?:\.\d+)?)\s*\)`)

// detectRVersion checks the DESCRIPTION file for "Depends: R (>= X.Y)"
// and .R-version file. Falls back to "4.5".
func detectRVersion(appPath string) string {
	const fallback = "4.5"

	// Check DESCRIPTION file for Depends: R (>= X.Y)
	descPath := filepath.Join(appPath, "DESCRIPTION")
	if data, err := os.ReadFile(descPath); err == nil {
		if m := rDependsRegex.FindStringSubmatch(string(data)); len(m) >= 2 {
			return m[1]
		}
	}

	// Check .R-version file
	versionFile := filepath.Join(appPath, ".R-version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return extractVersion(string(data), fallback)
	}

	return fallback
}

// stackResolverRegex matches "resolver: ghc-X.Y.Z" in stack.yaml.
var stackResolverRegex = regexp.MustCompile(`resolver:\s*ghc-(\d+\.\d+(?:\.\d+)?)`)

// detectHaskellVersion checks stack.yaml for a GHC resolver version,
// then .ghc-version file. Falls back to "9.12".
func detectHaskellVersion(appPath string) string {
	const fallback = "9.12"

	// Check stack.yaml for resolver: ghc-X.Y.Z
	stackYaml := filepath.Join(appPath, "stack.yaml")
	if data, err := os.ReadFile(stackYaml); err == nil {
		if m := stackResolverRegex.FindStringSubmatch(string(data)); len(m) >= 2 {
			return m[1]
		}
		// Try extracting any version from the resolver line
		for _, line := range strings.Split(string(data), "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "resolver:") {
				if v := extractVersion(line, ""); v != "" {
					return v
				}
			}
		}
	}

	// Check .ghc-version file
	versionFile := filepath.Join(appPath, ".ghc-version")
	if data, err := os.ReadFile(versionFile); err == nil {
		return extractVersion(string(data), fallback)
	}

	return fallback
}

// cpanfileRequiresRegex matches "requires 'perl', '>= X.Y'" or similar in cpanfile.
// Example: requires 'perl', '>= 5.36.0'; → captures "5.36.0"
var cpanfileRequiresRegex = regexp.MustCompile(`requires\s+['"]perl['"]\s*,\s*['"][>=<!\s]*(\d+\.\d+(?:\.\d+)?)\s*['"]`)

// detectPerlVersion checks .perl-version and cpanfile for a Perl version.
// Falls back to "5.40".
func detectPerlVersion(appPath string) string {
	const fallback = "5.40"

	// Check .perl-version file first
	perlVersionFile := filepath.Join(appPath, ".perl-version")
	if data, err := os.ReadFile(perlVersionFile); err == nil {
		return extractVersion(string(data), fallback)
	}

	// Check cpanfile for requires 'perl' line
	cpanfile := filepath.Join(appPath, "cpanfile")
	if data, err := os.ReadFile(cpanfile); err == nil {
		if m := cpanfileRequiresRegex.FindStringSubmatch(string(data)); len(m) >= 2 {
			return m[1]
		}
	}

	return fallback
}
