# Theme Infrastructure as Code

DevOpsMaestro supports Infrastructure as Code (IaC) workflows for themes, allowing you to version, share, and collaborate on theme configurations using YAML files.

---

## Overview

Theme IaC enables:

- **Version control** - Track theme changes in Git
- **Team sharing** - Share themes across your organization
- **Declarative configuration** - Define themes as code
- **Remote sourcing** - Apply themes from URLs and GitHub
- **Backup and restore** - Export/import theme configurations

---

## Basic IaC Operations

### Apply Theme from File

```bash
# Apply theme from local file
dvm apply -f my-theme.yaml

# Apply from URL
dvm apply -f https://themes.example.com/coolnight-variants.yaml

# Apply from GitHub (shorthand)
dvm apply -f github:user/repo/themes/company-theme.yaml

# Apply from stdin
cat theme.yaml | dvm apply -f -
```

### Export Theme for Sharing

```bash
# Export current workspace theme
dvm get nvim theme -o yaml > my-workspace-theme.yaml

# Export specific theme by name
dvm get nvim theme coolnight-ocean -o yaml > coolnight-ocean.yaml

# Export theme with hierarchy context
dvm get nvim theme --with-hierarchy -o yaml > full-theme-config.yaml
```

---

## Theme YAML Format

### Basic Theme Definition

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: my-custom-theme
  description: Corporate brand theme
  author: Platform Team
  version: "1.0.0"
  category: dark
  tags:
    - corporate
    - blue
    - professional
spec:
  plugin:
    repo: folke/tokyonight.nvim
    version: "v2.0.0"               # Optional: pin to specific version
  style: night                      # Theme variant
  transparent: false                # Transparent background
  colors:
    bg: "#1a1b26"
    fg: "#c0caf5"
    accent: "#7aa2f7"
    comment: "#565f89"
    keyword: "#bb9af7"
    string: "#9ece6a"
    function: "#7aa2f7"
    variable: "#c0caf5"
    type: "#2ac3de"
    constant: "#ff9e64"
    error: "#f7768e"
    warning: "#e0af68"
    info: "#7dcfff"
    hint: "#1abc9c"
    selection: "#33467c"
    border: "#29a4bd"
  options:                          # Plugin-specific options
    dim_inactive: true
    terminal_colors: true
    styles:
      comments: italic
      keywords: bold
      functions: none
      variables: none
```

### CoolNight Parametric Theme

```yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: company-blue
  description: Company-branded blue theme
  author: Design Team
spec:
  generator: coolnight
  parameters:
    hue: 210                        # Blue hue
    name: company-blue
    base_colors:
      bg: "#0a0e1a"                 # Custom background
    overrides:
      accent: "#0078d4"             # Company blue
      keyword: "#569cd6"            # Lighter company blue
```

---

## Hierarchy-Aware Theme IaC

### Multi-Level Theme Configuration

Define themes at multiple hierarchy levels in a single file:

```yaml
# team-themes.yaml
apiVersion: devopsmaestro.io/v1
kind: ThemeConfig
metadata:
  name: platform-team-themes
  description: Standardized themes for platform team
spec:
  hierarchy:
    ecosystem:
      name: corporate-platform
      theme: company-base
    domains:
      - name: backend-services
        theme: company-dark
      - name: frontend-apps  
        theme: company-light
    apps:
      - name: user-service
        domain: backend-services
        theme: company-blue          # Override domain theme
      - name: payment-service
        domain: backend-services
        theme: company-green         # Different override
---
# Define the referenced themes
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: company-base
spec:
  plugin:
    repo: folke/tokyonight.nvim
  style: night
  # ... theme definition
---
apiVersion: devopsmaestro.io/v1  
kind: NvimTheme
metadata:
  name: company-dark
spec:
  generator: coolnight
  parameters:
    hue: 220
    name: company-dark
---
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: company-light
spec:
  plugin:
    repo: catppuccin/nvim
  style: latte
  # ... light theme definition
```

Apply the multi-level configuration:

```bash
dvm apply -f team-themes.yaml
```

---

## Remote Theme Sources

### URL Sources

Apply themes directly from web URLs:

```bash
# Apply from theme library
dvm apply -f https://themes.devopsmaestro.io/coolnight-variants.yaml

