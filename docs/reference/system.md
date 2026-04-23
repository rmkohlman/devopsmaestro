# System YAML Reference

**Kind:** `System`  
**APIVersion:** `devopsmaestro.io/v1`

A System represents a logical grouping of related applications within a domain. It sits between Domain and App in the hierarchy, allowing you to cluster apps that share infrastructure, configuration, or deployment context.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: payments
  domain: backend
  ecosystem: my-platform
  labels:
    team: payments-team
    tier: critical
  annotations:
    description: "Payment processing and billing system"
    slack-channel: "#payments-team"
spec:
  theme: gruvbox-dark
  nvimPackage: go-dev
  terminalPackage: devops-shell
  build:
    args:
      NPM_REGISTRY: "https://npm.corp.com/registry"
  caCerts:
    - name: corp-root-ca
      vaultSecret: corp-root-ca-pem
    - name: internal-ca
      vaultSecret: internal-ca-pem
      vaultField: certificate
  apps:
    - payment-api
    - billing-service
    - invoice-worker
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `System` |
| `metadata.name` | string | ✅ | Unique name for the system |
| `metadata.domain` | string | ❌ | Parent domain name |
| `metadata.ecosystem` | string | ❌ | Parent ecosystem name — inherited from domain when omitted |
| `metadata.labels` | object | ❌ | Key-value labels for organization |
| `metadata.annotations` | object | ❌ | Key-value annotations for metadata |
| `spec.theme` | string | ❌ | Default theme for apps/workspaces in this system |
| `spec.nvimPackage` | string | ❌ | Default Neovim plugin package cascaded to all workspaces in this system |
| `spec.terminalPackage` | string | ❌ | Default terminal package cascaded to all workspaces in this system |
| `spec.build` | object | ❌ | Build configuration inherited by all workspaces in this system |
| `spec.build.args` | map[string]string | ❌ | Build arguments passed as Docker `--build-arg` to all workspace builds |
| `spec.caCerts` | array | ❌ | CA certificates cascaded to all workspace builds in this system |
| `spec.caCerts[].name` | string | ✅ | Certificate name (must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`; max 64 chars) |
| `spec.caCerts[].vaultSecret` | string | ✅ | MaestroVault secret name containing the PEM certificate |
| `spec.caCerts[].vaultEnvironment` | string | ❌ | Vault environment override |
| `spec.caCerts[].vaultField` | string | ❌ | Field within the secret (default: `cert`) |
| `spec.apps` | array | ❌ | List of app names in this system |

## Field Details

### metadata.name (required)
The unique identifier for the system within its domain.

**Examples:**
- `payments`
- `auth`
- `data-ingestion`
- `notification`

### metadata.domain (optional)
The name of the parent domain this system belongs to. Optional — when omitted, `dvm apply` resolves using the active context.

```yaml
metadata:
  name: payments
  domain: backend  # References Domain/backend
```

### metadata.ecosystem (optional)
The name of the parent ecosystem. Inherited from the parent domain when omitted. Provide explicitly for context-free apply.

```yaml
metadata:
  name: payments
  domain: backend
  ecosystem: my-platform  # Enables context-free apply
```

### spec.apps (optional)
List of application names that belong to this system. These are references to App resources. Populated automatically on `dvm get system -o yaml`.

```yaml
spec:
  apps:
    - payment-api        # References App/payment-api
    - billing-service    # References App/billing-service
    - invoice-worker     # References App/invoice-worker
```

### spec.theme (optional)
Default theme that applies to all apps and workspaces in this system, overriding the domain and ecosystem theme.

Theme hierarchy: `Workspace → App → System → Domain → Ecosystem → Global Default`

```yaml
spec:
  theme: gruvbox-dark  # Overrides domain theme for this system
```

### spec.nvimPackage (optional)
Default Neovim plugin package that cascades to all workspaces in this system. References a `NvimPackage` resource by name. Overrides the domain-level `nvimPackage`; overridden at App or Workspace level.

```yaml
spec:
  nvimPackage: go-dev   # References NvimPackage/go-dev
```

### spec.terminalPackage (optional)
Default terminal package that cascades to all workspaces in this system. References a `TerminalPackage` resource by name. Overrides the domain-level `terminalPackage`; overridden at App or Workspace level.

```yaml
spec:
  terminalPackage: devops-shell   # References TerminalPackage/devops-shell
```

### spec.build.args (optional)

Build arguments that cascade down to all apps and workspaces in this system. Each key-value pair is passed as `--build-arg KEY=VALUE` during `dvm build`. Values are not persisted in image layers (they map to `ARG` declarations in the generated Dockerfile, not `ENV`).

```yaml
spec:
  build:
    args:
      NPM_REGISTRY: "https://npm.corp.com/registry"
      GITHUB_PAT: "ghp_abc123"
```

**Cascade order (most specific level wins):**
```
global < ecosystem < domain < system < app < workspace
```

An arg defined at the system level overrides any matching arg from the domain, ecosystem, or global level, and is itself overridden by app- or workspace-level definitions. Use `dvm get build-args --effective --workspace <name>` to see the fully merged result with provenance for any workspace.

Manage system-level build args with:

```bash
dvm set build-arg NPM_REGISTRY "https://npm.corp.com/registry" --system payments
dvm get build-args --system payments
dvm delete build-arg NPM_REGISTRY --system payments
```

### spec.caCerts (optional)

CA certificates that cascade down to all apps and workspaces in this system. Each entry references a PEM certificate stored in MaestroVault. Certificates are fetched at build time and injected into the container image via `COPY certs/ /usr/local/share/ca-certificates/custom/` + `RUN update-ca-certificates`. Missing or invalid certificates are a fatal build error.

Note: `spec.caCerts` is a **top-level spec field**, not nested under `spec.build`.

```yaml
spec:
  caCerts:
    - name: corp-root-ca
      vaultSecret: corp-root-ca-pem
    - name: internal-ca
      vaultSecret: internal-ca-pem
      vaultField: certificate
```

**Cascade order (most specific level wins by cert name):**
```
global < ecosystem < domain < system < app < workspace
```

A cert defined at the system level overrides any matching cert from the domain, ecosystem, or global level, and is itself overridden by app- or workspace-level definitions. Use `dvm get ca-certs --effective --workspace <name>` to see the fully merged result with provenance for any workspace.

Manage system-level CA certs with:

```bash
dvm set ca-cert corp-root-ca --vault-secret corp-root-ca-pem --system payments
dvm get ca-certs --system payments
dvm delete ca-cert corp-root-ca --system payments
```

## Usage Examples

### Create System

```bash
# From YAML file
dvm apply -f system.yaml

# Imperative command
dvm create system payments --domain backend
```

### Set System Theme

```bash
# Set theme for system (affects all apps and workspaces)
dvm set theme gruvbox-dark --system payments
```

### List Systems

```bash
# List all systems
dvm get systems

# List systems in specific domain
dvm get systems --domain backend

# List systems in specific ecosystem
dvm get systems --ecosystem my-platform
```

### Export System

```bash
# Export to YAML
dvm get system payments -o yaml

# Export with all apps and workspaces
dvm get system payments --include-children -o yaml
```

## System Examples by Use Case

### Payment Processing System

```yaml
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: payments
  domain: backend
  ecosystem: company-platform
spec:
  theme: coolnight-ocean
  apps:
    - payment-api
    - billing-service
    - invoice-worker
```

### Authentication System

```yaml
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: auth
  domain: backend
  ecosystem: company-platform
spec:
  theme: coolnight-synthwave
  apps:
    - auth-api
    - token-service
    - user-directory
```

### Data Ingestion System

```yaml
apiVersion: devopsmaestro.io/v1
kind: System
metadata:
  name: data-ingestion
  domain: data
  ecosystem: company-platform
spec:
  theme: coolnight-forest
  apps:
    - kafka-consumer
    - event-router
    - data-validator
```

## Related Resources

- [Domain](domain.md) - Parent bounded context
- [App](app.md) - Applications within this system
- [Workspace](workspace.md) - Development environments
- [Ecosystem](ecosystem.md) - Top-level organizational grouping

## Validation Rules

- `metadata.name` must be unique within the parent domain
- `metadata.name` must be a valid DNS subdomain (lowercase, alphanumeric, hyphens)
- `metadata.domain`, if provided, must reference an existing Domain resource
- `metadata.ecosystem`, if provided, must reference an existing Ecosystem resource
- `spec.apps` references must exist as App resources within this system
- `spec.theme` must reference an existing theme (built-in or custom)
- `spec.nvimPackage` must reference an existing NvimPackage resource
- `spec.terminalPackage` must reference an existing TerminalPackage resource
