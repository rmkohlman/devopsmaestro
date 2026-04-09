# dvt Commands Reference

`dvt` is the DevOpsMaestro Terminal CLI for managing terminal prompts, plugins, packages, shell configs, profiles, emulators, and WezTerm integration.

**Global flags** available on every command:

| Flag | Description |
|------|-------------|
| `--config <dir>` | Config directory (default: `~/.dvt`) |
| `-v, --verbose` | Enable debug logging |
| `--log-file <path>` | Write logs to file (JSON format) |
| `--no-color` | Disable colored output |

---

## Getting Started

### dvt init

Initialize the `dvt` configuration directory, creating the required subdirectory structure.

**Usage:**

```
dvt init [--config <dir>]
```

Creates:
- `~/.dvt/prompts/`
- `~/.dvt/plugins/`
- `~/.dvt/shells/`
- `~/.dvt/profiles/`

**Examples:**

```bash
# Initialize with defaults (~/.dvt/)
dvt init

# Initialize at a custom path
dvt init --config ~/my-terminal-config
```

---

### dvt version

Print version information for the `dvt` binary.

**Usage:**

```
dvt version [--short]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--short` | Print only the version number |

**Examples:**

```bash
# Full version info (version, build time, commit)
dvt version

# Just the version number (useful for scripting)
dvt version --short
```

---

### dvt completion

Generate shell autocompletion scripts for `dvt`.

**Usage:**

```
dvt completion <bash|zsh|fish|powershell>
```

**Examples:**

```bash
# Load completions in the current Zsh session
source <(dvt completion zsh)

# Install permanently for Zsh on macOS (Homebrew)
dvt completion zsh > $(brew --prefix)/share/zsh/site-functions/_dvt

# Install permanently for Fish
dvt completion fish > ~/.config/fish/completions/dvt.fish
```

---

## Prompts

Prompts define how your shell prompt looks using Starship or P10k. Import from the built-in library, customize with `apply`, and generate config files with `generate`.

### dvt prompt library get

List available prompts in the built-in library.

**Usage:**

```
dvt prompt library get [--category <cat>] [--tag <tag>] [-o <format>]
```

**Aliases:** `dvt prompt library list` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `table` | Output format: `table`, `yaml`, `json` |
| `-c, --category` | | Filter by category |
| `-t, --tag` | | Filter by tag |

**Examples:**

```bash
# List all library prompts
dvt prompt library get

# Filter by category
dvt prompt library get --category minimal
```

---

### dvt prompt library describe

Show full details of a single prompt from the built-in library.

**Usage:**

```
dvt prompt library describe <name> [-o <format>]
```

**Aliases:** `dvt prompt library show` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `yaml` | Output format: `yaml`, `json` |

**Examples:**

```bash
# Show YAML definition of a library prompt
dvt prompt library describe starship-default

# Show as JSON
dvt prompt library describe starship-minimal -o json
```

---

### dvt prompt library import

Copy prompts from the built-in library to your local store (database + file store).

**Usage:**

```
dvt prompt library import <name>... [--all]
```

**Aliases:** `dvt prompt library install` (deprecated)

**Flags:**

| Flag | Description |
|------|-------------|
| `--all` | Import all prompts from the library |

**Examples:**

```bash
# Import a single prompt
dvt prompt library import starship-default

# Import multiple prompts at once
dvt prompt library import starship-minimal starship-powerline

# Import everything
dvt prompt library import --all
```

---

### dvt prompt library categories

List all prompt categories available in the built-in library.

**Usage:**

```
dvt prompt library categories
```

**Examples:**

```bash
# See all categories and prompt counts
dvt prompt library categories
```

---

### dvt prompt get

Get prompt definitions from the local store. With no arguments, lists all installed prompts.

**Usage:**

```
dvt prompt get [name] [-o <format>]
```

**Aliases:** `dvt prompt list` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `yaml` | Output format: `table`, `yaml`, `json` |

**Examples:**

```bash
# List all installed prompts
dvt prompt get

# Get a specific prompt as YAML
dvt prompt get coolnight

# Get as JSON
dvt prompt get coolnight -o json
```

---

### dvt prompt apply

Apply a prompt definition from a YAML file (kubectl-style). Creates or updates the prompt in the database.

**Usage:**

```
dvt prompt apply -f <file> [-f <file>...]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-f, --filename` | Prompt YAML file(s); use `-` for stdin |

**Examples:**