# Apply from company theme repository
dvm apply -f https://themes.company.com/corporate-brand.yaml

# Apply multiple themes
dvm apply -f https://themes.company.com/base-themes.yaml \
          -f https://themes.company.com/overrides.yaml
```

### GitHub Shorthand

Use GitHub shorthand for easier access:

```bash
# Public repository
dvm apply -f github:platform-team/themes/backend-theme.yaml

# Specific branch
dvm apply -f github:platform-team/themes/experimental-theme.yaml@feature-branch

# Specific tag/version
dvm apply -f github:platform-team/themes/stable-theme.yaml@v1.2.0

# Raw GitHub URLs are also supported
dvm apply -f https://raw.githubusercontent.com/platform-team/themes/main/base.yaml
```

### Private Repository Access

For private repositories, ensure your Git credentials are configured:

```bash
# SSH key authentication (recommended)
dvm apply -f github:company/private-themes/secret-theme.yaml

# Personal Access Token (via Git configuration)
git config --global credential.helper store
dvm apply -f github:company/private-themes/secret-theme.yaml
```

---

## Team Sharing Workflows

### 1. Centralized Theme Repository

Create a dedicated repository for team themes:

```bash
# Repository structure
themes/
├── base/
│   ├── corporate-dark.yaml
│   ├── corporate-light.yaml
│   └── corporate-minimal.yaml
├── teams/
│   ├── backend/
│   │   ├── service-theme.yaml
│   │   └── api-theme.yaml
│   └── frontend/
│       ├── webapp-theme.yaml
│       └── mobile-theme.yaml
├── environments/
│   ├── development.yaml
│   ├── staging.yaml
│   └── production.yaml
└── README.md
```

Team members apply themes:

```bash
# Apply base corporate theme
dvm apply -f github:company/themes/base/corporate-dark.yaml

# Apply team-specific customizations
dvm apply -f github:company/themes/teams/backend/service-theme.yaml
```

### 2. Project-Based Themes

Include theme configuration in project repositories:

```bash
# Project structure
my-service/
├── .devopsmaestro/
│   ├── workspace.yaml
│   └── theme.yaml              # Project theme
├── src/
└── README.md
```

Apply project theme when setting up workspace:

```bash
cd my-service
dvm apply -f .devopsmaestro/theme.yaml
dvm apply -f .devopsmaestro/workspace.yaml
```

### 3. Environment-Specific Themes

Use different themes for different environments:

```bash
# development-theme.yaml - High contrast for development
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: dev-theme
spec:
  generator: coolnight
  parameters:
    hue: 210
    name: dev-bright
```

```bash
# production-theme.yaml - Subdued colors for production work
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: prod-theme  
spec:
  generator: coolnight
  parameters:
    hue: 0
    name: prod-calm
    base_colors:
      bg: "#0f0f0f"               # Very dark background
```

Apply based on environment:

```bash
# Development
dvm apply -f github:company/themes/environments/development.yaml

# Production
dvm apply -f github:company/themes/environments/production.yaml
```

---

## Backup and Restore

### Export Current Configuration

```bash
# Export all themes and hierarchy settings
dvm get themes --all-levels -o yaml > devopsmaestro-backup-$(date +%Y%m%d).yaml

# Export specific ecosystem configuration
dvm get ecosystem my-platform --include-themes -o yaml > ecosystem-backup.yaml

# Export only theme definitions (no hierarchy)
dvm get nvim themes -o yaml > themes-only.yaml
```

### Restore from Backup

```bash
# Restore full configuration
dvm apply -f devopsmaestro-backup-20260219.yaml

# Restore only themes
dvm apply -f themes-only.yaml
```

---

## Version Control Integration

### Git Workflow

```bash
# 1. Create theme repository
mkdir company-themes && cd company-themes
git init

# 2. Export current themes
dvm get nvim themes -o yaml > base-themes.yaml

# 3. Commit and push
git add base-themes.yaml
git commit -m "Add base company themes"
git remote add origin git@github.com:company/themes.git
git push -u origin main

