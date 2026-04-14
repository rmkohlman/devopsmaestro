package models

import (
	"fmt"
	"strings"
)

// SandboxPreset defines a language runtime preset for ephemeral sandbox containers.
// Each preset specifies the base image, supported versions, and dependency handling.
type SandboxPreset struct {
	Language          string   // Canonical language name (e.g., "python", "golang")
	Versions          []string // Supported versions, newest first
	DefaultVersion    string   // Default version when none specified
	BaseImageTemplate string   // Base image with %s placeholder for version
	DepsFiles         []string // Recognized dependency file names
	DepsInstallCmd    string   // Install command template with %s for file path
}

// BaseImage returns the base image for a given version.
// If the template contains %s, it is substituted with the version.
// Otherwise, the template is returned as-is (for languages without versioned Docker images).
func (p SandboxPreset) BaseImage(version string) string {
	if strings.Contains(p.BaseImageTemplate, "%s") {
		return fmt.Sprintf(p.BaseImageTemplate, version)
	}
	return p.BaseImageTemplate
}

// builtinPresets contains the built-in language sandbox presets.
var builtinPresets = map[string]SandboxPreset{
	"python": {
		Language:          "python",
		Versions:          []string{"3.13", "3.12", "3.11", "3.10"},
		DefaultVersion:    "3.13",
		BaseImageTemplate: "python:%s-slim",
		DepsFiles:         []string{"requirements.txt", "Pipfile", "pyproject.toml"},
		DepsInstallCmd:    "pip install -r %s",
	},
	"golang": {
		Language:          "golang",
		Versions:          []string{"1.24", "1.23", "1.22"},
		DefaultVersion:    "1.24",
		BaseImageTemplate: "golang:%s-bookworm",
		DepsFiles:         []string{"go.mod", "go.sum"},
		DepsInstallCmd:    "go mod download",
	},
	"rust": {
		Language:          "rust",
		Versions:          []string{"1.86", "1.85", "1.84", "1.83"},
		DefaultVersion:    "1.86",
		BaseImageTemplate: "rust:%s-slim-bookworm",
		DepsFiles:         []string{"Cargo.toml", "Cargo.lock"},
		DepsInstallCmd:    "cargo fetch",
	},
	"node": {
		Language:          "node",
		Versions:          []string{"22", "20", "18"},
		DefaultVersion:    "22",
		BaseImageTemplate: "node:%s-slim",
		DepsFiles:         []string{"package.json", "package-lock.json", "yarn.lock"},
		DepsInstallCmd:    "npm install",
	},
	"cpp": {
		Language:          "cpp",
		Versions:          []string{"14", "13", "12"},
		DefaultVersion:    "14",
		BaseImageTemplate: "gcc:%s-bookworm",
		DepsFiles:         []string{"CMakeLists.txt", "Makefile"},
		DepsInstallCmd:    "make -f %s",
	},
	"dotnet": {
		Language:          "dotnet",
		Versions:          []string{"9.0", "8.0"},
		DefaultVersion:    "9.0",
		BaseImageTemplate: "mcr.microsoft.com/dotnet/sdk:%s",
		DepsFiles:         []string{"*.csproj", "*.sln", "NuGet.config"},
		DepsInstallCmd:    "dotnet restore",
	},
	"php": {
		Language:          "php",
		Versions:          []string{"8.4", "8.3", "8.2"},
		DefaultVersion:    "8.4",
		BaseImageTemplate: "php:%s-cli-alpine",
		DepsFiles:         []string{"composer.json", "composer.lock"},
		DepsInstallCmd:    "composer install",
	},
	"kotlin": {
		Language:          "kotlin",
		Versions:          []string{"21", "17"},
		DefaultVersion:    "21",
		BaseImageTemplate: "eclipse-temurin:%s-jdk-noble",
		DepsFiles:         []string{"build.gradle.kts", "settings.gradle.kts"},
		DepsInstallCmd:    "./gradlew dependencies --no-daemon",
	},
	"scala": {
		Language:          "scala",
		Versions:          []string{"21", "17"},
		DefaultVersion:    "21",
		BaseImageTemplate: "eclipse-temurin:%s-jdk-noble",
		DepsFiles:         []string{"build.sbt", "project/build.properties"},
		DepsInstallCmd:    "sbt update",
	},
	"elixir": {
		Language:          "elixir",
		Versions:          []string{"1.18", "1.17", "1.16"},
		DefaultVersion:    "1.18",
		BaseImageTemplate: "elixir:%s-slim",
		DepsFiles:         []string{"mix.exs", "mix.lock"},
		DepsInstallCmd:    "mix deps.get",
	},
	"swift": {
		Language:          "swift",
		Versions:          []string{"6.1", "6.0", "5.10"},
		DefaultVersion:    "6.1",
		BaseImageTemplate: "swift:%s-slim",
		DepsFiles:         []string{"Package.swift", "Package.resolved"},
		DepsInstallCmd:    "swift package resolve",
	},
	"zig": {
		Language:          "zig",
		Versions:          []string{"0.14", "0.13"},
		DefaultVersion:    "0.14",
		BaseImageTemplate: "ubuntu:22.04", // No official Zig Docker image; Zig installed at build time
		DepsFiles:         []string{"build.zig", "build.zig.zon"},
		DepsInstallCmd:    "zig build --fetch",
	},
	"dart": {
		Language:          "dart",
		Versions:          []string{"3.7", "3.6"},
		DefaultVersion:    "3.7",
		BaseImageTemplate: "dart:%s",
		DepsFiles:         []string{"pubspec.yaml", "pubspec.lock"},
		DepsInstallCmd:    "dart pub get",
	},
	"lua": {
		Language:          "lua",
		Versions:          []string{"5.4", "5.3"},
		DefaultVersion:    "5.4",
		BaseImageTemplate: "ubuntu:22.04", // No official Lua Docker image; Lua installed at build time
		DepsFiles:         []string{"*.rockspec"},
		DepsInstallCmd:    "luarocks install --deps-only",
	},
	"r": {
		Language:          "r",
		Versions:          []string{"4.5", "4.4"},
		DefaultVersion:    "4.5",
		BaseImageTemplate: "r-base:%s",
		DepsFiles:         []string{"DESCRIPTION", "renv.lock"},
		DepsInstallCmd:    "Rscript -e 'renv::restore()'",
	},
	"haskell": {
		Language:          "haskell",
		Versions:          []string{"9.12", "9.10", "9.8"},
		DefaultVersion:    "9.12",
		BaseImageTemplate: "haskell:%s-slim",
		DepsFiles:         []string{"*.cabal", "stack.yaml", "cabal.project"},
		DepsInstallCmd:    "cabal update && cabal build --only-dependencies",
	},
	"perl": {
		Language:          "perl",
		Versions:          []string{"5.40", "5.38"},
		DefaultVersion:    "5.40",
		BaseImageTemplate: "perl:%s-slim",
		DepsFiles:         []string{"cpanfile", "Makefile.PL", "Build.PL"},
		DepsInstallCmd:    "cpanm --installdeps .",
	},
	"jupyter": {
		Language:          "jupyter",
		Versions:          []string{"3.14", "3.13", "3.12"},
		DefaultVersion:    "3.13",
		BaseImageTemplate: "python:%s-slim",
		DepsFiles:         []string{"requirements.txt", "Pipfile", "pyproject.toml"},
		DepsInstallCmd:    "pip install euporie jupyter",
	},
}

