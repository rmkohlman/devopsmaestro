package main

import (
	"context"
	"devopsmaestro/db"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"devopsmaestro/pkg/terminalbridge"
	"github.com/rmkohlman/MaestroSDK/colors"
	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/rmkohlman/MaestroTerminal/terminalops/plugin"
	pluginlibrary "github.com/rmkohlman/MaestroTerminal/terminalops/plugin/library"
	"github.com/rmkohlman/MaestroTerminal/terminalops/profile"
	"github.com/rmkohlman/MaestroTerminal/terminalops/prompt"
	promptlibrary "github.com/rmkohlman/MaestroTerminal/terminalops/prompt/library"
	"github.com/rmkohlman/MaestroTerminal/terminalops/shell"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

var (
	// Global flags
	configDir string
	outputFmt string
	verbose   bool
	logFile   string
	noColor   bool
)

// getMigrationsFS creates a filesystem for migrations.
// Uses embedded migrations that are built into the dvt binary at compile time.
func getMigrationsFS() (fs.FS, error) {
	return GetEmbeddedMigrationsFS()
}

// shouldSkipAutoMigration determines if auto-migration should be skipped for this command.
func shouldSkipAutoMigration(cmd *cobra.Command) bool {
	commandName := cmd.Name()
	switch commandName {
	case "completion", "version", "help":
		return true
	}
	return false
}

// errSilent is returned by commands that have already displayed their error
// via render.Error(). It causes Cobra to set exit code 1 without double-printing.
var errSilent = fmt.Errorf("")

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "dvt",
	Short: "TerminalOps - DevOps-style terminal configuration management",
	Long: `dvt (TerminalOps) is a CLI tool for managing terminal configuration using a DevOps-style
YAML configuration approach. It provides:

  - YAML-based prompt definitions (Starship/P10k)
  - Built-in library of curated prompts and plugins
  - Config file generation (starship.toml, .zshrc)
  - Profile system to aggregate prompts, plugins, and shell settings

Quick Start:
  dvt prompt library list           # See available prompts
  dvt prompt library install starship-default
  dvt plugin library list           # See available plugins
  dvt profile generate              # Generate config files

Configuration is stored in ~/.dvt/ by default.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// setupDatabaseConfig configures database settings for dvt
func setupDatabaseConfig() error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Failed to get home directory", "error", err)
		render.ErrorfToStderr("Failed to get home directory: %v", err)
		return errSilent
	}

	// Set config path to ~/.devopsmaestro (same as dvm for shared database)
	configPath := paths.New(homeDir).Root()
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()

	// Try to read config, but don't fail if it doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Error("Failed to read config", "error", err)
			render.ErrorfToStderr("Failed to read config: %v", err)
			return errSilent
		}
		// Config not found is OK - use defaults
	}

	// Set default values if config is not found (use same database as dvm)
	if viper.GetString("database.type") == "" {
		viper.Set("database.type", "sqlite")
		viper.Set("database.path", "~/"+paths.DVMDirName+"/"+paths.DatabaseFile)
		viper.Set("store", "sql")
	}
	return nil
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configDir, "config", "", "Config directory (default: ~/.dvt)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Write logs to file (JSON format)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Initialize logging, color provider, and database before any command runs
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		initLogging()

		// Initialize ColorProvider - respect --no-color flag and NO_COLOR env var
		ctx, err := colors.InitColorProviderForCommand(
			cmd.Context(),
			nil, // dvt does not have its own theme store
			noColor,
		)
		if err != nil {
			slog.Warn("using default colors", "error", err)
		}
		cmd.SetContext(ctx)

		// Check if this is a command that doesn't need database
		skipDB := false
		commandName := cmd.Name()
		switch commandName {
		case "completion", "version", "help":
			skipDB = true
		}

		if !skipDB {
			// Setup database configuration
			if err := setupDatabaseConfig(); err != nil {
				return err
			}

			// Create DataStore instance
			dataStore, err := db.CreateDataStore()
			if err != nil {
				slog.Error("Failed to initialize database", "error", err)
				render.ErrorfToStderr("Failed to initialize database: %v", err)
				return errSilent
			}

			// Set the dataStore in context for resource operations
			ctx := context.WithValue(cmd.Context(), "dataStore", &dataStore)
			cmd.SetContext(ctx)

			// Auto-migrate database if needed (skip for commands that don't need DB)
			if shouldSkipAutoMigration(cmd) {
				return nil
			}

			// Auto-migrate database if needed
			driver := dataStore.Driver()
			if driver != nil {
				// Get migrations FS - for dvt we try to find migrations on disk
				migrationsFS, err := getMigrationsFS()
				if err != nil {
					// If migrations not found, warn but continue
					// This allows dvt to work even when migrations are not available
					if verbose {
						slog.Warn("migrations not available for auto-migration in dvt", "error", err)
						render.WarningfToStderr("Migrations not found, skipping auto-migration: %v", err)
					}
					return nil
				}

				// Use version-based auto-migration
				migrationsApplied, err := db.CheckVersionBasedAutoMigration(driver, migrationsFS, Version, verbose)
				if err != nil {
					slog.Error("auto-migration failed", "error", err)
					render.ErrorfToStderr("Failed to apply database migrations: %v", err)
					render.InfoToStderr("Please run 'dvm admin migrate' to fix migration issues.")
					return errSilent
				}
				if migrationsApplied && verbose {
					slog.Info("database migrations applied successfully")
				}
			}
		}
		return nil
	}

	// Add all commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(promptCmd)
	rootCmd.AddCommand(pluginCmd)
	rootCmd.AddCommand(packageCmd)
	rootCmd.AddCommand(shellCmd)
	rootCmd.AddCommand(profileCmd)
	rootCmd.AddCommand(emulatorCmd)
	rootCmd.AddCommand(weztermCmd)
	rootCmd.AddCommand(completionCmd)
}

// initLogging configures the global slog logger based on flags.
func initLogging() {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler

	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			render.WarningfToStderr("could not open log file %s: %v", logFile, err)
			handler = slog.NewTextHandler(os.Stderr, opts)
		} else {
			handler = slog.NewJSONHandler(f, opts)
		}
	} else if verbose {
		handler = slog.NewTextHandler(os.Stderr, opts)
	} else {
		handler = slog.NewTextHandler(io.Discard, opts)
	}

	slog.SetDefault(slog.New(handler))
}

// =============================================================================
// VERSION COMMAND
// =============================================================================

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		short, _ := cmd.Flags().GetBool("short")
		if short {
			fmt.Println(Version)
			return
		}
		fmt.Printf("dvt (TerminalOps) %s\n", Version)
		fmt.Printf("  Build time: %s\n", BuildTime)
		fmt.Printf("  Commit:     %s\n", Commit)
	},
}

func init() {
	versionCmd.Flags().Bool("short", false, "Print only version number")
}

// =============================================================================
// INIT COMMAND
// =============================================================================

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize dvt configuration",
	Long: `Initialize the dvt configuration directory.

This creates:
  ~/.dvt/
  ~/.dvt/prompts/      # Prompt YAML storage
  ~/.dvt/plugins/      # Plugin YAML storage
  ~/.dvt/shells/       # Shell YAML storage
  ~/.dvt/profiles/     # Profile YAML storage

You can specify a custom directory with --config or DVT_CONFIG_DIR.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getConfigDir()

		// Create directories
		dirs := []string{
			filepath.Join(dir, "prompts"),
			filepath.Join(dir, "plugins"),
			filepath.Join(dir, "shells"),
			filepath.Join(dir, "profiles"),
		}

		for _, d := range dirs {
			if err := os.MkdirAll(d, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", d, err)
			}
		}

		render.Successf("Initialized dvt at %s", dir)
		return nil
	},
}

