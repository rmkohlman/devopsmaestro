# Projects

!!! warning "Deprecation Notice"
    
    **Projects are deprecated as of v0.8.0.** The Project concept has been split into:
    
    - **Domain** - Bounded context grouping (what Project used to be conceptually)
    - **App** - The codebase/application (what Project used to store with `path`)
    
    Existing projects will be auto-migrated to Apps. See [Apps](apps.md) for the new approach.
    
    **Migration:** Your existing `dvm create project` commands will continue to work but will 
    create Apps in the default Domain with a deprecation warning.

---

Projects in dvm represent codebases on your filesystem.

> **Note:** This documentation describes the legacy Project object. For new projects, 
> please use [Apps](apps.md) instead.

---

## What is a Project?

A **project** is a pointer to a directory containing your code:

- Tracks the path to your codebase
- Can have multiple workspaces
- Stored in dvm's database

---

## Creating Projects

### From Current Directory

The most common way - create a project from where you are:

```bash
cd ~/Developer/my-app
dvm create project my-app --from-cwd
```

### From Specific Path

Point to a directory elsewhere:

```bash
dvm create project my-app --path ~/Developer/my-app
```

### With Description

Add context to remember what the project is:

```bash
dvm create project my-app --from-cwd --description "User authentication microservice"
```

---

## Managing Projects

### List All Projects

```bash
dvm get projects
# or
dvm get proj
```

Output:

```
NAME       PATH                           CREATED
● my-app   ~/Developer/my-app             2024-02-04 12:00
  api      ~/Developer/api                2024-02-04 11:30
  frontend ~/Developer/frontend           2024-02-04 10:00
```

The `●` indicates the active project.

### Get Project Details

```bash
dvm get project my-app
```

Or as YAML:

```bash
dvm get project my-app -o yaml
```

### Delete a Project

```bash
dvm delete project my-app
```

This will prompt for confirmation. Use `--force` to skip:

```bash
dvm delete project my-app --force
```

!!! warning "Deleting projects"
    
    Deleting a project also deletes all its workspaces.
    Your actual code files are **not** deleted.

---

## Setting Active Project

Set which project commands operate on by default:

```bash
dvm use project my-app
# or
dvm use proj my-app
```

Check the current context:

```bash
dvm get ctx
```

Clear the active project:

```bash
dvm use project --clear
```

---

## Multiple Projects

Manage multiple codebases:

```bash
# Add multiple projects
dvm create proj frontend --path ~/Developer/frontend
dvm create proj backend --path ~/Developer/backend
dvm create proj shared-lib --path ~/Developer/shared-lib

# Switch between them
dvm use proj backend
dvm build
dvm attach

# Later, switch to frontend
dvm use proj frontend
dvm build
dvm attach
```

---

## Best Practices

### Use Descriptive Names

```bash
# Good
dvm create proj user-service --from-cwd
dvm create proj payment-gateway --from-cwd

# Less helpful
dvm create proj app --from-cwd
dvm create proj test --from-cwd
```

### Add Descriptions

```bash
dvm create proj user-service --from-cwd \
  --description "User authentication and profile management"
```

### Organize by Team/Domain

```bash
# Team-based
dvm create proj auth-team-api --from-cwd
dvm create proj payments-team-service --from-cwd

# Domain-based
dvm create proj ecommerce-api --from-cwd
dvm create proj ecommerce-frontend --from-cwd
```

---

## Next Steps

- [Workspaces](workspaces.md) - Create environments for your projects
- [Building & Attaching](build-attach.md) - Run your containers
