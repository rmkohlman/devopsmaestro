//go:build !integration

package utils

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDetectPythonVersion verifies that DetectVersion("python", ...) correctly
// extracts a semantic version from various .python-version file contents.
//
// TDD Phase 2: These tests are written BEFORE the fix. Cases with non-standard
// inputs (prefixed, suffixed, lts-style) are expected to FAIL until
// detectPythonVersion() is updated to use regex extraction.
func TestDetectPythonVersion(t *testing.T) {
	const pythonDefault = "3.11"

	tests := []struct {
		name        string
		fileContent *string // nil means "don't create the file"
		expected    string
	}{
		{
			name:        "plain version",
			fileContent: strPtr("3.9.9"),
			expected:    "3.9.9",
		},
		{
			name:        "plain version with newline",
			fileContent: strPtr("3.9.9\n"),
			expected:    "3.9.9",
		},
		{
			name:        "prefixed version daa-api-3.9.9",
			fileContent: strPtr("daa-api-3.9.9"),
			expected:    "3.9.9",
		},
		{
			name:        "prefixed version myproject-3.11.2",
			fileContent: strPtr("myproject-3.11.2"),
			expected:    "3.11.2",
		},
		{
			name:        "suffixed version 3.10.5-dev",
			fileContent: strPtr("3.10.5-dev"),
			expected:    "3.10.5",
		},
		{
			name:        "complex prefix and suffix app-3.11.2-beta",
			fileContent: strPtr("app-3.11.2-beta"),
			expected:    "3.11.2",
		},
		{
			name:        "major.minor only 3.9",
			fileContent: strPtr("3.9"),
			expected:    "3.9",
		},
		{
			name:        "invalid string with no version digits",
			fileContent: strPtr("invalid-no-version"),
			expected:    pythonDefault,
		},
		{
			name:        "empty file",
			fileContent: strPtr(""),
			expected:    pythonDefault,
		},
		{
			name:        "no file exists",
			fileContent: nil,
			expected:    pythonDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.fileContent != nil {
				versionFile := filepath.Join(dir, ".python-version")
				if err := os.WriteFile(versionFile, []byte(*tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to write .python-version: %v", err)
				}
			}

			got := DetectVersion("python", dir)
			if got != tt.expected {
				t.Errorf("DetectVersion(\"python\", dir) = %q, want %q (input: %q)",
					got, tt.expected, safeDeref(tt.fileContent))
			}
		})
	}
}

// TestDetectNodeVersion verifies that DetectVersion("nodejs", ...) correctly
// extracts a semantic version from various .nvmrc file contents.
//
// TDD Phase 2: These tests are written BEFORE the fix. Cases with v-prefixed
// versions (e.g. "v18.17.0") or lts aliases ("lts/*", "lts/hydrogen", "node")
// are expected to FAIL until detectNodeVersion() is updated to strip the "v"
// prefix and return the default for non-numeric aliases.
func TestDetectNodeVersion(t *testing.T) {
	const nodeDefault = "20"

	tests := []struct {
		name        string
		fileContent *string // nil means "don't create the file"
		expected    string
	}{
		{
			name:        "plain numeric version",
			fileContent: strPtr("18.17.0"),
			expected:    "18.17.0",
		},
		{
			name:        "plain numeric version with newline",
			fileContent: strPtr("18.17.0\n"),
			expected:    "18.17.0",
		},
		{
			name:        "v-prefixed version v18.17.0",
			fileContent: strPtr("v18.17.0"),
			expected:    "18.17.0",
		},
		{
			name:        "lts alias lts/*",
			fileContent: strPtr("lts/*"),
			expected:    nodeDefault,
		},
		{
			name:        "lts named alias lts/hydrogen",
			fileContent: strPtr("lts/hydrogen"),
			expected:    nodeDefault,
		},
		{
			name:        "bare node alias",
			fileContent: strPtr("node"),
			expected:    nodeDefault,
		},
		{
			name:        "major.minor only 20.10",
			fileContent: strPtr("20.10"),
			expected:    "20.10",
		},
		{
			name:        "prefixed version myapp-16.14.0",
			fileContent: strPtr("myapp-16.14.0"),
			expected:    "16.14.0",
		},
		{
			name:        "no file exists",
			fileContent: nil,
			expected:    nodeDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			if tt.fileContent != nil {
				nvmrc := filepath.Join(dir, ".nvmrc")
				if err := os.WriteFile(nvmrc, []byte(*tt.fileContent), 0644); err != nil {
					t.Fatalf("failed to write .nvmrc: %v", err)
				}
			}

			got := DetectVersion("nodejs", dir)
			if got != tt.expected {
				t.Errorf("DetectVersion(\"nodejs\", dir) = %q, want %q (input: %q)",
					got, tt.expected, safeDeref(tt.fileContent))
			}
		})
	}
}