// =============================================================================
// PROMPT COMMANDS
// =============================================================================

var promptCmd = &cobra.Command{
	Use:     "prompt",
	Aliases: []string{"pr"},
	Short:   "Manage terminal prompts (Starship, P10k)",
	Long: `Manage terminal prompt configurations.

Prompts define how your shell prompt looks using tools like Starship or P10k.
Use the library to get started with pre-configured prompts, then customize as needed.`,
}

var promptLibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Browse and import prompts from the library",
}

var promptLibraryListCmd = &cobra.Command{
	Use:   "get",
	Short: "List available prompts in the library",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := promptlibrary.NewPromptLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		prompts := lib.List()

		// Filter by category if specified
		category, _ := cmd.Flags().GetString("category")
		if category != "" {
			prompts = lib.ListByCategory(category)
		}

		if len(prompts) == 0 {
			render.Info("No prompts found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPrompts(prompts, format)
	},
}

var promptLibraryShowCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Show details of a library prompt",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lib, err := promptlibrary.NewPromptLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		p, err := lib.Get(name)
		if err != nil {
			return fmt.Errorf("prompt not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPrompt(p, format)
	},
}

var promptLibraryInstallCmd = &cobra.Command{
	Use:   "import <name>...",
	Short: "Import prompts from library to local store",
	Long: `Copy prompt definitions from the built-in library to your local store.
You can then customize them or use them directly.

Examples:
  dvt prompt library import starship-default
  dvt prompt library import starship-minimal starship-powerline
  dvt prompt library import --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if !all && len(args) == 0 {
			return fmt.Errorf("specify prompt names or use --all")
		}

		lib, err := promptlibrary.NewPromptLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		fileStore := getPromptStore()

		// Also write to database so dvt prompt get/generate/set/delete work
		var dbStore *terminalbridge.DBPromptStore
		if ds := cmd.Context().Value("dataStore"); ds != nil {
			if dataStore, ok := ds.(*db.DataStore); ok {
				dbStore = terminalbridge.NewDBPromptStore(*dataStore)
			}
		}

		var prompts []*prompt.Prompt
		if all {
			prompts = lib.List()
		} else {
			for _, name := range args {
				p, err := lib.Get(name)
				if err != nil {
					render.WarningfToStderr("prompt not found in library: %s", name)
					continue
				}
				prompts = append(prompts, p)
			}
		}

		for _, p := range prompts {
			if err := fileStore.Save(p); err != nil {
				render.WarningfToStderr("failed to install %s: %v", p.Name, err)
				continue
			}
			// Sync to database
			if dbStore != nil {
				if err := dbStore.Upsert(p); err != nil {
					render.WarningfToStderr("installed %s to file store but failed to sync to database: %v", p.Name, err)
				}
			}
			render.Successf("Installed %s", p.Name)
		}

		return nil
	},
}

var promptLibraryCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List prompt categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := promptlibrary.NewPromptLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		categories := lib.Categories()
		render.Infof("Categories (%d):", len(categories))
		for _, c := range categories {
			prompts := lib.ListByCategory(c)
			render.Plainf("  %-15s (%d prompts)", c, len(prompts))
		}
		return nil
	},
}

var promptGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get prompt definition(s)",
	Long: `Get terminal prompts stored in the database.

With no arguments, lists all installed prompts.
With a name argument, gets a specific prompt definition.

Uses Resource/Handler pattern with database storage.

Examples:
  dvt prompt get                     # List all installed prompts
  dvt prompt get coolnight           # Get prompt as YAML
  dvt prompt get coolnight -o json   # Get prompt as JSON
  dvt prompt get coolnight -o table  # Get prompt as table`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// List mode
			store := getPromptStore()
			prompts, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list prompts: %w", err)
			}

			if len(prompts) == 0 {
				render.Info("No prompts installed")
				render.Info("Use 'dvt prompt library get' to see available prompts")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			return outputPrompts(prompts, format)
		}
		// Single get mode - delegate to resource handler
		return promptResourceGet(cmd, args)
	},
}

var promptApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a prompt definition from file (kubectl-style)",
	Long: `Apply a terminal prompt configuration from a YAML file using Resource/Handler pattern.

Uses database storage and supports theme variable resolution.

Examples:
  dvt prompt apply -f my-prompt.yaml
  dvt prompt apply -f -               # Read from stdin
  dvt prompt apply -f prompt1.yaml -f prompt2.yaml  # Apply multiple`,
	RunE: promptResourceApply,
}

var promptDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a prompt (kubectl-style)",
	Long: `Delete a terminal prompt from the database using Resource/Handler pattern.

Requires confirmation unless --force is used.

Examples:
  dvt prompt delete coolnight        # Delete with confirmation
  dvt prompt delete coolnight --force  # Delete without confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: promptResourceDelete,
}

var promptGenerateCmd = &cobra.Command{
	Use:   "generate <name>",
	Short: "Generate config file for a prompt (kubectl-style)",
	Long: `Generate the configuration file for a terminal prompt stored in the database.

Uses Resource/Handler pattern with theme variable resolution.
For Starship prompts, this outputs starship.toml content.

Examples:
  dvt prompt generate coolnight                    # Output to stdout
  dvt prompt generate coolnight > ~/.config/starship.toml  # Save to file`,
	Args: cobra.ExactArgs(1),
	RunE: promptResourceGenerate,
}

func init() {
	// Prompt subcommands
	promptCmd.AddCommand(promptLibraryCmd)
	promptCmd.AddCommand(promptGetCmd)
	promptCmd.AddCommand(promptApplyCmd)
	promptCmd.AddCommand(promptDeleteCmd)
	promptCmd.AddCommand(promptGenerateCmd)
	promptCmd.AddCommand(promptSetCmd)

	// Prompt library subcommands
	promptLibraryCmd.AddCommand(promptLibraryListCmd)
	promptLibraryCmd.AddCommand(promptLibraryShowCmd)
	promptLibraryCmd.AddCommand(promptLibraryInstallCmd)
	promptLibraryCmd.AddCommand(promptLibraryCategoriesCmd)

	// Flags
	promptLibraryListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	promptLibraryListCmd.Flags().StringP("category", "c", "", "Filter by category")
	promptLibraryShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	promptLibraryInstallCmd.Flags().Bool("all", false, "Import all prompts from library")
	promptGetCmd.Flags().StringP("output", "o", "yaml", "Output format: table, yaml, json")
	promptApplyCmd.Flags().StringSliceP("filename", "f", nil, "Prompt YAML file(s)")
	promptDeleteCmd.Flags().Bool("force", false, "Skip confirmation")

	// Hidden backward-compat aliases for deprecated verbs in prompt (after flags)
	promptLibraryCmd.AddCommand(hiddenAlias("list", promptLibraryListCmd))
	promptLibraryCmd.AddCommand(hiddenAlias("show", promptLibraryShowCmd))
	promptLibraryCmd.AddCommand(hiddenAlias("install", promptLibraryInstallCmd))
	promptCmd.AddCommand(hiddenAlias("list", promptGetCmd))
}

// =============================================================================
// PLUGIN COMMANDS
// =============================================================================

var pluginCmd = &cobra.Command{
	Use:     "plugin",
	Aliases: []string{"pl"},
	Short:   "Manage shell plugins (zsh-autosuggestions, etc.)",
	Long: `Manage shell plugin configurations.

Plugins enhance your shell with features like autosuggestions, syntax highlighting,
and fuzzy finding. The library provides pre-configured plugins that work well together.`,
}

var pluginLibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Browse and import plugins from the library",
}

