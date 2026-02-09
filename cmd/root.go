package cmd

import (
	"context"
	"devopsmaestro/db"
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
)

var rootCmd = &cobra.Command{
	Use:   "dvm",
	Short: "DevOpsMaestro CLI",
	Long: `DevOpsMaestro (dvm) is a CLI tool designed to manage development environments,
testing, deployments, and maintenance of code and software. It allows you to
create, manage, and deploy workspaces, apps, dependencies, and more.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(database *db.Database, dataStore *db.DataStore, executor *Executor, migrationsFS fs.FS) {
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		// Initialize logging
		initLogging()

		// Set the database, dataStore, and executor for all commands
		ctx := context.WithValue(cmd.Context(), "database", database)
		ctx = context.WithValue(ctx, "dataStore", dataStore)
		ctx = context.WithValue(ctx, "executor", executor)
		ctx = context.WithValue(ctx, "migrationsFS", migrationsFS)
		cmd.SetContext(ctx)
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags available to all commands
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable debug logging")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Write logs to file (JSON format)")
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
