package themebridge

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"

	"devopsmaestro/models"
)

// mockInheritanceStore implements ThemeDataStore for inheritance testing.
type mockInheritanceStore struct {
	themes map[string]*models.NvimThemeDB
}

func newMockInheritanceStore() *mockInheritanceStore {
	return &mockInheritanceStore{themes: make(map[string]*models.NvimThemeDB)}
}

func (m *mockInheritanceStore) CreateTheme(t *models.NvimThemeDB) error {
	m.themes[t.Name] = t
	return nil
}

func (m *mockInheritanceStore) GetThemeByName(name string) (*models.NvimThemeDB, error) {
	t, ok := m.themes[name]
	if !ok {
		return nil, fmt.Errorf("theme %q not found", name)
	}
	return t, nil
}

func (m *mockInheritanceStore) UpdateTheme(t *models.NvimThemeDB) error    { return nil }
func (m *mockInheritanceStore) DeleteTheme(name string) error              { return nil }
func (m *mockInheritanceStore) ListThemes() ([]*models.NvimThemeDB, error) { return nil, nil }
func (m *mockInheritanceStore) ListThemesByCategory(string) ([]*models.NvimThemeDB, error) {
	return nil, nil
}
func (m *mockInheritanceStore) GetActiveTheme() (*models.NvimThemeDB, error) { return nil, nil }
func (m *mockInheritanceStore) SetActiveTheme(string) error                  { return nil }
func (m *mockInheritanceStore) ClearActiveTheme() error                      { return nil }

