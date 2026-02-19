# Shell Completion for DevOpsMaestro

!!! warning "Deprecated Documentation Location"
    
    **This file has been moved to the proper location.**
    
    **New location:** [Configuration/Shell Completion](configuration/shell-completion.md)
    
    This file is kept for backward compatibility. Please use the configuration version for consistency with other configuration-related documentation.

---

This guide covers shell completion for both tools:
- **dvm** - DevOpsMaestro workspace/container management
- **nvp** - NvimOps Neovim plugin & theme manager

Both support shell autocompletion for **bash**, **zsh**, **fish**, and **PowerShell**.

## Features

- ✅ **Command completion** - Tab-complete all commands
- ✅ **Flag completion** - Tab-complete all flags
- ✅ **Dynamic completions** - Smart suggestions for plugin/theme names
- ✅ **Descriptions** - See helpful descriptions for each option

---

## Quick Install

### dvm Completions

#### macOS (Zsh)
```bash
dvm completion zsh > $(brew --prefix)/share/zsh/site-functions/_dvm
exec zsh
```

#### macOS (Bash)
```bash
brew install bash-completion@2
dvm completion bash > $(brew --prefix)/etc/bash_completion.d/dvm
exec bash
```

#### Linux
```bash
# Zsh
dvm completion zsh > "${fpath[1]}/_dvm"
exec zsh

# Bash
sudo dvm completion bash > /etc/bash_completion.d/dvm
exec bash
```

### nvp Completions

#### macOS (Zsh)
```bash
nvp completion zsh > $(brew --prefix)/share/zsh/site-functions/_nvp
exec zsh
```

#### macOS (Bash)
```bash
nvp completion bash > $(brew --prefix)/etc/bash_completion.d/nvp
exec bash
```

#### Linux
```bash
# Zsh
nvp completion zsh > "${fpath[1]}/_nvp"
exec zsh

# Bash
sudo nvp completion bash > /etc/bash_completion.d/nvp
exec bash
```

#### Fish
```bash
dvm completion fish > ~/.config/fish/completions/dvm.fish
nvp completion fish > ~/.config/fish/completions/nvp.fish
fish_config reload
```

---

## Testing Completion

### dvm Completions
```bash
dvm <TAB>
# Shows: admin attach build create delete get nvim plugin app version workspace

dvm get <TAB>
# Shows: platforms plugins apps workspaces
```

### nvp Completions
```bash
nvp <TAB>
# Shows: apply completion delete disable enable generate get init library list theme version

nvp library <TAB>
# Shows: categories info install list show tags

nvp theme <TAB>
# Shows: apply delete generate get library list use

nvp theme library <TAB>
# Shows: install list show
```

---

## Available Completions

### Commands
- `dvm <TAB>` - Show all available commands
- `dvm nvim <TAB>` - Show nvim subcommands (init, status, sync, push)
- `dvm get <TAB>` - Show get subcommands (plugins, apps, workspaces)

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

Shell completion is **built-in and automatic** with both tools. Just run:

```bash
# One-line install for both (macOS zsh)
dvm completion zsh > $(brew --prefix)/share/zsh/site-functions/_dvm && \
nvp completion zsh > $(brew --prefix)/share/zsh/site-functions/_nvp && \
exec zsh

# One-line install for both (Linux zsh)
dvm completion zsh > "${fpath[1]}/_dvm" && \
nvp completion zsh > "${fpath[1]}/_nvp" && \
exec zsh

# One-line install for both (bash)
dvm completion bash > ~/.local/share/bash-completion/completions/dvm && \
nvp completion bash > ~/.local/share/bash-completion/completions/nvp && \
exec bash
```

Then enjoy tab-completion for all commands, flags, and arguments!
