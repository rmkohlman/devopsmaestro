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

	"devopsmaestro/pkg/terminalops/plugin"
	pluginlibrary "devopsmaestro/pkg/terminalops/plugin/library"
	"devopsmaestro/pkg/terminalops/profile"
	"devopsmaestro/pkg/terminalops/prompt"
	promptlibrary "devopsmaestro/pkg/terminalops/prompt/library"
	"devopsmaestro/pkg/terminalops/shell"
	"devopsmaestro/pkg/terminalops/store"

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
}

// setupDatabaseConfig configures database settings for dvt
func setupDatabaseConfig() {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Failed to get home directory", "error", err)
		os.Exit(1)
	}

	// Set config path to ~/.devopsmaestro (same as dvm for shared database)
	configPath := filepath.Join(homeDir, ".devopsmaestro")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.AutomaticEnv()

	// Try to read config, but don't fail if it doesn't exist
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			slog.Error("Failed to read config", "error", err)
			os.Exit(1)
		}
		// Config not found is OK - use defaults
	}

	// Set default values if config is not found (use same database as dvm)
	if viper.GetString("database.type") == "" {
		viper.Set("database.type", "sqlite")
		viper.Set("database.path", "~/.devopsmaestro/devopsmaestro.db")
		viper.Set("store", "sql")
	}
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

	// Initialize logging and database before any command runs
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		initLogging()

		// Check if this is a command that doesn't need database
		skipDB := false
		commandName := cmd.Name()
		switch commandName {
		case "completion", "version", "help":
			skipDB = true
		}

		if !skipDB {
			// Setup database configuration
			setupDatabaseConfig()

			// Initialize database connection
			dbInstance, err := db.InitializeDBConnection()
			if err != nil {
				slog.Error("Failed to initialize database", "error", err)
				os.Exit(1)
			}

			// Create DataStore instance
			dataStore, err := db.StoreFactory(dbInstance)
			if err != nil {
				slog.Error("Failed to create DataStore", "error", err)
				os.Exit(1)
			}

			// Set the dataStore in context for resource operations
			ctx := context.WithValue(cmd.Context(), "dataStore", &dataStore)
			cmd.SetContext(ctx)

			// Auto-migrate database if needed (skip for commands that don't need DB)
			if shouldSkipAutoMigration(cmd) {
				return
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
						fmt.Printf("Warning: Migrations not found, skipping auto-migration: %v\n", err)
					}
					return
				}

				// Use version-based auto-migration
				migrationsApplied, err := db.CheckVersionBasedAutoMigration(driver, migrationsFS, Version, verbose)
				if err != nil {
					slog.Error("auto-migration failed", "error", err)
					fmt.Printf("Error: Failed to apply database migrations: %v\n", err)
					fmt.Println("Please run 'dvm admin migrate' to fix migration issues.")
					os.Exit(1)
				}
				if migrationsApplied && verbose {
					slog.Info("database migrations applied successfully")
				}
			}
		}
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
			fmt.Fprintf(os.Stderr, "Warning: could not open log file %s: %v\n", logFile, err)
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

		fmt.Printf("Initialized dvt at %s\n", dir)
		return nil
	},
}

// =============================================================================
// PROMPT COMMANDS
// =============================================================================

var promptCmd = &cobra.Command{
	Use:   "prompt",
	Short: "Manage terminal prompts (Starship, P10k)",
	Long: `Manage terminal prompt configurations.

Prompts define how your shell prompt looks using tools like Starship or P10k.
Use the library to get started with pre-configured prompts, then customize as needed.`,
}

var promptLibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Browse and install prompts from the library",
}

var promptLibraryListCmd = &cobra.Command{
	Use:   "list",
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
			fmt.Println("No prompts found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPrompts(prompts, format)
	},
}

