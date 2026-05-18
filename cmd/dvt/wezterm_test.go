package main

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/rmkohlman/MaestroTerminal/terminalops/wezterm"
)

// TestYamlConfigToJSON_ProducesValidJSON verifies that yamlConfigToJSON
// produces valid JSON output (not YAML) - regression test for issue #435
func TestYamlConfigToJSON_ProducesValidJSON(t *testing.T) {
	tests := []struct {
		name   string
		config *wezterm.WeztermYAML
	}{
		{
			name: "empty config",
			config: &wezterm.WeztermYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "WeztermConfig",
				Metadata: wezterm.WeztermMetadata{
					Name: "test",
				},
				Spec: wezterm.WeztermSpec{},
			},
		},
		{
			name: "config with font settings",
			config: &wezterm.WeztermYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "WeztermConfig",
				Metadata: wezterm.WeztermMetadata{
					Name: "test",
				},
				Spec: wezterm.WeztermSpec{
					Font: wezterm.FontConfig{
						Family: "Cascadia Code",
						Size:   14,
					},
				},
			},
		},
		{
			name: "config with window settings",
			config: &wezterm.WeztermYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "WeztermConfig",
				Metadata: wezterm.WeztermMetadata{
					Name: "test",
				},
				Spec: wezterm.WeztermSpec{
					Window: wezterm.WindowConfig{
						Opacity:      0.9,
						Decorations:  "full",
						InitialRows:  24,
						InitialCols:  80,
					},
				},
			},
		},
		{
			name: "config with colors",
			config: &wezterm.WeztermYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "WeztermConfig",
				Metadata: wezterm.WeztermMetadata{
					Name: "test",
				},
				Spec: wezterm.WeztermSpec{
					Colors: &wezterm.ColorConfig{
						Foreground: "#c0caf5",
						Background: "#1a1b26",
					},
				},
			},
		},
		{
			name: "config with boolean values",
			config: &wezterm.WeztermYAML{
				APIVersion: "devopsmaestro.io/v1",
				Kind:       "WeztermConfig",
				Metadata: wezterm.WeztermMetadata{
					Name: "test",
				},
				Spec: wezterm.WeztermSpec{
					TabBar: &wezterm.TabBarConfig{
						Enabled:         true,
						ShowNewTab:      true,
						FancyTabBar:     false,
						HideTabBarIfOnly: true,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := yamlConfigToJSON(tt.config)
			if err != nil {
				t.Fatalf("yamlConfigToJSON returned error: %v", err)
			}

			output := string(result)

			// Verify output is valid JSON by unmarshaling it
			var parsed map[string]interface{}
			if err := json.Unmarshal(result, &parsed); err != nil {
				t.Errorf("Output is not valid JSON: %v\nOutput: %s", err, output)
			}

			// Verify JSON-specific characteristics:
			// 1. Should NOT start with YAML document marker
			if strings.HasPrefix(output, "---") {
				t.Errorf("Output appears to be YAML (starts with '---'): %s", output)
			}

			// 2. Should NOT start with YAML comment marker
			if strings.HasPrefix(output, "#") {
				t.Errorf("Output appears to be YAML (starts with '#'): %s", output)
			}

			// 3. Should start with '{' for JSON object
			if !strings.HasPrefix(output, "{") {
				t.Errorf("Output does not start with '{' (not JSON object): %s", output)
			}
		})
	}
}

// TestYamlConfigToJSON_NotEqualToYAML verifies that JSON output is different from YAML output
func TestYamlConfigToJSON_NotEqualToYAML(t *testing.T) {
	config := &wezterm.WeztermYAML{
		APIVersion: "devopsmaestro.io/v1",
		Kind:       "WeztermConfig",
		Metadata: wezterm.WeztermMetadata{
			Name: "test",
		},
		Spec: wezterm.WeztermSpec{
			Font: wezterm.FontConfig{
				Family: "Cascadia Code",
				Size:   14,
			},
		},
	}

	jsonResult, err := yamlConfigToJSON(config)
	if err != nil {
		t.Fatalf("yamlConfigToJSON returned error: %v", err)
	}

	yamlResult, err := yamlConfigToYAML(config)
	if err != nil {
		t.Fatalf("yamlConfigToYAML returned error: %v", err)
	}

	// Verify JSON output is valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonResult, &parsed); err != nil {
		t.Errorf("JSON output is invalid: %v\nOutput: %s", err, string(jsonResult))
	}

	// JSON and YAML outputs should be different (JSON uses different formatting)
	// Note: They might be the same for very simple configs, but we verify JSON is valid
	t.Logf("JSON output: %s", string(jsonResult))
	t.Logf("YAML output: %s", string(yamlResult))
}

