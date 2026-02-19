# Hierarchical Theme System

The hierarchical theme system allows themes to cascade through the DevOpsMaestro object hierarchy, providing consistent styling while allowing for customization at any level.

---

## Theme Cascade Levels

Themes cascade through the object hierarchy from most specific to most general:

```
Workspace → App → Domain → Ecosystem → Global Default
 (most)                                    (least)
specific                                  specific
```

### Resolution Algorithm

When determining which theme to use, DevOpsMaestro follows this resolution order:

1. **Workspace-level theme** (highest priority)
2. **App-level theme**
3. **Domain-level theme**
4. **Ecosystem-level theme**
5. **Global default theme** (lowest priority)

The first theme found in this hierarchy is used.

---

## Setting Themes at Each Level

### Workspace Level

Set a theme for a specific workspace:

```bash
# Set theme for current workspace context
dvm set theme coolnight-ocean

# Set theme for specific workspace
dvm set theme gruvbox-dark --workspace my-workspace
```

### App Level

Set a theme for all workspaces in an app:

```bash
# Set theme for current app context
dvm set theme coolnight-synthwave --app

# Set theme for specific app
dvm set theme tokyonight-night --app my-api
```

### Domain Level

Set a theme for all apps in a domain:

```bash
# Set theme for current domain context
dvm set theme catppuccin-mocha --domain

# Set theme for specific domain
dvm set theme nord --domain backend
```

### Ecosystem Level

Set a theme for the entire ecosystem:

```bash
# Set theme for current ecosystem context
dvm set theme dracula --ecosystem

# Set theme for specific ecosystem
dvm set theme coolnight-matrix --ecosystem my-platform
```

### Global Default

The global default theme is used when no theme is set at any hierarchy level:

```bash
# Set global default theme (applies everywhere without override)
dvm config set theme coolnight-ocean
```

---

## Viewing Theme Cascade

### Show Current Theme

```bash
# Show active theme and where it's coming from
dvm get theme

# Example output:
# Theme: coolnight-ocean
# Source: App (my-api)
# Cascade: [none] → [none] → coolnight-ocean → [none] → [coolnight-ocean]
#          workspace   app        domain      ecosystem    global
```

### Show Cascade Effect

```bash
# Show how themes cascade through the hierarchy
dvm get theme --show-cascade

# Example output:
# Theme Cascade for workspace 'dev' in app 'my-api':
# ┌─────────────┬──────────────────┬─────────────────────┐
# │ Level       │ Theme           │ Status              │
# ├─────────────┼──────────────────┼─────────────────────┤
# │ Workspace   │ -               │ Not set             │
# │ App         │ coolnight-ocean │ ✓ Active (inherits) │
# │ Domain      │ gruvbox-dark    │ Overridden by app   │
# │ Ecosystem   │ -               │ Not set             │
# │ Global      │ tokyonight-night│ Default fallback    │
# └─────────────┴──────────────────┴─────────────────────┘
#
# Resolved theme: coolnight-ocean (from App level)
```

### Set and Show Cascade

```bash
# Set theme at app level and show cascade effect
dvm set theme gruvbox-dark --app my-api --show-cascade
```

---

## Clearing Themes

### Clear Specific Level

```bash
# Clear workspace-level theme (inherits from parent level)
dvm unset theme --workspace

# Clear app-level theme
dvm unset theme --app my-api

# Clear domain-level theme
dvm unset theme --domain backend

# Clear ecosystem-level theme
dvm unset theme --ecosystem my-platform
```

### Inherit from Parent

When you clear a theme at a level, it automatically inherits from the next level up:

```bash
# Before: Workspace has coolnight-ocean, App has gruvbox-dark
dvm get theme
# Theme: coolnight-ocean (from Workspace)

# Clear workspace theme - now inherits from app
dvm unset theme --workspace
dvm get theme
# Theme: gruvbox-dark (from App)
```

---

## Use Cases

### Per-Project Themes

Set different themes for different types of projects:

```bash
# Dark theme for backend services
dvm set theme coolnight-ocean --domain backend

# Light theme for frontend apps
dvm set theme catppuccin-latte --domain frontend

# Special theme for AI/ML projects
dvm set theme coolnight-matrix --app ml-service
```

### Team Defaults

Set consistent themes for team environments:

