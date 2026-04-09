# GlobalDefaults YAML Reference

**Kind:** `GlobalDefaults`  
**APIVersion:** `devopsmaestro.io/v1`

GlobalDefaults stores system-wide fallback values for theme, build args, CA certs, nvim/terminal packages, plugins, and registry routing. These values are the lowest-priority level in the cascade — any setting at the ecosystem, domain, app, or workspace level overrides the corresponding global default.

GlobalDefaults is a singleton resource. There is exactly one per installation. The `metadata.name` is always `global-defaults` and is ignored on apply.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: GlobalDefaults
metadata:
  name: global-defaults
spec:
  theme: tokyonight-night
  nvimPackage: go-dev
  terminalPackage: devops-shell
  plugins:
    - nvim-telescope
    - nvim-treesitter
  buildArgs:
    PIP_INDEX_URL: "https://pypi.corp.com/root/prod"
    NPM_REGISTRY: "https://npm.corp.com/registry"
  caCerts:
    - name: corp-root-ca
      vaultSecret: corp-root-ca-pem
    - name: internal-ca
      vaultSecret: internal-ca-pem
      vaultField: certificate
  registryOci: my-zot-registry
  registryPypi: my-devpi-registry
  registryNpm: my-verdaccio-registry
  registryGo: my-athens-registry
  registryHttp: my-squid-proxy
  registryIdleTimeout: 30m
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `GlobalDefaults` |
| `metadata.name` | string | ✅ | Always `global-defaults`; value is informational only |
| `spec.theme` | string | ❌ | Default theme cascaded to all workspaces (lowest priority) |
| `spec.nvimPackage` | string | ❌ | Default NvimPackage name cascaded to all workspaces |
| `spec.terminalPackage` | string | ❌ | Default TerminalPackage name cascaded to all workspaces |
| `spec.plugins` | []string | ❌ | Default plugin names applied globally |
| `spec.buildArgs` | map[string]string | ❌ | Global build args passed as `--build-arg`; lowest priority in cascade |
| `spec.caCerts` | array | ❌ | CA certificates injected globally into all workspace builds |
| `spec.caCerts[].name` | string | ✅ | Certificate name (must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`; max 64 chars) |
| `spec.caCerts[].vaultSecret` | string | ✅ | MaestroVault secret name containing the PEM certificate |
| `spec.caCerts[].vaultEnvironment` | string | ❌ | Vault environment override |
| `spec.caCerts[].vaultField` | string | ❌ | Field within the secret (default: `cert`) |
| `spec.registryOci` | string | ❌ | Default OCI registry resource name (used by `zot`) |
| `spec.registryPypi` | string | ❌ | Default PyPI registry resource name (used by `devpi`) |
| `spec.registryNpm` | string | ❌ | Default npm registry resource name (used by `verdaccio`) |
| `spec.registryGo` | string | ❌ | Default Go module proxy resource name (used by `athens`) |
| `spec.registryHttp` | string | ❌ | Default HTTP caching proxy resource name (used by `squid`) |
| `spec.registryIdleTimeout` | string | ❌ | Global default idle timeout for `on-demand` registries (e.g., `30m`, `1h`) |

## Field Details

### spec.theme (optional)

The global fallback theme name. Applied when no theme is set at the ecosystem, domain, app, or workspace level. References a built-in or custom NvimTheme by name.

**Cascade order (most specific wins):**
```
global < ecosystem < domain < app < workspace
```

```yaml
spec:
  theme: coolnight-ocean
```

### spec.nvimPackage (optional)

The global fallback NvimPackage. Applied to workspaces that have no NvimPackage set at any higher level. References a `NvimPackage` resource by name.

```yaml
spec:
  nvimPackage: go-dev
```

### spec.terminalPackage (optional)

The global fallback TerminalPackage. Applied to workspaces that have no TerminalPackage set at any higher level. References a `TerminalPackage` resource by name.

```yaml
spec:
  terminalPackage: devops-shell
```

### spec.plugins (optional)

A list of plugin names applied globally. These are stored as a JSON array in the defaults table internally.

```yaml
spec:
  plugins:
    - nvim-telescope
    - nvim-treesitter
    - nvim-lspconfig
```

### spec.buildArgs (optional)

Global build arguments passed as `--build-arg KEY=VALUE` to all workspace builds. These are the lowest-priority level — any ecosystem, domain, app, or workspace build arg with the same key overrides the global value.

```yaml
spec:
  buildArgs:
    PIP_INDEX_URL: "https://pypi.corp.com/root/prod"
    NPM_REGISTRY: "https://npm.corp.com/registry"
```

Use `dvm get build-args --effective --workspace <name>` to see the fully merged result with provenance.

### spec.caCerts (optional)

CA certificates injected into all workspace builds globally. Stored as a JSON array in the defaults table. Each entry references a PEM certificate in MaestroVault. Certificates are fetched at build time and installed via `update-ca-certificates`.

```yaml
spec:
  caCerts:
    - name: corp-root-ca
      vaultSecret: corp-root-ca-pem
    - name: internal-ca
      vaultSecret: internal-ca-pem
      vaultField: certificate
```

### spec.registryOci / registryPypi / registryNpm / registryGo / registryHttp (optional)

Global default registry resource names for each package type. When set, workspaces that do not specify a registry for a given type will use this global default. Each value is a Registry resource name.

```yaml
spec:
  registryOci: my-zot-registry      # Registry/my-zot-registry (type: zot)
  registryPypi: my-devpi-registry   # Registry/my-devpi-registry (type: devpi)
  registryNpm: my-verdaccio-registry # Registry/my-verdaccio-registry (type: verdaccio)
  registryGo: my-athens-registry    # Registry/my-athens-registry (type: athens)
  registryHttp: my-squid-proxy      # Registry/my-squid-proxy (type: squid)
```

### spec.registryIdleTimeout (optional)

Global default idle timeout for `on-demand` registries. Accepts a Go duration string (e.g., `30m`, `1h`, `1h30m`). Overridden by the `spec.idleTimeout` field on individual Registry resources.

```yaml
spec:
  registryIdleTimeout: 30m
```

## CLI Commands

### Apply GlobalDefaults

Restore or update global defaults from a YAML file:

```bash
dvm apply -f global-defaults.yaml
```

Applying GlobalDefaults is idempotent — only non-empty fields in the YAML overwrite existing defaults. Fields absent from the YAML are left unchanged.

### Get GlobalDefaults

```bash
dvm get globaldefaults
dvm get globaldefaults -o yaml
```

### Export for Backup

GlobalDefaults is included in a full export:

```bash
dvm get all -o yaml > backup.yaml
```

### Delete GlobalDefaults

Clears all global defaults (theme, build args, CA certs, packages, plugins, and all registry settings):

```bash
dvm delete globaldefaults global-defaults
```

## Singleton Behavior

- There is exactly one GlobalDefaults resource per installation.
- `dvm get globaldefaults` returns nothing if no defaults have been set.
- `dvm get all -o yaml` omits GlobalDefaults from the output if no defaults are set.
- `metadata.name` is always serialized as `global-defaults` on export and is ignored on apply.

## Validation Rules

- `spec.caCerts[].name` must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$` and be at most 64 characters
- `spec.caCerts[].vaultSecret` is required for each cert entry
- All other spec fields are optional; unset fields remain at their previous value on apply

## Related Resources

- [Ecosystem](ecosystem.md) — Ecosystem-level theme, build args, and CA certs (overrides GlobalDefaults)
- [Domain](domain.md) — Domain-level theme, build args, and CA certs
- [Registry](registry.md) — Registry resources referenced by `registryOci`, `registryPypi`, etc.
- [NvimPackage](nvim-package.md) — Package definitions referenced by `nvimPackage`
- [NvimTheme](nvim-theme.md) — Theme definitions referenced by `theme`
