package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// Test that all formatters implement the Formatter interface
func TestFormatterInterfaceCompliance(t *testing.T) {
	formatters := []struct {
		name string
		fn   func() Formatter
	}{
		{"Plain", Plain},
		{"Colored", Colored},
		{"JSON", JSON},
		{"YAML", YAML},
		{"Verbose", Verbose},
	}

	for _, tc := range formatters {
		t.Run(tc.name, func(t *testing.T) {
			var _ Formatter = tc.fn()
		})
	}
}

func TestNewFormatter_DefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Style != StyleColored {
		t.Errorf("DefaultConfig().Style = %v, want %v", cfg.Style, StyleColored)
	}
	if cfg.Theme != ThemeAuto {
		t.Errorf("DefaultConfig().Theme = %v, want %v", cfg.Theme, ThemeAuto)
	}
	if cfg.Verbose {
		t.Error("DefaultConfig().Verbose should be false")
	}
}

func TestNewFormatter_StyleSelection(t *testing.T) {
	tests := []struct {
		style    Style
		wantType string
	}{
		{StylePlain, "*output.PlainFormatter"},
		{StyleColored, "*output.ColoredFormatter"},
		{StyleJSON, "*output.JSONFormatter"},
		{StyleYAML, "*output.YAMLFormatter"},
		{StyleCompact, "*output.CompactFormatter"},
		{StyleVerbose, "*output.VerboseFormatter"},
		{StyleTable, "*output.TableOnlyFormatter"},
	}

	for _, tc := range tests {
		t.Run(string(tc.style), func(t *testing.T) {
			f := NewFormatter(FormatterConfig{Style: tc.style})
			if f.GetStyle() != tc.style {
				t.Errorf("GetStyle() = %v, want %v", f.GetStyle(), tc.style)
			}
		})
	}
}

func TestForOutput_FormatMapping(t *testing.T) {
	tests := []struct {
		format string
		want   Style
	}{
		{"json", StyleJSON},
		{"yaml", StyleYAML},
		{"plain", StylePlain},
		{"text", StylePlain},
		{"table", StyleTable},
		{"verbose", StyleVerbose},
		{"compact", StyleCompact},
		{"", StyleColored},        // default
		{"unknown", StyleColored}, // default
	}

	for _, tc := range tests {
		t.Run(tc.format, func(t *testing.T) {
			f := ForOutput(tc.format)
			if f.GetStyle() != tc.want {
				t.Errorf("ForOutput(%q).GetStyle() = %v, want %v", tc.format, f.GetStyle(), tc.want)
			}
		})
	}
}

func TestPlainFormatter_Output(t *testing.T) {
	var buf bytes.Buffer
	f := NewPlainFormatter(FormatterConfig{Writer: &buf}, PlainIcons())

	t.Run("Info", func(t *testing.T) {
		buf.Reset()
		f.Info("test message")
		if !strings.Contains(buf.String(), "test message") {
			t.Errorf("Info output = %q, should contain 'test message'", buf.String())
		}
	})

	t.Run("Success", func(t *testing.T) {
		buf.Reset()
		f.Success("success message")
		if !strings.Contains(buf.String(), "[OK]") || !strings.Contains(buf.String(), "success message") {
			t.Errorf("Success output = %q, should contain '[OK]' and 'success message'", buf.String())
		}
	})

	t.Run("Warning", func(t *testing.T) {
		buf.Reset()
		f.Warning("warning message")
		if !strings.Contains(buf.String(), "[WARN]") {
			t.Errorf("Warning output = %q, should contain '[WARN]'", buf.String())
		}
	})

	t.Run("Error", func(t *testing.T) {
		buf.Reset()
		f.Error("error message")
		if !strings.Contains(buf.String(), "[ERROR]") {
			t.Errorf("Error output = %q, should contain '[ERROR]'", buf.String())
		}
	})

	t.Run("Progress", func(t *testing.T) {
		buf.Reset()
		f.Progress("doing something")
		if !strings.Contains(buf.String(), "->") {
			t.Errorf("Progress output = %q, should contain '->'", buf.String())
		}
	})

	t.Run("Step", func(t *testing.T) {
		buf.Reset()
		f.Step(1, 5, "first step")
		if !strings.Contains(buf.String(), "[1/5]") {
			t.Errorf("Step output = %q, should contain '[1/5]'", buf.String())
		}
	})
}

