package cmd

// Tests for Issue #207: Table Output Rendering
//
// These tests verify that commands using --output table produce the expected
// output format (headers + horizontal dividers + data rows), while other
// formats (json, yaml, plain) remain backward-compatible.
//
// NOTE: Tests updated in Issue #229 to match the actual renderer behavior
// in MaestroSDK v0.1.6 (unbordered format with horizontal dividers).
// NOTE: Tests updated in Issue #230 to verify the new bordered table format
// with alternating row backgrounds and box-drawing borders.

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rmkohlman/MaestroSDK/render"
)

// ---------------------------------------------------------------------------
// Helper: detect horizontal separator characters used by the table renderer.
// The current renderer (MaestroSDK v0.1.6) produces headers + horizontal
// dividers (─) + data rows — without vertical borders.
// ---------------------------------------------------------------------------

func containsHorizontalDivider(s string) bool {
	return strings.Contains(s, "─")
}

// containsVerticalBorder detects vertical box-drawing borders (│).
func containsVerticalBorder(s string) bool {
	return strings.Contains(s, "│")
}

// containsBoxCorners detects box-drawing corner characters.
func containsBoxCorners(s string) bool {
	return strings.Contains(s, "┌") && strings.Contains(s, "┐") &&
		strings.Contains(s, "└") && strings.Contains(s, "┘")
}

// ---------------------------------------------------------------------------
// Helper: detect ANSI escape codes in output
// ---------------------------------------------------------------------------

func containsANSI(s string) bool {
	return strings.Contains(s, "\x1b[")
}

// ---------------------------------------------------------------------------
// 1. --output table produces expected format for TableData
// ---------------------------------------------------------------------------

func TestRenderTableData_WithTableFormat_ProducesBorderedOutput(t *testing.T) {
	r := render.Get(render.RendererTable)
	if r == nil {
		t.Fatal("Table renderer not registered")
	}

	var buf bytes.Buffer
	data := render.TableData{
		Headers: []string{"NAME", "STATUS", "APP"},
		Rows: [][]string{
			{"main", "running", "portal"},
			{"dev", "stopped", "api"},
		},
	}

	// The "table" renderer produces: headers, horizontal divider line (─), then data rows.
	// MaestroSDK v0.1.6 uses an unbordered format (no vertical box-drawing characters).
	err := r.Render(&buf, data, render.Options{Type: render.TypeTable})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()
	if !containsHorizontalDivider(output) {
		t.Errorf("--output table should produce output with horizontal divider (─) separating headers from rows.\nOutput:\n%s", output)
	}

	// Headers should appear in output
	if !strings.Contains(output, "NAME") || !strings.Contains(output, "STATUS") || !strings.Contains(output, "APP") {
		t.Errorf("--output table should include column headers.\nOutput:\n%s", output)
	}

	// Data rows should appear in output
	if !strings.Contains(output, "main") || !strings.Contains(output, "running") {
		t.Errorf("--output table should include data rows.\nOutput:\n%s", output)
	}
}

