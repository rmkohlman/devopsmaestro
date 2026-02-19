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
  labels:
    team: backend-team
    type: api
    language: golang
  annotations:
    description: "RESTful API for user management"
    repository: "https://github.com/company/api-service"
spec:
  path: /Users/dev/projects/api-service
  language:
    name: go
    version: "1.22"
  repo:
    url: "https://github.com/company/api-service"
    branch: "main"
  build:
    dockerfile: ./Dockerfile
    target: production
    context: .
    args:
      GITHUB_TOKEN: ${GITHUB_TOKEN}
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
  theme: coolnight-synthwave
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
| `metadata.labels` | object | ❌ | Key-value labels for organization |
| `metadata.annotations` | object | ❌ | Key-value annotations for metadata |
| `spec.path` | string | ✅ | Absolute path to source code |
| `spec.language` | object | ❌ | Programming language configuration |
| `spec.repo` | object | ❌ | Repository information |
| `spec.build` | object | ❌ | Build configuration |
| `spec.dependencies` | object | ❌ | Dependency management |
| `spec.services` | array | ❌ | External services (databases, etc.) |
| `spec.ports` | array | ❌ | Ports the app exposes |
| `spec.env` | object | ❌ | Application environment variables |
| `spec.theme` | string | ❌ | Default theme for workspaces |
| `spec.workspaces` | array | ❌ | List of workspace names |

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

### spec.language (optional)
Programming language and version information.

```yaml
spec:
  language:
    name: go          # go, python, node, rust, java, etc.
    version: "1.22"   # Language version
```

**Supported languages:**
- `go` - Go/Golang
- `python` - Python
- `node` - Node.js/JavaScript
- `rust` - Rust
- `java` - Java
- `dotnet` - .NET/C#

### spec.repo (optional)
Repository information for the source code.

```yaml
spec:
  repo:
    url: "https://github.com/company/api-service"
    branch: "main"    # Optional: default branch
```

### spec.build (optional)
Build configuration for containerization.

```yaml
spec:
  build:
    dockerfile: ./Dockerfile       # Path to Dockerfile
    buildpack: auto               # Or: go, python, node, etc.
    target: production            # Multi-stage build target
    context: .                    # Build context path
    args:                         # Build arguments
      GITHUB_TOKEN: ${GITHUB_TOKEN}
      BUILD_ENV: production
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
List of workspace names that belong to this app.

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

# From current directory
cd ~/projects/my-app
dvm create app my-app --from-cwd

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
- [NvimTheme](nvim-theme.md) - Theme definitions

## Validation Rules

- `metadata.name` must be unique within the parent domain
- `metadata.name` must be a valid DNS subdomain
- `metadata.domain` must reference an existing Domain resource
- `spec.path` must be an existing directory path
- `spec.language.name` must be a supported language
- `spec.ports` must be valid port mappings (1-65535)
- `spec.theme` must reference an existing theme
- `spec.workspaces` references must exist as Workspace resources