var pluginLibraryListCmd = &cobra.Command{
	Use:   "get",
	Short: "List available plugins in the library",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := pluginlibrary.NewPluginLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		plugins := lib.List()

		// Filter by category if specified
		category, _ := cmd.Flags().GetString("category")
		if category != "" {
			plugins = lib.ListByCategory(category)
		}

		if len(plugins) == 0 {
			render.Info("No plugins found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugins(plugins, format)
	},
}

var pluginLibraryShowCmd = &cobra.Command{
	Use:   "describe <name>",
	Short: "Show details of a library plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lib, err := pluginlibrary.NewPluginLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		p, err := lib.Get(name)
		if err != nil {
			return fmt.Errorf("plugin not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugin(p, format)
	},
}

var pluginLibraryInstallCmd = &cobra.Command{
	Use:   "import <name>...",
	Short: "Import plugins from library to local store",
	Long: `Copy plugin definitions from the built-in library to your local store.

Examples:
  dvt plugin library import zsh-autosuggestions
  dvt plugin library import zsh-autosuggestions zsh-syntax-highlighting
  dvt plugin library import --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if !all && len(args) == 0 {
			return fmt.Errorf("specify plugin names or use --all")
		}

		lib, err := pluginlibrary.NewPluginLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		store, err := getPluginStore(cmd)
		if err != nil {
			return err
		}

		var plugins []*plugin.Plugin
		if all {
			plugins = lib.List()
		} else {
			for _, name := range args {
				p, err := lib.Get(name)
				if err != nil {
					render.WarningfToStderr("plugin not found in library: %s", name)
					continue
				}
				plugins = append(plugins, p)
			}
		}

		for _, p := range plugins {
			if err := store.Upsert(p); err != nil {
				render.WarningfToStderr("failed to install %s: %v", p.Name, err)
				continue
			}
			render.Successf("Installed %s", p.Name)
		}

		return nil
	},
}

var pluginLibraryCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List plugin categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := pluginlibrary.NewPluginLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		categories := lib.Categories()
		render.Infof("Categories (%d):", len(categories))
		for _, c := range categories {
			plugins := lib.ListByCategory(c)
			render.Plainf("  %-20s (%d plugins)", c, len(plugins))
		}
		return nil
	},
}

var pluginGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get plugin definition(s)",
	Long: `Get shell plugin definitions.

With no arguments, lists all installed plugins.
With a name argument, gets a specific plugin definition.

Examples:
  dvt plugin get                        # List installed plugins
  dvt plugin get zsh-autosuggestions    # Get specific plugin`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// List mode
			store, err := getPluginStore(cmd)
			if err != nil {
				return err
			}
			plugins, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list plugins: %w", err)
			}

			if len(plugins) == 0 {
				render.Info("No plugins installed")
				render.Info("Use 'dvt plugin library get' to see available plugins")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			return outputPlugins(plugins, format)
		}
		// Single get mode
		name := args[0]
		store, err := getPluginStore(cmd)
		if err != nil {
			return err
		}

		p, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("plugin not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugin(p, format)
	},
}

var pluginGenerateCmd = &cobra.Command{
	Use:   "generate [name...]",
	Short: "Generate .zshrc plugin section (stdout)",
	Long: `Generate shell configuration for installing and loading plugins.

If no names specified, generates for all installed plugins.
Supports different plugin managers: zinit, oh-my-zsh, antigen, sheldon, manual.

Examples:
  dvt plugin generate                          # All installed plugins
  dvt plugin generate zsh-autosuggestions      # Specific plugin
  dvt plugin generate --manager zinit          # Use zinit format`,
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getPluginStore(cmd)
		if err != nil {
			return err
		}
		_ = cmd.Flags().Lookup("manager") // manager flag not used with current API

		var plugins []*plugin.Plugin

		if len(args) > 0 {
			// Specific plugins
			for _, name := range args {
				p, err := store.Get(name)
				if err != nil {
					// Try library
					lib, _ := pluginlibrary.NewPluginLibrary()
					if libPlugin, libErr := lib.Get(name); libErr == nil {
						p = libPlugin
					} else {
						render.WarningfToStderr("plugin not found: %s", name)
						continue
					}
				}
				plugins = append(plugins, p)
			}
		} else {
			// All installed plugins
			plugins, err = store.List()
			if err != nil {
				return fmt.Errorf("failed to list plugins: %w", err)
			}
		}

		if len(plugins) == 0 {
			fmt.Println("# No plugins to generate")
			return nil
		}

		gen := plugin.NewZshGenerator("")
		output, err := gen.Generate(plugins)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		fmt.Print(output)
		return nil
	},
}

func init() {
	// Plugin subcommands
	pluginCmd.AddCommand(pluginLibraryCmd)
	pluginCmd.AddCommand(pluginGetCmd)
	pluginCmd.AddCommand(pluginGenerateCmd)

	// Plugin library subcommands
	pluginLibraryCmd.AddCommand(pluginLibraryListCmd)
	pluginLibraryCmd.AddCommand(pluginLibraryShowCmd)
	pluginLibraryCmd.AddCommand(pluginLibraryInstallCmd)
	pluginLibraryCmd.AddCommand(pluginLibraryCategoriesCmd)

	// Flags
	pluginLibraryListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	pluginLibraryListCmd.Flags().StringP("category", "c", "", "Filter by category")
	pluginLibraryShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	pluginLibraryInstallCmd.Flags().Bool("all", false, "Import all plugins from library")
	pluginGetCmd.Flags().StringP("output", "o", "yaml", "Output format: table, yaml, json")
	pluginGenerateCmd.Flags().StringP("manager", "m", "manual", "Plugin manager: zinit, oh-my-zsh, antigen, sheldon, manual")

	// Hidden backward-compat aliases for deprecated verbs in plugin (after flags)
	pluginLibraryCmd.AddCommand(hiddenAlias("list", pluginLibraryListCmd))
	pluginLibraryCmd.AddCommand(hiddenAlias("show", pluginLibraryShowCmd))
	pluginLibraryCmd.AddCommand(hiddenAlias("install", pluginLibraryInstallCmd))
	pluginCmd.AddCommand(hiddenAlias("list", pluginGetCmd))
}

