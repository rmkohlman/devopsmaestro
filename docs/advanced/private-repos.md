# Private Repository Support Guide

DevOpsMaestro provides comprehensive support for building projects that depend on private repositories across all major programming languages.

## Supported Authentication Methods

### 1. SSH Keys (Recommended for Git)
**Most secure for private Git repositories**

```bash
# Ensure SSH agent is running with your keys loaded
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_rsa

# Build with SSH forwarding
dvm build
```

**How it works:**
- BuildKit mounts your SSH agent into the build
- Git operations use SSH authentication automatically
- No credentials embedded in Dockerfile or images

**Supported URL formats:**
- `git+ssh://git@github.com/company/repo.git`
- `git@github.com:company/repo.git`

---

### 2. Build Args (For HTTPS with Tokens)
**Use when SSH is not available**

```bash
# Set environment variables
export GITHUB_USERNAME=youruser
export GITHUB_PAT=ghp_yourtoken

# Build (args are automatically detected and passed)
dvm build
```

**How it works:**
- Detects `${VARIABLE}` patterns in dependency files
- Adds ARG declarations to Dockerfile
- Substitutes variables at build time using sed
- Credentials passed as build arguments

**Security note:** Build args are visible in image history. Use secrets for production.

---

### 3. BuildKit Secrets (Coming Soon)
**Most secure for tokens and credentials**

```bash
# Build with secrets
echo "ghp_token" | dvm build --secret GITHUB_TOKEN=-
```

---

## Language-Specific Support

### Python

**Detection:**
- Scans `requirements.txt` for:
  - `git+https://` URLs
  - `git+ssh://` URLs  
  - `${VARIABLE}` patterns

**SSH Example:**
```txt
# requirements.txt
my-package @ git+ssh://git@github.com/company/private-repo.git@v1.0.0
```

**HTTPS Example:**
```txt
# requirements.txt
my-package @ git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/company/repo.git@v1.0.0
```

**Generated Dockerfile:**
```dockerfile
ARG GITHUB_USERNAME
ARG GITHUB_PAT

FROM python:3.11-slim
RUN apt-get install git openssh-client
RUN --mount=type=ssh \
    pip install -r requirements.txt
```

---

### Go

**Detection:**
- Scans `go.mod` for:
  - Private `github.com/company/*` repos
  - `git@` SSH URLs
  - Non-standard repo URLs

**SSH Example:**
```go
// go.mod
require github.com/company/private-lib v1.2.3
```

**Generated Dockerfile:**
```dockerfile
FROM golang:1.22-alpine
RUN apk add git openssh-client

# Configure git with SSH
RUN --mount=type=ssh \
    git config --global url."ssh://git@github.com/".insteadOf "https://github.com/"
```

**HTTPS with Token:**
```dockerfile
ARG GITHUB_TOKEN

RUN git config --global url."https://${GITHUB_TOKEN}@github.com/".insteadOf "https://github.com/"
```

---

### Node.js / TypeScript

**Detection:**
- Scans `package.json` for:
  - `git+https://` / `git+ssh://` URLs
  - Scoped packages (`@company/package`)
  - `${VARIABLE}` patterns

**SSH Example:**
```json
{
  "dependencies": {
    "my-package": "git+ssh://git@github.com/company/private-repo.git#v1.0.0"
  }
}
```

**HTTPS with Token:**
```json
{
  "dependencies": {
    "my-package": "git+https://${NPM_TOKEN}@github.com/company/repo.git"
  }
}
```

**Generated Dockerfile:**
```dockerfile
ARG NPM_TOKEN

FROM node:18-alpine
RUN apk add git openssh-client

# Setup .npmrc
RUN echo "//registry.npmjs.org/:_authToken=${NPM_TOKEN}" > ~/.npmrc

# Or with SSH
RUN --mount=type=ssh \
    npm install
```

---

### Java (Maven/Gradle)

**Detection:**
- Scans `pom.xml` for `<repository>` tags
- Scans `build.gradle` for `maven {}` blocks

**Generated Dockerfile:**
```dockerfile
ARG MAVEN_USERNAME
ARG MAVEN_PASSWORD

FROM maven:3.9-eclipse-temurin-17
RUN echo "<servers><server><id>private</id><username>${MAVEN_USERNAME}</username><password>${MAVEN_PASSWORD}</password></server></servers>" > ~/.m2/settings.xml
```

---

### Rust

**Detection:**
- Scans `Cargo.toml` for:
  - `git = "https://..."` entries
  - `git = "ssh://..."` entries

**SSH Example:**
```toml
# Cargo.toml
[dependencies]
my-crate = { git = "ssh://git@github.com/company/repo.git", tag = "v1.0" }
```

**Generated Dockerfile:**
```dockerfile
FROM rust:1.75-alpine
RUN apk add git openssh-client

RUN --mount=type=ssh \
    cargo build --release
```

---

## Automatic Detection

DevOpsMaestro automatically:

1. **Scans dependency files** for private repo patterns
2. **Detects authentication method** (SSH vs HTTPS)
3. **Extracts required variables** from `${VAR}` patterns
4. **Generates appropriate Dockerfile** with:
   - ARG declarations
   - Git/SSH installation
   - Authentication setup
   - Dependency installation with mounts

---

## Security Best Practices

### ✅ DO

