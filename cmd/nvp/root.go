package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"

	"devopsmaestro/db"
	"devopsmaestro/pkg/colors"
	"devopsmaestro/pkg/nvimops"
	nvimconfig "devopsmaestro/pkg/nvimops/config"
	"devopsmaestro/pkg/nvimops/library"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
	"devopsmaestro/pkg/nvimops/sync"
	"devopsmaestro/pkg/nvimops/sync/sources"
	"devopsmaestro/pkg/nvimops/theme"
	themelibrary "devopsmaestro/pkg/nvimops/theme/library"
	"devopsmaestro/pkg/nvimops/theme/parametric"
	"devopsmaestro/pkg/resource"
	"devopsmaestro/pkg/resource/handlers"
	"devopsmaestro/pkg/source"

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
	case "completion", "version", "help":
		return true
	}
	return false
}

// setupDatabaseConfig configures database settings for nvp
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
  nvp library list              # See available plugins
  nvp library install telescope # Install from library  
  nvp generate                  # Generate Lua files

Configuration is stored in ~/.nvp/ by default.`,
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
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		initLogging()

		// Initialize ColorProvider for nvp
		// nvp uses its own theme path under ~/.nvp/themes
		nvpThemePath := filepath.Join(getConfigDir(), "themes")
		ctx, err := colors.InitColorProviderForCommand(
			cmd.Context(),
			nvpThemePath,
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
			setupDatabaseConfig()

			// Initialize database connection
			dbInstance, err := db.InitializeDBConnection()
			if err != nil {
				slog.Warn("Failed to initialize database (using file-based storage)", "error", err)
				return // Continue without database - nvp can work with file-based storage
			}

			// Create DataStore instance
			dataStore, err := db.StoreFactory(dbInstance)
			if err != nil {
				slog.Warn("Failed to create DataStore (using file-based storage)", "error", err)
				return // Continue without database
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
				// Get migrations FS - for nvp we try to find migrations on disk
				migrationsFS := getMigrationsFS()
				if migrationsFS == nil {
					// If migrations not found, warn but continue
					// This allows nvp to work even when migrations are not available
					if verbose {
						slog.Warn("migrations not available for auto-migration in nvp")
						fmt.Printf("Warning: Migrations not found, skipping auto-migration\n")
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
	rootCmd.AddCommand(listCmd)
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
		fmt.Printf("nvp (NvimOps) %s\n", Version)
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
	Short: "Initialize nvp configuration",
	Long: `Initialize the nvp configuration directory and plugin store.

This creates:
  ~/.nvp/
  ~/.nvp/plugins/     # Plugin YAML storage
  ~/.nvp/config.yaml  # Configuration file (optional)

You can specify a custom directory with --config or NVP_CONFIG_DIR.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getConfigDir()

		// Create directories
		pluginsDir := filepath.Join(dir, "plugins")
		if err := os.MkdirAll(pluginsDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		fmt.Printf("✓ Initialized nvp at %s\n", dir)
		return nil
	},
}

// =============================================================================
// LIBRARY COMMANDS
// =============================================================================

var libraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Browse and install from the plugin library",
	Long: `The plugin library contains curated, pre-configured plugin definitions
that work well together. Use these commands to explore and install plugins.`,
}

var libraryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all plugins in the library",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		plugins := lib.List()

		// Filter by category if specified
		category, _ := cmd.Flags().GetString("category")
		if category != "" {
			plugins = lib.ListByCategory(category)
		}

		// Filter by tag if specified
		tag, _ := cmd.Flags().GetString("tag")
		if tag != "" {
			plugins = lib.ListByTag(tag)
		}

		if len(plugins) == 0 {
			fmt.Println("No plugins found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugins(plugins, format)
	},
}

var libraryShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show details of a library plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		lib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		p, ok := lib.Get(name)
		if !ok {
			return fmt.Errorf("plugin not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugin(p, format)
	},
}

var libraryInstallCmd = &cobra.Command{
	Use:   "install <name>...",
	Short: "Install plugins from library to local store",
	Long: `Copy plugin definitions from the built-in library to your local store.
You can then customize them with 'nvp get' and 'nvp apply'.

Examples:
  nvp library install telescope
  nvp library install telescope treesitter lspconfig
  nvp library install --all`,
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if !all && len(args) == 0 {
			return fmt.Errorf("specify plugin names or use --all")
		}

		slog.Debug("loading library")
		lib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		var plugins []*plugin.Plugin
		if all {
			plugins = lib.List()
			slog.Info("installing all plugins from library", "count", len(plugins))
		} else {
			for _, name := range args {
				p, ok := lib.Get(name)
				if !ok {
					slog.Warn("plugin not found in library", "name", name)
					fmt.Fprintf(os.Stderr, "Warning: plugin not found in library: %s\n", name)
					continue
				}
				plugins = append(plugins, p)
			}
			slog.Info("installing plugins from library", "count", len(plugins), "names", args)
		}

		for _, p := range plugins {
			p.Enabled = true
			if err := mgr.Apply(p); err != nil {
				slog.Error("failed to install plugin", "name", p.Name, "error", err)
				fmt.Fprintf(os.Stderr, "Warning: failed to install %s: %v\n", p.Name, err)
				continue
			}
			slog.Debug("installed plugin", "name", p.Name)
			fmt.Printf("✓ Installed %s\n", p.Name)
		}

		return nil
	},
}

var libraryCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List all plugin categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		categories := lib.Categories()
		fmt.Printf("Categories (%d):\n", len(categories))
		for _, c := range categories {
			plugins := lib.ListByCategory(c)
			fmt.Printf("  %-15s (%d plugins)\n", c, len(plugins))
		}
		return nil
	},
}

var libraryTagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all plugin tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		lib, err := library.NewLibrary()
		if err != nil {
			return fmt.Errorf("failed to load library: %w", err)
		}

		tags := lib.Tags()
		fmt.Printf("Tags (%d):\n", len(tags))
		for _, t := range tags {
			plugins := lib.ListByTag(t)
			fmt.Printf("  %-20s (%d plugins)\n", t, len(plugins))
		}
		return nil
	},
}

