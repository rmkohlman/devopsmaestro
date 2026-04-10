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

### `dvm get all`

Show a kubectl-style overview of all resources. By default, resources are scoped to the active context (ecosystem, domain, or app). Use `-A` to ignore context and show everything.

```bash
dvm get all [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-e, --ecosystem <name>` | Filter to a specific ecosystem |
| `-d, --domain <name>` | Filter to a specific domain (requires ecosystem context) |
| `-a, --app <name>` | Filter to a specific app (requires domain context) |
| `-A, --all` | Show all resources (ignore active context) |
| `-o, --output <format>` | Output format: `json`, `yaml`, `wide`, `table` (default: human-readable table) |

**Sections displayed:** Ecosystems, Domains, Apps, Workspaces, Credentials, Registries, Git Repos, Nvim Plugins, Nvim Themes, Nvim Packages, Terminal Prompts, Terminal Packages. Empty sections show `(none)`.

Global resources (Registries, Git Repos, Nvim Plugins, Nvim Themes, Nvim Packages, Terminal Prompts, Terminal Packages) are always shown in table output regardless of scope.

**YAML/JSON output (`-o yaml` / `-o json`):**

`-o yaml` and `-o json` produce a `kind: List` document — a kubectl-style List wrapper where each item is the full resource YAML (identical to `dvm get <resource> <name> -o yaml`). Items are ordered for apply-safe dependency: Ecosystems → Domains → Apps → GitRepos → Registries → Credentials → Workspaces → NvimPlugins → NvimThemes → NvimPackages → TerminalPrompts → TerminalPackages.

When scope flags (`-e`/`-d`/`-a`) are used with `-o yaml/json`, global resources are excluded from the List output — only hierarchical resources (ecosystems, domains, apps, workspaces, credentials) matching the scope are exported. Table output (`-o wide` or no `-o`) always shows global resources.

The output is designed for round-trip use with `dvm apply -f`:

```bash
# Export all resources to a backup file
dvm get all -A -o yaml > backup.yaml

# Restore from backup
dvm apply -f backup.yaml

# Or pipe directly
dvm get all -A -o yaml | dvm apply -f -
```

**Examples:**

```bash
dvm get all                         # Show resources in active scope
dvm get all -A                      # Show all resources (ignore context)
dvm get all -e prod                 # Show resources in 'prod' ecosystem
dvm get all -e prod -d backend      # Show resources in 'backend' domain
dvm get all -o wide                 # Show additional columns
dvm get all -o json                 # Output as JSON (kind: List format)
dvm get all -o yaml                 # Output as YAML (kind: List format)
dvm get all -A -o yaml > backup.yaml  # Export all resources for backup
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

The build command:
- Detects the app language and generates a Dockerfile with dev tools
- Emits `ARG` declarations for all `spec.build.args` keys (not `ENV` — credentials are not persisted in image layers)
- Injects CA certificates from MaestroVault when `spec.build.caCerts` is configured — certificates are written to `/usr/local/share/ca-certificates/custom/`, `update-ca-certificates` is run, and `SSL_CERT_FILE`, `REQUESTS_CA_BUNDLE`, and `NODE_EXTRA_CA_CERTS` are set
- Sets the `USER` directive to the value of `container.user` (defaults to `dev` if unset)
- Builds the image using the detected container platform and tags it as `dvm-<workspace>-<app>:<timestamp>`
- Optionally pushes to local registry cache after build

**Supported platforms:** OrbStack, Docker Desktop, Podman (Docker API); Colima with containerd (BuildKit API). Use the `DVM_PLATFORM` environment variable to select a specific platform.

**Registry integration:** If `registry.enabled` is `true` in config and lifecycle is `on-demand` or `persistent`, the registry is automatically started before building to provide image caching.

**Hierarchy flags (`-A/--all`, `-e`, `-d`, `-a`, `-w`) — NEW in v0.74.0; additive behavior added in [#213](https://github.com/rmkohlman/devopsmaestro/issues/213):**

Scope flags allow building specific workspaces without first running `dvm use`. Use `-A/--all` to build every workspace across all apps, domains, and ecosystems in parallel. Scope flags (`-e`, `-d`, `-a`, `-w`) **compose additively** with `--all` — they narrow the set of workspaces to build rather than conflicting with it. `dvm build --all` does not require an active workspace to be set.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--all` | `-A` | bool | `false` | Build all workspaces across all apps, domains, and ecosystems |
| `--ecosystem <name>` | `-e` | string | `""` | Filter by ecosystem name |
| `--domain <name>` | `-d` | string | `""` | Filter by domain name |
| `--app <name>` | `-a` | string | `""` | Filter by app name |
| `--workspace <name>` | `-w` | string | `""` | Filter by workspace name |
| `--concurrency <n>` | | int | `4` | Maximum number of parallel builds when building multiple workspaces |
| `--detach` | | bool | `false` | Run the build session in the background; return immediately and monitor with `dvm build status` |
| `--force` | | bool | `false` | Force rebuild even if image exists |
| `--no-cache` | | bool | `false` | Build without using cache (skip registry cache) |
| `--target <stage>` | | string | `"dev"` | Build target stage |
| `--push` | | bool | `false` | Push built image to local registry after build |
| `--registry <endpoint>` | | string | `""` | Override registry endpoint (default: from config) |
| `--timeout <duration>` | | duration | `30m` | Timeout for the build operation (e.g., `30m`, `1h`) |
| `--dry-run` | | bool | `false` | Preview what would be built without executing |

**Examples:**

```bash
# Build active workspace
dvm build

# Force rebuild
dvm build --force

# Build without cache
dvm build --no-cache

# Build a specific app's workspace using hierarchy flags
dvm build -a my-api

# Build all workspaces in a specific domain
dvm build -d backend

# Build all workspaces across every app and ecosystem
dvm build --all

# Build all workspaces in a specific ecosystem (additive scoping)
dvm build --all --ecosystem beans-modules

# Build all workspaces in an ecosystem's domain (additive scoping)
dvm build --all --ecosystem beans-modules --domain services

# Run parallel build in the background (monitor with dvm build status)
dvm build --all --detach

# Run background build with higher concurrency
dvm build --all --ecosystem beans-modules --detach --concurrency 8

# Preview what --all would build without executing
dvm build --all --dry-run

# Build and push to local registry
dvm build --push

# Use a specific platform
DVM_PLATFORM=colima dvm build
```

### `dvm build status`

