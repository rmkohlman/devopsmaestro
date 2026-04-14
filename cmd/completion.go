package cmd

import (
	"bytes"
	"strings"

	"github.com/spf13/cobra"
)

// completionCmd is the custom completion command for dvm.
// It replaces Cobra's default completion command to fix zsh autoload
// compatibility. Cobra's GenZshCompletion emits a bare "compdef _dvm dvm"
// line that breaks zsh fpath-based autoloading (the function is not yet
// defined when compinit scans the file). Our wrapper strips that line so
// only the #compdef header remains, which is the correct zsh convention.
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion script",
	Long: `Generate the autocompletion script for dvm for the specified shell.

To load completions in your current shell session:

  source <(dvm completion bash)   # Bash
  source <(dvm completion zsh)    # Zsh
  dvm completion fish | source    # Fish

To install completions permanently:

  # Bash (Linux)
  dvm completion bash > /etc/bash_completion.d/dvm

  # Bash (macOS with Homebrew)
  dvm completion bash > $(brew --prefix)/etc/bash_completion.d/dvm

  # Zsh (macOS with Homebrew)
  dvm completion zsh > $(brew --prefix)/share/zsh/site-functions/_dvm

  # Zsh (Linux)
  dvm completion zsh > "${fpath[1]}/_dvm"

  # Fish
  dvm completion fish > ~/.config/fish/completions/dvm.fish

  # PowerShell
  dvm completion powershell > dvm.ps1
  # Then add '. dvm.ps1' to your PowerShell profile

You will need to start a new shell for permanent installations to take effect.`,
	Args:      cobra.ExactValidArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(out)
		case "zsh":
			return genZshCompletionFixed(cmd)
		case "fish":
			return rootCmd.GenFishCompletion(out, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletionWithDesc(out)
		}
		return nil
	},
}

// genZshCompletionFixed generates zsh completion output with the bare
// "compdef _dvm dvm" line removed. Cobra emits this line right after the
// #compdef header, but it is incompatible with zsh's fpath autoload
// mechanism: when compinit scans the file, only the #compdef header is
// needed to register the function. The bare compdef call can prevent the
// completion function from being properly autoloaded, causing "type _dvm
// → not found" and broken tab completion that requires a shell restart.
func genZshCompletionFixed(cmd *cobra.Command) error {
	var buf bytes.Buffer
	if err := rootCmd.GenZshCompletion(&buf); err != nil {
		return err
	}

	// Remove the bare "compdef _dvm dvm" line. We use line-by-line
	// filtering rather than a simple string replace so we only strip
	// exact matches and keep the #compdef header intact.
	out := cmd.OutOrStdout()
	raw := buf.String()
	lines := strings.Split(raw, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(line, "compdef _") {
			continue
		}
		filtered = append(filtered, line)
	}
	_, err := out.Write([]byte(strings.Join(filtered, "\n")))
	return err
}

// init for completion registration is in zz_completion_init.go to ensure
// it runs AFTER all command init() functions have registered their flags.
// Go processes init() functions in file-name order within a package;
// completion.go (starting with 'c') would run before files like get.go,
// set_build_arg.go, etc., causing RegisterFlagCompletionFunc to silently
// fail because the flags don't exist yet.

// registerDynamicCompletions registers custom completion functions
func registerDynamicCompletions() {
	// Complete template names for 'dvm nvim init'
	templateCompletion := func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		templates := []string{
			"kickstart\tMinimal, well-documented starter config",
			"lazyvim\tFeature-rich, batteries-included config",
			"astronvim\tAesthetically pleasing, fully featured config",
			"minimal\tMinimal config created by DevOpsMaestro",
			"custom\tClone from custom Git URL",
		}
		return templates, cobra.ShellCompDirectiveNoFileComp
	}

	// Register completions for nvim commands
	if nvimInitCmd != nil {
		nvimInitCmd.ValidArgsFunction = templateCompletion
	}

	if nvimSyncCmd != nil {
		nvimSyncCmd.ValidArgsFunction = completeWorkspaces
	}

	if nvimPushCmd != nil {
		nvimPushCmd.ValidArgsFunction = completeWorkspaces
	}

	// Register all resource completions (ecosystem, domain, app, workspace)
	// This wires up ValidArgsFunction for get/use/delete commands
	registerAllResourceCompletions()
}
