# GitRepo YAML Reference

**Kind:** `GitRepo`  
**APIVersion:** `devopsmaestro.io/v1`

A GitRepo defines a remote git repository that DevOpsMaestro mirrors locally. When applied, `dvm` stores the repository configuration in its database and clones a bare mirror to `~/.devopsmaestro/repos/`. Workspaces can reference the mirror for fast, offline-capable builds.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: api-service
  labels:
    team: backend
    language: go
  annotations:
    description: "Main API service repository"
spec:
  url: "https://github.com/myorg/api-service.git"
  defaultRef: main
  authType: ssh
  credential: github-ssh-key
  autoSync: true
  syncIntervalMinutes: 60
```

## Minimal Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: public-lib
spec:
  url: "https://github.com/someorg/public-lib.git"
```

## SSH Authenticated Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: GitRepo
metadata:
  name: private-service
  labels:
    visibility: private
spec:
  url: "git@github.com:myorg/private-service.git"
  defaultRef: develop
  authType: ssh
  credential: github-deploy-key
  autoSync: true
  syncIntervalMinutes: 30
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | Yes | Must be `devopsmaestro.io/v1` |
| `kind` | string | Yes | Must be `GitRepo` |
| `metadata.name` | string | Yes | Unique name for the git repository |
| `metadata.labels` | map[string]string | No | Key-value labels for filtering |
| `metadata.annotations` | map[string]string | No | Non-identifying metadata |
| `spec.url` | string | Yes | Remote git repository URL (HTTPS or SSH) |
| `spec.defaultRef` | string | No | Default branch or tag to check out (default: `main`) |
| `spec.authType` | string | No | Authentication method: `none`, `ssh`, `basic` (default: `none`) |
| `spec.credential` | string | No | Name of a Credential resource used for authentication |
| `spec.autoSync` | bool | No | Automatically sync the mirror on a schedule (default: `false`) |
| `spec.syncIntervalMinutes` | int | No | How often to sync in minutes when `autoSync` is `true` |

## Field Details

### metadata.name (required)

The unique identifier for this git repository. GitRepo names are global — they are not scoped to any app or workspace.

**Examples:**
- `api-service` — Main API service repository
- `frontend-app` — React frontend repository
- `shared-libs` — Internal shared library monorepo

### metadata.labels (optional)

Key-value labels for organizing and filtering repositories.

```yaml
metadata:
  name: api-service
  labels:
    team: backend
    language: go
    visibility: private
```

### metadata.annotations (optional)

Non-identifying metadata attached to the resource. Useful for descriptions and tooling metadata.

```yaml
metadata:
  name: api-service
  annotations:
    description: "Core REST API for the platform"
    owner: "backend-team@example.com"
```

### spec.url (required)

The remote git repository URL. Supports HTTPS and SSH formats.

```yaml
# HTTPS (public or with basic auth)
spec:
  url: "https://github.com/myorg/api-service.git"

# SSH (with deploy key credential)
spec:
  url: "git@github.com:myorg/private-service.git"
```

### spec.defaultRef (optional)

The branch, tag, or commit reference to check out by default. When omitted, defaults to `main`.

```yaml
spec:
  defaultRef: develop     # Branch
  # defaultRef: v2.3.1   # Tag
```

### spec.authType (optional)

The authentication method used to access the repository. When omitted, defaults to `none`.

| Value | Description |
|-------|-------------|
| `none` | No authentication — public repositories only (default) |
| `ssh` | SSH key authentication — requires a `credential` referencing an SSH key |
| `basic` | HTTP basic authentication — requires a `credential` with username/password |

```yaml
spec:
  authType: ssh
  credential: github-deploy-key
```

### spec.credential (optional)

The name of an existing [Credential](credential.md) resource used for authentication. The credential is looked up by name across all scopes at apply time.

- For `authType: ssh` — the credential should reference an SSH private key
- For `authType: basic` — the credential should provide username and password values

