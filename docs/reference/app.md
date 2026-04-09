# App YAML Reference

**Kind:** `App`  
**APIVersion:** `devopsmaestro.io/v1`

An App represents a codebase or application within a domain. App configuration focuses on **what the code needs to build and run** in production.

## Full Example

```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: api-service
  domain: backend
  ecosystem: my-platform
  labels:
    team: backend-team
    type: api
    language: golang
  annotations:
    description: "RESTful API for user management"
    repository: "https://github.com/company/api-service"
spec:
  path: /Users/dev/projects/api-service
  theme: coolnight-synthwave
  nvimPackage: go-dev
  terminalPackage: devops-shell
  gitRepo: api-service-repo
  language:
    name: go
    version: "1.22"
  build:
    dockerfile: ./Dockerfile
    target: production
    context: .
    buildpack: go
    args:
      GITHUB_TOKEN: ${GITHUB_TOKEN}
    caCerts:
      - name: corp-root-ca
        vaultSecret: corp-root-ca-pem
      - name: internal-ca
        vaultSecret: internal-ca-pem
        vaultField: certificate
  dependencies:
    file: go.mod
    install: go mod download
    extra:
      - github.com/gin-gonic/gin
  services:
    - name: postgres
      version: "15"
      port: 5432
      env:
        POSTGRES_USER: apiservice
        POSTGRES_PASSWORD: secret
        POSTGRES_DB: apiservice
    - name: redis
      version: "7"
      port: 6379
  ports:
    - "8080:8080"
    - "8443:8443"
    - "9090:9090"
  env:
    DATABASE_URL: postgres://localhost:5432/apiservice
    REDIS_URL: redis://localhost:6379
    LOG_LEVEL: debug
    GO111MODULE: "on"
  workspaces:
    - main
    - debug
    - testing
```

## Field Reference

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `apiVersion` | string | ✅ | Must be `devopsmaestro.io/v1` |
| `kind` | string | ✅ | Must be `App` |
| `metadata.name` | string | ✅ | Unique name for the app |
| `metadata.domain` | string | ✅ | Parent domain name |
| `metadata.ecosystem` | string | ❌ | Parent ecosystem name — enables context-free apply without `dvm use ecosystem` |
| `metadata.labels` | object | ❌ | Key-value labels for organization |
| `metadata.annotations` | object | ❌ | Key-value annotations for metadata |
| `spec.path` | string | ✅ | Absolute path to source code on the local filesystem |
| `spec.theme` | string | ❌ | Default theme for workspaces in this app |
| `spec.nvimPackage` | string | ❌ | Default Neovim plugin package for workspaces in this app |
| `spec.terminalPackage` | string | ❌ | Default terminal package for workspaces in this app |
| `spec.gitRepo` | string | ❌ | Name of a GitRepo resource to associate with this app |
| `spec.language` | object | ❌ | Programming language configuration |
| `spec.language.name` | string | ❌ | Language name: `go`, `python`, `node`, `rust`, `java`, `dotnet` |
| `spec.language.version` | string | ❌ | Language version (e.g., `"1.22"`, `"3.11"`, `"20"`) |
| `spec.build` | object | ❌ | Build configuration |
| `spec.build.dockerfile` | string | ❌ | Path to an existing Dockerfile |
| `spec.build.buildpack` | string | ❌ | Buildpack to use (`auto`, `go`, `python`, `node`, etc.) |
| `spec.build.target` | string | ❌ | Multi-stage Dockerfile build target |
| `spec.build.context` | string | ❌ | Build context path (defaults to app path) |
| `spec.build.args` | map[string]string | ❌ | Build arguments emitted as `ARG` declarations (not `ENV`) |
| `spec.build.caCerts` | array | ❌ | CA certificates fetched from MaestroVault at build time |
| `spec.build.caCerts[].name` | string | ✅ | Certificate name (must match `^[a-zA-Z0-9][a-zA-Z0-9_-]*$`; max 64 chars) |
| `spec.build.caCerts[].vaultSecret` | string | ✅ | MaestroVault secret name containing the PEM certificate |
| `spec.build.caCerts[].vaultEnvironment` | string | ❌ | Vault environment override |
| `spec.build.caCerts[].vaultField` | string | ❌ | Field within the secret (default: `cert`) |
| `spec.dependencies` | object | ❌ | Dependency management configuration |
| `spec.dependencies.file` | string | ❌ | Dependency file: `go.mod`, `requirements.txt`, `package.json` |
| `spec.dependencies.install` | string | ❌ | Command to install dependencies |
| `spec.dependencies.extra` | array | ❌ | Additional dependencies to install |
| `spec.services` | array | ❌ | Sidecar services the app depends on |
| `spec.services[].name` | string | ✅ | Service name (e.g., `postgres`, `redis`, `mongodb`) |
| `spec.services[].image` | string | ❌ | Custom Docker image (defaults to official image) |
| `spec.services[].version` | string | ❌ | Service version/tag |
| `spec.services[].port` | int | ❌ | Port to expose |
| `spec.services[].env` | map[string]string | ❌ | Service environment variables |
| `spec.ports` | array | ❌ | Port mappings the app exposes (format: `"host:container"`) |
| `spec.env` | map[string]string | ❌ | Application-level environment variables |
| `spec.workspaces` | array | ❌ | List of workspace names belonging to this app |

