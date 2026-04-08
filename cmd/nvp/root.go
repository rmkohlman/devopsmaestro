package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"devopsmaestro/db"
	"devopsmaestro/pkg/colorbridge"
	"devopsmaestro/pkg/resource/handlers"
	"github.com/rmkohlman/MaestroNvim/nvimops/sync"
	"github.com/rmkohlman/MaestroNvim/nvimops/sync/sources"
	"github.com/rmkohlman/MaestroSDK/colors"
	"github.com/rmkohlman/MaestroSDK/paths"
	"github.com/rmkohlman/MaestroSDK/render"
	theme "github.com/rmkohlman/MaestroTheme"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
// Uses embedded migrations filesystem for Homebrew compatibility.
func getMigrationsFS() fs.FS {
	// Use embedded migrations (available when built with sync-migrations)
	if embeddedFS, err := GetEmbeddedMigrationsFS(); err == nil {
		return embeddedFS
	}

	// Fallback to filesystem search for development
	possiblePaths := []string{
		"db/migrations",       // Development: from repo root
		"./db/migrations",     // Current directory
		"../db/migrations",    // One level up
		"../../db/migrations", // Two levels up (from cmd/nvp)
	}

	for _, path := range possiblePaths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			// Check if it has sqlite subdirectory (validating it's the right path)
			sqlitePath := filepath.Join(path, "sqlite")
			if sqliteInfo, err := os.Stat(sqlitePath); err == nil && sqliteInfo.IsDir() {
				return os.DirFS(path)
			}
		}
	}

	return nil
}

// shouldSkipAutoMigration determines if auto-migration should be skipped for this command.
func shouldSkipAutoMigration(cmd *cobra.Command) bool {
	commandName := cmd.Name()
	switch commandName {
	case "completion", "version", "help", "generate-docs":
		return true
	}
	return false
}

// commandRequiresDatabase returns true for commands that require database access.
// These commands will fail fast if database initialization fails.
func commandRequiresDatabase(cmd *cobra.Command) bool {
	// Get the full command path (e.g., "package install", "sync", etc.)
	commandPath := cmd.CommandPath()
	commandName := cmd.Name()

	// Commands that require database
	switch {
	case strings.Contains(commandPath, "package install"):
		return true
	case commandName == "install" && cmd.Parent() != nil && cmd.Parent().Name() == "package":
		return true
	case commandName == "sync":
		return true
	case commandName == "get" && cmd.Parent() != nil && cmd.Parent().Name() == "nvp":
		// "nvp get" reads from DB if available, but can fall back to file store
		return false
	default:
		return false
	}
}

// setupDatabaseConfig configures database settings for nvp
func setupDatabaseConfig() error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
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
			return fmt.Errorf("failed to read config: %w", err)
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

// errSilent is returned by commands that have already displayed their error
// via render.Error(). It causes Cobra to set exit code 1 without double-printing.
var errSilent = fmt.Errorf("")