func init() {
	libraryCmd.AddCommand(libraryListCmd)
	libraryCmd.AddCommand(libraryShowCmd)
	libraryCmd.AddCommand(libraryInstallCmd)
	libraryCmd.AddCommand(libraryCategoriesCmd)
	libraryCmd.AddCommand(libraryTagsCmd)

	libraryListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	libraryListCmd.Flags().StringP("category", "c", "", "Filter by category")
	libraryListCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	libraryShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	libraryInstallCmd.Flags().Bool("all", false, "Install all plugins from library")
}

// =============================================================================
// APPLY COMMAND
// =============================================================================

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a plugin definition from file or URL",
	Long: `Apply a plugin definition from a YAML file or URL to the local store.
If the plugin already exists, it will be updated.

The -f flag accepts local files, URLs, or stdin (use '-' for stdin).
URLs starting with http://, https://, or github: are fetched automatically.

GitHub shorthand: github:user/repo/path/file.yaml
   
Examples:
  nvp apply -f telescope.yaml
  nvp apply -f plugin1.yaml -f plugin2.yaml
  nvp apply -f https://raw.githubusercontent.com/user/repo/main/plugin.yaml
  nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml
  cat plugin.yaml | nvp apply -f -`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file or URL with -f flag")
		}

		// Create resource context for file-based storage
		ctx := resource.Context{
			ConfigDir: getConfigDir(),
		}

		// Process files and URLs using unified source resolution
		for _, src := range files {
			srcObj := source.Resolve(src)
			data, displayName, err := srcObj.Read()
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", src, err)
			}

			// Use unified resource pipeline
			res, err := resource.Apply(ctx, data, displayName)
			if err != nil {
				return fmt.Errorf("failed to apply from %s: %w", displayName, err)
			}

			// Determine if this was a create or update based on the resource type
			// For now, just report success
			slog.Info("resource applied", "kind", res.GetKind(), "name", res.GetName(), "source", displayName)
			fmt.Printf("✓ %s '%s' applied (from %s)\n", res.GetKind(), res.GetName(), displayName)
		}

		return nil
	},
}

func init() {
	applyCmd.Flags().StringSliceP("filename", "f", nil, "Plugin YAML file(s) or URL(s) to apply (use '-' for stdin)")
}

// =============================================================================
// LIST COMMAND
// =============================================================================

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List plugins in the local store",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		plugins, err := mgr.List()
		if err != nil {
			return fmt.Errorf("failed to list plugins: %w", err)
		}

		// Filter by category
		category, _ := cmd.Flags().GetString("category")
		if category != "" {
			var filtered []*plugin.Plugin
			for _, p := range plugins {
				if p.Category == category {
					filtered = append(filtered, p)
				}
			}
			plugins = filtered
		}

		// Filter by enabled/disabled
		enabled, _ := cmd.Flags().GetBool("enabled")
		disabled, _ := cmd.Flags().GetBool("disabled")
		if enabled || disabled {
			var filtered []*plugin.Plugin
			for _, p := range plugins {
				if (enabled && p.Enabled) || (disabled && !p.Enabled) {
					filtered = append(filtered, p)
				}
			}
			plugins = filtered
		}

		if len(plugins) == 0 {
			fmt.Println("No plugins found")
			return nil
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugins(plugins, format)
	},
}

func init() {
	listCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	listCmd.Flags().StringP("category", "c", "", "Filter by category")
	listCmd.Flags().Bool("enabled", false, "Show only enabled plugins")
	listCmd.Flags().Bool("disabled", false, "Show only disabled plugins")
}

// =============================================================================
// GET COMMAND
// =============================================================================

var getCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a plugin definition",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		p, err := mgr.Get(name)
		if err != nil {
			return fmt.Errorf("plugin not found: %s", name)
		}

		format, _ := cmd.Flags().GetString("output")
		return outputPlugin(p, format)
	},
}

func init() {
	getCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
}

// =============================================================================
// DELETE COMMAND
// =============================================================================

var deleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a plugin from the local store",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		// Check exists
		if _, err := mgr.Get(name); err != nil {
			return fmt.Errorf("plugin not found: %s", name)
		}

		// Confirm unless forced
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Delete plugin '%s'? (y/N): ", name)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted")
				return nil
			}
		}

		if err := mgr.Delete(name); err != nil {
			return fmt.Errorf("failed to delete plugin: %w", err)
		}

		fmt.Printf("✓ Plugin '%s' deleted\n", name)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
}

// =============================================================================
// ENABLE/DISABLE COMMANDS
// =============================================================================

var enableCmd = &cobra.Command{
	Use:   "enable <name>...",
	Short: "Enable plugins for Lua generation",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setPluginsEnabled(args, true)
	},
}

var disableCmd = &cobra.Command{
	Use:   "disable <name>...",
	Short: "Disable plugins (exclude from Lua generation)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return setPluginsEnabled(args, false)
	},
}

func setPluginsEnabled(names []string, enabled bool) error {
	mgr, err := getManager()
	if err != nil {
		return err
	}
	defer mgr.Close()

	action := "enabled"
	if !enabled {
		action = "disabled"
	}

	for _, name := range names {
		p, err := mgr.Get(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Warning: plugin not found: %s\n", name)
			continue
		}

		p.Enabled = enabled
		if err := mgr.Apply(p); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: failed to update %s: %v\n", name, err)
			continue
		}

		fmt.Printf("✓ Plugin '%s' %s\n", name, action)
	}

	return nil
}