// =============================================================================
// SHELL COMMANDS
// =============================================================================

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "Manage shell configurations (aliases, env vars, functions)",
	Long: `Manage shell configuration like aliases, environment variables, and functions.

Shell configs define the non-prompt parts of your shell setup. Use profiles
to combine shell configs with prompts and plugins.`,
}

var shellApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a shell configuration from file",
	Long: `Apply a shell configuration from a YAML file.

Examples:
  dvt shell apply -f my-shell.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file with -f flag")
		}

		store := getShellStore()

		for _, file := range files {
			var data []byte
			var err error
			var source string

			if file == "-" {
				data, err = io.ReadAll(os.Stdin)
				source = "stdin"
			} else {
				data, err = os.ReadFile(file)
				source = file
			}
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", source, err)
			}

			s, err := shell.Parse(data)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", source, err)
			}

			existing, _ := store.Get(s.Name)
			action := "created"
			if existing != nil {
				action = "updated"
			}

			if err := store.Save(s); err != nil {
				return fmt.Errorf("failed to save shell config: %w", err)
			}

			render.Successf("Shell config '%s' %s (from %s)", s.Name, action, source)
		}

		return nil
	},
}

var shellGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get shell configuration(s)",
	Long: `Get shell configurations.

With no arguments, lists all installed shell configurations.
With a name argument, gets a specific shell configuration.

Examples:
  dvt shell get              # List installed shell configs
  dvt shell get my-shell     # Get specific shell config`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// List mode
			store := getShellStore()
			shells, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list shell configs: %w", err)
			}

			if len(shells) == 0 {
				render.Info("No shell configs installed")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			return outputShells(shells, format)
		}
		// Single get mode
		name := args[0]
		store := getShellStore()

		s, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("shell config not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputShell(s, format)
	},
}

var shellGenerateCmd = &cobra.Command{
	Use:   "generate <name>",
	Short: "Generate shell config section (stdout)",
	Long: `Generate the shell configuration section for a shell config.

This outputs aliases, environment variables, and functions for your .zshrc/.bashrc.

Examples:
  dvt shell generate my-shell
  dvt shell generate my-shell >> ~/.zshrc`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store := getShellStore()

		s, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("shell config not found: %s", name)
		}

		gen := shell.NewGenerator()
		output, err := gen.Generate(s)
		if err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		fmt.Print(output)
		return nil
	},
}

func init() {
	// Shell subcommands
	shellCmd.AddCommand(shellApplyCmd)
	shellCmd.AddCommand(shellGetCmd)
	shellCmd.AddCommand(shellGenerateCmd)

	// Flags
	shellApplyCmd.Flags().StringSliceP("filename", "f", nil, "Shell YAML file(s)")
	shellGetCmd.Flags().StringP("output", "o", "yaml", "Output format: table, yaml, json")

	// Hidden backward-compat alias for deprecated verb in shell (after flags)
	shellCmd.AddCommand(hiddenAlias("list", shellGetCmd))
}

// =============================================================================
// PROFILE COMMANDS
// =============================================================================

var profileCmd = &cobra.Command{
	Use:   "profile",
	Short: "Manage terminal profiles (combines prompt, plugins, shell)",
	Long: `Manage terminal profiles that combine prompts, plugins, and shell settings.

Profiles are the recommended way to manage your terminal configuration. They:
  - Reference a prompt (Starship/P10k)
  - Include multiple plugins (autosuggestions, syntax-highlighting, etc.)
  - Include shell settings (aliases, env vars, functions)

Quick start with presets:
  dvt profile preset get        # See available presets
  dvt profile preset import default
  dvt profile generate          # Generate all config files`,
}

var profilePresetCmd = &cobra.Command{
	Use:   "preset",
	Short: "Manage profile presets",
}

var profilePresetListCmd = &cobra.Command{
	Use:   "get",
	Short: "List available profile presets",
	RunE: func(cmd *cobra.Command, args []string) error {
		presets := []struct {
			name        string
			description string
		}{
			{"default", "Balanced setup with Starship, autosuggestions, and syntax-highlighting"},
			{"minimal", "Lightweight setup with just Starship and basic plugins"},
			{"power-user", "Full-featured setup with all plugins and nerd font support"},
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tDESCRIPTION")
		for _, p := range presets {
			fmt.Fprintf(w, "%s\t%s\n", p.name, p.description)
		}
		w.Flush()
		return nil
	},
}

var profilePresetInstallCmd = &cobra.Command{
	Use:   "import <name>",
	Short: "Import a profile preset",
	Long: `Import a profile preset and all its dependencies.

Available presets:
  default     - Balanced setup with Starship, autosuggestions, syntax-highlighting
  minimal     - Lightweight setup with just Starship
  power-user  - Full-featured setup with all plugins

Examples:
  dvt profile preset import default`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		var p *profile.Profile
		switch name {
		case "default":
			p = profile.DefaultProfile()
		case "minimal":
			p = profile.MinimalProfile()
		case "power-user":
			p = profile.PowerUserProfile()
		default:
			return fmt.Errorf("unknown preset: %s (available: default, minimal, power-user)", name)
		}

		store := getProfileStore()
		if err := store.Save(p); err != nil {
			return fmt.Errorf("failed to save profile: %w", err)
		}

		render.Successf("Installed profile preset '%s'", name)
		render.Info("Run 'dvt profile generate' to generate config files")
		return nil
	},
}

var profileApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a profile from file",
	Long: `Apply a profile definition from a YAML file.

Examples:
  dvt profile apply -f my-profile.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file with -f flag")
		}

		store := getProfileStore()

		for _, file := range files {
			var data []byte
			var err error
			var source string

			if file == "-" {
				data, err = io.ReadAll(os.Stdin)
				source = "stdin"
			} else {
				data, err = os.ReadFile(file)
				source = file
			}
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", source, err)
			}

			p, err := profile.Parse(data)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", source, err)
			}

			existing, _ := store.Get(p.Name)
			action := "created"
			if existing != nil {
				action = "updated"
			}

			if err := store.Save(p); err != nil {
				return fmt.Errorf("failed to save profile: %w", err)
			}

			render.Successf("Profile '%s' %s (from %s)", p.Name, action, source)
		}

		return nil
	},
}

var profileGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get profile definition(s)",
	Long: `Get terminal profile definitions.

With no arguments, lists all installed profiles.
With a name argument, gets a specific profile definition.

Examples:
  dvt profile get              # List installed profiles
  dvt profile get default      # Get specific profile`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			// List mode
			store := getProfileStore()
			profiles, err := store.List()
			if err != nil {
				return fmt.Errorf("failed to list profiles: %w", err)
			}

			if len(profiles) == 0 {
				render.Info("No profiles installed")
				render.Info("Use 'dvt profile preset get' to see available presets")
				return nil
			}

			format, _ := cmd.Flags().GetString("output")
			return outputProfiles(profiles, format)
		}
		// Single get mode
		name := args[0]
		store := getProfileStore()

		p, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("profile not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputProfile(p, format)
	},
}

var profileGenerateCmd = &cobra.Command{
	Use:   "generate <name>",
	Short: "Generate all config files for a profile",
	Long: `Generate all configuration files for a profile.

This creates:
  - starship.toml (if using Starship prompt)
  - Plugin installation/loading code
  - Shell aliases, env vars, functions

By default outputs to stdout. Use --output-dir to write to files.

Examples:
  dvt profile generate default
  dvt profile generate default --output-dir ~/.config/`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store := getProfileStore()

		p, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("profile not found: %s", name)
		}

		outputDir, _ := cmd.Flags().GetString("output-dir")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		gen := profile.NewGenerator("")
		result, err := gen.Generate(p)
		if err != nil {
			return fmt.Errorf("failed to generate profile: %w", err)
		}

		if outputDir == "" || dryRun {
			// Output to stdout
			fmt.Println("# Generated by dvt profile generate")
			fmt.Println("#")
			fmt.Printf("# Profile: %s\n", p.Name)
			if p.Description != "" {
				fmt.Printf("# Description: %s\n", p.Description)
			}
			fmt.Println()

			if result.StarshipTOML != "" {
				fmt.Println("# === starship.toml ===")
				fmt.Println(result.StarshipTOML)
				fmt.Println()
			}

			if result.ZshrcPlugins != "" {
				fmt.Println("# === .zshrc (plugins) ===")
				fmt.Println(result.ZshrcPlugins)
				fmt.Println()
			}

			if result.ZshrcShell != "" {
				fmt.Println("# === .zshrc (shell) ===")
				fmt.Println(result.ZshrcShell)
			}

			if dryRun && outputDir != "" {
				fmt.Println("\n# Would write to:")
				fmt.Printf("#   %s/starship.toml\n", outputDir)
				fmt.Printf("#   %s/.zshrc.dvt\n", outputDir)
			}
			return nil
		}

		// Write to files
		// Expand ~
		if strings.HasPrefix(outputDir, "~") {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, outputDir[1:])
		}

		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if result.StarshipTOML != "" {
			starshipPath := filepath.Join(outputDir, "starship.toml")
			if err := os.WriteFile(starshipPath, []byte(result.StarshipTOML), 0644); err != nil {
				return fmt.Errorf("failed to write starship.toml: %w", err)
			}
			render.Successf("Wrote %s", starshipPath)
		}

		if result.ZshrcPlugins != "" || result.ZshrcShell != "" {
			zshrcContent := "# Generated by dvt - source this from your .zshrc\n\n"
			if result.ZshrcPlugins != "" {
				zshrcContent += "# Plugins\n" + result.ZshrcPlugins + "\n"
			}
			if result.ZshrcShell != "" {
				zshrcContent += "# Shell config\n" + result.ZshrcShell
			}

			zshrcPath := filepath.Join(outputDir, ".zshrc.dvt")
			if err := os.WriteFile(zshrcPath, []byte(zshrcContent), 0644); err != nil {
				return fmt.Errorf("failed to write .zshrc.dvt: %w", err)
			}
			render.Successf("Wrote %s", zshrcPath)
			render.Blank()
			render.Info("Add this line to your .zshrc:")
			render.Plainf("  source %s", zshrcPath)
		}

		return nil
	},
}

var profileUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the active profile",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store := getProfileStore()

		if err := store.SetActive(name); err != nil {
			return err
		}

		render.Successf("Active profile set to '%s'", name)
		render.Info("Run 'dvt profile generate' to regenerate config files")
		return nil
	},
}

