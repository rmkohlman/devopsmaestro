# Adding Existing Projects

Already have a project on your laptop? Here's how to add it to DevOpsMaestro.

---

## Overview

When you add an existing project to dvm:

1. **Your code stays where it is** - dvm just tracks the path
2. **Code is mounted into containers** - Changes sync automatically
3. **You get isolated environments** - Each workspace is a separate container

---

## Step-by-Step Guide

### 1. Initialize dvm (one-time)

If you haven't already:

```bash
dvm init
```

### 2. Navigate to Your Project

```bash
cd ~/Developer/my-existing-app
```

### 3. Create the Project in dvm

Use `--from-cwd` to use the current directory:

```bash
dvm create project my-app --from-cwd
```

Or specify the path explicitly:

```bash
dvm create project my-app --path ~/Developer/my-existing-app
```

You can also add a description:

```bash
dvm create project my-app --from-cwd --description "My REST API backend"
```

### 4. Set as Active Project

```bash
dvm use project my-app
```

### 5. Create a Workspace

Workspaces are your containerized environments. You might have different ones for different purposes:

```bash
# Main development workspace
dvm create workspace dev

# Set it as active
dvm use workspace dev
```

### 6. Build the Container

```bash
dvm build
```

dvm will:

- Auto-detect your project language (Go, Python, Node.js, etc.)
- Generate an appropriate Dockerfile
- Install language tools and LSP servers
- Configure Neovim with relevant plugins

### 7. Attach to Your Environment

```bash
dvm attach
```

You're now inside the container with your project mounted!

---

## Full Example

Here's a complete example adding a Go project:

```bash
# Your existing Go project
cd ~/Developer/my-go-api

# Verify the project exists
ls -la
# go.mod  go.sum  main.go  ...

# Initialize dvm (skip if already done)
dvm init

# Add project to dvm
dvm create project my-go-api --from-cwd --description "Go REST API"

# Set context
dvm use project my-go-api
dvm use workspace dev   # This will error - we need to create it first

# Create and use workspace
dvm create workspace dev
dvm use workspace dev

# Check your setup
dvm get ctx
# Project:   my-go-api
# Workspace: dev

# Build the container
dvm build

# Enter the container
dvm attach

# Inside container: your project is at /workspace
# Neovim is configured with Go LSP, gopls, etc.
```

---

## Adding Multiple Projects

You can manage multiple projects:

```bash
# Add first project
cd ~/Developer/frontend-app
dvm create proj frontend --from-cwd

# Add second project
cd ~/Developer/backend-api
dvm create proj backend --from-cwd

# Add third project
cd ~/Developer/shared-lib
dvm create proj shared --from-cwd

# List all projects
dvm get proj
# NAME       PATH                           CREATED
# frontend   ~/Developer/frontend-app       2024-02-04 12:00
# backend    ~/Developer/backend-api        2024-02-04 12:01
# shared     ~/Developer/shared-lib         2024-02-04 12:02

# Switch between projects
dvm use proj backend
dvm get ctx
```

---

## Multiple Workspaces per Project

Create different workspaces for different purposes:

```bash
# Development workspace
dvm create ws dev

# Testing workspace with different config
dvm create ws test --description "For running tests"

# Feature branch workspace
dvm create ws feature-auth --description "Auth feature development"

# List workspaces
dvm get ws
# NAME          PROJECT    IMAGE                    STATUS   CREATED
# dev           my-app     dvm-dev-my-app:20240204  ready    2024-02-04
# test          my-app     dvm-test-my-app:pending  pending  2024-02-04
# feature-auth  my-app     dvm-feature-auth:pending pending  2024-02-04

# Switch workspaces
dvm use ws test
dvm build  # Build this workspace's container
```

---

## Verify Your Setup

After adding a project:

```bash
# Check current context
dvm get ctx

# View project details
dvm get project my-app -o yaml

# List all workspaces
dvm get ws

# Check container platform
dvm get plat

# Full status
dvm status
```

---

## What Happens to My Code?

!!! info "Your code is safe"
    
    - dvm **does not modify** your project files
    - Your code is **mounted** into containers (not copied)
    - Changes made inside the container **persist** on your filesystem
    - Git, IDE access, etc. all work normally outside the container

---

## Next Steps

- [Building & Attaching](../dvm/build-attach.md) - Learn more about containers
- [Workspace Configuration](../configuration/yaml-schema.md) - Customize your environment
- [nvp Plugins](../nvp/plugins.md) - Configure Neovim plugins