## Field Details

### metadata.name (required)
The unique identifier for the app within the domain.

**Examples:**
- `api-service`
- `user-service`
- `web-frontend`

### metadata.domain (required)
The name of the parent domain this app belongs to. Must reference an existing Domain resource.

```yaml
metadata:
  name: api-service
  domain: backend  # References Domain/backend
```

### metadata.ecosystem (optional)
The name of the parent ecosystem. Optional but recommended — when present, `dvm apply` can resolve the app without requiring `dvm use ecosystem` to be set first.

```yaml
metadata:
  name: api-service
  domain: backend
  ecosystem: my-platform  # Enables context-free apply
```

### spec.path (required)
Absolute path to the source code directory on the local filesystem.

```yaml
spec:
  path: /Users/dev/projects/api-service
```

**Variable substitution supported:**
```yaml
spec:
  path: ${HOME}/projects/api-service
```

### spec.gitRepo (optional)
Name of a GitRepo resource to associate with this app. When set, `dvm apply` links the app to the named GitRepo by ID.

```yaml
spec:
  gitRepo: api-service-repo  # References GitRepo/api-service-repo
```

### spec.language (optional)
Programming language and version information.

```yaml
spec:
  language:
    name: go          # go, python, node, rust, java, dotnet
    version: "1.22"   # Language version
```

**Supported languages:**
- `go` - Go/Golang
- `python` - Python
- `node` - Node.js/JavaScript
- `rust` - Rust
- `java` - Java
- `dotnet` - .NET/C#

### spec.build (optional)
Build configuration for containerization.

```yaml
spec:
  build:
    dockerfile: ./Dockerfile       # Path to Dockerfile
    buildpack: auto               # Or: go, python, node, etc.
    target: production            # Multi-stage build target
    context: .                    # Build context path
    args:                         # Build arguments (ARG, not ENV)
      GITHUB_TOKEN: ${GITHUB_TOKEN}
      BUILD_ENV: production
    caCerts:                      # CA certificates from MaestroVault
      - name: corp-root-ca
        vaultSecret: corp-root-ca-pem
```

### spec.build.caCerts (optional)

CA certificates for this app's workspace builds. Each entry references a PEM certificate stored in MaestroVault. Certificates are fetched at build time and injected into the container image via `COPY certs/ /usr/local/share/ca-certificates/custom/` + `RUN update-ca-certificates`. Missing or invalid certificates are a fatal build error.

```yaml
spec:
  build:
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

An app-level cert overrides any matching cert from higher levels (domain, ecosystem, global). Individual workspaces can further override by defining a cert with the same name. Use `dvm get ca-certs --effective --workspace <name>` to see the fully merged result with provenance.

Manage app-level CA certs with:

```bash
dvm set ca-cert corp-root-ca --vault-secret corp-root-ca-pem --app my-api
dvm get ca-certs --app my-api
dvm delete ca-cert corp-root-ca --app my-api
```

### spec.dependencies (optional)
Dependency management configuration.

```yaml
spec:
  dependencies:
    file: go.mod                  # go.mod, requirements.txt, package.json
    install: go mod download      # Command to install dependencies
    extra:                        # Additional dependencies
      - github.com/gin-gonic/gin
      - github.com/lib/pq
