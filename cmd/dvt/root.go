package main

import (
	"context"
	"devopsmaestro/db"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/rmkohlman/MaestroSDK/colors"
	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Global flags
	configDir string
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
