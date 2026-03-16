# CustomResourceDefinition YAML Reference

**Kind:** `CustomResourceDefinition`  
**APIVersion:** `devopsmaestro.io/v1alpha1`

A CustomResourceDefinition (CRD) registers a new resource type with DevOpsMaestro. Once registered, instances of that type can be created, retrieved, and managed using `dvm apply`. CRDs are inspired by Kubernetes CRDs and allow teams to extend DevOpsMaestro with domain-specific resource types.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1alpha1
kind: CustomResourceDefinition
metadata:
  name: databases.custom.devopsmaestro.io
spec:
  group: custom.devopsmaestro.io
  names:
    kind: Database
    singular: database
    plural: databases
    shortNames:
      - db
  scope: App
  versions:
    - name: v1
      served: true
      storage: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            engine:
              type: string
            version:
              type: string
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | Yes | Must be `devopsmaestro.io/v1alpha1` |
| `kind` | string | Yes | Must be `CustomResourceDefinition` |
| `metadata.name` | string | Yes | Conventionally `<plural>.<group>` (e.g., `databases.custom.devopsmaestro.io`) |
| `spec.group` | string | No | API group for the custom resource (e.g., `custom.devopsmaestro.io`) |
| `spec.names.kind` | string | Yes | PascalCase type name (e.g., `Database`) |
| `spec.names.singular` | string | Yes | Lowercase singular name (e.g., `database`) |
| `spec.names.plural` | string | Yes | Lowercase plural name (e.g., `databases`) |
| `spec.names.shortNames` | array | No | Short aliases (e.g., `db`) |
| `spec.scope` | string | Yes | One of: `Global`, `Ecosystem`, `Domain`, `App`, or `Workspace` |
| `spec.versions` | array | Yes | Must contain at least one version entry |
| `spec.versions[].name` | string | Yes | Version identifier (e.g., `v1`) |
| `spec.versions[].served` | bool | Yes | Whether this version is actively served |
| `spec.versions[].storage` | bool | Yes | Whether this is the storage version |
| `spec.versions[].schema` | object | No | OpenAPI v3 schema for instance validation |

## Field Details

### metadata.name (required)

A unique identifier for this CRD. By convention, use `<plural>.<group>` to mirror Kubernetes naming:

```yaml
metadata:
  name: databases.custom.devopsmaestro.io
```

### spec.group (optional)

The API group for instances of this resource. Use a reverse-DNS style group to avoid collisions:

```yaml
spec:
  group: custom.devopsmaestro.io
```

### spec.names (required)

Defines the naming vocabulary for the resource type.

```yaml
spec:
  names:
    kind: Database        # PascalCase — used in YAML kind field
    singular: database    # Lowercase singular — used in CLI
    plural: databases     # Lowercase plural — used in CLI list commands
    shortNames:
      - db                # Optional aliases
```

### spec.scope (required)

Controls where instances of this resource can be scoped. Supported values:

| Value | Description |
|-------|-------------|
| `Global` | Not scoped to any hierarchy level |
| `Ecosystem` | Scoped to an ecosystem |
| `Domain` | Scoped to a domain |
| `App` | Scoped to an app |
| `Workspace` | Scoped to a workspace |

### spec.versions (required)

At least one version entry must be provided. Each version can carry its own schema.

```yaml
spec:
  versions:
    - name: v1
      served: true      # This version is active
      storage: true     # This version is used for storage
      schema:
        openAPIV3Schema:
          type: object
          properties:
            engine:
              type: string
            version:
              type: string
```

`storage: true` must be set on exactly one version. If multiple versions are defined, only the storage version is persisted.

### spec.versions[].schema (optional)

An OpenAPI v3 schema for validating instances of this resource. When omitted, instances are accepted without structural validation.

```yaml
schema:
  openAPIV3Schema:
    type: object
    properties:
      host:
        type: string
      port:
        type: integer
      tls:
        type: boolean
```

## CLI Commands

### Apply a CRD

Register a new custom resource type from a YAML file:

```bash
dvm apply -f database-crd.yaml
```

Applying an existing CRD by kind name updates it in place.

### Get a CRD

```bash
dvm get crd Database
```

### List all CRDs

```bash
dvm get crds
```

### Delete a CRD

```bash
dvm delete crd Database
```

A CRD cannot be deleted while instances of its kind exist.

## Validation Rules

The following fields are validated when a CRD is applied:

- `spec.names.kind` is required
- `spec.names.singular` is required
- `spec.names.plural` is required
- `spec.scope` is required
- `spec.versions` must contain at least one entry
- `spec.names.kind` must not shadow a built-in kind (see below)
- `spec.versions[].schema` is compiled against the OpenAPI v3 validator when present

**Built-in kinds that cannot be overridden:**

`Workspace`, `App`, `Domain`, `Ecosystem`, `NvimPlugin`, `NvimTheme`, `NvimPackage`, `TerminalPrompt`, `TerminalPackage`, `TerminalPlugin`, `TerminalEmulator`, `Registry`, `GitRepo`, `Credential`, `CustomResourceDefinition`

**Note:** The `Validate()` method on the `CustomResourceDefinition` model is stubbed and always returns `nil`. All meaningful validation is performed by `CRDHandler.Apply()` and the OpenAPI v3 schema compiler in `pkg/crd`. This is an intentional extensibility point for future validation logic.

## Notes

- CRDs are an alpha-stage extensibility feature (`v1alpha1`). The schema and behavior may change in future releases.
- Registering a CRD does not create any instances. Use `dvm apply -f` with a YAML file whose `kind` matches the registered `spec.names.kind` to create instances.
- The `metadata.name` field on a CRD is informational. The system indexes CRDs by `spec.names.kind`, not by `metadata.name`.
- Schema validation for instances is performed using the OpenAPI v3 schema supplied in `spec.versions[].schema`. Instances are only validated when a schema is present.

## Related Resources

- [Ecosystem](ecosystem.md) - Top-level scope
- [Domain](domain.md) - Domain scope
- [App](app.md) - App scope
- [Workspace](workspace.md) - Workspace scope
