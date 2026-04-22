//go:build !integration

package appkind_test

import (
	"os"
	"path/filepath"
	"testing"

	"devopsmaestro/models"
	"devopsmaestro/utils/appkind"
)

// helper: create a temp dir with the given files/dirs.
// dirs must end with "/".
func scaffold(t *testing.T, entries map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, content := range entries {
		fullPath := filepath.Join(dir, name)
		if len(name) > 0 && name[len(name)-1] == '/' {
			if err := os.MkdirAll(fullPath, 0755); err != nil {
				t.Fatalf("mkdir %s: %v", fullPath, err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
				t.Fatalf("mkdirAll for %s: %v", fullPath, err)
			}
			if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
				t.Fatalf("write %s: %v", fullPath, err)
			}
		}
	}
	return dir
}

func app(name string) *models.App {
	return &models.App{Name: name}
}

// ---------------------------------------------------------------------------
// Signal matrix — each signal should independently produce KindCICD
// ---------------------------------------------------------------------------

func TestAppKindDetector_SignalMatrix(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		appName  string
		wantKind appkind.Kind
	}{
		{
			name: "signal1_chart_yaml_with_apiVersion_and_name",
			files: map[string]string{
				"Chart.yaml": "apiVersion: v2\nname: my-chart\nversion: 1.0.0\n",
			},
			wantKind: appkind.KindCICD,
		},
		{
			name: "signal2_kustomization_yaml",
			files: map[string]string{
				"kustomization.yaml": "resources:\n  - deployment.yaml\n",
			},
			wantKind: appkind.KindCICD,
		},
		{
			name: "signal2_kustomization_yml",
			files: map[string]string{
				"kustomization.yml": "resources:\n  - deployment.yaml\n",
			},
			wantKind: appkind.KindCICD,
		},
		{
			name: "signal2_Kustomization_exact_case",
			files: map[string]string{
				"Kustomization": "resources:\n  - deployment.yaml\n",
			},
			wantKind: appkind.KindCICD,
		},
		{
			name: "signal3_argocd_directory",
			files: map[string]string{
				".argocd/": "",
			},
			wantKind: appkind.KindCICD,
		},
		{
			name: "signal3_Application_kind_in_root_yaml",
			files: map[string]string{
				"app.yaml": "apiVersion: argoproj.io/v1alpha1\nkind: Application\n",
			},
			wantKind: appkind.KindCICD,
		},
		{
			name: "signal4_HelmRelease_kind_in_root_yaml",
			files: map[string]string{
				"release.yaml": "apiVersion: helm.toolkit.fluxcd.io/v2\nkind: HelmRelease\n",
			},
			wantKind: appkind.KindCICD,
		},
		{
			name:     "signal5_name_heuristic_argocd",
			files:    map[string]string{},
			appName:  "my-argocd-config",
			wantKind: appkind.KindUnknown, // signal 5 alone must NOT decide
		},
		{
			name:     "signal5_name_heuristic_flux",
			files:    map[string]string{},
			appName:  "flux-manifests",
			wantKind: appkind.KindUnknown, // signal 5 alone must NOT decide
		},
		{
			name:     "signal5_name_heuristic_gitops",
			files:    map[string]string{},
			appName:  "gitops-platform",
			wantKind: appkind.KindUnknown, // signal 5 alone must NOT decide
		},
		{
			name: "signal6_yaml_only_repo_alone_no_signal5",
			files: map[string]string{
				"deployment.yaml": "apiVersion: apps/v1\nkind: Deployment\n",
				"service.yaml":    "apiVersion: v1\nkind: Service\n",
			},
			appName:  "generic-app",
			wantKind: appkind.KindUnknown, // yaml-only + no signal 5 → NOT KindCICD
		},
		{
			name: "signal6_yaml_only_with_signal5_name",
			files: map[string]string{
				"deployment.yaml": "apiVersion: apps/v1\nkind: Deployment\n",
			},
			appName:  "gitops-infra",
			wantKind: appkind.KindCICD, // yaml-only + signal 5 → KindCICD
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := scaffold(t, tt.files)
			name := tt.appName
			if name == "" {
				name = "test-app"
			}
			a := app(name)

			got, _, err := appkind.Detect(dir, a, "auto")
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if got != tt.wantKind {
				t.Errorf("Detect() = %q, want %q", got, tt.wantKind)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Precedence tie-break: Helm chart + Python files → KindCICD wins
// ---------------------------------------------------------------------------

func TestAppKindDetector_Precedence_HelmWithPython_IsCICD(t *testing.T) {
	dir := scaffold(t, map[string]string{
		"Chart.yaml":       "apiVersion: v2\nname: my-chart\nversion: 1.0.0\n",
		"hooks/install.py": "#!/usr/bin/env python3\nprint('hook')\n",
		"requirements.txt": "requests==2.31.0\n",
	})

	got, _, err := appkind.Detect(dir, app("helm-with-python"), "auto")
	if err != nil {
		t.Fatalf("Detect() error = %v", err)
	}
	if got != appkind.KindCICD {
		t.Errorf("Helm chart with Python files: got %q, want %q (deployable artifact wins)", got, appkind.KindCICD)
	}
}

// ---------------------------------------------------------------------------
// spec.kind override
// ---------------------------------------------------------------------------

func TestAppKindDetector_SpecKindOverride(t *testing.T) {
	tests := []struct {
		name     string
		files    map[string]string
		specKind string
		wantKind appkind.Kind
	}{
		{
			name:     "spec_kind_cicd_overrides_no_signals",
			files:    map[string]string{"main.go": "package main\n"},
			specKind: "cicd",
			wantKind: appkind.KindCICD,
		},
		{
			name: "spec_kind_language_overrides_cicd_signals",
			files: map[string]string{
				"Chart.yaml": "apiVersion: v2\nname: my-chart\nversion: 1.0.0\n",
			},
			specKind: "language",
			wantKind: appkind.KindLanguage,
		},
		{
			name: "spec_kind_auto_runs_detection",
			files: map[string]string{
				"Chart.yaml": "apiVersion: v2\nname: my-chart\nversion: 1.0.0\n",
			},
			specKind: "auto",
			wantKind: appkind.KindCICD,
		},
		{
			name:     "spec_kind_empty_defaults_to_auto",
			files:    map[string]string{"Chart.yaml": "apiVersion: v2\nname: my-chart\n"},
			specKind: "",
			wantKind: appkind.KindCICD,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := scaffold(t, tt.files)
			got, _, err := appkind.Detect(dir, app("override-test"), tt.specKind)
			if err != nil {
				t.Fatalf("Detect() error = %v", err)
			}
			if got != tt.wantKind {
				t.Errorf("Detect(specKind=%q) = %q, want %q", tt.specKind, got, tt.wantKind)
			}
		})
	}
}
