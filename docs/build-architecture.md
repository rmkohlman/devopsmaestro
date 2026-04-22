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

## Parallel Builds

`dvm build --all` (or any scoped parallel build) runs the full 7-phase pipeline for each workspace:

1. **Validate app path** — confirms the source directory exists
2. **Detect platform & registry** — identifies the active container runtime and registry
3. **Prepare workspace spec** — loads YAML config and resolves workspace settings
4. **Source & staging** — copies source into an isolated staging directory
5. **CA certs & nvim config** — resolves certificates and generates Neovim configuration
6. **Dockerfile generation & build** — generates the Dockerfile and executes the Docker build
7. **Post-build** — updates the workspace image field in the database and (optionally) pushes to registry

Each workspace gets its own isolated staging directory, keyed by `appName-workspaceName`. This prevents `Dockerfile.dvm` collisions when multiple workspaces from the same app are built in parallel.

---

## Build Session Tracking

Every `dvm build` run creates a **build session** in the database. A session records:

- **UUID** — unique session identifier
- **Timestamps** — start and completion time
- **Status** — `in_progress`, `succeeded`, `failed`, `partial`, `interrupted`, or `cancelled` (workspace-level)
- **Per-workspace results** — status, duration, built image tag, and any error message for each workspace in the build

After a successful build, the workspace record's image field is updated with the actual built image tag (e.g., `dvm-dev-my-api:20260409-153012`). Build sessions older than 30 days are automatically cleaned up.

If a build is interrupted (Ctrl-C / SIGTERM), the session is finalized synchronously as `interrupted` and in-flight workspace rows are set to `cancelled`. Sessions that were not cleanly finalized (e.g., process killed hard) are auto-healed to `interrupted` the next time `dvm build status` is run, using a 10-minute staleness threshold.

Succeeded and failed workspace counts displayed at the end of a build are sourced from the persisted build session in the database — not from in-memory counters — to ensure accuracy across parallel builds.

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
| OrbStack | docker buildx / BuildKit | Recommended for macOS |
| Docker Desktop | docker buildx / BuildKit | Full feature support |
| Colima | nerdctl/BuildKit | Uses containerd backend |
| Podman | Podman Build | Rootless support |

Check your detected platform:

```bash
dvm get platforms
```

---

## Build Cache

### Builder Stage Cache Mounts

All builder stages in the generated Dockerfile use `--mount=type=cache` to preserve package manager caches between builds:

| Stage | Cache Mount |
|-------|------------|
| `neovim-builder` | `apt` cache (`/var/cache/apt`, `/var/lib/apt`) |
| `lazygit-builder` | `apt` or `apk` cache depending on base image |
| `starship-builder` | `apt` or `apk` cache |
| `treesitter-builder` | `apt` or `apk` cache |
| `go-tools-builder` | Go module cache (`/root/go/pkg/mod`) |

BuildKit isolates each cache mount per stage, so there is no lock contention between parallel builds. Package caches are preserved across builds even when layers change — apt/apk packages are not re-downloaded on every build.

### docker buildx

DevOpsMaestro uses `docker buildx build` for all Docker-compatible platforms (OrbStack, Docker Desktop). This enables BuildKit features including `--mount=type=cache` and lays the groundwork for registry-based layer caching in a future phase.

### Cache Readiness Reporting

Before each build, DevOpsMaestro checks all configured registries and reports which caches are available:

```
Cache: 3/5 registries active (zot-layer-cache: connection refused, devpi: not started)
```

This output appears in the build log so you can see whether caches are warm before the Docker build begins.

### Local Directory Layer Cache

`dvm build` uses BuildKit's `type=local` cache to persist Docker build layers to disk, scoped per app and workspace:

```
~/.devopsmaestro/build-cache/<app>-<workspace>/
```

This cache survives `docker system prune` and Docker restarts, so warm rebuilds remain fast even after Docker's internal layer cache is cleared. The cache is populated automatically on every build and disabled when `--no-cache` is passed.

To manage the build cache:

```bash
# Preview what would be cleared (dry run)
dvm cache clear --dry-run

# Clear only the local layer cache and BuildKit cache
dvm cache clear --buildkit

# Clear all caches (layer cache + BuildKit + staging directories)
dvm cache clear --all
```

See [`dvm cache clear`](../dvm/commands.md#dvm-cache-clear) for the full flag reference.

### Registry-Based Layer Cache (Future)

`--cache-from`/`--cache-to` with a Zot OCI registry for cross-machine layer cache sharing is not yet enabled. Docker BuildKit defaults to HTTPS for non-localhost registries, which requires adding the registry to Docker's `insecure-registries` daemon config. The current `type=local` disk cache already provides excellent warm rebuild performance for single-machine workflows.

---

## Build Logs

Every `dvm build` run writes a structured log file in addition to its terminal output.

### File Layout

```
~/.devopsmaestro/logs/builds/
├── latest.log -> 550e8400-e29b-41d4-a716-446655440000.log   # atomic symlink
├── 550e8400-e29b-41d4-a716-446655440000.log                 # current session
├── 3f2a1b00-c3d2-4e5f-a617-335566221100.log.gz              # rotated (compressed)
└── ...
```

- Each session gets its own file named after the **build session UUID** (matching `dvm build status`)
- `latest.log` is an atomic symlink updated at the start of every build — always points to the most recent session
- Files are created with mode **`0o600`** (user-only read/write); the parent directory is **`0o700`**
- Logs are flushed and closed cleanly on success, failure, **and** interrupt (Ctrl-C / SIGTERM)

### Rotation

Log rotation is handled by [lumberjack](https://github.com/natefinish/lumberjack). Old sessions are rotated when the file exceeds `maxSizeMB` and compressed in the background.

### Configuration

Add a `buildLogs:` block to `~/.devopsmaestro/config.yaml` to customise behaviour:

```yaml
buildLogs:
  enabled: true                                    # set false to disable file logging
  directory: ~/.devopsmaestro/logs/builds          # where log files are written
  maxSizeMB: 100                                   # rotate when file exceeds this size
  maxAgeDays: 7                                    # delete rotated files older than N days
  maxBackups: 10                                   # keep at most N rotated files
  compress: true                                   # gzip rotated files
```

All keys are optional — defaults are shown above.

### Quick access

```bash
# Tail the current build log live
tail -f ~/.devopsmaestro/logs/builds/latest.log

# Open the log for a specific session
cat ~/.devopsmaestro/logs/builds/<session-uuid>.log

# List all stored build logs
ls -lh ~/.devopsmaestro/logs/builds/
```

> ⚠️ **Security notice:** Build logs may contain sensitive data — environment variable values,
> build args, tokens echoed by Makefiles, and other secrets passed to the build process.
> Log files are stored `0o600` (owner read/write only), but **do not share raw build log files**
> with others or upload them to public issue trackers without redacting secret values first.

---

## SSH Agent Forwarding

If your build needs SSH access (e.g., to clone private repos), attach with SSH agent forwarding:

```bash
dvm attach --ssh-agent
```

This makes your host SSH agent available inside the container without copying private keys.
