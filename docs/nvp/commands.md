# nvp Commands Reference

Complete reference for all `nvp` commands.

---

## Global Flags

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Enable debug logging |
| `--log-file <path>` | Write logs to file (JSON format) |
| `--config <path>` | Path to config file |
| `--no-color` | Disable color output |
| `-h, --help` | Show help for command |

---

## Initialization

### `nvp init`

Initialize nvp store.

```bash
nvp init
```

Creates `~/.nvp/` directory structure.

---

## Plugin Management

### `nvp get`

List all plugins in the local store, or get a specific plugin definition.

```bash
nvp get [name] [flags]
```

With no arguments, lists all plugins. With a name, shows that plugin's full definition.

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `yaml`, `json`, `table` (default: `yaml`) |
| `-c, --category <cat>` | Filter by category (list mode only) |
| `--enabled` | Show only enabled plugins |
| `--disabled` | Show only disabled plugins |
| `--show-deps` | Show dependency tree for a plugin (single-get mode) |

**Examples:**

```bash
nvp get                      # List all plugins
nvp get -c lsp               # List plugins filtered by category
nvp get telescope            # Get specific plugin as YAML
nvp get telescope -o json    # Get specific plugin as JSON
nvp get telescope --show-deps
```

### `nvp enable`

Enable one or more installed plugins.

```bash
nvp enable <name>...
```

**Example:**

```bash
nvp enable telescope
nvp enable telescope treesitter lspconfig
```

### `nvp disable`

Disable one or more installed plugins (without deleting them).

```bash
nvp disable <name>...
```

**Example:**

```bash
nvp disable copilot
nvp disable copilot codeium
```

### `nvp delete`

Delete a plugin from the local store.

```bash
nvp delete <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt |

**Example:**

```bash
nvp delete telescope
nvp delete telescope --force
```

---

## Applying Resources

### `nvp apply`

Apply a plugin or theme from a source.

```bash
nvp apply -f <source>
```

**Source types:**

| Type | Example |
|------|---------|
| File | `plugin.yaml` |
| URL | `https://example.com/plugin.yaml` |
| GitHub shorthand | `github:user/repo/plugin.yaml` |
| Stdin | `-` |

**Examples:**

```bash
# Install from local file
nvp apply -f my-plugin.yaml

# Install from URL
nvp apply -f https://example.com/plugin.yaml

# Install from GitHub
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml

# Install from stdin
cat plugin.yaml | nvp apply -f -
```

---

## Plugin Library

### `nvp library get`

List all plugins available in the remote library.

```bash
nvp library get [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `table`, `yaml`, `json` (default: `table`) |
| `-c, --category <cat>` | Filter by category |
| `-t, --tag <tag>` | Filter by tag |

### `nvp library describe`

Show details of a library plugin.

```bash
nvp library describe <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `yaml`, `json` (default: `yaml`) |

### `nvp library import`

Import plugins from the library to your local store.

```bash
nvp library import <name>... [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--all` | Import all plugins from the library |

**Example:**

```bash
nvp library import telescope
nvp library import telescope treesitter lspconfig
nvp library import --all
```

### `nvp library categories`

List available plugin categories in the library.

```bash
nvp library categories
```

### `nvp library tags`

List available plugin tags in the library.

```bash
nvp library tags
```

---

## Source Management

### `nvp source get`

List available external plugin sources.

```bash
nvp source get [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `table`, `yaml`, `json` (default: `table`) |

### `nvp source describe`

Show details of a source.

```bash
nvp source describe <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `yaml`, `json` (default: `yaml`) |

### `nvp source sync`

Sync plugins from an external source.

```bash
nvp source sync <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview changes without applying |
| `-l, --selector <key=value>` | Label selector to filter plugins (repeatable) |
| `--tag <version>` | Specific version/tag to sync from |
| `--force` | Overwrite existing plugins |
| `-o, --output <format>` | Output format: `table`, `yaml`, `json` (default: `table`) |

**Examples:**

```bash
nvp source sync lazyvim
nvp source sync lazyvim --dry-run
nvp source sync lazyvim -l category=lsp
nvp source sync lazyvim --tag v15.0.0
nvp source sync lazyvim --force
```

---

## Theme Management

### `nvp theme get`

List all installed themes, or get a specific theme definition.

```bash
nvp theme get [name] [flags]
```

With no arguments, lists all installed themes (active theme marked with `*`).
With a name, shows that theme's full definition.

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `yaml`, `json`, `table` (default: `yaml`) |

**Examples:**

```bash
nvp theme get
nvp theme get coolnight-ocean
nvp theme get coolnight-ocean -o json
```

### `nvp theme apply`

Apply a theme definition from a YAML file or URL.

```bash
nvp theme apply -f <source> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-f, --filename <source>` | Theme YAML file(s) or URL(s) to apply (repeatable) |

**Source types:**

| Type | Example |
|------|---------|
| File | `my-theme.yaml` |
| URL | `https://example.com/theme.yaml` |
| GitHub shorthand | `github:user/repo/themes/custom.yaml` |
| Stdin | `-` |