// rootCmd is the base command
var rootCmd = &cobra.Command{
	Use:   "nvp",
	Short: "NvimOps - DevOps-style Neovim plugin management",
	Long: `nvp (NvimOps) is a CLI tool for managing Neovim plugins using a DevOps-style
YAML configuration approach. It provides:

  - YAML-based plugin definitions (kubectl-style)
  - Built-in library of curated plugins
  - Lua code generation for lazy.nvim
  - File-based storage (no database required)

Quick Start:
  nvp library get                 # See available plugins
  nvp library import telescope    # Import from library  
  nvp generate                  # Generate Lua files

Configuration is stored in ~/.nvp/ by default.`,
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configDir, "config", "", "Config directory (default: ~/.nvp)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Write logs to file (JSON format)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Initialize logging and ColorProvider before any command runs
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		initLogging()

		// Initialize ColorProvider for nvp
		// nvp uses its own theme path under ~/.nvp/themes
		nvpThemePath := filepath.Join(getConfigDir(), "themes")
		var paletteProvider colors.PaletteProvider
		if nvpThemePath != "" {
			store := theme.NewFileStore(nvpThemePath)
			paletteProvider = colorbridge.NewThemeStoreAdapter(store)
		}
		ctx, err := colors.InitColorProviderForCommand(
			cmd.Context(),
			paletteProvider,
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

		// Initialize database connection for commands that need it
		// (nvp uses file-based storage by default, but some features like packages use the database)
		if !skipDB {
			// Setup database configuration
			if err := setupDatabaseConfig(); err != nil {
				return fmt.Errorf("database config: %w", err)
			}

			// Create DataStore instance
			dataStore, err := db.CreateDataStore()
			if err != nil {
				// Check if this command requires database
				if commandRequiresDatabase(cmd) {
					// Fail fast with clear error message
					slog.Error("Database required but unavailable", "error", err)
					render.ErrorToStderr("Database required but unavailable")
					render.InfoToStderr("Run 'dvm admin init' to initialize the database, or check ~/.devopsmaestro/devopsmaestro.db exists")
					render.ErrorfToStderr("Details: %v", err)
					return errSilent
				}
				// For optional DB commands, just warn and continue
				slog.Warn("Failed to initialize database (using file-based storage)", "error", err)
				return nil // Continue without database - nvp can work with file-based storage
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
				// Get migrations FS - for nvp we try to find migrations on disk
				migrationsFS := getMigrationsFS()
				if migrationsFS == nil {
					// If migrations not found, warn but continue
					// This allows nvp to work even when migrations are not available
					if verbose {
						slog.Warn("migrations not available for auto-migration in nvp")
						render.Warning("Migrations not found, skipping auto-migration")
					}
					return nil
				}

				// Use version-based auto-migration
				migrationsApplied, err := db.CheckVersionBasedAutoMigration(driver, migrationsFS, Version, verbose)
				if err != nil {
					slog.Error("auto-migration failed", "error", err)
					render.Errorf("Failed to apply database migrations: %v", err)
					render.Info("Please run 'dvm admin migrate' to fix migration issues.")
					return errSilent
				}
				if migrationsApplied && verbose {
					slog.Info("database migrations applied successfully")
				}
			}
		}
		return nil
	}

	// Register resource handlers for unified pipeline
	handlers.RegisterAll()

	// Initialize sync sources registry with builtin sources
	if err := sync.InitializeGlobalRegistry(); err != nil {
		// Log warning but don't fail - sync functionality will be limited
		slog.Warn("failed to initialize sync sources", "error", err)
	}

	// Register actual source handlers (replaces placeholder handlers)
	if err := sources.RegisterAllGlobalHandlers(); err != nil {
		// Log warning but don't fail - some sources will use placeholder handlers
		slog.Warn("failed to register source handlers", "error", err)
	}

	// Add all commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(libraryCmd)
	rootCmd.AddCommand(packageCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(enableCmd)
	rootCmd.AddCommand(disableCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(generateLuaCmd)
	rootCmd.AddCommand(themeCmd)
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(completionCmd)
}

// initLogging configures the global slog logger based on flags.
// - Default: Silent (logs discarded)
// - With --verbose: DEBUG level to stderr
// - With --log-file: JSON format to file
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
		// JSON format for file output (machine-readable)
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			render.WarningfToStderr("could not open log file %s: %v", logFile, err)
			handler = slog.NewTextHandler(os.Stderr, opts)
		} else {
			handler = slog.NewJSONHandler(f, opts)
		}
	} else if verbose {
		// Text format for terminal (human-readable), only when verbose
		handler = slog.NewTextHandler(os.Stderr, opts)
	} else {
		// Silent by default - discard logs unless verbose or log-file specified
		handler = slog.NewTextHandler(io.Discard, opts)
	}

	slog.SetDefault(slog.New(handler))
}
