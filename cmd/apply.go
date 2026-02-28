package cmd

import (
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/pkg/source"
	"devopsmaestro/render"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	// Register resource handlers
	handlers.RegisterAll()
}

// applyCmd is the root 'apply' command for kubectl-style resource application
// Usage: dvm apply -f <file> or dvm apply nvim plugin -f <file>
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a configuration from file",
	Long: `Apply a configuration to a resource from a YAML file, URL, directory, or stdin.

The -f flag accepts:
  - Local files: plugin.yaml, ./configs/theme.yaml
  - URLs: https://example.com/plugin.yaml
  - GitHub shorthand: github:user/repo/path/file.yaml
  - GitHub directories: github:user/repo/plugins/ (applies all YAML files)
  - Stdin: use '-' to read from stdin

Directory URLs (ending with / or no .yaml extension) will apply all YAML files
in that directory. Files are applied in alphabetical order.

The resource type is auto-detected from the 'kind' field in the YAML.
Supported kinds: NvimPlugin, NvimTheme, Workspace, TerminalPrompt

Secrets in YAML can be resolved from various providers:
  - macOS Keychain (default on macOS)
  - Environment variables (DVM_SECRET_<NAME> or GITHUB_TOKEN)
  
Use inline syntax: ${secret:name} or ${secret:name:provider}

Examples:
  # Apply single file
  dvm apply -f plugin.yaml
  
  # Apply multiple files
  dvm apply -f plugin.yaml -f theme.yaml
  
  # Apply from GitHub file
  dvm apply -f github:user/repo/plugins/telescope.yaml
  
  # Apply all plugins from GitHub directory
  dvm apply -f github:user/repo/plugins/
  
  # Apply from URL
  dvm apply -f https://example.com/workspace.yaml
  
  # Apply from stdin
  cat plugin.yaml | dvm apply -f -
  
  # Using secrets (token from keychain for private repos)
  dvm apply -f github:user/private-repo/config.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			// No -f flag provided, show help
			return cmd.Help()
		}

		return applyResources(cmd, files)
	},
}

// applyResources applies resources from the given sources using the unified pipeline.
func applyResources(cmd *cobra.Command, sources []string) error {
	// Build resource context
	ctx, err := buildResourceContext(cmd)
	if err != nil {
		return err
	}

	for _, src := range sources {
		// Check if this is a directory source
		if source.IsDirectory(src) && source.IsURL(src) {
			if err := applyDirectorySource(ctx, src); err != nil {
				return err
			}
		} else {
			// Single file apply (existing behavior)
			if err := applyResource(ctx, src); err != nil {
				return err
			}
		}
	}

	return nil
}

// applyDirectorySource handles applying all YAML files from a directory source.
func applyDirectorySource(ctx resource.Context, src string) error {
	// Create the directory source (currently only GitHub directories are supported)
	dirSource := source.NewGitHubDirectorySource(src)

	// List files using the DirectorySource interface
	files, err := dirSource.ListFiles()
	if err != nil {
		return fmt.Errorf("failed to list files from %s: %w", src, err)
	}

	if len(files) == 0 {
		render.Warning(fmt.Sprintf("No YAML files found in %s", src))
		return nil
	}

	render.Info(fmt.Sprintf("Found %d YAML files in %s", len(files), src))

	// Apply each file, showing progress
	var errors []error
	successCount := 0

	for i, file := range files {
		// Get source name for display
		sourceName := source.GetSourceName(file)
		render.Info(fmt.Sprintf("Applying %d/%d: %s...", i+1, len(files), sourceName))

		if err := applySourceFile(ctx, file, sourceName); err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", sourceName, err))
			render.Warning(fmt.Sprintf("  Failed: %v", err))
		} else {
			successCount++
		}
	}

	// Report summary
	if len(errors) > 0 {
		render.Warning(fmt.Sprintf("Applied %d/%d files successfully", successCount, len(files)))
		return fmt.Errorf("%d of %d files failed to apply", len(errors), len(files))
	}

	render.Success(fmt.Sprintf("Successfully applied all %d files from %s", len(files), src))
	return nil
}

// applySourceFile applies a single resource from a Source interface.
func applySourceFile(ctx resource.Context, src source.Source, sourceName string) error {
	// 1. Read data
	data, displayName, err := src.Read()
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", sourceName, err)
	}

	// 2. Detect kind from YAML
	kind, err := resource.DetectKind(data)
	if err != nil {
		return fmt.Errorf("failed to detect resource kind from %s: %w", displayName, err)
	}

	// 3. Get handler for this kind
	handler, err := resource.MustGetHandler(kind)
	if err != nil {
		return fmt.Errorf("unsupported resource kind '%s' in %s", kind, displayName)
	}

	// 4. Apply the resource
	res, err := handler.Apply(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to apply %s from %s: %w", kind, displayName, err)
	}

	render.Success(fmt.Sprintf("  %s '%s' applied", kind, res.GetName()))
	return nil
}

// applyResource applies a single resource from the given source.
func applyResource(ctx resource.Context, src string) error {
	// 1. Resolve source and read data
	s := source.Resolve(src)
	data, displayName, err := s.Read()
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", src, err)
	}

	// 2. Detect kind from YAML
	kind, err := resource.DetectKind(data)
	if err != nil {
		return fmt.Errorf("failed to detect resource kind from %s: %w", displayName, err)
	}

	// 3. Get handler for this kind
	handler, err := resource.MustGetHandler(kind)
	if err != nil {
		return fmt.Errorf("unsupported resource kind '%s' in %s", kind, displayName)
	}

	// 4. Check if resource exists (for messaging)
	// We need to parse first to get the name
	res, err := handler.Apply(ctx, data)
	if err != nil {
		return fmt.Errorf("failed to apply %s from %s: %w", kind, displayName, err)
	}

	render.Success(fmt.Sprintf("%s '%s' applied (from %s)", kind, res.GetName(), displayName))
	return nil
}

// buildResourceContext creates a resource.Context from the command.
func buildResourceContext(cmd *cobra.Command) (resource.Context, error) {
	datastore, err := getDataStore(cmd)
	if err != nil {
		return resource.Context{}, fmt.Errorf("failed to get datastore: %w", err)
	}

	return resource.Context{
		DataStore: datastore,
	}, nil
}

// applyNvimCmd is the 'nvim' subcommand under 'apply'
var applyNvimCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Apply nvim resources",
	Long: `Apply nvim-related resources (plugins, themes) from YAML files.

