# Architecture Overview

How `dvm` (DevOpsMaestro) is structured from a user's perspective.

---

## Object Hierarchy

Resources are organized in a five-level hierarchy:

```
Ecosystem → Domain → System → App → Workspace
```

| Object | Purpose |
|--------|---------|
| **Ecosystem** | Top-level platform grouping (e.g., your company or product area) |
| **Domain** | Bounded context within an ecosystem (e.g., a team or service area) |
| **System** | Logical grouping of related apps within a domain (optional) |
| **App** | Your codebase — the thing you build and maintain |
| **Workspace** | A development environment for an App (runs in a container) |

All intermediate levels (Ecosystem, Domain, System) are optional — only Workspace is required.

### Context

Your active context determines which ecosystem, domain, system, app, and workspace commands operate on by default. Check it at any time:

```bash
dvm get context     # Show active context
dvm get ctx         # Short alias
```

---

## Resource Types

DevOpsMaestro manages 8 built-in resource types:

| Resource | Kind | Purpose |
|----------|------|---------|
| Ecosystem | `Ecosystem` | Organizational grouping |
| Domain | `Domain` | Bounded context |
| System | `System` | Logical app cluster within a domain |
| App | `App` | Application registration |
| Workspace | `Workspace` | Dev environment container |
| Credential | `Credential` | Stored credentials for auth |
| Registry | `Registry` | OCI/package registry config |
| Git Repository | `GitRepo` | Bare git mirror |

You can also define your own resource types using [Custom Resource Definitions](../reference/custom-resource-definition.md).

---

## How `apply` Works

The `-f` flag accepts files, URLs, GitHub shorthands, and stdin. When you run `dvm apply -f something.yaml`, DevOpsMaestro:

1. **Resolves the source** — reads from a file, URL, GitHub path, or stdin
2. **Parses the YAML** — detects the resource `kind`
3. **Validates and saves** — validates the resource and stores it in the database

```
Source (file/URL/GitHub/stdin) → Parse YAML → Validate → Save
```

---

## Database

All state is stored in a single SQLite database:

```
~/.devopsmaestro/devopsmaestro.db
```

The database is automatically migrated to the latest schema on every startup. No manual migration steps are needed.

---

## Container Runtime Detection

`dvm` automatically detects your container platform in this order:

1. `DVM_PLATFORM` environment variable (override)
2. OrbStack
3. Docker Desktop
4. Colima
5. Podman
6. containerd (nerdctl)

Check what was detected:

```bash
dvm get platforms
```

---

## See Also

- [Source Types](source-types.md) — How the `-f` flag resolves sources
- [Custom Resource Definitions](../reference/custom-resource-definition.md) — Extending DevOpsMaestro with custom types
- [Contributing](https://github.com/rmkohlman/devopsmaestro/blob/main/CONTRIBUTING.md) — Contribute to development
