package cmd

import (
	"devopsmaestro/nvim"
	"devopsmaestro/ui"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var nvimCmd = &cobra.Command{
	Use:   "nvim",
	Short: "Manage local Neovim configuration",
	Long: `Manage local Neovim configuration and sync with workspace containers.

DevOpsMaestro's nvim commands help you:
  • Initialize local Neovim config from popular templates
  • Sync config between your local machine and workspace containers
  • Keep your Neovim setup consistent across environments

Examples:
  # Initialize with kickstart.nvim
  dvm nvim init kickstart

  # Initialize with LazyVim
  dvm nvim init lazyvim --git-clone

  # Check status
  dvm nvim status

  # Sync from workspace to local
  dvm nvim sync my-workspace

  # Push local changes to workspace
  dvm nvim push my-workspace`,
}

var nvimInitCmd = &cobra.Command{
	Use:   "init [template]",
	Short: "Initialize local Neovim configuration from template",
	Long: `Initialize your local Neovim configuration from a template.

Available templates:
  kickstart  - kickstart.nvim (minimal, well-documented starter)
  lazyvim    - LazyVim (feature-rich, batteries-included)
  astronvim  - AstroNvim (aesthetically pleasing, fully featured)
  minimal    - Minimal config created by DevOpsMaestro
  custom     - Clone from a custom Git URL (requires --git-url)
  
  Or use a direct URL:
  https://github.com/user/nvim-config - Clone from any Git repository

By default, creates a minimal config. Use --git-clone to clone from upstream repos.

Examples:
  # Create minimal config
  dvm nvim init minimal

  # Clone kickstart.nvim
  dvm nvim init kickstart --git-clone

  # Clone LazyVim starter
  dvm nvim init lazyvim --git-clone

  # Clone from GitHub URL directly
  dvm nvim init https://github.com/yourusername/nvim-config.git

  # Clone from short GitHub format
  dvm nvim init github:yourusername/nvim-config

  # Clone subdirectory from repo
  dvm nvim init https://github.com/user/repo.git --subdir templates/starter

  # Overwrite existing config
  dvm nvim init kickstart --git-clone --overwrite`,
	Args: cobra.ExactArgs(1),
	Run:  runNvimInit,
}

var nvimStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show local Neovim configuration status",
	Long: `Display the current status of your local Neovim configuration.

Shows:
  • Config location
  • Whether config exists
  • Last sync time
  • Workspace last synced with
  • Local changes since last sync
  • Template used

Example:
  dvm nvim status`,
	Run: runNvimStatus,
}

var nvimSyncCmd = &cobra.Command{
	Use:   "sync <workspace>",
	Short: "Sync local config with workspace (pull from workspace)",
	Long: `Synchronize your local Neovim config with a workspace container.

By default, this pulls config FROM the workspace TO your local machine.
Use 'dvm nvim push' to push local changes to the workspace.

Examples:
  # Pull config from workspace
  dvm nvim sync my-workspace

  # Pull with automatic conflict resolution (remote wins)
  dvm nvim sync my-workspace --remote-wins`,
	Args: cobra.ExactArgs(1),
	Run:  runNvimSync,
}

var nvimPushCmd = &cobra.Command{
	Use:   "push <workspace>",
	Short: "Push local config to workspace",
	Long: `Push your local Neovim configuration to a workspace container.

This copies your local config TO the workspace container, overwriting
the workspace's config.

Examples:
  # Push local config to workspace
  dvm nvim push my-workspace

  # Push and restart Neovim in workspace
  dvm nvim push my-workspace --restart`,
	Args: cobra.ExactArgs(1),
	Run:  runNvimPush,
}

// Flags
var (
	nvimConfigPath string
	nvimGitClone   bool
	nvimGitURL     string
	nvimSubdir     string
	nvimOverwrite  bool
	nvimRemoteWins bool
	nvimRestart    bool
)