// =============================================================================
// GENERATE COMMANDS
// =============================================================================

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Lua files for all enabled plugins",
	Long: `Generate lazy.nvim compatible Lua files for all enabled plugins.

By default, files are written to ~/.config/nvim/lua/plugins/nvp/
Use --output to specify a different directory.

Examples:
  nvp generate
  nvp generate --output ~/.config/nvim/lua/plugins/managed
  nvp generate --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		outputDir, _ := cmd.Flags().GetString("output")
		if outputDir == "" {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, ".config", "nvim", "lua", "plugins", "nvp")
		}

		// Expand ~
		if strings.HasPrefix(outputDir, "~") {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, outputDir[1:])
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		slog.Debug("generate command", "outputDir", outputDir, "dryRun", dryRun)

		plugins, err := mgr.List()
		if err != nil {
			return fmt.Errorf("failed to list plugins: %w", err)
		}

		// Filter to enabled only
		var enabled []*plugin.Plugin
		for _, p := range plugins {
			if p.Enabled {
				enabled = append(enabled, p)
			}
		}

		slog.Info("generating Lua files", "total", len(plugins), "enabled", len(enabled))

		if len(enabled) == 0 {
			fmt.Println("No enabled plugins to generate")
			return nil
		}

		if dryRun {
			fmt.Printf("Would generate %d Lua files to %s:\n", len(enabled), outputDir)
			for _, p := range enabled {
				fmt.Printf("  %s.lua\n", p.Name)
			}
			return nil
		}

		// Create output directory
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Generate files
		gen := plugin.NewGenerator()
		for _, p := range enabled {
			lua, err := gen.GenerateLuaFile(p)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to generate %s: %v\n", p.Name, err)
				continue
			}

			filename := filepath.Join(outputDir, p.Name+".lua")
			if err := os.WriteFile(filename, []byte(lua), 0644); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to write %s: %v\n", filename, err)
				continue
			}

			if verbose {
				fmt.Printf("  Generated %s\n", filename)
			}
		}

		fmt.Printf("✓ Generated %d Lua files to %s\n", len(enabled), outputDir)
		return nil
	},
}

var generateLuaCmd = &cobra.Command{
	Use:   "generate-lua <name>",
	Short: "Generate Lua for a single plugin (stdout)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		p, err := mgr.Get(name)
		if err != nil {
			// Try library as fallback
			lib, libErr := library.NewLibrary()
			if libErr == nil {
				if libPlugin, ok := lib.Get(name); ok {
					p = libPlugin
				}
			}
			if p == nil {
				return fmt.Errorf("plugin not found: %s", name)
			}
		}

		gen := plugin.NewGenerator()
		lua, err := gen.GenerateLuaFile(p)
		if err != nil {
			return fmt.Errorf("failed to generate Lua: %w", err)
		}

		fmt.Print(lua)
		return nil
	},
}

func init() {
	generateCmd.Flags().StringP("output", "o", "", "Output directory")
	generateCmd.Flags().Bool("dry-run", false, "Show what would be generated")
}

// =============================================================================
// THEME COMMANDS
// =============================================================================

var themeCmd = &cobra.Command{
	Use:   "theme",
	Short: "Manage Neovim themes",
	Long: `Manage Neovim colorscheme themes using YAML definitions.

Themes define:
  - The colorscheme plugin to use (tokyonight, catppuccin, etc.)
  - Color palette that other plugins can reference
  - Custom color overrides

The active theme's palette is exported as a Lua module that other plugins
(lualine, bufferline, telescope, etc.) can use for consistent styling.

Examples:
  nvp theme library list              # See available themes
  nvp theme library install catppuccin-mocha
  nvp theme use catppuccin-mocha      # Set as active theme
  nvp theme get                       # Show active theme`,
}

var themeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed themes",
	RunE: func(cmd *cobra.Command, args []string) error {
		themeStore := getThemeStore()

		themes, err := themeStore.List()
		if err != nil {
			return fmt.Errorf("failed to list themes: %w", err)
		}

		if len(themes) == 0 {
			fmt.Println("No themes installed")
			fmt.Println("\nUse 'nvp theme library list' to see available themes")
			return nil
		}

		// Get active theme
		active, _ := themeStore.GetActive()
		activeName := ""
		if active != nil {
			activeName = active.Name
		}

		format, _ := cmd.Flags().GetString("output")
		return outputThemes(themes, format, activeName)
	},
}

var themeGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get theme details (defaults to active theme)",
	RunE: func(cmd *cobra.Command, args []string) error {
		themeStore := getThemeStore()

		var t *theme.Theme
		var err error

		if len(args) > 0 {
			t, err = themeStore.Get(args[0])
		} else {
			t, err = themeStore.GetActive()
			if t == nil && err == nil {
				return fmt.Errorf("no active theme set. Use 'nvp theme use <name>' to set one")
			}
		}

		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("output")
		return outputTheme(t, format)
	},
}

var themeApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a theme from file or URL",
	Long: `Apply a theme definition from a YAML file or URL.

The -f flag accepts local files, URLs, or stdin (use '-' for stdin).
URLs starting with http://, https://, or github: are fetched automatically.

GitHub shorthand: github:user/repo/path/file.yaml
   
Examples:
  nvp theme apply -f my-theme.yaml
  nvp theme apply -f github:user/repo/themes/custom.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")

		if len(files) == 0 {
			return fmt.Errorf("must specify at least one file or URL with -f flag")
		}

		// Create resource context for file-based storage
		ctx := resource.Context{
			ConfigDir: getConfigDir(),
		}

		// Process files and URLs using unified source resolution
		for _, src := range files {
			srcObj := source.Resolve(src)
			data, displayName, err := srcObj.Read()
			if err != nil {
				return fmt.Errorf("failed to read %s: %w", src, err)
			}

			// Use unified resource pipeline
			res, err := resource.Apply(ctx, data, displayName)
			if err != nil {
				return fmt.Errorf("failed to apply from %s: %w", displayName, err)
			}

			slog.Info("resource applied", "kind", res.GetKind(), "name", res.GetName(), "source", displayName)
			fmt.Printf("✓ %s '%s' applied (from %s)\n", res.GetKind(), res.GetName(), displayName)
		}

		return nil
	},
}

var themeDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		themeStore := getThemeStore()

		// Check exists
		if _, err := themeStore.Get(name); err != nil {
			return fmt.Errorf("theme not found: %s", name)
		}

		// Confirm unless forced
		force, _ := cmd.Flags().GetBool("force")
		if !force {
			fmt.Printf("Delete theme '%s'? (y/N): ", name)
			var response string
			fmt.Scanln(&response)
			if response != "y" && response != "Y" {
				fmt.Println("Aborted")
				return nil
			}
		}

		if err := themeStore.Delete(name); err != nil {
			return fmt.Errorf("failed to delete theme: %w", err)
		}

		fmt.Printf("✓ Theme '%s' deleted\n", name)
		return nil
	},
}

var themeUseCmd = &cobra.Command{
	Use:   "use <name>",
	Short: "Set the active theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		themeStore := getThemeStore()

		if err := themeStore.SetActive(name); err != nil {
			return err
		}

		fmt.Printf("✓ Active theme set to '%s'\n", name)
		fmt.Println("\nRun 'nvp generate' to regenerate Lua files with the new theme")
		return nil
	},
}

// Theme library commands
var themeLibraryCmd = &cobra.Command{
	Use:   "library",
	Short: "Browse and install themes from the library",
}

var themeLibraryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available themes in the library",
	RunE: func(cmd *cobra.Command, args []string) error {
		themes, err := themelibrary.List()
		if err != nil {
			return fmt.Errorf("failed to list library themes: %w", err)
		}

		if len(themes) == 0 {
			fmt.Println("No themes in library")
			return nil
		}

		// Filter by category
		category, _ := cmd.Flags().GetString("category")
		if category != "" {
			themes, err = themelibrary.ListByCategory(category)
			if err != nil {
				return err
			}
		}

		format, _ := cmd.Flags().GetString("output")
		return outputThemeInfos(themes, format)
	},
}

var themeLibraryShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show details of a library theme",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		t, err := themelibrary.Get(name)
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("output")
		return outputTheme(t, format)
	},
}

var themeLibraryInstallCmd = &cobra.Command{
	Use:   "install <name>...",
	Short: "Install themes from library",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		themeStore := getThemeStore()
		if err := themeStore.Init(); err != nil {
			return err
		}

		setActive, _ := cmd.Flags().GetBool("use")
		var lastInstalled string

		for _, name := range args {
			t, err := themelibrary.Get(name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: theme not found in library: %s\n", name)
				continue
			}

			if err := themeStore.Save(t); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to install %s: %v\n", name, err)
				continue
			}

			fmt.Printf("✓ Installed theme '%s'\n", t.Name)
			lastInstalled = t.Name
		}

		// Set active if requested
		if setActive && lastInstalled != "" {
			if err := themeStore.SetActive(lastInstalled); err != nil {
				return err
			}
			fmt.Printf("✓ Active theme set to '%s'\n", lastInstalled)
		}

		return nil
	},
}

var themeLibraryCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List theme categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		categories, err := themelibrary.Categories()
		if err != nil {
			return err
		}

		fmt.Printf("Categories (%d):\n", len(categories))
		for _, c := range categories {
			themes, _ := themelibrary.ListByCategory(c)
			fmt.Printf("  %-10s (%d themes)\n", c, len(themes))
		}
		return nil
	},
}

var themePreviewCmd = &cobra.Command{
	Use:   "preview <name>",
	Short: "Preview a theme's colors in the terminal",
	Long: `Preview a theme's color palette directly in the terminal.

This shows how the theme's colors will look, including:
  - Background and foreground colors
  - Syntax highlighting colors (keywords, strings, comments, etc.)
  - UI elements (errors, warnings, selections, etc.)
  - Sample code snippet with the theme applied

The preview works with both installed themes and library themes.

Examples:
  nvp theme preview tokyonight-night
  nvp theme preview catppuccin-mocha
  nvp theme preview --all              # Preview all library themes`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")

		if all {
			// Preview all library themes
			themes, err := themelibrary.List()
			if err != nil {
				return err
			}
			for i, info := range themes {
				if i > 0 {
					fmt.Println()
					fmt.Println(strings.Repeat("─", 60))
					fmt.Println()
				}
				t, err := themelibrary.Get(info.Name)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not load %s: %v\n", info.Name, err)
					continue
				}
				printThemePreview(t)
			}
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("theme name required (or use --all)")
		}

		name := args[0]

		// Try installed themes first
		themeStore := getThemeStore()
		t, err := themeStore.Get(name)
		if err != nil {
			// Try library
			t, err = themelibrary.Get(name)
			if err != nil {
				return fmt.Errorf("theme not found: %s (not installed or in library)", name)
			}
		}

		printThemePreview(t)
		return nil
	},
}