func TestPlainFormatter_Table(t *testing.T) {
	var buf bytes.Buffer
	f := NewPlainFormatter(FormatterConfig{Writer: &buf}, PlainIcons())

	headers := []string{"NAME", "VALUE", "STATUS"}
	rows := [][]string{
		{"item1", "value1", "active"},
		{"item2", "value2", "inactive"},
	}

	f.Table(headers, rows)
	output := buf.String()

	// Check headers
	if !strings.Contains(output, "NAME") || !strings.Contains(output, "VALUE") {
		t.Error("Table should contain headers")
	}

	// Check rows
	if !strings.Contains(output, "item1") || !strings.Contains(output, "value1") {
		t.Error("Table should contain row data")
	}

	// Check alignment (headers should be at start of lines)
	lines := strings.Split(output, "\n")
	if len(lines) < 4 { // header + separator + 2 rows
		t.Errorf("Table should have at least 4 lines, got %d", len(lines))
	}
}

func TestPlainFormatter_Debug_Verbose(t *testing.T) {
	t.Run("verbose disabled", func(t *testing.T) {
		var buf bytes.Buffer
		f := NewPlainFormatter(FormatterConfig{Writer: &buf, Verbose: false}, PlainIcons())
		f.Debug("debug message")
		if buf.String() != "" {
			t.Error("Debug should not output when verbose is false")
		}
	})

	t.Run("verbose enabled", func(t *testing.T) {
		var buf bytes.Buffer
		f := NewPlainFormatter(FormatterConfig{Writer: &buf, Verbose: true}, PlainIcons())
		f.SetVerbose(true)
		f.Debug("debug message")
		if !strings.Contains(buf.String(), "[DEBUG]") {
			t.Error("Debug should output when verbose is true")
		}
	})
}

func TestJSONFormatter_Output(t *testing.T) {
	var buf bytes.Buffer
	f := NewJSONFormatter(FormatterConfig{Writer: &buf})

	t.Run("Info", func(t *testing.T) {
		buf.Reset()
		f.Info("test message")

		var result map[string]string
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}
		if result["level"] != "info" {
			t.Errorf("level = %q, want 'info'", result["level"])
		}
		if result["message"] != "test message" {
			t.Errorf("message = %q, want 'test message'", result["message"])
		}
	})

	t.Run("Table", func(t *testing.T) {
		buf.Reset()
		headers := []string{"name", "value"}
		rows := [][]string{{"item1", "val1"}, {"item2", "val2"}}
		f.Table(headers, rows)

		var result []map[string]string
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}
		if len(result) != 2 {
			t.Errorf("Expected 2 rows, got %d", len(result))
		}
		if result[0]["name"] != "item1" {
			t.Errorf("First row name = %q, want 'item1'", result[0]["name"])
		}
	})

	t.Run("Object", func(t *testing.T) {
		buf.Reset()
		data := struct {
			Name  string `json:"name"`
			Count int    `json:"count"`
		}{"test", 42}

		if err := f.Object(data); err != nil {
			t.Fatalf("Object() error: %v", err)
		}

		var result map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}
		if result["name"] != "test" {
			t.Errorf("name = %v, want 'test'", result["name"])
		}
	})
}

func TestYAMLFormatter_Output(t *testing.T) {
	var buf bytes.Buffer
	f := NewYAMLFormatter(FormatterConfig{Writer: &buf})

	t.Run("List", func(t *testing.T) {
		buf.Reset()
		f.List([]string{"item1", "item2", "item3"})
		output := buf.String()

		if !strings.Contains(output, "- item1") {
			t.Errorf("YAML list should contain '- item1', got: %s", output)
		}
	})

	t.Run("KeyValue", func(t *testing.T) {
		buf.Reset()
		f.KeyValue(map[string]string{"key1": "val1", "key2": "val2"})
		output := buf.String()

		if !strings.Contains(output, "key1: val1") && !strings.Contains(output, "key1: \"val1\"") {
			t.Errorf("YAML should contain key-value pairs, got: %s", output)
		}
	})
}

