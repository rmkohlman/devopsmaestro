// Package appkind provides app kind detection for DevOpsMaestro workspaces.
// An AppKind is a stable, declarative property of the app — it describes
// what the workspace is *for* (develop a language service, operate a CICD
// manifest set, etc.), one level above language detection.
//
// Detection runs a registry of signal-based detectors in priority order,
// per the precedence table in #404 §3:
//
//	Rank 1: Chart.yaml at root with apiVersion + name (Helm)
//	Rank 2: kustomization.yaml / .yml / Kustomization (Kustomize)
//	Rank 3: .argocd/ directory OR `kind: Application` in any root yaml (Argo)
//	Rank 4: flux-system/ OR `kind: HelmRelease` in root yaml (Flux)
//	Rank 5: app name contains argocd|flux|gitops (weak heuristic — never alone)
//	Rank 6: yaml-only repo (no language indicators) — only counts when paired
//	        with signal 5
//
// spec.kind on the App resource overrides auto-detection: "cicd"|"language"
// short-circuit immediately; "auto" or "" runs the registry.
package appkind

import (
	"strings"

	"devopsmaestro/models"
)

// Kind identifies the high-level category of an app workspace.
type Kind string

const (
	// KindCICD indicates a GitOps / Helm / Kustomize / Argo / Flux workspace.
	// The deployable artifact is YAML manifests, not a compiled language binary.
	KindCICD Kind = "cicd"

	// KindLanguage indicates a standard language development workspace.
	// Language detection runs normally to pick the compiler/interpreter.
	KindLanguage Kind = "language"

	// KindUnknown means no signal fired; caller falls back to generic ubuntu image.
	KindUnknown Kind = "unknown"
)

// Evidence records which signals fired during detection (for logging/debugging).
type Evidence struct {
	Signals []string
}

// Detector detects the app kind for a given source path and app configuration.
// Implementations are responsible for one signal each (Strategy pattern); the
// top-level Detect() composes them via the registry.
type Detector interface {
	// Detect returns true and a label if the signal fires for this path/app.
	Detect(path string, app *models.App) (fired bool, label string)
}

// Detect is the top-level entry point: consults specKind override first,
// then runs the signal registry per the precedence rules in §3 of #404.
//
// specKind values:
//
//	"cicd"     → return KindCICD immediately
//	"language" → return KindLanguage immediately
//	"auto"|""  → run signal registry
func Detect(path string, app *models.App, specKind string) (Kind, Evidence, error) {
	switch strings.ToLower(strings.TrimSpace(specKind)) {
	case "cicd":
		return KindCICD, Evidence{Signals: []string{"spec.kind=cicd"}}, nil
	case "language":
		return KindLanguage, Evidence{Signals: []string{"spec.kind=language"}}, nil
	case "", "auto":
		// fall through to detection
	default:
		// unrecognized value — treat as auto
	}

	ev := Evidence{}

	// Signals 1–4 are structurally unambiguous: any single hit decides KindCICD.
	hardSignals := []Detector{
		signalChartYAML{},
		signalKustomization{},
		signalArgoCD{},
		signalFlux{},
	}
	for _, d := range hardSignals {
		if fired, label := d.Detect(path, app); fired {
			ev.Signals = append(ev.Signals, label)
			return KindCICD, ev, nil
		}
	}

	// Signal 5 (name heuristic) and Signal 6 (yaml-only) are weak.
	// Per §3: signal 5 alone must NOT decide. signal 6 + signal 5 → KindCICD.
	nameHit, nameLabel := signalNameHeuristic{}.Detect(path, app)
	yamlOnly, yamlLabel := signalYAMLOnly{}.Detect(path, app)
	if yamlOnly && nameHit {
		ev.Signals = append(ev.Signals, yamlLabel, nameLabel)
		return KindCICD, ev, nil
	}

	return KindUnknown, ev, nil
}
