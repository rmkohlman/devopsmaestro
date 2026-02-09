# Adding Existing Apps

Already have a codebase on your laptop? Here's how to add it to DevOpsMaestro.

---

## Overview

When you add an existing app to dvm:

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

### 2. Navigate to Your App

```bash
cd ~/Developer/my-existing-app
```

### 3. Create the App in dvm

Use `--from-cwd` to use the current directory:

```bash
dvm create app my-app --from-cwd
```

Or specify the path explicitly:

```bash
dvm create app my-app --path ~/Developer/my-existing-app
```

You can also add a description:

```bash
dvm create app my-app --from-cwd --description "My REST API backend"
```

### 4. Set as Active App

```bash
dvm use app my-app
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

- Auto-detect your app language (Go, Python, Node.js, etc.)
- Generate an appropriate Dockerfile
- Install language tools and LSP servers
- Configure Neovim with relevant plugins

### 7. Attach to Your Environment

```bash
dvm attach
```

You're now inside the container with your app mounted!

---

## Full Example

Here's a complete example adding a Go app:

```bash
# Your existing Go app
cd ~/Developer/my-go-api

# Verify the app exists
ls -la
# go.mod  go.sum  main.go  ...

# Initialize dvm (skip if already done)
dvm init

# Add app to dvm
dvm create app my-go-api --from-cwd --description "Go REST API"

# Set context
dvm use app my-go-api
dvm use workspace dev   # This will error - we need to create it first

# Create and use workspace
dvm create workspace dev
dvm use workspace dev

# Check your setup
dvm get ctx
# App:       my-go-api
# Workspace: dev

# Build the container
dvm build

# Enter the container
dvm attach

# Inside container: your app is at /workspace
# Neovim is configured with Go LSP, gopls, etc.
```

---

## Adding Multiple Apps

You can manage multiple apps:

```bash
# Add first app
cd ~/Developer/frontend-app
dvm create app frontend --from-cwd

# Add second app
cd ~/Developer/backend-api
dvm create app backend --from-cwd

# Add third app
cd ~/Developer/shared-lib
dvm create app shared --from-cwd

# List all apps
dvm get apps
# NAME       PATH                           CREATED
# frontend   ~/Developer/frontend-app       2024-02-04 12:00
# backend    ~/Developer/backend-api        2024-02-04 12:01
# shared     ~/Developer/shared-lib         2024-02-04 12:02

# Switch between apps
dvm use app backend
dvm get ctx
```

---

## Multiple Workspaces per App

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
# NAME          APP        IMAGE                    STATUS   CREATED
# dev           my-app     dvm-dev-my-app:20240204  ready    2024-02-04
# test          my-app     dvm-test-my-app:pending  pending  2024-02-04
# feature-auth  my-app     dvm-feature-auth:pending pending  2024-02-04

# Switch workspaces
dvm use ws test
dvm build  # Build this workspace's container
```

---

## Verify Your Setup

After adding an app:

```bash
# Check current context
dvm get ctx

# View app details
dvm get app my-app -o yaml

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
    
    - dvm **does not modify** your app files
    - Your code is **mounted** into containers (not copied)
    - Changes made inside the container **persist** on your filesystem
    - Git, IDE access, etc. all work normally outside the container

---

## Next Steps

- [Building & Attaching](../dvm/build-attach.md) - Learn more about containers
- [Workspace Configuration](../configuration/yaml-schema.md) - Customize your environment
- [nvp Plugins](../nvp/plugins.md) - Configure Neovim plugins
