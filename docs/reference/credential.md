# Credential YAML Reference

**Kind:** `Credential`  
**APIVersion:** `devopsmaestro.io/v1`

A Credential stores a reference to a secret — not the secret itself. Credentials point to values stored in the macOS Keychain or in environment variables. They are scoped to exactly one resource (ecosystem, domain, app, or workspace) and resolve automatically when that scope is active.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-token
  app: api-service           # Scoped to this app
spec:
  source: keychain
  service: com.github.token
  description: "GitHub PAT for pulling private packages"
```

### Dual-Field Keychain Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: docker-registry
  ecosystem: prod-platform   # Scoped to this ecosystem
spec:
  source: keychain
  service: com.registry.docker
  description: "Docker Hub credentials"
  usernameVar: DOCKER_USERNAME   # Injected as this env var (from keychain account field)
  passwordVar: DOCKER_PASSWORD   # Injected as this env var (from keychain password field)
```

### Environment Variable Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: api-key
  domain: backend            # Scoped to this domain
spec:
  source: env
  envVar: MY_API_KEY
  description: "External API key read from the host environment"
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `Credential` |
| `metadata.name` | string | ✅ | Unique name for the credential within its scope |
| `metadata.ecosystem` | string | ❌ | Scope to an ecosystem (exactly one scope required) |
| `metadata.domain` | string | ❌ | Scope to a domain (exactly one scope required) |
| `metadata.app` | string | ❌ | Scope to an app (exactly one scope required) |
| `metadata.workspace` | string | ❌ | Scope to a workspace (exactly one scope required) |
| `spec.source` | string | ✅ | Where the secret lives: `keychain` or `env` |
| `spec.service` | string | ✅ (keychain) | macOS Keychain service name — required when `source: keychain` |
| `spec.envVar` | string | ✅ (env) | Environment variable name — required when `source: env` |
| `spec.description` | string | ❌ | Human-readable description of the credential |
| `spec.usernameVar` | string | ❌ | Env var name to receive the keychain account field (keychain only) |
| `spec.passwordVar` | string | ❌ | Env var name to receive the keychain password field (keychain only) |

## Field Details

