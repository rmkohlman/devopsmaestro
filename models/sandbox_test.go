package models

import (
	"sort"
	"testing"
)

func TestGetPreset_AllLanguages(t *testing.T) {
	languages := []string{"python", "golang", "rust", "node", "cpp"}
	for _, lang := range languages {
		preset, ok := GetPreset(lang)
		if !ok {
			t.Errorf("GetPreset(%q) returned false", lang)
			continue
		}
		if preset.Language != lang {
			t.Errorf("GetPreset(%q).Language = %q", lang, preset.Language)
		}
		if len(preset.Versions) == 0 {
			t.Errorf("GetPreset(%q).Versions is empty", lang)
		}
		if preset.DefaultVersion == "" {
			t.Errorf("GetPreset(%q).DefaultVersion is empty", lang)
		}
		if preset.BaseImageTemplate == "" {
			t.Errorf("GetPreset(%q).BaseImageTemplate is empty", lang)
		}
	}
}

func TestGetPreset_Aliases(t *testing.T) {
	tests := []struct {
		alias    string
		expected string
	}{
		{"py", "python"},
		{"python3", "python"},
		{"go", "golang"},
		{"rs", "rust"},
		{"nodejs", "node"},
		{"javascript", "node"},
		{"js", "node"},
		{"c++", "cpp"},
		{"c", "cpp"},
		{"gcc", "cpp"},
	}

	for _, tt := range tests {
		preset, ok := GetPreset(tt.alias)
		if !ok {
			t.Errorf("GetPreset(%q) returned false", tt.alias)
			continue
		}
		if preset.Language != tt.expected {
			t.Errorf("GetPreset(%q).Language = %q, want %q", tt.alias, preset.Language, tt.expected)
		}
	}
}

func TestGetPreset_Unknown(t *testing.T) {
	_, ok := GetPreset("fortran")
	if ok {
		t.Error("GetPreset(\"fortran\") should return false")
	}
}

func TestGetPreset_CaseInsensitive(t *testing.T) {
	preset, ok := GetPreset("PYTHON")
	if !ok {
		t.Error("GetPreset(\"PYTHON\") should be case-insensitive")
	}
	if preset.Language != "python" {
		t.Errorf("GetPreset(\"PYTHON\").Language = %q", preset.Language)
	}
}

func TestBaseImage(t *testing.T) {
	preset, _ := GetPreset("python")
	img := preset.BaseImage("3.12")
	if img != "python:3.12-slim" {
		t.Errorf("BaseImage(\"3.12\") = %q, want \"python:3.12-slim\"", img)
	}
}

func TestDefaultVersionInVersionList(t *testing.T) {
	for _, lang := range ListPresets() {
		preset, _ := GetPreset(lang)
		found := false
		for _, v := range preset.Versions {
			if v == preset.DefaultVersion {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s: DefaultVersion %q not in Versions %v", lang, preset.DefaultVersion, preset.Versions)
		}
	}
}

func TestListPresets(t *testing.T) {
	presets := ListPresets()
	if len(presets) < 5 {
		t.Errorf("ListPresets() returned %d presets, expected at least 5", len(presets))
	}

	sort.Strings(presets)
	expected := []string{"cpp", "golang", "node", "python", "rust"}
	sort.Strings(expected)
	for _, e := range expected {
		found := false
		for _, p := range presets {
			if p == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ListPresets() missing %q", e)
		}
	}
}
