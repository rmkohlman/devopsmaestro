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

// generate_template_test.go — Tests for the `dvm generate template` command
//
// Issue: #210
// TDD Phase 2 (Red) — These tests define the contract for the generate template
// command. They FAIL until @dvm-core implements the command in Phase 3.

// templateKindHeader is a minimal struct for validating apiVersion and kind.
type templateKindHeader struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// findSubCommand searches a cobra command's direct children for the given name.
func findSubCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, sub := range parent.Commands() {
		if sub.Name() == name {
			return sub
		}
	}
	return nil
}

// TestGenerateTemplateCmd_IsRegistered verifies `dvm generate template` exists.
func TestGenerateTemplateCmd_IsRegistered(t *testing.T) {
	generateCmd := findSubCommand(rootCmd, "generate")
	require.NotNil(t, generateCmd,
		"root command must have a 'generate' subcommand registered")

	templateCmd := findSubCommand(generateCmd, "template")
	require.NotNil(t, templateCmd,
		"'generate' must have a 'template' subcommand")
}

// TestGenerateTemplateCmd_Flags verifies required flags exist with correct types.
func TestGenerateTemplateCmd_Flags(t *testing.T) {
	generateCmd := findSubCommand(rootCmd, "generate")
	require.NotNil(t, generateCmd, "'generate' must be registered")

	templateCmd := findSubCommand(generateCmd, "template")
	require.NotNil(t, templateCmd, "'generate template' must be registered")

	tests := []struct {
		name     string
		wantType string
	}{
		{"output", "string"},
		{"all", "bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := templateCmd.Flags().Lookup(tt.name)
			require.NotNil(t, f, "--%s flag not found on generate template", tt.name)
			assert.Equal(t, tt.wantType, f.Value.Type(),
				"--%s should be type %q", tt.name, tt.wantType)
		})
	}
}

// TestGenerateTemplateCmd_SingleKind_OutputsValidYAML verifies that running
// `dvm generate template <kind>` outputs valid YAML with apiVersion and kind.
func TestGenerateTemplateCmd_SingleKind_OutputsValidYAML(t *testing.T) {
	tests := []struct {
		kind string
	}{
		{"ecosystem"},
		{"domain"},
		{"app"},
		{"workspace"},
		{"nvim-plugin"},
		{"terminal-prompt"},
		{"credential"},
		{"custom-resource-definition"},
	}

	for _, tt := range tests {
		t.Run(tt.kind, func(t *testing.T) {
			buf := new(bytes.Buffer)
			errBuf := new(bytes.Buffer)

			rootCmd.SetOut(buf)
			rootCmd.SetErr(errBuf)
			rootCmd.SetArgs([]string{"generate", "template", tt.kind})

			err := rootCmd.Execute()
			require.NoError(t, err, "generate template %s should not error", tt.kind)

			output := buf.String()
			assert.NotEmpty(t, output, "generate template %s should produce output", tt.kind)

			var header templateKindHeader
			parseErr := yaml.Unmarshal([]byte(output), &header)
			require.NoError(t, parseErr,
				"generate template %s output must be valid YAML", tt.kind)
			assert.NotEmpty(t, header.APIVersion,
				"template %s output must have apiVersion", tt.kind)
			assert.NotEmpty(t, header.Kind,
				"template %s output must have kind", tt.kind)

			rootCmd.SetOut(nil)
			rootCmd.SetErr(nil)
		})
	}
}

// TestGenerateTemplateCmd_AllFlag_OutputsMultiDocument verifies that
// `dvm generate template --all` outputs multi-document YAML with "---" separators.
func TestGenerateTemplateCmd_AllFlag_OutputsMultiDocument(t *testing.T) {
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	rootCmd.SetOut(buf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"generate", "template", "--all"})

	err := rootCmd.Execute()
	require.NoError(t, err, "generate template --all should not error")

	output := buf.String()
	assert.NotEmpty(t, output, "generate template --all should produce output")
	assert.Contains(t, output, "---",
		"generate template --all output must contain YAML document separators")

	// Verify multiple kinds appear in the combined output
	for _, kind := range []string{"ecosystem", "domain", "workspace", "nvim-plugin"} {
		assert.True(t, strings.Contains(output, kind),
			"--all output must contain %q", kind)
	}

	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
}

// TestGenerateTemplateCmd_JSONOutput verifies that `dvm generate template <kind> -o json`
// outputs valid JSON with apiVersion and kind fields.
func TestGenerateTemplateCmd_JSONOutput(t *testing.T) {
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	rootCmd.SetOut(buf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"generate", "template", "workspace", "-o", "json"})

	err := rootCmd.Execute()
	require.NoError(t, err, "generate template workspace -o json should not error")

	output := buf.String()
	assert.NotEmpty(t, output,
		"generate template workspace -o json should produce output")

	var doc map[string]interface{}
	parseErr := json.Unmarshal([]byte(output), &doc)
	require.NoError(t, parseErr, "generate template -o json output must be valid JSON")

	apiVersion, ok := doc["apiVersion"]
	require.True(t, ok, "JSON output must have 'apiVersion' field")
	assert.NotEmpty(t, apiVersion, "JSON output apiVersion must not be empty")

	kind, ok := doc["kind"]
	require.True(t, ok, "JSON output must have 'kind' field")
	assert.NotEmpty(t, kind, "JSON output kind must not be empty")

	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
}

// TestGenerateTemplateCmd_UnknownKind_ReturnsError verifies that providing an
// unrecognized kind name causes the command to return an error.
func TestGenerateTemplateCmd_UnknownKind_ReturnsError(t *testing.T) {
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	rootCmd.SetOut(buf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"generate", "template", "notakind"})

	err := rootCmd.Execute()
	assert.Error(t, err, "generate template notakind should return an error")

	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
}

// TestGenerateTemplateCmd_NoArgs_ReturnsError verifies that calling
// `dvm generate template` with no kind and no --all flag returns an error.
func TestGenerateTemplateCmd_NoArgs_ReturnsError(t *testing.T) {
	buf := new(bytes.Buffer)
	errBuf := new(bytes.Buffer)

	rootCmd.SetOut(buf)
	rootCmd.SetErr(errBuf)
	rootCmd.SetArgs([]string{"generate", "template"})

	err := rootCmd.Execute()
	assert.Error(t, err, "generate template with no args should return an error")

	rootCmd.SetOut(nil)
	rootCmd.SetErr(nil)
}
