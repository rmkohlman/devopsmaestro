package cmd

// =============================================================================
// Verification: Issue #401 — dvm system df table output formatting
// =============================================================================
// Bug: header rows were emitted via render.Info (which prefixes "ℹ ")
// while data rows went through render.Plain, causing columns to misalign
// and a literal "ℹ" symbol to appear at the start of header rows.
//
// Fix: render the entire table through a single tabwriter writer.
// =============================================================================

import (
	"bytes"
	"strings"
	"testing"

	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSystemDF_RenderTable_NoInfoPrefixAndAligned(t *testing.T) {
	// Capture render output.
	var buf bytes.Buffer
	orig := render.GetWriter()
	render.SetWriter(&buf)
	defer render.SetWriter(orig)

	categories := []DFCategory{
		{Type: "Build Cache", Count: 0, Active: 0, Size: "0 B", Reclaimable: "0 B"},
		{Type: "Build Staging", Count: 1103, Active: 0, Size: "3.9 MB", Reclaimable: "3.9 MB"},
		{Type: "Registries", Count: 65, Active: 0, Size: "274.4 MB", Reclaimable: "274.4 MB"},
	}

	renderDFTable(categories)

	output := buf.String()
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")

	// 1. No line should contain the info-log prefix character "ℹ".
	for i, line := range lines {
		assert.NotContains(t, line, "ℹ",
			"line %d must not contain the info-log prefix: %q", i, line)
	}

	// 2. Locate the header line (starts with "TYPE").
	var headerIdx = -1
	for i, line := range lines {
		if strings.HasPrefix(line, "TYPE") {
			headerIdx = i
			break
		}
	}
	require.GreaterOrEqual(t, headerIdx, 0, "header row must be present")

	header := lines[headerIdx]
	// Header must contain all column titles in order.
	for _, col := range []string{"TYPE", "COUNT", "ACTIVE", "SIZE", "RECLAIMABLE"} {
		assert.Contains(t, header, col, "header missing column %q", col)
	}

	// 3. Verify data rows are present and tabular.
	dataTypes := map[string]bool{
		"Build Cache":   true,
		"Build Staging": true,
		"Registries":    true,
	}
	dataRowsFound := 0
	for _, line := range lines {
		matched := false
		for prefix := range dataTypes {
			if strings.HasPrefix(line, prefix) {
				matched = true
				break
			}
		}
		if !matched {
			continue
		}
		dataRowsFound++

		// Tabwriter pads with at least 2 spaces between columns. Verify
		// the data row is consistent with the tabular layout: it must
		// contain at least 4 multi-space separators (one between each
		// of the 5 columns), proving the row participated in the same
		// tabwriter as the header.
		separators := strings.Count(line, "  ")
		assert.GreaterOrEqual(t, separators, 4,
			"data row %q does not look tabular (need >=4 multi-space separators, got %d)",
			line, separators)
	}
	assert.Equal(t, 3, dataRowsFound, "expected 3 data rows in output, got %d", dataRowsFound)
}
