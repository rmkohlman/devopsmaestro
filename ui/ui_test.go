package ui

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatProject(t *testing.T) {
	// Test active project
	result := FormatProject("my-project", true)
	assert.Contains(t, result, IconProject)
	assert.Contains(t, result, "my-project")

	// Test inactive project
	result = FormatProject("my-project", false)
	assert.Contains(t, result, IconProject)
	assert.Contains(t, result, "my-project")
}

func TestFormatWorkspace(t *testing.T) {
	// Test active workspace
	result := FormatWorkspace("main", true)
	assert.Contains(t, result, IconWorkspace)
	assert.Contains(t, result, "main")

	// Test inactive workspace
	result = FormatWorkspace("main", false)
	assert.Contains(t, result, IconWorkspace)
	assert.Contains(t, result, "main")
}

func TestFormatPlugin(t *testing.T) {
	// Test enabled plugin
	result := FormatPlugin("telescope", true)
	assert.Contains(t, result, IconPlugin)
	assert.Contains(t, result, "telescope")

	// Test disabled plugin
	result = FormatPlugin("telescope", false)
	assert.Contains(t, result, IconPlugin)
	assert.Contains(t, result, "telescope")
}

func TestFormatContext(t *testing.T) {
	result := FormatContext("test-project", "main-workspace")
	assert.Contains(t, result, IconContext)
	assert.Contains(t, result, "test-project")
	assert.Contains(t, result, "main-workspace")
	assert.Contains(t, result, "/")
}

func TestFormatBuildStep(t *testing.T) {
	// Test successful step
	result := FormatBuildStep("Build completed", true)
	assert.Contains(t, result, IconSuccess)
	assert.Contains(t, result, "Build completed")

	// Test in-progress step
	result = FormatBuildStep("Building...", false)
	assert.Contains(t, result, IconBuild)
	assert.Contains(t, result, "Building...")
}

func TestFormatResourceType(t *testing.T) {
	tests := []struct {
		resourceType string
		expectedIcon string
	}{
		{"project", IconProject},
		{"workspace", IconWorkspace},
		{"plugin", IconPlugin},
		{"pipeline", IconPipeline},
		{"workflow", IconWorkflow},
		{"task", IconTask},
		{"database", IconDatabase},
		{"datalake", IconDataLake},
		{"datastore", IconDataStore},
		{"storage", IconStorage},
		{"dependency", IconDependency},
		{"prototype", IconPrototype},
		{"unknown", IconFile},
	}

	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			result := FormatResourceType(tt.resourceType)
			assert.Contains(t, result, tt.expectedIcon, "Expected icon %s for resource type %s", tt.expectedIcon, tt.resourceType)
			assert.Contains(t, result, strings.ToUpper(tt.resourceType))
		})
	}
}

func TestRenderSuccess(t *testing.T) {
	result := RenderSuccess("Operation completed")
	assert.Contains(t, result, CheckMark)
	assert.Contains(t, result, "Operation completed")
}

func TestRenderError(t *testing.T) {
	result := RenderError("Operation failed")
	assert.Contains(t, result, CrossMark)
	assert.Contains(t, result, "Operation failed")
}

func TestRenderStatus(t *testing.T) {
	// Test active status
	result := RenderStatus("active")
	assert.Contains(t, result, "active")

	// Test inactive status
	result = RenderStatus("inactive")
	assert.Contains(t, result, "inactive")

	// Test running status
	result = RenderStatus("running")
	assert.Contains(t, result, "running")

	// Test stopped status
	result = RenderStatus("stopped")
	assert.Contains(t, result, "stopped")
}

func TestTableRenderer_AddRow(t *testing.T) {
	tr := NewTableRenderer([]string{"NAME", "VALUE"})
	tr.AddRow("test1", "value1")
	tr.AddRow("test2", "value2")

	assert.Len(t, tr.rows, 2)
	assert.Equal(t, "test1", tr.rows[0][0])
	assert.Equal(t, "value1", tr.rows[0][1])
}