Show the status of a build session. Updated in [#217](https://github.com/rmkohlman/devopsmaestro/issues/217) to use persisted session data.

```bash
dvm build status [flags]
```

Displays a table of all workspaces in a build session with their status, duration, and any errors. By default shows the most recent session. When no sessions exist, outputs a hint to run `dvm build --all` to start one.

**Table columns:** `WORKSPACE`, `APP`, `STATUS`, `DURATION`, `ERROR`

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--session-id <uuid>` | | string | `""` | Show a specific build session by UUID |
| `--history` | | bool | `false` | List the 10 most recent build sessions |
| `--output <format>` | `-o` | string | `""` | Output format: `table`, `json`, `yaml` |

**Examples:**

```bash
# Show the latest build session (human-readable table)
dvm build status

# Show a specific session by UUID
dvm build status --session-id 550e8400-e29b-41d4-a716-446655440000

# List the 10 most recent sessions
dvm build status --history

# Output as JSON
dvm build status -o json

# Output a specific session as YAML
dvm build status --session-id 550e8400-e29b-41d4-a716-446655440000 -o yaml
```

### `dvm detach`

Stop and detach from a workspace container.

```bash
dvm detach [flags]
```

Stops the currently active workspace container. The container is stopped but not removed, so you can quickly re-attach later with `dvm attach`. Use `-A/--all` to stop all DVM workspace containers at once.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--all` | `-A` | bool | `false` | Stop all DVM workspace containers |
| `--ecosystem <name>` | `-e` | string | `""` | Filter by ecosystem name |
| `--domain <name>` | `-d` | string | `""` | Filter by domain name |
| `--app <name>` | `-a` | string | `""` | Filter by app name |
| `--workspace <name>` | `-w` | string | `""` | Filter by workspace name |
| `--timeout <duration>` | | duration | `5m` | Timeout for the detach operation (e.g., `5m`, `30s`) |
| `--dry-run` | | bool | `false` | Preview which containers would be stopped without stopping them |

**Examples:**

```bash
# Stop the active workspace container
dvm detach

# Stop a specific app's workspace using hierarchy flags
dvm detach -a my-api

# Stop a specific workspace within an ecosystem
dvm detach -e my-platform -a my-api

# Stop all DVM workspace containers
dvm detach --all

# Preview what would be stopped without stopping anything
dvm detach --all --dry-run
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

**`kind: List` support:**

`dvm apply -f` accepts `kind: List` documents produced by `dvm get all -o yaml`. Each item in the list is applied individually in document order. If an item fails, the error is reported and processing continues with the next item (continue-on-error). The command exits with a non-zero code if any item failed.

This enables infrastructure-as-code backup and restore:

```bash
# Backup
dvm get all -A -o yaml > backup.yaml

# Restore
dvm apply -f backup.yaml
```

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

# Apply a List document (e.g., from dvm get all -o yaml)
dvm apply -f backup.yaml
dvm get all -A -o yaml | dvm apply -f -

# Apply theme IaC
dvm apply -f https://themes.devopsmaestro.io/coolnight-synthwave.yaml
dvm apply -f github:user/themes/my-custom-theme.yaml
```

**Resource Types Supported:**
- `Ecosystem` - Ecosystem definitions
- `Domain` - Domain definitions
- `App` - Application definitions
- `Workspace` - Workspace configurations
- `Credential` - Credential references
- `Registry` - Container registry configurations
- `GitRepo` - Git repository definitions
- `NvimTheme` - Custom theme definitions
- `NvimPlugin` - Plugin configurations
- `NvimPackage` - Neovim package definitions
- `TerminalPrompt` - Terminal prompt configurations
- `TerminalPackage` - Terminal package definitions
- `List` - Multi-resource list document (applies each item individually)
- `CustomResourceDefinition` - Custom resource type definitions

---

## Credentials

Credentials store references to secrets in MaestroVault or host environment variables. They are scoped to a specific resource (ecosystem, domain, app, or workspace).

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
| `--source <type>` | Secret source: `vault` or `env` (required) |
| `--vault-secret <name>` | MaestroVault secret name — required when `--source=vault` |
| `--vault-env <name>` | MaestroVault environment |
| `--vault-username-secret <name>` | MaestroVault secret name for username |
| `--vault-field <ENV_VAR=field>` | Map a vault field to an env var (repeatable) |
| `--env-var <name>` | Environment variable name — required when `--source=env` |
| `--description <text>` | Human-readable description |
| `--username-var <name>` | Env var for username (vault only) |
| `--password-var <name>` | Env var for password (vault only) |

**Scope flags (exactly one required):**

| Flag | Short | Description |
|------|-------|-------------|
| `--ecosystem` | `-e` | Scope to an ecosystem |
| `--domain` | `-d` | Scope to a domain |
| `--app` | `-a` | Scope to an app |
| `--workspace` | `-w` | Scope to a workspace |

**Examples:**

```bash
# GitHub PAT from MaestroVault
dvm create credential github-token \
  --source vault --vault-secret "github-pat" \
  --app my-api

# API key from environment variable
dvm create credential api-key \
  --source env --env-var MY_API_KEY \
  --ecosystem prod

# Docker Hub credentials with separate username and password vars
dvm create credential docker-registry \
  --source vault --vault-secret "hub.docker.com" \
  --username-var DOCKER_USERNAME \
  --password-var DOCKER_PASSWORD \
  --domain backend

# Vault secret with explicit field mapping
dvm create cred db-pass \
  --source vault --vault-secret "postgres-prod" \
  --vault-field DB_PASSWORD=password \
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

## Registries

Registries are local package mirrors (OCI, Go, Python, npm, HTTP proxy) managed as named resources. Each registry has a type, port, and lifecycle mode.

**Registry types:**

| Type | Binary | Default Port | Purpose |
|------|--------|-------------|---------|
| `zot` | zot | 5000 | OCI container image registry |
| `athens` | athens | 3000 | Go module proxy |
| `devpi` | devpi | 3141 | Python package index |
| `verdaccio` | verdaccio | 4873 | npm private registry |
| `squid` | squid | 3128 | HTTP/HTTPS caching proxy |

**Lifecycle modes:**

| Mode | Behavior |
|------|---------|
| `persistent` | Always running (starts with system) |
| `on-demand` | Starts when needed, stops after idle timeout |
| `manual` | User controls start/stop explicitly (default) |

### `dvm create registry`

Create a new package registry.

```bash
dvm create registry <name> [flags]
dvm create reg <name> [flags]        # Alias
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--type <type>` | `-t` | string | — | **Required.** Registry type: `zot`, `athens`, `devpi`, `verdaccio`, `squid` |
| `--port <number>` | `-p` | int | type-specific | Port number override |
| `--lifecycle <mode>` | `-l` | string | `manual` | Lifecycle mode: `persistent`, `on-demand`, `manual` |
| `--description <text>` | `-d` | string | — | Human-readable description |
| `--version <version>` | — | string | latest | Desired binary version (e.g., `2.1.15`) |

**Examples:**

```bash
# Create a zot OCI registry with default port (5000)
dvm create registry my-zot --type zot

# Create an npm registry on a custom port
dvm create registry my-npm --type verdaccio --port 4880

# Create a Go proxy that stays running permanently
dvm create registry go-proxy --type athens --lifecycle persistent

# Create a Python index with a description
dvm create registry pypi --type devpi --description "Internal Python packages"
```

### `dvm delete registry`

Delete a registry by name. Auto-stops the registry if it is running before removing the database record.

```bash
dvm delete registry <name> [flags]
dvm delete reg <name> [flags]        # Alias
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--force` | `-f` | bool | false | Skip confirmation prompt |

**Examples:**

```bash
# Delete with confirmation prompt
dvm delete registry my-zot

# Delete without confirmation
dvm delete registry my-zot --force
dvm delete reg my-zot -f
```

### `dvm get registries`

List all registries.

```bash
dvm get registries [flags]
dvm get regs [flags]                 # Alias
dvm get reg [flags]                  # Alias
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `-o, --output <format>` | — | string | table | Output format: `table`, `wide`, `json`, `yaml` |

**Examples:**

```bash
dvm get registries
dvm get regs                         # Short form
dvm get registries -o wide           # Show CREATED column
dvm get registries -o yaml
dvm get registries -o json
```

### `dvm get registry`

Get details for a specific registry.

```bash
dvm get registry <name> [flags]
dvm get reg <name> [flags]           # Alias
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `-o, --output <format>` | — | string | table | Output format: `table`, `json`, `yaml` |

**Examples:**

```bash
dvm get registry my-zot
dvm get reg my-zot                   # Short form
dvm get registry my-zot -o yaml
dvm get registry my-zot -o json
```

### `dvm start registry`

Start a registry instance.

```bash
dvm start registry <name>
dvm start reg <name>                 # Alias
```

The registry binary is automatically downloaded on first use. If the registry is already running, prints the current endpoint and exits cleanly.

**Examples:**

```bash
# Start a registry
dvm start registry my-zot

# Short form
dvm start reg my-zot
```

### `dvm stop registry`

Stop a running registry instance.

```bash
dvm stop registry <name> [flags]
dvm stop reg <name> [flags]          # Alias
```

By default sends SIGTERM and waits for graceful shutdown. If the registry is already stopped, exits cleanly with no error.

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--force` | — | bool | false | Force kill with SIGKILL instead of graceful SIGTERM |

**Examples:**

```bash
# Graceful stop
dvm stop registry my-zot

# Force kill
dvm stop registry my-zot --force
dvm stop reg my-zot --force
```

### `dvm registry enable`

Enable a registry type. Creates a new registry of that type (with default settings) if one does not already exist, and sets it as the default for that type.

```bash
dvm registry enable <type> [flags]
```

**Valid types:** `oci`, `pypi`, `npm`, `go`, `http`

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--lifecycle <mode>` | string | `on-demand` | Lifecycle mode: `persistent`, `on-demand`, `manual` |

**Examples:**

```bash
# Enable the OCI registry type (creates a zot registry if none exists)
dvm registry enable oci

# Enable with persistent lifecycle
dvm registry enable oci --lifecycle persistent

# Enable the Python index type
dvm registry enable pypi

# Enable the Go module proxy
dvm registry enable go --lifecycle on-demand
```

### `dvm registry disable`

Disable a registry type by clearing its default assignment. Does not delete or stop any existing registry instances.

```bash
dvm registry disable <type>
```

**Valid types:** `oci`, `pypi`, `npm`, `go`, `http`

**Examples:**

```bash
dvm registry disable oci
dvm registry disable pypi
dvm registry disable npm
```

### `dvm registry set-default`

Set the default registry to use for a specific type. The registry must already exist and its type must match the alias.

```bash
dvm registry set-default <type> <registry-name>
```

**Examples:**

```bash
# Set the default OCI registry
dvm registry set-default oci zot-local

# Set the default Python index
dvm registry set-default pypi devpi-local

# Set a custom npm registry as default
dvm registry set-default npm my-verdaccio
```

### `dvm registry get-defaults`

Display the default registry configured for each type.

```bash
dvm registry get-defaults [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `-o, --output <format>` | — | string | table | Output format: `table`, `json`, `yaml` |

Output shows `TYPE`, `REGISTRY`, `ENDPOINT`, and `STATUS` columns. Types with no default configured show `-` for registry/endpoint and `not configured` for status.

**Examples:**

```bash
dvm registry get-defaults
dvm registry get-defaults -o yaml
dvm registry get-defaults -o json
```

---

## Git Repos

Git repos are bare-clone mirrors of remote repositories. They are stored locally at `~/.devopsmaestro/repos/` and can be shared across workspaces. Workspaces clone from the local mirror (fast) instead of directly from the remote.

### `dvm create gitrepo`

Create a git repository mirror configuration and perform an initial clone.

```bash
dvm create gitrepo <name> [flags]
dvm create repo <name> [flags]       # Alias
dvm create gr <name> [flags]         # Alias
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--url <url>` | string | — | **Required.** Git repository URL (HTTPS or SSH) |
| `--auth-type <type>` | string | `none` | Authentication type: `none`, `ssh`, `token` |
| `--credential <name>` | string | — | Credential name for authentication (requires `--auth-type` other than `none`) |
| `--no-sync` | bool | false | Skip the initial clone (register only) |
| `--default-ref <branch>` | string | auto-detected | Default branch name (auto-detected from remote if not specified) |
| `--dry-run` | bool | false | Preview what would be created without making changes |

**Examples:**

```bash
# Create a public repository mirror (auto-detects default branch)
dvm create gitrepo my-repo --url https://github.com/org/repo.git

# Create with SSH authentication
dvm create gitrepo private-repo \
  --url git@github.com:org/repo.git \
  --auth-type ssh \
  --credential github-ssh

# Register only, skip the initial clone
dvm create gitrepo my-repo --url https://github.com/org/repo.git --no-sync

# Preview without creating
dvm create gitrepo my-repo --url https://github.com/org/repo.git --dry-run
```

### `dvm delete gitrepo`

Delete a git repository mirror. Removes both the database record and the local mirror directory by default.

```bash
dvm delete gitrepo <name> [flags]
dvm delete repo <name> [flags]       # Alias
dvm delete gr <name> [flags]         # Alias
```

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--keep-mirror` | bool | false | Remove only the database record; keep the local mirror directory on disk |
| `--force` | bool | false | Skip confirmation prompt |
| `--dry-run` | bool | false | Preview what would be deleted without making changes |

**Examples:**

```bash
# Delete repo record and mirror directory (with confirmation)
dvm delete gitrepo my-repo

# Delete only the database record, keep the mirror on disk
dvm delete gitrepo my-repo --keep-mirror

# Skip confirmation prompt
dvm delete gitrepo my-repo --force

# Preview without deleting
dvm delete gitrepo my-repo --dry-run
```

### `dvm get gitrepos`

List all git repository mirrors.

```bash
dvm get gitrepos [flags]
dvm get repos [flags]                # Alias
dvm get grs [flags]                  # Alias
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `-o, --output <format>` | — | string | table | Output format: `table`, `wide`, `json`, `yaml` |

The default table shows `NAME`, `URL`, `STATUS`, and `LAST_SYNCED`. The `wide` format adds `SLUG`, `REF`, and `AUTO_SYNC` columns.

**Examples:**

```bash
dvm get gitrepos
dvm get repos                        # Short form
dvm get gitrepos -o wide             # Show extra columns
dvm get gitrepos -o yaml
dvm get gitrepos -o json
```

### `dvm get gitrepo`

Get details for a specific git repository mirror.

```bash
dvm get gitrepo <name> [flags]
dvm get repo <name> [flags]          # Alias
dvm get gr <name> [flags]            # Alias
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `-o, --output <format>` | — | string | table | Output format: `table`, `json`, `yaml` |

**Examples:**

```bash
dvm get gitrepo my-repo
dvm get repo my-repo                 # Short form
dvm get gitrepo my-repo -o yaml
dvm get gitrepo my-repo -o json
```

### `dvm sync gitrepo`

Sync a single git repository mirror with its remote (fetch latest changes). If the local mirror does not exist yet, performs an initial clone.

```bash
dvm sync gitrepo <name>
dvm sync repo <name>                 # Alias
dvm sync gr <name>                   # Alias
```

**Examples:**

```bash
dvm sync gitrepo my-repo
dvm sync repo my-repo                # Short form
dvm sync gr my-repo                  # Short form
```

### `dvm sync gitrepos`

Sync all git repository mirrors with their remotes. Reports a count of successful and failed syncs.

```bash
dvm sync gitrepos
dvm sync repos                       # Alias
dvm sync grs                         # Alias
```

**Examples:**

```bash
dvm sync gitrepos
dvm sync repos                       # Short form
```

### `dvm describe gitrepo`

Show a rich status view for a single bare git mirror. Includes mirror health, disk usage, branch and tag counts, last sync time, credential status, and all apps and workspaces currently linked to the mirror.

```bash
dvm describe gitrepo <name> [flags]
dvm describe repo <name> [flags]     # Alias
dvm describe gr <name> [flags]       # Alias
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `-o, --output <format>` | — | string | table | Output format: `table`, `json`, `yaml` |

**Fields shown (table output):** Name, URL, Default Ref, Auth Type, Credential, Disk Usage, Branch Count, Tag Count, Mirror Status, Last Synced, Linked Apps, Linked Workspaces.

**Examples:**

```bash
# Show rich status for a git mirror
dvm describe gitrepo my-repo

# Output as YAML (useful for scripting)
dvm describe gitrepo my-repo -o yaml

# Using short alias
dvm describe gr my-repo
```

### `dvm get branches`

List all branches in a bare git mirror.

```bash
dvm get branches --repo <name> [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--repo <name>` | — | string | — | **Required.** Name of the git repo mirror |
| `-o, --output <format>` | — | string | table | Output format: `table`, `json`, `yaml` |

**Examples:**

```bash
# List branches in a mirror
dvm get branches --repo my-repo

# Output as YAML
dvm get branches --repo my-repo -o yaml
```

### `dvm get tags`

List all tags in a bare git mirror.

```bash
dvm get tags --repo <name> [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--repo <name>` | — | string | — | **Required.** Name of the git repo mirror |
| `-o, --output <format>` | — | string | table | Output format: `table`, `json`, `yaml` |

**Examples:**

```bash
# List tags in a mirror
dvm get tags --repo my-repo

# Output as JSON
dvm get tags --repo my-repo -o json
```

### `dvm create branch`

Create a new git branch in a workspace's repository checkout.

```bash
dvm create branch <name> [flags]
```

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--workspace <name>` | `-w` | string | active workspace | Workspace to create the branch in |
| `--app <name>` | `-a` | string | active app | App that owns the workspace |
| `--from <ref>` | — | string | `HEAD` | Base ref to branch from (branch, tag, or commit SHA) |

**Examples:**

```bash
# Create a branch in the active workspace
dvm create branch feature-auth

# Create a branch in a specific workspace
dvm create branch feature-auth --workspace dev

# Create a branch from a specific tag
dvm create branch hotfix-123 --from v1.2.0

# Create a branch in a named app's workspace
dvm create branch feature-x --workspace dev --app my-api
```

---

## Build Args (NEW in v0.55.0)

Build args cascade down the full `global → ecosystem → domain → app → workspace` hierarchy. The most specific level wins. Resolved args are injected as `--build-arg KEY=VALUE` at build time; values are not stored in image layers.

### `dvm set build-arg`

Set a build arg at any hierarchy level.

```bash
dvm set build-arg <KEY> <VALUE> [flags]
```

**Hierarchy flags (exactly one required):**

| Flag | Description |
|------|-------------|
| `--global` | Set at global level (lowest priority — inherited by everything) |
| `--ecosystem <name>` | Set at ecosystem level |
| `--domain <name>` | Set at domain level |
| `--app <name>` | Set at app level |
| `--workspace <name>` | Set at workspace level (highest priority) |

**Key validation:** Keys must be valid environment variable names. `DVM_`-prefixed keys and dangerous system variables (e.g., `PATH`, `HOME`) are rejected.

**Examples:**

```bash
# Set a global PyPI mirror (lowest priority)
dvm set build-arg PIP_INDEX_URL "https://pypi.corp.example/simple" --global

# Override at ecosystem level
dvm set build-arg PIP_INDEX_URL "https://pypi.eu.corp.example/simple" --ecosystem my-platform

# Set a build token only for a specific app
dvm set build-arg GITHUB_PAT "ghp_abc123" --app my-api

# Set at workspace level (highest priority — overrides all others)
dvm set build-arg NPM_TOKEN "npm_xyz789" --workspace dev
```

---

### `dvm get build-args`

List build args at the active or specified hierarchy level.

```bash
dvm get build-args [flags]
```

| Flag | Description |
|------|-------------|
| `--global` | Show global-level build args |
| `--ecosystem <name>` | Show build args for an ecosystem |
| `--domain <name>` | Show build args for a domain |
| `--app <name>` | Show build args for an app |
| `--workspace <name>` | Show build args for a workspace |
| `--effective` | Show the fully merged cascade with provenance |
| `--output <format>` | Output format: `table`, `yaml`, `json` |

**Build Args Cascade:**
Args inherit down the hierarchy unless overridden:
```
global → ecosystem → domain → app → workspace
```
The most specific level (workspace) wins. Use `--effective` to see the merged result with a provenance column showing which level each arg comes from.

**Examples:**

```bash
# List global build args
dvm get build-args --global

# Show the full effective cascade for a workspace (with provenance)
dvm get build-args --workspace dev --effective

# Machine-readable output
dvm get build-args --app my-api --output yaml
dvm get build-args --workspace dev --effective --output json
```

---

### `dvm delete build-arg`

Delete a build arg at a specific hierarchy level.

```bash
dvm delete build-arg <KEY> [flags]
```

**Hierarchy flags (exactly one required):**

| Flag | Description |
|------|-------------|
| `--global` | Delete from global level |
| `--ecosystem <name>` | Delete from ecosystem level |
| `--domain <name>` | Delete from domain level |
| `--app <name>` | Delete from app level |
| `--workspace <name>` | Delete from workspace level |

**Examples:**

```bash
# Remove a global build arg
dvm delete build-arg PIP_INDEX_URL --global

# Remove an override set at the app level
dvm delete build-arg GITHUB_PAT --app my-api

# Remove a workspace-level override
dvm delete build-arg NPM_TOKEN --workspace dev
```

---

## CA Certs (NEW in v0.56.0)

CA certificates cascade down the full `global → ecosystem → domain → app → workspace` hierarchy. The most specific level wins by cert name. Resolved certs are fetched from MaestroVault at build time and injected into the container image.

### `dvm set ca-cert`

Set a CA cert at any hierarchy level.

```bash
dvm set ca-cert <NAME> [flags]
```

**Vault source flags:**

| Flag | Description |
|------|-------------|
| `--vault-secret <name>` | **Required.** Name of the MaestroVault secret containing the certificate |
| `--vault-env <name>` | Optional vault environment override |
| `--vault-field <name>` | Field within the secret (default: `cert`) |

**Hierarchy flags (exactly one required):**

| Flag | Description |
|------|-------------|
| `--global` | Set at global level (lowest priority — inherited by everything) |
| `--ecosystem <name>` | Set at ecosystem level |
| `--domain <name>` | Set at domain level |
| `--app <name>` | Set at app level |
| `--workspace <name>` | Set at workspace level (highest priority) |

**Additional flags:**

| Flag | Description |
|------|-------------|
| `--dry-run` | Preview the operation without applying any changes |

**Name validation:** Names must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`; maximum 64 characters.

**Examples:**

```bash
# Set a global corporate CA cert
dvm set ca-cert corp-root-ca --vault-secret corp-root-ca-pem --global

# Override at ecosystem level with a different secret
dvm set ca-cert corp-root-ca --vault-secret corp-root-ca-eu-pem --ecosystem my-platform

# Set a cert only for a specific app
dvm set ca-cert internal-ca --vault-secret internal-ca-pem --app my-api

# Preview without applying
dvm set ca-cert corp-root-ca --vault-secret corp-root-ca-pem --global --dry-run
```

---

### `dvm get ca-certs`

List CA certs at the active or specified hierarchy level.

```bash
dvm get ca-certs [flags]
```

| Flag | Description |
|------|-------------|
| `--global` | Show global-level CA certs |
| `--ecosystem <name>` | Show CA certs for an ecosystem |
| `--domain <name>` | Show CA certs for a domain |
| `--app <name>` | Show CA certs for an app |
| `--workspace <name>` | Show CA certs for a workspace |
| `--effective` | Show the fully merged cascade with SOURCE provenance (requires `--workspace`) |
| `-o, --output <format>` | Output format: `table`, `yaml`, `json` |

**CA Certs Cascade:**
Certs inherit down the hierarchy unless overridden by cert name:
```
global → ecosystem → domain → app → workspace
```
The most specific level (workspace) wins. Use `--effective` to see the merged result with a SOURCE column showing which level each cert comes from.

**Examples:**

```bash
# List global CA certs
dvm get ca-certs --global

# Show the full effective cascade for a workspace (with provenance)
dvm get ca-certs --workspace dev --effective

# Machine-readable output
dvm get ca-certs --app my-api -o yaml
dvm get ca-certs --workspace dev --effective -o json
```

---

### `dvm delete ca-cert`

Delete a CA cert at a specific hierarchy level.

```bash
dvm delete ca-cert <NAME> [flags]
```

**Hierarchy flags (exactly one required):**

| Flag | Description |
|------|-------------|
| `--global` | Delete from global level |
| `--ecosystem <name>` | Delete from ecosystem level |
| `--domain <name>` | Delete from domain level |
| `--app <name>` | Delete from app level |
| `--workspace <name>` | Delete from workspace level |

**Additional flags:**

| Flag | Description |
|------|-------------|
| `-f, --force` | Skip confirmation prompt |

Deleting a cert that does not exist at the specified level is a no-op (exits cleanly with no error).

**Examples:**

```bash
# Remove a global CA cert (with confirmation prompt)
dvm delete ca-cert corp-root-ca --global

# Remove an override set at the app level, skipping prompt
dvm delete ca-cert internal-ca --app my-api --force

# Remove a workspace-level override
dvm delete ca-cert corp-root-ca --workspace dev
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

## Rollout

Manage resource rollouts following kubectl patterns. Supports `restart`, `status`, `history`, and `undo` operations for registries. Each restart operation creates a new revision entry in the rollout history for traceability.

### `dvm rollout restart`

Restart a resource (stops then starts it).

```bash
dvm rollout restart <resource> <name>
```

#### `dvm rollout restart registry`

Restart a registry by stopping it then starting it again. Creates a new revision entry in the rollout history on success or failure.

```bash
dvm rollout restart registry <name>
```

**Aliases:** `registry` → `reg`

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<name>` | Registry name (required) |

**Examples:**

```bash
# Restart a registry
dvm rollout restart registry local-cache

# Using the alias
dvm rollout restart reg local-cache
```

---

### `dvm rollout status`

Show the current rollout status of a resource.

```bash
dvm rollout status <resource> <name> [flags]
```

#### `dvm rollout status registry`

Show the current rollout status of a registry. Displays the registry name, enabled state, runtime status, endpoint, configuration, and latest revision details.

```bash
dvm rollout status registry <name> [flags]
```

**Aliases:** `registry` → `reg`

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<name>` | Registry name (required) |

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output <format>` | `-o` | string | `""` | Output format: `table`, `json`, `yaml` |

**Fields shown (table output):** Name, Enabled, Running, Endpoint, Lifecycle, Port, Storage, Latest Revision, Latest Action, Latest Status, Last Updated.

**Examples:**

```bash
# Show registry rollout status (human-readable)
dvm rollout status registry local-cache

# Output as JSON
dvm rollout status registry local-cache -o json

# Output as YAML
dvm rollout status registry local-cache -o yaml

# Using the alias
dvm rollout status reg local-cache
```

---

### `dvm rollout history`

Show the rollout history for a resource.

```bash
dvm rollout history <resource> <name> [flags]
```

#### `dvm rollout history registry`

Show all past revisions for a registry with their actions, statuses, and timestamps.

```bash
dvm rollout history registry <name> [flags]
```

**Aliases:** `registry` → `reg`

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<name>` | Registry name (required) |

**Flags:**

| Flag | Short | Type | Default | Description |
|------|-------|------|---------|-------------|
| `--output <format>` | `-o` | string | `""` | Output format: `table`, `json`, `yaml` |

**Table columns:** `REVISION`, `ACTION`, `STATUS`, `CREATED`, `COMPLETED`

**Examples:**

```bash
# Show registry rollout history (human-readable table)
dvm rollout history registry local-cache

# Output as JSON
dvm rollout history registry local-cache -o json

# Output as YAML
dvm rollout history registry local-cache -o yaml

# Using the alias
dvm rollout history reg local-cache
```

---

### `dvm rollout undo`

Rollback a resource to its previous version.

```bash
dvm rollout undo <resource> <name>
```

#### `dvm rollout undo registry`

Rollback a registry to its previous successful revision. Restores the configuration from the previous revision and restarts the registry with that configuration.

> **Note:** `dvm rollout undo registry` is not yet implemented. The command is defined and registered but returns an error when invoked.

```bash
dvm rollout undo registry <name>
```

**Aliases:** `registry` → `reg`

**Arguments:**

| Argument | Description |
|----------|-------------|
| `<name>` | Registry name (required) |

**Examples:**

```bash
# Undo the last rollout for a registry
dvm rollout undo registry local-cache

# Using the alias
dvm rollout undo reg local-cache
```

---

## Nvim Management

### `dvm nvim init`

Initialize local Neovim configuration from a template.

```bash
dvm nvim init <template> [flags]
```

**Templates:**

| Template | Description |
|----------|-------------|
| `kickstart` | kickstart.nvim — minimal, well-documented starter |
| `lazyvim` | LazyVim — feature-rich, batteries-included |
| `astronvim` | AstroNvim — aesthetically pleasing, fully featured |
| `minimal` | Minimal config created by DevOpsMaestro |
| `custom` | Clone from a custom Git URL (requires `--git-url`) |
| `<url>` | Any HTTPS or `github:user/repo` URL cloned directly |

By default, creates a minimal config. Pass `--git-clone` to clone from the template's upstream repository.

**Flags:**

| Flag | Description |
|------|-------------|
| `--config-path <path>` | Custom config path (default: `~/.config/nvim`) |
| `--git-clone` | Clone template from upstream Git repository |
| `--git-url <url>` | Custom Git URL (for `custom` template) |
| `--subdir <path>` | Subdirectory within repo to use as config root |
| `--overwrite` | Overwrite existing config |

**Examples:**

```bash
# Create a minimal config locally
dvm nvim init minimal

# Clone kickstart.nvim from upstream
dvm nvim init kickstart --git-clone

# Clone LazyVim starter
dvm nvim init lazyvim --git-clone

# Clone from a GitHub URL directly
dvm nvim init https://github.com/yourusername/nvim-config.git

# Clone using short GitHub format
dvm nvim init github:yourusername/nvim-config

# Clone a subdirectory from a repo
dvm nvim init https://github.com/user/repo.git --subdir templates/starter

# Overwrite an existing config
dvm nvim init kickstart --git-clone --overwrite
```

### `dvm nvim status`

Show current local Neovim configuration status.

```bash
dvm nvim status
```

Displays:
- Config location and whether config exists
- Template used
- Last sync time and which workspace was synced
- Whether there are local changes since the last sync

**Examples:**

```bash
dvm nvim status
```

### `dvm nvim sync`

Synchronize local Neovim config with a workspace container (pull from workspace).

```bash
dvm nvim sync <workspace> [flags]
```

Pulls the Neovim config FROM the workspace TO your local machine. Use `dvm nvim push` to push local changes to a workspace.

**Flags:**

| Flag | Description |
|------|-------------|
| `--remote-wins` | Remote changes win in conflicts |

**Examples:**

```bash
# Pull config from workspace
dvm nvim sync my-workspace

# Pull with automatic conflict resolution (remote wins)
dvm nvim sync my-workspace --remote-wins
```

### `dvm nvim push`

Push local Neovim configuration to a workspace container.

```bash
dvm nvim push <workspace> [flags]
```

Copies the local config TO the workspace container, overwriting the workspace's existing config.

**Flags:**

| Flag | Description |
|------|-------------|
| `--restart` | Restart Neovim in workspace after push |

**Examples:**

```bash
# Push local config to workspace
dvm nvim push my-workspace

# Push and restart Neovim in workspace
dvm nvim push my-workspace --restart
```

---

## Nvim Resources (Extended)

### `dvm edit nvim plugin`

Edit a nvim plugin definition in your default editor (`$EDITOR`). After saving and closing the editor, changes are automatically applied to the database.

```bash
dvm edit nvim plugin <name>
```

Falls back to `vi` if `$EDITOR` is not set.

**Examples:**

```bash
dvm edit nvim plugin telescope
EDITOR=vim dvm edit nvim plugin mason
```

### `dvm edit nvim theme`

Edit a nvim theme definition (redirects to `nvp` CLI).

```bash
dvm edit nvim theme <name>
```

> **Note:** Theme editing is currently available via the standalone `nvp` CLI. This command displays instructions for using `nvp theme get`, editing the YAML, and re-applying with `nvp theme apply -f`.

**Examples:**

```bash
dvm edit nvim theme tokyonight-night
```

### `dvm apply nvim plugin`

Apply a nvim plugin from a YAML file (IaC). See [`dvm apply`](#dvm-apply) for full source type support.

```bash
dvm apply -f plugin.yaml
```

**Resource kind:** `NvimPlugin`

### `dvm apply nvim theme`

Apply a nvim theme from a YAML file (IaC). See [`dvm apply`](#dvm-apply) for full source type support.

```bash
dvm apply -f theme.yaml
```

**Resource kind:** `NvimTheme`

### `dvm set nvim plugin`

Add nvim plugins to a workspace configuration or set global default plugins.

```bash
dvm set nvim plugin [names...] [flags]
```

Plugins must exist in the global library (`~/.nvp/plugins/`). Use `dvm get nvim plugins` to see available plugins.

**Flags:**

| Flag | Description |
|------|-------------|
| `-w, --workspace <name>` | Workspace to configure (required unless `--global`) |
| `-a, --app <name>` | App for workspace (defaults to active) |
| `--all` | Add all plugins from global library |
| `--clear` | Remove all plugins from workspace or clear global defaults |
| `--global` | Set global default plugins (mutually exclusive with `--workspace`) |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table`, `colored` |
| `--dry-run` | Preview changes without applying |

**Examples:**

```bash
# Add specific plugins to a workspace
dvm set nvim plugin -w dev treesitter lspconfig telescope

# Add all global plugins to workspace
dvm set nvim plugin -w dev --all

# Remove all plugins from workspace
dvm set nvim plugin -w dev --clear

# Specify app explicitly
dvm set nvim plugin -a myapp -w dev treesitter

# Set global default plugins
dvm set nvim plugin lazygit telescope --global

# Clear global default plugins
dvm set nvim plugin --clear --global
```

### `dvm get nvim packages`

List all nvim packages stored in the database.

```bash
dvm get nvim packages [flags]
```

**Aliases:** `pkg`, `pkgs`

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get nvim packages
dvm get nvim packages -o yaml
dvm get nvim packages -o json
```

### `dvm get nvim package`

Get a specific nvim package by name.

```bash
dvm get nvim package <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get nvim package core
dvm get nvim package core -o yaml
dvm get nvim package minimal -o json
```

### `dvm get nvim defaults`

Show current nvim default configuration values.

```bash
dvm get nvim defaults [flags]
```

Displays nvim-related defaults: `nvim-package`, `theme`, and `plugins`.

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get nvim defaults
dvm get nvim defaults -o yaml
dvm get nvim defaults -o json
```

### `dvm get nvim-package`

Show the resolved nvim plugin package for the current workspace context. Walks the hierarchy `workspace → app → domain → ecosystem → global default` and returns the first match.

```bash
dvm get nvim-package [flags]
```

Requires an active workspace context (set with `dvm use workspace <name>`).

**Flags:**

| Flag | Description |
|------|-------------|
| `--show-cascade` | Show full hierarchy walk with resolution path |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get nvim-package
dvm get nvim-package --show-cascade
dvm get nvim-package -o yaml
```

### `dvm set nvim-package`

Set nvim plugin package at any hierarchy level. Packages cascade down unless overridden: `global → Ecosystem → Domain → App → Workspace`.

```bash
dvm set nvim-package <name> [flags]
```

Use `none` as the name to clear the override at the specified level and inherit from the parent.

**Hierarchy flags (exactly one required):**

| Flag | Description |
|------|-------------|
| `--global` | Set as global default (lowest priority) |
| `--ecosystem <name>` | Set at ecosystem level |
| `--domain <name>` | Set at domain level |
| `--app <name>` | Set at app level |
| `--workspace <name>` | Set at workspace level (highest priority) |

**Other flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table`, `colored` |
| `--dry-run` | Preview changes without applying |
| `--show-cascade` | Show package cascade effect after setting |

**Examples:**

```bash
# Set at workspace level (highest priority)
dvm set nvim-package full-stack --workspace dev

# Set at app level (applies to all workspaces unless overridden)
dvm set nvim-package minimal --app my-api

# Clear workspace override — inherits from app
dvm set nvim-package none --workspace dev

# Set for an entire domain
dvm set nvim-package standard --domain auth

# Set the global default (inherited by everything)
dvm set nvim-package full-stack --global

# Clear the global default
dvm set nvim-package none --global
```

---

## Terminal Management

### `dvm get terminal packages`

List all terminal packages stored in the database.

```bash
dvm get terminal packages [flags]
```

**Aliases:** `pkg`, `pkgs`

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get terminal packages
dvm get terminal packages -o yaml
dvm get terminal packages -o json
```

### `dvm get terminal package`

Get a specific terminal package by name.

```bash
dvm get terminal package <name> [flags]
```

Displays: name, category, description, extends, plugins, prompts, profiles, tags.

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get terminal package dev-essentials
dvm get terminal package dev-essentials -o yaml
dvm get terminal package poweruser -o json
```

### `dvm get terminal defaults`

Show current terminal default configuration values.

```bash
dvm get terminal defaults [flags]
```

Displays terminal-related defaults: `terminal-package`.

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get terminal defaults
dvm get terminal defaults -o yaml
dvm get terminal defaults -o json
```

### `dvm set terminal prompt`

Set the terminal prompt configuration for a workspace.

```bash
dvm set terminal prompt [name] [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-w, --workspace <name>` | Workspace to configure (**required**) |
| `-a, --app <name>` | App for workspace (defaults to active) |
| `--clear` | Remove prompt from workspace |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table`, `colored` |
| `--dry-run` | Preview changes without applying |

**Examples:**

```bash
# Set the terminal prompt for a workspace
dvm set terminal prompt -w dev starship

# Set with a specific variant
dvm set terminal prompt -w dev starship-minimal

# Preview without applying
dvm set terminal prompt -w dev starship --dry-run

# Remove prompt from workspace
dvm set terminal prompt -w dev --clear
```

### `dvm set terminal plugin`

Add terminal plugins to a workspace configuration.

```bash
dvm set terminal plugin [names...] [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-w, --workspace <name>` | Workspace to configure (**required**) |
| `-a, --app <name>` | App for workspace (defaults to active) |
| `--all` | Add all plugins from library |
| `--clear` | Remove all plugins from workspace |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table`, `colored` |
| `--dry-run` | Preview changes without applying |

**Examples:**

```bash
# Add a single plugin
dvm set terminal plugin -w dev zsh-autosuggestions

# Add multiple plugins
dvm set terminal plugin -w dev zsh-autosuggestions zsh-syntax-highlighting

# Add all available plugins
dvm set terminal plugin -w dev --all

# Remove all plugins from workspace
dvm set terminal plugin -w dev --clear
```

### `dvm set terminal package`

Set a terminal package (bundle of plugins, prompts, and configurations) for a workspace.

```bash
dvm set terminal package [name] [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-w, --workspace <name>` | Workspace to configure (**required**) |
| `-a, --app <name>` | App for workspace (defaults to active) |
| `--clear` | Remove package from workspace |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table`, `colored` |
| `--dry-run` | Preview changes without applying |

**Examples:**

```bash
# Set terminal package for a workspace
dvm set terminal package -w dev poweruser

# Set a minimal package
dvm set terminal package -w dev minimal

# Remove package from workspace
dvm set terminal package -w dev --clear
```

### `dvm use terminal package`

Set the global default terminal package for new workspaces. New workspaces that don't specify a package will use this default.

```bash
dvm use terminal package <name>
```

Use `none` to clear the default.

**Examples:**

```bash
# Set default terminal package
dvm use terminal package developer-essentials

# Set poweruser as default
dvm use terminal package poweruser

# Clear default terminal package
dvm use terminal package none
```

### `dvm get terminal-package`

Show the resolved terminal package for the current workspace context. Walks the hierarchy `workspace → app → domain → ecosystem → global default` and returns the first match.

```bash
dvm get terminal-package [flags]
```

Requires an active workspace context (set with `dvm use workspace <name>`).

**Flags:**

| Flag | Description |
|------|-------------|
| `--show-cascade` | Show full hierarchy walk with resolution path |
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table` |

**Examples:**

```bash
dvm get terminal-package
dvm get terminal-package --show-cascade
dvm get terminal-package -o yaml
```

### `dvm set terminal-package`

Set terminal package at any hierarchy level. Packages cascade down unless overridden: `global → Ecosystem → Domain → App → Workspace`.

```bash
dvm set terminal-package <name> [flags]
```

Use `none` as the name to clear the override at the specified level and inherit from the parent.

**Hierarchy flags (exactly one required):**

| Flag | Description |
|------|-------------|
| `--global` | Set as global default (lowest priority) |
| `--ecosystem <name>` | Set at ecosystem level |
| `--domain <name>` | Set at domain level |
| `--app <name>` | Set at app level |
| `--workspace <name>` | Set at workspace level (highest priority) |

**Other flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml`, `plain`, `table`, `colored` |
| `--dry-run` | Preview changes without applying |
| `--show-cascade` | Show package cascade effect after setting |

**Examples:**

```bash
# Set at workspace level (highest priority)
dvm set terminal-package poweruser --workspace dev

# Set at app level (applies to all workspaces unless overridden)
dvm set terminal-package minimal --app my-api

# Clear workspace override — inherits from app
dvm set terminal-package none --workspace dev

# Set for an entire domain
dvm set terminal-package standard --domain auth

# Set the global default (inherited by everything)
dvm set terminal-package poweruser --global

# Clear the global default
dvm set terminal-package none --global
```

---

## Context Switching

### `dvm use ecosystem`

Set the active ecosystem. Clears domain, app, and workspace context downstream.

```bash
dvm use ecosystem <name> [flags]
dvm use eco <name> [flags]        # Alias
```

Use `none` as the name to clear the ecosystem context (also clears domain, app, and workspace).

**Flags:**

| Flag | Description |
|------|-------------|
| `--export` | Print `export DVM_ECOSYSTEM=<name>` for shell eval instead of updating DB |
| `--dry-run` | Preview the context switch without applying |

**Examples:**

```bash
# Set active ecosystem
dvm use ecosystem my-platform

# Short form
dvm use eco my-platform

# Switch to another ecosystem
dvm use ecosystem staging

# Clear ecosystem context (also clears domain, app, workspace)
dvm use ecosystem none

# Print export statement for eval in current shell tab
eval $(dvm use ecosystem my-platform --export)
```

### `dvm use domain`

Set the active domain. Requires an active ecosystem. Clears app and workspace context downstream.

```bash
dvm use domain <name> [flags]
dvm use dom <name> [flags]        # Alias
```

Use `none` as the name to clear the domain context (also clears app and workspace).

**Flags:**

| Flag | Description |
|------|-------------|
| `--export` | Print `export DVM_DOMAIN=<name>` for shell eval instead of updating DB |
| `--dry-run` | Preview the context switch without applying |

**Examples:**

```bash
# Set active domain (requires active ecosystem)
dvm use domain backend

# Short form
dvm use dom backend

# Switch to another domain
dvm use domain frontend

# Clear domain context (also clears app and workspace)
dvm use domain none

# Print export statement for eval in current shell tab
eval $(dvm use domain backend --export)
```

### `dvm use app`

Set the active app. Searches globally across all domains. Clears workspace context downstream.

```bash
dvm use app <name> [flags]
dvm use a <name> [flags]            # Short alias
dvm use application <name> [flags]  # Full alias
```

Use `none` as the name to clear the app context (also clears workspace).

**Flags:**

| Flag | Description |
|------|-------------|
| `--export` | Print `export DVM_APP=<name>` for shell eval instead of updating DB |
| `--dry-run` | Preview the context switch without applying |

**Examples:**

```bash
# Set active app
dvm use app my-api

# Short form
dvm use a my-api

# Switch to another app
dvm use app web-frontend

# Clear app context (also clears workspace)
dvm use app none

# Print export statement for eval in current shell tab
eval $(dvm use app my-api --export)
```

### `dvm use workspace`

Set the active workspace. Requires an active app.

```bash
dvm use workspace <name> [flags]
dvm use ws <name> [flags]       # Alias
```

Use `none` as the name to clear only the workspace context (keeps app active).

**Flags:**

| Flag | Description |
|------|-------------|
| `--export` | Print `export DVM_WORKSPACE=<name>` for shell eval instead of updating DB |
| `--dry-run` | Preview the context switch without applying |

**Examples:**

```bash
# Set active workspace (requires active app)
dvm use workspace dev

# Short form
dvm use ws dev

# Switch to another workspace
dvm use workspace feature-x

# Clear workspace only (app remains active)
dvm use workspace none

# Print export statement for eval in current shell tab
eval $(dvm use workspace dev --export)
```

### `dvm use -`

Toggle back to the previous context (like `cd -` in a shell).

```bash
dvm use -
```

**Examples:**

```bash
dvm use -
```

---

## Singular Get Commands

### `dvm get ecosystem`

Get details for a specific ecosystem.

```bash
dvm get ecosystem <name> [flags]
dvm get eco <name> [flags]       # Alias
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--show-theme` | Show theme resolution information |
| `-o, --output <format>` | Output format: `json`, `yaml` |

**Examples:**

```bash
# Show ecosystem details
dvm get ecosystem my-platform

# Show with theme info
dvm get ecosystem my-platform --show-theme

# Output as JSON (includes list of domain names)
dvm get ecosystem my-platform -o json
```

### `dvm get domain`

Get details for a specific domain.

```bash
dvm get domain <name> [flags]
dvm get dom <name> [flags]       # Alias
```

Looks up the domain in the active ecosystem. Use `--ecosystem` to target a specific ecosystem.

**Flags:**

| Flag | Description |
|------|-------------|
| `-e, --ecosystem <name>` | Look up domain in this ecosystem instead of the active one |
| `--show-theme` | Show theme resolution information |
| `-o, --output <format>` | Output format: `json`, `yaml` |

**Examples:**

```bash
# Show domain details (uses active ecosystem)
dvm get domain backend

# Target a specific ecosystem
dvm get domain backend --ecosystem my-platform

# Output as JSON (includes list of app names)
dvm get domain backend -o json
```

### `dvm get app`

Get details for a specific app.

```bash
dvm get app <name> [flags]
dvm get a <name> [flags]            # Short alias
dvm get application <name> [flags]  # Full alias
```

Looks up the app in the active domain. Use `--domain` to target a specific domain.

**Flags:**

| Flag | Description |
|------|-------------|
| `-d, --domain <name>` | Look up app in this domain instead of the active one |
| `--show-theme` | Show theme resolution information |
| `-o, --output <format>` | Output format: `json`, `yaml` |

**Examples:**

```bash
# Show app details (uses active domain)
dvm get app my-api

# Target a specific domain
dvm get app my-api --domain backend

# Output as JSON (includes workspace names and git repo)
dvm get app my-api -o json

# Show app with theme resolution
dvm get app my-api --show-theme
```

### `dvm get workspace`

Get details for a specific workspace.

```bash
dvm get workspace <name> [flags]
dvm get ws <name> [flags]       # Alias
```

If the workspace name matches multiple apps, dvm prints a disambiguation table listing all matches. Use hierarchy flags to narrow the scope.

**Flags:**

| Flag | Description |
|------|-------------|
| `-e, --ecosystem <name>` | Filter by ecosystem |
| `-d, --domain <name>` | Filter by domain |
| `-a, --app <name>` | Filter by app |
| `-w, --workspace <name>` | Alternative to positional argument |
| `--show-theme` | Show theme resolution information |
| `-o, --output <format>` | Output format: `json`, `yaml` |

**Examples:**

```bash
# Show workspace details (uses active app)
dvm get workspace dev

# Short form
dvm get ws dev

# Target a specific app
dvm get workspace dev --app my-api

# Output as YAML
dvm get workspace dev -o yaml

# Show with theme resolution
dvm get workspace dev --show-theme
```

---

## Defaults

### `dvm get defaults`

Display default configuration values for containers, shells, Neovim, and themes. Merges hardcoded defaults with any user-set values from the database (`nvim-package`, `terminal-package`, `theme`).

```bash
dvm get defaults [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `json`, `yaml` |

**Examples:**

```bash
# Show all defaults (human-readable)
dvm get defaults

# Output as JSON
dvm get defaults -o json

# Output as YAML
dvm get defaults -o yaml
```

---

## Library

Browse and import embedded plugin, theme, prompt, and package libraries without needing a database connection.

### `dvm library get`

List library resources.

```bash
dvm library get <type> [flags]
dvm lib get <type> [flags]     # Alias
dvm lib ls <type> [flags]      # Alias (ls = get)
```

**Resource types:**

| Type | Aliases | Description |
|------|---------|-------------|
| `plugins` | `np` | Neovim plugins |
| `themes` | `nt` | Neovim themes |
| `nvim packages` | `nvim-packages` | Neovim plugin bundles |
| `terminal prompts` | `terminal-prompts`, `tp` | Terminal prompt configs |
| `terminal plugins` | `terminal-plugins`, `tpl` | Shell plugins |
| `terminal packages` | `terminal-packages` | Terminal plugin/prompt bundles |
| `terminal emulators` | `terminal-emulators` | Terminal emulator configs |

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `table` (default), `json`, `yaml` |

**Examples:**

```bash
# List nvim plugins
dvm library get plugins

# Short form
dvm lib ls np

# List nvim themes
dvm library get themes

# List terminal prompts
dvm library get terminal prompts

# List terminal packages as YAML
dvm library get terminal packages -o yaml

# List terminal emulators
dvm library get terminal-emulators
```

### `dvm library describe`

Show details of a specific library resource.

```bash
dvm library describe <type> <name> [flags]
dvm lib describe <type> <name> [flags]   # Alias
```

**Resource types (singular form):**

| Type | Description |
|------|-------------|
| `plugin` | Neovim plugin |
| `theme` | Neovim theme |
| `nvim-package` | Neovim plugin bundle |
| `terminal prompt` | Terminal prompt config |
| `terminal plugin` | Shell plugin |
| `terminal-package` | Terminal bundle |
| `terminal-emulator` | Terminal emulator config |

**Flags:**

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: `table` (default), `json`, `yaml` |

**Examples:**

```bash
# Show nvim plugin details
dvm library describe plugin telescope

# Show theme details
dvm library describe theme coolnight-ocean

# Show nvim package details
dvm library describe nvim-package core

# Show terminal prompt details
dvm library describe terminal prompt starship-default

# Show terminal plugin details
dvm library describe terminal plugin zsh-autosuggestions

# Show terminal package details
dvm library describe terminal-package developer-essentials

# Show terminal emulator details
dvm library describe terminal-emulator wezterm

# Output as JSON
dvm library describe plugin telescope -o json
```

### `dvm library import`

Import embedded library resources into the local database, making them available for workspace configuration.

```bash
dvm library import <type...> [flags]
dvm lib import <type...> [flags]   # Alias
```

At least one resource type or `--all` is required.

**Resource types:**

| Type | Description |
|------|-------------|
| `nvim-plugins` | Import Neovim plugins |
| `nvim-themes` | Import Neovim themes |
| `nvim-packages` | Import Neovim plugin bundles |
| `terminal-prompts` | Import terminal prompt configs |
| `terminal-plugins` | Import shell plugins |
| `terminal-packages` | Import terminal bundles |
| `terminal-emulators` | Import terminal emulator configs |

**Flags:**

| Flag | Description |
|------|-------------|
| `--all` | Import all 7 resource types at once |
| `-o, --output <format>` | Output format: `table` (default), `json`, `yaml` |

**Examples:**

```bash
# Import all library resources
dvm library import --all

# Import only nvim plugins
dvm library import nvim-plugins

# Import nvim plugins and themes together
dvm library import nvim-plugins nvim-themes

# Import all terminal resources
dvm library import terminal-prompts terminal-plugins terminal-packages terminal-emulators

# Short form
dvm lib import --all
```

---

## Administration

### `dvm admin migrate`

Apply database migrations to ensure the schema is up-to-date. Run this after upgrading dvm to a new version.

```bash
dvm admin migrate
```

No flags beyond global flags.

**Examples:**

```bash
dvm admin migrate
```

---

## Cache

### `dvm cache`

Root command for cache management. Shows help when called with no subcommand.

```bash
dvm cache
```

### `dvm cache clear`

Clear build caches. Supports BuildKit, npm, pip, and build staging caches.

```bash
dvm cache clear [flags]
```

If no specific cache type flag is given, all caches are cleared (equivalent to `--all`).

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--all` | | Clear all caches |
| `--buildkit` | | Clear BuildKit build cache |
| `--npm` | | Clear npm cache mount |
| `--pip` | | Clear pip cache mount |
| `--staging` | | Clear build staging directory |
| `--force` | `-f` | Skip confirmation prompt |
| `--dry-run` | | Preview what would be cleared without making changes |

**Examples:**

```bash
# Clear all caches (no flags = same as --all)
dvm cache clear

# Clear only BuildKit cache
dvm cache clear --buildkit

# Clear npm and pip caches
dvm cache clear --npm --pip

# Preview what would be cleared
dvm cache clear --dry-run

# Clear all without confirmation
dvm cache clear --all --force
```

---

## Set Credential

### `dvm set credential`

Update properties on an existing credential. Currently supports setting or clearing expiration for rotation reminders.

```bash
dvm set credential <name> [flags]
```

**Flags:**

| Flag | Description |
|------|-------------|
| `--expires <duration>` | Set expiration duration (e.g., `90d`, `365d`, `24h`, `8760h`). Use `0` to clear expiration. |
| `--dry-run` | Preview changes without applying |

**Scope flags (exactly one required):**

| Flag | Short | Description |
|------|-------|-------------|
| `--ecosystem <name>` | `-e` | Target credential in this ecosystem |
| `--domain <name>` | `-d` | Target credential in this domain |
| `--app <name>` | `-a` | Target credential in this app |
| `--workspace <name>` | `-w` | Target credential in this workspace |

**Examples:**

```bash
# Set expiration to 90 days from now
dvm set credential github-token --expires 90d --app my-api

# Set expiration to 1 year (using hours)
dvm set credential api-key --expires 8760h --ecosystem prod

# Set expiration scoped to a domain
dvm set credential db-pass --expires 365d --domain backend

# Clear expiration
dvm set credential deploy-key --expires 0 --app my-api

# Preview change without applying
dvm set credential github-token --expires 90d --app my-api --dry-run
```

---

## Template Generation

### `dvm generate template`

Output an annotated, copy-paste-ready YAML template for any resource kind to stdout. Templates are embedded in the binary and include all fields with inline comments, required/optional annotations, valid value ranges, and a documentation link in the header.

```bash
dvm generate template <kind> [flags]
```

**Supported kinds:**

| Kind | Description |
|------|-------------|
| `ecosystem` | Top-level platform grouping |
| `domain` | Bounded context within an ecosystem |
| `app` | Application/codebase within a domain |
| `workspace` | Development environment for an app |
| `credential` | Secret reference (MaestroVault or env) |
| `registry` | Local package registry (OCI, Python, Go, npm, HTTP) |
| `nvim-plugin` | Neovim plugin configuration |
| `terminal-prompt` | Shell prompt configuration |
| `source` | External plugin source definition |
| `mirror` | Git repository mirror |
| `build-arg` | Build argument at any hierarchy level |
| `ca-cert` | CA certificate for corporate network builds |
| `color` | Color palette or theme definition |
| `env` | Environment variable injection |
| `infra` | Infrastructure resource definition |

Kind names accept both kebab-case (e.g., `nvim-plugin`) and PascalCase (e.g., `NvimPlugin`). Shell completion is built in for kind names.

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--output <format>` | `-o` | Output format: `yaml` (default), `json` |
| `--all` | `-A` | Output all 15 kinds as a multi-document YAML stream |

**Examples:**

```bash
# Generate a workspace template and save to a file
dvm generate template workspace > my-workspace.yaml

# Generate an ecosystem template
dvm generate template ecosystem > my-ecosystem.yaml

# Use PascalCase kind name
dvm generate template Workspace > my-workspace.yaml

# Generate in JSON format
dvm generate template workspace --output json > my-workspace.json

# Generate ALL resource kinds as a multi-document YAML
dvm generate template --all > all-resources.yaml

# Preview a template and pipe directly to an editor
dvm generate template app | vim -
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
| `registry` | `reg` |
| `registries` | `reg`, `regs` |
| `gitrepo` | `repo`, `gr` |
| `gitrepos` | `repos`, `grs` |
| `rollout restart registry` | `rollout restart reg` |
| `rollout status registry` | `rollout status reg` |
| `rollout history registry` | `rollout history reg` |
| `rollout undo registry` | `rollout undo reg` |
| `library` | `lib` |
| `library get` | `lib ls`, `lib get` |
| `library get plugins` | `lib ls np` |
| `library get themes` | `lib ls nt` |
| `library get terminal prompts` | `lib ls tp` |
| `library get terminal plugins` | `lib ls tpl` |

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