**Examples:**

```bash
nvp theme apply -f my-theme.yaml
nvp theme apply -f github:user/repo/themes/custom.yaml
```

### `nvp theme use`

Set the active theme.

```bash
nvp theme use <name>
```

**Example:**

```bash
nvp theme use coolnight-ocean
```

### `nvp theme create`

Create a custom CoolNight theme variant using the parametric generator.

```bash
nvp theme create --from <value> --name <name> [flags]
```

The `--from` value can be:
- A hue angle in degrees (`"210"`)
- A hex color (`"#8B00FF"`)
- A named preset (`"synthwave"`, `"ocean"`, `"forest"`)

**Flags:**

| Flag | Description |
|------|-------------|
| `--from <value>` | Base color: hue (0-360), hex (#rrggbb), or preset name **(required)** |
| `--name <name>` | Name for the new theme (required unless `--dry-run`) |
| `--use` | Set new theme as active after creation |
| `--dry-run` | Preview the generated theme without saving |
| `-o, --output <format>` | Output format for `--dry-run`: `yaml`, `json`, `table` (default: `yaml`) |

**Examples:**

```bash
nvp theme create --from "210" --name my-blue-theme
nvp theme create --from "#8B00FF" --name my-violet-theme
nvp theme create --from "synthwave" --name my-synth --use
```

### `nvp theme preview`

Preview a theme in the terminal.

```bash
nvp theme preview <name>
```

**Example:**

```bash
nvp theme preview coolnight-ocean
```

### `nvp theme delete`

Delete a theme.

```bash
nvp theme delete <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Skip confirmation prompt |

### `nvp theme generate`

Generate Lua files for the active theme only.

```bash
nvp theme generate [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--output-dir <dir>` | Output directory (default: `~/.config/nvim/lua`) |
| `--dry-run` | Show what would be generated without writing files |

**Generated files:**

- `lua/theme/palette.lua` — color palette module
- `lua/theme/init.lua` — theme setup and helpers
- `lua/plugins/nvp/colorscheme.lua` — lazy.nvim plugin spec

---

## Theme Library

### `nvp theme library get`

List themes available in the remote library.

```bash
nvp theme library get [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `table`, `yaml`, `json` (default: `table`) |
| `-c, --category <cat>` | Filter by category (e.g. `dark`, `light`) |

### `nvp theme library describe`

Show details of a library theme.

```bash
nvp theme library describe <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `yaml`, `json` (default: `yaml`) |

### `nvp theme library import`

Import a theme from the library to your local store.

```bash
nvp theme library import <name>... [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--use` | Set theme as active after importing |

**Example:**

```bash
nvp theme library import catppuccin-mocha --use
```

### `nvp theme library categories`

List available theme categories in the library.

```bash
nvp theme library categories
```

---

## Configuration

### `nvp config init`

Initialize nvp configuration file (`core.yaml`).

```bash
nvp config init [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Overwrite existing `core.yaml` |

### `nvp config describe`

Show current core configuration.

```bash
nvp config describe [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `yaml`, `json` (default: `yaml`) |

### `nvp config generate`

Generate a complete Neovim configuration from `core.yaml` and installed plugins.

```bash
nvp config generate [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--output-dir <dir>` | Output directory (default: `~/.config/nvim`) |
| `--dry-run` | Show what would be generated without writing files |

**Examples:**

```bash
nvp config generate
nvp config generate --output-dir /path/to/nvim/config
nvp config generate --dry-run
```

### `nvp config edit`

Open configuration file in editor.

```bash
nvp config edit
```

---

## Generation

### `nvp generate`

Generate Lua files from all enabled plugins (and active theme if set).

```bash
nvp generate [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--output-dir <dir>` | Output directory (default: `~/.config/nvim/lua/plugins/nvp`) |
| `--dry-run` | Show what would be generated without writing files |

**Examples:**

```bash
nvp generate
nvp generate --output-dir ~/my-nvim-config/lua/plugins
nvp generate --dry-run
```

### `nvp generate-lua`

Generate Lua configuration for a single plugin and print to stdout.

```bash
nvp generate-lua <name>
```

---

## Shell Completion

### `nvp completion`

Generate shell completion scripts.

```bash
nvp completion <shell>
```

**Supported shells:** `bash`, `zsh`, `fish`, `powershell`

**Examples:**

```bash
# Bash
nvp completion bash > /etc/bash_completion.d/nvp

# Zsh
nvp completion zsh > "${fpath[1]}/_nvp"

# Fish
nvp completion fish > ~/.config/fish/completions/nvp.fish
```

---

## Version

### `nvp version`

Show version information.

```bash
nvp version [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--short` | Print only the version number |

---

## Health Checks

### `nvp health`

Run health checks for installed Neovim plugins.

```bash
nvp health [plugin-name] [flags]
```

Without arguments, checks all enabled plugins. With a name, checks only that plugin.

Health checks verify:
- Lua modules are loadable (`require()` succeeds)
- Neovim commands exist
- Treesitter parsers are installed
- LSP servers are configured

By default, runs static checks only (no Neovim required). Use `--live` to run checks inside Neovim.

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `table`, `json` (default: `table`) |
| `--live` | Run checks inside Neovim (requires `nvim` on PATH) |
| `--generate-script` | Output the Lua health check script (for debugging) |

**Examples:**

```bash
nvp health                    # Static check all plugins
nvp health telescope          # Check specific plugin
nvp health --live             # Run checks inside Neovim
nvp health -o json            # Output as JSON
```

---

## Lock File

### `nvp lock`

Generate or verify a `lazy-lock.json` lock file for reproducible plugin versions.

```bash
nvp lock [flags]
```

Without flags, generates or updates the lock file from current plugin configuration.
With `--verify`, checks if the current config matches the existing lock file.

The lock file pins specific git commits for each plugin, ensuring reproducible builds across environments.

**Flags:**

| Flag | Description |
|------|-------------|
| `--verify` | Check if current config matches the existing lock file |
| `--output <path>` | Lock file path (default: `~/.nvp/lazy-lock.json`) |

**Examples:**

```bash
nvp lock                      # Generate/update lazy-lock.json
nvp lock --verify             # Check config matches lock file
nvp lock --output /path/to/lazy-lock.json
```

---

## Package Management

### `nvp package get`

List all packages or get details of a specific package.

```bash
nvp package get [name] [flags]
```

Alias: `nvp pkg get`

With no arguments, lists all available packages. With a name, shows the full package definition with all resolved plugins (including inherited ones).

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `yaml`, `json`, `table` (default: `yaml`) |
| `-c, --category <cat>` | Filter by category (list mode only) |
| `--library` | Show only library packages |
| `--user` | Show only user packages |
| `-w, --wide` | Show extended output (includes `EXTENDS` column) |

**Examples:**

```bash
nvp package get                      # List all packages
nvp package get --category language  # Filter by category
nvp package get core                 # Show package details
nvp package get go-dev -o yaml       # Show details as YAML
nvp package get go-dev -w            # Show with extended columns
```

### `nvp package install`

Install a package by adding all its plugins to your local store.

```bash
nvp package install <name> [flags]
```

Alias: `nvp pkg install`

Resolves package inheritance — installing `go-dev` will also install all plugins from its parent package (e.g. `core`).

**Flags:**

| Flag | Description |
|------|-------------|
| `--dry-run` | Show what would be installed without installing |

**Examples:**

```bash
nvp package install core
nvp package install go-dev
nvp package install go-dev --dry-run
```

---

## Common Workflows

### Install Plugins from Library

```bash
# Browse the library
nvp library get
nvp library get --category lsp

# Import individual plugins
nvp library import telescope
nvp library import treesitter

# Generate Lua files
nvp generate
```

### Install from a YAML File

```bash
# Apply a plugin definition from a file or URL
nvp apply -f my-plugin.yaml
nvp apply -f https://example.com/my-plugin.yaml
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml

# Generate Lua files
nvp generate
```

### Create a Custom Theme

```bash
# Create by hue
nvp theme create --from "280" --name my-synthwave --use

# Create from a preset
nvp theme create --from "synthwave" --name my-synth

# Generate Lua files
nvp generate
```

### Switch Themes

```bash
nvp theme use catppuccin-mocha
nvp generate
```

### Install Theme from Library

```bash
nvp theme library get
nvp theme library import catppuccin-mocha --use
nvp generate
```
