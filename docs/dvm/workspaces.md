# Workspaces

Workspaces are isolated container environments for your apps.

---

## What is a Workspace?

A **Workspace** is a containerized development environment:

- Belongs to an App (your codebase)
- Runs in a Docker container
- Mounts your app's source code directory
- Has its own image with tools/dependencies
- Can be customized with different configurations

Think of it as your App running in **dev mode**.

---

## Creating Workspaces

### Basic Workspace

Create a workspace for the active app:

```bash
dvm create workspace dev
# or
dvm create ws dev
```

### In a Specific App

```bash
dvm create ws dev -a my-api
```

### With Description

```bash
dvm create ws dev --description "Main development environment"
```

### With Custom Image Name

```bash
dvm create ws dev --image my-custom-image:latest
```

---

## Managing Workspaces

### List Workspaces

```bash
dvm get workspaces
# or
dvm get ws
```

Output:

```
NAME          APP        IMAGE                           STATUS   CREATED
● dev         my-app     dvm-dev-my-app:20240204-1200    ready    2024-02-04
  test        my-app     dvm-test-my-app:pending         pending  2024-02-04
  feature-x   my-app     dvm-feature-x-my-app:pending    pending  2024-02-04
```

### Get Workspace Details

```bash
dvm get workspace dev
dvm get ws dev -o yaml
```

### Delete a Workspace

```bash
dvm delete workspace dev
dvm delete ws dev --force
```

---

## Setting Active Workspace

Set which workspace commands operate on:

```bash
dvm use workspace dev
# or
dvm use ws dev
```

Check current context:

```bash
dvm get ctx
# Ecosystem: default
# Domain:    default  
# App:       my-app
# Workspace: dev
```

Clear active workspace:

```bash
dvm use ws none
```

---

## Multiple Workspaces

Create different workspaces for different purposes:

### Development Workspace

Your main coding environment:

```bash
dvm create ws dev --description "Daily development"
```

### Testing Workspace

For running test suites:

```bash
dvm create ws test --description "Running tests"
```

### Feature Branch Workspaces

Isolated environments for features:

```bash
dvm create ws feature-auth --description "Authentication feature"
dvm create ws feature-payments --description "Payment integration"
```

### Debug Workspace

With extra debugging tools:

```bash
dvm create ws debug --description "Debugging with extra tools"
```

---

## Workspace Lifecycle

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   create    │────▶│    build    │────▶│   attach    │
│  workspace  │     │   (image)   │     │ (container) │
└─────────────┘     └─────────────┘     └─────────────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │   detach    │
                                        │   (stop)    │
                                        └─────────────┘
```

1. **Create** - Define the workspace
2. **Build** - Build the container image
3. **Attach** - Start and enter the container
4. **Detach** - Stop the container

---

## Switching Workspaces

```bash
# Start with dev
dvm use ws dev
dvm build
dvm attach
# ... work on code ...
# Exit container (Ctrl+D or exit)

# Switch to test
dvm use ws test
dvm build
dvm attach
# ... run tests ...

# Switch to feature branch
dvm use ws feature-auth
dvm build
dvm attach
```

---

## Image Versioning

Workspace images are tagged with timestamps:

```
dvm-dev-my-app:20240204-120030
```

Format: `dvm-<workspace>-<app>:<YYYYMMDD-HHMMSS>`

When you run `dvm attach`:

- If the image changed since last attach, container is recreated
- Your code changes persist (they're mounted, not in the image)

---

## Best Practices

### Use Purpose-Specific Workspaces

```bash
# Good
dvm create ws dev
dvm create ws test
dvm create ws staging

# Less ideal - one workspace for everything
dvm create ws main
```

### Add Descriptions

```bash
dvm create ws dev --description "Daily development with hot reload"
dvm create ws ci --description "Mimics CI environment"
```

### Clean Up Unused Workspaces

```bash
# List all
dvm get ws

# Delete old feature workspaces
dvm delete ws feature-old --force
```

---

## Next Steps

- [Apps](apps.md) - Understand the App object
- [Building & Attaching](build-attach.md) - Container lifecycle details
- [Commands Reference](commands.md) - Full command documentation
