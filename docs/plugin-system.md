# DevOpsMaestro Plugin System

The plugin system is designed like Kubernetes - plugins are **first-class objects** stored in the database, and workspace YAMLs **reference** them by name.

## Architecture

```
┌─────────────────────────────────────────┐
│  Database (Source of Truth)             │
│                                          │
│  ┌────────────────────────────────┐     │
│  │  nvim_plugins table            │     │
│  │  - telescope                   │     │
│  │  - mason                       │     │
│  │  - treesitter                  │     │
│  │  - copilot                     │     │
│  └────────────────────────────────┘     │
│                                          │
│  ┌────────────────────────────────┐     │
│  │  workspace_plugins junction    │     │
│  │  workspace_id | plugin_id      │     │
│  └────────────────────────────────┘     │
└─────────────────────────────────────────┘
            ↑                    ↑
            │                    │
    ┌───────┴────────┐   ┌──────┴────────┐
    │  Plugin YAML   │   │ Workspace YAML│
    │  (view/input)  │   │ (references)  │
    └────────────────┘   └───────────────┘
```

## Three Ways to Build Configs

### 1. Apply Individual Plugin YAMLs (Build Library)

Create reusable plugin definitions:

```bash
# Create telescope.yaml
cat > telescope.yaml << 'EOF'
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
  description: "Fuzzy finder for files, grep, buffers"
  category: fuzzy-finder
  tags: ["finder", "search"]
spec:
  repo: nvim-telescope/telescope.nvim
  branch: 0.1.x
  dependencies:
    - nvim-lua/plenary.nvim
    - repo: nvim-telescope/telescope-fzf-native.nvim
      build: make
  config: |
    local telescope = require("telescope")
    telescope.setup({
      defaults = {
        path_display = { "smart" },
      },
    })
  keymaps:
    - key: "<leader>ff"
      mode: n
      action: "<cmd>Telescope find_files<cr>"
      desc: "Find files"
EOF

# Apply to database (becomes reusable object)
dvm plugin apply telescope.yaml
```

### 2. Reference Plugins by Name (Compose from Library)

Workspace YAML references plugins stored in DB:

```yaml
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: main
  project: my-project
spec:
  image:
    name: dvm-main-my-project
  nvim:
    structure: custom
    plugins:
      # Just reference by name - definitions are in DB!
      - telescope
      - mason
      - treesitter
      - copilot
      - gitsigns
      - lazygit
  container:
    user: dev
    uid: 1000
```

### 3. Multi-File Apply (Kubernetes Style)

Apply multiple plugin definitions at once:

```bash
# Directory structure:
plugins/
├── colorscheme.yaml
├── telescope.yaml
├── mason.yaml
├── treesitter.yaml
├── lsp.yaml
└── copilot.yaml

# Apply all plugins
dvm plugin apply plugins/*.yaml

# Or use a single multi-document YAML (like kubectl)
cat > nvim-stack.yaml << 'EOF'
---
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: telescope
spec:
  repo: nvim-telescope/telescope.nvim
---
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: mason
spec:
  repo: williamboman/mason.nvim
---
apiVersion: devopsmaestro.io/v1
kind: NvimPlugin
metadata:
  name: copilot
spec:
  repo: zbirenbaum/copilot.lua
EOF

dvm plugin apply nvim-stack.yaml
```

## Commands

### Apply Plugin Definitions
```bash
# Create/update a plugin from YAML
dvm plugin apply telescope.yaml

# Apply multiple files
dvm plugin apply plugins/*.yaml
```

### List Plugins
```bash
# List all plugins
dvm plugin list

# Filter by category
dvm plugin list --category lsp
dvm plugin list --category fuzzy-finder
```

### Get Plugin as YAML
```bash
# Export plugin definition as YAML
dvm plugin get telescope

# Save to file
dvm plugin get telescope > telescope.yaml
```

### Delete Plugin
```bash
# Delete with confirmation
dvm plugin delete telescope

# Force delete (no confirmation)
dvm plugin delete telescope --force
```

## Workflow Examples

### Scenario 1: Personal Plugin Library
```bash
# Build your personal nvim plugin library
cd ~/.devopsmaestro/plugins
dvm plugin apply *.yaml

# Reference in any workspace
cat > workspace.yaml << 'EOF'
apiVersion: devopsmaestro.io/v1
kind: Workspace
spec:
  nvim:
    plugins: [telescope, mason, copilot, gitsigns]
EOF
```

### Scenario 2: Team Shared Plugins
```bash
# Team maintains plugins in git repo
git clone git@github.com:myteam/dvm-plugins.git
cd dvm-plugins
dvm plugin apply *.yaml

# Everyone on team has same plugin definitions
# Each dev can customize their workspace.yaml
```

### Scenario 3: Export Existing Config
```bash
# Export current plugins to share
dvm plugin get telescope > plugins/telescope.yaml
dvm plugin get mason > plugins/mason.yaml

# Commit to git
git add plugins/
git commit -m "Add telescope and mason plugins"
```

### Scenario 4: Different Workspaces, Same Plugins
```yaml
# workspace-minimal.yaml - Only essentials
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: minimal
spec:
  nvim:
    plugins: [telescope, treesitter]
---
# workspace-full.yaml - Full IDE setup
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: full
spec:
  nvim:
    plugins:
      - telescope
      - mason
      - treesitter
      - copilot
      - gitsigns
      - lazygit
      - trouble
      - nvim-cmp
```

## Benefits

1. **Reusability** - Define once, use in multiple workspaces
2. **Version Control** - Plugin definitions can be in git, tracked separately from workspace configs
3. **Sharing** - Export/import plugin definitions between machines or team members
4. **Composition** - Mix and match plugins to create different development environments
5. **Single Source of Truth** - Database stores canonical definitions, YAML is just a view
6. **Declarative** - Just list plugin names, DVM handles the rest

## Similar to Kubernetes

Like Kubernetes ConfigMaps/Secrets referenced by Pods:

```yaml
# Kubernetes style
apiVersion: v1
kind: Pod
spec:
  containers:
    - name: app
      envFrom:
        - configMapRef:
            name: app-config  # Reference by name
        - secretRef:
            name: app-secrets # Reference by name
```

```yaml
# DVM style
apiVersion: devopsmaestro.io/v1
kind: Workspace
spec:
  nvim:
    plugins:
      - telescope  # Reference by name
      - mason      # Reference by name
      - copilot    # Reference by name
```

Both systems:
- Store objects in a database/API server
- Reference objects by name in higher-level configs
- Support apply/get/delete operations
- Treat YAML as a serialization format, not the source of truth
