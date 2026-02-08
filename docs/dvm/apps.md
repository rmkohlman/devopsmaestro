# Apps

Apps in dvm represent your codebases/applications - the things you build and run.

---

## What is an App?

An **App** is the core object in DevOpsMaestro - it represents your codebase:

- Has a `path` pointing to your source code on disk
- Can have multiple Workspaces (development environments)
- Belongs to a Domain (bounded context)
- Can run in **dev mode** (Workspace) or **live mode** (Operator)

### Hierarchy

```
Ecosystem → Domain → App → Workspace
```

| Level | Purpose | Example |
|-------|---------|---------|
| Ecosystem | Platform grouping | `customer-platform` |
| Domain | Bounded context | `auth-domain` |
| App | Your codebase | `auth-api` |
| Workspace | Dev environment | `dev`, `test` |

---

## Creating Apps

### From Current Directory

The most common way - create an app from where your code is:

```bash
cd ~/Developer/my-app
dvm create app my-app --from-cwd
```

### From Specific Path

Point to a directory elsewhere:

```bash
dvm create app my-app --path ~/Developer/my-app
```

### In a Specific Domain

```bash
dvm create app my-app --from-cwd -d auth-domain
```

### With Description

Add context to remember what the app is:

```bash
dvm create app my-app --from-cwd --description "User authentication microservice"
```

---

## Managing Apps

### List All Apps

```bash
dvm get apps
# or
dvm get app
```

Output:

```
NAME       DOMAIN        PATH                           CREATED
● my-app   auth-domain   ~/Developer/my-app             2024-02-04 12:00
  api      user-domain   ~/Developer/api                2024-02-04 11:30
  frontend frontend      ~/Developer/frontend           2024-02-04 10:00
```

The `●` indicates the active app.

### Get App Details

```bash
dvm get app my-app
```

Or as YAML:

```bash
dvm get app my-app -o yaml
```

### Delete an App

```bash
dvm delete app my-app
```

This will prompt for confirmation. Use `--force` to skip:

```bash
dvm delete app my-app --force
```

!!! warning "Deleting apps"
    
    Deleting an app also deletes all its workspaces.
    Your actual code files are **not** deleted.

---

## Setting Active App

Set which app commands operate on by default:

```bash
dvm use app my-app
```

Check the current context:

```bash
dvm get ctx
```

Clear the active app:

```bash
dvm use app --clear
```

---

## Apps and Domains

Apps belong to Domains. If you don't specify a domain, the default domain is used.

### Create Domain First (Optional)

```bash
dvm create domain auth-domain -e customer-platform
dvm use domain auth-domain
```

### Create App in Domain

```bash
dvm create app auth-api --from-cwd -d auth-domain
```

### List Apps in Current Domain

```bash
dvm get apps
```

### List Apps Across All Domains

```bash
dvm get apps -A
```

---

## Apps and Workspaces

Each App can have multiple Workspaces (development environments):

```bash
# Create app
dvm create app auth-api --from-cwd

# Create workspaces for the app
dvm use app auth-api
dvm create workspace dev --description "Daily development"
dvm create workspace test --description "Running tests"
dvm create workspace debug --description "Debugging with extra tools"

# Work in a workspace
dvm use workspace dev
dvm build
dvm attach
```

---

## Dev Mode vs Live Mode

Apps can run in two modes:

| Mode | How | Description |
|------|-----|-------------|
| **Dev mode** | `dvm attach` | Interactive workspace with nvim, tools, mounted source |
| **Live mode** | Operator | Production-like deployment (requires k8s) |

### Dev Mode (Workspace)

```bash
dvm use app auth-api
dvm use workspace dev
dvm build
dvm attach  # Enter the container, code with nvim
```

### Live Mode (Future - requires Operator)

```bash
# Once Operator is installed
dvm deploy app auth-api --mode live
```

---

## Best Practices

### Use Descriptive Names

```bash
# Good
dvm create app user-service --from-cwd
dvm create app payment-gateway --from-cwd

# Less helpful
dvm create app app --from-cwd
dvm create app test --from-cwd
```

### Add Descriptions

```bash
dvm create app user-service --from-cwd \
  --description "User authentication and profile management"
```

### Organize by Domain

```bash
# Create domains for your bounded contexts
dvm create domain auth-domain
dvm create domain payments-domain
dvm create domain frontend-domain

# Create apps in appropriate domains
dvm create app auth-api --from-cwd -d auth-domain
dvm create app payment-api --from-cwd -d payments-domain
dvm create app web-app --from-cwd -d frontend-domain
```

---

## Migration from Projects

If you have existing Projects, they will be auto-migrated to Apps:

- Each Project becomes an App with the same name
- The `path` is preserved
- All Workspaces are re-associated with the new App
- A default Domain is created if none exists

You can continue using `dvm create project` but will see deprecation warnings.

---

## Next Steps

- [Workspaces](workspaces.md) - Create development environments for your apps
- [Building & Attaching](build-attach.md) - Run your containers
