package main

import (
	"github.com/spf13/cobra"
)

// =============================================================================
// COMPLETION COMMAND
// =============================================================================

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate the autocompletion script for nvp for the specified shell.

To load completions in your current shell session:

  source <(nvp completion bash)   # Bash
  source <(nvp completion zsh)    # Zsh
  nvp completion fish | source    # Fish

To install completions permanently:

  # Bash (Linux)
  nvp completion bash > /etc/bash_completion.d/nvp

  # Bash (macOS with Homebrew)
  nvp completion bash > $(brew --prefix)/etc/bash_completion.d/nvp

  # Zsh (macOS with Homebrew)
  nvp completion zsh > $(brew --prefix)/share/zsh/site-functions/_nvp

  # Zsh (Linux)
  nvp completion zsh > "${fpath[1]}/_nvp"

  # Fish
  nvp completion fish > ~/.config/fish/completions/nvp.fish

  # PowerShell
  nvp completion powershell > nvp.ps1
  # Then add '. nvp.ps1' to your PowerShell profile

You will need to start a new shell for permanent installations to take effect.`,
	Args:      cobra.ExactValidArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(out)
		case "zsh":
			return rootCmd.GenZshCompletion(out)
		case "fish":
			return rootCmd.GenFishCompletion(out, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(out)
		}
		return nil
	},
}
