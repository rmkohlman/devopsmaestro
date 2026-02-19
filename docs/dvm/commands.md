# dvm Commands Reference

Complete reference for all `dvm` commands.

## Object Hierarchy

DevOpsMaestro uses a hierarchical structure:

```
Ecosystem → Domain → App → Workspace
```

| Level | Purpose | Example |
|-------|---------|---------|
| **Ecosystem** | Top-level platform grouping | `my-platform` |
| **Domain** | Bounded context (group of related apps) | `backend`, `frontend` |
| **App** | A codebase/application | `my-api`, `web-app` |
| **Workspace** | Development environment for an app | `dev`, `feature-x` |

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

## Ecosystems

An ecosystem is the top-level grouping in the hierarchy.

### `dvm create ecosystem`

Create a new ecosystem.

```bash
dvm create ecosystem <name> [flags]
```

**Aliases:** `eco`

**Flags:**

| Flag | Description |
|------|-------------|
| `--description <text>` | Ecosystem description |

**Examples:**

```bash
dvm create ecosystem my-platform
dvm create eco my-platform                    # Short form
dvm create ecosystem my-platform --description "Main development platform"
```

### `dvm get ecosystems`

List all ecosystems.

```bash
dvm get ecosystems [flags]
```

**Aliases:** `eco`

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get ecosystems
dvm get eco                    # Short form
dvm get eco -o yaml
```

### `dvm get ecosystem`

Get a specific ecosystem.

```bash
dvm get ecosystem <name> [flags]
```

**Aliases:** `eco`

**Examples:**

```bash
dvm get ecosystem my-platform
dvm get eco my-platform -o yaml
```

### `dvm delete ecosystem`

Delete an ecosystem.

```bash
dvm delete ecosystem <name> [flags]
```

**Aliases:** `eco`

> **Warning:** This will also delete all domains and apps within the ecosystem.

**Examples:**

```bash
dvm delete ecosystem my-platform
dvm delete eco my-platform      # Short form
```

### `dvm use ecosystem`

Set the active ecosystem.

```bash
dvm use ecosystem <name>
```

**Aliases:** `eco`

Use `none` as the name to clear the ecosystem context (also clears domain and app).

**Examples:**

```bash
dvm use ecosystem my-platform
dvm use eco my-platform         # Short form
dvm use ecosystem none          # Clear ecosystem context
```

---

## Domains

A domain represents a bounded context within an ecosystem, grouping related applications.

### `dvm create domain`

Create a new domain within an ecosystem.

```bash
dvm create domain <name> [flags]
```

**Aliases:** `dom`

**Flags:**

| Flag | Description |
|------|-------------|
| `--ecosystem <name>` | Ecosystem name (defaults to active ecosystem) |
| `--description <text>` | Domain description |

**Examples:**

```bash
dvm create domain backend
dvm create dom backend                        # Short form
dvm create domain backend --ecosystem my-platform
dvm create domain backend --description "Backend services"
```

### `dvm get domains`

List domains in an ecosystem.

```bash
dvm get domains [flags]
```

**Aliases:** `dom`

**Flags:**

| Flag | Description |
|------|-------------|
| `-e, --ecosystem <name>` | Ecosystem name (defaults to active ecosystem) |
| `--all` | List domains from all ecosystems |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get domains                              # List in active ecosystem
dvm get dom                                  # Short form
dvm get domains --ecosystem my-platform      # List in specific ecosystem
dvm get domains --all                        # List all domains
dvm get dom -o yaml
```

### `dvm get domain`

Get a specific domain.

```bash
dvm get domain <name> [flags]
```

**Aliases:** `dom`

**Flags:**

| Flag | Description |
|------|-------------|
| `-e, --ecosystem <name>` | Ecosystem name (defaults to active ecosystem) |

**Examples:**

```bash
dvm get domain backend
dvm get dom backend -o yaml
dvm get domain backend --ecosystem my-platform
```

### `dvm delete domain`

Delete a domain.

```bash
dvm delete domain <name> [flags]
```

**Aliases:** `dom`

**Flags:**

| Flag | Description |
|------|-------------|
| `-e, --ecosystem <name>` | Ecosystem name (defaults to active ecosystem) |

> **Warning:** This will also delete all apps within the domain.

**Examples:**

```bash
dvm delete domain backend
dvm delete dom backend          # Short form
dvm delete domain backend --ecosystem my-platform
```

### `dvm use domain`

Set the active domain.

```bash
dvm use domain <name>
```

**Aliases:** `dom`

Requires an active ecosystem to be set first. Use `none` as the name to clear the domain context (also clears app).

**Examples:**

```bash
dvm use domain backend
dvm use dom backend             # Short form
dvm use domain none             # Clear domain context
```

---

## Apps

An app represents a codebase/application within a domain.

### `dvm create app`

Create a new app within a domain.

```bash
dvm create app <name> [flags]
```

**Aliases:** `application`

**Flags:**