var promptLibraryShowCmd = &cobra.Command{
	Use:   "show <name>",
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
	Use:   "install <name>...",
	Short: "Install prompts from library to local store",
	Long: `Copy prompt definitions from the built-in library to your local store.
You can then customize them or use them directly.

Examples:
  dvt prompt library install starship-default
  dvt prompt library install starship-minimal starship-powerline
  dvt prompt library install --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if !all && len(args) == 0 {
			return fmt.Errorf("specify prompt names or use --all")
		}

		lib, err := promptlibrary.NewPromptLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		store := getPromptStore()

		var prompts []*prompt.Prompt
		if all {
			prompts = lib.List()
		} else {
			for _, name := range args {
				p, err := lib.Get(name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: prompt not found in library: %s\n", name)
					continue
				}
				prompts = append(prompts, p)
			}
		}

		for _, p := range prompts {
			if err := store.Save(p); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to install %s: %v\n", p.Name, err)
				continue
			}
			fmt.Printf("Installed %s\n", p.Name)
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
		fmt.Printf("Categories (%d):\n", len(categories))
		for _, c := range categories {
			prompts := lib.ListByCategory(c)
			fmt.Printf("  %-15s (%d prompts)\n", c, len(prompts))
		}
		return nil
	},
}

var promptListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed prompts",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := getPromptStore()
		prompts, err := store.List()
		if err != nil {
			return fmt.Errorf("failed to list prompts: %w", err)
		}

		if len(prompts) == 0 {
			fmt.Println("No prompts installed")
			fmt.Println("\nUse 'dvt prompt library list' to see available prompts")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPrompts(prompts, format)
	},
}

var promptGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a prompt definition (kubectl-style)",
	Long: `Get a terminal prompt stored in the database.

Uses Resource/Handler pattern with database storage.

Examples:
  dvt prompt get coolnight           # Get prompt as YAML
  dvt prompt get coolnight -o json   # Get prompt as JSON
  dvt prompt get coolnight -o table  # Get prompt as table`,
	Args: cobra.ExactArgs(1),
	RunE: promptResourceGet,
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
	promptCmd.AddCommand(promptListCmd)
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
	promptLibraryInstallCmd.Flags().Bool("all", false, "Install all prompts from library")
	promptListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	promptGetCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	promptApplyCmd.Flags().StringSliceP("filename", "f", nil, "Prompt YAML file(s)")
	promptDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
}

// =============================================================================
// PLUGIN COMMANDS
// =============================================================================

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage shell plugins (zsh-autosuggestions, etc.)",
	Long: `Manage shell plugin configurations.

Plugins enhance your shell with features like autosuggestions, syntax highlighting,
and fuzzy finding. The library provides pre-configured plugins that work well together.`,
}

var pluginLibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Browse and install plugins from the library",
}

var pluginLibraryListCmd = &cobra.Command{
	Use:   "list",
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
			fmt.Println("No plugins found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugins(plugins, format)
	},
}

var pluginLibraryShowCmd = &cobra.Command{
	Use:   "show <name>",
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
	Use:   "install <name>...",
	Short: "Install plugins from library to local store",
	Long: `Copy plugin definitions from the built-in library to your local store.

Examples:
  dvt plugin library install zsh-autosuggestions
  dvt plugin library install zsh-autosuggestions zsh-syntax-highlighting
  dvt plugin library install --all`,
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
					fmt.Fprintf(os.Stderr, "Warning: plugin not found in library: %s\n", name)
					continue
				}
				plugins = append(plugins, p)
			}
		}

		for _, p := range plugins {
			if err := store.Upsert(p); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to install %s: %v\n", p.Name, err)
				continue
			}
			fmt.Printf("Installed %s\n", p.Name)
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
		fmt.Printf("Categories (%d):\n", len(categories))
		for _, c := range categories {
			plugins := lib.ListByCategory(c)
			fmt.Printf("  %-20s (%d plugins)\n", c, len(plugins))
		}
		return nil
	},
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		store, err := getPluginStore(cmd)
		if err != nil {
			return err
		}
		plugins, err := store.List()
		if err != nil {
			return fmt.Errorf("failed to list plugins: %w", err)
		}

		if len(plugins) == 0 {
			fmt.Println("No plugins installed")
			fmt.Println("\nUse 'dvt plugin library list' to see available plugins")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugins(plugins, format)
	},
}

var pluginGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a plugin definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
						fmt.Fprintf(os.Stderr, "Warning: plugin not found: %s\n", name)
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
	pluginCmd.AddCommand(pluginListCmd)
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
	pluginLibraryInstallCmd.Flags().Bool("all", false, "Install all plugins from library")
	pluginListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	pluginGetCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	pluginGenerateCmd.Flags().StringP("manager", "m", "manual", "Plugin manager: zinit, oh-my-zsh, antigen, sheldon, manual")
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

			fmt.Printf("Shell config '%s' %s (from %s)\n", s.Name, action, source)
		}

		return nil
	},
}

var shellListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed shell configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := getShellStore()
		shells, err := store.List()
		if err != nil {
			return fmt.Errorf("failed to list shell configs: %w", err)
		}

		if len(shells) == 0 {
			fmt.Println("No shell configs installed")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputShells(shells, format)
	},
}

var shellGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a shell configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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
	shellCmd.AddCommand(shellListCmd)
	shellCmd.AddCommand(shellGetCmd)
	shellCmd.AddCommand(shellGenerateCmd)

	// Flags
	shellApplyCmd.Flags().StringSliceP("filename", "f", nil, "Shell YAML file(s)")
	shellListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	shellGetCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
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
  dvt profile preset list       # See available presets
  dvt profile preset install default
  dvt profile generate          # Generate all config files`,
}

var profilePresetCmd = &cobra.Command{
	Use:   "preset",
	Short: "Manage profile presets",
}

var profilePresetListCmd = &cobra.Command{
	Use:   "list",
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
	Use:   "install <name>",
	Short: "Install a profile preset",
	Long: `Install a profile preset and all its dependencies.

Available presets:
  default     - Balanced setup with Starship, autosuggestions, syntax-highlighting
  minimal     - Lightweight setup with just Starship
  power-user  - Full-featured setup with all plugins

Examples:
  dvt profile preset install default`,
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

		fmt.Printf("Installed profile preset '%s'\n", name)
		fmt.Println("\nRun 'dvt profile generate' to generate config files")
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

			fmt.Printf("Profile '%s' %s (from %s)\n", p.Name, action, source)
		}

		return nil
	},
}

var profileListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := getProfileStore()
		profiles, err := store.List()
		if err != nil {
			return fmt.Errorf("failed to list profiles: %w", err)
		}

		if len(profiles) == 0 {
			fmt.Println("No profiles installed")
			fmt.Println("\nUse 'dvt profile preset list' to see available presets")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputProfiles(profiles, format)
	},
}

var profileGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a profile definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
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

By default outputs to stdout. Use --output to write to files.

Examples:
  dvt profile generate default
  dvt profile generate default --output ~/.config/`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		store := getProfileStore()

		p, err := store.Get(name)
		if err != nil {
			return fmt.Errorf("profile not found: %s", name)
		}

		outputDir, _ := cmd.Flags().GetString("output")
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
			fmt.Printf("Wrote %s\n", starshipPath)
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
			fmt.Printf("Wrote %s\n", zshrcPath)
			fmt.Println("\nAdd this line to your .zshrc:")
			fmt.Printf("  source %s\n", zshrcPath)
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

		fmt.Printf("Active profile set to '%s'\n", name)
		fmt.Println("\nRun 'dvt profile generate' to regenerate config files")
		return nil
	},
}

func init() {
	// Profile subcommands
	profileCmd.AddCommand(profilePresetCmd)
	profileCmd.AddCommand(profileApplyCmd)
	profileCmd.AddCommand(profileListCmd)
	profileCmd.AddCommand(profileGetCmd)
	profileCmd.AddCommand(profileGenerateCmd)
	profileCmd.AddCommand(profileUseCmd)

	// Preset subcommands
	profilePresetCmd.AddCommand(profilePresetListCmd)
	profilePresetCmd.AddCommand(profilePresetInstallCmd)

	// Flags
	profileApplyCmd.Flags().StringSliceP("filename", "f", nil, "Profile YAML file(s)")
	profileListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	profileGetCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	profileGenerateCmd.Flags().StringP("output", "o", "", "Output directory (default: stdout)")
	profileGenerateCmd.Flags().Bool("dry-run", false, "Show what would be generated")
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

func getConfigDir() string {
	if configDir != "" {
		return configDir
	}
	if dir := os.Getenv("DVT_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".dvt")
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
	return store.NewDBPluginStore(*dataStore), nil
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
