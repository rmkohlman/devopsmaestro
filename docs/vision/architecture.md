# DevOpsMaestro Architecture Vision

> **Purpose:** Complete architectural vision for DevOpsMaestro's evolution from a dev environment tool into a full local DevOps platform.
>
> **Status:** Living Document - defines the target state we're building toward.
>
> **Audience:** AI assistants, contributors, and future-you trying to understand the complete vision.
>
> **Last Updated:** v2.0 - Refined hierarchy with Domain and App objects

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [Philosophy & Principles](#philosophy--principles)
3. [Complete Object Model](#complete-object-model)
4. [Object Relationships](#object-relationships)
5. [Scoping Rules](#scoping-rules)
6. [Core Workflows](#core-workflows)
7. [YAML Specifications](#yaml-specifications)
8. [CLI Command Structure](#cli-command-structure)
9. [Implementation Phases](#implementation-phases)
10. [Migration Path](#migration-path)

---

## Executive Summary

### What DevOpsMaestro Is Becoming

DevOpsMaestro is evolving from a **containerized development environment manager** into a **complete local DevOps platform**. The vision is to provide developers with:

1. **Development Environments** (Workspaces) - Where you write code (App in dev mode)
2. **Live Environments** (live mode) - Where your App runs in production-like conditions
3. **CI/CD Pipelines** (Tasks, Workflows, Pipelines) - How code moves from dev to live
4. **Operator** - Kubernetes Operator (CRD-based) that orchestrates everything

### The Core Hierarchy

```
Ecosystem → Domain → App → Workspace (dev mode)
                      ↓
                  (live mode - managed by Operator)
```

| Object | Purpose | Analogy |
|--------|---------|---------|
| **Ecosystem** | Top-level platform grouping | A product area or company |
| **Domain** | Bounded context within ecosystem | A team's area of responsibility |
| **App** | The codebase/application (lives for years) | The actual code you build |
| **Workspace** | Dev environment for an App | Where you write code |
| **live mode** | App running in production-like env | Managed by Operator |

### The Core Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DEVELOPMENT                                     │
│  ┌──────────────────┐                                                       │
│  │    WORKSPACE     │  ← You code here (nvim, tools, path to source)        │
│  │    auth-api      │                                                       │
│  │    (dev mode)    │                                                       │
│  └────────┬─────────┘                                                       │
│           │ git commit                                                      │
│           ▼                                                                 │
│  ┌──────────────────┐                                                       │
│  │      JOB         │  ← Triggered by commit, runs Pipeline                 │
│  │  on-commit-ci    │                                                       │
│  └────────┬─────────┘                                                       │
│           │                                                                 │
│           ▼                                                                 │
│  ┌──────────────────┐                                                       │
│  │    PIPELINE      │  ← Build → Test → Deploy stages                       │
│  │    ci-pipeline   │                                                       │
│  └────────┬─────────┘                                                       │
│           │ deploy on success                                               │
│           ▼                                                                 │
└─────────────────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                              LIVE ENVIRONMENT                                │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐          │
│  │       APP        │  │       APP        │  │       APP        │          │
│  │    auth-api      │  │    user-api      │  │   billing-api    │          │
│  │   (live mode)    │  │   (live mode)    │  │   (live mode)    │          │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘          │
│           │                    │                      │                     │
│           └────────────────────┼──────────────────────┘                     │
│                                ▼                                            │
│                    ┌──────────────────┐                                     │
│                    │   Global Redis   │  ← Shared Dependency                │
│                    │   (messaging)    │                                     │
│                    └──────────────────┘                                     │
│                                                                             │
│  + Chaos testing, integration tests, observability running continuously     │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Why This Matters

- **App** = The codebase that exists for years (the thing you build)
- **Workspace** = App in dev mode - where you develop (nvim, tools, source code)
- **live mode** = App running in production-like environment (Operator manages)
- Same App, different runtime modes
- Multiple Apps across Ecosystems can interact via shared Dependencies
- Enables **full microservices testing locally** on your laptop

### Two Operating Modes

| Mode | Tool | Requirements | Use Case |
|------|------|--------------|----------|
| **dvm alone** | `dvm` CLI | Docker only | Basic workspace management, single dev |
| **dvm + Operator** | `dvm` + k8s Operator | Docker + k8s (Colima/OrbStack) | Full DevOps: live mode, CI/CD, dependencies |

---

## Philosophy & Principles

### 1. Generic Over Specific

**Reuse generic objects rather than creating special-purpose objects.**

Why have a `Backup` object when it's really a `Task`? Why have a `Restore` object when it's also a `Task`?

```yaml
# BAD: Special-purpose objects
kind: Backup
spec:
  database: postgres
  schedule: "0 2 * * *"

# GOOD: Generic Task with specific action
kind: Task
metadata:
  name: backup-postgres
spec:
  action: data.backup
  target: postgres
  schedule: "0 2 * * *"
```

### 2. Composition Over Inheritance

Build complex behaviors by composing simple objects.

```yaml
# Workflow composes Tasks
kind: Workflow
spec:
  tasks:
    - ref: lint-code
    - ref: run-tests
      dependsOn: [lint-code]
    - ref: build-binary
      dependsOn: [run-tests]

# Pipeline composes Workflows
kind: Pipeline
spec:
  stages:
    - name: build
      workflow: ci-workflow
    - name: deploy
      workflow: deploy-workflow
```

### 3. Template Pattern

Reusable defaults for any object type. Templates are not a special object - they're a pattern applied to any resource.

```yaml
# Define a template
kind: Template
metadata:
  name: go-api-workspace
spec:
  targetKind: Workspace
  defaults:
    language: go
    languageVersion: "1.23"
    nvim:
      plugins: [telescope, treesitter, lsp-zero, go-nvim]
      theme: tokyonight

# Use the template
kind: Workspace
metadata:
  name: auth-api
spec:
  template: go-api-workspace  # Inherits all defaults
  # Override specific values
  languageVersion: "1.24"
```

### 4. CRD-like Extensibility

Users can define custom resource types, just like Kubernetes CRDs.

```yaml
kind: ResourceDefinition
metadata:
  name: featureflag
spec:
  group: mycompany.io
  names:
    kind: FeatureFlag
    plural: featureflags
  schema:
    properties:
      enabled: { type: boolean }
      rollout: { type: integer }
```

### 5. Declarative Everything

Every piece of configuration is YAML. Every operation can be expressed as `dvm apply -f`.

### 6. kubectl Patterns

Familiar commands for anyone who knows Kubernetes:

```bash
dvm get workspaces
dvm create project my-api
dvm apply -f workspace.yaml
dvm delete service auth-api
```

---

## Complete Object Model

### Object Categories

| Category | Objects | Purpose |
|----------|---------|---------|
| **Hierarchy** | Ecosystem, Domain, App, Workspace, Context | Organizational containers |
| **Execution** | Task, Workflow, Pipeline, Orchestration, Job | CI/CD and automation |
| **Data** | DataRecord, DataStore, DataLake | Test data management |
| **Infrastructure** | Dependency, Volume, Action, Runtime | Services and storage |
| **Reusability** | Template, ResourceDefinition | Templates and extensibility |
| **Dev Tools** | NvimPlugin, NvimTheme, TerminalPrompt, TerminalPlugin, TerminalTheme | Editor and shell config |
| **Operations** | Operator | Kubernetes Operator for live mode |

---

### Hierarchy Objects

#### Ecosystem

Top-level grouping of related domains. Think of it as a "platform" or "product area."

```yaml
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: customer-platform
  description: All customer-facing services
spec:
  # Shared dependencies available to all domains/apps
  dependencies:
    - name: platform-db        # Instance name
      ref: postgres            # References global Dependency definition
    - name: platform-cache
      ref: redis
  
  # Shared data for integration testing
  dataLakes:
    - ref: integration-test-data
```

#### Domain

Groups related Apps within an ecosystem. Think of it as a "bounded context" or "team area."
**Note:** This was previously called "Project" but renamed to better reflect its long-lived nature.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: auth-domain
  ecosystem: customer-platform
  description: Authentication and authorization services
spec:
  # Domain-level dependencies (in addition to ecosystem)
  dependencies:
    - name: auth-secrets
      ref: vault
  
  # Domain-level test data
  dataStores:
    - ref: auth-test-data
  
  # Default settings for apps in this domain
  defaults:
    language: go
    languageVersion: "1.23"
```

#### App

The codebase/application - the core object that exists for years. Has a `path` to source code.
Apps can run in **dev mode** (Workspace) or **live mode** (managed by Operator).

```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: auth-api
  domain: auth-domain
  description: Authentication API service
spec:
  # Path to source code on host machine
  path: ~/Developer/auth-api
  
  language: go
  languageVersion: "1.23"
  
  # Resource limits for live mode
  resources:
    memory: 512Mi
    cpu: 500m
  
  # Port mappings
  ports:
    - containerPort: 8080
      hostPort: 8080
  
  # Health check (used in live mode)
  healthCheck:
    path: /health
    port: 8080
    interval: 10s
  
  # Dependencies this app needs
  dependencies:
    - ref: platform-db       # Uses ecosystem's instance
```

#### Workspace

Development environment for an App. This is where you write code.
A Workspace is essentially an App running in **dev mode**.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: auth-api
  description: Main development workspace
spec:
  # Mount points into container
  mounts:
    - source: ~/Developer/auth-api
      target: /workspace
  
  # Neovim configuration
  nvim:
    plugins:
      - telescope
      - treesitter
      - lsp-zero
    theme: tokyonight
  
  # Terminal configuration
  terminal:
    prompt: starship
    plugins:
      - zsh-autosuggestions
      - zsh-syntax-highlighting
```

#### Context

Active selection state - which ecosystem/domain/app/workspace is currently selected.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Context
metadata:
  name: current
spec:
  ecosystem: customer-platform
  domain: auth-domain
  app: auth-api
  workspace: dev
```

---

### Execution Objects

#### Task

Single discrete action. The atomic unit of work.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Task
metadata:
  name: run-tests
  domain: auth-domain
spec:
  # Action to perform (extensible via Action objects)
  action: exec
  
  # Command to run
  command: go test ./... -v -race
  workdir: /workspace
  
  # Timeout
  timeout: 5m
  
  # Environment variables
  env:
    - name: GO_TEST_FLAGS
      value: "-v -race"
  
  # What to do on completion
  onSuccess:
    action: notify.slack
    channel: "#builds"
  onFailure:
    action: notify.alert
    severity: high
```

#### Workflow

Series of Tasks with dependencies. Tasks can run in parallel when no dependencies exist.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workflow
metadata:
  name: ci-workflow
  domain: auth-domain
spec:
  tasks:
    # These run in parallel (no dependencies)
    - name: lint
      taskRef: run-linter
    - name: security-scan
      taskRef: run-security-scan
    
    # This waits for both above to complete
    - name: test
      taskRef: run-tests
      dependsOn: [lint, security-scan]
    
    # This waits for test
    - name: build
      taskRef: build-binary
      dependsOn: [test]
  
  # Workflow-level callbacks
  onComplete:
    action: notify.team
```

#### Pipeline

Full CI/CD cycle with stages. Each stage contains a Workflow.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Pipeline
metadata:
  name: ci-pipeline
  domain: auth-domain
spec:
  stages:
    - name: build
      workflowRef: ci-workflow
    
    - name: deploy-to-service
      workflowRef: deploy-workflow
      dependsOn: [build]
    
    - name: integration-tests
      workflowRef: integration-test-workflow
      dependsOn: [deploy-to-service]
    
    - name: chaos-tests
      workflowRef: chaos-test-workflow
      dependsOn: [integration-tests]
      # Optional - only runs if explicitly requested
      manual: false
```

#### Orchestration

Multi-pipeline coordination. Runs multiple pipelines across projects/ecosystems.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Orchestration
metadata:
  name: full-platform-test
  ecosystem: customer-platform
spec:
  # Run these pipelines (can be across different projects)
  pipelines:
    - name: auth-ci
      pipelineRef: auth-domain/ci-pipeline
    - name: user-ci
      pipelineRef: user-domain/ci-pipeline
      dependsOn: [auth-ci]  # User depends on auth
    - name: billing-ci
      pipelineRef: billing-domain/ci-pipeline
      dependsOn: [auth-ci]  # Billing also depends on auth
  
  # After all pipelines complete
  onComplete:
    action: notify.slack
    message: "Full platform test complete"
```

#### Job

Trigger definition - when and how to run CI/CD objects. The "cron job" or "webhook" equivalent.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Job
metadata:
  name: on-commit-ci
  domain: auth-domain
spec:
  # What triggers this job
  trigger:
    # Watch the app path for git commits
    watch: app.path
    event: commit
    # Or schedule-based
    # schedule: "0 * * * *"  # Every hour
    # Or manual
    # manual: true
  
  # What to run when triggered
  runs: ci-pipeline
  
  # What to do on success/failure
  onSuccess:
    action: git.push
    remote: origin
    branch: main
  onFailure:
    action: notify.alert
    severity: critical
```

---

### Data Objects

Data objects are **composable** - they can be referenced by Tasks and shared across scopes.

#### DataRecord

Atomic data unit - a single row or entity. Can be reused across multiple Tasks.

```yaml
apiVersion: devopsmaestro.io/v1
kind: DataRecord
metadata:
  name: test-user-alice
  domain: auth-domain
spec:
  # Which DataStore this record belongs to
  dataStoreRef: auth-test-data
  
  # The actual data
  data:
    id: "usr_001"
    email: "alice@example.com"
    name: "Alice Smith"
    role: "admin"
    created_at: "2024-01-01T00:00:00Z"
```

#### DataStore

Schema + collection of DataRecords. Think of it as a "table" or "collection."

```yaml
apiVersion: devopsmaestro.io/v1
kind: DataStore
metadata:
  name: auth-test-data
  domain: auth-domain
spec:
  # Target dependency (where data will be inserted)
  target:
    dependencyRef: platform-db
    database: auth_db
    table: users
  
  # Schema definition
  schema:
    fields:
      - name: id
        type: string
        primary: true
      - name: email
        type: string
        unique: true
      - name: name
        type: string
      - name: role
        type: string
        enum: [admin, user, guest]
      - name: created_at
        type: timestamp
  
  # Records in this store (or reference external DataRecords)
  records:
    - ref: test-user-alice
    - ref: test-user-bob
    - inline:
        id: "usr_003"
        email: "charlie@example.com"
        name: "Charlie Brown"
        role: "user"
```

#### DataLake

Collection of DataStores. Used for ecosystem-wide test data.

```yaml
apiVersion: devopsmaestro.io/v1
kind: DataLake
metadata:
  name: integration-test-data
  ecosystem: customer-platform
spec:
  # DataStores in this lake
  dataStores:
    - ref: auth-domain/auth-test-data
    - ref: user-domain/user-test-data
    - ref: billing-domain/billing-test-data
  
  # Seed order matters for foreign key relationships
  seedOrder:
    - auth-test-data      # Users first
    - user-test-data      # User profiles second
    - billing-test-data   # Billing records last
  
  # Actions to run before/after seeding
  beforeSeed:
    action: data.truncate
    target: all
  afterSeed:
    action: data.verify
    target: all
```

---

### Infrastructure Objects

#### Dependency

Service definition. **All Dependencies are defined globally** and instantiated by reference.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Dependency
metadata:
  name: postgres
  # No scope - all dependencies are global definitions
spec:
  # Container image
  image: postgres:16-alpine
  
  # Default ports
  ports:
    - containerPort: 5432
      protocol: TCP
  
  # Default environment variables
  env:
    - name: POSTGRES_USER
      value: devops
    - name: POSTGRES_PASSWORD
      value: devops
    - name: POSTGRES_DB
      value: devops
  
  # Health check
  healthCheck:
    command: ["pg_isready", "-U", "devops"]
    interval: 5s
    timeout: 3s
    retries: 5
  
  # Volume for data persistence
  volumes:
    - name: data
      mountPath: /var/lib/postgresql/data
```

**How Dependencies Are Instantiated:**

```yaml
# Ecosystem creates an instance
kind: Ecosystem
metadata:
  name: customer-platform
spec:
  dependencies:
    - name: platform-db      # Instance name (unique within ecosystem)
      ref: postgres          # References global Dependency definition
      # Override defaults if needed
      env:
        - name: POSTGRES_DB
          value: platform_db

# App uses the SAME instance (same name)
kind: App
metadata:
  name: auth-api
spec:
  dependencies:
    - ref: platform-db       # Uses ecosystem's instance

# App creates its OWN instance (different name)
kind: App
metadata:
  name: isolated-tests
spec:
  dependencies:
    - name: test-db          # Different name = new instance
      ref: postgres
      env:
        - name: POSTGRES_DB
          value: test_db
```

#### Volume

Persistent storage that can be attached to Dependencies or Workspaces.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Volume
metadata:
  name: postgres-data
  ecosystem: customer-platform
spec:
  # Size
  size: 10Gi
  
  # Storage class (maps to local paths)
  storageClass: standard  # or 'fast' for SSD
  
  # Where to store on host
  hostPath: ~/.dvm/volumes/customer-platform/postgres-data
  
  # Backup settings
  backup:
    enabled: true
    schedule: "0 2 * * *"
    retention: 7d
```

#### Action

Extensible action handler. Actions are the verbs that Tasks use.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Action
metadata:
  name: data.insert
spec:
  description: Insert data into a data store
  
  # Different handlers for different dependency types
  handlers:
    postgres:
      type: sql
      template: |
        INSERT INTO {{ .table }} ({{ .columns }})
        VALUES ({{ .values }})
        ON CONFLICT ({{ .primaryKey }}) DO UPDATE SET {{ .updates }};
    
    redis:
      type: command
      template: |
        SET {{ .key }} '{{ .value | json }}'
    
    mongodb:
      type: command
      template: |
        db.{{ .collection }}.insertOne({{ .document | json }})
  
  # Input schema
  inputs:
    - name: target
      type: string
      required: true
    - name: data
      type: object
      required: true
```

**Built-in Actions:**

| Action | Description |
|--------|-------------|
| `exec` | Execute a command |
| `data.insert` | Insert data into a store |
| `data.delete` | Delete data from a store |
| `data.truncate` | Truncate a table/collection |
| `data.backup` | Backup a dependency |
| `data.restore` | Restore a dependency |
| `git.push` | Push to git remote |
| `git.pull` | Pull from git remote |
| `notify.slack` | Send Slack notification |
| `notify.alert` | Send alert notification |
| `service.deploy` | Deploy a service |
| `service.restart` | Restart a service |
| `chaos.kill` | Kill a random service instance |
| `chaos.latency` | Inject network latency |

#### Runtime

Container platform configuration.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Runtime
metadata:
  name: default
spec:
  # Platform type
  platform: docker  # or orbstack, colima, podman
  
  # Socket path (auto-detected if not specified)
  socket: /var/run/docker.sock
  
  # Resource limits for the runtime
  resources:
    maxMemory: 8Gi
    maxCPU: 4
  
  # Network settings
  network:
    name: dvm-network
    driver: bridge
```

---

### Reusability Objects

#### Template

Reusable defaults for any object type.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Template
metadata:
  name: go-api-workspace
spec:
  # What kind of object this template is for
  targetKind: Workspace
  
  # Default values
  defaults:
    language: go
    languageVersion: "1.23"
    
    nvim:
      plugins:
        - telescope
        - treesitter
        - lsp-zero
        - go-nvim
      theme: tokyonight
    
    terminal:
      prompt: starship
      plugins:
        - zsh-autosuggestions
        - git
    
    # Standard tasks for Go projects
    tasks:
      - ref: go-lint
      - ref: go-test
      - ref: go-build
```

**Using Templates:**

```yaml
kind: App
metadata:
  name: auth-api
spec:
  template: go-api-app  # Apply template defaults
  
  # Override specific values
  languageVersion: "1.24"
```

#### ResourceDefinition

Define custom resource types (CRD-like extensibility).

```yaml
apiVersion: devopsmaestro.io/v1
kind: ResourceDefinition
metadata:
  name: featureflags.mycompany.io
spec:
  # API group
  group: mycompany.io
  
  # Names
  names:
    kind: FeatureFlag
    plural: featureflags
    singular: featureflag
    shortNames: [ff]
  
  # Scope
  scope: Project  # or Ecosystem, Workspace, Global
  
  # Schema
  schema:
    type: object
    required: [spec]
    properties:
      spec:
        type: object
        required: [enabled]
        properties:
          enabled:
            type: boolean
            description: Whether the feature is enabled
          rolloutPercentage:
            type: integer
            minimum: 0
            maximum: 100
            default: 100
          allowedUsers:
            type: array
            items:
              type: string
```

**Using Custom Resources:**

```yaml
apiVersion: mycompany.io/v1
kind: FeatureFlag
metadata:
  name: new-dashboard
  project: frontend
spec:
  enabled: true
  rolloutPercentage: 25
  allowedUsers:
    - dev-team
    - beta-users
```

---

### Dev Tools Objects

#### NvimPlugin

Neovim plugin configuration.

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  description: Fuzzy finder for Neovim
  category: navigation
spec:
  repo: nvim-telescope/telescope.nvim
  branch: master
  lazy: true
  event: VimEnter
  
  dependencies:
    - plenary
    - nvim-web-devicons
  
  keys:
    - key: "<leader>ff"
      action: "Telescope find_files"
      desc: "Find files"
    - key: "<leader>fg"
      action: "Telescope live_grep"
      desc: "Live grep"
  
  config: |
    require("telescope").setup({
      defaults = {
        file_ignore_patterns = { "node_modules", ".git" },
      },
    })
```

#### NvimTheme

Neovim colorscheme with palette.

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: tokyonight-custom
  description: Tokyo Night colorscheme
spec:
  colorscheme: tokyonight
  background: dark
  repo: folke/tokyonight.nvim
  
  config: |
    require("tokyonight").setup({
      style = "night",
      transparent = false,
    })
  
  palette:
    primary: "#7aa2f7"
    secondary: "#bb9af7"
    accent: "#7dcfff"
    bg: "#1a1b26"
    fg: "#c0caf5"
    error: "#f7768e"
    warning: "#e0af68"
    info: "#7dcfff"
    hint: "#1abc9c"
```

#### TerminalPrompt

Shell prompt configuration.

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPrompt
metadata:
  name: starship-minimal
spec:
  type: starship
  
  config: |
    format = "$directory$git_branch$git_status$character"
    
    [directory]
    style = "blue bold"
    
    [git_branch]
    style = "purple"
    
    [character]
    success_symbol = "[❯](green)"
    error_symbol = "[❯](red)"
```

#### TerminalPlugin

Shell plugin configuration.

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalPlugin
metadata:
  name: zsh-autosuggestions
spec:
  repo: zsh-users/zsh-autosuggestions
  shell: zsh
  
  # How to source the plugin
  source: zsh-autosuggestions.zsh
  
  # Configuration
  config:
    ZSH_AUTOSUGGEST_STRATEGY: (history completion)
    ZSH_AUTOSUGGEST_HIGHLIGHT_STYLE: "fg=#666666"
```

#### TerminalTheme

Shell colors/theme.

```yaml
apiVersion: devopsmaestro.io/v1
kind: TerminalTheme
metadata:
  name: tokyonight-terminal
spec:
  # Palette (can reference NvimTheme palette)
  paletteRef: tokyonight-custom
  
  # Or define inline
  colors:
    background: "#1a1b26"
    foreground: "#c0caf5"
    cursor: "#c0caf5"
    
    # ANSI colors
    black: "#15161e"
    red: "#f7768e"
    green: "#9ece6a"
    yellow: "#e0af68"
    blue: "#7aa2f7"
    magenta: "#bb9af7"
    cyan: "#7dcfff"
    white: "#a9b1d6"
```

---

### Operations Objects

#### Operator

The Kubernetes Operator (CRD-based). Manages Apps in live mode, orchestrates CI/CD.
**Note:** Requires Kubernetes (Colima, OrbStack k8s, minikube). dvm works without Operator for basic workspace usage.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Operator
metadata:
  name: default
spec:
  # What to watch
  watches:
    - type: app
      events: [commit]
      # Run this job when event occurs
      jobRef: on-commit-ci
    
    - type: app
      mode: live
      events: [unhealthy]
      jobRef: app-recovery
  
  # Global settings
  settings:
    # How often to check for events
    pollInterval: 5s
    
    # Max concurrent jobs
    maxConcurrentJobs: 5
    
    # Auto-push on successful pipeline
    autoPush:
      enabled: true
      remote: origin
      branch: main
  
  # Logging
  logging:
    level: info
    output: ~/.dvm/logs/operator.log
```

---

## Object Relationships

### Visual Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              GLOBAL SCOPE                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│  Dependency (definitions): postgres, redis, kafka, mongodb, vault           │
│  Action (definitions): exec, data.insert, data.delete, git.push, notify.*   │
│  Template: go-api-app, python-service, node-frontend                        │
│  ResourceDefinition: custom CRDs                                            │
│  Runtime: docker, orbstack, colima                                          │
│  NvimPlugin, NvimTheme (library)                                            │
│  TerminalPrompt, TerminalPlugin, TerminalTheme (library)                    │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────────────┐
│                         ECOSYSTEM: customer-platform                         │
├─────────────────────────────────────────────────────────────────────────────┤
│  Dependencies (instances): platform-db, platform-cache, platform-queue      │
│  DataLake: integration-test-data                                            │
│  Volume: postgres-data, redis-data                                          │
│  Operator: default                                                          │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    DOMAIN: auth-domain                               │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  Dependencies (instances): auth-secrets                              │   │
│  │  DataStore: auth-test-data                                          │   │
│  │  Task, Workflow, Pipeline, Job definitions                          │   │
│  │                                                                      │   │
│  │  ┌─────────────────────────────────────────────────────────────┐    │   │
│  │  │                    APP: auth-api                             │    │   │
│  │  │  path: ~/Developer/auth-api                                  │    │   │
│  │  ├─────────────────────────────────────────────────────────────┤    │   │
│  │  │  ┌─────────────────┐    ┌─────────────────┐                 │    │   │
│  │  │  │ WORKSPACE: dev  │    │  (live mode)    │                 │    │   │
│  │  │  │ (dev mode)      │───▶│ managed by      │                 │    │   │
│  │  │  │ where you code  │    │ Operator        │                 │    │   │
│  │  │  └─────────────────┘    └─────────────────┘                 │    │   │
│  │  └─────────────────────────────────────────────────────────────┘    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    DOMAIN: user-domain                               │   │
│  │  (similar structure)                                                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Relationship Table

| Object | Contains/References | Contained By |
|--------|---------------------|--------------|
| **Ecosystem** | Domains, Dependency instances, DataLakes, Volumes, Operator | - |
| **Domain** | Apps, Tasks, Workflows, Pipelines, Jobs, DataStores | Ecosystem |
| **App** | Workspaces, path to source code, Dependencies | Domain |
| **Workspace** | NvimPlugins, NvimTheme, TerminalPrompt, TerminalPlugins | App |
| **Task** | Actions | Domain |
| **Workflow** | Tasks | Domain |
| **Pipeline** | Workflows | Domain |
| **Orchestration** | Pipelines | Ecosystem |
| **Job** | Pipeline/Workflow/Task reference, Trigger | Domain |
| **DataLake** | DataStores | Ecosystem |
| **DataStore** | DataRecords | Domain |
| **DataRecord** | Data | Domain |
| **Dependency** | (global definition) | Global |
| **Template** | Default values for any Kind | Global |

---

## Scoping Rules

### Scope Hierarchy

```
Global
  └── Ecosystem
        └── Domain
              └── App
                    └── Workspace
```

### Object Scope Summary

| Scope | Objects |
|-------|---------|
| **Global** | Dependency (definitions), Action, Template, ResourceDefinition, Runtime, NvimPlugin*, NvimTheme*, TerminalPrompt*, TerminalPlugin*, TerminalTheme* |
| **Ecosystem** | Ecosystem, Dependency (instances), DataLake, Volume, Operator, Orchestration |
| **Domain** | Domain, Task, Workflow, Pipeline, Job, DataStore, DataRecord |
| **App** | App (with path), Dependencies |
| **Workspace** | Workspace (inherits from App, references global dev tools) |

*Dev tools (Nvim*, Terminal*) are defined globally but can be overridden at Workspace level.

### Dependency Instance Sharing

**Key Rule:** Same instance name at the same scope level = shared instance.

```yaml
# Ecosystem creates instance
kind: Ecosystem
spec:
  dependencies:
    - name: platform-db      # Creates instance "platform-db"
      ref: postgres

# App A uses it
kind: App
metadata:
  name: auth-api
spec:
  dependencies:
    - ref: platform-db       # Uses ecosystem's "platform-db"

# App B also uses it
kind: App
metadata:
  name: user-api
spec:
  dependencies:
    - ref: platform-db       # Same instance as auth-api

# App C creates its own
kind: App
metadata:
  name: isolated-tests
spec:
  dependencies:
    - name: test-db          # NEW instance (different name)
      ref: postgres
```

---

## Core Workflows

### Workflow 1: Developer Coding Session

```
1. Developer runs: dvm attach auth-api
   └── Starts Workspace container for auth-api App
   └── Mounts ~/Developer/auth-api to /workspace
   └── Starts shared dependencies (platform-db)
   └── Opens shell with nvim configured

2. Developer codes in Workspace
   └── Full nvim setup with LSP, plugins, theme
   └── Terminal with starship prompt, zsh plugins
   └── Access to platform-db at localhost:5432

3. Developer commits: git commit -m "Add feature"
   └── Operator detects commit (if running)
   └── Triggers Job: on-commit-ci
   └── Job runs Pipeline: ci-pipeline
```

### Workflow 2: CI/CD Pipeline

```
1. Job triggers Pipeline: ci-pipeline
   └── Stage 1: build (ci-workflow)
       ├── Task: lint-code (parallel)
       ├── Task: security-scan (parallel)
       ├── Task: run-tests (depends on lint, scan)
       └── Task: build-binary (depends on tests)
   
   └── Stage 2: deploy-to-live
       └── Task: deploy auth-api to live mode
   
   └── Stage 3: integration-tests
       └── Task: run integration tests against live App
   
   └── Stage 4: chaos-tests (optional)
       └── Task: inject failures, verify resilience

2. On Pipeline success:
   └── Job.onSuccess: git.push to remote
   └── Notify: Slack message to #builds

3. On Pipeline failure:
   └── Job.onFailure: notify.alert
   └── App NOT deployed to live (stays at previous version)
```

### Workflow 3: Multi-App Interaction

```
1. Multiple Apps running in live mode:
   └── customer-platform ecosystem
       ├── auth-api App (live mode)
       ├── user-api App (live mode)
       └── billing-api App (live mode)
   
   └── internal-tools ecosystem
       ├── admin-api App (live mode)
       └── reporting-api App (live mode)

2. All Apps share:
   └── Global Redis (messaging)
   └── Global Kafka (events)

3. Developer can:
   └── Watch all Apps interact
   └── Run chaos tests across Apps
   └── Debug issues by checking logs
   └── Fix in Workspace, commit, auto-deploy
```

### Workflow 4: Test Data Seeding

```
1. Before integration tests:
   └── DataLake: integration-test-data
       └── Seed order: auth → user → billing
       └── Action: data.truncate all tables
       └── Action: data.insert from DataStores

2. During tests:
   └── Tests use seeded data
   └── DataRecords are reusable across Tasks

3. After tests:
   └── Action: data.verify (check data integrity)
   └── Optional: data.truncate (cleanup)
```

---

## CLI Command Structure

### Universal Pattern

```bash
dvm <verb> <object-type> [name] [flags]
```

### Verbs

| Verb | Purpose | Example |
|------|---------|---------|
| `create` | Create a new object | `dvm create project my-api` |
| `get` | List or view objects | `dvm get workspaces` |
| `apply` | Apply YAML configuration | `dvm apply -f workspace.yaml` |
| `delete` | Delete an object | `dvm delete workspace auth-api` |
| `use` | Set context | `dvm use project my-api` |
| `build` | Build workspace image | `dvm build -w auth-api` |
| `attach` | Attach to workspace | `dvm attach auth-api` |
| `detach` | Detach from workspace | `dvm detach` |
| `run` | Run a Task/Workflow/Pipeline | `dvm run task lint-code` |
| `logs` | View logs | `dvm logs service auth-api` |
| `status` | View status | `dvm status pipeline ci-pipeline` |

### Object Types by Category

#### Hierarchy
```bash
dvm get ecosystems
dvm create ecosystem customer-platform
dvm get domains [-e ecosystem]
dvm create domain auth-domain -e customer-platform
dvm get apps [-d domain]
dvm create app auth-api -d auth-domain --path ~/Developer/auth-api
dvm get workspaces [-a app]
dvm get context
dvm use ecosystem customer-platform
dvm use domain auth-domain
dvm use app auth-api
dvm use workspace dev
```

#### Execution
```bash
dvm get tasks [-d domain]
dvm run task lint-code
dvm get workflows [-d domain]
dvm run workflow ci-workflow
dvm get pipelines [-d domain]
dvm run pipeline ci-pipeline
dvm get orchestrations [-e ecosystem]
dvm run orchestration full-platform-test
dvm get jobs [-d domain]
dvm status job on-commit-ci
```

#### Data
```bash
dvm get datarecords [-d domain]
dvm get datastores [-d domain]
dvm get datalakes [-e ecosystem]
dvm apply -f test-data.yaml
```

#### Infrastructure
```bash
dvm get dependencies           # Global definitions
dvm get dependencies -e eco    # Instances in ecosystem
dvm get volumes [-e ecosystem]
dvm get actions                # Global actions
dvm get runtimes
```

#### Dev Tools
```bash
dvm get nvim plugins
dvm get nvim themes
dvm apply -f plugin.yaml
nvp plugin list                # Standalone nvp
nvp theme list
```

#### Reusability
```bash
dvm get templates
dvm apply -f template.yaml
dvm get resourcedefinitions
dvm get crds                   # Alias
```

#### Operations
```bash
dvm operator start
dvm operator stop
dvm operator status
dvm operator logs
```

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--ecosystem` | `-e` | Ecosystem context |
| `--domain` | `-d` | Domain context |
| `--app` | `-a` | App context |
| `--workspace` | `-w` | Workspace context |
| `--output` | `-o` | Output format (table, yaml, json) |
| `--all` | `-A` | All (across contexts) |
| `--file` | `-f` | YAML file to apply |
| `--verbose` | `-v` | Verbose output |

---

## Implementation Phases

### Phase 1: Hierarchy Foundation (v0.8.x)

**Goal:** Add Ecosystem, Domain, and App objects.

| Task | Description | Priority |
|------|-------------|----------|
| Add Ecosystem object | New table, CRUD operations | High |
| Add Domain object | Replaces Project concept | High |
| Add App object | Core object with `path` | High |
| Update Workspace | Now belongs to App | High |
| Update Context | Add ecosystem, domain, app | Medium |
| Update CLI | New commands for hierarchy | Medium |

### Phase 2: Migration (v0.9.x)

**Goal:** Migrate existing Projects to Apps, deprecate Project.

| Task | Description | Priority |
|------|-------------|----------|
| Auto-migrate Projects → Apps | One-time migration | High |
| Create default Ecosystem/Domain | For migrated data | High |
| Add deprecation warnings | On old project commands | Medium |
| Remove Project code | After migration complete | Low |

### Phase 3: Terminal Tools (v0.10.x)

**Goal:** Add TerminalPrompt, TerminalPlugin, TerminalTheme.

| Task | Description | Priority |
|------|-------------|----------|
| Add TerminalPrompt object | Starship, oh-my-posh configs | Medium |
| Add TerminalPlugin object | zsh/bash plugins | Medium |
| Add TerminalTheme object | Terminal color schemes | Medium |

### Phase 4: Templates (v0.11.x)

**Goal:** Template system for reusable configurations.

| Task | Description | Priority |
|------|-------------|----------|
| Add Template object | Reusable defaults for any Kind | Medium |
| Template inheritance | Apps/Workspaces use templates | Medium |

### Phase 5: Execution (v1.0.x - v1.1.x)

**Goal:** Local CI/CD with Task, Workflow, Pipeline, Job.

| Task | Description | Priority |
|------|-------------|----------|
| Add Task object | Single action execution | High |
| Add Action system | Built-in actions (exec, etc.) | High |
| Add Workflow object | Task composition | High |
| Add Pipeline object | Stage-based CI/CD | High |
| Add Job object | Trigger definitions | Medium |

### Phase 6: Operator (v1.2.x+)

**Goal:** Kubernetes Operator for live mode and automation.

| Task | Description | Priority |
|------|-------------|----------|
| Operator CRDs | Define CRDs for k8s | High |
| Live mode deployment | Deploy Apps to live | High |
| Watch for commits | Trigger Jobs on commit | Medium |
| Auto-push on success | Git push after pipeline | Medium |
| Chaos actions | chaos.kill, chaos.latency | Low |

---

## Migration Path

### From Current State (v0.7.x) to Phase 1 (v0.8.x)

**Current State:**
- Project has `path`
- Workspace belongs to Project
- No Ecosystem, Domain, App, etc.

**Migration Steps:**

1. **Create default Ecosystem:**
   ```sql
   INSERT INTO ecosystems (name, description) 
   VALUES ('default', 'Default ecosystem for migrated data');
   ```

2. **Create default Domain:**
   ```sql
   INSERT INTO domains (name, ecosystem_id, description)
   VALUES ('default', 1, 'Default domain for migrated projects');
   ```

3. **Migrate Projects to Apps:**
   ```sql
   -- For each Project, create an App
   INSERT INTO apps (name, domain_id, path, description)
   SELECT name, 1, path, description FROM projects;
   ```

4. **Update Workspaces to reference Apps:**
   ```sql
   ALTER TABLE workspaces ADD COLUMN app_id INTEGER;
   UPDATE workspaces SET app_id = (
     SELECT apps.id FROM apps 
     JOIN projects ON apps.name = projects.name 
     WHERE projects.id = workspaces.project_id
   );
   ALTER TABLE workspaces DROP COLUMN project_id;
   ```

5. **Update Context to include new fields:**
   ```sql
   ALTER TABLE context ADD COLUMN active_ecosystem_id INTEGER;
   ALTER TABLE context ADD COLUMN active_domain_id INTEGER;
   ALTER TABLE context ADD COLUMN active_app_id INTEGER;
   -- Migrate active_project_id to active_app_id
   ```

6. **Update CLI commands:**
   - Add `dvm get ecosystems`, `dvm get domains`, `dvm get apps`
   - Add deprecation warnings to `dvm create project`, `dvm get projects`

7. **Backward compatibility:**
   - `dvm create project` still works but warns and creates App in default Domain
   - Existing configs continue to work with warnings

---

## Summary

This document defines the complete vision for DevOpsMaestro as a **local DevOps platform**. The key concepts are:

1. **Hierarchy:** Ecosystem → Domain → App → Workspace
2. **App** is the core object - the codebase that exists for years
3. **Workspace** = App in dev mode (where you write code)
4. **live mode** = App running in production-like environment (Operator manages)
5. **Task** → **Workflow** → **Pipeline** → **Orchestration** - Composable CI/CD
6. **Operator** = Kubernetes Operator (CRD-based), not a local daemon
7. **dvm without Operator** = basic workspace management (Docker only)
8. **dvm with Operator** = full DevOps platform (requires k8s)
9. **Dependencies** - Global definitions, scoped instances
10. **Templates** - Reusable defaults for any object

The architecture follows these principles:
- Generic over specific
- Composition over inheritance
- Declarative everything
- kubectl patterns

Implementation is phased, starting with Hierarchy Foundation (Ecosystem, Domain, App) in v0.8.x and building toward full local DevOps with Operator, CI/CD, and chaos testing.

---

**Document Version:** 2.0  
**Last Updated:** Refined hierarchy with Domain and App objects  
**Next Steps:** Begin Phase 1 implementation (v0.8.0)