```yaml
spec:
  authType: ssh
  credential: github-deploy-key   # Must exist as a Credential resource
```

### spec.autoSync (optional)

When `true`, `dvm` periodically syncs the local mirror with the remote repository according to `syncIntervalMinutes`. Defaults to `false`.

```yaml
spec:
  autoSync: true
  syncIntervalMinutes: 60   # Sync every hour
```

### spec.syncIntervalMinutes (optional)

How frequently (in minutes) to sync the mirror when `autoSync` is `true`. Has no effect when `autoSync` is `false`.

```yaml
spec:
  autoSync: true
  syncIntervalMinutes: 30   # Sync every 30 minutes
```

## Behavior

### Mirror Initialization

When `dvm apply -f` processes a GitRepo resource, two things happen:

1. **Database record is created/updated** — the configuration is stored in the `dvm` database.
2. **Bare mirror is cloned** — `dvm` clones a bare mirror of the repository to `~/.devopsmaestro/repos/<slug>/`. If the clone fails (e.g., network unavailable), the database record is still saved and a warning is logged. The mirror can be cloned later via `dvm sync gitrepo <name>`.

### Slug Generation

Each GitRepo is assigned a filesystem-safe slug derived from its URL. The slug is used as the directory name under `~/.devopsmaestro/repos/`. Slugs are generated automatically and cannot be set in YAML.

### Idempotent Apply

`dvm apply -f` on an existing GitRepo updates the database record. If the mirror already exists on disk, it is not re-cloned. Use `dvm sync gitrepo <name>` to force a sync.

## CLI Commands

### `dvm apply` (YAML)

Create or update a git repository from a YAML file.

```bash
dvm apply -f gitrepo.yaml
```

---

### `dvm get gitrepo`

Get a single git repository by name.

```
dvm get gitrepo <name>
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output format: `json`, `yaml` |

**Examples:**

```bash
dvm get gitrepo api-service
dvm get gitrepo api-service -o yaml
dvm get gitrepo api-service -o json
```

---

### `dvm get gitrepos`

List all git repositories.

```
dvm get gitrepos
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output format: `json`, `yaml` |

**Examples:**

```bash
dvm get gitrepos
dvm get gitrepos -o yaml
```

---

### `dvm delete gitrepo`

Delete a git repository record from the database. Does not remove the local mirror from disk.

```
dvm delete gitrepo <name> [flags]
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip the confirmation prompt |

**Examples:**

```bash
dvm delete gitrepo old-service
dvm delete gitrepo old-service --force
dvm delete gitrepo old-service -f
```

## Validation Rules

- `metadata.name` is required and must be non-empty
- `spec.url` is required and must be non-empty
- `spec.defaultRef` defaults to `main` when omitted
- `spec.authType` defaults to `none` when omitted
- `spec.credential` must reference an existing Credential resource by name when set
- `spec.syncIntervalMinutes` has no effect unless `spec.autoSync` is `true`

## Notes

- **GitRepos are global.** They are not scoped to an app, domain, or workspace. Any workspace can reference any registered repository.
- **Mirrors live at `~/.devopsmaestro/repos/`.** The local bare mirror is used by workspace builds for fast, offline-capable clones.
- **`dvm apply` is idempotent.** Applying a GitRepo YAML when a record with the same name exists updates the database record. An existing mirror on disk is not re-cloned.
- **Clone failures are non-fatal.** If the remote is unreachable at apply time, the database record is saved and a warning is emitted. Sync the mirror later with `dvm sync gitrepo <name>`.
- **Credentials are resolved at apply time.** The `spec.credential` field is resolved to a credential ID when the resource is applied — not at sync time.

## Related Resources

- [Credential](credential.md) - Authentication secrets for private repositories
- [Workspace](workspace.md) - Development environments that may use git repositories
- [App](app.md) - Applications that may reference git repositories
