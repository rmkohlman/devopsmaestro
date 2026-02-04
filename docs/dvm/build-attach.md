# Building & Attaching

How to build container images and work inside them.

---

## Building Images

### Basic Build

Build the container for the active workspace:

```bash
dvm build
```

This will:

1. Detect your project language
2. Generate a Dockerfile (if needed)
3. Build the container image
4. Install language tools and LSP servers
5. Configure Neovim

### Force Rebuild

Rebuild even if the image exists:

```bash
dvm build --force
```

### Build Without Cache

Start fresh without Docker cache:

```bash
dvm build --no-cache
```

---

## What Gets Built?

dvm auto-detects your project and sets up:

| Language | Tools Installed |
|----------|-----------------|
| **Go** | go, gopls, golangci-lint |
| **Python** | python, pip, pyright, black |
| **Node.js** | node, npm, typescript-language-server |
| **Rust** | rust, cargo, rust-analyzer |

Plus common tools:

- Neovim with plugins
- git, curl, wget
- ripgrep, fd-find
- tmux

---

## Attaching to Containers

### Basic Attach

Enter the active workspace container:

```bash
dvm attach
```

### Attach to Specific Workspace

```bash
dvm attach dev
dvm attach test
```

---

## What Happens on Attach?

1. **Container starts** (if not running)
2. **Project mounted** at `/workspace`
3. **Terminal connects** with full resize support
4. **You're inside!** Ready to code

```
┌────────────────────────────────────────────────────┐
│  Your Terminal                                      │
│  ┌──────────────────────────────────────────────┐  │
│  │  Container                                    │  │
│  │  /workspace $ nvim .                         │  │
│  │                                              │  │
│  │  Your project files are here                 │  │
│  │  Changes sync automatically                  │  │
│  └──────────────────────────────────────────────┘  │
└────────────────────────────────────────────────────┘
```

---

## Working Inside the Container

Once attached:

```bash
# Your project is at /workspace
cd /workspace
ls

# Edit with Neovim (fully configured!)
nvim main.go

# Run your code
go run .
python main.py
npm start

# Use git
git status
git commit -m "feat: add feature"

# Exit when done
exit
# or Ctrl+D
```

---

## Detaching

### Exit the Container

Just type `exit` or press `Ctrl+D`.

### Stop the Container

Stop the container from outside:

```bash
dvm detach
```

Or stop a specific workspace:

```bash
dvm detach dev
```

---

## Auto-Recreate on Image Change

When you run `dvm attach`:

1. dvm checks if the image has changed
2. If yes, the container is recreated
3. Your code is still there (it's mounted)

This means after `dvm build --force`, the next `dvm attach` will use the new image automatically.

---

## Terminal Resize

The container terminal supports full resize:

- Resize your terminal window
- The container terminal adjusts automatically
- Works with Neovim splits, tmux, etc.

---

## Container Platforms

dvm detects and uses available platforms:

```bash
dvm get platforms
```

| Platform | Status |
|----------|--------|
| OrbStack | Recommended for macOS |
| Docker Desktop | Cross-platform |
| Podman | Rootless option |
| Colima | Lightweight |

### Force a Specific Platform

```bash
DVM_PLATFORM=colima dvm build
DVM_PLATFORM=docker dvm attach
```

---

## Troubleshooting

### Image Not Building

```bash
# Check platform detection
dvm get plat

# Try verbose mode
dvm build -v

# Check logs
dvm build --log-file build.log
```

### Container Won't Start

```bash
# Check status
dvm status

# Verbose attach
dvm attach -v
```

### Changes Not Persisting

Your project is **mounted**, so changes should persist. If they don't:

1. Make sure you're editing in `/workspace`
2. Check that the mount is working: `ls /workspace`

---

## Next Steps

- [Commands Reference](commands.md) - Full command list
- [YAML Configuration](../configuration/yaml-schema.md) - Customize workspaces
- [nvp Plugins](../nvp/plugins.md) - Configure Neovim
