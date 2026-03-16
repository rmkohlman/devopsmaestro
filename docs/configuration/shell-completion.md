# Shell Completion for DevOpsMaestro

This guide covers shell completion for both tools:
- **dvm** - DevOpsMaestro workspace/container management
- **nvp** - NvimOps Neovim plugin & theme manager

Both support shell autocompletion for **bash**, **zsh**, **fish**, and **PowerShell**.

## Features

- **Command completion** - Tab-complete all commands and subcommands
- **Flag completion** - Tab-complete all flags with descriptions
- **Dynamic completions** - Context-aware suggestions sourced from MaestroVault (your database)
- **Descriptions** - Inline descriptions for commands, flags, and resource names
- **Full resource coverage** - Completions for all resource types: ecosystems, domains, apps, workspaces, credentials, registries, git repos, nvim resources, and terminal resources

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
# Shows: admin attach build create delete get nvim registry rollout start stop terminal version

dvm get <TAB>
# Shows: app apps credential credentials domain domains ecosystem ecosystems gitrepo gitrepos registry registries workspace workspaces

dvm get ecosystem <TAB>
# Shows:  (ecosystem names from MaestroVault, with descriptions)
# platform-a  -- Primary cloud platform ecosystem
# platform-b  -- Secondary data pipeline ecosystem

dvm get workspace <TAB>
# Shows: (workspace names from MaestroVault, with descriptions)
# dev-workspace   -- Feature development workspace
# prod-workspace  -- Production workspace
```

### nvp Completions
```bash
nvp <TAB>
# Shows: apply completion config delete disable enable generate generate-lua get library list source theme version

nvp library <TAB>
# Shows: categories install list show tags

nvp theme <TAB>
# Shows: create delete generate get library list preview use

nvp theme library <TAB>
# Shows: categories install list show tags
```

---

## Available Completions

### Commands with Dynamic Argument Completions

The following commands complete positional arguments from your MaestroVault database:

| Command | Completes |
|---------|-----------|
| `dvm get ecosystem <TAB>` | Ecosystem names with description |
| `dvm use ecosystem <TAB>` | Ecosystem names |
| `dvm delete ecosystem <TAB>` | Ecosystem names |
| `dvm get domain <TAB>` | Domain names with description |
| `dvm use domain <TAB>` | Domain names |
| `dvm delete domain <TAB>` | Domain names |
| `dvm get app <TAB>` | App names with description |
| `dvm use app <TAB>` | App names |
| `dvm delete app <TAB>` | App names |
| `dvm get workspace <TAB>` | Workspace names with description |
| `dvm use workspace <TAB>` | Workspace names |
| `dvm delete workspace <TAB>` | Workspace names |
| `dvm get credential <TAB>` | Credential names |
| `dvm delete credential <TAB>` | Credential names |
| `dvm get registry <TAB>` | Registry names |
| `dvm delete registry <TAB>` | Registry names |
| `dvm start <TAB>` | Registry names |
| `dvm stop <TAB>` | Registry names |
| `dvm rollout restart <TAB>` | Registry names |
| `dvm rollout status <TAB>` | Registry names |
| `dvm rollout history <TAB>` | Registry names |
| `dvm rollout undo <TAB>` | Registry names |
| `dvm get gitrepo <TAB>` | Git repo names with URL |
| `dvm delete gitrepo <TAB>` | Git repo names |
| `dvm sync gitrepo <TAB>` | Git repo names |
| `dvm nvim get plugin <TAB>` | NvimPlugin names |
| `dvm nvim set plugin <TAB>` | NvimPlugin names |
| `dvm edit nvim-plugin <TAB>` | NvimPlugin names |
| `dvm delete nvim-plugin <TAB>` | NvimPlugin names |
| `dvm nvim get theme <TAB>` | NvimTheme names |
| `dvm edit nvim-theme <TAB>` | NvimTheme names |
| `dvm delete nvim-theme <TAB>` | NvimTheme names |
| `dvm nvim get package <TAB>` | NvimPackage names |
| `dvm use nvim-package <TAB>` | NvimPackage names |
| `dvm terminal get package <TAB>` | TerminalPackage names |
| `dvm terminal set package <TAB>` | TerminalPackage names |
| `dvm use terminal-package <TAB>` | TerminalPackage names |
| `dvm terminal set prompt <TAB>` | TerminalPrompt names |
| `dvm set theme <TAB>` | NvimTheme names |
| `dvm nvim sync <TAB>` | Workspace names |
| `dvm nvim push <TAB>` | Workspace names |

### Commands with Static Argument Completions

| Command | Completes |
|---------|-----------|
| `dvm nvim init <TAB>` | Template names (kickstart, lazyvim, astronvim, etc.) |
| `dvm registry enable <TAB>` | Registry types (oci, pypi, npm, go, http) |
| `dvm registry disable <TAB>` | Registry types (oci, pypi, npm, go, http) |

### Commands with Multi-Argument Completions

| Command | First Arg | Second Arg |
|---------|-----------|------------|
| `dvm registry set-default <TAB>` | Registry type (oci, pypi, npm, go, http) | Registry name from database |

### Flags with Dynamic Completions

| Flag | Completes | Used By |
|------|-----------|---------|
| `--ecosystem` | Ecosystem names | get domains, get domain, delete domain, create domain, create credential, get credentials, get credential, delete credential, set theme, attach, build, detach, get workspaces, get workspace |
| `--domain` | Domain names | get apps, get app, delete app, create app, set theme, attach, build, detach, get workspaces, get workspace |
| `--app` | App names | delete workspace, create workspace, nvim get plugins, nvim get plugin, set nvim-plugin, delete nvim-plugin, delete nvim-theme, terminal set prompt, terminal set plugin, terminal set package, create branch, set theme, attach, build, detach, get workspaces, get workspace |
| `--workspace` | Workspace names | nvim get plugins, nvim get plugin, set nvim-plugin, delete nvim-plugin, delete nvim-theme, terminal set prompt, terminal set plugin, terminal set package, create branch, attach, build, detach, get workspaces, get workspace |
| `--repo` | Git repo names with URL | create app, create workspace |
| `--credential` | Credential names | create gitrepo |

---

## Dynamic Completions

DevOpsMaestro provides **dynamic completions** that query MaestroVault for context-aware suggestions. All resource types are supported as of v0.42.0.

### Core Hierarchy (Ecosystems, Domains, Apps, Workspaces)

Completions for the core resource hierarchy include descriptions where available:

```bash
dvm get ecosystem <TAB>
# platform-a  -- Primary cloud platform ecosystem
# platform-b  -- Secondary data pipeline ecosystem

