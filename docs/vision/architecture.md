# DevOpsMaestro Architecture Vision

> **Purpose:** Complete architectural vision for DevOpsMaestro's evolution from a dev environment tool into a full local DevOps platform.
>
> **Status:** Living Document - defines the target state we're building toward.
>
> **Audience:** AI assistants, contributors, and future-you trying to understand the complete vision.

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

1. **Development Environments** (Workspaces) - Where you write code
2. **Running Services** (Services) - Where your code runs in production-like conditions
3. **CI/CD Pipelines** (Tasks, Workflows, Pipelines) - How code moves from dev to service
4. **Operator** - Automated daemon that watches for commits and orchestrates everything

### The Core Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DEVELOPMENT                                     │
│  ┌──────────────────┐                                                       │
│  │    WORKSPACE     │  ← You code here (nvim, tools, path to source)        │
│  │    auth-api      │                                                       │
│  │    env: dev      │                                                       │
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
│  │     SERVICE      │  │     SERVICE      │  │     SERVICE      │          │
│  │    auth-api      │  │    user-api      │  │   billing-api    │          │
│  │    env: live     │  │    env: live     │  │    env: live     │          │
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

- **Workspace** = `env: dev` - Where you develop (nvim, tools, source code)
- **Service** = `env: live` - Where your code runs after deployment
- Same codebase, different runtime contexts
- Multiple Services across Ecosystems can interact via shared Dependencies
- Enables **full microservices testing locally** on your laptop

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
| **Hierarchy** | Ecosystem, Project, Workspace, Service, Context | Organizational containers |
| **Execution** | Task, Workflow, Pipeline, Orchestration, Job | CI/CD and automation |
| **Data** | DataRecord, DataStore, DataLake | Test data management |
| **Infrastructure** | Dependency, Volume, Action, Runtime | Services and storage |
| **Reusability** | Template, ResourceDefinition | Templates and extensibility |
| **Dev Tools** | NvimPlugin, NvimTheme, TerminalPrompt, TerminalPlugin, TerminalTheme | Editor and shell config |
| **Operations** | Operator | Automation daemon |

---

### Hierarchy Objects

#### Ecosystem

Top-level grouping of related projects. Think of it as a "platform" or "product area."

```yaml
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: customer-platform
  description: All customer-facing services
spec:
  # Shared dependencies available to all projects/workspaces
  dependencies:
    - name: platform-db        # Instance name
      ref: postgres            # References global Dependency definition
    - name: platform-cache
      ref: redis
  
  # Shared data for integration testing
  dataLakes:
    - ref: integration-test-data
```

#### Project

Groups related workspaces within an ecosystem. Think of it as a "domain" or "bounded context."

```yaml
apiVersion: devopsmaestro.io/v1
kind: Project
metadata:
  name: auth-domain
  ecosystem: customer-platform
  description: Authentication and authorization services
spec:
  # Project-level dependencies (in addition to ecosystem)
  dependencies:
    - name: auth-secrets
      ref: vault
  
  # Project-level test data
  dataStores:
    - ref: auth-test-data
  
  # Default settings for workspaces in this project
  defaults:
    language: go
    languageVersion: "1.23"
```

#### Workspace

Development environment where you write code. Has a `path` to source code.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: auth-api
  project: auth-domain
  description: Authentication API service
spec:
  # Path to source code on host machine
  path: ~/Developer/auth-api
  
  language: go
  languageVersion: "1.23"
  
  # Environment is always 'dev' for workspaces
  environment: dev
  
  # Mount points into container
  mounts:
    - source: ~/Developer/auth-api
      target: /workspace
  
  # Use ecosystem's shared postgres instance
  dependencies:
    - ref: platform-db       # Same name = uses ecosystem's instance
  
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

#### Service

Running instance deployed from a Workspace. Lives in the "live" environment.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Service
metadata:
  name: auth-api
  project: auth-domain
  description: Running auth-api service
