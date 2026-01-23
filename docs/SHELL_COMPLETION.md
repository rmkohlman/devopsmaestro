# Shell Completion for DevOpsMaestro (dvm)

DevOpsMaestro supports shell autocompletion for all commands, flags, and arguments across **bash**, **zsh**, **fish**, and **PowerShell**.

## Features

- ✅ **Command completion** - Tab-complete all dvm commands
- ✅ **Flag completion** - Tab-complete all flags (--help, --config-path, etc.)
- ✅ **Dynamic completions** - Smart suggestions for workspace names, templates, etc.
- ✅ **Descriptions** - See helpful descriptions for each completion option

---

## Quick Install

### macOS

#### Zsh (default on macOS)
```bash
# Install completion script
dvm completion zsh > $(brew --prefix)/share/zsh/site-functions/_dvm

# Restart your shell
exec zsh
```

#### Bash (if using bash)
```bash
# Install bash-completion if not already installed
brew install bash-completion@2

# Install dvm completion
dvm completion bash > $(brew --prefix)/etc/bash_completion.d/dvm

# Restart your shell
exec bash
```

### Linux

#### Zsh
```bash
# Install completion script
dvm completion zsh > "${fpath[1]}/_dvm"

# Restart your shell
exec zsh
```

#### Bash
```bash
# Install to system completions directory
sudo dvm completion bash > /etc/bash_completion.d/dvm

# Or user-local (no sudo)
mkdir -p ~/.local/share/bash-completion/completions
dvm completion bash > ~/.local/share/bash-completion/completions/dvm

# Restart your shell
exec bash
```

#### Fish
```bash
# Install completion script
dvm completion fish > ~/.config/fish/completions/dvm.fish

# Reload fish config
fish_config reload
```

### Windows (PowerShell)

```powershell
# Generate completion script
dvm completion powershell > $HOME\Documents\PowerShell\Scripts\dvm-completion.ps1

# Add to your PowerShell profile
Add-Content $PROFILE ". $HOME\Documents\PowerShell\Scripts\dvm-completion.ps1"

# Restart PowerShell
```

---

## Manual Installation

If the quick install doesn't work, you can manually source the completion script:

### Zsh

Add to your `~/.zshrc`:

```bash
# Enable completion system
autoload -U compinit; compinit

# Load dvm completions
source <(dvm completion zsh)
```

### Bash

Add to your `~/.bashrc`:

```bash
# Load dvm completions
source <(dvm completion bash)
```

### Fish

Add to your `~/.config/fish/config.fish`:

```fish
# Load dvm completions
dvm completion fish | source
```

---

## Testing Completion

After installation, test that completion works:

```bash
# Type this and press TAB
dvm n<TAB>

# Should complete to:
dvm nvim

# Type this and press TAB
dvm nvim init <TAB>

# Should show:
kickstart  -- Minimal, well-documented starter config
lazyvim    -- Feature-rich, batteries-included config
astronvim  -- Aesthetically pleasing, fully featured config
minimal    -- Minimal config created by DevOpsMaestro
custom     -- Clone from custom Git URL
```

---

## Available Completions

### Commands
- `dvm <TAB>` - Show all available commands
- `dvm nvim <TAB>` - Show nvim subcommands (init, status, sync, push)
- `dvm get <TAB>` - Show get subcommands (plugins, projects, workspaces)

### Flags
- `dvm nvim init --<TAB>` - Show all flags (--config-path, --git-clone, --overwrite)
- `dvm nvim sync --<TAB>` - Show sync flags (--remote-wins)

### Arguments
- `dvm nvim init <TAB>` - Show template options (kickstart, lazyvim, etc.)
- `dvm nvim sync <TAB>` - Show workspace names (dynamic from database)
- `dvm nvim push <TAB>` - Show workspace names (dynamic from database)

---

## Dynamic Completions

DevOpsMaestro provides **dynamic completions** that query the database for context-aware suggestions:

### Workspace Names
When typing workspace-related commands, completion will suggest actual workspace names:

```bash
dvm nvim sync <TAB>
# Shows: my-workspace  dev-workspace  prod-workspace

dvm attach <TAB>
# Shows: my-workspace  dev-workspace  prod-workspace
```

### Template Names
Template suggestions include descriptions:

```bash
dvm nvim init <TAB>
# Shows:
# kickstart  -- Minimal, well-documented starter config
# lazyvim    -- Feature-rich, batteries-included config
# astronvim  -- Aesthetically pleasing, fully featured config
```

---

## Troubleshooting

### Completion not working

**1. Check completion is installed:**
```bash
# Zsh
ls -la $(brew --prefix)/share/zsh/site-functions/_dvm

# Bash
ls -la $(brew --prefix)/etc/bash_completion.d/dvm
```

**2. Verify shell completion is enabled:**

Zsh:
```bash
# Should show completion functions
echo $fpath
```

Bash:
```bash
# Check if bash-completion is loaded
complete -p | grep dvm
```

**3. Regenerate completion:**
```bash
# Remove old completion
rm $(brew --prefix)/share/zsh/site-functions/_dvm

# Regenerate
dvm completion zsh > $(brew --prefix)/share/zsh/site-functions/_dvm

# Restart shell
exec zsh
```

### Completion shows file paths

If completion suggests files instead of commands, your shell may be falling back to file completion. This usually means:

1. Completion script isn't installed
2. Completion script isn't loaded
3. Shell needs restart

**Fix:**
```bash
# Reinstall completion
dvm completion <your-shell> > <install-location>

# Restart shell
exec $SHELL
```

### Descriptions not showing (zsh)

Zsh can disable descriptions. Check your `.zshrc`:

```bash
# Ensure this is NOT in your .zshrc:
# zstyle ':completion:*' verbose false

# Or explicitly enable descriptions:
zstyle ':completion:*' verbose yes
zstyle ':completion:*:descriptions' format '%B%d%b'
```

---

## Advanced: Custom Completions

If you're a developer extending DevOpsMaestro, you can add custom completions:

### In Go code (cmd/*.go)

```go
// Add dynamic completion function
myCommand.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    // Query database or API
    items := []string{
        "option1\tDescription of option 1",
        "option2\tDescription of option 2",
    }
    return items, cobra.ShellCompDirectiveNoFileComp
}
```

### Register flag completions

```go
// Complete file paths
cmd.MarkFlagFilename("config", "yaml", "yml")

// Complete directory paths
cmd.MarkFlagDirname("output-dir")

// Custom flag values
cmd.RegisterFlagCompletionFunc("theme", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
    return []string{"catppuccin-mocha", "tokyo-night", "nord"}, cobra.ShellCompDirectiveDefault
})
```

---

## References

- Cobra Shell Completion: https://github.com/spf13/cobra/blob/main/shell_completions.md
- Zsh Completion System: https://zsh.sourceforge.io/Doc/Release/Completion-System.html
- Bash Programmable Completion: https://www.gnu.org/software/bash/manual/html_node/Programmable-Completion.html
- Fish Completion: https://fishshell.com/docs/current/completions.html

---

## Summary

Shell completion is **built-in and automatic** with DevOpsMaestro. Just run:

```bash
# One-line install (macOS zsh)
dvm completion zsh > $(brew --prefix)/share/zsh/site-functions/_dvm && exec zsh

# One-line install (Linux zsh)
dvm completion zsh > "${fpath[1]}/_dvm" && exec zsh

# One-line install (bash)
dvm completion bash > ~/.local/share/bash-completion/completions/dvm && exec bash
```

Then enjoy tab-completion for all commands, flags, and arguments! 🎉