// presetAliases maps alternative language names to canonical preset keys.
var presetAliases = map[string]string{
	"py":         "python",
	"python3":    "python",
	"go":         "golang",
	"rs":         "rust",
	"nodejs":     "node",
	"javascript": "node",
	"js":         "node",
	"c++":        "cpp",
	"c":          "cpp",
	"gcc":        "cpp",
	"csharp":     "dotnet",
	"cs":         "dotnet",
	"fsharp":     "dotnet",
	"fs":         "dotnet",
	"kt":         "kotlin",
	"sbt":        "scala",
	"ex":         "elixir",
	"flutter":    "dart",
	"luajit":     "lua",
	"rlang":      "r",
	"rmd":        "r",
	"hs":         "haskell",
	"ghc":        "haskell",
	"pl":         "perl",
	"notebook":   "jupyter",
	"ipynb":      "jupyter",
}

// GetPreset returns the sandbox preset for the given language name.
// It supports aliases (e.g., "py" → "python", "go" → "golang").
// Returns the preset and true if found, zero value and false otherwise.
func GetPreset(lang string) (SandboxPreset, bool) {
	lang = strings.ToLower(strings.TrimSpace(lang))

	// Check canonical name first
	if p, ok := builtinPresets[lang]; ok {
		return p, true
	}

	// Check aliases
	if canonical, ok := presetAliases[lang]; ok {
		if p, ok := builtinPresets[canonical]; ok {
			return p, true
		}
	}

	return SandboxPreset{}, false
}

// ListPresets returns the names of all available sandbox presets.
func ListPresets() []string {
	names := make([]string, 0, len(builtinPresets))
	for name := range builtinPresets {
		names = append(names, name)
	}
	return names
}
