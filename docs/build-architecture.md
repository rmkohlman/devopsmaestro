# How Builds Work

When you run `dvm build`, DevOpsMaestro builds a container image for your active workspace.

---

## What Gets Built

Each workspace has a Dockerfile generated from its YAML configuration. The build process:

1. Generates a Dockerfile based on your workspace's `language`, `version`, and plugin settings
2. Builds the image using your active container platform (OrbStack, Docker Desktop, Colima, or Podman)
3. Optionally pushes the image to a local OCI registry (Zot) for caching
4. Records the build outcome in the database and updates the workspace's image field with the built tag

---

## Build Commands

```bash
# Build workspace image
dvm build

# Build without using the local registry cache
dvm build --no-cache

# Build and push to local registry
dvm build --push

# Build against a specific registry
dvm build --registry <url>
```

---

## Build Session Tracking

Every `dvm build` run creates a **build session** in the database. A session records:

- **UUID** — unique session identifier
- **Timestamps** — start and completion time
- **Status** — `in_progress`, `succeeded`, or `failed`
- **Per-workspace results** — status, duration, built image tag, and any error message for each workspace in the build

After a successful build, the workspace record's image field is updated with the actual built image tag (e.g., `dvm-dev-my-api:20260409-153012`). Build sessions older than 30 days are automatically cleaned up.

```bash
# Show the most recent build session
dvm build status

# Show a specific session by UUID
dvm build status --session-id <uuid>

# List the 10 most recent sessions
dvm build status --history
```

---

## Local Registry Cache

DevOpsMaestro can use a local Zot OCI registry as a pull-through cache. This means:

- **First build**: Base images are pulled from Docker Hub and cached locally
- **Subsequent builds**: Images are served from the local cache — much faster
- **Offline support**: Cached images are available without network access
- **Rate limit avoidance**: Fewer pulls from Docker Hub (avoids the 100 pulls/6 hour limit)

Set up a local cache registry:

```bash
# Create and start a Zot registry
dvm create registry local-cache --type zot --port 5000
dvm start registry local-cache

# Build now uses the cache automatically
dvm build
```

See [Registry Management](../dvm/commands.md#registry-management) for full details.

---

## Supported Platforms

DevOpsMaestro automatically detects and uses your installed container platform:

| Platform | Build Tool | Notes |
|----------|-----------|-------|
| OrbStack | Docker/BuildKit | Recommended for macOS |
| Docker Desktop | Docker/BuildKit | Full feature support |
| Colima | nerdctl/BuildKit | Uses containerd backend |
| Podman | Podman Build | Rootless support |

Check your detected platform:

```bash
dvm get platforms
```

---

## SSH Agent Forwarding

If your build needs SSH access (e.g., to clone private repos), attach with SSH agent forwarding:

```bash
dvm attach --ssh-agent
```

This makes your host SSH agent available inside the container without copying private keys.
