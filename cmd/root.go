package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/pkg/colors"
	"devopsmaestro/pkg/crd"
	"devopsmaestro/pkg/nvimops/theme"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var (
	verbose bool
	logFile string
	noColor bool
)

// errSilent is returned by commands that have already displayed their error
// via render.Error(). It causes Cobra to set exit code 1 without double-printing.
var errSilent = fmt.Errorf("")

var rootCmd = &cobra.Command{
	Use:   "dvm",
	Short: "DevOpsMaestro CLI",
	Long: `DevOpsMaestro (dvm) is a CLI tool designed to manage development environments,
testing, deployments, and maintenance of code and software. It allows you to
create, manage, and deploy workspaces, apps, dependencies, and more.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(dataStore *db.DataStore, executor *Executor, migrationsFS fs.FS) {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Initialize logging
		initLogging()

		// Initialize ColorProvider - construct adapter chain at composition root
		themePath := colors.GetDefaultThemePath()
		var paletteProvider colors.PaletteProvider
		if themePath != "" {
			store := theme.NewFileStore(themePath)
			paletteProvider = colors.NewThemeStoreAdapter(store)
		}
		ctx, err := colors.InitColorProviderForCommand(
			cmd.Context(),
			paletteProvider,
			noColor,
		)
		if err != nil {
			slog.Warn("using default colors", "error", err)
		}

		// Set the dataStore and executor for all commands
		ctx = context.WithValue(ctx, "dataStore", dataStore)
		ctx = context.WithValue(ctx, "executor", executor)
		ctx = context.WithValue(ctx, "migrationsFS", migrationsFS)
		cmd.SetContext(ctx)

		// Auto-migrate database if needed (skip for commands that don't need DB)
		if shouldSkipAutoMigration(cmd) {
			return
		}

		if dataStore != nil && *dataStore != nil {
			driver := (*dataStore).Driver()
			if driver != nil {
				// Use version-based auto-migration for better performance
				migrationsApplied, err := db.CheckVersionBasedAutoMigration(driver, migrationsFS, Version, verbose)
				if err != nil {
					// Migration failure is critical - exit
					slog.Error("auto-migration failed", "error", err)
					fmt.Printf("Error: Failed to apply database migrations: %v\n", err)
					fmt.Println("Please run 'dvm admin migrate' to fix migration issues.")
					os.Exit(1)
				}

				if migrationsApplied && verbose {
					slog.Info("database migrations applied successfully")
				}
			}

			// Initialize CRD fallback handler for custom resources (v0.29.0)
			if err := crd.InitializeFallbackHandler(*dataStore); err != nil {
				slog.Warn("failed to initialize CRD handler", "error", err)
				// Don't exit - CRD support is optional, built-in resources still work
			}
		}
	}

	if err := rootCmd.Execute(); err != nil {
		// errSilent means the command already displayed the error via render.Error()
		if err != errSilent {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}

// shouldSkipAutoMigration determines if auto-migration should be skipped for this command.
// Skip for commands that don't need the database or handle migrations themselves.
func shouldSkipAutoMigration(cmd *cobra.Command) bool {
	cmdPath := cmd.CommandPath()

	// Skip for commands that don't need database
	skipCommands := []string{
		"dvm completion",
		"dvm version",
		"dvm help",
		"dvm admin init",    // init handles its own migrations
		"dvm admin migrate", // migrate command handles migrations explicitly
	}

	for _, skipCmd := range skipCommands {
		if cmdPath == skipCmd {
			return true
		}
	}

	return false
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Write logs to file (JSON format)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}

// initLogging configures the global slog logger based on flags.
// - Default: INFO level, text format to stderr (only shown with -v)
// - With --verbose: DEBUG level
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
			fmt.Fprintf(os.Stderr, "Warning: could not open log file %s: %v\n", logFile, err)
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