func init() {
	// Add subcommands
	nvimCmd.AddCommand(nvimInitCmd)
	nvimCmd.AddCommand(nvimStatusCmd)
	nvimCmd.AddCommand(nvimSyncCmd)
	nvimCmd.AddCommand(nvimPushCmd)

	// Add to root command
	rootCmd.AddCommand(nvimCmd)

	// Init flags
	nvimInitCmd.Flags().StringVar(&nvimConfigPath, "config-path", "", "Custom config path (default: ~/.config/nvim)")
	nvimInitCmd.Flags().BoolVar(&nvimGitClone, "git-clone", false, "Clone template from upstream Git repository")
	nvimInitCmd.Flags().StringVar(&nvimGitURL, "git-url", "", "Custom Git URL (for 'custom' template)")
	nvimInitCmd.Flags().StringVar(&nvimSubdir, "subdir", "", "Subdirectory within repo to use as config root")
	nvimInitCmd.Flags().BoolVar(&nvimOverwrite, "overwrite", false, "Overwrite existing config")

	// Sync flags
	nvimSyncCmd.Flags().BoolVar(&nvimRemoteWins, "remote-wins", false, "Remote changes win in conflicts")

	// Push flags
	nvimPushCmd.Flags().BoolVar(&nvimRestart, "restart", false, "Restart Neovim in workspace after push")

	// Register custom completions
	registerNvimCompletions()
}

// registerNvimCompletions registers custom completion functions for nvim commands
func registerNvimCompletions() {
	// Complete template names for 'dvm nvim init'
	nvimInitCmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// Only complete the first argument (template name)
		if len(args) >= 1 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		templates := []string{
			"kickstart\tMinimal, well-documented starter config",
			"lazyvim\tFeature-rich, batteries-included config",
			"astronvim\tAesthetically pleasing, fully featured config",
			"minimal\tMinimal config created by DevOpsMaestro",
			"custom\tClone from custom Git URL (requires --git-url)",
		}
		return templates, cobra.ShellCompDirectiveNoFileComp
	}

	// Complete workspace names for sync/push commands
	workspaceCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		if len(args) > 0 {
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
		// TODO: Query database for actual workspace names
		// For now, return empty to show this feature exists
		return []string{}, cobra.ShellCompDirectiveDefault
	}

	nvimSyncCmd.ValidArgsFunction = workspaceCompletion
	nvimPushCmd.ValidArgsFunction = workspaceCompletion
}

func runNvimInit(cmd *cobra.Command, args []string) {
	template := args[0]

	fmt.Println()
	fmt.Println(ui.SectionHeader("Initializing Neovim Configuration"))
	fmt.Println()

	mgr := nvim.NewManager()
	if nvimConfigPath != "" {
		mgr = nvim.NewManagerWithPath(nvimConfigPath)
	}

	// Detect if template is a URL
	isURL := nvim.IsGitURL(template)

	opts := nvim.InitOptions{
		ConfigPath: nvimConfigPath,
		Template:   template,
		Overwrite:  nvimOverwrite,
		GitClone:   nvimGitClone || isURL, // Auto-enable git-clone for URLs
		GitURL:     nvimGitURL,
		Subdir:     nvimSubdir,
	}

	// If template is a URL, use it as GitURL
	if isURL {
		opts.GitURL = template
		opts.Template = "custom"
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Source:"), ui.InfoStyle.Render(template))
	} else {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Template:"), ui.InfoStyle.Render(template))
	}

	if opts.GitClone {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Method:"), ui.InfoStyle.Render("Git clone"))
	} else {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Method:"), ui.InfoStyle.Render("Local template"))
	}

	if nvimSubdir != "" {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Subdirectory:"), ui.PathStyle.Render(nvimSubdir))
	}
	fmt.Println()

	if err := mgr.Init(opts); err != nil {
		fmt.Println(ui.ErrorBox(fmt.Sprintf("Failed to initialize: %v", err)))
		os.Exit(1)
	}

	status, _ := mgr.Status()

	fmt.Println(ui.SuccessBox("Neovim configuration initialized successfully!"))
	fmt.Println()
	fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Location:"), ui.PathStyle.Render(status.ConfigPath))
	if isURL {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Source:"), ui.InfoStyle.Render(template))
	} else {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Template:"), ui.InfoStyle.Render(template))
	}
	fmt.Println()
	fmt.Println(ui.MutedStyle.Render("  Next steps:"))
	fmt.Println(ui.MutedStyle.Render("    1. Run 'nvim' to open Neovim"))
	if template != "minimal" && opts.GitClone {
		fmt.Println(ui.MutedStyle.Render("    2. Wait for plugins to install"))
		fmt.Println(ui.MutedStyle.Render("    3. Restart Neovim"))
	}
	fmt.Println()
}

