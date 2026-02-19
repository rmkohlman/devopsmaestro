# nvp Commands Reference

Complete reference for all `nvp` commands in v0.12.0.

---

## Global Flags

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Enable debug logging |
| `--log-file <path>` | Write logs to file (JSON format) |
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

## Plugin Library

### `nvp plugin list`

List available plugins and packages.

```bash
nvp plugin list
```

### `nvp apply`

Apply plugin or package from various sources.

```bash
nvp apply -f <source>
```

**Source types:**

| Type | Example |
|------|---------|
| Package | `package:rkohlman-full` |
| Plugin | `plugin:telescope` |
| File | `plugin.yaml` |
| URL | `https://example.com/plugin.yaml` |
| GitHub | `github:user/repo/plugin.yaml` |
| Stdin | `-` |

**Examples:**

```bash
# Install complete plugin package
nvp apply -f package:rkohlman-full

# Install individual plugin
nvp apply -f plugin:telescope

# Install from file/URL
nvp apply -f my-plugin.yaml
nvp apply -f https://example.com/plugin.yaml
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml
cat plugin.yaml | nvp apply -f -
```

## Plugin Management

### `nvp plugin list`

List installed plugins.

```bash
nvp plugin list
```

### `nvp plugin get`

Get details of an installed plugin.

```bash
nvp plugin get <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `table` |

**Examples:**

```bash
nvp plugin get telescope
nvp plugin get telescope -o yaml
```

### `nvp plugin delete`

Delete an installed plugin.

```bash
nvp plugin delete <name>
```

**Example:**

```bash
nvp plugin delete telescope
```

---

## Theme Library

### `nvp theme list`

List all available themes (library + user).

```bash
nvp theme list
```

### `nvp theme create`

Create a custom CoolNight theme variant using the parametric generator.

```bash
nvp theme create --hue <degrees> --name <name>
```

**Examples:**

```bash
nvp theme create --hue 210 --name my-blue-theme
nvp theme create --hue 280 --name synthwave-purple
nvp theme create --hue 120 --name matrix-green
```

### `nvp theme use`

Set active theme.

```bash
nvp theme use <name>
```

**Example:**

```bash
nvp theme use coolnight-ocean
```

---

## Theme Management

### `nvp apply`

Apply theme from file.

```bash
nvp apply -f <file>
```

**Example:**

```bash
nvp apply -f my-theme.yaml
```

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

If no name provided, shows active theme.

**Examples:**

```bash
nvp theme get
nvp theme get coolnight-ocean
nvp theme get -o yaml
```

### `nvp theme use`

Set active theme.

```bash
nvp theme use <name>
```

**Example:**

```bash
nvp theme use coolnight-ocean
```

### `nvp theme delete`

Delete a theme.

```bash
nvp theme delete <name>
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
| `--output <dir>` | Output directory (default: `~/.config/nvim/lua/plugins/nvp`) |

**Examples:**

```bash
nvp generate
nvp generate --output ~/my-nvim-config/lua/plugins
```

### `nvp theme generate`

Generate Lua files for active theme only.

```bash
nvp theme generate
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

### Install Complete Setup

```bash
# Install complete plugin package
nvp apply -f package:rkohlman-full

# Set theme
nvp theme use coolnight-ocean

# Generate Lua files
nvp generate
```

### Install Individual Plugins

```bash
nvp apply -f plugin:telescope
nvp apply -f plugin:treesitter
nvp generate
```

### Create Custom Theme

```bash
nvp theme create --hue 280 --name my-synthwave
nvp theme use my-synthwave
nvp generate
```

### Switch Themes

```bash
nvp theme use catppuccin-mocha
nvp generate
```
