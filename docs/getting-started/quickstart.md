# Quick Start

Get up and running with DevOpsMaestro in 5 minutes.

---

## Prerequisites

1. **Install dvm** - See [Installation](installation.md)
2. **Have a container runtime** - OrbStack, Docker Desktop, Podman, or Colima
3. **Have an app** - Either existing or new

---

## Object Hierarchy

DevOpsMaestro organizes your development environments in a hierarchy:

```
Ecosystem → Domain → App → Workspace
```

| Level | Purpose | Example |
|-------|---------|---------|
| **Ecosystem** | Top-level platform grouping | `my-platform` |
| **Domain** | Bounded context (group of related apps) | `backend`, `frontend` |
| **App** | A codebase/application | `my-api`, `web-app` |
| **Workspace** | Development environment for an app | `dev`, `feature-x` |

---

## 5-Minute Setup

### Step 1: Initialize dvm

Run this once to set up the database:

```bash
dvm init
```

This creates `~/.devopsmaestro/devopsmaestro.db`.

### Step 2: Create the Hierarchy

Set up the organizational structure:

```bash
# Create an ecosystem (top-level grouping)
dvm create ecosystem my-platform

# Create a domain (bounded context)
dvm create domain backend
```

> **Tip:** Each `create` command automatically sets the created resource as active, so you can continue to the next level immediately.

### Step 3: Create an App

=== "Existing App"

    ```bash
    # Go to your app directory
    cd ~/Developer/my-app
    
    # Create an app from current directory
    dvm create app my-app --from-cwd
    ```

=== "New App"

    ```bash
    # Create and enter a new directory
    mkdir ~/Developer/my-app
    cd ~/Developer/my-app
    
    # Initialize your code (example: Go app)
    git init
    go mod init github.com/myuser/my-app
    
    # Create an app from current directory
    dvm create app my-app --from-cwd
    ```

### Step 4: Create a Workspace

A workspace defines your containerized development environment:

```bash
# Create a workspace named "dev"
dvm create workspace dev
```

### Step 5: Verify Context

Check that everything is set up correctly:

```bash
dvm get context
```

You should see output like:

```
Current Context
  App:       my-app
  Workspace: (none)
```

Set your workspace as active:

```bash
dvm use workspace dev
```

### Step 6: Build the Container

```bash
dvm build
```

This will:

- Detect your app language
- Generate a Dockerfile (if needed)
- Build a container image with dev tools
- Configure Neovim with plugins

### Step 7: Attach to the Container

```bash
dvm attach
```

You're now inside your containerized dev environment! Your app is mounted and ready to edit.

---

## One-Liner Setup (After Init)

Once you understand the hierarchy, you can set up a new project quickly:

```bash
cd ~/Developer/my-app
dvm create eco my-platform && dvm create dom backend && dvm create app my-app --from-cwd && dvm create ws dev
dvm build && dvm attach
```

---

## Shorthand Commands

Use kubectl-style aliases for faster workflows:

| Resource | Full | Alias |
|----------|------|-------|
| Ecosystem | `ecosystem` | `eco` |
| Domain | `domain` | `dom` |
| App | `app` | `a`, `application` |
| Workspace | `workspace` | `ws` |
| Context | `context` | `ctx` |
| Platforms | `platforms` | `plat` |

**Example using shorthand:**

```bash
dvm create eco my-platform
dvm create dom backend
dvm create app my-app --from-cwd
dvm create ws dev
dvm use ws dev
dvm build && dvm attach
```

---

## Common Commands

```bash
# Check your current context
dvm get ctx

# List all ecosystems
dvm get eco

# List all domains (in active ecosystem)
dvm get dom

# List all apps (in active domain)
dvm get apps

# List all apps (across all domains)
dvm get apps --all

# List workspaces in current app
dvm get ws

# Check container platform status
dvm get plat

# Full status overview
dvm status

# Stop and exit the container
dvm detach
```

---

## Switching Context

Navigate your hierarchy using `dvm use`:

```bash
# Switch ecosystem
dvm use eco my-platform

# Switch domain
dvm use dom frontend

# Switch app
dvm use app web-app

# Switch workspace
dvm use ws dev

# Clear a specific level (clears levels below too)
dvm use eco none      # Clears ecosystem, domain, app
dvm use dom none      # Clears domain, app
dvm use app none      # Clears app, workspace
dvm use ws none       # Clears workspace only

# Clear all context at once
dvm use --clear
```

---

## What's Next?

- [Working with Existing Apps](existing-projects.md) - Detailed guide for adding existing repos
- [Creating New Apps](new-projects.md) - Start fresh with dvm
- [dvm Commands Reference](../dvm/commands.md) - Complete command documentation
- [nvp Quick Start](../nvp/overview.md) - Manage Neovim plugins