dvm use domain <TAB>
# backend   -- Backend services domain
# frontend  -- Frontend applications domain

dvm delete app <TAB>
# api-server   -- REST API service
# web-client   -- React web client

dvm get workspace <TAB>
# dev-workspace   -- Feature development workspace
# prod-workspace  -- Production workspace
```

### Credentials and Registries

```bash
dvm get credential <TAB>
# aws-prod-creds
# github-token
# dockerhub-ro

dvm start <TAB>
# local-oci    -- Local OCI registry
# npm-cache    -- Cached npm proxy

dvm rollout restart <TAB>
# local-oci
# npm-cache
```

### Git Repos

Git repo completions include the repository URL as the description:

```bash
dvm get gitrepo <TAB>
# my-app-repo   -- https://github.com/org/my-app
# config-repo   -- https://github.com/org/config

dvm sync gitrepo <TAB>
# my-app-repo
# config-repo
```

### Nvim Resources

```bash
dvm nvim get plugin <TAB>
# telescope.nvim
# nvim-treesitter
# lualine.nvim

dvm nvim get theme <TAB>
# catppuccin-mocha
# tokyo-night
# nord

dvm use nvim-package <TAB>
# base
# full
# minimal
```

### Terminal Resources

```bash
dvm terminal get package <TAB>
# core-utils
# dev-tools

dvm terminal set prompt <TAB>
# starship
# oh-my-posh
```

### Registry Types (Static)

Registry type completions are static — they do not query the database:

```bash
dvm registry enable <TAB>
# oci    -- OCI (Docker) registry
# pypi   -- Python package index
# npm    -- Node package registry
# go     -- Go module proxy
# http   -- Generic HTTP registry

dvm registry disable <TAB>
# oci
# pypi
# npm
# go
# http
```

### Multi-Argument Completions (registry set-default)

The `registry set-default` command uses position-aware completion. The first argument completes registry types; the second argument completes registry names from the database filtered to the selected type:

```bash
dvm registry set-default <TAB>
# oci
# pypi
# npm
# go
# http

dvm registry set-default oci <TAB>
# local-oci
# remote-oci
```

### Template Names (Static)

```bash
dvm nvim init <TAB>
# kickstart  -- Minimal, well-documented starter config
# lazyvim    -- Feature-rich, batteries-included config
# astronvim  -- Aesthetically pleasing, fully featured config
```

### Flag Completions

Flags complete resource names from MaestroVault regardless of the command position:

```bash
dvm create workspace --app <TAB>
# api-server
# web-client

dvm get workspaces --ecosystem <TAB>
# platform-a
# platform-b

dvm create app --repo <TAB>
# my-app-repo  -- https://github.com/org/my-app
# config-repo  -- https://github.com/org/config

dvm create gitrepo --credential <TAB>
# aws-prod-creds
# github-token
```

---

## Interactive Menu Selection (Zsh)

By default, zsh lists completion candidates without interactive navigation. To enable a highlighted menu that can be navigated with arrow keys, add the following to your `~/.zshrc`:

```bash
# Enable interactive menu selection for tab completion
zstyle ':completion:*' menu select
```

After adding this line, reload your shell:

```bash
exec zsh
```

With `menu select` enabled, pressing `<TAB>` opens a highlighted list. Use the arrow keys to move through candidates and press `<Return>` to select. Press `<TAB>` again or `<Esc>` to dismiss the menu.

This setting affects all zsh completions system-wide, not just DevOpsMaestro.

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

### Dynamic completions return nothing

If resource completions (e.g. workspace names, ecosystem names) show no results, MaestroVault may not have any resources of that type yet, or the database path may not be configured. Verify with:

```bash
dvm get ecosystems
dvm get workspaces
```

If those commands return results but completions are empty, regenerate the completion script — an outdated script may be missing the dynamic completion functions:

```bash
dvm completion zsh > $(brew --prefix)/share/zsh/site-functions/_dvm
exec zsh
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

Then tab-complete all commands, flags, and resource names directly from MaestroVault.
