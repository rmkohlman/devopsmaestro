---

## Migration from Other Tools

### From Docker Compose

```bash
# Your current setup:
# docker-compose.yml with dev environment

# Keep your docker-compose.yml for production
# Use DevOpsMaestro for development:
cd ~/Developer/dockerized-app
dvm create app dockerized-app --from-cwd
dvm create workspace dev

# DevOpsMaestro provides:
# - Better Neovim integration
# - Language-specific tooling
# - Hierarchical organization
# - Theme management
```

### From VS Code Dev Containers

```bash
# Your current setup:
# .devcontainer/devcontainer.json

# Add to DevOpsMaestro for kubectl-style management:
cd ~/Developer/vscode-project  
dvm create app vscode-project --from-cwd
dvm create workspace dev

# Benefits over VS Code Dev Containers:
# - Works with any editor (Neovim, VS Code, etc.)
# - CLI-based workflow
# - Better for remote development
# - Hierarchical project organization
```

---

## Next Steps

After adding your existing projects:

1. **[Workspace Configuration](../configuration/yaml-schema.md)** - Customize environments with YAML
2. **[Theme System](../nvp/themes.md)** - Set up visual themes across your hierarchy  
3. **[Plugin Management](../nvp/plugins.md)** - Add language-specific Neovim plugins
4. **[Build & Attach Guide](../dvm/build-attach.md)** - Master the container development lifecycle
5. **[Advanced Patterns](../advanced/)** - Multi-container setups, CI/CD integration

---

## Cheat Sheet for Existing Projects

```bash
# Single existing project
cd ~/existing-app  
dvm create eco personal && dvm create dom apps && dvm create a existing-app --from-cwd && dvm create ws dev
dvm build && dvm attach

# Multiple related projects  
dvm create eco company && dvm create dom backend
cd ~/service1 && dvm create a service1 --from-cwd && dvm create ws dev
cd ~/service2 && dvm create a service2 --from-cwd && dvm create ws dev  

# Quick status check
dvm status                    # Overview
dvm get apps --all           # All apps
dvm get workspaces -A        # All workspaces

# Switch between projects
dvm use app service1 && dvm attach    # Work on service1
dvm use app service2 && dvm attach    # Work on service2
```