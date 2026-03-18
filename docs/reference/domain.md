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
  theme: gruvbox-dark
  build:
    args:
      NPM_REGISTRY: "https://npm.corp.com/registry"
  apps:
    - api-service
    - user-service
    - auth-service
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
| `spec.build` | object | ❌ | Build configuration inherited by all workspaces in this domain |
| `spec.build.args` | map[string]string | ❌ | Build arguments passed as Docker `--build-arg` to all workspace builds |

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

### spec.build.args (optional)

Build arguments that cascade down to all apps and workspaces in this domain. Each key-value pair is passed as `--build-arg KEY=VALUE` during `dvm build`. Values are not persisted in image layers (they map to `ARG` declarations in the generated Dockerfile, not `ENV`).

```yaml
spec:
  build:
    args:
      NPM_REGISTRY: "https://npm.corp.com/registry"
      GITHUB_PAT: "ghp_abc123"
```

**Cascade order (most specific level wins):**
```
global < ecosystem < domain < app < workspace
```

An arg defined at the domain level overrides any matching arg from the ecosystem or global level, and is itself overridden by app- or workspace-level definitions. Use `dvm get build-args --effective --workspace <name>` to see the fully merged result with provenance for any workspace.

Manage domain-level build args with:

```bash
dvm set build-arg NPM_REGISTRY "https://npm.corp.com/registry" --domain backend
dvm get build-args --domain backend
dvm delete build-arg NPM_REGISTRY --domain backend
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
  apps:
    - api-service
    - user-service
    - auth-service
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
  apps:
    - web-app
    - admin-portal
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
  apps:
    - data-pipeline
    - analytics-service
```

## Related Resources

- [Ecosystem](ecosystem.md) - Parent organizational grouping
- [App](app.md) - Applications within this domain
- [Workspace](workspace.md) - Development environments
- [Credential](credential.md) - Secrets scoped to this domain
- [NvimTheme](nvim-theme.md) - Theme definitions

## Validation Rules

- `metadata.name` must be unique within the parent ecosystem
- `metadata.name` must be a valid DNS subdomain (lowercase, alphanumeric, hyphens)
- `metadata.ecosystem` must reference an existing Ecosystem resource
- `spec.apps` references must exist as App resources within this domain
- `spec.theme` must reference an existing theme (built-in or custom)