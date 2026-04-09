package cmd

// Tests for Issue #207: Beautiful Table Output (TDD Phase 2 — RED)
//
// These tests verify that commands using --output table produce bordered,
// lipgloss-styled output with box-drawing characters, while other formats
// (json, yaml, plain) remain backward-compatible.
//
// All tests are expected to FAIL until the PrettyRenderer implementation is added.

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/rmkohlman/MaestroSDK/render"
)

// ---------------------------------------------------------------------------
// Helper: detect box-drawing characters that indicate a true bordered table.
// A proper bordered table has vertical bars AND corners — not just horizontal
// separator lines (which the current colored renderer already uses).
// ---------------------------------------------------------------------------

func containsBoxDrawing(s string) bool {
	// Require at least a vertical bar — the current colored renderer only
	// uses "─" for separators, but a true bordered table has "│" as well.
	verticalChars := []string{"│", "┌", "┐", "└", "┘", "├", "┤", "┬", "┴", "┼"}
	for _, ch := range verticalChars {
		if strings.Contains(s, ch) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Helper: detect ANSI escape codes in output
// ---------------------------------------------------------------------------

func containsANSI(s string) bool {
	return strings.Contains(s, "\x1b[")
}

// ---------------------------------------------------------------------------
// 1. --output table produces bordered output for TableData
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

	// After issue #207, the "table" renderer should produce bordered output
	// with vertical box-drawing characters (│, ┌, ┐, └, ┘, etc.)
	err := r.Render(&buf, data, render.Options{Type: render.TypeTable})
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	output := buf.String()
	if !containsBoxDrawing(output) {
		t.Errorf("--output table should produce bordered output with box-drawing characters (│, ┌, ┐, etc.).\nOutput:\n%s", output)
	}
}

func TestRenderTableData_WithPrettyFormat_ProducesBorderedOutput(t *testing.T) {
	// "pretty" format must produce bordered table output
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
	if !containsBoxDrawing(output) {
		t.Errorf("--output pretty should produce bordered output.\nOutput:\n%s", output)
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

	// Both should have the same bordered output (alias maps to same renderer)
	coloredHasBoxes := containsBoxDrawing(bufColored.String())
	tableHasBoxes := containsBoxDrawing(bufTable.String())

	if tableHasBoxes && !coloredHasBoxes {
		t.Errorf("'colored' alias should produce the same bordered output as 'table'.\nTable output:\n%s\nColored output:\n%s",
			bufTable.String(), bufColored.String())
	}
}
