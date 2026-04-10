# Architecture Overview

How DevOpsMaestro is structured from a user's perspective.

---

## Three Tools, One Database

DevOpsMaestro provides three CLI tools that share a single SQLite database at `~/.devopsmaestro/devopsmaestro.db`:

| Tool | Binary | What It Manages |
|------|--------|-----------------|
| **DevOpsMaestro** | `dvm` | Apps, workspaces, registries, git repos, credentials, CRDs |
| **NvimOps** | `nvp` | Neovim plugins and themes |
| **Terminal Operations** | `dvt` | Terminal prompts, plugins, and configurations |

All three tools share the same object hierarchy and context.

---

## Object Hierarchy

Resources are organized in a four-level hierarchy:

```
Ecosystem → Domain → App → Workspace
```

| Object | Purpose |
|--------|---------|
| **Ecosystem** | Top-level platform grouping (e.g., your company or product area) |
| **Domain** | Bounded context within an ecosystem (e.g., a team or service area) |
| **App** | Your codebase — the thing you build and maintain |
| **Workspace** | A development environment for an App (runs in a container) |

### Context

Your active context determines which ecosystem, domain, app, and workspace commands operate on by default. Check it at any time:

```bash
dvm get context     # Show active context
dvm get ctx         # Short alias
```

---

## Resource Types

DevOpsMaestro manages 12 built-in resource types:

| Resource | Kind | Purpose |
|----------|------|---------|
| Ecosystem | `Ecosystem` | Organizational grouping |
| Domain | `Domain` | Bounded context |
| App | `App` | Application registration |
| Workspace | `Workspace` | Dev environment container |
| Credential | `Credential` | Stored credentials for auth |
| Registry | `Registry` | OCI/package registry config |
| Git Repository | `GitRepo` | Bare git mirror |
| Neovim Plugin | `NvimPlugin` | Neovim plugin definition |
| Neovim Theme | `NvimTheme` | Neovim color theme |
| Neovim Package | `NvimPackage` | Curated plugin bundle |
| Terminal Prompt | `TerminalPrompt` | Prompt configuration |
| Terminal Package | `TerminalPackage` | Terminal extension bundle |

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

## External Modules

DevOpsMaestro is built on a set of focused libraries that can also be used independently:

| Module | Purpose |
|--------|---------|
| [MaestroNvim](https://github.com/rmkohlman/MaestroNvim) | Neovim plugin and theme management (`nvp`) |
| [MaestroTheme](https://github.com/rmkohlman/MaestroTheme) | Theme and color palette system |
| [MaestroTerminal](https://github.com/rmkohlman/MaestroTerminal) | Terminal prompt and plugin management (`dvt`) |
| [MaestroSDK](https://github.com/rmkohlman/MaestroSDK) | Shared foundation (paths, rendering, resource handling) |
| [MaestroPalette](https://github.com/rmkohlman/MaestroPalette) | Color palette primitives |

---

## See Also

- [Source Types](source-types.md) — How the `-f` flag resolves sources
- [Theme Hierarchy](theme-hierarchy.md) — How themes cascade through the object hierarchy
- [Custom Resource Definitions](../reference/custom-resource-definition.md) — Extending DevOpsMaestro with custom types
- [Contributing](../development/contributing.md) — Contribute to development
