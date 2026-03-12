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

### `dvm admin init`

Initialize DevOpsMaestro (creates database).

```bash
dvm admin init
```

Creates `~/.devopsmaestro/devopsmaestro.db`.

### `dvm delete ecosystem`

Delete an ecosystem and all its contents.

```bash
dvm delete ecosystem <name>
```

**Examples:**

```bash
dvm delete ecosystem my-platform
```

### `dvm delete domain`

Delete a domain and all its apps/workspaces.

```bash
dvm delete domain <ecosystem>/<domain>
```

**Examples:**

```bash
dvm delete domain my-platform/backend
```

### `dvm delete app`

Delete an app and all its workspaces.

```bash
dvm delete app <ecosystem>/<domain>/<app>
```

**Examples:**

```bash
dvm delete app my-platform/backend/my-api
```

### `dvm delete workspace`

Delete a workspace.

```bash
dvm delete workspace <ecosystem>/<domain>/<app>/<workspace>
```

**Examples:**

```bash
dvm delete workspace my-platform/backend/my-api/dev
```

### `dvm create ecosystem`

Create a new ecosystem.

```bash
dvm create ecosystem <name> [flags]
```

**Examples:**

```bash
dvm create ecosystem my-platform
dvm create ecosystem my-platform --description "Main development platform"

# Full path format also supported
dvm create ecosystem my-platform/backend/my-api/dev  # Creates full hierarchy
```

### `dvm create domain`

Create a new domain within an ecosystem.

```bash
dvm create domain <ecosystem>/<domain> [flags]
```

**Examples:**

```bash
dvm create domain my-platform/backend
dvm create domain my-platform/frontend --description "Frontend services"

# Context-aware (if ecosystem is set)
dvm use ecosystem my-platform
dvm create domain backend  # Creates my-platform/backend
```

### `dvm create app`

Create a new app within a domain.

```bash
dvm create app <ecosystem>/<domain>/<app> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--from-cwd` | Use current working directory as app path |
| `--path <path>` | Specific path for the app |
| `--repo <url>` | Git repository URL |
| `--language <name>` | Programming language (go, python, node, etc.) |
| `--description <text>` | App description |

**Examples:**

```bash
# Full path format
dvm create app my-platform/backend/my-api --from-cwd
dvm create app my-platform/backend/my-api --path ~/code/my-api
dvm create app my-platform/backend/my-api --repo https://github.com/user/my-api

# Context-aware (if ecosystem and domain are set)
dvm use ecosystem my-platform
dvm use domain backend
dvm create app my-api --from-cwd --language go
```

### `dvm create workspace`

Create a new workspace for an app.

```bash
dvm create workspace <ecosystem>/<domain>/<app>/<workspace> [flags]
```

**Examples:**

```bash
# Full path format
dvm create workspace my-platform/backend/my-api/dev
dvm create workspace my-platform/backend/my-api/feature-x

# Context-aware (if app is set)
dvm use ecosystem my-platform
dvm use domain backend  
dvm use app my-api
dvm create workspace dev --description "Development environment"
```

### `dvm get ecosystems`

List all ecosystems.

```bash
dvm get ecosystems [flags]
```

**Examples:**

```bash
dvm get ecosystems
dvm get ecosystems -o yaml
```

### `dvm get domains`

List domains in an ecosystem.