| Flag | Description |
|------|-------------|
| `--from-cwd` | Use current working directory as app path |
| `--path <path>` | Specific path for the app |
| `--domain <name>` | Domain name (defaults to active domain) |
| `--description <text>` | App description |

**Examples:**

```bash
# Create from current directory
dvm create app my-api --from-cwd

# Create with explicit path
dvm create app my-api --path ~/Developer/my-api

# Create in a specific domain
dvm create app my-api --from-cwd --domain backend

# Create with description
dvm create app my-api --from-cwd --description "REST API service"
```

### `dvm get apps`

List apps in a domain.

```bash
dvm get apps [flags]
```

**Aliases:** `app`, `application`, `applications`

**Flags:**

| Flag | Description |
|------|-------------|
| `-d, --domain <name>` | Domain name (defaults to active domain) |
| `--all` | List apps from all domains |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get apps                    # List in active domain
dvm get apps --domain backend   # List in specific domain
dvm get apps --all              # List all apps
dvm get apps -o yaml
```

### `dvm get app`

Get a specific app.

```bash
dvm get app <name> [flags]
```

**Aliases:** `application`

**Flags:**

| Flag | Description |
|------|-------------|
| `-d, --domain <name>` | Domain name (defaults to active domain) |

**Examples:**

```bash
dvm get app my-api
dvm get app my-api -o yaml
dvm get app my-api --domain backend
```

### `dvm delete app`

Delete an app.

```bash
dvm delete app <name> [flags]
```

**Aliases:** `application`

**Flags:**

| Flag | Description |
|------|-------------|
| `-d, --domain <name>` | Domain name (defaults to active domain) |

**Examples:**

```bash
dvm delete app my-api
dvm delete app my-api --domain backend
```

### `dvm use app`

Set the active app.

```bash
dvm use app <name>
```

**Aliases:** `a`

Use `none` as the name to clear the app context (also clears workspace).

**Examples:**

```bash
dvm use app my-api
dvm use a my-api               # Short form
dvm use app none               # Clear app context
```

---

## Workspaces

A workspace is an isolated development environment within an app.

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
dvm create ws dev                             # Short form
dvm create ws dev -a my-api                   # Specific app
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
dvm get ws                      # Short form
dvm get ws -a my-api
dvm get ws -o yaml
```

### `dvm get workspace`

Get a specific workspace.

```bash
dvm get workspace <name> [flags]
```

**Aliases:** `ws`

**Flags:**

| Flag | Description |
|------|-------------|
| `-a, --app <name>` | App name (defaults to active app) |

**Examples:**

```bash
dvm get workspace dev
dvm get ws dev -o yaml
dvm get ws dev -a my-api
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
dvm use workspace <name>
```

**Aliases:** `ws`

Requires an active app to be set first. Use `none` as the name to clear the workspace context (keeps app).

**Examples:**

```bash
dvm use workspace dev
dvm use ws dev                  # Short form
dvm use ws none                 # Clear workspace context
```

---

## Context

### `dvm get context`

Show current active ecosystem, domain, app, and workspace.

```bash
dvm get context [flags]
```

**Aliases:** `ctx`

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml` |

**Examples:**

```bash
dvm get context
dvm get ctx
dvm get ctx -o yaml
```

### `dvm use --clear`

Clear all context (ecosystem, domain, app, and workspace).

```bash
dvm use --clear
```

**Examples:**

```bash
dvm use --clear
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

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

### `dvm get nvim themes`

List nvim themes from user store and embedded library (34+ themes available instantly).

```bash
dvm get nvim themes [flags]
```

**Aliases:** `dvm get nt`

**Examples:**
```bash
dvm get nvim themes                    # Shows user + library themes
dvm get nvim theme coolnight-ocean     # Get specific library theme (no install needed)
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

### `dvm delete nvim plugin`

Delete nvim plugin.

```bash
dvm delete nvim plugin <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-f, --force` | Skip confirmation |
| `-w, --workspace <name>` | Remove from workspace (instead of global library) |
| `-a, --app <name>` | App for workspace |

**Examples:**

```bash
dvm delete nvim plugin telescope              # Delete from global library
dvm delete nvim plugin telescope --force      # Skip confirmation
dvm delete nvim plugin -w dev telescope       # Remove from workspace 'dev'
```

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

---

## Quick Reference: Command Aliases

| Full Command | Alias |
|--------------|-------|
| `ecosystem` | `eco` |
| `domain` | `dom` |
| `app` | `a`, `application` |
| `workspace` | `ws` |
| `context` | `ctx` |
| `platforms` | `plat` |
| `nvim plugins` | `np` |
| `nvim themes` | `nt` |

---

## Quick Reference: Typical Workflow

```bash
# 1. Initialize
dvm init

# 2. Set up hierarchy
dvm create ecosystem my-platform
dvm create domain backend
dvm create app my-api --from-cwd
dvm create workspace dev

# 3. Build and attach
dvm build && dvm attach

# 4. Check status
dvm status
dvm get context
```
