//go:build !integration

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetectLanguage_Go verifies that a directory containing go.mod is detected as "golang".
func TestDetectLanguage_Go(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\ngo 1.22\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "golang" {
		t.Errorf("expected Name == %q, got %q", "golang", lang.Name)
	}
}

// TestDetectLanguage_Rust verifies that a directory containing Cargo.toml is detected as "rust".
func TestDetectLanguage_Rust(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"test\"\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write Cargo.toml: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "rust" {
		t.Errorf("expected Name == %q, got %q", "rust", lang.Name)
	}
}

// TestDetectLanguage_Python verifies that a directory containing requirements.txt is detected as "python".
func TestDetectLanguage_Python(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("flask==3.0\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write requirements.txt: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "python" {
		t.Errorf("expected Name == %q, got %q", "python", lang.Name)
	}
}

// TestDetectLanguage_Gleam verifies that a directory containing gleam.toml is detected as "gleam".
func TestDetectLanguage_Gleam(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "gleam.toml"), []byte("name = \"test\"\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write gleam.toml: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language for gleam project")
	}
	if lang.Name != "gleam" {
		t.Errorf("expected Name == %q, got %q", "gleam", lang.Name)
	}
}

// TestDetectLanguage_EmptyDir verifies that an empty directory returns nil, nil
// (no language detected, no error).
func TestDetectLanguage_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error for empty dir: %v", err)
	}
	if lang != nil {
		t.Errorf("expected nil for empty dir, got %+v", lang)
	}
}

// TestDetectLanguage_BareRepo verifies that a directory with no language files
// (simulating a bare git repository) returns nil, nil.
// This documents the expected behavior: bare repos have no source code files,
// so language detection correctly returns nothing.
func TestDetectLanguage_BareRepo(t *testing.T) {
	dir := t.TempDir()

	// Simulate a bare git repo structure — contains only git metadata, no source files
	gitDir := filepath.Join(dir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "HEAD"), []byte("ref: refs/heads/main\n"), 0644); err != nil {
		t.Fatalf("failed to write HEAD: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error for bare repo: %v", err)
	}
	if lang != nil {
		t.Errorf("expected nil for bare repo dir, got %+v", lang)
	}
}

