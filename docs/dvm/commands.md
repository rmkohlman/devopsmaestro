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

## Projects

### `dvm create project`

Create a new project.

```bash
dvm create project <name> [flags]
```

**Aliases:** `proj`

**Flags:**

| Flag | Description |
|------|-------------|
| `--from-cwd` | Use current working directory as project path |
| `--path <path>` | Specific path for the project |
| `--description <text>` | Project description |

**Examples:**

```bash
# Create from current directory
dvm create project my-api --from-cwd

# Create with explicit path
dvm create project my-api --path ~/Developer/my-api

# Create with description
dvm create project my-api --from-cwd --description "REST API service"
```

### `dvm get projects`

List all projects.

```bash
dvm get projects [flags]
```

**Aliases:** `proj`

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get projects
dvm get proj
dvm get proj -o yaml
dvm get proj -o json
```

### `dvm get project`

Get a specific project.

```bash
dvm get project <name> [flags]
```

**Examples:**

```bash
dvm get project my-api
dvm get project my-api -o yaml
```

### `dvm delete project`

Delete a project.

```bash
dvm delete project <name> [flags]
```

**Aliases:** `proj`

**Flags:**

| Flag | Description |
|------|-------------|
| `-f, --force` | Skip confirmation prompt |

**Examples:**

```bash
dvm delete project my-api
dvm delete proj my-api --force
```

### `dvm use project`

Set the active project.

```bash
dvm use project <name> [flags]
```

**Aliases:** `proj`

**Flags:**

| Flag | Description |
|------|-------------|
| `--clear` | Clear active project |

**Examples:**

```bash
dvm use project my-api
dvm use proj my-api
dvm use project --clear
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
| `-p, --project <name>` | Project name (defaults to active project) |
| `--description <text>` | Workspace description |
| `--image <name>` | Custom image name |

**Examples:**

```bash
dvm create workspace dev
dvm create ws dev
dvm create ws dev -p my-api
dvm create ws staging --description "Staging environment"
```

### `dvm get workspaces`

List workspaces in a project.

```bash
dvm get workspaces [flags]
```

**Aliases:** `ws`

**Flags:**

| Flag | Description |
|------|-------------|
| `-p, --project <name>` | Project name (defaults to active project) |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get workspaces
dvm get ws
dvm get ws -p my-api
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
| `-p, --project <name>` | Project name (defaults to active project) |
| `-f, --force` | Skip confirmation prompt |

**Examples:**

```bash
dvm delete workspace dev
dvm delete ws dev --force
dvm delete ws dev -p my-api
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

Show current active project and workspace.

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
| `-p, --project <name>` | Project for workspace |

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