```bash
# Apply from a file
dvt prompt apply -f my-prompt.yaml

# Apply from stdin
cat my-prompt.yaml | dvt prompt apply -f -
```

---

### dvt prompt delete

Delete a prompt from the database (kubectl-style). Prompts for confirmation unless `--force` is used.

**Usage:**

```
dvt prompt delete <name> [--force]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt |

**Examples:**

```bash
# Delete with confirmation
dvt prompt delete coolnight

# Delete without confirmation
dvt prompt delete coolnight --force
```

---

### dvt prompt generate

Generate the config file for a prompt (e.g., `starship.toml` for Starship prompts). Output goes to stdout.

**Usage:**

```
dvt prompt generate <name>
```

**Examples:**

```bash
# Preview the generated starship.toml
dvt prompt generate coolnight

# Save directly to Starship's config location
dvt prompt generate coolnight > ~/.config/starship.toml
```

---

### dvt prompt set

Set the active prompt and validate its config (kubectl-style). Requires confirmation unless `--force` is used.

**Usage:**

```
dvt prompt set <name> [--force]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt |

**Examples:**

```bash
# Set a prompt as active
dvt prompt set coolnight

# Set without confirmation
dvt prompt set coolnight --force
```

---

### dvt get prompts

List all installed prompts using the kubectl-style top-level `get` verb.

**Usage:**

```
dvt get prompts [--type <type>] [-o <format>]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `table` | Output format: `table`, `yaml`, `json` |
| `--type` | | Filter by prompt type: `starship`, `powerlevel10k` |

**Examples:**

```bash
# List all prompts
dvt get prompts

# Filter to Starship prompts only
dvt get prompts --type starship
```

---

## Plugins

Plugins enhance your shell with features like autosuggestions, syntax highlighting, and fuzzy finding.

### dvt plugin library get

List available plugins in the built-in library.

**Usage:**

```
dvt plugin library get [--category <cat>] [--tag <tag>] [-o <format>]
```

**Aliases:** `dvt plugin library list` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `table` | Output format: `table`, `yaml`, `json` |
| `-c, --category` | | Filter by category |
| `-t, --tag` | | Filter by tag |

**Examples:**

```bash
# List all library plugins
dvt plugin library get

# Filter by category
dvt plugin library get --category completion
```

---

### dvt plugin library describe

Show full details of a single plugin from the built-in library.

**Usage:**

```
dvt plugin library describe <name> [-o <format>]
```

**Aliases:** `dvt plugin library show` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `yaml` | Output format: `yaml`, `json` |

**Examples:**

```bash
# Show plugin details
dvt plugin library describe zsh-autosuggestions

# Show as JSON
dvt plugin library describe fzf -o json
```

---

### dvt plugin library import

Copy plugins from the built-in library to the local store.

**Usage:**

```
dvt plugin library import <name>... [--all]
```

**Aliases:** `dvt plugin library install` (deprecated)

**Flags:**

| Flag | Description |
|------|-------------|
| `--all` | Import all plugins from the library |

**Examples:**

```bash
# Import a single plugin
dvt plugin library import zsh-autosuggestions

# Import multiple plugins
dvt plugin library import zsh-autosuggestions zsh-syntax-highlighting

# Import all library plugins
dvt plugin library import --all
```

---

### dvt plugin library categories

List all plugin categories available in the built-in library.

**Usage:**

```
dvt plugin library categories
```

**Examples:**

```bash
# See all categories and plugin counts
dvt plugin library categories
```

---

### dvt plugin get

Get plugin definitions from the local store. With no arguments, lists all installed plugins.

**Usage:**

```
dvt plugin get [name] [-o <format>]
```

**Aliases:** `dvt plugin list` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `yaml` | Output format: `table`, `yaml`, `json` |

**Examples:**

```bash
# List all installed plugins
dvt plugin get

# Get a specific plugin
dvt plugin get zsh-autosuggestions
```

---

### dvt plugin generate

Generate the `.zshrc` plugin section for installed plugins. Output goes to stdout.

**Usage:**

```
dvt plugin generate [name...] [-m <manager>]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-m, --manager` | `manual` | Plugin manager format: `zinit`, `oh-my-zsh`, `antigen`, `sheldon`, `manual` |

**Examples:**

```bash
# Generate for all installed plugins
dvt plugin generate

# Generate for a specific plugin using zinit format
dvt plugin generate zsh-autosuggestions --manager zinit