// TestDetectLanguage_NodeJS verifies that a directory containing package.json is detected as "nodejs".
func TestDetectLanguage_NodeJS(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"name":"test","version":"1.0.0"}`+"\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write package.json: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "nodejs" {
		t.Errorf("expected Name == %q, got %q", "nodejs", lang.Name)
	}
}

// TestDetectLanguage_Ruby verifies that a directory containing Gemfile is detected as "ruby".
func TestDetectLanguage_Ruby(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "Gemfile"), []byte("source 'https://rubygems.org'\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write Gemfile: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "ruby" {
		t.Errorf("expected Name == %q, got %q", "ruby", lang.Name)
	}
}

// TestDetectLanguage_Java verifies that a directory containing pom.xml is detected as "java".
func TestDetectLanguage_Java(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "pom.xml"), []byte("<project></project>\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write pom.xml: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "java" {
		t.Errorf("expected Name == %q, got %q", "java", lang.Name)
	}
}

// TestDetectLanguage_MultipleIndicators_GoWins verifies that when a directory contains
// multiple Go indicators (go.mod, go.sum, and .go files) with no other language indicators,
// the language with the most matched files wins and "golang" is detected.
func TestDetectLanguage_MultipleIndicators_GoWins(t *testing.T) {
	dir := t.TempDir()

	// Write go.mod, go.sum, and 3 .go files — 5 golang indicators total
	files := map[string]string{
		"go.mod":  "module test\ngo 1.22\n",
		"go.sum":  "// empty\n",
		"main.go": "package main\n",
		"foo.go":  "package main\n",
		"bar.go":  "package main\n",
	}
	for name, content := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", name, err)
		}
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "golang" {
		t.Errorf("expected Name == %q, got %q", "golang", lang.Name)
	}
}

func TestDetectLanguage_Dotnet(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{"csproj file", "MyApp.csproj", "<Project></Project>\n"},
		{"sln file", "MyApp.sln", "Microsoft Visual Studio Solution File\n"},
		{"cs file", "Program.cs", "using System;\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			err := os.WriteFile(filepath.Join(dir, tt.filename), []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("failed to write %s: %v", tt.filename, err)
			}

			lang, err := DetectLanguage(dir)
			if err != nil {
				t.Fatalf("DetectLanguage returned error: %v", err)
			}
			if lang == nil {
				t.Fatal("DetectLanguage returned nil, expected *Language")
			}
			if lang.Name != "dotnet" {
				t.Errorf("expected Name == %q, got %q", "dotnet", lang.Name)
			}
		})
	}
}

// TestDetectLanguage_PHP verifies that a directory containing composer.json is detected as "php".
func TestDetectLanguage_PHP(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "composer.json"), []byte(`{"require":{"php":">=8.2"}}`+"\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write composer.json: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "php" {
		t.Errorf("expected Name == %q, got %q", "php", lang.Name)
	}
}

// TestDetectLanguage_PHP_Artisan verifies that a directory containing artisan is detected as "php".
func TestDetectLanguage_PHP_Artisan(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "artisan"), []byte("#!/usr/bin/env php\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write artisan: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "php" {
		t.Errorf("expected Name == %q, got %q", "php", lang.Name)
	}
}

// TestDetectLanguage_Kotlin verifies that a directory containing build.gradle.kts
// and *.kt files is detected as "kotlin".
func TestDetectLanguage_Kotlin(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "build.gradle.kts"), []byte("plugins { kotlin(\"jvm\") version \"2.1.0\" }\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write build.gradle.kts: %v", err)
	}
	err = os.WriteFile(filepath.Join(dir, "Main.kt"), []byte("fun main() { println(\"Hello\") }\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write Main.kt: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "kotlin" {
		t.Errorf("expected Name == %q, got %q", "kotlin", lang.Name)
	}
}

// TestDetectLanguage_Elixir verifies that a directory containing mix.exs
// is detected as "elixir".
func TestDetectLanguage_Elixir(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "mix.exs"), []byte("defmodule MyApp.MixProject do\n  use Mix.Project\nend\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write mix.exs: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "elixir" {
		t.Errorf("expected Name == %q, got %q", "elixir", lang.Name)
	}
}

// TestDetectLanguage_Scala verifies that a directory containing build.sbt
// is detected as "scala".
func TestDetectLanguage_Scala(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "build.sbt"), []byte("name := \"myapp\"\nscalaVersion := \"3.6.4\"\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write build.sbt: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "scala" {
		t.Errorf("expected Name == %q, got %q", "scala", lang.Name)
	}
}

// TestDetectLanguage_Swift verifies that a directory containing Package.swift
// is detected as "swift".
func TestDetectLanguage_Swift(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "Package.swift"), []byte("// swift-tools-version:6.1\nimport PackageDescription\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write Package.swift: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "swift" {
		t.Errorf("expected Name == %q, got %q", "swift", lang.Name)
	}
}

// TestDetectSwiftVersion verifies Swift version detection from
// .swift-version file and Package.swift swift-tools-version comment.
func TestDetectSwiftVersion(t *testing.T) {
	tests := []struct {
		name            string
		swiftVersion    string // content of .swift-version (empty = no file)
		packageSwift    string // content of Package.swift (empty = no file)
		expectedVersion string
	}{
		{
			name:            "from .swift-version file",
			swiftVersion:    "5.10\n",
			expectedVersion: "5.10",
		},
		{
			name:            "from Package.swift swift-tools-version",
			packageSwift:    "// swift-tools-version:6.0\nimport PackageDescription\n",
			expectedVersion: "6.0",
		},
		{
			name:            "from Package.swift with space after colon",
			packageSwift:    "// swift-tools-version: 5.9\nimport PackageDescription\n",
			expectedVersion: "5.9",
		},
		{
			name:            ".swift-version takes priority over Package.swift",
			swiftVersion:    "5.10\n",
			packageSwift:    "// swift-tools-version:6.1\nimport PackageDescription\n",
			expectedVersion: "5.10",
		},
		{
			name:            "fallback when no version files exist",
			expectedVersion: "6.1",
		},
		{
			name:            "patch version from Package.swift",
			packageSwift:    "// swift-tools-version:5.9.2\nimport PackageDescription\n",
			expectedVersion: "5.9.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.swiftVersion != "" {
				if err := os.WriteFile(filepath.Join(dir, ".swift-version"), []byte(tt.swiftVersion), 0644); err != nil {
					t.Fatalf("failed to write .swift-version: %v", err)
				}
			}
			if tt.packageSwift != "" {
				if err := os.WriteFile(filepath.Join(dir, "Package.swift"), []byte(tt.packageSwift), 0644); err != nil {
					t.Fatalf("failed to write Package.swift: %v", err)
				}
			}

			got := DetectVersion("swift", dir)
			if got != tt.expectedVersion {
				t.Errorf("DetectVersion(\"swift\", ...) = %q, want %q", got, tt.expectedVersion)
			}
		})
	}
}

func TestDetectLanguage_Zig(t *testing.T) {
	dir := t.TempDir()

	err := os.WriteFile(filepath.Join(dir, "build.zig"), []byte("const std = @import(\"std\");\n"), 0644)
	if err != nil {
		t.Fatalf("failed to write build.zig: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "zig" {
		t.Errorf("expected Name == %q, got %q", "zig", lang.Name)
	}
}

// TestDetectZigVersion verifies Zig version detection from
// .zigversion file and build.zig.zon .minimum_zig_version.
func TestDetectZigVersion(t *testing.T) {
	tests := []struct {
		name            string
		zigVersion      string // content of .zigversion (empty = no file)
		buildZigZon     string // content of build.zig.zon (empty = no file)
		expectedVersion string
	}{
		{
			name:            "from .zigversion file",
			zigVersion:      "0.13.0\n",
			expectedVersion: "0.13.0",
		},
		{
			name:            "from build.zig.zon minimum_zig_version",
			buildZigZon:     ".{\n    .minimum_zig_version = \"0.13.0\",\n}\n",
			expectedVersion: "0.13.0",
		},
		{
			name:            ".zigversion takes priority over build.zig.zon",
			zigVersion:      "0.12.0\n",
			buildZigZon:     ".{\n    .minimum_zig_version = \"0.13.0\",\n}\n",
			expectedVersion: "0.12.0",
		},
		{
			name:            "fallback when no version files exist",
			expectedVersion: "0.14",
		},
		{
			name:            "major.minor from .zigversion",
			zigVersion:      "0.14\n",
			expectedVersion: "0.14",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.zigVersion != "" {
				if err := os.WriteFile(filepath.Join(dir, ".zigversion"), []byte(tt.zigVersion), 0644); err != nil {
					t.Fatalf("failed to write .zigversion: %v", err)
				}
			}
			if tt.buildZigZon != "" {
				if err := os.WriteFile(filepath.Join(dir, "build.zig.zon"), []byte(tt.buildZigZon), 0644); err != nil {
					t.Fatalf("failed to write build.zig.zon: %v", err)
				}
			}

			got := DetectVersion("zig", dir)
			if got != tt.expectedVersion {
				t.Errorf("DetectVersion(\"zig\", ...) = %q, want %q", got, tt.expectedVersion)
			}
		})
	}
}

// TestDetectLanguage_Dart verifies that a directory containing pubspec.yaml
// is detected as "dart".
func TestDetectLanguage_Dart(t *testing.T) {
	dir := t.TempDir()

	pubspec := "name: my_app\nenvironment:\n  sdk: ^3.7.0\n"
	err := os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte(pubspec), 0644)
	if err != nil {
		t.Fatalf("failed to write pubspec.yaml: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "dart" {
		t.Errorf("expected Name == %q, got %q", "dart", lang.Name)
	}
}

// TestDetectDartVersion verifies Dart version detection from
// pubspec.yaml environment.sdk constraint and .dart_tool/version file.
func TestDetectDartVersion(t *testing.T) {
	tests := []struct {
		name            string
		pubspecYaml     string // content of pubspec.yaml (empty = no file)
		dartToolVersion string // content of .dart_tool/version (empty = no file)
		expectedVersion string
	}{
		{
			name:            "from pubspec.yaml sdk caret constraint",
			pubspecYaml:     "name: my_app\nenvironment:\n  sdk: ^3.6.0\n",
			expectedVersion: "3.6.0",
		},
		{
			name:            "from pubspec.yaml sdk range constraint",
			pubspecYaml:     "name: my_app\nenvironment:\n  sdk: \">=3.5.0 <4.0.0\"\n",
			expectedVersion: "3.5.0",
		},
		{
			name:            "from pubspec.yaml sdk quoted constraint",
			pubspecYaml:     "name: my_app\nenvironment:\n  sdk: '>=3.7.0 <4.0.0'\n",
			expectedVersion: "3.7.0",
		},
		{
			name:            "from .dart_tool/version file",
			dartToolVersion: "3.6.1\n",
			expectedVersion: "3.6.1",
		},
		{
			name:            "pubspec.yaml takes priority over .dart_tool/version",
			pubspecYaml:     "name: my_app\nenvironment:\n  sdk: ^3.5.0\n",
			dartToolVersion: "3.7.0\n",
			expectedVersion: "3.5.0",
		},
		{
			name:            "fallback when no version files exist",
			expectedVersion: "3.7",
		},
		{
			name:            "major.minor only from pubspec.yaml",
			pubspecYaml:     "name: my_app\nenvironment:\n  sdk: ^3.7.0\n",
			expectedVersion: "3.7.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.pubspecYaml != "" {
				if err := os.WriteFile(filepath.Join(dir, "pubspec.yaml"), []byte(tt.pubspecYaml), 0644); err != nil {
					t.Fatalf("failed to write pubspec.yaml: %v", err)
				}
			}
			if tt.dartToolVersion != "" {
				dartToolDir := filepath.Join(dir, ".dart_tool")
				if err := os.MkdirAll(dartToolDir, 0755); err != nil {
					t.Fatalf("failed to create .dart_tool dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(dartToolDir, "version"), []byte(tt.dartToolVersion), 0644); err != nil {
					t.Fatalf("failed to write .dart_tool/version: %v", err)
				}
			}

			got := DetectVersion("dart", dir)
			if got != tt.expectedVersion {
				t.Errorf("DetectVersion(\"dart\", ...) = %q, want %q", got, tt.expectedVersion)
			}
		})
	}
}

// TestDetectLanguage_R verifies that a directory containing DESCRIPTION
// and .R files is detected as "r".
func TestDetectLanguage_R(t *testing.T) {
	dir := t.TempDir()

	desc := "Package: mypackage\nTitle: My Package\nDepends: R (>= 4.1.0)\n"
	err := os.WriteFile(filepath.Join(dir, "DESCRIPTION"), []byte(desc), 0644)
	if err != nil {
		t.Fatalf("failed to write DESCRIPTION: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "r" {
		t.Errorf("expected Name == %q, got %q", "r", lang.Name)
	}
}

// TestDetectRVersion verifies R version detection from
// DESCRIPTION file and .R-version file.
func TestDetectRVersion(t *testing.T) {
	tests := []struct {
		name            string
		description     string // content of DESCRIPTION (empty = no file)
		rVersionFile    string // content of .R-version (empty = no file)
		expectedVersion string
	}{
		{
			name:            "from DESCRIPTION Depends field",
			description:     "Package: mypackage\nDepends: R (>= 4.1.0)\n",
			expectedVersion: "4.1.0",
		},
		{
			name:            "from DESCRIPTION Depends major.minor only",
			description:     "Package: mypackage\nDepends: R (>= 4.3)\n",
			expectedVersion: "4.3",
		},
		{
			name:            "from .R-version file",
			rVersionFile:    "4.4.1\n",
			expectedVersion: "4.4.1",
		},
		{
			name:            "DESCRIPTION takes priority over .R-version",
			description:     "Package: mypackage\nDepends: R (>= 4.2.0)\n",
			rVersionFile:    "4.4.0\n",
			expectedVersion: "4.2.0",
		},
		{
			name:            "fallback when no version files exist",
			expectedVersion: "4.5",
		},
		{
			name:            "DESCRIPTION without Depends falls back to .R-version",
			description:     "Package: mypackage\nTitle: My Package\n",
			rVersionFile:    "4.3.2\n",
			expectedVersion: "4.3.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.description != "" {
				if err := os.WriteFile(filepath.Join(dir, "DESCRIPTION"), []byte(tt.description), 0644); err != nil {
					t.Fatalf("failed to write DESCRIPTION: %v", err)
				}
			}
			if tt.rVersionFile != "" {
				if err := os.WriteFile(filepath.Join(dir, ".R-version"), []byte(tt.rVersionFile), 0644); err != nil {
					t.Fatalf("failed to write .R-version: %v", err)
				}
			}

			got := DetectVersion("r", dir)
			if got != tt.expectedVersion {
				t.Errorf("DetectVersion(\"r\", ...) = %q, want %q", got, tt.expectedVersion)
			}
		})
	}
}

// TestDetectLanguage_Haskell verifies that a directory containing *.cabal
// is detected as "haskell".
func TestDetectLanguage_Haskell(t *testing.T) {
	dir := t.TempDir()

	cabal := "name: my-app\nversion: 0.1.0.0\nbuild-depends: base >=4.7 && <5\n"
	err := os.WriteFile(filepath.Join(dir, "my-app.cabal"), []byte(cabal), 0644)
	if err != nil {
		t.Fatalf("failed to write my-app.cabal: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "haskell" {
		t.Errorf("expected Name == %q, got %q", "haskell", lang.Name)
	}
}

// TestDetectHaskellVersion verifies Haskell version detection from
// stack.yaml resolver and .ghc-version file.
func TestDetectHaskellVersion(t *testing.T) {
	tests := []struct {
		name            string
		stackYaml       string // content of stack.yaml (empty = no file)
		ghcVersionFile  string // content of .ghc-version (empty = no file)
		expectedVersion string
	}{
		{
			name:            "from stack.yaml ghc resolver",
			stackYaml:       "resolver: ghc-9.8.4\n",
			expectedVersion: "9.8.4",
		},
		{
			name:            "from stack.yaml lts resolver",
			stackYaml:       "resolver: lts-22.43\n",
			expectedVersion: "22.43",
		},
		{
			name:            "from .ghc-version file",
			ghcVersionFile:  "9.10.1\n",
			expectedVersion: "9.10.1",
		},
		{
			name:            "stack.yaml takes priority over .ghc-version",
			stackYaml:       "resolver: ghc-9.8.4\n",
			ghcVersionFile:  "9.10.1\n",
			expectedVersion: "9.8.4",
		},
		{
			name:            "fallback when no version files exist",
			expectedVersion: "9.12",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.stackYaml != "" {
				if err := os.WriteFile(filepath.Join(dir, "stack.yaml"), []byte(tt.stackYaml), 0644); err != nil {
					t.Fatalf("failed to write stack.yaml: %v", err)
				}
			}
			if tt.ghcVersionFile != "" {
				if err := os.WriteFile(filepath.Join(dir, ".ghc-version"), []byte(tt.ghcVersionFile), 0644); err != nil {
					t.Fatalf("failed to write .ghc-version: %v", err)
				}
			}

			got := DetectVersion("haskell", dir)
			if got != tt.expectedVersion {
				t.Errorf("DetectVersion(\"haskell\", ...) = %q, want %q", got, tt.expectedVersion)
			}
		})
	}
}

// TestDetectLanguage_Perl verifies that a directory containing cpanfile
// is detected as "perl".
func TestDetectLanguage_Perl(t *testing.T) {
	dir := t.TempDir()

	cpanfile := "requires 'Mojolicious', '>= 9.0';\nrequires 'DBI';\n"
	err := os.WriteFile(filepath.Join(dir, "cpanfile"), []byte(cpanfile), 0644)
	if err != nil {
		t.Fatalf("failed to write cpanfile: %v", err)
	}

	lang, err := DetectLanguage(dir)
	if err != nil {
		t.Fatalf("DetectLanguage returned error: %v", err)
	}
	if lang == nil {
		t.Fatal("DetectLanguage returned nil, expected *Language")
	}
	if lang.Name != "perl" {
		t.Errorf("expected Name == %q, got %q", "perl", lang.Name)
	}
}

// TestDetectPerlVersion verifies Perl version detection from
// .perl-version file and cpanfile requires 'perl' line.
func TestDetectPerlVersion(t *testing.T) {
	tests := []struct {
		name            string
		perlVersionFile string // content of .perl-version (empty = no file)
		cpanfileContent string // content of cpanfile (empty = no file)
		expectedVersion string
	}{
		{
			name:            "from .perl-version file",
			perlVersionFile: "5.38.2\n",
			expectedVersion: "5.38.2",
		},
		{
			name:            "from cpanfile requires perl",
			cpanfileContent: "requires 'perl', '>= 5.36.0';\nrequires 'Mojolicious';\n",
			expectedVersion: "5.36.0",
		},
		{
			name:            ".perl-version takes priority over cpanfile",
			perlVersionFile: "5.38.2\n",
			cpanfileContent: "requires 'perl', '>= 5.36.0';\n",
			expectedVersion: "5.38.2",
		},
		{
			name:            "fallback when no version files exist",
			expectedVersion: "5.40",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.perlVersionFile != "" {
				if err := os.WriteFile(filepath.Join(dir, ".perl-version"), []byte(tt.perlVersionFile), 0644); err != nil {
					t.Fatalf("failed to write .perl-version: %v", err)
				}
			}
			if tt.cpanfileContent != "" {
				if err := os.WriteFile(filepath.Join(dir, "cpanfile"), []byte(tt.cpanfileContent), 0644); err != nil {
					t.Fatalf("failed to write cpanfile: %v", err)
				}
			}

			got := DetectVersion("perl", dir)
			if got != tt.expectedVersion {
				t.Errorf("DetectVersion(\"perl\", ...) = %q, want %q", got, tt.expectedVersion)
			}
		})
	}
}