func TestColoredFormatter_Output(t *testing.T) {
	var buf bytes.Buffer
	f := NewColoredFormatter(FormatterConfig{Writer: &buf}, DefaultIcons())

	t.Run("Success contains icon", func(t *testing.T) {
		buf.Reset()
		f.Success("success message")
		output := buf.String()

		// Should contain the success icon
		if !strings.Contains(output, DefaultIcons().Success) {
			t.Errorf("Success output should contain success icon, got: %s", output)
		}
	})

	t.Run("Table with data", func(t *testing.T) {
		buf.Reset()
		headers := []string{"COL1", "COL2"}
		rows := [][]string{{"a", "b"}, {"c", "d"}}
		f.Table(headers, rows)
		output := buf.String()

		if !strings.Contains(output, "COL1") {
			t.Errorf("Table should contain headers, got: %s", output)
		}
	})

	t.Run("Empty table", func(t *testing.T) {
		buf.Reset()
		f.Table([]string{"H1", "H2"}, [][]string{})
		output := buf.String()

		if !strings.Contains(output, "No data") {
			t.Errorf("Empty table should show 'No data', got: %s", output)
		}
	})
}

func TestVerboseFormatter_Timestamps(t *testing.T) {
	var buf bytes.Buffer
	f := NewVerboseFormatter(FormatterConfig{Writer: &buf}, DefaultIcons())

	f.Info("test message")
	output := buf.String()

	// Should contain a timestamp pattern (HH:MM:SS)
	if !strings.Contains(output, ":") {
		t.Errorf("Verbose output should contain timestamp, got: %s", output)
	}
}

func TestCompactFormatter_Table(t *testing.T) {
	var buf bytes.Buffer
	f := NewCompactFormatter(FormatterConfig{Writer: &buf}, DefaultIcons())

	headers := []string{"A", "B"}
	rows := [][]string{{"1", "2"}}
	f.Table(headers, rows)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Compact table should have only 2 lines (header + 1 row, no separator)
	if len(lines) != 2 {
		t.Errorf("Compact table should have 2 lines, got %d: %v", len(lines), lines)
	}
}

func TestTableOnlyFormatter_Suppression(t *testing.T) {
	var buf bytes.Buffer
	f := NewTableOnlyFormatter(FormatterConfig{Writer: &buf}, DefaultIcons())

	// These should not produce output
	f.Info("ignored")
	f.Success("ignored")
	f.Warning("ignored")
	f.Error("ignored")
	f.Progress("ignored")

	if buf.String() != "" {
		t.Errorf("TableOnlyFormatter should suppress non-table output, got: %s", buf.String())
	}

	// Table should still work
	f.Table([]string{"H"}, [][]string{{"V"}})
	if !strings.Contains(buf.String(), "H") {
		t.Error("TableOnlyFormatter should output tables")
	}
}

func TestWithWriter(t *testing.T) {
	var buf bytes.Buffer
	f := WithWriter(&buf, StylePlain)

	f.Info("test")
	if buf.String() == "" {
		t.Error("WithWriter should write to the provided buffer")
	}
}

func TestIcons(t *testing.T) {
	t.Run("DefaultIcons", func(t *testing.T) {
		icons := DefaultIcons()
		if icons.Success == "" {
			t.Error("DefaultIcons should have non-empty Success")
		}
		if icons.Error == "" {
			t.Error("DefaultIcons should have non-empty Error")
		}
	})

	t.Run("NerdFontIcons", func(t *testing.T) {
		icons := NerdFontIcons()
		// Nerd Font icons are defined - just verify the struct is populated
		// The actual glyphs may not render in test output
		if icons.Bullet == "" {
			t.Error("NerdFontIcons should have non-empty Bullet")
		}
	})

	t.Run("PlainIcons", func(t *testing.T) {
		icons := PlainIcons()
		// Plain icons should be ASCII only
		if !strings.HasPrefix(icons.Success, "[") {
			t.Errorf("PlainIcons.Success should be ASCII, got: %s", icons.Success)
		}
	})
}

func TestFormatterSetters(t *testing.T) {
	var buf1, buf2 bytes.Buffer

	f := NewPlainFormatter(FormatterConfig{Writer: &buf1}, PlainIcons())

	f.Info("to buf1")
	if buf1.String() == "" {
		t.Error("Should write to buf1")
	}

	f.SetWriter(&buf2)
	f.Info("to buf2")
	if buf2.String() == "" {
		t.Error("Should write to buf2 after SetWriter")
	}

	f.SetVerbose(true)
	buf2.Reset()
	f.Debug("debug message")
	if buf2.String() == "" {
		t.Error("Should output debug after SetVerbose(true)")
	}
}
