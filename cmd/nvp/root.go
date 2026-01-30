package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"devopsmaestro/pkg/nvimops"
	"devopsmaestro/pkg/nvimops/library"
	"devopsmaestro/pkg/nvimops/plugin"
	"devopsmaestro/pkg/nvimops/store"
	"devopsmaestro/pkg/nvimops/theme"
	themelibrary "devopsmaestro/pkg/nvimops/theme/library"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	// Global flags
	configDir string
	outputFmt string
	verbose   bool
	logFile   string
)

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

	// Initialize logging before any command runs
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		initLogging()
	}

	// Add all commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(libraryCmd)
	rootCmd.AddCommand(applyCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(enableCmd)
	rootCmd.AddCommand(disableCmd)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(generateLuaCmd)
	rootCmd.AddCommand(themeCmd)
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

URL Formats:
  https://example.com/plugin.yaml           # Direct URL
  github:user/repo/path/plugin.yaml         # GitHub shorthand (uses raw.githubusercontent.com)
  
Examples:
  nvp apply -f telescope.yaml
  nvp apply -f plugin1.yaml -f plugin2.yaml
  nvp apply --url https://raw.githubusercontent.com/user/repo/main/plugin.yaml
  nvp apply --url github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml
  cat plugin.yaml | nvp apply -f -`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")
		urls, _ := cmd.Flags().GetStringSlice("url")

		if len(files) == 0 && len(urls) == 0 {
			return fmt.Errorf("must specify at least one file with -f flag or URL with --url flag")
		}

		mgr, err := getManager()
		if err != nil {
			return err
		}
		defer mgr.Close()

		// Process files
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

			if err := applyPluginData(mgr, data, source); err != nil {
				return err
			}
		}

		// Process URLs
		for _, url := range urls {
			data, source, err := fetchURL(url)
			if err != nil {
				return fmt.Errorf("failed to fetch %s: %w", url, err)
			}

			if err := applyPluginData(mgr, data, source); err != nil {
				return err
			}
		}

		return nil
	},
}

// applyPluginData parses YAML data and applies it to the manager
func applyPluginData(mgr *nvimops.Manager, data []byte, source string) error {
	slog.Debug("parsing plugin YAML", "source", source, "bytes", len(data))

	p, err := plugin.ParseYAML(data)
	if err != nil {
		slog.Error("failed to parse YAML", "source", source, "error", err)
		return fmt.Errorf("failed to parse %s: %w", source, err)
	}

	// Check if exists for messaging
	existing, _ := mgr.Get(p.Name)
	action := "created"
	if existing != nil {
		action = "configured"
		slog.Debug("updating existing plugin", "name", p.Name)
	} else {
		slog.Debug("creating new plugin", "name", p.Name)
	}

	if err := mgr.Apply(p); err != nil {
		slog.Error("failed to apply plugin", "name", p.Name, "error", err)
		return fmt.Errorf("failed to apply %s: %w", p.Name, err)
	}

	slog.Info("plugin applied", "name", p.Name, "action", action, "source", source)
	fmt.Printf("✓ Plugin '%s' %s (from %s)\n", p.Name, action, source)
	return nil
}

// fetchURL fetches plugin YAML from a URL, supporting GitHub shorthand
func fetchURL(url string) ([]byte, string, error) {
	originalURL := url

	// Handle GitHub shorthand: github:user/repo/path/file.yaml
	if strings.HasPrefix(url, "github:") {
		path := strings.TrimPrefix(url, "github:")
		// Split into user/repo and the rest
		parts := strings.SplitN(path, "/", 3)
		if len(parts) < 3 {
			return nil, "", fmt.Errorf("invalid github URL format, expected github:user/repo/path/file.yaml")
		}
		user := parts[0]
		repo := parts[1]
		filePath := parts[2]
		url = fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", user, repo, filePath)
		slog.Debug("converted GitHub shorthand", "original", originalURL, "url", url)
	}

	slog.Debug("fetching URL", "url", url)

	// Fetch the URL
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		slog.Error("HTTP request failed", "url", url, "error", err)
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Error("HTTP request returned error", "url", url, "status", resp.StatusCode)
		return nil, "", fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("failed to read response body", "url", url, "error", err)
		return nil, "", err
	}

	slog.Info("fetched URL successfully", "url", url, "bytes", len(data))
	return data, url, nil
}

func init() {
	applyCmd.Flags().StringSliceP("filename", "f", nil, "Plugin YAML file(s) to apply (use '-' for stdin)")
	applyCmd.Flags().StringSlice("url", nil, "Plugin YAML URL(s) to fetch and apply")
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

URL Formats:
  https://example.com/theme.yaml           # Direct URL
  github:user/repo/path/theme.yaml         # GitHub shorthand
  
Examples:
  nvp theme apply -f my-theme.yaml
  nvp theme apply --url github:user/repo/themes/custom.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		files, _ := cmd.Flags().GetStringSlice("filename")
		urls, _ := cmd.Flags().GetStringSlice("url")

		if len(files) == 0 && len(urls) == 0 {
			return fmt.Errorf("must specify at least one file with -f flag or URL with --url flag")
		}

		themeStore := getThemeStore()
		if err := themeStore.Init(); err != nil {
			return err
		}

		// Process files
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

			if err := applyThemeData(themeStore, data, source); err != nil {
				return err
			}
		}

		// Process URLs
		for _, url := range urls {
			data, source, err := fetchURL(url)
			if err != nil {
				return fmt.Errorf("failed to fetch %s: %w", url, err)
			}

			if err := applyThemeData(themeStore, data, source); err != nil {
				return err
			}
		}

		return nil
	},
}

func applyThemeData(themeStore *theme.FileStore, data []byte, source string) error {
	slog.Debug("parsing theme YAML", "source", source, "bytes", len(data))

	t, err := theme.ParseYAML(data)
	if err != nil {
		slog.Error("failed to parse theme YAML", "source", source, "error", err)
		return fmt.Errorf("failed to parse %s: %w", source, err)
	}

	// Check if exists
	existing, _ := themeStore.Get(t.Name)
	action := "created"
	if existing != nil {
		action = "updated"
	}

	if err := themeStore.Save(t); err != nil {
		slog.Error("failed to save theme", "name", t.Name, "error", err)
		return fmt.Errorf("failed to save theme: %w", err)
	}

	slog.Info("theme applied", "name", t.Name, "action", action, "source", source)
	fmt.Printf("✓ Theme '%s' %s (from %s)\n", t.Name, action, source)
	return nil
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
	themeCmd.AddCommand(themeDeleteCmd)
	themeCmd.AddCommand(themeUseCmd)
	themeCmd.AddCommand(themeLibraryCmd)
	themeCmd.AddCommand(themeGenerateCmd)

	// Theme library subcommands
	themeLibraryCmd.AddCommand(themeLibraryListCmd)
	themeLibraryCmd.AddCommand(themeLibraryShowCmd)
	themeLibraryCmd.AddCommand(themeLibraryInstallCmd)
	themeLibraryCmd.AddCommand(themeLibraryCategoriesCmd)

	// Flags
	themeListCmd.Flags().StringP("output", "o", "table", "Output format: table, yaml, json")
	themeGetCmd.Flags().StringP("output", "o", "yaml", "Output format: yaml, json")
	themeApplyCmd.Flags().StringSliceP("filename", "f", nil, "Theme YAML file(s)")
	themeApplyCmd.Flags().StringSlice("url", nil, "Theme YAML URL(s)")
	themeDeleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation")
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