# Append plugin config to .zshrc
dvt plugin generate >> ~/.zshrc
```

---

## Packages

Packages group related terminal configuration into reusable bundles with inheritance. A `developer` package might extend `core` and add development-specific plugins and prompts.

### dvt package get

Show package details or list all packages when no name is given.

**Usage:**

```
dvt package get [name] [-o <format>] [-w] [--category <cat>]
```

**Aliases:** `dvt package list` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `table` | Output format: `table`, `yaml`, `json` |
| `-w, --wide` | | Show extended columns (plugins, prompts, profiles, extends) |
| `-c, --category` | | Filter by category |
| `--library` | | Show only library packages |
| `--user` | | Show only user packages |

**Examples:**

```bash
# List all packages
dvt package get

# Show extended info
dvt package get -w

# Show a specific package (YAML output with resolved inheritance)
dvt package get developer
```

---

### dvt package library get

List available packages in the built-in library.

**Usage:**

```
dvt package library get [--category <cat>] [-o <format>] [-w]
```

**Aliases:** `dvt package library list` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `table` | Output format: `table`, `yaml`, `json` |
| `-w, --wide` | | Show extended columns |
| `-c, --category` | | Filter by category |

**Examples:**

```bash
# List all library packages
dvt package library get

# Filter by category
dvt package library get --category development
```

---

### dvt package library describe

Show full details of a package from the built-in library, with resolved inheritance.

**Usage:**

```
dvt package library describe <name> [-o <format>]
```

**Aliases:** `dvt package library show` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `yaml` | Output format: `yaml`, `json` |

**Examples:**

```bash
# Show resolved package definition
dvt package library describe developer

# Show as JSON
dvt package library describe core -o json
```

---

### dvt package library import

Import a package and all its components (resolved through inheritance) into the database.

**Usage:**

```
dvt package library import <name> [--dry-run]
```

**Aliases:** `dvt package library install` (deprecated)

**Flags:**

| Flag | Description |
|------|-------------|
| `--dry-run` | Show what would be installed without installing |

**Examples:**

```bash
# Preview what the developer package would install
dvt package library import developer --dry-run

# Install the core package
dvt package library import core

# Install developer (also installs core via inheritance)
dvt package library import developer
```

---

## Shell

Shell configs define aliases, environment variables, and functions — the non-prompt parts of your shell setup.

### dvt shell apply

Apply a shell configuration from a YAML file. Creates or updates the config in the local store.

**Usage:**

```
dvt shell apply -f <file> [-f <file>...]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-f, --filename` | Shell config YAML file(s); use `-` for stdin |

**Examples:**

```bash
# Apply from a file
dvt shell apply -f my-shell.yaml

# Apply from stdin
cat my-shell.yaml | dvt shell apply -f -
```

---

### dvt shell get

Get shell configurations from the local store. With no arguments, lists all installed configs.

**Usage:**

```
dvt shell get [name] [-o <format>]
```

**Aliases:** `dvt shell list` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `yaml` | Output format: `table`, `yaml`, `json` |

**Examples:**

```bash
# List all installed shell configs
dvt shell get

# Get a specific shell config
dvt shell get my-shell
```

---

### dvt shell generate

Generate the shell configuration section (aliases, env vars, functions) for a shell config. Output goes to stdout.

**Usage:**

```
dvt shell generate <name>
```

**Examples:**

```bash
# Preview generated shell config
dvt shell generate my-shell

# Append to .zshrc
dvt shell generate my-shell >> ~/.zshrc
```

---

### dvt shell generate-workspace

Generate a workspace-scoped shell config file that composites host config, workspace env vars, installed plugins, shell configurations, and prompt initialization. Supports Bash, Zsh, and Fish.

**Usage:**

```
dvt shell generate-workspace [flags]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | stdout | Write to file instead of stdout |
| `-w, --workspace` | | Workspace name (for header comment) |
| `-s, --shell` | auto | Target shell: `bash`, `zsh`, `fish` (default: auto-detect from `$SHELL`) |
| `--no-host-config` | | Skip including the host shell config |
| `--host-config` | | Path to host shell config file |
| `--plugin-dir` | | Plugin installation directory |
| `--env` | | Environment variables as `KEY=VALUE` (repeatable) |
| `--prompt-init` | | Prompt init command (e.g., `eval "$(starship init zsh)"`) |

**Examples:**

```bash
# Generate workspace shell config to stdout
dvt shell generate-workspace

# Write to a file for use in a workspace
dvt shell generate-workspace --output .dvm/.zshrc.workspace

# Workspace-specific env vars and custom shell
dvt shell generate-workspace --shell zsh --env APP_ENV=dev --env DEBUG=1
```

