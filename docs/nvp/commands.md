# nvp Commands Reference

Complete reference for all `nvp` commands.

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

### `nvp library list`

List available plugins in the library.

```bash
nvp library list
```

### `nvp library get`

Show details of a library plugin.

```bash
nvp library get <name>
```

**Example:**

```bash
nvp library get telescope
```

### `nvp library install`

Install plugin(s) from library.

```bash
nvp library install <name> [name...]
```

**Examples:**

```bash
nvp library install telescope
nvp library install telescope treesitter lspconfig
```

---

## Plugin Management

### `nvp apply`

Apply plugin from file, URL, or stdin.

```bash
nvp apply -f <source>
```

**Source types:**

| Type | Example |
|------|---------|
| File | `-f plugin.yaml` |
| URL | `-f https://example.com/plugin.yaml` |
| GitHub | `-f github:user/repo/plugin.yaml` |
| Stdin | `-f -` |

**Examples:**

```bash
nvp apply -f my-plugin.yaml
nvp apply -f https://example.com/plugin.yaml
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml
cat plugin.yaml | nvp apply -f -
```

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

## Theme Library

### `nvp theme library list`

List available themes.

```bash
nvp theme library list
```

### `nvp theme library show`

Show theme details.

```bash
nvp theme library show <name>
```

### `nvp theme library install`

Install theme from library.

```bash
nvp theme library install <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--use` | Set as active theme after install |

**Examples:**

```bash
nvp theme library install tokyonight-custom
nvp theme library install catppuccin-mocha --use
```

---

## Theme Management

### `nvp theme apply`

Apply theme from file.

```bash
nvp theme apply -f <file>
```

**Example:**

```bash
nvp theme apply -f my-theme.yaml
```

### `nvp theme list`

List installed themes.

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
nvp theme get tokyonight-custom
nvp theme get -o yaml
```

### `nvp theme use`

Set active theme.

```bash
nvp theme use <name>
```

**Example:**

```bash
nvp theme use catppuccin-mocha
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

### Install and Generate

```bash
nvp library install telescope treesitter lspconfig
nvp theme library install tokyonight-custom --use
nvp generate
```

### Update a Plugin

```bash
nvp apply -f github:rmkohlman/nvim-yaml-plugins/plugins/telescope.yaml
nvp generate
```

### Switch Themes

```bash
nvp theme use catppuccin-mocha
nvp generate
```