```

### spec.services (optional)
External services the app depends on (databases, caches, message queues).

```yaml
spec:
  services:
    - name: postgres              # Service name
      version: "15"              # Service version
      port: 5432                 # Port number
      env:                       # Service environment variables
        POSTGRES_USER: myapp
        POSTGRES_PASSWORD: secret
        POSTGRES_DB: myapp
    - name: redis
      version: "7"
      port: 6379
```

### spec.ports (optional)
Ports that the application exposes.

```yaml
spec:
  ports:
    - "8080:8080"     # HTTP API
    - "8443:8443"     # HTTPS API
    - "9090:9090"     # Metrics endpoint
    - "2345:2345"     # Debug port
```

### spec.env (optional)
Application-level environment variables.

```yaml
spec:
  env:
    DATABASE_URL: postgres://localhost:5432/myapp
    REDIS_URL: redis://localhost:6379
    LOG_LEVEL: debug
    API_VERSION: v1
```

### spec.theme (optional)
Default theme for all workspaces in this app, overriding domain and ecosystem themes.

```yaml
spec:
  theme: coolnight-synthwave
```

### spec.workspaces (optional)
List of workspace names that belong to this app. Populated automatically on `dvm get app -o yaml`.

```yaml
spec:
  workspaces:
    - main      # Primary development workspace
    - debug     # Debugging workspace
    - testing   # Testing workspace
```

## Language-Specific Examples

### Go/Golang App

```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: go-api
  domain: backend
  ecosystem: my-platform
spec:
  path: /Users/dev/projects/go-api
  language:
    name: go
    version: "1.22"
  build:
    buildpack: go
  dependencies:
    file: go.mod
    install: go mod download
  ports:
    - "8080:8080"
  env:
    GO111MODULE: "on"
    GOOS: linux
```

### Python FastAPI App

```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: fastapi-service
  domain: backend
  ecosystem: my-platform
spec:
  path: /Users/dev/projects/fastapi-service
  language:
    name: python
    version: "3.11"
  build:
    dockerfile: ./Dockerfile
  dependencies:
    file: requirements.txt
    install: pip install -r requirements.txt
  services:
    - name: postgres
      version: "15"
      port: 5432
  ports:
    - "8000:8000"
  env:
    DATABASE_URL: postgres://localhost:5432/mydb
```

### Node.js App

```yaml
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: node-api
  domain: frontend
  ecosystem: my-platform
spec:
  path: /Users/dev/projects/node-api
  language:
    name: node
    version: "20"
  dependencies:
    file: package.json
    install: npm install
  ports:
    - "3000:3000"
    - "9229:9229"    # Debug port
  env:
    NODE_ENV: development
```

## Usage Examples

### Create App

```bash
# From YAML file
dvm apply -f app.yaml

# Imperative command
dvm create app backend/my-api
```

### Set App Theme

```bash
# Set theme for app (affects all workspaces)
dvm set theme tokyonight-night --app my-api
```

### Export App

```bash
# Export to YAML
dvm get app my-api -o yaml

# Export with all workspaces
dvm get app my-api --include-workspaces -o yaml
```

## Related Resources

- [Domain](domain.md) - Parent bounded context
- [Workspace](workspace.md) - Development environments for this app
- [Credential](credential.md) - Secrets scoped to this app
- [NvimPackage](nvim-package.md) - Plugin package definitions
- [NvimTheme](nvim-theme.md) - Theme definitions

## Validation Rules

- `metadata.name` must be unique within the parent domain
- `metadata.name` must be a valid DNS subdomain
- `metadata.domain` must reference an existing Domain resource
- `metadata.ecosystem`, if provided, must reference an existing Ecosystem resource
- `spec.path` must be an existing directory path
- `spec.gitRepo`, if provided, must reference an existing GitRepo resource
- `spec.language.name` must be a supported language
- `spec.ports` must be valid port mappings (1-65535)
- `spec.theme` must reference an existing theme
- `spec.workspaces` references must exist as Workspace resources