// printThemePreview renders a colorized preview of a theme
func printThemePreview(t *theme.Theme) {
	// Get colors with fallbacks
	getColor := func(key string, fallback string) string {
		if c, ok := t.Colors[key]; ok && c != "" {
			return c
		}
		return fallback
	}

	bg := getColor("bg", "#1a1b26")
	fg := getColor("fg", "#c0caf5")
	comment := getColor("comment", "#565f89")
	red := getColor("red", "#f7768e")
	green := getColor("green", "#9ece6a")
	yellow := getColor("yellow", "#e0af68")
	blue := getColor("blue", "#7aa2f7")
	magenta := getColor("magenta", "#bb9af7")
	cyan := getColor("cyan", "#7dcfff")
	orange := getColor("orange", "#ff9e64")

	// Header
	fmt.Printf("╭─────────────────────────────────────────────────────────╮\n")
	fmt.Printf("│ %s%-55s%s │\n", colorFgRGB(fg), centerText(t.Name, 55), "\033[0m")
	if t.Description != "" {
		desc := t.Description
		if len(desc) > 55 {
			desc = desc[:52] + "..."
		}
		fmt.Printf("│ %s%-55s%s │\n", colorFgRGB(comment), centerText(desc, 55), "\033[0m")
	}
	fmt.Printf("╰─────────────────────────────────────────────────────────╯\n\n")

	// Color swatches
	fmt.Printf("  %s████%s %s████%s %s████%s %s████%s %s████%s %s████%s %s████%s %s████%s\n",
		colorBgRGB(bg), "\033[0m",
		colorBgRGB(red), "\033[0m",
		colorBgRGB(green), "\033[0m",
		colorBgRGB(yellow), "\033[0m",
		colorBgRGB(blue), "\033[0m",
		colorBgRGB(magenta), "\033[0m",
		colorBgRGB(cyan), "\033[0m",
		colorBgRGB(fg), "\033[0m",
	)
	fmt.Printf("  %-4s %-4s %-4s %-4s %-4s %-4s %-4s %-4s\n",
		"bg", "red", "grn", "yel", "blu", "mag", "cyn", "fg")
	fmt.Println()

	// Sample code preview with syntax highlighting
	fmt.Printf("  %s// Sample code preview%s\n", colorFgRGB(comment), "\033[0m")
	fmt.Printf("  %sfunc%s %smain%s() {\n", colorFgRGB(magenta), "\033[0m", colorFgRGB(blue), "\033[0m")
	fmt.Printf("      %smessage%s %s:=%s %s\"Hello, World!\"%s\n",
		colorFgRGB(fg), "\033[0m",
		colorFgRGB(cyan), "\033[0m",
		colorFgRGB(green), "\033[0m")
	fmt.Printf("      %scount%s %s:=%s %s42%s\n",
		colorFgRGB(fg), "\033[0m",
		colorFgRGB(cyan), "\033[0m",
		colorFgRGB(orange), "\033[0m")
	fmt.Printf("      %sfmt%s.%sPrintln%s(%smessage%s, %scount%s)\n",
		colorFgRGB(cyan), "\033[0m",
		colorFgRGB(blue), "\033[0m",
		colorFgRGB(fg), "\033[0m",
		colorFgRGB(fg), "\033[0m")
	fmt.Printf("      %s// TODO: add more features%s\n", colorFgRGB(comment), "\033[0m")
	fmt.Printf("  }\n\n")

	// Diagnostics preview
	fmt.Printf("  %s✗ Error: something went wrong%s\n", colorFgRGB(red), "\033[0m")
	fmt.Printf("  %s⚠ Warning: deprecated function%s\n", colorFgRGB(yellow), "\033[0m")
	fmt.Printf("  %sℹ Info: build completed%s\n", colorFgRGB(blue), "\033[0m")
	fmt.Printf("  %s✓ Success: all tests passed%s\n", colorFgRGB(green), "\033[0m")
}

