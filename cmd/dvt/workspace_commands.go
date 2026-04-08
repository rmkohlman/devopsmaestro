package main

import (
	"fmt"
	"os"

	"devopsmaestro/db"
	"devopsmaestro/pkg/terminalbridge"
	"devopsmaestro/pkg/terminalbridge/shellgen"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTerminal/terminalops/plugin"

	"github.com/spf13/cobra"
)

// =============================================================================
// SHELL GENERATE-WORKSPACE COMMAND
// =============================================================================

var shellGenerateWorkspaceCmd = &cobra.Command{
	Use:   "generate-workspace",
	Short: "Generate a workspace shell configuration file",
	Long: `Generate a workspace-scoped shell config file that composites:
  - Host shell config as base (if present)
  - Workspace-specific environment variables
  - Installed shell plugins (from database)
  - Shell configurations
  - Prompt initialization

Supports Bash, Zsh, and Fish shells. The shell type is auto-detected from
$SHELL by default, or can be specified with --shell.

Output goes to stdout by default. Use --output to write to a file.

Examples:
  dvt shell generate-workspace
  dvt shell generate-workspace --shell bash
  dvt shell generate-workspace --shell fish
  dvt shell generate-workspace --output .dvm/.zshrc.workspace
  dvt shell generate-workspace --workspace my-project
  dvt shell generate-workspace --no-host-config
  dvt shell generate-workspace --env KEY=VALUE --env APP_ENV=dev`,
	RunE: runShellGenerateWorkspace,
}

func runShellGenerateWorkspace(cmd *cobra.Command, args []string) error {
	// Determine shell type
	shellType, _ := cmd.Flags().GetString("shell")
	if shellType == "" {
		shellType = shellgen.DetectShell()
	}

	// Create the appropriate generator
	gen, err := shellgen.NewShellGenerator(shellType)
	if err != nil {
		return err
	}

	// Build ShellConfig from flags and database
	config := shellgen.ShellConfig{}

	// Workspace name
	workspace, _ := cmd.Flags().GetString("workspace")
	config.WorkspaceName = workspace

	// Host config — support both new and legacy flags
	noHostConfig, _ := cmd.Flags().GetBool("no-host-config")
	noHostZshrc, _ := cmd.Flags().GetBool("no-host-zshrc")
	if noHostConfig || noHostZshrc {
		f := false
		config.IncludeHostConfig = &f
	}

	hostPath, _ := cmd.Flags().GetString("host-config")
	if hostPath == "" {
		hostPath, _ = cmd.Flags().GetString("host-zshrc")
	}
	if hostPath != "" {
		config.HostShellConfigPath = hostPath
	}

	// Plugin dir
	pluginDir, _ := cmd.Flags().GetString("plugin-dir")
	config.PluginDir = pluginDir

	// Environment variables from flags
	envFlags, _ := cmd.Flags().GetStringSlice("env")
	if err := parseEnvFlags(envFlags, &config); err != nil {
		return err
	}

	// Prompt config
	promptConfig, _ := cmd.Flags().GetString("prompt-init")
	config.PromptConfig = promptConfig

	// Load plugins from database
	plugins, err := loadPluginsFromDB(cmd)
	if err != nil {
		// Non-fatal — continue without plugins
		if verbose {
			render.WarningfToStderr("Could not load plugins from database: %v", err)
		}
	} else {
		config.Plugins = plugins
	}

	// Load shell configs from file store
	shellStore := getShellStore()
	shells, err := shellStore.List()
	if err == nil && len(shells) > 0 {
		config.ShellConfigs = shells
	}

	// Generate
	output, err := gen.Generate(config)
	if err != nil {
		return fmt.Errorf("failed to generate workspace shell config: %w", err)
	}

	// Write output
	outputPath, _ := cmd.Flags().GetString("output")
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write to %s: %w", outputPath, err)
		}
		render.Successf("Generated %s workspace config at %s", shellType, outputPath)
		return nil
	}

	fmt.Print(output)
	return nil
}

// parseEnvFlags parses KEY=VALUE flags into the config.
func parseEnvFlags(flags []string, config *shellgen.ShellConfig) error {
	if len(flags) == 0 {
		return nil
	}
	if config.EnvVars == nil {
		config.EnvVars = make(map[string]string)
	}
	for _, f := range flags {
		idx := indexOf(f, '=')
		if idx == -1 {
			return fmt.Errorf("invalid env format %q — use KEY=VALUE", f)
		}
		config.EnvVars[f[:idx]] = f[idx+1:]
	}
	return nil
}

// indexOf returns the index of the first occurrence of c in s, or -1.
func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// loadPluginsFromDB loads installed plugins from the database.
func loadPluginsFromDB(cmd *cobra.Command) ([]*plugin.Plugin, error) {
	dataStoreInterface := cmd.Context().Value("dataStore")
	if dataStoreInterface == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	dataStore, ok := dataStoreInterface.(*db.DataStore)
	if !ok {
		return nil, fmt.Errorf("dataStore has wrong type")
	}

	store := terminalbridge.NewDBPluginStore(*dataStore)
	return store.List()
}

func init() {
	shellCmd.AddCommand(shellGenerateWorkspaceCmd)

	shellGenerateWorkspaceCmd.Flags().StringP("output", "o", "", "Write to file instead of stdout")
	shellGenerateWorkspaceCmd.Flags().StringP("workspace", "w", "", "Workspace name (for header comment)")
	shellGenerateWorkspaceCmd.Flags().StringP("shell", "s", "", "Target shell: bash, zsh, fish (default: auto-detect from $SHELL)")
	shellGenerateWorkspaceCmd.Flags().Bool("no-host-config", false, "Skip including host shell config")
	shellGenerateWorkspaceCmd.Flags().String("host-config", "", "Path to host shell config file")
	shellGenerateWorkspaceCmd.Flags().String("plugin-dir", "", "Plugin installation directory")
	shellGenerateWorkspaceCmd.Flags().StringSlice("env", nil, "Environment variables (KEY=VALUE)")
	shellGenerateWorkspaceCmd.Flags().String("prompt-init", "", "Prompt init command (e.g., eval \"$(starship init zsh)\")")

	// Backward-compatible hidden flags
	shellGenerateWorkspaceCmd.Flags().Bool("no-host-zshrc", false, "Skip including host ~/.zshrc (deprecated: use --no-host-config)")
	shellGenerateWorkspaceCmd.Flags().String("host-zshrc", "", "Path to host .zshrc (deprecated: use --host-config)")
	_ = shellGenerateWorkspaceCmd.Flags().MarkHidden("no-host-zshrc")
	_ = shellGenerateWorkspaceCmd.Flags().MarkHidden("host-zshrc")
}
