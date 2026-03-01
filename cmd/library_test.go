//go:build integration
// +build integration

// Package cmd - library_test.go contains tests for the `dvm library` command group.
// These tests are tagged as integration tests because they require the library
// command implementation (TDD Phase 2 - RED state).
//
// Run these tests with: go test -tags=integration ./cmd/...
// Or run all tests except these: go test ./cmd/... (default, no tag)

package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ========== Test Library List Commands ==========

// TestLibraryListPlugins tests listing nvim plugins from the library
func TestLibraryListPlugins(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		wantErr        bool
		wantOutputType string // "table", "yaml", "json"
		wantContains   []string
	}{
		{
			name:           "list nvim plugins with default output",
			args:           []string{"library", "list", "plugins"},
			wantErr:        false,
			wantOutputType: "table",
			wantContains:   []string{"telescope", "treesitter", "lspconfig"},
		},
		{
			name:           "list nvim plugins with yaml output",
			args:           []string{"library", "list", "plugins", "-o", "yaml"},
			wantErr:        false,
			wantOutputType: "yaml",
			wantContains:   []string{"name:", "description:", "telescope"},
		},
		{
			name:           "list nvim plugins with json output",
			args:           []string{"library", "list", "plugins", "-o", "json"},
			wantErr:        false,
			wantOutputType: "json",
			wantContains:   []string{`"name"`, `"description"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create root command with library subcommand
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			// Capture output
			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			// Execute
			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := buf.String()

			// Verify output type and content
			for _, want := range tt.wantContains {
				assert.Contains(t, output, want, "output should contain %q", want)
			}

			// Verify output format
			switch tt.wantOutputType {
			case "yaml":
				var parsed interface{}
				err := yaml.Unmarshal([]byte(output), &parsed)
				assert.NoError(t, err, "output should be valid YAML")
			case "json":
				var parsed interface{}
				err := json.Unmarshal([]byte(output), &parsed)
				assert.NoError(t, err, "output should be valid JSON")
			case "table":
				// Table format should have header
				assert.True(t, strings.Contains(output, "NAME") || strings.Contains(output, "DESCRIPTION"),
					"table output should have headers")
			}
		})
	}
}

// TestLibraryListThemes tests listing nvim themes from the library
func TestLibraryListThemes(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
	}{
		{
			name:         "list nvim themes",
			args:         []string{"library", "list", "themes"},
			wantErr:      false,
			wantContains: []string{"coolnight", "tokyonight"},
		},
		{
			name:         "list nvim themes with yaml",
			args:         []string{"library", "list", "themes", "-o", "yaml"},
			wantErr:      false,
			wantContains: []string{"name:", "description:"},
		},
		{
			name:         "list nvim themes with json",
			args:         []string{"library", "list", "themes", "-o", "json"},
			wantErr:      false,
			wantContains: []string{`"name"`, `"description"`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := buf.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want)
			}
		})
	}
}

// TestLibraryListNvimPackages tests listing nvim packages from the library
func TestLibraryListNvimPackages(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
	}{
		{
			name:    "list nvim packages",
			args:    []string{"library", "list", "nvim", "packages"},
			wantErr: false,
			// May be empty initially, but command should succeed
		},
		{
			name:    "list nvim packages with yaml",
			args:    []string{"library", "list", "nvim", "packages", "-o", "yaml"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := buf.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want)
			}
		})
	}
}

// TestLibraryListTerminalPrompts tests listing terminal prompts from the library
func TestLibraryListTerminalPrompts(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
	}{
		{
			name:         "list terminal prompts",
			args:         []string{"library", "list", "terminal", "prompts"},
			wantErr:      false,
			wantContains: []string{"starship"},
		},
		{
			name:    "list terminal prompts with yaml",
			args:    []string{"library", "list", "terminal", "prompts", "-o", "yaml"},
			wantErr: false,
		},
		{
			name:    "list terminal prompts with json",
			args:    []string{"library", "list", "terminal", "prompts", "-o", "json"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := buf.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want)
			}
		})
	}
}

// TestLibraryListTerminalPlugins tests listing terminal/shell plugins from the library
func TestLibraryListTerminalPlugins(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
	}{
		{
			name:         "list terminal plugins",
			args:         []string{"library", "list", "terminal", "plugins"},
			wantErr:      false,
			wantContains: []string{"zsh-autosuggestions"},
		},
		{
			name:    "list terminal plugins with yaml",
			args:    []string{"library", "list", "terminal", "plugins", "-o", "yaml"},
			wantErr: false,
		},
		{
			name:    "list terminal plugins with json",
			args:    []string{"library", "list", "terminal", "plugins", "-o", "json"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			output := buf.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want)
			}
		})
	}
}

// TestLibraryListTerminalPackages tests listing terminal packages from the library
func TestLibraryListTerminalPackages(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "list terminal packages",
			args:    []string{"library", "list", "terminal", "packages"},
			wantErr: false,
		},
		{
			name:    "list terminal packages with yaml",
			args:    []string{"library", "list", "terminal", "packages", "-o", "yaml"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

// ========== Test Library Show Commands ==========

// TestLibraryShowPlugin tests showing details of a specific nvim plugin
func TestLibraryShowPlugin(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
		wantNotFound bool
	}{
		{
			name:         "show telescope plugin",
			args:         []string{"library", "show", "plugin", "telescope"},
			wantErr:      false,
			wantContains: []string{"telescope", "fuzzy", "finder"},
		},
		{
			name:         "show telescope plugin with yaml",
			args:         []string{"library", "show", "plugin", "telescope", "-o", "yaml"},
			wantErr:      false,
			wantContains: []string{"name:", "telescope", "description:"},
		},
		{
			name:         "show telescope plugin with json",
			args:         []string{"library", "show", "plugin", "telescope", "-o", "json"},
			wantErr:      false,
			wantContains: []string{`"name"`, `"telescope"`},
		},
		{
			name:         "show unknown plugin returns error",
			args:         []string{"library", "show", "plugin", "unknown-plugin-xyz"},
			wantErr:      true,
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					output := buf.String()
					assert.Contains(t, output, "not found", "error should indicate resource not found")
				}
				return
			}

			require.NoError(t, err)
			output := buf.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want)
			}
		})
	}
}

// TestLibraryShowTheme tests showing details of a specific nvim theme
func TestLibraryShowTheme(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
		wantNotFound bool
	}{
		{
			name:         "show coolnight-ocean theme",
			args:         []string{"library", "show", "theme", "coolnight-ocean"},
			wantErr:      false,
			wantContains: []string{"coolnight", "ocean"},
		},
		{
			name:         "show theme with yaml",
			args:         []string{"library", "show", "theme", "coolnight-ocean", "-o", "yaml"},
			wantErr:      false,
			wantContains: []string{"name:", "description:"},
		},
		{
			name:         "show unknown theme returns error",
			args:         []string{"library", "show", "theme", "unknown-theme-xyz"},
			wantErr:      true,
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					output := buf.String()
					assert.Contains(t, output, "not found", "error should indicate resource not found")
				}
				return
			}

			require.NoError(t, err)
			output := buf.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want)
			}
		})
	}
}

// TestLibraryShowTerminalPrompt tests showing details of a terminal prompt
func TestLibraryShowTerminalPrompt(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
		wantNotFound bool
	}{
		{
			name:         "show starship-default prompt",
			args:         []string{"library", "show", "terminal", "prompt", "starship-default"},
			wantErr:      false,
			wantContains: []string{"starship"},
		},
		{
			name:    "show prompt with yaml",
			args:    []string{"library", "show", "terminal", "prompt", "starship-default", "-o", "yaml"},
			wantErr: false,
		},
		{
			name:         "show unknown prompt returns error",
			args:         []string{"library", "show", "terminal", "prompt", "unknown-prompt-xyz"},
			wantErr:      true,
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					output := buf.String()
					assert.Contains(t, output, "not found", "error should indicate resource not found")
				}
				return
			}

			require.NoError(t, err)
			output := buf.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want)
			}
		})
	}
}

// TestLibraryShowTerminalPlugin tests showing details of a terminal plugin
func TestLibraryShowTerminalPlugin(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		wantContains []string
		wantNotFound bool
	}{
		{
			name:         "show zsh-autosuggestions plugin",
			args:         []string{"library", "show", "terminal", "plugin", "zsh-autosuggestions"},
			wantErr:      false,
			wantContains: []string{"autosuggestions"},
		},
		{
			name:    "show terminal plugin with json",
			args:    []string{"library", "show", "terminal", "plugin", "zsh-autosuggestions", "-o", "json"},
			wantErr: false,
		},
		{
			name:         "show unknown terminal plugin returns error",
			args:         []string{"library", "show", "terminal", "plugin", "unknown-plugin-xyz"},
			wantErr:      true,
			wantNotFound: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantNotFound {
					output := buf.String()
					assert.Contains(t, output, "not found", "error should indicate resource not found")
				}
				return
			}

			require.NoError(t, err)
			output := buf.String()

			for _, want := range tt.wantContains {
				assert.Contains(t, output, want)
			}
		})
	}
}

// ========== Test Library Command Aliases ==========

// TestLibraryAliases tests various command and subcommand aliases
func TestLibraryAliases(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		// lib as alias for library
		{
			name:    "lib alias for library",
			args:    []string{"lib", "list", "plugins"},
			wantErr: false,
		},
		// ls as alias for list
		{
			name:    "ls alias for list",
			args:    []string{"library", "ls", "plugins"},
			wantErr: false,
		},
		// Combined aliases
		{
			name:    "lib ls combined aliases",
			args:    []string{"lib", "ls", "themes"},
			wantErr: false,
		},
		// Resource type aliases - nvim
		{
			name:    "np alias for plugins (nvim plugins)",
			args:    []string{"library", "list", "np"},
			wantErr: false,
		},
		{
			name:    "nt alias for themes (nvim themes)",
			args:    []string{"library", "list", "nt"},
			wantErr: false,
		},
		// Resource type aliases - terminal
		{
			name:    "tp alias for terminal prompts",
			args:    []string{"library", "list", "tp"},
			wantErr: false,
		},
		{
			name:    "tpl alias for terminal plugins",
			args:    []string{"library", "list", "tpl"},
			wantErr: false,
		},
		// Show command aliases
		{
			name:    "lib show with alias",
			args:    []string{"lib", "show", "plugin", "telescope"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}

// ========== Test Library Error Handling ==========

// TestLibraryErrorHandling tests error cases
func TestLibraryErrorHandling(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		wantErr      bool
		errContains  string
		helpExpected bool
	}{
		{
			name:         "library with no subcommand shows help",
			args:         []string{"library"},
			wantErr:      false,
			helpExpected: true,
		},
		{
			name:        "library list with unknown type",
			args:        []string{"library", "list", "unknown-type"},
			wantErr:     true,
			errContains: "unknown",
		},
		{
			name:        "library show with no resource type",
			args:        []string{"library", "show"},
			wantErr:     true,
			errContains: "required",
		},
		{
			name:        "library show with unknown resource type",
			args:        []string{"library", "show", "unknown-type", "something"},
			wantErr:     true,
			errContains: "unknown",
		},
		{
			name:        "library show plugin with no name",
			args:        []string{"library", "show", "plugin"},
			wantErr:     true,
			errContains: "required",
		},
		{
			name:        "invalid output format",
			args:        []string{"library", "list", "plugins", "-o", "xml"},
			wantErr:     true,
			errContains: "invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					output := buf.String()
					errStr := ""
					if err != nil {
						errStr = err.Error()
					}
					combined := output + errStr
					assert.Contains(t, combined, tt.errContains,
						"error/output should contain %q", tt.errContains)
				}
				return
			}

			assert.NoError(t, err)

			if tt.helpExpected {
				output := buf.String()
				assert.Contains(t, output, "Available Commands", "should show help text")
			}
		})
	}
}

// ========== Test Library Output Counts ==========

// TestLibraryOutputCounts verifies reasonable counts of library resources
func TestLibraryOutputCounts(t *testing.T) {
	tests := []struct {
		name         string
		args         []string
		minExpected  int // Minimum number of resources expected
		resourceType string
	}{
		{
			name:         "nvim plugins should have multiple entries",
			args:         []string{"library", "list", "plugins"},
			minExpected:  30, // We have ~38 plugins
			resourceType: "plugins",
		},
		{
			name:         "nvim themes should have multiple entries",
			args:         []string{"library", "list", "themes"},
			minExpected:  30, // We have ~34 themes
			resourceType: "themes",
		},
		{
			name:         "terminal prompts should have entries",
			args:         []string{"library", "list", "terminal", "prompts"},
			minExpected:  3, // We have ~5 starship prompts
			resourceType: "prompts",
		},
		{
			name:         "terminal plugins should have entries",
			args:         []string{"library", "list", "terminal", "plugins"},
			minExpected:  5, // We have ~8 terminal plugins
			resourceType: "plugins",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createTestLibraryCommand()
			cmd.SetArgs(tt.args)

			buf := new(bytes.Buffer)
			cmd.SetOut(buf)
			cmd.SetErr(buf)

			err := cmd.Execute()
			require.NoError(t, err)

			output := buf.String()
			// Count non-empty lines (rough heuristic)
			lines := strings.Split(output, "\n")
			nonEmptyLines := 0
			for _, line := range lines {
				if strings.TrimSpace(line) != "" {
					nonEmptyLines++
				}
			}

			// Should have at least minExpected + header lines
			assert.GreaterOrEqual(t, nonEmptyLines, tt.minExpected,
				"should have at least %d %s", tt.minExpected, tt.resourceType)
		})
	}
}

// ========== Test Helpers ==========

// createTestLibraryCommand creates a test library command structure
// This is a placeholder that will be replaced with actual implementation
func createTestLibraryCommand() *cobra.Command {
	// Root command
	rootCmd := &cobra.Command{
		Use: "dvm",
		Run: func(cmd *cobra.Command, args []string) {
			// No-op for testing
		},
	}

	// Library command with alias
	libraryCmd := &cobra.Command{
		Use:     "library",
		Aliases: []string{"lib"},
		Short:   "Browse library resources",
		Run: func(cmd *cobra.Command, args []string) {
			// Show help if no subcommand
			cmd.Help()
		},
	}

	// List command with alias
	listCmd := &cobra.Command{
		Use:     "list [resource-type]",
		Aliases: []string{"ls"},
		Short:   "List library resources",
		Run: func(cmd *cobra.Command, args []string) {
			// Implementation will be in actual code
			cmd.Println("library list not implemented")
		},
	}

	// Show command
	showCmd := &cobra.Command{
		Use:   "show [resource-type] [name]",
		Short: "Show library resource details",
		Run: func(cmd *cobra.Command, args []string) {
			// Implementation will be in actual code
			cmd.Println("library show not implemented")
		},
	}

	// Add output format flag
	listCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")
	showCmd.Flags().StringP("output", "o", "table", "Output format (table|yaml|json)")

	// Assemble command hierarchy
	libraryCmd.AddCommand(listCmd)
	libraryCmd.AddCommand(showCmd)
	rootCmd.AddCommand(libraryCmd)

	return rootCmd
}
