# YAML Reference

Complete YAML schemas for all DevOpsMaestro resource types.

## Resource Types

DevOpsMaestro supports the following resource types with kubectl-style YAML configuration:

### Core Resources

| Resource | APIVersion | Description |
|----------|------------|-------------|
| [Ecosystem](ecosystem.md) | `devopsmaestro.io/v1` | Top-level platform grouping (organization) |
| [Domain](domain.md) | `devopsmaestro.io/v1` | Bounded context within an ecosystem |
| [System](system.md) | `devopsmaestro.io/v1` | Logical grouping of related apps within a domain (optional) |
| [App](app.md) | `devopsmaestro.io/v1` | Application/codebase within a domain |
| [Workspace](workspace.md) | `devopsmaestro.io/v1` | Development environment for an app |
| [Credential](credential.md) | `devopsmaestro.io/v1` | Secret reference (MaestroVault or env) scoped to an ecosystem, domain, app, or workspace |

### Extensibility Resources

| Resource | APIVersion | Description |
|----------|------------|-------------|
| [CustomResourceDefinition](custom-resource-definition.md) | `devopsmaestro.io/v1alpha1` | Register a custom resource type to extend DevOpsMaestro |

### Infrastructure Resources

| Resource | APIVersion | Description |
|----------|------------|-------------|
| [Registry](registry.md) | `devopsmaestro.io/v1` | Local package registry (OCI, Python, Go, npm, HTTP proxy) |

## Object Hierarchy

```
Ecosystem → Domain → App → Workspace
   (org)    (context) (code)  (dev env)
```

Resources are organized hierarchically, with themes and configurations cascading down through the hierarchy.

## Base YAML Structure

All DevOpsMaestro resources follow Kubernetes-style YAML structure:

```yaml
apiVersion: devopsmaestro.io/v1  # or devopsmaestro.io/v1alpha1 for CRDs
kind: <ResourceType>
metadata:
  name: <resource-name>
  labels: {}                     # Optional labels
  annotations: {}                # Optional annotations
spec:
  # Resource-specific configuration
```

## Common Usage Patterns

### Export Resources

```bash
# Export any resource to YAML
dvm get ecosystem my-platform -o yaml
dvm get app my-api -o yaml
dvm get workspace dev -o yaml
dvm get nvim theme coolnight-ocean -o yaml
```

### Apply Resources

```bash
# Apply from file
dvm apply -f resource.yaml

# Apply from URL
dvm apply -f https://themes.example.com/theme.yaml

# Apply from GitHub (shorthand)
dvm apply -f github:user/repo/theme.yaml
```

### Multi-Document YAML

You can combine multiple resources in a single YAML file:

```yaml
---
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: my-theme
spec:
  # theme configuration
---
apiVersion: devopsmaestro.io/v1
kind: Workspace
metadata:
  name: dev
  app: my-app
spec:
  nvim:
    theme: my-theme
```

## Validation

DevOpsMaestro validates all YAML resources on import:

- **Required fields** - `apiVersion`, `kind`, `metadata.name`
- **Field types** - String, integer, boolean, array validation
- **Enum values** - Valid values for enumerated fields
- **References** - Theme names, plugin dependencies
- **Format** - Color codes, repository URLs, paths

See each resource page for specific validation rules.