```bash
# Company-wide branding
dvm set theme company-theme --ecosystem corporate-platform

# Team-specific themes
dvm set theme coolnight-synthwave --domain platform-team
dvm set theme gruvbox-dark --domain data-team
```

### Personal Overrides

Override team themes for personal preference:

```bash
# Team uses gruvbox-dark at domain level
# But you prefer coolnight-ocean for your workspace
dvm set theme coolnight-ocean --workspace my-dev-env
```

### Environment-Specific Themes

Use different themes for different environments:

```bash
# Development environments use dark themes
dvm set theme coolnight-ocean --workspace dev

# Staging uses warning colors
dvm set theme coolnight-ember --workspace staging

# Production uses minimal themes
dvm set theme coolnight-mono-slate --workspace prod
```

---

## Theme Commands Reference

### Setting Themes

| Command | Scope | Example |
|---------|-------|---------|
| `dvm set theme <name>` | Current workspace | `dvm set theme coolnight-ocean` |
| `dvm set theme <name> --workspace <ws>` | Specific workspace | `dvm set theme gruvbox-dark --workspace dev` |
| `dvm set theme <name> --app [app]` | App level | `dvm set theme nord --app my-api` |
| `dvm set theme <name> --domain [domain]` | Domain level | `dvm set theme dracula --domain backend` |
| `dvm set theme <name> --ecosystem [eco]` | Ecosystem level | `dvm set theme coolnight-matrix --ecosystem platform` |

### Getting Themes

| Command | Purpose | Example |
|---------|---------|---------|
| `dvm get theme` | Show active theme | `dvm get theme` |
| `dvm get theme --show-cascade` | Show cascade effect | `dvm get theme --show-cascade` |
| `dvm get themes` | List all set themes | `dvm get themes --all-levels` |

### Clearing Themes

| Command | Purpose | Example |
|---------|---------|---------|
| `dvm unset theme` | Clear workspace theme | `dvm unset theme` |
| `dvm unset theme --app <app>` | Clear app theme | `dvm unset theme --app my-api` |
| `dvm unset theme --domain <domain>` | Clear domain theme | `dvm unset theme --domain backend` |
| `dvm unset theme --ecosystem <eco>` | Clear ecosystem theme | `dvm unset theme --ecosystem platform` |

---

## Integration with nvp

The theme system integrates seamlessly with nvp (Neovim Plugin manager):

```bash
# Set theme via dvm (with hierarchy)
dvm set theme coolnight-ocean --app my-api

# Use the resolved theme in nvp
nvp theme use $(dvm get theme --name-only)

# Generate Neovim config with hierarchical theme
nvp generate
```

---

## YAML Configuration

Themes can also be set via YAML configuration:

```yaml
# ecosystem-config.yaml
apiVersion: devopsmaestro.io/v1
kind: Ecosystem
metadata:
  name: my-platform
spec:
  theme: coolnight-ocean
---
apiVersion: devopsmaestro.io/v1
kind: Domain
metadata:
  name: backend
  ecosystem: my-platform
spec:
  theme: gruvbox-dark
---
apiVersion: devopsmaestro.io/v1
kind: App
metadata:
  name: user-service
  domain: backend
  ecosystem: my-platform
spec:
  theme: coolnight-synthwave  # Overrides domain theme
```

Apply the configuration:

```bash
dvm apply -f ecosystem-config.yaml
```

---

## Troubleshooting

### Theme Not Applying

1. **Check the cascade:**
   ```bash
   dvm get theme --show-cascade
   ```

2. **Verify theme exists:**
   ```bash
   nvp theme list | grep your-theme-name
   ```

3. **Check workspace context:**
   ```bash
   dvm get workspace
   ```

### Theme Conflicts

If you have unexpected theme behavior, clear all levels and start fresh:

```bash
# Clear all theme settings
dvm unset theme --workspace
dvm unset theme --app
dvm unset theme --domain
dvm unset theme --ecosystem

# Set new theme at desired level
dvm set theme coolnight-ocean --app
```

---

## Next Steps

- [Themes Documentation](../nvp/themes.md) - Available themes and variants
- [Commands Reference](../dvm/commands.md) - Full command documentation
- [YAML Schema](../configuration/yaml-schema.md) - YAML configuration format