# Registry YAML Reference

**Kind:** `Registry`  
**APIVersion:** `devopsmaestro.io/v1`

A Registry is a locally managed package registry or proxy — an OCI image registry, a language package index, or an HTTP proxy cache. Registries are standalone resources, not scoped to any app or workspace. They run as local processes managed by `dvm`.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: zot-local
  description: "OCI container image registry for local development"
spec:
  type: zot
  version: "2.1.15"
  port: 5001
  lifecycle: on-demand
  config:
    log:
      level: warn
    storage:
      dedupe: true
```

## Other Registry Type Examples

### Go Module Proxy (athens)

```yaml
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: go-proxy
  description: "Athens Go module proxy"
spec:
  type: athens
  lifecycle: persistent
```

### Python Package Index (devpi)

```yaml
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: pypi-local
spec:
  type: devpi
  port: 3141
  lifecycle: manual
```

### npm Registry (verdaccio)

```yaml
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: npm-local
spec:
  type: verdaccio
  lifecycle: on-demand
```

### HTTP Proxy Cache (squid)

```yaml
apiVersion: devopsmaestro.io/v1
kind: Registry
metadata:
  name: squid-proxy
  description: "HTTP/HTTPS caching proxy"
spec:
  type: squid
  port: 3128
  lifecycle: persistent
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | Yes | Must be `devopsmaestro.io/v1` |
| `kind` | string | Yes | Must be `Registry` |
| `metadata.name` | string | Yes | Unique name for the registry |
| `metadata.description` | string | No | Human-readable description |
| `spec.type` | string | Yes | Registry type: `zot`, `athens`, `devpi`, `verdaccio`, `squid` |
| `spec.version` | string | No | Desired binary version (e.g., `2.1.15`, `1.0.0-rc1`) — empty means use strategy default |
| `spec.enabled` | bool | No | Whether the registry is enabled (default: `true`) |
| `spec.port` | int | No | Port to listen on (0 means auto-assign from type default; must be 1024–65535 if set) |
| `spec.lifecycle` | string | No | How the process is managed: `persistent`, `on-demand`, `manual` (default: `manual`) |
| `spec.storage` | string | No | Storage path for registry data (default: type-specific path, e.g., `/var/lib/zot`) |
| `spec.idleTimeout` | int | No | Seconds before auto-stopping an `on-demand` registry (default: `1800`; min: `60` if set) |
| `spec.config` | map | No | Registry-type-specific configuration as key-value pairs |
| `status.state` | string | Read-only | Live process state: `stopped`, `starting`, `running`, `error` |
| `status.endpoint` | string | Read-only | Registry endpoint URL (e.g., `http://localhost:5001`) |

## Field Details

### metadata.name (required)

The unique identifier for this registry. Registry names are global — they are not scoped to any app or workspace.

**Examples:**
- `zot-local` — Local OCI container image registry
- `go-proxy` — Go module proxy for private modules
- `npm-local` — npm registry for private packages
- `pypi-internal` — Internal Python package index

### metadata.description (optional)

A human-readable note about what this registry is for.

```yaml
metadata:
  name: zot-local
  description: "OCI registry for dev and CI image caching"
```

### spec.type (required)