func TestTableRenderer_RenderSimple(t *testing.T) {
	tr := NewTableRenderer([]string{"NAME", "VALUE"})
	tr.AddRow("test1", "value1")
	tr.AddRow("test2", "value2")

	result := tr.RenderSimple()
	assert.NotEmpty(t, result)
	assert.Contains(t, result, "NAME")
	assert.Contains(t, result, "VALUE")
	assert.Contains(t, result, "test1")
	assert.Contains(t, result, "value1")
}

func TestTableRenderer_EmptyTable(t *testing.T) {
	tr := NewTableRenderer([]string{"NAME", "VALUE"})

	result := tr.RenderSimple()
	assert.Contains(t, result, "No data available")
}

func TestProjectsTableRenderer(t *testing.T) {
	tr := ProjectsTableRenderer("active-project")
	assert.NotNil(t, tr)
	assert.Len(t, tr.headers, 3)
	assert.Contains(t, tr.headers[0], IconProject)
	assert.Contains(t, tr.headers[2], IconTime)
}

func TestWorkspacesTableRenderer(t *testing.T) {
	tr := WorkspacesTableRenderer("active-workspace")
	assert.NotNil(t, tr)
	assert.Len(t, tr.headers, 5)
	assert.Contains(t, tr.headers[0], IconWorkspace)
	assert.Contains(t, tr.headers[1], IconProject)
	assert.Contains(t, tr.headers[2], IconImage)
	assert.Contains(t, tr.headers[4], IconTime)
}

func TestPluginsTableRenderer(t *testing.T) {
	tr := PluginsTableRenderer([]string{})
	assert.NotNil(t, tr)
	assert.Len(t, tr.headers, 4)
	assert.Contains(t, tr.headers[0], IconPlugin)
	assert.Contains(t, tr.headers[1], IconCategory)
	assert.Contains(t, tr.headers[2], IconGit)
	assert.Contains(t, tr.headers[3], IconTag)
}

func TestFormatActiveItem(t *testing.T) {
	// Test active item
	result := FormatActiveItem("test-item", true)
	assert.Contains(t, result, ActiveIndicator)
	assert.Contains(t, result, "test-item")

	// Test inactive item
	result = FormatActiveItem("test-item", false)
	assert.NotContains(t, result, ActiveIndicator)
	assert.Contains(t, result, "test-item")
}

func TestProgressMessage(t *testing.T) {
	result := ProgressMessage("Loading resources...")
	assert.Contains(t, result, Arrow)
	assert.Contains(t, result, "Loading resources...")
}

func TestStepMessage(t *testing.T) {
	// Test successful step
	result := StepMessage("Completed successfully", true)
	assert.Contains(t, result, CheckMark)
	assert.Contains(t, result, "Completed successfully")

	// Test failed step
	result = StepMessage("Failed to complete", false)
	assert.Contains(t, result, CrossMark)
	assert.Contains(t, result, "Failed to complete")
}

func TestSectionHeader(t *testing.T) {
	result := SectionHeader("Test Section")
	assert.Contains(t, result, "Test Section")
	assert.NotEmpty(t, result)
}

func TestInfoBox(t *testing.T) {
	result := InfoBox("Important information")
	assert.Contains(t, result, "Important information")
}

func TestSuccessBox(t *testing.T) {
	result := SuccessBox("Operation succeeded")
	assert.Contains(t, result, CheckMark)
	assert.Contains(t, result, "Operation succeeded")
}

func TestErrorBox(t *testing.T) {
	result := ErrorBox("Operation failed")
	assert.Contains(t, result, CrossMark)
	assert.Contains(t, result, "Operation failed")
}

func TestListItem(t *testing.T) {
	// Test level 0
	result := ListItem("Item 1", 0)
	assert.Contains(t, result, Bullet)
	assert.Contains(t, result, "Item 1")

	// Test level 1 (should have indentation)
	result = ListItem("Subitem", 1)
	assert.Contains(t, result, Bullet)
	assert.Contains(t, result, "Subitem")
	assert.True(t, strings.HasPrefix(result, "  "))
}

func TestStripANSI(t *testing.T) {
	// Test with ANSI codes
	input := "\x1b[31mRed Text\x1b[0m"
	result := stripANSI(input)
	assert.Equal(t, "Red Text", result)

	// Test without ANSI codes
	input = "Plain Text"
	result = stripANSI(input)
	assert.Equal(t, "Plain Text", result)
}
