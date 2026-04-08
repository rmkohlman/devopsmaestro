package main

import (
	"fmt"
	"strings"

	"devopsmaestro/pkg/terminalbridge/toolconfig"

	"github.com/rmkohlman/MaestroSDK/colors"
	"github.com/rmkohlman/MaestroSDK/render"

	"github.com/spf13/cobra"
)

// =============================================================================
// TOOL-CONFIG COMMANDS
// =============================================================================

var toolConfigCmd = &cobra.Command{
	Use:   "tool-config",
	Short: "Generate theme-aware config for terminal tools (bat, delta, fzf, dircolors)",
	Long: `Generate configuration snippets for terminal tools using the active theme palette.

Supported tools: bat, delta, fzf, dircolors

Each tool gets colors mapped from the workspace theme, ensuring visual
consistency across your entire development environment.

Examples:
  dvt tool-config generate fzf           # Generate fzf color config
  dvt tool-config generate --all         # Generate config for all tools
  dvt tool-config generate delta --format raw  # Raw config without shell export
  dvt tool-config list                   # List supported tools`,
}

var toolConfigGenerateCmd = &cobra.Command{
	Use:   "generate [tool-name]",
	Short: "Generate config snippet for a terminal tool",
	Long: `Generate a config snippet for the specified tool using the active color palette.

By default, output is formatted as shell export statements. Use --format raw
to get just the config value.

Examples:
  dvt tool-config generate fzf
  dvt tool-config generate fzf --format raw
  dvt tool-config generate --all
  dvt tool-config generate delta --format env`,
	Args: cobra.MaximumNArgs(1),
	RunE: runToolConfigGenerate,
}

var toolConfigListCmd = &cobra.Command{
	Use:   "list",
	Short: "List supported tool config generators",
	Run: func(cmd *cobra.Command, args []string) {
		render.Info("Supported tool config generators:")
		render.Blank()
		for _, gen := range toolconfig.AllGenerators() {
			render.Plainf("  %-12s %s", gen.ToolName(), gen.Description())
		}
	},
}

func runToolConfigGenerate(cmd *cobra.Command, args []string) error {
	allTools, _ := cmd.Flags().GetBool("all")
	format, _ := cmd.Flags().GetString("format")

	if !allTools && len(args) == 0 {
		return fmt.Errorf("specify a tool name or use --all\navailable: %s",
			strings.Join(toolconfig.AvailableTools(), ", "))
	}

	// Get palette from the active color provider in context
	provider := colors.FromContextOrDefault(cmd.Context())
	pal := colors.ToPalette(provider)

	// Determine which generators to run
	var generators []toolconfig.ToolConfigGenerator
	if allTools {
		generators = toolconfig.AllGenerators()
	} else {
		gen, err := toolconfig.NewToolConfigGenerator(args[0])
		if err != nil {
			return err
		}
		generators = []toolconfig.ToolConfigGenerator{gen}
	}

	// Generate and output for each tool
	for i, gen := range generators {
		if i > 0 && allTools {
			fmt.Println() // blank line between tools
		}

		output, err := gen.Generate(pal)
		if err != nil {
			render.WarningfToStderr("Failed to generate %s config: %v",
				gen.ToolName(), err)
			continue
		}

		switch toolconfig.OutputFormat(format) {
		case toolconfig.FormatRaw:
			fmt.Print(output)
		case toolconfig.FormatEnv:
			printEnvFormat(gen.ToolName(), output)
		default:
			printEnvFormat(gen.ToolName(), output)
		}
	}

	return nil
}

// printEnvFormat wraps tool output in shell export statements.
func printEnvFormat(toolName, output string) {
	switch toolName {
	case "bat":
		// bat uses a config file content, export BAT_THEME
		fmt.Printf("export BAT_THEME=\"ansi\"\n")
	case "fzf":
		// Extract the --color= line from the output
		for _, line := range strings.Split(output, "\n") {
			if strings.HasPrefix(line, "--color=") {
				fmt.Printf("export FZF_DEFAULT_OPTS=\"%s\"\n", line)
				break
			}
		}
	case "dircolors":
		// Extract the LS_COLORS value (non-comment line)
		for _, line := range strings.Split(output, "\n") {
			if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
				fmt.Printf("export LS_COLORS=\"%s\"\n", line)
				break
			}
		}
	case "delta":
		// Delta config goes in gitconfig, output as-is with comment
		fmt.Printf("# Add to ~/.config/git/config or ~/.gitconfig:\n")
		fmt.Print(output)
	default:
		fmt.Print(output)
	}
}

func init() {
	toolConfigCmd.AddCommand(toolConfigGenerateCmd)
	toolConfigCmd.AddCommand(toolConfigListCmd)

	toolConfigGenerateCmd.Flags().Bool("all", false, "Generate config for all tools")
	toolConfigGenerateCmd.Flags().String("format", "env",
		"Output format: env (shell exports) or raw (config value only)")
}