func nullStr(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

func TestResolveWithInheritance(t *testing.T) {
	tests := []struct {
		name      string
		themes    map[string]*models.NvimThemeDB
		resolve   string
		wantErr   string
		checkFunc func(t *testing.T, result *models.NvimThemeDB)
	}{
		{
			name: "no inheritance passthrough",
			themes: map[string]*models.NvimThemeDB{
				"base": {
					Name:       "base",
					PluginRepo: "catppuccin/nvim",
					Style:      nullStr("mocha"),
					Colors:     nullStr(`{"bg":"#1e1e2e","fg":"#cdd6f4"}`),
				},
			},
			resolve: "base",
			checkFunc: func(t *testing.T, result *models.NvimThemeDB) {
				if result.Name != "base" {
					t.Errorf("expected name 'base', got %q", result.Name)
				}
				if result.PluginRepo != "catppuccin/nvim" {
					t.Errorf("expected plugin 'catppuccin/nvim', got %q", result.PluginRepo)
				}
				if !result.Style.Valid || result.Style.String != "mocha" {
					t.Errorf("expected style 'mocha', got %v", result.Style)
				}
			},
		},
		{
			name: "single-level inheritance",
			themes: map[string]*models.NvimThemeDB{
				"base": {
					Name:       "base",
					PluginRepo: "catppuccin/nvim",
					Style:      nullStr("mocha"),
					Author:     nullStr("catppuccin"),
					Colors:     nullStr(`{"bg":"#1e1e2e","fg":"#cdd6f4","accent":"#89b4fa"}`),
				},
				"child": {
					Name:     "child",
					Inherits: nullStr("base"),
					Colors:   nullStr(`{"bg":"#000000"}`),
					Style:    nullStr("latte"),
				},
			},
			resolve: "child",
			checkFunc: func(t *testing.T, result *models.NvimThemeDB) {
				if result.Name != "child" {
					t.Errorf("expected name 'child', got %q", result.Name)
				}
				// Plugin inherited from parent
				if result.PluginRepo != "catppuccin/nvim" {
					t.Errorf("expected inherited plugin, got %q", result.PluginRepo)
				}
				// Style overridden by child
				if !result.Style.Valid || result.Style.String != "latte" {
					t.Errorf("expected style 'latte', got %v", result.Style)
				}
				// Author inherited from parent
				if !result.Author.Valid || result.Author.String != "catppuccin" {
					t.Errorf("expected inherited author 'catppuccin', got %v", result.Author)
				}
				// Colors: child bg overrides, parent fg and accent inherited
				if !result.Colors.Valid {
					t.Fatal("expected merged colors to be valid")
				}
			},
		},
		{
			name: "multi-level inheritance A->B->C",
			themes: map[string]*models.NvimThemeDB{
				"grandparent": {
					Name:       "grandparent",
					PluginRepo: "folke/tokyonight.nvim",
					Author:     nullStr("folke"),
					Colors:     nullStr(`{"bg":"#111111","fg":"#ffffff"}`),
				},
				"parent": {
					Name:     "parent",
					Inherits: nullStr("grandparent"),
					Style:    nullStr("night"),
					Colors:   nullStr(`{"bg":"#222222"}`),
				},
				"child": {
					Name:     "child",
					Inherits: nullStr("parent"),
					Colors:   nullStr(`{"accent":"#ff0000"}`),
				},
			},
			resolve: "child",
			checkFunc: func(t *testing.T, result *models.NvimThemeDB) {
				if result.PluginRepo != "folke/tokyonight.nvim" {
					t.Errorf("expected inherited plugin from grandparent, got %q", result.PluginRepo)
				}
				if !result.Author.Valid || result.Author.String != "folke" {
					t.Errorf("expected inherited author from grandparent")
				}
				if !result.Style.Valid || result.Style.String != "night" {
					t.Errorf("expected inherited style from parent")
				}
			},
		},
		{
			name: "circular inheritance detected",
			themes: map[string]*models.NvimThemeDB{
				"a": {
					Name:       "a",
					PluginRepo: "test/plugin",
					Inherits:   nullStr("b"),
				},
				"b": {
					Name:       "b",
					PluginRepo: "test/plugin",
					Inherits:   nullStr("a"),
				},
			},
			resolve: "a",
			wantErr: "circular theme inheritance detected",
		},
		{
			name: "max depth exceeded",
			themes: func() map[string]*models.NvimThemeDB {
				m := make(map[string]*models.NvimThemeDB)
				for i := 0; i <= MaxInheritanceDepth+1; i++ {
					name := fmt.Sprintf("theme-%d", i)
					parent := ""
					if i > 0 {
						parent = fmt.Sprintf("theme-%d", i-1)
					}
					m[name] = &models.NvimThemeDB{
						Name:       name,
						PluginRepo: "test/plugin",
						Inherits:   nullStr(parent),
					}
				}
				return m
			}(),
			resolve: fmt.Sprintf("theme-%d", MaxInheritanceDepth+1),
			wantErr: "depth exceeded maximum",
		},
		{
			name: "parent not found",
			themes: map[string]*models.NvimThemeDB{
				"orphan": {
					Name:       "orphan",
					PluginRepo: "test/plugin",
					Inherits:   nullStr("nonexistent"),
				},
			},
			resolve: "orphan",
			wantErr: "failed to resolve parent theme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := newMockInheritanceStore()
			for k, v := range tt.themes {
				store.themes[k] = v
			}

			result, err := ResolveWithInheritance(tt.resolve, store)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !containsSubstring(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && searchSubstring(s, substr)
}

func searchSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMergeJSONMaps_DeepMergeColors(t *testing.T) {
	parent := nullStr(`{"bg":"#1e1e2e","fg":"#cdd6f4","accent":"#89b4fa"}`)
	child := nullStr(`{"bg":"#000000","warn":"#f9e2af"}`)

	merged, err := mergeJSONMaps(child, parent)
	if err != nil {
		t.Fatalf("mergeJSONMaps error: %v", err)
	}

	if !merged.Valid {
		t.Fatal("expected merged to be valid")
	}

	// Verify merged JSON contains all keys
	var result map[string]string
	if err := jsonUnmarshal(merged.String, &result); err != nil {
		t.Fatalf("failed to unmarshal merged: %v", err)
	}

	// Child override
	if result["bg"] != "#000000" {
		t.Errorf("expected child bg '#000000', got %q", result["bg"])
	}
	// Parent inherited
	if result["fg"] != "#cdd6f4" {
		t.Errorf("expected parent fg '#cdd6f4', got %q", result["fg"])
	}
	if result["accent"] != "#89b4fa" {
		t.Errorf("expected parent accent '#89b4fa', got %q", result["accent"])
	}
	// Child addition
	if result["warn"] != "#f9e2af" {
		t.Errorf("expected child warn '#f9e2af', got %q", result["warn"])
	}
}

func TestMergeJSONMaps_DeepMergeHighlights(t *testing.T) {
	parent := nullStr(`{"Normal":{"fg":"#cdd6f4","bg":"#1e1e2e"},"Comment":{"fg":"#6c7086","italic":true}}`)
	child := nullStr(`{"Normal":{"bg":"#000000"},"Error":{"fg":"#f38ba8","bold":true}}`)

	merged, err := mergeJSONMaps(child, parent)
	if err != nil {
		t.Fatalf("mergeJSONMaps error: %v", err)
	}

	if !merged.Valid {
		t.Fatal("expected merged to be valid")
	}

	var result map[string]interface{}
	if err := jsonUnmarshal(merged.String, &result); err != nil {
		t.Fatalf("failed to unmarshal merged: %v", err)
	}

	// Child overrides entire Normal group
	if _, ok := result["Normal"]; !ok {
		t.Error("expected Normal group in merged result")
	}
	// Parent Comment inherited
	if _, ok := result["Comment"]; !ok {
		t.Error("expected Comment group inherited from parent")
	}
	// Child Error added
	if _, ok := result["Error"]; !ok {
		t.Error("expected Error group from child")
	}
}

func TestMergeJSONMaps_OnlyParent(t *testing.T) {
	parent := nullStr(`{"bg":"#1e1e2e"}`)
	child := sql.NullString{} // invalid

	merged, err := mergeJSONMaps(child, parent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged.String != parent.String {
		t.Errorf("expected parent value %q, got %q", parent.String, merged.String)
	}
}

func TestMergeJSONMaps_OnlyChild(t *testing.T) {
	parent := sql.NullString{} // invalid
	child := nullStr(`{"bg":"#000000"}`)

	merged, err := mergeJSONMaps(child, parent)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged.String != child.String {
		t.Errorf("expected child value %q, got %q", child.String, merged.String)
	}
}

func TestMergeJSONMaps_BothEmpty(t *testing.T) {
	merged, err := mergeJSONMaps(sql.NullString{}, sql.NullString{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if merged.Valid {
		t.Error("expected invalid NullString when both are empty")
	}
}

// jsonUnmarshal is a helper to unmarshal JSON for test assertions.
func jsonUnmarshal(s string, v interface{}) error {
	return json.Unmarshal([]byte(s), v)
}
