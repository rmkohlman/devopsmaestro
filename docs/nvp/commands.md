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

### `nvp list`

List installed plugins.

```bash
nvp list
```

### `nvp get`

Get details of an installed plugin.

```bash
nvp get <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `table` |

**Examples:**

```bash
nvp get telescope
nvp get telescope -o yaml
```

### `nvp enable`

Enable an installed plugin.

```bash
nvp enable <name>
```

**Example:**

```bash
nvp enable telescope
```

### `nvp disable`

Disable an installed plugin (without deleting it).

```bash
nvp disable <name>
```

**Example:**

```bash
nvp disable copilot
```

### `nvp delete`

Delete an installed plugin.

```bash
nvp delete <name>
```

**Example:**

```bash
nvp delete telescope
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

### `nvp library list`

List plugins available in the remote library.

```bash
nvp library list [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--category <cat>` | Filter by category |
| `--tags <tags>` | Filter by tags |

### `nvp library show`

Show details of a library plugin.

```bash
nvp library show <name>
```

### `nvp library install`

Install a plugin from the library.

```bash
nvp library install <name>
```

**Example:**

```bash
nvp library install telescope
nvp library install treesitter
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

### `nvp source list`

List configured plugin sources.

```bash
nvp source list
```

### `nvp source show`

Show details of a source.

```bash
nvp source show <name>
```

### `nvp source sync`

Sync plugins from a source.

```bash
nvp source sync <name>
```

---

## Theme Management

### `nvp theme list`

List all available themes (library + user).

```bash
nvp theme list
```

### `nvp theme get`

Get theme details.

```bash
nvp theme get [name] [flags]
```

If no name provided, shows the active theme.

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `table` |

**Examples:**

```bash
nvp theme get
nvp theme get coolnight-ocean
nvp theme get coolnight-ocean -o yaml
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
| `--from <value>` | Base color: hue (0-360), hex (#rrggbb), or preset name |
| `--name <name>` | Name for the new theme |
| `--use` | Set new theme as active after creation |

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
nvp theme delete <name>
```

### `nvp theme generate`

Generate Lua files for the active theme only.

```bash
nvp theme generate
```

---

## Theme Library

### `nvp theme library list`

List themes available in the remote library.

```bash
nvp theme library list [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--category <cat>` | Filter by category |

### `nvp theme library show`

Show details of a library theme.

```bash
nvp theme library show <name>
```

### `nvp theme library install`

Install a theme from the library.

```bash
nvp theme library install <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--use` | Set theme as active after installing |

**Example:**

```bash
nvp theme library install catppuccin-mocha --use
```

### `nvp theme library categories`

List available theme categories in the library.

```bash
nvp theme library categories
```

### `nvp theme library tags`

List available theme tags in the library.

```bash
nvp theme library tags
```

---

## Configuration

### `nvp config init`

Initialize nvp configuration file.

```bash
nvp config init
```

### `nvp config show`

Show current configuration.

```bash
nvp config show
```

### `nvp config generate`

Generate a default configuration file.

```bash
nvp config generate
```

### `nvp config edit`

Open configuration file in editor.

```bash
nvp config edit
```

---

## Generation

### `nvp generate`

Generate Lua files from installed plugins and active theme.

```bash
nvp generate [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <dir>` | Output directory (default: `~/.config/nvim/lua/plugins/nvp`) |

**Examples:**

```bash
nvp generate
nvp generate --output ~/my-nvim-config/lua/plugins
```

### `nvp generate-lua`

Generate Lua configuration files (alias for `nvp generate`).

```bash
nvp generate-lua [flags]
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
nvp version
```

---

## Common Workflows

### Install Plugins from Library

```bash
# Browse the library
nvp library list
nvp library list --category lsp

# Install individual plugins
nvp library install telescope
nvp library install treesitter

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
nvp theme library list
nvp theme library install catppuccin-mocha --use
nvp generate
```
