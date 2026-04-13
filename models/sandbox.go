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
func (p SandboxPreset) BaseImage(version string) string {
	return fmt.Sprintf(p.BaseImageTemplate, version)
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
