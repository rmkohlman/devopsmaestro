# Quick Start: Themes

Change your workspace theme, explore built-in options, and understand how themes cascade through your development hierarchy.

---

## Overview

DevOpsMaestro ships with **34+ built-in themes** embedded directly in the binary — no installation, no network access required. You can apply any of them instantly.

**The three things you need to know:**

1. **Browse** — `dvm library get themes` lists all 34+ built-in themes
2. **Apply** — `dvm set theme <name>` sets a theme (global by default, or scoped with flags)
3. **Inspect** — `dvm get nvim themes` shows all themes in use; `dvm get nvim theme <name>` shows details

---

## Step 1: See What Themes Are Available

```bash
# List all built-in themes (34+)
dvm library get themes

# Short alias
dvm lib ls nt
```

Example output:

```
NAME                      DESCRIPTION
coolnight-ocean           Default blue ocean theme
coolnight-arctic          Ice-cold blue variant
coolnight-synthwave       Cyberpunk purple aesthetic
coolnight-matrix          Matrix-inspired green
tokyonight-night          Tokyo Night — dark blue
catppuccin-mocha          Catppuccin mocha flavor
gruvbox-dark              Warm retro dark theme
...
```

See the full list with descriptions in the [Themes Reference](../nvp/themes.md).

---

## Step 2: View a Theme's Details

```bash
# Show theme details
dvm library describe theme coolnight-ocean
```

Output:

```
Name:         coolnight-ocean
Description:  Default blue ocean theme
Author:       devopsmaestro
Category:     dark
Style:        ocean
Repository:   rmkohlman/coolnight.nvim
```

---

## Step 3: Set Your Theme

### Set a Global Default

No scope flags = global default. Every workspace inherits this unless overridden.

```bash
dvm set theme coolnight-ocean
```

### Set for a Specific Workspace

```bash
dvm set theme coolnight-synthwave --workspace dev
# or short form
dvm set theme coolnight-synthwave -w dev
```

### Set for an Entire App

All workspaces in the app inherit this theme.

```bash
dvm set theme tokyonight-night --app my-api
# or short form
dvm set theme tokyonight-night -a my-api
```

### Set at Domain or Ecosystem Level

```bash
# Domain level (all apps in that domain)
dvm set theme gruvbox-dark --domain backend

# Ecosystem level (everything in that ecosystem)
dvm set theme catppuccin-mocha --ecosystem my-platform
```

---

## Step 4: See What Theme Is Active

```bash
# Show effective theme for current context with resolution path
dvm get nvim themes

# Show details for a specific theme
dvm get nvim theme coolnight-ocean
```

The `dvm get nvim themes` command shows both built-in (library) and user-defined themes with color swatches when your terminal supports true color.

---

## Step 5: Apply a Theme via YAML (IaC)

Themes are full YAML resources. You can apply them declaratively:

```bash
# Apply from a local file
dvm apply -f my-theme.yaml

# Apply from a URL
dvm apply -f https://example.com/themes/my-theme.yaml

# Apply from GitHub
dvm apply -f github:user/repo/themes/my-theme.yaml
```

---

## Clearing a Theme Override

Pass an empty string to clear the theme at a level, inheriting from the parent:

```bash
# Clear workspace override (inherits from app)
dvm set theme "" --workspace dev

# Clear app override (inherits from domain)
dvm set theme "" --app my-api

# Clear global default (falls back to hardcoded default: coolnight-ocean)
dvm set theme "" --global
```

---

## Preview the Cascade Effect

See how a theme propagates through your hierarchy before applying:

```bash
# Set theme and show the cascade tree
dvm set theme coolnight-synthwave --app my-api --show-cascade
```

Example output:

```
global          → coolnight-ocean
└─ dev-platform → (inherit from global)
   └─ backend   → (inherit from ecosystem)
      └─ my-api → coolnight-synthwave ← SET HERE
         └─ dev → (inherit from app)
```

---

## Common Workflows

### Change Theme for a Single Session (workspace only)

```bash
dvm set theme coolnight-matrix -w dev
dvm build dev --app my-api   # Rebuild to apply
dvm attach
```

### Set a Company-Wide Default

```bash
# All workspaces use this unless overridden at a lower level
dvm set theme coolnight-ocean --ecosystem my-company
```

### Try a Theme Without Committing

```bash
# Dry-run shows what would change
dvm set theme dracula --workspace dev --dry-run
```

### Use Different Themes per Domain

```bash
# Security team: high-contrast
dvm set theme coolnight-matrix --domain security

# Frontend team: modern feel
dvm set theme catppuccin-mocha --domain frontend

# Data team: warm and comfortable
dvm set theme gruvbox-dark --domain data
```

### Export Current Theme to Share with Team

```bash
dvm get nvim theme coolnight-synthwave -o yaml > team-theme.yaml

# Team members apply it
dvm apply -f team-theme.yaml
```

---

## Theme Hierarchy Summary

Themes cascade from most specific to least specific:

```
Workspace → App → Domain → Ecosystem → Global Default
```

The first theme found walking up the hierarchy is used. Setting a theme at a higher level affects all children that haven't set their own override.

| Scope Flag | What It Sets |
|------------|-------------|
| `--workspace/-w <name>` | One workspace |
| `--app/-a <name>` | All workspaces in that app |
| `--domain/-d <name>` | All apps and workspaces in that domain |
| `--ecosystem/-e <name>` | Everything in that ecosystem |
| `--global` | Global fallback for everything |
| *(no flags)* | Global fallback (same as `--global`) |

---

## Next Steps

- **[All Built-in Themes](themes.md)** — Complete theme catalog with descriptions
- **[Built-in Packages](packages.md)** — Nvim plugin packages reference
- **[Theme Hierarchy](../advanced/theme-hierarchy.md)** — Deep dive into cascade rules
- **[NvimTheme YAML Reference](../reference/nvim-theme.md)** — Define custom themes
- **[CoolNight Collection](coolnight.md)** — All 21 CoolNight variants explained