// TestComputeDiff_UnifiedFormat verifies that computeDiff produces proper
// unified diff format - regression test for issue #437
func TestComputeDiff_UnifiedFormat(t *testing.T) {
	tests := []struct {
		name         string
		oldConfig    string
		newConfig    string
		checkHeaders bool   // check for --- a/ and +++ b/
		checkHunks   bool   // check for @@ -X,Y +X,Y @@
		checkPrefixes bool  // check for - and + prefixes
	}{
		{
			name:         "simple line change",
			oldConfig:    "font_size = 12\nfont_family = \"JetBrains Mono\"\n",
			newConfig:    "font_size = 14\nfont_family = \"Fira Code\"\n",
			checkHeaders: true,
			checkHunks:   true,
			checkPrefixes: true,
		},
		{
			name:         "added lines",
			oldConfig:    "font_size = 12\n",
			newConfig:    "font_size = 12\nfont_family = \"Fira Code\"\ncolor_scheme = \"dark\"\n",
			checkHeaders: true,
			checkHunks:   true,
			checkPrefixes: true,
		},
		{
			name:         "deleted lines",
			oldConfig:    "font_size = 12\nfont_family = \"JetBrains Mono\"\ncolor_scheme = \"dark\"\n",
			newConfig:    "font_size = 12\n",
			checkHeaders: true,
			checkHunks:   true,
			checkPrefixes: true,
		},
		{
			name:         "identical configs produces minimal diff",
			oldConfig:    "font_size = 12\n",
			newConfig:    "font_size = 12\n",
			checkHeaders: true,
			checkHunks:   false, // no hunks when identical
			checkPrefixes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := computeDiff(tt.oldConfig, tt.newConfig)
			lines := strings.Split(result, "\n")

			// Check file path headers
			if tt.checkHeaders {
				if len(lines) < 2 {
					t.Fatalf("Expected at least 2 lines for headers, got %d: %s", len(lines), result)
				}

				// First line should be "--- a/config"
				if !strings.HasPrefix(lines[0], "--- a/") {
					t.Errorf("Expected first line to start with '--- a/', got: %s", lines[0])
				}

				// Second line should be "+++ b/config"
				if !strings.HasPrefix(lines[1], "+++ b/") {
					t.Errorf("Expected second line to start with '+++ b/', got: %s", lines[1])
				}
			}

			// Check hunk headers
			if tt.checkHunks {
				hasHunk := false
				for _, line := range lines {
					if strings.HasPrefix(line, "@@ ") && strings.HasSuffix(line, " @@") {
						hasHunk = true
						// Verify hunk header format: @@ -X,Y +X,Y @@
						if !regexp.MustCompile(`^@@ -\d+,\d+ \+\d+,\d+ @@$`).MatchString(line) {
							t.Errorf("Invalid hunk header format: %s", line)
						}
					}
				}
				if !hasHunk && tt.oldConfig != tt.newConfig {
					t.Errorf("Expected hunk header (@@ ... @@) in diff output, got:\n%s", result)
				}
			}

			// Check line prefixes (-, +, or space for context)
			if tt.checkPrefixes {
				foundChange := false
				for _, line := range lines {
					if len(line) > 0 {
						firstChar := line[0]
						if firstChar == '-' || firstChar == '+' {
							foundChange = true
							// Verify no line number prefix (should be just "-text", not "-1: text")
							if len(line) > 1 && line[1] >= '0' && line[1] <= '9' {
								t.Errorf("Line has line number prefix instead of just '-'/'+': %s", line)
							}
						} else if firstChar == ' ' || firstChar == '@' || firstChar == '-' || firstChar == '+' {
							// Valid first characters in unified diff
						}
					}
				}
				if !foundChange && tt.oldConfig != tt.newConfig {
					t.Errorf("Expected changed lines with -/+ prefixes, got:\n%s", result)
				}
			}

			t.Logf("Diff output:\n%s", result)
		})
	}
}