func init() {
	// Profile subcommands
	profileCmd.AddCommand(profilePresetCmd)
	profileCmd.AddCommand(profileApplyCmd)
	profileCmd.AddCommand(profileGetCmd)
	profileCmd.AddCommand(profileGenerateCmd)
	profileCmd.AddCommand(profileUseCmd)

	// Preset subcommands
	profilePresetCmd.AddCommand(profilePresetListCmd)
	profilePresetCmd.AddCommand(profilePresetInstallCmd)

	// Flags
	profileApplyCmd.Flags().StringSliceP("filename", "f", nil, "Profile YAML file(s)")
	profileGetCmd.Flags().StringP("output", "o", "yaml", "Output format: table, yaml, json")
	profileGenerateCmd.Flags().String("output-dir", "", "Output directory (default: stdout)")
	profileGenerateCmd.Flags().Bool("dry-run", false, "Show what would be generated")

	// Hidden backward-compat aliases for deprecated verbs in profile (after flags)
	profilePresetCmd.AddCommand(hiddenAlias("list", profilePresetListCmd))
	profilePresetCmd.AddCommand(hiddenAlias("install", profilePresetInstallCmd))
	profileCmd.AddCommand(hiddenAlias("list", profileGetCmd))
}

// =============================================================================
// COMPLETION COMMAND
// =============================================================================

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for dvt.

