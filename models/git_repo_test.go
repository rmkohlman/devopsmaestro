package models

import (
	"strings"
	"testing"
)

// =============================================================================
// GitRepoDB ToYAML Tests
// =============================================================================

func TestGitRepoDB_ToYAML(t *testing.T) {
	repo := &GitRepoDB{
		ID:                  42,
		Name:                "my-repo",
		URL:                 "https://github.com/user/my-repo",
		DefaultRef:          "main",
		AuthType:            "none",
		AutoSync:            true,
		SyncIntervalMinutes: 30,
	}

	yamlDoc := repo.ToYAML()

	if yamlDoc.APIVersion != "devopsmaestro.io/v1" {
		t.Errorf("ToYAML() APIVersion = %q, want %q", yamlDoc.APIVersion, "devopsmaestro.io/v1")
	}
	if yamlDoc.Kind != "GitRepo" {
		t.Errorf("ToYAML() Kind = %q, want %q", yamlDoc.Kind, "GitRepo")
	}
	if yamlDoc.Metadata.Name != "my-repo" {
		t.Errorf("ToYAML() Metadata.Name = %q, want %q", yamlDoc.Metadata.Name, "my-repo")
	}
	if yamlDoc.Spec.URL != "https://github.com/user/my-repo" {
		t.Errorf("ToYAML() Spec.URL = %q, want %q", yamlDoc.Spec.URL, "https://github.com/user/my-repo")
	}
	if yamlDoc.Spec.DefaultRef != "main" {
		t.Errorf("ToYAML() Spec.DefaultRef = %q, want %q", yamlDoc.Spec.DefaultRef, "main")
	}
	if yamlDoc.Spec.AuthType != "none" {
		t.Errorf("ToYAML() Spec.AuthType = %q, want %q", yamlDoc.Spec.AuthType, "none")
	}
	if !yamlDoc.Spec.AutoSync {
		t.Errorf("ToYAML() Spec.AutoSync = false, want true")
	}
	if yamlDoc.Spec.SyncIntervalMinutes != 30 {
		t.Errorf("ToYAML() Spec.SyncIntervalMinutes = %d, want 30", yamlDoc.Spec.SyncIntervalMinutes)
	}
}

func TestGitRepoDB_ToYAML_MinimalFields(t *testing.T) {
	repo := &GitRepoDB{
		Name: "minimal-repo",
		URL:  "https://github.com/user/minimal",
	}

	yamlDoc := repo.ToYAML()

	if yamlDoc.APIVersion != "devopsmaestro.io/v1" {
		t.Errorf("ToYAML() APIVersion = %q, want %q", yamlDoc.APIVersion, "devopsmaestro.io/v1")
	}
	if yamlDoc.Kind != "GitRepo" {
		t.Errorf("ToYAML() Kind = %q, want %q", yamlDoc.Kind, "GitRepo")
	}
	if yamlDoc.Metadata.Name != "minimal-repo" {
		t.Errorf("ToYAML() Metadata.Name = %q, want %q", yamlDoc.Metadata.Name, "minimal-repo")
	}
	if yamlDoc.Spec.URL != "https://github.com/user/minimal" {
		t.Errorf("ToYAML() Spec.URL = %q, want %q", yamlDoc.Spec.URL, "https://github.com/user/minimal")
	}
	// DefaultRef should be empty or zero for minimal case
	if yamlDoc.Spec.AutoSync {
		t.Errorf("ToYAML() Spec.AutoSync = true, want false for zero-value")
	}
}

// =============================================================================
// GitRepoDB FromYAML Tests
// =============================================================================

