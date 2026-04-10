package cmd

import (
	"encoding/json"
	"fmt"

	"devopsmaestro/pkg/templates"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate resources and templates",
	Long:  `Generate resources, templates, and configuration for dvm.`,
}

var generateTemplateCmd = &cobra.Command{
	Use:   "template [kind]",
	Short: "Generate an annotated YAML template for a resource kind",
	Long: `Generate an annotated YAML template for a dvm resource kind.

The template includes all fields with placeholder values and helpful comments.
Use this as a starting point when creating new resources with 'dvm apply -f'.

Examples:
  dvm generate template ecosystem
  dvm generate template nvim-plugin
  dvm generate template workspace -o json
  dvm generate template --all`,
	ValidArgsFunction: completeTemplateKinds,
	RunE:              runGenerateTemplate,
}

func init() {
	generateTemplateCmd.Flags().StringP("output", "o", "yaml",
		"Output format: yaml, json")
	generateTemplateCmd.Flags().BoolP("all", "A", false,
		"Output all templates as a multi-document YAML")

	generateCmd.AddCommand(generateTemplateCmd)
	rootCmd.AddCommand(generateCmd)
}

// resetLocalFlags resets a command's local flags to their default values.
// Cobra does not reset local flag state between successive Execute() calls on a
// shared command tree, so without this, flags set in one invocation (e.g. --all)
// leak into the next. Called via defer in RunE to ensure cleanup even on error.
func resetLocalFlags(cmd *cobra.Command) {
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		_ = f.Value.Set(f.DefValue)
		f.Changed = false
	})
}

func runGenerateTemplate(cmd *cobra.Command, args []string) error {
	defer resetLocalFlags(cmd)

	allFlag, _ := cmd.Flags().GetBool("all")
	outputFmt, _ := cmd.Flags().GetString("output")

	store, err := templates.NewTemplateStore()
	if err != nil {
		return fmt.Errorf("failed to initialize template store: %w", err)
	}

	// Require either --all or a kind argument
	if !allFlag && len(args) == 0 {
		return fmt.Errorf("specify a resource kind or use --all")
	}

	if allFlag {
		data, err := store.GetAll()
		if err != nil {
			return fmt.Errorf("failed to get all templates: %w", err)
		}
		cmd.Print(string(data))
		return nil
	}

	kind := args[0]
	data, err := store.Get(kind)
	if err != nil {
		return err
	}

	switch outputFmt {
	case "json":
		var doc map[string]interface{}
		if err := yaml.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("failed to parse template YAML: %w", err)
		}
		jsonBytes, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to convert to JSON: %w", err)
		}
		cmd.Print(string(jsonBytes))
	default:
		cmd.Print(string(data))
	}

	return nil
}

// completeTemplateKinds provides shell completion for template kind names.
func completeTemplateKinds(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) > 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	store, err := templates.NewTemplateStore()
	if err != nil {
		return nil, cobra.ShellCompDirectiveError
	}

	return store.List(), cobra.ShellCompDirectiveNoFileComp
}