---

## Profiles

Profiles are the recommended way to manage your terminal configuration — they combine a prompt, plugins, and shell settings into a single named unit.

### dvt profile preset get

List available profile presets.

**Usage:**

```
dvt profile preset get
```

**Aliases:** `dvt profile preset list` (deprecated)

Available presets:

| Preset | Description |
|--------|-------------|
| `default` | Balanced setup with Starship, autosuggestions, and syntax-highlighting |
| `minimal` | Lightweight setup with just Starship and basic plugins |
| `power-user` | Full-featured setup with all plugins and nerd font support |

**Examples:**

```bash
# See all presets
dvt profile preset get
```

---

### dvt profile preset import

Import a profile preset and all its dependencies into the local store.

**Usage:**

```
dvt profile preset import <name>
```

**Aliases:** `dvt profile preset install` (deprecated)

**Examples:**

```bash
# Import the default preset
dvt profile preset import default

# Import the power-user preset
dvt profile preset import power-user
```

---

### dvt profile apply

Apply a profile definition from a YAML file. Creates or updates the profile in the local store.

**Usage:**

```
dvt profile apply -f <file> [-f <file>...]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-f, --filename` | Profile YAML file(s); use `-` for stdin |

**Examples:**

```bash
# Apply a profile from file
dvt profile apply -f my-profile.yaml

# Apply from stdin
cat my-profile.yaml | dvt profile apply -f -
```

---

### dvt profile get

Get profile definitions from the local store. With no arguments, lists all installed profiles.

**Usage:**

```
dvt profile get [name] [-o <format>]
```

**Aliases:** `dvt profile list` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `yaml` | Output format: `table`, `yaml`, `json` |

**Examples:**

```bash
# List all installed profiles
dvt profile get

# Get a specific profile
dvt profile get default
```

---

### dvt profile generate

Generate all configuration files for a profile — `starship.toml`, plugin loading code, and shell config. Output goes to stdout by default.

**Usage:**

```
dvt profile generate <name> [--output-dir <dir>] [--dry-run]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--output-dir` | Directory to write config files (default: stdout) |
| `--dry-run` | Show what would be generated without writing files |

**Examples:**

```bash
# Preview all generated config to stdout
dvt profile generate default

# Write config files to ~/.config/
dvt profile generate default --output-dir ~/.config/

# Preview what would be written
dvt profile generate default --output-dir ~/.config/ --dry-run
```

---

### dvt profile use

Set the active profile by name.

**Usage:**

```
dvt profile use <name>
```

**Examples:**

```bash
# Set default as the active profile
dvt profile use default

# Switch to minimal profile
dvt profile use minimal
```

---

## Emulators

Manage configurations for terminal emulators like WezTerm, Alacritty, Kitty, and iTerm2.

### dvt emulator get

Show emulator configuration details or list all emulators when no name is given.

**Usage:**

```
dvt emulator get [name] [--type <type>] [--category <cat>] [-o <format>]
```

**Aliases:** `dvt emulator list` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `table` | Output format: `table`, `yaml`, `json` |
| `--type` | | Filter by emulator type: `wezterm`, `alacritty`, `kitty`, `iterm2` |
| `--category` | | Filter by category |

**Examples:**

```bash
# List all emulators
dvt emulator get

# Filter by type
dvt emulator get --type wezterm

# Show a specific emulator
dvt emulator get wezterm-default
```

---

### dvt emulator enable

Enable a terminal emulator configuration by setting `enabled=true`.

**Usage:**

```
dvt emulator enable <name>
```

**Examples:**

```bash
# Enable an emulator config
dvt emulator enable wezterm-default
```

---

### dvt emulator disable

Disable a terminal emulator configuration by setting `enabled=false`.

**Usage:**

```
dvt emulator disable <name>
```

**Examples:**

```bash
# Disable an emulator config
dvt emulator disable old-config
```

---

### dvt emulator apply

Apply a terminal emulator configuration from a YAML file.

**Usage:**

```
dvt emulator apply -f <file> [--dry-run]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-f, --filename` | YAML file to apply (required); use `-` for stdin |
| `--dry-run` | Show what would be applied without applying |

**Examples:**

```bash
# Apply from a file
dvt emulator apply -f my-emulator.yaml

# Preview what would be applied
dvt emulator apply -f my-emulator.yaml --dry-run
```

