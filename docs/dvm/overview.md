# dvm Overview

`dvm` (DevOpsMaestro) is a kubectl-style CLI for managing containerized development environments with hierarchical organization.

---

## What is dvm?

dvm provides:

- **Hierarchical organization** - Ecosystem → Domain → App → Workspace structure for scalable development
- **App management** - Track your codebases with rich metadata and configuration
- **Workspace management** - Isolated container environments per App
- **Container orchestration** - Build, attach, detach from dev containers
- **Theme inheritance** - Themes cascade through the hierarchy for consistent styling
- **Neovim integration** - Pre-configured editor with LSP support

---

## Object Hierarchy

DevOpsMaestro organizes your development work using a four-level hierarchy:

```
Ecosystem → Domain → App → Workspace
   (org)    (context) (code)  (dev env)
```

| Object | Purpose | Example | Created With |
|--------|---------|---------|--------------|
| **Ecosystem** | Top-level platform/org grouping | `mycompany`, `aws`, `homelab` | `dvm create ecosystem` |
| **Domain** | Bounded context within ecosystem | `mycompany/backend`, `mycompany/frontend` | `dvm create domain` |
| **App** | The actual codebase/application | `mycompany/backend/api-service` | `dvm create app` |
| **Workspace** | Development environment for an App | `mycompany/backend/api-service/dev` | `dvm create workspace` |

### Why This Hierarchy?

This structure provides:

- **Scalability** - Organize hundreds of apps across teams and platforms
- **Context** - Clear bounded contexts prevent naming conflicts
- **Theme Inheritance** - Consistent styling cascades from top to bottom
- **Team Organization** - Domains map to team boundaries
- **Multi-Platform** - Separate ecosystems for different environments

### Hierarchy Examples

#### Single Platform
```
mycompany/                    # Ecosystem
├── backend/                  # Domain
│   ├── auth-service/         # App
│   │   ├── dev/              # Workspace
│   │   └── test/             # Workspace
│   └── payment-service/      # App
│       └── dev/              # Workspace
└── frontend/                 # Domain
    └── web-app/              # App
        ├── dev/              # Workspace
        └── e2e/              # Workspace
```

#### Multi-Platform
```
aws/                          # Ecosystem
├── production/               # Domain
│   └── api-gateway/          # App
└── staging/                  # Domain
    └── api-gateway/          # App

homelab/                      # Ecosystem
├── infrastructure/           # Domain
│   └── monitoring/           # App
└── experiments/              # Domain
    └── ai-project/           # App
```

### Theme Inheritance

Themes cascade through the hierarchy, allowing you to set consistent styling at any level:

```
Workspace Theme (highest priority)
    ↓ (inherits from)
App Theme
    ↓ (inherits from)
Domain Theme
    ↓ (inherits from)
Ecosystem Theme
    ↓ (inherits from)
Global Default Theme (lowest priority)
```

Example:
```bash
# Set company-wide theme
dvm set theme coolnight-corporate --ecosystem mycompany

# Override for backend team
dvm set theme coolnight-dark --domain mycompany/backend

# Override for specific app
dvm set theme coolnight-synthwave --app mycompany/backend/api-service

# Override for debugging workspace
dvm set theme coolnight-debug --workspace mycompany/backend/api-service/debug
```

---

## Core Concepts

### Apps

An **App** represents your codebase - the thing you build and run:

```bash
dvm create app my-api --from-cwd
```

- Points to a directory containing your source code
- Has language detection and build configuration
- Can have multiple workspaces (dev environments)
- Belongs to a Domain (bounded context)

### Workspaces

A **Workspace** is an isolated container environment for an App:

```bash
dvm create workspace dev
```

- Belongs to a specific App
- Has its own container image and configuration
- Mounts the App's source code directory
- Can have different tools, plugins, and themes

### Context

The **context** tracks your currently active ecosystem, domain, app, and workspace:

```bash
dvm get ctx
# Ecosystem: mycompany
# Domain:    backend  
# App:       api-service
# Workspace: dev
```

Commands operate on the active context by default.

---

## Typical Workflow

### Quick Start (Single App)

```
┌─────────────────────────────────────────────────────┐
│  1. dvm init           Initialize dvm (one-time)    │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  2. dvm create app my-app --from-cwd                │
│     Create an app pointing to your code             │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  3. dvm use app my-app                              │
│     Set the active app                              │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  4. dvm create workspace dev                        │
│     Create a workspace (container environment)      │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  5. dvm use workspace dev                           │
│     Set the active workspace                        │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  6. dvm build                                       │
│     Build the container image                       │
└─────────────────────────────────────────────────────┘
                          │
                          ▼
┌─────────────────────────────────────────────────────┐
│  7. dvm attach                                      │
│     Enter the container and start coding!           │
└─────────────────────────────────────────────────────┘
```

### Enterprise Workflow (Full Hierarchy)

For teams managing multiple apps across domains:

