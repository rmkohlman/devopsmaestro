package appkind

import (
	"os"
	"path/filepath"
	"strings"

	"devopsmaestro/models"
)

// signalChartYAML — Rank 1: Helm Chart.yaml at root with apiVersion: + name:
type signalChartYAML struct{}

func (signalChartYAML) Detect(path string, _ *models.App) (bool, string) {
	data, err := os.ReadFile(filepath.Join(path, "Chart.yaml"))
	if err != nil {
		return false, ""
	}
	s := string(data)
	if hasYAMLKey(s, "apiVersion") && hasYAMLKey(s, "name") {
		return true, "signal1:Chart.yaml"
	}
	return false, ""
}

// signalKustomization — Rank 2: kustomization.yaml | kustomization.yml | Kustomization
type signalKustomization struct{}

func (signalKustomization) Detect(path string, _ *models.App) (bool, string) {
	for _, name := range []string{"kustomization.yaml", "kustomization.yml", "Kustomization"} {
		if _, err := os.Stat(filepath.Join(path, name)); err == nil {
			return true, "signal2:" + name
		}
	}
	return false, ""
}

// signalArgoCD — Rank 3: .argocd/ dir OR `kind: Application` in any root yaml
type signalArgoCD struct{}

func (signalArgoCD) Detect(path string, _ *models.App) (bool, string) {
	if info, err := os.Stat(filepath.Join(path, ".argocd")); err == nil && info.IsDir() {
		return true, "signal3:.argocd/"
	}
	if hit := scanRootYAMLForKind(path, "Application"); hit != "" {
		return true, "signal3:Application(" + hit + ")"
	}
	return false, ""
}

// signalFlux — Rank 4: flux-system/ dir OR `kind: HelmRelease` in root yaml
type signalFlux struct{}

func (signalFlux) Detect(path string, _ *models.App) (bool, string) {
	if info, err := os.Stat(filepath.Join(path, "flux-system")); err == nil && info.IsDir() {
		return true, "signal4:flux-system/"
	}
	if hit := scanRootYAMLForKind(path, "HelmRelease"); hit != "" {
		return true, "signal4:HelmRelease(" + hit + ")"
	}
	return false, ""
}

// signalNameHeuristic — Rank 5: app name contains argocd|flux|gitops (case-insensitive)
// Per §3 this MUST NOT decide alone — caller pairs it with signal 6.
type signalNameHeuristic struct{}

func (signalNameHeuristic) Detect(_ string, app *models.App) (bool, string) {
	if app == nil {
		return false, ""
	}
	n := strings.ToLower(app.Name)
	for _, kw := range []string{"argocd", "flux", "gitops"} {
		if strings.Contains(n, kw) {
			return true, "signal5:name~" + kw
		}
	}
	return false, ""
}

// signalYAMLOnly — Rank 6: repo contains *.yaml/*.yml files and no source matching
// language indicators. Walks the root only (not recursive — avoid pathological costs).
type signalYAMLOnly struct{}

func (signalYAMLOnly) Detect(path string, _ *models.App) (bool, string) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, ""
	}
	hasYAML := false
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".yaml", ".yml":
			hasYAML = true
		case ".go", ".py", ".js", ".ts", ".rs", ".java", ".kt", ".rb", ".php",
			".cs", ".cpp", ".c", ".swift", ".scala", ".ex", ".exs", ".dart",
			".lua", ".pl", ".hs", ".r", ".zig":
			return false, ""
		}
		// Common language manifest files at root → not yaml-only.
		switch name {
		case "go.mod", "package.json", "requirements.txt", "pyproject.toml",
			"Cargo.toml", "pom.xml", "build.gradle", "build.gradle.kts",
			"Gemfile", "composer.json", "mix.exs", "Package.swift":
			return false, ""
		}
	}
	if hasYAML {
		return true, "signal6:yaml-only"
	}
	return false, ""
}

// hasYAMLKey returns true if `key:` appears at column 0 on any line
// (sufficient for detecting top-level Chart.yaml fields).
func hasYAMLKey(content, key string) bool {
	prefix := key + ":"
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimRight(line, "\r")
		if strings.HasPrefix(trimmed, prefix) {
			return true
		}
	}
	return false
}

// scanRootYAMLForKind walks root-level *.yaml/*.yml files (non-recursive) looking
// for `kind: <wanted>`. Returns the first matching filename or "".
func scanRootYAMLForKind(path, wanted string) string {
	entries, err := os.ReadDir(path)
	if err != nil {
		return ""
	}
	target := "kind: " + wanted
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(path, e.Name()))
		if err != nil {
			continue
		}
		for _, line := range strings.Split(string(data), "\n") {
			if strings.TrimSpace(strings.TrimRight(line, "\r")) == target {
				return e.Name()
			}
		}
	}
	return ""
}