func runNvimStatus(cmd *cobra.Command, args []string) {
	mgr := nvim.NewManager()
	status, err := mgr.Status()
	if err != nil {
		fmt.Println(ui.ErrorBox(fmt.Sprintf("Failed to get status: %v", err)))
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println(ui.SectionHeader("Neovim Configuration Status"))
	fmt.Println()

	fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Config Path:  "), ui.PathStyle.Render(status.ConfigPath))

	if status.Exists {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Status:       "), ui.SuccessStyle.Render("✓ Exists"))
	} else {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Status:       "), ui.ErrorStyle.Render("✗ Not found"))
		fmt.Println()
		fmt.Println(ui.InfoBox("No Neovim config found. Run 'dvm nvim init <template>' to create one."))
		fmt.Println()
		return
	}

	if status.Template != "" {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Template:     "), ui.InfoStyle.Render(status.Template))
	}

	if !status.LastSync.IsZero() {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Last Sync:    "), ui.DateStyle.Render(status.LastSync.Format("2006-01-02 15:04:05")))
	}

	if status.SyncedWith != "" {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Synced With:  "), ui.InfoStyle.Render(status.SyncedWith))
	}

	if status.LocalChanges {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Local Changes:"), ui.WarningStyle.Render("Yes"))
	} else {
		fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Local Changes:"), ui.MutedStyle.Render("No"))
	}

	fmt.Println()
}

func runNvimSync(cmd *cobra.Command, args []string) {
	workspace := args[0]

	fmt.Println()
	fmt.Println(ui.SectionHeader("Syncing Neovim Configuration"))
	fmt.Println()
	fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Workspace:"), ui.InfoStyle.Render(workspace))
	fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Direction:"), ui.InfoStyle.Render("workspace → local (pull)"))
	fmt.Println()

	mgr := nvim.NewManager()
	if err := mgr.Sync(workspace, nvim.SyncPull); err != nil {
		fmt.Println(ui.ErrorBox(fmt.Sprintf("Failed to sync: %v", err)))
		os.Exit(1)
	}

	fmt.Println(ui.SuccessBox("Configuration synced successfully!"))
	fmt.Println()
}

func runNvimPush(cmd *cobra.Command, args []string) {
	workspace := args[0]

	fmt.Println()
	fmt.Println(ui.SectionHeader("Pushing Neovim Configuration"))
	fmt.Println()
	fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Workspace:"), ui.InfoStyle.Render(workspace))
	fmt.Printf("  %s %s\n", ui.MutedStyle.Render("Direction:"), ui.InfoStyle.Render("local → workspace (push)"))
	fmt.Println()

	mgr := nvim.NewManager()
	if err := mgr.Push(workspace); err != nil {
		fmt.Println(ui.ErrorBox(fmt.Sprintf("Failed to push: %v", err)))
		os.Exit(1)
	}

	fmt.Println(ui.SuccessBox("Configuration pushed successfully!"))
	fmt.Println()
}