```bash
dvm get domains [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--ecosystem <name>` | Ecosystem name (required) |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get domains --ecosystem my-platform
dvm get domains --ecosystem my-platform -o yaml
```

### `dvm get apps`

List apps in a domain.

```bash
dvm get apps [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--domain <name>` | Domain name (defaults to active domain if set) |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get apps --domain backend
dvm get apps --domain my-platform/backend  # Full path format
dvm get apps --domain backend -o yaml
```

### `dvm get workspaces`

List workspaces for an app.

```bash
dvm get workspaces [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--app <name>` | App name (defaults to active app if set) |
| `-A, --all` | List all workspaces across every app |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get workspaces --app my-api
dvm get workspaces --app my-platform/backend/my-api  # Full path format
dvm get workspaces --app my-api -o yaml
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
| `-e, --ecosystem <name>` | Filter by ecosystem name |
| `-d, --domain <name>` | Filter by domain name |
| `-a, --app <name>` | Filter by app name |
| `-w, --workspace <name>` | Filter by workspace name |
| `--force` | Force rebuild even if image exists |
| `--no-cache` | Build without cache (skip registry cache, pull fresh) |
| `--target <stage>` | Build target stage (default: `dev`) |
| `--push` | Push built image to local registry after build |
| `--registry <endpoint>` | Override registry endpoint (default: from config) |

**Examples:**

```bash
# Build active workspace
dvm build

# Force rebuild
dvm build --force

# Build without cache
dvm build --no-cache

# Build specific app's workspace
dvm build -a my-api

# Build and push to local registry
dvm build --push

# Specify ecosystem and app
dvm build -e my-platform -a my-api

# Use specific platform
DVM_PLATFORM=colima dvm build
```

### `dvm attach`

Attach to workspace container (starts if not running; builds image if it doesn't exist).

```bash
dvm attach [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-e, --ecosystem <name>` | Filter by ecosystem name |
| `-d, --domain <name>` | Filter by domain name |
| `-a, --app <name>` | Filter by app name |
| `-w, --workspace <name>` | Filter by workspace name |
| `--no-sync` | Skip syncing git mirror before attach |

**Examples:**

```bash
# Attach to active workspace
dvm attach

# Skip mirror sync
dvm attach --no-sync

# Attach to specific app's workspace
dvm attach -a my-api

# Specify app and workspace name
dvm attach -a my-api -w staging

# Specify ecosystem and app
dvm attach -e my-platform -a my-api
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

---

## Configuration & IaC

### `dvm apply`

Apply configuration from file using Infrastructure as Code.

```bash
dvm apply -f <file> [flags]
```

**Source types:**

| Type | Example | Description |
|------|---------|-------------|
| File | `-f workspace.yaml` | Local file |
| URL | `-f https://example.com/config.yaml` | Remote HTTP/HTTPS |
| GitHub | `-f github:user/repo/path.yaml` | GitHub repository |
| Stdin | `-f -` | Standard input |

**Examples:**

```bash
# Apply local file
dvm apply -f workspace.yaml
dvm apply -f theme.yaml

# Apply from URL
dvm apply -f https://example.com/workspace.yaml

# Apply from GitHub
dvm apply -f github:rmkohlman/configs/workspace.yaml

# Apply from stdin
cat workspace.yaml | dvm apply -f -

# Apply theme IaC
dvm apply -f https://themes.devopsmaestro.io/coolnight-synthwave.yaml
dvm apply -f github:user/themes/my-custom-theme.yaml
```

**Resource Types Supported:**
- `NvimTheme` - Custom theme definitions
- `NvimPlugin` - Plugin configurations  
- `Workspace` - Workspace configurations
- `App` - Application definitions

---

## Credentials

Credentials store references to secrets in the macOS Keychain or host environment variables. They are scoped to a specific resource (ecosystem, domain, app, or workspace).

See [Credential YAML Reference](../reference/credential.md) for full YAML spec and field details.

### `dvm create credential`

Create a new credential.

```bash
dvm create credential <name> [flags]
dvm create cred <name> [flags]        # Alias
```

**Source flags:**

| Flag | Description |
|------|-------------|
| `--source <type>` | Secret source: `keychain` or `env` (required) |
| `--keychain-label <label>` | Keychain entry label — required when `--source=keychain` |
| `--keychain-type <type>` | Keychain item type: `generic` or `internet` (default: `internet`) |
| `--env-var <name>` | Environment variable name — required when `--source=env` |
| `--description <text>` | Human-readable description |
| `--username-var <name>` | Env var for keychain account field (keychain only) |
| `--password-var <name>` | Env var for keychain password field (keychain only) |
| `--service <name>` | **Deprecated.** Use `--keychain-label` instead |

**Scope flags (exactly one required):**

| Flag | Short | Description |
|------|-------|-------------|
| `--ecosystem` | `-e` | Scope to an ecosystem |
| `--domain` | `-d` | Scope to a domain |
| `--app` | `-a` | Scope to an app |
| `--workspace` | `-w` | Scope to a workspace |

**Examples:**

```bash
# GitHub PAT from Passwords app / iCloud Keychain
dvm create credential github-token \
  --source keychain --keychain-label "GitHub PAT" \
  --app my-api

# API key from environment variable
dvm create credential api-key \
  --source env --env-var MY_API_KEY \
  --ecosystem prod

# Docker Hub credentials (split into username + password vars)
dvm create credential docker-registry \
  --source keychain --keychain-label "hub.docker.com" \
  --keychain-type internet \
  --username-var DOCKER_USERNAME \
  --password-var DOCKER_PASSWORD \
  --domain backend

# Generic keychain entry (Keychain Access app)
dvm create cred db-pass \
  --source keychain --keychain-label "Postgres prod" \
  --keychain-type generic \
  --description "Postgres prod password" \
  --app my-api
```

### `dvm get credential`

Get a specific credential by name within a scope.

```bash
dvm get credential <name> [scope-flag]
dvm get cred <name> [scope-flag]       # Alias
```

**Scope flags (exactly one required):** `-e/--ecosystem`, `-d/--domain`, `-a/--app`, `-w/--workspace`

**Examples:**

```bash
dvm get credential github-token --app my-api
dvm get credential api-key --ecosystem prod
dvm get cred db-pass --domain backend
```

### `dvm get credentials`

List credentials by scope or across all scopes.

```bash
dvm get credentials [flags]
dvm get creds [flags]                  # Alias
```

| Flag | Short | Description |
|------|-------|-------------|
| `--all` | `-A` | List all credentials across every scope |
| `--ecosystem` | `-e` | Filter to an ecosystem |
| `--domain` | `-d` | Filter to a domain |
| `--app` | `-a` | Filter to an app |
| `--workspace` | `-w` | Filter to a workspace |

**Examples:**

```bash
dvm get credentials -A             # All credentials
dvm get credentials --app my-api   # Credentials for a specific app
```

### `dvm delete credential`

Delete a credential by name within a scope.

```bash
dvm delete credential <name> [scope-flag] [flags]
dvm delete cred <name> [scope-flag] [flags]    # Alias
```

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip confirmation prompt |
| `--ecosystem` | `-e` | Scope to an ecosystem |
| `--domain` | `-d` | Scope to a domain |
| `--app` | `-a` | Scope to an app |
| `--workspace` | `-w` | Scope to a workspace |

**Examples:**

```bash
dvm delete credential github-token --app my-api
dvm delete credential api-key --ecosystem prod --force
dvm delete cred db-pass --domain backend -f
```

---

## Themes (NEW in v0.12.0)

### `dvm set theme`

Set Neovim theme at any hierarchy level with cascading inheritance.

```bash
dvm set theme <theme-name> [flags]
```

**Hierarchy levels (one required):**

| Flag | Description |
|------|-------------|
| `--workspace <name>` | Set theme at workspace level |
| `--app <name>` | Set theme at app level |
| `--domain <name>` | Set theme at domain level |
| `--ecosystem <name>` | Set theme at ecosystem level |

**Other flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table`, `colored` |
| `--dry-run` | Preview changes without applying |
| `--show-cascade` | Show theme cascade effect |

**Theme Cascade:**
Themes inherit down the hierarchy unless overridden:
```
Ecosystem → Domain → App → Workspace
```

**Examples:**

```bash
# Set workspace theme (highest priority)
dvm set theme coolnight-synthwave --workspace dev

# Set app theme (applies to all workspaces in app unless overridden)
dvm set theme tokyonight-night --app my-api

# Set domain theme (applies to all apps/workspaces in domain)
dvm set theme gruvbox-dark --domain backend

# Set ecosystem theme (applies globally unless overridden)
dvm set theme catppuccin-mocha --ecosystem my-platform

# Clear theme to inherit from parent level
dvm set theme "" --workspace dev

# Preview changes
dvm set theme coolnight-ocean --workspace dev --dry-run

# Show cascade effect
dvm set theme gruvbox-dark --app my-api --show-cascade
```

**Available Themes:**
- **Library themes**: 34+ instantly available (coolnight-ocean, tokyonight-night, catppuccin-mocha, etc.)
- **CoolNight variants**: 21 algorithmic variants (ocean, synthwave, matrix, sunset, etc.)
- **User themes**: Custom themes via `dvm apply -f theme.yaml`

Use `dvm get nvim themes` to see all available themes.

---

## Nvim Resources

### `dvm get nvim plugins`

List nvim plugins from global library or workspace-specific.

```bash
dvm get nvim plugins [flags]
```

**Aliases:** `dvm get np`

**Flags:**

| Flag | Description |
|------|-------------|
| `-w, --workspace <name>` | Filter by workspace |
| `-a, --app <name>` | App for workspace |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get nvim plugins                  # List all global plugins
dvm get nvim plugins -w dev           # List plugins for workspace 'dev'
dvm get nvim plugins -a myapp -w dev  # Explicit app and workspace
dvm get nvim plugins -o yaml          # Output as YAML
```

### `dvm get nvim plugin`

Get a specific nvim plugin by name.

```bash
dvm get nvim plugin <name> [flags]
```

**Examples:**

```bash
dvm get nvim plugin telescope
dvm get nvim plugin telescope -o yaml
dvm get nvim plugin lspconfig -o json
```

### `dvm get nvim themes`

List nvim themes from user store and embedded library (34+ themes available instantly).

```bash
dvm get nvim themes [flags]
```

**Aliases:** `dvm get nt`

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get nvim themes                    # Shows user + library themes
dvm get nvim themes -o yaml            # YAML format
```

### `dvm get nvim theme`

Get a specific nvim theme by name.

```bash
dvm get nvim theme <name> [flags]
```

**Examples:**

```bash
dvm get nvim theme coolnight-ocean     # Get specific library theme
dvm get nvim theme coolnight-ocean -o yaml  # Export as YAML for sharing
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
| `credential` | `cred` |
| `credentials` | `cred`, `creds` |
| `context` | `ctx` |
| `platforms` | `plat` |
| `nvim plugins` | `np` |
| `nvim themes` | `nt` |

---

## Quick Reference: Typical Workflow

```bash
# 1. Initialize
dvm admin init

# 2. Set up hierarchy (new path-based approach)
dvm create ecosystem my-platform
dvm create domain my-platform/backend
dvm create app my-platform/backend/my-api --from-cwd
dvm create workspace my-platform/backend/my-api/dev

# 3. Set theme at any level
dvm set theme coolnight-synthwave --workspace my-platform/backend/my-api/dev
# OR set at app level (affects all workspaces)
dvm set theme tokyonight-night --app my-platform/backend/my-api

# 4. Build and attach
dvm build my-platform/backend/my-api/dev
dvm attach my-platform/backend/my-api/dev

# 5. Apply IaC themes and configs
dvm apply -f https://themes.example.com/custom-theme.yaml
dvm apply -f workspace-config.yaml

# 6. Check status
dvm status
dvm get context
```
