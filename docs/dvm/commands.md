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
- Builds the image using the detected container platform and tags it as `dvm-<workspace>-<app>:latest`

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