### metadata.name (required)
The unique identifier for this credential within its scope. Two credentials in different scopes may share the same name — the narrower scope wins at resolution time (see [Scope Hierarchy](#scope-hierarchy)).

**Examples:**
- `github-token` — GitHub Personal Access Token
- `docker-registry` — Container registry credentials
- `db-password` — Database password
- `api-key` — External service API key

### metadata scope (exactly one required)
Credentials must be scoped to exactly one resource. Specify the name of the target resource in the matching field:

```yaml
metadata:
  name: my-credential
  ecosystem: prod-platform   # Option 1: ecosystem scope
  # domain: backend          # Option 2: domain scope
  # app: my-api              # Option 3: app scope
  # workspace: dev           # Option 4: workspace scope
```

### spec.source (required)
Defines where the secret value is stored at runtime.

| Value | Description |
|-------|-------------|
| `keychain` | macOS Keychain item (requires `spec.service`) |
| `env` | Host environment variable (requires `spec.envVar`) |

### spec.service (required for keychain)
The macOS Keychain service name used to look up the item.

```yaml
spec:
  source: keychain
  service: com.github.token   # Keychain service identifier
```

The Keychain item's **password field** is the value retrieved by default. Use `usernameVar` / `passwordVar` to split a single Keychain entry into two environment variables.

### spec.envVar (required for env)
The name of a host environment variable. The value is read at resolution time from `os.Getenv`.

```yaml
spec:
  source: env
  envVar: GITHUB_TOKEN   # Must exist in the host environment
```

### spec.usernameVar and spec.passwordVar (optional, keychain only)
Dual-field extraction splits a single Keychain entry into two injected environment variables:

- `usernameVar` — receives the Keychain item's **account** field
- `passwordVar` — receives the Keychain item's **password** field

Both are optional and independent — you can specify one or both.

```yaml
spec:
  source: keychain
  service: com.registry.docker
  usernameVar: DOCKER_USERNAME   # account field → DOCKER_USERNAME
  passwordVar: DOCKER_PASSWORD   # password field → DOCKER_PASSWORD
```

`usernameVar` and `passwordVar` values must be valid environment variable names (uppercase letters, digits, and underscores; must not start with a digit).

### spec.description (optional)
A human-readable note about what this credential is used for.

```yaml
spec:
  description: "GitHub PAT with read:packages and repo scopes"
```

## Scope Hierarchy

Credentials are resolved from broadest to narrowest scope. A credential defined at the workspace level overrides the same credential name at the app, domain, or ecosystem level.

```
ecosystem → domain → app → workspace
 (lowest)                   (highest)
```

When dvm resolves credentials for an active workspace, it merges all credentials from each ancestor scope, with narrower scopes taking priority.

**Override behavior:**  
If `GITHUB_TOKEN` is defined at both the ecosystem scope and the app scope, the app-scoped value wins when working in that app.

**Environment variable precedence:**  
Host environment variables always take the highest priority. If `DOCKER_PASSWORD` is already set in the host environment, that value is used regardless of any Keychain or credential configuration.

## CLI Commands

### `dvm create credential`

Create a new credential.

```
dvm create credential <name> [flags]
dvm create cred <name> [flags]        # Alias
```

**Source flags:**

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `--source` | string | ✅ | Secret source: `keychain` or `env` |
| `--service` | string | ✅ (keychain) | Keychain service name |
| `--env-var` | string | ✅ (env) | Environment variable name |
| `--description` | string | ❌ | Human-readable description |
| `--username-var` | string | ❌ | Env var for keychain account field (keychain only) |
| `--password-var` | string | ❌ | Env var for keychain password field (keychain only) |

**Scope flags (exactly one required):**

| Flag | Short | Description |
|------|-------|-------------|
| `--ecosystem` | `-e` | Scope to an ecosystem |
| `--domain` | `-d` | Scope to a domain |
| `--app` | `-a` | Scope to an app |
| `--workspace` | `-w` | Scope to a workspace |

**Examples:**

```bash
# Keychain credential scoped to an app
dvm create credential github-token \
  --source keychain --service com.github.token \
  --app my-api

# Environment variable credential scoped to an ecosystem
dvm create credential api-key \
  --source env --env-var MY_API_KEY \
  --ecosystem prod

# Dual-field keychain credential scoped to a domain
dvm create credential docker-registry \
  --source keychain --service com.registry.docker \
  --username-var DOCKER_USERNAME \
  --password-var DOCKER_PASSWORD \
  --domain backend

# With description
dvm create cred db-pass \
  --source keychain --service com.db.password \
  --description "Postgres prod password" \
  --app my-api
```

---

### `dvm get credential`

Get a single credential by name within a scope.

```
dvm get credential <name> [scope-flag]
dvm get cred <name> [scope-flag]       # Alias
```

**Scope flags (exactly one required):**

| Flag | Short | Description |
|------|-------|-------------|
| `--ecosystem` | `-e` | Look up in an ecosystem |
| `--domain` | `-d` | Look up in a domain |
| `--app` | `-a` | Look up in an app |
| `--workspace` | `-w` | Look up in a workspace |

**Examples:**

```bash
dvm get credential github-token --app my-api
dvm get credential api-key --ecosystem prod
dvm get cred db-pass --domain backend
```

**Output:**

```
Name:      github-token
Scope:     app (ID: 3)
Source:    keychain
Service:   com.github.token
Desc:      GitHub PAT for pulling private packages
```

---

### `dvm get credentials`

List credentials, either by scope or across all scopes.

```
dvm get credentials [flags]
dvm get creds [flags]                  # Alias
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--all` | `-A` | List all credentials across every scope |
| `--ecosystem` | `-e` | Filter to an ecosystem |
| `--domain` | `-d` | Filter to a domain |
| `--app` | `-a` | Filter to an app |
| `--workspace` | `-w` | Filter to a workspace |

**Examples:**

```bash
# List all credentials (every scope)
dvm get credentials --all
dvm get credentials -A

# List credentials for a specific app
dvm get credentials --app my-api

# List credentials for an ecosystem
dvm get credentials --ecosystem prod
```

---

### `dvm delete credential`

Delete a credential by name within a scope.

```
dvm delete credential <name> [scope-flag] [flags]
dvm delete cred <name> [scope-flag] [flags]    # Alias
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip the confirmation prompt |
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

### `dvm apply` (YAML)

Create a credential from a YAML file.

```bash
dvm apply -f credential.yaml
```

## Examples

### Simple Keychain Credential

A GitHub PAT stored in the macOS Keychain, available to a single app:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-token
  app: api-service
spec:
  source: keychain
  service: com.github.token
  description: "GitHub PAT with read:packages scope"
```

**Create equivalent:**
```bash
dvm create credential github-token \
  --source keychain --service com.github.token \
  --app api-service
```

---

### Dual-Field Keychain Credential

A single Keychain entry split into username and password environment variables — useful for container registries:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: docker-registry
  ecosystem: prod-platform
spec:
  source: keychain
  service: com.registry.docker
  description: "Docker Hub login"
  usernameVar: DOCKER_USERNAME
  passwordVar: DOCKER_PASSWORD
```

When resolved, `DOCKER_USERNAME` receives the Keychain account field and `DOCKER_PASSWORD` receives the password field.

---

### Environment Variable Credential

A credential that reads from a host environment variable — useful for CI pipelines or values that rotate frequently:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: api-key
  domain: backend
spec:
  source: env
  envVar: EXTERNAL_API_KEY
  description: "External payment gateway key"
```

---

### Scoped Credentials at Different Levels

Credentials can be defined at every level of the hierarchy. Narrower scopes override broader ones.

```yaml
# Ecosystem-level: applies to all resources in this ecosystem
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-token
  ecosystem: prod-platform
spec:
  source: keychain
  service: com.github.prod.token
---
# App-level: overrides the ecosystem credential for this specific app
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-token
  app: special-service
spec:
  source: keychain
  service: com.github.special.token
  description: "Separate token for this app's private repos"
```

---

### Apply from YAML

```bash
# Apply a single credential
dvm apply -f credential.yaml

# Apply multiple resources in one file (multi-document YAML)
dvm apply -f all-credentials.yaml
```

## Validation Rules

- `metadata.name` is required and must be non-empty
- Exactly one scope field (`ecosystem`, `domain`, `app`, or `workspace`) must be set in `metadata`
- `spec.source` is required and must be `keychain` or `env`
- `spec.service` is required when `spec.source` is `keychain`
- `spec.envVar` is required when `spec.source` is `env`
- `spec.usernameVar` and `spec.passwordVar` are only valid when `spec.source` is `keychain`
- `spec.usernameVar` and `spec.passwordVar` must be valid environment variable names (e.g., `UPPER_SNAKE_CASE`)
- Plaintext values are not supported — `source` must be `keychain` or `env`

## Notes

- **Credentials never store secret values.** Only the source reference (Keychain service name or env var name) is stored in the database.
- **macOS Keychain only.** The `keychain` source uses the macOS Security framework. It is not available on Linux or Windows.
- **Host env takes precedence.** If the same variable name already exists in the host environment, the host value is always used — Keychain and config values do not override it.
- **Dual-field credentials** inject two separate environment variables from a single Keychain item, which is common for services requiring a username and password pair (container registries, npm, Maven).
- **Scope resolution** is performed at `dvm` runtime, not at creation time. A credential scoped to an app is available to all workspaces of that app.

## Related Resources

- [Ecosystem](ecosystem.md) - Broadest credential scope
- [Domain](domain.md) - Domain-level credential scope
- [App](app.md) - App-level credential scope
- [Workspace](workspace.md) - Narrowest credential scope
