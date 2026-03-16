# Ecosystem YAML Reference

**Kind:** `Ecosystem`  
**APIVersion:** `devopsmaestro.io/v1`

An Ecosystem represents the top-level organizational grouping in DevOpsMaestro. It typically represents a company, platform, or major organizational unit.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: my-platform
  labels:
    environment: production
    organization: acme-corp
  annotations:
    description: "Production platform for Acme Corp"
    contact: "platform-team@acme.com"
spec:
  domains:
    - backend
    - frontend
    - data
  theme: coolnight-ocean
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `Ecosystem` |
| `metadata.name` | string | ✅ | Unique name for the ecosystem |
| `metadata.labels` | object | ❌ | Key-value labels for organization |
| `metadata.annotations` | object | ❌ | Key-value annotations for metadata |
| `spec.domains` | array | ❌ | List of domain names in this ecosystem |
| `spec.theme` | string | ❌ | Default theme for all domains/apps/workspaces |

## Field Details

### metadata.name (required)
The unique identifier for the ecosystem. Must be a valid DNS subdomain name.

**Examples:**
- `my-platform`
- `acme-corp`
- `prod-env`

### spec.domains (optional)
List of domain names that belong to this ecosystem. These are references to Domain resources.

```yaml
spec:
  domains:
    - backend      # References Domain/backend
    - frontend     # References Domain/frontend
    - data         # References Domain/data
```

### spec.theme (optional)
Default theme name that cascades down to all domains, apps, and workspaces in this ecosystem unless overridden.

**Built-in themes available:**
- `coolnight-ocean` (default)
- `coolnight-synthwave`
- `tokyonight-night`
- `catppuccin-mocha`
- `gruvbox-dark`

See [Theme Hierarchy](../advanced/theme-hierarchy.md) for complete list.

## Usage Examples

### Create Ecosystem

```bash
# From YAML file
dvm apply -f ecosystem.yaml

# Imperative command
dvm create ecosystem my-platform
```

### Set Ecosystem Theme

```bash
# Set theme for entire ecosystem (affects all children)
dvm set theme coolnight-synthwave --ecosystem my-platform
```

### Export Ecosystem

```bash
# Export to YAML
dvm get ecosystem my-platform -o yaml

# Export with all domains and apps
dvm get ecosystem my-platform --include-children -o yaml
```

## Related Resources

- [Domain](domain.md) - Bounded contexts within an ecosystem
- [App](app.md) - Applications within domains
- [Workspace](workspace.md) - Development environments
- [Credential](credential.md) - Secrets scoped to this ecosystem
- [NvimTheme](nvim-theme.md) - Theme definitions

## Validation Rules

- `metadata.name` must be unique across all ecosystems
- `metadata.name` must be a valid DNS subdomain (lowercase, alphanumeric, hyphens)
- `spec.domains` references must exist as Domain resources
- `spec.theme` must reference an existing theme (built-in or custom)