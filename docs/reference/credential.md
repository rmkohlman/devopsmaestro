# Credential YAML Reference

**Kind:** `Credential`  
**APIVersion:** `devopsmaestro.io/v1`

A Credential stores a reference to a secret — not the secret itself. Credentials point to values stored in MaestroVault or in environment variables. They are scoped to exactly one resource (ecosystem, domain, app, or workspace) and resolve automatically when that scope is active.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-token
  app: api-service           # Scoped to this app
spec:
  source: vault
  vaultSecret: "github-pat"
  vaultEnvironment: production
  description: "GitHub PAT for pulling private packages"
```

## Vault Fields Example

A single MaestroVault secret with multiple fields, each mapped to a separate environment variable:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: db-creds
  app: api-service
spec:
  source: vault
  vaultSecret: "database/prod"
  description: "Database connection credentials"
  vaultFields:
    DB_HOST: host
    DB_PORT: port
    DB_PASSWORD: password
```

## Dual-Field Vault Example

A single MaestroVault secret split into a username variable and a password variable:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: docker-registry
  ecosystem: prod-platform   # Scoped to this ecosystem
spec:
  source: vault
  vaultSecret: "hub.docker.com"
  description: "Docker Hub credentials"
  usernameVar: DOCKER_USERNAME   # Injected from vault username field
  passwordVar: DOCKER_PASSWORD   # Injected from vault password field
```

## Environment Variable Example

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
| `apiVersion` | string | Yes | Must be `devopsmaestro.io/v1` |
| `kind` | string | Yes | Must be `Credential` |
| `metadata.name` | string | Yes | Unique name for the credential within its scope |
| `metadata.ecosystem` | string | No | Scope to an ecosystem (exactly one scope required) |
| `metadata.domain` | string | No | Scope to a domain (exactly one scope required) |
| `metadata.app` | string | No | Scope to an app (exactly one scope required) |
| `metadata.workspace` | string | No | Scope to a workspace (exactly one scope required) |
| `spec.source` | string | Yes | Where the secret lives: `vault` or `env` |
| `spec.vaultSecret` | string | Yes (vault) | MaestroVault secret name — required when `source: vault` |
| `spec.vaultEnvironment` | string | No | MaestroVault environment (e.g., `production`) — vault source only |
| `spec.vaultUsernameSecret` | string | No | Separate MaestroVault secret name for the username — vault source only; requires `usernameVar` |
| `spec.vaultFields` | map[string]string | No | Map of env var names to vault field names — vault source only; mutually exclusive with `usernameVar`/`passwordVar` |
| `spec.envVar` | string | Yes (env) | Environment variable name — required when `source: env` |
| `spec.description` | string | No | Human-readable description of the credential |
| `spec.usernameVar` | string | No | Env var name to receive the vault username value — vault source only |
| `spec.passwordVar` | string | No | Env var name to receive the vault password value — vault source only |

## Field Details

### metadata.name (required)

The unique identifier for this credential within its scope. Two credentials in different scopes may share the same name — the narrower scope wins at resolution time (see [Scope Hierarchy](#scope-hierarchy)).

For simple vault credentials (no `usernameVar`, `passwordVar`, or `vaultFields`), the credential name becomes the environment variable key when the credential is resolved.

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

Defines where the secret value is retrieved from at runtime.

| Value | Description |
|-------|-------------|
| `vault` | MaestroVault secret (requires `spec.vaultSecret`) |
| `env` | Host environment variable (requires `spec.envVar`) |

### spec.vaultSecret (required for vault)

The MaestroVault secret name used to look up the credential.

```yaml
spec:
  source: vault
  vaultSecret: "github-pat"
```

For simple credentials, the entire secret value is retrieved and injected as the environment variable named after the credential. For dual-field and vault fields credentials, the secret value is split into multiple environment variables.

### spec.vaultEnvironment (optional, vault only)

The MaestroVault environment to target when retrieving the secret. When omitted, MaestroVault uses its default environment.

```yaml
spec:
  source: vault
  vaultSecret: "github-pat"
  vaultEnvironment: production
```

### spec.vaultUsernameSecret (optional, vault only)

An alternate MaestroVault secret name used exclusively for the username field of a dual-field credential. When set, `usernameVar` reads from this secret instead of from `vaultSecret`. Requires `usernameVar` to be set.

```yaml
spec:
  source: vault
  vaultSecret: "docker-hub-password"
  vaultUsernameSecret: "docker-hub-username"
  usernameVar: DOCKER_USERNAME
  passwordVar: DOCKER_PASSWORD
```

### spec.vaultFields (optional, vault only)

A map of environment variable names to vault field names. Each entry fans out to one injected environment variable. Maximum 50 entries.

- Mutually exclusive with `usernameVar` and `passwordVar`
- Requires `vaultSecret`
- Environment variable keys must be valid env var names (uppercase letters, digits, and underscores; must not start with a digit)
- Field names cannot be empty

```yaml
spec:
  source: vault
  vaultSecret: "database/prod"
  vaultFields:
    DB_HOST: host
    DB_PORT: port
    DB_PASSWORD: password
    DB_USER: username