spec:
  # Which workspace this service is deployed from
  workspaceRef: auth-api
  
  # Environment is 'live' for services
  environment: live
  
  # Resource limits for local runtime
  resources:
    memory: 512Mi
    cpu: 500m
  
  # Port mappings
  ports:
    - containerPort: 8080
      hostPort: 8080
  
  # Health check
  healthCheck:
    path: /health
    port: 8080
    interval: 10s
  
  # Uses same dependencies as workspace
  dependencies:
    - ref: platform-db
```

#### Context

Active selection state - which ecosystem/project/workspace is currently selected.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Context
metadata:
  name: current
spec:
  ecosystem: customer-platform
  project: auth-domain
  workspace: auth-api
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
  project: auth-domain
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
  project: auth-domain
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
  project: auth-domain
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
  project: auth-domain
spec:
  # What triggers this job
  trigger:
    # Watch the workspace path for git commits
    watch: workspace.path
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
  project: auth-domain
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
  project: auth-domain
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

# Workspace uses the SAME instance (same name)
kind: Workspace
metadata:
  name: auth-api
spec:
  dependencies:
    - ref: platform-db       # Uses ecosystem's instance

# Workspace creates its OWN instance (different name)
kind: Workspace
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
kind: Workspace
metadata:
  name: auth-api
spec:
  template: go-api-workspace  # Apply template defaults
  
  # Override specific values
  languageVersion: "1.24"
  
  # Add additional plugins
  nvim:
    plugins:
      - copilot  # Added on top of template defaults
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

The automation daemon. Watches for events and orchestrates everything.

```yaml
apiVersion: devopsmaestro.io/v1
kind: Operator
metadata:
  name: default
spec:
  # What to watch
  watches:
    - type: workspace
      events: [commit]
      # Run this job when event occurs
      jobRef: on-commit-ci
    
    - type: service
      events: [unhealthy]
      jobRef: service-recovery
  
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
│  Template: go-api-workspace, python-service, node-frontend                  │
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
│  │                    PROJECT: auth-domain                              │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  Dependencies (instances): auth-secrets                              │   │
│  │  DataStore: auth-test-data                                          │   │
│  │  Task, Workflow, Pipeline, Job definitions                          │   │
│  │                                                                      │   │
│  │  ┌─────────────────────┐    ┌─────────────────────┐                 │   │
│  │  │ WORKSPACE: auth-api │    │ SERVICE: auth-api   │                 │   │
│  │  │ env: dev            │───▶│ env: live           │                 │   │
│  │  │ path: ~/auth-api    │    │ (deployed instance) │                 │   │
│  │  └─────────────────────┘    └─────────────────────┘                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │                    PROJECT: user-domain                              │   │
│  │  (similar structure)                                                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Relationship Table

| Object | Contains/References | Contained By |
|--------|---------------------|--------------|
| **Ecosystem** | Projects, Dependency instances, DataLakes, Volumes, Operator | - |
| **Project** | Workspaces, Services, Tasks, Workflows, Pipelines, Jobs, DataStores | Ecosystem |
| **Workspace** | NvimPlugins, NvimTheme, TerminalPrompt, TerminalPlugins | Project |
| **Service** | (deployed from Workspace) | Project |
| **Task** | Actions | Project |
| **Workflow** | Tasks | Project |
| **Pipeline** | Workflows | Project |
| **Orchestration** | Pipelines | Ecosystem |
| **Job** | Pipeline/Workflow/Task reference, Trigger | Project |
| **DataLake** | DataStores | Ecosystem |
| **DataStore** | DataRecords | Project |
| **DataRecord** | Data | Project |
| **Dependency** | (global definition) | Global |
| **Template** | Default values for any Kind | Global |

---

## Scoping Rules

### Scope Hierarchy

```
Global
  └── Ecosystem
        └── Project
              └── Workspace / Service
```

### Object Scope Summary