---

### dvt emulator library get

List available emulators in the built-in library.

**Usage:**

```
dvt emulator library get [--type <type>] [--category <cat>] [-o <format>]
```

**Aliases:** `dvt emulator library list` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `table` | Output format: `table`, `yaml`, `json` |
| `--type` | | Filter by emulator type |
| `--category` | | Filter by category |

**Examples:**

```bash
# List all library emulators
dvt emulator library get

# Filter to WezTerm configs only
dvt emulator library get --type wezterm
```

---

### dvt emulator library describe

Show details of a specific emulator from the built-in library.

**Usage:**

```
dvt emulator library describe <name> [-o <format>]
```

**Aliases:** `dvt emulator library show` (deprecated)

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `-o, --output` | `yaml` | Output format: `yaml`, `json` |

**Examples:**

```bash
# Show emulator details
dvt emulator library describe maestro

# Show as YAML
dvt emulator library describe minimal -o yaml
```

---

### dvt emulator library import

Import a terminal emulator configuration from the built-in library into the database.

**Usage:**

```
dvt emulator library import <name> [--force] [--dry-run]
```

**Aliases:** `dvt emulator install` (deprecated)

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Overwrite if the emulator already exists |
| `--dry-run` | Show what would be installed without installing |

**Examples:**

```bash
# Preview what would be imported
dvt emulator library import maestro --dry-run

# Import a library emulator
dvt emulator library import maestro

# Overwrite an existing config
dvt emulator library import minimal --force
```

---

## WezTerm

Manage WezTerm terminal configurations using YAML presets. Presets can reference themes for automatic color resolution.

### dvt wezterm get

List available WezTerm presets from the built-in library.

**Usage:**

```
dvt wezterm get
```

**Aliases:** `dvt wezterm list` (deprecated)

**Examples:**

```bash
# List all WezTerm presets
dvt wezterm get
```

---

### dvt wezterm describe

Show details of a specific WezTerm preset — font, opacity, theme reference, and key bindings.

**Usage:**

```
dvt wezterm describe <name>
```

**Aliases:** `dvt wezterm show` (deprecated)

**Examples:**

```bash
# Show preset details
dvt wezterm describe default

# Show details of the minimal preset
dvt wezterm describe minimal
```

---

### dvt wezterm generate

Generate WezTerm Lua configuration from a named preset and print to stdout. Resolves theme colors if a `theme_ref` is set on the preset.

**Usage:**

```
dvt wezterm generate <name>
```

**Examples:**

```bash
# Preview generated Lua config
dvt wezterm generate default

# Save to a custom file
dvt wezterm generate default > ~/custom.wezterm.lua
```

---

### dvt wezterm use

Apply a WezTerm preset by writing the generated Lua config to `~/.wezterm.lua` (or a custom path). Resolves theme colors automatically if a `theme_ref` is configured.

**Usage:**

```
dvt wezterm use <name> [--output-file <path>]
```

**Aliases:** `dvt wezterm apply`

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--output-file` | `~/.wezterm.lua` | Output file path |

**Examples:**

```bash
# Apply the default preset to ~/.wezterm.lua
dvt wezterm use default

# Write to a custom location
dvt wezterm use default --output-file ~/dotfiles/.wezterm.lua
```

---

## Tool Config

Generate theme-aware configuration snippets for terminal tools using the active color palette. Ensures visual consistency across your entire development environment.

Supported tools: `bat`, `delta`, `fzf`, `dircolors`

### dvt tool-config list

List all supported tool config generators with their descriptions.

**Usage:**

```
dvt tool-config list
```

**Examples:**

```bash
# See all supported tools
dvt tool-config list
```

---

### dvt tool-config generate

Generate a theme-aware config snippet for a terminal tool. By default, output is formatted as shell `export` statements.

**Usage:**

```
dvt tool-config generate [tool-name] [--all] [--format <format>]
```

**Flags:**

| Flag | Default | Description |
|------|---------|-------------|
| `--all` | | Generate config for all supported tools |
| `--format` | `env` | Output format: `env` (shell exports) or `raw` (config value only) |

**Examples:**

```bash
# Generate fzf color config as shell export
dvt tool-config generate fzf

# Generate raw config value (no export wrapping)
dvt tool-config generate fzf --format raw

# Generate for all tools at once
dvt tool-config generate --all

# Generate delta config (outputs gitconfig snippet)
dvt tool-config generate delta
```