- Use SSH keys for Git authentication (most secure)
- Load keys into ssh-agent before building
- Use environment variables for tokens
- Rotate tokens regularly
- Use BuildKit secrets for production builds

### ❌ DON'T

- Embed tokens directly in Dockerfiles
- Commit `.env` files with credentials
- Use weak or shared tokens
- Disable SSH host key checking

---

## Troubleshooting

### SSH Agent Not Available

```bash
# Start SSH agent
eval "$(ssh-agent -s)"

# Add your key
ssh-add ~/.ssh/id_rsa

# Verify
ssh-add -l
```

### Git Clone Fails with SSH

```bash
# Test SSH connection
ssh -T git@github.com

# Add GitHub to known_hosts
ssh-keyscan github.com >> ~/.ssh/known_hosts
```

### Build Args Not Substituted

```bash
# Ensure variables are exported
export GITHUB_USERNAME=myuser
export GITHUB_PAT=mytoken

# Verify
echo $GITHUB_USERNAME
```

### Permission Denied (publickey)

Your SSH key may not be loaded or doesn't have access to the repo.

```bash
# Ensure key is added
ssh-add ~/.ssh/id_rsa

# Verify repo access
git clone git@github.com:company/repo.git /tmp/test
```

---

## Advanced Configuration

### Custom SSH Key Location

```bash
# Add specific key
ssh-add ~/.ssh/company_rsa

# Build with forwarding
dvm build
```

### Multiple Git Hosts

The system automatically adds known_hosts for:
- github.com
- gitlab.com

For custom hosts:

```bash
# Add to your ~/.ssh/config
Host git.company.com
  HostName git.company.com
  User git
  IdentityFile ~/.ssh/company_rsa
```

---

## How It Works Internally

### 1. Detection Phase (`utils/private_repo_detector.go`)

```go
privateRepoInfo := utils.DetectPrivateRepos(projectPath, language)
// Returns:
// - NeedsGit: bool
// - NeedsSSH: bool
// - RequiredBuildArgs: []string
// - GitURLType: "ssh", "https", or "mixed"
```

### 2. Generation Phase (`builders/dockerfile_generator.go`)

- Adds ARG declarations if needed
- Installs git + openssh-client
- Configures SSH known_hosts
- Adds `--mount=type=ssh` to RUN commands
- Configures git credential helpers

### 3. Build Phase (`builders/api_image_builder.go`)

- Creates BuildKit session
- Attaches SSH agent provider
- Forwards `SSH_AUTH_SOCK`
- Passes build args from environment

---

## Example: Complete Python Project

**Project structure:**
```
my-project/
├── requirements.txt
├── pyproject.toml
└── src/
```

**requirements.txt:**
```txt
# Public packages
requests==2.31.0
pandas==2.1.0

# Private packages with SSH
internal-lib @ git+ssh://git@github.com/company/internal-lib.git@v2.0.0

# Private with HTTPS token
api-client @ git+https://${GITHUB_USERNAME}:${GITHUB_PAT}@github.com/company/api-client.git@v1.5.0
```

**Build process:**

```bash
# Setup authentication
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_rsa
export GITHUB_USERNAME=myuser
export GITHUB_PAT=ghp_xxxxx

# Create and build
dvm create project my-project --from-cwd
dvm use project my-project
dvm use workspace main
dvm build
```

**Generated Dockerfile.dvm:**
```dockerfile
# Build arguments for private repositories
ARG GITHUB_USERNAME
ARG GITHUB_PAT

# Base stage
FROM python:3.11-slim AS base

# Install git and SSH for private repos
RUN apt-get update && apt-get install -y \
    git \
    openssh-client \
    gcc \
    python3-dev \
    && rm -rf /var/lib/apt/lists/*

# Setup SSH
RUN mkdir -p /root/.ssh && chmod 700 /root/.ssh
RUN --mount=type=ssh \
    ssh-keyscan github.com >> /root/.ssh/known_hosts

# Process requirements with variable substitution
COPY requirements.txt /tmp/requirements-template.txt
RUN cat /tmp/requirements-template.txt | \
    sed "s/\${GITHUB_USERNAME}/$GITHUB_USERNAME/g" | \
    sed "s/\${GITHUB_PAT}/$GITHUB_PAT/g" | \
    tee /tmp/requirements.txt

# Install with SSH mount
RUN --mount=type=ssh \
    pip install --no-cache-dir -r /tmp/requirements.txt

# Dev stage
FROM base AS dev
...
```

**Result:**
- ✅ SSH keys used for internal-lib
- ✅ Tokens substituted for api-client
- ✅ No credentials in final image
- ✅ All dependencies installed successfully

---

## Supported Platforms

- ✅ GitHub (github.com)
- ✅ GitLab (gitlab.com)
- ✅ Bitbucket (bitbucket.org)
- ✅ Azure DevOps (dev.azure.com)
- ✅ Self-hosted Git servers
- ✅ Private npm registries
- ✅ Private Maven/Artifactory
- ✅ Private Cargo registries

---

## Coming Soon

- [ ] BuildKit secrets API (more secure than build args)
- [ ] `.netrc` file support
- [ ] Credential helper integration
- [ ] AWS CodeArtifact support
- [ ] Google Artifact Registry
- [ ] Certificate bundle mounting
- [ ] Proxy configuration