func TestGitRepoDB_FromYAML(t *testing.T) {
	yamlDoc := GitRepoYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "GitRepo",
		Metadata: GitRepoMetadata{
			Name: "from-yaml-repo",
			Annotations: map[string]string{
				"description": "Test repository from YAML",
			},
		},
		Spec: GitRepoSpec{
			URL:                 "https://github.com/user/from-yaml-repo",
			DefaultRef:          "develop",
			AuthType:            "ssh",
			AutoSync:            true,
			SyncIntervalMinutes: 60,
		},
	}

	repo := &GitRepoDB{}
	repo.FromYAML(yamlDoc)

	if repo.Name != "from-yaml-repo" {
		t.Errorf("FromYAML() Name = %q, want %q", repo.Name, "from-yaml-repo")
	}
	if repo.URL != "https://github.com/user/from-yaml-repo" {
		t.Errorf("FromYAML() URL = %q, want %q", repo.URL, "https://github.com/user/from-yaml-repo")
	}
	if repo.DefaultRef != "develop" {
		t.Errorf("FromYAML() DefaultRef = %q, want %q", repo.DefaultRef, "develop")
	}
	if repo.AuthType != "ssh" {
		t.Errorf("FromYAML() AuthType = %q, want %q", repo.AuthType, "ssh")
	}
	if !repo.AutoSync {
		t.Errorf("FromYAML() AutoSync = false, want true")
	}
	if repo.SyncIntervalMinutes != 60 {
		t.Errorf("FromYAML() SyncIntervalMinutes = %d, want 60", repo.SyncIntervalMinutes)
	}
}

// =============================================================================
// GitRepoDB Round-Trip Tests
// =============================================================================

func TestGitRepoDB_RoundTrip(t *testing.T) {
	original := &GitRepoDB{
		Name:                "roundtrip-repo",
		URL:                 "https://github.com/user/roundtrip",
		DefaultRef:          "main",
		AuthType:            "none",
		AutoSync:            true,
		SyncIntervalMinutes: 15,
	}

	// ToYAML then FromYAML
	yamlDoc := original.ToYAML()
	restored := &GitRepoDB{}
	restored.FromYAML(yamlDoc)

	if restored.Name != original.Name {
		t.Errorf("RoundTrip Name = %q, want %q", restored.Name, original.Name)
	}
	if restored.URL != original.URL {
		t.Errorf("RoundTrip URL = %q, want %q", restored.URL, original.URL)
	}
	if restored.DefaultRef != original.DefaultRef {
		t.Errorf("RoundTrip DefaultRef = %q, want %q", restored.DefaultRef, original.DefaultRef)
	}
	if restored.AuthType != original.AuthType {
		t.Errorf("RoundTrip AuthType = %q, want %q", restored.AuthType, original.AuthType)
	}
	if restored.AutoSync != original.AutoSync {
		t.Errorf("RoundTrip AutoSync = %v, want %v", restored.AutoSync, original.AutoSync)
	}
	if restored.SyncIntervalMinutes != original.SyncIntervalMinutes {
		t.Errorf("RoundTrip SyncIntervalMinutes = %d, want %d", restored.SyncIntervalMinutes, original.SyncIntervalMinutes)
	}
}

// =============================================================================
// GitRepoYAML Struct Field Tests
// =============================================================================

func TestGitRepoYAML_FieldAccess(t *testing.T) {
	// This test verifies all struct fields exist and are accessible.
	// It will fail to compile if any required field is missing from the struct definition.
	doc := GitRepoYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       KindGitRepoModel,
		Metadata: GitRepoMetadata{
			Name:        "test",
			Annotations: map[string]string{"key": "value"},
		},
		Spec: GitRepoSpec{
			URL:                 "https://github.com/user/repo",
			DefaultRef:          "main",
			AuthType:            "none",
			AutoSync:            false,
			SyncIntervalMinutes: 0,
		},
	}

	// Verify values are accessible
	if !strings.Contains(doc.APIVersion, "devopsmaestro") {
		t.Errorf("APIVersion should contain 'devopsmaestro', got %q", doc.APIVersion)
	}
	if doc.Kind != KindGitRepoModel {
		t.Errorf("Kind = %q, want %q", doc.Kind, KindGitRepoModel)
	}
}