| Scope | Objects |
|-------|---------|
| **Global** | Dependency (definitions), Action, Template, ResourceDefinition, Runtime, NvimPlugin*, NvimTheme*, TerminalPrompt*, TerminalPlugin*, TerminalTheme* |
| **Ecosystem** | Ecosystem, Dependency (instances), DataLake, Volume, Operator, Orchestration |
| **Project** | Project, Workspace, Service, Task, Workflow, Pipeline, Job, DataStore, DataRecord |
| **Workspace** | (inherits from Project, references global dev tools) |

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

# Workspace A uses it
kind: Workspace
metadata:
  name: auth-api
spec:
  dependencies:
    - ref: platform-db       # Uses ecosystem's "platform-db"

# Workspace B also uses it
kind: Workspace
metadata:
  name: user-api
spec:
  dependencies:
    - ref: platform-db       # Same instance as auth-api

# Workspace C creates its own
kind: Workspace
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
   └── Starts Workspace container
   └── Mounts ~/Developer/auth-api to /workspace
   └── Starts shared dependencies (platform-db)
   └── Opens shell with nvim configured

2. Developer codes in Workspace
   └── Full nvim setup with LSP, plugins, theme
   └── Terminal with starship prompt, zsh plugins
   └── Access to platform-db at localhost:5432

3. Developer commits: git commit -m "Add feature"
   └── Operator detects commit
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
   
   └── Stage 2: deploy-to-service
       └── Task: deploy auth-api Service
   
   └── Stage 3: integration-tests
       └── Task: run integration tests against live Service
   
   └── Stage 4: chaos-tests (optional)
       └── Task: inject failures, verify resilience

2. On Pipeline success:
   └── Job.onSuccess: git.push to remote
   └── Notify: Slack message to #builds

3. On Pipeline failure:
   └── Job.onFailure: notify.alert
   └── Service NOT updated (stays at previous version)
```

### Workflow 3: Multi-Service Interaction

```
1. Multiple Ecosystems running:
   └── customer-platform
       ├── auth-api Service (env: live)
       ├── user-api Service (env: live)
       └── billing-api Service (env: live)
   
   └── internal-tools
       ├── admin-api Service (env: live)
       └── reporting-api Service (env: live)

2. All Services share:
   └── Global Redis (messaging)
   └── Global Kafka (events)

3. Developer can:
   └── Watch all Services interact
   └── Run chaos tests across Services
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
dvm get projects [-e ecosystem]
dvm create project auth-domain -e customer-platform
dvm get workspaces [-p project]
dvm get services [-p project]
dvm get context
dvm use ecosystem customer-platform
dvm use project auth-domain
dvm use workspace auth-api
```

#### Execution
```bash
dvm get tasks [-p project]
dvm run task lint-code
dvm get workflows [-p project]
dvm run workflow ci-workflow
dvm get pipelines [-p project]
dvm run pipeline ci-pipeline
dvm get orchestrations [-e ecosystem]
dvm run orchestration full-platform-test
dvm get jobs [-p project]
dvm status job on-commit-ci
```

#### Data
```bash
dvm get datarecords [-p project]
dvm get datastores [-p project]
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
| `--project` | `-p` | Project context |
| `--workspace` | `-w` | Workspace context |
| `--output` | `-o` | Output format (table, yaml, json) |
| `--all` | `-A` | All (across contexts) |
| `--file` | `-f` | YAML file to apply |
| `--verbose` | `-v` | Verbose output |

---

## Implementation Phases

### Phase 1: Foundation (v0.8.x - v0.9.x)

**Goal:** Add Ecosystem and restructure Workspace to have `path`.

| Task | Description | Priority |
|------|-------------|----------|
| Add Ecosystem object | New table, CRUD operations | High |
| Move `path` to Workspace | Currently on Project, move to Workspace | High |
| Add Service object | Basic Service definition | High |
| Update Context | Add ecosystem to context | Medium |
| Update CLI | New commands for ecosystem | Medium |

**Migration:** Existing Projects become Projects with a default Ecosystem.

### Phase 2: Execution (v1.0.x - v1.1.x)

