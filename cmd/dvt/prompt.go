package main

import (
	"fmt"

	"devopsmaestro/db"
	"devopsmaestro/pkg/palette"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/pkg/source"
	"devopsmaestro/pkg/terminalops/prompt"
	"devopsmaestro/render"

	"github.com/spf13/cobra"
)

// Global get command for kubectl-style usage
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get resources (kubectl-style)",
	Long:  `Get resources using kubectl-style patterns. Supports prompts and other resources.`,
}

// Get prompts subcommand
var getPromptsCmd = &cobra.Command{
	Use:   "prompts",
	Short: "List all terminal prompts",
	Long: `List all terminal prompts stored in the database.

Uses Resource/Handler pattern with database storage.

Examples:
  dvt get prompts                    # List all prompts
  dvt get prompts --type starship    # Filter by type
  dvt get prompts -o yaml            # Output as YAML
  dvt get prompts -o json            # Output as JSON`,
	RunE: promptResourceList,
}

// Set prompt command for dvt prompt set <name>
var promptSetCmd = &cobra.Command{
	Use:   "set <name>",
	Short: "Set the active prompt (kubectl-style)",
	Long: `Set the active terminal prompt and generate its config.

Uses Resource/Handler pattern with database storage.

Examples:
  dvt prompt set coolnight           # Set coolnight as active prompt
  dvt prompt set coolnight --force   # Set without confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: promptResourceSet,
}

// buildPromptResourceContext creates a resource.Context for prompt operations
func buildPromptResourceContext(cmd *cobra.Command) (resource.Context, error) {
	ctx := cmd.Context()
	dataStoreValue := ctx.Value("dataStore")
	if dataStoreValue == nil {
		return resource.Context{}, fmt.Errorf("dataStore not found in context")
	}

	dataStore, ok := dataStoreValue.(*db.DataStore)
	if !ok {
		return resource.Context{}, fmt.Errorf("dataStore has wrong type: %T", dataStoreValue)
	}

	return resource.Context{
		DataStore: *dataStore,
	}, nil
}

// promptResourceGet gets a specific prompt using Resource/Handler pattern
// This function is intended to replace the existing promptGetCmd.RunE function
func promptResourceGet(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Build resource context
	ctx, err := buildPromptResourceContext(cmd)
	if err != nil {
		return err
	}

	// Use resource system to get the prompt
	res, err := resource.Get(ctx, prompt.KindTerminalPrompt, name)
	if err != nil {
		return fmt.Errorf("prompt '%s' not found in database: %w", name, err)
	}

	promptRes, ok := res.(*handlers.TerminalPromptResource)
	if !ok {
		return fmt.Errorf("unexpected resource type: %T", res)
	}

	p := promptRes.Prompt()

	// Get output format from flags
	format, _ := cmd.Flags().GetString("output")

	// For JSON/YAML output
	if format == "json" || format == "yaml" || format == "" {
		// Default to YAML for single resource (kubectl pattern)
		if format == "" {
			format = "yaml"
		}
		return render.OutputWith(format, p.ToYAML(), render.Options{})
	}

	// For table output, show details
	enabled := "yes"
	if !p.Enabled {
		enabled = "no"
	}
	created := ""
	if p.CreatedAt != nil {
		created = p.CreatedAt.Format("2006-01-02 15:04:05")
	}

	kvData := render.NewOrderedKeyValueData(
		render.KeyValue{Key: "Name", Value: p.Name},
		render.KeyValue{Key: "Type", Value: string(p.Type)},
		render.KeyValue{Key: "Category", Value: p.Category},
		render.KeyValue{Key: "Description", Value: p.Description},
		render.KeyValue{Key: "Enabled", Value: enabled},
		render.KeyValue{Key: "Created", Value: created},
	)

	return render.OutputWith(format, kvData, render.Options{
		Type:  render.TypeKeyValue,
		Title: "Prompt Details",
	})
}

// promptResourceApply applies prompts from files using Resource/Handler pattern
// This function is intended to replace the existing promptApplyCmd.RunE function
func promptResourceApply(cmd *cobra.Command, args []string) error {
	files, _ := cmd.Flags().GetStringSlice("filename")

	if len(files) == 0 {
		return fmt.Errorf("must specify at least one file with -f flag")
	}

	// Build resource context
	ctx, err := buildPromptResourceContext(cmd)
	if err != nil {
		return err
	}

	for _, src := range files {
		if err := applyResourcePrompt(ctx, src); err != nil {
			return err
		}
	}

	return nil
}

// applyResourcePrompt applies a single prompt from a source
func applyResourcePrompt(ctx resource.Context, src string) error {
	// Resolve source and read data
	s := source.Resolve(src)
	data, displayName, err := s.Read()
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", src, err)
	}

	// Detect kind from YAML
	kind, err := resource.DetectKind(data)
	if err != nil {
		return fmt.Errorf("failed to detect resource kind from %s: %w", displayName, err)
	}

	// Verify it's a TerminalPrompt
	if kind != prompt.KindTerminalPrompt {
		return fmt.Errorf("expected TerminalPrompt, got %s in %s", kind, displayName)
	}

	// Get handler
	handler, err := resource.MustGetHandler(kind)
	if err != nil {
		return fmt.Errorf("no handler for %s: %w", kind, err)
	}

	// Apply the resource
	res, err := handler.Apply(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to apply %s from %s: %w", kind, displayName, err)
	}

	render.Success(fmt.Sprintf("TerminalPrompt '%s' applied (from %s)", res.GetName(), displayName))
	return nil
}

// promptResourceDelete deletes a prompt using Resource/Handler pattern
// This function is intended to replace the existing promptDeleteCmd.RunE function
func promptResourceDelete(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Build resource context
	ctx, err := buildPromptResourceContext(cmd)
	if err != nil {
		return err
	}

	// Check if prompt exists
	_, err = resource.Get(ctx, prompt.KindTerminalPrompt, name)
	if err != nil {
		return fmt.Errorf("prompt '%s' not found in database: %w", name, err)
	}

	// Check force flag
	force, _ := cmd.Flags().GetBool("force")

	// Confirm deletion unless --force is used
	if !force {
		fmt.Printf("Delete prompt '%s'? (y/N): ", name)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted")
			return nil
		}
	}

	// Get handler and delete
	handler, err := resource.MustGetHandler(prompt.KindTerminalPrompt)
	if err != nil {
		return fmt.Errorf("no handler for TerminalPrompt: %w", err)
	}

	if err := handler.Delete(ctx, name); err != nil {
		return fmt.Errorf("failed to delete prompt: %w", err)
	}

	render.Success(fmt.Sprintf("TerminalPrompt '%s' deleted", name))
	return nil
}

// promptResourceGenerate generates config file for a prompt using Resource/Handler pattern
// This function is intended to replace the existing promptGenerateCmd.RunE function
func promptResourceGenerate(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Build resource context
	ctx, err := buildPromptResourceContext(cmd)
	if err != nil {
		return err
	}

	// Get the prompt
	res, err := resource.Get(ctx, prompt.KindTerminalPrompt, name)
	if err != nil {
		return fmt.Errorf("prompt '%s' not found in database: %w", name, err)
	}

	promptRes, ok := res.(*handlers.TerminalPromptResource)
	if !ok {
		return fmt.Errorf("unexpected resource type: %T", res)
	}

	p := promptRes.Prompt()

	// Convert to PromptYAML for rendering
	promptYAML := p.ToYAML()

	// Create renderer
	renderer := prompt.NewStarshipRenderer()

	// Create a basic palette for variable resolution
	// TODO: This should use the active theme palette from theme resolver
	pal := &palette.Palette{
		Name: "default",
		Colors: map[string]string{
			// Basic color defaults for demo
			"bg":      "#1e1e2e",
			"fg":      "#cdd6f4",
			"primary": "#89b4fa",
			"red":     "#f38ba8",
			"green":   "#a6e3a1",
			"yellow":  "#f9e2af",
			"blue":    "#89b4fa",
			"purple":  "#cba6f7",
			"cyan":    "#94e2d5",
			"white":   "#bac2de",
		},
	}

	// Generate config
	config, err := renderer.Render(promptYAML, pal)
	if err != nil {
		return fmt.Errorf("failed to generate config for prompt '%s': %w", name, err)
	}

	// Handle output - just output to stdout for now
	fmt.Print(config)
	return nil
}

// promptResourceList lists all prompts using Resource/Handler pattern
// This is for the kubectl-style 'dvt get prompts' command
func promptResourceList(cmd *cobra.Command, args []string) error {
	// Build resource context
	ctx, err := buildPromptResourceContext(cmd)
	if err != nil {
		return err
	}

	// Get handler
	handler, err := resource.MustGetHandler(prompt.KindTerminalPrompt)
	if err != nil {
		return fmt.Errorf("no handler for TerminalPrompt: %w", err)
	}

	// List all prompts
	resources, err := handler.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list prompts: %w", err)
	}

	if len(resources) == 0 {
		fmt.Println("No prompts found")
		return nil
	}

	// Get output format and type filter from flags
	format, _ := cmd.Flags().GetString("output")
	typeFilter, _ := cmd.Flags().GetString("type")

	// Convert to prompts and filter by type if specified
	prompts := make([]*prompt.Prompt, 0, len(resources))
	for _, res := range resources {
		promptRes, ok := res.(*handlers.TerminalPromptResource)
		if !ok {
			continue
		}
		p := promptRes.Prompt()

		// Apply type filter
		if typeFilter != "" && string(p.Type) != typeFilter {
			continue
		}

		prompts = append(prompts, p)
	}

	if len(prompts) == 0 {
		if typeFilter != "" {
			fmt.Printf("No prompts found with type '%s'\n", typeFilter)
		} else {
			fmt.Println("No prompts found")
		}
		return nil
	}

	// For YAML/JSON output, output all prompts
	if format == "yaml" || format == "json" {
		// Convert to YAML documents
		yamlDocs := make([]*prompt.PromptYAML, len(prompts))
		for i, p := range prompts {
			yamlDocs[i] = p.ToYAML()
		}
		return render.OutputWith(format, yamlDocs, render.Options{})
	}

	// For table output (default)
	headers := []string{"NAME", "TYPE", "CATEGORY", "DESCRIPTION", "ENABLED"}
	var rows [][]string

	for _, p := range prompts {
		enabled := "yes"
		if !p.Enabled {
			enabled = "no"
		}

		description := p.Description
		if len(description) > 50 {
			description = description[:47] + "..."
		}

		rows = append(rows, []string{
			p.Name,
			string(p.Type),
			p.Category,
			description,
			enabled,
		})
	}

	tableData := render.TableData{
		Headers: headers,
		Rows:    rows,
	}
	return render.OutputWith(format, tableData, render.Options{
		Type:  render.TypeTable,
		Title: "Terminal Prompts",
	})
}

// promptResourceSet sets the active prompt and generates its config
// This function is for the kubectl-style 'dvt prompt set <name>' command
func promptResourceSet(cmd *cobra.Command, args []string) error {
	name := args[0]

	// Build resource context
	ctx, err := buildPromptResourceContext(cmd)
	if err != nil {
		return err
	}

	// Get the prompt to verify it exists
	res, err := resource.Get(ctx, prompt.KindTerminalPrompt, name)
	if err != nil {
		return fmt.Errorf("prompt '%s' not found in database: %w", name, err)
	}

	promptRes, ok := res.(*handlers.TerminalPromptResource)
	if !ok {
		return fmt.Errorf("unexpected resource type: %T", res)
	}

	p := promptRes.Prompt()

	// Check force flag
	force, _ := cmd.Flags().GetBool("force")

	// Confirm setting unless --force is used
	if !force {
		fmt.Printf("Set '%s' as the active prompt? (y/N): ", name)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Aborted")
			return nil
		}
	}

	// Generate config for the prompt
	promptYAML := p.ToYAML()
	renderer := prompt.NewStarshipRenderer()

	// Create a basic palette for variable resolution
	// TODO: This should use the active theme palette from theme resolver
	pal := &palette.Palette{
		Name: "default",
		Colors: map[string]string{
			// Basic color defaults for demo
			"bg":      "#1e1e2e",
			"fg":      "#cdd6f4",
			"primary": "#89b4fa",
			"red":     "#f38ba8",
			"green":   "#a6e3a1",
			"yellow":  "#f9e2af",
			"blue":    "#89b4fa",
			"purple":  "#cba6f7",
			"cyan":    "#94e2d5",
			"white":   "#bac2de",
		},
	}

	// Generate config
	_, err = renderer.Render(promptYAML, pal)
	if err != nil {
		return fmt.Errorf("failed to generate config for prompt '%s': %w", name, err)
	}

	// TODO: This should write to the appropriate config file (e.g., ~/.config/starship.toml)
	// For now, output instructions
	render.Success(fmt.Sprintf("Prompt '%s' set as active", name))
	fmt.Printf("\nTo use this prompt, save the config to ~/.config/starship.toml:\n")
	fmt.Printf("dvt prompt generate %s > ~/.config/starship.toml\n\n", name)
	fmt.Printf("Or view the generated config with:\n")
	fmt.Printf("dvt prompt generate %s\n", name)

	return nil
}

func init() {
	// Global get command structure
	getCmd.AddCommand(getPromptsCmd)

	// Flags for get prompts command
	getPromptsCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	getPromptsCmd.Flags().StringP("type", "", "", "Filter by prompt type (starship, powerlevel10k)")

	// Flags for prompt set command
	promptSetCmd.Flags().BoolP("force", "f", false, "Skip confirmation")

	// Register handlers to ensure TerminalPrompt is available
	handlers.RegisterAll()
}