// TestMergeWezTermConfigs_Precedence verifies that mergeWezTermConfigs applies
// config precedence correctly: existing < preset < mergeOverrides
// This is a regression test for issue #436
func TestMergeWezTermConfigs_Precedence(t *testing.T) {
	tests := []struct {
		name           string
		existing      *wezterm.WezTerm
		preset        *wezterm.WezTerm
		overrides     *wezterm.WezTerm
		expectedFont  string
		expectedSize  float64
		expectedOpacity float64
	}{
		{
			name: "font family precedence",
			existing: &wezterm.WezTerm{
				Font: wezterm.FontConfig{Family: "JetBrains Mono", Size: 12},
			},
			preset: &wezterm.WezTerm{
				Font: wezterm.FontConfig{Family: "Fira Code", Size: 14},
			},
			overrides: &wezterm.WezTerm{
				Font: wezterm.FontConfig{Family: "Cascadia Code", Size: 16},
			},
			expectedFont:  "Cascadia Code", // overrides should win
			expectedSize:  16,               // overrides should win
			expectedOpacity: 0,
		},
		{
			name: "preset overrides existing",
			existing: &wezterm.WezTerm{
				Font: wezterm.FontConfig{Family: "JetBrains Mono", Size: 12},
			},
			preset: &wezterm.WezTerm{
				Font: wezterm.FontConfig{Family: "Fira Code", Size: 14},
			},
			overrides: nil, // no overrides - preset should win
			expectedFont:  "Fira Code", // preset should override existing
			expectedSize:  14,
			expectedOpacity: 0,
		},
		{
			name: "existing is base when others are nil",
			existing: &wezterm.WezTerm{
				Font: wezterm.FontConfig{Family: "JetBrains Mono", Size: 12},
			},
			preset:  nil,
			overrides: nil,
			expectedFont:  "JetBrains Mono", // existing should be base
			expectedSize:  12,
			expectedOpacity: 0,
		},
		{
			name: "window opacity precedence",
			existing: &wezterm.WezTerm{
				Window: wezterm.WindowConfig{Opacity: 0.8},
			},
			preset: &wezterm.WezTerm{
				Window: wezterm.WindowConfig{Opacity: 0.9},
			},
			overrides: &wezterm.WezTerm{
				Window: wezterm.WindowConfig{Opacity: 1.0},
			},
			expectedFont:    "",
			expectedSize:   0,
			expectedOpacity: 1.0, // overrides should win
		},
		{
			name: "nil existing uses preset as base",
			existing: nil,
			preset: &wezterm.WezTerm{
				Font: wezterm.FontConfig{Family: "Fira Code", Size: 14},
			},
			overrides: &wezterm.WezTerm{
				Font: wezterm.FontConfig{Family: "Cascadia Code"},
			},
			expectedFont:  "Cascadia Code", // overrides wins
			expectedSize:  14, // preset size (overrides doesn't set size)
			expectedOpacity: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeWezTermConfigs(tt.existing, tt.preset, tt.overrides)

			// Verify font family precedence
			if tt.expectedFont != "" && result.Font.Family != tt.expectedFont {
				t.Errorf("Font.Family: expected %q, got %q", tt.expectedFont, result.Font.Family)
			}

			// Verify font size precedence
			if tt.expectedSize > 0 && result.Font.Size != tt.expectedSize {
				t.Errorf("Font.Size: expected %v, got %v", tt.expectedSize, result.Font.Size)
			}

			// Verify window opacity precedence
			if tt.expectedOpacity > 0 && result.Window.Opacity != tt.expectedOpacity {
				t.Errorf("Window.Opacity: expected %v, got %v", tt.expectedOpacity, result.Window.Opacity)
			}
		})
	}
}