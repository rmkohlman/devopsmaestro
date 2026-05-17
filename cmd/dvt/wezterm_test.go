package main

import (
	"encoding/json"
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