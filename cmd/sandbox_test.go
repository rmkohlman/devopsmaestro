package cmd

import (
	"devopsmaestro/models"
	"testing"
)

func TestGenerateSandboxName(t *testing.T) {
	tests := []struct {
		lang string
	}{
		{"python"},
		{"golang"},
		{"rust"},
		{"node"},
		{"cpp"},
	}

	for _, tt := range tests {
		name := generateSandboxName(tt.lang)
		prefix := "dvm-sandbox-" + tt.lang + "-"
		if len(name) <= len(prefix) {
			t.Errorf("generateSandboxName(%q) = %q, too short", tt.lang, name)
		}
		if name[:len(prefix)] != prefix {
			t.Errorf("generateSandboxName(%q) = %q, missing prefix %q", tt.lang, name, prefix)
		}
	}

	// Verify uniqueness (probabilistic — two 4-hex-char IDs have ~1/65536 chance of collision)
	a := generateSandboxName("python")
	b := generateSandboxName("python")
	if a == b {
		t.Logf("WARNING: two generated names are identical: %q (unlikely but possible)", a)
	}
}

func TestBuildSandboxLabels(t *testing.T) {
	labels := buildSandboxLabels("python", "3.12", "my-sandbox")

	expected := map[string]string{
		"dvm.sandbox":         "true",
		"dvm.sandbox.lang":    "python",
		"dvm.sandbox.version": "3.12",
		"dvm.sandbox.name":    "my-sandbox",
	}

	for k, v := range expected {
		got, ok := labels[k]
		if !ok {
			t.Errorf("missing label %q", k)
			continue
		}
		if got != v {
			t.Errorf("label %q = %q, want %q", k, got, v)
		}
	}
}

func TestIsValidVersion(t *testing.T) {
	tests := []struct {
		lang    string
		version string
		want    bool
	}{
		{"python", "3.13", true},
		{"python", "3.12", true},
		{"python", "2.7", false},
		{"golang", "1.24", true},
		{"golang", "1.20", false},
		{"node", "22", true},
		{"node", "16", false},
	}

	for _, tt := range tests {
		preset, ok := models.GetPreset(tt.lang)
		if !ok {
			t.Fatalf("preset %q not found", tt.lang)
		}
		got := isValidVersion(preset, tt.version)
		if got != tt.want {
			t.Errorf("isValidVersion(%q, %q) = %v, want %v", tt.lang, tt.version, got, tt.want)
		}
	}
}
