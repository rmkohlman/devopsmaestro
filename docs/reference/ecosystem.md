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
  description: "Production platform for Acme Corp"
  theme: coolnight-ocean
  nvimPackage: go-dev
  terminalPackage: devops-shell
  build:
    args:
      PIP_INDEX_URL: "https://pypi.corp.com/root/prod"
  caCerts:
    - name: corp-root-ca
      vaultSecret: corp-root-ca-pem
    - name: internal-ca
      vaultSecret: internal-ca-pem
      vaultField: certificate
  domains:
    - backend
    - frontend
    - data
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `Ecosystem` |
| `metadata.name` | string | ✅ | Unique name for the ecosystem |
| `metadata.labels` | object | ❌ | Key-value labels for organization |
| `metadata.annotations` | object | ❌ | Key-value annotations for metadata |
| `spec.description` | string | ❌ | Human-readable description of the ecosystem |
| `spec.theme` | string | ❌ | Default theme for all domains/apps/workspaces |
| `spec.nvimPackage` | string | ❌ | Default Neovim plugin package cascaded to all workspaces |
| `spec.terminalPackage` | string | ❌ | Default terminal package cascaded to all workspaces |
| `spec.build` | object | ❌ | Build configuration inherited by all workspaces in this ecosystem |
| `spec.build.args` | map[string]string | ❌ | Build arguments passed as Docker `--build-arg` to all workspace builds |
| `spec.caCerts` | array | ❌ | CA certificates cascaded to all workspace builds in this ecosystem |
| `spec.caCerts[].name` | string | ✅ | Certificate name (must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`; max 64 chars) |
| `spec.caCerts[].vaultSecret` | string | ✅ | MaestroVault secret name containing the PEM certificate |
| `spec.caCerts[].vaultEnvironment` | string | ❌ | Vault environment override |
| `spec.caCerts[].vaultField` | string | ❌ | Field within the secret (default: `cert`) |
| `spec.domains` | array | ❌ | List of domain names in this ecosystem |

## Field Details

### metadata.name (required)
The unique identifier for the ecosystem. Must be a valid DNS subdomain name.

**Examples:**
- `my-platform`
- `acme-corp`
- `prod-env`

### spec.domains (optional)
List of domain names that belong to this ecosystem. These are references to Domain resources. Populated automatically on `dvm get ecosystem -o yaml`.

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

See [Theme Hierarchy](https://rmkohlman.github.io/MaestroTheme/configuration/theme-hierarchy/) for complete list.

### spec.nvimPackage (optional)
Default Neovim plugin package that cascades to all workspaces in this ecosystem. References a `NvimPackage` resource by name. Overridden at Domain, App, or Workspace level.

```yaml
spec:
  nvimPackage: go-dev   # References NvimPackage/go-dev
```

### spec.terminalPackage (optional)
Default terminal package that cascades to all workspaces in this ecosystem. References a `TerminalPackage` resource by name. Overridden at Domain, App, or Workspace level.

```yaml
spec:
  terminalPackage: devops-shell   # References TerminalPackage/devops-shell
```

### spec.build.args (optional)

Build arguments that cascade down to all domains, apps, and workspaces in this ecosystem. Each key-value pair is passed as `--build-arg KEY=VALUE` during `dvm build`. Values are not persisted in image layers (they map to `ARG` declarations in the generated Dockerfile, not `ENV`).

```yaml
spec:
  build:
    args:
      PIP_INDEX_URL: "https://pypi.corp.com/root/prod"
      NPM_REGISTRY: "https://npm.corp.com/registry"
```

**Cascade order (most specific level wins):**
```
global < ecosystem < domain < app < workspace
```

An arg defined at the ecosystem level is inherited by all domains, apps, and workspaces in this ecosystem unless overridden at a more specific level. Use `dvm get build-args --effective --workspace <name>` to see the fully merged result with provenance for any workspace.

Manage ecosystem-level build args with:

```bash
dvm set build-arg PIP_INDEX_URL "https://pypi.corp.com/root/prod" --ecosystem my-platform
dvm get build-args --ecosystem my-platform
dvm delete build-arg PIP_INDEX_URL --ecosystem my-platform
```

### spec.caCerts (optional)

CA certificates that cascade down to all domains, apps, and workspaces in this ecosystem. Each entry references a PEM certificate stored in MaestroVault. Certificates are fetched at build time and injected into the container image via `COPY certs/ /usr/local/share/ca-certificates/custom/` + `RUN update-ca-certificates`. Missing or invalid certificates are a fatal build error.

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
global < ecosystem < domain < app < workspace
```

A cert defined at the ecosystem level is inherited by all domains, apps, and workspaces in this ecosystem unless overridden at a more specific level. Use `dvm get ca-certs --effective --workspace <name>` to see the fully merged result with provenance for any workspace.

Manage ecosystem-level CA certs with:

```bash
dvm set ca-cert corp-root-ca --vault-secret corp-root-ca-pem --ecosystem my-platform
dvm get ca-certs --ecosystem my-platform
dvm delete ca-cert corp-root-ca --ecosystem my-platform
```

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

## Validation Rules

- `metadata.name` must be unique across all ecosystems
- `metadata.name` must be a valid DNS subdomain (lowercase, alphanumeric, hyphens)
- `spec.domains` references must exist as Domain resources
- `spec.theme` must reference an existing theme (built-in or custom)
- `spec.nvimPackage` must reference an existing NvimPackage resource (see [MaestroNvim](https://rmkohlman.github.io/MaestroNvim/))
- `spec.terminalPackage` must reference an existing TerminalPackage resource (see [MaestroTerminal](https://rmkohlman.github.io/MaestroTerminal/))