**Goal:** Local CI/CD with Task, Workflow, Pipeline, Job.

| Task | Description | Priority |
|------|-------------|----------|
| Add Task object | Single action execution | High |
| Add Action system | Built-in actions (exec, etc.) | High |
| Add Workflow object | Task composition | High |
| Add Pipeline object | Stage-based CI/CD | High |
| Add Job object | Trigger definitions | Medium |
| Basic Operator | Watch for commits, trigger Jobs | Medium |

### Phase 3: Services & Data (v1.2.x - v1.3.x)

**Goal:** Deploy Services, manage test data.

| Task | Description | Priority |
|------|-------------|----------|
| Service deployment | Deploy Workspace to Service | High |
| Dependency instances | Shared dependency management | High |
| DataRecord/DataStore | Test data management | Medium |
| DataLake | Ecosystem-wide test data | Medium |
| Volume management | Persistent storage | Medium |

### Phase 4: Operations (v1.4.x - v1.5.x)

**Goal:** Full Operator with auto-push, chaos testing.

| Task | Description | Priority |
|------|-------------|----------|
| Full Operator | Watches, triggers, manages Services | High |
| Auto-push on success | Git push after successful pipeline | High |
| Orchestration | Multi-pipeline coordination | Medium |
| Chaos actions | chaos.kill, chaos.latency | Medium |
| Observability | Logs, metrics, status dashboards | Medium |

### Phase 5: Extensibility (v2.0.x+)

**Goal:** CRD-like extensibility, Templates, advanced features.

| Task | Description | Priority |
|------|-------------|----------|
| Template system | Reusable defaults | Medium |
| ResourceDefinition | Custom resource types | Medium |
| Custom Actions | User-defined actions | Low |
| Remote Runtime | Deploy to remote (future) | Low |

---

## Migration Path

### From Current State (v0.7.x) to Phase 1

**Current State:**
- Project has `path`
- Workspace has no `path`
- No Ecosystem, Service, Task, etc.

**Migration Steps:**

1. **Create default Ecosystem:**
   ```sql
   INSERT INTO ecosystems (name, description) 
   VALUES ('default', 'Default ecosystem for migrated projects');
   ```

2. **Link existing Projects to default Ecosystem:**
   ```sql
   ALTER TABLE projects ADD COLUMN ecosystem_id INTEGER;
   UPDATE projects SET ecosystem_id = 1;  -- default ecosystem
   ```

3. **Move `path` from Project to Workspace:**
   ```sql
   ALTER TABLE workspaces ADD COLUMN path TEXT;
   UPDATE workspaces SET path = (
     SELECT path FROM projects WHERE projects.id = workspaces.project_id
   );
   ALTER TABLE projects DROP COLUMN path;
   ```

4. **Update CLI commands:**
   - Add `dvm get ecosystems`, `dvm create ecosystem`, etc.
   - Update context to include ecosystem

5. **Backward compatibility:**
   - `dvm create project` without `-e` uses 'default' ecosystem
   - Existing configs continue to work

---

## Summary

This document defines the complete vision for DevOpsMaestro as a **local DevOps platform**. The key concepts are:

1. **Workspace** (dev) → **Service** (live) - Same code, different environments
2. **Task** → **Workflow** → **Pipeline** → **Orchestration** - Composable CI/CD
3. **Operator** - Watches commits, triggers Pipelines, deploys Services
4. **Dependencies** - Global definitions, scoped instances
5. **Templates** - Reusable defaults for any object
6. **DataRecord/DataStore/DataLake** - Composable test data

The architecture follows these principles:
- Generic over specific
- Composition over inheritance
- Declarative everything
- kubectl patterns

Implementation is phased, starting with Foundation (Ecosystem, Workspace path, Service) and building toward full local DevOps with Operator, CI/CD, and chaos testing.

---

**Document Version:** 1.0  
**Last Updated:** During session defining complete architecture vision  
**Next Steps:** Begin Phase 1 implementation
