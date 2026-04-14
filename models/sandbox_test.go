package models

import (
	"sort"
	"testing"
)

func TestGetPreset_AllLanguages(t *testing.T) {
	languages := []string{"python", "golang", "rust", "node", "cpp", "dotnet", "php", "kotlin", "scala", "elixir", "swift", "zig", "dart", "lua", "r", "haskell", "perl", "jupyter"}
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
		{"csharp", "dotnet"},
		{"cs", "dotnet"},
		{"fsharp", "dotnet"},
		{"fs", "dotnet"},
		{"kt", "kotlin"},
		{"sbt", "scala"},
		{"ex", "elixir"},
		{"flutter", "dart"},
		{"luajit", "lua"},
		{"rlang", "r"},
		{"rmd", "r"},
		{"hs", "haskell"},
		{"ghc", "haskell"},
		{"pl", "perl"},
		{"notebook", "jupyter"},
		{"ipynb", "jupyter"},
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

func TestBaseImage_Dotnet(t *testing.T) {
	preset, _ := GetPreset("dotnet")
	img := preset.BaseImage("9.0")
	if img != "mcr.microsoft.com/dotnet/sdk:9.0" {
		t.Errorf("BaseImage(\"9.0\") = %q, want \"mcr.microsoft.com/dotnet/sdk:9.0\"", img)
	}
}

func TestBaseImage_Php(t *testing.T) {
	preset, ok := GetPreset("php")
	if !ok {
		t.Fatal("GetPreset(\"php\") returned false")
	}
	img := preset.BaseImage("8.3")
	if img != "php:8.3-cli-alpine" {
		t.Errorf("BaseImage(\"8.3\") = %q, want \"php:8.3-cli-alpine\"", img)
	}
}

func TestBaseImage_Kotlin(t *testing.T) {
	preset, ok := GetPreset("kotlin")
	if !ok {
		t.Fatal("GetPreset(\"kotlin\") returned false")
	}
	img := preset.BaseImage("21")
	if img != "eclipse-temurin:21-jdk-noble" {
		t.Errorf("BaseImage(\"21\") = %q, want \"eclipse-temurin:21-jdk-noble\"", img)
	}
}

func TestBaseImage_Kotlin_Alias(t *testing.T) {
	preset, ok := GetPreset("kt")
	if !ok {
		t.Fatal("GetPreset(\"kt\") returned false")
	}
	if preset.Language != "kotlin" {
		t.Errorf("GetPreset(\"kt\").Language = %q, want \"kotlin\"", preset.Language)
	}
}

func TestBaseImage_Elixir(t *testing.T) {
	preset, ok := GetPreset("elixir")
	if !ok {
		t.Fatal("GetPreset(\"elixir\") returned false")
	}
	img := preset.BaseImage("1.17")
	if img != "elixir:1.17-slim" {
		t.Errorf("BaseImage(\"1.17\") = %q, want \"elixir:1.17-slim\"", img)
	}
}

func TestBaseImage_Elixir_Alias(t *testing.T) {
	preset, ok := GetPreset("ex")
	if !ok {
		t.Fatal("GetPreset(\"ex\") returned false")
	}
	if preset.Language != "elixir" {
		t.Errorf("GetPreset(\"ex\").Language = %q, want \"elixir\"", preset.Language)
	}
}

func TestBaseImage_Scala(t *testing.T) {
	preset, ok := GetPreset("scala")
	if !ok {
		t.Fatal("GetPreset(\"scala\") returned false")
	}
	img := preset.BaseImage("21")
	if img != "eclipse-temurin:21-jdk-noble" {
		t.Errorf("BaseImage(\"21\") = %q, want \"eclipse-temurin:21-jdk-noble\"", img)
	}
}

func TestBaseImage_Scala_Alias(t *testing.T) {
	preset, ok := GetPreset("sbt")
	if !ok {
		t.Fatal("GetPreset(\"sbt\") returned false")
	}
	if preset.Language != "scala" {
		t.Errorf("GetPreset(\"sbt\").Language = %q, want \"scala\"", preset.Language)
	}
}

func TestBaseImage_Swift(t *testing.T) {
	preset, ok := GetPreset("swift")
	if !ok {
		t.Fatal("GetPreset(\"swift\") returned false")
	}
	img := preset.BaseImage("6.0")
	if img != "swift:6.0-slim" {
		t.Errorf("BaseImage(\"6.0\") = %q, want \"swift:6.0-slim\"", img)
	}
}

func TestBaseImage_Zig(t *testing.T) {
	preset, ok := GetPreset("zig")
	if !ok {
		t.Fatal("GetPreset(\"zig\") returned false")
	}
	// Zig has no official Docker image; BaseImageTemplate is a fixed string
	img := preset.BaseImage("0.14")
	if img != "ubuntu:22.04" {
		t.Errorf("BaseImage(\"0.14\") = %q, want \"ubuntu:22.04\"", img)
	}
}

func TestBaseImage_Dart(t *testing.T) {
	preset, ok := GetPreset("dart")
	if !ok {
		t.Fatal("GetPreset(\"dart\") returned false")
	}
	img := preset.BaseImage("3.7")
	if img != "dart:3.7" {
		t.Errorf("BaseImage(\"3.7\") = %q, want \"dart:3.7\"", img)
	}
}

func TestBaseImage_R(t *testing.T) {
	preset, ok := GetPreset("r")
	if !ok {
		t.Fatal("GetPreset(\"r\") returned false")
	}
	img := preset.BaseImage("4.5")
	if img != "r-base:4.5" {
		t.Errorf("BaseImage(\"4.5\") = %q, want \"r-base:4.5\"", img)
	}
}

func TestBaseImage_Haskell(t *testing.T) {
	preset, ok := GetPreset("haskell")
	if !ok {
		t.Fatal("GetPreset(\"haskell\") returned false")
	}
	img := preset.BaseImage("9.12")
	if img != "haskell:9.12-slim" {
		t.Errorf("BaseImage(\"9.12\") = %q, want \"haskell:9.12-slim\"", img)
	}
}

func TestBaseImage_Perl(t *testing.T) {
	preset, ok := GetPreset("perl")
	if !ok {
		t.Fatal("GetPreset(\"perl\") returned false")
	}
	img := preset.BaseImage("5.40")
	if img != "perl:5.40-slim" {
		t.Errorf("BaseImage(\"5.40\") = %q, want \"perl:5.40-slim\"", img)
	}
}

func TestBaseImage_Jupyter(t *testing.T) {
	preset, ok := GetPreset("jupyter")
	if !ok {
		t.Fatal("GetPreset(\"jupyter\") returned false")
	}
	img := preset.BaseImage("3.13")
	if img != "python:3.13-slim" {
		t.Errorf("BaseImage(\"3.13\") = %q, want \"python:3.13-slim\"", img)
	}
}

func TestJupyterPreset_Details(t *testing.T) {
	preset, ok := GetPreset("jupyter")
	if !ok {
		t.Fatal("GetPreset(\"jupyter\") returned false")
	}
	if preset.DefaultVersion != "3.13" {
		t.Errorf("jupyter DefaultVersion = %q, want \"3.13\"", preset.DefaultVersion)
	}
	if preset.DepsInstallCmd != "pip install euporie jupyter" {
		t.Errorf("jupyter DepsInstallCmd = %q, want \"pip install euporie jupyter\"", preset.DepsInstallCmd)
	}
	// Verify versions
	wantVersions := []string{"3.14", "3.13", "3.12"}
	if len(preset.Versions) != len(wantVersions) {
		t.Fatalf("jupyter Versions = %v, want %v", preset.Versions, wantVersions)
	}
	for i, v := range wantVersions {
		if preset.Versions[i] != v {
			t.Errorf("jupyter Versions[%d] = %q, want %q", i, preset.Versions[i], v)
		}
	}
}

func TestJupyterPreset_Aliases(t *testing.T) {
	for _, alias := range []string{"notebook", "ipynb"} {
		preset, ok := GetPreset(alias)
		if !ok {
			t.Errorf("GetPreset(%q) returned false", alias)
			continue
		}
		if preset.Language != "jupyter" {
			t.Errorf("GetPreset(%q).Language = %q, want \"jupyter\"", alias, preset.Language)
		}
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
	if len(presets) < 10 {
		t.Errorf("ListPresets() returned %d presets, expected at least 10", len(presets))
	}

	sort.Strings(presets)
	expected := []string{"cpp", "dart", "dotnet", "elixir", "golang", "haskell", "jupyter", "kotlin", "lua", "node", "perl", "php", "python", "r", "rust", "scala", "swift", "zig"}
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