The registry implementation to use. Each type runs a different software stack and serves a different package ecosystem. See [Registry Types](#registry-types) for details.

```yaml
spec:
  type: zot
```

### spec.version (optional)

The desired binary version to run. Must be semver format without a leading `v` (e.g., `2.1.15`, `1.0.0-rc1`). When empty, `dvm` applies the strategy-layer default version for the registry type.

```yaml
spec:
  version: "2.1.15"
```

### spec.enabled (optional)

Whether the registry is enabled. Defaults to `true` when omitted. Set to `false` to disable the registry without deleting it.

```yaml
spec:
  enabled: false
```

### spec.port (optional)

The port the registry process will listen on. When omitted or set to `0`, the type-specific default port is applied automatically. If set, must be in the range 1024–65535.

```yaml
spec:
  port: 5001
```

### spec.lifecycle (optional)

Controls how `dvm` manages the registry process. Defaults to `manual` when not set.

| Value | Description |
|-------|-------------|
| `persistent` | Always running; starts with the system |
| `on-demand` | Auto-starts when needed, auto-stops after idle timeout |
| `manual` | User controls start and stop (default) |

For `on-demand` registries, the idle timeout defaults to 1800 seconds (30 minutes). A custom value must be at least 60 seconds.

```yaml
spec:
  lifecycle: on-demand
```

### spec.config (optional)

Registry-type-specific configuration passed directly to the registry process. The structure varies by type. The value must be a valid YAML map (serialized as JSON internally).

```yaml
spec:
  config:
    log:
      level: warn
    storage:
      dedupe: true
```

### spec.storage (optional)

The filesystem path where the registry stores its data. When omitted, the type-specific default storage path is applied automatically (see [Registry Types](#registry-types)). The resolved path is stored in the database and cannot be empty.

```yaml
spec:
  storage: "/data/my-zot-registry"
```

### spec.idleTimeout (optional)

Only applies to `on-demand` registries. The number of seconds after the last request before the registry process is automatically stopped. Defaults to `1800` seconds (30 minutes) when not set. Must be at least `60` if specified explicitly. A value of `0` means "use the default".

```yaml
spec:
  lifecycle: on-demand
  idleTimeout: 3600   # Auto-stop after 60 minutes of inactivity
```

### status (read-only)

The `status` section is populated by `dvm` at runtime. It is not applied from YAML — any `status` values in an applied YAML file are ignored.

```yaml
status:
  state: running        # stopped, starting, running, error
  endpoint: http://localhost:5001
```

## Registry Types

| Type | Description | Default Port | Default Storage |
|------|-------------|-------------|-----------------|
| `zot` | OCI container image registry | 5001 | `/var/lib/zot` |
| `athens` | Go module proxy | 3000 | `/var/lib/athens` |
| `devpi` | Python package index | 3141 | `/var/lib/devpi` |
| `verdaccio` | npm private registry | 4873 | `/var/lib/verdaccio` |
| `squid` | HTTP/HTTPS caching proxy | 3128 | `/var/cache/squid` |

Storage paths are applied as defaults when `storage` is not configured. The `dvm` datastore tracks the resolved storage path — it cannot be empty.

## Lifecycle

The `lifecycle` field controls how `dvm` manages the registry process over time.

### persistent

The registry is always running. `dvm` starts it on initialization and keeps it alive.

```yaml
spec:
  lifecycle: persistent
```

Use `persistent` for registries that must always be available — for example, a shared OCI registry accessed by CI scripts and containers alike.

### on-demand

The registry auto-starts when a dependent process or workspace requests it, and auto-stops after the idle timeout expires. The idle timeout defaults to 1800 seconds (30 minutes) and must be at least 60 seconds if set explicitly.

```yaml
spec:
  lifecycle: on-demand
```

Use `on-demand` for registries that are used intermittently and should not consume resources when idle.

### manual (default)

The user explicitly starts and stops the registry with `dvm registry start` and `dvm registry stop`. This is the default when `lifecycle` is not specified.

```yaml
spec:
  lifecycle: manual
```

## CLI Commands

### `dvm create registry`

Create a new registry.

```
dvm create registry <name> [flags]
dvm create reg <name> [flags]        # Alias
```

**Flags:**

| Flag | Short | Type | Required | Description |
|------|-------|------|----------|-------------|
| `--type` | `-t` | string | Yes | Registry type: `zot`, `athens`, `devpi`, `verdaccio`, `squid` |
| `--port` | `-p` | int | No | Port number (default: type-specific) |
| `--lifecycle` | `-l` | string | No | Lifecycle mode: `persistent`, `on-demand`, `manual` (default: `manual`) |
| `--version` | | string | No | Desired binary version (e.g., `2.1.15`) |
| `--description` | `-d` | string | No | Registry description |

**Examples:**

```bash
# Create a Zot OCI registry with defaults
dvm create registry zot-local --type zot

# Create a verdaccio npm registry on a custom port
dvm create registry npm-local --type verdaccio --port 4880

# Create a persistent Athens Go module proxy
dvm create registry go-proxy --type athens --lifecycle persistent

# Create a devpi registry with a description
dvm create registry pypi --type devpi --description "Python packages"

# Pin a specific version
dvm create registry zot-local --type zot --version 2.1.15
```

---

### `dvm get registry`

Get a single registry by name.

```
dvm get registry <name>
dvm get reg <name>                   # Alias
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output format: `json`, `yaml` |

**Examples:**

```bash
dvm get registry zot-local
dvm get reg go-proxy
dvm get registry zot-local -o yaml
dvm get registry zot-local -o json
```

**Output:**

```
Registry Details
Name:        zot-local
Type:        zot
Version:     2.1.15
Port:        5001
Lifecycle:   on-demand
Status:      running
Description: OCI registry for dev and CI image caching
Created:     2026-03-15T10:00:00Z
```

---

### `dvm get registries`

List all registries.

```
dvm get registries
dvm get reg                          # Alias
dvm get regs                         # Alias
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--output` | `-o` | Output format: `json`, `yaml`, `wide` |

**Examples:**

```bash
# List all registries (table view)
dvm get registries

# Show additional columns (including CREATED timestamp)
dvm get registries -o wide

# Output as YAML
dvm get registries -o yaml
```

**Output columns:** `NAME`, `TYPE`, `VERSION`, `PORT`, `LIFECYCLE`, `STATE`, `UPTIME`  
**Wide output adds:** `CREATED`

---

### `dvm delete registry`

Delete a registry by name. If the registry is running, it is automatically stopped before the database record is removed. If stopping fails, the record is not deleted.

```
dvm delete registry <name> [flags]
dvm delete reg <name> [flags]        # Alias
```

**Flags:**

| Flag | Short | Description |
|------|-------|-------------|
| `--force` | `-f` | Skip the confirmation prompt |

**Examples:**

```bash
dvm delete registry zot-local
dvm delete reg go-proxy
dvm delete registry npm-local --force
dvm delete registry pypi -f
```

---

### `dvm apply` (YAML)

Create or update a registry from a YAML file. If a registry with the same name already exists, it is updated.

```bash
dvm apply -f registry.yaml
```

---

### `dvm registry enable` / `dvm registry disable`

Enable or disable a registry by its package-ecosystem type alias. These commands operate on type aliases (`oci`, `pypi`, `npm`, `go`, `http`) rather than registry names, and manage the default registry mapping for each type.

```
dvm registry enable <type> [--lifecycle <mode>]
dvm registry disable <type>
```

**Lifecycle flag for enable:**

| Flag | Default | Description |
|------|---------|-------------|
| `--lifecycle` | `on-demand` | Lifecycle mode: `persistent`, `on-demand`, `manual` |

**Examples:**

```bash
dvm registry enable oci
dvm registry enable pypi --lifecycle persistent
dvm registry disable npm
```

---

### `dvm registry set-default` / `dvm registry get-defaults`

Manage which registry is used by default for each package type.

```
dvm registry set-default <type> <registry-name>
dvm registry get-defaults
```

**Examples:**

```bash
# Set a specific registry as the default for OCI
dvm registry set-default oci zot-local

# Set the default Go module proxy
dvm registry set-default go go-proxy

# Show all current defaults
dvm registry get-defaults
```

**`get-defaults` output columns:** `TYPE`, `REGISTRY`, `ENDPOINT`, `STATUS`

## Validation Rules

- `metadata.name` is required and must be non-empty
- `spec.type` is required and must be one of: `zot`, `athens`, `devpi`, `verdaccio`, `squid`
- `spec.port` must be `0` (auto-assign) or between `1024` and `65535` inclusive
- `spec.lifecycle` must be one of: `persistent`, `on-demand`, `manual` — defaults to `manual` when omitted
- `spec.version` must be semver format (e.g., `2.1.15`, `1.0.0-rc1`) when set — empty is valid and triggers the strategy-layer default
- For `on-demand` registries, a custom `idleTimeout` must be at least 60 seconds; `0` means use the default of 1800 seconds (30 minutes)
- Storage path is required and cannot be empty — defaults are applied automatically per type
- `spec.config` must be valid if present (serializable as JSON)

## Notes

- **Registries are global.** They are not scoped to an app, domain, or workspace. Any workspace can reference any registry.
- **Status is live.** The `state` field in `status` reflects the actual process state at query time (checked via PID file), not a cached database value.
- **`dvm apply` is idempotent.** Applying a registry YAML when a registry of the same name already exists updates the existing record.
- **Deleting a running registry auto-stops it.** `dvm delete registry` calls `Stop()` before removing the database record. If stopping fails, the record is preserved and an error is returned.
- **Default ports are applied automatically.** If `spec.port` is `0` or omitted, the type-specific default port is used. The resolved port is stored in the database.
- **Default storage paths are applied automatically.** The resolved path is stored in the database and cannot be empty after defaults are applied.
- **Version defaults come from the strategy layer.** When `spec.version` is empty, the registry strategy for that type supplies the default version string.

## Related Resources

- [Workspace](workspace.md) - Development environments that may use registries
- [App](app.md) - Applications whose workspaces consume registry services
- [Credential](credential.md) - Secrets for registry authentication
