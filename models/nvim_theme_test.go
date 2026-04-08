package models

import (
	"database/sql"
	"testing"
)

func TestNvimThemeDB_ToYAML_WithCustomHighlights(t *testing.T) {
	db := &NvimThemeDB{
		Name:       "custom-theme",
		PluginRepo: "folke/tokyonight.nvim",
		CustomHighlights: sql.NullString{
			String: `{"MyGroup":{"fg":"#ff0000","bg":"#000000","bold":true},"LinkedGroup":{"link":"Comment"}}`,
			Valid:  true,
		},
	}

	yamlDoc, err := db.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}

	if len(yamlDoc.Spec.CustomHighlights) != 2 {
		t.Fatalf("expected 2 custom highlights, got %d", len(yamlDoc.Spec.CustomHighlights))
	}

	myGroup, ok := yamlDoc.Spec.CustomHighlights["MyGroup"]
	if !ok {
		t.Fatal("expected MyGroup in custom highlights")
	}
	if myGroup.Fg != "#ff0000" {
		t.Errorf("expected fg '#ff0000', got %q", myGroup.Fg)
	}
	if myGroup.Bg != "#000000" {
		t.Errorf("expected bg '#000000', got %q", myGroup.Bg)
	}
	if !myGroup.Bold {
		t.Error("expected bold = true")
	}

	linked, ok := yamlDoc.Spec.CustomHighlights["LinkedGroup"]
	if !ok {
		t.Fatal("expected LinkedGroup in custom highlights")
	}
	if linked.Link != "Comment" {
		t.Errorf("expected link 'Comment', got %q", linked.Link)
	}
}

func TestNvimThemeDB_FromYAML_WithCustomHighlights(t *testing.T) {
	yamlDoc := NvimThemeYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimTheme",
		Metadata: ThemeMetadata{
			Name: "test-theme",
		},
		Spec: ThemeSpec{
			Plugin: ThemePluginSpec{
				Repo: "catppuccin/nvim",
			},
			CustomHighlights: map[string]ThemeHighlight{
				"ErrorHL": {
					Fg:        "#ff0000",
					Bold:      true,
					Undercurl: true,
					Sp:        "#ff0000",
				},
				"LinkHL": {
					Link: "Normal",
				},
			},
		},
	}

	var db NvimThemeDB
	err := db.FromYAML(yamlDoc)
	if err != nil {
		t.Fatalf("FromYAML failed: %v", err)
	}

	if !db.CustomHighlights.Valid {
		t.Fatal("expected CustomHighlights to be valid")
	}

	// Round-trip back to YAML
	yamlDoc2, err := db.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML round-trip failed: %v", err)
	}

	if len(yamlDoc2.Spec.CustomHighlights) != 2 {
		t.Fatalf("expected 2 custom highlights after round-trip, got %d",
			len(yamlDoc2.Spec.CustomHighlights))
	}

	errorHL, ok := yamlDoc2.Spec.CustomHighlights["ErrorHL"]
	if !ok {
		t.Fatal("expected ErrorHL after round-trip")
	}
	if errorHL.Fg != "#ff0000" || !errorHL.Bold || !errorHL.Undercurl || errorHL.Sp != "#ff0000" {
		t.Errorf("ErrorHL attrs mismatch after round-trip: %+v", errorHL)
	}

	linkHL, ok := yamlDoc2.Spec.CustomHighlights["LinkHL"]
	if !ok {
		t.Fatal("expected LinkHL after round-trip")
	}
	if linkHL.Link != "Normal" {
		t.Errorf("LinkHL link mismatch: got %q", linkHL.Link)
	}
}

func TestNvimThemeDB_NoCustomHighlights(t *testing.T) {
	db := &NvimThemeDB{
		Name:       "plain-theme",
		PluginRepo: "folke/tokyonight.nvim",
	}

	yamlDoc, err := db.ToYAML()
	if err != nil {
		t.Fatalf("ToYAML failed: %v", err)
	}

	if yamlDoc.Spec.CustomHighlights != nil {
		t.Errorf("expected nil custom highlights, got %v", yamlDoc.Spec.CustomHighlights)
	}
}

func TestNvimThemeDB_FromYAML_EmptyHighlights(t *testing.T) {
	yamlDoc := NvimThemeYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "NvimTheme",
		Metadata: ThemeMetadata{
			Name: "minimal-theme",
		},
		Spec: ThemeSpec{
			Plugin: ThemePluginSpec{
				Repo: "folke/tokyonight.nvim",
			},
		},
	}

	var db NvimThemeDB
	err := db.FromYAML(yamlDoc)
	if err != nil {
		t.Fatalf("FromYAML failed: %v", err)
	}

	if db.CustomHighlights.Valid {
		t.Error("expected CustomHighlights to be invalid when empty")
	}
}
