package cmd

import (
	"context"
	"devopsmaestro/db"
	"devopsmaestro/pkg/colorbridge"
	"devopsmaestro/pkg/crd"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/utils"
	"fmt"
	"github.com/rmkohlman/MaestroSDK/colors"
	"github.com/rmkohlman/MaestroSDK/render"
	theme "github.com/rmkohlman/MaestroTheme"
	"io/fs"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var (
	verbose      bool
	logLevel     string
	logFormat    string
	logFile      string
	noColor      bool
	outputFormat string
	themeFlag    string
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
	SilenceErrors: true,
	SilenceUsage:  true,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(dataStore *db.DataStore, executor *Executor, migrationsFS fs.FS) {
	// Explicit initialization: register all resource handlers at startup
	handlers.RegisterAll()

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Initialize logging
		initLogging()

		// Initialize ColorProvider - construct adapter chain at composition root
		themePath := colors.GetDefaultThemePath()
		var paletteProvider colors.PaletteProvider
		if themePath != "" {
			store := theme.NewFileStore(themePath)
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

		// Set the dataStore and executor for all commands
		ctx = context.WithValue(ctx, CtxKeyDataStore, dataStore)
		ctx = context.WithValue(ctx, ctxKeyExecutor, executor)
		ctx = context.WithValue(ctx, ctxKeyMigrationsFS, migrationsFS)
		cmd.SetContext(ctx)

		// Auto-migrate database if needed (skip for commands that don't need DB)
		if shouldSkipAutoMigration(cmd) {
			return nil
		}

		if dataStore != nil && *dataStore != nil {
			driver := (*dataStore).Driver()
			if driver != nil {
				// Use version-based auto-migration for better performance
				migrationsApplied, err := db.CheckVersionBasedAutoMigration(driver, migrationsFS, Version, verbose)
				if err != nil {
					// Migration failure is critical - return error via errSilent
					slog.Error("auto-migration failed", "error", err)
					render.Errorf("Failed to apply database migrations: %v", err)
					render.Info("Please run 'dvm admin migrate' to fix migration issues.")
					return errSilent
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
		return nil
	}

	if err := rootCmd.ExecuteContext(buildSignalContext()); err != nil {
		// errSilent means the command already displayed the error via render.Error()
		if err != errSilent {
			render.Errorf("%s", err)
		}
		os.Exit(1)
	}
}

// buildSignalContext returns a context that is cancelled on SIGINT (Ctrl-C)
// or SIGTERM. The cancellation propagates through cobra's cmd.Context() so
// that the parallel build engine can mark in-flight workspaces and the build
// session as "interrupted" rather than leaving them stuck in "running" /
// "building" forever (#399).
//
// We deliberately use signal.NotifyContext rather than relying on Go's default
// SIGINT handler — the default aborts the process immediately, bypassing all
// defers and the engine's finalization writes.
func buildSignalContext() context.Context {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	return ctx
}

// shouldSkipAutoMigration determines if auto-migration should be skipped for this command.
// Skip for commands that don't need the database or handle migrations themselves.
func shouldSkipAutoMigration(cmd *cobra.Command) bool {
	cmdPath := cmd.CommandPath()

	// Skip for Cobra's hidden completion commands (__complete, __completeNoDesc).
	// These are invoked by the shell on every TAB press and must not trigger
	// database migrations or any other side effects.
	if strings.HasPrefix(cmdPath, "dvm __complete") {
		return true
	}

	// Skip for commands that don't need database
	skipCommands := []string{
		"dvm completion",
		"dvm version",
		"dvm help",
		"dvm generate-docs",     // dev tool: no database needed
		"dvm generate template", // template generation: no database needed
		"dvm admin init",        // init handles its own migrations
		"dvm admin migrate",     // migrate command handles migrations explicitly
		"dvm sandbox",           // sandboxes are runtime-only, no database needed
		"dvm sandbox create",
		"dvm sandbox get",
		"dvm sandbox attach",
		"dvm sandbox delete",
		"dvm system info",  // system maintenance: runtime-only, no database needed
		"dvm system df",    // system maintenance: runtime-only, no database needed
		"dvm system prune", // system maintenance: runtime-only, no database needed
	}

	for _, skipCmd := range skipCommands {
		if cmdPath == skipCmd {
			return true
		}
	}

	return false
}

func init() {
	// Disable Cobra's auto-generated completion command — we provide our own
	// custom completionCmd (in completion.go) that fixes zsh autoload
	// compatibility by stripping the bare "compdef" line from the output.
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Register our custom completion command
	rootCmd.AddCommand(completionCmd)

	// Global flags available to all commands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging (shortcut for --log-level=debug)")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "warn", "Set log level (debug, info, warn, error)")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Set log format (text, json)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Write logs to file (JSON format)")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Output format flag — persistent so all subcommands inherit it
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "table",
		"Output format: table, json, yaml, plain, compact, wide")

	// Theme flag — persistent so all subcommands inherit it
	rootCmd.PersistentFlags().StringVar(&themeFlag, "theme", "",
		"Color theme for output (overrides DVM_THEME and config)")
}

// initLogging configures the global slog logger based on flags.
// - Default: WARN level, text format (logs discarded unless level elevated)
// - With --verbose / -v: DEBUG level to stderr
// - With --log-level: sets the minimum log level
// - With --log-format: sets output format (text or json)
// - With --log-file: JSON format to file (overrides --log-format)
func initLogging() {
	// --verbose is a shortcut for --log-level=debug
	effectiveLevel := logLevel
	if verbose {
		effectiveLevel = "debug"
	}

	// When writing to a log file, always use JSON format
	if logFile != "" {
		f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			render.WarningfToStderr("Could not open log file %s: %v", logFile, err)
			utils.InitLogger(effectiveLevel, logFormat)
			return
		}
		lvl := utils.ParseLogLevel(effectiveLevel)
		opts := &slog.HandlerOptions{Level: lvl}
		handler := slog.NewJSONHandler(f, opts)
		slog.SetDefault(slog.New(handler))
		return
	}

	utils.InitLogger(effectiveLevel, logFormat)
}