# 4. Team members apply themes
dvm apply -f github:company/themes/base-themes.yaml
```

### Theme Versioning

Use Git tags for theme versioning:

```bash
# Tag a stable theme version
git tag -a v1.0.0 -m "Stable theme release v1.0.0"
git push origin v1.0.0

# Apply specific version
dvm apply -f github:company/themes/base-themes.yaml@v1.0.0
```

### Branching Strategy

```bash
# Development branch for theme experiments
git checkout -b feature/new-corporate-theme

# Edit themes
vim corporate-blue-v2.yaml

# Test theme
dvm apply -f corporate-blue-v2.yaml
nvp generate

# Commit and create PR
git add corporate-blue-v2.yaml
git commit -m "Add corporate blue v2 theme"
git push origin feature/new-corporate-theme
```

---

## Validation and Testing

### Theme Validation

DevOpsMaestro automatically validates theme YAML files:

```bash
# This will validate the YAML and report errors
dvm apply -f invalid-theme.yaml

# Example validation error:
# Error: Theme validation failed
# - metadata.name is required
# - spec.colors.bg: invalid color format "#gggggg"
# - spec.plugin.repo: repository "invalid/repo" not found
```

### Testing Theme Changes

```bash
# 1. Apply test theme
dvm apply -f test-theme.yaml

# 2. Generate Neovim config
nvp generate

# 3. Test in Neovim (open test files)
nvim ~/test-files/

# 4. Revert if needed
dvm set theme previous-theme
```

---

## Advanced IaC Patterns

### Template Themes

Create reusable theme templates:

```yaml
# theme-template.yaml
apiVersion: devopsmaestro.io/v1
kind: NvimTheme
metadata:
  name: corporate-template
  annotations:
    template: "true"
spec:
  generator: coolnight
  parameters:
    hue: ${HUE}                     # Parameter placeholder
    name: ${THEME_NAME}
    base_colors:
      accent: ${BRAND_COLOR}
```

Use with environment variables or tooling:

```bash
# Using envsubst for templating
export HUE=210
export THEME_NAME=corporate-blue
export BRAND_COLOR="#0078d4"

envsubst < theme-template.yaml | dvm apply -f -
```

### Dynamic Theme Selection

```bash
# Select theme based on project type
PROJECT_TYPE=$(dvm get app --template '{{.metadata.labels.type}}')

case $PROJECT_TYPE in
  "backend")
    dvm apply -f github:company/themes/backend-theme.yaml
    ;;
  "frontend")
    dvm apply -f github:company/themes/frontend-theme.yaml
    ;;
  "ml")
    dvm apply -f github:company/themes/data-science-theme.yaml
    ;;
esac
```

---

## Integration with CI/CD

### Automated Theme Updates

```yaml
# .github/workflows/theme-update.yml
name: Update Team Themes
on:
  push:
    paths:
      - 'themes/**'
  
jobs:
  update-themes:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install DevOpsMaestro
        run: |
          curl -fsSL https://install.devopsmaestro.io | sh
          
      - name: Validate themes
        run: |
          for theme in themes/*.yaml; do
            dvm apply -f "$theme" --dry-run
          done
          
      - name: Deploy to theme server
        run: |
          # Deploy validated themes to central server
          rsync -av themes/ theme-server:/themes/
```

---

## Troubleshooting

### Common Issues

1. **Invalid YAML format:**
   ```bash
   # Check YAML syntax
   yamllint my-theme.yaml
   
   # Validate against schema
   dvm apply -f my-theme.yaml --dry-run
   ```

2. **Network access issues:**
   ```bash
   # Test URL accessibility
   curl -I https://themes.example.com/theme.yaml
   
   # Check GitHub access
   git ls-remote https://github.com/user/themes.git
   ```

3. **Theme not applying:**
   ```bash
   # Check theme was created
   nvp theme list | grep my-theme
   
   # Verify theme content
   dvm get nvim theme my-theme -o yaml
   ```

---

## Next Steps

- [Theme Hierarchy](theme-hierarchy.md) - Understanding theme inheritance
- [Themes Documentation](../nvp/themes.md) - Available themes and variants
- [YAML Schema](../configuration/yaml-schema.md) - Complete YAML format reference
- [Source Types](../advanced/source-types.md) - All supported source formats