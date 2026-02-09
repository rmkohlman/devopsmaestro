# dvm Commands Reference

Complete reference for all `dvm` commands.

---

## Global Flags

These flags work with any command:

| Flag | Description |
|------|-------------|
| `-v, --verbose` | Enable debug logging |
| `--log-file <path>` | Write logs to file (JSON format) |
| `-h, --help` | Show help for command |

---

## Initialization

### `dvm init`

Initialize DevOpsMaestro (creates database).

```bash
dvm init
```

Creates `~/.devopsmaestro/devopsmaestro.db`.

---

## Apps

### `dvm create app`

Create a new app.

```bash
dvm create app <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--from-cwd` | Use current working directory as app path |
| `--path <path>` | Specific path for the app |
| `--description <text>` | App description |

**Examples:**

```bash
# Create from current directory
dvm create app my-api --from-cwd

# Create with explicit path
dvm create app my-api --path ~/Developer/my-api

# Create with description
dvm create app my-api --from-cwd --description "REST API service"
```

### `dvm get apps`

List all apps.

```bash
dvm get apps [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get apps
dvm get app
dvm get app -o yaml
dvm get app -o json
```

### `dvm get app`

Get a specific app.

```bash
dvm get app <name> [flags]
```

**Examples:**

```bash
dvm get app my-api
dvm get app my-api -o yaml
```

### `dvm delete app`

Delete an app.

```bash
dvm delete app <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-f, --force` | Skip confirmation prompt |

**Examples:**

```bash
dvm delete app my-api
dvm delete app my-api --force
```

### `dvm use app`

Set the active app.

```bash
dvm use app <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--clear` | Clear active app |

**Examples:**

```bash
dvm use app my-api
dvm use app --clear
```

---

## Workspaces

### `dvm create workspace`

Create a new workspace.

```bash
dvm create workspace <name> [flags]
```

**Aliases:** `ws`

**Flags:**

| Flag | Description |
|------|-------------|
| `-a, --app <name>` | App name (defaults to active app) |
| `--description <text>` | Workspace description |
| `--image <name>` | Custom image name |

**Examples:**

```bash
dvm create workspace dev
dvm create ws dev
dvm create ws dev -a my-api
dvm create ws staging --description "Staging environment"
```

### `dvm get workspaces`

List workspaces in an app.

```bash
dvm get workspaces [flags]
```

**Aliases:** `ws`

**Flags:**

| Flag | Description |
|------|-------------|
| `-a, --app <name>` | App name (defaults to active app) |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get workspaces
dvm get ws
dvm get ws -a my-api
dvm get ws -o yaml
```

### `dvm get workspace`

Get a specific workspace.

```bash
dvm get workspace <name> [flags]
```

**Examples:**

```bash
dvm get workspace dev
dvm get ws dev -o yaml
```

### `dvm delete workspace`

Delete a workspace.

```bash
dvm delete workspace <name> [flags]
```

**Aliases:** `ws`

**Flags:**

| Flag | Description |
|------|-------------|
| `-a, --app <name>` | App name (defaults to active app) |
| `-f, --force` | Skip confirmation prompt |

**Examples:**

```bash
dvm delete workspace dev
dvm delete ws dev --force
dvm delete ws dev -a my-api
```

### `dvm use workspace`

Set the active workspace.

```bash
dvm use workspace <name> [flags]
```

**Aliases:** `ws`

**Examples:**

```bash
dvm use workspace dev
dvm use ws dev
dvm use ws none  # Clear active workspace
```

---

## Context

### `dvm get context`

Show current active app and workspace.

```bash
dvm get context [flags]
```

**Aliases:** `ctx`

**Examples:**

```bash
dvm get context
dvm get ctx
dvm get ctx -o yaml
```

---

## Build & Runtime

### `dvm build`

Build workspace container image.

```bash
dvm build [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--force` | Force rebuild even if image exists |
| `--no-cache` | Build without using cache |
| `--target <stage>` | Build target stage (default: `dev`) |

**Examples:**

```bash
dvm build
dvm build --force
dvm build --no-cache
```

### `dvm attach`

Attach to workspace container.

```bash
dvm attach [workspace] [flags]
```

**Examples:**

```bash
dvm attach        # Attach to active workspace
dvm attach dev    # Attach to specific workspace
```

### `dvm detach`

Stop and detach from workspace container.

```bash
dvm detach [workspace] [flags]
```

**Examples:**

```bash
dvm detach        # Detach from active workspace
dvm detach dev    # Detach specific workspace
```

---

## Status

### `dvm status`

Show current context, runtime info, and containers.

```bash
dvm status [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml` |

**Examples:**

```bash
dvm status
dvm status -o json
dvm status -o yaml
```

### `dvm get platforms`

List detected container platforms.

```bash
dvm get platforms [flags]
```

**Aliases:** `plat`

**Examples:**

```bash
dvm get platforms
dvm get plat
dvm get plat -o yaml
```

---

## Configuration

### `dvm apply`

Apply configuration from file.

```bash
dvm apply -f <file> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-f, --file <path>` | Path to YAML file (or URL, or `-` for stdin) |

**Source types:**

| Type | Example |
|------|---------|
| File | `-f workspace.yaml` |
| URL | `-f https://example.com/config.yaml` |
| GitHub | `-f github:user/repo/path.yaml` |
| Stdin | `-f -` |

**Examples:**

```bash
dvm apply -f workspace.yaml
dvm apply -f https://example.com/workspace.yaml
dvm apply -f github:rmkohlman/configs/workspace.yaml
cat workspace.yaml | dvm apply -f -
```

---

## Nvim Resources

### `dvm get nvim plugins`

List nvim plugins.

```bash
dvm get nvim plugins [flags]
```

**Aliases:** `dvm get np`

### `dvm get nvim plugin`

Get specific nvim plugin.

```bash
dvm get nvim plugin <name> [flags]
```

### `dvm delete nvim plugin`

Delete nvim plugin.

```bash
dvm delete nvim plugin <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-f, --force` | Skip confirmation |
| `-w, --workspace <name>` | Remove from workspace (instead of global) |
| `-a, --app <name>` | App for workspace |

---

## Shell Completion

### `dvm completion`

Generate shell completion scripts.

```bash
dvm completion <shell>
```

**Supported shells:** `bash`, `zsh`, `fish`, `powershell`

**Examples:**

```bash
# Bash
dvm completion bash > /etc/bash_completion.d/dvm

# Zsh
dvm completion zsh > "${fpath[1]}/_dvm"

# Fish
dvm completion fish > ~/.config/fish/completions/dvm.fish
```

---

## Version

### `dvm version`

Show version information.

```bash
dvm version
```
