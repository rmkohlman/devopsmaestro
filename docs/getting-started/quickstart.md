# Quick Start

Get up and running with DevOpsMaestro in 5 minutes.

---

## Prerequisites

1. **Install dvm** - See [Installation](installation.md)
2. **Have a container runtime** - OrbStack, Docker Desktop, Podman, or Colima
3. **Have a project** - Either existing or new

---

## 5-Minute Setup

### Step 1: Initialize dvm

Run this once to set up the database:

```bash
dvm init
```

This creates `~/.devopsmaestro/devopsmaestro.db`.

### Step 2: Create a Project

=== "Existing Project"

    ```bash
    # Go to your project directory
    cd ~/Developer/my-app
    
    # Create a project from current directory
    dvm create project my-app --from-cwd
    ```

=== "New Project"

    ```bash
    # Create and enter a new directory
    mkdir ~/Developer/my-app
    cd ~/Developer/my-app
    
    # Initialize your code (example: Go project)
    git init
    go mod init github.com/myuser/my-app
    
    # Create a project from current directory
    dvm create project my-app --from-cwd
    ```

### Step 3: Set Active Context

```bash
# Set the active project
dvm use project my-app

# Verify
dvm get ctx
```

### Step 4: Create a Workspace

A workspace defines your containerized development environment:

```bash
# Create a workspace named "dev"
dvm create workspace dev

# Set it as active
dvm use workspace dev
```

### Step 5: Build the Container

```bash
dvm build
```

This will:

- Detect your project language
- Generate a Dockerfile (if needed)
- Build a container image with dev tools
- Configure Neovim with plugins

### Step 6: Attach to the Container

```bash
dvm attach
```

You're now inside your containerized dev environment! Your project is mounted and ready to edit.

---

## Shorthand Commands

Use kubectl-style aliases for faster workflows:

| Full Command | Shorthand |
|--------------|-----------|
| `dvm get projects` | `dvm get proj` |
| `dvm get workspaces` | `dvm get ws` |
| `dvm get context` | `dvm get ctx` |
| `dvm create project` | `dvm create proj` |
| `dvm create workspace` | `dvm create ws` |
| `dvm use project` | `dvm use proj` |
| `dvm use workspace` | `dvm use ws` |

**Example using shorthand:**

```bash
dvm create proj my-app --from-cwd
dvm use proj my-app
dvm create ws dev
dvm use ws dev
dvm build
dvm attach
```

---

## Common Commands

```bash
# Check your current context
dvm get ctx

# List all projects
dvm get proj

# List workspaces in current project
dvm get ws

# Check container platform status
dvm get plat

# Full status overview
dvm status

# Stop and exit the container
dvm detach
```

---

## What's Next?

- [Working with Existing Projects](existing-projects.md) - Detailed guide for adding existing repos
- [Creating New Projects](new-projects.md) - Start fresh with dvm
- [dvm Commands Reference](../dvm/commands.md) - Complete command documentation
- [nvp Quick Start](../nvp/overview.md) - Manage Neovim plugins
