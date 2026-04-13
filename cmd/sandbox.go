package cmd

import (
	"github.com/rmkohlman/MaestroSDK/render"
	"github.com/spf13/cobra"
)

// sandboxFlags holds the flags for sandbox create.
var sandboxFlags struct {
	version string
	deps    string
	name    string
	repo    string
	noCache bool
}

// sandboxDeleteAll controls the --all flag on sandbox delete.
var sandboxDeleteAll bool

// sandboxCmd is the parent command for sandbox management.
var sandboxCmd = &cobra.Command{
	Use:     "sandbox [language]",
	Aliases: []string{"sb"},
	Short:   "Manage ephemeral sandbox workspaces",
	Long: `Create and manage ephemeral, single-use development containers for quick
testing, prototyping, or learning — no ecosystem/domain/app hierarchy required.

When invoked with a language argument (e.g., dvm sandbox python), it is a
shorthand for "dvm sandbox create <lang>".

Supported languages: python, golang, rust, node, cpp
(with aliases: py, go, rs, js, c++, etc.)

Examples:
  # Quick sandbox (shorthand for create)
  dvm sandbox python
  dvm sandbox golang --version 1.24

  # Explicit create
  dvm sandbox create node --version 22

  # List active sandboxes
  dvm sandbox get

  # Attach to a running sandbox
  dvm sandbox attach my-sandbox

  # Delete all sandboxes
  dvm sandbox delete --all`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return cmd.Help()
		}
		// Shorthand: dvm sandbox <lang> → dvm sandbox create <lang>
		return runSandboxCreate(cmd, args[0])
	},
}

// sandboxCreateCmd creates a new sandbox.
var sandboxCreateCmd = &cobra.Command{
	Use:   "create <language>",
	Short: "Create an ephemeral sandbox for a language runtime",
	Long: `Create and attach to an ephemeral development container.

On exit, the container is automatically stopped and removed.
The sandbox image is cached for fast re-creation.

Examples:
  dvm sandbox create python
  dvm sandbox create golang --version 1.24
  dvm sandbox create node --deps package.json
  dvm sandbox create rust --name my-rust-test
  dvm sandbox create python --repo https://github.com/user/project.git`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSandboxCreate(cmd, args[0])
	},
}

// sandboxGetCmd lists active sandboxes.
var sandboxGetCmd = &cobra.Command{
	Use:   "get",
	Short: "List active sandboxes",
	Long: `List all active sandbox containers.

Sandboxes are discovered by querying the container runtime for containers
with the dvm.sandbox=true label. No database records are used.

Examples:
  dvm sandbox get
  dvm sandbox get -o json
  dvm sandbox get -o yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSandboxGet(cmd)
	},
}

// sandboxAttachCmd re-attaches to a running sandbox.
var sandboxAttachCmd = &cobra.Command{
	Use:   "attach <name>",
	Short: "Attach to a running sandbox",
	Long: `Attach an interactive terminal to an existing running sandbox.

Unlike "sandbox create", this does NOT auto-remove the container on exit,
so you can re-attach multiple times.

Examples:
  dvm sandbox attach dvm-sandbox-python-a3f2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSandboxAttach(cmd, args[0])
	},
}

// sandboxDeleteCmd removes sandboxes.
var sandboxDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete one or all sandboxes",
	Long: `Delete a specific sandbox by name, or all sandboxes with --all.

Examples:
  dvm sandbox delete dvm-sandbox-python-a3f2
  dvm sandbox delete --all`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if sandboxDeleteAll {
			return runSandboxDeleteAll(cmd)
		}
		if len(args) == 0 {
			render.Error("Specify a sandbox name or use --all")
			return errSilent
		}
		return runSandboxDelete(cmd, args[0])
	},
}

func init() {
	rootCmd.AddCommand(sandboxCmd)

	// Subcommands
	sandboxCmd.AddCommand(sandboxCreateCmd)
	sandboxCmd.AddCommand(sandboxGetCmd)
	sandboxCmd.AddCommand(sandboxAttachCmd)
	sandboxCmd.AddCommand(sandboxDeleteCmd)

	// Flags on create
	for _, cmd := range []*cobra.Command{sandboxCmd, sandboxCreateCmd} {
		cmd.Flags().StringVar(&sandboxFlags.version, "version", "", "Language version (interactive picker if omitted)")
		cmd.Flags().StringVarP(&sandboxFlags.deps, "deps", "d", "", "Path to dependency file (requirements.txt, package.json, etc.)")
		cmd.Flags().StringVarP(&sandboxFlags.name, "name", "n", "", "Custom sandbox name (default: auto-generated)")
		cmd.Flags().StringVarP(&sandboxFlags.repo, "repo", "r", "", "Git repo URL to clone into sandbox")
		cmd.Flags().BoolVar(&sandboxFlags.noCache, "no-cache", false, "Force rebuild the sandbox image")
	}

	// Flags on delete
	sandboxDeleteCmd.Flags().BoolVar(&sandboxDeleteAll, "all", false, "Delete all sandboxes")

	// Skip auto-migration for sandbox commands (no DB needed)
}