```bash
# 1. Initialize and set up organization structure
dvm init
dvm create ecosystem mycompany --description "My Company Platform"
dvm create domain backend -e mycompany --description "Backend services"
dvm create domain frontend -e mycompany --description "Frontend applications"

# 2. Set company-wide theme
dvm set theme coolnight-corporate --ecosystem mycompany
dvm set theme coolnight-dark --domain backend  # Override for backend team

# 3. Create apps in appropriate domains
cd ~/Developer/api-service
dvm create app api-service --from-cwd -d backend
cd ~/Developer/web-app  
dvm create app web-app --from-cwd -d frontend

# 4. Work on specific app
dvm use app api-service
dvm create workspace dev --description "Daily development"
dvm create workspace test --description "Testing environment"

# 5. Develop
dvm use workspace dev
dvm build
dvm attach  # Enter container and start coding
```

---

## kubectl-style Commands

dvm follows kubectl patterns for a familiar experience:

| Pattern | Examples |
|---------|----------|
| **get** | `dvm get ecosystems`, `dvm get apps`, `dvm get workspaces` |
| **create** | `dvm create ecosystem`, `dvm create app`, `dvm create workspace` |
| **delete** | `dvm delete ecosystem`, `dvm delete app`, `dvm delete workspace` |
| **apply** | `dvm apply -f workspace.yaml`, `dvm apply -f theme.yaml` |
| **use** | `dvm use ecosystem`, `dvm use app`, `dvm use workspace` |

### Resource Hierarchy Commands

Manage each level of the hierarchy:

| Resource | Aliases | Examples |
|----------|---------|----------|
| `ecosystems` | `eco` | `dvm get eco`, `dvm create eco production` |
| `domains` | `dom` | `dvm get dom`, `dvm create dom backend -e production` |
| `apps` | `app` | `dvm get app`, `dvm create app api --from-cwd` |
| `workspaces` | `ws` | `dvm get ws`, `dvm create ws dev` |
| `context` | `ctx` | `dvm get ctx`, `dvm use ctx --clear` |

### Cross-Hierarchy Queries

Use flags to query across hierarchy levels:

| Command | Scope | Example |
|---------|-------|---------|
| `dvm get apps` | Current domain | Apps in active domain |
| `dvm get apps -A` | All domains | Apps across all domains and ecosystems |
| `dvm get workspaces` | Current app | Workspaces for active app |
| `dvm get workspaces -A` | All apps | Workspaces across all apps |

### Declarative Configuration

Apply resources from YAML files:

```bash
# Apply from local files
dvm apply -f ecosystem.yaml
dvm apply -f app.yaml
dvm apply -f workspace.yaml

# Apply from URLs
dvm apply -f https://example.com/workspace.yaml

# Apply from GitHub
dvm apply -f github:user/repo/configs/app.yaml

# Apply multiple files
dvm apply -f workspace.yaml -f theme.yaml
```

---

## Container Platforms

dvm supports multiple container runtimes:

| Platform | Type | Notes |
|----------|------|-------|
| **OrbStack** | Docker | Recommended for macOS |
| **Docker Desktop** | Docker | Cross-platform |
| **Podman** | Docker-compatible | Rootless containers |
| **Colima** | Docker or containerd | Lightweight alternative |

Check detected platforms:

```bash
dvm get platforms
```

---

## Migration from Projects (v0.8.0+)

If you're upgrading from a version before v0.8.0, Projects have been replaced by the new hierarchy:

| Old Concept | New Concept | Notes |
|-------------|-------------|-------|
| **Project** | **App** | Your codebase with path and metadata |
| | **Domain** | Bounded context grouping (new) |
| | **Ecosystem** | Platform/organization grouping (new) |

### Automatic Migration

- Existing Projects are automatically migrated to Apps
- A default Domain and Ecosystem are created if needed
- All Workspaces remain associated with their Apps
- `dvm create project` still works but shows deprecation warnings

### Manual Migration (Recommended)

For better organization:

```bash
# Check current projects
dvm get projects

# Create proper hierarchy
dvm create ecosystem mycompany
dvm create domain backend -e mycompany

# Move apps to domains (if desired)
# Note: This requires manual recreation
dvm create app api-service --from-cwd -d backend
dvm delete project old-api-service --force
```

---

## Best Practices

### Start Simple, Scale Up

```bash
# Minimal - just apps
dvm create app my-api --from-cwd

# Scale to teams - add domains  
dvm create domain backend
dvm create app api-service --from-cwd -d backend

# Scale to enterprise - add ecosystems
dvm create ecosystem production
dvm create domain backend -e production
```

### Use Descriptive Names

```bash
# Good
dvm create ecosystem customer-platform
dvm create domain payment-services
dvm create app stripe-integration

# Less helpful
dvm create ecosystem prod
dvm create domain stuff
dvm create app app1
```

### Leverage Theme Inheritance

```bash
# Set company branding at top level
dvm set theme company-brand --ecosystem customer-platform

# Team-specific overrides
dvm set theme dark-mode --domain backend-services

# App-specific needs
dvm set theme high-contrast --app accessibility-service
```

---

## Next Steps

- [Apps](apps.md) - Managing apps (your codebases)
- [Workspaces](workspaces.md) - Managing development environments
- [Building & Attaching](build-attach.md) - Container lifecycle
- [Commands Reference](commands.md) - Complete command list