Examples:
  dvm apply nvim plugin -f telescope.yaml
  dvm apply nvim theme -f tokyonight.yaml`,
}

// applyNvimPluginCmd applies a nvim plugin from file or URL
// Usage: dvm apply nvim plugin -f <file|url>
var applyNvimPluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Apply a nvim plugin from file or URL",
	Long: `Apply nvim plugin definition(s) from a YAML file, URL, or directory.
If the plugin already exists, it will be updated.

The -f flag accepts local files, URLs, GitHub shorthand, or directories.
Directory URLs apply all .yaml files in that directory.

Examples:
  dvm apply nvim plugin -f telescope.yaml
  dvm apply nvim plugin -f plugin1.yaml -f plugin2.yaml
  dvm apply nvim plugin -f github:rmkohlman/nvim-yaml-plugins/plugins/
  dvm apply nvim plugin -f https://raw.githubusercontent.com/user/repo/main/plugin.yaml
  cat plugin.yaml | dvm apply nvim plugin -f -`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file or URL with -f flag")
		}

		// Use the unified apply pipeline
		return applyResources(cmd, files)
	},
}

// applyNvimThemeCmd applies a nvim theme from file or URL
// Usage: dvm apply nvim theme -f <file|url>
var applyNvimThemeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Apply a nvim theme from file or URL",
	Long: `Apply nvim theme definition(s) from a YAML file, URL, or directory.
If the theme already exists, it will be updated.

The -f flag accepts local files, URLs, GitHub shorthand, or directories.
Directory URLs apply all .yaml files in that directory.

Examples:
  dvm apply nvim theme -f tokyonight.yaml
  dvm apply nvim theme -f theme1.yaml -f theme2.yaml
  dvm apply nvim theme -f github:rmkohlman/nvim-yaml-plugins/themes/
  dvm apply nvim theme -f https://raw.githubusercontent.com/user/repo/main/theme.yaml
  cat theme.yaml | dvm apply nvim theme -f -`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file or URL with -f flag")
		}

		// Use the unified apply pipeline
		return applyResources(cmd, files)
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)

	// Add -f flag to root apply command
	applyCmd.Flags().StringSliceP("filename", "f", []string{}, "Resource YAML file(s) or URL(s) to apply (use '-' for stdin)")

	// Add nvim subcommand to apply
	applyCmd.AddCommand(applyNvimCmd)

	// Add plugin and theme under nvim
	applyNvimCmd.AddCommand(applyNvimPluginCmd)
	applyNvimCmd.AddCommand(applyNvimThemeCmd)

	// Add flags for subcommands (backward compatibility)
	applyNvimPluginCmd.Flags().StringSliceP("filename", "f", []string{}, "Plugin YAML file(s) or URL(s) to apply (use '-' for stdin)")
	applyNvimThemeCmd.Flags().StringSliceP("filename", "f", []string{}, "Theme YAML file(s) or URL(s) to apply (use '-' for stdin)")
}