// colorFgRGB returns ANSI escape code for true color foreground
func colorFgRGB(hex string) string {
	r, g, b, err := parseHexColor(hex)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// colorBgRGB returns ANSI escape code for true color background
func colorBgRGB(hex string) string {
	r, g, b, err := parseHexColor(hex)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

// parseHexColor converts hex color to RGB values
func parseHexColor(hex string) (r, g, b int, err error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return 0, 0, 0, fmt.Errorf("invalid hex color: %s", hex)
	}
	_, err = fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return
}

// centerText centers text within a given width
func centerText(s string, width int) string {
	if len(s) >= width {
		return s[:width]
	}
	padding := (width - len(s)) / 2
	return strings.Repeat(" ", padding) + s + strings.Repeat(" ", width-len(s)-padding)
}

var themeCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom theme using parametric generation",
	Long: `Create a new theme using parametric generation from a base color.
	
This command uses the CoolNight Ocean theme as a reference and generates
a new theme by shifting the color hue while maintaining the same saturation
and lightness relationships. Semantic colors (red, green, yellow) are kept
fixed for consistency.

The --from flag accepts:
  - Hex colors: #FF5733, #3498db
  - Hue values: 270 (degrees 0-360)
  - Preset names: synthwave, matrix, arctic

Examples:
  nvp theme create --from "#8B00FF" --name my-purple-theme
  nvp theme create --from "150" --name my-green-theme
  nvp theme create --from "synthwave" --dry-run
  nvp theme create --from "#FF6B35" --name sunset-coding -o json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fromValue, _ := cmd.Flags().GetString("from")
		name, _ := cmd.Flags().GetString("name")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		output, _ := cmd.Flags().GetString("output")

		if fromValue == "" {
			return fmt.Errorf("--from flag is required (hex color, hue, or preset name)")
		}

		// Generate theme name if not provided and not dry run
		if name == "" && !dryRun {
			return fmt.Errorf("--name flag is required (or use --dry-run to preview)")
		}

		generator := parametric.NewGenerator()

		var generatedTheme *theme.Theme
		var err error

		// Check if it's a preset name first
		if preset, exists := parametric.GetPreset(fromValue); exists {
			_ = preset // Use preset to avoid unused variable
			generatedTheme, err = parametric.GeneratePreset(fromValue)
			if err != nil {
				return fmt.Errorf("failed to generate preset %s: %w", fromValue, err)
			}
		} else {
			// Try parsing as hex color
			if strings.HasPrefix(fromValue, "#") {
				description := fmt.Sprintf("Custom theme generated from %s", fromValue)
				if name != "" {
					description = fmt.Sprintf("Custom %s theme", name)
				}
				generatedTheme, err = generator.GenerateFromHex(fromValue, name, description)
				if err != nil {
					return fmt.Errorf("invalid hex color %s: %w", fromValue, err)
				}
			} else {
				// Try parsing as hue value
				hue, parseErr := strconv.ParseFloat(fromValue, 64)
				if parseErr != nil {
					return fmt.Errorf("invalid input %s: expected hex color (#rrggbb), hue (0-360), or preset name", fromValue)
				}

				if hue < 0 || hue >= 360 {
					return fmt.Errorf("hue must be between 0 and 360, got %.1f", hue)
				}

				description := fmt.Sprintf("Custom theme generated from hue %.1f°", hue)
				if name != "" {
					description = fmt.Sprintf("Custom %s theme (hue %.1f°)", name, hue)
				}
				generatedTheme = generator.GenerateFromHue(hue, name, description)
			}
		}

		// Dry run - just output the theme
		if dryRun {
			fmt.Printf("Generated theme preview:\n")
			return outputTheme(generatedTheme, output)
		}

		// Save theme to store
		themeStore := getThemeStore()
		if err := themeStore.Init(); err != nil {
			return err
		}

		if err := themeStore.Save(generatedTheme); err != nil {
			return fmt.Errorf("failed to save theme: %w", err)
		}

		fmt.Printf("✓ Created theme '%s'\n", generatedTheme.Name)

		// Optionally set as active
		setActive, _ := cmd.Flags().GetBool("use")
		if setActive {
			if err := themeStore.SetActive(generatedTheme.Name); err != nil {
				return err
			}
			fmt.Printf("✓ Set '%s' as active theme\n", generatedTheme.Name)
		}

		// Show what to do next
		fmt.Println("\nNext steps:")
		if !setActive {
			fmt.Printf("  nvp theme use %s      # Set as active theme\n", generatedTheme.Name)
		}
		fmt.Println("  nvp generate          # Generate Lua files")
		fmt.Printf("  nvp theme preview %s  # Preview colors\n", generatedTheme.Name)

		return nil
	},
}

var themeGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Lua files for the active theme",
	Long: `Generate Lua files for the active theme including:
  - theme/palette.lua   - Color palette module for other plugins
  - theme/init.lua      - Theme setup and helpers
  - plugins/colorscheme.lua - Lazy.nvim plugin spec

Other plugins can use the palette:
  local palette = require("theme").palette
  local bg = palette.colors.bg`,
	RunE: func(cmd *cobra.Command, args []string) error {
		themeStore := getThemeStore()

		t, err := themeStore.GetActive()
		if err != nil {
			return err
		}
		if t == nil {
			return fmt.Errorf("no active theme set. Use 'nvp theme use <name>' first")
		}

		outputDir, _ := cmd.Flags().GetString("output")
		if outputDir == "" {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, ".config", "nvim", "lua")
		}

		// Expand ~
		if strings.HasPrefix(outputDir, "~") {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, outputDir[1:])
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")

		gen := theme.NewGenerator()
		generated, err := gen.Generate(t)
		if err != nil {
			return fmt.Errorf("failed to generate theme: %w", err)
		}

		files := map[string]string{
			filepath.Join(outputDir, "theme", "palette.lua"):              generated.PaletteLua,
			filepath.Join(outputDir, "theme", "init.lua"):                 generated.InitLua,
			filepath.Join(outputDir, "plugins", "nvp", "colorscheme.lua"): generated.PluginLua,
		}

		if dryRun {
			fmt.Printf("Would generate theme files for '%s':\n", t.Name)
			for path := range files {
				fmt.Printf("  %s\n", path)
			}
			return nil
		}

		for path, content := range files {
			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", path, err)
			}
			if verbose {
				fmt.Printf("  Generated %s\n", path)
			}
		}

		fmt.Printf("✓ Generated theme '%s' to %s\n", t.Name, outputDir)
		fmt.Println("\nOther plugins can now use: require(\"theme\").palette")
		return nil
	},
}

func init() {
	// Theme subcommands
	themeCmd.AddCommand(themeListCmd)
	themeCmd.AddCommand(themeGetCmd)
	themeCmd.AddCommand(themeApplyCmd)
	themeCmd.AddCommand(themeCreateCmd)
	themeCmd.AddCommand(themeDeleteCmd)
	themeCmd.AddCommand(themeUseCmd)
	themeCmd.AddCommand(themeLibraryCmd)
	themeCmd.AddCommand(themeGenerateCmd)
	themeCmd.AddCommand(themePreviewCmd)

	// Theme library subcommands
	themeLibraryCmd.AddCommand(themeLibraryListCmd)
	themeLibraryCmd.AddCommand(themeLibraryShowCmd)
	themeLibraryCmd.AddCommand(themeLibraryInstallCmd)
	themeLibraryCmd.AddCommand(themeLibraryCategoriesCmd)

	// Flags
	themeListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	themeGetCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	themeApplyCmd.Flags().StringSliceP("filename", "f", nil, "Theme YAML file(s) or URL(s) to apply")
	themeCreateCmd.Flags().String("from", "", "Base color (hex #rrggbb, hue 0-360, or preset name)")
	themeCreateCmd.Flags().String("name", "", "Theme name (required unless --dry-run)")
	themeCreateCmd.Flags().Bool("dry-run", false, "Preview without saving")
	themeCreateCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json, table")
	themeCreateCmd.Flags().Bool("use", false, "Set as active theme after creation")
	themeDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
	themePreviewCmd.Flags().Bool("all", false, "Preview all library themes")
	themeLibraryListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	themeLibraryListCmd.Flags().StringP("category", "c", "", "Filter by category (dark, light)")
	themeLibraryShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	themeLibraryInstallCmd.Flags().Bool("use", false, "Set as active theme after install")
	themeGenerateCmd.Flags().StringP("output", "o", "", "Output directory (default: ~/.config/nvim/lua)")
	themeGenerateCmd.Flags().Bool("dry-run", false, "Show what would be generated")
}

func getThemeStore() *theme.FileStore {
	dir := getConfigDir()
	return theme.NewFileStore(dir)
}

func outputThemes(themes []*theme.Theme, format string, activeName string) error {
	switch format {
	case "yaml":
		for i, t := range themes {
			if i > 0 {
				fmt.Println("---")
			}
			data, err := t.ToYAML()
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		}
	case "json":
		data, err := json.MarshalIndent(themes, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCATEGORY\tPLUGIN\tACTIVE\tDESCRIPTION")
		for _, t := range themes {
			active := ""
			if t.Name == activeName {
				active = "*"
			}
			desc := t.Description
			if len(desc) > 35 {
				desc = desc[:32] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", t.Name, t.Category, t.Plugin.Repo, active, desc)
		}
		w.Flush()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputTheme(t *theme.Theme, format string) error {
	switch format {
	case "yaml", "":
		data, err := t.ToYAML()
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(t, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

func outputThemeInfos(themes []themelibrary.ThemeInfo, format string) error {
	switch format {
	case "yaml":
		data, err := yaml.Marshal(themes)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		data, err := json.MarshalIndent(themes, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	case "table", "":
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCATEGORY\tPLUGIN\tDESCRIPTION")
		for _, t := range themes {
			desc := t.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", t.Name, t.Category, t.Plugin, desc)
		}
		w.Flush()
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}

// =============================================================================
// COMPLETION COMMAND
// =============================================================================

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate shell completion script for nvp.

Examples:
  # Bash
  nvp completion bash > /etc/bash_completion.d/nvp
  
  # Zsh
  nvp completion zsh > "${fpath[1]}/_nvp"
  
  # Fish
  nvp completion fish > ~/.config/fish/completions/nvp.fish`,
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
// CONFIG COMMANDS
// =============================================================================

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage Neovim core configuration",
	Long: `Manage Neovim core configuration (options, keymaps, autocmds).

The config command generates a complete Neovim configuration structure
from YAML definitions, matching the lua/workspace/ directory pattern.

Generated structure:
  ~/.config/nvim/
  ├── init.lua                    # Entry point
  └── lua/workspace/
      ├── lazy.lua                # lazy.nvim bootstrap
      ├── core/
      │   ├── init.lua
      │   ├── options.lua         # vim.opt settings
      │   ├── keymaps.lua         # Key mappings
      │   └── autocmds.lua        # Autocommands
      └── plugins/
          ├── init.lua            # Base plugins
          └── *.lua               # Plugin configs

Quick Start:
  nvp config init                 # Create default core.yaml
  nvp config show                 # View current config
  nvp config generate             # Generate full nvim structure`,
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize core.yaml with default settings",
	Long: `Create a default core.yaml configuration file.

This creates a sensible starting configuration with:
  - Common vim options (line numbers, tabs, search, etc.)
  - Essential keymaps (leader key, window splits, etc.)
  - Useful autocmds (yank highlight, etc.)
  - Base plugins (plenary, tmux-navigator)

The file is created at ~/.nvp/core.yaml by default.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getConfigDir()
		configPath := filepath.Join(dir, "core.yaml")

		// Check if already exists
		force, _ := cmd.Flags().GetBool("force")
		if _, err := os.Stat(configPath); err == nil && !force {
			return fmt.Errorf("core.yaml already exists at %s (use --force to overwrite)", configPath)
		}

		// Create directory if needed
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Write default config
		cfg := nvimconfig.DefaultCoreConfig()
		if err := cfg.WriteYAMLFile(configPath); err != nil {
			return fmt.Errorf("failed to write core.yaml: %w", err)
		}

		fmt.Printf("✓ Created %s\n", configPath)
		fmt.Println("\nEdit this file to customize your Neovim configuration.")
		fmt.Println("Then run 'nvp config generate' to create the Lua files.")
		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current core configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadCoreConfig()
		if err != nil {
			return err
		}

		format, _ := cmd.Flags().GetString("output")
		switch format {
		case "yaml", "":
			data, err := cfg.ToYAML()
			if err != nil {
				return err
			}
			fmt.Print(string(data))
		case "json":
			data, err := json.MarshalIndent(cfg, "", "  ")
			if err != nil {
				return err
			}
			fmt.Println(string(data))
		default:
			return fmt.Errorf("unknown format: %s", format)
		}
		return nil
	},
}

var configGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate complete Neovim configuration",
	Long: `Generate a complete Neovim configuration from core.yaml and installed plugins.

This creates the full lua/workspace/ directory structure:
  - init.lua (entry point)
  - lua/workspace/lazy.lua (lazy.nvim bootstrap)
  - lua/workspace/core/*.lua (options, keymaps, autocmds)
  - lua/workspace/plugins/*.lua (plugin configurations)

By default, files are written to ~/.config/nvim/
Use --output to specify a different directory.

Examples:
  nvp config generate
  nvp config generate --output /path/to/nvim/config
  nvp config generate --dry-run`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load core config
		cfg, err := loadCoreConfig()
		if err != nil {
			// If no core.yaml exists, use defaults
			if os.IsNotExist(err) {
				fmt.Println("No core.yaml found, using defaults...")
				cfg = nvimconfig.DefaultCoreConfig()
			} else {
				return err
			}
		}

		// Load plugins
		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		plugins, err := mgr.List()
		if err != nil {
			return fmt.Errorf("failed to list plugins: %w", err)
		}

		// Filter to enabled
		var enabled []*plugin.Plugin
		for _, p := range plugins {
			if p.Enabled {
				enabled = append(enabled, p)
			}
		}

		// Output directory
		outputDir, _ := cmd.Flags().GetString("output")
		if outputDir == "" {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, ".config", "nvim")
		}

		// Expand ~
		if strings.HasPrefix(outputDir, "~") {
			home, _ := os.UserHomeDir()
			outputDir = filepath.Join(home, outputDir[1:])
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		ns := cfg.Namespace
		if ns == "" {
			ns = "workspace"
		}

		if dryRun {
			fmt.Printf("Would generate Neovim config to %s:\n", outputDir)
			fmt.Printf("  init.lua\n")
			fmt.Printf("  lua/%s/lazy.lua\n", ns)
			fmt.Printf("  lua/%s/core/init.lua\n", ns)
			fmt.Printf("  lua/%s/core/options.lua\n", ns)
			fmt.Printf("  lua/%s/core/keymaps.lua\n", ns)
			fmt.Printf("  lua/%s/core/autocmds.lua\n", ns)
			fmt.Printf("  lua/%s/plugins/init.lua\n", ns)
			for _, p := range enabled {
				fmt.Printf("  lua/%s/plugins/%s.lua\n", ns, p.Name)
			}
			// Check for active theme
			themeStore := getThemeStore()
			if activeTheme, _ := themeStore.GetActive(); activeTheme != nil {
				fmt.Printf("  lua/%s/plugins/colorscheme.lua (theme: %s)\n", ns, activeTheme.Name)
				fmt.Printf("  lua/theme/init.lua\n")
				fmt.Printf("  lua/theme/palette.lua\n")
			}
			return nil
		}

		// Generate
		gen := nvimconfig.NewGenerator()
		if err := gen.WriteToDirectory(cfg, enabled, outputDir); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		// Generate theme if active
		themeStore := getThemeStore()
		activeTheme, _ := themeStore.GetActive()
		if activeTheme != nil {
			themeGen := theme.NewGenerator()
			generated, err := themeGen.Generate(activeTheme)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to generate theme: %v\n", err)
			} else {
				// Write theme files
				themeFiles := map[string]string{
					filepath.Join(outputDir, "lua", "theme", "palette.lua"):           generated.PaletteLua,
					filepath.Join(outputDir, "lua", "theme", "init.lua"):              generated.InitLua,
					filepath.Join(outputDir, "lua", ns, "plugins", "colorscheme.lua"): generated.PluginLua,
				}

				for path, content := range themeFiles {
					dir := filepath.Dir(path)
					if err := os.MkdirAll(dir, 0755); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to create %s: %v\n", dir, err)
						continue
					}
					if err := os.WriteFile(path, []byte(content), 0644); err != nil {
						fmt.Fprintf(os.Stderr, "Warning: failed to write %s: %v\n", path, err)
						continue
					}
				}
				fmt.Printf("  Theme: %s (colorscheme.lua)\n", activeTheme.Name)
			}
		}

		fmt.Printf("✓ Generated Neovim configuration to %s\n", outputDir)
		fmt.Printf("  Core files: init.lua, lua/%s/core/*.lua\n", ns)
		fmt.Printf("  Plugin files: %d plugins in lua/%s/plugins/\n", len(enabled), ns)
		fmt.Println("\nRestart Neovim to apply changes.")
		return nil
	},
}

var configEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open core.yaml in editor",
	RunE: func(cmd *cobra.Command, args []string) error {
		dir := getConfigDir()
		configPath := filepath.Join(dir, "core.yaml")

		// Create default if doesn't exist
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			fmt.Println("No core.yaml found, creating default...")
			if err := os.MkdirAll(dir, 0755); err != nil {
				return err
			}
			cfg := nvimconfig.DefaultCoreConfig()
			if err := cfg.WriteYAMLFile(configPath); err != nil {
				return err
			}
		}

		// Find editor
		editor := os.Getenv("EDITOR")
		if editor == "" {
			editor = os.Getenv("VISUAL")
		}
		if editor == "" {
			editor = "vim"
		}

		// Open editor
		editorCmd := fmt.Sprintf("%s %s", editor, configPath)
		fmt.Printf("Opening %s in %s...\n", configPath, editor)
		return runCommand(editorCmd)
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGenerateCmd)
	configCmd.AddCommand(configEditCmd)

	configInitCmd.Flags().Bool("force", false, "Overwrite existing core.yaml")
	configShowCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	configGenerateCmd.Flags().StringP("output", "o", "", "Output directory (default: ~/.config/nvim)")
	configGenerateCmd.Flags().Bool("dry-run", false, "Show what would be generated")
}

func loadCoreConfig() (*nvimconfig.CoreConfig, error) {
	dir := getConfigDir()
	configPath := filepath.Join(dir, "core.yaml")
	return nvimconfig.ParseYAMLFile(configPath)
}

func runCommand(command string) error {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	cmd := parts[0]
	args := parts[1:]

	proc := &os.ProcAttr{
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}

	path, err := findExecutable(cmd)
	if err != nil {
		return err
	}

	process, err := os.StartProcess(path, append([]string{cmd}, args...), proc)
	if err != nil {
		return err
	}

	_, err = process.Wait()
	return err
}

func findExecutable(name string) (string, error) {
	if filepath.IsAbs(name) {
		return name, nil
	}

	paths := strings.Split(os.Getenv("PATH"), string(os.PathListSeparator))
	for _, dir := range paths {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}
	return "", fmt.Errorf("executable not found: %s", name)
}

// =============================================================================
// HELPERS
// =============================================================================

func getConfigDir() string {
	if configDir != "" {
		return configDir
	}
	if dir := os.Getenv("NVP_CONFIG_DIR"); dir != "" {
		return dir
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".nvp")
}

func getManager() (*nvimops.Manager, error) {
	dir := getConfigDir()
	pluginsDir := filepath.Join(dir, "plugins")

	// Auto-create if doesn't exist
	if err := os.MkdirAll(pluginsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	fileStore, err := store.NewFileStore(pluginsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	return nvimops.NewWithOptions(nvimops.Options{
		Store: fileStore,
	})
}

func outputPlugins(plugins []*plugin.Plugin, format string) error {
	// Sort by name
	sort.Slice(plugins, func(i, j int) bool {
		return plugins[i].Name < plugins[j].Name
	})

	switch format {
	case "yaml":
		for i, p := range plugins {
			if i > 0 {
				fmt.Println("---")
			}
			yml := p.ToYAML()
			data, err := yaml.Marshal(yml)
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
		fmt.Fprintln(w, "NAME\tCATEGORY\tENABLED\tDESCRIPTION")
		for _, p := range plugins {
			enabled := "yes"
			if !p.Enabled {
				enabled = "no"
			}
			desc := p.Description
			if len(desc) > 40 {
				desc = desc[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.Name, p.Category, enabled, desc)
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
		yml := p.ToYAML()
		data, err := yaml.Marshal(yml)
		if err != nil {
			return err
		}
		fmt.Print(string(data))
	case "json":
		yml := p.ToYAML()
		data, err := json.MarshalIndent(yml, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
	default:
		return fmt.Errorf("unknown format: %s", format)
	}
	return nil
}