```

When resolved, this produces four injected environment variables: `DB_HOST`, `DB_PORT`, `DB_PASSWORD`, and `DB_USER`.

### spec.envVar (required for env)

The name of a host environment variable. The value is read from the host environment at resolution time.

```yaml
spec:
  source: env
  envVar: GITHUB_TOKEN   # Must exist in the host environment
```

### spec.usernameVar and spec.passwordVar (optional, vault only)

Dual-field extraction splits a single vault secret into two injected environment variables:

- `usernameVar` — receives the vault secret's username value
- `passwordVar` — receives the vault secret's password value

Both are optional and independent — you can specify one or both. When `vaultUsernameSecret` is also set, `usernameVar` reads from that separate secret instead of from `vaultSecret`.

Mutually exclusive with `vaultFields`.

```yaml
spec:
  source: vault
  vaultSecret: "hub.docker.com"
  usernameVar: DOCKER_USERNAME
  passwordVar: DOCKER_PASSWORD
```

`usernameVar` and `passwordVar` values must be valid environment variable names.

### spec.description (optional)

A human-readable note about what this credential is used for.

```yaml
spec:
  description: "GitHub PAT with read:packages and repo scopes"
```

## Scope Hierarchy

Credentials are resolved from broadest to narrowest scope. A credential defined at the workspace level overrides the same credential name at the app, domain, or ecosystem level.

```
global → ecosystem → domain → app → workspace
(lowest)                              (highest)
```

When `dvm` resolves credentials for an active workspace, it merges all credentials from each ancestor scope, with narrower scopes taking priority.

**Override behavior:**  
If `GITHUB_TOKEN` is defined at both the ecosystem scope and the app scope, the app-scoped value wins when working in that app.

**Environment variable precedence:**  
Host environment variables always take the highest priority. If `DOCKER_PASSWORD` is already set in the host environment, that value is used regardless of any vault or credential configuration.

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
| `--source` | string | Yes | Secret source: `vault` or `env` |
| `--vault-secret` | string | Yes (vault) | MaestroVault secret name |
| `--vault-env` | string | No | MaestroVault environment (e.g., `production`) |
| `--vault-username-secret` | string | No | MaestroVault secret name for username (vault only) |
| `--env-var` | string | Yes (env) | Environment variable name |
| `--description` | string | No | Human-readable description |
| `--username-var` | string | No | Env var for vault username value (vault only) |
| `--password-var` | string | No | Env var for vault password value (vault only) |
| `--vault-field` | string | No | Map vault field to env var (repeatable) — see format below |

**`--vault-field` format:**

```
--vault-field ENV_VAR=field_name   # Explicit: map field_name to ENV_VAR
--vault-field FIELD_NAME           # Auto-map: env var name equals field name
```

Repeatable. Maximum 50 entries. Mutually exclusive with `--username-var`, `--password-var`, and `--vault-username-secret`.

**Scope flags (exactly one required):**

| Flag | Short | Description |
|------|-------|-------------|
| `--ecosystem` | `-e` | Scope to an ecosystem |
| `--domain` | `-d` | Scope to a domain |
| `--app` | `-a` | Scope to an app |
| `--workspace` | `-w` | Scope to a workspace |

**Examples:**

```bash
# Simple vault credential scoped to an app
dvm create credential github-token \
  --source vault \
  --vault-secret "github-pat" \
  --vault-env production \
  --app my-api

# Environment variable credential scoped to an ecosystem
dvm create credential api-key \
  --source env \
  --env-var MY_API_KEY \
  --ecosystem prod

# Dual-field vault credential scoped to a domain
dvm create credential docker-registry \
  --source vault \
  --vault-secret "hub.docker.com" \
  --username-var DOCKER_USERNAME \
  --password-var DOCKER_PASSWORD \
  --domain backend

# Vault fields credential — one secret, multiple env vars
dvm create credential db-creds \
  --source vault \
  --vault-secret "database/prod" \
  --vault-field DB_HOST=host \
  --vault-field DB_PORT=port \
  --vault-field DB_PASSWORD=password \
  --app my-api

# Dual-field with a separate username secret
dvm create credential github-creds \
  --source vault \
  --vault-secret "github-pat" \
  --vault-username-secret "github-username" \
  --username-var GITHUB_USERNAME \
  --password-var GITHUB_PAT \
  --ecosystem myorg
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
Source:    vault
Secret:    github-pat
Vault Env: production
Desc:      GitHub PAT for pulling private packages
```

Fields are only printed when set. A vault fields credential additionally prints a `Fields:` block:

```
Name:      db-creds
Scope:     app (ID: 3)
Source:    vault
Secret:    database/prod
Fields:
  DB_HOST <- host
  DB_PASSWORD <- password
  DB_PORT <- port
```

When the env var name equals the field name (auto-map), only the name is printed:

```
Fields:
  DB_PASSWORD
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