func TestRenderTableData_WithPrettyFormat_ProducesBorderedOutput(t *testing.T) {
	// "pretty" format falls back to the colored renderer in MaestroSDK v0.1.6
	// (no RendererPretty registered), producing headers + horizontal dividers + rows.
	var buf bytes.Buffer
	data := render.TableData{
		Headers: []string{"ECOSYSTEM", "DOMAINS", "APPS"},
		Rows: [][]string{
			{"prod", "3", "5"},
			{"staging", "2", "3"},
		},
	}

	err := render.OutputTo(&buf, "pretty", data, render.Options{Type: render.TypeTable})
	if err != nil {
		t.Fatalf("OutputTo('pretty') failed: %v", err)
	}

	output := buf.String()
	if !containsHorizontalDivider(output) {
		t.Errorf("--output pretty should produce output with horizontal divider (─) separating headers from rows.\nOutput:\n%s", output)
	}

	// Headers should appear in output
	if !strings.Contains(output, "ECOSYSTEM") || !strings.Contains(output, "DOMAINS") || !strings.Contains(output, "APPS") {
		t.Errorf("--output pretty should include column headers.\nOutput:\n%s", output)
	}

	// Data rows should appear in output
	if !strings.Contains(output, "prod") || !strings.Contains(output, "staging") {
		t.Errorf("--output pretty should include data rows.\nOutput:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// 2. --output json produces valid JSON (backward compat)
// ---------------------------------------------------------------------------

func TestRenderTableData_WithJSONFormat_ProducesValidJSON(t *testing.T) {
	var buf bytes.Buffer
	data := render.TableData{
		Headers: []string{"NAME", "STATUS"},
		Rows: [][]string{
			{"workspace-1", "active"},
			{"workspace-2", "stopped"},
		},
	}

	err := render.OutputTo(&buf, "json", data, render.Options{Type: render.TypeTable})
	if err != nil {
		t.Fatalf("OutputTo('json') failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("JSON output should not be empty")
	}

	// Must be valid JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("--output json should produce valid JSON.\nOutput: %s\nError: %v", output, err)
	}
}

// ---------------------------------------------------------------------------
// 3. --output yaml produces valid YAML (backward compat)
// ---------------------------------------------------------------------------

func TestRenderTableData_WithYAMLFormat_ProducesYAMLOutput(t *testing.T) {
	var buf bytes.Buffer
	data := render.TableData{
		Headers: []string{"NAME", "STATUS"},
		Rows: [][]string{
			{"workspace-1", "active"},
		},
	}

	err := render.OutputTo(&buf, "yaml", data, render.Options{Type: render.TypeTable})
	if err != nil {
		t.Fatalf("OutputTo('yaml') failed: %v", err)
	}

	output := buf.String()
	if output == "" {
		t.Fatal("YAML output should not be empty")
	}

	// YAML output should contain at least one field from the data
	if !strings.Contains(output, "workspace-1") && !strings.Contains(output, "active") {
		t.Errorf("YAML output should contain table data.\nOutput:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// 4. --output plain produces NO ANSI escape codes
// ---------------------------------------------------------------------------

func TestRenderTableData_WithPlainFormat_ProducesNoANSI(t *testing.T) {
	var buf bytes.Buffer
	data := render.TableData{
		Headers: []string{"NAME", "STATUS"},
		Rows: [][]string{
			{"workspace-1", "active"},
			{"workspace-2", "stopped"},
		},
	}

	err := render.OutputTo(&buf, "plain", data, render.Options{Type: render.TypeTable})
	if err != nil {
		t.Fatalf("OutputTo('plain') failed: %v", err)
	}

	output := buf.String()
	if containsANSI(output) {
		t.Errorf("--output plain should produce no ANSI escape codes.\nOutput:\n%s", output)
	}
}

func TestRenderMessage_WithPlainFormat_ProducesNoANSI(t *testing.T) {
	var buf bytes.Buffer
	r := render.Get(render.RendererPlain)
	if r == nil {
		t.Fatal("Plain renderer not registered")
	}

	err := r.RenderMessage(&buf, render.Message{
		Level:   render.LevelSuccess,
		Content: "Operation completed successfully",
	})
	if err != nil {
		t.Fatalf("RenderMessage failed: %v", err)
	}

	output := buf.String()
	if containsANSI(output) {
		t.Errorf("Plain renderer RenderMessage should produce no ANSI codes.\nOutput:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// 5. "colored" format alias produces same output type as "table"
// ---------------------------------------------------------------------------

func TestRenderTableData_ColoredAlias_MapsToTableRenderer(t *testing.T) {
	// The "colored" alias should resolve to the table renderer for TableData
	var bufColored, bufTable bytes.Buffer

	data := render.TableData{
		Headers: []string{"NAME"},
		Rows:    [][]string{{"ws-1"}},
	}

	errColored := render.OutputTo(&bufColored, "colored", data, render.Options{Type: render.TypeTable})
	errTable := render.OutputTo(&bufTable, "table", data, render.Options{Type: render.TypeTable})

	if errColored != nil {
		t.Fatalf("OutputTo('colored') failed: %v", errColored)
	}
	if errTable != nil {
		t.Fatalf("OutputTo('table') failed: %v", errTable)
	}

	// Both should produce divider output — they map to the same rendering style
	coloredHasDivider := containsHorizontalDivider(bufColored.String())
	tableHasDivider := containsHorizontalDivider(bufTable.String())

	if tableHasDivider && !coloredHasDivider {
		t.Errorf("'colored' alias should produce the same output style as 'table'.\nTable output:\n%s\nColored output:\n%s",
			bufTable.String(), bufColored.String())
	}
}

// ---------------------------------------------------------------------------
// 6. Colored renderer produces bordered table with box-drawing (Issue #230)
// ---------------------------------------------------------------------------

func TestRenderTableData_ColoredFormat_HasBoxBorders(t *testing.T) {
	r := render.Get(render.RendererColored)
	if r == nil {
		t.Fatal("Colored renderer not registered")
	}

	var buf bytes.Buffer
	data := render.TableData{
		Headers: []string{"NAME", "STATUS", "APP"},
		Rows: [][]string{
			{"main", "running", "portal"},
			{"dev", "stopped", "api"},
		},
	}

	err := r.Render(&buf, data, render.Options{Type: render.TypeTable})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()

	// Should have box-drawing borders
	if !containsVerticalBorder(output) {
		t.Errorf("colored table should have vertical borders (│).\nOutput:\n%s", output)
	}
	if !containsBoxCorners(output) {
		t.Errorf("colored table should have box corners (┌┐└┘).\nOutput:\n%s", output)
	}

	// Headers and data should still be present
	if !strings.Contains(output, "NAME") || !strings.Contains(output, "STATUS") {
		t.Errorf("colored table should include column headers.\nOutput:\n%s", output)
	}
	if !strings.Contains(output, "main") || !strings.Contains(output, "running") {
		t.Errorf("colored table should include data rows.\nOutput:\n%s", output)
	}
}

// ---------------------------------------------------------------------------
// 7. --output json/yaml still work as overrides (backward compat with #230)
// ---------------------------------------------------------------------------

func TestRenderTableData_JSONOverride_StillWorks(t *testing.T) {
	var buf bytes.Buffer
	data := render.TableData{
		Headers: []string{"NAME", "STATUS"},
		Rows: [][]string{
			{"ws-1", "active"},
		},
	}

	err := render.OutputTo(&buf, "json", data, render.Options{Type: render.TypeTable})
	if err != nil {
		t.Fatalf("OutputTo('json') failed: %v", err)
	}

	output := buf.String()
	// JSON output should not contain box-drawing
	if containsVerticalBorder(output) || containsBoxCorners(output) {
		t.Errorf("--output json should not produce box-drawing borders.\nOutput:\n%s", output)
	}

	// Must be valid JSON
	var parsed interface{}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Errorf("--output json should produce valid JSON.\nOutput: %s\nError: %v", output, err)
	}
}

func TestRenderTableData_YAMLOverride_StillWorks(t *testing.T) {
	var buf bytes.Buffer
	data := render.TableData{
		Headers: []string{"NAME", "STATUS"},
		Rows:    [][]string{{"ws-1", "active"}},
	}

	err := render.OutputTo(&buf, "yaml", data, render.Options{Type: render.TypeTable})
	if err != nil {
		t.Fatalf("OutputTo('yaml') failed: %v", err)
	}

	output := buf.String()
	// YAML output should not contain box-drawing
	if containsVerticalBorder(output) || containsBoxCorners(output) {
		t.Errorf("--output yaml should not produce box-drawing borders.\nOutput:\n%s", output)
	}
	// Should still contain data
	if !strings.Contains(output, "ws-1") {
		t.Errorf("YAML output should contain table data.\nOutput:\n%s", output)
	}
}