Examples:
  # Bash
  dvt completion bash > /etc/bash_completion.d/dvt
  
  # Zsh
  dvt completion zsh > "${fpath[1]}/_dvt"
  
  # Fish
  dvt completion fish > ~/.config/fish/completions/dvt.fish`,
	Args:      cobra.ExactValidArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
		}
		return nil
	},
}

// =============================================================================
// HELPERS
// =============================================================================

// hiddenAlias creates a hidden command that acts as a backward-compatible alias
// for a deprecated verb. The command prints a deprecation notice when used.
func hiddenAlias(name string, target *cobra.Command) *cobra.Command {
	alias := *target
	alias.Use = name
	alias.Aliases = nil
	alias.Hidden = true
	alias.Short = target.Short + " (deprecated: use " + target.Name() + ")"
	alias.Deprecated = "use '" + target.Name() + "' instead"
	return &alias
}

func getConfigDir() string {
	if configDir != "" {
		return configDir
	}
	if dir := os.Getenv("DVT_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return paths.New(home).DVTRoot()
}

// =============================================================================
// FILE STORES
// =============================================================================

// PromptFileStore implements prompt.PromptStore using file-based storage
type PromptFileStore struct {
	dir string
}

func getPromptStore() *PromptFileStore {
	dir := filepath.Join(getConfigDir(), "prompts")
	os.MkdirAll(dir, 0755)
	return &PromptFileStore{dir: dir}
}

func (s *PromptFileStore) Save(p *prompt.Prompt) error {
	data, err := yaml.Marshal(p.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, p.Name+".yaml"), data, 0644)
}

func (s *PromptFileStore) Get(name string) (*prompt.Prompt, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return prompt.Parse(data)
}

func (s *PromptFileStore) List() ([]*prompt.Prompt, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var prompts []*prompt.Prompt
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		p, err := s.Get(name)
		if err != nil {
			continue
		}
		prompts = append(prompts, p)
	}
	return prompts, nil
}

func (s *PromptFileStore) Delete(name string) error {
	return os.Remove(filepath.Join(s.dir, name+".yaml"))
}

func (s *PromptFileStore) Exists(name string) bool {
	_, err := os.Stat(filepath.Join(s.dir, name+".yaml"))
	return err == nil
}

func (s *PromptFileStore) Close() error { return nil }

// PluginFileStore implements plugin storage
type PluginFileStore struct {
	dir string
}

// getPluginStore extracts DataStore from command context and returns database-backed plugin store
func getPluginStore(cmd *cobra.Command) (plugin.PluginStore, error) {
	// Extract DataStore from context (following established dvt pattern)
	dataStoreInterface := cmd.Context().Value("dataStore")
	if dataStoreInterface == nil {
		return nil, fmt.Errorf("database not initialized - run 'dvt init' or check configuration")
	}

	dataStore := dataStoreInterface.(*db.DataStore)

	// Return database-backed plugin store via factory
	return terminalbridge.NewDBPluginStore(*dataStore), nil
}

func (s *PluginFileStore) Save(p *plugin.Plugin) error {
	data, err := yaml.Marshal(p.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, p.Name+".yaml"), data, 0644)
}

func (s *PluginFileStore) Get(name string) (*plugin.Plugin, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return plugin.Parse(data)
}

func (s *PluginFileStore) List() ([]*plugin.Plugin, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var plugins []*plugin.Plugin
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		p, err := s.Get(name)
		if err != nil {
			continue
		}
		plugins = append(plugins, p)
	}
	return plugins, nil
}

func (s *PluginFileStore) Delete(name string) error {
	return os.Remove(filepath.Join(s.dir, name+".yaml"))
}

// ShellFileStore implements shell config storage
type ShellFileStore struct {
	dir string
}

func getShellStore() *ShellFileStore {
	dir := filepath.Join(getConfigDir(), "shells")
	os.MkdirAll(dir, 0755)
	return &ShellFileStore{dir: dir}
}

func (s *ShellFileStore) Save(sh *shell.Shell) error {
	data, err := yaml.Marshal(sh.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, sh.Name+".yaml"), data, 0644)
}

func (s *ShellFileStore) Get(name string) (*shell.Shell, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return shell.Parse(data)
}

func (s *ShellFileStore) List() ([]*shell.Shell, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var shells []*shell.Shell
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		sh, err := s.Get(name)
		if err != nil {
			continue
		}
		shells = append(shells, sh)
	}
	return shells, nil
}

// ProfileFileStore implements profile storage
type ProfileFileStore struct {
	dir        string
	activePath string
}

func getProfileStore() *ProfileFileStore {
	dir := filepath.Join(getConfigDir(), "profiles")
	os.MkdirAll(dir, 0755)
	return &ProfileFileStore{
		dir:        dir,
		activePath: filepath.Join(getConfigDir(), ".active-profile"),
	}
}

func (s *ProfileFileStore) Save(p *profile.Profile) error {
	data, err := yaml.Marshal(p.ToYAML())
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.dir, p.Name+".yaml"), data, 0644)
}

func (s *ProfileFileStore) Get(name string) (*profile.Profile, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, name+".yaml"))
	if err != nil {
		return nil, err
	}
	return profile.Parse(data)
}

func (s *ProfileFileStore) List() ([]*profile.Profile, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var profiles []*profile.Profile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		p, err := s.Get(name)
		if err != nil {
			continue
		}
		profiles = append(profiles, p)
	}
	return profiles, nil
}

func (s *ProfileFileStore) SetActive(name string) error {
	// Verify profile exists
	if _, err := s.Get(name); err != nil {
		return fmt.Errorf("profile not found: %s", name)
	}
	return os.WriteFile(s.activePath, []byte(name), 0644)
}

func (s *ProfileFileStore) GetActive() (*profile.Profile, error) {
	data, err := os.ReadFile(s.activePath)
	if err != nil {
		return nil, nil
	}
	name := strings.TrimSpace(string(data))
	return s.Get(name)
}

// =============================================================================
// OUTPUT HELPERS
// =============================================================================

func outputPrompts(prompts []*prompt.Prompt, format string) error {
	sort.Slice(prompts, func(i, j int) bool {
		return prompts[i].Name < prompts[j].Name
	})

	switch format {
	case "yaml":
		for i, p := range prompts {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := yaml.Marshal(p.ToYAML())
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		var items []*prompt.PromptYAML
		for _, p := range prompts {
			items = append(items, p.ToYAML())
		}
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tTYPE\tCATEGORY\tDESCRIPTION")
		for _, p := range prompts {
			desc := p.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Type, p.Category, desc)
		}
		w.Flush()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputPrompt(p *prompt.Prompt, format string) error {
	switch format {
	case "yaml", "":
		data, err := yaml.Marshal(p.ToYAML())
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(p.ToYAML(), "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputPlugins(plugins []*plugin.Plugin, format string) error {
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})

	switch format {
	case "yaml":
		for i, p := range plugins {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := yaml.Marshal(p.ToYAML())
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		var items []*plugin.PluginYAML
		for _, p := range plugins {
			items = append(items, p.ToYAML())
		}
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tREPO\tMANAGER\tDESCRIPTION")
		for _, p := range plugins {
			desc := p.Description
			if len(desc) > 35 {
				desc = desc[:32] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Repo, p.Manager, desc)
		}
		w.Flush()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputPlugin(p *plugin.Plugin, format string) error {
	switch format {
	case "yaml", "":
		data, err := yaml.Marshal(p.ToYAML())
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(p.ToYAML(), "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputShells(shells []*shell.Shell, format string) error {
	sort.Slice(shells, func(i, j int) bool {
		return shells[i].Name < shells[j].Name
	})

	switch format {
	case "yaml":
		for i, s := range shells {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := yaml.Marshal(s.ToYAML())
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		var items []*shell.ShellYAML
		for _, s := range shells {
			items = append(items, s.ToYAML())
		}
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tSHELL\tALIASES\tENV VARS\tFUNCTIONS")
		for _, s := range shells {
			fmt.Fprintf(w, "%s\t%s\t%d\t%d\t%d\n",
				s.Name, s.ShellType, len(s.Aliases), len(s.Env), len(s.Functions))
		}
		w.Flush()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputShell(s *shell.Shell, format string) error {
	switch format {
	case "yaml", "":
		data, err := yaml.Marshal(s.ToYAML())
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(s.ToYAML(), "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputProfiles(profiles []*profile.Profile, format string) error {
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	switch format {
	case "yaml":
		for i, p := range profiles {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := yaml.Marshal(p.ToYAML())
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		var items []*profile.ProfileYAML
		for _, p := range profiles {
			items = append(items, p.ToYAML())
		}
		data, err := json.MarshalIndent(items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tPROMPT\tPLUGINS\tDESCRIPTION")
		for _, p := range profiles {
			desc := p.Description
			if len(desc) > 35 {
				desc = desc[:32] + "..."
			}
			promptName := ""
			if p.Prompt != nil {
				promptName = p.Prompt.Name
			}
			fmt.Fprintf(w, "%s\t%s\t%d\t%s\n", p.Name, promptName, len(p.Plugins), desc)
		}
		w.Flush()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputProfile(p *profile.Profile, format string) error {
	switch format {
	case "yaml", "":
		data, err := yaml.Marshal(p.ToYAML())
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(p.ToYAML(), "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}