### Simple Vault Credential

A GitHub PAT stored in MaestroVault, available to a single app. The credential name (`github-token`) becomes the injected environment variable key.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-token
  app: api-service
spec:
  source: vault
  vaultSecret: "github-pat"
  vaultEnvironment: production
  description: "GitHub PAT with read:packages scope"
```

**Create equivalent:**
```bash
dvm create credential github-token \
  --source vault \
  --vault-secret "github-pat" \
  --vault-env production \
  --app api-service
```

When resolved, the value of the `github-pat` secret is injected into the credentials map with the key `github-token` (the credential name).

---

### Dual-Field Vault Credential

A single vault secret split into username and password environment variables — useful for container registries, npm, or Maven:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: docker-registry
  ecosystem: prod-platform
spec:
  source: vault
  vaultSecret: "hub.docker.com"
  description: "Docker Hub login"
  usernameVar: DOCKER_USERNAME
  passwordVar: DOCKER_PASSWORD
```

**Create equivalent:**
```bash
dvm create credential docker-registry \
  --source vault \
  --vault-secret "hub.docker.com" \
  --username-var DOCKER_USERNAME \
  --password-var DOCKER_PASSWORD \
  --ecosystem prod-platform
```

When resolved, `DOCKER_USERNAME` receives the vault username value and `DOCKER_PASSWORD` receives the vault password value.

---

### Vault Fields Credential

One vault secret with multiple named fields, each mapped to a separate environment variable:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: db-creds
  app: api-service
spec:
  source: vault
  vaultSecret: "database/prod"
  description: "Database connection credentials"
  vaultFields:
    DB_HOST: host
    DB_PORT: port
    DB_USER: username
    DB_PASSWORD: password
```

**Create equivalent:**
```bash
dvm create credential db-creds \
  --source vault \
  --vault-secret "database/prod" \
  --vault-field DB_HOST=host \
  --vault-field DB_PORT=port \
  --vault-field DB_USER=username \
  --vault-field DB_PASSWORD=password \
  --app api-service
```

When resolved, this fans out to four environment variables: `DB_HOST`, `DB_PORT`, `DB_USER`, and `DB_PASSWORD`.

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

Credentials can be defined at every level of the hierarchy. Narrower scopes override broader ones when both define the same credential name.

```yaml
# Ecosystem-level: applies to all resources in this ecosystem
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-token
  ecosystem: prod-platform
spec:
  source: vault
  vaultSecret: "github-pat-shared"
  vaultEnvironment: production
---
# App-level: overrides the ecosystem credential for this specific app
apiVersion: devopsmaestro.io/v1
kind: Credential
metadata:
  name: github-token
  app: special-service
spec:
  source: vault
  vaultSecret: "github-pat-special-service"
  vaultEnvironment: production
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

- `kind` must be `Credential`
- `metadata.name` is required and must be non-empty
- Exactly one scope field (`ecosystem`, `domain`, `app`, or `workspace`) must be set in `metadata`
- `spec.source` is required and must be `vault` or `env`
- `spec.vaultSecret` is required when `spec.source` is `vault`
- `spec.envVar` is required when `spec.source` is `env`
- `spec.usernameVar` and `spec.passwordVar` are only valid when `spec.source` is `vault`
- `spec.vaultUsernameSecret` requires `spec.usernameVar` to be set
- `spec.vaultFields` is only valid when `spec.source` is `vault`
- `spec.vaultFields` requires `spec.vaultSecret`
- `spec.vaultFields` is mutually exclusive with `spec.usernameVar` and `spec.passwordVar`
- `spec.vaultFields` maximum 50 entries
- Vault field env var keys must be valid environment variable names (uppercase letters, digits, underscores; must not start with a digit)
- Vault field names cannot be empty

## Notes

- **Credentials never store secret values.** Only the source reference (vault secret name or env var name) is stored in the database.
- **MaestroVault is required for vault-sourced credentials.** The `vault` source requires MaestroVault to be installed and accessible at resolution time.
- **Simple vault credentials** use the credential name as the environment variable key in the resolved credentials map.
- **Dual-field credentials** inject two separate environment variables from a single vault secret, which is common for services requiring a username and password pair (container registries, npm, Maven).
- **Vault fields credentials** fan out a single vault secret with N fields into N injected environment variables. They cannot be combined with `usernameVar` or `passwordVar`.
- **`vaultUsernameSecret`** allows the username to come from a different vault secret than the password, for cases where the two values are stored separately.
- **Host env takes precedence.** If the same variable name already exists in the host environment, the host value always wins — vault and config values do not override it.
- **Scope resolution** is performed at `dvm` runtime, not at creation time. A credential scoped to an app is available to all workspaces of that app.

## Related Resources

- [Ecosystem](ecosystem.md) - Broadest credential scope
- [Domain](domain.md) - Domain-level credential scope
- [App](app.md) - App-level credential scope
- [Workspace](workspace.md) - Narrowest credential scope