// strPtr is a helper that returns a pointer to a string literal.
func strPtr(s string) *string {
	return &s
}

// safeDeref returns the string value of a pointer, or "<nil>" if the pointer is nil.
func safeDeref(s *string) string {
	if s == nil {
		return "<nil>"
	}
	return *s
}

// TestDetectDotnetVersion verifies that DetectVersion("dotnet", ...) correctly
// extracts the SDK version from global.json (priority 1), *.csproj TargetFramework
// (priority 2), or falls back to "9.0".
func TestDetectDotnetVersion(t *testing.T) {
	const dotnetDefault = "9.0"

	tests := []struct {
		name     string
		files    map[string]string // filename -> content
		expected string
	}{
		{
			name: "global.json with SDK version 9.0.100",
			files: map[string]string{
				"global.json": `{"sdk": {"version": "9.0.100"}}`,
			},
			expected: "9.0",
		},
		{
			name: "global.json with SDK version 8.0.301",
			files: map[string]string{
				"global.json": `{
  "sdk": {
    "version": "8.0.301"
  }
}`,
			},
			expected: "8.0",
		},
		{
			name: "csproj with TargetFramework net8.0",
			files: map[string]string{
				"MyApp.csproj": `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
  </PropertyGroup>
</Project>`,
			},
			expected: "8.0",
		},
		{
			name: "csproj with TargetFramework net7.0",
			files: map[string]string{
				"MyApp.csproj": `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net7.0</TargetFramework>
  </PropertyGroup>
</Project>`,
			},
			expected: "7.0",
		},
		{
			name: "global.json takes priority over csproj",
			files: map[string]string{
				"global.json": `{"sdk": {"version": "9.0.100"}}`,
				"MyApp.csproj": `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>net8.0</TargetFramework>
  </PropertyGroup>
</Project>`,
			},
			expected: "9.0",
		},
		{
			name:     "no files — fallback",
			files:    map[string]string{},
			expected: dotnetDefault,
		},
		{
			name: "global.json without version field",
			files: map[string]string{
				"global.json": `{"sdk": {"rollForward": "latestFeature"}}`,
			},
			expected: dotnetDefault,
		},
		{
			name: "csproj without TargetFramework",
			files: map[string]string{
				"MyApp.csproj": `<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <OutputType>Exe</OutputType>
  </PropertyGroup>
</Project>`,
			},
			expected: dotnetDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write %s: %v", name, err)
				}
			}

			got := DetectVersion("dotnet", dir)
			if got != tt.expected {
				t.Errorf("DetectVersion(\"dotnet\", dir) = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestDetectPhpVersion verifies that DetectVersion("php", ...) correctly
// extracts the PHP version from composer.json (require.php) or .php-version.
func TestDetectPhpVersion(t *testing.T) {
	const phpDefault = "8.4"

	tests := []struct {
		name     string
		files    map[string]string // filename -> content
		expected string
	}{
		{
			name: "composer.json with >=8.2",
			files: map[string]string{
				"composer.json": `{"require":{"php":">=8.2"}}`,
			},
			expected: "8.2",
		},
		{
			name: "composer.json with ^8.3",
			files: map[string]string{
				"composer.json": `{"require":{"php":"^8.3"}}`,
			},
			expected: "8.3",
		},
		{
			name: "composer.json with ~8.1.0",
			files: map[string]string{
				"composer.json": `{"require":{"php":"~8.1.0"}}`,
			},
			expected: "8.1.0",
		},
		{
			name: "composer.json with version range >=8.2 <8.4",
			files: map[string]string{
				"composer.json": `{"require":{"php":">=8.2 <8.4"}}`,
			},
			expected: "8.2",
		},
		{
			name: "composer.json with || constraint",
			files: map[string]string{
				"composer.json": `{"require":{"php":"^7.4 || ^8.0"}}`,
			},
			expected: "7.4",
		},
		{
			name: "composer.json without php require",
			files: map[string]string{
				"composer.json": `{"require":{"laravel/framework":"^11.0"}}`,
			},
			expected: phpDefault,
		},
		{
			name: ".php-version file takes over when no composer.json",
			files: map[string]string{
				".php-version": "8.3\n",
			},
			expected: "8.3",
		},
		{
			name: "composer.json takes priority over .php-version",
			files: map[string]string{
				"composer.json": `{"require":{"php":">=8.2"}}`,
				".php-version":  "8.1\n",
			},
			expected: "8.2",
		},
		{
			name:     "no files — fallback to default",
			files:    map[string]string{},
			expected: phpDefault,
		},
		{
			name: ".php-version with prefixed content",
			files: map[string]string{
				".php-version": "php-8.3.1",
			},
			expected: "8.3.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write %s: %v", name, err)
				}
			}

			got := DetectVersion("php", dir)
			if got != tt.expected {
				t.Errorf("DetectVersion(\"php\", dir) = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestDetectKotlinVersion verifies that DetectVersion("kotlin", ...) correctly
// extracts the JVM target version from build.gradle.kts.
func TestDetectKotlinVersion(t *testing.T) {
	const kotlinDefault = "21"

	tests := []struct {
		name     string
		files    map[string]string // filename -> content
		expected string
	}{
		{
			name: "jvmTarget = 17",
			files: map[string]string{
				"build.gradle.kts": `plugins {
    kotlin("jvm") version "2.1.0"
}

kotlin {
    jvmToolchain(17)
}

tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile> {
    kotlinOptions.jvmTarget = "17"
}`,
			},
			expected: "17",
		},
		{
			name: "jvmTarget = 21",
			files: map[string]string{
				"build.gradle.kts": `plugins {
    kotlin("jvm") version "2.1.0"
}

tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile> {
    kotlinOptions.jvmTarget = "21"
}`,
			},
			expected: "21",
		},
		{
			name: "jvmTarget = 11",
			files: map[string]string{
				"build.gradle.kts": `kotlin {
    jvmToolchain(11)
}
tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile> {
    kotlinOptions.jvmTarget = "11"
}`,
			},
			expected: "11",
		},
		{
			name: "no jvmTarget — fallback to default",
			files: map[string]string{
				"build.gradle.kts": `plugins {
    kotlin("jvm") version "2.1.0"
}`,
			},
			expected: kotlinDefault,
		},
		{
			name:     "no build.gradle.kts — fallback to default",
			files:    map[string]string{},
			expected: kotlinDefault,
		},
		{
			name: "jvmTarget with major.minor format",
			files: map[string]string{
				"build.gradle.kts": `tasks.withType<org.jetbrains.kotlin.gradle.tasks.KotlinCompile> {
    kotlinOptions.jvmTarget = "1.8"
}`,
			},
			expected: "1.8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write %s: %v", name, err)
				}
			}

			got := DetectVersion("kotlin", dir)
			if got != tt.expected {
				t.Errorf("DetectVersion(\"kotlin\", dir) = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestDetectElixirVersion verifies that DetectVersion("elixir", ...) correctly
// extracts the Elixir version from mix.exs or .tool-versions.
func TestDetectElixirVersion(t *testing.T) {
	const elixirDefault = "1.18"

	tests := []struct {
		name     string
		files    map[string]string // filename -> content
		expected string
	}{
		{
			name: "mix.exs with ~> 1.16",
			files: map[string]string{
				"mix.exs": `defmodule MyApp.MixProject do
  use Mix.Project

  def project do
    [
      app: :my_app,
      version: "0.1.0",
      elixir: "~> 1.16",
      deps: deps()
    ]
  end
end`,
			},
			expected: "1.16",
		},
		{
			name: "mix.exs with >= 1.14.0",
			files: map[string]string{
				"mix.exs": `defmodule MyApp.MixProject do
  use Mix.Project

  def project do
    [
      app: :my_app,
      elixir: ">= 1.14.0",
      deps: deps()
    ]
  end
end`,
			},
			expected: "1.14.0",
		},
		{
			name: "mix.exs with exact version 1.17.2",
			files: map[string]string{
				"mix.exs": `defmodule MyApp.MixProject do
  use Mix.Project

  def project do
    [
      app: :my_app,
      elixir: "1.17.2",
      deps: deps()
    ]
  end
end`,
			},
			expected: "1.17.2",
		},
		{
			name: "mix.exs without elixir version",
			files: map[string]string{
				"mix.exs": `defmodule MyApp.MixProject do
  use Mix.Project

  def project do
    [
      app: :my_app,
      version: "0.1.0",
      deps: deps()
    ]
  end
end`,
			},
			expected: elixirDefault,
		},
		{
			name: ".tool-versions with elixir 1.15.7",
			files: map[string]string{
				".tool-versions": "erlang 26.2.1\nelixir 1.15.7-otp-26\n",
			},
			expected: "1.15.7",
		},
		{
			name: "mix.exs takes priority over .tool-versions",
			files: map[string]string{
				"mix.exs": `defmodule MyApp.MixProject do
  use Mix.Project

  def project do
    [
      app: :my_app,
      elixir: "~> 1.16",
      deps: deps()
    ]
  end
end`,
				".tool-versions": "elixir 1.14.0\n",
			},
			expected: "1.16",
		},
		{
			name:     "no files — fallback to default",
			files:    map[string]string{},
			expected: elixirDefault,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write %s: %v", name, err)
				}
			}

			got := DetectVersion("elixir", dir)
			if got != tt.expected {
				t.Errorf("DetectVersion(\"elixir\", dir) = %q, want %q", got, tt.expected)
			}
		})
	}
}

// TestDetectScalaVersion verifies that DetectVersion("scala", ...) correctly
// determines the JDK version based on the Scala version in build.sbt.
func TestDetectScalaVersion(t *testing.T) {
	const scalaDefault = "21" // JDK version

	tests := []struct {
		name     string
		files    map[string]string
		expected string
	}{
		{
			name: "Scala 3.6.4 — JDK 21",
			files: map[string]string{
				"build.sbt": `name := "myapp"
scalaVersion := "3.6.4"

libraryDependencies += "org.scalatest" %% "scalatest" % "3.2.18" % Test`,
			},
			expected: "21",
		},
		{
			name: "Scala 3.3.0 — JDK 21",
			files: map[string]string{
				"build.sbt": `scalaVersion := "3.3.0"`,
			},
			expected: "21",
		},
		{
			name: "Scala 2.13.12 — JDK 17",
			files: map[string]string{
				"build.sbt": `scalaVersion := "2.13.12"
name := "legacy-app"`,
			},
			expected: "17",
		},
		{
			name: "Scala 2.12.19 — JDK 17",
			files: map[string]string{
				"build.sbt": `scalaVersion := "2.12.19"`,
			},
			expected: "17",
		},
		{
			name: "no scalaVersion in build.sbt — fallback to default",
			files: map[string]string{
				"build.sbt": `name := "myapp"
libraryDependencies += "org.typelevel" %% "cats-core" % "2.10.0"`,
			},
			expected: scalaDefault,
		},
		{
			name:     "no build.sbt — fallback to default",
			files:    map[string]string{},
			expected: scalaDefault,
		},
		{
			name: "scalaVersion with spaces around :=",
			files: map[string]string{
				"build.sbt": `scalaVersion  :=  "3.5.0"`,
			},
			expected: "21",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()

			for name, content := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644); err != nil {
					t.Fatalf("failed to write %s: %v", name, err)
				}
			}

			got := DetectVersion("scala", dir)
			if got != tt.expected {
				t.Errorf("DetectVersion(\"scala\", dir) = %q, want %q", got, tt.expected)
			}
		})
	}
}
