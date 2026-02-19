# Domain YAML Reference

**Kind:** `Domain`  
**APIVersion:** `devopsmaestro.io/v1`

A Domain represents a bounded context within an ecosystem. It groups related applications together based on business domain boundaries.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: backend
  ecosystem: my-platform
  labels:
    team: backend-team
    tech-stack: microservices
  annotations:
    description: "Backend services and APIs"
    slack-channel: "#backend-team"
spec:
  apps:
    - api-service
    - user-service
    - auth-service
  theme: gruvbox-dark
  defaults:
    language:
      name: go
      version: "1.22"
    shell:
      theme: starship
    nvim:
      plugins:
        - neovim/nvim-lspconfig
        - fatih/vim-go
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `Domain` |
| `metadata.name` | string | ✅ | Unique name for the domain |
| `metadata.ecosystem` | string | ✅ | Parent ecosystem name |
| `metadata.labels` | object | ❌ | Key-value labels for organization |
| `metadata.annotations` | object | ❌ | Key-value annotations for metadata |
| `spec.apps` | array | ❌ | List of app names in this domain |
| `spec.theme` | string | ❌ | Default theme for apps/workspaces in this domain |
| `spec.defaults` | object | ❌ | Default configurations for apps and workspaces |

## Field Details

### metadata.name (required)
The unique identifier for the domain within the ecosystem.

**Examples:**
- `backend`
- `frontend`
- `data-platform`
- `infrastructure`

### metadata.ecosystem (required)
The name of the parent ecosystem this domain belongs to. Must reference an existing Ecosystem resource.

```yaml
metadata:
  name: backend
  ecosystem: my-platform  # References Ecosystem/my-platform
```

### spec.apps (optional)
List of application names that belong to this domain. These are references to App resources.

```yaml
spec:
  apps:
    - api-service      # References App/api-service
    - user-service     # References App/user-service
    - auth-service     # References App/auth-service
```

### spec.theme (optional)
Default theme that applies to all apps and workspaces in this domain, overriding the ecosystem theme.

Theme hierarchy: `Workspace → App → Domain → Ecosystem → System Default`

```yaml
spec:
  theme: gruvbox-dark  # Overrides ecosystem theme for this domain
```

### spec.defaults (optional)
Default configuration values inherited by all apps and workspaces in this domain.

```yaml
spec:
  defaults:
    language:
      name: go           # Default language for apps
      version: "1.22"    # Default language version
    shell:
      type: zsh
      theme: starship
    nvim:
      structure: lazyvim
      plugins:
        - neovim/nvim-lspconfig
        - fatih/vim-go     # Go-specific plugins
    container:
      resources:
        cpus: "1.0"
        memory: "2G"
```

## Usage Examples

### Create Domain

```bash
# From YAML file
dvm apply -f domain.yaml

# Imperative command
dvm create domain my-platform/backend
```

### Set Domain Theme

```bash
# Set theme for domain (affects all apps and workspaces)
dvm set theme gruvbox-dark --domain backend
```

### List Domains

```bash
# List all domains
dvm get domains

# List domains in specific ecosystem
dvm get domains --ecosystem my-platform
```

### Export Domain

```bash
# Export to YAML
dvm get domain backend -o yaml

# Export with all apps and workspaces
dvm get domain backend --include-children -o yaml
```

## Domain Examples by Use Case

### Backend Services Domain

```yaml
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: backend
  ecosystem: company-platform
spec:
  theme: coolnight-ocean
  defaults:
    language:
      name: go
      version: "1.22"
    nvim:
      plugins:
        - neovim/nvim-lspconfig
        - fatih/vim-go
        - ray-x/go.nvim
```

### Frontend Domain

```yaml
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: frontend
  ecosystem: company-platform
spec:
  theme: coolnight-synthwave
  defaults:
    language:
      name: node
      version: "20"
    nvim:
      plugins:
        - neovim/nvim-lspconfig
        - nvim-treesitter/nvim-treesitter
        - windwp/nvim-autopairs
```

### Data Platform Domain

```yaml
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: data
  ecosystem: company-platform
spec:
  theme: coolnight-forest
  defaults:
    language:
      name: python
      version: "3.11"
    nvim:
      plugins:
        - neovim/nvim-lspconfig
        - nvim-treesitter/nvim-treesitter
        - jupyter-vim/jupyter-vim
```

## Related Resources

- [Ecosystem](ecosystem.md) - Parent organizational grouping
- [App](app.md) - Applications within this domain
- [Workspace](workspace.md) - Development environments
- [NvimTheme](nvim-theme.md) - Theme definitions

## Validation Rules

- `metadata.name` must be unique within the parent ecosystem
- `metadata.name` must be a valid DNS subdomain (lowercase, alphanumeric, hyphens)
- `metadata.ecosystem` must reference an existing Ecosystem resource
- `spec.apps` references must exist as App resources within this domain
- `spec.theme` must reference an existing theme (built-in or custom)
- `spec.defaults` values must be valid configuration options